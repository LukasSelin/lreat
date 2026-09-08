// Package report keeps what each run of the simulation said, on disk, so
// that a run can be read again after the terminal it was printed in has
// gone. A settlement is only ever worth what can be learned from it
// afterwards: the numbers a batch ends on, the shape of the population
// curve, the flags it was run under, and which tree it was run against.
// Held together those are the difference between "it got worse" and
// "it got worse here, under these priors, and here is the run that says so".
//
// Reports are local and not part of the tree. They go under runs/ in a
// folder named for the branch, because several branches are usually being
// run at once and their numbers mean nothing mixed together. Set LREAT_RUNS
// to gather the folders of several worktrees in one place, or LREAT_REPORTS
// to off to keep none at all.
package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Run is one run of one command, from the moment it starts to the moment it
// says how it went. Everything written through Out is kept verbatim; Score
// picks out the few numbers that are worth comparing between runs and puts
// them in the index as well as the report.
type Run struct {
	command string
	args    []string
	start   time.Time
	tree    tree
	body    bytes.Buffer
	scores  []score
	dir     string // empty when nothing will be written
}

type score struct {
	name  string
	value float64
}

// Open starts a report for a command. It never fails in a way the caller has
// to handle: a run that cannot be filed is still a run, and the command goes
// on printing to the terminal exactly as it did before.
func Open(command string) *Run {
	r := &Run{command: command, args: os.Args[1:], start: time.Now(), tree: look()}
	if strings.EqualFold(os.Getenv("LREAT_REPORTS"), "off") {
		return r
	}
	dir := filepath.Join(root(), folder(r.tree.branch))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return r
	}
	r.dir = dir
	return r
}

// Out returns a writer that copies everything into the report on its way to
// w. A command prints what it always printed, and the report is what was on
// the terminal rather than a second account of it that can drift from it.
func (r *Run) Out(w io.Writer) io.Writer {
	if r == nil {
		return w
	}
	return io.MultiWriter(w, &r.body)
}

// Write puts text into the report without printing it.
func (r *Run) Write(p []byte) (int, error) {
	if r == nil {
		return len(p), nil
	}
	return r.body.Write(p)
}

// Score records one number worth holding against another run's. Order is
// kept: they read in the order the command found them.
func (r *Run) Score(name string, value float64) {
	if r == nil {
		return
	}
	r.scores = append(r.scores, score{name, value})
}

// Close writes the report and returns a line saying where it went, or "" if
// nothing was written. The line is meant to be printed: a report nobody can
// find is a report nobody reads.
func (r *Run) Close() string {
	if r == nil || r.dir == "" {
		return ""
	}
	name := r.start.Format("20060102-150405") + "-" + r.command + ".md"
	path := filepath.Join(r.dir, name)
	if err := os.WriteFile(path, []byte(r.markdown()), 0o644); err != nil {
		return ""
	}
	r.appendIndex()
	// Said from where the command was run, so that the line can be opened
	// by whoever is reading it.
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, path); err == nil && !strings.HasPrefix(rel, "..") {
			path = rel
		}
	}
	return "report " + filepath.ToSlash(path) + "\n"
}

// markdown is the report itself: what tree it was run against and under what
// flags, the numbers, and then everything the run printed.
func (r *Run) markdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s — %s\n\n", r.command, r.start.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "    command   go run ./cmd/%s %s\n", r.command, strings.Join(r.args, " "))
	fmt.Fprintf(&b, "    branch    %s\n", r.tree.branch)
	commit := r.tree.commit
	if r.tree.dirty {
		commit += " (uncommitted changes)"
	}
	fmt.Fprintf(&b, "    commit    %s\n", commit)
	fmt.Fprintf(&b, "    took      %s\n", time.Since(r.start).Truncate(time.Second))
	if len(r.scores) > 0 {
		b.WriteString("\n## scores\n\n")
		for _, s := range r.scores {
			fmt.Fprintf(&b, "    %-22s %8.3f\n", s.name, s.value)
		}
	}
	b.WriteString("\n## run\n\n```\n")
	b.WriteString(strings.TrimRight(r.body.String(), "\n"))
	b.WriteString("\n```\n")
	return b.String()
}

// appendIndex adds one line of JSON to the folder's index, which is how a
// hundred runs are read at once: every score of every run on this branch, in
// the order they were taken, greppable without opening any of them.
func (r *Run) appendIndex() {
	scores := map[string]float64{}
	for _, s := range r.scores {
		scores[s.name] = s.value
	}
	line, err := json.Marshal(struct {
		At      string             `json:"at"`
		Command string             `json:"command"`
		Args    []string           `json:"args"`
		Branch  string             `json:"branch"`
		Commit  string             `json:"commit"`
		Dirty   bool               `json:"dirty"`
		Seconds float64            `json:"seconds"`
		Scores  map[string]float64 `json:"scores"`
	}{
		At: r.start.Format(time.RFC3339), Command: r.command, Args: r.args,
		Branch: r.tree.branch, Commit: r.tree.commit, Dirty: r.tree.dirty,
		Seconds: time.Since(r.start).Seconds(), Scores: scores,
	})
	if err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(r.dir, "index.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(append(line, '\n'))
}

// tree is which version of the simulation the run was of. A number without
// it is a number about nothing.
type tree struct {
	branch string
	commit string
	dirty  bool
}

// look asks git where it is. Git is not required — the reports are still
// worth keeping without it — so every answer it fails to give is left blank
// rather than treated as an error.
func look() tree {
	t := tree{branch: git("rev-parse", "--abbrev-ref", "HEAD"), commit: git("rev-parse", "--short", "HEAD")}
	if t.branch == "" || t.branch == "HEAD" {
		t.branch = filepath.Base(root())
	}
	t.dirty = git("status", "--porcelain") != ""
	return t
}

func git(args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// root is where the reports go: LREAT_RUNS if it is set, and otherwise a
// runs/ beside the go.mod of the tree being run, which for a worktree is the
// worktree's own.
func root() string {
	if dir := os.Getenv("LREAT_RUNS"); dir != "" {
		return dir
	}
	dir, err := os.Getwd()
	if err != nil {
		return "runs"
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "runs")
		}
		up := filepath.Dir(dir)
		if up == dir {
			return "runs"
		}
		dir = up
	}
}

// folder is a branch name made safe to be one. Branches are commonly named
// with slashes and the folder is one folder, not a path.
func folder(branch string) string {
	safe := strings.Map(func(r rune) rune {
		if strings.ContainsRune(`/\:*?"<>|`, r) || r < 0x20 {
			return '-'
		}
		return r
	}, branch)
	if safe = strings.Trim(safe, ". "); safe == "" {
		return "unknown"
	}
	return safe
}

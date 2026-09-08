package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A run leaves one file behind and one line in the index, and the file holds
// both what was printed and the numbers it was scored on. Everything a
// report is for is in that sentence.
func TestRunIsKept(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LREAT_RUNS", dir)

	r := Open("tune")
	var terminal bytes.Buffer
	fmt.Fprint(r.Out(&terminal), "gates: fed 0.61\n")
	r.Score("fed", 0.61)
	if line := r.Close(); line == "" {
		t.Fatal("nothing was written")
	}

	if got := terminal.String(); got != "gates: fed 0.61\n" {
		t.Errorf("the terminal got %q, not what the command printed", got)
	}
	body := readOne(t, dir)
	for _, want := range []string{"gates: fed 0.61", "fed", "0.610", "go run ./cmd/tune"} {
		if !strings.Contains(body, want) {
			t.Errorf("the report does not say %q:\n%s", want, body)
		}
	}
}

func TestIndexHoldsTheScores(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LREAT_RUNS", dir)
	for _, pop := range []float64{12, 34} {
		r := Open("headless")
		r.Score("pop", pop)
		r.Close()
	}
	lines := strings.Split(strings.TrimSpace(index(t, dir)), "\n")
	if len(lines) != 2 {
		t.Fatalf("two runs left %d lines in the index", len(lines))
	}
	var got struct {
		Command string
		Scores  map[string]float64
	}
	if err := json.Unmarshal([]byte(lines[1]), &got); err != nil {
		t.Fatalf("the index is not JSON: %v", err)
	}
	if got.Command != "headless" || got.Scores["pop"] != 34 {
		t.Errorf("the second run reads as %+v", got)
	}
}

// Turned off, a run is still a run: it prints what it always printed and
// leaves nothing behind.
func TestOffWritesNothing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LREAT_RUNS", dir)
	t.Setenv("LREAT_REPORTS", "off")
	r := Open("watch")
	var terminal bytes.Buffer
	fmt.Fprint(r.Out(&terminal), "nothing ran\n")
	if line := r.Close(); line != "" {
		t.Errorf("it filed a report anyway: %q", line)
	}
	if terminal.String() != "nothing ran\n" {
		t.Errorf("the terminal got %q", terminal.String())
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("it left %d entries behind", len(entries))
	}
}

// A branch is one folder however it is named, and the slashes people put in
// branch names are not path separators here.
func TestFolderIsOneFolder(t *testing.T) {
	for branch, want := range map[string]string{
		"master":                "master",
		"claude/reports-3459bb": "claude-reports-3459bb",
		`feature\odd:name`:      "feature-odd-name",
		"":                      "unknown",
	} {
		if got := folder(branch); got != want {
			t.Errorf("folder(%q) = %q, want %q", branch, got, want)
		}
	}
}

// readOne returns the single report under the branch folder in dir.
func readOne(t *testing.T, dir string) string {
	t.Helper()
	paths, _ := filepath.Glob(filepath.Join(dir, "*", "*.md"))
	if len(paths) != 1 {
		t.Fatalf("expected one report, found %v", paths)
	}
	b, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func index(t *testing.T, dir string) string {
	t.Helper()
	paths, _ := filepath.Glob(filepath.Join(dir, "*", "index.jsonl"))
	if len(paths) != 1 {
		t.Fatalf("expected one index, found %v", paths)
	}
	b, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The pre-commit hook only checks the document when a commit touches something
// the document is built from, and it reads that list out of a file. A list
// like that is the same hazard as a table with a row missing: wrong in the
// quiet direction, because a package left out is not a false alarm, it is a
// commit that changes the ontology and is never checked.
//
// So the list is derived here rather than trusted. `go list -deps` knows what
// the generator actually reads, and it knew before this test was written that
// core/need is one of them - it arrives through core/entity, and a list
// written by hand would not have had it.

const hookPaths = "../.githooks/ontology-paths"

func TestTheHookWatchesWhatTheDocumentIsBuiltFrom(t *testing.T) {
	want, err := generatorDeps()
	if err != nil {
		t.Skipf("cannot ask what the generator reads: %v", err)
	}
	got, err := watchedPaths()
	if err != nil {
		t.Fatalf("cannot read %s: %v", hookPaths, err)
	}

	for _, pkg := range want {
		if !got[pkg] {
			t.Errorf("the document is built from %s and the hook does not watch it:\n"+
				"    add %s to .githooks/ontology-paths", pkg, pkg)
		}
		delete(got, pkg)
	}
	for pkg := range got {
		t.Errorf("the hook watches %s and the document is not built from it:\n"+
			"    remove %s from .githooks/ontology-paths", pkg, pkg)
	}
}

// The hook has to be a hook: readable, and executable by the shell git hands
// it to.
func TestTheHookIsInstallable(t *testing.T) {
	hook := filepath.Join("..", ".githooks", "pre-commit")
	b, err := os.ReadFile(hook)
	if err != nil {
		t.Fatalf("no pre-commit hook: %v", err)
	}
	if !strings.HasPrefix(string(b), "#!") {
		t.Error("the hook has no interpreter line, so git cannot run it")
	}
	// Git carries the executable bit in the index rather than on the file, so
	// that is what has to be right for anyone who clones this.
	out, err := exec.Command("git", "ls-files", "-s", ".githooks/pre-commit").Output()
	if err != nil {
		t.Skipf("cannot read the index: %v", err)
	}
	if mode := strings.Fields(string(out)); len(mode) > 0 && mode[0] != "100755" {
		t.Errorf("the hook is committed %s and has to be 100755:\n"+
			"    git update-index --chmod=+x .githooks/pre-commit", mode[0])
	}
}

// generatorDeps is every package of this repository the generator reads,
// as repo-relative directories.
func generatorDeps() ([]string, error) {
	out, err := exec.Command("go", "list", "-deps", "lreat/owl").Output()
	if err != nil {
		return nil, err
	}
	var pkgs []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(line, "lreat/"); ok && rest != "" {
			pkgs = append(pkgs, rest)
		}
	}
	sort.Strings(pkgs)
	return pkgs, nil
}

func watchedPaths() (map[string]bool, error) {
	b, err := os.ReadFile(hookPaths)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out[line] = true
	}
	return out, nil
}

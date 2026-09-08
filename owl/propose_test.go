package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// propose applies a proposal to the live trees, because the only honest answer
// to "what would this cost" is the catalog Instantiate then returns - not a
// guess about it. That makes it a one-shot: a case run in this process would
// leave the trees changed for every test after it, and the failures would land
// somewhere else entirely. So each case runs in a child, which is also the way
// the command is really used.

const childEnv = "OWL_PROPOSE_CHILD"

// TestProposeChild is not a case of its own. It is the body a child runs when
// runPropose re-execs the test binary, and it skips in the parent.
func TestProposeChild(t *testing.T) {
	path := os.Getenv(childEnv)
	if path == "" {
		t.Skip("runs only as a child of runPropose")
	}
	o, err := Build()
	if err != nil {
		t.Fatal(err)
	}
	o.Sort()
	if err := propose(path, o, os.Stdout); err != nil {
		os.Stdout.WriteString("error: " + err.Error() + "\n")
	}
}

func runPropose(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "proposal.ofn")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestProposeChild$")
	cmd.Env = append(os.Environ(), childEnv+"="+path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("child failed: %v\n%s", err, out)
	}
	return string(out)
}

func TestPropose(t *testing.T) {
	for _, tc := range []struct {
		name     string
		proposal string
		want     []string
		notWant  []string
	}{{
		name: "a new material, and somewhere to get it",
		proposal: `Declaration(Class(:Pitch))
SubClassOf(:Pitch :Material)
SubClassOf(:Pitch :Burnable)
SubClassOf(:Pitch DataHasValue(:lack "0.5"^^xsd:decimal))
SubClassOf(:Wood ObjectSomeValuesFrom(:affords :Pitch))`,
		want: []string{
			`Pitch = lack(New("pitch", Material, Burnable, habit.Signature{}), 0.5)`,
			"Wood: {..., Pitch},",
			"+ take/pitch@wood",
		},
		notWant: []string{"Nothing affords it"},
	}, {
		// The same class with nowhere to get it is a class nobody can have,
		// and saying so is most of the point of answering with the catalog.
		name: "a new material with nothing affording it",
		proposal: `Declaration(Class(:Pitch))
SubClassOf(:Pitch :Material)`,
		want:    []string{"Nothing affords it", "no change"},
		notWant: []string{"take/pitch"},
	}, {
		// The expensive one. Stone is a leaf, so hanging a child off it moves
		// every act keyed on stone to a key nobody has learned.
		name: "a leaf turned into a branch",
		proposal: `Declaration(Class(:Granite))
SubClassOf(:Granite :Stone)
SubClassOf(:Granite :Heavy)`,
		want: []string{
			"stone had no children of its own",
			"- take/stone@outcrop",
			"+ take/granite@outcrop",
			"keeps a slot nothing reads",
		},
	}, {
		name: "a site, and a season on what it grows",
		proposal: `Declaration(Class(:Orchard))
SubClassOf(:Orchard :Ground)
SubClassOf(:Orchard :Living)
SubClassOf(:Orchard DataHasValue(:at-near "0.5"^^xsd:decimal))
SubClassOf(:Orchard ObjectSomeValuesFrom(:affords :Berries))`,
		want: []string{
			`at(New("orchard", Ground, Living, habit.Signature{}), habit.Signature{habit.Near: 0.5})`,
			"+ take/berries@orchard",
		},
	}, {
		// The vocabulary is the guard, and a slip has to stop everything: a
		// proposal half applied is worse than one refused.
		name: "a misspelling stops the whole proposal",
		proposal: `Declaration(Class(:Pitch))
SubClassOf(:Pitch :Materal)`,
		want: []string{
			"unknown class :Materal",
			"did you mean :Material?",
			"nothing applied",
		},
		notWant: []string{"class.go"},
	}, {
		// Declaring is the one way past the vocabulary, so a term used without
		// one is refused even when it is spelt perfectly well.
		name:     "an undeclared term is not a new term",
		proposal: `SubClassOf(:Pitch :Material)`,
		want:     []string{"unknown class :Pitch", "or declare it", "Declaration(Class(:Pitch))"},
	}, {
		name: "something the trees have no room for",
		proposal: `Declaration(Class(:Ferment))
SubClassOf(:Ferment :Transform)`,
		want: []string{"no declaration to paste", "Transforms in core/ontology/relation.go"},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			out := runPropose(t, tc.proposal)
			for _, want := range tc.want {
				if !strings.Contains(out, want) {
					t.Errorf("missing %q in:\n%s", want, out)
				}
			}
			for _, no := range tc.notWant {
				if strings.Contains(out, no) {
					t.Errorf("unwanted %q in:\n%s", no, out)
				}
			}
		})
	}
}

// Re-asserting what the document already says has to come back as nothing to
// do. A proposal is usually mostly things that are already true - it is the
// generated document with a few lines added - so a tool that reported every
// line as a change would bury the ones that are not.
func TestProposeIsIdempotentAgainstItself(t *testing.T) {
	out := runPropose(t, "SubClassOf(:Timber :Buildable)")
	if !strings.Contains(out, "The trees already say this") {
		t.Errorf("re-asserting what the document already says moved something:\n%s", out)
	}
	if strings.Contains(out, "catalog:") {
		t.Errorf("nothing changed, so there is no catalog to report:\n%s", out)
	}
}

package main

import (
	"strings"
	"testing"
)

// A removal is only expressible against the whole document, so each case here
// is the generated document with lines taken out of it. That is also how it
// would really be used: edit the .ofn, and ask what deleting that costs.

func withoutLines(t *testing.T, drop ...string) string {
	t.Helper()
	o, err := Build()
	if err != nil {
		t.Fatal(err)
	}
	o.Sort()
	var kept []string
	dropped := map[string]bool{}
	for _, line := range strings.Split(o.Functional(), "\n") {
		trimmed := strings.TrimSpace(line)
		var skip bool
		for _, d := range drop {
			if trimmed == d {
				skip, dropped[d] = true, true
			}
		}
		if !skip {
			kept = append(kept, line)
		}
	}
	for _, d := range drop {
		if !dropped[d] {
			t.Fatalf("the document has no line %q to take out; the test is stale", d)
		}
	}
	return strings.Join(kept, "\n")
}

func TestRemoval(t *testing.T) {
	for _, tc := range []struct {
		name    string
		drop    []string
		want    []string
		notWant []string
	}{{
		// The one that can be applied, so the catalog answers for real.
		name: "a row out of Affords",
		drop: []string{"SubClassOf(:Wood ObjectSomeValuesFrom(:affords :Timber))"},
		want: []string{
			"stops affording Timber",
			"drop Timber from Wood's row in Affords",
			"- take/timber@wood",
		},
	}, {
		// The one that cannot, so it is answered with what names it.
		name: "a whole class",
		drop: []string{
			"Declaration(Class(:Coin))",
			"SubClassOf(:Coin :Material)",
		},
		want: []string{
			":Coin  (thing/material/coin) would go",
			"exchange/coin>provision@market",
			"will not compile",
			"does not count the class removals",
		},
	}, {
		// A trait a class owns can go, and what else reads it is the thing
		// worth knowing before it does.
		name: "a trait the class owns",
		drop: []string{"SubClassOf(:Dwelling :Hearth)"},
		want: []string{
			"loses hearth",
			"drop hearth from the traits on dwelling",
			"also carried by: dwelling, tavern",
			"placed by it: make provision+timber > meal @hearth",
		},
	}, {
		// A trait nothing in the ontology reads is a trait that has stopped
		// earning its place, and saying so is the useful answer.
		name: "a trait nothing here reads",
		drop: []string{"SubClassOf(:Timber :Burnable)"},
		want: []string{"Nothing in the ontology reads it", "grep the trait name"},
	}, {
		// Everything that goes with a deleted class goes quietly. Listing the
		// label and the disjointness beside it would be reporting one removal
		// four times in a form nobody can act on separately.
		name: "what a deleted class drags with it is not reported twice",
		drop: []string{
			"Declaration(Class(:Coin))",
			"SubClassOf(:Coin :Material)",
			`AnnotationAssertion(rdfs:label :Coin "coin"^^xsd:string)`,
			"DisjointClasses(:Provision :Timber :Stone :Tool :Coin)",
		},
		want:    []string{"would go"},
		notWant: []string{"not read as any of the above"},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			out := runPropose(t, withoutLines(t, tc.drop...))
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

// The document as generated proposes nothing, in either direction. If this
// fails, the diff is seeing a difference that is not there - a rendering that
// does not survive a round trip through the parser.
func TestTheDocumentProposesNothingAgainstItself(t *testing.T) {
	out := runPropose(t, withoutLines(t))
	if !strings.Contains(out, "nothing proposed") {
		t.Errorf("the document differs from itself:\n%s", out)
	}
}

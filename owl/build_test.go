package main

import (
	"os"
	"slices"
	"strings"
	"testing"

	"gowl/el"
	"gowl/lint"
	"gowl/owl"

	"lreat/core/habit"
	"lreat/core/ontology"
)

// The point of these is not that the document says any particular thing. It
// is that it says what core/ontology says: every test here reads both sides
// and compares them, so a class added to the trees fails nothing and a class
// the rendering quietly drops fails here.

func build(t *testing.T) *owl.Ontology {
	t.Helper()
	o, err := Build()
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func TestEveryClassIsRendered(t *testing.T) {
	o := build(t)
	for _, root := range []*ontology.Class{ontology.Thing, ontology.Site} {
		for _, c := range root.Family() {
			cls := o.Class(classIRI(c))
			if !o.IsDeclared(cls) {
				t.Errorf("%s is in the trees but not in the document", c.Path())
				continue
			}
			if c.Parent == nil {
				continue
			}
			if !hasSuper(o, cls, o.Class(classIRI(c.Parent))) {
				t.Errorf("%s is not under %s", c.Path(), c.Parent.Name)
			}
		}
	}
}

// A trait is a bitfield there and a superclass here, so what has to survive
// the crossing is Has: everything a class has, own or inherited, has to come
// back out of the ancestor closure.
func TestTraitsBecomeSuperclasses(t *testing.T) {
	o := build(t)
	for _, root := range []*ontology.Class{ontology.Thing, ontology.Site} {
		for _, c := range root.Family() {
			ancestors := map[owl.Class]bool{}
			for _, a := range o.AncestorsOf(o.Class(classIRI(c))) {
				ancestors[a] = true
			}
			for bit := ontology.Trait(1); bit != 0; bit <<= 1 {
				if bit.String() == "" || !c.Has(bit) {
					continue
				}
				if !ancestors[o.Class(traitIRI(bit))] {
					t.Errorf("%s has %s but is not under %s", c.Name, bit, traitIRI(bit))
				}
			}
		}
	}
}

func TestEveryActIsRendered(t *testing.T) {
	o := build(t)
	acts := ontology.Instantiate()
	if len(acts) == 0 {
		t.Fatal("the catalog is empty")
	}
	for _, in := range acts {
		cls := o.Class(actIRI(in.Key))
		if !o.IsDeclared(cls) {
			t.Errorf("act %s is in the catalog but not in the document", in.Key)
			continue
		}
		if got := o.Label(cls); got != in.Key {
			t.Errorf("act %s is labelled %q", in.Key, got)
		}
		if !hasSuper(o, cls, o.Class(":Act")) {
			t.Errorf("act %s is not an :Act", in.Key)
		}
		if !hasSome(o, cls, ":verb", verbIRI(in.Schema.Verb)) {
			t.Errorf("act %s does not name its verb", in.Key)
		}
	}
	// One class per act, and nothing else wearing an act name.
	var n int
	for _, ax := range o.AxiomsReferencing(o.Class(":Act")) {
		if sc, ok := owl.Unwrap(ax).(owl.SubClassOf); ok && sc.Super == o.Class(":Act") {
			n++
		}
	}
	if n != len(acts) {
		t.Errorf("%d acts in the catalog, %d in the document", len(acts), n)
	}
}

// Affords is what makes Take instantiate at all: a map with no outcrop has no
// quarrying, and a document that forgets a row would say the same of a map
// that has one.
func TestAffordsSurvives(t *testing.T) {
	o := build(t)
	for holder, materials := range ontology.Affords {
		for _, m := range materials {
			if !hasSome(o, o.Class(classIRI(holder)), ":affords", classIRI(m)) {
				t.Errorf("%s affords %s, and the document does not say so", holder.Name, m.Name)
			}
		}
	}
}

func TestProcessesAndTransforms(t *testing.T) {
	o := build(t)
	for _, p := range ontology.Processes {
		cls := o.Class(":" + slug(p.Name) + "-process")
		if !o.IsDeclared(cls) {
			t.Errorf("process %s is missing", p.Name)
			continue
		}
		if !hasSome(o, cls, ":processOf", classIRI(p.Of)) {
			t.Errorf("process %s does not say what it happens to", p.Name)
		}
		for i, s := range p.Stages {
			if !o.IsDeclared(o.Class(":" + slug(p.Name) + "-" + slug(s.Name))) {
				t.Errorf("stage %d of %s (%s) is missing", i, p.Name, s.Name)
			}
		}
	}
	for i := range ontology.Transforms {
		tr := &ontology.Transforms[i]
		if !o.IsDeclared(o.Class(transformIRI(tr))) {
			t.Errorf("transform %s is missing", transformLabel(tr))
		}
	}
}

// A sanity check on the numbers, which are the easiest thing to render into
// the wrong place. Seizing is the act with the most to say about itself.
func TestSignaturesReachTheDocument(t *testing.T) {
	o := build(t)
	var seize *ontology.Instance
	for _, in := range ontology.Instantiate() {
		if strings.Contains(in.Key, "<holder") {
			c := in
			seize = &c
			break
		}
	}
	if seize == nil {
		t.Skip("no seizing act in the catalog")
	}
	cls := o.Class(actIRI(seize.Key))
	for i, v := range seize.Prior {
		if v == 0 {
			continue
		}
		want := owl.DataValue(o.DataProperty(":prior-"+habit.Names[i]), dec(v))
		if !hasSuperExpression(o, cls, want) {
			t.Errorf("%s: prior %s is %v in the catalog and not in the document", seize.Key, habit.Names[i], v)
		}
	}
}

// The document is checked in, so a stale one is a real failure rather than a
// cosmetic one: it is what anybody reading the repository sees.
func TestCheckedInDocumentIsCurrent(t *testing.T) {
	want, err := os.ReadFile("lreat.ofn")
	if err != nil {
		t.Skipf("no checked-in document: %v", err)
	}
	o := build(t)
	o.Sort()
	if got := o.Functional(); got != strings.ReplaceAll(string(want), "\r\n", "\n") {
		t.Error("owl/lreat.ofn is out of date; run: go run ./owl")
	}
}

func TestLintIsQuiet(t *testing.T) {
	o := build(t)
	o.Sort()
	for _, f := range lint.Run(o, lint.Default()) {
		// The roots are orphans by construction, which gowl reports at info.
		if f.Severity >= lint.Warning {
			t.Errorf("%v", f)
		}
	}
}

// The document lands in OWL 2 EL, which is what makes it worth classifying:
// siblings are disjoint and traits are superclasses, so a class that has two
// traits which cannot both hold, or that ends up under two disjoint siblings,
// comes out unsatisfiable here rather than going unnoticed in a bitfield.
func TestTheTreesAreCoherent(t *testing.T) {
	o := build(t)
	if got := owl.Profiles(o); !slices.Contains(got, owl.ProfileEL) {
		t.Fatalf("the document is no longer in EL, so it cannot be classified: %v", got)
	}
	c := el.Classify(o)
	if !c.IsCoherent() {
		t.Errorf("unsatisfiable classes: %v", c.UnsatisfiableClasses())
	}
}

func hasSuper(o *owl.Ontology, sub, super owl.Class) bool {
	for _, s := range o.SuperClassesOf(sub) {
		if s == super {
			return true
		}
	}
	return false
}

func hasSome(o *owl.Ontology, sub owl.Class, prop, filler string) bool {
	return hasSuperExpression(o, sub, owl.Some(o.ObjectProperty(prop), o.Class(filler)))
}

func hasSuperExpression(o *owl.Ontology, sub owl.Class, want owl.ClassExpression) bool {
	for _, ax := range o.AxiomsReferencing(sub) {
		sc, ok := owl.Unwrap(ax).(owl.SubClassOf)
		if ok && sc.Sub == sub && owl.Equal(sc.Super, want) {
			return true
		}
	}
	return false
}

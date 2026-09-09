package ontology

import "testing"

// walkIsA is what IsA was before it was a bit test: a walk up the parents.
func walkIsA(c, anc *Class) bool {
	for x := c; x != nil; x = x.Parent {
		if x == anc {
			return true
		}
	}
	return false
}

// IsA reads an ancestry worked out when the trees were declared. It has to
// answer for every pair of classes exactly what the walk answered, or every
// question asked about what a tile is has a different answer.
func TestIsAAnswersWhatTheWalkAnswered(t *testing.T) {
	var all []*Class
	for _, root := range []*Class{Thing, Site} {
		all = append(all, root.Family()...)
	}
	if len(all) < 20 {
		t.Fatalf("only %d classes found; the trees are not being walked", len(all))
	}
	for _, c := range all {
		for _, anc := range all {
			if got, want := c.IsA(anc), walkIsA(c, anc); got != want {
				t.Fatalf("%s is-a %s: %v, the walk says %v", c.Path(), anc.Path(), got, want)
			}
		}
		if c.IsA(nil) {
			t.Fatalf("%s is a nothing", c.Path())
		}
	}
	var none *Class
	if none.IsA(Thing) {
		t.Fatal("nothing is a thing")
	}
}

// Every class must have an id of its own, or two of them share a bit and
// each is the other's ancestor.
func TestEveryClassHasAnAncestryOfItsOwn(t *testing.T) {
	seen := map[int]*Class{}
	for _, root := range []*Class{Thing, Site} {
		for _, c := range root.Family() {
			if was, ok := seen[c.id]; ok {
				t.Fatalf("%s and %s were both declared %d", was.Path(), c.Path(), c.id)
			}
			seen[c.id] = c
			if !c.IsA(c) {
				t.Fatalf("%s is not itself", c.Path())
			}
		}
	}
}

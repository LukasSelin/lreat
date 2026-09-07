package action

import (
	"testing"

	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/world"
)

// bear puts a child of a's beside it, of the given age, with nothing.
func bear(w *world.World, a *entity.Agent, age int) *entity.Agent {
	c := w.SpawnAt("c", w.RandomPersonality(), a.Pos)
	c.Parent = a.ID
	c.Born = w.Tick - age
	c.Inventory[entity.Food] = 0
	c.Needs[need.Physiological] = 0.5
	return c
}

// A parent feeds its own before they are desperate. Half fed is not a
// stranger's business and it is a parent's, and that gap is the whole
// mechanical difference between charity and rearing.
func TestAParentFeedsItsOwnBeforeAStrangerWould(t *testing.T) {
	w, a := shore(t)
	a.Inventory[entity.Food] = 3
	c := bear(w, a, 2*clock.Year)

	if _, ok := Give.Target(a, w); ok {
		t.Fatal("a child half fed is not yet anybody's charity")
	}
	if !run(w, a, FeedChild) {
		t.Fatal("a parent with food and a hungry child of its own should feed it")
	}
	if c.Inventory[entity.Food] != 1 {
		t.Fatalf("the child holds %.2f food, want the unit its parent handed over", c.Inventory[entity.Food])
	}
	if a.Inventory[entity.Food] != 2 {
		t.Fatalf("the parent holds %.2f food, want one unit fewer", a.Inventory[entity.Food])
	}
}

// Somebody else's child is somebody else's. Rearing is preferential or it
// is nothing: it is the one thing in the catalog that asks whose you are.
func TestAParentDoesNotRearTheNeighboursChildren(t *testing.T) {
	w, a := shore(t)
	a.Inventory[entity.Food] = 3
	c := bear(w, a, 2*clock.Year)
	c.Parent = a.ID + 100 // somebody else's, standing in the same place

	if _, ok := FeedChild.Target(a, w); ok {
		t.Fatal("a stranger's child is not one's own to feed")
	}
	if TeachChild.Available(a, w) {
		t.Fatal("a stranger's child is not one's own to bring up")
	}
}

// A grown child feeds itself. The role runs out at maturity, which is what
// keeps a household from being a lifelong dependency.
func TestRearingEndsAtMaturity(t *testing.T) {
	w, a := shore(t)
	a.Inventory[entity.Food] = 3
	bear(w, a, entity.Maturity+clock.Year)

	if _, ok := FeedChild.Target(a, w); ok {
		t.Fatal("a grown child is not a child to be fed")
	}
}

// What a household keeps is its craft. A parent shows a child the work, and
// the child comes away with more of it than a pupil in the square would.
func TestAParentBringsAChildUpInTheWork(t *testing.T) {
	w, a := shore(t)
	a.AddSkill(entity.Farming, 0.6)
	c := bear(w, a, 3*clock.Year)
	before := c.Skills[entity.Farming]

	if !run(w, a, TeachChild) {
		t.Fatal("a parent with a craft and a child of its own should pass it on")
	}
	if got := c.Skills[entity.Farming] - before; got <= 0 {
		t.Fatalf("the child learned %.3f, want a share of what its parent knows", got)
	}
	if c.Skills[entity.Farming]-before < 0.04 {
		t.Fatalf("a child took %.3f from a showing, want more than a stranger's lesson gives",
			c.Skills[entity.Farming]-before)
	}
}

// Nothing to show, nothing to pass on.
func TestAParentWithNoCraftTeachesNothing(t *testing.T) {
	w, a := shore(t)
	for s := range a.Skills {
		a.Skills[s] = 0
	}
	bear(w, a, 3*clock.Year)
	if TeachChild.Available(a, w) {
		t.Fatal("a parent who can do nothing has nothing to bring a child up in")
	}
}

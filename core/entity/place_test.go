package entity

import "testing"

// A child is shown the country it grows up in - on a slice of its own. Habits
// and Reach are arrays and copy on assignment; places are not, and a shared
// backing store would let either party's next discovery land in the other's
// memory.
func TestPlacesArePassedOnAndNotShared(t *testing.T) {
	parent := &Agent{}
	for i := 0; i < 3; i++ {
		parent.Remember(Pos{X: i}, float64(i), 0)
	}
	child := &Agent{Places: parent.CopyPlaces()}
	if len(child.Places) != len(parent.Places) {
		t.Fatalf("child knows %d places, parent %d", len(child.Places), len(parent.Places))
	}
	child.Remember(Pos{X: 99}, 99, 1)
	if _, ok := parent.Knows(Pos{X: 99}); ok {
		t.Fatal("the child's discovery was written into its parent's memory")
	}
	parent.Remember(Pos{X: 0}, 50, 1)
	if p, _ := child.Knows(Pos{X: 0}); p.Worth != 0 {
		t.Fatalf("the parent's second look changed the child's memory to %.0f", p.Worth)
	}
}

// Nobody carries the whole country in their head.
func TestAHeadHasARoom(t *testing.T) {
	a := &Agent{}
	for i := 0; i < MaxPlaces*3; i++ {
		a.Remember(Pos{X: i}, float64(i%7), i)
	}
	if len(a.Places) > MaxPlaces {
		t.Fatalf("carrying %d places, cap is %d", len(a.Places), MaxPlaces)
	}
}

// Ground that turns out to be no use is dropped.
func TestGroundCanBeForgotten(t *testing.T) {
	a := &Agent{}
	a.Remember(Pos{X: 1}, 5, 0)
	a.Remember(Pos{X: 2}, 5, 0)
	a.Forget(Pos{X: 1})
	if _, ok := a.Knows(Pos{X: 1}); ok {
		t.Fatal("forgotten ground is still in mind")
	}
	if _, ok := a.Knows(Pos{X: 2}); !ok {
		t.Fatal("forgetting one place lost another")
	}
}

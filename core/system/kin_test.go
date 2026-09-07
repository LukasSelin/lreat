package system

import (
	"testing"

	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/world"
)

// A roof covers the people asleep under it. Shelter is kept per body and a
// house is not, so without this a child holds whatever it was born with and
// watches it rot for the fifteen years before it can carry a beam.
func TestARoofCoversTheChildren(t *testing.T) {
	w := world.New(5)
	parent := w.Spawn("p", need.Neutral())
	parent.Shelter = 0.9
	child := w.SpawnAt("c", need.Neutral(), parent.Pos)
	child.Parent, child.Born, child.Shelter = parent.ID, w.Tick-2*clock.Year, 0

	Household(w)
	if child.Shelter != parent.Shelter {
		t.Fatalf("the child is sheltered %.2f under a roof worth %.2f", child.Shelter, parent.Shelter)
	}

	// Never more than the roof is worth: a household is exactly as well
	// housed as the person who built it.
	parent.Shelter = 0.3
	Household(w)
	if child.Shelter <= parent.Shelter {
		t.Fatal("a child's shelter should be its own once it is better than the roof it was given")
	}
	child.Shelter = 0
	Household(w)
	if child.Shelter != 0.3 {
		t.Fatalf("the child is sheltered %.2f, want the 0.30 the roof is now worth", child.Shelter)
	}
}

// The roof is a parent's own, not the settlement's. Nobody is housed by
// somebody else's walls for being nearby.
func TestARoofCoversNobodyElse(t *testing.T) {
	w := world.New(5)
	parent := w.Spawn("p", need.Neutral())
	parent.Shelter = 0.9
	stranger := w.SpawnAt("s", need.Neutral(), parent.Pos)
	stranger.Shelter = 0
	grown := w.SpawnAt("g", need.Neutral(), parent.Pos)
	grown.Parent, grown.Born, grown.Shelter = parent.ID, w.Tick-entity.Maturity-clock.Year, 0

	Household(w)
	if stranger.Shelter != 0 {
		t.Fatalf("a stranger standing by is sheltered %.2f, want none of it", stranger.Shelter)
	}
	if grown.Shelter != 0 {
		t.Fatalf("a grown child is sheltered %.2f, want a roof of its own or none", grown.Shelter)
	}
}

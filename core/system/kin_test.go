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

	Rearing(w)
	if child.Shelter != parent.Shelter {
		t.Fatalf("the child is sheltered %.2f under a roof worth %.2f", child.Shelter, parent.Shelter)
	}

	// Never more than the roof is worth: a household is exactly as well
	// housed as the person who built it.
	parent.Shelter = 0.3
	Rearing(w)
	if child.Shelter <= parent.Shelter {
		t.Fatal("a child's shelter should be its own once it is better than the roof it was given")
	}
	child.Shelter = 0
	Rearing(w)
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

	Rearing(w)
	if stranger.Shelter != 0 {
		t.Fatalf("a stranger standing by is sheltered %.2f, want none of it", stranger.Shelter)
	}
	if grown.Shelter != 0 {
		t.Fatalf("a grown child is sheltered %.2f, want a roof of its own or none", grown.Shelter)
	}
}

// rear runs a childhood of the given length and says what came of it: how
// many of n children lived, and the body the survivors grew into. tend is
// what a parent does about each of them, in tending a year.
func rear(t *testing.T, n int, tendAYear float64) (lived int, body float64) {
	t.Helper()
	w := world.New(11)
	parent := w.Spawn("p", need.Neutral())
	parent.Shelter = 0
	kids := make([]*entity.Agent, n)
	for i := range kids {
		c := w.SpawnAt("c", need.Neutral(), parent.Pos)
		c.Parent, c.Born, c.Tended, c.Vitality = parent.ID, w.Tick, 1, 1
		kids[i] = c
	}
	alive := map[*entity.Agent]bool{}
	for _, c := range kids {
		alive[c] = true
	}
	for day := 0; day < entity.Maturity; day++ {
		w.Tick++
		Rearing(w)
		for _, c := range kids {
			if !alive[c] {
				continue
			}
			if tendAYear > 0 && day%(clock.Year/4) == 0 {
				c.Tended = min(1, c.Tended+tendAYear/4)
			}
			if neglected(w, c, c.Age(w.Tick)) {
				alive[c] = false
			}
		}
	}
	for _, c := range kids {
		if alive[c] {
			lived++
			body += c.Vitality
		}
	}
	if lived > 0 {
		body /= float64(lived)
	}
	return lived, body
}

// A child nobody comes back to is a child a third of whom do not reach
// fifteen. This is the other half of the ledger: without it, tending is a
// kindness rather than the thing a settlement lives by.
func TestAnUntendedChildOftenDoesNotGrowUp(t *testing.T) {
	lived, _ := rear(t, 200, 0)
	if share := float64(lived) / 200; share > 0.8 {
		t.Errorf("%.2f of untended children grew up; neglect should cost a childhood", share)
	}
	if share := float64(lived) / 200; share < 0.4 {
		t.Errorf("%.2f of untended children grew up; neglect should be a risk, not a cull", share)
	}
}

// And a child somebody keeps an eye on mostly lives, and grows into the
// better body. A handful of visits a year is all it takes, which is the
// cadence the acts are set to.
func TestATendedChildGrowsUpAndGrowsWell(t *testing.T) {
	tended, tendedBody := rear(t, 200, 0.5)
	untended, untendedBody := rear(t, 200, 0)
	if tended <= untended {
		t.Fatalf("%d tended children lived against %d untended; rearing has to pay", tended, untended)
	}
	if share := float64(tended) / 200; share < 0.9 {
		t.Errorf("%.2f of tended children grew up, want nearly all of them", share)
	}
	if tendedBody <= untendedBody {
		t.Errorf("a tended child grew into %.2f of a body and an untended one %.2f; a childhood should tell",
			tendedBody, untendedBody)
	}
	if tendedBody < 1 {
		t.Errorf("a tended child grew into %.2f of a body, want an ordinary one at least", tendedBody)
	}
}

// Nothing reads a childhood after it is over: what it did is in the body by
// then, and a grown agent is not still being reared.
func TestRearingStopsAtMaturity(t *testing.T) {
	if got := entity.Neglect(entity.Maturity, 0); got != 0 {
		t.Errorf("a grown agent carries a neglect risk of %g, want none", got)
	}
	if young, old := entity.Neglect(0, 0), entity.Neglect(entity.Maturity-clock.Year, 0); young <= old {
		t.Errorf("an infant risks %g and a near-grown child %g; dependence should fall away", young, old)
	}
	if got := entity.Neglect(clock.Year, 1); got != 0 {
		t.Errorf("a tended child risks %g, want none", got)
	}
}

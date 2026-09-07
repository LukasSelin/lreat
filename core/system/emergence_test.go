package system

import (
	"testing"

	"lreat/core/belief"
	"lreat/core/clock"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/observe"
)

// TestSocialLifeEmerges runs a settlement with no player and no scripted
// content and checks that the belief layer actually produces a social life:
// people ask each other for work, some of it gets done, and the moral acts
// happen because agents chose them rather than because anything triggered
// them. It logs the tallies, which is the fastest way to see the effect of a
// tuning change.
func TestSocialLifeEmerges(t *testing.T) {
	w := valueWorld(31)
	for i := 0; i < 25; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	Run(w, 15*clock.Year)

	counts := map[event.Kind]int{}
	for _, e := range w.Log.All() {
		counts[e.Kind]++
	}
	s := observe.Take(w)

	t.Logf("population %d, houses %d, fields %d, techs %v",
		s.Population, s.Houses, s.Fields, s.Techs)
	t.Logf("asked %d, fulfilled %d, unmet %d, open now %d (%d by name)",
		counts[event.Requested], counts[event.Fulfilled], counts[event.Unmet],
		s.OpenRequests, s.NamedRequests)
	t.Logf("stolen %d, given %d", counts[event.Stolen], counts[event.Given])
	t.Logf("norms honesty %.2f charity %.2f industry %.2f tradition %.2f",
		s.MeanNorms[belief.Honesty], s.MeanNorms[belief.Charity],
		s.MeanNorms[belief.Industry], s.MeanNorms[belief.Tradition])
	t.Logf("outcasts %d", len(Outcast(w)))

	if counts[event.Requested] == 0 {
		t.Error("no request was ever posted")
	}
	if counts[event.Fulfilled] == 0 {
		t.Error("no request was ever fulfilled")
	}
	if counts[event.Stolen] == 0 && counts[event.Given] == 0 {
		t.Error("nobody ever stole or gave; the moral layer never engaged")
	}
	// A settlement where theft is the normal way to get food has a broken
	// economy, not a moral one.
	if counts[event.Stolen] > counts[event.Fulfilled]*20 {
		t.Errorf("theft (%d) dwarfs honest work (%d)",
			counts[event.Stolen], counts[event.Fulfilled])
	}
}

// TestNormsDriftTogether checks that living side by side makes people more
// alike. Culture has to be able to form, or nothing above the individual can.
func TestNormsDriftTogether(t *testing.T) {
	w := valueWorld(32)
	a := w.Spawn("a", need.Neutral())
	b := w.SpawnAt("b", need.Neutral(), a.Pos)
	a.Norms = belief.Norms{belief.Honesty: 1, belief.Charity: 1, belief.Industry: 1, belief.Tradition: 1}
	b.Norms = belief.Norms{}
	a.AddBond(b.ID, 0.8)
	b.AddBond(a.ID, 0.8)

	before := a.Norms[belief.Honesty] - b.Norms[belief.Honesty]
	for i := 0; i < 400; i++ {
		contagion(w)
	}
	after := a.Norms[belief.Honesty] - b.Norms[belief.Honesty]

	if after >= before {
		t.Fatalf("neighbours stayed as different as ever: gap %.3f then %.3f", before, after)
	}
}

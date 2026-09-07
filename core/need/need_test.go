package need

import "testing"

func TestLowerTiersAreSteeper(t *testing.T) {
	// Same deficit everywhere, foundation irrelevant: bottom tier must be most urgent.
	l := Levels{0.7, 0.7, 0.7, 0.7, 0.7}
	// Isolate the curve by removing damping: give each tier a full foundation.
	prev := 2.0
	for _, tier := range Tiers() {
		full := Levels{1, 1, 1, 1, 1}
		full[tier] = l[tier]
		u := Urgency(tier, full)
		if u >= prev {
			t.Fatalf("tier %s urgency %.3f not below previous %.3f", tier, u, prev)
		}
		prev = u
	}
}

func TestHigherTiersAreDampedButLeak(t *testing.T) {
	starving := Levels{0, 1, 1, 1, 0}
	sated := Levels{1, 1, 1, 1, 0}
	uStarving := Urgency(Actualization, starving)
	uSated := Urgency(Actualization, sated)
	if uStarving >= uSated {
		t.Fatalf("starving agent's curiosity %.3f should be below sated agent's %.3f", uStarving, uSated)
	}
	if uStarving <= 0 {
		t.Fatalf("curiosity must leak through even when starving, got %.3f", uStarving)
	}
	empty := Levels{}
	if got := Urgency(Actualization, empty); got < Leak*0.99 {
		t.Fatalf("fully unmet agent should keep at least Leak=%.2f of top-tier urgency, got %.3f", Leak, got)
	}
}

func TestSatedTierHasNoUrgency(t *testing.T) {
	if u := Urgency(Physiological, Levels{1, 0, 0, 0, 0}); u != 0 {
		t.Fatalf("full tier should have zero urgency, got %.3f", u)
	}
}

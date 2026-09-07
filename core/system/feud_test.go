package system

import (
	"testing"

	"lreat/core/action"
	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/observe"
	"lreat/core/world"
)

// wronged sets up a comfortable victim standing beside the thief who robbed
// them, with the grudge a theft leaves behind and no values to complicate it.
func wronged(seed uint64) (*world.World, *entity.Agent, *entity.Agent) {
	w := world.New(seed)
	victim := w.Spawn("victim", need.Neutral())
	thief := w.SpawnAt("thief", need.Neutral(), victim.Pos)
	for _, a := range []*entity.Agent{victim, thief} {
		a.Needs = need.Levels{1, 0.9, 0.9, 0.5, 0.9}
		a.Norms = belief.Norms{}
		a.Inventory[entity.Food] = 3
	}
	victim.Judge(thief.ID, -0.6, w.Tick)
	return w, victim, thief
}

func TestTheWrongedGetEven(t *testing.T) {
	w, victim, _ := wronged(41)
	if d, _ := Choose(victim, w); d != action.Retaliate {
		t.Fatalf("a wronged agent with nothing better to do chose %q, want retaliate", d.Name)
	}

	// The charitable forgive, or at least refrain.
	w, saint, _ := wronged(41)
	saint.Norms[belief.Charity] = 1
	if d, _ := Choose(saint, w); d == action.Retaliate {
		t.Fatal("an agent who prizes charity took revenge")
	}
}

func TestRetaliationMakesAFeud(t *testing.T) {
	w, victim, thief := wronged(42)
	if thief.Regard(victim.ID) < 0 {
		t.Fatal("the thief should start with nothing against the victim")
	}
	action.Retaliate.Apply(victim, w)

	if thief.Regard(victim.ID) > -0.3 || victim.Regard(thief.ID) > -0.3 {
		t.Fatalf("no feud: thief thinks %.2f of victim, victim %.2f of thief",
			thief.Regard(victim.ID), victim.Regard(thief.ID))
	}
	if thief.Inventory[entity.Food] != 2 || victim.Inventory[entity.Food] != 4 {
		t.Fatalf("nothing was taken: thief %.0f victim %.0f",
			thief.Inventory[entity.Food], victim.Inventory[entity.Food])
	}
	// Now the thief holds a grudge of its own, and may answer in kind.
	if !action.Retaliate.Available(thief, w) {
		t.Fatal("the struck thief has no grudge to act on")
	}
}

func TestReprisalTeachesCaution(t *testing.T) {
	w, victim, thief := wronged(43)
	bystander := w.SpawnAt("bystander", need.Neutral(), victim.Pos)
	far := w.SpawnAt("far", need.Neutral(), entity.Pos{X: 0, Y: 0})

	action.Retaliate.Apply(victim, w)

	if thief.Caution <= bystander.Caution || bystander.Caution <= 0 {
		t.Fatalf("caution: thief %.2f bystander %.2f; the struck should learn most, the watching some",
			thief.Caution, bystander.Caution)
	}
	if far.Caution != 0 {
		t.Fatal("someone out of sight learned caution")
	}

	before := thief.Caution
	for i := 0; i < 50; i++ {
		Beliefs(w)
	}
	if thief.Caution >= before {
		t.Fatal("caution did not fade when nothing reinforced it")
	}
}

func TestCautionDetersTheft(t *testing.T) {
	// Same starving thief, same easy mark, no scruples either way. The only
	// difference is what it has learned to expect from the neighbors.
	lawless, thief, _ := desperate(t, 44)
	thief.Norms[belief.Honesty] = 0
	thief.Caution = 0
	if d, _ := Choose(thief, lawless); d != action.Steal {
		t.Fatalf("in a lawless place the thief chose %q over an easy theft", d.Name)
	}

	policed, wary, _ := desperate(t, 44)
	wary.Norms[belief.Honesty] = 0
	wary.Caution = 1
	if d, _ := Choose(wary, policed); d == action.Steal {
		t.Fatal("an agent who has seen thieves punished stole anyway")
	}
}

// TestFeudsFormInALivingSettlement runs settlements with no player and checks
// that the retaliation loop closes on its own: wrongs are answered, some
// answers are answered in turn, and the answering teaches people caution.
//
// It runs several seeds because one is a coin toss. Whether a particular pair
// falls out badly enough to keep a grudge turns on who happened to be standing
// where when a theft went down, and about half of settlements get through four
// thousand ticks without one hardening. The claim being made is that the loop
// closes in a living settlement, not that it closes in every settlement, so
// the test asks the question it means to ask.
func TestFeudsFormInALivingSettlement(t *testing.T) {
	var stolen, avenged, feuds int
	var mostCaution float64
	for seed := uint64(1); seed <= 6; seed++ {
		w := world.New(seed)
		for i := 0; i < 20; i++ {
			w.Spawn("a", w.RandomPersonality())
		}
		for i := 0; i < 4000; i++ {
			Step(w)
			for _, e := range w.Log.Since(w.Tick) {
				switch e.Kind {
				case event.Stolen:
					stolen++
				case event.Avenged:
					avenged++
				}
			}
		}
		s := observe.Take(w)
		var caution float64
		for _, a := range w.Agents {
			caution += a.Caution
		}
		if len(w.Agents) > 0 {
			caution /= float64(len(w.Agents))
		}
		if caution > mostCaution {
			mostCaution = caution
		}
		feuds += s.Feuds
		t.Logf("seed %d: pop %d, stolen %d, avenged %d, feuds %d, friendships %d, mean caution %.2f",
			seed, s.Population, stolen, avenged, s.Feuds, s.Friendships, caution)
	}

	if avenged == 0 {
		t.Fatal("nobody ever got even")
	}
	if feuds == 0 {
		t.Fatal("retaliation never hardened into a feud in any settlement")
	}
	if mostCaution == 0 {
		t.Fatal("reprisals taught nobody anything")
	}
}

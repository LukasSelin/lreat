package system

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/world"
)

// wintering puts one unroofed agent in a world and runs a day of weather on
// it, returning what the day cost its body and its condition.
func wintering(warmth float64) (fed, health float64) {
	w := world.NewSized(1, 40, 12)
	w.Mods.Warmth = warmth
	w.Climate.Temp = -10 // the bitterest cold
	a := w.Spawn("a", need.Neutral())
	a.Shelter = 0
	a.Health = 0.5
	a.Needs[need.Physiological] = 0.8
	Decay(w)
	return a.Needs[need.Physiological], a.Health
}

// What a settlement has learned to put between a body and the weather comes
// off both halves of a winter: the hunger of staying warm, and the condition
// it takes.
func TestWeavingTakesTheEdgeOffAWinter(t *testing.T) {
	bare, bareHealth := wintering(1)
	clad, cladHealth := wintering(0.6)
	if clad <= bare {
		t.Fatalf("a clad body kept %.4f of its day and a bare one %.4f; the cloak should cost less", clad, bare)
	}
	if cladHealth <= bareHealth {
		t.Fatalf("a clad body came to %.4f condition and a bare one %.4f; the cloak should spare both halves", cladHealth, bareHealth)
	}
}

// A cloak is worth nothing in June. The modifier multiplies the cold, so
// where there is no cold there is nothing to multiply, and a settlement that
// learns to weave in a mild place has learned something it cannot use.
func TestWeavingIsWorthNothingInAMildSeason(t *testing.T) {
	mild := func(warmth float64) float64 {
		w := world.NewSized(1, 40, 12)
		w.Mods.Warmth = warmth
		w.Climate.Temp = 18
		a := w.Spawn("a", need.Neutral())
		a.Shelter, a.Needs[need.Physiological] = 0, 0.8
		Decay(w)
		return a.Needs[need.Physiological]
	}
	if mild(1) != mild(0.6) {
		t.Fatal("weaving changed a summer day, and it should touch nothing but the cold")
	}
}

// Medicine speeds the climb back and not the fall. A body whose
// circumstances have improved gets there sooner; one that is failing fails
// at the rate it always did.
func TestMedicineHealsAndDoesNotHarm(t *testing.T) {
	run := func(healing, health, food, shelter float64) float64 {
		w := world.NewSized(1, 40, 12)
		w.Mods.Healing = healing
		w.Climate.Temp = 15
		a := w.Spawn("a", need.Neutral())
		a.Health, a.Shelter = health, shelter
		a.Needs[need.Physiological] = food
		Decay(w)
		return a.Health
	}
	// Well fed and housed, with a body still worn from before: recovering.
	plain := run(1, 0.2, 1, 1)
	tended := run(2, 0.2, 1, 1)
	if tended <= plain {
		t.Fatalf("a tended body came back to %.5f and an untended one to %.5f; medicine should be the quicker", tended, plain)
	}
	// Starving and unhoused, with a body still sound: failing.
	fallPlain := run(1, 0.9, 0, 0)
	fallTended := run(2, 0.9, 0, 0)
	if fallTended != fallPlain {
		t.Fatalf("medicine changed how fast a body fails, %.5f against %.5f; it should touch only the climb back", fallTended, fallPlain)
	}
}

// The deep technologies are the far end of a run and must ask for masters,
// because a master is the one tier nobody can be given. A catalog whose top
// could be reached by being taught would be a catalog every settlement holds
// entire by year fifteen, which is what this one used to be.
func TestTheDeepTechnologiesAskForMasters(t *testing.T) {
	deep := map[world.Tech]bool{"the arch": true, "medicine": true, "the plough": true}
	seen := 0
	for _, d := range Discoveries {
		if !deep[d.Tech] {
			continue
		}
		seen++
		if d.Craft == nil {
			t.Fatalf("%s has no craft, so nobody can ever be said to have mastered it", d.Tech)
		}
		w := world.NewSized(1, 40, 12)
		w.Knowledge = 10000
		for _, t := range []world.Tech{"agriculture", "masonry", "writing", "metallurgy", "quarrying", "trapping"} {
			w.Unlock(t)
		}
		// Everybody a journeyman of everything, which under the ladder is
		// as far as being taught can carry anyone.
		for i := 0; i < 12; i++ {
			a := w.Spawn("a", need.Neutral())
			for s := entity.Skill(0); s < entity.SkillCount; s++ {
				a.Skills[s] = 3*entity.TierBand - 0.01
				a.Practice[s] = 1000
			}
		}
		if d.Condition(w) {
			t.Fatalf("%s unlocked for a settlement of journeymen; the deep end must want masters", d.Tech)
		}
	}
	if seen != len(deep) {
		t.Fatalf("checked %d of the deep technologies, want %d", seen, len(deep))
	}
}

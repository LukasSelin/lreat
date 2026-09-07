package action

import (
	"math"

	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/world"
)

// Retaliation is where feuds come from. A wrong leaves a grudge in the
// victim's bond; a deep enough grudge makes getting even worth the trouble;
// getting even leaves a grudge on the other side. Whether that cycle burns
// out or becomes a feud depends on the two temperaments, on what each holds
// to be right, and on how much reprisal each has learned to expect.

func init() {
	Catalog = append(Catalog, Retaliate)
}

const (
	// GrudgeThreshold is how badly an agent must think of someone before
	// getting even becomes a thing it might do.
	GrudgeThreshold = -0.4
	// settled is how much regard a retaliation restores in the avenger. It is
	// less than a fresh wrong costs, so one act of retaliation rarely clears
	// the slate on its own.
	settled = 0.25
	// struck is how the target's regard for the avenger moves.
	struck = -0.5
)

// Grudge returns the living, reachable agent this one most wants to get even
// with, and the bond that says so. Nil when no grudge is worth acting on.
func Grudge(a *entity.Agent, w *world.World) (*entity.Agent, *entity.Bond) {
	var worst *entity.Bond
	var target *entity.Agent
	for i := range a.Bonds {
		b := &a.Bonds[i]
		if b.Regard > GrudgeThreshold {
			continue
		}
		if worst != nil && b.Regard >= worst.Regard {
			continue
		}
		o := w.Find(b.To)
		if o == nil || entity.Dist(a.Pos, o.Pos) > searchRadius {
			continue
		}
		worst, target = b, o
	}
	return target, worst
}

// deter teaches everyone nearby that wrongs are answered in this place.
func deter(a *entity.Agent, w *world.World) {
	for _, o := range w.Agents {
		if o == a || entity.Dist(a.Pos, o.Pos) > reachRadius {
			continue
		}
		o.Caution = belief.Clamp(o.Caution + belief.CautionLearned)
	}
}

// Retaliate is getting even. It restores the avenger's standing and takes
// something from the target, and it is the only act that raises Caution, so
// without it nothing in the settlement deters anything.
var Retaliate = &Def{
	Name: "retaliate", Ticks: 1,
	Available: func(a *entity.Agent, w *world.World) bool {
		t, _ := Grudge(a, w)
		return t != nil
	},
	Target: func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
		t, _ := Grudge(a, w)
		if t == nil {
			return entity.Pos{}, false
		}
		return t.Pos, true
	},
	Expect: func(a *entity.Agent, w *world.World, _ entity.Pos) need.Levels {
		_, b := Grudge(a, w)
		if b == nil {
			return need.Levels{}
		}
		depth := -b.Regard
		return need.Levels{
			need.Esteem: 0.4 * depth,  // standing restored
			need.Safety: 0.15 * depth, // they will think twice next time
		}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		t, b := Grudge(a, w)
		if t == nil || entity.Dist(a.Pos, t.Pos) > 2 {
			return // they got away, for now
		}
		depth := -b.Regard

		// Take what can be taken: food if they have it, else a fine.
		switch {
		case t.Inventory[entity.Food] >= 1:
			t.Inventory[entity.Food]--
			a.Inventory[entity.Food]++
		case t.Wealth > 0:
			fine := math.Min(t.Wealth, 2*w.Market.Price[entity.Food])
			t.Wealth -= fine
			a.Wealth += fine
		}
		t.Needs.Add(need.Safety, -0.2)
		t.Needs.Add(need.Esteem, -0.15)

		// The score is settled, for now.
		b.Regard = clampUnit(b.Regard + settled)
		a.Needs.Add(need.Esteem, 0.25*depth)

		// The target learns two things: that this person is an enemy, and
		// that wrongdoing is answered here.
		t.Judge(a.ID, struck, w.Tick)
		if bt := t.Look(a.ID); bt != nil {
			bt.Strength = belief.Clamp(bt.Strength - 0.3)
			bt.Expect = clampUnit(bt.Expect - 0.5)
		}
		t.Caution = belief.Clamp(t.Caution + belief.CautionSuffered)

		witness(a, w, "retaliate")
		deter(a, w)
		remorse(a, "retaliate")
		w.Emit(event.Avenged, a.ID, t.ID, "%s got even with %s", a.Name, t.Name)
	},
}

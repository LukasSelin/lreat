package system

import (
	"math"

	"lreat/core/action"
	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/world"
)

// Noise is the standard deviation of the jitter added to every score. It
// keeps agents with identical circumstances from marching in lockstep.
const Noise = 0.01

// Score is the utility of an expected gain for this agent right now:
// gain per tier, weighted by how urgent that tier is and how much this
// particular agent cares about it.
func Score(a *entity.Agent, gain need.Levels, urgency [need.Count]float64) float64 {
	var s float64
	for t := range gain {
		s += gain[t] * urgency[t] * a.Personality[t]
	}
	return s
}

// Choose picks the best available action for an agent and where to do it.
// Scores are divided by total time, travel included, so a quick fix nearby
// beats a slow one far away. Travel is costed over the terrain in the way,
// so the shape of the ground, not just distance, is how the map shapes
// behavior.
func Choose(a *entity.Agent, w *world.World) (*action.Def, entity.Pos) {
	urgency := need.Urgencies(a.Needs)
	best, bestPos, bestScore := action.Rest, a.Pos, math.Inf(-1)
	for _, d := range action.Catalog {
		if !d.Available(a, w) {
			continue
		}
		target, ok := d.Target(a, w)
		if !ok {
			continue
		}
		cost := float64(d.Ticks) + w.Grid.TravelCost(a.Pos, target)/a.Vigor()
		// Conscience sits beside need rather than inside it. It is not scaled
		// by urgency, so a principle holds until hunger grows big enough to
		// outweigh it, and then it gives way.
		// Expected reprisal sits beside conscience: the same act is cheap where
		// nobody answers wrongdoing and dear where somebody does.
		v := action.ValenceOf(d.Name)
		worth := Score(a, d.Expect(a, w, target), urgency) + belief.Conscience(v, a.Norms) - belief.Reprisal(v, a.Caution)
		s := worth/cost + w.RNG.NormFloat64()*Noise
		if s > bestScore {
			best, bestPos, bestScore = d, target, s
		}
	}
	return best, bestPos
}

// Decide gives every idle agent a plan.
func Decide(w *world.World) {
	for _, a := range w.Agents {
		if a.Plan != nil {
			continue
		}
		d, target := Choose(a, w)
		a.Plan = &entity.Plan{Action: d.Name, Target: target, Remaining: d.Ticks, Total: d.Ticks}
	}
}

// Exertion is the physiological cost of one tick's worth of walking, for an
// ordinary body. Hard ground is slow and tiring in the same measure: a tile
// that takes three ticks to cross also takes three ticks of hunger with it.
// A stronger frame both keeps a faster pace and spends less per tick, so the
// same errand costs a hale agent noticeably less than a worn one.
const Exertion = 0.006

// Act moves agents toward their targets, then advances and applies plans.
func Act(w *world.World) {
	for _, a := range w.Agents {
		if a.Plan == nil {
			continue
		}
		if a.Pos != a.Plan.Target {
			next := w.Grid.StepToward(a.Pos, a.Plan.Target)
			a.Travel += a.Vigor()
			a.Needs.Add(need.Physiological, -Exertion/a.Endurance())
			if cost := w.Grid.MoveCost(next); a.Travel >= cost {
				a.Travel -= cost
				a.Pos = next
			}
			continue
		}
		a.Travel = 0
		a.Plan.Remaining--
		if a.Plan.Remaining > 0 {
			continue
		}
		d := action.ByName(a.Plan.Action)
		a.Plan = nil
		// Circumstances may have changed since the plan was made: the last
		// unit of food may have been bought by someone faster. A plan that is
		// no longer possible simply fails.
		if d == nil || !d.Available(a, w) {
			continue
		}
		d.Apply(a, w)
		w.Emit(event.Acted, a.ID, 0, "%s finished %s", a.Name, d.Name)
	}
}

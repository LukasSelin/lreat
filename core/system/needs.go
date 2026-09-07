package system

import (
	"math"

	"lreat/core/need"
	"lreat/core/world"
)

// decay is the per-tick satisfaction lost for tiers that drain on their own.
// Safety is not listed: it tracks the agent's environment instead.
var decay = [need.Count]float64{
	need.Physiological: 0.02,
	need.Belonging:     0.006,
	need.Esteem:        0.004,
	need.Actualization: 0.003,
}

// HealthRate is how fast condition follows circumstance. It is slow enough
// that health is a record of how an agent has lived, not of what it ate today.
const HealthRate = 0.01

// Decay drains needs, lets shelter rot, and relaxes safety toward what the
// agent's circumstances actually provide.
func Decay(w *world.World) {
	for _, a := range w.Agents {
		for t := range decay {
			a.Needs.Add(need.Tier(t), -decay[t])
		}

		a.Shelter = math.Max(0, a.Shelter-w.Mods.ShelterDecay)

		// Health follows nourishment and housing, at a hundredth of the rate
		// they move themselves. A lean week barely shows; a lean season
		// leaves a body that walks slower and tires sooner.
		condition := 0.35 + 0.45*need.Clamp(a.Needs[need.Physiological]) + 0.2*a.Shelter
		a.Health = need.Clamp(a.Health + (condition-a.Health)*HealthRate)

		// Safety is derived: housing, public order, and savings each matter.
		target := 0.5*a.Shelter + 0.35*w.Safety + 0.15*math.Min(a.Wealth/20, 1)
		a.Needs[need.Safety] += (target - a.Needs[need.Safety]) * 0.1

		if a.Needs[need.Physiological] <= 0.02 {
			a.Starving++
		} else {
			a.Starving = 0
		}
	}
	w.Safety = math.Max(0, w.Safety-0.01)
}

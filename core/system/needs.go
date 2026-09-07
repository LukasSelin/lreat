package system

import (
	"math"

	"lreat/core/entity"
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

// ColdDrain is what the bitterest cold costs a body per tick with no roof
// over it, on top of the ordinary drain. A roof takes all of it: what makes
// winter dangerous is not the cold but being caught out in it, and a
// settlement that has built for itself hardly feels the season in its
// larder. A seventh of the ordinary drain is enough to make an unhoused
// winter tell; at half again the ordinary drain, which is what it was first
// set to, winter did not press on the unhoused, it pressed on the
// settlement, and the population fell with it. Nobody is fully roofed all
// the time - shelter rots - so this is a cost everyone pays a little of.
const ColdDrain = 0.003

// ColdCondition is how much of a body's condition the same exposure takes.
// It works through health, which moves at a hundredth of the rate needs do,
// so a winter outdoors shows up as a body that is still slower come spring.
const ColdCondition = 0.2

// Household is a parent's roof reaching over its own young. Shelter is kept
// per body, and a house is not: a roof raised by one person covers everybody
// asleep under it, and the people asleep under it are that person's children.
// Without this a five-year-old holds whatever shelter it was born with and
// watches it rot, for the ten years before it can carry a length of timber -
// which is not a settlement failing to house its children, it is a
// settlement in which houses do not have families in them.
//
// It is not an act, and it is the only part of rearing that is not: feeding
// and teaching are things a parent does or fails to do, and a roof is a
// thing that stands. Nobody decides each morning to go on sheltering their
// children. The roof is worth what the parent's is, never more, so a
// household is exactly as well housed as the person who built it.
func Household(w *world.World) {
	roofs := make(map[entity.ID]float64, len(w.Agents))
	for _, a := range w.Agents {
		roofs[a.ID] = a.Shelter
	}
	for _, a := range w.Agents {
		if a.Parent == 0 || entity.Adult(a.Age(w.Tick)) {
			continue
		}
		if roof, ok := roofs[a.Parent]; ok && roof > a.Shelter {
			a.Shelter = roof
		}
	}
}

// Decay drains needs, lets shelter rot, and relaxes safety toward what the
// agent's circumstances actually provide.
func Decay(w *world.World) {
	chill := w.Climate.Chill()
	for _, a := range w.Agents {
		for t := range decay {
			a.Needs.Add(need.Tier(t), -decay[t])
		}

		a.Shelter = math.Max(0, a.Shelter-w.Mods.ShelterDecay)

		// What is carried rots as what is stored does. An agent's larder is
		// its own: no granary stands behind it, and it keeps by the weather
		// and by the roof over it.
		Spoil(w, a)

		// What the weather costs is what the agent stands in it unroofed.
		// In a mild season it is nothing whatever anyone has built; in the
		// deep of a hard winter a body without a house burns half again
		// what it otherwise would just staying warm.
		exposure := chill * (1 - a.Shelter)
		a.Needs.Add(need.Physiological, -ColdDrain*exposure)

		// Health follows nourishment and housing, at a hundredth of the rate
		// they move themselves. A lean week barely shows; a lean season
		// leaves a body that walks slower and tires sooner.
		condition := 0.35 + 0.45*need.Clamp(a.Needs[need.Physiological]) + 0.2*a.Shelter - ColdCondition*exposure
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

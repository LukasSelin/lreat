package system

import (
	"math"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/ontology"
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

// roofWear is what a roof loses per tick, read from the ontology once at
// start: the settlement's building knowledge scales it, and the ontology
// says what it scales.
var roofWear = func() float64 {
	t, ok := ontology.Spoiling(ontology.Dwelling, ontology.Person)
	if !ok {
		panic("system: nothing wears a roof")
	}
	return t.Rate
}()

// Decay drains needs, lets shelter rot, and relaxes safety toward what the
// agent's circumstances actually provide.
//
// Each body is worn by its own day, and by nothing anybody else's day
// does: what one agent's needs, roof and health come to reads the weather
// where it stands and the settlement's order and knowledge, which the day
// does not change until everybody has been worn, and draws no chance. So
// everybody is worn at once, over goroutines, and the settlement's order
// fades afterwards.
func Decay(w *world.World) {
	n := len(w.Agents)
	world.InParallel(n, world.WorkersOver(n), func(i, _ int) { wear(w.Agents[i], w) })
	w.Safety = math.Max(0, w.Safety-0.01)
}

// wear is one agent's day of Decay.
func wear(a *entity.Agent, w *world.World) {
	for t := range decay {
		d := decay[t]
		if need.Tier(t) == need.Physiological {
			// What a body spends staying alive is the body's own. The
			// rest of the tiers are wants and drain alike for everyone;
			// this one is a fire that some people bank higher.
			d *= a.Body.Burn()
		}
		a.Needs.Add(need.Tier(t), -d)
	}

	a.Shelter = math.Max(0, a.Shelter-roofWear*w.Mods.ShelterDecay)

	// What is carried rots as what is stored does. An agent's larder is
	// its own: no granary stands behind it, and it keeps by the weather
	// and by the roof over it.
	Spoil(w, a)

	// What the weather costs is what the agent stands in it unroofed.
	// In a mild season it is nothing whatever anyone has built; in the
	// deep of a hard winter a body without a house burns half again
	// what it otherwise would just staying warm.
	// What the cold costs is what the body standing in it can stand.
	// One exposure carries both halves of a winter - the hunger of
	// staying warm below, and the condition it takes further down - so
	// a hardy frame is spared both and a frail one pays both.
	// What the settlement has learned to put between a body and the
	// weather comes off here, on the one line both halves of a winter
	// are read from.
	exposure := w.ChillAt(a.Pos) * (1 - a.Shelter) / a.Body.Hardy() * w.Mods.Warmth
	a.Needs.Add(need.Physiological, -ColdDrain*exposure)

	// Health follows nourishment and housing, at a hundredth of the rate
	// they move themselves. A lean week barely shows; a lean season
	// leaves a body that walks slower and tires sooner.
	condition := 0.35 + 0.45*need.Clamp(a.Needs[need.Physiological]) + 0.2*a.Shelter - ColdCondition*exposure
	// What is known about tending a body speeds the climb back and not
	// the fall: a settlement that has learned medicine gets its people
	// on their feet sooner, and does not lose them any faster for it.
	rate := HealthRate
	if condition > a.Health {
		rate *= w.Mods.Healing
	}
	a.Health = need.Clamp(a.Health + (condition-a.Health)*rate)

	// Safety is derived: housing, public order, and savings each matter.
	target := 0.5*a.Shelter + 0.35*w.Safety + 0.15*math.Min(a.Wealth/20, 1)
	a.Needs[need.Safety] += (target - a.Needs[need.Safety]) * 0.1

	if a.Needs[need.Physiological] <= 0.02 {
		a.Starving++
	} else {
		a.Starving = 0
	}
}

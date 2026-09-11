package system

import (
	"math"

	"lreat/core/action"
	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// What each tier loses on its own each day is the species' own, in
// entity.Species.Decay: a person's is what was always written here, and
// safety is not among them for anybody, since it tracks circumstances
// instead.

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
	// Frozen while the bodies are worn over goroutines: a creature's day
	// asks the agent file who is standing near it, and the file must not
	// be tidying itself under the asking.
	w.Freeze(true)
	world.InParallel(n, world.WorkersOver(n), func(i, _ int) { wear(w.Agents[i], w) })
	w.Freeze(false)
	w.Safety = math.Max(0, w.Safety-0.01)
}

// wear is one agent's day of Decay.
func wear(a *entity.Agent, w *world.World) {
	sp := a.Species()
	decay := sp.Decay
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
	if !sp.Settles {
		weather(a, w)
		return
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

	starve(a)
}

// Fright is how fast a creature's safety falls when somebody is near, and
// Settling how fast it comes back once nobody is. A person's safety follows
// a roof and the settlement's order at a tenth a day, which is the pace
// those things change at; a deer's fear is the sight of somebody and is
// felt at once, or the deer is eaten before it has finished feeling it. The
// two rates are what make the alarmed moment its own moment rather than a
// slow lean on the calm one.
const (
	Fright   = 0.5
	Settling = 0.1
)

// HerdComfort is what a day in the herd is worth to a creature's belonging.
// A person's belonging is fed by going to see somebody; a deer's is fed by
// standing where the others stand, so it is a little more than the tier
// drains, and a deer that has lost its herd feels it within the month.
const HerdComfort = 0.012

// weather is a creature's day of Decay past the draining of its wants: the
// same winter a person has, read against the trees over it rather than a
// roof, and a safety that is the cover it stands in and whether anybody is
// near. It has no roof to rot, no pack to spoil, no purse and no settlement
// keeping order for it, and nothing anybody has learned comes between it
// and the cold.
func weather(a *entity.Agent, w *world.World) {
	cover := action.Cover(a, w)
	exposure := w.ChillAt(a.Pos) * (1 - cover) / a.Body.Hardy()
	a.Needs.Add(need.Physiological, -ColdDrain*exposure)
	condition := 0.35 + 0.45*need.Clamp(a.Needs[need.Physiological]) + 0.2*cover - ColdCondition*exposure
	a.Health = need.Clamp(a.Health + (condition-a.Health)*HealthRate)
	// Safe is under trees with nobody about: wholly so, and less than half
	// so in the open. Somebody about is not safe at all, whatever the
	// cover, and is felt at once.
	target, rate := 0.4+0.6*cover, Settling
	if action.Alarm(a, w) != nil {
		target, rate = 0, Fright
	}
	a.Needs[need.Safety] += (target - a.Needs[need.Safety]) * rate
	if action.InHerd(a, w) {
		a.Needs.Add(need.Belonging, HerdComfort)
	}
	starve(a)
}

// starve counts the days at the bottom of the physiological tier.
func starve(a *entity.Agent) {
	if a.Needs[need.Physiological] <= 0.02 {
		a.Starving++
	} else {
		a.Starving = 0
	}
}

package entity

import (
	"math"

	"lreat/core/clock"
	"lreat/core/need"
)

// What kind of creature an agent is.
//
// Everything an agent does runs on one psychology: needs that drain and
// press, a moment read as a point in habit space, an act recognised as
// belonging to that moment, a plan walked and done. A species is not a
// second psychology. It is what a kind of creature brings to that one:
// how long it lives and when it bears, the body and the cast of mind it
// hands its young, what it wants at birth and how fast each want returns,
// and whether it is the kind of creature that settles - holds values, learns
// crafts, keeps a house and a purse, asks things of its neighbours - or the
// kind that only lives on the ground.
//
// What a species senses and what it can do are not here, because this
// package cannot see the world or the catalog. They are registered against
// the species by the packages that can: core/action keeps a sense and an
// actor per species, core/system a way of wearing and bearing. Adding a
// creature is a value here, a few schemas in the ontology, and a sense.
//
// A species is compared by identity. There is one Human and one Deer, and an
// agent points at whichever it is; an agent that points at nothing is a
// person, so that an agent put together by hand in a test is what it always
// was.

// Life is the span of one, in ticks. A creature grows into its body by
// Maturity, carries it whole until Prime, and declines from there; Lifespan
// is the age by which a body has usually failed, not a wall it hits.
type Life struct {
	Maturity, Prime, Lifespan int
}

// Factor is how much of its vitality a body of this age can bring to bear.
// A child has half a body and grows into the rest; an elder gives it back.
// Nothing here is a cliff: the young are slower rather than helpless, and
// the old are frail rather than finished.
func (l Life) Factor(age int) float64 {
	switch {
	case age < 0:
		return 1
	case age < l.Maturity:
		return 0.5 + 0.5*float64(age)/float64(l.Maturity)
	case age < l.Prime:
		return 1
	}
	decline := float64(age-l.Prime) / float64(l.Lifespan-l.Prime)
	return math.Max(0.25, 1-0.75*decline)
}

// endRisk is the daily chance a body right at the end of its decline gives
// out. It is set against the length of the decline rather than chosen: what
// has to hold is how much of a decline a body is likely to survive, and that
// is the risk multiplied by the days there are to run it in. Lengthen the
// year and this falls, or old age arrives the same number of days after a
// prime that is now decades away.
func (l Life) endRisk() float64 { return 9.6 / float64(l.Lifespan-l.Prime) }

// Frailty is the daily chance that a body past its prime simply fails. It is
// zero until Prime and then climbs quadratically, so old age is a rising
// risk rather than an appointment: most die somewhere in their decline and a
// few live well past Lifespan.
func (l Life) Frailty(age int) float64 {
	if age < l.Prime {
		return 0
	}
	d := float64(age-l.Prime) / float64(l.Lifespan-l.Prime)
	return l.endRisk() * d * d
}

// Adult reports whether one of this age has grown into its own body.
func (l Life) Adult(age int) bool { return age >= l.Maturity }

// Fertile reports whether one of this age can bear. Bearing belongs to the
// years between growing up and starting to decline.
func (l Life) Fertile(age int) bool { return age >= l.Maturity && age < l.Prime }

// Species is a kind of creature: what it brings to the one psychology.
type Species struct {
	Name string
	Life Life
	// Body and Mind are the kind's own measure, what an ordinary one of them
	// is; an individual is drawn around these and hands its own on with
	// drift. See trait.go.
	Body Body
	Mind Mind
	// Needs is how met each tier is at birth, and Decay how much of each
	// tier drains on its own each day. A tier a species never feels is
	// born met and never drains, and so is never urgent: that is how a
	// creature with no standing to win is spared the want of it.
	Needs need.Levels
	Decay [need.Count]float64
	// Bears is the daily chance one whose lower tiers are met has young.
	Bears float64
	// Swims says whether this kind takes to deep water when it is carrying
	// nothing. A person does, and is kept out of it only by an armful; a
	// deer keeps to the bank, so a river is the edge of its country and a
	// herd on one side of it is not the herd on the other.
	Swims bool
	// Wary is how many tiles off a person is noticed, and Herds how near
	// another of the kind must stand to be company. Both are read by the
	// creature's senses and its acts, and neither may reach past
	// world.NearbyLimit; a person has neither, since a person's company and
	// order are read otherwise.
	Wary, Herds int
	// Edge is what standing in the open beside a wood is worth to this
	// kind as cover, in [0,1]: nothing to a deer, which is as plain in a
	// field beside the trees as in the middle of it, and most of a wood to
	// a hare, which lives in the hedge.
	Edge float64
	// Settles says whether this is the kind of creature that holds values,
	// learns crafts, keeps a house and a purse, trades, asks and is asked,
	// and discovers things: the whole of a settlement's life above the
	// ground. A creature that does not settle still wants, senses, chooses
	// and walks; it is simply not on the settlement's books.
	Settles bool
}

// Human is a person, and everything a person always was: the figures the
// settlements are made of, with the childhood, the bearing years and the
// decline that were written in age.go before there was anything else.
var Human = &Species{
	Name:  "human",
	Life:  Life{Maturity: Maturity, Prime: Prime, Lifespan: Lifespan},
	Body:  Body{Vitality: 1, Metabolism: 1, Hardiness: 1},
	Mind:  Mind{Plasticity: 1, Resolve: 1, Horizon: 1},
	Needs: need.Levels{0.7, 0.3, 0.5, 0.3, 0.2},
	// Safety is not listed: it tracks the agent's circumstances instead.
	Decay:   [need.Count]float64{need.Physiological: 0.02, need.Belonging: 0.006, need.Esteem: 0.004, need.Actualization: 0.003},
	Bears:   1.0 / clock.Year,
	Swims:   true,
	Settles: true,
}

// The creatures. Each feels hunger, danger and the want of its own kind, and
// a little of the pull to range; none has standing to win, so that tier is
// born met and never drains. What tells them apart is the body, the span,
// how often they bear, how far off they notice a person, and what they eat,
// which is on their acts.

// Deer is the first creature that is not a person: quick, hardy, slow to
// take to anything, a dozen years long, and shy of anybody six tiles off.
var Deer = &Species{
	Name:  "deer",
	Life:  Life{Maturity: 3 * clock.Year / 2, Prime: 8 * clock.Year, Lifespan: 14 * clock.Year},
	Body:  Body{Vitality: 1.6, Metabolism: 1.2, Hardiness: 1.5},
	Mind:  Mind{Plasticity: 0.3, Resolve: 1, Horizon: 1},
	Needs: need.Levels{0.7, 0.5, 0.5, 1, 0.5},
	Decay: [need.Count]float64{need.Physiological: 0.02, need.Belonging: 0.006, need.Actualization: 0.001},
	Bears: 0.5 / clock.Year,
	Wary:  6,
	Herds: 8,
}

// Boar is heavier and bolder: it burns more, stands the cold, lets a person
// come nearer before it bolts, and keeps a tighter sounder. It lives on
// what a deer lives on and on the fields besides.
var Boar = &Species{
	Name:  "boar",
	Life:  Life{Maturity: clock.Year, Prime: 6 * clock.Year, Lifespan: 10 * clock.Year},
	Body:  Body{Vitality: 1.2, Metabolism: 1.3, Hardiness: 1.3},
	Mind:  Mind{Plasticity: 0.3, Resolve: 1, Horizon: 1},
	Needs: need.Levels{0.7, 0.5, 0.5, 1, 0.5},
	Decay: [need.Count]float64{need.Physiological: 0.02, need.Belonging: 0.006, need.Actualization: 0.001},
	Bears: 0.6 / clock.Year,
	Wary:  4,
	Herds: 6,
}

// Hare is the quick, short-lived, many-bearing thing of the open ground: it
// grazes the sward at the edge of the wood, lives a few years, and bears
// several times a year.
var Hare = &Species{
	Name:  "hare",
	Life:  Life{Maturity: clock.Year / 2, Prime: 3 * clock.Year, Lifespan: 5 * clock.Year},
	Body:  Body{Vitality: 1.8, Metabolism: 1, Hardiness: 1},
	Mind:  Mind{Plasticity: 0.2, Resolve: 1, Horizon: 0.6},
	Needs: need.Levels{0.7, 0.5, 0.5, 1, 0.5},
	Decay: [need.Count]float64{need.Physiological: 0.02, need.Belonging: 0.006, need.Actualization: 0.001},
	Bears: 1.5 / clock.Year,
	Wary:  5,
	Herds: 4,
	Edge:  0.8,
}

// Creatures is every kind that is not a person, in the order they were
// thought of, which is the order they are put down in and shown in.
var Creatures = []*Species{Deer, Boar, Hare}

// Species is what kind of creature this agent is. One that was never told is
// a person, which is what a test that builds an agent by hand means.
func (a *Agent) Species() *Species {
	if a.Kind == nil {
		return Human
	}
	return a.Kind
}

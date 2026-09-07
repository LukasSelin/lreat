package action

import (
	"sort"

	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/world"
)

// The situation an agent finds itself in, as a point in habit space. Most of
// it is the same whichever action is being weighed; the last three
// dimensions depend on the candidate, because how far one must go, whom one
// would be dealing with, and how able one believes oneself all do.

// bipolar maps [0,1] onto [-1,1], clamping outside it.
func bipolar(x float64) float64 { return 2*need.Clamp(x) - 1 }

// Knees are the amounts of each stock at which its dimension saturates.
// They match the knees the value-based catalog already uses: food is worth
// nothing past four units, a house's frame is as much wood as anybody needs
// at once, twenty coins is rich.
//
// The wood knee has to be the frame's price and not an errand's. Read against
// an armful, a person holding two lengths already feels flush, and feeling
// flush is what stops them going back to the woods: they would spend the
// afternoon paving or whittling instead, and never once in a life stand in
// front of enough timber to raise a wall. Read against the frame, wood stays
// something to be short of until there is a house's worth of it, which is
// what makes gathering toward a house a thing an agent will keep at for days.
const (
	foodKnee   = 4
	woodKnee   = raisingTimber
	wealthKnee = 20
	// nearKnee is the travel cost at which a target reads as far as can be.
	nearKnee = 30
)

// Shared is the part of the situation that is the same for every candidate.
func Shared(a *entity.Agent, w *world.World) habit.Signature {
	var s habit.Signature
	u := need.Urgencies(a.Needs)
	for t := range u {
		s[t] = bipolar(u[t] * a.Personality[t])
	}
	s[habit.Food] = bipolar(a.Inventory[entity.Food] / foodKnee)
	s[habit.Wood] = bipolar(a.Inventory[entity.Wood] / woodKnee)
	s[habit.Wealth] = bipolar(a.Wealth / wealthKnee)
	// Shelter is read as a lack, not around a midpoint: half a house is
	// still half a house short, and a roof that is not kept up rots to
	// nothing. Read around a midpoint the unsheltered moment only began once
	// a house had mostly rotted, safety sat below what a birth needs except
	// in the spike after each rebuild, and a settlement lived or died on
	// how many such spikes fell in its founders' fertile years. Over 48
	// seeds this took survivors from 35 to 40 and extinctions from 2 to 0.
	s[habit.Shelter] = a.Shelter - 1
	if w.Neighbor(a, reachRadius) != nil {
		s[habit.Company] = 1
	} else {
		s[habit.Company] = -1
	}
	// The weather, read as cold and one-sided: 0 through the mild half of
	// the year, rising to 1 at the bottom of a hard winter. It is the one
	// coordinate that says the same thing to everyone at once, which for
	// any other coordinate would be a tax on whichever acts mention it -
	// but this one is silent for half the year and turns with the rest, so
	// what it taxes is farming in February and nothing in June. Read
	// bipolar it was never silent: a mild spring read as a strong -1, every
	// act that named the cold was penalised the year round, and farming,
	// alone in naming the warmth, outranked studying for the curious and
	// getting even for the wronged.
	s[habit.Chill] = need.Clamp(w.Climate.Chill())
	// Exposure is the cold this particular body is actually in: the weather
	// times what it is not sheltered from. It is the same product the world
	// charges a body for standing out in the winter, and it is here because
	// a prior is linear and so cannot say "cold and unroofed" with the
	// weather and the roof as separate coordinates - it can only say "cold"
	// and "unroofed" and add them. The difference matters. Told only that
	// it is cold, a settlement sends everybody to the woods in February,
	// the people who already have roofs included, and stops farming to do
	// it; told what each body is actually suffering, the roofed keep to
	// their fields and the ones out in it go for timber. One is a season
	// the settlement endures together, the other is the season sorting out
	// who needs to do something about it.
	s[habit.Exposure] = need.Clamp(w.Climate.Chill() * (1 - a.Shelter))
	// The moral coordinates are one-sided, as the belief layer defines them:
	// a norm of 0 is holding nothing, caution of 0 is having learned of no
	// reprisal, safety of 0 is nobody keeping order. Read that way an
	// ordinary agent minds a wrong about half as much as a saint, which is
	// what conscience in the value rule charges too. Centred at 0.5 they
	// would fall silent for everyone but the extremes.
	s[habit.Order] = need.Clamp(w.Safety)
	for n := range a.Norms {
		s[habit.Honesty+n] = need.Clamp(a.Norms[n])
	}
	s[habit.Caution] = need.Clamp(a.Caution)
	return s
}

// Situation completes shared for candidate d done at target.
func Situation(a *entity.Agent, w *world.World, d *Def, target entity.Pos, shared habit.Signature) habit.Signature {
	return SituationOn(a, w, w.Routers(1)[0], d, target, shared)
}

// SituationOn is Situation costing the walk on a given router, so that agents
// sizing up a moment at the same time each route on working memory of their
// own.
func SituationOn(a *entity.Agent, w *world.World, r *world.Router, d *Def, target entity.Pos, shared habit.Signature) habit.Signature {
	s := shared
	// Costed with what the agent is holding, so that a target it would have
	// to swim to with an armful reads as far off as the walk round is.
	cost := r.Carrying(a.Load()).TravelCost(a.Pos, target) / a.Vigor(w.Tick)
	s[habit.Near] = 1 - 2*need.Clamp(cost/nearKnee)
	if d.With != nil {
		if o := d.With(a, w, target); o != nil {
			s[habit.Rapport] = rapport(a, w, d, o)
		}
	}
	if d.Skilled != nil {
		if sk, ok := d.Skilled(a, w); ok {
			s[habit.Skill] = bipolar(a.Believes(sk))
		}
	}
	return s
}

// rapport is how the agent feels about the person an action involves. For
// most acts it is how it expects a meeting to go. For getting even it is the
// depth of the grudge, so that a deep grudge reads as a strong fit.
func rapport(a *entity.Agent, w *world.World, d *Def, o *entity.Agent) float64 {
	if d == Retaliate {
		return clampUnit(-a.Regard(o.ID))
	}
	return Anticipate(a, o)
}

// Fit is how well a candidate's situation matches the agent's habit for it,
// less how far out of reach the action still is.
func Fit(a *entity.Agent, i int, s habit.Signature) float64 {
	return habit.Cosine(s, a.Habits[i]) - habit.ReachPenalty*(1-a.Reach[i])
}

// Candidate is one action the agent could take now, with everything the
// chooser and the learner need to know about it.
type Candidate struct {
	Def       *Def
	Index     int
	Target    entity.Pos
	Situation habit.Signature
	Fit       float64
}

// Candidates lists every action available to the agent, in catalog order,
// with its situation and fit. Targets are chosen the same way value-based
// choice chooses them, including any randomness that involves.
func Candidates(a *entity.Agent, w *world.World) []Candidate {
	return CandidatesOn(a, w, w.Routers(1)[0])
}

// CandidatesOn is Candidates routing on a given router. Everything it writes
// belongs to the agent it is called for - its habits and reach, seeded on
// first use - so several agents may be sized up at once, one per router.
func CandidatesOn(a *entity.Agent, w *world.World, r *world.Router) []Candidate {
	Imprint(a)
	w.Room()
	for i := range Catalog {
		a.Reach[i] = max(a.Reach[i], w.ReachFloor[i])
	}
	shared := Shared(a, w)
	out := make([]Candidate, 0, Count)
	for i, d := range Catalog {
		if !d.Available(a, w) {
			continue
		}
		target, ok := d.Target(a, w)
		if !ok {
			continue
		}
		s := SituationOn(a, w, r, d, target, shared)
		out = append(out, Candidate{Def: d, Index: i, Target: target, Situation: s, Fit: Fit(a, i, s)})
	}
	return out
}

// Rank is Candidates sorted by fit, best first. Ties keep catalog order.
func Rank(a *entity.Agent, w *world.World) []Candidate {
	c := Candidates(a, w)
	sort.SliceStable(c, func(i, j int) bool { return c[i].Fit > c[j].Fit })
	return c
}

// Imprint seeds an agent's habits and reach from the catalog priors, once
// per slot. It is lazy because the world cannot see the catalog when it
// spawns, and a newborn's first decision is the earliest the habits are
// needed. Agents whose habits were inherited or copied are marked
// imprinted by whoever gave them. A slot given after that is fresh for
// everyone alive and is seeded here the next time each of them decides.
func Imprint(a *entity.Agent) {
	a.Room()
	for i := a.Seeded; i < len(Catalog); i++ {
		a.Habits[i] = Catalog[i].Prior
		a.Reach[i] = Catalog[i].Reach0
	}
	a.Seeded = len(Catalog)
	a.Imprinted = true
}

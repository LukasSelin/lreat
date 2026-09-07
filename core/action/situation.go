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

// nearKnee is the travel cost at which a target reads as far as can be.
// It is the ground people live on, not the width of the map. Read at
// thirty an errand three times as far as another read as only a little
// worse, and agents spent their lives walking; read at six, everything
// past the near ground reads as far as can be and the choice between two
// errands is decided where it is actually made. It is the most valuable
// number in the file: on its own it took settlements that replaced their
// founders from 39 in 48 to 46, and with the food knee beside it the
// median settlement roughly doubled. The store knees are in store.go.
const nearKnee = 6

// Shared is the part of the situation that is the same for every candidate.
func Shared(a *entity.Agent, w *world.World) habit.Signature {
	var s habit.Signature
	u := need.Urgencies(a.Needs)
	for t := range u {
		s[t] = bipolar(u[t] * a.Personality[t])
	}
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
	// Public order is read as a lack, the way shelter is: order of 0 is
	// nobody keeping any, and that is the loudest the coordinate ever gets.
	// Read one-sided it was silent exactly when it mattered - an act whose
	// moment is the ungoverned one, which is standing guard, could say so
	// only by naming a coordinate that reads 0 in an ungoverned settlement -
	// and nobody ever stood a watch in a place that had never had one.
	s[habit.Order] = need.Clamp(w.Safety) - 1
	// The moral coordinates are one-sided, as the belief layer defines them:
	// a norm of 0 is holding nothing and caution of 0 is having learned of no
	// reprisal. Read that way an ordinary agent minds a wrong about half as
	// much as a saint, which is what conscience in the value rule charges
	// too. Centred at 0.5 they would fall silent for everyone but the
	// extremes.
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
	// What the act would bring and spend, or, for an act that moves
	// nothing in particular, how the agent stands in general: every
	// candidate's moment has the same shape, so none is nearer for saying
	// less.
	if d.Supply != nil {
		s[habit.Lack], s[habit.Stock] = d.Supply(a)
	} else {
		p := plenty(a)
		s[habit.Lack], s[habit.Stock] = -p, p
	}
	// What the act would bring and spend, or, for an act that moves
	// nothing in particular, how the agent stands in general: every
	// candidate's moment has the same shape, so none is nearer for saying
	// less.
	if d.Supply != nil {
		s[habit.Lack], s[habit.Stock] = d.Supply(a)
	} else {
		p := plenty(a)
		s[habit.Lack], s[habit.Stock] = -p, p
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
	// The store coordinates count toward how well the moment matches the
	// habit, not toward how big the moment is. They are read per
	// candidate, and a candidate is not further from its habit for being
	// about something the agent is short of: with them in the norm, an
	// act that moves nothing had the longer moment and lost on cosine to
	// one that does, whatever either was about.
	h := a.Habits[i]
	shape := s
	shape[habit.Lack], shape[habit.Stock] = 0, 0
	ns, nh := habit.Norm(shape), habit.Norm(h)
	if ns < habit.Epsilon || nh < habit.Epsilon {
		return -habit.ReachPenalty * (1 - a.Reach[i])
	}
	return max(-1, min(1, habit.Dot(s, h)/(ns*nh))) - habit.ReachPenalty*(1-a.Reach[i])
}

// Candidate is one action the agent could take now, with everything the
// chooser needs to know about it.
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
	// One spread of the ground around the agent serves every errand it is
	// about to weigh: they all start where it stands, and none of them is
	// read past the near knee.
	r.Survey(a.Pos, a.Load(), nearKnee*a.Vigor(w.Tick))
	defer r.Forget()
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
// per slot, with a little idiosyncrasy on the varying coordinates. It is
// lazy because the world cannot see the catalog when it spawns, and a
// newborn's first decision is the earliest the habits are needed. Agents
// whose habits were inherited or copied are marked imprinted by whoever
// gave them. A slot given after that is fresh for everyone alive and is
// seeded here the next time each of them decides.
//
// The idiosyncrasy is what makes a settlement more than twenty copies of one
// person. Founders seeded from the bare priors all recognise the same moment
// as calling for the same act, so they forage together, build together, and
// go hungry together; there is no division of labour to be had, because
// there is nobody who reads a morning differently. Drawn from the agent's own
// luck, so it is fixed at birth and the same in any run of the same seed.
func Imprint(a *entity.Agent) {
	a.Room()
	for i := a.Seeded; i < len(Catalog); i++ {
		prior := Catalog[i].Prior
		a.Habits[i] = prior
		if a.Luck != nil && BornNoise > 0 && habit.Norm(prior) >= habit.Epsilon {
			for k := range a.Habits[i] {
				if habit.Varying[k] {
					a.Habits[i][k] += a.Luck.NormFloat64() * BornNoise
				}
			}
			habit.ClampNorm(&a.Habits[i], habit.MinNorm, habit.MaxNorm)
		}
		a.Reach[i] = Catalog[i].Reach0
	}
	a.Seeded = len(Catalog)
	a.Imprinted = true
}

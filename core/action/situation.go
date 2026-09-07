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
// nothing past four units, two wood build a shelter, twenty coins is rich.
const (
	foodKnee   = 4
	woodKnee   = 2
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
	s[habit.Shelter] = bipolar(a.Shelter)
	if w.Neighbor(a, reachRadius) != nil {
		s[habit.Company] = 1
	} else {
		s[habit.Company] = -1
	}
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
	s := shared
	cost := w.Grid.TravelCost(a.Pos, target) / a.Vigor(w.Tick)
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
	Imprint(a)
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
		s := Situation(a, w, d, target, shared)
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

// Imprint seeds an agent's habits and reach from the catalog priors, once.
// It is lazy because the world cannot see the catalog when it spawns, and
// a newborn's first decision is the earliest the habits are needed. Agents
// whose habits were inherited or copied are marked imprinted by whoever
// gave them.
func Imprint(a *entity.Agent) {
	if a.Imprinted {
		return
	}
	for i, d := range Catalog {
		a.Habits[i] = d.Prior
		a.Reach[i] = d.Reach0
	}
	a.Imprinted = true
}

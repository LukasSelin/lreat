package system

import (
	"math"
	"runtime"
	"sync"

	"lreat/core/action"
	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/habit"
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
// beats a slow one far away. Travel is costed over the ground in the way, on
// the route the agent would really walk, so the shape of the land and what
// has been built on it — a river across the way, a street running to the
// market — is how the map shapes behavior.
func Choose(a *entity.Agent, w *world.World) (*action.Def, entity.Pos) {
	return choose(a, w, w.Routers(1)[0])
}

// choose is Choose over a given router, so that agents deciding at the same
// time each route on working memory of their own.
func choose(a *entity.Agent, w *world.World, r *world.Router) (*action.Def, entity.Pos) {
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
		cost := float64(d.Ticks) + r.TravelCost(a.Pos, target)/a.Vigor(w.Tick)
		// Conscience sits beside need rather than inside it. It is not scaled
		// by urgency, so a principle holds until hunger grows big enough to
		// outweigh it, and then it gives way.
		// Expected reprisal sits beside conscience: the same act is cheap where
		// nobody answers wrongdoing and dear where somebody does.
		v := action.ValenceOf(d.Name)
		worth := Score(a, d.Expect(a, w, target), urgency) + belief.Conscience(v, a.Norms) - belief.Reprisal(v, a.Caution)
		s := worth/cost + a.Luck.NormFloat64()*Noise
		if s > bestScore {
			best, bestPos, bestScore = d, target, s
		}
	}
	return best, bestPos
}

// Workers is how many goroutines deciding may spread over. Deciding is the
// bulk of a tick - most of it is agents sizing up errands over the ground -
// and it only reads the world, so it is the one phase that parallelises. Set
// it to 1 to decide one agent at a time.
var Workers = runtime.NumCPU()

// decision is what one agent worked out for itself, held aside until every
// agent has finished. Nothing here touches the world; it is all installed
// afterwards, in agent order.
type decision struct {
	plan    *entity.Plan
	entropy float64
	counted bool
}

// Decide gives every idle agent a plan, by value or by fit as the world's
// rules say. Agents decide at the same time and commit in turn: deciding only
// reads the world and the agent's own habits, so it can be spread over
// goroutines, while the plans - and the tally of how undecided the settlement
// was - land in a fixed order out of a slice indexed by agent. That is what
// keeps a run the same however the goroutines are scheduled.
func Decide(w *world.World) {
	w.Choices, w.Entropy = 0, 0
	idle := make([]*entity.Agent, 0, len(w.Agents))
	for _, a := range w.Agents {
		if a.Plan == nil {
			idle = append(idle, a)
		}
	}
	if len(idle) == 0 {
		return
	}
	out := make([]decision, len(idle))
	routers := w.Routers(workersFor(len(idle)))
	inParallel(len(idle), len(routers), func(i, worker int) {
		out[i] = decide(idle[i], w, routers[worker])
	})
	for i, a := range idle {
		if out[i].plan == nil {
			continue
		}
		a.Plan = out[i].plan
		if out[i].counted {
			w.Choices++
			w.Entropy += out[i].entropy
		}
	}
}

// decide works out one agent's next plan without touching anything outside
// that agent. The way to the target is worked out here too, while there are
// goroutines to work it out on: it is the last of the routing that used to
// happen a step at a time in Act, where only one agent could be served.
func decide(a *entity.Agent, w *world.World, r *world.Router) decision {
	if w.Rules.Fit {
		c, entropy := recognise(a, w, r)
		if c == nil {
			return decision{}
		}
		return decision{
			plan:    newPlan(a, w, r, c.Def, c.Target, c.Index, c.Situation),
			entropy: entropy,
			counted: true,
		}
	}
	d, target := choose(a, w, r)
	action.Imprint(a)
	s := action.SituationOn(a, w, r, d, target, action.Shared(a, w))
	return decision{plan: newPlan(a, w, r, d, target, action.Index(d), s)}
}

// workersFor is how many goroutines to spread n agents over. One agent each
// is worth it: a decision costs far more than handing one to a goroutine, and
// batching them up only leaves cores idle - insisting on four agents per
// worker cost a third of the speedup when this was measured.
func workersFor(n int) int {
	k := Workers
	if n < k {
		k = n
	}
	if k < 1 {
		k = 1
	}
	return k
}

// inParallel runs f for every index below n, spread over workers goroutines.
// f must not write anything another call to f can see; results belong in a
// slice indexed by i.
func inParallel(n, workers int, f func(i, worker int)) {
	if workers <= 1 {
		for i := 0; i < n; i++ {
			f(i, 0)
		}
		return
	}
	var wg sync.WaitGroup
	for k := 0; k < workers; k++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			for i := k; i < n; i += workers {
				f(i, k)
			}
		}(k)
	}
	wg.Wait()
}

// Recognise is fit-based choice: it samples one of the agent's available
// actions in proportion to how well each fits the moment, sharper the more
// pressing the moment is. No value is computed here. Nil only when nothing
// at all is available, which the always-available rest prevents.
func Recognise(a *entity.Agent, w *world.World) *action.Candidate {
	c, entropy := recognise(a, w, w.Routers(1)[0])
	if c == nil {
		return nil
	}
	w.Choices++
	w.Entropy += entropy
	return c
}

// recognise is Recognise on a given router, returning how undecided the
// moment was rather than adding it to the world's tally. Agents recognising
// at the same time cannot share a tally: the order floating-point additions
// land in would decide the total. The caller adds them up in agent order.
func recognise(a *entity.Agent, w *world.World, r *world.Router) (*action.Candidate, float64) {
	cs := action.CandidatesOn(a, w, r)
	if len(cs) == 0 {
		return nil, 0
	}
	eff := make([]float64, len(cs))
	for i := range cs {
		eff[i] = cs[i].Fit
	}
	temp := w.Rules.Temperature / Intensity(a)
	// The draw comes from the agent's own luck, not the world's one stream,
	// so what it settles on does not depend on who else was deciding beside
	// it. See entity.Agent.Luck.
	i := habit.Sample(a.Luck, eff, temp)
	return &cs[i], habit.Entropy(eff, temp)
}

// Intensity is how pressing an agent's moment is: the sum of its urgencies
// as it weighs them, offset so that a moment with no urgency at all is
// still decided at a finite temperature. It sharpens fit-based sampling
// without ever changing the order of the candidates.
func Intensity(a *entity.Agent) float64 {
	u := need.Urgencies(a.Needs)
	s := 0.5
	for t := range u {
		s += u[t] * a.Personality[t]
	}
	return s
}

// Commit makes a plan for d at target and installs it on the agent. It is
// the one place plans are made, for agents and for the player alike, so the
// lesson drawn when the plan ends is always available. The situation is
// recorded as the agent sees it now, whatever rule chose the action.
func Commit(a *entity.Agent, w *world.World, d *action.Def, target entity.Pos) *entity.Plan {
	r := w.Routers(1)[0]
	action.Imprint(a)
	s := action.SituationOn(a, w, r, d, target, action.Shared(a, w))
	a.Plan = newPlan(a, w, r, d, target, action.Index(d), s)
	return a.Plan
}

// newPlan builds a plan without installing it, so that plans worked out side
// by side can be installed afterwards in a fixed order.
func newPlan(a *entity.Agent, w *world.World, r *world.Router, d *action.Def, target entity.Pos, index int, s habit.Signature) *entity.Plan {
	return &entity.Plan{
		Action: d.Name, Target: target, Remaining: d.Ticks, Total: d.Ticks,
		Index:     index,
		Situation: s,
		Before: habit.Ledger{
			Needs: a.Needs, Urgency: need.Urgencies(a.Needs),
			Food: a.Inventory[entity.Food], Shelter: a.Shelter,
		},
		Started: w.Tick,
		Route:   r.Path(a.Pos, target),
	}
}

// Exertion is the physiological cost of one tick's worth of walking, for an
// ordinary body. Hard ground is slow and tiring in the same measure: a tile
// that takes three ticks to cross also takes three ticks of hunger with it.
// A stronger frame both keeps a faster pace and spends less per tick, so the
// same errand costs a hale agent noticeably less than a worn one. Paving
// discounts the tick as well as shortening the walk, so an errand down a
// street is cheap on both counts.
const Exertion = 0.006

// Act moves agents toward their targets, then advances and applies plans.
func Act(w *world.World) {
	for _, a := range w.Agents {
		if a.Plan == nil {
			continue
		}
		if a.Pos != a.Plan.Target {
			// A plan made by deciding already knows its way. One set by hand -
			// a player's order, a test - works it out on arrival here.
			if len(a.Plan.Route) == 0 {
				a.Plan.Route = w.Grid.Path(a.Pos, a.Plan.Target)
				if len(a.Plan.Route) == 0 {
					a.Plan = nil // nowhere to go from here
					continue
				}
			}
			next := a.Plan.Route[0]
			a.Travel += a.Vigor(w.Tick)
			a.Needs.Add(need.Physiological, -Exertion*w.Grid.MoveDrain(next)/a.Endurance(w.Tick))
			// A tick buys a budget of walking, and the agent spends all of it
			// it can. Over ordinary ground that is the one tile it always
			// was; on ground cheap enough to cross for less than the budget —
			// which is to say on a road — it is more than one, and that is
			// where the speed of a street comes from.
			for len(a.Plan.Route) > 0 {
				step := a.Plan.Route[0]
				cost := w.Grid.MoveCost(step)
				if a.Travel < cost {
					break
				}
				a.Travel -= cost
				a.Pos = step
				a.Plan.Route = a.Plan.Route[1:]
				w.Grid.Tread(step)
			}
			continue
		}
		a.Travel = 0
		a.Plan.Remaining--
		if a.Plan.Remaining > 0 {
			continue
		}
		p := a.Plan
		d := action.ByName(p.Action)
		a.Plan = nil
		// Circumstances may have changed since the plan was made: the last
		// unit of food may have been bought by someone faster. A plan that is
		// no longer possible simply fails, and the failure is a lesson too.
		if d != nil && d.Available(a, w) {
			d.Apply(a, w)
			w.Emit(event.Acted, a.ID, 0, "%s finished %s", a.Name, d.Name)
		}
		Learn(a, w, p)
	}
}

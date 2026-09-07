package system

import (
	"math"
	"runtime"
	"sync"

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
// bulk of a tick - most of it is agents costing errands over the ground - and
// it only reads the world, so it is the one phase that parallelises. Set it
// to 1 to decide one agent at a time.
var Workers = runtime.NumCPU()

// Decide gives every idle agent a plan. Agents choose at the same time and
// commit in turn: choosing only reads the world, so it can be spread over
// goroutines, while the plans land in a fixed order out of a slice indexed by
// agent, which is what keeps a run the same however the goroutines are
// scheduled.
func Decide(w *world.World) {
	idle := make([]*entity.Agent, 0, len(w.Agents))
	for _, a := range w.Agents {
		if a.Plan == nil {
			idle = append(idle, a)
		}
	}
	if len(idle) == 0 {
		return
	}
	plans := make([]*entity.Plan, len(idle))
	routers := w.Routers(workersFor(len(idle)))
	inParallel(len(idle), len(routers), func(i, worker int) {
		a, r := idle[i], routers[worker]
		d, target := choose(a, w, r)
		// The way there is worked out here too, while there are goroutines to
		// work it out on. It is the last of the routing that used to happen a
		// step at a time in Act, where only one agent could be served at once.
		plans[i] = &entity.Plan{
			Action: d.Name, Target: target,
			Remaining: d.Ticks, Total: d.Ticks,
			Route: r.Path(a.Pos, target),
		}
	})
	for i, a := range idle {
		a.Plan = plans[i]
	}
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

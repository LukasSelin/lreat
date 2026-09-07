package system

import (
	"lreat/core/action"
	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/world"
)

// Learn draws the lesson of a finished plan. The outcome is judged by what
// the agent wanted when it decided, against what it has come to expect of
// outcomes in general; the difference pulls the habit for that action toward
// the moment it was chosen in, or pushes it away. Recent earlier actions
// share in the lesson with diminishing weight, so an act that only set up a
// later gain still learns from it. Doing an action also brings it further
// into reach.
//
// This is the only place value touches habit. Choice never sees it.
//
// Credit also follows provenance. A unit of food remembers the act that
// produced it, and when it is eaten the meal's reward reaches that act; a
// roof remembers the act that raised it, and safety that rises afterwards
// reaches that act. Without this a needs-only reward never reaches the
// acts that fill a larder or keep a house standing, because their payoff
// comes many plans later, and the settlement lives from hand to mouth.
func Learn(a *entity.Agent, w *world.World, p *entity.Plan) {
	if p == nil || p.Index < 0 || p.Index >= action.Count || habit.Norm(p.Situation) < habit.Epsilon {
		return // a plan made outside the builder carries no lesson
	}
	action.Imprint(a)
	after := habit.Ledger{Needs: a.Needs, Food: a.Inventory[entity.Food], Shelter: a.Shelter}
	r := habit.Reward(p.Before, after, a.Personality)
	adv := habit.Advantage(r, expectation(a, p.Index))
	a.Baseline += habit.BaselineRate * (r - a.Baseline)
	a.Baselines[p.Index] += habit.ActionBaselineRate * (r - a.Baselines[p.Index])
	provenance(a, p, r, adv, after)

	a.Trace.Push(habit.Step{Index: p.Index, Situation: p.Situation})
	for k, st := range a.Trace {
		h := &a.Habits[st.Index]
		habit.Update(h, st.Situation, adv, habit.Weight(k), habit.Eta)
		habit.ClampNorm(h, habit.MinNorm, habit.MaxNorm)
	}
	h := &a.Habits[p.Index]
	habit.Retain(h, action.Catalog[p.Index].Prior, habit.Lambda)
	habit.ClampNorm(h, habit.MinNorm, habit.MaxNorm)
	a.Reach[p.Index] = min(1, a.Reach[p.Index]+habit.ReachGain)
}

// settle judges the i-th harvest by all it brought against what producing
// acts usually bring, credits the act that made it, and forgets it. This
// is where a field that feeds three meals learns it did better than the
// forest that fed one, and where a poor field learns it did worse.
func settle(a *entity.Agent, i int) {
	h := a.Larder[i]
	credit(a, h.Step, habit.Advantage(h.Returned, a.Harvest))
	a.Harvest += habit.HarvestRate * (h.Returned - a.Harvest)
	a.Larder = append(a.Larder[:i], a.Larder[i+1:]...)
}

// expectation is what the agent has come to expect of an outcome of act i:
// a blend of that act's usual outcome and outcomes in general.
func expectation(a *entity.Agent, i int) float64 {
	return habit.BaselineMix*a.Baselines[i] + (1-habit.BaselineMix)*a.Baseline
}

// credit hands an advantage to an earlier act by provenance. Credit carries
// an advantage, not a reward, on purpose: a raw reward is always good news,
// and an act thanked for every meal drifts onto the average moment and fits
// everything. An advantage can be bad news, and is what lets one way of
// getting food be recognised as better than another.
func credit(a *entity.Agent, st habit.Step, adv float64) {
	h := &a.Habits[st.Index]
	habit.Update(h, st.Situation, adv, habit.ProvenanceWeight, habit.Eta)
	habit.ClampNorm(h, habit.MinNorm, habit.MaxNorm)
}

// provenance keeps the larder and the roof, and pays credit from them. r and
// adv are the reward and the advantage of the plan that just ended.
func provenance(a *entity.Agent, p *entity.Plan, r, adv float64, after habit.Ledger) {
	st := habit.Step{Index: p.Index, Situation: p.Situation}
	// Food that came in is a harvest that remembers this act. Food that went
	// out adds this plan's reward to the harvests it came from, oldest
	// first, and a harvest whose last unit has gone is judged by all it
	// brought. Eating is the case that matters; selling and giving add their
	// smaller rewards the same way, and food lost to a thief adds whatever
	// this plan happened to earn, which is about how theft feels to a farmer.
	switch delta := after.Food - p.Before.Food; {
	case delta > 0.3:
		if len(a.Larder) == habit.LarderCap {
			settle(a, 0) // the oldest harvest is judged by what it has brought so far
		}
		a.Larder = append(a.Larder, habit.Harvest{Step: st, Left: delta})
	case delta < -0.3:
		gone := -delta
		for gone > 0 && len(a.Larder) > 0 {
			h := &a.Larder[0]
			take := min(h.Left, gone)
			h.Returned += r * take / -delta
			h.Left -= take
			gone -= take
			if h.Left > 0.05 {
				break
			}
			settle(a, 0)
		}
	}
	// A roof raised remembers this act, and so does a watch kept. Safety
	// that rose during any later plan thanks them, for the part of the
	// reward that safety accounts for.
	switch {
	case after.Shelter > p.Before.Shelter+0.05:
		a.Roof, a.HasRoof = st, true
	case p.Index == action.Index(action.Guard):
		a.Watch, a.HasWatch = st, true
	default:
		gain := after.Needs[need.Safety] - p.Before.Needs[need.Safety]
		if gain <= 0 {
			return
		}
		r := gain * p.Before.Urgency[need.Safety] * a.Personality[need.Safety]
		if a.HasRoof {
			credit(a, a.Roof, habit.Advantage(r, expectation(a, a.Roof.Index)))
		}
		if a.HasWatch {
			credit(a, a.Watch, habit.Advantage(r, expectation(a, a.Watch.Index)))
		}
	}
}

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
	provenance(a, p, adv, after)

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

// expectation is what the agent has come to expect of an outcome of act i:
// a blend of that act's usual outcome and outcomes in general.
func expectation(a *entity.Agent, i int) float64 {
	return habit.BaselineMix*a.Baselines[i] + (1-habit.BaselineMix)*a.Baseline
}

// credit hands an advantage to an earlier act by provenance: the meal was
// better or worse than meals usually are, and the act that made it possible
// learns from that. Credit carries the advantage, not the reward, on
// purpose. A raw reward is always good news, and an act thanked for every
// meal drifts onto the average moment and fits everything; an advantage is
// bad news for the forage that fed a full belly, so production learns to
// stop when the larder is full.
func credit(a *entity.Agent, st habit.Step, adv float64) {
	h := &a.Habits[st.Index]
	habit.Update(h, st.Situation, adv, habit.ProvenanceWeight, habit.Eta)
	habit.ClampNorm(h, habit.MinNorm, habit.MaxNorm)
}

// provenance keeps the larder and the roof, and pays credit from them. adv
// is the advantage of the plan that just ended.
func provenance(a *entity.Agent, p *entity.Plan, adv float64, after habit.Ledger) {
	st := habit.Step{Index: p.Index, Situation: p.Situation}
	// Food that came in remembers this act; food that went out thanks the
	// act that brought it, with this act's reward. Eating is the case that
	// matters; selling and giving pass on their smaller rewards the same way,
	// and food lost to a thief thanks its producer with whatever this plan
	// happened to earn, which is about how theft feels to a farmer.
	switch delta := after.Food - p.Before.Food; {
	case delta > 0.3:
		for n := max(1, int(delta+0.5)); n > 0 && len(a.Larder) < habit.LarderCap; n-- {
			a.Larder = append(a.Larder, st)
		}
	case delta < -0.3:
		for n := max(1, int(-delta+0.5)); n > 0 && len(a.Larder) > 0; n-- {
			credit(a, a.Larder[0], adv)
			a.Larder = a.Larder[1:]
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

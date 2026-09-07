package system

import (
	"lreat/core/action"
	"lreat/core/entity"
	"lreat/core/habit"
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
func Learn(a *entity.Agent, w *world.World, p *entity.Plan) {
	if p == nil || p.Index < 0 || p.Index >= action.Count || habit.Norm(p.Situation) < habit.Epsilon {
		return // a plan made outside the builder carries no lesson
	}
	action.Imprint(a)
	after := habit.Ledger{Needs: a.Needs}
	r := habit.Reward(p.Before, after, a.Personality)
	adv := habit.Advantage(r, a.Baseline)
	a.Baseline += habit.BaselineRate * (r - a.Baseline)

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

package action

import (
	"math"

	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/world"
)

// Rearing. A settlement that waits fifteen years for a birth to become a
// worker is a settlement with dependants in it, and dependants are the one
// thing nothing in the catalog used to be about. Everything an agent did for
// anybody else it did for whoever was nearest and worst off - which is
// charity, and charity is not what a household is. A parent feeds its own
// before they are desperate, teaches its own before they have shown they are
// worth teaching, and keeps a roof over its own without being asked.
//
// None of that is staged. Kinship is a fact the world records (see
// entity.Agent.Parent), the ontology has a role for it (ontology.Child), and
// what a parent does about it is two ordinary acts it recognises or fails to
// recognise like any other. A cold parent rears badly. That is the point of
// putting it here rather than in a system.

// FeedChild is a parent putting food into its own child's hands. It is the
// same move as giving to a needy stranger and it is a different act: see the
// terms in move.go, and ontology.Child for the moment it belongs to.
var FeedChild = mover("transfer/provision>child")

// ownChild is the nearest child of a's that has not grown up yet, within
// radius, or nil. Ties go to the nearest, then to the earliest in the slice,
// which keeps runs deterministic.
func ownChild(a *entity.Agent, w *world.World, radius int) *entity.Agent {
	return nearestWith(a, w, radius, func(o *entity.Agent) bool {
		return o.Parent == a.ID && !entity.Adult(o.Age(w.Tick))
	})
}

// teachable reports whether a has anything worth passing on and anybody of
// its own to pass it to.
func teachable(a *entity.Agent, w *world.World) bool {
	if _, level := a.BestSkill(); level < 0.2 {
		return false
	}
	return ownChild(a, w, reachRadius) != nil
}

// TeachChild is bringing a child up in the work. It is teaching, and it
// differs from teaching at large in the two ways a parent differs from a
// master. It asks less of the teacher - a fifth of a practice against two
// fifths, because a parent shows a child what it does whether or not it does
// it well - and it pays in no standing at all, because there is nobody in
// the doorway to be impressed. What it pays in is the child, and a household
// that keeps its craft.
var TeachChild = &Def{
	Name: "teach child", Ticks: 3, Available: teachable,
	Target: func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
		if o := ownChild(a, w, reachRadius); o != nil {
			return o.Pos, true
		}
		return entity.Pos{}, false
	},
	With: func(a *entity.Agent, w *world.World, _ entity.Pos) *entity.Agent {
		return ownChild(a, w, reachRadius)
	},
	Expect: func(a *entity.Agent, w *world.World, target entity.Pos) need.Levels {
		gain := need.Levels{need.Belonging: rearingWarmth, need.Actualization: 0.03}
		if o := intended(a, w, target); o != nil {
			// What a parent gets out of it is most where the child has
			// furthest to come: teaching a child that already does the work
			// is watching it work.
			skill, level := a.BestSkill()
			gain[need.Actualization] += 0.03 * math.Max(0, level-o.Skills[skill])
		}
		return gain
	},
	Apply: func(a *entity.Agent, w *world.World) {
		o := ownChild(a, w, companionRadius)
		if o == nil {
			return
		}
		skill, level := a.BestSkill()
		o.AddSkill(skill, childLearns)
		// The craft, not the disposition: a child already carries its
		// parent's habits and what a showing adds is the doing. See Pass.
		Reachable(a, o, skill)
		o.Rate(a.ID, skill, level, w.Tick)
		o.AddBond(a.ID, 0.04)
		a.AddBond(o.ID, 0.04)
		a.Needs.Add(need.Belonging, rearingWarmth)
		a.Needs.Add(need.Actualization, 0.03)
		o.Needs.Add(need.Belonging, 0.08)
		o.Tended = min(1, o.Tended+TendTeaching)
		w.Emit(event.Taught, a.ID, o.ID, "%s brought %s up to %s", a.Name, o.Name, skill)
	},
}

// rearingWarmth is what a parent gets out of tending its own, and it is
// deliberately small - half of what giving to a stranger pays, where giving
// is already the least rewarded thing in the catalog. Nobody is in the
// doorway to be impressed and it is a duty rather than a kindness, so the
// only thing it should answer is the tie itself.
//
// It was 0.12 first, which is more than a stranger's charity pays, and that
// was a mistake of the kind the roads have their own note about. It did not
// show up as a settlement that reared too much - rearing stayed under three
// per cent of what anybody did - but as one whose belonging and standing sat
// at four fifths satisfied all year. A population with nothing left to want
// does not go out, and the share of fertile adults over all three birth
// gates at once fell from 0.31 to 0.17 while their average needs rose.
const rearingWarmth = 0.04

// What the two rearing acts are worth to a childhood. Feeding is worth more
// than a showing because it is the one a child cannot do without, and both
// are worth several months of the fade: a few visits a year is a childhood
// somebody is keeping an eye on, and that is the cadence these are set to.
// See system.Rearing, where the fade is.
const (
	TendFeeding  = 0.12
	TendTeaching = 0.10
)

// childLearns is what one showing passes to a child. It is half again what a
// pupil in the square takes from a lesson: a child is shown the work over and
// over, from the age it can carry anything, by somebody who has no reason to
// stop.
const childLearns = 0.06

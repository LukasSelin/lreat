package action

import (
	"math/rand/v2"
	"slices"

	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/world"
)

// Reach is the distance gate. An action that begins far out of reach is one
// whose moment an agent cannot yet recognise as its own, however well the
// coordinates line up. Four things bring it closer: doing it, which is
// simply practice; studying, which widens what one can imagine doing; being
// taught, which hands over a teacher's recognition ready-made; and living in
// a settlement that has discovered the thing, which puts it in front of
// everyone. None of them asks whether the doing went well - reach is
// familiarity, not success. Reach only ever grows.

const (
	// PractisedReach is how much doing an action once brings it further
	// within reach.
	PractisedReach = 0.02
	// StudyReach is how much of the remaining distance one study closes on
	// every gated action, before the settlement's study rate.
	StudyReach = 0.03
	// TaughtReach is the share of a teacher's reach a student receives.
	TaughtReach = 0.6
	// TaughtHabit is how far a student's habit moves toward the teacher's.
	TaughtHabit = 0.3
	// Opened is the reach everyone has for an action the settlement has
	// discovered.
	Opened = 0.8
	// InheritReach is the share of a parent's reach a child is born with.
	InheritReach = 0.7
)

// BornNoise is the drift on each varying coordinate of a founder's habits,
// which are otherwise the bare priors. It is what makes one founder read a
// morning differently from the next.
var BornNoise = 0.15

// InheritNoise is the drift on each varying coordinate a child's habits take
// from a parent's.
var InheritNoise = 0.05

// Gated lists the catalog positions of actions that do not start fully
// within reach.
func Gated() []int {
	var out []int
	for i, d := range Catalog {
		if d.Reach0 < 1 {
			out = append(out, i)
		}
	}
	return out
}

// ForSkill is the action a skill is exercised in, or nil.
func ForSkill(s entity.Skill) *Def {
	switch s {
	case entity.Farming:
		return Farm
	case entity.Building:
		return BuildShelter
	case entity.Crafting:
		return Craft
	case entity.Scholarship:
		return Study
	case entity.Guarding:
		return Guard
	case entity.Fishing:
		return Fish
	}
	return nil
}

// Practise is what doing an action does to reach: having done a thing once
// it is that much less foreign the next time, whatever came of it.
func Practise(a *entity.Agent, i int) {
	if i < 0 || i >= Count {
		return
	}
	Imprint(a)
	// How much a turn at the work brings it within reach is the learner's
	// own. Everyone climbed the same ladder at the same rate before this,
	// so who ended up a mason was whoever happened to lay the first stone.
	a.Reach[i] = min(1, a.Reach[i]+PractisedReach*a.Mind.Learns())
}

// Broaden is what study does to reach: every gated action comes a little
// closer, faster where the settlement has learned to keep records.
func Broaden(a *entity.Agent, w *world.World) {
	Imprint(a)
	for _, i := range Gated() {
		a.Reach[i] += StudyReach * (1 - a.Reach[i]) * w.Mods.StudyRate
	}
}

// Pass is what teaching does: the student receives a share of the teacher's
// reach for the action the skill lives in, and its habit for that action
// moves toward the teacher's. Recognition is handed over, not just skill.
func Pass(teacher, student *entity.Agent, s entity.Skill) {
	d := ForSkill(s)
	if d == nil {
		return
	}
	Imprint(teacher)
	Imprint(student)
	i := Index(d)
	student.Reach[i] = max(student.Reach[i], TaughtReach*teacher.Reach[i])
	for k := range student.Habits[i] {
		student.Habits[i][k] += TaughtHabit * (teacher.Habits[i][k] - student.Habits[i][k])
	}
	habit.ClampNorm(&student.Habits[i], habit.MinNorm, habit.MaxNorm)
}

// Inherit gives a child its parent's habits and a share of its reach. With
// rng the varying coordinates drift a little, so children are like their
// parents and not copies; without it they are copies, for runs that must
// not draw on the RNG for this.
func Inherit(child, parent *entity.Agent, w *world.World, rng *rand.Rand) {
	Imprint(parent)
	child.Room()
	w.Room()
	child.Habits = slices.Clone(parent.Habits)
	child.Seeded = parent.Seeded
	// Only the slots of the child's own kind drift and are reached: the
	// rest are empty in the parent and stay empty, and drawing a drift
	// for them would be chance a settlement never used to spend.
	mine := For(child.Species())
	if rng != nil {
		for _, i := range mine {
			for k := range child.Habits[i] {
				child.Habits[i][k] += rng.NormFloat64() * InheritNoise
			}
			habit.ClampNorm(&child.Habits[i], habit.MinNorm, habit.MaxNorm)
		}
	}
	for _, i := range mine {
		child.Reach[i] = max(w.ReachFloor[i], InheritReach*parent.Reach[i])
	}
	// A child is shown the country it grows up in. Without this every
	// generation would be born knowing nowhere and would have to walk the
	// neighbourhood again before it could site anything, which is not how
	// anybody learns where the good ground is. It has to be a fresh slice:
	// Habits and Reach are arrays and copy on assignment, but a shared slice
	// would give parent and child one backing store to append into.
	child.Places = parent.CopyPlaces()
	child.Imprinted = true
}

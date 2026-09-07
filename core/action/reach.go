package action

import (
	"math/rand/v2"

	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/world"
)

// Reach is the distance gate. An action that begins far out of reach is one
// whose moment an agent cannot yet recognise as its own, however well the
// coordinates line up. Three things bring it closer: studying, which widens
// what one can imagine doing; being taught, which hands over a teacher's
// recognition ready-made; and living in a settlement that has discovered
// the thing, which puts it in front of everyone. Doing it brings it closer
// too, in the learning step. Reach only ever grows.

const (
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
	// InheritNoise is the drift on each learnable coordinate a child's
	// habits take from a parent's.
	InheritNoise = 0.05
	// InheritReach is the share of a parent's reach a child is born with.
	InheritReach = 0.7
)

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
// rng the learnable coordinates drift a little, so children are like their
// parents and not copies; without it they are copies, for runs that must
// not draw on the RNG for this.
func Inherit(child, parent *entity.Agent, w *world.World, rng *rand.Rand) {
	Imprint(parent)
	child.Habits = parent.Habits
	if rng != nil {
		for i := range Catalog {
			for k := range child.Habits[i] {
				if habit.Learnable[k] {
					child.Habits[i][k] += rng.NormFloat64() * InheritNoise
				}
			}
			habit.ClampNorm(&child.Habits[i], habit.MinNorm, habit.MaxNorm)
		}
	}
	for i := range Catalog {
		child.Reach[i] = max(w.ReachFloor[i], InheritReach*parent.Reach[i])
	}
	child.Imprinted = true
}

// Package entity defines agents and the plain-struct components they carry.
//
// The player is an ordinary Agent. Nothing in the simulation may branch on
// whether an agent is the player; commands only set that agent's plan.
package entity

import (
	"lreat/core/belief"
	"lreat/core/habit"
	"lreat/core/need"
)

// ID identifies an agent. Zero means "no agent".
type ID int

// Pos is a tile coordinate on the world grid.
type Pos struct{ X, Y int }

// Dist is the Chebyshev distance: the number of eight-directional steps
// needed to walk from a to b.
func Dist(a, b Pos) int {
	dx, dy := a.X-b.X, a.Y-b.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dx > dy {
		return dx
	}
	return dy
}

// StepToward returns from moved one tile toward to.
func StepToward(from, to Pos) Pos {
	return Pos{X: from.X + sign(to.X-from.X), Y: from.Y + sign(to.Y-from.Y)}
}

func sign(v int) int {
	switch {
	case v < 0:
		return -1
	case v > 0:
		return 1
	}
	return 0
}

// Good is a tradeable resource.
type Good int

const (
	Food Good = iota
	Wood
	Tools
	GoodCount
)

var goodNames = [GoodCount]string{"food", "wood", "tools"}

func (g Good) String() string { return goodNames[g] }

// Skill is a learnable competence that scales action yields.
type Skill int

const (
	Farming Skill = iota
	Building
	Crafting
	Scholarship
	Guarding
	SkillCount
)

var skillNames = [SkillCount]string{"farming", "building", "crafting", "scholarship", "guarding"}

func (s Skill) String() string { return skillNames[s] }

// Bond is what one agent knows and feels about another. Everything in it is
// belief: Competence is what the holder thinks the other can do, which may be
// nothing like the truth.
type Bond struct {
	To       ID
	Strength float64 // affinity, raised by time spent together
	Regard   float64 // moral standing in [-1,1], set by judging their conduct
	// Competence is the believed level of each of the other agent's skills.
	Competence [SkillCount]float64
	LastSeen   int

	// Expect is how the holder anticipates a meeting with this person will
	// go, in [-1,1]. It is seeded by first impression and hearsay and then
	// corrected by how meetings actually went, so two people can hold very
	// different expectations of the same third person.
	Expect float64
	// Met counts meetings in person. Zero means the holder only knows of
	// this person through what others have said.
	Met int
}

// Plan is an action the agent has committed to, and where it happens. The
// agent walks to Target first; Remaining only counts down once it is there.
// Agents do not re-decide while a plan is running; that inertia is deliberate.
type Plan struct {
	Action    string
	Target    Pos
	Remaining int
	Total     int

	// Index is the catalog position of Action, so the outcome can be
	// credited to the right habit without a name lookup.
	Index int
	// Situation is the moment as the agent saw it when it chose this action.
	// The lesson drawn on completion is about that moment, not about the
	// one the agent finds itself in afterwards.
	Situation habit.Signature
	// Before is what the agent wanted and had when it decided, so that the
	// outcome is judged by the urgencies of the time.
	Before habit.Ledger
	// Started is the tick the plan was made.
	Started int
}

// Agent is one actor in the world.
type Agent struct {
	ID   ID
	Name string
	Born int

	Needs       need.Levels
	Personality need.Weights

	Pos      Pos
	Home     Pos
	HasHome  bool
	Field    Pos
	HasField bool

	Inventory [GoodCount]float64
	Skills    [SkillCount]float64
	Wealth    float64
	// Reputation is public standing: the visible record of things made,
	// taught, and done that a stranger can see or has heard tell of. It is
	// not what anyone thinks of the agent's character; that lives in other
	// people's Bonds and differs from person to person.
	Reputation float64
	Shelter    float64 // quality of housing in [0,1]; decays

	// Vitality is the body an agent was born with: how much effort it can
	// carry, around 1 and rarely far from it. It is drawn at birth and
	// inherited with drift, so bodies vary the way personalities do, by a
	// little and across generations.
	Vitality float64
	// Health is present condition in [0,1]. It follows how well the agent is
	// fed and housed, but slowly, so it reads as a constitution worn down or
	// built back up over a long stretch rather than a second hunger bar.
	Health float64

	// Temperament is how the agent approaches other people, independent of
	// what it needs from them.
	Temperament Temperament

	// Efficacy is what the agent believes about its own skills. It trails the
	// truth in Skills and can be badly wrong in either direction, which is
	// what makes an agent hire out work it could have done, or take on work
	// it cannot finish.
	Efficacy [SkillCount]float64
	// Norms is what this agent holds to be right. It judges itself and
	// everyone else by these, and passes them to its children.
	Norms belief.Norms

	// Caution is the learned expectation that wrongdoing is answered here.
	// It is not a rule and not a fear of any particular punishment: it rises
	// from seeing reprisals happen and decays when they stop, so deterrence
	// exists only where somebody keeps enforcing.
	Caution float64

	// Travel is effort banked toward entering the next tile. An agent puts
	// one tick of walking into it per tick and steps once the tile it is
	// entering has been paid for, so hard ground is crossed slowly rather
	// than in the same stride as open grass.
	Travel float64

	Starving int // consecutive ticks at the bottom of the physiological tier
	Bonds    []Bond
	Plan     *Plan

	// Habits is what this agent has come to recognise as the kind of moment
	// each action belongs to, indexed by catalog position. Each starts as a
	// copy of the action's shared prior and is moved by the agent's own
	// outcomes, copied by teachers, and inherited with drift by children.
	// It is the agent's character as revealed in what it does.
	Habits [habit.MaxActions]habit.Signature
	// Reach is how far into reach each action is for this agent, in [0,1].
	// Everyday living is fully in reach from birth; crafts and learning
	// begin far off and are brought closer by study, teaching, and what the
	// settlement has discovered.
	Reach [habit.MaxActions]float64
	// Baseline is the agent's slow-moving sense of what an ordinary outcome
	// feels like, and Baselines the same for each action on its own.
	// Lessons are drawn from how an outcome differs from a blend of the two.
	Baseline  float64
	Baselines [habit.MaxActions]float64
	// Trace is the short memory of recent actions that share in the next
	// reward, so that an action that only set up a later gain still learns.
	Trace habit.Trace
	// Imprinted is set once Habits and Reach have been seeded from the
	// catalog priors. Seeding is lazy because the world cannot see the
	// catalog, and a newborn's first decision is the earliest it is needed.
	Imprinted bool
}

// ordinaryBody is the vitality of an agent constructed without one, as
// tooling and tests sometimes do. Treating it as average keeps a bodyless
// agent moving normally instead of standing still forever.
const ordinaryBody = 1

// Endurance is how well the agent's frame carries effort: a larger store to
// spend walking out of. It is the body it was born with, condition aside.
func (a *Agent) Endurance() float64 {
	if a.Vitality <= 0 {
		return ordinaryBody
	}
	return a.Vitality
}

// Vigor is the pace the agent can actually keep, its frame discounted by the
// condition that frame is currently in. A hale agent covers hard ground
// faster; a worn-down one labours over the same tile.
func (a *Agent) Vigor() float64 {
	return a.Endurance() * (0.6 + 0.4*need.Clamp(a.Health))
}

// AddBond strengthens (or creates) the bond to another agent.
func (a *Agent) AddBond(to ID, delta float64) {
	for i := range a.Bonds {
		if a.Bonds[i].To == to {
			a.Bonds[i].Strength = need.Clamp(a.Bonds[i].Strength + delta)
			return
		}
	}
	a.Bonds = append(a.Bonds, Bond{To: to, Strength: need.Clamp(delta)})
}

// BondWith returns the strength of the bond to another agent, zero if none.
func (a *Agent) BondWith(to ID) float64 {
	for _, b := range a.Bonds {
		if b.To == to {
			return b.Strength
		}
	}
	return 0
}

// BestSkill returns the agent's strongest skill and its level.
func (a *Agent) BestSkill() (Skill, float64) {
	best, level := Skill(0), a.Skills[0]
	for s := Skill(1); s < SkillCount; s++ {
		if a.Skills[s] > level {
			best, level = s, a.Skills[s]
		}
	}
	return best, level
}

// AddSkill raises a skill and clamps it to [0,1].
func (a *Agent) AddSkill(s Skill, d float64) {
	a.Skills[s] = need.Clamp(a.Skills[s] + d)
}

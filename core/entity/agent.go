// Package entity defines agents and the plain-struct components they carry.
//
// The player is an ordinary Agent. Nothing in the simulation may branch on
// whether an agent is the player; commands only set that agent's plan.
package entity

import (
	"math/rand/v2"

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
	Stone
	Meals // cooked food: keeps, and feeds more
	GoodCount
)

var goodNames = [GoodCount]string{"food", "wood", "tools", "stone", "meals"}

func (g Good) String() string { return goodNames[g] }

// Skill is a learnable competence that scales action yields.
type Skill int

const (
	Farming Skill = iota
	Building
	Crafting
	Scholarship
	Guarding
	Fishing
	SkillCount
)

var skillNames = [SkillCount]string{"farming", "building", "crafting", "scholarship", "guarding", "fishing"}

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

	// Route is the way to Target that was cheapest when the plan was made,
	// the tiles still to be walked, nearest first. It is worked out once, at
	// the moment of deciding, rather than asked again every tick: the same
	// inertia that keeps an agent on a plan keeps it on the way it set out
	// by. Somebody may pave a better street while it is walking, and it will
	// not notice until its next errand takes it that way.
	Route []Pos

	// NoWay is set on a plan made by deciding whose way was looked for and
	// not found, and Waters is the water as it was then. Acting on it drops
	// it rather than looking again, unless the water has moved since - a
	// bridge gone up between the deciding and the acting - in which case
	// the way is looked for once more, as it always was.
	NoWay  bool
	Waters int

	// Index is the catalog position of Action, so a finished plan can be
	// tied back to the action it ran without a name lookup.
	Index int
	// Started is the tick the plan was made.
	Started int
}

// Agent is one actor in the world.
type Agent struct {
	ID   ID
	Name string
	Born int

	// Luck is the agent's own stream of chance, seeded when it is born. An
	// agent draws from this rather than from the world's one stream so that
	// what it decides depends on what it has drawn before and not on who
	// else happened to draw in between. That is what lets a whole population
	// decide at the same time and still come out the same as if they had
	// gone one at a time - and it is why only deciding may run in parallel:
	// Luck covers choosing, while acting on the choice still draws on the
	// world.
	Luck *rand.Rand

	Needs       need.Levels
	Personality need.Weights

	Pos     Pos
	Home    Pos
	HasHome bool

	// Field is the first furrow a farmer broke and the tile the household is
	// anchored to; Parcel is the whole holding, that furrow and every strip
	// broken beside it since. A house is one tile and a holding is many,
	// because a family eats far more ground than it sleeps on.
	Field    Pos
	HasField bool
	Parcel   []Pos

	Inventory [GoodCount]float64
	Skills    [SkillCount]float64
	Wealth    float64
	// Reputation is public standing: the visible record of things made,
	// taught, and done that a stranger can see or has heard tell of. It is
	// not what anyone thinks of the agent's character; that lives in other
	// people's Bonds and differs from person to person.
	Reputation float64
	Shelter    float64 // quality of housing in [0,1]; decays

	// Body and Mind are what this one is made of and thinks with: the frame
	// it was born with, what it burns, what cold it can stand, how fast it
	// takes to a craft, how firmly it decides, how far it will go. All of
	// them drawn at birth and inherited with drift, so a settlement's people
	// vary the way their personalities do - by a little, and across
	// generations. See trait.go.
	Body Body
	Mind Mind
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
	// Places is the ground this agent has stood on and thought worth
	// remembering. It is the whole of what it knows of the country: an agent
	// sites a house, a field, and a move out of this list and nothing else,
	// so where a settlement grows is decided by where its people have
	// actually walked. See place.go.
	Places []Place
	Plan   *Plan

	// NoWay is the last way this agent looked for and did not find: from
	// where, to where, carrying what, and with the water as it was. Looking
	// for a way that is not there opens everything the walker could reach,
	// and an agent that wants the same thing tomorrow from the same place
	// would open it all again to learn what it already knows.
	NoWay Impasse

	// Habits is the kind of moment this agent recognises each action as
	// belonging to, by habit slot, which is catalog position. Each starts
	// as a copy of the action's shared prior with a little drift of its
	// own; teachers copy theirs onto students, and children inherit their
	// parents' with drift again. Nothing an outcome does moves them. It is
	// the agent's character as it was handed down and passed on, not as it
	// was earned. Room keeps it and Reach as long as there are slots; see
	// habit.Register.
	Habits []habit.Signature
	// Reach is how far into reach each action is for this agent, in [0,1].
	// Everyday living is fully in reach from birth; crafts and learning
	// begin far off and are brought closer by practice, study, teaching,
	// and what the settlement has discovered.
	Reach []float64
	// Seeded is how many slots have been given a starting habit and reach,
	// so that a slot given after this agent was imprinted can be told from
	// one it has lived with.
	Seeded int
	// Imprinted is set once Habits and Reach have been seeded from the
	// catalog priors. Seeding is lazy because the world cannot see the
	// catalog, and a newborn's first decision is the earliest it is needed.
	Imprinted bool

	// Company is the room this agent looks over a crowd in: whoever was
	// within reach the last time it wondered who to go and see. It is kept
	// rather than taken afresh because it is taken afresh every time
	// anybody weighs going to see somebody, and in a crowd it is as long
	// as the crowd - that one list was more than half of everything the
	// simulation allocated. What is in it between decisions means nothing.
	Company []*Agent
	// Sizing is the same thing for the list of errands this agent is
	// weighing up. It is an any because what is in it is a candidate
	// action, which is the action package's to describe and this package
	// must not know: an agent carries its own working memory, and the only
	// thing it can say about this piece of it is whose it is. It holds a
	// pointer to the list rather than the list, so that lengthening the
	// list does not have to be stored back and boxed again.
	//
	// Both of these are an agent's own scratch, and that is what makes
	// them safe while everybody decides at once: one agent is sized up by
	// one goroutine, so nobody else is ever looking at this room.
	Sizing any
}

// ordinaryBody is the vitality of an agent constructed without one, as
// tooling and tests sometimes do. Treating it as average keeps a bodyless
// agent moving normally instead of standing still forever.
const ordinaryBody = 1

// Endurance is how well the agent's frame carries effort: a larger store to
// spend walking out of. It is the body it was born with, grown into or given
// back with age, condition aside.
func (a *Agent) Endurance(tick int) float64 {
	return a.Body.Frame() * AgeFactor(a.Age(tick))
}

// Vigor is the pace the agent can actually keep: its frame at this age,
// discounted by the condition that frame is currently in. A hale adult covers
// hard ground faster; a child, an elder, or a worn-down agent labours over the
// same tile.
func (a *Agent) Vigor(tick int) float64 {
	return a.Endurance(tick) * (0.6 + 0.4*need.Clamp(a.Health))
}

// Load is how much an agent has in its arms, counted in units of goods with
// every good the same. It is a yes-or-no dressed as a number: everything that
// reads it compares it against world.SwimLoad and asks whether this walker
// may take to the water, and a person holding a crumb is as shut out of the
// river as one holding a harvest.
//
// So the counting here is deliberately coarse, and what one thing is to carry
// against another is not its question. That is world.Hauled, which reads the
// ontology's Heavy off the trees and is what wears the ground; this package
// cannot ask, because the ontology names skills from here and the two would
// import in a circle.
func (a *Agent) Load() float64 {
	total := 0.0
	for _, q := range a.Inventory {
		if q > 0 {
			total += q
		}
	}
	return total
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

// Room makes the per-act tables as long as there are slots. A slot given
// after this agent was born is a fresh, empty one for it, to be seeded
// from the act's prior the next time it is imprinted.
func (a *Agent) Room() {
	n := habit.Slots()
	a.Habits = habit.Grow(a.Habits, n)
	a.Reach = habit.Grow(a.Reach, n)
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

// Holds reports whether p is ground this agent has broken and works.
func (a *Agent) Holds(p Pos) bool {
	for _, q := range a.Parcel {
		if q == p {
			return true
		}
	}
	return false
}

// Impasse is a way that was looked for and was not there.
type Impasse struct {
	From, To Pos
	Laden    bool
	// Waters is the water as it was when the way was looked for; see
	// world.Grid.Waters. A bridge since changes the answer.
	Waters int
	Known  bool
}

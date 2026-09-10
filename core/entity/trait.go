package entity

// What one creature is, as against what it wants and what it has learned.
//
// An agent's character was already enormous - one signature of twenty
// numbers for each of thirty-odd actions, and a personality and a set of
// values besides - but all of it is a way of weighing a moment. None of it
// could say that this person eats more than that one, or feels the cold, or
// picks a craft up quickly, or will walk a long way for good ground. Those
// were one number apiece for the whole population, written in the systems
// that used them, and a settlement of five hundred had exactly one body and
// exactly one cast of mind between them.
//
// So they live here. Every one is a multiplier on something the world does
// to an agent or the agent does to itself, 1 being the ordinary measure, and
// every one is drawn at birth and inherited with drift - the same way
// Vitality and Norms and Temperament already were. Nothing an outcome does
// moves them: they are what somebody is, not what has happened to them.
//
// They are two structs and not eight fields on Agent because a deer is
// coming. What a species is, is a body and a mind it hands its young; an
// individual differs from its kind by drift, exactly as a child differs
// from its parent. See the Species work.

// Body is the frame a creature was born with: what it can carry, what it
// burns to stay alive, and what weather it can stand.
type Body struct {
	// Vitality is how much effort the frame can carry, and it is the oldest
	// of these: it scales the pace of a walk and what a walk costs, through
	// Endurance and Vigor.
	Vitality float64
	// Metabolism is how fast this body spends what it has eaten. A big
	// frame or a hot one burns through a larder that would keep a smaller
	// one; it is the first thing a famine sorts people by, and it makes
	// hunger a thing some people are more often than others rather than a
	// clock everybody keeps together.
	Metabolism float64
	// Hardiness is what the cold costs this body, inverted: what a hardy
	// one shrugs off leaves a frail one slower come spring. It is the whole
	// of what a winter without a roof does, both the hunger of staying warm
	// and the condition it takes.
	Hardiness float64
}

// Mind is the cast of a creature's thought: how fast it takes to a thing,
// how firmly it settles on one, and how far its world reaches.
type Mind struct {
	// Plasticity is how fast practice brings an action within reach. It is
	// what lets a settlement have people who take to a craft and people who
	// never quite do, rather than everyone climbing the same ladder at the
	// same rate.
	Plasticity float64
	// Resolve is how decidedly this mind settles a moment: it sharpens the
	// temperature the choice is drawn at, so a resolute agent takes the
	// best fit it can see and a wavering one is talked round by a near
	// second. The pressure of the moment already sharpens everyone; this is
	// the part of it that is the person rather than the day.
	Resolve float64
	// Horizon is how far this one will look for ground worth having. A
	// homebody breaks its field beside the house it has; someone with a
	// wider horizon walks past a poor field for a better one, and founds
	// the edge of a settlement doing it.
	Horizon float64
}

// Ordinary is a body and a mind with nothing remarkable about either: every
// measure at 1. It is what a creature has where nobody has said otherwise,
// which is what a test usually means and what an agent made by hand gets.
func Ordinary() (Body, Mind) {
	return Body{Vitality: 1, Metabolism: 1, Hardiness: 1},
		Mind{Plasticity: 1, Resolve: 1, Horizon: 1}
}

// or is a measure with the ordinary one standing in where none was given, so
// that an agent put together by hand in a test behaves like an ordinary one
// rather than like a corpse. Every reading of a trait goes through it.
func or(v float64) float64 {
	if v <= 0 {
		return 1
	}
	return v
}

// The readings. Each is the trait as the systems should use it, with the
// ordinary measure standing in for one never drawn.

func (b Body) Frame() float64   { return or(b.Vitality) }
func (b Body) Burn() float64    { return or(b.Metabolism) }
func (b Body) Hardy() float64   { return or(b.Hardiness) }
func (m Mind) Learns() float64  { return or(m.Plasticity) }
func (m Mind) Decides() float64 { return or(m.Resolve) }
func (m Mind) Reaches() float64 { return or(m.Horizon) }

// Package habit is the vector space in which agents recognise what a moment
// calls for.
//
// A situation is a point in that space: how urgent each need is, what the
// agent has on hand, who is around, what it holds to be right, and how the
// candidate action sits relative to it. Each action carries a Signature in the
// same space, the kind of moment it belongs to. Choosing is recognition:
// which signature lies closest in direction to the moment at hand. No value
// is computed anywhere in it. A signature is a disposition, not a score, and
// nothing an outcome does can move one: an agent is born with the ones its
// parents had, receives a teacher's when it is taught, and keeps them.
//
// This package is a leaf. It knows nothing of agents, actions, or the world.
package habit

import (
	"math"
	"math/rand/v2"
)

// Dims is the size of the space.
const Dims = 20

// Slots are places in every agent's tables, one per act, keyed by the
// act's canonical name. A slot is given once, in the order acts are first
// registered, and is never reused or renumbered: an act that enters later
// takes a fresh slot after everything before it, and every table already
// in use grows to make room. So tables stay dense arrays indexed by slot,
// and what a slot means never moves under a habit.
var slots = struct {
	keys  []string
	index map[string]int
}{index: map[string]int{}}

// Register gives key a slot, or returns the one it has.
func Register(key string) int {
	if i, ok := slots.index[key]; ok {
		return i
	}
	i := len(slots.keys)
	slots.keys = append(slots.keys, key)
	slots.index[key] = i
	return i
}

// Slot is the slot key holds, if it holds one.
func Slot(key string) (int, bool) {
	i, ok := slots.index[key]
	return i, ok
}

// Slots is how many slots have been given.
func Slots() int { return len(slots.keys) }

// Keys lists every slot's key, by slot.
func Keys() []string { return append([]string(nil), slots.keys...) }

// Grow extends s to at least n entries, zero-filled, keeping what it had.
func Grow[T any](s []T, n int) []T {
	if len(s) >= n {
		return s
	}
	return append(s, make([]T, n-len(s))...)
}

// The dimensions. The first five are urgencies weighted by personality;
// the next five are surroundings: shelter, company, the weather and what
// it is doing to this body, and whether anybody keeps order; five are what
// the agent holds to be right and how much reprisal it expects; the last
// five are patched per candidate action, because how near a thing is, how
// one feels about the person involved, how able one believes oneself, and
// how short one is of what the act brings or how well supplied in what it
// costs all depend on which action is being considered.
//
// Lack and Stock are how an agent's stores enter the space. There is no
// coordinate for food or wood or coin as such: a moment is short of what
// this act would bring, or well supplied in what it would spend, whatever
// that is. That is what lets a material the trees did not have yesterday
// be recognised today, by an agent that was never told its name. An act
// that moves nothing in particular reads the stores in general on the
// same two coordinates - short of things, or well off - so that every
// candidate's moment has the same shape, and a habit for resting or
// meeting can still learn that it belongs to ease and not to want.
const (
	Hunger = iota
	Unsafe
	Lonely
	Unproven
	Curious
	Shelter
	Company
	Chill
	Exposure
	Order
	Honesty
	Charity
	Industry
	Tradition
	Caution
	Near
	Rapport
	Skill
	Lack
	Stock
)

// Names labels each dimension for reports.
var Names = [Dims]string{
	"hunger", "unsafe", "lonely", "unproven", "curious",
	"shelter", "company", "chill", "exposure", "order",
	"honesty", "charity", "industry", "tradition", "caution",
	"near", "rapport", "skill", "lack", "stock",
}

// Signature is a point or direction in the space, each coordinate roughly
// in [-1, 1] with 0 meaning neutral.
type Signature [Dims]float64

// Varying marks the dimensions that may differ from one agent to the next.
// The moral dimensions are fixed: an agent's values are its own and identical
// across every candidate it weighs, so a signature that drifted on them would
// only say what the agent already says everywhere, and the dimension would
// cancel out of the choice. Values belong to the agent; signatures describe
// situations.
var Varying = func() [Dims]bool {
	var l [Dims]bool
	for i := range l {
		l[i] = i < Honesty || i > Caution
	}
	return l
}()

// Tuning constants, grouped so tuning is one edit.
const (
	// Epsilon is the norm below which a vector has no direction.
	Epsilon = 1e-9
	// MinNorm and MaxNorm bound a signature's length whenever one is
	// blended with another, so a signature can neither vanish into noise
	// nor grow without limit.
	MinNorm = 0.2
	MaxNorm = 2.0
	// Temperature is the base softmax temperature. It is divided by the
	// intensity of the situation, so a screaming moment is decided sharply
	// and a bland one loosely.
	Temperature = 0.15
	// ReachPenalty is subtracted from fit in proportion to how far out of
	// reach an action still is.
	ReachPenalty = 0.6
)

// Dot is the inner product.
func Dot(a, b Signature) float64 {
	var s float64
	for i := range a {
		s += a[i] * b[i]
	}
	return s
}

// Norm is the length.
func Norm(a Signature) float64 { return math.Sqrt(Dot(a, a)) }

// Unit is the direction, or the zero vector when there is none.
func Unit(a Signature) Signature {
	n := Norm(a)
	if n < Epsilon {
		return Signature{}
	}
	for i := range a {
		a[i] /= n
	}
	return a
}

// Cosine is the fit between two directions in [-1, 1]. Either vector having
// no direction gives 0, never NaN.
func Cosine(a, b Signature) float64 {
	na, nb := Norm(a), Norm(b)
	if na < Epsilon || nb < Epsilon {
		return 0
	}
	c := Dot(a, b) / (na * nb)
	return math.Max(-1, math.Min(1, c))
}

// ClampNorm scales s so its length lies in [lo, hi]. A vector with no
// direction is left alone.
func ClampNorm(s *Signature, lo, hi float64) {
	n := Norm(*s)
	if n < Epsilon {
		return
	}
	f := 1.0
	if n < lo {
		f = lo / n
	} else if n > hi {
		f = hi / n
	}
	if f != 1 {
		for i := range s {
			s[i] *= f
		}
	}
}

// Room is how many candidates a draw or a reading of one can weigh without
// asking for memory. It is comfortably more than the catalog holds, and a
// longer list than this is weighed on borrowed room instead - correctly,
// and no slower than it used to be.
const Room = 64

// weigh fills dst with the unnormalised softmax weights of eff at temp,
// and returns their total and the best of eff. dst must be as long as eff.
// Nothing here normalises: a draw does not need it, and the two readings
// that do divide by the total themselves.
func weigh(dst, eff []float64, temp float64) (total float64, best int) {
	for i, e := range eff {
		if e > eff[best] {
			best = i
		}
	}
	for i, e := range eff {
		dst[i] = math.Exp((e - eff[best]) / temp)
		total += dst[i]
	}
	return total, best
}

// Sample draws one index from a softmax over eff at the given temperature,
// consuming exactly one number from rng. A temperature at or below zero is
// an argmax. An empty eff returns -1.
func Sample(rng *rand.Rand, eff []float64, temp float64) int {
	if len(eff) == 0 {
		return -1
	}
	if temp <= 0 {
		best := 0
		for i, e := range eff {
			if e > eff[best] {
				best = i
			}
		}
		return best
	}
	// Written where the choice is made: a helper that handed this slice
	// back would be handing out the frame's own array, and the compiler
	// would have to put the array on the heap to allow it. So the choice
	// is made here, and the array stays where it was declared.
	var buf [Room]float64
	weights := buf[:]
	if len(eff) <= Room {
		weights = buf[:len(eff)]
	} else {
		weights = make([]float64, len(eff))
	}
	total, _ := weigh(weights, eff, temp)
	r := rng.Float64() * total
	for i, w := range weights {
		r -= w
		if r < 0 {
			return i
		}
	}
	return len(eff) - 1
}

// Softmax is the chance of each candidate being drawn by Sample, in the same
// order. A temperature at or below zero puts all of it on the best one. It
// is not what Sample draws with - Sample needs no normalising and does none -
// but it is the same distribution, which is what lets an onlooker be shown
// the odds an agent actually decided under.
func Softmax(eff []float64, temp float64) []float64 {
	if len(eff) == 0 {
		return nil
	}
	best := 0
	for i, e := range eff {
		if e > eff[best] {
			best = i
		}
	}
	p := make([]float64, len(eff))
	if temp <= 0 {
		p[best] = 1
		return p
	}
	total, _ := weigh(p, eff, temp)
	for i := range p {
		p[i] /= total
	}
	return p
}

// Entropy is the entropy in nats of the softmax over eff, a measure of how
// open the choice was. Zero means one action was certain.
func Entropy(eff []float64, temp float64) float64 {
	if temp <= 0 || len(eff) == 0 {
		return 0
	}
	// The same weights Softmax would have returned, divided by the same
	// total, written in the caller's own frame rather than in a slice made
	// to be read once and dropped. This is taken for every decision.
	var buf [Room]float64
	p := buf[:]
	if len(eff) <= Room {
		p = buf[:len(eff)]
	} else {
		p = make([]float64, len(eff))
	}
	total, _ := weigh(p, eff, temp)
	var h float64
	for _, w := range p {
		if q := w / total; q > 0 {
			h -= q * math.Log(q)
		}
	}
	return h
}

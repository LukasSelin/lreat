// Package habit is the vector space in which agents recognise what a moment
// calls for.
//
// A situation is a point in that space: how urgent each need is, what the
// agent has on hand, who is around, what it holds to be right, and how the
// candidate action sits relative to it. Each action carries a Signature in the
// same space, the kind of moment it belongs to. Choosing is recognition:
// which signature lies closest in direction to the moment at hand. No value
// is computed at the moment of choice. Value enters only afterwards, when a
// completed action's outcome pulls its signature toward the situation it was
// taken in or pushes it away, so habits are the residue of past reward.
//
// This package is a leaf. It knows nothing of agents, actions, or the world.
package habit

import (
	"math"
	"math/rand/v2"

	"lreat/core/need"
)

// Dims is the size of the space.
const Dims = 19

// MaxActions bounds the catalog so per-agent habit tables can be arrays and
// iteration order can never depend on a map.
const MaxActions = 32

// The dimensions. The first five are urgencies weighted by personality; the
// next six are stock and surroundings; five are what the agent holds to be
// right and how much reprisal it expects; the last three are patched per
// candidate action, because how near a thing is, how one feels about the
// person involved, and how able one believes oneself all depend on which
// action is being considered.
const (
	Hunger = iota
	Unsafe
	Lonely
	Unproven
	Curious
	Food
	Wood
	Wealth
	Shelter
	Company
	Order
	Honesty
	Charity
	Industry
	Tradition
	Caution
	Near
	Rapport
	Skill
)

// Names labels each dimension for reports.
var Names = [Dims]string{
	"hunger", "unsafe", "lonely", "unproven", "curious",
	"food", "wood", "wealth", "shelter", "company", "order",
	"honesty", "charity", "industry", "tradition", "caution",
	"near", "rapport", "skill",
}

// Signature is a point or direction in the space, each coordinate roughly
// in [-1, 1] with 0 meaning neutral.
type Signature [Dims]float64

// Learnable marks the dimensions that experience may move. The moral
// dimensions are frozen: an agent's values are its own and identical across
// every candidate it weighs, so if habits could learn them every habit would
// soon carry the same coordinates and the dimension would cancel out of the
// choice. Values belong to the agent; habits learn situations.
var Learnable = func() [Dims]bool {
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
	// MinNorm and MaxNorm bound a habit's length after every update, so a
	// habit can neither vanish into noise nor grow without limit.
	MinNorm = 0.2
	MaxNorm = 2.0
	// Temperature is the base softmax temperature. It is divided by the
	// intensity of the situation, so a screaming moment is decided sharply
	// and a bland one loosely.
	Temperature = 0.15
	// ReachPenalty is subtracted from fit in proportion to how far out of
	// reach an action still is.
	ReachPenalty = 0.6
	// Eta is the learning rate of a habit toward or away from a situation.
	Eta = 0.10
	// Lambda is the retention rate, a slow pull back to the prior.
	Lambda = 0.003
	// Gamma discounts credit given to the actions before the one that
	// completed, so a farm that fed a later meal earns some of that meal.
	Gamma = 0.5
	// TraceLen is how many recent actions share in a reward.
	TraceLen = 3
	// BaselineRate is how fast an agent's sense of an ordinary outcome moves.
	BaselineRate = 0.02
	// ActionBaselineRate is the same for its sense of an ordinary outcome of
	// one particular act, which it sees less often.
	ActionBaselineRate = 0.05
	// BaselineMix is how much of the expectation a lesson is judged against
	// is the act's own. Against the act's own baseline a lesson is about
	// when the act pays; against the agent's general one it is about
	// whether it pays at all. It is set to the act's own alone. With the
	// general baseline in the mix every act whose direct outcome is modest,
	// which is every instrumental act and above all standing guard, is
	// pushed a little further from its own moments each time it happens,
	// and the settlement loses public order and the births that need it.
	BaselineMix = 1.0
	// AdvantageClamp bounds the size of one lesson.
	AdvantageClamp = 0.5
	// ReachGain is how much doing an action brings it further into reach.
	ReachGain = 0.02
	// ProvenanceWeight is the share of a later reward that reaches the act
	// which made it possible: the meal's share for the act that grew the
	// food, the safety's share for the act that raised the roof.
	ProvenanceWeight = 0.7
	// LarderCap bounds how many units of food an agent remembers the
	// origin of.
	LarderCap = 16
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

// Sample draws one index from a softmax over eff at the given temperature,
// consuming exactly one number from rng. A temperature at or below zero is
// an argmax. An empty eff returns -1.
func Sample(rng *rand.Rand, eff []float64, temp float64) int {
	if len(eff) == 0 {
		return -1
	}
	best := 0
	for i, e := range eff {
		if e > eff[best] {
			best = i
		}
	}
	if temp <= 0 {
		return best
	}
	var total float64
	weights := make([]float64, len(eff))
	for i, e := range eff {
		weights[i] = math.Exp((e - eff[best]) / temp)
		total += weights[i]
	}
	r := rng.Float64() * total
	for i, w := range weights {
		r -= w
		if r < 0 {
			return i
		}
	}
	return len(eff) - 1
}

// Entropy is the entropy in nats of the softmax over eff, a measure of how
// open the choice was. Zero means one action was certain.
func Entropy(eff []float64, temp float64) float64 {
	if len(eff) == 0 || temp <= 0 {
		return 0
	}
	best := eff[0]
	for _, e := range eff {
		best = math.Max(best, e)
	}
	var total float64
	for _, e := range eff {
		total += math.Exp((e - best) / temp)
	}
	var h float64
	for _, e := range eff {
		p := math.Exp((e-best)/temp) / total
		if p > 0 {
			h -= p * math.Log(p)
		}
	}
	return h
}

// Update moves the learnable coordinates of h toward s when adv is positive
// and away when it is negative, by eta scaled by weight. It is the lesson of
// one outcome: this moment was, or was not, one that called for this act.
func Update(h *Signature, s Signature, adv, weight, eta float64) {
	step := eta * weight * adv
	for i := range h {
		if Learnable[i] {
			h[i] += step * (s[i] - h[i])
		}
	}
}

// Retain pulls every coordinate of h back toward prior by lambda. It is
// forgetting, and it is what keeps a habit from wandering off for good.
func Retain(h *Signature, prior Signature, lambda float64) {
	for i := range h {
		h[i] += lambda * (prior[i] - h[i])
	}
}

// Ledger is what an agent remembers about its state when it committed to
// an action, so that the outcome can be judged by what it wanted then. Food
// and Shelter are not valued; they are how the learner tells that an act
// produced or consumed something, so that credit can follow provenance.
type Ledger struct {
	Needs   need.Levels
	Urgency [need.Count]float64
	Food    float64
	Shelter float64
}

// Reward is the value of what changed between before and after, judged by
// the urgencies at the time of the decision and by personality. It is the
// only place in the habit layer where value is computed.
func Reward(before, after Ledger, personality need.Weights) float64 {
	var r float64
	for t := range before.Needs {
		r += (after.Needs[t] - before.Needs[t]) * before.Urgency[t] * personality[t]
	}
	return r
}

// Advantage is a reward relative to what the agent has come to expect,
// bounded so no single outcome can overwrite a lifetime of habit.
func Advantage(reward, baseline float64) float64 {
	return math.Max(-AdvantageClamp, math.Min(AdvantageClamp, reward-baseline))
}

// Step is one entry of an eligibility trace: an action taken and the moment
// it was taken in.
type Step struct {
	Index     int
	Situation Signature
}

// Trace is the short memory of recent actions that share in each reward.
// Newest first.
type Trace []Step

// Push records a step, dropping the oldest beyond TraceLen.
func (t *Trace) Push(s Step) {
	*t = append(*t, Step{})
	copy((*t)[1:], (*t)[:len(*t)-1])
	(*t)[0] = s
	if len(*t) > TraceLen {
		*t = (*t)[:TraceLen]
	}
}

// Weight is the share of a reward the k-th most recent step receives.
func Weight(k int) float64 { return math.Pow(Gamma, float64(k)) }

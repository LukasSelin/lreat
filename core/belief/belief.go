// Package belief holds what agents think is true and what they think is right.
//
// Nothing here is guaranteed to match the world. Agents plan with beliefs and
// act with reality, and the gap between the two is deliberate: it is what
// produces misplaced confidence, disappointment, gossip, and the requests
// people make of each other.
package belief

import "math"

// Norm is a value an agent may hold. Norms are not rules the simulation
// enforces; they are weights each agent applies to its own conduct and uses to
// judge everyone else's.
type Norm int

const (
	// Honesty condemns taking what is not yours.
	Honesty Norm = iota
	// Charity approves of giving and serving others at cost to yourself.
	Charity
	// Industry approves of labour and condemns idleness.
	Industry
	// Tradition approves of ritual and continuity, and is suspicious of novelty.
	Tradition
	NormCount
)

var normNames = [NormCount]string{"honesty", "charity", "industry", "tradition"}

func (n Norm) String() string { return normNames[n] }

// Norms is how strongly an agent holds each value, each in [0,1].
type Norms [NormCount]float64

// Valence is how much an act upholds (positive) or violates (negative) each
// norm. It is a property of the act, shared by everyone; what differs between
// agents is how much they care.
type Valence [NormCount]float64

// ConscienceScale converts moral weight into the same units as need
// satisfaction. Raising it makes the population more principled and less
// responsive to hunger.
const ConscienceScale = 0.35

// JudgementScale converts moral weight into a change of regard for the actor.
const JudgementScale = 0.5

// Conscience is what an act is worth morally to an agent, given what that
// agent values. Positive means it feels right.
func Conscience(v Valence, n Norms) float64 {
	var s float64
	for i := range v {
		s += v[i] * n[i]
	}
	return s * ConscienceScale
}

// Judgement is how much a witness's regard for an actor moves after seeing an
// act. A witness who does not hold the violated norm barely reacts.
func Judgement(v Valence, witness Norms) float64 {
	var s float64
	for i := range v {
		s += v[i] * witness[i]
	}
	return s * JudgementScale
}

// Guilt is the esteem an actor loses for doing something their own values
// condemn. Acts they consider virtuous produce no guilt, only the esteem the
// act itself grants.
func Guilt(v Valence, n Norms) float64 {
	c := Conscience(v, n)
	if c >= 0 {
		return 0
	}
	return -c
}

// LearnRate is how fast a single outcome moves a belief toward evidence.
const LearnRate = 0.25

// Update moves a belief toward observed evidence.
func Update(prior, evidence, rate float64) float64 {
	return prior + (evidence-prior)*rate
}

// Drift moves a value toward a target by at most step. Used for the slow
// spread of norms between people who spend time together.
func Drift(value, target, step float64) float64 {
	d := target - value
	if math.Abs(d) <= step {
		return target
	}
	if d > 0 {
		return value + step
	}
	return value - step
}

// Clamp limits x to [0,1].
func Clamp(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

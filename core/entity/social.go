package entity

import (
	"math"
	"math/rand/v2"
)

// Temperament is an agent's stance toward other people. It is not a need and
// not a value; it is the disposition that decides how a stranger is received
// before anything is known about them, and how much company is worth once it
// is had.
type Temperament struct {
	// Trust is the good faith extended to strangers and the weight given to
	// what friends say about people. Low trust means every relationship must
	// be built from nothing in person.
	Trust float64
	// Warmth is how much belonging a meeting yields and how quickly bonds
	// strengthen. Cold agents need fewer people and take longer to warm to them.
	Warmth float64
}

// RandomTemperament draws a disposition around the middle.
func RandomTemperament(r *rand.Rand) Temperament {
	return Temperament{
		Trust:  clamp01(0.5 + r.NormFloat64()*0.2),
		Warmth: clamp01(0.5 + r.NormFloat64()*0.2),
	}
}

// InheritTemperament returns a child's disposition drawn from a parent's.
func InheritTemperament(parent Temperament, r *rand.Rand) Temperament {
	return Temperament{
		Trust:  clamp01(parent.Trust + r.NormFloat64()*0.1),
		Warmth: clamp01(parent.Warmth + r.NormFloat64()*0.1),
	}
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// Affinity is how alike two agents are in what they value and what they
// want, in [0,1]. Like draws to like: it sets the tone of a first meeting and
// the quality of every meeting after.
func Affinity(a, o *Agent) float64 {
	var norms float64
	for i := range a.Norms {
		norms += math.Abs(a.Norms[i] - o.Norms[i])
	}
	norms /= float64(len(a.Norms))

	var wants float64
	for i := range a.Personality {
		wants += math.Abs(a.Personality[i] - o.Personality[i])
	}
	wants /= float64(len(a.Personality)) * 1.7 // weights span [0.3,2]

	return clamp01(1 - (0.6*norms + 0.4*wants))
}

// Know returns the bond record for id, creating one if needed. It is the
// exported face of bond for packages that seed opinions.
func (a *Agent) Know(id ID, tick int) *Bond { return a.bond(id, tick) }

// HasMet reports whether the agent has met id in person.
func (a *Agent) HasMet(id ID) bool {
	b := a.Look(id)
	return b != nil && b.Met > 0
}

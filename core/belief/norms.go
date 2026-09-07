package belief

import "math/rand/v2"

// RandomNorms draws a set of values around a moderate baseline. Two people
// raised in the same place still disagree about what matters, which is where
// moral disagreement in the settlement starts.
func RandomNorms(r *rand.Rand) Norms {
	var n Norms
	for i := range n {
		n[i] = Clamp(0.5 + r.NormFloat64()*0.25)
	}
	return n
}

// Inherit returns a child's values, drawn from a parent's with drift. Culture
// passes down but never exactly, so a community's values move over
// generations without anyone deciding to change them.
func Inherit(parent Norms, r *rand.Rand) Norms {
	var n Norms
	for i := range n {
		n[i] = Clamp(parent[i] + r.NormFloat64()*0.12)
	}
	return n
}

// ContagionStep is how far one agent's values move toward another's during a
// single sustained interaction. Small, because people change slowly.
const ContagionStep = 0.004

// Spread moves the listener's values a step toward the speaker's. Repeated
// across a neighborhood this is what makes local cultures cohere.
func Spread(listener *Norms, speaker Norms, weight float64) {
	for i := range listener {
		listener[i] = Clamp(Drift(listener[i], speaker[i], ContagionStep*weight))
	}
}

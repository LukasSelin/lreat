// Package need models the leaky hierarchy of needs that drives every agent.
//
// Each agent has a satisfaction level in [0,1] per tier. Urgency grows with the
// deficit of a tier, but higher tiers are damped by how well the tiers below
// them are met. The damping is never total: a fraction Leak of the urgency
// always survives, so a hungry agent still occasionally socializes and a poor
// one can still be curious. That leak is what keeps the population from
// collapsing into identical behavior.
package need

import "math"

// Tier is one level of the hierarchy, ordered from most to least basic.
type Tier int

const (
	Physiological Tier = iota
	Safety
	Belonging
	Esteem
	Actualization
	Count
)

var names = [Count]string{"physiological", "safety", "belonging", "esteem", "actualization"}

func (t Tier) String() string { return names[t] }

// Tiers returns every tier in ascending order.
func Tiers() [Count]Tier {
	var ts [Count]Tier
	for i := range ts {
		ts[i] = Tier(i)
	}
	return ts
}

// Levels holds an agent's satisfaction per tier, each in [0,1].
type Levels [Count]float64

// Weights holds an agent's personal multiplier per tier. 1.0 is neutral.
// Personality lives here: a risk-taker has low Safety and high Esteem weights.
type Weights [Count]float64

// Neutral returns weights of 1.0 for every tier.
func Neutral() Weights {
	var w Weights
	for i := range w {
		w[i] = 1
	}
	return w
}

// exponent shapes how urgency grows with deficit. Values below 1 make urgency
// rise sharply on small deficits, values above 1 keep it low until the deficit
// is large. Lower tiers are steeper: mild hunger already matters, mild
// boredom does not.
var exponent = [Count]float64{0.5, 0.7, 1.0, 1.4, 2.0}

// Leak is the fraction of a higher tier's urgency that survives when every
// lower tier is completely unmet.
const Leak = 0.2

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

// Foundation is the mean satisfaction of the tiers below t. It is 1 for the
// bottom tier, which has nothing beneath it.
func Foundation(t Tier, l Levels) float64 {
	if t == 0 {
		return 1
	}
	var sum float64
	for i := Tier(0); i < t; i++ {
		sum += Clamp(l[i])
	}
	return sum / float64(t)
}

// Urgency returns how pressing tier t is for an agent with levels l.
// It is the deficit shaped by the tier's curve, damped by the foundation
// beneath it, with Leak guaranteeing it never fully vanishes.
func Urgency(t Tier, l Levels) float64 {
	deficit := Clamp(1 - l[t])
	raw := math.Pow(deficit, exponent[t])
	return raw * (Leak + (1-Leak)*Foundation(t, l))
}

// Urgencies returns Urgency for every tier.
func Urgencies(l Levels) [Count]float64 {
	var u [Count]float64
	for t := range u {
		u[t] = Urgency(Tier(t), l)
	}
	return u
}

// Add raises tier t by d and clamps the result.
func (l *Levels) Add(t Tier, d float64) {
	l[t] = Clamp(l[t] + d)
}

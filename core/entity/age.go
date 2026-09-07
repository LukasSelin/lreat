package entity

import "math"

// The span of a life, in ticks. An agent grows into its body by Maturity,
// carries it whole until Prime, and declines from there; Lifespan is the age
// by which a body has usually failed, not a wall it hits. The numbers are
// scaled so that a settlement turns over several generations in a run long
// enough to develop, which is the point: what a population believes has to
// outlive the people who first believed it.
const (
	Maturity = 500
	Prime    = 2800
	Lifespan = 5200
)

// Age returns how many ticks the agent has been alive.
func (a *Agent) Age(tick int) int { return tick - a.Born }

// AgeFactor is how much of its vitality a body of this age can bring to bear.
// A child has half a body and grows into the rest; an elder gives it back.
// Nothing here is a cliff: the young are slower rather than helpless, and the
// old are frail rather than finished.
func AgeFactor(age int) float64 {
	switch {
	case age < 0:
		return 1
	case age < Maturity:
		return 0.5 + 0.5*float64(age)/Maturity
	case age < Prime:
		return 1
	}
	decline := float64(age-Prime) / float64(Lifespan-Prime)
	return math.Max(0.25, 1-0.75*decline)
}

// Frailty is the per-tick chance that a body past its prime simply fails. It
// is zero until Prime and then climbs quadratically, so old age is a rising
// risk rather than an appointment: most agents die somewhere in their decline
// and a few live well past Lifespan.
func Frailty(age int) float64 {
	if age < Prime {
		return 0
	}
	d := float64(age-Prime) / float64(Lifespan-Prime)
	return 0.004 * d * d
}

// Adult reports whether an agent of this age has grown into its own body.
func Adult(age int) bool { return age >= Maturity }

// Fertile reports whether an agent of this age can have children. Bearing
// belongs to the years between growing up and starting to decline.
func Fertile(age int) bool { return age >= Maturity && age < Prime }

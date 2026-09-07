package entity

import (
	"math"

	"lreat/core/clock"
)

// The span of a life, in years. An agent grows into its body by Maturity,
// carries it whole until Prime, and declines from there; Lifespan is the age
// by which a body has usually failed, not a wall it hits.
//
// These are human. A childhood is fifteen years, the bearing years run to
// forty, and a body that is kept has given out by sixty-five. They were not
// always: while the year was a hundred ticks long a life had to be counted
// in ticks, and it came out at five, twenty-eight and fifty-two - a
// settlement whose children were grown before they could walk to the next
// field. What made the human figures affordable is the calendar, which is
// worth stating plainly: what a run has to cover is a number of generations,
// and a generation is a span of years, so lengthening the year is what buys
// a childhood.
//
// A childhood is not free, and the cost is the point of having one. Under
// the old span a settlement carried about a tenth of its people as children;
// under this one it carries a quarter, each of them eating and each working
// at half a body and growing into the rest. A settlement that cannot feed
// its children does not grow, and now it has children to fail to feed.
const (
	Maturity = 15 * clock.Year
	Prime    = 40 * clock.Year
	Lifespan = 65 * clock.Year
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

// endRisk is the daily chance a body right at the end of its decline gives
// out. It is set against the length of the decline rather than chosen: what
// has to hold is how much of a decline a body is likely to survive, and that
// is the risk multiplied by the days there are to run it in. Lengthen the
// year and this falls, or old age arrives the same number of days after a
// prime that is now decades away.
var endRisk = 9.6 / float64(Lifespan-Prime)

// Frailty is the daily chance that a body past its prime simply fails. It is
// zero until Prime and then climbs quadratically, so old age is a rising
// risk rather than an appointment: most agents die somewhere in their decline
// and a few live well past Lifespan.
func Frailty(age int) float64 {
	if age < Prime {
		return 0
	}
	d := float64(age-Prime) / float64(Lifespan-Prime)
	return endRisk * d * d
}

// Adult reports whether an agent of this age has grown into its own body.
func Adult(age int) bool { return age >= Maturity }

// Fertile reports whether an agent of this age can have children. Bearing
// belongs to the years between growing up and starting to decline.
func Fertile(age int) bool { return age >= Maturity && age < Prime }

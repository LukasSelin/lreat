package entity

import "lreat/core/clock"

// The span of a human life, in years. A childhood is fifteen years, the
// bearing years run to forty, and a body that is kept has given out by
// sixty-five. They were not always: while the year was a hundred ticks long
// a life had to be counted in ticks, and it came out at five, twenty-eight
// and fifty-two - a settlement whose children were grown before they could
// walk to the next field. What made the human figures affordable is the
// calendar, which is worth stating plainly: what a run has to cover is a
// number of generations, and a generation is a span of years, so
// lengthening the year is what buys a childhood.
//
// A childhood is not free, and the cost is the point of having one. Under
// the old span a settlement carried about a tenth of its people as children;
// under this one it carries a quarter, each of them eating and each working
// at half a body and growing into the rest. A settlement that cannot feed
// its children does not grow, and now it has children to fail to feed.
//
// The span itself is a Life, and every creature has one of its own; see
// species.go. What is here is the people's, kept under the old names so
// that everything that asks about a person's years asks as it always did.
const (
	Maturity = 15 * clock.Year
	Prime    = 40 * clock.Year
	Lifespan = 65 * clock.Year
)

// Age returns how many ticks the agent has been alive.
func (a *Agent) Age(tick int) int { return tick - a.Born }

// AgeFactor is Life.Factor for a person.
func AgeFactor(age int) float64 { return Human.Life.Factor(age) }

// Frailty is Life.Frailty for a person.
func Frailty(age int) float64 { return Human.Life.Frailty(age) }

// Adult is Life.Adult for a person.
func Adult(age int) bool { return Human.Life.Adult(age) }

// Fertile is Life.Fertile for a person.
func Fertile(age int) bool { return Human.Life.Fertile(age) }

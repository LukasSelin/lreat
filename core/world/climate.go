package world

import (
	"math"
	"math/rand/v2"

	"lreat/core/ontology"
)

// The climate. A settlement that is founded in one weather and lives in it
// forever is a settlement with nothing to save for. Here the year turns:
// things grow fast in the warm half and slowly in the cold one, so the wild
// food an idle summer left standing is most of the wild food there is in
// February, and a roof is worth most in the months a body loses heat
// fastest. On top of the year sit two slower wanderings, so no two winters
// are the same and no two decades are either.
//
// The place is temperate. Midsummer is warm and midwinter bites without
// being fatal on its own; what kills is a winter met without a roof.

// Year is the length of a year in ticks, and Season a quarter of it. A life
// of Lifespan ticks is some fifty of these, so a run long enough to develop
// sees dozens of winters.
//
// The calendar is the ontology's, because what grows is measured in it and
// the ontology is where what grows is stated. These are that same calendar
// under the names the rest of the world already calls it by.
const (
	Year   = ontology.Year
	Season = ontology.Season
)

// The shape of the year. MeanTemp is the annual mean in degrees, Swing half
// the distance from midwinter to midsummer. Tick zero is early spring:
// people arrive with the growing season ahead of them, not behind them.
const (
	MeanTemp = 10.0
	Swing    = 12.0
)

// The two wanderings, each an AR(1) process. Drift is the slow one, a run of
// kind or unkind decades; Spell is the fast one, the warm week and the cold
// snap. Keep is how much of the anomaly survives a tick - one part in the
// process's timescale is shed - and Shock the standard deviation of what is
// added, chosen so the anomaly settles at a standard deviation of Wander.
const (
	driftTime   = 800.0 // ticks, eight years
	driftWander = 1.2   // degrees
	spellTime   = 6.0   // ticks
	spellWander = 2.5   // degrees
)

var (
	driftKeep  = 1 - 1/driftTime
	driftShock = driftWander * math.Sqrt(1-driftKeep*driftKeep)
	spellKeep  = 1 - 1/spellTime
	spellShock = spellWander * math.Sqrt(1-spellKeep*spellKeep)
)

// Climate is the weather of the whole map at one tick. It is one temperature
// for everywhere: the settlement is small enough that the difference between
// its ends is nothing beside the difference between its seasons.
type Climate struct {
	Temp  float64 // this tick's temperature, in degrees
	Drift float64 // the slow anomaly, a kind or unkind decade
	Spell float64 // the fast anomaly, a warm week or a cold snap
}

// NewClimate is the weather a world is founded in: an ordinary early spring,
// with neither wandering underway.
func NewClimate() Climate {
	return Climate{Temp: seasonal(0)}
}

// seasonal is the temperature the turning year alone would give at tick t.
func seasonal(tick int) float64 {
	return MeanTemp + Swing*math.Sin(2*math.Pi*float64(tick)/Year)
}

// Advance moves the weather on one tick. It consumes exactly two numbers
// from rng, so a seed still reproduces a run.
func (c *Climate) Advance(tick int, rng *rand.Rand) {
	c.Drift = driftKeep*c.Drift + driftShock*rng.NormFloat64()
	c.Spell = spellKeep*c.Spell + spellShock*rng.NormFloat64()
	c.Temp = seasonal(tick) + c.Drift + c.Spell
}

// The thresholds the living world reads temperature by. Green things grow
// at their slowest below Frost and at their fullest from Thrive up. Cold
// begins to be felt below Mild and presses as hard as it ever does at
// Bitter.
const (
	Frost  = 4.0
	Thrive = 14.0
	Mild   = 12.0
	Bitter = -5.0
)

// WinterGrowth is what the land still gives with the year at its coldest.
// It is not zero: this is a temperate place, not an arctic one, and a winter
// that stops the forest dead is a winter the settlement cannot get through.
// At zero the year became a siege - median population over 24 seeds fell
// from 128 to 21, with the first extinction the sweep had seen - and at 0.35
// seed 1 still bred not once in 2500 ticks and died with its founders. Half
// is a winter that pauses a settlement's growth without ending it: the land
// gives about two thirds of the old flat rate in the depth of it and a third
// again as much at midsummer.
const WinterGrowth = 0.5

// growthNorm holds the year's total growth where it was before the seasons
// existed. Averaged over a year the ramp between Frost and Thrive is worth
// 0.527 of full growth, so the whole curve averages WinterGrowth + 0.527 of
// what is above it, and full growth is worth the reciprocal of that: what
// changed is when a forest grows, not how much it grows in a year.
const growthNorm = 1 / (WinterGrowth + (1-WinterGrowth)*0.527)

// Growth is how much the season lets green things grow, in multiples of the
// old year-round rate. It is two thirds of that in the depth of winter and a
// third again as much at midsummer, and averages 1 over a year, so tuning
// done before the seasons still holds.
func (c Climate) Growth() float64 {
	return growthNorm * (WinterGrowth + (1-WinterGrowth)*ramp(c.Temp, Frost, Thrive))
}

// Chill is how hard the cold presses on a body, 0 in mild weather and 1 in
// the bitterest cold this place sees.
func (c Climate) Chill() float64 {
	return 1 - ramp(c.Temp, Bitter, Mild)
}

// ramp is x placed on [0,1] between lo and hi, clamped at both ends.
func ramp(x, lo, hi float64) float64 {
	switch {
	case x <= lo:
		return 0
	case x >= hi:
		return 1
	}
	return (x - lo) / (hi - lo)
}

// Names of the four seasons, from the one tick zero falls in.
var seasons = [4]string{"spring", "summer", "autumn", "winter"}

// SeasonOf names the quarter of the year tick falls in.
func SeasonOf(tick int) string {
	q := (tick % Year) / Season
	if q < 0 {
		q += 4
	}
	return seasons[q]
}

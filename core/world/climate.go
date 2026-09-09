package world

import (
	"math"
	"math/rand/v2"

	"lreat/core/clock"
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

// Year is the length of a year, and Season a quarter of it. A life is some
// fifty of these, so a run long enough to develop sees dozens of winters.
//
// The calendar itself is package clock's, where a tick is a day. The
// ontology names it too, because what grows is measured in it; these are
// that same calendar under the names the weather already calls it by.
const (
	Year   = clock.Year
	Season = clock.Season
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
	driftTime   = 8.0 * clock.Year // a run of kind or unkind decades
	driftWander = 1.2              // degrees
	spellTime   = 1.0 * clock.Week // a warm week, a cold snap
	spellWander = 2.5              // degrees
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
	Temp  float64 // this tick's temperature, in degrees, in the temperate latitudes
	Drift float64 // the slow anomaly, a kind or unkind decade
	Spell float64 // the fast anomaly, a warm week or a cold snap
	// rows is how many rows the map has, and globe whether it is one. On a
	// globe the weather is read by latitude - see TempAt - and Temp is the
	// weather of the temperate latitudes, which is the weather a valley
	// has everywhere.
	rows  int
	globe bool
}

// NewClimate is the weather a world is founded in: an ordinary early spring,
// with neither wandering underway.
func NewClimate() Climate {
	return Climate{Temp: seasonal(0)}
}

// NewClimateOn is the weather a world of the given shape is founded in.
func NewClimateOn(cfg Config) Climate {
	c := NewClimate()
	c.rows, c.globe = cfg.Height, cfg.Wrap
	return c
}

// The globe's weather. Temperate is the latitude the default map's weather
// is the weather of; a globe reads that weather there, warmer toward the
// middle and colder toward the poles by LatSwing across the whole of the
// curve, with the year's swing turning over in the south and fading out
// at the equator.
const (
	Temperate = 45.0
	LatSwing  = 30.0
)

// latitude is the latitude of row y in degrees, from ninety at the top
// row to minus ninety at the bottom.
func (c Climate) latitude(y int) float64 {
	return 90 - 180*(float64(y)+0.5)/float64(c.rows)
}

// warmth is what a latitude adds to the temperate mean the year round.
func warmth(lat float64) float64 {
	return LatSwing * (math.Cos(lat*math.Pi/180) - math.Cos(Temperate*math.Pi/180))
}

// TempAt is this tick's temperature on row y. On a valley it is Temp
// everywhere, to the bit.
func (c Climate) TempAt(y int) float64 {
	if !c.globe {
		return c.Temp
	}
	lat := c.latitude(y)
	hemi := math.Copysign(math.Min(1, math.Abs(lat)/Temperate), lat)
	season := c.Temp - MeanTemp - c.Drift - c.Spell
	return MeanTemp + hemi*season + c.Drift + c.Spell + warmth(lat)
}

// MeanAt is the mean temperature of row y over a year.
func (c Climate) MeanAt(y int) float64 {
	if !c.globe {
		return MeanTemp
	}
	return MeanTemp + warmth(c.latitude(y))
}

// GrowthAt is Growth on row y, and ChillAt is Chill there.
func (c Climate) GrowthAt(y int) float64 {
	if !c.globe {
		return c.Growth()
	}
	return growthOf(c.TempAt(y))
}

// ChillAt is Chill on row y.
func (c Climate) ChillAt(y int) float64 {
	if !c.globe {
		return c.Chill()
	}
	return chillOf(c.TempAt(y))
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
func (c Climate) Growth() float64 { return growthOf(c.Temp) }

func growthOf(temp float64) float64 {
	return growthNorm * (WinterGrowth + (1-WinterGrowth)*ramp(temp, Frost, Thrive))
}

// Chill is how hard the cold presses on a body, 0 in mild weather and 1 in
// the bitterest cold this place sees.
func (c Climate) Chill() float64 { return chillOf(c.Temp) }

func chillOf(temp float64) float64 { return 1 - ramp(temp, Bitter, Mild) }

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

// SeasonOf names the quarter of the year tick falls in. It is the calendar's
// answer; this is here so that callers reading the weather need not reach
// past it for the date.
func SeasonOf(tick int) clock.Quarter { return clock.SeasonOf(tick) }

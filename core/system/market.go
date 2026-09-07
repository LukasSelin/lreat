package system

import (
	"math"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/world"
)

// basePrice is what each good costs when the market holds a normal stock.
var basePrice = [entity.GoodCount]float64{
	entity.Food:  1,
	entity.Wood:  0.5,
	entity.Tools: 3,
	entity.Stone: 1.5,
	entity.Meals: 2,
}

// spoilage is the fraction of stock lost per tick. Food and meals spoil
// less in a settlement with granaries; see Modifiers.Keeping.
var spoilage = [entity.GoodCount]float64{
	entity.Food:  0.01,
	entity.Wood:  0.001,
	entity.Tools: 0.0005,
	entity.Stone: 0,
	entity.Meals: 0.003,
}

// perishable is what a granary keeps, and what the cold keeps.
var perishable = [entity.GoodCount]bool{entity.Food: true, entity.Meals: true}

// ColdKeeping is how much of a perishable's spoilage the bitterest cold
// stops. A cold store is the oldest one there is: what the year takes from
// the settlement in the growing it stops doing, it gives back a little of in
// the keeping. It is what makes an autumn surplus worth holding rather than
// selling, and it is the one thing in the world that gets better as the
// weather gets worse.
const ColdKeeping = 0.6

// Keeping is the share of the usual spoilage a perishable actually suffers
// here today: what the settlement's granaries stop, and what the weather
// stops on top of that.
func Keeping(w *world.World) float64 {
	return w.Mods.Keeping * (1 - ColdKeeping*w.Climate.Chill())
}

// RoofKeeping is how much of a private larder's spoilage a roof of one's
// own stops. A house is not a granary and a granary is not a house: the
// settlement's stores keep the market's food, and what keeps an agent's is
// the weather and whatever it has built over its own head.
const RoofKeeping = 0.5

// Larder is the share of the usual spoilage the food in an agent's own
// hands suffers. Nothing carried is kept for ever now, which is what makes
// a full larder in June worth less than the same larder in January and
// gives the cold something to be good for close to home.
func Larder(w *world.World, a *entity.Agent) float64 {
	return (1 - ColdKeeping*w.Climate.Chill()) * (1 - RoofKeeping*need.Clamp(a.Shelter))
}

// Spoil rots what an agent is carrying, by the same rates the market's
// stock rots at.
func Spoil(w *world.World, a *entity.Agent) {
	k := Larder(w, a)
	for g := range a.Inventory {
		if perishable[g] {
			a.Inventory[g] *= 1 - spoilage[g]*k
		}
	}
}

// MarketStep moves prices against stock and lets goods spoil. Prices lag
// supply, so gluts and shortages both overshoot, which is what gives traders
// something to notice.
func MarketStep(w *world.World) {
	m := &w.Market
	for g := range m.Stock {
		loss := spoilage[g]
		if perishable[g] {
			loss *= Keeping(w)
		}
		m.Stock[g] = math.Max(0, m.Stock[g]*(1-loss))
		target := basePrice[g] * math.Sqrt(10/(m.Stock[g]+2))
		target = math.Max(0.2*basePrice[g], math.Min(5*basePrice[g], target))
		m.Price[g] += (target - m.Price[g]) * 0.05
	}
}

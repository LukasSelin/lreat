package system

import (
	"math"

	"lreat/core/entity"
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

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

// perishable is what a granary keeps.
var perishable = [entity.GoodCount]bool{entity.Food: true, entity.Meals: true}

// MarketStep moves prices against stock and lets goods spoil. Prices lag
// supply, so gluts and shortages both overshoot, which is what gives traders
// something to notice.
func MarketStep(w *world.World) {
	m := &w.Market
	for g := range m.Stock {
		loss := spoilage[g]
		if perishable[g] {
			loss *= w.Mods.Keeping
		}
		m.Stock[g] = math.Max(0, m.Stock[g]*(1-loss))
		target := basePrice[g] * math.Sqrt(10/(m.Stock[g]+2))
		target = math.Max(0.2*basePrice[g], math.Min(5*basePrice[g], target))
		m.Price[g] += (target - m.Price[g]) * 0.05
	}
}

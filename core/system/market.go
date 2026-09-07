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

// LarderSpoil is the share of what an agent carries that goes off each
// tick in mild weather. It is a fifth of what the same food loses sitting
// in the market, and flat across food and meals, because what is carried
// is eaten within days while the market's stock sits out whole seasons.
//
// It was the market's own rates first, and that was far too much: a third
// again on top of what an agent eats, on the loop the whole economy runs
// on. Over 24 seeds to 6000 ticks it took the median settlement from 94 to
// 30 and killed two. The point was never to punish a full larder in June,
// only to make one in January worth more.
const LarderSpoil = 0.002

// Larder is what an agent's own food loses this tick. Nothing a granary
// does reaches it - the settlement's stores keep the settlement's food -
// so the cold is the whole of its mercy, and in the deep of winter it
// keeps two and a half times as well as it does in the summer.
func Larder(w *world.World) float64 {
	return LarderSpoil * (1 - ColdKeeping*w.Climate.Chill())
}

// Spoil rots what an agent is carrying.
func Spoil(w *world.World, a *entity.Agent) {
	k := Larder(w)
	for g := range a.Inventory {
		if perishable[g] {
			a.Inventory[g] *= 1 - k
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

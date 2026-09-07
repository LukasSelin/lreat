package action

import (
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// Exchanging is one verb. Selling and buying are the same act with the
// sides swapped: at the market, a good for coin or coin for a good, at the
// market's price. The ontology says which way it runs; the terms say what
// is handed over and at what point it is worth the walk. Coin is a thing
// in the trees so that it can one day be minted, given, and stolen; in the
// pack it is still the one scalar, Wealth, and exchange is the only act
// that touches it.

// sale is what a seller lets go of: everything of a good above what they
// keep, once they have at least enough for the trip to be worth making.
type sale struct {
	Good        entity.Good
	Keep, Least float64
}

// purchase is what a buyer comes for: some of a good, when they have less
// than they want of it and the market has it to sell.
type purchase struct {
	Good         entity.Good
	Want, Amount float64
}

// terms is what an exchange hands over, by key. A selling lists sales, a
// buying one purchase.
type terms struct {
	Name  string
	Sells []sale
	Buys  purchase
	// Worth is what the exchange is worth, by need tier. The esteem it
	// promises is the esteem it gives; the rest comes with what was got.
	Worth need.Levels
}

var trades = map[string]terms{
	// Savings buy safety; being a seller of note buys a little esteem.
	"exchange/material>coin@market": {
		Name: "sell",
		Sells: []sale{
			{entity.Food, 3, 4},
			{entity.Tools, 0, 1},
			{entity.Meals, 2, 3},
			{entity.Stone, 3, 4},
		},
		Worth: need.Levels{need.Safety: 0.08, need.Esteem: 0.03},
	},
	"exchange/coin>provision@market": {
		Name:  "buy food",
		Buys:  purchase{entity.Food, 1, 1},
		Worth: need.Levels{need.Physiological: 0.3},
	},
}

// exchanging is the act the ontology entails for an exchange with terms,
// carried out from them, or nil where there are none or the terms do not
// fit which way the exchange runs.
func exchanging(in ontology.Instance) *Def {
	t, ok := trades[in.Key]
	if !ok {
		return nil
	}
	selling := in.Schema.Output == ontology.Coin
	buying := len(in.Schema.Inputs) == 1 && in.Schema.Inputs[0] == ontology.Coin
	switch {
	case selling && len(t.Sells) > 0 && t.Buys.Amount == 0:
	case buying && len(t.Sells) == 0 && t.Buys.Amount > 0 && goods[in.Schema.Output] == t.Buys.Good:
	default:
		return nil
	}
	d := &Def{Name: t.Name, Ticks: in.Ticks, Target: atMarket}
	d.Expect = func(*entity.Agent, *world.World, entity.Pos) need.Levels { return t.Worth }
	if selling {
		d.Available = func(a *entity.Agent, _ *world.World) bool {
			for _, s := range t.Sells {
				if a.Inventory[s.Good] >= s.Least {
					return true
				}
			}
			return false
		}
		d.Apply = func(a *entity.Agent, w *world.World) {
			var earned float64
			for _, s := range t.Sells {
				if surplus := a.Inventory[s.Good] - s.Keep; surplus > 0 {
					a.Inventory[s.Good] -= surplus
					w.Market.Stock[s.Good] += surplus
					earned += surplus * w.Market.Price[s.Good]
				}
			}
			a.Wealth += earned
			a.Needs.Add(need.Esteem, t.Worth[need.Esteem])
			w.Emit(event.Traded, a.ID, 0, "%s sold goods for %.1f", a.Name, earned)
		}
		return d
	}
	b := t.Buys
	d.Available = func(a *entity.Agent, w *world.World) bool {
		return a.Inventory[b.Good] < b.Want &&
			w.Market.Stock[b.Good] >= b.Amount &&
			a.Wealth >= w.Market.Price[b.Good]*b.Amount
	}
	d.Apply = func(a *entity.Agent, w *world.World) {
		p := w.Market.Price[b.Good] * b.Amount
		a.Wealth -= p
		a.Inventory[b.Good] += b.Amount
		w.Market.Stock[b.Good] -= b.Amount
		a.Needs.Add(need.Esteem, t.Worth[need.Esteem])
		w.Emit(event.Traded, a.ID, 0, "%s bought %s for %.2f", a.Name, b.Good, p)
	}
	return d
}

// trade is the exchange the ontology entails under key, for the acts the
// rest of the package refers to by name.
func trade(key string) *Def {
	d := exchanging(instances[key])
	if d == nil {
		panic("action: no terms for " + key)
	}
	return d
}

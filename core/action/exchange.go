package action

import (
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// Exchanging is one verb. Selling and buying are the same act with the
// sides swapped: at the market, move a material from the pack to the
// shelf and coin the other way, or the reverse, at the market's price.
// The ontology says which way it runs; the terms say what is handed over
// and at what point it is worth the walk. Coin is a thing in the trees so
// that it can one day be minted, given, and stolen; in the pack it is the
// purse, and the market's purse is bottomless.

// sale is what a seller lets go of: everything of a material above what
// they keep, once they have at least enough for the trip to be worth
// making.
type sale struct {
	Of          *ontology.Class
	Keep, Least float64
}

// purchase is what a buyer comes for: some of a material, when they have
// less than they want of it and the market has it to sell.
type purchase struct {
	Of           *ontology.Class
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
			{ontology.Provision, 3, 4},
			{ontology.Tool, 0, 1},
			{ontology.Meal, 2, 3},
			{ontology.Stone, 3, 4},
		},
		Worth: need.Levels{need.Safety: 0.08, need.Esteem: 0.03},
	},
	"exchange/coin>provision@market": {
		Name:  "buy food",
		Buys:  purchase{ontology.Provision, 1, 1},
		Worth: need.Levels{need.Physiological: 0.3},
	},
}

// price is what the market asks for a material.
func price(w *world.World, m *ontology.Class) float64 {
	g, _ := good(m)
	return w.Market.Price[g]
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
		for _, s := range t.Sells {
			if _, ok := good(s.Of); !ok {
				return nil
			}
		}
	case buying && len(t.Sells) == 0 && t.Buys.Amount > 0 && t.Buys.Of.IsA(in.Schema.Output):
		if _, ok := good(t.Buys.Of); !ok {
			return nil
		}
	default:
		return nil
	}
	d := &Def{Name: t.Name, Ticks: in.Ticks, Target: atMarket}
	d.Expect = func(*entity.Agent, *world.World, entity.Pos) need.Levels { return t.Worth }
	if selling {
		d.Available = func(a *entity.Agent, _ *world.World) bool {
			for _, s := range t.Sells {
				if held, _ := pack(a, s.Of); held.Held() >= s.Least {
					return true
				}
			}
			return false
		}
		d.Apply = func(a *entity.Agent, w *world.World) {
			var earned float64
			for _, s := range t.Sells {
				mine, _ := pack(a, s.Of)
				if surplus := mine.Held() - s.Keep; surplus > 0 {
					theirs, _ := shelf(w, s.Of)
					mine.Move(-surplus)
					theirs.Move(surplus)
					earned += surplus * price(w, s.Of)
				}
			}
			purse, _ := pack(a, ontology.Coin)
			purse.Move(earned)
			a.Needs.Add(need.Esteem, t.Worth[need.Esteem])
			w.Emit(event.Traded, a.ID, 0, "%s sold goods for %.1f", a.Name, earned)
		}
		return d
	}
	b := t.Buys
	d.Available = func(a *entity.Agent, w *world.World) bool {
		mine, _ := pack(a, b.Of)
		theirs, _ := shelf(w, b.Of)
		purse, _ := pack(a, ontology.Coin)
		return mine.Held() < b.Want && theirs.Held() >= b.Amount && purse.Held() >= price(w, b.Of)*b.Amount
	}
	d.Apply = func(a *entity.Agent, w *world.World) {
		mine, _ := pack(a, b.Of)
		theirs, _ := shelf(w, b.Of)
		purse, _ := pack(a, ontology.Coin)
		p := price(w, b.Of) * b.Amount
		purse.Move(-p)
		mine.Move(b.Amount)
		theirs.Move(-b.Amount)
		a.Needs.Add(need.Esteem, t.Worth[need.Esteem])
		g, _ := good(b.Of)
		w.Emit(event.Traded, a.ID, 0, "%s bought %s for %.2f", a.Name, g, p)
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

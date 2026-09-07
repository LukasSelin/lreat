package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/ontology"
)

// An exchange is composed from terms over a schema, not written: which way
// it runs comes from the ontology, what is handed over from the terms.
// Nobody buys tools at the market in the trees; that they could is the
// point.
func TestAnExchangeIsComposedFromTerms(t *testing.T) {
	w, a := shore(t)
	key := "exchange/coin>tool@market"
	trades[key] = terms{Name: "buy tools", Buys: purchase{ontology.Tool, 1, 2}, Worth: need.Levels{need.Esteem: 0.01}}
	defer delete(trades, key)
	sc := &ontology.Schema{Verb: ontology.Exchange, Inputs: []*ontology.Class{ontology.Coin}, Output: ontology.Tool, Site: ontology.Market}
	d := exchanging(ontology.Instance{Key: key, Schema: sc, Ticks: 1})
	if d == nil {
		t.Fatal("terms over a schema should be an act")
	}
	if d.Name != "buy tools" {
		t.Fatalf("composed %q", d.Name)
	}
	w.Market.Stock[entity.Tools], w.Market.Price[entity.Tools] = 5, 2
	a.Wealth = 3
	if d.Available(a, w) {
		t.Fatal("two tools at two coin each should be beyond three coin")
	}
	a.Wealth = 4
	if !run(w, a, d) {
		t.Fatal("buying with the coin and the stock should be possible")
	}
	if a.Pos != w.MarketPos {
		t.Fatalf("bought at %v, not the market", a.Pos)
	}
	if a.Inventory[entity.Tools] != 2 || a.Wealth != 0 || w.Market.Stock[entity.Tools] != 3 {
		t.Fatalf("tools %v coin %v stock %v after buying, want 2, 0, 3", a.Inventory[entity.Tools], a.Wealth, w.Market.Stock[entity.Tools])
	}
	if d.Available(a, w) {
		t.Fatal("with tools in hand there should be nothing to come for")
	}
}

// Terms that run the wrong way for their schema are no act: a selling
// over a buying schema, or a purchase of something other than what the
// schema says is bought.
func TestTermsMustFitTheirSchema(t *testing.T) {
	buying := &ontology.Schema{Verb: ontology.Exchange, Inputs: []*ontology.Class{ontology.Coin}, Output: ontology.Provision, Site: ontology.Market}
	trades["exchange/wrong-way"] = terms{Name: "wrong", Sells: []sale{{ontology.Provision, 0, 1}}}
	trades["exchange/wrong-good"] = terms{Name: "wrong", Buys: purchase{ontology.Stone, 1, 1}}
	defer delete(trades, "exchange/wrong-way")
	defer delete(trades, "exchange/wrong-good")
	for _, key := range []string{"exchange/wrong-way", "exchange/wrong-good", "exchange/none"} {
		if d := exchanging(ontology.Instance{Key: key, Schema: buying}); d != nil {
			t.Fatalf("%s: %q was composed", key, d.Name)
		}
	}
}

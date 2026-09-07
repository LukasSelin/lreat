package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// A move is composed from terms over a schema, whoever the far side is.
// The same interpreter reaches the ground, the market, and a person, and
// the trees say which. None of the three acts here is entailed by the
// trees; that they run all the same is the point.

func TestAMoveFromTheGround(t *testing.T) {
	w, a := shore(t)
	for y := 3; y < 6; y++ {
		w.Grid.At(entity.Pos{X: 2, Y: y}).Terrain = world.Rock
	}
	w.Grid.At(entity.Pos{X: 2, Y: 3}).Wood = 1 // one outcrop with driftwood on it
	key := "take/timber@outcrop"
	moves[key] = move{Name: "salvage",
		Each:  map[*ontology.Class]part{ontology.Timber: {Quantity: fixed(0.25), Drain: 0.5, Least: 0.75}},
		Worth: levels(need.Levels{need.Safety: 0.01}),
	}
	defer delete(moves, key)
	sc := &ontology.Schema{Verb: ontology.Take, Object: ontology.Material, Site: ontology.Ground}
	d := moving(ontology.Instance{Key: key, Schema: sc, Object: ontology.Timber, Site: ontology.Outcrop, Ticks: 2})
	if d == nil {
		t.Fatal("terms over a taking should be an act")
	}
	if !run(w, a, d) {
		t.Fatal("the move should find the outcrop")
	}
	if w.Grid.At(a.Pos).Terrain != world.Rock {
		t.Fatalf("stood on %v, not the outcrop", a.Pos)
	}
	if a.Inventory[entity.Wood] != 0.25 || w.Grid.At(a.Pos).Wood != 0.5 {
		t.Fatalf("pack %v ground %v; want 0.25 gained and 0.5 drained", a.Inventory[entity.Wood], w.Grid.At(a.Pos).Wood)
	}
	if _, ok := d.Target(a, w); ok {
		t.Fatal("with the outcrops below the least, there should be nowhere worth going")
	}
}

func TestAMoveWithTheMarket(t *testing.T) {
	w, a := shore(t)
	key := "exchange/coin>tool@market"
	moves[key] = move{Name: "buy tools",
		Each:  map[*ontology.Class]part{ontology.Tool: {Quantity: fixed(2), Want: 1, Least: 2}},
		Worth: levels(need.Levels{need.Esteem: 0.01}),
		Event: event.Traded, Report: func(a, _ *entity.Agent, _, paid float64) string { return "bought tools" },
	}
	defer delete(moves, key)
	sc := &ontology.Schema{Verb: ontology.Exchange, Inputs: []*ontology.Class{ontology.Coin}, Output: ontology.Tool, Site: ontology.Market}
	d := moving(ontology.Instance{Key: key, Schema: sc, Site: ontology.Market, Ticks: 1})
	if d == nil {
		t.Fatal("terms over an exchange should be an act")
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

func TestAMoveBetweenPeople(t *testing.T) {
	w := world.New(11)
	thief, mark := blank(w, "thief"), blank(w, "mark")
	mark.Pos = entity.Pos{X: thief.Pos.X + 1, Y: thief.Pos.Y}
	key := "transfer/timber<holder"
	moves[key] = move{Name: "pilfer wood", Regard: -0.2, Rift: 0.1,
		Each:  map[*ontology.Class]part{ontology.Timber: {Quantity: fixed(0.5), Want: 1}},
		Worth: levels(need.Levels{need.Safety: 0.02, need.Physiological: 0.5}),
		Gives: need.Levels{need.Safety: 0.02},
		Event: event.Stolen, Report: func(a, o *entity.Agent, _, _ float64) string { return a.Name + " pilfered wood from " + o.Name },
	}
	defer delete(moves, key)
	sc := &ontology.Schema{Verb: ontology.Transfer, Object: ontology.Timber, Role: &ontology.Holder, Dir: ontology.Seize}
	d := moving(ontology.Instance{Key: key, Schema: sc, Object: ontology.Timber, Ticks: 1})
	if d == nil {
		t.Fatal("terms over a transfer should be an act")
	}
	if d.With == nil {
		t.Fatal("a move between people is done with somebody")
	}
	if d.Available(thief, w) {
		t.Fatal("with nobody holding wood there is nobody to take it from")
	}
	mark.Inventory[entity.Wood] = 2
	thief.Inventory[entity.Wood] = 1
	if d.Available(thief, w) {
		t.Fatal("a thief with wood enough should not want more")
	}
	thief.Inventory[entity.Wood] = 0
	if got := d.Expect(thief, w, mark.Pos); got[need.Physiological] != 0.5 {
		t.Fatalf("sized up from a step away, the move should be worth its terms: %v", got)
	}
	safety := thief.Needs[need.Safety]
	if !run(w, thief, d) {
		t.Fatal("taking from a neighbour with wood to spare should be possible")
	}
	if thief.Inventory[entity.Wood] != 0.5 || mark.Inventory[entity.Wood] != 1.5 {
		t.Fatalf("thief %v mark %v after the taking, want 0.5 and 1.5", thief.Inventory[entity.Wood], mark.Inventory[entity.Wood])
	}
	if mark.Regard(thief.ID) >= 0 {
		t.Fatal("the victim always knows")
	}
	if !(thief.Needs[need.Safety] > safety) {
		t.Fatal("what a move gives, it gives")
	}
}

// Terms that do not fit their schema are no act: no terms, terms for a
// material the schema is not about, or a far side the interpreter cannot
// reach.
func TestTermsMustFitTheirMove(t *testing.T) {
	sc := &ontology.Schema{Verb: ontology.Take, Object: ontology.Material, Site: ontology.Ground}
	if d := moving(ontology.Instance{Key: "take/none", Schema: sc, Object: ontology.Timber, Site: ontology.Wood}); d != nil {
		t.Fatalf("no terms, yet %q was composed", d.Name)
	}
	moves["take/odd"] = move{Name: "odd", Each: map[*ontology.Class]part{ontology.Stone: {Quantity: fixed(1)}}}
	defer delete(moves, "take/odd")
	if d := moving(ontology.Instance{Key: "take/odd", Schema: sc, Object: ontology.Timber, Site: ontology.Wood}); d != nil {
		t.Fatalf("terms for stone over a taking of timber, yet %q was composed", d.Name)
	}
	if d := moving(ontology.Instance{Key: "take/odd", Schema: sc, Object: ontology.Stone, Site: ontology.Field}); d != nil {
		t.Fatalf("a field is no ground to take from, yet %q was composed", d.Name)
	}
}

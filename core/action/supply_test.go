package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/ontology"
)

// A moment is short of what an act would bring and supplied in what it
// would spend, whatever those are. Stone never had a coordinate of its
// own; quarrying reads the lack of it all the same, and a material the
// trees get tomorrow will read the same way.
func TestAMomentIsShortOfWhatTheActBrings(t *testing.T) {
	w, a := workshop(t)
	a.Inventory[entity.Tools] = 1
	shared := Shared(a, w)
	target, ok := Quarry.Target(a, w)
	if !ok {
		t.Fatal("the workshop has an outcrop")
	}
	if got := Situation(a, w, Quarry, target, shared)[habit.Lack]; got != 1 {
		t.Fatalf("with no stone, quarrying should read as a full lack: %v", got)
	}
	a.Inventory[entity.Stone] = knee(ontology.Stone)
	if got := Situation(a, w, Quarry, target, shared)[habit.Lack]; got != -1 {
		t.Fatalf("with stone enough, quarrying should read as no lack at all: %v", got)
	}
	// Selling reads how well stocked one is in whichever thing one has to
	// spare; buying, how short of food and how well off in coin.
	a.Inventory[entity.Food], a.Wealth = 0, 20
	if s := Situation(a, w, Sell, w.MarketPos, shared); s[habit.Stock] != 1 || s[habit.Lack] != -1 {
		t.Fatalf("a stone-rich, coin-rich seller should read stock 1, lack -1: %v %v", s[habit.Stock], s[habit.Lack])
	}
	if s := Situation(a, w, Buy, w.MarketPos, shared); s[habit.Lack] != 1 || s[habit.Stock] != 1 {
		t.Fatalf("a foodless buyer with a full purse should read lack 1, stock 1: %v %v", s[habit.Lack], s[habit.Stock])
	}
	// A making is supplied only as far as its scarcest input.
	a.Inventory[entity.Wood] = 0
	if s := Situation(a, w, Smelt, a.Home, shared); s[habit.Stock] != -1 {
		t.Fatalf("smelting with stone but no wood should read as unsupplied: %v", s[habit.Stock])
	}
	// An act that moves nothing in particular reads the stores in general.
	if s := Situation(a, w, Rest, a.Pos, shared); s[habit.Lack] != -plenty(a) || s[habit.Stock] != plenty(a) {
		t.Fatal("resting should read how the agent stands in general")
	}
}

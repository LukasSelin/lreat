package action

import (
	"math"

	"lreat/core/entity"
	"lreat/core/ontology"
	"lreat/core/world"
)

// Stores. A material is always somewhere: in the ground, in a pack, in a
// purse, on the market's shelves. The ontology says which kinds of thing
// hold which materials (ontology.Affords); this says where that is at run
// time, as a number that can be read and moved. Every generic verb moves
// material between stores it names by class, so a verb never says
// "inventory" or "tile" or "wealth": it says the ground here, this
// person's pack, the market.

// A store is some of a material somewhere.
type store interface {
	// Held is how much is there.
	Held() float64
	// Move adds delta, which may be negative, and never goes below nothing.
	Move(delta float64)
}

// cell is a store that is a number in the world.
type cell struct{ at *float64 }

func (c cell) Held() float64  { return *c.at }
func (c cell) Move(d float64) { *c.at = max(0, *c.at+d) }

// bottomless is a store with no end of what it holds and no room to fill:
// the stone in an outcrop, the market's coin. Taking from it takes
// nothing away and giving to it gives nothing back.
type bottomless struct{}

func (bottomless) Held() float64 { return math.Inf(1) }
func (bottomless) Move(float64)  {}

// pack is what a person holds of m: their inventory, or their purse for
// coin. ok is false for a thing nobody can carry.
func pack(a *entity.Agent, m *ontology.Class) (store, bool) {
	if m == ontology.Coin {
		return cell{&a.Wealth}, true
	}
	g, ok := world.GoodOf(m)
	if !ok {
		return nil, false
	}
	return cell{&a.Inventory[g]}, true
}

// shelf is what the market holds of m. The market's coin is bottomless:
// it pays whatever is brought and takes whatever is paid.
func shelf(w *world.World, m *ontology.Class) (store, bool) {
	if m == ontology.Coin {
		return bottomless{}, true
	}
	g, ok := world.GoodOf(m)
	if !ok {
		return nil, false
	}
	return cell{&w.Market.Stock[g]}, true
}

// soil is what a tile holds of m, given that the ground it stands for
// holds it at all: its stock, or a bottomless store where it has none.
func soil(t *world.Tile, m *ontology.Class) store {
	if s, ok := world.Stock(t, m); ok {
		return cell{s}
	}
	return bottomless{}
}

// price is what the market asks for a material.
func price(w *world.World, m *ontology.Class) float64 {
	g, _ := world.GoodOf(m)
	return w.Market.Price[g]
}

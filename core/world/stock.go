package world

import (
	"lreat/core/entity"
	"lreat/core/ontology"
)

// Where the ontology's names are to be found on the map.
//
// The ontology says what a thing is; this says which number on which tile,
// or in which slot of a pack, that comes to at run time. It is the whole of
// the binding between the two, and it lives here rather than in the packages
// that use it because more than one of them needs it: an act moves a
// material between stores, and the world's own processes wear those same
// stores down without anybody asking.

// classes is what a tile is, in the ontology's terms. The structure is
// asked first: a tile with something built on it is that building, whatever
// the ground under it used to be. Nothing is lost by asking in that order,
// because a structure only ever goes up on cleared ground - the one
// exception being a road carried over water, and a bridge is a road.
var classes = map[Structure]*ontology.Class{
	House:   ontology.Dwelling,
	Market:  ontology.Market,
	Granary: ontology.Granary,
	Tavern:  ontology.Tavern,
	Road:    ontology.Road,
}

// grounds is what bare ground is, in the ontology's terms.
var grounds = map[Terrain]*ontology.Class{
	Grass:  ontology.Open,
	Forest: ontology.Wood,
	Water:  ontology.Water,
	Field:  ontology.Field,
	Rock:   ontology.Outcrop,
}

// ClassOf is what this tile is: what stands on it if anything does, and
// otherwise the ground itself.
func ClassOf(t *Tile) *ontology.Class {
	if c, ok := classes[t.Structure]; ok {
		return c
	}
	return grounds[t.Terrain]
}

// goods is what a material is in a pack. A material that is nothing in a
// pack cannot be carried.
var goods = map[*ontology.Class]entity.Good{
	ontology.Provision: entity.Food,
	ontology.Timber:    entity.Wood,
	ontology.Stone:     entity.Stone,
	ontology.Tool:      entity.Tools,
	ontology.Meal:      entity.Meals,
}

// GoodOf is what m is in a pack or on a shelf, walking up the tree: berries
// are food. It is the coarsest thing the binding does, and deliberately: a
// pack has one number for everything edible, which is why what befalls a
// provision has to be said of provisions and cannot be said of berries.
func GoodOf(m *ontology.Class) (entity.Good, bool) {
	for c := m; c != nil; c = c.Parent {
		if g, ok := goods[c]; ok {
			return g, true
		}
	}
	return 0, false
}

// stocks is what the ground holds of a material, by tile. A material the
// ground holds but has no stock of (stone in an outcrop) has no entry.
var stocks = map[*ontology.Class]func(*Tile) *float64{
	ontology.Berries: func(t *Tile) *float64 { return &t.Wild },
	ontology.Game:    func(t *Tile) *float64 { return &t.Wild },
	ontology.Timber:  func(t *Tile) *float64 { return &t.Wood },
	ontology.Fish:    func(t *Tile) *float64 { return &t.Fish },
}

// Stock is the number on this tile that stands for how much of m is here.
// ok is false where the ground affords m without keeping a count of it.
func Stock(t *Tile, m *ontology.Class) (*float64, bool) {
	if s, ok := stocks[m]; ok {
		return s(t), true
	}
	return nil, false
}

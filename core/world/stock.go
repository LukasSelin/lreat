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

// ClassOf is what this tile is: what stands on it if anything does, and
// otherwise the ground itself.
func ClassOf(t *Tile) *ontology.Class {
	if c, ok := classes[t.Structure]; ok {
		return c
	}
	return t.Terrain.Class()
}

// GroundOf is what the bare ground of this tile is, whatever has been put on
// top of it. It is not ClassOf and the difference is the bridge: a road over
// water is a road to walk on and still a river to fish in, so what a tile
// affords is the ground's to say and never the structure's.
func GroundOf(t *Tile) *ontology.Class { return t.Terrain.Class() }

// Is reports whether the ground here is c, or any kind of c. It is the
// identity question asked of the ontology instead of of the terrain, so that
// a wood goes on being a wood however many kinds of wood the trees come to
// hold, and the caller does not have to be found again when they do.
func (t *Tile) Is(c *ontology.Class) bool {
	g := GroundOf(t)
	return g != nil && g.IsA(c)
}

// Affords reports whether the ground here is the kind that has m to give at
// all, whatever is standing on it today: an outcrop affords stone and a
// meadow does not. It is the ontology's question asked of a tile, and it
// is all a search needs that is looking for a kind of ground rather than
// for what is left on it.
func (t *Tile) Affords(m *ontology.Class) bool {
	return ontology.Offers(GroundOf(t), m)
}

// Offers is how much of m tile i has to give: nothing where the ground is
// not the kind that affords m at all, the standing stock where there is a
// count of it, and one where the ground affords m without keeping a count,
// stone in an outcrop being bottomless.
//
// This is the question most callers of the terrain actually meant. What a
// timber search looks for is not a forest, it is somewhere with wood standing
// on it; asking the first while meaning the second is what makes every new
// kind of ground a hunt through the callers, since the compiler has nothing
// to say about a comparison that stayed valid and stopped being right.
func (g *Grid) Offers(i int, m *ontology.Class) float64 {
	if !g.Tiles[i].Affords(m) {
		return 0
	}
	if s, ok := g.Stock(i, m); ok {
		return *s
	}
	return 1
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

// hefts is what one unit of each good is to carry, in armfuls, read off the
// ontology's Heavy trait once at start-up rather than looked up per step.
// Anything the trees do not name is an ordinary armful.
var hefts = func() [entity.GoodCount]float64 {
	var h [entity.GoodCount]float64
	for i := range h {
		h[i] = 1
	}
	for c, g := range goods {
		h[g] = ontology.Heft(c)
	}
	return h
}()

// Hauled is what an agent is carrying, in armfuls, counting a heavy material
// for what it actually is to carry. It is what marks the ground - see
// Grid.Tread - and it is the only place the weight of a pack is read as a
// quantity rather than as the yes-or-no of entity.Agent.Load, which asks
// whether a walker may take to the water and nothing else.
//
// A road is laid where the settlement's hauling runs, and hauling stone is
// the heaviest of it. Stone is the only material the trees call Heavy, and it
// is quarried rarely, so this moves the map less than it moves the meaning:
// it is here so that Heavy is a fact with a consequence rather than a label,
// and so that calling anything else heavy takes effect without another edit.
func Hauled(a *entity.Agent) float64 {
	total := 0.0
	for g, q := range a.Inventory {
		if q > 0 {
			total += q * hefts[g]
		}
	}
	return total
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

// stocks is what the ground holds of a material: the layer of the map that
// keeps the count, indexed by tile. A material the ground holds but has no
// stock of (stone in an outcrop) has no entry. Berries and game are one
// count between them, which is why two of these are the same layer.
var stocks = map[*ontology.Class]func(*Grid) []float64{
	ontology.Berries: func(g *Grid) []float64 { return g.Wild },
	ontology.Game:    func(g *Grid) []float64 { return g.Wild },
	ontology.Browse:  func(g *Grid) []float64 { return g.Wild },
	ontology.Sward:   func(g *Grid) []float64 { return g.Sward },
	ontology.Mast:    func(g *Grid) []float64 { return g.Wood },
	ontology.Timber:  func(g *Grid) []float64 { return g.Wood },
	ontology.Fish:    func(g *Grid) []float64 { return g.Fish },
}

// StockOf is Stock resolved for a material instead of for a tile: the layer
// of a map that holds m, or nil where the ground has no stock of it and is
// bottomless. It is for the searches that ask the same question of every tile
// on the map, which would otherwise look the material up on each of them.
func StockOf(m *ontology.Class) func(*Grid) []float64 { return stocks[m] }

// Stock is the number that stands for how much of m is on tile i. ok is
// false where the ground affords m without keeping a count of it. The
// number is where it is kept, so what is moved through it is moved on the
// map: a layer is made once with the map and never grows, so the pointer
// stands as long as the map does.
func (g *Grid) Stock(i int, m *ontology.Class) (*float64, bool) {
	if s, ok := stocks[m]; ok {
		return &s(g)[i], true
	}
	return nil, false
}

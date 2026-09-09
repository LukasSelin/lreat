package world

import "lreat/core/ontology"

// What a kind of ground is, in one place.
//
// Before this there was no such place. What a terrain was called existed
// nowhere at all - a Terrain printed as a number - and the rest of it was
// spread over a map in stock.go and a switch in erode.go, with two more
// switches in other packages keyed off the same five values. Adding a kind of
// ground meant finding four of those and being right about all four, and
// nothing would have said so if one was missed: a switch with no case for the
// new terrain falls through to whatever the default happens to be, and a map
// with no row returns the zero value, which for a class is nil.
//
// So the facts that are the ground's own live here, one row each, and the
// things that read them read the row. What is deliberately not here is
// anything belonging to somebody else: how fast a terrain puts back what is
// taken off it is tuning and belongs to the system that runs it, and how a
// terrain is drawn belongs to the renderer. Those keep their own tables, over
// the same terrains, and a test in each says the table covers them all.
type terrain struct {
	// name is what this ground is called, and is the ontology's word for it
	// rather than a second vocabulary: what world calls Forest, the ontology
	// calls a wood, and there is no gain in the map having its own name for
	// the same thing.
	name string
	// class is what a tile of this ground is in the ontology's terms.
	class *ontology.Class
	// hold is how well the ground holds its soil against the weather, from
	// nothing to all of it. Bare rock keeps almost none, a wood most of what
	// falls on it, and a channel and a worked field are counted whole - the
	// one because cutting down is how a valley deepens, the other because a
	// field is soil by definition.
	hold float64
}

var terrains = [TerrainCount]terrain{
	Grass:  {"open", ontology.Open, 0.6},
	Forest: {"wood", ontology.Wood, 0.25},
	Water:  {"water", ontology.Water, 1},
	Field:  {"field", ontology.Field, 1},
	Rock:   {"outcrop", ontology.Outcrop, 0.15},
}

// String is what this ground is called.
func (t Terrain) String() string {
	if int(t) >= len(terrains) {
		return "unknown"
	}
	return terrains[t].name
}

// Class is what a tile of this ground is, in the ontology's terms. It is the
// bare ground and not what has been built on it; see GroundOf and ClassOf.
func (t Terrain) Class() *ontology.Class {
	if int(t) >= len(terrains) {
		return nil
	}
	return terrains[t].class
}

// Hold is how much of its soil this ground keeps against the weather.
func (t Terrain) Hold() float64 {
	if int(t) >= len(terrains) {
		return 0
	}
	return terrains[t].hold
}

// Terrains is every kind of ground, for the tables that have to cover them
// all and the tests that check they do.
func Terrains() []Terrain {
	out := make([]Terrain, 0, TerrainCount)
	for t := Terrain(0); t < TerrainCount; t++ {
		out = append(out, t)
	}
	return out
}

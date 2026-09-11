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
	// wet says this is water and not ground at all. It is the one thing about
	// a terrain that half the map-maker asks and none of it used to be able
	// to: trees do not grow on it, silt does not settle on it, it stands
	// above nothing so it has no drain, and nobody stands on it either.
	//
	// It is not the same question as whether a river runs here. That is Flow,
	// and it is a number on the tile, because a lake is water that does not
	// flow and a river in spate is the same channel carrying more. The two
	// coincide while there is one kind of water and stop the moment there are
	// two, which is why they are apart before rather than after.
	//
	// It is here and not a trait of the ontology's on purpose. The ontology's
	// traits are what verbs test, and no act asks whether the ground is wet -
	// what an act wants of water is the fish, and Affords says that already.
	// This is the map's own fact about its own ground.
	wet bool
	// hold is how well the ground holds its soil against the weather, from
	// nothing to all of it. Bare rock keeps almost none, a wood most of what
	// falls on it, and a channel and a worked field are counted whole - the
	// one because cutting down is how a valley deepens, the other because a
	// field is soil by definition.
	hold float64
}

var terrains = [TerrainCount]terrain{
	Grass:  {name: "open", class: ontology.Open, hold: 0.6},
	Forest: {name: "wood", class: ontology.Wood, hold: 0.25},
	Water:  {name: "water", class: ontology.Water, wet: true, hold: 1},
	Field:  {name: "field", class: ontology.Field, hold: 1},
	Rock:   {name: "outcrop", class: ontology.Outcrop, hold: 0.15},
	// Ice is wet: it is the sea, and the map-maker's questions about water
	// all have the sea's answer here. Nothing grows on it, nothing settles
	// on it, and it stands above nothing, so it has no drain. What it does
	// not share with open water is that somebody can walk on it; that is
	// Tile.Deep, which is the walker's question and not the map-maker's.
	Ice: {name: "ice", class: ontology.Ice, wet: true, hold: 1},
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

// Wet reports whether this is water rather than ground.
func (t Terrain) Wet() bool {
	if int(t) >= len(terrains) {
		return false
	}
	return terrains[t].wet
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

// KindSet is a set of kinds of ground, one bit each. It is what a search
// over the ground says it is looking for, so that the search can be
// answered from the chunk counts before it is walked.
type KindSet uint32

// Kinds is the set holding just these.
func Kinds(ts ...Terrain) KindSet {
	var s KindSet
	for _, t := range ts {
		s |= 1 << t
	}
	return s
}

// KindsOf is every kind of ground that is c, or a kind of c, in the
// ontology's terms. It is derived from the same table that says what each
// ground is rather than written out again: a kind of ground added to the
// trees joins the sets it belongs to without anybody being found and told.
func KindsOf(c *ontology.Class) KindSet {
	var s KindSet
	for t := Terrain(0); t < TerrainCount; t++ {
		if cl := t.Class(); cl != nil && cl.IsA(c) {
			s |= 1 << t
		}
	}
	return s
}

// Has reports whether t is in the set.
func (s KindSet) Has(t Terrain) bool { return s&(1<<t) != 0 }

// KindsOffering is every kind of ground that affords m, in the ontology's
// terms: the ground a search for m has any business looking at. It is the
// set to hand AnyWithin when what is wanted is a material rather than a
// kind of ground - stone rather than an outcrop, fish rather than water -
// so that what affords what stays the trees' to say.
func KindsOffering(m *ontology.Class) KindSet {
	var s KindSet
	for t := Terrain(0); t < TerrainCount; t++ {
		if ontology.Offers(t.Class(), m) {
			s |= 1 << t
		}
	}
	return s
}

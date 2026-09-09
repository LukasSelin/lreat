package world

import (
	"testing"

	"lreat/core/ontology"
)

// A kind of ground with no row is the failure this table exists to stop. A
// map with no entry returns nil and a switch with no case falls through to
// whatever the default happens to be, so a terrain added and half wired up
// used to run, and to run wrong.
func TestEveryTerrainHasARow(t *testing.T) {
	seen := map[string]Terrain{}
	for _, kind := range Terrains() {
		if kind.String() == "" || kind.String() == "unknown" {
			t.Errorf("terrain %d has no name", kind)
		}
		if kind.Class() == nil {
			t.Errorf("%s is no class of the ontology's", kind)
		}
		if prev, ok := seen[kind.String()]; ok {
			t.Errorf("%s and %s are both called %q", prev, kind, kind.String())
		}
		seen[kind.String()] = kind
	}
}

// The map's kinds of ground and the ontology's are the same set said twice,
// and the whole of the binding is that they line up. A ground class the map
// cannot make is one no act will ever be instantiated at.
func TestTheGroundsAndTheOntologyAgree(t *testing.T) {
	fromMap := map[*ontology.Class]bool{}
	for _, kind := range Terrains() {
		fromMap[kind.Class()] = true
	}
	for _, c := range ontology.Ground.Leaves() {
		if !fromMap[c] {
			t.Errorf("the ontology has %s and no terrain makes one", c.Path())
		}
		delete(fromMap, c)
	}
	for c := range fromMap {
		t.Errorf("a terrain is %s, which is not a ground in the ontology", c.Path())
	}
}

// Hold is a share of the soil kept, so it belongs in [0,1]. It is read
// straight into an erosion rate, where a number outside that range would not
// fail, it would quietly wash the map away or freeze it.
func TestHoldIsSane(t *testing.T) {
	for _, kind := range Terrains() {
		if h := kind.Hold(); h < 0 || h > 1 {
			t.Errorf("%s holds %v of its soil", kind, h)
		}
	}
}

// Wet ground is water and not ground at all, and three things follow that the
// map would otherwise have to be told separately. Nothing grows in the
// channel, nothing is built on it, and it costs more to cross than any ground
// does - the last being what makes a bridge worth building rather than a
// detour worth walking.
func TestWetGroundIsNotGround(t *testing.T) {
	var wet int
	for _, kind := range Terrains() {
		if !kind.Wet() {
			continue
		}
		wet++
		if kind.Class().Has(ontology.Living) {
			t.Errorf("%s is wet and the ontology calls it living: nothing comes on in the channel", kind)
		}
		if (&Tile{Terrain: kind}).Buildable() {
			t.Errorf("%s is wet and buildable", kind)
		}
		for _, dry := range Terrains() {
			if !dry.Wet() && moveCost[kind] <= moveCost[dry] {
				t.Errorf("%s costs %v to cross and %s costs %v: water has to be dearer than ground",
					kind, moveCost[kind], dry, moveCost[dry])
			}
		}
	}
	if wet == 0 {
		t.Error("no terrain is wet, so the map has no water in it")
	}
}

// Whether a tile is water and whether a river runs through it are two
// questions, and they are apart on purpose: a lake is water that does not
// flow, and a river in spate is the same channel carrying more. Nothing may
// read one for the other.
func TestWetIsNotFlow(t *testing.T) {
	for _, kind := range Terrains() {
		still := &Tile{Terrain: kind, Flow: 0}
		spate := &Tile{Terrain: kind, Flow: 1000}
		if still.Wet() != kind.Wet() || spate.Wet() != kind.Wet() {
			t.Errorf("%s changes whether it is wet when the flow does", kind)
		}
	}
}

// Every terrain costs something to cross. A row left out reads as free
// movement, which is worse than slow: it is a tile every route prefers.
func TestEveryTerrainCostsSomethingToCross(t *testing.T) {
	for _, kind := range Terrains() {
		if moveCost[kind] <= 0 {
			t.Errorf("%s costs %v to cross", kind, moveCost[kind])
		}
	}
}

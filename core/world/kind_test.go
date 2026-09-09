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

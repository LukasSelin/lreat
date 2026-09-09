package system

import (
	"testing"

	"lreat/core/ontology"
	"lreat/core/world"
)

// Not every ground puts something back - an outcrop is stone and does not
// grow - so a nil row is meant. What is not meant is a ground that keeps a
// stock the ontology says it affords and never recovers it, which is a
// settlement that can strip something and never see it again.
func TestGroundThatIsDrawnOnRecovers(t *testing.T) {
	// Wood and wild food come back through ripen rather than through this
	// table, which reads the age rather than the terrain.
	byAge := map[*ontology.Class]bool{ontology.Wood: true}

	for _, kind := range world.Terrains() {
		c := kind.Class()
		if byAge[c] || recovery[kind] != nil {
			continue
		}
		for _, m := range ontology.Affords[c] {
			if _, ok := world.Stock(&world.Tile{Terrain: kind}, m); ok {
				t.Errorf("%s affords %s and keeps a count of it, and nothing puts it back", kind, m.Name)
			}
		}
	}
}

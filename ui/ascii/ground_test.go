package ascii

import (
	"testing"

	"lreat/core/world"
)

// Every kind of ground has to be drawable. A missing row is a nil call and a
// panic in the middle of a run, which is a poor way to find out that a
// terrain was added and the renderer was not told.
func TestEveryTerrainIsDrawn(t *testing.T) {
	seen := map[rune]world.Terrain{}
	for _, kind := range world.Terrains() {
		draw := ground[kind]
		if draw == nil {
			t.Errorf("%s is not drawn", kind)
			continue
		}
		c := draw(&world.Tile{Terrain: kind})
		if c.Ch == 0 {
			t.Errorf("%s draws as nothing at all", kind)
		}
		if prev, ok := seen[c.Ch]; ok {
			t.Errorf("%s and %s are both drawn %q", prev, kind, c.Ch)
		}
		seen[c.Ch] = kind
	}
}

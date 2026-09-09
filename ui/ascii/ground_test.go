package ascii

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/world"
)

// Every kind of ground has to be drawable. A missing row is a nil call and a
// panic in the middle of a run, which is a poor way to find out that a
// terrain was added and the renderer was not told. Two grounds may share a
// glyph - a boulder field and a crag are both '^', told apart by colour - so
// what is checked is that each draws as something.
func TestEveryTerrainIsDrawn(t *testing.T) {
	g := world.NewGrid(3, 3)
	p := entity.Pos{X: 1, Y: 1}
	for _, kind := range world.Terrains() {
		draw := ground[kind]
		if draw == nil {
			t.Errorf("%s is not drawn", kind)
			continue
		}
		tile := g.At(p)
		tile.Terrain = kind
		if c := draw(scene{g: g, p: p, t: tile, b: 0}); c.Ch == 0 {
			t.Errorf("%s draws as nothing at all", kind)
		}
	}
}

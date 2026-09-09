package system

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/world"
)

// Every count a chunk keeps is what a walk over its ground would find,
// after a settlement has built, farmed, paved, planted, moved house and let
// things fall down on it for eight years.
func TestChunkCountsAgreeWithTheGround(t *testing.T) {
	w := world.New(2)
	for i := 0; i < 30; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	Run(w, 3000)
	g := w.Grid
	for ci := range g.Chunks {
		c := &g.Chunks[ci]
		var want world.Chunk
		for y := c.Y0; y < c.Y0+c.H; y++ {
			for x := c.X0; x < c.X0+c.W; x++ {
				tile := g.At(entity.Pos{X: x, Y: y})
				switch tile.Structure {
				case world.House:
					want.Houses++
				case world.Road:
					want.Roads++
				case world.Granary:
					want.Granaries++
				case world.Market:
					want.Markets++
				case world.Tavern:
					want.Taverns++
				}
				if tile.Structure != world.None {
					want.Built++
				}
				switch tile.Terrain {
				case world.Field:
					want.Fields++
				case world.Forest:
					want.Forest++
				}
				if tile.Owner != 0 {
					want.Owned++
				}
			}
		}
		got := *c
		got.X0, got.Y0, got.W, got.H, got.Trodden = 0, 0, 0, 0, false
		if got != want {
			t.Fatalf("chunk %d keeps %+v; the ground says %+v", ci, got, want)
		}
	}
	if g.Houses() == 0 || g.Roads() == 0 || g.Fields() == 0 {
		t.Fatalf("houses %d roads %d fields %d: the settlement did too little to test the counts", g.Houses(), g.Roads(), g.Fields())
	}
}

func TestTheDefaultMapIsTwoChunks(t *testing.T) {
	w := world.New(1)
	if w.Grid.CW != 2 || w.Grid.CH != 1 || len(w.Grid.Chunks) != 2 {
		t.Fatalf("the default map is %d by %d chunks", w.Grid.CW, w.Grid.CH)
	}
	if c := w.Grid.Chunks[1]; c.X0 != 64 || c.W != 16 || c.H != 36 {
		t.Fatalf("the east chunk is %+v", c)
	}
	if w.Grid.ChunkAt(entity.Pos{X: 79, Y: 35}) != &w.Grid.Chunks[1] || w.Grid.ChunkAt(entity.Pos{X: 63, Y: 0}) != &w.Grid.Chunks[0] {
		t.Fatal("a position is put in the wrong chunk")
	}
}

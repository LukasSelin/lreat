package world

import (
	"testing"

	"lreat/core/entity"
)

func TestNearestWalksOutward(t *testing.T) {
	g := NewGrid(20, 20)
	g.At(entity.Pos{X: 10, Y: 13}).Terrain = Forest // distance 3
	g.At(entity.Pos{X: 15, Y: 10}).Terrain = Forest // distance 5
	p, ok := g.Nearest(entity.Pos{X: 10, Y: 10}, 10, func(_ entity.Pos, tile *Tile) bool {
		return tile.Terrain == Forest
	})
	if !ok || p != (entity.Pos{X: 10, Y: 13}) {
		t.Fatalf("nearest = %v ok=%v, want (10,13)", p, ok)
	}
	if _, ok := g.Nearest(entity.Pos{X: 0, Y: 0}, 2, func(_ entity.Pos, tile *Tile) bool { return tile.Terrain == Forest }); ok {
		t.Fatal("found a forest beyond the search radius")
	}
}

func TestTerrainHasRiverForestAndMarketOnGrass(t *testing.T) {
	w := New(5)
	g := w.Grid
	if g.Count(func(tile *Tile) bool { return tile.Terrain == Water }) < g.H {
		t.Fatal("river should span the map")
	}
	if g.Count(func(tile *Tile) bool { return tile.Terrain == Forest }) == 0 {
		t.Fatal("no forest generated")
	}
	if m := g.At(w.MarketPos); m.Structure != Market || m.Terrain != Grass {
		t.Fatalf("market tile = %+v, want market on grass", *m)
	}
	// Fertility should fall away from the river.
	var nearSum, farSum float64
	var near, far int
	for i := range g.Tiles {
		tile := &g.Tiles[i]
		if tile.Terrain == Water {
			continue
		}
		p := entity.Pos{X: i % g.W, Y: i / g.W}
		if g.HasNeighbor(p, func(n *Tile) bool { return n.Terrain == Water }) {
			nearSum += tile.Fertility
			near++
		} else {
			farSum += tile.Fertility
			far++
		}
	}
	if near == 0 || far == 0 || nearSum/float64(near) <= farSum/float64(far) {
		t.Fatalf("riverbank fertility %.2f should exceed inland %.2f", nearSum/float64(near), farSum/float64(far))
	}
}

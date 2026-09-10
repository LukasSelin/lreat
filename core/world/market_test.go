package world

import (
	"testing"

	"lreat/core/entity"
)

// A settlement may hold more than one square, and an errand is walked to
// the one at hand rather than to the one the town was founded on.
func TestTradeGoesToTheNearestSquare(t *testing.T) {
	w := NewSized(3, 24, 12)
	w.Grid.Layers = NewLayers(len(w.Grid.Tiles))
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i] = Tile{Terrain: Grass}
	}
	old, far := entity.Pos{X: 2, Y: 6}, entity.Pos{X: 20, Y: 6}
	for _, p := range []entity.Pos{old, far} {
		w.Grid.Build(p, Market)
	}
	w.markets = nil
	w.MarketPos = old
	w.FoundMarket(old)
	w.FoundMarket(far)

	if got, _ := w.NearestMarket(entity.Pos{X: 3, Y: 6}); got != old {
		t.Errorf("somebody beside the old square trades at %v", got)
	}
	if got, _ := w.NearestMarket(entity.Pos{X: 19, Y: 6}); got != far {
		t.Errorf("somebody who has walked down the valley trades at %v, not the square beside them", got)
	}
	if n := len(w.Markets()); n != 2 {
		t.Errorf("the settlement holds %d squares, want both", n)
	}
}

// Founding the same square twice does not make two of it.
func TestASquareIsFoundedOnce(t *testing.T) {
	w := NewSized(3, 12, 6)
	w.markets = nil
	p := entity.Pos{X: 4, Y: 2}
	w.FoundMarket(p)
	w.FoundMarket(p)
	if n := len(w.Markets()); n != 1 {
		t.Fatalf("founding one square twice gave %d", n)
	}
}

// The list of squares is a cache and the ground is the truth. A square that
// has been built over or moved is passed over, and a settlement whose cache
// has gone stale still trades on the square it was founded on - because
// every errand in the catalog is walked to whatever this returns.
func TestAStaleSquareIsNotTradedOn(t *testing.T) {
	w := NewSized(3, 24, 12)
	w.Grid.Layers = NewLayers(len(w.Grid.Tiles))
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i] = Tile{Terrain: Grass}
	}
	gone, real := entity.Pos{X: 2, Y: 6}, entity.Pos{X: 20, Y: 6}
	w.Grid.Build(real, Market) // gone has no market on it
	w.markets = nil
	w.MarketPos = real
	w.FoundMarket(gone)

	got, ok := w.NearestMarket(entity.Pos{X: 3, Y: 6})
	if !ok {
		t.Fatal("a settlement with a square found none")
	}
	if got != real {
		t.Fatalf("traded at %v, which has no market on it", got)
	}
	if w.Grid.At(got).Structure != Market {
		t.Fatal("the square traded at is not a market")
	}
}

// A world with no square at all says so rather than pointing at bare ground.
func TestNoSquareIsNoSquare(t *testing.T) {
	w := NewSized(3, 12, 6)
	w.Grid.Layers = NewLayers(len(w.Grid.Tiles))
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i] = Tile{Terrain: Grass}
	}
	w.markets = nil
	if p, ok := w.NearestMarket(entity.Pos{X: 1, Y: 1}); ok {
		t.Fatalf("a settlement with no square trades at %v", p)
	}
}

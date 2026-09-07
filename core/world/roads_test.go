package world

import (
	"testing"

	"lreat/core/entity"
)

func TestPaveRespectsWhatIsAlreadyThere(t *testing.T) {
	w := NewSized(1, 12, 12)
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i] = Tile{Terrain: Grass}
	}
	g := w.Grid

	wood := entity.Pos{X: 2, Y: 2}
	g.At(wood).Terrain, g.At(wood).Wood = Forest, 0.8
	if !g.Pave(wood) {
		t.Fatal("a road would not go through the woods")
	}
	if tile := g.At(wood); tile.Terrain != Grass || tile.Wood != 0 {
		t.Fatalf("the trees survived the road: %+v", tile)
	}

	river := entity.Pos{X: 4, Y: 4}
	g.At(river).Terrain = Water
	if g.Pave(river) {
		t.Fatal("a road was laid across the water; there are no bridges yet")
	}

	claimed := entity.Pos{X: 6, Y: 6}
	g.At(claimed).Owner = 7
	if g.Pave(claimed) {
		t.Fatal("a road was laid over land somebody had claimed")
	}

	home := entity.Pos{X: 8, Y: 8}
	g.At(home).Structure = House
	if g.Pave(home) {
		t.Fatal("a road was laid through a house")
	}
	if g.At(wood).Buildable() {
		t.Fatal("a paved tile is still open to building on; a way laid is a way kept")
	}
}

// The streets should join every house to the market, and running the paving
// again should extend the network rather than start over.
func TestPaveStreetsConnectsTheSettlement(t *testing.T) {
	w := NewSized(2, 24, 16)
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i] = Tile{Terrain: Grass}
	}
	w.MarketPos = entity.Pos{X: 12, Y: 8}
	w.Grid.At(w.MarketPos).Structure = Market

	homes := []entity.Pos{{X: 3, Y: 3}, {X: 20, Y: 13}, {X: 4, Y: 12}}
	for _, h := range homes {
		w.Grid.At(h).Structure = House
	}
	if laid := w.PaveStreets(); laid == 0 {
		t.Fatal("paving the settlement laid no road at all")
	}
	for _, h := range homes {
		if !w.Grid.HasNeighbor(h, func(t *Tile) bool { return t.Structure == Road }) {
			t.Fatalf("the house at %v has no street outside it", h)
		}
		routes := w.Grid.Routes(h)
		onRoad := 0
		for _, p := range routes.Path(w.MarketPos) {
			if w.Grid.At(p).Structure == Road {
				onRoad++
			}
		}
		if onRoad == 0 {
			t.Fatalf("the walk from %v to the market touches no road", h)
		}
	}

	before := w.Grid.Count(func(t *Tile) bool { return t.Structure == Road })
	if again := w.PaveStreets(); again != 0 {
		t.Fatalf("paving an already-paved settlement laid %d more tiles", again)
	}
	newHome := entity.Pos{X: 21, Y: 2}
	w.Grid.At(newHome).Structure = House
	if w.PaveStreets() == 0 {
		t.Fatal("a house built after the streets were laid got no street")
	}
	if after := w.Grid.Count(func(t *Tile) bool { return t.Structure == Road }); after <= before {
		t.Fatalf("road tiles went from %d to %d; the network should have grown", before, after)
	}
}

// Ground remembers being walked on, and forgets when nobody comes.
func TestGroundRemembersBeingWalkedOn(t *testing.T) {
	g := NewGrid(10, 10)
	busy, quiet := entity.Pos{X: 3, Y: 3}, entity.Pos{X: 7, Y: 7}
	for i := 0; i < 20; i++ {
		g.Tread(busy)
	}
	g.Tread(quiet)
	if g.At(busy).Traffic <= g.At(quiet).Traffic {
		t.Fatal("the well-walked tile is no more worn than the once-walked one")
	}

	before := g.At(busy).Traffic
	for i := 0; i < 200; i++ {
		g.Weather()
	}
	if after := g.At(busy).Traffic; after >= before {
		t.Fatalf("wear went from %.2f to %.2f with nobody walking; it should fade", before, after)
	}
}

// Somebody looking for where to lay a road finds the most-walked open ground,
// and does not offer to pave what is already built on or claimed.
func TestBusiestFindsTheWornWay(t *testing.T) {
	g := NewGrid(20, 20)
	from := entity.Pos{X: 10, Y: 10}
	way := entity.Pos{X: 12, Y: 10}
	for i := 0; i < 30; i++ {
		g.Tread(way)
	}
	// A busier tile, but somebody's field: property is not a right of way.
	claimed := entity.Pos{X: 8, Y: 10}
	g.At(claimed).Owner = 3
	for i := 0; i < 60; i++ {
		g.Tread(claimed)
	}
	p, worn, ok := g.Busiest(from, 6)
	if !ok || p != way {
		t.Fatalf("Busiest picked %v (ok=%v), want the worn open ground at %v", p, ok, way)
	}
	if worn <= 0 {
		t.Fatalf("Busiest reported wear %v", worn)
	}
	if _, _, ok := g.Busiest(entity.Pos{X: 2, Y: 2}, 1); ok {
		t.Fatal("Busiest found somewhere worth paving in untrodden wilderness")
	}
	// Paving settles the question, so the ground stops asking.
	g.Pave(way)
	if g.At(way).Traffic != 0 {
		t.Fatal("a paved tile still reads as ground crying out for a road")
	}
}

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

	// Over water a road is a bridge. The river stays a river underneath it.
	river := entity.Pos{X: 4, Y: 4}
	g.At(river).Terrain = Water
	if !g.Pave(river) {
		t.Fatal("a road would not cross the water")
	}
	if tile := g.At(river); !tile.Bridged() || tile.Terrain != Water {
		t.Fatalf("the crossing is not a bridge over water: %+v", tile)
	}
	if c := g.MoveCost(river); c >= moveCost[Water] {
		t.Fatalf("crossing the bridge costs %v, no better than wading at %v", c, moveCost[Water])
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

// Somebody looking for where to lay a road reads the open ground and never
// offers what is built on or claimed - but a busy doorway is exactly the
// reason to lay a street past it, so ground that can never be paved lends its
// wear to the gaps beside it.
func TestBusiestFindsTheWornWay(t *testing.T) {
	g := NewGrid(30, 12)

	// A quiet corner with one lightly walked open tile.
	lone := entity.Pos{X: 24, Y: 6}
	for i := 0; i < 30; i++ {
		g.Tread(lone)
	}
	if p, _, ok := g.Busiest(entity.Pos{X: 24, Y: 6}, 3, nil); !ok || p != lone {
		t.Fatalf("in open country Busiest picked %v (ok=%v), want the worn tile at %v", p, ok, lone)
	}

	// A thronged doorway. The house itself is no thoroughfare, so the case it
	// makes is for the ground beside it.
	door := entity.Pos{X: 5, Y: 6}
	g.At(door).Structure = House
	for i := 0; i < 100; i++ {
		g.Tread(door)
	}
	p, worn, ok := g.Busiest(entity.Pos{X: 5, Y: 6}, 3, nil)
	if !ok {
		t.Fatal("a thronged doorway made no case for a street beside it")
	}
	if p == door {
		t.Fatal("Busiest offered to pave the house itself")
	}
	if entity.Dist(p, door) != 1 {
		t.Fatalf("Busiest picked %v, want ground next to the doorway at %v", p, door)
	}
	if worn <= g.At(p).Traffic {
		t.Fatalf("the case for %v is %.1f, no more than the tile's own wear; the doorway lent nothing", p, worn)
	}

	// Open ground speaks only for itself: were it to lend too, paving would
	// come out in patches rather than in lines.
	quiet := entity.Pos{X: 24, Y: 8}
	if g.Draw(quiet) != g.At(quiet).Traffic {
		t.Fatal("open ground lent its wear to a neighbour")
	}

	// A street already carries what it carries; it does not argue for another
	// street beside it.
	g.At(door).Structure = Road
	if d := g.Draw(entity.Pos{X: 5, Y: 7}); d != 0 {
		t.Fatalf("ground beside a road drew %.1f; traffic on a street is already served", d)
	}

	if _, _, ok := g.Busiest(entity.Pos{X: 15, Y: 1}, 1, nil); ok {
		t.Fatal("Busiest found somewhere worth paving in untrodden wilderness")
	}

	// Paving settles the question, so the ground stops asking.
	g.Pave(lone)
	if g.At(lone).Traffic != 0 {
		t.Fatal("a paved tile still reads as ground crying out for a road")
	}
}

// The reading of the ground the whole settlement shares must answer exactly
// what walking the neighbourhood by hand answered, ties and all: it is the
// same question asked once instead of once per person, and if it resolved a
// tie differently the settlement would pave somewhere else.
func TestWaysSaysWhatWalkingTheGroundSaid(t *testing.T) {
	const radius = 12
	w := NewSized(9, 40, 24)
	g := w.Grid
	rng := w.RNG
	for i := range g.Tiles {
		tile := &g.Tiles[i]
		*tile = Tile{Terrain: Grass}
		switch n := rng.IntN(10); {
		case n < 2:
			tile.Terrain = Water
		case n < 3:
			tile.Terrain = Forest
		case n < 4:
			tile.Structure = House
		case n < 5:
			tile.Structure = Road
		}
		// Wear in whole crossings, so that ties are common rather than a
		// thing floating point makes vanishingly rare.
		tile.Traffic = float64(rng.IntN(4))
	}
	dry := func(t *Tile) bool { return t.Terrain != Water }
	wet := func(t *Tile) bool { return t.Terrain == Water }
	ways := w.Ways()
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			from := entity.Pos{X: x, Y: y}
			gotDry, gotWet := ways.Busiest(from, radius)
			for _, c := range []struct {
				name string
				got  Pick
				ok   func(*Tile) bool
			}{{"dry", gotDry, dry}, {"water", gotWet, wet}} {
				p, worn, found := g.Busiest(from, radius, c.ok)
				if c.got.Found != found || c.got.Worn != worn || (found && c.got.Pos != p) {
					t.Fatalf("from %v the %s reading says %v/%.1f/%v, walking it says %v/%.1f/%v",
						from, c.name, c.got.Pos, c.got.Worn, c.got.Found, p, worn, found)
				}
			}
		}
	}
}

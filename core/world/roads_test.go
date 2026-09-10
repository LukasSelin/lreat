package world

import (
	"testing"

	"lreat/core/entity"
)

func TestPaveRespectsWhatIsAlreadyThere(t *testing.T) {
	w := NewSized(1, 12, 12)
	w.Grid.Layers = NewLayers(len(w.Grid.Tiles))
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i] = Tile{Terrain: Grass}
	}
	g := w.Grid

	wood := entity.Pos{X: 2, Y: 2}
	g.Turn(wood, Forest)
	g.At(wood).Wood = 0.8
	if !g.Pave(wood) {
		t.Fatal("a road would not go through the woods")
	}
	if tile := g.At(wood); tile.Terrain != Grass || tile.Wood != 0 {
		t.Fatalf("the trees survived the road: %+v", tile)
	}

	// Over water a road is a bridge. The river stays a river underneath it.
	river := entity.Pos{X: 4, Y: 4}
	g.Turn(river, Water)
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
	g.Claim(claimed, 7)
	if g.Pave(claimed) {
		t.Fatal("a road was laid over land somebody had claimed")
	}

	home := entity.Pos{X: 8, Y: 8}
	g.Build(home, House)
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
	w.Grid.Layers = NewLayers(len(w.Grid.Tiles))
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i] = Tile{Terrain: Grass}
	}
	w.MarketPos = entity.Pos{X: 12, Y: 8}
	w.Grid.Build(w.MarketPos, Market)

	homes := []entity.Pos{{X: 3, Y: 3}, {X: 20, Y: 13}, {X: 4, Y: 12}}
	for _, h := range homes {
		w.Grid.Build(h, House)
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
	w.Grid.Build(newHome, House)
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
		g.Tread(busy, 0)
	}
	g.Tread(quiet, 0)
	if g.Traffic[g.Index(busy)] <= g.Traffic[g.Index(quiet)] {
		t.Fatal("the well-walked tile is no more worn than the once-walked one")
	}

	before := g.Traffic[g.Index(busy)]
	for i := 0; i < 200; i++ {
		g.Weather()
	}
	if after := g.Traffic[g.Index(busy)]; after >= before {
		t.Fatalf("wear went from %.2f to %.2f with nobody walking; it should fade", before, after)
	}
}

// Somebody looking for where to lay a road reads the open ground and never
// offers what is built on or claimed - but a busy doorway is exactly the
// reason to lay a street past it, so ground that can never be paved lends its
// wear to the gaps beside it.
func TestBusiestFindsTheWornWay(t *testing.T) {
	g := NewGrid(30, 12)

	// A quiet corner with one well walked open tile: walked enough to be
	// worth a road, or the reading passes it over as nothing.
	lone := entity.Pos{X: 24, Y: 6}
	for i := 0; i < WorthPaving+30; i++ {
		g.Tread(lone, 0)
	}
	if p, _, ok := g.Busiest(entity.Pos{X: 24, Y: 6}, 3, nil); !ok || p != lone {
		t.Fatalf("in open country Busiest picked %v (ok=%v), want the worn tile at %v", p, ok, lone)
	}

	// A thronged doorway. The house itself is no thoroughfare, so the case it
	// makes is for the ground beside it.
	door := entity.Pos{X: 5, Y: 6}
	g.Build(door, House)
	for i := 0; i < 8*WorthPaving+100; i++ { // shared among eight gaps, and still worth a road each
		g.Tread(door, 0)
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
	if worn <= g.Traffic[g.Index(p)] {
		t.Fatalf("the case for %v is %.1f, no more than the tile's own wear; the doorway lent nothing", p, worn)
	}

	// Open ground speaks only for itself: were it to lend too, paving would
	// come out in patches rather than in lines.
	quiet := entity.Pos{X: 24, Y: 8}
	if g.Draw(quiet) != g.Traffic[g.Index(quiet)] {
		t.Fatal("open ground lent its wear to a neighbour")
	}

	// A street already carries what it carries; it does not argue for another
	// street beside it.
	g.Build(door, Road)
	if d := g.Draw(entity.Pos{X: 5, Y: 7}); d != 0 {
		t.Fatalf("ground beside a road drew %.1f; traffic on a street is already served", d)
	}

	if _, _, ok := g.Busiest(entity.Pos{X: 15, Y: 1}, 1, nil); ok {
		t.Fatal("Busiest found somewhere worth paving in untrodden wilderness")
	}

	// Paving settles the question, and the ground stops asking because a road
	// is not Pavable and lends nothing, not because the wear is thrown away.
	// The wear stays on as the road's keep: see Walked.
	worn = g.Traffic[g.Index(lone)]
	g.Pave(lone)
	if _, _, ok := g.Busiest(lone, 0, nil); ok {
		t.Fatal("a paved tile is still offered as ground crying out for a road")
	}
	if g.Traffic[g.Index(lone)] != worn {
		t.Fatalf("the road kept %.1f of the %.1f of wear that made the case for it",
			g.Traffic[g.Index(lone)], worn)
	}
}

// A doorway lends its wear once. Lent whole to every gap around it, one busy
// house made the case for a street on all eight sides, and paving one side
// took nothing off what it had to say about the other seven.
func TestADoorwayLendsItsWearOnce(t *testing.T) {
	g := NewGrid(20, 20)
	door := entity.Pos{X: 10, Y: 10}
	g.Build(door, House)
	for i := 0; i < 800; i++ {
		g.Tread(door, 0)
	}

	// Eight ways out, so each gap hears an eighth of the errands.
	gap := entity.Pos{X: 10, Y: 9}
	if d := g.Draw(gap); d < 90 || d > 110 {
		t.Fatalf("the gap beside a doorway worn %.0f drew %.1f, want about an eighth of it",
			g.Traffic[g.Index(door)], d)
	}

	// Hem the house in and the one way left carries the lot: that gap really
	// is the doorway, and there is nowhere else for the road to go.
	for _, off := range dirs {
		q := entity.Pos{X: door.X + off.X, Y: door.Y + off.Y}
		if q != gap {
			g.Build(q, House)
		}
	}
	if d := g.Draw(gap); d < g.Traffic[g.Index(door)] {
		t.Fatalf("the only gap out of a hemmed-in house drew %.1f of its %.1f of wear",
			d, g.Traffic[g.Index(door)])
	}

	// With a street outside it the house is served, and says nothing more.
	g.Build(gap, Road)
	other := entity.Pos{X: 10, Y: 11}
	g.Build(other, None)
	g.Traffic[g.Index(other)] = 0
	if d := g.Draw(other); d != 0 {
		t.Fatalf("a house with a street outside it drew %.1f for a second one", d)
	}
}

// Ground the streets already run past is passed over. A road is the cheapest
// going on the map, so the tiles beside one carry the traffic that funnels on
// and off it, and read as bare wear that is a standing case for paving the
// next tile out for as long as anybody walks.
func TestPavingDoesNotWidenAStreetItAlreadyHas(t *testing.T) {
	g := NewGrid(20, 20)

	// A length of street running east to west.
	for x := 5; x <= 9; x++ {
		g.Build(entity.Pos{X: x, Y: 10}, Road)
	}

	// Alongside it: three roads that already reach each other without it.
	if !g.Served(entity.Pos{X: 7, Y: 11}) {
		t.Fatal("the tile alongside a street is not on the street")
	}
	// Off its end: one road, and paving carries the way onward.
	if g.Served(entity.Pos{X: 10, Y: 10}) {
		t.Fatal("the ground off the end of a street counts as already served")
	}
	// Open country, with no street to be on.
	if g.Served(entity.Pos{X: 2, Y: 2}) {
		t.Fatal("untouched ground counts as already served")
	}

	// Two stubs that do not otherwise meet: the tile between them joins them,
	// which is the one thing a road can do that its neighbours cannot.
	g2 := NewGrid(20, 20)
	g2.Build(entity.Pos{X: 4, Y: 10}, Road)
	g2.Build(entity.Pos{X: 6, Y: 10}, Road)
	if g2.Served(entity.Pos{X: 5, Y: 10}) {
		t.Fatal("the gap between two separate ways counts as already served")
	}

	// So Busiest offers the end of the street and never its flank, however
	// worn the flank is.
	flank := entity.Pos{X: 7, Y: 11}
	end := entity.Pos{X: 10, Y: 10}
	for i := 0; i < 900; i++ {
		g.Tread(flank, 0)
	}
	for i := 0; i < 800; i++ { // worn past the least case anything is done about
		g.Tread(end, 0)
	}
	p, _, ok := g.Busiest(entity.Pos{X: 8, Y: 10}, 4, nil)
	if !ok {
		t.Fatal("nowhere worth paving beside a worn street")
	}
	if p == flank {
		t.Fatalf("paving aimed at %v, alongside a street that already runs there", p)
	}
	if p != end {
		t.Fatalf("paving aimed at %v, want the end of the street at %v", p, end)
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
	g.Layers = NewLayers(len(g.Tiles))
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
		// thing floating point makes vanishingly rare, and in fifties so
		// that cases fall on both sides of what a road is worth.
		for n := rng.IntN(4) * 50; n > 0; n-- {
			g.Tread(g.PosOf(i), 0)
		}
	}
	g.Recount() // the ground was remade wholesale
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

// A road costs the same timber wherever it goes and does not save the same
// amount, so ground walked alike does not make an equal case for paving. Read
// on wear alone a settlement paves the flat it was already crossing easily
// and leaves the thicket and the river, which is where a road is the point.
func TestTheCaseForARoadIsWeighedByWhatItSaves(t *testing.T) {
	g := NewGrid(20, 20)

	// Grass is the unit: open ground reads exactly as worn as it is walked.
	meadow := entity.Pos{X: 3, Y: 3}
	for i := 0; i < 100; i++ {
		g.Tread(meadow, 0)
	}
	if d := g.Draw(meadow); d != g.Traffic[g.Index(meadow)] {
		t.Fatalf("grass drew %.1f against %.1f of wear; it is meant to be the unit",
			d, g.Traffic[g.Index(meadow)])
	}

	// The same wear on dearer going makes a better case, in the order the
	// ground is dear to cross: grass, field, rock, wood, water.
	var last float64
	for _, c := range []struct {
		terrain Terrain
		name    string
	}{{Grass, "grass"}, {Field, "field"}, {Rock, "rock"}, {Forest, "wood"}, {Water, "water"}} {
		p := entity.Pos{X: 10, Y: 10}
		g.Turn(p, c.terrain)
		g.Traffic[g.Index(p)] = 100
		d := g.Draw(p)
		if d <= last {
			t.Fatalf("%s drew %.1f, no better than the easier going before it at %.1f",
				c.name, d, last)
		}
		last = d
	}

	// And a ford is worth six lengths of ordinary street to whoever crosses
	// it, which is what lets a crossing clear the same bar as a lane on a
	// sixth of the wear rather than on an allowance of its own.
	ford := entity.Pos{X: 10, Y: 10} // still water from the walk above
	lane := entity.Pos{X: 14, Y: 14}
	g.Traffic[g.Index(ford)] = 100
	g.Traffic[g.Index(lane)] = 600
	if w, l := g.Draw(ford), g.Draw(lane); w < l*0.99 || w > l*1.01 {
		t.Fatalf("a ford worn 100 drew %.1f against a lane worn 600 at %.1f; want them level", w, l)
	}
}

// Wear is a record of the settlement's work, not of its wandering. A road is
// laid because grain has to come off the field and timber out of the wood, so
// the crossing that most wants one is the laden crossing, and counting it the
// same as an idler's made the two indistinguishable.
func TestHaulingMarksTheGroundMoreThanStrolling(t *testing.T) {
	g := NewGrid(10, 10)
	lane := entity.Pos{X: 2, Y: 2}
	cartway := entity.Pos{X: 6, Y: 6}

	// Twice as many people stroll down the lane as haul along the cartway.
	for i := 0; i < 20; i++ {
		g.Tread(lane, 0)
	}
	for i := 0; i < 10; i++ {
		g.Tread(cartway, 4) // four sacks on the back
	}
	if l, c := g.Traffic[g.Index(lane)], g.Traffic[g.Index(cartway)]; c <= l {
		t.Fatalf("the cartway is worn %.0f against the strolled lane's %.0f; "+
			"carrying is meant to tell", c, l)
	}

	// An empty-handed crossing still marks the ground, because a road is a
	// footpath too and being walked at all is what keeps it - see Walked.
	quiet := entity.Pos{X: 8, Y: 8}
	g.Tread(quiet, 0)
	if g.Traffic[g.Index(quiet)] != Wear {
		t.Fatalf("an empty-handed crossing marked %.2f, want %v", g.Traffic[g.Index(quiet)], Wear)
	}
}

package system

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/world"
)

// The trigger, end to end: nobody is told to lay roads and nothing lays them
// on the settlement's behalf. Agents recognise worn ground as calling for a
// road and pave it themselves, and the network that results sits where the
// settlement's own errands run.
func TestASettlementLaysItsOwnRoads(t *testing.T) {
	w := world.New(5)
	for i := 0; i < 25; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	Run(w, 4000)

	roads := w.Grid.Count(func(t *world.Tile) bool { return t.Structure == world.Road })
	if roads == 0 {
		t.Fatal("nobody in the settlement ever decided to lay a road")
	}
	laid := 0
	for _, e := range w.Log.Since(0) {
		if e.Kind == event.Built && e.Actor != 0 {
			laid++
		}
	}
	if laid == 0 {
		t.Fatal("roads appeared but no agent is recorded as having laid one")
	}
	t.Logf("%d road tiles, %d agents alive", roads, len(w.Agents))

	// The roads should be among the settlement, not scattered over the map:
	// they follow traffic, and traffic is where people live and work.
	near := 0
	for i := range w.Grid.Tiles {
		if w.Grid.Tiles[i].Structure != world.Road {
			continue
		}
		p := entity.Pos{X: i % w.Grid.W, Y: i / w.Grid.W}
		if w.Grid.HasNeighbor(p, func(t *world.Tile) bool {
			return t.Structure != world.None || t.Terrain == world.Field
		}) {
			near++
		}
	}
	if near*2 < roads {
		t.Fatalf("only %d of %d road tiles touch anything built or farmed; roads should follow the settlement", near, roads)
	}
}

// Roads are laid where people walk. A settlement whose ground nobody has worn
// has nothing to pave, however much timber is lying about.
func TestPavingWaitsForTraffic(t *testing.T) {
	w := world.New(6)
	for i := 0; i < 12; i++ {
		a := w.Spawn("a", w.RandomPersonality())
		a.Inventory[entity.Wood] = 8
		a.Shelter = 1
	}
	// One tick of living, then wipe the record of where anyone walked.
	Run(w, 1)
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i].Traffic = 0
	}
	for _, a := range w.Agents {
		a.Plan = nil
	}
	Decide(w)
	for _, a := range w.Agents {
		if a.Plan != nil && a.Plan.Action == "lay road" {
			t.Fatal("an agent set out to pave ground nobody had walked")
		}
	}
}

// A settlement grows on both banks, because the ground near water is the
// ground worth farming. It should answer the river it straddles by bridging
// it, rather than by wading it forever.
func TestASettlementBridgesTheRiverItStraddles(t *testing.T) {
	bridged, waded, crossed := 0, 0.0, 0.0
	for _, seed := range []uint64{1, 3, 5, 7} {
		w := world.New(seed)
		for i := 0; i < 20; i++ {
			w.Spawn("a", w.RandomPersonality())
		}
		for tick := 0; tick < 4000; tick++ {
			Step(w)
			for _, a := range w.Agents {
				if tile := w.Grid.At(a.Pos); tile.Terrain == world.Water {
					crossed++
					if !tile.Bridged() {
						waded++
					}
				}
			}
		}
		n := w.Grid.Count(func(t *world.Tile) bool { return t.Bridged() })
		t.Logf("seed %d: %d bridge tiles", seed, n)
		if n > 0 {
			bridged++
		}
	}
	if bridged == 0 {
		t.Fatal("no settlement ever bridged the river it lives on both sides of")
	}
	if crossed > 0 && waded/crossed > 0.8 {
		t.Fatalf("%.0f%% of time spent in the water was wading; the bridges are not being used",
			100*waded/crossed)
	}
}

// A road is kept by feet. A way people still come down stands; a way that has
// been out of everybody's road for a generation goes back to grass, which is
// what stops a settlement's streets from being the only thing about it that
// can never change. See ontology.Transforms and world.Walked.
func TestAnUnwalkedRoadGrowsOver(t *testing.T) {
	w := world.NewSized(4, 16, 16)
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i] = world.Tile{Terrain: world.Grass}
	}
	walked := entity.Pos{X: 4, Y: 4}
	forgotten := entity.Pos{X: 11, Y: 11}
	for _, p := range []entity.Pos{walked, forgotten} {
		for i := 0; i < 200; i++ {
			w.Grid.Tread(p)
		}
		if !w.Grid.Pave(p) {
			t.Fatalf("could not lay a road at %v", p)
		}
	}
	// A road starts life with the wear that made the case for it, so neither
	// is in any danger yet however long nobody comes.
	for i := 0; i < 200; i++ {
		Upkeep(w)
	}
	for _, p := range []entity.Pos{walked, forgotten} {
		if w.Grid.At(p).Structure != world.Road {
			t.Fatalf("the road at %v went while it still had wear on it", p)
		}
	}
	// Now let the years pass, with people still using the one and not the
	// other. Weather is what fades the wear, so it runs as it does in Land.
	for i := 0; i < 4000; i++ {
		w.Grid.Weather()
		w.Grid.Tread(walked)
		Upkeep(w)
	}
	if w.Grid.At(walked).Structure != world.Road {
		t.Fatal("a road people were still walking every day grew over")
	}
	if s := w.Grid.At(forgotten).Structure; s != world.None {
		t.Fatalf("a road nobody had walked for four thousand ticks is still there as %v", s)
	}
	if !w.Grid.At(forgotten).Buildable() {
		t.Fatal("the ground a road grew off is not open ground again")
	}
}

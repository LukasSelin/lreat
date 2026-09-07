package system

import (
	"testing"

	"lreat/core/action"
	"lreat/core/entity"
	"lreat/core/world"
)

// pressured is a world with knowledge enough for every land discovery and
// none of the pressures that call for them.
func pressured(seed uint64) *world.World {
	w := fitWorld(seed)
	populate(w, 4)
	w.Knowledge = 100
	w.Unlock("agriculture")
	return w
}

func has(w *world.World, techs ...world.Tech) []world.Tech {
	var out []world.Tech
	for _, t := range techs {
		if w.Has(t) {
			out = append(out, t)
		}
	}
	return out
}

func TestTheLandsAnswersWaitForTheirPressures(t *testing.T) {
	w := pressured(41)
	Discover(w)
	if got := has(w, "fishing", "trapping", "irrigation", "forestry"); len(got) != 0 {
		t.Fatalf("discovered %v with the land untouched", got)
	}
}

func TestAThinForestByARiverTeachesFishing(t *testing.T) {
	w := pressured(42)
	for i := range w.Grid.Tiles {
		if w.Grid.Tiles[i].Terrain == world.Forest {
			w.Grid.Tiles[i].Wild = 0.1
		}
	}
	Discover(w)
	if !w.Has("fishing") {
		t.Fatal("a thin forest by a river should teach fishing")
	}
	if w.ReachFloor[action.Index(action.Fish)] < action.Opened {
		t.Fatal("fishing should open fish to everyone")
	}
	if w.Has("trapping") {
		t.Fatal("trapping should wait for tools as well")
	}
	w.Agents[0].Inventory[entity.Tools], w.Agents[1].Inventory[entity.Tools] = 1, 1
	Discover(w)
	if !w.Has("trapping") {
		t.Fatal("a thin forest and tools in hand should teach trapping")
	}
}

func TestWornFieldsTeachIrrigation(t *testing.T) {
	w := pressured(43)
	n := 0
	for i := range w.Grid.Tiles {
		p := entity.Pos{X: i % w.Grid.W, Y: i / w.Grid.W}
		t := &w.Grid.Tiles[i]
		if t.Terrain == world.Grass && entity.Dist(p, w.MarketPos) < 5 && n < 4 {
			t.Terrain, t.Fertility, t.Rich = world.Field, 0.15, 0.6
			n++
		}
	}
	Discover(w)
	if !w.Has("irrigation") {
		t.Fatal("worn fields near the market should teach irrigation")
	}
	if w.ReachFloor[action.Index(action.Irrigate)] < action.Opened {
		t.Fatal("irrigation should open irrigate to everyone")
	}
}

func TestClearedWoodsTeachForestry(t *testing.T) {
	w := pressured(44)
	cleared := 0
	for i := range w.Grid.Tiles {
		if w.Grid.Tiles[i].Terrain == world.Forest && float64(cleared) < 0.5*float64(w.Forest0) {
			w.Grid.Tiles[i].Terrain = world.Grass
			cleared++
		}
	}
	Discover(w)
	if !w.Has("forestry") {
		t.Fatalf("clearing half the forest should teach forestry (forest0 %d, cleared %d)", w.Forest0, cleared)
	}
	if w.Mods.Regrowth <= 1 {
		t.Fatal("forestry should make the land come back faster")
	}
}

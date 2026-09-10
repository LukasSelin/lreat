package system

import (
	"testing"

	"lreat/core/action"
	"lreat/core/belief"
	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
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

// wanting makes everybody feel the moment the land's answers are answers to,
// because the land being thin is no longer enough on its own: somebody has
// to be hungry about it. See the Signature on each discovery.
func wanting(w *world.World) {
	for _, a := range w.Agents {
		a.Needs[need.Physiological] = 0
		a.Needs[need.Belonging] = 0
		a.Needs[need.Actualization] = 0
		for i := range a.Personality {
			a.Personality[i] = 1
		}
		a.Norms = belief.Norms{1, 1, 1, 1}
	}
}

// insist runs the discovery pass over and over, which is what a settlement
// does with a want: a thing is not worked out on the day it is first needed.
// A year of days is enough for anything the land answers.
func insist(w *world.World) {
	for i := 0; i < 20*clock.Year; i++ {
		Discover(w)
		w.Tick++
	}
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
	wanting(w)
	Discover(w)
	if w.Has("fishing") {
		t.Fatal("fishing arrived on the first day of wanting it; the land's answers are worked toward, not handed over")
	}
	insist(w)
	if !w.Has("fishing") {
		t.Fatal("a thin forest by a river, wanted long enough, should teach fishing")
	}
	if w.ReachFloor[action.Index(action.Fish)] < action.Opened {
		t.Fatal("fishing should open fish to everyone")
	}
	if w.Has("trapping") {
		t.Fatal("trapping should wait for tools as well")
	}
	w.Agents[0].Inventory[entity.Tools], w.Agents[1].Inventory[entity.Tools] = 1, 1
	insist(w)
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
	wanting(w)
	insist(w)
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
			w.Grid.Turn(w.Grid.PosOf(i), world.Grass)
			cleared++
		}
	}
	wanting(w)
	insist(w)
	if !w.Has("forestry") {
		t.Fatalf("clearing half the forest should teach forestry (forest0 %d, cleared %d)", w.Forest0, cleared)
	}
	if w.Mods.Regrowth <= 1 {
		t.Fatal("forestry should make the land come back faster")
	}
}

// The point of measuring in pressure rather than in time: a settlement with
// nothing much wrong with it makes no progress toward anything, however long
// it stands in exactly the circumstances that would teach a hungrier one.
// Nothing anywhere says a comfortable people stagnate. It is what adding
// nearly nothing for twenty years comes to.
func TestAComfortableSettlementWorksNothingOut(t *testing.T) {
	thin := func(w *world.World) {
		for i := range w.Grid.Tiles {
			if w.Grid.Tiles[i].Terrain == world.Forest {
				w.Grid.Tiles[i].Wild = 0.1
			}
		}
	}

	hungry := pressured(51)
	thin(hungry)
	wanting(hungry)
	insist(hungry)

	easy := pressured(51)
	thin(easy)
	for _, a := range easy.Agents {
		for i := range a.Needs {
			a.Needs[i] = 1 // wanting for nothing
		}
	}
	insist(easy)

	if !hungry.Has("fishing") {
		t.Fatal("the hungry settlement should have worked out fishing")
	}
	if easy.Has("fishing") {
		t.Fatalf("the comfortable settlement worked out fishing on %.0f of pressure; want it to stay where it is",
			easy.Pressed("fishing"))
	}
	if easy.Pressed("fishing") >= hungry.Pressed("fishing") {
		t.Fatalf("comfort pressed %.1f and want pressed %.1f; want should press harder",
			easy.Pressed("fishing"), hungry.Pressed("fishing"))
	}
}

// A discovery paced by pressure names whoever was most in the moment for it,
// which is the whole of what the per-agent reading buys: a technology stops
// being a thing that happened to a settlement.
func TestADiscoveryNamesWhoeverFeltItWorst(t *testing.T) {
	w := pressured(52)
	for i := range w.Grid.Tiles {
		if w.Grid.Tiles[i].Terrain == world.Forest {
			w.Grid.Tiles[i].Wild = 0.1
		}
	}
	wanting(w)
	insist(w)
	if !w.Has("fishing") {
		t.Fatal("fishing should have arrived")
	}
	var found bool
	for _, e := range w.Log.All() {
		if e.Kind == event.Discovered && e.Actor != 0 {
			found = true
		}
	}
	if !found {
		t.Fatal("no discovery named the person it happened to")
	}
}

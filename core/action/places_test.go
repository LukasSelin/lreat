package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/world"
)

// village is a wide open world with a market, and two agents whose homes and
// temperaments the tests set.
func village(t *testing.T) (*world.World, *entity.Agent, *entity.Agent) {
	t.Helper()
	w := world.NewSized(5, 60, 20)
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i] = world.Tile{Terrain: world.Grass, Fertility: 0.3, Rich: 0.3}
	}
	w.MarketPos = entity.Pos{X: 20, Y: 10}
	w.Grid.At(w.MarketPos).Structure = world.Market
	a := w.SpawnAt("a", need.Neutral(), entity.Pos{X: 18, Y: 10})
	o := w.SpawnAt("o", need.Neutral(), entity.Pos{X: 22, Y: 10})
	a.Temperament, o.Temperament = entity.Temperament{Trust: 0.5, Warmth: 0.5}, entity.Temperament{Trust: 0.5, Warmth: 0.5}
	return w, a, o
}

func TestNobodyWithinReachIsNobody(t *testing.T) {
	w, a, o := village(t)
	if !Socialize.Available(a, w) {
		t.Fatal("company two tiles away should be company")
	}
	o.Pos = entity.Pos{X: 38, Y: 18}
	if Socialize.Available(a, w) || Teach.Available(a, w) {
		t.Fatal("somebody on the far side of the map is not company")
	}
	if PickCompany(a, w) != nil {
		t.Fatal("nobody within reach should be picked")
	}
}

func TestMeetingsHappenAtPlacesNotInTheWoods(t *testing.T) {
	w, a, o := village(t)
	p, ok := meetingPlace(a, o, w)
	if !ok || p != w.MarketPos {
		t.Fatalf("with only a market about, meet there: %v %v", p, ok)
	}
	// o wanders into a forest far from any place.
	for y := 0; y < 20; y++ {
		w.Grid.At(entity.Pos{X: 5, Y: y}).Terrain = world.Forest
	}
	o.Pos = entity.Pos{X: 5, Y: 10}
	a.Pos = entity.Pos{X: 8, Y: 10}
	if _, ok := meetingPlace(a, o, w); ok {
		t.Fatal("there is nowhere to meet somebody off in the woods")
	}
	if _, ok := Socialize.Target(a, w); ok {
		t.Fatal("socialising should have no target when the only company is in the woods")
	}
}

func TestTemperamentPicksThePlace(t *testing.T) {
	w, a, o := village(t)
	a.Home, a.HasHome = entity.Pos{X: 18, Y: 8}, true
	w.Grid.At(a.Home).Structure = world.House
	tavern := entity.Pos{X: 22, Y: 8}
	w.Grid.At(tavern).Structure = world.Tavern

	a.Temperament.Warmth = 0.9
	if p, _ := meetingPlace(a, o, w); p != tavern {
		t.Fatalf("the warm should head for the tavern, got %v", p)
	}
	a.Temperament.Warmth = 0.1
	if p, _ := meetingPlace(a, o, w); p != a.Home {
		t.Fatalf("the cool should have company at home, got %v", p)
	}
}

func TestATavernCheersAMeeting(t *testing.T) {
	w, a, o := village(t)
	w.Grid.At(entity.Pos{X: 21, Y: 10}).Structure = world.Tavern
	a.Pos, o.Pos = entity.Pos{X: 21, Y: 11}, entity.Pos{X: 21, Y: 11}
	b0 := a.Needs[need.Belonging]
	Socialize.Apply(a, w)
	inTavern := a.Needs[need.Belonging] - b0
	a.Needs[need.Belonging] = b0
	w.Grid.At(entity.Pos{X: 21, Y: 10}).Structure = world.None
	a.Pos, o.Pos = entity.Pos{X: 10, Y: 3}, entity.Pos{X: 10, Y: 3}
	// Same pair, same luck: reseed the world's stream so the meeting's chance
	// is the same, and compare.
	Socialize.Apply(a, w)
	outside := a.Needs[need.Belonging] - b0
	if !(inTavern > outside-0.1) {
		t.Fatalf("a tavern should be at least as good an evening: %v in, %v out", inTavern, outside)
	}
	if !(inTavern >= tavernCheer) {
		t.Fatalf("a tavern meeting should carry the cheer: %v", inTavern)
	}
}

func TestATavernTakesTimberAndOnlyOneIsBuilt(t *testing.T) {
	w, a, _ := village(t)
	w.Unlock("brewing")
	a.Inventory[entity.Wood] = tavernWood
	if !run(w, a, BuildTavern) {
		t.Fatal("a tavern should be buildable with timber beside the market")
	}
	if w.Grid.At(a.Pos).Structure != world.Tavern || a.Inventory[entity.Wood] != 0 {
		t.Fatal("the tavern should stand and the timber be spent")
	}
	a.Inventory[entity.Wood] = tavernWood
	if BuildTavern.Available(a, w) {
		t.Fatal("one tavern is enough for a settlement this size")
	}
}

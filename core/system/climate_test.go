package system

import (
	"testing"

	"lreat/core/action"
	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/world"
)

// forestAt makes p a forest tile stripped of what grows on it, so that
// whatever the season puts back is measurable.
func forestAt(w *world.World, p entity.Pos) *world.Tile {
	t := w.Grid.At(p)
	t.Terrain, t.Wood, t.Wild = world.Forest, 0.1, 0.1
	t.Standing() // a wood that has stood a while, not a planting
	return t
}

// A forest keeps the season's hours: it puts wood and wild food back fast in
// the warmth and barely at all in a frost.
func TestLittleGrowsInAFrost(t *testing.T) {
	p := entity.Pos{X: 5, Y: 5}

	warm := fitWorld(3)
	warm.Climate = world.Climate{Temp: world.Thrive + 2}
	tile := forestAt(warm, p)
	before := tile.Wood
	Land(warm)
	grown := warm.Grid.At(p).Wood - before

	cold := fitWorld(3)
	cold.Climate = world.Climate{Temp: world.Frost - 5}
	tile = forestAt(cold, p)
	before = tile.Wood
	Land(cold)
	frozen := cold.Grid.At(p).Wood - before

	if grown <= 0 {
		t.Fatalf("summer put back %.6f of a forest, want growth", grown)
	}
	if frozen >= grown/2 {
		t.Errorf("a frost put back %.6f of a forest against summer's %.6f, which is no season at all", frozen, grown)
	}
}

// What winter costs is being caught out in it. Two identical bodies in the
// same cold, one under a roof and one not, should not fare the same; in
// mild weather the roof makes no difference to the larder at all.
func TestAWinterOutdoorsCostsWhatAWinterIndoorsDoesNot(t *testing.T) {
	spend := func(temp, shelter float64) float64 {
		w := fitWorld(5)
		w.Climate = world.Climate{Temp: temp}
		a := w.Spawn("a", need.Neutral())
		a.Shelter = shelter
		a.Needs[need.Physiological] = 1
		Decay(w)
		return 1 - a.Needs[need.Physiological]
	}
	housed := spend(world.Bitter, 1)
	exposed := spend(world.Bitter, 0)
	if exposed <= housed {
		t.Errorf("a winter outdoors cost %.4f, a winter indoors %.4f", exposed, housed)
	}
	if got := exposed - housed; got < ColdDrain*0.9 {
		t.Errorf("the deepest cold costs the unhoused %.4f, want about %.4f", got, ColdDrain)
	}
	if spend(world.Mild+5, 0) != spend(world.Mild+5, 1) {
		t.Error("a mild season costs the unhoused more than the housed")
	}
}

// The moment reads different in different weather, and the acts that belong
// to the cold and to the growing season swap places with it. This is what
// makes the settlement's year a year rather than a number on a panel.
func TestTheSeasonChangesWhatFits(t *testing.T) {
	fit := func(temp float64, d *action.Def, i int) float64 {
		w := fitWorld(9)
		w.Climate = world.Climate{Temp: temp}
		a := w.Spawn("a", need.Neutral())
		a.Shelter = 0.4
		a.Inventory[entity.Wood] = 1
		action.Imprint(a)
		return habit.Cosine(action.Shared(a, w), a.Habits[i])
	}
	var farm, wood int
	for i, d := range action.Catalog {
		switch d {
		case action.Farm:
			farm = i
		case action.GatherWood:
			wood = i
		}
	}
	summer, winter := world.Thrive+5, world.Bitter
	if fit(summer, action.Farm, farm) <= fit(winter, action.Farm, farm) {
		t.Error("farming does not fit summer better than winter")
	}
	if fit(winter, action.GatherWood, wood) <= fit(summer, action.GatherWood, wood) {
		t.Error("gathering wood does not fit winter better than summer")
	}
}

// The weather is part of the world, so it is part of a reproducible run.
func TestWeatherIsPartOfTheSeededRun(t *testing.T) {
	a, b := fitWorld(21), fitWorld(21)
	populate(a, 5)
	populate(b, 5)
	Run(a, 300)
	Run(b, 300)
	if a.Climate != b.Climate {
		t.Fatalf("same seed, different weather: %+v vs %+v", a.Climate, b.Climate)
	}
	if a.Climate.Temp == world.MeanTemp {
		t.Error("the weather never moved")
	}
}

// A cold store is the oldest one there is. What the year takes from the
// settlement in the growing it stops doing, it gives back a little of in the
// keeping - and a granary does the same thing in August that the weather
// does in January.
func TestFoodKeepsInTheColdAndInTheGranary(t *testing.T) {
	left := func(temp float64, granaries int) float64 {
		w := fitWorld(11)
		w.Climate = world.Climate{Temp: temp}
		for i := 0; i < granaries; i++ {
			w.Mods.Keeping *= 0.4
		}
		w.Market.Stock[entity.Food] = 100
		MarketStep(w)
		return w.Market.Stock[entity.Food]
	}
	summer, winter := world.Mild+5, world.Bitter
	if !(left(winter, 0) > left(summer, 0)) {
		t.Errorf("food keeps no better in the cold: %.4f left against %.4f", left(winter, 0), left(summer, 0))
	}
	if !(left(summer, 1) > left(summer, 0)) {
		t.Errorf("a granary keeps nothing: %.4f left against %.4f", left(summer, 1), left(summer, 0))
	}
	// The two are worth having together: a granary through a winter keeps
	// what neither keeps alone, which is what lets an autumn be put by.
	if !(left(winter, 1) > left(winter, 0) && left(winter, 1) > left(summer, 1)) {
		t.Errorf("a granary in winter (%.4f) does not beat a granary in summer (%.4f) or a winter without one (%.4f)",
			left(winter, 1), left(summer, 1), left(winter, 0))
	}
	if got := Keeping(&world.World{Mods: world.DefaultModifiers(), Climate: world.Climate{Temp: world.Mild + 5}}); got != 1 {
		t.Errorf("mild weather keeps %.2f of the usual spoilage, want all of it", got)
	}
}

// What an agent carries goes off slowly, and the cold is the whole of the
// mercy: a January larder beats a June one, and no granary reaches either.
func TestWhatIsCarriedRotsToo(t *testing.T) {
	held := func(temp float64, granaries int) float64 {
		w := fitWorld(13)
		w.Climate = world.Climate{Temp: temp}
		for i := 0; i < granaries; i++ {
			w.Mods.Keeping *= 0.4
		}
		a := w.Spawn("a", need.Neutral())
		a.Inventory[entity.Food] = 10
		Spoil(w, a)
		return a.Inventory[entity.Food]
	}
	summer, winter := world.Mild+5, world.Bitter
	if got := held(summer, 0); got >= 10 {
		t.Errorf("food in hand kept perfectly through a summer: %.4f of 10", got)
	}
	if !(held(winter, 0) > held(summer, 0)) {
		t.Error("a larder keeps no better in the cold, which is the whole point of it")
	}
	if held(summer, 3) != held(summer, 0) {
		t.Error("the settlement's granaries kept an agent's own food, which is the market's business")
	}
	// Slowly is the point: a day's spoilage should be nothing beside a
	// day's eating, or the larder is a tax on carrying food at all.
	if lost := 10 - held(summer, 0); lost > 0.1 {
		t.Errorf("a tick took %.4f of ten units, which is not a low rate", lost)
	}
}

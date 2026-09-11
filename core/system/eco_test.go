package system

import (
	"testing"

	"lreat/core/action"
	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// What the creatures do to the country: a herd holds a young wood back and
// leaves an old one alone, a warren keeps the ground it grazes open and
// makes it richer, and a sounder turns the soil and spreads the wood.

// woodBeside is a wood tile near the square with buildable open ground
// beside it that would hold a wood.
func woodBeside(t *testing.T, w *world.World) (entity.Pos, entity.Pos) {
	t.Helper()
	g := w.Grid
	var open entity.Pos
	wood, ok := g.NearestOfKind(w.MarketPos, 40, world.KindsOf(ontology.Wood), func(p entity.Pos, tile *world.Tile) bool {
		if !tile.Is(ontology.Wood) {
			return false
		}
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				q := g.Norm(entity.Pos{X: p.X + dx, Y: p.Y + dy})
				if (dx != 0 || dy != 0) && g.In(q) && g.At(q).Buildable() && g.HoldsWood(q) {
					open = q
					return true
				}
			}
		}
		return false
	})
	if !ok {
		t.Fatal("no wood with open ground beside it near the square")
	}
	return wood, open
}

func TestBrowsingHoldsAThicketBackAndLeavesAWoodAlone(t *testing.T) {
	w := world.New(6)
	g := w.Grid
	wood, _ := woodBeside(t, w)
	i := g.Index(wood)
	deer := w.SpawnKind(entity.Deer, "doe", w.RandomPersonality(), wood)
	deer.Needs[need.Physiological] = 0.2

	g.Age[i] = 300 // a thicket a year in
	g.Wild[i] = 1
	action.Browse.Apply(deer, w)
	if g.Age[i] >= 300 {
		t.Errorf("a browsed thicket is %.0f days on; want it set back", g.Age[i])
	}

	g.Standing(i) // an old wood
	was := g.Age[i]
	g.Wild[i] = 1
	action.Browse.Apply(deer, w)
	if g.Age[i] != was {
		t.Errorf("a browsed old wood moved from %.0f to %.0f days; an old wood does not mind a herd", was, g.Age[i])
	}
}

func TestGrazingManuresTheMeadowAndKeepsItOpen(t *testing.T) {
	w := world.New(6)
	g := w.Grid
	_, open := woodBeside(t, w)
	i := g.Index(open)
	hare := w.SpawnKind(entity.Hare, "hare", w.RandomPersonality(), open)
	hare.Needs[need.Physiological] = 0.2
	g.Rich[i], g.Fertility[i], g.Sward[i] = 0.4, 0.4, 1
	action.Graze.Apply(hare, w)
	if g.Rich[i] <= 0.4 || g.Fertility[i] <= 0.4 {
		t.Errorf("after a grazing the ground holds %.3f of a best %.3f; want both up", g.Fertility[i], g.Rich[i])
	}
	if g.SeedTakes(i) != (g.Sward[i] >= world.SeedTakes) {
		t.Fatal("SeedTakes does not read the sward")
	}
	g.Sward[i] = world.SeedTakes / 2
	if g.SeedTakes(i) {
		t.Error("a seed takes on ground grazed to half of what it needs")
	}
	// Over a fresh valley with every sward grazed away, no wood comes back
	// for as long as the sward is down, which with nothing grazing it is
	// most of a year; with the sward standing, some does. The seed falls
	// the same either way, since the ground is asked before any chance is
	// spent on it.
	forestOn := func(grazed bool) int {
		w := world.New(6)
		if grazed {
			for i := range w.Grid.Sward {
				w.Grid.Sward[i] = 0
			}
		}
		for i := 0; i < 5; i++ {
			w.Spawn("a", w.RandomPersonality())
		}
		was := w.Grid.Forest()
		Run(w, 240)
		return w.Grid.Forest() - was
	}
	if grew := forestOn(true); grew > 0 {
		t.Errorf("%d tiles of wood came back over ground grazed bare", grew)
	}
	if grew := forestOn(false); grew <= 0 {
		t.Errorf("no wood came back over standing sward in eight months")
	}
}

func TestRootingTurnsTheSoilAndSpreadsTheWood(t *testing.T) {
	w := world.New(6)
	g := w.Grid
	wood, open := woodBeside(t, w)
	i, j := g.Index(wood), g.Index(open)
	boar := w.SpawnKind(entity.Boar, "boar", w.RandomPersonality(), wood)
	g.Rich[i], g.Fertility[i] = 0.4, 0.4
	planted := false
	for k := 0; k < 200 && !planted; k++ {
		boar.Needs[need.Physiological] = 0.2
		g.Wild[i] = 1
		action.Root.Apply(boar, w)
		planted = g.At(open).Is(ontology.Wood) || g.Tiles[j].Terrain == world.Forest
	}
	if g.Rich[i] <= 0.4 {
		t.Error("rooting did not turn the soil")
	}
	// Any of the open tiles round the wood may be the one planted: the
	// first by the bearings that would hold a wood.
	planted = false
	for dy := -1; dy <= 1 && !planted; dy++ {
		for dx := -1; dx <= 1; dx++ {
			q := g.Norm(entity.Pos{X: wood.X + dx, Y: wood.Y + dy})
			if (dx != 0 || dy != 0) && g.In(q) && g.At(q).Is(ontology.Wood) && g.Age[g.Index(q)] == 0 {
				planted = true
				break
			}
		}
	}
	if !planted {
		t.Error("two hundred rootings beside open ground planted no wood")
	}
}

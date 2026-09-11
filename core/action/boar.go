package action

import (
	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// A boar roots in the wood for the same brush a deer browses, and raids the
// fields: a sounder in a strip in ear takes the grain and tramples what it
// does not take. It wallows, bolts from anybody near - though not from as
// far off as a deer does - keeps to its sounder, and ranges.
var (
	Root    = feeding("root", ontology.Browse, ontology.Wood)
	Wallow  = bedding("wallow")
	Bolt    = fleeing("bolt")
	Sounder = herding("sounder")
	Range   = roaming("range")
)

const (
	// raidWear is the fertility a raid takes off a strip, a few harvests'
	// worth of wear in a night; raidSetback is how much of the crop's
	// coming on is trampled, as a share of the age it had.
	raidWear    = 0.02
	raidSetback = 0.5
	raidGain    = 0.4
	// raidLeast is how far along a strip's crop must be for a raid to be
	// worth the walk: something has to be standing in it.
	raidLeast = 0.25
)

// raidSite is the nearest strip with a crop worth raiding.
func raidSite(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	kinds := world.KindsOf(ontology.Field)
	return w.Grid.NearestOfKind(a.Pos, ranging(a), kinds, func(p entity.Pos, t *world.Tile) bool {
		return t.Is(ontology.Field) && w.Grid.Along(w.Grid.Index(p), ontology.Crop) >= raidLeast
	})
}

// Raid is a boar in a field. What it eats comes off the crop, and what it
// tramples sets the crop back and wears the strip, so a farmer whose field
// is raided reaps later and less. Nobody answers it yet; that is for the
// day people hunt what is on the map.
var Raid = &Def{
	Name: "raid", Ticks: 1, Available: always, Target: raidSite, Supply: belly,
	Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
		return need.Levels{need.Physiological: raidGain}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		g := w.Grid
		i := g.Index(a.Pos)
		if !g.At(a.Pos).Is(ontology.Field) || g.Along(i, ontology.Crop) < raidLeast {
			return
		}
		g.Fertility[i] = max(0, g.Fertility[i]-raidWear)
		g.Age[i] *= 1 - raidSetback
		a.Needs.Add(need.Physiological, raidGain)
	},
}

func init() {
	wary(entity.Boar, Root, Wallow, Bolt, Sounder, Range, feedPrior)
	// A raid is a taking of grain off a field, as the ontology composes it:
	// hungry, short, near, and of the warm half of the year, which is when
	// there is grain standing to be had.
	seed(Raid, reachEveryday, habit.Signature{habit.Hunger: 0.8, habit.Near: 0.6, habit.Lack: 0.8, habit.Chill: -0.5})
}

package action

import (
	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// What a deer does, and what a deer senses.
//
// A deer decides as a person does: its needs press, the moment is read as a
// point in habit space, and the act whose habit lies nearest is drawn. What
// differs is what is in the moment and what can be drawn. A deer reads the
// same coordinates a person reads and finds different things on them - a
// wood for a roof, a herd for company, a person in earshot where a settler
// reads the settlement's order - and it has five acts of its own, keyed by
// its actor in the ontology and kept in slots after everything a person
// does. Nothing here draws on any chance: a target is asked of every
// candidate every decision, and again by anybody looking at the deer, so
// what a deer would do must not move for having been looked at.

const (
	// alarmRadius is how far off a deer notices a person. herdRadius is
	// how near another deer must be to count as company, and fleeRange how
	// far a flight goes. All three are within world.NearbyLimit, past which
	// the agent file will not be asked, and within world.IslandReach, so
	// that deer and people out of each other's reach can act at once.
	alarmRadius = 6
	herdRadius  = 8
	fleeRange   = 6

	// browseTake is what one browsing takes off the brush, and browseLeast
	// how thin a stand may be and still be worth going to. The brush is the
	// one count a forager picks and a trapper hunts over, so a herd near a
	// settlement thins what its people find in the wood.
	// A browse takes what a forage takes, so a deer costs the wood what a
	// forager does and the two are in plain competition for it; what the
	// brush puts on in a year is then what bounds the herd.
	browseTake  = 0.08
	browseLeast = 0.15
	// browseGain is what a full browsing restores. A deer eats every day it
	// can and keeps no larder, so this is what a day's grazing is worth
	// against the day's burn.
	browseGain = 0.4
	// bedGain is what bedding down restores, and it is next to nothing on
	// purpose. A person's rest restores a little more than a day burns,
	// which is harmless for a person, who is also walking, working and
	// wearing out shelter; a deer that got that from lying down never had
	// to eat, and a herd of five hundred lay in the wood with full bellies
	// and the brush untouched.
	bedGain = 0.005

	fleeGain = 0.2
	herdGain = 0.15
	roamGain = 0.1
	// roamSpan is how many days a deer keeps a bearing before turning.
	roamSpan = 30
)

// Cover is how sheltered a creature standing at p is from the weather and
// from being seen: whole under trees, nothing in the open. It is what a deer
// has instead of a roof.
func Cover(w *world.World, p entity.Pos) float64 {
	if w.Grid.At(p).Is(ontology.Wood) {
		return 1
	}
	return 0
}

// Alarm is the nearest person within a deer's notice, or nil.
func Alarm(a *entity.Agent, w *world.World) *entity.Agent {
	return w.ClosestOf(a.Pos, alarmRadius, entity.Human, func(*entity.Agent) bool { return true })
}

// InHerd reports whether another of the creature's kind stands within
// herd reach: the company a deer has instead of neighbours.
func InHerd(a *entity.Agent, w *world.World) bool { return w.Neighbor(a, herdRadius) != nil }

// fellow is the nearest other deer the file can be asked for, or nil.
func fellow(a *entity.Agent, w *world.World) *entity.Agent {
	return w.ClosestOf(a.Pos, world.NearbyLimit, a.Species(), func(o *entity.Agent) bool { return o != a })
}

// belly is a deer's Supply: a deer keeps no pack, so what it is short of is
// read off its hunger, and it is well supplied in nothing.
func belly(a *entity.Agent) (lack, stock float64) {
	return bipolar(1 - a.Needs[need.Physiological]), 0
}

// senseDeer is the shared part of a deer's moment. The first five
// coordinates are its urgencies, as anybody's. Shelter is cover: a deer
// under trees is housed and one in the open is not. Company is a herd
// within reach. Chill is the weather and exposure the weather on a deer
// with no trees over it. Order is the alarm: a person in earshot reads as
// the loudest the coordinate goes, the way an ungoverned square does to a
// settler, and it is silent otherwise. A deer holds nothing to be right and
// has learned no reprisal, so the moral coordinates are silent.
func senseDeer(a *entity.Agent, w *world.World) habit.Signature {
	var s habit.Signature
	u := need.Urgencies(a.Needs)
	for t := range u {
		s[t] = bipolar(u[t] * a.Personality[t])
	}
	cover := Cover(w, a.Pos)
	s[habit.Shelter] = cover - 1
	if InHerd(a, w) {
		s[habit.Company] = 1
	} else {
		s[habit.Company] = -1
	}
	chill := need.Clamp(w.ChillAt(a.Pos))
	s[habit.Chill] = chill
	s[habit.Exposure] = need.Clamp(chill * (1 - cover))
	if Alarm(a, w) != nil {
		s[habit.Order] = -1
	}
	return s
}

// browseSite is the nearest stand of brush worth browsing, which is very
// often the tile the deer stands on.
func browseSite(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	held := world.StockOf(ontology.Browse)
	return w.Grid.NearestOfKind(a.Pos, ranging(a), world.KindsOf(ontology.Wood), func(p entity.Pos, t *world.Tile) bool {
		return t.Is(ontology.Wood) && held(w.Grid)[w.Grid.Index(p)] >= browseLeast
	})
}

// Browse is a deer eating: the brush under it, a little off the count the
// wood keeps, into the belly rather than into any pack.
var Browse = &Def{
	Name: "browse", Ticks: 1, Available: always, Target: browseSite, Supply: belly,
	Expect: func(a *entity.Agent, _ *world.World, _ entity.Pos) need.Levels {
		return need.Levels{need.Physiological: browseGain}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		s, ok := w.Grid.Stock(w.Grid.Index(a.Pos), ontology.Browse)
		if !ok || !w.Grid.At(a.Pos).Is(ontology.Wood) {
			return
		}
		take := min(browseTake, max(0, *s))
		*s -= take
		a.Needs.Add(need.Physiological, browseGain*take/browseTake)
	},
}

// BedDown is a deer's rest, and the fallback for a deer as resting is for a
// person: always there, and worth next to nothing.
var BedDown = &Def{
	Name: "bed down", Ticks: 1, Available: always, Target: here, Supply: belly,
	Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
		return need.Levels{need.Physiological: bedGain}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		a.Needs.Add(need.Physiological, bedGain)
	},
}

// fleeTo is where a deer runs from o: the tile within fleeRange furthest
// from them, under trees for choice, and never into deep water. Ties go to
// the first found, walking the square in a fixed order.
func fleeTo(a *entity.Agent, w *world.World, o *entity.Agent) (entity.Pos, bool) {
	g := w.Grid
	best, bestScore := a.Pos, -1
	for dy := -fleeRange; dy <= fleeRange; dy++ {
		for dx := -fleeRange; dx <= fleeRange; dx++ {
			p := g.Norm(entity.Pos{X: a.Pos.X + dx, Y: a.Pos.Y + dy})
			if !g.In(p) || g.At(p).Deep() {
				continue
			}
			score := 2 * g.Dist(p, o.Pos)
			if g.At(p).Is(ontology.Wood) {
				score++
			}
			if score > bestScore {
				best, bestScore = p, score
			}
		}
	}
	return best, best != a.Pos
}

// Flee is the alarmed deer putting ground between itself and a person.
var Flee = &Def{
	Name: "flee", Ticks: 1, Supply: belly,
	Available: func(a *entity.Agent, w *world.World) bool { return Alarm(a, w) != nil },
	Target: func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
		o := Alarm(a, w)
		if o == nil {
			return a.Pos, false
		}
		return fleeTo(a, w, o)
	},
	Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
		return need.Levels{need.Safety: fleeGain}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		a.Needs.Add(need.Safety, fleeGain)
	},
}

// Herd is a deer with no herd going to the nearest of its kind.
var Herd = &Def{
	Name: "herd", Ticks: 2, Supply: belly,
	Available: func(a *entity.Agent, w *world.World) bool {
		return !InHerd(a, w) && fellow(a, w) != nil
	},
	Target: func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
		o := fellow(a, w)
		if o == nil {
			return a.Pos, false
		}
		return o.Pos, true
	},
	Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
		return need.Levels{need.Belonging: herdGain}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		a.Needs.Add(need.Belonging, herdGain)
	},
}

// roamStep is how far along its bearing a deer looks for the next wood.
const roamStep = 8

// roamTo is the wood a deer rambles toward: a step out along a bearing it
// keeps for a month, chosen from who it is and what day it is rather than
// from any draw (the bearings are the scout's, in ground.go), and the
// nearest wood to that point. A bearing that leaves
// the map is a ramble not taken.
func roamTo(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	g := w.Grid
	b := bearings[(int(a.ID)+w.Tick/roamSpan)%len(bearings)]
	out := g.Norm(entity.Pos{X: a.Pos.X + b.X*roamStep, Y: a.Pos.Y + b.Y*roamStep})
	if !g.In(out) {
		return a.Pos, false
	}
	p, ok := g.NearestOfKind(out, roamStep, world.KindsOf(ontology.Wood), func(p entity.Pos, t *world.Tile) bool {
		return t.Is(ontology.Wood) && !t.Deep()
	})
	if !ok || p == a.Pos {
		return a.Pos, false
	}
	return p, true
}

// Roam is the fed deer's ramble.
var Roam = &Def{
	Name: "roam", Ticks: 1, Available: always, Target: roamTo, Supply: belly,
	Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
		return need.Levels{need.Actualization: roamGain}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		a.Needs.Add(need.Actualization, roamGain)
	},
}

func init() {
	// The hand priors here are what the ontology composes, written out so
	// the golden test that holds the two against each other has something
	// to hold. See priors.go for why what runs is the composed one.
	seed(Browse, reachEveryday, habit.Signature{habit.Hunger: 0.8, habit.Near: 0.6, habit.Lack: 0.8})
	seed(BedDown, reachEveryday, habit.Signature{
		habit.Hunger: -1, habit.Unsafe: -0.6, habit.Lonely: -0.6,
		habit.Unproven: -0.4, habit.Curious: -0.4,
	})
	seed(Flee, reachEveryday, habit.Signature{habit.Unsafe: 1, habit.Order: -1})
	seed(Herd, reachEveryday, habit.Signature{habit.Lonely: 0.7, habit.Company: -0.5})
	seed(Roam, reachEveryday, habit.Signature{habit.Curious: 0.6, habit.Hunger: -0.3, habit.Shelter: -0.4})
	Perceive(entity.Deer, senseDeer)
}

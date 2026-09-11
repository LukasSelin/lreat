package action

import (
	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// What a creature does, and what a creature senses.
//
// A creature decides as a person does: its needs press, the moment is read
// as a point in habit space, and the act whose habit lies nearest is drawn.
// What differs is what is in the moment and what can be drawn. A creature
// reads the same coordinates a person reads and finds different things on
// them - a wood for a roof, a herd for company, a person in earshot where a
// settler reads the settlement's order - and it has acts of its own, keyed
// by its actor in the ontology and kept in slots after everything a person
// does. Nothing here draws on any chance: a target is asked of every
// candidate every decision, and again by anybody looking at the creature,
// so what it would do must not move for having been looked at.
//
// The acts are made here and named by each kind in its own file, because a
// deer's bedding down and a boar's wallowing are the same act with two
// habits: a slot is a kind's own, so that a person never seeds a deer's and
// a deer never a boar's. What a kind is wary of and how near its own must
// stand to be company are the kind's, on its species, and the senses read
// them there. What every kind shares - how far a flight goes, what a
// browse takes, what a bed is worth - is here once.

const (
	// fleeRange is how far a flight goes. It is within world.NearbyLimit,
	// past which the agent file will not be asked, and within
	// world.IslandReach, so that creatures and people out of each other's
	// reach can act at once; so must every species' Wary and Herds be, and
	// a test says so.
	fleeRange = 6

	// browseTake is what one feeding takes off the stand, and browseLeast
	// how thin a stand may be and still be worth going to. A feeding takes
	// half what a forage takes and a thin stand still feeds, so a creature
	// lives on less ground than a forager does: what the stand puts on in a
	// year is still what bounds the herd, but a wood the settlement has
	// felled to a third of itself still carries a herd. At a forage's take
	// and a fifteenth for a stand, the herds followed the woods down and
	// were gone from a valley of a hundred people while the woods stood.
	browseTake  = 0.04
	browseLeast = 0.05
	// mastTake is what rooting takes off the timber count of an old wood:
	// the acorns and beech mast under the trees, a boar's own living that
	// no deer browses and no forager picks, so a sounder and a herd can
	// share a wood without one eating the other out of it. It is a small
	// bite, since the count is the trees and a sounder does not fell them.
	mastTake = 0.01
	// browseGain is what a full feeding restores. A creature eats every day
	// it can and keeps no larder, so this is what a day's grazing is worth
	// against the day's burn.
	browseGain = 0.5
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

	// What a feeding leaves behind it, besides less to eat.
	//
	// browseSetback is how many growing days a browsing sets a stand that
	// is still coming on back by. An old wood does not mind a herd; a
	// thicket browsed every day never comes on to timber, and a clearing
	// with a heavy herd in it stays a clearing. It is a stand's age and
	// not the calendar that is set back, so a felled wood the deer keep
	// down is exactly a wood that has had less growing weather.
	browseSetback = 20
	// dung is what a grazing leaves in the ground: open ground a warren
	// keeps creeps up toward the best soil there is, so the meadows a herd
	// has kept for years are the strips a settlement breaks next.
	dung = 0.01
	// rootTurns is what rooting does to the soil under the trees, turned
	// over and lifted the same way; rootPlant is the chance that a rooting
	// beside open ground plants a wood there, since a boar carries acorns
	// and buries most of them.
	rootTurns = 0.01
	rootPlant = 0.05
	// roamSpan is how many days a creature keeps a bearing before turning,
	// and roamStep how far along it it looks for the next cover.
	roamSpan = 30
	roamStep = 8
)

// Cover is how sheltered a creature is where it stands, from the weather
// and from being seen: whole under trees, what the kind makes of a hedge
// beside them, and nothing in the open. It is what a creature has instead
// of a roof.
func Cover(a *entity.Agent, w *world.World) float64 {
	g := w.Grid
	if g.At(a.Pos).Is(ontology.Wood) {
		return 1
	}
	if edge := a.Species().Edge; edge > 0 && g.HasNeighbor(a.Pos, func(t *world.Tile) bool { return t.Is(ontology.Wood) }) {
		return edge
	}
	return 0
}

// Alarm is the nearest person within the creature's notice, or nil.
func Alarm(a *entity.Agent, w *world.World) *entity.Agent {
	return w.ClosestOf(a.Pos, a.Species().Wary, entity.Human, func(*entity.Agent) bool { return true })
}

// InHerd reports whether another of the creature's kind stands within
// herd reach: the company a creature has instead of neighbours.
func InHerd(a *entity.Agent, w *world.World) bool { return w.Neighbor(a, a.Species().Herds) != nil }

// fellow is the nearest other of the creature's kind the file can be asked
// for, or nil.
func fellow(a *entity.Agent, w *world.World) *entity.Agent {
	return w.ClosestOf(a.Pos, world.NearbyLimit, a.Species(), func(o *entity.Agent) bool { return o != a })
}

// belly is a creature's Supply: it keeps no pack, so what it is short of is
// read off its hunger, and it is well supplied in nothing.
func belly(a *entity.Agent) (lack, stock float64) {
	return bipolar(1 - a.Needs[need.Physiological]), 0
}

// senseWary is the shared part of a creature's moment. The first five
// coordinates are its urgencies, as anybody's. Shelter is cover: a creature
// under trees is housed and one in the open is not. Company is a herd within
// reach. Chill is the weather and exposure the weather on a body with no
// trees over it. Order is the alarm: a person in earshot reads as the
// loudest the coordinate goes, the way an ungoverned square does to a
// settler, and it is silent otherwise. A creature holds nothing to be right
// and has learned no reprisal, so the moral coordinates are silent.
func senseWary(a *entity.Agent, w *world.World) habit.Signature {
	var s habit.Signature
	u := need.Urgencies(a.Needs)
	for t := range u {
		s[t] = bipolar(u[t] * a.Personality[t])
	}
	cover := Cover(a, w)
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

// feeding is a creature eating what stands on a kind of ground: the nearest
// stand of it worth going to, which is very often the tile it stands on,
// and a little off the count the ground keeps, into the belly rather than
// into any pack. Ground with something built on it is not grazed. What the
// feeding leaves behind it on the ground is the kind's, in leaves.
func feeding(name string, of, site *ontology.Class, take float64, leaves func(w *world.World, i int)) *Def {
	held := world.StockOf(of)
	kinds := world.KindsOf(site)
	return &Def{
		Name: name, Ticks: 1, Available: always, Supply: belly,
		Target: func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
			return w.Grid.NearestOfKind(a.Pos, ranging(a), kinds, func(p entity.Pos, t *world.Tile) bool {
				return t.Is(site) && t.Structure == world.None && held(w.Grid)[w.Grid.Index(p)] >= browseLeast
			})
		},
		Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
			return need.Levels{need.Physiological: browseGain}
		},
		Apply: func(a *entity.Agent, w *world.World) {
			t := w.Grid.At(a.Pos)
			s, ok := w.Grid.Stock(w.Grid.Index(a.Pos), of)
			if !ok || !t.Is(site) || t.Structure != world.None {
				return
			}
			took := min(take, max(0, *s))
			*s -= took
			a.Needs.Add(need.Physiological, browseGain*took/take)
			if leaves != nil && took > 0 {
				leaves(w, w.Grid.Index(a.Pos))
			}
		},
	}
}

// browsed is what a herd's browsing leaves a wood: a stand still coming on
// is set back, and an old wood is left alone.
func browsed(w *world.World, i int) {
	g := w.Grid
	if g.Along(i, ontology.Timbering) < 1 {
		g.Age[i] = max(0, g.Age[i]-browseSetback)
	}
}

// manured is what a grazing leaves open ground: a little richer.
func manured(w *world.World, i int) { enrich(w.Grid, i, dung) }

// enrich lifts what the ground could grow at best, and what it has in it,
// by d, up to the most any ground has.
func enrich(g *world.Grid, i int, d float64) {
	g.Rich[i] = min(1, g.Rich[i]+d)
	g.Fertility[i] = min(g.Rich[i], g.Fertility[i]+d)
}

// rooted is what a boar's rooting leaves a wood: the stand set back as a
// browsing sets it, the soil under it turned and lifted, and, now and
// then, a wood planted on the open ground beside it. The ground it takes
// on is the first, by the scout's bearings, that would hold a wood; the
// chance is the world's, as every act's is, and is spent only where there
// is such ground.
func rooted(w *world.World, i int) {
	g := w.Grid
	browsed(w, i)
	enrich(g, i, rootTurns)
	p := g.PosOf(i)
	for _, b := range bearings {
		q := g.Norm(entity.Pos{X: p.X + b.X, Y: p.Y + b.Y})
		if !g.In(q) || !g.At(q).Buildable() || !g.HoldsWood(q) {
			continue
		}
		if w.RNG.Float64() < rootPlant {
			j := g.Index(q)
			g.Turn(q, world.Forest)
			g.Wood[j], g.Wild[j] = 0, 0
			g.Sow(j) // a seedling wood, with nothing on it yet
		}
		return
	}
}

// bedding is a creature's rest, and the fallback for it as resting is for a
// person: always there, and worth next to nothing.
func bedding(name string) *Def {
	return &Def{
		Name: name, Ticks: 1, Available: always, Target: here, Supply: belly,
		Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
			return need.Levels{need.Physiological: bedGain}
		},
		Apply: func(a *entity.Agent, w *world.World) {
			a.Needs.Add(need.Physiological, bedGain)
		},
	}
}

// fleeTo is where a creature runs from o: the tile within fleeRange furthest
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

// fleeing is the alarmed creature putting ground between itself and a person.
func fleeing(name string) *Def {
	return &Def{
		Name: name, Ticks: 1, Supply: belly,
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
}

// herding is a creature with no herd going to the nearest of its kind.
func herding(name string) *Def {
	return &Def{
		Name: name, Ticks: 2, Supply: belly,
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
}

// roamTo is the cover a creature rambles toward: a step out along a bearing
// it keeps for a month, chosen from who it is and what day it is rather
// than from any draw (the bearings are the scout's, in ground.go), and the
// nearest wood to that point. A bearing that leaves the map is a ramble not
// taken.
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

// roaming is the fed creature's ramble, always toward trees, so it is also
// what one caught in the open does about it.
func roaming(name string) *Def {
	return &Def{
		Name: name, Ticks: 1, Available: always, Target: roamTo, Supply: belly,
		Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
			return need.Levels{need.Actualization: roamGain}
		},
		Apply: func(a *entity.Agent, w *world.World) {
			a.Needs.Add(need.Actualization, roamGain)
		},
	}
}

// The hand priors every wary creature's acts are held to, written out as
// the ontology composes them so the golden test that holds the two against
// each other has something to hold. See priors.go for why what runs is the
// composed one. A feeding's prior depends on what is fed on and where, and
// is written by each kind.
var (
	restPrior = habit.Signature{
		habit.Hunger: -1, habit.Unsafe: -0.6, habit.Lonely: -0.6,
		habit.Unproven: -0.4, habit.Curious: -0.4,
	}
	fleePrior = habit.Signature{habit.Unsafe: 1, habit.Order: -1}
	herdPrior = habit.Signature{habit.Lonely: 0.7, habit.Company: -0.5}
	roamPrior = habit.Signature{habit.Curious: 0.6, habit.Hunger: -0.3, habit.Shelter: -0.4}
	// A feeding off a wood or the open: hungry, short, and near.
	feedPrior = habit.Signature{habit.Hunger: 0.8, habit.Near: 0.6, habit.Lack: 0.8}
)

// wary seeds the five acts every wary creature has and gives the kind its
// sense. Each kind calls it from its own file with its own acts.
func wary(sp *entity.Species, feed, bed, flee, herd, roam *Def, feeds habit.Signature) {
	seed(feed, reachEveryday, feeds)
	seed(bed, reachEveryday, restPrior)
	seed(flee, reachEveryday, fleePrior)
	seed(herd, reachEveryday, herdPrior)
	seed(roam, reachEveryday, roamPrior)
	Perceive(sp, senseWary)
}

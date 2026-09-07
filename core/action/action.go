// Package action is the catalog of things an agent can do.
//
// Each action declares where it happens and what it expects to give per need
// tier. The decision system multiplies those expectations by the agent's
// current urgencies and personality, then divides by the time it takes,
// travel included. Emergence lives here: add an action and the whole
// population can discover new ways to live.
package action

import (
	"math"

	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// Def describes one action.
type Def struct {
	// Key is the act's name in the ontology, take/timber@wood: what it is
	// in terms of what it does with what and where. It is what habit
	// slots are keyed by. Name is what it is called.
	Key   string
	Name  string
	Ticks int
	// Available reports whether the agent can start the action now.
	Available func(a *entity.Agent, w *world.World) bool
	// Target says where the action must be performed. ok is false when no
	// suitable place exists, which makes the action impossible for now.
	Target func(a *entity.Agent, w *world.World) (pos entity.Pos, ok bool)
	// Expect estimates the satisfaction gained per tier on completion at target.
	Expect func(a *entity.Agent, w *world.World, target entity.Pos) need.Levels
	// Apply mutates the world when the action completes. The agent is at
	// its plan's target by then, so Apply works on a.Pos.
	Apply func(a *entity.Agent, w *world.World)

	// Prior is the kind of moment this action belongs to, the signature every
	// agent starts from before experience moves its own copy. It is the seed
	// of recognition-based choice; Expect remains the seed of value-based
	// choice. See package habit. It is composed from what the act is about
	// by the ontology; Tuned is the prior written by hand before it was,
	// kept so the two can be held against each other.
	Prior habit.Signature
	Tuned habit.Signature
	// Reach0 is how far into reach the action starts for a newborn, in
	// [0,1]. Ordinary living starts at 1. Crafts and learning start lower and
	// are brought closer by study, teaching, and discovery.
	Reach0 float64
	// Skilled names the skill the action draws on for this agent right now,
	// if any. It sets the skill dimension of the situation the agent sees
	// for this candidate. Nil means the action takes no skill.
	Skilled func(a *entity.Agent, w *world.World) (entity.Skill, bool)
	// With is the other agent the action is done to or with, given where it
	// would be done, so that rapport toward that person is part of the
	// situation. Nil means the action involves nobody in particular.
	With func(a *entity.Agent, w *world.World, target entity.Pos) *entity.Agent
}

// Count is the size of the catalog, set once it is assembled.
var Count int

// Catalog lists every action in a fixed order: the order of their keys,
// which is what per-agent habit tables are indexed by. It is what the
// ontology entails, each act bound to its mechanics below. It is assembled
// in init rather than declared, because some actions reach back into the
// catalog when they run (study broadens reach, teaching passes it on) and a
// declaration would make that a cycle.
var Catalog []*Def

// instances is what the ontology entailed, by key.
var instances = func() map[string]ontology.Instance {
	m := map[string]ontology.Instance{}
	for _, in := range ontology.Instantiate() {
		m[in.Key] = in
	}
	return m
}()

// Derived selects the composed prior over the hand-tuned one.
const Derived = true

// mechanics binds an act the ontology entails to the code that carries it
// out. Moves, makings, and raisings are not listed: one interpreter
// carries each, from the move (see move.go), the recipe (see make.go), or
// the plan (see raise.go), and those named here are named only because
// the rest of the package refers to them. An act the ontology entails and nothing carries
// is a catalog that cannot be assembled, and says so at start.
var mechanics = map[string]*Def{
	"take/berries@wood":                 Forage,
	"take/game@wood":                    Hunt,
	"take/timber@wood":                  GatherWood,
	"take/fish@water":                   Fish,
	"take/stone@outcrop":                Quarry,
	"take/grain@field":                  Farm,
	"tend/clear@open":                   Clear,
	"tend/water@field":                  Irrigate,
	"tend/plant@open":                   PlantTrees,
	"make/timber>tool@bench":            Craft,
	"make/stone+timber>tool@forge":      Smelt,
	"make/provision+timber>meal@hearth": Cook,
	"raise/timber>dwelling@open":        BuildShelter,
	"raise/timber+stone>granary@open":   BuildGranary,
	"raise/timber>tavern@open":          BuildTavern,
	"raise/timber>road@ground":          Pave,
	"consume/provision":                 Eat,
	"dwell/rest":                        Rest,
	"dwell/meet@tavern>neighbour":       Socialize,
	"dwell/look":                        Scout,
	"dwell/guard@market":                Guard,
	"exchange/material>coin@market":     Sell,
	"exchange/coin>provision@market":    Buy,
	"transfer/provision>needy":          Give,
	"transfer/material>requester":       Fulfil,
	"transfer/provision<holder":         Steal,
	"pass/practice>pupil":               Teach,
	"pass/practice>self":                Study,
	"strike/person>wrongdoer":           Retaliate,
	"move@dwelling":                     MoveHouse,
}

func init() {
	for _, in := range ontology.Instantiate() {
		d := mechanics[in.Key]
		if d == nil {
			switch in.Schema.Verb {
			case ontology.Take, ontology.Exchange, ontology.Transfer:
				d = moving(in)
			case ontology.Make:
				d = making(in)
			case ontology.Raise:
				d = raising(in)
			}
		}
		if d == nil {
			panic("action: nothing carries out " + in.Key)
		}
		if d.Key != "" {
			panic("action: " + d.Name + " bound twice, to " + d.Key + " and " + in.Key)
		}
		d.Key = in.Key
		if habit.Register(in.Key) != len(Catalog) {
			panic("action: slot for " + in.Key + " is not its catalog position")
		}
		Catalog = append(Catalog, d)
	}
	Count = len(Catalog)
}

// ByKey returns the action with that key, or nil.
func ByKey(key string) *Def {
	for _, d := range Catalog {
		if d.Key == key {
			return d
		}
	}
	return nil
}

// ByName returns the action with that name, or nil.
func ByName(name string) *Def {
	for _, d := range Catalog {
		if d.Name == name {
			return d
		}
	}
	return nil
}

// Index is the catalog position of d, or -1 if it is not in the catalog.
func Index(d *Def) int {
	for i, c := range Catalog {
		if c == d {
			return i
		}
	}
	return -1
}

// searchRadius bounds how far agents look for a suitable tile.
const searchRadius = 40

func always(*entity.Agent, *world.World) bool { return true }

func here(a *entity.Agent, _ *world.World) (entity.Pos, bool) { return a.Pos, true }

func atMarket(_ *entity.Agent, w *world.World) (entity.Pos, bool) { return w.MarketPos, true }

// foodValue is how much one more unit of food is worth to the physiological
// tier. It falls off as the larder fills, so nobody hoards for its own sake.
func foodValue(a *entity.Agent) float64 {
	return 0.25 * math.Max(0, 1-a.Inventory[entity.Food]/4)
}

// A rest restores a little of the body. A rest that restored more under a
// roof was tried: the reward reinforced an idle act, and under a seasoned
// year the median settlement fell by a third. Rest is a fallback and stays
// one.
const restGain = 0.03

// Rest is the fallback. It is always available and barely worth anything.
var Rest = &Def{
	Name: "rest", Ticks: 1, Available: always, Target: here,
	Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
		return need.Levels{need.Physiological: restGain}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		a.Needs.Add(need.Physiological, restGain)
	},
}

// A meal is as much as it takes to be full, up to a few units, or what there
// is. Yields from the land are fractional, so an agent that could only eat
// whole units would starve with most of a meal in its pack; and an agent
// that could only eat one unit at a sitting would, with a larder full,
// still go hungry between the sittings it finds time for.
const (
	mouthful    = 0.25 // the least worth stopping to eat
	feast       = 3.0  // the most eaten at one sitting
	nourished   = 0.35 // what one unit of raw food restores
	mealNourish = 0.5  // what one cooked meal restores
)

// Edible is everything an agent could eat, raw and cooked.
func Edible(a *entity.Agent) float64 { return a.Inventory[entity.Food] + a.Inventory[entity.Meals] }

// helping is what an agent would eat now: cooked meals first, then raw
// food, enough to be full, within what it has and what a sitting can hold.
// It returns the meals and the raw food taken, and what they restore.
func helping(a *entity.Agent) (meals, raw, restores float64) {
	room := 1 - a.Needs[need.Physiological]
	budget := feast
	meals = min(a.Inventory[entity.Meals], min(budget, room/mealNourish))
	room -= meals * mealNourish
	budget -= meals
	raw = min(a.Inventory[entity.Food], min(budget, max(0, room/nourished)))
	if meals+raw < mouthful {
		// Not hungry enough to be worth the sitting, but there is something:
		// take a mouthful of whichever there is.
		if a.Inventory[entity.Meals] >= mouthful {
			meals, raw = mouthful, 0
		} else {
			meals, raw = 0, min(mouthful, a.Inventory[entity.Food])
		}
	}
	return meals, raw, meals*mealNourish + raw*nourished
}

// A meal eaten in company at the tavern is a little belonging as well.
const tableCheer = 0.03

var Eat = &Def{
	Name: "eat", Ticks: 1, Target: here,
	Available: func(a *entity.Agent, _ *world.World) bool { return Edible(a) >= mouthful },
	Expect: func(a *entity.Agent, w *world.World, target entity.Pos) need.Levels {
		_, _, restores := helping(a)
		gain := need.Levels{need.Physiological: restores}
		if inTavern(w, target) && intended(a, w, target) != nil {
			gain[need.Belonging] = tableCheer
		}
		return gain
	},
	Apply: func(a *entity.Agent, w *world.World) {
		meals, raw, restores := helping(a)
		a.Inventory[entity.Meals] -= meals
		a.Inventory[entity.Food] -= raw
		a.Needs.Add(need.Physiological, restores)
		if inTavern(w, a.Pos) && w.Neighbor(a, 1) != nil {
			a.Needs.Add(need.Belonging, tableCheer)
		}
	},
}

func isForest(_ entity.Pos, t *world.Tile) bool { return t.Terrain == world.Forest }

// forageTake is how much of a forest's wild food one forage consumes.
const forageTake = 0.08

// forageYield is what a forest with this much left gives. A picked forest
// still gives something, so a settlement is pushed toward the river and
// the field rather than into the ground.
func forageYield(wild float64) float64 { return 0.45 + 0.55*wild }

var Forage = mover("take/berries@wood")

// farmWear is the fertility one farming takes from a field, and wornField
// the least a field is worn down to. A field farmed without rest goes poor
// in a few dozen harvests and comes back over a long fallow.
const (
	farmWear  = 0.006
	wornField = 0.1
)

func farmYield(a *entity.Agent, w *world.World, fertility float64) float64 {
	return (0.8 + 2*a.Skills[entity.Farming]) * w.Mods.FarmYield * (0.3 + 0.7*fertility)
}

// fieldTiles is the ground one household works: how many strips a farmer
// goes on breaking before the holding is as much land as the family needs.
//
// A house is one tile. A holding is not, and the gap is not small. A year's
// bread for a family of five is on the order of a tonne of grain. Wheat
// before the plough of our own age gave perhaps a tonne to the hectare in a
// good year, a quarter of which went back into the ground as next year's
// seed, and half the holding lay fallow while the other half bore - so the
// family needed something like three hectares to hold to eat from one. Set
// against the sixty square metres they slept under, the field they lived off
// was hundreds of times the house.
//
// This map cannot carry that ratio: at eighty by thirty-six tiles, three
// hectares to a household would give the world room for a dozen families.
// What it can carry was measured rather than argued. Three strips a
// household is half again as much cultivated land on the map and near twice
// as much per head, for a population two batches of seeds cannot tell from
// one-tile fields; eight buys a quarter more ground again and costs a
// fifteenth of the population and a third of the median. See
// docs/action-space.md.
const fieldTiles = 3

// fieldSoil is the least fertile ground worth breaking. It is what a farmer
// asks of the tile they first clear, and of every strip they add after: a
// holding grows into land that will bear, and stops at the sand.
const fieldSoil = 0.3

// worked is the strip of a holding a farmer goes to: the nearest one. A
// holding is one farm and not eight fields - the household works the whole
// of it in a season and eats the whole of it, so which strip they are
// standing on when the day's work is done does not matter, and making them
// cross their own land to reach the best of it only cost them the walk.
func worked(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	if len(a.Parcel) == 0 {
		return a.Field, a.HasField // a holding of one, from before it was a holding
	}
	best, near := a.Parcel[0], entity.Dist(a.Pos, a.Parcel[0])
	for _, p := range a.Parcel[1:] {
		if d := entity.Dist(a.Pos, p); d < near {
			best, near = p, d
		}
	}
	return best, true
}

// bearing is what a holding has to give: the fertility of the ground taken
// together, because the harvest comes off all of it. A worn strip is carried
// by the rest, which is what a holding large enough to rotate is for.
func bearing(a *entity.Agent, w *world.World) float64 {
	if len(a.Parcel) == 0 {
		return 0
	}
	sum := 0.0
	for _, p := range a.Parcel {
		sum += w.Grid.At(p).Fertility
	}
	return sum / float64(len(a.Parcel))
}

// plough reports whether open ground is worth breaking: soil the crop will
// come up in, and not the yard of somebody's house. A settlement keeps its
// built ground - the roofs, and the gaps between them the lanes run along -
// and the holdings lie outside it, which is where a village puts its fields.
func plough(w *world.World, p entity.Pos) bool {
	t := w.Grid.At(p)
	return t.Buildable() && t.Fertility >= fieldSoil && !w.Grid.HasNeighbor(p, (*world.Tile).Roofed)
}

// newGround is the best ground worth breaking beside the holding: where the
// next strip goes when the family has not yet broken all the land it eats.
// It must be at least as good as what the household already works, because
// the harvest comes off the holding as a whole - taking on poorer ground
// would only pull down the crop the family lives on.
func newGround(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	var best entity.Pos
	found := false
	fertility := bearing(a, w)
	for _, p := range a.Parcel {
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				q := entity.Pos{X: p.X + dx, Y: p.Y + dy}
				if !w.Grid.In(q) || !plough(w, q) {
					continue
				}
				if t := w.Grid.At(q); t.Fertility >= fertility {
					best, fertility, found = q, t.Fertility, true
				}
			}
		}
	}
	return best, found
}

// fieldSite is where a farmer goes to make ground into field: the next
// strip to break if the holding is still short of what the household
// eats, otherwise nothing. Someone with no field at all takes the nearest
// open ground near home that will bear a crop - a man with no land takes
// what he can get, even the strip behind his neighbour's house; it is only
// in adding to a holding that a farmer leaves the neighbourhood its ground.
func fieldSite(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	if a.HasField {
		if len(a.Parcel) < fieldTiles {
			return newGround(a, w)
		}
		return entity.Pos{}, false
	}
	anchor := a.Pos
	if a.HasHome {
		anchor = a.Home
	}
	return w.Grid.Nearest(anchor, searchRadius, func(_ entity.Pos, t *world.Tile) bool {
		return t.Buildable() && t.Fertility >= fieldSoil
	})
}

// breakGround turns open ground beside a holding into another strip of it.
func breakGround(a *entity.Agent, w *world.World) bool {
	t := w.Grid.At(a.Pos)
	if !plough(w, a.Pos) || len(a.Parcel) >= fieldTiles {
		return false
	}
	if !w.Grid.HasNeighbor(a.Pos, func(n *world.Tile) bool { return n.Terrain == world.Field && n.Owner == a.ID }) {
		return false
	}
	t.Terrain, t.Owner = world.Field, a.ID
	a.Parcel = append(a.Parcel, a.Pos)
	return true
}

// Clear is the half of farming that makes ground into a field: the first
// strip of a holding claimed and broken, or the next one added beside it
// while the household still eats more than it holds. Nothing comes off it
// yet. It is its own act so that a holding is a thing an agent has, lacks,
// or is still adding to.
var Clear = &Def{
	Name: "clear field", Ticks: 4, Target: fieldSite,
	Available: func(a *entity.Agent, _ *world.World) bool {
		return !a.HasField || len(a.Parcel) < fieldTiles
	},
	Expect: func(a *entity.Agent, w *world.World, target entity.Pos) need.Levels {
		// Instrumental: a strip is worth what its first harvest will be.
		return need.Levels{
			need.Physiological: foodValue(a) * farmYield(a, w, w.Grid.At(target).Fertility) * 0.5,
			need.Esteem:        0.02,
		}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		t := w.Grid.At(a.Pos)
		if a.HasField {
			if !breakGround(a, w) {
				return
			}
		} else {
			if !t.Buildable() {
				return // claimed by someone else first
			}
			t.Terrain = world.Field
			t.Owner = a.ID
			a.Field, a.HasField = a.Pos, true
			a.Parcel = []entity.Pos{a.Pos}
		}
		a.AddSkill(entity.Farming, 0.005)
		a.Needs.Add(need.Esteem, 0.02)
		w.Emit(event.Built, a.ID, 0, "%s cleared a field", a.Name)
	},
}

// Farm is the harvest: the household's holding, worked and worn.
var Farm = &Def{
	Name: "farm", Ticks: 4,
	Available: func(a *entity.Agent, _ *world.World) bool { return a.HasField },
	Target:    worked,
	Expect: func(a *entity.Agent, w *world.World, _ entity.Pos) need.Levels {
		return need.Levels{
			need.Physiological: foodValue(a) * farmYield(a, w, bearing(a, w)),
			need.Esteem:        0.02,
		}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		if !a.Holds(a.Pos) {
			return // the strip went to someone else while we walked
		}
		yield := farmYield(a, w, bearing(a, w))
		// A tool makes the work go further, and wears with it.
		if a.Inventory[entity.Tools] >= 0.5 {
			yield *= toolFarming
			a.Inventory[entity.Tools] -= farmWearTool
		}
		a.Inventory[entity.Food] += yield
		// One harvest takes one harvest's worth out of the ground however
		// much ground it came off, so the draw is shared over the holding.
		// Eight strips are not eight fields' worth of food; they are one
		// family's, off land that gets a rest between crops.
		wear := farmWear / float64(len(a.Parcel))
		for _, p := range a.Parcel {
			f := w.Grid.At(p)
			f.Fertility = max(wornField, f.Fertility-wear)
		}
		a.AddSkill(entity.Farming, 0.01)
		a.Needs.Add(need.Esteem, 0.02)
	},
}

// A day in the woods and what it is worth. Felling is the slow half of every
// building: treeTake is how much standing timber one day's work brings down,
// and armful how much of that a person can drag home before the light goes.
// Most of a tree is left where it falls. When an armful was three lengths of
// wood a single tree housed a family twice over, everybody was under a roof
// inside the first season, and the forest was decoration rather than the
// thing a settlement is built out of. At half a length the woods are what a
// house is made of, and a settlement's shape follows the treeline.
const (
	treeTake = 0.4
	armful   = 0.5
)

var GatherWood = mover("take/timber@wood")

// Raising a house and keeping one are the same act to the person doing it
// and quite different things to the forest. raisingTimber is the frame: the
// walls and the roof beams, a winter's felling, more wood than anybody
// carries about with them and so a thing that has to be gathered toward on
// purpose over many days. roofingTimber is what patching that roof takes
// afterwards, which is exactly one day in the woods: an armful, carried home
// and nailed on the same evening.
//
// Before they were told apart a house cost what a repair costs, and a
// settlement of twenty was housed to the last person inside two hundred
// ticks, on wood nobody had to go looking for. Charging the frame properly
// while leaving the patching cheap is what puts the founding years back:
// people live rough among half-built walls for a good while, the ones with
// a roof keep it easily, and where the houses go is decided by where the
// timber was rather than by where the first day's walk happened to end.
const (
	raisingTimber = 6
	roofingTimber = armful
)

// timberToBuild is what the next day's building costs this agent: a frame if
// they have no house, a course of repair if they have.
func timberToBuild(a *entity.Agent) float64 {
	if a.HasHome {
		return roofingTimber
	}
	return raisingTimber
}

func shelterGain(a *entity.Agent, w *world.World) float64 {
	return math.Min(1-a.Shelter, 0.4*w.Mods.BuildEfficiency*(0.5+a.Skills[entity.Building]))
}

// plotNear is the closest place to anchor where a house can stand with its
// own ground around it. A settlement that builds wall to wall has nowhere
// left to put a street, so a plot is looked for first and open ground only
// taken as it comes when the neighbourhood has run out of room.
func plotNear(w *world.World, anchor entity.Pos) (entity.Pos, bool) {
	if p, ok := w.Grid.Nearest(anchor, searchRadius, func(p entity.Pos, _ *world.Tile) bool {
		return w.Grid.RoomToBuild(p)
	}); ok {
		return p, true
	}
	return w.Grid.Nearest(anchor, searchRadius, func(_ entity.Pos, t *world.Tile) bool { return t.Buildable() })
}

// buildSite is the agent's house, or the best plot it knows of.
//
// It used to be a plot near what the agent cared about, and what it was taken
// to care about was the market - a position nobody in the settlement had
// chosen and every one of them was measured from. Since the anchor was shared
// and the search was deterministic, every homeless agent alive was handed the
// same tile on the same tick and queued for it. Siting now runs off what the
// agent has personally walked over, which no two of them have the same list
// of, and off what that ground is actually worth. See ground.go.
func buildSite(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	if a.HasHome {
		return a.Home, true
	}
	return KnownPlot(a, w)
}

// roomNearby reports whether a plot with its own ground around it is still
// to be had within reach of p.
func roomNearby(w *world.World, p entity.Pos) bool {
	_, ok := w.Grid.Nearest(p, searchRadius, func(q entity.Pos, _ *world.Tile) bool {
		return w.Grid.RoomToBuild(q)
	})
	return ok
}

var BuildShelter = &Def{
	Name: "build shelter", Ticks: 3, Target: buildSite,
	Available: func(a *entity.Agent, _ *world.World) bool {
		return a.Inventory[entity.Wood] >= timberToBuild(a) && a.Shelter < 0.95
	},
	Expect: func(a *entity.Agent, w *world.World, _ entity.Pos) need.Levels {
		return need.Levels{need.Safety: shelterGain(a, w) * 0.8, need.Esteem: 0.05}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		// The frame is charged for here rather than up front, because a
		// raising that finds the plot taken is a wasted walk and not a
		// wasted winter's timber.
		cost := timberToBuild(a)
		if a.Inventory[entity.Wood] < cost {
			return
		}
		if !a.HasHome {
			t := w.Grid.At(a.Pos)
			if !t.Buildable() {
				return
			}
			// Somebody may have built next door during the walk over. Go
			// looking again rather than raise a wall against theirs, unless
			// there is no plot left within reach to go looking for.
			if !w.Grid.RoomToBuild(a.Pos) && roomNearby(w, a.Pos) {
				return
			}
			t.Structure = world.House
			t.Owner = a.ID
			a.Home, a.HasHome = a.Pos, true
			w.Emit(event.Built, a.ID, 0, "%s built a house", a.Name)
		}
		a.Inventory[entity.Wood] -= cost
		gain := shelterGain(a, w)
		// A stone in the walls makes a house that stands.
		if a.Inventory[entity.Stone] >= 1 {
			a.Inventory[entity.Stone]--
			gain = min(1-a.Shelter, gain*stoneHouse)
		}
		a.Shelter = need.Clamp(a.Shelter + gain)
		a.AddSkill(entity.Building, 0.02)
		a.Needs.Add(need.Esteem, 0.05)
	},
}

// pavingWood is the timber one length of road takes: two days in the woods,
// a small fraction of a house, so laying a way is a far smaller commitment
// than raising a roof while competing with it for the same wood. A bridge
// takes twice that, because it has to hold itself up over the water, and it
// is the one piece of road worth walking a long way to build: a river is
// otherwise something a settlement can only put up with.
//
// These were once set against a house that cost two lengths of timber, when
// an agent spent its wood the moment it had any and nobody ever held more
// than about two and a half lengths at once. A frame now costs six and gets
// saved up for, so both are cheap against it on purpose: the roads are what
// a settled person does with the wood left over, not what they choose
// instead of a roof.
const (
	pavingWood = 1
	bridgeWood = 2
)

// timberFor is what a length of road costs on this ground.
func timberFor(t *world.Tile) float64 {
	if t.Terrain == world.Water {
		return bridgeWood
	}
	return pavingWood
}

// wornEnough is how beaten the ground must be before anyone thinks of paving
// it. Below this the wear is somebody having passed once, not a route.
const wornEnough = 60

// pavingRadius is how far somebody will go to lay a road. Roads are laid
// where the layer already lives and walks, not wherever the settlement's
// worst bottleneck happens to be: nobody has that view of the place.
const pavingRadius = 12

// paveSite is the most walked-on unpaved ground near the agent. Roads follow
// the settlement's own errands: the way people already take is the way that
// gets made, which is why nobody has to plan the network for it to appear.
func paveSite(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	afford := func(t *world.Tile) bool { return a.Inventory[entity.Wood] >= timberFor(t) }
	// A crossing comes before a street. A street can go round whatever is in
	// its way, so the busiest ground will be paved sooner or later whoever
	// gets to it; a river is the one thing a road cannot go round, and the
	// ford is never the busiest ground in a settlement because everybody who
	// can avoid it does. Left to compete on wear alone a bridge is never
	// built, and the two banks stay two settlements.
	ford := func(t *world.Tile) bool { return t.Terrain == world.Water && afford(t) }
	if p, worn, ok := w.Grid.Busiest(a.Pos, pavingRadius, ford); ok && worn >= wornEnough {
		return p, true
	}
	p, worn, ok := w.Grid.Busiest(a.Pos, pavingRadius, afford)
	if !ok || worn < wornEnough {
		return entity.Pos{}, false
	}
	return p, true
}

// Pave is the settlement's first work on the common ground: a stretch of road
// that does the layer no direct good beyond the credit of having laid it, and
// that everybody who walks it afterwards is quicker and less worn for. Like
// standing guard it is a public good, and like standing guard it pays in
// standing rather than in bread, which is the only reason anybody learns to
// keep doing it.
var Pave = &Def{
	Name: "lay road", Ticks: 2, Target: paveSite,
	Available: func(a *entity.Agent, _ *world.World) bool {
		return a.Inventory[entity.Wood] >= pavingWood
	},
	Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
		// The road is for everyone who walks it. What comes back to the one
		// who laid it is the credit of having laid it, and that is deliberately
		// less per tick than standing guard pays: paving that repaid its own
		// effort would be an esteem farm, and the settlement would pave itself
		// into a yard. Measured at twice this, agents laid a third of the map.
		return need.Levels{need.Esteem: 0.03, need.Belonging: 0.02}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		// Somebody may have built here, or paved it, while this one walked,
		// and the timber may have gone on something else on the way.
		cost := timberFor(w.Grid.At(a.Pos))
		if a.Inventory[entity.Wood] < cost || !w.Grid.Pave(a.Pos) {
			return
		}
		a.Inventory[entity.Wood] -= cost
		a.AddSkill(entity.Building, 0.01)
		a.Needs.Add(need.Esteem, 0.03)
		a.Needs.Add(need.Belonging, 0.02)
		what := "laid a road"
		if w.Grid.At(a.Pos).Bridged() {
			what = "bridged the river"
		}
		w.Emit(event.Built, a.ID, 0, "%s %s", a.Name, what)
	},
}

var Sell = mover("exchange/material>coin@market")

var Buy = mover("exchange/coin>provision@market")

var Guard = &Def{
	Name: "guard", Ticks: 3, Available: worthGuarding, Target: atMarket,
	Expect: func(a *entity.Agent, w *world.World, _ entity.Pos) need.Levels {
		return need.Levels{
			need.Safety:    0.12 * (1 - w.Safety),
			need.Belonging: 0.03,
			need.Esteem:    0.04,
		}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		w.Safety = need.Clamp(w.Safety + 0.08)
		a.Needs.Add(need.Belonging, 0.03)
		a.Needs.Add(need.Esteem, 0.04)
		w.Emit(event.Guarded, a.ID, 0, "%s stood guard", a.Name)
	},
}

// companionRadius is how close two agents must be to interact.
const companionRadius = 3

// intended is the person an agent expects to find at a target.
func intended(a *entity.Agent, w *world.World, target entity.Pos) *entity.Agent {
	return w.AgentAt(target, companionRadius, a)
}

var Socialize = &Def{
	Name: "socialize", Ticks: 2, Available: hasCompany, Target: towardCompany,
	Expect: func(a *entity.Agent, w *world.World, target entity.Pos) need.Levels {
		o := intended(a, w, target)
		if o == nil {
			return need.Levels{need.Belonging: 0.08}
		}
		gain := need.Levels{need.Belonging: encounterGain(Anticipate(a, o), a.Temperament)}
		// Being seen with someone of standing reflects on you.
		if o.Reputation > a.Reputation {
			gain[need.Esteem] = 0.03 * math.Min(1, o.Reputation-a.Reputation)
		}
		return gain
	},
	Apply: func(a *entity.Agent, w *world.World) {
		o := w.Neighbor(a, companionRadius)
		if o == nil {
			return // nobody home; a wasted walk
		}
		Encounter(a, o, w)
		// A tavern is a better evening than a doorstep.
		if inTavern(w, a.Pos) {
			a.Needs.Add(need.Belonging, tavernCheer)
			o.Needs.Add(need.Belonging, tavernCheer)
		}
	},
}

func craftQuality(a *entity.Agent, w *world.World) float64 {
	return (0.3 + a.Skills[entity.Crafting]) * w.Mods.CraftQuality
}

var Craft = product("make/timber>tool@bench")

var Teach = &Def{
	Name: "teach", Ticks: 3, Target: towardCompany,
	Available: func(a *entity.Agent, w *world.World) bool {
		_, level := a.BestSkill()
		return level >= 0.4 && hasCompany(a, w)
	},
	Expect: func(a *entity.Agent, w *world.World, target entity.Pos) need.Levels {
		gain := need.Levels{need.Esteem: 0.2, need.Belonging: 0.05}
		if o := intended(a, w, target); o != nil {
			// A willing student is a pleasure; a hostile one is a chore.
			gain[need.Belonging] += 0.1 * math.Max(0, Anticipate(a, o))
		}
		return gain
	},
	Apply: func(a *entity.Agent, w *world.World) {
		o := w.Neighbor(a, companionRadius)
		if o == nil {
			return
		}
		Introduce(a, o, w.Tick)
		Introduce(o, a, w.Tick)
		skill, level := a.BestSkill()
		o.AddSkill(skill, 0.04)
		Pass(a, o, skill)
		// The student now knows what the teacher can do, and thinks a little
		// better of them for the trouble taken.
		o.Rate(a.ID, skill, level, w.Tick)
		o.Judge(a.ID, 0.05, w.Tick)
		o.AddBond(a.ID, 0.04)
		a.Reputation += 0.05
		a.Needs.Add(need.Esteem, 0.2)
		a.Needs.Add(need.Belonging, 0.1)
		o.Needs.Add(need.Belonging, 0.05)
		w.Emit(event.Taught, a.ID, o.ID, "%s taught %s %s", a.Name, o.Name, skill)
	},
}

var Study = &Def{
	Name: "study", Ticks: 4, Available: hasPlace(desk), Target: desk,
	Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
		return need.Levels{need.Actualization: 0.3, need.Esteem: 0.03}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		w.Knowledge += (0.2 + a.Skills[entity.Scholarship]) * w.Mods.StudyRate
		a.AddSkill(entity.Scholarship, 0.02)
		Broaden(a, w)
		a.Needs.Add(need.Actualization, 0.3)
		a.Needs.Add(need.Esteem, 0.03)
	},
}

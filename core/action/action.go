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
	"lreat/core/world"
)

// Def describes one action.
type Def struct {
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
	// choice. See package habit.
	Prior habit.Signature
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

// Count is the size of the catalog. It is checked at init so that a table
// indexed by catalog position can be a fixed array everywhere.
const Count = 17

// Catalog lists every action in a fixed order. Order matters for
// determinism, and position is what per-agent habit tables are indexed by.
// It is assembled in init rather than declared, because some actions reach
// back into the catalog when they run (study broadens reach, teaching
// passes it on) and a declaration would make that a cycle.
var Catalog []*Def

func init() {
	Catalog = []*Def{
		Rest, Eat, Forage, Farm, GatherWood, BuildShelter, Sell, Buy,
		Guard, Socialize, Craft, Teach, Study,
		Steal, Give, Fulfil, Retaliate,
	}
	if len(Catalog) != Count {
		panic("action: Catalog length does not match Count")
	}
	if Count > habit.MaxActions {
		panic("action: Catalog exceeds habit.MaxActions")
	}
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

func hasCompany(_ *entity.Agent, w *world.World) bool { return len(w.Agents) > 1 }

func here(a *entity.Agent, _ *world.World) (entity.Pos, bool) { return a.Pos, true }

func atHome(a *entity.Agent, _ *world.World) (entity.Pos, bool) {
	if a.HasHome {
		return a.Home, true
	}
	return a.Pos, true
}

func atMarket(_ *entity.Agent, w *world.World) (entity.Pos, bool) { return w.MarketPos, true }

// foodValue is how much one more unit of food is worth to the physiological
// tier. It falls off as the larder fills, so nobody hoards for its own sake.
func foodValue(a *entity.Agent) float64 {
	return 0.25 * math.Max(0, 1-a.Inventory[entity.Food]/4)
}

// Rest is the fallback. It is always available and barely worth anything.
var Rest = &Def{
	Name: "rest", Ticks: 1, Available: always, Target: here,
	Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
		return need.Levels{need.Physiological: 0.03}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		a.Needs.Add(need.Physiological, 0.03)
	},
}

var Eat = &Def{
	Name: "eat", Ticks: 1, Target: here,
	Available: func(a *entity.Agent, _ *world.World) bool { return a.Inventory[entity.Food] >= 1 },
	Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
		return need.Levels{need.Physiological: 0.35}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		a.Inventory[entity.Food]--
		a.Needs.Add(need.Physiological, 0.35)
	},
}

func isForest(_ entity.Pos, t *world.Tile) bool { return t.Terrain == world.Forest }

var Forage = &Def{
	Name: "forage", Ticks: 2, Available: always,
	Target: func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
		return w.Grid.Nearest(a.Pos, searchRadius, isForest)
	},
	Expect: func(a *entity.Agent, _ *world.World, _ entity.Pos) need.Levels {
		return need.Levels{need.Physiological: foodValue(a) * 1.0}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		if w.Grid.At(a.Pos).Terrain != world.Forest {
			return // somebody cleared it while we walked
		}
		a.Inventory[entity.Food] += 1.0 * (0.6 + 0.8*w.RNG.Float64())
	},
}

func farmYield(a *entity.Agent, w *world.World, fertility float64) float64 {
	return (0.8 + 2*a.Skills[entity.Farming]) * w.Mods.FarmYield * (0.3 + 0.7*fertility)
}

// farmSite is the agent's field, or the best unclaimed ground near home.
func farmSite(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	if a.HasField {
		return a.Field, true
	}
	anchor := a.Pos
	if a.HasHome {
		anchor = a.Home
	}
	return w.Grid.Nearest(anchor, searchRadius, func(_ entity.Pos, t *world.Tile) bool {
		return t.Buildable() && t.Fertility >= 0.3
	})
}

var Farm = &Def{
	Name: "farm", Ticks: 4, Available: always, Target: farmSite,
	Expect: func(a *entity.Agent, w *world.World, target entity.Pos) need.Levels {
		return need.Levels{
			need.Physiological: foodValue(a) * farmYield(a, w, w.Grid.At(target).Fertility),
			need.Esteem:        0.02,
		}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		t := w.Grid.At(a.Pos)
		if !a.HasField {
			if !t.Buildable() {
				return // claimed by someone else first
			}
			t.Terrain = world.Field
			t.Owner = a.ID
			a.Field, a.HasField = a.Pos, true
			w.Emit(event.Built, a.ID, 0, "%s cleared a field", a.Name)
		}
		if a.Pos != a.Field {
			return
		}
		a.Inventory[entity.Food] += farmYield(a, w, t.Fertility)
		a.AddSkill(entity.Farming, 0.01)
		a.Needs.Add(need.Esteem, 0.02)
	},
}

var GatherWood = &Def{
	Name: "gather wood", Ticks: 2, Available: always,
	Target: func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
		return w.Grid.Nearest(a.Pos, searchRadius, func(_ entity.Pos, t *world.Tile) bool {
			return t.Terrain == world.Forest && t.Wood >= 0.3
		})
	},
	Expect: func(a *entity.Agent, _ *world.World, _ entity.Pos) need.Levels {
		// Instrumental: wood is only worth something if you lack shelter or craft.
		want := 0.0
		if a.Inventory[entity.Wood] < 2 {
			want = 0.1*(1-a.Shelter) + 0.03*a.Skills[entity.Crafting]
		}
		return need.Levels{need.Safety: want, need.Esteem: want * 0.3}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		t := w.Grid.At(a.Pos)
		if t.Terrain != world.Forest {
			return
		}
		t.Wood -= 0.4
		a.Inventory[entity.Wood] += 1.5
		if t.Wood < 0.1 {
			t.Terrain, t.Wood = world.Grass, 0
		}
	},
}

func shelterGain(a *entity.Agent, w *world.World) float64 {
	return math.Min(1-a.Shelter, 0.4*w.Mods.BuildEfficiency*(0.5+a.Skills[entity.Building]))
}

// buildSite is the agent's house, or open ground near what they care about:
// the market for the safety-minded, their field for everyone else.
func buildSite(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	if a.HasHome {
		return a.Home, true
	}
	anchor := w.MarketPos
	if a.HasField && a.Personality[need.Safety] < 1 {
		anchor = a.Field
	}
	return w.Grid.Nearest(anchor, searchRadius, func(_ entity.Pos, t *world.Tile) bool { return t.Buildable() })
}

var BuildShelter = &Def{
	Name: "build shelter", Ticks: 3, Target: buildSite,
	Available: func(a *entity.Agent, _ *world.World) bool {
		return a.Inventory[entity.Wood] >= 2 && a.Shelter < 0.95
	},
	Expect: func(a *entity.Agent, w *world.World, _ entity.Pos) need.Levels {
		return need.Levels{need.Safety: shelterGain(a, w) * 0.8, need.Esteem: 0.05}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		if !a.HasHome {
			t := w.Grid.At(a.Pos)
			if !t.Buildable() {
				return
			}
			t.Structure = world.House
			t.Owner = a.ID
			a.Home, a.HasHome = a.Pos, true
			w.Emit(event.Built, a.ID, 0, "%s built a house", a.Name)
		}
		a.Inventory[entity.Wood] -= 2
		a.Shelter = need.Clamp(a.Shelter + shelterGain(a, w))
		a.AddSkill(entity.Building, 0.02)
		a.Needs.Add(need.Esteem, 0.05)
	},
}

var Sell = &Def{
	Name: "sell", Ticks: 1, Target: atMarket,
	Available: func(a *entity.Agent, _ *world.World) bool {
		return a.Inventory[entity.Food] > 4 || a.Inventory[entity.Tools] >= 1
	},
	Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
		// Savings buy safety; being a seller of note buys a little esteem.
		return need.Levels{need.Safety: 0.08, need.Esteem: 0.03}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		var earned float64
		if surplus := a.Inventory[entity.Food] - 3; surplus > 0 {
			a.Inventory[entity.Food] -= surplus
			w.Market.Stock[entity.Food] += surplus
			earned += surplus * w.Market.Price[entity.Food]
		}
		if tools := a.Inventory[entity.Tools]; tools > 0 {
			a.Inventory[entity.Tools] = 0
			w.Market.Stock[entity.Tools] += tools
			earned += tools * w.Market.Price[entity.Tools]
		}
		a.Wealth += earned
		a.Needs.Add(need.Esteem, 0.03)
		w.Emit(event.Traded, a.ID, 0, "%s sold goods for %.1f", a.Name, earned)
	},
}

var Buy = &Def{
	Name: "buy food", Ticks: 1, Target: atMarket,
	Available: func(a *entity.Agent, w *world.World) bool {
		return a.Inventory[entity.Food] < 1 &&
			w.Market.Stock[entity.Food] >= 1 &&
			a.Wealth >= w.Market.Price[entity.Food]
	},
	Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
		return need.Levels{need.Physiological: 0.3}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		p := w.Market.Price[entity.Food]
		a.Wealth -= p
		a.Inventory[entity.Food]++
		w.Market.Stock[entity.Food]--
		w.Emit(event.Traded, a.ID, 0, "%s bought food for %.2f", a.Name, p)
	},
}

var Guard = &Def{
	Name: "guard", Ticks: 3, Available: hasCompany, Target: atMarket,
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

// towardCompany heads for the person the agent would most like to see. They
// may have moved by the time we arrive; then whoever is nearby will do.
func towardCompany(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	o := PickCompany(a, w)
	if o == nil {
		return entity.Pos{}, false
	}
	return o.Pos, true
}

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
	},
}

func craftQuality(a *entity.Agent, w *world.World) float64 {
	return (0.3 + a.Skills[entity.Crafting]) * w.Mods.CraftQuality
}

var Craft = &Def{
	Name: "craft", Ticks: 3, Target: atHome,
	Available: func(a *entity.Agent, _ *world.World) bool { return a.Inventory[entity.Wood] >= 1 },
	Expect: func(a *entity.Agent, w *world.World, _ entity.Pos) need.Levels {
		q := craftQuality(a, w)
		return need.Levels{need.Esteem: 0.15 * q, need.Safety: 0.03 * q}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		q := craftQuality(a, w)
		a.Inventory[entity.Wood]--
		a.Inventory[entity.Tools] += q
		a.Reputation += 0.05 * q
		a.AddSkill(entity.Crafting, 0.015)
		a.Needs.Add(need.Esteem, 0.15*q)
	},
}

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
	Name: "study", Ticks: 4, Available: always,
	Target: func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
		if a.HasHome {
			return a.Home, true
		}
		return w.MarketPos, true
	},
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

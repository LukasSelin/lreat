package action

import (
	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// Taking is one verb. Foraging, hunting, felling, fishing, and cutting
// stone are all the same act: go to where a thing lies, bring some of it
// in, leave the ground a little poorer for it. What differs between them
// is where the thing lies and what bringing it in costs, and that is
// stated here as data rather than as five copies of the act. The
// ontology says which materials lie where; a lode says what a material
// is like to take, and a ground what a site is like to stand on. An act
// the ontology entails for a material with a lode, at a site with a
// ground, needs no code of its own.

// lode is how a material lies in the land and what taking it is like.
type lode struct {
	// Name is what the taking is called.
	Name string
	// Good is what it brings in.
	Good entity.Good
	// Stock is the tile's store the taking draws down, or nil where the
	// land has no end of it.
	Stock func(t *world.Tile) *float64
	// Least is the stock below which a site is not worth walking to, and
	// Spent the stock at or below which there is nothing left to take.
	// Negative means the land always gives something, however picked.
	Least, Spent float64
	// Take is how much stock one taking removes.
	Take float64
	// Yield is what one taking brings in, given the stock it draws on.
	Yield func(a *entity.Agent, w *world.World, stock float64) float64
	// Luck spreads the yield by (lo + span*chance), so that the land gives
	// unevenly. A zero span is a sure thing.
	Luck struct{ Lo, Span float64 }
	// Tool is how much of a tool the taking wears; zero means it needs
	// none.
	Tool float64
	// Skill is what the taking exercises, by Learn per taking.
	Skill   entity.Skill
	Skilled bool
	Learn   float64
	// Pride is the esteem a taking brings.
	Pride float64
	// Worth is what a taking with this yield is worth to the agent, by
	// need tier.
	Worth func(a *entity.Agent, w *world.World, yield float64) need.Levels
	// Gone is what becomes of the tile once the taking has drawn on it.
	Gone func(t *world.Tile)
}

// ground is what a site is like to take from: how to know a tile of it,
// and which tile the taking draws on when standing there. Fishing stands
// on the bank and draws on the water beside it.
type ground struct {
	Here  func(w *world.World, p entity.Pos, t *world.Tile) bool
	Drawn func(w *world.World, p entity.Pos) *world.Tile
}

func at(w *world.World, p entity.Pos) *world.Tile { return w.Grid.At(p) }

var grounds = map[*ontology.Class]ground{
	ontology.Wood:    {Here: func(_ *world.World, p entity.Pos, t *world.Tile) bool { return isForest(p, t) }, Drawn: at},
	ontology.Outcrop: {Here: func(_ *world.World, p entity.Pos, t *world.Tile) bool { return isRock(p, t) }, Drawn: at},
	ontology.Water:   {Here: func(w *world.World, p entity.Pos, t *world.Tile) bool { return bank(w)(p, t) }, Drawn: bestWater},
}

func wildOf(t *world.Tile) *float64 { return &t.Wild }
func woodOf(t *world.Tile) *float64 { return &t.Wood }
func fishOf(t *world.Tile) *float64 { return &t.Fish }

// fed is what a taking of food is worth: a meal, less the more one has.
func fed(a *entity.Agent, _ *world.World, yield float64) need.Levels {
	return need.Levels{need.Physiological: foodValue(a) * yield}
}

var lodes = map[*ontology.Class]lode{
	// The nearest forest, thin as it may be. Going further for a fuller
	// patch was tried: the walk each way, on the errand the whole economy
	// runs on, cost more than the fuller patch gave, and a forager who
	// stays put and finds less is what turns a settlement toward the river
	// and the field.
	ontology.Berries: {
		Name: "forage", Good: entity.Food, Stock: wildOf, Take: forageTake, Spent: -1,
		Yield: func(_ *entity.Agent, _ *world.World, wild float64) float64 { return forageYield(wild) },
		Luck:  struct{ Lo, Span float64 }{0.6, 0.8}, Worth: fed,
	},
	ontology.Game: {
		Name: "hunt", Good: entity.Food, Stock: wildOf, Least: 0.3, Spent: 0.1, Take: huntTake,
		Yield: func(_ *entity.Agent, w *world.World, wild float64) float64 { return huntYield(w, wild) },
		Luck:  struct{ Lo, Span float64 }{0.5, 1.0}, Tool: toolWear, Worth: fed,
	},
	ontology.Fish: {
		Name: "fish", Good: entity.Food, Stock: fishOf, Take: fishTake,
		Yield: func(a *entity.Agent, w *world.World, fish float64) float64 { return fishYield(a, w, fish) },
		Luck:  struct{ Lo, Span float64 }{0.5, 1.0}, Skill: entity.Fishing, Skilled: true, Learn: 0.015, Worth: fed,
	},
	ontology.Timber: {
		Name: "gather wood", Good: entity.Wood, Stock: woodOf, Least: 0.3, Spent: -1, Take: treeTake,
		Yield: func(*entity.Agent, *world.World, float64) float64 { return armful },
		Luck:  struct{ Lo, Span float64 }{1, 0},
		// Instrumental: wood is only worth something if you lack shelter
		// or craft.
		Worth: func(a *entity.Agent, _ *world.World, _ float64) need.Levels {
			want := 0.0
			if a.Inventory[entity.Wood] < raisingTimber {
				want = 0.1*(1-a.Shelter) + 0.03*a.Skills[entity.Crafting]
			}
			return need.Levels{need.Safety: want, need.Esteem: want * 0.3}
		},
		// A wood felled to the last stand is a clearing.
		Gone: func(t *world.Tile) {
			if t.Wood < 0.1 {
				t.Terrain, t.Wood = world.Grass, 0
			}
		},
	},
	ontology.Stone: {
		Name: "quarry", Good: entity.Stone,
		Yield: func(a *entity.Agent, _ *world.World, _ float64) float64 { return quarryYield(a) },
		Luck:  struct{ Lo, Span float64 }{1, 0}, Tool: quarryWear,
		Skill: entity.Building, Skilled: true, Learn: 0.01, Pride: 0.03,
		// Stone is for building; it is worth the safety of the house it
		// will go into, if the agent lacks one.
		Worth: func(a *entity.Agent, _ *world.World, yield float64) need.Levels {
			return need.Levels{need.Safety: 0.06 * (1 - a.Shelter) * yield, need.Esteem: 0.03}
		},
	},
}

// stock is what a tile holds of a lode, or 1 where the land has no end of it.
func (l *lode) stock(t *world.Tile) float64 {
	if l.Stock == nil {
		return 1
	}
	return *l.Stock(t)
}

// taking is the act the ontology entails for taking a material at a site,
// carried out from its lode and ground, or nil where either is unknown.
func taking(in ontology.Instance) *Def {
	l, ok := lodes[in.Object]
	g, ok2 := grounds[in.Site]
	if !ok || !ok2 {
		return nil
	}
	tech := world.Tech(in.Tech)
	d := &Def{Name: l.Name, Ticks: in.Ticks}
	d.Available = func(a *entity.Agent, w *world.World) bool {
		if l.Tool > 0 && a.Inventory[entity.Tools] < 0.5 {
			return false
		}
		return tech == "" || known(a, w, tech, l.Name)
	}
	d.Target = func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
		return w.Grid.Nearest(a.Pos, searchRadius, func(p entity.Pos, t *world.Tile) bool {
			if !g.Here(w, p, t) {
				return false
			}
			drawn := g.Drawn(w, p)
			return drawn != nil && l.stock(drawn) >= l.Least
		})
	}
	d.Expect = func(a *entity.Agent, w *world.World, target entity.Pos) need.Levels {
		t := g.Drawn(w, target)
		if t == nil {
			return need.Levels{}
		}
		return l.Worth(a, w, l.Yield(a, w, l.stock(t)))
	}
	d.Apply = func(a *entity.Agent, w *world.World) {
		t := g.Drawn(w, a.Pos)
		if t == nil || !g.Here(w, a.Pos, w.Grid.At(a.Pos)) || l.stock(t) <= l.Spent {
			return // taken, cleared, or fished out while we walked
		}
		luck := l.Luck.Lo
		if l.Luck.Span > 0 {
			luck += l.Luck.Span * w.RNG.Float64()
		}
		a.Inventory[l.Good] += l.Yield(a, w, l.stock(t)) * luck
		if l.Stock != nil {
			s := l.Stock(t)
			*s = max(0, *s-l.Take)
		}
		if l.Tool > 0 {
			a.Inventory[entity.Tools] = max(0, a.Inventory[entity.Tools]-l.Tool)
		}
		if l.Skilled {
			a.AddSkill(l.Skill, l.Learn)
		}
		if l.Pride > 0 {
			a.Needs.Add(need.Esteem, l.Pride)
		}
		if l.Gone != nil {
			l.Gone(t)
		}
	}
	return d
}

// take is the taking the ontology entails under key, for the acts the rest
// of the package refers to by name.
func take(key string) *Def {
	d := taking(instances[key])
	if d == nil {
		panic("action: no lode or ground for " + key)
	}
	return d
}

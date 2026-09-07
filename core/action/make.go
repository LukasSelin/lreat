package action

import (
	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// Making is one verb. Crafting a tool, smelting one, and cooking a meal are
// the same act: at a workplace, turn what is on hand into something that
// is worth more. The ontology says what goes into what and at what kind
// of place; a recipe says how much, what comes out, and what the making
// is worth. An act the ontology entails for a schema with a recipe needs
// no code of its own.

// workplaces is where an act placed by trait is done. A dwelling has all
// of them; the market lends a bench and a desk, the tavern a hearth.
var workplaces = map[ontology.Trait]func(*entity.Agent, *world.World) (entity.Pos, bool){
	ontology.Bench:  bench,
	ontology.Hearth: hearth,
	ontology.Forge:  forge,
	ontology.Desk:   desk,
}

// recipe is how much a making takes of each of its inputs, in the order
// the schema lists them, and what the making is like.
type recipe struct {
	Name string
	// Amounts is what the making consumes of each input.
	Amounts []float64
	// Reserve is what the maker keeps back of a good beyond what the
	// making takes: fuel comes out of what is left after a house's worth.
	Reserve struct {
		Good   entity.Good
		Amount float64
	}
	// Yield is how much of the output one making gives.
	Yield func(a *entity.Agent, w *world.World) float64
	// Skill is what the making exercises, by Learn per making.
	Skill entity.Skill
	Learn float64
	// Renown is the standing a making earns, per unit of yield.
	Renown float64
	// Worth is what a making with this yield is worth, by need tier. The
	// esteem it promises is the esteem it gives.
	Worth func(a *entity.Agent, w *world.World, yield float64) need.Levels
}

// made is what a tool is worth: the standing of having made it, and a
// little of the safety it will earn.
func made(_ *entity.Agent, _ *world.World, q float64) need.Levels {
	return need.Levels{need.Esteem: 0.15 * q, need.Safety: 0.03 * q}
}

var recipes = map[string]recipe{
	"make/timber>tool@bench": {
		Name: "craft", Amounts: []float64{1},
		Yield: craftQuality, Skill: entity.Crafting, Learn: 0.015, Renown: 0.05, Worth: made,
	},
	"make/stone+timber>tool@forge": {
		Name: "smelt", Amounts: []float64{smeltStone, smeltWood},
		Yield: func(a *entity.Agent, w *world.World) float64 { return craftQuality(a, w) * smeltYield },
		Skill: entity.Crafting, Learn: 0.02, Renown: 0.08, Worth: made,
	},
	// A cooking is a batch: cookBatch units of food over cookFuel of wood
	// make cookBatch meals, out of the wood left after a house's worth.
	"make/provision+timber>meal@hearth": {
		Name: "cook", Amounts: []float64{cookBatch, cookFuel},
		Reserve: struct {
			Good   entity.Good
			Amount float64
		}{entity.Wood, cookReserve},
		Yield: func(*entity.Agent, *world.World) float64 { return cookBatch },
		Skill: entity.Crafting, Learn: 0.005,
		// A cooked meal feeds more than a raw one; that difference is what
		// cooking is worth today.
		Worth: func(a *entity.Agent, _ *world.World, batch float64) need.Levels {
			return need.Levels{need.Physiological: (mealNourish - nourished) * batch * foodValue(a) * 4, need.Esteem: 0.02}
		},
	},
}

// making is the act the ontology entails for a schema with a recipe,
// carried out from it, or nil where there is none.
func making(in ontology.Instance) *Def {
	r, ok := recipes[in.Key]
	place, ok2 := workplaces[in.Schema.SiteTrait]
	_, ok3 := world.GoodOf(in.Schema.Output)
	if !ok || !ok2 || !ok3 || len(r.Amounts) != len(in.Schema.Inputs) {
		return nil
	}
	for _, c := range in.Schema.Inputs {
		if _, ok := world.GoodOf(c); !ok {
			return nil
		}
	}
	takes, gives := in.Schema.Inputs, in.Schema.Output
	tech := world.Tech(in.Tech)
	d := &Def{Name: r.Name, Ticks: in.Ticks, Target: place}
	d.Available = func(a *entity.Agent, w *world.World) bool {
		for i, m := range takes {
			wants := r.Amounts[i]
			if g, _ := world.GoodOf(m); g == r.Reserve.Good {
				wants += r.Reserve.Amount
			}
			if mine, _ := pack(a, m); mine.Held() < wants {
				return false
			}
		}
		if _, ok := place(a, w); !ok {
			return false
		}
		return tech == "" || known(a, w, tech, r.Name)
	}
	d.Expect = func(a *entity.Agent, w *world.World, _ entity.Pos) need.Levels {
		return r.Worth(a, w, r.Yield(a, w))
	}
	d.Apply = func(a *entity.Agent, w *world.World) {
		yield := r.Yield(a, w)
		for i, m := range takes {
			mine, _ := pack(a, m)
			mine.Move(-r.Amounts[i])
		}
		made, _ := pack(a, gives)
		made.Move(yield)
		a.Reputation += r.Renown * yield
		a.AddSkill(r.Skill, r.Learn)
		a.Needs.Add(need.Esteem, r.Worth(a, w, yield)[need.Esteem])
	}
	return d
}

// product is the making the ontology entails under key, for the acts the
// rest of the package refers to by name.
func product(key string) *Def {
	d := making(instances[key])
	if d == nil {
		panic("action: no recipe for " + key)
	}
	return d
}

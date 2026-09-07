package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// A making is composed from a recipe over a schema, not written: what goes
// in and where comes from the ontology, how much and what for from the
// recipe. Stone alone makes no tool in the trees, so this act is never
// entailed; that it can be carried out all the same is the point.
func TestAMakingIsComposedFromARecipe(t *testing.T) {
	w, a := workshop(t)
	key := "make/stone>tool@bench"
	recipes[key] = recipe{
		Name: "knap", Amounts: []float64{2},
		Yield: func(*entity.Agent, *world.World) float64 { return 0.5 },
		Skill: entity.Crafting, Learn: 0.01, Renown: 0.1,
		Worth: func(*entity.Agent, *world.World, float64) need.Levels { return need.Levels{need.Esteem: 0.04} },
	}
	defer delete(recipes, key)
	sc := &ontology.Schema{Verb: ontology.Make, Inputs: []*ontology.Class{ontology.Stone}, Output: ontology.Tool, SiteTrait: ontology.Bench}
	d := making(ontology.Instance{Key: key, Schema: sc, Ticks: 2})
	if d == nil {
		t.Fatal("a recipe over a schema should be an act")
	}
	if d.Name != "knap" || d.Ticks != 2 {
		t.Fatalf("composed %q over %d ticks", d.Name, d.Ticks)
	}
	if d.Available(a, w) {
		t.Fatal("knapping with no stone should not be possible")
	}
	a.Inventory[entity.Stone] = 3
	esteem := a.Needs[need.Esteem]
	if !run(w, a, d) {
		t.Fatal("knapping with stone and a bench should be possible")
	}
	if a.Pos != a.Home {
		t.Fatalf("knapped at %v, not at the bench at home", a.Pos)
	}
	if a.Inventory[entity.Stone] != 1 || a.Inventory[entity.Tools] != 0.5 {
		t.Fatalf("stone %v tools %v after knapping, want 1 and 0.5", a.Inventory[entity.Stone], a.Inventory[entity.Tools])
	}
	if a.Reputation != 0.05 || a.Skills[entity.Crafting] != 0.01 {
		t.Fatalf("renown %v skill %v, want 0.05 and 0.01", a.Reputation, a.Skills[entity.Crafting])
	}
	if a.Needs[need.Esteem] <= esteem {
		t.Fatal("the esteem a making promises is the esteem it gives")
	}
}

// A schema with no recipe is no act, and neither is a recipe that does not
// fit its schema.
func TestAMakingWithoutARecipeIsNoAct(t *testing.T) {
	sc := &ontology.Schema{Verb: ontology.Make, Inputs: []*ontology.Class{ontology.Stone}, Output: ontology.Tool, SiteTrait: ontology.Bench}
	if d := making(ontology.Instance{Key: "make/nothing", Schema: sc}); d != nil {
		t.Fatalf("no recipe, yet %q was composed", d.Name)
	}
	recipes["make/short"] = recipe{Name: "short", Amounts: []float64{1, 1}}
	defer delete(recipes, "make/short")
	if d := making(ontology.Instance{Key: "make/short", Schema: sc}); d != nil {
		t.Fatalf("a recipe for two inputs over a schema with one, yet %q was composed", d.Name)
	}
}

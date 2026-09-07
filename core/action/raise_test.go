package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// A raising is composed from a plan over a schema, not written: what it is
// built of and what it becomes come from the ontology, how much and where
// and what for from the plan. A granary of stone alone is not in the
// trees; that it can be raised all the same is the point.
func TestARaisingIsComposedFromAPlan(t *testing.T) {
	w, a := workshop(t)
	key := "raise/stone>granary@open"
	plans[key] = plan{
		Name: "wall a granary", Amounts: []float64{2},
		Site: publicPlot(granaryRadius), Learn: 0.02, Renown: 0.1,
		Worth: func(*entity.Agent, *world.World) need.Levels {
			return need.Levels{need.Esteem: 0.1, need.Belonging: 0.05, need.Safety: 0.5}
		},
		Done:  func(w *world.World) { w.Mods.Keeping = 0 },
		Built: "walled a granary",
	}
	defer delete(plans, key)
	sc := &ontology.Schema{Verb: ontology.Raise, Inputs: []*ontology.Class{ontology.Stone}, Output: ontology.Granary, Site: ontology.Open}
	d := raising(ontology.Instance{Key: key, Schema: sc, Ticks: 3})
	if d == nil {
		t.Fatal("a plan over a schema should be an act")
	}
	if d.Name != "wall a granary" || d.Ticks != 3 {
		t.Fatalf("composed %q over %d ticks", d.Name, d.Ticks)
	}
	if d.Available(a, w) {
		t.Fatal("raising with no stone should not be possible")
	}
	a.Inventory[entity.Stone] = 2
	esteem, safety := a.Needs[need.Esteem], a.Needs[need.Safety]
	if !run(w, a, d) {
		t.Fatal("raising with stone and a plot should be possible")
	}
	if w.Grid.At(a.Pos).Structure != world.Granary {
		t.Fatalf("stood on %v, and no granary there", a.Pos)
	}
	if a.Inventory[entity.Stone] != 0 || w.Mods.Keeping != 0 {
		t.Fatal("the raising should take its stone and do what the structure does")
	}
	if a.Reputation != 0.1 || a.Skills[entity.Building] != 0.02 {
		t.Fatalf("renown %v skill %v, want 0.1 and 0.02", a.Reputation, a.Skills[entity.Building])
	}
	if !(a.Needs[need.Esteem] > esteem) || a.Needs[need.Safety] != safety {
		t.Fatal("standing is the builder's on the day; safety is the settlement's to give")
	}
}

// A schema with no plan is no act, and a plan whose output is not a
// structure is not one either.
func TestARaisingWithoutAPlanIsNoAct(t *testing.T) {
	sc := &ontology.Schema{Verb: ontology.Raise, Inputs: []*ontology.Class{ontology.Stone}, Output: ontology.Granary, Site: ontology.Open}
	if d := raising(ontology.Instance{Key: "raise/nothing", Schema: sc}); d != nil {
		t.Fatalf("no plan, yet %q was composed", d.Name)
	}
	plans["raise/tool"] = plan{Name: "tool", Amounts: []float64{1}}
	defer delete(plans, "raise/tool")
	odd := &ontology.Schema{Verb: ontology.Raise, Inputs: []*ontology.Class{ontology.Stone}, Output: ontology.Tool, Site: ontology.Open}
	if d := raising(ontology.Instance{Key: "raise/tool", Schema: odd}); d != nil {
		t.Fatalf("a tool is no structure, yet %q was composed", d.Name)
	}
}

package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// A transfer is composed from a hand over a schema, not written: what
// moves, which way, and to whom come from the ontology, how much and what
// passes with it from the hand. Nobody in the trees takes wood off a
// neighbour; that they could is the point.
func TestATransferIsComposedFromAHand(t *testing.T) {
	w := world.New(11)
	thief, mark := blank(w, "thief"), blank(w, "mark")
	mark.Pos = entity.Pos{X: thief.Pos.X + 1, Y: thief.Pos.Y}
	key := "transfer/timber<holder"
	hands[key] = hand{
		Name: "pilfer wood", Amount: 0.5, Want: 1, Regard: -0.2, Rift: 0.1,
		Worth: func(*entity.Agent) need.Levels { return need.Levels{need.Safety: 0.02, need.Physiological: 0.5} },
		Event: event.Stolen, Text: "pilfered wood from",
	}
	defer delete(hands, key)
	sc := &ontology.Schema{Verb: ontology.Transfer, Object: ontology.Timber, Role: &ontology.Holder, Dir: ontology.Seize}
	d := transferring(ontology.Instance{Key: key, Schema: sc, Ticks: 1})
	if d == nil {
		t.Fatal("a hand over a schema should be an act")
	}
	if d.Name != "pilfer wood" || d.With == nil {
		t.Fatalf("composed %q, with someone: %v", d.Name, d.With != nil)
	}
	if d.Available(thief, w) {
		t.Fatal("with nobody holding wood there is nobody to take it from")
	}
	mark.Inventory[entity.Wood] = 2
	thief.Inventory[entity.Wood] = 1
	if d.Available(thief, w) {
		t.Fatal("a thief with wood enough should not want more")
	}
	thief.Inventory[entity.Wood] = 0
	safety := thief.Needs[need.Safety]
	if !run(w, thief, d) {
		t.Fatal("taking from a neighbour with wood to spare should be possible")
	}
	if thief.Inventory[entity.Wood] != 0.5 || mark.Inventory[entity.Wood] != 1.5 {
		t.Fatalf("thief %v mark %v after the taking, want 0.5 and 1.5", thief.Inventory[entity.Wood], mark.Inventory[entity.Wood])
	}
	if mark.Regard(thief.ID) >= 0 {
		t.Fatal("the victim always knows")
	}
	if !(thief.Needs[need.Safety] > safety) {
		t.Fatal("what a transfer promises of standing it gives")
	}
}

// A hand for a role nobody stands in, or over a thing that is not a good,
// is no act.
func TestATransferWithoutAHandIsNoAct(t *testing.T) {
	sc := &ontology.Schema{Verb: ontology.Transfer, Object: ontology.Timber, Role: &ontology.Holder, Dir: ontology.Seize}
	if d := transferring(ontology.Instance{Key: "transfer/none", Schema: sc}); d != nil {
		t.Fatalf("no hand, yet %q was composed", d.Name)
	}
	hands["transfer/practice"] = hand{Name: "odd", Amount: 1}
	defer delete(hands, "transfer/practice")
	odd := &ontology.Schema{Verb: ontology.Transfer, Object: ontology.Practice, Role: &ontology.Holder}
	if d := transferring(ontology.Instance{Key: "transfer/practice", Schema: odd}); d != nil {
		t.Fatalf("a practice is nothing in a pack, yet %q was composed", d.Name)
	}
}

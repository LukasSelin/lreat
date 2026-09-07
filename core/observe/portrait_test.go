package observe

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/world"
)

func TestPortraitTellsWhatAnAgentIsGoodAt(t *testing.T) {
	w := world.NewSized(1, 40, 20)
	a := w.Spawn("Ada", need.Neutral())
	a.AddSkill(entity.Farming, 0.2)
	a.AddSkill(entity.Crafting, 0.6)
	a.Efficacy[entity.Crafting] = 0.3 // it does not know how good it is

	p := Look(w, a.ID)
	if p == nil {
		t.Fatal("no portrait of a living agent")
	}
	if p.Calling != entity.Crafting {
		t.Fatalf("calling is %v, want crafting", p.Calling)
	}
	if p.Level != 0.6 || p.Efficacy[p.Calling] != 0.3 {
		t.Fatalf("portrait loses the gap between what it can do (%.2f) and what it thinks (%.2f)", p.Level, p.Efficacy[p.Calling])
	}
	if p.Name != "Ada" || p.ID != a.ID {
		t.Fatal("portrait is of somebody else")
	}
}

func TestPortraitShowsTheErrand(t *testing.T) {
	w := world.NewSized(1, 40, 20)
	a := w.SpawnAt("Ada", need.Neutral(), entity.Pos{X: 2, Y: 2})
	a.Plan = &entity.Plan{Action: "farm", Target: entity.Pos{X: 9, Y: 4}, Remaining: 3, Total: 4, Route: []entity.Pos{{X: 3, Y: 3}, {X: 4, Y: 3}}}

	p := Look(w, a.ID)
	if p.Errand == nil {
		t.Fatal("an agent with a plan has no errand in its portrait")
	}
	if !p.Errand.Walking || p.Errand.Steps != 2 {
		t.Fatalf("errand does not show the walk still ahead: %+v", p.Errand)
	}
	a.Pos = a.Plan.Target
	if p := Look(w, a.ID); p.Errand.Walking {
		t.Fatal("an agent standing on its target is still shown walking")
	}
}

// A portrait carries the watched agent's thinking, and only that agent's.
func TestPortraitCarriesTheThinking(t *testing.T) {
	w := world.NewSized(1, 40, 20)
	a := w.Spawn("Ada", need.Neutral())
	b := w.Spawn("Bo", need.Neutral())
	w.Watch(a.ID)
	w.Remember(world.Deliberation{Tick: 1, Agent: a.ID, Weighed: []world.Weighed{
		{Action: "farm", Weight: 0.4, Chance: 0.6, Chosen: true},
		{Action: "rest", Weight: 0.1, Chance: 0.4},
	}})

	if got := Look(w, a.ID).Thinking; len(got) != 1 || got[0].Chose() != "farm" {
		t.Fatalf("the watched agent's thinking is missing: %+v", got)
	}
	if got := Look(w, b.ID).Thinking; len(got) != 0 {
		t.Fatal("an unwatched agent was given somebody else's thinking")
	}
}

// The one being followed may die; a portrait of nobody is how the view is
// told, rather than a portrait of whoever inherited the index.
func TestPortraitOfTheDeadIsNothing(t *testing.T) {
	w := world.NewSized(1, 40, 20)
	a := w.Spawn("Ada", need.Neutral())
	id := a.ID
	w.Agents = nil
	if Look(w, id) != nil {
		t.Fatal("a portrait was drawn of an agent that is gone")
	}
}

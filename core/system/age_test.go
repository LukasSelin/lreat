package system

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/world"
)

// Agents die of old age, and a life is a spread rather than a fixed term.
func TestAgentsDieOfOldAge(t *testing.T) {
	w := world.NewSized(4, 20, 10)
	for i := 0; i < 40; i++ {
		a := w.Spawn("elder", need.Neutral())
		a.Born = -entity.Prime
		a.Health = 1
		a.Needs = need.Levels{1, 1, 1, 1, 1}
	}
	var died []int
	for w.Tick = 0; w.Tick < entity.Lifespan*2 && len(w.Agents) > 0; w.Tick++ {
		before := len(w.Agents)
		Population(w)
		for i := 0; i < before-len(w.Agents); i++ {
			died = append(died, w.Tick)
		}
		for _, a := range w.Agents {
			a.Needs = need.Levels{1, 1, 1, 1, 1} // keep them fed; only age should kill
		}
	}
	if len(w.Agents) > 0 {
		t.Fatalf("%d of 40 agents outlived twice their lifespan", len(w.Agents))
	}
	if died[0] == died[len(died)-1] {
		t.Fatal("every agent died on the same tick; old age should be a risk, not an appointment")
	}
}

// A body wears with age: the same errand costs an elder more than an adult.
func TestElderlyBodiesAreSlower(t *testing.T) {
	w := world.NewSized(5, 20, 3)
	w.Tick = entity.Lifespan
	adult := w.SpawnAt("adult", need.Neutral(), entity.Pos{X: 0, Y: 1})
	elder := w.SpawnAt("elder", need.Neutral(), entity.Pos{X: 0, Y: 1})
	child := w.SpawnAt("child", need.Neutral(), entity.Pos{X: 0, Y: 1})
	for _, a := range []*entity.Agent{adult, elder, child} {
		a.Vitality, a.Health = 1, 1
	}
	adult.Born = w.Tick - entity.Prime + 1
	elder.Born = w.Tick - entity.Lifespan
	child.Born = w.Tick

	if elder.Vigor(w.Tick) >= adult.Vigor(w.Tick) {
		t.Fatalf("elder vigor %.2f, adult %.2f; the old should be slower", elder.Vigor(w.Tick), adult.Vigor(w.Tick))
	}
	if child.Vigor(w.Tick) >= adult.Vigor(w.Tick) {
		t.Fatalf("child vigor %.2f, adult %.2f; the young should be slower", child.Vigor(w.Tick), adult.Vigor(w.Tick))
	}
}

// Children do not have children, however well they are looked after.
func TestChildrenDoNotBear(t *testing.T) {
	w := world.NewSized(6, 20, 10)
	for i := 0; i < 20; i++ {
		a := w.Spawn("young", need.Neutral())
		a.Born = w.Tick
		a.Needs = need.Levels{1, 1, 1, 1, 1}
		a.Health = 1
	}
	before := len(w.Agents)
	for i := 0; i < entity.Maturity; i++ {
		for _, a := range w.Agents {
			a.Needs = need.Levels{1, 1, 1, 1, 1}
		}
		Population(w)
	}
	if len(w.Agents) != before {
		t.Fatalf("population went from %d to %d before anyone grew up", before, len(w.Agents))
	}
}

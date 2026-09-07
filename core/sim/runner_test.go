package sim

import (
	"context"
	"testing"
	"time"

	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/world"
)

func TestRunnerAppliesCommandsAtTickBoundary(t *testing.T) {
	w := world.New(3)
	for i := 0; i < 4; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	player := w.Agents[0].ID

	r := New(w, 0) // clock stopped; we step by hand
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Run(ctx)

	r.Send(Intend{Agent: player, Action: "study"})
	r.StepOnce()

	var plan *entity.Plan
	r.Inspect(func(w *world.World) {
		if w.Tick != 1 {
			t.Errorf("tick = %d, want 1", w.Tick)
		}
		plan = w.Find(player).Plan
	})
	if plan == nil || plan.Action != "study" {
		t.Fatalf("player plan = %+v, want study in progress", plan)
	}
	if plan.Situation == (habit.Signature{}) {
		t.Fatal("the player's plan should record the moment it was made in, so it can teach")
	}

	select {
	case s := <-r.Snapshots():
		if s.Tick != 1 || s.Population != 4 {
			t.Fatalf("snapshot = %+v", s)
		}
	case <-time.After(time.Second):
		t.Fatal("no snapshot published")
	}
}

func TestRunnerTicksOnItsOwn(t *testing.T) {
	w := world.New(3)
	w.Spawn("a", w.RandomPersonality())
	r := New(w, 1000)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Run(ctx)

	deadline := time.After(2 * time.Second)
	for {
		select {
		case s := <-r.Snapshots():
			if s.Tick >= 5 {
				return
			}
		case <-deadline:
			t.Fatal("runner did not reach tick 5 in 2s")
		}
	}
}

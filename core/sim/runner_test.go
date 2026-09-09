package sim

import (
	"context"
	"testing"
	"time"

	"lreat/core/entity"
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

// A viewer that has not collected the last snapshot is not sent another
// every tick: the map is copied when somebody will look at it, or when the
// copy waiting is a graph column old.
func TestASnapshotNobodyCollectedIsNotRetakenEveryTick(t *testing.T) {
	w := world.New(3)
	w.Spawn("a", w.RandomPersonality())
	r := New(w, 0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Run(ctx)

	for i := 0; i < SnapshotEvery/2; i++ {
		r.StepOnce()
	}
	r.Inspect(func(*world.World) {})
	s := <-r.Snapshots()
	if s.Tick != 1 {
		t.Fatalf("the waiting snapshot is of tick %d; nobody collected the first, so it should still be tick 1", s.Tick)
	}
	for i := 0; i < SnapshotEvery-1; i++ {
		r.StepOnce()
	}
	r.Inspect(func(*world.World) {})
	s = <-r.Snapshots()
	if s.Tick != SnapshotEvery/2+1 {
		t.Fatalf("after collecting, the next tick's snapshot should be waiting; got tick %d", s.Tick)
	}
	// The next tick is taken because the last was collected; left
	// uncollected for a whole column after that, it is replaced.
	for i := 0; i < SnapshotEvery+1; i++ {
		r.StepOnce()
	}
	r.Inspect(func(*world.World) {})
	s = <-r.Snapshots()
	if want := SnapshotEvery/2 + 2*SnapshotEvery; s.Tick != want {
		t.Fatalf("a snapshot a column old should have been replaced by tick %d; got tick %d", want, s.Tick)
	}
}

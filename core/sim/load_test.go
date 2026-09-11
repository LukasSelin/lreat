package sim

import (
	"context"
	"testing"
	"time"

	"lreat/core/world"
)

// The runner reports what it is costing, so that a watcher can tell a world
// that has slowed down from a world where little is happening.
func TestLoadReportsWhatTicksCost(t *testing.T) {
	w := world.New(3)
	for i := 0; i < 4; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	r := New(w, 0) // the clock is stopped; the ticks here are stepped by hand
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Run(ctx)

	if l := r.Load(); l != (Load{}) {
		t.Fatalf("a runner that has not ticked reports %+v", l)
	}
	// Enough ticks, over more than one window, that a reading has been
	// published and is of ticks rather than of the pause between them.
	deadline := time.Now().Add(3 * loadWindow)
	for time.Now().Before(deadline) {
		r.StepOnce()
	}
	r.StepOnce()

	l := r.Load()
	if l.Rate <= 0 {
		t.Fatalf("ticks were stepped and the rate is %v", l.Rate)
	}
	if l.Step <= 0 {
		t.Fatalf("a tick took %v, which is no time at all", l.Step)
	}
	if l.Peak < l.Step {
		t.Fatalf("the worst tick (%v) was quicker than the mean one (%v)", l.Peak, l.Step)
	}
	if l.Busy <= 0 || l.Busy > 1 {
		t.Fatalf("the share of the window spent stepping is %.2f", l.Busy)
	}
}

// Measuring is around the tick and never inside the world: whatever the
// clock says, the run itself has to come out the same.
func TestMeasuringDoesNotMoveTheWorld(t *testing.T) {
	step := func() int {
		w := world.New(7)
		for i := 0; i < 8; i++ {
			w.Spawn("a", w.RandomPersonality())
		}
		r := New(w, 0)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go r.Run(ctx)
		for i := 0; i < 200; i++ {
			r.StepOnce()
		}
		var pop int
		r.Inspect(func(w *world.World) { pop = len(w.Agents) })
		return pop
	}
	if a, b := step(), step(); a != b {
		t.Fatalf("two runs of the same seed ended with %d and %d people", a, b)
	}
}

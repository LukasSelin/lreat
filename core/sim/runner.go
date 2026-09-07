// Package sim runs a World in the background.
//
// The Runner owns the world on a single goroutine. Everything outside sends
// Commands in and receives Snapshots out. Wall-clock time lives here and
// nowhere else in the core: the systems only know ticks.
package sim

import (
	"context"
	"time"

	"lreat/core/action"
	"lreat/core/entity"
	"lreat/core/observe"
	"lreat/core/system"
	"lreat/core/world"
)

// Command is a request to change the world, applied at the next tick boundary.
type Command interface {
	Apply(w *world.World)
}

// Intend makes an agent commit to an action on its next decision. This is how
// the player acts: through the same plan slot every other agent uses.
type Intend struct {
	Agent  entity.ID
	Action string
}

func (c Intend) Apply(w *world.World) {
	a := w.Find(c.Agent)
	if a == nil {
		return
	}
	d := action.ByName(c.Action)
	if d == nil {
		return
	}
	target, ok := d.Target(a, w)
	if !ok {
		return
	}
	// Through the same builder as everyone else, so that what the player
	// does teaches the player's habits.
	system.Commit(a, w, d, target)
}

// Func adapts a closure into a Command, for tooling and tests.
type Func func(w *world.World)

func (f Func) Apply(w *world.World) { f(w) }

// Runner drives a world at a chosen speed.
type Runner struct {
	w     *world.World
	cmds  chan Command
	ctrl  chan func()
	snaps chan observe.Snapshot
	tps   float64
	pause bool
}

// New wraps a world. ticksPerSecond sets the initial speed.
func New(w *world.World, ticksPerSecond float64) *Runner {
	return &Runner{
		w:     w,
		cmds:  make(chan Command, 256),
		ctrl:  make(chan func(), 16),
		snaps: make(chan observe.Snapshot, 1),
		tps:   ticksPerSecond,
	}
}

// Run loops until ctx is cancelled. Call it in its own goroutine.
func (r *Runner) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case f := <-r.ctrl:
			f()
			ticker.Reset(r.interval())
		case <-ticker.C:
			if r.pause {
				continue
			}
			r.tick()
		}
	}
}

func (r *Runner) interval() time.Duration {
	if r.tps <= 0 {
		return time.Hour
	}
	return time.Duration(float64(time.Second) / r.tps)
}

// tick applies pending commands, steps once, and publishes a snapshot.
// The snapshot channel holds one item; a slow consumer sees the latest and
// never slows the simulation.
func (r *Runner) tick() {
	for {
		select {
		case c := <-r.cmds:
			c.Apply(r.w)
			continue
		default:
		}
		break
	}
	system.Step(r.w)
	s := observe.Take(r.w)
	select {
	case r.snaps <- s:
	default:
		select {
		case <-r.snaps:
		default:
		}
		r.snaps <- s
	}
}

// Send queues a command for the next tick.
func (r *Runner) Send(c Command) { r.cmds <- c }

// Snapshots yields the latest snapshot after each tick.
func (r *Runner) Snapshots() <-chan observe.Snapshot { return r.snaps }

// SetSpeed changes ticks per second. Zero stops the clock without pausing,
// which is useful together with StepOnce.
func (r *Runner) SetSpeed(ticksPerSecond float64) {
	r.ctrl <- func() { r.tps = ticksPerSecond }
}

// Pause halts ticking; commands still queue.
func (r *Runner) Pause() { r.ctrl <- func() { r.pause = true } }

// Resume continues ticking.
func (r *Runner) Resume() { r.ctrl <- func() { r.pause = false } }

// StepOnce advances exactly one tick regardless of pause state.
func (r *Runner) StepOnce() { r.ctrl <- r.tick }

// Inspect runs f on the simulation goroutine and waits for it. Use it to read
// state safely from tests and tools; never hold the *World after f returns.
func (r *Runner) Inspect(f func(w *world.World)) {
	done := make(chan struct{})
	r.ctrl <- func() {
		f(r.w)
		close(done)
	}
	<-done
}

package sim

import "time"

// What a run costs is invisible from inside it. The world only knows ticks,
// and a tick is the same thing to it whether it took forty microseconds on a
// valley or nine milliseconds on a globe with five thousand people on it —
// so a settlement that has quietly stopped keeping up with the clock looks
// exactly like one that is keeping up, and the only symptom is that the map
// moves more slowly than the speed at the top of the panel says it should.
//
// The meter is here rather than in the view because this is the only
// goroutine that knows: it owns the world, it is the one doing the work, and
// wall-clock time already lives here and nowhere else in the core. It hands
// the answer over through an atomic pointer, which is the cheapest way to
// let a reader that is not in step with the simulation ask a question that
// does not have to be answered on any particular tick.
//
// None of it touches the world, so none of it can move a run: the numbers
// are read off the clock around the simulation, never out of it.

// Load is what the last window of running cost. It is over a window rather
// than a tick because a single tick is noise — a garbage collection, a
// snapshot, the operating system looking elsewhere for a moment — and the
// question this answers is whether the run is keeping up, which is a
// question about a stretch of seconds.
type Load struct {
	// Rate is the ticks per second actually achieved. Against the speed the
	// runner was asked for, it is the whole of whether the machine is
	// keeping up: a run set to sixty and managing eleven is a run whose
	// every other reading is arriving at a fifth of the rate it seems to.
	Rate float64
	// Step is the mean wall time of one tick, everything included: the
	// systems, and the snapshot taken for whoever is watching. Peak is the
	// worst single tick of the window, which is where a stall shows —
	// a mean of two milliseconds hides a hundred-millisecond hitch every
	// second, and the hitch is what is felt.
	Step time.Duration
	Peak time.Duration
	// Busy is the share of the window spent inside a tick rather than
	// waiting for the next one. It is the headroom, read the other way
	// round: at a tenth there is an order of magnitude of speed left to
	// ask for, and at one there is none and the clock is already slipping.
	Busy float64
}

// loadWindow is how long the meter gathers before it publishes. Short enough
// that speeding a run up shows in the panel while the finger is still on the
// key, long enough that a single slow tick does not become the reading.
const loadWindow = 500 * time.Millisecond

// meter is the window being gathered. It is only ever touched on the
// simulation goroutine, so it needs no lock; only the finished reading
// crosses over, and that goes through the atomic.
type meter struct {
	since time.Time
	ticks int
	spent time.Duration
	peak  time.Duration
}

// Load reports the last finished window. It is the zero Load until the first
// one closes, and while a run is paused it stands at whatever it last was:
// nothing is being spent, so there is nothing to measure, and blanking it
// would throw away the reading of the run the pause interrupted.
func (r *Runner) Load() Load {
	if l := r.load.Load(); l != nil {
		return *l
	}
	return Load{}
}

// measure folds one tick into the window and publishes when the window is
// full. start is when the tick began; everything since then was the tick.
func (r *Runner) measure(start time.Time) {
	now := time.Now()
	spent := now.Sub(start)
	if r.meter.since.IsZero() {
		r.meter.since = start
	}
	r.meter.ticks++
	r.meter.spent += spent
	r.meter.peak = max(r.meter.peak, spent)
	elapsed := now.Sub(r.meter.since)
	if elapsed < loadWindow || r.meter.ticks == 0 {
		return
	}
	l := Load{
		Rate: float64(r.meter.ticks) / elapsed.Seconds(),
		Step: r.meter.spent / time.Duration(r.meter.ticks),
		Peak: r.meter.peak,
		Busy: float64(r.meter.spent) / float64(elapsed),
	}
	r.load.Store(&l)
	// The next window starts now rather than at the last publication, so a
	// run stepped by hand after a long pause does not report its one tick
	// spread over the minutes nobody was ticking.
	r.meter = meter{since: now}
}

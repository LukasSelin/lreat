package main

import (
	"fmt"
	"runtime"
	"runtime/metrics"
	"time"

	"github.com/gdamore/tcell/v2"

	"lreat/core/sim"
)

// Everything else in this view is about the settlement. This block is about
// the machine underneath it, and it is here because the two are told apart
// by nothing else on the screen: a world that has slowed to a third of the
// speed it was asked for looks exactly like a world where nothing much is
// happening. The map still moves, the figures still walk, the day count
// still climbs — only slower, and the panel's "60 t/s" goes on saying sixty
// because that is the speed that was asked for and not the speed being had.
//
// So the numbers are in two halves, and the split is the useful part. What
// the run costs — the rate it is actually managing, what a tick takes, how
// much of the wall clock goes into the world — comes from the simulation
// goroutine, which is the only thing that knows. What the program costs —
// the memory, the churn through it, the share of the processors burnt —
// is read off the Go runtime here. The first says whether the world is
// keeping up; the second says what it is keeping up on, and which of the
// two is the thing to fix when it stops.
//
// Reading it costs nothing worth counting. The runtime figures come out of
// runtime/metrics, which is a handful of words copied out of the scheduler
// rather than the stop-the-world that reading a full set of memory stats
// would be, and they are sampled a couple of times a second rather than
// every frame — a number that changes faster than it can be read is not a
// number anybody is reading.
const (
	// sampleEvery is how often the runtime is asked what it is using. Fast
	// enough to catch a heap climbing, slow enough that the asking is not
	// itself part of what is being measured.
	sampleEvery = 500 * time.Millisecond
	// frameWindow is the span the view's own cost is measured over: how
	// long a frame takes to draw and how many were dropped because the next
	// event was already waiting. Longer than the simulation's window
	// because frames are rarer than ticks at any speed worth watching at.
	frameWindow = time.Second
	// loadRows is how many lines the block takes, heading included. The
	// panel shares its height between this, the technologies and the card
	// of whoever is being followed, and the block is anchored at the foot
	// of it, so everything above has to know what it will take before any
	// of it is drawn. Six lines is what this is worth: the card above it
	// is the reason anybody opened the panel, and a machine reading that
	// crowds out an agent's thinking has the priority backwards.
	loadRows = 6
)

// gauge is what the view knows about what it is costing. The drawing half is
// gathered here frame by frame; the runtime half is read off the runtime and
// kept until the next sampling, so that every frame in between draws the
// same figures instead of a column of them flickering.
type gauge struct {
	// The window being gathered: when it opened, how many frames were drawn
	// in it and what they cost, and how many were thrown away unfinished.
	since   time.Time
	count   int
	spent   time.Duration
	skipped int

	// The last finished reading of that window.
	frame  time.Duration
	missed float64

	// The runtime as of taken, and the cumulative counters kept beside it
	// so the next sampling can take the difference. What is wanted of a
	// counter here is a rate, and the total since the program started would
	// flatten every burst into the same tired average.
	taken  time.Time
	memory uint64
	allocs uint64
	cpu    float64
	idle   float64
	gcTime float64
	churn  float64
	busy   float64
	gc     float64
	// read is the sample set, kept rather than rebuilt so that the names
	// are not looked up in the runtime's table on every sampling.
	read []metrics.Sample
}

// The runtime figures worth a place in the block. The memory one is the
// whole of what the program has taken off the operating system and not
// given back, rather than what is live in the heap: what is being asked
// here is what the machine is giving up to this, and the machine cannot
// use a page back merely because the collector has finished with it. The
// allocation total beside it is differenced into a rate, because a program
// making a gigabyte a second of short-lived rubbish and one quietly holding
// a gigabyte are the same number here and nothing alike to run. The
// processor classes are differenced the same way: what share of the cores
// the runtime was given is actually being burnt, and how much of that is
// the collector rather than the world.
const (
	mTotal = iota
	mAllocs
	mCPU
	mIdle
	mGC
)

var wanted = []string{
	mTotal:  "/memory/classes/total:bytes",
	mAllocs: "/gc/heap/allocs:bytes",
	mCPU:    "/cpu/classes/total:cpu-seconds",
	mIdle:   "/cpu/classes/idle:cpu-seconds",
	mGC:     "/cpu/classes/gc/total:cpu-seconds",
}

// drew folds one finished frame into the window, and closes the window once
// it is old enough. start is when the frame began.
func (g *gauge) drew(start time.Time) {
	now := time.Now()
	if g.since.IsZero() {
		g.since = start
	}
	g.count++
	g.spent += now.Sub(start)
	if now.Sub(g.since) < frameWindow || g.count == 0 {
		return
	}
	g.frame = g.spent / time.Duration(g.count)
	g.missed = float64(g.skipped) / float64(g.count+g.skipped)
	g.since, g.count, g.spent, g.skipped = now, 0, 0, 0
}

// dropped counts a frame abandoned because another event was already
// waiting. It is worth counting rather than passing over: a view dropping
// four frames in five is showing numbers a second old whatever the
// simulation underneath is managing. See view.draw.
func (g *gauge) dropped() { g.skipped++ }

// machine reads what the process is using, at most every sampleEvery. Rates
// are taken against the last reading rather than against the start of the
// run, so what the block shows is what is happening now.
func (g *gauge) machine() {
	now := time.Now()
	if !g.taken.IsZero() && now.Sub(g.taken) < sampleEvery {
		return
	}
	if g.read == nil {
		g.read = make([]metrics.Sample, len(wanted))
		for i, name := range wanted {
			g.read[i].Name = name
		}
	}
	metrics.Read(g.read)
	// A runtime that does not keep one of these answers with a kind of
	// Bad, and a blank in the panel is better than a confident zero.
	u := func(i int) uint64 {
		if g.read[i].Value.Kind() != metrics.KindUint64 {
			return 0
		}
		return g.read[i].Value.Uint64()
	}
	f := func(i int) float64 {
		if g.read[i].Value.Kind() != metrics.KindFloat64 {
			return 0
		}
		return g.read[i].Value.Float64()
	}
	allocs := u(mAllocs)
	cpu, idle, gc := f(mCPU), f(mIdle), f(mGC)
	// The first reading has nothing behind it to difference against, so it
	// sets the marks and leaves the rates at zero rather than reporting the
	// whole history of the process as though it had happened just now.
	if !g.taken.IsZero() {
		if secs := now.Sub(g.taken).Seconds(); secs > 0 && allocs >= g.allocs {
			g.churn = float64(allocs-g.allocs) / secs
		}
		// The processor classes are shares of one budget — every core the
		// runtime was given, for every second of the window — so the
		// denominator is the total's own increase and not the wall clock.
		if d := cpu - g.cpu; d > 0 {
			g.busy = clamp((d - (idle - g.idle)) / d)
			g.gc = clamp((gc - g.gcTime) / d)
		}
	}
	g.memory, g.allocs = u(mTotal), allocs
	g.cpu, g.idle, g.gcTime = cpu, idle, gc
	g.taken = now
}

// clamp keeps a share inside the nought-to-one it is supposed to be. The
// processor counters are gathered by several threads at slightly different
// moments and can cross each other by a hair; a panel briefly reporting
// 103% of the cores would look like a bug in the world rather than in the
// arithmetic of reading it.
func clamp(f float64) float64 {
	return min(max(f, 0), 1)
}

// drawLoad writes the block at the foot of the panel, in the panel's own
// two-column grid so that it reads down the same edges as the settlement's
// figures above it. l is what the simulation goroutine has to say, and the
// size of the world is put in the heading beside it: the same tick on a
// valley and on a globe are two different amounts of work, and without the
// ground it covers a duration is a number with nothing to be large or small
// against.
func (v *view) drawLoad(x, y int, l sim.Load, tiles int) {
	sc := v.screen
	dim := tcell.StyleDefault.Dim(true)
	plain := tcell.StyleDefault
	puts(sc, x, y, tcell.StyleDefault.Bold(true), trim(fmt.Sprintf("─ what it costs ─ %s tiles", count(tiles)), panelWidth))
	line := y + 1
	row := func(l1 string, s1 tcell.Style, v1, l2 string, s2 tcell.Style, v2 string) {
		puts(sc, x, line, dim, l1)
		puts(sc, x+statLabel, line, s1, fmt.Sprintf("%*s", statValue, trim(v1, statValue)))
		puts(sc, x+statCol, line, dim, l2)
		puts(sc, x+statCol+statLabel, line, s2, fmt.Sprintf("%*s", statValue, trim(v2, statValue)))
		line++
	}

	// The rate against the speed asked for is the line the whole block is
	// here for, and the only one coloured by how far apart two numbers are:
	// from the moment a run stops keeping up, every other reading on this
	// panel is arriving more slowly than it appears to.
	managed, style := fmt.Sprintf("%.4g", l.Rate), plain
	switch {
	case v.paused:
		managed, style = "paused", dim
	case l.Rate <= 0:
		managed, style = "—", dim
	case l.Rate < 0.5*v.speed:
		style = tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
	case l.Rate < 0.9*v.speed:
		style = tcell.StyleDefault.Foreground(tcell.ColorYellow)
	}
	row("t/s", style, managed, "stepping", share(l.Busy), pct(l.Busy))
	row("a tick", plain, span(l.Step), "worst", plain, span(l.Peak))
	row("a frame", plain, span(v.gauge.frame), "dropped", dim, pct(v.gauge.missed))
	row("memory", plain, size(v.gauge.memory), "churn", plain, rate(v.gauge.churn))
	// The processors get their count beside the share, because a tenth of
	// twenty-four cores and a tenth of two are not the same machine, and a
	// world that steps on one thread is meant to read as the small number
	// it is rather than as an idle one.
	row("cpu", share(v.gauge.busy), fmt.Sprintf("%s/%d", pct(v.gauge.busy), runtime.GOMAXPROCS(0)),
		"gc", plain, pct(v.gauge.gc))
}

// share colours a proportion the way a gauge is read: quiet while there is
// headroom, warm as it runs out. It is for the two figures that are shares
// of something finite — the wall clock and the processors — and not for the
// ones that merely happen to be written as percentages.
func share(f float64) tcell.Style {
	switch {
	case f >= 0.9:
		return tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
	case f >= 0.6:
		return tcell.StyleDefault.Foreground(tcell.ColorYellow)
	}
	return tcell.StyleDefault
}

// pct writes a proportion as a whole percentage, and says there is nothing
// there rather than writing a confident 0%.
func pct(f float64) string {
	switch {
	case f <= 0:
		return "—"
	case f < 0.01:
		return "<1%"
	}
	return fmt.Sprintf("%.0f%%", f*100)
}

// span writes a duration in the largest unit that still has digits in it,
// inside the six columns the panel gives a number. A tick is microseconds
// on a valley and milliseconds on a globe, and both want reading without
// counting zeroes.
func span(d time.Duration) string {
	switch {
	case d <= 0:
		return "—"
	case d < time.Microsecond:
		return fmt.Sprintf("%dns", d.Nanoseconds())
	case d < time.Millisecond:
		return fmt.Sprintf("%.3gµs", float64(d.Nanoseconds())/1e3)
	case d < time.Second:
		return fmt.Sprintf("%.3gms", float64(d.Microseconds())/1e3)
	}
	return fmt.Sprintf("%.3gs", d.Seconds())
}

// rate writes a quantity of bytes a second, and stays blank rather than
// writing "—/s" while there is nothing to report: the first sampling has
// nothing behind it to take a difference against.
func rate(bytes float64) string {
	if bytes <= 0 {
		return "—"
	}
	return size(uint64(bytes)) + "/s"
}

// size writes a quantity of bytes short enough to sit in the grid: the
// runtime deals in bytes and nobody reads nine digits of them.
func size(n uint64) string {
	switch {
	case n == 0:
		return "—"
	case n < 1<<10:
		return fmt.Sprintf("%dB", n)
	case n < 1<<20:
		return fmt.Sprintf("%.3gk", float64(n)/(1<<10))
	case n < 1<<30:
		return fmt.Sprintf("%.3gM", float64(n)/(1<<20))
	}
	return fmt.Sprintf("%.3gG", float64(n)/(1<<30))
}

// count is the same for plain quantities: a globe is half a million tiles
// and the whole number of them is not worth five columns of the panel.
func count(n int) string {
	switch {
	case n < 1000:
		return fmt.Sprintf("%d", n)
	case n < 1000000:
		return fmt.Sprintf("%.3gk", float64(n)/1000)
	}
	return fmt.Sprintf("%.3gM", float64(n)/1000000)
}

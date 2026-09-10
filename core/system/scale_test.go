package system

import (
	"math"
	"testing"

	"lreat/core/entity"
	"lreat/core/world"
)

// How a tick costs as the world fills up.
//
// BenchmarkStepGlobe2000 spawns two thousand people and measures four
// hundred: Spawn puts everybody within four tiles of the one market, two
// thousand of them starve down to what that ground will feed, and the
// settling run is over before the benchmark starts timing. It is a fine
// measure of one crowded settlement and it says nothing about a full
// world, which is what the ground was made big for.
//
// So these scatter parties of forty over the temperate half of the globe,
// each on its own dry ground, and settle them only briefly - long enough
// that people are working and walking and not so long that the weakest
// settlements have died. What they report is the population they actually
// measured and the cost of a tick per head, which is the number that says
// whether a tick costs what the people in it cost or something worse.

// dryish is roughly what share of the globe is land, used only to guess
// how close the lattice has to be to find the sites it needs. Guessing it
// low costs a tighter lattice and nothing else.
const dryish = 0.25

// party is how many people are put down together. It is the founding party
// the rest of the simulation is tuned around.
const party = 40

// scatter puts about n people over the globe in parties, on a lattice of
// dry ground walked in a fixed order, so the same seed gives the same
// settlements in the same places.
func scatter(seed uint64, n int) *world.World {
	w := world.NewWith(seed, world.Globe())
	g := w.Grid
	// Far enough apart that the parties are settlements and not one
	// crowd, close enough that there is room for as many as were asked
	// for. Most of a globe is water, so the lattice has to offer several
	// times the sites it needs and let the wet ones fall out.
	sites := (n + party - 1) / party
	step := int(math.Sqrt(float64(g.W) * float64(g.H/2) * dryish / float64(sites)))
	step = max(2, min(24, step))
	put := 0
	for y := g.H / 4; y < 3*g.H/4 && put < n; y += step {
		for x := 0; x < g.W && put < n; x += step {
			p := entity.Pos{X: x, Y: y}
			if g.At(p).Wet() {
				continue
			}
			for i := 0; i < party && put < n; i++ {
				w.SpawnAt("a", w.RandomPersonality(), p)
				put++
			}
		}
	}
	return w
}

// settling is how long the parties are left before the clock starts. Long
// enough to be living rather than standing where they were put; short
// enough that the population is still near what was asked for.
const settling = 100

func benchScale(b *testing.B, n int) {
	w := scatter(1, n)
	Run(w, settling)
	before := len(w.Agents)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Step(w)
	}
	b.StopTimer()
	// The population moves while it is being measured, so the cost per
	// head is charged against the middle of it rather than either end.
	pop := float64(before+len(w.Agents)) / 2
	b.ReportMetric(pop, "agents")
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/pop, "ns/agent")
}

func BenchmarkStepScale1k(b *testing.B)  { benchScale(b, 1000) }
func BenchmarkStepScale5k(b *testing.B)  { benchScale(b, 5000) }
func BenchmarkStepScale20k(b *testing.B) { benchScale(b, 20000) }
func BenchmarkStepScale50k(b *testing.B) { benchScale(b, 50000) }

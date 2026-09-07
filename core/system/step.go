// Package system holds the per-tick rules. Each system is a function over
// the world, run in a fixed order. None of them reads the clock or any
// randomness outside World.RNG.
package system

import "lreat/core/world"

// Step advances the world by one tick.
func Step(w *world.World) {
	w.Tick++
	Decay(w)
	Land(w)
	Beliefs(w)
	Requests(w)
	Decide(w)
	Act(w)
	MarketStep(w)
	Population(w)
	Upkeep(w)
	Discover(w)
}

// Run advances the world by n ticks.
func Run(w *world.World, n int) {
	for i := 0; i < n; i++ {
		Step(w)
	}
}

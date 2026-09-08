// Package system holds the rules of a day. Each system is a function over
// the world, run in a fixed order. None of them reads the wall clock or any
// randomness outside World.RNG; how long anything takes is said in package
// clock, where a tick is a day.
package system

import "lreat/core/world"

// Step advances the world by one day.
func Step(w *world.World) {
	w.Tick++
	Climate(w)
	Decay(w)
	Land(w)
	Beliefs(w)
	Requests(w)
	Decide(w)
	Act(w)
	MarketStep(w)
	MoveMarket(w)
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

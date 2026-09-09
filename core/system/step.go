// Package system holds the rules of a day. Each system is a function over
// the world, run in a fixed order. None of them reads the wall clock or any
// randomness outside World.RNG; how long anything takes is said in package
// clock, where a tick is a day.
package system

import "lreat/core/world"

// Phase is one system of the day, named so that a tool outside the core can
// say what a tick was spent on. The core itself never reads a clock.
type Phase struct {
	Name string
	Run  func(*world.World)
}

// Phases is the day in order. The order is the order the world's chance is
// drawn in, which every settlement's history depends on; see the remarks on
// upkeep and population for what moving one would cost.
var Phases = []Phase{
	{"climate", Climate},
	{"decay", Decay},
	{"land", Land},
	{"beliefs", Beliefs},
	{"requests", Requests},
	{"decide", Decide},
	{"act", Act},
	{"market", MarketStep},
	{"move-market", MoveMarket},
	{"population", Population},
	{"upkeep", Upkeep},
	{"discover", Discover},
}

// Step advances the world by one day.
func Step(w *world.World) {
	w.Tick++
	for _, p := range Phases {
		p.Run(w)
	}
}

// StepWith is Step with each phase handed to around, which must call run
// exactly once. It is how a runner times the day without the day knowing
// it is being timed.
func StepWith(w *world.World, around func(name string, run func())) {
	w.Tick++
	for _, p := range Phases {
		around(p.Name, func() { p.Run(w) })
	}
}

// Run advances the world by n ticks.
func Run(w *world.World, n int) {
	for i := 0; i < n; i++ {
		Step(w)
	}
}

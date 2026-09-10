package observe

import (
	"testing"

	"lreat/core/system"
	"lreat/core/world"
)

// What a snapshot costs. It is worth watching because of who pays: Take
// reads the whole world, so it runs on the simulation's own goroutine with
// the day stopped around it, and on a globe it costs more than the day does.
// A viewer that asked for one every tick would be the slowest thing in the
// run. See sim.SnapshotEvery, which is how often one is taken for a viewer
// that has not looked at the last.

func settled(agents int, cfg world.Config, ticks int) *world.World {
	w := world.NewWith(1, cfg)
	for i := 0; i < agents; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	system.Run(w, ticks)
	return w
}

// BenchmarkTakeValley is a snapshot of the default map with forty people on
// it: the settlement every other number is taken against.
func BenchmarkTakeValley(b *testing.B) {
	w := settled(40, world.DefaultConfig(), 600)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Take(w)
	}
}

// BenchmarkTakeGlobe is a snapshot of the globe with a founding party on it.
// Nearly all of it is the ground: half a million tiles copied for a viewer,
// whoever is standing on them.
func BenchmarkTakeGlobe(b *testing.B) {
	w := settled(300, world.Globe(), 200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Take(w)
	}
}

// BenchmarkTakeGlobeCrowd is the same ground with two thousand people, which
// is what says whether a snapshot costs the world or the people in it.
func BenchmarkTakeGlobeCrowd(b *testing.B) {
	w := settled(2000, world.Globe(), 200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Take(w)
	}
}

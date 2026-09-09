package system

import (
	"testing"

	"lreat/core/world"
)

// BenchmarkStepDefault40 is a day in a settled settlement of about forty on
// the default map: the number every change to a tick is held against.
func BenchmarkStepDefault40(b *testing.B) {
	w := world.New(1)
	for i := 0; i < 40; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	Run(w, 600)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Step(w)
	}
}

// BenchmarkStepGlobe40 is a day on the globe with a founding party of forty:
// what the ground costs when almost nobody is on it.
func BenchmarkStepGlobe40(b *testing.B) {
	w := world.NewWith(1, world.Globe())
	for i := 0; i < 40; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	Run(w, 600)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Step(w)
	}
}

// BenchmarkStepGlobe2000 is a day on the globe with two thousand people,
// all of them for now around the one market: the crowd the scheduler has
// to carry.
func BenchmarkStepGlobe2000(b *testing.B) {
	w := world.NewWith(1, world.Globe())
	for i := 0; i < 2000; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	Run(w, 200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Step(w)
	}
}

// BenchmarkGenerateGlobe is the making of the globe.
func BenchmarkGenerateGlobe(b *testing.B) {
	for i := 0; i < b.N; i++ {
		world.NewWith(uint64(i+1), world.Globe())
	}
}

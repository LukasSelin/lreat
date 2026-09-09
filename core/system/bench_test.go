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

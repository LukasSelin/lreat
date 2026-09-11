package system

import (
	"testing"
	"time"

	"lreat/core/world"
)

// A day among two parties out of each other's reach, one worker against
// many: what the islands buy the acting. Run with -bench 'TwoParties' and
// read the two against each other; the rest of the day is spread either
// way, so the difference is the acting alone.
func BenchmarkTwoPartiesStep(b *testing.B) {
	for _, c := range []struct {
		name    string
		workers int
	}{{"1worker", 1}, {"24workers", 24}} {
		b.Run(c.name, func(b *testing.B) {
			was := world.Workers
			defer func() { world.Workers = was }()
			w := twoParties(b, 9, c.workers, 400)
			Run(w, 200) // let the parties settle in before timing them
			var acting time.Duration
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				StepWith(w, func(name string, run func()) {
					if name != "act" {
						run()
						return
					}
					t0 := time.Now()
					run()
					acting += time.Since(t0)
				})
			}
			b.ReportMetric(float64(w.Isles.Islands), "islands")
			b.ReportMetric(float64(acting.Nanoseconds())/float64(b.N), "act-ns/op")
		})
	}
}

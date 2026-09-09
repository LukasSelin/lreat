package observe

import (
	"testing"

	"lreat/core/action"
	"lreat/core/habit"
	"lreat/core/system"
	"lreat/core/world"
)

// habitsByTable is the measure as it was first written: a table of every
// agent's unit signature for every act, read twice. Kept here so that the
// measure without the table can be shown to be the same measure.
func habitsByTable(w *world.World) (spread, mean, gated float64) {
	n := float64(len(w.Agents))
	if n == 0 {
		return 0, 0, 0
	}
	units := make([][]habit.Signature, action.Count)
	var gatedN float64
	w.Room()
	for i, d := range action.Catalog {
		units[i] = make([]habit.Signature, 0, len(w.Agents))
		for _, a := range w.Agents {
			h, r := d.Prior, max(d.Reach0, w.ReachFloor[i])
			if a.Imprinted {
				h, r = a.Habits[i], max(a.Reach[i], w.ReachFloor[i])
			}
			units[i] = append(units[i], habit.Unit(h))
			mean += r
			if d.Reach0 < 1 {
				gated += r
				gatedN++
			}
		}
	}
	mean /= n * float64(action.Count)
	if gatedN > 0 {
		gated /= gatedN
	}
	for i := range units {
		var centre habit.Signature
		for _, u := range units[i] {
			for k := range centre {
				centre[k] += u[k] / n
			}
		}
		for _, u := range units[i] {
			var d habit.Signature
			for k := range d {
				d[k] = u[k] - centre[k]
			}
			spread += habit.Norm(d)
		}
	}
	spread /= n * float64(action.Count)
	return spread, mean, gated
}

func TestHabitSpreadIsTheSameWithoutTheScratchTables(t *testing.T) {
	w := world.New(4)
	for i := 0; i < 25; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	system.Run(w, 600)
	s1, m1, g1 := habitsByTable(w)
	s2, m2, g2 := habits(w)
	if s1 != s2 || m1 != m2 || g1 != g2 {
		t.Fatalf("by table %v %v %v, streaming %v %v %v", s1, m1, g1, s2, m2, g2)
	}
}

package world

import (
	"math/rand/v2"
	"testing"

	"lreat/core/entity"
)

// nearbyByWalking is what Nearby replaced: everyone within radius, in the
// order of the population.
func nearbyByWalking(w *World, p entity.Pos, radius int) []entity.ID {
	var out []entity.ID
	for _, a := range w.Agents {
		if w.Grid.Dist(p, a.Pos) <= radius {
			out = append(out, a.ID)
		}
	}
	return out
}

func scatter(w *World, rng *rand.Rand) {
	for _, a := range w.Agents {
		a.Pos = entity.Pos{X: rng.IntN(w.Grid.W), Y: rng.IntN(w.Grid.H)}
		w.Moved(a)
	}
}

func TestNearbyVisitsInTheOrderTheSliceWould(t *testing.T) {
	for _, wrap := range []bool{false, true} {
		width := 300
		if wrap {
			width = 320 // a globe is a whole number of chunks round
		}
		w := NewWith(7, Config{Width: width, Height: 150, Wrap: wrap})
		for i := 0; i < 400; i++ {
			w.Spawn("a", w.RandomPersonality())
		}
		rng := rand.New(rand.NewPCG(1, 2))
		for round := 0; round < 20; round++ {
			scatter(w, rng)
			for q := 0; q < 50; q++ {
				p := entity.Pos{X: rng.IntN(w.Grid.W), Y: rng.IntN(w.Grid.H)}
				radius := 1 + rng.IntN(ChunkSide)
				want := nearbyByWalking(w, p, radius)
				var got []entity.ID
				w.Nearby(p, radius, func(a *entity.Agent) bool { got = append(got, a.ID); return true })
				if len(got) != len(want) {
					t.Fatalf("wrap %v: around %v within %d the file found %d, the walk %d", wrap, p, radius, len(got), len(want))
				}
				for i := range got {
					if got[i] != want[i] {
						t.Fatalf("wrap %v: around %v within %d the file found %v, the walk %v", wrap, p, radius, got, want)
					}
				}
			}
		}
	}
}

// A position set by hand, with nobody told, is put right by the next reader
// that looks at the chunk it was filed under. That is any reader on the
// default map, whose two chunks every reader looks at; on a larger map a
// test that moves an agent by hand must say so with Moved.
func TestAPositionSetByHandIsFoundWhereItIs(t *testing.T) {
	w := NewWith(3, Config{Width: 200, Height: 100})
	a := w.Spawn("a", w.RandomPersonality())
	b := w.Spawn("b", w.RandomPersonality())
	a.Pos, b.Pos = entity.Pos{X: 5, Y: 5}, entity.Pos{X: 110, Y: 80}
	w.Reindex()
	a.Pos = entity.Pos{X: 100, Y: 70} // walked two chunks over without saying
	if got := w.Neighbor(b, 12); got != a {
		t.Fatalf("b's neighbour is %v; a is standing ten tiles off", got)
	}
	if got := w.AgentAt(entity.Pos{X: 6, Y: 6}, 12, nil); got != nil {
		t.Fatalf("someone is still filed where a used to stand: %v", got.Name)
	}
}

func TestFindIsTheSameAfterDeaths(t *testing.T) {
	w := New(4)
	for i := 0; i < 10; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	dead := w.Agents[3]
	w.Agents = append(w.Agents[:3], w.Agents[4:]...)
	w.Reindex()
	if w.Find(dead.ID) != nil {
		t.Fatal("the dead are still on file")
	}
	for _, a := range w.Agents {
		if w.Find(a.ID) != a {
			t.Fatalf("%d is not found", a.ID)
		}
	}
	if w.Find(0) != nil || w.Find(entity.ID(1000)) != nil {
		t.Fatal("nobody and the unborn are found")
	}
}

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
				radius := 1 + rng.IntN(NearbyLimit)
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

// A position set by hand, with nobody told, is put right by the next
// reader that looks at the cell it was filed under. A reader looks at the
// cells its radius touches and no further, so a walk beyond those has to
// be declared with Moved - which is what everything that walks an agent
// does, and what a test that stands one somewhere must do too.
func TestAPositionSetByHandIsFoundWhereItIs(t *testing.T) {
	w := NewWith(3, Config{Width: 200, Height: 100})
	a := w.Spawn("a", w.RandomPersonality())
	b := w.Spawn("b", w.RandomPersonality())
	a.Pos, b.Pos = entity.Pos{X: 100, Y: 70}, entity.Pos{X: 110, Y: 80}
	w.Reindex()
	a.Pos = entity.Pos{X: 104, Y: 74} // a step over, into the next cell
	if got := w.Neighbor(b, 12); got != a {
		t.Fatalf("b's neighbour is %v; a is standing six tiles off", got)
	}
	if got := w.AgentAt(entity.Pos{X: 100, Y: 70}, 1, nil); got != nil {
		t.Fatalf("someone is still filed where a used to stand: %v", got.Name)
	}
	a.Pos = entity.Pos{X: 5, Y: 5} // right across the map, and said so
	w.Moved(a)
	if got := w.Neighbor(b, 12); got != nil {
		t.Fatalf("b's neighbour is %v; a is a hundred tiles off", got)
	}
	if got := w.AgentAt(entity.Pos{X: 6, Y: 6}, 12, nil); got != a {
		t.Fatalf("a is not where it said it had gone: %v", got)
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

// closestByWalking is what Closest must agree with: the nearest agent the
// filter keeps, ties to the earliest born, found by walking everybody.
func closestByWalking(w *World, p entity.Pos, radius int, keep func(*entity.Agent) bool) *entity.Agent {
	var best *entity.Agent
	bestD := radius + 1
	for _, a := range w.Agents {
		if !keep(a) {
			continue
		}
		if d := w.Grid.Dist(p, a.Pos); d < bestD {
			best, bestD = a, d
		}
	}
	return best
}

// Closest stops as soon as no cell it has not opened can hold anybody
// nearer. What it must not do is stop early enough to miss somebody, or to
// take the younger of two people standing equally near.
func TestClosestFindsWhoTheWalkWouldFind(t *testing.T) {
	for _, wrap := range []bool{false, true} {
		width := 300
		if wrap {
			width = 320 // a globe is a whole number of chunks round
		}
		w := NewWith(11, Config{Width: width, Height: 150, Wrap: wrap})
		for i := 0; i < 400; i++ {
			w.Spawn("a", w.RandomPersonality())
		}
		rng := rand.New(rand.NewPCG(3, 4))
		for round := 0; round < 20; round++ {
			scatter(w, rng)
			for q := 0; q < 50; q++ {
				p := entity.Pos{X: rng.IntN(w.Grid.W), Y: rng.IntN(w.Grid.H)}
				radius := 1 + rng.IntN(NearbyLimit)
				// A filter that turns some of them down, so that the
				// search cannot stop at the first person it comes across.
				odd := func(a *entity.Agent) bool { return a.ID%3 != 0 }
				want := closestByWalking(w, p, radius, odd)
				if got := w.Closest(p, radius, odd); got != want {
					t.Fatalf("wrap %v: nearest to %v within %d is %v; the walk found %v",
						wrap, p, radius, got, want)
				}
			}
		}
	}
}

// A crowd all standing on one another is where ties happen, and where the
// order the file is walked in could otherwise show through.
func TestClosestBreaksTiesByAge(t *testing.T) {
	w := NewWith(5, Config{Width: 64, Height: 64})
	rng := rand.New(rand.NewPCG(7, 8))
	for i := 0; i < 200; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	for _, a := range w.Agents {
		a.Pos = entity.Pos{X: 20 + rng.IntN(6), Y: 20 + rng.IntN(6)}
		w.Moved(a)
	}
	all := func(*entity.Agent) bool { return true }
	for q := 0; q < 200; q++ {
		p := entity.Pos{X: 16 + rng.IntN(16), Y: 16 + rng.IntN(16)}
		radius := 1 + rng.IntN(NearbyLimit)
		if got, want := w.Closest(p, radius, all), closestByWalking(w, p, radius, all); got != want {
			t.Fatalf("in the crowd, nearest to %v within %d is %v; the walk found %v", p, radius, got, want)
		}
	}
}

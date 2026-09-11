package system

import (
	"testing"

	"lreat/core/action"
	"lreat/core/world"
)

// runCrowd is runSettlement with a population big enough that the cheap
// passes over it - who is standing beside whom in Beliefs, the wearing of
// each body in Decay, everybody's situation in Discover - are spread over
// goroutines too; thirty people are done in turn. See world.WorkersOver.
func runCrowd(t *testing.T, seed uint64, workers, people, ticks int) string {
	t.Helper()
	was := world.Workers
	world.Workers = workers
	defer func() { world.Workers = was }()

	w := world.New(seed)
	for i := 0; i < people; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	Run(w, ticks)
	return fingerprint(w) + norms(w)
}

// norms is what everybody holds to be right, which is what the beliefs pass
// moves and what the fingerprint does not read.
func norms(w *world.World) string {
	out := ""
	for _, a := range w.Agents {
		for _, v := range a.Norms {
			out += string(rune('a' + int(v*1000)%26))
		}
	}
	return out
}

// The cheap passes over the population read everybody at once and write in
// turn afterwards, or write only each body's own. One worker or many, the
// same people must be worn the same, find the same neighbours, and so hold
// the same values.
func TestSpreadingTheCheapPassesOverGoroutinesChangesNothing(t *testing.T) {
	for _, seed := range []uint64{1, 5} {
		serial := runCrowd(t, seed, 1, 300, 120)
		for _, workers := range []int{2, 8, 24} {
			if got := runCrowd(t, seed, workers, 300, 120); got != serial {
				t.Fatalf("seed %d: %d workers produced a different crowd than 1 worker", seed, workers)
			}
		}
	}
}

// Everybody's situation, read for a discovery to weigh, is the same read
// side by side as read in turn, and the same as reading each alone.
func TestSituationsReadSideBySideAreTheSame(t *testing.T) {
	was := world.Workers
	defer func() { world.Workers = was }()

	w := world.New(7)
	for i := 0; i < 300; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	world.Workers = 1
	Run(w, 60)
	alone := situations(w)
	for i, a := range w.Agents {
		if got := action.Shared(a, w); got != alone[i] {
			t.Fatalf("agent %d: situation read in turn differs from reading it alone", a.ID)
		}
	}
	for _, workers := range []int{2, 8, 24} {
		world.Workers = workers
		spread := situations(w)
		for i := range alone {
			if spread[i] != alone[i] {
				t.Fatalf("%d workers: agent %d's situation differs from the one read in turn", workers, w.Agents[i].ID)
			}
		}
	}
}

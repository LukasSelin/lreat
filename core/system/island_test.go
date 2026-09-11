package system

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/world"
)

// twoParties is a world with two settlements founded far enough apart to be
// islands of their own: a market at each end of a strip of chunks, and
// people spawned around each. They stay apart, because everybody trades at
// the nearest square and children are born where their parents stand.
func twoParties(t testing.TB, seed uint64, workers, each int) *world.World {
	t.Helper()
	world.Workers = workers
	w := world.NewSized(seed, 10*world.ChunkSide, world.ChunkSide)
	w.Agents = w.Agents[:0]
	w.Reindex()
	east := entity.Pos{X: 9*world.ChunkSide + 20, Y: world.ChunkSide / 2}
	w.MarketPos = entity.Pos{X: 20, Y: world.ChunkSide / 2}
	w.Grid.Build(w.MarketPos, world.Market)
	w.FoundMarket(w.MarketPos)
	w.Grid.Build(east, world.Market)
	w.FoundMarket(east)
	for i := 0; i < each; i++ {
		for _, m := range []entity.Pos{w.MarketPos, east} {
			p := entity.Pos{X: m.X + w.RNG.IntN(9) - 4, Y: m.Y + w.RNG.IntN(9) - 4}
			w.SpawnAt("a", w.RandomPersonality(), w.Grid.Norm(p))
		}
	}
	return w
}

// runParties runs a two-party world and says what came of it, and how many
// islands the acting was cut into at most.
func runParties(t *testing.T, seed uint64, workers, ticks int) (string, int) {
	t.Helper()
	was := world.Workers
	defer func() { world.Workers = was }()
	w := twoParties(t, seed, workers, 30)
	most := 0
	for i := 0; i < ticks; i++ {
		Step(w)
		most = max(most, w.Isles.Islands)
	}
	return fingerprint(w) + norms(w), most
}

// Two parties out of each other's reach act on goroutines of their own.
// One worker or many, the same seed must give the same two settlements:
// the islands' order is a fact about the run, not about the scheduling.
func TestPartiesActingApartChangeNothingBetweenWorkers(t *testing.T) {
	for _, seed := range []uint64{2, 11} {
		serial, most := runParties(t, seed, 1, 400)
		if most < 2 {
			t.Fatalf("seed %d: the parties were never two islands; this exercised nothing", seed)
		}
		for _, workers := range []int{2, 8} {
			if got, _ := runParties(t, seed, workers, 400); got != serial {
				t.Fatalf("seed %d: %d workers produced different parties than 1 worker", seed, workers)
			}
		}
	}
}

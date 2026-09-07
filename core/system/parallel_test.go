package system

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/observe"
	"lreat/core/world"
)

// fingerprint is everything about a settlement that ought to be the same
// between two runs of the same seed: where everyone is, what shape they are
// in, what they are doing, and what has been built.
func fingerprint(w *world.World) string {
	s := observe.Take(w)
	out := ""
	for _, a := range w.Agents {
		plan := "-"
		if a.Plan != nil {
			plan = a.Plan.Action
		}
		out += string(rune('0'+a.ID%10)) + plan +
			posKey(a.Pos) + posKey(entity.Pos{X: int(a.Needs[0] * 1000), Y: int(a.Health * 1000)})
	}
	out += posKey(entity.Pos{X: s.Houses, Y: s.Roads})
	out += posKey(entity.Pos{X: s.Fields, Y: s.Population})
	out += posKey(entity.Pos{X: int(s.Knowledge), Y: s.Friendships})
	return out
}

func posKey(p entity.Pos) string {
	return string(rune('a'+p.X%26)) + string(rune('a'+p.Y%26)) +
		string(rune('a'+(p.X/26)%26)) + string(rune('a'+(p.Y/26)%26))
}

func runSettlement(t *testing.T, seed uint64, workers, ticks int) string {
	t.Helper()
	was := Workers
	Workers = workers
	defer func() { Workers = was }()

	w := world.New(seed)
	for i := 0; i < 30; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	Run(w, ticks)
	return fingerprint(w)
}

// The whole point of giving each agent its own luck is that spreading the
// deciding over goroutines must not change what the settlement does. One
// worker or many, the same seed has to produce the same settlement.
func TestDecidingInParallelChangesNothing(t *testing.T) {
	for _, seed := range []uint64{1, 3, 9} {
		serial := runSettlement(t, seed, 1, 600)
		for _, workers := range []int{2, 3, 8, 16} {
			if got := runSettlement(t, seed, workers, 600); got != serial {
				t.Fatalf("seed %d: %d workers produced a different settlement than 1 worker", seed, workers)
			}
		}
	}
}

// Deciding reads the world from several goroutines at once. Under -race this
// is what catches anything in the read path that quietly writes.
func TestDecidingConcurrentlyIsRaceFree(t *testing.T) {
	was := Workers
	Workers = 8
	defer func() { Workers = was }()

	w := world.New(2)
	for i := 0; i < 40; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	Run(w, 400)
	if len(w.Agents) == 0 {
		t.Fatal("everyone died; this exercised nothing")
	}
}

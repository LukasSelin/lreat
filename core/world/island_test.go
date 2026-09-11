package world

import (
	"reflect"
	"strings"
	"sync"
	"testing"

	"lreat/core/entity"
	"lreat/core/event"
)

// Every field of World and of Grid has a policy for what an island's view
// does with it, and nothing has a policy that is not a field. Adding a
// field to either is deciding what an island does with it.
func TestEveryFieldHasAnIslandPolicy(t *testing.T) {
	seen := map[string]bool{}
	for _, typ := range []reflect.Type{reflect.TypeOf(World{}), reflect.TypeOf(Grid{})} {
		for i := 0; i < typ.NumField(); i++ {
			name := typ.Field(i).Name
			seen[name] = true
			if _, ok := islandPolicy[name]; !ok {
				t.Errorf("%s.%s has no island policy; say what a view does with it in islandPolicy", typ.Name(), name)
			}
		}
	}
	for name, policy := range islandPolicy {
		if !seen[name] {
			t.Errorf("islandPolicy names %q, which is no field of World or Grid", name)
		}
		switch policy {
		case "shared", "read", "own", "scratch", "kept":
		default:
			t.Errorf("islandPolicy[%q] = %q is not a policy", name, policy)
		}
	}
}

// party is a world with people standing where the test says, filed where
// they stand.
func party(t *testing.T, w *World, at ...entity.Pos) []*entity.Agent {
	t.Helper()
	var out []*entity.Agent
	for _, p := range at {
		a := w.SpawnAt("a", w.RandomPersonality(), p)
		out = append(out, a)
	}
	return out
}

// People in chunks far enough apart that nobody on one could reach ground
// anybody on the other reaches are islands of their own; people any nearer
// are one island. The islands come out in order of their lowest chunk, and
// each holds its people in agent order.
func TestPeopleOutOfReachAreIslands(t *testing.T) {
	w := NewSized(3, 10*ChunkSide, ChunkSide) // ten chunks in a row
	w.Agents = w.Agents[:0]
	w.Reindex()
	// Chunks 0, 3 and 5 hold people: 3 and 5 are two apart, within reach
	// of each other's ground; 0 is three from 3, out of reach.
	party(t, w,
		entity.Pos{X: 5*ChunkSide + 10, Y: 10},
		entity.Pos{X: 3*ChunkSide + 10, Y: 20},
		entity.Pos{X: 10, Y: 30},
		entity.Pos{X: 5*ChunkSide + 20, Y: 40},
	)
	isles := w.Islands()
	if len(isles) != 2 {
		t.Fatalf("got %d islands, want 2: %+v", len(isles), isles)
	}
	if isles[0].Key != 0 || isles[1].Key != 3 {
		t.Fatalf("islands keyed %d and %d, want 0 and 3", isles[0].Key, isles[1].Key)
	}
	if got := isles[1].Chunks; len(got) != 2 || got[0] != 3 || got[1] != 5 {
		t.Fatalf("island 3 holds chunks %v, want [3 5]", got)
	}
	names := func(is Island) string {
		var s []string
		for _, a := range is.Agents {
			s = append(s, string(rune('0'+a.ID)))
		}
		return strings.Join(s, "")
	}
	if names(isles[0]) != "3" || names(isles[1]) != "124" {
		t.Fatalf("islands hold %q and %q, want \"3\" and \"124\"", names(isles[0]), names(isles[1]))
	}
	if islandApart != 3 {
		t.Fatalf("islandApart = %d; the chunks above were laid out for 3", islandApart)
	}
}

// On a globe the east edge joins the west, so people at either end of the
// map are one island.
func TestIslandsGoRoundAGlobe(t *testing.T) {
	w := NewWith(4, Config{Width: 10 * ChunkSide, Height: ChunkSide, Wrap: true, Settlements: 1})
	w.Agents = w.Agents[:0]
	w.Reindex()
	party(t, w, entity.Pos{X: 10, Y: 10}, entity.Pos{X: 9*ChunkSide + 10, Y: 10})
	if isles := w.Islands(); len(isles) != 1 {
		t.Fatalf("got %d islands across the seam, want 1", len(isles))
	}
}

// Where the people are one island, they act on the world itself, in the
// one order there always was.
func TestOneIslandActsOnTheWorldItself(t *testing.T) {
	w := New(5)
	for i := 0; i < 8; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	calls, held := 0, 0
	var handed *World
	w.EachIsland(func(v *World, is Island) {
		calls++
		handed, held = v, len(is.Agents)
	})
	if calls != 1 {
		t.Fatalf("f was called %d times, want once", calls)
	}
	if handed != w {
		t.Fatal("one island was handed a view rather than the world")
	}
	if held != len(w.Agents) {
		t.Fatalf("the one island holds %d of %d people", held, len(w.Agents))
	}
	if w.Isles.Islands != 1 || w.Isles.Largest != 1 {
		t.Fatalf("Isles = %+v, want one island holding everybody", w.Isles)
	}
}

// What islands do to the settlement as a whole is put together afterwards,
// island by island in key order: what they told, in that order; what they
// closed and founded; how far they moved the order, the knowledge and the
// market; whether they moved the water. The island the principal square
// stands on draws the world's own chance; any other draws its own.
func TestIslandsPutTheDayTogether(t *testing.T) {
	w := NewSized(6, 10*ChunkSide, ChunkSide)
	w.Agents = w.Agents[:0]
	w.Reindex()
	w.MarketPos = entity.Pos{X: 8*ChunkSide + 10, Y: 10}
	w.FoundMarket(w.MarketPos)
	party(t, w, entity.Pos{X: 10, Y: 10}, w.MarketPos)
	markets := len(w.Markets())
	w.Safety, w.Knowledge = 0.5, 10
	w.Market.Stock[entity.Food] = 100
	first := w.Post(entity.Request{Requester: w.Agents[0].ID, Reward: 1})
	second := w.Post(entity.Request{Requester: w.Agents[1].ID, Reward: 1})
	if first == 0 || second == 0 {
		t.Fatal("could not post")
	}
	waters := w.Grid.Waters()
	worldRNG := w.RNG
	// The islands run on goroutines of their own, so what they find out is
	// taken down under a lock and judged afterwards.
	var mu sync.Mutex
	var drewWorld, drewOwn, handedWorld bool
	var odd []int
	w.EachIsland(func(v *World, is Island) {
		mu.Lock()
		defer mu.Unlock()
		if v == w {
			handedWorld = true
		}
		switch is.Key {
		case 0:
			v.Emit(event.Acted, 0, 0, "west")
			v.Safety += 0.3
			v.Knowledge += 1
			v.Market.Stock[entity.Food] -= 10
			v.Close(first)
			v.FoundMarket(entity.Pos{X: 20, Y: 20})
			v.Grid.wet()
			drewOwn = v.RNG != worldRNG
		case 8:
			v.Emit(event.Acted, 0, 0, "east")
			v.Safety += 0.4
			v.Knowledge += 2
			v.Market.Stock[entity.Food] -= 20
			v.Grid.wet()
			drewWorld = v.RNG == worldRNG
		default:
			odd = append(odd, is.Key)
		}
	})
	if handedWorld {
		t.Fatal("two islands were handed the world itself")
	}
	if len(odd) > 0 {
		t.Fatalf("islands keyed %v", odd)
	}
	told := w.Log.Since(0)
	if len(told) < 2 || told[len(told)-2].Text != "west" || told[len(told)-1].Text != "east" {
		t.Fatalf("the log ends %+v, want west then east", told)
	}
	if w.Safety != 1 { // 0.5 + 0.3 + 0.4, clamped
		t.Fatalf("Safety = %v, want 1", w.Safety)
	}
	if w.Knowledge != 13 {
		t.Fatalf("Knowledge = %v, want 13", w.Knowledge)
	}
	if w.Market.Stock[entity.Food] != 70 {
		t.Fatalf("food stock = %v, want 70", w.Market.Stock[entity.Food])
	}
	if w.HasRequest(first) || !w.HasRequest(second) {
		t.Fatal("the board should have lost the first request and kept the second")
	}
	if len(w.Markets()) != markets+1 {
		t.Fatalf("%d markets, want %d: those there were and the one the west island founded", len(w.Markets()), markets+1)
	}
	if w.Grid.Waters() != waters+2 {
		t.Fatalf("the water moved %d times, want 2", w.Grid.Waters()-waters)
	}
	if !w.Grid.regionsStale {
		t.Fatal("the water moved and the labels are not stale")
	}
	if !drewWorld || !drewOwn {
		t.Fatalf("the market's island drew the world's chance: %v; the other drew its own: %v", drewWorld, drewOwn)
	}
	if w.Isles.Islands != 2 || w.Isles.Largest != 0.5 {
		t.Fatalf("Isles = %+v, want two islands of half each", w.Isles)
	}
}

// An island's own chance is the same stream every day it is that island,
// and a different one from every other island's.
func TestAnIslandKeepsItsOwnChance(t *testing.T) {
	w := NewSized(7, 10*ChunkSide, ChunkSide)
	w.Agents = w.Agents[:0]
	w.Reindex()
	w.MarketPos = entity.Pos{X: 8*ChunkSide + 10, Y: 10}
	party(t, w, entity.Pos{X: 10, Y: 10}, entity.Pos{X: 4*ChunkSide + 10, Y: 10}, w.MarketPos)
	var mu sync.Mutex
	streams := map[int]float64{}
	w.EachIsland(func(v *World, is Island) {
		mu.Lock()
		defer mu.Unlock()
		streams[is.Key] = v.RNG.Float64()
	})
	again := map[int]float64{}
	w2 := NewSized(7, 10*ChunkSide, ChunkSide)
	w2.Agents = w2.Agents[:0]
	w2.Reindex()
	w2.MarketPos = w.MarketPos
	party(t, w2, entity.Pos{X: 10, Y: 10}, entity.Pos{X: 4*ChunkSide + 10, Y: 10}, w2.MarketPos)
	w2.EachIsland(func(v *World, is Island) {
		mu.Lock()
		defer mu.Unlock()
		again[is.Key] = v.RNG.Float64()
	})
	if len(streams) != 3 {
		t.Fatalf("%d islands, want 3", len(streams))
	}
	for k := range streams {
		if streams[k] != again[k] {
			t.Fatalf("island %d drew %v one run and %v the next", k, streams[k], again[k])
		}
	}
	if streams[0] == streams[4] {
		t.Fatal("two islands drew the same chance")
	}
}

// A view that brings somebody into the world is stopped: that is the
// settlement's to do, one at a time.
func TestAnIslandThatSpawnsIsStopped(t *testing.T) {
	w := NewSized(8, 10*ChunkSide, ChunkSide)
	w.Agents = w.Agents[:0]
	w.Reindex()
	party(t, w, entity.Pos{X: 10, Y: 10}, entity.Pos{X: 8*ChunkSide + 10, Y: 10})
	defer func() {
		if recover() == nil {
			t.Fatal("an island spawned somebody and nothing stopped it")
		}
	}()
	w.EachIsland(func(v *World, is Island) {
		if is.Key == 0 {
			v.SpawnAt("b", v.RandomPersonality(), entity.Pos{X: 11, Y: 11})
		}
	})
}

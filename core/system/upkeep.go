package system

import (
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/world"
)

// Standing is not the same as being kept. A house stands because somebody
// lives in it and a field is a field because somebody works it, so when the
// person is gone what they made starts going too. Without this a settlement
// only ever accumulates: every roof its founders raised on ground they had
// just arrived on is still standing three generations later, in everybody's
// way, and the land under it can never be put to anything else. Ruin is what
// gives a settlement the room to be something other than what it first was.
const (
	// leftToFall is the per-tick chance that a building nobody keeps falls
	// in, and leftToWeeds the chance a field nobody works goes back to
	// grass. About three hundred ticks either way: long enough that a house
	// outlives its builder and an heir could take it on, short enough that a
	// settlement is not walled in by its dead.
	leftToFall  = 1.0 / 300
	leftToWeeds = 1.0 / 300
)

// Upkeep lets what nobody keeps fall down. A tile belongs to whoever built
// it; when that person is no longer alive, nobody is keeping it, and it goes
// back to the land at its own pace.
func Upkeep(w *world.World) {
	alive := make(map[entity.ID]bool, len(w.Agents))
	for _, a := range w.Agents {
		alive[a.ID] = true
	}
	g := w.Grid
	for i := range g.Tiles {
		t := &g.Tiles[i]
		if t.Owner == 0 || alive[t.Owner] {
			continue
		}
		p := entity.Pos{X: i % g.W, Y: i / g.W}
		switch {
		case t.Structure == world.House:
			if w.RNG.Float64() < leftToFall && g.Raze(p) {
				w.Emit(event.Ruined, 0, 0, "an empty house fell in")
			}
		case t.Terrain == world.Field:
			if w.RNG.Float64() < leftToWeeds {
				g.Raze(p)
			}
		default:
			// Something claimed but neither lived in nor sown: the claim
			// goes, and the ground is open to whoever comes next.
			g.Raze(p)
		}
	}
}

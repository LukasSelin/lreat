package system

import (
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/ontology"
	"lreat/core/world"
)

// Upkeep lets what nobody keeps fall down. A tile belongs to whoever built
// it; when that person is no longer alive, nobody is keeping it, and it goes
// back to the land at its own pace. What that pace is, for each kind of
// thing, is stated in ontology.Transforms.
func Upkeep(w *world.World) { wither(w) }

// wither is the runner for what nobody is left to keep.
//
// The order it draws the world's chance in is the whole of what has to be
// preserved here: the living are counted first, the tiles are walked by
// index, a tile still in somebody's keeping is passed over before any luck
// is spent on it, and what is certain to go draws no luck at all. A number
// drawn and not used is a different settlement three generations on.
func wither(w *world.World) {
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
		// What becomes of it is settled before any luck is spent, so that
		// a certainty costs the world no draw.
		tr := ontology.Unkept(world.ClassOf(t))
		if tr == nil {
			continue
		}
		if tr.Rate < 1 && w.RNG.Float64() >= tr.Rate {
			continue
		}
		p := entity.Pos{X: i % g.W, Y: i / g.W}
		// Nobody did this, so there is no act to name - only the ground it
		// happened on.
		if g.Raze(p) && tr.Says != "" {
			w.EmitAt(event.Ruined, 0, 0, "", p, "%s", tr.Says)
		}
	}
}

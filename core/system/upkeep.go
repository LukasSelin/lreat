package system

import (
	"lreat/core/event"
	"lreat/core/world"
)

// Upkeep lets what nobody keeps fall down. A tile mostly belongs to whoever
// built it; when that person is no longer alive, nobody is keeping it, and
// it goes back to the land at its own pace. A public work belongs to nobody
// and goes anyway, slower. What that pace is, for each kind of thing, is
// stated in ontology.Transforms.
func Upkeep(w *world.World) { wither(w) }

// wither is the runner for what nobody is left to keep.
//
// The order it draws the world's chance in is the whole of what has to be
// preserved here: the living are counted first, the tiles are walked by
// index, a tile still in somebody's keeping is passed over before any luck
// is spent on it, and what is certain to go draws no luck at all. A number
// drawn and not used is a different settlement three generations on.
func wither(w *world.World) {
	g := w.Grid
	// Only the awake ground is walked, in the same order the whole would
	// be. Nothing stands on sleeping ground and nobody holds it, so nothing
	// there could befall and no luck would have been spent on it.
	// A chunk with nothing built on it and nothing held is passed over
	// whole: nothing befalls bare ground on its own, which the ontology
	// says and world.BareBefalls checks.
	var only func(c int) bool
	if !world.BareBefalls {
		only = func(c int) bool { return g.Chunks[c].Built > 0 || g.Chunks[c].Owned > 0 }
	}
	g.EachActive(only, func(i, _ int, t *world.Tile) {
		// What becomes of it is settled before any luck is spent, so that
		// a certainty costs the world no draw, and so that ground with
		// nothing standing on it costs none either.
		tr := t.Befalling(unkept(t, w))
		if tr == nil {
			return
		}
		if tr.Rate < 1 && w.RNG.Float64() >= tr.Rate {
			return
		}
		p := g.PosOf(i)
		// Nobody did this, so there is no act to name - only the ground it
		// happened on.
		if g.Raze(p) && tr.Says != "" {
			w.EmitAt(event.Ruined, 0, 0, "", p, tr.Says)
		}
	})
}

// unkept says whether whatever was holding this tile against the weather is
// gone, which is what lets the Kept transforms apply to it.
//
// For nearly everything a settlement builds that is a person: a house stands
// because somebody lives in it, a field is a field because somebody works it,
// and when they are dead there is nobody. A road is the exception. Nobody
// owns one, so no death takes it and no life saves it, and asked the ordinary
// question a road answered "still kept" for ever. What keeps a road is the
// walking on it, which the ground itself records: see world.Walked.
func unkept(t *world.Tile, w *world.World) bool {
	if t.Structure == world.Road {
		return t.Traffic < world.Walked
	}
	return t.Owner != 0 && w.Find(t.Owner) == nil
}

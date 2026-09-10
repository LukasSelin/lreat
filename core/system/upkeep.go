package system

import (
	"slices"

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
//
// So it is two walks. The first works out what becomes of every tile and
// writes down the few it found something for; that is all of the reading and
// none of the chance, and it is spread over goroutines. The second goes
// through what the first found, in tile order, and is where the luck is
// spent and the ground is razed - one tile at a time, as it always was.
//
// The second walk is short. Nearly all of the ground in a chunk with a
// settlement on it is grass and wood that nothing becomes of, and the whole
// cost of the day's ruin used to be walking over it to find that out; what
// is left to do in order is the handful of houses, fields and roads that
// something might actually become of.
//
// Nothing the first walk reads is anything the second walk writes. Razing a
// tile changes that tile and its chunk's tallies and nothing else, so what
// befalls a tile is the same worked out before the walk as during it.
func wither(w *world.World) {
	g := w.Grid
	r := w.Ruin()
	// Only the awake ground is walked, in the same order the whole would
	// be. Nothing stands on sleeping ground and nobody holds it, so nothing
	// there could befall and no luck would have been spent on it.
	// A chunk with nothing built on it and nothing held is passed over
	// whole: nothing befalls bare ground on its own, which the ontology
	// says and world.BareBefalls checks - so this is a saving and never a
	// rule, and settling it once for both walks, rather than asking again
	// as razing empties chunks under the second, decides nothing. What the
	// tiles of such a chunk would answer is nil either way.
	for c := range r.Admit {
		r.Admit[c] = world.BareBefalls || g.Chunks[c].Built > 0 || g.Chunks[c].Owned > 0
		if r.Admit[c] {
			r.Found[c] = r.Found[c][:0]
		}
	}
	only := func(c int) bool { return r.Admit[c] }
	w.Freeze(true)
	g.EachActiveOver(only, func(i, c int, t *world.Tile) {
		// Ground nothing can befall either way is left before anybody asks
		// whose it is: that question is a look into the population by a
		// number off the tile, and most of the ground here would answer
		// nothing.
		if !t.Befalls() {
			return
		}
		// What becomes of it is settled before any luck is spent, so that
		// a certainty costs the world no draw, and so that ground with
		// nothing standing on it costs none either.
		tr := t.Befalling(unkept(i, t, w))
		if tr == nil {
			return
		}
		r.Befall[i] = tr
		r.Found[c] = append(r.Found[c], int32(i))
	})
	w.Freeze(false)
	// The chunks' findings gathered into one walk in tile order, which for
	// a map kept row by row is the order the whole ground would be walked
	// in - and so the order the old single walk drew its chance in.
	r.Order = r.Order[:0]
	for c := range r.Found {
		if r.Admit[c] && g.Awake(c) {
			r.Order = append(r.Order, r.Found[c]...)
		}
	}
	slices.Sort(r.Order)
	for _, i := range r.Order {
		tr := r.Befall[i]
		if tr.Rate < 1 && w.RNG.Float64() >= tr.Rate {
			continue
		}
		p := g.PosOf(int(i))
		// Nobody did this, so there is no act to name - only the ground it
		// happened on.
		if g.Raze(p) && tr.Says != "" {
			w.EmitAt(event.Ruined, 0, 0, "", p, tr.Says)
		}
	}
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
//
// It only reads, which is what lets the first of wither's two walks ask it
// on several goroutines at once. The look into the population is by number
// into a file the world is frozen over; see world.Freeze.
func unkept(i int, t *world.Tile, w *world.World) bool {
	if t.Structure == world.Road {
		return w.Grid.Traffic[i] < world.Walked
	}
	return t.Owner != 0 && w.Find(t.Owner) == nil
}

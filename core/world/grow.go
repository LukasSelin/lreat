package world

import "lreat/core/ontology"

// What grows on a tile takes time to come on, and that time is not the same
// for everything growing. How long each thing takes, and what stages it
// passes through on the way, is stated in the ontology - see
// ontology.Processes. This is where those stages are read off an actual
// tile, and system.Land is what advances them.
//
// These are growing ticks rather than ticks: what is measured is how much
// growing weather a stand has had, so a wood raised in the autumn stands
// still until the thaw.

// alive is the same question read as a table over the ground, worked out
// once from the same bindings rather than asked of the ontology again. Every
// tile on the map is asked whether it is alive on every tick, and gathering
// the processes that run on it to answer allocates the gathering.
var alive = func() (a [Tavern + 1][Rock + 1]bool) {
	for s := range a {
		for t := range a[s] {
			tile := Tile{Structure: Structure(s), Terrain: Terrain(t)}
			a[s][t] = len(ontology.Growing(ClassOf(&tile))) > 0
		}
	}
	return a
}()

// Alive reports whether this tile carries a standing crop, which is to say
// something that had to grow before it could be taken. It is exactly the
// tiles some process runs on: the ontology says which ground carries
// something alive, and ClassOf says which terrain that is.
func (t *Tile) Alive() bool { return alive[t.Structure][t.Terrain] }

// The age of what stands on a tile is kept in a layer beside the map - see
// Layers - so what asks after it asks the grid, by the tile's index.

// Along is how far through process p what stands on tile i has come, in
// [0,1]. It is the reading a stage is asked for against, and it is p's
// share of the tile's one age: a wood is coming on as brush and as timber
// at once, and is far further along the first than the second.
func (g *Grid) Along(i int, p *ontology.Process) float64 {
	return g.Grown(i, p.Full())
}

// Reached reports whether what stands on tile i has come as far as ph. It
// is what lets an act ask for a stage - a harvest wants a field in ear -
// and it is written as the share of the whole rather than as an age in
// ticks because the two disagree in the last bit, and one field flipping on
// one tick re-rolls every draw in the run after it.
func (g *Grid) Reached(i int, ph ontology.Phase) bool {
	return g.Along(i, ph.Process) >= ph.Share()
}

// Grown is how far along what grows on tile i is, in [0,1], against the
// time such a thing takes to come on. It only rises: a wood that has made
// its timber holds it, and a stand does not go over and take the wood with
// it. What starts it again is the ground being cleared and something else
// sown on it.
func (g *Grid) Grown(i int, full float64) float64 {
	if full <= 0 {
		return 1
	}
	return clamp01(g.Age[i] / full)
}

// Sow starts whatever is to grow on tile i over: the ground is bare, and
// what stands on it from now on is this year's, not last year's. It is
// called wherever the terrain changes hands - a wood seeded or planted, a
// wood felled to a clearing, a strip broken, a strip harvested, a road laid
// over any of them - so that nothing inherits the age of what it replaced.
func (g *Grid) Sow(i int) { g.Age[i] = 0 }

// Standing puts tile i's growth at full, for ground that is meant to have
// been there all along: the woods a map is made with are old woods.
func (g *Grid) Standing(i int) { g.Age[i] = ontology.Timbering.Full() }

// growing is which processes run on each kind of ground, worked out once
// in the same way as alive. Asking the ontology gathers them afresh, and
// this is asked of every tile on every day.
// A process and the stock on the tile it fills, if any: the two are looked
// up together once so that the day's pass looks up nothing.
type filling struct {
	p     *ontology.Process
	stock func(*Tile) *float64
}

var growing = func() (g [Tavern + 1][Rock + 1][]filling) {
	for s := range g {
		for t := range g[s] {
			tile := Tile{Structure: Structure(s), Terrain: Terrain(t)}
			for _, p := range ontology.Growing(ClassOf(&tile)) {
				g[s][t] = append(g[s][t], filling{p: p, stock: StockOf(p.Yields)})
			}
		}
	}
	return g
}()

// grown puts back what growing weather puts back, up to what the stand's
// age accounts for. Age bounds what a stand grows into, and nothing else:
// it never takes away what is already standing, so a wood is only ever
// held back from filling out, never thinned by the calendar.
func grown(have, ceiling, by float64) float64 {
	if have >= ceiling {
		return have
	}
	return min(ceiling, have+by)
}

// Ripen advances what is growing on this tile by k of growing weather: the
// stand gets that much older, and whatever it is coming on toward fills a
// little further, bounded by the age it has had. A wood does both of these
// twice over, on two clocks - the brush under it within a few years, the
// timber over a lifetime - which is why the age is the tile's and the
// filling is the process's.
//
// A process whose yield the ground keeps no count of only ages the tile. A
// field is the case: a crop is cut once and wholly rather than drawn down,
// so what a strip has to give is read off how far along it is and there is
// no stock to put back.
//
// It is the day's pass with k a day's weather, and the catching up of a
// chunk that slept with k a season's; see active.go.
func (g *Grid) Ripen(i int, k float64) {
	t := &g.Tiles[i]
	ps := growing[t.Structure][t.Terrain]
	if len(ps) == 0 {
		return
	}
	// A stand ages by the weather it gets, not by the calendar: what a
	// winter gives it is nothing, and that is the same clock everything
	// else growing keeps.
	g.Age[i] += k
	for _, f := range ps {
		if f.p.Rate == 0 || f.stock == nil {
			continue
		}
		s := f.stock(t)
		*s = grown(*s, g.Along(i, f.p), f.p.Rate*k)
	}
}

// Green is how much of what could be growing here is standing, in [0,1]. It
// is the reading a satellite takes rather than the one a surveyor takes: not
// what the ground could grow, which is Rich and does not change from one year
// to the next, but what is on it this morning.
//
// A stand that keeps a count of itself is read off the count, because that is
// what a taking draws down: a wood gathered to nothing reads as nothing while
// the trees are still called a wood. A stand that keeps no count - a crop is
// cut once and wholly - is read off how far along it is, so a sown strip is
// bare, a strip in ear is full, and the same strip is bare again the day
// after the harvest. Where a tile has more than one thing growing on it, as a
// wood has its brush and its timber, the reading is the mean of them.
//
// Ground with nothing growing on it at all is nothing: bare rock, open water,
// what is under a roof, and open grass, which carries no crop anybody can
// take. That last one is the reading disagreeing with the eye, and it is the
// simulation's own answer rather than a picture of one: what this map shades
// is what there is to be had.
func (g *Grid) Green(i int) float64 {
	t := &g.Tiles[i]
	ps := growing[t.Structure][t.Terrain]
	if len(ps) == 0 {
		return 0
	}
	sum := 0.0
	for _, f := range ps {
		if f.stock != nil {
			sum += clamp01(*f.stock(t))
			continue
		}
		sum += g.Along(i, f.p)
	}
	return sum / float64(len(ps))
}

// How much a water tile's fish come back per growing day, and how much
// worn fertility a field recovers per growing day toward what the land
// can hold. Neither of these is a process: a shoal is a stock that
// replenishes, not a crop that has to come on, and worn soil is resting
// rather than growing.
const (
	FishRegrowth = 0.0012
	Fallow       = 0.0006
)

// Replenish is what k of growing weather puts back on tile i that is not a
// stand coming on: the fish in the water and the rest a worn field gets.
func (g *Grid) Replenish(i int, k float64) {
	t := &g.Tiles[i]
	switch t.Terrain {
	case Water:
		t.Fish = min(1, t.Fish+FishRegrowth*k)
	case Field:
		g.Fertility[i] = min(g.Rich[i], g.Fertility[i]+Fallow*k)
	}
}

// befalling is what befalls each kind of ground kept and unkept, worked out
// once in the same way as growing. Asking the ontology walks the class up
// its parents and the transforms along, and this is asked of every awake
// tile on every day.
var befalling = func() (b [Tavern + 1][Rock + 1][2]*ontology.Transform) {
	for s := range b {
		for t := range b[s] {
			tile := Tile{Structure: Structure(s), Terrain: Terrain(t)}
			b[s][t][0] = ontology.Befalling(ClassOf(&tile), false)
			b[s][t][1] = ontology.Befalling(ClassOf(&tile), true)
		}
	}
	return b
}()

// befalls is whether anything at all becomes of this kind of ground on its
// own, kept or unkept. It is the two entries above read as one question, so
// that ground nothing can befall either way is passed over before anybody
// asks whose it is - which is most of the ground in a settled chunk, and
// the asking is a look into the population by a number off the tile.
var befalls = func() (b [Tavern + 1][Rock + 1]bool) {
	for s := range b {
		for t := range b[s] {
			b[s][t] = befalling[s][t][0] != nil || befalling[s][t][1] != nil
		}
	}
	return b
}()

// Befalls reports whether anything can become of this ground on its own,
// whoever is or is not keeping it. Where it is false Befalling is nil both
// ways, so nothing follows from asking.
func (t *Tile) Befalls() bool { return befalls[t.Structure][t.Terrain] }

// Befalling is what becomes of this tile on its own, given whether whoever
// was keeping it is gone. Nil where nothing does.
func (t *Tile) Befalling(unkept bool) *ontology.Transform {
	if unkept {
		return befalling[t.Structure][t.Terrain][1]
	}
	return befalling[t.Structure][t.Terrain][0]
}

// BareBefalls is whether anything befalls bare ground on its own: ground
// with nothing built on it and nobody holding it. The ontology says nothing
// does, and this is read off the same table the ruin is, so that a pass
// over what nobody keeps can leave such ground alone in the certainty
// that it would have drawn no luck.
var BareBefalls = func() bool {
	for t := range befalling[None] {
		if befalling[None][t][0] != nil {
			return true
		}
	}
	return false
}()

// Recovers reports whether ground of this kind puts something back on its
// own when it is left alone, besides what grows on it by its age: the fish
// in the water, the rest a worn field gets. See Replenish, which is where
// the pace is; an outcrop is stone and does not grow, and that is meant.
func (k Terrain) Recovers() bool {
	switch k {
	case Water, Field:
		return true
	}
	return false
}

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

// Along is how far through process p what stands here has come, in [0,1].
// It is the reading a stage is asked for against, and it is p's share of
// the tile's one age: a wood is coming on as brush and as timber at once,
// and is far further along the first than the second.
func (t *Tile) Along(p *ontology.Process) float64 {
	return t.Grown(p.Full())
}

// Reached reports whether what stands here has come as far as ph. It is
// what lets an act ask for a stage - a harvest wants a field in ear - and
// it is written as the share of the whole rather than as an age in ticks
// because the two disagree in the last bit, and one field flipping on one
// tick re-rolls every draw in the run after it.
func (t *Tile) Reached(ph ontology.Phase) bool {
	return t.Along(ph.Process) >= ph.Share()
}

// Grown is how far along what grows here is, in [0,1], against the time such
// a thing takes to come on. It only rises: a wood that has made its timber
// holds it, and a stand does not go over and take the wood with it. What
// starts it again is the ground being cleared and something else sown on it.
func (t *Tile) Grown(full float64) float64 {
	if full <= 0 {
		return 1
	}
	return clamp01(t.Age / full)
}

// Sow starts whatever is to grow here over: the ground is bare, and what
// stands on it from now on is this year's, not last year's. It is called
// wherever the terrain changes hands - a wood seeded or planted, a wood
// felled to a clearing, a strip broken, a strip harvested, a road laid over
// any of them - so that nothing inherits the age of what it replaced.
func (t *Tile) Sow() { t.Age = 0 }

// Standing puts a tile's growth at full, for ground that is meant to have
// been there all along: the woods a map is made with are old woods.
func (t *Tile) Standing() { t.Age = ontology.Timbering.Full() }

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
func (t *Tile) Ripen(k float64) {
	ps := growing[t.Structure][t.Terrain]
	if len(ps) == 0 {
		return
	}
	// A stand ages by the weather it gets, not by the calendar: what a
	// winter gives it is nothing, and that is the same clock everything
	// else growing keeps.
	t.Age += k
	for _, f := range ps {
		if f.p.Rate == 0 || f.stock == nil {
			continue
		}
		s := f.stock(t)
		*s = grown(*s, t.Along(f.p), f.p.Rate*k)
	}
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

// Replenish is what k of growing weather puts back that is not a stand
// coming on: the fish in the water and the rest a worn field gets.
func (t *Tile) Replenish(k float64) {
	switch t.Terrain {
	case Water:
		t.Fish = min(1, t.Fish+FishRegrowth*k)
	case Field:
		t.Fertility = min(t.Rich, t.Fertility+Fallow*k)
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

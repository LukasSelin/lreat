package world

import (
	"math"

	"lreat/core/clock"
)

// The ground that is awake. Every day the weather passes over every tile
// of the map - wear fades, stands grow, fish come back, worn soil rests,
// what nobody keeps falls down - and on a map big enough to hold more than
// one settlement nearly all of that ground has nobody on it and nothing
// built on it. A chunk with nobody in it, nothing standing on it, nobody
// across it these two seasons, and none of the first two next door, is
// asleep: the passes skip it, and what the weather would have done to it
// is done in one go when it wakes, or on a slow sweep so that nothing
// sleeps longer than a season.
//
// Catching up is the same arithmetic as the day's pass with a season's
// growing weather in place of a day's, so it is deterministic, but it is
// not the pass taken a day at a time to the last bit: a stand that would
// have been held back by its age one day and let go the next comes out a
// hair different. So the map a settlement is measured on must never have
// a chunk asleep, and on the default map none ever is - the market's chunk
// is always occupied and the other is beside it. A test says so.

// Growing is the growing weather the world has had since it was made, in
// growing days: the sum over every day of what that day let green things
// grow. A chunk stamps it when it is passed over, and what the chunk is
// owed is the difference.
//
// Weathered is the day a chunk was last weathered, for the wear that fades
// by the day whatever the season.

// sweepOver is how long the sweep takes to reach every sleeping chunk once.
const sweepOver = clock.Season

// wearMemory is how long a crossing keeps ground awake. Wear fades by a
// third in a season, so after one a single crossing is well under what
// counts as a way, and what is left of it can be faded when the ground
// next wakes.
const wearMemory = clock.Season

// Wake works out which chunks are awake this day, catches up any that were
// asleep and are not now, and sweeps a few that still are.
func (w *World) Wake() {
	g := w.Grid
	w.whole()
	if len(g.Active) != len(g.Chunks) {
		g.Active = make([]bool, len(g.Chunks))
	}
	// Settled ground - something built on it or held - is awake and wakes
	// its neighbours, so that what a settlement reads across a chunk's edge
	// is read off ground the day has passed over. Ground with people on it,
	// or walked on lately, is awake on its own account and wakes nothing: a
	// scout on the far side of the country is not a settlement, and neither
	// is the trail behind it.
	settled := make([]bool, len(g.Chunks))
	for i := range g.Chunks {
		c := &g.Chunks[i]
		if c.Trodden {
			c.Trodden, c.Trod = false, w.Tick
		}
		settled[i] = c.Built > 0 || c.Owned > 0
	}
	w.Awake = AwakeCount{Chunks: len(g.Chunks)}
	for i := range g.Chunks {
		was := g.Active[i]
		c := &g.Chunks[i]
		peopled, worn := len(w.cells[i]) > 0, c.Trod >= 0 && w.Tick-c.Trod <= wearMemory
		g.Active[i] = peopled || worn
		cx, cy := i%g.CW, i/g.CW
		for dy := -1; dy <= 1 && !g.Active[i]; dy++ {
			y := cy + dy
			if y < 0 || y >= g.CH {
				continue
			}
			for dx := -1; dx <= 1; dx++ {
				x := cx + dx
				if g.Wrap {
					x = ((x % g.CW) + g.CW) % g.CW
				} else if x < 0 || x >= g.CW {
					continue
				}
				if settled[y*g.CW+x] {
					g.Active[i] = true
					break
				}
			}
		}
		switch {
		case settled[i]:
			w.Awake.Settled++
		case peopled:
			w.Awake.Peopled++
		case worn:
			w.Awake.Worn++
		case g.Active[i]:
			w.Awake.Beside++
		}
		switch {
		case g.Active[i] && !was:
			w.CatchUp(i)
		case was && !g.Active[i] && w.ways != nil:
			// Nothing here asks for a road any more; see readWays, which
			// only reads the ground that is awake.
			g.eachIn(i, func(j int, _ *Tile) { w.ways.draw[j] = 0 })
		}
	}
	// The sweep: enough sleeping chunks a day that every one is caught up
	// once a season, taken in turn.
	n := (len(g.Chunks) + sweepOver - 1) / sweepOver
	for k := 0; k < n; k++ {
		i := (w.swept + k) % len(g.Chunks)
		if !g.Active[i] {
			w.CatchUp(i)
		}
	}
	w.swept = (w.swept + n) % len(g.Chunks)
}

// CatchUp does to chunk i what the days it slept through would have done,
// and stamps it as weathered up to yesterday. The day's own pass follows.
func (w *World) CatchUp(i int) {
	g := w.Grid
	c := &g.Chunks[i]
	growth := w.Growing[i/g.CW] - c.Grown
	days := w.Tick - 1 - c.Weathered
	if growth > 0 || days > 0 {
		fade := math.Pow(Fade, float64(max(0, days)))
		g.eachIn(i, func(_ int, t *Tile) {
			if days > 0 && t.Traffic > 0 {
				t.Traffic *= fade
			}
			if growth > 0 {
				t.Ripen(growth)
				t.Replenish(growth)
			}
		})
	}
	c.Grown, c.Weathered = w.Growing[i/g.CW], w.Tick-1
}

// CatchUpAll catches up every sleeping chunk, for before the whole ground is
// remade at once.
func (w *World) CatchUpAll() {
	g := w.Grid
	for i := range g.Chunks {
		if len(g.Active) == len(g.Chunks) && !g.Active[i] {
			w.CatchUp(i)
		}
	}
}

// Stamp marks every awake chunk as passed over today with the growing
// weather so far. The day's passes call it when they are done.
func (g *Grid) Stamp(growing []float64, tick int) {
	for i := range g.Chunks {
		if len(g.Active) != len(g.Chunks) || g.Active[i] {
			g.Chunks[i].Grown, g.Chunks[i].Weathered = growing[i/g.CW], tick
		}
	}
}

// Rates is what this day's weather lets green things grow on each chunk
// row, as the growing weather of the row's middle: the weather goes by
// latitude, and a chunk is the finest the sleeping ground is reckoned by.
// On a valley every row reads the same.
func (w *World) Rates() []float64 {
	g := w.Grid
	if len(w.rates) != g.CH {
		w.rates = make([]float64, g.CH)
	}
	for cy := range w.rates {
		mid := min(g.H-1, cy*ChunkSide+ChunkSide/2)
		w.rates[cy] = w.Mods.Regrowth * w.Climate.GrowthAt(mid)
	}
	return w.rates
}

// EachActive visits every tile of every awake chunk that only admits, or of
// every awake chunk when only is nil, in the order a walk over the whole
// map would visit them: row by row, and along each row. The caller is told
// which chunk each tile is in, which it would otherwise have to divide for.
// That order is the order the world's chance is drawn in by what nobody
// keeps, so it is kept exactly. A map that has never been woken is read as
// all awake, so that the day's systems can be run on their own.
func (g *Grid) EachActive(only func(c int) bool, f func(i, c int, t *Tile)) {
	all := len(g.Active) != len(g.Chunks)
	for cy := 0; cy < g.CH; cy++ {
		y0, y1 := cy*ChunkSide, min(g.H, (cy+1)*ChunkSide)
		for y := y0; y < y1; y++ {
			row := y * g.W
			for cx := 0; cx < g.CW; cx++ {
				c := cy*g.CW + cx
				if (!all && !g.Active[c]) || (only != nil && !only(c)) {
					continue
				}
				x0, x1 := cx*ChunkSide, min(g.W, (cx+1)*ChunkSide)
				for i := row + x0; i < row+x1; i++ {
					f(i, c, &g.Tiles[i])
				}
			}
		}
	}
}

// eachIn visits every tile of chunk i, row by row.
func (g *Grid) eachIn(i int, f func(j int, t *Tile)) {
	c := &g.Chunks[i]
	for y := c.Y0; y < c.Y0+c.H; y++ {
		row := y * g.W
		for j := row + c.X0; j < row+c.X0+c.W; j++ {
			f(j, &g.Tiles[j])
		}
	}
}

// Awake reports whether chunk i is awake. A map never woken is all awake.
func (g *Grid) Awake(i int) bool {
	return len(g.Active) != len(g.Chunks) || g.Active[i]
}

// worn is which chunks anybody has walked in these two seasons, or that
// are beside one, as of day tick. It is what the case for a road is read
// over: wear a step away is the most a tile's case can be made of.
func (g *Grid) worn(tick int) []bool {
	lately := make([]bool, len(g.Chunks))
	for i := range g.Chunks {
		c := &g.Chunks[i]
		lately[i] = c.Trodden || (c.Trod >= 0 && tick-c.Trod <= wearMemory)
	}
	return g.spread(lately)
}

// spread is which chunks are marked or beside a marked one.
func (g *Grid) spread(mark []bool) []bool {
	out := make([]bool, len(g.Chunks))
	for i := range g.Chunks {
		cx, cy := i%g.CW, i/g.CW
		for dy := -1; dy <= 1 && !out[i]; dy++ {
			y := cy + dy
			if y < 0 || y >= g.CH {
				continue
			}
			for dx := -1; dx <= 1; dx++ {
				x := cx + dx
				if g.Wrap {
					x = ((x % g.CW) + g.CW) % g.CW
				} else if x < 0 || x >= g.CW {
					continue
				}
				if mark[y*g.CW+x] {
					out[i] = true
					break
				}
			}
		}
	}
	return out
}

// AwakeCount is how many chunks are awake and on what account, for a runner
// that says what a day was spent on.
type AwakeCount struct {
	Chunks, Settled, Beside, Peopled, Worn int
}

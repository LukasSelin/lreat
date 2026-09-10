package system

import (
	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/ontology"
	"lreat/core/world"
)

const (
	// How much a water tile's fish come back per tick, and how much worn
	// fertility a field recovers per tick toward what the land can hold.
	// Neither of these is a process: a shoal is a stock that replenishes,
	// not a crop that has to come on, and worn soil is resting rather than
	// growing. What a wood and a field have standing on them is in
	// ontology.Processes; see ripen.
	// reseedSamples is how many random tiles per tick are checked for
	// spontaneous reforestation from a wooded neighbor, on the default map;
	// a bigger map is checked in proportion, so that a wood comes back at
	// the same rate per acre wherever it is.
	reseedSamples = 3
	reseedPer     = world.DefaultWidth * world.DefaultHeight
	reseedChance  = 0.3
	// ErodeEvery is how long an age of weather covers. An age is a decade,
	// not a season: the land is not weather, and a hillside that moved every
	// year would be a different thing altogether - a settlement would watch
	// its own river wander in the time it took to raise a barn. What the
	// ground does, it does over decades and centuries, so a run of a few
	// generations sees the slopes it cleared begin to go and a run of
	// several sees the valley they went into.
	//
	// It is also why the land is worked over as a whole rather than a little
	// each day: the drainage has to be recomputed for the whole map at once
	// for the rivers to be anywhere sensible, and doing that every day would
	// buy nothing but a slower run.
	ErodeEvery = 10 * clock.Year
)

func isForest(t *world.Tile) bool { return t.Is(ontology.Wood) }

// Land lets forests regrow and slowly reclaim unclaimed grass beside them,
// weathers the ground every so often so that the hills wear into the valleys
// and the rivers go where the new heights send them,
// lets wild food and fish come back, and lets a worn field recover toward
// what the land can hold. Clearing land is therefore reversible only where
// nobody has settled, so the built footprint of a settlement persists while
// the wild edges shift; and taking too much from any one place leaves it
// poorer for a long while, which is what makes a settlement move on to
// other ways of living.
func Land(w *world.World) {
	g := w.Grid
	if w.Tick%ErodeEvery == 0 {
		w.Erode()
	}
	// Green things keep the season's hours. In the cold half of the year
	// nothing regrows, so what a summer left standing is what a winter has
	// to live on; over a whole year the growth is what it was before the
	// seasons existed. Fish and fallow follow the same clock: the water
	// under ice gives back nothing, and worn ground rests until it thaws.
	// Only the ground that is awake is passed over; what is asleep is owed
	// the growing weather from here on, and gets it when it wakes. See
	// world.Wake. The weather goes by latitude and by height, so the rate
	// does by chunk.
	rates := w.Rates()
	for i, k := range rates {
		w.Growing[i] += k
	}
	// The wear fades in the same pass as the growing. They were two passes
	// and the first came first, but neither reads what the other writes,
	// so one walk over the awake ground does both and it is the same day.
	//
	// The walk takes the ground a row of a chunk at a time and does each
	// row as loops over the layers - the wear, then what grows, then what
	// comes back - rather than a tile at a time; see world.Grow, which is
	// Ripen and Replenish tile by tile to the last bit. Nothing in it reads
	// a tile's numbers but that tile's, writes any but those, or draws the
	// world's chance, so the rows are spread over goroutines - see
	// world.EachActiveRow. It is the largest thing a day spends itself on
	// once the country is bigger than one settlement: it grows with the
	// ground that is awake rather than with the people, so the more of the
	// world is lived in the more this is the day.
	g.EachActiveRow(nil, func(lo, hi, c int) {
		g.FadeWear(lo, hi, world.Fade)
		g.Grow(lo, hi, rates[c])
	})
	g.Stamp(w.Growing, w.Tick)
	// The hedges are read off the fields as they now stand, so a strip broken
	// yesterday is inside its block's fence today and a holding given up is
	// open ground again. See world.Fence.
	g.Fence()
	samples := max(reseedSamples, (reseedSamples*len(g.Tiles)+reseedPer/2)/reseedPer)
	for k := 0; k < samples; k++ {
		p := entity.Pos{X: w.RNG.IntN(g.W), Y: w.RNG.IntN(g.H)}
		t := g.At(p)
		if !t.Buildable() || !g.HasNeighbor(p, isForest) {
			continue
		}
		// Seed falls everywhere and takes where the ground will hold it. A
		// wood that spread wherever a seed landed closed over the whole map
		// and left nowhere to break a field; a wood that stops at the tree
		// line comes back over what was cleared and no further.
		if !g.HoldsWood(p) {
			continue
		}
		// Seed falls in the growing season, not on frozen ground.
		if w.RNG.Float64() < reseedChance*w.GrowthAt(p) {
			g.Turn(p, world.Forest)
			i := g.Index(p)
			g.Wood[i], g.Wild[i] = 0, 0
			g.Sow(i) // a seedling wood, with nothing on it yet
			// Seed that falls on sleeping ground is owed nothing of the
			// growing weather the ground slept through, but that ground
			// will be given all of it when it wakes. So the seedling is
			// sown that much before its time, and comes out at nought.
			if c := g.ChunkOf(i); !g.Awake(c) {
				g.Age[i] -= w.Growing[c] - g.Chunks[c].Grown
			}
		}
	}
}

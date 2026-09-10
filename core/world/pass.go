package world

// The day's pass over the ground, as flat loops over the layers.
//
// What the weather does to a tile in a day is stated tile by tile in
// grow.go - Ripen, then Replenish - and that is the definition. This is the
// same arithmetic done to a run of tiles at once, a layer at a time, so
// that the pass streams the layers rather than picking each tile's numbers
// out of the map: what kind of ground each tile is - what stands on it and
// what it is made of, as one word - is kept beside the map as a layer of
// its own, and every loop here is a loop over one slice asking that word
// whether the tile is the kind it is looking for. Nothing in the pass
// reads a tile. It is the shape the pass has to be in for the arithmetic
// to be done several tiles at a time, which is what pass_simd_amd64.go
// does with it where the build and the processor allow; pass_noasm.go is
// the same loops one tile at a time, and the helpers here are the tails of
// the runs either way.
//
// Nothing here changes a result to the last bit. Each tile is given the
// same operations on the same operands in the same order as Ripen and
// Replenish give it, and no tile's numbers are read by another tile's, so
// the order the tiles are taken in is not a fact about anything. The rate
// of a filling times the day's weather is worked out once for the run
// rather than once per tile; it is the same product. What that does rule
// out is the compiler fusing a multiply into the add that follows it, which
// Go allows and which the amd64 compiler does not do at the default level
// the golden numbers are taken at; a machine that fused it before would not
// fuse it now, and would be off by a last bit from itself. See
// core/system/golden_test.go.

// A kind is what stands on a tile and what it is made of, as one word:
// the structure in the high byte and the terrain in the low, so that the
// terrain alone can be read back off it with a mask. It is a word rather
// than a byte because the pass compares it lane for lane against the
// numbers it gates, and the numbers are eight bytes wide. The map keeps
// one for every tile in Layers.Kinds; Build and Turn write it, and Recount
// takes it afresh.
const (
	kindShift = 8
	kindMask  = 1<<kindShift - 1
	// kindSpan is one more than the largest kind, for the tables by kind.
	kindSpan = int(Tavern)<<kindShift | int(Rock) + 1
)

// kindOf is the kind of a tile.
func kindOf(t *Tile) int64 {
	return int64(t.Structure)<<kindShift | int64(t.Terrain)
}

// ages is which kinds of ground carry something growing, and so get older
// with the weather: the alive table read by kind. isWater and isField are
// which kinds are water and which are field, whatever stands on them, which
// is the question Replenish asks. aging is the kinds that age listed out,
// for a pass that asks the question of several tiles at once.
var ages, isWater, isField [kindSpan]bool
var aging []int64

func init() {
	for s := 0; s <= int(Tavern); s++ {
		for t := 0; t < int(TerrainCount); t++ {
			k := kindOf(&Tile{Structure: Structure(s), Terrain: Terrain(t)})
			ages[k] = alive[s][t]
			isWater[k] = Terrain(t) == Water
			isField[k] = Terrain(t) == Field
			if ages[k] {
				aging = append(aging, k)
			}
		}
	}
}

// stocking is one process that fills a stock on one kind of ground: which
// kind, how long the process takes in full, how much of the stock a growing
// day puts on, and the layer the stock is kept in. It is a filling with its
// kind beside it and its span worked out, so that the pass asks the
// ontology nothing.
type stocking struct {
	kind       int64
	full, rate float64
	// spanned is whether the process takes any time at all - full over
	// nought - settled here so that the pass need not compare it.
	spanned bool
	stock   func(*Grid) []float64
}

// stocked is every process that fills a stock, by kind and then in the
// order the growing table has them: the order Ripen fills them in.
var stocked = func() (out []stocking) {
	for s := 0; s <= int(Tavern); s++ {
		for t := 0; t < int(TerrainCount); t++ {
			for _, f := range growing[s][t] {
				if f.p.Rate == 0 || f.stock == nil {
					continue
				}
				out = append(out, stocking{
					kind:    kindOf(&Tile{Structure: Structure(s), Terrain: Terrain(t)}),
					full:    f.p.Full(),
					rate:    f.p.Rate,
					spanned: f.p.Full() > 0,
					stock:   f.stock,
				})
			}
		}
	}
	return
}()

// FadeWear fades the wear on the tiles [lo, hi) by the given share: a day's
// Fade, or a sleep's worth of days at once. Wear is never below nothing -
// what marks the ground adds to it and what fades it multiplies it by a
// share of one - so ground with none on it is left with none, and there is
// nothing to ask before multiplying.
func (g *Grid) FadeWear(lo, hi int, by float64) {
	fade(g.Traffic[lo:hi], by)
}

// Grow gives the tiles [lo, hi) k of growing weather: what grows on them
// gets that much older and fills toward what its age accounts for, the
// water gets its fish back and worn fields rest. It is Ripen and then
// Replenish, tile by tile, done as loops over the layers; see the remarks
// at the top of the file.
func (g *Grid) Grow(lo, hi int, k float64) {
	ks := g.Kinds[lo:hi]
	// A stand ages by the weather it gets, not by the calendar; see Ripen.
	// The age is put on before anything reads it.
	age := g.Age[lo:hi]
	grow(age, ks, k)
	// Whatever is coming on fills a little further, bounded by the age it
	// has had; see grown. Each filling is a pass of its own over the run,
	// in the order the growing table has them.
	for _, e := range stocked {
		fill(e.stock(g)[lo:hi], age, ks, e.kind, e.spanned, e.full, e.rate*k)
	}
	// And what comes back that is not a stand coming on; see Replenish.
	shoal(g.Fish[lo:hi], ks, FishRegrowth*k)
	rest(g.Fertility[lo:hi], g.Rich[lo:hi], ks, Fallow*k)
}

// The passes one tile at a time. They are the whole of the pass where the
// arithmetic is not done several tiles at once, and the tail of every run
// where it is, and each is the statement in grow.go with the tile's kind
// read off ks rather than off the tile.

// fadeScalar is FadeWear over a run.
func fadeScalar(wear []float64, by float64) {
	for i := range wear {
		wear[i] *= by
	}
}

// growScalar puts k of weather on the age of every tile that has something
// growing on it.
func growScalar(age []float64, ks []int64, k float64) {
	for j, kk := range ks {
		if ages[kk] {
			age[j] += k
		}
	}
}

// fillScalar fills the stock s on every tile of the given kind by what by
// puts on, up to what its age over full accounts for; see grown and Grown.
func fillScalar(s, age []float64, ks []int64, kind int64, full, by float64) {
	for j, kk := range ks {
		if kk != kind {
			continue
		}
		ceiling := 1.0
		if full > 0 {
			ceiling = clamp01(age[j] / full)
		}
		s[j] = grown(s[j], ceiling, by)
	}
}

// shoalScalar puts the fish back in the water, up to full.
func shoalScalar(fish []float64, ks []int64, by float64) {
	for j, kk := range ks {
		if isWater[kk] {
			fish[j] = min(1, fish[j]+by)
		}
	}
}

// restScalar rests the fields, up to what the ground can hold.
func restScalar(fert, rich []float64, ks []int64, by float64) {
	for j, kk := range ks {
		if isField[kk] {
			fert[j] = min(rich[j], fert[j]+by)
		}
	}
}

package world

// The day's pass over the ground, as flat loops over the layers.
//
// What the weather does to a tile in a day is stated tile by tile in
// grow.go - Ripen, then Replenish - and that is the definition. This is the
// same arithmetic done to a run of tiles at once, a layer at a time, so
// that the pass streams the layers rather than picking each tile's numbers
// out of the map: the kind of every tile in the run is read off the map
// once into a byte, and everything after that is a loop over one slice
// asking that byte whether the tile is the kind it is looking for. It is
// the shape the pass has to be in before the arithmetic can be done several
// tiles at a time.
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

// kinds is how many kinds of ground the pass tells apart: what stands on a
// tile and what it is made of, since between them they say what grows on
// it and what comes back to it. It is asked to fit in a byte.
const kinds = int(Tavern+1) * int(TerrainCount)

var _ [255 - kinds]struct{} // a kind has to fit in a byte

// kindOf is the kind of a tile.
func kindOf(t *Tile) uint8 {
	return uint8(int(t.Structure)*int(TerrainCount) + int(t.Terrain))
}

// ages is which kinds of ground carry something growing, and so get older
// with the weather: the alive table read by kind. isWater and isField are
// which kinds are water and which are field, whatever stands on them, which
// is the question Replenish asks.
var ages, isWater, isField = func() (a, w, f [kinds]bool) {
	for s := 0; s <= int(Tavern); s++ {
		for t := 0; t < int(TerrainCount); t++ {
			k := s*int(TerrainCount) + t
			a[k] = alive[s][t]
			w[k] = Terrain(t) == Water
			f[k] = Terrain(t) == Field
		}
	}
	return
}()

// stocking is one process that fills a stock on one kind of ground: which
// kind, how long the process takes in full, how much of the stock a growing
// day puts on, and the layer the stock is kept in. It is a filling with its
// kind beside it and its span worked out, so that the pass asks the
// ontology nothing.
type stocking struct {
	kind       uint8
	full, rate float64
	stock      func(*Grid) []float64
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
					kind: uint8(s*int(TerrainCount) + t), full: f.p.Full(), rate: f.p.Rate, stock: f.stock,
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
	wear := g.Traffic[lo:hi]
	for i := range wear {
		wear[i] *= by
	}
}

// Grow gives the tiles [lo, hi) k of growing weather: what grows on them
// gets that much older and fills toward what its age accounts for, the
// water gets its fish back and worn fields rest. It is Ripen and then
// Replenish, tile by tile, done as loops over the layers; see the remarks
// at the top of the file. The run is taken a chunk's width at a time, which
// is the run a day hands it, so the kinds of a run fit on the stack.
func (g *Grid) Grow(lo, hi int, k float64) {
	fishBy, fallowBy := FishRegrowth*k, Fallow*k
	for lo < hi {
		n := min(ChunkSide, hi-lo)
		var kind [ChunkSide]uint8
		tiles := g.Tiles[lo : lo+n]
		for j := range tiles {
			kind[j] = kindOf(&tiles[j])
		}
		ks := kind[:n]
		// A stand ages by the weather it gets, not by the calendar; see
		// Ripen. The age is put on before anything reads it.
		age := g.Age[lo : lo+n]
		for j, kk := range ks {
			if ages[kk] {
				age[j] += k
			}
		}
		// Whatever is coming on fills a little further, bounded by the age
		// it has had; see grown. Each filling is a pass of its own over the
		// run, in the order the growing table has them.
		for _, e := range stocked {
			by := e.rate * k
			s := e.stock(g)[lo : lo+n]
			for j, kk := range ks {
				if kk != e.kind {
					continue
				}
				ceiling := 1.0
				if e.full > 0 {
					ceiling = clamp01(age[j] / e.full)
				}
				s[j] = grown(s[j], ceiling, by)
			}
		}
		// And what comes back that is not a stand coming on; see Replenish.
		fish := g.Fish[lo : lo+n]
		for j, kk := range ks {
			if isWater[kk] {
				fish[j] = min(1, fish[j]+fishBy)
			}
		}
		fert, rich := g.Fertility[lo:lo+n], g.Rich[lo:lo+n]
		for j, kk := range ks {
			if isField[kk] {
				fert[j] = min(rich[j], fert[j]+fallowBy)
			}
		}
		lo += n
	}
}

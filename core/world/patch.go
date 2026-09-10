package world

import "lreat/core/entity"

// What kind of ground lies where, at a size worth asking about.
//
// A patch is a small square of the map with one count per kind of ground
// on it. It exists for one question - is there any wood, water, rock
// within so far of here - and that question exists because the commonest
// and dearest thing an agent does is look for ground that is not there.
// Such a search walks every ring out to its radius before it can say so.
// Asked of the patches first, it does not have to be walked at all.
//
// The size is the whole of the tuning. A patch is consulted whole, so the
// ground it answers for is the ground the search would cover rounded up to
// the patch, and every tile of that difference is a search walked for
// nothing. Chunks were tried first, since they were already counting: at
// sixty-four tiles against a radius of forty they answered for five and a
// half times the ground the search covered, said yes wherever a settlement
// had any wood at all in the district, and turned away only four searches
// in a hundred. At sixteen the overshoot is under half again.
//
// It is not a second copy of what the chunks know. The chunks stopped
// counting kinds when this began; they count what is built and what is
// owned, which is what the passes over the ground ask them, and what kind
// of ground a tile is is asked here and nowhere else.

// PatchSide is how many tiles a patch is across and down.
const PatchSide = 16

// patch is how many tiles of each kind of ground one patch holds. It is
// int32 because a patch cannot hold more tiles than it has.
type patch [TerrainCount]int32

// layPatches divides the grid into patches, all counts empty. Recount
// fills them.
func (g *Grid) layPatches() {
	g.PW, g.PH = (g.W+PatchSide-1)/PatchSide, (g.H+PatchSide-1)/PatchSide
	g.patches = make([]patch, g.PW*g.PH)
}

// PatchOf is the patch the tile kept at i is in.
func (g *Grid) PatchOf(i int) int {
	return (i/g.W/PatchSide)*g.PW + (i%g.W)/PatchSide
}

// mark adds d to the patch's tally of the kind of ground at i. It is
// called on either side of every change to what a tile is, as the chunk's
// own counts are: see Grid.Turn.
func (g *Grid) mark(i int, t *Tile, d int32) {
	if t.Terrain < TerrainCount {
		g.patches[g.PatchOf(i)][t.Terrain] += d
	}
}

// repatch takes every patch's counts afresh from the ground.
func (g *Grid) repatch() {
	if len(g.patches) != g.PW*g.PH {
		g.layPatches()
	}
	for i := range g.patches {
		g.patches[i] = patch{}
	}
	for i := range g.Tiles {
		g.mark(i, &g.Tiles[i], 1)
	}
}

// patchAt is the patch p is in. p must be on the map and normalised.
func (g *Grid) patchAt(p entity.Pos) int {
	return (p.Y/PatchSide)*g.PW + p.X/PatchSide
}

// patchHolds reports whether patch i holds any tile of these kinds.
func (g *Grid) patchHolds(i int, kinds KindSet) bool {
	held := &g.patches[i]
	for t := Terrain(0); t < TerrainCount; t++ {
		if kinds.Has(t) && held[t] > 0 {
			return true
		}
	}
	return false
}

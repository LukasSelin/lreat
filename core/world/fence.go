package world

import "lreat/core/entity"

// Fences, and what they do to a journey.
//
// A strip or two of corn is a thing people walk round without being asked to.
// A block of them is a different thing: it is somebody's year, standing up to
// the knee and ruined by being walked through, and what goes round it is a
// fence, a hedge or a ditch. Nobody in here decides to build one. A holding
// large enough to be worth the trouble is enclosed the way a wood large
// enough is a wood: it follows from what is on the ground.
//
// The point of it is the going. Before this, a field cost 1.3 to cross
// against grass at 1, which is to say a shortcut through the corn was very
// nearly the shortest way and every errand took it. That is not what people
// do. So the toll is not on the ground of the field, which is easy enough
// underfoot; it is on the line round it, paid once by whoever climbs in and
// once by whoever climbs out, and it makes the way round a big holding
// cheaper than the way across it without ever making the way across it
// impossible.
const (
	// fenceSize is how many strips have to lie together before the block is
	// worth enclosing. A holding is three strips, so this is two households'
	// worth of ground lying side by side: the size at which a block is
	// several tiles across whichever way it is crossed, which is what makes
	// walking round it a detour worth pricing. One family's three strips are
	// left open, because a fence round them would cost more to go round than
	// it saved anybody, and because nobody hedges a patch they can see across.
	fenceSize = 6
	// fenceToll is the ticks of climbing the fence, in and out. It has to
	// beat the detour round a block of fenceSize or the fence is decoration:
	// such a block is two or three tiles across, so going round it costs two
	// or three tiles of walking, and at 4 the way round wins.
	//
	// It is a toll and not a wall. A walker who has no way round - a farmer
	// whose neighbours' strips lie between the lane and their own, somebody
	// cut off by a river on the other side - climbs over and pays for it,
	// which is what people do. Nothing on this map is ever made unreachable
	// by anything anybody built.
	fenceToll = 4
)

// Fence reads the fields and marks the ones lying in a block large enough to
// be enclosed. It is called once a day, from Land, so a strip broken this
// morning is inside the hedge tomorrow: raising a fence takes a while, and
// the ground it goes round has to be there first.
//
// Blocks are the eight-neighbour kind, the same as everything else that reads
// the map, so two farmers whose holdings touch at a corner are inside one
// fence. That is the honest reading of it - what is enclosed is the block of
// worked ground, not one household's title - and it is why the toll is
// charged on the line and not on the tile.
//
// Only the fields are visited, and only the patches that have any - see
// patch.go - since a hedge stands round a field and nothing else: the
// hedge comes off a tile as the tile stops being a field, in Turn, so there
// is nothing to take off any other ground. A field is somebody's, and
// ground that is somebody's is awake, so the patches are those of awake
// chunks and the fields found are the fields the day would have found
// walking all its ground. What a block comes to does not depend on which
// of its strips it was found from, so the order the patches are taken in
// is not a fact about anything. The ground already answered for is marked
// with the day's stamp rather than cleared and marked afresh, which on a
// globe was a walk over the whole map for a few dozen strips of corn.
func (g *Grid) Fence() {
	n := len(g.Tiles)
	if len(g.fenceSeen) != n {
		g.fenceSeen, g.fenceGen = make([]uint32, n), 0
	}
	g.fenceGen++
	if g.fenceGen == 0 { // the stamp has come round: start it over
		clear(g.fenceSeen)
		g.fenceGen = 1
	}
	gen, seen := g.fenceGen, g.fenceSeen
	// The patches with any field on awake ground, in patch order, and then
	// the fields of each in the order the patch is walked in. Finding them
	// is a walk over every tile of every such patch for the few that are
	// fields, and was nearly the whole cost of the day's fences; it reads
	// the ground and writes a list of its own patch, so the patches are
	// found side by side. The blocks are then read off the lists in the
	// one order the walk always took.
	patches := g.fencePatches[:0]
	for pi := range g.patches {
		if g.patches[pi][Field] == 0 || !g.Awake(g.chunkOfPatch(pi)) {
			continue
		}
		patches = append(patches, int32(pi))
	}
	g.fencePatches = patches
	if len(g.fenceFields) != len(g.patches) {
		g.fenceFields = make([][]int32, len(g.patches))
	}
	fields := g.fenceFields
	InParallel(len(patches), WorkersFor(len(patches)/fencePatchesEach), func(k, _ int) {
		pi := int(patches[k])
		found := fields[pi][:0]
		g.eachInPatch(pi, func(i int, t *Tile) {
			if t.Terrain == Field {
				found = append(found, int32(i))
			}
		})
		fields[pi] = found
	})
	block := g.fenceBlock[:0]
	stack := g.fenceStack[:0]
	for _, pi := range patches {
		for _, i := range fields[pi] {
			if seen[i] == gen {
				continue // already answered for with the rest of its block
			}
			// Everything that lies with this strip, found in a fixed order so
			// that the same map always gives the same blocks.
			block, stack = block[:0], append(stack[:0], int32(i))
			seen[i] = gen
			for len(stack) > 0 {
				j := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				block = append(block, j)
				p := g.PosOf(int(j))
				for d := range dirs {
					q := entity.Pos{X: p.X + dirs[d].X, Y: p.Y + dirs[d].Y}
					if !g.In(q) {
						continue
					}
					k := int32(g.Index(q))
					if seen[k] == gen || g.Tiles[k].Terrain != Field {
						continue
					}
					seen[k] = gen
					stack = append(stack, k)
				}
			}
			enclosed := len(block) >= fenceSize
			for _, j := range block {
				g.Tiles[j].Fenced = enclosed
			}
		}
	}
	g.fenceBlock, g.fenceStack = block[:0], stack[:0]
}

// fencePatchesEach is how many patches of fields each goroutine finding
// them is worth: a patch is walked in about a microsecond, and handing one
// out costs a few, so a valley's dozen are found in turn and a globe's
// hundreds are dealt out.
const fencePatchesEach = 16

// fenceCost is what crossing the line between two tiles costs. It is paid
// where one side is inside a fence and the other is not, and it is not paid
// by the holder of the ground: a farmer has a gate into their own field, and
// charging them to reach the strip they live off would only have made
// farming dearer than it is.
func (g *Grid) fenceCost(from, to int32, holder entity.ID) float64 {
	f, t := &g.Tiles[from], &g.Tiles[to]
	if f.Fenced == t.Fenced {
		return 0 // both in the same enclosure, or both outside one
	}
	enclosed := f
	if t.Fenced {
		enclosed = t
	}
	if holder != 0 && enclosed.Owner == holder {
		return 0
	}
	return fenceToll
}

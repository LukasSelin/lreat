package world

import (
	"fmt"

	"lreat/core/entity"
	"lreat/core/ontology"
)

// releaseRange is how far from the square the creatures a world is made
// with are put down. It keeps them in the settlement's own country, which
// on a valley is the whole map and on a globe is the one part of it anybody
// is watching - and, since a creature standing on a chunk keeps that chunk
// awake, the one part of it the day is already spent on.
const releaseRange = 60

// habitats is where each kind of creature is put down: a deer and a boar
// under trees, a hare on open ground with trees beside it.
var habitats = map[*entity.Species]func(g *Grid, p entity.Pos, t *Tile) bool{
	entity.Deer: func(_ *Grid, _ entity.Pos, t *Tile) bool { return t.Is(ontology.Wood) },
	entity.Boar: func(_ *Grid, _ entity.Pos, t *Tile) bool { return t.Is(ontology.Wood) },
	entity.Hare: func(g *Grid, p entity.Pos, t *Tile) bool {
		return t.Is(ontology.Open) && t.Structure == None && g.HasNeighbor(p, func(n *Tile) bool { return n.Is(ontology.Wood) })
	},
}

// Populate puts on the ground the creatures the world was configured with,
// each kind in its own habitat around the settlement. It is called once the
// founders stand, and it draws nothing when there is nothing to put down,
// so a world with no creatures in it draws exactly the chance it always
// drew, and a world with them draws its own after everything the founding
// drew. The kinds are put down in the order they are declared.
func (w *World) Populate() {
	for _, sp := range entity.Creatures {
		// A kind is released as a herd and not scattered: the first is
		// put down anywhere in its habitat within range of the square, and
		// the rest within herd reach of the last, so that from the first
		// day each has its kind for company. Six boar put down sixty tiles
		// apart never found one another, belonged to nothing, and bore
		// nothing.
		var last entity.Pos
		for i := 0; i < w.Config.Of(sp); i++ {
			var p entity.Pos
			var ok bool
			if i == 0 {
				p, ok = w.covert(sp, w.MarketPos, releaseRange)
			} else {
				p, ok = w.covert(sp, last, sp.Herds)
			}
			if !ok {
				break
			}
			w.SpawnKind(sp, fmt.Sprintf("%s-%d", sp.Name, i+1), w.RandomPersonality(), p)
			last = p
		}
	}
}

// covert is somewhere in the kind's habitat within reach of a place: a few
// draws at random, and the nearest such ground to the place when none of
// them lands on any.
func (w *World) covert(sp *entity.Species, near entity.Pos, reach int) (entity.Pos, bool) {
	g := w.Grid
	fit := habitats[sp]
	if fit == nil {
		return entity.Pos{}, false
	}
	for try := 0; try < 40; try++ {
		p := g.Norm(entity.Pos{
			X: near.X + w.RNG.IntN(2*reach+1) - reach,
			Y: near.Y + w.RNG.IntN(2*reach+1) - reach,
		})
		if g.In(p) && fit(g, p, g.At(p)) {
			return p, true
		}
	}
	return g.Nearest(near, reach, func(p entity.Pos, t *Tile) bool { return fit(g, p, t) })
}

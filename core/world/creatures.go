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

// Populate puts on the ground the creatures the world was configured with:
// its deer, in the woods around the settlement. It is called once the
// founders stand, and it draws nothing when there is nothing to put down,
// so a world with no creatures in it draws exactly the chance it always
// drew, and a world with them draws its own after everything the founding
// drew.
func (w *World) Populate() {
	for i := 0; i < w.Config.Deer; i++ {
		p, ok := w.covert()
		if !ok {
			return
		}
		w.SpawnKind(entity.Deer, fmt.Sprintf("deer-%d", i+1), w.RandomPersonality(), p)
	}
}

// covert is somewhere wooded within the release range of the square: a few
// draws at random, and the nearest wood to the square when none of them
// lands under trees.
func (w *World) covert() (entity.Pos, bool) {
	g := w.Grid
	for try := 0; try < 40; try++ {
		p := g.Norm(entity.Pos{
			X: w.MarketPos.X + w.RNG.IntN(2*releaseRange+1) - releaseRange,
			Y: w.MarketPos.Y + w.RNG.IntN(2*releaseRange+1) - releaseRange,
		})
		if g.In(p) && g.At(p).Is(ontology.Wood) {
			return p, true
		}
	}
	return g.NearestOfKind(w.MarketPos, releaseRange, KindsOf(ontology.Wood), func(_ entity.Pos, t *Tile) bool {
		return t.Is(ontology.Wood)
	})
}

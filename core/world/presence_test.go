package world

import (
	"math/rand/v2"
	"testing"

	"lreat/core/entity"
	"lreat/core/ontology"
)

// AnyWithin is allowed to say yes when there is nothing - it reads whole
// chunks, and a chunk the radius only clips is counted as reached - but a
// no has to mean there is truly nothing, because a no is taken for an
// answer and the search is not walked. This holds it against the walk.
func TestAnyWithinNeverMissesGroundThatIsThere(t *testing.T) {
	for _, wrap := range []bool{false, true} {
		width := 200
		if wrap {
			width = 192 // a globe is a whole number of chunks round
		}
		w := NewWith(13, Config{Width: width, Height: 128, Wrap: wrap})
		g := w.Grid
		rng := rand.New(rand.NewPCG(5, 6))
		sets := []KindSet{
			Kinds(Forest), Kinds(Water), Kinds(Rock), Kinds(Field),
			Kinds(Forest, Water), KindsOf(ontology.Wood), KindsOffering(ontology.Stone),
		}
		says, truly := 0, 0
		for q := 0; q < 3000; q++ {
			p := entity.Pos{X: rng.IntN(g.W), Y: rng.IntN(g.H)}
			radius := 1 + rng.IntN(40)
			kinds := sets[rng.IntN(len(sets))]
			there := false
			for dy := -radius; dy <= radius && !there; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					q := entity.Pos{X: p.X + dx, Y: p.Y + dy}
					if !g.In(g.Norm(q)) || q.Y < 0 || q.Y >= g.H {
						continue
					}
					if kinds.Has(g.At(g.Norm(q)).Terrain) {
						there = true
						break
					}
				}
			}
			said := g.AnyWithin(p, radius, kinds)
			if there && !said {
				t.Fatalf("wrap %v: within %d of %v there is ground of %b, and the counts say there is none",
					wrap, radius, p, kinds)
			}
			if said {
				says++
			}
			if there {
				truly++
			}
		}
		// If it said yes to everything it would be sound and useless.
		if says == truly {
			t.Logf("wrap %v: %d of 3000 said yes, %d truly had some", wrap, says, truly)
		}
		if says >= 3000 {
			t.Fatalf("wrap %v: the counts said yes to every one of 3000 questions", wrap)
		}
	}
}

// An empty set is nothing to look for, and nothing is never anywhere.
func TestAnyWithinOfNothingIsNo(t *testing.T) {
	w := New(2)
	if w.Grid.AnyWithin(entity.Pos{X: 4, Y: 4}, 40, 0) {
		t.Fatal("the counts found ground of no kind at all")
	}
}

// countKind walks the tiles, which is what the patches are there to save.
func countKind(g *Grid, t Terrain) int {
	n := 0
	for i := range g.Tiles {
		if g.Tiles[i].Terrain == t {
			n++
		}
	}
	return n
}

// The patches are the only tally of what kind of ground the map holds
// where, so a tally that drifts from the ground does not show up as a
// wrong number anywhere - it shows up as a search that is never walked
// because the counts said there was nothing to find. This holds them
// against the ground as it is made and as it is turned.
func TestPatchesAgreeWithTheGround(t *testing.T) {
	for _, wrap := range []bool{false, true} {
		width := 200
		if wrap {
			width = 192
		}
		w := NewWith(17, Config{Width: width, Height: 128, Wrap: wrap})
		g := w.Grid
		check := func(when string) {
			for k := Terrain(0); k < TerrainCount; k++ {
				if got, want := g.Kind(k), countKind(g, k); got != want {
					t.Fatalf("wrap %v, %s: the patches count %d of %v, the ground has %d",
						wrap, when, got, k, want)
				}
			}
		}
		check("as made")
		rng := rand.New(rand.NewPCG(9, 10))
		for i := 0; i < 500; i++ {
			p := entity.Pos{X: rng.IntN(g.W), Y: rng.IntN(g.H)}
			g.Turn(p, Terrain(rng.IntN(int(TerrainCount))))
		}
		check("after turning five hundred tiles")
		g.Recount()
		check("after a recount")
	}
}

// Clone is how a snapshot is taken, and a snapshot whose patches were
// empty would report a map with no ground of any kind on it.
func TestACloneCarriesItsPatches(t *testing.T) {
	w := New(6)
	c := w.Grid.Clone()
	for k := Terrain(0); k < TerrainCount; k++ {
		if got, want := c.Kind(k), w.Grid.Kind(k); got != want {
			t.Fatalf("the copy counts %d of %v, the original %d", got, k, want)
		}
	}
}

// NearestOfKind steps over ground that cannot hold the answer. It must
// come back with the very tile Nearest would have come back with - not
// merely a tile that satisfies the predicate, but the same one, since ties
// between equally near tiles are settled by the order the rings are walked
// in and a settlement built on the other one is a different settlement.
func TestNearestOfKindFindsWhatNearestFinds(t *testing.T) {
	for _, wrap := range []bool{false, true} {
		width := 200
		if wrap {
			width = 192
		}
		w := NewWith(23, Config{Width: width, Height: 128, Wrap: wrap})
		g := w.Grid
		rng := rand.New(rand.NewPCG(11, 12))
		// Each of these is true only of ground of the kinds beside it.
		cases := []struct {
			kinds KindSet
			ok    func(p entity.Pos, t *Tile) bool
		}{
			{Kinds(Forest), func(_ entity.Pos, t *Tile) bool { return t.Terrain == Forest }},
			{Kinds(Water), func(_ entity.Pos, t *Tile) bool { return t.Terrain == Water }},
			{Kinds(Rock), func(_ entity.Pos, t *Tile) bool { return t.Terrain == Rock }},
			{KindsOffering(ontology.Timber), func(_ entity.Pos, t *Tile) bool { return t.Offers(ontology.Timber) >= 0.3 }},
			{KindsOffering(ontology.Stone), func(_ entity.Pos, t *Tile) bool { return t.Offers(ontology.Stone) > 0 }},
			// One that is true of only some of its kind, so the search
			// cannot stop at the first patch that holds any.
			{Kinds(Forest), func(p entity.Pos, t *Tile) bool { return t.Terrain == Forest && p.X%3 == 0 }},
		}
		for q := 0; q < 4000; q++ {
			p := entity.Pos{X: rng.IntN(g.W), Y: rng.IntN(g.H)}
			radius := 1 + rng.IntN(40)
			c := cases[rng.IntN(len(cases))]
			want, wantOK := g.Nearest(p, radius, c.ok)
			got, gotOK := g.NearestOfKind(p, radius, c.kinds, c.ok)
			if gotOK != wantOK || got != want {
				t.Fatalf("wrap %v: nearest to %v within %d: told the kinds it found %v/%v, told nothing %v/%v",
					wrap, p, radius, got, gotOK, want, wantOK)
			}
		}
	}
}

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

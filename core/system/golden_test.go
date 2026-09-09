package system

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"lreat/core/world"
)

// golden is what three settlements on the default map came to after fifteen
// hundred days, as a hash of everything about them: where everyone is and
// what shape they are in, every tile of the ground, and how much has
// happened. It was taken before the world was made big enough to hold more
// than one settlement, and it is what every step of making it so is held
// against. It was taken again when master was merged in, on master itself,
// and the merged tree came to the same three numbers. The full batch in docs/baseline.md is the proof; this is the
// check that runs in seconds.
//
// A change meant to alter what a settlement does must retake these three
// numbers in the commit that makes it, and say so. A change that was not
// meant to and moves them was not what it was meant to be.
//
// These numbers are a hash of every field of every tile, so a change to what
// a tile *is* moves them even when nothing that happens on one has changed.
// That is what the plate and the epoch a rock dates from did: a drawn world
// is untouched to the last bit - its heights, its drainage and its soils all
// check out identical against the commit before - and these three numbers
// moved anyway because there are two more fields in the struct being hashed.
// When that happens, prove it the way it was proved then: sum the heights,
// the drainage and the fertility of a few seeds on both trees and compare.
//
// Retaken for the rivers: a channel now cuts the outside of its own bends and
// walks sideways across its valley, and what an age of weather takes off a
// tile is divided by the rock under it, so soft beds go and hard ones are
// left standing. Both move every map from the first age onward.
//
// Retaken before that for the soil's make-up: there is rock under the ground now, the
// soil over it is a mixture of sand, silt and clay weathered out of that
// rock, and the water sorts what it carries so that the fine stuff ends up
// where the water slows. Fertility reads the mixture and so does how fast a
// hillside washes, which moves every map from the moment it is made; the full
// batch is in docs/baseline.md.
//
// Retaken before that for the lapse rate: the weather falls with the height
// of the ground, so the cold a body feels and the growing weather the ground
// gets are read where they are rather than off the row.
var golden = map[uint64]string{
	1: "c984d655f4906307",
	3: "4a34be73e4103abb",
	9: "e45e762d53c4ed7b",
}

// digest is the hash the golden numbers are of.
func digest(w *world.World) string {
	h := sha256.New()
	fmt.Fprint(h, fingerprint(w))
	for i := range w.Grid.Tiles {
		fmt.Fprintf(h, "%v", w.Grid.Tiles[i])
	}
	fmt.Fprintf(h, "%d", w.Log.Len())
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func TestTheSameSeedStillMakesTheSameSettlement(t *testing.T) {
	for seed, want := range golden {
		seed, want := seed, want
		t.Run(fmt.Sprintf("seed%d", seed), func(t *testing.T) {
			t.Parallel()
			w := world.New(seed)
			for i := 0; i < 30; i++ {
				w.Spawn("a", w.RandomPersonality())
			}
			Run(w, 1500)
			if got := digest(w); got != want {
				t.Fatalf("seed %d came to %s after 1500 days; the tree this was taken on came to %s", seed, got, want)
			}
		})
	}
}

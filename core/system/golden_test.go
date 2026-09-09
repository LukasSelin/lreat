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
// Retaken for the lapse rate: the weather now falls with the height of the
// ground, so the cold a body feels and the growing weather the ground gets
// are read where they are rather than off the row. On the valley that is two
// degrees between the river and the highest ground, which moves the run
// without changing what it is; the full batch is in docs/baseline.md.
var golden = map[uint64]string{
	1: "8655d764af56fdab",
	3: "d98fdb39def62c60",
	9: "e6dcfeae58cd3759",
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

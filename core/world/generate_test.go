package world

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

// Making a world is mostly the same arithmetic done half a million times -
// the ground drawn, then measured, then read - and most of those passes are
// spread over goroutines. What comes out must not know that.
//
// The chance is what makes this worth saying. A seed means the order the
// world's chance comes out in, and the passes that draw - the corners of a
// lattice, the luck thrown on a tile's suitability for trees - are kept in
// walks of their own with only the arithmetic spread. If any draw ever
// wandered into a spread pass, these hashes would part company on the
// second run.
func digest(w *World) string {
	h := sha256.New()
	for i := range w.Grid.Tiles {
		fmt.Fprintf(h, "%v", w.Grid.Tiles[i])
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func madeOver(t *testing.T, seed uint64, cfg Config, workers int) string {
	t.Helper()
	was := Workers
	Workers = workers
	defer func() { Workers = was }()
	return digest(NewWith(seed, cfg))
}

func TestMakingAWorldDoesNotDependOnTheGoroutines(t *testing.T) {
	worlds := map[string]Config{
		"globe":   Globe(),
		"ancient": Ancient(),
		"valley":  DefaultConfig(),
	}
	// Not in parallel with one another: what is being varied is a package
	// variable, and two subtests varying it at once would be measuring each
	// other. Making a globe five times is a few seconds, which is what this
	// costs and what it is worth.
	for name, cfg := range worlds {
		t.Run(name, func(t *testing.T) {
			one := madeOver(t, 1, cfg, 1)
			for _, workers := range []int{2, 3, 8, 16} {
				if got := madeOver(t, 1, cfg, workers); got != one {
					t.Fatalf("%s: %d goroutines made a different world than 1 did", name, workers)
				}
			}
		})
	}
}

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
// Retaken for what a discovery asks of a settlement: a technology used to be
// unlocked by a count of people over a level, and is now unlocked by people
// who have both reached a tier and done the work the tier is in. Agriculture
// asked for two people at a fifth of farming, in settlements whose best
// farmer had broken ground five times in sixty years; it asks for two
// apprentices with eighty days on the ground. See system.adept. It
// moves every settlement from the first discovery on, and the ones it moves
// furthest are the ones that were living off a technology they never earned.
//
// Retaken before that for the moral coordinates: all twenty dimensions of a habit drift
// per agent now, where the five moral ones used to be held fixed across the
// population. Five more draws off every founder's own luck at imprinting and
// off the world's at every birth, so these move whatever the drift does.
//
// Retaken before that for how competence is got: learning curves rather than
// stepping, each tier of it slower than the one under it, and what a person
// is shown stops short of what the person showing them knows. Being taught
// was a flat step from any teacher at any level, so a settlement rose at the
// rate of its best member and everybody arrived at mastery together; now only
// the work reaches the last tier and the settlement has to make its masters
// one at a time. See core/entity/learn.go. It moves every settlement from the
// first lesson given.
//
// Retaken before that for bodies and minds: what an agent burns, what cold it can stand,
// how fast it takes to a craft, how firmly it decides and how far it will go
// were one number apiece for the whole population and are now drawn per agent
// and inherited. Five more draws at every spawn move the world's chance from
// the first tick, so these had to move whatever the traits did. The land did
// not: all six map hashes are as they were, because a world is made before
// anybody is put on it. See entity.Body.
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
	1: "1a4878c5c4fec607",
	3: "74fcf6fd0048f38d",
	9: "39d590378ddae2f9",
}

// digest is the hash the golden numbers are of.
//
// Each tile is written out field by field, in the order and the form fmt
// gives a Tile printed whole, so that where a field is kept - on the tile
// or in a layer beside it, see Layers - is not part of what is hashed. It
// was fmt's own rendering of the struct, and the numbers were taken on that;
// this writes the same bytes, and the test proves it by not having moved.
func digest(w *world.World) string {
	h := sha256.New()
	fmt.Fprint(h, fingerprint(w))
	g := w.Grid
	for i := range g.Tiles {
		t := &g.Tiles[i]
		fmt.Fprintf(h, "{%v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v}",
			t.Terrain, t.Structure, t.Owner, g.Fertility[i], g.Rich[i], t.Wood, t.Wild, t.Fish,
			t.Height, t.Flow, t.Drain, t.Bedrock, t.Sand, t.Clay, t.Plate, t.Formed,
			g.Age[i], t.Fenced, g.Traffic[i])
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

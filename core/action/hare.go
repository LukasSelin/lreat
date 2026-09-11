package action

import (
	"lreat/core/entity"
	"lreat/core/ontology"
)

// A hare grazes the sward on open ground and lives at the edge of the wood,
// which is its cover: it crouches, dashes from anybody near, keeps to its
// warren, and lopes back toward the trees.
var (
	Graze  = feeding("graze", ontology.Sward, ontology.Open)
	Crouch = bedding("crouch")
	Dash   = fleeing("dash")
	Warren = herding("warren")
	Lope   = roaming("lope")
)

func init() { wary(entity.Hare, Graze, Crouch, Dash, Warren, Lope, feedPrior) }

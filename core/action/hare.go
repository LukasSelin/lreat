package action

import (
	"lreat/core/entity"
	"lreat/core/ontology"
)

// A hare grazes the sward on open ground and lives at the edge of the wood,
// which is its cover: it crouches, dashes from anybody near, keeps to its
// warren, and lopes back toward the trees. What a warren grazes it keeps
// open, since a wood's seed takes only among shoots, and what it grazes it
// manures, so the meadow at a wood's edge is a warren's making.
var (
	Graze  = feeding("graze", ontology.Sward, ontology.Open, browseTake, manured)
	Crouch = bedding("crouch")
	Dash   = fleeing("dash")
	Warren = herding("warren")
	Lope   = roaming("lope")
)

func init() { wary(entity.Hare, Graze, Crouch, Dash, Warren, Lope, feedPrior) }

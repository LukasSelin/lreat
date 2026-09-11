package action

import (
	"lreat/core/entity"
	"lreat/core/ontology"
)

// A deer browses the brush under the trees, beds down, runs from anybody
// who comes near, goes to its kind when it has lost the herd, and rambles.
// See creature.go for what each of these is.
var (
	Browse  = feeding("browse", ontology.Browse, ontology.Wood, browseTake, browsed)
	BedDown = bedding("bed down")
	Flee    = fleeing("flee")
	Herd    = herding("herd")
	Roam    = roaming("roam")
)

func init() { wary(entity.Deer, Browse, BedDown, Flee, Herd, Roam, feedPrior) }

package ontology

// Affords says what a site has to give. Take instantiates one act for every
// (material, site) pair listed here and none otherwise: a map with no
// outcrop has no quarrying.
var Affords = map[*Class][]*Class{
	Wood:    {Berries, Game, Timber},
	Water:   {Fish},
	Outcrop: {Stone},
	Field:   {Grain},
}

// Yields reports whether site has material to give.
func Yields(site, material *Class) bool {
	for _, m := range Affords[site] {
		if material.IsA(m) {
			return true
		}
	}
	return false
}

// Transform is a change that happens to a thing rather than one somebody
// does to it: food spoils, a field left alone goes back to grass. It is the
// world's half of the ontology. Transforms never instantiate as actions and
// never take a habit slot; they belong to upkeep. To is nil when the thing
// is simply gone. Unless names a site that arrests the change.
type Transform struct {
	From, To *Class
	Rate     float64 // share per tick
	Unless   *Class
}

// Transforms lists what happens on its own. Rates are placeholders to be
// read off the existing upkeep once this replaces it; for now the list is
// the shape, not the numbers.
var Transforms = []Transform{
	{From: Meal, Rate: 0.05},
	{From: Berries, Rate: 0.02, Unless: Granary},
	{From: Game, Rate: 0.03, Unless: Granary},
	{From: Fish, Rate: 0.03, Unless: Granary},
	{From: Grain, Rate: 0.005, Unless: Granary},
	{From: Field, To: Open, Rate: 0.01},
}

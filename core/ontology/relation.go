package ontology

// Holds says what kinds of thing hold what materials: the ground holds
// what lies in it, a person holds what is in their pack and their purse,
// the market holds what is on its shelves. Every act that moves a material
// moves it between two holders named by class, and Take instantiates one
// act for every (material, ground) pair here and none otherwise: a map
// with no outcrop has no quarrying.
var Holds = map[*Class][]*Class{
	Wood:    {Berries, Game, Timber},
	Water:   {Fish},
	Outcrop: {Stone},
	Field:   {Grain},
	Person:  {Material},
	Market:  {Material},
}

// Held reports whether a holder of class c holds material.
func Held(c, material *Class) bool {
	for _, m := range Holds[c] {
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

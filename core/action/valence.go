package action

import "lreat/core/belief"

// valences say what each act is worth morally. This is a property of the act
// itself; how much it matters depends entirely on the norms of whoever is
// doing the judging.
//
// It is kept as a table beside the catalog rather than a field on Def so that
// the moral layer can be read, tuned, and argued about in one place.
var valences = map[string]belief.Valence{
	// Idleness offends the industrious.
	"rest": {belief.Industry: -0.4},

	// Ordinary work is mildly virtuous to them.
	"farm":        {belief.Industry: 0.3},
	"clear field": {belief.Industry: 0.3},
	"gather wood": {belief.Industry: 0.25},
	"craft":       {belief.Industry: 0.3},
	"forage":      {belief.Industry: 0.2},

	"build shelter": {belief.Industry: 0.3, belief.Tradition: 0.1},
	"sell":          {belief.Industry: 0.15},
	"buy food":      {},
	"eat":           {},

	// Serving the settlement reads as both work and generosity.
	"guard":     {belief.Industry: 0.3, belief.Charity: 0.4},
	"lay road":  {belief.Industry: 0.4, belief.Charity: 0.3},
	"socialize": {belief.Charity: 0.1, belief.Tradition: 0.15},
	"teach":     {belief.Charity: 0.5, belief.Industry: 0.2, belief.Tradition: 0.3},

	// Enquiry unsettles the traditional. A settlement that prizes continuity
	// will suppress its own scholars without ever forbidding anything.
	"study": {belief.Industry: 0.2, belief.Tradition: -0.5},

	// The morally loaded acts the rest of the system turns on.
	"steal":          {belief.Honesty: -1, belief.Charity: -0.4},
	"give":           {belief.Charity: 0.8, belief.Honesty: 0.1},
	"fulfil request": {belief.Charity: 0.5, belief.Industry: 0.35},

	// Retaliation is rough justice. To the honest it answers a wrong and to
	// the traditional it keeps order; to the charitable it is harm done.
	"retaliate": {belief.Honesty: 0.6, belief.Charity: -0.4, belief.Tradition: 0.3},
}

// ValenceOf returns the moral weight of an act. Unlisted acts are neutral.
func ValenceOf(name string) belief.Valence { return valences[name] }

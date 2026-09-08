package world

import "lreat/core/ontology"

// What grows on a tile takes time to come on, and that time is not the same
// for everything growing. Timber is the slow one: a stand a planter raised is
// firewood for a lifetime before it is beams. The brush under it - the
// berries and the game that live off them - comes back within a few years of
// a clearing being made. A crop is the fast one: sown, in ear within the
// quarter, and gone again the moment it is cut.
//
// These are growing days rather than days: what is measured is how much
// growing weather a stand has had, so a wood raised in the autumn stands
// still until the thaw. See system.Land, which advances them.
//
// The map does not decide them. How long a thing takes to come on is a fact
// about the thing and not about the ground it happens to be standing on, so
// the ontology holds it - see ontology.Class.Comes - and these are the three
// the map has occasion to ask for. What a wood is worth waiting for is the
// same answer wherever the wood is.
//
// They are what the regrowth rates in system.Land were tuned against, and
// the pair has to be read together: an age says what a stand may hold, a
// rate says how fast it fills toward it, and it is the slower of the two
// that a settlement actually feels. Lengthening the year did not lengthen
// these, and that is deliberate - the day is what the land renews by and
// the day is what people take by, so the balance between them is the one
// thing the calendar must not touch.
var (
	TimberAge = float64(ontology.Timber.Ripens())
	BrushAge  = float64(ontology.Berries.Ripens())
	CropAge   = float64(ontology.Grain.Ripens())
)

// living is where the ontology's Living sites are to be found on the map. It
// is the whole of the binding between the two: the ontology says which ground
// carries something alive, and this says which terrain that is. A terrain
// missing from here has no stand on it, and its stocks - the water's fish,
// an outcrop's stone - are not waited on.
var living = map[Terrain]*ontology.Class{
	Forest: ontology.Wood,
	Field:  ontology.Field,
}

// Alive reports whether this tile carries a standing crop, which is to say
// something that had to grow before it could be taken.
func (t *Tile) Alive() bool {
	_, ok := living[t.Terrain]
	return ok
}

// Grown is how far along what grows here is, in [0,1], against the time such
// a thing takes to come on. It only rises: a wood that has made its timber
// holds it, and a stand does not go over and take the wood with it. What
// starts it again is the ground being cleared and something else sown on it.
func (t *Tile) Grown(full float64) float64 {
	if full <= 0 {
		return 1
	}
	return clamp01(t.Age / full)
}

// Sow starts whatever is to grow here over: the ground is bare, and what
// stands on it from now on is this year's, not last year's. It is called
// wherever the terrain changes hands - a wood seeded or planted, a wood
// felled to a clearing, a strip broken, a strip harvested, a road laid over
// any of them - so that nothing inherits the age of what it replaced.
func (t *Tile) Sow() { t.Age = 0 }

// Standing puts a tile's growth at full, for ground that is meant to have
// been there all along: the woods a map is made with are old woods.
func (t *Tile) Standing() { t.Age = TimberAge }

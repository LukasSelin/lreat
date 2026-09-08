package world

import "lreat/core/ontology"

// What grows on a tile takes time to come on, and that time is not the same
// for everything growing. How long each thing takes, and what stages it
// passes through on the way, is stated in the ontology - see
// ontology.Processes. This is where those stages are read off an actual
// tile, and system.Land is what advances them.
//
// These are growing ticks rather than ticks: what is measured is how much
// growing weather a stand has had, so a wood raised in the autumn stands
// still until the thaw.

// alive is the same question read as a table over the ground, worked out
// once from the same bindings rather than asked of the ontology again. Every
// tile on the map is asked whether it is alive on every tick, and gathering
// the processes that run on it to answer allocates the gathering.
var alive = func() (a [Tavern + 1][Rock + 1]bool) {
	for s := range a {
		for t := range a[s] {
			tile := Tile{Structure: Structure(s), Terrain: Terrain(t)}
			a[s][t] = len(ontology.Growing(ClassOf(&tile))) > 0
		}
	}
	return a
}()

// Alive reports whether this tile carries a standing crop, which is to say
// something that had to grow before it could be taken. It is exactly the
// tiles some process runs on: the ontology says which ground carries
// something alive, and ClassOf says which terrain that is.
func (t *Tile) Alive() bool { return alive[t.Structure][t.Terrain] }

// Along is how far through process p what stands here has come, in [0,1].
// It is the reading a stage is asked for against, and it is p's share of
// the tile's one age: a wood is coming on as brush and as timber at once,
// and is far further along the first than the second.
func (t *Tile) Along(p *ontology.Process) float64 {
	return t.Grown(p.Full())
}

// Reached reports whether what stands here has come as far as ph. It is
// what lets an act ask for a stage - a harvest wants a field in ear - and
// it is written as the share of the whole rather than as an age in ticks
// because the two disagree in the last bit, and one field flipping on one
// tick re-rolls every draw in the run after it.
func (t *Tile) Reached(ph ontology.Phase) bool {
	return t.Along(ph.Process) >= ph.Share()
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
func (t *Tile) Standing() { t.Age = ontology.Timbering.Full() }

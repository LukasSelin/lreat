package ontology

import "lreat/core/clock"

// What happens on its own.
//
// The rest of this package states what an agent can do and to what. This
// states the other half: what the world does whether or not anybody is
// attending to it. A crop comes on while the farmer is asleep, a wood the
// planter died before seeing is somebody else's timber, and a house nobody
// is left to keep falls in. None of it is anyone's action, none of it takes
// a habit slot, and nothing has to decide for any of it to happen.
//
// There are two kinds, and they are kept apart on purpose. A Process is a
// maturing: monotone, on a clock, and readable in advance, which is what
// lets an act ask for it - a harvest wants a field in ear. A Transform is a
// hazard or an attrition: a rate, whose whole point is that it cannot be
// read in advance. Forcing them into one declaration would give every entry
// a dead Rate or a dead set of Stages, so they stay two.
//
// See system.ripen, system.wither and system.spoil, which run them, and
// world.ClassOf, which says what a tile is in these terms.

// Year is the length of a year, and Season a quarter of it. They are named
// here because what grows is measured in them - a crop takes a month, brush
// a couple of years, timber six - and the ontology is where what grows is
// stated. The lengths themselves are the calendar's: package clock, where a
// tick is a day. The world keeps the same one; see world.Year.
const (
	Year   = clock.Year
	Season = clock.Season
)

// Stage is one named span of a process. Ticks is how much growing weather
// it takes, not how many ticks pass: a wood raised in the autumn stands
// still until the thaw. A stage has no inputs and consumes nothing, because
// nothing here does - what wears a field out is the harvest taken off it,
// which is an act, not the growing.
type Stage struct {
	Name  string
	Ticks float64
}

// Process is a change a thing goes through on its own, in stages, over
// time. Of is the class it happens to and Yields what it brings on. Rate is
// how much of a full stock the stand puts on per growing tick, where what
// grows is a quantity to be taken; a crop has none, because a field is cut
// once and wholly rather than drawn down.
//
// Two processes may run on one thing - a wood is making brush and timber at
// once, on different clocks - which is why how far along it is has to be a
// reading of the thing's one age rather than a record kept per process.
type Process struct {
	Name   string
	Of     *Class
	Yields *Class
	Rate   float64
	Stages []Stage
}

// Full is how much growing weather the whole process takes.
func (p *Process) Full() float64 {
	var sum float64
	for _, s := range p.Stages {
		sum += s.Ticks
	}
	return sum
}

// Phase names a point a process has reached, so that an act can ask for it.
// It is the ontology's half of "a harvest wants a ripe field"; where the age
// is kept is the world's half, and it is world.Tile.Reached that answers.
type Phase struct {
	Process *Process
	Stage   int
}

// Share is how far into the process this phase begins, in [0,1].
func (p Phase) Share() float64 {
	var before float64
	for _, s := range p.Process.Stages[:p.Stage] {
		before += s.Ticks
	}
	return before / p.Process.Full()
}

// Phase finds a stage of p by name. It panics on a name p has no stage of,
// because every caller is a package variable and a mistake there is a
// mistake in the trees, not in a run.
func (p *Process) Phase(name string) Phase {
	for i, s := range p.Stages {
		if s.Name == name {
			return Phase{p, i}
		}
	}
	panic("ontology: " + p.Name + " has no stage " + name)
}

// What grows, and how long it takes.
//
// Timber is the slow one: a stand a planter raised is firewood for a
// lifetime before it is beams. The brush under it - the berries and the game
// that live off them - comes back within a few years of a clearing being
// made, so a young stand feeds a forager long before it is worth felling. A
// crop is the fast one: sown, in ear within the quarter, and gone again the
// moment it is cut.
//
// A crop is the one thing here with a stage boundary that anybody consults,
// and it is what makes a holding worth having. A strip is not worth cutting
// until it is in ear; a strip cut is bare ground again and has to come on
// before it can be cut a second time. So a household with one strip waits
// and a household with three works them in turn. That is the whole of crop
// rotation, and nobody had to be told it: the ground says when, and the size
// of the holding says how often. Where the line sits does not change what a
// strip gives over a year - the yield is the growth, so half a crop taken
// twice as often comes to the same bread - it decides what a holding is for.
//
// The spans are in days of growing weather, and they did not lengthen when
// the year did. That is deliberate and it is the one thing the calendar was
// not allowed to touch: the day is what the land renews by and the day is
// what people take by, so an age and the Rate beside it have to keep the
// balance they were tuned to. What changed is what those days are called - a
// crop is a month rather than a quarter of a hundred-tick year, and a wood
// six years rather than twenty - not how many of them there are.
var (
	Crop = &Process{Name: "crop", Of: Field, Yields: Grain,
		Stages: []Stage{{"sown", clock.Month / 2}, {"ear", clock.Month / 2}}}
	Brush = &Process{Name: "brush", Of: Wood, Yields: Berries, Rate: 0.0012,
		Stages: []Stage{{"scrub", 2 * Year}}}
	Timbering = &Process{Name: "timber", Of: Wood, Yields: Timber, Rate: 0.0004,
		Stages: []Stage{{"thicket", 6 * Year}}}
)

// Processes is everything that grows. A class no process names carries
// nothing standing, and a taking there draws on a stock that is simply
// there: an outcrop is stone and does not grow, and the water's fish come
// back without having to come on.
var Processes = []*Process{Crop, Brush, Timbering}

// InEar is the crop a harvest asks for.
var InEar = Crop.Phase("ear")

// Growing lists the processes that run on class c.
func Growing(c *Class) []*Process {
	var out []*Process
	for _, p := range Processes {
		if c == p.Of {
			out = append(out, p)
		}
	}
	return out
}

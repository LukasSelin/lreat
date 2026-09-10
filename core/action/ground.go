package action

import (
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// Ground, and what anybody makes of it.
//
// Every siting decision used to be the same call: the nearest tile that was
// legal to build on, found by walking square rings outward from an anchor. It
// had two faults and they compounded. The anchor was the market, which nobody
// chose, so a whole cohort of homeless agents was handed the identical plot
// and queued for it; and the test was legality rather than quality, so
// excellent ground one tile further out lost to any bare patch of grass. The
// land carries fertility, drainage, height and standing timber - the same
// readings that decide where the market itself is founded - and not one of
// them was ever consulted about a house.
//
// Here they are consulted, in a single figure, and by an agent that only
// knows the ground it has actually walked over. That last part is what turns
// siting into something with a decision in it: a settler with poor ground in
// mind may take it and be housed by nightfall, or walk out past the edge of
// what it knows and look for better. Neither is written down as the right
// answer. Which one it does is recognition, like everything else.

// The worth of ground, in tiles of walking saved a day. That is the currency
// homeCost and worthMoving are already written in, so the readings below can
// simply be added to them. A term of 6 means "this feature of the land is
// worth as much as living six tiles closer to everything you do".
const (
	// goodSoil is what fertile ground under and around a house is worth. A
	// field is cleared near home, and a rich one feeds two or three times
	// what a poor one does.
	goodSoil = 6
	// nearWaterWorth is what a river within reach is worth: something to
	// drink, fish to take, and a field that can be irrigated later.
	nearWaterWorth = 4
	// standingTimber is what a wood within reach is worth, and timberReach
	// how far away a wood may be and still count as one's own. This is the
	// term that answers a settlement felling its way to the horizon: ground
	// far from timber is worth less to live on, so where people settle
	// follows the treeline instead of receding from it.
	standingTimber = 6
	timberReach    = 10
	// floodRisk is what a house on ground that stands barely above the water
	// it drains into gives up. A water meadow grows a fine crop and is no
	// place to sleep.
	floodRisk = 8
	// steepGround is what a hillside costs. A tenth of a rise is a gentle
	// hill; a half is ground you would not plough and could not stand a wall
	// on.
	steepGround = 10
	// neighbourly is what living in the middle of the settlement is worth
	// rather than at the edge of the map, and neighbourReach the distance
	// over which that falls away.
	// It is the largest term here on purpose: people are worth more to each
	// other than any soil is worth to any of them, and a settlement is the
	// proof. Set at the size of the land terms it lost to them - founders
	// scattered to whatever ground read best, each raised a roof alone
	// twenty tiles from the next, and a country of hermits never traded,
	// never taught, never discovered anything and did not breed.
	//
	// This is what replaced the market as the reason a settlement holds
	// together. The old rule named the market in the code, which meant the
	// shape of every town was a decision taken once, by hand, before anybody
	// had lived there. This one names nobody: ground is worth more for having
	// neighbours on it, the neighbours are wherever people have already
	// chosen to build, and the first of them chose on the strength of the
	// land alone. Where a town goes is now an outcome rather than a constant.
	neighbourly    = 40
	neighbourReach = 8
	// elbowRoom is what a plot with its own ground all round it is worth
	// over one wedged between two walls.
	//
	// Kept as a price rather than a rule, so that a crowded settlement still
	// roofs its last arrival rather than leaving them outside. And priced
	// high, because the gaps it keeps are the only thing a road can ever be
	// laid along: at six, houses closed up, errands had nowhere to
	// concentrate, no ground wore through to the sixty crossings paving asks
	// for, and settlements that used to lay eighty tiles of street laid two.
	//
	// It is the knob between many small towns and few large ones, and it is
	// the one figure here the rest of the change is most sensitive to. It
	// also has to be re-measured whenever the ground does, which is worth
	// knowing: over 24 seeds it read best at 16 against the holdings alone,
	// and best at 10 once the woods were kept off the slopes and a loaded
	// agent could no longer swim. At 10 now: 20 settlements of 24, no
	// extinctions, median 115, and 71 tiles of street to master's 54.
	elbowRoom = 10
	// settlingWalk is how many days of a settled life the walk out to a plot
	// is set against. Worth is counted per day and lasts as long as the
	// house does; the walk to get there is paid once. Subtracting the one
	// from the other directly, as this first did, values a morning's walk
	// like a lifetime of them, and a settler forty tiles from anywhere then
	// always builds exactly where it is standing - which is how a founding
	// party of twenty turned into seven hamlets that never traded with each
	// other. What holds an agent back from a long walk is not arithmetic
	// here; it is habit.Near, in the fit of the act itself, which is
	// per-agent and learned rather than a constant.
	settlingWalk = 8
	// poorGround is the worth below which a tile is not worth carrying in
	// one's head at all, so that walking over a bog does not cost anybody a
	// memory slot.
	poorGround = -4
)

// bestAmenity bounds what the two searches in landWorth can add on top of
// what the tile says about itself. It lets a tile be dismissed on its cheap
// readings alone, which is what keeps an appraisal on every step affordable.
const bestAmenity = nearWaterWorth + standingTimber

// groundWorth is what a tile says about itself: soil, footing, drainage, and
// whether a way people use runs across it. Everything here is read off the
// tile and its eight neighbours, so it is cheap enough to ask on every step.
func groundWorth(w *world.World, p entity.Pos) float64 {
	t := w.Grid.At(p)
	// The soil that matters is not the soil under the floorboards - a house
	// stands on anything - but the best a field near the door could be.
	soil := t.Rich
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			q := entity.Pos{X: p.X + dx, Y: p.Y + dy}
			if w.Grid.In(q) {
				if r := w.Grid.At(q).Rich; r > soil {
					soil = r
				}
			}
		}
	}
	v := goodSoil * soil
	v -= floodRisk * clampUnit(1-t.Drain/world.FloodDepth)
	v -= steepGround * w.Grid.Slope(p)
	// A house astride a thoroughfare is a house people walk through. The
	// same reading, at the same price, that makes a mover step out of a
	// street's way in home.go.
	v -= inTheWay * t.Traffic / wornEnough
	return v
}

// landWorth is what the country round a tile is worth as ground: water and
// timber. With groundWorth it is the whole of what a place is like, and all
// of it changes slowly or not at all, which is what makes it worth carrying
// in one's head for years. It costs two short searches, so it is only asked
// of ground that has already proved itself on the cheap readings.
// timberKinds is the ground timber stands on, asked of the trees rather
// than named here. Siting a house reads the worth of every plot it
// considers and every plot reads this, so it is the second commonest
// search a settlement makes.
var timberKinds = world.KindsOffering(ontology.Timber)

func landWorth(w *world.World, p entity.Pos) float64 {
	v := groundWorth(w, p)
	if nearWater(w, p) {
		v += nearWaterWorth
	}
	if q, ok := w.Grid.NearestOfKind(p, timberReach, timberKinds, func(_ entity.Pos, t *world.Tile) bool {
		return t.Offers(ontology.Timber) >= 0.3
	}); ok {
		// A wood at the door is worth all of it; one at the edge of reach,
		// almost none. Wood is fetched an armful at a time, so the walk is
		// paid on every single load.
		v += standingTimber * (1 - float64(w.Grid.Dist(p, q))/timberReach)
	}
	return v
}

// companyWorth is what a plot is worth for the people round it: the share of
// the settlement living within reach, falling off with distance.
//
// It is deliberately not part of what an agent remembers about a place, and
// that separation is the whole of what makes a town hold together. A memory
// of ground is a memory of soil and water and timber, none of which move.
// Who lives where moves constantly - and when the social reading was frozen
// into the memory alongside the soil, an agent who had once walked past a
// spot on a day five people happened to be standing near it remembered that
// spot as excellent for ever, and went and built there alone years later.
// The settlement scattered into hamlets that never met, and belonging - which
// is what a birth needs as much as food - never rose at all. Nobody needs to
// remember where their people are. They can see.
//
// Counted as a share rather than a headcount so it cannot saturate, and as a
// falling-off rather than a radius so there is a gradient at every distance
// instead of a cliff at one. The gradient is what does the work, and it has
// to be steeper than the walk it is set against, or nobody at the edge of the
// map ever has a reason to come home.
func companyWorth(w *world.World, p entity.Pos) float64 {
	if len(w.Agents) == 0 {
		return 0
	}
	share := 0.0
	for _, o := range w.Agents {
		// Where somebody lives, or where they are if they live nowhere yet.
		// Their door is the honest reading: an agent is out at the treeline
		// half its life and is not thereby a neighbour of the treeline. The
		// homeless are counted where they stand because at the founding
		// there are no doors at all, and a reading of zero everywhere would
		// leave the first settlers nothing to gather around.
		at := o.Pos
		if o.HasHome {
			at = o.Home
		}
		share += 1 / (1 + float64(w.Grid.Dist(at, p))/neighbourReach)
	}
	return neighbourly * share / float64(len(w.Agents))
}

// LandWorth is the whole appraisal as it stands right now: the ground, and
// the people round it. Siting reads it through this; memory keeps only the
// landWorth half.
func LandWorth(w *world.World, p entity.Pos) float64 {
	return landWorth(w, p) + companyWorth(w, p)
}

// Notice is what an agent takes in of the ground under its feet. It is called
// for every tile walked onto and for the tile an agent stands on while it
// works, which is the whole of how anybody comes to know anywhere: there is
// no survey, and nobody is told. An agent that has spent its life between its
// door and one wood knows two places and will site its house in one of them.
//
// It draws no randomness and writes nothing outside the agent, so it is safe
// wherever in the tick it is called from.
func Notice(a *entity.Agent, w *world.World) bool {
	p := a.Pos
	// Any ground that could be built on, not only ground with its own sides
	// clear. Remembering only the roomy plots was tried and it emptied the
	// towns: inside a settlement almost nothing has its sides clear, so
	// nobody could remember anywhere they actually lived, every head filled
	// up with open country miles out, and that is where the next generation
	// went. Elbow room is worth something when a plot is chosen, in
	// KnownPlot, and nothing at all when ground is merely looked at. Over
	// sixteen seeds the distinction doubled the median settlement.
	if !w.Grid.At(p).Buildable() {
		return false
	}
	base := groundWorth(w, p)
	if base+bestAmenity < poorGround {
		return false
	}
	// With a head already full of better ground, even the most generous
	// reading of the country round this tile could not earn it a slot.
	if len(a.Places) >= entity.MaxPlaces {
		worst := a.Places[0].Worth
		for i := range a.Places {
			if a.Places[i].Worth < worst {
				worst = a.Places[i].Worth
			}
		}
		if _, known := a.Knows(p); !known && base+bestAmenity <= worst {
			return false
		}
	}
	if v := landWorth(w, p); v >= poorGround {
		return a.Remember(p, v, w.Tick)
	}
	return false
}

// KnownPlot is the best plot this agent knows of, weighed against the walk
// out to it, including the ground it is standing on this minute. Worth is in
// tiles of walking, so the two subtract.
//
// This is the whole of siting. There is no scan: an agent cannot site a house
// on ground it has never seen, which is why walking about is worth anything
// and why two agents standing side by side, having led different lives,
// disagree about where to build.
func KnownPlot(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	best, bestValue, found := entity.Pos{}, 0.0, false
	consider := func(p entity.Pos, worth float64) {
		if !w.Grid.At(p).Buildable() {
			return
		}
		// Elbow room is a preference and not a rule. A settlement that
		// builds wall to wall has nowhere left to put a street, so a plot
		// with its own ground round it is worth a good deal more - but when
		// there is no such plot left, a roof beside a neighbour's still
		// beats no roof at all, and the old hard test would have left the
		// last settler in a crowded place standing outside.
		if !w.Grid.RoomToBuild(p) {
			worth -= elbowRoom
		}
		if v := worth - float64(w.Grid.Dist(a.Pos, p))/settlingWalk; !found || v > bestValue {
			best, bestValue, found = p, v, true
		}
	}
	for i := range a.Places {
		// What is remembered is the ground. Who is living round it now is
		// looked at now.
		consider(a.Places[i].Pos, a.Places[i].Worth+companyWorth(w, a.Places[i].Pos))
	}
	// The ground here and now is always in the running, appraised fresh
	// rather than remembered, so that settling for where one happens to be
	// standing is never off the table however little one has seen.
	if p, ok := plotNear(a, w, a.Pos); ok {
		consider(p, LandWorth(w, p))
	}
	// And so is the nearest ground that can be built on at all, which in a
	// crowded neighbourhood is not the same tile.
	if p, ok := w.Grid.Nearest(a.Pos, searchRadius, func(_ entity.Pos, t *world.Tile) bool {
		return t.Buildable()
	}); ok {
		consider(p, LandWorth(w, p))
	}
	return best, found
}

// Scouting. Every other errand in the catalog is a reason to be somewhere in
// particular; this is the one that is a reason to be somewhere else. It
// belongs to the moment when the lower needs are quiet enough that a person
// can afford to wonder what is over the hill, and its prior is the only one
// in the catalog that asks for distance rather than shrinking from it.
//
// What it pays back is not in the act. A scout comes home with nothing but
// places in mind, and places in mind are only worth something later, when one
// of them turns out to be where the house goes. It earns through the trace
// when that happens soon enough, and carries a small satisfaction of its own
// so that a lifetime of fruitless walking is discouraged rather than a habit
// extinguished outright - the same problem paving has, and answered the same
// way and at the same size.
const (
	// scoutRange is how far out a scout sets its sights: past the edge of
	// what a settled life covers, and not so far that the walk is a season.
	scoutRange = 14
	// scoutFinding is what coming back knowing somewhere new is worth to the
	// person who went.
	scoutFinding = 0.03
	// contentedGround is the worth of ground at which somebody stops looking
	// for better. It is the gate on the whole act, and it has to be there.
	//
	// Recognition does not divide by how long a thing takes - only by how
	// far off it is, through habit.Near - so an act whose moment keeps
	// arriving is taken however much of the day it eats. Ungated, scouting
	// was a seventh of everything anybody did: the founding party spent its
	// first years walking in opposite directions, was never in one place
	// long enough to meet or trade, and belonging never came near what a
	// birth asks for. Over sixteen seeds it cost eleven settlements of
	// sixteen and turned no extinctions into seven.
	//
	// Gated this way the act says what it should say: you go looking when
	// what you know of is not good enough to settle for, and you stop when
	// it is. Nobody wanders for the sake of it, and nobody has to be told
	// when to stop.
	contentedGround = 10
)

// bearings are the eight ways out of anywhere, in a fixed order.
var bearings = [8]entity.Pos{
	{X: 0, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 0}, {X: 1, Y: 1},
	{X: 0, Y: 1}, {X: -1, Y: 1}, {X: -1, Y: 0}, {X: -1, Y: -1},
}

// unknownWay is the direction with least of what this agent already knows in
// it: the way out of here that is most worth walking. Ties are broken from a
// bearing of the agent's own, so that a party who have seen nothing at all do
// not all set off in single file the way the old siting rule sent them all to
// one plot.
//
// It draws no randomness. Target is asked of every candidate every time an
// agent decides, and an agent that spent luck on being asked would decide
// differently for having been asked.
func unknownWay(a *entity.Agent, w *world.World) entity.Pos {
	var known [8]int
	for i := range a.Places {
		known[bearingOf(w.Grid.Delta(a.Pos, a.Places[i].Pos))]++
	}
	own := int(a.ID) % len(bearings)
	best := own
	for k := 0; k < len(bearings); k++ {
		b := (own + k) % len(bearings)
		if known[b] < known[best] {
			best = b
		}
	}
	return bearings[best]
}

// bearingOf is which of the eight ways an offset points.
func bearingOf(d entity.Pos) int {
	for i, b := range bearings {
		if sign(d.X) == b.X && sign(d.Y) == b.Y {
			return i
		}
	}
	return 0
}

func sign(v int) int {
	switch {
	case v < 0:
		return -1
	case v > 0:
		return 1
	}
	return 0
}

// worthLooking reports whether this agent has reason to go and see the
// country: nothing in mind good enough to settle on, or a house on ground it
// would rather not have. Somebody content with what they know stays home.
func worthLooking(a *entity.Agent, _ *world.World) bool {
	if a.HasHome {
		return false
	}
	best := float64(poorGround)
	for i := range a.Places {
		if a.Places[i].Worth > best {
			best = a.Places[i].Worth
		}
	}
	return best < contentedGround
}

// scoutSite is dry land a good way off in the direction least known.
func scoutSite(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	d := unknownWay(a, w)
	aim := entity.Pos{X: a.Pos.X + d.X*scoutRange, Y: a.Pos.Y + d.Y*scoutRange}
	if w.Grid.Wrap {
		aim = w.Grid.Norm(aim) // east of the east edge is the west
	} else {
		aim.X = clampInt(aim.X, 0, w.Grid.W-1)
	}
	aim.Y = clampInt(aim.Y, 0, w.Grid.H-1)
	if aim == a.Pos {
		return entity.Pos{}, false
	}
	return w.Grid.Nearest(aim, scoutRange, func(_ entity.Pos, t *world.Tile) bool {
		return !t.Is(ontology.Water)
	})
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

var Scout = &Def{
	Name: "scout", Ticks: 1, Available: worthLooking, Target: scoutSite,
	Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
		// Only the value rule reads this. Under recognition an act is chosen
		// by the moment it belongs to and learned from what actually
		// happened, and neither consults Expect.
		return need.Levels{need.Actualization: scoutFinding}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		// Paid for the finding, not for the going. An unconditional
		// satisfaction here is a reward for an idle act, which is the same
		// mistake a rest that restored more under a roof was: it reinforces
		// the doing of it, agents wander instead of living, and a
		// settlement whose people are all half a day away from each other
		// stops meeting, stops trading, and stops having children. A day in
		// country one already knew turns up nothing and is worth nothing.
		if !Notice(a, w) {
			return
		}
		a.Needs.Add(need.Actualization, scoutFinding)
		w.Emit(event.Acted, a.ID, 0, "%s found new country", a.Name)
	},
}

func init() {
	// Going to look belongs to somebody with no roof of their own, fed
	// enough to spare the day, and with some curiosity left over once the
	// tiers below have been quieted. Curiosity is the smaller half of it:
	// asked for as strongly as studying asks, scouting outranked study in
	// the one moment study exists for, and a settlement of wanderers learns
	// nothing.
	//
	// It is the only prior in the catalog that says nothing about Near, and
	// that silence is the whole design. Every other errand wants its target
	// close and loses fit as the walk grows; this one simply does not mind.
	// So when everything an agent knows of has been used up and all that is
	// left is far away, going to look is what is left standing - which is
	// the pressure that makes anybody explore, and it arrives without a
	// single line saying "explore when the neighbourhood is exhausted".
	seed(Scout, reachEveryday, habit.Signature{
		habit.Shelter: -0.7, habit.Hunger: -0.5, habit.Curious: 0.4,
	})
}

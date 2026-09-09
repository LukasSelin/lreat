package world

import (
	"math"
	"sort"

	"lreat/core/entity"
)

// A world made by what happened to it.
//
// The land used to be a picture: two fields of noise, one for the lie of the
// country and one for the ridges, with the high ground masked in where a
// third field said so. It makes handsome maps and it makes them all at once,
// out of nothing, in the shape somebody decided mountains ought to be.
// Nothing on such a map has a reason - a range is where it is because a
// lattice was high there, and the rock under it is what another lattice said.
//
// Here the land is the leavings of a history instead, run forward from a
// world too hot to have a surface worth the name. It is not geology and does
// not pretend to be: there is no mantle in it, no heat budget, no isostasy.
// What it has that a picture cannot have is causes - a range stands where two
// plates met, the rock in it is what that meeting made of what was there, the
// basin beside it is full of the range's own debris, and the good ground at
// the mouth of the valley is what the rivers carried out of both. Anything a
// later change wants to put on a map - where the ore is, where the coal is,
// where the ground still shakes - can then be asked of what happened rather
// than painted on afterwards.
//
// It runs behind Config.Epochs and no preset uses it yet. The constants the
// whole settlement model is tuned against - the sixty metres of lowland, the
// share of a map that can be ploughed - were measured on the picture, so a
// history has to be shown to hand them the same kind of map before it can be
// allowed to make the only one. See normalise, which is where that join is.

// The three eras. Molten is a world with no rigid crust at all, where the
// surface is convection and nothing that forms outlasts the forming of it;
// then the crust goes rigid and breaks; and the rest is the plate era, which
// is all the history a map actually keeps.
//
// moltenChurns is how many times the molten world turns itself over. Each
// churn half forgets the one before, so the era leaves the residue of all of
// them and the shape of none, which is the honest way to draw a surface that
// had no time to become anything.
const (
	moltenChurns = 6
	moltenMemory = 0.5
	// How much of a lowland's worth of relief the churning leaves behind. It
	// is small on purpose: what survives the molten era is a floor with a
	// swell in it, not country. Left at a whole Relief the lumps were sixty
	// metres over five tiles, and normalise then stretched them into a map
	// twice as steep as a drawn one, which nothing downstream could plough
	// or walk.
	moltenRelief = 0.8 * Relief
)

// How many plates a world breaks into, quoted at the width the rest of the
// map's constants are quoted at and scaled with the map: a plate is a piece
// of a world, so a bigger world has bigger pieces rather than more of them.
//
// oceanShare is how many of them are ocean floor - thin, dense, low, and the
// first to go down when two plates meet. More than half, because ocean floor
// is most of a world's surface, and it is the term deciding whether a map
// comes out a continent with seas in it or an ocean with islands in it.
const (
	plateCount = 5
	plateSpan  = DefaultWidth
	// oceanShare is how many of a world's plates are ocean floor, and it is
	// read off how much sea the finished map asks for. A world is mostly
	// ocean floor; a valley cut out of one and asking for no sea at all is a
	// piece of it that was never under any, so its plates should be the
	// continent they plainly are. Fixed at a world's own share, the default
	// map came out with seven tiles in ten standing on ocean floor and a
	// bedrock map that was mostly basalt - true of Iceland and of nowhere a
	// valley like this one is.
	oceanFloor  = 0.25
	oceanPerSea = 1.0
	// beltWidth is how far from a seam the ground is raised or dropped by
	// what is happening at it, in tiles. A collision does not make a wall one
	// tile wide: it thickens a belt of country either side, which is why a
	// range has flanks and foothills and a pass through it. Applying the lift
	// only where two plates actually touch made a map of knife edges - the
	// steep ground doubled and there was nothing to walk up.
	beltWidth = 10.0
	// axisWidth is how near the seam itself a tile has to be for what is
	// happening there to change what it is made of, rather than only how
	// high it stands. A collision lifts a belt of country ten tiles wide and
	// cooks the rock in the middle of it; a rift drops a valley and floors
	// the axis of it with melt. Recording the rock across the whole belt made
	// seven tiles in ten of a map igneous or metamorphic, which is a map of
	// nothing but seams.
	axisWidth = 2.0
)

// How far a plate moves in an epoch, in tiles, at the start of the plate era
// and at the end of it. A young world convects hard and its plates race; an
// old one has cooled and slowed. That decay is the whole of what cooling
// means here, and it is why the ranges raised early are worn down to shields
// by the end while the ones raised late still stand.
// historySea is how much of a young world is under water while its history
// runs. A world has oceans whatever the map cut out of it at the end does -
// the default valley asks for no sea at all - and what lay under one age
// after age is where limestone comes from. So a history floods itself to this
// share, records who was drowned, and hands the finished ground to the map's
// own sea share, which is why a valley with no sea in it can still have
// limestone country: that ground was a seabed once.
//
// A share of the map and not a level read off where the two kinds of crust
// are riding, which was tried and is the better-sounding rule: the water
// fills what is low because what is low is ocean floor. It is worse in the
// one way that matters. The plates are moved about by everything else that
// happens to them, so on two seeds of five the level came out under
// everything and nothing was ever drowned - no seabed, and so no limestone
// anywhere on the map. A share always drowns something.
const historySea = 0.35

// marineMud is how much a sea bed off a shore takes in an epoch, against the
// one an epoch of burial on land is worth. fillEnough is how much has to have
// fallen on a tile before what it is made of is the fill rather than whatever
// was underneath - a couple of epochs' worth, so that ground which dipped
// below its river once or twice is still the basement it always was. And
// Which of the two a fill makes is settled by the sand in it against the
// clay, which is the sorting asked the only question it can answer: what
// stopped here, and what went on past.
const (
	marineMud  = 0.5
	fillEnough = 3.0
	// coarseShare is how much of a world's filled ground comes out sandstone
	// rather than shale: the sandiest third of it. It is a share and not a
	// cutoff for the reason every other share on this map is - see
	// waterShare - and here the reason is sharper than usual. What a deposit
	// is made of hardly varies while a history is running, because until the
	// rock is settled at the end every tile is weathering the same basalt,
	// so the sand in one basin and the next differ by a few hundredths and
	// any fixed line puts nearly all of them on one side of it: at sand
	// against clay the map came out a third sandstone, and one step stricter
	// it came out with none at all. Ranking the fills against each other asks
	// the only thing the sorting can actually answer - which of these
	// stopped soonest - and a third is about the share of the world's
	// sedimentary rock that is sandstone.
	coarseShare = 0.35
)

// deepWeather is how many ages of weather an epoch of the earth is worth.
// Erode's age is a decade, and an epoch here is not: mountains raised and
// never worn stand as walls, and a map of walls is one nobody can cross.
// What this figure is really setting is the balance between how fast the
// ground goes up at the seams and how fast the weather takes it down again,
// which is the balance that decides whether a map comes out as ridges or as
// country.
const deepWeather = 1.5

// smoothing is how many times a finished history is softened before its
// heights are matched to a drawn map's spread. A seam raises a range narrower
// than the country a drawn map puts its high ground over, and since the
// matching hands out the same heights either way, a narrower range means a
// steeper one: on one seed of three the steepest tenth came out at 1.03
// against a drawn 0.38. Softening spreads the extremes over the ground around
// them, and the matching afterwards puts the heights back exactly, so what it
// costs is sharpness and not scale.
const smoothing = 2

// marginRamp is how far the step at the edge of a plate is spread, in passes
// of a nine-tile average - so a handful of tiles either side, which is a
// continental margin at this scale.
const marginRamp = 6

var smoothingAt = smoothing

// A map is a region and not a world - eighty tiles at TileSpan is two
// kilometres of country - so these are what a boundary's *works* travel
// across the ground, not what a plate does in any real sense. At a tile an
// epoch a seam crosses a fifth of a default map over a whole history, which
// leaves plate interiors that were never touched by anything. Faster than
// that and every tile on the map has been in a mountain range at some point,
// which is the same uniformity this was meant to replace.
const (
	driftFast = 1.0
	driftSlow = 0.2
)

// What a meeting of plates does to the ground, in metres per epoch at a
// head-on closing of one tile. Continents crumple and stay up because they
// are too light to go down; ocean floor meeting a continent goes under it,
// which trenches the one and lifts an arc of volcanoes on the other; two
// floors meeting make islands out of open water. Parting drops the ground and
// floors it with what comes up.
const (
	orogeny  = 45.0
	arcLift  = 28.0
	trench   = -20.0
	islandUp = 18.0
	rifting  = -15.0
)

// How high a plate floats before anything happens at its edges: a continent
// stands above the ocean floor because it is thicker and lighter, and that
// one fact is what gives a world coasts at all rather than an even skin of
// water. The floor is not at nothing, because a height of nothing is where
// the ground stops being allowed to fall - see wear - and a sea floor pinned
// against that stop comes out as a dead flat plain: half the map at a slope
// of a hundredth, against a tenth on a drawn one. What matters is the
// distance between the two levels and not either figure, since normalise
// rescales the lot.
//
// settling is how much of the way to its own level a plate comes in an
// epoch, so that crust which changes hands rises or sinks over an age rather
// than jumping. It moves a whole plate by what its middle is short of, so a
// plate keeps the country it is carrying; see tectonics.
const (
	oceanFreeboard     = 100.0
	continentFreeboard = 340.0
	settling           = 0.15
)

// How many places in a world are fed from below rather than at their edges,
// and how far each one's works reach. They stay where they are while the
// plates come and go over them, which on the earth draws a chain of islands;
// here the crust does not travel across the map, so what a hotspot leaves is
// a volcanic province rather than a chain. It is the one place this admits to
// being a model of a model.
const (
	hotspots     = 2
	hotspotReach = 7.0
	hotspotLift  = 10.0
	// hotspotWakes is how often one of them is awake in an epoch. A volcano
	// is not a thing that happens continuously for the age of a world: it
	// goes off, and then it is quiet for longer than anybody watching it will
	// be alive. Left erupting every epoch at forty-five metres a time, two
	// hotspots built seven-hundred-metre cones five tiles across - three
	// perfect circles that were, on two seeds of four, the steepest ground on
	// the map by a wide margin and the whole of what made a made world
	// steeper than a drawn one.
	hotspotWakes = 0.35
)

// Plate is one piece of a world's crust: where its middle is, where it is
// going, and whether it is ocean floor or continent. Positions are in tiles
// and are not whole numbers, because a plate moves less than a tile in an
// epoch late on and rounding that off would stop it moving at all.
type Plate struct {
	X, Y   float64
	DX, DY float64
	Ocean  bool
}

// seam is what is happening at the nearest place two plates meet: how much
// the ground there rises or falls in an epoch, and how far off that place is.
// It is the generator's working, kept on the grid only so that a history does
// not allocate a map's worth of it every epoch.
type seam struct {
	lift  float64
	away  float64
	makes made
	found bool
}

// made is what a meeting makes of the rock at its axis, as against what it
// does to the height of the country round it. A trench makes nothing: it is
// ground going down and away.
type made uint8

const (
	nothing made = iota
	crushed
	arc
	melt
)

// record is what has been done to a tile over the whole history, which is
// what decides the rock it ends up being. It is the generator's working and
// is not kept: the tile keeps the answer.
type record struct {
	// melt is what came up and cooled at the surface, and pluton what melted
	// under an arc and cooled at depth. They are the same fire and they make
	// different rock, which is the whole reason for keeping them apart: what
	// reaches the air is basalt, and what stops on the way is the granite
	// that a few million years of weather then lays bare.
	melt   float64
	pluton float64
	crush  float64 // metres raised by two plates meeting, and cooked doing it
	laid   [Grains]float64
	// submerged is how many epochs this tile spent under the sea a history
	// floods itself to. Ground that lay there quietly, with no river mud
	// reaching it, is where limestone comes from.
	submerged int
}

// history makes a world by running one. It leaves every tile with a height, a
// rock, the plate it rides and the epoch that rock dates from, and it leaves
// the drainage worked out, so that everything after it in Generate - the
// woods, the outcrops, the soils, the market - reads the same kind of ground
// it would have read from the picture.
func (w *World) history(g *Grid, epochs int, sea float64) {
	w.molten(g)
	// The water has to have something to carry. Soil is made from the rock
	// beneath it, and at the end of the molten era that is basalt everywhere;
	// what the epochs then do is sort it, which is what makes the fill of one
	// basin coarse and the next one fine - and so which of them becomes
	// sandstone and which shale.
	g.fill()
	g.drain()
	g.soilTexture()
	plates := w.firstPlates(g, sea)
	book := make([]record, len(g.Tiles))

	for e := 0; e < epochs; e++ {
		// How far through the era we are, which is how far the world has
		// cooled: the plates slow as it goes.
		through := float64(e) / math.Max(1, float64(epochs-1))
		g.partition(plates)
		w.tectonics(g, plates, book, e)
		// An age of weather between the ages of the earth. What was raised
		// this epoch starts coming down in the next, and what comes off it is
		// what fills the basins - which is where a finished map's sandstone
		// and shale come from.
		g.wear(deepWeather)
		g.fill()
		g.drain()
		g.keepBook(book, e)
		drift(plates, through)
	}

	g.settleRock(book, plates, epochs)
	for k := 0; k < smoothing; k++ {
		g.soften()
	}
	w.normalise(g)
	g.fill()
	g.drain()
}

// molten is the world before it had a crust. Each churn is a new surface half
// blended into the last, so the era leaves a lumpy floor with no lasting
// shape - no ranges, no basins, nothing that outlives the making of it. All
// of it is basalt, because that is what a world cools into, and any that is
// still basalt at the end is the oldest ground on the map.
func (w *World) molten(g *Grid) {
	h := make([]float64, len(g.Tiles))
	for c := 0; c < moltenChurns; c++ {
		// Fine cells rather than broad swells, and finer as the churns go on:
		// convection at this stage is small and furious.
		cell := w.lattice(g, math.Max(4, float64(g.Span())/float64(2+2*c)))
		for i := range h {
			h[i] = moltenMemory*h[i] + (1-moltenMemory)*cell[i]
		}
	}
	for i := range g.Tiles {
		t := &g.Tiles[i]
		t.Height = moltenRelief * h[i]
		t.Bedrock, t.Formed = Basalt, 0
	}
}

// firstPlates is the crust going rigid: it breaks, and the pieces start
// moving. Where the breaks fall is drawn rather than derived - a world's
// first plates are an accident of how it cooled and there is nothing to read
// them off - but which pieces are ocean and which are continent decides the
// shape of everything after.
func (w *World) firstPlates(g *Grid, sea float64) []Plate {
	n := max(3, plateCount*g.Span()/plateSpan)
	ocean := clamp01(oceanFloor + oceanPerSea*sea)
	plates := make([]Plate, n)
	for i := range plates {
		a := 2 * math.Pi * w.RNG.Float64()
		plates[i] = Plate{
			X:     w.RNG.Float64() * float64(g.W),
			Y:     w.RNG.Float64() * float64(g.H),
			DX:    driftFast * math.Cos(a),
			DY:    driftFast * math.Sin(a),
			Ocean: w.RNG.Float64() < ocean,
		}
	}
	// A world has both kinds in it. Left to the draw, a valley - which asks
	// for no sea and so for few ocean plates - came out on two seeds of five
	// with nothing but continent, and a world with no floor anywhere has no
	// arcs in it, nothing going under anything, and so no granite and no
	// islands: three of the six rocks lose the only place they come from.
	ocean, land := 0, 0
	for i := range plates {
		if plates[i].Ocean {
			ocean++
		} else {
			land++
		}
	}
	switch {
	case ocean == 0:
		plates[0].Ocean = true
	case land == 0:
		plates[0].Ocean = false
	}
	return plates
}

// drift moves every plate on by one epoch and slows it by how far the world
// has cooled. A plate keeps its bearing: what changes is how fast it holds
// it, so a boundary sweeps across the ground in one direction for the whole
// history and the range it raises is longer than the epoch that raised it.
func drift(plates []Plate, through float64) {
	speed := driftFast + (driftSlow-driftFast)*through
	for i := range plates {
		p := &plates[i]
		if d := math.Hypot(p.DX, p.DY); d > 0 {
			p.DX, p.DY = p.DX/d*speed, p.DY/d*speed
		}
		p.X += p.DX
		p.Y += p.DY
	}
}

// partition says which plate each tile rides: the nearest middle, which on a
// globe is measured the short way round. It is taken again every epoch,
// because the middles have moved and the boundaries move with them.
func (g *Grid) partition(plates []Plate) {
	for i := range g.Tiles {
		x, y := float64(i%g.W), float64(i/g.W)
		best, at := math.Inf(1), 0
		for k := range plates {
			dx := plates[k].X - x
			if g.Wrap {
				dx = math.Remainder(dx, float64(g.W))
			}
			dy := plates[k].Y - y
			if d := dx*dx + dy*dy; d < best {
				best, at = d, k
			}
		}
		g.Tiles[i].Plate = uint8(at)
	}
}

// tectonics is one epoch of what the plates do to the ground they carry.
//
// Every tile is asked what is happening at its own edges: for each neighbour
// riding a different plate, how fast the two are closing or parting along the
// line between them. Closing raises ground, and some of what it raises comes
// up as melt; parting drops it and fills the drop from below. Away from an
// edge nothing happens at all, which is why the middle of a plate is the
// oldest, flattest ground on a map and everything worth looking at is at the
// seams.
func (w *World) tectonics(g *Grid, plates []Plate, book []record, epoch int) {
	n := len(g.Tiles)
	if len(g.seam) != n {
		g.seam = make([]seam, n)
		g.seamQueue = make([]int32, 0, n)
	}
	for i := range g.seam {
		g.seam[i] = seam{}
	}
	g.seamQueue = g.seamQueue[:0]

	// Where the plates actually touch, and what is happening there.
	for i := range g.Tiles {
		t := &g.Tiles[i]
		p := entity.Pos{X: i % g.W, Y: i / g.W}
		mine := plates[t.Plate]
		worst := 0.0
		var at *Plate
		for _, off := range dirs {
			q := entity.Pos{X: p.X + off.X, Y: p.Y + off.Y}
			if g.Wrap {
				q = g.Norm(q)
			}
			if !g.In(q) || g.At(q).Plate == t.Plate {
				continue
			}
			other := &plates[g.At(q).Plate]
			// How fast the two are closing along the line between them, as a
			// share of a head-on meeting of one tile in an epoch.
			d := math.Hypot(float64(off.X), float64(off.Y))
			nx, ny := float64(off.X)/d, float64(off.Y)/d
			closing := (mine.DX-other.DX)*nx + (mine.DY-other.DY)*ny
			if math.Abs(closing) > math.Abs(worst) {
				worst, at = closing, other
			}
		}
		if at == nil {
			continue
		}
		lift, makes := liftOf(mine, *at, worst/driftFast)
		g.seam[i] = seam{lift: lift, makes: makes, found: true}
		g.seamQueue = append(g.seamQueue, int32(i))
	}

	// And how far its works reach either side of it. The lift falls off with
	// distance from the seam, so a range has flanks rather than sides.
	for k := 0; k < len(g.seamQueue); k++ {
		i := g.seamQueue[k]
		here := g.seam[i]
		p := entity.Pos{X: int(i) % g.W, Y: int(i) / g.W}
		for _, off := range dirs {
			q := entity.Pos{X: p.X + off.X, Y: p.Y + off.Y}
			if g.Wrap {
				q = g.Norm(q)
			}
			if !g.In(q) {
				continue
			}
			j := g.Index(q)
			step := here.away + math.Hypot(float64(off.X), float64(off.Y))
			if step >= beltWidth {
				continue
			}
			if g.seam[j].found && g.seam[j].away <= step {
				continue
			}
			g.seam[j] = seam{lift: here.lift, away: step, makes: here.makes, found: true}
			g.seamQueue = append(g.seamQueue, int32(j))
		}
	}

	// Where each plate floats. A plate rides at its own level because of what
	// it is made of, and it carries whatever country it has on its back while
	// it does: so the whole plate is moved by what its middle is short of,
	// and not each tile by what it is short of itself. Pulling every tile
	// toward the level directly was the first way this was written, and over
	// sixteen epochs it left nothing but plains and cliffs - a tenth of a
	// slope was the ninetieth percentile of the drawn map and this could not
	// manage it at the fiftieth, because each pass took another seventh of
	// whatever texture the ground had.
	var sum [256]float64
	var count [256]float64
	for i := range g.Tiles {
		sum[g.Tiles[i].Plate] += g.Tiles[i].Height
		count[g.Tiles[i].Plate]++
	}
	var shift [256]float64
	for k := range plates {
		want := oceanFreeboard
		if !plates[k].Ocean {
			want = continentFreeboard
		}
		if count[k] > 0 {
			shift[k] = (want - sum[k]/count[k]) * settling
		}
	}
	// Spread the step at the edge of a plate into a ramp. A continent stands
	// a quarter of a kilometre above the floor beside it, and where the two
	// meet is a margin - a shelf, and a slope down off it - rather than a
	// wall. Applied as a step, that one boundary was the steepest ground on
	// the map by a mile: on two seeds of four the steepest tenth came out at
	// three times a drawn map's, and no amount of softening the finished
	// heights could undo it, because a step a tile wide is where the height
	// actually is.
	rise := make([]float64, len(g.Tiles))
	for i := range g.Tiles {
		rise[i] = shift[g.Tiles[i].Plate]
	}
	for k := 0; k < marginRamp; k++ {
		rise = g.spread(rise)
	}

	for i := range g.Tiles {
		t := &g.Tiles[i]
		t.Height += rise[i]

		if s := g.seam[i]; s.found {
			// Eased and not cut, so that a range has feet. The drawn
			// generator learned the same thing about its upland mask: cut
			// straight, the high country began at a wall with no approach to
			// it, and the ninetieth percentile of slope came out at twice
			// what a drawn map has.
			by := s.lift * smooth(1-s.away/beltWidth)
			t.Height += by
			if s.away <= axisWidth {
				switch s.makes {
				case crushed:
					book[i].crush += math.Abs(by)
				case arc:
					// An arc cooks what it pushes up and melts what goes
					// under it, and it is mostly the melting: two parts fire
					// to one of crushing, where a collision is all crushing
					// and no fire at all. Split evenly, an arc could never
					// come out as anything but the crushed rock, and the
					// granite it should leave had nowhere to come from.
					book[i].crush += math.Abs(by) / 3
					book[i].pluton += 2 * math.Abs(by) / 3
				case melt:
					book[i].melt += math.Abs(by)
				}
				if s.makes != nothing {
					t.Formed = uint8(epoch)
				}
			}
		}
		t.Height = math.Max(0, t.Height)
	}
	w.hotspot(g, book)
}

// liftOf is what a meeting does to the ground at the seam itself, in metres
// an epoch: which of the four kinds of meeting this is, times how hard. A
// negative closing is a parting.
func liftOf(mine, other Plate, closing float64) (float64, made) {
	switch {
	case closing <= 0:
		// They are parting: the ground drops and melt fills the axis.
		return rifting * -closing, melt
	case !mine.Ocean && !other.Ocean:
		// Two continents. Neither will go down, so both go up, and the rock
		// in the middle of it is cooked and squeezed.
		return orogeny * closing, crushed
	case !mine.Ocean:
		// The floor goes under us and melts on the way: an arc of volcanoes
		// on a rising edge, half crush and half fire.
		return arcLift * closing, arc
	case !other.Ocean:
		// We are the floor going under them, and what we are made of goes
		// down with us.
		return trench * closing, nothing
	default:
		// Two floors: islands come up out of open water.
		return islandUp * closing, melt
	}
}

// hotspot is melt coming up in the middle of a plate rather than at its edge,
// and it is what puts a volcano where nothing is colliding. Where they are is
// drawn once for a world and does not move, so the same places go on erupting
// age after age.
func (w *World) hotspot(g *Grid, book []record) {
	if g.hot == nil {
		g.hot = make([]entity.Pos, hotspots)
		for i := range g.hot {
			g.hot[i] = entity.Pos{X: w.RNG.IntN(g.W), Y: w.RNG.IntN(g.H)}
		}
	}
	r := int(hotspotReach)
	for _, h := range g.hot {
		if w.RNG.Float64() > hotspotWakes {
			continue // quiet this age
		}
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				d := math.Hypot(float64(dx), float64(dy))
				if d > hotspotReach {
					continue
				}
				q := entity.Pos{X: h.X + dx, Y: h.Y + dy}
				if g.Wrap {
					q = g.Norm(q)
				}
				if !g.In(q) {
					continue
				}
				lift := hotspotLift * smooth(1-d/hotspotReach)
				g.At(q).Height += lift
				book[g.Index(q)].melt += lift
			}
		}
	}
}

// keepBook writes down, after an epoch of weather, what the epoch left on
// each tile: what was buried, and what lay under water. The soil's own
// make-up carries the first - the water sorted what it laid down, so a tile
// buried in sand reads as sand - and the second is simply counted.
func (g *Grid) keepBook(book []record, epoch int) {
	h := make([]float64, len(g.Tiles))
	for i := range g.Tiles {
		h[i] = g.Tiles[i].Height
	}
	sea := quantile(h, historySea)
	for i := range g.Tiles {
		t := &g.Tiles[i]
		if t.Wet() || t.Height <= sea {
			book[i].submerged++
			// What a sea bed gets depends on whether anything is being
			// washed into it. Off a shore there is mud, and mud makes shale;
			// out where no land is near enough to send any, the water is
			// quiet and what settles is what lived there, which makes
			// limestone. Nothing here knows how far the shore is, only
			// whether it is next door, which is enough to tell a bed that
			// silts up from one that does not.
			if g.offshore(entity.Pos{X: i % g.W, Y: i / g.W}, sea) {
				book[i].laid[Clay] += marineMud * 0.7
				book[i].laid[Silt] += marineMud * 0.3
				t.Formed = uint8(epoch)
			}
			continue
		}
		// Ground below the water it drains into is ground being filled in,
		// and rock made of what is falling on it now dates from now.
		if t.Drain < FloodDepth/2 {
			book[i].laid[Sand] += t.Sand
			book[i].laid[Silt] += t.Silt()
			book[i].laid[Clay] += t.Clay
			t.Formed = uint8(epoch)
		}
	}
}

// settleRock is the history read back as geology: what a tile is made of,
// given everything that happened to it. The order is the order that decides
// it - what came up as melt is what it is, whatever was done to it after;
// short of that, what was cooked and squeezed by a collision; short of that,
// what it was buried under; and short of everything at all, whatever the
// plate it rides was made of - basalt if it is ocean floor, granite if it is
// the old body of a continent.
func (g *Grid) settleRock(book []record, plates []Plate, epochs int) {
	// Where the line between a coarse fill and a fine one falls on this
	// world, read off its own fills rather than fixed. See coarseShare.
	sandy := make([]float64, 0, len(g.Tiles))
	for i := range g.Tiles {
		if fill := carrying(book[i].laid); fill > fillEnough {
			sandy = append(sandy, book[i].laid[Sand]/fill)
		}
	}
	coarse := math.Inf(1)
	if len(sandy) > 0 {
		coarse = quantile(sandy, 1-coarseShare)
	}

	for i := range g.Tiles {
		t := &g.Tiles[i]
		b := book[i]
		fill := carrying(b.laid)
		switch {
		case b.melt > b.crush && b.melt > b.pluton && b.melt > fill:
			// It came up and cooled in the air.
			t.Bedrock = Basalt
		case b.pluton > b.crush && b.pluton > fill:
			// It melted under an arc and cooled at depth, and the weather has
			// since taken off what stood over it. This is where granite comes
			// from, and saying so is what gave the rock a place on the map at
			// all: as the leftover case - ground nothing ever happened to -
			// it never came up once in sixteen epochs, because something
			// happens to everything.
			t.Bedrock = Granite
		case b.crush > fill && b.crush > 0:
			t.Bedrock = Schist
		case fill > fillEnough && b.laid[Sand]/fill >= coarse:
			// Coarse fill: the near end of a basin, where what came off the
			// hill did not travel far before it was dropped.
			t.Bedrock = Sandstone
		case fill > fillEnough:
			// Fine fill: what did travel, and the mud off a shore.
			t.Bedrock = Shale
		case b.submerged > epochs/2:
			// Ground that lay under water for most of a history, with no
			// river reaching it, is where limestone comes from: what settles
			// there is what lived there.
			t.Bedrock = Limestone
		case plates[t.Plate].Ocean:
			// Ocean floor nothing ever happened to is the basalt it cooled
			// as, and the oldest rock on the map.
			t.Bedrock = Basalt
		default:
			// The old body of a continent, showing through.
			t.Bedrock = Granite
		}
	}
}

// normalise brings a history's relief back to the scale the rest of the world
// is built to, and this is the join between a world that made itself and a
// world that has to be liveable. It is the part most likely to be wrong, so
// it is here in the open rather than folded into the passes above.
//
// The history says where the high ground is; the drawn map says how high a
// map's ground is spread. Every tile keeps its place in the order - the
// hundredth-highest tile of a history is the hundredth-highest tile of the
// map that comes out - and takes the height the drawn generator would have
// put at that place in its own order. So the causes are the history's and
// the scale is the one every constant downstream was measured against: the
// sixty metres of lowland, the high country standing on a fifth of it,
// fourteen metres to the top of a flood plain.
//
// Stretching each band onto its own range was the first way this was written,
// and it made plains and cliffs: the lowland was squeezed four to one while
// the mountains were left at one to one, so half the map came out at a slope
// of a hundredth and the steepest tenth at twice what a drawn map has.
// Matching the whole spread rather than its ends is what fixed that.
func (w *World) normalise(g *Grid) {
	spread := w.relief(g)
	sort.Float64s(spread)

	order := make([]int32, len(g.Tiles))
	for i := range order {
		order[i] = int32(i)
	}
	sort.Slice(order, func(a, b int) bool {
		ha, hb := g.Tiles[order[a]].Height, g.Tiles[order[b]].Height
		if ha != hb {
			return ha < hb
		}
		return order[a] < order[b] // ties by position, so a world repeats
	})
	for rank, i := range order {
		g.Tiles[i].Height = spread[rank]
	}
}

// soften eases the finished ground: the same blur the river valleys are cut
// with, run over the heights. It is the one pass here that is not a process -
// nothing in the earth averages a hillside with its neighbours - and it is
// here because a history raises narrower ranges than a drawn map does. What
// it costs is sharpness and not scale, since the matching afterwards hands
// out the same heights either way. See smoothing.
func (g *Grid) soften() {
	h := make([]float64, len(g.Tiles))
	for i := range g.Tiles {
		h[i] = g.Tiles[i].Height
	}
	h = g.spread(h)
	for i := range g.Tiles {
		g.Tiles[i].Height = h[i]
	}
}

// offshore reports whether any of the eight tiles around p stands above the
// sea. It is how a sea bed is told from a shore: what is next to land gets
// what the land sends it.
func (g *Grid) offshore(p entity.Pos, sea float64) bool {
	for _, off := range dirs {
		q := entity.Pos{X: p.X + off.X, Y: p.Y + off.Y}
		if g.Wrap {
			q = g.Norm(q)
		}
		if g.In(q) && !g.At(q).Wet() && g.At(q).Height > sea {
			return true
		}
	}
	return false
}

// seaLevel is where the water stands this epoch: up from where the ocean
// floor is riding toward where the continents are, by historySea of the way.
// Both kinds of crust are always present - see firstPlates - so there is
// always a level that drowns some floor and leaves some land dry.
func (g *Grid) seaLevel(plates []Plate) float64 {
	var deep, high, deepN, highN float64
	for i := range g.Tiles {
		if plates[g.Tiles[i].Plate].Ocean {
			deep += g.Tiles[i].Height
			deepN++
		} else {
			high += g.Tiles[i].Height
			highN++
		}
	}
	if deepN == 0 || highN == 0 {
		return math.Inf(-1)
	}
	deep, high = deep/deepN, high/highN
	return deep + historySea*(high-deep)
}

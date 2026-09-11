// Package observe is the read side of the simulation.
//
// Snapshot is the immutable view published every tick for renderers, tests,
// and headless tooling. Perceive is the player's perception layer: a filter
// over the event log that returns only what a given agent could know.
package observe

import (
	"slices"
	"sort"
	"sync"

	"lreat/core/action"
	"lreat/core/belief"
	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/world"
)

// Activity counts agents currently committed to an action.
type Activity struct {
	Action string
	Agents int
}

// Mark is an agent's position and what it is doing, for the map. The ID is
// there so that a view can follow one particular figure from tick to tick
// rather than whoever happens to be standing where it was.
type Mark struct {
	ID     entity.ID
	Pos    entity.Pos
	Action string
	// Kind is what kind of creature stands here, by its species' name, so
	// that a map can draw a deer as a deer.
	Kind string
}

// MapView is a copy of the grid plus where everyone is.
type MapView struct {
	W, H int
	// Wrap is the shape of the ground this was taken off: a globe's east
	// edge is joined to its west, and anything drawn from these tiles has
	// to go round it the same way the world does. Without it a viewer
	// reading a tile's neighbours gets the seam wrong, and one looking
	// eastward past the last column falls off a map that has no edge.
	Wrap  bool
	Tiles []world.Tile
	// Layers is the ground that changes by the day, copied beside the
	// tiles; see world.Layers.
	world.Layers
	Agents []Mark
	Market entity.Pos
}

// At returns the tile at p.
func (m *MapView) At(p entity.Pos) *world.Tile { return &m.Tiles[p.Y*m.W+p.X] }

// Snapshot is a copy of the aggregate state at one tick.
type Snapshot struct {
	// Tick is the day the world has reached, and Date is that day said as a
	// calendar date. Everything a settlement does is easier to read against
	// the second than the first: a famine in the winter of year nine means
	// something, and a famine at tick 3140 does not.
	Tick int
	Date clock.Date
	// Population is the people: everyone who settles. Creatures is
	// everything else alive on the map, which is on the settlement's map
	// and not on its books, so none of the means below count it.
	Population int
	Creatures  int
	MeanNeeds  need.Levels
	// MeanHealth is the population's average condition. It moves slowly, so
	// a settlement that is wearing its people down shows here long before it
	// shows in the death log.
	MeanHealth float64
	// MeanAge, in years, and Elders describe the shape of the generations: a
	// settlement of the old is one that has stopped replacing itself,
	// whatever its headcount says today.
	MeanAge    int
	Elders     int        // agents past their prime
	Activity   []Activity // most common first
	WealthGini float64
	Knowledge  float64
	Techs      []world.Tech
	// Worked is the same technologies with the dates on them: when each was
	// worked out and when the settlement first had a master of its craft.
	// Techs is kept beside it because most readers only want the names.
	Worked    []Worked
	Safety    float64
	FoodPrice float64
	Events    int
	// The weather. Temp is this tick's temperature and Season names the
	// quarter of the year it falls in; Growth is what the season lets the
	// land put back, 1 being an ordinary year's average.
	Temp   float64
	Season clock.Quarter
	Growth float64

	// The moral and contractual state of the settlement. MeanNorms is what
	// this population currently holds to be right, which drifts on its own.
	MeanNorms     belief.Norms
	OpenRequests  int
	NamedRequests int // asked of someone in particular rather than of anyone

	// The shape of the social graph. Friendships and Feuds are mutual;
	// Hearsay counts opinions held about people never met in person.
	Friendships int
	Feuds       int
	Hearsay     int

	// The habit layer. HabitSpread is how far agents' habits lie apart, 0
	// when everyone recognises the same moments the same way. It is what
	// birth, teaching, and inheritance have made of one shared table, and
	// it does not move within a life. GatedReach is how far the crafts and
	// learning have come within reach on average. ChoiceEntropy is how open this tick's decisions were, in
	// nats, 0 when every choice was certain. Deaths is cumulative.
	HabitSpread   float64
	MeanReach     float64
	GatedReach    float64
	ChoiceEntropy float64
	Deaths        int

	// The demographic record. Vitals says what the deaths were of and what
	// stood between everyone else and a child; Chronicle is the last of the
	// births, deaths, and discoveries in the settlement's own words. A
	// population curve says when a settlement died out. These two are how
	// it is told why.
	Vitals    world.Vitals
	Chronicle []world.Note

	// The shape of the generations under the headcount. Children have not
	// grown up yet and Bearing are in their fertile years; Elders is above.
	// A settlement can be at full strength and already finished, if none of
	// the strength is of bearing age.
	Children int
	Bearing  int
	// Starving is how many are at the bottom of the physiological tier
	// right now, each of them on a clock that runs out in
	// system.StarvationTicks. It moves hundreds of ticks ahead of the
	// deaths it becomes.
	Starving int
	// What there is to live on. FoodStock is the market's, MeanFood is what
	// the average agent is carrying, MeanShelter how much roof it has.
	FoodStock   float64
	MeanFood    float64
	MeanShelter float64

	Houses, Fields, Forest, Roads int
	// Forest0 is how much forest there was before anyone touched it. Beside
	// Forest it says what the settlement has taken out of the land, which
	// the count on its own never can.
	Forest0 int
	Map     *MapView
}

// Take builds a Snapshot. It must run on the simulation goroutine: it reads
// the whole world, and nothing may change under it while it does.
//
// Three things are read out and none of them reads what another writes. The
// ground is copied, which on a globe is seventy megabytes and is memory
// rather than arithmetic; the population is counted up; and the habit layer
// is measured. So all three are done at once, each on a goroutine of its
// own, and a snapshot costs the longest of them rather than the sum. Each
// keeps its own order within itself - the ground by tile, the population by
// agent, the habits by act and then by agent - so every number here is to
// the last bit the number it was when they were done one after another.
func Take(w *world.World) Snapshot {
	s := Snapshot{
		Tick:       w.Tick,
		Date:       clock.At(w.Tick),
		Population: w.People(),
		Knowledge:  w.Knowledge,
		Techs:      w.Techs(),
		Worked:     worked(w),
		Safety:     w.Safety,
		FoodPrice:  w.Market.Price[entity.Food],
		Temp:       w.Climate.Temp,
		Season:     world.SeasonOf(w.Tick),
		Growth:     w.Climate.Growth(),
		Events:     w.Log.Len(),
		Houses:     w.Grid.Houses(),
		Fields:     w.Grid.Fields(),
		Forest:     w.Grid.Forest(),
		Roads:      w.Grid.Roads(),
		Forest0:    w.Forest0,

		OpenRequests: len(w.Requests),
		Deaths:       w.Deaths,
		Vitals:       w.Vitals,
		FoodStock:    w.Market.Stock[entity.Food],
	}
	s.Chronicle = append(s.Chronicle, w.Chronicle...)
	if w.Choices > 0 {
		s.ChoiceEntropy = w.Entropy / float64(w.Choices)
	}
	for _, r := range w.Requests {
		if r.Directed != 0 {
			s.NamedRequests++
		}
	}

	m := &MapView{W: w.Grid.W, H: w.Grid.H, Wrap: w.Grid.Wrap, Market: w.MarketPos}
	s.Map = m
	if len(w.Agents) == 0 {
		m.Tiles, m.Layers = ground(w)
		return s
	}

	// Measuring the habit layer wants the room grown to the slots there now
	// are. It is the one write in any of this, so it is done here, before
	// anybody is reading beside anybody else.
	w.Room()
	// The file of who is where is put right and then held still: counting
	// the population looks people up in it, and a reader that put it right
	// would be writing where the others are looking. See world.Freeze.
	w.Freeze(true)
	var c census
	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); m.Tiles, m.Layers = ground(w) }()
	go func() { defer wg.Done(); c = count(w) }()
	go func() { defer wg.Done(); s.HabitSpread, s.MeanReach, s.GatedReach = habits(w) }()
	wg.Wait()
	w.Freeze(false)

	m.Agents = c.marks
	s.Creatures = c.creatures
	if c.people == 0 {
		return s
	}
	n := float64(c.people)
	s.MeanNeeds = c.needs
	s.MeanNorms = c.norms
	s.MeanHealth, s.MeanShelter, s.MeanFood = c.health/n, c.shelter/n, c.food/n
	s.MeanAge = clock.Years(c.age / c.people)
	s.Starving, s.Children, s.Bearing, s.Elders = c.starving, c.children, c.bearing, c.elders
	s.Friendships, s.Feuds, s.Hearsay = c.friendships, c.feuds, c.hearsay
	for t := range s.MeanNeeds {
		s.MeanNeeds[t] /= n
	}
	for i := range s.MeanNorms {
		s.MeanNorms[i] /= n
	}
	for name, k := range c.counts {
		s.Activity = append(s.Activity, Activity{Action: name, Agents: k})
	}
	sort.Slice(s.Activity, func(i, j int) bool {
		if s.Activity[i].Agents != s.Activity[j].Agents {
			return s.Activity[i].Agents > s.Activity[j].Agents
		}
		return s.Activity[i].Action < s.Activity[j].Action
	})
	s.WealthGini = Gini(c.wealth)
	return s
}

// ground is the copy of the map a snapshot carries. It is cloned rather than
// made and copied into: making a slice zeroes it, and the copy that follows
// writes over every byte of the zeroes, so the ground was being walked twice
// for a picture of it taken once. On a globe that was five milliseconds of
// every snapshot. The layers come with it, copied the same way.
func ground(w *world.World) ([]world.Tile, world.Layers) {
	return slices.Clone(w.Grid.Tiles), w.Grid.Layers.Copy()
}

// census is what one pass over the population comes to, gathered up so that
// the pass can be made beside the copying of the ground rather than after
// it. The means are still sums here; who divides them is Take.
type census struct {
	needs                 need.Levels
	norms                 belief.Norms
	health, shelter, food float64
	age                   int
	starving              int
	children, bearing     int
	elders                int
	friendships, feuds    int
	hearsay               int
	people, creatures     int
	marks                 []Mark
	wealth                []float64
	counts                map[string]int
}

// count walks the population once and adds up everything a snapshot says
// about it. It only reads, and it reads the population in agent order, which
// is the order the sums were always made in.
func count(w *world.World) census {
	c := census{
		counts: map[string]int{},
		marks:  make([]Mark, 0, len(w.Agents)),
		wealth: make([]float64, 0, len(w.Agents)),
	}
	for _, a := range w.Agents {
		mark := Mark{ID: a.ID, Pos: a.Pos, Kind: a.Species().Name}
		if a.Plan != nil {
			mark.Action = a.Plan.Action
		}
		c.marks = append(c.marks, mark)
		// A creature is on the map and nowhere else: nothing about it
		// enters what is said of the people.
		if !a.Species().Settles {
			c.creatures++
			continue
		}
		c.people++
		for t := range c.needs {
			c.needs[t] += a.Needs[t]
		}
		for i := range c.norms {
			c.norms[i] += a.Norms[i]
		}
		c.health += a.Health
		c.shelter += a.Shelter
		c.food += a.Inventory[entity.Food] + a.Inventory[entity.Meals]
		if a.Starving > 0 {
			c.starving++
		}
		age := a.Age(w.Tick)
		c.age += age
		switch {
		case age < entity.Maturity:
			c.children++
		case age < entity.Prime:
			c.bearing++
		default:
			c.elders++
		}
		if a.Plan != nil {
			c.counts[a.Plan.Action]++
		}
		c.wealth = append(c.wealth, a.Wealth)

		for i := range a.Bonds {
			b := &a.Bonds[i]
			if b.Met == 0 {
				c.hearsay++
			}
			if b.To < a.ID {
				continue // count each pair once
			}
			o := w.Find(b.To)
			if o == nil {
				continue
			}
			back := o.Look(a.ID)
			if back == nil {
				continue
			}
			if b.Strength > 0.5 && back.Strength > 0.5 {
				c.friendships++
			}
			if b.Regard < -0.3 && back.Regard < -0.3 {
				c.feuds++
			}
		}
	}
	return c
}

// habits measures the habit layer: how far apart agents' recognition lies,
// and how far the gated actions have come within reach. Agents not
// yet imprinted are read as holding the priors.
//
// It only reads. The room it measures against has to have been grown first -
// see World.Room - because growing it is a write, and this runs beside the
// other two passes a snapshot makes.
func habits(w *world.World) (spread, mean, gated float64) {
	n := float64(w.People())
	if n == 0 {
		return 0, 0, 0
	}
	// The people's acts, over the people: a creature's slots are empty in
	// everybody's tables and a creature holds nothing on the people's.
	mine := action.For(entity.Human)
	var gatedN float64
	// Each act is measured in two passes over the population rather than
	// from a table of every unit signature: the centre first, then how far
	// each agent stands from it. The unit signature is worked out twice for
	// that, which is cheaper than keeping thirty of them per agent per tick
	// and adds the same numbers in the same order, so the measure is the
	// one it always was.
	for _, i := range mine {
		d := action.Catalog[i]
		var centre habit.Signature
		for _, a := range w.Agents {
			if !a.Species().Settles {
				continue
			}
			h, r := d.Prior, max(d.Reach0, w.ReachFloor[i])
			if a.Imprinted {
				h, r = a.Habits[i], max(a.Reach[i], w.ReachFloor[i])
			}
			u := habit.Unit(h)
			for k := range centre {
				centre[k] += u[k] / n
			}
			mean += r
			if d.Reach0 < 1 {
				gated += r
				gatedN++
			}
		}
		for _, a := range w.Agents {
			if !a.Species().Settles {
				continue
			}
			h := d.Prior
			if a.Imprinted {
				h = a.Habits[i]
			}
			u := habit.Unit(h)
			var diff habit.Signature
			for k := range diff {
				diff[k] = u[k] - centre[k]
			}
			spread += habit.Norm(diff)
		}
	}
	mean /= n * float64(len(mine))
	if gatedN > 0 {
		gated /= gatedN
	}
	spread /= n * float64(len(mine))
	return spread, mean, gated
}

// Gini returns the Gini coefficient of a distribution, 0 for perfect equality.
func Gini(values []float64) float64 {
	n := len(values)
	if n == 0 {
		return 0
	}
	sorted := make([]float64, n)
	copy(sorted, values)
	sort.Float64s(sorted)
	var cum, total float64
	for i, v := range sorted {
		total += v
		cum += v * float64(2*(i+1)-n-1)
	}
	if total == 0 {
		return 0
	}
	return cum / (float64(n) * total)
}

// Perceive returns events since tick that viewer could know about: things
// they did, things done to them, things done by people they are bonded to,
// and public events such as deaths and discoveries. Everything else stays
// hidden. The player learns about the world through this and nothing else.
func Perceive(w *world.World, viewer entity.ID, since int) []event.Event {
	v := w.Find(viewer)
	if v == nil {
		return nil
	}
	var out []event.Event
	for _, e := range w.Log.Since(since) {
		switch {
		case e.Kind == event.Discovered || e.Kind == event.Died:
			out = append(out, e)
		case e.Actor == viewer || e.Target == viewer:
			out = append(out, e)
		case e.Actor != 0 && v.BondWith(e.Actor) > 0.3:
			out = append(out, e)
		}
	}
	return out
}

// Worked is one technology and what has become of it. The dates are ticks;
// clock.At turns them into something worth reading, which is the whole
// reason they are carried rather than left as a count.
type Worked struct {
	Tech     world.Tech
	Found    int
	Mastered int // zero until somebody is a master of its craft
}

// worked reads the dates off the world in the order Techs gives them, so
// that two snapshots of the same settlement list them the same way.
func worked(w *world.World) []Worked {
	techs := w.Techs()
	if len(techs) == 0 {
		return nil
	}
	out := make([]Worked, 0, len(techs))
	for _, t := range techs {
		k := w.Known(t)
		out = append(out, Worked{Tech: t, Found: k.Found, Mastered: k.Mastered})
	}
	return out
}

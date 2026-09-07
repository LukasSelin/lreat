// Package observe is the read side of the simulation.
//
// Snapshot is the immutable view published every tick for renderers, tests,
// and headless tooling. Perceive is the player's perception layer: a filter
// over the event log that returns only what a given agent could know.
package observe

import (
	"sort"

	"lreat/core/action"
	"lreat/core/belief"
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
}

// MapView is a copy of the grid plus where everyone is.
type MapView struct {
	W, H   int
	Tiles  []world.Tile
	Agents []Mark
	Market entity.Pos
}

// At returns the tile at p.
func (m *MapView) At(p entity.Pos) *world.Tile { return &m.Tiles[p.Y*m.W+p.X] }

// Snapshot is a copy of the aggregate state at one tick.
type Snapshot struct {
	Tick       int
	Population int
	MeanNeeds  need.Levels
	// MeanHealth is the population's average condition. It moves slowly, so
	// a settlement that is wearing its people down shows here long before it
	// shows in the death log.
	MeanHealth float64
	// MeanAge and Elders describe the shape of the generations: a settlement
	// of the old is one that has stopped replacing itself, whatever its
	// headcount says today.
	MeanAge    int
	Elders     int        // agents past their prime
	Activity   []Activity // most common first
	WealthGini float64
	Knowledge  float64
	Techs      []world.Tech
	Safety     float64
	FoodPrice  float64
	Events     int
	// The weather. Temp is this tick's temperature and Season names the
	// quarter of the year it falls in; Growth is what the season lets the
	// land put back, 1 being an ordinary year's average.
	Temp   float64
	Season string
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

	Houses, Fields, Forest, Roads int
	Map                           *MapView
}

// Take builds a Snapshot. It must run on the simulation goroutine.
func Take(w *world.World) Snapshot {
	s := Snapshot{
		Tick:       w.Tick,
		Population: len(w.Agents),
		Knowledge:  w.Knowledge,
		Techs:      w.Techs(),
		Safety:     w.Safety,
		FoodPrice:  w.Market.Price[entity.Food],
		Temp:       w.Climate.Temp,
		Season:     world.SeasonOf(w.Tick),
		Growth:     w.Climate.Growth(),
		Events:     w.Log.Len(),
		Houses:     w.Grid.Count(func(t *world.Tile) bool { return t.Structure == world.House }),
		Fields:     w.Grid.Count(func(t *world.Tile) bool { return t.Terrain == world.Field }),
		Forest:     w.Grid.Count(func(t *world.Tile) bool { return t.Terrain == world.Forest }),
		Roads:      w.Grid.Count(func(t *world.Tile) bool { return t.Structure == world.Road }),

		OpenRequests: len(w.Requests),
		Deaths:       w.Deaths,
	}
	if w.Choices > 0 {
		s.ChoiceEntropy = w.Entropy / float64(w.Choices)
	}
	for _, r := range w.Requests {
		if r.Directed != 0 {
			s.NamedRequests++
		}
	}

	m := &MapView{W: w.Grid.W, H: w.Grid.H, Tiles: make([]world.Tile, len(w.Grid.Tiles)), Market: w.MarketPos}
	copy(m.Tiles, w.Grid.Tiles)
	s.Map = m

	if len(w.Agents) == 0 {
		return s
	}

	counts := map[string]int{}
	wealth := make([]float64, 0, len(w.Agents))
	m.Agents = make([]Mark, 0, len(w.Agents))
	for _, a := range w.Agents {
		for t := range s.MeanNeeds {
			s.MeanNeeds[t] += a.Needs[t]
		}
		for n := range s.MeanNorms {
			s.MeanNorms[n] += a.Norms[n]
		}
		s.MeanHealth += a.Health
		age := a.Age(w.Tick)
		s.MeanAge += age
		if age >= entity.Prime {
			s.Elders++
		}
		mark := Mark{ID: a.ID, Pos: a.Pos}
		if a.Plan != nil {
			counts[a.Plan.Action]++
			mark.Action = a.Plan.Action
		}
		m.Agents = append(m.Agents, mark)
		wealth = append(wealth, a.Wealth)

		for i := range a.Bonds {
			b := &a.Bonds[i]
			if b.Met == 0 {
				s.Hearsay++
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
				s.Friendships++
			}
			if b.Regard < -0.3 && back.Regard < -0.3 {
				s.Feuds++
			}
		}
	}
	for t := range s.MeanNeeds {
		s.MeanNeeds[t] /= float64(len(w.Agents))
	}
	s.MeanHealth /= float64(len(w.Agents))
	s.MeanAge /= len(w.Agents)
	for n := range s.MeanNorms {
		s.MeanNorms[n] /= float64(len(w.Agents))
	}
	for name, n := range counts {
		s.Activity = append(s.Activity, Activity{Action: name, Agents: n})
	}
	sort.Slice(s.Activity, func(i, j int) bool {
		if s.Activity[i].Agents != s.Activity[j].Agents {
			return s.Activity[i].Agents > s.Activity[j].Agents
		}
		return s.Activity[i].Action < s.Activity[j].Action
	})
	s.WealthGini = Gini(wealth)
	s.HabitSpread, s.MeanReach, s.GatedReach = habits(w)
	return s
}

// habits measures the habit layer: how far apart agents' recognition lies,
// and how far the gated actions have come within reach. Agents not
// yet imprinted are read as holding the priors.
func habits(w *world.World) (spread, mean, gated float64) {
	n := float64(len(w.Agents))
	if n == 0 {
		return 0, 0, 0
	}
	units := make([][]habit.Signature, action.Count)
	var gatedN float64
	w.Room()
	for i, d := range action.Catalog {
		units[i] = make([]habit.Signature, 0, len(w.Agents))
		for _, a := range w.Agents {
			h, r := d.Prior, max(d.Reach0, w.ReachFloor[i])
			if a.Imprinted {
				h, r = a.Habits[i], max(a.Reach[i], w.ReachFloor[i])
			}
			units[i] = append(units[i], habit.Unit(h))
			mean += r
			if d.Reach0 < 1 {
				gated += r
				gatedN++
			}
		}
	}
	mean /= n * float64(action.Count)
	if gatedN > 0 {
		gated /= gatedN
	}
	for i := range units {
		var centre habit.Signature
		for _, u := range units[i] {
			for k := range centre {
				centre[k] += u[k] / n
			}
		}
		for _, u := range units[i] {
			var d habit.Signature
			for k := range d {
				d[k] = u[k] - centre[k]
			}
			spread += habit.Norm(d)
		}
	}
	spread /= n * float64(action.Count)
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

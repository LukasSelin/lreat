// Package world holds the complete simulation state.
//
// Exactly one goroutine may drive a World. The sim package enforces that;
// everything else assumes it. Within a tick, deciding is spread over several
// goroutines - it only reads - while everything that changes the world runs
// one at a time.
//
// Randomness flows through World.RNG, except for what an agent draws while
// deciding, which comes from that agent's own Luck. Both are seeded from the
// world seed, so a seed plus a command log still reproduces a run exactly,
// and it reproduces it whatever the goroutines do.
package world

import (
	"fmt"
	"math/rand/v2"
	"sort"

	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/habit"
	"lreat/core/need"
)

// DefaultWidth and DefaultHeight size the map when none is given. They fit a
// standard terminal beside a stats panel.
const (
	DefaultWidth  = 80
	DefaultHeight = 36
)

// Tech is a capability the settlement has discovered. Nobody chooses a Tech;
// the discovery system detects the conditions for it.
type Tech string

// Modifiers are the knobs technologies turn.
type Modifiers struct {
	FarmYield       float64
	BuildEfficiency float64
	StudyRate       float64
	CraftQuality    float64
	ShelterDecay    float64
	FishYield       float64
	HuntYield       float64
	Regrowth        float64 // how fast forest, wild food, and fish come back
	Keeping         float64 // how much of the market's food spoils, as a share of the usual
}

// DefaultModifiers is the pre-technology baseline.
func DefaultModifiers() Modifiers {
	return Modifiers{
		FarmYield:       1,
		BuildEfficiency: 1,
		StudyRate:       1,
		CraftQuality:    1,
		ShelterDecay:    0.002,
		FishYield:       1,
		HuntYield:       1,
		Regrowth:        1,
		Keeping:         1,
	}
}

// Rules selects how agents choose. Unlike Modifiers these are never changed
// by anything inside the simulation; they are set once when a world is made.
type Rules struct {
	// Fit makes agents choose by recognition, sampling the action whose
	// habit best fits the moment, instead of by expected value. See package
	// habit.
	Fit bool
	// Temperature is the base softmax temperature of fit-based choice. Zero
	// takes the best fit every time.
	Temperature float64
}

// DefaultRules is recognition. The value rule, the original, stays behind
// the flag for comparison. See docs/action-space.md.
func DefaultRules() Rules {
	return Rules{Fit: true, Temperature: 0.15}
}

// MarketState is a single shared marketplace with a stock and a price per good.
type MarketState struct {
	Stock [entity.GoodCount]float64
	Price [entity.GoodCount]float64
}

// World is the whole simulation.
type World struct {
	Tick   int
	RNG    *rand.Rand
	Agents []*entity.Agent

	Grid      *Grid
	MarketPos entity.Pos

	Market    MarketState
	Climate   Climate // the weather over the whole map this tick
	Safety    float64 // public safety in [0,1], raised by guarding, decays
	Knowledge float64 // accumulated by study, consumed by nothing
	Mods      Modifiers
	Rules     Rules
	Log       *event.Log

	// Requests is the open board of work one agent wants another to do.
	// Nothing here is authored; the request system posts and clears it.
	Requests []*entity.Request

	// Forest0 is how much forest the world was made with, so that how much
	// of it a settlement has taken can be told.
	Forest0 int

	// ReachFloor is how far into reach each action is for everyone here,
	// by habit slot, raised when the settlement discovers a thing. Room
	// keeps it as long as there are slots.
	ReachFloor []float64
	// Deaths counts everyone who has died here.
	Deaths int
	// Choices and Entropy record this tick's fit-based decisions: how many
	// were made and how open they were in total, for observation.
	Choices int
	Entropy float64

	// watched is the agent whose deliberations are being kept, and thoughts
	// is what it has been weighing. Both are for observation only; see
	// watch.go.
	watched  entity.ID
	thoughts []Deliberation

	techs     map[Tech]bool
	nextID    entity.ID
	nextReqID entity.RequestID

	// routers is the working memory deciding routes on, one per goroutine.
	routers []*Router
}

// New creates a world with default-sized terrain, seeded for determinism.
func New(seed uint64) *World {
	return NewSized(seed, DefaultWidth, DefaultHeight)
}

// NewSized creates a world with terrain of the given size.
func NewSized(seed uint64, width, height int) *World {
	w := &World{
		RNG:     rand.New(rand.NewPCG(seed, seed*0x9E3779B97F4A7C15+1)),
		Mods:    DefaultModifiers(),
		Rules:   DefaultRules(),
		Climate: NewClimate(),
		Market: MarketState{
			Price: [entity.GoodCount]float64{1, 0.5, 3, 1.5, 2},
		},
		Log:    event.NewLog(50_000),
		techs:  map[Tech]bool{},
		nextID: 1,
	}
	w.GenerateTerrain(width, height)
	w.Room()
	return w
}

// Room keeps the reach floor as long as there are slots.
func (w *World) Room() { w.ReachFloor = habit.Grow(w.ReachFloor, habit.Slots()) }

// Spawn adds an agent on open ground near the market.
func (w *World) Spawn(name string, p need.Weights) *entity.Agent {
	pos := w.MarketPos
	for try := 0; try < 20; try++ {
		c := entity.Pos{X: w.MarketPos.X + w.RNG.IntN(9) - 4, Y: w.MarketPos.Y + w.RNG.IntN(9) - 4}
		if w.Grid.In(c) && w.Grid.At(c).Terrain != Water {
			pos = c
			break
		}
	}
	return w.SpawnAt(name, p, pos)
}

// SpawnAt adds an agent at a specific position. Spawned agents are young
// adults of assorted ages: grown into their bodies, with most of their
// fertile years ahead of them, and not all due to reach old age in the same
// week. A founding party is people who set out, not a nursery and not a
// retirement. Only birth makes a newborn: the population system sets the
// child's Born itself.
func (w *World) SpawnAt(name string, p need.Weights, pos entity.Pos) *entity.Agent {
	a := &entity.Agent{
		ID:          w.nextID,
		Name:        name,
		Luck:        rand.New(rand.NewPCG(w.RNG.Uint64(), w.RNG.Uint64())),
		Born:        w.Tick - entity.Maturity - w.RNG.IntN((entity.Prime-entity.Maturity)/3),
		Pos:         pos,
		Needs:       need.Levels{0.7, 0.3, 0.5, 0.3, 0.2},
		Personality: p,
		Norms:       belief.RandomNorms(w.RNG),
		Temperament: entity.RandomTemperament(w.RNG),
		Vitality:    w.RandomVitality(),
		Health:      0.9,
	}
	// Everyone starts believing they are unremarkable. Confidence is earned
	// by doing, and can outrun or lag the skill it is meant to describe.
	for s := range a.Efficacy {
		a.Efficacy[s] = 0.15
	}
	a.Inventory[entity.Food] = 2
	w.nextID++
	w.Agents = append(w.Agents, a)
	a.Room()
	return a
}

// RandomPersonality draws per-tier weights around neutral. This is the main
// source of heterogeneity in the population.
func (w *World) RandomPersonality() need.Weights {
	var p need.Weights
	for i := range p {
		p[i] = clampWeight(1 + w.RNG.NormFloat64()*0.3)
	}
	return p
}

// RandomVitality draws a body around the ordinary one. The spread is narrow
// on purpose: bodies differ, but a settlement's fortunes should turn on what
// people want and believe, not on who was born strong.
func (w *World) RandomVitality() float64 {
	return clampVitality(1 + w.RNG.NormFloat64()*0.12)
}

// InheritVitality returns a child's body derived from a parent's.
func (w *World) InheritVitality(v float64) float64 {
	if v <= 0 {
		return w.RandomVitality()
	}
	return clampVitality(v + w.RNG.NormFloat64()*0.08)
}

func clampVitality(v float64) float64 {
	if v < 0.7 {
		return 0.7
	}
	if v > 1.3 {
		return 1.3
	}
	return v
}

// Mutate returns a child's personality derived from a parent's.
func (w *World) Mutate(p need.Weights) need.Weights {
	for i := range p {
		p[i] = clampWeight(p[i] + w.RNG.NormFloat64()*0.15)
	}
	return p
}

func clampWeight(v float64) float64 {
	if v < 0.3 {
		return 0.3
	}
	if v > 2 {
		return 2
	}
	return v
}

// Other picks a random agent that is not a. Returns nil if a is alone. The
// draw comes from a's own luck, because this is reached while deciding, which
// several agents may be doing at once.
func (w *World) Other(a *entity.Agent) *entity.Agent {
	if len(w.Agents) < 2 {
		return nil
	}
	for {
		o := w.Agents[a.Luck.IntN(len(w.Agents))]
		if o != a {
			return o
		}
	}
}

// Routers returns n routers over the world's map, made once and kept between
// ticks so that deciding allocates nothing. Each is for one goroutine.
func (w *World) Routers(n int) []*Router {
	for len(w.routers) < n {
		w.routers = append(w.routers, w.Grid.Router())
	}
	return w.routers[:n]
}

// Neighbor returns the closest other agent within radius tiles of a, or nil.
// Ties go to the agent earliest in the slice, which keeps runs deterministic.
func (w *World) Neighbor(a *entity.Agent, radius int) *entity.Agent {
	var best *entity.Agent
	bestD := radius + 1
	for _, o := range w.Agents {
		if o == a {
			continue
		}
		if d := entity.Dist(a.Pos, o.Pos); d < bestD {
			best, bestD = o, d
		}
	}
	return best
}

// Find returns the agent with the given id, or nil.
func (w *World) Find(id entity.ID) *entity.Agent {
	for _, a := range w.Agents {
		if a.ID == id {
			return a
		}
	}
	return nil
}

// Emit appends an event at the current tick.
func (w *World) Emit(kind event.Kind, actor, target entity.ID, format string, args ...any) {
	w.Log.Append(event.Event{
		Tick:   w.Tick,
		Kind:   kind,
		Actor:  actor,
		Target: target,
		Text:   fmt.Sprintf(format, args...),
	})
}

// Has reports whether a technology has been discovered.
func (w *World) Has(t Tech) bool { return w.techs[t] }

// Unlock marks a technology as discovered.
func (w *World) Unlock(t Tech) { w.techs[t] = true }

// Techs returns discovered technologies in a stable order.
func (w *World) Techs() []Tech {
	out := make([]Tech, 0, len(w.techs))
	for t := range w.techs {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

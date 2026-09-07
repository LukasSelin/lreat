// Package world holds the complete simulation state.
//
// Exactly one goroutine may touch a World at a time. The sim package enforces
// that; everything else assumes it. All randomness flows through World.RNG so
// that a seed plus a command log reproduces a run exactly.
package world

import (
	"fmt"
	"math/rand/v2"
	"sort"

	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/event"
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
}

// DefaultModifiers is the pre-technology baseline.
func DefaultModifiers() Modifiers {
	return Modifiers{
		FarmYield:       1,
		BuildEfficiency: 1,
		StudyRate:       1,
		CraftQuality:    1,
		ShelterDecay:    0.002,
	}
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
	Safety    float64 // public safety in [0,1], raised by guarding, decays
	Knowledge float64 // accumulated by study, consumed by nothing
	Mods      Modifiers
	Log       *event.Log

	// Requests is the open board of work one agent wants another to do.
	// Nothing here is authored; the request system posts and clears it.
	Requests []*entity.Request

	techs     map[Tech]bool
	nextID    entity.ID
	nextReqID entity.RequestID
}

// New creates a world with default-sized terrain, seeded for determinism.
func New(seed uint64) *World {
	return NewSized(seed, DefaultWidth, DefaultHeight)
}

// NewSized creates a world with terrain of the given size.
func NewSized(seed uint64, width, height int) *World {
	w := &World{
		RNG:  rand.New(rand.NewPCG(seed, seed*0x9E3779B97F4A7C15+1)),
		Mods: DefaultModifiers(),
		Market: MarketState{
			Price: [entity.GoodCount]float64{1, 0.5, 3},
		},
		Log:    event.NewLog(50_000),
		techs:  map[Tech]bool{},
		nextID: 1,
	}
	w.GenerateTerrain(width, height)
	return w
}

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

// Other picks a random agent that is not a. Returns nil if a is alone.
func (w *World) Other(a *entity.Agent) *entity.Agent {
	if len(w.Agents) < 2 {
		return nil
	}
	for {
		o := w.Agents[w.RNG.IntN(len(w.Agents))]
		if o != a {
			return o
		}
	}
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

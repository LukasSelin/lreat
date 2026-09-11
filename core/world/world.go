// Package world holds the complete simulation state.
//
// Exactly one goroutine may drive a World. The sim package enforces that;
// everything else assumes it. Within a tick the passes that only read are
// spread over several goroutines - agents deciding, the day's walk over the
// ground, and what the day's ruin works out - while everything that changes
// the world runs one at a time. See parallel.go for the rule they keep.
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
	ShelterDecay    float64 // how fast a roof wears, as a share of the usual
	FishYield       float64
	HuntYield       float64
	Regrowth        float64 // how fast forest, wild food, and fish come back
	Keeping         float64 // how much of the market's food spoils, as a share of the usual
	// Warmth is how much of the cold a body actually feels, as a share of
	// the usual. It is the one modifier that does nothing at all most of
	// the year: a cloak is worth nothing in June and worth a life in
	// February, so what it buys a settlement depends on where the
	// settlement is and what winters it gets rather than on a flat rate.
	Warmth float64
	// Healing is how much faster a body climbs back toward the condition
	// its circumstances would give it. It works one way only - see
	// system.Decay - because knowing what to do for a fever helps somebody
	// recover and does not make anybody fall ill quicker.
	Healing float64
}

// Granaries is how many granaries are standing. It is what a granary does
// for the settlement rather than what building one did: a store that has
// fallen in keeps nothing, and the market's food has to notice that. Kept
// as a count rather than folded into Keeping when one goes up, because a
// one-way multiplier cannot be undone when one comes down.
func (w *World) Granaries() int {
	return w.Grid.Granaries()
}

// DefaultModifiers is the pre-technology baseline.
func DefaultModifiers() Modifiers {
	return Modifiers{
		FarmYield:       1,
		BuildEfficiency: 1,
		StudyRate:       1,
		CraftQuality:    1,
		ShelterDecay:    1,
		FishYield:       1,
		HuntYield:       1,
		Regrowth:        1,
		Warmth:          1,
		Healing:         1,
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

	Grid *Grid
	// MarketPos is the principal square: where the settlement was founded,
	// where its roads are measured from, and what "in the settlement" is
	// reckoned around. A town may hold several squares - see Markets - but
	// one of them is still the middle of it.
	MarketPos entity.Pos
	// markets is every square trade may be done at, the principal one
	// among them, in the order they were founded. Kept rather than counted
	// because the nearest square is asked for by every agent every tick.
	markets []entity.Pos

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
	// Deaths counts everyone who has died here. Vitals is the same
	// turnover written out: by cause, by tick, and with the reason nobody
	// else was born beside it. Chronicle is the last of the births,
	// deaths, and discoveries, kept so a viewer that has fallen behind can
	// still say what happened.
	Deaths    int
	Vitals    Vitals
	Chronicle []Note
	// Choices and Entropy record this tick's fit-based decisions: how many
	// were made and how open they were in total, for observation.
	Choices int
	Entropy float64

	// watched is the agent whose deliberations are being kept, and thoughts
	// is what it has been weighing. Both are for observation only; see
	// watch.go.
	watched  entity.ID
	thoughts []Deliberation

	techs     map[Tech]Known
	pressed   map[Tech]float64
	nextID    entity.ID
	nextReqID entity.RequestID

	// index is where everybody is filed; see index.go.
	index

	// Growing is the growing weather the world has had since it was made,
	// in growing days, and swept is where the sweep of sleeping chunks has
	// got to; see active.go.
	Growing []float64 // by chunk, because the weather goes by latitude and height
	swept   int
	rates   []float64

	// Config is the terms this world was made on.
	Config Config

	// Stuck counts the plans made whose way was looked for and not found,
	// and Awake says why the ground is awake, both for a runner's timing
	// line: what a tick is spent on is mostly what these say.
	Stuck int
	Awake AwakeCount

	// seed is what the world was made from, kept for the streams of chance
	// the islands draw on; see island.go.
	seed uint64
	// isles is the working memory of cutting the population into islands
	// and acting on them, and Isles how the last day's acting was cut.
	isles isles
	Isles IsleCount

	// routers is the working memory deciding routes on, one per goroutine.
	routers []*Router
	// ways is this tick's reading of the worn ground, kept between ticks so
	// that taking it again writes into the same buffers.
	ways *Ways
	// ruin is the working memory of what nobody is left to keep, kept
	// between ticks so that the day's ruin allocates nothing. See Ruin.
	ruin Ruin
}

// Config is the terms a world is made on: how big the ground is and what
// shape. Nothing in it changes once the world is made.
type Config struct {
	Width, Height int
	// Wrap joins the east edge to the west: the map is a globe drawn as a
	// cylinder rather than a valley with edges. See Grid.
	Wrap bool
	// SeaShare is how much of the ground lies under the sea. A valley has
	// none: its water leaves at the edges. A globe has no edges but the
	// poles, and without a sea every river on it runs to a pole and every
	// laden walker is cut off by one.
	SeaShare float64
	// Settlements is how many parties are founded. One, until the world can
	// hold more than one settlement.
	Settlements int
	// LogCapacity is how many events the log keeps; zero is the usual.
	LogCapacity int
	// Epochs is how many ages of the earth to run before the world is
	// handed over: 0 draws the land, and anything else makes it out of its
	// own history. See history.go.
	Epochs int
	// Deer is how many deer are put down in the woods around the settlement
	// once it is founded; see Populate. None, unless somebody asks: the
	// runs a settlement is measured on have no creatures in them.
	Deer int
}

// DefaultConfig is the valley every settlement was founded in before there
// was anywhere else: the default size, with edges.
func DefaultConfig() Config {
	return Config{Width: DefaultWidth, Height: DefaultHeight, Settlements: 1}
}

// Globe is a world with room for several settlements: a cylinder sixteen
// chunks round and eight down, a third of it sea, cold at the poles and
// warm at the middle. Nothing measured on the default map is measured on
// this; it has a baseline of its own.
//
// It is made out of its own history rather than drawn, which the default
// valley is not. A valley is two kilometres of country and you see one corner
// of one plate boundary on it, so drawing the ground is as true as running
// for it and costs a fiftieth as much. A globe is the whole diagram at once,
// and a drawn one shows it: the eye picks out the lattice the mountains were
// masked in with, however carefully that mask is shaped. Run instead, the
// coasts are where continents ended up, the ranges are where they met, and
// the rock in them is what the meeting made. It costs about fourteen seconds
// a world against a second, which is a price worth paying once at the start
// of a game and not worth paying for a valley. See history.go.
func Globe() Config {
	return Config{Width: 1024, Height: 512, Wrap: true, SeaShare: 0.3, Settlements: 4, LogCapacity: 200_000, Epochs: 16}
}

// Ancient is the default valley made out of its own history rather than
// drawn: the same size, the same sea, sixteen ages of the earth before
// anybody arrives. Nothing measured on the valley is measured on this - it is
// a different map of the same kind - and it is a preset so that a made world
// can be run and looked at without being the only kind there is. See
// history.go, and normalise, which is the join that makes it the same kind.
func Ancient() Config {
	cfg := DefaultConfig()
	cfg.Epochs = 16
	return cfg
}

// New creates a world with default-sized terrain, seeded for determinism.
func New(seed uint64) *World {
	return NewWith(seed, DefaultConfig())
}

// NewSized creates a world with terrain of the given size.
func NewSized(seed uint64, width, height int) *World {
	return NewWith(seed, Config{Width: width, Height: height, Settlements: 1})
}

// NewWith creates a world on the given terms. A globe is a whole number of
// chunks round: the nine chunks around a place hold everything within a
// chunk of it only if no chunk is narrower than the rest.
func NewWith(seed uint64, cfg Config) *World {
	if cfg.Wrap && cfg.Width%ChunkSide != 0 {
		panic("world: a globe must be a whole number of chunks round")
	}
	w := &World{
		seed:    seed,
		RNG:     rand.New(rand.NewPCG(seed, seed*0x9E3779B97F4A7C15+1)),
		Mods:    DefaultModifiers(),
		Rules:   DefaultRules(),
		Climate: NewClimateOn(cfg),
		Config:  cfg,
		Market: MarketState{
			Price: [entity.GoodCount]float64{1, 0.5, 3, 1.5, 2},
		},
		Log:     event.NewLog(max(50_000, cfg.LogCapacity)),
		techs:   map[Tech]Known{},
		pressed: map[Tech]float64{},
		nextID:  1,
	}
	w.Generate(cfg)
	w.Growing = make([]float64, len(w.Grid.Chunks))
	w.Room()
	return w
}

// Markets is every square in the settlement, oldest first.
func (w *World) Markets() []entity.Pos { return w.markets }

// FoundMarket records a square. A settlement founds its first at its
// founding and raises the rest when it has spread too far to walk to the
// ones it has.
func (w *World) FoundMarket(p entity.Pos) {
	for _, q := range w.markets {
		if q == p {
			return
		}
	}
	w.markets = append(w.markets, p)
}

// CloseMarket forgets a square, for one that has been moved or built over.
func (w *World) CloseMarket(p entity.Pos) {
	for i, q := range w.markets {
		if q == p {
			w.markets = append(w.markets[:i], w.markets[i+1:]...)
			return
		}
	}
}

// NearestMarket is the square p would trade at: the closest of them, and
// the older where two are equally close, so that the answer does not
// depend on which was raised last.
//
// The list is a cache and the ground is the truth, so a square that has
// been moved or built over is passed over here rather than relied upon;
// and a settlement whose cache has nothing left in it still has the square
// it was founded on. What this must never do is return a tile with no
// market on it, because every errand in the catalog is walked to it.
func (w *World) NearestMarket(p entity.Pos) (entity.Pos, bool) {
	best, bestD, found := entity.Pos{}, 0, false
	for _, q := range w.markets {
		if !w.isMarket(q) {
			continue
		}
		if d := w.Grid.Dist(p, q); !found || d < bestD {
			best, bestD, found = q, d, true
		}
	}
	if !found && w.isMarket(w.MarketPos) {
		return w.MarketPos, true
	}
	return best, found
}

func (w *World) isMarket(p entity.Pos) bool {
	return w.Grid != nil && w.Grid.In(p) && w.Grid.At(p).Structure == Market
}

// Room keeps the reach floor as long as there are slots.
func (w *World) Room() { w.ReachFloor = habit.Grow(w.ReachFloor, habit.Slots()) }

// Spawn adds an agent on open ground near the market.
func (w *World) Spawn(name string, p need.Weights) *entity.Agent {
	pos := w.MarketPos
	for try := 0; try < 20; try++ {
		c := entity.Pos{X: w.MarketPos.X + w.RNG.IntN(9) - 4, Y: w.MarketPos.Y + w.RNG.IntN(9) - 4}
		if w.Grid.In(c) && !w.Grid.At(c).Wet() {
			pos = w.Grid.Norm(c)
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
		Kind:        entity.Human,
		Luck:        rand.New(rand.NewPCG(w.RNG.Uint64(), w.RNG.Uint64())),
		Born:        w.Tick - entity.Maturity - w.RNG.IntN((entity.Prime-entity.Maturity)/3),
		Pos:         pos,
		Needs:       need.Levels{0.7, 0.3, 0.5, 0.3, 0.2},
		Personality: p,
		Norms:       belief.RandomNorms(w.RNG),
		Temperament: entity.RandomTemperament(w.RNG),
		Body:        w.RandomBody(),
		Mind:        w.RandomMind(),
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
	w.enroll(a)
	a.Room()
	return a
}

// SpawnKind adds a creature of some other kind at a position: grown, of an
// assorted age as a founder is, with the wants its species is born with and
// a body and mind drawn around its species' own. It holds no values, no
// temperament, no belief in any craft and nothing in its arms, because a
// creature that does not settle has none of those to hold. It is never
// called for a person, and a world with no such creatures in it never
// calls it, so a settlement's chance is drawn exactly as it always was.
func (w *World) SpawnKind(sp *entity.Species, name string, p need.Weights, pos entity.Pos) *entity.Agent {
	life := sp.Life
	a := &entity.Agent{
		ID:          w.nextID,
		Name:        name,
		Kind:        sp,
		Luck:        rand.New(rand.NewPCG(w.RNG.Uint64(), w.RNG.Uint64())),
		Born:        w.Tick - life.Maturity - w.RNG.IntN(max(1, (life.Prime-life.Maturity)/3)),
		Pos:         pos,
		Needs:       sp.Needs,
		Personality: p,
		Body:        w.RandomBodyOf(sp),
		Mind:        w.RandomMindOf(sp),
		Health:      0.9,
	}
	w.nextID++
	w.Agents = append(w.Agents, a)
	w.enroll(a)
	a.Room()
	return a
}

// People is how many of those here settle: the population as the settlement
// counts it, which is everyone but the creatures. It is counted when asked
// rather than kept, because everything that asks is already walking the
// agents or is asked a few times a day.
func (w *World) People() int {
	n := 0
	for _, a := range w.Agents {
		if a.Species().Settles {
			n++
		}
	}
	return n
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

// TraitSpread is how far a drawn measure strays from the ordinary one, and
// TraitDrift how far a child's strays from its parent's. Both are narrow on
// purpose: people differ, but a settlement's fortunes should turn on what
// they want and believe, not on who was born strong or quick. A generation
// is where the drift adds up, which is why the second number is the one to
// reach for when a population ought to spread out over time.
const (
	TraitSpread = 0.12
	TraitDrift  = 0.08
)

// RandomTrait draws one measure around the ordinary one, and InheritTrait a
// child's from a parent's. Every measure of a body or a mind is drawn this
// way, so they are all the same shape of thing and none of them is anybody's
// special case.
func (w *World) RandomTrait() float64 { return clampTrait(1 + w.RNG.NormFloat64()*TraitSpread) }

func (w *World) InheritTrait(v float64) float64 {
	if v <= 0 {
		return w.RandomTrait()
	}
	return clampTrait(v + w.RNG.NormFloat64()*TraitDrift)
}

// RandomBody and RandomMind draw a whole creature. The order the measures
// are drawn in is fixed and is part of what a seed means, so a measure added
// here goes on the end.
func (w *World) RandomBody() entity.Body {
	return entity.Body{
		Vitality:   w.RandomTrait(),
		Metabolism: w.RandomTrait(),
		Hardiness:  w.RandomTrait(),
	}
}

func (w *World) RandomMind() entity.Mind {
	return entity.Mind{
		Plasticity: w.RandomTrait(),
		Resolve:    w.RandomTrait(),
		Horizon:    w.RandomTrait(),
	}
}

// RandomBodyOf and RandomMindOf draw a creature of some kind: the same
// draw as a person's, scaled by what an ordinary one of that kind is. A
// person's kind is the ordinary measure, so for a person they are
// RandomBody and RandomMind exactly.
func (w *World) RandomBodyOf(sp *entity.Species) entity.Body {
	b := w.RandomBody()
	b.Vitality *= sp.Body.Frame()
	b.Metabolism *= sp.Body.Burn()
	b.Hardiness *= sp.Body.Hardy()
	return b
}

func (w *World) RandomMindOf(sp *entity.Species) entity.Mind {
	m := w.RandomMind()
	m.Plasticity *= sp.Mind.Learns()
	m.Resolve *= sp.Mind.Decides()
	m.Horizon *= sp.Mind.Reaches()
	return m
}

// InheritBody and InheritMind are a child's, drifted from its parent's.
func (w *World) InheritBody(b entity.Body) entity.Body {
	return entity.Body{
		Vitality:   w.InheritTrait(b.Vitality),
		Metabolism: w.InheritTrait(b.Metabolism),
		Hardiness:  w.InheritTrait(b.Hardiness),
	}
}

func (w *World) InheritMind(m entity.Mind) entity.Mind {
	return entity.Mind{
		Plasticity: w.InheritTrait(m.Plasticity),
		Resolve:    w.InheritTrait(m.Resolve),
		Horizon:    w.InheritTrait(m.Horizon),
	}
}

func clampTrait(v float64) float64 {
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

// Routers returns n routers over the world's map, made once and kept between
// ticks so that deciding allocates nothing. Each is for one goroutine.
func (w *World) Routers(n int) []*Router {
	for len(w.routers) < n {
		w.routers = append(w.routers, w.Grid.Router())
	}
	return w.routers[:n]
}

// Ways is this tick's reading of the ground people have walked, taken if it
// has not been taken yet. Deciding only reads the world, so one reading
// serves every agent deciding on the tick; whoever decides first takes it.
// Agents deciding side by side must not, so it is taken for them before they
// start - see action.Ready.
func (w *World) Ways() *Ways {
	if w.ways != nil && w.ways.stamp == w.Tick+1 && sameGround(w.ways.g, w.Grid) {
		return w.ways
	}
	w.ways = w.Grid.readWays(w.ways, w.Tick)
	return w.ways
}

// Emit appends an event at the current tick.
func (w *World) Emit(kind event.Kind, actor, target entity.ID, format string, args ...any) {
	e := event.Event{
		Tick:   w.Tick,
		Kind:   kind,
		Actor:  actor,
		Target: target,
		Text:   fmt.Sprintf(format, args...),
	}
	w.Log.Append(e)
	w.note(e)
}

// EmitAt records an event that came out of a particular act on a particular
// tile, so that a reader can count what a settlement did and where without
// reading the sentence it was told in. act is an ontology key.
//
// It takes the sentence made rather than a format and its parts. This is
// the one event a settlement raises constantly - one per errand finished,
// by everybody, all day - and formatting it cost a slice for the parts, a
// box for each part, and a walk over the format, to make a sentence that
// is dropped unread whenever nobody is watching. Both callers can say what
// they mean by joining two strings.
func (w *World) EmitAt(kind event.Kind, actor, target entity.ID, act string, where entity.Pos, text string) {
	w.Log.Append(event.Event{
		Tick:   w.Tick,
		Kind:   kind,
		Actor:  actor,
		Target: target,
		Act:    act,
		Where:  where,
		Placed: true,
		Text:   text,
	})
}

// Known is what a settlement has done with a technology: the tick it worked
// the thing out, and the tick it first had somebody who was a master of the
// craft the thing lives in. The two are a long way apart and the distance
// between them is the interesting part - knowing how a field is rotated is
// not the same as having a farmer, and a settlement can hold a technology
// for a generation before anybody is really any good at it.
//
// Mastered is zero until it happens, and a technology whose craft nobody
// names - the tavern, the fish trap - is never mastered because there is no
// craft to be master of. Neither ever comes undone: a master who dies does
// not take the date with them, because the question the date answers is
// when this settlement first got there.
type Known struct {
	Found    int
	Mastered int
}

// Has reports whether a technology has been discovered.
func (w *World) Has(t Tech) bool { _, ok := w.techs[t]; return ok }

// Unlock marks a technology as discovered on this tick. Discovering
// something twice is not a thing that happens, and if it did the first time
// would be the one worth keeping.
func (w *World) Unlock(t Tech) {
	if _, ok := w.techs[t]; !ok {
		w.techs[t] = Known{Found: w.Tick}
	}
}

// Master marks a technology as one this settlement now has a master of, the
// first time it is true.
func (w *World) Master(t Tech) {
	k, ok := w.techs[t]
	if !ok || k.Mastered != 0 {
		return
	}
	k.Mastered = w.Tick
	w.techs[t] = k
}

// Known returns what the settlement has done with a technology.
func (w *World) Known(t Tech) Known { return w.techs[t] }

// Press adds a day of pressure toward a technology and returns what has
// accumulated. Pressure is measured in the only unit that makes sense for
// it - how hard the moment pressed, summed over the days it pressed - so a
// settlement in real trouble arrives in a few years and one mildly
// inconvenienced takes decades. A settlement under no pressure at all
// arrives never, which is not a rule written anywhere: it is what adding
// zero repeatedly comes to.
func (w *World) Press(t Tech, by float64) float64 {
	if by > 0 {
		w.pressed[t] += by
	}
	return w.pressed[t]
}

// Pressed is what has accumulated toward a technology.
func (w *World) Pressed(t Tech) float64 { return w.pressed[t] }

// Techs returns discovered technologies in a stable order.
func (w *World) Techs() []Tech {
	out := make([]Tech, 0, len(w.techs))
	for t := range w.techs {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Preset is a configuration by name, for a runner asked for one: "valley"
// is the default map and "globe" is Globe. ok is false for a name nobody
// has given a world.
func Preset(name string) (cfg Config, ok bool) {
	switch name {
	case "", "valley":
		return DefaultConfig(), true
	case "globe":
		return Globe(), true
	case "ancient":
		return Ancient(), true
	}
	return Config{}, false
}

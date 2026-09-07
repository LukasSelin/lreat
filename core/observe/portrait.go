package observe

import (
	"slices"
	"sort"

	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/world"
)

// A Portrait is one agent seen whole: who it is, what it has become good at,
// what it is doing about the moment it is in, and what it weighed before
// settling on that. The aggregate snapshot answers what a settlement is
// doing; it cannot answer why any particular figure on the map walked where
// it walked. This is the answer to that question, and it is for tuning and
// debugging: it shows what no agent, the player included, could know about
// somebody else.

// Errand is an agent's plan as an onlooker sees it: what it committed to,
// where, and how much of the walking and the working is left.
type Errand struct {
	Action string
	Target entity.Pos
	// Walking is true while the agent is still on its way; Steps is how many
	// tiles of the route remain. Work only begins once Steps reaches zero.
	Walking   bool
	Steps     int
	Remaining int // ticks of work left
	Total     int // ticks of work the action takes
	Started   int // the tick the plan was made
}

// Tie is one bond as its holder understands it. Everything in it is belief:
// see entity.Bond.
type Tie struct {
	To       entity.ID
	Name     string
	Strength float64
	Regard   float64
	Expect   float64
	Met      int
}

// Portrait is everything worth knowing about one agent at one tick.
type Portrait struct {
	ID   entity.ID
	Name string
	Age  int
	Pos  entity.Pos

	Home     entity.Pos
	HasHome  bool
	Field    entity.Pos
	HasField bool

	// What it wants, and how much it is the sort of person who wants it.
	Needs       need.Levels
	Urgency     [need.Count]float64
	Personality need.Weights
	// Intensity is how pressing the whole moment is, which is what decides
	// how sharply the agent is choosing right now.
	Intensity float64

	// What it is made of and has to hand.
	Health     float64
	Vitality   float64
	Shelter    float64
	Wealth     float64
	Reputation float64
	Inventory  [entity.GoodCount]float64

	// What it can do, and what it believes it can do. The two come apart,
	// and an agent acts on the second.
	Skills   [entity.SkillCount]float64
	Efficacy [entity.SkillCount]float64
	// Calling is the skill the agent is best at: the nearest thing to a role
	// in a settlement where nobody is given one. Level is how far it has
	// come with it.
	Calling entity.Skill
	Level   float64

	// What it holds to be right, and how much it expects wrongdoing to be
	// answered here.
	Norms       belief.Norms
	Caution     float64
	Temperament entity.Temperament

	// Reach is how far into reach each action has come for this agent, by
	// catalog position. A craft nobody has taught it is out of reach however
	// well it would fit the moment.
	Reach []float64

	// Errand is what it is doing now, nil when it is between plans.
	Errand *Errand

	// Who it knows. Ties are its strongest bonds, closest first.
	Known, Friends, Feuds, Hearsay int
	Ties                           []Tie

	// Thinking is the run of decisions this agent has been watched making,
	// oldest first, empty until the world is told to watch it. See
	// world.World.Watch.
	Thinking []world.Deliberation
}

// TieCount is how many of an agent's bonds a portrait carries.
const TieCount = 4

// Look draws a portrait of one agent, nil if there is no such agent - which
// is what a watcher sees when the figure it was following has died. It must
// run on the simulation goroutine.
func Look(w *world.World, id entity.ID) *Portrait {
	a := w.Find(id)
	if a == nil {
		return nil
	}
	p := &Portrait{
		ID: a.ID, Name: a.Name, Age: a.Age(w.Tick), Pos: a.Pos,
		Home: a.Home, HasHome: a.HasHome, Field: a.Field, HasField: a.HasField,
		Needs: a.Needs, Urgency: need.Urgencies(a.Needs), Personality: a.Personality,
		Health: a.Health, Vitality: a.Vitality, Shelter: a.Shelter,
		Wealth: a.Wealth, Reputation: a.Reputation, Inventory: a.Inventory,
		Skills: a.Skills, Efficacy: a.Efficacy,
		Norms: a.Norms, Caution: a.Caution, Temperament: a.Temperament,
		Reach: slices.Clone(a.Reach),
	}
	p.Calling, p.Level = a.BestSkill()
	for t := range p.Urgency {
		p.Intensity += p.Urgency[t] * a.Personality[t]
	}
	if a.Plan != nil {
		p.Errand = &Errand{
			Action: a.Plan.Action, Target: a.Plan.Target,
			Walking: a.Pos != a.Plan.Target, Steps: len(a.Plan.Route),
			Remaining: a.Plan.Remaining, Total: a.Plan.Total, Started: a.Plan.Started,
		}
	}
	p.Known = len(a.Bonds)
	for i := range a.Bonds {
		b := &a.Bonds[i]
		if b.Met == 0 {
			p.Hearsay++
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
			p.Friends++
		}
		if b.Regard < -0.3 && back.Regard < -0.3 {
			p.Feuds++
		}
	}
	p.Ties = ties(w, a)
	if w.Watching() == a.ID {
		p.Thinking = w.Recall()
	}
	return p
}

// ties returns the bonds that say most about an agent: the ones it feels
// most strongly about either way, closest first. A grudge is as telling as
// a friendship, so both are ranked by how far from indifference they are.
func ties(w *world.World, a *entity.Agent) []Tie {
	out := make([]Tie, 0, len(a.Bonds))
	for i := range a.Bonds {
		b := &a.Bonds[i]
		t := Tie{To: b.To, Strength: b.Strength, Regard: b.Regard, Expect: b.Expect, Met: b.Met}
		if o := w.Find(b.To); o != nil {
			t.Name = o.Name
		}
		out = append(out, t)
	}
	weight := func(t Tie) float64 {
		r := t.Regard
		if r < 0 {
			r = -r
		}
		return t.Strength + r
	}
	sort.SliceStable(out, func(i, j int) bool {
		if wi, wj := weight(out[i]), weight(out[j]); wi != wj {
			return wi > wj
		}
		return out[i].To < out[j].To
	})
	if len(out) > TieCount {
		out = out[:TieCount]
	}
	return out
}

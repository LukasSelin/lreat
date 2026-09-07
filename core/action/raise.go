package action

import (
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// Raising is one verb. A granary and a tavern are the same act: carry
// materials to a plot, and leave a structure standing on it that the
// settlement has where it had none. The ontology says what it is built of
// and what it becomes; a plan says how much, where it stands, and what
// raising it is worth. An act the ontology entails for a schema with a
// plan needs no code of its own. A house and a road are raised by hand
// still: a house is kept as well as built, and a road is laid where the
// ground is worn and costs more over water.

// structures is what a built class is on the map.
var structures = map[*ontology.Class]world.Structure{
	ontology.Dwelling: world.House,
	ontology.Granary:  world.Granary,
	ontology.Tavern:   world.Tavern,
	ontology.Road:     world.Road,
}

// plan is how much a raising takes of each of its inputs, in the order the
// schema lists them, and what the raising is like.
type plan struct {
	Name string
	// Amounts is what the raising consumes of each input.
	Amounts []float64
	// Site is where it stands.
	Site func(a *entity.Agent, w *world.World) (entity.Pos, bool)
	// Room reports whether there is call for another; nil means always.
	Room func(a *entity.Agent, w *world.World) bool
	// Learn is what raising it teaches of building, and Renown the
	// standing it earns.
	Learn, Renown float64
	// Worth is what raising it is worth, by need tier. What it promises is
	// what it gives.
	Worth func(a *entity.Agent, w *world.World) need.Levels
	// Done is what the structure does for the settlement once it stands.
	Done func(w *world.World)
	// Built is what the event says.
	Built string
}

// publicPlot is a plot beside the market with its own ground around it, or
// failing that any open ground there. A public building is a door people
// come to, and a door with a wall against it is no use to anybody.
func publicPlot(radius int) func(*entity.Agent, *world.World) (entity.Pos, bool) {
	return func(_ *entity.Agent, w *world.World) (entity.Pos, bool) {
		if p, ok := w.Grid.Nearest(w.MarketPos, radius, func(p entity.Pos, _ *world.Tile) bool {
			return w.Grid.RoomToBuild(p)
		}); ok {
			return p, true
		}
		return w.Grid.Nearest(w.MarketPos, radius, func(_ entity.Pos, t *world.Tile) bool { return t.Buildable() })
	}
}

var plans = map[string]plan{
	// A granary is a public good: the market keeps what it holds. Its
	// worth to the builder is standing, and a little safety in a
	// settlement that will not run short.
	"raise/timber+stone>granary@open": {
		Name: "build granary", Amounts: []float64{granaryWood, granaryStone},
		Site: publicPlot(granaryRadius), Learn: 0.03, Renown: 0.2,
		Worth: func(_ *entity.Agent, w *world.World) need.Levels {
			return need.Levels{need.Esteem: 0.2, need.Safety: 0.05 * w.Mods.Keeping}
		},
		Done:  func(w *world.World) { w.Mods.Keeping *= granaryKeeping },
		Built: "built a granary",
	},
	// A tavern is where people meet of an evening. One is enough for a
	// settlement; a second would only split the company.
	"raise/timber>tavern@open": {
		Name: "build tavern", Amounts: []float64{tavernWood},
		Site: publicPlot(tavernRadius), Learn: 0.03, Renown: 0.3,
		Room: func(_ *entity.Agent, w *world.World) bool {
			_, taken := nearPlace(w, w.MarketPos, tavernApart, world.Tavern)
			return !taken
		},
		Worth: func(*entity.Agent, *world.World) need.Levels {
			return need.Levels{need.Esteem: 0.25, need.Belonging: 0.1}
		},
		Built: "built a tavern",
	},
}

// raising is the act the ontology entails for a schema with a plan,
// carried out from it, or nil where there is none.
func raising(in ontology.Instance) *Def {
	p, ok := plans[in.Key]
	structure, ok2 := structures[in.Schema.Output]
	if !ok || !ok2 || len(p.Amounts) != len(in.Schema.Inputs) {
		return nil
	}
	for _, c := range in.Schema.Inputs {
		if _, ok := world.GoodOf(c); !ok {
			return nil
		}
	}
	takes := in.Schema.Inputs
	tech := world.Tech(in.Tech)
	d := &Def{Name: p.Name, Ticks: in.Ticks, Target: p.Site, Supply: supply(nil, takes, false)}
	d.Available = func(a *entity.Agent, w *world.World) bool {
		for i, m := range takes {
			if mine, _ := pack(a, m); mine.Held() < p.Amounts[i] {
				return false
			}
		}
		if tech != "" && !known(a, w, tech, p.Name) {
			return false
		}
		return p.Room == nil || p.Room(a, w)
	}
	d.Expect = func(a *entity.Agent, w *world.World, _ entity.Pos) need.Levels {
		return p.Worth(a, w)
	}
	d.Apply = func(a *entity.Agent, w *world.World) {
		t := w.Grid.At(a.Pos)
		if !t.Buildable() {
			return // somebody got there first
		}
		t.Structure = structure
		for i, m := range takes {
			mine, _ := pack(a, m)
			mine.Move(-p.Amounts[i])
		}
		if p.Done != nil {
			p.Done(w)
		}
		a.Reputation += p.Renown
		a.AddSkill(entity.Building, p.Learn)
		// The standing and the company a raising promises are the
		// builder's on the day; the safety it promises is the
		// settlement's to give, and comes as the structure does its work.
		for tier, gain := range p.Worth(a, w) {
			if need.Tier(tier) != need.Safety && gain > 0 {
				a.Needs.Add(need.Tier(tier), gain)
			}
		}
		w.Emit(event.Built, a.ID, 0, "%s %s", a.Name, p.Built)
	}
	return d
}

// raise is the raising the ontology entails under key, for the acts the
// rest of the package refers to by name.
func raise(key string) *Def {
	d := raising(instances[key])
	if d == nil {
		panic("action: no plan for " + key)
	}
	return d
}

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
	ontology.Market:   world.Market,
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
	// Room reports whether there is call for another where this one would
	// stand. It is asked at the site and not at the builder, because those
	// are not the same place: a public building goes up beside a square,
	// and somebody standing well away from the last one would otherwise
	// raise a second right on top of it. nil means always.
	Room func(w *world.World, site entity.Pos) bool
	// Learn is what raising it teaches of building, and Renown the
	// standing it earns.
	Learn, Renown float64
	// Worth is what raising it is worth, by need tier. What it promises is
	// what it gives.
	Worth func(a *entity.Agent, w *world.World) need.Levels
	// Done is what the structure does for the settlement once it stands,
	// given the ground it stands on.
	Done func(w *world.World, p entity.Pos)
	// Built is what the event says.
	Built string
}

// marketWood and marketStone are what a square is laid with, and
// marketApart how far it must stand from the next one. Apart is past the
// distance at which people stop calling a square theirs - see
// system.marketDrift - so a second is raised only where the first has
// stopped being any use.
const (
	marketWood  = 3
	marketStone = 3
	marketApart = 14
)

// ownPlot is the ground the builder is standing on, or the nearest with
// room around it. A public building beside the old square is no use to the
// people who have walked away from it, so a square goes where they are.
func ownPlot(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	if w.Grid.RoomToBuild(a.Pos) {
		return a.Pos, true
	}
	return w.Grid.Nearest(a.Pos, 3, func(p entity.Pos, _ *world.Tile) bool { return w.Grid.RoomToBuild(p) })
}

// publicPlot is a plot beside the market with its own ground around it, or
// failing that any open ground there. A public building is a door people
// come to, and a door with a wall against it is no use to anybody.
func publicPlot(radius int) func(*entity.Agent, *world.World) (entity.Pos, bool) {
	return func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
		square, ok := w.NearestMarket(a.Pos)
		if !ok {
			return entity.Pos{}, false
		}
		if p, ok := w.Grid.Nearest(square, radius, func(p entity.Pos, _ *world.Tile) bool {
			return w.Grid.RoomToBuild(p)
		}); ok {
			return p, true
		}
		return w.Grid.Nearest(square, radius, func(_ entity.Pos, t *world.Tile) bool { return t.Buildable() })
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
			return need.Levels{need.Esteem: 0.2, need.Safety: 0.05 * GranaryKeeping(w)}
		},
		Built: "built a granary",
	},
	// A square is what a town raises when it has spread too far to walk to
	// the one it has. It is not sited beside a market, for obvious reasons:
	// it is sited where the builder is, so that a settlement which has
	// walked down the valley gets its square where the walking ended.
	"raise/timber+stone>market@open": {
		Name: "found market", Amounts: []float64{marketWood, marketStone},
		Site: ownPlot, Learn: 0.03, Renown: 0.3,
		Room: func(w *world.World, site entity.Pos) bool {
			return roomApart(w, site, marketApart, world.Market)
		},
		Worth: func(*entity.Agent, *world.World) need.Levels {
			return need.Levels{need.Esteem: 0.25, need.Belonging: 0.1}
		},
		Done:  func(w *world.World, p entity.Pos) { w.FoundMarket(p) },
		Built: "founded a market",
	},
	// A tavern is where people meet of an evening. A settlement holds as
	// many as it has room for at tavernApart, which for a small one is one.
	"raise/timber>tavern@open": {
		Name: "build tavern", Amounts: []float64{tavernWood},
		Site: publicPlot(tavernRadius), Learn: 0.03, Renown: 0.3,
		Room: func(w *world.World, site entity.Pos) bool {
			return roomApart(w, site, tavernApart, world.Tavern)
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
		if p.Room == nil {
			return true
		}
		site, ok := p.Site(a, w)
		return ok && p.Room(w, site)
	}
	d.Expect = func(a *entity.Agent, w *world.World, _ entity.Pos) need.Levels {
		return p.Worth(a, w)
	}
	d.Apply = func(a *entity.Agent, w *world.World) {
		t := w.Grid.At(a.Pos)
		if !t.Buildable() {
			return // somebody got there first
		}
		w.Grid.Build(a.Pos, structure)
		for i, m := range takes {
			mine, _ := pack(a, m)
			mine.Move(-p.Amounts[i])
		}
		if p.Done != nil {
			p.Done(w, a.Pos)
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

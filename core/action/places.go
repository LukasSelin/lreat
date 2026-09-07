package action

import (
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/world"
)

// Places. A social act needs two things the old catalog did not ask for:
// somebody within reach, rather than anybody alive, and somewhere to meet.
// People meet at the market, in a house, in a granary's yard, or in a
// tavern once the settlement has learned to keep one. Nobody socialises in
// the middle of a forest. Which place an agent heads for is a matter of
// temperament: the warm go where the crowd is, the cool would rather have
// company at home.

const (
	// meetRadius is how far away company may be and still be company.
	meetRadius = 12
	// placeRadius is how far from the companion a meeting place may lie.
	placeRadius = 6
	// tavernWood is what a tavern is built of, and tavernRadius how far from
	// the market it may stand; tavernApart keeps a settlement to one until
	// it has outgrown it.
	tavernWood   = 3
	tavernRadius = 6
	tavernApart  = 12
	// tavernCheer is the belonging a meeting in a tavern adds for each side.
	tavernCheer = 0.05
)

// isPlace reports whether a tile is somewhere people meet.
func isPlace(t *world.Tile) bool {
	switch t.Structure {
	case world.Market, world.House, world.Granary, world.Tavern:
		return true
	}
	return false
}

// placeAt returns p itself if it is a place, or a place beside it.
func placeAt(w *world.World, p entity.Pos) (entity.Pos, bool) {
	return w.Grid.Nearest(p, 1, func(_ entity.Pos, t *world.Tile) bool { return isPlace(t) })
}

// nearPlace finds the nearest tile of the given structure within radius of p.
func nearPlace(w *world.World, p entity.Pos, radius int, s world.Structure) (entity.Pos, bool) {
	return w.Grid.Nearest(p, radius, func(_ entity.Pos, t *world.Tile) bool { return t.Structure == s })
}

// outgoing reports whether an agent would rather meet where the crowd is.
func outgoing(a *entity.Agent) bool { return a.Temperament.Warmth >= 0.5 }

// meetingPlace is where a would go to meet o: a place within reach of where
// o is now, chosen by a's temperament. The warm prefer a tavern, then the
// market, then wherever o already is; the cool prefer their own house, then
// o's, then wherever o already is. If there is no place near o at all, the
// meeting cannot happen: o is off in the woods.
func meetingPlace(a, o *entity.Agent, w *world.World) (entity.Pos, bool) {
	tavern := func() (entity.Pos, bool) { return nearPlace(w, o.Pos, placeRadius, world.Tavern) }
	market := func() (entity.Pos, bool) {
		if entity.Dist(o.Pos, w.MarketPos) <= placeRadius {
			return w.MarketPos, true
		}
		return entity.Pos{}, false
	}
	there := func() (entity.Pos, bool) { return placeAt(w, o.Pos) }
	home := func(h *entity.Agent) func() (entity.Pos, bool) {
		return func() (entity.Pos, bool) {
			if h.HasHome && entity.Dist(o.Pos, h.Home) <= placeRadius {
				return h.Home, true
			}
			return entity.Pos{}, false
		}
	}
	var order []func() (entity.Pos, bool)
	if outgoing(a) {
		order = []func() (entity.Pos, bool){tavern, market, there, home(o), home(a)}
	} else {
		order = []func() (entity.Pos, bool){home(a), home(o), there, tavern, market}
	}
	for _, f := range order {
		if p, ok := f(); ok {
			return p, true
		}
	}
	return entity.Pos{}, false
}

// hasCompany reports whether there is anybody within reach to meet.
func hasCompany(a *entity.Agent, w *world.World) bool { return w.Neighbor(a, meetRadius) != nil }

// towardCompany heads for the place a would meet the person it would most
// like to see. They may have moved by the time we arrive; then whoever is
// nearby will do.
func towardCompany(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	o := PickCompany(a, w)
	if o == nil {
		return entity.Pos{}, false
	}
	return meetingPlace(a, o, w)
}

// Workplaces. Making things needs somewhere to make them, as meeting needs
// somewhere to meet: a bench, a hearth, a forge, a quiet corner. An agent
// with a house has all of these at home. One without has the market for a
// bench and a desk and the tavern for a hearth, and no forge at all.

// settlementRadius is how far from the market the settlement is taken to
// reach: the distance within which the market counts as one's own.
const settlementRadius = 20

// bench is where an agent crafts: at home, or at a stall at the market.
func bench(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	if a.HasHome {
		return a.Home, true
	}
	if entity.Dist(a.Pos, w.MarketPos) <= settlementRadius {
		return w.MarketPos, true
	}
	return entity.Pos{}, false
}

// hearth is where an agent cooks: at home, or at the tavern.
func hearth(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	if a.HasHome {
		return a.Home, true
	}
	return nearPlace(w, a.Pos, settlementRadius, world.Tavern)
}

// forge is where an agent smelts: at home, and nowhere else.
func forge(a *entity.Agent, _ *world.World) (entity.Pos, bool) {
	if a.HasHome {
		return a.Home, true
	}
	return entity.Pos{}, false
}

// desk is where an agent studies: at home, or at the market where the
// records are kept. Not the tavern.
func desk(a *entity.Agent, w *world.World) (entity.Pos, bool) { return bench(a, w) }

// hasPlace turns a place-finder into an availability check.
func hasPlace(find func(*entity.Agent, *world.World) (entity.Pos, bool)) func(*entity.Agent, *world.World) bool {
	return func(a *entity.Agent, w *world.World) bool {
		_, ok := find(a, w)
		return ok
	}
}

// worthGuarding reports whether there is a market within the settlement's
// reach with somebody at it to guard. A guard at an empty square in a
// settlement that has moved on is a guard of nothing.
func worthGuarding(a *entity.Agent, w *world.World) bool {
	if entity.Dist(a.Pos, w.MarketPos) > settlementRadius {
		return false
	}
	return w.AgentAt(w.MarketPos, placeRadius, a) != nil
}

// Comfort. Resting and eating happen wherever an agent is, since a body
// that must be fed cannot be made to walk home first; but within a short
// walk they prefer a roof, home or the tavern by temperament, and are the
// better for it.

// comfortRadius is how far an agent will go to rest or eat under a roof: a
// single step. Recognition reads distance as poor fit, so a meal sent a few
// tiles off fits worse exactly when it should happen, and a one-tick act
// becomes a walk for the starving. At four tiles this preference killed
// two thirds of settlements over 96 seeds; at two it cost a dozen; at one
// it costs nothing.
const comfortRadius = 1

// comfort is where an agent would rest or eat: home or the tavern if either
// is close, the warm preferring the tavern and the cool their own hearth;
// otherwise right here.
func comfort(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	home := func() (entity.Pos, bool) {
		if a.HasHome && entity.Dist(a.Pos, a.Home) <= comfortRadius {
			return a.Home, true
		}
		return entity.Pos{}, false
	}
	tavern := func() (entity.Pos, bool) { return nearPlace(w, a.Pos, comfortRadius, world.Tavern) }
	order := []func() (entity.Pos, bool){home, tavern}
	if outgoing(a) {
		order = []func() (entity.Pos, bool){tavern, home}
	}
	for _, f := range order {
		if p, ok := f(); ok {
			return p, true
		}
	}
	return a.Pos, true
}

// underRoof reports whether p is at the agent's own house or in a tavern.
func underRoof(a *entity.Agent, w *world.World, p entity.Pos) bool {
	return (a.HasHome && p == a.Home) || inTavern(w, p)
}

// inTavern reports whether p is in or beside a tavern.
func inTavern(w *world.World, p entity.Pos) bool {
	_, ok := nearPlace(w, p, 1, world.Tavern)
	return ok
}

// tavernSite is a plot beside the market. Like a house and a granary it
// wants its own ground around it: a tavern is a door people come to of an
// evening, and a door with a wall against it is no use to anybody.
func tavernSite(_ *entity.Agent, w *world.World) (entity.Pos, bool) {
	if p, ok := w.Grid.Nearest(w.MarketPos, tavernRadius, func(p entity.Pos, _ *world.Tile) bool {
		return w.Grid.RoomToBuild(p)
	}); ok {
		return p, true
	}
	return w.Grid.Nearest(w.MarketPos, tavernRadius, func(_ entity.Pos, t *world.Tile) bool { return t.Buildable() })
}

var BuildTavern = &Def{
	Name: "build tavern", Ticks: 4, Target: tavernSite,
	Available: func(a *entity.Agent, w *world.World) bool {
		if a.Inventory[entity.Wood] < tavernWood || !known(a, w, "brewing", "build tavern") {
			return false
		}
		_, taken := nearPlace(w, w.MarketPos, tavernApart, world.Tavern)
		return !taken
	},
	Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
		return need.Levels{need.Esteem: 0.25, need.Belonging: 0.1}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		t := w.Grid.At(a.Pos)
		if !t.Buildable() {
			return
		}
		t.Structure = world.Tavern
		a.Inventory[entity.Wood] -= tavernWood
		a.Reputation += 0.3
		a.AddSkill(entity.Building, 0.03)
		a.Needs.Add(need.Esteem, 0.25)
		a.Needs.Add(need.Belonging, 0.1)
		w.Emit(event.Built, a.ID, 0, "%s built a tavern", a.Name)
	},
}

const reachTavern = 0.2

func init() {
	// A tavern belongs to the lonely with standing to win and timber to
	// spare, among people they mean to keep company with.
	seed(BuildTavern, reachTavern, habit.Signature{
		habit.Lonely: 0.5, habit.Unproven: 0.6, habit.Wood: 0.8, habit.Company: 0.5, habit.Charity: 0.3, habit.Tradition: 0.3,
	})
	BuildTavern.Skilled = uses(entity.Building)
}

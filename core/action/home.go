package action

import (
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/world"
)

// A house is where somebody decided to put it on the day they could afford
// it, which is rarely where it belongs once there is a settlement around it.
// The market moves nobody, but a field is cleared on the far side of the
// water, neighbours build and the errands between them wear a way straight
// through somebody's parlour, and the good ground by the square falls empty
// when its owner dies. None of that is a reason to raise a second house; it
// is a reason to move the one there is. Moving is what lets a settlement be
// rearranged by the people living in it rather than only added to.

const (
	// movingWood is the timber a move costs. Most of a house comes apart and
	// goes up again, so what is lost is what the taking down wastes: the
	// price of a length of road rather than of a house.
	movingWood = 1
	// movingRadius is how far somebody will carry their house. A move is a
	// rearrangement of the neighbourhood, not emigration.
	movingRadius = 10
	// worthMoving is how much better the new plot has to be, measured in
	// tiles of walking saved every day. Below it the settlement would shuffle
	// itself for ever over differences nobody would notice.
	worthMoving = 3
	// inTheWay is what a thoroughfare through the house is worth in those
	// same tiles, at the wear a road is thought worth laying at. Houses can
	// be walked through, slowly, so ground worn inside one is the settlement
	// saying a way wants to run exactly there. Moving out of its line is how
	// a street gets to be straight: nobody plans the square, it is what is
	// left when the houses in the way have stepped aside.
	inTheWay = 12
)

// homeCost is how poorly a house at p serves the life its owner leads: the
// walk to their field, and the ground itself, which is worth what living on
// it is worth and is counted in the same tiles of walking.
//
// It used to open with the walk to the market. That was the last place in the
// simulation where the shape of a settlement was written down rather than
// arrived at: wherever the map generator had put the market, every house for
// the rest of time was pulled toward it, and a town could not grow anywhere
// else however good the ground was. What holds a settlement together now is
// in LandWorth - people are worth more to each other than to any building -
// and the market earns whatever pull it has by being somewhere people go.
func homeCost(a *entity.Agent, w *world.World, p entity.Pos) float64 {
	c := -LandWorth(w, p)
	if a.HasField {
		c += float64(entity.Dist(p, a.Field))
	}
	return c
}

// betterPlot is the best plot its owner knows of, if it would serve them
// better by enough to be worth the move.
//
// It used to weigh every tile in a square twenty-one across, which was both
// expensive and a kind of omniscience: a householder considering a move had
// the whole neighbourhood in front of them, surveyed, including ground they
// had never once walked. A move now runs off the same list of remembered
// places a first house is sited from, which is a shorter list and an honest
// one - you move somewhere you have been.
func betterPlot(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	if !a.HasHome {
		return entity.Pos{}, false
	}
	bestCost := homeCost(a, w, a.Home) - worthMoving
	best, found := entity.Pos{}, false
	for i := range a.Places {
		p := a.Places[i].Pos
		if entity.Dist(a.Home, p) > movingRadius || !w.Grid.RoomToBuild(p) {
			continue
		}
		if c := homeCost(a, w, p); c < bestCost {
			best, bestCost, found = p, c, true
		}
	}
	return best, found
}

// MoveHouse pulls a house down and raises it again on better ground. It
// gives the mover nothing they did not already have - the roof that comes
// with them is the roof they had - and costs a day and some timber. What it
// pays back is every walk they take afterwards, which is why it is worth
// making at all and why it is only ever worth making toward ground that is
// better by a distance a person would notice.
var MoveHouse = &Def{
	Name: "move house", Ticks: 4, Target: betterPlot,
	Available: func(a *entity.Agent, _ *world.World) bool {
		return a.HasHome && a.Inventory[entity.Wood] >= movingWood
	},
	Expect: func(a *entity.Agent, w *world.World, target entity.Pos) need.Levels {
		// What a move is worth is the walking it saves, counted small: a
		// tile a day off the errands of a settled life, against the day the
		// move itself takes.
		saved := homeCost(a, w, a.Home) - homeCost(a, w, target)
		return need.Levels{need.Physiological: 0.004 * saved, need.Safety: 0.01, need.Esteem: 0.02}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		if !a.HasHome || a.Inventory[entity.Wood] < movingWood {
			return
		}
		// Somebody may have built on the plot during the walk over, and the
		// old house may have fallen in behind them.
		if !w.Grid.RoomToBuild(a.Pos) {
			return
		}
		if old := w.Grid.At(a.Home); old.Structure == world.House && old.Owner == a.ID {
			w.Grid.Raze(a.Home)
		}
		t := w.Grid.At(a.Pos)
		t.Structure, t.Owner = world.House, a.ID
		a.Home = a.Pos
		a.Inventory[entity.Wood] -= movingWood
		a.AddSkill(entity.Building, 0.02)
		a.Needs.Add(need.Esteem, 0.02)
		w.Emit(event.Built, a.ID, 0, "%s moved house", a.Name)
	},
}

// The moment moving belongs to is the settled one, as paving does: somebody
// under a roof with timber past what the roof took. What separates it from
// paving is that it is a private rearrangement rather than a public work, so
// it says nothing of charity or company, and that the plot has to be near:
// a house is carried, not sent.
//
// It is within reach from birth, unlike paving. Nobody has to be taught that
// they could live somewhere better; what they need is a roof to move and the
// timber to spare, and the prior asks for both, which is gate enough. Started
// half out of reach it was chosen too seldom to rearrange anything.
func init() {
	seed(MoveHouse, reachEveryday, habit.Signature{
		habit.Shelter: 0.7, habit.Industry: 0.5, habit.Near: 0.8,
	})
	MoveHouse.Skilled = uses(entity.Building)
}

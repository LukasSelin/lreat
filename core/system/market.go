package system

import (
	"math"

	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/ontology"
	"lreat/core/world"
)

// basePrice is what each good costs when the market holds a normal stock.
var basePrice = [entity.GoodCount]float64{
	entity.Food:  1,
	entity.Wood:  0.5,
	entity.Tools: 3,
	entity.Stone: 1.5,
	entity.Meals: 2,
}

// What is kept where, and what it loses for being kept there. The rates
// are ontology.Transforms' - a shelf and a pack are two keepings of the
// same food, and the ontology says what each costs - read into tables
// keyed by good once at start, because the loops below run every tick over
// every shelf and every pack and must not walk the class tree to do it.
//
// sheltered is what a granary and the cold can reach: the goods whose
// transform names a site that arrests it.
var (
	shelfLoss [entity.GoodCount]float64
	packLoss  [entity.GoodCount]float64
	sheltered [entity.GoodCount]bool
)

func init() {
	for _, c := range ontology.Material.Family() {
		g, ok := world.GoodOf(c)
		if !ok {
			continue
		}
		if t, ok := ontology.Spoiling(c, ontology.Market); ok {
			shelfLoss[g] = t.Rate
			sheltered[g] = t.Unless != nil
		}
		if t, ok := ontology.Spoiling(c, ontology.Person); ok {
			packLoss[g] = t.Rate
		}
	}
}

// ColdKeeping is how much of a perishable's spoilage the bitterest cold
// stops. A cold store is the oldest one there is: what the year takes from
// the settlement in the growing it stops doing, it gives back a little of in
// the keeping. It is what makes an autumn surplus worth holding rather than
// selling, and it is the one thing in the world that gets better as the
// weather gets worse.
const ColdKeeping = 0.6

// Keeping is the share of the usual spoilage a perishable actually suffers
// here today: what the settlement's granaries stop, and what the weather
// stops on top of that.
func Keeping(w *world.World) float64 {
	return w.Mods.Keeping * (1 - ColdKeeping*w.Climate.Chill())
}

// Larder is Keeping for a pack rather than a shelf: the share of the usual
// spoilage what an agent carries actually suffers. Nothing a granary does
// reaches a pack - the settlement's stores keep the settlement's food - so
// the cold is the whole of its mercy, and in the deep of winter a pack
// keeps two and a half times as well as it does in the summer.
func Larder(w *world.World) float64 {
	return 1 - ColdKeeping*w.Climate.Chill()
}

// Spoil rots what an agent is carrying.
func Spoil(w *world.World, a *entity.Agent) {
	k := Larder(w)
	for g := range a.Inventory {
		if packLoss[g] != 0 {
			a.Inventory[g] *= 1 - packLoss[g]*k
		}
	}
}

// MarketStep moves prices against stock and lets goods spoil. Prices lag
// supply, so gluts and shortages both overshoot, which is what gives traders
// something to notice.
func MarketStep(w *world.World) {
	m := &w.Market
	for g := range m.Stock {
		loss := shelfLoss[g]
		if sheltered[g] {
			loss *= Keeping(w)
		}
		m.Stock[g] = math.Max(0, m.Stock[g]*(1-loss))
		target := basePrice[g] * math.Sqrt(10/(m.Stock[g]+2))
		target = math.Max(0.2*basePrice[g], math.Min(5*basePrice[g], target))
		m.Price[g] += (target - m.Price[g]) * 0.05
	}
}

// The market follows the town.
//
// A market is not a monument. It is the tile everybody's errands run to, and
// what makes it that tile is that everybody is near it - which was true on
// the day it was founded and need not stay true afterwards. It used to have
// to stay true, because siting named MarketPos: every house ever built was
// pulled toward the square, so the settlement could not go anywhere the
// square was not. Now that people site their houses on the ground they think
// best, a town can and does walk - toward the treeline, up out of a flood
// plain, along the river - and leave its market behind it.
//
// A market left behind is worse than a market in the wrong place. Half the
// catalog reads from MarketPos: the crafting bench and the guard's post are
// only there for people inside settlementRadius of it, and buying and
// selling are a walk to it. A settlement that moved twenty-seven tiles off
// its square stopped guarding, stopped crafting, stopped trading, never
// discovered anything, and never had a child, because safety never rose to
// what a birth asks for. So the square goes after the town.
//
// Nothing here decides where anybody lives. It reads where they already do.
const (
	// marketDrift is how far the middle of the settlement may wander from
	// its market before the market is moved to meet it. A square that
	// chased every birth and death would never sit still; this is about the
	// distance at which people would stop calling it their market.
	marketDrift = 10
	// marketEvery is how often the question is asked. Moving a market is a
	// generation's work, not a morning's.
	marketEvery = 200
)

// middle is the settlement's centre of gravity: the mean of its roofs, or of
// its people while it has no roofs yet. Houses are the better reading once
// there are any - they are where people sleep rather than where they happen
// to be standing this tick - but a founding party has none.
func middle(w *world.World) (entity.Pos, bool) {
	sx, sy, n := 0, 0, 0
	for y := 0; y < w.Grid.H; y++ {
		for x := 0; x < w.Grid.W; x++ {
			if w.Grid.At(entity.Pos{X: x, Y: y}).Structure == world.House {
				sx, sy, n = sx+x, sy+y, n+1
			}
		}
	}
	if n == 0 {
		for _, a := range w.Agents {
			sx, sy, n = sx+a.Pos.X, sy+a.Pos.Y, n+1
		}
	}
	if n == 0 {
		return entity.Pos{}, false
	}
	return entity.Pos{X: sx / n, Y: sy / n}, true
}

// MoveMarket carries the square to the middle of the settlement when the
// settlement has walked away from it.
func MoveMarket(w *world.World) {
	if w.Tick%marketEvery != 0 {
		return
	}
	mid, ok := middle(w)
	if !ok || entity.Dist(mid, w.MarketPos) <= marketDrift {
		return
	}
	// Somewhere in the middle of things with room round it to stand a crowd,
	// and failing that any open ground at all.
	site, ok := w.Grid.Nearest(mid, marketDrift, func(p entity.Pos, _ *world.Tile) bool {
		return w.Grid.RoomToBuild(p)
	})
	if !ok {
		site, ok = w.Grid.Nearest(mid, marketDrift, func(_ entity.Pos, t *world.Tile) bool {
			return t.Buildable()
		})
	}
	if !ok {
		return
	}
	if old := w.Grid.At(w.MarketPos); old.Structure == world.Market {
		old.Structure = world.None
	}
	t := w.Grid.At(site)
	t.Terrain, t.Structure = world.Grass, world.Market
	w.MarketPos = site
	w.Emit(event.Built, 0, 0, "the market moved to where the town had gone")
}

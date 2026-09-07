package action

import (
	"math"

	"lreat/core/entity"
	"lreat/core/ontology"
	"lreat/core/world"
)

// Stores. A material is always somewhere: in the ground, in a pack, in a
// purse, on the market's shelves. The ontology says which kinds of thing
// hold which materials (ontology.Holds); this says where that is at run
// time, as a number that can be read and moved. Every generic verb moves
// material between stores it names by class, so a verb never says
// "inventory" or "tile" or "wealth": it says the ground here, this
// person's pack, the market.

// A store is some of a material somewhere.
type store interface {
	// Held is how much is there.
	Held() float64
	// Move adds delta, which may be negative, and never goes below nothing.
	Move(delta float64)
}

// cell is a store that is a number in the world.
type cell struct{ at *float64 }

func (c cell) Held() float64  { return *c.at }
func (c cell) Move(d float64) { *c.at = max(0, *c.at+d) }

// bottomless is a store with no end of what it holds and no room to fill:
// the stone in an outcrop, the market's coin. Taking from it takes
// nothing away and giving to it gives nothing back.
type bottomless struct{}

func (bottomless) Held() float64 { return math.Inf(1) }
func (bottomless) Move(float64)  {}

// goods is what a material is in a pack. A material that is nothing in a
// pack cannot be carried.
var goods = map[*ontology.Class]entity.Good{
	ontology.Provision: entity.Food,
	ontology.Timber:    entity.Wood,
	ontology.Stone:     entity.Stone,
	ontology.Tool:      entity.Tools,
	ontology.Meal:      entity.Meals,
}

// good is what m is in a pack, walking up the tree: berries are food.
func good(m *ontology.Class) (entity.Good, bool) {
	for c := m; c != nil; c = c.Parent {
		if g, ok := goods[c]; ok {
			return g, true
		}
	}
	return 0, false
}

// stocks is what the ground holds of a material, by tile. A material the
// ground holds but has no stock of (stone in an outcrop) is bottomless.
var stocks = map[*ontology.Class]func(*world.Tile) *float64{
	ontology.Berries: func(t *world.Tile) *float64 { return &t.Wild },
	ontology.Game:    func(t *world.Tile) *float64 { return &t.Wild },
	ontology.Timber:  func(t *world.Tile) *float64 { return &t.Wood },
	ontology.Fish:    func(t *world.Tile) *float64 { return &t.Fish },
}

// pack is what a person holds of m: their inventory, or their purse for
// coin. ok is false for a thing nobody can carry.
func pack(a *entity.Agent, m *ontology.Class) (store, bool) {
	if m == ontology.Coin {
		return cell{&a.Wealth}, true
	}
	g, ok := good(m)
	if !ok {
		return nil, false
	}
	return cell{&a.Inventory[g]}, true
}

// shelf is what the market holds of m. The market's coin is bottomless:
// it pays whatever is brought and takes whatever is paid.
func shelf(w *world.World, m *ontology.Class) (store, bool) {
	if m == ontology.Coin {
		return bottomless{}, true
	}
	g, ok := good(m)
	if !ok {
		return nil, false
	}
	return cell{&w.Market.Stock[g]}, true
}

// stock is soil resolved once for a material rather than once per tile: what
// the ground holds of m as a number on a tile, or nil where the ground has
// no stock of it and is bottomless. It is for the searches that ask the same
// question of the whole map.
func stock(m *ontology.Class) func(*world.Tile) *float64 { return stocks[m] }

// soil is what a tile holds of m, given that the ground it stands for
// holds it at all: its stock, or a bottomless store where it has none.
func soil(t *world.Tile, m *ontology.Class) store {
	if s, ok := stocks[m]; ok {
		return cell{s(t)}
	}
	return bottomless{}
}

// price is what the market asks for a material.
func price(w *world.World, m *ontology.Class) float64 {
	g, _ := good(m)
	return w.Market.Price[g]
}

// Knees are the amounts of each material at which its store reads as
// full. They are what an agent reads as plenty, and plenty is the whole of
// what stops it going out to get more. Without a learner to notice that a
// full larder went on feeding a family for a fortnight, these are the only
// place that judgement lives.
//
// The food knee is a fortnight's eating, not a day's. Read against four
// units - the point past which one more unit is worth nothing to eat today
// - a settlement lived hand to mouth: two units in the basket read as half
// plenty, nobody went out, and a fertile adult was fed well enough to think
// of a child about a third of the time. What a larder is for is the week
// that has not happened yet, and read at sixteen the same person keeps
// working.
//
// The wood knee has to be the frame's price and not an errand's. Read
// against an armful, a person holding two lengths already feels flush, and
// feeling flush is what stops them going back to the woods; read against
// the frame, wood stays something to be short of until there is a house's
// worth of it, which is what makes gathering toward a house a thing an
// agent will keep at for days. Twenty coins is rich.
var knees = map[*ontology.Class]float64{
	ontology.Provision: 16,
	ontology.Timber:    raisingTimber,
	ontology.Coin:      20,
	ontology.Tool:      1,
	ontology.Meal:      3,
	ontology.Stone:     4,
}

// knee is the amount at which a store of m reads as full, walking up the
// tree.
func knee(m *ontology.Class) float64 {
	for c := m; c != nil; c = c.Parent {
		if k, ok := knees[c]; ok {
			return k
		}
	}
	return 1
}

// fullness is how well supplied a is in m, in [-1, 1].
func fullness(a *entity.Agent, m *ontology.Class) float64 {
	s, ok := pack(a, m)
	if !ok {
		return 0
	}
	return bipolar(s.Held() / knee(m))
}

// supply is how an agent's stores bear on an act that brings gets and
// spends gives: how short of the former, how well supplied in the latter.
// An act that needs all of what it spends reads as supplied only as far
// as its scarcest input; one that spends whichever it has to spare reads
// as supplied as far as its fullest. Where an act brings nothing in
// particular, or spends nothing, that side reads the stores in general,
// so that every act's moment has both coordinates live and a habit can
// learn that felling is for the well fed as much as that it is for the
// short of wood.
func supply(gets, gives []*ontology.Class, any bool) func(a *entity.Agent) (lack, stock float64) {
	if len(gets) == 0 && len(gives) == 0 {
		return nil
	}
	return func(a *entity.Agent) (lack, stock float64) {
		lack, stock = -plenty(a), plenty(a)
		for i, m := range gets {
			if l := -fullness(a, m); i == 0 || l > lack {
				lack = l
			}
		}
		for i, m := range gives {
			f := fullness(a, m)
			if i == 0 || (any && f > stock) || (!any && f < stock) {
				stock = f
			}
		}
		return lack, stock
	}
}

// everyday is what everybody keeps some of and runs short of: the stores
// a moment is well off or badly off in, taken together.
var everyday = []*ontology.Class{ontology.Provision, ontology.Timber, ontology.Coin}

// plenty is how well off an agent is in what it keeps, taken all
// together: the mean fullness of its everyday stores. It is what an act
// that moves nothing in particular reads. It is read over the everyday stores
// and not over everything with a knee, because tools, meals, and stone
// are empty for almost everyone almost always, and a mean that counts
// them is the same for everyone - which is a tax on whichever acts learn
// it and tells nobody anything.
func plenty(a *entity.Agent) float64 {
	var sum float64
	for _, m := range everyday {
		sum += fullness(a, m)
	}
	return sum / float64(len(everyday))
}

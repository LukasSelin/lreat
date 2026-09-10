package action

import (
	"fmt"

	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// Moving is one verb, and taking, exchanging, and handing over are three
// names for it. Every one of them is: go to a holder of a material - the
// ground, the market, a person who stands in some relation to you - and
// move some of it between their store and yours, one way or the other.
// What the trees say is what moves, from what kind of holder, and which
// way. What a move says is how much, when it is worth the walk, what the
// far side loses if that differs from what you gain, what is paid, and
// what passes between the two parties besides the thing. One interpreter
// carries all nine acts, and any act the trees entail over a holder it
// knows how to reach.

// ground is what a site is like to take from: how to know a tile of it,
// and which tile the taking draws on when standing there. Fishing stands
// on the bank and draws on the water beside it.
type ground struct {
	Here  func(w *world.World, p entity.Pos, t *world.Tile) bool
	Drawn func(w *world.World, p entity.Pos) *world.Tile
	// Kinds is the ground Here can be true of, or true near: no tile of
	// any other kind can satisfy it. Margin is how far off that ground a
	// tile may stand and still satisfy it - nought where Here asks about
	// the tile itself, and one for a bank, which is dry ground beside
	// water rather than water.
	//
	// Between them they let a search that is going to find nothing say so
	// from the chunk counts instead of walking every ring to prove it.
	// They are derived from the trees and not written out: what affords
	// stone is the ontology's to say, and this asks it.
	Kinds  world.KindSet
	Margin int
}

func at(w *world.World, p entity.Pos) *world.Tile { return w.Grid.At(p) }

var grounds = map[*ontology.Class]ground{
	ontology.Wood: {
		Here:  func(_ *world.World, p entity.Pos, t *world.Tile) bool { return isForest(p, t) },
		Drawn: at, Kinds: world.KindsOf(ontology.Wood),
	},
	ontology.Outcrop: {
		Here:  func(_ *world.World, p entity.Pos, t *world.Tile) bool { return isRock(p, t) },
		Drawn: at, Kinds: world.KindsOffering(ontology.Stone),
	},
	ontology.Water: {
		Here:  func(w *world.World, p entity.Pos, t *world.Tile) bool { return bank(w)(p, t) },
		Drawn: bestWater, Kinds: world.KindsOffering(ontology.Fish), Margin: 1,
	},
}

// whom is who stands in a role to the actor, given the material the act
// is about: a holder has some to spare, the needy have none and are
// hungry.
var whom = map[*ontology.Role]func(m *ontology.Class) func(o *entity.Agent) bool{
	&ontology.Holder: func(m *ontology.Class) func(o *entity.Agent) bool {
		return func(o *entity.Agent) bool { theirs, _ := pack(o, m); return theirs.Held() >= 1 }
	},
	&ontology.Needy: func(m *ontology.Class) func(o *entity.Agent) bool {
		return func(o *entity.Agent) bool {
			theirs, _ := pack(o, m)
			return o.Needs[need.Physiological] < 0.4 && theirs.Held() < 1
		}
	},
}

// quantity is how much of a material one move brings to the receiving
// store, given both stores and the luck of the day.
type quantity func(a *entity.Agent, w *world.World, from, to store, luck float64) float64

// fixed is a set amount.
func fixed(x float64) quantity {
	return func(*entity.Agent, *world.World, store, store, float64) float64 { return x }
}

// surplus is everything the giving store holds above what is kept.
func surplus(keep float64) quantity {
	return func(_ *entity.Agent, _ *world.World, from, _ store, _ float64) float64 { return from.Held() - keep }
}

// yields is what the land gives from what it holds, by luck.
func yields(f func(a *entity.Agent, w *world.World, stock float64) float64) quantity {
	return func(a *entity.Agent, w *world.World, from, _ store, luck float64) float64 {
		return f(a, w, from.Held()) * luck
	}
}

// part is the terms for one material of a move.
type part struct {
	// Quantity is what arrives.
	Quantity quantity
	// Drain is what the far store loses instead, where the land gives
	// other than what it loses; zero means it loses what arrives.
	Drain float64
	// Luck spreads the quantity by (lo + span*chance). A zero span is a
	// sure thing.
	Luck struct{ Lo, Span float64 }
	// Want is what the actor must hold less than to receive; Spare what
	// they must hold at least to give. Zero is no condition.
	Want, Spare float64
	// Least is what the far store must hold for the move to be worth
	// making, and Spent the level at or below which there is nothing
	// there; negative means there always is.
	Least, Spent float64
}

// move is what a move is like beyond what the trees say.
type move struct {
	Name string
	// Each is the terms per material. A move over a class covers every
	// material listed here that is of that class.
	Each map[*ontology.Class]part
	// Tool is how much of a tool the move wears; zero means it needs none.
	Tool float64
	// Skill is what the move exercises, by Learn.
	Skill   entity.Skill
	Skilled bool
	Learn   float64
	// Regard is how the other party comes to think of the actor for it,
	// Bonds what it adds to the tie each way, and Rift what it takes from
	// the other's tie to the actor.
	Regard float64
	Bonds  struct{ Theirs, Mine float64 }
	Rift   float64
	// Worth is what a move bringing this much is worth, by need tier, and
	// Gives what of that the actor feels on the day: standing and company;
	// the body's tiers come with what was got.
	Worth func(a *entity.Agent, w *world.World, quantity float64) need.Levels
	Gives need.Levels
	// Gone is what becomes of the tile once the move has drawn on it.
	Gone func(w *world.World, p entity.Pos, t *world.Tile)
	// Event is what the act is called in the record, and Report how it
	// reads; nil for a move nobody records.
	Event  event.Kind
	Report func(a, o *entity.Agent, quantity, paid float64) string
}

// fed is what a move of food is worth: a meal, less the more one has.
func fed(a *entity.Agent, _ *world.World, q float64) need.Levels {
	return need.Levels{need.Physiological: foodValue(a) * q}
}

func levels(l need.Levels) func(*entity.Agent, *world.World, float64) need.Levels {
	return func(*entity.Agent, *world.World, float64) need.Levels { return l }
}

var moves = map[string]move{
	// The nearest forest, thin as it may be. Going further for a fuller
	// patch was tried: the walk each way, on the errand the whole economy
	// runs on, cost more than the fuller patch gave, and a forager who
	// stays put and finds less is what turns a settlement toward the river
	// and the field. A picked forest still gives something, on purpose.
	"take/berries@wood": {
		Name: "forage", Worth: fed,
		Each: map[*ontology.Class]part{ontology.Berries: {
			Quantity: yields(func(_ *entity.Agent, _ *world.World, wild float64) float64 { return forageYield(wild) }),
			Drain:    forageTake, Spent: -1, Luck: struct{ Lo, Span float64 }{0.6, 0.8},
		}},
	},
	"take/game@wood": {
		Name: "hunt", Tool: toolWear, Worth: fed,
		Each: map[*ontology.Class]part{ontology.Game: {
			Quantity: yields(func(_ *entity.Agent, w *world.World, wild float64) float64 { return huntYield(w, wild) }),
			Drain:    huntTake, Least: 0.3, Spent: 0.1, Luck: struct{ Lo, Span float64 }{0.5, 1.0},
		}},
	},
	"take/fish@water": {
		Name: "fish", Skill: entity.Fishing, Skilled: true, Learn: 0.015, Worth: fed,
		Each: map[*ontology.Class]part{ontology.Fish: {
			Quantity: yields(fishYield), Drain: fishTake, Luck: struct{ Lo, Span float64 }{0.5, 1.0},
		}},
	},
	// Instrumental: wood is only worth something if you lack shelter or
	// craft. A wood felled to the last stand is a clearing.
	//
	// A stand is felled when it is grown. Least was three tenths, from when a
	// wood was a stock that sat near full whatever its age: a day's work
	// takes four tenths out of a tile, so a stand felled at three left
	// nothing standing and the tile became a clearing on the first visit.
	// With stands that come on slowly that emptied whole maps - one seed
	// ended with eleven wooded tiles left. At six tenths a felling thins a
	// wood instead of ending it, and a thicket is left to grow into one.
	"take/timber@wood": {
		Name: "gather wood",
		Each: map[*ontology.Class]part{ontology.Timber: {
			Quantity: fixed(armful), Drain: treeTake, Least: 0.6, Spent: -1,
		}},
		Worth: func(a *entity.Agent, _ *world.World, _ float64) need.Levels {
			want := 0.0
			if a.Inventory[entity.Wood] < raisingTimber {
				want = 0.1*(1-a.Shelter) + 0.03*a.Skills[entity.Crafting]
			}
			return need.Levels{need.Safety: want, need.Esteem: want * 0.3}
		},
		Gone: func(w *world.World, p entity.Pos, t *world.Tile) {
			if t.Wood < 0.1 {
				w.Grid.Turn(p, world.Grass)
				t.Wood = 0
				w.Grid.Sow(w.Grid.Index(p)) // the stand is gone; what comes back starts from nothing
			}
		},
	},
	// Stone is for building; it is worth the safety of the house it will
	// go into, if the agent lacks one.
	"take/stone@outcrop": {
		Name: "quarry", Tool: quarryWear, Skill: entity.Building, Skilled: true, Learn: 0.01,
		Each: map[*ontology.Class]part{ontology.Stone: {
			Quantity: yields(func(a *entity.Agent, _ *world.World, _ float64) float64 { return quarryYield(a) }),
		}},
		Worth: func(a *entity.Agent, _ *world.World, q float64) need.Levels {
			return need.Levels{need.Safety: 0.06 * (1 - a.Shelter) * q, need.Esteem: 0.03}
		},
		Gives: need.Levels{need.Esteem: 0.03},
	},

	// A seller brings whatever is over their keep, of anything, once they
	// have enough of something for the trip to be worth making. Savings buy
	// safety; being a seller of note buys a little esteem.
	"exchange/material>coin@market": {
		Name: "sell",
		Each: map[*ontology.Class]part{
			ontology.Provision: {Quantity: surplus(3), Spare: 4},
			ontology.Tool:      {Quantity: surplus(0), Spare: 1},
			ontology.Meal:      {Quantity: surplus(2), Spare: 3},
			ontology.Stone:     {Quantity: surplus(3), Spare: 4},
		},
		Worth: levels(need.Levels{need.Safety: 0.08, need.Esteem: 0.03}),
		Gives: need.Levels{need.Esteem: 0.03},
		Event: event.Traded,
		Report: func(a, _ *entity.Agent, _, paid float64) string {
			return fmt.Sprintf("%s sold goods for %.1f", a.Name, paid)
		},
	},
	"exchange/coin>provision@market": {
		Name:  "buy food",
		Each:  map[*ontology.Class]part{ontology.Provision: {Quantity: fixed(1), Want: 1, Least: 1}},
		Worth: levels(need.Levels{need.Physiological: 0.3}),
		Event: event.Traded,
		Report: func(a, _ *entity.Agent, _, paid float64) string {
			return fmt.Sprintf("%s bought food for %.2f", a.Name, paid)
		},
	},

	// Stealing is the option that makes conscience mean something. It is
	// fast, it works, and the only thing standing against it is what the
	// agent believes about itself and what it thinks the neighbours will
	// make of it. The victim always knows; bystanders may or may not care.
	"transfer/provision<holder": {
		Name: "steal", Regard: -0.6, Rift: 0.3,
		Each: map[*ontology.Class]part{ontology.Provision: {Quantity: fixed(1), Want: 1}},
		Worth: func(a *entity.Agent, _ *world.World, _ float64) need.Levels {
			return need.Levels{need.Physiological: foodValue(a) * 1.5}
		},
		Event: event.Stolen,
		Report: func(a, o *entity.Agent, _, _ float64) string {
			return fmt.Sprintf("%s stole food from %s", a.Name, o.Name)
		},
	},
	// Giving costs the giver and helps the receiver. Nothing in the need
	// model rewards it much; the charitable do it because their conscience
	// pays them.
	"transfer/provision>needy": {
		Name: "give", Regard: 0.4, Bonds: struct{ Theirs, Mine float64 }{0.15, 0.1},
		Each:  map[*ontology.Class]part{ontology.Provision: {Quantity: fixed(1), Spare: 2}},
		Worth: levels(need.Levels{need.Belonging: 0.08, need.Esteem: 0.05}),
		Gives: need.Levels{need.Belonging: 0.08, need.Esteem: 0.05},
		Event: event.Given,
		Report: func(a, o *entity.Agent, _, _ float64) string {
			return fmt.Sprintf("%s gave food to %s", a.Name, o.Name)
		},
	},
}

// far is the other party to a move: where their store of a material is,
// how to find them, and who they are if anybody.
type far struct {
	// Find is where to go.
	Find func(a *entity.Agent, w *world.World) (entity.Pos, bool)
	// Who is the person on the other side, if the far side is a person.
	Who func(a *entity.Agent, w *world.World, radius int) *entity.Agent
	// Store is their store of m, standing at p, or nil if it is not
	// there. A person is looked for within radius: as far as one would go
	// when sizing the act up, within arm's reach when doing it.
	Store func(a *entity.Agent, w *world.World, p entity.Pos, m *ontology.Class, radius int) store
	// There reports whether the far side is still at p.
	There func(a *entity.Agent, w *world.World, p entity.Pos) bool
	// Priced says coin moves the other way, at the market's price.
	Priced bool
}

// farSide reads the other party off the schema: a person in a role, the
// market, or a kind of ground. parts are the terms, for what the ground
// must hold to be worth walking to.
func farSide(in ontology.Instance, parts []*ontology.Class, terms map[*ontology.Class]part) (far, bool) {
	sc, site := in.Schema, in.Site
	switch {
	case sc.Role != nil:
		role, ok := whom[sc.Role]
		if !ok || len(parts) != 1 {
			return far{}, false
		}
		other := role(parts[0])
		who := func(a *entity.Agent, w *world.World, radius int) *entity.Agent {
			return nearestWith(a, w, radius, other)
		}
		return far{
			Find: func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
				if o := who(a, w, reachRadius); o != nil {
					return o.Pos, true
				}
				return entity.Pos{}, false
			},
			Who: who,
			Store: func(a *entity.Agent, w *world.World, _ entity.Pos, m *ontology.Class, radius int) store {
				o := who(a, w, radius)
				if o == nil {
					return nil
				}
				s, _ := pack(o, m)
				return s
			},
			There: func(a *entity.Agent, w *world.World, _ entity.Pos) bool { return who(a, w, 2) != nil },
		}, true
	case site == ontology.Market:
		return far{
			Find: atMarket,
			Store: func(_ *entity.Agent, w *world.World, _ entity.Pos, m *ontology.Class, _ int) store {
				s, _ := shelf(w, m)
				return s
			},
			There:  func(*entity.Agent, *world.World, entity.Pos) bool { return true },
			Priced: sc.Verb == ontology.Exchange,
		}, true
	case site != nil:
		g, ok := grounds[site]
		if !ok || len(parts) != 1 {
			return far{}, false
		}
		m, t := parts[0], terms[parts[0]]
		// What the ground holds of this material, looked up once here rather
		// than on every tile the search below walks over.
		held := world.StockOf(m)
		return far{
			Find: func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
				// What this search is looking for stands on ground of a
				// kind, so the counts can turn the whole search away and
				// step over the ground between what is left of it. Where
				// the answer only stands *near* that ground - a bank, in
				// reach of water - only the first of those is true, and
				// the reach is asked over the margin. See
				// world.Grid.NearestOfKind.
				kinds := g.Kinds
				if g.Margin > 0 {
					if !w.Grid.AnyWithin(a.Pos, searchRadius+g.Margin, kinds) {
						return entity.Pos{}, false
					}
					kinds = 0
				}
				return w.Grid.NearestOfKind(a.Pos, searchRadius, kinds, func(p entity.Pos, tile *world.Tile) bool {
					if !g.Here(w, p, tile) {
						return false
					}
					drawn := g.Drawn(w, p)
					if drawn == nil {
						return false
					}
					// Ground with no stock of it holds no end of it.
					return held == nil || *held(drawn) >= t.Least
				})
			},
			Store: func(_ *entity.Agent, w *world.World, p entity.Pos, m *ontology.Class, _ int) store {
				tile := g.Drawn(w, p)
				if tile == nil {
					return nil
				}
				return soil(tile, m)
			},
			There: func(_ *entity.Agent, w *world.World, p entity.Pos) bool {
				return g.Here(w, p, w.Grid.At(p)) && g.Drawn(w, p) != nil
			},
		}, true
	}
	return far{}, false
}

// moving is the act the ontology entails for a move with terms, carried
// out from them, or nil where there are none, none fit the material, or
// the far side is nowhere the interpreter knows how to reach.
func moving(in ontology.Instance) *Def {
	mv, ok := moves[in.Key]
	if !ok {
		return nil
	}
	object := in.Object
	if object == nil {
		// An exchange names its material as the input or the output.
		object = in.Schema.Output
		if in.Schema.Output == ontology.Coin {
			object = in.Schema.Inputs[0]
		}
	}
	// The materials this move covers, in the order the trees declare them.
	var parts []*ontology.Class
	for _, c := range ontology.Material.Family() {
		if _, ok := mv.Each[c]; ok && c.IsA(object) {
			if _, carried := world.GoodOf(c); carried {
				parts = append(parts, c)
			}
		}
	}
	if len(parts) == 0 {
		return nil
	}
	// Which way it runs: to the actor unless the trees say otherwise.
	receiving := true
	switch in.Schema.Verb {
	case ontology.Exchange:
		receiving = in.Schema.Output != ontology.Coin
	case ontology.Transfer:
		receiving = in.Schema.Dir == ontology.Seize
	}
	other, ok := farSide(in, parts, mv.Each)
	if !ok {
		return nil
	}
	tech := world.Tech(in.Tech)
	d := &Def{Name: mv.Name, Ticks: in.Ticks, Target: other.Find}
	// What the move brings and what it spends: the parts one way, coin the
	// other at the market. A move that gives spends whichever part it has
	// to spare.
	var gets, gives []*ontology.Class
	if receiving {
		gets = parts
		if other.Priced {
			gives = []*ontology.Class{ontology.Coin}
		}
	} else {
		gives = parts
		if other.Priced {
			gets = []*ontology.Class{ontology.Coin}
		}
	}
	d.Supply = supply(gets, gives, !receiving)

	if other.Who != nil {
		d.With = func(a *entity.Agent, w *world.World, _ entity.Pos) *entity.Agent { return other.Who(a, w, reachRadius) }
	}
	// mine reports whether the actor's side of a part allows the move.
	mine := func(a *entity.Agent, c *ontology.Class) bool {
		t := mv.Each[c]
		s, _ := pack(a, c)
		if receiving {
			return t.Want == 0 || s.Held() < t.Want
		}
		return s.Held() >= t.Spare
	}
	d.Available = func(a *entity.Agent, w *world.World) bool {
		if mv.Tool > 0 && a.Inventory[entity.Tools] < 0.5 {
			return false
		}
		if tech != "" && !known(a, w, tech, mv.Name) {
			return false
		}
		// Somebody to deal with, and something worth dealing in.
		if other.Who != nil && other.Who(a, w, reachRadius) == nil {
			return false
		}
		for _, c := range parts {
			if !mine(a, c) {
				continue
			}
			if other.Who != nil || other.Priced && !receiving {
				return true // the ground and the market are checked on arrival
			}
			if other.Priced {
				t := mv.Each[c]
				square, _ := w.NearestMarket(a.Pos)
				theirs := other.Store(a, w, square, c, reachRadius)
				purse, _ := pack(a, ontology.Coin)
				if theirs.Held() >= t.Least && purse.Held() >= price(w, c)*t.Quantity(a, w, theirs, nil, 1) {
					return true
				}
				continue
			}
			return true
		}
		return false
	}
	d.Expect = func(a *entity.Agent, w *world.World, target entity.Pos) need.Levels {
		var q float64
		for _, c := range parts {
			theirs := other.Store(a, w, target, c, reachRadius)
			if theirs == nil {
				return need.Levels{}
			}
			s, _ := pack(a, c)
			from, to := theirs, s
			if !receiving {
				from, to = s, theirs
			}
			q += mv.Each[c].Quantity(a, w, from, to, 1)
		}
		return mv.Worth(a, w, q)
	}
	d.Apply = func(a *entity.Agent, w *world.World) {
		if !other.There(a, w, a.Pos) {
			return // cleared, fished out, or moved on while we walked
		}
		var o *entity.Agent
		if other.Who != nil {
			o = other.Who(a, w, 2)
		}
		var moved, paid float64
		for _, c := range parts {
			t := mv.Each[c]
			theirs := other.Store(a, w, a.Pos, c, 2)
			s, _ := pack(a, c)
			from, to := theirs, s
			if !receiving {
				from, to = s, theirs
			}
			if receiving && from.Held() <= t.Spent {
				continue // taken while we walked
			}
			luck := t.Luck.Lo
			if t.Luck.Span > 0 {
				luck += t.Luck.Span * w.RNG.Float64()
			}
			if luck == 0 {
				luck = 1
			}
			q := t.Quantity(a, w, from, to, luck)
			if q <= 0 {
				continue
			}
			to.Move(q)
			drain := q
			if t.Drain > 0 {
				drain = t.Drain
			}
			from.Move(-drain)
			moved += q
			if other.Priced {
				paid += q * price(w, c)
			}
		}
		if other.Priced {
			purse, _ := pack(a, ontology.Coin)
			if receiving {
				purse.Move(-paid)
			} else {
				purse.Move(paid)
			}
		}
		if mv.Tool > 0 {
			a.Inventory[entity.Tools] = max(0, a.Inventory[entity.Tools]-mv.Tool)
		}
		if mv.Skilled {
			a.Learn(mv.Skill, mv.Learn)
		}
		if o != nil {
			o.Judge(a.ID, mv.Regard, w.Tick)
			if mv.Bonds.Theirs > 0 {
				o.AddBond(a.ID, mv.Bonds.Theirs)
			}
			if mv.Bonds.Mine > 0 {
				a.AddBond(o.ID, mv.Bonds.Mine)
			}
			if mv.Rift > 0 {
				if b := o.Look(a.ID); b != nil {
					b.Strength = belief.Clamp(b.Strength - mv.Rift)
				}
			}
		}
		for tier, gain := range mv.Gives {
			if gain > 0 {
				a.Needs.Add(need.Tier(tier), gain)
			}
		}
		if mv.Gone != nil {
			if tile := grounds[in.Site]; tile.Drawn != nil {
				mv.Gone(w, a.Pos, tile.Drawn(w, a.Pos))
			}
		}
		if o != nil {
			witness(a, w, mv.Name)
			remorse(a, mv.Name)
		}
		if mv.Report != nil {
			to := entity.ID(0)
			if o != nil {
				to = o.ID
			}
			w.Emit(mv.Event, a.ID, to, "%s", mv.Report(a, o, moved, paid))
		}
	}
	return d
}

// mover is the move the ontology entails under key, for the acts the rest
// of the package refers to by name.
func mover(key string) *Def {
	d := moving(instances[key])
	if d == nil {
		panic("action: no move for " + key)
	}
	return d
}

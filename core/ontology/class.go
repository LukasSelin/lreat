// Package ontology is what the world is made of, arranged so that what can
// be done with it follows from what it is.
//
// Three trees and a set of relations. Things are what an act operates on:
// materials, people, and the practices themselves. Sites are where an act
// happens. Roles are how a person stands relative to the actor. Verbs are
// schemas over classes; the catalog of concrete actions is instantiated by
// walking the trees, so adding a class adds every act the trees entail for
// it and removing one removes them. A class exists only if some verb treats
// it differently from its siblings; anything else is a trait.
//
// This package is a leaf. It knows nothing of agents, actions, or the
// world; it names them.
package ontology

import (
	"strings"

	"lreat/core/habit"
)

// Trait is a property a class has or inherits. Traits are what verbs test:
// Consume wants Edible, Raise wants Buildable, a hearth is wherever Hearth
// holds.
type Trait uint32

const (
	// Material traits.
	Edible Trait = 1 << iota
	Perishable
	Burnable
	Buildable
	Heavy
	Wears // a tool: multiplies work and is used up by it

	// Site traits.
	Living // carries something growing: what stands here came on with age
	Roofed
	Owned
	Public
	Passable

	// Workplaces. A dwelling has all of them; the market lends a bench and
	// a desk to those without one, the tavern a hearth.
	Bench
	Hearth
	Forge
	Desk
	Company
	Trade
	Store
)

var traitNames = map[Trait]string{
	Edible: "edible", Perishable: "perishable", Burnable: "burnable", Buildable: "buildable",
	Heavy: "heavy", Wears: "wears", Living: "living", Roofed: "roofed", Owned: "owned", Public: "public",
	Passable: "passable", Bench: "bench", Hearth: "hearth", Forge: "forge", Desk: "desk",
	Company: "company", Trade: "trade", Store: "store",
}

func (t Trait) String() string {
	if s, ok := traitNames[t]; ok {
		return s
	}
	var parts []string
	for b := Trait(1); b != 0; b <<= 1 {
		if t&b != 0 {
			parts = append(parts, traitNames[b])
		}
	}
	return strings.Join(parts, "+")
}

// Class is a node in one of the trees. Single inheritance: a class has one
// parent and every trait of its ancestors. Prior is the delta this class
// adds to the moment of lacking it, on top of whatever its ancestors add:
// for a material its stock coordinate, for a structure what having none
// is like. At is the moment of being there, for sites; see prior.go.
type Class struct {
	Name   string
	Parent *Class
	Traits Trait
	Prior  habit.Signature
	At     habit.Signature
	// Lack is how strongly being short of this registers, for a material:
	// what a class adds on the lack coordinate when an act would bring it,
	// and on the stock coordinate when an act would spend it. Inherited.
	Lack float64

	// Warmth is which half of the year getting this belongs to: positive
	// for the green half, negative for the cold one, zero for a thing had
	// in any weather. See season. How long a thing takes to come on is the
	// other half of its time and is a Process; see process.go.
	Warmth float64

	children []*Class
}

// New declares a class under parent. Declaration order is preserved and is
// the order leaves are walked in, which keeps instantiation deterministic.
func New(name string, parent *Class, traits Trait, prior habit.Signature) *Class {
	c := &Class{Name: name, Parent: parent, Traits: traits, Prior: prior}
	if parent != nil {
		parent.children = append(parent.children, c)
	}
	return c
}

// IsA reports whether c is anc or descends from it.
func (c *Class) IsA(anc *Class) bool {
	for x := c; x != nil; x = x.Parent {
		if x == anc {
			return true
		}
	}
	return false
}

// All is every trait c has, its own and inherited.
func (c *Class) All() Trait {
	var t Trait
	for x := c; x != nil; x = x.Parent {
		t |= x.Traits
	}
	return t
}

// Has reports whether c has every bit of t, own or inherited.
func (c *Class) Has(t Trait) bool { return c.All()&t == t }

// Leaves are the classes under c with no children, in declaration order.
// A childless c is its own leaf.
func (c *Class) Leaves() []*Class {
	if len(c.children) == 0 {
		return []*Class{c}
	}
	var out []*Class
	for _, k := range c.children {
		out = append(out, k.Leaves()...)
	}
	return out
}

// Family is c and every class under it, in declaration order.
func (c *Class) Family() []*Class {
	out := []*Class{c}
	for _, k := range c.children {
		out = append(out, k.Family()...)
	}
	return out
}

// Children are the classes directly under c.
func (c *Class) Children() []*Class { return c.children }

// Path is the ancestry from root to c, for reports.
func (c *Class) Path() string {
	var parts []string
	for x := c; x != nil; x = x.Parent {
		parts = append([]string{x.Name}, parts...)
	}
	return strings.Join(parts, "/")
}

// DerivedPrior is the sum of Prior deltas from root to c.
func (c *Class) DerivedPrior() habit.Signature {
	var s habit.Signature
	for x := c; x != nil; x = x.Parent {
		for i := range s {
			s[i] += x.Prior[i]
		}
	}
	return s
}

// Things: what an act operates on.
var (
	Thing    = New("thing", nil, 0, habit.Signature{})
	Material = New("material", Thing, 0, habit.Signature{})

	// Provisions are what feeds. They are one class to Consume, which
	// picks whatever is on hand, and several to Take, because fishing and
	// foraging and hunting are different habits.
	// A class's prior is its stock coordinate and nothing else: what it is
	// for comes through Wanting, and what taking it takes is in takeDetail.
	Provision = lack(New("provision", Material, Edible, habit.Signature{}), 0.8)
	// The brush a forager and a trapper live off is one stand and comes on
	// as one: a couple of years from a clearing being made. Picking it
	// keeps the season's hours and hunting over it does not, which is the
	// two facts pulling apart on the same stand.
	Berries = season(New("berries", Provision, Perishable, habit.Signature{}), 0.3)
	Game    = New("game", Provision, Perishable, habit.Signature{})
	// A shoal is a stock and not a crop: it comes back without having to
	// come on, so there is no age on it to wait out.
	Fish = New("fish", Provision, Perishable, habit.Signature{})
	// The fastest thing the year makes, and the most of its season: sown,
	// in ear within the month, and cut before it is anything else.
	Grain = season(New("grain", Provision, 0, habit.Signature{}), 0.5)
	Meal  = New("meal", Provision, Perishable, habit.Signature{})

	// The slowest thing the year makes, and the one thing here had in the
	// cold half rather than the green one: a stand a planter raised is
	// years off being beams, and the felling of it is winter work whoever
	// does it.
	Timber = season(lack(New("timber", Material, Burnable|Buildable, habit.Signature{}), 0.6), -0.3)
	Stone  = lack(New("stone", Material, Buildable|Heavy, habit.Signature{}), 0.5)
	Tool   = lack(New("tool", Material, Wears, habit.Signature{habit.Unproven: 0.8, habit.Skill: 0.5}), 0.5)
	// Coin is a thing so that money can be made, given, and stolen like
	// anything else. A young settlement barters; coin arrives when
	// somebody mints it.
	Coin = lack(New("coin", Material, 0, habit.Signature{}), 0.6)

	// Person is one class. How a person stands to the actor is a Role, not
	// a subclass: nobody is a pupil the way an oak is timber.
	Person = New("person", Thing, 0, habit.Signature{habit.Rapport: 0.5})
	// Practice is an act itself as the object of another: what is taught
	// and studied. The ontology contains its own catalog.
	Practice = New("practice", Thing, 0, habit.Signature{})
)

// Sites: where an act happens.
var (
	Site = New("site", nil, 0, habit.Signature{})
	// Being near to hand is what every kind of ground has in common, so it
	// is said here once rather than on each of them.
	Ground = at(New("ground", Site, Passable, habit.Signature{}), habit.Signature{habit.Near: 0.5})
	// Open ground is unclaimed grass: nothing to take, room to build.
	Open = New("open", Ground, 0, habit.Signature{})
	// A wood and a field are Living: what a taking draws on there is not a
	// seam that is simply there, it is a standing crop, and a stand has an
	// age. A thicket a planter raised this spring is not an old wood, and a
	// strip cut yesterday is not a strip in ear. What such a site holds is
	// bounded by how long it has been growing - see world.Tile.Grown - which
	// is why planting is for those who come after, and why a holding is
	// worked in turn rather than all at once.
	//
	// The other grounds are not. An outcrop is stone and does not grow; the
	// water's fish come back, but a shoal is a stock that replenishes, not a
	// crop that has to come on before it can be cut.
	Wood    = New("wood", Ground, Living, habit.Signature{})
	Water   = New("water", Ground, 0, habit.Signature{})
	Outcrop = New("outcrop", Ground, 0, habit.Signature{})
	Field   = New("field", Ground, Living|Owned, habit.Signature{})

	Built = New("built", Site, 0, habit.Signature{})
	// Lacking a roof is the unsafe, unsheltered moment, and a little more
	// so for whoever is out in the cold - exposure, not the weather: the
	// weather is the same news to everybody, and an act that names it is
	// one the whole settlement turns to at once. Lacking a granary or a
	// tavern is not a want of the body but of standing: they belong to
	// those with something to win and a settlement they mean to stay in.
	Dwelling = at(New("dwelling", Built, Roofed|Owned|Bench|Hearth|Forge|Desk,
		habit.Signature{habit.Unsafe: 1, habit.Shelter: -1, habit.Exposure: 0.3}), habit.Signature{habit.Near: 0.4})
	// Lacking a square is wanting somewhere to bring what one has made.
	// It is not a want of the body: it belongs to a person with more on
	// hand than they need, standing to win, and neighbours far enough off
	// that the walk to the old square is a day's work.
	Market = at(New("market", Built, Public|Bench|Desk|Trade,
		habit.Signature{habit.Unproven: 0.6, habit.Stock: 0.6, habit.Company: 0.4, habit.Charity: 0.3, habit.Tradition: 0.3}),
		habit.Signature{habit.Near: 0.4})
	Granary = New("granary", Built, Roofed|Public|Store, habit.Signature{habit.Unproven: 0.7, habit.Charity: 0.4, habit.Tradition: 0.4})
	Tavern  = at(New("tavern", Built, Roofed|Public|Hearth|Company,
		habit.Signature{habit.Unproven: 0.6, habit.Lonely: 0.4, habit.Charity: 0.4, habit.Tradition: 0.3}), habit.Signature{habit.Company: 0.6})
	// A road is a public work like the granary and the tavern, and what
	// wanting one is like is said here for the same reason theirs is: it is
	// a fact about roads, not about the one schema that happens to raise
	// them. It belongs to somebody already roofed - the timber is what is
	// left over once there is a roof, not what is chosen instead of one -
	// with neighbours around to walk it and enough charity to spend a day
	// on ground that is nobody's. It says nothing of hunger and nothing of
	// skill: a newcomer believes itself unskilled at everything, and a
	// prior that mentions skill taxes exactly the acts a young settlement
	// needs.
	Road = New("road", Built, Passable|Public, habit.Signature{
		habit.Shelter: 0.7, habit.Company: 0.6, habit.Charity: 0.5, habit.Industry: 0.3})
)

// Time, as the ontology has it, in two halves that are held apart on
// purpose.
//
// How long a thing takes to come on is a Process: stages, on a clock, over
// the thing they happen to. See process.go, which states them, and note that
// nothing about that is a fact about the *getting* - a wood takes six years
// whether anybody fells it or not.
//
// Which half of the year the getting belongs to is Warmth, and it is here,
// on the thing, because it is a fact about the thing. It is signed, because
// the year has two halves and things are had in both: grain and berries are
// of the green half, and felling is winter work whoever does it, the sap
// being down and there being least else to do. The two are independent,
// which is the point of holding both - timber is slow and of the cold half,
// a crop is quick and wholly of the warm one, and neither can be read off
// the other.
//
// Before this the whole of the ontology's account of the second was a chill
// coordinate written by hand onto two classes, with nothing saying why
// berries carried one and stone did not, a third hidden in the take detail
// for timber, and a fourth in the residue of clearing a field.
//
// What reads it: the composed prior takes Warmth as the chill coordinate, so
// an act's season follows from what it is about rather than from a number
// somebody wrote on it.
func season(c *Class, warmth float64) *Class {
	c.Warmth = warmth
	c.Prior[habit.Chill] = -warmth
	return c
}

// lack sets how strongly being short of a material registers.
func lack(c *Class, l float64) *Class {
	c.Lack = l
	return c
}

// Short is how strongly being short of c registers, walking up the tree.
func (c *Class) Short() float64 {
	for x := c; x != nil; x = x.Parent {
		if x.Lack != 0 {
			return x.Lack
		}
	}
	return 0
}

// HeavyLoad is what an armful of a heavy material is to carry, in armfuls of
// anything else. A creel of stone is not a sack of grain: the mason walks
// slower under it, rests sooner, and minds the rough ground more, and twice
// is a modest reading of a difference that is nearer threefold by weight.
//
// One number, because Heavy is one trait. The ontology says which materials
// are heavy and how much that means; what carrying it comes to for a
// particular walker is world.Hauled, which is where the names here are bound
// to the goods in a pack.
const HeavyLoad = 2

// Heft is what one unit of c is to carry, in armfuls. It is the trait read
// as a number, and it is the whole of what Heavy means anywhere.
func Heft(c *Class) float64 {
	if c.Has(Heavy) {
		return HeavyLoad
	}
	return 1
}

// at sets what being at a site is like, as a delta on what being at the kind
// of site it is is like. See DerivedAt.
func at(c *Class, s habit.Signature) *Class {
	c.At = s
	return c
}

// DerivedAt is the sum of At deltas from root to c, the way DerivedPrior is
// for Prior. Being at a site is being at every kind of site it is: what is
// true of standing on ground is true of standing on a meadow, and saying it
// on the meadow, the wood, the water, the outcrop and the field separately
// left it unsaid for anybody standing on ground as such. The one act whose
// site is ground itself is paving - a road may be laid on any of them - and
// it was the one act that fell through the gap.
func (c *Class) DerivedAt() habit.Signature {
	var s habit.Signature
	for x := c; x != nil; x = x.Parent {
		for i := range s {
			s[i] += x.At[i]
		}
	}
	return s
}

// Role is how a person stands to the actor, decided at the moment of acting
// rather than fixed in a tree.
type Role struct {
	Name  string
	Prior habit.Signature
}

var (
	// Self is the actor: study is teaching turned inward.
	Self = Role{"self", habit.Signature{habit.Curious: 1, habit.Hunger: -0.5, habit.Unsafe: -0.3, habit.Tradition: -0.4}}
	// Neighbour is anyone within range.
	Neighbour = Role{"neighbour", habit.Signature{habit.Lonely: 0.5, habit.Company: 0.5, habit.Rapport: 0.5}}
	// Requester has asked for something the actor can supply.
	Requester = Role{"requester", habit.Signature{habit.Unproven: 0.5, habit.Industry: 0.5, habit.Skill: 0.5, habit.Lack: 0.5}}
	// Needy has none of something and feels the lack.
	Needy = Role{"needy", habit.Signature{habit.Lonely: 0.3, habit.Rapport: 0.5}}
	// Holder has more of something than they need.
	Holder = Role{"holder", habit.Signature{habit.Hunger: 1, habit.Rapport: -0.4}}
	// Pupil reaches less far in some practice than the actor. Teaching is
	// the moment of having a skill to show and someone to show it to; what
	// is wanted of the practice is all on this side.
	Pupil = Role{"pupil", habit.Signature{habit.Unproven: 0.8, habit.Company: 0.7, habit.Charity: 0.3, habit.Rapport: 0.5, habit.Skill: 1}}
	// Wrongdoer owes the actor for a wrong done.
	Wrongdoer = Role{"wrongdoer", habit.Signature{habit.Unproven: 0.6, habit.Rapport: 1, habit.Caution: -0.4}}
)

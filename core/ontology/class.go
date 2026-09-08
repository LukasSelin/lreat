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

	"lreat/core/clock"
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

	// Time, in the two ways a thing can have any. Comes is how long a stand
	// of this takes to grow before it is all there, in days of growing
	// weather, and Warmth is which half of the year getting it belongs to:
	// positive for the green half, negative for the cold one, zero for a
	// thing had in any weather. See grows.
	Comes  int
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
	Berries = grows(New("berries", Provision, Perishable, habit.Signature{}), 2*clock.Year, 0.3)
	Game    = grows(New("game", Provision, Perishable, habit.Signature{}), 2*clock.Year, 0)
	// A shoal is a stock and not a crop: it comes back without having to
	// come on, so there is no age on it to wait out.
	Fish = New("fish", Provision, Perishable, habit.Signature{})
	// The fastest thing the year makes, and the most of its season: sown,
	// in ear within the month, and cut before it is anything else.
	Grain = grows(New("grain", Provision, 0, habit.Signature{}), clock.Month, 0.5)
	Meal  = New("meal", Provision, Perishable, habit.Signature{})

	// The slowest thing the year makes, and the one thing here had in the
	// cold half rather than the green one: a stand a planter raised is
	// years off being beams, and the felling of it is winter work whoever
	// does it.
	Timber = grows(lack(New("timber", Material, Burnable|Buildable, habit.Signature{}), 0.6), 6*clock.Year, -0.3)
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
	Site   = New("site", nil, 0, habit.Signature{})
	Ground = New("ground", Site, Passable, habit.Signature{})
	// Open ground is unclaimed grass: nothing to take, room to build.
	Open = at(New("open", Ground, 0, habit.Signature{}), habit.Signature{habit.Near: 0.5})
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
	Wood    = at(New("wood", Ground, Living, habit.Signature{}), habit.Signature{habit.Near: 0.5})
	Water   = at(New("water", Ground, 0, habit.Signature{}), habit.Signature{habit.Near: 0.5})
	Outcrop = at(New("outcrop", Ground, 0, habit.Signature{}), habit.Signature{habit.Near: 0.5})
	Field   = at(New("field", Ground, Living|Owned, habit.Signature{}), habit.Signature{habit.Near: 0.5})

	Built = New("built", Site, 0, habit.Signature{})
	// Lacking a roof is the unsafe, unsheltered moment, and a little more
	// so for whoever is out in the cold - exposure, not the weather: the
	// weather is the same news to everybody, and an act that names it is
	// one the whole settlement turns to at once. Lacking a granary or a
	// tavern is not a want of the body but of standing: they belong to
	// those with something to win and a settlement they mean to stay in.
	Dwelling = at(New("dwelling", Built, Roofed|Owned|Bench|Hearth|Forge|Desk,
		habit.Signature{habit.Unsafe: 1, habit.Shelter: -1, habit.Exposure: 0.3}), habit.Signature{habit.Near: 0.4})
	Market  = at(New("market", Built, Public|Bench|Desk|Trade, habit.Signature{}), habit.Signature{habit.Near: 0.4})
	Granary = New("granary", Built, Roofed|Public|Store, habit.Signature{habit.Unproven: 0.7, habit.Charity: 0.4, habit.Tradition: 0.4})
	Tavern  = at(New("tavern", Built, Roofed|Public|Hearth|Company,
		habit.Signature{habit.Unproven: 0.6, habit.Lonely: 0.4, habit.Charity: 0.4, habit.Tradition: 0.3}), habit.Signature{habit.Company: 0.6})
	Road = New("road", Built, Passable, habit.Signature{})
)

// Time, as the ontology has it.
//
// The year turns in core/world and the calendar is in core/clock, and
// neither of them is a statement about what things are. This is: some things
// the year makes and some things are simply there, and everything the season
// does to what can be done follows from which. Before it, the whole of the
// ontology's account of time was a chill coordinate written by hand onto two
// classes, with no concept behind it saying why berries carried one and
// stone did not, and how long a stand took to come on lived out in the map,
// which is the one place it is not a fact about the thing.
//
// Two facts, and they are independent, which is the point of having both.
// Comes is how long the growing takes: timber is years and a crop is a
// month, and a thing with no Comes - stone, a shoal of fish - is a stock
// that is there or is not, rather than a crop that has to come on. Warmth is
// which half of the year the getting belongs to, and it is signed, because
// the year has two halves and things are had in both: berries and grain are
// of the green half, and felling is winter work whoever does it, the sap
// being down and there being least else to do. A thing can therefore be slow
// and of the cold half, or quick and wholly of the warm one, and the two say
// different things about it.
//
// What reads them: the composed prior takes Warmth as the chill coordinate,
// so an act's season follows from what it is about rather than from a number
// somebody wrote on it; and world.Tile.Grown takes Comes, so the map asks
// the ontology how long a wood is in coming instead of keeping its own
// table of ages.
func grows(c *Class, comes int, warmth float64) *Class {
	c.Comes, c.Warmth = comes, warmth
	c.Prior[habit.Chill] = -warmth
	return c
}

// Ripens is how long a stand of c takes to come on, walking up the tree as
// Short does. Zero means the year does not make this: it is a stock, and
// what there is of it is what there is.
func (c *Class) Ripens() int {
	for x := c; x != nil; x = x.Parent {
		if x.Comes != 0 {
			return x.Comes
		}
	}
	return 0
}

// Grown reports whether c is something the year makes.
func (c *Class) Grown() bool { return c.Ripens() != 0 }

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

// at sets what being at a site is like.
func at(c *Class, s habit.Signature) *Class {
	c.At = s
	return c
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
	// Child is one the actor bore and that is not grown yet. It is the one
	// role nobody chooses and nobody competes for: a needy stranger is
	// somebody a charitable person may or may not notice, and a child is
	// already yours. So the moment is charity and custom rather than
	// sociability - the warmth is high because the tie is, and there is no
	// coordinate for what the actor gets out of it, because there is not
	// much.
	Child = Role{"child", habit.Signature{habit.Charity: 0.6, habit.Tradition: 0.5, habit.Rapport: 0.7, habit.Company: 0.3}}
)

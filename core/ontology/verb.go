package ontology

import (
	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/habit"
)

// Verb is a kind of act: what changes when it is done. Eleven cover the
// catalog. Each is one Effect over object and site classes; the concrete
// acts are the schemas below, instantiated over the trees.
type Verb int

const (
	Take     Verb = iota // site stock into inventory
	Make                 // materials into a material, at a workplace
	Raise                // materials into a structure, on ground
	Tend                 // a change to a tile
	Consume              // a material into needs
	Dwell                // presence at a site into needs
	Exchange             // a material for another, at a market
	Transfer             // a material to or from a person
	Pass                 // a practice to a person, or to oneself
	Strike               // harm to a person
	Move                 // home to another site
	VerbCount
)

var verbNames = [VerbCount]string{
	"take", "make", "raise", "tend", "consume", "dwell", "exchange", "transfer", "pass", "strike", "move",
}

func (v Verb) String() string { return verbNames[v] }

// Direction is which way a Transfer runs. Seize is a gift the other way
// round, and everything that makes theft theft is in the traits and the
// valence, not in a separate verb.
type Direction int

const (
	Give Direction = iota
	Seize
)

// Schema is an act stated over classes. Instantiate expands it over the
// leaves of Object and Site unless told to collapse, and only where the
// relations allow: Take needs the site to afford the object.
//
// Name names the change where the classes alone do not say it: what a Tend
// does to the ground, what a Dwell is for. Inputs and Output are what Make
// and Raise and Exchange turn into what. SiteTrait is an alternative to
// Site: any site with that trait, which is how a workplace is named without
// saying whose. Prior and Valence are what is irreducible to the act's
// verb, classes, site, and role, and should stay small; if one grows, a
// class or trait is missing.
type Schema struct {
	Verb           Verb
	Name           string
	Object         *Class
	CollapseObject bool
	Inputs         []*Class
	// Optional inputs improve the output without being needed for it: a
	// stone in the walls of a house. They count for what a material is
	// wanted for, not for whether the act can start.
	Optional     []*Class
	Output       *Class
	Site         *Class
	CollapseSite bool
	SiteTrait    Trait
	Role         *Role
	Dir          Direction

	Ticks   int
	Skill   entity.Skill
	Skilled bool
	// Tech gates the act on a discovery, the settlement's or the actor's.
	Tech string
	// Reach0 is how far into reach the act starts for a newborn.
	Reach0 float64

	Prior   habit.Signature
	Valence belief.Valence
}

// Reach at birth, by kind of act.
const (
	reachEveryday = 1.0
	reachCraft    = 0.5
	reachPave     = 0.5
	reachStudy    = 0.6
	reachTeach    = 0.3
	reachFish     = 0.4
	reachHunt     = 0.5
	reachWater    = 0.3
	reachForest   = 0.3
	reachCook     = 0.4
	reachQuarry   = 0.4
	reachGranary  = 0.2
	reachSmelt    = 0.2
	reachTavern   = 0.2
)

// Schemas is the catalog stated over classes. Order here does not matter;
// instances are sorted by key.
var Schemas = []Schema{
	// Taking is one schema: every material a site affords, at that site.
	// Skills and reach differ by what is taken, so they are set per leaf
	// in takeDetail rather than here.
	{Verb: Take, Object: Material, Site: Ground, Ticks: 2, Reach0: reachEveryday,
		Valence: belief.Valence{belief.Industry: 0.25}},

	// Tending changes the ground. Clearing is the half of farming that
	// makes a field; harvesting is the Take above.
	{Verb: Tend, Name: "clear", Site: Open, Ticks: 4, Skill: entity.Farming, Skilled: true, Reach0: reachEveryday,
		Prior:   habit.Signature{habit.Chill: -0.5, habit.Industry: 0.7, habit.Skill: 0.3, habit.Lack: 0.3},
		Valence: belief.Valence{belief.Industry: 0.3}},
	{Verb: Tend, Name: "water", Inputs: []*Class{Timber}, Site: Field, Ticks: 4, Skill: entity.Farming, Skilled: true, Reach0: reachWater, Tech: "irrigation",
		Prior:   habit.Signature{habit.Industry: 0.7, habit.Skill: 0.4},
		Valence: belief.Valence{belief.Industry: 0.3}},
	{Verb: Tend, Name: "plant", Site: Open, Ticks: 2, Reach0: reachForest, Tech: "forestry",
		Prior:   habit.Signature{habit.Charity: 0.4, habit.Tradition: 0.4, habit.Lack: 0.5},
		Valence: belief.Valence{belief.Industry: 0.3, belief.Charity: 0.3}},

	// Making, at a workplace.
	{Verb: Make, Inputs: []*Class{Timber}, Output: Tool, SiteTrait: Bench, Ticks: 3, Skill: entity.Crafting, Skilled: true, Reach0: reachCraft,
		Valence: belief.Valence{belief.Industry: 0.3}},
	{Verb: Make, Inputs: []*Class{Stone, Timber}, Output: Tool, SiteTrait: Forge, Ticks: 3, Skill: entity.Crafting, Skilled: true, Reach0: reachSmelt, Tech: "metallurgy",
		Valence: belief.Valence{belief.Industry: 0.3}},
	{Verb: Make, Inputs: []*Class{Provision, Timber}, Output: Meal, SiteTrait: Hearth, Ticks: 2, Reach0: reachCook, Tech: "pottery",
		Prior:   habit.Signature{habit.Hunger: -0.5},
		Valence: belief.Valence{belief.Industry: 0.3}},

	// Raising, on ground.
	{Verb: Raise, Inputs: []*Class{Timber}, Optional: []*Class{Stone}, Output: Dwelling, Site: Open, Ticks: 3, Skill: entity.Building, Skilled: true, Reach0: reachEveryday,
		Valence: belief.Valence{belief.Industry: 0.3, belief.Tradition: 0.1}},
	{Verb: Raise, Inputs: []*Class{Timber, Stone}, Output: Granary, Site: Open, Ticks: 4, Skill: entity.Building, Skilled: true, Reach0: reachGranary, Tech: "masonry",
		Valence: belief.Valence{belief.Industry: 0.3, belief.Charity: 0.3}},
	{Verb: Raise, Inputs: []*Class{Timber}, Output: Tavern, Site: Open, Ticks: 4, Skill: entity.Building, Skilled: true, Reach0: reachTavern, Tech: "brewing",
		Valence: belief.Valence{belief.Industry: 0.3, belief.Charity: 0.3}},
	{Verb: Raise, Inputs: []*Class{Timber}, Output: Road, Site: Ground, CollapseSite: true, Ticks: 2, Reach0: reachPave,
		Prior:   habit.Signature{habit.Shelter: 0.7, habit.Company: 0.6, habit.Charity: 0.5, habit.Industry: 0.3},
		Valence: belief.Valence{belief.Industry: 0.4, belief.Charity: 0.3}},

	// Eating is one act; which provision goes is decided when it is done.
	{Verb: Consume, Object: Provision, CollapseObject: true, Ticks: 1, Reach0: reachEveryday},

	// Being somewhere.
	// Rest is the moment when nothing presses: fed, safe enough, in no
	// particular want of company, of standing, or of anything to wonder
	// at. Naming every urgency negative is what makes it a fallback rather
	// than a rival, since any urgency at all turns an agent away from it.
	// Read as a mild version of eating's moment - which is what it was
	// while outcomes could still move habits and rest could be learned
	// away - a third of every settlement's waking life went on resting off
	// a hunger it was not answering.
	{Verb: Dwell, Name: "rest", Ticks: 1, Reach0: reachEveryday,
		Prior: habit.Signature{
			habit.Hunger: -1, habit.Unsafe: -0.6, habit.Lonely: -0.6,
			habit.Unproven: -0.4, habit.Curious: -0.4,
		},
		Valence: belief.Valence{belief.Industry: -0.4}},
	{Verb: Dwell, Name: "meet", Site: Tavern, Role: &Neighbour, Ticks: 2, Reach0: reachEveryday,
		Prior:   habit.Signature{habit.Lonely: 0.5},
		Valence: belief.Valence{belief.Charity: 0.1, belief.Tradition: 0.15}},
	// Going to look is presence too, and the presence it turns into a need
	// is presence somewhere one has not been. It carries no site on purpose.
	// Every other site class contributes its At to the composed prior, and
	// open ground's At is a nearness; scouting is the one act in the catalog
	// whose moment says nothing whatever about how far off the target is,
	// and a Near in its prior is not a small error - it is the act inverted.
	// What is left is the moment itself: no roof of one's own, fed enough to
	// spare the day, and some curiosity surviving the tiers underneath it.
	{Verb: Dwell, Name: "look", Ticks: 1, Reach0: reachEveryday,
		Prior: habit.Signature{habit.Shelter: -0.7, habit.Hunger: -0.5, habit.Curious: 0.4}},

	{Verb: Dwell, Name: "guard", Site: Market, Ticks: 3, Skill: entity.Guarding, Skilled: true, Reach0: reachEveryday,
		Prior:   habit.Signature{habit.Unsafe: 0.6, habit.Company: 0.5, habit.Order: -1, habit.Charity: 0.4, habit.Tradition: 0.3},
		Valence: belief.Valence{belief.Industry: 0.3, belief.Charity: 0.4}},

	// Trading, at the market. Provision for coin and coin for provision are
	// the same schema with the sides swapped.
	// A seller brings whatever is over their keep, of anything; what a
	// household mostly has over its keep is food, and that is the moment
	// selling belongs to.
	{Verb: Exchange, Inputs: []*Class{Material}, Output: Coin, Site: Market, Ticks: 1, Reach0: reachEveryday,
		Prior:   habit.Signature{habit.Stock: 0.8},
		Valence: belief.Valence{belief.Industry: 0.15}},
	{Verb: Exchange, Inputs: []*Class{Coin}, Output: Provision, Site: Market, Ticks: 1, Reach0: reachEveryday},

	// Handing over, and its inverse.
	{Verb: Transfer, Object: Provision, CollapseObject: true, Role: &Needy, Ticks: 1, Reach0: reachEveryday,
		Prior:   habit.Signature{habit.Hunger: -0.4, habit.Charity: 1},
		Valence: belief.Valence{belief.Charity: 0.8, belief.Honesty: 0.1}},
	{Verb: Transfer, Object: Material, CollapseObject: true, Role: &Requester, Ticks: 3, Skilled: true, Reach0: reachEveryday,
		Valence: belief.Valence{belief.Charity: 0.5, belief.Industry: 0.35}},
	{Verb: Transfer, Object: Provision, CollapseObject: true, Role: &Holder, Dir: Seize, Ticks: 1, Reach0: reachEveryday,
		Prior:   habit.Signature{habit.Order: -0.3, habit.Caution: -0.7},
		Valence: belief.Valence{belief.Honesty: -1, belief.Charity: -0.4}},

	// Knowledge, outward and inward.
	{Verb: Pass, Object: Practice, Role: &Pupil, Ticks: 3, Skilled: true, Reach0: reachTeach,
		Valence: belief.Valence{belief.Charity: 0.5, belief.Industry: 0.2, belief.Tradition: 0.3}},
	{Verb: Pass, Object: Practice, Role: &Self, Ticks: 4, Skill: entity.Scholarship, Skilled: true, Reach0: reachStudy,
		Valence: belief.Valence{belief.Industry: 0.2, belief.Tradition: -0.5}},

	{Verb: Strike, Object: Person, Role: &Wrongdoer, Ticks: 1, Reach0: reachEveryday,
		Valence: belief.Valence{belief.Honesty: 0.6, belief.Charity: -0.4, belief.Tradition: 0.3}},

	// Moving house is the settled moment: under a roof, with wood past
	// what it needed, and the will to do something with it. Left to the
	// verb alone the prior was a bare nearness that fit every moment a
	// little, and agents moved house instead of living in one.
	{Verb: Move, Site: Dwelling, Ticks: 4, Reach0: reachEveryday,
		Prior: habit.Signature{habit.Shelter: 0.7, habit.Industry: 0.5, habit.Near: 0.4}},
}

// takeDetail is what differs between takings of different things: the
// skill drawn on, the reach at birth, how long it takes, and what the
// taking itself is like beyond wanting the thing. Anything afforded but
// unlisted takes the schema's defaults.
var takeDetail = map[*Class]struct {
	Skill   entity.Skill
	Skilled bool
	Reach0  float64
	Ticks   int
	Tech    string
	Prior   habit.Signature
}{
	Berries: {Reach0: reachEveryday, Ticks: 2},
	Game:    {Reach0: reachHunt, Ticks: 3, Tech: "trapping", Prior: habit.Signature{habit.Skill: 0.4}},
	Fish:    {Skill: entity.Fishing, Skilled: true, Reach0: reachFish, Ticks: 2, Tech: "fishing", Prior: habit.Signature{habit.Skill: 0.3}},
	// A harvest is not what hunger calls for; foraging is. It is what an
	// industrious person with a field does whether or not the larder is
	// low, and that tradition is what carries farming through the bad
	// years a first field has.
	Grain: {Skill: entity.Farming, Skilled: true, Reach0: reachEveryday, Ticks: 4,
		Prior: habit.Signature{habit.Hunger: -0.8, habit.Industry: 0.7, habit.Skill: 0.3, habit.Lack: -0.5}},
	// Felling is winter work whoever does it: the sap is down and there is
	// least else to do.
	Timber: {Reach0: reachEveryday, Ticks: 2, Prior: habit.Signature{habit.Chill: 0.3}},
	Stone:  {Skill: entity.Building, Skilled: true, Reach0: reachQuarry, Ticks: 3, Tech: "quarrying", Prior: habit.Signature{habit.Industry: 0.5, habit.Skill: 0.3}},
}

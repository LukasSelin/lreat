package ontology

import "lreat/core/clock"

// Affords says what a site or a holder offers: the ground offers what lies
// in it, a person what is in their pack and their purse, the market what is
// on its shelves. It is the static half of opportunity - what a kind of
// place is good for, before anybody has gone to look at one. Every act that
// moves a material moves it between two holders named by class, and Take
// instantiates one act for every (material, ground) pair here and none
// otherwise: a map with no outcrop has no quarrying.
//
// The dynamic half is not a table and is deliberately not an object. What is
// on offer here, now, to this person is recomputed every tick from the ground
// truth by Def.Available and Def.Target, and an opportunity is only a thing
// worth keeping when it can be wrong or when it can wait - which is what
// entity.Place (believed, and often stale) and entity.Request (posted, and
// outliving the tick) are, and what a candidate recomputed from truth is not.
var Affords = map[*Class][]*Class{
	Wood:    {Berries, Game, Timber, Browse, Mast},
	Open:    {Sward},
	Water:   {Fish},
	Outcrop: {Stone},
	Field:   {Grain},
	Person:  {Material},
	Market:  {Material},
}

// Offers reports whether a holder of class c affords material.
func Offers(c, material *Class) bool {
	for _, m := range Affords[c] {
		if material.IsA(m) {
			return true
		}
	}
	return false
}

// Transform is a change that happens to a thing rather than one somebody
// does to it: food spoils, a house nobody keeps falls in, a field left alone
// goes back to grass. It is the other half of what the world does on its
// own - see Process for the half that grows - and unlike a process it cannot
// be read in advance: a rate is a share lost where what changes is a
// quantity, and a chance where what changes is a thing. Transforms never
// instantiate as actions and never take a habit slot.
//
// From is what changes and To what it becomes, nil when the thing is simply
// gone. In is whose keeping it is in - the market's shelves, a person's own
// pack - and nil for a site, which is in nobody's. Unless names a site that
// arrests the change. Kept means somebody living to keep it arrests it,
// which is what makes ruin the fate of what is left behind rather than of
// everything. Says is the sentence for the settlement's record, empty for
// what goes unremarked.
//
// Two things this table cannot say, and it is better to say why than to
// pretend otherwise. Rates cannot sit on leaf classes: a pack and a shelf
// keep one number for everything edible, so what befalls a provision has to
// be said of provisions, and berries cannot spoil faster than grain. And To
// earns little: razing a tile already knows a field goes back to grass, so
// To is here to say what happens, not to make it happen.
type Transform struct {
	From, To *Class
	In       *Class
	Rate     float64
	Unless   *Class
	Kept     bool
	Says     string
}

// Transforms lists what happens on its own.
//
// The keeping rates are what the same food loses in the two places it sits.
// On the market's shelves it sits out whole seasons; in a pack it is eaten
// within days, which is why a pack loses a fifth of what a shelf does. The
// market's rates were tried on packs and were far too much - a third again
// on top of what an agent eats, on the loop the whole economy runs on - and
// over 24 seeds to 6000 ticks that took the median settlement from 94 to 30
// and killed two. The point was never to punish a full larder in June, only
// to make one in January worth more. What a granary stops and what the cold
// stops on top of it are modifiers over these; see system.Keeping.
//
// Standing is not the same as being kept. A house stands because somebody
// lives in it and a field is a field because somebody works it, so when the
// person is gone what they made starts going too. Without this a settlement
// only ever accumulates: every roof its founders raised is still standing
// three generations later, in everybody's way, and the land under it can
// never be put to anything else. Ruin is what gives a settlement the room to
// be something other than what it first was. About three years either way:
// long enough that a house outlives its builder and an heir could take it
// on, short enough that a settlement is not walled in by its dead. What was claimed and then neither lived in nor sown goes at once,
// because there is nothing there to fall down - only a claim, and it lapses.
//
// A road is Kept, but not by anybody: what keeps a road is feet. It is the
// one thing here whose keeping is not a person, which is why system.Upkeep
// has to ask a different question of it - see world.Walked. Nobody owns a
// road and nobody's dying takes it, but a way nobody walks is a way the grass
// is closing over, and six years of that is enough. Without this roads were
// the only built thing exempt from ruin, and so the only thing a settlement
// could not stop having: every length ever laid was still there, and since a
// road is the cheapest going on the map, the traffic that funnelled along it
// made the case for the next one. Twenty thousand ticks in, settlements had
// paved a third of the map and were down to a house or two.
//
// The pace is slow on purpose - twice a house's, because a roof falls in and
// a road only has grass grow over it - and it is the second of two delays. A
// road is laid on ground already worn past action.wornEnough and keeps that
// wear, so before it is even unkept the wear has to fade from sixty to five
// with nobody at all coming that way, which is about five hundred days, a
// season and a half either side of a year. Only then does this rate start
// running. A street that goes quiet for a winter is in no danger; one that
// has been out of everybody's way since the last generation is.
//
// A public work is nobody's to keep, which is why the granary is not Kept:
// no one person's dying takes it and no one person's living saves it. It
// stands thirty years against a house's three, and the settlement puts
// up another if it still wants one - which it does, because a granary is
// cheap in standing and there is no limit on how many may go up. Without
// this it stood forever, since nothing ever owns one: the raising sets a
// structure on the tile and no owner, so ruin never looked at it. A granary
// that cannot fall is also one that only ever accumulates, and five of them
// leave a hundredth of the usual spoilage - a settlement whose food does
// not go off.
//
// The tavern was tried here once and taken out again, because a settlement
// then got only one - Room refused a second - so a tavern that went was one
// that might never come back. Left to decay it took the meeting place with
// it for good and feuds stopped forming anywhere: the social life thinned
// out rather than turned over. Ruin is only worth having where the thing
// ruined can be built again, and now that a settlement may hold as many
// taverns as it has room for, it can be.
var Transforms = []Transform{
	{From: Provision, In: Market, Rate: 0.01, Unless: Granary},
	{From: Meal, In: Market, Rate: 0.003, Unless: Granary},
	{From: Timber, In: Market, Rate: 0.001},
	{From: Tool, In: Market, Rate: 0.0005},
	{From: Stone, In: Market, Rate: 0},

	{From: Provision, In: Person, Rate: 0.002},
	{From: Meal, In: Person, Rate: 0.002},

	// A roof over somebody's head is in their keeping the way the food in
	// their pack is, and it wears at about the same pace. This is the roof
	// as shelter - the condition of it, which a person makes good by
	// building again - and not the house as a tile, which is the Dwelling
	// below and goes only when there is nobody left to keep it. The two are
	// the same thing at different grains: what rots daily and what falls in
	// once. Nobody is fully roofed all the time, which is why the cold is a
	// cost everyone pays a little of.
	{From: Dwelling, In: Person, Rate: 0.002},

	{From: Granary, To: Open, Rate: 1.0 / (30 * clock.Year), Says: "a granary fell in"},
	{From: Tavern, To: Open, Rate: 1.0 / (20 * clock.Year), Says: "the tavern fell empty"},

	{From: Dwelling, To: Open, Rate: 1.0 / (3 * clock.Year), Kept: true, Says: "an empty house fell in"},
	{From: Road, To: Open, Rate: 1.0 / (6 * clock.Year), Kept: true, Says: "a road nobody walked grew over"},
	{From: Field, To: Open, Rate: 1.0 / (3 * clock.Year), Kept: true},
	{From: Site, To: Open, Rate: 1, Kept: true},
}

// Befalling is what becomes of a thing of class c standing in the world.
// ownerGone says whether whoever kept it is dead, which is what makes the
// Kept transforms apply at all.
//
// The walk is most specific first, so a dwelling goes at a dwelling's pace
// and everything else at the pace of what it is more generally - and a
// granary goes at a granary's pace whether or not anyone owned it, because
// its own row is found before the catch-all that would lapse it at once.
func Befalling(c *Class, ownerGone bool) *Transform {
	for x := c; x != nil; x = x.Parent {
		for i := range Transforms {
			t := &Transforms[i]
			if t.In != nil || t.From != x {
				continue
			}
			if t.Kept && !ownerGone {
				continue
			}
			return t
		}
	}
	return nil
}

// Unkept is what becomes of a thing of class c that nobody is left to keep.
func Unkept(c *Class) *Transform { return Befalling(c, true) }

// Spoiling is what a thing of class c loses per tick while it is kept in a
// holder of class in. ok is false where nothing befalls it there.
func Spoiling(c, in *Class) (*Transform, bool) {
	for x := c; x != nil; x = x.Parent {
		for i := range Transforms {
			if t := &Transforms[i]; !t.Kept && t.In == in && t.From == x {
				return t, true
			}
		}
	}
	return nil, false
}

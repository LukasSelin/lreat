package action

import (
	"math"
	"os"
	"slices"
	"testing"

	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// TestMain runs the action tests with founder idiosyncrasy switched off.
// These tests are about what the shared priors mean - which moment each act
// belongs to - and a founder's own drift on top of them is exactly the noise
// that question does not want. What the drift itself does is tested in
// TestFoundersDifferFromOneAnother.
func TestMain(m *testing.M) {
	BornNoise = 0
	os.Exit(m.Run())
}

// position is where d ranks for a, or -1 if it is not a candidate.
func position(r []Candidate, d *Def) int {
	for i, c := range r {
		if c.Def == d {
			return i
		}
	}
	return -1
}

// fitOf is the fit of d among ranked candidates, or -2 if it is not one.
func fitOf(r []Candidate, d *Def) float64 {
	if i := position(r, d); i >= 0 {
		return r[i].Fit
	}
	return -2
}

func names(r []Candidate) []string {
	out := make([]string, len(r))
	for i, c := range r {
		out[i] = c.Def.Name
	}
	return out
}

func TestImprintCopiesPriorsOnce(t *testing.T) {
	w := world.New(1)
	a := blank(w, "a")
	Imprint(a)
	for i, d := range Catalog {
		if a.Habits[i] != d.Prior || a.Reach[i] != d.Reach0 {
			t.Fatalf("%s not imprinted", d.Name)
		}
	}
	a.Habits[Index(Eat)][habit.Hunger] = 0
	Imprint(a)
	if a.Habits[Index(Eat)][habit.Hunger] != 0 {
		t.Fatal("imprint overwrote habits an agent already had")
	}
}

// Founders are seeded from the same table, so without a little idiosyncrasy
// on top of it every one of them reads every moment identically and the
// settlement is twenty copies of one person: they forage together, build
// together, and starve together. The drift is fixed at birth and is the
// only thing that ever moves a habit off its prior for a founder.
func TestFoundersDifferFromOneAnother(t *testing.T) {
	BornNoise = 0.15
	defer func() { BornNoise = 0 }()
	w := world.New(1)
	a, b := blank(w, "a"), blank(w, "b")
	Imprint(a)
	Imprint(b)
	if slices.Equal(a.Habits, b.Habits) {
		t.Fatal("two founders were imprinted identically")
	}
	// How far the drift carries is a distribution and not a number, so it is
	// asked of a crowd and not of one founder. Over two hundred of them the
	// cosine to the prior runs from 0.66 to 0.96 with a median of 0.89 and a
	// mean of 0.886 - so a bar of 0.9 on a single draw, which is what this
	// asked for, is a coin toss that had been coming up heads. It came up
	// tails the first time the world's own draw shifted under it, which any
	// change to the ground does, because the ground is drawn from the same
	// stream before any founder is.
	eat := Index(Eat)
	var sum float64
	const founders = 200
	for i := 0; i < founders; i++ {
		f := blank(w, "f")
		Imprint(f)
		sum += habit.Cosine(f.Habits[eat], Eat.Prior)
	}
	if mean := sum / founders; mean < 0.85 {
		t.Fatalf("founders' habits wandered off the prior: mean cosine %v over %d of them", mean, founders)
	}
	before := slices.Clone(a.Habits)
	Imprint(a)
	if !slices.Equal(a.Habits, before) {
		t.Fatal("imprint drew fresh idiosyncrasy for an agent that had some")
	}
}

func TestEveryPriorHasADirection(t *testing.T) {
	for _, d := range Catalog {
		if habit.Norm(d.Prior) < habit.MinNorm {
			t.Errorf("%s prior is too faint to recognise anything", d.Name)
		}
		if d.Reach0 <= 0 || d.Reach0 > 1 {
			t.Errorf("%s reach0 = %v", d.Name, d.Reach0)
		}
	}
}

func TestSharedReflectsValuesAndStock(t *testing.T) {
	w := world.New(2)
	a := blank(w, "a")
	a.Norms[belief.Honesty] = 1
	a.Caution = 0
	a.Inventory[entity.Food] = knee(ontology.Provision) // a larder that reads as plenty
	s := Shared(a, w)
	if s[habit.Honesty] != 1 || s[habit.Charity] != 0.5 || s[habit.Caution] != 0 {
		t.Fatalf("shared = %v", s)
	}
	if s[habit.Company] != -1 {
		t.Fatal("alone should read as no company")
	}
	blank(w, "b")
	if Shared(a, w)[habit.Company] != 1 {
		t.Fatal("a neighbour should read as company")
	}
}

// bareCountry empties the whole map of everything a taking could be about:
// no timber to fetch, no berries to pick, no game, no fish. It is for the
// tests that are about the agent rather than about the land - what is
// standing within reach decides how well gathering and foraging fit, and a
// ranking taken beside a full wood is a ranking of the wood.
//
// The whole map and not a radius, because the search that finds a wood goes
// forty tiles and a wood at the edge of that still outranked what was being
// tested. The woods are taken away rather than emptied, because an empty wood
// is still somewhere to go and look: a taking asks the ground for at least
// nothing of what it wants, so a picked-over wood goes on offering itself.
//
// It is here because the ground has now been remade twice under tests that
// never meant to say anything about it, and both times what moved was which
// errand happened to have a good tile within reach.
func bareCountry(w *world.World) {
	for i := range w.Grid.Tiles {
		t := &w.Grid.Tiles[i]
		w.Grid.Wood[i], w.Grid.Wild[i], w.Grid.Fish[i] = 0, 0, 0
		if t.Terrain == world.Forest {
			w.Grid.Turn(entity.Pos{X: i % w.Grid.W, Y: i / w.Grid.W}, world.Grass)
		}
	}
}

func TestStarvingWithFoodEats(t *testing.T) {
	w := world.New(3)
	a := blank(w, "a")
	// With a wood at the door, going to pick something is nearly as good a
	// moment as eating what one has, and on two seeds in ten it is better.
	// That is a fact about foraging and not about eating; what is asked here
	// is that an agent with food in its pack and nothing else on offer eats
	// it rather than setting out.
	bareCountry(w)
	a.Needs[need.Physiological] = 0.05
	a.Inventory[entity.Food] = 2
	r := Rank(a, w)
	if r[0].Def != Eat {
		t.Fatalf("starving agent with food ranked %v", names(r))
	}
}

func TestStarvingWithoutFoodGoesLookingForIt(t *testing.T) {
	w := world.New(4)
	a := blank(w, "a")
	a.Needs[need.Physiological] = 0.05
	a.Inventory[entity.Food] = 0
	r := Rank(a, w)
	if r[0].Def != Forage && r[0].Def != Farm {
		t.Fatalf("starving agent without food ranked %v", names(r))
	}
}

func TestDishonestyMakesTheftFitBetter(t *testing.T) {
	rank := func(honesty, caution float64) int {
		w := world.New(5)
		thief := blank(w, "thief")
		victim := blank(w, "victim")
		victim.Pos = entity.Pos{X: thief.Pos.X + 1, Y: thief.Pos.Y}
		victim.Inventory[entity.Food] = 3
		thief.Needs[need.Physiological] = 0.05
		thief.Inventory[entity.Food] = 0
		thief.Norms[belief.Honesty] = honesty
		thief.Caution = caution
		return position(Rank(thief, w), Steal)
	}
	crook, saint := rank(0, 0), rank(1, 1)
	if crook < 0 || saint < 0 {
		t.Fatal("steal was not a candidate")
	}
	if !(crook < saint) {
		t.Fatalf("steal ranks %d for the crook and %d for the saint", crook, saint)
	}
}

func TestSatedCuriousAgentStudiesWhenItIsInReach(t *testing.T) {
	// Asked of sixty worlds rather than of one. What an agent ranks depends on
	// the habits it was imprinted with, and those are drawn from the world's
	// own stream after the ground has been drawn from it - so a single world
	// seed is a single draw, and pinned to one of them this says nothing about
	// the behaviour and everything about which seed was picked.
	//
	// The bar is well under the rate and not at it. Over sixty worlds the
	// curious moment calls for study on sixty-eight of a hundred, and the
	// first attempt at fixing this counted twenty worlds, saw seventeen, and
	// asked for fifteen - which is above the rate, so it failed on the very
	// next change to the ground. A bar set at what was measured is a bar that
	// fails half the time by construction.
	held := 0
	const worlds = 60
	for seed := uint64(1); seed <= worlds; seed++ {
		w := world.New(seed)
		a := blank(w, "a")
		a.Needs = need.Levels{0.95, 0.95, 0.9, 0.9, 0.1}
		a.Inventory[entity.Food] = 3
		a.Shelter = 1
		Imprint(a)
		a.Reach[Index(Study)] = 1
		r := Rank(a, w)
		// Rest is what an agent with nothing pressing does, and a moment with
		// one need in it is still a quiet one; what is asked here is that of
		// the acts that do something, the curious moment calls for study.
		if r[0].Def == Rest && r[1].Def == Study {
			held++
		}
	}
	if held < 30 {
		t.Fatalf("the curious moment called for study on %d worlds of %d", held, worlds)
	}
}

func TestReachHoldsAnActionBack(t *testing.T) {
	w := world.New(7)
	a := blank(w, "a")
	a.Needs = need.Levels{0.95, 0.95, 0.9, 0.9, 0.1}
	a.Inventory[entity.Food] = 3
	Imprint(a)
	a.Reach[Index(Study)] = 0
	far := position(Rank(a, w), Study)
	a.Reach[Index(Study)] = 1
	near := position(Rank(a, w), Study)
	if !(near < far) {
		t.Fatalf("study ranks %d out of reach and %d in reach", far, near)
	}
}

func TestColdAgentWithWoodBuilds(t *testing.T) {
	w := world.New(8)
	a := blank(w, "a")
	a.Needs = need.Levels{0.9, 0.05, 0.8, 0.8, 0.8}
	a.Inventory[entity.Food] = 3
	a.Inventory[entity.Wood] = raisingTimber
	a.Shelter = 0
	r := Rank(a, w)
	if r[0].Def != BuildShelter {
		t.Fatalf("cold agent with wood ranked %v", names(r))
	}
}

func TestLonelyAgentWithCompanySocializes(t *testing.T) {
	w := world.New(9)
	a := blank(w, "a")
	o := blank(w, "o")
	o.Pos = entity.Pos{X: a.Pos.X + 2, Y: a.Pos.Y}
	a.Needs = need.Levels{0.9, 0.9, 0.05, 0.8, 0.8}
	a.Inventory[entity.Food] = 3
	a.Shelter = 1
	// A settlement that keeps some order. Order is read as a lack, so a
	// place where nobody has ever kept any calls on everybody to stand a
	// watch, whatever else they want, and that is the point of reading it
	// that way. It is not the moment this test is about.
	w.Safety = 0.5
	r := Rank(a, w)
	if r[0].Def != Socialize {
		t.Fatalf("lonely agent ranked %v", names(r))
	}
}

func TestSituationIsPerCandidate(t *testing.T) {
	w := world.New(10)
	a := blank(w, "a")
	a.Inventory[entity.Food] = 2
	for _, c := range Candidates(a, w) {
		if c.Def == Eat && c.Situation[habit.Near] != 1 {
			t.Fatalf("eating here should be as near as can be, got %v", c.Situation[habit.Near])
		}
		if math.IsNaN(c.Fit) {
			t.Fatalf("%s fit is NaN", c.Def.Name)
		}
	}
}

// Twins of the value-rule tests in core/system, asserting order under
// recognition rather than the choice the old rule made.

func TestLonelyButHungryAgentStillEatsByFit(t *testing.T) {
	w := world.New(11)
	a := blank(w, "a")
	blank(w, "o")
	a.Needs = need.Levels{0.1, 0.8, 0.0, 0.8, 0.8}
	a.Inventory[entity.Food] = 2
	// The claim is about hunger against loneliness, not about hunger against
	// every way of answering it: two units of food is nearly out of food on
	// a twelve-unit reading, so going to the woods fits such a moment too.
	// What must not happen is that company comes first.
	r := Rank(a, w)
	if position(r, Eat) > position(r, Socialize) || position(r, Eat) > 1 {
		t.Fatalf("hungry lonely agent ranked %v", names(r))
	}
}

func TestCautionMakesTheftFitWorse(t *testing.T) {
	fit := func(caution float64) float64 {
		w := world.New(12)
		thief := blank(w, "thief")
		victim := blank(w, "victim")
		victim.Pos = entity.Pos{X: thief.Pos.X + 1, Y: thief.Pos.Y}
		victim.Inventory[entity.Food] = 3
		thief.Needs[need.Physiological] = 0.05
		thief.Inventory[entity.Food] = 0
		thief.Norms[belief.Honesty] = 0
		thief.Caution = caution
		return fitOf(Rank(thief, w), Steal)
	}
	// Read as a fit rather than a place in the order: caution is one
	// coordinate among twenty and moves the fit without always moving the
	// rank past the act above it.
	lawless, policed := fit(0), fit(1)
	if !(lawless > policed) {
		t.Fatalf("theft fits as well where it is answered (%v) as where it is not (%v)", policed, lawless)
	}
}

func TestTheWrongedRecogniseAMomentToGetEven(t *testing.T) {
	setup := func(charity float64) (*world.World, *entity.Agent) {
		w := world.New(13)
		victim := blank(w, "victim")
		thief := blank(w, "thief")
		thief.Pos = victim.Pos
		for _, a := range []*entity.Agent{victim, thief} {
			a.Needs = need.Levels{1, 0.9, 0.9, 0.5, 0.9}
			a.Inventory[entity.Food] = 3
			a.Shelter = 1
		}
		victim.Norms[belief.Charity] = charity
		victim.Judge(thief.ID, -0.6, w.Tick)
		// Nothing to fetch anywhere, so that where getting even lands is
		// decided by the grievance and not by whether the pair happen to
		// have stopped beside a wood.
		bareCountry(w)
		return w, victim
	}
	w, victim := setup(0)
	r := Rank(victim, w)
	vengeful := position(r, Retaliate)
	// A settled moment belongs to rest, and an unpoliced settlement calls
	// on anyone to stand a watch, so getting even is not the whole of what
	// a wronged and otherwise comfortable agent recognises. What is asked
	// is that it is near the front, and - below - that it is nearer the
	// front for the wronged than for the charitable.
	if vengeful < 0 || vengeful > 2 {
		t.Fatalf("a wronged agent with nothing better to do ranked %v", names(r))
	}
	w, saint := setup(1)
	rs := Rank(saint, w)
	// Charity is read on the same scale for every candidate, so it shows in
	// how well getting even fits rather than always in where it lands.
	if !(fitOf(rs, Retaliate) < fitOf(r, Retaliate)) {
		t.Fatalf("getting even fits the charitable (%v) as well as the wronged (%v)",
			fitOf(rs, Retaliate), fitOf(r, Retaliate))
	}
}

// The cold is one thing and standing out in it is another. A roofed body and
// an unroofed one read the same weather, so a coordinate that carries only
// the weather says the same to both, and every act that names it is an act
// the whole settlement turns to at once. Exposure is the product the world
// actually charges a body for the season, and it is what lets the same
// prior send the unroofed to the woods and leave the roofed in the field.
func TestExposureTellsTheRoofedFromTheUnroofed(t *testing.T) {
	w := world.New(3)
	for w.Climate.Chill() < 0.5 {
		w.Climate.Advance(w.Tick, w.RNG)
		w.Tick++
	}
	roofed, cold := blank(w, "roofed"), blank(w, "cold")
	roofed.Shelter, cold.Shelter = 1, 0
	// Standing on the same ground, so that the only difference between them
	// is the roof. The weather is read where a body is - the same day is
	// colder higher up; see world.Lapse - so two agents on two tiles need
	// not read the same Chill, and that is not what this is about.
	cold.Pos = roofed.Pos

	warm := Shared(roofed, w)
	bare := Shared(cold, w)
	if warm[habit.Chill] != bare[habit.Chill] {
		t.Fatal("the weather is the same news to everybody standing in the same place")
	}
	if !(bare[habit.Exposure] > warm[habit.Exposure]) {
		t.Fatalf("the unroofed should feel the winter more: %v against %v",
			bare[habit.Exposure], warm[habit.Exposure])
	}
	if warm[habit.Exposure] != 0 {
		t.Fatalf("a roof should take all of it, not %v", warm[habit.Exposure])
	}
	// And in a mild season it says nothing to anybody, roof or no roof.
	w.Climate.Temp = world.Mild + 5
	if s := Shared(cold, w); s[habit.Exposure] != 0 {
		t.Fatalf("there is nothing to be exposed to in mild weather: %v", s[habit.Exposure])
	}
}

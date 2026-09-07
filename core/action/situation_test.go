package action

import (
	"math"
	"testing"

	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/world"
)

// position is where d ranks for a, or -1 if it is not a candidate.
func position(r []Candidate, d *Def) int {
	for i, c := range r {
		if c.Def == d {
			return i
		}
	}
	return -1
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
		t.Fatal("imprint overwrote a learned habit")
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
	a.Inventory[entity.Food] = 4
	s := Shared(a, w)
	if s[habit.Honesty] != 1 || s[habit.Charity] != 0.5 || s[habit.Caution] != 0 || s[habit.Food] != 1 {
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

func TestStarvingWithFoodEats(t *testing.T) {
	w := world.New(3)
	a := blank(w, "a")
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
	w := world.New(6)
	a := blank(w, "a")
	a.Needs = need.Levels{0.95, 0.95, 0.9, 0.9, 0.1}
	a.Inventory[entity.Food] = 3
	a.Shelter = 1
	Imprint(a)
	a.Reach[Index(Study)] = 1
	r := Rank(a, w)
	if r[0].Def != Study {
		t.Fatalf("curious agent ranked %v", names(r))
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
	if r := Rank(a, w); r[0].Def != Eat {
		t.Fatalf("hungry lonely agent ranked %v", names(r))
	}
}

func TestCautionMakesTheftFitWorse(t *testing.T) {
	rank := func(caution float64) int {
		w := world.New(12)
		thief := blank(w, "thief")
		victim := blank(w, "victim")
		victim.Pos = entity.Pos{X: thief.Pos.X + 1, Y: thief.Pos.Y}
		victim.Inventory[entity.Food] = 3
		thief.Needs[need.Physiological] = 0.05
		thief.Inventory[entity.Food] = 0
		thief.Norms[belief.Honesty] = 0
		thief.Caution = caution
		return position(Rank(thief, w), Steal)
	}
	lawless, policed := rank(0), rank(1)
	if !(lawless < policed) {
		t.Fatalf("steal ranks %d where nothing is enforced and %d where it is", lawless, policed)
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
		return w, victim
	}
	w, victim := setup(0)
	r := Rank(victim, w)
	if r[0].Def != Retaliate {
		t.Fatalf("a wronged agent with nothing better to do ranked %v", names(r))
	}
	vengeful := position(r, Retaliate)
	w, saint := setup(1)
	forgiving := position(Rank(saint, w), Retaliate)
	if !(vengeful < forgiving) {
		t.Fatalf("retaliate ranks %d for the wronged and %d for the charitable", vengeful, forgiving)
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

	warm := Shared(roofed, w)
	bare := Shared(cold, w)
	if warm[habit.Chill] != bare[habit.Chill] {
		t.Fatal("the weather is the same news to everybody")
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

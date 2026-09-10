package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/world"
)

// Good ground reads as good ground. The appraisal is a single figure, so the
// only thing worth asserting about it is the order it puts places in.
func TestGroundIsJudgedByWhatIsOnIt(t *testing.T) {
	w, _ := shore(t)
	rich := entity.Pos{X: 3, Y: 4}
	poor := entity.Pos{X: 3, Y: 5}
	w.Grid.At(rich).Fertility, w.Grid.At(rich).Rich = 1, 1
	w.Grid.At(poor).Fertility, w.Grid.At(poor).Rich = 0, 0
	if landWorth(w, rich) <= landWorth(w, poor) {
		t.Fatalf("rich ground at %.1f, poor at %.1f", landWorth(w, rich), landWorth(w, poor))
	}
}

// A way worn across a plot is a way somebody wants to walk, and a house on it
// is a house they walk through.
func TestWornGroundIsWorthLess(t *testing.T) {
	w, _ := shore(t)
	p := entity.Pos{X: 3, Y: 4}
	before := landWorth(w, p)
	for i := 0; i < 200; i++ {
		w.Grid.Tread(p, 0)
	}
	if landWorth(w, p) >= before {
		t.Fatalf("worn ground worth %.1f, was %.1f before anyone walked it", landWorth(w, p), before)
	}
}

// Ground beside the settlement is worth more than ground at the far end of
// the map, and nothing in the code says where the settlement is.
func TestGroundIsWorthMoreWhereThePeopleAre(t *testing.T) {
	w, _ := shore(t)
	crowd := entity.Pos{X: 2, Y: 2}
	for i := 0; i < 5; i++ {
		w.SpawnAt("neighbour", w.RandomPersonality(), crowd)
	}
	far := entity.Pos{X: 11, Y: 5}
	if companyWorth(w, crowd) <= companyWorth(w, far) {
		t.Fatalf("company beside a crowd %.1f, out on its own %.1f",
			companyWorth(w, crowd), companyWorth(w, far))
	}
}

// An agent knows the ground it has stood on and no other. This is the whole
// difference between siting a house and reading a survey.
func TestAnAgentOnlyKnowsWhereItHasBeen(t *testing.T) {
	w, a := shore(t)
	if len(a.Places) != 0 {
		t.Fatalf("a newcomer already knew %d places", len(a.Places))
	}
	a.Pos = entity.Pos{X: 3, Y: 4}
	if !Notice(a, w) {
		t.Fatal("standing on open ground taught the agent nothing")
	}
	if _, ok := a.Knows(a.Pos); !ok {
		t.Fatalf("stood at %v and does not know it", a.Pos)
	}
	if _, ok := a.Knows(entity.Pos{X: 10, Y: 5}); ok {
		t.Fatal("knows ground it has never been near")
	}
}

// Looking at the same ground twice teaches nothing the second time, which is
// what stops a day spent walking in circles from paying.
func TestNoticingTheSameGroundTwiceIsNoDiscovery(t *testing.T) {
	w, a := shore(t)
	a.Pos = entity.Pos{X: 3, Y: 4}
	if !Notice(a, w) {
		t.Fatal("the first look taught nothing")
	}
	if Notice(a, w) {
		t.Fatal("the second look at the same tile counted as a discovery")
	}
}

// A head only holds so much, and what it drops is the worst of it.
func TestAHeadFullOfGroundKeepsTheBest(t *testing.T) {
	w, a := shore(t)
	for i := 0; i < entity.MaxPlaces+4; i++ {
		a.Remember(entity.Pos{X: i, Y: 0}, float64(i), w.Tick)
	}
	if len(a.Places) != entity.MaxPlaces {
		t.Fatalf("carrying %d places, cap is %d", len(a.Places), entity.MaxPlaces)
	}
	for _, p := range a.Places {
		if p.Worth < 4 {
			t.Fatalf("kept a place worth %.0f when better ones were offered", p.Worth)
		}
	}
}

// Between poor ground underfoot and good ground it has walked, an agent
// takes the good ground - and the other way about when the good ground is
// where it stands.
func TestSitingTakesTheBestGroundKnown(t *testing.T) {
	w, _ := shore(t)
	good := entity.Pos{X: 2, Y: 4}
	w.Grid.At(good).Fertility, w.Grid.At(good).Rich = 1, 1
	a := blank(w, "settler")
	a.Pos = entity.Pos{X: 9, Y: 5}
	for _, p := range []entity.Pos{good, {X: 8, Y: 5}} {
		here := a.Pos
		a.Pos = p
		Notice(a, w)
		a.Pos = here
	}
	p, ok := KnownPlot(a, w)
	if !ok {
		t.Fatal("a settler who has seen good ground found nowhere to build")
	}
	if p != good {
		t.Fatalf("sited at %v, want the good ground it had walked over at %v", p, good)
	}
}

// Two settlers who have led different lives disagree about where to build.
// Under the old rule they were handed the same tile, on the same tick, every
// time, and queued for it.
func TestTwoSettlersDoNotPickTheSamePlot(t *testing.T) {
	w, _ := shore(t)
	one, two := blank(w, "one"), blank(w, "two")
	one.Pos, two.Pos = entity.Pos{X: 2, Y: 4}, entity.Pos{X: 10, Y: 1}
	Notice(one, w)
	Notice(two, w)
	p, ok1 := KnownPlot(one, w)
	q, ok2 := KnownPlot(two, w)
	if !ok1 || !ok2 {
		t.Fatal("a settler found nowhere to build")
	}
	if p == q {
		t.Fatalf("both settlers were sent to %v", p)
	}
}

// Somebody with nowhere decent in mind goes to look. Somebody under a roof
// has what looking is for and stays home.
func TestLookingIsForPeopleWithNowhereToLive(t *testing.T) {
	w, _ := shore(t)
	a := blank(w, "settler")
	if !worthLooking(a, w) {
		t.Fatal("a settler who knows nowhere at all had no reason to look")
	}
	a.Remember(entity.Pos{X: 3, Y: 4}, contentedGround+1, w.Tick)
	if worthLooking(a, w) {
		t.Fatal("a settler with good ground in mind still went wandering")
	}
	a.Places = nil
	a.HasHome, a.Home = true, entity.Pos{X: 3, Y: 4}
	if worthLooking(a, w) {
		t.Fatal("a householder went looking for somewhere to live")
	}
}

// A scout heads for country it does not know, and two of them standing in the
// same spot do not walk off in single file.
func TestScoutsHeadForUnknownCountry(t *testing.T) {
	w := world.NewSized(4, 60, 60)
	w.Grid.Layers = world.NewLayers(len(w.Grid.Tiles))
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i] = world.Tile{Terrain: world.Grass}
	}
	a := blank(w, "a")
	a.Pos = entity.Pos{X: 30, Y: 30}
	p, ok := scoutSite(a, w)
	if !ok {
		t.Fatal("a scout in the middle of an open map found nowhere to go")
	}
	if entity.Dist(a.Pos, p) < scoutRange/2 {
		t.Fatalf("scouted %d tiles off, which is no journey", entity.Dist(a.Pos, p))
	}
	var apart bool
	for i := 0; i < 8; i++ {
		b := blank(w, "b")
		b.Pos = a.Pos
		if q, ok := scoutSite(b, w); ok && q != p {
			apart = true
			break
		}
	}
	if !apart {
		t.Fatal("every scout in the party set off the same way")
	}
}

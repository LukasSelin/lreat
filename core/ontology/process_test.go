package ontology

import "testing"

// The stage durations are written out again here rather than referred to,
// so that the numbers cannot be retuned by editing one side. If a stage
// changes, this test is where the change has to be said out loud.
func TestWhatGrowsTakesTheTimeItTook(t *testing.T) {
	for _, c := range []struct {
		p    *Process
		full float64
	}{
		{Crop, 25},        // a season
		{Brush, 600},      // six years
		{Timbering, 2000}, // twenty
	} {
		if got := c.p.Full(); got != c.full {
			t.Errorf("%s comes on in %v, want %v", c.p.Name, got, c.full)
		}
	}
}

// A crop is worth cutting once it is in ear, which is halfway. The share
// has to be exactly a half: it is compared against a tile's age every tick
// for every field, and a hair either way is a different settlement.
func TestACropIsWorthCuttingHalfwayThrough(t *testing.T) {
	if got := InEar.Share(); got != 0.5 {
		t.Fatalf("a crop is in ear at %v of its growing, want exactly 0.5", got)
	}
	if got := Crop.Phase("sown").Share(); got != 0 {
		t.Fatalf("a crop is sown at %v, want the start", got)
	}
}

// What grows and what does not. A wood is making brush and timber at once,
// off the one clock; an outcrop is stone and does not grow, and the water's
// fish come back without having to come on.
func TestOnlyLivingGroundGrows(t *testing.T) {
	for _, c := range Site.Family() {
		living := c.Has(Living)
		if grows := len(Growing(c)) > 0; grows != living {
			t.Errorf("%s: grows=%v but Living=%v", c.Name, grows, living)
		}
	}
	if n := len(Growing(Wood)); n != 2 {
		t.Errorf("a wood runs %d processes, want brush and timber", n)
	}
	for _, p := range Processes {
		if !p.Of.Has(Living) {
			t.Errorf("%s runs on %s, which is not Living", p.Name, p.Of.Name)
		}
		if p.Yields == nil || !Offers(p.Of, p.Yields) {
			t.Errorf("%s yields %v, which %s does not afford", p.Name, p.Yields, p.Of.Name)
		}
	}
}

// The rates are written out again here rather than referred to, for the
// same reason the stage durations are: a number that appears in one place
// can be retuned by accident, and these were measured.
func TestWhatHappensOnItsOwnHappensAtTheRateItDid(t *testing.T) {
	for _, c := range []struct {
		what     string
		from, in *Class
		rate     float64
	}{
		{"food on a shelf", Provision, Market, 0.01},
		{"meals on a shelf", Meal, Market, 0.003},
		{"timber on a shelf", Timber, Market, 0.001},
		{"tools on a shelf", Tool, Market, 0.0005},
		{"stone on a shelf", Stone, Market, 0},
		{"food in a pack", Provision, Person, 0.002},
		{"meals in a pack", Meal, Person, 0.002},
		// Berries are perishable and grain is not, but both are food in a
		// pack and on a shelf, so both go at a provision's rate. It is the
		// binding that is coarse, not the trees.
		{"berries in a pack", Berries, Person, 0.002},
		{"grain in a pack", Grain, Person, 0.002},
	} {
		tr, ok := Spoiling(c.from, c.in)
		if !ok {
			t.Errorf("%s: nothing befalls it", c.what)
			continue
		}
		if tr.Rate != c.rate {
			t.Errorf("%s goes at %v, want %v", c.what, tr.Rate, c.rate)
		}
	}
	// A pack keeps what a shelf does not, because what is carried is eaten
	// within days while a shelf's stock sits out whole seasons.
	shelf, _ := Spoiling(Provision, Market)
	pack, _ := Spoiling(Provision, Person)
	if pack.Rate*5 != shelf.Rate {
		t.Errorf("a pack loses %v against a shelf's %v, want a fifth", pack.Rate, shelf.Rate)
	}
}

// What nobody is left to keep, and how fast it goes. A dwelling and a field
// go at their own pace; anything else claimed and neither lived in nor sown
// goes at once, because there is nothing there to fall down.
func TestWhatNobodyKeepsGoesAtItsOwnPace(t *testing.T) {
	for _, c := range []struct {
		of   *Class
		rate float64
	}{
		{Dwelling, 1.0 / 300},
		{Field, 1.0 / 300},
		{Open, 1},
		{Road, 1},
		{Wood, 1},
	} {
		tr := Unkept(c.of)
		if tr == nil {
			t.Errorf("%s left unkept becomes nothing at all", c.of.Name)
			continue
		}
		if tr.Rate != c.rate {
			t.Errorf("an unkept %s goes at %v, want %v", c.of.Name, tr.Rate, c.rate)
		}
	}
	// A certainty is drawn for by nobody: see system.wither, which spends no
	// luck on a rate of one, and would be a different settlement if it did.
	if Unkept(Open).Rate != 1 {
		t.Fatal("a lapsed claim is not certain to lapse")
	}
}

// A public work is nobody's to keep. No one person's dying takes it and no
// one person's living saves it, so it falls in on its own long clock -
// which is also the only reason a granary is not built without end.
func TestAPublicWorkFallsInWhoeverIsAlive(t *testing.T) {
	for _, c := range []struct {
		of   *Class
		rate float64
	}{
		{Granary, 1.0 / 3000},
		{Tavern, 1.0 / 2000},
	} {
		for _, ownerGone := range []bool{false, true} {
			tr := Befalling(c.of, ownerGone)
			if tr == nil {
				t.Fatalf("a %s stands forever (owner gone: %v)", c.of.Name, ownerGone)
			}
			if tr.Rate != c.rate {
				t.Errorf("a %s goes at %v (owner gone: %v), want %v", c.of.Name, tr.Rate, ownerGone, c.rate)
			}
		}
		// It outlasts a house by a long way, or the settlement spends its
		// whole life rebuilding what it shares.
		if !(Befalling(c.of, false).Rate < Unkept(Dwelling).Rate/5) {
			t.Errorf("a %s goes too near the pace of a house", c.of.Name)
		}
	}
	// Ordinary ground that somebody is alive to hold is not going anywhere,
	// and must cost no draw: see system.wither.
	for _, c := range []*Class{Open, Wood, Field, Dwelling, Road, Water} {
		if tr := Befalling(c, false); tr != nil {
			t.Errorf("a kept %s decays at %v with its holder alive", c.Name, tr.Rate)
		}
	}
}

// Only what is worth remarking on speaks up when it goes. The event log is
// bounded, and what the settlement's watchers read is in it.
func TestOnlyBuildingsSpeakUpWhenTheyGo(t *testing.T) {
	for _, tr := range Transforms {
		if tr.Says != "" && !tr.From.IsA(Built) {
			t.Errorf("%s speaks up when it goes", tr.From.Name)
		}
	}
}

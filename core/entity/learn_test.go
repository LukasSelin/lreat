package entity

import (
	"math"
	"testing"
)

// near compares levels the way the world does, which is not to the last bit:
// a ceiling of teacher-TaughtGap is a subtraction of two floats and lands a
// hair off the number anybody would write down.
func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// reps counts how many acts of size d it takes to get from level to want,
// which is the only honest way to read a curve: the numbers in the rate
// table mean nothing beside each other, and everything beside the hours.
func reps(d, from, want float64, learn func(a *Agent) float64) int {
	a := &Agent{}
	a.Skills[Farming] = from
	for n := 1; n <= 100000; n++ {
		if learn(a) == 0 {
			return -1
		}
		if a.Skills[Farming] >= want {
			return n
		}
	}
	return -1
}

func doing(a *Agent) float64 { return a.Learn(Farming, 0.01) }

func TestEachTierCostsMoreThanTheOneBelow(t *testing.T) {
	var last int
	for tier := Novice; tier < TierCount; tier++ {
		from := float64(tier) * TierBand
		n := reps(0.01, from, from+TierBand, doing)
		if n < 0 {
			t.Fatalf("%s never crossed its tier by working", tier)
		}
		if tier > Novice && n <= last {
			t.Fatalf("%s took %d acts to cross and the tier under it took %d; each tier should cost more than the one below", tier, n, last)
		}
		last = n
	}
}

func TestTheFirstTierIsAsQuickAsItEverWas(t *testing.T) {
	// The opening years of a settlement must not have got harder. A novice
	// learns at the flat old rate, which is what the rate table's first
	// entry says, so twenty-five acts of a hundredth cross the tier.
	if n := reps(0.01, 0, TierBand, doing); n != 25 {
		t.Fatalf("a novice took %d acts to leave the tier, want the 25 a flat rate took", n)
	}
}

func TestWorkingIsTheOnlyWayToMastery(t *testing.T) {
	a := &Agent{}
	for i := 0; i < 100000 && a.Skills[Farming] < Mastery; i++ {
		a.Learn(Farming, 0.01)
	}
	if a.Skills[Farming] != Mastery {
		t.Fatalf("a lifetime of the work came to %.2f, want mastery", a.Skills[Farming])
	}
}

func TestBeingShownStopsShortOfTheTeacher(t *testing.T) {
	pupil := &Agent{}
	for i := 0; i < 100000; i++ {
		if pupil.Taught(Farming, TeachRateForTest, 0.6) == 0 {
			break
		}
	}
	if want := 0.6 - TaughtGap; !near(pupil.Skills[Farming], want) {
		t.Fatalf("a journeyman's pupil came to %.2f, want %.2f - the last of it is not the teacher's to give", pupil.Skills[Farming], want)
	}
}

func TestEvenAMastersPupilIsOnlyAJourneyman(t *testing.T) {
	pupil := &Agent{}
	for i := 0; i < 100000; i++ {
		if pupil.Taught(Farming, TeachRateForTest, Mastery) == 0 {
			break
		}
	}
	if !near(pupil.Skills[Farming], TaughtCeiling) {
		t.Fatalf("a master's pupil came to %.2f, want the taught ceiling %.2f", pupil.Skills[Farming], TaughtCeiling)
	}
	// The ceiling is the threshold of mastery and not a step inside it: from
	// here the pupil is on master-tier rates and has only the work left.
	if got := pupil.Taught(Farming, TeachRateForTest, Mastery); got != 0 {
		t.Fatalf("a pupil at the ceiling was taught a further %.4f, want nothing", got)
	}
	if TierOf(pupil.Skills[Farming]) != Master {
		t.Fatal("the taught ceiling should sit exactly on the threshold of the master tier")
	}
}

func TestAMasterTeachesFasterThanAnAmateur(t *testing.T) {
	master := reps(0, 0, TierBand, func(a *Agent) float64 { return a.Taught(Farming, TeachRateForTest, Mastery) })
	amateur := reps(0, 0, TierBand, func(a *Agent) float64 { return a.Taught(Farming, TeachRateForTest, 0.5) })
	if master < 0 || amateur < 0 {
		t.Fatal("both should be able to carry a beginner out of the novice tier")
	}
	if amateur <= master {
		t.Fatalf("an amateur took %d lessons and a master %d; the master should be the quicker", amateur, master)
	}
}

func TestTwoAmateursHaveNothingToPass(t *testing.T) {
	// The pair the old flat step rewarded most: two people at the same
	// level, each raising the other, neither having anything to show.
	one, two := &Agent{}, &Agent{}
	one.Skills[Farming], two.Skills[Farming] = 0.5, 0.5
	if one.CanBeTaught(Farming, two.Skills[Farming]) {
		t.Fatal("two equals should have nothing to teach each other")
	}
	if got := one.Taught(Farming, TeachRateForTest, two.Skills[Farming]); got != 0 {
		t.Fatalf("an equal passed on %.4f, want nothing", got)
	}
}

func TestNobodyBelowTheFloorTeachesAnything(t *testing.T) {
	pupil := &Agent{}
	if pupil.CanBeTaught(Farming, TeachFloor-0.01) {
		t.Fatal("somebody still in the novice tier has nothing to show")
	}
	if got := pupil.Taught(Farming, TeachRateForTest, TeachFloor-0.01); got != 0 {
		t.Fatalf("a novice passed on %.4f, want nothing", got)
	}
}

func TestReadingStopsAtTheDoorOfTheWork(t *testing.T) {
	a := &Agent{}
	for i := 0; i < 100000; i++ {
		if a.Study(Scholarship, 0.02) == 0 {
			break
		}
	}
	if !near(a.Skills[Scholarship], StudyCeiling) {
		t.Fatalf("a life at a desk came to %.2f, want the study ceiling %.2f", a.Skills[Scholarship], StudyCeiling)
	}
}

func TestTierOfCoversTheWholeRange(t *testing.T) {
	for _, c := range []struct {
		level float64
		want  Tier
	}{{0, Novice}, {0.24, Novice}, {0.25, Apprentice}, {0.5, Journeyman}, {0.75, Master}, {Mastery, Master}} {
		if got := TierOf(c.level); got != c.want {
			t.Fatalf("%.2f reads as %s, want %s", c.level, got, c.want)
		}
	}
}

// TeachRateForTest is action.TeachRate, which package entity cannot see.
const TeachRateForTest = 0.04

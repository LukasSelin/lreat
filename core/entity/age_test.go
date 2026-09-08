package entity

import (
	"testing"

	"lreat/core/clock"
)

func TestAgeFactorRisesThenFalls(t *testing.T) {
	child, adult, elder := AgeFactor(0), AgeFactor(Prime-1), AgeFactor(Lifespan)
	if child >= adult {
		t.Fatalf("a newborn body (%.2f) should be weaker than an adult one (%.2f)", child, adult)
	}
	if adult != 1 {
		t.Fatalf("an adult in its prime brings %.2f of its vitality, want all of it", adult)
	}
	if elder >= adult {
		t.Fatalf("a body at the end of life (%.2f) should be weaker than one in its prime (%.2f)", elder, adult)
	}
	if AgeFactor(Lifespan*3) < 0.25 {
		t.Fatal("decline should bottom out rather than reach nothing")
	}
	for age := 0; age < Lifespan*2; age += 50 {
		if f := AgeFactor(age); f <= 0 || f > 1 {
			t.Fatalf("age %d gives factor %.2f, want it within (0,1]", age, f)
		}
	}
}

func TestFrailtyStartsAtPrimeAndClimbs(t *testing.T) {
	if Frailty(Prime-1) != 0 {
		t.Fatal("bodies should not fail of old age before their prime is over")
	}
	if Frailty(Lifespan) <= Frailty(Prime+100) {
		t.Fatal("the risk of dying old should climb with the years")
	}
	// Half a lifetime of decline should carry most agents off.
	survival := 1.0
	for age := Prime; age < Lifespan; age++ {
		survival *= 1 - Frailty(age)
	}
	if survival > 0.5 {
		t.Fatalf("%.0f%% of agents reach Lifespan, want old age to be the usual end", survival*100)
	}
}

func TestOnlyGrownAgentsBear(t *testing.T) {
	if Fertile(Maturity/2) || Adult(Maturity/2) {
		t.Fatal("a child should be neither grown nor fertile")
	}
	if !Fertile(Maturity) || !Adult(Maturity) {
		t.Fatal("an agent that has reached maturity should be both")
	}
	if Fertile(Prime) {
		t.Fatal("bearing should end when decline begins")
	}
	if !Adult(Lifespan) {
		t.Fatal("an elder is still a grown agent")
	}
}

// A life is human, and a settlement is meant to feel it. A childhood that is
// a rounding error on a lifetime is a settlement with no dependants in it -
// which is what these spans were while the year was a hundred ticks long and
// a life had to be counted in ticks to fit a run inside it.
func TestALifeIsHuman(t *testing.T) {
	if got := clock.Years(Maturity); got < 12 || got > 18 {
		t.Errorf("a childhood lasts %d years, want a human one", got)
	}
	if got := clock.Years(Lifespan); got < 55 || got > 80 {
		t.Errorf("a life lasts %d years, want a human one", got)
	}
	if got := clock.Years(Prime - Maturity); got < 20 {
		t.Errorf("the bearing years are %d long, want a generation's worth", got)
	}
	// The dependants are the cost of the childhood and the reason to have
	// one: a settlement that cannot feed its children does not grow.
	if share := float64(Maturity) / float64(Lifespan); share < 0.15 {
		t.Errorf("children are %.2f of a life, want a settlement that carries some", share)
	}
}

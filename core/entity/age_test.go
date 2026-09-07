package entity

import "testing"

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

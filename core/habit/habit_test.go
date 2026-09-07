package habit

import (
	"math"
	"math/rand/v2"
	"testing"
)

func TestCosineIsBoundedAndNeverNaN(t *testing.T) {
	var zero Signature
	a := Signature{Hunger: 1, Food: -1}
	if c := Cosine(a, zero); c != 0 {
		t.Fatalf("cosine with zero vector = %v, want 0", c)
	}
	if c := Cosine(a, a); math.Abs(c-1) > 1e-12 {
		t.Fatalf("self cosine = %v, want 1", c)
	}
	b := a
	for i := range b {
		b[i] = -b[i]
	}
	if c := Cosine(a, b); math.Abs(c+1) > 1e-12 {
		t.Fatalf("opposite cosine = %v, want -1", c)
	}
	if math.IsNaN(Cosine(zero, zero)) {
		t.Fatal("cosine of zeros is NaN")
	}
}

func TestUnitOfZeroIsZero(t *testing.T) {
	if u := Unit(Signature{}); u != (Signature{}) {
		t.Fatalf("unit of zero = %v", u)
	}
	if n := Norm(Unit(Signature{Near: 3, Skill: -4})); math.Abs(n-1) > 1e-12 {
		t.Fatalf("unit norm = %v", n)
	}
}

func TestClampNormBounds(t *testing.T) {
	s := Signature{Hunger: 10}
	ClampNorm(&s, MinNorm, MaxNorm)
	if n := Norm(s); math.Abs(n-MaxNorm) > 1e-12 {
		t.Fatalf("norm after clamp high = %v", n)
	}
	s = Signature{Hunger: 0.01}
	ClampNorm(&s, MinNorm, MaxNorm)
	if n := Norm(s); math.Abs(n-MinNorm) > 1e-12 {
		t.Fatalf("norm after clamp low = %v", n)
	}
	var zero Signature
	ClampNorm(&zero, MinNorm, MaxNorm)
	if zero != (Signature{}) {
		t.Fatal("clamp moved the zero vector")
	}
}

func TestSampleIsDeterministicAndFavoursTheBest(t *testing.T) {
	eff := []float64{0.1, 0.9, 0.3}
	a, b := rand.New(rand.NewPCG(7, 8)), rand.New(rand.NewPCG(7, 8))
	for i := 0; i < 50; i++ {
		if x, y := Sample(a, eff, 0.15), Sample(b, eff, 0.15); x != y {
			t.Fatalf("draw %d differs: %d vs %d", i, x, y)
		}
	}
	counts := [3]int{}
	r := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < 2000; i++ {
		counts[Sample(r, eff, 0.3)]++
	}
	if counts[1] < counts[0] || counts[1] < counts[2] {
		t.Fatalf("best action not favoured: %v", counts)
	}
	if counts[0] == 0 && counts[2] == 0 {
		t.Fatalf("sampling never explored: %v", counts)
	}
	if Sample(r, eff, 0) != 1 {
		t.Fatal("zero temperature is not argmax")
	}
	if Sample(r, nil, 0.1) != -1 {
		t.Fatal("empty eff did not return -1")
	}
}

func TestEntropyShrinksWithSharpness(t *testing.T) {
	eff := []float64{0.1, 0.9, 0.3}
	loose, tight := Entropy(eff, 1), Entropy(eff, 0.05)
	if !(loose > tight) {
		t.Fatalf("entropy loose=%v tight=%v", loose, tight)
	}
	if Entropy(eff, 0) != 0 {
		t.Fatal("argmax entropy should be 0")
	}
}

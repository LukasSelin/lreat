package habit

import "testing"

func TestSlotsAreGivenOnceAndNeverMove(t *testing.T) {
	n := Slots()
	a := Register("test/a")
	b := Register("test/b")
	if a != n || b != n+1 {
		t.Fatalf("slots %d, %d; want %d, %d", a, b, n, n+1)
	}
	if again := Register("test/a"); again != a {
		t.Fatalf("registering test/a again gave %d, want %d", again, a)
	}
	if i, ok := Slot("test/b"); !ok || i != b {
		t.Fatalf("Slot(test/b) = %d, %v", i, ok)
	}
	if _, ok := Slot("test/none"); ok {
		t.Fatal("an unregistered key should hold no slot")
	}
	if keys := Keys(); keys[a] != "test/a" || keys[b] != "test/b" {
		t.Fatalf("keys = %v", keys)
	}
}

func TestGrowKeepsWhatItHad(t *testing.T) {
	s := []float64{1, 2}
	g := Grow(s, 4)
	if len(g) != 4 || g[0] != 1 || g[1] != 2 || g[2] != 0 || g[3] != 0 {
		t.Fatalf("grown = %v", g)
	}
	if same := Grow(g, 3); len(same) != 4 {
		t.Fatal("growing to less should change nothing")
	}
}

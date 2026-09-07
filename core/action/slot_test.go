package action

import (
	"testing"

	"lreat/core/habit"
	"lreat/core/world"
)

// An act that enters the catalog after an agent was imprinted takes a
// fresh slot for it: seeded from the act's prior at its next decision,
// with everything it had already learned left where it was.
func TestALateActTakesAFreshSlot(t *testing.T) {
	w := world.New(9)
	a := blank(w, "a")
	Imprint(a)
	i := Index(Forage)
	a.Habits[i][habit.Lack] = 0.9 // something learned, to see it kept
	a.Reach[i] = 0.5

	late := &Def{Key: "test/late", Name: "late", Ticks: 1, Reach0: 0.3,
		Prior: habit.Signature{habit.Curious: 1}}
	if habit.Register(late.Key) != len(Catalog) {
		t.Fatal("a late act should take the next slot")
	}
	Catalog = append(Catalog, late)
	Count = len(Catalog)
	defer func() {
		Catalog = Catalog[:len(Catalog)-1]
		Count = len(Catalog)
	}()

	Imprint(a)
	w.Room()
	j := Index(late)
	if len(a.Habits) <= j || len(w.ReachFloor) <= j {
		t.Fatal("tables should have grown to hold the late act")
	}
	if a.Habits[j] != late.Prior || a.Reach[j] != late.Reach0 {
		t.Fatalf("late slot = %v reach %v, want its prior", a.Habits[j], a.Reach[j])
	}
	if a.Habits[i][habit.Lack] != 0.9 || a.Reach[i] != 0.5 {
		t.Fatal("seeding a fresh slot should leave learned ones alone")
	}
	if a.Seeded != len(Catalog) {
		t.Fatalf("seeded = %d, want %d", a.Seeded, len(Catalog))
	}
}

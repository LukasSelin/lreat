package action

import (
	"math/rand/v2"
	"slices"
	"testing"

	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/world"
)

func TestGatedIsTheCraftsAndLearning(t *testing.T) {
	// Paving is gated with the crafts: a settlement that has not yet learned
	// to work in stone has more pressing uses for its timber than the common
	// ground, and masonry is what opens it.
	want := map[*Def]bool{
		Craft: true, Teach: true, Study: true, Pave: true,
		Fish: true, Hunt: true, Irrigate: true, PlantTrees: true,
		Cook: true, Quarry: true, BuildGranary: true, Smelt: true, BuildTavern: true,
		FoundMarket: true,
	}
	for _, i := range Gated() {
		if !want[Catalog[i]] {
			t.Errorf("%s should not be gated", Catalog[i].Name)
		}
		delete(want, Catalog[i])
	}
	for d := range want {
		t.Errorf("%s should be gated", d.Name)
	}
}

func TestStudyBroadensOnlyWhatIsGated(t *testing.T) {
	w := world.New(1)
	a := blank(w, "a")
	Imprint(a)
	before := slices.Clone(a.Reach)
	Broaden(a, w)
	for i, d := range Catalog {
		switch {
		case d.Reach0 < 1 && !(a.Reach[i] > before[i]):
			t.Errorf("%s did not come closer", d.Name)
		case d.Reach0 == 1 && a.Reach[i] != before[i]:
			t.Errorf("%s moved though it was already in reach", d.Name)
		}
	}
	w.Mods.StudyRate = 2
	b := blank(w, "b")
	Imprint(b)
	Broaden(b, w)
	i := Index(Study)
	if !(b.Reach[i] > a.Reach[i]) {
		t.Fatal("written records should make study broaden reach faster")
	}
}

func TestTeachingPassesRecognitionOn(t *testing.T) {
	w := world.New(2)
	teacher, student := blank(w, "teacher"), blank(w, "student")
	Imprint(teacher)
	Imprint(student)
	i := Index(Craft)
	teacher.Reach[i] = 1
	teacher.Habits[i][habit.Stock] = -0.9 // an odd habit, to see it travel
	Pass(teacher, student, entity.Crafting)
	if student.Reach[i] != TaughtReach {
		t.Fatalf("student reach = %v, want %v", student.Reach[i], TaughtReach)
	}
	if !(student.Habits[i][habit.Stock] < Craft.Prior[habit.Stock]) {
		t.Fatal("student's habit did not move toward the teacher's")
	}
	student.Reach[i] = 1
	Pass(teacher, student, entity.Crafting)
	if student.Reach[i] != 1 {
		t.Fatal("teaching should never take reach away")
	}
}

func TestChildrenInheritHabitsAndAShareOfReach(t *testing.T) {
	w := world.New(3)
	parent, child := blank(w, "parent"), blank(w, "child")
	Imprint(parent)
	i := Index(Study)
	parent.Reach[i] = 1
	parent.Habits[i][habit.Lack] = 0.8
	w.ReachFloor[Index(Craft)] = Opened

	Inherit(child, parent, w, nil)
	if !child.Imprinted {
		t.Fatal("an heir should count as imprinted")
	}
	if !slices.Equal(child.Habits, parent.Habits) {
		t.Fatal("without drift a child should copy its parent exactly")
	}
	if child.Reach[i] != InheritReach {
		t.Fatalf("child reach = %v, want %v", child.Reach[i], InheritReach)
	}
	if child.Reach[Index(Craft)] != Opened {
		t.Fatal("a child should be born with what the settlement has opened")
	}

	drifted := blank(w, "drifted")
	Inherit(drifted, parent, w, rand.New(rand.NewPCG(1, 2)))
	if slices.Equal(drifted.Habits, parent.Habits) {
		t.Fatal("with drift a child should differ from its parent")
	}
	for k := habit.Honesty; k <= habit.Caution; k++ {
		if drifted.Habits[i][k] != parent.Habits[i][k] {
			t.Fatal("drift touched a frozen coordinate")
		}
	}
}

func TestTheSettlementsFloorRaisesEveryoneAtTheirNextDecision(t *testing.T) {
	w := world.New(4)
	a := blank(w, "a")
	Imprint(a)
	i := Index(Study)
	w.ReachFloor[i] = Opened
	Candidates(a, w)
	if a.Reach[i] != Opened {
		t.Fatalf("reach = %v, want the floor %v", a.Reach[i], Opened)
	}
}

package system

import (
	"testing"

	"lreat/core/action"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/observe"
)

func TestDiscoveryOpensActionsToEveryone(t *testing.T) {
	w := fitWorld(11)
	populate(w, 3)
	w.Knowledge = 1000
	for i := 0; i < 3; i++ {
		w.Agents[i].Skills[entity.Scholarship] = 0.5
	}
	Discover(w)
	if !w.Has("writing") {
		t.Fatal("writing should have been discovered")
	}
	for _, name := range []string{"study", "teach"} {
		if f := w.ReachFloor[action.Index(action.ByName(name))]; f != action.Opened {
			t.Fatalf("%s floor = %v, want %v", name, f, action.Opened)
		}
	}
	Decide(w)
	for _, a := range w.Agents {
		if a.Reach[action.Index(action.Study)] < action.Opened {
			t.Fatalf("%s was not raised to the settlement's floor", a.Name)
		}
	}
}

func TestChildrenAreBornRecognising(t *testing.T) {
	w := fitWorld(12)
	parent := w.Spawn("parent", need.Neutral())
	parent.Needs = need.Levels{1, 1, 1, 1, 1}
	action.Imprint(parent)
	parent.Reach[action.Index(action.Craft)] = 1
	born := 0
	for i := 0; i < 5000 && born == 0; i++ {
		Population(w)
		born = len(w.Agents) - 1
	}
	if born == 0 {
		t.Skip("no birth in 5000 ticks on this seed")
	}
	child := w.Agents[1]
	if !child.Imprinted {
		t.Fatal("a child should be born imprinted")
	}
	if got := child.Reach[action.Index(action.Craft)]; got != action.InheritReach {
		t.Fatalf("child's reach = %v, want %v", got, action.InheritReach)
	}
}

func TestDeathsAreCounted(t *testing.T) {
	w := fitWorld(13)
	a := w.Spawn("a", need.Neutral())
	a.Starving = StarvationTicks + 1
	Population(w)
	if w.Deaths != 1 || len(w.Agents) != 0 {
		t.Fatalf("deaths = %d, agents = %d", w.Deaths, len(w.Agents))
	}
	if observe.Take(w).Deaths != 1 {
		t.Fatal("snapshot should carry the death count")
	}
}

func TestHabitsGrowApartUnderRecognition(t *testing.T) {
	w := fitWorld(14)
	populate(w, 15)
	if s := observe.Take(w); s.HabitSpread > 1e-9 {
		t.Fatalf("before anyone has decided, spread = %v", s.HabitSpread)
	}
	Run(w, 1500)
	s := observe.Take(w)
	t.Logf("spread %.3f, reach %.2f, gated %.2f, entropy %.2f, deaths %d", s.HabitSpread, s.MeanReach, s.GatedReach, s.ChoiceEntropy, s.Deaths)
	if !(s.HabitSpread > 0) {
		t.Fatal("experience should have moved habits apart")
	}
	if !(s.GatedReach > 0.5) {
		t.Fatal("study and teaching should have brought the crafts closer")
	}
	if s.ChoiceEntropy <= 0 && w.Choices > 0 {
		t.Fatal("sampled choices should carry some entropy")
	}
	n := 0
	for _, e := range w.Log.All() {
		if e.Kind == event.Born {
			n++
		}
	}
	t.Logf("born %d", n)
}

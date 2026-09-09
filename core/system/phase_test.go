package system

import (
	"strings"
	"testing"

	"lreat/core/action"
	"lreat/core/habit"
	"lreat/core/world"
)

// A day handed out phase by phase is the same day, in the same order.
func TestStepWithRunsThePhasesInStepOrder(t *testing.T) {
	a, b := world.New(5), world.New(5)
	for i := 0; i < 20; i++ {
		a.Spawn("a", a.RandomPersonality())
		b.Spawn("a", b.RandomPersonality())
	}
	var names []string
	for i := 0; i < 300; i++ {
		Step(a)
		names = names[:0]
		StepWith(b, func(name string, run func()) {
			names = append(names, name)
			run()
		})
	}
	want := make([]string, len(Phases))
	for i, p := range Phases {
		want[i] = p.Name
	}
	if got := strings.Join(names, " "); got != strings.Join(want, " ") {
		t.Fatalf("the phases ran as %q", got)
	}
	if digest(a) != digest(b) {
		t.Fatal("a day taken phase by phase came out different from a day taken whole")
	}
}

// The reach floor grows when a slot is given, and every agent used to grow
// it for itself while deciding - the one write in the read-only phase.
// Ready grows it once, before anybody decides.
func TestReadySetsTheReachFloorBeforeAnyoneDecides(t *testing.T) {
	w := world.New(1)
	w.ReachFloor = nil
	action.Ready(w)
	if len(w.ReachFloor) != habit.Slots() {
		t.Fatalf("floor has %d slots after Ready, want %d", len(w.ReachFloor), habit.Slots())
	}
}

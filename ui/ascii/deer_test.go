package ascii_test

import (
	"testing"

	"lreat/core/observe"
	"lreat/core/system"
	"lreat/core/world"
	"lreat/ui/ascii"
)

// A herd shows on the map: every deer is drawn as a d in the deer's colour,
// and a person standing on the same ground is drawn over it.
func TestDeerAreDrawnAsDeer(t *testing.T) {
	cfg := world.DefaultConfig()
	cfg.Deer = 12
	w := world.NewWith(3, cfg)
	for i := 0; i < 10; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	w.Populate()
	system.Run(w, 20)
	s := observe.Take(w)
	deer, people := 0, 0
	for _, row := range ascii.Render(s.Map) {
		for _, c := range row {
			switch c.Ch {
			case 'd':
				deer++
				if c.Color != ascii.AgentDeer {
					t.Errorf("a deer is drawn in colour %d, not the deer's", c.Color)
				}
			case '@':
				people++
			}
		}
	}
	if deer == 0 {
		t.Fatal("no deer drawn on a map with a dozen in the woods")
	}
	if deer > s.Creatures {
		t.Fatalf("%d deer drawn, %d alive", deer, s.Creatures)
	}
	if people == 0 {
		t.Fatal("no people drawn")
	}
}

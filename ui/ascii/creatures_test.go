package ascii_test

import (
	"testing"

	"lreat/core/observe"
	"lreat/core/system"
	"lreat/core/world"
	"lreat/ui/ascii"
)

// The creatures show on the map: every deer is drawn as a d, every boar a b
// and every hare an h, in the creatures' colour, and a person standing on
// the same ground is drawn over them.
func TestCreaturesAreDrawnAsWhatTheyAre(t *testing.T) {
	cfg := world.DefaultConfig()
	cfg.Deer, cfg.Boar, cfg.Hare = 12, 6, 12
	w := world.NewWith(3, cfg)
	for i := 0; i < 10; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	w.Populate()
	system.Run(w, 20)
	s := observe.Take(w)
	drawn := map[rune]int{}
	for _, row := range ascii.Render(s.Map) {
		for _, c := range row {
			switch c.Ch {
			case 'd', 'b', 'h':
				drawn[c.Ch]++
				if c.Color != ascii.AgentCreature {
					t.Errorf("a %c is drawn in colour %d, not the creatures'", c.Ch, c.Color)
				}
			case '@':
				drawn['@']++
			}
		}
	}
	for _, g := range []rune{'d', 'b', 'h', '@'} {
		if drawn[g] == 0 {
			t.Errorf("no %c drawn on a map with some", g)
		}
	}
	if drawn['d']+drawn['b']+drawn['h'] > s.Creatures {
		t.Fatalf("%d creatures drawn, %d alive", drawn['d']+drawn['b']+drawn['h'], s.Creatures)
	}
}

package system

import (
	"fmt"

	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/world"
)

const (
	// StarvationTicks is how long an agent survives at the bottom of the
	// physiological tier.
	StarvationTicks = 60
	// BirthChance is the per-tick probability that a thriving agent has a child.
	BirthChance = 0.0015
	// MaxPopulation caps growth so runs stay bounded.
	MaxPopulation = 400
)

// Population handles deaths and births. Births need the three lower tiers
// met, which is why a city that cannot feed and protect its people does not
// grow no matter how much food is in the market.
func Population(w *world.World) {
	alive := w.Agents[:0]
	for _, a := range w.Agents {
		if a.Starving > StarvationTicks {
			w.Emit(event.Died, a.ID, 0, "%s starved", a.Name)
			continue
		}
		alive = append(alive, a)
	}
	w.Agents = alive

	n := len(w.Agents)
	for i := 0; i < n; i++ {
		a := w.Agents[i]
		if len(w.Agents) >= MaxPopulation {
			break
		}
		if a.Needs[need.Physiological] < 0.7 || a.Needs[need.Safety] < 0.6 || a.Needs[need.Belonging] < 0.6 {
			continue
		}
		if w.RNG.Float64() >= BirthChance {
			continue
		}
		child := w.SpawnAt(fmt.Sprintf("%s-%d", a.Name, w.Tick), w.Mutate(a.Personality), a.Pos)
		child.Inventory[entity.Food] = 1
		child.Shelter = a.Shelter * 0.8
		// Values are inherited, not drawn fresh. This is what lets a
		// settlement keep a character across generations.
		child.Norms = belief.Inherit(a.Norms, w.RNG)
		child.Temperament = entity.InheritTemperament(a.Temperament, w.RNG)
		// Parent and child start as intimates, not strangers: the bond is
		// strong, warm, and already counts as a meeting.
		for _, pair := range [][2]*entity.Agent{{child, a}, {a, child}} {
			b := pair[0].Know(pair[1].ID, w.Tick)
			b.Strength, b.Regard, b.Expect, b.Met = 0.8, 0.5, 0.6, 1
		}
		w.Emit(event.Born, child.ID, a.ID, "%s was born to %s", child.Name, a.Name)
	}
}

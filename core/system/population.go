package system

import (
	"fmt"

	"lreat/core/action"
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
	BirthChance = 0.006
	// MaxPopulation caps growth so runs stay bounded.
	MaxPopulation = 400
	// InheritedSkill is the share of a parent's skills a child is born with,
	// under recognition.
	InheritedSkill = 0.5
)

// Population handles deaths and births. Births need the three lower tiers
// met, which is why a city that cannot feed and protect its people does not
// grow no matter how much food is in the market. They also need a parent in
// the years between growing up and declining, so a settlement's ability to
// replace itself depends on how many of its people are that age.
//
// It also writes the settlement's vital record on the way past: what each
// death was of, and — for everyone who had no child — the first thing that
// stood in the way. Neither is anything the simulation reads; both are what
// makes a population curve afterwards say why it went the way it did.
func Population(w *world.World) {
	w.Vitals.Born, w.Vitals.Died, w.Vitals.Gates = 0, 0, [world.GateCount]int{}
	alive := w.Agents[:0]
	for _, a := range w.Agents {
		if a.Starving > StarvationTicks {
			w.Deaths++
			w.Vitals.Starved++
			w.Vitals.Died++
			w.Emit(event.Died, a.ID, 0, "%s starved", a.Name)
			continue
		}
		// Old age is a rising risk, not an appointment, and a body already
		// worn down by hunger and bad housing gives out sooner than a kept
		// one of the same years.
		age := a.Age(w.Tick)
		if w.RNG.Float64() < entity.Frailty(age)*(1.5-need.Clamp(a.Health)) {
			w.Deaths++
			w.Vitals.Failed++
			w.Vitals.Died++
			w.Emit(event.Died, a.ID, 0, "%s died of old age at %d", a.Name, age)
			continue
		}
		alive = append(alive, a)
	}
	w.Agents = alive

	n := len(w.Agents)
	for i := 0; i < n; i++ {
		a := w.Agents[i]
		if len(w.Agents) >= MaxPopulation {
			w.Vitals.Gates[world.Crowded] += n - i
			break
		}
		// Every agent is counted under the first thing standing between it
		// and a child, in the order the conditions are checked, so that the
		// gates add up to the population and can be read as a funnel.
		age := a.Age(w.Tick)
		if !entity.Fertile(age) {
			if age < entity.Maturity {
				w.Vitals.Gates[world.Young]++
			} else {
				w.Vitals.Gates[world.Spent]++
			}
			continue
		}
		switch {
		case a.Needs[need.Physiological] < 0.7:
			w.Vitals.Gates[world.Hungry]++
			continue
		case a.Needs[need.Safety] < 0.6:
			w.Vitals.Gates[world.Unsafe]++
			continue
		case a.Needs[need.Belonging] < 0.6:
			w.Vitals.Gates[world.Alone]++
			continue
		}
		// Nothing was in the way; from here it is only the draw.
		w.Vitals.Gates[world.Ready]++
		if w.RNG.Float64() >= BirthChance {
			continue
		}
		w.Vitals.Births++
		w.Vitals.Born++
		child := w.SpawnAt(fmt.Sprintf("%s-%d", a.Name, w.Tick), w.Mutate(a.Personality), a.Pos)
		child.Born = w.Tick
		child.Inventory[entity.Food] = 1
		child.Shelter = a.Shelter * 0.8
		child.Vitality = w.InheritVitality(a.Vitality)
		// Values are inherited, not drawn fresh. This is what lets a
		// settlement keep a character across generations.
		child.Norms = belief.Inherit(a.Norms, w.RNG)
		child.Temperament = entity.InheritTemperament(a.Temperament, w.RNG)
		// Under recognition a child grows up in its parent's work and starts
		// with a share of the skill, so a farming family stays a farming
		// family. Without this every generation began at nothing and the
		// fields emptied with each founder's death. The value rule keeps
		// its original design, in which every child starts from nothing.
		if w.Rules.Fit {
			for s := range child.Skills {
				child.Skills[s] = InheritedSkill * a.Skills[s]
			}
		}
		// Habits pass down too: what a parent has come to recognise as
		// calling for what, the child starts out recognising. Under
		// recognition they drift a little; under value they are copied, so
		// that rule's runs draw nothing extra from the RNG.
		if w.Rules.Fit {
			action.Inherit(child, a, w, w.RNG)
		} else {
			action.Inherit(child, a, w, nil)
		}
		// Parent and child start as intimates, not strangers: the bond is
		// strong, warm, and already counts as a meeting.
		for _, pair := range [][2]*entity.Agent{{child, a}, {a, child}} {
			b := pair[0].Know(pair[1].ID, w.Tick)
			b.Strength, b.Regard, b.Expect, b.Met = 0.8, 0.5, 0.6, 1
		}
		w.Emit(event.Born, child.ID, a.ID, "%s was born to %s", child.Name, a.Name)
	}
}

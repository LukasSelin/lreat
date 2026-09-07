package system

import (
	"fmt"

	"lreat/core/action"
	"lreat/core/belief"
	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/world"
)

const (
	// Starvation is how long an agent survives at the bottom of the
	// physiological tier. Two months of it is the end.
	Starvation = 2 * clock.Month
	// BirthChance is the daily probability that a thriving agent has a
	// child. It is written against the year because what has to stay fixed
	// as the calendar changes is how many children a fertile life brings,
	// not how many a day does: one chance a year, over the twenty-five
	// fertile years, for a parent whose lower three tiers are all met at
	// once - which happens about three tenths of the time, so a fertile life
	// that runs its course brings seven or eight children.
	//
	// That is a human number and it is set by the childhood. A settlement
	// that waits fifteen years for a birth to become a worker, and feeds it
	// the whole way, needs the fertility a pre-modern people actually had;
	// at three fifths of this, which is what the five-year childhood was
	// tuned with, the same settlements came out at a median of 92 against
	// 163. See docs/action-space.md.
	BirthChance = 1.0 / clock.Year
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
func Population(w *world.World) {
	alive := w.Agents[:0]
	for _, a := range w.Agents {
		if a.Starving > Starvation {
			w.Deaths++
			w.Emit(event.Died, a.ID, 0, "%s starved", a.Name)
			continue
		}
		// Old age is a rising risk, not an appointment, and a body already
		// worn down by hunger and bad housing gives out sooner than a kept
		// one of the same years.
		age := a.Age(w.Tick)
		if w.RNG.Float64() < entity.Frailty(age)*(1.5-need.Clamp(a.Health)) {
			w.Deaths++
			w.Emit(event.Died, a.ID, 0, "%s died of old age at %d", a.Name, clock.Years(age))
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
		if !entity.Fertile(a.Age(w.Tick)) {
			continue
		}
		if a.Needs[need.Physiological] < 0.7 || a.Needs[need.Safety] < 0.6 || a.Needs[need.Belonging] < 0.6 {
			continue
		}
		if w.RNG.Float64() >= BirthChance {
			continue
		}
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

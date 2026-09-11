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
	// InheritedSkill is the share of a parent's skills a child is born with,
	// under recognition.
	InheritedSkill = 0.5
)

// MaxPopulation is a guard on the machine and not a fact about the world.
// Nothing in the settlement knows it is there: agents do not feel crowded at
// 4,999 and free at 5,001, and the land, the larder and the roofs are what a
// settlement is supposed to run out of. It exists because most of a tick is
// spent in loops over everybody - a fair number of them once per agent, so
// the cost goes as the square - and a run that grew without limit would stop
// finishing rather than tell anybody anything.
//
// It was 400, which was low enough to bind. A settlement held at a ceiling
// looks exactly like a settlement that found its level, and the difference
// is the whole of what a batch is read on, so the number it reported was
// being quietly decided here rather than out on the land. Seed 1 at sixty
// years stood at 400 with the cap on and climbed to 502 without it, and its
// people were worse fed and lonelier at the top - which is what finding a
// real ceiling looks like. The seeds that never reached 400 came out
// identical to the digit.
//
// Five thousand is far enough above what the map has ever carried that the
// land binds first, and near enough that a settlement which somehow ran away
// still stops rather than running the batch into the ground. world.Crowded
// counts who was turned away, which is how anybody can tell whether it is
// binding again.
//
// Zero, or anything below it, takes the guard off altogether: births are then
// gated by the land and the larder and nothing else, and the run is as long
// as the machine will bear. That is the honest setting for asking what a
// world actually carries - a globe is a great deal more ground than the
// valley the number was chosen against - and it is the caller's business to
// know that a day costs what the population squared costs.
var MaxPopulation = 5000

// MaxCreatures is the same guard for each kind of creature. A herd with
// nothing hunting it is bounded by its browse and by nothing else, and this
// is here for the day the browse is not enough. It is a kind's own, so that
// a warren at its ceiling does not stand between a doe and her fawn.
var MaxCreatures = 300

// Room says whether the settlement may take one more. An unset ceiling - zero
// or below - is no ceiling: nothing but the world stands between a fertile
// pair and a child.
func Room(n int) bool { return MaxPopulation <= 0 || n < MaxPopulation }

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
		// The settlement's record is of its people. A creature dies the
		// same two ways and is written down the same way, and is on no
		// other books.
		settles := a.Species().Settles
		if a.Starving > Starvation {
			if settles {
				w.Deaths++
				w.Vitals.Starved++
				w.Vitals.Died++
			}
			w.Emit(event.Died, a.ID, 0, "%s starved", a.Name)
			continue
		}
		// Old age is a rising risk, not an appointment, and a body already
		// worn down by hunger and bad housing gives out sooner than a kept
		// one of the same years.
		age := a.Age(w.Tick)
		if w.RNG.Float64() < a.Species().Life.Frailty(age)*(1.5-need.Clamp(a.Health)) {
			if settles {
				w.Deaths++
				w.Vitals.Failed++
				w.Vitals.Died++
			}
			w.Emit(event.Died, a.ID, 0, "%s died of old age at %d", a.Name, clock.Years(age))
			continue
		}
		alive = append(alive, a)
	}
	w.Agents = alive
	w.Reindex()

	// The people are counted apart from the creatures: the ceiling is on
	// the settlement, the funnel is the settlement's, and a creature is
	// walked past to its own bearing. What is born during the walk counts
	// toward the ceiling as it always did, and nothing born today is walked.
	n := len(w.Agents)
	people := w.People()
	kinds := map[*entity.Species]int{}
	for _, a := range w.Agents {
		if !a.Species().Settles {
			kinds[a.Species()]++
		}
	}
	were, seen, crowded := people, 0, false
	for i := 0; i < n; i++ {
		a := w.Agents[i]
		if !a.Species().Settles {
			if bear(a, w, kinds[a.Species()]) {
				kinds[a.Species()]++
			}
			continue
		}
		if crowded {
			continue
		}
		if !Room(people) {
			w.Vitals.Gates[world.Crowded] += were - seen
			crowded = true
			continue
		}
		seen++
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
		people++
		child := w.SpawnAt(fmt.Sprintf("%s-%d", a.Name, w.Tick), w.Mutate(a.Personality), a.Pos)
		child.Born = w.Tick
		child.Inventory[entity.Food] = 1
		child.Shelter = a.Shelter * 0.8
		child.Body = w.InheritBody(a.Body)
		child.Mind = w.InheritMind(a.Mind)
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

// bear is a creature's chance of young today, and the young if it has any.
// It is the same ladder a person climbs - grown, fed, safe, in company, and
// then the draw - at the species' own chance, and the young is of the
// parent's kind, with its body, its mind and its habits drifted a little,
// as a child's are. It holds nothing else, because its kind holds nothing
// else. Nothing here is counted in the settlement's vitals.
func bear(a *entity.Agent, w *world.World, creatures int) bool {
	sp := a.Species()
	if creatures >= MaxCreatures || !sp.Life.Fertile(a.Age(w.Tick)) {
		return false
	}
	if a.Needs[need.Physiological] < 0.7 || a.Needs[need.Safety] < 0.6 || a.Needs[need.Belonging] < 0.6 {
		return false
	}
	if w.RNG.Float64() >= sp.Bears {
		return false
	}
	child := w.SpawnKind(sp, fmt.Sprintf("%s-%d", a.Name, w.Tick), w.Mutate(a.Personality), a.Pos)
	child.Born = w.Tick
	child.Body = w.InheritBody(a.Body)
	child.Mind = w.InheritMind(a.Mind)
	if w.Rules.Fit {
		action.Inherit(child, a, w, w.RNG)
	} else {
		action.Inherit(child, a, w, nil)
	}
	w.Emit(event.Born, child.ID, a.ID, "%s was born to %s", child.Name, a.Name)
	return true
}

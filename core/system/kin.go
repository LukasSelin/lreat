package system

import (
	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/world"
)

// The household: what a settlement does about the people in it who cannot
// yet do anything about themselves.
//
// A child used to be a small adult that happened to be weak, and nothing
// whatever depended on anybody looking after it. That made rearing a charge
// on a parent with nothing on the other side of the ledger, and it measured
// exactly as badly as that sounds: over twenty-four seeds the acts cost
// three fifths of the median population and bought a third of the deaths.
// The acts were not the mistake. What was missing was the other half - that
// a child nobody tends does badly - without which tending is a kindness
// rather than the thing a settlement lives or dies by.
//
// So a childhood is now a stock, entity.Agent.Tended, that being fed, taught
// and housed puts in and that time takes out, and two things read it: whether
// the child lives (entity.Neglect, applied in Population) and what body it
// grows into (entity.RearedBody, drifted here). Neither is a rule about what
// anybody must do. A parent that never comes back is free not to, and its
// line ends.

const (
	// tendFade is the share of a childhood's tending that time takes back
	// each day. Three years of nobody doing anything is a childhood with
	// nothing left in it.
	tendFade = 1.0 / (3 * clock.Year)
	// tendRoof is what a day under a parent's roof is worth to a childhood,
	// at the full worth of the roof. It carries most of a childhood, and
	// that is the balance the whole thing turns on. A roof is not a
	// generous reading of rearing - it is the daily fact of having somewhere
	// to be, against acts that are occasional by nature - and putting the
	// weight here is what decides who pays for the childhood. Read the
	// other way, with the acts carrying it, every settlement pays a
	// standing tax it cannot choose to stop paying, because agents choose
	// by recognition and cannot decide to rear harder; over six seeds that
	// read took the median settlement to 30. Here the charge falls where a
	// settlement is already failing - the unhoused, the orphaned, the ones
	// with nothing to spare - which is where a childhood should bite and
	// where it means something.
	tendRoof = 0.0008
	// bodyRate is how fast the body a child is building follows how it is
	// being reared. Five years or so to come round, so a bad stretch tells
	// and a bad week does not - and so a childhood put right in its middle
	// still comes out most of the way.
	bodyRate = 1.0 / (5 * clock.Year)
)

// Rearing is the childhood: a parent's roof reaching over its own young, the
// slow fade of what has been done for them, and the body that is being built
// out of it. Shelter is kept per body and a house is not - a roof raised by
// one person covers everybody asleep under it, and the people asleep under it
// are that person's children.
//
// It is not an act, and it is the only part of rearing that is not: feeding
// and teaching are things a parent does or fails to do, and a roof is a thing
// that stands. Nobody decides each morning to go on sheltering their
// children. The roof is worth what the parent's is, never more, so a
// household is exactly as well housed as the person who built it.
func Rearing(w *world.World) {
	roofs := make(map[entity.ID]float64, len(w.Agents))
	for _, a := range w.Agents {
		roofs[a.ID] = a.Shelter
	}
	for _, a := range w.Agents {
		if entity.Adult(a.Age(w.Tick)) {
			continue
		}
		// What was done for this child yesterday counts for a little less
		// today. Tending is a stock and not a mark on a record: a parent
		// that fed a child once and never again reared it once.
		a.Tended -= tendFade * a.Tended

		if roof, ok := roofs[a.Parent]; ok && a.Parent != 0 {
			if roof > a.Shelter {
				a.Shelter = roof
			}
			a.Tended = min(1, a.Tended+tendRoof*roof)
		}

		// The body follows the rearing. It is still the body it was born
		// with underneath - what a childhood does is move it, over years,
		// toward what that childhood is worth.
		a.Vitality += (entity.RearedBody(a.Tended) - a.Vitality) * bodyRate
	}
}

// Neglected carries off the children nobody came back to. It runs inside
// Population, with the other ways of dying, because it is one of them: an
// infant that is not fed does not starve in the way an adult does, over
// sixty days of an emptying larder. It simply does not wake up.
func neglected(w *world.World, a *entity.Agent, age int) bool {
	risk := entity.Neglect(age, a.Tended)
	if risk <= 0 || w.RNG.Float64() >= risk {
		return false
	}
	w.Deaths++
	w.Vitals.Lost++
	w.Vitals.Died++
	w.Emit(event.Died, a.ID, a.Parent, "%s was lost at %d, untended", a.Name, clock.Years(age))
	return true
}

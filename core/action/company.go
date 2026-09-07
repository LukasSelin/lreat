package action

import (
	"math"
	"sort"

	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/world"
)

// This file is the social life of the settlement: whom an agent seeks out,
// what it expects of them, how the meeting actually goes, and what it hears
// about everyone else while it is there. Reputation is nothing more than the
// sum of these private, second-hand, often wrong opinions.

const (
	// exploreBase is the chance an agent seeks out someone new regardless of
	// expectations; traditionalists explore less.
	exploreBase  = 0.1
	exploreRange = 0.25
	// expectRate is how far one meeting moves the expectation of the next.
	expectRate = 0.3
	// gossipItems is how many people a speaker mentions in one conversation.
	gossipItems = 3
	// gossipRate is how far hearsay moves the listener's opinion, before it
	// is scaled by trust and by how close the two are.
	gossipRate = 0.3
)

// clampUnit limits x to [-1,1].
func clampUnit(x float64) float64 {
	return math.Max(-1, math.Min(1, x))
}

// encounterGain converts the quality of a meeting into belonging. Even a bad
// meeting is company; a good one with a warm agent is worth a great deal.
func encounterGain(quality float64, t entity.Temperament) float64 {
	pleasant := math.Max(0, (quality+1)/2)
	return (0.08 + 0.22*pleasant) * (0.7 + 0.6*t.Warmth)
}

// FirstImpression is what a decides about o on meeting for the first time,
// before a word is said: good faith by temperament, kinship by affinity, and
// a little deference to visible standing.
func FirstImpression(a, o *entity.Agent) float64 {
	aff := entity.Affinity(a, o)
	return a.Temperament.Trust*0.3 + 0.5*(aff-0.5) + 0.1*math.Min(o.Reputation, 1)
}

// Introduce makes sure a has an opinion of o. On a first meeting the opinion
// is the first impression laid over whatever hearsay already seeded it. It
// returns the bond either way.
func Introduce(a, o *entity.Agent, tick int) *entity.Bond {
	b := a.Know(o.ID, tick)
	if b.Met > 0 {
		return b
	}
	first := FirstImpression(a, o)
	b.Regard = clampUnit(b.Regard + first)
	b.Expect = clampUnit(first + 0.3*b.Regard)
	return b
}

// Anticipate is how a expects a meeting with o to go. Known people are judged
// on the record; strangers on temperament and hearsay.
func Anticipate(a, o *entity.Agent) float64 {
	if b := a.Look(o.ID); b != nil {
		if b.Met > 0 {
			return b.Expect
		}
		return clampUnit(a.Temperament.Trust*0.4 - 0.2 + 0.5*b.Regard)
	}
	return clampUnit(a.Temperament.Trust*0.4 - 0.2)
}

// PickCompany chooses whom a would go and see. Mostly it is whoever they
// expect to enjoy, discounted by the walk; sometimes it is a stranger, more
// often for agents who do not hold tradition dear.
func PickCompany(a *entity.Agent, w *world.World) *entity.Agent {
	if len(w.Agents) < 2 {
		return nil
	}
	explore := exploreBase + exploreRange*(1-a.Norms[belief.Tradition])
	if a.Luck.Float64() < explore {
		return w.Other(a)
	}
	var best *entity.Agent
	bestScore := math.Inf(-1)
	for _, o := range w.Agents {
		if o == a {
			continue
		}
		s := Anticipate(a, o) - 0.01*float64(entity.Dist(a.Pos, o.Pos))
		if s > bestScore {
			best, bestScore = o, s
		}
	}
	return best
}

// Encounter is a meeting between a and o. Its quality comes from how alike
// they are and how they already regard each other, plus the luck of the day.
// Both sides update their bond toward what actually happened, and both hear
// the other's news. It returns the quality.
func Encounter(a, o *entity.Agent, w *world.World) float64 {
	ba := Introduce(a, o, w.Tick)
	bo := Introduce(o, a, w.Tick)

	aff := entity.Affinity(a, o)
	q := clampUnit((aff-0.5)*1.6 + 0.25*(ba.Regard+bo.Regard) + w.RNG.NormFloat64()*0.2)

	a.Needs.Add(need.Belonging, encounterGain(q, a.Temperament))
	o.Needs.Add(need.Belonging, encounterGain(q, o.Temperament)*0.6)

	for _, side := range []struct {
		b *entity.Bond
		t entity.Temperament
	}{{ba, a.Temperament}, {bo, o.Temperament}} {
		if q > 0 {
			side.b.Strength = belief.Clamp(side.b.Strength + (0.05+0.08*q)*(0.6+0.8*side.t.Warmth))
		} else {
			side.b.Strength = belief.Clamp(side.b.Strength + 0.05*q)
		}
		side.b.Regard = clampUnit(side.b.Regard + 0.05*q)
		side.b.Expect = clampUnit(belief.Update(side.b.Expect, q, expectRate))
		side.b.Met++
		side.b.LastSeen = w.Tick
	}

	Gossip(a, o, w)
	Gossip(o, a, w)

	switch {
	case q > 0.4:
		w.Emit(event.Met, a.ID, o.ID, "%s got on well with %s", a.Name, o.Name)
	case q < -0.3:
		w.Emit(event.Met, a.ID, o.ID, "%s fell out with %s", a.Name, o.Name)
	default:
		w.Emit(event.Met, a.ID, o.ID, "%s spent time with %s", a.Name, o.Name)
	}
	return q
}

// Gossip has speaker tell listener about the people the speaker feels most
// strongly about. The listener's opinion moves toward the speaker's by an
// amount set by how much the listener trusts people in general and this
// speaker in particular. This is how someone acquires an opinion of a person
// they have never met, and how a reputation travels faster than its owner.
func Gossip(listener, speaker *entity.Agent, w *world.World) {
	bl := listener.Look(speaker.ID)
	closeness := 0.3
	if bl != nil {
		closeness += 0.7 * bl.Strength
	}
	weight := gossipRate * listener.Temperament.Trust * closeness
	if weight <= 0 {
		return
	}

	idx := make([]int, 0, len(speaker.Bonds))
	for i := range speaker.Bonds {
		if speaker.Bonds[i].To != listener.ID {
			idx = append(idx, i)
		}
	}
	sort.SliceStable(idx, func(i, j int) bool {
		ri, rj := math.Abs(speaker.Bonds[idx[i]].Regard), math.Abs(speaker.Bonds[idx[j]].Regard)
		if ri != rj {
			return ri > rj
		}
		return speaker.Bonds[idx[i]].To < speaker.Bonds[idx[j]].To
	})

	told := 0
	for _, i := range idx {
		if told >= gossipItems {
			break
		}
		said := speaker.Bonds[i]
		if w.Find(said.To) == nil {
			continue
		}
		heard := listener.Know(said.To, w.Tick)
		// What you have seen for yourself outweighs what you are told. A
		// grudge earned first-hand survives a good deal of friendly talk.
		sway := weight / (1 + float64(heard.Met))
		heard.Regard = clampUnit(belief.Update(heard.Regard, said.Regard, sway))
		for s := range heard.Competence {
			heard.Competence[s] = belief.Clamp(belief.Update(heard.Competence[s], said.Competence[s], sway))
		}
		if heard.Met == 0 {
			// A person known only by report is anticipated on that report.
			heard.Expect = clampUnit(0.5 * heard.Regard)
		}
		told++
	}
}

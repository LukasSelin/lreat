package world

import "lreat/core/event"

// Vitals is the settlement's demographic record: how the population turned
// over, and — the part a headcount can never show — why it did not turn over
// more. A settlement dies out for one of two reasons, and the difference is
// everything: either it is losing people faster than it makes them, or it
// stopped making them a long time ago and is only now running out. The
// counts here are what tells those apart afterwards.
//
// Nothing in the simulation reads any of this. It is written on the way
// past by the population system and only ever looked at from outside.
type Vitals struct {
	// The cumulative record. Starved and Failed together are every death
	// there has been, so the two of them against Births is the whole of
	// whether a settlement is replacing itself.
	Births  int
	Starved int // ran out of food and stayed out
	Failed  int // a body gave out with age

	// What this tick alone did, so that a rate can be read off a run of
	// ticks rather than inferred from totals that only ever climb.
	Born, Died int

	// Gates is this tick's fertility funnel: every living agent counted
	// once, under the first condition that stopped it having a child. It is
	// the answer to "why is nobody being born", which is otherwise
	// invisible: an agent that never comes close to bearing looks exactly
	// like one that came close and lost the draw.
	Gates [GateCount]int
}

// Gate is why an agent had no child on a given tick. Every agent falls under
// exactly one of these each tick, in the order the population system checks
// them.
type Gate int

const (
	Young   Gate = iota // has not grown up yet
	Spent               // past its prime, out of its bearing years
	Hungry              // the larder is too thin to bear on
	Unsafe              // does not feel safe enough
	Alone               // has nobody enough to bear with
	Crowded             // the settlement is at its cap
	Ready               // nothing was in the way; the draw simply went the other way
	GateCount
)

var gateNames = [GateCount]string{"too young", "past prime", "hungry", "unsafe", "alone", "crowded", "ready"}

func (g Gate) String() string { return gateNames[g] }

// Gates returns every gate in the order the population system checks them.
func Gates() [GateCount]Gate {
	var gs [GateCount]Gate
	for i := range gs {
		gs[i] = Gate(i)
	}
	return gs
}

// Note is one thing worth telling afterwards: a birth, a death, a discovery.
type Note struct {
	Tick int
	Kind event.Kind
	Text string
}

// Chronicled is how many notes the world keeps. It is the settlement's short
// memory: enough to read the last of a collapse off, not a record of it.
const Chronicled = 64

// note keeps an event worth telling, dropping the oldest beyond Chronicled.
// It exists because a viewer cannot rely on seeing every tick — snapshots are
// dropped when the screen falls behind, and the ticks that get dropped in a
// collapse are exactly the ones worth reading. Keeping the last of them here
// means the view can always say who died last, however far behind it was.
func (w *World) note(e event.Event) {
	switch e.Kind {
	case event.Born, event.Died, event.Discovered:
	default:
		return
	}
	w.Chronicle = append(w.Chronicle, Note{Tick: e.Tick, Kind: e.Kind, Text: e.Text})
	if len(w.Chronicle) > Chronicled {
		w.Chronicle = w.Chronicle[len(w.Chronicle)-Chronicled:]
	}
}

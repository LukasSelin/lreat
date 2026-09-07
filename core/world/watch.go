package world

import "lreat/core/entity"

// Watching is how a single agent is followed through a settlement that has
// hundreds of them. The map shows where everyone went; it cannot show what
// any one of them had in mind on the way, and a decision leaves no trace
// once its plan is made. So the world keeps, for one agent at a time, the
// whole of what that agent weighed each time it decided: every action it
// could have taken, how well each one fitted the moment, and which one it
// settled on. One agent at a time because it is a debugging window and not
// a record: keeping it for everybody would cost every tick something, to be
// read for nobody.
//
// Nothing in the simulation may read any of this. It is written on the way
// past and only ever looked at from outside.

// Weighed is one action as it stood at the moment of choosing: where it
// would have been done, how strongly it recommended itself, and how likely
// the agent was to settle on it.
type Weighed struct {
	Action string
	Index  int // catalog position
	Target entity.Pos
	// Weight is how strongly the action put itself forward. Under the
	// recognition rule that is fit, in [-1,1]; under the value rule it is
	// worth per tick, unbounded.
	Weight float64
	// Reach is how far into reach the action was for this agent, 1 being
	// wholly within it. An action can fit a moment perfectly and still be
	// passed over for being out of reach.
	Reach float64
	// Chance is the probability the agent would have picked it. Under the
	// value rule the best is taken every time, so it is 1 or 0.
	Chance float64
	Chosen bool
}

// Deliberation is one decision, whole. Weighed is sorted with the strongest
// candidate first, whichever the agent actually took.
type Deliberation struct {
	Tick  int
	Agent entity.ID
	// Rule is "fit" when the agent chose by recognition and "value" when it
	// chose by expected worth, so the Weight column can be read correctly.
	Rule string
	// Intensity is how pressing the moment was, which is what sharpens a
	// recognition-based choice.
	Intensity float64
	// Entropy is how open the choice was, in nats. Zero is a decision that
	// made itself.
	Entropy float64
	Weighed []Weighed
}

// Chose returns the action the agent settled on.
func (d Deliberation) Chose() string {
	for _, c := range d.Weighed {
		if c.Chosen {
			return c.Action
		}
	}
	return ""
}

// Thoughts is how many of a watched agent's decisions are kept. It is a few
// minutes of a busy agent's life: enough to see a run of choices for what it
// is rather than one decision out of context.
const Thoughts = 32

// Watch follows an agent's decisions, forgetting whoever was followed
// before. Zero follows nobody, which is where every world starts.
func (w *World) Watch(id entity.ID) {
	if w.watched == id {
		return
	}
	w.watched, w.thoughts = id, nil
}

// Watching is the agent whose decisions are being kept, zero for nobody.
func (w *World) Watching() entity.ID { return w.watched }

// Remember keeps a decision, dropping the oldest beyond Thoughts.
func (w *World) Remember(d Deliberation) {
	w.thoughts = append(w.thoughts, d)
	if len(w.thoughts) > Thoughts {
		w.thoughts = w.thoughts[len(w.thoughts)-Thoughts:]
	}
}

// Recall returns the watched agent's kept decisions, oldest first.
func (w *World) Recall() []Deliberation { return w.thoughts }

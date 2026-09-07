package system

import (
	"math"
	"testing"

	"lreat/core/observe"
	"lreat/core/world"
)

// Following an agent keeps every decision it makes, whole: the actions it
// had before it, the odds it drew on, and the one it took.
func TestWatchingKeepsWholeDecisions(t *testing.T) {
	w := fitWorld(7)
	populate(w, 6)
	w.Watch(w.Agents[0].ID)
	Run(w, 60)

	ds := w.Recall()
	if len(ds) == 0 {
		t.Fatal("sixty ticks of a watched agent produced no decisions")
	}
	for _, d := range ds {
		if d.Agent != w.Agents[0].ID {
			t.Fatalf("kept somebody else's decision: %d", d.Agent)
		}
		if d.Rule != "fit" {
			t.Fatalf("recognition-based choice recorded as %q", d.Rule)
		}
		var total float64
		chosen := 0
		for _, c := range d.Weighed {
			total += c.Chance
			if c.Chosen {
				chosen++
			}
		}
		if chosen != 1 {
			t.Fatalf("tick %d: %d actions marked as taken, want exactly one", d.Tick, chosen)
		}
		if math.Abs(total-1) > 1e-9 {
			t.Fatalf("tick %d: the odds come to %.4f, not 1", d.Tick, total)
		}
		for i := 1; i < len(d.Weighed); i++ {
			if d.Weighed[i-1].Weight < d.Weighed[i].Weight {
				t.Fatal("candidates are not strongest first")
			}
		}
		if d.Chose() == "" {
			t.Fatal("a decision with nothing chosen in it")
		}
	}
}

// The value rule is watchable too, and says so, since a score per tick is
// not read the way a fit is.
func TestWatchingUnderTheValueRule(t *testing.T) {
	w := valueWorld(7)
	populate(w, 4)
	w.Watch(w.Agents[0].ID)
	Run(w, 40)

	ds := w.Recall()
	if len(ds) == 0 {
		t.Fatal("no decisions kept under the value rule")
	}
	for _, d := range ds {
		if d.Rule != "value" {
			t.Fatalf("value-based choice recorded as %q", d.Rule)
		}
		if d.Weighed[0].Chosen != true {
			t.Fatal("the value rule took something other than the best score")
		}
	}
}

// Nobody is watched unless somebody asks, and only one agent at a time.
func TestWatchingIsOneAgentAndOptional(t *testing.T) {
	w := fitWorld(3)
	populate(w, 5)
	Run(w, 30)
	if len(w.Recall()) != 0 {
		t.Fatal("decisions were kept for a world watching nobody")
	}

	w.Watch(w.Agents[0].ID)
	Run(w, 30)
	if len(w.Recall()) == 0 {
		t.Fatal("nothing kept for the agent being watched")
	}
	w.Watch(w.Agents[1].ID)
	if len(w.Recall()) != 0 {
		t.Fatal("looking at somebody else kept the last one's thinking")
	}
	Run(w, world.Thoughts*20)
	if got := len(w.Recall()); got > world.Thoughts {
		t.Fatalf("kept %d decisions, more than the %d it bounds itself to", got, world.Thoughts)
	}
}

// Watching is a window and not a hand: a run must come out the same whether
// or not somebody was looking at one of its agents.
func TestWatchingChangesNothing(t *testing.T) {
	seen, unseen := fitWorld(11), fitWorld(11)
	populate(seen, 12)
	populate(unseen, 12)
	seen.Watch(seen.Agents[3].ID)
	Run(seen, 400)
	Run(unseen, 400)
	if !equalSnapshots(observe.Take(seen), observe.Take(unseen)) {
		t.Fatal("a watched run diverged from an unwatched one")
	}
}

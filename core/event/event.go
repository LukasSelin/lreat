// Package event is the append-only record of everything that happened.
//
// Every state change emits an Event. The log is the replay, the debugger, and
// the raw material of the player's perception layer, which is a filter over it.
package event

import "lreat/core/entity"

// Kind classifies an event.
type Kind string

const (
	Acted Kind = "acted"
	Built Kind = "built"
	// Ruined is the other end of Built: something standing has gone back
	// to the ground it stood on.
	Ruined     Kind = "ruined"
	Traded     Kind = "traded"
	Met        Kind = "met"
	Taught     Kind = "taught"
	Guarded    Kind = "guarded"
	Born       Kind = "born"
	Died       Kind = "died"
	Discovered Kind = "discovered"

	// The moral and contractual life of the settlement.
	Stolen    Kind = "stolen"
	Given     Kind = "given"
	Requested Kind = "requested"
	Fulfilled Kind = "fulfilled"
	Unmet     Kind = "unmet"
	Avenged   Kind = "avenged"
)

// Event is one thing that happened at a tick.
type Event struct {
	Tick   int
	Kind   Kind
	Actor  entity.ID // zero for world-level events
	Target entity.ID // zero when nobody else was involved
	Text   string
}

// Log is a bounded append-only event store. When full it drops the oldest
// events and counts them, so long headless runs stay within memory.
type Log struct {
	capacity int
	events   []Event
	dropped  int
}

// NewLog creates a log that keeps at most capacity events.
func NewLog(capacity int) *Log {
	return &Log{capacity: capacity}
}

// Append records an event.
func (l *Log) Append(e Event) {
	if len(l.events) >= l.capacity {
		// Drop the oldest tenth in one move to amortize the shift.
		n := l.capacity / 10
		if n < 1 {
			n = 1
		}
		copy(l.events, l.events[n:])
		l.events = l.events[:len(l.events)-n]
		l.dropped += n
	}
	l.events = append(l.events, e)
}

// All returns every retained event, oldest first. The slice is shared;
// callers must not modify it.
func (l *Log) All() []Event { return l.events }

// Since returns retained events at or after tick.
func (l *Log) Since(tick int) []Event {
	i := len(l.events)
	for i > 0 && l.events[i-1].Tick >= tick {
		i--
	}
	return l.events[i:]
}

// Len is the number of retained events.
func (l *Log) Len() int { return len(l.events) }

// Dropped is the number of events discarded because the log was full.
func (l *Log) Dropped() int { return l.dropped }

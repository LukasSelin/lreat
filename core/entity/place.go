package entity

// What an agent knows of the country it lives in. An agent used to know all
// of it: every siting decision read the whole grid and took the nearest tile
// that was legal to build on, so nobody ever had a reason to go and look, and
// everybody standing in the same place picked the same plot. Knowledge of
// ground is now like knowledge of people - held by whoever gathered it, held
// in a small number of slots, and wrong as often as memory is wrong.

// MaxPlaces caps how many spots an agent carries in its head, as MaxBonds
// caps how many people. Nobody knows the whole country, and a settler who
// did would have no reason to walk out of it.
const MaxPlaces = 8

// Place is a piece of ground an agent has stood on and what it made of it.
//
// Worth is belief, in the way Bond.Competence is belief: what the ground
// promised the last time this agent looked at it, which may be nothing like
// what it promises now. A wood remembered beside a plot has since been felled
// by somebody else; the neighbours remembered two fields over have died. The
// only thing that corrects the record is standing there again. That gap
// between what is remembered and what is true is the whole reason a walk out
// to look at somewhere is worth anything.
type Place struct {
	Pos   Pos
	Worth float64 // in tiles of walking saved a day: the currency siting is costed in
	Seen  int     // the tick it was last looked at
}

// Remember writes down what this agent made of the ground at p. A place
// already known is simply brought up to date. A new one takes a free slot if
// there is one, and otherwise displaces the least promising place in mind,
// but only if it beats it: a head full of good ground is not emptied by
// walking over a bog.
//
// It reports whether the agent came away knowing anything it did not know
// before - a place it had never stood on, or one that has changed. That is
// what a day spent looking is judged by; a day that turns up nothing has to
// count for nothing or looking becomes its own reward.
func (a *Agent) Remember(p Pos, worth float64, tick int) bool {
	worst := -1
	for i := range a.Places {
		if a.Places[i].Pos == p {
			changed := a.Places[i].Worth != worth
			a.Places[i].Worth, a.Places[i].Seen = worth, tick
			return changed
		}
		if worst < 0 || a.Places[i].Worth < a.Places[worst].Worth {
			worst = i
		}
	}
	if len(a.Places) < MaxPlaces {
		a.Places = append(a.Places, Place{Pos: p, Worth: worth, Seen: tick})
		return true
	}
	if worth > a.Places[worst].Worth {
		a.Places[worst] = Place{Pos: p, Worth: worth, Seen: tick}
		return true
	}
	return false
}

// Knows returns what this agent thinks of the ground at p, and whether it has
// ever stood there.
func (a *Agent) Knows(p Pos) (Place, bool) {
	for i := range a.Places {
		if a.Places[i].Pos == p {
			return a.Places[i], true
		}
	}
	return Place{}, false
}

// Forget drops a place from mind, for ground that has turned out to be no
// use: built over by somebody else, or under water since the river moved.
func (a *Agent) Forget(p Pos) {
	for i := range a.Places {
		if a.Places[i].Pos == p {
			a.Places = append(a.Places[:i], a.Places[i+1:]...)
			return
		}
	}
}

// CopyPlaces returns a fresh slice of what this agent knows, for handing to
// somebody else. It has to be a copy: Habits and Reach are arrays and may be
// assigned across agents safely, but a shared slice would give parent and
// child one backing store, and the first append by either would be written
// into the other's memory.
func (a *Agent) CopyPlaces() []Place {
	if len(a.Places) == 0 {
		return nil
	}
	return append([]Place(nil), a.Places...)
}

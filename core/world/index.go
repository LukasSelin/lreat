package world

import "lreat/core/entity"

// Where everybody is. Every question about the people near a place used
// to walk the whole population, and most of a tick is such questions, so
// the cost of a settlement went as the square of it. The world files each
// agent under the chunk it stands in, and a question about a place reads
// the nine chunks around it and nothing else.
//
// What has to be kept is the order. Every one of those questions breaks
// its ties by taking the earliest agent in the population, and the
// population is in order of birth, which is order of ID; and the draw a
// visitor makes among the people it finds indexes the list it found them
// in. So each chunk's file is in ascending ID, and a walk over nine of
// them is a merge of nine sorted lists - the same people, in the same
// order, as the walk over everybody would have found.
//
// The file is rebuilt from the population once a tick, before anybody
// decides, and kept by the one thing that moves an agent between ticks.
// A test that sets a position by hand is caught when the file is next
// read: outside the deciding phase, a reader puts right any agent it finds
// filed under ground it is no longer standing on. On the default map the
// nine chunks are the whole map, so nothing set by hand there is missed.

// index is the file: who is in each chunk, and which chunk each is in.
type index struct {
	cells [][]*entity.Agent // by chunk, each in ascending ID
	filed []int32           // by ID: the chunk the agent is filed under, or -1
	byID  []*entity.Agent   // by ID, nil for the dead and the unborn
	// counted is how many were on file when it was last known to be whole.
	// A population that has changed size under it is refiled before it is
	// read, which is how one struck off by hand is found to be gone.
	counted int
	// frozen is set while agents decide side by side, when the file may be
	// read from several goroutines and so must not be put right by any.
	frozen bool
}

// Reindex files the whole population afresh.
func (w *World) Reindex() {
	g := w.Grid
	if len(w.cells) != len(g.Chunks) {
		w.cells = make([][]*entity.Agent, len(g.Chunks))
	}
	for i := range w.cells {
		w.cells[i] = w.cells[i][:0]
	}
	for i := range w.byID {
		w.byID[i] = nil
	}
	w.counted = 0
	for _, a := range w.Agents {
		w.file(a)
	}
}

// whole refiles the population if it has changed size since it was filed.
func (w *World) whole() {
	if !w.frozen && (len(w.cells) != len(w.Grid.Chunks) || w.counted != len(w.Agents)) {
		w.Reindex()
	}
}

// file puts a where it stands, at the end of that chunk's list. The
// population is filed in ID order, so the end is the right place.
func (w *World) file(a *entity.Agent) {
	id := int(a.ID)
	for len(w.byID) <= id {
		w.byID = append(w.byID, nil)
		w.filed = append(w.filed, -1)
	}
	w.byID[id] = a
	w.counted++
	c := w.Grid.ChunkOf(w.Grid.Index(a.Pos))
	w.cells[c] = append(w.cells[c], a)
	w.filed[id] = int32(c)
}

// enroll files a newcomer, already the last of the population. A file that
// has never been laid out for this map is laid out whole instead.
func (w *World) enroll(a *entity.Agent) {
	if len(w.cells) != len(w.Grid.Chunks) || w.counted != len(w.Agents)-1 {
		w.Reindex()
		return
	}
	w.file(a)
}

// Moved refiles a after its position has changed. Cheap when it is still
// in the chunk it was filed under, which is nearly always.
func (w *World) Moved(a *entity.Agent) {
	id := int(a.ID)
	if id >= len(w.filed) {
		w.file(a)
		return
	}
	c := int32(w.Grid.ChunkOf(w.Grid.Index(a.Pos)))
	if w.filed[id] == c {
		return
	}
	w.refile(a, c)
}

// refile moves a from the chunk it is filed under to chunk c, keeping both
// lists in ID order.
func (w *World) refile(a *entity.Agent, c int32) {
	id := int(a.ID)
	if old := w.filed[id]; old >= 0 {
		cell := w.cells[old]
		for i, o := range cell {
			if o == a {
				w.cells[old] = append(cell[:i], cell[i+1:]...)
				break
			}
		}
	}
	cell := w.cells[c]
	i := len(cell)
	for i > 0 && cell[i-1].ID > a.ID {
		i--
	}
	cell = append(cell, nil)
	copy(cell[i+1:], cell[i:])
	cell[i] = a
	w.cells[c] = cell
	w.filed[id] = c
}

// Freeze says whether the file may be put right by readers. It is frozen
// while agents decide side by side.
func (w *World) Freeze(frozen bool) { w.frozen = frozen }

// Find returns the agent with the given id, or nil.
func (w *World) Find(id entity.ID) *entity.Agent {
	w.whole()
	if id < 0 || int(id) >= len(w.byID) {
		return nil
	}
	return w.byID[id]
}

// Nearby visits every agent within radius of p, in ascending ID, until
// visit returns false. The radius may not exceed a chunk's side, so that
// the nine chunks around p hold everyone within it.
func (w *World) Nearby(p entity.Pos, radius int, visit func(*entity.Agent) bool) {
	if radius > ChunkSide {
		panic("world: Nearby past a chunk's side")
	}
	w.whole()
	g := w.Grid
	p = g.Norm(p)
	cx, cy := p.X/ChunkSide, p.Y/ChunkSide
	var around [9]int
	n := 0
	for dy := -1; dy <= 1; dy++ {
		y := cy + dy
		if y < 0 || y >= g.CH {
			continue
		}
		for dx := -1; dx <= 1; dx++ {
			x := cx + dx
			if g.Wrap {
				x = ((x % g.CW) + g.CW) % g.CW
			} else if x < 0 || x >= g.CW {
				continue
			}
			c := y*g.CW + x
			dup := false
			for _, seen := range around[:n] {
				if seen == c {
					dup = true
				}
			}
			if !dup {
				around[n] = c
				n++
			}
		}
	}
	if !w.frozen {
		for _, c := range around[:n] {
			w.heal(c)
		}
	}
	// A merge of the lists, smallest ID first.
	var head [9]int
	for {
		best, bestID := -1, entity.ID(0)
		for k := 0; k < n; k++ {
			cell := w.cells[around[k]]
			if head[k] < len(cell) && (best < 0 || cell[head[k]].ID < bestID) {
				best, bestID = k, cell[head[k]].ID
			}
		}
		if best < 0 {
			return
		}
		a := w.cells[around[best]][head[best]]
		head[best]++
		if g.Dist(p, a.Pos) <= radius && !visit(a) {
			return
		}
	}
}

// heal refiles anyone in chunk c who is no longer standing in it.
func (w *World) heal(c int) {
	cell := w.cells[c]
	for i := 0; i < len(cell); {
		a := cell[i]
		if at := int32(w.Grid.ChunkOf(w.Grid.Index(a.Pos))); at != int32(c) {
			w.refile(a, at)
			cell = w.cells[c]
			continue
		}
		i++
	}
}

// Neighbor returns the closest other agent within radius tiles of a, or nil.
// Ties go to the earliest born, which keeps runs deterministic.
func (w *World) Neighbor(a *entity.Agent, radius int) *entity.Agent {
	var best *entity.Agent
	bestD := radius + 1
	w.Nearby(a.Pos, radius, func(o *entity.Agent) bool {
		if o == a {
			return true
		}
		if d := w.Grid.Dist(a.Pos, o.Pos); d < bestD {
			best, bestD = o, d
		}
		return true
	})
	return best
}

// AgentAt returns the agent closest to p within radius, ignoring except.
// Ties go to the earliest born, which keeps runs deterministic.
func (w *World) AgentAt(p entity.Pos, radius int, except *entity.Agent) *entity.Agent {
	var best *entity.Agent
	bestD := radius + 1
	w.Nearby(p, radius, func(o *entity.Agent) bool {
		if o == except {
			return true
		}
		if d := w.Grid.Dist(p, o.Pos); d < bestD {
			best, bestD = o, d
		}
		return true
	})
	return best
}

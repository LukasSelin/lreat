package world

import "lreat/core/entity"

// Where everybody is. Every question about the people near a place used
// to walk the whole population, and most of a tick is such questions, so
// the cost of a settlement went as the square of it. The world files each
// agent under the cell of ground it stands in, and a question about a
// place reads the cells that place's radius actually touches.
//
// A cell is not a chunk. A chunk is a reading of the ground, sixty-four
// tiles across, sized so that a question about the whole map is a sum over
// few of them. A cell is sized to the questions asked about people, the
// widest of which reaches twelve tiles, so that asking who is near
// somebody reads the ground around them and not the district they live in.
// Filed by chunk, a crowd all standing round one market answered every
// such question with the whole crowd, and the square came back: on two
// thousand people this file was sixty-nine per cent of a tick.
//
// What has to be kept is the order. Every one of those questions breaks
// its ties by taking the earliest agent in the population, and the
// population is in order of birth, which is order of ID; and the draw a
// visitor makes among the people it finds indexes the list it found them
// in. So each cell's file is in ascending ID, and a walk over the cells in
// reach is a merge of sorted lists - the same people, in the same order,
// as the walk over everybody would have found.
//
// The file is kept by the one thing that moves an agent between ticks, and
// by whoever moves one by hand saying so with Moved. A reader still puts
// right anyone it finds filed under ground they are no longer standing on,
// but it can only put right what it looks at, and it now looks at very
// little: a position set by hand and not declared is a position the file
// does not know about. It was covered before by accident - on a small map
// the nine chunks around anywhere were the whole map - and the accident is
// gone with the chunks.

// CellSide is how many tiles a cell of the agent file is across and down.
// Nearly every radius anyone is looked for over is this or shorter, so a
// question reads a handful of cells at most.
const CellSide = 8

// NearbyLimit is the furthest anyone may be looked for in one question. It
// is not a property of the file - the cells in reach are worked out from
// the radius, whatever it is - but a bound on how many of them there can
// be, which is what lets the merge run out of fixed arrays and allocate
// nothing.
const NearbyLimit = 20

// maxCells is how many cells a question at NearbyLimit can touch: six each
// way, a span of forty-one tiles laid across cells of eight at the worst
// offset.
const maxCells = 36

// index is the file: who is in each cell, and which cell each is in.
type index struct {
	cells [][]*entity.Agent // by cell, each in ascending ID
	filed []int32           // by ID: the cell the agent is filed under, or -1
	byID  []*entity.Agent   // by ID, nil for the dead and the unborn
	// cw and ch are how many cells the map is across and down, as it was
	// when the file was last laid out.
	cw, ch int
	// counted is how many were on file when it was last known to be whole.
	// A population that has changed size under it is refiled before it is
	// read, which is how one struck off by hand is found to be gone.
	counted int
	// frozen is set while agents decide side by side, when the file may be
	// read from several goroutines and so must not be put right by any.
	frozen bool
}

// cellDims is how many cells across and down the map is now.
func (w *World) cellDims() (int, int) {
	g := w.Grid
	return (g.W + CellSide - 1) / CellSide, (g.H + CellSide - 1) / CellSide
}

// cellOf is the cell the agent standing at p is filed under.
func (w *World) cellOf(p entity.Pos) int {
	p = w.Grid.Norm(p)
	return (p.Y/CellSide)*w.cw + p.X/CellSide
}

// Reindex files the whole population afresh.
func (w *World) Reindex() {
	w.cw, w.ch = w.cellDims()
	if len(w.cells) != w.cw*w.ch {
		w.cells = make([][]*entity.Agent, w.cw*w.ch)
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

// laid reports whether the file is laid out for the map as it stands.
func (w *World) laid() bool {
	cw, ch := w.cellDims()
	return w.cw == cw && w.ch == ch && len(w.cells) == cw*ch
}

// whole refiles the population if it has changed size since it was filed.
func (w *World) whole() {
	if !w.frozen && (!w.laid() || w.counted != len(w.Agents)) {
		w.Reindex()
	}
}

// file puts a where it stands, at the end of that cell's list. The
// population is filed in ID order, so the end is the right place.
func (w *World) file(a *entity.Agent) {
	id := int(a.ID)
	for len(w.byID) <= id {
		w.byID = append(w.byID, nil)
		w.filed = append(w.filed, -1)
	}
	w.byID[id] = a
	w.counted++
	c := w.cellOf(a.Pos)
	w.cells[c] = append(w.cells[c], a)
	w.filed[id] = int32(c)
}

// enroll files a newcomer, already the last of the population. A file that
// has never been laid out for this map is laid out whole instead.
func (w *World) enroll(a *entity.Agent) {
	if !w.laid() || w.counted != len(w.Agents)-1 {
		w.Reindex()
		return
	}
	w.file(a)
}

// Moved refiles a after its position has changed. Cheap when it is still
// in the cell it was filed under, which is nearly always.
func (w *World) Moved(a *entity.Agent) {
	id := int(a.ID)
	if id >= len(w.filed) || !w.laid() {
		// Not on file, or on a file laid out for another map. Laying it
		// out afresh files this one where it now stands with everybody
		// else; only somebody the population has never heard of is left
		// to file alone.
		w.whole()
		if id >= len(w.filed) {
			w.file(a)
		}
		return
	}
	c := int32(w.cellOf(a.Pos))
	if w.filed[id] == c {
		return
	}
	w.refile(a, c)
}

// refile moves a from the cell it is filed under to cell c, keeping both
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
// while a pass reads the world on several goroutines at once: putting the
// file right is a write, and a reader that did it would be writing where
// the others are looking.
//
// Freezing puts it right first, so that what the readers are held to is a
// file that answers for the population as it now stands. Whoever unfreezes
// hands the world back as they found it.
func (w *World) Freeze(frozen bool) {
	if frozen {
		w.whole()
	}
	w.frozen = frozen
}

// Find returns the agent with the given id, or nil.
func (w *World) Find(id entity.ID) *entity.Agent {
	w.whole()
	if id < 0 || int(id) >= len(w.byID) {
		return nil
	}
	return w.byID[id]
}

// reach is a cell a question touches, and how far off the nearest ground
// in it is. A search for the closest person walks these in order and stops
// once what it has found is nearer than any cell left can be.
type reach struct {
	cell int
	off  int
}

// inReach lists the cells holding every tile within radius of p into out,
// and returns how many. It heals what it lists, so a reader that goes on
// to walk them finds anybody filed under ground they have left.
func (w *World) inReach(p entity.Pos, radius int, out *[maxCells]reach) int {
	if radius > NearbyLimit {
		panic("world: nobody is looked for that far")
	}
	w.whole()
	g := w.Grid
	p = g.Norm(p)
	y0, y1 := p.Y-radius, p.Y+radius
	if y0 < 0 {
		y0 = 0
	}
	if y1 > g.H-1 {
		y1 = g.H - 1
	}
	x0, x1 := p.X-radius, p.X+radius
	if !g.Wrap {
		if x0 < 0 {
			x0 = 0
		}
		if x1 > g.W-1 {
			x1 = g.W - 1
		}
	}
	n := 0
	for cy := y0 / CellSide; cy <= y1/CellSide; cy++ {
		for cx := floorDiv(x0, CellSide); cx <= floorDiv(x1, CellSide); cx++ {
			x := cx
			if g.Wrap {
				x = ((x % w.cw) + w.cw) % w.cw
			} else if x < 0 || x >= w.cw {
				continue
			}
			c := cy*w.cw + x
			off := w.cellReach(p, c)
			if off > radius {
				continue
			}
			seen := false
			for k := 0; k < n; k++ {
				if out[k].cell == c {
					seen = true
					break
				}
			}
			if seen {
				continue
			}
			out[n] = reach{cell: c, off: off}
			n++
		}
	}
	if !w.frozen {
		for k := 0; k < n; k++ {
			w.heal(out[k].cell)
		}
	}
	return n
}

// floorDiv divides rounding down, so that a cell west of the map's west
// edge numbers -1 and not 0.
func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// cellReach is how far p is from the nearest tile of cell c, zero when p
// stands in it. Distances east and west are asked of the grid, so the way
// round a globe counts where it is the shorter one.
func (w *World) cellReach(p entity.Pos, c int) int {
	g := w.Grid
	cx, cy := c%w.cw, c/w.cw
	x0, y0 := cx*CellSide, cy*CellSide
	x1, y1 := min(x0+CellSide-1, g.W-1), min(y0+CellSide-1, g.H-1)
	dy := 0
	if p.Y < y0 {
		dy = y0 - p.Y
	} else if p.Y > y1 {
		dy = p.Y - y1
	}
	dx := 0
	if p.X < x0 || p.X > x1 {
		dx = min(g.Dist(p, entity.Pos{X: x0, Y: p.Y}), g.Dist(p, entity.Pos{X: x1, Y: p.Y}))
	}
	return max(dx, dy)
}

// Nearby visits every agent within radius of p, in ascending ID, until
// visit returns false. The radius may not exceed NearbyLimit.
func (w *World) Nearby(p entity.Pos, radius int, visit func(*entity.Agent) bool) {
	var around [maxCells]reach
	n := w.inReach(p, radius, &around)
	g := w.Grid
	p = g.Norm(p)
	// A merge of the lists, smallest ID first.
	var head [maxCells]int
	for {
		best, bestID := -1, entity.ID(0)
		for k := 0; k < n; k++ {
			cell := w.cells[around[k].cell]
			if head[k] < len(cell) && (best < 0 || cell[head[k]].ID < bestID) {
				best, bestID = k, cell[head[k]].ID
			}
		}
		if best < 0 {
			return
		}
		a := w.cells[around[best].cell][head[best]]
		head[best]++
		if g.Dist(p, a.Pos) <= radius && !visit(a) {
			return
		}
	}
}

// closest is the agent nearest p within radius that keep accepts, ties to
// the earliest born. It walks the cells in reach nearest ground first and
// stops as soon as what it holds is nearer than any cell left can be,
// which on crowded ground is after one or two of them. Asking who is
// standing near somebody should cost what the ground around them holds,
// not what the radius does: it is a question about a neighbour, and it was
// being answered by counting the parish.
func (w *World) closest(p entity.Pos, radius int, keep func(*entity.Agent) bool) *entity.Agent {
	var around [maxCells]reach
	n := w.inReach(p, radius, &around)
	// Nearest ground first. There are a handful of cells, so the sort is
	// an insertion, and equal cells keep the order they were listed in.
	for i := 1; i < n; i++ {
		r := around[i]
		j := i
		for j > 0 && around[j-1].off > r.off {
			around[j] = around[j-1]
			j--
		}
		around[j] = r
	}
	g := w.Grid
	p = g.Norm(p)
	var best *entity.Agent
	bestD := radius + 1
	for k := 0; k < n; k++ {
		// Everyone left stands at least this far off, so nobody left can
		// be nearer than what is in hand, nor stand as near and be older.
		if around[k].off > bestD {
			break
		}
		for _, o := range w.cells[around[k].cell] {
			d := g.Dist(p, o.Pos)
			if d > radius || d > bestD {
				continue
			}
			if d == bestD && (best == nil || o.ID > best.ID) {
				continue
			}
			if !keep(o) {
				continue
			}
			best, bestD = o, d
		}
	}
	return best
}

// heal refiles anyone in cell c who is no longer standing in it.
func (w *World) heal(c int) {
	cell := w.cells[c]
	for i := 0; i < len(cell); {
		a := cell[i]
		if at := int32(w.cellOf(a.Pos)); at != int32(c) {
			w.refile(a, at)
			cell = w.cells[c]
			continue
		}
		i++
	}
}

// Closest is the agent nearest p within radius that keep accepts, or nil.
// Ties go to the earliest born. It is what every question of the form "the
// nearest somebody who..." should be asked through: walking everyone in
// the radius and keeping the best of them costs what the radius holds,
// which on crowded ground is most of a settlement, where this costs what
// the ground between the two of them holds.
func (w *World) Closest(p entity.Pos, radius int, keep func(*entity.Agent) bool) *entity.Agent {
	return w.closest(p, radius, keep)
}

// ClosestOf is Closest among one kind of creature: the nearest of that
// species within radius that keep accepts, or nil. Every question of the
// form "the nearest somebody who..." is asked of a kind, because what a
// person wants of a neighbour a deer is not, and a deer standing in the
// crowd would otherwise be counted as company, judged as a witness, and
// drawn as a stranger to go and meet.
func (w *World) ClosestOf(p entity.Pos, radius int, sp *entity.Species, keep func(*entity.Agent) bool) *entity.Agent {
	return w.closest(p, radius, func(o *entity.Agent) bool { return o.Species() == sp && keep(o) })
}

// NearbyOf is Nearby among one kind of creature.
func (w *World) NearbyOf(p entity.Pos, radius int, sp *entity.Species, visit func(*entity.Agent) bool) {
	w.Nearby(p, radius, func(o *entity.Agent) bool {
		if o.Species() != sp {
			return true
		}
		return visit(o)
	})
}

// Neighbor returns the closest other agent of a's own kind within radius
// tiles of a, or nil. Ties go to the earliest born, which keeps runs
// deterministic.
func (w *World) Neighbor(a *entity.Agent, radius int) *entity.Agent {
	return w.ClosestOf(a.Pos, radius, a.Species(), func(o *entity.Agent) bool { return o != a })
}

// AgentAt returns the agent closest to p within radius of except's kind,
// ignoring except itself; with nobody excepted it is the nearest person.
// Ties go to the earliest born, which keeps runs deterministic.
func (w *World) AgentAt(p entity.Pos, radius int, except *entity.Agent) *entity.Agent {
	sp := entity.Human
	if except != nil {
		sp = except.Species()
	}
	return w.ClosestOf(p, radius, sp, func(o *entity.Agent) bool { return o != except })
}

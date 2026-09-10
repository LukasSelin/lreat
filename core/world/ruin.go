package world

import "lreat/core/ontology"

// The working memory of what nobody is left to keep.
//
// The day's ruin is two walks over the ground rather than one: what becomes
// of each tile is worked out on goroutines, and then what was found is gone
// through one tile at a time, in tile order, where the world's chance is
// drawn and the ground is razed. This is what the first walk hands the
// second. It is kept on the world between ticks so that the ruin allocates
// nothing, and what is in it belongs to whichever walk last filled it.
//
// See system.Upkeep, where the reasons are, and parallel.go for the rule
// the first walk keeps.
type Ruin struct {
	// Befall is what becomes of the tile at each index, nil where nothing
	// does. Only the entries Found names are worth reading.
	Befall []*ontology.Transform
	// Found is the tiles something becomes of, gathered by the chunk they
	// are in - a chunk to a goroutine, so no two workers write the same
	// list - and ascending within each.
	Found [][]int32
	// Order is those lists in one, in tile order, which is the order the
	// ground is walked in and so the order the chance is drawn in.
	Order []int32
	// Admit is which chunks the walk looks at, settled before the first
	// walk so that both are held to the same answer.
	Admit []bool
}

// Ruin is the day's ruin's working memory, sized to the ground as it stands.
func (w *World) Ruin() *Ruin {
	r := &w.ruin
	if len(r.Befall) != len(w.Grid.Tiles) {
		r.Befall = make([]*ontology.Transform, len(w.Grid.Tiles))
	}
	if len(r.Found) != len(w.Grid.Chunks) {
		r.Found = make([][]int32, len(w.Grid.Chunks))
		r.Admit = make([]bool, len(w.Grid.Chunks))
	}
	return r
}

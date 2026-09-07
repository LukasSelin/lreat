package system

import "lreat/core/world"

// Climate turns the year. It runs first in the tick, before anything reads
// the temperature: what the land grows this tick and what a body loses to
// the air both depend on it, and both should be answering today's weather
// rather than yesterday's. It is not called Weather because Grid.Weather is
// already the wearing away of roads.
func Climate(w *world.World) {
	w.Climate.Advance(w.Tick, w.RNG)
}

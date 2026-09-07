package system

import (
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/world"
)

// Discovery is a technology waiting to be noticed. Knowledge is the raw
// threshold; Condition is the social shape the settlement must have. Nobody
// chooses these. They happen to a city that has become a certain kind of place.
type Discovery struct {
	Tech      world.Tech
	Knowledge float64
	Condition func(w *world.World) bool
	Effect    func(w *world.World)
	Text      string
}

// skilled counts agents with at least level in a skill.
func skilled(w *world.World, s entity.Skill, level float64) int {
	n := 0
	for _, a := range w.Agents {
		if a.Skills[s] >= level {
			n++
		}
	}
	return n
}

// Discoveries is the catalog, checked in order every tick.
var Discoveries = []Discovery{
	{
		Tech: "agriculture", Knowledge: 15,
		Condition: func(w *world.World) bool { return skilled(w, entity.Farming, 0.2) >= 2 },
		Effect:    func(w *world.World) { w.Mods.FarmYield *= 1.8 },
		Text:      "farmers learned to rotate their fields",
	},
	{
		Tech: "masonry", Knowledge: 40,
		Condition: func(w *world.World) bool { return skilled(w, entity.Building, 0.3) >= 2 },
		Effect: func(w *world.World) {
			w.Mods.BuildEfficiency *= 1.6
			w.Mods.ShelterDecay *= 0.5
		},
		Text: "builders began working in stone",
	},
	{
		Tech: "writing", Knowledge: 90,
		Condition: func(w *world.World) bool { return skilled(w, entity.Scholarship, 0.3) >= 3 },
		Effect:    func(w *world.World) { w.Mods.StudyRate *= 2 },
		Text:      "scholars started keeping written records",
	},
	{
		Tech: "metallurgy", Knowledge: 200,
		Condition: func(w *world.World) bool {
			return w.Has("writing") && skilled(w, entity.Crafting, 0.4) >= 2 && w.Market.Stock[entity.Tools] >= 5
		},
		Effect: func(w *world.World) {
			w.Mods.CraftQuality *= 2
			w.Mods.FarmYield *= 1.3
		},
		Text: "smiths learned to work metal",
	},
}

// Discover unlocks any technology whose conditions the world now meets.
func Discover(w *world.World) {
	for _, d := range Discoveries {
		if w.Has(d.Tech) || w.Knowledge < d.Knowledge {
			continue
		}
		if d.Condition != nil && !d.Condition(w) {
			continue
		}
		w.Unlock(d.Tech)
		d.Effect(w)
		w.Emit(event.Discovered, 0, 0, "%s: %s", d.Tech, d.Text)
	}
}

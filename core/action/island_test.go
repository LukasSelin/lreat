package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/world"
)

// Nothing anybody does in a day reads or writes further from where they
// stand than world.IslandReach: that is what lets people out of each
// other's reach act at the same time. Every radius an act looks or works
// within is held to it here, at the furthest anybody's mind stretches it,
// so that a new act that looks further has to say so where the islands
// are cut.
func TestNothingAnActDoesReachesPastAnIsland(t *testing.T) {
	farthest := entity.Mind{Plasticity: 1, Resolve: 1, Horizon: 1.3}
	radii := map[string]int{
		"ranging":           ranging(&entity.Agent{Mind: farthest}),
		"searchRadius":      searchRadius,
		"pavingRadius":      pavingRadius,
		"companionRadius":   companionRadius,
		"movingRadius":      movingRadius,
		"plantRadius":       plantRadius,
		"granaryRadius":     granaryRadius,
		"meetRadius":        meetRadius,
		"placeRadius":       placeRadius,
		"tavernRadius":      tavernRadius,
		"settlementRadius":  settlementRadius,
		"timberReach":       timberReach,
		"scoutRange":        scoutRange,
		"waterReach":        waterReach,
		"reachRadius":       reachRadius,
		"alarmRadius":       alarmRadius,
		"herdRadius":        herdRadius,
		"fleeRange":         fleeRange,
		"roamStep":          2 * roamStep,
		"world.Window":      world.Window,
		"world.NearbyLimit": world.NearbyLimit,
	}
	for name, r := range radii {
		if r > world.IslandReach {
			t.Errorf("%s = %d reaches past an island (%d); islands are cut for nothing further", name, r, world.IslandReach)
		}
	}
	if ranging(&entity.Agent{Mind: farthest}) <= searchRadius {
		t.Fatal("the farthest mind does not look further than the ordinary one; the bound above tests less than it should")
	}
}

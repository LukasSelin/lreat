package world

import "testing"

// Every rock a history can make has to turn up on a map made by one, and none
// of them may take the whole thing over. The two that decide this are granite
// and shale, because they are the two that had nowhere to come from when the
// generator was first written: granite was the leftover case - ground nothing
// ever happened to - which on sixteen epochs is no ground at all, and shale
// lost every basin to sandstone because a fill's make-up hardly varies while
// a history is running. Granite is the root of an arc now, laid bare, and the
// coarse and the fine fill are ranked against each other rather than against
// a fixed line.
func TestEveryRockAHistoryMakesTurnsUp(t *testing.T) {
	pooled := map[Bedrock]int{}
	tiles := 0
	for _, seed := range []uint64{1, 2, 3, 4, 5} {
		w := NewWith(seed, Ancient())
		var seen [BedrockCount]int
		for i := range w.Grid.Tiles {
			seen[w.Grid.Tiles[i].Bedrock]++
			pooled[w.Grid.Tiles[i].Bedrock]++
			tiles++
		}
		for _, b := range Bedrocks() {
			if seen[b] == 0 {
				t.Errorf("seed %d has no %s on it at all", seed, b)
			}
		}
	}
	// Pooled over the seeds, no rock may be more than half a world or less
	// than a fortieth of one. The band is wide because which rocks a world
	// gets is the whole point - a world with little ocean floor has little
	// granite, and that is a fact about the world and not a fault in it.
	for _, b := range Bedrocks() {
		share := float64(pooled[b]) / float64(tiles)
		if share > 0.5 || share < 0.025 {
			t.Errorf("%s is %.1f%% of the ground pooled over five worlds", b, 100*share)
		}
	}
}

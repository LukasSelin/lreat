package world

import (
	"testing"

	"lreat/core/ontology"
)

// The world and the ontology keep one calendar. What grows is stated in the
// ontology and the weather that advances it is stated here, so if the two
// ever came apart a season of growing would stop being a season of weather.
func TestTheWorldKeepsTheOntologysCalendar(t *testing.T) {
	if Year != ontology.Year || Season != ontology.Season {
		t.Fatalf("world year %d/%d, ontology year %d/%d", Year, Season, ontology.Year, ontology.Season)
	}
	if Season != 25 {
		t.Fatalf("a season is %d ticks; the crop stages are written for 25", Season)
	}
}

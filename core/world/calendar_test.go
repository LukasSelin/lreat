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
	// A crop is the fastest thing that grows and a wood the slowest, and
	// both are read in growing days off this same calendar. The crop is a
	// month rather than a season: what a stand takes to come on kept its
	// length in days when the year lengthened, because the day is what the
	// land renews by and the day is what people take by. See
	// ontology.Processes.
	if crop, wood := ontology.Crop.Full(), ontology.Timbering.Full(); crop >= wood {
		t.Fatalf("a crop comes on in %.0f days and a wood in %.0f; the crop is the fast one", crop, wood)
	}
	if crop := ontology.Crop.Full(); crop > float64(Season) {
		t.Fatalf("a crop takes %.0f days, longer than the %d-day season it is sown and cut within", crop, Season)
	}
}

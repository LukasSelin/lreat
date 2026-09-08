package ontology_test

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"testing"

	"lreat/core/action"
	"lreat/core/belief"
	"lreat/core/habit"
	"lreat/core/ontology"
)

// golden maps every instance the trees should entail to the hand-written
// action it stands in for.
var golden = map[string]string{
	"take/berries@wood":                 "forage",
	"take/game@wood":                    "hunt",
	"take/timber@wood":                  "gather wood",
	"take/fish@water":                   "fish",
	"take/stone@outcrop":                "quarry",
	"take/grain@field":                  "farm",
	"tend/clear@open":                   "clear field",
	"tend/water@field":                  "irrigate",
	"tend/plant@open":                   "plant trees",
	"make/timber>tool@bench":            "craft",
	"make/stone+timber>tool@forge":      "smelt",
	"make/provision+timber>meal@hearth": "cook",
	"raise/timber>dwelling@open":        "build shelter",
	"raise/timber+stone>granary@open":   "build granary",
	"raise/timber>tavern@open":          "build tavern",
	"raise/timber+stone>market@open":    "found market",
	"raise/timber>road@ground":          "lay road",
	"consume/provision":                 "eat",
	"dwell/rest":                        "rest",
	"dwell/meet@tavern>neighbour":       "socialize",
	"dwell/look":                        "scout",
	"dwell/guard@market":                "guard",
	"exchange/material>coin@market":     "sell",
	"exchange/coin>provision@market":    "buy food",
	"transfer/provision>needy":          "give",
	"transfer/provision>child":          "feed child",
	"transfer/material>requester":       "fulfil request",
	"transfer/provision<holder":         "steal",
	"pass/practice>pupil":               "teach",
	"pass/practice>child":               "teach child",
	"pass/practice>self":                "study",
	"strike/person>wrongdoer":           "retaliate",
	"move@dwelling":                     "move house",
}

// TestInstantiateMatchesCatalog is the golden test: the trees entail
// exactly the acts the hand-written catalog has, and nothing else.
func TestInstantiateMatchesCatalog(t *testing.T) {
	got := map[string]bool{}
	for _, in := range ontology.Instantiate() {
		if got[in.Key] {
			t.Errorf("duplicate key %q", in.Key)
		}
		got[in.Key] = true
	}
	for k := range golden {
		if !got[k] {
			t.Errorf("missing %q", k)
		}
	}
	for k := range got {
		if _, ok := golden[k]; !ok {
			t.Errorf("unexpected %q", k)
		}
	}
	covered := map[string]bool{}
	for _, name := range golden {
		covered[name] = true
	}
	for _, d := range action.Catalog {
		if !covered[d.Name] {
			t.Errorf("catalog action %q has no instance", d.Name)
		}
	}
}

// TestInstantiateIsDeterministic: two walks are the same walk.
func TestInstantiateIsDeterministic(t *testing.T) {
	a, b := ontology.Instantiate(), ontology.Instantiate()
	if len(a) != len(b) {
		t.Fatalf("lengths differ: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i].Key != b[i].Key || a[i].Prior != b[i].Prior {
			t.Errorf("instance %d differs between walks", i)
		}
	}
	if !sort.SliceIsSorted(a, func(i, j int) bool { return a[i].Key < a[j].Key }) {
		t.Error("instances are not sorted by key")
	}
}

// TestDerivedPriorsAgree reports how far the composed priors are from the
// hand-tuned ones, act by act. A low fit is not a failure of the ontology
// but a disagreement to look at: either the composition is missing a
// class or trait, or the hand prior was carrying something the act's kind
// does not explain. The floor is where the composition stands today; raise
// it as the trees improve.
func TestDerivedPriorsAgree(t *testing.T) {
	const floor = 0.75
	type row struct {
		key, name string
		fit       float64
	}
	var rows []row
	for _, in := range ontology.Instantiate() {
		name := golden[in.Key]
		d := action.ByName(name)
		if d == nil {
			continue
		}
		fit := habit.Cosine(in.Prior, d.Tuned)
		rows = append(rows, row{in.Key, name, fit})
		if fit < floor {
			t.Errorf("%-34s vs %-14s fit %.2f\n  derived %s\n  hand    %s",
				in.Key, name, fit, show(in.Prior), show(d.Tuned))
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].fit < rows[j].fit })
	for _, r := range rows {
		t.Logf("%.2f  %-34s %s", r.fit, r.key, r.name)
	}
}

// TestDerivedValencesAgree does the same for the moral weights. The hand
// table leaves most of the later crafts unlisted, which reads as an
// omission rather than a judgement that quarrying is idle, so an act with
// no entry there is reported and not failed.
func TestDerivedValencesAgree(t *testing.T) {
	for _, in := range ontology.Instantiate() {
		name := golden[in.Key]
		want := action.ValenceOf(name)
		if want == (belief.Valence{}) && name != "eat" && name != "buy food" {
			t.Logf("%-34s vs %-14s derived %v, hand has no entry", in.Key, name, showV(in.Valence))
			continue
		}
		var diff float64
		for i := range want {
			diff = math.Max(diff, math.Abs(want[i]-in.Valence[i]))
		}
		if diff > 0.05 {
			t.Errorf("%-34s vs %-14s valence differs by %.2f: derived %v hand %v",
				in.Key, name, diff, showV(in.Valence), showV(want))
		}
	}
}

// TestReachAndTicksAgree: what the instance inherits from the schema is
// what the hand-written act had.
func TestReachAndTicksAgree(t *testing.T) {
	for _, in := range ontology.Instantiate() {
		name := golden[in.Key]
		d := action.ByName(name)
		if in.Ticks != d.Ticks {
			t.Errorf("%s ticks %d, %s has %d", in.Key, in.Ticks, name, d.Ticks)
		}
		if math.Abs(in.Reach0-d.Reach0) > 1e-9 {
			t.Errorf("%s reach %.2f, %s has %.2f", in.Key, in.Reach0, name, d.Reach0)
		}
	}
}

func show(s habit.Signature) string {
	var parts []string
	for i, v := range s {
		if v != 0 {
			parts = append(parts, habit.Names[i]+":"+trim(v))
		}
	}
	return "{" + strings.Join(parts, " ") + "}"
}

func showV(v belief.Valence) string {
	var parts []string
	for i, x := range v {
		if x != 0 {
			parts = append(parts, belief.Norm(i).String()+":"+trim(x))
		}
	}
	return "{" + strings.Join(parts, " ") + "}"
}

func trim(v float64) string {
	s := strings.TrimRight(strings.TrimRight(strconv.FormatFloat(v, 'f', 2, 64), "0"), ".")
	if s == "" || s == "-0" {
		return "0"
	}
	return s
}

// The ground that carries something growing says so, and the ground that
// does not says that too: a wood and a field have a standing crop with an age
// to it, an outcrop is stone, and a shoal of fish is a stock rather than a
// crop that has to come on before it can be cut.
func TestWoodsAndFieldsAreLiving(t *testing.T) {
	for _, c := range []*ontology.Class{ontology.Wood, ontology.Field} {
		if !c.Has(ontology.Living) {
			t.Errorf("%s carries a standing crop and is not living", c.Path())
		}
	}
	for _, c := range []*ontology.Class{ontology.Open, ontology.Water, ontology.Outcrop, ontology.Road, ontology.Dwelling, ontology.Market} {
		if c.Has(ontology.Living) {
			t.Errorf("%s is not something that grows", c.Path())
		}
	}
}

// Time, as the ontology has it: how long a thing takes to come on, which is
// a Process, and which half of the year the getting belongs to, which is the
// thing's own Warmth. The two have to stay independent, because the whole
// reason for holding both is that a stand can be slow and of the cold half
// or quick and wholly of the warm one.
func TestTimeIsTwoFacts(t *testing.T) {
	// Slow, and of the cold half: years in the growing, felled in winter.
	if ontology.Timbering.Full() < float64(2*ontology.Year) {
		t.Errorf("timber comes on in %.0f days, want years of it", ontology.Timbering.Full())
	}
	if ontology.Timber.Warmth >= 0 {
		t.Errorf("timber's warmth is %.2f; felling is winter work", ontology.Timber.Warmth)
	}
	// Quick, and wholly of its season.
	if ontology.Crop.Full() >= ontology.Timbering.Full() {
		t.Error("a crop should come on faster than a wood")
	}
	if !(ontology.Grain.Warmth > ontology.Berries.Warmth) {
		t.Error("a crop keeps the season's hours more closely than the brush does")
	}
	// A stock is neither: there or not, with no age to wait out, and no
	// half of the year it belongs to.
	for _, c := range []*ontology.Class{ontology.Fish, ontology.Stone} {
		if len(ontology.Growing(c)) > 0 || c.Warmth != 0 {
			t.Errorf("%s is a stock; it should neither come on nor keep a season", c.Name)
		}
	}
}

// The season an act belongs to is composed from what the act is about, not
// written onto it. Nothing but a class's Warmth may put the chill
// coordinate into a prior, or the concept is decoration.
func TestTheSeasonOfAnActFollowsFromItsThing(t *testing.T) {
	for _, in := range ontology.Instantiate() {
		chill := in.Prior[habit.Chill]
		if chill == 0 {
			continue
		}
		var want float64
		for _, c := range []*ontology.Class{in.Object, in.Schema.Output, in.Schema.Season} {
			if c != nil && c.Warmth != 0 {
				want = -c.Warmth
			}
		}
		if math.Abs(chill-want) > 1e-9 {
			t.Errorf("%s carries chill %.2f, but nothing it is about is worth %.2f", in.Key, chill, want)
		}
	}
}

// Anything the year makes is taken off ground that carries a standing crop,
// and anything taken off such ground is something the year makes. The two
// halves of the same statement live in different places - the process on one
// side, the Living trait on the site on the other - and this is what keeps
// them one statement.
func TestWhatGrowsIsTakenOffLivingGround(t *testing.T) {
	for _, in := range ontology.Instantiate() {
		if in.Schema.Verb != ontology.Take || in.Object == nil || in.Site == nil {
			continue
		}
		// A stand is one stand however many things are had off it: the
		// brush a forager picks is the brush a trapper hunts over, and the
		// ontology names one process for it. So the question is whether
		// anything grows here at all, not whether a process is named for
		// this particular thing.
		grows := len(ontology.Growing(in.Site)) > 0
		if living := in.Site.Has(ontology.Living); grows != living {
			t.Errorf("%s takes %s (grown %v) off %s (living %v); the two should agree",
				in.Key, in.Object.Name, grows, in.Site.Name, living)
		}
	}
}

package system

import (
	"strings"
	"testing"

	"lreat/core/action"
	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/observe"
	"lreat/core/ontology"
	"lreat/core/world"
)

// A deer is an agent like anybody: it wants, it reads its moment, it
// recognises an act as belonging to it, and it walks. What these hold it to
// is the rest of the design - that it lives on the same catalog and the same
// tables as the people without ever crossing into theirs, that it keeps to
// the woods and eats them, that it runs from anybody who comes near, that
// its young are its kind, and that the settlement's books never count it.

// herdWorld founds the valley with a party and a herd in the woods around it.
func herdWorld(seed uint64, people, deer int) *world.World {
	cfg := world.DefaultConfig()
	cfg.Deer = deer
	w := world.NewWith(seed, cfg)
	for i := 0; i < people; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	w.Populate()
	return w
}

func herdOf(w *world.World) []*entity.Agent {
	var out []*entity.Agent
	for _, a := range w.Agents {
		if a.Species() == entity.Deer {
			out = append(out, a)
		}
	}
	return out
}

func TestDeerKeepToTheWoodsAndEatThem(t *testing.T) {
	// Where a herd stands after six years is read over four valleys, not
	// one: a herd that a party happens to spook onto open ground stays
	// out a while, and on any one seed the share under trees swings
	// between two thirds and nearly all of them with nothing changed but
	// which of two ways of the same cost somebody walked.
	inWood, herdSize := 0, 0
	for seed := uint64(1); seed <= 4; seed++ {
		w := herdWorld(seed, 20, 20)
		Run(w, 2000)
		for _, d := range herdOf(w) {
			herdSize++
			if w.Grid.At(d.Pos).Is(ontology.Wood) {
				inWood++
			}
		}
	}
	if inWood*5 < herdSize*4 {
		t.Errorf("%d of %d deer stand under trees over four valleys; deer keep to the woods", inWood, herdSize)
	}
	w := herdWorld(3, 20, 20)
	Run(w, 2000)
	herd := herdOf(w)
	if len(herd) == 0 {
		t.Fatal("the herd died out inside six years")
	}
	// A deer never swims: nothing it has planned takes it through deep
	// water, and none stands in any.
	for _, d := range herd {
		if w.Grid.At(d.Pos).Deep() {
			t.Errorf("%s stands in deep water", d.Name)
		}
		if d.Plan != nil {
			for _, p := range d.Plan.Route {
				if w.Grid.At(p).Deep() {
					t.Errorf("%s is routed through deep water", d.Name)
					break
				}
			}
		}
	}
	browsed := 0
	for _, e := range w.Log.All() {
		if e.Kind == event.Acted && e.Act == "deer:take/browse@wood" {
			browsed++
		}
	}
	if browsed == 0 {
		t.Error("nobody browsed anything in six years")
	}
	// What a herd eats is the brush the people forage: a tile a deer has
	// browsed holds less of it than it held.
	for _, d := range herd {
		if i := w.Grid.Index(d.Pos); w.Grid.At(d.Pos).Is(ontology.Wood) && w.Grid.Wild[i] < 1 {
			return
		}
	}
	t.Error("no deer stands on a stand of brush it has thinned")
}

func TestADeerRunsFromAPerson(t *testing.T) {
	w := world.New(7)
	isWood := func(_ entity.Pos, tile *world.Tile) bool { return tile.Is(ontology.Wood) }
	covert, ok := w.Grid.NearestOfKind(w.MarketPos, 40, world.KindsOf(ontology.Wood), isWood)
	if !ok {
		t.Fatal("no wood near the square")
	}
	doe := w.SpawnKind(entity.Deer, "doe", w.RandomPersonality(), covert)
	// Fed, in company, and already frightened: the one thing pressing is
	// the person two tiles off.
	doe.Needs = need.Levels{0.9, 0.05, 0.9, 1, 0.9}
	var beside entity.Pos
	found := false
	for r := 2; r <= 3 && !found; r++ {
		for dy := -r; dy <= r && !found; dy++ {
			for dx := -r; dx <= r && !found; dx++ {
				p := w.Grid.Norm(entity.Pos{X: covert.X + dx, Y: covert.Y + dy})
				if w.Grid.Dist(p, covert) == r && w.Grid.In(p) && !w.Grid.At(p).Wet() {
					beside, found = p, true
				}
			}
		}
	}
	if !found {
		t.Fatal("nowhere dry beside the covert")
	}
	man := w.SpawnAt("man", w.RandomPersonality(), beside)
	// Standing still for the length of the test, on a plan that finishes
	// long after it.
	man.Plan = &entity.Plan{Action: "rest", Target: beside, Remaining: 500, Total: 500}
	was := w.Grid.Dist(doe.Pos, man.Pos)

	ranked := action.Rank(doe, w)
	if len(ranked) == 0 || ranked[0].Def != action.Flee {
		var names []string
		for _, c := range ranked {
			names = append(names, c.Def.Name)
		}
		t.Fatalf("a frightened deer beside a person ranks %s; want flee first", strings.Join(names, ", "))
	}
	Run(w, 30)
	if now := w.Grid.Dist(doe.Pos, man.Pos); now <= was {
		t.Errorf("after a month the deer stands %d from the man, no further than the %d it started at", now, was)
	}
}

func TestFawnsAreDeer(t *testing.T) {
	w := herdWorld(11, 10, 20)
	Run(w, 2*world.Year)
	var fawn *entity.Agent
	for _, d := range herdOf(w) {
		if d.Born > 0 {
			fawn = d
			break
		}
	}
	if fawn == nil {
		t.Fatal("no fawn born in two years to a herd of twenty")
	}
	if fawn.Norms != (belief.Norms{}) || fawn.Skills != [entity.SkillCount]float64{} || fawn.HasHome || fawn.Wealth != 0 {
		t.Errorf("a fawn holds what a person holds: norms %v skills %v home %v wealth %v", fawn.Norms, fawn.Skills, fawn.HasHome, fawn.Wealth)
	}
	if fawn.Species().Life != entity.Deer.Life {
		t.Error("a fawn does not live a deer's life")
	}
}

// Nobody weighs, seeds, inherits or does an act of another kind: a person's
// tables are empty on a deer's slots and a deer's on a person's, and nothing
// in the record shows either finishing the other's work.
func TestNobodyDoesAnotherKindsAct(t *testing.T) {
	w := herdWorld(5, 20, 20)
	Run(w, 600)
	kind := map[entity.ID]*entity.Species{}
	for _, a := range w.Agents {
		kind[a.ID] = a.Species()
		for _, i := range action.For(entity.Human) {
			if a.Species() == entity.Deer && (a.Habits[i] != (habit.Signature{}) || a.Reach[i] != 0) {
				t.Fatalf("%s holds a habit for %s", a.Name, action.Catalog[i].Name)
			}
		}
		for _, i := range action.For(entity.Deer) {
			if a.Species() == entity.Human && (a.Habits[i] != (habit.Signature{}) || a.Reach[i] != 0) {
				t.Fatalf("%s holds a habit for %s", a.Name, action.Catalog[i].Name)
			}
		}
		if a.Plan != nil && !action.Owns(a.Species(), action.ByName(a.Plan.Action)) {
			t.Fatalf("%s, a %s, is off to %s", a.Name, a.Species().Name, a.Plan.Action)
		}
	}
	for _, e := range w.Log.All() {
		if e.Kind != event.Acted {
			continue
		}
		sp, alive := kind[e.Actor]
		if !alive {
			continue
		}
		if d := action.ByKey(e.Act); d != nil && !action.Owns(sp, d) {
			t.Fatalf("a %s finished %s", sp.Name, e.Act)
		}
	}
}

func TestDeerDecidingInParallelChangesNothing(t *testing.T) {
	run := func(workers int) string {
		was := world.Workers
		world.Workers = workers
		defer func() { world.Workers = was }()
		w := herdWorld(9, 20, 20)
		Run(w, 400)
		return fingerprint(w)
	}
	one := run(1)
	if many := run(8); many != one {
		t.Fatal("a herd deciding over eight workers came out differently from one")
	}
}

// The settlement's books are of its people: its population, its means and
// its funnel count nobody with hooves, and the map still shows every deer.
func TestTheCensusCountsThePeople(t *testing.T) {
	w := herdWorld(2, 15, 10)
	Run(w, 50)
	s := observe.Take(w)
	people, deer := w.People(), len(herdOf(w))
	if s.Population != people || s.Creatures != deer {
		t.Fatalf("census says %d people and %d creatures; there are %d and %d", s.Population, s.Creatures, people, deer)
	}
	marked := 0
	for _, m := range s.Map.Agents {
		if m.Kind == "deer" {
			marked++
		}
	}
	if marked != deer {
		t.Fatalf("%d deer marked on the map of %d", marked, deer)
	}
	gates := 0
	for _, n := range w.Vitals.Gates {
		gates += n
	}
	if gates != people {
		t.Fatalf("the fertility funnel counts %d, the people are %d", gates, people)
	}
}

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

// A creature is an agent like anybody: it wants, it reads its moment, it
// recognises an act as belonging to it, and it walks. What these hold every
// kind to is the rest of the design - that it lives on the same catalog and
// the same tables as the people without ever crossing into theirs or into
// another kind's, that it keeps to its ground and eats it, that it runs from
// anybody who comes near, that its young are its kind, and that the
// settlement's books never count it.

// wildWorld founds the valley with a party and every kind of creature in
// its habitat around it.
func wildWorld(seed uint64, people, each int) *world.World {
	cfg := world.DefaultConfig()
	cfg.Deer, cfg.Boar, cfg.Hare = each, each, each
	w := world.NewWith(seed, cfg)
	for i := 0; i < people; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	w.Populate()
	return w
}

func kindOf(w *world.World, sp *entity.Species) []*entity.Agent {
	var out []*entity.Agent
	for _, a := range w.Agents {
		if a.Species() == sp {
			out = append(out, a)
		}
	}
	return out
}

// Where each kind lives - a deer under trees, a boar under trees or in
// somebody's field, a hare at the edge of the wood - how many in ten are
// found there on any day, and the key of the act it eats by. A deer hardly
// leaves the trees; a boar is forever crossing the open between the wood
// and the fields, and a hare between the hedge and the sward.
var habitats = map[*entity.Species]struct {
	ground func(g *world.Grid, p entity.Pos) bool
	inTen  int
	feeds  string
}{
	entity.Deer: {func(g *world.Grid, p entity.Pos) bool { return g.At(p).Is(ontology.Wood) }, 8, "deer:take/browse@wood"},
	entity.Boar: {func(g *world.Grid, p entity.Pos) bool {
		t := g.At(p)
		return t.Is(ontology.Wood) || t.Is(ontology.Field) || (t.Is(ontology.Open) && besideWood(g, p))
	}, 6, "boar:take/mast@wood"},
	entity.Hare: {func(g *world.Grid, p entity.Pos) bool {
		t := g.At(p)
		return t.Is(ontology.Wood) || t.Is(ontology.Open)
	}, 8, "hare:take/sward@open"},
}

func besideWood(g *world.Grid, p entity.Pos) bool {
	return g.HasNeighbor(p, func(t *world.Tile) bool { return t.Is(ontology.Wood) })
}

func TestCreaturesKeepToTheirGroundAndEatIt(t *testing.T) {
	// Where a kind stands after six years is read over four valleys, not
	// one: a herd that a party happens to spook onto open ground stays
	// out a while, and on any one seed the share on its ground swings
	// with nothing changed but which of two ways of the same cost
	// somebody walked.
	at, of := map[*entity.Species]int{}, map[*entity.Species]int{}
	for seed := uint64(1); seed <= 4; seed++ {
		w := wildWorld(seed, 20, 20)
		Run(w, 2000)
		for _, sp := range entity.Creatures {
			for _, a := range kindOf(w, sp) {
				of[sp]++
				if habitats[sp].ground(w.Grid, a.Pos) {
					at[sp]++
				}
			}
		}
	}
	for _, sp := range entity.Creatures {
		if h := habitats[sp]; at[sp]*10 < of[sp]*h.inTen {
			t.Errorf("%d of %d %s stand on their ground over four valleys; want %d in ten", at[sp], of[sp], sp.Name, h.inTen)
		}
	}
	w := wildWorld(3, 20, 20)
	Run(w, 2000)
	fed := map[string]int{}
	for _, e := range w.Log.All() {
		if e.Kind == event.Acted {
			fed[e.Act]++
		}
	}
	for _, sp := range entity.Creatures {
		kind := kindOf(w, sp)
		if len(kind) == 0 {
			t.Errorf("the %s died out inside six years", sp.Name)
			continue
		}
		h := habitats[sp]
		for _, a := range kind {
			// No creature swims: none stands in deep water and none is
			// routed through any.
			if w.Grid.At(a.Pos).Deep() {
				t.Errorf("%s stands in deep water", a.Name)
			}
			if a.Plan != nil {
				for _, p := range a.Plan.Route {
					if w.Grid.At(p).Deep() {
						t.Errorf("%s is routed through deep water", a.Name)
						break
					}
				}
			}
		}
		if fed[h.feeds] == 0 {
			t.Errorf("no %s ever fed by %s", sp.Name, h.feeds)
		}
	}
	// What a herd eats is the brush the people forage, and what a warren
	// eats is the sward: somewhere a creature stands on ground it has
	// thinned.
	thinned := false
	for _, a := range w.Agents {
		i := w.Grid.Index(a.Pos)
		switch a.Species() {
		case entity.Deer, entity.Boar:
			thinned = thinned || (w.Grid.At(a.Pos).Is(ontology.Wood) && w.Grid.Wild[i] < 1)
		case entity.Hare:
			thinned = thinned || (w.Grid.At(a.Pos).Is(ontology.Open) && w.Grid.Sward[i] < 1)
		}
	}
	if !thinned {
		t.Error("no creature stands on ground it has thinned")
	}
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

// A boar in a strip in ear eats the crop and tramples what it does not eat:
// the strip is worn and the crop set back.
func TestABoarRaidsAField(t *testing.T) {
	w := world.New(4)
	g := w.Grid
	// A strip beside the square, broken and in ear.
	strip, ok := g.Nearest(w.MarketPos, 20, func(p entity.Pos, t *world.Tile) bool { return t.Buildable() })
	if !ok {
		t.Fatal("nowhere to break a strip")
	}
	g.Turn(strip, world.Field)
	i := g.Index(strip)
	g.Age[i] = ontology.Crop.Full()
	g.Fertility[i], g.Rich[i] = 0.8, 0.8
	boar := w.SpawnKind(entity.Boar, "boar", w.RandomPersonality(), strip)
	boar.Needs = need.Levels{0.2, 1, 1, 1, 1} // hungry, and nothing else
	if p, ok := action.Raid.Target(boar, w); !ok || p != strip {
		t.Fatalf("a hungry boar on a strip in ear does not see the raid: %v %v", p, ok)
	}
	fed := boar.Needs[need.Physiological]
	action.Raid.Apply(boar, w)
	if g.Fertility[i] >= 0.8 || g.Age[i] >= ontology.Crop.Full() {
		t.Errorf("after a raid the strip holds %.2f and its crop is %.0f days on; want both less", g.Fertility[i], g.Age[i])
	}
	if boar.Needs[need.Physiological] <= fed {
		t.Error("the boar got nothing out of the raid")
	}
}

func TestYoungAreTheirParentsKind(t *testing.T) {
	w := wildWorld(11, 10, 20)
	Run(w, 2*world.Year)
	for _, sp := range entity.Creatures {
		var young *entity.Agent
		for _, a := range kindOf(w, sp) {
			if a.Born > 0 {
				young = a
				break
			}
		}
		if young == nil {
			t.Errorf("no %s born in two years to twenty", sp.Name)
			continue
		}
		if young.Norms != (belief.Norms{}) || young.Skills != [entity.SkillCount]float64{} || young.HasHome || young.Wealth != 0 {
			t.Errorf("a young %s holds what a person holds: norms %v skills %v home %v wealth %v", sp.Name, young.Norms, young.Skills, young.HasHome, young.Wealth)
		}
		if young.Species().Life != sp.Life {
			t.Errorf("a young %s does not live its kind's life", sp.Name)
		}
	}
}

// Nobody weighs, seeds, inherits or does an act of another kind: every
// kind's tables are empty on every other kind's slots, and nothing in the
// record shows any finishing another's work.
func TestNobodyDoesAnotherKindsAct(t *testing.T) {
	w := wildWorld(5, 20, 12)
	Run(w, 600)
	kinds := append([]*entity.Species{entity.Human}, entity.Creatures...)
	kind := map[entity.ID]*entity.Species{}
	for _, a := range w.Agents {
		kind[a.ID] = a.Species()
		for _, other := range kinds {
			if other == a.Species() {
				continue
			}
			for _, i := range action.For(other) {
				if a.Habits[i] != (habit.Signature{}) || a.Reach[i] != 0 {
					t.Fatalf("%s, a %s, holds a habit for %s", a.Name, a.Species().Name, action.Catalog[i].Name)
				}
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

func TestCreaturesDecidingInParallelChangeNothing(t *testing.T) {
	run := func(workers int) string {
		was := world.Workers
		world.Workers = workers
		defer func() { world.Workers = was }()
		w := wildWorld(9, 20, 12)
		Run(w, 400)
		return fingerprint(w)
	}
	one := run(1)
	if many := run(8); many != one {
		t.Fatal("the creatures deciding over eight workers came out differently from one")
	}
}

// The settlement's books are of its people: its population, its means and
// its funnel count nobody with hooves or paws, and the map still shows
// every creature by its kind.
func TestTheCensusCountsThePeople(t *testing.T) {
	w := wildWorld(2, 15, 6)
	Run(w, 50)
	s := observe.Take(w)
	people, wild := w.People(), len(w.Agents)-w.People()
	if s.Population != people || s.Creatures != wild {
		t.Fatalf("census says %d people and %d creatures; there are %d and %d", s.Population, s.Creatures, people, wild)
	}
	marked := map[string]int{}
	for _, m := range s.Map.Agents {
		marked[m.Kind]++
	}
	for _, sp := range entity.Creatures {
		if marked[sp.Name] != len(kindOf(w, sp)) {
			t.Fatalf("%d %s marked on the map of %d", marked[sp.Name], sp.Name, len(kindOf(w, sp)))
		}
	}
	gates := 0
	for _, n := range w.Vitals.Gates {
		gates += n
	}
	if gates != people {
		t.Fatalf("the fertility funnel counts %d, the people are %d", gates, people)
	}
}

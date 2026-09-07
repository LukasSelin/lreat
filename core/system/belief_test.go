package system

import (
	"testing"

	"lreat/core/action"
	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/world"
)

// desperate puts an agent one step from starving, with a well-fed neighbor
// standing next to it, and returns both.
func desperate(t *testing.T, seed uint64) (*world.World, *entity.Agent, *entity.Agent) {
	t.Helper()
	w := valueWorld(seed)
	thief := w.Spawn("thief", need.Neutral())
	victim := w.SpawnAt("victim", need.Neutral(), thief.Pos)
	thief.Needs = need.Levels{0.05, 0.9, 0.9, 0.9, 0.9}
	thief.Inventory[entity.Food] = 0
	victim.Inventory[entity.Food] = 5
	// Start everyone morally blank so each test isolates the one value it is
	// about, rather than depending on what the seed happened to draw.
	thief.Norms = belief.Norms{}
	victim.Norms = belief.Norms{}
	return w, thief, victim
}

func TestConscienceRestrainsTheft(t *testing.T) {
	// The same agent in the same predicament, differing only in what it holds
	// to be right, must make different choices. That difference is the whole
	// point of the moral layer.
	w, honest, _ := desperate(t, 5)
	honest.Norms[belief.Honesty] = 1
	if d, _ := Choose(honest, w); d == action.Steal {
		t.Fatal("an agent who values honesty stole rather than starve slowly")
	}

	w, crooked, _ := desperate(t, 5)
	crooked.Norms[belief.Honesty] = 0.05
	if d, _ := Choose(crooked, w); d != action.Steal {
		t.Fatalf("an agent with no scruples chose %q over an easy theft", d.Name)
	}
}

func TestHungerEventuallyOverwhelmsPrinciple(t *testing.T) {
	// Conscience is a weight, not a gate. Push the need far enough and it
	// gives way, which is what keeps the pyramid leaky in both directions.
	w, a, _ := desperate(t, 6)
	a.Norms[belief.Honesty] = 0.55
	a.Needs[need.Physiological] = 0.9
	if d, _ := Choose(a, w); d == action.Steal {
		t.Fatal("a comfortable agent stole")
	}
	a.Needs[need.Physiological] = 0
	if d, _ := Choose(a, w); d != action.Steal {
		t.Fatalf("a starving agent still chose %q over theft", d.Name)
	}
}

func TestTheftCostsReputationAndPeace(t *testing.T) {
	w, thief, victim := desperate(t, 7)
	thief.Norms[belief.Honesty] = 0
	bystander := w.SpawnAt("bystander", need.Neutral(), thief.Pos)
	bystander.Norms = belief.Norms{belief.Honesty: 1}

	thief.Pos = victim.Pos
	action.Steal.Apply(thief, w)

	if victim.Inventory[entity.Food] != 4 || thief.Inventory[entity.Food] != 1 {
		t.Fatalf("food did not move: victim=%.1f thief=%.1f",
			victim.Inventory[entity.Food], thief.Inventory[entity.Food])
	}
	if victim.Regard(thief.ID) >= 0 {
		t.Fatalf("victim's regard for the thief is %.2f, want negative", victim.Regard(thief.ID))
	}
	if bystander.Regard(thief.ID) >= 0 {
		t.Fatalf("an honest witness's regard is %.2f, want negative", bystander.Regard(thief.ID))
	}
}

func TestWitnessesJudgeByTheirOwnValues(t *testing.T) {
	w, thief, _ := desperate(t, 8)
	strict := w.SpawnAt("strict", need.Neutral(), thief.Pos)
	strict.Norms = belief.Norms{belief.Honesty: 1}
	lax := w.SpawnAt("lax", need.Neutral(), thief.Pos)
	lax.Norms = belief.Norms{belief.Honesty: 0}

	action.Steal.Apply(thief, w)

	if strict.Regard(thief.ID) >= lax.Regard(thief.ID) {
		t.Fatalf("the strict witness judged %.2f, the lax one %.2f; the strict should judge harder",
			strict.Regard(thief.ID), lax.Regard(thief.ID))
	}
}

func TestSelfDoubtProducesARequest(t *testing.T) {
	w := valueWorld(11)
	a := w.Spawn("asker", need.Neutral())
	w.Spawn("other", need.Neutral())
	// Wants shelter, has money, and believes it cannot build.
	a.Needs = need.Levels{0.9, 0.1, 0.9, 0.9, 0.9}
	a.Shelter = 0
	a.Wealth = 20
	a.Efficacy[entity.Building] = 0.01

	r, ok := compose(a, w)
	if !ok {
		t.Fatal("an agent who needs shelter and cannot build asked nobody")
	}
	if r.Kind != entity.Serve || r.Skill != entity.Building {
		t.Fatalf("asked for the wrong thing: %+v", r)
	}
	if r.Reward <= 0 || r.Reward > a.Wealth {
		t.Fatalf("reward %.2f is not payable from wealth %.2f", r.Reward, a.Wealth)
	}
}

func TestCapableAgentsDoNotAskUntilTheyAreRich(t *testing.T) {
	w := valueWorld(12)
	a := w.Spawn("builder", need.Neutral())
	w.Spawn("other", need.Neutral())
	a.Needs = need.Levels{0.9, 0.1, 0.9, 0.9, 0.9}
	a.Shelter = 0
	a.Wealth = 5
	a.Efficacy[entity.Building] = 0.9

	if _, ok := compose(a, w); ok {
		t.Fatal("a confident builder of modest means hired someone else")
	}

	// Wealth changes the calculation: time now costs more than the fee.
	a.Wealth = 40
	if _, ok := compose(a, w); !ok {
		t.Fatal("a rich confident builder still would not pay to have it done")
	}
}

func TestRequestersHireWhoTheyBelieveIsBest(t *testing.T) {
	w := valueWorld(13)
	a := w.Spawn("asker", need.Neutral())
	good := w.Spawn("believed good", need.Neutral())
	bad := w.Spawn("believed bad", need.Neutral())

	// Belief, not truth: the one the asker rates highly is actually useless.
	a.Rate(good.ID, entity.Building, 0.9, w.Tick)
	a.Rate(bad.ID, entity.Building, 0.1, w.Tick)
	good.Skills[entity.Building] = 0
	bad.Skills[entity.Building] = 0.9

	if pick := bestKnown(a, w, entity.Building); pick != good.ID {
		t.Fatalf("hired %d, want the agent the asker believes in (%d)", pick, good.ID)
	}
}

func TestFulfillingWorkPaysAndTeachesBothSides(t *testing.T) {
	w := valueWorld(14)
	client := w.Spawn("client", need.Neutral())
	doer := w.SpawnAt("doer", need.Neutral(), client.Pos)
	client.Wealth = 10
	client.Shelter = 0
	doer.Skills[entity.Building] = 0.8

	w.Post(entity.Request{
		Requester: client.ID, Kind: entity.Serve, Skill: entity.Building,
		Reward: 4, Deadline: w.Tick + 50,
	})

	if !action.Fulfil.Available(doer, w) {
		t.Fatal("a capable neighbour could not see the job")
	}
	action.Fulfil.Apply(doer, w)

	if doer.Wealth != 4 || client.Wealth != 6 {
		t.Fatalf("payment wrong: doer=%.1f client=%.1f", doer.Wealth, client.Wealth)
	}
	if client.Shelter <= 0 {
		t.Fatal("the work had no effect on the client")
	}
	if len(w.Requests) != 0 {
		t.Fatalf("%d requests left on the board, want 0", len(w.Requests))
	}
	if got := client.Competence(doer.ID, entity.Building); got <= 0.15 {
		t.Fatalf("client's belief in the doer is %.2f, want it raised by the evidence", got)
	}
}

func TestUnansweredRequestsSourTheRelationship(t *testing.T) {
	w := valueWorld(15)
	client := w.Spawn("client", need.Neutral())
	named := w.Spawn("named", need.Neutral())
	client.Rate(named.ID, entity.Building, 0.8, w.Tick)

	w.Post(entity.Request{
		Requester: client.ID, Directed: named.ID, Kind: entity.Serve,
		Skill: entity.Building, Reward: 3, Deadline: w.Tick + 1,
	})
	w.Tick += 5
	expire(w)

	if len(w.Requests) != 0 {
		t.Fatal("the expired request stayed on the board")
	}
	if client.Regard(named.ID) >= 0 {
		t.Fatalf("regard for the no-show is %.2f, want negative", client.Regard(named.ID))
	}
	if got := client.Competence(named.ID, entity.Building); got >= 0.8 {
		t.Fatalf("belief in the no-show is still %.2f, want it knocked down", got)
	}
}

func TestConfidenceTrailsCompetence(t *testing.T) {
	w := valueWorld(16)
	a := w.Spawn("a", need.Neutral())
	a.Skills[entity.Farming] = 1
	a.Efficacy[entity.Farming] = 0.15

	Beliefs(w)
	first := a.Efficacy[entity.Farming]
	if first <= 0.15 {
		t.Fatal("doing the work taught the agent nothing about itself")
	}
	if first >= 1 {
		t.Fatal("one tick made the agent fully self-aware; belief should lag")
	}
	for i := 0; i < 200; i++ {
		Beliefs(w)
	}
	if a.Efficacy[entity.Farming] <= first {
		t.Fatal("confidence stopped climbing toward demonstrated skill")
	}
}

func TestRequestsAppearInALivingSettlement(t *testing.T) {
	w := valueWorld(21)
	for i := 0; i < 20; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	Run(w, 3000)

	var asked, done int
	for _, e := range w.Log.All() {
		switch e.Kind {
		case event.Requested:
			asked++
		case event.Fulfilled:
			done++
		}
	}
	if asked == 0 {
		t.Fatal("nobody asked anybody for anything in 3000 ticks")
	}
	if done == 0 {
		t.Fatalf("%d requests posted, none ever fulfilled", asked)
	}
}

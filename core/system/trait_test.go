package system

import (
	"testing"

	"lreat/core/action"
	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/world"
)

// Every measure of a body and a mind is a multiplier on something the world
// does to an agent or the agent does to itself. These say that each of them
// reaches whatever it was meant to reach - a trait nothing reads is a trait
// that is not there, and it would be easy to add one and wire it nowhere.

// two spawns a pair on the same ground with ordinary bodies and minds, so
// that a test can vary one measure and nothing else.
func two(t *testing.T) (*world.World, *entity.Agent, *entity.Agent) {
	t.Helper()
	w := world.New(1)
	a, b := w.Spawn("a", need.Neutral()), w.Spawn("b", need.Neutral())
	a.Body, a.Mind = entity.Ordinary()
	b.Body, b.Mind = entity.Ordinary()
	b.Pos = a.Pos
	return w, a, b
}

// A body that burns more is hungrier at the end of the same day.
func TestABodyThatBurnsMoreIsHungrierSooner(t *testing.T) {
	w, steady, hungry := two(t)
	hungry.Body.Metabolism = 1.3
	for i := 0; i < 20; i++ {
		Decay(w)
	}
	if hungry.Needs[need.Physiological] >= steady.Needs[need.Physiological] {
		t.Fatalf("the body that burns more is at %.4f and the ordinary one at %.4f; it should be lower",
			hungry.Needs[need.Physiological], steady.Needs[need.Physiological])
	}
}

// A hardy body is spared what a frail one pays for standing out in the cold,
// in what it burns staying warm and in the condition it costs.
func TestAHardyBodyFeelsTheColdLess(t *testing.T) {
	w, hardy, frail := two(t)
	hardy.Body.Hardiness, frail.Body.Hardiness = 1.3, 0.7
	hardy.Shelter, frail.Shelter = 0, 0
	// The deep of winter, with nothing over either of them.
	for w.Tick = 0; clock.SeasonOf(w.Tick) != clock.Winter; w.Tick++ {
	}
	w.Climate.Advance(w.Tick, w.RNG)
	if w.ChillAt(hardy.Pos) <= 0 {
		t.Skip("no cold on this ground in this season; the test has nothing to weigh")
	}
	hardy.Needs[need.Physiological], frail.Needs[need.Physiological] = 1, 1
	hardy.Health, frail.Health = 1, 1
	for i := 0; i < 40; i++ {
		Decay(w)
	}
	if frail.Needs[need.Physiological] >= hardy.Needs[need.Physiological] {
		t.Fatalf("the frail body is no hungrier for the winter: %.4f against %.4f",
			frail.Needs[need.Physiological], hardy.Needs[need.Physiological])
	}
	if frail.Health >= hardy.Health {
		t.Fatalf("the frail body is no worse for the winter: %.4f against %.4f", frail.Health, hardy.Health)
	}
}

// A quick mind brings a craft within reach in fewer turns at it. Everyday
// living is in reach from birth and cannot be learned toward, so this has to
// be asked of a gated action - a craft - which is the only kind practice
// moves at all.
func TestAQuickMindLearnsFaster(t *testing.T) {
	craft := -1
	for i, d := range action.Catalog {
		if d.Reach0 < 1 {
			craft = i
			break
		}
	}
	if craft < 0 {
		t.Skip("nothing in the catalog is out of reach to begin with")
	}
	_, quick, slow := two(t)
	quick.Mind.Plasticity, slow.Mind.Plasticity = 1.3, 0.7
	for _, a := range []*entity.Agent{quick, slow} {
		for i := 0; i < 5; i++ {
			action.Practise(a, craft)
		}
	}
	if quick.Reach[craft] <= slow.Reach[craft] {
		t.Fatalf("after the same five turns at %s the quick mind is at %.4f and the slow one at %.4f",
			action.Catalog[craft].Name, quick.Reach[craft], slow.Reach[craft])
	}
}

// A resolute mind settles the same moment at a sharper temperature, which is
// what makes it take the best fit it can see rather than a near second.
func TestAResoluteMindDecidesMoreSharply(t *testing.T) {
	w, firm, wavering := two(t)
	firm.Mind.Resolve, wavering.Mind.Resolve = 1.3, 0.7
	firm.Needs, wavering.Needs = need.Levels{0.5, 0.5, 0.5, 0.5, 0.5}, need.Levels{0.5, 0.5, 0.5, 0.5, 0.5}
	if Temper(w, firm) >= Temper(w, wavering) {
		t.Fatalf("the resolute mind decides at %.4f and the wavering one at %.4f; the first should be the sharper",
			Temper(w, firm), Temper(w, wavering))
	}
}

// An agent put together by hand, with nothing drawn for it, lives like an
// ordinary one rather than like a corpse. Every reading of a trait goes
// through the fallback that makes this so, and a test that built an agent
// literally would otherwise get a body that burns nothing and never tires.
func TestAnAgentWithNothingDrawnIsOrdinary(t *testing.T) {
	var a entity.Agent
	if got := a.Body.Burn(); got != 1 {
		t.Fatalf("an undrawn body burns %v, want the ordinary 1", got)
	}
	if got := a.Mind.Decides(); got != 1 {
		t.Fatalf("an undrawn mind decides at %v, want the ordinary 1", got)
	}
	if got := a.Endurance(0); got <= 0 {
		t.Fatalf("an undrawn body endures %v; it should carry an ordinary frame", got)
	}
}

// A founding party is not one body and one mind repeated, and what its
// people are is handed to their children rather than drawn afresh.
func TestBodiesAndMindsAreDrawnApartAndHandedOn(t *testing.T) {
	w := world.New(3)
	var burn []float64
	for i := 0; i < 30; i++ {
		burn = append(burn, w.Spawn("a", w.RandomPersonality()).Body.Metabolism)
	}
	same := 0
	for _, v := range burn {
		if v == burn[0] {
			same++
		}
	}
	if same == len(burn) {
		t.Fatal("every founder burns exactly alike; the party is one body repeated")
	}
	// A child is its parent drifted, not a stranger: over many children the
	// mean sits nearer the parent than the ordinary measure does.
	parent := w.Agents[0]
	parent.Body.Metabolism = 1.25
	var sum float64
	const children = 200
	for i := 0; i < children; i++ {
		sum += w.InheritBody(parent.Body).Metabolism
	}
	mean := sum / children
	if d := mean - parent.Body.Metabolism; d > 0.05 || d < -0.05 {
		t.Fatalf("children of a body at %.2f average %.2f; they are not its children",
			parent.Body.Metabolism, mean)
	}
}

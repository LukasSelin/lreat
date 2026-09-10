package action

import (
	"fmt"
	"math"
	"math/rand/v2"
	"testing"

	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/world"
)

// blank spawns an agent with neutral personality, blank norms, and a middling
// temperament, so each test controls the one thing it is about.
func blank(w *world.World, name string) *entity.Agent {
	a := w.Spawn(name, need.Neutral())
	a.Norms = belief.Norms{0.5, 0.5, 0.5, 0.5}
	a.Temperament = entity.Temperament{Trust: 0.5, Warmth: 0.5}
	return a
}

func TestStrangersStartDifferentlyByTemperament(t *testing.T) {
	w := world.New(1)
	trusting := blank(w, "trusting")
	wary := blank(w, "wary")
	stranger := blank(w, "stranger")
	trusting.Temperament.Trust = 0.9
	wary.Temperament.Trust = 0.1

	rt := Introduce(trusting, stranger, w.Tick).Regard
	rw := Introduce(wary, stranger, w.Tick).Regard
	if rt <= rw {
		t.Fatalf("trusting agent's first regard %.2f should exceed wary agent's %.2f", rt, rw)
	}
}

func TestAffinityShapesFirstImpression(t *testing.T) {
	w := world.New(2)
	a := blank(w, "a")
	kin := blank(w, "kin")
	alien := blank(w, "alien")
	a.Norms = belief.Norms{1, 1, 0, 0}
	kin.Norms = belief.Norms{1, 1, 0, 0}
	alien.Norms = belief.Norms{0, 0, 1, 1}

	if FirstImpression(a, kin) <= FirstImpression(a, alien) {
		t.Fatal("an agent should warm to someone who shares its values over someone who does not")
	}
}

func TestGossipSeedsOpinionsOfPeopleNeverMet(t *testing.T) {
	w := world.New(3)
	listener := blank(w, "listener")
	speaker := blank(w, "speaker")
	villain := blank(w, "villain")
	speaker.Judge(villain.ID, -0.9, w.Tick)
	listener.AddBond(speaker.ID, 0.8) // a close friend's word carries

	if listener.Look(villain.ID) != nil {
		t.Fatal("listener should start with no opinion of the villain")
	}
	Gossip(listener, speaker, w)

	b := listener.Look(villain.ID)
	if b == nil {
		t.Fatal("gossip left the listener with no opinion of the villain")
	}
	if b.Met != 0 {
		t.Fatal("hearsay must not count as a meeting")
	}
	if b.Regard >= 0 || b.Expect >= 0 {
		t.Fatalf("hearsay should sour the listener: regard %.2f expect %.2f", b.Regard, b.Expect)
	}

	// The same news lands more softly on a distrustful listener.
	skeptic := blank(w, "skeptic")
	skeptic.Temperament.Trust = 0.1
	skeptic.AddBond(speaker.ID, 0.8)
	Gossip(skeptic, speaker, w)
	if skeptic.Regard(villain.ID) <= listener.Regard(villain.ID) {
		t.Fatalf("skeptic %.2f should be moved less than the trusting listener %.2f",
			skeptic.Regard(villain.ID), listener.Regard(villain.ID))
	}
}

func TestMeetingsUpdateExpectationsTowardWhatHappened(t *testing.T) {
	w := world.New(4)
	a := blank(w, "a")
	friend := blank(w, "friend")
	foe := blank(w, "foe")
	a.Norms = belief.Norms{1, 1, 1, 1}
	friend.Norms = belief.Norms{1, 1, 1, 1}
	foe.Norms = belief.Norms{0, 0, 0, 0}
	friend.Personality = a.Personality
	foe.Personality = need.Weights{2, 2, 0.3, 0.3, 2}

	var qFriend, qFoe float64
	for i := 0; i < 5; i++ {
		qFriend += Encounter(a, friend, w)
		qFoe += Encounter(a, foe, w)
	}
	if qFriend <= qFoe {
		t.Fatalf("meetings with a like mind (%.2f) should go better than with an opposite (%.2f)", qFriend, qFoe)
	}
	bf, bo := a.Look(friend.ID), a.Look(foe.ID)
	if bf.Expect <= bo.Expect {
		t.Fatalf("expectation of friend %.2f should exceed foe %.2f", bf.Expect, bo.Expect)
	}
	if bf.Strength <= bo.Strength {
		t.Fatalf("bond with friend %.2f should exceed foe %.2f", bf.Strength, bo.Strength)
	}
	if bf.Met != 5 || bo.Met != 5 {
		t.Fatalf("meetings not counted: %d and %d", bf.Met, bo.Met)
	}
}

func TestAgentsSeekTheCompanyTheyExpectToEnjoy(t *testing.T) {
	w := world.New(5)
	a := blank(w, "a")
	liked := blank(w, "liked")
	disliked := blank(w, "disliked")
	a.Norms[belief.Tradition] = 1 // never explores, so the choice is pure expectation
	// Put the disliked one closer so distance alone would favor them.
	disliked.Pos = a.Pos
	liked.Pos = entity.Pos{X: a.Pos.X + 6, Y: a.Pos.Y}
	bl := a.Know(liked.ID, w.Tick)
	bl.Expect, bl.Met = 0.8, 1
	bd := a.Know(disliked.ID, w.Tick)
	bd.Expect, bd.Met = -0.6, 1

	if got := PickCompany(a, w); got != liked {
		t.Fatalf("picked %s, want the friend six tiles away", got.Name)
	}

	// And the expected value of the visit reflects who it is.
	eLiked := Socialize.Expect(a, w, liked.Pos)[need.Belonging]
	eDisliked := Socialize.Expect(a, w, disliked.Pos)[need.Belonging]
	if eLiked <= eDisliked {
		t.Fatalf("expected belonging from a friend %.3f should exceed a foe %.3f", eLiked, eDisliked)
	}
}

func TestWarmthChangesWhatCompanyIsWorth(t *testing.T) {
	w := world.New(6)
	warm := blank(w, "warm")
	cold := blank(w, "cold")
	other := blank(w, "other")
	warm.Temperament.Warmth = 0.9
	cold.Temperament.Warmth = 0.1
	if Socialize.Expect(warm, w, other.Pos)[need.Belonging] <= Socialize.Expect(cold, w, other.Pos)[need.Belonging] {
		t.Fatal("a warm agent should expect more from company than a cold one")
	}
}

// byTheWholeCrowd is how the most welcome company used to be found: every
// person within reach, in order of birth, the best kept with a strict
// better-than so that ties fell to the elder.
func byTheWholeCrowd(a *entity.Agent, w *world.World) *entity.Agent {
	var best *entity.Agent
	bestScore := math.Inf(-1)
	w.Nearby(a.Pos, meetRadius, func(o *entity.Agent) bool {
		if o == a {
			return true
		}
		if s := Anticipate(a, o) - 0.01*float64(w.Grid.Dist(a.Pos, o.Pos)); s > bestScore {
			best, bestScore = o, s
		}
		return true
	})
	return best
}

// preferred looks at the people it knows and the nearest it does not,
// instead of at everybody standing near. It has to land on the same person
// the whole crowd would have given - not merely somebody as welcome, but
// the same one, since who is visited decides who befriends whom.
//
// The crowd is built by hand rather than lived into being, so that the
// cases that could tell the two apart are all present: people standing on
// one another, so that ties on welcome and on distance both happen; some
// known and most not; bonds ranging from loathing to devotion; and some
// bonds to people who have wandered out of reach or died, which the short
// way has to skip and the long way never sees.
func TestPreferredIsWhoTheWholeCrowdWouldGive(t *testing.T) {
	w := world.New(5)
	rng := rand.New(rand.NewPCG(19, 20))
	var all []*entity.Agent
	for i := 0; i < 80; i++ {
		a := blank(w, "a")
		a.Pos = entity.Pos{X: 20 + rng.IntN(9), Y: 8 + rng.IntN(9)}
		w.Moved(a)
		a.Temperament.Trust = rng.Float64()
		all = append(all, a)
	}
	// A few standing well outside everybody's reach, to be bonded to and
	// never met.
	for i := 0; i < 6; i++ {
		a := blank(w, "far")
		a.Pos = entity.Pos{X: 60 + rng.IntN(5), Y: 30}
		w.Moved(a)
		all = append(all, a)
	}
	for _, a := range all {
		for k := 0; k < rng.IntN(6); k++ {
			o := all[rng.IntN(len(all))]
			if o == a {
				continue
			}
			b := entity.Bond{To: o.ID, Met: rng.IntN(3)}
			b.Regard = rng.Float64()*2 - 1
			b.Expect = rng.Float64()*2 - 1
			a.Bonds = append(a.Bonds, b)
		}
	}
	// A bond to somebody who is not in the world at all.
	all[0].Bonds = append(all[0].Bonds, entity.Bond{To: entity.ID(9999), Met: 1, Expect: 1})
	for _, a := range all {
		want := byTheWholeCrowd(a, w)
		if got := preferred(a, w); got != want {
			t.Fatalf("%d would visit %v; the whole crowd says %v", a.ID, id(got), id(want))
		}
	}
}

func id(a *entity.Agent) string {
	if a == nil {
		return "nobody"
	}
	return fmt.Sprint(a.ID)
}

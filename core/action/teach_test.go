package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/world"
)

// pair puts two blank agents on the same tile, which is what standing close
// enough to teach amounts to.
func pair(w *world.World) (*entity.Agent, *entity.Agent) {
	a, o := blank(w, "teacher"), blank(w, "pupil")
	o.Pos = a.Pos
	return a, o
}

func TestNobodyTeachesWhatTheOtherAlreadyKnows(t *testing.T) {
	w := world.New(1)
	a, o := pair(w)
	a.Skills[entity.Farming] = 0.6
	o.Skills[entity.Farming] = 0.6
	if Teach.Available(a, w) {
		t.Fatal("two people at the same level have nothing to pass, and an afternoon of it should not be on offer")
	}
	o.Skills[entity.Farming] = 0
	if !Teach.Available(a, w) {
		t.Fatal("a journeyman standing next to a beginner should have a lesson to give")
	}
}

func TestALessonCarriesLessFromAWorseTeacher(t *testing.T) {
	// Over twenty worlds and not one. What a lesson carries depends on the
	// habits of the two people in it, and those come off the world's own
	// stream after the ground has been drawn from it - so one world seed is
	// one draw. Counted, this held on eighteen of twenty; on the other two the
	// apprentice had nothing to give, which is a thing that can happen to an
	// apprentice and not a fault in the rule.
	held := 0
	const worlds = 20
	for seed := uint64(1); seed <= worlds; seed++ {
		w := world.New(seed)
		master, keen := pair(w)
		master.Skills[entity.Farming] = entity.Mastery
		Teach.Apply(master, w)
		fromMaster := keen.Skills[entity.Farming]

		w2 := world.New(seed)
		amateur, other := pair(w2)
		amateur.Skills[entity.Farming] = 2 * entity.TierBand
		Teach.Apply(amateur, w2)
		fromAmateur := other.Skills[entity.Farming]

		if fromAmateur > 0 && fromAmateur < fromMaster {
			held++
		}
		if fromAmateur > fromMaster {
			t.Fatalf("seed %d: an hour with an apprentice gave %.4f and an hour with a master %.4f; the master should be worth more",
				seed, fromAmateur, fromMaster)
		}
	}
	if held < 15 {
		t.Fatalf("an apprentice was worth something, and less than a master, on %d worlds of %d", held, worlds)
	}
}

func TestALifetimeOfLessonsIsNotMastery(t *testing.T) {
	w := world.New(3)
	master, pupil := pair(w)
	master.Skills[entity.Farming] = entity.Mastery
	// Every day of a working life at the master's elbow.
	for i := 0; i < 20000; i++ {
		Teach.Apply(master, w)
	}
	if got := pupil.Skills[entity.Farming]; got > entity.TaughtCeiling+1e-9 {
		t.Fatalf("a life of lessons came to %.3f, past the %.2f being shown can reach", got, entity.TaughtCeiling)
	}
	if !Teach.Available(master, w) {
		return // the lesson correctly ran out
	}
	t.Fatal("a pupil at the ceiling should have nothing left to be taught")
}

package entity

import "lreat/core/need"

// How a person gets good at something.
//
// Skill runs from nothing to mastery in [0,1], and it used to be raised the
// same amount by every act that touched it: a flat step, from whatever
// source, at whatever level. That made competence a thing the settlement
// accumulated rather than a thing anybody earned. A newcomer stood next to a
// farmer for a season and was a farmer; the farmer stood next to a third
// person and so was that one; and because being shown cost nothing and
// carried as far as doing, the whole settlement rose together at the rate of
// its best member. Nothing about it got harder the further it went.
//
// Two things are true of learning anything and neither was here.
//
// The first is that it curves. The step from knowing nothing to being some
// use is short and the step from being good to being the best there is takes
// most of a life, and the same hour of work buys a great deal of the first
// and very little of the last. So the gain is scaled by the tier the learner
// is standing in, and each tier is slower than the one under it.
//
// The second is that being shown only carries so far. A teacher hands over
// what can be said and demonstrated, and there is always a part that cannot
// be - the part that is only got by having done the thing badly enough times.
// So learning has a ceiling that depends on where it came from: being taught
// stops short of the teacher and well short of mastery, reading stops sooner
// than that, and only the work itself goes all the way up.
//
// Together they mean a settlement can raise a generation of competent people
// quickly and cannot raise a generation of masters at all. Masters are made
// one at a time, by working, and when one dies what they knew goes down a
// tier and has to be climbed again.

// Tier is the band of competence a level stands in. The four are equal
// quarters of the range, so a tier is a quarter and TierOf is a division.
type Tier int

const (
	// Novice has been shown the thing and is no good at it.
	Novice Tier = iota
	// Apprentice can do it under supervision and is worth feeding.
	Apprentice
	// Journeyman can do it alone and can teach the first two tiers.
	Journeyman
	// Master is as good as the settlement has, and got there by working.
	Master
	TierCount
)

var tierNames = [TierCount]string{"novice", "apprentice", "journeyman", "master"}

func (t Tier) String() string { return tierNames[t] }

// TierBand is how much of the range one tier covers.
const TierBand = 1.0 / float64(TierCount)

// TierOf is the tier a level stands in. Mastery itself is the master tier
// rather than a fifth one off the end.
func TierOf(level float64) Tier {
	t := Tier(level / TierBand)
	if t >= TierCount {
		t = TierCount - 1
	}
	if t < 0 {
		t = 0
	}
	return t
}

// tierRate is what a tier costs, as a share of what the novice tier costs.
// The first is left at one so that the opening years of a settlement - the
// years it is short of everything and has nobody to spare - are exactly as
// hard as they were. What has changed is the far end, where the same act
// that carried a beginner a quarter of the way now barely moves somebody who
// is already good. All four together make mastery by work about three times
// the labour a flat rate made it.
var tierRate = [TierCount]float64{1, 0.55, 0.3, 0.15}

// Mastery is the top of the range and the ceiling on learning by doing.
const Mastery = 1.0

// The ceilings. What a person can reach depends on where the learning came
// from, and only one of these ways reaches the top.
const (
	// TaughtGap is how far short of the teacher a lesson leaves the pupil.
	// A teacher cannot hand over the whole of what they have: what is left
	// at the end is the part that was never sayable.
	TaughtGap = 0.2
	// TaughtCeiling is where being shown stops however good the teacher is.
	// It sits exactly on the threshold of the master tier: the best lessons
	// there are carry a pupil to the door of mastery and never a step
	// through it, so every part of the last tier is worked for. That is the
	// whole of the point - the last tier is the one nobody can be given.
	TaughtCeiling = 3 * TierBand
	// StudyCeiling is where reading on your own stops. It is the top of the
	// apprentice tier, which is exactly the level the settlement asks for
	// before it will let somebody teach: books carry a scholar to the door
	// of the work and no further, and the work carries them the rest.
	StudyCeiling = 2 * TierBand
	// TeachFloor is the level below which a person has nothing to show. It
	// is the top of the novice tier: you must be out of it yourself before
	// anybody is better off standing next to you.
	TeachFloor = TierBand
)

// Learn is what doing the work teaches. It is the only way to mastery, and
// the only thing that counts as having done it: every call is one more time
// this pair of hands has been on the thing, whether or not the skill had any
// room left to rise.
func (a *Agent) Learn(s Skill, d float64) float64 {
	a.Practice[s]++
	return a.toward(s, d, Mastery)
}

// Study is what reading alone teaches: the same curve, stopped early.
func (a *Agent) Study(s Skill, d float64) float64 { return a.toward(s, d, StudyCeiling) }

// Taught is what being shown teaches, and returns how much was actually
// passed - which may be nothing, and often is.
//
// The teacher's level tells twice over. It sets how far the lessons can
// carry, by TaughtGap under the teacher and never past TaughtCeiling; and it
// sets how fast they carry, in proportion, so that an hour with somebody who
// has just cleared the floor is worth a fraction of an hour with a master.
// That is the difference the old flat step had no way of saying: two
// amateurs teaching each other went up as quickly as a master teaching
// either of them, and rather more often, because there were more of them.
func (a *Agent) Taught(s Skill, d, teacher float64) float64 {
	if teacher < TeachFloor {
		return 0
	}
	return a.toward(s, d*teacher, min(teacher-TaughtGap, TaughtCeiling))
}

// CanBeTaught reports whether this person has anything to gain from being
// shown the skill by someone at that level. It is what keeps a lesson with
// nothing in it from being worth anybody's afternoon; see action.Teach.
func (a *Agent) CanBeTaught(s Skill, teacher float64) bool {
	return teacher >= TeachFloor && a.Skills[s] < min(teacher-TaughtGap, TaughtCeiling)
}

// toward raises a skill by d, curved by the tier it is raised from and
// stopped at a ceiling. The rate is read at the level the gain starts from,
// which is exact for the small steps every caller actually takes and would
// not be for a step that crossed a tier in one go; nothing raises a skill by
// a quarter at a time.
func (a *Agent) toward(s Skill, d, ceiling float64) float64 {
	was := a.Skills[s]
	if d <= 0 || was >= ceiling {
		return 0
	}
	got := min(d*tierRate[TierOf(was)], ceiling-was)
	a.Skills[s] = need.Clamp(was + got)
	return a.Skills[s] - was
}

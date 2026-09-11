package system

import (
	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/world"
)

const (
	// ConfidenceRate is how fast doing a thing teaches you what you can do.
	// Slow, so belief lags reality and the gap has time to matter.
	ConfidenceRate = 0.02
	// ConfidenceFloor keeps a shred of self-belief alive in everyone, so an
	// agent who has never tried something can still be talked into it.
	ConfidenceFloor = 0.05
	// TalkRadius is how close agents must be for values to rub off.
	TalkRadius = 3
)

// Beliefs keeps what agents think roughly in touch with what is true.
//
// Self-belief drifts toward demonstrated skill, never arriving: people who
// work at something grow confident a little after they grow capable, and
// people who stop doing it stay confident a little after they stop.
func Beliefs(w *world.World) {
	for _, a := range w.Agents {
		for s := range a.Efficacy {
			a.Efficacy[s] = belief.Clamp(belief.Update(a.Efficacy[s], a.Skills[s], ConfidenceRate))
			if a.Efficacy[s] < ConfidenceFloor {
				a.Efficacy[s] = ConfidenceFloor
			}
		}
		// Deterrence is a memory, and memories fade. A settlement stays
		// orderly only while somebody keeps answering wrongs.
		a.Caution = belief.Clamp(a.Caution - belief.CautionDecay)
	}
	contagion(w)
}

// contagion spreads values between neighbors. Nobody argues anybody into
// anything; people simply come to resemble those they stand next to, and
// districts end up with characters their founders never chose.
//
// Who is standing next to whom is looked up for everybody at once, over
// goroutines, and the values are spread afterwards one agent at a time in
// agent order. The lookup only reads where people stand, which nothing here
// moves, so it answers the same side by side as in turn; the spreading is
// not so - what an agent takes on is read off a neighbour who may already
// have taken something on today - and it stays in the one order it always
// had. Looking was nearly the whole cost of the day's beliefs on a crowded
// map, and the spreading is a few multiplications each.
func contagion(w *world.World) {
	n := len(w.Agents)
	near := make([]*entity.Agent, n)
	w.Freeze(true)
	world.InParallel(n, world.WorkersOver(n), func(i, _ int) {
		near[i] = w.Neighbor(w.Agents[i], TalkRadius)
	})
	w.Freeze(false)
	for i, a := range w.Agents {
		o := near[i]
		if o == nil {
			continue
		}
		// You take on the values of people you like and admire, and drift
		// away from those you have judged badly.
		weight := 1.0
		if b := a.Look(o.ID); b != nil {
			weight = 0.4 + b.Strength + b.Regard
		}
		if weight <= 0 {
			continue
		}
		belief.Spread(&a.Norms, o.Norms, weight)
	}
}

// MeanNorms is the settlement's average values, for observation and tests.
func MeanNorms(w *world.World) belief.Norms {
	var m belief.Norms
	if len(w.Agents) == 0 {
		return m
	}
	for _, a := range w.Agents {
		for i := range m {
			m[i] += a.Norms[i]
		}
	}
	for i := range m {
		m[i] /= float64(len(w.Agents))
	}
	return m
}

// Outcast reports agents the settlement has turned against: those whose
// average regard among the people who know them has gone negative.
func Outcast(w *world.World) []entity.ID {
	sum := map[entity.ID]float64{}
	count := map[entity.ID]int{}
	for _, a := range w.Agents {
		for i := range a.Bonds {
			sum[a.Bonds[i].To] += a.Bonds[i].Regard
			count[a.Bonds[i].To]++
		}
	}
	var out []entity.ID
	for _, a := range w.Agents { // fixed order, not map order
		if n := count[a.ID]; n >= 3 && sum[a.ID]/float64(n) < -0.3 {
			out = append(out, a.ID)
		}
	}
	return out
}

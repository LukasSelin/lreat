package action

import (
	"math"

	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// reachRadius is how far an agent will look for a person to act upon.
const reachRadius = 8

// Witness has up to this many bystanders form an opinion about a public act.
const witnessCount = 3

// witness lets nearby agents judge an act by their own values. It is how a
// reputation forms without any central record of who did what.
func witness(a *entity.Agent, w *world.World, name string) {
	v := ValenceOf(name)
	seen := 0
	w.NearbyOf(a.Pos, reachRadius, a.Species(), func(o *entity.Agent) bool {
		if o == a {
			return true
		}
		o.Judge(a.ID, belief.Judgement(v, o.Norms), w.Tick)
		seen++
		return seen < witnessCount
	})
}

// remorse charges the actor for doing something their own values condemn.
func remorse(a *entity.Agent, name string) {
	if g := belief.Guilt(ValenceOf(name), a.Norms); g > 0 {
		a.Needs.Add(need.Esteem, -g)
	}
}

// nearestWith returns the closest other agent satisfying ok, or nil. It
// asks the world for the nearest rather than sizing up everyone in the
// radius, because ok is asked of a stranger and the answer is the same
// whoever asks: the nearest holder of a thing is the nearest holder of it,
// and finding them should not cost the crowd standing behind them.
func nearestWith(a *entity.Agent, w *world.World, radius int, ok func(*entity.Agent) bool) *entity.Agent {
	return w.ClosestOf(a.Pos, radius, a.Species(), func(o *entity.Agent) bool { return o != a && ok(o) })
}

func hasSpareFood(o *entity.Agent) bool { return o.Inventory[entity.Food] >= 1 }

// Steal is the option that makes conscience mean something. It is fast, it
// works, and the only thing standing against it is what the agent believes
// about itself and what it thinks the neighbors will make of it.
var Steal = mover("transfer/provision<holder")

func inNeed(o *entity.Agent) bool {
	return o.Needs[need.Physiological] < 0.4 && o.Inventory[entity.Food] < 1
}

// Give costs the giver and helps the receiver. Nothing in the need model
// rewards it much; the charitable do it because their conscience pays them.
var Give = mover("transfer/provision>needy")

// rewardValue converts a fee into need satisfaction. Money is worth what it
// can buy, so the same wage means a great deal to the hungry and little to
// the comfortable. That gradient is the labour market.
func rewardValue(a *entity.Agent, w *world.World, reward float64) need.Levels {
	price := math.Max(w.Market.Price[entity.Food], 0.05)
	units := math.Min(reward/price, 4)
	before := math.Min(a.Wealth/20, 1)
	after := math.Min((a.Wealth+reward)/20, 1)
	return need.Levels{
		need.Physiological: foodValue(a) * units * 0.5,
		need.Safety:        0.15 * (after - before),
	}
}

// CanFulfil reports whether the agent is in a position to take this job on.
// For deliveries that means having the goods; for service it means only
// believing you can do the work, which is not the same as being able to.
func CanFulfil(a *entity.Agent, w *world.World, r *entity.Request) bool {
	if !r.Wants(a.ID) || w.Find(r.Requester) == nil {
		return false
	}
	if r.Kind == entity.Deliver {
		return a.Inventory[r.Good] >= r.Amount
	}
	return true
}

// BestRequest is the job an agent would take if it took one: the most
// rewarding work it believes it can do, discounted by the walk. Selection uses
// believed competence, so agents cheerfully accept work they will do badly.
func BestRequest(a *entity.Agent, w *world.World) *entity.Request {
	var best *entity.Request
	bestScore := 0.0
	for _, r := range w.Requests {
		if !CanFulfil(a, w, r) {
			continue
		}
		confidence := 1.0
		if r.Kind == entity.Serve {
			confidence = 0.2 + a.Believes(r.Skill)
		}
		req := w.Find(r.Requester)
		score := r.Reward * confidence / float64(1+w.Grid.Dist(a.Pos, req.Pos))
		// People would rather work for someone they like and trust.
		score *= 1 + 0.3*a.Regard(r.Requester)
		if score > bestScore {
			best, bestScore = r, score
		}
	}
	return best
}

// Fulfil is one agent doing what another asked. It is the only action whose
// benefit lands on somebody else, which is what makes the settlement a
// settlement rather than a set of people standing near each other.
var Fulfil = &Def{
	Name: "fulfil request", Ticks: 3, Supply: supply([]*ontology.Class{ontology.Coin}, nil, false),
	Available: func(a *entity.Agent, w *world.World) bool { return BestRequest(a, w) != nil },
	Target: func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
		r := BestRequest(a, w)
		if r == nil {
			return entity.Pos{}, false
		}
		return w.Find(r.Requester).Pos, true
	},
	Expect: func(a *entity.Agent, w *world.World, _ entity.Pos) need.Levels {
		r := BestRequest(a, w)
		if r == nil {
			return need.Levels{}
		}
		gain := rewardValue(a, w, r.Reward)
		// Being asked, and being equal to it, is worth something in itself.
		if r.Kind == entity.Serve {
			gain[need.Esteem] += 0.12 * a.Believes(r.Skill)
		} else {
			gain[need.Esteem] += 0.04
		}
		if b := a.Look(r.Requester); b != nil {
			gain[need.Belonging] += 0.15 * b.Strength
		}
		if r.Directed == a.ID {
			gain[need.Esteem] += 0.08 // asked for by name
		}
		return gain
	},
	Apply: func(a *entity.Agent, w *world.World) {
		r := BestRequest(a, w)
		if r == nil {
			return
		}
		req := w.Find(r.Requester)
		if req == nil || w.Grid.Dist(a.Pos, req.Pos) > reachRadius {
			return
		}

		paid := math.Min(r.Reward, req.Wealth)
		req.Wealth -= paid
		a.Wealth += paid

		// What the work is actually worth depends on real skill, not on what
		// anyone believed. This is where confidence meets the world.
		var shown float64
		switch r.Kind {
		case entity.Deliver:
			a.Inventory[r.Good] -= r.Amount
			req.Inventory[r.Good] += r.Amount
			shown = a.Skills[entity.Farming]
		case entity.Serve:
			shown = a.Skills[r.Skill]
			serve(a, req, w, r.Skill)
		}

		// Both sides learn. The requester learns what this person can do; the
		// doer learns what they are worth.
		req.Rate(a.ID, skillOf(r), shown, w.Tick)
		req.Judge(a.ID, 0.25, w.Tick)
		req.AddBond(a.ID, 0.12)
		a.AddBond(req.ID, 0.08)
		a.Efficacy[skillOf(r)] = belief.Clamp(belief.Update(a.Efficacy[skillOf(r)], shown, belief.LearnRate))
		a.Needs.Add(need.Esteem, 0.1)

		witness(a, w, "fulfil request")
		w.Close(r.ID)
		w.Emit(event.Fulfilled, a.ID, req.ID, "%s did %s's work for %.1f", a.Name, req.Name, paid)
	},
}

func skillOf(r *entity.Request) entity.Skill {
	if r.Kind == entity.Deliver {
		return entity.Farming
	}
	return r.Skill
}

// serve applies skilled work done on someone else's behalf.
func serve(doer, client *entity.Agent, w *world.World, s entity.Skill) {
	level := doer.Skills[s]
	switch s {
	case entity.Building:
		client.Shelter = need.Clamp(client.Shelter + 0.35*w.Mods.BuildEfficiency*(0.3+level))
		doer.Learn(entity.Building, 0.02)
	case entity.Guarding:
		w.Safety = need.Clamp(w.Safety + 0.05*(0.3+level))
		client.Needs.Add(need.Safety, 0.15*(0.3+level))
		doer.Learn(entity.Guarding, 0.02)
	case entity.Scholarship:
		// Tutoring is teaching that was asked for and paid for, and it
		// carries exactly as far as teaching that was not: to a gap under
		// the tutor, and never to mastery. It used to scale by 0.3+level,
		// which meant a tutor who knew nothing still taught something; now
		// a tutor who knows nothing teaches nothing, and only the doing of
		// it - the line under this one - takes anybody to the top.
		client.Taught(entity.Scholarship, 0.05, level)
		w.Knowledge += (0.1 + level) * w.Mods.StudyRate
		doer.Learn(entity.Scholarship, 0.015)
	case entity.Crafting:
		client.Inventory[entity.Tools] += (0.3 + level) * w.Mods.CraftQuality
		doer.Learn(entity.Crafting, 0.015)
	case entity.Farming:
		client.Inventory[entity.Food] += (0.8 + 2*level) * w.Mods.FarmYield
		doer.Learn(entity.Farming, 0.01)
	case entity.Fishing:
		client.Inventory[entity.Food] += (0.6 + 1.5*level) * w.Mods.FishYield
		doer.Learn(entity.Fishing, 0.015)
	}
	client.Needs.Add(need.Belonging, 0.05)
}

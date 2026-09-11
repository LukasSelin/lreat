package system

import (
	"math"

	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/world"
)

const (
	// SelfDoubt is the believed skill below which an agent stops thinking of
	// a job as something it could do itself.
	SelfDoubt = 0.3
	// GreedWealth is the fortune above which an agent starts preferring to
	// pay rather than spend its own time, even when perfectly capable.
	GreedWealth = 12
	// RequestLife is how long a request stays on the board before the
	// requester gives up on it: the better part of a season.
	RequestLife = 80 * clock.Day
	// AskChance throttles how often an eligible agent actually asks.
	AskChance = 0.08
)

// task is what an agent wants done, derived from whichever need is loudest.
type task struct {
	kind   entity.RequestKind
	good   entity.Good
	amount float64
	skill  entity.Skill
	gain   need.Levels // what having it done would be worth
}

// wanted maps the agent's most pressing need onto a job somebody could do for
// it. Not every need can be bought; belonging in particular cannot, which is
// why the lonely rich are still lonely.
func wanted(a *entity.Agent, w *world.World) (task, bool) {
	urgency := need.Urgencies(a.Needs)
	top := need.Tier(0)
	for t := need.Tier(1); t < need.Count; t++ {
		if urgency[t] > urgency[top] {
			top = t
		}
	}

	switch top {
	case need.Physiological:
		if a.Inventory[entity.Food] < 1.5 {
			return task{
				kind: entity.Deliver, good: entity.Food, amount: 2, skill: entity.Farming,
				gain: need.Levels{need.Physiological: 0.5},
			}, true
		}
	case need.Safety:
		if a.Shelter < 0.5 {
			return task{
				kind: entity.Serve, skill: entity.Building,
				gain: need.Levels{need.Safety: 0.35},
			}, true
		}
		if w.Safety < 0.5 {
			return task{
				kind: entity.Serve, skill: entity.Guarding,
				gain: need.Levels{need.Safety: 0.2},
			}, true
		}
	case need.Esteem:
		if a.Inventory[entity.Tools] < 1 {
			return task{
				kind: entity.Serve, skill: entity.Crafting,
				gain: need.Levels{need.Esteem: 0.2},
			}, true
		}
	case need.Actualization:
		if a.Skills[entity.Scholarship] < 0.5 {
			return task{
				kind: entity.Serve, skill: entity.Scholarship,
				gain: need.Levels{need.Actualization: 0.3, need.Esteem: 0.05},
			}, true
		}
	}
	return task{}, false
}

// bestKnown returns the agent this one believes is most capable at a skill,
// among the people it actually knows and does not despise. Belief, not truth:
// a well-regarded incompetent gets hired over a capable stranger.
func bestKnown(a *entity.Agent, w *world.World, s entity.Skill) entity.ID {
	var pick entity.ID
	best := 0.2 // must beat a plain stranger to be worth naming
	for i := range a.Bonds {
		b := &a.Bonds[i]
		if b.Regard < -0.2 || w.Find(b.To) == nil {
			continue
		}
		score := b.Competence[s] * (1 + 0.3*b.Regard)
		if score > best {
			pick, best = b.To, score
		}
	}
	return pick
}

// Requests posts new work and clears out what nobody did.
func Requests(w *world.World) {
	expire(w)
	for _, a := range w.Agents {
		if !a.Species().Settles {
			continue
		}
		if w.HasRequestFrom(a.ID) || w.RNG.Float64() >= AskChance {
			continue
		}
		if r, ok := compose(a, w); ok {
			id := w.Post(r)
			if id == 0 {
				continue
			}
			motive := "needing it done"
			if r.Greed {
				motive = "having better things to do"
			}
			w.Emit(event.Requested, a.ID, r.Directed, "%s asked for %s, %s", a.Name, describe(r), motive)
		}
	}
}

// compose decides whether an agent asks for help, and on what terms.
//
// Three things must line up, and they are exactly the three the design calls
// for: the agent wants something, it believes it will not get it by itself,
// and it believes the fee is worth less to it than the result.
func compose(a *entity.Agent, w *world.World) (entity.Request, bool) {
	t, ok := wanted(a, w)
	if !ok {
		return entity.Request{}, false
	}

	// Would I do this myself? Either I doubt I can, or I am comfortable
	// enough that my time is worth more to me than the fee.
	doubts := a.Believes(t.skill) < SelfDoubt
	greedy := a.Wealth > GreedWealth && a.Needs[need.Physiological] > 0.6
	if !doubts && !greedy {
		return entity.Request{}, false
	}

	urgency := need.Urgencies(a.Needs)
	reward := math.Min(a.Wealth*0.4, 1.5+5*Score(a, t.gain, urgency))
	if reward < 0.5 || reward > a.Wealth {
		return entity.Request{}, false
	}

	// The belief of benefit: what I get must be worth more to me than what
	// paying for it costs me. Both sides are valued through my own needs.
	benefit := Score(a, t.gain, urgency)
	cost := Score(a, rewardCost(a, w, reward), urgency)
	if benefit <= cost {
		return entity.Request{}, false
	}

	return entity.Request{
		Requester: a.ID,
		Directed:  bestKnown(a, w, t.skill),
		Kind:      t.kind,
		Good:      t.good,
		Amount:    t.amount,
		Skill:     t.skill,
		Reward:    reward,
		Deadline:  w.Tick + RequestLife,
		Greed:     greedy && !doubts,
	}, true
}

// rewardCost is what parting with a fee costs the payer, in need terms.
func rewardCost(a *entity.Agent, w *world.World, reward float64) need.Levels {
	before := math.Min(a.Wealth/20, 1)
	after := math.Min(math.Max(a.Wealth-reward, 0)/20, 1)
	price := math.Max(w.Market.Price[entity.Food], 0.05)
	return need.Levels{
		need.Safety:        0.15 * (before - after),
		need.Physiological: 0.1 * math.Min(reward/price, 3) * (1 - a.Needs[need.Physiological]),
	}
}

// expire drops requests nobody took. The requester learns something from it:
// the person they named goes down in their estimation, which is how a
// reputation for unreliability forms.
func expire(w *world.World) {
	kept := w.Requests[:0]
	for _, r := range w.Requests {
		if w.Tick < r.Deadline && w.Find(r.Requester) != nil {
			kept = append(kept, r)
			continue
		}
		if req := w.Find(r.Requester); req != nil {
			if r.Directed != 0 {
				req.Judge(r.Directed, -0.15, w.Tick)
				b := req.Look(r.Directed)
				if b != nil {
					b.Competence[r.Skill] *= 0.7
				}
			}
			w.Emit(event.Unmet, req.ID, r.Directed, "nobody answered %s's request for %s", req.Name, describe(*r))
		}
	}
	w.Requests = kept
}

func describe(r entity.Request) string {
	if r.Kind == entity.Deliver {
		return r.Good.String()
	}
	return r.Skill.String()
}

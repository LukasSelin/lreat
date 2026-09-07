package action

import (
	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// Transferring is one verb. Giving and stealing are the same act with the
// direction reversed: find a person who stands in some relation to you,
// go to them, and move something between the two packs. The ontology
// says what moves, which way, and to whom; a hand says how much, when it
// is worth doing, and what passes between the two people besides the
// thing. Fulfilling a request is bound by hand still: it is as often
// skilled work as a handing over, and what it hands over is whatever was
// asked.

// whom is who stands in a role to the actor, given the good the act is
// about: a holder has some to spare, the needy have none and are hungry.
var whom = map[*ontology.Role]func(good entity.Good) func(o *entity.Agent) bool{
	&ontology.Holder: func(good entity.Good) func(o *entity.Agent) bool {
		return func(o *entity.Agent) bool { return o.Inventory[good] >= 1 }
	},
	&ontology.Needy: func(good entity.Good) func(o *entity.Agent) bool {
		return func(o *entity.Agent) bool { return o.Needs[need.Physiological] < 0.4 && o.Inventory[good] < 1 }
	},
}

// hand is what a transfer moves and what moves with it.
type hand struct {
	Name string
	// Amount is how much of the good changes hands.
	Amount float64
	// Want is what the actor must be short of to take; Spare what they
	// must have to give.
	Want, Spare float64
	// Regard is how the other comes to think of the actor for it.
	Regard float64
	// Bonds is what the act adds to the tie each way, and Rift what it
	// takes from the other's tie to the actor.
	Bonds struct{ Theirs, Mine float64 }
	Rift  float64
	// Worth is what the transfer is worth to the actor, by need tier. What
	// it promises of company and standing it gives; what it promises the
	// body comes with what was got.
	Worth func(a *entity.Agent) need.Levels
	// Event is what the act is called in the record, and Text how it reads.
	Event event.Kind
	Text  string
}

var hands = map[string]hand{
	// Stealing is the option that makes conscience mean something. It is
	// fast, it works, and the only thing standing against it is what the
	// agent believes about itself and what it thinks the neighbours will
	// make of it. The victim always knows; bystanders may or may not care.
	"transfer/provision<holder": {
		Name: "steal", Amount: 1, Want: 1, Regard: -0.6, Rift: 0.3,
		Worth: func(a *entity.Agent) need.Levels { return need.Levels{need.Physiological: foodValue(a) * 1.5} },
		Event: event.Stolen, Text: "stole food from",
	},
	// Giving costs the giver and helps the receiver. Nothing in the need
	// model rewards it much; the charitable do it because their conscience
	// pays them.
	"transfer/provision>needy": {
		Name: "give", Amount: 1, Spare: 2, Regard: 0.4,
		Bonds: struct{ Theirs, Mine float64 }{0.15, 0.1},
		Worth: func(*entity.Agent) need.Levels { return need.Levels{need.Belonging: 0.08, need.Esteem: 0.05} },
		Event: event.Given, Text: "gave food to",
	},
}

// transferring is the act the ontology entails for a transfer with a hand,
// carried out from it, or nil where there is none.
func transferring(in ontology.Instance) *Def {
	h, ok := hands[in.Key]
	good, ok2 := goods[in.Schema.Object]
	role, ok3 := whom[in.Schema.Role]
	if !ok || !ok2 || !ok3 {
		return nil
	}
	other := role(good)
	seize := in.Schema.Dir == ontology.Seize
	d := &Def{Name: h.Name, Ticks: in.Ticks}
	d.Available = func(a *entity.Agent, w *world.World) bool {
		if seize && a.Inventory[good] >= h.Want || !seize && a.Inventory[good] < h.Spare {
			return false
		}
		return nearestWith(a, w, reachRadius, other) != nil
	}
	d.Target = func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
		if o := nearestWith(a, w, reachRadius, other); o != nil {
			return o.Pos, true
		}
		return entity.Pos{}, false
	}
	d.With = func(a *entity.Agent, w *world.World, _ entity.Pos) *entity.Agent {
		return nearestWith(a, w, reachRadius, other)
	}
	d.Expect = func(a *entity.Agent, _ *world.World, _ entity.Pos) need.Levels { return h.Worth(a) }
	d.Apply = func(a *entity.Agent, w *world.World) {
		o := nearestWith(a, w, 2, other)
		if o == nil {
			return // moved on, or fed, while we walked
		}
		from, to := a, o
		if seize {
			from, to = o, a
		}
		from.Inventory[good] -= h.Amount
		to.Inventory[good] += h.Amount
		o.Judge(a.ID, h.Regard, w.Tick)
		if h.Bonds.Theirs > 0 {
			o.AddBond(a.ID, h.Bonds.Theirs)
		}
		if h.Bonds.Mine > 0 {
			a.AddBond(o.ID, h.Bonds.Mine)
		}
		if h.Rift > 0 {
			if b := o.Look(a.ID); b != nil {
				b.Strength = belief.Clamp(b.Strength - h.Rift)
			}
		}
		for tier, gain := range h.Worth(a) {
			if need.Tier(tier) != need.Physiological && gain > 0 {
				a.Needs.Add(need.Tier(tier), gain)
			}
		}
		witness(a, w, h.Name)
		remorse(a, h.Name)
		w.Emit(h.Event, a.ID, o.ID, "%s %s %s", a.Name, h.Text, o.Name)
	}
	return d
}

// handing is the transfer the ontology entails under key, for the acts the
// rest of the package refers to by name.
func handing(key string) *Def {
	d := transferring(instances[key])
	if d == nil {
		panic("action: no hand for " + key)
	}
	return d
}

package ontology

import "lreat/core/habit"

// The prior of an instantiated act is composed rather than written: the
// verb says what kind of moment the act is, the classes say what it is
// about, the site where and the role with whom. What a class adds depends
// on how the act touches it. Something the act wants (what Take takes,
// what Make makes) contributes the class's Wanting; something the act must
// already have (an input) contributes its Having, which is the stock
// coordinate turned round. A schema's own Prior is the residue that none
// of those explain.

// VerbPrior is the moment each kind of act belongs to before anything is
// said about what it is done with.
var VerbPrior = [VerbCount]habit.Signature{
	Take:     {habit.Near: 0.1},
	Make:     {habit.Industry: 0.4},
	Raise:    {habit.Industry: 0.3},
	Tend:     {habit.Near: 0.4},
	Consume:  {habit.Hunger: 1, habit.Near: 0.5},
	Dwell:    {},
	Exchange: {habit.Near: 0.4},
	Transfer: {habit.Charity: 0.6, habit.Near: 0.5},
	Pass:     {},
	Strike:   {habit.Honesty: 0.2, habit.Charity: -0.5, habit.Near: 0.5},
	Move:     {},
}

// TraitPrior is what wanting something with a trait adds. Edible things are
// wanted by the hungry; a workplace is wanted by nobody in particular.
var TraitPrior = map[Trait]habit.Signature{
	Edible: {habit.Hunger: 0.8},
}

// SeizePrior is what taking rather than giving adds: the moral coordinates
// that make an honest agent fail to recognise the moment.
var SeizePrior = habit.Signature{habit.Honesty: -1, habit.Charity: -0.6}

// ForShare is how much of the wanting of what a material makes reaches the
// wanting of the material: timber is wanted for the roof it will be. Only
// one step is taken, and outputs weigh by how far into reach making them
// starts, so a newborn wants wood for a house and not for a forge.
const ForShare = 0.6

// Having is the moment of already holding c: its stock coordinates turned
// round, and nothing of what it is for.
func Having(c *Class) habit.Signature {
	s := c.DerivedPrior()
	for i := range s {
		s[i] = -s[i]
	}
	return s
}

// Wanting is the moment of lacking c: its own coordinates, what its traits
// draw when it is to be got rather than made, and a share of the wanting
// of whatever it goes into. Making something edible is not a hungry act;
// it is a keeping one, so the trait is left out of a made thing's want.
func Wanting(c *Class, getting bool) habit.Signature {
	s := c.DerivedPrior()
	if getting {
		for t, p := range TraitPrior {
			if c.Has(t) {
				add(&s, p, 1)
			}
		}
	}
	var down habit.Signature
	var weight float64
	for i := range Schemas {
		sc := &Schemas[i]
		if sc.Output == nil || (sc.Verb != Make && sc.Verb != Raise) {
			continue
		}
		for _, in := range sc.Inputs {
			if c.IsA(in) {
				add(&down, sc.Output.DerivedPrior(), sc.Reach0)
				weight += sc.Reach0
			}
		}
	}
	if weight > 0 {
		add(&s, down, ForShare/weight)
	}
	return s
}

// Compose derives the prior of one instance.
func Compose(sc *Schema, object, site *Class) habit.Signature {
	s := VerbPrior[sc.Verb]
	add(&s, sc.Prior, 1)
	for _, in := range sc.Inputs {
		add(&s, Having(in), 1)
	}
	if sc.Output != nil {
		// Refining, a meal out of provisions, does not want the class it
		// already holds; only what the finer thing adds on top of it.
		refined := false
		for _, in := range sc.Inputs {
			if sc.Output.IsA(in) {
				add(&s, sc.Output.DerivedPrior(), 1)
				add(&s, in.DerivedPrior(), -1)
				refined = true
				break
			}
		}
		if !refined {
			add(&s, Wanting(sc.Output, sc.Verb == Exchange), 1)
		}
	}
	if object != nil {
		switch {
		case sc.Verb == Consume, sc.Verb == Transfer && sc.Dir == Give:
			add(&s, Having(object), 1)
		default:
			add(&s, Wanting(object, true), 1)
		}
	}
	if site != nil {
		add(&s, site.At, 1)
	}
	if sc.Role != nil {
		add(&s, sc.Role.Prior, 1)
	}
	if sc.Dir == Seize {
		add(&s, SeizePrior, 1)
	}
	for i := range s {
		s[i] = max(-1, min(1, s[i]))
	}
	return s
}

func add(dst *habit.Signature, src habit.Signature, k float64) {
	for i := range dst {
		dst[i] += k * src[i]
	}
}

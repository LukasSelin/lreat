package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gowl/owl"

	"lreat/core/belief"
	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/ontology"
)

// This is core/ontology said again in OWL 2, and nothing here decides
// anything: every axiom is read off the Go declarations at run time, so the
// document goes stale only if this file is not run again.
//
// What the two say differently is worth stating, because it is most of the
// mapping. lreat's trees are single-inheritance and a trait is a bitfield;
// OWL has no bitfields, so a trait becomes the class of the things that have
// it and "Provision SubClassOf Edible" is how the bit is written down. That
// turns Has(Edible) from a mask test into a subsumption, which is what a
// reasoner is for. Roles stay out of the Person tree here as they do there -
// nobody is a pupil the way an oak is timber - but they are classes rather
// than values, because an act names one and what an act names has to be
// something a class expression can hold.
//
// Acts are the instantiated catalog rather than the schemas. A schema is a
// statement over classes and the trees are what make it concrete, so the
// thing worth writing down is what the trees currently entail: add a class
// and the document gains every act the trees entail for it, which is the
// same property Instantiate has.
//
// Signatures are the one place where a number rather than a shape is
// carried. Each habit coordinate is a data property, so the moment an act
// belongs to is queryable rather than opaque, and zero coordinates are left
// out: a signature is sparse and saying so is the point.

const (
	base = "https://github.com/LukasSelin/lreat/ontology"
	ns   = base + "#"
)

// Build renders the whole of core/ontology as an OWL 2 ontology.
func Build() (*owl.Ontology, error) {
	b := &builder{
		o:        owl.New(base),
		declared: map[string]bool{},
		names:    map[string]string{},
	}
	b.o.Prefix("", ns)
	b.o.Annotate(owl.RDFSComment, owl.Str(preamble()))

	b.upper()
	b.traits()
	b.trees()
	b.affords()
	b.roles()
	b.verbs()
	b.skills()
	b.acts()
	b.processes()
	b.transforms()

	if err := b.o.Err(); err != nil {
		return nil, err
	}
	return b.o, nil
}

func preamble() string {
	return "lreat's ontology, rendered from core/ontology by ./owl. Classes are what the " +
		"world is made of; acts are the catalog the trees entail, one class per " +
		"instantiated act, labelled with the key its habit slot is kept under. The " +
		"numbers are the simulation's own: a tick is a day and a year is " +
		strconv.Itoa(clock.Year) + " of them, reach0 is how far into reach an act starts for a " +
		"newborn, and prior-* is the moment an act or a class belongs to, one data " +
		"property per habit coordinate. Nothing here is written by hand. Regenerate " +
		"rather than edit."
}

type builder struct {
	o *owl.Ontology
	// declared guards against declaring an entity twice, which would show up
	// as a duplicate axiom rather than as an error.
	declared map[string]bool
	// names guards against two different things wanting one local name,
	// which sanitising keys into IRIs makes possible in principle.
	names map[string]string
}

// upper is the handful of classes and properties that hold the rest
// together. They have no counterpart in core/ontology: they are what OWL
// needs in order to say what Go says with a struct field.
func (b *builder) upper() {
	b.class(":Act", "act", "One act of the instantiated catalog: a schema bound to the leaf classes it operates on. The label is the key its habit slot is kept under.")
	b.class(":Verb", "verb", "A kind of act - what changes when it is done.")
	b.class(":Role", "role", "How a person stands to the actor, decided at the moment of acting rather than fixed in a tree.")
	b.class(":Skill", "skill", "A competence an act draws on and improves.")
	b.class(":Technology", "technology", "A discovery an act is gated on, the settlement's or the actor's.")
	b.class(":Direction", "direction", "Which way a transfer runs.")
	b.class(":Process", "process", "A maturing a thing goes through on its own: monotone, on a clock, and readable in advance, which is what lets an act ask for it.")
	b.class(":Stage", "stage", "One named span of a process, measured in growing weather rather than in ticks passed: a wood raised in the autumn stands still until the thaw.")
	b.class(":Transform", "transform", "A change that happens to a thing rather than one somebody does to it. Unlike a process it cannot be read in advance.")
	b.class(":Workplace", "workplace", "A site trait that lends an act the tools it needs. A dwelling has all of them; the market lends a bench and a desk, the tavern a hearth.")

	b.op(":verb", "verb", ":Act", ":Verb", "What kind of act this is.")
	b.op(":actor", "actor", ":Act", ":Thing", "Who does the act: a person, unless the schema names another creature.")
	b.op(":handles", "handles", ":Act", "", "What the act operates on: the thing it is done to or with.")
	b.op(":at", "at", ":Act", ":Site", "Where the act happens.")
	b.op(":atWorkplace", "at workplace", ":Act", ":Workplace", "The act is done wherever this trait holds, rather than at a named site.")
	b.op(":consumes", "consumes", ":Act", ":Material", "What the act spends. Without it the act cannot start.")
	b.op(":mayUse", "may use", ":Act", ":Material", "What improves the act without being needed for it: a stone in the walls of a house.")
	b.op(":yields", "yields", "", "", "What the act or the process brings on.")
	b.op(":toward", "toward", ":Act", ":Role", "Whom the act is done to or with.")
	b.op(":direction", "direction", ":Act", ":Direction", "Which way the transfer runs.")
	b.op(":requiresSkill", "requires skill", ":Act", ":Skill", "The competence the act draws on.")
	b.op(":requiresTech", "requires technology", ":Act", ":Technology", "The discovery the act is gated on.")
	b.op(":season", "season", ":Act", "", "The thing whose half of the year this act keeps, where the act itself handles none of it: clearing is warm-half work because a field is for grain.")
	b.op(":affords", "affords", "", ":Thing", "What a site or a holder offers: the ground what lies in it and the brush a deer lives on, a person what is in their pack and their purse, the market what is on its shelves.")
	b.op(":processOf", "process of", ":Process", "", "What the process happens to.")
	b.op(":stageOf", "stage of", ":Stage", ":Process", "The process this stage belongs to.")
	b.op(":changes", "changes", ":Transform", "", "What the transform acts on.")
	b.op(":becomes", "becomes", ":Transform", "", "What it turns into. A transform with none leaves nothing behind.")
	b.op(":heldIn", "held in", ":Transform", "", "Whose keeping the thing is in while the transform runs. A transform with none is about a site, which is in nobody's.")
	b.op(":arrestedBy", "arrested by", ":Transform", ":Site", "A site that stops the change.")

	b.dp(":ticks", "ticks", ":Act", owl.XSDInteger, "How long the doing takes, in days.")
	b.dp(":reach0", "reach at birth", ":Act", owl.XSDDecimal, "How far into reach the act starts for a newborn.")
	b.dp(":skilled", "skilled", ":Act", owl.XSDBoolean, "Whether how the act goes turns on how good the actor is at it.")
	b.dp(":lack", "lack", ":Material", owl.XSDDecimal, "How strongly being short of this registers. Inherited down the tree.")
	b.dp(":warmth", "warmth", "", owl.XSDDecimal, "Which half of the year getting this belongs to: positive for the green half, negative for the cold one, zero for a thing had in any weather.")
	b.dp(":rate", "rate", "", owl.XSDDecimal, "For a transform, the share lost or the chance taken per tick; for a process, how much of a full stock the stand puts on per growing tick.")
	b.dp(":kept", "kept", ":Transform", owl.XSDBoolean, "Whether somebody living to keep the thing arrests the change, which is what makes ruin the fate of what is left behind rather than of everything.")
	b.dp(":stageTicks", "stage ticks", ":Stage", owl.XSDDecimal, "How much growing weather the stage takes, in days.")
	b.dp(":stageIndex", "stage index", ":Stage", owl.XSDInteger, "Where the stage falls in its process, from zero.")

	b.dp(":prior", "prior", "", owl.XSDDecimal, "The moment a thing belongs to, one sub-property per habit coordinate. On an act it is the composed prior; on a verb, a role or a trait it is what that contributes to one; on a class of thing it is the delta the class adds to the moment of lacking it.")
	b.dp(":at-moment", "at", ":Site", owl.XSDDecimal, "What being at a site is like, one sub-property per habit coordinate.")
	b.dp(":valence", "valence", ":Act", owl.XSDDecimal, "How much an act upholds or violates each norm, one sub-property per norm. It is a property of the act, shared by everyone; what differs between agents is how much they care.")
	for _, n := range habit.Names {
		b.subDP(":prior-"+n, "prior "+n, ":prior")
		b.subDP(":at-"+n, "at "+n, ":at-moment")
	}
	for n := belief.Norm(0); n < belief.NormCount; n++ {
		b.subDP(":valence-"+n.String(), "valence "+n.String(), ":valence")
	}

	b.class(":Give", "give", "Handing over.")
	b.class(":Seize", "seize", "A gift the other way round. Everything that makes theft theft is in the traits and the valence, not in a separate verb.")
	b.sub(":Give", ":Direction")
	b.sub(":Seize", ":Direction")
	b.o.Add(owl.DisjointClasses{b.o.Class(":Give"), b.o.Class(":Seize")})
	b.signature(":Seize", ":prior-", ontology.SeizePrior)
}

// traits are qualities, and a quality is the class of the things that have
// it: Edible is everything edible, and Provision is under it. Every trait
// belongs wholly to one side of the world - nothing edible is a site and
// nothing roofed is a material - so each is anchored under the root it
// applies to rather than left free of the trees altogether. Edible is
// anchored one step up from the rest of a material's traits, because the
// brush a deer lives on is edible and is a thing rather than a material:
// a material is what a person carries and trades, and what a deer eats is
// neither.
func (b *builder) traits() {
	b.sub(":Workplace", ":Site")
	groups := []struct {
		under  string
		traits []ontology.Trait
	}{
		{":Thing", []ontology.Trait{ontology.Edible}},
		{":Material", []ontology.Trait{ontology.Perishable, ontology.Burnable, ontology.Buildable, ontology.Heavy, ontology.Wears}},
		{":Site", []ontology.Trait{ontology.Living, ontology.Roofed, ontology.Owned, ontology.Public, ontology.Passable}},
		{":Workplace", []ontology.Trait{ontology.Bench, ontology.Hearth, ontology.Forge, ontology.Desk, ontology.Company, ontology.Trade, ontology.Store}},
	}
	for _, g := range groups {
		for _, t := range g.traits {
			b.claim(traitIRI(t), "trait "+t.String())
			b.class(traitIRI(t), t.String(), "Everything with the "+t.String()+" trait.")
			b.sub(traitIRI(t), g.under)
		}
	}

	// A trait may say what wanting something with it is like. One does.
	for t, s := range ontology.TraitPrior {
		b.signature(traitIRI(t), ":prior-", s)
	}

	// Every named bit has to land in one of the groups above, or the trait
	// was added to core/ontology and not to this file.
	for bit := ontology.Trait(1); bit != 0; bit <<= 1 {
		if bit.String() != "" && !b.declared[traitIRI(bit)] {
			panic("owl: trait " + bit.String() + " is in core/ontology but in no group here")
		}
	}
}

// trees walks the two trees and writes each class down with its parent, its
// own traits, and whatever it says about the moment of lacking it.
func (b *builder) trees() {
	for _, root := range []*ontology.Class{ontology.Thing, ontology.Site} {
		for _, c := range root.Family() {
			name := b.claim(classIRI(c), "class "+c.Path())
			b.class(name, c.Name, "")
			if c.Parent != nil {
				b.sub(name, classIRI(c.Parent))
			}
			for bit := ontology.Trait(1); bit != 0; bit <<= 1 {
				if c.Traits&bit != 0 {
					b.sub(name, traitIRI(bit))
				}
			}
			b.signature(name, ":prior-", c.Prior)
			b.signature(name, ":at-", c.At)
			if c.Lack != 0 {
				b.value(name, ":lack", dec(c.Lack))
			}
			if c.Warmth != 0 {
				b.value(name, ":warmth", dec(c.Warmth))
			}
			// Siblings are alternatives: a thing is one of them, not two.
			if kids := c.Children(); len(kids) > 1 {
				ces := make([]owl.ClassExpression, len(kids))
				for i, k := range kids {
					ces[i] = b.o.Class(classIRI(k))
				}
				b.o.Add(owl.DisjointClasses(ces))
			}
		}
	}
}

func (b *builder) affords() {
	holders := make([]*ontology.Class, 0, len(ontology.Affords))
	for c := range ontology.Affords {
		holders = append(holders, c)
	}
	sort.Slice(holders, func(i, j int) bool { return holders[i].Name < holders[j].Name })
	for _, h := range holders {
		for _, m := range ontology.Affords[h] {
			b.some(classIRI(h), ":affords", classIRI(m))
		}
	}
}

// roles are the eight core/ontology names, in the order it declares them.
// There is no exported list of them there, so the list is here.
var roles = []*ontology.Role{
	&ontology.Self, &ontology.Neighbour, &ontology.Requester,
	&ontology.Needy, &ontology.Holder, &ontology.Pupil, &ontology.Wrongdoer,
	&ontology.Fellow,
}

func (b *builder) roles() {
	for _, r := range roles {
		b.claim(roleIRI(r), "role "+r.Name)
		b.class(roleIRI(r), r.Name, "")
		b.sub(roleIRI(r), ":Role")
		b.signature(roleIRI(r), ":prior-", r.Prior)
	}
}

func (b *builder) verbs() {
	for v := ontology.Verb(0); v < ontology.VerbCount; v++ {
		b.claim(verbIRI(v), "verb "+v.String())
		b.class(verbIRI(v), v.String(), "")
		b.sub(verbIRI(v), ":Verb")
		b.signature(verbIRI(v), ":prior-", ontology.VerbPrior[v])
	}
}

func (b *builder) skills() {
	for s := entity.Skill(0); s < entity.SkillCount; s++ {
		b.claim(skillIRI(s), "skill "+s.String())
		b.class(skillIRI(s), s.String(), "")
		b.sub(skillIRI(s), ":Skill")
	}
}

// acts writes the catalog the trees currently entail.
func (b *builder) acts() {
	for _, in := range ontology.Instantiate() {
		name := b.claim(actIRI(in.Key), "act "+in.Key)
		b.class(name, in.Key, "")
		b.sub(name, ":Act")
		sc := in.Schema

		b.some(name, ":verb", verbIRI(sc.Verb))
		actor := in.Actor
		if actor == nil {
			actor = ontology.Person
		}
		b.some(name, ":actor", classIRI(actor))
		if in.Object != nil {
			b.some(name, ":handles", classIRI(in.Object))
		}
		if in.Site != nil {
			b.some(name, ":at", classIRI(in.Site))
		}
		if sc.SiteTrait != 0 {
			b.some(name, ":atWorkplace", traitIRI(sc.SiteTrait))
		}
		for _, c := range sc.Inputs {
			b.some(name, ":consumes", classIRI(c))
		}
		for _, c := range sc.Optional {
			b.some(name, ":mayUse", classIRI(c))
		}
		if sc.Output != nil {
			b.some(name, ":yields", classIRI(sc.Output))
		}
		if sc.Role != nil {
			b.some(name, ":toward", roleIRI(sc.Role))
			if sc.Verb == ontology.Transfer {
				if sc.Dir == ontology.Seize {
					b.some(name, ":direction", ":Seize")
				} else {
					b.some(name, ":direction", ":Give")
				}
			}
		}
		if sc.Season != nil {
			b.some(name, ":season", classIRI(sc.Season))
		}
		if in.Skilled {
			b.value(name, ":skilled", owl.Bool(true))
			if namesItsSkill(sc) {
				b.some(name, ":requiresSkill", skillIRI(in.Skill))
			} else {
				b.comment(name, "Skilled, but in nothing this document names: what it draws on is the practice being handed over, and that is not known until the act is done.")
			}
		}
		if in.Tech != "" {
			b.claim(techIRI(in.Tech), "technology "+in.Tech)
			b.class(techIRI(in.Tech), in.Tech, "")
			b.sub(techIRI(in.Tech), ":Technology")
			b.some(name, ":requiresTech", techIRI(in.Tech))
		}
		b.value(name, ":ticks", owl.Int(in.Ticks))
		if in.Reach0 != 0 {
			b.value(name, ":reach0", dec(in.Reach0))
		}
		b.signature(name, ":prior-", in.Prior)
		for n := belief.Norm(0); n < belief.NormCount; n++ {
			if v := in.Valence[n]; v != 0 {
				b.value(name, ":valence-"+n.String(), dec(v))
			}
		}
	}
}

func (b *builder) processes() {
	for _, p := range ontology.Processes {
		name := b.claim(":"+slug(p.Name)+"-process", "process "+p.Name)
		b.class(name, p.Name, "")
		b.sub(name, ":Process")
		b.some(name, ":processOf", classIRI(p.Of))
		b.some(name, ":yields", classIRI(p.Yields))
		if p.Rate != 0 {
			b.value(name, ":rate", dec(p.Rate))
		}
		for i, s := range p.Stages {
			stage := b.claim(":"+slug(p.Name)+"-"+slug(s.Name), "stage "+p.Name+"/"+s.Name)
			b.class(stage, p.Name+"/"+s.Name, "")
			b.sub(stage, ":Stage")
			b.some(stage, ":stageOf", name)
			b.value(stage, ":stageTicks", dec(s.Ticks))
			b.value(stage, ":stageIndex", owl.Int(i))
		}
	}
	inEar := ontology.InEar
	b.comment(":"+slug(inEar.Process.Name)+"-"+slug(inEar.Process.Stages[inEar.Stage].Name),
		"The phase a harvest asks for: a strip is not worth cutting until it is in ear, and a strip cut is bare ground again.")
}

func (b *builder) transforms() {
	for i := range ontology.Transforms {
		t := &ontology.Transforms[i]
		name := b.claim(transformIRI(t), "transform "+transformLabel(t))
		b.class(name, transformLabel(t), t.Says)
		b.sub(name, ":Transform")
		b.some(name, ":changes", classIRI(t.From))
		if t.To != nil {
			b.some(name, ":becomes", classIRI(t.To))
		}
		if t.In != nil {
			b.some(name, ":heldIn", classIRI(t.In))
		}
		if t.Unless != nil {
			b.some(name, ":arrestedBy", classIRI(t.Unless))
		}
		b.value(name, ":rate", dec(t.Rate))
		if t.Kept {
			b.value(name, ":kept", owl.Bool(true))
		}
	}
}

// namesItsSkill reports whether a schema says which competence it draws on.
// Two do not: handing a requester what they asked for, and teaching. Both set
// Skilled and leave Skill alone, so the field reads back as its zero value,
// farming, rather than as a fact - and what they actually draw on is the
// practice being handed over, which nothing knows until the act is done. No
// Take or Tend schema is in that position, which is what the verbs rule out.
func namesItsSkill(sc *ontology.Schema) bool {
	if !sc.Skilled {
		return false
	}
	if sc.Skill != entity.Farming {
		return true
	}
	return sc.Verb == ontology.Take || sc.Verb == ontology.Tend
}

// --- writing ---------------------------------------------------------------

func (b *builder) class(name, label, comment string) {
	if b.declared[name] {
		return
	}
	b.declared[name] = true
	b.o.Declare(b.o.Class(name))
	if label != "" {
		b.label(name, label)
	}
	if comment != "" {
		b.comment(name, comment)
	}
}

func (b *builder) sub(name, super string) {
	b.o.Add(owl.SubClassOf{Sub: b.o.Class(name), Super: b.o.Class(super)})
}

func (b *builder) some(name, prop, filler string) {
	b.o.Add(owl.SubClassOf{
		Sub:   b.o.Class(name),
		Super: owl.Some(b.o.ObjectProperty(prop), b.o.Class(filler)),
	})
}

func (b *builder) value(name, prop string, lit owl.Literal) {
	b.o.Add(owl.SubClassOf{
		Sub:   b.o.Class(name),
		Super: owl.DataValue(b.o.DataProperty(prop), lit),
	})
}

// signature writes the nonzero coordinates of s under a family of data
// properties. A coordinate left out is one the moment says nothing about.
func (b *builder) signature(name, family string, s habit.Signature) {
	for i, v := range s {
		if v != 0 {
			b.value(name, family+habit.Names[i], dec(v))
		}
	}
}

func (b *builder) label(name, text string) {
	b.o.Add(owl.AnnotationAssertion{Property: owl.RDFSLabel, Subject: b.iri(name), Value: owl.Str(text)})
}

func (b *builder) comment(name, text string) {
	b.o.Add(owl.AnnotationAssertion{Property: owl.RDFSComment, Subject: b.iri(name), Value: owl.Str(text)})
}

func (b *builder) iri(name string) owl.IRI { return owl.IRI(ns + strings.TrimPrefix(name, ":")) }

func (b *builder) op(name, label, domain, rng, comment string) {
	p := b.o.ObjectProperty(name)
	b.o.Declare(p)
	if domain != "" {
		b.o.Add(owl.ObjectPropertyDomain{Property: p, Domain: b.o.Class(domain)})
	}
	if rng != "" {
		b.o.Add(owl.ObjectPropertyRange{Property: p, Range: b.o.Class(rng)})
	}
	b.label(name, label)
	b.comment(name, comment)
}

func (b *builder) dp(name, label, domain string, rng owl.Datatype, comment string) {
	p := b.o.DataProperty(name)
	b.o.Declare(p)
	if domain != "" {
		b.o.Add(owl.DataPropertyDomain{Property: p, Domain: b.o.Class(domain)})
	}
	b.o.Add(owl.DataPropertyRange{Property: p, Range: rng})
	b.label(name, label)
	b.comment(name, comment)
}

func (b *builder) subDP(name, label, super string) {
	p := b.o.DataProperty(name)
	b.o.Declare(p)
	b.o.Add(owl.SubDataPropertyOf{Sub: p, Super: b.o.DataProperty(super)})
	b.label(name, label)
}

// claim reserves a local name for one thing, so that two names sanitising to
// the same IRI is a build failure rather than two classes silently merged.
func (b *builder) claim(name, of string) string {
	if prev, ok := b.names[name]; ok && prev != of {
		panic(fmt.Sprintf("owl: %s and %s both want %s", prev, of, name))
	}
	b.names[name] = of
	return name
}

// --- naming ----------------------------------------------------------------

func classIRI(c *ontology.Class) string { return ":" + title(c.Name) }
func traitIRI(t ontology.Trait) string  { return ":" + title(t.String()) }
func verbIRI(v ontology.Verb) string    { return ":" + title(v.String()) }
func skillIRI(s entity.Skill) string    { return ":" + title(s.String()) }

// techIRI carries a suffix because a settlement discovers fishing and a person
// is good at fishing, and those are two things: without it the skill and the
// technology are one class that is under both :Skill and :Technology, which is
// what a run of gowl classify showed. Every other kind of name here is claimed
// against the same map, so a second collision of this shape is a build failure
// rather than something to be spotted in a taxonomy.
func techIRI(t string) string         { return ":" + title(t) + "Tech" }
func roleIRI(r *ontology.Role) string { return ":" + title(r.Name) }

// actIRI turns a catalog key into a local name. The key itself stays on the
// class as its label, because the key is the real name of an act - it is what
// a habit slot is kept under - and the local name is only what an IRI takes.
var keyToName = strings.NewReplacer("@", "-at-", "+", "-and-", ">", "-to-", "<", "-from-", "/", "-")

func actIRI(key string) string { return ":" + keyToName.Replace(key) }

func transformIRI(t *ontology.Transform) string { return ":" + slug(transformLabel(t)) }

func transformLabel(t *ontology.Transform) string {
	s := t.From.Name
	if t.In != nil {
		s += " in " + t.In.Name
	}
	if t.To != nil {
		s += " to " + t.To.Name
	}
	return s
}

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func slug(s string) string { return strings.ReplaceAll(s, " ", "-") }

// dec renders a float as an xsd:decimal, which has no exponent form.
func dec(v float64) owl.Literal {
	return owl.Typed(strconv.FormatFloat(v, 'f', -1, 64), owl.XSDDecimal)
}

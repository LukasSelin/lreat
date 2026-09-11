package ontology

import (
	"sort"
	"strings"

	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/habit"
)

// Instance is one concrete act: a schema bound to the leaf classes it
// operates on. Key is its canonical name, take/timber@wood, and is what a
// habit slot is keyed by, so an instance keeps its habits across any change
// to the trees that leaves it standing.
type Instance struct {
	Key          string
	Schema       *Schema
	Object, Site *Class
	// Actor is who does it: nil for a person, as on the schema.
	Actor *Class

	Ticks   int
	Skill   entity.Skill
	Skilled bool
	Tech    string
	Reach0  float64
	Prior   habit.Signature
	Valence belief.Valence
}

// Instantiate walks the schemas over the trees and returns every act they
// entail, in order: everything a person does first, by key, and then each
// other actor's acts in the order the actors were declared. It is
// deterministic: same trees, same catalog.
func Instantiate() []Instance {
	var out []Instance
	for i := range Schemas {
		sc := &Schemas[i]
		for _, obj := range objects(sc) {
			for _, site := range sites(sc) {
				if sc.Verb == Take && !Offers(site, obj) {
					continue
				}
				out = append(out, bind(sc, obj, site))
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return Ordered(&out[i], &out[j]) })
	return out
}

// Ordered is the order the catalog is in: by actor first, people before
// every creature declared after them, and by key within an actor. Habit
// slots are catalog positions, so an actor added later takes its slots
// after everything that was already there, and nothing a person does
// moves for a deer having been thought of.
func Ordered(a, b *Instance) bool {
	if ra, rb := actorRank(a.Actor), actorRank(b.Actor); ra != rb {
		return ra < rb
	}
	return a.Key < b.Key
}

// actorRank is where an actor came in the declaring: a person for nil.
func actorRank(c *Class) int {
	if c == nil {
		return Person.id
	}
	return c.id
}

// objects are the classes a schema's object expands to: the leaves under
// it, or the class itself when collapsed, or nothing when it has none.
func objects(sc *Schema) []*Class {
	switch {
	case sc.Object == nil:
		return []*Class{nil}
	case sc.CollapseObject:
		return []*Class{sc.Object}
	}
	return sc.Object.Leaves()
}

// sites likewise. A schema placed by trait has no site class; the trait
// names the place in the key.
func sites(sc *Schema) []*Class {
	switch {
	case sc.Site == nil:
		return []*Class{nil}
	case sc.CollapseSite:
		return []*Class{sc.Site}
	}
	return sc.Site.Leaves()
}

func bind(sc *Schema, obj, site *Class) Instance {
	in := Instance{
		Schema: sc, Object: obj, Site: site, Actor: sc.Actor,
		Ticks: sc.Ticks, Skill: sc.Skill, Skilled: sc.Skilled, Tech: sc.Tech, Reach0: sc.Reach0,
		Valence: sc.Valence,
	}
	in.Key = key(sc, obj, site)
	in.Prior = Compose(sc, obj, site)
	if sc.Verb == Take {
		if d, ok := takeDetail[obj]; ok {
			in.Skill, in.Skilled, in.Reach0, in.Ticks, in.Tech = d.Skill, d.Skilled, d.Reach0, d.Ticks, d.Tech
			add(&in.Prior, d.Prior, 1)
		}
	}
	return in
}

// key renders [actor:]verb[/name][/inputs>output | /object][@site][>role | <role].
// The actor is named only where it is not a person, so that every key a
// person's act ever had is the key it has.
func key(sc *Schema, obj, site *Class) string {
	var b strings.Builder
	if sc.Actor != nil {
		b.WriteString(sc.Actor.Name + ":")
	}
	b.WriteString(sc.Verb.String())
	if sc.Name != "" {
		b.WriteString("/" + sc.Name)
	}
	switch {
	case sc.Output != nil:
		names := make([]string, len(sc.Inputs))
		for i, c := range sc.Inputs {
			names[i] = c.Name
		}
		b.WriteString("/" + strings.Join(names, "+") + ">" + sc.Output.Name)
	case obj != nil:
		b.WriteString("/" + obj.Name)
	}
	switch {
	case site != nil:
		b.WriteString("@" + site.Name)
	case sc.SiteTrait != 0:
		b.WriteString("@" + sc.SiteTrait.String())
	}
	if sc.Role != nil {
		if sc.Dir == Seize {
			b.WriteString("<")
		} else {
			b.WriteString(">")
		}
		b.WriteString(sc.Role.Name)
	}
	return b.String()
}

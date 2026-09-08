package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"gowl/owl"

	"lreat/core/habit"
	"lreat/core/ontology"
)

// Taking something out is not adding something in reverse, and the tool should
// not pretend otherwise.
//
// An addition can be applied and the catalog asked what it did. Two of the
// three removals can be too: Affords is an exported map and a class's own
// traits are an exported field, so dropping a row or a bit and instantiating
// again gives the real before and after. The third cannot. A class is attached
// to its parent through an unexported slice and core/ontology offers no way to
// detach one, which is right - nothing in a running world should be pulling
// classes out of the trees - and it leaves this with no way to ask what the
// catalog would be without one.
//
// So a class removal is answered differently, and better: not with a count but
// with everything that names it. That is the question anyway. What stops a
// class being deleted is never the size of the catalog, it is the four schemas
// that will not compile and the transform whose From is now nil, and those are
// all readable straight off the exported tables.
//
// Nothing is written either way. A removal is reported and left; the deleting
// is done by hand, in the file where the prose lives.

// dropping is one thing a proposal takes away, gathered from every axiom that
// takes part of it away.
type dropping struct {
	name string
	// exactly one of these is set.
	class   *ontology.Class // the whole class goes
	traitOf *ontology.Class // a class loses traits
	holder  *ontology.Class // a holder stops affording things

	traits     ontology.Trait
	offered    []*ontology.Class
	data       []string
	other      []owl.Axiom
	unparented bool // the link to its parent was dropped
	self       *ontology.Class

	// Everything below is read off the trees by survey, before any removal is
	// applied to them. It has to be: a proposal is usually several removals,
	// and each one applied makes the trees stop saying what the next one is
	// about to be asked. Dropping a class and the Affords row that named it
	// is one edit, and the class was left reporting no row at all.
	ownTraits       ontology.Trait
	inheritedTraits ontology.Trait
	readers         []string
	acts            []string
	schemas         []string
	tables          []string
	kids            []string
	lastChild       bool
}

// survey reads what the removal costs while the trees still say it.
func (d *dropping) survey() {
	if c := d.class; c != nil {
		d.acts = actsNaming(c)
		d.schemas = schemasNaming(c)
		d.tables = tablesNaming(c)
		for _, k := range c.Children() {
			d.kids = append(d.kids, k.Name)
		}
		d.lastChild = c.Parent != nil && len(c.Parent.Children()) == 1
	}
	if d.traitOf != nil && d.class == nil {
		d.ownTraits = d.traits & d.traitOf.Traits
		d.inheritedTraits = d.traits &^ d.traitOf.Traits
		d.readers = traitReaders(d.traits)
	}
}

// habitZero is the empty signature, to compare a derived prior against.
var habitZero habit.Signature

// gatherRemovals reads the removed axioms into one entry per thing removed. A
// class going is written as a Declaration and a handful of SubClassOf axioms;
// they are one decision and are answered as one.
func gatherRemovals(axioms []owl.Axiom) []*dropping {
	byName := map[string]*dropping{}
	var order []*dropping
	get := func(c *ontology.Class) *dropping {
		if d, ok := byName[c.Name]; ok {
			return d
		}
		d := &dropping{name: c.Name, self: c}
		byName[c.Name] = d
		order = append(order, d)
		return d
	}

	var loose []owl.Axiom
	for _, ax := range axioms {
		switch a := owl.Unwrap(ax).(type) {
		case owl.Declaration:
			if c := lreatClass(owl.Class(a.Entity.IRI())); c != nil {
				get(c).class = c
				continue
			}
		case owl.SubClassOf:
			sub, ok := a.Sub.(owl.Class)
			if !ok {
				break
			}
			c := lreatClass(sub)
			if c == nil {
				break
			}
			switch super := a.Super.(type) {
			case owl.Class:
				if t := lreatTrait(super); t != 0 {
					d := get(c)
					d.traitOf, d.traits = c, d.traits|t
					continue
				}
				// The link to its own parent. Usually that is the class
				// itself going and says nothing on its own; on its own it
				// is a class left hanging, which says a great deal.
				if c.Parent != nil && classIRI(c.Parent) == shortOf(super.IRI()) {
					get(c).unparented = true
					continue
				}
			case owl.ObjectSomeValuesFrom:
				p, _ := super.Property.(owl.ObjectProperty)
				f, _ := super.Filler.(owl.Class)
				if shortOf(p.IRI()) == ":affords" {
					if m := lreatClass(f); m != nil {
						d := get(c)
						d.holder, d.offered = c, append(d.offered, m)
						continue
					}
				}
			case owl.DataHasValue:
				p, _ := super.Property.(owl.DataProperty)
				d := get(c)
				d.data = append(d.data, strings.TrimPrefix(shortOf(p.IRI()), ":")+" "+super.Value.Value)
				continue
			}
		}
		loose = append(loose, ax)
	}
	sort.SliceStable(order, func(i, j int) bool { return order[i].name < order[j].name })

	// A class whose parent is going was written down only because its link to
	// that parent went with it. It is not a decision of its own and it is
	// already counted, under the parent, as one of the children hanging off
	// it. Left in, deleting one class reports six removals, five of them
	// blank.
	leaving := map[*ontology.Class]bool{}
	for _, d := range order {
		if d.class != nil {
			leaving[d.class] = true
		}
	}
	kept := order[:0]
	for _, d := range order {
		if d.unparented && d.self != nil && leaving[d.self.Parent] {
			d.unparented = false
		}
		if d.class != nil || d.traitOf != nil || d.holder != nil || d.unparented || len(d.data) > 0 {
			kept = append(kept, d)
		}
	}
	order = kept

	// An axiom left over because it mentions a class that is going is not
	// left over at all: a label on a deleted class, the disjointness it was
	// part of, the acts that named it. Reporting those beside the class
	// would be reporting the same removal several times over, in a form
	// nobody can act on separately.
	going := map[owl.IRI]bool{}
	for _, d := range order {
		if d.class != nil {
			going[owl.IRI(ns+title(d.class.Name))] = true
		}
	}
	var rest []owl.Axiom
	for _, ax := range loose {
		if !mentionsAny(ax, going) {
			rest = append(rest, ax)
		}
	}
	if len(rest) > 0 {
		order = append(order, &dropping{other: rest})
	}
	return order
}

func mentionsAny(ax owl.Axiom, iris map[owl.IRI]bool) bool {
	// An annotation's subject is a bare IRI rather than an entity, so it is
	// not in the signature and has to be asked for.
	if a, ok := owl.Unwrap(ax).(owl.AnnotationAssertion); ok && iris[a.Subject] {
		return true
	}
	for _, e := range owl.Signature(ax) {
		if iris[e.IRI()] {
			return true
		}
	}
	return false
}

// applyRemoval takes away what can be taken away, and reports whether it did.
// A class removal never can be; see the note at the top of this file.
func (d *dropping) applyRemoval() bool {
	var did bool
	if d.holder != nil {
		gone := map[*ontology.Class]bool{}
		for _, m := range d.offered {
			gone[m] = true
		}
		var kept []*ontology.Class
		for _, m := range ontology.Affords[d.holder] {
			if !gone[m] {
				kept = append(kept, m)
			}
		}
		ontology.Affords[d.holder] = kept
		did = true
	}
	// Only a class's own traits can go. An inherited one belongs to an
	// ancestor, and taking it off here would take it off every sibling too,
	// which is a different proposal.
	if d.traitOf != nil && d.class == nil && d.ownTraits != 0 {
		d.traitOf.Traits &^= d.ownTraits
		did = true
	}
	return did
}

func (d *dropping) report(w io.Writer) {
	switch {
	case d.class != nil:
		d.reportClass(w)
	case d.traitOf != nil:
		d.reportTraits(w)
	case d.holder != nil:
		d.reportAffords(w)
	case d.unparented:
		d.reportUnparented(w)
	case len(d.data) > 0:
		fmt.Fprintf(w, ":%s  loses %s\n", title(d.name), strings.Join(d.data, ", "))
		fmt.Fprintln(w, "  core/ontology/class.go: drop it from the declaration.")
	}
	if len(d.other) > 0 {
		fmt.Fprintln(w, "dropped, and not read as any of the above:")
		for _, ax := range d.other {
			fmt.Fprintf(w, "      %s\n", rendered(ax))
		}
	}
}

func (d *dropping) reportClass(w io.Writer) {
	c := d.class
	fmt.Fprintf(w, ":%s  (%s) would go\n", title(c.Name), c.Path())
	fmt.Fprintln(w, "  core/ontology/class.go: delete the declaration.")

	if len(d.kids) > 0 {
		fmt.Fprintf(w, "\n  It is not a leaf. %s hang off it and go too, or\n", strings.Join(d.kids, ", "))
		fmt.Fprintln(w, "  need somewhere else to hang; this counts none of what they carry.")
	}
	if d.lastChild {
		p := c.Parent
		fmt.Fprintf(w, "\n  ! %s has no other child, so it becomes a leaf again and the acts\n", p.Name)
		fmt.Fprintf(w, "    below re-key onto %s rather than simply going. Either way the keys\n", p.Name)
		fmt.Fprintln(w, "    that exist now stop existing, and their habit slots with them.")
	}

	if len(d.acts) > 0 {
		fmt.Fprintf(w, "\n  %d act(s) in the catalog name it:\n", len(d.acts))
		for _, k := range d.acts {
			fmt.Fprintf(w, "      %s\n", k)
		}
	} else {
		fmt.Fprintln(w, "\n  No act in the catalog names it.")
	}

	if len(d.schemas) > 0 {
		fmt.Fprintf(w, "\n  %d schema(s) in core/ontology/verb.go name it and will not compile:\n", len(d.schemas))
		for _, s := range d.schemas {
			fmt.Fprintf(w, "      %s\n", s)
		}
	}
	for _, line := range d.tables {
		fmt.Fprintf(w, "\n  %s\n", line)
	}
}

func (d *dropping) reportTraits(w io.Writer) {
	c := d.traitOf
	fmt.Fprintf(w, ":%s  loses %s\n", title(c.Name), d.traits)

	if d.ownTraits != 0 {
		fmt.Fprintf(w, "  core/ontology/class.go: drop %s from the traits on %s.\n", d.ownTraits, c.Name)
	}
	if d.inheritedTraits != 0 {
		for x := c.Parent; x != nil; x = x.Parent {
			if x.Traits&d.inheritedTraits != 0 {
				fmt.Fprintf(w, "\n  ! %s is not %s's own, it is %s's. Taking it off here would take\n", d.inheritedTraits, c.Name, x.Name)
				fmt.Fprintf(w, "    it off every class under %s, so nothing was applied for it.\n", x.Name)
				break
			}
		}
	}
	if len(d.readers) > 0 {
		fmt.Fprintln(w, "\n  what reads it:")
		for _, u := range d.readers {
			fmt.Fprintf(w, "      %s\n", u)
		}
	} else {
		fmt.Fprintln(w, "\n  Nothing in the ontology reads it. The simulation may still test it;")
		fmt.Fprintln(w, "  grep the trait name before deleting it outright.")
	}
}

// reportUnparented answers a class cut loose from its parent while its parent
// stays. The trees are single-inheritance and everything hangs off thing or
// site, so there is no such state to move to: either it goes somewhere else or
// it goes.
func (d *dropping) reportUnparented(w io.Writer) {
	c := d.self
	fmt.Fprintf(w, ":%s  would no longer be under :%s\n", title(c.Name), title(c.Parent.Name))
	fmt.Fprintln(w, "  Nothing in the trees can hold a class with no parent, and its parent is")
	fmt.Fprintln(w, "  staying. Say which class it hangs off instead, and this will answer for")
	fmt.Fprintln(w, "  the move; as written it is neither a removal nor a reparenting.")
	fmt.Fprintf(w, "\n  What it would stop inheriting from %s: %s\n", c.Parent.Name, inherits(c))
}

// inherits is what a class gets from above it and would lose on the way out.
func inherits(c *ontology.Class) string {
	var parts []string
	if t := c.Parent.All(); t != 0 {
		parts = append(parts, "traits "+t.String())
	}
	if l := c.Parent.Short(); l != 0 {
		parts = append(parts, fmt.Sprintf("lack %s", num(l)))
	}
	if p := c.Parent.DerivedPrior(); p != (habitZero) {
		parts = append(parts, "the prior "+sigExpr(p))
	}
	if len(parts) == 0 {
		return "nothing it does not already have of its own"
	}
	return strings.Join(parts, ", ")
}

func (d *dropping) reportAffords(w io.Writer) {
	names := make([]string, len(d.offered))
	for i, m := range d.offered {
		names[i] = title(m.Name)
	}
	fmt.Fprintf(w, ":%s  stops affording %s\n", title(d.holder.Name), strings.Join(names, ", "))
	fmt.Fprintf(w, "  core/ontology/relation.go: drop %s from %s's row in Affords.\n",
		strings.Join(names, ", "), title(d.holder.Name))
}

// --- what names a class ----------------------------------------------------

func actsNaming(c *ontology.Class) []string {
	var out []string
	for _, in := range ontology.Instantiate() {
		if in.Object == c || in.Site == c || schemaNames(in.Schema, c) {
			out = append(out, in.Key)
		}
	}
	sort.Strings(out)
	return out
}

func schemasNaming(c *ontology.Class) []string {
	var out []string
	for i := range ontology.Schemas {
		if sc := &ontology.Schemas[i]; schemaNames(sc, c) || sc.Object == c || sc.Site == c {
			out = append(out, schemaDesc(sc))
		}
	}
	return out
}

// schemaNames is about what a schema turns into rather than what it ranges
// over, so that an act naming a class through its inputs is caught even when
// the object and site are something else entirely.
func schemaNames(sc *ontology.Schema, c *ontology.Class) bool {
	if sc.Output == c || sc.Season == c {
		return true
	}
	for _, in := range sc.Inputs {
		if in == c {
			return true
		}
	}
	for _, in := range sc.Optional {
		if in == c {
			return true
		}
	}
	return false
}

func schemaDesc(sc *ontology.Schema) string {
	var b strings.Builder
	b.WriteString(sc.Verb.String())
	if sc.Name != "" {
		b.WriteString("/" + sc.Name)
	}
	if len(sc.Inputs) > 0 {
		names := make([]string, len(sc.Inputs))
		for i, c := range sc.Inputs {
			names[i] = c.Name
		}
		b.WriteString(" " + strings.Join(names, "+"))
	}
	if sc.Output != nil {
		b.WriteString(" > " + sc.Output.Name)
	}
	if sc.Object != nil {
		b.WriteString(" " + sc.Object.Name)
	}
	switch {
	case sc.Site != nil:
		b.WriteString(" @" + sc.Site.Name)
	case sc.SiteTrait != 0:
		b.WriteString(" @" + sc.SiteTrait.String())
	}
	if sc.Role != nil {
		b.WriteString(" " + sc.Role.Name)
	}
	return b.String()
}

// tablesNaming reads the rest of core/ontology for mentions: the relations,
// what happens on its own, and what grows.
func tablesNaming(c *ontology.Class) []string {
	var out []string

	var rows []string
	for holder, offered := range ontology.Affords {
		if holder == c {
			rows = append(rows, holder.Name+" (the whole row)")
			continue
		}
		for _, m := range offered {
			if m == c {
				rows = append(rows, holder.Name)
			}
		}
	}
	sort.Strings(rows)
	if len(rows) > 0 {
		out = append(out, fmt.Sprintf("Affords in relation.go names it under: %s", strings.Join(rows, ", ")))
	}

	var trs []string
	for i := range ontology.Transforms {
		t := &ontology.Transforms[i]
		if t.From == c || t.To == c || t.In == c || t.Unless == c {
			trs = append(trs, transformLabel(t))
		}
	}
	if len(trs) > 0 {
		out = append(out, fmt.Sprintf("Transforms in relation.go: %s", strings.Join(trs, ", ")))
	}

	var ps []string
	for _, p := range ontology.Processes {
		if p.Of == c || p.Yields == c {
			ps = append(ps, p.Name)
		}
	}
	if len(ps) > 0 {
		out = append(out, fmt.Sprintf("Processes in process.go: %s", strings.Join(ps, ", ")))
	}
	return out
}

// traitReaders is what in the ontology tests a trait. A trait exists only
// because some verb treats a class differently from its siblings, so a trait
// nothing reads is one that has stopped earning its place.
func traitReaders(t ontology.Trait) []string {
	var out []string
	for i := range ontology.Schemas {
		if sc := &ontology.Schemas[i]; sc.SiteTrait&t != 0 {
			out = append(out, "placed by it: "+schemaDesc(sc))
		}
	}
	for bit, s := range ontology.TraitPrior {
		if bit&t != 0 {
			out = append(out, fmt.Sprintf("TraitPrior in prior.go: wanting it adds %s", sigExpr(s)))
		}
	}
	var carriers []string
	for _, root := range []*ontology.Class{ontology.Thing, ontology.Site} {
		for _, c := range root.Family() {
			if c.Traits&t != 0 {
				carriers = append(carriers, c.Name)
			}
		}
	}
	if len(carriers) > 1 {
		out = append(out, "also carried by: "+strings.Join(carriers, ", "))
	}
	sort.Strings(out)
	return out
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"gowl/owl"

	"lreat/core/habit"
	"lreat/core/ontology"
)

// The document runs the other way here. Rendering core/ontology as OWL is a
// reading of what is already decided; this is for deciding, which is a
// different job and wants a different guard.
//
// The guard is the vocabulary. A proposal is read against the closed set of
// terms the current document defines, so :requiresTech cannot be misspelt into
// existence and a class cannot be invented by referring to it. What a proposal
// may do is declare - one Declaration and the term is yours, and everything
// after it has to stay inside the vocabulary as extended. That is the whole
// rule, and it is the difference between growing an ontology and drifting
// away from one.
//
// What comes back is not a patch. Go stays the source, because the prose in
// core/ontology carries the reasoning and no generator is going to write "the
// slowest thing the year makes". What comes back is the declaration to paste
// and, more usefully, what the trees would then entail: the catalog is
// instantiated before and after, so a proposal is answered with the acts it
// wins and the acts it costs.
//
// The costly ones are the point. Hanging a child off a class that had none
// turns a leaf into a branch, and Instantiate walks leaves - so every act
// keyed on that leaf is re-keyed under its children, which orphans the habit
// slots kept under the old keys. That is invisible in the axioms and obvious
// in the catalog, which is the whole reason for instantiating rather than
// reasoning about it.

// propose reads a proposal and reports what accepting it would mean.
func propose(path string, current *owl.Ontology, w io.Writer) error {
	axioms, removed, err := readProposal(path, current)
	if err != nil {
		return err
	}
	if len(axioms) == 0 && len(removed) == 0 {
		fmt.Fprintln(w, "nothing proposed")
		return nil
	}

	vocab := owl.NewVocabulary(current)
	fresh := declaredIn(axioms)
	if err := checkTerms(axioms, vocab, fresh, current, w); err != nil {
		return err
	}

	rendered = current.Render
	concepts := gather(axioms, fresh)
	drops := gatherRemovals(removed)
	before := keysOf(ontology.Instantiate())
	var applied int

	// Every removal is surveyed before any is applied, because applying one
	// stops the trees saying what the next is about to be asked. Dropping a
	// class and the Affords row that named it is one edit, and done in the
	// other order the class reports no row at all.
	for _, d := range drops {
		d.survey()
	}

	// Removals go first, and not only in the report. Dropping a row from
	// Affords and adding a class that wants one are the same proposal, and
	// applying them the other way round would leave the new class briefly
	// afforded by a row on its way out.
	for _, d := range drops {
		if d.applyRemoval() {
			applied++
		}
	}

	// Everything is applied to the trees before anything is reported, because
	// a concept is answered by what the trees then entail and the trees are
	// not whole until the last one is in: a material proposed in one axiom and
	// afforded in the next is one concept written in two places. Affords is a
	// second pass for the same reason, since a row may name a class an earlier
	// concept has only just created.
	for _, c := range concepts {
		if c.apply() {
			applied++
		}
	}
	for _, c := range concepts {
		c.applyAffords()
	}

	fmt.Fprintf(w, "%d axiom(s) in, %d out: %d concept(s), %d removal(s)\n",
		len(axioms), len(removed), len(concepts), len(drops))
	for _, d := range drops {
		fmt.Fprintln(w)
		d.report(w)
	}
	for _, c := range concepts {
		fmt.Fprintln(w)
		c.report(w)
	}
	if applied == 0 && !detached(drops) {
		return nil
	}

	fmt.Fprintln(w)
	reportCatalog(before, keysOf(ontology.Instantiate()), w)
	if detached(drops) {
		fmt.Fprintln(w, "\n  The catalog above does not count the class removals. A class cannot be")
		fmt.Fprintln(w, "  taken off its parent from out here, which is why they are answered")
		fmt.Fprintln(w, "  with everything that names them instead.")
	}
	return nil
}

// detached reports whether any removal is one that could not be applied.
func detached(drops []*dropping) bool {
	for _, d := range drops {
		if d.class != nil {
			return true
		}
	}
	return false
}

// readProposal accepts either a whole document, which is diffed against the
// current one, or a file of one axiom per line, which is read as additions.
// The second is what a person writes by hand; the first is what comes back
// from editing the generated document in a tool.
func readProposal(path string, current *owl.Ontology) (added, removed []owl.Axiom, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	src := string(data)
	if isDocument(src) {
		authored, err := owl.ParseFunctionalString(src)
		if err != nil {
			return nil, nil, err
		}
		d := owl.DiffOntologies(current, authored)
		return d.AddedAxioms, d.RemovedAxioms, nil
	}

	var out []owl.Axiom
	s := bufio.NewScanner(strings.NewReader(src))
	for line := 1; s.Scan(); line++ {
		text := strings.TrimSpace(s.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		ax, err := owl.ParseAxiom(text, current.Prefixes)
		if err != nil {
			return nil, nil, fmt.Errorf("%s:%d: %w", path, line, err)
		}
		out = append(out, ax)
	}
	// A line-oriented proposal has no way to say what it takes away, and
	// should not: a removal is a thing to be looked at against the whole
	// document, which is what the document form is for.
	return out, nil, s.Err()
}

func isDocument(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return strings.HasPrefix(line, "Prefix(") || strings.HasPrefix(line, "Ontology(")
	}
	return false
}

// declaredIn is the set of terms the proposal introduces. Declaring is the one
// way past the vocabulary, and it has to be deliberate: a term used without
// being declared is a misspelling far more often than it is an intention.
func declaredIn(axioms []owl.Axiom) map[owl.IRI]owl.Entity {
	out := map[owl.IRI]owl.Entity{}
	for _, ax := range axioms {
		if d, ok := owl.Unwrap(ax).(owl.Declaration); ok {
			out[d.Entity.IRI()] = d.Entity
		}
	}
	return out
}

func checkTerms(axioms []owl.Axiom, v *owl.Vocabulary, fresh map[owl.IRI]owl.Entity, o *owl.Ontology, w io.Writer) error {
	var bad int
	for _, ax := range axioms {
		for _, e := range v.Unknown(ax) {
			if _, ok := fresh[e.IRI()]; ok || owl.IsBuiltin(e) {
				continue
			}
			bad++
			name := shortOf(e.IRI())
			fmt.Fprintf(w, "unknown %s %s\n", strings.ToLower(e.Kind().String()), name)
			fmt.Fprintf(w, "  in %s\n", o.Render(ax))
			if near := nearby(v, e); len(near) > 0 {
				fmt.Fprintf(w, "  did you mean %s?\n", strings.Join(near, ", "))
			}
			fmt.Fprintln(w, "  or declare it, if it is meant to be new:")
			fmt.Fprintf(w, "      Declaration(%s(%s))\n", e.Kind(), name)
		}
	}
	if bad > 0 {
		return fmt.Errorf("%d unknown term(s); nothing applied", bad)
	}
	return nil
}

// nearby offers terms of the same kind that a name is plausibly a slip for.
// gowl's own Resolve suggests only across case and punctuation, deliberately,
// so this covers the rest of the near misses: one name sitting inside the
// other, or the two opening alike.
func nearby(v *owl.Vocabulary, e owl.Entity) []string {
	want := fold(shortOf(e.IRI()))
	var out []string
	for _, t := range v.TermsOfKind(e.Kind()) {
		got := fold(t.Name)
		switch {
		case strings.Contains(got, want), strings.Contains(want, got):
		case len(want) >= 4 && len(got) >= 4 && got[:4] == want[:4]:
		default:
			continue
		}
		if out = append(out, t.Name); len(out) == 3 {
			break
		}
	}
	return out
}

func fold(s string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(s), ":"))
}

func shortOf(i owl.IRI) string {
	if n := strings.LastIndexAny(string(i), "#/"); n >= 0 {
		return ":" + string(i)[n+1:]
	}
	return string(i)
}

// concept is one proposed thing, gathered from every axiom about it. A
// proposal names a concept in as many axioms as it likes and they are answered
// together, because one class is one decision however it is written down.
type concept struct {
	name    string // local name, without the colon
	entity  owl.Entity
	parent  owl.Class
	traits  []owl.Class
	data    map[string]float64
	affords []owl.Class // what this holder offers
	other   []owl.Axiom // axioms about it this does not read

	// what apply worked out.
	added       *ontology.Class
	parentClass *ontology.Class
	existing    *ontology.Class // named a class the trees already have
	wasLeaf     bool
	newTraits   ontology.Trait // the ones existing did not already have
	newAffords  []*ontology.Class
}

func gather(axioms []owl.Axiom, fresh map[owl.IRI]owl.Entity) []*concept {
	byIRI := map[owl.IRI]*concept{}
	var order []*concept
	get := func(i owl.IRI, e owl.Entity) *concept {
		if c, ok := byIRI[i]; ok {
			return c
		}
		c := &concept{name: strings.TrimPrefix(shortOf(i), ":"), entity: e, data: map[string]float64{}}
		byIRI[i] = c
		order = append(order, c)
		return c
	}
	for i, e := range fresh {
		get(i, e)
	}

	for _, ax := range axioms {
		sc, ok := owl.Unwrap(ax).(owl.SubClassOf)
		if !ok {
			continue
		}
		sub, ok := sc.Sub.(owl.Class)
		if !ok {
			continue
		}
		// An axiom about an existing class is a change to that class, and is
		// gathered under it the same way.
		c := get(sub.IRI(), sub)
		switch super := sc.Super.(type) {
		case owl.Class:
			if isTrait(super) {
				c.traits = append(c.traits, super)
			} else {
				c.parent = super
			}
		case owl.ObjectSomeValuesFrom:
			p, _ := super.Property.(owl.ObjectProperty)
			f, _ := super.Filler.(owl.Class)
			if shortOf(p.IRI()) == ":affords" {
				c.affords = append(c.affords, f)
			} else {
				c.other = append(c.other, ax)
			}
		case owl.DataHasValue:
			p, _ := super.Property.(owl.DataProperty)
			v, err := strconv.ParseFloat(super.Value.Value, 64)
			if err != nil {
				c.other = append(c.other, ax)
				continue
			}
			c.data[strings.TrimPrefix(shortOf(p.IRI()), ":")] = v
		default:
			c.other = append(c.other, ax)
		}
	}
	sort.Slice(order, func(i, j int) bool { return order[i].name < order[j].name })
	return order
}

// apply puts the concept into the live trees, so the catalog can be
// instantiated as it would be if the proposal were accepted. The process is
// about to exit and nothing else reads the trees after this.
func (c *concept) apply() bool {
	if parent := lreatClass(c.parent); parent != nil {
		c.parentClass = parent
		c.wasLeaf = len(parent.Children()) == 0
		added := ontology.New(strings.ToLower(c.name), parent, c.traitMask(), c.prior("prior-"))
		if l, ok := c.data["lack"]; ok {
			added.Lack = l
		}
		if wm, ok := c.data["warmth"]; ok {
			added.Warmth = wm
			added.Prior[habit.Chill] = -wm
		}
		if at := c.prior("at-"); at != (habit.Signature{}) {
			added.At = at
		}
		c.added = added
		// A class the proposal declares has to be findable by the concepts
		// after it: that is how one axiom adds a material and the next says
		// what affords it.
		classByName[classIRI(added)] = added
		return true
	}
	// Not a new class, then, but perhaps a change to one the trees have: a
	// trait it should have had, a row in Affords, a number retuned.
	if existing := lreatClass(owl.Class(c.entity.IRI())); existing != nil {
		c.existing = existing
		c.newTraits = c.traitMask() &^ existing.All()
		existing.Traits |= c.newTraits
		return c.newTraits != 0 || len(c.affords) > 0
	}
	return false
}

func (c *concept) applyAffords() {
	holder := c.existing
	if holder == nil {
		holder = c.added
	}
	if holder == nil {
		return
	}
	for _, a := range c.affords {
		m := lreatClass(a)
		if m == nil || ontology.Offers(holder, m) {
			continue
		}
		c.newAffords = append(c.newAffords, m)
		ontology.Affords[holder] = append(ontology.Affords[holder], m)
	}
}

// report says what one concept costs.
func (c *concept) report(w io.Writer) {
	switch {
	case c.added != nil:
		c.reportClass(w)
	case c.existing != nil:
		c.reportExisting(w)
	default:
		fmt.Fprintf(w, ":%s\n", c.name)
		fmt.Fprintln(w, "  Not a class under either tree, so there is no declaration to paste.")
		fmt.Fprintf(w, "  %s\n", whereItWouldGo(c))
	}
}

func (c *concept) reportClass(w io.Writer) {
	parent, added := c.parentClass, c.added
	root := "Things"
	if parent.IsA(ontology.Site) {
		root = "Sites"
	}

	fmt.Fprintf(w, ":%s  (class, under :%s)\n", c.name, title(parent.Name))
	fmt.Fprintf(w, "  core/ontology/class.go, in the %s block:\n\n", root)
	fmt.Fprintf(w, "      %s\n", c.goDecl(parent))

	// A proposal that adds a material usually means it to be got somewhere,
	// and saying where is not optional: Take instantiates one act per
	// (material, ground) pair in Affords, and none otherwise.
	if added.IsA(ontology.Material) && !gettable(added) {
		fmt.Fprintf(w, "\n  Nothing affords it, so no act can get it. Add a row to Affords in\n")
		fmt.Fprintf(w, "  core/ontology/relation.go, or say so in the proposal:\n")
		fmt.Fprintf(w, "      SubClassOf(:Wood ObjectSomeValuesFrom(:affords :%s))\n", c.name)
	}
	if c.wasLeaf {
		fmt.Fprintf(w, "\n  ! %s had no children of its own. Adding one makes it a branch, and\n", parent.Name)
		fmt.Fprintf(w, "    Instantiate walks leaves, so every act keyed on %s is re-keyed\n", parent.Name)
		fmt.Fprintf(w, "    under its children. See the catalog below.\n")
	}
	for _, ax := range c.other {
		fmt.Fprintf(w, "\n  not read: %s\n", rendered(ax))
	}
}

// reportExisting answers a proposal about a class the trees already have. The
// common case is that it says nothing new, and saying so plainly is worth more
// than an edit nobody needs to make.
func (c *concept) reportExisting(w io.Writer) {
	e := c.existing
	fmt.Fprintf(w, ":%s  (%s, already in the trees)\n", c.name, e.Path())
	if c.newTraits == 0 && len(c.newAffords) == 0 {
		fmt.Fprintln(w, "  The trees already say this. Nothing to change.")
		return
	}
	if c.newTraits != 0 {
		fmt.Fprintf(w, "  core/ontology/class.go, on %s:\n\n", e.Name)
		fmt.Fprintf(w, "      add %s to its traits\n", c.newTraits)
	}
	if len(c.newAffords) > 0 {
		names := make([]string, len(c.newAffords))
		for i, m := range c.newAffords {
			names[i] = title(m.Name)
		}
		fmt.Fprintf(w, "  core/ontology/relation.go, in Affords:\n\n")
		fmt.Fprintf(w, "      %s: {..., %s},\n", title(e.Name), strings.Join(names, ", "))
	}
}

// goDecl renders the declaration to paste, wrapped the way class.go wraps
// them: New innermost, with lack, season and at outside it.
func (c *concept) goDecl(parent *ontology.Class) string {
	expr := fmt.Sprintf("New(%q, %s, %s, %s)",
		strings.ToLower(c.name), title(parent.Name), traitExpr(c.traits), sigExpr(c.prior("prior-")))
	if l, ok := c.data["lack"]; ok {
		expr = fmt.Sprintf("lack(%s, %s)", expr, num(l))
	}
	if wm, ok := c.data["warmth"]; ok {
		expr = fmt.Sprintf("season(%s, %s)", expr, num(wm))
	}
	if at := c.prior("at-"); at != (habit.Signature{}) {
		expr = fmt.Sprintf("at(%s, %s)", expr, sigExpr(at))
	}
	return title(c.name) + " = " + expr
}

func (c *concept) traitMask() ontology.Trait {
	var t ontology.Trait
	for _, tc := range c.traits {
		t |= lreatTrait(tc)
	}
	return t
}

// prior reads one family of coordinate properties back into a signature.
func (c *concept) prior(family string) habit.Signature {
	var s habit.Signature
	for k, v := range c.data {
		for i, n := range habit.Names {
			if k == family+n {
				s[i] = v
			}
		}
	}
	return s
}

// --- rendering Go ----------------------------------------------------------

func traitExpr(traits []owl.Class) string {
	if len(traits) == 0 {
		return "0"
	}
	names := make([]string, 0, len(traits))
	for _, t := range traits {
		names = append(names, title(strings.TrimPrefix(shortOf(t.IRI()), ":")))
	}
	sort.Strings(names)
	return strings.Join(names, "|")
}

func sigExpr(s habit.Signature) string {
	var parts []string
	for i, v := range s {
		if v != 0 {
			parts = append(parts, fmt.Sprintf("habit.%s: %s", title(habit.Names[i]), num(v)))
		}
	}
	return "habit.Signature{" + strings.Join(parts, ", ") + "}"
}

func num(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }

// whereItWouldGo names the file for the kinds of thing this does not paste.
// Everything below the line is a Go declaration with prose attached, and the
// prose is the part worth writing by hand.
func whereItWouldGo(c *concept) string {
	switch shortOf(c.parent.IRI()) {
	case ":Act":
		return "Acts are instantiated, not declared: add a Schema to core/ontology/verb.go and the trees will expand it over their leaves."
	case ":Transform":
		return "Add a row to Transforms in core/ontology/relation.go."
	case ":Process", ":Stage":
		return "Add a Process to core/ontology/process.go; a stage is one of its Stages."
	case ":Verb":
		return "Add a Verb constant in core/ontology/verb.go and its moment to VerbPrior in prior.go."
	case ":Role":
		return "Add a Role var in core/ontology/class.go."
	case ":Skill":
		return "Skills are core/entity's, not the ontology's: add one to entity.Skill."
	case ":Technology":
		return "A technology is a string on a Schema: set Tech on one in core/ontology/verb.go."
	}
	return "Nothing says what it is under, so nothing here can say where it goes."
}

// --- the catalog before and after ------------------------------------------

// rendered is set once, so that an axiom echoed back reads the way the
// proposal wrote it rather than as a line of full IRIs.
var rendered = owl.Functional

func keysOf(in []ontology.Instance) map[string]bool {
	out := make(map[string]bool, len(in))
	for _, i := range in {
		out[i.Key] = true
	}
	return out
}

func reportCatalog(before, after map[string]bool, w io.Writer) {
	var gained, lost []string
	for k := range after {
		if !before[k] {
			gained = append(gained, k)
		}
	}
	for k := range before {
		if !after[k] {
			lost = append(lost, k)
		}
	}
	sort.Strings(gained)
	sort.Strings(lost)

	fmt.Fprintf(w, "catalog: %d acts before, %d after\n", len(before), len(after))
	if len(gained) == 0 && len(lost) == 0 {
		fmt.Fprintln(w, "  no change")
		return
	}
	for _, k := range lost {
		fmt.Fprintf(w, "  - %s\n", k)
	}
	for _, k := range gained {
		fmt.Fprintf(w, "  + %s\n", k)
	}
	if len(lost) > 0 {
		fmt.Fprintf(w, "\n  A key that goes takes its habit slot with it: everyone who had learned\n")
		fmt.Fprintf(w, "  the act keeps a slot nothing reads, and whatever replaced it starts\n")
		fmt.Fprintf(w, "  unlearned for the whole settlement.\n")
	}
}

// --- looking lreat things up by their rendered name ------------------------

var (
	classByName = map[string]*ontology.Class{}
	traitByName = map[string]ontology.Trait{}
)

func init() {
	for _, root := range []*ontology.Class{ontology.Thing, ontology.Site} {
		for _, c := range root.Family() {
			classByName[classIRI(c)] = c
		}
	}
	for bit := ontology.Trait(1); bit != 0; bit <<= 1 {
		if bit.String() != "" {
			traitByName[traitIRI(bit)] = bit
		}
	}
}

func lreatClass(c owl.Class) *ontology.Class {
	if c == "" {
		return nil
	}
	return classByName[shortOf(c.IRI())]
}

func lreatTrait(c owl.Class) ontology.Trait { return traitByName[shortOf(c.IRI())] }

func isTrait(c owl.Class) bool { _, ok := traitByName[shortOf(c.IRI())]; return ok }

// gettable reports whether any ground affords m, which is the condition Take
// instantiates on. A person and the market afford Material in general - that
// is how a pack and a shelf hold anything - so asking whether m is afforded
// at all would answer yes for everything and say nothing.
func gettable(m *ontology.Class) bool {
	for holder := range ontology.Affords {
		if holder.IsA(ontology.Ground) && ontology.Offers(holder, m) {
			return true
		}
	}
	return false
}

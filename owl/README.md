# owl

`core/ontology` said again in OWL 2, so that what the simulation is made of can
be read, queried and reasoned over by something other than the simulation.

[`lreat.ofn`](lreat.ofn) is the document, in OWL 2 functional-style syntax. It
is generated, not written: every axiom in it is read off the Go declarations at
run time, so the way to change it is to change `core/ontology` and run this
again.

```bash
go run ./owl                       # rewrite owl/lreat.ofn
go run ./owl -o - | less
go run ./owl -lint                 # what gowl makes of it
go run ./owl -propose new.ofn      # what accepting a change would cost
```

`go test ./owl` compares the two sides rather than the document against a
golden copy: a class added to the trees, an act the catalog gains, a row added
to `Affords` — each has to come out of the document again, and the checked-in
copy has to match what a fresh run produces.

## What crosses over, and how

| in `core/ontology` | in the document |
|---|---|
| a `Class` in either tree | a class, under its parent |
| a `Trait` bit | a class of the things that have it, so `Has` becomes subsumption |
| the leaf classes under one parent | `DisjointClasses`: a thing is one of them, not two |
| `Affords` | `SubClassOf(:Wood ObjectSomeValuesFrom(:affords :Berries))` |
| a `Role` | a class under `:Role`, still outside the `Person` tree |
| an instantiated act | a class under `:Act`, labelled with its catalog key |
| `Schema` fields | `:verb`, `:handles`, `:at`, `:consumes`, `:yields`, `:toward`, … |
| a `habit.Signature` | one `:prior-<coordinate>` data value per nonzero coordinate |
| a `belief.Valence` | one `:valence-<norm>` data value per nonzero norm |
| a `Process` and its stages | a class under `:Process`, one under `:Stage` per stage |
| a `Transform` | a class under `:Transform` |

Acts are the instantiated catalog rather than the schemas, because a schema is
a statement over classes and the trees are what make it concrete. Adding a
class adds every act the trees entail for it, here as there.

The document is in OWL 2 EL, so `gowl`'s classifier will take it:

```bash
gowl classify owl/lreat.ofn   # 142 classes, 169 inferred, consistent, coherent
gowl stats owl/lreat.ofn
```

It classifies the taxonomy and passes over the numbers. `gowl/el` implements
the EL fragment without data properties or property ranges, so the ~450
`DataHasValue`, `DataPropertyRange` and `ObjectPropertyRange` axioms come back
under `Unsupported` rather than being reasoned over — every `:prior-*`,
`:ticks` and `:rate` among them. That costs completeness and not soundness:
what it derives holds, and a subsumption it does not derive is not settled.
Nothing skipped could move the taxonomy here anyway, since no class in this
document is defined by a number.

The tests classify it too, and fail on an incoherence — a class that ends up
under two disjoint siblings, or with two traits that cannot both hold, is
unsatisfiable there and invisible in a bitfield here.

Running it is also how the collision behind `techIRI` turned up: a settlement
discovers fishing and a person is good at fishing, and until the technology
took a suffix those were one class, sitting under both `:Skill` and
`:Technology`. Every generated name is now claimed against one map, so the
next collision of that shape is a build failure instead.

## Thinking in the ontology first

`-propose` runs the other way. You write what you want in OWL and it answers
with what the world would then be — which is the order you want when the
concept is not settled yet, because the axioms are quicker to write than the
Go and much quicker to throw away.

A proposal is a file of axioms, one per line, or a whole edited document.

```
Declaration(Class(:Pitch))
SubClassOf(:Pitch :Material)
SubClassOf(:Pitch :Burnable)
SubClassOf(:Pitch DataHasValue(:lack "0.5"^^xsd:decimal))
SubClassOf(:Wood ObjectSomeValuesFrom(:affords :Pitch))
```

```
:Pitch  (class, under :Material)
  core/ontology/class.go, in the Things block:

      Pitch = lack(New("pitch", Material, Burnable, habit.Signature{}), 0.5)

:Wood  (thing/site/ground/wood, already in the trees)
  core/ontology/relation.go, in Affords:

      Wood: {..., Pitch},

catalog: 31 acts before, 32 after
  + take/pitch@wood
```

Two things make this worth running rather than reasoning about.

**The vocabulary is closed.** Every term has to be one the document already
defines, or be declared in the proposal itself. `:Materal` is refused with
`did you mean :Material?` and nothing is applied — a proposal half accepted is
worse than one turned down. Declaring is the one way past it, and it has to be
deliberate, because a term used without one is a slip far more often than an
intention.

**The catalog is instantiated, not predicted.** The proposal is applied to the
trees and `Instantiate` is run again, so the answer is the real before and
after. That is how the expensive edits show themselves:

```
Declaration(Class(:Granite))
SubClassOf(:Granite :Stone)
```
```
  ! stone had no children of its own. Adding one makes it a branch, and
    Instantiate walks leaves, so every act keyed on stone is re-keyed
    under its children.

catalog: 31 acts before, 31 after
  - take/stone@outcrop
  + take/granite@outcrop

  A key that goes takes its habit slot with it: everyone who had learned
  the act keeps a slot nothing reads, and whatever replaced it starts
  unlearned for the whole settlement.
```

Nothing in the axioms says that. It is a fact about `Leaves`, and the only way
to see it is to run the thing.

What comes back is a declaration to paste, not a patch. Go stays the source:
the prose in `core/ontology` carries the reasoning, and no generator is going
to write *the slowest thing the year makes*.

### Taking things out

Edit the generated document, delete lines, and pass the whole file. Removals
are diffed out of it — a line-oriented proposal has no way to say what it
takes away, and should not, since a deletion is a thing to weigh against the
whole document rather than in isolation.

Taking a row out of `Affords` or a trait off a class is applied, so the
catalog answers for real:

```
:Wood  stops affording Timber
  core/ontology/relation.go: drop Timber from Wood's row in Affords.

catalog: 31 acts before, 30 after
  - take/timber@wood
```

Deleting a whole class is not applied, and cannot be: a class hangs off its
parent through an unexported slice and `core/ontology` offers no way to detach
one — which is right, since nothing in a running world should be pulling
classes out of the trees. So it is answered with everything that names it,
which is the better answer anyway. What stops a class being deleted is never
the size of the catalog; it is the schemas that will no longer compile.

```
:Coin  (thing/material/coin) would go
  core/ontology/class.go: delete the declaration.

  2 act(s) in the catalog name it:
      exchange/coin>provision@market
      exchange/material>coin@market

  2 schema(s) in core/ontology/verb.go name it and will not compile:
      exchange material > coin @market
      exchange coin > provision @market
```

Changing which class a `SubClassOf` points at is a move, and is read as one
rather than as a deletion and a declaration:

```
:Fish  would move from :Provision to :Game
  core/ontology/class.go: change the parent in fish's declaration.

  it stops inheriting: traits edible, lack 0.8
  it starts inheriting: traits edible+perishable, lack 0.8

  ! game had no children of its own, so it becomes a branch and the acts
    keyed on game re-key under fish.
```

What a class inherits is the interesting half: traits, lack and prior all come
from above it, so a move changes what a thing *is* before it changes where it
sits. A move cannot be applied either, for the same reason a deletion cannot,
and the catalog says so rather than counting it.

It reads `Affords`, `Transforms` and `Processes` for mentions too, warns when
the parent is left with no other child (the leaf case again, running the other
way), and says when a trait nothing in the ontology reads has stopped earning
its place — *a class exists only if some verb treats it differently from its
siblings*, and the same is true of a trait.

## Why this is its own module

lreat has no dependency beyond the terminal library, and `go build ./...` at
the repository root walks past a nested module, so nothing in the simulation
gains a dependency from this directory existing.

[gowl](https://github.com/LukasSelin/gowl) declares its module path as `gowl`
rather than as a fetchable one, so it is reached by a `replace` and is expected
to sit beside lreat:

```
repos/
  gowl/
  lreat/
    owl/
```

Point the `replace` in [go.mod](go.mod) somewhere else if it does not.

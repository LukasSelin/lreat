# owl

`core/ontology` said again in OWL 2, so that what the simulation is made of can
be read, queried and reasoned over by something other than the simulation.

[`lreat.ofn`](lreat.ofn) is the document, in OWL 2 functional-style syntax. It
is generated, not written: every axiom in it is read off the Go declarations at
run time, so the way to change it is to change `core/ontology` and run this
again.

```bash
go run ./owl            # rewrite owl/lreat.ofn
go run ./owl -o - | less
go run ./owl -lint      # what gowl makes of it
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
gowl classify owl/lreat.ofn
gowl stats owl/lreat.ofn
```

The tests classify it too, and fail on an incoherence — a class that ends up
under two disjoint siblings, or with two traits that cannot both hold, is
unsatisfiable there and invisible in a bitfield here.

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

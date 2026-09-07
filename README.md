# lreat

A simulation of people living in a world. A settlement is seeded with a handful
of figures on a generated map, and everything after that — fields, houses,
roads, trade, friendships, feuds, marriages, children, funerals — is what the
figures do rather than something the simulation stages.

There is no player yet, and no learning anywhere in it. The three commands here
are the ways to watch a settlement and to judge whether a change to it helped.

## Running it

Go 1.25 or newer. No dependencies beyond the terminal library.

Watch a settlement develop live:

```bash
go run ./cmd/watch
```

Space pauses, `+` and `-` change speed, `.` steps once while paused, `r` lays
streets, `tab` and `shift-tab` (or a click) pick a figure out of the crowd and
open it up beside the map — who it is, what it is good at, what errand it is on,
and everything it weighed before setting out. `esc` drops it, `q` quits.

`d` swaps the map for the settlement's vital record, which is how a run that
ended is read rather than guessed at: the population curve coloured by how well
fed it was at the time, a ribbon under it saying which way each stretch of the
run was going, what people died of, how the generations are shaped, what there
is to live on — and the fertility funnel, which is every living figure counted
under the first thing standing between it and a child. A settlement that starved
and one that simply stopped bearing look identical on the map and nothing alike
there.

Run a settlement with no display and print how it went:

```bash
go run ./cmd/headless -ticks 5000 -every 250 -map
```

Run a batch of seeds and compare the outcomes, which is how one set of priors is
judged against another — a single seed is a coin toss:

```bash
go run ./cmd/tune -seeds 24 -ticks 6000
```

What that batch prints on master is checked in at
[docs/baseline.md](docs/baseline.md), so a change is judged against a file
rather than against a batch run to find out where master already stood. Take the
one run that measures the change once it is settled, and refresh the file in the
commit that lands it. The `baseline` skill in
[.claude/skills](.claude/skills/baseline/SKILL.md) is that procedure written
out, including which of the numbers are steady enough to believe.

All three take `-seed` and `-agents`. A seed plus a command log reproduces a run
exactly, however the goroutines happen to interleave.

```bash
go test ./...
```

## How choosing works

An agent does not price its options. Every action carries a *signature* in a
20-dimensional space — the kind of moment it belongs to — and every decision
builds a point in that same space out of how urgent each need is, what the agent
has on hand, who is nearby, what it holds to be right, and how the candidate
action sits relative to all of it. The agent does whatever signature lies closest
in direction to the moment at hand. No value is computed anywhere in it.

Signatures never move. An agent is born with the ones its parents had, receives a
teacher's when it is taught, and keeps them; how an action turned out changes
nothing about the next one. That makes the prior table the whole of the design:
to change behaviour, change what a moment is written to call for.

The original expected-value rule is still there behind `-value` (or
`world.Rules.Fit = false`) and both rules stay under test.
[docs/action-space.md](docs/action-space.md) is the full account, including what
the removed reinforcement-learning layer used to do and what the priors have to
carry on their own now.

## Layout

| package | what it holds |
|---|---|
| `core/world` | the complete simulation state: terrain, climate, roads, growth, erosion |
| `core/system` | the per-tick rules, each a function over the world, run in a fixed order |
| `core/entity` | agents and the plain-struct components they carry |
| `core/need` | the leaky hierarchy of needs that drives every agent |
| `core/habit` | the space in which a moment is recognised; a leaf package |
| `core/action` | the catalog of what can be done, and the priors behind it |
| `core/ontology` | what the world is made of, arranged so that what can be done follows from what it is |
| `core/belief` | what agents think is true and what they think is right — neither guaranteed to match the world |
| `core/event` | the append-only record of everything that happened |
| `core/observe` | the read side: snapshots for renderers, and the perception filter |
| `core/sim` | runs a world on its own goroutine; the only place wall-clock time lives |
| `ui/ascii` | the terminal rendering |
| `cmd/` | `watch`, `headless`, `tune` |

Exactly one goroutine drives a world. Deciding is spread over several — it only
reads — while everything that changes the world runs one at a time.

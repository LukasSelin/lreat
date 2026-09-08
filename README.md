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

That opens the start screen rather than a settlement. `start` founds one on
the defaults straight away; `options` is the terms it would be founded on —
the seed, how many figures, how much ground, what speed to begin at, and
whether anyone chooses by recognition or by expected value — each written out
in a line under the one being looked at, arrows to change it or a number
typed straight in, `r` for a fresh seed, `d` to put everything back, and a
`start` at the foot of the page so a tuned run is founded from where it was
tuned. The flags are still there and still mean the same things; they are
what the menu opens on now, and `-start` skips the menu for a run launched by
a script rather than by hand.

The ground is taken off the window by default. The `map` line says so —
`fit to terminal` — and the width and height under it are then read rather
than set: the largest map that still leaves room for the panel beside it and
the graph under it, measured when the settlement is founded. A map bigger
than the window shows an apology instead of a settlement and one much smaller
wastes ground, and neither number is knowable before the program has looked
at the terminal it was started in. Set the line to `by hand` for a map of a
stated size; `-fit=false`, or simply giving `-width` or `-height`, is the same
from the command line. Nothing that has to be reproduced is founded here —
`headless` and `tune`, which the baseline is taken on, both keep the fixed
default size.

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

`w` is the same for the ground rather than the people: the wood taken out of
the map against what was there to begin with, the fields and houses and roads
that went up, what food cost and how much of it was kept, what the settlement
knew, how equally it held what it had — each drawn over the whole run, so that
every figure is read against its own past — and under them the whole run's
weave of what it has been spending its people on, with everything it ever
worked out and the tick it got there.

Both pages are also printed in plain words when the run ends, so a settlement
that is over can still be reported on rather than only restarted.

Run a settlement with no display and print how it went:

```bash
go run ./cmd/headless -ticks 18000 -every 1800 -map
```

Run a batch of seeds and compare the outcomes, which is how one set of priors is
judged against another — a single seed is a coin toss:

```bash
go run ./cmd/tune -seeds 24 -ticks 21600
```

What that batch prints on master is checked in at
[docs/baseline.md](docs/baseline.md), so a change is judged against a file
rather than against a batch run to find out where master already stood. Take the
one run that measures the change once it is settled, and refresh the file in the
commit that lands it. The `baseline` skill in
[.claude/skills](.claude/skills/baseline/SKILL.md) is that procedure written
out, including which of the numbers are steady enough to believe.

All three take `-seed` and `-agents`, and count in ticks, which are days: the
defaults above are fifty and sixty years. A seed plus a command log reproduces a
run exactly, however the goroutines happen to interleave.

```bash
go test ./...
```

## How time works

A tick is a day. That is the only convention, and `core/clock` is the only place
that says so: a week is seven days, a month thirty, a season ninety, a year four
seasons. Everything with a duration is written against those units rather than as
a bare number of ticks — a request stands eighty days, an unkept house falls in
about three years, a body is grown at five and old at fifty — so the calendar can
be changed in one place and the world follows it.

It was not always so. A year used to be a hundred ticks and a season
twenty-five, which is shorter than a walk across the map: a settler who set out
for a plot in the spring arrived in the autumn, and nobody could sow, cut and
store within a season because a season was shorter than an errand. Seasons that
nobody can act inside are weather, not seasons. A season is ninety days now, so
it holds a journey, a harvest, and the storing of one.

A life is human, and the calendar is what made it affordable. A childhood is
fifteen years, the bearing years run to forty, and a body that is kept has given
out by sixty-five. Under the old hundred-tick year a life had to be counted in
ticks and came out at five, twenty-eight and fifty-two — a settlement whose
children were grown before they could walk to the next field. What a run has to
cover is a number of generations, and a generation is a span of years, so
lengthening the year is what buys a childhood.

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
| `core/clock` | the calendar: a tick is a day, and every duration in the world is said against it |
| `core/world` | the complete simulation state: terrain, climate, roads, growth, erosion |
| `core/system` | the rules of a day, each a function over the world, run in a fixed order |
| `core/entity` | agents and the plain-struct components they carry |
| `core/need` | the leaky hierarchy of needs that drives every agent |
| `core/habit` | the space in which a moment is recognised; a leaf package |
| `core/action` | the catalog of what can be done, and the priors behind it |
| `core/ontology` | what the world is made of, arranged so that what can be done follows from what it is — and what happens to it with nobody doing anything |
| `core/belief` | what agents think is true and what they think is right — neither guaranteed to match the world |
| `core/event` | the append-only record of everything that happened |
| `core/observe` | the read side: snapshots for renderers, and the perception filter |
| `core/sim` | runs a world on its own goroutine; the only place wall-clock time lives |
| `ui/ascii` | the terminal rendering |
| `cmd/` | `watch`, `headless`, `tune` |

Exactly one goroutine drives a world. Deciding is spread over several — it only
reads — while everything that changes the world runs one at a time.

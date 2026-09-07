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

Run a settlement with no display and print how it went:

```bash
go run ./cmd/headless -ticks 18000 -every 1800 -map
```

Run a batch of seeds and compare the outcomes, which is how one set of priors is
judged against another — a single seed is a coin toss:

```bash
go run ./cmd/tune -seeds 24 -ticks 21600
```

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

The one place the calendar is deliberately compressed is a life. Five years to
grow up and fifty-two to a lifetime is not a human childhood; it is what lets a
run somebody will sit through turn a settlement over three or four times, which
is the only way to see whether what one generation believed outlived it.

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
| `core/ontology` | what the world is made of, arranged so that what can be done follows from what it is |
| `core/belief` | what agents think is true and what they think is right — neither guaranteed to match the world |
| `core/event` | the append-only record of everything that happened |
| `core/observe` | the read side: snapshots for renderers, and the perception filter |
| `core/sim` | runs a world on its own goroutine; the only place wall-clock time lives |
| `ui/ascii` | the terminal rendering |
| `cmd/` | `watch`, `headless`, `tune` |

Exactly one goroutine drives a world. Deciding is spread over several — it only
reads — while everything that changes the world runs one at a time.

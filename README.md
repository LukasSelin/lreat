# lreat

A simulation of people living in a world. A settlement is seeded with a handful
of figures on a generated map, and everything after that — fields, houses,
roads, trade, friendships, feuds, marriages, children, funerals — is what the
figures do rather than something the simulation stages.

There is no player yet, and no learning anywhere in it. The three commands here
are the ways to watch a settlement and to judge whether a change to it helped.

## Running it

Go 1.27 or newer. No dependencies beyond the terminal library.

The day's pass over the ground can be done four tiles at a time on a
processor with AVX2, through Go's experimental vector packages. They exist
only when the go command is asked for them, so the build has to be asked
too:

```bash
GOEXPERIMENT=simd go run ./cmd/watch
```

(In PowerShell, `$env:GOEXPERIMENT = 'simd'` first.) What a settlement does
is the same either way, to the last bit; a test holds the vector arithmetic
to the tile-by-tile statement of it, and the golden numbers and the batch
come out identical. What it changes is how long a day takes, and so far on
the machine it was measured on that is nothing worth having: the arithmetic
over a row of ground is twice as quick, but a day spends its time reading
the tiles and drawing the hedges, not on the arithmetic. It is there for
when that is no longer true.

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
from the command line. The `world` line above it chooses what is being
founded at all — the valley, which has edges and is sized to whoever is
watching, or the globe, which has none and comes at its own size; `-preset
globe` says the same from the command line, and on a globe the three lines
under it are read rather than set. Nothing that has to be reproduced is founded here —
`headless` and `tune`, which the baseline is taken on, both keep the fixed
default size.

A map larger than the terminal is looked at through a window that moves. The
arrows move it a third of a screen — `hjkl` do the same for a hand already on
the letters — `c` comes back to the settlement, which on a globe is a fraction
of a per cent of the map and the only part of it anybody is watching, and `f`
keeps the window on whoever is being followed, so that picking a figure out of
the crowd goes to them rather than merely naming them in the panel. Looking
around by hand lets go of them again. The line under the graph says which part
of the world is on the screen, because on a globe every view looks alike.

`z` draws the map at twice as much ground to the cell and `Z` back in again,
up to the scale that holds the whole world and no further; the wheel does the
same. Panning answers where else to look and only this answers what shape the
place is: a continent is four hundred tiles across, and where its coast runs
is not something anybody can be told a screenful at a time. A cell that stands
for a block of ground is drawn as the most telling tile in it rather than as
the average of it — what people have built first, then the water where there
is enough of it to be a feature of the block, and otherwise the ground itself.
An average loses exactly what a map is for: a river is a tile wide and a
settlement a dozen across, and the mean of the block holding either of them is
the country around it. The line under the graph says the scale whenever it is
not one tile to the cell.

The arrows go to whichever of the two the page has. Where there is more map
than screen they look around it and the graphs answer to `pgup` and `pgdn`;
on a map the terminal holds whole — every valley run — and on both pages of
graphs, they open a graph out as they always did, and `pgup` and `pgdn` do
the same there. The keys at the foot of the panel say which case it is in, so
this is never something to remember.

Space pauses, `+` and `-` change speed, `.` steps once while paused, `r` lays
streets, `tab` and `shift-tab` (or a click) pick a figure out of the crowd and
open it up beside the map — who it is, what it is good at, what errand it is on,
and everything it weighed before setting out. `esc` drops it, `q` quits.

`m` turns the map to the next reading of the land and `M` back to the last.
The settlement is the one to watch a run on — trees, water, roofs, people —
and the rest each ask the ground one question and answer it over the whole map
at once: how high it stands, how wet it is, what it will grow, what rock is
under it, what its soil is made of, what is standing on it, how much of what
could be growing there is green this morning, where people have actually worn
it, what the water has in it, and who holds what. The row under the map names the reading and both ends of its
shading, so which end is the good ground is never a guess.

Bedrock and texture are the ground's own composition. The rock under a tile
never changes and is laid down in regions, four kinds of it; the soil over it
starts as what that rock weathers to and is then carried about by the water,
which sorts it — sand drops first, silt travels to the flood plain, clay stays
up in the water longest. So the texture view is mostly the shape of the
bedrock view with the valleys rewritten by the river, and ground that reads
nothing like the rock beneath it is ground that was carried there. Texture is
the one scale where neither end is the good end: the best farmland is a loam
in the middle of it, which is why its legend names sand and clay rather than
better and worse.

Holdings is the one that is not a quantity. Its colours stand for different
owners rather than for more and less of one thing, so the row under it is a
key and not a scale — a scale there would say that one farmer is somehow more
than another. Six colours go round however many settlers there are, so it is
for seeing the shape of a holding against the unclaimed ground rather than for
telling one farmer from another by eye.

Soil and green are drawn in the same shading on purpose, so that turning from
one to the other is a comparison rather than a change of subject. Soil is what
the ground could grow and hardly moves in a lifetime; green is what is on it
today, and it moves with every season, every felling and every harvest. Ground
that is bright on the first and dark on the second is good land with nothing
standing on it.

They are readings and not decorations. Every one is a number the settlement
already keeps and already acts on — the soil view is the fertility a settler
weighs when choosing where to break a field, the texture under it is half of
what decides that fertility and all of what decides how fast the hillside
washes away, the green view is what a gatherer finds when it gets there, and
the wear view is what somebody reads before laying a road — so a run where the fields are not on the bright ground, or the
roads are not on the worn ground, is a run worth asking about.

Every page here shows a dozen measurements at once and squeezes each into a
row or a band. `↑` and `↓` step through whatever the page has — the kinds of
work under the map, the measures down the world page, the curve and the two
flows that make it on the vitals page — and open the one stepped onto out
over the page's largest space, scaled to its own high-water mark, saying
where it stands now and what its full height means. A kind of work holding a
twentieth of the population is not drawn at all in a weave shared with six
others, and is a chart of its own when it is stepped onto. The rest of the
page stays where it was, with the row being read marked. Stepping past the
last one, or `esc`, gives the whole page back. `pgup` and `pgdn` do the same
stepping everywhere, and are what to use where the arrows are busy looking
around a map larger than the screen.

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

`headless` and `tune` also take `-preset ancient`, which makes the default
valley out of its own history instead of drawing it. A world starts molten,
with no rigid crust and nothing that forms outlasting the forming of it; the
crust goes rigid and breaks into plates; and then sixteen ages of the earth
run, each drifting the plates, raising ground where two of them close,
dropping it where they part, waking the odd volcano, and weathering all of it
before the next age. What comes out is a map where things have reasons — a
range stands where two plates met, the rock in it is what that meeting made,
the basin beside it is full of the range's own debris — and every tile keeps
the plate it rides and the epoch its rock dates from, so later work can ask
where the ore is rather than paint it on. `-epochs N` is the same term by
hand on any preset. It is not what any settlement is measured on: see
"A world made by what happened to it" in core/world/history.go for what the
join to the tuned constants does and why it is the part most likely to be
wrong.

All three take `-preset globe`, which founds the world on a
globe instead of the valley: a cylinder a thousand tiles round and five
hundred down, joined at the east and west edges, a third of it sea, cold at
the poles and warm at the middle, with the weather read by latitude and by
height: the air cools six and a half degrees a kilometre, so the mountains
carry a tree line and a snow cap of their own and the peaks are bare for the
same reason the poles are. On the valley the highest ground is two degrees
colder than the river and no more. On `headless` and `tune`, `-width`,
`-height` and `-wrap` are the same terms by hand and override the preset
where given; a globe must be a whole number of chunks round. `watch` takes a
globe at its own size and moves a window over it instead — see above. The map is kept
in chunks of sixty-four tiles, and ground with nobody on it, nothing built
on it and nobody across it lately sleeps: the day's passes skip it and it is
caught up in one go when it wakes, which on the default map never happens,
so nothing measured there moves. `headless -timing` says what each report
interval spent on each phase of the day, how much of the ground was awake
and why, how many islands the day's acting was cut into and how much of the
population the biggest held, and how many plans found no way; `-cpuprofile`
writes a profile. The islands are the pieces of the population that cannot
touch each other's ground in a day - parties living further apart than
anybody walks, looks or routes - and each acts on a goroutine of its own,
with what it does to the settlement as a whole put together afterwards in
one order; a world whose people all live in one place is one island and is
acted on as it always was. See `core/world/island.go`.
What the globe batch prints is checked in at
[docs/baseline-globe.md](docs/baseline-globe.md), taken on eight seeds, and
it is a different settlement from the valley's in every number.

```bash
go test ./...
```

One check does not run from there. `owl/` is its own module, so the root suite
walks past it and the test that keeps the OWL rendering of the ontology honest
never fires. A hook covers it, once per clone:

```bash
git config core.hooksPath .githooks
```

A commit touching anything that ontology document is built from then has to
leave it current. See [owl/README.md](owl/README.md).

## What a run leaves behind

Every run of all three commands is kept, and says on its last line where it
went:

```
runs/<branch>/20260908-234526-tune.md
```

The file is the terms the run was founded on — the command line, the branch,
the commit and whether it was clean — then its scores, then everything it
printed, verbatim. Beside the reports is an `index.jsonl` with one line per
run: the same scores as JSON, so a folder of a hundred runs can be read
without opening any of them.

```bash
cat runs/*/index.jsonl | jq -r '[.at, .scores["mean-pop"], .scores["gates-all"]] | @tsv'
```

The folder is named for the branch because several branches are usually being
run at once and their numbers mean nothing mixed together. `runs/` is ignored
and is meant to be: the numbers belong to the tree they were taken on, and
master's are checked in at [docs/baseline.md](docs/baseline.md) instead.

`tune -quiet` is about the terminal and not about the record — the table and
the tally still go into the file. Set `LREAT_RUNS` to gather the branch
folders of several worktrees in one place, or `LREAT_REPORTS=off` to keep
nothing at all. A run killed part-way through leaves no report; it is written
when the command ends of its own accord.

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
| `report` | what each run left behind, kept under `runs/` a branch at a time |
| `cmd/` | `watch`, `headless`, `tune` |
| `owl/` | `core/ontology` rendered as an OWL 2 document — its own module, so the simulation gains no dependency from it |

Exactly one goroutine drives a world. Deciding is spread over several — it only
reads — while everything that changes the world runs one at a time.

# Baseline

What `tune` prints on master, checked in so that nobody has to run it to find
out. The run below is the default batch:

```bash
go run ./cmd/tune -seeds 24 -ticks 21600
```

Nothing in it is sampled from the wall clock, so the same commit gives the same
numbers however the goroutines interleave. That is what makes a checked-in
baseline worth anything: the file is the run, and re-running it only confirms
it.

Read it before a change rather than measuring master again. Two hours of work
begins with a few minutes of a batch that already has an answer here, and the
answer does not change while the work is going on. Take the comparison run once
the change is settled instead, and when it lands, replace what is below with the
new numbers in the same commit - the file describes the tree it is committed in,
so a change to behaviour that leaves it alone has left it wrong.

The headline is the last line: how many of the 24 settlements were still at
least their founding size at the end, how many died out, and where the mean
population landed. The `gates` line in front of it is the share of fertile
agents fed, safe and held - `all` is the share that were all three at once,
which is the gate a birth has to pass, and the one worth watching.

## Why the batch is 21,600 ticks and not 6000

It is the same sixty years it always was. A tick used to be a hundredth of a
year and is now a day, so the number in the command had to change for the run
to go on meaning the same thing; see "The calendar" in docs/action-space.md.
Anybody holding these numbers against a batch measured before that should
compare years and not ticks - and should know that sixty years is not quite the
same stretch of a life either way, because a life was 52 years then and is 65
now. A settlement is read a little earlier in its generations here than it used
to be.

## master

```
seed  pop  died births houses fields |  phys  safe belng  estm  actl | order
   1   64    97     27     10     44 |  0.70  0.45  0.50  0.70  0.06 |  1.00
   2  196   159     29     86     47 |  0.55  0.50  0.72  0.65  0.08 |  1.00
   3  120   124     25     36     68 |  0.58  0.47  0.66  0.67  0.05 |  1.00
   4   66   141     36     25     73 |  0.49  0.52  0.41  0.62  0.02 |  1.00
   5  281   369     22     95    138 |  0.62  0.51  0.53  0.67  0.09 |  1.00
   6  318   262     48    142    211 |  0.64  0.55  0.56  0.58  0.01 |  1.00
   7   41    47     38     33     10 |  0.55  0.57  0.86  0.80  0.07 |  0.79
   8  140   173     23     42     67 |  0.66  0.51  0.39  0.72  0.02 |  1.00
   9   69   154     25      7     34 |  0.58  0.45  0.54  0.67  0.13 |  1.00
  10  424   520     55    115    191 |  0.56  0.47  0.43  0.47  0.04 |  1.00
  11   30   153     75     23     28 |  0.57  0.62  0.54  0.67  0.04 |  0.99
  12  333   190     11     38    146 |  0.60  0.43  0.68  0.83  0.09 |  1.00
  13    0    33     13      0      0 |  0.00  0.00  0.00  0.00  0.00 |  0.00
  14  257   265     27     97    164 |  0.60  0.50  0.52  0.69  0.04 |  1.00
  15  145   283     25     73     75 |  0.58  0.55  0.51  0.71  0.08 |  1.00
  16  202   222     26     88    116 |  0.60  0.50  0.48  0.59  0.09 |  1.00
  17  147   420     49     67    109 |  0.57  0.52  0.56  0.60  0.10 |  1.00
  18   92   340     41     38     56 |  0.58  0.56  0.53  0.67  0.09 |  1.00
  19  221   218     27     71     66 |  0.61  0.45  0.75  0.77  0.02 |  1.00
  20  166   115     39     52    107 |  0.63  0.50  0.68  0.59  0.10 |  1.00
  21  543   389     33     57    258 |  0.64  0.53  0.61  0.63  0.18 |  1.00
  22  111   179     22     29     86 |  0.71  0.51  0.47  0.57  0.07 |  1.00
  23  338   357     33     94    237 |  0.65  0.49  0.47  0.54  0.04 |  1.00
  24   82    79     23     17     91 |  0.53  0.49  0.77  0.78  0.07 |  1.00

dwell/rest                         2847009  26.4%
take/berries@wood                  2712612  25.1%
consume/provision                  1815720  16.8%
dwell/guard@market                 1056765   9.8%
pass/practice>pupil                 973886   9.0%
take/fish@water                     346450   3.2%
dwell/meet@tavern>neighbour         297465   2.8%
take/grain@field                    146852   1.4%
take/timber@wood                    132517   1.2%
transfer/provision>needy            117447   1.1%
pass/practice>self                  111375   1.0%
exchange/material>coin@market        49028   0.5%
raise/timber>dwelling@open           42665   0.4%
tend/plant@open                      39254   0.4%
exchange/coin>provision@market       38026   0.4%
transfer/material>requester          20963   0.2%
transfer/provision<holder            11883   0.1%
make/timber>tool@bench                5635   0.1%
tend/clear@open                       5225   0.0%
strike/person>wrongdoer               4740   0.0%
take/game@wood                        3284   0.0%
tend/water@field                      2886   0.0%
take/stone@outcrop                    2364   0.0%
raise/timber>road@ground              2328   0.0%
make/provision+timber>meal@hearth      442   0.0%
dwell/look                             267   0.0%
raise/timber>tavern@open               110   0.0%
move@dwelling                           84   0.0%
raise/timber+stone>granary@open         74   0.0%
make/stone+timber>tool@forge            39   0.0%
raise/timber+stone>market@open          18   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.56 safe 0.48 held 0.83 all 0.269 food 4.64 hungry-with-food 0.26 | lasted 23/24 extinct 1 mean 182.8 median 147 | phys 0.60 safe 0.51 belng 0.57 estm 0.66
```

## What the last change did

The ground got a make-up. There is rock under every tile now - granite,
limestone, sandstone or shale, laid down in regions by two coarse lattices
crossed at their own middles - and the soil over it is a mixture of sand, silt
and clay weathered out of that rock. Two shares are kept and the third is the
remainder, so a soil cannot disagree with itself.

The part worth having is that the water sorts it. Erode carried one load and
now carries three, each settling at its own rate: sand goes down at the first
slackening, silt travels to where the river spills over its bank, clay stays
up in the water longest. The three average within a hundredth of the single
figure they replace, so a map silts up at about the rate the model was
measured at, and what is new is where each grain of it lands. Held against the
bedrock, the texture of a map is the geology with the valleys rewritten by the
river.

Then two things read it. Fertility is multiplied by how near a soil is to a
loam - both ends of that scale are poor, and the best ground is in the middle
of it - and what an age of weather strips off a tile is multiplied by how
sandy it is, so a settlement that ploughs its sandy slopes loses them faster
than one that ploughs its clay. See bedrock.go.

**It is a draw on the settlements, on two batches.** That is the intended
result: this is structure and not difficulty.

```
offset  0, master: gates: fed 0.56 safe 0.52 held 0.83 all 0.291 food 4.76 | lasted 22/24 extinct 0 mean 186.4 median 159
offset  0, soils:  gates: fed 0.56 safe 0.48 held 0.83 all 0.269 food 4.64 | lasted 23/24 extinct 1 mean 182.8 median 147
offset 24, master: gates: fed 0.55 safe 0.53 held 0.86 all 0.295 food 4.31 | lasted 23/24 extinct 0 mean 161.8 median 100
offset 24, soils:  gates: fed 0.56 safe 0.52 held 0.84 all 0.290 food 4.40 | lasted 23/24 extinct 0 mean 171.2 median 140
```

The default batch alone would have been worth a second look: `safe` down 0.04,
`all` - the gate a birth has to pass - down 0.022, and one settlement dying
where none had. None of that survives the independent batch, where the same
two readings are down 0.01 and 0.005 and nothing goes extinct. Pooled over the
forty-eight seeds `all` is down 0.014 and `safe` 0.025, both an order inside
what chance moves them by, and the population disagrees in direction between
the two batches in the way it always does. `fed` is the same to a hundredth on
both.

What did move, and is not in these numbers, is how much the seeds differ from
each other about their ground. A settlement founded over sandstone has a
poorer valley than one founded over shale, and that is a new thing for a seed
to decide. Anybody reading a population figure off a small batch after this
should expect it to be looser than it was.

## Earlier changes

What the changes before this one did, each measured against the master of its
own day. They are kept for the method rather than for the numbers: none of
them is a comparison with the run above.

### What the lapse rate did

The weather learned about height. It was read by latitude alone, so the top of
a mountain was exactly as warm as the valley it stood over - `Height` was the
one thing a tile carried that nothing in the weather ever asked for. The air
now cools by `world.Lapse`, six and a half degrees a kilometre, and the cold a
body feels, the growing weather the ground gets, what a pack keeps by, and
where seed will take are all read where they are rather than off the row. Two
rules became one on the way: ground whose year never warms past `Frost` is bare
outcrop, which used to be said of the poles alone and now covers a peak without
naming one, and the woods stop below it.

**On this map it was behaviourally neutral, and that was the expectation
going in.** The default valley is sixty metres of relief with two hundred and sixty
of high country on it, so the highest ground on it is two degrees colder than
the river and no ground on it is ever frozen: the ploughable share of the map
does not move, and the tree line has nothing to bite on. What the run measures
is two degrees spread over the shoulders of a valley.

Nothing crossed a threshold on either batch:

```
offset  0, master: gates: fed 0.55 safe 0.51 held 0.83 all 0.278 food 4.34 | lasted 22/24 extinct 0 mean 143.3 median 132
offset  0, lapse:  gates: fed 0.56 safe 0.52 held 0.83 all 0.291 food 4.76 | lasted 22/24 extinct 0 mean 186.4 median 159
offset 24, master: gates: fed 0.58 safe 0.51 held 0.85 all 0.297 food 4.52 | lasted 23/24 extinct 0 mean 163.9 median 151
offset 24, lapse:  gates: fed 0.55 safe 0.53 held 0.86 all 0.295 food 4.31 | lasted 23/24 extinct 0 mean 161.8 median 100
```

The default seeds gained a third of their mean population and a fifth of their
median, which read like a win and is not one. The independent batch does not
confirm it: on seeds the baseline never draws from the mean is flat - 163.9 to
161.8 - and the median falls by a third, the opposite direction. `fed` went up
0.01 on one batch and down 0.03 on the other, `all` up 0.013 and down 0.002,
and `lasted` and `extinct` are identical on both. Pooled over the forty-eight
seeds, `fed` is down 0.01 and `all` up 0.005, both an order inside what chance
moves them by. So the honest reading is a draw, and the population figures on
either batch are the loosest number in the file doing what it always does.

What the change buys is not on this map at all. On a globe the mountains are
ten times as high, and there the same constant is the difference between a
world with a tree line and one without: at 1024 by 512 the frozen share of the
dry land goes from 33% to 39%, and the 6% is snow on the peaks rather than more
ice at the poles - the highest ground there is 2344 metres against a frost line
of 2275 at the temperate latitudes. That is a reading taken at generation, on
seed 1; what a globe settlement then does with it is measured in
[baseline-globe.md](baseline-globe.md).

### What fencing moved

Hedges round the large holdings; see "Fences" in docs/action-space.md. This is
an account of an earlier change, kept because its second batch is the clearest
worked example in the file of a population move that turned out to be a draw.
It was measured against the master of its own day, not against the run above.

Against the batch it replaced - fed 0.56, safe 0.53, all 0.291, lasted 24 of
24, mean 238.2, median 198 - the gates did not move at all:
`fed` held, `safe` and `all` came up by 0.01 and 0.007, and one settlement of
twenty-four fell below its founding size. What moved was the population on these
seeds, mean by a sixth and median by a third, and an independent batch does not
confirm it:

```
offset 24, before: gates: fed 0.57 safe 0.52 held 0.83 all 0.292 | lasted 24/24 extinct 0 mean 229.5 median 202
offset 24, fenced: gates: fed 0.58 safe 0.50 held 0.84 all 0.295 | lasted 24/24 extinct 0 mean 218.1 median 211
```

On seeds the baseline never draws from, the gates are the same to within a
hundredth, nothing was lost, the mean is down 5% and the median is up. So the
fall on the default seeds is a draw and not a cost: two batches agree that
hedging the big holdings is behaviourally neutral, and what it buys is that
people stop walking through the corn.

### What the mountains cost

The land grew mountains. `raise` used to be one texture at one amplitude scaled
to sixty metres, and every seed came out the same gentle bowl - nine tiles in
ten under a tenth of a slope, with the highest ground only the largest of the
same lumps. It now raises a lowland to exactly that same sixty metres and
stands two hundred and sixty metres of ridged high country on a fifth of it,
with the rivers cutting their own valleys into what is left. The high country
is scaled to the width of the map and the lowland is not, so a wider map is a
bigger country at the same ruggedness and its valley floor is unchanged. See
relief.go.

What that costs is ground. The mountains are not farmland, and the gentle open
ground a settlement can plough falls from 72% of the map to 55% over the first
eight seeds.

What that costs in people is about a sixteenth. Pooled over all seventy-two
seeds the mean population goes 230 to 215 and the median 198 to 176, and all
three batches agree in direction - down a thirtieth, a fifteenth and an
eleventh.

Nothing else moved, and two things moved the wrong way for a change that costs
ground: pooled, `fed` went 0.563 to 0.570 and `all` - the gate a birth has to
pass - 0.294 to 0.298, both up and both far inside their thresholds. `lasted`
went 72 of 72 to 71 and nothing went extinct on any batch. The settlements are
not doing worse on the land they have; there is less of it and slightly fewer
of them on it.

**These three batches were taken on c8513bc, before the hedges landed**, and
are compared against that commit's own numbers. The hedges went in while they
were running and changed how a journey goes, so the figures above are a
measurement of the mountains and not of the tree they are committed in. They
were not taken again because master took four behavioural changes during the
day this branch was open, each roughly as often as a three-batch run takes to
finish, and a number chased under those conditions never lands. What is worth
holding is the direction, which held across three separate attempts on three
different masters. Whoever next has cause to refresh this file should simply
refresh it; nothing here needs preserving.

This was taken with the map at its default eighty by thirty-six, which is the
width the high country is quoted at. A batch taken on a bigger map is a
different measurement and not comparable to this one: the mountains grow with
the map, the valley does not, and the ploughable share of the tiles rises from
55% to about 63% by two hundred and forty wide.

## How much of that is chance

The same code on two further batches of twenty-four seeds, which `-offset`
draws from a stretch of the seed space the baseline never touched:

```bash
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 24 -quiet
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 48 -quiet
```

```
offset  0: gates: fed 0.55 safe 0.51 held 0.83 all 0.278 food 4.34 hungry-with-food 0.28 | lasted 22/24 extinct 0 mean 143.3 median 132 | phys 0.60 safe 0.52 belng 0.53 estm 0.67
offset 24: gates: fed 0.58 safe 0.51 held 0.85 all 0.297 food 4.52 hungry-with-food 0.28 | lasted 23/24 extinct 0 mean 163.9 median 151 | phys 0.62 safe 0.51 belng 0.56 estm 0.59
offset 48: gates: fed 0.57 safe 0.50 held 0.84 all 0.275 food 5.67 hungry-with-food 0.26 | lasted 21/24 extinct 1 mean 170.1 median 161 | phys 0.58 safe 0.50 belng 0.59 estm 0.64
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 143, 164, 170, the median 132, 151, 161. `lasted` was 22, 23
and 21 of 24, and one of the three lost a settlement outright. The mean is the
loosest reading here and always has been: a settlement that runs away is worth
as much as the twenty that did not, and the handful sitting over four hundred
are most of the distance between the draws. The median is looser on this tree
than it used to be, and the mountains are why: how much of a map's lowland a
range happens to cover is now one of the things a seed decides, so the seeds
have more to differ about and the middle of them moves further.

What holds still is the per-agent side. `fed` moved by 0.03 across the three,
the mean needs by 0.02 to 0.08, and `all` sat between 0.275 and 0.297. So:

| reading | moved by chance | worth believing at |
|---|---|---|
| fed, and the four mean needs | 0.02-0.05 | 0.05 on two batches that agree |
| gates all | 0.022 | 0.05 |
| lasted, extinct | 2 of 24, 1 of 24 | 5 of 24 |
| mean population | 16% of itself | all three batches, pooled |
| median population | 18% of itself | all three batches, pooled |

So a change that only moves the population numbers has not been shown to do
anything, and one batch cannot show it either way. Take all three and pool them
before believing any of it, and say in the commit what the pooled figure was.

The change that raised the mountains is the worked example, and it is worth
keeping because of how far the answer wandered. Measured against the master of
the day, its first batch alone said a third fewer people; the three batches
together said a seventh, with one of them saying the population had gone up.
Measured again after merging the master it first landed on, the same three said
a fourteenth. Measured a third time, on a master where the wear on a road had
changed underneath it, all three agreed in direction for the first time and
said a sixteenth. None of those batches was wrong and none was mismeasured. The
first was simply not an answer, and neither is any single batch taken after it.

The other half of that lesson is about the tree rather than the seeds. Two of
those three measurements were stale before they could be committed, because
master took four behavioural changes in a day and each moved the numbers the
comparison was against. When that is happening, refreshing this file is a race
and not a task: take the batches, say which commit they were taken on, and let
whoever needs them next take them again.

## What the ceiling is doing

`system.MaxPopulation` is 5000, and none of the seventy-two settlements above
comes near it: the largest is 508. That is the point of where it sits. It was
400 until the batch above, which two of these seeds stood exactly on, and a
settlement held at a ceiling reads the same as one that found its level - so
the headline number was being decided in a constant rather than out on the
land.

Lifting it moved nothing but the tail. Every per-agent reading came out the
same to the digit - `fed` 0.57, `all` 0.293, 0.277, 0.309 against 0.293, 0.277,
0.310 - and every median was unchanged, because the two seeds that were pinned
were the only ones affected: seed 1 went 400 to 502, seed 21 400 to 508, and
the mean moved with them. Both are worse fed and lonelier at their new size
than they were at 400, which is what a settlement finding its own ceiling looks
like.

`world.Crowded` counts everyone turned away by the cap on a tick, and it is
what to read if the question comes up again. While it stays at nothing, the
land is doing the binding.

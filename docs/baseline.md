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
   1  113   120     32     37     13 |  0.54  0.49  0.50  0.61  0.06 |  1.00
   2  784   716     30    111    283 |  0.61  0.49  0.55  0.49  0.08 |  1.00
   3  317   205     14     61    333 |  0.61  0.45  0.59  0.71  0.10 |  1.00
   4  134   269     58     66    107 |  0.64  0.57  0.50  0.58  0.09 |  0.99
   5  235   337     33     51    179 |  0.58  0.49  0.53  0.58  0.12 |  1.00
   6   81   290     67     41     90 |  0.60  0.58  0.60  0.57  0.04 |  0.99
   7   27    89     27      9     25 |  0.73  0.50  0.72  0.78  0.09 |  0.86
   8  136   311     16     64     87 |  0.63  0.56  0.43  0.72  0.03 |  1.00
   9  244   266     30     83    233 |  0.60  0.49  0.60  0.75  0.06 |  0.99
  10  209   225     26     56     75 |  0.65  0.45  0.51  0.62  0.13 |  1.00
  11   43    73     11      8     28 |  0.69  0.45  0.45  0.92  0.10 |  1.00
  12  130   250     32     53     46 |  0.60  0.49  0.60  0.77  0.05 |  1.00
  13  213   266     26     96    149 |  0.61  0.53  0.47  0.72  0.08 |  1.00
  14  472   424     50    114    398 |  0.54  0.50  0.56  0.67  0.04 |  1.00
  15  106   421     41     55     98 |  0.66  0.54  0.47  0.48  0.03 |  1.00
  16  282   275     24     51    255 |  0.65  0.49  0.60  0.77  0.14 |  1.00
  17   50    98     17     23     13 |  0.57  0.47  0.60  0.79  0.05 |  1.00
  18  132   349     39     29     87 |  0.55  0.53  0.68  0.66  0.20 |  1.00
  19  106   226     42     54     97 |  0.59  0.49  0.50  0.57  0.04 |  0.99
  20  762   261     34    108    488 |  0.69  0.49  0.54  0.57  0.04 |  1.00
  21  228   430     42    110    131 |  0.56  0.53  0.58  0.54  0.04 |  1.00
  22  236   243     29     87    134 |  0.57  0.49  0.58  0.84  0.08 |  1.00
  23  143   213     43     77    120 |  0.58  0.52  0.58  0.64  0.04 |  1.00
  24  344   243     41    148     98 |  0.54  0.51  0.46  0.58  0.02 |  1.00

dwell/rest                         3608059  27.4%
take/berries@wood                  3162251  24.0%
consume/provision                  2314358  17.6%
pass/practice>pupil                1247976   9.5%
dwell/guard@market                 1002865   7.6%
take/fish@water                     518258   3.9%
dwell/meet@tavern>neighbour         319265   2.4%
take/grain@field                    181510   1.4%
take/timber@wood                    178699   1.4%
transfer/provision>needy            172511   1.3%
pass/practice>self                  137246   1.0%
tend/plant@open                      69356   0.5%
exchange/material>coin@market        64672   0.5%
raise/timber>dwelling@open           59010   0.4%
exchange/coin>provision@market       52387   0.4%
transfer/material>requester          28316   0.2%
transfer/provision<holder            11093   0.1%
make/timber>tool@bench                8992   0.1%
take/game@wood                        8306   0.1%
tend/clear@open                       7226   0.1%
strike/person>wrongdoer               4563   0.0%
tend/water@field                      4405   0.0%
raise/timber>road@ground              3675   0.0%
take/stone@outcrop                    2178   0.0%
make/provision+timber>meal@hearth      649   0.0%
dwell/look                             443   0.0%
raise/timber>tavern@open               134   0.0%
make/stone+timber>tool@forge           131   0.0%
move@dwelling                           93   0.0%
raise/timber+stone>granary@open         59   0.0%
raise/timber+stone>market@open          31   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.57 safe 0.52 held 0.84 all 0.292 food 4.68 hungry-with-food 0.27 | lasted 24/24 extinct 0 mean 230.3 median 209 | phys 0.61 safe 0.50 belng 0.55 estm 0.66
```

### What fencing moved

The run above is the tree with hedges round the large holdings; see "Fences" in
docs/action-space.md. Against the batch it replaced - fed 0.56, safe 0.53, all
0.291, lasted 24 of 24, mean 238.2, median 198 - the gates did not move at all:
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

## What the last change did

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
offset  0: gates: fed 0.57 safe 0.52 held 0.84 all 0.292 food 4.68 hungry-with-food 0.27 | lasted 24/24 extinct 0 mean 230.3 median 209 | phys 0.61 safe 0.50 belng 0.55 estm 0.66
offset 24: gates: fed 0.58 safe 0.56 held 0.85 all 0.323 food 4.57 hungry-with-food 0.27 | lasted 24/24 extinct 0 mean 213.8 median 157 | phys 0.59 safe 0.52 belng 0.56 estm 0.62
offset 48: gates: fed 0.56 safe 0.50 held 0.84 all 0.279 food 4.27 hungry-with-food 0.26 | lasted 23/24 extinct 0 mean 201.4 median 163 | phys 0.58 safe 0.50 belng 0.59 estm 0.67
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 230, 214, 201, the median 209, 157, 163. `lasted` was 24, 24
and 23 of 24, and none of them lost a settlement outright. The mean is the
loosest reading here and always has been: a settlement that runs away is worth
as much as the twenty that did not, and the handful sitting over four hundred
are most of the distance between the draws. The median is looser on this tree
than it used to be, and the mountains are why: how much of a map's lowland a
range happens to cover is now one of the things a seed decides, so the seeds
have more to differ about and the middle of them moves further.

What holds still is the per-agent side. `fed` moved by 0.02 across the three,
the mean needs by 0.02 to 0.05, and `all` sat between 0.279 and 0.323. So:

| reading | moved by chance | worth believing at |
|---|---|---|
| fed, and the four mean needs | 0.02-0.05 | 0.05 on two batches that agree |
| gates all | 0.044 | 0.05 |
| lasted, extinct | 1 of 24, 0 of 24 | 5 of 24 |
| mean population | 13% of itself | all three batches, pooled |
| median population | 29% of itself | all three batches, pooled |

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

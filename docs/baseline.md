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
   1   27   118     43     18      2 |  0.55  0.55  0.44  0.74  0.06 |  0.99
   2  385   409     21    120     89 |  0.59  0.47  0.53  0.52  0.03 |  1.00
   3  240   273     18     81    307 |  0.60  0.47  0.58  0.68  0.07 |  1.00
   4   15    64     15     11     10 |  0.32  0.41  0.65  0.77  0.03 |  0.50
   5  431   481     34     49    284 |  0.58  0.49  0.53  0.57  0.13 |  1.00
   6   51   263     53     36     49 |  0.56  0.58  0.56  0.56  0.03 |  1.00
   7   19    61     10      5     21 |  0.60  0.51  0.81  0.82  0.12 |  1.00
   8  283   287     25    104    118 |  0.58  0.53  0.34  0.65  0.04 |  1.00
   9  242   260     27     96    123 |  0.59  0.51  0.45  0.67  0.05 |  1.00
  10  218   186     17     63     75 |  0.64  0.46  0.55  0.74  0.10 |  1.00
  11   68   172     27     26     38 |  0.57  0.53  0.47  0.79  0.07 |  1.00
  12  134    86     20     50     36 |  0.59  0.48  0.60  0.70  0.03 |  1.00
  13  224   206     18     62    190 |  0.60  0.48  0.51  0.78  0.06 |  1.00
  14  151   331     77     77    180 |  0.57  0.61  0.45  0.68  0.07 |  1.00
  15  153   415     51     77     71 |  0.44  0.53  0.42  0.47  0.04 |  1.00
  16  148   224     17     42    157 |  0.58  0.46  0.50  0.69  0.14 |  1.00
  17  273   226     36     57    113 |  0.60  0.46  0.59  0.69  0.13 |  1.00
  18   86   234     45     38     65 |  0.58  0.54  0.45  0.62  0.13 |  1.00
  19   47   211     33     22     55 |  0.58  0.52  0.48  0.53  0.04 |  0.98
  20  370   397     29    129    254 |  0.61  0.48  0.40  0.45  0.03 |  1.00
  21  261   584     50    120    192 |  0.52  0.51  0.52  0.54  0.01 |  1.00
  22   63    66     17      9     57 |  0.61  0.45  0.67  0.94  0.08 |  1.00
  23  124   151     33     39    120 |  0.63  0.52  0.64  0.69  0.09 |  1.00
  24   89   146     49     57     37 |  0.60  0.51  0.51  0.69  0.03 |  0.71

dwell/rest                         3387979  28.9%
take/berries@wood                  2948954  25.1%
consume/provision                  1867728  15.9%
dwell/guard@market                 1120378   9.6%
pass/practice>pupil                 959974   8.2%
take/fish@water                     428540   3.7%
dwell/meet@tavern>neighbour         271156   2.3%
take/timber@wood                    158115   1.3%
take/grain@field                    115310   1.0%
pass/practice>self                  114002   1.0%
transfer/provision>needy            100314   0.9%
raise/timber>dwelling@open           59523   0.5%
exchange/material>coin@market        44272   0.4%
exchange/coin>provision@market       43288   0.4%
tend/plant@open                      42445   0.4%
transfer/material>requester          18893   0.2%
transfer/provision<holder            14251   0.1%
make/timber>tool@bench                8428   0.1%
tend/clear@open                       5870   0.1%
take/game@wood                        5863   0.0%
strike/person>wrongdoer               5265   0.0%
tend/water@field                      3666   0.0%
raise/timber>road@ground              2096   0.0%
take/stone@outcrop                    2038   0.0%
make/provision+timber>meal@hearth      670   0.0%
dwell/look                             339   0.0%
raise/timber>tavern@open               145   0.0%
move@dwelling                           79   0.0%
make/stone+timber>tool@forge            62   0.0%
raise/timber+stone>granary@open         58   0.0%
raise/timber+stone>market@open          31   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.54 safe 0.49 held 0.83 all 0.262 food 4.28 hungry-with-food 0.27 | lasted 22/24 extinct 0 mean 170.9 median 151 | phys 0.57 safe 0.50 belng 0.53 estm 0.67
```

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

What that costs in people is about a fourteenth. Pooled over all seventy-two
seeds the mean population goes 212 to 197 and the median 185 to 172. No single
batch says that: batch for batch against master the three went down a quarter,
down a twelfth, and up a sixth, which is a wider spread than the effect they
average to.

Little else moved. Pooled, `fed` went 0.570 to 0.560 and `all` - the gate a
birth has to pass - 0.301 to 0.277, both inside their thresholds; `lasted` went
71 of 72 to 69, and nothing went extinct on any batch. The settlements are not
doing much worse on the land they have. There is less of it, and a few fewer of
them.

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
offset  0: gates: fed 0.54 safe 0.49 held 0.83 all 0.262 food 4.28 hungry-with-food 0.27 | lasted 22/24 extinct 0 mean 170.9 median 151 | phys 0.57 safe 0.50 belng 0.53 estm 0.67
offset 24: gates: fed 0.56 safe 0.53 held 0.83 all 0.293 food 4.14 hungry-with-food 0.28 | lasted 23/24 extinct 0 mean 195.3 median 184 | phys 0.58 safe 0.51 belng 0.57 estm 0.63
offset 48: gates: fed 0.58 safe 0.49 held 0.84 all 0.277 food 5.18 hungry-with-food 0.26 | lasted 24/24 extinct 0 mean 224.5 median 181 | phys 0.61 safe 0.50 belng 0.58 estm 0.65
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 171, 195, 225, the median 151, 184, 181. `lasted` was 22, 23
and 24 of 24, and none of the three lost a settlement outright. The mean is the
loosest reading here and always has been: a settlement that runs away is worth
as much as the twenty that did not, and the handful above sitting over four
hundred are most of the distance between the draws. The spread is wider on this
tree than it was before the mountains, and the mountains are why: how much of a
map's lowland a range happens to cover is now one of the things a seed decides,
so the seeds have more to differ about.

What holds still is the per-agent side. `fed` moved by 0.04 across the three,
the mean needs by 0.02 to 0.05, and `all` sat between 0.262 and 0.293. So:

| reading | moved by chance | worth believing at |
|---|---|---|
| fed, and the four mean needs | 0.02-0.05 | 0.05 on two batches that agree |
| gates all | 0.031 | 0.05 |
| lasted, extinct | 2 of 24, 0 of 24 | 5 of 24 |
| mean population | 27% of itself | all three batches, pooled |
| median population | 19% of itself | all three batches, pooled |

So a change that only moves the population numbers has not been shown to do
anything, and one batch cannot show it either way. Take all three and pool them
before believing any of it, and say in the commit what the pooled figure was.
The change that raised the mountains is the worked example: measured against
the master of the day, its first batch alone said a third fewer people and its
third batch said a tenth more, and the three together said a seventh. Pooled
again after merging the master it actually landed on, it said a fourteenth.
None of those batches was wrong; the first one was simply not an answer.

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

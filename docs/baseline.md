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
   1   55    65     10     17     22 |  0.54  0.48  0.70  0.72  0.03 |  0.96
   2  374   624     34    112    171 |  0.60  0.48  0.55  0.55  0.06 |  1.00
   3  238   254     25     66    233 |  0.64  0.47  0.62  0.68  0.07 |  1.00
   4  107   207     69     48    104 |  0.57  0.55  0.60  0.65  0.06 |  0.97
   5  261   301     24     60    170 |  0.54  0.46  0.46  0.51  0.14 |  1.00
   6   93   353     55     53     98 |  0.52  0.54  0.56  0.54  0.04 |  0.78
   7   21    93     22     12     33 |  0.50  0.53  0.71  0.89  0.06 |  0.99
   8   69   149      9     37     18 |  0.68  0.52  0.23  0.66  0.02 |  1.00
   9  213   293     28     98    114 |  0.58  0.53  0.56  0.71  0.05 |  1.00
  10  224   188     29     82     75 |  0.58  0.48  0.51  0.71  0.08 |  1.00
  11   35    77     15      7     25 |  0.64  0.44  0.34  0.92  0.04 |  1.00
  12  124   356     31     44    108 |  0.56  0.53  0.55  0.69  0.04 |  1.00
  13  334   313     29    128    249 |  0.61  0.49  0.49  0.71  0.05 |  1.00
  14  321   326     52     82    242 |  0.56  0.55  0.59  0.73  0.10 |  1.00
  15  182   543     55     88    156 |  0.58  0.55  0.46  0.48  0.03 |  1.00
  16  161   245     30     61    137 |  0.63  0.51  0.53  0.73  0.07 |  1.00
  17   65   120     33     31     36 |  0.60  0.53  0.49  0.64  0.20 |  0.99
  18   33   171     56     23     20 |  0.50  0.63  0.44  0.58  0.08 |  0.99
  19   91   192     26     39    105 |  0.58  0.51  0.60  0.73  0.03 |  1.00
  20  128   249     19     29    117 |  0.66  0.48  0.46  0.59  0.02 |  1.00
  21  217   579     49    104    151 |  0.52  0.53  0.51  0.50  0.04 |  0.99
  22  153   137     12     55    146 |  0.66  0.50  0.56  0.84  0.07 |  1.00
  23   80   180     32     41     72 |  0.62  0.53  0.65  0.78  0.09 |  1.00
  24  137   198     52     73     70 |  0.48  0.52  0.46  0.50  0.01 |  0.96

dwell/rest                         2980463  27.6%
take/berries@wood                  2588519  24.0%
consume/provision                  1801266  16.7%
dwell/guard@market                 1018976   9.4%
pass/practice>pupil                 944441   8.7%
take/fish@water                     450748   4.2%
dwell/meet@tavern>neighbour         259378   2.4%
take/timber@wood                    153482   1.4%
take/grain@field                    129310   1.2%
transfer/provision>needy            116003   1.1%
pass/practice>self                  103240   1.0%
exchange/material>coin@market        51221   0.5%
raise/timber>dwelling@open           48652   0.5%
exchange/coin>provision@market       44333   0.4%
tend/plant@open                      37970   0.4%
transfer/material>requester          16804   0.2%
transfer/provision<holder            15447   0.1%
make/timber>tool@bench                9334   0.1%
tend/clear@open                       6151   0.1%
take/game@wood                        6102   0.1%
strike/person>wrongdoer               5159   0.0%
tend/water@field                      3597   0.0%
raise/timber>road@ground              3546   0.0%
take/stone@outcrop                    1808   0.0%
make/provision+timber>meal@hearth      392   0.0%
dwell/look                             345   0.0%
raise/timber>tavern@open               140   0.0%
move@dwelling                           81   0.0%
raise/timber+stone>granary@open         62   0.0%
make/stone+timber>tool@forge            39   0.0%
raise/timber+stone>market@open          27   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.54 safe 0.50 held 0.82 all 0.268 food 3.94 hungry-with-food 0.27 | lasted 24/24 extinct 0 mean 154.8 median 137 | phys 0.58 safe 0.51 belng 0.53 estm 0.67
```

## What the last change did

The land grew mountains. `raise` used to be one texture at one amplitude
scaled to sixty metres, and every seed came out the same gentle bowl; it now
raises a lowland to exactly that same sixty metres and stands two hundred and
sixty metres of ridged high country on a fifth of it, with the rivers cutting
their own valleys into what is left. See relief.go.

What that costs is ground. The mountains are not farmland, and the gentle open
ground a settlement can plough fell from 72% of the map to 55%, over the first
eight seeds.

What that costs in people is less than one batch made it look, and the three
batches below are the reason to distrust one. Against master seed for seed the
mean population went 223.6 to 154.8 on the baseline draw, 230.2 to 190.2 on the
next, and 201.3 to 218.7 on the third - down a third, down a sixth, and up a
tenth. Pooled over all seventy-two seeds it is 218 to 188, about a seventh
fewer people, and the spread between the draws is wider than the effect. A
seventh is the number to hold; the third of the first batch was chance dressed
as a finding, and taking that batch alone as the answer is exactly the mistake
this file exists to prevent.

Nothing per-agent moved at all. Pooled over the three batches `fed` went 0.570
to 0.563 and `all` - the gate a birth has to pass - went 0.288 to 0.292, both
of them smaller moves than the batches make on their own. `lasted` is 24 of 24
on all three with none extinct. The settlements are not doing worse on the
land they have; there is a little less of it and a few fewer of them.

This was taken with the map at its default eighty by thirty-six, which is the
width Upland is quoted at. The high country is scaled to the width of the map
actually being made, so a wider map is a bigger country at the same
ruggedness - at two hundred and forty wide the peaks stand at some seven
hundred and fifty metres instead of two hundred and sixty and the flanks keep
their slope, where holding the height fixed had put the same mountains over
three times the ground and flattened them. The valley is deliberately not
scaled with them, so the flood plain and the soil that reads off it are the
same on any map, and the lowland is simply a larger share of a larger map:
ploughable ground goes from 55% of the tiles at the default width to about
63% at two hundred and forty. A batch taken on a bigger map is still a
different measurement and not comparable to this one.

## How much of that is chance

The same code on two further batches of twenty-four seeds, which `-offset`
draws from a stretch of the seed space the baseline never touched:

```bash
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 24 -quiet
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 48 -quiet
```

```
offset  0: gates: fed 0.54 safe 0.50 held 0.82 all 0.268 food 3.94 hungry-with-food 0.27 | lasted 24/24 extinct 0 mean 154.8 median 137 | phys 0.58 safe 0.51 belng 0.53 estm 0.67
offset 24: gates: fed 0.57 safe 0.54 held 0.85 all 0.307 food 4.37 hungry-with-food 0.27 | lasted 24/24 extinct 0 mean 190.2 median 166 | phys 0.59 safe 0.51 belng 0.56 estm 0.60
offset 48: gates: fed 0.58 safe 0.52 held 0.84 all 0.300 food 4.95 hungry-with-food 0.26 | lasted 24/24 extinct 0 mean 218.7 median 202 | phys 0.60 safe 0.50 belng 0.56 estm 0.65
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 155, 190, 219, the median 137, 166, 202. `lasted` was 24 of 24
on all three and none of them lost a settlement outright. The mean is the
loosest reading here and always has been: a settlement that runs away is worth
as much as the twenty that did not, and a couple of seeds over five hundred are
most of the distance between the draws. The spread is if anything wider on this
tree than it was on the last, because a map with mountains on it gives its
seeds more to differ about - how much of the lowland a range happens to cover
is now part of what a seed decides.

What holds still is the per-agent side. `fed` moved by 0.04 across the three,
the mean needs by 0.01 to 0.07, and `all` sat between 0.268 and 0.307. So:

| reading | moved by chance | worth believing at |
|---|---|---|
| fed, and the four mean needs | 0.01-0.07 | 0.05 on two batches that agree |
| gates all | 0.039 | 0.05 |
| lasted, extinct | 0 of 24, 0 of 24 | 5 of 24 |
| mean population | 34% of itself | all three batches, pooled |
| median population | 39% of itself | all three batches, pooled |

The population readings are looser on this tree than they were on the last -
a third of themselves across the draws, where it used to be a quarter - and
the mountains are why: how much of a map's lowland a range happens to cover is
now one of the things a seed decides, so the seeds have more to differ about.

So a change that only moves the population numbers has not been shown to do
anything, and one batch cannot show it either way. Take all three and pool
them before believing any of it, and say in the commit what the pooled figure
was. The change that landed the mountains is the worked example: its first
batch said a third fewer people, and three batches said a seventh.

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

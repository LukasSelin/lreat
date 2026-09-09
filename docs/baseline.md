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
   1  591   630     39    169    250 |  0.60  0.46  0.47  0.51  0.03 |  1.00
   2  276   313     34     69    127 |  0.54  0.51  0.59  0.79  0.11 |  1.00
   3  113   190     43     69     55 |  0.54  0.53  0.54  0.61  0.03 |  1.00
   4  233   379     49     81    158 |  0.58  0.55  0.48  0.63  0.09 |  1.00
   5   53   145     26     34     36 |  0.57  0.58  0.38  0.44  0.03 |  0.99
   6  146   421     21     72    102 |  0.55  0.55  0.67  0.82  0.05 |  1.00
   7  259   393     52    108    177 |  0.52  0.51  0.56  0.61  0.06 |  1.00
   8   98   187     35     50     68 |  0.65  0.55  0.60  0.54  0.06 |  1.00
   9  343   730     34    147    198 |  0.57  0.55  0.61  0.61  0.07 |  1.00
  10  192   184     27     71    110 |  0.61  0.53  0.59  0.74  0.11 |  1.00
  11  446   343     24    158    441 |  0.60  0.48  0.47  0.53  0.05 |  1.00
  12   41   145     52     21     10 |  0.53  0.58  0.59  0.61  0.15 |  1.00
  13   30   159     55     24     17 |  0.50  0.48  0.54  0.54  0.05 |  0.46
  14   73   168     71     55     31 |  0.48  0.60  0.41  0.42  0.04 |  0.86
  15  469   780     57    126    325 |  0.58  0.48  0.52  0.54  0.05 |  1.00
  16   68    72      9      4     63 |  0.60  0.39  0.74  0.81  0.06 |  0.98
  17  139   253     18     42    109 |  0.54  0.49  0.56  0.67  0.08 |  1.00
  18  447   359     33     96    322 |  0.59  0.51  0.51  0.67  0.07 |  1.00
  19  358   322     41    171    131 |  0.64  0.50  0.44  0.53  0.05 |  0.99
  20  273   201     38    128     63 |  0.61  0.53  0.56  0.66  0.12 |  1.00
  21  448   837     29     99    357 |  0.72  0.47  0.61  0.64  0.08 |  1.00
  22  104   298     37     12     56 |  0.56  0.52  0.69  0.75  0.19 |  0.99
  23  202   294     70     81    167 |  0.57  0.54  0.57  0.64  0.03 |  1.00
  24   84   288     42     51     68 |  0.56  0.59  0.47  0.60  0.08 |  1.00

dwell/rest                         4372009  29.2%
take/berries@wood                  3730454  24.9%
consume/provision                  2510438  16.8%
pass/practice>pupil                1262482   8.4%
dwell/guard@market                 1118705   7.5%
take/fish@water                     613007   4.1%
dwell/meet@tavern>neighbour         382742   2.6%
take/timber@wood                    196266   1.3%
take/grain@field                    157340   1.1%
transfer/provision>needy            157336   1.1%
pass/practice>self                  152941   1.0%
tend/plant@open                      64436   0.4%
raise/timber>dwelling@open           59813   0.4%
exchange/coin>provision@market       56538   0.4%
exchange/material>coin@market        54156   0.4%
transfer/provision<holder            21774   0.1%
transfer/material>requester          20779   0.1%
tend/clear@open                       7441   0.0%
make/timber>tool@bench                7245   0.0%
strike/person>wrongdoer               5768   0.0%
take/game@wood                        5251   0.0%
tend/water@field                      4245   0.0%
raise/timber>road@ground              3541   0.0%
take/stone@outcrop                    2584   0.0%
dwell/look                             476   0.0%
make/provision+timber>meal@hearth      169   0.0%
raise/timber>tavern@open               167   0.0%
move@dwelling                           85   0.0%
raise/timber+stone>granary@open         66   0.0%
raise/timber+stone>market@open          41   0.0%
make/stone+timber>tool@forge            31   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.55 safe 0.54 held 0.82 all 0.295 food 4.43 hungry-with-food 0.27 | lasted 24/24 extinct 0 mean 228.6 median 202 | phys 0.58 safe 0.52 belng 0.55 estm 0.62
```

## How much of that is chance

The same code on two further batches of twenty-four seeds, which `-offset`
draws from a stretch of the seed space the baseline never touched:

```bash
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 24 -quiet
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 48 -quiet
```

```
offset  0: gates: fed 0.55 safe 0.54 held 0.82 all 0.295 food 4.43 hungry-with-food 0.27 | lasted 24/24 extinct 0 mean 228.6 median 202 | phys 0.58 safe 0.52 belng 0.55 estm 0.62
offset 24: gates: fed 0.58 safe 0.51 held 0.82 all 0.287 food 4.88 hungry-with-food 0.25 | lasted 22/24 extinct 1 mean 239.4 median 211 | phys 0.61 safe 0.50 belng 0.58 estm 0.66
offset 48: gates: fed 0.55 safe 0.55 held 0.83 all 0.292 food 4.85 hungry-with-food 0.28 | lasted 24/24 extinct 0 mean 199.5 median 190 | phys 0.59 safe 0.52 belng 0.53 estm 0.65
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 229, 239, 200, the median 202, 211, 190. `lasted` was 24, 22
and 24 of 24, and one of the three lost a settlement outright where the others
lost none. The mean is the
loosest reading here and always has been: a settlement that runs away is worth
as much as the twenty that did not, and the two seeds above sitting over five
hundred are most of the distance between the draws.

What holds still is the per-agent side. `fed` moved by 0.03 across the three,
the mean needs by 0.02 to 0.05, and `all` sat between 0.287 and 0.295. So:

| reading | moved by chance | worth believing at |
|---|---|---|
| fed, and the four mean needs | 0.00-0.07 | 0.05 on two batches that agree |
| gates all | 0.008 | 0.05 |
| lasted, extinct | 2 of 24, 1 of 24 | 5 of 24 |
| mean population | 8% of itself, 27% before | a second batch that agrees |
| median population | 11% of itself | a second batch that agrees |

A change that only moves the population numbers has not been shown to do
anything. Run it again on `-offset 24` before believing it, and say in the
commit that both batches agreed.

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

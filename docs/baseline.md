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
   1  363   382     28    127    167 |  0.66  0.49  0.51  0.56  0.02 |  1.00
   2  452   226     21    105    274 |  0.64  0.49  0.62  0.79  0.11 |  1.00
   3   36   108     43     20     15 |  0.49  0.53  0.61  0.75  0.12 |  0.99
   4  118   319     62     70     76 |  0.60  0.57  0.52  0.68  0.10 |  0.98
   5   78   208     41     36     98 |  0.48  0.55  0.47  0.46  0.06 |  0.99
   6  115   251     28     28    106 |  0.56  0.52  0.64  0.72  0.09 |  1.00
   7  201   228     43     81    154 |  0.60  0.51  0.59  0.62  0.07 |  1.00
   8  113   254     28     44     75 |  0.67  0.55  0.61  0.65  0.11 |  1.00
   9  206   579     57    102    151 |  0.60  0.55  0.58  0.59  0.05 |  1.00
  10  190   182     27     60    124 |  0.53  0.51  0.56  0.69  0.07 |  1.00
  11  625   343     33    158    599 |  0.53  0.49  0.50  0.56  0.06 |  1.00
  12   76   188     46     52     24 |  0.49  0.58  0.58  0.53  0.08 |  0.99
  13  162   363     39     90    112 |  0.61  0.57  0.63  0.65  0.05 |  0.98
  14   75   158     39     37     16 |  0.59  0.56  0.42  0.47  0.05 |  0.92
  15  480   535     49    128    225 |  0.58  0.46  0.48  0.43  0.03 |  1.00
  16    5    54      7      2      4 |  0.10  0.46  0.45  0.70  0.03 |  0.99
  17  134   294     33     39    124 |  0.52  0.49  0.56  0.66  0.10 |  1.00
  18  256   260     28     86    176 |  0.56  0.51  0.47  0.60  0.05 |  1.00
  19  168   281     31    121    125 |  0.61  0.56  0.61  0.70  0.04 |  0.99
  20  263   266     44    107     36 |  0.54  0.50  0.55  0.56  0.08 |  1.00
  21  382   840     36     93    286 |  0.61  0.50  0.57  0.61  0.06 |  1.00
  22  264   339     38     42    134 |  0.56  0.49  0.64  0.70  0.15 |  1.00
  23   92   206     71     39     69 |  0.44  0.58  0.57  0.60  0.04 |  0.98
  24  188   293     39     72    143 |  0.60  0.52  0.48  0.66  0.09 |  1.00

dwell/rest                         3715441  27.7%
take/berries@wood                  3335765  24.9%
consume/provision                  2249727  16.8%
pass/practice>pupil                1190186   8.9%
dwell/guard@market                 1101392   8.2%
take/fish@water                     552988   4.1%
dwell/meet@tavern>neighbour         368489   2.7%
take/timber@wood                    171995   1.3%
take/grain@field                    143633   1.1%
pass/practice>self                  143070   1.1%
transfer/provision>needy            140585   1.0%
raise/timber>dwelling@open           61243   0.5%
tend/plant@open                      61237   0.5%
exchange/coin>provision@market       54971   0.4%
exchange/material>coin@market        51683   0.4%
transfer/material>requester          19307   0.1%
transfer/provision<holder            18561   0.1%
make/timber>tool@bench                8039   0.1%
tend/clear@open                       7159   0.1%
raise/timber>road@ground              7151   0.1%
take/game@wood                        6042   0.0%
strike/person>wrongdoer               5833   0.0%
tend/water@field                      4283   0.0%
take/stone@outcrop                    2794   0.0%
dwell/look                             609   0.0%
make/provision+timber>meal@hearth      288   0.0%
raise/timber>tavern@open               166   0.0%
raise/timber+stone>granary@open        100   0.0%
move@dwelling                           88   0.0%
raise/timber+stone>market@open          42   0.0%
make/stone+timber>tool@forge            38   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.56 safe 0.53 held 0.84 all 0.293 food 4.36 hungry-with-food 0.27 | lasted 23/24 extinct 0 mean 210.1 median 188 | phys 0.55 safe 0.52 belng 0.55 estm 0.62
```

## How much of that is chance

The same code on two further batches of twenty-four seeds, which `-offset`
draws from a stretch of the seed space the baseline never touched:

```bash
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 24 -quiet
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 48 -quiet
```

```
offset  0: gates: fed 0.56 safe 0.53 held 0.84 all 0.293 food 4.36 hungry-with-food 0.27 | lasted 23/24 extinct 0 mean 210.1 median 188 | phys 0.55 safe 0.52 belng 0.55 estm 0.62
offset 24: gates: fed 0.58 safe 0.49 held 0.83 all 0.291 food 5.22 hungry-with-food 0.25 | lasted 23/24 extinct 0 mean 226.8 median 190 | phys 0.61 safe 0.51 belng 0.59 estm 0.65
offset 48: gates: fed 0.56 safe 0.55 held 0.84 all 0.302 food 5.97 hungry-with-food 0.29 | lasted 23/24 extinct 0 mean 217.0 median 211 | phys 0.58 safe 0.52 belng 0.56 estm 0.67
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 210, 227, 217, the median 188, 190, 211. `lasted` was 23 of 24
in all three, and none of the three lost a settlement outright, where a draw
before this one lost one. The mean is the loosest reading here and always has been: a
settlement that runs away is worth as much as the twenty that did not, and the
two seeds above sitting over five hundred are most of the distance between the
draws.

What holds still is the per-agent side. `fed` moved by 0.02 across the three,
the mean needs by 0.01 to 0.06, and `all` sat between 0.291 and 0.302. So:

| reading | moved by chance | worth believing at |
|---|---|---|
| fed, and the four mean needs | 0.00-0.07 | 0.05 on two batches that agree |
| gates all | 0.011 | 0.05 |
| lasted, extinct | 0 of 24, 0 of 24 | 5 of 24 |
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

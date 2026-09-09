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
   1  522   552     32    130    129 |  0.63  0.45  0.41  0.43  0.01 |  1.00
   2  212   252     38     71    104 |  0.56  0.50  0.61  0.75  0.08 |  1.00
   3  177   297     39     81     89 |  0.53  0.49  0.53  0.52  0.04 |  1.00
   4   82   221     48     61     57 |  0.54  0.67  0.51  0.60  0.12 |  1.00
   5   56   175     22     37     37 |  0.51  0.56  0.59  0.65  0.04 |  0.86
   6  193   329     22     48     99 |  0.61  0.48  0.70  0.77  0.17 |  1.00
   7  212   314     37    102    191 |  0.62  0.49  0.57  0.61  0.06 |  0.99
   8  198   153     23     67     37 |  0.67  0.49  0.48  0.48  0.02 |  1.00
   9  217   602     51    118    104 |  0.56  0.58  0.54  0.55  0.06 |  1.00
  10  179   174     32     55     76 |  0.65  0.49  0.43  0.59  0.09 |  1.00
  11  376   388     23    129    412 |  0.51  0.47  0.50  0.64  0.03 |  1.00
  12   56   125     50     24     24 |  0.60  0.53  0.50  0.40  0.17 |  0.99
  13   95   287     55     54    106 |  0.57  0.60  0.54  0.66  0.07 |  0.96
  14   47   157     49     49     32 |  0.51  0.53  0.54  0.56  0.02 |  0.44
  15  476   576     43    135    300 |  0.59  0.46  0.56  0.53  0.03 |  1.00
  16   87   155     51     47     50 |  0.60  0.54  0.58  0.71  0.04 |  1.00
  17   53   158     22     26     57 |  0.69  0.53  0.66  0.76  0.12 |  0.99
  18  272   375     21    111    178 |  0.56  0.51  0.31  0.55  0.07 |  1.00
  19  448   329     31    197    213 |  0.59  0.45  0.41  0.48  0.03 |  1.00
  20  224   209     35    118     56 |  0.56  0.52  0.62  0.64  0.10 |  0.97
  21  963  1102     20    127    608 |  0.65  0.48  0.64  0.64  0.08 |  1.00
  22  194   324     43     15     66 |  0.57  0.50  0.66  0.66  0.16 |  1.00
  23  163   292     37     78    153 |  0.58  0.56  0.56  0.69  0.05 |  1.00
  24  214   390     44     89    152 |  0.58  0.53  0.43  0.59  0.10 |  1.00

dwell/rest                         4331381  29.1%
take/berries@wood                  3648000  24.5%
consume/provision                  2448079  16.5%
pass/practice>pupil                1261100   8.5%
dwell/guard@market                 1183577   8.0%
take/fish@water                     610824   4.1%
dwell/meet@tavern>neighbour         379132   2.5%
take/timber@wood                    202252   1.4%
take/grain@field                    169327   1.1%
transfer/provision>needy            156801   1.1%
pass/practice>self                  148935   1.0%
exchange/coin>provision@market       69420   0.5%
tend/plant@open                      62652   0.4%
exchange/material>coin@market        62547   0.4%
raise/timber>dwelling@open           60332   0.4%
transfer/material>requester          23731   0.2%
transfer/provision<holder            21236   0.1%
make/timber>tool@bench                8749   0.1%
tend/clear@open                       7087   0.0%
strike/person>wrongdoer               6344   0.0%
take/game@wood                        5077   0.0%
tend/water@field                      4478   0.0%
raise/timber>road@ground              3687   0.0%
take/stone@outcrop                    3321   0.0%
dwell/look                             490   0.0%
make/provision+timber>meal@hearth      260   0.0%
raise/timber>tavern@open               162   0.0%
raise/timber+stone>granary@open        116   0.0%
move@dwelling                          110   0.0%
make/stone+timber>tool@forge            77   0.0%
raise/timber+stone>market@open          50   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.56 safe 0.53 held 0.83 all 0.291 food 4.42 hungry-with-food 0.27 | lasted 24/24 extinct 0 mean 238.2 median 198 | phys 0.58 safe 0.52 belng 0.54 estm 0.60
```

## How much of that is chance

The same code on two further batches of twenty-four seeds, which `-offset`
draws from a stretch of the seed space the baseline never touched:

```bash
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 24 -quiet
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 48 -quiet
```

```
offset  0: gates: fed 0.56 safe 0.53 held 0.83 all 0.291 food 4.42 hungry-with-food 0.27 | lasted 24/24 extinct 0 mean 238.2 median 198 | phys 0.58 safe 0.52 belng 0.54 estm 0.60
offset 24: gates: fed 0.57 safe 0.52 held 0.83 all 0.292 food 4.47 hungry-with-food 0.25 | lasted 24/24 extinct 0 mean 229.5 median 202 | phys 0.58 safe 0.51 belng 0.58 estm 0.64
offset 48: gates: fed 0.56 safe 0.55 held 0.84 all 0.299 food 5.02 hungry-with-food 0.29 | lasted 24/24 extinct 0 mean 221.2 median 193 | phys 0.60 safe 0.52 belng 0.55 estm 0.66
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 238, 230, 221, the median 198, 202, 193. `lasted` was 24 of 24
in all three, and none of them lost a settlement outright. The mean is the
loosest reading here and always has been: a settlement that runs away is worth
as much as the twenty that did not, and the two seeds above sitting over five
hundred are most of the distance between the draws.

What holds still is the per-agent side. `fed` moved by 0.01 across the three,
the mean needs by 0.02 to 0.06, and `all` sat between 0.291 and 0.299. So:

| reading | moved by chance | worth believing at |
|---|---|---|
| fed, and the four mean needs | 0.00-0.07 | 0.05 on two batches that agree |
| gates all | 0.008 | 0.05 |
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

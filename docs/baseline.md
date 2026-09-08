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
   1  435   521     39    121    149 |  0.62  0.47  0.50  0.47  0.01 |  1.00
   2  378   297     25     89    236 |  0.66  0.49  0.65  0.81  0.12 |  1.00
   3   69   110     20     24     31 |  0.53  0.51  0.75  0.80  0.13 |  1.00
   4  106   215     26     37     87 |  0.64  0.58  0.72  0.76  0.16 |  0.95
   5   93   202     23     30     56 |  0.66  0.49  0.66  0.76  0.17 |  1.00
   6  195   315     31     61    150 |  0.57  0.52  0.65  0.75  0.09 |  1.00
   7  160   239     31     69    169 |  0.58  0.48  0.66  0.55  0.05 |  1.00
   8  102   174     19     43     67 |  0.65  0.53  0.48  0.44  0.05 |  1.00
   9  249   551     44     97    162 |  0.56  0.54  0.56  0.63  0.08 |  1.00
  10  414   196     20     64    233 |  0.65  0.46  0.55  0.71  0.08 |  1.00
  11  326   278     28    120    263 |  0.52  0.50  0.52  0.66  0.06 |  1.00
  12   56   231     69     42     30 |  0.51  0.57  0.56  0.46  0.13 |  0.98
  13  226   341     34     68    195 |  0.59  0.51  0.65  0.72  0.18 |  1.00
  14   59   121     37     31     11 |  0.59  0.51  0.50  0.60  0.00 |  0.91
  15  405   555     34    109    364 |  0.64  0.46  0.53  0.52  0.03 |  1.00
  16   33   101     20     11     36 |  0.57  0.44  0.71  0.82  0.07 |  0.75
  17   90   297     48     39    101 |  0.60  0.52  0.62  0.71  0.09 |  0.98
  18  348   454     35     76    233 |  0.61  0.49  0.49  0.63  0.07 |  1.00
  19  171   308     31     93     60 |  0.66  0.53  0.64  0.58  0.05 |  1.00
  20  471   312     49    165     60 |  0.52  0.50  0.49  0.48  0.11 |  1.00
  21  416   626     38    100    273 |  0.64  0.50  0.56  0.61  0.10 |  1.00
  22  262   533     23     30    122 |  0.56  0.50  0.67  0.79  0.18 |  1.00
  23  127   193     31     54     87 |  0.53  0.52  0.57  0.60  0.03 |  1.00
  24  176   221     34     61    134 |  0.58  0.51  0.43  0.61  0.05 |  1.00

dwell/rest                         4276959  29.2%
take/berries@wood                  3598012  24.5%
consume/provision                  2365911  16.1%
pass/practice>pupil                1271607   8.7%
dwell/guard@market                 1186243   8.1%
take/fish@water                     572065   3.9%
dwell/meet@tavern>neighbour         399802   2.7%
take/timber@wood                    177434   1.2%
pass/practice>self                  168628   1.2%
take/grain@field                    153728   1.0%
transfer/provision>needy            142154   1.0%
exchange/coin>provision@market       67275   0.5%
raise/timber>dwelling@open           67250   0.5%
exchange/material>coin@market        59476   0.4%
tend/plant@open                      58541   0.4%
transfer/provision<holder            28752   0.2%
transfer/material>requester          22357   0.2%
make/timber>tool@bench                9215   0.1%
take/game@wood                        6995   0.0%
tend/clear@open                       6855   0.0%
raise/timber>road@ground              6603   0.0%
strike/person>wrongdoer               6263   0.0%
tend/water@field                      4483   0.0%
take/stone@outcrop                    3468   0.0%
dwell/look                             567   0.0%
make/provision+timber>meal@hearth      193   0.0%
raise/timber>tavern@open               163   0.0%
raise/timber+stone>granary@open        144   0.0%
move@dwelling                           97   0.0%
make/stone+timber>tool@forge            66   0.0%
raise/timber+stone>market@open          49   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.56 safe 0.52 held 0.84 all 0.285 food 4.22 hungry-with-food 0.27 | lasted 24/24 extinct 0 mean 223.6 median 195 | phys 0.59 safe 0.51 belng 0.59 estm 0.65
```

## How much of that is chance

The same code on two further batches of twenty-four seeds, which `-offset`
draws from a stretch of the seed space the baseline never touched:

```bash
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 24 -quiet
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 48 -quiet
```

```
offset  0: gates: fed 0.56 safe 0.52 held 0.84 all 0.285 food 4.22 hungry-with-food 0.27 | lasted 24/24 extinct 0 mean 223.6 median 195 | phys 0.59 safe 0.51 belng 0.59 estm 0.65
offset 24: gates: fed 0.58 safe 0.50 held 0.82 all 0.279 food 4.46 hungry-with-food 0.25 | lasted 23/24 extinct 0 mean 230.2 median 198 | phys 0.59 safe 0.50 belng 0.55 estm 0.63
offset 48: gates: fed 0.57 safe 0.54 held 0.84 all 0.300 food 5.21 hungry-with-food 0.28 | lasted 24/24 extinct 0 mean 201.3 median 175 | phys 0.60 safe 0.53 belng 0.57 estm 0.67
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 224, 230, 201, the median 195, 198, 175. `lasted` was 24, 23
and 24 of 24, and none of the three lost a settlement outright. The mean is the loosest reading here and always has been: a
settlement that runs away is worth as much as the twenty that did not, and the
two seeds above sitting over five hundred are most of the distance between the
draws.

What holds still is the per-agent side. `fed` moved by 0.02 across the three,
the mean needs by 0.01 to 0.05, and `all` sat between 0.279 and 0.300. So:

| reading | moved by chance | worth believing at |
|---|---|---|
| fed, and the four mean needs | 0.00-0.07 | 0.05 on two batches that agree |
| gates all | 0.021 | 0.05 |
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

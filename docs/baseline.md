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
   1  547   490     25    144    184 |  0.68  0.46  0.50  0.47  0.01 |  1.00
   2  348   457     43    112    216 |  0.55  0.50  0.65  0.71  0.10 |  1.00
   3   88   213     39     44     51 |  0.51  0.55  0.58  0.64  0.13 |  1.00
   4   89   220     43     63     33 |  0.56  0.62  0.47  0.59  0.11 |  1.00
   5   52   108     20     13     28 |  0.69  0.45  0.64  0.71  0.06 |  1.00
   6  129   224     45     80    111 |  0.52  0.59  0.60  0.71  0.04 |  1.00
   7  107   163     19     39     66 |  0.54  0.43  0.55  0.70  0.09 |  1.00
   8  128   113     11     27     24 |  0.67  0.50  0.58  0.62  0.02 |  1.00
   9  244   616     71    123     87 |  0.57  0.56  0.52  0.53  0.06 |  1.00
  10  174    99     16     36     54 |  0.64  0.44  0.72  0.72  0.09 |  1.00
  11  328   413     36    116    310 |  0.58  0.49  0.58  0.69  0.06 |  1.00
  12   63   324     56     55     33 |  0.56  0.60  0.51  0.57  0.08 |  0.98
  13  206   279     47     47    116 |  0.62  0.51  0.66  0.73  0.13 |  1.00
  14   53   133     44     42     12 |  0.57  0.61  0.61  0.69  0.08 |  0.91
  15  541   447     27    125    414 |  0.62  0.46  0.54  0.55  0.04 |  1.00
  16    1    34     15      1      1 |  0.84  0.63  0.00  0.43  0.37 |  0.00
  17   33   181     69     23     28 |  0.52  0.57  0.51  0.67  0.04 |  0.64
  18  234   377     33    112    171 |  0.58  0.56  0.41  0.61  0.08 |  1.00
  19  393   328     43    146    104 |  0.58  0.46  0.42  0.43  0.02 |  0.96
  20  313   229     42    140     21 |  0.58  0.52  0.56  0.62  0.14 |  1.00
  21  360   540     44     93    236 |  0.58  0.53  0.58  0.58  0.07 |  1.00
  22  102   270     34     19     58 |  0.54  0.50  0.55  0.64  0.07 |  1.00
  23   62   178     53     40     44 |  0.56  0.57  0.64  0.71  0.07 |  1.00
  24  131   238     55     81     93 |  0.54  0.59  0.44  0.52  0.08 |  1.00

dwell/rest                         3581053  29.4%
take/berries@wood                  3116172  25.5%
consume/provision                  1961636  16.1%
dwell/guard@market                  985613   8.1%
pass/practice>pupil                 984972   8.1%
take/fish@water                     457471   3.8%
dwell/meet@tavern>neighbour         324247   2.7%
take/timber@wood                    176015   1.4%
pass/practice>self                  119366   1.0%
take/grain@field                    118710   1.0%
transfer/provision>needy            114190   0.9%
raise/timber>dwelling@open           52340   0.4%
tend/plant@open                      47956   0.4%
exchange/coin>provision@market       47209   0.4%
exchange/material>coin@market        43915   0.4%
transfer/material>requester          17460   0.1%
transfer/provision<holder            15615   0.1%
make/timber>tool@bench                6193   0.1%
tend/clear@open                       5566   0.0%
strike/person>wrongdoer               5349   0.0%
take/game@wood                        4683   0.0%
raise/timber>road@ground              3647   0.0%
tend/water@field                      3345   0.0%
take/stone@outcrop                    2225   0.0%
dwell/look                             793   0.0%
make/provision+timber>meal@hearth      329   0.0%
raise/timber>tavern@open               157   0.0%
move@dwelling                           95   0.0%
raise/timber+stone>granary@open         61   0.0%
raise/timber+stone>market@open          32   0.0%
make/stone+timber>tool@forge            27   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.56 safe 0.54 held 0.83 all 0.298 food 4.24 hungry-with-food 0.27 | lasted 23/24 extinct 0 mean 196.9 median 131 | phys 0.59 safe 0.53 belng 0.53 estm 0.62
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

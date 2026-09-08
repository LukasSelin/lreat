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
   1  400   502      7    131    212 |  0.72  0.47  0.60  0.64  0.03 |  1.00
   2  279   291     32    111    130 |  0.56  0.51  0.62  0.65  0.09 |  1.00
   3  220   339     31     69    104 |  0.53  0.50  0.58  0.58  0.08 |  1.00
   4   80   220     23     32     70 |  0.56  0.56  0.60  0.69  0.15 |  1.00
   5   58   158     10     40     49 |  0.65  0.53  0.48  0.47  0.01 |  1.00
   6  158   298     26     78    178 |  0.58  0.56  0.71  0.78  0.06 |  1.00
   7  202   270     44     81    142 |  0.53  0.50  0.50  0.57  0.10 |  0.98
   8   66   225     19     19     72 |  0.58  0.57  0.54  0.66  0.05 |  1.00
   9  227   627     50    115    145 |  0.57  0.55  0.54  0.50  0.04 |  1.00
  10  168   235     20     46    150 |  0.62  0.51  0.53  0.68  0.12 |  1.00
  11  344   341     42    142    338 |  0.58  0.53  0.46  0.53  0.05 |  1.00
  12   58   215     77     45     33 |  0.55  0.62  0.61  0.41  0.08 |  0.98
  13  120   317     38     53    134 |  0.57  0.49  0.57  0.66  0.06 |  1.00
  14   12    96     38     12      6 |  0.50  0.54  0.68  0.74  0.07 |  0.34
  15  357   573     39    111    306 |  0.54  0.48  0.52  0.54  0.03 |  1.00
  16  101   165     59     53     66 |  0.54  0.58  0.48  0.51  0.05 |  0.96
  17   40   184     57     24     55 |  0.45  0.61  0.68  0.68  0.10 |  0.97
  18  203   299     28     88    123 |  0.56  0.49  0.25  0.46  0.04 |  1.00
  19  246   309     62    135    127 |  0.61  0.51  0.51  0.57  0.05 |  0.95
  20  239   271     37     99     68 |  0.55  0.51  0.45  0.56  0.11 |  1.00
  21  400   614      4     97    338 |  0.72  0.52  0.71  0.68  0.08 |  1.00
  22   22   175     57     15     24 |  0.64  0.58  0.61  0.68  0.13 |  0.87
  23   96   272     42     50     88 |  0.58  0.58  0.58  0.72  0.03 |  1.00
  24  157   219     45     71    141 |  0.56  0.53  0.47  0.57  0.06 |  1.00

dwell/rest                         3655042  29.2%
take/berries@wood                  2875806  23.0%
consume/provision                  2131539  17.0%
pass/practice>pupil                1047544   8.4%
dwell/guard@market                  915102   7.3%
take/fish@water                     660131   5.3%
dwell/meet@tavern>neighbour         324545   2.6%
take/timber@wood                    186009   1.5%
take/grain@field                    155603   1.2%
transfer/provision>needy            128931   1.0%
pass/practice>self                  124776   1.0%
tend/plant@open                      56326   0.4%
exchange/coin>provision@market       56157   0.4%
raise/timber>dwelling@open           54813   0.4%
exchange/material>coin@market        53433   0.4%
transfer/provision<holder            24452   0.2%
raise/timber>road@ground             19727   0.2%
transfer/material>requester          17523   0.1%
make/timber>tool@bench                9018   0.1%
tend/clear@open                       7008   0.1%
strike/person>wrongdoer               6152   0.0%
take/game@wood                        4553   0.0%
tend/water@field                      4153   0.0%
take/stone@outcrop                    2501   0.0%
dwell/look                             458   0.0%
raise/timber>tavern@open               173   0.0%
make/provision+timber>meal@hearth      169   0.0%
move@dwelling                          112   0.0%
raise/timber+stone>granary@open         67   0.0%
make/stone+timber>tool@forge            49   0.0%
raise/timber+stone>market@open          44   0.0%

born 0.15 inherit 0.05 temp 0.15

gates: fed 0.57 safe 0.54 held 0.80 all 0.293 food 4.58 hungry-with-food 0.26 | lasted 23/24 extinct 0 mean 177.2 median 168 | phys 0.58 safe 0.53 belng 0.55 estm 0.61
```

## How much of that is chance

The same code on two further batches of twenty-four seeds, which `-offset`
draws from a stretch of the seed space the baseline never touched:

```bash
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 24 -quiet
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 48 -quiet
```

```
offset  0: gates: fed 0.57 safe 0.54 held 0.80 all 0.293 food 4.58 hungry-with-food 0.26 | lasted 23/24 extinct 0 mean 177.2 median 168 | phys 0.58 safe 0.53 belng 0.55 estm 0.61
offset 24: gates: fed 0.57 safe 0.51 held 0.81 all 0.277 food 4.49 hungry-with-food 0.24 | lasted 23/24 extinct 1 mean 173.8 median 192 | phys 0.59 safe 0.51 belng 0.59 estm 0.63
offset 48: gates: fed 0.57 safe 0.55 held 0.83 all 0.310 food 6.19 hungry-with-food 0.28 | lasted 23/24 extinct 0 mean 189.8 median 172 | phys 0.59 safe 0.54 belng 0.59 estm 0.68
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 177, 174, 190, the median 168, 192, 172. `lasted` was 23 of 24
in all three, and one of the three lost a settlement outright where the others
lost none. A settlement that runs away is worth as much as the twenty that did
not - two seeds in the batch above sat on the 400 cap - so the thresholds below
are kept wide.

What holds still is the per-agent side. `fed` was 0.57 in all three, the mean
needs moved by 0.01 to 0.07, and `all` sat between 0.277 and 0.310. So:

| reading | moved by chance | worth believing at |
|---|---|---|
| fed, and the four mean needs | 0.00-0.07 | 0.05 on two batches that agree |
| gates all | 0.033 | 0.05 |
| lasted, extinct | 0 of 24, 1 of 24 | 5 of 24 |
| mean and median population | 9% and 14% of themselves | a second batch that agrees |

A change that only moves the population numbers has not been shown to do
anything. Run it again on `-offset 24` before believing it, and say in the
commit that both batches agreed.

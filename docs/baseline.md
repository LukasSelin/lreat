# Baseline

What `tune` prints on master, checked in so that nobody has to run it to find
out. The run below is the default batch:

```bash
go run ./cmd/tune -seeds 24 -ticks 6000
```

Nothing in it is sampled from the wall clock, so the same commit gives the same
numbers however the goroutines interleave — the two runs behind the table below
were byte-identical. That is what makes a checked-in baseline worth anything:
the file is the run, and re-running it only confirms it.

Read it before a change rather than measuring master again. Two hours of work
begins with two minutes of a batch that already has an answer here, and the
answer does not change while the work is going on. Take the comparison run once
the change is settled instead, and when it lands, replace what is below with the
new numbers in the same commit — the file describes the tree it is committed in,
so a change to behaviour that leaves it alone has left it wrong.

The headline is the last line: how many of the 24 settlements were still at
least their founding size at tick 6000, how many died out, and where the mean
population landed. The `gates` line in front of it is the share of fertile
agents fed, safe and held — `all` is the share that were all three at once,
which is the gate a birth has to pass, and the one worth watching.

## master

```
seed  pop  died births houses fields |  phys  safe belng  estm  actl | order
   1   21    63     51     15      3 |  0.51  0.46  0.46  0.79  0.08 |  0.82
   2  121    75     23     63     38 |  0.54  0.45  0.72  0.84  0.05 |  1.00
   3   53    39     56     27     18 |  0.62  0.49  0.37  0.21  0.01 |  1.00
   4  178   129     76     78     18 |  0.60  0.44  0.44  0.52  0.08 |  0.58
   5   15    72     44     12      7 |  0.51  0.48  0.64  0.75  0.11 |  0.90
   6   83   112     44     16     42 |  0.68  0.49  0.71  0.72  0.10 |  1.00
   7  111    72     69     67     15 |  0.66  0.51  0.69  0.69  0.21 |  1.00
   8   46    80     45     24     48 |  0.48  0.45  0.55  0.60  0.03 |  0.88
   9  186   137     95     83     16 |  0.53  0.55  0.61  0.56  0.06 |  1.00
  10   67    55     65     26      7 |  0.52  0.55  0.62  0.67  0.08 |  0.99
  11  340   105     41    150    174 |  0.47  0.48  0.55  0.51  0.03 |  1.00
  12  263   117     69     89     36 |  0.56  0.51  0.56  0.36  0.10 |  0.99
  13  329   101     59    149     41 |  0.34  0.28  0.50  0.01  0.00 |  0.59
  14  142    68     88     70     24 |  0.62  0.52  0.51  0.44  0.04 |  0.97
  15  333    54     64     69    117 |  0.63  0.51  0.57  0.55  0.14 |  1.00
  16   87   100     90     57      9 |  0.46  0.39  0.48  0.57  0.02 |  0.49
  17  397   116     23    109    105 |  0.57  0.47  0.55  0.71  0.18 |  0.99
  18  154    66     64     16     44 |  0.72  0.48  0.50  0.45  0.08 |  1.00
  19  226   116     52    117     17 |  0.63  0.42  0.66  0.50  0.06 |  0.83
  20  329    58    108     90     24 |  0.61  0.42  0.63  0.47  0.15 |  0.66
  21   86    41     25     51     44 |  0.60  0.51  0.51  0.15  0.03 |  0.99
  22   26    65     13      5     12 |  0.47  0.40  0.75  0.86  0.15 |  1.00
  23  167   165     63    100     45 |  0.42  0.54  0.52  0.61  0.04 |  0.99
  24  140    91     66     62     40 |  0.51  0.51  0.70  0.59  0.10 |  1.00

take/berries@wood                   960399  28.3%
dwell/rest                          827216  24.4%
consume/provision                   580867  17.1%
dwell/guard@market                  235542   6.9%
take/fish@water                     177696   5.2%
pass/practice>pupil                 169328   5.0%
dwell/meet@tavern>neighbour         127635   3.8%
take/timber@wood                    106139   3.1%
tend/plant@open                      57284   1.7%
pass/practice>self                   29737   0.9%
transfer/material>requester          29261   0.9%
raise/timber>dwelling@open           21817   0.6%
take/grain@field                     10933   0.3%
exchange/coin>provision@market       10015   0.3%
take/game@wood                        7799   0.2%
exchange/material>coin@market         6493   0.2%
transfer/provision>needy              6440   0.2%
make/timber>tool@bench                5939   0.2%
raise/timber>road@ground              4703   0.1%
transfer/provision<holder             4075   0.1%
take/stone@outcrop                    3358   0.1%
strike/person>wrongdoer               3241   0.1%
tend/clear@open                       2585   0.1%
dwell/look                             617   0.0%
make/provision+timber>meal@hearth      473   0.0%
raise/timber+stone>granary@open        170   0.0%
raise/timber>tavern@open               134   0.0%
tend/water@field                        67   0.0%
move@dwelling                           37   0.0%
raise/timber+stone>market@open          31   0.0%
make/stone+timber>tool@forge             6   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.47 safe 0.29 held 0.61 all 0.135 food 2.21 hungry-with-food 0.25 | lasted 23/24 extinct 0 mean 162.5 median 142 | phys 0.55 safe 0.47 belng 0.57 estm 0.55
```

## How much of that is chance

The same code on two further batches of twenty-four seeds, which `-offset`
draws from a stretch of the seed space the baseline never touched:

```bash
go run ./cmd/tune -seeds 24 -ticks 6000 -offset 24 -quiet
go run ./cmd/tune -seeds 24 -ticks 6000 -offset 48 -quiet
```

```
offset  0: gates: fed 0.47 safe 0.29 held 0.61 all 0.135 food 2.21 hungry-with-food 0.25 | lasted 23/24 extinct 0 mean 162.5 median 142 | phys 0.55 safe 0.47 belng 0.57 estm 0.55
offset 24: gates: fed 0.49 safe 0.26 held 0.66 all 0.124 food 1.82 hungry-with-food 0.21 | lasted 21/24 extinct 0 mean 144.8 median 102 | phys 0.60 safe 0.47 belng 0.64 estm 0.60
offset 48: gates: fed 0.46 safe 0.27 held 0.66 all 0.126 food 2.13 hungry-with-food 0.27 | lasted 22/24 extinct 0 mean 112.7 median 102 | phys 0.58 safe 0.51 belng 0.61 estm 0.56
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 163, 145, 113, the median 142, 102, 102. The median is the
unsteady one: a batch is a handful of settlements that run away and a long tail
that get by, so which side of that gap the twelfth seed falls on moves the
median further than any change to the rules has moved it. Three batches are
nowhere near enough to say the swing has narrowed. A settlement that runs away
is worth as much as the twenty that did not, and one of them still lands in
every few batches, so the thresholds below are kept where the wider spread put
them.

What holds still is the per-agent side. `fed` moved 0.03 across all three, the
mean needs 0.01 to 0.07, `all` 0.124 to 0.135, and `lasted` and `extinct` a
couple of settlements. So:

| reading | moved by chance | worth believing at |
|---|---|---|
| fed, and the four mean needs | 0.01-0.07 | 0.05 on two batches that agree |
| gates all | 0.009 | 0.04 |
| lasted, extinct | 2 of 24, 1 of 24 | 5 of 24 |
| mean and median population | 12% of itself here, 60% before | a second batch that agrees |

A change that only moves the population numbers has not been shown to do
anything. Run it again on `-offset 24` before believing it, and say in the
commit that both batches agreed.

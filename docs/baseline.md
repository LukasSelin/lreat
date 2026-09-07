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
   1  192    81     38    120      8 |  0.62  0.49  0.34  0.52  0.03 |  1.00
   2   68    76     65      2     31 |  0.62  0.47  0.76  0.85  0.37 |  1.00
   3  240    74     93     77     40 |  0.60  0.50  0.48  0.11  0.08 |  1.00
   4  400   253     32     23     16 |  0.57  0.53  0.58  0.47  0.20 |  1.00
   5    4    36     20      4      3 |  0.91  0.27  0.37  0.73  0.09 |  0.00
   6  331   234     46     14     16 |  0.64  0.50  0.66  0.44  0.15 |  1.00
   7  316   112     42    139     46 |  0.60  0.47  0.42  0.35  0.05 |  1.00
   8   81    36     66     32     36 |  0.62  0.51  0.55  0.42  0.11 |  0.99
   9  221    68     55     24     45 |  0.64  0.46  0.61  0.64  0.17 |  1.00
  10   20    30     26      7     20 |  0.47  0.39  0.71  0.82  0.15 |  0.98
  11  115    90     57     59      9 |  0.64  0.51  0.64  0.40  0.03 |  0.99
  12  157    72     36     10     48 |  0.71  0.49  0.50  0.74  0.42 |  1.00
  13  136    98     50     66    103 |  0.56  0.39  0.59  0.69  0.10 |  0.68
  14  115   114     61     80     25 |  0.67  0.47  0.59  0.52  0.05 |  0.61
  15  139   108     99     75     12 |  0.60  0.50  0.45  0.36  0.01 |  0.84
  16   91    97     88     50     13 |  0.62  0.55  0.66  0.62  0.04 |  0.99
  17   63    66     44     22     49 |  0.49  0.42  0.66  0.87  0.12 |  0.90
  18   68    60     31     40     39 |  0.60  0.47  0.76  0.60  0.06 |  1.00
  19  193   104     79     92     19 |  0.62  0.44  0.43  0.52  0.04 |  0.85
  20   54    55     51     33     16 |  0.59  0.45  0.62  0.51  0.01 |  0.91
  21  123    65     70     50     49 |  0.73  0.48  0.50  0.01  0.01 |  1.00
  22  257   106     51     98     75 |  0.49  0.49  0.77  0.52  0.22 |  0.99
  23  311   126     60    157    280 |  0.49  0.51  0.64  0.36  0.03 |  1.00
  24   54    68     93     33     11 |  0.51  0.56  0.54  0.52  0.08 |  0.99

take/berries@wood                   894714  27.3%
dwell/rest                          842769  25.7%
consume/provision                   546995  16.7%
dwell/guard@market                  257869   7.9%
take/fish@water                     164238   5.0%
pass/practice>pupil                 142313   4.3%
dwell/meet@tavern>neighbour         123453   3.8%
take/timber@wood                     97711   3.0%
tend/plant@open                      46402   1.4%
transfer/material>requester          40055   1.2%
pass/practice>self                   31961   1.0%
raise/timber>dwelling@open           14386   0.4%
take/grain@field                     12595   0.4%
transfer/provision>needy             11690   0.4%
raise/timber>road@ground             10589   0.3%
exchange/coin>provision@market        8469   0.3%
exchange/material>coin@market         5896   0.2%
take/game@wood                        5213   0.2%
make/timber>tool@bench                4980   0.2%
transfer/provision<holder             3833   0.1%
strike/person>wrongdoer               2709   0.1%
take/stone@outcrop                    2692   0.1%
tend/clear@open                       2526   0.1%
dwell/look                             424   0.0%
make/provision+timber>meal@hearth      185   0.0%
raise/timber+stone>granary@open        139   0.0%
move@dwelling                           68   0.0%
tend/water@field                        59   0.0%
raise/timber>tavern@open                26   0.0%
make/stone+timber>tool@forge             6   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.49 safe 0.27 held 0.62 all 0.135 food 2.22 hungry-with-food 0.25 | lasted 23/24 extinct 0 mean 156.2 median 136 | phys 0.61 safe 0.47 belng 0.58 estm 0.52
```

## How much of that is chance

The same code on two further batches of twenty-four seeds, which `-offset`
draws from a stretch of the seed space the baseline never touched:

```bash
go run ./cmd/tune -seeds 24 -ticks 6000 -offset 24 -quiet
go run ./cmd/tune -seeds 24 -ticks 6000 -offset 48 -quiet
```

```
offset  0: gates: fed 0.49 safe 0.27 held 0.62 all 0.135 food 2.22 hungry-with-food 0.25 | lasted 23/24 extinct 0 mean 156.2 median 136 | phys 0.61 safe 0.47 belng 0.58 estm 0.52
offset 24: gates: fed 0.48 safe 0.24 held 0.59 all 0.114 food 1.98 hungry-with-food 0.23 | lasted 21/24 extinct 0 mean 120.5 median 131 | phys 0.56 safe 0.47 belng 0.59 estm 0.52
offset 48: gates: fed 0.48 safe 0.25 held 0.68 all 0.120 food 2.07 hungry-with-food 0.24 | lasted 20/24 extinct 2 mean 95.2 median 91 | phys 0.58 safe 0.49 belng 0.60 estm 0.54
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 156, 121, 95. Twenty-four seeds is not enough to say anything
with it, and neither is the median, which went 136, 131, 91. A settlement that
runs away is worth as much as the twenty that did not, and one of them lands in
every few batches.

What holds still is the per-agent side. `fed` moved 0.01 across all three, the
mean needs 0.01 to 0.03, `all` 0.11 to 0.14, and `lasted` and `extinct` a couple
of settlements. So:

| reading | moved by chance | worth believing at |
|---|---|---|
| fed, and the four mean needs | 0.01-0.03 | 0.05 |
| gates all | 0.026 | 0.04 |
| lasted, extinct | 3 of 24, 2 of 24 | 5 of 24 |
| mean and median population | 60% of itself | a second batch that agrees |

A change that only moves the population numbers has not been shown to do
anything. Run it again on `-offset 24` before believing it, and say in the
commit that both batches agreed.

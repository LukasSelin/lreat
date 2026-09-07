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
   3  257    66     93     88     42 |  0.60  0.49  0.50  0.11  0.07 |  0.92
   4  400   252     49     22     20 |  0.65  0.51  0.61  0.59  0.25 |  1.00
   5    4    36     20      4      3 |  0.91  0.27  0.37  0.73  0.09 |  0.00
   6  375   303     99     25     21 |  0.63  0.54  0.70  0.45  0.16 |  1.00
   7  316   112     42    139     46 |  0.60  0.47  0.42  0.35  0.05 |  1.00
   8   67    33     55     31     33 |  0.52  0.52  0.52  0.56  0.07 |  1.00
   9  226    60     59     22     28 |  0.65  0.48  0.62  0.66  0.19 |  1.00
  10   20    30     26      7     20 |  0.47  0.39  0.71  0.82  0.15 |  0.98
  11  143    85     63     60     11 |  0.65  0.51  0.62  0.39  0.01 |  0.99
  12  210    91     63      5     53 |  0.66  0.50  0.38  0.62  0.34 |  1.00
  13  147   116     65     74    115 |  0.49  0.50  0.65  0.66  0.07 |  0.97
  14   79   123     53     67     20 |  0.54  0.32  0.49  0.49  0.03 |  0.33
  15  139   108     99     75     12 |  0.60  0.50  0.45  0.36  0.01 |  0.84
  16   80    77     62     44     18 |  0.59  0.49  0.74  0.73  0.04 |  0.89
  17   63    66     44     22     49 |  0.49  0.42  0.66  0.87  0.12 |  0.90
  18   93    59     40     49     49 |  0.65  0.46  0.71  0.57  0.04 |  1.00
  19   91   141     34     26     14 |  0.66  0.32  0.49  0.61  0.03 |  0.31
  20   73    51     54     40      7 |  0.58  0.58  0.45  0.44  0.03 |  0.98
  21  123    65     70     50     49 |  0.73  0.48  0.50  0.01  0.01 |  1.00
  22  260   114     30    106     71 |  0.49  0.47  0.79  0.48  0.23 |  0.88
  23  311   126     60    157    280 |  0.49  0.51  0.64  0.36  0.03 |  1.00
  24   41    78     90     25     11 |  0.56  0.58  0.58  0.43  0.12 |  0.99

forage           896888  27.3%
rest             857693  26.1%
eat              551754  16.8%
guard            257154   7.8%
fish             159121   4.8%
teach            144327   4.4%
socialize        122530   3.7%
gather wood       96931   3.0%
plant trees       44385   1.4%
fulfil request    36327   1.1%
study             31526   1.0%
build shelter     13810   0.4%
farm              12183   0.4%
give              10914   0.3%
lay road          10486   0.3%
buy food           8659   0.3%
sell               6047   0.2%
hunt               5335   0.2%
craft              5056   0.2%
steal              3533   0.1%
quarry             2915   0.1%
retaliate          2736   0.1%
clear field        2569   0.1%
scout               423   0.0%
build granary       141   0.0%
cook                139   0.0%
irrigate             73   0.0%
move house           70   0.0%
build tavern         27   0.0%
smelt                 9   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.49 safe 0.28 held 0.62 all 0.136 food 2.17 hungry-with-food 0.24 | lasted 23/24 extinct 0 mean 157.4 median 139 | phys 0.60 safe 0.47 belng 0.57 estm 0.53
```

## How much of that is chance

The same code on two further batches of twenty-four seeds, which `-offset`
draws from a stretch of the seed space the baseline never touched:

```bash
go run ./cmd/tune -seeds 24 -ticks 6000 -offset 24 -quiet
go run ./cmd/tune -seeds 24 -ticks 6000 -offset 48 -quiet
```

```
offset  0: gates: fed 0.49 safe 0.28 held 0.62 all 0.136 food 2.17 hungry-with-food 0.24 | lasted 23/24 extinct 0 mean 157.4 median 139 | phys 0.60 safe 0.47 belng 0.57 estm 0.53
offset 24: gates: fed 0.48 safe 0.24 held 0.59 all 0.110 food 2.01 hungry-with-food 0.23 | lasted 20/24 extinct 0 mean 116.2 median 122 | phys 0.57 safe 0.46 belng 0.60 estm 0.53
offset 48: gates: fed 0.48 safe 0.25 held 0.67 all 0.122 food 2.10 hungry-with-food 0.24 | lasted 20/24 extinct 2 mean 96.5 median 75 | phys 0.57 safe 0.49 belng 0.59 estm 0.53
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 157, 116, 97. Twenty-four seeds is not enough to say anything
with it, and neither is the median, which went 139, 122, 75. A settlement that
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

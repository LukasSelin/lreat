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
begins with half a minute of a batch that already has an answer here, and the
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
   1  154    85     64     84     28 |  0.69  0.52  0.45  0.64  0.04 |  1.00
   2  170    90     34     43     39 |  0.55  0.46  0.68  0.82  0.14 |  1.00
   3   94   114     56     53     17 |  0.61  0.55  0.68  0.75  0.03 |  1.00
   4  148   171     59    108     25 |  0.45  0.53  0.43  0.62  0.10 |  0.73
   5   36    50     63     13     11 |  0.56  0.56  0.56  0.53  0.11 |  0.99
   6   97   101     46     21     45 |  0.61  0.51  0.62  0.71  0.11 |  1.00
   7   53    71     70     31     28 |  0.60  0.33  0.57  0.71  0.19 |  0.20
   8  158    73     53     57     88 |  0.52  0.50  0.55  0.54  0.07 |  1.00
   9   62    68     50     25     20 |  0.42  0.50  0.59  0.45  0.13 |  1.00
  10  107    53     32     31     23 |  0.69  0.46  0.78  0.74  0.10 |  0.99
  11  174    87     17    105     92 |  0.51  0.51  0.56  0.73  0.05 |  1.00
  12  159   114     58     87     15 |  0.49  0.48  0.55  0.29  0.07 |  0.99
  13   92   130     49     81     22 |  0.53  0.45  0.57  0.55  0.05 |  0.49
  14   10    52     42     13      5 |  0.36  0.45  0.66  0.76  0.00 |  0.26
  15  123    59     21     22     47 |  0.64  0.43  0.61  0.42  0.14 |  1.00
  16  140    96     69     13     33 |  0.57  0.50  0.61  0.53  0.19 |  1.00
  17  131    86     49     69     58 |  0.61  0.27  0.57  0.79  0.04 |  0.45
  18   70    54     38     20     47 |  0.60  0.45  0.53  0.45  0.07 |  1.00
  19   94    61     43     60     19 |  0.64  0.47  0.63  0.67  0.03 |  0.93
  20  235    68     39     85     21 |  0.59  0.49  0.65  0.59  0.11 |  1.00
  21   55    42     31     25     29 |  0.59  0.52  0.76  0.66  0.10 |  1.00
  22  124    99     42     19     35 |  0.47  0.46  0.47  0.49  0.16 |  0.96
  23  106    64     95     63     35 |  0.52  0.53  0.50  0.62  0.02 |  0.99
  24   58    80     61     23     24 |  0.58  0.47  0.78  0.78  0.07 |  0.99

dwell/rest                          751644  25.9%
take/berries@wood                   734464  25.3%
consume/provision                   480491  16.6%
dwell/guard@market                  235438   8.1%
take/fish@water                     180045   6.2%
pass/practice>pupil                 160651   5.5%
dwell/meet@tavern>neighbour         101793   3.5%
take/timber@wood                     88021   3.0%
tend/plant@open                      31792   1.1%
pass/practice>self                   25279   0.9%
transfer/material>requester          23411   0.8%
raise/timber>dwelling@open           13617   0.5%
take/grain@field                     11098   0.4%
exchange/coin>provision@market       10486   0.4%
raise/timber>road@ground              8981   0.3%
transfer/provision>needy              8539   0.3%
take/game@wood                        7311   0.3%
exchange/material>coin@market         6203   0.2%
make/timber>tool@bench                4793   0.2%
transfer/provision<holder             4763   0.2%
strike/person>wrongdoer               3521   0.1%
take/stone@outcrop                    2873   0.1%
tend/clear@open                       2431   0.1%
dwell/look                             610   0.0%
make/provision+timber>meal@hearth      162   0.0%
raise/timber>tavern@open               132   0.0%
raise/timber+stone>granary@open        124   0.0%
tend/water@field                        34   0.0%
move@dwelling                           27   0.0%
raise/timber+stone>market@open          21   0.0%
make/stone+timber>tool@forge            12   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.46 safe 0.26 held 0.60 all 0.115 food 2.09 hungry-with-food 0.25 | lasted 23/24 extinct 0 mean 110.4 median 107 | phys 0.56 safe 0.48 belng 0.60 estm 0.62
```

## How much of that is chance

The same code on two further batches of twenty-four seeds, which `-offset`
draws from a stretch of the seed space the baseline never touched:

```bash
go run ./cmd/tune -seeds 24 -ticks 6000 -offset 24 -quiet
go run ./cmd/tune -seeds 24 -ticks 6000 -offset 48 -quiet
```

```
offset  0: gates: fed 0.46 safe 0.26 held 0.60 all 0.115 food 2.09 hungry-with-food 0.25 | lasted 23/24 extinct 0 mean 110.4 median 107 | phys 0.56 safe 0.48 belng 0.60 estm 0.62
offset 24: gates: fed 0.47 safe 0.28 held 0.60 all 0.124 food 2.12 hungry-with-food 0.24 | lasted 23/24 extinct 0 mean 120.9 median 108 | phys 0.57 safe 0.48 belng 0.59 estm 0.55
offset 48: gates: fed 0.46 safe 0.26 held 0.64 all 0.115 food 2.23 hungry-with-food 0.26 | lasted 21/24 extinct 1 mean 107.8 median 111 | phys 0.57 safe 0.49 belng 0.62 estm 0.59
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 110, 121, 108, the median 107, 108, 111. The three draws happen
to sit closer together than they have before — an earlier tree gave 123, 105, 85
on the same offsets — and three batches are nowhere near enough to say the swing
has narrowed. A settlement that runs away is worth as much as the twenty that
did not, and one of them still lands in every few batches, so the thresholds
below are kept where the wider spread put them.

What holds still is the per-agent side. `fed` moved 0.01 across all three, the
mean needs 0.01 to 0.07, `all` 0.115 to 0.124, and `lasted` and `extinct` a
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

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
   1   76    80     91     33      3 |  0.55  0.48  0.24  0.43  0.04 |  0.78
   2  109    87     58     66     15 |  0.48  0.47  0.52  0.54  0.04 |  0.99
   3   38    40     41     11      5 |  0.36  0.41  0.59  0.24  0.03 |  0.96
   4  249   134    104     91     47 |  0.47  0.53  0.46  0.44  0.12 |  1.00
   5   87    70     81     45      3 |  0.64  0.51  0.58  0.43  0.00 |  1.00
   6   90   105     38     30     78 |  0.56  0.45  0.67  0.78  0.08 |  1.00
   7   59    54     54     33     38 |  0.58  0.46  0.65  0.77  0.07 |  0.87
   8   47    92     69     38     13 |  0.46  0.43  0.51  0.55  0.01 |  0.47
   9   59    53     52     18     20 |  0.44  0.45  0.71  0.60  0.03 |  0.99
  10  125    39     59     69     77 |  0.65  0.51  0.55  0.62  0.08 |  0.99
  11  164    86     45     90     77 |  0.42  0.50  0.54  0.52  0.04 |  1.00
  12  302   144     72    110     29 |  0.56  0.50  0.52  0.42  0.11 |  0.99
  13  276   129     48    107    140 |  0.43  0.44  0.47  0.01  0.00 |  0.91
  14  124    93     80     97     20 |  0.54  0.42  0.51  0.36  0.02 |  0.55
  15   78    50     33     23     78 |  0.59  0.45  0.53  0.51  0.02 |  1.00
  16   22    36     20      9     25 |  0.69  0.44  0.78  0.82  0.10 |  0.94
  17  115    88     74     50     92 |  0.55  0.52  0.44  0.60  0.08 |  0.99
  18  130    59    105     12     75 |  0.58  0.57  0.67  0.72  0.32 |  1.00
  19  213    76     51    124     12 |  0.62  0.37  0.51  0.37  0.04 |  0.69
  20  171    56     52     91     10 |  0.60  0.38  0.63  0.58  0.11 |  0.64
  21  102    64     33     42     27 |  0.52  0.48  0.51  0.36  0.04 |  1.00
  22   69    64     34     19      7 |  0.44  0.41  0.63  0.54  0.21 |  0.95
  23  142   201     42    103     33 |  0.40  0.54  0.51  0.62  0.02 |  0.98
  24  108    77     92     62     24 |  0.47  0.55  0.46  0.49  0.04 |  1.00

take/berries@wood                   742767  26.4%
dwell/rest                          674699  24.0%
consume/provision                   495374  17.6%
dwell/guard@market                  215719   7.7%
take/fish@water                     172439   6.1%
pass/practice>pupil                 129674   4.6%
take/timber@wood                    106563   3.8%
dwell/meet@tavern>neighbour         106339   3.8%
tend/plant@open                      45605   1.6%
pass/practice>self                   20502   0.7%
transfer/material>requester          17850   0.6%
raise/timber>dwelling@open           14521   0.5%
take/grain@field                     12680   0.5%
raise/timber>road@ground             12214   0.4%
exchange/coin>provision@market        8840   0.3%
transfer/provision>needy              8128   0.3%
take/game@wood                        5985   0.2%
exchange/material>coin@market         5401   0.2%
make/timber>tool@bench                5200   0.2%
transfer/provision<holder             4326   0.2%
strike/person>wrongdoer               3192   0.1%
tend/clear@open                       2471   0.1%
take/stone@outcrop                    2466   0.1%
dwell/look                             503   0.0%
make/provision+timber>meal@hearth      150   0.0%
raise/timber+stone>granary@open        114   0.0%
raise/timber>tavern@open               114   0.0%
tend/water@field                        48   0.0%
move@dwelling                           45   0.0%
raise/timber+stone>market@open          17   0.0%
make/stone+timber>tool@forge             2   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.45 safe 0.28 held 0.60 all 0.121 food 2.11 hungry-with-food 0.24 | lasted 24/24 extinct 0 mean 123.1 median 109 | phys 0.53 safe 0.47 belng 0.55 estm 0.51
```

## How much of that is chance

The same code on two further batches of twenty-four seeds, which `-offset`
draws from a stretch of the seed space the baseline never touched:

```bash
go run ./cmd/tune -seeds 24 -ticks 6000 -offset 24 -quiet
go run ./cmd/tune -seeds 24 -ticks 6000 -offset 48 -quiet
```

```
offset  0: gates: fed 0.45 safe 0.28 held 0.60 all 0.121 food 2.11 hungry-with-food 0.24 | lasted 24/24 extinct 0 mean 123.1 median 109 | phys 0.53 safe 0.47 belng 0.55 estm 0.51
offset 24: gates: fed 0.46 safe 0.23 held 0.65 all 0.103 food 1.93 hungry-with-food 0.22 | lasted 18/24 extinct 0 mean 105.2 median 77 | phys 0.59 safe 0.46 belng 0.63 estm 0.62
offset 48: gates: fed 0.47 safe 0.25 held 0.64 all 0.112 food 2.26 hungry-with-food 0.25 | lasted 21/24 extinct 0 mean 85.2 median 72 | phys 0.60 safe 0.47 belng 0.61 estm 0.65
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 123, 105, 85. Twenty-four seeds is not enough to say anything
with it, and neither is the median, which went 109, 77, 72. A settlement that
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

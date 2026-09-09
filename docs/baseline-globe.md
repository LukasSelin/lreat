# Baseline on the globe

What `tune` prints on the globe, checked in so that nobody has to run it to
find out. The run below is the globe batch:

```bash
go run ./cmd/tune -preset globe -seeds 8 -ticks 21600
```

It is eight seeds rather than twenty-four because a globe is a hundred and
eighty times the ground of the valley and each seed is a single thread here,
so the batch is half an hour where the valley's is a few minutes. It is taken
on the same terms otherwise: sixty years, twenty founders, one settlement
founded where the terrain scorer puts it.

This is not the valley's baseline and none of its numbers compare to
[baseline.md](baseline.md). The world is a cylinder a thousand tiles round,
a third of it sea, cold at the poles and warm at the middle, and a founding
party is set down on a coast where a laden walker is cut off by water on
one side or another. One of eight settlements died with its founders and none
grew past four hundred; that is the globe's ecology as it stands, before the
settlements to come have bounded land of their own, and it is what a change to
the globe is held against. Everything that is measured on the valley is
measured on the valley still, and is unchanged by any of this.

Taken on the commit that gave the weather a lapse rate, where a mountain is
colder than the valley under it and the peaks carry snow and a tree line; see
"What the last change did" in [baseline.md](baseline.md). **It is not a
measurement of that change.** The run it replaces was taken on `a49cf5f`, the
end of the first phase of the globe, and master has had the mountains, the
hedges, the roads wearing in and much else since - so everything between these
two batches is in the difference, not the lapse rate alone. What moved is
worth knowing and not worth attributing: `held` 0.47 to 0.66, `lasted` 4 of 8
to 6 of 8, one extinction instead of two, and the mean population 302 to 102
on eight seeds, which is the loosest number in either file on the smallest
batch either file takes. Whoever next changes the globe should take this
batch again before and after, on one tree, and then the difference will mean
something.

The cost of a day was measured on `a49cf5f` and has not been taken again: on
an idle machine of twenty-four cores a day here with forty people cost four to
six milliseconds, twenty of the hundred and twenty-eight chunks awake, against
half a millisecond for a day on the valley with forty. The whole batch above -
eight seeds, sixty years each - took 2m38s.

## The batch

```
seed  pop  died births houses fields |  phys  safe belng  estm  actl | order
   1    3    21      4      0      8 |  0.64  0.41  0.93  0.92  0.00 |  0.94
   2   37    96    113     24     60 |  0.66  0.56  0.70  0.72  0.07 |  0.93
   3   87    79     78     33     70 |  0.66  0.49  0.55  0.75  0.01 |  1.00
   4  219    37     47     23    181 |  0.76  0.43  0.86  0.78  0.03 |  1.00
   5   35    67     82     23     20 |  0.51  0.52  0.41  0.77  0.04 |  1.00
   6   98   227    187     51    131 |  0.71  0.60  0.65  0.81  0.11 |  1.00
   7  341    80    104     57    151 |  0.59  0.44  0.91  0.83  0.09 |  0.95
   8    0    25      5      1      0 |  0.00  0.00  0.00  0.00  0.00 |  0.00

dwell/rest                          509829  27.4%
take/berries@wood                   375100  20.2%
consume/provision                   334603  18.0%
dwell/guard@market                  201680  10.8%
pass/practice>pupil                 185089   9.9%
dwell/meet@tavern>neighbour          74556   4.0%
take/grain@field                     42894   2.3%
transfer/provision>needy             33047   1.8%
take/timber@wood                     23326   1.3%
take/fish@water                      20449   1.1%
pass/practice>self                   14976   0.8%
dwell/look                           13042   0.7%
raise/timber>dwelling@open            7993   0.4%
exchange/material>coin@market         7623   0.4%
exchange/coin>provision@market        4929   0.3%
tend/plant@open                       3042   0.2%
transfer/material>requester           1872   0.1%
transfer/provision<holder             1806   0.1%
make/timber>tool@bench                1281   0.1%
tend/clear@open                       1237   0.1%
take/game@wood                        1036   0.1%
raise/timber>road@ground               575   0.0%
tend/water@field                       424   0.0%
take/stone@outcrop                     390   0.0%
strike/person>wrongdoer                270   0.0%
move@dwelling                           69   0.0%
make/provision+timber>meal@hearth       54   0.0%
raise/timber>tavern@open                33   0.0%
make/stone+timber>tool@forge            32   0.0%
raise/timber+stone>granary@open         19   0.0%
raise/timber+stone>market@open           3   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.63 safe 0.44 held 0.66 all 0.204 food 19.58 hungry-with-food 0.23 | lasted 6/8 extinct 1 mean 102.5 median 87 | phys 0.65 safe 0.50 belng 0.71 estm 0.80
```

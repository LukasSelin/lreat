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
one side or another. Two of eight settlements died with their founders and
one grew past a thousand; that is the globe's ecology as it stands, before
the settlements to come have bounded land of their own, and it is what a
change to the globe is held against. Everything that is measured on the
valley is measured on the valley still, and is unchanged by any of this.

Taken on `a49cf5f`, the end of the first phase of the globe: the wrap, the
chunks, the sleeping ground, the file of where everybody is, the windowed
router, the sea and the weather by latitude. On an idle machine of
twenty-four cores a day here with forty people costs four to six
milliseconds, twenty of the hundred and twenty-eight chunks awake; a day on
the valley with forty costs half a millisecond.

## The batch

```
seed  pop  died births houses fields |  phys  safe belng  estm  actl | order
   1 1008   187    110    521    879 |  0.67  0.55  0.71  0.74  0.04 |  0.99
   2    0    20      0      0      0 |  0.00  0.00  0.00  0.00  0.00 |  0.00
   3   59    51     90     28     86 |  0.75  0.63  0.64  0.62  0.07 |  1.00
   4    0    21      1      0      0 |  0.00  0.00  0.00  0.00  0.00 |  0.00
   5 1211   588    120    498   1038 |  0.68  0.50  0.55  0.70  0.04 |  1.00
   6  132   196    107     47    136 |  0.67  0.51  0.63  0.79  0.08 |  1.00
   7    7    27     14      4     10 |  0.65  0.63  0.91  0.97  0.15 |  0.97
   8    1    24      5      1      3 |  0.37  0.63  0.00  0.04  0.00 |  0.00

dwell/rest                         1811419  39.4%
take/berries@wood                   933646  20.3%
consume/provision                   717629  15.6%
pass/practice>pupil                 473891  10.3%
take/fish@water                     144508   3.1%
dwell/guard@market                  134867   2.9%
dwell/meet@tavern>neighbour          92800   2.0%
dwell/look                           77627   1.7%
take/timber@wood                     64722   1.4%
take/grain@field                     43914   1.0%
pass/practice>self                   27395   0.6%
transfer/provision>needy             26113   0.6%
raise/timber>dwelling@open           18999   0.4%
tend/plant@open                       9778   0.2%
exchange/material>coin@market         6271   0.1%
exchange/coin>provision@market        3680   0.1%
tend/clear@open                       3140   0.1%
transfer/provision<holder             2851   0.1%
take/game@wood                        2566   0.1%
make/timber>tool@bench                2494   0.1%
transfer/material>requester           1340   0.0%
raise/timber>road@ground              1172   0.0%
tend/water@field                       782   0.0%
take/stone@outcrop                     512   0.0%
strike/person>wrongdoer                385   0.0%
make/provision+timber>meal@hearth      139   0.0%
move@dwelling                           42   0.0%
raise/timber+stone>granary@open         39   0.0%
raise/timber>tavern@open                36   0.0%
make/stone+timber>tool@forge            36   0.0%
raise/timber+stone>market@open           8   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.67 safe 0.45 held 0.47 all 0.196 food 19.89 hungry-with-food 0.25 | lasted 4/8 extinct 2 mean 302.2 median 59 | phys 0.63 safe 0.57 belng 0.57 estm 0.64
```

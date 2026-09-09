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
one side or another. One of the eight died out, one is down to three souls,
and three grew past five hundred; that is the globe's ecology as it stands,
before the settlements to come have bounded land of their own, and it is what
a change to the globe is held against. Everything that is measured on the valley is
measured on the valley still, and is unchanged by any of this.

Taken on the commit that gave the rivers a shape: a channel cuts the outside
of its own bends and walks across its valley, and how fast it cuts is divided
by the rock it is cutting. See "What the last change did" in
[baseline.md](baseline.md).

Like the batch before it, this is a measurement of that change and of nothing
else: it was taken on the commit immediately previous, with nothing in
between.

```
before the rivers: gates: fed 0.70 safe 0.52 held 0.62 all 0.257 food 22.36 | lasted 6/8 extinct 0 mean 133.2 median 90
after the rivers:  gates: fed 0.68 safe 0.51 held 0.66 all 0.272 food 14.06 | lasted 6/8 extinct 1 mean 254.9 median 124
```

Eight seeds is too small a batch to believe, and this one shows why: the mean
population has doubled, one settlement has died where none did, and the food
on hand has fallen by a third - three readings pointing three ways. What is
worth noting is `food`, which is the one figure here far outside anything the
valley's thresholds would call noise: 22.4 to 14.1. On a globe a settlement
lives off wild ground far more than a valley one does, and rivers that move
are rivers that drown and re-make the ground it forages. Whether that is the
change or the draw wants a batch that could carry the question.

The cost of a day was measured on `a49cf5f` and has not been taken again: on
an idle machine of twenty-four cores a day here with forty people cost four to
six milliseconds, twenty of the hundred and twenty-eight chunks awake, against
half a millisecond for a day on the valley with forty. The whole batch above -
eight seeds, sixty years each - took 2m38s.

## The batch

```
seed  pop  died births houses fields |  phys  safe belng  estm  actl | order
   1   45   132    157     24     62 |  0.65  0.59  0.74  0.76  0.17 |  1.00
   2   28   189    137     19     47 |  0.80  0.67  0.80  0.73  0.09 |  0.97
   3  124   112     75     51     78 |  0.66  0.48  0.61  0.84  0.03 |  1.00
   4  772   121    144    336    407 |  0.67  0.53  0.54  0.67  0.03 |  1.00
   5    0    25      5      0      0 |  0.00  0.00  0.00  0.00  0.00 |  0.00
   6  515   329    133    270    471 |  0.68  0.55  0.70  0.69  0.07 |  0.99
   7  552   164    136    207    256 |  0.65  0.50  0.65  0.82  0.06 |  1.00
   8    3    23      6      3      6 |  0.62  0.74  0.99  0.99  0.21 |  0.89

dwell/rest                         1416252  37.5%
take/berries@wood                   654753  17.3%
consume/provision                   648057  17.1%
pass/practice>pupil                 374915   9.9%
dwell/guard@market                  256471   6.8%
dwell/meet@tavern>neighbour          94983   2.5%
transfer/provision>needy             64412   1.7%
take/grain@field                     62911   1.7%
take/timber@wood                     58430   1.5%
take/fish@water                      53683   1.4%
pass/practice>self                   26049   0.7%
raise/timber>dwelling@open           15538   0.4%
exchange/material>coin@market        12941   0.3%
dwell/look                           11015   0.3%
tend/plant@open                       9632   0.3%
exchange/coin>provision@market        5239   0.1%
make/timber>tool@bench                2531   0.1%
raise/timber>road@ground              2418   0.1%
tend/clear@open                       2407   0.1%
transfer/provision<holder             2230   0.1%
transfer/material>requester           2086   0.1%
take/game@wood                        1927   0.1%
tend/water@field                       952   0.0%
make/provision+timber>meal@hearth      319   0.0%
strike/person>wrongdoer                288   0.0%
move@dwelling                          226   0.0%
take/stone@outcrop                     115   0.0%
raise/timber>tavern@open                34   0.0%
make/stone+timber>tool@forge            19   0.0%
raise/timber+stone>granary@open          6   0.0%
raise/timber+stone>market@open           3   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.68 safe 0.51 held 0.66 all 0.272 food 14.06 hungry-with-food 0.22 | lasted 6/8 extinct 1 mean 254.9 median 124 | phys 0.68 safe 0.58 belng 0.72 estm 0.78
```

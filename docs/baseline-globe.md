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
one side or another. None of the eight died out, one is down to a single
soul, and two grew past three hundred; that is the globe's ecology as it
stands, before the settlements to come have bounded land of their own, and it
is what a change to the globe is held against. Everything that is measured on the valley is
measured on the valley still, and is unchanged by any of this.

Taken on the commit that gave the soil a make-up: there is rock under every
tile, the soil over it is a mixture weathered out of that rock, and the water
sorts what it carries. See "What the last change did" in
[baseline.md](baseline.md).

This one *is* a measurement of that change, and it is the first entry here
that is: the batch before it was taken on the commit immediately previous, and
nothing else moved in between.

```
before the soils: gates: fed 0.63 safe 0.44 held 0.66 all 0.204 food 19.58 | lasted 6/8 extinct 1 mean 102.5 median 87
after the soils:  gates: fed 0.70 safe 0.52 held 0.62 all 0.257 food 22.36 | lasted 6/8 extinct 0 mean 133.2 median 90
```

Eight seeds is far too small a batch to believe any of it - the valley's
thresholds are for twenty-four - but the direction is worth writing down
because it is the opposite of the valley's. There, soils were a draw; here
`fed` is up 0.07, `all` up 0.05, and the extinction is gone. The reading that
suggests itself is that a globe is where a soil's make-up has room to matter:
a founding party set down on one coast of a world may land on ground that its
own rock made good, and there are eight coasts rather than one valley. Nobody
should act on that until it is measured on a batch that could carry it.

The cost of a day was measured on `a49cf5f` and has not been taken again: on
an idle machine of twenty-four cores a day here with forty people cost four to
six milliseconds, twenty of the hundred and twenty-eight chunks awake, against
half a millisecond for a day on the valley with forty. The whole batch above -
eight seeds, sixty years each - took 2m38s.

## The batch

```
seed  pop  died births houses fields |  phys  safe belng  estm  actl | order
   1    1    20      1      1      3 |  0.84  0.15  0.00  0.03  0.00 |  0.00
   2   90   119    164     63     98 |  0.59  0.61  0.72  0.65  0.04 |  0.99
   3   31    84     95     13     31 |  0.73  0.55  0.69  0.53  0.02 |  0.98
   4  411    92     55    110    229 |  0.73  0.47  0.53  0.84  0.03 |  1.00
   5   30    46     56     14     33 |  0.59  0.52  0.49  0.95  0.05 |  1.00
   6  177   178    124     87    149 |  0.70  0.53  0.56  0.75  0.01 |  1.00
   7   15    50     45      2     15 |  0.78  0.44  0.82  0.82  0.02 |  0.99
   8  311   157    144    198    320 |  0.65  0.56  0.72  0.73  0.07 |  1.00

dwell/rest                         1013897  39.9%
take/berries@wood                   469791  18.5%
consume/provision                   363624  14.3%
dwell/guard@market                  254553  10.0%
pass/practice>pupil                 195683   7.7%
dwell/meet@tavern>neighbour          55502   2.2%
take/timber@wood                     35858   1.4%
take/grain@field                     35757   1.4%
take/fish@water                      28215   1.1%
transfer/provision>needy             23665   0.9%
dwell/look                           16546   0.7%
pass/practice>self                   12800   0.5%
raise/timber>dwelling@open            9974   0.4%
exchange/coin>provision@market        5661   0.2%
exchange/material>coin@market         4193   0.2%
make/timber>tool@bench                2218   0.1%
tend/plant@open                       2057   0.1%
raise/timber>road@ground              1963   0.1%
transfer/material>requester           1802   0.1%
tend/clear@open                       1654   0.1%
take/game@wood                        1399   0.1%
transfer/provision<holder              921   0.0%
tend/water@field                       485   0.0%
make/provision+timber>meal@hearth      450   0.0%
strike/person>wrongdoer                317   0.0%
move@dwelling                          205   0.0%
raise/timber>tavern@open                40   0.0%
take/stone@outcrop                       1   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.70 safe 0.52 held 0.62 all 0.257 food 22.36 hungry-with-food 0.21 | lasted 6/8 extinct 0 mean 133.2 median 90 | phys 0.70 safe 0.48 belng 0.57 estm 0.66
```

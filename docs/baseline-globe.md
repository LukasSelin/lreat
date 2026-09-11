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
one side or another. All eight came through, two are under a hundred souls,
and one grew past five hundred; that is the globe's ecology as it stands,
before the settlements to come have bounded land of their own, and it is what
a change to the globe is held against. Everything that is measured on the valley is
measured on the valley still, and is unchanged by any of this.

Taken on the commit that gave the globe mountains instead of blobs: the
ridges, the mask that says where the high country stands, and the height it
stands to are all measured in the length of a mountain range now, and not in
the width of the map. A globe's high ground was four or five smooth domes two
hundred tiles across and three thousand three hundred metres high, with slopes
a walker could not stand on and nothing on them to see; it is a dozen ranges
of eight hundred metres with summits and saddles in them. See rangeSpan in
core/world/relief.go. The valley is untouched by all of it - its heights hash
the same tile for tile - so [baseline.md](baseline.md) still stands.

Like the batch before it, this is a measurement of that change and of nothing
else: it was taken on the commit immediately previous, with nothing in
between.

```
before the mountains: gates: fed 0.68 safe 0.51 held 0.66 all 0.272 food 14.06 | lasted 6/8 extinct 1 mean 254.9 median 124
after the mountains:  gates: fed 0.69 safe 0.56 held 0.69 all 0.299 food 15.32 | lasted 8/8 extinct 0 mean 206.9 median 127
```

Nothing here is outside what eight seeds will do on their own, and it is not
claimed that the mountains fed anybody. `lasted` went 6/8 to 8/8 and the one
extinction went away, which is the right direction and is one settlement
either way; `gates all` moved 0.027, half of what the valley's table asks for
before a move is worth believing. The reading that matters is that a change
this large to the shape of the ground moved the settlements hardly at all -
they live on the lowland, and the lowland is the one part of the height field
this did not touch.

The batch before this one is kept for the same reason it always was:

```
before the rivers: gates: fed 0.70 safe 0.52 held 0.62 all 0.257 food 22.36 | lasted 6/8 extinct 0 mean 133.2 median 90
```

The cost of a day was measured on `a49cf5f` and has not been taken again: on
an idle machine of twenty-four cores a day here with forty people cost four to
six milliseconds, twenty of the hundred and twenty-eight chunks awake, against
half a millisecond for a day on the valley with forty. The whole batch above -
eight seeds, sixty years each - took 2m38s.

## The batch

```
seed  pop  died births houses fields |  phys  safe belng  estm  actl | order
   1  461   227    130    122    265 |  0.69  0.56  0.68  0.29  0.17 |  1.00
   2   34    63     77     17     26 |  0.70  0.63  0.45  0.35  0.02 |  1.00
   3  261   251    151    134    236 |  0.68  0.59  0.68  0.27  0.11 |  1.00
   4   64    43     60     20     68 |  0.68  0.46  0.63  0.36  0.02 |  1.00
   5   78   348    123     39     24 |  0.69  0.59  0.52  0.62  0.16 |  1.00
   6  127   221     68     67     87 |  0.84  0.56  0.69  0.18  0.04 |  1.00
   7  540   439    210    439     26 |  0.71  0.66  0.58  0.06  0.03 |  0.99
   8   90    44     37     19     85 |  0.66  0.50  0.72  0.33  0.02 |  1.00

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

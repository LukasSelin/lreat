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
one side or another. Three of the eight came to nothing and one grew past three hundred;
that is the globe's ecology as it stands,
before the settlements to come have bounded land of their own, and it is what
a change to the globe is held against. Everything that is measured on the valley is
measured on the valley still, and is unchanged by any of this.

Taken on the commit that put the slope-area law on its published form: a
channel is where A·S^theta is greatest, with theta one rather than a half,
which makes the reading A·S - the stream power index. See "What the last
change did" in [baseline.md](baseline.md), where the same change is set out
against the valley, and channelTheta in core/world/relief.go.

```
before: gates: fed 0.69 safe 0.42 held 0.62 all 0.234 food 7.06 | lasted 6/8 extinct 1 mean 191.9 median 209
after:  gates: fed 0.67 safe 0.36 held 0.59 all 0.196 food 12.99 | lasted 4/8 extinct 3 mean 75.6 median 46
```

This is worse again, and worse in the same direction as the batch before it.
Three settlements of eight died where one did; the mean population is a
hundred and sixteen down on a hundred and ninety-two. Over the three changes
on this branch the globe has gone 0.309, 0.234, 0.196 on the gate a birth has
to pass, and 0, 1, 3 on the extinctions. Three points is a trend or it is three
draws, and eight seeds cannot tell which.

What is known is that the ground and the dying do not line up. Measuring the
fertility and the flood-plain tiles within ten of each market, before this
change and after:

	seed 1      fertility -0.8%   flood plain  -4.3%   lived
	seed 2 died fertility  0.0%   flood plain   0.0%
	seed 3      fertility  0.0%   flood plain   0.0%   lived
	seed 4      fertility -5.4%   flood plain  -4.0%   lived
	seed 5 died fertility -14.3%  flood plain -15.4%
	seed 6      fertility  0.0%   flood plain   0.0%   lived
	seed 7 died fertility -1.9%   flood plain  -2.7%
	seed 8      fertility -7.2%   flood plain  -3.6%   lived, and largest

Seed 2 died on ground that did not move by a thousandth, and seed 7 on ground
that moved by two per cent. Seed 8 lost seven per cent of its fertility and
grew larger than anything else in the batch. Only seed 5 lost enough to argue
about. Whatever is killing these settlements, it is not mainly what happened to
the ground under them.

What did move is who the founders are. The world's own stream is drawn from for
the ground before it is drawn from for the people - carve takes a number from
it for the fish on every tile that becomes water - so any change to how many
tiles are river deals every founder a different hand. Three tests in
core/action fell over on exactly this at the same time, and one of them turned
out to have been passing on a coin toss for as long as it had existed. A globe
batch is eight seeds, and eight settlements founded by eight differently drawn
sets of people is an eight-sample lottery on top of whatever the terrain did.

So a bigger batch is what settles it, and a bigger batch is what should be run
before anyone concludes the rivers are starving the globe. Until then this is
recorded as what it is: a worse run, of a size that cannot say why.

## Before that

Taken on the commit that put the rivers where water would actually put them.
Two readings were wrong: `Grid.Aspect` took the lowest neighbour rather than
the steepest fall, so on evenly falling ground the water always left
cornerways and the plains were ruled with parallel lines at forty-five
degrees; and a river was picked by how much water crossed it alone, so the
heads of them were never in the hills. Rain is heavier on high ground now and
none falls on the sea, a channel is picked by the water against the fall, and
each is laid from its head the whole way to the sea so a trunk on its own flood
plain is still a river. See "What the last change did" in
[baseline.md](baseline.md), where the same change is set out against the
valley.

```
before the rivers: gates: fed 0.69 safe 0.56 held 0.73 all 0.309 food 8.98 | lasted 7/8 extinct 0 mean 163.6 median 159
after the rivers:  gates: fed 0.69 safe 0.42 held 0.62 all 0.234 food 7.06 | lasted 6/8 extinct 1 mean 191.9 median 209
```

This is a worse run and it should be said plainly rather than absorbed. `safe`
fell 0.14 and `gates all` 0.075, both well past what the valley's table calls
noise, and one settlement died where none had. Seed 1 came out with nobody at
all and seed 7 with two souls; the other six are healthy and three of them are
the best this preset has produced.

The obvious explanation is wrong. It would be that the new drainage left those
two markets dry, and it did not: both have water one tile from the square, both
sit wholly on flood plain, and their ground is the most fertile of the eight -
0.97 and 0.90 against a 0.72 for the seed that grew to five hundred. Whatever
killed them, it was not the water or the soil.

So this is recorded and not explained. Eight seeds can say that something moved
and cannot say what, and the next thing worth doing is to run the two dead
seeds and watch them rather than to guess again and put another unmeasured
change on top of this one.

## Before that

Taken on the commit that made a globe out of its own history rather than
drawing it. The crust breaks into pieces that drift, weld when two continents
have driven into each other for long enough, rift when one grows too big to
have anything happening inside it, and are taken into their neighbours when
they are ground down too small to be plates; the boundaries between them are
bent off the straight line they used to run in. A globe breaks into sixty-four
pieces and ends with about thirty. See the block above rangeSpan in
core/world/relief.go for the mountains and the one above plateWarp in
core/world/history.go for the crust.

It costs fourteen seconds a world against one, and the batch is 3m41s against
2m55s.

```
drawn:  gates: fed 0.69 safe 0.56 held 0.69 all 0.299 food 15.32 | lasted 8/8 extinct 0 mean 206.9 median 127
run:    gates: fed 0.62 safe 0.52 held 0.73 all 0.309 food 8.98 | lasted 7/8 extinct 0 mean 163.6 median 159
```

The gate a birth has to pass is where it was - 0.299 to 0.309 is nothing - and
so is the tally: nobody more died out and one settlement fewer held its
founding size, which is one settlement. What is not nothing is `food`, which
has fallen by two fifths, and `fed` with it by 0.07. A made globe feeds a
settlement worse than a drawn one, and by enough to say so.

The reason is not mysterious and is not settled either. A made world puts its
high ground at the boundaries of its plates, and the boundary of a continental
plate is its coast, so a founding party set down on a coast - which is where
the globe puts them, see above - is set down under a mountain far more often
than on a drawn map, with less wild ground within reach of it. Whether the
answer is to found them elsewhere, or to give plate interiors relief so that
the coast stops being the only high ground, is the next thing to find out. It
is recorded here rather than fixed because eight seeds of a batch this size
can say that something moved and cannot say what.

## Before that

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
   1   15    42     37      7     19 |  0.68  0.58  0.49  0.48  0.00 |  1.00
   2    0    20      0      0      0 |  0.00  0.00  0.00  0.00  0.00 |  0.00
   3  112   146     84     36    110 |  0.81  0.50  0.65  0.31  0.12 |  0.99
   4   46   192    113     25     26 |  0.64  0.53  0.69  0.24  0.05 |  0.99
   5    0    39     19      0      0 |  0.00  0.00  0.00  0.00  0.00 |  0.00
   6   56    69     86     10     23 |  0.77  0.46  0.77  0.20  0.08 |  1.00
   7    0    21      1      0      0 |  0.00  0.00  0.00  0.00  0.00 |  0.00
   8  376   291    161    180    181 |  0.71  0.55  0.66  0.15  0.01 |  1.00

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

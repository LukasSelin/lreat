# Baseline on the globe

What `tune` prints on the globe, checked in so that nobody has to run it to
find out. The run below is the globe batch:

```bash
go run ./cmd/tune -preset globe -seeds 8 -ticks 21600
```

It is eight seeds rather than twenty-four because a globe is a hundred and
eighty times the ground of the valley and each seed is a single thread here,
so the batch is half an hour where the valley's is a few minutes.

Eight is enough for the gates and it is not enough for the dying. Twice on this
branch an eight-seed batch has said something about `extinct` and `lasted` that
a twenty-four-seed batch then contradicted - once inventing a decline the
rivers had not caused, once matching a real one by luck. A move in those two
readings is worth nothing at this size; take twenty-four before believing one,
which is twelve minutes and not half an hour now that a globe is made in
fourteen seconds. It is taken
on the same terms otherwise: sixty years, twenty founders, one settlement
founded where the terrain scorer puts it.

This is not the valley's baseline and none of its numbers compare to
[baseline.md](baseline.md). The world is a cylinder a thousand tiles round,
a third of it sea, cold at the poles and warm at the middle, and a founding
party is set down on a coast where a laden walker is cut off by water on
one side or another. Three of the eight came to nothing and one grew past five hundred;
that is the globe's ecology as it stands,
before the settlements to come have bounded land of their own, and it is what
a change to the globe is held against. Everything that is measured on the valley is
measured on the valley still, and is unchanged by any of this.

Taken on the commit that stopped a history paving itself in basalt.

Every scrap of basalt on a made world came from melt at a rifting axis, and
melt was kept as a running total. A rift axis drifts, so a tile that was ever
near one accrued a sum nothing else could outweigh and could never lose again -
and the longer a history ran the more of its continent turned to ocean floor. A
made valley came out fifteen parts in a hundred basalt after one epoch and
fifty-four after sixteen, with its granite falling from forty-nine to five,
which is the opposite of what running a history longer should do. Melt is the
deepest single flooring now and not the sum of all of them, because flooring
ground twice leaves it floored and not floored twice as much; and nothing
remakes a tile until fifteen metres of ground have moved, where the fill has
always had such a bar and the fire and the crushing had none. On a made valley
the basalt falls from fifty-four in a hundred to twenty-four and no rock owns
the map. See madeEnough in core/world/history.go.

The globe was never the case that was wrong - its plates are large enough that
the seams do not sweep it - and its basalt only goes from ten parts in a
hundred to five. What moved here is the settlements, and not by way of the
ground: the land's slope reads identically to three figures before and after,
median 0.017 and ninetieth 0.657, on rock four per cent softer. It is the soil,
which is read off the bedrock, and the founders, who are dealt again whenever
anything at all about the water changes.

	                     extinct  lasted  gates all  held  median
	before                 7/24   16/24     0.189    0.55    60
	after                  7/24   13/24     0.165    0.51    40

Both moves are inside what this batch is allowed to see - the table in
baseline.md asks for 0.05 on the gate and five settlements of twenty-four - and
both go the same way as every change on this branch has. That is worth setting
down on its own. Across the five commits the globe has gone 0.214, 0.219,
0.202, 0.189, 0.165 on the gate a birth has to pass, each step inside the noise
and the whole of it, 0.049, sitting exactly on the line where a move starts
being worth believing. Nothing here says which step did it, or whether five
reshuffles of who the founders are would have done as much on their own. A
batch that could tell would be the same twenty-four seeds run on two trees that
differ only in the world's luck, and it has not been run.

## Before that

Taken on the commit that gave the mountain flanks their spurs and hollows.

A belt of country raised smoothly is an inclined plane, and an inclined plane
drains in parallel lines - every tile on it sends its water the same way, so
the rills run side by side down the flank and no two of them ever meet. The
mountainsides came out combed: even, parallel streams at a tile's spacing
running straight into the sea. Ground steeper than 0.3 had three or more
neighbours draining into it on four tiles in a thousand, against twenty-seven
in a thousand on gentle ground; the flanks had no hollows for water to gather
in. With the spurs, the streams on steep ground join at twenty-four in a
hundred where they joined at eight. See beltGrain in core/world/history.go.

How rough to make them was settled by this batch and not by the look of it,
because rough flanks are ground nobody can build on or plough. Twenty-four
seeds:

	flank roughness   extinct  lasted  gates all  held  median
	none                6/24    17/24    0.202    0.62    127
	a tenth             7/24    16/24    0.189    0.55     60
	a quarter           7/24    11/24    0.145    0.48     17

A quarter is the better-looking map and it takes six settlements off the batch.
A tenth keeps nearly all of the joining - twenty-four in a hundred against
twenty-seven - and comes back to within a settlement and a hundredth and a half
of where the globe stood, which is inside what this batch can see.

The eight-seed batch on the quarter reported nought extinct and looked healthy.
It took twenty-four seeds to find the eleven. That is the second time the note
above has earned itself in as many changes.

## Before that

Taken on the commit that tightened how far a great river's flood reaches: onto
the ground no higher than the water, where it used to take every neighbour
standing up to a metre above the channel. See "What the last change did" in
[baseline.md](baseline.md).

This is the other half of the bank correction, and it gives back most of what
that correction cost. Twenty-four seeds, one variable at a time, which is the
size those two readings need:

	tree                                    extinct  lasted  gates all  safe  median
	rivers, before the slope-area law         3/24   20/24     0.219    0.41    124
	slope-area law, flood rising a metre      9/24   13/24     0.185    0.32     47
	slope-area law, flood no higher           6/24   17/24     0.202    0.39    127

The middling settlement is back where it was before any of it - a hundred and
twenty-seven against a hundred and twenty-four - and `safe` within two
hundredths. Three of the six extra deaths come back and three do not, which is
the standing price of theta at one, recorded as such and deliberately paid: the
mountains are worth it on a preset whose settlements are not yet what is being
tuned.

Three of the eight above came to nothing and one grew past nine hundred, which
is the widest spread this preset has shown. On eight seeds that is two
extinctions; on twenty-four it is six. Take the twenty-four before believing
either.

## Before that

Taken on the commit that put the slope-area law on its published form: a
channel is where A·S^theta is greatest, with theta one rather than a half,
which makes the reading A·S - the stream power index. See "What the last
change did" in [baseline.md](baseline.md), where the same change is set out
against the valley, and channelTheta in core/world/relief.go.

```
before: gates: fed 0.69 safe 0.42 held 0.62 all 0.234 food 7.06 | lasted 6/8 extinct 1 mean 191.9 median 209
after:  gates: fed 0.67 safe 0.36 held 0.59 all 0.196 food 12.99 | lasted 4/8 extinct 3 mean 75.6 median 46
```

This is worse, and this time it was run down rather than guessed at. The batch
above is eight seeds and eight seeds could not say why, so the same batch was
taken at twenty-four on this commit and on the two before it, and once more on
this commit with the exponent put back. Four runs, one variable at a time:

	tree                                     extinct  lasted  gates all  median
	plates                                     4/24   19/24     0.214      90
	rivers                                     3/24   20/24     0.219     124
	this commit, theta 0.5, banks off flow     3/24   16/24     0.189      61
	this commit, as it stands                  9/24   13/24     0.185      47

Read down the column and the two halves of this commit did two different
things. Raising theta from a half to one is what triples the dying - three in
twenty-four to nine, with the bank reading held still. Reading the banks off
the flow instead of off the work is what costs the gate a birth has to pass -
0.219 to 0.189 - and halves the middling settlement, while leaving the dying
where it was.

That second one is the uncomfortable half, because it is a correctness fix. The
version that scored better was the version in which nine tenths of a globe's
bank-flooding happened on mountainsides while the flood plains stayed dry. The
better number was being bought with a wrong map.

Both are kept. The mountains are worth a third of the settlements on a preset
whose settlements are not yet what is being tuned, and a map that floods its
ridges is not worth keeping for a better score.

What the bank reading costs was then run down, and it is not food. Counted
within ten tiles of the market, over the eight globes, mean per world:

	                          water  fish  fertility  buildable
	banks off the work (bug)  124.6  106.5    247.1      272.0
	banks off the flow (fix)  149.5  127.3    232.2      240.4

The fix moves the flooding off the ridges and onto the flood plains, and the
flood plains are where the markets are - so a settlement gains twenty-five
tiles of water and twenty-one of fish, and loses thirty-two tiles it could have
built on, an eighth of its building ground. That is the whole of the cost, and
it shows in the readings: `food` went up while `held` and `safe` went down. It
is a housing problem and not a hunger one.

Which points at the spreading rule rather than at the reading. A great river
takes every neighbour standing no more than a metre above its channel, and a
metre is a great deal of flood plain when the ground is flat. That figure was
settled when this fired on ridges and hardly ever on a flood plain; it is doing
far more work now that it fires where it should. Tightening it is the next
thing to try, and it wants its own batch.

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

This read as a worse run and it was not one. Taken at eight seeds it showed
`safe` down 0.14, `gates all` down 0.075 and one settlement dead where none had
been, and it was written up here as a real decline. Taken again at twenty-four
seeds afterwards, against the commit before it at the same size, the rivers
moved nothing: four dead becoming three, and 0.214 on the gate becoming 0.219.
The decline was the batch being too small to see through, and the write-up
below is left standing as a caution about that rather than corrected away.

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
   1  156    33     81     20     96 |  0.68  0.49  0.72  0.66  0.09 |  1.00
   2    0    20      0      0      0 |  0.00  0.00  0.00  0.00  0.00 |  0.00
   3  546   394    114    110    349 |  0.66  0.49  0.57  0.30  0.04 |  1.00
   4    0    20      0      0      0 |  0.00  0.00  0.00  0.00  0.00 |  0.00
   5   40   224    174     21     32 |  0.84  0.65  0.80  0.26  0.08 |  1.00
   6   24    77     81     11     19 |  0.69  0.54  0.74  0.09  0.00 |  0.98
   7  100    57     58     22     38 |  0.71  0.48  0.49  0.60  0.04 |  1.00
   8    0    21      1      0      0 |  0.00  0.00  0.00  0.00  0.00 |  0.00

dwell/rest                          620911  27.9%
take/berries@wood                   457267  20.5%
dwell/guard@market                  380549  17.1%
consume/provision                   339575  15.3%
dwell/meet@tavern>neighbour          84491   3.8%
transfer/material>requester          44467   2.0%
take/timber@wood                     41220   1.9%
take/grain@field                     40517   1.8%
take/fish@water                      38049   1.7%
transfer/provision>needy             33723   1.5%
exchange/coin>provision@market       26840   1.2%
tend/plant@open                      22681   1.0%
exchange/material>coin@market        21840   1.0%
pass/practice>self                   16520   0.7%
pass/practice>pupil                  13578   0.6%
take/game@wood                       13220   0.6%
dwell/look                           12695   0.6%
make/timber>tool@bench                8171   0.4%
raise/timber>dwelling@open            5171   0.2%
transfer/provision<holder             1606   0.1%
tend/clear@open                       1338   0.1%
strike/person>wrongdoer               1301   0.1%
raise/timber>road@ground               540   0.0%
tend/water@field                       308   0.0%
take/stone@outcrop                      63   0.0%
raise/timber>tavern@open                23   0.0%
move@dwelling                           14   0.0%
make/provision+timber>meal@hearth       10   0.0%
make/stone+timber>tool@forge             9   0.0%
raise/timber+stone>granary@open          2   0.0%
raise/timber+stone>market@open           2   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.58 safe 0.31 held 0.49 all 0.171 food 9.33 hungry-with-food 0.26 | lasted 5/8 extinct 3 mean 108.2 median 40 | phys 0.72 safe 0.53 belng 0.66 estm 0.38
```

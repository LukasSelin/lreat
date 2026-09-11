# Baseline

What `tune` prints on master, checked in so that nobody has to run it to find
out. The run below is the default batch:

```bash
go run ./cmd/tune -seeds 24 -ticks 21600
```

Nothing in it is sampled from the wall clock, so the same commit gives the same
numbers however the goroutines interleave. That is what makes a checked-in
baseline worth anything: the file is the run, and re-running it only confirms
it.

Read it before a change rather than measuring master again. Two hours of work
begins with a few minutes of a batch that already has an answer here, and the
answer does not change while the work is going on. Take the comparison run once
the change is settled instead, and when it lands, replace what is below with the
new numbers in the same commit - the file describes the tree it is committed in,
so a change to behaviour that leaves it alone has left it wrong.

The headline is the last line: how many of the 24 settlements were still at
least their founding size at the end, how many died out, and where the mean
population landed. The `gates` line in front of it is the share of fertile
agents fed, safe and held - `all` is the share that were all three at once,
which is the gate a birth has to pass, and the one worth watching.

## Why the batch is 21,600 ticks and not 6000

It is the same sixty years it always was. A tick used to be a hundredth of a
year and is now a day, so the number in the command had to change for the run
to go on meaning the same thing; see "The calendar" in docs/action-space.md.
Anybody holding these numbers against a batch measured before that should
compare years and not ticks - and should know that sixty years is not quite the
same stretch of a life either way, because a life was 52 years then and is 65
now. A settlement is read a little earlier in its generations here than it used
to be.

## master

```
seed  pop  died births houses fields |  phys  safe belng  estm  actl | order
   1   53   137     34     18     15 |  0.67  0.59  0.71  0.32  0.12 |  1.00
   2  185   351     35     90     49 |  0.58  0.51  0.36  0.12  0.01 |  1.00
   3  110   116     16     25     43 |  0.62  0.48  0.45  0.21  0.05 |  1.00
   4   58   246     23     24     22 |  0.47  0.51  0.56  0.17  0.05 |  1.00
   5  415   377     31    151     82 |  0.63  0.51  0.60  0.16  0.06 |  1.00
   6  237   209     32     88    112 |  0.60  0.54  0.49  0.15  0.06 |  1.00
   7  182   261     26     35     91 |  0.60  0.56  0.53  0.29  0.11 |  1.00
   8  231   435     43     87    206 |  0.59  0.56  0.46  0.26  0.04 |  1.00
   9  121   341     25     60     22 |  0.68  0.47  0.49  0.05  0.01 |  0.64
  10  165   275     13     50     78 |  0.61  0.52  0.34  0.16  0.07 |  1.00
  11   71   244     26     24     35 |  0.58  0.55  0.37  0.20  0.02 |  1.00
  12   93   196     35     30     35 |  0.65  0.54  0.69  0.13  0.09 |  1.00
  13  375   291     26    129    122 |  0.68  0.49  0.44  0.12  0.05 |  1.00
  14  605   292     31     98    214 |  0.67  0.53  0.51  0.29  0.10 |  1.00
  15   35   260     76     33     44 |  0.59  0.58  0.62  0.08  0.10 |  0.82
  16   64   123     31     19     32 |  0.62  0.55  0.68  0.10  0.06 |  1.00
  17  150   279     44     74    152 |  0.62  0.55  0.54  0.17  0.03 |  1.00
  18  343   176     22    124    207 |  0.68  0.50  0.44  0.15  0.06 |  1.00
  19  249   269     31     78     39 |  0.65  0.57  0.58  0.34  0.10 |  1.00
  20   30   146     33     25     66 |  0.57  0.49  0.45  0.03  0.00 |  0.74
  21  148   345     50     99    113 |  0.53  0.61  0.37  0.11  0.02 |  1.00
  22   99   195     17     43     32 |  0.65  0.53  0.64  0.25  0.04 |  1.00
  23  215   299     29     70     80 |  0.55  0.48  0.31  0.12  0.10 |  1.00
  24   84   193     30     68     79 |  0.67  0.69  0.28  0.09  0.02 |  1.00

dwell/rest                         3844063  30.0%
take/berries@wood                  3083932  24.1%
consume/provision                  2194965  17.1%
dwell/guard@market                 1307284  10.2%
take/fish@water                     533611   4.2%
dwell/meet@tavern>neighbour         500825   3.9%
take/timber@wood                    244903   1.9%
exchange/coin>provision@market      170050   1.3%
transfer/material>requester         134751   1.1%
take/grain@field                    113941   0.9%
pass/practice>self                  111944   0.9%
tend/plant@open                     107648   0.8%
pass/practice>pupil                 105403   0.8%
exchange/material>coin@market        99396   0.8%
transfer/provision>needy             78821   0.6%
take/game@wood                       48825   0.4%
raise/timber>dwelling@open           46728   0.4%
make/timber>tool@bench               41675   0.3%
take/stone@outcrop                   12806   0.1%
strike/person>wrongdoer              11099   0.1%
transfer/provision<holder             9926   0.1%
tend/clear@open                       4957   0.0%
raise/timber>road@ground              2500   0.0%
tend/water@field                      2253   0.0%
raise/timber+stone>granary@open        989   0.0%
make/provision+timber>meal@hearth      478   0.0%
make/stone+timber>tool@forge           340   0.0%
dwell/look                             314   0.0%
raise/timber>tavern@open               188   0.0%
move@dwelling                           83   0.0%
raise/timber+stone>market@open          62   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.59 safe 0.51 held 0.69 all 0.262 food 3.73 hungry-with-food 0.26 | lasted 24/24 extinct 0 mean 179.9 median 150 | phys 0.61 safe 0.54 belng 0.50 estm 0.17
```

## What the last change did

The creatures made to last. Over sixty years on the valley the herds followed
the woods down and were gone from a settlement of a hundred people while the
woods still stood, boar starved walking forty tiles to strips cut before they
arrived, and deer and boar took turns eating each other out of the one brush.
A feeding now takes half what a forage takes and a thin stand still feeds; a
raid is local; a boar lives on the mast under the old trees, a count kept on
the timber that nothing else draws on, and on the sward besides, and a kind is
released as a herd rather than scattered. Every kind was still there at sixty
years on two valleys, one of them at a hundred and fourteen people. None of
it is read by a run with no creatures in it, so the golden numbers and the
batch stand.

### What the creatures do to the country

What the creatures do to the country: a browsing sets a young stand back, a
wood's seed takes only among shoots so a grazed meadow stays open, a grazing
manures the ground, and a boar turns the soil and now and then plants a wood
beside it. None of it is read by a run with no creatures in it - the one
line in the settlement's own passes, the seed asking the sward before any
chance is spent on it, never fires on ground nobody grazes - so the golden
numbers and the batch stand where the merge below left them.

### Boar and hares

Boar and hares, after the deer. A boar is a deer that lets a person come
nearer and raids the fields; a hare grazes the sward, which is new: the grass
standing on open ground, a stock beside the map like the fish, drawn down by
grazing and put back by the growing weather. **The sward moved the three
golden numbers and nothing else.** No run a settlement is measured on has a
creature in it, so the sward stands at one on every open tile from the first
day to the last and nothing a settlement does reads it; the numbers moved
because there is one more value in the line being hashed, and it was proved
the way the plate and the epoch were, by running the digest without the sward
on the tree with it and getting the three numbers before. The batch is not
retaken for it: what a batch measures does not read the sward. The numbers
were taken on the tree with the landmarks and the walking limit merged in,
whose own numbers the digest without the sward reproduced on that tree.

## Earlier changes

### What the landmarks and the walking limit did, merged

The landmarks and the walking limit, merged with master's islands,
mountains, the flood's level and the deer. The landmarks bound the walk a
guided search has left off a few tables (`world.Landmarks`), so it opens
fewer tiles on its way to the same destination at the same cost; the
walking limit lets a scout give up on a way round that is more than three
times the distance meant, as soon as the landmarks say so; and the islands
act on shallow copies of the map, so each island's view now keeps its own
copy of the landmarks' slack and puts what it built into the tables'
account afterwards. The accounts of the two on their own branch are below;
this is the merged tree against master.

**On the batch it is a draw.**

```
master:  gates: fed 0.58 safe 0.50 held 0.69 all 0.251 food 3.61 | lasted 24/24 mean 187.5 median 149
merged:  gates: fed 0.59 safe 0.51 held 0.69 all 0.262 food 3.73 | lasted 24/24 mean 179.9 median 150
```

Nothing is past a threshold. The golden numbers were retaken on the merged
tree, both sides having moved every seed on their own; the deer moved none,
as their own account says, and the batch came out to the same line before
and after the deer were kept to the bank. The one test that moved was the
herd's: whether four fifths of a herd stand under trees after six years
swings between two thirds and nearly all on any one seed - master's own
seed 4 sits under the bar - with nothing changed but which of two ways of
the same cost somebody walked, so it now reads the share over four valleys.

### What the deer did

Deer. A second kind of agent lives in the woods around the settlement now, on
the same psychology the people run on - the same needs, the same reading of a
moment as a point in habit space, the same recognition of an act, the same
walking - and differs from them by its species: how long it lives and when it
bears, the body and mind it hands its young, what it wants at birth and how
fast each want returns, and what it finds on the coordinates a person reads.
See core/entity/species.go, core/action/deer.go, and the README.

**It changes nothing a settlement is measured on.** No run the batch takes has
a deer in it, and the whole of the work was written so that a world with none
draws exactly the chance it always drew: every act a person does keeps its
slot in the catalog, because the acts are ordered by their actor before their
key and deer come after people; a person never seeds, drifts or weighs a
deer's slot, because seeding draws luck for every coordinate of every slot it
seeds; a person's neighbour, witness and stranger are asked for among people;
and the settlement's books - its population, its means, its funnel, the market's
middle, what its scholars count - are of the people. The batch above was taken
again on the tree with the deer in it and came out the same to the digit, and
the three golden seeds did not move.

What a herd does when it is let in is not measured here and there is no
baseline for it yet. On one seed, ten founders and twenty deer over twenty
years, the herd grew to about sixty-five and the brush in the wood fell from
three hundred units to a hundred and twenty, which is the herd finding the
wood's limit: a browse takes what a forage takes, so a deer and a forager are
in plain competition for the same brush. Three things had to be found out on
the way, each by watching a herd do something absurd. A bed must restore
next to nothing, or a deer never eats: a person's rest gives a little more
than a day burns, which a person walks and works off, and a herd of five
hundred lay in the wood with full bellies and the brush untouched. Fear must
fall on the safety tier at once, not at the tenth a day a roof does, or the
one coordinate that carries the alarm never out-fits bedding down. And a deer
standing in its herd must be fed belonging by standing there, or the tier
drains to nothing and no fawn is ever born.

### What the walking limit did

A scout aims fourteen tiles out as the crow flies, and where the water lay
between, the way round could be many times that or run out of the search's
window altogether - and the search opened everything it could reach to
find that out, for every scout on that shore, every day. `action.Def.Walk`
now bounds how far a plan may set out on, and only scouting sets it, at
three times its range; `Router.Within` carries it into the search, which
stops as soon as the landmark bound says the rest is too far. A scout
whose way round is longer than that has no way, where before it had a
long one, so more plans come back with no way and fewer people walk
round lakes.

**On the batch it is a draw.**

```
master:    gates: fed 0.59 safe 0.51 held 0.69 all 0.256 food 4.02 | lasted 23/24 mean 152.5 median 139
landmarks: gates: fed 0.58 safe 0.53 held 0.70 all 0.259 food 4.22 | lasted 24/24 mean 177.7 median 168
walk:      gates: fed 0.58 safe 0.52 held 0.69 all 0.252 food 4.03 | lasted 24/24 mean 158.8 median 117
```

Nothing is past a threshold against either the landmarks batch it follows
or master: `fed`, `safe` and `held` are within 0.01 of both, `all` within
0.007, one more settlement lasted than on master, and the population is
back to within six of master's after the landmarks batch had it up 25 -
which is the swing three batches of master show with no change at all.
`food` went up 0.20 with the landmarks and came back down 0.19 here, so
neither move was the code's doing.

Only golden seed 9 moved: the other two send nobody that far round within
fifteen hundred days.

### What the landmarks did

A guided search now bounds the walk that is left off landmark tables as well
as the straight line (`world.Landmarks`, core/world/landmark.go), so it
opens fewer tiles on its way to the same destination: 2.2 times fewer on
the routes settled agents were walking on the valley, 1.4 times fewer on
the scattered globe. The bound is never above the true cost, so every
destination is settled at the same cost and by the same first step as
before. What can differ is which of several routes of exactly the same cost
is the one settled on, and the one walked is where the wear falls and the
roads come, so the runs do diverge from master tick by tick. This batch is
what that divergence amounts to over sixty years.

**On the batch it is a draw.**

```
master:    gates: fed 0.59 safe 0.51 held 0.69 all 0.256 food 4.02 | lasted 23/24 mean 152.5 median 139
landmarks: gates: fed 0.58 safe 0.53 held 0.70 all 0.259 food 4.22 | lasted 24/24 mean 177.7 median 168
```

Nothing is past a threshold. `fed` is down 0.01, `safe` up 0.02, `held` up
0.01 and `all` up 0.003; one more settlement lasted, which is one of the
five it would take to mean anything; and the population is up 25, which is
inside the swing three batches of master showed with no change at all (177,
174, 190 - see "How much of that is chance"). `food` is up 0.20, the one
reading that is more than noise-sized, but on a single batch it is not
evidence either way. The change was made for the machine and not for the
people, and the batch says the people did not notice.

The golden numbers moved, for the reason above: a run walks a different
route of the same cost within its first fifteen hundred days.

### What the flood was let rise to

A great river spreads onto the ground beside it that is no higher than the
water. It used to take every neighbour standing up to a metre above its
channel, while the line of comment above it said "no higher" - the comment and
the code had disagreed since the day it was written, and the code was the
generous one.

It went unnoticed for as long as the reading that picks a great river was
wrong. A river picked by how hard it is cutting is a rill near a ridge, and a
rill near a ridge has no flat ground beside it to give away; corrected to read
off the flow, the rule started firing on flood plains, which is where the
markets are, and a metre of flood took an eighth of a settlement's building
ground with it.

Counted within ten tiles of a market over eight globes, by how far the flood is
let rise:

```
rise    water   fish   fertility   buildable   river, share of map
1.00    149.5   127.3    232.2       240.4           7.18%
0.50    146.0   124.6    237.8       244.1           7.05%
0.25    142.5   121.2    242.1       247.9           6.85%
0.00    130.1   111.2    254.2       256.4           5.68%
```

Nothing gives back half the ground the correction cost, against 272 when the
rule was firing on ridges, and still leaves a settlement more water and more
fish than it had then. It also brings the share of a map that comes out as
watercourse back toward the waterShare it asks for - the banks are laid after
the channels are counted, so whatever they add is over the top of it, and at a
metre they were adding two thirds again.

```
before: gates: fed 0.58 safe 0.51 held 0.69 all 0.247 food 3.73 | lasted 23/24 extinct 0 mean 134.4 median 97
after:  gates: fed 0.58 safe 0.50 held 0.69 all 0.251 food 3.61 | lasted 24/24 extinct 0 mean 187.5 median 149
```

Every settlement held its founding size, which has not happened before on this
batch, and the gate is back within a hundredth of where master had it before
any of this branch touched the water. The mean and the median are up by half,
which the table above says to believe never on its own.

`TestSoilGoesWithTheGround` had to be re-aimed and it is worth saying why,
because the behaviour it guards did not change. It asked for the ground that
was still a hillside at the end, and asked it by standing more than FloodDepth
above the nearest water. Drain is measured to the nearest water, so what counts
as off the flood plain moves whenever the amount of water on the map moves:
with the banks tightened, ground the river still feeds stopped clearing the
line and stayed in the reckoning, and the whole set came out richer. Every part
of it that is really a hillside came out poorer - over a tenth of a fall, 0.150
to 0.054; over a fifth, 0.150 to 0.027 - so it asks the fall now. The fall is
what a slope is.

And the two tests fixed in the commit before this one had to be fixed again,
which is the more useful lesson. They were re-aimed to ask their question of
twenty worlds instead of one, and the bar was set at what those twenty worlds
happened to give: seventeen. The true rate is sixty-eight in a hundred, so a
bar at seventeen in twenty is above the rate and fails by construction on the
next change to the ground - which is exactly what it did, one commit later. A
bar set at what was measured is a bar that fails half the time. They ask sixty
worlds now and the bars sit at about half.


What the changes before this one did, each measured against the master of its
own day. They are kept for the method rather than for the numbers: none of
them is a comparison with the run above.

### What the slope-area law did

Which tiles are river is read by the slope-area law now, properly and on a
published exponent. A channel is where A·S^theta is greatest - the water times
the fall, raised to a power - and theta is one, which makes the reading A·S,
the stream power index, and is the bottom of the range the channel-initiation
literature quotes.

It was a half before, and the half sat in a constant that nothing referenced:
`carve` hardcoded a square root and `channelFall = 0.5` stood beside it saying
what the square root was for, so the two could have drifted apart without a
word. The exponent is swept and the table is in the constant's comment. The
high fifth of a valley holds sixteen river tiles in a hundred at one where it
held eight at a half - and none at all when a river was picked by the water
alone.

Incision was left as it was, and that is a decision rather than an oversight.
Its comment claimed to take the same reading erode.go takes, and erode.go takes
the stream power law entire, with the fall in it. Giving incision the fall
moves the mean cut on the low half of a valley from 2.49 metres to 0.60 and
puts it on the top fifth instead; the valley's own trunk - the tile where the
river leaves the map, whose fall is exactly zero - goes from 26.7 metres of cut
to 0.03, and the valley the settlement lives in stops existing. The difference
is that erosion runs an age at a time and incision runs once: stream power says
how fast a channel is cutting now, and a channel on its own flood plain is
cutting nothing now and still lies at the bottom of a valley, because it spent
ages getting there. The comment says so now.

One bug came out of the measuring. Whether a river is great enough to spread
onto its banks had been read off the same cutting reading, and it should not
be: how hard the water is working and how much of it there is point opposite
ways, and the hardest-working tile on a map is a steep rill near a ridge
carrying nothing. At theta of one, nine tenths of the bank-flooding on a globe
was happening in the top fifth of the ground - mountainsides in flood with the
flood plains dry. Read off the flow again, it is a seventh, and the share of
river tiles draining onto dry ground fell from between two and seven in a
hundred to between one and three in a thousand.

```
before: gates: fed 0.57 safe 0.50 held 0.69 all 0.245 food 5.28 | lasted 22/24 extinct 0 mean 155.2 median 147
after:  gates: fed 0.58 safe 0.51 held 0.69 all 0.247 food 3.73 | lasted 23/24 extinct 0 mean 134.4 median 97
```

The valley did not move. Every gate is inside a hundredth, `lasted` went the
right way by one, and `food` came back to about where master had it before the
rivers were touched at all - which is the more likely reading of the third it
had gained than that the rivers gave it and the channels took it away. The
mean population fell by a fifth and the median by a third, which the table
above says to believe on its own never.

Three tests in core/action had to be fixed, and they are worth a word because
they had nothing to do with the ground. They draw their founders from the
world's own stream, and the ground is drawn from that stream first, so any
change to the terrain deals every founder a different hand.
`TestFoundersDifferFromOneAnother` asked that one founder's habits stay within
a cosine of 0.9 of the prior; over two hundred founders that runs from 0.66 to
0.96 with a median of 0.89, so it was a coin toss, and it cleared the bar on
eight of twenty world seeds. It had been coming up heads. All three ask their
question of a crowd now - two hundred founders, or twenty worlds - so the next
change to the ground will not trip them.

### What the rivers did

Two readings of where water goes were wrong, and both moved every river on the
map.

`Grid.Aspect` took the lowest neighbour rather than the steepest fall. A
diagonal neighbour is half again as far off, so on ground that falls evenly it
sits lower and won - and the same diagonal won every time, because the first
direction offered takes a tie. Two river tiles in three left their tile
cornerways, and what that drew was a set of parallel lines at forty-five
degrees ruled across every plain on the map. It is a quarter of them now.
`Grid.Slope`, three functions above it, had been dividing the drop by the
distance all along; `Aspect` was the one reading of the same eight neighbours
that did not.

And a river was picked out by how much water crossed it and nothing else, so
the head of one sat wherever a catchment had gathered enough - which on flat
country is nowhere in particular and never in the hills. Three things answer
that. Rain is heavier on high ground and none of it falls on the sea. A channel
is picked by the water against the fall rather than by the water alone, because
steep ground needs less of a catchment to cut one. And a channel is then laid
from its head the whole way down to the sea, so that a trunk crossing its own
flood plain - all the water on the map and no fall at all - is still a river.

The high fifth of a valley held none of its river tiles, one, and none over
three seeds. It holds fifteen, twenty-nine and fifteen.

```
before: gates: fed 0.59 safe 0.51 held 0.69 all 0.256 food 4.02 | lasted 23/24 extinct 0 mean 152.5 median 139
after:  gates: fed 0.57 safe 0.50 held 0.69 all 0.245 food 5.28 | lasted 22/24 extinct 0 mean 155.2 median 147
```

The gates are where they were - 0.256 to 0.245 is half of what the table above
asks for before a move is worth believing - and so is the tally, at one
settlement either way. The one figure outside noise is `food`, up by a third.
A settlement lives on its flood plain and moving the rivers moved the flood
plains; whether a third more food on hand is that or the draw wants a batch
that could carry the question, and twenty-four seeds cannot.

The golden numbers moved, and had to: the rivers are in different places, so
the soil is, so the fields are.

### What the unreachable pressures did

Irrigation and forestry were unreachable, and had been since they were
written. The pressure spike made it visible - a zero standing in `w.Pressed`
where the other four were accumulating - but it was not the spike's doing.
Both conditions asked the world for something it does not do.

They were wrong in two different ways, and only one of them was a number.

### Forestry was reading the wrong ground

`forestGone` asked whether the map had lost two fifths of its forest. One
settlement cannot do that. Over sixty years the whole-map share bottomed out
at 1.00, 0.68, 0.61 and 0.98 on four probed seeds, against a bar of 0.60 -
and 1.00 means a settlement that cleared nothing net at all, because regrowth
kept pace with it everywhere but where it actually was.

Where it actually was, it cleared plenty: the wood share of the ground around
the market fell to between 0.03 and 0.18. Every other pressure in the file
reads the near ground for the reason `usedForest` gives - the woods a
settlement lives off are the few tiles nearest it and the country beyond
stays full whatever it does. This one was reading the country.

It reads the settlement's own ground against the country now, and fires below
seven tenths of the ambient share. That is scale-free, which matters because
maps differ: a people who cleared their surroundings stand out however wooded
the map they were dropped on, and a people who settled in a clearing and cut
nothing do not. On six probed seeds the ratio bottomed between 0.32 and 1.18,
so the bar separates them rather than passing or failing all of them.

### Irrigation was asking for a thing this world does not do

`fieldsWorn` wanted the mean fertility of the fields near the market under
0.45. It was never once true, on any seed, in sixty years - and that is not a
threshold set too low. **Fields do not wear.** `Fallow` puts fertility back
toward `Rich` faster than farming takes it out, so a settlement's fields sit
between 0.94 and 1.00 of what the ground can hold for the whole of a run.
There is no threshold that would have worked, because there is no signal:
the measure is flat at its ceiling.

`TestFieldsDoNotWearOut` now says so, and says what should happen if it ever
stops being true. Land that pushes back is most of what would make a
settlement move on, and making fields wear is a change worth having - but it
is a change to what the ground does rather than to what a discovery asks of
it, and it wants a batch of its own.

So irrigation answers the other thing irrigation is for. A field high above
the river is dry whatever its soil is worth, and a channel is the answer to
that and to nothing else. `fieldsDry` reads the mean dampness of the fields
near the market - 1 on the valley floor, 0 at `FloodDepth` above it - and
fires under 0.65. Settlements pick good land and farm at a median of 0.64 to
0.83, so this picks out the ones that have been pushed uphill rather than the
ones that chose where to be.

It is also a better pressure than wear would have been: it can come and go as
a holding spreads, where wear only ever accumulates.

### Where they land now

Twelve seeds:

```
              was          now
irrigation    0/12         9/12    years 16-50
forestry      0/12         4/12    years  6-11
```

The seeds that still get neither are settlements whose condition genuinely
never holds - they did not clear their woods and they farm the valley floor -
and they show a zero in `Pressed` for the honest reason rather than because
nothing could ever have moved it.

**On the batch it is a draw.**

```
offset  0, master: gates: fed 0.59 safe 0.52 held 0.69 all 0.261 food 3.93 | lasted 23/24 mean 151.5 median 139
offset  0, reach:  gates: fed 0.59 safe 0.51 held 0.69 all 0.256 food 4.02 | lasted 23/24 mean 152.5 median 139
offset 24, master: gates: fed 0.59 safe 0.52 held 0.70 all 0.264 food 3.76 | lasted 24/24 mean 147.8 median 126
offset 24, reach:  gates: fed 0.59 safe 0.52 held 0.69 all 0.258 food 4.01 | lasted 24/24 mean 167.1 median 153
```

Nothing is past a threshold. `fed` and `held` did not move at all, `all` went
down 0.005 and 0.006, and the population is flat on one batch and up 19 on
the other, which is what population always does. `food` is up 0.09 and 0.25,
both batches the same direction, which is the only thing here that reads like
a consequence rather than like noise - nine settlements in twelve now get a
fifth more off their fields, and it shows up as a fuller larder and not as
more people.

The golden numbers did not move. Forestry arrives at year six at the
earliest and the golden run is fifteen hundred days, so nothing this changed
happens inside it.

### What pressing toward the land's answers did

The land's answers are worked toward now rather than handed over, and what
does the working is a person.

Each of the six carries two new things: a `habit.Signature`, which is the
kind of moment it belongs to, in the same twenty-dimensional space agents
recognise their own moments in; and a `Cost` in pressure-days. Every tick,
the settlement's most affected person presses on it, and when the pressing
comes to the cost the thing is worked out.

The pressing is one operation:

```go
push := habit.Dot(action.Shared(a, w), habit.Unit(d.Signature))
```

That is the projection of a person's own situation onto the discovery's
direction, and it says both of the things wanted at once. The direction says
whether this settlement is working on the thing at all - a people who are
cold and fed are pointed somewhere other than a people who are warm and
starving. The length says how fast, because how far the situation reaches
along that direction is how hard the moment is actually pressing.

`action.Fit` throws this second half away on purpose: it normalises both
sides, and [situation.go](../core/action/situation.go) even zeroes `Lack` and
`Stock` out of the norm, because a long moment was beating a well-matched one
on cosine. For choosing an action that is right. For arriving at a
technology it is exactly backwards, and the half `Fit` discards is the half
that paces research.

### That a comfortable people stagnate is not a rule here

It is what adding nearly nothing for twenty years comes to. A settlement with
little wrong with it has a short situation vector in every direction, so its
projection onto any discovery is near zero and it arrives at nothing - not
because anything says so, but because that is what the arithmetic does.
`TestAComfortableSettlementWorksNothingOut` stands two identical settlements
on identically thinned ground, makes one hungry and leaves the other wanting
for nothing, and runs twenty years of days over both: the first works out
fishing and the second does not, and has less pressure to show for the same
two decades.

### It happens to a person

The projection is the maximum over everybody rather than the mean or the sum,
which is the whole of the per-agent reading. A thing is worked out by the
person it is happening to hardest, and a settlement of five hundred
comfortable people with one desperate one in it is a settlement where
somebody is about to think of something; a mean would drown them. The
discovery event names whoever that was, so a technology stops being a thing
that happened to a settlement.

### Where they land

Twelve seeds, master against this tree, mean year of arrival:

```
                master        spike
pottery         year  6.0     year 15.1     12/12 both
brewing         year  4.8     year 12.9     12/12 both
fishing         year 11.8     year 14.8     12/12 both
trapping        year 12.3     year 15.9     12/12 both
```

The two that moved most are the two whose conditions were nearly free.
Brewing wanted twelve people and five food on the market shelf, which a
settlement has in its fourth year and then has forever; pottery wanted eight
food. Neither was a pressure, and both are now paced by one - loneliness
among company for the tavern, and curiosity with the leisure to indulge it
for the pot. Fishing and trapping moved less because `forestThin` was already
doing the binding.

Irrigation and forestry are unchanged at nothing: their conditions -
`fieldsWorn`, `forestGone` - almost never read true, so they never begin
accumulating. That was so before this change and is not this change's doing,
but it is now visible as a zero in `w.Pressed` rather than as an absence.

**On the batch it is a draw, and the population came down a little from where
the five new technologies had put it.**

```
offset  0, master: gates: fed 0.60 safe 0.52 held 0.68 all 0.260 | lasted 24/24 mean 164.4 median 176 | estm 0.17
offset  0, press:  gates: fed 0.59 safe 0.52 held 0.69 all 0.261 | lasted 23/24 mean 151.5 median 139 | estm 0.18
offset 24, master: gates: fed 0.57 safe 0.51 held 0.69 all 0.253 | lasted 24/24 mean 152.0 median 129 | estm 0.17
offset 24, press:  gates: fed 0.59 safe 0.52 held 0.70 all 0.264 | lasted 24/24 mean 147.8 median 126 | estm 0.15
```

Nothing is past a threshold. `all` moved 0.001 and 0.011, both up; `fed`,
`held` and `safe` a hundredth or two each. `food` is down 0.40 and 0.16, and
the mean population down 13 and 4, both batches the same direction - which is
the delay showing up as slightly less settlement, and is about half of what
the five technologies added in the commit before.

### What this is a spike of, and what it is not

The skill-gated half of the catalog is untouched. A settlement does not feel
its way to a master mason, and the arch still asks for two of them: capability
and motive are different things and collapsing them into one vector would let
a desperate people invent the arch with nobody who can lay stone. What the
vector decides is when and how fast among the things a settlement is already
in a position to work out.

The cost of going further is legibility. Six hand-written predicates have
become six predicates and six sparse vectors, and the vectors are harder to
argue with: "why did this settlement never get medicine" is a question a
condition answers and a cosine does not. The vectors are kept sparse and
commented for that reason - two or three coordinates each, with the moment
they stand for written beside them - and that is the whole of the mitigation
there is.

### What the deep technologies did

Five technologies at the far end of the tree, and two modifiers that are not
flat multipliers.

The catalog stopped at metallurgy, which a settlement reaches around year
forty-nine, and everything before it was held by nearly everybody. There was
nothing at the top for a settlement to still be climbing at sixty years. The
five are weaving, husbandry, the arch, medicine and the plough, and what they
ask for is masters - which under [core/entity/learn.go](../core/entity/learn.go)
is the one tier nobody can be given, so the top of the catalog cannot be got
by being taught.

Two of them turn knobs that did not exist:

`Warmth` is how much of the cold a body actually feels. Weaving takes it to
three fifths, and it comes off `system.Decay` on the one line both halves of
a winter are read from - the hunger of staying warm and the condition it
takes. It is the first modifier here that does nothing most of the year: a
cloak is worth nothing in June and worth a life in February, so what it buys
depends on where a settlement is and what winters it gets rather than on a
rate. A settlement that housed itself early never learns to weave at all,
which is right, because it solved the same problem another way.

`Healing` is how fast a body climbs back toward the condition its
circumstances would give it, and medicine doubles it. It works one way only:
knowing what to do for a fever gets somebody on their feet sooner and does
not make anybody fall ill quicker. There is a test for each half of that.

Where each of them lands, over twelve seeds:

```
weaving      12/12   years 12-22   the winter, and a quarter of the people out in it
husbandry    12/12   years 13-27   trapped-out woods, tools, and builders for the pens
the arch      4/12   years 36-59   quarrying, two master masons with eighty raisings
medicine      5/12   years 25-36   writing, a master scholar with forty turns at it
the plough    2/12   years 39-57   metal, three master farmers with three hundred days
```

Two of them had to be given a real bar before they were technologies at all.
Husbandry first asked for what trapping asks for and a little more, and fired
on the same day trapping did on eleven seeds of twelve - which is not a
second technology but a longer sentence about the first. It asks for builders
now, because a kept beast needs somewhere to be kept, and it lands a decade
after trapping. The plough asked for two master farmers at a hundred and
fifty days, which a settlement with metal already had, so it arrived the same
day metallurgy did on six seeds of seven; at three master farmers with three
hundred days it is past what any but a seriously farming people reach.
Medicine went the other way - at two master scholars it fired on nothing at
all, because reading stops at the top of the apprentice tier and a master
scholar is somebody who has tutored for years.

**On the batch it is close to a draw, and what moved is the population,
upward, on both.**

```
offset  0, master: gates: fed 0.59 safe 0.54 held 0.70 all 0.272 | lasted 24/24 mean 146.9 median 141 | estm 0.16
offset  0, deep:   gates: fed 0.60 safe 0.52 held 0.68 all 0.260 | lasted 24/24 mean 164.4 median 176 | estm 0.17
offset 24, master: gates: fed 0.57 safe 0.51 held 0.70 all 0.252 | lasted 24/24 mean 131.1 median 108 | estm 0.15
offset 24, deep:   gates: fed 0.57 safe 0.51 held 0.69 all 0.253 | lasted 24/24 mean 152.0 median 129 | estm 0.17
```

Nothing is past a threshold. `all` went down 0.012 and then up 0.001; `fed`,
`held`, `safe` and `estm` all moved by a hundredth or two and disagreed about
the sign. `lasted` is 24 of 24 on all four with nothing extinct.

The mean population is up 17 and 21 and the median up 35 and 21, on both
batches, in the same direction. This file says population is never worth
believing on its own and that stands - but the direction is worth writing
down, because it runs against the last few changes. Five new multipliers are
five new multipliers and the settlements are a little larger for them. What
keeps it to about a tenth rather than a third is that three of the five are
rare and all five are late: nothing here arrives before year twelve, and the
two that lift a yield outright reach two and twelve settlements, one of them
only in the last third of a run. A settlement that has made no masters gets
weaving and nothing else.

`food` is up from 3.95 to 4.33 on the default batch and unchanged on the
other, which is the same story told quietly.

### What earning the research did

A discovery now asks a settlement what it has actually done, not how many of
its people are over a line.

Every skill-gated technology was unlocked by `skilled`, a count of agents at
or above some level of a skill. That was a fair question when a level could
only be got one way. It stopped being one when a lesson began stopping short
of the teacher: somebody can now stand in the journeyman tier having been
shown the whole of it and never once had their hands on the work, and it was
that person the catalog was reading. Agriculture asked for two people at a
fifth of farming, in settlements whose most practised farmer had broken
ground five times in sixty years. The field was being invented by people who
had never really farmed.

So `entity.Agent.Practice` counts, per skill, how many times this pair of
hands has actually done the work. Only `Learn` raises it - being taught and
reading do not - so it is the half of a person's history that `Skills`
cannot tell you. `system.adept` asks for both halves, and the four
skill-gated discoveries ask it:

| | asked before | asks now |
|---|---|---|
| agriculture | 2 at 0.20 farming | 2 apprentices, 80 days on the ground |
| masonry | 2 at 0.30 building | 2 journeymen, 40 raisings |
| writing | 3 at 0.30 scholarship | 3 journeymen, 20 turns at the work |
| metallurgy | 2 at 0.40 crafting | 2 masters, 100 makings |

The numbers are read off what settlements reach rather than chosen. A probe
over six seeds printed every agent's tier and practice at sixty years, and
each threshold is set where it separates the settlements that do the work
from the settlements that do not - which for farming is a real divide, since
three of the six had masters with two to seven hundred days on the ground and
the other three had nobody past thirty.

**On the summary line this change did nothing, and the summary line is the
wrong instrument for it.**

```
offset  0, master: gates: fed 0.59 safe 0.49 held 0.67 all 0.239 | lasted 24/24 mean 146.0 median 121 | estm 0.20
offset  0, adept:  gates: fed 0.59 safe 0.54 held 0.70 all 0.272 | lasted 24/24 mean 146.9 median 141 | estm 0.16
offset 24, master: gates: fed 0.58 safe 0.51 held 0.70 all 0.252 | lasted 23/24 mean 145.3 median 123 | estm 0.19
offset 24, adept:  gates: fed 0.57 safe 0.51 held 0.70 all 0.252 | lasted 24/24 mean 131.1 median 108 | estm 0.15
```

`all` went up 0.033 and then did not move at all. `safe` up 0.05 and then
0.00. `fed`, `held` and `lasted` are unmoved on both. The population points
opposite ways in the way it always does. The one reading that agrees with
itself is `estm`, down 0.04 on both batches, which is inside the threshold
this file sets and is the tail of the change before this one rather than
anything here.

What moved is the shape of a settlement's history, which no gate in that line
measures. The same probe run on master and on this tree, twelve seeds, the
year each technology arrived:

```
                master          adept
agriculture     year  3.9       year 12.4      12/12 both
masonry         year  4.6       year 13.1      12/12 both
writing         year  6.6       year 18.9      12/12 both
metallurgy      year 15.2       year 49.2      12/12 -> 9/12
```

On master a settlement holds the whole catalog by year fifteen, which is a
quarter of one life. It now takes until year forty-nine to work metal, and
three settlements in twelve never do. The gap between founding and the last
technology went from eleven years to thirty-seven.

That is the change, and the reason the year-sixty gates cannot see it is that
sixty years is long enough to arrive anyway. A settlement kept off masonry
for eight extra years builds fewer houses in those years and the same number
by the end; what it does not have is the compounding it used to get from
holding every multiplier before its founders were dead. Anybody wanting this
to show in the gates should read a shorter batch, where the difference has
not yet been slept off.

Agriculture is the one that had to be raised twice. At five-and-twenty days
it did nothing at all, because twenty-five days of farming is exactly what
carries somebody to the apprentice tier and the practice bar sat on top of
the tier bar rather than beyond it. Eighty is past it, and moved the mean
arrival from year 5.1 to year 12.4. It is still 12 of 12: every settlement
eventually farms, and this asks it to farm first and learn the rotation
after, which is the order the thing actually happened in.

The land's own answers were left alone. Fishing, trapping, forestry,
irrigation, pottery and brewing are unlocked by need and not by skill - a
hungry people by a river will fish, whatever they know - and the catalog
already says so where they are defined.

### What the moral coordinates did

Every dimension of a habit drifts per agent now. Five of the twenty - the
moral ones, honesty through caution - used to be held fixed across the whole
population, on the reasoning that an agent's values are the same across every
candidate it weighs and so drift on them would cancel out of the choice.

It does not cancel. What the situation carries there is the agent's own norms,
one number apiece and the same for every errand; what the habit carries is per
action, and the fit is their product summed, so a habit reading a little
differently on honesty for this act and not that one changes which act wins.

But it buys less than the arithmetic suggests. The situation never varies on
those coordinates, so what drift adds is a standing disposition rather than a
judgement made afresh: a pull toward some acts and away from others, sized by
how much the agent holds the value it hangs on. A scrupulous agent's own
reading of which work is honest shapes what it does; a careless one's is
multiplied by a norm near nothing.

**It is a draw, and what moved is how sharply people decide rather than how
far apart they are.**

```
offset  0, tiers: gates: fed 0.58 safe 0.45 held 0.71 all 0.244 | lasted 24/24 mean 174.3 median 110 | belng 0.50 estm 0.21
offset  0, moral: gates: fed 0.59 safe 0.49 held 0.67 all 0.239 | lasted 24/24 mean 146.0 median 121 | belng 0.47 estm 0.20
offset 24, tiers: gates: fed 0.57 safe 0.48 held 0.72 all 0.253 | lasted 24/24 mean 145.6 median 104 | belng 0.50 estm 0.20
offset 24, moral: gates: fed 0.58 safe 0.51 held 0.70 all 0.252 | lasted 23/24 mean 145.3 median 123 | belng 0.50 estm 0.19
```

Nothing reached its threshold. `all`, which is the gate a birth passes, did not
move at all: 0.244 to 0.239 and 0.253 to 0.252. `held` is down 0.04 and 0.02;
`fed` up 0.01 on both. The mean population fell 28 on one batch and stood still
on the other, which is what this file says population does.

The one reading that keeps its sign is the safety gate, up 0.04 and 0.03 here
and up 0.03 and 0.03 when the same change was measured against the tree before
the tiers landed. Four batches, four times up, never once reaching 0.05. That
is worth writing down and is not worth believing yet; if it is real it is small,
and the thing that would settle it is more seeds rather than more readings of
these twenty-four.

What is worth recording is a reading `tune` does not print. Over sixty years on
two seeds outside the batch, the population's habit spread went 0.33 and 0.26
to 0.32 and 0.27 - which is to say nowhere - while the choice entropy went 0.83
to 0.93 and 0.81 to 0.86. Five more coordinates of drift bring candidates
nearer one another in fit, and a moment with less between its candidates is
decided less sharply. So the change did not spread the population out; it made
each of them a little less certain. (Those two seeds were measured on the tree
before the tiers; the tiers move learning and not this, but the numbers are
from the older tree and are quoted as the shape of the thing rather than as
readings of the one committed here.)

That is the thing to weigh before turning this dial further. If what is wanted
is two people looking at one piece of work and disagreeing about it, this is
not the lever: that wants the act's own valence in the situation rather than
only the agent's norms, the way the value rule already reads it through
belief.Conscience. See action.ValenceOf.

### What earning the last of it did

Competence stopped being something the settlement accumulated and became
something people earn. Three things changed and they are one change; see
[core/entity/learn.go](../core/entity/learn.go).

Learning curves. A skill used to go up by the same flat step from any source
at any level, so the hour that took somebody from nothing to some use was
worth exactly what the hour that would have made them the best there is was
worth. Now the step is scaled by the tier it is taken from - novice,
apprentice, journeyman, master, the four equal quarters of the range - at 1,
0.55, 0.3 and 0.15. The first is left at one deliberately: the opening years,
when a settlement is short of everything, are exactly as hard as they were,
and what got harder is the far end. Mastery by work is about three times the
labour it was.

Being shown stops short. A lesson leaves the pupil `TaughtGap` under the
teacher and never past the threshold of the master tier, so a master's pupil
comes out a journeyman with the whole last quarter still to work for, and
somebody still in the novice tier has nothing to show at all. Reading alone
stops sooner, at the top of the apprentice tier, which is the level the
settlement asks for before it will let anybody tutor. The teacher's level
tells twice over, because it sets the rate as well as the ceiling: an hour
with somebody who has just cleared the floor is worth a fraction of an hour
with a master. The old flat step could not say that, and the pair it rewarded
most was two amateurs teaching each other what they both already half knew.

And `action.Teach` is gated on there being something to teach, without which
the act would have gone on paying its esteem for lessons nobody learned
anything from.

**It is a real and large change, and the two batches agree about the size and
the direction of all of it.**

```
offset  0, master: gates: fed 0.57 safe 0.51 held 0.82 all 0.285 | lasted 23/24 mean 178.6 median 160 | belng 0.57 estm 0.64
offset  0, tiers:  gates: fed 0.58 safe 0.45 held 0.71 all 0.244 | lasted 24/24 mean 174.3 median 110 | belng 0.50 estm 0.21
offset 24, master: gates: fed 0.57 safe 0.50 held 0.82 all 0.292 | lasted 22/24 mean 178.4 median 157 | belng 0.57 estm  n/a
offset 24, tiers:  gates: fed 0.57 safe 0.48 held 0.72 all 0.253 | lasted 24/24 mean 145.6 median 104 | belng 0.50 estm 0.20
```

`estm` is down from 0.64 to 0.21 and 0.20, `held` down 0.11 and 0.10, and
`belng` down 0.07 on both. Those three are past the thresholds this file sets
and both batches say the same thing. `all` is down 0.041 and 0.039, which is
just inside the 0.05 this file asks for - but two batches agreeing that
closely on the direction and the size is worth more than either of them
alone, and it is the same story the other three tell. `fed` did not move.

Esteem is the one to understand rather than to mourn. `pass/practice>pupil`
was the settlement's fourth commonest act at 8.1% of everything anybody did,
997,606 lessons paying 0.2 of esteem each, and it is under 1% now. Most of
that esteem was never earned by teaching anybody anything: it was a
settlement of journeymen sitting down with each other to confirm what they
both knew, and it went away when a lesson had to have something in it. `held`
and `belng` follow it down for the same reason - the belonging in the act
went with the act. If the esteem tier should be fuller than 0.21, the honest
way is to pay it for being good at something rather than for the ceremony of
saying so, and that is a change of its own with a batch either side of it.

**The runaway is not gone, and this batch is the evidence against saying it
is.** The typical settlement is smaller: the median fell from 160 and 157 to
110 and 104, on two batches that agree. But seed 8 came to 807, which is
larger than any settlement master has on either batch, and it is what holds
the mean at 174.3 against master's 178.6 while the median falls by fifty.
What the change did to the distribution is not to flatten it but to make the
top of it rarer and more contingent - and, since bodies and minds landed
first, more a matter of who happened to be born. A settlement that produces
an exceptional individual now compounds harder than it used to, because
masters are scarce and worth more. Anybody wanting the tail actually cut has
to cut it somewhere else; this only raised the price of getting there.

Nothing is dying for it. `lasted` went to 24 of 24 on both batches, against
23 and 22, with nothing extinct on any of the four.

The larder agrees. `food` is down from 4.33 to 3.07 and 3.42, and
`take/fish@water` rose while the farmed and gathered staples fell: a
settlement that cannot mint master farmers by standing people next to one
lives closer to the margin and spreads its bets.

### What turning the constants buys

Nothing that can be measured, which is worth writing down so that nobody
spends the batches finding it out again. Seven full batches, each moving one
of the numbers above and leaving the rest:

```
committed     gap 0.20  curve 1,.55,.3,.15    all 0.236  held 0.71  estm 0.19
              gap 0.15                        all 0.254  held 0.71  estm 0.22
              gap 0.10                        all 0.243  held 0.71  estm 0.20
              gap 0.20  curve 1,.7,.45,.25    all 0.248  held 0.72  estm 0.19
              gap 0.10  curve 1,.7,.45,.25    all 0.249  held 0.70  estm 0.18
              gap 0.20  curve 1,.3,.12,.05    all 0.238  held 0.70  estm 0.22
              InheritedSkill 0.5 to 0.25      all 0.246  held 0.71  estm 0.19
              InheritedSkill 0.5 to 0.0       all 0.241  held 0.71  estm 0.22
```

Those were taken on the tree before bodies and minds merged in, so the
absolute numbers are not the run above and are kept for the shape rather than
for themselves. The shape is the point: every variant lands inside `all`
0.236 to 0.254 against a threshold of 0.05, `held` 0.69 to 0.72, `estm` 0.18
to 0.22, and 24 of 24 lasted with nothing extinct. Halving all three upper
tier rates - journeyman from 0.3 to 0.12, master from 0.15 to 0.05 - moved
`all` by 0.002.

The reason is that teaching saturates. Lessons sit under 1% of all acts under
every one of those settings, against 8.1% before the change: everybody rises
to whatever ceiling they are given and then there is nothing left to pass,
wherever it is put. So what this change does is carried by its shape - that a
ceiling exists at all, and that Teach is gated on there being something to
teach - and not by the numbers in it. That is a good property, because none
of it is balanced on a knife edge, and it also means there is no dial here
worth turning.

`InheritedSkill` is the surprise. Half a parent's skill at birth looked like
the last free ride in the model, and taking it to zero moved nothing, because
children are born to parents whose skills are mostly middling and half of
middling lands in the novice tier, where learning was already at full rate.

### What this does not yet read

`Mind.Plasticity` is how fast one person takes to a thing, and it reaches
reach and not skill: `action.Practise` scales by it and nothing in
`entity.toward` does. That was true before this change and is only worth
saying now because the ladder it would scale has just been built. A settlement
where some people climb faster than others is what the trait was written for -
its own comment says so - and wiring it through the tiers is the obvious next
change, with a batch either side of it.

### What bodies and minds did

Bodies and minds. What an agent burns to stay alive, what cold it can stand,
how fast practice brings a craft within reach, how firmly it settles a moment
and how far it will walk for good ground were one number apiece for the whole
population, written into the systems that read them. A settlement of five
hundred had exactly one body and one cast of mind between them.

They are drawn per agent now and inherited with drift, the way Vitality and
Norms and Temperament already were, and they live in `entity.Body` and
`entity.Mind` rather than as loose fields - because a species is a body and a
mind it hands its young, and the deer are coming. See core/entity/trait.go.

Each measure is a multiplier on something real and there is a test for each
saying so: a body that burns more is hungrier by the end of the same day, a
frail one pays for a winter a hardy one shrugs off in both hunger and
condition, a quick mind is further into a craft after the same five turns at
it, a resolute one settles the moment at a sharper temperature, a wider
horizon looks further for a field.

**It is a draw on the settlements, and the two batches disagree about the one
thing that moved.**

```
offset  0, master: gates: fed 0.56 safe 0.51 held 0.81 all 0.285 | lasted 23/24 extinct 0 mean 211.9 median 199 | belng 0.52
offset  0, traits: gates: fed 0.57 safe 0.51 held 0.82 all 0.285 | lasted 23/24 extinct 0 mean 178.6 median 160 | belng 0.57
offset 24, master: gates: fed 0.56 safe 0.52 held 0.84 all 0.290 | lasted 23/24 extinct 0 mean 171.2 median 140 | belng 0.58
offset 24, traits: gates: fed 0.57 safe 0.50 held 0.82 all 0.292 | lasted 22/24 extinct 0 mean 178.4 median 157 | belng 0.57
```

The default batch had `belng` up 0.05, which is exactly the figure this file
says is worth believing on two batches that agree; the independent batch has it
down 0.01. `all` did not move on either - 0.285 to 0.285, and 0.290 to 0.292.
`fed` is up 0.01 on both, which is a fifth of its threshold. `lasted` went 23
and then 22 against 23 and 23. The mean population fell 33 on one batch and
rose 7 on the other, which is what this file says population always does.

So: **the traits are behaviourally neutral at the spread they were given.**
That is not a disappointment, it is the spread doing what it was set to do.
`world.TraitSpread` is 0.12 and `TraitDrift` 0.08, taken from Vitality, which
chose those numbers so that a settlement's fortunes would turn on what people
want and believe rather than on who was born strong. At 0.12 the traits decide
which of two neighbours breaks the further field and which of them takes to
masonry, and they do not decide whether the settlement eats.

That dial is the thing to turn if they should matter more, and it should be
turned on its own, with a batch either side of it - not folded into the change
that introduced them, where a moved number could not be told apart from the
traits themselves being a good idea.

### What the rivers did

The rivers got a shape. A channel used to be the steepest way down and nothing
else, redrawn from scratch every time the ground moved, which on a smooth
hillside is a straight line along one of eight bearings - and it cut the same
way whether it was a gully or the drainage of half the map. Now it cuts the
outside of every bend it makes and lays most of what it cuts on the inside, so
the bend grows and the channel walks sideways across its own valley. See
meander.go.

The other half is that the rock underneath decides what the water can cut.
`hardness` runs from granite at one and a half down to shale at not quite a
half, and it divides both the incision that cuts the first valleys and the
bank a meander takes - so a river takes a gorge out of shale and is turned
aside by granite. The geology has been on the map since the soils landed and
this is the first thing that reads it.

Over three seeds and forty ages of weather the share of river tiles that turn
goes from about a half to two thirds. `bankCut` at 2.5 metres an age for the
greatest river on the map is the knee of that: twice as much buys a hundredth
more and starts pulling the hillsides about, half as much buys half the bends.

**It is a draw on the settlements, and the two batches disagree about
everything that moved.**

```
offset  0, master: gates: fed 0.56 safe 0.48 held 0.83 all 0.269 | lasted 23/24 extinct 1 mean 182.8 median 147 | belng 0.57
offset  0, rivers: gates: fed 0.56 safe 0.51 held 0.81 all 0.285 | lasted 23/24 extinct 0 mean 211.9 median 199 | belng 0.52
offset 24, master: gates: fed 0.56 safe 0.52 held 0.84 all 0.290 | lasted 23/24 extinct 0 mean 171.2 median 140 | belng 0.58
offset 24, rivers: gates: fed 0.55 safe 0.47 held 0.82 all 0.251 | lasted 23/24 extinct 0 mean 159.3 median 123 | belng 0.57
```

The default batch had `belng` down 0.05, which is exactly the figure this file
says is worth believing on two batches that agree - and there was a good story
ready for it, that a channel walking across a valley walks between neighbours
and cuts a settlement in half. The independent batch has it down 0.01. `all`
went up 0.016 and then down 0.039; `safe` up 0.03 and then down 0.05. Pooled
over the forty-eight seeds `all` is down 0.011 and `belng` down 0.03, both
inside what chance moves them by, and the population points opposite ways in
the way it always does. The story was fitted to one batch and is not in the
numbers.

### What this cannot see

The golden numbers in core/system did not move for the meandering at all, and
that is worth knowing about them: a golden run is 1500 days and an age of
weather is 3600, so no run in that test ever reaches one. They cover the map
as it is made and the settlement's first four years on it. Anything about
erosion - this change, the soils before it, the ages that wear a valley down
over a lifetime - is measured here and nowhere else.

### What the soils did

The ground got a make-up. There is rock under every tile now - granite,
limestone, sandstone or shale, laid down in regions by two coarse lattices
crossed at their own middles - and the soil over it is a mixture of sand, silt
and clay weathered out of that rock. Two shares are kept and the third is the
remainder, so a soil cannot disagree with itself.

The part worth having is that the water sorts it. Erode carried one load and
now carries three, each settling at its own rate: sand goes down at the first
slackening, silt travels to where the river spills over its bank, clay stays
up in the water longest. The three average within a hundredth of the single
figure they replace, so a map silts up at about the rate the model was
measured at, and what is new is where each grain of it lands. Held against the
bedrock, the texture of a map is the geology with the valleys rewritten by the
river.

Then two things read it. Fertility is multiplied by how near a soil is to a
loam - both ends of that scale are poor, and the best ground is in the middle
of it - and what an age of weather strips off a tile is multiplied by how
sandy it is, so a settlement that ploughs its sandy slopes loses them faster
than one that ploughs its clay. See bedrock.go.

**It was a draw on the settlements, on two batches.** That was the intended
result: structure and not difficulty.

```
offset  0, master: gates: fed 0.56 safe 0.52 held 0.83 all 0.291 food 4.76 | lasted 22/24 extinct 0 mean 186.4 median 159
offset  0, soils:  gates: fed 0.56 safe 0.48 held 0.83 all 0.269 food 4.64 | lasted 23/24 extinct 1 mean 182.8 median 147
offset 24, master: gates: fed 0.55 safe 0.53 held 0.86 all 0.295 food 4.31 | lasted 23/24 extinct 0 mean 161.8 median 100
offset 24, soils:  gates: fed 0.56 safe 0.52 held 0.84 all 0.290 food 4.40 | lasted 23/24 extinct 0 mean 171.2 median 140
```

The default batch alone would have been worth a second look: `safe` down 0.04,
`all` - the gate a birth has to pass - down 0.022, and one settlement dying
where none had. None of that survives the independent batch, where the same
two readings are down 0.01 and 0.005 and nothing goes extinct. Pooled over the
forty-eight seeds `all` is down 0.014 and `safe` 0.025, both an order inside
what chance moves them by, and the population disagrees in direction between
the two batches in the way it always does. `fed` is the same to a hundredth on
both.

What did move, and is not in these numbers, is how much the seeds differ from
each other about their ground. A settlement founded over sandstone has a
poorer valley than one founded over shale, and that is a new thing for a seed
to decide. Anybody reading a population figure off a small batch after this
should expect it to be looser than it was.

### What the lapse rate did

The weather learned about height. It was read by latitude alone, so the top of
a mountain was exactly as warm as the valley it stood over - `Height` was the
one thing a tile carried that nothing in the weather ever asked for. The air
now cools by `world.Lapse`, six and a half degrees a kilometre, and the cold a
body feels, the growing weather the ground gets, what a pack keeps by, and
where seed will take are all read where they are rather than off the row. Two
rules became one on the way: ground whose year never warms past `Frost` is bare
outcrop, which used to be said of the poles alone and now covers a peak without
naming one, and the woods stop below it.

**On this map it was behaviourally neutral, and that was the expectation
going in.** The default valley is sixty metres of relief with two hundred and sixty
of high country on it, so the highest ground on it is two degrees colder than
the river and no ground on it is ever frozen: the ploughable share of the map
does not move, and the tree line has nothing to bite on. What the run measures
is two degrees spread over the shoulders of a valley.

Nothing crossed a threshold on either batch:

```
offset  0, master: gates: fed 0.55 safe 0.51 held 0.83 all 0.278 food 4.34 | lasted 22/24 extinct 0 mean 143.3 median 132
offset  0, lapse:  gates: fed 0.56 safe 0.52 held 0.83 all 0.291 food 4.76 | lasted 22/24 extinct 0 mean 186.4 median 159
offset 24, master: gates: fed 0.58 safe 0.51 held 0.85 all 0.297 food 4.52 | lasted 23/24 extinct 0 mean 163.9 median 151
offset 24, lapse:  gates: fed 0.55 safe 0.53 held 0.86 all 0.295 food 4.31 | lasted 23/24 extinct 0 mean 161.8 median 100
```

The default seeds gained a third of their mean population and a fifth of their
median, which read like a win and is not one. The independent batch does not
confirm it: on seeds the baseline never draws from the mean is flat - 163.9 to
161.8 - and the median falls by a third, the opposite direction. `fed` went up
0.01 on one batch and down 0.03 on the other, `all` up 0.013 and down 0.002,
and `lasted` and `extinct` are identical on both. Pooled over the forty-eight
seeds, `fed` is down 0.01 and `all` up 0.005, both an order inside what chance
moves them by. So the honest reading is a draw, and the population figures on
either batch are the loosest number in the file doing what it always does.

What the change buys is not on this map at all. On a globe the mountains are
ten times as high, and there the same constant is the difference between a
world with a tree line and one without: at 1024 by 512 the frozen share of the
dry land goes from 33% to 39%, and the 6% is snow on the peaks rather than more
ice at the poles - the highest ground there is 2344 metres against a frost line
of 2275 at the temperate latitudes. That is a reading taken at generation, on
seed 1; what a globe settlement then does with it is measured in
[baseline-globe.md](baseline-globe.md).

### What fencing moved

Hedges round the large holdings; see "Fences" in docs/action-space.md. This is
an account of an earlier change, kept because its second batch is the clearest
worked example in the file of a population move that turned out to be a draw.
It was measured against the master of its own day, not against the run above.

Against the batch it replaced - fed 0.56, safe 0.53, all 0.291, lasted 24 of
24, mean 238.2, median 198 - the gates did not move at all:
`fed` held, `safe` and `all` came up by 0.01 and 0.007, and one settlement of
twenty-four fell below its founding size. What moved was the population on these
seeds, mean by a sixth and median by a third, and an independent batch does not
confirm it:

```
offset 24, before: gates: fed 0.57 safe 0.52 held 0.83 all 0.292 | lasted 24/24 extinct 0 mean 229.5 median 202
offset 24, fenced: gates: fed 0.58 safe 0.50 held 0.84 all 0.295 | lasted 24/24 extinct 0 mean 218.1 median 211
```

On seeds the baseline never draws from, the gates are the same to within a
hundredth, nothing was lost, the mean is down 5% and the median is up. So the
fall on the default seeds is a draw and not a cost: two batches agree that
hedging the big holdings is behaviourally neutral, and what it buys is that
people stop walking through the corn.

### What the mountains cost

The land grew mountains. `raise` used to be one texture at one amplitude scaled
to sixty metres, and every seed came out the same gentle bowl - nine tiles in
ten under a tenth of a slope, with the highest ground only the largest of the
same lumps. It now raises a lowland to exactly that same sixty metres and
stands two hundred and sixty metres of ridged high country on a fifth of it,
with the rivers cutting their own valleys into what is left. The high country
is scaled to the width of the map and the lowland is not, so a wider map is a
bigger country at the same ruggedness and its valley floor is unchanged. See
relief.go.

What that costs is ground. The mountains are not farmland, and the gentle open
ground a settlement can plough falls from 72% of the map to 55% over the first
eight seeds.

What that costs in people is about a sixteenth. Pooled over all seventy-two
seeds the mean population goes 230 to 215 and the median 198 to 176, and all
three batches agree in direction - down a thirtieth, a fifteenth and an
eleventh.

Nothing else moved, and two things moved the wrong way for a change that costs
ground: pooled, `fed` went 0.563 to 0.570 and `all` - the gate a birth has to
pass - 0.294 to 0.298, both up and both far inside their thresholds. `lasted`
went 72 of 72 to 71 and nothing went extinct on any batch. The settlements are
not doing worse on the land they have; there is less of it and slightly fewer
of them on it.

**These three batches were taken on c8513bc, before the hedges landed**, and
are compared against that commit's own numbers. The hedges went in while they
were running and changed how a journey goes, so the figures above are a
measurement of the mountains and not of the tree they are committed in. They
were not taken again because master took four behavioural changes during the
day this branch was open, each roughly as often as a three-batch run takes to
finish, and a number chased under those conditions never lands. What is worth
holding is the direction, which held across three separate attempts on three
different masters. Whoever next has cause to refresh this file should simply
refresh it; nothing here needs preserving.

This was taken with the map at its default eighty by thirty-six, which is the
width the high country is quoted at. A batch taken on a bigger map is a
different measurement and not comparable to this one: the mountains grow with
the map, the valley does not, and the ploughable share of the tiles rises from
55% to about 63% by two hundred and forty wide.

## How much of that is chance

The same code on two further batches of twenty-four seeds, which `-offset`
draws from a stretch of the seed space the baseline never touched:

```bash
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 24 -quiet
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 48 -quiet
```

```
offset  0: gates: fed 0.55 safe 0.51 held 0.83 all 0.278 food 4.34 hungry-with-food 0.28 | lasted 22/24 extinct 0 mean 143.3 median 132 | phys 0.60 safe 0.52 belng 0.53 estm 0.67
offset 24: gates: fed 0.58 safe 0.51 held 0.85 all 0.297 food 4.52 hungry-with-food 0.28 | lasted 23/24 extinct 0 mean 163.9 median 151 | phys 0.62 safe 0.51 belng 0.56 estm 0.59
offset 48: gates: fed 0.57 safe 0.50 held 0.84 all 0.275 food 5.67 hungry-with-food 0.26 | lasted 21/24 extinct 1 mean 170.1 median 161 | phys 0.58 safe 0.50 belng 0.59 estm 0.64
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 143, 164, 170, the median 132, 151, 161. `lasted` was 22, 23
and 21 of 24, and one of the three lost a settlement outright. The mean is the
loosest reading here and always has been: a settlement that runs away is worth
as much as the twenty that did not, and the handful sitting over four hundred
are most of the distance between the draws. The median is looser on this tree
than it used to be, and the mountains are why: how much of a map's lowland a
range happens to cover is now one of the things a seed decides, so the seeds
have more to differ about and the middle of them moves further.

What holds still is the per-agent side. `fed` moved by 0.03 across the three,
the mean needs by 0.02 to 0.08, and `all` sat between 0.275 and 0.297. So:

| reading | moved by chance | worth believing at |
|---|---|---|
| fed, and the four mean needs | 0.02-0.05 | 0.05 on two batches that agree |
| gates all | 0.022 | 0.05 |
| lasted, extinct | 2 of 24, 1 of 24 | 5 of 24 |
| mean population | 16% of itself | all three batches, pooled |
| median population | 18% of itself | all three batches, pooled |

So a change that only moves the population numbers has not been shown to do
anything, and one batch cannot show it either way. Take all three and pool them
before believing any of it, and say in the commit what the pooled figure was.

The change that raised the mountains is the worked example, and it is worth
keeping because of how far the answer wandered. Measured against the master of
the day, its first batch alone said a third fewer people; the three batches
together said a seventh, with one of them saying the population had gone up.
Measured again after merging the master it first landed on, the same three said
a fourteenth. Measured a third time, on a master where the wear on a road had
changed underneath it, all three agreed in direction for the first time and
said a sixteenth. None of those batches was wrong and none was mismeasured. The
first was simply not an answer, and neither is any single batch taken after it.

The other half of that lesson is about the tree rather than the seeds. Two of
those three measurements were stale before they could be committed, because
master took four behavioural changes in a day and each moved the numbers the
comparison was against. When that is happening, refreshing this file is a race
and not a task: take the batches, say which commit they were taken on, and let
whoever needs them next take them again.

## What the ceiling is doing

`system.MaxPopulation` is 5000, and none of the seventy-two settlements above
comes near it: the largest is 508. That is the point of where it sits. It was
400 until the batch above, which two of these seeds stood exactly on, and a
settlement held at a ceiling reads the same as one that found its level - so
the headline number was being decided in a constant rather than out on the
land.

Lifting it moved nothing but the tail. Every per-agent reading came out the
same to the digit - `fed` 0.57, `all` 0.293, 0.277, 0.309 against 0.293, 0.277,
0.310 - and every median was unchanged, because the two seeds that were pinned
were the only ones affected: seed 1 went 400 to 502, seed 21 400 to 508, and
the mean moved with them. Both are worse fed and lonelier at their new size
than they were at 400, which is what a settlement finding its own ceiling looks
like.

`world.Crowded` counts everyone turned away by the cap on a tick, and it is
what to read if the question comes up again. While it stays at nothing, the
land is doing the binding.

The ceiling is a flag on all three commands - `-cap` on headless, tune and
watch, and a line on watch's start screen - and `-cap=0` takes it off
altogether, leaving the land as the only thing that stops a settlement. That
is the setting to reach for when the question really is what a world carries,
which a globe makes worth asking; it is not the setting to take a baseline
batch under, because a day costs what the population squared costs and a seed
that runs away takes the batch with it. Any batch quoted in this file was
taken at the default 5000 unless it says otherwise.

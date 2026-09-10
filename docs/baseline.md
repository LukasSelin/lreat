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
   1  105   135     17     28     31 |  0.62  0.49  0.66  0.13  0.06 |  1.00
   2  143   169     19      2     55 |  0.64  0.46  0.55  0.12  0.06 |  1.00
   3  350   279     28     50    118 |  0.66  0.51  0.58  0.24  0.05 |  1.00
   4  234   171     14     73     65 |  0.62  0.51  0.41  0.24  0.03 |  1.00
   5  229   212     26     64    140 |  0.68  0.52  0.36  0.33  0.10 |  1.00
   6  315   358     28     24    227 |  0.62  0.47  0.63  0.24  0.10 |  1.00
   7   45   143     19     20     47 |  0.66  0.59  0.55  0.08  0.06 |  0.97
   8  807   375     28    107    154 |  0.68  0.52  0.60  0.15  0.06 |  1.00
   9   96   253     33     19     69 |  0.60  0.55  0.51  0.30  0.11 |  1.00
  10  108   241     29     13     53 |  0.54  0.46  0.47  0.36  0.07 |  1.00
  11   62   169     29     24     49 |  0.69  0.49  0.41  0.27  0.04 |  1.00
  12  110   115     12     33     21 |  0.57  0.50  0.32  0.40  0.04 |  1.00
  13   91   145     33     42     53 |  0.63  0.57  0.64  0.18  0.06 |  1.00
  14   70   131     24     36     45 |  0.68  0.52  0.49  0.26  0.04 |  1.00
  15   79   137     22     18     40 |  0.54  0.47  0.36  0.26  0.03 |  1.00
  16   71   134     28     24     39 |  0.61  0.56  0.60  0.22  0.03 |  1.00
  17   88   235     35     45     60 |  0.58  0.60  0.58  0.17  0.09 |  1.00
  18  217   279     27     57     95 |  0.57  0.51  0.59  0.15  0.08 |  1.00
  19  303   293     41     43    238 |  0.58  0.60  0.62  0.24  0.14 |  1.00
  20   81   274     20     46     96 |  0.53  0.52  0.40  0.04  0.01 |  0.99
  21  157   193     27     56     78 |  0.57  0.50  0.43  0.16  0.02 |  1.00
  22  168   255     26     42     43 |  0.49  0.50  0.46  0.10  0.03 |  1.00
  23  162   201     40     75    133 |  0.63  0.52  0.58  0.13  0.04 |  1.00
  24   93   257     27     36    111 |  0.62  0.57  0.31  0.17  0.06 |  1.00

take/berries@wood                  2920967  25.2%
dwell/rest                         2909562  25.1%
consume/provision                  2100473  18.1%
dwell/guard@market                 1319276  11.4%
dwell/meet@tavern>neighbour         543603   4.7%
take/fish@water                     526760   4.5%
take/timber@wood                    203994   1.8%
transfer/material>requester         174059   1.5%
exchange/coin>provision@market      161143   1.4%
exchange/material>coin@market       124498   1.1%
pass/practice>self                  117601   1.0%
take/grain@field                     88512   0.8%
pass/practice>pupil                  84732   0.7%
tend/plant@open                      80212   0.7%
transfer/provision>needy             55243   0.5%
take/game@wood                       52542   0.5%
make/timber>tool@bench               44912   0.4%
raise/timber>dwelling@open           39097   0.3%
take/stone@outcrop                    9754   0.1%
transfer/provision<holder             8828   0.1%
strike/person>wrongdoer               8459   0.1%
tend/clear@open                       4855   0.0%
tend/water@field                      1606   0.0%
raise/timber>road@ground              1373   0.0%
make/stone+timber>tool@forge           665   0.0%
raise/timber+stone>granary@open        461   0.0%
dwell/look                             312   0.0%
make/provision+timber>meal@hearth      278   0.0%
raise/timber>tavern@open               230   0.0%
raise/timber+stone>market@open          84   0.0%
move@dwelling                           73   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.58 safe 0.45 held 0.71 all 0.244 food 3.07 hungry-with-food 0.27 | lasted 24/24 extinct 0 mean 174.3 median 110 | phys 0.61 safe 0.52 belng 0.50 estm 0.21
```

## What the last change did

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

## Earlier changes

What the changes before this one did, each measured against the master of its
own day. They are kept for the method rather than for the numbers: none of
them is a comparison with the run above.

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

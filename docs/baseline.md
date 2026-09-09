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
   1  540   268     38    137     92 |  0.70  0.47  0.58  0.47  0.04 |  1.00
   2  290   264     30     76    196 |  0.58  0.47  0.60  0.78  0.09 |  1.00
   3   33   100     40     23     20 |  0.60  0.61  0.74  0.78  0.07 |  1.00
   4  110   362     65     85     79 |  0.52  0.62  0.46  0.64  0.15 |  0.99
   5   38   144     26     19     33 |  0.49  0.55  0.42  0.45  0.06 |  0.96
   6  155   173     30     57     91 |  0.52  0.51  0.60  0.72  0.14 |  1.00
   7  183   264     39     90    105 |  0.60  0.54  0.45  0.67  0.12 |  1.00
   8   57   153     33     21     63 |  0.58  0.52  0.69  0.69  0.06 |  0.98
   9  360   491     28    182     76 |  0.62  0.53  0.49  0.42  0.03 |  0.99
  10  121   195     32     43     82 |  0.51  0.48  0.53  0.61  0.03 |  1.00
  11  395   372     45    141    387 |  0.55  0.50  0.44  0.48  0.02 |  1.00
  12   99   239     51     57     68 |  0.57  0.58  0.63  0.64  0.14 |  1.00
  13   34   152     36     24     43 |  0.60  0.59  0.55  0.80  0.10 |  0.90
  14   15   112     57     11      7 |  0.54  0.50  0.68  0.84  0.10 |  0.53
  15  276   356     48     92    185 |  0.54  0.49  0.41  0.49  0.03 |  1.00
  16   32    76     21      9     35 |  0.67  0.42  0.81  0.88  0.05 |  0.82
  17   38   118     32     16     34 |  0.60  0.49  0.68  0.63  0.10 |  0.90
  18  275   410     26    114    146 |  0.58  0.53  0.33  0.59  0.06 |  1.00
  19  444   479     39    182    185 |  0.61  0.49  0.49  0.48  0.07 |  1.00
  20  298   197     39    119     67 |  0.57  0.52  0.62  0.62  0.13 |  1.00
  21  253   380     32     79    151 |  0.62  0.50  0.55  0.59  0.08 |  1.00
  22   17    97     27      2      9 |  0.55  0.40  0.79  0.77  0.20 |  0.30
  23   94   250     93     53    146 |  0.57  0.63  0.55  0.58  0.06 |  1.00
  24  107   174     26     39    117 |  0.53  0.50  0.41  0.61  0.03 |  1.00

dwell/rest                         3077935  28.5%
take/berries@wood                  2748659  25.4%
consume/provision                  1777293  16.4%
dwell/guard@market                  914321   8.5%
pass/practice>pupil                 881414   8.2%
take/fish@water                     413170   3.8%
dwell/meet@tavern>neighbour         278966   2.6%
take/timber@wood                    155128   1.4%
pass/practice>self                  111966   1.0%
take/grain@field                    109729   1.0%
transfer/provision>needy            100281   0.9%
raise/timber>dwelling@open           55329   0.5%
tend/plant@open                      50051   0.5%
exchange/coin>provision@market       40381   0.4%
exchange/material>coin@market        37086   0.3%
transfer/provision<holder            14683   0.1%
transfer/material>requester          14603   0.1%
make/timber>tool@bench                6282   0.1%
tend/clear@open                       5666   0.1%
strike/person>wrongdoer               5424   0.1%
take/game@wood                        3800   0.0%
tend/water@field                      3429   0.0%
raise/timber>road@ground              2542   0.0%
take/stone@outcrop                    2031   0.0%
dwell/look                             434   0.0%
raise/timber>tavern@open               134   0.0%
make/provision+timber>meal@hearth      130   0.0%
move@dwelling                           83   0.0%
raise/timber+stone>granary@open         58   0.0%
raise/timber+stone>market@open          31   0.0%
make/stone+timber>tool@forge            18   0.0%

born 0.15 inherit 0.05 temp 0.15
gates: fed 0.55 safe 0.53 held 0.83 all 0.281 food 4.89 hungry-with-food 0.27 | lasted 22/24 extinct 0 mean 177.7 median 121 | phys 0.58 safe 0.52 belng 0.56 estm 0.63
```

### What fencing moved

The run above is the tree with hedges round the large holdings; see "Fences" in
docs/action-space.md. Against the batch it replaced - fed 0.57, all 0.313,
lasted 24/24, mean 229.1, median 179 - nothing crossed a threshold: `fed` fell
0.02, `all` 0.032, `lasted` 2 of 24, all inside what the section below shows the
seeds doing on their own. The population numbers fell further than that, mean by
a fifth and median by a third, and an independent batch does not confirm it:

```
offset 24, master: gates: fed 0.59 safe 0.51 held 0.83 all 0.300 | lasted 23/24 extinct 0 mean 213.7 median 191
offset 24, fenced: gates: fed 0.59 safe 0.51 held 0.83 all 0.301 | lasted 23/24 extinct 0 mean 191.6 median 184
```

On seeds the baseline never draws from, the gates are identical to the third
decimal and the median is down 4%. So what the fences cost is at most a slice of
the population on the default draw, and on the evidence of two batches they cost
nothing that can be told from chance.

The three lines in the next section were measured before the fences and have not
been retaken: they are about how far the seeds swing with the code held still,
which is what the thresholds are read off, and that has not changed.

## How much of that is chance

The same code on two further batches of twenty-four seeds, which `-offset`
draws from a stretch of the seed space the baseline never touched:

```bash
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 24 -quiet
go run ./cmd/tune -seeds 24 -ticks 21600 -offset 48 -quiet
```

```
offset  0: gates: fed 0.57 safe 0.56 held 0.84 all 0.313 food 4.53 hungry-with-food 0.27 | lasted 24/24 extinct 0 mean 229.1 median 179 | phys 0.58 safe 0.52 belng 0.53 estm 0.58
offset 24: gates: fed 0.59 safe 0.51 held 0.83 all 0.300 food 5.13 hungry-with-food 0.24 | lasted 23/24 extinct 0 mean 213.7 median 191 | phys 0.63 safe 0.48 belng 0.55 estm 0.63
offset 48: gates: fed 0.55 safe 0.53 held 0.84 all 0.289 food 5.01 hungry-with-food 0.28 | lasted 24/24 extinct 0 mean 193.2 median 186 | phys 0.57 safe 0.52 belng 0.57 estm 0.69
```

Nothing changed between those three but which seeds were drawn, and the mean
population went 229, 214, 193, the median 179, 191, 186. `lasted` was 24, 23
and 24 of 24, and none of the three lost a settlement outright. The mean is the
loosest reading here and always has been: a settlement that runs away is worth
as much as the twenty that did not, and the two seeds above sitting over five
hundred are most of the distance between the draws.

What holds still is the per-agent side. `fed` moved by 0.04 across the three,
the mean needs by 0.03 to 0.11, and `all` sat between 0.289 and 0.313. So:

| reading | moved by chance | worth believing at |
|---|---|---|
| fed, and the four mean needs | 0.00-0.07 | 0.05 on two batches that agree |
| gates all | 0.024 | 0.05 |
| lasted, extinct | 0 of 24, 0 of 24 | 5 of 24 |
| mean population | 8% of itself, 27% before | a second batch that agrees |
| median population | 11% of itself | a second batch that agrees |

A change that only moves the population numbers has not been shown to do
anything. Run it again on `-offset 24` before believing it, and say in the
commit that both batches agreed.

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

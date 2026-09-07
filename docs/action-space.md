# Fit-based action space

Status: all six phases implemented, plus credit by provenance. **Recognition is the default rule.** The value rule stays behind `world.Rules.Fit = false`, or `-value` on the headless runner, and its tests run through a `valueWorld` helper so both rules stay covered. Recognition settlements replace their founders on 21 of 24 seeds with no extinctions (see Robustness below), farming overtakes foraging as fields and farmers improve (see Harvests below), and the land pushes back and is answered (see The land below).

## Why

`system.Choose` is an expected-value maximizer. For every action it multiplies the action's expected need gains by the agent's urgencies and personality, adds conscience, subtracts expected reprisal, divides by time cost, and takes the argmax with a little noise. It works, but it means every agent is a small economist, and the only way to change behaviour is to change payoffs.

People do not choose that way most of the time. They recognise what kind of moment they are in and do what such moments call for. The value of the outcome shapes what they will recognise next time. This document describes a decision rule built on that idea: **choice is recognition, learning is where value lives, and the two never meet in the same step.**

Three commitments, agreed up front:

1. Ranking is by cosine fit between the moment and each action's signature. The overall intensity of the moment only sharpens or loosens the sampling. It never reorders candidates.
2. Signatures ("habits") are per agent, seeded from one shared prior per action. Teaching copies them. Children inherit them.
3. The learning signal counts need changes only. No stock valuation, no pride term. Actions that only set up a later gain learn through credit that follows provenance: a meal thanks the act that grew the food, and rising safety thanks the act that raised the roof or kept the watch. A short eligibility trace also hands part of each lesson to the acts just before it.

## The space

`core/habit` defines a 20-dimensional space. Every coordinate is roughly in [-1, 1] with 0 meaning neutral.

Shared dimensions, computed once per agent per decision:

| dim | name | source | mapping |
|---|---|---|---|
| 0-4 | hunger, unsafe, lonely, unproven, curious | `need.Urgencies(a.Needs)[t] * a.Personality[t]` | `2*clamp01(x) - 1` |
| 5 | food | `Inventory[Food]` | `2*clamp01(food/4) - 1` (same knee as `foodValue`) |
| 6 | wood | `Inventory[Wood]` | `2*clamp01(wood/2) - 1` (the cost of a house, so +1 means "can build") |
| 7 | wealth | `Wealth` | `2*clamp01(wealth/20) - 1` |
| 8 | shelter | `Shelter` | `shelter - 1`, a lack: half a house is half a house short |
| 9 | company | `w.Neighbor(a, 8) != nil` | +1 or -1 |
| 10 | chill | `w.Climate.Chill()` | `chill`, one-sided: 0 through the mild half of the year, 1 at the bottom of a hard winter |
| 11 | order | `w.Safety` | `safety`, one-sided |
| 12-15 | honesty, charity, industry, tradition | `Norms` | `n`, one-sided, frozen |
| 16 | caution | `Caution` | `c`, one-sided, frozen |

Per-candidate dimensions, patched for each action being weighed:

| dim | name | source |
|---|---|---|
| 17 | near | `1 - 2*clamp01(TravelCost(a.Pos, target) / Vigor() / 30)` |
| 18 | rapport | `Anticipate(a, other)` for actions done to a person; sign flipped for retaliate; 0 otherwise |
| 19 | skill | `2*Efficacy[def.Skill] - 1` for skilled actions; 0 otherwise. Belief, not truth. |

Personality is folded into the urgency dimensions rather than kept separate. The moral coordinates and order are one-sided because the belief layer defines them that way: a norm of 0 is holding nothing, caution of 0 is having learned of no reprisal, safety of 0 is nobody keeping order. Read on that scale an ordinary agent minds a wrong about half as much as a saint, which is what conscience charges in the value rule. Centred at 0.5 they fell silent for everyone but the extremes, and theft ran wild (see the comparison below). The moral dimensions are frozen: they are the same across every candidate an agent weighs, so if habits could learn them every habit would soon carry the agent's own norms and the dimension would cancel out of the choice. Values belong to the agent. Habits learn situations.

Dropped on purpose: health (tracks hunger and shelter), tools (sell and craft are hard-gated), market price, knowledge (global and unbounded).

## Signatures and habits

- `action.Def.Prior` is the shared signature of the moment an action belongs to. It is composed by the ontology from what the act is about (see The ontology below) and never learned; the prior each act was tuned to by hand before that is kept as `Def.Tuned` and held against it in a golden test. `Def.Skilled` names the skill a candidate draws on for this agent right now, and `Def.With` names the other agent it involves, so the per-candidate dimensions can be filled in.
- `entity.Agent.Habits[i]` is the agent's own copy for habit slot `i`, which is catalog position, seeded from the prior on first decision (`Imprinted`) and moved by experience.
- `entity.Agent.Reach[i]` in [0, 1] is how far into reach action `i` is for this agent. `Def.Reach0` seeds it.
- Habit length is clamped to [0.2, 2] after every update. Cosine of a zero vector is 0, never NaN.

## Choice

For every action that is `Available` and has a target:

```
eff_i = cos(S_i, H_i) - 0.6 * (1 - Reach_i)
```

Sample from a softmax over `eff` at temperature `τ / intensity`, with `τ = 0.15` and `intensity = 0.5 + Σ_t urgency[t] * personality[t]` over raw urgencies. A moment with strong urgencies is decided sharply. A bland moment is decided loosely. The norm of the whole situation vector is not a usable intensity, because the stock and surroundings coordinates saturate at -1 for most agents most of the time. Exactly one `w.RNG.Float64()` is consumed per decision.

Nothing is divided by cost. Nearness is a dimension the habit learns about. Hard physical gates in `Available` stay (eating needs food). The only soft judgement gate in the catalog today, teach's skill floor, becomes reach.

### Writing priors

Fit is by direction, which has three consequences that phase 3 ran into:

- **A prior should name only the coordinates that predict its moment.** Every extra coordinate dilutes the ones that matter. The first draft put a skill coordinate on farm, build, and guard; since every newcomer believes itself unskilled, that read as a tax on exactly the acts a settlement needs first.
- **There is no constant fallback.** A rest prior of "nothing is urgent" matched a sated agent on four axes and beat every specific act. A prior with one always-maximal coordinate has a constant fit that real but moderate matches lose to. Rest's prior is therefore "recover", a mild version of eating's. Under sampling the least bad candidate is fallback enough.
- **Stock coordinates that sit at -1 for everyone are attractors.** Gather wood fits any agent with no wood and no house rather well, whatever else is going on. That is arguably true, and it is what gets houses built, but it is worth remembering when reading activity tables.

`action.Rank` orders the available candidates by fit and is the ordering oracle for tests: a test asserts which action ranks first for a canonical moment, never what the sampler drew.

## Learning

At every plan end, whether or not `Apply` ran (a plan whose `Available` went false while walking is a lesson too):

```
r   = Σ_t (Needs_after[t] - Needs_before[t]) * Urgency_at_decision[t] * Personality[t]
adv = clamp(r - Baselines[index], -0.5, 0.5)                                  // BaselineMix = 1
Baseline += 0.02 * (r - Baseline);  Baselines[index] += 0.05 * (r - Baselines[index])
Trace.Push(index, S_at_decision)             // keep 3, newest first
for k, step in Trace:
    H[step.Index] += 0.10 * 0.5^k * adv * (step.Situation - H[step.Index])   // learnable dims only
H[index] += 0.003 * (Prior[index] - H[index])                                // retention, all dims
clamp |H[index]| to [0.2, 2]
Reach[index] = min(1, Reach[index] + 0.02)                                   // doing is learning
```

The reward uses urgencies from the moment of the decision, so an outcome is judged by what the agent wanted then. A lesson is judged against the act's own baseline, so it is about *when* the act pays, never *whether*. The machinery for a blend with the agent's general baseline is there (`BaselineMix`) and is set to the act's own alone. With the general baseline in the mix, an act whose direct outcome is always modest, which is every instrumental act and above all the public good of standing guard, is pushed a little further from its own moments every time it happens, and the settlement loses public order and with it the safety that births need. With only the general baseline (phases 4 and 5) every positive reward reinforced and the eat habit drifted toward barely-hungry moments.

### Credit by provenance

Food on hand remembers the act that produced it and the moment it was taken in, as a harvest (`Agent.Larder`, oldest first, capped at 16 harvests). As the food is consumed, by meals, sales, gifts, or a thief, each consuming plan's reward is added to the harvest it came from, and when the last of a harvest goes the act that made it is judged by all it brought (see Harvests below). A roof remembers the act that last raised shelter (`Agent.Roof`), a watch the act by which the agent last kept public order (`Agent.Watch`); safety that rises during any later plan credits both, for the part of that plan's reward safety accounts for.

What the credit carries is an **advantage**, never a raw reward. This was the difference between a working economy and a runaway one. Thanked with the raw reward, a forage got good news from every meal, its habit drifted onto the average moment and fit everything, and agents foraged every six ticks into a larder of thirty units while houses rotted. Thanked with the advantage, a forage that fed a full belly is pushed away from that moment, and production regulates itself: a meal better than meals usually are pulls the field toward the moment it was worked in, a needless one pushes it away.

### Harvests

The owner's brief for farming: bad and slow when first discovered, then improving until it is the mainstay. Recognition has no efficiency channel at the moment of choice, by design, so the only place a field can be found better than the forest is in what its harvests fed. A harvest is settled when its last unit is gone: the act that made it is credited with `Advantage(Returned, Harvest)`, where `Returned` is the need satisfaction everything it fed brought and `Agent.Harvest` is the agent's running sense of what a producing act usually brings. A forage brings one unit and feeds one meal. A first field on ordinary ground brings 0.4 to 0.8 and feeds less, so it is learned away; after agriculture (yield times 1.8) and with the work learned (plus two times skill) it brings two or three, and the same lesson pulls it in. A harvest evicted unsettled, because the larder is full, is judged by what it brought so far, which is how overproduction is learned away too.

Two things had to be true for the arc to show:

- **Tradition has to carry farming through its bad years.** A farm prior with a hunger coordinate gave farming a foothold in hungry moments, and the harvests learned it away before agriculture arrived on either seed (agriculture needs two farmers at skill 0.2, which is twenty farms each). The farm prior is the industrious moment alone, a person with a field nearby working it whether or not the larder is low, and retention toward that prior is what keeps a poor field worked until it is a good one.
- **Skill has to survive a generation.** With the founders dead, their children started at skill zero, and on seed 7 farming fell from 15 acts per agent to one inside 500 ticks. A child is now born with half its parent's skills (`system.InheritedSkill`, under recognition; the value rule keeps its original design), so a farming family stays a farming family. Teaching passes the rest.

With both, on seed 7 farming goes from 7 acts per agent per 500 ticks to 16 while foraging falls from 35 to 18, farming skill reaches 0.8, and the population reaches 127 by tick 6000; on seed 21 farming doubles and skill reaches 0.4 while the forest still feeds most meals, the arc in progress. On the 24-seed sweep this took survivors from 20 to 22 and the median population from 118 to 127.

Guard starts fully in reach. It is the one public good in the catalog, and a settlement that has to discover it first has died of disorder before it does. The value rule's safety turned out to come from public order too, not from houses: its agents guard several hundred times per 500 ticks and hold order at 1.0, while their shelter is as low as recognition's. Long actions carry more decay in `r`, which is the old time cost re-emerging from physics rather than from a formula.

The trace is what keeps the economy alive under a needs-only reward. Farm and forage never touch needs, only the larder. When a later eat pays +0.35, the trace hands half of that to the previous plan and a quarter to the one before. If liveness runs still show farming being learned away, the documented fallback is to add stock terms to `r`. That would move a value judgement into learning, which is the agreed place for it, but it has been rejected for now.

### Second tuning pass and the subsistence finding (phase 6)

Flipping the default broke one test: on seed 21 nobody posted a request in 3000 ticks. Diagnostics showed why. Wealth stayed at zero because nobody sold; nobody sold because food per agent never rose above about half a unit; and that was because the farm prior said "hungry and out of food", so agents farmed only when hungry and foraged the rest of the time. Three changes, all inside the design:

- **The farm prior no longer mentions hunger.** Foraging is what hunger calls for. Farming is what an industrious person with a field nearby does, whether or not the larder is low. That is the only moment in which a larder ever fills past today.
- **Wood is measured against the cost of a house**, so "enough wood" reads as +1 exactly when a shelter can be built, and the gather and build priors are worded around the unsheltered moment.
- **Per-action baselines** (above), so that a good outcome for an act is judged against that act's usual outcome.

After these, seed 21 reaches 41 fulfilled requests per 500 ticks by tick 2000 and wealth grows; the social seed keeps its request rate (276 asked, 230 fulfilled) with theft at 293 and giving at 222.

What did not move is safety, and with it growth. Agents forage every 17 ticks and gather wood a tenth as often, houses rot faster than they are rebuilt, and mean safety sits near 0.3 against 0.5 under the value rule, below the 0.6 a birth requires. The cause is structural. Under the leaky hierarchy a moderately hungry agent has its safety urgency damped, so it chases food; recognition has no notion that a field is more efficient than the forest, so it chases food the slow way; and a needs-only reward gives no credit for surplus, so nothing it learns changes that. **Recognition with a needs-only reward produces a subsistence society: fed, housed after a fashion, social, literate in time, and not growing.** Whether that is a bug or the point is a design call. The two levers left, both rejected so far, are an efficiency or stock term in the reward, and a longer credit horizon (a trace of 8 at 0.75 was tried and made requests collapse, because it spread each meal's credit over everything).

### Robustness

Single seeds are a coin toss near the edge. Whether a settlement lasts turns on how many births fall in its founders' fertile years, and that turns on safety crossing 0.6 at the moment of a roll, so the same change can send one seed from 0 to 144 and another the other way. Ranking a change on one or two seeds is meaningless; the method that worked is a batch of 24 seeds, in parallel, counting settlements that replaced their founders (population at least 20 at tick 6000), extinctions, and the median population, confirmed on a second independent batch:

```bash
go build -o /tmp/h.exe ./cmd/headless && seq 1 24 | xargs -P 8 -I{} sh -c '/tmp/h.exe -seed {} -ticks 6000 -every 6000 | grep "^  6000" | awk -v s={} "{print s, \$2}"' | sort -n
```

| variant | seeds 1-24 | seeds 25-48 |
|---|---|---|
| committed after provenance | 16 survived, 2 extinct, median 42 | 19 survived, 0 extinct, median 43 |
| **shelter read as a lack** (adopted) | 20 survived, 0 extinct, median 118 | 20 survived, 0 extinct, median 77 |
| plus unsheltered gather and build priors | 21 survived, 1 extinct, median 78 | |
| plus eat at milder hunger | 16 survived, 2 extinct, median 94 | |
| order with a midpoint, guard prior on it | worse on every seed tried | |
| guard prior less repelled by order | 4 extinctions in 8 seeds | |

What the winning change is about. Safety urgency is gone by the time safety reaches 0.4, a birth needs 0.6, and a house rots at 0.002 per tick. Read around a midpoint, the shelter coordinate only called for gathering and building once a house had mostly rotted, so shelter sat near 0.2 in both rules and safety only crossed 0.6 in the spike after a rebuild. Read as a lack, half a house is still a moment that calls for wood, houses are kept nearer 0.5, and the spikes are no longer what a settlement lives or dies on.

What did not work is as telling. Anything that made disorder a louder call to stand guard, whether through the mapping or the guard prior, starved settlements at their posts. The milder eat prior, meant to keep agents above the birth mark, undid the gain. And seed 21, the seed that started this, is still on the edge: 13 before, 10 after, 63 under a variant that was worse overall. It is not a special seed, it is an ordinary one that fell on the wrong side of a knife edge, and the fix was to blunt the edge for everyone rather than to tune for it.

### First side-by-side run

Same seeds, same populations, same tick counts; the two modes are different RNG streams so this is a comparison of character, not of trajectories. Phase 4, before any tuning.

| test | value mode | fit mode |
|---|---|---|
| seed 7, 20 agents, 6000 ticks | pop 40, 4 starved, 36 houses, 3 techs | pop 22, 0 starved, 22 houses, 3 techs |
| seed 31, 25 agents, 6000 ticks | 148 asked, 134 fulfilled, 9 stolen, 84 given | 105 asked, 103 fulfilled, 510 stolen, 39 given |
| seed 1, 20 agents, 4000 ticks | 7 stolen, 31 avenged, 1 feud | 397 avenged, 45 feuds |

After the first tuning pass (moral coordinates one-sided, phase 5):

| test | fit mode, tuned |
|---|---|
| seed 7 | pop 18, 2 starved, 20 houses, 3 techs, mean physiological 0.71 |
| seed 31 | 216 asked, 202 fulfilled, 310 stolen, 214 given, 298 avenged |
| seed 1 | 94 avenged, 4 feuds |

What the numbers say:

- **Nobody starves, but nobody thrives.** Mean physiological need sits near 0.55 in fit mode against 0.85 in value mode, and the population barely grows. Agents under recognition satisfy the pressing need and stop; value maximisers overshoot into surplus, which is what feeds births. The eligibility trace did keep farming and foraging alive with a needs-only reward.
- **Theft was fifty times more common, now thirty.** Before tuning the moral coordinates were centred so that a norm of 0.5 read as 0, which made them silent for the average agent, whereas value mode charges everyone conscience in proportion to their honesty. Reading them one-sided, as the belief layer defines them, cut theft by two fifths, multiplied giving by five, and collapsed feuds from 45 to 4 on seed 1. What remains is honest dynamics: stealing works, the trace hands the meal's reward back to the theft, and guilt lands as an esteem loss that a hungry agent barely weighs. A society with a self-centred reward and no enforcement steals. The remaining levers are the steal prior itself and how hard remorse lands, both left for phase 6.
- **Feuds cluster and persist**, which is the predicted consequence of habits being individual: once an agent has learned that a grudge calls for getting even, it keeps recognising that moment.
- **Requests are fulfilled at the same rate**, so the contract layer works under recognition without changes.

## The land

The world was static ground the agents drew on without limit. Now it pushes back, and a settlement under pressure finds new ways to live. `core/action/land.go`, `core/system/land.go`, `core/system/discover.go`.

**What the land has to give.** A forest tile carries wild food (`Tile.Wild`) that foraging takes a little of and hunting a lot; a water tile carries fish (`Tile.Fish`); a field carries fertility that each harvest wears down toward a floor, and `Tile.Rich`, the most it can recover to when left fallow. All of it comes back slowly, faster once the settlement has learned forestry. A picked forest still gives something, so a settlement is pushed toward the river and the field rather than into the ground.

**Four answers**, each far out of reach until discovered:

| action | belongs to the moment | takes | gives |
|---|---|---|---|
| fish | hungry by the water, skilled at it | fish from the best water beside the bank | food, fishing skill |
| hunt | hungry with a tool in hand | wild food, three times a forage; wears the tool | more food than a forage from a full forest |
| irrigate | an industrious farmer with wood to spare and water within eight tiles | a unit of wood | a quarter more the field can hold |
| plant trees | no wood, and a mind for those who come after | two ticks | a young forest tile, feeding nobody today |

Fishing is a new skill; teaching and serving pass it like the others.

**Discovery is a response.** Each tech needs the settlement to be under the pressure it answers, so different worlds take different paths:

| tech | pressure | knowledge | opens | effect |
|---|---|---|---|---|
| fishing | the forest the settlement lives off is picked thin, and there is water near | none | fish | fish yield ×1.5 |
| trapping | the forest is thin and two people hold tools | none | hunt | hunt yield ×1.5 |
| irrigation | agriculture known and the fields near the market have gone poor | 10 | irrigate | farm yield ×1.2 |
| forestry | less than six tenths of the founding forest is left | 10 | plant trees | regrowth ×2 |

"The forest the settlement lives off" is the two dozen forest tiles nearest the market. Foragers go to the nearest forest, so those few tiles are picked bare while the woods beyond stay full, and it is those that say whether the settlement feels the land pushing back. Measured over the whole neighbourhood the forest never thinned and nothing was ever discovered.

Under the value rule the land actions are unavailable until their tech is known or the agent has come within reach of them on its own; the value rule has no reach gate, and without this a starving value-mode agent walked to a river nobody had fished.

**Three things that broke and what they taught.** The first version killed every settlement on every seed, under both rules, and the forest was not the cause: wild food near the market never fell below three quarters. Foragers were sent up to ten tiles for a fuller patch, and that walk each way, on the errand the whole economy runs on, cost more than the fuller patch gave; foragers now go to the nearest forest, thin as it is. Meals had to be whole units, and with fractional yields agents starved holding most of a meal; a meal is now as much as it takes to be full, up to three units, or what there is. And field wear at 0.02 a harvest wore the value rule's fields, farmed without rest, to the floor within a generation; it is 0.006, with fallow recovery at 0.0006 a tick.

Study starts nearer to reach (0.6, from 0.4). Recognition settlements otherwise rarely studied, and every discovery waited on knowledge they did not have.

**What the runs show.** Over 6000 ticks: seed 3 learns fishing at tick 200, before it has learned to farm; seeds 2, 7, and 16 farm first and turn to the river between ticks 1500 and 2100, once the forest they live off is thin, with hundreds to fifteen hundred fishing acts following. Trapping arrives with fishing wherever tools are held. Irrigation and forestry did not fire on these seeds: with wear at 0.006 the fields near the market stay above the 0.45 that counts as poor, and forest reseeding more than replaces what is cleared, so the founding forest is never down by four tenths. Both remain reachable by pioneers through study and teaching, and are used that way (fifty to a hundred irrigations and several hundred plantings per run), but as discoveries they wait for a scarcer world or a larger settlement. On the 24-seed sweep the package leaves survival at 21 of 24 with no extinctions and a median population of 127 to 144.

## Making and keeping

The longer chain that turns what the land gives into things that last. `core/action/making.go`.

**Two new goods.** Stone, cut from rock outcrops the map now carries, goes into houses (a stone in the walls makes a house half again as good) and granaries. Meals are cooked food: they keep (spoiling at a third of the rate) and feed more (0.5 a unit against 0.35 raw), and a hungry agent eats them first. Tools now make a field go further, a quarter more per harvest, and wear with the work.

| act | belongs to the moment | takes | gives |
|---|---|---|---|
| cook | a full larder, wood to spare beyond a house's worth, no hunger | two units of food and a fifth of a unit of wood | two meals |
| quarry | the unsheltered with a tool in hand and stone near | tool wear | stone, building skill |
| build granary | standing to win, stone and wood to spare, a settlement one means to stay in | three stone and two wood, beside the market | the market's food spoils half as fast, for everyone |
| smelt | a skilled crafter with stone and fuel | a stone and a unit of wood | twice a crafting's tools |

| tech | pressure | knowledge | opens | effect |
|---|---|---|---|---|
| pottery | food piling up at the market | 30 | cook | market spoilage ×0.7 |
| quarrying | masonry known and rock near the market | 40 | quarry | build efficiency ×1.2 |
| masonry | (existing) | | build granary | |
| metallurgy | (existing) | | smelt | |

**What the runs show.** Pottery arrives between ticks 1100 and 1900, quarrying with masonry soon after; a settlement cooks and quarries a few hundred times over a run, one in four builds a granary, and smelting stays rare because metallurgy waits on five tools at the market. On 96 seeds the package holds survival at 82 against 86 for the code before it, with the median population 144 against 147. Merged with the settlement laying its own roads, the same 96 seeds give 92 survivors, no extinctions, and a median of 231.

**What broke, and the lesson about measuring.** The first version cost survival badly, and finding out why took most of the work. Cooking burned half a unit of wood per meal, four meals to a house, and pottery came early enough that whole settlements learned to cook and stopped building. But no 24-seed sweep could show that, because a child inherits its habits with random drift on every coordinate of every catalog action, so a bigger catalog draws more from the world's random stream at every birth and every trajectory after the first birth is rerolled. Variants that changed nothing behavioural moved the 24-seed sweep by three survivors and halved its median. The comparisons that settled it were 96 seeds each: 86 survivors before the package, 68 with it, 84 with cooking removed, 82 with cooking made a hearth batch on spare wood. Anything that grows the catalog has to be judged on that scale.

## Places

A social act needs two things the old catalog did not ask for: somebody within reach, rather than anybody alive, and somewhere to meet. `core/action/places.go`.

- **Company is within twelve tiles.** Socialising and teaching are unavailable otherwise, and the companion an agent picks is chosen among those within reach, so nobody crosses the map to see somebody.
- **Meetings happen at places**: the market, a house, a granary's yard, or a tavern. The place is chosen by temperament. The warm head for a tavern, then the market, then wherever the companion already is; the cool would rather have company at their own house, then the companion's. If there is no place within six tiles of the companion, the meeting cannot happen: they are off in the woods. Across four seeds every one of tens of thousands of meetings took place at a market, a house, or a tavern, and none elsewhere.
- **Taverns.** Brewing answers a settlement big enough to be lonely in with grain to spare (twelve people, five food at the market, knowledge 25) and opens building one, for three units of wood beside the market, one per settlement until it outgrows it. A meeting in a tavern is a better evening for both sides. Every seed builds its tavern within a few hundred ticks of brewing, and between a sixth and a half of all meetings then happen there. At four units of wood none was ever built: recognition agents seldom hold that much, since gathering yields a unit and a half and a house takes two.

On 96 seeds, asking social acts to have somewhere to happen took survivors from 92 to 95 and the median population from 231 to 287; the tavern's cost brought the median back to 237.

**Workplaces.** Making things needs somewhere to make them, as meeting needs somewhere to meet. Before, an agent without a house crafted, cooked, smelted, and studied wherever it stood, forest included.

| act | place |
|---|---|
| craft | a bench: at home, or a stall at the market |
| study | a desk: at home, or at the market where the records are; not the tavern |
| cook | a hearth: at home, or at the tavern |
| smelt | a forge: at home, and nowhere else |
| guard | a market within the settlement's reach with somebody at it; an empty square is nothing to guard |

**Comfort, and what it cost.** The brief was that resting and eating should prefer home or a tavern too. They cannot. Recognition reads distance as poor fit, a meal is the loop the whole economy runs on, and the year has made the margins thin. Sending a meal or a rest even one step toward a roof, at a house tile's toll, cost a tenth of settlements over two independent batches of 96 seeds under the seasoned year (61 and 60 survivors against 70 and 73, extinctions 10 and 17 against 10 and 10, median 97 and 42 against 150 and 162); at four tiles it had killed two thirds before the year came. Nor can a rest be made better under a roof: a rest that restored more at home reinforced an idle act and the median fell by a third (101 against 150). What remains is the table: a meal eaten in company at the tavern is a little belonging, which costs nothing (71 survivors, median 132). Agents are at home for their crafting, cooking, and studying, and in the tavern for their meetings; where they eat and rest is where they are.

Worth recording how this was found, because it is the first case here of two changes that each measure clean and are wrong together: comfort was measured in a world with no winter, the seasons were measured before comfort existed, both passed their own sweeps, and their merge killed seed 3 of `TestCityDevelopsByRecognition` outright - twenty founders and not one birth in six thousand ticks. The liveness sweeps catch this only when they are run on the merge.

The settlement's reach is twenty tiles from the market. Everything else already had its place: the field, the forest, the bank, the outcrop, the market, the companion's side, the requester's door. On 96 seeds the change leaves survival within the band, 92 against 94, and the median population at 356.

## Reach

Implemented in `core/action/reach.go`; the constants live there.

- `Reach0` per action. Everyday living (rest, eat, forage, farm, gather, build, sell, buy, guard, socialize, give, steal, retaliate, fulfil) starts at 1. Gated: craft 0.5, lay road 0.5, teach 0.3, study 0.6, fish 0.4, hunt 0.5, irrigate 0.3, plant trees 0.3.
- **Study broadens.** Every gated action comes closer: `Reach += 0.03 * (1 - Reach) * Mods.StudyRate`, so written records make study widen reach twice as fast.
- **Teaching passes recognition on.** The student's reach for the taught skill's action becomes `max(own, 0.6 * teacher)`, and its habit moves 0.3 of the way toward the teacher's. `ForSkill` maps farming, building, crafting, scholarship, and guarding to farm, build, craft, study, and guard.
- **Discovery opens.** A `Discovery` has an `Opens` list; masonry opens craft, writing opens study and teach, metallurgy opens craft and guard. Masonry opens laying roads too. Each raises `world.ReachFloor` for that action to 0.8, and every agent is lifted to the floor at its next decision.
- **Children inherit.** A child takes its parent's habits with N(0, 0.05) drift on the learnable coordinates and `Reach = max(floor, 0.7 * parent)`. Under the value rule the copy is exact so that rule draws nothing extra from the RNG.
- Doing an action raises its own reach by 0.02 in the learning step.

Reach is the "distance gate": a far action is one whose signature the agent cannot yet reach, and study, teaching, and discovery bring it closer. Because it is a penalty on fit rather than a lock, a curious agent can occasionally reach a far action early. Pioneers fall out of the sampling.

## Expected impact on the simulation

- **Roles.** Agents keep doing what fitted before, so farmers, scholars, and thieves become individuals rather than a population-wide threshold. Feuds cluster around people.
- **Slower reaction to acute need** until temperature and learning rate are tuned. Intensity sharpening is the safety valve. Watch starvation counts; seed 7 today has 4 deaths in 6000 ticks.
- **Techs later.** Study is gated and knowledge only comes from study and scholar service.
- **Culture.** Teaching and births transmit habits, so settlements diverge across seeds more than they do now.
- **Determinism** holds: habit tables are arrays, one RNG draw per decision, fixed catalog order. Fit mode and value mode are different RNG streams for the same seed, so comparison is statistical, not trajectory-level.
- **Player.** `sim.Intend` must build plans through the same builder, so player commands teach the player's habits.

Metrics in `observe.Snapshot`: `HabitSpread` (mean distance of each agent's unit habit from the population's mean unit habit, over all actions; 0 means everyone recognises the same moments the same way), `MeanReach`, `GatedReach` (over the four gated actions), `ChoiceEntropy` (mean entropy in nats of the tick's sampled decisions; 0 in value mode), `Deaths` (cumulative). Headless prints them as `died`, `reach`, `sprd`, `open`; the TUI has one line for them.

## Pitfalls recorded

- The catalog used to be assembled by two `init()` appenders in different files, which run in filename order and put retaliate before steal. Phase 0 replaced them with one literal and `action.Count`, checked at init.
- Rewards are misattributed when other agents change your needs mid-plan (theft, retaliation, encounters). Accepted as noise, which is why the learning rate is small and the baseline slow.
- Only the chosen action's habit moves toward the situation, so winners drift toward "the moment I am usually in" and win more. Retention toward the prior, the norm clamp, and sampling instead of argmax all push against lock-in. `HabitSpread` makes it visible.
- `world` cannot import `action`, so habits are imprinted lazily in `system` on an agent's first decision.

## Phases

| phase | scope | status |
|---|---|---|
| 0 | One catalog literal, `Count`, `Index`, new `Def` fields | done |
| 1 | `core/habit`: space, cosine, sampling, update, ledger, trace, tests | done |
| 2 | Agent `Habits`, `Reach`, `Baseline`, `Trace`, `Imprinted`; Plan `Index`, `Situation`, `Before`, `Started`; `world.Rules` | done |
| 3 | `action.Shared`, `action.Situation`, `action.Candidates`, `action.Rank`, `action.Imprint`; priors and `Reach0` for all 17 actions; canonical-moment ranking tests | done |
| 4 | `system.Recognise` sampling in `Decide`, `system.Learn` at every plan end in `Act`, `system.Commit` as the one plan builder (used by `sim.Intend`), headless `-fit` and `-temp`; fit-mode determinism and liveness tests | done |
| 5 | `action.Broaden`, `action.Pass`, `Discovery.Opens` and `world.ReachFloor`, `action.Inherit` at birth; `HabitSpread`, `GatedReach`, `ChoiceEntropy`, `Deaths` in snapshot, headless, TUI; moral coordinates one-sided | done |
| 7 | Credit by provenance: larder, roof, and watch; credit carries the advantage; per-action baseline alone; guard fully in reach | done |
| 8 | Robustness: 48-seed sweeps; shelter read as a lack | done |
| 9 | Harvests: producing acts judged by all they fed; industrious farm prior; children inherit half their parents' skills | done |
| 10 | The land: wild food, fish, field wear and fallow; fish, hunt, irrigate, plant trees; fishing, trapping, irrigation, forestry discovered under pressure; meals sized to hunger | done |
| 11 | Making and keeping: stone and meals; cook, quarry, build granary, smelt; pottery and quarrying; tools on the farm; stone houses; 96-seed comparisons | done |
| 12 | Places: company within reach, meetings at market, house, or tavern by temperament; brewing and taverns | done |
| 13 | Workplaces: a bench, a desk, a hearth, a forge, and a market worth guarding | done |
| 14 | Comfort: tried and mostly withdrawn; a meal in company at the tavern is a little belonging | done |
| 6 | Recognition is the default (reverted once after the aging merge, restored with provenance); headless `-value`; value-rule tests run through `valueWorld`; recognition twins at full length; ordering twins for every value-rule choice test in `core/action/situation_test.go`; per-action baselines, industrious farm prior, wood knee at the house cost, guard reach 0.8 | done |

Tests under fit mode assert ordering (which action ranks first), not the sampled outcome. `TestHungerEventuallyOverwhelmsPrinciple` is about magnitude and stays value-mode only. The four liveness tests run in both modes from phase 4 onward so tuning is visible before the default flips.

## Laying roads (the second public good)

`Pave` ("lay road", catalog position 13) is the first action added after the six phases. It is the trigger for the road material that already existed: nothing lays streets on the settlement's behalf any more, agents decide to.

**The ground remembers.** `Tile.Traffic` rises by 1 whenever an agent steps onto a tile and fades by 0.5% a tick, a memory about 140 ticks long. It costs nothing to walk a worn tile - wear is not a road - it is only a record of where the settlement's errands actually run. `Grid.Busiest` returns the most-worn *pavable* tile within a radius, so houses and claimed fields are never offered: roads form in the gaps between buildings, which is what streets are.

**The moment.** `Pave`'s prior is the settled one: `Wood 0.6, Shelter 0.7, Company 0.6, Charity 0.5, Industry 0.6, Near 0.5`. Shelter is what separates it from gathering and building, which want the opposite - you improve the common ground once your own roof is up. Charity and industry are what it shares with standing guard. It mentions neither hunger nor skill, for the reasons in "Writing priors".

**Why it is not an esteem farm.** Paving has no provenance channel: a road pays back as slightly cheaper walking, forever, for everybody, and a needs-only reward would extinguish it. So like guard it pays the layer in standing - but deliberately less per tick than guard does. At esteem 0.05 / belonging 0.03 agents paved a third of the map and the settlement was worse for it (seed 5: 984 road tiles, pop 135). At 0.03 / 0.02 paving is common but self-limiting.

**Tuning.** `wornEnough` was measured, not guessed. Equilibrium wear on a tile is roughly crossings-per-tick x 200, so the threshold is what decides whether roads are a big-city luxury or an ordinary act. Sweeping it on seed 5 (6000 ticks, 25 founders):

| `wornEnough` | pop | roads |
|---|---|---|
| 6 | 135 | 984 |
| 20 | 222 | 973 |
| 40 | 332 | 245 |
| 80 | 400 | 170 |
| 150 | 305 | 60 |

At 80 the threshold was population-sensitive - only the largest settlements ever wore ground that far, and seeds 7, 11 and 21 laid 15, 26 and 2 tiles. Dropping the payoff to 0.03/0.02 and the threshold to 45 gave paving on every seed tried, and against the same seeds with no paving action at all (comparison is statistical; adding an action re-rolls the stream):

| seed | without | with |
|---|---|---|
| 5 | pop 170, 111 houses | pop 391, 181 houses, 402 roads |
| 7 | pop 64, 80 houses | pop 163, 117 houses, 181 roads |
| 11 | pop 91, 87 houses | pop 182, 139 houses, 179 roads |
| 21 | pop 132, 122 houses | pop 163, 103 houses, 103 roads |
| 31 | pop 92, 77 houses | pop 97, 77 houses, 75 roads |
| 44 | pop 4, 29 houses | pop 3, 29 houses, 27 roads |

Equal or better on all six, which is the first change in this document to move the subsistence finding rather than work around it. The mechanism is the one the house toll exposed: recognition reads distance straight off the ground, so making the settlement cheaper to cross improves the judgement of everyone in it.

**Where a street belongs.** A tile's case for a road (`Grid.Draw`) is its own wear plus the wear on neighbouring ground that could never be a street: houses, the market, claimed fields. Most of the traffic a street carries is not on the street but on what it runs between, so read on its own the gap beside a thronged doorway looks like empty ground - and in a close-built settlement that left nowhere at all worth paving. Two things deliberately lend nothing. Open ground speaks only for itself: when it lent too, every tile near a busy one read as busy and paving came out in patches instead of lines. Roads lend nothing either, because traffic already on a street is already served, and counting it paved the settlement outward from its first road until the ground ran out (seed 11: 1350 tiles, near half the map).

This is what brought the value rule back. With `wornEnough` at 60, over six seeds and 6000 ticks:

| | recognition, total pop | value rule, total pop | value rule, seeds that paved |
|---|---|---|---|
| own wear only, threshold 45 | 999 | 225 | 1 of 6 (2 tiles) |
| neighbours lend, threshold 60 | 877 | 386 | 6 of 6 |

The value rule improves on every one of the six seeds and lays roads on all of them. Recognition's total falls 12%, which is inside this system's noise - two of the six seeds improved, and the same configuration swings between 19 and 350 across seeds - but it is a fall, and it is recorded here rather than rounded away.

**Cost.** `Busiest` scans a 25x25 window and reads nine tiles per candidate, once per deciding agent per decision, and it is now the most expensive thing in a tick: 2000 ticks of 40 agents went from 650ms to 917ms across the machine. It sits in the parallel phase, so the machine absorbs it. If it ever needs to be cheaper, `Draw` can be computed for the whole grid once per tick in `Weather` instead of per candidate.

### Bridges

A road may be carried over water, and then it is a bridge: the tile stays a river to look at and to fish in, and costs a road to cross. It was added because the settlement straddles its river - the ground worth farming is the ground near the water, so a quarter of the houses end up on the far bank - and with no way to span it those people waded, for ever. It also cut the road network in two, since a way that cannot cross water can only run along its own bank.

Three things had to be true before a bridge was ever built, and each was found by measuring rather than by reasoning:

- **A crossing has to come before a street.** The busiest ground in a settlement is always a lane between houses, never the ford, because everyone who can avoid the water does. Competing on wear alone, a bridge is never the best site and never gets built. `paveSite` therefore looks for an affordable ford first: a street can go round what is in its way, and a river cannot.
- **A bridge cannot cost more than a house.** It was first set at three lengths of timber against a road's one. Nobody in any settlement ever holds three: agents gather toward the roof they want and spend it the moment they have enough, and the most anyone was ever seen holding was two and a half. The action could not fire at all. It costs two now, the same as a house.
- **The water must stay cheap.** Making the river dearer to wade is the obvious answer and it is the wrong one. At 5, 7 and 9 the wading barely fell and the population dropped by up to a quarter: a river nobody can afford to cross is a river nobody wears a ford in, and a ford nobody wears is a ford nobody bridges. The cheapest water is what gets a bridge built.

Six seeds, 6000 ticks, 25 founders, against the same seeds with no bridges:

| | population | share of agent-time spent wading |
|---|---|---|
| no bridges | 1277 | 2.05% |
| bridges | 1535 | 0.63% |

Better on both counts, which is unusual for a change made for the look of the thing.

### The land underneath

The map used to be a sine wave with a river drawn along it and fertility measured as distance from that river. It is now a piece of ground, and everything else is read off it. `core/world/relief.go` holds the whole of it, and the order is the one a landscape obeys:

1. **Raise** the ground — five octaves of smoothed random lattice, scaled to `Relief` (60 m over a map).
2. **Fill** every hollow to the level at which it would spill, by priority-flood inward from the edges, so no ground is left with nowhere to send its water.
3. **Drain** — each tile's water goes to its lowest neighbour, and `Flow` is the share of the map passing through it. Settled highest-first, so a tile's own total is complete before it is passed on.
4. **Carve** — the wettest `waterShare` of the map is the river. Nothing about the water is drawn; it is where the water went.
5. **Height above drainage** (`Tile.Drain`) — how far a tile stands above the water it drains into, got by following its flow down and adding up the fall.

`Slope`, `Aspect` and `Sunlight` are read off `Height` on demand. Woods, outcrops and soil are then scored and thresholded against each map's own distribution rather than against fixed numbers, because a fixed cutoff gives one map a river and the next a puddle: over a handful of seeds the heaviest-draining tile carried between a fifth and four fifths of the map.

**Drain, not flow, is what soil moisture means.** The first version read fertility off flow accumulation and produced a dead world - mean fertility 0.17, essentially no farmland, and three settlements in five collapsed. Flow is a terrible proxy: a tile on the valley floor beside the river carries hardly any flow of its own and is still a water meadow, while a tile halfway up a hillside may carry a gully's worth and be dry as a bone. Reading it off height-above-drainage instead gave mean fertility 0.33-0.45 and 435-747 good tiles per map.

**Walking answers the ground.** `Grid.StepCost(from, to)` adds `Climb` per metre of ascent and `Descend` per metre of fall to the cost of the tile entered, so routing rounds the shoulder of a hill rather than going over it - and since roads are laid where the ground is worn, the streets follow the contours and the valley floors without anybody deciding they should.

**What it cost.** Six seeds, 6000 ticks, 25 founders, against the flat map:

| seed | flat | with relief |
|---|---|---|
| 1 | 255 | 65 |
| 3 | 400 | 359 |
| 5 | 400 | 203 |
| 7 | 304 | 20 |
| 11 | 272 | 399 |
| 21 | 98 | 334 |
| **total** | **1729** | **1380** |

Down a fifth overall, but not uniformly: seeds 11 and 21 grew where they had struggled, and seed 7 nearly died where it had thrived. That is the change doing what it is for - the ground now has quality, and a valley is worth more than a hillside. Raising the fertility floor to lift the weak maps was tried at 0.25 and 0.35 and made the total worse (1311, 1096), because it flattens the very differences the good settlements are living on.

Left for later: nothing erodes yet, and `Flow` is a static share rather than water with a season to it. Both are why the drainage is derived rather than drawn - re-run the four steps on changed ground and the rivers move by themselves.

### Weather

`core/world/erode.go` runs every `ErodeEvery` ticks and is the only thing that changes the shape of the land after the map is made. An age of weather strips soil in proportion to the root of the water crossing a tile, times its steepness, times how little is holding it down; carries what it strips downhill; and lays it down where the water slows. Then the four generation steps run again - fill, drain, carve, height-above-drainage - so the rivers are wherever the new ground sends them. Nothing is moved by hand.

**The settlement causes it.** `hold` is the share of soil that stays put, by land cover: woods 0.25, grass 0.6, ploughed field 1.0, rock 0.15, anything roofed or paved 0. Measured on one map with every slope above the flood plain clothed one way or the other, over forty ages: **ploughed slopes lost 3.7x the soil wooded ones did**, and silted nearly twice as much into the valley. Nobody decides that; it falls out of where the fields were put.

Three things had to be right, and none was right first time:

- **The root, not the amount.** With erosion in proportion to flow, the valley floor carried so much of the map's water that it scoured itself out - the opposite of what a flood plain is. `sqrt(flow) * slope` is the usual reading of stream power and with it the channel still cuts down while the ground beside it fills.
- **Rivers spill.** With silt deposited only where it was carried, it stayed in the channel and the flood plain slowly washed away instead of being fed. `Overbank` puts 0.7 of what a river lays down onto the low ground either side, which is what a flood plain actually is: not ground the river spared but ground the river made.
- **Channels do not flicker.** `carve` will not run a river through anything built or claimed, so as the drainage shifted, a settlement holding the ground the new course wanted wiped out its own river - the old bed dried the moment it fell under the line and the new one could never form. Water now needs the full threshold to appear and keeps its bed until less than half of that remains.

**What it costs.** Ten seeds, 6000 ticks, 25 founders: 2627 people without weather, 2423 with, a fall of 8%. Two seeds are much the poorer for it and one is better. Faster weather was tried and is worse in a way worth recording - at twice this rate the total drops to 1910 and two settlements collapse outright, because the loop above is real: a people that farms its slopes strips them, and the yields it depends on go with the soil. Erosion adds 4-18% to a tick.

Over 6000 ticks a map moves about 1.6 m of ground on average, and around 170 tiles per map change between land and water - the rivers really do shift.

**Not yet joined up:** the weathering runs at an average year's rate and takes no notice of the season, though there is now a year to take notice of (see the climate above). `Tile.Flow` is a share of the map rather than a volume of water, so there are no floods - and a flood is where most of a century's erosion actually happens. Reading the age's rate off `Climate` is the obvious next thread to pull.

## Watching one agent decide

Everything above is measured over settlements, and a settlement is the one
thing a fit-based choice cannot be read off. A decision here is a draw from a
softmax over every action available in the moment; what it leaves behind is a
plan, and a plan is only the winner's name. Tuning a prior meant guessing
from the aggregate which candidate it had beaten, and the map — hundreds of
figures walking — shows the choices and not one reason for any of them.

So the world will follow one agent at a time (`World.Watch`). While it does,
every decision that agent makes is kept whole (`world.Deliberation`): each
action it had before it, where it would have been done, how well it fitted
the moment, how far into reach it still was, and the odds the draw actually
ran on — `habit.Softmax` over the same fits `habit.Sample` drew from, at the
same temperature, which is `Rules.Temperature` divided by how pressing the
moment was. The value rule is watchable too and says so, since a score per
tick is not read the way a fit is.

Three things kept it from being a change to the simulation rather than a
window on it:

- **One agent.** Keeping this for everybody would cost every tick something
  to be read for nobody, and following somebody else forgets the last one.
- **Written on the way past.** Deliberations ride home from the parallel
  deciding phase inside `decision`, beside the plan and the entropy, and are
  filed in agent order. Watching an agent must not decide what a run does or
  when it does it, and `TestWatchingChangesNothing` holds a watched run
  against an unwatched one.
- **Nothing inside may read it.** No agent knows what anyone weighed,
  including itself. It is for the operator, and `observe.Look` — a whole
  agent, including what it can do against what it believes it can do — is
  frankly omniscient in the same way the rest of the watch command is.

In `cmd/watch`, tab or a click picks a figure out of the crowd and opens it up
beside the map; the last decision is shown as the candidates it beat, with
the odds, over the run of choices behind it. A run of choices is what says
whether an agent is getting anywhere or turning on the spot between the same
two errands — which is the thing the aggregate graph, by construction,
averages away.

## The turning year

The world had one weather and kept it for ever. Now it has a temperate year, and the settlement has a season to get through rather than a steady state to sit in. `core/world/climate.go`, `core/system/climate.go`.

**The shape of it.** A year is 100 ticks, so a life is some fifty years and a run long enough to develop sees dozens of winters. Temperature is a sine about a mean of 10 degrees with a swing of 12, putting midsummer at 22 and midwinter at -2, and two AR(1) wanderings sit on top of it: a slow one, eight years to shed an anomaly and a standard deviation of 1.2 degrees, which makes a decade kind or unkind, and a fast one, six ticks and 2.5 degrees, which makes the warm week and the cold snap. Tick zero is early spring, so founders arrive with a growing season in front of them rather than behind them. It costs two draws from `World.RNG` a tick and nothing else, so a seed and a command log still reproduce a run exactly.

**Three things read it**, and the whole of the feature is in the three:

| reader | what the season does |
|---|---|
| `system.Land` | forest, wild food, fish, fallow and reseeding all keep the season's hours. `Climate.Growth()` is two thirds of the old flat rate in the depth of winter and a third again as much at midsummer, and averages exactly 1 over a year, so the land tuning that came before the seasons still holds: what changed is when a forest grows, not how much it grows in a year. |
| `system.Decay` | cold costs a body `Chill * (1 - Shelter)`, in the larder and, at a hundredth of the rate, in its health. A roof takes all of it. What makes winter dangerous is not the cold, it is being caught out in it. |
| `habit.Chill` | the twentieth dimension. Farming and foraging name the warmth, gathering wood and building name the cold, so the same settlement does different work in February and in June without anyone being told to. |
| `system.MarketStep` | perishables keep better in the cold. `Keeping` is what the granaries stop and what the weather stops on top of it, and at the bottom of a hard winter food spoils at two fifths of its usual rate. |

**What it cost, measured.** Twenty-four seeds, 4000 ticks, 20 founders, population at the end:

| | median | mean | extinctions |
|---|---|---|---|
| no seasons | 128 | 132 | 0 |
| seasons that nothing reads | 140 | 144 | 0 |
| first build: no winter growth, chill read bipolar | 21 | 28 | 1 |
| the same, chill read one-sided | 51 | 69 | 1 |
| winter at 0.35 of full growth, no keeping | 69 | 87 | 0 |
| shipped: winter at 0.5, and the cold keeps food | 60 | 86 | 1 |

The second row is the control, and it is the one that made the rest legible. The weather draws from `World.RNG` whether or not anything reads it, which reshuffles every downstream draw in the simulation, so a seasoned run and an unseasoned run of the same seed are different worlds and cannot be compared one to one. A climate that turns and draws and that nothing reads costs nothing: 140 against 128 is the same number twice. Everything below the control is winter, not noise.

**Two things had to give, and a third was a plain mistake.**

- **A winter that stops the forest dead is a siege, not a season.** Nothing in this world can put a summer by: the larder is a few meals and they spoil. At zero winter growth the median settlement fell to 21 and one seed died outright, which is the first extinction this sweep has seen. `WinterGrowth` is 0.35 of full growth now - about half the old flat rate once the year is renormalised - and the year is lean in its cold half rather than empty.
- **Cold has to press on the unhoused, not on the settlement.** `ColdDrain` began at 0.010, half again the ordinary drain of 0.02. That is a cost the whole population pays, because shelter rots and nobody is fully roofed all the time. At 0.005, a quarter of the ordinary drain at its very worst, an unhoused winter tells and a housed one does not.
- **The chill coordinate must be silent in mild weather.** Read bipolar, +1 in the cold and -1 in the warmth, it was never silent: an ordinary spring read as a strong -1 in every situation vector, and farming, alone in naming the warmth, outranked studying for the curious and getting even for the wronged. Two `core/action` tests caught it. Chill is one-sided now, like order and caution: 0 through the mild half of the year, rising to 1 at the bottom of a hard winter. It taxes farming in February and nothing in June. Between those two readings, with identical physics, the median settlement went from 21 to 51 - the habit layer earning its keep, and the largest single effect measured here.

**Four thousand ticks was the wrong horizon to tune against.** At 4000 the floor at 0.5 and the floor at 0.35 gave the same median, 71 against 69, and it looked as though the depth of the winter cost nothing. It does not show at 4000 because a settlement that has stopped breeding still has its founders. Run to 6000, which is what `TestCityDevelopsByRecognition` runs, seed 1 died outright at 0.35: seventeen founders, not one birth in 2500 ticks, and then old age. Births need physiological, safety and belonging all above their thresholds at once, and physiological's is a cliff at 0.7. A settlement whose mean sits at 0.55 to 0.68 does not breed slowly, it does not breed. That cliff is why a small per-tick cost buys a large population effect, and why measuring at a horizon shorter than a generation hides it.

**Somewhere to put a summer.** The answer to a lean winter is not a milder winter, it is a store. Perishables now keep better in the cold - `ColdKeeping` stops three fifths of their spoilage at the bottom of the year - and a granary keeps more of them than it did, 0.4 of the usual spoilage per granary against 0.5. A cold store is the oldest one there is, and it is the one thing in this world that gets better as the weather gets worse: what the year takes from the settlement in the growing it stops doing, it gives a little of back in the keeping. On the deep season it was the difference between seed 1 ending at 0 and ending at 17.

| 6000 ticks, six seeds | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|
| winter 0.35, drain 0.005, no keeping | 0 | 400 | 400 | 99 | 400 | 203 |
| winter 0.35, drain 0.005, keeping | 17 | 400 | 281 | 8 | 399 | 400 |
| shipped: winter 0.5, drain 0.003, keeping | 43 | 364 | 399 | 113 | 167 | 142 |

Keeping saves the deep winter but leaves two settlements hanging on by their fingernails, at 17 and at 8. The shipped season leaves none: it is a two-to-one year rather than a three-to-one one, still a season you can watch on the map, and it costs nothing to read.

**What the season costs, on the ground it ran on then.** Everything above was measured on the flat map, before the terrain underneath was rebuilt. Re-measured on the ground of "The land underneath", 24 seeds to 6000 ticks, against a control in which the year turns and draws from `World.RNG` but `Growth` returns 1 and `Chill` returns 0:

| | median | mean | extinctions | settlements at the cap |
|---|---|---|---|---|
| the year turns, nothing reads it | 378 | 273 | 2 | 11 |
| shipped, winter at 0.5 | 89 | 185 | 2 | 8 |
| winter at 0.6 | 136 | 172 | 4 | 6 |

Two things to take from this and one not to. The season costs the median settlement a great deal - 378 to 89 - and it costs it in growth rather than in survival: the control has the same two extinctions the shipped rules do. On this map extinctions are the terrain's doing, not the winter's; the flat map's tidier story, no deaths without seasons and three with, did not survive the ground being rebuilt underneath it.

What not to take from it is the third row. A milder winter is not obviously better: the median rises, the mean falls, the extinctions double and fewer settlements reach the cap, which is four metrics disagreeing. The distribution is bimodal - a settlement either takes off and pins at the cap or founders under twenty - so the median mostly reports which side of that split the middle seeds fell, and 24 seeds cannot separate 0.5 from 0.6. Nor can the per-seed columns be compared: changing `WinterGrowth` changes `growthNorm`, which shifts every float after it and re-rolls the world, so seed 4 at 0.5 and seed 4 at 0.6 are not the same settlement in different weather. 0.5 is kept because it is the deeper season and nothing measured argues against it.

**And what it costs now.** That reading did not last either. A house came to cost a forest, the weather began to move the ground, and a household's bread was given the ground it takes; measured again on the tree those landed in, the same pair reads:

| 24 seeds, 6000 ticks | median | mean | extinctions | under twenty | largest |
|---|---|---|---|---|---|
| the year turns, nothing reads it | 108 | 119 | 0 | 0 | 206 |
| the season live | 94 | 79 | 0 | 5 | 167 |

The season costs about a seventh of the median where a tree earlier it cost three quarters, and the reason is not the season: the whole distribution has changed shape. It used to be two humps - a third of the seeds pinned at the four-hundred cap and the rest floundering near zero - which is why its median jumped about so much that 0.5 and 0.6 could not be told apart. Now nothing reaches the cap, nothing dies, and both columns sit in a single band between forty and two hundred. Whatever did that belongs to the three changes above rather than to the weather.

What the season does on this ground is thin the weak without killing them: the mean falls from 119 to 79 while the median barely moves, five settlements come in under twenty where the control has none, and seventeen of twenty-four seeds end lower. That is the shape the feature was for. It also retires the extinction question this section spent so long on - neither column loses a settlement now - and with it the ceiling that `WinterGrowth` was being tuned against. If the depth of the winter is ever revisited, it should be re-argued from this distribution and not from the one above.

**What is left on the table.** A seasoned settlement is smaller than an unseasoned one, and it should be: a year with a lean half in it is a harder world, and the population it carries is the population its worst season carries. The store is only a market store, though. An agent's own larder still does not spoil and still holds almost nothing, and nothing in the habit layer has yet learned to sell into a granary in August and buy out of it in February. Whether recognition can learn a habit whose reward is a season away is the interesting question here, and this package does not answer it.

## What a household eats

A field was one tile, the same ground a house stands on, and that is not what a family lives off.

**The calculation.** A year's bread for a household of five is on the order of a tonne of grain. Wheat before the plough of our own age gave perhaps a tonne to the hectare in a good year, a quarter of which went back into the ground as next year's seed, and half the holding lay fallow while the other half bore. That is something like three hectares held to eat from one - against the sixty square metres the family slept under, a field some hundreds of times the house. No settlement has ever been laid out the other way round.

The map cannot carry that ratio. At eighty by thirty-six tiles, three hectares to a household would leave room for a dozen families and nothing else. So it is compressed, and how far it compresses was measured rather than argued: `fieldTiles` is 3.

**A holding is one farm, not three fields.** The household works the whole of it in a season and eats the whole of it, so the harvest is read off the mean fertility of the strips together and takes one harvest's worth out of the ground, shared over all of them. A worn strip is carried by the rest, which is what land enough to rotate is for, and the fallow keeps up with a draw a third the size. Three strips are not three families' worth of food; they are one family's, off land that gets a rest between crops. The farmer goes to the *nearest* strip they hold - making them cross their own land to reach the best of it only cost them the walk, and the walk came out of the time they had for everybody else.

**How a holding grows.** It is not claimed in one act. A farmer short of what the household eats breaks another strip beside the ones they have, and that harvest comes off ground that was grass the same morning. Three rules bound where:

- **Only soil that will bear.** A strip must hold fertility 0.3, the same as the first furrow, so a holding grows toward the valley floor and stops on the dry hillside.
- **Never poorer than what you have.** New ground must be at least as good as the holding's mean. This is the rule the whole change turned on. Without it a farmer took in whatever was next to them, the mean fell, and because the harvest is read off the mean their crop fell with it: the same seeds ran an eighth below one-tile fields, and it read as land hunger when it was really a farmer working ground that made them poorer. With it the cost disappears.
- **Not against a wall.** New ground may not touch a roof, so the built core keeps the gaps its streets are laid along and the holdings lie outside it, which is where a village puts its fields. The *first* furrow is exempt: like a roof, a holding is what a settlement wants and not what it owes, and a man with no land takes the ground he can get. Requiring the first furrow to clear the houses too was tried and cost three settlements of twenty-four - newcomers in a built-up place had to walk out past everything to start at all.

Nothing else was needed to give the land back: `Upkeep` already lets a field nobody is left to work go to weeds, and that matters more the bigger a holding is.

**Three is what the world carries.** Two independent batches of 24 seeds, 6000 ticks, 25 founders, on the turning year:

| | mean population | median population | field tiles over 48 seeds | field tiles a head |
|---|---|---|---|---|
| one tile a farmer | 156 | 102 | 5865 | 0.82 |
| holding of three | 150 | 93 | 8722 | 1.48 |
| holding of eight | 145 | 64 | 10994 | 1.68 |
| holding of three, taking whatever was next to it | 127 | 54 | 9276 | 1.64 |

Half again as much cultivated ground and near twice as much per head, for a population the batches cannot tell apart. Eight buys a quarter more ground again for a fifteenth of the mean population and a third of the median, which is not a trade worth making: past three, a holding stops being what a household eats and starts being what a household holds while its neighbours have none.

**Land is the binding constraint,** which is the point and was not designed in. On the drain-fed map the soil worth ploughing is a band along the valley floor, and holdings claim it by mid-run: on four seeds looked at, three quarters to all of the farmers short of a full holding had nothing left touching them that was both worth breaking and no poorer than what they already work. Which is why the cap is not what most people get. Mean holdings run 1.0 to 1.8 strips and a good half of the holders are still on their first furrow; the full three is what the farmer who broke ground early has, and everyone after them works what is left. The change is not that every field is three times bigger - it is that the settlement's cultivated ground is half again as large, that the people who farm properly hold a parcel instead of a plot, and that ground is now something a valley can run out of.

Cropping the best strip and letting the rest rest, rather than reading the harvest off the holding together, was also tried: it grows the holdings faster, and it was worse on both batches (mean 139, median 20 and 116, twelve settlements of twenty-four lasting on the first). A holding that carries its own poor ground is the version that works.

## The ontology

The catalog was a list of twenty-eight hand-written acts, each five closures and a prior tuned by hand, in a fixed order that every habit table was indexed by and a fixed bound of thirty-two that it could not outgrow. That is a linear cost per act, and it never gets to an extensive catalog. `core/ontology` replaces the list with a statement of what the world is made of, from which what can be done with it follows.

**Three trees and a set of relations.** Things are what an act operates on; sites are where it happens; roles are how a person stands to the actor. A class exists only if some verb treats it differently from its siblings; anything else is a trait.

```
Thing                                 Site
├─ Material                           ├─ Ground            [passable]
│  ├─ Provision   [edible]            │  ├─ Open
│  │  ├─ Berries  [perishable]        │  ├─ Wood           affords Berries, Game, Timber
│  │  ├─ Game     [perishable]        │  ├─ Water          affords Fish
│  │  ├─ Fish     [perishable]        │  ├─ Outcrop        affords Stone
│  │  ├─ Grain                        │  └─ Field [owned]  affords Grain
│  │  └─ Meal     [perishable]        └─ Built
│  ├─ Timber      [burnable, buildable]  ├─ Dwelling  [roofed, owned, bench, hearth, forge, desk]
│  ├─ Stone       [buildable, heavy]     ├─ Market    [public, bench, desk, trade]
│  ├─ Tool        [wears]                ├─ Granary   [roofed, public, store]
│  └─ Coin                               ├─ Tavern    [roofed, public, hearth, company]
├─ Person                                └─ Road      [passable]
└─ Practice
```

`Provision` is one class to eating, which takes whatever is on hand, and five to taking, because fishing and foraging and hunting are different habits. `Coin` is a thing so that money can be made, given, and stolen like anything else; a young settlement barters, and coin arrives when somebody mints it. `Practice` is an act itself as the object of another - what is taught and studied - so the ontology contains its own catalog. `Person` is one class: nobody is a pupil the way an oak is timber, so how a person stands to the actor is a **role** (self, neighbour, requester, holder, pupil, wrongdoer), a predicate decided at the moment of acting rather than a subclass. Workplaces are traits on built sites, which is how a hearth is named without saying whose: a dwelling has all of them, the market lends a bench and a desk to those without one, the tavern a hearth.

**Verbs are schemas over classes.** Eleven cover the catalog: take, make, raise, tend, consume, dwell, exchange, transfer, pass, strike, move. A schema states an act over classes - `Take Material @ Ground`, `Make Timber → Tool @ bench`, `Transfer Provision > neighbour` - and `Instantiate` expands it over the leaves of its object and site classes, only where the relations allow (a taking needs the site to afford the material), unless told to collapse (eating is one act). Theft is a transfer the other way round; everything that makes theft theft is in the traits and the valence, not in a separate verb. Transforms - food spoiling, a field going back to grass - are the world's half of the ontology: they never instantiate as acts and belong to upkeep.

**Every act has a key.** `take/timber@wood`, `make/stone+timber>tool@forge`, `transfer/provision<holder`, `pass/practice>self`: what it does, with what, where, and to whom. The walk is deterministic and its output is sorted by key. The catalog is what the walk entails, twenty-nine acts, each bound to the code that carries it out; an act the trees entail and nothing carries stops the program at start. Farming is two acts, as the trees have it: `tend/clear@open` makes ground into a strip of a holding, `take/grain@field` harvests one.

**Priors are composed, not written.** An act's prior is the sum of the verb's moment, what the act must already hold (an input's stock coordinate turned round), what it wants (the output's or the object's), the site it happens at, and the role it is done to, plus a small residue on the schema for what none of those explain. `Wanting` a material includes a share of the wanting of what it makes - timber is wanted for the roof it will be - taken from the one thing nearest to hand, the output whose making starts furthest into reach. The hand-tuned priors are kept beside the composed ones as `Def.Tuned` and held against them in the golden test at a cosine floor of 0.75; the city test replaces its founders on the composed ones. What the comparison caught, each of them a failing test first:

- **A class's prior is its stock coordinate and nothing else.** With the shelter motive written into timber, holding timber turned it round, and building a house came out as the unsafe moment with the sign flipped.
- **A want must not average.** Weighted over everything timber could become, the want spread so thin over so many coordinates that gathering wood fit every moment a little and no moment well.
- **Preconditions are not the moment.** Eating inheriting the stock it needs, and study the skill it needs, both lost to gathering wood on cosine: those coordinates sit dead against the situation and dilute the ones that matter. Consume adds no holding; the practice-want lives on the pupil role.
- **An underspecified prior is a universal one.** Moving house came out as bare nearness, which fits nearly any moment, and agents moved house instead of living in one.
- **Refining does not want what it holds.** A meal out of provisions is a keeping act, not a hungry one; a make whose output is-a its input adds only the output's delta over the input.

**Slots.** Habit tables were arrays bounded at thirty-two. They are slices indexed by slot, where `habit.Register` gives an act's key a slot once, in registration order, and never renumbers it: an act that enters the catalog later takes a fresh slot after everything before it, every table grows to make room (`Agent.Room`, `World.Room`), and an agent alive when that happens is seeded for the new act from its prior at its next decision and keeps what it had learned of the rest (`Agent.Seeded`).

**Taking is one act.** Foraging, hunting, felling, fishing, and cutting stone were five copies of the same act. `take.go` states them as data: a *lode* per material (the tile stock it draws on, how much a taking takes and gives, the luck in it, what it wears and teaches, what the tile becomes) and a *ground* per site (how to know a tile of it, and which tile the taking draws on - fishing stands on the bank and draws on the water beside it). Any taking the trees entail for a material with a lode at a site with a ground is carried out without code of its own. The five keep their names and numbers and the order they draw luck in; a picked forest still gives something, on purpose, and the paving test fails if it does not.

**Adding a material** is now: a class under `Material` with its traits, an entry in `Affords` for where it lies, a lode row, and a stock on the tile if it draws one down. The walk entails the taking, the composition gives it a prior, the registry gives it a slot, and the binding at start says if anything is missing. Making, raising, and the rest are still bound by hand to the tuned mechanics they had - a meal's `helping`, a house's `RoomToBuild`, a bridge's cost - and go the same way verb by verb as their bespoke logic allows.

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

`core/habit` defines a 19-dimensional space. Every coordinate is roughly in [-1, 1] with 0 meaning neutral.

Shared dimensions, computed once per agent per decision:

| dim | name | source | mapping |
|---|---|---|---|
| 0-4 | hunger, unsafe, lonely, unproven, curious | `need.Urgencies(a.Needs)[t] * a.Personality[t]` | `2*clamp01(x) - 1` |
| 5 | food | `Inventory[Food]` | `2*clamp01(food/4) - 1` (same knee as `foodValue`) |
| 6 | wood | `Inventory[Wood]` | `2*clamp01(wood/2) - 1` (the cost of a house, so +1 means "can build") |
| 7 | wealth | `Wealth` | `2*clamp01(wealth/20) - 1` |
| 8 | shelter | `Shelter` | `shelter - 1`, a lack: half a house is half a house short |
| 9 | company | `w.Neighbor(a, 8) != nil` | +1 or -1 |
| 10 | order | `w.Safety` | `safety`, one-sided |
| 11-14 | honesty, charity, industry, tradition | `Norms` | `n`, one-sided, frozen |
| 15 | caution | `Caution` | `c`, one-sided, frozen |

Per-candidate dimensions, patched for each action being weighed:

| dim | name | source |
|---|---|---|
| 16 | near | `1 - 2*clamp01(TravelCost(a.Pos, target) / Vigor() / 30)` |
| 17 | rapport | `Anticipate(a, other)` for actions done to a person; sign flipped for retaliate; 0 otherwise |
| 18 | skill | `2*Efficacy[def.Skill] - 1` for skilled actions; 0 otherwise. Belief, not truth. |

Personality is folded into the urgency dimensions rather than kept separate. The moral coordinates and order are one-sided because the belief layer defines them that way: a norm of 0 is holding nothing, caution of 0 is having learned of no reprisal, safety of 0 is nobody keeping order. Read on that scale an ordinary agent minds a wrong about half as much as a saint, which is what conscience charges in the value rule. Centred at 0.5 they fell silent for everyone but the extremes, and theft ran wild (see the comparison below). The moral dimensions are frozen: they are the same across every candidate an agent weighs, so if habits could learn them every habit would soon carry the agent's own norms and the dimension would cancel out of the choice. Values belong to the agent. Habits learn situations.

Dropped on purpose: health (tracks hunger and shelter), tools (sell and craft are hard-gated), market price, knowledge (global and unbounded).

## Signatures and habits

- `action.Def.Prior` is the shared signature of the moment an action belongs to. It is hand-seeded in `core/action/priors.go` and never learned. `Def.Skilled` names the skill a candidate draws on for this agent right now, and `Def.With` names the other agent it involves, so the per-candidate dimensions can be filled in.
- `entity.Agent.Habits[i]` is the agent's own copy for catalog position `i`, seeded from the prior on first decision (`Imprinted`) and moved by experience.
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

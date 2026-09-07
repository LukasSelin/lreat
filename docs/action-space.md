# Fit-based action space

Status: all six phases implemented, plus credit by provenance. **Recognition is the default rule.** The value rule stays behind `world.Rules.Fit = false`, or `-value` on the headless runner, and its tests run through a `valueWorld` helper so both rules stay covered. Recognition settlements now replace their founders: on seed 7 the population goes from 20 to 65 over 6000 ticks with all four techs; seeds 31 and 1 reach 63 and 114; seed 21 survives but shrinks to 13, the weak case to watch.

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
| 8 | shelter | `Shelter` | `2*shelter - 1` |
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

A unit of food remembers the act that produced it and the moment it was taken in (`Agent.Larder`, oldest first, capped at 16). When a unit is consumed, by a meal, a sale, a gift, or a thief, the act that brought it is credited. A roof remembers the act that last raised shelter (`Agent.Roof`), a watch the act by which the agent last kept public order (`Agent.Watch`); safety that rises during any later plan credits both, for the part of that plan's reward safety accounts for.

What the credit carries is the **advantage** of the consuming plan, not its reward. This was the difference between a working economy and a runaway one. Thanked with the raw reward, a forage got good news from every meal, its habit drifted onto the average moment and fit everything, and agents foraged every six ticks into a larder of thirty units while houses rotted. Thanked with the advantage, a forage that fed a full belly is pushed away from that moment, and production regulates itself: a meal better than meals usually are pulls the field toward the moment it was worked in, a needless one pushes it away.

Guard starts fully in reach. It is the one public good in the catalog, and a settlement that has to discover it first has died of disorder before it does. The value rule's safety turned out to come from public order too, not from houses: its agents guard several hundred times per 500 ticks and hold order at 1.0, while their shelter is as low as recognition's. Long actions carry more decay in `r`, which is the old time cost re-emerging from physics rather than from a formula.

The trace is what keeps the economy alive under a needs-only reward. Farm and forage never touch needs, only the larder. When a later eat pays +0.35, the trace hands half of that to the previous plan and a quarter to the one before. If liveness runs still show farming being learned away, the documented fallback is to add stock terms to `r`. That would move a value judgement into learning, which is the agreed place for it, but it has been rejected for now.

### Second tuning pass and the subsistence finding (phase 6)

Flipping the default broke one test: on seed 21 nobody posted a request in 3000 ticks. Diagnostics showed why. Wealth stayed at zero because nobody sold; nobody sold because food per agent never rose above about half a unit; and that was because the farm prior said "hungry and out of food", so agents farmed only when hungry and foraged the rest of the time. Three changes, all inside the design:

- **The farm prior no longer mentions hunger.** Foraging is what hunger calls for. Farming is what an industrious person with a field nearby does, whether or not the larder is low. That is the only moment in which a larder ever fills past today.
- **Wood is measured against the cost of a house**, so "enough wood" reads as +1 exactly when a shelter can be built, and the gather and build priors are worded around the unsheltered moment.
- **Per-action baselines** (above), so that a good outcome for an act is judged against that act's usual outcome.

After these, seed 21 reaches 41 fulfilled requests per 500 ticks by tick 2000 and wealth grows; the social seed keeps its request rate (276 asked, 230 fulfilled) with theft at 293 and giving at 222.

What did not move is safety, and with it growth. Agents forage every 17 ticks and gather wood a tenth as often, houses rot faster than they are rebuilt, and mean safety sits near 0.3 against 0.5 under the value rule, below the 0.6 a birth requires. The cause is structural. Under the leaky hierarchy a moderately hungry agent has its safety urgency damped, so it chases food; recognition has no notion that a field is more efficient than the forest, so it chases food the slow way; and a needs-only reward gives no credit for surplus, so nothing it learns changes that. **Recognition with a needs-only reward produces a subsistence society: fed, housed after a fashion, social, literate in time, and not growing.** Whether that is a bug or the point is a design call. The two levers left, both rejected so far, are an efficiency or stock term in the reward, and a longer credit horizon (a trace of 8 at 0.75 was tried and made requests collapse, because it spread each meal's credit over everything).

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

## Reach

Implemented in `core/action/reach.go`; the constants live there.

- `Reach0` per action. Everyday living (rest, eat, forage, farm, gather, build, sell, buy, socialize, give, steal, retaliate, fulfil) starts at 1. Gated: craft 0.5, guard 0.6, teach 0.3, study 0.4.
- **Study broadens.** Every gated action comes closer: `Reach += 0.03 * (1 - Reach) * Mods.StudyRate`, so written records make study widen reach twice as fast.
- **Teaching passes recognition on.** The student's reach for the taught skill's action becomes `max(own, 0.6 * teacher)`, and its habit moves 0.3 of the way toward the teacher's. `ForSkill` maps farming, building, crafting, scholarship, and guarding to farm, build, craft, study, and guard.
- **Discovery opens.** A `Discovery` has an `Opens` list; masonry opens craft, writing opens study and teach, metallurgy opens craft and guard. Each raises `world.ReachFloor` for that action to 0.8, and every agent is lifted to the floor at its next decision.
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
| 6 | Recognition is the default (reverted once after the aging merge, restored with provenance); headless `-value`; value-rule tests run through `valueWorld`; recognition twins at full length; ordering twins for every value-rule choice test in `core/action/situation_test.go`; per-action baselines, industrious farm prior, wood knee at the house cost, guard reach 0.8 | done |

Tests under fit mode assert ordering (which action ranks first), not the sampled outcome. `TestHungerEventuallyOverwhelmsPrinciple` is about magnitude and stays value-mode only. The four liveness tests run in both modes from phase 4 onward so tuning is visible before the default flips.

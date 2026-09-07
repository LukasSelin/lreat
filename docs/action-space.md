# Fit-based action space

Status: phases 0 to 2 implemented (habit package, catalog hygiene, agent state). Value-based choice remains the default until the fit chooser is tuned. Switch: `world.Rules.Fit`.

## Why

`system.Choose` is an expected-value maximizer. For every action it multiplies the action's expected need gains by the agent's urgencies and personality, adds conscience, subtracts expected reprisal, divides by time cost, and takes the argmax with a little noise. It works, but it means every agent is a small economist, and the only way to change behaviour is to change payoffs.

People do not choose that way most of the time. They recognise what kind of moment they are in and do what such moments call for. The value of the outcome shapes what they will recognise next time. This document describes a decision rule built on that idea: **choice is recognition, learning is where value lives, and the two never meet in the same step.**

Three commitments, agreed up front:

1. Ranking is by cosine fit between the moment and each action's signature. The overall intensity of the moment only sharpens or loosens the sampling. It never reorders candidates.
2. Signatures ("habits") are per agent, seeded from one shared prior per action. Teaching copies them. Children inherit them.
3. The learning signal counts need changes only. No stock valuation, no pride term. Actions that only set up a later gain (farm, forage, gather, sell) learn through an eligibility trace that hands part of a later reward back to the actions that preceded it.

## The space

`core/habit` defines a 19-dimensional space. Every coordinate is roughly in [-1, 1] with 0 meaning neutral.

Shared dimensions, computed once per agent per decision:

| dim | name | source | mapping |
|---|---|---|---|
| 0-4 | hunger, unsafe, lonely, unproven, curious | `need.Urgencies(a.Needs)[t] * a.Personality[t]` | `2*clamp01(x) - 1` |
| 5 | food | `Inventory[Food]` | `2*clamp01(food/4) - 1` (same knee as `foodValue`) |
| 6 | wood | `Inventory[Wood]` | `2*clamp01(wood/3) - 1` |
| 7 | wealth | `Wealth` | `2*clamp01(wealth/20) - 1` |
| 8 | shelter | `Shelter` | `2*shelter - 1` |
| 9 | company | `w.Neighbor(a, 8) != nil` | +1 or -1 |
| 10 | order | `w.Safety` | `2*safety - 1` |
| 11-14 | honesty, charity, industry, tradition | `Norms` | `2n - 1`, frozen |
| 15 | caution | `Caution` | `2c - 1`, frozen |

Per-candidate dimensions, patched for each action being weighed:

| dim | name | source |
|---|---|---|
| 16 | near | `1 - 2*clamp01(TravelCost(a.Pos, target) / Vigor() / 30)` |
| 17 | rapport | `Anticipate(a, other)` for actions done to a person; sign flipped for retaliate; 0 otherwise |
| 18 | skill | `2*Efficacy[def.Skill] - 1` for skilled actions; 0 otherwise. Belief, not truth. |

Personality is folded into the urgency dimensions rather than kept separate. The moral dimensions are frozen: they are the same across every candidate an agent weighs, so if habits could learn them every habit would soon carry the agent's own norms and the dimension would cancel out of the choice. Values belong to the agent. Habits learn situations.

Dropped on purpose: health (tracks hunger and shelter), tools (sell and craft are hard-gated), market price, knowledge (global and unbounded).

## Signatures and habits

- `action.Def.Prior` is the shared signature of the moment an action belongs to. It is hand-seeded and never learned.
- `entity.Agent.Habits[i]` is the agent's own copy for catalog position `i`, seeded from the prior on first decision (`Imprinted`) and moved by experience.
- `entity.Agent.Reach[i]` in [0, 1] is how far into reach action `i` is for this agent. `Def.Reach0` seeds it.
- Habit length is clamped to [0.2, 2] after every update. Cosine of a zero vector is 0, never NaN.

## Choice

For every action that is `Available` and has a target:

```
eff_i = cos(S_i, H_i) - 0.6 * (1 - Reach_i)
```

Sample from a softmax over `eff` at temperature `τ / |S_shared|`, with `τ = 0.15`. A moment with strong urgencies has a large norm, so it is decided sharply. A bland moment is decided loosely. Exactly one `w.RNG.Float64()` is consumed per decision.

Nothing is divided by cost. Nearness is a dimension the habit learns about. Hard physical gates in `Available` stay (eating needs food). The only soft judgement gate in the catalog today, teach's skill floor, becomes reach.

## Learning

At every plan end, whether or not `Apply` ran (a plan whose `Available` went false while walking is a lesson too):

```
r   = Σ_t (Needs_after[t] - Needs_before[t]) * Urgency_at_decision[t] * Personality[t]
adv = clamp(r - Baseline, -0.5, 0.5)
Baseline += 0.02 * (r - Baseline)
Trace.Push(index, S_at_decision)             // keep 3, newest first
for k, step in Trace:
    H[step.Index] += 0.10 * 0.5^k * adv * (step.Situation - H[step.Index])   // learnable dims only
H[index] += 0.003 * (Prior[index] - H[index])                                // retention, all dims
clamp |H[index]| to [0.2, 2]
Reach[index] = min(1, Reach[index] + 0.02)                                   // doing is learning
```

The reward uses urgencies from the moment of the decision, so an outcome is judged by what the agent wanted then. The baseline is global per agent, not per action, so uniformly poor actions are learned away. Long actions carry more decay in `r`, which is the old time cost re-emerging from physics rather than from a formula.

The trace is what keeps the economy alive under a needs-only reward. Farm and forage never touch needs, only the larder. When a later eat pays +0.35, the trace hands half of that to the previous plan and a quarter to the one before. If liveness runs still show farming being learned away, the documented fallback is to add stock terms to `r`. That would move a value judgement into learning, which is the agreed place for it, but it has been rejected for now.

## Reach

- `Reach0` per action. Everyday living (rest, eat, forage, farm, gather, build, sell, buy, socialize, give, steal, retaliate, fulfil) starts at 1. Gated: craft 0.5, guard 0.6, teach 0.3, study 0.4.
- Study raises reach of gated actions: `Reach += 0.03 * (1 - Reach) * Mods.StudyRate`.
- Teach: the student's reach for the taught skill's action becomes `max(own, 0.6 * teacher)`, and its habit lerps 0.3 toward the teacher's.
- Discoveries gain an `Opens` list and raise a world reach floor for those actions.
- Births copy the parent's habits with N(0, 0.05) noise on learnable dims and `Reach = max(floor, 0.7 * parent)`.

Reach is the "distance gate": a far action is one whose signature the agent cannot yet reach, and study, teaching, and discovery bring it closer. Because it is a penalty on fit rather than a lock, a curious agent can occasionally reach a far action early. Pioneers fall out of the sampling.

## Expected impact on the simulation

- **Roles.** Agents keep doing what fitted before, so farmers, scholars, and thieves become individuals rather than a population-wide threshold. Feuds cluster around people.
- **Slower reaction to acute need** until temperature and learning rate are tuned. Intensity sharpening is the safety valve. Watch starvation counts; seed 7 today has 4 deaths in 6000 ticks.
- **Techs later.** Study is gated and knowledge only comes from study and scholar service.
- **Culture.** Teaching and births transmit habits, so settlements diverge across seeds more than they do now.
- **Determinism** holds: habit tables are arrays, one RNG draw per decision, fixed catalog order. Fit mode and value mode are different RNG streams for the same seed, so comparison is statistical, not trajectory-level.
- **Player.** `sim.Intend` must build plans through the same builder, so player commands teach the player's habits.

New metrics for `observe.Snapshot`: `HabitSpread` (mean distance of unit habits from the population mean; 0 means collapse), `MeanReach`, `GatedReach`, `Starved` (cumulative), `ChoiceEntropy`. One extra column pair in headless, one line in the TUI.

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
| 3 | `action.Situation(a, w, d, target)`, priors and `Reach0` for all 17 actions, `Imprint`, canonical-situation ranking tests | next |
| 4 | Fit chooser in `system.Decide`, learning at plan end in `system.Act`, `sim.Intend` through the shared builder, headless `-fit` and `-temp` flags | |
| 5 | Reach growth from study, teach, discovery; inheritance at birth; metrics in snapshot, headless, TUI | |
| 6 | Flip `DefaultRules` to fit; port choose tests to ordering twins via a `Rank` helper; keep value mode behind the flag | |

Tests under fit mode assert ordering (which action ranks first), not the sampled outcome. `TestHungerEventuallyOverwhelmsPrinciple` is about magnitude and stays value-mode only. The four liveness tests run in both modes from phase 4 onward so tuning is visible before the default flips.

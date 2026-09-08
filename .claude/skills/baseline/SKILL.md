---
name: baseline
description: Judge whether a change to the simulation helped, using the checked-in master baseline instead of a fresh batch. Use when changing anything the settlements do — priors in core/action, habit coordinates, needs, systems, world rules — or when asked whether a change helped, to compare priors, to run cmd/tune, to read or refresh docs/baseline.md, or when a refactor of simulation code needs to be shown to have changed nothing.
---

# Judging a change against the baseline

`go run ./cmd/tune` is the only thing here that says whether a change to the
simulation helped. It costs a few minutes for the default batch, and it is
deterministic: the same tree gives the same numbers, however the goroutines
interleave.

Because it is deterministic, master's numbers are checked in at
[docs/baseline.md](../../../docs/baseline.md). They are the run. Re-running them
only confirms them.

## When to run it

**Do not run `tune` to find out where master stood.** That is what the file is.
Open it at the start if you need the numbers in front of you; that is the whole
of the before-measurement.

**Run it once, at the end**, when the change is settled and `go test ./...`
passes. A batch taken over unfinished work measures work that no longer exists,
and every rerun after that is a few minutes buying nothing.

If the change is large enough that you want to know mid-way whether you are
going the right direction, run a cheap probe — `-seeds 8 -ticks 3000 -quiet` —
and treat it as a smoke test, not as evidence. Only the full batch below
compares to the file.

## The run

```bash
go run ./cmd/tune -seeds 24 -ticks 21600
```

Exactly those flags. `-seeds`, `-ticks`, `-agents`, `-born`, `-inherit`, `-temp`
and `-value` all move the numbers, so a batch taken under any other flags is not
comparable to the file and must not be reported as if it were.

## Reading it

The last line is the headline. Most of what is on it is noise:

| reading | worth believing at |
|---|---|
| fed, and the four mean needs | a move of 0.05 on two batches that agree |
| gates all | a move of 0.05 |
| lasted, extinct | a move of 5 of 24 |
| mean and median population | never on its own |

Twenty-four seeds swing the mean population by about a tenth of itself with no
change to the code at all — 177, 174, 190 on three batches of master, and the
median by more. The "How much of that is chance" section of docs/baseline.md has
the evidence. So:

- A change that only moved population has not been shown to do anything. Confirm
  it on an independent batch, `-offset 24`, before believing it, and say in the
  commit that both batches agreed.
- A change that moved nothing past the thresholds is a change that did nothing.
  Say so plainly rather than reading a win out of the third decimal.
- **A refactor that moves any of these numbers was not a refactor.** Determinism
  makes `tune` an exact behavioural test: if the summary line differs from the
  file after a change meant to be behaviour-preserving, find out why before
  landing it.

## Landing it

If the change moved behaviour, refresh docs/baseline.md **in the commit that
lands the change** — the whole default-batch output, plus the commit's own
account of what moved. The file describes the tree that carries it, so a
behavioural change that leaves it alone has left it wrong.

Leave the file alone for changes that cannot move behaviour: docs, comments,
tests, tooling. The `-offset` section only needs redoing when the thresholds
themselves look wrong.

## Reporting

Give the summary line, not the twenty-four-row table — the table is in the file
for anyone who wants it. Name what moved past its threshold and what did not,
and if nothing did, say the change is behaviourally neutral.

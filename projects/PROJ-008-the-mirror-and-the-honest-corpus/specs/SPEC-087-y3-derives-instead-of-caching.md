---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-087
  type: chore                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
  blocked: false
  priority: high                   # sequencing, not size: it should land
                                   # BEFORE SPEC-086 design, which creates
                                   # DEC-050 and would otherwise be the sixth
                                   # consecutive hand re-pin of Y3.
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5
  created_at: 2026-09-06
  framed_at: 2026-09-06

insight:
  confidence: 0.75                 # the PROBLEM is measured over five specs;
                                   # the FIX shape (what independent source Y3
                                   # derives from without going vacuous) is the
                                   # open question this spec exists to settle.

references:
  decisions: []
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-086                     # split out of its re-framing, 2026-09-06
    - SPEC-085                     # the fifth consecutive hand re-pin
    - SPEC-084                     # the third; DEC-048 and the golden-length lesson
---

# SPEC-087: `Y3` derives instead of caching

> **Cycle: frame.** **GO** at complexity **S**. Split out of SPEC-086's
> re-framing on 2026-09-06 rather than absorbed, because taking it on there
> would have pushed an already-split spec back to L. Measured against `main` at
> `df369e9`.

## Context

STAGE-022's close promoted this to a stage-level lesson and then **held** it,
with the trigger *"routed to the next spec that opens that file."* SPEC-085
**was** that spec — it inserted a new assertion group into
`scripts/test-docs.sh` — and hand-re-pinned `Y3` anyway. That makes **five
consecutive specs: SPEC-081, 082, 083, 084, 085.**

STAGE-023 named **SPEC-086** as the owner, on the finding that *an anonymous
"next spec" is not an owner*. SPEC-086's re-framing accepted the ownership and
discharged it by **splitting rather than absorbing** — which is what this file
is. The routing note is now a spec with an id, and the id is reserved by a file
on disk, which matters mechanically: `scripts/_lib.sh:107-119` computes
`next_id` by scanning **filenames** (`find … -name "SPEC-*.md"`), so an id
reserved only in backlog prose would be handed straight to the next
`just new-spec`.

## Goal

Make `Y3` prove its claim without caching a literal that a human must move every
time a decision record is added — while keeping the failure mode it exists to
catch. If the answer turns out to be *"the cache is correct and the duplication
is the price"*, say so with evidence and close the lesson rather than carrying
it a sixth time.

## What the two assertions actually do — measured, not assumed

Re-derived on 2026-09-06. **They are not redundant**, which is the thing
five specs' worth of routing notes never quite said:

| | `X3` (`scripts/test-docs.sh:1449-1472`) | `Y3` (`scripts/test-docs.sh:1586-1626`) |
|---|---|---|
| Mechanism | runs `scripts/inventory.sh`, diffs its **whole output** byte-for-byte against the block between the `inventory:begin/end` markers in `docs/engineering-practices.md` | runs `scripts/inventory.sh`, then `grep -F -q` for two **literal rows**: `Decision records \| 48 \|` and `Decision numbers reserved, not yet decided \| 1 \|` |
| Catches | the page going stale against the script | the **script** going wrong — e.g. someone edits the `insight.type` filter so the DEC-041 reservation tombstone starts counting as a decision |
| Blind to | a script and a page that agree on a **wrong** number | nothing, while the pin is current — but the pin is a hand-maintained literal |
| Remedy on failure | mechanical: `just inventory`, paste | **manual: a human edits the number** |

`Y3`'s own comment states the failure mode precisely: *"script and page would
still agree, just agree on 48."* So the duplication is deliberate and the guard
has teeth. **The defect is not that `Y3` exists; it is that its expected value
is a literal.**

Current live values, re-derived: `Decision records | 48 |`,
`Decision numbers reserved, not yet decided | 1 |`. The inventory block and the
script agree (`X3` green, whole table byte-identical).

## The forks this spec must settle

### Fork 1 — is there an independent derivation at all?

This is the load-bearing question, and it is genuinely open.

`Y3` cannot derive its expectation from `inventory.sh`, because that is the
thing under test — it would become vacuously true, which is strictly worse than
a stale literal (a stale literal at least fails loudly). So a derivation must
come from a **second, mechanically different** count of the same corpus. The
obvious candidate is counting `decisions/DEC-*.md` front-matter directly in the
harness, by a different expression than `inventory.sh` uses.

**The trap, and it is the repo's own:** two greps over the same files, written
by the same author on the same afternoon, are not independent. AGENTS.md §9
already records five counts that were *"wrong the same way — grep-shaped
heuristics that read as authoritative."* Design must argue why its second
derivation fails differently from the first, and prove it with a **paired
mutation**: break `inventory.sh`'s filter and confirm `Y3` fires; break `Y3`'s
own derivation and confirm it fires too. A guard that survives both mutations is
vacuous. **A green guard is not evidence its claim is true.**

### Fork 2 — or is the right fix to delete `Y3`'s value pin and strengthen `X3`?

The alternative shape: keep one assertion, not two. If `inventory.sh` grew a
self-check — a row that reports *how many `DEC-*.md` files it examined and how
many it rejected, and why* — then `X3`'s byte-for-byte diff would catch a filter
change, because the rejection count would move and the page would go stale. That
converts a hand-maintained literal into a regenerated table row.

Cost: it changes the inventory table, which is a user-facing document, for a
harness-internal reason. Design must weigh that honestly.

### Fork 3 — the re-pin-note convention

`Y3`'s comment carries its own history (46→47 at SPEC-084, 47→48 at SPEC-085),
and SPEC-085's ship correction established that **nothing requires that note** —
it *"lives only in the comments it has already produced, which means it survives
exactly as long as the next author happens to read one."* If Fork 1 or Fork 2
lands, the convention becomes moot. If neither does, the convention is the only
thing standing between the next re-pin and a wrong baseline, and it should
become a requirement rather than a habit.

## GO / NO-GO

**GO**, at complexity **S**, and it should be **sequenced before SPEC-086's
design**.

The sequencing is the whole point. SPEC-086's design authors **DEC-050**, which
moves `Decision records` 48 → 49. If this spec has not landed by then, SPEC-086
becomes the **sixth** consecutive hand re-pin, and the lesson STAGE-022 promoted
will have failed under a *named* owner as well as an anonymous one — which is a
materially worse result than the current N=5, because it would rule out
"nobody owned it" as the explanation.

**What this is not.** It is not a rule to codify. STAGE-023 recorded the
duplication at N=2 same-outcome and this repo wants N=3 (AGENTS.md §12
codification meta-rule). This spec removes the duplication; it does not write it
into AGENTS.md.

## Acceptance Criteria

*Written at design.* Must include: the paired mutation results from Fork 1
(break each side, confirm each fires), and a statement of what `Y3` costs a
future spec that adds a `DEC-*` — ideally zero edits.

## Failing Tests

*Written at design.*

## Implementation Context

### Files this spec opens
- `scripts/test-docs.sh` — `X3` (~1449-1472), `Y3` (~1586-1626), `Y4`
  (~1628-1660, the same shape for the questions register and **not in scope**
  unless Fork 2 makes it free).
- `scripts/inventory.sh` — the derivation under test; `docs/engineering-
  practices.md` holds the generated block.

### Prior related work
- **STAGE-022 close** — promoted the lesson, held it, routed it anonymously.
- **SPEC-081…085** — the five hand re-pins.
- **SPEC-085 ship** — corrected verify's diagnosis of *why* the `Y3` note was
  missing, and established that the note convention is nowhere a requirement.

### Out of scope
- `Y4` (the open-questions pin), unless Fork 2 makes it fall out for free.
- Anything in `internal/`. This is a harness spec.

## Notes for the Implementer

*Written at design.* Two mechanical rules that bit twice in this cycle:

- **Quote `--include` globs.** Unquoted `--include=*.go` is expanded by zsh
  before `grep` runs and the search silently never happens.
- **`just inventory` only prints.** Paste its output between the markers with no
  blank lines inside them — `X3` compares byte-for-byte.

## Build Completion

*Filled at build.*

### Build-phase reflection (3 questions, short answers)

## Reflection (Ship)

- **What can a user do now that they couldn't before?** — one sentence,
  before → after. Capture this before closing the cycle.

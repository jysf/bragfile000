---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-086
  type: story                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
  blocked: true                    # on SPEC-085 — there is no `failed` row to
                                   # section off until the verb ships
  priority: high
  complexity: M                    # S | M | L  (L means split it)

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5
  created_at: 2026-09-05
  framed_at: 2026-09-05

insight:
  confidence: 0.80                 # the WHAT is decided by the user and the
                                   # measurement is solid; the two DEC-048
                                   # count renames are the open shape.

references:
  decisions:
    - DEC-014                      # the envelope
    - DEC-028                      # impact's two-number count + the 4-key projection
    - DEC-030                      # wrapped's section arc
    - DEC-048                      # a count must name what it counted
    - DEC-049                      # the reserved `failed` value (SPEC-085)
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-085                     # ships the verb and the value this consumes
---

# SPEC-086: what the celebratory digests do with a failure

> **Cycle: frame.** **GO** at complexity **M**. Split out of SPEC-085 at design
> on 2026-09-05, when the user's Fork 4 decision turned out to need a renderer
> change on two surfaces plus two DEC-048 count renames. STAGE-023's spec
> backlog pre-authorised exactly this split. All numbers below were measured on
> 2026-09-05 against `main` at `81e639d`, corpus **398**.

## Why this exists as its own spec

STAGE-023's backlog carried a conditional second spec:

> *"(not yet written) — digest posture: what `wrapped` / `impact` / `summary` /
> `story` do with a failure. **Only if SPEC-085's Fork 4 turns out to need more
> than a documented default.**"*

**The condition fired.** Put to the user at SPEC-085 design, Fork 4 resolved to
a named section rather than a documented default — which is renderer work on
two surfaces, two headline-count renames under DEC-048, and a decision record
of its own. Absorbing it would have made SPEC-085 an L; framing's own guidance
says split rather than run an L.

## The decision, and who made it

**Chosen by the user, 2026-09-05:** failures get their **own named section** —
`## What didn't work` — on `brag wrapped` and `brag impact`. They are pulled
*out* of *Impact moments* / the impact body, not hidden and not mixed in.

**Rejected, with the user's reasoning:**

- **Exclude by default.** `wrapped` stays purely celebratory, failures reachable
  only behind a flag. Rejected because it re-creates, inside the digest people
  actually read, the flattery PROJ-008 exists to remove.
- **Include silently** (status quo, zero code). Rejected — and independently
  ruled out by STAGE-023's success criterion that *"the celebratory digests do
  not silently absorb failures."*

The shape the user selected:

```
## Impact moments

### AI Development Factory

- 74: Shipped the single-worker AI dev factory PoC
  Proved the loop end-to-end...

## What didn't work

### AI Development Factory

- 411: Tried a shared-worker pool to cut cold starts
  Cost two days and produced nothing reusable.
```

## What was measured, so framing does not have to re-derive it

All on 2026-09-05, `main` at `81e639d`, corpus 398:

1. **`wrapped`'s *Impact moments* and `impact`'s body are the same document.**
   Extracted and diffed: **697 vs 696 lines**, differing by exactly one
   trailing blank line. Both carry **all 324** impact-bearing entries
   (`grep -cE '^- [0-9]+: '` → 324 on each).
2. **Neither surface renders `type`, in either format.**
   `internal/export/impact.go:60-61` emits `- %d: %s` + `  %s` (id, title,
   impact). The JSON projection `impactEntry` is a deliberately narrow **4-key**
   shape — `id`, `title`, `project`, `impact` (DEC-028 choice 4). `wrapped`
   inlines the same `aggregate.WithImpact` + `GroupEntriesByProject` pair
   (`wrapped.go:95`). **So a failure there is not merely unflagged — it is
   unrepresentable as a failure without a renderer change.** This is the fact
   that made "include silently" unacceptable.
3. **The headline counts today:** `impact` → `Entries: 324/398 with impact`;
   `wrapped` → `Entries: 398`.
4. **`--type` is exact-match INCLUSION; there is no negation** —
   `store.go:388`: `conds = append(conds, "e.type = ?")`. So *"wrapped, minus
   failures"* is code, not configuration. **But a negation precedent exists in
   the same struct**: `ListFilter.Author`'s human branch is
   `"NOT " + provenanceExistsClause` (`store.go:404`). Class-based rather than
   generic, and the pattern to follow.
5. **Seven commands carry `--type`** — `list`, `export`, `summary`, `impact`,
   `wrapped`, `story`, `coverage` — not the five framing listed. The scope
   question below has to answer for all seven, not four.

## The forks this spec must settle

### Fork A — the DEC-048 obligation, which is the real work

Once *Impact moments* stops carrying every with-impact entry, **its count
changes meaning and DEC-048 forbids silently redefining it.** Both surfaces
need an answer:

- `impact`'s `Entries: 324/398 with impact` — does it become a three-number
  form? `brag impact`'s existing two-number shape (DEC-028) is the precedent
  for how this repo does it.
- `wrapped`'s `Entries: 398` — the corpus count is still honest, but the
  *section* now needs its own accounting.
- The **JSON envelopes** move too: `impactEnvelope` has
  `entries_in_window` / `entries_with_impact`, and a new section needs a key.
  That is a breaking wire change on a DEC-014 surface, exactly like the
  `entries` → `candidates` rename DEC-048 already legislated once.

### Fork B — scope: which of the seven surfaces?

The user's decision names `wrapped` and `impact`. STAGE-023's scope guard says
whatever is chosen *"must apply to `impact`, `wrapped`, `summary` and `story`
consistently"* — and the measurement adds `export` and `coverage` to the list
of surfaces that filter by type. Design must decide, per surface, and say why:
`wrapped`/`impact` are celebratory; `summary`/`story`/`export`/`coverage` are
neutral, and a neutral surface arguably needs no section at all. **Consistency
must be stated for all seven even where the answer is "nothing".**

### Fork C — where the predicate lives

`aggregate.WithImpact` is the current split point. A failure predicate
(`Type == cli.FailureType`) must not live in `internal/cli/` twice, and
`internal/aggregate` is SQL-free and dependency-free by design — which makes it
the natural home, but it means `FailureType` moves out of `internal/cli`. Note
the single-sourcing precedent: `aggregate.IsAgentAuthored` is kept in agreement
with storage's SQL clause by a cross-package drift-guard test
(`TestProvenanceClassifier_GoPredicateMatchesSQLClause`). The same shape
probably applies.

### Fork D — the empty case

On a corpus with no failures — which is every corpus today, and most corpora
for a while — does `## What didn't work` render empty, or is it omitted?
DEC-014 part 4 has a precedent (`wrapped` omits body sections on an empty
period). Getting this wrong makes every existing user's `wrapped` grow an empty
accusatory heading.

### Fork E — goldens

`internal/export`'s byte-exact goldens all move. One of them
(`internal/export/memory_test.go:247`) is the assertion AGENTS.md §9 cites as
*not reachable by any grep* — it caches a golden document's byte length. Design
must enumerate which goldens move and how they are regenerated.

## GO / NO-GO

**GO**, but **blocked on SPEC-085** — there is no `failed` row to section off
until the verb exists, and the empty-case fork (D) is the only part testable
before it.

**Interim risk, accepted and named at SPEC-085 design:** between SPEC-085
shipping and this spec shipping, a `failed` entry carrying an impact statement
*would* appear silently in *Impact moments*. Bounded: the corpus holds zero
such entries, the user controls when the first is written, and the alternative
— shipping the digest change first — would mean designing a section for a row
type that does not exist.

## Acceptance Criteria

*Written at design.* Must be numbers to diff against: the exact `Entries:` /
new-count lines on each touched surface before and after, the line counts of
each changed golden, and the corpus count of `type: failed` entries used in the
fixtures.

## Failing Tests

*Written at design.* Note the standing trap: **a NOT-contains assertion proves
the old phrase is gone, never that the new one is true.** SPEC-085 closed this
with paired assertions (see its LD7 and mutation M-E); follow that shape.

## Implementation Context

### Decisions that apply
- **DEC-049** — the reserved `failed` value this spec consumes. Read it first.
- **DEC-048** — a count must name what it counted. **The binding constraint.**
- **DEC-028** — `impact`'s two-number count and its 4-key JSON projection.
- **DEC-030** — `wrapped`'s locked section arc; a new section changes it.
- **DEC-014** — the envelope, including part 4's empty-period rule (Fork D).

### Prior related work
- **SPEC-085** — the verb, the value, and the measurement above.
- **SPEC-084 / DEC-048** — the `Entries:` → `Candidates:` rename. The precedent
  for how this repo renames a count that stopped meaning what it said, and the
  source of the golden-byte-length lesson in Fork E.

### Out of scope
- The impact-quality classifier (STAGE-024). **This spec must not smuggle in a
  ranking rule** — sectioning by `type` is not ranking by quality.
- The memory pool fix (`memory-pool-composition-excludes-older-entries`).
- The mirror; story-surface v2.

## Notes for the Implementer

*Written at design.* Run the **§12(b) pre-flight**: SPEC-085's found five
things framing had wrong, two of which would have shipped a guard that could
never fail.

## Build Completion

*Filled at build.*

### Build-phase reflection (3 questions, short answers)

## Reflection (Ship)

- **What can a user do now that they couldn't before?** — one sentence,
  before → after. Capture this before closing the cycle.

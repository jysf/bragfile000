---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-049                        # stable, never reused
  type: decision                     # decision | analysis | recommendation | observation
  confidence: 0.88                   # honest: the schema home is measured and
                                     # near-certain; the choice of the literal
                                     # value "failed" over "abandoned" is the
                                     # soft half.
  audience:
    - developer
    - agent

agent:
  id: claude-opus-5
  session_id: null

project:
  id: PROJ-008
repo:
  id: bragfile

created_at: 2026-09-05
supersedes: null
superseded_by: null

tags:
  - schema
  - capture
  - type
---

# DEC-049: a failure is a reserved `type` value, pinned by a verb

## Decision

Work that did not work is recorded as an ordinary entry whose
`entries.type` is the reserved value **`failed`**, written by a dedicated
verb **`brag learn`** that does not expose `--type`.

Three parts, and the third is the one that carries the weight:

1. **The schema home is `entries.type`** — not a reserved tag, not a new
   column.
2. **The value is `failed`** — a new value, not a redefinition of the
   occupied `learned`.
3. **The verb pins the value; the field is NOT validated.** `brag add
   --type` stays free-form and accepts anything, as it does today. This
   decision adds one reserved value and one verb that always writes it. It
   does **not** introduce a closed set, a reserved-prefix rule, or any
   validation on `brag add`.

## Context

Measured on the live corpus 2026-09-05 (398 entries, `main` at `81e639d`):

- **Zero entries record a failure.** Not suppression — there was no verb, no
  value, and no documented convention for one.
- `entries.type` is free-form. `internal/cli/add.go:101` calls it *"free-form
  category"* and nothing validates it. **18 distinct non-empty values** across
  398 entries, and **113 entries carry no type at all**.
- The field already demonstrates what an unpinned convention does:
  **`shipped` 202 / `ship` 15**, and **`fixed` 2 / `bugfix` 1**. Near-duplicate
  spellings of the same idea, in live data, with no mechanism to have stopped
  them.
- **`learned` is occupied** — 8 entries. Reading all 8, they are lessons, and
  three of them (ids 80, 361, 383) take a *failure as their subject* while
  framing it as a win recovered (*"Verify caught an unpinned test seam…"*).
  One (id 18) is a PROJ-001 smoke-test artifact and not a lesson at all.
  Redefining `learned` would therefore mean `brag list --type learned`
  returning a mix of wins and failures — which fails the stage's own
  criterion that a failure be *retrievable as a failure*.

The read path is what settles part 1. DEC-044's memory line
(`internal/memory/memory.go:224`) is:

```
- <id> <YYYY-MM-DD> [<project>/<type>] <title> — <impact>
```

**`type` renders there; tags do not.** So a failure marked by `type` is
labelled as a failure on the agent-facing read surface at zero cost, while a
failure marked by a reserved tag is invisible there unless DEC-044's locked
line shape is amended. Verified by running it, not by reading it:

```
- 1 2026-09-05 [bragfile/failed] shared-worker pool did not cut cold starts — cost two days and produced nothing reusable
```

Retrieval also needs no new code: `--type` is an exact-match inclusion filter
on **seven** commands (`list`, `export`, `summary`, `impact`, `wrapped`,
`story`, `coverage`), so `brag list --type failed` answers *"what has not
worked?"* the day the verb ships.

## Alternatives Considered

- **A reserved tag (`outcome:failed`), in the `agent:`/`model:` family.**
  Cheaper than an early framing input assumed — DEC-015 normalized tags into
  `tags` + `taggings` with two indexes on 2026-06-06, so a reserved tag is an
  indexed lookup, not a `LIKE` scan. Rejected anyway, and on the read path
  rather than the write path: **tags do not render in the DEC-044 memory
  line**, so the agent-facing surface this stage exists to serve would not
  show it without amending the one line SPEC-073 stabilised.
- **A new column (`outcome`, or a boolean).** A forward-only migration
  (DEC-002) for a distinction `entries.type` already expresses, plus a new
  concept in every renderer and filter. Bought nothing the `type` value does
  not.
- **Redefining `learned`.** Collides with 8 rows of live user data (above).
  Re-typing them was considered and rejected separately: rewriting a user's
  corpus to fit a schema decision is a cost this work should not impose.
- **A documented convention with no verb** — i.e. tell people to type
  `brag add --type failed`. Rejected on the corpus's own evidence: `shipped`
  202 / `ship` 15 is what an unpinned convention produces here. The verb *is*
  the enforcement, scoped to one value.
- **Validating `type` globally** — a closed set or a reserved-prefix rule.
  Rejected as out of proportion and out of scope: it is a behaviour change on
  `brag add` for every existing user, it would reject `ship` (15 live entries)
  and `bugfix` (1) on write, and it owes an answer about the 113 typeless
  entries. It is also unnecessary — pinning ONE value needs one verb, not a
  validator. This is the choice that kept SPEC-085 an M rather than an L.
- **`abandoned` / `dead-end` as the value.** `failed` matches the field's
  existing house style — 7 of the 18 live values are past participles
  (`shipped`, `learned`, `fixed`, `designed`, `documented`, `hardened`,
  `verified`) — and reads correctly in the memory line's `[project/type]`
  slot. `abandoned` is narrower than the case (a wrong call that was carried
  to completion is not abandoned); `dead-end` reads poorly in that slot.

## Consequences

- **The verb and the value are deliberately different words.** `brag learn`
  writes `type: failed`. The verb names what the user is doing (extracting the
  lesson); the value names what is true of the entry (it did not work). The
  seam is real and is documented in the command's own help rather than
  smoothed over: naming the verb `brag fail` would describe the entry and
  discourage the use.
- **`brag learn` emits no milestone line.** `brag add` prints a
  congratulatory nudge to stderr on a TTY. Firing `🎉 N brags and counting —
  nice work!` at the moment someone records a failure is precisely the
  flattery this work exists to remove, so `insertLearned` does not call
  `emitMilestone`. Guarded by a paired assertion — `learn` silent AND `add`
  still firing on the same corpus at the same threshold.
- **One reserved value, not a validated field.** An agent or a user can still
  write `brag add --type failure`, and nothing stops them. The mitigation is
  documentation (`BRAG.md`, the `brag_add` MCP `type` parameter description)
  plus the verb being the obvious path — not a gate. Accepted knowingly.
- **The 8 `type: learned` entries are untouched.** A later rule that reads
  `learned` expecting failures will be wrong in 8 places; that is recorded
  here so it is a known boundary rather than a surprise.
- **The read path is NOT fixed by this decision.** `brag memory`'s candidate
  pool is the 200 most recent entries, so a failure stops being a candidate
  once 200 newer entries exist — measured at a 61-day horizon on a 398-entry
  corpus, with 198 entries already outside it. See
  `memory-pool-composition-excludes-older-entries` in
  `guidance/questions.yaml`. The line shape labels a failure correctly; the
  pool decides whether it is there to label.

## Validation

- `brag learn -t "…"` writes `entries.type = "failed"`; `brag list --type
  failed` returns it.
- `NewLearnCmd().Flags().Lookup("type")` is nil, and so is
  `ShorthandLookup("k")` — a structural check, not a help-text substring
  check, because the help text deliberately *does* contain `--type` (it
  teaches `brag list --type failed`).
- An editor buffer with a user-added `Type: shipped` header still stores
  `failed`.
- With the milestone TTY seam forced on and a 10th-entry threshold crossed,
  `brag learn` writes nothing to stderr while `brag add` writes
  `🎉 10 brags and counting`.

## References

- DEC-002 — migrations are forward-only (why a new column is expensive).
- DEC-015 — tags are normalized; supersedes DEC-004. Why the reserved-tag
  option was costed honestly and still rejected.
- DEC-043 / DEC-044 — the memory slice and its line shape. DEC-044's line is
  the fact that decides the schema home.
- SPEC-085 — the spec that makes this decision.
- SPEC-086 — what the celebratory digests do with a `failed` entry. Decided
  by the user at SPEC-085 design; implemented there.

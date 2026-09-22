---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-095
  type: story                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
                                   # Created at SPEC-094 framing (2026-09-22) to
                                   # CLAIM the id in the same edit as the
                                   # STAGE-023 line that routes the `summary`
                                   # half here. It carries SPEC-094 framing's
                                   # verdict for this half — GO, S — and the
                                   # measurement behind the split.
  blocked: false
  priority: high                   # gates v0.7.0 with SPEC-094: every failure in
                                   # a `summary` window is listed as a highlight
  complexity: S                    # one renderer file, one JSON key, a posture
                                   # DEC-050 row 3 already states in full

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-5-5
  implementer: claude-opus-5-5     # usually same Claude, different session
  created_at: 2026-09-22

references:
  decisions:
    - DEC-050                      # row 3 is this spec's posture, in full; it
                                   # implements it and decides nothing new
    - DEC-048                      # no count changes what it counts
    - DEC-014                      # the envelope, incl. part 4's empty state
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-086                     # the same fix on `impact` + `wrapped`
    - SPEC-094                     # the `story` half; split from it at framing
---

# SPEC-095: summary moves a failure out of Highlights

## Context

> **Cycle: frame.** This file exists so the id is claimed by a file rather
> than reserved in prose (SPEC-087 LD6). It was created at SPEC-094's framing,
> in the same edit as the STAGE-023 backlog line that points here. Nothing
> below is designed.

SPEC-094 was filed for `summary` **and** `story`, on the premise that both
*"already carry the data and drop it only in markdown."* Framing re-measured
that premise and it is true of `story` and **false of `summary`**. That is
why this half is its own spec.

Measured 2026-09-22 against a frozen copy of the live corpus
(`sqlite3 ~/.bragfile/db.sqlite ".backup …"`): 621 entries, 4 `type: failed`
(ids 420, 433, 465, 473), all 4 carrying an impact. Binary built from `main`
at `c0b840e`.

**`brag summary --range month`:**

- `## Summary → By type` prints `- failed: 4`, and JSON `counts_by_type`
  carries `"failed": 4`. That aggregate is honest, as DEC-050 row 3 says.
- `## Highlights` lists all four as `- <id>: <title>`. Measured with
  `/usr/bin/grep -c -E '^- (473|465|433|420): '` → **4**.
- **The JSON highlight entry is `{"id","title"}`, with no `type`.** So a
  failure cannot be told apart from a win *per entry in either format*. Only
  the aggregate count carries the type. That is the defect shape SPEC-086
  measured on `impact` and `wrapped` (DEC-028's 4-key projection has no
  `type`), **not** the lossy-markdown shape `story` has, where the JSON beat
  carries `"type": "failed"`.
- `--type failed` renders a `## Highlights` holding only the four failures:
  `bragfile` 3, `contextcore-pilot-harness` 1.

So this is SPEC-086's fix on a third surface, and DEC-050 row 3 already
states all of it:

| DEC-050 | What `summary` does |
|---|---|
| rule 1 | the predicate is `aggregate.IsFailure`; reuse it, do not restate it |
| rule 2 | a partition: the rows that leave `## Highlights` are exactly the rows that land in `## What didn't work` |
| rule 3 | `By type`, `By project`, `counts_by_type` and `counts_by_project` keep their definitions and still count both sections |
| rule 4 | the markdown section renders only when it has an entry; JSON always carries the key, as `[]` when empty |
| rule 5 | heading `## What didn't work`; JSON key `failures_by_project`, grouped as `highlights` is: `[{project, entries: [{id, title}]}]` |

## What framing found that design must settle

These are mechanics inside row 3's posture, not posture:

1. **The partition covers every in-window entry, not just the with-impact
   subset.** `summary`'s highlights list every entry, with or without an
   impact. So its section lists a failure that has no impact, which `impact`
   and `wrapped` do not (DEC-050's accepted T5 consequence). That difference
   follows from each surface's existing shape rather than from a new rule, and
   it costs nothing today: **0 of 4** failures lack an impact
   (`select count(*) from entries where type='failed' and impact=''` → 0).
   Design states it so nobody later reads it as drift.
2. **A failures-only window.** `summary --range month --type failed` would
   leave `## Highlights` with no entries. SPEC-086 pinned that a
   failures-only window renders no bare `## Impact`
   (`TestToImpactMarkdown_SectionsRenderOnlyWhenNonEmpty`), and the same rule
   applies here. What JSON `highlights` carries then is `[]`, per DEC-014
   part 4.
3. **The window cliff.** `--range month` is a **rolling 30 days**, not a
   calendar month. The four failures are dated 2026-09-06 to 2026-09-08, so
   they leave `summary --range month` on the live corpus around
   **2026-10-06**. After that, every NOT-contains check against the live
   corpus passes on a window with no failure in it. Verify must either run
   before then or measure against a seeded store with an injected `Now`, and
   **must pair every NOT-contains check with a positive one** either way.
4. **DEC-050 still routes row 3 to SPEC-094.** Its table's *Implemented by*
   column, its Context (*"SPEC-094 covers rows 3 and 4"*), its *"until
   SPEC-094 ships"* consequence and its References all predate the split.
   Framing does not edit a decision record, so design amends DEC-050 to name
   SPEC-095 for row 3, as SPEC-086's design amended DEC-028 and DEC-030. The
   posture itself stays untouched.

## Out of scope

- `story`: SPEC-094.
- `impact` and `wrapped`: SPEC-086 shipped them.
- **`brag review`**, which also lists all four failures as `- <id>: <title>`,
  measured with `brag review --month`. It is not one of DEC-050's seven
  `--type` surfaces (it has no `--type` flag), and it lists them under a
  neutral `## Entries` heading followed by reflection questions, not as
  highlights. It does not break DEC-050's invariant as written, so it is not
  routed here. It is named in SPEC-094's framing as an open question.
- `--type` negation, which is a separate `bug` entry on STAGE-023.

## Complexity

**S.** One renderer file (`internal/export/summary.go`), one new JSON key,
`aggregate.SplitFailures` reused as is, and no decision record: DEC-050 row
3 is complete. Its size and shape match one of SPEC-086's two surfaces, minus
the DEC-028/DEC-030 amendments, because `summary`'s envelope (DEC-014) has no
locked section arc. It also touches `docs/api-contract.md`'s `brag summary`
section (line 348) and the goldens.

## GO / NO-GO

**GO**, at complexity **S**. Nothing blocks it. It does **not** depend on
SPEC-094's open maintainer question (see SPEC-094 *Framing → T3*). That is
the practical reason to split: this half can go through design, build,
verify and ship while `story`'s mechanism waits on one confirmation.

**It gates v0.7.0** alongside SPEC-094: DEC-050 row 3 is half of what
STAGE-023's Success Criterion 4 still owes.

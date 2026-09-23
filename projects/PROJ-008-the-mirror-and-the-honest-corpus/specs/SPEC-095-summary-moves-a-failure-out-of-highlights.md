---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-095
  type: story                      # epic | story | task | bug | chore
  cycle: verify                    # frame | design | build | verify | ship
                                   # Designed 2026-09-22 against main 55063e9,
                                   # GO at S, held. Framing's record is kept
                                   # below as written.
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
  designed_at: 2026-09-22              # main 55063e9, after SPEC-094 shipped (#230)

references:
  decisions:
    - DEC-050                      # row 3 is this spec's posture, in full; it
                                   # implements it and decides nothing new.
                                   # AMENDED at this spec's design (row 3's
                                   # implementer, and its mechanics)
    - DEC-048                      # no count changes what it counts
    - DEC-014                      # the envelope, incl. part 4's empty state
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-086                     # the same fix on `impact` + `wrapped`
    - SPEC-094                     # the `story` half; split from it at framing
---

# SPEC-095: summary moves a failure out of Highlights

> **Cycle: design.** Designed 2026-09-22 against `main` at `55063e9`, after
> SPEC-094 shipped (#230). **GO at S, held.** Framing's record is kept below
> as it was written. Wherever design re-measured a number or overturned a
> claim, the design sections say so and take precedence. **Start at *What
> design settled*.**

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

---

## What design settled (2026-09-22, `main` at `55063e9`)

> Every live-corpus number was measured against a **frozen file copy** of
> `~/.bragfile/db.sqlite`, taken with `sqlite3 .backup`, which takes a read
> lock only. Its SHA-256 (`c93395b6e9bc…`) was the same before and after every
> run against it, and the live corpus was never opened for writing. Two
> binaries ran against it: `brag-main`, built from `55063e9`, and
> `brag-proto`, built from a prototype worktree at `55063e9` carrying every
> literal in *Notes for the Implementer*.

- **Mechanic 1: the partition covers every in-window entry.**
  `summary` calls `aggregate.SplitFailures(entries)` on the whole in-window
  slice, never on `aggregate.WithImpact(entries)`. So `## What didn't work`
  on `summary` lists a failure with **no** impact, which `impact` and
  `wrapped` do not. That follows from `summary`'s highlights listing every
  entry, and it is not drift. It is pinned by
  `TestToSummary_PartitionCoversEveryInWindowEntry` and by the end-to-end
  test, whose `brag learn` row deliberately carries no impact. Re-measured:
  **0 of 4** failures lack an impact.
- **Mechanic 2: a window holding only failures renders no bare
  `## Highlights`.** Each body section renders only when it has an entry.
  JSON `highlights` is then `[]` (DEC-014 part 4), and `failures_by_project`
  is always present. Pinned by `TestToSummary_SectionsRenderOnlyWhenNonEmpty`.
  An empty *document* is unchanged: provenance only.
- **Mechanic 3: the window cliff. Chosen: seeded stores, and no Go test
  or acceptance criterion reads the live corpus.** `summary` has **no
  `--since` flag**, so SPEC-094's `--since 2026-09-01` escape does not
  exist here. The Go tests render fixed fixtures through the renderer, and
  `summary` gets no clock injection because none is needed. The end-to-end
  test writes its rows with `brag add` and `brag learn` at run time and
  reads `--range week`, so its window always holds them. The acceptance
  criteria do the same against a scratch `--db`. **Every NOT-contains is
  paired with a positive on the same section.** The live-corpus measurement
  below is design's record for 2026-09-22 and is dated. Its cliff is exact:
  id 420 leaves `--range month` at **2026-10-06T00:44:44Z**, and the last
  failure, id 473, leaves at **2026-10-08T18:50:42Z**.
- **Mechanic 4: DEC-050 is amended, and there is no new record.** SPEC-094's
  design had already moved row 3's *Implemented by* and the *"until SPEC-094
  ships"* consequence (its `## Amendment (2026-09-22, SPEC-094 design)`).
  This commit adds `## Amendment (2026-09-22, SPEC-095 design)`. It re-points
  the three sentences still naming SPEC-094 alone for row 3 (Context,
  Validation, References) and records the mechanics above. The posture and
  the original text are untouched. **DEC-055 is not claimed.**
- **The JSON change:** keys go from `generated_at, scope, filters,
  counts_by_type, counts_by_project, highlights` to the same six plus
  `failures_by_project` last. The new key has the same
  `{project, entries: [{id, title}]}` shape as `highlights`. Both count maps
  are byte-identical to `main` on the frozen copy.

---

## Re-measurement at design (2026-09-22, `main` at `55063e9`)

Design does not inherit framing's numbers. Before any equality was believed,
each side was checked: the first line is `# Bragfile Summary`, stderr is
empty, and every JSON file parses (written to files, never through zsh
`echo`).

| Framing said (2026-09-22, `c0b840e`) | Measured at design (`55063e9`) |
|---|---|
| 621 entries, 4 `type: failed` (420, 433, 465, 473) | **623 entries**, max id 645. The two new rows are 644 and 645, the SPEC-086 and SPEC-094 ship brags, neither a failure. `lower(type) like '%fail%'` → only `failed\|4`, the same four ids. |
| 0 of 4 failures lack an impact | **Same**: `select count(*) from entries where type='failed' and impact=''` → `0` |
| `summary --range month`: `- failed: 4` under By type, JSON `counts_by_type.failed` 4 | **Same** |
| `/usr/bin/grep -c -E '^- (473\|465\|433\|420): '` → 4, all under `## Highlights` | **Same**: 4, all under `## Highlights` |
| JSON highlight entry is `{id, title}` | **Same**. Keys: `generated_at,scope,filters,counts_by_type,counts_by_project,highlights` |
| `--type failed` → a `## Highlights` of only the four: `bragfile` 3, `contextcore-pilot-harness` 1 | **Same** |
| (not measured) | `--range month` holds 233 entries. `--range week` holds **0** failures. |

**What the prototype does to the same copy:**

| Cell | `brag-main` | `brag-proto` |
|---|---|---|
| `--range month`, markdown | 272 lines; the four failures under `## Highlights` | 280 lines. `diff` is exactly 4 deletions from `## Highlights` plus a 12-line `## What didn't work` appended (`bragfile`: 433, 465, 473; `contextcore-pilot-harness`: 420). Nothing else moves. |
| `--range month`, By type | `- failed: 4` | `- failed: 4` |
| `--range month`, JSON counts | `counts_by_type` and `counts_by_project` | **equal** (`jq` `==` → `true`; `counts_by_project` hash `1c7a76e430e4` on both) |
| `--range month`, JSON `highlights` | 233 entries | equals `main`'s with ids 420/433/465/473 removed and emptied groups dropped (`cmp` of the `jq -c` projections) |
| `--range month`, JSON `failures_by_project` | key absent (`jq` iterating it errors on `null`, a shape check) | `[{bragfile:[433,465,473]}, {contextcore-pilot-harness:[420]}]` |
| `--range month --type failed` | `## Summary`, `## Highlights` | `## Summary`, `## What didn't work`. JSON `"highlights": []`, 2 failure groups, `counts_by_type` `{"failed":4}` |
| `--range week` (0 failures) | 68 lines | **byte-identical** with `Generated:` removed (`cmp`); JSON identical after deleting `generated_at` and the new key, which is `[]` |

The `--range week` equality is a real measurement, not a default. Both sides
were 68 lines, started `# Bragfile Summary`, had empty stderr, and the window
was confirmed to hold no failure.

---

## The four mechanics — settled

### 1. The partition covers every in-window entry

`impact` and `wrapped` split `aggregate.WithImpact(entries)`, because their
celebratory section only ever held the with-impact subset. `summary`'s
`## Highlights` holds **every** in-window entry, so rule 2 (a partition,
not a filter) splits every entry. Splitting the with-impact subset here
would leave an impact-less failure under `## Highlights`, still listed as a
highlight, which is the defect. Mutant M-4 is that shortcut, and five tests
catch it.

So the one visible asymmetry is this: an impact-less failure is **listed**
by `summary` and **counted but not listed** by `impact` and `wrapped`. It is
stated in DEC-050's new amendment, in `SplitFailures`'s doc comment, in the
contract, and in the CHANGELOG, so nobody later reads it as drift. It costs
nothing today, because 0 of 4 failures lack an impact.

### 2. A window holding only failures

`summary --type failed` today prints a `## Highlights` holding only failures.
After this spec it prints `## Summary` then `## What didn't work`, with no
bare `## Highlights`, the same rule SPEC-086 pinned for `## Impact`. JSON:
`"highlights": []`. The whole-document empty state (no entries in window)
is unchanged: provenance only, both arrays `[]`.

### 3. The window cliff

| Surface | Loses the four failures | Used here? |
|---|---|---|
| `summary --range month` on the live corpus | 420 at 2026-10-06T00:44:44Z … 473 at 2026-10-08T18:50:42Z | design-time record only |
| `summary --range week` on the live corpus | already (0 failures on 2026-09-22) | as the *no difference* control |
| a scratch store written by `brag add` / `brag learn` at run time, read with `--range week` | **never**, because its rows are always minutes old | the end-to-end Go test and AC-1 to AC-5 |
| Go renderer fixtures (`summaryFailureFixture`, dated April 2026, `Now` fixed) | never, because the renderer receives already-windowed entries | every export test |

`summary` has no `--since`, so the anchor SPEC-094 used does not exist. No
clock seam is added: `runSummary` reads `time.Now()`, and a store written at
run time needs none. **Verify may still re-run the live-corpus rows above
before 2026-10-06** as a real-data cross-check. After that date the rows
pass vacuously, and verify should say so rather than cite them.

### 4. DEC-050

See *Mechanic 4* above, and the record's
`## Amendment (2026-09-22, SPEC-095 design)`. It is written in this commit,
not at build.

---

## §12(b) design-time pre-flight: what the tools actually said

Every literal in *Notes for the Implementer* ran through its real tool:
`go build`, `gofmt -l` (empty), `go vet` (clean), `just lint` (**0
issues**), and `go test -count=1 ./...` (14 packages `ok`, **1119** passing
tests counting subtests, up from 1113). The help sentence ran through
cobra's `--help`, the markdown and JSON through `brag-proto` on the frozen
copy, and the doc hunks and Group `AF` through `./scripts/test-docs.sh`
(ALL OK, **213** `OK:` lines, up from 210). The inventory ran through
`scripts/inventory.sh`.

**The literals are `git diff` output from the prototype, base `55063e9`.**
They are 11 files and 22 hunks. `internal/export/summary_test.go` is diffed
at `-U12` and the rest at the default `-U3`, because at `-U3` the test
file's end-of-file append had a 3-line context that occurs 3 times. **Every
hunk's old side occurs exactly once in its base file**, checked by a script
that rebuilt each hunk's context-plus-`-` text and counted it in
`git show 55063e9:<file>`: 22 of 22 unique. `git apply --check` accepts the
whole patch against this design branch. Applied to a pristine
`git archive 55063e9`, it reproduces all 11 prototype files byte for byte
(hashes below), and `go test ./...` is green there.

**Finding 1: fail-first was run, not predicted.** The three new and modified
test files were copied onto an unmodified `git archive 55063e9`:

```
internal/export  FAIL  6: TestToSummaryJSON_DEC014ShapeGolden, TestToSummary_EmptyEntriesEmitsProvenanceOnly
                           (the 2 planned rewrites), and all 4 new export tests
internal/cli     FAIL  2: TestLearnCmd_SummarySectionsWhatItWrote, TestSummaryCmd_HelpNamesTheFailureSection
(every other package ok; nothing failed to compile, so no stubs are needed)
```

`TestToSummaryMarkdown_DEC014FullDocumentGolden` **passes on `main` and
after the change**, byte for byte. Its fixture holds no failure, so the
markdown is unchanged. That is a measured *no difference*, and the test is
now also the clean-window guard: M-3 turns it red.

**Finding 2: one probe was a compile red, and it was discarded, not
credited.** M-5 (`highlightGroups(entries)`) left `worked` unused in
`ToSummaryJSON`, so four packages failed to build. That is a red for the
wrong reason (SPEC-094 Finding 3). It was re-run as M-5′, which compiles
and is killed by four tests.

**Finding 3: M-D2 survived, and the probe was wrong, not the guard.** The
contract's `summary` section names `## What didn't work` **twice**, once as
the bullet head and once in the omission sentence. Renaming one leaves
`AF1` green, correctly: `AF1` asserts the section *names* the heading. M-D2′
renames both and `AF1` fires. This is SPEC-094's Finding 4 again. **AF1 does
not check every mention**, and nothing here claims it does.

**Finding 4: the one `test-docs` failure on the archive export is AC2, and
it is correct.** AC2 fails on a tree that tracks no files (*"git ls-files
listed no tracked files"*), so the pristine-export run reports it. On the
prototype worktree, a real checkout, every assertion is OK.

**Finding 5: the NOT-contains self-audit.** Every NOT-contains this spec
adds is paired with a positive on the same section (LD5): `- <failID>:`
absent from `## Highlights` alongside the win present there, and
`- <winID>:` absent from `## What didn't work` alongside the failure present
there. No NOT-contains reads help text or the live corpus.

### Mutation matrix: 17 probes, each confirmed by content hash before its gates ran

The probe helper (`probe.py`, in the design session's scratchpad) **refused
any edit whose old text does not occur exactly once** (the SPEC-094-ship
clause, as a guard). It **refused to run the gates until the file's hash
had moved**, then ran `go test -count=1 ./...` and
`./scripts/test-docs.sh`, restored the file, and confirmed the hash was back
at baseline. Baselines, on the prototype (= `55063e9` + the literals):
`internal/export/summary.go` `3b81e79b3fdf`, `internal/cli/summary.go`
`6bf6857adafd`, `docs/api-contract.md` `3da3b8d8a35f`, `AGENTS.md`
`5fb3a6d60c68`, `CHANGELOG.md` `33c7ebd14725`. **Every target was back at
its baseline at the end.** Each *Diff* cell's old text occurs exactly once
in its baseline file, so each row has one literal reading. *(Verify, V-F1:
true of every old side, and false of M-9's new side, which was written as a
sketch. The row now states it literally.)*

| # | File | Diff (text replaced → replacement) | Hash | Fired |
|---|---|---|---|---|
| **M-1** | `export/summary.go` | `worked, failed := aggregate.SplitFailures(entries)` + next line `if len(worked) > 0 {` → `worked, failed := entries, []storage.Entry{}` + the same next line (markdown never splits) | `3b81e79b3fdf`→`9856d2b3899b` | `…FailureSectionGolden` (md), `…PartitionCoversEveryInWindowEntry`, `…SectionsRenderOnlyWhenNonEmpty`, the `learn` e2e |
| **M-2** | `export/summary.go` | `if len(worked) > 0 {` → `if true {` (a bare `## Highlights` on a failures-only window) | →`2d11f9555c90` | **only** `TestToSummary_SectionsRenderOnlyWhenNonEmpty` |
| **M-3** | `export/summary.go` | `if len(failed) > 0 {` → `if true {` (an empty `## What didn't work` on a clean window) | →`633c6867b377` | `…SectionsRenderOnlyWhenNonEmpty`, and the **existing** `TestToSummaryMarkdown_DEC014FullDocumentGolden` |
| **M-4** | `export/summary.go` | `worked, failed := aggregate.SplitFailures(entries)` + next line `if len(worked) > 0 {` → `worked, failed := aggregate.SplitFailures(aggregate.WithImpact(entries))` + the same next line (**mechanic 1's shortcut**) | →`f1fdbcf9dea8` | `…FailureSectionGolden` (md), `…PartitionCoversEveryInWindowEntry`, the e2e, and the **existing** `…DEC014FullDocumentGolden` and `TestSummaryCmd_ScopeFieldAndMarkdownDefault` |
| ~~M-5~~ | `export/summary.go` | `env.Highlights = highlightGroups(worked)` → `env.Highlights = highlightGroups(entries)` | →`2754812dd137` | **Discarded**: `worked` unused, a compile red in 4 packages (Finding 2) |
| **M-5′** | `export/summary.go` | `env.Highlights = highlightGroups(worked)` → `env.Highlights = highlightGroups(append(worked, failed...))` (JSON filters instead of partitioning: a failure in both keys) | →`48307e8815e0` | `…FailureSectionGolden` (JSON), `…PartitionCoversEveryInWindowEntry`, `…SectionsRenderOnlyWhenNonEmpty`, the e2e |
| **M-6** | `export/summary.go` | `env.FailuresByProject = highlightGroups(failed)` → `env.FailuresByProject = highlightGroups(failed[:0])` (JSON drops failures silently) | →`8dcf174e9e81` | the same four |
| **M-7** | `export/summary.go` | `` `json:"failures_by_project"` `` → `` `json:"failures"` `` | →`fdf06c45d5ec` | 6, including the **rewritten** `TestToSummaryJSON_DEC014ShapeGolden` and `TestToSummary_EmptyEntriesEmitsProvenanceOnly` |
| **M-8** | `export/summary.go` | `out := []highlightGroup{}` → `var out []highlightGroup` (both arrays `null` when empty) | →`da5c7a31434b` | `…DEC014ShapeGolden`, `…EmptyEntriesEmitsProvenanceOnly`, `…SectionsRenderOnlyWhenNonEmpty` |
| **M-9** | `export/summary.go` | the one line `env.CountsByType[tc.Type] = tc.Count` (`:134`, two tabs) → the three gofmt lines `if tc.Type != "failed" {` / `env.CountsByType[tc.Type] = tc.Count` / `}` at two, three and two tabs (**a count stops counting a section**, DEC-048). Base `3b81e79b3fdf` as of `6590772` on `main`. *Pinned at verify (V-F1): as designed, the new side read "wrapped in `if tc.Type != "failed" { … }`", which build reproduced as a one-line wrap, `c840a90180f2`.* | →`e9404d9921ee` | `…FailureSectionGolden` (JSON), `…PartitionCoversEveryInWindowEntry`, the e2e |
| **M-10** | `cli/summary.go` | ` The by-type and by-project counts still include it.` → *(empty)* | `6bf6857adafd`→`32972312a7d7` | **only** `TestSummaryCmd_HelpNamesTheFailureSection` |
| **M-D1** | `docs/api-contract.md` | `` `{project, entries:[{id, title}]}` groups, failures excluded), and `` + next line `` `failures_by_project` (the same shape, holding only the failures). `` → the same first line + `a second array (the same shape, holding only the failures).` | `3da3b8d8a35f`→`e2e9c98f0528` | `AF1` |
| M-D2 | `docs/api-contract.md` | `` - **`## What didn't work`** (markdown): the in-window entries whose `` → `- **What didn't work** (markdown): the in-window entries whose` | →`0fbc82a55a63` | **none**: the probe was too weak (Finding 3) |
| **M-D2′** | `docs/api-contract.md` | the 8-line block from `` - **`## What didn't work`** (markdown): the in-window entries whose `` through `  heading,` (the first line after `` a window with no recorded failure has no `## What didn't work` ``), with **both** `## What didn't work` → `## What did not work` | →`ec00cdb2ae0a` | `AF1` |
| **M-D3** | `docs/api-contract.md` | `` ### `brag summary --range week|month` (STAGE-004) `` → `` ### `brag  summary --range week|month` (STAGE-004) `` (two spaces) | →`10ef99590a1b` | `AF1: … section not found`, the non-vacuity branch |
| **M-D4** | `AGENTS.md` | `` `brag summary` lists every failure in its window `` → `` `brag summary` lists each failure in its window `` | `5fb3a6d60c68`→`3e34c51a9a38` | `AF2` |
| **M-D5** | `CHANGELOG.md` | `` `brag summary --format json` no longer lists failures in `` → `` `brag summary --format json` drops failures from `` | `33c7ebd14725`→`2ec7cc1682c2` | `AF3` |

**Two probes matter most.** M-4 is mechanic 1's exact shortcut, copying
`impact`'s `SplitFailures(WithImpact(…))`. Five tests catch it, two of them
pre-existing, because it also drops impact-less *wins* from `## Highlights`.
M-2 is caught **only** by the section test, because every golden fixture
with a failure also has a win.

### Inventory: regenerated and diffed, not predicted

`scripts/inventory.sh` was run on three trees and diffed: the block in
`docs/engineering-practices.md` at `55063e9`, this design branch, and the
prototype.

| Row | `main` | this design commit | after build |
|---|---:|---:|---:|
| Decision records | 53 | 53 | 53 |
| …of those, carrying an explicit `## Amendment` section | ~~5~~ **6** | ~~5~~ **6** (DEC-050 already carried one) | ~~5~~ **6** |
| Go test files | 81 | 81 | 81 |
| Go test functions | 856 | 856 | **862** |
| Documentation assertions (distinct ids) | 209 | 209 | **212** |

Every other row is unchanged. The design commit changes no inventory row,
so `docs/engineering-practices.md` is **not** in it. Build regenerates the
block with `just inventory`, and its literal is below. Passing tests
counting subtests (`go test -v ./... | grep -c -- '--- PASS'`) go from 1113
to **1119**, and `test-docs` `OK:` lines from 210 to **213**.

### §9(b): the harness grepped by VALUE

```
$ for v in '| 856 |' '| 862 |' '| 209 |' '| 212 |' '!=856' '!=209' '1113' '1119' '210' '213'; do
    /usr/bin/grep -n -F -- "$v" scripts/test-docs.sh | /usr/bin/grep -v '^[0-9]*:[[:space:]]*#'
  done
(no hits for any value)
$ /usr/bin/grep -rn --include='*_test.go' -E '\b(856|1113|209)\b' internal cmd
(no hits)
```

**No literal pin exists for any value that moves.** Only `X3`'s page block
moves, and it is regenerated.

---

## Locked design decisions

**LD1 — the partition covers every in-window entry.** Both renderers call
`aggregate.SplitFailures(entries)` on the full slice they receive, never on
`aggregate.WithImpact(entries)`. A failure with no impact is listed under
`## What didn't work` and in `failures_by_project`. The predicate is
`aggregate.IsFailure`, reached only through `SplitFailures` and never
restated (DEC-050 rule 1).

**LD2 — each body section renders only when it has an entry.**
`## Highlights` renders iff at least one non-failure is in the window, and
`## What didn't work` iff at least one failure is. The order is `## Summary`
→ `## Highlights` → `## What didn't work`. An empty document is unchanged:
the provenance block only, as today.

**LD3 — no count changes what it counts** (DEC-050 rule 3, DEC-048).
`By type`, `By project`, `counts_by_type` and `counts_by_project` are still
computed from `entries`, before the split, and count both sections. No
per-section count line is added (DEC-048 Alternative 2).

**LD4 — the JSON key.** `failures_by_project`, declared **after**
`highlights` in `summaryEnvelope`, so it is the last key. It is typed
`[]highlightGroup`, the same 2-key `{id, title}` entry as `highlights`, and
no `type` is added to either. It is always present, as `[]` when empty.
Both arrays come from one helper, `highlightGroups`, which returns a
non-nil slice.

**LD5 — no test or criterion depends on the date.** Export tests use
fixtures. The end-to-end test and the acceptance criteria write their rows
at run time into a scratch store and read `--range week`. **Every
NOT-contains is paired with a positive on the same section.** No clock seam
is added to `runSummary`.

**LD6 — `--help` says where a failure goes.** One sentence is appended to
`Long` as its own paragraph after the `--range` paragraph:
`Work recorded with brag learn is listed under its own "What didn't work" heading instead of among the highlights, and that heading is left out when there is none. The by-type and by-project counts still include it.`
The first sentence is `impact`'s (SPEC-086 LD9) with *impact* → *highlights*.

**LD7 — one markdown helper.** `writeHighlightGroups(buf, entries)`
renders `### <project>` blocks of `- <id>: <title>` for both sections,
through `aggregate.GroupForHighlights`, so a failure reads in the shape it
had as a highlight.

**LD8 — `SplitFailures`'s doc comment names both call shapes.** It said
*"The digests call it on the with-impact subset"*, which stopped being
true at SPEC-094 (`story.OmitFailures` calls it on every entry). It now
says `impact` and `wrapped` split the with-impact subset, and `story` and
`summary` split every in-window entry. No code in `aggregate` changes.

**LD9 — docs and guards.** `docs/api-contract.md`'s `brag summary` section
documents the section, the asymmetry, and the full JSON key list, which
it never had. `CHANGELOG.md` `[Unreleased]` → `### Changed` gains two
bullets (the markdown change, then the **Breaking** JSON change).
`AGENTS.md`'s `learn` glossary entry names `summary`. `scripts/test-docs.sh`
gains Group `AF` (3 ids), scoped with Group AD's helpers.

**LD10 — records.** DEC-050 is amended in the design commit. No new
record. **DEC-014 is not amended**: its choice 2 lists `summary`'s keys as
an example of flat payload keys and says *"per-spec payload keys are
documented in each consuming spec"*. A new flat key is additive under
choice 2, and choice 4 already requires `[]` for it. The contract and
DEC-050's amendment carry the new key list.

### Rejected alternatives (build-time)

- **Add `type` to each highlight entry, and leave failures in place.** It
  fixes JSON and leaves markdown unlabelled. DEC-050's posture for row 3 is
  a section, not a label.
- **Split only the with-impact subset, as `impact` does.** That is M-4. An
  impact-less failure stays under `## Highlights` as a win.
- **Keep a bare `## Highlights` heading on a failures-only window.** It
  breaks rule 4, and a heading over nothing reads as *"nothing went well"*
  where it means *"nothing but failures was recorded"*.
- **Put `failures_by_project` before `highlights`.** It would change the
  position of an existing key. Appending keeps every existing key's
  position, which DEC-014's key-order test pins.
- **Add a clock seam so a fixed-date store can be read with
  `--range month`.** It is test-only surface (SPEC-018 rejected the same for
  `rangeCutoff`), and a store written at run time needs none.

---

## Outputs

### New files (0)

None. Every new test goes into an existing file.

### Modified files, at build (11)

| File | Change |
|---|---|
| `internal/export/summary.go` | LD1–LD4, LD7: split, two sections, the new key, and two helpers |
| `internal/export/summary_test.go` | 4 new tests, the new fixture, and 2 planned rewrites (below) |
| `internal/aggregate/aggregate.go` | LD8: `SplitFailures` doc comment only |
| `internal/cli/summary.go` | LD6: the `--help` sentence |
| `internal/cli/summary_test.go` | 1 new test |
| `internal/cli/learn_test.go` | 1 new test; `runDigestCorpus` registers `NewSummaryCmd()` and its comment says so |
| `docs/api-contract.md` | the `brag summary` section: 2 hunks |
| `CHANGELOG.md` | `[Unreleased]` → `### Changed`: 2 bullets |
| `AGENTS.md` | the `learn` glossary entry: 1 sentence |
| `scripts/test-docs.sh` | Group `AF`, 3 ids |
| `docs/engineering-practices.md` | the inventory block, **regenerated by `just inventory`, never hand-edited**: 856 → 862, 209 → 212 |

### Modified or created at design (this commit)

| File | Change |
|---|---|
| this spec | `cycle: design`, and everything from *What design settled* on |
| `decisions/DEC-050-a-failure-is-never-rendered-as-a-win.md` | `## Amendment (2026-09-22, SPEC-095 design)` |
| `projects/PROJ-008-…/stages/STAGE-023-…md` | SPEC-095's backlog line: `(frame)` → `(design)`, plus a design note |

### The CHANGELOG entries

`[Unreleased]` → `### Changed` gains two bullets, directly after SPEC-086's
breaking `impact`/`wrapped` JSON bullet and before SPEC-094's `story`
bullets. The breaking one reads:

> **Breaking: `brag summary --format json` no longer lists failures in
> `highlights`.** The envelope gains `failures_by_project`, after
> `highlights`, in the same `{project, entries:[{id, title}]}` shape. It is
> always present and `[]` when empty. `counts_by_type` and
> `counts_by_project` are unchanged. A script that read every entry from
> `highlights` now reads both keys.

### Premise audit (§9), run at design against the repo

**Case 1: inversion or removal → planned test rewrites.** The change
removes failures from `highlights` and adds a key. Found **by execution**
(the renderer change alone, then the fail-first run), not by grep:

| Test | Why it moves | Plan |
|---|---|---|
| `TestToSummaryJSON_DEC014ShapeGolden` | the envelope gains `"failures_by_project": []`, and the key-order walk gains a 7th key | **rewrite**: golden + `wantKeys` |
| `TestToSummary_EmptyEntriesEmitsProvenanceOnly/json` | the empty envelope gains `"failures_by_project": []` | **rewrite**: golden |
| `TestToSummaryMarkdown_DEC014FullDocumentGolden` | its fixture holds no failure | **unchanged**, measured byte-identical |
| every `internal/cli` summary test | none of their stores holds a failure | **unchanged**, measured green |

`runDigestCorpus` (a helper, not a test) gains `NewSummaryCmd()`.

**Case 2: addition → planned count bumps.** Go test functions 856 → 862,
distinct `test-docs` ids 209 → 212, passing tests 1113 → 1119, and
`test-docs` `OK:` lines 210 → 213. There is no literal pin on any of them
(*§9(b)* above). Only the regenerated inventory block moves.

**Case 3: status change → doc references, grepped by value:**

| Grep | Hits | Lands |
|---|---|---|
| `/usr/bin/grep -n -i 'brag summary\|summary --range' docs/*.md README.md AGENTS.md` | `docs/api-contract.md:348` (the section) | **updated**: 2 hunks (LD9) |
| same | `docs/tutorial.md:496`, `:517`, `:536`: each names `summary` in passing (*"the right command if you want filter composition"*, *"windowed digests"*, *"rolling windows"*) | **no change**: none describes its sections or keys. The tutorial has no `summary` section. |
| same | `README.md:222-223`: two example commands | **no change**: no shape claim |
| same | `AGENTS.md:290` (**digest**), `:296` (**summary**) | **no change**: *"grouped by project/type"* still holds |
| `/usr/bin/grep -n -- '- \*\*learn\*\* —' AGENTS.md` | `AGENTS.md:294`, which names `impact`, `wrapped` and `story` as readers of a failure | **updated** (LD9), pinned by `AF2` |
| `git grep -n 'Highlights\|"highlights"\|\.highlights'` outside `projects/`, `.claude/worktrees/` and tests | `internal/export/summary.go`, `internal/aggregate/aggregate.go` (`GroupForHighlights`), `guidance/questions.yaml:260-271` (a closed question on grouping axes) | the renderer is this spec. `questions.yaml` is out of scope and still true: project is the only grouping axis, in both sections. |
| `git grep -n 'summary --range\|counts_by_type'` outside `projects/` | `BRAG.md:438`, `docs/blog/why-bragfile.md:84`, `CHANGELOG.md:617` (0.1.0's entry): commands only. `DEC-014:50,55,151` | **no change**. DEC-014: see LD10. |
| the `--help` text | `internal/cli/summary.go` `Long` | **updated** (LD6), pinned by `TestSummaryCmd_HelpNamesTheFailureSection` |
| `git grep -n 'ToSummary\|NewSummaryCmd'` outside tests | `cmd/brag/main.go:47` and the renderer. No MCP tool, plugin or skill calls `summary`. | no other consumer |

---

## Acceptance Criteria

Every criterion runs against a **scratch store written at run time**, never
the live corpus (LD5). Set up once, with `B` a binary built from the build
branch:

```sh
T=$(mktemp -d)
W=$($B --db "$T/db.sqlite" add -t "shipped the cache" -p alpha -k shipped)
F=$($B --db "$T/db.sqlite" learn -t "tried a worker pool" -p alpha)   # no impact, on purpose (LD1)
$B --db "$T/db.sqlite" summary --range week > "$T/s.md"
$B --db "$T/db.sqlite" summary --range week --format json > "$T/s.json"
```

First check each artifact's shape: `head -1 "$T/s.md"` is
`# Bragfile Summary`, and `jq -e . "$T/s.json"` succeeds.

- [ ] **AC-1: a failure leaves `## Highlights` for `## What didn't work`,
  with or without an impact.** Using `awk` over `$T/s.md` between `##`
  headings: `- $F: tried a worker pool` is **under `## What didn't work`**
  and **not under `## Highlights`**, and `- $W: shipped the cache` is
  **under `## Highlights`** and **not under `## What didn't work`**. That is
  two positives, each paired with a negative on the same section.
- [ ] **AC-2: the JSON partitions.**
  `jq -c '[.highlights[].entries[].id]'` → `[$W]`,
  `jq -c '[.failures_by_project[].entries[].id]'` → `[$F]`, and
  `jq -r 'keys_unsorted|join(",")'` →
  `generated_at,scope,filters,counts_by_type,counts_by_project,highlights,failures_by_project`.
- [ ] **AC-3: no count changes what it counts.** `$T/s.md` contains
  `- failed: 1` and `- shipped: 1` under `**By type**`, and `- alpha: 2`
  under `**By project**`. `jq '.counts_by_type.failed'` → `1` and
  `jq '.counts_by_project.alpha'` → `2`.
- [ ] **AC-4: a failures-only window.**
  `summary --range week --type failed` prints exactly the `##` headings
  `## Summary`, `## What didn't work`, with no `## Highlights`. Its JSON has
  `.highlights == []` → `true` and `.failures_by_project | length` → `1`.
- [ ] **AC-5: a clean window.** `summary --range week --type shipped` prints
  exactly `## Summary`, `## Highlights`, with no `## What didn't work`. Its
  JSON has `.failures_by_project == []` → `true` and `.highlights | length`
  → `1`.
- [ ] **AC-6: `--help`** contains LD6's sentence verbatim.
- [ ] **AC-7: the Go suite:** the 6 new tests and the 2 rewrites in *Failing
  Tests* exist under those names and pass. `TestToSummaryMarkdown_DEC014FullDocumentGolden`
  is unmodified: `git diff main -- internal/export/summary_test.go` leaves
  its `want` literal alone.
- [ ] **AC-8: docs:** `./scripts/test-docs.sh` prints `OK:   AF1`, `AF2` and
  `AF3`, and `ALL OK`. The inventory block equals `just inventory`'s
  output (`X3`).
- [ ] **AC-9: all five gates:** `just test`, `just test-docs`, `just lint`,
  `gofmt -l .` (empty) and `go vet ./...`.
- [ ] **AC-10: DEC-050** carries `## Amendment (2026-09-22, SPEC-095 design)`
  and its text above that heading is unchanged from `55063e9` (written at
  design; build only checks it).

---

## Failing Tests

Written at build first, and observed failing against the unmodified
renderer (design ran this: *Finding 1*).

### New: `internal/export/summary_test.go` (4)

- `TestToSummaryMarkdown_FailureSectionGolden` (LOAD-BEARING): the whole
  document on `summaryFailureFixture`, byte for byte.
- `TestToSummaryJSON_FailureSectionGolden` (LOAD-BEARING): the whole
  envelope, byte for byte.
- `TestToSummary_PartitionCoversEveryInWindowEntry`: mechanic 1 and rules
  2 and 3. The ids under each section in both formats are `1,2,3,4,5` and
  `6,7,8`, id 7 being the impact-less failure, and both count maps sum to 8.
- `TestToSummary_SectionsRenderOnlyWhenNonEmpty`: mechanic 2 and rule 4.
  The `##` headings, and each JSON key's group count (never `null`), on no
  failures, failures only, and both.

### New: `internal/cli` (2)

- `TestLearnCmd_SummarySectionsWhatItWrote` (`learn_test.go`): writer to
  reader through a real store, with an impact-less `brag learn` row and
  `--range week`. Every negative is paired.
- `TestSummaryCmd_HelpNamesTheFailureSection` (`summary_test.go`): LD6.

### Changed, as planned rewrites (2 tests, 1 helper)

- `TestToSummaryJSON_DEC014ShapeGolden`: golden gains
  `"failures_by_project": []`, and `wantKeys` gains it last.
- `TestToSummary_EmptyEntriesEmitsProvenanceOnly`, `json` subtest: golden
  gains `"failures_by_project": []`.
- `runDigestCorpus` (helper): registers `NewSummaryCmd()`.

### New: `scripts/test-docs.sh`, Group `AF` (3 ids)

- `AF1`: the contract's `### \`brag summary` section names
  `## What didn't work` and `` `failures_by_project` ``.
- `AF2`: `AGENTS.md`'s `- **learn** —` line names
  `` `brag summary` lists every failure ``.
- `AF3`: `CHANGELOG.md` `[Unreleased]` names
  `` `brag summary --format json` no longer lists failures ``.

### Mutation checks

Build re-runs the 16 non-discarded probes in the matrix, each from its
stated edit, and records the hashes it reached. A hash that differs from
design's while the same tests fire is reported, not hidden.

### Decision-to-test mapping (§9)

| Decision | Test |
|---|---|
| DEC-050 rule 1 (the predicate) | reached only through `SplitFailures`, which has `TestSplitFailures_PartitionsInOrderNeverNil` and `TestFailureClassifier_GoPredicateMatchesTypeFilter` (unchanged) |
| DEC-050 rule 2 / LD1 | `TestToSummary_PartitionCoversEveryInWindowEntry`, `TestLearnCmd_SummarySectionsWhatItWrote` |
| DEC-050 rule 3 / LD3 / DEC-048 | `TestToSummary_PartitionCoversEveryInWindowEntry` (the count sums), both failure goldens, M-9 |
| DEC-050 rule 4 / LD2 | `TestToSummary_SectionsRenderOnlyWhenNonEmpty`, `TestToSummaryMarkdown_DEC014FullDocumentGolden` |
| DEC-050 rule 5 / LD4 | `TestToSummaryJSON_FailureSectionGolden`, `TestToSummaryJSON_DEC014ShapeGolden` |
| DEC-014 part 4 | `TestToSummary_EmptyEntriesEmitsProvenanceOnly`, `TestToSummary_SectionsRenderOnlyWhenNonEmpty` |
| LD6 | `TestSummaryCmd_HelpNamesTheFailureSection` |
| LD9 | `AF1`–`AF3` |

---

## Implementation Context

### Decisions that apply

- **DEC-050**: row 3 and rules 1–5, plus both amendments. This spec
  implements it and decides nothing new.
- **DEC-048**: no count changes what it counts.
- **DEC-014**: the envelope, key order by struct declaration, and part 4's
  `[]`.

### Constraints that apply

- `one-spec-per-pr`.
- `AGENTS.md` §12: a *no difference* is a measurement. Every pinned diff
  has exactly one literal reading.

### Prior related work

- **SPEC-086**: the same partition on `impact` and `wrapped`. Its
  `writeImpactGroups` is the model for `writeHighlightGroups`.
- **SPEC-094**: `story`. Its `OmitFailures` is the other caller of
  `SplitFailures` on every entry.

### Out of scope (for this spec specifically)

`story`, `impact` and `wrapped` (shipped); **`brag review`**, which also lists
failures under a neutral `## Entries` (framing's reasoning stands); SPEC-096;
`--type` negation; SPEC-090 through SPEC-093; STAGE-027; `brag lint`;
STAGE-020; `test-docs` in CI; `guidance/questions.yaml`; the v0.7.0 release
cut; brags.

---

## Corrections and open questions (design)

1. **Mechanic 4 was mostly done before this session.** The handoff said
   DEC-050's *Implemented by* column and its *"until SPEC-094 ships"*
   consequence predate the split. SPEC-094's design amendment had already
   re-pointed both. What remained were three sentences (Context,
   Validation, References), which this amendment covers.
2. **The corpus moved: 623 entries, not 621.** The two new rows are ship
   brags, neither a failure. The failure set is unchanged.
3. **`summary` has no `--since`.** The handoff offered *"a seeded store with
   an injected clock, or `--since`"*. The second does not exist, and the
   first needs no clock seam, because rows written at run time are always
   inside `--range week`.
4. **`SplitFailures`'s doc comment was already stale** after SPEC-094 (it
   said *"the digests call it on the with-impact subset"* while
   `story.OmitFailures` split every entry). LD8 corrects it while making it
   true for `summary`.
5. **The contract's `brag summary` section never listed its JSON keys.**
   LD9 adds the full list, not just the new one.
6. **Open, for the maintainer, not blocking:** whether DEC-014 wants its own
   `## Amendment` naming `failures_by_project` among `summary`'s keys. LD10
   says no, on choice 2's own wording. If the maintainer prefers the key
   list in DEC-014 too, it is a one-paragraph amendment at build, with no
   code impact.

---

## Notes for the Implementer

### Order of work

1. `git switch main && git pull --ff-only`, confirm this design merged,
   and branch `build/spec-095-summary-honesty`.
2. **Apply the literals.** Each `diff` block below is `git diff` output
   against `55063e9`. Save each to a file and `git apply` it, or apply them
   together. `git apply --check` accepted all 22 hunks against the design
   branch. If `main` has moved and a hunk no longer applies, apply it by
   hand from its context: every hunk's old side occurs exactly once in its
   file.
3. **Fail-first:** apply the three test files' diffs first and run
   `go test ./internal/export/ ./internal/cli/`. Expect exactly the 8
   failures in *Finding 1*. Then apply the rest.
4. Run `just inventory` and paste its output between the markers. The
   `docs/engineering-practices.md` diff below is what that should produce.
   **Do not hand-edit it.**
5. Re-run the matrix (16 probes) and the five gates, then run AC-1 to AC-5
   on a scratch store.
6. `just advance-cycle SPEC-095 build`, then **restore the inline enum
   comment it strips** from `cycle:`.

### Traps

- **`runSummary` reads `time.Now()`.** Do not write a Go test that seeds
  rows with fixed past dates and reads `--range`: it has a cliff. Write rows
  at run time, or test the renderer.
- **`markdownSection` in `learn_test.go` ends a section at the next `## `
  line.** `## Summary`'s body therefore includes `**By type**` and
  `**By project**`, which AC-3's `- failed: 1` check relies on.
- **`summary_test.go`'s diff is at `-U12` on purpose** (see *§12(b)*). Do not
  regenerate the literal at `-U3` when recording what you applied.
- **zsh:** write JSON to files, never pipe it through `echo`. Never name a
  variable `path`, and brace `${B}` before a `:`.
- **AC2** fails on any tree that is not a git checkout. Run `test-docs` in
  the worktree.
- Count with `/usr/bin/grep` or `git grep`, and keep out of
  `.claude/worktrees/`.

### §1. The renderer (LD1–LD4, LD7)

`internal/export/summary.go`: 7 hunks, base `55063e9`, file hash after `3b81e79b3fdf`.

```diff
diff --git a/internal/export/summary.go b/internal/export/summary.go
index ab03aa8..16c540d 100644
--- a/internal/export/summary.go
+++ b/internal/export/summary.go
@@ -27,6 +27,12 @@ type SummaryOptions struct {
 // DEC-014. Returns bytes with the trailing "\n" stripped (matches
 // ToJSON / ToMarkdown). On empty input, only the header + provenance
 // block is emitted; the Summary and Highlights sections are omitted.
+//
+// The recorded failures (aggregate.SplitFailures, over EVERY in-window
+// entry, because highlights list every entry) leave ## Highlights for
+// ## What didn't work, in the same grouped shape (DEC-050 row 3). Each of
+// the two sections is emitted only when it has an entry. The By type and
+// By project counts still count both.
 func ToSummaryMarkdown(entries []storage.Entry, opts SummaryOptions) ([]byte, error) {
 	var buf bytes.Buffer
 	fmt.Fprintln(&buf, "# Bragfile Summary")
@@ -51,17 +57,33 @@ func ToSummaryMarkdown(entries []storage.Entry, opts SummaryOptions) ([]byte, er
 	for _, pc := range aggregate.ByProject(entries) {
 		fmt.Fprintf(&buf, "- %s: %d\n", pc.Project, pc.Count)
 	}
-	fmt.Fprintln(&buf)
-	fmt.Fprintln(&buf, "## Highlights")
-	for _, group := range aggregate.GroupForHighlights(entries) {
+	worked, failed := aggregate.SplitFailures(entries)
+	if len(worked) > 0 {
 		fmt.Fprintln(&buf)
-		fmt.Fprintf(&buf, "### %s\n", group.Project)
+		fmt.Fprintln(&buf, "## Highlights")
+		writeHighlightGroups(&buf, worked)
+	}
+	if len(failed) > 0 {
 		fmt.Fprintln(&buf)
+		fmt.Fprintln(&buf, "## What didn't work")
+		writeHighlightGroups(&buf, failed)
+	}
+	return trimTrailingNewline(buf.Bytes()), nil
+}
+
+// writeHighlightGroups renders entries grouped by project as `### <project>`
+// blocks of `- <id>: <title>` lines. ## Highlights and ## What didn't work
+// both render through it, so a failure reads in the same shape as the
+// highlight it used to be.
+func writeHighlightGroups(buf *bytes.Buffer, entries []storage.Entry) {
+	for _, group := range aggregate.GroupForHighlights(entries) {
+		fmt.Fprintln(buf)
+		fmt.Fprintf(buf, "### %s\n", group.Project)
+		fmt.Fprintln(buf)
 		for _, ref := range group.Entries {
-			fmt.Fprintf(&buf, "- %d: %s\n", ref.ID, ref.Title)
+			fmt.Fprintf(buf, "- %d: %s\n", ref.ID, ref.Title)
 		}
 	}
-	return trimTrailingNewline(buf.Bytes()), nil
 }
 
 // summaryEnvelope is the on-the-wire shape for ToSummaryJSON. Field
@@ -74,6 +96,9 @@ type summaryEnvelope struct {
 	CountsByType    map[string]int    `json:"counts_by_type"`
 	CountsByProject map[string]int    `json:"counts_by_project"`
 	Highlights      []highlightGroup  `json:"highlights"`
+	// FailuresByProject is DEC-050 rule 5's key: the recorded failures,
+	// grouped exactly as Highlights is. Always present, [] when empty.
+	FailuresByProject []highlightGroup `json:"failures_by_project"`
 }
 
 type highlightGroup struct {
@@ -88,9 +113,12 @@ type highlightEntry struct {
 
 // ToSummaryJSON renders the JSON envelope per DEC-014: single object,
 // flat top-level keys (generated_at, scope, filters, counts_by_type,
-// counts_by_project, highlights), pretty-printed with 2-space indent.
-// Empty-state values per DEC-014 choice (4): counts maps render as
-// {} and highlights as [], never null.
+// counts_by_project, highlights, failures_by_project), pretty-printed
+// with 2-space indent. highlights carries the entries that are not
+// recorded failures and failures_by_project the ones that are (DEC-050);
+// both counts maps still count every in-window entry. Empty-state values
+// per DEC-014 choice (4): counts maps render as {} and both arrays as
+// [], never null.
 func ToSummaryJSON(entries []storage.Entry, opts SummaryOptions) ([]byte, error) {
 	env := summaryEnvelope{
 		GeneratedAt:     opts.Now.UTC().Format(time.RFC3339),
@@ -98,7 +126,6 @@ func ToSummaryJSON(entries []storage.Entry, opts SummaryOptions) ([]byte, error)
 		Filters:         opts.FiltersJSON,
 		CountsByType:    map[string]int{},
 		CountsByProject: map[string]int{},
-		Highlights:      []highlightGroup{},
 	}
 	if env.Filters == nil {
 		env.Filters = map[string]string{}
@@ -109,6 +136,17 @@ func ToSummaryJSON(entries []storage.Entry, opts SummaryOptions) ([]byte, error)
 	for _, pc := range aggregate.ByProject(entries) {
 		env.CountsByProject[pc.Project] = pc.Count
 	}
+	worked, failed := aggregate.SplitFailures(entries)
+	env.Highlights = highlightGroups(worked)
+	env.FailuresByProject = highlightGroups(failed)
+	return json.MarshalIndent(env, "", "  ")
+}
+
+// highlightGroups projects entries onto the {project, entries: [{id,
+// title}]} group shape both summary arrays share. Non-nil for empty
+// input, so either key renders as [] rather than null.
+func highlightGroups(entries []storage.Entry) []highlightGroup {
+	out := []highlightGroup{}
 	for _, group := range aggregate.GroupForHighlights(entries) {
 		hg := highlightGroup{
 			Project: group.Project,
@@ -119,7 +157,7 @@ func ToSummaryJSON(entries []storage.Entry, opts SummaryOptions) ([]byte, error)
 				ID: ref.ID, Title: ref.Title,
 			})
 		}
-		env.Highlights = append(env.Highlights, hg)
+		out = append(out, hg)
 	}
-	return json.MarshalIndent(env, "", "  ")
+	return out
 }
```

### §2. `SplitFailures`'s doc comment (LD8)

`internal/aggregate/aggregate.go`: 1 hunk, base `55063e9`, file hash after `70b2f84f8208`.

```diff
diff --git a/internal/aggregate/aggregate.go b/internal/aggregate/aggregate.go
index 9e2a6b7..fefba38 100644
--- a/internal/aggregate/aggregate.go
+++ b/internal/aggregate/aggregate.go
@@ -332,10 +332,13 @@ func IsFailure(e storage.Entry) bool {
 
 // SplitFailures partitions entries into those that are not failures and those
 // that are (IsFailure), preserving input order within each. Both results are
-// non-nil so JSON callers never see null. The digests call it on the
+// non-nil so JSON callers never see null. impact and wrapped call it on the
 // with-impact subset, never in place of WithImpact: WithImpact keeps meaning
 // "non-empty impact", and this decides only which section a row renders in
-// (DEC-050).
+// (DEC-050). story's OmitFailures and summary call it on every in-window
+// entry instead: summary's highlights list every entry, impact or not, so its
+// section lists a failure with no impact, which impact's and wrapped's do not
+// (SPEC-095).
 func SplitFailures(entries []storage.Entry) (others, failures []storage.Entry) {
 	others = make([]storage.Entry, 0, len(entries))
 	failures = make([]storage.Entry, 0)
```

### §3. The `--help` sentence (LD6)

`internal/cli/summary.go`: 1 hunk, base `55063e9`, file hash after `6bf6857adafd`.

```diff
diff --git a/internal/cli/summary.go b/internal/cli/summary.go
index 6b73ad5..301e68c 100644
--- a/internal/cli/summary.go
+++ b/internal/cli/summary.go
@@ -25,6 +25,8 @@ Output is markdown (default) or a single-object JSON envelope (--format json) pe
 
 --range is required: week = last 7 UTC days from time.Now(); month = last 30 UTC days. Rolling, NOT calendar. Filter flags --tag/--project/--type compose with the range.
 
+Work recorded with brag learn is listed under its own "What didn't work" heading instead of among the highlights, and that heading is left out when there is none. The by-type and by-project counts still include it.
+
 Examples:
   brag summary --range week                          # last 7 UTC days, markdown
   brag summary --range month --format json           # last 30 UTC days, JSON envelope
```

### §4. Export tests: 4 new, 2 rewrites (diffed at `-U12`)

`internal/export/summary_test.go`: 4 hunks, base `55063e9`, file hash after `c8568ecaf0ee`.

```diff
diff --git a/internal/export/summary_test.go b/internal/export/summary_test.go
index 30659e0..b4bc274 100644
--- a/internal/export/summary_test.go
+++ b/internal/export/summary_test.go
@@ -1,17 +1,21 @@
 package export
 
 import (
 	"bytes"
 	"encoding/json"
+	"fmt"
+	"reflect"
+	"sort"
+	"strconv"
 	"strings"
 	"testing"
 	"time"
 
 	"github.com/jysf/bragfile000/internal/storage"
 )
 
 // summaryFixture is the load-bearing fixture shared across the markdown
 // + JSON goldens (tests #5, #6) and the empty-state + filters-echo
 // tests (#7, #8). 5 entries spanning 3 projects + (no project), with
 // chrono-ordering chosen to exercise within-alpha chrono-ASC (1 → 4,
 // IDs and timestamps NOT monotonic together so ID tie-break is
@@ -193,46 +197,47 @@ func TestToSummaryJSON_DEC014ShapeGolden(t *testing.T) {
         }
       ]
     },
     {
       "project": "(no project)",
       "entries": [
         {
           "id": 3,
           "title": "unbound-mid"
         }
       ]
     }
-  ]
+  ],
+  "failures_by_project": []
 }`
 
 	got, err := ToSummaryJSON(summaryFixture, opts)
 	if err != nil {
 		t.Fatalf("ToSummaryJSON: %v", err)
 	}
 	if !bytes.Equal(got, []byte(want)) {
 		t.Fatalf("DEC-014 JSON golden mismatch\nwant:\n%s\n\ngot:\n%s", want, string(got))
 	}
 
 	// Verify struct-tag declaration order on top-level keys via a
 	// json.Decoder walk. DEC-014 rests on this key order.
 	dec := json.NewDecoder(bytes.NewReader(got))
 	tok, err := dec.Token()
 	if err != nil {
 		t.Fatalf("decoder.Token open: %v", err)
 	}
 	if d, ok := tok.(json.Delim); !ok || d != '{' {
 		t.Fatalf("expected opening {, got %v", tok)
 	}
-	wantKeys := []string{"generated_at", "scope", "filters", "counts_by_type", "counts_by_project", "highlights"}
+	wantKeys := []string{"generated_at", "scope", "filters", "counts_by_type", "counts_by_project", "highlights", "failures_by_project"}
 	for _, k := range wantKeys {
 		tok, err := dec.Token()
 		if err != nil {
 			t.Fatalf("decoder.Token key %q: %v", k, err)
 		}
 		gotKey, ok := tok.(string)
 		if !ok {
 			t.Fatalf("expected string key %q, got %T(%v)", k, tok, tok)
 		}
 		if gotKey != k {
 			t.Fatalf("expected key %q, got %q", k, gotKey)
 		}
@@ -326,25 +331,26 @@ Filters: (none)`
 		opts := SummaryOptions{
 			Scope:       "week",
 			Filters:     "(none)",
 			FiltersJSON: map[string]string{},
 			Now:         summaryFixedNow,
 		}
 		want := `{
   "generated_at": "2026-04-25T12:00:00Z",
   "scope": "week",
   "filters": {},
   "counts_by_type": {},
   "counts_by_project": {},
-  "highlights": []
+  "highlights": [],
+  "failures_by_project": []
 }`
 		got, err := ToSummaryJSON([]storage.Entry{}, opts)
 		if err != nil {
 			t.Fatalf("ToSummaryJSON: %v", err)
 		}
 		if !bytes.Equal(got, []byte(want)) {
 			t.Fatalf("empty-state JSON mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
 		}
 		var m map[string]any
 		if err := json.Unmarshal(got, &m); err != nil {
 			t.Fatalf("json.Unmarshal: %v", err)
 		}
@@ -453,12 +459,375 @@ func TestToSummaryMarkdown_FiltersLineFormat(t *testing.T) {
 		found := false
 		for _, ln := range strings.Split(string(got), "\n") {
 			if ln == "Filters: --project platform --tag auth" {
 				found = true
 				break
 			}
 		}
 		if !found {
 			t.Errorf("expected line %q in:\n%s", "Filters: --project platform --tag auth", string(got))
 		}
 	})
 }
+
+// summaryFailureFixture is summaryFixture plus three rows typed "failed" —
+// the literal DEC-049 persists, spelled out rather than taken from
+// aggregate.FailureType so these tests also fail if the constant drifts from
+// the stored value. id 6 is an alpha failure dated between alpha's two
+// highlights, so it must leave their group. id 7 is a failure with NO impact,
+// and it still lands in ## What didn't work: summary's highlights list every
+// entry, so its partition covers every entry (SPEC-095 LD1), unlike impact's
+// and wrapped's. delta holds only id 7, so it is a failures-only project and
+// must not appear under ## Highlights. id 8 is a (no project) failure, so the
+// section's (no project)-last rule is exercised too. 8 in window: 5
+// highlights, 3 failures.
+var summaryFailureFixture = append(append([]storage.Entry{}, summaryFixture...),
+	storage.Entry{ID: 6, Title: "pool-dead-end",
+		Project: "alpha", Type: "failed",
+		Impact:    "cost two days and produced nothing reusable",
+		CreatedAt: time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC),
+		UpdatedAt: time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)},
+	storage.Entry{ID: 7, Title: "retry-noimpact",
+		Project: "delta", Type: "failed",
+		Impact:    "", // no impact, and still listed (LD1)
+		CreatedAt: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC),
+		UpdatedAt: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)},
+	storage.Entry{ID: 8, Title: "unbound-dead-end",
+		Type:      "failed", // (no project) group
+		Impact:    "ruled out the vendor SDK",
+		CreatedAt: time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC),
+		UpdatedAt: time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC)},
+)
+
+var summaryFailureOpts = SummaryOptions{
+	Scope:       "week",
+	Filters:     "(none)",
+	FiltersJSON: map[string]string{},
+	Now:         summaryFixedNow,
+}
+
+// TestToSummaryMarkdown_FailureSectionGolden ▲ SPEC-095 LD2/LD3
+// (LOAD-BEARING). A recorded failure leaves ## Highlights and renders under
+// ## What didn't work, in the same grouped `- <id>: <title>` shape. By type
+// and By project are unchanged in meaning: they count all eight entries,
+// both sections, and `failed: 3` stays where it always was.
+func TestToSummaryMarkdown_FailureSectionGolden(t *testing.T) {
+	got, err := ToSummaryMarkdown(summaryFailureFixture, summaryFailureOpts)
+	if err != nil {
+		t.Fatalf("ToSummaryMarkdown: %v", err)
+	}
+	want := `# Bragfile Summary
+
+Generated: 2026-04-25T12:00:00Z
+Scope: week
+Filters: (none)
+
+## Summary
+
+**By type**
+- failed: 3
+- shipped: 3
+- fixed: 1
+- learned: 1
+
+**By project**
+- alpha: 3
+- beta: 1
+- delta: 1
+- gamma: 1
+- (no project): 2
+
+## Highlights
+
+### alpha
+
+- 1: alpha-old
+- 4: alpha-new
+
+### beta
+
+- 2: beta-mid
+
+### gamma
+
+- 5: gamma-only
+
+### (no project)
+
+- 3: unbound-mid
+
+## What didn't work
+
+### alpha
+
+- 6: pool-dead-end
+
+### delta
+
+- 7: retry-noimpact
+
+### (no project)
+
+- 8: unbound-dead-end`
+	if string(got) != want {
+		t.Errorf("markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
+	}
+}
+
+// TestToSummaryJSON_FailureSectionGolden ▲ SPEC-095 LD3/LD4 (LOAD-BEARING).
+// The failures leave highlights and arrive in failures_by_project, in the
+// same {project, entries: [{id, title}]} shape; both counts maps still count
+// every entry.
+func TestToSummaryJSON_FailureSectionGolden(t *testing.T) {
+	got, err := ToSummaryJSON(summaryFailureFixture, summaryFailureOpts)
+	if err != nil {
+		t.Fatalf("ToSummaryJSON: %v", err)
+	}
+	want := `{
+  "generated_at": "2026-04-25T12:00:00Z",
+  "scope": "week",
+  "filters": {},
+  "counts_by_type": {
+    "failed": 3,
+    "fixed": 1,
+    "learned": 1,
+    "shipped": 3
+  },
+  "counts_by_project": {
+    "(no project)": 2,
+    "alpha": 3,
+    "beta": 1,
+    "delta": 1,
+    "gamma": 1
+  },
+  "highlights": [
+    {
+      "project": "alpha",
+      "entries": [
+        {
+          "id": 1,
+          "title": "alpha-old"
+        },
+        {
+          "id": 4,
+          "title": "alpha-new"
+        }
+      ]
+    },
+    {
+      "project": "beta",
+      "entries": [
+        {
+          "id": 2,
+          "title": "beta-mid"
+        }
+      ]
+    },
+    {
+      "project": "gamma",
+      "entries": [
+        {
+          "id": 5,
+          "title": "gamma-only"
+        }
+      ]
+    },
+    {
+      "project": "(no project)",
+      "entries": [
+        {
+          "id": 3,
+          "title": "unbound-mid"
+        }
+      ]
+    }
+  ],
+  "failures_by_project": [
+    {
+      "project": "alpha",
+      "entries": [
+        {
+          "id": 6,
+          "title": "pool-dead-end"
+        }
+      ]
+    },
+    {
+      "project": "delta",
+      "entries": [
+        {
+          "id": 7,
+          "title": "retry-noimpact"
+        }
+      ]
+    },
+    {
+      "project": "(no project)",
+      "entries": [
+        {
+          "id": 8,
+          "title": "unbound-dead-end"
+        }
+      ]
+    }
+  ]
+}`
+	if string(got) != want {
+		t.Errorf("json golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
+	}
+}
+
+// TestToSummary_PartitionCoversEveryInWindowEntry ▲ SPEC-095 LD1/LD3 — DEC-050
+// rules 2 and 3 on summary. Every in-window entry, with or without an impact,
+// is listed exactly once across the two sections: a failure in ## What didn't
+// work, anything else in ## Highlights. The impact-less failure (id 7) is the
+// row impact and wrapped would not list; here it must be listed, or it is
+// counted in By type and shown nowhere. Both counts maps still sum to every
+// entry. Checked in both formats.
+func TestToSummary_PartitionCoversEveryInWindowEntry(t *testing.T) {
+	wantHighlights := "1,2,3,4,5"
+	wantFailures := "6,7,8"
+
+	md, err := ToSummaryMarkdown(summaryFailureFixture, summaryFailureOpts)
+	if err != nil {
+		t.Fatalf("ToSummaryMarkdown: %v", err)
+	}
+	// ids listed under a `## <heading>`, in ascending order, by line.
+	sectionIDs := func(heading string) string {
+		var ids []int
+		in := false
+		for _, ln := range strings.Split(string(md), "\n") {
+			if strings.HasPrefix(ln, "## ") {
+				in = ln == "## "+heading
+				continue
+			}
+			var id int
+			if in && strings.HasPrefix(ln, "- ") {
+				if _, err := fmt.Sscanf(ln, "- %d:", &id); err == nil {
+					ids = append(ids, id)
+				}
+			}
+		}
+		sort.Ints(ids)
+		var out []string
+		for _, id := range ids {
+			out = append(out, strconv.Itoa(id))
+		}
+		return strings.Join(out, ",")
+	}
+	if got := sectionIDs("Highlights"); got != wantHighlights {
+		t.Errorf("## Highlights ids = %q, want %q\n%s", got, wantHighlights, md)
+	}
+	if got := sectionIDs("What didn't work"); got != wantFailures {
+		t.Errorf("## What didn't work ids = %q, want %q\n%s", got, wantFailures, md)
+	}
+
+	raw, err := ToSummaryJSON(summaryFailureFixture, summaryFailureOpts)
+	if err != nil {
+		t.Fatalf("ToSummaryJSON: %v", err)
+	}
+	var env struct {
+		CountsByType      map[string]int   `json:"counts_by_type"`
+		CountsByProject   map[string]int   `json:"counts_by_project"`
+		Highlights        []highlightGroup `json:"highlights"`
+		FailuresByProject []highlightGroup `json:"failures_by_project"`
+	}
+	if err := json.Unmarshal(raw, &env); err != nil {
+		t.Fatalf("json.Unmarshal: %v", err)
+	}
+	groupIDs := func(groups []highlightGroup) string {
+		var ids []int
+		for _, g := range groups {
+			for _, e := range g.Entries {
+				ids = append(ids, int(e.ID))
+			}
+		}
+		sort.Ints(ids)
+		var out []string
+		for _, id := range ids {
+			out = append(out, strconv.Itoa(id))
+		}
+		return strings.Join(out, ",")
+	}
+	if got := groupIDs(env.Highlights); got != wantHighlights {
+		t.Errorf("highlights ids = %q, want %q", got, wantHighlights)
+	}
+	if got := groupIDs(env.FailuresByProject); got != wantFailures {
+		t.Errorf("failures_by_project ids = %q, want %q", got, wantFailures)
+	}
+	sum := func(m map[string]int) int {
+		n := 0
+		for _, v := range m {
+			n += v
+		}
+		return n
+	}
+	if got := sum(env.CountsByType); got != len(summaryFailureFixture) {
+		t.Errorf("counts_by_type sums to %d, want %d (both sections)", got, len(summaryFailureFixture))
+	}
+	if got := sum(env.CountsByProject); got != len(summaryFailureFixture) {
+		t.Errorf("counts_by_project sums to %d, want %d (both sections)", got, len(summaryFailureFixture))
+	}
+}
+
+// TestToSummary_SectionsRenderOnlyWhenNonEmpty ▲ SPEC-095 LD2 — DEC-050 rule 4
+// on summary, and the failures-only window. Each `##` body section renders
+// only when it has an entry, so a window holding only failures has no bare
+// `## Highlights` heading (SPEC-086 pinned the same for `## Impact`). JSON
+// always carries both keys: the failures-only window renders `"highlights":
+// []` (DEC-014 part 4), and the clean window `"failures_by_project": []`.
+func TestToSummary_SectionsRenderOnlyWhenNonEmpty(t *testing.T) {
+	headings := func(md []byte) []string {
+		var out []string
+		for _, ln := range strings.Split(string(md), "\n") {
+			if strings.HasPrefix(ln, "## ") {
+				out = append(out, ln)
+			}
+		}
+		return out
+	}
+	failuresOnly := []storage.Entry{summaryFailureFixture[5], summaryFailureFixture[6]}
+	cases := []struct {
+		name         string
+		entries      []storage.Entry
+		want         []string
+		wantHL       int
+		wantFailures int
+	}{
+		{"no failures", summaryFixture, []string{"## Summary", "## Highlights"}, 4, 0},
+		{"failures only", failuresOnly, []string{"## Summary", "## What didn't work"}, 0, 2},
+		{"both", summaryFailureFixture, []string{"## Summary", "## Highlights", "## What didn't work"}, 4, 3},
+	}
+	for _, c := range cases {
+		md, err := ToSummaryMarkdown(c.entries, summaryFailureOpts)
+		if err != nil {
+			t.Fatalf("%s: ToSummaryMarkdown: %v", c.name, err)
+		}
+		if got := headings(md); !reflect.DeepEqual(got, c.want) {
+			t.Errorf("%s: ## headings = %q, want %q\n%s", c.name, got, c.want, md)
+		}
+
+		raw, err := ToSummaryJSON(c.entries, summaryFailureOpts)
+		if err != nil {
+			t.Fatalf("%s: ToSummaryJSON: %v", c.name, err)
+		}
+		var env map[string]json.RawMessage
+		if err := json.Unmarshal(raw, &env); err != nil {
+			t.Fatalf("%s: json.Unmarshal: %v", c.name, err)
+		}
+		for key, wantLen := range map[string]int{"highlights": c.wantHL, "failures_by_project": c.wantFailures} {
+			v, ok := env[key]
+			if !ok {
+				t.Errorf("%s: JSON is missing %q:\n%s", c.name, key, raw)
+				continue
+			}
+			var groups []highlightGroup
+			if err := json.Unmarshal(v, &groups); err != nil || groups == nil {
+				t.Errorf("%s: %q = %s, want an array (never null)", c.name, key, v)
+				continue
+			}
+			if len(groups) != wantLen {
+				t.Errorf("%s: %q has %d groups, want %d:\n%s", c.name, key, len(groups), wantLen, raw)
+			}
+		}
+	}
+}
```

### §5. The end-to-end test and `runDigestCorpus`

`internal/cli/learn_test.go`: 2 hunks, base `55063e9`, file hash after `4b332bb26398`.

```diff
diff --git a/internal/cli/learn_test.go b/internal/cli/learn_test.go
index 0ffe2a5..eec0283 100644
--- a/internal/cli/learn_test.go
+++ b/internal/cli/learn_test.go
@@ -168,17 +168,17 @@ func TestLearnCmd_EmptyTitleIsUserError(t *testing.T) {
 }
 
 // runDigestCorpus runs one brag invocation against dbPath on a fresh root
-// carrying the two writers (add, learn), the two digests that section
-// failures (impact, wrapped), and story, which labels or omits them
-// (SPEC-094). A fresh root per call, so no flag value leaks from one
-// invocation into the next.
+// carrying the two writers (add, learn), the three digests that section
+// failures (impact, wrapped, summary — SPEC-095), and story, which labels or
+// omits them (SPEC-094). A fresh root per call, so no flag value leaks from
+// one invocation into the next.
 func runDigestCorpus(t *testing.T, dbPath string, args ...string) string {
 	t.Helper()
 	t.Setenv("BRAGFILE_DB", "")
 	addStderrIsTTY = func() bool { return false }
 	t.Cleanup(func() { addStderrIsTTY = defaultStderrIsTTY })
 	root := NewRootCmd("test")
-	root.AddCommand(NewAddCmd(), NewLearnCmd(), NewImpactCmd(), NewWrappedCmd(), NewStoryCmd())
+	root.AddCommand(NewAddCmd(), NewLearnCmd(), NewImpactCmd(), NewWrappedCmd(), NewSummaryCmd(), NewStoryCmd())
 	var outBuf, errBuf bytes.Buffer
 	root.SetOut(&outBuf)
 	root.SetErr(&errBuf)
@@ -357,3 +357,68 @@ func TestLearnCmd_StoryLabelsOrOmitsWhatItWrote(t *testing.T) {
 		t.Errorf("exec omitted_failure_count = %v, want 1", env.OmittedFailureCount)
 	}
 }
+
+// TestLearnCmd_SummarySectionsWhatItWrote ▲ SPEC-095 LD1/LD2/LD5 — the
+// writer-to-reader check on brag summary, through a real store and with no
+// constant in sight. The brag learn entry carries NO impact on purpose: it is
+// the row impact and wrapped leave out, and summary must still list it, under
+// "What didn't work" and not under "Highlights" (LD1). Both rows are written
+// now and the window is --range week, so the test has no window cliff (LD5):
+// it cannot pass on a window that holds no failure. Every negative is paired
+// with a positive on the same section.
+func TestLearnCmd_SummarySectionsWhatItWrote(t *testing.T) {
+	dbPath := filepath.Join(t.TempDir(), "test.db")
+	winID := strings.TrimSpace(runDigestCorpus(t, dbPath, "add", "-t", "shipped the cache", "-p", "alpha", "-k", "shipped"))
+	failID := strings.TrimSpace(runDigestCorpus(t, dbPath, "learn", "-t", "tried a worker pool", "-p", "alpha"))
+
+	md := runDigestCorpus(t, dbPath, "summary", "--range", "week")
+	highlights := markdownSection(md, "Highlights")
+	failed := markdownSection(md, "What didn't work")
+	if !strings.Contains(highlights, "- "+winID+": shipped the cache") {
+		t.Errorf("## Highlights is missing the brag add entry %s:\n%s", winID, md)
+	}
+	if strings.Contains(highlights, "- "+failID+":") {
+		t.Errorf("## Highlights still carries the brag learn entry %s:\n%s", failID, md)
+	}
+	if !strings.Contains(failed, "- "+failID+": tried a worker pool") {
+		t.Errorf("## What didn't work is missing the brag learn entry %s:\n%s", failID, md)
+	}
+	if strings.Contains(failed, "- "+winID+":") {
+		t.Errorf("## What didn't work carries the brag add entry %s:\n%s", winID, md)
+	}
+	if !strings.Contains(markdownSection(md, "Summary"), "- failed: 1\n") {
+		t.Errorf("By type should still count the brag learn entry:\n%s", md)
+	}
+
+	type group struct {
+		Entries []struct {
+			ID int64 `json:"id"`
+		} `json:"entries"`
+	}
+	var env struct {
+		CountsByType      map[string]int `json:"counts_by_type"`
+		Highlights        []group        `json:"highlights"`
+		FailuresByProject []group        `json:"failures_by_project"`
+	}
+	if err := json.Unmarshal([]byte(runDigestCorpus(t, dbPath, "summary", "--range", "week", "--format", "json")), &env); err != nil {
+		t.Fatalf("json unmarshal: %v", err)
+	}
+	ids := func(groups []group) string {
+		var out []string
+		for _, g := range groups {
+			for _, e := range g.Entries {
+				out = append(out, strconv.FormatInt(e.ID, 10))
+			}
+		}
+		return strings.Join(out, ",")
+	}
+	if got := ids(env.Highlights); got != winID {
+		t.Errorf("highlights ids = %q, want %q", got, winID)
+	}
+	if got := ids(env.FailuresByProject); got != failID {
+		t.Errorf("failures_by_project ids = %q, want %q", got, failID)
+	}
+	if env.CountsByType["failed"] != 1 || env.CountsByType["shipped"] != 1 {
+		t.Errorf("counts_by_type = %v, want failed:1 and shipped:1", env.CountsByType)
+	}
+}
```

### §6. The help test

`internal/cli/summary_test.go`: 1 hunk, base `55063e9`, file hash after `62e3bf34f0ca`.

```diff
diff --git a/internal/cli/summary_test.go b/internal/cli/summary_test.go
index 1c1ff09..572422a 100644
--- a/internal/cli/summary_test.go
+++ b/internal/cli/summary_test.go
@@ -385,3 +385,19 @@ func lineMatches(s string, re *regexp.Regexp) bool {
 	}
 	return false
 }
+
+// TestSummaryCmd_HelpNamesTheFailureSection ▲ SPEC-095 LD6: summary's --help
+// says where a brag learn entry goes, since it is no longer among the
+// highlights, and that the counts still include it. "What didn't work" is a
+// token no cobra-generated line can produce.
+func TestSummaryCmd_HelpNamesTheFailureSection(t *testing.T) {
+	root, outBuf, _ := newSummaryTestRoot(t)
+	root.SetArgs([]string{"summary", "--help"})
+	if err := root.Execute(); err != nil {
+		t.Fatalf("unexpected error: %v", err)
+	}
+	want := `Work recorded with brag learn is listed under its own "What didn't work" heading instead of among the highlights, and that heading is left out when there is none. The by-type and by-project counts still include it.`
+	if !bytes.Contains(outBuf.Bytes(), []byte(want)) {
+		t.Errorf("expected the failure-section sentence in help:\n%s", outBuf.String())
+	}
+}
```

### §7. The contract's `brag summary` section (LD9)

`docs/api-contract.md`: 2 hunks, base `55063e9`, file hash after `3da3b8d8a35f`.

```diff
diff --git a/docs/api-contract.md b/docs/api-contract.md
index 6e17cb7..5b1c2c7 100644
--- a/docs/api-contract.md
+++ b/docs/api-contract.md
@@ -364,7 +364,18 @@ carrying:
   list).
 - **Highlights:** entry titles + IDs grouped by project,
   chronological-ASC within group; descriptions are intentionally
-  elided for the "skim before pasting" goal.
+  elided for the "skim before pasting" goal. Recorded failures are
+  not highlights (next item).
+- **`## What didn't work`** (markdown): the in-window entries whose
+  `type` is the reserved `failed` that `brag learn` writes, pulled out
+  of `## Highlights` and rendered after it in the same grouped shape.
+  Unlike `brag impact` and `brag wrapped`, it lists a failure with no
+  impact statement too, because `summary`'s highlights list every
+  entry. Each of the two sections appears only when it has an entry:
+  a window with no recorded failure has no `## What didn't work`
+  heading, and a window holding only failures has no `## Highlights`
+  heading. `By type` and `By project` count both sections. Locked by
+  [DEC-050](../decisions/DEC-050-a-failure-is-never-rendered-as-a-win.md).
 
 Flags:
 - `--range week|month` REQUIRED. `week` = last 7 UTC days from
@@ -374,6 +385,12 @@ Flags:
   single-object envelope (NOT an array — diverges from DEC-011's
   list shape because aggregations carry metadata). Shape locked by
   [DEC-014](../decisions/DEC-014-rule-based-output-shape.md).
+  Top-level keys: `generated_at`, `scope`, `filters`,
+  `counts_by_type`, `counts_by_project` (both count every in-window
+  entry, failures included), `highlights` (an array of
+  `{project, entries:[{id, title}]}` groups, failures excluded), and
+  `failures_by_project` (the same shape, holding only the failures).
+  Both arrays are always present, as `[]` when empty.
 - `--tag <token>`, `--project <name>`, `--type <name>` reuse `brag
   list`'s `ListFilter` semantics. No `--since`/`--limit`/`--out` on
   summary in MVP.
```

### §8. `CHANGELOG.md` (LD9)

`CHANGELOG.md`: 1 hunk, base `55063e9`, file hash after `33c7ebd14725`.

```diff
diff --git a/CHANGELOG.md b/CHANGELOG.md
index b7f0717..7ed8f79 100644
--- a/CHANGELOG.md
+++ b/CHANGELOG.md
@@ -46,6 +46,20 @@ and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0
   still sums to `entries_with_impact`. A consumer that read every
   with-impact entry from `impact_by_project` or `impact_moments` now reads
   both keys.
+- **`brag summary` lists work that did not work under its own
+  `## What didn't work` heading, not among `## Highlights`**
+  ([DEC-050](decisions/DEC-050-a-failure-is-never-rendered-as-a-win.md)).
+  Its highlights carry no type, so an entry recorded with `brag learn` read
+  as a highlight like any other. It now renders in the new section, after
+  the highlights, in the same shape. Because `summary` lists every entry,
+  the section lists a failure with no impact too, which `impact` and
+  `wrapped` leave out. `By type` and `By project` still count it.
+- **Breaking: `brag summary --format json` no longer lists failures in
+  `highlights`.** The envelope gains `failures_by_project`, after
+  `highlights`, in the same `{project, entries:[{id, title}]}` shape. It is
+  always present and `[]` when empty. `counts_by_type` and
+  `counts_by_project` are unchanged. A script that read every entry from
+  `highlights` now reads both keys.
 - **`brag story` no longer marks work that did not work as a win**
   ([DEC-054](decisions/DEC-054-story-candor-decides-what-a-failure-renders-as.md)).
   An entry recorded with `brag learn` rendered as `- ★ <id>:` on every
```

### §9. `AGENTS.md`, the `learn` glossary entry (LD9)

`AGENTS.md`: 1 hunk, base `55063e9`, file hash after `5fb3a6d60c68`.

```diff
diff --git a/AGENTS.md b/AGENTS.md
index 3b6b5a1..f23f93f 100644
--- a/AGENTS.md
+++ b/AGENTS.md
@@ -291,7 +291,7 @@ DECs are stable; specs come and go. DECs don't reciprocally list specs.
 - **Store** — the `*storage.Store` Go type that owns the `*sql.DB` and all typed methods. The only package that imports a SQL driver.
 - **migration** — a single `NNNN_*.sql` file under `internal/storage/migrations/`, embedded into the binary, applied automatically in lexical order on `storage.Open`.
 - **export** — a one-shot dump of entries, either as a Markdown report (stdout or `--out file.md`) or as a portable SQLite file copy (via `VACUUM INTO`).
-- **learn** — `brag learn`: the capture verb for work that did **not** work, and the only place `entries.type` is pinned rather than free-form. Writes the reserved type value `failed` (DEC-049) from flag mode or editor mode; no `--type` flag, no `-k`, no `--json`, and — deliberately — no milestone nudge on stderr, because every milestone line is a congratulation. `entries.type` stays free-form everywhere else: the verb pins one value, it does not validate the field. Retrieval is the existing `brag list --type failed`; the DEC-044 memory line renders `[<project>/failed]`, so a failure is labelled as one on the agent-facing read surface for free. `brag impact` and `brag wrapped` list a failure that carries an impact under their own `## What didn't work` section, never among the wins (DEC-050, SPEC-086). `brag story` labels one `✗ <id> (failed)` on a candid audience (`me`, `manager`) and leaves it out on a promotional one (`skip`, `exec`), with an `Omitted:` count and a directive clause telling the LLM to say so (DEC-054, SPEC-094). The constant is `aggregate.FailureType`, read through `aggregate.IsFailure`. PROJ-008 STAGE-023 (SPEC-085).
+- **learn** — `brag learn`: the capture verb for work that did **not** work, and the only place `entries.type` is pinned rather than free-form. Writes the reserved type value `failed` (DEC-049) from flag mode or editor mode; no `--type` flag, no `-k`, no `--json`, and — deliberately — no milestone nudge on stderr, because every milestone line is a congratulation. `entries.type` stays free-form everywhere else: the verb pins one value, it does not validate the field. Retrieval is the existing `brag list --type failed`; the DEC-044 memory line renders `[<project>/failed]`, so a failure is labelled as one on the agent-facing read surface for free. `brag impact` and `brag wrapped` list a failure that carries an impact under their own `## What didn't work` section, never among the wins (DEC-050, SPEC-086). `brag summary` lists every failure in its window there too, with or without an impact, instead of among its highlights (SPEC-095). `brag story` labels one `✗ <id> (failed)` on a candid audience (`me`, `manager`) and leaves it out on a promotional one (`skip`, `exec`), with an `Omitted:` count and a directive clause telling the LLM to say so (DEC-054, SPEC-094). The constant is `aggregate.FailureType`, read through `aggregate.IsFailure`. PROJ-008 STAGE-023 (SPEC-085).
 - **review** — `brag review --week | --month`: prints recent entries grouped by project followed by three hard-coded reflection questions ("What pattern do you see in this period?", "What did you underestimate?", "What's missing here that should be?"). Markdown elides per-entry descriptions for compactness; JSON includes the full DEC-011 entry shape. Designed to be pasted into an external AI session for guided self-reflection. STAGE-004 (SPEC-019).
 - **summary** — a rule-based (non-LLM) aggregation of entries grouped by project/type over a rolling 7- or 30-day time window (`brag summary --range week|month`). STAGE-004.
 - **stats** — `brag stats`: six lifetime aggregations (total entries, entries/week rolling average, current streak, longest streak, top-5 most-common tags, top-5 most-common projects, corpus span). STAGE-004 (SPEC-020).
```

### §10. `scripts/test-docs.sh`, Group `AF` (LD9)

`scripts/test-docs.sh`: 1 hunk, base `55063e9`, file hash after `6385bfe6bde1`.

```diff
diff --git a/scripts/test-docs.sh b/scripts/test-docs.sh
index 31f8f66..c0fea14 100755
--- a/scripts/test-docs.sh
+++ b/scripts/test-docs.sh
@@ -2308,6 +2308,26 @@ assert_section_names "AE3" "$(grep -F -- '- **learn** —' AGENTS.md)" \
 assert_section_names "AE4" "$(ad_section CHANGELOG.md '## [Unreleased]' '## [')" \
     "CHANGELOG.md, the [Unreleased] section" '`omitted_failure_count`' 'DEC-054'
 
+# ===== Group AF — summary sections a failure (SPEC-095 / DEC-050 row 3) =====
+#
+# SPEC-095 changes what `brag summary` PRINTS: a recorded failure leaves
+# `## Highlights` for `## What didn't work`, and the JSON envelope gains
+# `failures_by_project`. The Go suite pins the output; these ids pin the docs
+# that describe it, scoped with Group AD's helpers so the `impact` and
+# `wrapped` sections, which already name both needles, cannot satisfy them.
+
+# AF1 — the contract documents the section AND the key on `summary`.
+assert_section_names "AF1" "$(ad_section docs/api-contract.md '### `brag summary' '### ')" \
+    "docs/api-contract.md, the brag summary section" "$ad_heading" '`failures_by_project`'
+
+# AF2 — the agent-facing glossary's learn entry names the fourth reader.
+assert_section_names "AF2" "$(grep -F -- '- **learn** —' AGENTS.md)" \
+    "AGENTS.md, the learn glossary entry" '`brag summary` lists every failure'
+
+# AF3 — the unreleased changelog names the breaking change to `highlights`.
+assert_section_names "AF3" "$(ad_section CHANGELOG.md '## [Unreleased]' '## [')" \
+    "CHANGELOG.md, the [Unreleased] section" '`brag summary --format json` no longer lists failures'
+
 # ===== finalise =====
 
 if [ "$FAIL_COUNT" -gt 0 ]; then
```

### §11. `docs/engineering-practices.md`: regenerate with `just inventory`, never hand-edit

`docs/engineering-practices.md`: 1 hunk, base `55063e9`, file hash after `a492ae543dfd`.

```diff
diff --git a/docs/engineering-practices.md b/docs/engineering-practices.md
index 9c183c4..ee60ba4 100644
--- a/docs/engineering-practices.md
+++ b/docs/engineering-practices.md
@@ -39,8 +39,8 @@ The section that says the most about how this repository is run is
 | …of those, also carrying a build-phase reflection | 81 | `### Build-phase reflection` in those files |
 | Go source files | 70 | `internal/`, `cmd/` |
 | Go test files | 81 | `internal/`, `cmd/` |
-| Go test functions | 856 | `func Test*` in `*_test.go` |
-| Documentation assertions (distinct ids) | 209 | `scripts/test-docs.sh`, run by `just test-docs` |
+| Go test functions | 862 | `func Test*` in `*_test.go` |
+| Documentation assertions (distinct ids) | 212 | `scripts/test-docs.sh`, run by `just test-docs` |
 | …of those, replacing a manual release-checklist item | 6 | the `W`-series in `scripts/test-docs.sh` |
 | Questions tracked in guidance/questions.yaml | 21 | `guidance/questions.yaml` |
 | …of those, still open | 8 | `status: open` in the same file |
```

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `build/spec-095-summary-honesty`
- **PR (if applicable):** none — the orchestrator opens it.
- **All acceptance criteria met?** yes (AC-1 through AC-10), all reproduced
  on a scratch store written at run time (LD5), never the live corpus. The
  live-corpus numbers were independently re-derived on a read-only
  `sqlite3 .backup` copy (hash `c93395b6e9bc2f…`, matching design's stated
  hash exactly, unchanged before and after): 623 entries, 4 failures
  (420/433/465/473), 0 impact-less, `--range month` holds 233 entries,
  `--range week` holds 0 failures — all identical to design's numbers.
- **New decisions emitted:** none. DEC-050's `## Amendment (2026-09-22,
  SPEC-095 design)` was written at design; this cycle only implements it.
- **Deviations from spec:** none in production code, docs, or tests — all 11
  literal diffs applied via `git apply` with a clean `--check` both before
  and after, and every one is byte-identical to the spec's block (the
  `internal/export/summary_test.go` block only matches at `-U12` context,
  per the spec's own trap; `git diff -U3`'s default framing differs but the
  content is identical). Two things worth recording, neither a spec defect:
  1. **M-9's reproduced hash** (`c840a90180f2`) differs from the spec's
     stated `e9404d9921ee`. The matrix's *Diff* column gives M-9's edit only
     in prose ("the same line wrapped in `if tc.Type != "failed" { … }`"),
     not an exact literal, unlike every other row. My one-line wrap differs
     byte-for-byte from design's (unstated) formatting. The three tests the
     spec says M-9 fires — `TestToSummaryJSON_FailureSectionGolden`,
     `TestToSummary_PartitionCoversEveryInWindowEntry`, and the e2e test —
     fired, and only those three, which is what the probe exists to show.
  2. **The design-time inventory table's "main" column understates the
     Amendment-section row.** It states `5` for `main` (`55063e9`), but the
     file actually committed at that commit already reads `6`
     (`decisions/DEC-025/028/029/030/050/054`, checked with `git grep -l
     '^## Amendment' 55063e9`). This predates this build — the row was
     already `6` in `docs/engineering-practices.md` at `HEAD` before
     `just inventory` ran — so the regenerated block still reads `6` and the
     row is unchanged by this spec's work, consistent with "every other row
     is unchanged." The design commit's own re-derivation table just cites a
     stale number for that one cell.
- **Follow-up work identified:** none beyond what the spec already routes
  (SPEC-096, `--type` negation, the v0.7.0 release cut).

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing in the spec's own mechanics. The one friction was mechanical,
   same shape as SPEC-094's build: the mutation matrix's *Diff* column gives
   exact edit text for every probe except M-9, whose wrap is described only
   in prose. The *behavior* (which tests fire) reproduces exactly, which is
   what M-9 exists to show; only the incidental final hash differs.

2. **Was there a constraint or decision that should have been listed but
   wasn't?**
   — No. `one-spec-per-pr`, `no-sql-in-cli-layer`,
   `stdout-is-for-data-stderr-is-for-humans`, and `test-before-implementation`
   all applied cleanly with no friction.

3. **If you did this task again, what would you do differently?**
   — Diff the regenerated `docs/engineering-practices.md` against the spec's
   embedded literal by *content* (splice the block, `diff` the result) rather
   than trying `git apply --check` on the embedded patch directly — the
   patch's context no longer matches once the file has already moved past
   the pre-inventory state, which is expected and not a failure to
   investigate.

---

## Verification

*Filled in at the end of the **verify** cycle. The orchestrator had already
confirmed that the 22 hunks rebuild the build tree byte for byte from a clean
`main`, and that all five gates pass. Those are re-checked cheaply here and
hold. This record covers what a passing build could not see: M-9's record,
the real output a user reads, mutants nobody imagined, and whether anything
outside `summary` moved.*

- **Branch:** `verify/spec-095-summary-honesty`, off `main` at `6590772`
  (PR #232, the build, merged as a squash; no stacking).
- **Verdict:** ⚠ **PUNCH LIST: two record corrections and one test
  strengthened (V-F3, second pass), all fixed in this cycle. No functional
  defect, and no line of production Go or harness changed.**
  **21 mutants** stand against it. The matrix has 16: M-9 was re-run here
  from its corrected record, and the other 15 are as build reproduced them
  and were not re-run. The other 5 are new in this cycle. A named test kills
  every one of them except M-D2, which design recorded as a probe too weak
  to fire (its Finding 3). The real output is
  correct on every window and format tried, including a seeded impact-less
  failure.

### The attack list

| # | Attack | Result |
|---|---|---|
| A | M-9's record | **V-F1: under-specified. FIXED.** It is now one literal edit that reproduces `e9404d9921ee`. Codification ruling below |
| B | Real output: `--range week` and `month`, `--type failed`, a clean window, both formats | **HOLDS.** Only the four failures move, and both count maps are identical to the pre-build binary |
| C | An impact-less failure, seeded | **HOLDS.** Listed under `## What didn't work` by `summary`, and counted but not listed by `impact`, as designed |
| D | 5 mutants the matrix did not try | **HOLDS.** All 5 killed; 3 predictions exact, 2 imprecise (below) |
| E | Nothing else changed | **HOLDS.** 28 invocations across 10 other surfaces, byte-identical across the two binaries, with every side shape-checked |
| F | Who reads the changed shape? | **HOLDS.** No consumer outside the contract, which is updated |
| G | The build reflection | **HONEST.** Both items reproduce. **V-F2**: the stale `5` is corrected in the design table |
| H | AC-1 to AC-7, AC-10, re-run rather than trusted | **HOLDS** |
| I | STAGE-023 candidate 2's named test (SPEC-094's `M-D1` at `c3f707a`) | **Reproduces after its target moved.** Reported for ship |

### V-F1: M-9 pinned. FIXED, and the codification ruling

The row's old side was unique; its new side, *"the same line wrapped in
`if tc.Type != "failed" { … }`"*, was not. The `…` leaves the line breaks
and the indentation open. Five readings were hashed against the baseline
`3b81e79b3fdf` (the file at `6590772` on `main`) before any probe ran:

| Reading | Hash |
|---|---|
| gofmt, three lines: `if … {` / body one tab deeper / `}` | **`e9404d9921ee`**, design's |
| one line, `if … { body }` | `c840a90180f2`, build's |
| three lines, body at the `if`'s own indent | `d59f0833ca21` |
| one line, no inner spaces | `f425f45d637d` |
| one line, `body; }` | `111efe2bc390` |

The row now states design's reading in full and names its base. Run through
the verify helper, it reproduces **`3b81e79b3fdf` → `e9404d9921ee`** and fires
exactly `TestToSummaryJSON_FailureSectionGolden`,
`TestToSummary_PartitionCoversEveryInWindowEntry` and
`TestLearnCmd_SummarySectionsWhatItWrote`. The file hash returned to its
baseline afterwards. The printed diff:

```diff
@@ -134 +134,3 @@
-		env.CountsByType[tc.Type] = tc.Count
+		if tc.Type != "failed" {
+			env.CountsByType[tc.Type] = tc.Count
+		}
```

The matrix preamble's *"each row has one literal reading"* now carries a
note that it held for every old side but not for M-9's new side.

**Ruling: M-9 does not advance candidate 1, and more rule text is not the
answer.**

1. **It is a uniqueness failure, not a fidelity failure.** SPEC-094's ship
   re-scoped STAGE-023 so that each failed property has one home. *Unique*
   (item 5, codified) means the stated edit has one literal reading.
   *Faithful* (candidate 1) means that reading is the edit that ran. M-9's
   sketch has at least five readings, and design's is one of them. The row
   visibly showed its own gap with a `…`, and design's hash reproduced as
   soon as the reading was named. That is `M-D1`'s shape (*faithful but
   ambiguous*), not `M-A0`'s (one clear edit that was not the one that ran).
   SPEC-094's ship already weighed and rejected the counter-argument that
   the preamble's false claim makes the row a composition failure. The same
   reasoning applies here: had the row been captured from the helper, it
   would also have been unique.
2. **The codified text already covers it, and its guard does not.** Item
   5's first sentence says *"applying it to the target admits exactly one
   result"*, and a sketch fails that. Its mechanical test and the guard it
   prescribes check only the **old** side: *"the replaced text occurs exactly
   once"*. Design's helper applied that guard, and it worked: none of the 17
   rows has an ambiguous old side, where SPEC-094 had one. The hole is where
   no guard looks.
3. **So the fix is in the tool, not the rule.** Across three specs, each
   non-reproducing row failed a different property (`M-A0` fidelity, `M-D1`
   old-side uniqueness, `M-9` new-side uniqueness). Each prose clause
   prevented its own case, and the next case showed up just outside it. A
   fourth sentence would follow the same pattern. What closed the only hole
   that stayed closed was a guard, and one mechanism covers all three cases:
   the helper **prints the diff it applied, and the matrix cell is that
   diff**. That is candidate 1's remedy of capture. It is also item 5's
   uniqueness test (a printed diff has one reading), and it cannot contain a
   `…`. Every row in this Verification is that printed output. This
   cycle's helper also refuses a new side containing an ellipsis, as the
   cheap half.
4. **The structural reason the guard keeps missing:** the probe helper is
   written again in each session's scratchpad (design's `probe.py`,
   SPEC-094 verify's, this one), so it covers what that session happened to
   think of, and no version of it persists. **Recommendation to ship:** do
   not codify new rule text. Record M-9 on STAGE-023's page under item 5 as a
   **guard-scope** case, not under candidate 1. Route **"commit the probe
   helper to `scripts/`, emitting the matrix row from its own diff"** as a
   STAGE-023 backlog item after v0.7.0, owned by whoever frames the next
   harness spec. **Ship codifies**, and may weigh this differently.

### V-F2: the design inventory table's `## Amendment` row. CORRECTED

Build's reflection is right. `git grep -l '^## Amendment' 55063e9 --
decisions` lists **6** records (DEC-025, 028, 029, 030, 050 and 054), and
`docs/engineering-practices.md` read `6` at `55063e9`, at `dfee276` and on
this branch. The design table typed `5` in all three columns, under a
heading that says *"regenerated and diffed, not predicted"*. All three cells
are now struck through and corrected to **6**. It did no harm, because `X3`
diffs the regenerated block and never read the table. **Not for promotion.**
It is one more instance of the rule CLAUDE.md and §9(c) already state
(numbers in docs are derived, not typed), and an instance of an
already-codified rule does not advance anything.

### Real output, read as a user would

Everything ran against a frozen copy (`sqlite3 ~/.bragfile/db.sqlite
".backup …"`), SHA-256 **`c93395b6e9bc`**: design's hash, and unchanged at the
end. There were two binaries: `brag-pre`, built from `git archive dfee276`,
and `brag-cur`, built from `6590772`. Every output started with its expected
header, stderr was empty, and every JSON file parsed. Diffs drop the
`Generated:` and `generated_at` lines.

| Window | Markdown, pre → cur | JSON |
|---|---|---|
| `--range month` | 4 lines removed from `## Highlights` (433, 465, 473 under `bragfile`; 420 under `contextcore-pilot-harness`), and a 12-line `## What didn't work` appended. **Nothing else moved** | keys gain `failures_by_project` last. `counts_by_type` and `counts_by_project` are `==` pre. `highlights` `==` pre with the four removed and empty groups dropped. Everything else `==` |
| `--range month --type failed` | the one line `## Highlights` → `## What didn't work`; no bare `## Highlights` | `"highlights": []` (raw text), 2 failure groups, `counts_by_type.failed` 4 |
| `--range week` (0 failures: `/usr/bin/grep -c -E '^- (420\|433\|465\|473): '` → 0) | **identical**, 69 lines; headings `## Summary`, `## Highlights` | `"failures_by_project": []` (raw text); everything else `==` pre |
| `--range week --type failed` (empty window) | identical, provenance only | identical plus `[]` |

`By type` still prints `- failed: 4` on the month window. Both count maps span
both sections.

**Seeded impact-less failure.** In a second copy (`seed.sqlite`, never the
live file), `brag-cur learn -t "VERIFYSEED tried sharding the FTS index" -p
bragfile` wrote id **646**, `type failed`, impact length **0**.

- `summary --range week`: 646 appears only under `## What didn't work`
  (an awk section walk finds it in one section). The diff against the
  unseeded week is exactly `- failed: 1`, `bragfile: 4 → 5`, and the new
  section. JSON: 646 is absent from `highlights`, `failures_by_project` is
  `[{bragfile:[646]}]`, and `counts_by_type.failed` is 1.
- `summary --range month`: 646 joins `bragfile`'s failure group
  `[433,465,473,646]`, and `failed` is 5.
- `impact --since 2026-09-01` on the same seed: `Entries: 232/234 with
  impact`, so 646 is counted in the window, and it appears in no section or
  JSON key. That is the asymmetry mechanic 1 states, observed rather than
  inferred.

**Window cliff.** This ran on 2026-09-23, before 420 leaves `--range month` at
2026-10-06T00:44:44Z, so the month rows above are real. Every NOT-contains
here is paired with a positive on the same section.

### Novel mutants: 5, each predicted before it ran

Predictions were written to a file before the first run (sha256
`c4bccfdd5aba…`). The helper refuses an old side that does not occur exactly
once and a new side that contains an ellipsis. It refuses to run the gates
until the hash has moved, runs `go test -count=1 ./...` and `test-docs`,
restores from a scratchpad backup with `cp`, and confirms the hash is back.
Base `3b81e79b3fdf` as of `6590772`. Each *Edit* is the diff the helper
printed. All five restored to `3b81e79b3fdf`, and `test-docs` stayed green
under all five.

| # | Edit (printed) | Prediction | Fired | Hash |
|---|---|---|---|---|
| **V-1** | `:101` `` `json:"failures_by_project"` `` → `` `json:"failures_by_project,omitempty"` `` (the key omitted when empty) | `DEC014ShapeGolden`, `EmptyEntries…/json`; `SectionsRenderOnly…` only if it reads raw key presence | `DEC014ShapeGolden`, `EmptyEntriesEmitsProvenanceOnly` (+`/json`), `SectionsRenderOnlyWhenNonEmpty`. **Matched**, conditional included | →`b048e901cf88` |
| **V-2** | `:139` (the JSON path's) `worked, failed := aggregate.SplitFailures(entries)` → `…SplitFailures(aggregate.WithImpact(entries))`, anchored on the next line `env.Highlights` (M-4's shortcut on the JSON path) | JSON `FailureSectionGolden`, `PartitionCovers…`; `DEC014ShapeGolden` iff its fixture has an impact-less entry | those three **plus** the e2e, `SectionsRenderOnlyWhenNonEmpty` and the pre-existing `TestSummaryCmd_FormatJSON_RangeWeekAndFiltersCompose`, **6 in all. Under-predicted**: impact-less wins are everywhere in the fixtures | →`37450a4810ed` |
| **V-3** | `:64` `writeHighlightGroups(&buf, worked)` → `writeHighlightGroups(&buf, entries)` (a failure in both sections) | md `FailureSectionGolden`, `PartitionCovers…`, the e2e | exactly those three. **Matched** | →`1e7d19730731` |
| **V-4** | `:53` `fmt.Fprintf(&buf, "- %s: %d\n", tc.Type, tc.Count)` → the gofmt three-line wrap in `if tc.Type != "failed" {` … `}` (the **markdown** `By type` narrowed; M-9 covers only JSON) | md `FailureSectionGolden`, `PartitionCovers…`, the e2e | md `FailureSectionGolden`, the e2e. **Over-predicted**: `PartitionCovers` sums only the JSON maps | →`d9ebd814e33b` |
| **V-5** | the two `if len(…) > 0 {` blocks swapped, so `## What didn't work` comes before `## Highlights` (LD2's order) | md `FailureSectionGolden`, `SectionsRenderOnly…`; the e2e survives | exactly those two. **Matched** | →`9c114f3483ee` |

**What V-4 tells ship:** the decision-to-test map credits DEC-050 rule 3 to
`PartitionCovers…` for its *"count sums"*. Those sums read only the JSON.
In markdown, rule 3 is held by the full-document golden and by the e2e's
AC-3 check. Each is enough, so there is no gap, but the map overstates what
one test does.

### Nothing else changed

The same frozen copy, both binaries, and **28 invocations**:
`impact` ×4, `wrapped` ×4 (`2026`, `2026 Q3`, json, `--type failed`),
`story` ×5 (`--audience me|manager|exec|skip`, two in json), `export` ×3,
`coverage` ×2, `list` ×3, `stats` ×2, `memory` ×2, `review` ×2 and `search`
×1. Every window reaches the four failures (`--since 2026-09-01`,
`--quarter`, `2026`, `--month`) except `review --week`. Each side was checked
before comparing: exit 0, empty stderr, at least 3 lines, and JSON that
parses. Then `cmp` ran with the timestamp lines dropped. **28 of 28
EQUAL**, and the store hash was unchanged.

The shape check did its job once. The first sweep ran `wrapped 2026-Q3`,
which should be two arguments. Both binaries printed the same `user error:
invalid year`, which compared **equal**. The guard flagged it as `rc=1` /
`short(0)` and the result was thrown away. The corrected `wrapped 2026 Q3`
row renders the digest.

### Consumers of the changed shape. No unupdated reader

`git grep -n -E
'highlights|failures_by_project|summary --|ToSummary|NewSummaryCmd|brag
summary|counts_by_type'` over `docs` (minus `engineering-practices.md`),
`BRAG.md`, `README.md`, `internal/mcpserver`, `plugin`, `.claude-plugin`,
`examples`, `cmd`, `GETTING_STARTED.md`, `scripts`, `justfile` and
`.github`, plus `.claude` without `worktrees`:

| Hit | Reads the JSON shape? |
|---|---|
| `BRAG.md:438`, `README.md:222-223`, `docs/blog/why-bragfile.md:84` | no: example commands |
| `cmd/brag/main.go:47` | no: registration |
| `docs/api-contract.md:348-392` | **yes**: the section, updated by build (keys, both arrays, the asymmetry) |
| `docs/api-contract.md:494, 553, 587`, `docs/tutorial.md:496, 536` | no: `summary` named in passing |
| `docs/api-contract.md:579, 594, 677` | `impact`/`wrapped`'s own `failures_by_project` (SPEC-086) |
| `docs/api-contract.md:1531`, `docs/data-model.md:217` | DEC-014's one-line description: no key list |
| `scripts/test-docs.sh` | Groups A, AD and AF: command lists and needles |
| `internal/mcpserver`, `plugin`, `.claude-plugin`, `examples`, `.claude` | **0** hits for `summary` in any form |

### AC re-run

AC-1 to AC-6 ran on a fresh scratch store written by `brag-cur` at run time
(`W=1`, `F=2`). AC-1: 1 / 0 / 1 / 0 (each positive paired). AC-2: `[1]`,
`[2]`, and the 7-key order. AC-3: all three count lines present, `failed`
1, `alpha` 2. AC-4: `## Summary|## What didn't work|`, `true`, 1. AC-5:
`## Summary|## Highlights|`, `true`, 1. AC-6: the LD6 sentence, verbatim, 1
match. AC-7: all six new tests exist once each, and
`TestToSummaryMarkdown_DEC014FullDocumentGolden`'s body is `cmp`-identical to
`dfee276`'s (77 lines). AC-10: DEC-050 lines 1–309 are `cmp`-identical to
`55063e9`, and the diff since has **0** removed lines.

### STAGE-023 candidate 2: its named test ran

The candidate says its nearest test is *"SPEC-094's `M-D1` at `c3f707a`, once
SPEC-095 has edited `docs/api-contract.md`"*. SPEC-095 did: the file moved
from `1cd16e2a3d7a` to `3da3b8d8a35f` at `6590772`. Starting from
`git show c3f707a:docs/api-contract.md` (`1cd16e2a3d7a`), the row's stated
edit has **1** occurrence and hashes to **`471eb6137945`**, exactly as
stated. That is the payoff the candidate describes: *a spec other than
SPEC-086 reproduces a `main`-named row after its target moved*. It counts
toward **N=2 paired-opposing** against `V-F0`. **Reported, not ruled.** Ship
codifies.

One detail from reproducing it: the row's old side is a markdown code span
that visually ends in `(an integer, `. CommonMark strips one leading and one
trailing space from a code span, and the real text ends at the comma. A
first attempt that kept the space matched **0** times, and the uniqueness
guard is what refused it. Worth one line if ship writes the rule: a pinned
edit with edge whitespace should not live inside a code span.

### Gates: all five green on this branch

| gate | result |
|---|---|
| `just test` | exit 0 · `go test -count=1 -v ./...`: **1119** `--- PASS` · **0** `--- FAIL` · **14** packages `ok` |
| `just test-docs` | exit 0 · **213** `OK:` / **212** distinct ids / **0** `SKIP:` / **0** `FAIL:` · `ALL OK` |
| `just lint` | `0 issues.` |
| `gofmt -l .` | empty |
| `go vet ./...` | exit 0 |

No derived number moved: this cycle changes one spec file, which no
inventory row counts, so the block was not regenerated. `X3` is green.

### What this cycle changed

- This spec only: M-9's row and the matrix preamble note (V-F1), the
  inventory table's `## Amendment` cells (V-F2), and this section. **No Go,
  test, doc or harness change.** `cycle:` stays `verify`. Ship advances it.
  *(Superseded by the second pass below: V-F3 adds 7 lines of assertions to
  one existing test, `internal/cli/learn_test.go`.)*

### Not findings, checked and clean

- **The corpus was never written to.** The only `brag` run against
  `~/.bragfile/db.sqlite` was §13.5's read-only `brag memory --project
  bragfile`, plus the `sqlite3 .backup`. The seed went into a scratchpad copy.
  **No brag was captured**; that comes at ship.
- **Counts** were taken with `/usr/bin/grep`, `git grep`, `jq` or Python, and
  diffs with `/usr/bin/diff`. The shell's `diff` is rewritten by a hook and
  printed both files whole on the first try, so that output was discarded.
- **Out of scope, untouched:** `story`, `impact`, `wrapped`, SPEC-096, DEC-014
  (the open maintainer question stays open), `guidance/questions.yaml` and
  `AGENTS.md`.
- **An observation, not routed:** `By type` on the live month window
  prints `- : 29`, for entries with an empty type. It is byte-identical on
  the pre-build binary, so it predates this spec.

### Second, independent pass (2026-09-23)

*A fresh session was handed the same verify brief after the pass above was
pushed. The maintainer chose an independent re-run: every attack below was
run before this pass read the section above, and then the two were
reconciled. The live corpus had moved, so it was re-frozen: 624 entries,
max id 646 (a `learned` row, not a failure), `sqlite3 .backup` SHA-256
`b831fbb2c11d…`, unchanged at the end, and the file kept read-only
(`chmod a-w`). Binaries: `brag-old` from `git archive dfee276` (`main`
before `build(SPEC-095)`) and `brag-new` from this branch.*

**Where the two passes agree** (each reproduced here independently):

- **M-9.** From base `3b81e79b3fdf` (`internal/export/summary.go` at
  `6590772`), the gofmt three-line wrap reproduces design's `e9404d9921ee`
  and the one-line wrap reproduces build's `c840a90180f2`. Both fire exactly
  `TestLearnCmd_SummarySectionsWhatItWrote`,
  `TestToSummaryJSON_FailureSectionGolden` and
  `TestToSummary_PartitionCoversEveryInWindowEntry`. V-F1's corrected row
  is right.
- **Real output** on the re-frozen copy. `--range week`: the markdown is
  byte-identical to `brag-old` with `Generated:` removed, and the JSON is
  identical after deleting `generated_at` and the new key, which is `[]`.
  `--range month`: `diff` is exactly the 4 deletions (433, 465, 473 from
  `### bragfile`, 420 from `### contextcore-pilot-harness`) plus the 12-line
  `## What didn't work`. `By type` still prints `- failed: 4`, and both
  count maps are `==` to `brag-old`'s. `highlights` holds 230 ids, down from
  234. `--range month --type failed`: `## Summary` then `## What didn't
  work`, no `## Highlights`, `"highlights": []`, `counts_by_type`
  `{"failed":4}`, key order ending `highlights,failures_by_project`.
  `--range week --type failed` (an empty window): provenance only, and JSON
  `highlights` and `failures_by_project` both `[]`. Every side was checked
  first: rc 0, empty stderr, first line `# Bragfile Summary` or `{`.
- **An impact-less failure, seeded** into a second copy with `brag learn -t
  "seeded impactless dead end" -p zz-seed` (id 647, `impact=''`). It is
  under `## What didn't work` in `--range week` and `--range month`, is in
  `failures_by_project` and absent from `highlights`, and is counted
  (`failed: 1`/`5`, `zz-seed: 1`) in both formats. `impact --since 1d`
  reads `Entries: 3/4 with impact` and lists id 647 nowhere in either
  format. That is the designed asymmetry.
- **Nothing else changed.** 17 more invocations were byte-identical across
  the two binaries, with `Generated`/`generated_at` removed: `impact
  --quarter` (both formats), `impact --month --previous`, `wrapped 2026`
  (both formats), `story --audience manager|exec --since 2026-09-01`,
  `export` (both formats), `coverage --quarter`, `coverage --year --format
  json`, `list`, `stats` (both formats), `memory --project bragfile`,
  `review --month` and `search failure`. **The shape guard fired on this
  pass's own first attempt:** bare `coverage` printed the same usage error
  from both binaries (0 stdout lines, rc 1), which compared *equal*. It was
  discarded and re-run with a window, as the section above did with
  `wrapped 2026-Q3`.
- **Consumers.** `git grep` for `"highlights"`, `.highlights`,
  `failures_by_project`, and `summary … json` across `docs`, `BRAG.md`,
  `README.md`, `AGENTS.md`, `plugin`, `.claude-plugin`,
  `internal/mcpserver`, `cmd`, `scripts` and `justfile` (excluding
  `docs/research` and `docs/blog`). The only reader of the shape is the
  contract, which build updated, and the `test-docs` guards on it.
  `README.md:223` is a bare command. `internal/mcpserver`, `plugin` and
  `.claude-plugin` have no `summary` mention at all.

**Five more mutants, none of them among M-1…M-10 or V-1…V-5.** The same
kind of helper, written fresh: it refuses an old side that does not occur
exactly once, refuses the gates until the hash moves, prints the diff it
applied, runs `go test -count=1 ./...`, restores with `cp` from `/tmp`, and
confirms the hash is back. All five went back to `3b81e79b3fdf`.
Predictions were stated before the first run.

| # | Edit (from the printed diff) | Predicted | Fired | Hash |
|---|---|---|---|---|
| **N-1** | JSON: the 3-line `for _, pc := range aggregate.ByProject(entries) {` loop plus the next line `worked, failed := aggregate.SplitFailures(entries)` (`:136`–`:139`) → the split line first, then the loop over `aggregate.ByProject(worked)` (`counts_by_project` stops counting failures) | JSON golden, `PartitionCovers…`; e2e survives | exactly those two. **Matched**. **After V-F3, the e2e fires too** | →`d74d20504e66` |
| **N-2** | the same reorder on the markdown path (`:57`–`:60`, the `By project` loop's `fmt.Fprintf(&buf, "- %s: %d\n", pc.Project, pc.Count)`) | md golden only | md golden only. **Matched**: one killer, which is V-F3. **After V-F3, the e2e fires too** | →`4e0c88901494` |
| **N-3** | `:68` `"## What didn't work")` → `"## What didn’t work")`, with U+2019 (a heading that looks right and matches nothing) | md golden, `PartitionCovers…`, `SectionsRenderOnly…`, the e2e | exactly those four. **Matched** | →`14d1b6842279` |
| **N-4** | `:141` `env.FailuresByProject = highlightGroups(failed)` → `…highlightGroups(failed)[:min(1, len(highlightGroups(failed)))]` (only the first failure group survives) | JSON golden, `PartitionCovers…`, `SectionsRenderOnly…`; the e2e survives (its one failure is one group) | exactly those three. **Matched** | →`38b07c3b128d` |
| **N-5** | `:141` the same line → wrapped in gofmt `if len(failed) > 0 {` … `}` (`null`, not `[]`, on a clean window; M-8 nulls both keys, V-1 omits this one) | `DEC014ShapeGolden`, `EmptyEntries…/json`, `SectionsRenderOnly…` | exactly those three (+ the parent `EmptyEntries…`). **Matched** | →`c16866f2771c` |

#### V-F3: `By project` had no end-to-end pin. FIXED

AC-3 names `- alpha: 2` under `**By project**` and
`.counts_by_project.alpha` → `2`, but `TestLearnCmd_SummarySectionsWhatItWrote`
checked only `By type` in both formats. So N-2, a CLI-visible break of
DEC-048 on the markdown path, was killed by exactly one renderer golden, and
N-1 by two renderer tests and no CLI test. The section above found the
`By type` half of this (V-4) and correctly called it *no gap*. `By project`
is the thinner half. It is still covered, so this is a strengthening, not a
defect. The test now asserts `- alpha: 2\n` in the markdown `## Summary`
and `counts_by_project["alpha"] == 2` in the JSON, 7 added lines. It passes
on the build, and **re-running N-1 and N-2 makes it fire on both** (hashes
as in the table). It adds no test function and no subtest, so no derived
number moves. **The decision-to-test map** row for rule 3 should now also
name the e2e, which ship can fold in with V-4's note.

#### Where the two passes disagree: the codification ruling

Both passes recommend the same **remedy**: a probe helper committed to
`scripts/` that emits the matrix row, meaning the diff it applied, the base
hash and its commit, from its own run. No new AGENTS.md sentence. They
disagree on **what M-9 counts toward**, and ship should decide with both
arguments in view.

The section above files M-9 under item 5 as a *new-side uniqueness* case
and holds candidate 1 at N=1. This pass argues M-9 is **candidate 1's
second negative**:

1. **Candidate 1's property is where the cell came from, not how many
   readings it has.** Design's helper applied one literal edit, because a
   hash cannot come from a sketch, and `e9404d9921ee` proves which edit it
   was. The cell was then written *afterwards, as a description*, with a
   `…` where the applied text had been. That is candidate 1's text word for
   word: *"record the diff the probe printed rather than a description
   written afterwards."* M-A0 and M-9 share this cause, and differ only in
   which symptom the description produced: a wrong reading, or several.
2. **Filing by symptom is how the three cases escaped.** Item 5's guard
   checks the old side. A new-side guard would check the new side. Every
   symptom-level clause leaves the next symptom outside it, which is the
   section above's own point 3. Filing by cause puts all three cases, and
   the next one, under the clause whose remedy closes them.
3. **M-D1 is the case that does not fit this reading.** SPEC-094's ship
   filed it under uniqueness, and it could have been either. This pass
   does not relitigate it, and counts only M-A0 and M-9.
4. **The paired positive is in this spec.** Every row that reached the
   matrix as literal text reproduced its hash first try in build and again
   here: design's other 15, V-1 to V-5 above, and N-1 to N-5 in this pass,
   all recorded from the diff the helper printed. So candidate 1 would be at
   **N=2 paired-opposing** (M-A0 and M-9 negative, printed diffs positive
   across two independent verify passes), which is this repo's bar.

**On "is more rule text the answer?":** no, and the two passes agree. The
one mechanical guard in this family (item 5's refusal) is what actually
held, and the misses are always where no guard looks. If ship promotes
candidate 1, the codified form should be one sentence pointing at a
committed helper, *the matrix cell is the diff `scripts/probe` printed*,
not a fourth description of a property. Until the helper exists, a
promoted clause is still discipline. That is the argument for routing the
helper as the first deliverable after v0.7.0 and codifying with it, rather
than before it.

#### Not findings

- `aggregate.IsFailure` is `e.Type == "failed"`, case-sensitive, so a
  hand-typed `-k Failed` stays under `## Highlights`. That is DEC-050 rule
  1's shared predicate, the same one `--type failed` uses and
  `TestFailureClassifier_GoPredicateMatchesTypeFilter` pins. It is not
  `summary`'s to change, and not routed.
- The `- : 29` empty-type line reproduces here too, identical on
  `brag-old`, so it predates this spec.
- Nothing was written to `~/.bragfile/db.sqlite`: one `.backup` and §13.5's
  `brag memory`. No brag was captured.
- **Gates, re-run on the final tree after V-F3:** `go test -count=1 -v
  ./...` **1119** `--- PASS`, **0** `--- FAIL`, **14** packages `ok` (the
  count is unchanged, because V-F3 adds assertions and no test);
  `test-docs` **213** `OK:` / **212** distinct / **0** `FAIL` / **0**
  `SKIP`, `ALL OK`; `just lint` `0 issues.`; `gofmt -l .` empty; `go vet`
  clean. No derived number moved, so the inventory was not regenerated.

---

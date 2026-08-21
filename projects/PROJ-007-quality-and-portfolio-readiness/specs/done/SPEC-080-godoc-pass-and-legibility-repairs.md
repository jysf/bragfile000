---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-080
  type: chore                      # epic | story | task | bug | chore
  cycle: ship
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-007
  stage: STAGE-021
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-19

references:
  decisions:
    # The page and the new package comments cite these; the packages'
    # boundaries are stated against them, and DEC-041's tombstone points at
    # DEC-040/DEC-042 as its numeric neighbors.
    - DEC-001                      # pure-Go SQL driver — cited in the storage package comment
    - DEC-003                      # --db resolution order — cited in the config package comment
    - DEC-009                      # editor buffer format — answers editor-template-format
    - DEC-014                      # rule-based output shape (by-project/by-type convention) — answers summary-grouping-heuristics
    - DEC-017                      # entries/project relationship — cited in the storage package comment
    - DEC-019                      # project-here resolution policy — cited in the storage package comment and the DEC-041 tombstone
    - DEC-020                      # project-location editing semantics — cited in the storage package comment and the DEC-041 tombstone
    - DEC-021                      # migration auto-backup — cited in the storage package comment
    - DEC-026                      # dev/prod migration guard — cited in the storage and config package comments
    - DEC-029                      # story audience profiles — cited in the story package comment
    - DEC-040                      # numeric neighbor of the DEC-041 tombstone
    - DEC-042                      # numeric neighbor of the DEC-041 tombstone
  constraints:
    - test-before-implementation   # blocking, paths ["**"] — Failing Tests below are written at design
    - one-spec-per-pr              # blocking, paths ["**"]
    - no-sql-in-cli-layer          # blocking, ["internal/cli/**"] — the internal/cli package comment restates it
    - no-secrets-in-code           # blocking, paths ["**"] — repo-wide, nothing here handles secrets
  related_specs:
    - SPEC-079                     # the practices entry point this completes the stage alongside; LD1's inventory.sh pattern this spec extends
    - SPEC-070                     # deferred; holds the unwritten DEC-041 reservation
    - SPEC-009                     # brag edit + editor package — origin of DEC-009, which answers editor-template-format
    - SPEC-010                     # brag add no-args editor launch — reuses DEC-009's format
    - SPEC-018                     # brag summary + aggregate package — origin of GroupForHighlights, which answers summary-grouping-heuristics
---

# SPEC-080: godoc pass and the two legibility repairs

> **Cycle: design.** The go/no-go lives in this spec's git history (framed
> 2026-08-19, GO at complexity S). This revision re-measures the framing's
> claims against the repo (both were wrong), settles the two forks framing
> left open with rejected alternatives recorded, and embeds the literal
> artifacts. Per AGENTS.md §6, build happens in a **fresh session** and
> transcribes the literals under `## Notes for the Implementer` verbatim.

## Context

STAGE-021's second and final spec. SPEC-079 built the practices entry point;
this one closes the three remaining scope items — a godoc pass, the
`DEC-041` gap, and open-questions hygiene.

Parent: `STAGE-021-make-the-discipline-legible`, spec 2 of 2.
Project: `PROJ-007`.

## Goal

`go doc` on this codebase reads as intentional, `decisions/` has no
unexplained numbering gap, and `guidance/questions.yaml` describes live
uncertainty only.

## Re-measurement: the framing spec's own numbers were wrong

Framing measured "175 exported declarations, 3 undocumented, 15 packages, 7
lacking a package comment" on 2026-08-19. Re-measured at design, against the
**same, unchanged tree** — `git log --since=2026-08-14 -- internal/ cmd/`
returns no commits, so this is not drift, it is a **counting-method error**
in the framing pass:

| | Framing claimed | Re-measured (AST-exact) |
|---|---:|---:|
| Exported declarations | 175 | **191** |
| …lacking a doc comment | 3 | 3 (confirmed) |
| Packages | 15 | 15 (confirmed) |
| …lacking a package doc comment | 7 | **5** |

**The exported-declaration undercount.** A grep-shaped heuristic undercounts
by 16 because it does not credit a doc comment on a parenthesized
`const ( … )` / `var ( … )` block as documenting every name inside it — which
*is* the rule `go/doc` and `go doc` both follow. `internal/capture/validate.go`
(7 named caps), `internal/memory/memory.go` (7 named constants), and
`internal/story/thread.go` (2 named constants) each carry exactly this shape:
one doc comment above the block, no doc comment on each individual name. A
tool built via Go's own `go/parser` + `go/ast` (`parser.ParseComments`,
crediting a `GenDecl`'s `Doc` to every `Spec` inside an undocumented-at-the-
spec-level parenthesized group) gives 191, and the same tool independently
reproduces the "3 undocumented" figure exactly — `(*ErrQuery).Error`,
`(*ErrQuery).Unwrap` (`internal/memory/pool.go:49-50`), and
`(*tagsField).UnmarshalJSON` (`internal/cli/add_json.go:22`) — so the
declaration-count method is what was wrong, not the undocumented list.

**The missing-package-comment overcount.** `internal/export` and
`internal/mcpserver` were named as missing; both already carry a package
comment (`internal/export/json.go:1-4`, `internal/mcpserver/provenance.go:1-5`
— re-confirmed via a live `go doc ./internal/export` / `go doc
./internal/mcpserver` run, not just grep). The real gap, unchanged in
substance from framing's finding, is five packages — still the largest and
most load-bearing in the tree, which is how the absence went unnoticed:

```
cmd/brag   internal/cli   internal/config   internal/storage   internal/story
```

This **shrinks** the spec's real surface area (5 comments, not 7) without
changing its shape or its complexity rating.

## Scope

Three items, all from STAGE-021's In-scope list.

1. **Five package doc comments** — `cmd/brag`, `internal/cli`,
   `internal/config`, `internal/storage`, `internal/story` — each saying what
   the package is *for* and what its boundary is, not restating its name.
2. **The `DEC-041` gap.** `decisions/` jumps 040 → 042. Settled below (Fork
   1): a reservation tombstone lands in `decisions/`.
3. **Open-questions hygiene.** Settled below (Fork 2 + the per-question
   triage): the hygiene number is now derived, and 2 of the 8 open questions
   are closed as answered-in-practice.

## Fork 1 — what shape does the `DEC-041` marker take?

**Decided: a tombstone file lands in `decisions/`, marked
`insight.type: reservation` (not `decision`), and `inventory.sh`'s Decision
records count is redefined to filter on that field rather than the bare
`decisions/DEC-*.md` glob.**

`decisions/DEC-041-reserved-brag-project-goto-multi-location-primary-policy.md`
(literal ② below) names the deferred feature holding the number (`brag
project goto`, SPEC-070), states the un-made choice it would decide
(first-registered-location = primary), links the full design preserved in
`projects/PROJ-001-mvp/backlog.md`, and states its own disposition (if built,
the decision replaces this file's content; if the backlog item is deleted
outright, this file says so; the number is never reused for anything else —
the backlog's own words).

**Why a tombstone beats leaving the gap explained only in prose (framing's
third candidate).** SPEC-079 already put "`DEC-041` is reserved rather than
lost" into `docs/engineering-practices.md`'s prose, pointing at the backlog —
and that was not enough, which is exactly why STAGE-021 kept this item open. A
reader who does `ls decisions/` — the natural first move for anyone
evaluating whether this repo's decision log is complete — sees `040`, `042`,
and nothing between them. The practices page is one path into the repo, not
the only one; a directory listing is a more fundamental one, and it is the
literal shape of STAGE-021's success criterion ("`decisions/` has no
unexplained numbering gap") — worded about the directory, not about the docs
page.

**The constraint the framing spec flagged, worked through.** A tombstone
living in `decisions/DEC-*.md` would be counted as a decision by the
*existing* `inventory.sh` line (`ls decisions/DEC-*.md | wc -l`), which would
silently inflate "Decision records" from 45 to 46 — a wrong number the moment
the tombstone lands, in the one document on the whole page that is guarded by
`X3`'s diff. The fix is **not** to exclude the tombstone from the directory
(that defeats the point) and **not** to leave the count wrong (that violates
the whole thesis of `docs/engineering-practices.md`): it is to make
"Decision records" mean what it has always informally meant — *actual
decisions* — explicitly, by filtering on a field every one of the 45 existing
records already carries consistently
(`grep -l '^  type: decision' decisions/DEC-*.md` matches all 45, confirmed
at design; `grep -L` finds zero exceptions). The tombstone's own
`insight.type: reservation` is what excludes it.

**What happens to the `Decision records` row, precisely.** It stays **45** —
unchanged, because the filter now counts what it always meant to count. A new
sibling row, **`Decision numbers reserved, not yet decided | 1`**, makes the
reservation itself a derived, guarded number rather than a fact only readable
in the page's prose — closing the same defect class LD1 (SPEC-079) named:
*"if you find yourself wanting to type a current-state number into the
prose, add a row here instead."* `X3`'s existing script-vs-page diff already
re-verifies this on every run; a **new** assertion (`Y3`, below) additionally
pins the *semantic* values (45 and 1), because `X3` alone would happily pass
if the type filter ever silently broke and both sides agreed on a wrong
number.

**Rejected alternatives (build-time):**

- **Release the number, let SPEC-070 take the next free id if ever built —
  REJECTED.** `projects/PROJ-001-mvp/backlog.md`'s own DEC-041 note says so
  explicitly: *"Do not reuse the number for anything else."* Overriding an
  explicit prior instruction to make this spec simpler is not this spec's
  call to make.
- **Leave the gap unexplained in `decisions/` itself, rely on the practices
  page's prose — REJECTED**, per above: that is the status quo SPEC-079
  already shipped, and STAGE-021 named the gap as still open *after* that
  page existed.
- **Exclude the tombstone from `decisions/DEC-*.md` entirely (a different
  directory or filename shape) — REJECTED.** It would stop `ls decisions/`
  from resolving the gap at all, defeating Fork 1's entire premise.
- **Retrofit all 45 existing records with an explicit `insight.type:
  decision` if it were missing — NOT NEEDED.** Verified at design: every
  existing record already carries it (`grep -L` found zero exceptions), so
  the filter is a pure addition with no retrofit cost.

**A known, deliberate side effect, not fixed here.** `scripts/status.sh`'s
casual `just status` summary line (`total_decisions=$(count_matching
"$decisions_dir" "DEC-*.md")`) still counts by bare filename glob, so it will
report 46, not 45 — a minor, informational-only inconsistency against the
*guarded* 45 on the practices page. `status.sh` is not part of the
`test-docs.sh`-guarded system SPEC-079 built and is out of this spec's named
scope; its low-confidence-decisions loop (the only *other* place it reads a
DEC file's front-matter) already tolerates a missing `confidence:` field
safely (`[ -n "$conf" ]` guards it), so the tombstone does not break it,
it just isn't counted by the same rule. Left as-is rather than silently
widening scope; a future spec touching `status.sh` should apply the same
`insight.type: decision` filter for consistency.

## Fork 2 — does the hygiene pass derive its own count?

**Decided: yes.** `inventory.sh` gains two rows — `Questions tracked in
guidance/questions.yaml` and `…of those, still open` — computed the same way
LD1 (SPEC-079) computes every other current-state number: derived by the
script, pasted into the guarded block, diffed by `X3` on every run.

**Why, weighed against the row count the page already carries (framing's
question).** The page already carries 16 rows; three more (one for Fork 1,
two for Fork 2) is a 19% growth in a table whose whole purpose is being the
place current-state numbers live rather than rot in prose. That is exactly
what the table is *for* — LD1 explicitly anticipated it growing: *"every
number describing the repository's current state lives inside the guarded
block."* The alternative is a fourth hand-typed restatement of a number that
has now been wrong three times in three days, in the same file that already
demonstrates, in its own `## Decisions` section, the fix for exactly this
failure mode.

**The bug that must not recur, and how the new computation avoids it.** The
"9 of 18" miscount came from `grep -c 'status: open'` with no anchor, which
matched the file's own header comment — `#   - status: open | investigating
| answered` — as a false positive, because that literal substring is
present in the comment too. The fix is not "be more careful"; it is
**anchoring to the exact structural shape only a real entry has**: every
question entry's `status:` field is indented **4 spaces** (a nested key
under a `- id:` list item), while the comment line begins with `#`. The new
computation, `grep -cE '^    status: open$'`, cannot match the comment under
any input — verified at design (§12(b) row 3 below): it returns 8 against the
*current* register, matching the number STAGE-021 already (correctly, this
time) carries. The total similarly anchors on `^  - id: ` (2-space list-item
indent), returning 18.

**Rejected alternatives (build-time):**

- **A `just` recipe only, no page row — REJECTED**, same reasoning LD1
  already rejected this shape for: "here is how to get the number" is a
  weaker page for a reader who will not clone the repo, and the recipe
  (`just inventory`) already exists and already prints this row — the
  rejection is of *omitting the row*, not of the recipe.
- **Restate the number a fourth time, more carefully — REJECTED.** This is
  the exact move that failed three times already; "wrong three times in
  three days" is direct evidence a hand-restated number in this specific
  file is not a durable fix, independent of how carefully any one restatement
  is done.
- **Fold "open" and "total" into one combined cell (e.g. `"8 of 18"`) —
  REJECTED.** `inventory.sh`'s numeric column (`|---:|`) is one value per
  row throughout the table; a combined string breaks that shape and cannot
  be `awk '$1 >= …'`-style validated the way every other row can be. Two
  rows, matching the existing "N / …of those, subset of N" pattern already
  used for decisions and specs, is the established idiom.

## The per-question triage

STAGE-021 asked for a **disposition**, not a restated count, for each of the
8 open questions. Each question's own stated revisit trigger was read and
tested against the repo. Two answered themselves; one gets a sharpened
resolve condition; five are genuinely, verifiably still open and are left
untouched.

| Question | Raised | Disposition | Evidence |
|---|---|---|---|
| `editor-template-format` | 2026-04-19 | **CLOSED — answered-in-practice** | `DEC-009` (2026-04-20, the very next day) locked RFC822-style `net/textproto` headers, explicitly rejecting YAML front-matter. `internal/editor/editor.go`'s shipped `Render`/`EmptyTemplate`/`Parse` — and the package doc comment this spec adds — still say exactly that. |
| `summary-grouping-heuristics` | 2026-04-19 | **CLOSED — answered-in-practice** | `internal/aggregate/aggregate.go`'s `GroupForHighlights` (shipped, SPEC-018/DEC-014): entries bucket by **project only** (alpha-ASC, `(no project)` last), chrono-ASC within each bucket. `type` gets its own separate flat count list, never nested; tags are not a grouping axis at all — exactly the "leaning toward" guess the original note recorded, never closed. |
| `shareable-ids` | 2026-04-19 | **STAYS OPEN — trigger sharpened** | Trigger is "sync or cross-machine export enters scope." `docs/research/duckdb-federation-spike.md` is real, relevant prior art (it names "bragfile has no global entry identity") but is explicitly a pre-framing *spike report*, never pulled into a shipped feature — the trigger has not fired. Note updated to say so, so a future reader does not mistake the spike's existence for the trigger firing. |
| `tag-rename-into-existing` | 2026-06-07 | **STAYS OPEN — untouched** | Trigger is a real user report that the two-step rename→error→merge flow is annoying. No mechanically-checkable repo state settles a UX preference; nothing to test. |
| `spark-month-semantic-drift` | 2026-07-10 | **STAYS OPEN — untouched** | Trigger is dogfooding `brag spark --month` and finding the rolling-28-day reading misleading against other commands' calendar-month `--month`. `brag spark` **is** shipped (SPEC-059, verified: `internal/cli/spark.go` exists and is wired into `cmd/brag/main.go`) — so the command exists to be dogfooded — but a search of the live corpus (`brag list --format json`, 10 hits on "spark") surfaces no entry recording that the ergonomic wart has actually bitten. The trigger requires a *subjective read*, not a fact this pass can check for the user. |
| `memory-slice-fusion-constants` | 2026-08-08 | **STAYS OPEN — untouched** | Trigger is "if the top of the [memory] slice reads wrong" during dogfooding — an inherently subjective judgment call the question's own text says needs to be made by using the command, not by grepping the repo. |
| `json-input-transport-stdin-only` | 2026-08-12 | **STAYS OPEN — untouched** | The question's own disposition note already records the user's explicit call: *"NOT WORTH BUILDING NOW … but we should log for the future."* Deliberately open by design; nothing to action. |
| `dec-amendment-heading-convention` | 2026-08-16 | **STAYS OPEN — untouched** | Resolve conditions: derived `## Amendment` count reaches 4+ (still 1, confirmed by this spec's own `inventory.sh` row), a reader asks for the number, or PROJ-007 closes with the count still at 1 (PROJ-007 has not closed — STAGE-022 is still ahead). None have fired. |

**Result: open-questions count moves from 8 to 6, of 18 total (unchanged —
no questions added or deleted).** `guidance/questions.yaml`'s intake rules
offer three dispositions for a settled question — mark answered and link the
answer, delete if resolved informally, or restate the resolve condition. Both
closures here get the first treatment (matching the register's own established
pattern for `tags-storage-model`, `project-state-note-shape`, and
`tag-ordering-projection`): the original entry is **retained below a
provenance divider**, not deleted, because losing an already-considered
alternative (YAML front-matter; type-first grouping) to a future reader is a
worse outcome than a longer file.

**Did not invent an answer for anything.** Every one of the five still-open
questions has an explicit trigger that a specific, checkable fact would need
to change — a user report, a dogfooding read, a corpus count crossing a
threshold, or an explicit already-recorded user call. None of those facts
changed at this pass. `shareable-ids` is the one edge case: real, relevant
prior art exists (the federation spike) without the trigger itself having
fired, so its note is sharpened to say precisely that distinction rather than
either closing it (inventing an answer) or leaving a future reader to
wonder whether the spike *was* the trigger.

## Out of scope (for this spec specifically)

- **The README restructure** — promoting the MCP call-to-action out of the
  Status blockquote, and switching `test-docs` A1 from `wc -l` to `wc -w`.
  Agreed as a separate, third STAGE-021 spec, sequenced after this one.
  `README.md` is untouched by this spec.
- Anything in STAGE-022 (lint, coverage, the `Entries:` envelope).
- Rewriting `godoc` prose that already exists in the 10 packages that have
  it (re-measured: not 8, see above).
- Answering an open question that is genuinely still open — five of the
  eight, per the triage above.
- Fixing `scripts/status.sh`'s casual decision count — see Fork 1's "known,
  deliberate side effect."

## §12(b) design-time verification

Every literal below was run through its target tool during design —
gofmt, go vet, go build, `go test ./...`, and a full `./scripts/test-docs.sh`
run — with all five package-doc-comment edits, the DEC-041 file, and the
`inventory.sh`/`test-docs.sh`/`questions.yaml`/`docs/engineering-practices.md`
changes staged locally in the working tree, then reverted (per AGENTS.md §6,
design does not commit build's changes — only this spec file and STAGE-021's
backlog line are part of this commit). Nothing below is a prediction.

| # | What was pre-flighted | Result |
|---|---|---|
| 1 | A `go/parser`+`go/ast` AST tool (not grep) counting exported declarations and undocumented ones, crediting a `GenDecl`'s doc comment to every `Spec` in a parenthesized block | **191 exported, 3 undocumented, 15 packages, 5 missing a package comment** — the corrected figures used throughout this spec. |
| 2 | `git log --since=2026-08-14 -- internal/ cmd/` | **No commits.** Confirms the framing-vs-design delta is a counting-method error, not code drift, between 2026-08-19 and now. |
| 3 | `go doc ./internal/export` and `go doc ./internal/mcpserver` (the actual rendering surface, not just source-grep — §12(b) refinement: behavior surface over shape surface) | Both render their existing package comment in full; framing's "missing" claim for these two is false. |
| 4 | All five package doc comments applied; `gofmt -l .` | **Empty** (clean) on all five files. |
| 5 | `go vet ./...` and `go build ./...` with all five comments applied | **Exit 0**, both. |
| 6 | `go test ./...` with all five comments applied | **All packages pass** (`internal/storage`, `internal/cli`, `internal/config`, `internal/story`, `cmd/brag` included). |
| 7 | `grep -l '^  type: decision' decisions/DEC-*.md \| wc -l` vs `ls decisions/DEC-*.md \| wc -l`, against the 45 existing records | **Both 45** — the type filter is a pure addition; zero existing records lack the field (`grep -L` returns nothing). |
| 8 | The DEC-041 tombstone staged; `./scripts/inventory.sh` re-run | `Decision records` row: **45** (unchanged). New `Decision numbers reserved, not yet decided` row: **1**. |
| 9 | `grep -cE '^    status: open$' guidance/questions.yaml` against the **unmodified** register (baseline, before closing the two questions) | **8** — matches STAGE-021's currently-stated (correct, this time) count, confirming the anchored pattern does not reproduce the header-comment false positive. |
| 10 | The same query, and `^  - id: ` for the total, after marking `editor-template-format` and `summary-grouping-heuristics` answered | **Open: 6. Total: 18** (unchanged — no entries added or removed). |
| 11 | `internal/aggregate/aggregate.go`'s `GroupForHighlights` and `internal/export/summary.go`'s `ToSummaryMarkdown` read directly | Confirms project-only grouping, alpha-ASC with `(no project)` last, chrono-ASC within group — the exact claim the `summary-grouping-heuristics` closure makes. |
| 12 | `internal/editor/editor.go` read directly (`Render`, `EmptyTemplate`, `Parse`) | Confirms RFC822-header format, matching DEC-009 exactly — the claim the `editor-template-format` closure makes. |
| 13 | `docs/engineering-practices.md` line count after all edits | **283** lines — inside `X7`'s 150–300 band, no re-pin needed (unlike SPEC-079's A1, which was at its ceiling). |
| 14 | Full `./scripts/test-docs.sh` run with every artifact staged | **All assertions pass**, including the new `Y1`–`Y5` and the re-diffed `X3`. `just test`, `gofmt -l .`, `go vet ./...` unaffected. |
| 15 | `scripts/status.sh`'s low-confidence-decisions loop, read directly, against a DEC file with no `confidence:` field | `[ -n "$conf" ]` guards it — the tombstone (which deliberately carries no `insight.confidence`) is silently skipped, not a crash. Confirms the Fork-1 side-effect note is accurate. |

## Outputs

- **Files created:**
  - `decisions/DEC-041-reserved-brag-project-goto-multi-location-primary-policy.md`
    — the reservation tombstone. **70 lines**, transcribed verbatim from
    literal ②.
- **Files modified:**
  - `cmd/brag/main.go`, `internal/cli/root.go`, `internal/config/config.go`,
    `internal/storage/store.go`, `internal/story/bundle.go` — one package
    doc comment inserted immediately before each `package` declaration, from
    literal ①. No other line in any of these five files changes.
  - `scripts/inventory.sh` — from literal ③: the `decs=` computation is
    redefined to filter on `insight.type: decision`; a new `decs_reserved`
    variable and table row; two new variables (`questions_total`,
    `questions_open`) and two new table rows.
  - `scripts/test-docs.sh` — from literal ④: new **Group Y** (`Y1`–`Y5`)
    appended after Group X, before `# ===== finalise =====`.
  - `guidance/questions.yaml` — from literal ⑤: `editor-template-format` and
    `summary-grouping-heuristics` marked `status: answered` with resolution
    notes (original text retained below a provenance divider, matching the
    file's own established pattern); `shareable-ids`'s note gains one
    paragraph sharpening its resolve condition. No entry is deleted; no
    entry is added.
  - `docs/engineering-practices.md` — from literal ⑥: the guarded inventory
    block re-pasted (one row's "Where it lives" text changes, one row is
    added for the reservation, two rows are added for questions, the
    doc-assertion count moves 171 → 176); the `## Decisions` section's second
    paragraph rewritten to point at the tombstone file directly, ahead of
    the backlog pointer.
- **New exports:** none. No exported Go identifier is added, renamed, or
  removed — every code change is a comment.
- **Database changes:** none.

### Premise audit (§9), run at design against the repo

The additive case applies twice (a new decision-adjacent file lands in a
directory whose count is asserted; the doc-assertion count grows), both
executed, not merely enumerated:

| Grep | Hits | Disposition |
|---|---|---|
| `decisions/DEC-\*\.md` or `decisions/DEC-` in `scripts/*.sh` (which scripts read the directory besides `inventory.sh`) | `status.sh` (casual summary count), `weekly-review.sh` (lists filenames only, no count), `lifetime-report.sh` (its own separate count) | **`status.sh`'s count is now off-by-one from the guarded figure** (46 vs 45) — recorded above as a known, deliberate, out-of-scope side effect. `weekly-review.sh` is unaffected (lists filenames, asserts nothing). `lifetime-report.sh` uses its own independent `find … \| wc -l`, also now off-by-one from the guarded figure for the same reason — same disposition as `status.sh`: informational-only, not guarded by `test-docs.sh`, out of this spec's named scope. |
| `assert_line_count_band "X7"` in `scripts/test-docs.sh` | 1 (unchanged assertion, existing 150–300 band) | **No change needed.** Confirmed at design (§12(b) row 13): the page lands at 283 lines, inside the existing band with headroom — unlike SPEC-079's A1, no re-pin is required. |
| `docs/engineering-practices.md`'s `## Decisions` section for other mentions of "`DEC-041`" or "backlog.md" that might now read inconsistently with the new tombstone-first phrasing | 1 (the paragraph being edited) | **No other hit.** The rewritten paragraph is the only place the page discusses DEC-041. |

## Locked design decisions

Both forks framing left open are settled above, with rejected alternatives
recorded (AGENTS.md §12, "Decide at design time when decidable"). Summarized:

- **LD1 (Fork 1).** A reservation tombstone lands in `decisions/`, typed
  `insight.type: reservation`; `inventory.sh`'s Decision records count is
  redefined by that field, staying at 45; a new derived row counts the
  reservation (1) separately.
- **LD2 (Fork 2).** The open-questions count becomes a derived `inventory.sh`
  row (two rows: total, open), anchored on the real entries' 4-space
  `status:` indent to structurally exclude the file's own header comment —
  the exact class of bug that produced "9 of 18."
- **LD3 (per-question triage).** Two of eight open questions close as
  answered-in-practice (`editor-template-format` via DEC-009,
  `summary-grouping-heuristics` via SPEC-018/DEC-014); one gets a sharpened
  resolve condition (`shareable-ids`); five stay open untouched because their
  triggers require a subjective dogfooding read or an external report, not a
  repo fact this pass can check.
- **LD4 (measurement correction).** The framing spec's own counts were wrong
  — 191 exported declarations (not 175), 5 packages missing a comment (not
  7) — from a grep-shaped counting method that does not credit a
  parenthesized block's doc comment to the names inside it, and from a
  "missing" claim for two packages (`internal/export`, `internal/mcpserver`)
  that already had comments. Corrected before locking any other decision in
  this spec, per the explicit "re-measure everything" requirement.

## Acceptance Criteria

- [ ] All five named packages carry a package doc comment saying what the
      package is *for*, not restating its name (`Y1`).
- [ ] `decisions/DEC-041-*.md` exists, is marked `insight.type: reservation`,
      and is not counted in "Decision records" (`Y2`, `Y3`).
- [ ] `inventory.sh`'s Decision records row stays 45; a new row reports 1
      reserved decision number (`Y3`).
- [ ] `inventory.sh`'s open-questions rows report the true, derived values —
      18 total, 6 open — and `X3` confirms the page matches (`Y4`, `X3`).
- [ ] `editor-template-format` and `summary-grouping-heuristics` are marked
      `status: answered` in `guidance/questions.yaml`, each citing what
      settled it (`Y5`).
- [ ] The five still-open questions are unchanged in disposition; `shareable-ids`
      carries a sharpened resolve-condition note (verified by reading the
      file; no assertion — restated prose is not itself a mechanically
      checkable fact, and inventing one would misstate what changed).
- [ ] `docs/engineering-practices.md`'s inventory block matches
      `./scripts/inventory.sh` byte-for-byte (`X3`); the page stays within
      its 150–300 line band (`X7`, no re-pin required).
- [ ] `just test-docs` passes with 176 distinct assertions; `just test`,
      `gofmt -l .` and `go vet ./...` are unaffected.

## Failing Tests

Written during **design**, BEFORE build. Every locked decision above has at
least one test that would fail without it (AGENTS.md §9). All assertions
live in **`scripts/test-docs.sh`**, new **Group Y**; the literal source is ④.

- **`"Y1"`** — asserts each of the five named files carries a `//` comment
  line immediately preceding its `package` declaration. Pins the godoc-pass
  scope itself. *(Fails today: none of the five files have one yet.)*
- **`"Y2"`** — asserts `decisions/DEC-041-*.md` exists and contains
  `  type: reservation`. Pins **LD1**'s tombstone. *(Fails today: no such
  file.)*
- **`"Y3"`** — asserts `./scripts/inventory.sh`'s live output contains
  `Decision records | 45 |` and `Decision numbers reserved, not yet decided
  | 1 |`. Pins **LD1**'s semantic values directly — not merely `X3`'s
  script-vs-page self-consistency, which would pass even if both sides
  silently agreed on a wrong number after a future edit to the type filter.
  *(Fails today: `inventory.sh` has neither the filter nor the new row; its
  current "Decision records" line has no `insight.type:` qualifier text and
  no reserved-count row exists at all.)*
- **`"Y4"`** — asserts `./scripts/inventory.sh`'s live output contains
  `Questions tracked in guidance/questions.yaml | 18 |` and `of those, still
  open | 6 |`. Pins **LD2**'s semantic values, same rationale as `Y3`.
  *(Fails today: neither row exists; even after adding the rows alone
  — without also closing the two questions in `Y5` — this would read `18`/`8`,
  not `18`/`6`, so `Y4` and `Y5` are cross-checking, not redundant.)*
- **`"Y5"`** — asserts `guidance/questions.yaml`'s `editor-template-format`
  and `summary-grouping-heuristics` entries each have `status: answered`.
  Pins **LD3**'s two closures. *(Fails today: both are `status: open`.)*

- **`scripts/test-docs.sh` — existing assertions affected**
  - `"X3"` — unchanged assertion, new expected content. The counting guard
    re-diffs the (now 19-row) inventory block against the live script;
    without the `docs/engineering-practices.md` re-paste, `X3` fails on the
    added rows and the doc-assertion-count row (171 → 176). Verified at
    design (§12(b) row 14): with all six literals staged, `X3` is the only
    assertion that would fail if the page were *not* re-pasted, and passes
    once it is.
  - No other existing assertion changes shape or expected value. `X7`
    continues to pass unmodified (283 lines, inside 150–300).

### §12 NOT-contains self-audit

No new NOT-contains assertion is introduced by this spec (`Y1`–`Y5` are all
positive/structural checks), so the self-audit does not apply to a new
assertion. The existing `X5` (`rigorous|comprehensive|…`) and `B`-series
NOT-contains assertions are re-verified at design against every piece of
load-bearing prose this spec adds — the five package comments, the DEC-041
tombstone, and the rewritten `## Decisions` paragraph — with zero hits
(confirmed by reading each literal below; none uses any of `X5`'s forbidden
tokens or the `B`-series forbidden phrases, and none of this spec's changes
touch `README.md` at all).

## Implementation Context

*Read this section (and the files it points to) before starting the build
cycle.*

### Decisions that apply

- `DEC-001`, `DEC-003`, `DEC-017`, `DEC-019`, `DEC-020`, `DEC-021`, `DEC-026`,
  `DEC-029` — each is cited by name inside one of the five new package
  comments (literal ①); build must not paraphrase the citation beyond what
  the literal already says.
- `DEC-009` — the record that answers `editor-template-format`; the closure
  note in literal ⑤ restates its Option E choice. Do not re-derive it; the
  DEC itself is the source of truth if a discrepancy is ever suspected.
- `DEC-014` — the record whose §3 by-type/by-project convention is what
  `summary-grouping-heuristics`'s closure cites.
- `DEC-040`, `DEC-042` — the tombstone's numeric neighbors; not otherwise
  touched.

### Constraints that apply

- `test-before-implementation` — blocking, `["**"]`. Group Y above is
  written; build makes it pass and does not add assertions beyond literal ④
  without saying so under Deviations.
- `one-spec-per-pr` — blocking, `["**"]`. One PR, referencing SPEC-080.
- `no-sql-in-cli-layer` — blocking, `["internal/cli/**"]`. The
  `internal/cli` package comment (literal ①) restates this constraint as
  part of stating the package's boundary; build must not add an import that
  would make the comment false.
- `no-secrets-in-code` — blocking, `["**"]`. Nothing here handles
  credentials; listed because it is repo-wide.

No other Go-scoped constraint applies: this spec adds comments only, no
behavior, no SQL, no timestamps, no migrations.

### Prior related work

- `SPEC-079` (shipped) — built `scripts/inventory.sh`, established LD1's
  "derive it, guard it with `X3`" pattern this spec extends twice (the
  reservation row, the two questions rows), and closed
  `tag-ordering-projection` as the worked example of "a question whose own
  stated condition has already been decided by reality" — the same shape
  both closures in this spec follow.
- `SPEC-009` / `SPEC-010` (shipped) — origin of `DEC-009`, which answers
  `editor-template-format`.
- `SPEC-018` (shipped) — origin of `internal/aggregate.GroupForHighlights`
  and `DEC-014`, which answer `summary-grouping-heuristics`.

### Out of scope (for this spec specifically)

- The README restructure (STAGE-021's third, not-yet-scaffolded spec).
- STAGE-022 (lint, coverage, the `Entries:` envelope).
- Rewriting the 10 packages that already carry a package comment.
- The five still-open questions (see the triage table — each stays open on
  its own evidence, not by default).
- `scripts/status.sh`'s now-off-by-one casual decision count (Fork 1's known
  side effect).

## Notes for the Implementer

**Transcribe the six literals below verbatim.** Build should produce zero
prose of its own. Verify diffs the working tree against these literals.

---

### ① Five package doc comments (insert immediately before each `package` line)

**`cmd/brag/main.go`** — insert before `package main` (currently line 1):

```go
// Package main is the brag CLI entrypoint. It resolves the build version
// (goreleaser's ldflags, or the embedded module version for a
// `go install ./cmd/brag@latest` build — see resolveVersion), wires every
// internal/cli command onto the root cobra command, and maps the returned
// error to an exit code: ErrUser and storage.ErrDevProdMigrate exit 1
// (user-actionable), everything else exits 2 (internal fault). It holds no
// business logic of its own — see internal/cli for the commands and
// internal/storage for persistence.
```

**`internal/cli/root.go`** — insert before `package cli` (currently line 1):

```go
// Package cli implements every brag subcommand as a cobra.Command
// constructor — one file per verb (add.go, list.go, ...) — plus the
// scaffolding they share: ErrUser/UserErrorf for exit-code classification
// (errors.go), the injectable clock seam tests substitute (clock.go),
// atomic same-directory-rename config writes (atomicwrite.go), and the
// calendar-window flag parsing shared by impact, story and coverage
// (window.go). Its production code imports no SQL driver and no
// database/sql — the no-sql-in-cli-layer boundary, held today by
// convention and review, not an automated test (internal/mcpserver has
// one, TestNoSQLImport; internal/cli — the package the constraint's path
// glob actually covers — does not; see STAGE-022) — so every command
// reaches persistence only through internal/storage, keeping the CLI a
// thin shell a future frontend (TUI, API) could replace.
```

*(Corrected at the SPEC-080 punch-list build, 2026-08-20 — P1. The original
literal claimed the boundary was "enforced by the no-sql-in-cli-layer
constraint." Verify demonstrated empirically that nothing enforces it for
`internal/cli`: adding `_ "database/sql"` to `internal/cli/root.go` still
passes `go build`, `go vet`, `gofmt -l .`, `just test`, and `just
test-docs`. Fork chosen: narrow the claim (option a), not port
`TestNoSQLImport` (option b) — see "Punch List Resolution" below.)*

*(Amended at the re-verify return trip, 2026-08-20 — O1. `impact/story` →
`impact, story and coverage`: `internal/cli/coverage.go` calls both
`selectedWindow` (`coverage.go:59`) and `windowCutoff` (`coverage.go:68`),
exactly as `impact.go` and `story.go` do. Those three are the complete
caller set — `spark.go` names `selectedWindow` only in comments, to record
that it deliberately does not use it. No other clause in this literal
changed.)*

**`internal/config/config.go`** — insert before `package config` (currently
line 1):

```go
// Package config resolves the bragfile database path. ResolveDBPath applies
// DEC-003's fixed order — an explicit flag value, then the BRAGFILE_DB env
// var, then DefaultDBPath's ~/.bragfile/db.sqlite — expanding a leading
// `~/` and returning an absolute path. internal/cli's commands call it to
// resolve the --db flag; internal/storage's dev/prod migration guard
// (DEC-026) calls it independently to find the real database regardless of
// what path the caller passed. It has no dependency on internal/storage, so
// both can import it without a cycle.
```

**`internal/storage/store.go`** — insert before `package storage` (currently
line 1):

```go
// Package storage is the core of the storage layer — the only layer
// whose production code imports a SQL driver (modernc.org/sqlite, pure
// Go — no CGO, DEC-001). Test files elsewhere import it too; the
// storagetest test-helper subpackage exists so they need not, keeping
// raw SQL inside internal/storage. Store wraps *sql.DB and owns every
// persistence operation: Open applies pending embedded migrations
// (migrate.go) behind a pre-migration backup safety belt (backup.go,
// DEC-021) and a dev/prod migration guard (devguard.go, DEC-026); Entry
// and ListFilter (entry.go) are the query vocabulary; project.go holds
// the projects/locations operations over the 0004_add_projects.sql
// schema (DEC-017/019/020). Every other package's production code
// reaches the database only through a *Store — the no-sql-in-cli-layer
// boundary on internal/cli, held today by convention and review rather
// than an automated test (see STAGE-022) — which is what keeps commands
// testable and a future frontend feasible.
```

*(Corrected at the SPEC-080 punch-list build, 2026-08-20 — P1, P3, P4. Three
false claims in the original literal: (P1) "enforced by the
no-sql-in-cli-layer constraint" — not true, see the `internal/cli` literal
above; (P3) "the only package that imports a SQL driver" — false,
`internal/storage/storagetest/storagetest.go` (a distinct, non-test-file
package) also imports `modernc.org/sqlite`, so the claim is fixed to name
the storage **layer**, the repo's own established term; (P4) "project.go
adds the projects/locations schema" — `project.go` has zero DDL, the
schema is added by `internal/storage/migrations/0004_add_projects.sql`
(DEC-017 line 45 says so explicitly), so the claim is corrected to what
`project.go` actually holds: the operations over that schema.)*

*(Corrected again at the re-verify return trip, 2026-08-20 — R1. The P3 fix
above was still false: `package` → `layer` escaped the `storagetest`
counterexample without re-testing the narrowed claim, and `package cli`
imports `modernc.org/sqlite` directly in `story_test.go`, `wrapped_test.go`,
`coverage_test.go` and `impact_test.go`. Both uniqueness claims in this
literal are now scoped to **production code** — the leading "only layer"
claim and the later "Every other package reaches the database only through a
`*Store`", which was false for the same reason and by the same test (four
`internal/cli` test files call `sql.Open("sqlite", …)` and `db.Exec`
directly). The scoping matches `internal/cli/root.go`'s own qualifier in
literal ① above, so the two comments now agree. See "Re-verify Punch List
Resolution" below for the rejected alternatives.)*

**`internal/story/bundle.go`** — insert before `package story` (currently
line 1):

```go
// Package story builds and renders `brag story` (SPEC-049): audience-shaped
// bundles of Threads and Beats over a window of entries. BuildThreads and
// BuildThroughline (thread.go) project []storage.Entry into the threading
// shape a Profile selects; ToStoryMarkdown/ToStoryJSON (bundle.go, this
// file) render the result per DEC-014's envelope. Profile (profile.go) is
// DATA, not a Go enum — bundled defaults live under profiles/*.yaml and
// directives/*.md (embed.go, DEC-029 choice 2), with an optional
// ~/.bragfile/story-profiles/ user override shadowing a bundled name by
// name. The package reads storage.Entry values but never a *storage.Store
// or a SQL driver, and reads no clock itself — the CLI resolves the window,
// the directive text, and "now", and passes them in via StoryOptions.
```

---

### ② `decisions/DEC-041-reserved-brag-project-goto-multi-location-primary-policy.md` (new file, 70 lines)

```markdown
---
insight:
  id: DEC-041
  type: reservation
  audience:
    - developer
    - agent

agent:
  id: claude-opus-5
  session_id: null

project:
  id: PROJ-007
repo:
  id: bragfile

created_at: 2026-08-19
supersedes: null
superseded_by: null

tags:
  - reservation
  - backlog
---

# DEC-041: reserved — `brag project goto` / multi-location primary policy

## This is not a decision

`DEC-041` is reserved, not written. It exists only so `decisions/` has no
unexplained gap between [`DEC-040`](DEC-040-distribution-binary-formula-over-cask.md)
and [`DEC-042`](DEC-042-mcp-time-window-filter-parity.md) — there is no choice
recorded here to read, weigh, or cite.

The number was set aside on 2026-07-16 for `brag project goto` (SPEC-070), a
navigation-ergonomics feature that was drafted, then deferred — not built, not
rejected — on 2026-08-13.

## What DEC-041 would decide, if `brag project goto` is ever built

A project can have many locations (`project_locations`, one-to-many —
[`DEC-017`](DEC-017-entries-project-relationship.md),
[`DEC-019`](DEC-019-project-here-resolution-policy.md),
[`DEC-020`](DEC-020-project-location-editing-semantics.md)), so a `goto`
command has to pick one as "the" directory to jump to. The deferred draft
assumed **first-registered = primary** — a real choice, not an obvious one,
which is why it needed a decision record at all rather than a one-line
default.

## Where the rest of the design lives

The full navigation design — why a subprocess can't `cd` its parent shell,
the two-part `goto` (resolver) + `shell-init` (shell-function wrapper) shape,
and the trigger for revisiting it — is preserved in
[`../projects/PROJ-001-mvp/backlog.md`](../projects/PROJ-001-mvp/backlog.md),
under "`brag project goto` + `brag shell-init` — jump to a project's
directory."

## Disposition

- **If `brag project goto` is ever built**, it takes this number: the
  decision gets written into *this* file, replacing this notice, rather than
  claiming a new one.
- **If the backlog item is ever deleted outright** instead, this file is
  updated to say so — so the gap stays explained rather than becoming a
  second mystery.
- **Do not reuse `DEC-041` for anything else.** The backlog entry says so
  explicitly, and reusing a reserved number would leave a decision record and
  its own reservation notice referring to two different things.
```

---

### ③ `scripts/inventory.sh` — diff

```diff
@@ -44,9 +44,14 @@ confidences() {
     grep -h '^  confidence:' decisions/DEC-*.md | sed 's/#.*//' | awk '{print $2+0}'
 }

-decs=$(n "$(ls decisions/DEC-*.md 2>/dev/null | wc -l)")
+decs=$(n "$(grep -l '^  type: decision' decisions/DEC-*.md 2>/dev/null | wc -l)")
 decs_superseded=$(n "$(grep -l '^superseded_by: DEC-' decisions/DEC-*.md 2>/dev/null | wc -l)")
 decs_amended=$(n "$(grep -l '^## Amendment' decisions/DEC-*.md 2>/dev/null | wc -l)")
+# A tombstone (insight.type: reservation, e.g. DEC-041) holds a number
+# without deciding anything — SPEC-080. Counted separately so it does not
+# silently inflate "Decision records", and so its own existence is a derived
+# number too, not a fact only readable in prose.
+decs_reserved=$(n "$(grep -l '^  type: reservation' decisions/DEC-*.md 2>/dev/null | wc -l)")
 conf_min=$(confidences | sort -g | head -1)
 conf_max=$(confidences | sort -g | tail -1)
 conf_certain=$(n "$(confidences | awk '$1 >= 1.0' | wc -l)")
@@ -62,12 +67,21 @@ wseries=$(n "$(grep -cE '^# W[0-9]+ ' scripts/test-docs.sh)")
 docasserts=$(n "$(grep -oE '(^|[[:space:]])(ok|fail|skip|assert_[a-z_]+) "[A-Za-z0-9][A-Za-z0-9._-]*"' \
     scripts/test-docs.sh | grep -oE '"[A-Za-z0-9][A-Za-z0-9._-]*"' | tr -d '"' | sort -u | wc -l)")

+# Open-questions hygiene (SPEC-080). STAGE-021's own count was wrong three
+# times in three days — once from a plain `grep -c 'status: open'` that
+# matched this file's OWN header comment (`#   - status: open |
+# investigating | answered`). Anchored to the exact 4-space indent real
+# entries use, which the `#`-prefixed header comment can never match.
+questions_total=$(n "$(grep -cE '^  - id: ' guidance/questions.yaml 2>/dev/null || true)")
+questions_open=$(n "$(grep -cE '^    status: open$' guidance/questions.yaml 2>/dev/null || true)")
+
 cat <<EOF
 | What | Value | Where it lives |
 |---|---:|---|
-| Decision records | ${decs} | \`decisions/DEC-*.md\` |
+| Decision records | ${decs} | \`decisions/DEC-*.md\` (\`insight.type: decision\`) |
 | …of those, superseded by a later record | ${decs_superseded} | \`superseded_by:\` in the front-matter |
 | …of those, carrying an explicit \`## Amendment\` section | ${decs_amended} | \`decisions/DEC-*.md\` |
+| Decision numbers reserved, not yet decided | ${decs_reserved} | \`decisions/DEC-*.md\` (\`insight.type: reservation\`) |
 | Lowest confidence value on a decision record | ${conf_min} | \`insight.confidence\` in the front-matter |
 | Highest confidence value on a decision record | ${conf_max} | \`insight.confidence\` in the front-matter |
 | Decision records claiming confidence 1.0 | ${conf_certain} | \`insight.confidence\` in the front-matter |
@@ -80,5 +94,7 @@ cat <<EOF
 | Go test functions | ${testfuncs} | \`func Test*\` in \`*_test.go\` |
 | Documentation assertions (distinct ids) | ${docasserts} | \`scripts/test-docs.sh\`, run by \`just test-docs\` |
 | …of those, replacing a manual release-checklist item | ${wseries} | the \`W\`-series in \`scripts/test-docs.sh\` |
+| Questions tracked in guidance/questions.yaml | ${questions_total} | \`guidance/questions.yaml\` |
+| …of those, still open | ${questions_open} | \`status: open\` in the same file |
 | Benchmarks | ${benchmarks} | none exist — see "What this does not measure" |
 EOF
```

---

### ④ `scripts/test-docs.sh` — new Group Y (insert after Group X, before `# ===== finalise =====`)

```bash
# ===== Group Y — godoc pass + legibility repairs (SPEC-080) =====

# Y1 — the five packages the godoc pass names now carry a package doc
# comment: a `//` comment line immediately preceding `package X`, the exact
# adjacency rule go/doc (and `go doc`) key off. Framing's "7 missing" was
# wrong — internal/export and internal/mcpserver already had one; the real
# gap, re-measured at design, was these five.
y1_missing=""
for y1_target in \
    "cmd/brag/main.go" \
    "internal/cli/root.go" \
    "internal/config/config.go" \
    "internal/storage/store.go" \
    "internal/story/bundle.go"
do
    if [ ! -f "$y1_target" ]; then
        y1_missing="$y1_missing $y1_target(missing-file)"
        continue
    fi
    y1_pkgline=$(grep -n '^package ' "$y1_target" | head -1 | cut -d: -f1)
    if [ -z "$y1_pkgline" ]; then
        y1_missing="$y1_missing $y1_target(no-package-decl)"
        continue
    fi
    y1_prev=$((y1_pkgline - 1))
    if [ "$y1_prev" -lt 1 ] || ! sed -n "${y1_prev}p" "$y1_target" | grep -q '^//'; then
        y1_missing="$y1_missing $y1_target"
    fi
done
if [ -z "$y1_missing" ]; then
    ok "Y1"
else
    fail "Y1" "missing a doc comment immediately before 'package':$y1_missing"
fi

# Y2 — the DEC-041 gap is explained IN decisions/, not only in the backlog:
# a tombstone file exists and is explicitly marked `insight.type:
# reservation`, not `decision` — the marker the Decision records count
# (Y3/inventory.sh) relies on to NOT count it as a decision.
y2_file=$(ls decisions/DEC-041-*.md 2>/dev/null | head -1)
if [ -z "$y2_file" ]; then
    fail "Y2" "no decisions/DEC-041-*.md tombstone file"
elif grep -q '^  type: reservation' "$y2_file"; then
    ok "Y2"
else
    fail "Y2" "$y2_file exists but is missing '  type: reservation' in its front-matter"
fi

# Y3 — the tombstone does not inflate the Decision records count, and its
# own reservation is counted separately. Pins the SEMANTIC values, not just
# X3-style script-vs-page self-consistency, which would happily pass even if
# both sides agreed on a wrong number (the failure mode this pin exists to
# catch: someone edits the type filter in inventory.sh and it silently starts
# counting the tombstone as a decision — script and page would still agree,
# just agree on 46).
if [ ! -x scripts/inventory.sh ]; then
    fail "Y3" "scripts/inventory.sh is missing or not executable"
else
    y3_out=$(./scripts/inventory.sh)
    y3_bad=""
    printf '%s\n' "$y3_out" | grep -F -q 'Decision records | 45 |' || y3_bad="$y3_bad decision-records!=45"
    printf '%s\n' "$y3_out" | grep -F -q 'Decision numbers reserved, not yet decided | 1 |' || y3_bad="$y3_bad reserved-decisions!=1"
    if [ -z "$y3_bad" ]; then
        ok "Y3"
    else
        fail "Y3" "inventory.sh row value(s) wrong:$y3_bad"
    fi
fi

# Y4 — the open-questions count is DERIVED (inventory.sh), not restated: the
# fix for the STAGE-021 line that has been wrong three times in three days
# (8 of 18 wrong total; 9 of 18 wrong on both halves, from a grep that
# matched the file's own header comment; 7 of 17 correct until a merge
# landed a new question). Pins the semantic values, same rationale as Y3.
if [ ! -x scripts/inventory.sh ]; then
    fail "Y4" "scripts/inventory.sh is missing or not executable"
else
    y4_out=$(./scripts/inventory.sh)
    y4_bad=""
    printf '%s\n' "$y4_out" | grep -F -q 'Questions tracked in guidance/questions.yaml | 18 |' || y4_bad="$y4_bad questions-total!=18"
    printf '%s\n' "$y4_out" | grep -F -q 'of those, still open | 6 |' || y4_bad="$y4_bad questions-open!=6"
    if [ -z "$y4_bad" ]; then
        ok "Y4"
    else
        fail "Y4" "inventory.sh row value(s) wrong:$y4_bad"
    fi
fi

# Y5 — the two questions answered in practice by this spec (editor-template-
# format by DEC-009; summary-grouping-heuristics by SPEC-018/DEC-014 +
# aggregate.GroupForHighlights) are marked closed in the register, not left
# stale the way both sat since 2026-04-19 despite the answer already
# existing.
y5_missing=""
for y5_id in "editor-template-format" "summary-grouping-heuristics"; do
    y5_status=$(awk -v want="  - id: $y5_id" '
        $0 == want { f=1; next }
        f && /^  - id: / { exit }
        f && /^    status: / { print; exit }
    ' guidance/questions.yaml)
    if [ "$y5_status" != "    status: answered" ]; then
        y5_missing="$y5_missing $y5_id"
    fi
done
if [ -z "$y5_missing" ]; then
    ok "Y5"
else
    fail "Y5" "guidance/questions.yaml entries not marked 'status: answered':$y5_missing"
fi

# ===== finalise =====
```

---

### ⑤ `guidance/questions.yaml` — diff

```diff
@@ -206,14 +206,43 @@ questions:
       `ulid TEXT UNIQUE` column, backfill, switch user-facing CLI to
       accept either form. No data loss.

+      STILL OPEN AT SPEC-080 HYGIENE PASS (2026-08-19): the trigger has NOT
+      fired. `docs/research/duckdb-federation-spike.md` (a pre-framing
+      research spike, explicitly "not the repo; only this report is
+      durable") found "bragfile has no global entry identity" and recommended
+      a file/export-level identity stamp over a per-entry one — but the spike
+      was never pulled into a shipped feature; no cross-machine sync or
+      export exists today. That report is prior art for WHEN this is
+      actioned, not evidence the trigger condition has been met.
+
   - id: editor-template-format
     question: "What exact format does the editor-launch markdown buffer use — YAML front-matter plus a body, or a simpler key-value header?"
     priority: medium
-    status: open
+    status: answered
     raised_by: claude
     raised_at: 2026-04-19
+    answered_at: 2026-08-19
     assigned_to: null
+    answered_by: DEC-009 (SPEC-009 / SPEC-010)
     notes: |
+      ANSWERED IN PRACTICE, FOUND STALE AT SPEC-080 DESIGN. This question was
+      raised 2026-04-19; DEC-009 settled it the very next day, 2026-04-20, and
+      the entry was simply never closed — the same shape as
+      `tag-ordering-projection`, a question whose own condition had already
+      been decided by reality.
+
+      NOT YAML front-matter — Option E: RFC822-style headers parseable via
+      stdlib `net/textproto.Reader.ReadMIMEHeader`, a blank line, then a
+      free-form markdown body as the entry's Description. Fixed header order
+      (Title, Tags, Project, Type, Impact); empty-valued fields omitted from
+      the render. Confirmed still exactly this in the shipped code
+      (`internal/editor/editor.go`'s `Render`/`EmptyTemplate`/`Parse`, and the
+      package doc comment SPEC-080 added, which states the format inline).
+      YAML/TOML front-matter were rejected in DEC-009 specifically to avoid a
+      new top-level dependency (`no-new-top-level-deps-without-decision`).
+
+      --- original entry, retained for provenance ---
+
       Arrives in STAGE-002 (SPEC for `brag edit` / `brag add` no-args).
       Leading candidate: YAML front-matter for optional fields (tags,
       project, type, impact), then `# <title>` and the description as
@@ -223,11 +252,28 @@ questions:
   - id: summary-grouping-heuristics
     question: "How does `brag summary --range week` decide on grouping order (project first vs type first vs tag first) when an entry has all three?"
     priority: low
-    status: open
+    status: answered
     raised_by: claude
     raised_at: 2026-04-19
+    answered_at: 2026-08-19
     assigned_to: null
+    answered_by: SPEC-018 (DEC-014) / aggregate.GroupForHighlights
     notes: |
+      ANSWERED IN PRACTICE, FOUND STALE AT SPEC-080 DESIGN. Same shape as
+      `editor-template-format`: the "leaning toward" guess below is exactly
+      what shipped, and the entry was never closed.
+
+      PROJECT is the only grouping axis for `## Highlights`. Confirmed in
+      `internal/aggregate/aggregate.go`'s `GroupForHighlights`: entries bucket
+      by `Project` (alpha-ASC, `(no project)` forced last per DEC-013/DEC-014's
+      count-ordering rule), and within each project bucket sort chrono-ASC by
+      `CreatedAt` with `ID` as tie-break. `Type` is never a grouping key for
+      Highlights — it gets its own separate flat `**By type**` count list,
+      never nested with project. Tags are not a grouping axis at all. DEC-014
+      §3 locks the by-type/by-project convention this implements.
+
+      --- original entry, retained for provenance ---
+
       Lands in STAGE-003. Leaning toward: always group by `project`
       first, then `type` second, with tags shown inline per entry.
       Make this choice when writing the summary spec — not now.
```

---

### ⑥ `docs/engineering-practices.md` — diff

```diff
@@ -26,9 +26,10 @@ The section that says the most about how this repository is run is
 <!-- inventory:begin — generated by scripts/inventory.sh (`just inventory`); pinned by test-docs X3; do not hand-edit -->
 | What | Value | Where it lives |
 |---|---:|---|
-| Decision records | 45 | `decisions/DEC-*.md` |
+| Decision records | 45 | `decisions/DEC-*.md` (`insight.type: decision`) |
 | …of those, superseded by a later record | 1 | `superseded_by:` in the front-matter |
 | …of those, carrying an explicit `## Amendment` section | 1 | `decisions/DEC-*.md` |
+| Decision numbers reserved, not yet decided | 1 | `decisions/DEC-*.md` (`insight.type: reservation`) |
 | Lowest confidence value on a decision record | 0.65 | `insight.confidence` in the front-matter |
 | Highest confidence value on a decision record | 0.95 | `insight.confidence` in the front-matter |
 | Decision records claiming confidence 1.0 | 0 | `insight.confidence` in the front-matter |
@@ -39,8 +40,10 @@ The section that says the most about how this repository is run is
 | Go source files | 69 | `internal/`, `cmd/` |
 | Go test files | 78 | `internal/`, `cmd/` |
 | Go test functions | 812 | `func Test*` in `*_test.go` |
-| Documentation assertions (distinct ids) | 171 | `scripts/test-docs.sh`, run by `just test-docs` |
+| Documentation assertions (distinct ids) | 176 | `scripts/test-docs.sh`, run by `just test-docs` |
 | …of those, replacing a manual release-checklist item | 6 | the `W`-series in `scripts/test-docs.sh` |
+| Questions tracked in guidance/questions.yaml | 18 | `guidance/questions.yaml` |
+| …of those, still open | 6 | `status: open` in the same file |
 | Benchmarks | 0 | none exist — see "What this does not measure" |
 <!-- inventory:end -->

@@ -53,9 +56,14 @@ same trap already documented at assertion `E2`.
 ## Decisions

 [`../decisions/`](../decisions/) holds one file per non-obvious choice. Ids are
-repo-global and never reused, and `DEC-041` is reserved rather than lost — the
-reservation is recorded at
+repo-global and never reused, and `DEC-041` is reserved rather than lost: a
+tombstone file,
+[`DEC-041`](../decisions/DEC-041-reserved-brag-project-goto-multi-location-primary-policy.md),
+sits in the directory between `DEC-040` and `DEC-042` naming the deferred
+feature that holds the number and pointing at the full design in
 [`../projects/PROJ-001-mvp/backlog.md`](../projects/PROJ-001-mvp/backlog.md).
+It is marked `insight.type: reservation`, not `decision`, so it does not
+count itself in the row above.

 - **Each record carries an honest confidence value.** `insight.confidence` is
   part of every record's front-matter, and no record claims certainty — see the
```

*(Corrected at the SPEC-080 punch-list build, 2026-08-20 — P5. The original
literal's added line read "so it does not count itself in the row below,"
directionally wrong: the inventory table this sentence refers to sits
**above** the `## Decisions` section, not below, and the very next bullet
in the same section correctly says "see the lowest, highest and 1.0 rows
above." One word changed: `below` → `above`.)*

**Note on the cached derived number in this diff (per the framing prompt's
trap-2 warning).** The `Documentation assertions` value (176) and both new
questions-row values (18, 6) are **derived, not authored** — they were
produced by actually running `./scripts/inventory.sh` against the fully
staged tree at design time (§12(b) row 14), not typed by hand. If build lands
this diff after any *other* change has landed on `main` that alters any row
in the inventory table (a new spec archived, a new DEC filed, a new Go test
function), **the derivation outranks this cached diff**: build must re-run
`just inventory` and re-paste the live output before declaring `X3` green,
exactly as SPEC-079's own ship reflection found necessary when `PROJ-008`/
`PROJ-009` landed between its design and build sessions.

## Confidence

**0.90.** Both forks are settled with rejected alternatives recorded and are
backed by evidence checked directly against the repo, not asserted from
memory: the type-filter approach was verified against all 45 existing
records (zero exceptions) and against `status.sh`'s tolerant handling of a
missing `confidence:` field; the anchored open-question grep was verified
against the *current* register before any edits, reproducing the number
STAGE-021 already had right. Every literal was run through its real target
tool (gofmt, go vet, go build, go test, the full test-docs.sh harness) with
all six changes staged together, and all pass. The one soft spot —
`status.sh`'s casual decision count drifting to an informational-only
off-by-one — is named explicitly rather than silently accepted, and is below
the threshold of a design decision (no test, no published claim, no guard
depends on it). No new item is filed in `guidance/questions.yaml`: nothing
here is below the §14 0.7 threshold, and the five questions this pass leaves
open already carry their own (unresolved, correctly so) entries.

## Gates

```
just test
just test-docs
gofmt -l .
go vet ./...
```

All four re-verified at design with every literal staged (§12(b) rows 4–6,
14); all pass. Build should see the identical result after transcribing the
six literals — nothing here is order-dependent or timing-sensitive.

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `build/spec-080-godoc-and-legibility-repairs`.
- **PR (if applicable):** none opened (per the invoking instruction — do not
  push, do not open a PR).
- **All acceptance criteria met?** yes. `just test-docs` is green at **176
  distinct assertion ids** (was 171; delta is exactly `Y1`–`Y5`, confirmed by
  `comm` against the pre-build id set — nothing added beyond the five, nothing
  lost). `just test`, `gofmt -l .` and `go vet ./...` are unaffected (all
  clean). `go build ./...` exits 0. `go doc` on all five named packages
  renders the new comment in full (verified individually, not just via
  `gofmt`'s adjacency check).
- **Order followed:** literal ③ (`scripts/inventory.sh`) and literal ②
  (the DEC-041 tombstone) landed first, per the invoking instruction.
  `./scripts/inventory.sh` was then run against that intermediate state —
  `Decision records | 45 |` and `Decision numbers reserved... | 1 |` were
  already correct at that point, and `Documentation assertions` /
  `Questions...` rows still read the pre-build values (171; 18/8) because
  literals ④ and ⑤ hadn't landed yet, confirming the script tracks live
  state rather than anything cached. After ④ and ⑤ landed, `inventory.sh`
  was re-run once more before touching literal ⑥, producing **176 / 18 / 6**
  — the exact values literal ⑥ predicted. Literal ⑥ was then applied
  unmodified; no reconciliation was needed because the script's live output
  and the spec's cached diff agreed exactly.
- **New decisions emitted:**
  - none. Every choice was settled at design (LD1–LD4); build produced no
    non-trivial decision of its own.
- **Deviations from spec:**
  - none. All six literals transcribed byte-for-byte — verified by `git diff`
    against each literal's exact text for ①, ③, ⑥, and by direct line-count
    (70) and content comparison for ②, ④, ⑤.
  - `cycle:` was edited to `build` by hand rather than via
    `just advance-cycle`, per the build instruction.
- **Follow-up work identified:**
  - none beyond what the spec already named as deliberately deferred
    (`status.sh`'s off-by-one 46, the README restructure, STAGE-022 scope).

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing. This is the second consecutive spec in this stage (after
   SPEC-079) to apply the literal-artifact-as-spec contract, and the
   "re-run the derivation first" ordering instruction (itself the
   generalised lesson from SPEC-079's own `Projects | 7 → 9` drift) meant
   there was no ambiguity about precedence: land ③/②, re-derive, reconcile,
   *then* judge everything else. Unlike SPEC-079, no literal had gone stale
   between design and build — `./scripts/inventory.sh`'s live output matched
   literal ⑥ exactly on the first run, so there was nothing to reconcile.
2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. `no-sql-in-cli-layer` was restated as prose inside the new
   `internal/cli` package comment (literal ①) and required no code change to
   satisfy — the comment is accurate against the current import set, which
   `go vet`/`go build` passing confirms indirectly (an added SQL import would
   have broken the `no-sql-in-cli-layer`-guarded build, not just made the
   comment stale). Everything else was comment-only, so no Go-scoped
   constraint had any surface to violate.

   **CORRECTED AT THE SPEC-080 PUNCH-LIST BUILD (2026-08-20), P2.** The
   parenthetical above is false, and the record of having believed it is
   kept rather than deleted. Verify falsified it directly: inserting
   `_ "database/sql"` into the production `internal/cli/root.go` still
   passes `go build`, `go vet`, `gofmt -l .`, `just test`, and
   `just test-docs` — nothing catches it. `no-sql-in-cli-layer` is held
   for `internal/cli` by convention and review, not by any automated
   guard; the one guard that exists in the repo
   (`internal/mcpserver/import_audit_test.go`'s `TestNoSQLImport`) was
   written for `internal/mcpserver`, whose own comment says the
   constraint's path glob covers `internal/cli/**` only — the package the
   rule was actually written for has no test. The package comment (P1)
   was corrected to say this honestly rather than claim enforcement.
   Closing the gap itself (a ported audit test, or a `depguard`
   golangci-lint rule scoped to `internal/cli/**`) is out of this spec's
   scope and is now a noted follow-up in STAGE-022 — see this spec's
   "Punch List Resolution" section below.
3. **If you did this task again, what would you do differently?**
   — Nothing procedurally. One minor observation: literal ④'s Y1 check reads
   the line immediately preceding each `package` declaration via
   `sed -n "${y1_prev}p"`, which is exactly the adjacency rule `go doc` uses —
   worth calling out that this build additionally ran `go doc` on all five
   packages by hand (a stronger, tool-level confirmation than the shell
   assertion alone gives), per the invoking instruction's explicit warning
   that a blank-line-separated comment "compiles fine and documents nothing."
   Both checks agreed on all five files; a future spec doing the same kind of
   pass could fold the `go doc` check into `test-docs.sh` itself rather than
   relying on Y1's line-adjacency proxy plus a manual `go doc` pass — Y1
   already gets this exactly right, so this is a note, not a defect.

---

## Verify Findings

*Filled in during the **verify** cycle (fresh session, per AGENTS.md §6).*

**Verdict: ⚠ PUNCH LIST.** The mechanical half is clean — all five gates
green, all six literals transcribed byte-for-byte (verified by diff, not by
assertion), the id delta is exactly `Y1`–`Y5` with nothing lost, and all ten
untouched inventory numbers independently recompute correctly. Every one of
`Y1`–`Y5` was mutation-tested and caught its mutant, each mutant hash-verified
as a real change. The punch list is entirely in the **judgement** half: three
factual claims inside the new godoc prose do not survive checking, and the
build reflection asserts a guard that does not exist.

Nothing below was fixed in place — each spans both the spec's embedded literal
and the shipped artifact, so fixing one without the other breaks the
literal-artifact contract.

### P1 — "enforced by the `no-sql-in-cli-layer` constraint" is not true (blocking)

`internal/cli/root.go` and `internal/storage/store.go` both state the SQL
boundary is *"enforced by the no-sql-in-cli-layer constraint."* Nothing
enforces it. Verified empirically: inserting `_ "database/sql"` into the
**production** file `internal/cli/root.go` — six lines below the comment
denying it — passes `go build`, `go vet`, `gofmt -l .`, `just test`, and
`just test-docs`. No gate caught it.

The repo already contains the guard this needs:
`internal/mcpserver/import_audit_test.go`'s `TestNoSQLImport`, whose own
comment reads *"The constraint's path glob covers `internal/cli/**` only;
this test covers the gap for the new package."* The package the constraint
was actually written for is the one without the test.

**Fix:** either soften the wording in both literals to what the repo's own
prose already says elsewhere (`internal/storage/storagetest`: *"without
violating"*; `impact_test.go`: *"the production CLI layer stays SQL-free"*),
or port `TestNoSQLImport` to `internal/cli` and make the word "enforced"
true. The second is the better outcome and is small, but it is new test
surface this spec did not scope.

### P2 — build reflection Q2 is factually wrong (blocking)

Q2 answers that *"an added SQL import would have broken the
`no-sql-in-cli-layer`-guarded build, not just made the comment stale."* P1's
mutation falsifies this directly. The reflection is honest in intent, but a
future spec reading it will assume a guard exists that does not. Correct the
answer to record that the constraint is convention-only for `internal/cli`.

### P3 — "the only package that imports a SQL driver" is false (minor)

`internal/storage/store.go`'s comment opens *"Package storage is the only
package that imports a SQL driver."* `internal/storage/storagetest` is a
distinct package whose **non-test** file `storagetest.go` imports
`_ "modernc.org/sqlite"`. The repo's own established phrasing is *the storage
**layer***, not the package — `storagetest`'s own comment says *"Living under
`internal/storage/` keeps the `database/sql` dependency inside the storage
layer."* Changing "package" to "layer" makes the sentence true.

### P4 — "project.go adds the projects/locations schema" is contradicted by its own citation (minor)

`internal/storage/store.go`'s comment says *"project.go adds the
projects/locations schema (DEC-017/019/020)."* `project.go` contains zero
`CREATE`/`ALTER`/`DROP` statements — it holds the `Store` methods
(`CreateProject`, `AddLocation`, `ProjectForPath`, …). The schema is added by
`internal/storage/migrations/0004_add_projects.sql`, which **DEC-017 — the
first DEC cited — states explicitly** (line 45: *"The `0004_add_projects.sql`
migration adds two tables"*). Suggested: *"project.go holds the
projects/locations operations over the `0004_add_projects.sql` schema."*

### P5 — "the row below" points the wrong way (minor)

`docs/engineering-practices.md` (literal ⑥, spec line 1021): *"so it does not
count itself in the row below."* The inventory table ends at
`<!-- inventory:end -->` **above** that prose; the next bullet in the same
section correctly says *"see the lowest, highest and 1.0 rows **above**."*
One word: `below` → `above`.

### Observations — not blocking, no action required in this spec

- **The `insight.type` seam is the right call, with one new silent-loss
  mode.** It reuses a field every one of the 45 records already carries, the
  distinction is semantic rather than positional, and `Y3` pins it
  semantically. Confirmed by scratch file: `type: decision` → 46/1,
  `type: reservation` → 45/2. But `type: bogus` and *no type field* both
  → 45/1 — counted by **neither** row. The old bare glob could not lose a
  file; the filter can. Nothing asserts totality
  (`decs + decs_reserved == ls decisions/DEC-*.md | wc -l`), and
  `reservation` is absent from `decisions/_template.md`'s documented enum
  (`decision | analysis | recommendation | observation`). A future spec
  touching either should close both gaps.
- **`Y3` earns its place, proven.** Mutating `inventory.sh` *and* the page
  together so both agreed on 46 left `X3` **passing** and `Y3` the only
  failure — exactly the blind spot the design predicted.
- **`Y1` catches what no Go gate can.** A blank line between the comment and
  `package config` passes `go build`, `go vet`, and `gofmt -l .` while
  `go doc ./internal/config` renders **no package comment at all**. Only `Y1`
  fails.
- **`status.sh` behaves exactly as predicted** — 46 by filename glob against
  the guarded 45. Confirmed, not touched.
- **176 vs 177.** `test-docs.sh` prints 177 `OK:` lines because `S3` is
  emitted twice; distinct ids are 176, matching both the script and the page.
  The duplicate is unchanged from `HEAD~1` and the row is honestly labelled
  *"distinct ids"*. Pre-existing, correct as defined.
- **`internal/config`, nit only.** *"internal/storage's dev/prod migration
  guard (DEC-026) calls it independently"* — the guard calls
  `config.DefaultDBPath` (`devguard.go:39`), not `ResolveDBPath`, which is
  what "it" refers to in the preceding clause. Substantively right, referent
  loose.

### Judgement on each of the five comments

| Package | Verdict |
|---|---|
| `cmd/brag` | **Sound.** Every claim verified — `resolveVersion`, all 21 `AddCommand` calls (literally *every* `internal/cli` constructor), and the exact `ErrUser`/`ErrDevProdMigrate` → 1, else 2 mapping. |
| `internal/cli` | **Sound but for P1.** File-to-concern map (`errors.go`, `clock.go`, `atomicwrite.go`, `window.go`) all correct; the SQL claim is true of production code but "enforced" is not. |
| `internal/config` | **Sound.** Resolution order, `~/` expansion, absolute-path return, and the no-cycle claim all verified against the code. Only the P-list nit above. |
| `internal/storage` | **Weakest of the five** — two false claims (P3, P4). The `Open` sequence (devguard → backup → apply) and `entry.go` vocabulary are exactly right. |
| `internal/story` | **Strongest of the five.** Every claim verified, including the non-obvious one: `DEC-014`'s envelope legitimately governs `story` because **DEC-029 choice 5** explicitly *"EXTENDS DEC-014's envelope."* Purity claim (no `*storage.Store`, no driver, no clock) confirmed by grep. |

Voice and depth sit comfortably inside the range of the ten pre-existing
comments — denser than `export`/`aggregate`/`mcpserver`, comparable to
`editor`/`memory`, less discursive than `timewindow`/`ftsquery`. No new
comment is markedly thinner than the existing set.

### Question closures — both hold

- **`editor-template-format` → DEC-009: holds, cleanly.** The register asked a
  binary ("YAML front-matter … or a simpler key-value header?"). DEC-009's
  **Option E (chosen)** is `net/textproto` header + markdown body; **Option A
  (YAML front-matter) is explicitly rejected**, and its stated reason —
  *"Requires `gopkg.in/yaml.v3` (a new top-level dep needing its own DEC)"* —
  is exactly the rationale the closure note cites. `EmptyTemplate()` returns
  `"Title: \nTags: \nProject: \nType: \nImpact: \n\n"`, matching the claimed
  fixed header order verbatim.
- **`summary-grouping-heuristics` → `aggregate.GroupForHighlights`: holds.**
  Read independently: buckets on `e.Project` only; `NoProjectKey` forced last;
  remaining projects ordered `out[i].Project < out[j].Project` (alpha-ASC);
  within a bucket, `CreatedAt` ASC with `ID` tie-break. `Type` is never a
  grouping key — `internal/export/summary.go` renders it as a separate flat
  `**By type**` list. Tags are not an axis. Exactly the closure's claim, and
  exactly the "leaning toward" guess the 2026-04-19 entry recorded.
  *One imprecision:* the note's parenthetical *"(alpha-ASC, `(no project)`
  forced last per DEC-013/DEC-014's count-ordering rule)"* — that rule is
  **count-DESC with alpha-ASC tiebreak**; only the `(no project)`-last clause
  comes from it. The behavioural description is right; the attribution reads
  as covering both properties.

**The six that stayed open all still have live triggers.** Confirmed:
`## Amendment` count is **1** against a 4+ resolve condition and PROJ-007 is
still `status: active`; `brag spark --month` exists and is wired
(`spark.go:43`, help text confirming the rolling-28-day reading the question
flags); `docs/research/duckdb-federation-spike.md` exists and is a spike, not
a shipped feature. The remaining three turn on a user report or a subjective
dogfooding read. No answer was invented to shrink the register.

### Mutation results — all five caught, every mutant hash-verified

| Mutant | Change proven by | Result |
|---|---|---|
| `Y1` blank line before `package config` | sha256 `14589d46…` → `6983aab8…` | `FAIL: Y1` — while `go build`/`vet`/`gofmt` all pass |
| `Y2` tombstone retyped `reservation`→`decision` | `30d3c43c…` → `830b5cad…` | `FAIL: Y2`, `Y3`, `X3` |
| `Y3` filter reverted to bare glob | `e2db9558…` → `331ae2a7…` | `FAIL: Y3`, `X3` |
| `Y3` **script + page both** set to 46 | both hashes changed | `FAIL: Y3` only — `X3` passes |
| `Y4` open-count grep de-anchored (the historical bug) | `e2db9558…` → `3e0ce8e9…` | `FAIL: Y4`, `X3` |
| `Y5` `editor-template-format` → `status: open` | `ed611ad6…` → `45fc1976…` | `FAIL: Y5`, `Y4`, `X3` |

`Y4`/`Y5` cross-check as designed. Every mutation restored; tree verified
clean after each.

**Id delta:** `61 → 66` distinct `ok`/`fail` labels, added set exactly
`{Y1,Y2,Y3,Y4,Y5}`, removed set **empty** (`comm` against `HEAD~1`).
Page `171 → 176`, matching the script's static derivation.

### Re-verify

Per the stage's own history (SPEC-075's punch-list delta found two more
defects; SPEC-078 skipped the re-verify and shipped an assertion pinning the
wrong proposition), **the punch-list delta deserves its own re-verify pass
before this spec advances.** P1 in particular can be closed two ways with
materially different scope.

---

## Punch List Resolution (build return trip, 2026-08-20)

*A fresh build session, per AGENTS.md §6, closing the five findings above.
Each fix lands in both the shipped artifact and this spec's literal, per the
invoking instruction — the literal is the source of truth for a future
rebuild.*

### P1 — fork chosen: (a) narrow the claim, not (b) port `TestNoSQLImport`

**Chosen: (a).** Both package comments (`internal/cli/root.go`,
`internal/storage/store.go`, literal ①) now say the boundary is *"held
today by convention and review, not an automated test"* and name the gap
explicitly — `internal/cli` is the package the constraint's path glob
actually covers, and it is the one without a test; `internal/mcpserver`
has one it didn't strictly need to have written first.

**Rejected alternative: (b), port `TestNoSQLImport` to `internal/cli` —
REJECTED for this spec.** Verified the cost the invoking instruction
warned about is real, not hypothetical: `internal/cli/coverage_test.go`,
`impact_test.go`, `project_test.go`, and `wrapped_test.go` all import
`database/sql` today (confirmed by direct grep), so a naive port of
`TestNoSQLImport`'s walk-every-`.go`-file shape fails immediately against
production's own test suite. Closing it properly requires first deciding
whether the constraint covers test files — a real scope question, not a
mechanical one — and separately fixing a stale comment at
`internal/cli/list_test.go:275` (*"CLI tests cannot import database/sql
per the no-sql-in-cli-layer"*), which those same four files already
disprove. That is new test surface and a new scope decision this
docs-and-comments spec was never framed to carry (its own Outputs section
says "every code change is a comment"); deciding it inside a punch-list
return trip would repeat the exact failure pattern
`guidance/constraints.yaml`'s `no-sql-in-cli-layer` entry itself warns
against elsewhere in this repo's history (STAGE-021/constraints.yaml
§"Packaging-shape changes are decisions, not hygiene" — small-looking
diffs that are actually decisions deserve a decision pass, not an
inline fix).

**Where (b) landed instead.** Per the invoking instruction's own steering
("STAGE-022's stated scope is signals that gate CI... An audit test may
belong there"), a Design Notes entry was added to
`projects/PROJ-007-quality-and-portfolio-readiness/stages/STAGE-022-measured-and-enforced.md`
naming the gap, the four colliding test files, the stale
`list_test.go:275` comment, and a candidate mechanism this repo doesn't
have yet: a `depguard` golangci-lint rule scoped to `internal/cli/**`,
which would pair naturally with STAGE-022's already-scoped lint-config
item rather than requiring a hand-rolled walker test.

### P2 — build reflection Q2

Corrected in place (original claim retained below a provenance-style
divider, matching this spec's own established pattern for closed
questions in literal ⑤): the reflection now records that
`no-sql-in-cli-layer` had no automated guard for `internal/cli` at build
time, contrary to what Q2 originally asserted.

### P3 — "the only package that imports a SQL driver"

Fixed in `internal/storage/store.go`'s comment (and literal ①) by naming
the storage **layer** — the repo's own established term
(`storagetest.go`'s own comment: *"keeps the database/sql dependency
inside the storage layer"*) — and explicitly naming
`internal/storage/storagetest` as the sibling package that also imports
`modernc.org/sqlite`.

### P4 — "project.go adds the projects/locations schema"

Fixed in the same comment: `project.go` now reads *"holds the
projects/locations operations over the 0004_add_projects.sql schema
(DEC-017/019/020)"*. Confirmed directly: `project.go` contains zero
`CREATE`/`ALTER`/`DROP` statements (it is all `*Store` methods —
`CreateProject`, `AddLocation`, `ProjectForPath`, etc.); the schema comes
from `internal/storage/migrations/0004_add_projects.sql`
(`CREATE TABLE projects`, `CREATE TABLE project_locations`); DEC-017 line
45 — the first DEC the original comment cited — states this itself:
*"The `0004_add_projects.sql` migration adds two tables and backfills
nothing."*

### P5 — "the row below"

One word, fixed in both `docs/engineering-practices.md` and literal ⑥:
`below` → `above`. Confirmed by position: the guarded inventory table
(`<!-- inventory:begin -->` … `<!-- inventory:end -->`) sits above the
`## Decisions` section in the file; the next bullet in the same section
already correctly says *"see the lowest, highest and 1.0 rows above."*

### P6 — the reservation-filter totality gap (non-blocking, decided)

**Decided: add `reservation` to `decisions/_template.md`'s documented
enum now; defer the totality assertion (`decs + decs_reserved == ls
decisions/DEC-*.md | wc -l`) with this reason recorded.**

The template fix is a one-line, zero-risk documentation correction —
`decisions/_template.md` isn't itself a `DEC-*.md` file, so it is outside
every inventory count and every `test-docs.sh` assertion; there is
nothing to reconcile or re-derive. It closes half the gap verify named
(the enum silently omitting a type every future tombstone author would
otherwise have to infer from DEC-041 alone).

The totality assertion is deferred, not skipped. Adding it is a real
design decision this punch-list return trip should not make by
implication: it requires choosing what a violation *means* (a third,
untyped `insight.type` on a `DEC-*.md` file is either a typo to fail the
build on, or a legitimately new category `inventory.sh` hasn't been
taught yet — those two failure modes want different assertion shapes),
and, per this spec's own ground rules, any new `test-docs.sh` assertion
must be added by anchoring on line-start matches with the match count
verified before writing — exactly the discipline that caught this
spec's own earlier truncation incident (1,779 lines → 715). Rushing that
inside a five-finding punch-list fix risks the same class of mistake the
ground rules are warning against. The gap itself is low-severity by
construction: it requires a *new* `DEC-*.md` file to carry a bogus or
absent `insight.type`, and every one of the 46 files that exist today
(45 decisions + the DEC-041 reservation) was independently verified at
SPEC-080 design and build to carry a correct one (§12(b) row 7; Y3's
mutation coverage). A future spec adding to `decisions/` — or STAGE-022,
whose whole charter is closing exactly this class of "measured but not
enforced" gap — is the right place to design and land the assertion
deliberately.

*(Routed at the re-verify return trip, 2026-08-20 — R3. Naming two candidate
venues and writing to neither is not a deferral, it is a disappearance. The
assertion now has its own bullet in STAGE-022's `## Design Notes`, carrying
the gap, the reason it is deferred, what would resolve it, and a trigger. See
"Re-verify Punch List Resolution" below.)*

---

## Re-verify Findings (scoped delta review, 2026-08-20)

*A fresh session per AGENTS.md §6, reviewing `git diff 19764dd..b5a2c66` only.
The mechanical half approved at `19764dd` was not redone. Verdict:
**⚠ PUNCH LIST** — three findings, one of them the same claim verify #1
already flagged.*

**Gates, re-run on the delta:** `just test` PASS; `just test-docs` PASS
(177 `OK:` lines / 176 distinct ids, `S3` the documented double-emitter);
`gofmt -l .` clean; `go vet ./...` clean; `go build ./...` clean. The
inventory block round-trips **byte-identical** against `./scripts/inventory.sh`
(21 rows). Artifact↔literal parity holds for **all five** literal ① comments
(byte-identical by extracted line range) and for all three hunks of literal ⑥.

**P1 demonstration, reproduced independently.** `_ "database/sql"` was added
to the production `internal/cli/root.go`; `gofmt -l .`, `go build ./...`,
`go vet ./...`, `just test` (and a forced `go test -count=1 ./internal/cli/`,
to rule out a cached result — `ok 2.299s`), and `just test-docs` (177 `OK:`,
exit 0) **all passed**. The mutation was reverted and the tree confirmed clean
by `git status --porcelain` (empty) and by hash: `git hash-object
internal/cli/root.go` == `git rev-parse HEAD:internal/cli/root.go` ==
`a4e5ab6ee40168ee960a2882b4577833ee7867c0`. The corrected P1 claim is **true**,
and every component of it independently confirmed: `TestNoSQLImport` exists at
`internal/mcpserver/import_audit_test.go` and covers `internal/mcpserver`; the
constraint's `paths:` glob is exactly `["internal/cli/**"]`; `internal/cli` has
no equivalent test. P2, P4 and P5 are likewise **true** (evidence in each
finding below where relevant).

### R1 — P3's replacement is narrower but still false (blocking)

`internal/storage/store.go` now claims storage is *"the only **layer** that
imports a SQL driver."* It is not. Package `cli` imports `modernc.org/sqlite`
directly in four files:

```
internal/cli/story_test.go:14    _ "modernc.org/sqlite"
internal/cli/wrapped_test.go:17  _ "modernc.org/sqlite"
internal/cli/coverage_test.go:16 _ "modernc.org/sqlite"
internal/cli/impact_test.go:15   _ "modernc.org/sqlite"
```

All four declare `package cli` — the CLI layer, not the storage layer.

The `package` → `layer` edit escapes the one counterexample verify #1 showed
it (`storagetest`) without checking whether others exist. The obvious defence —
"layer means production architecture; test files don't count" — is **not
available to this sentence**, because its own second clause admits a
test-support package into the claim's scope: *"its storagetest test-helper
subpackage imports the same driver for tests that need raw SQL."* Having
counted test-support imports as in-scope, the claim cannot then exclude
`internal/cli`'s. The sibling comment written **in the same commit** carries
exactly the qualifier this one lacks — `root.go`'s *"Its **production code**
imports no SQL driver"* — so the delta scopes one claim correctly and leaves
its twin unscoped. Two comments in one commit that cannot both be true.

Fix is a qualifier, not a rewrite (e.g. *"the only layer whose production code
imports a SQL driver"*), and must land in **both** `internal/storage/store.go`
and literal ① in this spec.

### R2 — STAGE-022's cost estimate is short by one file for the mechanism it recommends first (minor)

The routing note names four colliding test files (`coverage_test.go`,
`impact_test.go`, `project_test.go`, `wrapped_test.go`) as already importing
`database/sql`. That is **exactly right for a ported `TestNoSQLImport`**, which
greps the literal string `"database/sql"` — verified against the four files.

But the note recommends, first and as the "natural fit," *a `depguard`
golangci-lint rule scoped to `internal/cli/**`*, and the constraint it would
implement forbids *"database/sql **or any SQL driver**."* A depguard rule
written to the constraint also trips `internal/cli/story_test.go`, which
imports the driver but **not** `database/sql` — five files, not four. The note
under-states the cost of its own preferred mechanism.

Two related facts belong in the same note: (a) `TestNoSQLImport` greps only
`"database/sql"`, so a straight port closes only **half** the constraint and
would not catch a driver-only import; (b) the constraint's text is
`"Files under internal/cli/ must not import database/sql or any SQL driver"` —
unqualified by production-vs-test. On its literal reading, five test files
violate a **blocking** constraint today. The note frames this correctly as
"first decide whether the constraint covers test files," but does not flag that
the text as written already answers it in the direction that makes the current
tree non-compliant. That strengthens the case for declining it here (see R4)
and sharpens what STAGE-022 has to decide.

### R3 — the deferred totality assertion was not routed anywhere durable (minor)

P1's follow-up **was** routed well (see R4). The totality assertion was not
routed at all. `STAGE-022-measured-and-enforced.md` mentions `SPEC-080` exactly
once — in the P1 bullet — and contains no occurrence of `decs_reserved`,
`totality`, or `reservation`. The deferral exists only in this spec's
`## Punch List Resolution`, and `scripts/archive-spec.sh` moves a shipped spec
into `done/`.

The spec's own wording is the tell: *"A future spec adding to `decisions/` — or
STAGE-022 … — is the right place."* It names two candidate venues and writes to
neither. This repo already has an established venue for exactly this — STAGE-022's
own second Design Note points at `projects/PROJ-001-mvp/backlog.md` for the two
other deferred items and says "do not re-derive them." One bullet, in STAGE-022
or that backlog, makes "deferred, not skipped" true.

### R4 — judgements the delta got right

- **P1's fork (a) over (b) — correct, and for the right reason.** This spec's
  own Outputs says *"every code change is a comment"*; porting the audit test
  adds production test surface **and** forces a constraints.yaml scope decision
  the constraint's text leaves genuinely unsettled (see R2(b)). Declining that
  inside a comments-only punch-list return trip was right. Crucially the fork
  was closed by making the comment honest rather than by leaving the claim
  standing — the correct direction, and the routing to STAGE-022 is real: the
  bullet is **first** in `## Design Notes`, states the defect declaratively,
  names the mechanism, the empirical proof, the scope question, the four files,
  and the stale `list_test.go:275` comment, and hooks explicitly to the stage's
  own in-scope item #1 (lint config). Someone framing that spec would find and
  act on it. Its one weakness is that the Spec Backlog one-liner for the lint
  item does not mention it — but AGENTS.md §12 requires reading the parent
  `STAGE-*.md`, so it is findable by the documented path.
- **The totality split is principled in reasoning, incomplete in execution.**
  Not the easy half taken and the hard half renamed: the two halves differ in
  kind. The template edit is a documentation fix to a non-inventoried file, and
  it is genuinely count-neutral — every `inventory.sh` glob is
  `decisions/DEC-*.md`, which `_template.md` does not match, and the inventory
  block round-trips byte-identical after the edit (verified). The assertion
  really does require choosing what a violation *means* — a typo to fail the
  build on, or a new category `inventory.sh` has not been taught — and those two
  want different assertion shapes. That is a design decision, not a chore. Only
  the routing failed (R3).
- **P2's correction is honest.** The original wrong belief is retained verbatim
  (*"an added SQL import would have broken the `no-sql-in-cli-layer`-guarded
  build"*) with the correction below it, explicitly stating *"the record of
  having believed it is kept rather than deleted."* Every factual claim in the
  correction was independently reproduced above.
- **P4 confirmed by direct reading.** `internal/storage/project.go` contains
  zero DDL — 15 `*Store` methods, and the only `CREATE`-matching strings are
  `fmt.Errorf("create project: …")` error texts. The schema comes from
  `internal/storage/migrations/0004_add_projects.sql`
  (`CREATE TABLE projects`, `CREATE TABLE project_locations`,
  `CREATE INDEX idx_project_locations_project`). DEC-017 line 45 reads
  *"The `0004_add_projects.sql` migration **adds two tables and backfills
  nothing**"* — the citation says what the comment implies.
- **P5 confirmed by position.** The guarded inventory table spans lines 26–48
  of `docs/engineering-practices.md`; the corrected sentence is at line 66.
  The table is **above**. The sibling bullet four lines later independently
  agrees (*"see the lowest, highest and 1.0 rows above"*). Both the artifact
  and literal ⑥ carry `above`; the two surviving instances of `below` at spec
  lines 1055 and 1268 are quoted history, not live claims — correct.
- **Adjacency survives.** `go doc` renders the package comment for all five
  packages (`internal/cli`, `internal/config`, `internal/storage`,
  `internal/story`, `internal/mcpserver`); none was pushed off its `package`
  line.

### Untouched-claim spot-check

Sampled claims the punch list did not touch, in the two changed comments and
the three unchanged ones:

| Claim | Verdict |
|---|---|
| `store.go`: "Entry and ListFilter (entry.go) are the query vocabulary" | ✅ `entry.go:11`, `entry.go:26` |
| `store.go`: Open behind backup (backup.go, DEC-021) + devguard (devguard.go, DEC-026) | ✅ both called in `Open`; both DEC ids match their filenames |
| `root.go`: "ErrUser/UserErrorf for exit-code classification (errors.go)" | ✅ `errors.go:11`, `errors.go:13` |
| `root.go`: "the injectable clock seam tests substitute (clock.go)" | ✅ `clock.go` |
| `root.go`: "atomic same-directory-rename config writes (atomicwrite.go)" | ✅ `CreateTemp(filepath.Dir(path))` + `os.Rename`, `atomicwrite.go:25-53` |
| `bundle.go`: "reads storage.Entry values but never a `*storage.Store` or a SQL driver" | ✅ only occurrence of `storage.Store` in the package is the comment itself |
| `config.go`: "no dependency on internal/storage, so both can import it without a cycle" | ✅ `internal/config` has no internal imports at all |
| `root.go`: "calendar-window flag parsing shared by impact/story (window.go)" | ⚠ incomplete — see O1 |

### Observations — not blocking, no action required in this spec

- **O1. The `window.go` enumeration names two of three consumers.**
  `internal/cli/coverage.go` calls both `selectedWindow` (`coverage.go:59`) and
  `windowCutoff` (`coverage.go:68`), same as `impact.go` and `story.go`. The
  claim is not false — it asserts no uniqueness — but it is incomplete in a
  comment whose product is orientation. One word (`impact/story/coverage`)
  if a later spec is in the file anyway.
- **O2. `decisions/_template.md` is absent from `## Outputs`.** The P6 fix
  modified it; the Outputs "Files modified" list still enumerates six artifacts
  without it. It is documented in `## Punch List Resolution`, so it is
  traceable, but a rebuild driven off Outputs would miss it.

### On what is now pinnable

The genuinely pinnable proposition in this delta is the SQL boundary itself,
not the prose about it. Once R1 is fixed to a production-scoped claim, a grep
over non-`_test.go` files under `internal/` and `cmd/` for `"database/sql"` and
`modernc.org/sqlite`, asserting no hit outside `internal/storage/`, checks the
world rather than the words — and it passes today (only `store.go`, `migrate.go`,
`backup.go`, `devguard.go`, `project.go`, and `storagetest.go` match, all inside
`internal/storage/`). It is correctly **left unpinned here**: that assertion is
precisely the STAGE-022 follow-up, and adding it in this spec would contradict
P1's own reasoning for choosing fork (a).

What must **not** be added is an `assert_contains` on the comment text. It would
go green on the words while saying nothing about whether `internal/cli` still
has no driver import — the SPEC-078 H9 failure mode, already on this repo's
record, and the reason this delta needed its own review at all.

---

## Re-verify Punch List Resolution (second build return trip, 2026-08-20)

*A third fresh session, per AGENTS.md §6, closing R1–R3. R4's judgements were
approved and are not revisited; the six-literal transcription, `Y1`–`Y5` and
their mutations, the 171 → 176 id delta, the reservation seam, and P1/P2/P4/P5
are untouched.*

### R1 — chosen wording: scope the claim to production code

**Chosen: qualify both uniqueness claims in `internal/storage/store.go`'s
package comment with "production code," matching the qualifier
`internal/cli/root.go` already carries.**

- *"the only layer that imports a SQL driver"* → *"the only layer **whose
  production code** imports a SQL driver"*.
- *"Every other package reaches the database only through a `*Store`"* →
  *"Every other package**'s production code** reaches the database only
  through a `*Store`"*.

The second edit is the one the finding did not name. R1's instruction was to
test the **whole sentence**, not the counterexample handed over; doing so
surfaced a second, unscoped uniqueness claim eleven lines further down the same
comment, false for exactly the same reason — `internal/cli`'s test files reach
the database directly, not through a `*Store`
(`coverage_test.go:56`, `impact_test.go:57`, `project_test.go:1004`,
`wrapped_test.go:60` each call `sql.Open("sqlite", dbPath)` and then `db.Exec`).
Scoping one and leaving the other would have reproduced the exact defect R1
described — two claims in one comment that cannot both be read the same way.

The `storagetest` clause was also re-worded so it can no longer be read as
admitting test-support code into the uniqueness claim's scope: *"Test files
elsewhere import it too; the storagetest test-helper subpackage exists so they
need not, keeping raw SQL inside `internal/storage`."* The comment now states
the counterexample itself rather than leaving it for the next reader's grep to
find.

**Evidence the whole sentence is true — not just that the named counterexample
is gone.** Every non-`_test.go` file in the repo that imports `database/sql` or
`modernc.org/sqlite`, enumerated exhaustively
(`grep -rn '"database/sql"\|modernc.org/sqlite' --include='*.go' .`, import
lines only, test files excluded):

| Non-test file | Imports | Layer |
|---|---|---|
| `internal/storage/store.go` | both | storage |
| `internal/storage/backup.go` | `database/sql` | storage |
| `internal/storage/devguard.go` | `database/sql` | storage |
| `internal/storage/migrate.go` | `database/sql` | storage |
| `internal/storage/project.go` | `database/sql` | storage |
| `internal/storage/storagetest/storagetest.go` | both | storage |

Six files, one layer, no other. The claim holds under **both** available
readings of "production code" — as *non-`_test.go` files* (`storagetest.go` is
one, and it is inside `internal/storage/`), and as *code that ships in the
binary* (`storagetest.go` is not, so it is simply out of scope). It is not
narrowed to survive one counterexample; it is the proposition the STAGE-022
grep the re-verify named would actually check.

Cross-checked in the other direction: `internal/cli` has **no** non-`_test.go`
file importing either — so `root.go`'s *"Its production code imports no SQL
driver and no database/sql"* and `store.go`'s *"the only layer whose production
code imports a SQL driver"* are now the same proposition seen from each side.
The two comments agree.

#### Rejected alternatives (AGENTS.md §12)

- **Widen the exception list instead of scoping the claim** — *"the only layer
  that imports a SQL driver, apart from four `internal/cli` test files."*
  **Rejected.** It is the narrow-to-survive-this-counterexample move a third
  time, and it pins a file count that drifts the next time a test opens a
  database. A scope is stable; an exception list is a countdown.
- **Drop the uniqueness claim and say what storage *is*** — *"Package storage
  owns the SQL driver and every persistence operation."* **Rejected.** The
  uniqueness *is* the architectural fact the package comment exists to carry
  (DEC-001 plus the `no-sql-in-cli-layer` boundary); removing it because it was
  hard to scope trades a checkable claim for a vaguer one, and would leave
  `root.go`'s production-scoped claim with no counterpart on the storage side.
- **State the driver's location without a superlative** — *"the driver is
  imported in `store.go` and in the `storagetest` subpackage."* **Rejected.**
  A file enumeration in a package comment goes stale on the next file that
  opens a database; the scoped superlative is shorter, says more, and a grep
  can check it.
- **Rely on an unstated "test files don't count" convention.** **Rejected** —
  and it was never available: the sentence's own second clause counts
  `storagetest`, a test-helper package. Making the scope explicit is the fix.
- **Also fix `store.go`'s `Store` *type* comment in this spec** (*"All
  persistence flows through a Store; no other package imports a SQL driver"* —
  false in the same way: `internal/storage/storagetest` is another package and
  imports one, as do four `internal/cli` test files). **Rejected for this
  spec.** It is unchanged since SPEC-002 (`git blame`: `02dcd0e`, 2026-04-20),
  it is not part of literal ①, and this spec's `## Outputs` states *"No other
  line in any of these five files changes"* — editing it would break the
  artifact↔literal contract verify diffs against. Routed to STAGE-022 instead,
  in the same bullet that already carries the stale `list_test.go:275` comment.

### R2 — STAGE-022's routing note now costs the mechanism it recommends

Corrected in
`projects/PROJ-007-quality-and-portfolio-readiness/stages/STAGE-022-measured-and-enforced.md`,
first Design Note. Both missing facts added, both verified first:

- **`TestNoSQLImport` closes half the constraint.** `import_audit_test.go:27`
  matches one string only — the import path as it appears in source,
  `"database/sql"`, double quotes included — and nothing else. A driver-only
  import walks past it.
- **Five files, not four.** `database/sql` in `internal/cli`:
  `coverage_test.go:5`, `impact_test.go:5`, `project_test.go:5`,
  `wrapped_test.go:5`. `modernc.org/sqlite` in `internal/cli`:
  `coverage_test.go:16`, `impact_test.go:15`, `story_test.go:14`,
  `wrapped_test.go:17`. The two sets are not the same four:
  `project_test.go` imports `database/sql` but not the driver (it relies on the
  blank import in its siblings to register `sqlite`), and `story_test.go`
  imports the driver but not `database/sql`. Union = **five**. The old note's
  four are correct for a ported `TestNoSQLImport` and short by one for the
  `depguard` rule it recommends first.
- **Which way the constraint's text already points.** Its rule is *"Files under
  `internal/cli/` must not import `database/sql` or any SQL driver"* — no
  production-vs-test qualifier. On the literal reading, five test files violate
  a blocking constraint today. The note now says so, and says explicitly that
  amending the constraint is a decision for the spec that picks this up, not a
  side effect of wiring a lint rule. **`guidance/constraints.yaml` is
  unchanged.**
- **One fact added beyond what R2 asked, because testing the note the way R1
  demanded surfaced it.** The re-verify's own sketch of the future assertion —
  a grep over non-`_test.go` files for `database/sql` and `modernc.org/sqlite`
  — false-positives today: `internal/cli/root.go:8` is the only non-test hit
  outside `internal/storage/`, and it is the package comment *describing* the
  boundary, not an import. (It already read that way at `f3514cb`.) The
  mechanism must match import lines rather than file text —
  `TestNoSQLImport` handles the same hazard by excluding itself by filename
  (`import_audit_test.go:20`); `depguard` avoids it by working on the import
  graph. Recorded in the same Design Note so the assertion is not written twice.

### R3 — the totality assertion, routed to STAGE-022

**Venue: STAGE-022's `## Design Notes`, as its own bullet.**

Why there and not `projects/PROJ-001-mvp/backlog.md`: the gap is CI-shaped —
its resolution is one `scripts/test-docs.sh` assertion — and STAGE-022's
charter is precisely "measured but not enforced." Its first backlog spec is the
lint-and-CI pass, and AGENTS.md §12 ("During build") requires reading the
parent `STAGE-*.md`, so anyone framing that spec meets the bullet on the
documented path. The backlog is the right venue for a deferred *idea* with a
revisit trigger; this is a deferred *check* with a decision attached to it.

Why it survives this spec's archiving: `scripts/archive-spec.sh` moves spec
files into `done/`. There is no `archive-stage.sh` and no `just` recipe that
touches stage files — `justfile:83` is the only archive recipe and its argument
is a `SPEC-NNN`. A stage file stays in
`projects/PROJ-007-quality-and-portfolio-readiness/stages/` when the stage
ships; only its `status:` changes.

The bullet carries what the finding required: **what the gap is** (nothing
asserts *Decision records* + *Decision numbers reserved* equals
`ls decisions/DEC-*.md | wc -l`, so a file with a third `insight.type` or none
is counted by neither row and vanishes silently while the page still
round-trips); **why it is deferred** (the assertion cannot be written without
deciding whether an unknown type is a typo to fail on or a category
`inventory.sh` has not been taught — different assertion shapes); **what would
resolve it** (pick the meaning, add the one assertion, keep
`decisions/_template.md` outside the glob as it already is); and a **trigger**
(the next spec adding to `decisions/`, or this stage's lint-and-CI spec).
Current state recorded for the next reader: 45 + 1 = 46, exact today.

### O1 — folded in

The re-verify left this to judgement. **Folded in**, because it was a
one-word factual completion in a comment whose whole product is orientation,
and leaving a known-incomplete enumeration in place to be re-found later is the
same deferral habit R3 was about. `internal/cli/root.go` now reads *"calendar-
window flag parsing shared by impact, story and coverage (window.go)"*.
Verified the enumeration is now complete, not merely longer: `selectedWindow`
and `windowCutoff` have exactly three non-test callers —
`coverage.go:59`/`:68`, `impact.go:66`/`:75`, `story.go:88`/`:96`/`:108`.
`spark.go` names `selectedWindow` only in comments (`spark.go:73`, `:166`), to
record that it deliberately uses a different flag set. Literal ① updated to
match.

### What was deliberately not done

- **No `assert_contains` on comment text.** SPEC-078's H9 failure mode; it goes
  green on the words and says nothing about the world.
- **No production-code grep assertion in `scripts/test-docs.sh`.** The
  re-verify named a good one and left it for STAGE-022; adding it here is new
  production test surface, which is the same reasoning that settled P1's fork.
  `test-docs.sh` is unchanged by this return trip — still 176 distinct ids.
- **No change to `guidance/constraints.yaml`.** R2's finding is about a note
  describing the constraint, not about the constraint.
- **No change to `README.md`** (A1 at 260, zero headroom) and no change to
  `scripts/inventory.sh`, `guidance/questions.yaml`, or
  `docs/engineering-practices.md` — this trip touches two Go comments, one
  stage file, and this spec.

---

## Second Re-verify Findings (final scoped pass, 2026-08-20)

**✅ APPROVED.** A fourth fresh session per AGENTS.md §6, scoped to
`git diff f3514cb..9039818`. Everything the two prior passes approved was
taken as approved and not re-derived. No blocking finding. Two non-blocking
observations, both recorded below; the first has been routed into STAGE-022
(a one-sentence addition to the Design Note it qualifies, made here rather
than sent back — see *What this pass changed*).

### R1 — both claims tested independently, exhaustively

Gathered from the module's own import graph rather than from grep, so the
answer does not depend on how a comment happens to be worded:
`go list -f '{{.Imports}}' ./...` for production imports, `.TestImports` +
`.XTestImports` for test-only imports, cross-checked per file against every
tracked `.go` file's import block.

**Claim A — *"the only layer whose production code imports a SQL driver."***
Exactly two non-`_test.go` files in the module import `modernc.org/sqlite`:
`internal/storage/store.go:29` and
`internal/storage/storagetest/storagetest.go:13`. Both are inside
`internal/storage/`. Widening to `database/sql` adds `backup.go:5`,
`devguard.go:5`, `migrate.go:5`, `project.go:5`, `store.go:20`,
`storagetest.go:9` — same directory, no others anywhere. At package
granularity, `go list` reports production driver imports for
`internal/storage` and `internal/storage/storagetest` and for no other
package in the module. **True**, and true under both readings of "production
code" the return trip named: as *non-`_test.go` file* (`storagetest.go` is
one, and it is inside `internal/storage/`) and as *code that ships*
(`go list -deps ./cmd/brag` contains no `storagetest` — the subpackage is
imported only from `_test.go` files, in `internal/cli`, `internal/mcpserver`
and `internal/storage`). The `.claude/worktrees/` tree was excluded
throughout: it is gitignored, untracked, and absent from `go list ./...`.

**Claim B — *"Every other package's production code reaches the database
only through a `*Store`."*** True under the ships-in-the-binary reading, for
the reason above. Under the non-`_test.go`-file reading it has exactly one
counterexample, `internal/storage/storagetest/storagetest.go`: a distinct Go
package whose non-test file calls `sql.Open("sqlite", dbPath)` and
`db.Exec("UPDATE entries SET created_at = ? WHERE id = ?", …)` without a
`*Store` in sight. Claim A's noun is **layer** and survives both readings;
Claim B's noun is **package** and survives one. See observation **V1** — it
is not blocking, and the reasoning is recorded there rather than here.

Beyond `storagetest`, Claim B holds without qualification: no non-`_test.go`
file outside `internal/storage/` imports `database/sql` or a driver at all,
so no other package has any way to reach the database except through a
`*Store`.

### Do the two comments state the same proposition from each side?

Yes, over the boundary they both describe, and not by accident.
`internal/cli/root.go:7` claims *"Its production code imports no SQL driver
and no database/sql"*; `internal/storage/store.go:11` claims *"Every other
package's production code reaches the database only through a `*Store`."*
Both are grounded in one verified fact — `internal/cli` has zero
non-`_test.go` files importing either — and they fail together: a driver
import added to any `internal/cli` production file falsifies `root.go`
directly and falsifies `store.go` in the same edit, because that file would
then reach the database other than through a `*Store`. They carry the same
`production code` qualifier, cite the same constraint by id, and point at the
same STAGE-022. The agreement is structural, not coincidental. `root.go`'s
claim is the locally stronger one (it also excludes bare `database/sql`);
`store.go`'s is the wider one (it ranges past `internal/cli`), which is
where V1 lives.

### The `storagetest` clause

*"Test files elsewhere import it too; the storagetest test-helper subpackage
exists so they need not, keeping raw SQL inside `internal/storage`."*
Accurate on both halves, and it does close the ambiguity. On *why it
exists*: `storagetest.go`'s own package comment says the same thing
independently — *"Living under internal/storage/ keeps the database/sql
dependency inside the storage layer, which lets CLI tests use these helpers
without violating the no-sql-in-cli-layer constraint."* The description and
its citation agree. On the *ambiguity*: the earlier wording was false because
a reader could not tell whether test files counted. This wording counts them
out loud in its first clause — *"Test files elsewhere import it too"* — and
then explains them, rather than relying on an unstated convention. It states
its own counterexample instead of leaving it for the next grep, and it is
honest that `storagetest` has not displaced the direct imports (five
`internal/cli` test files still carry them); *"so they need not"* is a
statement of purpose, which is what it should be.

### Leaving `store.go`'s `Store` **type** comment unfixed — judgement

The comment (`store.go:32-34`, *"All persistence flows through a Store; no
other package imports a SQL driver"*) is false, unqualified, in both halves:
`storagetest` is another package that imports a driver, and
`storagetest.Backdate` is a persistence write that does not flow through a
`Store`. `git blame` puts it at `02dcd0e`, 2026-04-20 — SPEC-002, pre-dating
this spec by four months.

**The build's call was right, and I would take the same one.** Three
reasons, in order of weight:

1. **AGENTS.md §9 already decides this case.** The audit-grep cross-check
   bullet (SPEC-018 lesson) tells build to treat a delta between what it
   finds and what `## Outputs` enumerates as *a question for the spec author,
   not a unilateral expansion of scope*. A true defect found mid-return-trip
   is exactly that delta. The build held the line and routed it — the
   codified behaviour, not an improvised one.
2. **The `Outputs` contract is what makes the literals checkable at all.**
   This spec's product is six literal artifacts that verify diffs against
   byte-for-byte. If build may take one more line because the fix is
   obviously right, the parity check stops being a check — and this spec, of
   all specs, is about claims that survive scrutiny by being scrutinizable.
   The alternative reading ("a spec about true comments should not ship a
   false one") is real, but it buys one corrected line at the cost of the
   mechanism that catches the next twenty.
3. **The cost is bounded and disclosed.** The defect is pre-existing, has no
   new victim, and now sits eleven lines below a package comment that says
   the true thing in the reader's path. It is recorded in two places with a
   named owner.

**Routing exists and is actionable.** STAGE-022 `## Design Notes`, the
*"Stale comments to fix alongside"* bullet: it names the file, quotes the
exact false substring, names both counterexample classes (`storagetest` and
the `internal/cli` test files), records the blame provenance and why
SPEC-080 could not touch it, and pairs it with the second stale comment at
`internal/cli/list_test.go:275` (verified: that line does read *"CLI tests
cannot import database/sql per the no-sql-in-cli-layer constraint"*, and the
four files disprove it). It sits in the same note as the mechanism that will
make the boundary enforceable, so it gets fixed by the spec with the most
reason to care.

### R2 / R3 — would someone framing STAGE-022 act on these?

Yes. Each was re-derived from the repo, not read.

- **Five, not four — verified by set difference.** `database/sql` in
  `internal/cli`: `coverage_test.go:5`, `impact_test.go:5`,
  `project_test.go:5`, `wrapped_test.go:5`. `modernc.org/sqlite`:
  `coverage_test.go:16`, `impact_test.go:15`, `story_test.go:14`,
  `wrapped_test.go:17`. `project_test.go` imports `database/sql` and not the
  driver; `story_test.go` imports the driver and not `database/sql`. Union =
  **five**. All eight cited line numbers are exact.
- **The driver-vs-`database/sql` distinction is real.**
  `import_audit_test.go:27` is `strings.Contains(string(b), `"database/sql"`)`
  — one literal, file text, nothing else; `:20` is the self-exclusion by
  filename. A driver-only import walks past it, and `story_test.go` is that
  case today. Both cited lines are exact.
- **The false-positive note is the most useful thing in the delta, and it is
  true.** Run over tracked files, `grep -rn
  'database/sql\|modernc.org/sqlite' --include='*.go' .` minus `_test.go`
  minus `^./internal/storage/` returns exactly one line:
  `internal/cli/root.go:8` — the package comment *describing* the boundary,
  not an import. The note now leads with **"A content grep must match import
  lines, not file text"** and points at both the hazard
  (`TestNoSQLImport`'s own filename exclusion) and the mechanism that
  sidesteps it (`depguard`, on the import graph). Without this, the first
  person to write the assertion writes a rule that fires on the sentence
  telling them the rule is needed. *(Caveat for whoever writes it: run
  verbatim on a working copy the grep also sweeps untracked scratch trees —
  `.claude/worktrees/` added 13 hits here. Scope to tracked or module files;
  the recommended import-graph mechanism avoids this too.)*
- **The constraint's text is quoted correctly** and
  `guidance/constraints.yaml` is unchanged (`internal/cli/**`, `blocking`,
  no production-vs-test qualifier — so the note's "five files violate a
  blocking constraint on its literal reading" is right, and it correctly
  refuses to amend the constraint as a side effect).
- **The totality bullet survives archiving.** `scripts/archive-spec.sh`
  performs exactly one `mv` (line 49) and it moves the spec file; it only
  *reads* the stage file (`find_stage`, line 56) to echo its path. No
  `sed -i`, no redirect onto a stage path. `justfile:83` is the only archive
  recipe and its argument is a `SPEC-NNN`; there is no `archive-stage.sh` in
  `scripts/`. The bullet's own current-state number is exact: 46 `DEC-*.md`
  files = 45 `type: decision` + 1 `type: reservation`.

### O1 — the enumeration is complete, not merely longer

`selectedWindow` and `windowCutoff` have exactly three non-test **caller**
files: `coverage.go:59`/`:68`, `impact.go:66`/`:75`, `story.go:88`/`:96`/
`:108`. Every other tracked occurrence is a comment (`impact.go:15`/`:58`,
`story.go:76`/`:77`/`:84`, `window.go` itself, `store_test.go:811`) or a test
call (`impact_test.go`, nine sites). `spark.go` names `selectedWindow` only
at `:73` and `:166`, both comments, both recording that it deliberately uses
a different flag set. `coverage.go:43-46` registers the same four flags
`windowFlagNames` holds, so `coverage` is a real member of the set, not a
coincidental caller. *"impact, story and coverage"* is the complete list.

### Parity, adjacency, gates

- **Literal ① — byte-identical, diffed by extracted line range.** The five
  fenced `go` blocks under `### ①` were extracted programmatically (spec
  lines 571-698) and diffed against each file's leading `//` block: `main.go` 8
  lines, `root.go` 13, `config.go` 8, `store.go` 15, `bundle.go` 11 — five
  of five identical, sha256 match on each, `diff` empty on each.
- **The `Outputs` contract holds exactly.** `git diff main...HEAD` over the
  five files is **55 insertions, 0 deletions, 0 modifications** — every hunk
  is `@@ -0,0 +1,N @@`. *"No other line in any of these five files changes"*
  is literally true.
- **`go doc` on all five packages renders the comment.** Adjacency survived
  every edit: `./cmd/brag`, `./internal/cli`, `./internal/config`,
  `./internal/storage`, `./internal/story`.
- **Gates.** `just test` ok, `just test-docs` ok, `gofmt -l .` clean,
  `go vet ./...` clean, `go build ./...` clean. **177 `OK:` lines, 176
  distinct ids**, the single duplicate being `S3` — the documented
  pre-existing wart. Inventory block round-trips byte-identical against
  `./scripts/inventory.sh` (21 rows, `docs/engineering-practices.md:27`).
- **No truncation.** Spec 1496 → 1694 → 1917 lines across the three commits,
  26 fences, balanced; STAGE-022 168 → 239. Monotonic, nothing lost.

### Untouched-claim spot-check

Four claims neither punch list touched, checked against their own citations:

- `root.go` *"the injectable clock seam tests substitute (clock.go)"* —
  `clock.go` is `var clock = time.Now`, a package var; `list_test.go`
  substitutes it. **Holds.**
- `root.go` *"atomic same-directory-rename config writes (atomicwrite.go)"* —
  `atomicwrite.go:25` takes `filepath.Dir(path)`, creates the temp file
  there rather than in `os.TempDir()` (`:21` says so), and `:53` is
  `os.Rename`. The adjective "same-directory" is load-bearing and correct.
  **Holds.**
- `store.go` *"Open applies pending embedded migrations behind a
  pre-migration backup safety belt and a dev/prod migration guard"* — `Open`
  runs `devProdMigrateGuard` → `backupBeforeMigrations` → `applyMigrations`.
  "Behind A and B" asserts both precede, which they do; it asserts no order
  between A and B, and the code's own comment records that the guard is
  deliberately first. **Holds.**
- `bundle.go` *"reads storage.Entry values but never a `*storage.Store` or a
  SQL driver, and reads no clock itself"* — no production file in
  `internal/story` references `storage.Store`, `sql.`, `modernc`, or
  `time.Now()`; the only textual hit is the comment making the claim.
  **Holds.**

Every line number cited in the delta's own write-up (twelve of them, across
R1 and R2) resolves to the line claimed. That is a change from the earlier
passes, which found claims contradicted by their own citations.

### Observations — non-blocking, no action required in this spec

- **V1 — the two uniqueness claims use different nouns, and only one of them
  is airtight.** Claim A says *"only **layer**"*; Claim B says *"Every other
  **package**"*. `internal/storage/storagetest` is inside the layer and is
  another package, so the same subpackage that Claim A absorbs, Claim B
  admits as a counterexample under the non-`_test.go` reading. The delta's
  evidence paragraph runs the both-readings test and states its result — but
  states it for Claim A only; Claim B is never put through it. Why this is
  an observation and not a punch-list item: (a) Claim B is *true* under the
  ships-in-the-binary reading, which is a reading the return trip named and
  which `go list -deps` confirms — unlike P3 and R1's predecessors, which had
  no true reading at all; (b) the comment names `storagetest` as its own
  exception one sentence earlier, so the harm the finding guards against — a
  reader trusting the claim and being surprised — does not occur; (c) the
  operative gloss of the sentence is *"the no-sql-in-cli-layer boundary on
  `internal/cli`"*, which is unconditionally true; (d) the fix is a one-word
  noun swap inside literal ①, and re-opening the transcription surface for a
  noun is a poor trade in a stage where a literal-copy edit has already
  truncated a spec once. **Routed:** STAGE-022's *"Stale comments to fix
  alongside"* bullet certifies the package comments as *"the wording a
  mechanism should be checked against"*, which is the one place this could
  bite — a `depguard` rule written to the **package** noun would flag
  `storagetest`, one written to the **layer** noun would not. A qualifying
  sentence was added there by this pass.
- **V2 — `window.go:12-13` has the incompleteness O1 just fixed one file
  over.** `windowFlagNames`'s comment reads *"the canonical, ordered set of
  calendar-window flags shared by `brag impact` and `brag story`"*, but
  `coverage.go:43-46` registers the same four flags and `coverage.go:59`
  calls `selectedWindow`, which iterates the set. Same defect class as O1,
  same one-word cost, and it is the *definition site* — the place a reader
  looks first. Out of scope here: `window.go` is not one of literal ①'s five
  files and the `Outputs` contract is what makes this spec checkable. **Not
  routed**, and I am not inventing an owner for it: STAGE-022 is lint,
  coverage and the `Entries:` envelope, and a backlog entry costs more than
  the fix. It belongs to whoever next opens `window.go`. Recorded here so
  that person has it.

### What this pass changed

- `projects/.../stages/STAGE-022-measured-and-enforced.md` — one sentence
  appended to the *"Stale comments to fix alongside"* Design Note, carrying
  V1's noun distinction to the place a mechanism gets written. Additive,
  prose only, no assertion weakened. Called out explicitly per the pass's
  own ground rule against silent repairs.
- This section, and `cycle: build` → `cycle: verify` in the front-matter.
- **Not** changed: no `assert_contains` on comment text (SPEC-078 H9 — green
  on the words, silent on the world); no Go file; no literal; no
  `scripts/test-docs.sh`; no `guidance/constraints.yaml`.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — **Test the claim, not the counterexample.** The same defect recurred three
   times in this one spec: a comment claimed something false, verify named one
   counterexample, the fix narrowed the wording just enough to survive *that*
   example, and the narrowed claim was still false in general. `package` →
   `layer` fixed `storagetest` and stayed false for `internal/cli`. The round
   that finally worked did the opposite — it tested the whole sentence against
   the repo and found a **second** unscoped claim eleven lines further down that
   no finding had named. Three verify passes and two return trips is what
   narrow-to-survive costs.

2. **Does any template, constraint, or decision need updating?**
   — **`decisions/_template.md` gained `reservation`** to its `insight.type`
   enum — done in this spec, count-neutral.
   — **`no-sql-in-cli-layer` needs a decision this spec deliberately did not
   make.** It is a `severity: blocking` constraint whose path glob names
   `internal/cli/**`, and **nothing tests it there** — demonstrated twice by
   adding `_ "database/sql"` to `internal/cli/root.go` and watching every gate
   pass. Worse, its text is unqualified by production-vs-test, so on a literal
   reading **five `internal/cli` test files violate a blocking constraint
   today**. Amending the constraint is a decision; routed to STAGE-022 with the
   evidence rather than settled inside a comments-only spec.

3. **Is there a follow-up spec I should write now before I forget?**
   — **The README restructure** — the MCP call-to-action promoted out of the
   Status blockquote, and `test-docs` A1 switched from `wc -l` to `wc -w`. It is
   STAGE-021's third backlog item, agreed and not yet scaffolded.
   — **Four items are routed to STAGE-022**, all with evidence: the missing
   `internal/cli` audit test (with `depguard` named as the mechanism, the five
   colliding files, and the note that a content grep must match *import lines*,
   not file text — the sketched assertion false-positives on `root.go:8`, the
   comment describing the boundary); the deferred totality assertion; the stale
   comments (`store.go`'s `Store` type comment, false since 2026-04-20;
   `list_test.go:275`); and V1's package-vs-layer noun distinction.
   — **V2 is recorded and unrouted:** `internal/storage/window.go:12-13` says the
   window flags are "shared by `brag impact` and `brag story`" while
   `coverage.go` registers the same four and calls `selectedWindow`. The same
   defect this spec fixed one file over, at the definition site. Whoever next
   edits prose should fold it in.

4. **What can a user do now that they couldn't before?**
   — `go doc` on the five largest previously-undocumented packages — `cmd/brag`,
   `internal/cli`, `internal/config`, `internal/storage`, `internal/story` — now
   returns a comment saying what the package is *for* and where its boundary
   sits; `decisions/` no longer has an unexplained 040→042 gap; and the question
   register describes live uncertainty only, **6 open of 18** rather than 8, with
   both closures answered by artifacts that already existed on disk.

### What this spec proved about its own stage

STAGE-021's premise is that the discipline exists and is unsurfaced. This spec
is the second confirmation: framing expected to *write documentation* and found
**191 exported declarations with 3 undocumented**, all conventional interface
methods. The real gap was five package comments. Both of the framing pass's own
numbers were wrong — not stale, but produced by grep-shaped heuristics against an
unchanged tree — which is the same failure the practices page's derived counts
were built to prevent one spec earlier.

---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-080
  type: chore                      # epic | story | task | bug | chore
  cycle: build                     # frame | design | build | verify | ship
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
// calendar-window flag parsing shared by impact/story (window.go). It
// imports no SQL driver and no database/sql — enforced by the
// no-sql-in-cli-layer constraint — so every command reaches persistence
// only through internal/storage, keeping the CLI a thin shell a future
// frontend (TUI, API) could replace.
```

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
// Package storage is the only package that imports a SQL driver
// (modernc.org/sqlite, pure Go — no CGO, DEC-001). Store wraps *sql.DB and
// owns every persistence operation: Open applies pending embedded
// migrations (migrate.go) behind a pre-migration backup safety belt
// (backup.go, DEC-021) and a dev/prod migration guard (devguard.go,
// DEC-026); Entry and ListFilter (entry.go) are the query vocabulary;
// project.go adds the projects/locations schema (DEC-017/019/020). Every
// other package reaches the database only through a *Store — enforced by
// the no-sql-in-cli-layer constraint on internal/cli — which is what keeps
// commands testable and a future frontend feasible.
```

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
+count itself in the row below.

 - **Each record carries an honest confidence value.** `insight.confidence` is
   part of every record's front-matter, and no record claims certainty — see the
```

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

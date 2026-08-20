---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-022
  status: proposed                  # proposed | active | shipped | cancelled | on_hold
  priority: medium
  target_complete: null

project:
  id: PROJ-007
repo:
  id: bragfile

created_at: 2026-08-15
shipped_at: null
---

# STAGE-022: quality measured and enforced, known defects closed

## What This Stage Is

The standard signals a reviewer greps for exist and **gate CI** rather than
being available to run: a lint configuration, a coverage number with an honest
floor, and an empty known-defect list.

Two halves, both small:

- **Measured and enforced** — `golangci-lint` configured and gating,
  `go test -cover` reporting a number, a badge, and a floor chosen from the
  measured value rather than aspiration.
- **The one remaining known defect closed** — the `Entries:` envelope
  inconsistency. Two sibling items (`MergeTags` position density, `$EDITOR`
  quoting) were **deferred to `PROJ-001/backlog.md` on 2026-08-18** once each
  was measured: neither has a current victim, and both carry their evidence and
  a real revisit trigger. Deferring measured items is not the same as leaving
  the list untouched — what would make the project read as theatre is a quality
  pass that never looked.

## Why Now

Verified 2026-08-15, all four gaps still hold: no `golangci-lint` config, no
`-cover` anywhere, `func Benchmark` count **0**, no surfaced godoc. CI runs
`gofmt`, `go vet` and `go test` and nothing else.

**Second, not first** — see STAGE-021's *Why Now*. Lint and coverage are table
stakes; the discipline page is the differentiated half, and it establishes what
a coverage number is allowed to claim. Reversing the order invites a percentage
with no story attached, which is the exact failure mode SPEC-073 hit four times.

## Success Criteria

- `golangci-lint` runs in CI and **fails the build**, with a config whose
  enabled linters were each chosen, not inherited from a default set.
- Coverage is measured, reported, and carries a **floor that the current number
  actually clears** — set from measurement, and stated as a floor rather than a
  target.
- The coverage claim on the practices page is **mutation-checked**: break
  something the tests supposedly pin and watch the check fail, then confirm the
  mutant actually mutated. (SPEC-073's coverage sentence was wrong four times;
  SPEC-078's H9 was green, had teeth, and still pinned the wrong proposition.)
- `brag memory`'s header no longer describes its candidate pool using the same
  word its five sibling exporters use for entries-in-scope.

## Scope

### In scope

1. **`golangci-lint`** — config, per-linter rationale, CI gate.
2. **Coverage** — `-cover` in CI, a reported number, a badge, an honest floor.
3. **The `Entries:` envelope inconsistency** — six exporters emit that header
   line; five use `len(entries)`, and `memory` alone uses `result.Candidates`,
   the pool capped at `PoolLimit=200`. On a 368-entry corpus the header reads
   `Entries: 200`, which a reader takes for the corpus size. Changing it moves
   byte-exact goldens **and MCP resource output** (the resources are
   byte-identical to `brag memory`), so it is an envelope decision, not a
   label edit. This is now the stage's **only** correctness item.

### Explicitly out of scope

- **Benchmarks and any scale/perf or concurrency harness.** Deferred to
  **PROJ-009** with the crustyimg methodology harvest that it depends on. This
  keeps PROJ-007 short and leaves room for what comes up.
- Heavy performance optimization — never this project's game.
- Raising coverage by writing tests to hit a number. Measure what is there; if
  the number is embarrassing, say so on the practices page rather than gaming it.
- Any feature work.

## Spec Backlog

Ordered list of specs composing this stage. IDs assigned at creation.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [ ] (not yet framed) — **lint + coverage in CI.** golangci-lint config with
      per-linter rationale, `-cover`, badge, measured floor; the coverage claim
      mutation-checked before it is written down.
- [ ] (not yet framed) — **the `Entries:` envelope semantics.** One decision,
      one fix, one regression test. Narrowed 2026-08-18 from "the three coupled
      defects" after the other two were measured and deferred.

**Count:** 0 shipped / 0 active / 2 pending

## Design Notes

- **`no-sql-in-cli-layer` has no automated guard for the package it was
  written for.** Surfaced at SPEC-080 verify/punch-list-build
  (2026-08-20, P1; costed correctly at the re-verify return trip, R2):
  `internal/mcpserver/import_audit_test.go`'s `TestNoSQLImport` enforces
  the boundary for `internal/mcpserver`, but the constraint's path glob
  (`guidance/constraints.yaml`) covers `internal/cli/**`, which has no
  equivalent test. Verified empirically — adding `_ "database/sql"` to
  `internal/cli/root.go` still passes `go build`, `go vet`, `gofmt -l .`,
  `just test`, and `just test-docs`.

  **Two candidate mechanisms, and they do not close the same thing.** A
  `depguard` golangci-lint rule scoped to `internal/cli/**` (natural fit
  alongside this stage's lint work), or a ported `TestNoSQLImport`.
  Measured 2026-08-20:

  - **A straight port of `TestNoSQLImport` closes half the constraint.**
    It greps the literal string `"database/sql"` and nothing else
    (`import_audit_test.go:27`), so a driver-only import walks past it.
    `internal/cli/story_test.go` is exactly that case — it imports
    `modernc.org/sqlite` and never `database/sql`.
  - **A depguard rule written to the constraint trips five files, not
    four.** The constraint reads *"must not import `database/sql` **or any
    SQL driver**."* `database/sql`: `coverage_test.go`, `impact_test.go`,
    `project_test.go`, `wrapped_test.go`. `modernc.org/sqlite`:
    `coverage_test.go`, `impact_test.go`, `story_test.go`,
    `wrapped_test.go`. Union: **five** — the four above plus
    `story_test.go`.

  **Decide the scope question first — and note which way the text already
  points.** The constraint's text is unqualified by production-vs-test, so
  on its literal reading **five test files violate a blocking constraint
  today**. Amending it to say production-only, or accepting the violations
  and reworking those tests through `storagetest`, is a decision a spec has
  to make; do not change `guidance/constraints.yaml` as a side effect of
  wiring a mechanism that assumes an answer.

  **A content grep must match import lines, not file text.** The two package
  comments that *describe* the boundary contain the literal strings, so a naive
  whole-file grep false-positives on them: `internal/cli/root.go:8` is the only
  non-`_test.go` hit for `database/sql` outside `internal/storage/`, and it is
  a comment. `TestNoSQLImport` sidesteps this by excluding itself by filename
  (`import_audit_test.go:20`) — a port would need the same care, and a
  `depguard` rule avoids the problem entirely by working on the import graph
  rather than on text. Measured 2026-08-20:
  `grep -rn 'database/sql\|modernc.org/sqlite' --include='*.go' . |
  grep -v '_test.go' | grep -v '^./internal/storage/'` returns exactly that one
  comment line.

  **Stale comments to fix alongside.** Two comments assert the boundary
  more broadly than it holds: `internal/cli/list_test.go:275` (*"CLI tests
  cannot import database/sql per the no-sql-in-cli-layer constraint"* —
  the four files above disprove it), and `internal/storage/store.go`'s
  `Store` **type** comment (*"no other package imports a SQL driver"* —
  `internal/storage/storagetest` does, as do the `internal/cli` test
  files; unchanged since SPEC-002, and outside SPEC-080's scope, whose
  Outputs permit only the package doc comment to change in that file). The
  **package** comments on `internal/storage/store.go` and
  `internal/cli/root.go` were scoped to production code at SPEC-080's
  re-verify return trip and are correct as written — they are the wording
  a mechanism should be checked against — with one edge worth knowing before
  the wording is copied (SPEC-080 second re-verify, 2026-08-20).
  `store.go`'s leading claim uses the noun **layer** (*"the only layer whose
  production code imports a SQL driver"*) and holds under both readings of
  "production code": as *non-`_test.go` file* and as *code that ships*. Its
  later claim uses the noun **package** (*"Every other package's production
  code reaches the database only through a `*Store`"*) and holds only under
  the second — `go list -deps ./cmd/brag` contains no `storagetest`, so
  nothing that ships bypasses `*Store`, but `storagetest.go` is a
  non-`_test.go` file in another package that calls `sql.Open` and `db.Exec`
  directly. It is the same package this bullet cites two sentences up. A
  mechanism written to the **layer** noun (or to `internal/storage/**`) has
  no such edge; one written to the **package** noun would flag
  `storagetest`. Non-blocking there and here: the comment names
  `storagetest` as its own exception one sentence earlier, so no reader is
  misled — but the two nouns are not interchangeable and a lint rule should
  not treat them as if they were.

- **The `decisions/` totality gap — deferred at SPEC-080, routed here.**
  `scripts/inventory.sh` emits two rows over `decisions/DEC-*.md`:
  *Decision records* (`insight.type: decision`) and *Decision numbers
  reserved, not yet decided* (`insight.type: reservation`). **Nothing
  asserts the two rows add up to the number of `DEC-*.md` files on disk.**
  A file carrying a third `insight.type`, or none at all, is counted by
  neither row and vanishes from the inventory silently — while
  `docs/engineering-practices.md` still round-trips byte-identical, because
  the page is generated from the same two filters.
  - **No current victim.** The sum is exact today (45 + 1 = 46, matching
    `ls decisions/DEC-*.md | wc -l`), and every existing file's
    `insight.type` was checked at SPEC-080 design and build. The gap needs
    a *new* `DEC-*.md` to bite.
  - **Why deferred rather than written.** The assertion cannot be added
    without first deciding what a violation *means*: an untyped or
    unknown-typed `DEC-*.md` is either a typo that should fail the build,
    or a legitimately new category `inventory.sh` has not been taught yet.
    Those want different assertion shapes (hard fail vs. a "teach
    `inventory.sh` first" signal), so it is a design decision, not a chore
    — which is why SPEC-080's punch-list return trip declined to make it by
    implication.
  - **What would resolve it.** Pick the meaning, then add one
    `scripts/test-docs.sh` assertion that `decs + decs_reserved` equals
    `ls decisions/DEC-*.md | wc -l`. `decisions/_template.md` is
    deliberately outside every count (it does not match `DEC-*.md`) and
    must stay outside this one.
  - **Trigger.** The next spec that adds a file to `decisions/`, or this
    stage's lint-and-CI spec — whichever comes first.
- **The two deferred items keep their working.** `MergeTags` position density and
  `$EDITOR` quoting were measured during PROJ-007 framing and deferred
  2026-08-18; the drafted rules, the rejected alternatives, and the measurements
  behind them live in `projects/PROJ-001-mvp/backlog.md`. Do not re-derive them
  if either is ever picked up. Their coupled question,
  `tag-ordering-projection`, is **already answered** — 260 of 278 tagged entries
  are not in name-ASC order, so `position` stays.
- **Pick linters deliberately.** A default `golangci-lint` enable-all produces a
  wall of findings that gets silenced with `//nolint`, which is worse than no
  lint. Enable a small set, justify each, and let the config itself be
  legible — it is going to be read by the same person the practices page is for.
- **The coverage number is not the deliverable; the honest floor is.** A repo
  with 78 test files against 69 source files does not need to defend a
  percentage. State the number, set the floor below it, and explain in the
  practices page what the tests actually pin — which STAGE-021 will have already
  worked out.

## Dependencies

### Depends on
- **STAGE-021** — the practices page establishes what the coverage number is
  presented against, and what the test regime genuinely guarantees.
- Prior art, not a blocker: crustyimg already wired lint and coverage. Read its
  config as reference when framing the first spec. The **formal methodology
  harvest** (harness structure, perf approach) travels with the deferred
  benchmark work to PROJ-009.

### Enables
- **PROJ-009** — a scale/perf and concurrency baseline, on a substrate whose
  quality is already measured and whose known-defect list is empty.

## Stage-Level Reflection

*Filled in when status moves to shipped. Run Prompt 1c (Stage Ship) in
FIRST_SESSION_PROMPTS.md to draft this.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <one-line items>

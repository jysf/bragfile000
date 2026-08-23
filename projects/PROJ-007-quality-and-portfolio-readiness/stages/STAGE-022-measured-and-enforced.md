---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-022
  status: active
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

Re-verified 2026-08-20 at stage activation. **Three of the original four gaps
hold; one is closed.** No `golangci-lint` config, no `-cover` anywhere,
`func Benchmark` count **0**. CI runs `gofmt`, `go vet` and `go test` and
nothing else.

**"No surfaced godoc" is closed** — SPEC-080 landed the five missing package
comments, so all **15 of 15** packages now render one under `go doc`. Corrected
here rather than carried, because a stage that opens on a stale premise is the
defect STAGE-021 spent three specs learning to catch.

**The coverage number this stage must present honestly, measured today:**
**83.5% of statements** module-wide. Three packages sit at 100%
(`internal/timewindow`, `internal/spark`, `internal/ftsquery`); the floor is
`cmd/brag` at **23.7%**, which is the entrypoint and mostly `AddCommand`
wiring. ~~**Zero packages have no test files.**~~ **Corrected at SPEC-082
design, 2026-08-21: it is 1 of 15.** `go test ./...` prints
`? internal/storage/storagetest [no test files]`. Fourteen of the fifteen
packages carry at least one `_test.go`; the exception is the test-helper
package SPEC-007 created so CLI tests could backdate rows without importing
`database/sql`, and it is exercised only through its callers. The honest form
is *every package that ships behaviour has tests; the one that does not is a
test helper* — which is a better claim than the round zero and survives
someone running the command. Corrected in place rather than carried, for the
reason given two paragraphs up. These are the figures the floor argument has
to be built from — not a target to move toward.

**`cmd/brag`'s 23.7% is one function**, re-measured at SPEC-082 design:
`go tool cover -func` lists `main` at 0.0% and `resolveVersion` at 100.0%.
The package is **38 statements** of ~3,900 — 1.0% of the module — so its
entire shortfall is `main`'s **29 uncovered statements, 0.74% of the module**.
Taking it to 100% would move the headline figure by about seven tenths of a
point and would require a test that invokes `main()`.

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

- [x] SPEC-082 (shipped 2026-08-21) — **lint + coverage in CI.** Framed 2026-08-20,
      **GO at complexity M**. Measured: **83.5%** module-wide statement
      coverage, **zero** packages without test files, three packages at 100%,
      floor `cmd/brag` at **23.7%** — the entrypoint, mostly `AddCommand`
      wiring whose failure modes are covered one layer down in `internal/cli`.
      That number is **not a problem to fix**: a floor set to force it upward is
      the test-writing-to-hit-a-number this stage forbids. `golangci-lint` is
      not installed locally, so how CI obtains it is a design fork. Four forks
      handed over — which linters and why each; the floor and what it protects;
      whether a badge may depend on an external service given this repo's
      no-network identity; and whether `depguard` for `no-sql-in-cli-layer`
      belongs here at all, given the unresolved production-vs-test scope
      question underneath it.
      **Designed 2026-08-21.** All four forks settled with rejected
      alternatives, every literal run through its tool at design, and the
      enforcement mutation-checked five times. Re-measurement held on coverage
      (83.5% module-wide; 23.7% / 75.4% / 78.8% / 82.4%; three packages at
      100%) and on fork 4's evidence (depguard confirms **8 import lines
      across exactly 5 files**, from the import graph rather than grep — and
      `list_test.go`, which a whole-file grep hits, correctly does not appear).
      **One framing figure was wrong:** *zero packages without test files* is
      **1 of 15** — `internal/storage/storagetest`, corrected in *Why Now*
      above. **And golangci-lint's own `uniq-by-line: true` default was hiding
      43% of the evidence** — `default: all` reports 8,713 issues with it and
      **15,300** without; `errorlint` reads as 2 and is 7, `rowserrcheck` (3)
      and `sqlclosecheck` (1) read as **zero** and are real. Both are now
      enabled; the shipped config turns the default off.
      **Fork 1:** nine linters, `default: none`, each argued in `.golangci.yml`
      itself, with the measured count for every rejected one recorded —
      `gosec` 65, `modernize` 50, `revive` 16, `perfsprint` 14, `unparam` 10,
      down to twenty candidates at zero that are still not enabled. The gate
      goes green by **fixing all 19 findings, not excluding them** (two are
      genuine production defects: a `!= io.EOF` comparison that a wrapped error
      slips past, and a `%v` that drops the error chain — the first things
      `errors-wrap-with-context` has caught since it was written in April).
      **Fork 2:** module-wide **80.0%**, in one place (`scripts/coverage.sh`),
      with the unit named — `go tool cover -func`, no `-coverpkg`, because
      `-coverpkg=./...` reports **86.2%** on the identical suite. Per-package
      floors rejected on `cmd/brag`. **Fork 3:** **no badge and no service** —
      argued from *this page's* derived-not-typed rule, explicitly **not** from
      `SECURITY.md`, whose no-network claim is scoped to the binary and would
      survive Codecov intact. **Fork 4:** taken on, production files only
      (`!$test`); `guidance/constraints.yaml` is **not** amended and the scope
      question is logged in `guidance/questions.yaml` with both candidate
      resolutions costed. Also claims this stage's routed `decisions/` totality
      gap (`Z7`, hard-fail, mutation-checked), and moves `X7` from lines to
      words — SPEC-081's own stated trigger, fired by this spec's page edit
      landing at **301 lines, one over the ceiling**. Confidence 0.88.
      **Shipped 2026-08-21 (#178).** Built in one session: 19 lint findings
      fixed and none excluded (zero `//nolint` in the tree), 177 → **184**
      assertion ids, `golangci-lint run` at **0 issues**, coverage **83.5%**
      against an enforced **80.0%** floor, all five mutations fired with each
      mutant confirmed present before its failure was credited. All nine
      literals transcribed byte-exact — the practices page landed at **301
      lines / 2,468 words**, matching design's prediction to the word. **One
      deviation:** `Y4` also pinned the two question-register counts as
      literals (18/6 → 19/7); design routed the move through `X3` and the
      inventory block and never enumerated the second cacher. Re-pinned, not
      widened. Two package comments were **falsified by this spec** and routed
      rather than fixed — `root.go:9` and `store.go:13` both still claim the
      boundary is held "by convention and review, not an automated test" — held
      because the correct rewording depends on LD7's deferred scope question.
- [x] SPEC-083 (shipped on 2026-08-22) — **`no-sql-in-cli-layer` binds production code.**
      Framed 2026-08-21, **GO at complexity S**. Answers the
      `no-sql-in-cli-layer-test-scope` question SPEC-082 logged rather than
      resolved (LD7), on the user's decision that the constraint binds
      production code and not test code. Re-measured at framing from the import
      graph: **5 of the 30 `*_test.go` files** under `internal/cli/` carry
      **8 offending import lines** — and what they *do* with SQL is the argument.
      Four backdate fixtures (`UPDATE entries` / `UPDATE projects`); the fifth,
      `story_test.go`, does nothing at all — its blank driver import is **dead**,
      confirmed at framing by deleting it and watching `go vet` stay clean and
      the package's tests pass. **None uses SQL for persistence**, which is the
      clause the rule is about. Three artifacts already assume this reading
      (depguard's `!$test`, `root.go`'s package comment, and the existence of
      `internal/storage/storagetest`) and only the rule text disagrees.
      Ships `DEC-047` — **the first amendment to `guidance/constraints.yaml` in
      the repo's history**, unchanged since SPEC-001 — plus the rule text, the
      question's `answered` status, and the three stale comments the ambiguity
      was blocking, including `list_test.go:275`, which is wrong under *both*
      readings. Accepted cost recorded, not hidden: `storagetest`'s backdating
      gaps stay unfixed, so CLI tests will keep reaching for `sql.Open`.
      `severity: blocking` is unchanged and the gate is not weakened —
      M-A/M-B must still fire. Confidence 0.92.
      **Shipped 2026-08-22 (#183).** `DEC-047` minted — the first amendment to
      `guidance/constraints.yaml` in the repo's history. The gate was *proved*
      untouched, not asserted: `severity` / `paths` / `added_by` / `added_at`
      byte-identical to main (the whole diff is **two lines**, `rule:` and
      `rationale:`), `golangci-lint run` still **0 issues**, both depguard
      mutations still fire. Five comments repaired, `story_test.go`'s dead
      import deleted, the other four files untouched — LD5 restraint held. New
      Group `AA`, of which `AA1` is a standing anti-erosion guard that fails if
      `severity` ever leaves `blocking`. 184 → **187** assertion ids, 45 → **46**
      decisions, 7 → **6** open questions, all derived and pasted. All seven
      mutations re-run at verify with every mutant hash-confirmed and every
      restore returned to its exact pre-mutation hash; `root.go`'s `7145bf1e…`
      matched across design, build **and** verify, proving nothing drifted
      across three sessions. **One build incident, self-reported and caught by
      the protocol:** an M-F restore used `git checkout` instead of the `/tmp`
      backup and discarded an uncommitted comment edit; the hash check caught
      it and it was reapplied. Verify confirmed completeness independently —
      all 16 fenced literals present, byte-identical, exactly once each.
      **Designed 2026-08-22.** All five forks settled with rejected
      alternatives; every literal written into a throwaway worktree, run
      through its tool, and verified back — **16 of 16 embedded literals are
      byte-verbatim** against the tree that passed the battery. Framing's three
      numbers held on re-derivation from the import graph (**5 of 30 files,
      8 import lines**, 29 production files); after the spec it is **4 files,
      7 lines**, cross-validated by removing `!$test` and reading depguard's
      own count. **Fork 1:** `rule:` carries the scope, `paths:` does not move
      (the glob syntax has no negation, and `paths:` names *territory* while
      the rule names *compliance*); the amendment convention is set here —
      change `rule:`, cite the DEC in `rationale:`, leave `added_at` alone.
      **Fork 2:** four triggers (T1 volume=6, T2 kind, T3 subpackage shape,
      T4 a real frontend), each with its re-derivation command; T1 and T2
      resolve differently on purpose. **Fork 3:** **five** comments, not
      three — `storagetest.go`'s reason clause (falsified by the decision) and
      `store.go`'s `Store` **type** comment (falsified by the measurement,
      routed by this stage's own Design Notes) are named scope additions, not
      quiet ones. **Fork 4:** `story_test.go` only; the other three blank
      driver imports are redundant-but-accurate and stay. **Fork 5:** no
      assertion on the test half — both candidates are wrong, one by
      contradiction and one by firing on improvement — and three assertions on
      the *amendment* instead (`AA1` pins `severity: blocking`; 184 → **187**
      ids).
      **Two things the pre-flight caught that reading would not have.**
      (1) `Y3` independently pins the decision count that `X3` also pins, and
      the §9 half-(b) grep missed it because searching for the *concept* finds
      `Y4` while only searching for the *value* finds `Y3` — its worked example
      even uses `46` as the wrong answer, a number this spec makes right.
      (2) A locked decision was **killed**: design had added a driver import to
      `project_test.go` to fix a cross-file free-ride that does not exist —
      `internal/storage`'s production code registers the driver, proven by
      stripping all four blank imports and watching the one test that calls
      `sql.Open` pass. Seven mutations, each with the mutant confirmed present
      by content hash: M-A/M-B still fail the lint gate, M-C…M-G fire the new
      assertions. Confidence 0.94.
- [ ] (not yet framed) — **the `Entries:` envelope semantics.** One decision,
      one fix, one regression test. Narrowed 2026-08-18 from "the three coupled
      defects" after the other two were measured and deferred.

**Count:** 2 shipped / 0 active / 1 pending

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

  **DEC-047 made this one decidable (SPEC-083 verify, 2026-08-22).** The
  package-vs-layer noun above was an unresolvable wording debate while
  "production code" had no written definition. DEC-047 supplies one — *every
  `*.go` except `*_test.go`* — and under it `store.go:11`'s *"Every other
  package's production code reaches the database only through a `*Store`"* is
  plainly **false**: `storagetest.go` is a non-test file in another package
  that reaches the DB through `sql.Open`. The claim is defensible in context
  (the same comment carves out `storagetest` twice) and SPEC-083 deliberately
  did not touch it, but the item's status has changed from *"stale, needs a
  judgment call"* to *"stale, fix is now mechanical."*

- **A numeral SPEC-083 introduced, with no guard and a trigger that is not one**
  (SPEC-083 verify, 2026-08-22). `internal/cli/root.go:13` and
  `.golangci.yml:71` both now say **four** CLI test files open a database
  directly. True on 2026-08-22 — exactly `coverage_test.go`, `impact_test.go`,
  `project_test.go`, `wrapped_test.go` call `sql.Open` — and unguarded: `AA3`
  is NOT-contains only, so it proves the old false phrase is gone and can never
  prove the new sentence true.

  **`DEC-047`'s T1 trigger is not the guard, and should not be made into one.**
  T1 fires at **six** and asks an economic question — when does writing
  `storagetest.BackdateUpdated` / `BackdateProject` cost less than maintaining
  the copies. "Is this comment accurate?" is a different question that breaks at
  **five**. Lowering T1 to five would damage a working trigger to patch an
  unrelated hole. An assertion pinning *exactly four* is separately foreclosed
  by LD6's own reasoning: it would fire the day someone ports those tests to
  `storagetest`, which is the outcome the triggers exist to encourage.

  **The fix is to drop the numeral** (*"some do"*) from both comments and leave
  T1 alone. Not done at verify because both are literal artifacts of a
  literal-artifact spec, and an in-transit prose edit would leave the tree
  disagreeing with the spec's own §4 literal with no record of why. Belongs to
  whichever spec next opens those files.

  **Two more, created by SPEC-082 itself (build, 2026-08-21).** The package
  comments called *correct as written* above are still correctly **scoped** to
  production code — that clause is untouched and remains the wording a
  mechanism should be checked against. What SPEC-082 falsified is the
  *enforcement* clause each one also carries. `internal/cli/root.go:9` says the
  boundary is *"held today by convention and review, not an automated test
  (`internal/mcpserver` has one, `TestNoSQLImport`; `internal/cli` — the package
  the constraint's path glob actually covers — does not; see STAGE-022)"*, and
  `internal/storage/store.go:13` makes the same claim, *"held today by
  convention and review rather than an automated test (see STAGE-022)"*. Both
  point forward at **this stage**, and this stage's first spec is the automated
  test they say does not exist: `depguard` now fails the `lint` job on either
  import, mutation-checked both halves (M-A, M-B).

  Neither was fixed in SPEC-082's build, for the reason that spec gives for
  `list_test.go:275`: the correct rewording depends on **LD7's deferred
  production-vs-test scope question**, and a comment rewritten before that
  answer lands would have to be rewritten again. All three now share one
  resolve condition. Note what does *not* catch these — `Y1` asserts only that
  a package doc comment **exists**, not what it claims, so `just test-docs`
  stays green with both sentences false. That is the stage's own honest limit
  restated at file scope: **the guards reach numbers, not claims.**

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
  - **CLAIMED by SPEC-082 (design, 2026-08-21) as assertion `Z7`.** The meaning
    was decided rather than deferred again: a **hard fail**, because the two
    readings of an untyped `DEC-*.md` differ in intent but not in consequence —
    either way the file is counted by neither row and the page under-reports —
    and this harness has no warning tier. The failure message names both
    remedies instead. Mutation-checked at design: a `DEC-999` carrying
    `insight.type: bogus` left `scripts/inventory.sh` printing `45` and `1`
    with **47** files on disk, so `X3` stayed green and only `Z7` caught it.

- **A pre-existing goreleaser deprecation, surfaced at SPEC-082's design
  pre-flight and deliberately left alone.** `goreleaser check` (2.17.1) fails
  with `DEPRECATED: brews should not be used anymore` — **before and after**
  SPEC-082's ldflags edit, so it is not that spec's doing. Per AGENTS.md §4
  this is exactly the class that must not be treated as hygiene: the v0.1.0
  formula→cask switch was a four-line deprecation fix that silently changed the
  artifact's OS-trust path and produced the whole signing/Gatekeeper backlog.
  It wants its own spec, run against `docs/distribution-decisions.md`'s
  clean-host-trust category. Named here so it is not lost.
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

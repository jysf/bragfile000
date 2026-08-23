---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-047                        # stable, never reused
  type: decision                     # decision | analysis | recommendation | observation
  confidence: 0.92                   # 0.0 - 1.0, honest assessment
  audience:                          # who needs to know?
    - developer
    - agent

agent:
  id: claude-opus-5
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-007                       # the project during which this was decided
repo:
  id: bragfile

created_at: 2026-08-22
supersedes: null
superseded_by: null

tags:
  - constraints
  - architecture-boundary
  - testing
  - lint
  - depguard
---

# DEC-047: `no-sql-in-cli-layer` binds production code, not test code

## Decision

The `no-sql-in-cli-layer` constraint binds **production files under
`internal/cli/` — every `*.go` except `*_test.go`**. Its `severity` stays
`blocking` and its `paths` glob stays `internal/cli/**`; only the rule text
changes, to say what `depguard` already enforces and what three other
artifacts already assumed. Test files in `internal/cli/` may open a database
directly; `internal/storage/storagetest` is the preferred route, by
preference and not by prohibition.

## Context

`guidance/constraints.yaml` has been unchanged since SPEC-001 (#1, 2026-04-19).
This is its first amendment.

The rule text was unqualified — *"Files under internal/cli/ must not import
database/sql or any SQL driver"* — so on its literal reading, **5 of the 30
`*_test.go` files under `internal/cli/`, across 8 import lines, violated a
blocking constraint**. Measured from the import graph (`go/parser` over the
package, cross-checked against `depguard` with `!$test` removed) on 2026-08-21
and re-derived at design 2026-08-22:

| File | Imports | What it does with SQL |
|---|---|---|
| `coverage_test.go` | `database/sql`, `_ modernc.org/sqlite` | `UPDATE entries` |
| `impact_test.go` | `database/sql`, `_ modernc.org/sqlite` | `UPDATE entries` |
| `project_test.go` | `database/sql` | `UPDATE projects` |
| `story_test.go` | `_ modernc.org/sqlite` | **nothing** |
| `wrapped_test.go` | `database/sql`, `_ modernc.org/sqlite` | `UPDATE entries` |

`internal/cli/list_test.go` is **not** in that table. A whole-file grep hits
it; the import graph does not, because line 275 is a comment. That is the
trap this constraint has set twice, and it is why every count here comes from
the import graph.

Three artifacts already assumed the production-only reading, and only the rule
text disagreed with them:

1. `.golangci.yml`'s `depguard` rule ships with `!$test` (SPEC-082 LD7).
2. `internal/cli/root.go`'s package comment says *"Its **production code**
   imports no SQL driver"* (scoped at SPEC-080).
3. `internal/storage/storagetest` exists so CLI tests need not import SQL.

SPEC-082 mechanised the production half and deliberately declined to amend the
rule, on the principle that a spec must not amend the constraint it is
mechanising. It logged `no-sql-in-cli-layer-test-scope` with both resolutions
costed. This record is that question's answer.

**Who decided:** the user, 2026-08-21, at SPEC-083 framing. Framing's job was
to check the decision, not to relitigate it — and the code supports it more
strongly than the principle does. Four of the five files use raw SQL as a time
machine to age a fixture that `Store.Add` stamps with `time.Now()`; the fifth
is litter. **None uses SQL for persistence**, which is the clause the rule is
actually about. The rationale points the same way: *"future frontends (TUI,
API) feasible"* — a TUI reusing `internal/cli` links its production code, and
`_test.go` files never compile into a dependent.

## Alternatives Considered

- **Option A (chosen): amend the rule text to production files.**
  - What it is: `rule:` gains *"Production files under internal/cli/ — every
    `*.go` except `*_test.go`"*; `severity` and `paths` unchanged; `rationale`
    cites this record.
  - Why selected: it makes the rule agree with the mechanism that enforces it
    and with the two package comments and the sub-package that already assume
    it. It is the cheapest of the two, and the only one that leaves a reader
    of `constraints.yaml` with a true statement.

- **Option B: rework the five test files through `storagetest` and enforce the
  rule as literally written.**
  - What it is: `story_test.go` needs only its dead import deleted; the other
    four need two new helpers — `entries.updated_at` (today's `Backdate`
    covers `created_at` only) and a `projects`-table backdate. Then `depguard`
    drops `!$test` and the constraint becomes true as written.
  - Why rejected: it enforces a boundary that does not exist. The rule protects
    a *linkage* property — what a future frontend pulls in when it imports
    `internal/cli` — and a `_test.go` file is never in that graph. Paying two
    helpers and five file rewrites to defend a property that cannot be violated
    is ceremony. It also mistakes the symptom for the rule: what would actually
    matter is a CLI test using SQL for persistence or querying, and none does.
    Kept as the shape the revisit triggers below would restore.

- **Option C: leave the rule text alone and let the mechanism be narrower than
  the rule.**
  - What it is: what main had on 2026-08-21 — a blocking rule with five known
    violators, none of which anyone intends to fix.
  - Why rejected: a blocking constraint with standing, tolerated violations
    teaches every future reader that blocking constraints are advisory. That is
    a more expensive loss than either amendment.

- **Option D: express the scope in `paths:` instead of the rule text.**
  - What it is: narrow `paths: ["internal/cli/**"]` to something production-only.
  - Why rejected: the glob syntax used in this file has no negation — the one
    test-scoped constraint in the file (`storage-tests-use-tempdir`) narrows
    *toward* tests with a positive `**_test.go` suffix, which cannot be
    inverted. There is also a division of labour worth keeping: `paths:` names
    the territory a spec must check when it touches it; the rule text names
    which files inside that territory must comply. Widening the glob's job to
    carry scope would make every constraint's `paths:` ambiguous.

- **Option E: add `amended_at` / `amended_by` fields to the constraint entry.**
  - What it is: new front-matter keys recording that the text moved.
  - Why rejected: the file's header block declares six required fields, one of
    which — `rationale`, documented as *"why (link to DEC-XXX preferred)"* — is
    already the sanctioned place for exactly this. An optional seventh field
    carried by one constraint out of eleven is a field that rots. **The
    convention this record sets for future amendments:** change `rule:`, cite
    the `DEC-*` in `rationale:` with the date, and leave `added_at` at the
    original date.

## Consequences

- **Positive:** the rule text, `depguard`'s `!$test`, `internal/cli/root.go`'s
  package comment and the existence of `internal/storage/storagetest` now all
  say the same thing. The repo stops carrying five standing violations of a
  blocking constraint, without weakening a single check: `golangci-lint run`
  still reports 0, and both mutations (`_ "database/sql"` and
  `_ "modernc.org/sqlite"` in `internal/cli/root.go`) still fail the gate.

- **Negative — the accepted cost, named so it is a choice.**
  `internal/storage/storagetest` exists precisely so CLI tests need not import
  SQL, and production-only scope means its gaps stay unfixed: `Backdate` covers
  `entries.created_at` but not `updated_at`, and nothing covers the `projects`
  table. So CLI tests will keep reaching for `sql.Open` to age a row, and each
  one is another copy of storage's schema living outside `internal/storage`.
  This is accepted, not overlooked. It is bounded by the triggers below.

- **Neutral — a residual gap between the rule and its mechanism, named.**
  `depguard`'s file glob is `**/internal/cli/*.go`, which covers this package
  only; the constraint's `paths:` is `internal/cli/**`, which covers
  subpackages too. `internal/cli` has no subdirectories today (verified at
  design), so the two agree exactly. A subpackage would open a gap — see
  trigger T3.

- **Neutral — the blank driver imports in those test files are redundant,
  measured at design.** `internal/storage`'s *production* code imports
  `_ "modernc.org/sqlite"` (`store.go:29`), so every package that imports
  `internal/storage` — which is every CLI test — already has the driver
  registered. Verified by stripping all four blank driver imports from
  `internal/cli/*_test.go` and running `TestProjectStatus_OrderedByRecency`,
  the test that calls `sql.Open("sqlite", …)`: it passes. Only
  `story_test.go`'s is deleted here, because that file contains no SQL at
  all; the other three sit beside a live `sql.Open` naming the driver it
  uses, which is ordinary Go and is left alone.

- **Neutral:** `internal/mcpserver`'s `TestNoSQLImport` walks *all* `.go` files
  including tests, so that package holds itself to a stricter line than this
  constraint requires. That is a package's own convention, left alone
  deliberately; it is not evidence about `internal/cli`'s scope.

## Validation

This decision was right if the boundary keeps holding where it matters — no
SQL in the CLI's production code — while nobody spends a cycle rewriting test
fixtures to satisfy a property that cannot be violated. `golangci-lint run`
reporting 0 with both mutations still failing is the standing evidence.

**Revisit when any of these fires.** Each is mechanically checkable; the
command that re-derives it is given.

- **T1 — volume.** The number of `*_test.go` files under `internal/cli/` that
  import `database/sql` or a SQL driver reaches **six**. It is **four** as of
  2026-08-22 (`coverage_test.go`, `impact_test.go`, `project_test.go`,
  `wrapped_test.go`). Two more copies is where writing
  `storagetest.BackdateUpdated` and `storagetest.BackdateProject` costs less
  than maintaining the copies, and the answer is then to build those helpers —
  not to re-widen the rule.
  Re-derive: `golangci-lint run` with `!$test` removed from `.golangci.yml`,
  or the `go/parser` sweep recorded in SPEC-083's §12(b) section.

- **T2 — kind.** Any CLI test uses raw SQL for something other than backdating
  a fixture: inserting rows, reading rows back to assert on them, or poking the
  schema. That is SQL used as persistence or as a query surface in the CLI
  layer, which *is* the clause the rule is about, and it reopens the scope
  question itself rather than the helper question.
  Re-derive: `grep -n 'db\.\(Exec\|Query\|QueryRow\)' internal/cli/*_test.go`
  and read each hit — every one today is an `UPDATE` against `created_at` /
  `updated_at`.

- **T3 — shape.** A subpackage appears under `internal/cli/`. `depguard`'s
  `**/internal/cli/*.go` glob would not cover it while the constraint's
  `paths:` would, so the mechanism must be widened to
  `**/internal/cli/**/*.go` in the same PR that adds the subpackage.
  Re-derive: `ls -d internal/cli/*/` — empty on 2026-08-22.

- **T4 — premise.** A frontend that links `internal/cli` actually appears (a
  TUI, an API server). The reasoning above rests on `_test.go` files never
  entering a dependent's import graph; a real dependent is the moment to
  confirm that held rather than to assume it.

## References

- Related specs: SPEC-083 (this amendment), SPEC-082 (wired `depguard`, filed
  the question, declined to amend the rule), SPEC-080 (measured the gap
  empirically and routed it), SPEC-007 (the verify punch list that first hit
  this constraint, 2026-04-20)
- Related decisions: DEC-001 (pure-Go SQLite driver — the deny list names it)
- Related constraints: `no-sql-in-cli-layer` (the subject)
- Register: `guidance/questions.yaml` → `no-sql-in-cli-layer-test-scope`,
  answered by this record
- Enforcement: `.golangci.yml` → `linters.settings.depguard.rules.no-sql-in-cli-layer`

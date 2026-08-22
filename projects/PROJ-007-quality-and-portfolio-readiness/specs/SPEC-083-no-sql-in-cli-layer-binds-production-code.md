---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-083
  type: chore                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-007
  stage: STAGE-022
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-21
  framed_at: 2026-08-21

insight:
  confidence: 0.92

references:
  decisions:
    - DEC-001                      # pure-Go sqlite driver — the deny list names it
  constraints:
    - no-sql-in-cli-layer          # THE SUBJECT — this spec amends its rule text
    - one-spec-per-pr
  related_specs:
    - SPEC-007                     # the verify punch list that first hit this constraint
    - SPEC-080                     # measured the gap and routed it
    - SPEC-082                     # mechanised the production half; refused to amend the rule
---

# SPEC-083: `no-sql-in-cli-layer` binds production code

> **Cycle: frame.** GO at complexity **S**. The decision itself is already
> made — see *The decision, and who made it*. This spec records it and makes
> four artifacts agree with it.

## Context

`guidance/constraints.yaml`'s `no-sql-in-cli-layer` is **blocking** and its
rule text is unqualified:

> *Files under internal/cli/ must not import database/sql or any SQL driver.
> All persistence goes through internal/storage.*

On that literal reading, **5 of the 30 `*_test.go` files under `internal/cli/`
violate a blocking constraint today**, across **8 import lines** — re-measured
from the import graph at framing (2026-08-21), not from grep:

| File | Imports |
|---|---|
| `coverage_test.go` | `database/sql`, `_ modernc.org/sqlite` |
| `impact_test.go` | `database/sql`, `_ modernc.org/sqlite` |
| `project_test.go` | `database/sql` |
| `story_test.go` | `_ modernc.org/sqlite` |
| `wrapped_test.go` | `database/sql`, `_ modernc.org/sqlite` |

Note `internal/cli/list_test.go` is **not** among them — a whole-file grep hits
it, but line 275 is a comment. That is the trap this constraint has set twice.

**Three artifacts already assume the production-only reading**, and only the
rule text disagrees with them:

1. `.golangci.yml`'s depguard rule ships with `!$test` (SPEC-082, LD7).
2. `internal/cli/root.go`'s package comment says *"Its **production code**
   imports no SQL driver and no `database/sql`"* — scoped at SPEC-080.
3. `internal/storage/storagetest` exists so CLI tests can reach raw SQL
   without importing it.

SPEC-082 mechanised the production half and **deliberately refused** to amend
the rule, on the principle that a spec must not amend the constraint it is
mechanising. It logged the question as `no-sql-in-cli-layer-test-scope` with
both resolutions costed. This spec is that question's answer.

### The decision, and who made it

**The user decided it (2026-08-21): the constraint binds production code, not
test code.** Framing's job was to check it, not to relitigate it — and the code
supports it more strongly than the principle does. What those five files
actually *do* with SQL, measured at framing:

- `coverage_test.go`, `impact_test.go`, `wrapped_test.go` — `UPDATE entries`
- `project_test.go` — `UPDATE projects`
- `story_test.go` — **nothing.** The blank driver import is dead: deleting it
  leaves `go vet ./internal/cli/` clean and the package's tests passing
  (confirmed at framing by actually removing it and running both).

Four of the five use raw SQL as a **time machine to age fixtures**; the fifth is
litter. **None uses it for persistence**, which is the clause the rule is
actually about. And the rationale points the same way — *"future frontends
(TUI, API) feasible"* — because a TUI reusing `internal/cli` links its
production code, and `_test.go` files never compile into a dependent.

### What this costs, named so it is a choice

`internal/storage/storagetest` exists precisely so CLI tests need not import
SQL. Deciding production-only means its gaps stay unfixed — `Backdate` covers
`entries.created_at` but not `updated_at`, and nothing covers the projects
table — so future CLI tests will keep reaching for `sql.Open` to backdate a
row. That is an **accepted cost**, not an oversight, and design should record it
as one with a trigger for revisiting.

## Goal

Make the rule text say what the mechanism enforces and what the code already
assumes; record the decision as a `DEC-*`; and fix the three comments the
ambiguity was blocking.

## Inputs

- **Files to read:** `guidance/constraints.yaml` (the rule), `guidance/questions.yaml`
  (`no-sql-in-cli-layer-test-scope`, both resolutions costed), `.golangci.yml`
  (the `!$test` comment), `decisions/_template.md`, the SPEC-082 archived spec
  (LD6/LD7).
- **Related code paths:** `internal/cli/*_test.go`, `internal/storage/storagetest/`.

## Outputs

- **`decisions/DEC-047-*.md`** (new) — the decision record. **This is the first
  amendment to `guidance/constraints.yaml` in the repo's history** (unchanged
  since SPEC-001, #1), which is why it gets a DEC rather than a commit message.
- **`guidance/constraints.yaml`** — `no-sql-in-cli-layer` rule text scoped to
  production files. `severity: blocking` is **unchanged**.
- **`guidance/questions.yaml`** — `no-sql-in-cli-layer-test-scope` → `answered`,
  citing DEC-047.
- **`.golangci.yml`** — the `!$test` comment currently says that half is
  *"open, named, and unguarded"*. It stops being open.
- **Three stale comments**, all routed and all blocked on this answer:
  `internal/cli/list_test.go:275` (wrong under **both** readings),
  `internal/cli/root.go:9` and `internal/storage/store.go:13` (both still say
  the boundary is held *"by convention and review, not an automated test"* —
  falsified by SPEC-082).
- **`internal/cli/story_test.go`** — delete the dead driver import.

## Acceptance Criteria

*Frame-level. Design tightens these into assertions.*

1. `guidance/constraints.yaml`'s rule text is scoped to production files and
   `severity` is still `blocking`.
2. `DEC-047` exists, states the decision, the rejected alternative, and the
   accepted cost with a revisit trigger.
3. `no-sql-in-cli-layer-test-scope` is `answered` and cites DEC-047.
4. The three stale comments are true as written, checked **against the paths
   they cite** — the SPEC-080 lesson: *test the claim, not the counterexample.*
5. `story_test.go`'s driver import is gone; `just test`, `just lint`,
   `gofmt -l .`, `go vet ./...` all clean.
6. `golangci-lint run` still reports **0 issues**, and depguard still fires on
   a production-file import (M-A/M-B still pass — this spec must not weaken
   the gate it is documenting).
7. `just test-docs` **ALL OK**, with the inventory block re-derived and pasted
   if any row moves.

## Forks handed to design

1. **Exact rule wording.** `paths:` is `internal/cli/**` and YAML cannot express
   `!$test`, so the production-only scope has to live in the *rule text*. Decide
   whether `paths:` also changes and whether the rule should name `storagetest`
   as the sanctioned route.
2. **DEC-047's revisit trigger** for the accepted cost above.
3. **The three comments' rewording** — `list_test.go:275` is the delicate one:
   it is wrong under *both* readings, so it needs a new claim, not a patch.
4. **Whether `story_test.go`'s deletion belongs here.** Framing says yes: same
   subject, proven dead, one line. Design should confirm or split it out.
5. **Whether a test should pin any of this.** The rule is now enforced by
   depguard for production; the *test* half is by definition unenforced. Decide
   whether that asymmetry deserves an assertion or a comment.

## Verdict

**GO at complexity S.** The decision is made, the measurements are in hand and
re-derived at framing, and the blast radius is four documentation artifacts plus
one dead import. No production behaviour changes. The one thing that makes it
non-trivial is that it is the repo's first constraint amendment, so the record
matters more than the diff.

## Implementation Context

*Filled in during the design cycle.*

## Notes for the Implementer

*Filled in during the design cycle.*

## Build Completion

*Filled in during the build cycle.*

- **Deviations from the spec:** <none / list>
- **New DEC-* files created:** <none / list>
- **Constraints checked:** <list>
- **Gates:** <list>

### Build-phase reflection (3 questions, short answers)

- **What was unclear in the spec?**
- **What was missing that you had to decide yourself?**
- **What would you do differently?**

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>

4. **What can a user do now that they couldn't before?** — one sentence,
   before → after; quote the confirming number if one exists, name the outcome
   if not. Write `none` if this spec has no user-visible outcome — that is a
   real, greppable result, not a blank. **If this answer is not `none`, capture
   it before closing the cycle.** Evidence ref: tag `pr:<n>`.
   — <answer | none>

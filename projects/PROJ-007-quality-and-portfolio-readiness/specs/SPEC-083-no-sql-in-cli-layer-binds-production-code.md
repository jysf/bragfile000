---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-083
  type: chore                      # epic | story | task | bug | chore
  cycle: design                    # frame | design | build | verify | ship
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
  designed_at: 2026-08-22

insight:
  confidence: 0.94

references:
  decisions:
    - DEC-001                      # pure-Go sqlite driver — the deny list names it
    - DEC-047                      # NEW — this spec's own record
  constraints:
    - no-sql-in-cli-layer          # THE SUBJECT — this spec amends its rule text
    - one-spec-per-pr
  related_specs:
    - SPEC-007                     # the verify punch list that first hit this constraint
    - SPEC-080                     # measured the gap and routed it
    - SPEC-082                     # mechanised the production half; refused to amend the rule
---

# SPEC-083: `no-sql-in-cli-layer` binds production code

> **Cycle: design.** GO at complexity **S**. The decision itself was made by
> the user on 2026-08-21 — see *The decision, and who made it*. Design's job
> was to write it down correctly, make the artifacts agree with it, and prove
> the amendment does not weaken the gate it documents. Every literal below was
> written into a throwaway worktree and run through its own tool; see
> [§12(b)](#12b-design-time-pre-flight--what-each-tool-actually-said).

## Context

`guidance/constraints.yaml`'s `no-sql-in-cli-layer` is **blocking** and its
rule text is unqualified:

> *Files under internal/cli/ must not import database/sql or any SQL driver.
> All persistence goes through internal/storage.*

On that literal reading, **5 of the 30 `*_test.go` files under `internal/cli/`
violate a blocking constraint today**, across **8 import lines** — measured
from the import graph, not from grep.

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
supports it more strongly than the principle does. Four of the five files use
raw SQL as a **time machine to age fixtures**; the fifth is litter. **None uses
it for persistence**, which is the clause the rule is actually about. And the
rationale points the same way — *"future frontends (TUI, API) feasible"* —
because a TUI reusing `internal/cli` links its production code, and `_test.go`
files never compile into a dependent.

### What this costs, named so it is a choice

`internal/storage/storagetest` exists precisely so CLI tests need not import
SQL. Deciding production-only means its gaps stay unfixed — `Backdate` covers
`entries.created_at` but not `updated_at`, and nothing covers the projects
table — so future CLI tests will keep reaching for `sql.Open` to backdate a
row. That is an **accepted cost**, not an oversight. DEC-047 records it as one
and names four triggers that reopen it (LD3).

## Goal

Make the rule text say what the mechanism enforces and what the code already
assumes; record the decision as `DEC-047`; fix the five comments the ambiguity
was blocking; and guard the amendment so it cannot silently widen back or lose
its severity.

## Re-measurement at design (2026-08-22)

Framing's numbers were re-derived from scratch, by a `go/parser` sweep over
`internal/cli/` that reads each file's import list — **not** by grep, and not
by copying framing's table. All three numbers hold.

| File | Imports | What it does with SQL |
|---|---|---|
| `coverage_test.go` | `database/sql`, `_ modernc.org/sqlite` | `UPDATE entries SET created_at, updated_at` |
| `impact_test.go` | `database/sql`, `_ modernc.org/sqlite` | `UPDATE entries SET created_at, updated_at` |
| `project_test.go` | `database/sql` | `UPDATE projects SET updated_at, created_at` |
| `story_test.go` | `_ modernc.org/sqlite` | **nothing** |
| `wrapped_test.go` | `database/sql`, `_ modernc.org/sqlite` | `UPDATE entries SET created_at, updated_at` |

- **30** `*_test.go` files in the package, **29** production `.go` files.
- **5** offending test files, **8** offending import lines.
- `internal/cli/list_test.go` is **not** in the table. A whole-file grep hits
  it; the import graph does not, because line 275 is a comment.
- `internal/cli/` has **no subdirectories** (`ls -d internal/cli/*/` → empty),
  so depguard's `**/internal/cli/*.go` and the constraint's
  `paths: ["internal/cli/**"]` cover exactly the same files today. That
  agreement is contingent, not structural — DEC-047 trigger **T3**.

**After this spec: 4 files, 7 import lines.** Cross-validated by removing
`!$test` from `.golangci.yml` and running `golangci-lint run`, which reported
`depguard: 7` across `coverage_test.go`, `impact_test.go`, `project_test.go`,
`wrapped_test.go` — and, again, not `list_test.go`.

### The thing the re-measurement changed

`story_test.go`'s blank driver import is dead — but so is a claim design
almost wrote about the other three. See LD5 and the §12(b) section: the
pre-flight killed an entire locked decision before it reached build.

## §12(b) design-time pre-flight — what each tool actually said

Every literal this spec embeds was written into a detached `git worktree`, run
through its own tool, and the worktree removed. Nothing in this section is
predicted. Two findings below were **caught here that no amount of reading
would have found**, which is the whole argument for the practice.

| Tool | Version | What was run | Result |
|---|---|---|---|
| `python3 -c "import yaml…"` | PyYAML | parse `guidance/constraints.yaml` | **parses** — 11 constraints; `rule` / `severity: blocking` / `paths: ['internal/cli/**']` read back exactly as written |
| `python3 -c "import yaml…"` | PyYAML | parse `guidance/questions.yaml` | **parses** — 19 questions, **6** open, the subject entry `status: answered` with `answered_by: DEC-047 (SPEC-083) …` |
| `golangci-lint` | **v2.13.1** | `golangci-lint config verify` | **exit 0** |
| `golangci-lint` | v2.13.1 | `golangci-lint run` | **0 issues** |
| `gofmt -l .` | Go 1.26.6 | full tree, all edits staged | **clean** (the `story_test.go` import-group deletion leaves a gofmt-stable block) |
| `go vet ./...` | Go 1.26.6 | full tree | **clean** |
| `go test ./...` | Go 1.26.6 | full tree | **14 packages `ok`**, 1 with no test files (`storagetest`) |
| `./scripts/inventory.sh` | — | with `DEC-047` and Group `AA` staged | decisions **45 → 46**, doc assertions **184 → 187**, open questions **7 → 6**; every other row unchanged |
| `./scripts/test-docs.sh` | — | full harness, all artifacts staged | **ALL OK**, 187 assertion ids |
| literal round-trip | — | every fenced literal in *Notes for the Implementer* searched for verbatim in the pre-flighted tree | **16 of 16 byte-verbatim** — the literal-artifact contract, checked rather than asserted |
| `wc -w docs/engineering-practices.md` | — | after the prose edit + regenerated inventory block | **2,525 words** (was 2,494) — inside `X7`'s 1800..2700 band with 175 words of ceiling headroom |

### The gate is not weakened — two mutations, each with the mutant confirmed present

Confirmed by **content hash**, not `git diff --quiet`: the file's SHA-256 was
taken before and after the mutation, the diff printed, and the hash re-checked
after restore. `git diff --quiet` is blind to untracked files and would have
credited a red for the wrong reason.

| # | Mutation | Hash moved | `golangci-lint run` said | Restored |
|---|---|---|---|---|
| **M-A** | `_ "database/sql"` added to `internal/cli/root.go` | `7145bf1e` → `5a7a73a5` | `internal/cli/root.go:17:2: import 'database/sql' is not allowed from list 'no-sql-in-cli-layer' … (depguard)` — **1 issue** | hash back to `7145bf1e` |
| **M-B** | `_ "modernc.org/sqlite"` in the same place | `7145bf1e` → `c2b8ecad` | `import 'modernc.org/sqlite' is not allowed from list 'no-sql-in-cli-layer' … (depguard)` — **1 issue** | hash back to `7145bf1e` |

Unmutated, immediately after both restores: **0 issues**. The amended rule text
does **not** let either mutation pass.

### The three new assertions have teeth — five mutations, all fired

| Mutation | Assertion that fired |
|---|---|
| `severity: blocking` → `warning` | `FAIL: AA1 … [severity is no longer blocking]` |
| rule text reverted to the unqualified wording | `FAIL: AA1 … [rule text is not scoped to production files]` |
| register entry back to `status: open` | `FAIL: AA2 … [status is not answered]` (and `X3`, `Y4`) |
| `"convention and review"` restored in `root.go` | `FAIL: AA3 … internal/cli/root.go still says "convention and review"` |
| `"no other package imports a"` restored in `store.go` | `FAIL: AA3 … internal/storage/store.go still says "no other package imports a"` |

### Finding 1 — `Y3` pins the decision count, and the §9 grep did not find it

The premise audit correctly predicted that adding `DEC-047` moves
`Decision records | 45 |` and that `X3` (byte-exact inventory block) would
catch it. It **missed that `Y3` independently pins the same value** with a
literal `grep -F -q 'Decision records | 45 |'`. Running the harness at design
found it in one command:

```
FAIL: Y3: inventory.sh row value(s) wrong: decision-records!=45
```

This is AGENTS.md §9 half (b) — *"before writing `## Failing Tests`, grep the
harness for every literal occurrence of any value this spec changes"* — landing
for the second time in two specs, and it is worth naming *how* the grep failed:
searching the harness for the **concept** (`questions`, `open`) found `Y4`;
only searching for the **value** (`45`) finds `Y3`. `Y3`'s own comment also
carries a worked example that uses the number `46` as the *wrong* answer — a
number this spec makes correct — so the fix is a re-pin **plus** a comment
repair, not a one-character edit. Both are in LD7.

### Finding 2 — a locked decision the pre-flight killed

Design had locked an extra edit: give `internal/cli/project_test.go` its own
`_ "modernc.org/sqlite"`, on the reasoning that it calls `sql.Open("sqlite", …)`
without importing the driver and was therefore *free-riding* on the blank
imports in its sibling test files — so deleting `story_test.go`'s would leave a
landmine for whoever ports the rest to `storagetest`.

**That reasoning was wrong, and the pre-flight proved it.** All four blank
driver imports under `internal/cli/*_test.go` were stripped and the one test
that calls `sql.Open` was run by name:

```
=== RUN   TestProjectStatus_OrderedByRecency
--- PASS: TestProjectStatus_OrderedByRecency (0.01s)
ok      github.com/jysf/bragfile000/internal/cli    0.363s
```

`internal/storage`'s **production** code imports `_ "modernc.org/sqlite"`
(`store.go:29`), so registration is already in `internal/cli`'s production
dependency graph (`go list -deps ./internal/cli` matches the driver). There is
no free ride and no landmine: every CLI test imports `internal/storage`.

The invented fix was dropped. `story_test.go`'s import is still deleted — that
file contains **no SQL at all** — and the other three are left alone with the
reasoning recorded (LD5). Recording this here rather than discovering it at
build is the entire value of the practice; the run cost about a minute.

### Finding 3 — a fifth stale comment, and the grep that could not have found it

The premise-audit sweep greps for the constraint **id**.
`internal/storage/store.go`'s `Store` **type** comment — *"All persistence flows
through a Store; no other package imports a SQL driver"* — never names the
constraint, so the sweep walks straight past it. It surfaced from STAGE-022's
own Design Notes, which had analysed and routed it at SPEC-080.

Checking it took the import graph, not a grep. `grep -l 'modernc.org/sqlite'`
over non-test files returns **three**: `store.go`, `storagetest.go`, and
`backup.go` — whose hit is a **comment** (`// supported on modernc.org/sqlite
v1.51.0`). The parser says **two**. That is the third time in this one spec that
a grep over this exact subject produced a plausible wrong number
(`list_test.go:275`, `Y3`'s literal, and now `backup.go`) — which is the whole
reason AGENTS.md §9 says a count produced by grep is a hypothesis.

**A grep for a constraint's id is not a grep for its claims.** The premise audit
catches artifacts that *cite* the rule; it cannot catch artifacts that merely
*assert* what the rule is about. Both classes exist, and only the first is
mechanically enumerable.

---

## Fork 1 — the exact rule wording

**Settled.** `rule:` gains the production scope; `severity:` and `paths:` do
not move; `rationale:` carries the amendment provenance and names
`storagetest` as the sanctioned route.

```yaml
    rule: "Production files under internal/cli/ — every *.go except *_test.go — must not import database/sql or any SQL driver. All persistence goes through internal/storage."
```

Three sub-questions, each decided:

**(a) Does `paths:` change? No.** The glob syntax in this file has no negation.
The one test-scoped constraint in it — `storage-tests-use-tempdir`,
`paths: ["internal/storage/**_test.go"]` — narrows *toward* tests with a
positive suffix, and that cannot be inverted. There is also a division of
labour worth keeping: **`paths:` names the territory a spec must check when it
touches it; the rule text names which files inside that territory must
comply.** Making `paths:` carry scope would make every constraint's glob
ambiguous about which of the two jobs it is doing.

**(b) Does the rule name `storagetest`? Not in `rule:` — in `rationale:`.**
`rule:` is the enforceable sentence; a preference embedded in it reads as
enforced, and this preference is explicitly *not* enforced (fork 5). The file's
header documents `rationale` as *"why (link to DEC-XXX preferred)"*, which is
exactly where a sanctioned-route pointer belongs.

**(c) How is the amendment recorded, given this is the first one?** In
`rationale:`, citing `DEC-047` with the date, and `added_at` stays
`2026-04-19`. **This sets the convention for future amendments** (DEC-047,
Option E): change `rule:`, cite the `DEC-*` in `rationale:`, leave `added_at`
at the original date.

### Rejected alternatives

- **Narrow `paths:` instead of the rule text.** No negation in the glob
  syntax; conflates the two jobs `paths:` does. (a) above.
- **Add `amended_at` / `amended_by` fields.** The header block declares six
  required fields and `rationale` already covers this. An optional seventh
  field carried by one constraint out of eleven is a field that rots. Rejected
  in DEC-047 as Option E.
- **Say "non-test files" instead of naming the suffix.** `*_test.go` is the
  exact thing `depguard`'s `!$test` matches and the exact thing a reader
  greps for. "Non-test" is one indirection further from both.
- **Leave the rule text and let the mechanism stay narrower** (DEC-047 Option
  C). A blocking constraint with five standing, tolerated violations teaches
  every future reader that blocking constraints are advisory.

---

## Fork 2 — DEC-047's revisit trigger

**Settled: four triggers, each mechanically checkable, each shipping the
command that re-derives it.** The accepted cost is that `storagetest`'s gaps
stay unfixed and CLI tests keep hand-rolling `sql.Open` to age a row. A cost
with no trigger is just a regret.

| | Trigger | Value today | Re-derive with |
|---|---|---|---|
| **T1** | **Volume** — CLI test files importing SQL reaches **6** | **4** | `golangci-lint run` with `!$test` removed |
| **T2** | **Kind** — a CLI test uses raw SQL for anything but backdating a fixture (insert, read-back-and-assert, schema poke) | all 4 are `UPDATE`s on `created_at`/`updated_at` | `grep -n 'db\.\(Exec\|Query\|QueryRow\)' internal/cli/*_test.go` |
| **T3** | **Shape** — a subpackage appears under `internal/cli/`, which depguard's `*.go` glob would miss while `paths:` covers it | none | `ls -d internal/cli/*/` |
| **T4** | **Premise** — a frontend that actually links `internal/cli` ships (TUI, API) | none | — |

**T1 is the count trigger and T2 is the kind trigger, and they resolve
differently on purpose.** T1 firing means *build the two missing helpers*
(`BackdateUpdated`, `BackdateProject`) — the cost has outgrown the copies. T2
firing means *reopen the scope question itself*, because SQL used as a query
surface in the CLI layer is the thing the rule is actually about. Conflating
them would answer a scope question with a helper.

### Rejected alternatives

- **A time-based trigger ("revisit at PROJ-008").** Calendar triggers fire when
  nothing has changed and stay silent when everything has.
- **T1 at 5 instead of 6.** 5 is one more copy; 6 is the point where two
  helpers cost less than the copies. Picking the threshold at *today+1* makes
  the trigger fire on the next ordinary test, which is noise.
- **No trigger at all, on the grounds that the cost is small.** The cost is
  small *today*, at four copies. That is precisely the state a trigger exists
  to notice leaving.

---

## Fork 3 — the five comments

**Settled: five, not three.** Framing routed `list_test.go:275`, `root.go:9`
and `store.go:13`. Two more are design-time scope additions, and both are
named here rather than absorbed quietly:

- **`internal/storage/storagetest/storagetest.go:1`** gives the constraint as
  the *reason* CLI tests route through the sub-package — which stops being a
  reason the moment the constraint is scoped to production files. Falsified by
  **this decision**; found by the premise audit.
- **`internal/storage/store.go:32`, the `Store` *type* comment** — *"no other
  package imports a SQL driver"* — is false and was already false: the import
  graph says **exactly two non-`_test.go` files import a driver**, `store.go`
  itself and `internal/storage/storagetest/storagetest.go`. (A grep says three;
  `backup.go`'s hit is a comment. Third time in this spec.) It is **not**
  falsified by this decision — it is falsified by the measurement this decision
  rests on. STAGE-022's Design Notes analysed it, called it non-blocking and
  routed it as outside SPEC-080's scope.

**Why the `Store` type comment is included anyway.** It is one sentence, in a
file this spec already edits, twenty lines below a comment this spec repairs,
and it is wrong under **both** readings of the rule — the same defect class as
`list_test.go:275`, which framing *did* route. Leaving it would make `store.go`
internally contradictory the moment the package comment above it starts citing
DEC-047. The cost of including it is one extra correct sentence; the cost of
excluding it is a known-false claim shipping in the file this spec certified.
Recorded as an addition to the framed Outputs, not as a discovery at build.

| File | What was false | The new claim |
|---|---|---|
| `internal/cli/list_test.go:274` | *"CLI tests **cannot** import database/sql"* — false under **both** readings | they **may**; this helper delegates for **hygiene**, because a raw `UPDATE` is a copy of storage's schema |
| `internal/cli/root.go:9` | *"held today by convention and review, not an automated test"* — falsified by SPEC-082 | enforced by depguard on every non-test file; scoped to those files by DEC-047 |
| `internal/storage/store.go:13` | same clause, same falsification | same repair, one sentence shorter |
| `internal/storage/store.go:33` | *"no other package imports a SQL driver"* — `storagetest` does | names the **two** non-test driver importers, and points at `go list -deps ./cmd/brag` for the shipping claim |
| `internal/storage/storagetest/storagetest.go:3` | *"lets CLI tests use these helpers **without violating** the constraint"* | preferred **because** it keeps the SQL in one place, **not** because the constraint forbids the alternative |

`list_test.go` is the delicate one and the reason it needs a **new claim, not a
patch**: the file is a *good citizen* — it really does route through
`storagetest`. Only its stated reason is wrong. Patching "cannot" to "should
not" would leave the sentence still wrong about *why*, which is SPEC-080's
recurring lesson (*test the claim, not the counterexample*) in miniature.

The production-scoping clause in `root.go` and `store.go` — *"Its production
code imports no SQL driver"* — is **correct and survives verbatim**. Only the
enforcement clause moves.

### Rejected alternatives

- **Patch `list_test.go`'s verb only** (`cannot` → `should not`). Leaves the
  sentence false about its reason. Named above.
- **Delete the `list_test.go` comment.** The helper's non-obvious property is
  *why* it delegates when it needn't; deleting the comment deletes the only
  answer.
- **Leave `storagetest.go` for a later spec.** It is one sentence, falsified by
  this exact decision, in a file this spec's rationale points at.
- **Leave the `Store` type comment routed, as STAGE-022 left it.** Defensible —
  it is not blocked on this decision. Rejected on file coherence: see *Why the
  `Store` type comment is included anyway* above.
- **Rewrite the `Store` type comment to say "no other package's production code
  imports a driver".** Still false — `storagetest.go` is not a `_test.go` file.
  The honest claim is about what *ships*, which is why the new wording cites
  `go list -deps ./cmd/brag` instead of a file-suffix reading.

---

## Fork 4 — does `story_test.go`'s deletion belong here?

**Settled: yes, and only that one file.** Confirmed, not split.

`story_test.go` carries `_ "modernc.org/sqlite"` and contains **no SQL at
all** — no `database/sql` import, no `sql.` reference anywhere in the file. It
is litter, one line, in the same package and on the same subject. Deleting it
also makes DEC-047's **T1** trigger honest: T1 counts files, and litter
inflates the count toward a threshold it has no business moving.

**The other three blank driver imports stay.** They were nearly deleted too —
the §12(b) run showed all four are redundant, because `internal/storage`'s
production code registers the driver and every CLI test imports
`internal/storage`. They are kept because they are **redundant registration but
accurate documentation**: each sits beside a live `sql.Open("sqlite", …)` and
names the driver that call uses, which is ordinary Go. `story_test.go`'s names
a driver no line in the file uses. That is the line, and it is a real one.

### Rejected alternatives

- **Delete all four blank driver imports.** Tidier, and takes the count from 7
  import lines to 4 — but it makes three files depend invisibly on
  `internal/storage` continuing to import the driver, and it widens a
  framing-bounded one-line change into a four-file sweep on design's own
  authority.
- **Add `_ "modernc.org/sqlite"` to `project_test.go`.** Locked, then killed by
  the pre-flight — the free-ride it was fixing does not exist. §12(b) Finding 2.
- **Split the deletion into its own spec.** One line, proven dead twice (at
  framing and again at design), same subject. A PR of its own costs more than
  the line.

---

## Fork 5 — does the now-permanently-unenforced test half get an assertion?

**Settled: no assertion on the test half. A comment in `.golangci.yml`, and
DEC-047. What *does* get assertions is the amendment itself.**

The asymmetry is real: production is guarded by depguard; the test half is, by
definition, unguarded. But **both available assertions would be wrong**:

- *"No CLI test imports SQL"* contradicts the decision this spec is recording.
- *"Some CLI test imports SQL"* fires on the day someone finally ports them to
  `storagetest` — punishing the exact outcome DEC-047's T1 trigger wants. A
  guard that goes red when the codebase improves gets deleted, and takes its
  neighbours' credibility with it.

So the asymmetry is **documented, not asserted**: the rewritten `!$test`
comment says the test half is out of scope *by decision*, and DEC-047 says what
would bring it back.

The guarding effort goes where the actual risk is. This spec **narrows a
blocking constraint's prose**, which is the manoeuvre where enforcement quietly
erodes — the mechanism keeps running while the rule it points at drifts. Three
assertions pin the narrowing to exactly what was decided:

- **`AA1`** — the rule text is production-scoped, `severity` is **still
  `blocking`**, `paths` is unchanged, `rationale` cites `DEC-047`.
- **`AA2`** — the register entry is `answered`, cites `DEC-047`, and that record
  exists on disk.
- **`AA3`** — the four repaired claims stay repaired (NOT-contains on the false
  phrase, so the prose stays free to be reworded — just not back).

`AA1`'s severity check is the load-bearing one: **the spec that narrowed this
rule must not be remembered as the spec that softened it.**

### Rejected alternatives

- **A Go test asserting no CLI test file imports SQL.** Contradicts the decision.
- **A Go test asserting the count of CLI test files that do.** Fires on
  improvement; see above. The count lives in DEC-047's T1 trigger, where a
  human reads it deliberately, not in a gate that goes red on its own.
- **Nothing at all — a comment is enough.** A comment does not fail CI when
  `severity` drops to `warning`. This spec is the one place where that check
  is cheap to add and obviously motivated.
- **Widen `depguard` to `**/internal/cli/**/*.go` while we are here.** Correct
  eventually (T3), but `internal/cli` has no subpackages today, and it would
  break `Z3`'s literal `internal/cli/*.go` needle. Out of scope; recorded as a
  trigger instead.

---

## Outputs

### New files (1)

| Path | What |
|---|---|
| `decisions/DEC-047-no-sql-in-cli-layer-binds-production-code.md` | The decision record. 237 lines, embedded verbatim in *Notes for the Implementer*. **Must carry `  type: decision`** in its front-matter or `Z7` fails and the inventory silently under-reports. |

### Modified files (10)

| Path | Change |
|---|---|
| `guidance/constraints.yaml` | `no-sql-in-cli-layer` — `rule:` production-scoped, `rationale:` extended. `severity` / `paths` / `added_by` / `added_at` **unchanged**. |
| `guidance/questions.yaml` | `no-sql-in-cli-layer-test-scope` → `status: answered`, `+ answered_at`, `+ answered_by`, notes prefixed with the resolution and suffixed with what actually resolved it. |
| `.golangci.yml` | The `!$test` comment. **No config key changes** — `Z3`'s five needles all survive. |
| `internal/cli/root.go` | Package comment, enforcement clause only. |
| `internal/storage/store.go` | **Two** comments: the package comment's enforcement clause, and the `Store` **type** comment's *"no other package imports a SQL driver"* claim. |
| `internal/storage/storagetest/storagetest.go` | Package comment — the "without violating" reason. |
| `internal/cli/list_test.go` | `mustBackdate`'s comment — a new claim. |
| `internal/cli/story_test.go` | Delete `_ "modernc.org/sqlite"` and the blank line above it. |
| `docs/engineering-practices.md` | The depguard bullet's last sentence **+ the regenerated inventory block**. |
| `scripts/test-docs.sh` | `Y3` re-pin + comment repair, `Y4` re-pin, `X6` comment, new Group `AA`. |

### Premise audit (§9), run at design against the repo

The sweep is `grep -rn "no-sql-in-cli-layer" . --exclude-dir=.git` —
**312 matches in 91 files** before filtering. Filtered to live, non-historical
files (`--exclude-dir=.claude --exclude-dir=done --exclude-dir=reports`),
**36 files** remain. Every one was audited for a *status claim* about the
constraint. Result:

| Hit | Verdict |
|---|---|
| `guidance/constraints.yaml:31` | **EDIT** — the subject. |
| `.golangci.yml:32,54,71,77,79` | **EDIT line 67-70 only.** Lines 32/54/77/79 describe what depguard denies and stay true verbatim. |
| `internal/cli/root.go:8` | **EDIT** — false enforcement clause. |
| `internal/storage/store.go:13` | **EDIT** — same clause. |
| `internal/storage/store.go:33` (no literal `no-sql-in-cli-layer`; found via STAGE-022's Design Notes, not the sweep) | **EDIT** — *"no other package imports a SQL driver"*, false under both readings. The sweep does **not** surface it: the comment never names the constraint. A grep for the constraint id is not a grep for its claims. |
| `internal/storage/storagetest/storagetest.go:4` | **EDIT** — found by this audit, not by framing. |
| `internal/cli/list_test.go:275` | **EDIT** — false under both readings. |
| `docs/engineering-practices.md:142-145` | **EDIT** — *"the test-file half of that constraint is open"* stops being true. Named in `questions.yaml` as one of the three places SPEC-082 made the open half visible; all three move together. |
| `guidance/questions.yaml:896` | **EDIT** — the register entry. |
| `scripts/test-docs.sh:1653,1662` | **EDIT for other reasons** (`Y3`/`Y4`/`X6`/`AA`), but `Z3`'s own claim needs none: its *"depguard denies BOTH halves"* means the `database/sql` half and the driver half of the **deny list**, not production/test. True verbatim. |
| `internal/cli/impact_test.go:45`, `wrapped_test.go:47` | **NO CHANGE.** Both say *"Raw SQL is confined to the test file; the production CLI layer stays SQL-free (no-sql-in-cli-layer)"* — a claim about **production**, true before and after. Checked against the path each cites. |
| `internal/mcpserver/import_audit_test.go:10` | **NO CHANGE, deliberately.** `TestNoSQLImport` walks *all* `.go` files including tests, so that package holds itself **stricter** than this constraint requires. Its comment says the constraint's *path glob* covers `internal/cli/**` only — still true. A package's own convention, left alone; noted in DEC-047 Consequences so a reader does not mistake it for evidence about `internal/cli`. |
| `AGENTS.md:238` | **NO CHANGE.** *"No SQL outside `internal/storage`. Enforced by the `no-sql-in-cli-layer` constraint."* — a pointer, not a scope claim. |
| `AGENTS.md:320` | **NO CHANGE.** The SPEC-007 §12 lesson narrative — a dated historical account of a 2026-04-20 punch list, correct as history. |
| `docs/api-contract.md:1256,1295` | **NO CHANGE.** Both assert specific *commands* are SQL-free. Production claims, still true. |
| 15× `decisions/DEC-0*.md` | **NO CHANGE.** Each says its own feature keeps SQL in `internal/storage` — production claims, all still true. DECs are historical records; none claims the test scope. |
| `projects/**` (briefs, stages, non-archived specs) | **NO CHANGE.** Dated planning records. |
| `docs/reports/**`, `projects/**/specs/done/**` | **NO CHANGE.** Archived. |
| `NEXT-SESSION-PROMPT.md` | **NO CHANGE** — session handoff scratch, not repo content. |
| `.claude/worktrees/**` | **NOT REPO CONTENT.** Gitignored second copy; the trap `scripts/inventory.sh` and `test-docs.sh` E2 both call out by name. |

**The literal-value half (§9(b)) — grep the harness for every value this spec
moves.** Three values move. Four guards pin them:

| Value | 45 → 46 | 184 → 187 | 7 → 6 |
|---|---|---|---|
| `X3` (byte-exact inventory block) | ✔ | ✔ | ✔ |
| `Y3` (`Decision records \| 45 \|`) | ✔ **← the one the concept-grep missed** | | |
| `Y4` (`of those, still open \| 7 \|`) | | | ✔ |
| `X6` comment (`1 of 45 DECs`) | ✔ (prose only) | | |

`X7`'s word band is a fifth guard on a value this spec moves indirectly:
2,494 → 2,525 words, band 1800..2700. Verified, 175 words of headroom.

### §12 NOT-contains self-audit

`AA3` asserts five phrases are **absent** across four `.go` files. Every
replacement literal in *Notes for the Implementer* was grepped for its own
forbidden phrase at design:

| File | Forbidden phrase | Present in the new literal? |
|---|---|---|
| `internal/cli/root.go` | `convention and review` | **no** |
| `internal/storage/store.go` | `convention and review` | **no** |
| `internal/storage/store.go` | `no other package imports a` | **no** |
| `internal/cli/list_test.go` | `cannot import database/sql` | **no** |
| `internal/storage/storagetest/storagetest.go` | `without violating` | **no** |

Scope note: `AA3` reads those four `.go` files only. This spec's own markdown,
and the needle list inside `scripts/test-docs.sh`, contain the phrases and are
not scanned by `AA3` or by any Group P sweep (Group P is file-scoped to
`README.md`, `AGENTS.md`, `docs/architecture.md`, `docs/tutorial.md`,
`cmd/brag/main.go`, `BRAG.md`). Confirmed by the full harness run: **ALL OK**.

---

## Locked design decisions

**LD1 — the rule text is scoped in `rule:`; `severity`, `paths`, `added_by`
and `added_at` do not move.** Exact literal in Notes §1. `severity: blocking`
is unchanged and pinned by `AA1`. Fork 1.

**LD2 — the amendment convention for this repo: change `rule:`, cite the
`DEC-*` in `rationale:` with the date, leave `added_at` at the original date.**
No new front-matter fields. This is the repo's first constraint amendment, so
the convention is set here and recorded in DEC-047 Option E. Fork 1(c).

**LD3 — DEC-047 carries four revisit triggers (T1 volume=6, T2 kind, T3
subpackage shape, T4 a real frontend), each with its re-derivation command and
its current value.** T1 and T2 resolve differently and must not be merged.
Fork 2.

**LD4 — five comments are repaired, not three; each gets a new *claim*, not a
patched verb.** The `storagetest` package comment is the fourth (falsified by
this decision, found by the premise audit); `store.go`'s `Store` **type**
comment is the fifth (falsified by the measurement, routed by STAGE-022, and
included here on file-coherence grounds — named as a scope addition, not
absorbed). `root.go` and `store.go`'s **package** comments keep their
production-scoping clause verbatim; only the enforcement clause moves. Fork 3.

**LD5 — `story_test.go`'s dead import is deleted here; the other three blank
driver imports are kept.** The line: `story_test.go` names a driver no line in
the file uses; the other three sit beside a live `sql.Open("sqlite", …)`. All
four are *redundant* (the driver is registered transitively by
`internal/storage`'s production import, proven at §12(b) Finding 2) — but
redundant-and-accurate is not the same as dead. Fork 4.

**LD6 — no assertion on the test half; three assertions on the amendment
(`AA1`/`AA2`/`AA3`).** Both possible test-half assertions are wrong, one by
contradiction and one by firing on improvement. Fork 5.

**LD7 — `Y3` is re-pinned 45 → 46 *and* its worked example is repaired.**
`Y3`'s comment illustrates the failure mode it catches with *"script and page
would still agree, just agree on 46"* — a number this spec makes correct.
Left alone, the comment reads as if the correct count were the bug. The
example moves to **47** (46 decisions + the DEC-041 reservation miscounted as
a decision), which is what that failure mode would actually produce after this
spec. A re-pin note in `Y4`'s style records the move as deliberate.

**LD8 — the practices page sentence changes and the inventory block is
regenerated in the same edit.** `X3` diffs the block byte-for-byte against
`scripts/inventory.sh`; three of its rows move. The remedy is mechanical:
`just inventory`, paste between the markers. Never hand-edit a row.

### Rejected alternatives (build-time)

- **Hand-editing the three moved rows in the inventory block** instead of
  running `just inventory`. `X3` recomputes the whole table; a hand-edit that
  gets one row right and another wrong fails with a full diff anyway. Run the
  script.
- **Bumping `X7`'s word band** if the practices-page edit overshoots. It does
  not overshoot (2,525 of 2,700, measured). If a build-time reword did push
  it over, the fix is a shorter sentence — raising a band to make a page fit
  is the disarming SPEC-079 LD5 refused.
- **Adding the amendment to `AGENTS.md` §8** (*"No SQL outside
  `internal/storage`. Enforced by the `no-sql-in-cli-layer` constraint."*).
  Audited: it is a pointer to the constraint, not a restatement of its scope,
  so it stays true. Editing it would duplicate the scope in a second place
  that can then drift — the exact failure this spec exists to fix.
- **Renaming the constraint id** to something like
  `no-sql-in-cli-production-code`. Ids are stable; 457 references exist.

---

## Acceptance Criteria

1. `guidance/constraints.yaml`'s `no-sql-in-cli-layer` `rule:` is the LD1
   literal; `severity: blocking`, `paths: ["internal/cli/**"]`, `added_by` and
   `added_at` are **byte-identical to main**; `rationale:` cites `DEC-047`.
   The file parses as YAML.
2. `decisions/DEC-047-no-sql-in-cli-layer-binds-production-code.md` exists,
   carries `  type: decision`, and states the decision, the four rejected
   alternatives, and the accepted cost with its four triggers.
3. `no-sql-in-cli-layer-test-scope` is `status: answered`, carries
   `answered_at` / `answered_by`, and cites `DEC-047`. The file parses as YAML.
4. All five comments are true as written, **checked against the paths they
   cite**; `root.go` and `store.go`'s **package** comments retain their
   production-scoping clause verbatim.
5. `internal/cli/story_test.go` has no `modernc.org/sqlite` import;
   `project_test.go`, `coverage_test.go`, `impact_test.go` and `wrapped_test.go`
   are **untouched**. `gofmt -l .`, `go vet ./...`, `just test` all clean.
6. `golangci-lint config verify` exits 0 and `golangci-lint run` reports
   **0 issues**; **M-A and M-B still fire**, each with the mutant confirmed
   present by content hash before the failure is credited.
7. `just test-docs` **ALL OK** at **187** assertion ids, with the inventory
   block regenerated and pasted (`45 → 46`, `184 → 187`, `7 → 6`).
8. `./scripts/inventory.sh` and the block between the `inventory:begin` /
   `inventory:end` markers are byte-identical (`X3`).

---

## Failing Tests

Every literal below was run at design; the exact text is in *Notes for the
Implementer* §7. Write them first, watch them fail for the **stated** reason,
then make them pass.

### New — `scripts/test-docs.sh`, Group `AA` (`AA1`–`AA3`)

| Id | Asserts | Fails before the change because |
|---|---|---|
| `AA1` | the `no-sql-in-cli-layer` block in `constraints.yaml` has a production-scoped `rule:`, `severity: blocking`, an unchanged `paths:` glob, and a `rationale:` citing `DEC-047` | the rule text is still unqualified and nothing cites `DEC-047` → `[rule text is not scoped to production files] [rationale does not cite DEC-047]` |
| `AA2` | the register entry is `answered`, cites `DEC-047`, and `decisions/DEC-047-*.md` exists | `status: open`, no citation, no file → all three sub-reasons |
| `AA3` | five falsified phrases are gone from four `.go` files | all five still present → five entries in the failure list |

### Changed — `Y3`

Re-pin `'Decision records | 45 |'` → `| 46 |` and `decision-records!=45` →
`!=46`, plus the comment repair (LD7). Fails before `DEC-047` is added:
`inventory.sh row value(s) wrong: decision-records!=46`.

### Changed — `Y4`

Re-pin `'of those, still open | 7 |'` → `| 6 |` and `questions-open!=7` →
`!=6`. Fails until the register entry flips to `answered`.

### Changed — `X3`

No code change; the **inventory block on the practices page** must be
regenerated. Fails with a full script-vs-page diff until `just inventory` is
run and the output pasted between the markers.

### Changed — `X6`

Comment only (`1 of 45` → `1 of 46`; *"matches all 45"* → *"matches all of
them"*). No assertion behaviour changes — `DEC-047` carries no `## Amendment`
heading, so the "only 1" claim still holds. Verified at design: the
`…carrying an explicit ## Amendment section` inventory row stays at **1**.

### Mutation checks (run by build, recorded in Build Completion)

All seven ran at design; build re-runs them and pastes the output.

| # | Mutation | Expected |
|---|---|---|
| **M-A** | `_ "database/sql"` in `internal/cli/root.go` | `golangci-lint run` → 1 depguard issue naming `database/sql` |
| **M-B** | `_ "modernc.org/sqlite"` in `internal/cli/root.go` | 1 depguard issue naming `modernc.org/sqlite` |
| **M-C** | `severity: blocking` → `warning` | `FAIL: AA1 … [severity is no longer blocking]` |
| **M-D** | rule text reverted to the unqualified wording | `FAIL: AA1 … [rule text is not scoped to production files]` |
| **M-E** | register entry back to `status: open` | `FAIL: AA2 … [status is not answered]` |
| **M-F** | `"convention and review"` restored in `root.go` | `FAIL: AA3 … root.go still says "convention and review"` |
| **M-G** | `"no other package imports a"` restored in `store.go` | `FAIL: AA3 … store.go still says "no other package imports a"` |

**Confirm every mutant actually mutated, by content hash.** `shasum -a 256` the
file before and after; a red with an unmoved hash is a red for the wrong
reason. `git diff --quiet` does **not** substitute — it is blind to untracked
files. After each probe, restore and confirm the hash returns to its
pre-mutation value.

### Decision-to-test mapping (§9)

Every locked decision has an assertion that fails without it.

| Decision | Test |
|---|---|
| LD1 rule text + `severity`/`paths` unchanged | `AA1` (all four sub-checks); M-C, M-D |
| LD2 amendment convention (cite the DEC in `rationale`) | `AA1`'s `DEC-047` needle |
| LD3 DEC-047 exists with the decision and its triggers | `AA2`'s on-disk check; `Z7` (it must be `type: decision`) |
| LD4 five comments repaired | `AA3`; M-F, M-G |
| LD5 `story_test.go` deleted, other three untouched | `just test` + `gofmt -l .` + `go vet` clean; the AC-5 untouched check |
| LD6 no test-half assertion; three amendment assertions | `AA1`/`AA2`/`AA3` exist and fire (M-C…M-F); **no** assertion references the count of CLI test files |
| LD7 `Y3` re-pinned and its example repaired | `Y3` |
| LD8 practices sentence + regenerated block | `X3`, `X7`, `Z5` |

---

## Implementation Context

### Decisions that apply

- **DEC-047** (this spec creates it) — the whole subject. Read it first; the
  spec's forks are its reasoning, and its four triggers are the only reason
  the accepted cost is a choice rather than a shrug.
- **DEC-001** — pure-Go SQLite driver. The reason depguard's deny list can
  name `modernc.org/sqlite` specifically, and the reason
  `internal/storage/store.go:29` carries the blank driver import that makes
  all four CLI-test blank imports redundant (§12(b) Finding 2).

### Constraints that apply

- **`no-sql-in-cli-layer` (blocking)** — the subject. This spec **amends** it.
  Nothing here may weaken it: `severity` stays `blocking`, `golangci-lint run`
  stays at 0, and M-A/M-B must still fire. If the amended wording would let
  either mutation pass, the wording is wrong.
- **`one-spec-per-pr` (blocking)** — one branch, one PR.
- **`test-before-implementation`** — Group `AA` and the `Y3`/`Y4` re-pins go in
  first and are watched failing for their stated reasons.
- **`no-cgo`, `timestamps-in-utc-rfc3339`, `storage-tests-use-tempdir`,
  `stdout-is-for-data-stderr-is-for-humans`, `errors-wrap-with-context`** —
  N/A. No production behaviour changes; no Go code changes except one deleted
  import line.

### Prior related work

- **SPEC-082** (#178) mechanised the production half with depguard, filed
  `no-sql-in-cli-layer-test-scope` with both resolutions costed, and
  deliberately did **not** amend the rule — a spec must not amend the
  constraint it is mechanising. Its LD7 is the direct antecedent.
- **SPEC-080** (#176) proved the gap empirically: adding `_ "database/sql"` to
  `internal/cli/root.go` passed `go build`, `go vet`, `gofmt`, `just test` and
  `just test-docs`. It also scoped `root.go`'s package comment to *"production
  code"* — one of the three artifacts that already assumed this reading — and
  routed the three stale comments this spec repairs.
- **SPEC-007** (2026-04-20) is where this constraint first bit: a verify punch
  list caught a spec offering a test helper that imported `database/sql` under
  `internal/cli/`. That episode is the source of AGENTS.md §12's "spec prose
  cannot relax a blocking constraint" rule — worth reading before amending the
  same constraint's prose.

### Out of scope (for this spec specifically)

- **Filling `storagetest`'s gaps** (`BackdateUpdated`, `BackdateProject`).
  That is DEC-047's T1 trigger, at 4 of 6.
- **Deleting the other three blank driver imports.** LD5.
- **Widening depguard to `**/internal/cli/**/*.go`.** DEC-047 T3; also breaks
  `Z3`'s literal needle.
- **Extending depguard to `internal/mcpserver`.** Same class of side effect —
  it would widen a blocking constraint's path glob. Named in the register
  entry and left there.
- **Touching `impact_test.go` / `wrapped_test.go` comments.** Audited: both
  claim the *production* layer is SQL-free, true before and after.

---

## Notes for the Implementer

**Literal-artifact spec.** Transcribe verbatim; verify diffs against these
blocks. Every one was written into a worktree, run through its tool, and the
outputs are recorded in the §12(b) section above. Do not improve the prose in
transit — if something reads wrong, that is a question for the spec, not a
build-time edit.

### Order of work

1. `scripts/test-docs.sh` — Group `AA`, `Y3`, `Y4`, `X6` (§7). Run
   `./scripts/test-docs.sh` and confirm `AA1`/`AA2`/`AA3`/`Y3`/`Y4` fail for
   the reasons the Failing Tests table names.
2. `decisions/DEC-047-*.md` (§8) — creating it flips `Y3` and half of `AA2`.
3. `guidance/constraints.yaml` (§1), `guidance/questions.yaml` (§2) — flips
   `AA1`, `AA2`, `Y4`.
4. The four comments (§4) and the `story_test.go` deletion (§5) — flips `AA3`.
5. `.golangci.yml` (§3).
6. `docs/engineering-practices.md` (§6): edit the sentence, then run
   `just inventory` and paste its output between the `inventory:begin` /
   `inventory:end` markers. **Regenerate; never hand-edit a row.**
7. Full battery: `gofmt -l .`, `go vet ./...`, `just test`,
   `golangci-lint config verify`, `golangci-lint run`, `just test-docs`.
8. The seven mutation checks, each with the mutant confirmed present by
   `shasum -a 256` before and after, and the hash confirmed restored.

`golangci-lint` is not on `PATH` by default — it lives at `~/go/bin`. Either
`export PATH="$HOME/go/bin:$PATH"` or
`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1`.

### 1. `guidance/constraints.yaml` — replace two lines in the `no-sql-in-cli-layer` entry

Both are single-line YAML scalars. `severity`, `paths`, `added_by` and
`added_at` are **not touched**.

```yaml
  - id: no-sql-in-cli-layer
    rule: "Production files under internal/cli/ — every *.go except *_test.go — must not import database/sql or any SQL driver. All persistence goes through internal/storage."
    severity: blocking
    paths: ["internal/cli/**"]
    added_by: claude
    added_at: 2026-04-19
    rationale: "Architecture principle 2 (architecture.md) — CLI is a thin shell over storage. Keeps commands testable and future frontends (TUI, API) feasible: a TUI reusing internal/cli links its production code, never its *_test.go files. Scoped to production files by DEC-047 (2026-08-22), this rule's first amendment — added_at above remains the original date. Enforced by depguard (.golangci.yml, rule no-sql-in-cli-layer). Test files needing raw SQL should prefer internal/storage/storagetest."
```

The em-dashes and the `:` inside the `rationale` scalar are safe — it is
double-quoted. Verified: `yaml.safe_load` reads back 11 constraints and this
entry's three fields unchanged.

### 2. `guidance/questions.yaml` — two hunks in `no-sql-in-cli-layer-test-scope`

**2a — the header block.** Note `answered_at` goes after `raised_at` and
`answered_by` after `assigned_to`, matching `editor-template-format`:

```yaml
    priority: medium
    status: answered
    raised_by: claude-opus-5
    raised_at: 2026-08-21
    answered_at: 2026-08-22
    assigned_to: null
    answered_by: DEC-047 (SPEC-083) — user decision, 2026-08-21
    notes: |
      ANSWERED — resolution (A). The constraint binds PRODUCTION code only.
      guidance/constraints.yaml's rule text now says so, severity is unchanged
      at blocking, and DEC-047 carries the reasoning, the rejected alternative
      and the accepted cost. The original filing, unedited, follows.

      Filed at SPEC-082 design (LD7). The constraint's RULE TEXT is unqualified
```

Everything from `Filed at SPEC-082 design (LD7).` onward is the existing text,
**unedited**, except hunk 2b.

**2b — the trailer.** The original `RESOLVE WHEN:` paragraph is preserved as
filed (re-labelled, and re-wrapped on the first two lines only), with a new
paragraph after it. It is the last block in the file:

```yaml
      RESOLVE WHEN (as filed): the next spec that opens internal/cli/list_test.go
      (whose line-275 comment is wrong under BOTH readings and is on STAGE-021's
      routed stale-comment list), or the next spec that needs a sixth CLI test
      to backdate a row — whichever comes first. Either one has to answer this
      to write a correct comment or a correct helper.

      WHAT ACTUALLY RESOLVED IT: the first trigger, one spec later. SPEC-083
      opened list_test.go, could not write a true comment without an answer,
      and the user decided the scope question directly. DEC-047 records it and
      names the revisit triggers for the cost that resolution (A) accepts.
```

### 3. `.golangci.yml` — replace the four `WHY PRODUCTION FILES ONLY` comment lines

Comment only. **No config key moves** — `Z3`'s five needles
(`no-sql-in-cli-layer:`, `internal/cli/*.go`, `!$test`, `pkg: "database/sql"`,
`pkg: "modernc.org/sqlite"`) all survive, and `golangci-lint config verify`
still exits 0.

```yaml
        # WHY PRODUCTION FILES ONLY (`!$test`). Because production files are
        # the whole constraint. DEC-047 (2026-08-22) scoped the rule text to
        # them, so `!$test` is the mechanism AGREEING with the rule, not
        # falling short of it. Four *_test.go files under internal/cli/ open a
        # database directly, all four to backdate a fixture the store stamps
        # with time.Now(); that is out of scope by decision, and
        # internal/storage/storagetest is the preferred route for new ones.
        no-sql-in-cli-layer:
          files:
            - "**/internal/cli/*.go"
            - "!$test"
```

### 4. The four comments

**4a — `internal/cli/root.go`**, the tail of the package comment. Everything
above `// (window.go). Its production code…` is unchanged:

```go
// (window.go). Its production code imports no SQL driver and no
// database/sql — the no-sql-in-cli-layer boundary, enforced on every
// non-test file here by depguard since SPEC-082 (.golangci.yml) and
// scoped to exactly those files by DEC-047 — so every command reaches
// persistence only through internal/storage, keeping the CLI a thin
// shell a future frontend (TUI, API) could replace. Test files in this
// package may open a database directly; four do, to backdate a fixture.
package cli
```

**4b — `internal/storage/store.go`**, the tail of the package comment:

```go
// schema (DEC-017/019/020). Every other package's production code
// reaches the database only through a *Store — the no-sql-in-cli-layer
// boundary on internal/cli, enforced there by depguard since SPEC-082
// and scoped to production files by DEC-047 — which is what keeps
// commands testable and a future frontend feasible.
package storage
```

Lines 3–5 (*"Test files elsewhere import it too; the storagetest test-helper
subpackage exists so they need not…"*) are **unchanged** — audited, still true.

**4c — `internal/storage/store.go`**, the `Store` **type** comment (a *second,
separate* hunk in this file, ~20 lines below 4b):

```go
// Store is the typed wrapper around *sql.DB for the bragfile database.
// Everything that ships persists through a Store: only two non-test files
// import a SQL driver — this one and internal/storage/storagetest, which
// is absent from `go list -deps ./cmd/brag`. Test code is the deliberate
// exception; four internal/cli test files open a database directly to age
// a fixture (DEC-047).
type Store struct {
```

Both claims were verified at design: `go list -deps ./cmd/brag | grep -c
storagetest` → **0**, and the `go/parser` sweep over non-test files → exactly
`store.go` and `storagetest.go`.

**4d — `internal/storage/storagetest/storagetest.go`**, the whole package
comment:

```go
// Package storagetest exposes test-only helpers that need raw SQL
// access to a Bragfile database. Living under internal/storage/ keeps
// the database/sql dependency — and the schema knowledge that rides
// with it — inside the storage layer. A CLI test may open a database
// directly if it wants to (no-sql-in-cli-layer binds production files
// only, DEC-047); this package is the preferred route because it keeps
// that SQL in one place, not because the constraint forbids the other.
package storagetest
```

**4e — `internal/cli/list_test.go`**, `mustBackdate`'s doc comment:

```go
// mustBackdate forwards to storagetest.Backdate and t.Fatals on error.
// no-sql-in-cli-layer binds production files only (DEC-047), so a CLI
// test MAY open a database directly — four in this package do. This one
// does not: the helper already exists, and every raw UPDATE is one more
// copy of storage's schema living outside internal/storage.
func mustBackdate(t *testing.T, dbPath string, id int64, at time.Time) {
```

### 5. `internal/cli/story_test.go` — delete two lines

Delete the blank line **and** the import. The result is a two-group import
block that `gofmt` leaves alone (verified):

```go
import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)
```

**Do not touch** `coverage_test.go`, `impact_test.go`, `project_test.go` or
`wrapped_test.go` (LD5).

### 6. `docs/engineering-practices.md` — one sentence, then regenerate the block

Replace the last sentence of the `**A blocking constraint has a machine behind
it.**` bullet:

```markdown
  to `no-sql-in-cli-layer`: adding `_ "database/sql"` to
  `internal/cli/root.go` passed `go build`, `go vet`, `gofmt`, `just test` and
  `just test-docs`, and now fails the `lint` job. It guards production files
  only, and since DEC-047 that is the whole rule: the constraint text was
  amended to match the mechanism rather than the mechanism widened to match
  the text, and the four test files that open a database to backdate a fixture
  are out of scope by decision.
```

Then `just inventory` and paste between the markers. Three rows move —
`Decision records` 45→46, `Documentation assertions (distinct ids)` 184→187,
`…of those, still open` 7→6. Everything else is byte-identical. Result:
**2,525 words**, inside `X7`'s 1800..2700 band.

### 7. `scripts/test-docs.sh` — four edits

**7a — `X6`'s comment, two lines** (the illustrative DEC count; no assertion
behaviour changes):

```sh
# correction" (only 1 of 46 DECs carries an explicit `## Amendment` heading, and
# a keyword grep matches all of them because ordinary prose uses those words), so
```

**7b — `Y3`: the worked example, the re-pin note, and the pinned value.**
Two edits inside one block, shown here as the **final contiguous region**
(replacing the old single line `# just agree on 46).` and the old
`Decision records | 45 |` needle; the six lines between them are unchanged):

```sh
# just agree on 47).
#
# RE-PINNED 45 -> 46 at SPEC-083, which adds DEC-047. Deliberate corpus
# change, not drift. Note this is the SECOND guard on a number SPEC-083
# moves — Y4 pins the other — which is exactly the pair AGENTS.md §9's
# half (b) exists for: grep the harness for every literal the spec moves.
if [ ! -x scripts/inventory.sh ]; then
    fail "Y3" "scripts/inventory.sh is missing or not executable"
else
    y3_out=$(./scripts/inventory.sh)
    y3_bad=""
    printf '%s\n' "$y3_out" | grep -F -q 'Decision records | 46 |' || y3_bad="$y3_bad decision-records!=46"
```

**7c — `Y4`: a re-pin note after the SPEC-082 note, and the pinned value.**
Same shape — final contiguous region, unchanged lines included for anchoring:

```sh
# RE-PINNED 18/6 -> 19/7 at SPEC-082, which appends the
# no-sql-in-cli-layer-test-scope question (LD7). This is a deliberate
# corpus change, not drift: the pin moves with the register it describes.
#
# RE-PINNED AGAIN 19/7 -> 19/6 at SPEC-083, which ANSWERS that same question
# (DEC-047). The total does not move: the entry is closed, not removed.
if [ ! -x scripts/inventory.sh ]; then
    fail "Y4" "scripts/inventory.sh is missing or not executable"
else
    y4_out=$(./scripts/inventory.sh)
    y4_bad=""
    printf '%s\n' "$y4_out" | grep -F -q 'Questions tracked in guidance/questions.yaml | 19 |' || y4_bad="$y4_bad questions-total!=19"
    printf '%s\n' "$y4_out" | grep -F -q 'of those, still open | 6 |' || y4_bad="$y4_bad questions-open!=6"
```

**7d — Group `AA`, inserted immediately before the `# ===== finalise =====`
line** (the block below ends with that line, so paste it *over* the existing
`# ===== finalise =====` line; its trailing blank line stays put):

```sh
# ===== Group AA — the constraint amendment (SPEC-083 / DEC-047) =====
#
# SPEC-083 NARROWED a blocking constraint's prose. That is the manoeuvre where
# enforcement quietly erodes: the mechanism keeps running while the rule it
# points at drifts out from under it. These three assertions pin the narrowing
# to exactly what was decided, so a later widening, a severity downgrade or a
# revert of the repaired comments has to be a deliberate edit to this file too.
#
# What is deliberately NOT asserted: anything about the test half. It is
# unguarded BY DECISION (DEC-047), and both available assertions would be
# wrong. "No CLI test imports SQL" contradicts the decision. "Some CLI test
# imports SQL" fires the day someone finally ports them to storagetest, which
# is the outcome the decision's own revisit triggers want. An unenforced half
# stays unenforced; it is documented in .golangci.yml and in DEC-047 instead.

CONSTRAINTS_YAML="guidance/constraints.yaml"

# AA1 — THE ANTI-EROSION GUARD. Reads the no-sql-in-cli-layer entry as a block
# (its `- id:` line to the next one) and checks four things inside it: the rule
# text is production-scoped, the severity is STILL blocking, the path glob is
# unchanged, and the rationale cites the record that authorised the change.
# Severity is the load-bearing one — the spec that narrowed this rule must not
# be remembered as the spec that softened it.
if [ ! -f "$CONSTRAINTS_YAML" ]; then
    fail "AA1" "$CONSTRAINTS_YAML does not exist"
else
    aa1_block=$(awk '
        /^  - id: no-sql-in-cli-layer$/ { f=1; print; next }
        f && /^  - id: / { exit }
        f { print }
    ' "$CONSTRAINTS_YAML")
    aa1_bad=""
    if [ -z "$aa1_block" ]; then
        aa1_bad=" [no constraint with id no-sql-in-cli-layer]"
    else
        printf '%s\n' "$aa1_block" | grep -q '^    rule: "Production files under internal/cli/' \
            || aa1_bad="$aa1_bad [rule text is not scoped to production files]"
        printf '%s\n' "$aa1_block" | grep -q '^    severity: blocking$' \
            || aa1_bad="$aa1_bad [severity is no longer blocking]"
        printf '%s\n' "$aa1_block" | grep -F -q 'paths: ["internal/cli/**"]' \
            || aa1_bad="$aa1_bad [paths glob changed]"
        printf '%s\n' "$aa1_block" | grep -F -q 'DEC-047' \
            || aa1_bad="$aa1_bad [rationale does not cite DEC-047]"
    fi
    if [ -z "$aa1_bad" ]; then
        ok "AA1"
    else
        fail "AA1" "$CONSTRAINTS_YAML no-sql-in-cli-layer:$aa1_bad"
    fi
fi

# AA2 — the register entry is closed, cites the record that closed it, and that
# record exists on disk. Y5's idiom plus the citation: a question marked
# answered with no pointer to the answer is how `editor-template-format` sat
# stale for four months with its answer already shipped.
aa2_block=$(awk '
    /^  - id: no-sql-in-cli-layer-test-scope$/ { f=1; print; next }
    f && /^  - id: / { exit }
    f { print }
' guidance/questions.yaml)
aa2_bad=""
if [ -z "$aa2_block" ]; then
    aa2_bad=" [no question with id no-sql-in-cli-layer-test-scope]"
else
    printf '%s\n' "$aa2_block" | grep -q '^    status: answered$' \
        || aa2_bad="$aa2_bad [status is not answered]"
    printf '%s\n' "$aa2_block" | grep -F -q 'DEC-047' \
        || aa2_bad="$aa2_bad [does not cite DEC-047]"
fi
ls decisions/DEC-047-*.md >/dev/null 2>&1 \
    || aa2_bad="$aa2_bad [decisions/DEC-047-*.md does not exist]"
if [ -z "$aa2_bad" ]; then
    ok "AA2"
else
    fail "AA2" "guidance/questions.yaml no-sql-in-cli-layer-test-scope:$aa2_bad"
fi

# AA3 — the five comments this spec repaired stay repaired. Every phrase below
# was a FALSE claim standing on main: two package comments said the boundary
# was held "by convention and review, not an automated test" months after
# SPEC-082 made it a lint gate; store.go's Store TYPE comment said "no other
# package imports a SQL driver" while internal/storage/storagetest does (the
# import graph says exactly two non-test files import a driver, and that is the
# other one); list_test.go said CLI tests "cannot import database/sql", which
# was wrong under BOTH readings of the rule; and storagetest's package comment
# gave the constraint as the reason CLI tests route through it, which stopped
# being a reason when DEC-047 scoped the constraint to production files.
#
# NOT-contains by design: each needle is the false claim itself, so a revert
# fails while the replacement prose stays free to be reworded — just not back.
aa3_bad=""
while IFS='|' read -r aa3_file aa3_phrase; do
    [ -n "$aa3_file" ] || continue
    if [ ! -f "$aa3_file" ]; then
        aa3_bad="$aa3_bad [$aa3_file is missing]"
    elif grep -F -q -- "$aa3_phrase" "$aa3_file"; then
        aa3_bad="$aa3_bad [$aa3_file still says \"$aa3_phrase\"]"
    fi
done <<EOF
internal/cli/root.go|convention and review
internal/storage/store.go|convention and review
internal/storage/store.go|no other package imports a
internal/cli/list_test.go|cannot import database/sql
internal/storage/storagetest/storagetest.go|without violating
EOF
if [ -z "$aa3_bad" ]; then
    ok "AA3"
else
    fail "AA3" "a claim SPEC-083 corrected is back in the tree:$aa3_bad"
fi

```

Note the `while … done <<EOF` in `AA3`: a heredoc redirect runs in the current
shell, so `aa3_bad` survives the loop. A pipe would not — the same reason `E2`
uses this idiom. Note also that `internal/storage/store.go` appears **twice** in
the pair list, once per repaired claim.

### 8. `decisions/DEC-047-no-sql-in-cli-layer-binds-production-code.md` (new, 237 lines)

Transcribe verbatim. The front-matter `  type: decision` line is load-bearing:
`Z7` fails and the inventory silently under-reports without it.

```markdown
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
```


---

## Build Completion

*Filled in during the build cycle.*

- **Deviations from the spec:** <none / list>
- **New DEC-* files created:** DEC-047 (planned — not a deviation)
- **Constraints checked:** <list>
- **Gates:** <list>
- **Mutation checks M-A…M-G:** paste each result **and** the before/after
  content hashes that confirm the mutant was present.

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

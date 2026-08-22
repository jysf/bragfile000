---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-082
  type: chore                      # epic | story | task | bug | chore
  cycle: build                     # frame | design | build | verify | ship
  blocked: false
  priority: medium
  complexity: M                    # S | M | L  (L means split it)

project:
  id: PROJ-007
  stage: STAGE-022
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-20
  designed_at: 2026-08-21

insight:
  confidence: 0.88

references:
  decisions:
    - DEC-001                      # pure-Go sqlite driver — the depguard deny list names it
    - DEC-026                      # dev/prod migration guard — the ST1005 fix edits its message
  constraints:
    - no-sql-in-cli-layer          # mechanised by depguard (LD6/LD7); the scope question is LD7
    - no-new-top-level-deps-without-decision  # read at design, does not apply — see LD3
    - errors-wrap-with-context     # mechanised by errorlint (LD2)
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-023                     # the goreleaser ldflags that fed the two unused vars
    - SPEC-073                     # the coverage sentence that was wrong four times
    - SPEC-079                     # the practices page a coverage number is presented against
    - SPEC-080                     # routed the no-sql-in-cli-layer gap and the decisions/ totality gap
    - SPEC-081                     # built assert_word_count_band and wrote the switch trigger X7 now hits
---

# SPEC-082: lint and coverage, gating CI

> **Cycle: design.** Four forks settled with rejected alternatives, every
> literal run through its target tool at design (§12(b)), and the enforcement
> mutation-checked in four places. Build transcribes; verify diffs.

## Context

STAGE-022's first spec, and the one that closes PROJ-007's last named gap: the
standard signals a reviewer greps for exist and **gate CI** rather than being
available to run.

Parent: `STAGE-022-measured-and-enforced`, spec 1 of 2.
Project: `PROJ-007`.

## Goal

`golangci-lint` fails the build on a violation, coverage is reported with a floor
the current number actually clears, and the claim the practices page makes about
either one is mutation-checked before it is written down.

## What was measured at framing

Re-verified 2026-08-20. CI (`.github/workflows/ci.yml`) runs `gofmt -l .`,
`go vet ./...` and `go test ./...` — nothing else.

| | Measured |
|---|---|
| `golangci-lint` config | **none** (`.golangci.yml` / `.yaml` absent) |
| `-cover` anywhere in CI | **0 occurrences** |
| `func Benchmark` | **0** — out of scope, deferred to PROJ-009 |
| Module-wide statement coverage | **83.5%** |
| Packages with **no test files** | **0** |
| Highest | `internal/timewindow`, `internal/spark`, `internal/ftsquery` — **100%** |
| Lowest | `cmd/brag` — **23.7%** |
| Next lowest | `internal/editor` 75.4%, `internal/storage` 78.8%, `internal/cli` 82.4% |

**`golangci-lint` is not installed on this machine.** Design must decide whether
CI installs it pinned, uses the official action, or something else — and confirm
whether that counts under `no-new-top-level-deps-without-decision` (it is a CI
tool, not a module dependency, but the constraint should be read rather than
assumed).

**The 23.7% is the interesting number, and it is not a problem to fix.**
`cmd/brag` is the entrypoint — mostly `AddCommand` wiring, whose failure modes
are covered by the `internal/cli` tests one layer down. The stage's own scope
says: *if the number is embarrassing, say so on the practices page rather than
gaming it.* A floor set to force `cmd/brag` upward would be exactly the
test-writing-to-hit-a-number this stage forbids.

## Re-measurement at design (2026-08-21)

Every framing figure re-run. **One was wrong; one turns out to depend on a flag
nobody named; one hidden default was suppressing 43% of the evidence.**

| Framing said | Design measured | Verdict |
|---|---|---|
| Module coverage **83.5%** | **83.5%** | **holds** |
| `cmd/brag` **23.7%** | **23.7%** | **holds** |
| `internal/editor` 75.4% / `storage` 78.8% / `cli` 82.4% | identical | **holds** |
| Three packages at **100%** | `timewindow`, `spark`, `ftsquery` | **holds** |
| Packages with **no test files: 0** | **1 of 15** — `internal/storage/storagetest` | **WRONG** |
| No `.golangci.yml`, no `-cover` in CI, `func Benchmark` = 0 | unchanged | **holds** |
| depguard would trip "five files, not four" | **8 import lines across exactly 5 files** | **holds** |

### The drift: "zero packages without test files" is 1 of 15

`go test ./...` prints
`?  github.com/jysf/bragfile000/internal/storage/storagetest  [no test files]`.
**14 of the module's 15 packages carry at least one `_test.go`; one does not.**
The exception is `internal/storage/storagetest`, the test-helper package
SPEC-007 created so that CLI tests could backdate rows without importing
`database/sql` — it has 8 statements, 0.0% covered, and is exercised only
through the packages that call it.

Stating the unit matters here, because "0 packages without tests" and "1 of 15"
are the same repository. The honest sentence is: **every package that ships
behaviour has tests; the one that does not is a test helper.** That is a better
claim than the round zero, and it survives someone running the command.

### Two numbers that were never one number

- **Coverage depends on a flag nobody had named.** `go test ./...
  -coverprofile` reports **83.5%**. The same suite with `-coverpkg=./...`
  reports **86.2%** — 2.7 points higher for no additional test, because
  `-coverpkg` credits a package for statements some *other* package's tests
  happened to execute. LD5 pins the first definition and `scripts/coverage.sh`
  states it in a comment, because a percentage without its invocation is not a
  measurement.
- **The same profile yields 83.33% or 83.5% depending on the aggregation.**
  Summing the raw statement blocks in the profile gives 3,254 covered of
  **3,905** → 83.33%; `go tool cover -func` reports **83.5%** on that identical
  file. The cause was not run to ground and is not needed: the enforced number
  is defined as *the `total:` line of `go tool cover -func`*, full stop. The
  3,905-block figure is used below only as a scale for the headroom arithmetic.

### `golangci-lint`'s own default was hiding 43% of the evidence

`issues.uniq-by-line` defaults to **true**: golangci-lint reports one issue per
source line and silently drops the rest. Measured on this tree with
`linters.default: all`:

| | Issues reported |
|---|---:|
| Default (`uniq-by-line: true`) | **8,713** |
| `uniq-by-line: false` | **15,300** |

This is not a footnote — it changed three of this spec's own decisions.
`errorlint` reads as **2** findings under the default and is actually **7**;
`rowserrcheck` (**3**) and `sqlclosecheck` (**1**) read as **zero** and are
real. Both are now enabled. Any per-linter count quoted below was taken with
`uniq-by-line: false`, and LD4 turns it off in the shipped config.

### `cmd/brag` at 23.7% is one function

`go tool cover -func` lists exactly two functions in `cmd/brag`: `main` at
**0.0%** and `resolveVersion` at **100.0%**. The package is **38 statements**
of a ~3,900-statement module — **1.0%** — and its entire shortfall is `main`'s
**29 uncovered statements**, i.e. **0.74% of the module**. Taking `cmd/brag` to
100% would move the headline number by about seven tenths of a point, and would
require a test that invokes `main()` — which is the definition of a test
written to move a number. It is left where it is, and the practices page says so
in those terms.

Per-package statement counts, re-derived at design:

| Package | Statements | Coverage |
|---|---:|---:|
| `internal/cli` | 1728 | 82.4% |
| `internal/storage` | 638 | 78.8% |
| `internal/export` | 570 | 86.3% |
| `internal/story` | 203 | 85.7% |
| `internal/aggregate` | 187 | 90.9% |
| `internal/mcpserver` | 181 | 90.1% |
| `internal/capture` | 113 | 89.4% |
| `internal/memory` | 103 | 97.1% |
| `internal/editor` | 61 | 75.4% |
| `cmd/brag` | 38 | 23.7% |
| `internal/timewindow` | 30 | 100.0% |
| `internal/spark` | 18 | 100.0% |
| `internal/config` | 18 | 83.3% |
| `internal/ftsquery` | 9 | 100.0% |
| `internal/storage/storagetest` | 8 | 0.0% |

## §12(b) design-time pre-flight — what each tool actually said

Every literal this spec embeds was written into the working tree, run through
its own tool, and reverted. Nothing below is predicted.

| Tool | Version | What was run | Result |
|---|---|---|---|
| `golangci-lint` | **v2.13.1** (installed at design; `go install …@v2.13.1`) | `golangci-lint config verify` | **exit 0** — config validates against the JSON schema the CI action also checks |
| `golangci-lint` | v2.13.1 | `golangci-lint run` (no path args, config auto-discovered at repo root) | **19 issues** before the code fixes, **0 issues** after — see *The build punch list* |
| `actionlint` | installed | `actionlint .github/workflows/ci.yml` | **exit 0** |
| `goreleaser` | **2.17.1** | `goreleaser check` | **fails identically before and after** this spec's ldflags edit — a **pre-existing** `brews` deprecation. See LD8. |
| `scripts/coverage.sh` | — | run at `FLOOR=80.0` | passes, prints `total: 83.5%   floor: 80.0%` |
| `scripts/coverage.sh` | — | mutated to `FLOOR=90.0` | **exit 1** with the floor message — the gate has teeth |
| `just test-docs` | — | full harness with every artifact staged | **ALL OK**, 184 assertion ids |
| `gofmt -l .` / `go vet ./...` / `go test ./...` | Go 1.26.6 | all three, with every artifact and all 19 code fixes staged | clean / clean / all packages `ok` |

### Four mutation checks, each with the mutant confirmed to have mutated

1. **`depguard`, `database/sql` half.** Added `_ "database/sql"` to
   `internal/cli/root.go` — the exact edit STAGE-022 proved passes `go build`,
   `go vet`, `gofmt -l .`, `just test` and `just test-docs`. Result:
   `internal/cli/root.go:17:2: import 'database/sql' is not allowed from list
   'no-sql-in-cli-layer' (depguard)`.
2. **`depguard`, driver-only half.** Replaced it with `_ "modernc.org/sqlite"`
   — the case a straight port of `internal/mcpserver`'s `TestNoSQLImport` walks
   past, because that test greps the literal `"database/sql"` and nothing else.
   Result: `import 'modernc.org/sqlite' is not allowed from list
   'no-sql-in-cli-layer' (depguard)`.
3. **`nolintlint`.** Added a bare `//nolint` above `NewRootCmd`. Result:
   `directive `//nolint` should mention specific linter such as
   `//nolint:my-linter` (nolintlint)`.
4. **`Z7`, the `decisions/` totality assertion.** Added
   `decisions/DEC-999-mutation-probe.md` with `insight.type: bogus`. Confirmed
   the mutant mutated *and stayed invisible*: `scripts/inventory.sh` still
   printed `Decision records | 45` and `reserved | 1` with 47 files on disk, so
   **`X3` would have stayed green**. `Z7` reported
   `the inventory covers 46 of 47 decisions/DEC-*.md files`.

### The scoping check depguard makes, verified both ways

With `!$test` in the rule, depguard reports **0**. With `!$test` removed, it
reports **8 import lines across exactly 5 files** — `coverage_test.go`,
`impact_test.go`, `project_test.go`, `story_test.go`, `wrapped_test.go` —
confirming STAGE-022's count from the import graph rather than from grep. Note
what does **not** appear: `internal/cli/list_test.go`, which a whole-file grep
for `database/sql` *does* hit, because line 275 is a comment. That is the trap
the Design Notes flagged, and depguard is immune to it by construction.

---

## Fork 1 — which linters, and why each one

**Settled: nine, `default: none`, each argued in the config file itself.**

The stage's warning is quantified: `linters.default: all` reports **15,300**
issues on this tree. A short list is not modesty, it is the only thing that
keeps the gate from being disarmed.

### Enabled (9)

| Linter | Findings today | Why it is here |
|---|---:|---|
| `errcheck` | 236 raw → **5** after the exclusions in LD1 | An unchecked error is the bug class AGENTS.md §8 and `errors-wrap-with-context` both presuppose away. |
| `ineffassign` | **0** | A value assigned and never read is a logic error. Standard set, zero config. |
| `staticcheck` | **4** → **1** after `-QF1001` | SA/S/ST correctness and simplification. Standard set. |
| `unused` | **2** | AGENTS.md §8: *"No dead code. Delete, don't comment out."* It found real dead code on day one — see LD8. |
| `depguard` | **0** production / 8 in tests | Mechanises the blocking `no-sql-in-cli-layer` constraint. Fork 4. |
| `errorlint` | **7** | The only linter that checks `%w` and `errors.Is` — i.e. the only mechanism for `errors-wrap-with-context`. |
| `nolintlint` | **0** (zero `//nolint` in the tree) | Guards this gate's own escape hatch, which is the failure mode the stage named. |
| `rowserrcheck` | **3** | An unchecked `rows.Err()` silently **truncates** a result set. `internal/storage` is the only layer that holds state. |
| `sqlclosecheck` | **1** | `rows`/`stmt` closed without `defer` leaks on an early return. Same layer, same reason. |

`rowserrcheck` and `sqlclosecheck` are here **because of the re-measurement**:
both read as zero under golangci-lint's default `uniq-by-line`, and both are
real. They are also the only two picks that are about *this* application rather
than about Go in general.

### Rejected, with the count that decided it

Measured individually with `uniq-by-line: false`, 2026-08-21.

| Linter | Findings | Why not |
|---|---:|---|
| `govet` | 0 | CI already runs `go vet ./...` as its own step, pinned by test-docs `M8`. Running the same analyzers twice under two names doubles the report without adding a check. |
| `gosec` | 65 | 20× G304 (file inclusion via variable) is inherent to a CLI whose `--db`/`--out` take user paths; G306/G301 are file/dir permissions on files the tool exists to write; G204 is the `$EDITOR` exec, which is the feature. Every one would need a `//nolint`. |
| `modernize` | 50 | `strings.SplitSeq`, `slices.Contains`, `errors.AsType`. A 25-file refactor, not a gate. Worth its own spec; not this one. |
| `revive` | 16 | Local identifiers named `min`/`max`/`new`, unused test-helper params, and the two blank driver imports that are *correct and documented*. Style with no bug class. |
| `perfsprint` | 14 | Performance is explicitly out of scope for this stage. |
| `unparam` | 10 | 8 of 10 are deliberately-general test helpers ("`name` always receives `bragfile`"). |
| `forcetypeassert` | 6 | All 6 in tests, where the panic **is** the failure signal. |
| `predeclared` | 5 | Same `min`/`max`/`new` shadowing as `revive`. |
| `nilerr` | 2 | Both are `internal/storage/devguard.go`'s deliberate, commented degrade-to-nil paths (`// let the normal apply path surface the real error`). Enabling it buys exactly two `//nolint`s. |
| `gocritic` | 2 | Both false positives on map keys in `internal/capture/validate_test.go` whose whitespace **is the fixture under test**. |
| `dupword` | 2 | `NULL, NULL,` inside SQL string literals. |
| `errname` | 1 | Would rename the exported `memory.ErrQuery` **type** across four packages to satisfy a naming convention. |
| the wall | — | `wsl` 4826, `wsl_v5` 4062, `paralleltest` 937, `exhaustruct_v5` 767, `goconst` 655, `noinlineerr` 615, `varnamelen` 580, `nlreturn` 447, `lll` 139, `mnd` 54, `funlen` 54, `cyclop` 105, `wrapcheck` 66, `noctx` 65, `tagliatelle` 88, `testpackage` 75, `nonamedreturns` 80, `err113` 49… |

**Zero findings today and still not enabled** — because a linter that guards
nothing this repository has written down is set dressing, and the config is
supposed to be readable: `misspell`, `copyloopvar`, `usetesting`, `unconvert`,
`wastedassign`, `durationcheck`, `makezero`, `nilnesserr`, `bodyclose`,
`asasalint`, `gocheckcompilerdirectives`, `reassign`, `recvcheck`, `exptostd`,
`testableexamples`, `thelper`, `tparallel`, `fatcontext`, `iotamixing`,
`godoclint` — **all 0**. (`ineffassign` and `nolintlint` are also at 0 and *are*
enabled: the first is standard-set and guards a real bug class, the second
guards this gate's own escape hatch. The line is "does it guard a written
rule or a bug class in the core", not "has it fired".)

### The build punch list — 19 findings, all of them fixed, none silenced

The locked config reports exactly these before the code changes, and **0 issues**
after. Applied at design and reverted; the exact edits are in *Notes for the
Implementer*.

| # | Linter | Site |
|---:|---|---|
| 1 | errcheck | `internal/cli/mcp_test.go:120` `cw.Close()` |
| 2 | errcheck | `internal/editor/launch.go:40` `defer os.Remove(path)` |
| 3 | errcheck | `internal/mcpserver/list_filters_test.go:49` `cs.Close()` |
| 4 | errcheck | `internal/mcpserver/server_test.go:42` `cs.Close()` |
| 5 | errcheck | `internal/mcpserver/transport_test.go:42` `io.Copy` |
| 6 | errorlint | `internal/cli/mcp_test.go:58` non-wrapping `%v` |
| 7 | errorlint | `internal/editor/editor.go:76` `err != io.EOF` — **production** |
| 8 | errorlint | `internal/mcpserver/memory.go:54` non-wrapping `%v` — **production** |
| 9–11 | errorlint | `internal/storage/fts_test.go:295,337,470` `== / != sql.ErrNoRows` |
| 12 | errorlint | `internal/storage/migrate_test.go:122` `!= sql.ErrNoRows` |
| 13–15 | rowserrcheck | `internal/storage/store_test.go:1125,1226,1364` `rows.Err` unchecked |
| 16 | sqlclosecheck | `internal/storage/store_test.go:1137` `Close` should use `defer` |
| 17 | staticcheck | `internal/storage/devguard.go:86` ST1005 trailing punctuation |
| 18–19 | unused | `cmd/brag/main.go:26,27` `commit`, `date` |

Two of these are genuine production defects rather than tidiness: `editor.go:76`
compares `err != io.EOF` where a wrapped `io.EOF` would slip past, and
`mcpserver/memory.go:54` drops the error chain with `%v`. Both are one-line
fixes, and both are the constraint `errors-wrap-with-context` catching something
for the first time since it was written on 2026-04-19.

---

## Fork 2 — the coverage floor

**Settled: module-wide, `80.0%`, checked against the `total:` line of
`go tool cover -func`, owned by `scripts/coverage.sh`.**

### The argument

Measured **83.5%**. The floor is **80.0%** — deliberately 3.5 points below.

- **What it protects against.** One thing: a large untested surface landing at
  once — a new package with no tests, or a well-covered package being replaced
  by an uncovered one. At ~3,900 statements, 3.5 points is roughly **135
  statements**, which is bigger than `internal/editor` (61), `cmd/brag` (38) and
  `internal/config` (18) put together. Nothing short of a genuinely large
  untested arrival gets near it.
- **What it deliberately does not protect against.** Drift. A refactor that
  deletes a well-covered helper, or an error branch that lands without a test,
  moves the number a tenth of a point and must not turn CI red. A floor within
  half a point of the measurement is a **ratchet**, and the cheapest way to
  satisfy a ratchet is a filler test — the exact behaviour STAGE-022 forbids.
- **Why not 83%.** That is the measurement, not a floor. A floor equal to
  today's number cannot be cleared tomorrow by anything except more tests.
- **Why not 70%.** A floor only a catastrophe can breach protects nothing and
  reads as a number chosen to be safe.

### Rejected alternatives

- **Per-package floors.** Rejected on `cmd/brag`. A per-package floor there
  would have to be ≤ 23.7%, a number with no meaning that would nonetheless be
  written into the repo and read as a standard. And the failure mode is
  inverted: per-package floors turn *deleting* a small, well-covered file into a
  CI failure. If per-package granularity is ever wanted, the honest form is a
  per-package **report** (which `go test ./...` already prints, and
  `scripts/coverage.sh` shows) with one module-wide gate.
- **A floor on `cmd/brag` specifically.** Rejected: the only way to move it is a
  test that invokes `main()`. 29 statements, 0.74% of the module.
- **`-coverpkg=./...`.** Rejected: reports **86.2%** on the identical suite by
  crediting a package for statements another package's tests happened to run. A
  bigger number for no additional test is the definition of gaming the metric.
- **`-covermode=atomic`.** Rejected: it exists for `-race` runs; CI does not run
  `-race` today, and `set` is what produced every number in this spec. The mode
  is stated explicitly in the script so the number is reproducible.
- **A ratchet that raises the floor automatically to the last measurement.**
  Rejected outright — it is the ratchet argument above, automated.
- **No floor at all, just a reported number.** Rejected: the stage's success
  criterion is a floor, and a reported number with nothing enforcing it is what
  the repo already had for `gofmt` before CI existed.

---

## Fork 3 — the badge

**Settled: no badge, and no coverage service. The number is printed by a
command and enforced by CI; the floor is on the practices page, pinned by a new
test-docs assertion.**

### The argument, and the argument this spec refused to make

The tempting argument is that Codecov violates the repo's no-network identity.
**It does not, and saying so would be false.** `SECURITY.md`'s claim is scoped:
*"the binary makes no network calls of any kind."* A CI job uploading a
coverage profile is not the binary. The claim survives Codecov intact, and this
spec does not get to borrow its authority.

The argument that does hold is the practices page's own **first rule**:

> Every number describing the repository's current state is **derived, not
> typed**.

A badge is a *cached* number, rendered by a third party, that a reader trusts
without being able to check. That is the opposite of the mechanism the page was
built on (`scripts/inventory.sh` + assertion `X3`: recompute the whole table
from the repo and diff). Publishing coverage to a service and pinning an image
to the README would put the single most quotable number on the page outside the
one discipline the page exists to demonstrate.

Two smaller, real costs: an account and a token become a dependency of the
build's green/red status, and a README image is a third-party fetch for every
reader of a repository that otherwise asks nothing of the network.

**What is lost is stated plainly**, on the page and here: a reader now has to
run `just coverage` where a badge would have shown a number. That is a real
trade, not a free win, and the practices page says so in the sentence that
replaces the old "no lint gate and no coverage number" bullet.

### Rejected alternatives

- **Codecov / Coveralls.** Rejected on the derived-not-typed rule above, plus
  account + token + third-party availability coupled to CI status. Explicitly
  **not** rejected on SECURITY.md grounds.
- **A self-generated SVG badge committed to the repo and pinned by an
  assertion.** Genuinely tempting — it is the `X3` idiom applied to a badge, and
  it needs no network. Rejected on maintenance: the number moves on most PRs, so
  a stale-badge assertion would fail most PRs, and the remedy ("regenerate the
  badge") is friction with no information in it. The `X3` idiom works for the
  inventory because those counts move rarely.
- **A badge showing the *floor* rather than the measurement** (`coverage ≥80%`),
  which would be stable and honest. Rejected as a worse form of a sentence the
  page already carries, and one a reader would reasonably mistake for the
  measurement.
- **shields.io with a dynamic JSON endpoint in the repo.** Rejected: still a
  third-party image host, and the JSON is a typed number that rots.
- **Printing coverage into `$GITHUB_STEP_SUMMARY`.** Not rejected on principle,
  just not done: `scripts/coverage.sh` already prints the per-package table and
  the total to stdout, which appears in the job log, and adding a
  GitHub-specific output path would put a second consumer in the script that
  `just coverage` cannot exercise.

---

## Fork 4 — `depguard` for `no-sql-in-cli-layer`, and the scope question

**Settled: yes, this spec takes the mechanism on. It guards production files.
The production-vs-test scope question is decided *as a deferral*, recorded as
`LD7`, logged in `guidance/questions.yaml`, and stated inside the config where
anyone reading the rule will see it. `guidance/constraints.yaml` is not
touched.**

### Why the mechanism belongs here

`depguard` *is* a golangci-lint linter. Putting it in a separate spec means
either a second edit to the same `.golangci.yml` two specs later, or a Go test
that duplicates what the config does — and STAGE-022 already measured that the
Go-test route is the weaker one: `internal/mcpserver`'s `TestNoSQLImport` greps
the literal `"database/sql"` and nothing else, so `internal/cli/story_test.go`'s
`_ "modernc.org/sqlite"` walks straight past a straight port. Verified at design
by mutation check #2 above: depguard catches the driver-only case.

### Why production-only, and why that is a decision rather than a config choice

The constraint's **rule text** is unqualified:

> Files under `internal/cli/` must not import `database/sql` or any SQL driver.

On that literal reading, **5 of the 30 `*_test.go` files** under `internal/cli/`
violate a `blocking` constraint today, across **8 import lines** (verified above
by depguard, not by grep).

Three things point the same way about what the rule was *for*:

1. Its own **rationale** field: *"Architecture principle 2 — CLI is a thin shell
   over storage. Keeps commands testable and future frontends (TUI, API)
   feasible."* A future TUI links `internal/cli`'s production code; it does not
   link its `_test.go` files.
2. `internal/cli/root.go`'s **package comment**, scoped at SPEC-080's re-verify
   and named by STAGE-022 as *"the wording a mechanism should be checked
   against"*: *"Its **production code** imports no SQL driver and no
   database/sql."*
3. `internal/storage/storagetest` exists **only** so CLI tests can reach raw SQL
   without importing it — and its own package comment says so. Four of the five
   offending files predate or bypass that helper.

That is a strong case for amending the rule text. **It is not this spec's to
make**, for the reason STAGE-022 gives: a mechanism must not amend the rule it
mechanises. What this spec does instead is close the half that is not in
dispute, and make the open half *visible* rather than latent:

- `depguard` guards `internal/cli/**` **production** files. That is exactly the
  hole SPEC-080 demonstrated empirically (`_ "database/sql"` in `root.go` passed
  every check the repo had) and it is now closed, both halves, verified twice.
- The `!$test` line in the config carries a comment saying the test half is
  open and where the decision lives.
- The practices page says the same thing, in the same sentence that claims the
  gate.
- `guidance/questions.yaml` gains `no-sql-in-cli-layer-test-scope` with the
  measured evidence, both candidate resolutions, and a resolve trigger.

**This is a narrowing, so it gets tested as a whole claim.** The claim being
made is *"depguard closes the production half of `no-sql-in-cli-layer` for
`internal/cli`"* — not *"depguard enforces `no-sql-in-cli-layer`."* The second
sentence would be false and this spec does not write it anywhere.

### Rejected alternatives

- **Port `TestNoSQLImport` into `internal/cli`.** Rejected: measured
  driver-blind (STAGE-022, re-verified here), and it works on file text, which
  means it must special-case `root.go`'s package comment — the exact
  false-positive the Design Notes warned about. `internal/mcpserver`'s copy
  sidesteps it by excluding itself by filename; a port would need the same care
  and would still miss the driver.
- **Both: depguard *and* a ported Go test.** Rejected as two mechanisms for one
  rule with different blind spots, which is how they drift.
- **Rework the five test files through `storagetest` and enforce the constraint
  as literally written.** This is the option that would make the rule true as
  stated, and it is not crazy: `story_test.go` needs only a dead blank import
  deleted, and the other four backdate rows via `sql.Open` + `UPDATE`, which
  `storagetest.Backdate` half-covers already (it rewrites `created_at`; the
  callers also need `updated_at`, and `project_test.go` needs the `projects`
  table). Rejected **here** on scope: it is a five-file test refactor plus two
  new helpers, riding into a CI-wiring spec, and it *presumes* the answer to the
  scope question just as firmly as amending the constraint would — only in the
  other direction. It is written down in the question entry as candidate B so
  the next spec does not re-derive it.
- **Amend `guidance/constraints.yaml` to say "production code".** Rejected as a
  side effect, explicitly, per STAGE-022. Recorded as candidate A.
- **Scope depguard to `internal/cli/**` including tests and accept a red gate.**
  Rejected: a gate that is red on arrival is not a gate.
- **Extend depguard to `internal/mcpserver` too** (which has a Go test but a
  driver-blind one). Rejected: the constraint's `paths` glob is
  `internal/cli/**`, and widening a blocking constraint's reach is the same
  class of side effect as narrowing it. Noted in the question entry.

---

## Outputs

### New files (2)

1. **`.golangci.yml`** — 124 lines. Literal in *Notes for the Implementer*.
2. **`scripts/coverage.sh`** — 46 lines, mode `0755`. Literal below.

### Modified files (18)

3. **`.github/workflows/ci.yml`** — two new jobs (`lint`, `coverage`); the
   `test` job is untouched. Full literal below.
4. **`justfile`** — two recipes, `lint` and `coverage`.
5. **`scripts/test-docs.sh`** — new **Group Z** (`Z1`–`Z7`), and `X7` switched
   from `assert_line_count_band` to `assert_word_count_band` (LD9).
6. **`docs/engineering-practices.md`** — three prose hunks + the regenerated
   inventory block. **The derivation outranks the cache:** build MUST run
   `just inventory` and paste the output between the markers. **Three** rows
   are expected to move — `Documentation assertions (distinct ids)`
   **177 → 184** from Group Z, and, because output 8 adds a question,
   `Questions tracked in guidance/questions.yaml` **18 → 19** and
   `…of those, still open` **6 → 7**. If `just inventory` prints anything else,
   the paste wins and this spec's arithmetic is what was wrong.
7. **`AGENTS.md`** — §3 Tech Stack, the `Linter / Formatter` line, which
   currently says *"`golangci-lint` welcome but not required in CI yet"* and is
   falsified by this spec.
8. **`guidance/questions.yaml`** — one new entry,
   `no-sql-in-cli-layer-test-scope`.
9. **`cmd/brag/main.go`** — delete the `commit` and `date` vars (LD8).
10. **`.goreleaser.yaml`** — delete the two `-X` ldflags that fed them (LD8).
11–18. The lint punch list, one to three lines each:
    `internal/cli/mcp_test.go`, `internal/editor/editor.go`,
    `internal/editor/launch.go`, `internal/mcpserver/list_filters_test.go`,
    `internal/mcpserver/memory.go`, `internal/mcpserver/server_test.go`,
    `internal/mcpserver/transport_test.go`, `internal/storage/devguard.go`,
    `internal/storage/fts_test.go`, `internal/storage/migrate_test.go`,
    `internal/storage/store_test.go`.

### Premise audit (§9), run at design against the repo

Three audits, each executed, not merely enumerated.

1. **Additive case — the tracked collection is test-docs assertion ids.**
   `scripts/inventory.sh` derives `Documentation assertions (distinct ids)`
   from the script text. Adding `Z1`–`Z7` moves it. Ran `./scripts/inventory.sh`
   with `Z1`–`Z6` staged: **183**; with `Z7` alone staged: **178**. Both consistent
   with a base of 177, so the combined value is **184**. `X3` fails until the
   block is re-pasted — confirmed by running the harness in that state.
   **The same audit catches a second tracked collection this spec adds to:**
   `guidance/questions.yaml`. `inventory.sh` derives two rows from it, anchored
   on `^  - id: ` (**18** today) and `^    status: open$` (**6** today), so the
   new `no-sql-in-cli-layer-test-scope` entry moves both to **19** and **7**.
   Verified at design by parsing the appended literal: 19 questions, no
   duplicate ids. Three inventory rows move in total, not one.
2. **Status-change case — every doc claim about lint/coverage.**
   `grep -rn 'golangci' --include='*.md'` outside `projects/` returns exactly
   three hits: `AGENTS.md:76` (fixed, item 7), `docs/engineering-practices.md`
   (fixed, item 6), and `NEXT-SESSION-PROMPT.md:79` — a session scratch prompt,
   not a repository claim, deliberately left alone.
3. **Inversion case — existing assertions whose premise this spec changes.**
   `M6`/`M7`/`M8` pin `go test ./...`, `gofmt -l .` and `go vet ./...` inside
   `ci.yml`; the `test` job is left byte-identical so all three still hold —
   verified by running the full harness against the new workflow. `L9` pins
   `-X main.version=` in `.goreleaser.yaml`; only `main.commit` and `main.date`
   are removed, so `L9` holds — also verified. `X7`'s premise *is* inverted, by
   this spec's own edit; that is LD9.

### §12 NOT-contains self-audit

This spec adds **no** NOT-contains assertions. Group Z is entirely positive or
derive-and-diff. The one pre-existing NOT-contains that the new prose could
trip is `X5` (`rigorous|comprehensive|world-class|best-in-class|battle-tested|
cutting-edge|state-of-the-art` on the practices page) — the three new hunks were
staged and the harness run: `X5` green.

---

## Locked design decisions

**LD1 — `errcheck` is enabled with an enumerated exclusion list, not an
exclusion preset.** 236 findings unconfigured; 231 of them are two idioms
(`fmt.Fprint*` to the command's own writer; deferred `Close`/`Rollback`). The
eight `exclude-functions` entries are written out by name so a reader can weigh
each, rather than `exclusions.presets: [std-error-handling]`, which achieves the
same suppression behind a name the reader has to go look up. The remaining 5 are
fixed in code.

**LD2 — `errorlint` is enabled because it is the only mechanism
`errors-wrap-with-context` has ever had.** That constraint was written
2026-04-19 and has been enforced by review since. It catches 7 things, two of
them in production code.

**LD3 — `no-new-top-level-deps-without-decision` does not apply, read rather
than assumed.** Its `paths` are `["go.mod", "go.sum"]`; golangci-lint is
installed by the CI action and never enters the module graph. Confirmed
empirically: `go install …@v2.13.1` was run at design from `/tmp` and
`git status` stayed clean. **The alternative that *would* have tripped it is
named in LD4's rejected list.**

**LD4 — CI obtains golangci-lint from the official action, SHA-pinned, with the
version pinned in the workflow.**
`golangci/golangci-lint-action@ba0d7d2ec06a0ea1cb5fa41b2e4a3ab91d21278a  # v9.3.0`
with `version: v2.13.1`. SHA-pinning matches the two actions already in this
workflow. The action's default `verify: true` also validates `.golangci.yml`
against golangci-lint's JSON schema, so a malformed config fails loudly rather
than silently linting less. **`issues.uniq-by-line: false`** is set in the
config, for the reason measured above.

*Rejected:* `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1`
in the workflow — upstream explicitly discourages `go install` for this tool,
and the symptom is observable: the binary built that way at design reports
`built with go1.26.6 from (unknown, modified: ?)`. It also rebuilds ~2–3 minutes
on every PR with no binary cache. *Rejected:* a `tool` directive in `go.mod` —
this **would** trip `no-new-top-level-deps-without-decision` and would pull the
linter's transitive module graph into the application's. *Rejected:*
`brew install golangci-lint` — unpinned and macOS-only.

**LD5 — the coverage floor is module-wide `80.0%`, defined as the `total:` line
of `go tool cover -func` over `go test ./... -covermode=set` with no
`-coverpkg`, and it lives in exactly one place: `FLOOR=` in
`scripts/coverage.sh`.** CI calls that script; `just coverage` calls that
script; assertion `Z5` reads the value back out and asserts the practices page
states the same number. Full argument in Fork 2.

**LD6 — `depguard` is the mechanism for `no-sql-in-cli-layer`.** Not a ported
`TestNoSQLImport`, and not both. It reads the import list, so it is immune to
the package-comment false positive; it names the driver, so it catches the case
a `"database/sql"` grep misses. Both halves mutation-checked at design.

**LD7 — the depguard rule is scoped to production files (`!$test`), and the
production-vs-test scope question is deferred as a decision, not answered as a
config choice.** `guidance/constraints.yaml` is **not** amended. The deferral is
recorded in three places a future reader will actually hit: the `!$test` line's
own comment, the practices page sentence that claims the gate, and a new
`guidance/questions.yaml` entry carrying both candidate resolutions and the
measured evidence. Full argument in Fork 4.

**LD8 — `unused`'s two findings are fixed by deleting the dead code, and the
two goreleaser ldflags that fed them go with it.** `main.commit` and
`main.date` have been written by the linker and read by nothing since SPEC-023
(2026-04-26). AGENTS.md §8: *"No dead code. Delete, don't comment out."*
Leaving `-X main.commit=` pointing at a symbol that no longer exists would be
silently ignored by the Go linker and actively misleading to a reader, so both
lines go.

*The distribution downsides pass, run because AGENTS.md §4 demands one for any
diff near the packaging blocks:* `builds.ldflags` is **not** one of §4's
enumerated packaging blocks (`brews:`, `homebrew_casks:`, `nfpms:`, `archives:`,
`signs:`/`notarize:`), the artifact type is unchanged, the tap is unchanged, the
install path is unchanged, and the clean-host trust path is untouched — so no
category in `docs/distribution-decisions.md`'s trust matrix moves.
`-X main.version=` — the one that is read, and the one test-docs `L9` pins —
stays. `goreleaser check` was run before and after and returns the **identical**
result.

*A separate, pre-existing finding surfaced by that run, deliberately not fixed
here:* `goreleaser check` fails on **`DEPRECATED: brews should not be used
anymore`** (goreleaser 2.17.1) — with and without this spec's edit. Per
AGENTS.md §4 that is precisely the class of change that must **not** be treated
as hygiene: the v0.1.0 formula→cask switch was a four-line deprecation fix that
silently changed the artifact's OS-trust path. It is named here so it is not
lost and is explicitly **out of scope**; it wants its own spec running the
`docs/distribution-decisions.md` checklist.

*Rejected:* keep the vars and add `//nolint:unused` with an explanation —
tempting, because `nolintlint` would make the directive specific and explained,
and a linker-injected symbol is a legitimate blind spot for `unused`. Rejected
because these two symbols are not linker-injected-and-read, they are
linker-injected-and-ignored: there is no blind spot, the code is simply dead.
*Rejected:* wire them into `brag --version` output so they are used — that is
feature work, and a user-visible output change, arriving inside a CI spec.
*Rejected:* exclude `cmd/brag/main.go` from `unused` — an exclusion to preserve
dead code.

**LD9 — `X7` switches from `assert_line_count_band` to `assert_word_count_band`,
band `1800 2700`.** Not a judgement call: SPEC-081 built the helper and wrote
the trigger — *"switch a caller to this helper when its swing exceeds its
headroom, not before"* — and named `X7` as the next closest caller (headroom 17
lines, measured reflow swing +15). This spec's edit to the practices page takes
it to **301 lines**, one over the 300 ceiling; the harness was run in that state
and `X7` failed exactly there. The band tightens the guard: converted at the
file's own 7.93 words/line, the old `150..300` line band was ~1,190…2,380 words,
an admissible span of ~1,190; `1800..2700` is a span of **900 — 24% narrower** —
with 232 words of ceiling headroom against a reflow swing of at most ±4 words
(only the `-`/`>` prefixes on wrapped list items move). Measured after the edit:
**2,468 words**, comfortably inside.

*Rejected:* raise the line band to 340. That is the
bump-the-number-until-green disarming SPEC-079's LD5 refused and SPEC-081
codified against. *Rejected:* shrink the edit until it fits under 300. The
sentence being added is the one the stage requires, and cutting a true claim to
fit a guard is the guard driving the document.

**LD10 — `Z7` closes STAGE-022's routed `decisions/` totality gap, and the
meaning of a violation is a hard fail.** SPEC-080 deferred it because the
assertion could not be written without deciding what an untyped `DEC-*.md`
*means*: a typo that should fail the build, or a legitimately new category
`inventory.sh` has not been taught. **They differ in intent, not in
consequence** — either way the file is counted by neither row, vanishes from
the page, and `X3` stays green because the page is generated from the same two
filters that lost it (proved by mutation check #4). This harness has no warning
tier; every assertion is `ok`/`fail`/`skip`. So it fails, and the failure
message names both remedies: *fix its front-matter, or teach
`scripts/inventory.sh` a row for the new type.* `decisions/_template.md` stays
outside every count, as before, because it does not match `DEC-*.md`.

*Rejected:* a `skip` when the sum is short — a warning tier invented for one
assertion, which the next author has to learn. *Rejected:* teaching
`inventory.sh` an "other" bucket — it would make the sum always balance and the
assertion permanently green, which is the same as not writing it.

### Rejected alternatives (build-time)

Things build might reasonably reach for, decided here instead:

- **Do not add `//nolint` anywhere.** The tree has zero directives today,
  `nolintlint` is enabled, and every one of the 19 findings has a code fix
  specified below. If a fix does not work, that is a spec defect to raise, not a
  directive to add.
- **Do not widen `errcheck`'s `exclude-functions`.** The list is closed at eight
  entries. A ninth entry means a finding was suppressed instead of fixed.
- **Do not switch the `test` job to `go test ./... -coverprofile=…`.** Coverage
  is a separate job on one OS, on purpose (Fork 2 / LD5). Editing the `test`
  job's `go test ./...` line risks `M6`, for no gain.
- **Do not run the `lint` job on the matrix.** One OS; the reason is in the
  workflow comment.
- **`internal/storage/store_test.go` site 1: use `defer rows.Close()` plus a
  `rows.Err()` check**, not a closure and not `rows.Close()` retained alongside.
  Verified at design that the whole suite stays green with the rows held open
  across the remaining read-only queries in that test.
- **Do not reorder or re-comment `.golangci.yml`.** The comments are the
  deliverable — the stage's requirement is that the config be legible to the
  same reader as the practices page. Transcribe verbatim.

---

## Acceptance Criteria

1. `golangci-lint run` (v2.13.1, with the repo's `.golangci.yml`) reports
   **0 issues**.
2. `golangci-lint config verify` exits 0.
3. Adding `_ "database/sql"` **or** `_ "modernc.org/sqlite"` to
   `internal/cli/root.go` makes `golangci-lint run` fail with a `depguard`
   finding naming `no-sql-in-cli-layer`. (Both, separately.)
4. Adding a bare `//nolint` anywhere makes `golangci-lint run` fail with a
   `nolintlint` finding.
5. `./scripts/coverage.sh` exits 0 and prints `total: 83.5%   floor: 80.0%`
   (the total may differ by a tenth; it must be ≥ 80.0).
6. `./scripts/coverage.sh` with `FLOOR` temporarily set above the measured value
   exits **1** and prints the floor message. (Restore `FLOOR=80.0`.)
7. `just lint` and `just coverage` both run.
8. `actionlint .github/workflows/ci.yml` exits 0.
9. `just test`, `gofmt -l .` (empty), `go vet ./...` all clean.
10. `just test-docs` is **ALL OK**, and `./scripts/inventory.sh` reports
    **184** documentation assertions, **19** questions tracked and **7** open —
    with the inventory block on `docs/engineering-practices.md` re-pasted from
    `just inventory`, not hand-edited. If any of the three derived values
    differs, paste what the script printed and raise the discrepancy; do not
    hand-edit the block toward this spec.
11. `goreleaser check` returns the same result as before the change (the
    pre-existing `brews` deprecation, and nothing new).
12. `guidance/constraints.yaml` is **unchanged**. `git diff` on it must be empty.
13. No `//nolint` directive exists anywhere in the tree:
    `grep -rn 'nolint' --include='*.go' .` returns nothing outside
    `.claude/worktrees/`.

## Failing Tests

Written at design, made to pass at build. Seven new assertion ids, one changed
assertion, and four executable mutation checks.

### New — `scripts/test-docs.sh`, Group Z (`Z1`–`Z7`)

Full literal in *Notes for the Implementer*. Each would fail today:

| Id | Asserts | Fails today because |
|---|---|---|
| `Z1` | `.golangci.yml` exists | it does not |
| `Z2` | `linters.default: none` **and** exactly the nine locked linters, derived from the file and diffed against the expected sorted list | no config |
| `Z3` | the depguard rule names `no-sql-in-cli-layer`, scopes to `internal/cli/*.go`, carries `!$test`, and denies **both** `database/sql` and `modernc.org/sqlite` | no config |
| `Z4` | `ci.yml` runs `golangci/golangci-lint-action@` **and** `./scripts/coverage.sh` | CI runs neither |
| `Z5` | the `FLOOR=` value in `scripts/coverage.sh` appears, suffixed `%`, in `docs/engineering-practices.md` | neither the script nor the claim exists |
| `Z6` | the `justfile` has `^lint:` and `^coverage:` recipes | neither exists |
| `Z7` | `decisions` + `reservation` counts **sum to** `ls decisions/DEC-*.md \| wc -l` | passes today (45+1=46) — it fails on the mutant, see below |

`Z7` is the one assertion here that is green on arrival. That is deliberate and
is the whole reason SPEC-080 deferred it: it has no current victim, and its
value is entirely preventive. Its teeth are demonstrated by the mutation check,
not by an initial red.

### Changed — `X7`

`assert_line_count_band "X7" "$PRACTICES_DOC" 150 300` →
`assert_word_count_band "X7" "$PRACTICES_DOC" 1800 2700`.
**Would fail without the change:** with this spec's practices-page edits applied
and `X7` left as-is, the harness reports
`FAIL: X7: docs/engineering-practices.md has 301 lines (expected 150..300)`.
Verified at design.

### Changed — `X3`

Not edited, but it **will fail** until `just inventory` is re-run and pasted:
`Documentation assertions (distinct ids)` moves 177 → 184, and the new
`guidance/questions.yaml` entry moves two more rows (18 → 19 tracked, 6 → 7
open). Verified at design by running the harness in exactly that state. The
remedy is mechanical and is in `X3`'s own failure message.

### Mutation checks (run by build, recorded in Build Completion)

- **M-A** `_ "database/sql"` in `internal/cli/root.go` → depguard fires. Revert.
- **M-B** `_ "modernc.org/sqlite"` in `internal/cli/root.go` → depguard fires.
  Revert.
- **M-C** a bare `//nolint` in `internal/cli/root.go` → nolintlint fires. Revert.
- **M-D** `FLOOR=90.0` in `scripts/coverage.sh` → the script exits 1 with the
  floor message. Revert to `80.0`.
- **M-E** a `decisions/DEC-999-*.md` with `insight.type: bogus` → `Z7` fails
  **and** `scripts/inventory.sh` still prints `Decision records | 45`, proving
  `X3` would not have caught it. Delete the probe file.

Each mutation must be confirmed to have *mutated* — i.e. the tree really changed
— before the failure is credited. All five were run at design.

### Decision-to-test mapping (§9)

Every locked decision has an assertion or a mutation check that fails without it:

| Decision | Pinned by |
|---|---|
| LD1 (errcheck enumerated exclusions) | `Z2` (errcheck enabled) + AC-1 (0 issues) |
| LD2 (errorlint) | `Z2` |
| LD3 (constraint does not apply) | AC-12 (`constraints.yaml` unchanged) + `go.mod` untouched |
| LD4 (action, SHA-pinned, `uniq-by-line: false`) | `Z4`, AC-2, AC-8 |
| LD5 (module-wide 80.0 floor, one home) | `Z5`, `Z6`, AC-5, AC-6 / **M-D** |
| LD6 (depguard, both halves) | `Z3`, AC-3 / **M-A**, **M-B** |
| LD7 (production-only scope, constraint untouched) | `Z3` (`!$test` present), AC-12 |
| LD8 (delete the dead vars + ldflags) | AC-1 (`unused` clean), AC-11, test-docs `L9` |
| LD9 (`X7` on words) | `X7` itself, at band `1800 2700` |
| LD10 (`Z7` hard-fails) | `Z7` / **M-E** |
| nolintlint has teeth | AC-4, AC-13 / **M-C** |

---

## Implementation Context

### Decisions that apply

- **DEC-001 — pure-Go SQLite driver.** `modernc.org/sqlite` is the only driver
  in `go.mod`, which is why depguard's deny list can name it. The list is a
  list, not a category; a second driver would have to be added.
- **DEC-026 — dev/prod migration guard.** The ST1005 fix edits that guard's
  user-facing error message (dropping one trailing period). No test asserts the
  message tail — checked at design — and the behaviour is unchanged.

### Constraints that apply

- **`no-sql-in-cli-layer` (blocking)** — mechanised, production half, by
  depguard. **Do not edit `guidance/constraints.yaml`.** See LD7.
- **`errors-wrap-with-context` (warning)** — mechanised by errorlint. Two of
  its seven findings are the first production violations it has ever surfaced.
- **`no-new-top-level-deps-without-decision` (warning)** — does not apply; see
  LD3. `go.mod` and `go.sum` must be unchanged by this spec.
- **`test-before-implementation` (blocking)** — Group Z and the `X7` change are
  written first; the artifacts follow.
- **`one-spec-per-pr` (blocking)** — one PR, SPEC-082.
- **`storage-tests-use-tempdir` (blocking)** — unaffected; the storage test
  edits add `rows.Err()` checks and an import, no path changes.

### Prior related work

- **SPEC-080** left this spec two things: the `no-sql-in-cli-layer` gap with its
  full measurement, and the `decisions/` totality gap with STAGE-022 naming
  *"this stage's lint-and-CI spec"* as one of its two triggers. Both are closed
  here.
- **SPEC-079** built the derived-not-typed discipline (`scripts/inventory.sh` +
  `X3`) that Fork 3's badge decision rests on, and `X7`, which LD9 changes.
- **SPEC-081** built `assert_word_count_band` and wrote the switch trigger LD9
  now hits — including naming `X7` as the next caller in line.
- **SPEC-023** introduced the `-X main.commit` / `-X main.date` ldflags that
  LD8 removes.
- **SPEC-073** is the cautionary case the whole stage is organised around: a
  coverage sentence that was wrong four times.

### Out of scope (for this spec specifically)

- **Benchmarks and any scale/perf harness** — PROJ-009.
- **Raising coverage by writing tests.** Nothing in the punch list adds a test
  for coverage's sake; the three `rows.Err()` additions are assertions the
  linter demanded, and coverage is unchanged at 83.5% after all 19 fixes
  (measured).
- **The `Entries:` envelope inconsistency** — STAGE-022's second spec.
- **The five stale-comment items routed from STAGE-021** — `store.go`'s `Store`
  type comment, `list_test.go:275`, V1's noun, A1's "smallest deep-dive doc",
  A11's stale A1 reference. Note that `list_test.go:275` is *about* this spec's
  subject (it claims CLI tests cannot import `database/sql`, which four files
  disprove) — it is still not fixed here, because it is a one-word fix that
  belongs with whichever spec next opens that file, and because rewording it
  correctly depends on the answer to LD7's deferred question.
- **The `goreleaser` `brews` deprecation** surfaced by the design pre-flight —
  see LD8. It needs `docs/distribution-decisions.md` run against it.
- **Running `just test-docs` in CI.** Still owned by nothing; the practices
  page continues to say so.
- **`NEXT-SESSION-PROMPT.md`** — a session scratch file, not a repository claim.

---

## Notes for the Implementer

Everything below is a literal. Transcribe verbatim; verify diffs. Each was
written into the working tree at design, run through its tool, and reverted.

### 1. `.golangci.yml` (new, 124 lines)

```yaml
# .golangci.yml — the lint gate. Nine linters, each chosen; nothing inherited.
#
# WHY A SHORT LIST. Run with `linters.default: all`, golangci-lint reports
# **15,300** issues on this tree (measured 2026-08-21, v2.13.1). A gate that
# size is not a gate: it gets silenced with `//nolint` and then means nothing.
# Every linter below is enabled because it guards a rule this repository has
# already written down, or a bug class in the one layer that holds state. The
# measured finding count for every linter considered and REJECTED is in
# SPEC-082 — including the ones that report zero today.
#
# HOW TO RUN IT. `just lint` locally, and the `lint` job in
# .github/workflows/ci.yml on every pull request. The linter version is
# pinned in that workflow.

version: "2"

linters:
  default: none
  enable:
    # --- standard-set linters that earn their place here ---
    - errcheck        # an unchecked error is the bug class that AGENTS.md §8
                      # and the errors-wrap-with-context constraint both
                      # presuppose away. See settings.errcheck for what is
                      # excluded and why.
    - ineffassign     # a value assigned and never read is a logic error, not
                      # a style opinion. Zero findings today; zero config.
    - staticcheck     # the SA (correctness) / S (simplify) / ST (style) suite.
                      # The QF (quickfix) family is excluded below, with reason.
    - unused          # AGENTS.md §8: "No dead code. Delete, don't comment out."

    # --- linters that mechanise a rule this repo already wrote down ---
    - depguard        # guidance/constraints.yaml `no-sql-in-cli-layer`,
                      # severity BLOCKING, which had no automated guard for the
                      # package its own path glob names. See settings.depguard.
    - errorlint       # the errors-wrap-with-context constraint is about %w and
                      # errors.Is; errorlint is the only linter that checks
                      # either one.
    - nolintlint      # guards this gate's own escape hatch. There are zero
                      # //nolint directives in the tree today; this keeps any
                      # future one specific, explained, and load-bearing.

    # --- linters chosen for what this app is ---
    # An embedded-SQLite CLI. internal/storage is the only layer that holds
    # state, and every list / search / digest path reads rows out of it.
    - rowserrcheck    # an unchecked rows.Err() silently TRUNCATES a result set:
                      # the query "succeeds" and returns fewer entries.
    - sqlclosecheck   # rows/stmt closed without defer leaks on an early return.

  settings:
    depguard:
      rules:
        # Mechanises: "Files under internal/cli/ must not import database/sql
        # or any SQL driver. All persistence goes through internal/storage."
        # (guidance/constraints.yaml, no-sql-in-cli-layer, severity: blocking.)
        #
        # WHY DEPGUARD AND NOT A GREP. The package comments that DESCRIBE this
        # boundary contain the literal strings, so a whole-file grep
        # false-positives on the documentation of the rule
        # (internal/cli/root.go). depguard reads the import list, not the text.
        #
        # WHY THE DENY LIST NAMES A DRIVER. The constraint says "or any SQL
        # driver" — a category no import-path list can express. This is a LIST:
        # a second driver would have to be added to it. modernc.org/sqlite is
        # the only one in go.mod (DEC-001), and a driver imported WITHOUT
        # database/sql is exactly the case a "database/sql" grep walks past.
        #
        # WHY PRODUCTION FILES ONLY (`!$test`). The production-vs-test scope
        # question is decided in SPEC-082 LD7, not here. Five of the 30
        # *_test.go files under internal/cli/ carry eight offending imports
        # today; that half of the constraint is open, named, and unguarded.
        no-sql-in-cli-layer:
          files:
            - "**/internal/cli/*.go"
            - "!$test"
          deny:
            - pkg: "database/sql"
              desc: "internal/cli is a thin shell over internal/storage; persistence goes through *storage.Store (guidance/constraints.yaml no-sql-in-cli-layer)"
            - pkg: "modernc.org/sqlite"
              desc: "the SQL driver is imported by internal/storage only (DEC-001, guidance/constraints.yaml no-sql-in-cli-layer)"

    errcheck:
      # Unconfigured, errcheck reports 236 issues on this tree. 231 of them are
      # two idioms where the returned error is unactionable at the call site.
      # They are listed here BY NAME rather than hidden behind an exclusion
      # preset, so a reader can weigh each one instead of looking it up:
      exclude-functions:
        # (1) Writes to the command's own io.Writer. The CLI has no logging
        # channel to report a failed write to (AGENTS.md §8: "Logging: none"),
        # and the writer is a *bytes.Buffer in every test.
        - fmt.Fprint
        - fmt.Fprintf
        - fmt.Fprintln
        # (2) Deferred cleanup. A failed Close on a read path, or a Rollback
        # after a successful Commit (which always returns sql.ErrTxDone),
        # carries nothing a caller can act on.
        - (*database/sql.DB).Close
        - (*database/sql.Rows).Close
        - (*database/sql.Tx).Rollback
        - (*os.File).Close
        - (*github.com/jysf/bragfile000/internal/storage.Store).Close

    nolintlint:
      require-explanation: true   # a directive must say why
      require-specific: true      # bare `//nolint` is not allowed; name the linter
      allow-unused: false         # a directive that suppresses nothing is deleted

  exclusions:
    rules:
      # QF1001 "could apply De Morgan's law" is a quickfix SUGGESTION, not a
      # correctness check. All three sites are ordering assertions of the form
      # `if !(a < b && b < c)`, where the De Morgan inverse reads worse.
      - linters:
          - staticcheck
        text: "QF1001"

issues:
  # golangci-lint reports ONE issue per line by default and silently drops the
  # rest. Measured 2026-08-21: `linters.default: all` reports 8,713 issues with
  # that default and 15,300 with it off — the default hides 43% of them. A repo
  # whose documentation rule is "every number is derived, not typed" does not
  # ship a linter configured to under-report.
  uniq-by-line: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

### 2. `scripts/coverage.sh` (new, 46 lines, `chmod +x`)

```bash
#!/usr/bin/env bash
# scripts/coverage.sh — measure Go statement coverage and enforce the floor.
#
# SINGLE SOURCE for three consumers:
#   1. `just coverage` (local),
#   2. the `coverage` job in .github/workflows/ci.yml, and
#   3. test-docs assertion Z5, which reads FLOOR out of this file and asserts
#      docs/engineering-practices.md states the same number.
# The floor is written down once, here. Nothing else may restate it.
#
# THE UNIT — because "coverage" is not one number. This script means exactly
# one thing by it: the `total:` line of `go tool cover -func` over a profile
# from `go test ./... -covermode=set`, with NO -coverpkg. Each package is
# credited only for what its OWN tests execute. Adding `-coverpkg=./...`
# reports 86.2% on the same tree (measured 2026-08-21) because it credits a
# package for statements some other package's tests happened to run — a bigger
# number for no additional test, so it is not the one enforced.
#
# IT IS A FLOOR, NOT A TARGET. It sits BELOW the measured value on purpose,
# with headroom, so that ordinary refactoring never fails CI and nobody is ever
# rewarded for writing a test that exists to move a percentage. It catches one
# thing: a large untested surface landing at once. See SPEC-082 LD5.

set -eu

FLOOR=80.0

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
cd "$REPO_ROOT"

PROFILE=$(mktemp)
trap 'rm -f "$PROFILE"' EXIT

go test ./... -covermode=set -coverprofile="$PROFILE"

total=$(go tool cover -func="$PROFILE" | awk '/^total:/ {print $NF}')
pct=${total%\%}

printf '\ntotal: %s   floor: %s%%\n' "$total" "$FLOOR"

if awk -v p="$pct" -v f="$FLOOR" 'BEGIN { exit !(p < f) }'; then
    printf 'coverage %s is below the floor of %s%% — see docs/engineering-practices.md\n' \
        "$total" "$FLOOR" >&2
    exit 1
fi
```

### 3. `.github/workflows/ci.yml` (full replacement, 91 lines)

The `test` job is byte-identical to what is there today — `M6`/`M7`/`M8` depend
on it. Only the header comment and the two new jobs are new.

```yaml
# .github/workflows/ci.yml — PR + main-push gating
# Runs gofmt, go vet, go test on macOS-latest + ubuntu-latest with
# the Go version declared in go.mod, plus golangci-lint and the
# coverage floor on ubuntu-latest.

name: ci

on:
  pull_request:
  push:
    branches:
      - main

permissions:
  contents: read

jobs:
  test:
    strategy:
      fail-fast: false
      matrix:
        os: [macos-latest, ubuntu-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - name: Checkout
        uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1  # v7.0.1

      - name: Set up Go
        uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e  # v7.0.0
        with:
          go-version-file: go.mod
          check-latest: true
          cache: true

      - name: Verify gofmt
        run: |
          unformatted=$(gofmt -l .)
          if [ -n "$unformatted" ]; then
            printf 'gofmt -l reports unformatted files:\n%s\n' "$unformatted" >&2
            exit 1
          fi

      - name: Run go vet ./...
        run: go vet ./...

      - name: Run go test ./...
        run: go test ./...

  # golangci-lint. One OS is enough: the linters are static analysis over the
  # same source, and the two OS-specific surfaces this repo has (path handling
  # and the $EDITOR launch) are exercised by the matrix above.
  lint:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1  # v7.0.1

      - name: Set up Go
        uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e  # v7.0.0
        with:
          go-version-file: go.mod
          check-latest: true
          cache: true

      # The golangci-lint version is pinned HERE and nowhere else; match it
      # locally to reproduce a CI failure. The action's default `verify: true`
      # also validates .golangci.yml against golangci-lint's own JSON schema,
      # so a malformed config fails loudly instead of silently linting less.
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@ba0d7d2ec06a0ea1cb5fa41b2e4a3ab91d21278a  # v9.3.0
        with:
          version: v2.13.1

  # Coverage. scripts/coverage.sh owns both the floor and the definition of the
  # number; neither is restated here, so the workflow and the local
  # `just coverage` cannot disagree.
  coverage:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1  # v7.0.1

      - name: Set up Go
        uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e  # v7.0.0
        with:
          go-version-file: go.mod
          check-latest: true
          cache: true

      - name: Measure coverage and enforce the floor
        run: ./scripts/coverage.sh
```

### 4. `justfile` — insert immediately after the `test-hook` recipe

```make
# Run the lint gate (.golangci.yml). Needs golangci-lint v2.13.1 — the version
# pinned in .github/workflows/ci.yml — to reproduce CI exactly.
lint:
    @golangci-lint run

# Measure Go statement coverage and fail below the floor (scripts/coverage.sh).
coverage:
    @./scripts/coverage.sh
```

### 5. `scripts/test-docs.sh` — replace the `X7` block

Old (delete):

```sh
# X7 — the page stays an INDEX, not an essay. Same idiom as A1/C2/D2. The
# stated failure mode for this document is that it turns into new prose about
# the project instead of a route to artifacts that already exist; a length band
# is the cheap mechanical form of that. Band, not a tight pin, because unlike
# A1 there is no agreed length for a brand-new artifact to be re-pinned to.
assert_line_count_band "X7" "$PRACTICES_DOC" 150 300
```

New:

```sh
# X7 — the page stays an INDEX, not an essay. Same idiom as A1/C2/D2. The
# stated failure mode for this document is that it turns into new prose about
# the project instead of a route to artifacts that already exist; a size band
# is the cheap mechanical form of that.
#
# WORDS, NOT LINES, since SPEC-082. The switch trigger is the one SPEC-081
# wrote when it built assert_word_count_band: "switch a caller to this helper
# when its swing exceeds its headroom, not before." X7 was named there as the
# next closest caller — headroom 17 lines, measured reflow swing +15 — and
# SPEC-082's own edit to this page took it to 301 lines, one over the old
# ceiling. The alternative was to raise the line band, which is the
# bump-the-number-until-green disarming that SPEC-079 LD5 refused.
#
# The band tightens the guard rather than loosening it. Converted at this
# file's own 7.93 words/line, the old 150..300 line band was ~1,190..2,380
# words, an admissible span of ~1,190. 1800..2700 is a span of 900 — 24%
# narrower — with 232 words of ceiling headroom, against a reflow swing of
# at most ±4 words (only the `-`/`>` prefixes on wrapped list items move).
assert_word_count_band "X7" "$PRACTICES_DOC" 1800 2700
```

### 6. `scripts/test-docs.sh` — insert Group Z immediately before `# ===== finalise =====`

```sh
# ===== Group Z — lint gate + coverage floor (SPEC-082) =====

GOLANGCI=".golangci.yml"
COVERAGE_SH="scripts/coverage.sh"

# Z1 — the lint config exists at all. Before SPEC-082 there was none.
assert_file_exists "Z1" "$GOLANGCI"

# Z2 — the enabled set is CHOSEN, not inherited: `default: none` plus exactly
# the nine linters SPEC-082 locked, each with its own argument in the file.
# This is the assertion that fires if someone "just turns on a few more":
# `linters.default: all` reports 15,300 issues on this tree, and a gate that
# size gets silenced with //nolint, which is worse than no gate.
if [ ! -f "$GOLANGCI" ]; then
    fail "Z2" "$GOLANGCI does not exist"
elif ! grep -Eq '^  default: none$' "$GOLANGCI"; then
    fail "Z2" "$GOLANGCI must declare '  default: none' — nothing inherited"
else
    z2_want="depguard errcheck errorlint ineffassign nolintlint rowserrcheck sqlclosecheck staticcheck unused"
    z2_got=$(awk '/^  enable:$/{f=1; next} f && /^  [a-z]/{f=0} f' "$GOLANGCI" \
        | grep -oE '^    - [a-z]+' | awk '{print $2}' | sort | tr '\n' ' ' | sed 's/ $//')
    if [ "$z2_got" = "$z2_want" ]; then
        ok "Z2"
    else
        fail "Z2" "enabled linters are [$z2_got]; expected [$z2_want]"
    fi
fi

# Z3 — depguard denies BOTH halves of the no-sql-in-cli-layer constraint, and
# is scoped to the package its path glob names. The driver-only half is the
# whole reason depguard was chosen over porting internal/mcpserver's
# TestNoSQLImport: that test greps the literal "database/sql" and nothing
# else, so an import of modernc.org/sqlite alone walks straight past it.
if [ ! -f "$GOLANGCI" ]; then
    fail "Z3" "$GOLANGCI does not exist"
else
    z3_missing=""
    for z3_needle in "no-sql-in-cli-layer:" "internal/cli/*.go" "!\$test" \
                     'pkg: "database/sql"' 'pkg: "modernc.org/sqlite"'; do
        if ! grep -F -q -- "$z3_needle" "$GOLANGCI"; then
            z3_missing="$z3_missing [$z3_needle]"
        fi
    done
    if [ -z "$z3_missing" ]; then
        ok "Z3"
    else
        fail "Z3" "$GOLANGCI depguard rule is missing:$z3_missing"
    fi
fi

# Z4 — both gates actually run in CI. A config nobody runs is documentation.
if [ ! -f "$CI_WORKFLOW" ]; then
    fail "Z4" "$CI_WORKFLOW does not exist"
else
    z4_missing=""
    grep -F -q -- "golangci/golangci-lint-action@" "$CI_WORKFLOW" || z4_missing="$z4_missing [golangci-lint-action]"
    grep -F -q -- "./scripts/coverage.sh" "$CI_WORKFLOW" || z4_missing="$z4_missing [scripts/coverage.sh]"
    if [ -z "$z4_missing" ]; then
        ok "Z4"
    else
        fail "Z4" "$CI_WORKFLOW does not run:$z4_missing"
    fi
fi

# Z5 — THE FLOOR GUARD, same idiom as X3: derive, then diff. The floor is
# written down once, in scripts/coverage.sh. This reads it back out and
# asserts the practices page states the same number, so the page cannot claim
# a floor CI is not enforcing (or the reverse).
if [ ! -f "$COVERAGE_SH" ]; then
    fail "Z5" "$COVERAGE_SH does not exist"
elif [ ! -f "$PRACTICES_DOC" ]; then
    fail "Z5" "$PRACTICES_DOC does not exist"
else
    z5_floor=$(grep -E '^FLOOR=' "$COVERAGE_SH" | head -1 | cut -d= -f2)
    if [ -z "$z5_floor" ]; then
        fail "Z5" "$COVERAGE_SH has no '^FLOOR=' line"
    elif grep -F -q -- "${z5_floor}%" "$PRACTICES_DOC"; then
        ok "Z5"
    else
        fail "Z5" "$PRACTICES_DOC does not state the enforced floor ${z5_floor}% from $COVERAGE_SH"
    fi
fi

# Z6 — the remedies named in the config and on the page actually exist as
# commands. Same idiom as X8: `just inventory` had to be real for X3's failure
# message to mean anything.
if [ ! -f justfile ]; then
    fail "Z6" "justfile does not exist"
else
    z6_missing=""
    grep -E -q '^lint:' justfile || z6_missing="$z6_missing [lint]"
    grep -E -q '^coverage:' justfile || z6_missing="$z6_missing [coverage]"
    if [ -z "$z6_missing" ]; then
        ok "Z6"
    else
        fail "Z6" "justfile is missing recipe(s):$z6_missing"
    fi
fi

# Z7 — the two decisions/ rows in the inventory must ADD UP to the files on
# disk. Deferred at SPEC-080, routed here by STAGE-022's Design Notes.
# scripts/inventory.sh counts `insight.type: decision` and
# `insight.type: reservation` separately; a DEC-*.md carrying a third type, or
# none at all, is counted by NEITHER row and vanishes from the page silently —
# while X3 still round-trips green, because the page is generated from the same
# two filters that lost it.
#
# THE MEANING, decided at SPEC-082 LD10: a hard fail. The two readings of an
# untyped file — a typo, or a category inventory.sh has not been taught yet —
# differ in intent, not in consequence: either way the file is invisible and
# the page under-reports. This harness has no warning tier, so the failure
# message names both remedies instead of inventing one.
#
# decisions/_template.md is deliberately outside every count (it does not match
# DEC-*.md) and stays outside this one.
z7_files=$(ls decisions/DEC-*.md 2>/dev/null | wc -l | tr -d ' ')
z7_decisions=$(grep -l '^  type: decision' decisions/DEC-*.md 2>/dev/null | wc -l | tr -d ' ')
z7_reserved=$(grep -l '^  type: reservation' decisions/DEC-*.md 2>/dev/null | wc -l | tr -d ' ')
z7_sum=$((z7_decisions + z7_reserved))
if [ "$z7_sum" -eq "$z7_files" ]; then
    ok "Z7"
else
    fail "Z7" "the inventory covers $z7_sum of $z7_files decisions/DEC-*.md files ($z7_decisions decision + $z7_reserved reservation). A DEC-*.md with a missing or unknown 'insight.type' is counted by neither row: fix its front-matter, or teach scripts/inventory.sh a row for the new type."
fi
```

### 7. `docs/engineering-practices.md` — three prose hunks

Apply by exact match. **Each old block occurs exactly once** — assert that
before replacing; do not anchor on `^## ` headings, which repeat.

**Hunk A** — in `## What this does not measure`. Old:

```md
- **No lint gate and no coverage number.** CI
  ([`../.github/workflows/ci.yml`](../.github/workflows/ci.yml)) runs `gofmt -l
  .`, `go vet ./...` and `go test ./...`. There is no `golangci-lint`, no
  `-cover`, and no badge.
```

New:

```md
- **No coverage badge, and no coverage service.** The number is not published
  to Codecov or Coveralls, and there is no badge, because a badge is a *cached*
  number served by a third party and the first rule on this page is that a
  current number is derived. What exists instead: `just coverage` prints it and
  CI fails below the floor. The trade is real — a reader has to run a command
  where a badge would have shown a number.
```

**Hunk B** — in `## What the tests actually pin`, appended after the
`Documentation has its own test suite` bullet. Keep that bullet as-is and add
these two immediately after it:

```md
- **A blocking constraint has a machine behind it.**
  [`../.golangci.yml`](../.golangci.yml) enables nine linters, each argued in
  the file itself; run with `linters.default: all`, golangci-lint reports
  15,300 issues on this tree, which is why the list is short and why every
  rejected linter's count is written down. One of the nine is `depguard`, wired
  to `no-sql-in-cli-layer`: adding `_ "database/sql"` to
  `internal/cli/root.go` passed `go build`, `go vet`, `gofmt`, `just test` and
  `just test-docs`, and now fails the `lint` job. It guards production files
  only — the test-file half of that constraint is open, and the config says so.
- **Coverage carries a floor, not a target.**
  [`../scripts/coverage.sh`](../scripts/coverage.sh) — `just coverage` locally,
  the `coverage` job in CI — fails below **80.0%**. The floor sits below the
  measurement (83.5% of statements on 2026-08-21) on purpose, so that ordinary
  refactoring never turns CI red and nobody is ever rewarded for a test written
  to move a percentage. The lowest package is `cmd/brag` at 23.7%: one
  function, `main`, 29 statements out of ~3,900. It is left there.
```

**Hunk D** — the closing `STAGE-022` paragraph. Old (the lines after the
`[`STAGE-022`](…)` link line, which is unchanged):

```md
closes the lint gate and the coverage number, and one of the known defects — the
`Entries:` envelope inconsistency. It does not close the rest: its *Explicitly
out of scope* section defers benchmarks to
[`PROJ-009`](../projects/PROJ-009-scale-baseline-and-harness/brief.md), running
the documentation assertions in CI is owned by nothing today, and the no-network
claim stays enforced by review.
```

New:

```md
closed the lint gate and the coverage floor, both described above. It does not
close the rest: its *Explicitly out of scope* section defers benchmarks to
[`PROJ-009`](../projects/PROJ-009-scale-baseline-and-harness/brief.md), the
`Entries:` envelope inconsistency is its remaining correctness item, running
the documentation assertions in CI is owned by nothing today, and the
no-network claim stays enforced by review.
```

**Then regenerate the inventory block.** Run `just inventory` and paste its
output between the `<!-- inventory:begin … -->` and `<!-- inventory:end -->`
markers. **The derivation outranks the cache**: whatever `just inventory` prints
is correct, even where it disagrees with this spec. Expected: three rows change
— `Documentation assertions (distinct ids)` 177 → 184,
`Questions tracked in guidance/questions.yaml` 18 → 19, and
`…of those, still open` 6 → 7 (the last two because §9 above adds a question —
do that step before this one, as the *Order of work* has it).

Resulting size, measured at design with all three hunks and the re-pasted
block: **301 lines / 2,468 words** — which is why `X7` moves to words (LD9).

### 8. `AGENTS.md` §3 — one line

Old:

```md
- **Linter / Formatter:** `gofmt` (enforced) + `go vet`. `golangci-lint` welcome but not required in CI yet.
```

New:

```md
- **Linter / Formatter:** `gofmt` (enforced) + `go vet` + `golangci-lint` (v2.13.1, pinned in `.github/workflows/ci.yml`; nine linters, each argued in `.golangci.yml`). All three gate CI. Run the lint gate locally with `just lint`.
```

### 9. `guidance/questions.yaml` — append one entry

```yaml
  - id: no-sql-in-cli-layer-test-scope
    question: "Does the no-sql-in-cli-layer constraint bind test files, or only production code?"
    priority: medium
    status: open
    raised_by: claude-opus-5
    raised_at: 2026-08-21
    assigned_to: null
    notes: |
      Filed at SPEC-082 design (LD7). The constraint's RULE TEXT is unqualified
      — "Files under internal/cli/ must not import database/sql or any SQL
      driver" — so on its literal reading FIVE of the 30 *_test.go files under
      internal/cli/ violate a BLOCKING constraint today, across EIGHT import
      lines. Measured by depguard (the import graph, not grep) 2026-08-21:
      coverage_test.go, impact_test.go, project_test.go, wrapped_test.go import
      database/sql; coverage_test.go, impact_test.go, story_test.go,
      wrapped_test.go import modernc.org/sqlite. Note internal/cli/list_test.go
      is NOT among them — a whole-file grep hits it, but line 275 is a comment.

      Three things point the same way about intent: the constraint's own
      rationale is about a future TUI/API frontend reusing internal/cli, which
      links production code and not _test.go files; internal/cli/root.go's
      package comment (scoped at SPEC-080's re-verify) already says "Its
      PRODUCTION CODE imports no SQL driver and no database/sql"; and
      internal/storage/storagetest exists solely so CLI tests can reach raw SQL
      without importing it.

      TWO CANDIDATE RESOLUTIONS, both costed at SPEC-082 design:
      (A) Amend the rule text to say "production code", making the mechanism
          and the rule agree. Cheapest, and matches all three signals above.
          Requires deciding it deliberately — SPEC-082 deliberately did NOT do
          this, because a spec must not amend the rule it is mechanising.
      (B) Rework the five test files through internal/storage/storagetest and
          enforce the rule as literally written. story_test.go needs only a dead
          blank import deleted (it imports the driver and never calls sql.Open).
          The other four backdate rows via sql.Open + UPDATE; storagetest.Backdate
          already covers entries.created_at but not updated_at, and
          project_test.go backdates the projects table, so this needs two new
          helpers. Then depguard drops `!$test` and the constraint is true as
          written.

      WHAT SPEC-082 DID INSTEAD: wired depguard on internal/cli production files
      only — closing the gap SPEC-080 demonstrated empirically (adding
      `_ "database/sql"` to internal/cli/root.go passed go build, go vet, gofmt,
      just test AND just test-docs) — and made the open half visible in three
      places: the `!$test` line's comment in .golangci.yml, the practices-page
      sentence that claims the gate, and this entry. guidance/constraints.yaml
      was NOT touched.

      ALSO OPEN, same shape, deliberately not widened here: internal/mcpserver
      has TestNoSQLImport but it greps the literal "database/sql" only, so it is
      driver-blind. Extending depguard there would widen a blocking constraint's
      path glob (internal/cli/**), which is the same class of side effect as
      narrowing it.

      RESOLVE WHEN: the next spec that opens internal/cli/list_test.go (whose
      line-275 comment is wrong under BOTH readings and is on STAGE-021's
      routed stale-comment list), or the next spec that needs a sixth CLI test
      to backdate a row — whichever comes first. Either one has to answer this
      to write a correct comment or a correct helper.
```

### 10. The lint punch list — 19 edits

Each is an exact one-block replacement; each old block occurs **exactly once**
in its file (asserted at design). Applied at design; `golangci-lint run` then
reported **0 issues**, and `gofmt -l .`, `go vet ./...` and `go test ./...` were
all clean.

**errcheck (5).**

`internal/cli/mcp_test.go`:
```go
	cw.Close()     // client closes stdin while the request is in flight
```
→
```go
	_ = cw.Close() // client closes stdin while the request is in flight
```

`internal/editor/launch.go`:
```go
	defer os.Remove(path)
```
→
```go
	defer func() { _ = os.Remove(path) }()
```
(the same shape `internal/cli/atomicwrite.go:33` already uses)

`internal/mcpserver/list_filters_test.go` **and**
`internal/mcpserver/server_test.go` (one occurrence each):
```go
	t.Cleanup(func() { cs.Close() })
```
→
```go
	t.Cleanup(func() { _ = cs.Close() })
```

`internal/mcpserver/transport_test.go`:
```go
	io.Copy(&buf, r)
```
→
```go
	_, _ = io.Copy(&buf, r)
```

**errorlint (7).**

`internal/cli/mcp_test.go`:
```go
	serverClosing := fmt.Errorf("%w: %v", &jsonrpc.Error{Code: -32004, Message: "server is closing"}, io.EOF)
```
→
```go
	serverClosing := fmt.Errorf("%w: %w", &jsonrpc.Error{Code: -32004, Message: "server is closing"}, io.EOF)
```

`internal/editor/editor.go`, import block:
```go
	"bufio"
	"bytes"
	"fmt"
```
→
```go
	"bufio"
	"bytes"
	"errors"
	"fmt"
```
and:
```go
	if err != nil && err != io.EOF {
```
→
```go
	if err != nil && !errors.Is(err, io.EOF) {
```

`internal/mcpserver/memory.go`:
```go
				return nil, nil, fmt.Errorf("brag_memory: %v", qe)
```
→
```go
				return nil, nil, fmt.Errorf("brag_memory: %w", qe)
```

`internal/storage/fts_test.go`, import block:
```go
	"context"
	"database/sql"
	"io/fs"
```
→
```go
	"context"
	"database/sql"
	"errors"
	"io/fs"
```
and three comparisons:
```go
	if err == sql.ErrNoRows {
		t.Fatalf("entries_fts has no row for id=%d", inserted.ID)
```
→
```go
	if errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("entries_fts has no row for id=%d", inserted.ID)
```
```go
	if err != sql.ErrNoRows {
		t.Fatalf("MATCH %s returned rowid=%d err=%v, want ErrNoRows", oldExpr, oldRow, err)
```
→
```go
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("MATCH %s returned rowid=%d err=%v, want ErrNoRows", oldExpr, oldRow, err)
```
```go
	if err != sql.ErrNoRows {
		t.Fatalf("MATCH xxx_missing_tag: rowid=%d err=%v, want ErrNoRows", missing, err)
```
→
```go
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("MATCH xxx_missing_tag: rowid=%d err=%v, want ErrNoRows", missing, err)
```

`internal/storage/migrate_test.go`, import block:
```go
	"context"
	"database/sql"
	"path/filepath"
```
→
```go
	"context"
	"database/sql"
	"errors"
	"path/filepath"
```
and:
```go
	if err != sql.ErrNoRows {
		t.Errorf("expected no bad table; got name=%q err=%v", got, err)
```
→
```go
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected no bad table; got name=%q err=%v", got, err)
```

**rowserrcheck (3) + sqlclosecheck (1)**, all in
`internal/storage/store_test.go`.

Site 1 — both linters at once:
```go
	rows, err := db.Query("SELECT name FROM tags ORDER BY name")
	if err != nil {
		t.Fatalf("query tags: %v", err)
	}
	var tagNames []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatalf("scan tag: %v", err)
		}
		tagNames = append(tagNames, n)
	}
	rows.Close()
```
→
```go
	rows, err := db.Query("SELECT name FROM tags ORDER BY name")
	if err != nil {
		t.Fatalf("query tags: %v", err)
	}
	defer rows.Close()
	var tagNames []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatalf("scan tag: %v", err)
		}
		tagNames = append(tagNames, n)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate tags: %v", err)
	}
```

Site 2:
```go
		pairs = append(pairs, p)
	}
```
→
```go
		pairs = append(pairs, p)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate taggings: %v", err)
	}
```

Site 3:
```go
		names = append(names, n)
	}
	if len(names) != 2 || names[0] != "b" || names[1] != "c" {
```
→
```go
		names = append(names, n)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate taggings: %v", err)
	}
	if len(names) != 2 || names[0] != "b" || names[1] != "c" {
```

**staticcheck ST1005 (1).** `internal/storage/devguard.go` — drop the trailing
period from the DEC-026 guard's message. No test asserts the tail (checked).
```go
		"throwaway copy, or set %s=1 to override.",
```
→
```go
		"throwaway copy, or set %s=1 to override",
```

**unused (2) + the ldflags that fed them (LD8).**

`cmd/brag/main.go`:
```go
// version is set to "dev" for local builds. goreleaser injects the
// real values via ldflags (-X main.version=... -X main.commit=...
// -X main.date=...) at release-build time. See .goreleaser.yaml.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)
```
→
```go
// version is set to "dev" for local builds. goreleaser injects the real
// value via an ldflag (-X main.version=...) at release-build time. See
// .goreleaser.yaml. main.commit and main.date were removed at SPEC-082:
// goreleaser set them and nothing ever read them.
var version = "dev"
```

`.goreleaser.yaml`:
```yaml
      - -X main.version={{ .Version }}
      - -X main.commit={{ .Commit }}
      - -X main.date={{ .Date }}
```
→
```yaml
      - -X main.version={{ .Version }}
```

### 11. Order of work

1. `scripts/test-docs.sh` — Group Z and the `X7` switch **first**
   (`test-before-implementation`). Run `just test-docs`; `Z1`–`Z6` must fail,
   `Z7` must pass, `X3` must fail on the assertion count.
2. `.golangci.yml`, `scripts/coverage.sh` (`chmod +x`), `justfile`.
3. `golangci-lint run` → confirm **19** issues, matching the punch-list table.
4. Apply the 19 edits. `golangci-lint run` → **0 issues**.
5. `.github/workflows/ci.yml`, `AGENTS.md`, `guidance/questions.yaml`.
6. `docs/engineering-practices.md` — three hunks, then `just inventory` and
   paste.
7. Run every gate in the Acceptance Criteria, plus mutation checks **M-A**
   through **M-E**, reverting each mutant.

---

## Build Completion

*Filled in during the build cycle.*

- **Deviations from the spec:** **one**, forced and mechanical. `Y4` pins the
  question-register counts as literals (`18`/`6`) inside `scripts/test-docs.sh`.
  Output 8 (the new `no-sql-in-cli-layer-test-scope` entry) moves them to
  `19`/`7`, so `Y4` failed once every other artifact was staged. The spec's
  *Failing Tests* section tracked that move through `X3` and the inventory block
  only; `Y4` was missed. Re-pinned `18/6 -> 19/7` with a comment naming SPEC-082
  and stating that the corpus changed deliberately — the same re-pin discipline
  SPEC-079 LD5 used, not a band-widening. Assertion count is unaffected (`Y4` is
  one id either way). Nothing else deviates; every literal in *Notes for the
  Implementer* was transcribed byte-for-byte.
- **New DEC-* files created:** none.
- **Constraints checked:** `no-sql-in-cli-layer` (mechanised, production half;
  `guidance/constraints.yaml` **unchanged** — `git diff` empty, AC-12);
  `errors-wrap-with-context` (errorlint; two first-ever production violations
  fixed); `no-new-top-level-deps-without-decision` (does not apply — `go.mod`
  and `go.sum` diffs both empty); `test-before-implementation` (Group Z and the
  `X7` switch landed first and were observed red — `Z1`–`Z6` failing, `Z7`
  green, `X3` failing — before any artifact was written); `one-spec-per-pr`;
  `storage-tests-use-tempdir` (unaffected).
- **Gates:** `just test` **pass** · `just test-docs` **ALL OK** · `gofmt -l .`
  **empty** · `go vet ./...` **clean** · `just lint` **0 issues** ·
  `golangci-lint config verify` **exit 0** · `just coverage`
  **`total: 83.5%   floor: 80.0%`** · `actionlint .github/workflows/ci.yml`
  **exit 0** · `goreleaser check` **unchanged** — the pre-existing `brews`
  deprecation, byte-identical output before and after this spec's ldflags edit
  (captured both ways), nothing new.
- **Mutation checks M-A…M-E:** all five fired; each mutant confirmed present on
  disk before the failure was credited, and confirmed reverted after.
  **M-A** `_ "database/sql"` in `internal/cli/root.go` → depguard, at
  `root.go:17:2`, the design's exact site. **M-B** `_ "modernc.org/sqlite"` →
  depguard at `root.go:19:2`. **M-C** bare `//nolint` → nolintlint, 3 issues
  (unspecific / unused / unexplained), confirming all three settings have teeth.
  **M-D** `FLOOR=90.0` → exit **1**, `coverage 83.5% is below the floor of
  90.0%`; reverted file is byte-identical to the spec's literal (sha compared).
  Note: `git diff --quiet` is blind to `scripts/coverage.sh` because the file is
  **untracked**, so the first M-D run reported "did not mutate" when the mutant
  was in fact on disk — detection was redone against content hashes. The
  detector was wrong, not the mutant. **M-E** `decisions/DEC-999-mutation-probe.md`
  with `insight.type: bogus` → `Z7` fails with *"the inventory covers 46 of 47"*
  while `scripts/inventory.sh` still prints `Decision records | 45`. Run twice:
  the first probe carried `confidence: 0.50`, which moved the *lowest-confidence*
  row and took `X3` red for an unrelated reason. Re-run at `confidence: 0.90` to
  isolate the variable, `X3` stayed **green** and `Z7` was the **only** failure —
  which is the proposition LD10 actually claims. Probe deleted.
- **`just inventory` value pasted:** **184** documentation assertions, with
  `Questions tracked` **19** and `…of those, still open` **7** — the three rows
  the spec predicted, derived and pasted, not hand-edited. The derivation agreed
  with the spec's arithmetic in all three cases.

### Build-phase reflection (3 questions, short answers)

- **What was unclear in the spec?** Almost nothing — this is the least
  ambiguous spec in the project so far, because §12(b) ran every literal through
  its own tool at design and recorded the output, so build could *diff against a
  prediction* rather than judge a result. Two numbers in the spec's own prose are
  off and neither mattered: *Outputs* says "Modified files (18)" while its own
  list enumerates **19** (item range "11–18" names eleven files), and §10 is
  titled "19 edits" while it specifies **21** replacement blocks for 19 lint
  findings (three `errors` imports are separate blocks). Both are counts of the
  spec's own contents — the same hypothesis-not-measurement failure the stage is
  organised around, arriving from the direction nobody guards.
- **What was missing that you had to decide yourself?** The `Y4` re-pin (see
  *Deviations*). It is the second-order form of the thing this spec is about: the
  spec correctly reasoned that adding a question moves the two inventory rows, and
  correctly routed that through `X3` and the pasted block — but a *second* guard
  also pinned those same two numbers as literals, and nothing enumerated the
  pinners. `X3` and `Y4` both cache the same derived value in different places.
- **What would you do differently?** Extract literals mechanically rather than
  retyping them. Every artifact here was cut from the spec's own fenced blocks by
  line range and applied with a uniqueness assertion, and that caught a real
  error: the `AGENTS.md` hunk was first anchored on a line range that resolved to
  the closing ``` fence, which occurs **7** times in `AGENTS.md` — a
  silent 7-way corruption if it had been applied. That is the same truncation
  failure this project has already suffered once. The assertion, not the care,
  is what caught it. Corroborating evidence that the transcription is exact:
  `docs/engineering-practices.md` landed at **301 lines / 2,468 words**, matching
  the design's measured prediction to the word.

---

## Reflection (Ship)

*Filled in during the ship cycle.*

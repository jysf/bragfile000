---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-082
  type: chore                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
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

references:
  decisions: []
  constraints:
    - no-sql-in-cli-layer          # the depguard candidate would mechanise this; the scope question is open
    - no-new-top-level-deps-without-decision  # golangci-lint is a CI tool, not a module dep — confirm at design
  related_specs:
    - SPEC-073                     # the coverage sentence that was wrong four times — this stage's cautionary case
    - SPEC-079                     # the practices page a coverage number gets presented against
    - SPEC-080                     # routed the no-sql-in-cli-layer gap and the stale comments here
---

# SPEC-082: lint and coverage, gating CI

> **Cycle: frame.** A **go/no-go**, not a design. It records the measurements,
> states what is genuinely undecided, and hands design the forks. Per AGENTS.md
> §6, design happens in a **fresh session**.

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

## Decisions for the design session

1. **Which linters, and why each one?** The stage's Design Notes are explicit
   that a default enable-all produces a wall of findings that gets silenced with
   `//nolint`, which is worse than no lint. **Enable a small set and justify each
   individually.** Note the config itself will be read by the same audience the
   practices page is written for. Run the candidate set against the repo *at
   design* and report what it actually finds — a linter that fires 200 times on
   existing code is a decision about remediation, not a config line.
2. **What is the coverage floor, and what is it protecting against?** It must be
   a floor the current number clears, set from measurement. State whether it is
   module-wide (83.5%) or per-package, and if per-package, what happens to
   `cmd/brag` at 23.7%. **A floor that would fail today is a target in
   disguise.**
3. **Does the badge come from a service or from the repo?** A README badge that
   depends on an external service (Codecov, Coveralls) means a network
   dependency and an account — weigh that against this repo's local-first,
   no-network posture, which is a stated identity and not just a preference.
   A self-generated badge or a plain reported number are both live options.
4. **The `no-sql-in-cli-layer` mechanism — and the scope question underneath
   it.** STAGE-022's Design Notes carry the full evidence from SPEC-080:
   `depguard` scoped to `internal/cli/**` is a natural fit alongside this
   spec's lint work, but **the constraint's text is unqualified by
   production-vs-test, so on its literal reading five test files violate a
   blocking constraint today.** Decide whether this spec takes the mechanism on
   at all, or whether it belongs in its own spec — and if it does take it on,
   **the scope question is a decision, not a config choice**. Do not amend
   `guidance/constraints.yaml` as a side effect of wiring a mechanism that
   assumes an answer.

## Out of scope

- **Benchmarks and any scale/perf harness** — PROJ-009, with the crustyimg
  methodology harvest it depends on.
- **Raising coverage by writing tests to hit a number.** Measure what is there.
- **The `Entries:` envelope inconsistency** — STAGE-022's second spec.
- The five stale-comment items routed here from STAGE-021 (`store.go`'s `Store`
  type comment, `list_test.go:275`, V1's noun, A1's "smallest deep-dive doc",
  A11's stale A1 reference). They are one-word fixes best folded into whichever
  spec next opens those files; naming them here so they are not lost.

## Go / no-go

**GO.** Complexity **M**, not S: the lint config is a set of individually
justified choices measured against real findings, the floor is a judgement that
has to survive the practices page, and fork 4 carries an unresolved scope
question about a blocking constraint.

**Why now:** it is the stage's first spec, and STAGE-021 established what a
coverage number is allowed to claim — which was the stated reason for doing the
docs stage first.

**What would make this a no-go:** if the coverage work were framed as *raising*
the number. It is not. The deliverable is a measured number, an honest floor, and
a claim that survives a mutation check — the same bar SPEC-073's coverage
sentence failed four times.

---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-052                        # stable, never reused
  type: decision                     # decision | analysis | recommendation | observation | reservation
  confidence: 0.9                    # small, well-precedented contract tightening —
                                     # mirrors the existing `add` stdout-ID
                                     # contract verbatim, fully covered by
                                     # paired tests.
  audience:
    - developer
    - agent

agent:
  id: claude-sonnet-5
  session_id: null

project:
  id: PROJ-008
repo:
  id: bragfile

created_at: 2026-09-07
supersedes: null
superseded_by: null

tags:
  - cli
  - stdout-contract
  - edit
---

# DEC-052: `brag edit` emits the mutated entry ID on stdout as the applied/no-op signal

## Decision

`brag edit <id>` prints the mutated entry's ID to stdout when (and only
when) it writes a change, and prints nothing to stdout on a no-op. The
existing human-facing `Updated.` / `No changes.` lines stay on stderr,
unchanged. This extends the `stdout-is-for-data-stderr-is-for-humans`
contract that `brag add` already implements (`add.go:179`) to `edit`.

## Context

`runEdit` wrote `Updated.` (on write) and `No changes.` (on no-op) both to
stderr and returned `nil` either way. That stderr choice is correct — both
are human messages, and the `stdout-is-for-data-stderr-is-for-humans`
constraint pins them there. But it left a real gap: **both outcomes
produced empty stdout and exit 0**, so a batch script had no way to
distinguish "applied" from "no-op" on the data channel or the exit code.

Reported from a field audit: a 13-entry batch script read empty stdout on
every call, concluded `0 updated, 13 failed` even though all 13 had
applied, then "retried" one of them and double-applied it. The fix
respects the constraint rather than relaxing it — the entry ID is data, it
belongs on stdout, and its presence *is* the applied signal.

## Alternatives Considered

- **Option A: move `Updated.`/`No changes.` to stdout.** What it is: print
  the human prose to stdout instead of stderr so a script has something to
  grep. Why rejected: violates `stdout-is-for-data-stderr-is-for-humans`
  directly — those lines are prose, not data, and moving them does not give
  a script a machine-checkable signal, only a string to parse.

- **Option B: print a structured JSON result on every invocation.** What
  it is: emit e.g. `{"id":42,"updated":true}` to stdout always. Why
  rejected: `edit` has no `--json`/`--format` flag and no other output
  mode; introducing one is new surface area beyond what this bug fix
  needs, and out of scope for a two-defect capture-integrity spec.

- **Option C (chosen): print the ID to stdout on write, nothing on no-op.**
  What it is: bind `s.Update`'s return value (previously discarded with
  `_`) and `fmt.Fprintln(cmd.OutOrStdout(), inserted.ID)` before the
  existing stderr `Updated.` line; leave the `!changed` branch untouched.
  Why selected: mirrors `add`'s existing stdout-ID contract exactly, so the
  applied/no-op distinction becomes observable on stdout alone with the
  smallest possible surface change — no new flags, no new output shape.

## Consequences

- **Positive:** A batch script driving `brag edit` can now tell "applied"
  from "no-op" from stdout alone, closing the exact gap that produced a
  double-apply in the field report. Consistent with `add`'s existing
  contract, so the two mutating commands behave the same way.
- **Negative:** Any script or human that previously assumed `edit`'s stdout
  is always empty must account for the new ID line on a successful write.
  This is the intended behavior change.
- **Neutral:** No change to `runEdit`'s signature, exit codes, or the
  no-op path. No database or migration impact.

## Validation

Covered by `TestEditCmd_PrintsIDToStdoutOnUpdate` (the load-bearing case)
and the boundary guard `TestEditCmd_NoChangesEmitsNoStdout` (proves the new
write does not leak onto the no-op path). `TestEditCmd_HappyPath`'s
pre-existing assertion that stdout stays empty on a write was invalidated
by this decision and updated to assert the ID instead. Revisit if `edit`
ever grows a `--json`/`--format` flag — the plain-ID stdout contract would
need to be reconciled with a structured output mode at that point.

## References

- Related specs: SPEC-089
- Related decisions: DEC-007 (`ErrUser` sentinel and CLI error/output
  conventions), `stdout-is-for-data-stderr-is-for-humans` constraint
  (`guidance/constraints.yaml`)
- External docs: `internal/cli/add.go:179` (the `add` stdout-ID contract
  this mirrors)

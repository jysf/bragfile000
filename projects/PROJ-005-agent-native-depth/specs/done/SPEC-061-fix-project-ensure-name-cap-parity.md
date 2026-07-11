---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-061
  type: story                      # epic | story | task | bug | chore
  cycle: ship
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-005
  stage: STAGE-016
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7     # usually same Claude, different session
  created_at: 2026-07-10

references:
  decisions: [DEC-036]
  constraints: [stdout-is-for-data-stderr-is-for-humans, errors-wrap-with-context, no-sql-in-cli-layer]
  related_specs: [SPEC-057]
---

# SPEC-061: fix project ensure name cap parity

## Context

A pre-release adversarial sweep found that `brag project ensure` caps the
project name length in **runes**, but the two capture paths that actually
write brag entries — `brag add --json` and the MCP `brag_add` tool — cap in
**bytes**. The comment above the `ensure` check claims the cap "matches" the
capture cap so an ensured name can always soft-match a normally-added entry
(DEC-036), but that invariant is FALSE for multibyte names: a 40-CJK-character
name is 40 runes (ensure accepts) but 120 bytes (capture paths reject). So
`ensure` can register a project name that the agent-native capture surfaces can
never write an entry for — the exact failure DEC-036's parity clause exists to
prevent. This lives in `STAGE-016` (polish) under `PROJ-005`.

## Goal

Align `brag project ensure`'s project-name length cap to count **bytes** (not
runes), matching `add --json` and MCP `brag_add`, so any name ensure accepts is
also acceptable to the capture paths (DEC-036 parity).

## Inputs

- **Files to read:** `internal/cli/project.go` (ensure cap ~L157–161),
  `internal/cli/add_json.go` (`len(in.Project) > 64`),
  `internal/mcpserver/server.go` (`len(in.Project) > 64`) — the byte-cap
  the fix aligns to.
- **Related code paths:** `internal/cli/project_test.go`

## Outputs

- **Files modified:**
  - `internal/cli/project.go` — change `len([]rune(name)) > 64` to
    `len(name) > 64` in `runProjectEnsure`; update the adjacent comment to state
    a 64-BYTE cap matching the capture paths (keep the DEC-036 rationale).
  - `internal/cli/project_test.go` — add fail-first coverage.
- **Database changes:** none.

## Acceptance Criteria

- [x] A name ≤64 runes but >64 bytes (e.g. 40× a 3-byte CJK rune = 120 bytes)
      passed to `ensure` is REJECTED with a `UserError` (exit 1, empty stdout,
      message on stderr) — matching `add --json`.
- [x] A 64-byte ASCII name is accepted; a 65-byte one is rejected.
- [x] The same over-cap name fed through `add --json` is also rejected
      (symmetry proven in-test).
- [x] The rejected name never appears in `project list`.

## Failing Tests

- **`internal/cli/project_test.go`**
  - `"TestProjectEnsure_MultibyteOverByteCapErrUser"` — asserts ensure rejects
    a 40-CJK-rune (120-byte) name with `ErrUser`, empty stdout, no list row;
    proves `add --json` rejects the same name.
  - `"TestProjectEnsure_ByteCapAcceptsExactly64ASCII"` — asserts a 64-byte
    ASCII name is accepted.

## Implementation Context

### Decisions that apply

- `DEC-036` — project ensure is an idempotent upsert whose name cap must match
  the capture paths so an ensured name can always soft-match a normally-added
  entry; this fix restores that parity for multibyte names.

### Constraints that apply

- `stdout-is-for-data-stderr-is-for-humans` — rejection is a `UserError`;
  stdout stays empty, the human message goes to stderr.
- `errors-wrap-with-context` — unchanged; the cap returns a `UserErrorf`.
- `no-sql-in-cli-layer` — unchanged; the cap is a pure length check.

### Prior related work

- `SPEC-057` (shipped) — introduced `brag project ensure` and the rune-based
  cap this fix corrects.

### Out of scope (for this spec specifically)

- The capture paths (`add_json.go`, `server.go`) are correct and unchanged.
- `brag project new`'s no-cap behavior is a separate pre-existing asymmetry,
  not addressed here.

## Notes for the Implementer

The fix is a one-token change (`len([]rune(name))` → `len(name)`) plus a comment
update. Write the failing test first; the current rune-based code accepts the
multibyte name, so the test fails before the fix and passes after.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `fix/spec-061-ensure-cap-parity` (stacked on `fix/spec-060-spark-until-bound`)
- **PR (if applicable):** base `main`
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - none — this restores the parity DEC-036 already mandates.
- **Deviations from spec:**
  - none
- **Follow-up work identified:**
  - `brag project new` has no name cap at all (pre-existing asymmetry) — a
    candidate for a future small-wins spec, deliberately out of scope here.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing. The bug was fully characterized (rune-vs-byte, exact lines, repro).

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. DEC-036 is the load-bearing decision and it was cited.

3. **If you did this task again, what would you do differently?**
   — Nothing material; one-line fix with a fail-first multibyte test is the
     right shape. Worth noting the same byte/rune question exists for the other
     capped fields (title, tags, type) as a future consistency sweep.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — Nothing material — a one-token rune→byte change. Outcome: `brag project
   ensure` now rejects exactly the multibyte names the capture paths reject,
   so any ensured name is always writable (DEC-036 parity restored).

2. **Does any template, constraint, or decision need updating?**
   — No. DEC-036 already mandates the ensure↔capture parity this fix restores;
   no template or constraint change.

3. **Is there a follow-up spec I should write now before I forget?**
   — The remaining asymmetry is `brag project new`, which still has no name cap
   at all — a candidate for a future small-wins spec. (The title/tags/type
   byte-cap parity was subsequently closed by SPEC-064's shared validator.)

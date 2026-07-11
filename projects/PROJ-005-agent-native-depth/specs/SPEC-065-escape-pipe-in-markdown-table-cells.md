---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-065
  type: bug                        # epic | story | task | bug | chore
  cycle: verify
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-005
  stage: STAGE-016
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-10

references:
  decisions: [DEC-013]
  constraints: [stdout-is-for-data-stderr-is-for-humans]
  related_specs: [SPEC-015, SPEC-064]
---

# SPEC-065: escape pipe in markdown table cells

## Context

Pre-release audit MEDIUM #6. `internal/export/markdown.go`
`RenderEntry` (used by `brag show` and `brag export --format markdown`)
writes field values raw into `| field | value |` GFM table cells. A `|`
in a value splits the cell, producing extra columns and garbling the
rendered table. Repro: `brag add -i 'cut latency | doubled QPS'` then
`brag show 1` renders `| impact | cut latency | doubled QPS |` — four
columns instead of two. The same breaks for a `|` in tags / project /
type. Sits under `STAGE-016` (polish) in `PROJ-005`.

## Goal

Escape `|` as `\|` (the standard GFM table-cell escape) in every field
value rendered into a markdown table cell, so a `|` in a value keeps
the row at exactly two columns. Do not escape `|` in non-table contexts
(headings, description body prose), where it is literal and harmless.

## Inputs

- **Files to read:** `internal/export/markdown.go` — the `RenderEntry`
  table-cell rendering; `internal/export/markdown_test.go` — existing
  golden/render tests.
- **Related code paths:** `internal/cli/show.go`,
  `internal/cli/export.go` (both call `RenderEntry`).

## Outputs

- **Files modified:**
  - `internal/export/markdown.go` — wrap tags/project/type/impact cell
    values in a new `escapeTableCell` helper.
  - `internal/export/markdown_test.go` — new fail-first test.

## Acceptance Criteria

- [x] A `|` in tags / project / type / impact renders as `\|`, keeping
      the table row at exactly two columns.
- [x] A `|` in the description body block stays literal (not escaped).
- [x] Existing DEC-013 golden tests stay green (their fixtures contain
      no `|`, so their bytes are unchanged).

## Failing Tests

- **`internal/export/markdown_test.go`**
  - `TestRenderEntry_EscapesPipeInTableCells` — renders an entry whose
    tags/project/type/impact each contain a `|`; asserts each value
    cell carries the escaped `\|`, every table row has exactly 3
    unescaped pipes (two columns), and the description body `|` stays
    literal.

## Implementation Context

*Read this section (and the files it points to) before starting
the build cycle. It is the equivalent of a handoff document, folded
into the spec since there is no separate receiving agent.*

### Decisions that apply

- `DEC-NNN` — <one-line summary of why this matters here>
- `DEC-MMM` — <one-line summary>

### Constraints that apply

These constraints apply to the paths touched by this task (see
`/guidance/constraints.yaml` for full text):

- `constraint-id-1` — <one-line summary>
- `constraint-id-2` — <one-line summary>

### Prior related work

- `SPEC-YYY` (shipped) — <one-line summary, if relevant>
- `PR #NNN` — <link, if relevant>

### Out of scope (for this spec specifically)

Explicit list of what this spec does NOT include. If any of these feel
necessary during build, create a new spec rather than expanding this one.

- ...

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `fix/spec-065-markdown-cell-escape` (stacked on
  `fix/spec-064-capture-input-hardening`)
- **PR (if applicable):** see PR opened against `main`
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - none — this is a bug fix within DEC-013's fixed shape; escaping a
    cell value does not change the documented shape.
- **Deviations from spec:**
  - none
- **Follow-up work identified:**
  - none. Summary-block bullets (`writeSummaryByType` /
    `writeSummaryByProject`) render type/project as list items, not
    table cells, so a `|` there is harmless — deliberately left out of
    scope.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing. The one judgment call — whether description needs
   escaping — resolved on reading the code: description renders as a
   body block below the table, not a cell, so its `|` is harmless.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. DEC-013 (shape) and SPEC-064 (control-char rejection at
   ingress, which guarantees `|` is the only cell-breaking char that
   can reach the renderer) were the relevant references and both fit.

3. **If you did this task again, what would you do differently?**
   — Nothing material. A single-char escape via one shared helper was
   the minimal correct fix.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>

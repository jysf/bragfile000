---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-080
  type: chore                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-007
  stage: STAGE-021
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-19

references:
  decisions: []
  constraints: []
  related_specs:
    - SPEC-079                     # the practices entry point this completes the stage alongside
    - SPEC-070                     # deferred; holds the unwritten DEC-041 reservation
---

# SPEC-080: godoc pass and the two legibility repairs

> **Cycle: frame.** This is a **go/no-go**, not a design. It records what was
> measured, resizes the work against that measurement, and hands design the two
> real decisions. Per AGENTS.md §6, design happens in a **fresh session**.

## Context

STAGE-021's second and final spec. SPEC-079 built the entry point; this one
closes the three remaining items in the stage's scope — a godoc pass, the
`DEC-041` gap, and open-questions hygiene.

Parent: `STAGE-021-make-the-discipline-legible`, spec 2 of 2.
Project: `PROJ-007`.

## Goal

`go doc` on this codebase reads as intentional, `decisions/` has no unexplained
numbering gap, and `guidance/questions.yaml` describes live uncertainty only.

## The measurement that resizes this spec

PROJ-007's brief lists **"No surfaced godoc / API docs"** among four gaps. That
phrasing implies the doc comments are missing. **They are not.** Measured
2026-08-19:

| | Count |
|---|---|
| Exported declarations across `internal/` + `cmd/` | **175** |
| …lacking a preceding doc comment | **3 (1.7%)** |
| Packages | **15** |
| …lacking a package doc comment | **7** |

And the three undocumented declarations are all conventional interface
satisfactions — `(*ErrQuery).Error`, `(*ErrQuery).Unwrap`
(`internal/memory/pool.go:49-50`), and `(*tagsField).UnmarshalJSON`
(`internal/cli/add_json.go:22`). Go style does not ask for doc comments on
those, so **symbol-level coverage is effectively complete.**

The real gap is **seven missing package doc comments**:

```
cmd/brag   internal/cli   internal/config   internal/export
internal/mcpserver   internal/storage   internal/story
```

Those are, notably, the largest and most load-bearing packages in the tree —
which is how the absence went unnoticed: nobody reads a package comment they
never had.

**This is the same shape SPEC-079 found**: the discipline exists and is not
surfaced. It reduces the godoc item from "document the codebase" to "write
seven package comments," which is why this spec is **S** and not **M**.

## Scope

Three items, all from STAGE-021's In-scope list.

1. **Seven package doc comments.** One per package above, saying what the
   package is *for* and what its boundary is — not restating its name.
2. **The `DEC-041` gap.** `decisions/` jumps 040 → 042. The reservation belongs
   to the deferred SPEC-070 (`brag project goto`) and its unwritten
   "multi-location primary policy" decision. The gap is explained in
   `projects/PROJ-001-mvp/backlog.md`, which a browsing reader never opens.
3. **Open-questions hygiene.** **8 of 18 are open.** Three date from
   **2026-04-19** and have not moved in four months — `shareable-ids`,
   `editor-template-format`, `summary-grouping-heuristics`. At least
   `editor-template-format` is likely answered-in-practice: the editor-launch
   buffer shipped in SPEC-010, so the format was chosen, just never recorded.

## Decisions for the design session

Framing deliberately leaves these open.

1. **What shape does the `DEC-041` marker take?** A tombstone file in
   `decisions/` naming the holder and linking the backlog item keeps the
   reservation and kills the mystery — but it puts a non-decision in a directory
   whose contents are otherwise all decisions, and the practices page now counts
   that directory. Alternatives: release the number and let SPEC-070 take the
   next free id if it is ever built (simplest, but the backlog says do not
   reuse it); or leave it and note the gap on the practices page instead.
   **Whichever is chosen, check what it does to `scripts/inventory.sh`'s
   `Decision records` row** — a tombstone would be counted as a decision unless
   the script excludes it, and `X3` will notice.

2. **Does the hygiene pass derive its own count?** STAGE-021's hygiene line has
   been wrong **three times in three days** — "8 of 18" (wrong total), "9 of 18"
   (wrong both halves, from a grep that matched the file's own header comment),
   and "7 of 17" (correct until the #167/#168 merge landed a new question). That
   is a strong argument for an `inventory.sh` row rather than a fourth restated
   number. Weigh it against the row count the page already carries.

## Out of scope (for this spec specifically)

- **The README restructure** — promoting the MCP call-to-action out of the
  Status blockquote, and switching `test-docs` A1 from `wc -l` to `wc -w`. Agreed
  as a separate, third STAGE-021 spec, sequenced after this one.
- Anything in STAGE-022 (lint, coverage, the `Entries:` envelope).
- Rewriting `godoc` prose that already exists in the 8 packages that have it.
- Answering an open question that is genuinely still open. The pass closes what
  is *already decided*, restates resolve conditions for what is not, and does not
  invent answers to fill the register.

## Go / no-go

**GO.** Complexity **S**, resized down from the brief's framing by the
measurement above: seven package comments, one tombstone decision, and a
register triage.

**Why now:** it is the stage's last backlog item, it blocks nothing, and
STAGE-021 cannot close without it — `just archive-spec` already claimed the
stage was complete once (it counts written specs, not backlog items), which is
exactly the misreading this spec existing prevents.

**What would make this a no-go:** if the godoc item were really "write API
documentation." It is not — that was the brief's assumption, and the measurement
refutes it.

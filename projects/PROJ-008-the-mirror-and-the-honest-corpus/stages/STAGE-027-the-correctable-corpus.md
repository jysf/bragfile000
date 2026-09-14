---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-027
  status: proposed                  # proposed | active | shipped | cancelled | on_hold
  priority: high
  target_complete: null

project:
  id: PROJ-008
repo:
  id: bragfile

created_at: 2026-09-14
shipped_at: null
---

# STAGE-027: the correctable corpus

> **Cycle: frame. GO.** Framed 2026-09-14 from two rounds of `~/ContextCore` field
> feedback plus maintainer direction. Decision recorded in **DEC-053**. Not yet
> ordered against STAGE-024/025/026 — see *Why Now* for the pull-forward argument.

## What This Stage Is

The corpus becomes **correctable in place, non-interactively, over both surfaces
an agent uses.** When all its specs ship, an agent (or a script) can fix a field
of an entry it just wrote — a mis-framed `impact`, the wrong `project`, a bad tag,
a typo — through a non-interactive CLI edit (`brag edit --set field=value…` /
`--json`, a **partial** update) and a `brag_edit` **MCP tool**, without opening an
interactive `$EDITOR`. This is the everyday **fixup** path, and it is what makes
low-friction, unapproved auto-capture safe to rely on. It is deliberately distinct
from STAGE-026's supersede/link, which is for *narratable retractions*, not
fixups (DEC-053).

## Why Now

**The friction is live and it recurred within one week.** A heavy-agent audit hit
it twice (§3.4 "`edit` has no flag mode"; §4.2 "buffer format undocumented"), and
then an agent working in the repo could not figure out how to edit a brag from the
CLI at all, and the maintainer had to ask whether it was even possible over MCP.
Two independent parties, same gap.

It is also **upstream of STAGE-023's bet.** STAGE-023 makes the corpus hold
failures via low-friction capture the maintainer does not pre-approve. Unapproved
capture is only safe if the inevitable mistakes are cheap to correct — otherwise
every fat-fingered auto-brag is permanent or spawns a correction-pair. So this
stage is a **strong pull-forward candidate**; the ordering against
STAGE-024/025/026 is a call for the maintainer at activation, but the argument for
going early is that it de-risks work already in flight rather than adding a new
surface.

Nothing technical blocks it: `Store.Update`, `internal/capture.ValidateChanged`
(DEC-046, already validates only changed fields and strips reserved provenance
tags), and the `internal/mcpserver` tool-registration path all exist. This is
mostly surface, not engine.

## Success Criteria

- A `brag_edit` MCP tool applies a **partial** update to an entry by id (only the
  fields provided change; others are untouched), and an agent can use it to fix an
  entry it wrote in the same session.
- A non-interactive CLI edit exists: `brag edit 42 --set impact="cut p99 40%"`
  (and/or `--json` on stdin) updates without launching `$EDITOR`, printing the id
  on stdout on a real change (consistent with DEC-052) and staying silent on a
  no-op.
- Reserved provenance tags (`agent:`/`model:`/`session:`/`cost:`/`tokens:`) survive
  a tag edit rather than being clobbered (reuse `validateTagsChanged`, DEC-046).
- `delete` is **not** added to the MCP surface (DEC-053).
- `add --help` / `edit --help` (or `docs/for-ai-agents.md`) state which correction
  mechanism to use when: in-place fixup vs. superseding retraction.

## Scope

### In scope
- A non-interactive **CLI** edit: `brag edit --set field=value` (repeatable)
  and/or `brag edit --json` on stdin, doing a **partial** update via `Store.Update`
  + `capture.ValidateChanged`.
- A **`brag_edit` MCP tool**: partial update by id, same engine, provenance-tag
  safe.
- Documenting the two correction mechanisms (fixup vs. retraction) on the surfaces
  a user/agent reads.

### Explicitly out of scope
- **Delete over MCP** (DEC-053 — human-gated at the CLI).
- **Shadow-history / audit trail** on edits (MVP is `updated_at`-bump only,
  DEC-053; add later only if a substantive-edit history proves necessary).
- **Per-agent edit scoping** — MVP is any-entry-by-id, matching CLI `brag edit`
  (DEC-053); tightening to own-entries-only is a later, additive call.
- **Supersede/link** (`brag link corrects|supersedes`) — that is STAGE-026, a
  different mechanism for a different job.
- Write-time `brag add --check` dup warning and `brag lint` (mirror, STAGE-025).

## Spec Backlog

Ordered list of specs composing this stage. IDs are **claimed with a file at
design**, not reserved in prose — `scripts/_lib.sh next_id` counts filenames only,
so a prose-reserved id (as STAGE-024/025/026 and SPEC-088/090 showed) is invisible
to it and collides on the next `new-*`.

- [ ] (not yet written) — non-interactive CLI edit: `brag edit --set`/`--json`, partial update, reusing `capture.ValidateChanged`.
- [ ] (not yet written) — `brag_edit` MCP tool: partial update by id, provenance-tag safe; run it through actual client **registration** at design, not just schema validation (validate ≠ registration — the SPEC-041 lesson).
- [ ] (not yet written) — docs/help: state the fixup-vs-retraction split (`add`/`edit --help`, `docs/for-ai-agents.md`); may fold into SPEC-091.

**Count:** 0 shipped / 0 active / 3 pending

## Design Notes

- **Partial-update semantics.** The whole point is "change one field." `--json`
  should merge (unspecified keys untouched), not replace; `--set` is per-field.
  `capture.ValidateChanged(old, new)` (DEC-046) is the validator — it already
  checks only the fields that differ and grandfathers over-cap fields the caller
  does not touch.
- **Tags on edit.** `validateTagsChanged` (DEC-046) strips reserved-namespace tags
  before applying caps; a tag edit must preserve stamped `agent:`/`model:`/… tags
  rather than treat the stored joined string as user input. Decide set-vs-add
  semantics for `--tags` at design (add/remove flags are a possible refinement).
- **CLI/MCP one engine.** Both surfaces call the same `Store.Update` +
  `ValidateChanged` path (mirrors how `internal/memory.Gather` is shared by CLI and
  MCP, SPEC-074). No second validation path.
- **stdout contract.** The CLI edit keeps DEC-052: id on stdout on a real change,
  human prose on stderr, silent no-op.

## Dependencies

### Depends on
- Nothing technical — `Store.Update`, `internal/capture.ValidateChanged` (DEC-046),
  and the `internal/mcpserver` tool path all ship today.
- **DEC-053** — the posture this stage implements.

### Enables
- Safe unapproved auto-capture (STAGE-023's premise): mistakes become cheap to fix.
- A cleaner division of labor with STAGE-026: fixups here, narratable retractions
  there.

## Stage-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <one-line items>
- **What can a user do now that they couldn't before, at STAGE scope?**
  - <answer | none>
</content>

---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-018
  status: proposed                  # proposed | active | shipped | cancelled | on_hold
  priority: low
  target_complete: null

project:
  id: PROJ-006
repo:
  id: bragfile

created_at: 2026-07-12
shipped_at: null
---

# STAGE-018: capture-validation caps + the v0.5.0 audit backlog

## What This Stage Is

**Reframed 2026-08-09.** This began as a parking lot for the v0.5.0 audit's
LOW/NIT items. It now has an anchor that is not a nit: **the capture-validation
caps are wrong, and they are wrong in a way that already rejects a third of the
corpus.** The nits ride along because most of them live in the same code.

The coherent outcome: **`capture.Validate` becomes a rule you can defend** —
every cap derived from what the field is for rather than inherited from a JSON
schema, enforced on *every* ingress path including edit, and expressed once
rather than duplicated. The audit nits are the same "harden the edges" pass they
always were, now with a reason to schedule them.

Three coupled pieces, all touching `internal/capture`:

1. **Re-shape the caps.** `MaxImpact = 256` is exceeded by **81 of 274 entries
   (30%)**, with a **median of 227** — a cap sitting in the middle of ordinary
   usage, not catching outliers. `MaxTitle = 200` is exceeded by 33 of 348.
   `MaxTags = 64` caps the *comma-joined string* rather than each tag, which
   penalises using many tags (legitimate) instead of one absurd tag (the actual
   abuse) — a leftover from DEC-004's comma-joined era that DEC-015's normalised
   `tags` table made obsolete. `project`/`type`/`description` are fine and must
   NOT be widened reflexively (3x, 6x and 48x headroom respectively).
2. **Enforce on the edit path.** `capture.Validate` is called from
   `cli/add.go` (x2), `cli/add_json.go` and MCP `brag_add` — but **not** from
   edit / `Store.Update`. This is the pre-existing backlog item, and it is
   **blocked on (1)**: wiring it up first would make every over-cap entry
   uneditable, so a user could not fix a title without deleting tags.
3. **Express each cap once.** `cli/project.go:162` hard-codes `len(name) > 64`
   instead of referencing `capture.MaxProject`, with a comment directly above it
   reasoning about keeping the two in agreement (DEC-036's soft-match
   invariant) — exactly the case for one constant rather than two a comment asks
   a reader to align.

## Why Now

The original answer was "not now, deliberately." That changed when a user hit
the tags cap in real use and the investigation found the pattern was general,
measured against the live corpus (347 entries): the caps were never derived.
DEC-012 introduced them for the `add --json` schema and records **no rationale
for any value**; SPEC-064 then correctly unified per-path validation and
generalised those numbers to every path — which retroactively made ~30% of the
existing corpus unwritable. Nobody measured that at the time.

It is PROJ-006's business rather than PROJ-007's: the coupling is to items
already in this project, and `impact` is the field this project's own memory
slice reads. Deferring across a project boundary would strand that coupling.

## Success Criteria

- Every cap has a **recorded rationale** tied to what the field is for. No
  surviving number that nobody can defend.
- **No entry in the existing corpus is unwritable.** The measured over-cap
  populations (81 impacts, 33 titles, 9 tag-strings) either fit the new caps or
  are explicitly grandfathered — stated, not discovered.
- `capture.Validate` runs on **every** ingress path, edit included, and that is
  true *after* the caps are fixed, never before.
- Each cap is defined **once**; no hard-coded duplicate survives.
- **DEC-044's budget model is re-derived, not silently invalidated.** It cites
  these caps twice for the memory slice's worst-case line ("626 bytes / 157
  tokens"), the argument that one long entry cannot starve a slice. A new
  `MaxImpact` moves that number.
- Each remaining audit item is fixed with a regression test, or closed WONTFIX
  with a one-line rationale.
- Full gate set green.

## Scope

### In scope

**The anchor — capture-validation caps** (see `capture-caps-were-inherited-not-derived`
in `guidance/questions.yaml` for the full measurement and reasoning):
- Re-shape `MaxImpact` / `MaxTitle`, and re-shape `MaxTags` from a joined-string
  cap into a per-tag + tag-count cap. Emits a DEC (a validation-contract change
  across four ingress paths, plus the DEC-044 re-derivation).
- Leave `MaxProject` / `MaxType` / `MaxDescription` alone — measured, ample.
- Grandfathering policy for existing over-cap rows, decided explicitly.

**The v0.5.0 audit LOW/NIT backlog** (inherited verbatim from the project brief):

- `mcp_install` atomic write (temp + rename).
- ~~MCP `list`/`search` negative-`limit` parity with the CLI.~~ **DONE — folded
  into SPEC-072** (STAGE-019), which already edited both handlers. A negative
  `limit` is now a tool error on `brag_list` and `brag_search`; `0` still means
  unlimited. This is the "fold into the stage touching the same code" rule in
  Design Notes doing its job.
- `search -foo` → clear cobra-flag error (not an FTS query).
- `brag project new` name cap — **identified**: `internal/cli/project.go:162`
  hard-codes `if len(name) > 64` instead of referencing `capture.MaxProject`,
  while the comment above it reasons about keeping the two in agreement
  (DEC-036's soft-match invariant). One-line fix, but it must travel WITH any
  change to the caps — see the question below. Also: run the edit /
  `Store.Update` path through `internal/capture.Validate`.
  > ⚠ **COUPLED — read before actioning.** Wiring the edit path to
  > `capture.Validate` would make a large slice of the corpus
  > **uneditable**: measured today, **81 of 274 impacts (30%)**, 33 titles and
  > 9 tag-strings already exceed their caps, because flag/editor mode did not
  > enforce them until SPEC-064 unified validation. Changing one of those
  > entries' titles would then require truncating impact or deleting tags. Resolve
  > `capture-caps-were-inherited-not-derived` in `guidance/questions.yaml` in the
  > same change, or explicitly grandfather existing rows. Do not action this
  > item alone.
- `brag spark` same-second exclusive-edge.
- backup-filename same-second collision.
- empty-`type` sentinel handling.
- export-md sort id-tiebreak.
- `MergeTags` position dup.
- double-wrapped db-path error.
- `$EDITOR`-with-spaces handling.

### Explicitly out of scope
- The deeper agent-native pillars (memory / signed provenance / capture
  completeness / benchmark) — separate PROJ-006 stages.
- `ParseSince` wall-clock impurity — **already handled**: STAGE-017 / SPEC-068
  folded in the `since.go` clock-seam fix (audit L4). Verify before re-listing;
  do not double-count.

## Spec Backlog

Ordered list of specs composing this stage. IDs assigned at creation.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [ ] (not yet written) — **the caps spec.** Re-shape the caps + the DEC +
      DEC-044 re-derivation + the grandfathering call. Then, in the SAME spec or
      one immediately after it, the edit-path wiring and the `cli/project.go`
      duplicate — they are the two items blocked on the caps decision.
- [ ] (not yet written) — the remaining audit nits, batched by touched package
      per the Design Notes (an "MCP/CLI parity" cluster, a "filesystem-write
      robustness" cluster). Split at framing.

**Count:** 0 shipped / 0 active / 2 pending (unframed — stage is `proposed`)

## Design Notes

- Batch by touched package, not by audit order — several of these live in the
  same file and share a test. One spec per cluster keeps review cheap.
- Prefer folding an item into a deeper-pillar stage when that stage already
  edits the same code, rather than a standalone visit. (SPEC-072 did exactly
  this with the negative-`limit` item — already ticked above.)
- **Ordering is load-bearing, not a preference.** The caps decision must land
  before the edit-path wiring. Reversed, the wiring makes 30% of the corpus
  uneditable; there is no ordering in which that is acceptable.
- The caps spec is the one that needs a **DEC**; the nits do not. Do not let the
  nits' cheapness set the review altitude for the caps change.
- Resist widening every cap. The measurement says `description` has 48x headroom
  and `impact` is over-subscribed — the fix is per-field, derived from what the
  field is *for*, not a uniform multiplier.

## Dependencies

### Depends on
- Nothing new for the nits — all against shipped v0.5.0 code.
- The caps spec depends on **DEC-044** (shipped, STAGE-019): its worst-case-line
  math has to be re-derived in the same change, not after.

### Enables
- Writing a brag without compressing the one field every digest reads. `impact`
  is what `brag impact`, `wrapped`, `story` and `memory` all surface, and what a
  spec's Reflection-Q4 is transcribed into.
- The edit path finally validating like every other ingress path.
- A marginally more robust substrate for the deeper PROJ-006 stages that build
  on the same MCP / capture / storage paths.

## Stage-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>

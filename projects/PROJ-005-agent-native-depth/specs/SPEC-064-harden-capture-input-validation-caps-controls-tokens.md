---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-064
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
  decisions: [DEC-004, DEC-011, DEC-027]
  constraints: [no-sql-in-cli-layer, errors-wrap-with-context, stdout-is-for-data-stderr-is-for-humans, test-before-implementation, storage-tests-use-tempdir]
  related_specs: [SPEC-046, SPEC-061]
---

# SPEC-064: harden capture input validation caps controls tokens

## Context

A pre-release audit of the capture ingress found three input-validation gaps,
all on the paths that create an entry. There are four such paths, and they had
drifted apart:

- `internal/cli/add.go` — flag mode (`runAddFlags`) and editor mode (`runAddEditor`)
- `internal/cli/add_json.go` — `add --json`
- `internal/mcpserver/server.go` — the `brag_add` MCP tool

`add --json` and `brag_add` each enforced field byte-caps inline; flag and
editor mode enforced **nothing**. None of the four rejected control characters,
and any path could smuggle a bogus `cost:`/`tokens:` value into the store via the
freeform `tags` field, bypassing the validated dedicated params.

This spec fixes all three (audit findings #5 length-cap parity, #7 control-char
rejection + a NUL LOW, and the narrow half of #3 reserved-numeric tags) once, at
a single shared validator, so the per-path drift that caused #5 cannot recur.

Parent stage: `STAGE-016` (PROJ-005 polish). Project: `PROJ-005` (agent-native
depth) — capture correctness underpins the provenance work SPEC-046/DEC-027 added.

## Goal

Every entry-creation path validates identical field byte-caps, rejects embedded
control characters in single-line fields (and NUL in the multi-line
description), and validates reserved `cost:`/`tokens:` tokens in freeform tags —
all through one shared validator so no path can drift.

## Inputs

- **Files to read:**
  - `internal/cli/add.go`, `internal/cli/add_json.go` — cli ingress
  - `internal/mcpserver/server.go`, `internal/mcpserver/provenance.go` — MCP ingress + numeric normalizers
  - `internal/storage/store.go` (`Store.Add`, `canonicalizeTags`) — the common sink
- **Related code paths:** `internal/editor/` (buffer parse feeds editor mode)

## Outputs

- **Files created:**
  - `internal/capture/validate.go` — shared, SQL-free, cobra-free validator (`Fields`, `Validate`) plus the moved `NormalizeCost`/`NormalizeTokens`/`isDecimal`
  - `internal/capture/validate_test.go` — unit tests (caps, control chars, reserved tags) + the migrated normalizer tests
- **Files modified:**
  - `internal/cli/add.go` — flag + editor modes call `capture.Validate`, wrap as `ErrUser`
  - `internal/cli/add_json.go` — inline caps replaced by `capture.Validate`
  - `internal/mcpserver/server.go` — inline caps replaced by `capture.Validate`; numeric params use `capture.NormalizeCost`/`NormalizeTokens`
  - `internal/mcpserver/provenance.go` — normalizers moved out (kept `reservedTag`/`stampProvenance`)
  - `internal/mcpserver/provenance_test.go` — normalizer tests migrated to capture
- **New exports:** `capture.Fields`, `capture.Validate`, `capture.NormalizeCost`, `capture.NormalizeTokens`, `capture.Max*`
- **Database changes:** none

## Acceptance Criteria

- [x] Flag mode rejects an over-cap field (title 201 bytes, etc.) with `ErrUser`; a field exactly at cap is accepted.
- [x] Editor mode rejects an over-cap field with `ErrUser`.
- [x] `add --json` and `brag_add` still reject over-cap fields (regression), sharing one implementation.
- [x] Single-line fields (title, tags, project, type, impact) reject C0 control bytes (NUL, tab, newline, CR) on flag, editor, `--json`, and MCP paths.
- [x] `description` (multi-line) accepts tab/newline but rejects NUL, on all paths.
- [x] An accepted entry's `brag list` output stays exactly one line.
- [x] A bogus `cost:`/`tokens:` token in freeform tags (`cost:-9`, `tokens:xyz`, `cost:$5`) is rejected on all paths; a valid `cost:12.50` and non-numeric `agent:x` are accepted.
- [x] Caps are byte counts (`len`), matching the SPEC-061 byte decision.

## Failing Tests

Written before build; the build made them pass.

- **`internal/cli/add_hardening_test.go`**
  - `TestAdd_FlagFieldCapsAreUserError` — each field over cap via flag mode → `ErrUser`, 0 entries.
  - `TestAdd_FlagTitleAtCapAccepted` — 200-byte title accepted.
  - `TestAdd_EditorOverCapFieldIsUserError` — over-cap impact via editor buffer → `ErrUser`.
  - `TestAdd_FlagControlCharSingleLineFieldIsUserError` — newline/tab/NUL/CR in title → `ErrUser`.
  - `TestAdd_FlagDescriptionAllowsNewlineRejectsNUL` — newline+tab accepted; NUL rejected.
  - `TestAdd_AcceptedEntryListsOnOneLine` — accepted entry lists on one line.
  - `TestAdd_FlagReservedNumericTagIsUserError` / `TestAdd_FlagValidReservedTagsAccepted` — (C).
- **`internal/mcpserver/server_hardening_test.go`**
  - `TestServer_AddRejectsControlCharsInTitle`, `TestServer_AddDescriptionNewlineOkNulRejected`, `TestServer_AddRejectsBadReservedTagInTags`, `TestServer_AddAcceptsValidReservedTagInTags`.
- **`internal/capture/validate_test.go`**
  - `TestValidate_*` (caps at/over boundary, control chars in every single-line field, description multiline, reserved numeric tags) + migrated `TestNormalizeCost`/`TestNormalizeTokens`.

## Implementation Context

### Decisions that apply

- `DEC-004` — tags are a comma-joined string, not an array; the reserved-tag check splits on `,` and the tags cap applies to that string.
- `DEC-011` — the created-entry output shape is unchanged; validation happens before insert.
- `DEC-027` — `cost:`/`tokens:` are validated reserved provenance tags; the same normalizers now guard freeform tags too.

### Constraints that apply

- `no-sql-in-cli-layer` — the validator is SQL-free; cli stays a thin shell.
- `errors-wrap-with-context` — each boundary wraps the validator error (cli → `ErrUser`; MCP → `brag_add: %w`).
- `stdout-is-for-data-stderr-is-for-humans` — rejections are errors (stderr/exit code), stdout stays empty.
- `test-before-implementation` — failing tests written and confirmed red first.
- `storage-tests-use-tempdir` — unchanged; new tests use the existing temp-DB helpers.

### Prior related work

- `SPEC-061` (shipped) — established byte-count caps; this spec keeps `len()` bytes for parity.
- `SPEC-046` / `DEC-027` (shipped) — introduced the reserved `cost:`/`tokens:` provenance tags this spec now guards at freeform ingress.

### Out of scope (for this spec specifically)

- Edit path (`Store.Update` / `brag edit`) — not a capture ingress; left as-is (future defensive spec if wanted).
- `agent:`/`model:`/`session:` freeform validation — intentionally left opaque; `--tags "agent:x"` is the documented CLI provenance path and must keep working.
- Rune/grapheme caps or Unicode normalization — caps remain byte counts per SPEC-061.

## Notes for the Implementer

- Validate the caller's **raw** tags on the MCP path *before* `stampProvenance`, so the 64-byte tags cap applies to what the caller sent (the post-stamp string is longer by design).
- Keep the `add --json` error prefix `--json input:` so existing message assertions hold; the shared validator emits quoted-field messages (`"title" exceeds 200-character limit`) that read correctly under every boundary's prefix.

---

## Build Completion

- **Branch:** `fix/spec-064-capture-input-hardening`
- **PR (if applicable):** see PR opened against `main` (PROJ-005 / STAGE-016 / SPEC-064)
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - none (reuses DEC-004/DEC-011/DEC-027; validator placement recorded here, not as a DEC)
- **Deviations from spec:**
  - Validator lives in a neutral `internal/capture` package called at each of the four boundaries, **not** inside `storage.Store.Add`. Rationale: the 64-byte tags cap must be checked on the caller's raw tags *before* MCP provenance stamping appends `agent:/model:/…` — a `Store.Add` chokepoint would only ever see the post-stamp string and reject legitimate input. Boundary calls also preserve each path's error typing (`ErrUser` for cli exit-1; tool error for MCP) without threading a new storage sentinel through `main.go`. Drift is still ended: one `Validate` definition, one set of caps.
- **Follow-up work identified:**
  - Optional: apply the same validation defensively to the edit path (`Store.Update`).

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing material; the audit findings were precise. The one judgment call — validator placement — resolved cleanly once the MCP tags-cap-before-stamp ordering was considered.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — The SPEC-061 byte-cap decision was the load-bearing one; it is now referenced.

3. **If you did this task again, what would you do differently?**
   — Nothing significant. Writing the four-path failing tests first made the shared-validator shape obvious.

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

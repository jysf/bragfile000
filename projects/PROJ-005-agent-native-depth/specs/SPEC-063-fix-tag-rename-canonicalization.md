---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-063
  type: bug                        # epic | story | task | bug | chore
  cycle: verify
  blocked: false
  priority: high
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
  decisions:
    - DEC-004   # tags comma-joined for MVP → comma is the separator
    - DEC-016   # tag mutation semantics (rename/merge, FTS re-sync)
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
  related_specs: []
---

# SPEC-063: fix tag rename canonicalization

## Context

A pre-release CLI audit flagged a HIGH-severity, silent data-corruption
bug in `brag tag rename <old> <new>` (part of STAGE-016 polish under
PROJ-005). The rename path (`internal/cli/tag.go` → `storage.RenameTag`,
raw `UPDATE tags SET name=?`) does NOT run `<new>` through the tag
canonicalization every add/edit ingress uses (`canonicalizeTags` in
`store.go`: split on comma, trim, drop empties). It only rejects
`new == ""` and `old == new`, then writes `<new>` verbatim.

That means a non-canonical tag name is stored and silently corrupts
membership on the entry's NEXT edit:
- **comma:** `brag tag rename auth "a,b"` succeeds; the tag projection
  becomes `"a,b,perf"`. A later `brag edit` re-splits `"a,b"` into two
  tags — the entry is silently re-tagged, the `a,b` row is orphaned, and
  `brag list --tag "a,b"` returns nothing.
- **whitespace-only:** `rename auth "   "` accepted (check is `== ""`,
  not trim); vanishes on next edit.
- **surrounding whitespace:** `rename auth "  spaced  "` stored untrimmed;
  trimmed on the next round-trip → drift/orphan.

## Goal

Make `brag tag rename` apply the SAME normalization/rejection the capture
paths enforce, so a renamed tag can never round-trip into different or
missing membership. Validation lives at the CLI input boundary (per
`no-sql-in-cli-layer`); the storage `RenameTag` signature is unchanged.

## Inputs

- **Files to read:** `internal/cli/tag.go` — the rename command;
  `internal/storage/store.go` (`canonicalizeTags`, `RenameTag`) — the
  canonical form to mirror.
- **Related code paths:** `internal/cli/edit.go` (the Update round-trip
  that exposes the corruption).

## Outputs

- **Files modified:** `internal/cli/tag.go` — trim the new name; reject
  (UserError) if empty after trim or if it contains a comma; use the
  trimmed value for the rename. `internal/cli/tag_test.go` — new
  failing-first tests.
- **New exports:** none.
- **Database changes:** none.

## Acceptance Criteria

- [x] `brag tag rename auth "a,b"` → UserError (exit 1), stdout empty,
      message on stderr; entry membership unchanged (`auth` still resolves,
      `a,b` never created).
- [x] `brag tag rename auth "   "` → UserError; membership unchanged.
- [x] `brag tag rename auth "  spaced  "` → succeeds, tag stored trimmed
      as `spaced`.
- [x] Valid rename (`auth`→`authz`) still works; membership/FTS correct.
- [x] Round-trip: after a valid rename, editing the entry preserves
      membership (no drift).
- [x] Old bug gone: after a rejected comma rename, editing the entry does
      not silently re-split/corrupt membership.

## Failing Tests

- **`internal/cli/tag_test.go`**
  - `TestTagCmd_RenameCommaRejected` — comma target rejected, stdout empty,
    membership untouched.
  - `TestTagCmd_RenameWhitespaceOnlyRejected` — whitespace-only rejected.
  - `TestTagCmd_RenameTrimsSurroundingWhitespace` — surrounding whitespace
    trimmed and accepted.
  - `TestTagCmd_RenameRoundTripPreservesMembership` — valid rename + edit
    preserves membership.
  - `TestTagCmd_RenameCommaBugStaysFixed` — rejected comma rename + edit
    does not corrupt membership.

## Implementation Context

*Read this section (and the files it points to) before starting
the build cycle. It is the equivalent of a handoff document, folded
into the spec since there is no separate receiving agent.*

### Decisions that apply

- `DEC-004` — tags are comma-joined for the MVP; the comma is the join
  separator, so a single tag name can never contain one.
- `DEC-016` — tag mutation semantics; `RenameTag` fires the `tags_au`
  trigger to re-sync FTS. Fix keeps that path intact (validation is
  added ahead of it, at the CLI boundary).

### Constraints that apply

- `no-sql-in-cli-layer` — the fix is pure string validation in `tag.go`;
  no `database/sql` import added to cli.
- `stdout-is-for-data-stderr-is-for-humans` — rejection writes nothing to
  stdout; the confirmation ("Renamed.") stays on stderr.
- `errors-wrap-with-context` — rejections use `UserErrorf` (wraps
  `ErrUser`); storage errors keep their `%w` wrapping.

### Out of scope (for this spec specifically)

- Changing the `storage.RenameTag` signature or adding server-side
  canonicalization (validation belongs at the CLI input boundary).
- `brag tag merge` — both operands must already exist as valid rows, so it
  cannot introduce a non-canonical name. Untouched.

## Notes for the Implementer

`canonicalizeTags` is unexported in the storage package and reachable only
by importing SQL-adjacent code, so it is not reused directly from cli
(would risk `no-sql-in-cli-layer`). The rename target is a single tag
token, so the relevant subset of canonicalization is just trim +
reject-comma + reject-empty — implemented explicitly in `tag.go`.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `fix/spec-063-tag-rename-canonicalize` (stacked on
  `fix/spec-062-sqlite-concurrency`)
- **PR (if applicable):** see PR opened against `main`
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - none — the fix follows DEC-004/DEC-016; no new decision needed.
- **Deviations from spec:**
  - none.
- **Follow-up work identified:**
  - none.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing; the audit finding pinpointed the file, line, and repro.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. The three constraints and DEC-004/DEC-016 covered it.

3. **If you did this task again, what would you do differently?**
   — Consider a shared canonicalization helper reachable from cli without
   SQL, so add/edit/rename share one code path instead of a duplicated
   subset — but that is a refactor, not this fix.

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

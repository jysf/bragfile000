---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-089
  type: bug                        # epic | story | task | bug | chore
  cycle: build                     # frame | design | build | verify | ship
  blocked: false
  priority: high
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-09-07

references:
  decisions: [DEC-009, DEC-007, DEC-046]
  constraints: [stdout-is-for-data-stderr-is-for-humans, test-before-implementation]
  related_specs: [SPEC-085, SPEC-064]
---

# SPEC-089: the buffer silently drops a field, and edit gives no applied/no-op signal

> **Cycle: frame+design (collapsed by user direction).** **GO** at complexity
> **S**. Frame and design were written in one session at the user's request
> while triaging field feedback from heavy agent use of `brag` (2026-09-06→08),
> reversing §6's separate-session default for this bug spec only. A build
> session picks it up from here: the `## Failing Tests` are written; make them
> pass. Two independent defects, one PR, because both are one-line-ish capture
> integrity fixes on adjacent surfaces (`internal/editor` parse + `brag edit`
> output) and neither is worth a branch of its own.

## Context

This spec exists because of a concrete field report. An agent auditing a
450-entry corpus over two days drove every `brag` verb under script and found
two capture-integrity defects — both about the corpus quietly holding
something other than what the author wrote, which is the exact failure PROJ-008
("The Mirror and the Honest Corpus") is named for.

**Bug A — a duplicated header silently discards a field.** `editor.Parse`
([internal/editor/editor.go:81](../../../internal/editor/editor.go)) reads the
DEC-009 buffer with `net/textproto` and pulls each field via `hdr.Get(key)`.
`Get` returns only the *first* value for a repeated key; every later value is
dropped with no error. A buffer that ends up with two `Impact:` lines (a
scripted `$EDITOR` shim that inserts rather than replaces a header is the
reported cause) keeps one and silently loses the other. Reachable from both
`brag add` (editor mode) and `brag edit`. The reporter hit this: a double-apply
where a second `Type:` was tolerated silently — it happened to be the same
value, so nothing broke, but a duplicate `Title:` or `Impact:` would have
discarded real content.

**Bug B — `brag edit` has no machine-distinguishable applied/no-op signal.**
`runEdit` writes `Updated.` (on write) and `No changes.` (on no-op) both to
**stderr** and returns nil either way ([edit.go:90,137](../../../internal/cli/edit.go)).
That stderr choice is *correct* — those are human messages, and the
`stdout-is-for-data-stderr-is-for-humans` constraint pins them there. But it
leaves a real gap: **both outcomes produce empty stdout and exit 0**, so a
batch script cannot tell "applied" from "no-op" from either the data channel or
the exit code. The reporter's 13-entry batch read empty stdout, reported
`0 updated, 13 failed` while having applied all 13, then "retried" one and
double-applied it. The fix that respects the constraint is to emit the mutated
entry's **ID on stdout when a write happens** (and nothing on no-op) — exactly
what `brag add` already does ([add.go:179](../../../internal/cli/add.go)). The
ID is data; it belongs on stdout; its presence *is* the applied signal.

The reporter's other findings were triaged separately and are **out of scope
here** (see that section) — several are already shipped (`delete --yes`, the
field caps enforced by `internal/capture`), one is a design gap needing a DEC
(project-registry drift), and the rest are net-new features (write-time dup
detection, relationship links, `verified_at`, `brag lint`). This spec is only
the two true correctness bugs.

**Stage placement.** This is a `bug` spec attached to the active stage
(STAGE-023); it is *not* part of that stage's failure-capture narrative
backlog and does not gate the stage's ship. It rides here because it is
current work and because a silently-dropped field is a corpus-honesty defect —
PROJ-008's thesis — not because it extends the `brag learn` arc.

## Goal

`editor.Parse` rejects a buffer that repeats any of the five canonical headers
instead of keeping the first and dropping the rest; and `brag edit` prints the
edited entry's ID to stdout when (and only when) it writes a change, so a
script can distinguish applied from no-op on the data channel.

## Inputs

- **Files to read:** `internal/editor/editor.go` — `Parse`, the `hdr.Get`
  calls, the "Unknown headers are silently ignored" contract.
- **Files to read:** `internal/cli/edit.go` — `runEdit`, the `!changed`
  branch, the trailing `Updated.` line.
- **Files to read:** `internal/cli/add.go:179` — the existing `add` stdout-ID
  contract Bug B mirrors.
- **Related code paths:** `internal/cli/add_json.go` — the `--json` ingress
  does NOT go through `editor.Parse` (it decodes JSON), so Bug A does not touch
  it; confirm during build that no JSON test regresses.

## Outputs

- **Files modified:** `internal/editor/editor.go` — `Parse` gains a
  duplicate-canonical-header check between `ReadMIMEHeader` and field
  extraction; returns a descriptive error naming the offending header.
- **Files modified:** `internal/cli/edit.go` — on the write path, print
  `inserted.ID` (the `Update` return) to `cmd.OutOrStdout()` before the
  stderr `Updated.` line. No change to the `!changed` / no-op path.
- **Files modified:** `internal/editor/editor_test.go`,
  `internal/cli/edit_test.go` — new failing tests (below).
- **New exports:** none (behavior change only; `Parse`/`runEdit` signatures
  unchanged).
- **Database changes:** none.
- **New decisions (emit at build):**
  - `DEC-050` — the editor buffer rejects a repeated canonical header rather
    than silently keeping the first (refines DEC-009's parse contract).
  - `DEC-051` — `brag edit` emits the mutated entry ID on stdout as the
    applied/no-op signal, silent on no-op (extends the DEC-007/`add` stdout
    contract to `edit`).

## Acceptance Criteria

- [ ] A buffer with two `Title:` headers (differing values) makes `Parse`
      return a non-nil error that names `title`; the first value is NOT
      silently returned.
- [ ] Same for a repeated `Impact:` (the data-loss case): differing values →
      error, not first-wins.
- [ ] Duplicate detection is case-insensitive per `net/textproto`
      canonicalization: `Title:` + `title:` counts as a duplicate.
- [ ] A repeated **unknown** header (e.g. two `X-Note:`) is still silently
      ignored — the guard covers only the five canonical keys, preserving the
      existing "unknown headers ignored" contract.
- [ ] `brag edit <id>` that changes a field prints exactly `<id>\n` to stdout
      and `Updated.` to stderr; stdout carries only the ID.
- [ ] `brag edit <id>` with an unchanged buffer prints nothing to stdout and
      `No changes.` to stderr (the applied/no-op distinction is observable on
      stdout alone).
- [ ] `brag edit <id>` on a buffer with a duplicated header returns a UserError
      (exit 1), writes nothing to stdout, and leaves the stored row unchanged
      (no partial write) — Bug A's fix reaches the edit path for free through
      the existing `UserErrorf("invalid buffer: %v", err)` wrap.
- [ ] `just test` and `just lint` pass; no existing add/edit/JSON test regresses.

## Failing Tests

Written during design, BEFORE build. Make these pass in build.

- **`internal/editor/editor_test.go`**
  - `"TestParse_DuplicateTitleHeaderIsError"` — buffer with two `Title:` lines
    (values `"a"` then `"b"`); asserts `Parse` returns a non-nil error whose
    message contains `title` and `duplicate`; asserts the returned `Fields` is
    the zero value (nothing salvaged). This is the load-bearing failure: today
    `Parse` returns `Title:"a"`, err nil.
  - `"TestParse_DuplicateImpactHeaderIsError"` — two `Impact:` lines, differing
    values; asserts error naming `impact`. The corpus-data-loss case.
  - `"TestParse_DuplicateHeaderIsCaseInsensitive"` — `Title: a\n` then
    `title: b\n`; asserts error (textproto canonicalizes both to `Title`, so
    the values collide and must be caught).
  - `"TestParse_DuplicateUnknownHeaderStillIgnored"` — valid single `Title:`
    plus two `X-Note:` lines; asserts `Parse` succeeds (err nil) and returns the
    Title. Pins the guard's scope to the five canonical keys and protects the
    existing "unknown headers silently ignored" contract from over-tightening.

- **`internal/cli/edit_test.go`**
  - `"TestEditCmd_PrintsIDToStdoutOnUpdate"` — seed an entry; `editFn` returns a
    buffer with a changed Title; run `edit <id>`; assert `outBuf` trimmed ==
    the entry's ID, `errBuf` contains `Updated.`, and the stored Title changed.
    Fails today: `outBuf` is empty on update.
  - `"TestEditCmd_NoChangesEmitsNoStdout"` — `editFn` returns the buffer
    unchanged; assert `outBuf.Len() == 0` and `errBuf` contains `No changes.`
    Boundary guard: proves the new stdout write on the apply path does not leak
    onto the no-op path (passes today; must keep passing).
  - `"TestEditCmd_DuplicateHeaderIsUserErrorNoWrite"` — seed an entry; `editFn`
    returns a buffer that changes Impact but injects a second `Impact:` line;
    assert the command returns an `ErrUser` (per `errors.As`), `outBuf` is
    empty, `errBuf` mentions `invalid buffer`, and the stored row is byte-for-
    byte unchanged. Proves Bug A's fix reaches edit and that rejection is
    atomic (no partial update).

## Implementation Context

*Read this section (and the files it points to) before starting the build
cycle.*

### Decisions that apply

- `DEC-009` — pins the editor buffer format (`net/textproto` headers, blank
  line, markdown body). Bug A refines its *parse* contract: a repeated
  canonical header was undefined behavior that resolved to first-wins; it is
  now a hard reject. DEC-050 records this.
- `DEC-007` — inline positional-arg validation returns the `ErrUser` sentinel
  for user-facing exit codes; `edit` already wraps a parse failure as
  `UserErrorf("invalid buffer: %v", err)`, so Bug A surfaces correctly on the
  edit path with no extra wiring.
- `DEC-046` — validate-on-changed. Ordering note: the duplicate-header reject
  runs in `Parse`, strictly *before* `capture.ValidateChanged`, so a
  malformed buffer never reaches the cap/validation layer.

### Constraints that apply

- `stdout-is-for-data-stderr-is-for-humans` (blocking, `internal/cli/**`) —
  Bug B is *defined by* this constraint: the entry ID is data (stdout), the
  `Updated.`/`No changes.` prose stays human (stderr). Do not move the prose.
- `test-before-implementation` — the failing tests above are the contract;
  run `go test ./...` first and confirm the two load-bearing ones fail for the
  asserted reason (dup-header returns first-wins; edit stdout empty on update),
  not a compile error.

### Prior related work

- `SPEC-064` / `DEC-046` — created `internal/capture` to end per-path
  validation drift. Bug A is the same class one layer up (the *parser*, not the
  validator, was the silent-drop site).
- `SPEC-085` — `brag learn`, the stage's failure-capture verb, uses the same
  `editor.Parse`; it inherits Bug A's fix for free.

### CLI-test conventions (AGENTS.md §9)

- Use separate `outBuf` / `errBuf` (`cmd.SetOut` / `cmd.SetErr`) and assert no
  cross-leakage — the ID-on-stdout test must assert `errBuf` has the prose and
  `outBuf` has *only* the ID.
- The edit harness is `newRootWithEdit(t, editFn)` + `seedEditEntry` +
  `getEntry` (already in `edit_test.go`); reuse them.

### Out of scope (for this spec specifically)

Create a new spec rather than pulling any of these in:

- **`brag delete` printing its ID on stdout.** Same shape as Bug B, but `delete`
  already has a confirmation prompt and clear semantics; do it as its own small
  spec if wanted. Noted as a follow-up.
- **Reserved-header duplicate stamping.** `agent:`/`model:`/etc. are *tags*, not
  headers, and are joined into one `Tags:` line; the header guard does not touch
  them.
- **The rest of the field report (design work, not this bug fix):**
  project↔registry drift warning on `add` (touches DEC-017's free-text choice —
  needs a DEC); `type` near-miss/blank warning + `brag lint`; write-time
  duplicate detection on `add`; relationship links (`supersedes`/`corrects`);
  `verified_at`/staleness; and surfacing `brag learn`/`failed` more prominently
  (thematically STAGE-023 — route there). Docs gaps to route to a docs spec:
  the 1024-char `impact` cap absent from `add --help`; the `$EDITOR` buffer
  format undocumented in `edit --help`; and the fact that provenance
  (`agent:`/`model:`) is auto-stamped **only on the MCP ingress path**, never
  on CLI `brag add` — which is what made the reporter's scripted CLI entries
  look un-stamped.

## Notes for the Implementer

- **Bug A shape.** After `tp.ReadMIMEHeader()`, `hdr` is a
  `textproto.MIMEHeader` (= `map[string][]string`). Guard the canonical keys
  with `len(hdr["Title"]) > 1`, etc. — use the textproto-canonical form
  (`Title`, `Tags`, `Project`, `Type`, `Impact`), which is what
  `ReadMIMEHeader` stores. Return e.g.
  `fmt.Errorf("parse buffer: duplicate %q header (a field may only appear once)", name)`.
  Iterate a fixed slice of the five canonical keys so the check is exhaustive
  and greppable; do not iterate `hdr` (that would catch unknown dupes and break
  the ignored-unknown contract). Keep the existing `io.EOF` tolerance on
  `ReadMIMEHeader`.
- **Bug B shape.** `Update` already returns the updated entry; print its `.ID`
  via `fmt.Fprintln(cmd.OutOrStdout(), updated.ID)` on the write path, then the
  existing `Fprintln(cmd.ErrOrStderr(), "Updated.")`. `runEdit` currently
  discards the `Update` return with `_`; bind it. Leave the `!changed` branch
  untouched.
- **DEC-050 / DEC-051** each need at least one paired failing test above
  (AGENTS.md §9: a locked decision without a paired test is aspirational) —
  they do (`TestParse_Duplicate*` and `TestEditCmd_PrintsIDToStdoutOnUpdate`
  respectively). Write the DEC files during build with honest confidence
  (both are small, well-precedented contract tightenings → ~0.9).
- **Fail-first check.** Before touching implementation, run the four editor
  tests and confirm the three dup-header ones fail (the fourth,
  unknown-dupe-ignored, should already pass), and confirm
  `TestEditCmd_PrintsIDToStdoutOnUpdate` fails on empty stdout.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `fix/spec-089-editor-integrity`
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - `DEC-050` — editor buffer rejects a repeated canonical header
  - `DEC-051` — `brag edit` emits the mutated entry ID on stdout
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - `brag delete` ID-on-stdout (mirror of Bug B) — own spec if wanted.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — <answer>

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — <answer>

3. **If you did this task again, what would you do differently?**
   — <answer>

---

## Reflection (Ship)

*Appended during the **ship** cycle.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>

4. **What can a user do now that they couldn't before?** — one sentence.
   — <answer>
</content>
</invoke>

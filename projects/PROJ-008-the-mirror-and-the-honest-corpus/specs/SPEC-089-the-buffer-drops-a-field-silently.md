---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-089
  type: bug                        # epic | story | task | bug | chore
  cycle: verify                    # frame | design | build | verify | ship
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
  - `DEC-051` — the editor buffer rejects a repeated canonical header rather
    than silently keeping the first (refines DEC-009's parse contract).
  - `DEC-052` — `brag edit` emits the mutated entry ID on stdout as the
    applied/no-op signal, silent on no-op (extends the DEC-007/`add` stdout
    contract to `edit`).

  > **Renumbered 2026-09-07, from `DEC-050`/`DEC-051`. Do not renumber back.**
  > This spec and **SPEC-086** both claimed `DEC-050` in prose while neither had
  > built, so whichever built first would have taken it and left the other
  > pointing at the wrong record. SPEC-086's claim is load-bearing and far more
  > entangled — it is cited in SPEC-086's Fork A, four times on STAGE-023, and
  > about fourteen times in the **shipped** SPEC-087, including `LD7`, `AC-11`
  > and simulation `S-1`, whose whole justification was *"`DEC-050` stays free
  > for SPEC-086."* This spec's claim was five lines in one unbuilt document, so
  > it is the one that moves. `DEC-050` remains reserved for SPEC-086.
  >
  > This is the **third** instance of an id reserved in prose rather than by a
  > file — after the `SPEC-088` near-miss and corpus entry 433, which recorded
  > the same mechanism for the SPEC id class. `scripts/_lib.sh` derives
  > `next_id` from **filenames**, and no equivalent guard exists for `DEC-*`
  > ids at all. Routed to **SPEC-088**, which owns the harness-id family.

## Acceptance Criteria

- [x] A buffer with two `Title:` headers (differing values) makes `Parse`
      return a non-nil error that names `title`; the first value is NOT
      silently returned.
- [x] Same for a repeated `Impact:` (the data-loss case): differing values →
      error, not first-wins.
- [x] Duplicate detection is case-insensitive per `net/textproto`
      canonicalization: `Title:` + `title:` counts as a duplicate.
- [x] A repeated **unknown** header (e.g. two `X-Note:`) is still silently
      ignored — the guard covers only the five canonical keys, preserving the
      existing "unknown headers ignored" contract.
- [x] `brag edit <id>` that changes a field prints exactly `<id>\n` to stdout
      and `Updated.` to stderr; stdout carries only the ID.
- [x] `brag edit <id>` with an unchanged buffer prints nothing to stdout and
      `No changes.` to stderr (the applied/no-op distinction is observable on
      stdout alone).
- [x] `brag edit <id>` on a buffer with a duplicated header returns a UserError
      (exit 1), writes nothing to stdout, and leaves the stored row unchanged
      (no partial write) — Bug A's fix reaches the edit path for free through
      the existing `UserErrorf("invalid buffer: %v", err)` wrap.
- [x] `just test` and `just lint` pass; no existing add/edit/JSON test regresses.

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
  now a hard reject. DEC-051 records this.
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
- **DEC-051 / DEC-052** each need at least one paired failing test above
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
- **PR (if applicable):** opened, not merged — see PR description for number.
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - `DEC-051` — editor buffer rejects a repeated canonical header
  - `DEC-052` — `brag edit` emits the mutated entry ID on stdout
- **Deviations from spec:**
  - `TestEditCmd_HappyPath`'s pre-existing assertion (`outBuf.Len() != 0` is
    an error) was invalidated by DEC-052 — a real write now legitimately
    prints the ID to stdout. Updated the assertion to expect the mutated ID
    rather than emptiness, per the AGENTS.md §9 premise-audit convention
    (an inverted/changed behavior gets its existing-test consequences
    enumerated, not discovered as a build-time surprise). Not listed under
    the spec's `## Outputs`/`## Files modified` — a design-time gap, not a
    build-time deviation from what was designed, but flagged here because it
    is the one place build touched a test whose premise (not just its
    presence) the spec didn't anticipate.
  - `TestEditCmd_DuplicateHeaderIsUserErrorNoWrite`'s spec text says
    "`errBuf` mentions `invalid buffer`." The root command sets
    `SilenceErrors: true` (`root.go:37`), so `UserErrorf`'s message surfaces
    on the returned `error`, not on `errBuf` — cobra's `Execute()` never
    writes it to stderr itself; that's `main.go`'s job in production, which
    the test harness doesn't invoke. Wrote the assertion against
    `err.Error()` instead, matching the existing sibling pattern in
    `TestEditCmd_ChangedImpactOverCapIsUserErrorNoWrite`. Same defect class
    as the AGENTS.md §12(b) "design-time pre-flight covers the test's own
    expected-value literals" rule, one layer up: the literal here was which
    *channel* carries a `UserErrorf` message, not a value, and it wasn't
    checked against `root.go`'s actual `SilenceErrors` config at design.
  - Fail-first sequencing: wrote the seven new tests against the
    pre-mutation baseline first (stashed the two implementation edits),
    confirmed the expected 5 load-bearing failures + 2 already-passing
    boundary guards, then restored the implementation. See reflection Q3.
- **Follow-up work identified:**
  - `brag delete` ID-on-stdout (mirror of Bug B) — own spec if wanted.

### Mutation protocol (§12)

Two mutants, both against the fixed (post-build) code, backed up to
`/tmp` before mutating and restored from that backup (never `git checkout`,
since both files carried uncommitted work) after confirming the gate caught
each one. Hash confirmed to move before running the gate, and confirmed to
return to its pre-mutation value after restore, per §12's clauses (1) and
its refinement.

- **M-1** (`internal/editor/editor.go`) — changed the duplicate-header
  threshold `len(hdr[key]) > 1` → `len(hdr[key]) > 2` (only 3+ repeats would
  trigger the guard, letting the spec's 2-line duplicate cases through
  silently). Pre-mutation hash
  `928bf2324912d421c26dd028d36187eaecc941bdf5c8145148cc003cd27c5a4b` →
  post-mutation `f785cdcdf758c5b22298c21e85d444f2da5cee99a02111e7fda4f79ec546a925`
  (confirmed different). Gate: `TestParse_DuplicateTitleHeaderIsError`,
  `TestParse_DuplicateImpactHeaderIsError`, and
  `TestParse_DuplicateHeaderIsCaseInsensitive` all turned red (`expected
  error, got nil`); `TestParse_DuplicateUnknownHeaderStillIgnored` (unaffected
  by this key) stayed green. Restored; hash returned to
  `928bf232…`; the four tests passed again.
- **M-2** (`internal/cli/edit.go`) — changed the ID print's destination on
  the write path from `cmd.OutOrStdout()` to `cmd.ErrOrStderr()` (the ID
  would land on the wrong channel). Pre-mutation hash
  `448f717b418cd46d525cf0b246b3ed2441d926da2399932821ffc70d720ff15f` →
  post-mutation `59495d687bb8f5a088baa41b466b4ac263ab02de4e90630c0ecd66adef6b5347`
  (confirmed different). Gate: `TestEditCmd_HappyPath` and
  `TestEditCmd_PrintsIDToStdoutOnUpdate` both turned red; the boundary guard
  `TestEditCmd_NoChangesEmitsNoStdout` (no-op path, untouched by this
  mutation) stayed green, proving the new stdout write is correctly scoped
  to the write path only. Restored; hash returned to `448f717b…`; all 15
  `TestEditCmd_*` tests passed again.

Positive controls run alongside (per the "a green guard is not evidence"
trap): `TestParse_HappyPath`, `TestRoundTrip_AllFields`, and
`TestParse_UnknownHeadersIgnored` all stayed green throughout — a valid
single-header buffer and a buffer with unknown headers both still parse.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing structurally unclear; the two Deviations above are places
   where a spec-stated literal (a test assertion's target channel) didn't
   match the actual codebase mechanism (`SilenceErrors: true`). Both were
   cheap once found — five minutes each, caught by the fail-first run
   itself rather than by a later gate.

2. **Was there a constraint or decision that should have been listed but
   wasn't?**
   — Not a constraint, but `root.go:37`'s `SilenceErrors: true` is exactly
   the kind of "test's own expected-value literal" AGENTS.md §12(b)'s
   extension already names — a design-time pre-flight that grepped
   `root.go` for `SilenceErrors` before writing "errBuf mentions..." would
   have caught it before build. Not proposing a new rule; this is a same-
   outcome instance of the existing one.

3. **If you did this task again, what would you do differently?**
   — Nothing on sequencing — writing the tests first, confirming the
   fail-first run in one shot (stash implementation, run new tests, restore
   implementation), then applying the fix was clean and gave the pinned
   before/after evidence in `## Failing Tests`'s own terms. The two
   deviations above are the only things I'd fix earlier: run the test file
   itself, not just the artifact under test, against `root.go` before
   locking the exact assertion target.

---

## Verification

*Filled in at the end of the **verify** cycle. Build's gates passed and both
fixes were independently reproduced end to end before this cycle started;
none of that is re-litigated here. This section records only what a passing
build could not see.*

- **Branch:** `verify/spec-089-editor-integrity`, off `main` at `59b230d`
  (PR #208 merged 2026-09-08T07:39:05Z as a squash; no stacking).
- **Verdict:** ⚠ **PUNCH LIST — all items fixed in this cycle.** Both defects
  are genuinely fixed and reach every path they claim to. Four findings, three
  fixed here, one routed.

### The attack list

| # | Attack | Result |
|---|---|---|
| A | Does Bug A's fix reach **every** `Parse` caller, not just `edit`? | **HOLDS** |
| B | Is the fix a breaking change for a **legitimate** buffer a template can emit? | **V-F2 — two of three generators unguarded. FIXED HERE** |
| C | Is the five-key `canonicalHeaders` slice complete, and is its completeness *enforced*? | **V-F1 — complete, but three of five keys unenforced. FIXED HERE** |
| D | Does Bug B's signal survive `edit`'s other output modes / a post-write error? | **HOLDS** |
| E | Does the reporter's 13-entry batch now report correctly? | **HOLDS** |
| F | Are the two DEC records' own claims true? | **V-F4 — DEC-051's `## Validation` overclaimed. CORRECTED HERE** |
| G | Do the docs still describe the old contract? | **V-F3 — `api-contract.md` + CHANGELOG stale. FIXED HERE** |
| H | Does the same defect exist on the ingress the spec scoped out? | **V-F5 — yes, and it is now the *only* silent one. ROUTED → STAGE-023** |

### A — Bug A's blast radius. HOLDS on all three editor ingresses

`editor.Parse` has exactly three non-test callers repo-wide
(`grep -rn 'editor\.Parse' --include='*.go' .`, excluding the unrelated
pre-existing `.claude/worktrees/` checkout): `internal/cli/edit.go:94`,
`internal/cli/add.go:198`, `internal/cli/learn.go:117`. All three wrap
identically as `UserErrorf("invalid buffer: %v", err)` and return **before**
any `storage.Open`/insert, so the rejection is atomic by construction, not by
assertion. The spec's acceptance criteria exercise only `edit`; the other two
were driven end to end here against a `t.TempDir()`-equivalent scratch DB:

```
$ EDITOR=<shim writing two Impact: lines> brag --db "$DB" add
exit=1
stdout=[]
stderr=[brag: user error: invalid buffer: parse buffer: duplicate "impact" header (a field may only appear once)]
row count before = 1 ; after = 1

$ EDITOR=<same shim> brag --db "$DB" learn
exit=1
stdout=[]
stderr=[brag: user error: invalid buffer: parse buffer: duplicate "impact" header (a field may only appear once)]
row count before = 1 ; after = 1
```

`brag learn` also rejects a duplicated `Type:` — a header its own
`FailureTemplate` deliberately omits and whose value it overwrites anyway.
Over-strict in the narrow sense, but correct: the reject runs in `Parse`,
which cannot know the caller will discard the field, and a buffer the user
typed twice is not a buffer to guess about.

### C — V-F1: the canonical-key slice is complete, but its completeness was not enforced. FIXED

`canonicalHeaders` matches the five `hdr.Get` calls in `Parse` exactly, and
the five headers `Render` emits. That was verified by inspection **and** it
was not the question. The question is whether anything *keeps* it matching —
and nothing did. Only `Title` and `Impact` had tests. Mutation, §12 protocol,
hash confirmed to move before the gate ran (`shasum -a 256`, restore from
`/tmp` backup, hash confirmed to return):

| mutant | `internal/editor/editor.go` after | `go test -count=1 ./...` (pre-fix) |
|---|---|---|
| drop `"Title"` | `7276e3f0…`* | RED (2 tests) |
| drop `"Tags"` | `fd8b48d3…`* | **GREEN — all 14 packages ok** |
| drop `"Project"` | `69c91dcc…`* | **GREEN** |
| drop `"Type"` | `bbaca034…`* | **GREEN** |
| drop `"Impact"` | `1df5a621…`* | RED (2 tests) |

Edit, stated so the hash is reproducible: replace the line
`var canonicalHeaders = []string{"Title", "Tags", "Project", "Type", "Impact"}`
with the same line minus the named key. Baseline
`928bf2324912d421c26dd028d36187eaecc941bdf5c8145148cc003cd27c5a4b`.

> \*A first pass used `perl -0pi -e 's/"Tags", //'`, which also matched
> `write("Tags", f.Tags)` inside `Render` and produced a compile error — the
> right verdict for the wrong reason. Those five results were **discarded**
> under §12's third clause (*confirm the mutant changed only what you
> intended*) and re-run against the slice line alone; the hashes above are the
> re-run. A second pass mis-exported the mutator as a shell function and
> applied nothing at all; the hash-gated helper refused to run the gate and
> discarded the result, which is §12(b) working exactly as SPEC-087's verify
> built it. Both are recorded because a discarded probe that is not recorded
> looks like a probe that was never run.

Consequence: a sixth editable field added to `Render`/`Parse` and forgotten in
`canonicalHeaders` reintroduces this exact bug on that field, with a fully
green suite — the "green guard proves nothing" shape.

**Fixed by `TestParse_DuplicateGuardCoversEveryHeaderRenderEmits`**
(`internal/editor/editor_test.go`), which *derives* the set to test from
`Render`'s actual output instead of re-typing the list, and carries a
non-vacuity floor: the header count must equal
`reflect.TypeOf(Fields{}).NumField() - 1`, so a `Render` that stopped emitting
headers cannot make the loop iterate zero times and pass. All three previously-
green mutants now name the right key:

```
--- FAIL: TestParse_DuplicateGuardCoversEveryHeaderRenderEmits/Tags
    Parse tolerated a duplicate "Tags" header — canonicalHeaders is missing it,
    so that field is still silently droppable
```

and the floor fires on its own (delete `write("Tags", f.Tags)` from `Render`):
`Render emitted 4 headers [Title Project Type Impact], want 5`.

Honest limit: the derivation is anchored on `Render`. A field added to `Parse`
but never to `Render` is not covered — but such a field is already dead code
under `TestRoundTrip_AllFields`, which locks the two together.

### B — V-F2: the "no generator can emit a duplicate" property was two-thirds unguarded. FIXED

The fix converts a previously-harmless condition into a hard failure, so the
three buffer generators must be *incapable* of emitting a duplicate. Proved by
construction, not assertion — mutate each, then build a real binary and run the
real command:

| generator | mutant | pre-fix suite | user-facing effect |
|---|---|---|---|
| `Render` | write `Impact` twice | RED (`TestRoundTrip_AllFields`) | guarded already |
| `EmptyTemplate` | second `Impact:` line | RED (`TestEmptyTemplate_ParsesToMissingTitleError`) | guarded **by accident** |
| `EmptyTemplate` | second **`Title:`** line | **GREEN** | `brag add` broken for everyone |
| `FailureTemplate` | second `Impact:` line | **GREEN** | `brag learn` broken for everyone |

`EmptyTemplate`'s existing guard is coincidental: the test asserts the parse
error *contains* `"title"`, and a duplicate-**Impact** error preempts the
title error so the assertion fails — but a duplicate-**Title** error also
contains `"title"`, so it passes for the wrong reason. Driven for real from
binaries built from each mutant, with an ordinary user edit (fill in the
`Title:` line the template ships with, change nothing else):

```
$ EDITOR=<sed 's/^Title: $/Title: a real failure/'> brag-m4 --db "$DB" learn
exit=1
stderr=[brag: user error: invalid buffer: parse buffer: duplicate "impact" header (a field may only appear once)]
   # FailureTemplate mutant d153ec7d1a87… — 100% of `brag learn` invocations fail, suite green

$ EDITOR=<same> brag-m3 --db "$DB" add
exit=1
stderr=[brag: user error: invalid buffer: parse buffer: duplicate "title" header (a field may only appear once)]
   # EmptyTemplate mutant 83e9983e5b8e… — 100% of `brag add` editor-mode invocations fail, suite green
```

**Fixed by `TestTemplates_CannotEmitADuplicateHeader`**, which round-trips both
templates through `Parse` and fails on a `duplicate` error specifically (not on
the expected missing-`Title` error), then re-parses each with a filled-in
`Title:` to prove the template is still usable. Both mutants now RED.

### D — Bug B under the other output modes. HOLDS

- **No other output mode exists.** `brag edit --help` shows exactly two flags:
  `-h/--help` and the persistent `--db`. `brag edit 1 --format json` →
  `brag: user error: unknown flag: --format`, exit 1. DEC-052's Option-B
  rejection ("`edit` has no `--json`/`--format` flag and no other output mode")
  is factually correct.
- **No path writes and skips the ID.** After `s.Update` succeeds, `runEdit` has
  only the two `Fprintln`s and `return nil` — no error branch between them.
- **Broken pipe** is the one case where a write happens and no ID lands
  (`brag edit 1 | true` → exit **141**, empty stdout, empty stderr, row
  updated). `brag add` behaves identically under the same test, so this is
  `add`'s pre-existing contract inherited verbatim, not new unreliability —
  and exit 141 ≠ 0 is itself a distinguishable signal.
- **The signal is honest about what it means.** "ID on stdout" tracks *a row
  was written*, not *a field value differed*: an edit that changes only an
  unknown header (`X-Note:`) still prints the ID and still bumps `updated_at`
  (`07:47:43Z` → `07:47:45Z` measured), because `Launch`'s `changed` is a
  SHA-256 over the buffer. That is the right semantics for a batch driver.

### E — the reporter's 13-entry batch, replayed

The acceptance test the spec's criteria imply but never state: the reported
*symptom*, not its diagnosis. Driver = the reporter's own decision rule
(*success iff stdout is non-empty*); the scripted edit appends `" [reviewed]"`
to the title, so a double-apply is visible in the data.

**Pre-fix** (binary built from `b84636b`, the commit before #208):

```
--- pass 1 ---   0 updated, 13 failed        (all 13 had in fact applied)
--- pass 2 ---   0 updated, 13 failed        (the script retries the "failures")
1 entry 1 [reviewed] [reviewed]              ... all 13 double-applied
```

**Post-fix** (`59b230d`):

```
--- pass 1 ---  13 updated, 0 failed
1 entry 1 [reviewed]                          ... applied exactly once, no retry
```

Residual, reported as an observation rather than a defect: a *no-op* batch
still reports `0 updated, 13 failed` under that naive rule, because DEC-052
is deliberately silent on no-op. The retry is now harmless (a no-op retried is
a no-op), so the dangerous direction is closed. The full three-state space is
observable — applied `(id, exit 0)` / no-op `(empty, exit 0)` / error
`(empty, exit 1)` — but only from stdout **and** the exit code together;
stdout alone separates applied from not-applied, which is what the spec
claimed and delivered. Now stated explicitly in `docs/api-contract.md`.

### F — V-F4: DEC-051's `## Validation` claimed something its tests did not prove. CORRECTED

It read: `TestParse_DuplicateUnknownHeaderStillIgnored` "proves the guard's
scope is **exactly** the five canonical keys." It proves the *upper* bound
only. The lower bound was unpinned — that is V-F1, measured above. The
decision itself is right; the sentence was not. Corrected in place (no
`## Amendment` heading, so the inventory's amendment row does not move), citing
the two new tests that make the claim true.

### G — V-F3: the CLI contract document still described the old behaviour. FIXED

The spec's `## Outputs` enumerates two Go files, two test files and two DEC
files, and no documentation. AGENTS.md §9's premise audit (*status change →
planned doc references update*) was not run: `grep -rn 'brag edit' docs/
README.md CHANGELOG.md` reaches `docs/api-contract.md:233`, the repo's CLI
contract, which said —

> `- Saving a successful edit prints `Updated.` to stderr, exit 0.`

— with no mention of stdout, and listed exactly one user-error case for the
buffer (missing/empty `Title:`). Both of this spec's changes were invisible
there. Updated the `brag edit` and `brag add` (editor mode) contract blocks
for the stdout ID and the new duplicate-header rejection, and added a
`### Fixed` pair to `CHANGELOG.md`'s `[Unreleased]` — matching the precedent
of the two immediately preceding specs on this stage (SPEC-084 and SPEC-085
both updated the CHANGELOG at **build**, `d17bc0a` / `d8c69d7`).

### H — V-F5: the fourth ingress has the same defect, and is now the only silent one. ROUTED

The spec's `## Inputs` says the `--json` ingress "does NOT go through
`editor.Parse` (it decodes JSON), so Bug A does not touch it." True about the
code path — and it treats Bug A as a `Parse` defect rather than an *ingress*
defect. `internal/cli/add_json.go:24` uses `encoding/json`, whose behaviour for
a repeated object key is **last-wins, silently**:

```
$ echo '{"title":"json dup test","impact":"REAL VALUE","impact":"CLOBBERED"}' | brag --db "$DB" add --json
exit=0
stdout=[2]
stderr=[]
$ brag --db "$DB" list --format json
2 'json dup test' 'CLOBBERED'
```

So the corpus's two write ingresses now *disagree* about what a repeated field
means — the editor rejects, the scripted path silently takes the last — and the
silent one is the path an agent drives. This is the defect PROJ-008 is named
for, on the surface the field report was generated from. Not pulled in: it
needs a decision (reject vs. warn) and a `json.Decoder`-token pre-pass, since
`encoding/json` has no duplicate-key hook. Routed to STAGE-023's backlog as a
`bug`, with the reproduction.

### Gates — all five green on this branch

| gate | result |
|---|---|
| `go test -count=1 ./...` | ok, all 14 packages · **1086** `=== RUN` lines |
| `just test-docs` | `ALL OK` · **199** `OK:` lines / **198** distinct ids (S3 double-emits) |
| `just lint` | `0 issues.` |
| `gofmt -l .` | empty |
| `go vet ./...` | exit 0 |

Build's `1070 → 1077` claim was re-derived independently rather than taken on
trust, by counting `=== RUN` lines in a detached worktree at each commit:
`b84636b` → **1070**, `59b230d` → **1077**. This cycle's two new test functions
carry five and two subtests, so `1077 + 9 = 1086`. ✓

`just inventory` regenerated and pasted wholesale; **exactly one row moved**,
and it was not predicted in advance: `Go test functions 827 → 829`. The
`Decision records` row stays 50 and the amendment row stays 1 — the DEC-051
correction deliberately avoids an `## Amendment` heading.

### Corpus discipline

Live corpus re-derived read-only from a copy (never opened by a `brag` binary,
since `storage.Open` runs migrations): **454 entries, `max(id)` 466**. Entry
466 is `contextcore-pilot-harness` / `learned` — another project's write, not
this cycle's. Every reproduction above ran against an explicit `--db` under the
session scratchpad. `~/.bragfile/db.sqlite` is byte-identical to its state at
the start of this session:
`87d7e01f11e7d9badaf033a1af7e08f00fac92d1e095a49179555f537bc19676`.

### What this cycle changed

- `internal/editor/editor_test.go` — two new tests (+9 run cases). No
  production code changed: both fixes were correct as built.
- `decisions/DEC-051-…md` — corrected a false claim in `## Validation`.
- `docs/api-contract.md` — the `brag edit` and `brag add` editor-mode contract
  blocks, for both of this spec's behaviour changes.
- `CHANGELOG.md` — a `### Fixed` pair under `[Unreleased]`.
- `docs/engineering-practices.md` — regenerated inventory block.
- `projects/…/stages/STAGE-023-…md` — V-F5 routed onto the backlog.

### Not findings, checked and clean

- Build's two disclosed deviations are both correct. `root.go:37` does set
  `SilenceErrors: true`, so asserting on `err.Error()` rather than `errBuf` is
  the right call and matches the sibling
  `TestEditCmd_ChangedImpactOverCapIsUserErrorNoWrite`.
- `scripts/test-docs.sh` untouched; X3/Y3/Z7 green. SPEC-087's fix absorbed two
  new decision records with zero hand-edits on its first real exercise.
- `DEC-051`/`DEC-052` both carry `type: decision`; `decisions/DEC-050*` still
  matches nothing, so SPEC-086's reservation is intact.
- `NEXT-SESSION-PROMPT.md` was modified and uncommitted at session start and is
  left untouched — twelve consecutive cycles now.

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

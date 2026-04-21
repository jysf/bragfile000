---
task:
  id: SPEC-009
  type: story
  cycle: verify
  blocked: false
  priority: high
  complexity: M

project:
  id: PROJ-001
  stage: STAGE-002
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7
  created_at: 2026-04-20

references:
  decisions:
    - DEC-005  # integer autoincrement IDs
    - DEC-006  # cobra framework
    - DEC-007  # RunE-validated flags AND positional args
    - DEC-009  # editor buffer format (emitted during this spec's design)
  constraints:
    - no-cgo
    - no-sql-in-cli-layer
    - no-new-top-level-deps-without-decision
    - storage-tests-use-tempdir
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - timestamps-in-utc-rfc3339
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-002  # shipped; Store, Entry, scan patterns, timestamp discipline
    - SPEC-006  # shipped; show command, Store.Get, ErrNotFound,
                # DEC-007 positional-arg extension
    - SPEC-008  # shipped; delete command, stdin interaction pattern
---

# SPEC-009: `brag edit <id>` command + `internal/editor` package

## Context

Fifth spec in STAGE-002. Introduces the first genuinely new
architectural surface since SPEC-002: an `internal/editor` package
that handles `$EDITOR` spawning, temp-file round-trip, and
header+body buffer rendering/parsing. `brag edit <id>` is the first
consumer of that package and **is also the canonical update
mechanism for PROJ-001** — flag-based update (e.g., `brag update 42
-t "x"`) is deliberately deferred to a future polish spec, possibly
in STAGE-003, after real usage confirms it's needed.

Reframing the spec's role explicitly: **this is the "update an
existing entry" spec.** The editor-launch shape is the *how*; the
*what* is "close the CRUD loop so users can correct typos, add
descriptions to previously terse entries, and revise entries over
time."

Parent stage: `STAGE-002-capture-and-retrieval.md`. Project: PROJ-001.

## Goal

Ship three interconnected pieces as one coherent spec:

1. **`internal/editor` package** — `Render(Entry) []byte` +
   `Parse([]byte) (Entry, error)` + `Launch(initial []byte, edit
   EditFunc) ([]byte, error)` + a default `EditFunc` that resolves
   `$EDITOR` → `$VISUAL` → `vi` and runs it. DEC-009 pins the buffer
   format to `net/textproto` header + markdown body.
2. **`Store.Update(id int64, e Entry) (Entry, error)`** — preserves
   `id` and `created_at`, bumps `updated_at`, replaces every other
   user-editable field. Returns `ErrNotFound` (wrapped) if no row
   matches. Returns the hydrated Entry so the caller gets the bumped
   timestamps.
3. **`brag edit <id>` cobra subcommand** — fetches the entry via
   `Store.Get`, renders to the buffer, spawns `$EDITOR`, compares
   the saved bytes to the initial buffer (SHA-256), parses on
   change, calls `Store.Update`, prints `Updated.` to stderr. On
   unchanged save: prints `No changes.` to stderr and exits 0. On
   parse error: `UserErrorf` with context, exit 1. On editor
   failure: wrapped internal error, exit 2.

When this spec ships, users can fix any mistake in any captured
entry through the primary tool, without needing `sqlite3` as an
escape hatch for mutation.

## Inputs

- **Files to read:**
  - `docs/api-contract.md` — `brag edit <id>` section. **Note:**
    this spec UPDATES the doc to reflect the editor-launch as the
    update mechanism and to document exit-code semantics for the
    new command. Updates land in this spec's PR.
  - `docs/data-model.md` — `entries` column nullability (same
    concern as SPEC-006's scan path — `description/tags/project/
    type/impact` are nullable TEXT, so the Update SQL uses plain
    parameter binding and the renderer omits empty-valued rows).
  - `AGENTS.md` §8 (conventions), §9 (testing: separate buffers,
    fail-first, assertion specificity), §12 "During design".
  - `/decisions/DEC-005-integer-autoincrement-ids.md`
  - `/decisions/DEC-006-cobra-cli-framework.md`
  - `/decisions/DEC-007-required-flag-validation-in-runE.md` —
    applies to `brag edit`'s positional-arg validation.
  - `/decisions/DEC-009-editor-buffer-format.md` — the buffer
    format (this spec's companion DEC, emitted during design).
  - `/guidance/constraints.yaml`
  - `internal/cli/show.go` — reference shape for ID-taking commands
    + the existing `renderEntry` helper (note: `show`'s renderer
    produces markdown *for reading*; `editor`'s renderer produces
    *for editing* — different output shapes, no sharing).
  - `internal/cli/delete.go` — reference for the "fetch-then-mutate"
    flow (Store.Get before the user-facing action).
  - `internal/cli/errors.go` — `ErrUser` + `UserErrorf`.
  - `internal/storage/store.go` — existing `Store.Get` to read the
    current entry; new `Store.Update` lives alongside `Add` / `Get`
    / `Delete` / `List`.
  - `internal/storage/errors.go` — `ErrNotFound`.
- **External APIs:** stdlib `net/textproto`, `os/exec`, `os`,
  `crypto/sha256`. No new top-level go.mod entries.
- **Related code paths:** `internal/cli/`, `internal/storage/`,
  new `internal/editor/` package, `cmd/brag/main.go`.

## Outputs

- **Files created:**
  - `internal/editor/editor.go` — `Entry` re-declaration (or a thin
    import from `storage`; see Notes), `Render(Entry) []byte`,
    `Parse([]byte) (Entry, error)`.
    - **Design:** this package should not import `internal/storage`
      directly (to avoid a dependency cycle risk and to keep the
      format concern independent). Define a local `Fields` struct
      in `internal/editor` that mirrors the editable subset of
      `storage.Entry` (Title, Description, Tags, Project, Type,
      Impact — NOT id, created_at, updated_at). The CLI layer
      translates between `editor.Fields` and `storage.Entry`.
  - `internal/editor/launch.go` — `EditFunc` type, `Launch(initial
    []byte, edit EditFunc) ([]byte, bool, error)` (returns edited
    bytes, a "changed" bool, and error), `Default` (resolves
    `$EDITOR` → `$VISUAL` → `vi` and runs it), `resolveEditor() []
    string` unexported helper.
  - `internal/editor/editor_test.go` — render + parse + round-trip
    tests, no filesystem or subprocess.
  - `internal/editor/launch_test.go` — `Launch` with an injected
    fake `EditFunc` (pure Go, no subprocess); `resolveEditor` unit
    tests.
  - `internal/cli/edit.go` — `NewEditCmd() *cobra.Command` +
    `runEdit` RunE handler + package-level `testEditFunc` var that
    tests can set to override `editor.Default`.
  - `internal/cli/edit_test.go` — CLI tests using `testEditFunc` to
    simulate editing without spawning a real editor.
- **Files modified:**
  - `internal/storage/store.go` — add `Store.Update(id int64, e
    Entry) (Entry, error)`.
  - `internal/storage/store_test.go` — add `TestUpdate_*` tests
    (happy, not-found, preserves-id-and-created-at, bumps-updated-
    at).
  - `cmd/brag/main.go` — register the subcommand with one added
    line: `root.AddCommand(cli.NewEditCmd())`.
  - `docs/api-contract.md` — flesh out the `brag edit <id>` section
    with exit-code semantics, the buffer format pointer (to
    DEC-009), and explicit framing as the "update an entry"
    mechanism.
  - `docs/tutorial.md` — strike `brag edit <id>` from §9 "What's
    NOT there yet"; add a short "Edit an entry" mini-section under
    §4 showing the flow.
- **New exports:**
  - `editor.Fields` (struct)
  - `editor.Render(Fields) []byte`
  - `editor.Parse([]byte) (Fields, error)`
  - `editor.EditFunc` (func type: `func(path string) error`)
  - `editor.Launch(initial []byte, edit EditFunc) (edited []byte, changed bool, err error)`
  - `editor.Default` (package-level `EditFunc` that execs $EDITOR)
  - `(*storage.Store).Update(id int64, e Entry) (Entry, error)`
  - `cli.NewEditCmd() *cobra.Command`
- **Database changes:** none. Schema already supports updates; no
  new migration.

## Locked design decisions (inline — no new DEC needed beyond DEC-009)

1. **Buffer format: DEC-009.** Header + blank line + body. Empty-
   valued fields are OMITTED from render (not rendered as `Tags:`
   with nothing after). Description preserved verbatim, including
   embedded blank lines.

2. **Title is required.** If the parsed buffer has no `Title`
   header or the Title value is empty/whitespace-only, `Parse`
   returns an error. Mirrors `brag add --title` validation.

3. **Case-insensitive header keys.** Per `net/textproto` semantics.
   `TAGS:` and `tags:` both parse. Render always uses canonical
   case (`Title`, `Tags`, etc.).

4. **Unknown headers are ignored silently.** User typing `Mood:
   tired` into the buffer does not error — it's just discarded.
   (Better than failing, which would surprise users; the fields we
   know about are the ones we persist.)

5. **Change detection: SHA-256 of bytes.** If edited bytes hash
   identically to initial, abort with "No changes." — even if the
   user "saved" without modifications. Stricter than semantic-
   equivalence comparison; simpler.

6. **Editor resolution: `$EDITOR` → `$VISUAL` → `vi`.** Matches
   git's convention (actually git uses `GIT_EDITOR` → `core.editor`
   → `VISUAL` → `EDITOR` → `vi`; we adopt the simpler chain). If
   resolved editor exec fails with non-zero status AND the buffer
   was modified, treat as error. If exec fails with non-zero status
   AND the buffer was unmodified, treat as aborted → return nil
   error with changed=false. This matches how git behaves when the
   user `:cq`'s out of vim.

7. **Temp file location.** `os.CreateTemp("", "brag-edit-*.md")` —
   stdlib handles permissions and uniqueness. `.md` suffix helps
   editors pick syntax highlighting. Deleted via `defer os.Remove`.

8. **`Store.Update` semantics.**
   - Replaces every user-editable field (Title, Description, Tags,
     Project, Type, Impact).
   - Preserves `id` and `created_at`.
   - Sets `updated_at = time.Now().UTC().Truncate(time.Second)`.
   - Does NOT validate Title emptiness at storage layer — that's
     the CLI/editor layer's job (Parse catches empty Title).
   - Returns `ErrNotFound` (wrapped) if `RowsAffected() == 0`.
   - Returns the hydrated Entry (via a follow-up `Get`) so callers
     see the final state.

9. **CLI flow is Get → Render → Launch → Parse → Update.**
   Six steps, all wrapped with context on error. The first Get
   surfaces `ErrNotFound` as `UserErrorf` (exit 1); the final
   Update *also* can surface `ErrNotFound` (exit 1) for the
   between-steps-disappeared race, defensively handled.

10. **Test injection: package-level `testEditFunc` variable in
    `internal/cli`.** Simpler than threading an interface through
    cobra's context. Tests set the var in `TestMain` or per-test,
    reset on cleanup. Production leaves it nil; `runEdit` falls
    back to `editor.Default`.

## Acceptance Criteria

### Editor package

- [ ] `editor.Render(Fields{Title: "x"})` returns exactly `Title: x\n\n`
      (title header, blank line, empty body). No `Tags:`, no
      `Project:`, etc. when those fields are empty.
      *[TestRender_MinimalEntry]*
- [ ] `editor.Render` with every field populated produces a buffer
      whose first line is `Title: <title>`, header block ends with
      a blank line, body is the description verbatim.
      *[TestRender_FullEntry]*
- [ ] `editor.Parse` on a well-formed buffer returns `Fields` with
      every populated field matching the header/body content.
      *[TestParse_HappyPath]*
- [ ] `editor.Parse` on a buffer with `TAGS: foo` and `tags: bar`
      reads one (the last one wins, matching stdlib behavior) but
      does not error. Document the behavior in a test comment.
      *[TestParse_CaseInsensitiveHeaders]*
- [ ] `editor.Parse` on a buffer missing the `Title` header returns
      a non-nil error mentioning "title". *[TestParse_MissingTitle]*
- [ ] `editor.Parse` on a buffer with `Title:` (empty value) returns
      a non-nil error mentioning "title". *[TestParse_EmptyTitle]*
- [ ] `editor.Parse` on a buffer with unknown headers (e.g., `Mood:
      tired`) silently ignores them; the Fields struct comes back
      populated with only the known fields.
      *[TestParse_UnknownHeadersIgnored]*
- [ ] `editor.Parse` on a buffer with multiline body (with embedded
      blank lines) preserves the body verbatim into Description.
      *[TestParse_MultilineDescription]*
- [ ] Round-trip: `Parse(Render(f))` equals `f` for any populated
      Fields (ignoring unknown-key behavior). *[TestRoundTrip_All_Fields]*

### Editor launch

- [ ] `editor.Launch` with an injected fake `EditFunc` that
      modifies the temp file returns the new bytes with
      `changed=true`. *[TestLaunch_ChangedBufferReturnsNewBytes]*
- [ ] `editor.Launch` with an injected fake that leaves the file
      byte-identical returns the original bytes with `changed=false`.
      *[TestLaunch_UnchangedBufferReturnsFalse]*
- [ ] `editor.Launch` with an injected fake that returns an error
      returns a wrapped error. *[TestLaunch_EditFuncErrorPropagates]*
- [ ] `resolveEditor` returns `["foo"]` when `$EDITOR=foo`.
      *[TestResolveEditor_UsesEditor]*
- [ ] `resolveEditor` returns `["bar"]` when `$EDITOR` is empty and
      `$VISUAL=bar`. *[TestResolveEditor_FallsBackToVisual]*
- [ ] `resolveEditor` returns `["vi"]` when both env vars are
      empty. *[TestResolveEditor_DefaultsToVi]*
- [ ] `resolveEditor` handles editor commands with arguments
      (`$EDITOR="code --wait"` → `["code", "--wait"]`).
      *[TestResolveEditor_SplitsOnWhitespace]*

### Storage Update

- [ ] `Store.Update(id, Entry{Title: "new", ...})` on an existing
      entry writes every user-editable field. Follow-up `Store.Get`
      returns the new values.
      *[TestUpdate_ReplacesUserEditableFields]*
- [ ] `Store.Update` preserves the entry's original `ID`.
      *[TestUpdate_PreservesID]*
- [ ] `Store.Update` preserves the entry's original `CreatedAt`.
      *[TestUpdate_PreservesCreatedAt]*
- [ ] `Store.Update` sets `UpdatedAt` to a time strictly greater
      than or equal to the call time (allow 1s tolerance), and
      stored in UTC. *[TestUpdate_BumpsUpdatedAt]*
- [ ] `Store.Update` on a missing ID returns an error matching
      `errors.Is(err, storage.ErrNotFound)`.
      *[TestUpdate_NotFoundReturnsErrNotFound]*
- [ ] `Store.Update` returns the hydrated Entry (not the input).
      The returned Entry has the new `UpdatedAt`.
      *[TestUpdate_ReturnsHydratedEntry]*

### CLI edit command

- [ ] `brag edit <id>` with the test edit func that modifies the
      buffer: fetches entry, opens it in the (fake) editor, writes
      the modified content via `Store.Update`, prints `Updated.`
      to stderr, `outBuf.Len() == 0`, exits 0.
      *[TestEditCmd_HappyPath]*
- [ ] `brag edit <id>` with the test edit func that leaves the
      buffer unchanged: prints `No changes.` to stderr,
      `outBuf.Len() == 0`, exits 0, the row is byte-identical
      afterward (UpdatedAt not bumped).
      *[TestEditCmd_UnchangedBufferPrintsNoChanges]*
- [ ] `brag edit 999` when id 999 does not exist:
      `errors.Is(err, cli.ErrUser)`, `outBuf.Len() == 0`, the
      editor is NOT invoked (testEditFunc wasn't called).
      *[TestEditCmd_NotFoundIsUserError]*
- [ ] `brag edit <id>` with a test edit func that writes a buffer
      missing the Title header: `errors.Is(err, cli.ErrUser)`, the
      row is unchanged in the DB.
      *[TestEditCmd_ParseErrorIsUserError]*
- [ ] `brag edit <id>` with a test edit func that returns an error
      (editor failed): err is not `ErrUser` (maps to exit 2, internal
      error). Row unchanged. *[TestEditCmd_EditorErrorIsInternal]*
- [ ] `brag edit` with no positional arg: `ErrUser`.
      *[TestEditCmd_NoArgsIsUserError]*
- [ ] `brag edit 1 2`: `ErrUser`. *[TestEditCmd_TooManyArgsIsUserError]*
- [ ] `brag edit abc`: `ErrUser`. *[TestEditCmd_NonNumericArgIsUserError]*
- [ ] `brag edit 0`: `ErrUser`. *[TestEditCmd_NonPositiveArgIsUserError]*
- [ ] `brag edit --help` prints usage with an `Examples:` block
      and `outBuf.Len() > 0` / `errBuf.Len() == 0`.
      *[TestEditCmd_HelpShape]*
- [ ] Existing SPEC-001..008 tests remain green.
      *[manual: go test ./...]*
- [ ] `gofmt -l .` empty, `go vet ./...` clean, `CGO_ENABLED=0 go
      build ./...` succeeds, `go test ./...` green.
- [ ] `docs/api-contract.md` updated for `brag edit <id>` section.
- [ ] `docs/tutorial.md` §9 struck `brag edit <id>`; a short "Edit
      an entry" mini-section added somewhere sensible (§4 or a
      new §4.5).

## Failing Tests

Written now. Every CLI test uses separate `outBuf` / `errBuf` and
asserts on both (§9 SPEC-001 lesson). Every output assertion targets
distinctive content, not generic cobra boilerplate (§9 SPEC-005
lesson). Fail-first run before implementation (§9 SPEC-003 lesson).
Every implementation-path option in this spec's Notes passes every
blocking constraint (§12 "During design" rule from SPEC-007).

### `internal/editor/editor_test.go` (new; no subprocess, no filesystem)

Imports: `testing`, `strings`, package under test.

- **`TestRender_MinimalEntry`** — `Render(Fields{Title: "x"})`
  equals exactly `"Title: x\n\n"`. No other headers appear.
- **`TestRender_FullEntry`** — Render with all fields populated.
  Assert substring-ordered: first `"Title: <title>\n"` appears
  before `"Tags:"`, `"Tags:"` appears before blank-line, blank-line
  appears before the description text.
- **`TestRender_OmitsEmptyHeaders`** — Render with `Fields{Title:
  "x", Description: "body"}`. Assert the output does NOT contain
  `"Tags:"`, `"Project:"`, `"Type:"`, `"Impact:"`.
- **`TestParse_HappyPath`** — Parse a hand-authored buffer with all
  fields populated. Assert each struct field matches.
- **`TestParse_CaseInsensitiveHeaders`** — buffer with `TAGS: foo`.
  Parse. Assert `Fields.Tags == "foo"`.
- **`TestParse_MissingTitle`** — buffer with no `Title:` line.
  Parse returns non-nil error whose `.Error()` contains "title"
  (case-insensitive).
- **`TestParse_EmptyTitle`** — buffer with `Title:\n`. Same error
  shape.
- **`TestParse_UnknownHeadersIgnored`** — buffer has `Mood: tired`
  plus valid headers. Parse succeeds; Fields has the known fields
  populated, no panic, no error.
- **`TestParse_MultilineDescription`** — body has three paragraphs
  separated by blank lines. Parse. Assert `Fields.Description`
  equals the body verbatim including embedded blanks.
- **`TestRoundTrip_AllFields`** — for a Fields value with all
  fields populated, `Parse(Render(f))` returns a Fields equal to
  the original.

### `internal/editor/launch_test.go` (new)

Imports: `testing`, `os`, `bytes`, package under test.

- **`TestLaunch_ChangedBufferReturnsNewBytes`** — define a fake
  `EditFunc` that, given the temp file path, writes `[]byte("NEW")`
  to it. Call `Launch([]byte("OLD"), fake)`. Assert edited equals
  `[]byte("NEW")`, changed is true, err is nil.
- **`TestLaunch_UnchangedBufferReturnsFalse`** — fake EditFunc is
  a no-op (just returns nil). Call `Launch([]byte("X"), fake)`.
  Assert edited equals `[]byte("X")`, changed is false, err is
  nil.
- **`TestLaunch_EditFuncErrorPropagates`** — fake returns
  `errors.New("boom")`. Assert err is non-nil and its `.Error()`
  contains "boom".
- **`TestResolveEditor_UsesEditor`** — `t.Setenv("EDITOR", "foo")`,
  `t.Setenv("VISUAL", "")`. Assert `resolveEditor()` returns
  `[]string{"foo"}`.
- **`TestResolveEditor_FallsBackToVisual`** — `t.Setenv("EDITOR",
  "")`, `t.Setenv("VISUAL", "bar")`. Assert returns
  `[]string{"bar"}`.
- **`TestResolveEditor_DefaultsToVi`** — both empty. Returns
  `[]string{"vi"}`.
- **`TestResolveEditor_SplitsOnWhitespace`** — `t.Setenv("EDITOR",
  "code --wait")`. Returns `[]string{"code", "--wait"}`.

### `internal/storage/store_test.go` (append to existing file)

Reuse the existing `newTestStore` helper.

- **`TestUpdate_ReplacesUserEditableFields`** — `Add` an entry;
  `Update(id, Entry{Title: "new", Description: "new body", Tags:
  "x,y", Project: "p", Type: "t", Impact: "i"})`; follow-up `Get`
  returns the new values.
- **`TestUpdate_PreservesID`** — `Update` returns an entry whose
  `ID` equals the pre-update value.
- **`TestUpdate_PreservesCreatedAt`** — `Add`, remember
  `CreatedAt`, sleep 1s, `Update`, assert returned `CreatedAt`
  equals original.
- **`TestUpdate_BumpsUpdatedAt`** — `Add`, remember `UpdatedAt`,
  sleep 1s, `Update`. Assert new `UpdatedAt.After(original)` AND
  `UpdatedAt.Location().String() == "UTC"`.
- **`TestUpdate_NotFoundReturnsErrNotFound`** — fresh store.
  `Update(999, Entry{Title: "x"})`. Assert
  `errors.Is(err, storage.ErrNotFound)`.
- **`TestUpdate_ReturnsHydratedEntry`** — `Update` returns a
  non-zero Entry with the new UpdatedAt reflected.

### `internal/cli/edit_test.go` (new)

Use a new `newRootWithEdit(t, editFn editor.EditFunc) (*cobra.Command,
string)` helper that:
1. Sets `testEditFunc` to the given `editFn` (nil-safe).
2. Registers `t.Cleanup` to reset `testEditFunc = nil`.
3. Builds root + edit subcommand.
4. Returns root + temp DB path.

Every test uses separate `outBuf` / `errBuf` and asserts on both.
Any test that exercises the edit flow injects an `editor.EditFunc`
that directly manipulates the file at the given path (no subprocess
spawn).

- **`TestEditCmd_HappyPath`** — seed one entry. Inject editFn that
  reads the file, rewrites it with `Title: NEW TITLE\n\nNEW BODY\n`,
  and returns nil. Run `edit 1`. Assert:
  - err nil
  - `outBuf.Len() == 0`
  - `errBuf` contains `"Updated."`
  - Open the store, `Get(1)`, assert Title is `"NEW TITLE"` and
    Description is `"NEW BODY\n"` (or similar — tests assert on
    distinctive post-update content).

- **`TestEditCmd_UnchangedBufferPrintsNoChanges`** — seed one
  entry. Inject editFn that is a no-op (returns nil without
  modifying the file). Run `edit 1`. Assert:
  - err nil
  - `outBuf.Len() == 0`
  - `errBuf` contains `"No changes."` (distinctive literal)
  - Open the store, `Get(1)`, assert Title is unchanged AND
    UpdatedAt equals the pre-edit UpdatedAt (within 1s tolerance).

- **`TestEditCmd_NotFoundIsUserError`** — fresh store. Inject an
  editFn that, if called, sets a sentinel flag (using a closure
  capturing a local bool). Run `edit 999`. Assert:
  - `errors.Is(err, ErrUser)`
  - `outBuf.Len() == 0`
  - the sentinel flag is false (edit was NOT invoked — the Get
    happens before the editor spawn).

- **`TestEditCmd_ParseErrorIsUserError`** — seed one entry. Inject
  editFn that writes an invalid buffer (e.g., no `Title:` line).
  Assert `errors.Is(err, ErrUser)` and post-edit
  `Store.Get(1).Title` is unchanged from pre-edit value.

- **`TestEditCmd_EditorErrorIsInternal`** — seed one entry. Inject
  editFn that returns `errors.New("editor exited non-zero")`.
  Assert `err != nil`, `!errors.Is(err, ErrUser)`, row unchanged.

- **`TestEditCmd_NoArgsIsUserError`** — `edit`. `ErrUser`.
- **`TestEditCmd_TooManyArgsIsUserError`** — `edit 1 2`. `ErrUser`.
- **`TestEditCmd_NonNumericArgIsUserError`** — `edit abc`. `ErrUser`.
- **`TestEditCmd_NonPositiveArgIsUserError`** — `edit 0`. `ErrUser`.
- **`TestEditCmd_HelpShape`** — `edit --help`. Assert nil error,
  `errBuf.Len() == 0`, `outBuf` contains `"Examples:"` (distinctive
  label — SPEC-005 assertion-specificity lesson).

Notes for the implementer on testing patterns:

- Fail-first run: after writing all the new tests, run `go test
  ./...` once and confirm each fails for the expected reason
  (undefined symbol, missing substring, wrong error class). If any
  unexpectedly passes, tighten the assertion.
- Reset `testEditFunc = nil` in every test's cleanup, even if
  `newRootWithEdit` already does — defense in depth against test
  ordering.
- The package-level `testEditFunc` in `internal/cli` is a standard
  Go test-hook pattern. Document with a comment on the var that
  says "Only set by tests; production leaves nil."

## Implementation Context

*Read before starting build. Self-contained handoff.*

### Decisions that apply

- `DEC-005` — int64 IDs, positional arg parsed via `strconv.ParseInt`.
- `DEC-006` — Cobra. `NewEditCmd` mirrors `NewShowCmd` / `NewDeleteCmd`.
- `DEC-007` — RunE-validated args; no `cobra.ExactArgs`.
- `DEC-009` — Editor buffer format is `net/textproto` header +
  blank line + markdown body. Parser uses stdlib
  `net/textproto.Reader.ReadMIMEHeader`. No new deps.

### Constraints that apply

All blocking constraints were checked against every implementation
option in the Notes below. None are violated.

For `internal/editor/**`, `internal/cli/**`, `internal/storage/**`,
`cmd/brag/**`, `docs/**`:

- `no-cgo` — blocking. `os/exec` is pure Go. OK.
- `no-sql-in-cli-layer` — blocking. `edit.go` imports only
  `config`, `storage`, `editor` (+ stdlib). No SQL. `editor`
  package imports no SQL either.
- `no-new-top-level-deps-without-decision` — warning. No new deps
  introduced. DEC-009's rationale is specifically that
  `net/textproto` suffices and yaml.v3 stays out of go.mod.
- `storage-tests-use-tempdir` — blocking. All storage tests use
  `t.TempDir()`.
- `stdout-is-for-data-stderr-is-for-humans` — blocking. Edit
  produces no stdout — `"Updated."` and `"No changes."` go to
  stderr. Every happy-path CLI test asserts `outBuf.Len() == 0`.
- `errors-wrap-with-context` — warning. Every returned error is
  wrapped (`get entry`, `render buffer`, `launch editor`, `parse
  buffer`, `update entry`).
- `timestamps-in-utc-rfc3339` — blocking. `Store.Update` sets
  `updated_at` via `time.Now().UTC().Truncate(time.Second).Format(
  time.RFC3339)`. Matches `Store.Add`'s discipline from SPEC-002.
- `test-before-implementation` — blocking.
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-009-brag-edit-command-and-editor-package`.

### AGENTS.md lessons that apply

- §9 separate buffers (SPEC-001).
- §9 monotonic tie-break — not directly relevant (no ordering
  here).
- §9 fail-first (SPEC-003).
- §9 assertion specificity (SPEC-005) — help test, "Updated." /
  "No changes." assertions target distinctive literals.
- §12 "During design" (SPEC-007) — every "option" in the Notes
  below was mentally run against the blocking constraints. No
  "either is acceptable" language.
- §2 note on agents.architect/implementer template default
  (SPEC-008) — applies to this spec's frontmatter as well.

### Prior related work

- **SPEC-002** (shipped). Storage shape, scan patterns, timestamp
  discipline — `Store.Update` mirrors all of this.
- **SPEC-006** (shipped). ID-taking command shape (`show`),
  `Store.Get` fetch-before-act pattern, `ErrNotFound` mapping.
- **SPEC-008** (shipped). `delete` established the fetch-then-
  mutate flow and the stderr-for-feedback discipline. `edit`
  follows the same shape but adds the editor-launch step in the
  middle.
- **DEC-009** (new). Buffer format. Emitted during this spec's
  design.

### Out of scope (for this spec specifically)

If any of these feels necessary during build, write a new spec.

- **Flag-based update** (`brag update <id> -t "new"`). User-agreed
  follow-up after real usage. DO NOT add update-by-flag here.
- **Partial updates in `Store.Update`** (e.g., map-of-fields
  signature). Full-Entry replacement only; flag-based update will
  do Get+mutate+Update in the CLI layer when it ships.
- **Undo** or **edit history**. Out of MVP scope.
- **`brag edit` with no args** to edit the most recent entry. Out
  of scope — specific `<id>` only.
- **Batch editing** (`brag edit 1 2 3`). Single ID.
- **Editor-side syntax highlighting config**. Whatever the user's
  editor does with `.md` files is fine.
- **Locking** to prevent two `brag edit <id>` invocations in
  parallel. Single-user local tool; race is theoretical.
- **Detecting mid-edit deletion** (entry deleted from another
  process while the user is editing). `Store.Update` surfaces
  `ErrNotFound` cleanly; CLI layer maps to `UserErrorf`.

## Notes for the Implementer

- **`internal/editor/editor.go` shape.**
  ```go
  package editor

  import (
      "bufio"
      "bytes"
      "fmt"
      "io"
      "net/textproto"
      "strings"
  )

  // Fields is the user-editable subset of a brag entry. Storage and
  // CLI layers translate between this and storage.Entry. Editor
  // deliberately does NOT import storage — keeps format code
  // independent and avoids a dependency cycle risk.
  type Fields struct {
      Title       string
      Description string
      Tags        string
      Project     string
      Type        string
      Impact      string
  }

  // canonicalHeaders is the render order + canonical-case display.
  var canonicalHeaders = []struct{ key, field string }{
      {"Title", "Title"},
      {"Tags", "Tags"},
      {"Project", "Project"},
      {"Type", "Type"},
      {"Impact", "Impact"},
  }

  func Render(f Fields) []byte {
      var b bytes.Buffer
      write := func(k, v string) {
          if v != "" {
              fmt.Fprintf(&b, "%s: %s\n", k, v)
          }
      }
      write("Title", f.Title)
      write("Tags", f.Tags)
      write("Project", f.Project)
      write("Type", f.Type)
      write("Impact", f.Impact)
      b.WriteString("\n")
      b.WriteString(f.Description)
      if f.Description != "" && !strings.HasSuffix(f.Description, "\n") {
          b.WriteString("\n")
      }
      return b.Bytes()
  }

  func Parse(buf []byte) (Fields, error) {
      tp := textproto.NewReader(bufio.NewReader(bytes.NewReader(buf)))
      hdr, err := tp.ReadMIMEHeader()
      if err != nil && err != io.EOF {
          return Fields{}, fmt.Errorf("parse buffer headers: %w", err)
      }
      f := Fields{
          Title:   strings.TrimSpace(hdr.Get("Title")),
          Tags:    strings.TrimSpace(hdr.Get("Tags")),
          Project: strings.TrimSpace(hdr.Get("Project")),
          Type:    strings.TrimSpace(hdr.Get("Type")),
          Impact:  strings.TrimSpace(hdr.Get("Impact")),
      }
      if f.Title == "" {
          return Fields{}, fmt.Errorf("parse buffer: Title header is required and must be non-empty")
      }
      // Remaining body (after the blank line consumed by
      // ReadMIMEHeader) is the description.
      remaining, err := io.ReadAll(tp.R)
      if err != nil {
          return Fields{}, fmt.Errorf("parse buffer body: %w", err)
      }
      f.Description = strings.TrimPrefix(string(remaining), "\n")
      return f, nil
  }
  ```
  Unknown headers in the source come back as map entries that
  `Fields` simply doesn't read — matches the "silently ignore"
  requirement.

- **`internal/editor/launch.go` shape.**
  ```go
  package editor

  import (
      "crypto/sha256"
      "fmt"
      "os"
      "os/exec"
      "strings"
  )

  type EditFunc func(path string) error

  // Default runs $EDITOR / $VISUAL / vi on the given path. Production
  // code uses this; tests inject a fake EditFunc.
  var Default EditFunc = func(path string) error {
      argv := resolveEditor()
      argv = append(argv, path)
      c := exec.Command(argv[0], argv[1:]...)
      c.Stdin = os.Stdin
      c.Stdout = os.Stdout
      c.Stderr = os.Stderr
      if err := c.Run(); err != nil {
          return fmt.Errorf("editor exited: %w", err)
      }
      return nil
  }

  func Launch(initial []byte, edit EditFunc) ([]byte, bool, error) {
      f, err := os.CreateTemp("", "brag-edit-*.md")
      if err != nil {
          return nil, false, fmt.Errorf("create temp: %w", err)
      }
      path := f.Name()
      defer os.Remove(path)
      if _, err := f.Write(initial); err != nil {
          f.Close()
          return nil, false, fmt.Errorf("write temp: %w", err)
      }
      f.Close()
      initialHash := sha256.Sum256(initial)
      if err := edit(path); err != nil {
          return nil, false, fmt.Errorf("edit: %w", err)
      }
      edited, err := os.ReadFile(path)
      if err != nil {
          return nil, false, fmt.Errorf("read edited temp: %w", err)
      }
      editedHash := sha256.Sum256(edited)
      return edited, initialHash != editedHash, nil
  }

  func resolveEditor() []string {
      for _, env := range []string{"EDITOR", "VISUAL"} {
          if v := strings.TrimSpace(os.Getenv(env)); v != "" {
              return strings.Fields(v)
          }
      }
      return []string{"vi"}
  }
  ```
  Wait — the resolveEditor above iterates `EDITOR` then `VISUAL`;
  the tests want EDITOR first, then fall back to VISUAL. Verify
  the iteration order in the final code (EDITOR wins if both set).

- **`Store.Update` implementation.**
  ```go
  func (s *Store) Update(id int64, e Entry) (Entry, error) {
      now := time.Now().UTC().Truncate(time.Second)
      res, err := s.db.Exec(`
          UPDATE entries
          SET title = ?, description = ?, tags = ?, project = ?,
              type = ?, impact = ?, updated_at = ?
          WHERE id = ?`,
          e.Title, e.Description, e.Tags, e.Project, e.Type, e.Impact,
          now.Format(time.RFC3339), id)
      if err != nil {
          return Entry{}, fmt.Errorf("update entry %d: %w", id, err)
      }
      n, err := res.RowsAffected()
      if err != nil {
          return Entry{}, fmt.Errorf("update entry %d rows: %w", id, err)
      }
      if n == 0 {
          return Entry{}, fmt.Errorf("update entry %d: %w", id, ErrNotFound)
      }
      return s.Get(id)
  }
  ```

- **`runEdit` structure.**
  ```go
  var testEditFunc editor.EditFunc  // nil in production; tests set.

  func runEdit(cmd *cobra.Command, args []string) error {
      if len(args) != 1 {
          return UserErrorf("edit requires exactly one <id> argument")
      }
      id, err := strconv.ParseInt(args[0], 10, 64)
      if err != nil {
          return UserErrorf("invalid id %q: must be a positive integer", args[0])
      }
      if id <= 0 {
          return UserErrorf("invalid id %d: must be positive", id)
      }
      dbFlag := getFlagString(cmd, "db")
      path, err := config.ResolveDBPath(dbFlag)
      if err != nil {
          return fmt.Errorf("resolve db path: %w", err)
      }
      s, err := storage.Open(path)
      if err != nil {
          return fmt.Errorf("open store: %w", err)
      }
      defer s.Close()
      current, err := s.Get(id)
      if err != nil {
          if errors.Is(err, storage.ErrNotFound) {
              return UserErrorf("no entry with id %d", id)
          }
          return fmt.Errorf("get entry: %w", err)
      }
      initial := editor.Render(editor.Fields{
          Title: current.Title, Description: current.Description,
          Tags: current.Tags, Project: current.Project,
          Type: current.Type, Impact: current.Impact,
      })
      editFn := testEditFunc
      if editFn == nil {
          editFn = editor.Default
      }
      edited, changed, err := editor.Launch(initial, editFn)
      if err != nil {
          return fmt.Errorf("launch editor: %w", err)
      }
      if !changed {
          fmt.Fprintln(cmd.ErrOrStderr(), "No changes.")
          return nil
      }
      f, err := editor.Parse(edited)
      if err != nil {
          return UserErrorf("invalid buffer: %v", err)
      }
      _, err = s.Update(id, storage.Entry{
          Title: f.Title, Description: f.Description, Tags: f.Tags,
          Project: f.Project, Type: f.Type, Impact: f.Impact,
      })
      if err != nil {
          if errors.Is(err, storage.ErrNotFound) {
              return UserErrorf("no entry with id %d", id)
          }
          return fmt.Errorf("update entry: %w", err)
      }
      fmt.Fprintln(cmd.ErrOrStderr(), "Updated.")
      return nil
  }
  ```

- **`NewEditCmd` shape.**
  ```go
  func NewEditCmd() *cobra.Command {
      return &cobra.Command{
          Use:   "edit <id>",
          Short: "Edit a brag entry in $EDITOR",
          Long: `Open a brag entry in $EDITOR and save changes back to the database.
  This is the update mechanism for brag entries — edit any field by rewriting
  the buffer and saving. Exit the editor without modifying the buffer to abort
  (no changes written).

  Examples:
    brag edit 42        # open entry 42 in $EDITOR
    EDITOR=code brag edit 42  # override editor for one invocation`,
          RunE: runEdit,
      }
  }
  ```

- **`docs/api-contract.md` amendment.** In the `brag edit <id>`
  section, add:
  - Explicit note: "This is the primary mechanism for updating
    entries in PROJ-001. Flag-based update (e.g., `brag update <id>
    -t "new"`) is a future polish spec."
  - Exit codes: 0 on success OR no-change abort; 1 on user error
    (invalid arg, missing id, parse failure — e.g., deleted
    Title); 2 on internal error (storage failure, editor exec
    failure).
  - Pointer to DEC-009 for the buffer format.

- **`docs/tutorial.md` update.**
  1. §9 "What's NOT there yet" table: strike `brag edit <id>`.
     Remaining unshipped items are `brag add` no-args editor,
     `brag search`, export, summary, brew install.
  2. New mini-section "### Edit an entry" under §4, showing the
     round-trip briefly. Something like:
     ```
     brag edit 42          # opens $EDITOR on entry 42
     # edit the header fields or body, save and quit → Updated.
     # quit without saving / no changes → No changes.
     ```

- **No `init()` functions** (§8).

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-009-brag-edit-command-and-editor-package`
- **PR (if applicable):** opened on advance-cycle
- **All acceptance criteria met?** yes
  - `go test ./...` green (editor: 17 new tests; storage: 6 new
    Update tests; cli: 11 new edit tests; all prior specs still pass).
  - `gofmt -l .` empty, `go vet ./...` clean, `CGO_ENABLED=0 go
    build ./...` succeeds.
- **New decisions emitted:**
  - None in build. DEC-009 was emitted during design; no additional
    non-trivial choices surfaced that warranted a new DEC.
- **Deviations from spec:**
  - **Parse body handling.** The Notes reference implementation
    included `strings.TrimPrefix(string(remaining), "\n")` after
    reading the post-header body. I dropped the TrimPrefix: Go's
    `textproto.Reader.ReadMIMEHeader` already consumes the blank-
    line terminator, so no leading `\n` remains. Adding TrimPrefix
    would silently drop a legitimate first-line newline inside a
    description that happens to start with a blank line. Round-trip
    + multiline-description tests confirm the simpler form is
    correct; the spec's snippet was defensive against a non-existent
    condition.
  - **`Launch` tempfile close error.** Checked the explicit
    `f.Close()` error in `Launch` rather than the spec's bare
    `f.Close()` call — `errors-wrap-with-context` is a warning-
    severity constraint and the close error is a real failure mode
    (disk full on last flush, etc.); this keeps parity with the
    preceding `f.Write` error handling.
- **Follow-up work identified:**
  - No new backlog items. SPEC-010 (`brag add` no-args editor) is
    already queued and will reuse `editor.Render` / `editor.Parse` /
    `editor.Launch` as-is.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Very little. The Notes section included a near-complete
   implementation, including a self-flagged bug ("Wait — the
   resolveEditor above iterates EDITOR then VISUAL…") which on
   re-reading was a false alarm (the iteration order in the snippet
   is actually correct: EDITOR is tried first, VISUAL second). The
   one genuine nit was the `strings.TrimPrefix(remaining, "\n")`
   line — tracing `ReadMIMEHeader`'s behavior made it clear the
   TrimPrefix was unnecessary and potentially destructive on leading-
   blank-line descriptions, but that took a careful read rather
   than being obvious.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. Every constraint that fired (`stdout-is-for-data-stderr-is-
   for-humans`, `no-sql-in-cli-layer`, `errors-wrap-with-context`,
   `test-before-implementation`, `timestamps-in-utc-rfc3339`) was
   already in the spec's front-matter and its Implementation
   Context's "Constraints that apply" rundown. The one small thing
   worth noting for future specs: `editor.Default` is a package-
   level `var` (assignable) rather than a `func`, which is
   deliberate — it lets tests substitute editors without a new
   test-hook mechanism. Future specs that extend `editor` should
   preserve this var-based extensibility.

3. **If you did this task again, what would you do differently?**
   — One-second `time.Sleep`s in three storage tests
   (`TestUpdate_PreservesCreatedAt`, `TestUpdate_BumpsUpdatedAt`,
   `TestUpdate_ReturnsHydratedEntry`) cost ~3.3s of wall-clock in
   the test suite. The forced sleep is because `Store.Update`
   truncates to second precision via `time.Now().UTC().Truncate(time.
   Second)` (matching `Store.Add`'s SPEC-002 discipline). A future
   refactor could parameterize the clock on `Store` (or add a test-
   only seam) to make the UpdatedAt comparison deterministic
   without sleeps — but that's a cross-cutting change belonging to
   a dedicated spec, not a drive-by here. For this build, the
   sleeps were the lowest-risk way to exercise the behavior the
   acceptance criteria required.

---

## Reflection (Ship)

*Appended during the **ship** cycle.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>

---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-017
  type: story                      # epic | story | task | bug | chore
  cycle: ship
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-001
  stage: STAGE-003
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7     # usually same Claude, different session
  created_at: 2026-04-24

references:
  decisions:
    - DEC-004   # tags comma-joined TEXT — JSON ingress rejects array form, names DEC-004 in the error
    - DEC-006   # cobra framework — new `--json` flag on `brag add`
    - DEC-007   # RunE-validated flags — `--json` + field-flag mutual exclusion + all JSON-shape errors route through UserErrorf
    - DEC-011   # shared JSON output shape — SPEC-017's accepted stdin = DEC-011 shape MINUS (id, created_at, updated_at)
    - DEC-012   # EMITTED HERE — stdin-JSON schema for `brag add --json`
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
  related_specs:
    - SPEC-003   # shipped; original `brag add` flag-mode path + stdout-is-just-the-ID contract that json-mode preserves
    - SPEC-005   # shipped; add shorthand flags — the six entry-field flags json-mode is mutually exclusive with
    - SPEC-010   # shipped; editor-launch dispatch — two-branch runAdd that SPEC-017 extends to three branches
    - SPEC-014   # shipped; DEC-011 + `list --format json` — the output side of the round-trip SPEC-017 closes; its byte-identical cross-path test is SPEC-017's load-bearing-test template
    - SPEC-015   # shipped; markdown export + DEC-013 — precedent for design-time DEC emission alongside the spec
---

# SPEC-017: `brag add --json` + DEC-012 — stdin-JSON ingress and schema lock

## Context

Last spec in STAGE-003 (SPEC-013/014/015 shipped 2026-04-23/24;
SPEC-016 deferred to backlog 2026-04-23). Closes the machine-readable
I/O loop. SPEC-014 shipped the output side (`brag list --format
json|tsv`, `brag export --format json`, `internal/export.ToJSON`,
DEC-011's six-choice shape lock). SPEC-017 is the symmetric input
side: `brag add --json` reads a single JSON object from stdin,
validates it against a locked schema, and inserts via `Store.Add`.

The load-bearing contribution beyond the new flag is **the round-trip
is real**: `brag list --format json | jq '.[0]' | brag add --json`
produces a new entry whose user-owned fields match the source, with
no intermediate `jq del(.id, .created_at, .updated_at)` shape
transform. This is what DEC-011's empty-string-not-omit,
field-names-match-SQL, and naked-array choices were designed to
enable, and what DEC-012's tolerate-and-ignore-server-fields choice
lets cash out. One load-bearing test
(`TestAddCmd_JSON_RoundTripWithListJSON`) locks the full contract;
if it ever fails, DEC-011 or DEC-012 has drifted.

Parent stage: `STAGE-003-reports-and-ai-friendly-i-o.md` — Design
Notes → "Stdin-JSON schema (DEC-012 scope)" is the authoritative lock
for the six choices DEC-012 codifies; "Premise-audit hot spots →
SPEC-017" calls out the `runAdd` dispatcher status-change.
Project: PROJ-001 (MVP).

After SPEC-017 ships, STAGE-003 closes (run Prompt 1d); the remaining
project work is STAGE-004 (polish, provisional — the escape hatch
may dissolve it) and STAGE-005 (distribution).

## Goal

Ship (a) DEC-012 as a new decision file pinning the six stdin-JSON
schema choices; (b) a `--json` flag on `brag add` that reads a single
JSON object from stdin, validates against DEC-012, and inserts via
`Store.Add` with the same stdout-is-just-the-ID contract flag-mode
uses; (c) a third branch in the existing two-branch `runAdd`
dispatcher that routes to the new json-mode when `--json` is set and
is mutually exclusive with the six entry-field flags; (d) one
load-bearing round-trip test proving DEC-011 (output) and DEC-012
(input) agree.

## Inputs

- **Files to read:**
  - `/AGENTS.md` — §6 cycle rules; §7 spec anatomy; §8 DEC emission +
    honest confidence; §9 premise-audit family (SPEC-017 is
    ADDITION + STATUS-CHANGE around `runAdd`'s documented dispatch
    rule) + the markdown-heading-substring-trap addendum (not
    directly applicable — no heading assertions in this spec — but
    the general "assertion specificity" principle applies); §12 CLI
    test harness.
  - `/projects/PROJ-001-mvp/session-log.md` — the 2026-04-24 entry's
    "Pick up here next session" paragraph is the context this spec
    picks up from.
  - `/projects/PROJ-001-mvp/brief.md` — "Detail on individual ideas
    → Stdin-JSON ingress" section is authoritative for rationale
    (machine-readable input closes the AI-integration loop).
  - `/projects/PROJ-001-mvp/stages/STAGE-003-reports-and-ai-friendly-i-o.md`
    — Design Notes → "Stdin-JSON schema (DEC-012 scope)" locks all
    six DEC-012 choices; "Premise-audit hot spots → SPEC-017" names
    the dispatcher status-change.
  - `/projects/PROJ-001-mvp/backlog.md` — NOT for scope; for
    awareness of deferred siblings: "NDJSON / array-batch stdin for
    `brag add --json`" and "Lenient-accept mode for `brag add
    --json`". Do NOT pull from here for SPEC-017.
  - `/docs/api-contract.md` — current `brag add` section (lines
    31–77) has two dispatch branches documented (flag mode + editor
    mode); SPEC-017 adds a third (json mode) and rewrites the
    dispatch rule.
  - `/docs/tutorial.md` — §2 "Capture your first brag" and §3
    "Capture with full metadata" + "Capture in $EDITOR" subsection
    have the `brag add` examples; SPEC-017 adds a new "Capture from
    a script: `--json`" subsection. §4's machine-readable-output
    block already forward-references SPEC-017 (tutorial.md line 213
    mentions "piping one into `brag add --json` (SPEC-017) will
    round-trip"); that forward reference becomes present-tense.
  - `/docs/data-model.md` — gains a DEC-012 cross-reference at the
    bottom of the References list; no schema change.
  - `/guidance/constraints.yaml`
  - `/decisions/DEC-004-tags-comma-joined-for-mvp.md` — tags stay
    comma-joined string on JSON ingress; an array form is a clear
    reject naming DEC-004 in the error message.
  - `/decisions/DEC-006-cobra-cli-framework.md`
  - `/decisions/DEC-007-required-flag-validation-in-runE.md` —
    applies to both `--json` + field-flag mutual exclusion AND all
    schema-validation errors (missing title, unknown key, tags
    array, invalid JSON). All go through `UserErrorf`, never
    `MarkFlagRequired` or custom sentinels.
  - `/decisions/DEC-011-json-output-shape.md` — THE shape SPEC-017
    consumes. SPEC-017's accepted stdin = DEC-011's 9-key object
    MINUS (`id`, `created_at`, `updated_at`). Read verbatim.
  - `/decisions/DEC-012-brag-add-json-stdin-schema.md` — EMITTED BY
    THIS SPEC; the six-choice lock.
  - `/projects/PROJ-001-mvp/specs/done/SPEC-003-brag-add-command.md`
    — original `brag add` flag-mode contract; stdout-is-just-the-ID
    behavior json-mode preserves.
  - `/projects/PROJ-001-mvp/specs/done/SPEC-005-brag-add-ergonomic-polish.md`
    — shorthand-flag landing; the six entry-field flags in
    `addFieldFlags` that json-mode is mutually exclusive with.
  - `/projects/PROJ-001-mvp/specs/done/SPEC-010-brag-add-no-args-editor-launch.md`
    — the two-branch dispatch in `runAdd` that SPEC-017 extends
    to three branches. Design decision #2 ("`--db` alone does NOT
    trigger flag mode") is the pattern SPEC-017 follows — `--json`
    is NOT in `addFieldFlags`; it's a dispatch signal, not an entry
    field.
  - `/projects/PROJ-001-mvp/specs/done/SPEC-014-json-trio-and-shared-shape-dec.md`
    — ship-reflection Q3 ("write the byte-identical cross-path test
    first, not sixth") is why
    `TestAddCmd_JSON_RoundTripWithListJSON` is test #1 in this spec.
  - `/internal/cli/add.go` — existing command body; `runAdd`
    dispatcher (lines 62–70) gains a third branch. `addFieldFlags`
    (line 17) is the list of six flags json-mode is mutually
    exclusive with.
  - `/internal/cli/add_test.go` — existing flag-mode + editor-mode
    tests stay unchanged; json-mode tests appended.
  - `/internal/cli/errors.go` — `ErrUser` sentinel + `UserErrorf`;
    all SPEC-017 user errors route through these.
  - `/internal/export/json.go` — DEC-011's `entryRecord` struct +
    `ToJSON`; the output-side contract this spec closes. SPEC-017
    does NOT import from `internal/export`; the round-trip test
    feeds JSON bytes through `list --format json` naturally, which
    already uses `ToJSON`.
  - `/internal/storage/entry.go` — `Entry` struct; json-mode
    constructs one with the seven user-owned fields (title +
    description + tags + project + type + impact; `id`/`CreatedAt`/
    `UpdatedAt` set by `Store.Add`).
- **External APIs:** none. stdlib `encoding/json` only (already in
  use across the codebase).
- **Related code paths:** `internal/cli/`, `docs/`, `decisions/`.

## Outputs

- **Files created:**
  - `/decisions/DEC-012-brag-add-json-stdin-schema.md` — emitted in
    design alongside this spec. Six locked choices with honest
    confidence (0.85); rejected alternatives (lenient-accept, batch,
    tags-as-array, server-field-reject, silent-ignore-unknown-keys);
    consequences; revisit criteria; cross-refs to SPEC-017 and
    DEC-004/006/007/011.
  - `/internal/cli/add_json.go` — new file, same `package cli`.
    Holds: (a) the `addJSONInput` struct with `json` struct tags
    for the six user-owned fields; (b) a local `tagsField` type
    with a custom `UnmarshalJSON` that rejects array form naming
    DEC-004; (c) `parseAddJSON(r io.Reader) (storage.Entry,
    error)` which uses `json.Decoder` with `DisallowUnknownFields()`
    + `Decode` semantics to reject unknown keys AND reject
    trailing garbage (two calls to `Decode`; second returning
    non-EOF → trailing garbage error); (d) `runAddJSON(cmd
    *cobra.Command, _ []string) error` which opens the store,
    calls `parseAddJSON(cmd.InOrStdin())`, calls `Store.Add`,
    prints the ID to stdout.
  - `/internal/cli/add_json_test.go` — new file with 11 new tests
    (see Failing Tests). Reuses `newRootWithAdd(t)` from
    `add_test.go` (same package).
- **Files modified:**
  - `/internal/cli/add.go` — (a) declare `cmd.Flags().Bool("json",
    false, "read a single JSON entry from stdin; cannot combine
    with field flags")` after the six existing field flags, before
    `return cmd`. (b) Update the Long-description block to mention
    the new json-mode branch, mirroring the pattern used for
    flag-mode and editor-mode. (c) Rewrite `runAdd` to dispatch
    into three branches: json-mode takes priority; if `--json` is
    combined with any field flag return `UserErrorf` naming the
    first offending field flag; else fall through to the existing
    flag-mode / editor-mode logic verbatim. `addFieldFlags` stays
    as-is (six elements — `json` is deliberately NOT in this list
    because `--json` is a dispatch signal, not an entry-field flag;
    the mutual-exclusion check inspects both lists separately).
  - `/docs/api-contract.md` — `brag add` section rewrite: (a)
    dispatch rule (lines 56–60) grows to three branches; (b) new
    "STAGE-003 (JSON stdin form)" subsection added between the
    existing STAGE-002 editor-launch block and the "`brag list`"
    section, documenting `--json` usage, the accepted stdin schema
    (cross-linked to DEC-012), stdout = inserted ID contract,
    error behaviors (missing title / unknown key / tags array /
    mutual exclusion), and the round-trip example `brag list
    --format json | jq '.[0]' | brag add --json`. Cross-link
    `DEC-012` in the References list at the bottom of the file.
  - `/docs/tutorial.md` — (a) new subsection "Capture from a
    script: `--json`" added inside §3 (after "Capture in `$EDITOR`"
    subsection, before the horizontal rule that ends §3). One
    example: round-trip via `jq`. One example: minimal `echo
    '{"title":"x"}' | brag add --json`. Cross-link to DEC-012. (b)
    §4's machine-readable-output block (tutorial.md lines 209–213)
    currently reads "piping one into `brag add --json` (SPEC-017)
    will round-trip"; rewrite to present tense and drop the
    SPEC-017 parenthetical now that the feature ships.
  - `/docs/data-model.md` — append one reference bullet at the end
    for `DEC-012` (stdin-JSON schema for `brag add --json`). No
    schema change.
- **New exports:** none at the package level (`parseAddJSON` and
  `runAddJSON` are unexported — they're package-local helpers in
  `internal/cli`).
- **Database changes:** none. Pure new-entry write path that reuses
  `Store.Add` verbatim.

## Locked design decisions

Reproduced here so build / verify don't re-litigate. Each is paired
with at least one failing test below per AGENTS.md §9 (SPEC-009 ship
lesson). Decisions 1–6 are DEC-012's six sub-choices; decisions 7–9
are spec-local command-surface choices that don't rise to DEC level.

1. **Single object input only.** The decoder accepts one JSON object.
   Array input (`[{...}]`) is a user error. NDJSON / batch mode is
   deferred to backlog. *Pair:
   `TestAddCmd_JSON_ArrayInputIsUserError`.*
2. **`title` is required and non-empty.** Missing key, empty string,
   and whitespace-only title all return `UserErrorf`. Matches
   flag-mode's `strings.TrimSpace` check verbatim (SPEC-003 /
   DEC-007). *Pair: `TestAddCmd_JSON_MissingTitleIsUserError` +
   `TestAddCmd_JSON_EmptyTitleIsUserError`.*
3. **Optional user-owned fields are free-form text.**
   `description`, `tags`, `project`, `type`, `impact` accept any
   string (empty included). `tags` specifically stays a
   comma-joined string per DEC-004; an array form is a clear
   reject naming DEC-004. *Pair:
   `TestAddCmd_JSON_ValidInputInsertsEntryAndEmitsID` (happy path
   for all five) +
   `TestAddCmd_JSON_TagsAsArrayRejectedWithDEC004Reference`.*
4. **Server-owned fields are tolerated-and-ignored.** `id`,
   `created_at`, `updated_at` in the input are silently dropped;
   the new row gets a fresh ID and fresh timestamps from
   `Store.Add`. The user's values for those three fields are NOT
   stored. This is what makes the round-trip `brag list --format
   json | jq '.[0]' | brag add --json` work without `jq del`.
   *Pair: `TestAddCmd_JSON_ServerFieldsToleratedAndIgnored` +
   `TestAddCmd_JSON_RoundTripWithListJSON`.*
5. **Unknown keys are strict-rejected with the offending key
   named.** `json.Decoder.DisallowUnknownFields()` emits
   `json: unknown field "titl"`; SPEC-017 wraps via
   `UserErrorf("invalid JSON input: %v", err)` so `ErrUser`
   propagates. The rendered error contains the literal substring
   `unknown field "titl"` (with the actual typo) — strong signal
   to the user that they mistyped a key before it became a
   silently-missing entry. Lenient-accept mode is deferred to
   backlog. *Pair: `TestAddCmd_JSON_UnknownFieldNamedInError`.*
6. **Output on success is the inserted ID, one line, on stdout.**
   Matches SPEC-003's flag-mode stdout contract verbatim so
   `id=$(echo '{"title":"x"}' | brag add --json)` works identically
   to `id=$(brag add --title "x")`. Stderr is empty on success.
   *Pair: `TestAddCmd_JSON_ValidInputInsertsEntryAndEmitsID` asserts
   the stdout shape `^\d+\n$` + `errBuf.Len() == 0`.*

Spec-local command-surface decisions (no DEC needed):

7. **Mutual exclusion: `--json` + any field flag is a user error.**
   The dispatcher inspects both `--json` and `addFieldFlags` in
   `runAdd`; if `--json` is set AND any field flag from
   `addFieldFlags` is set, return `UserErrorf("--json cannot be
   combined with --%s", offender)` naming the FIRST field flag
   found (iteration order = `addFieldFlags` declaration order:
   `title, description, tags, project, type, impact`). Error
   surfaces before stdin is read — no silent stdin consumption on
   the error path. *Pair:
   `TestAddCmd_JSON_MutuallyExclusiveWithFieldFlags`.*
8. **Dispatch priority: json-mode > flag-mode > editor-mode.**
   `--json` checked first (and for mutual exclusion). If absent, the
   existing two-branch logic (any field flag → flag-mode; else →
   editor-mode) runs unchanged — SPEC-010's design decision #1
   ("no flags → editor mode") still holds when `--json` is unset.
   `--db` persistent flag stays orthogonal: `brag add --json --db
   /tmp/x.db` dispatches to json-mode (same way `brag add --db
   /tmp/x.db` with no other flags dispatches to editor-mode per
   SPEC-010 decision #2). *Pair: existing
   `TestAddCmd_NoFlagsOpensEditor` + `TestAddCmd_DbFlagAloneStillOpensEditor`
   + `TestAddCmd_SingleFieldFlagForcesFlagMode` all stay green
   unchanged as regression locks; new
   `TestAddCmd_JSON_AloneDispatchesToJSONMode` covers the new
   branch.*
9. **Invalid JSON (syntax errors, trailing garbage, EOF on empty
   input) is a user error, not an internal error.** Parser errors
   from the decoder wrap via `UserErrorf("invalid JSON input: %v",
   err)`. The stdlib decoder's error messages are legible enough to
   surface verbatim (contrast: raw storage errors would be wrapped
   with internal-error shape). *Pair:
   `TestAddCmd_JSON_InvalidJSONSyntaxIsUserError`.*

**Out of scope (by design — backlog entries exist for each):**

- `--json --batch` / NDJSON / array-stdin. Backlog: "NDJSON /
  array-batch stdin for `brag add --json`".
- `--json --lenient` / permissive unknown keys. Backlog:
  "Lenient-accept mode for `brag add --json`".
- Tags as a JSON array. Rejected with a clear error that names
  DEC-004; the backlog alternative lives inside DEC-012 itself as
  a rejected alternative, not a standalone entry (tags-as-array
  is tied to DEC-004's migration, not to a standalone feature).
- Strict timestamp parsing for `created_at` / `updated_at` — those
  fields are tolerated-and-ignored per decision 4; we never parse
  the user's values, so invalid RFC3339 there is a no-op, not an
  error.
- Validating impossible field combinations (e.g. "`type` must be
  one of X/Y/Z"). All optionals are free-form TEXT per DEC-012
  choice 3; anything goes.
- Hard-rejecting server-owned fields (`id`, `created_at`,
  `updated_at`). Decision 4 is tolerate-and-ignore, not
  hard-reject; covered explicitly as a rejected alternative in
  DEC-012.
- Structured-output on success (e.g., emit the inserted entry's
  JSON on stdout). Decision 6 locks stdout = inserted ID only;
  matches flag-mode. A future spec could add `--json-out` if a
  consumer asks.
- A new `internal/ingest` package. Covered under "Structural
  recommendation" below — parsing lives inline in
  `internal/cli/add_json.go` for this spec; promotion to a
  dedicated package is a one-`git mv` follow-up if NDJSON batch
  ever lands.

## Structural recommendation — parser placement

The task prompt asked: does parsing live inline in
`internal/cli/add.go`, or does it get a dedicated `internal/ingest`
package with `ParseJSON(r io.Reader) (storage.Entry, error)`
mirroring `internal/export.ToJSON`?

**Recommendation: inline** — in a package-local file
`internal/cli/add_json.go` (same `package cli`, separate file for
readability). Reasoning:

- `internal/export` earned its package when SPEC-014 knew two
  immediate consumers were coming (`list --format json` + `export
  --format json`) plus a third imminent (`export --format
  markdown` in SPEC-015). Today SPEC-017 has exactly one consumer.
  The only concrete "second consumer" is the backlogged NDJSON
  batch spec, and that spec has no revisit trigger fired yet.
- YAGNI per AGENTS.md §8. Extracting now trades zero-today benefit
  for a package the reader must load on every `add.go` change, and
  a new import graph edge.
- Promotion cost later is trivial: one `git mv internal/cli/add_json.go
  internal/ingest/json.go`, one import change, and upgrade
  `parseAddJSON` → exported `ingest.ParseJSON`. The batch spec will
  likely want a different reader shape (reader-of-N vs reader-of-1)
  anyway; extracting prematurely risks building the wrong seam.
- SPEC-014 ship reflection explicitly held on a similar one-data-
  point dedup question ("Implementation Context vs Notes for the
  Implementer") pending more data points. Same discipline applies
  here.

Counter-argument the build session or a later reader might make:
symmetry with `internal/export` makes the I/O loop legible at the
directory level — a reader sees `internal/export/` for the output
side and would expect `internal/ingest/` for the input side.
Acknowledged; the cost is a cross-file reference instead of a
cross-package one, and the benefit is one-fewer package. If the
batch spec ever lands, promote at that point with a one-line DEC
note rather than pre-emptively.

## Premise audit (AGENTS.md §9 — addition + status-change)

SPEC-017 is an **addition** case (new flag, new file, new DEC,
new test file) with **status-change** flavor (existing
`runAdd` dispatcher changes from two branches to three; existing
docs that describe the dispatch rule need updates). Both §9
heuristics apply.

**Addition heuristics** (SPEC-011 ship lesson — grep tracked
collections for count coupling):

- `addFieldFlags` slice (`internal/cli/add.go:17`): 6 elements
  today. SPEC-017 does NOT modify this slice — `--json` is NOT an
  entry field, it's a dispatch signal. The mutual-exclusion check
  iterates this slice; adding `--json` to it would break the
  mutual-exclusion semantics. Explicit no-op for this collection.
- `cmd/brag/main.go` `AddCommand` list: unchanged. No new
  subcommand; `--json` is a flag on the existing `brag add`.
- DEC collection: `/decisions/DEC-001..DEC-013` currently; SPEC-017
  adds `DEC-012`. No test or doc asserts the count.
- Existing `TestAdd_HelpListsAllFlags` (add_test.go:231) asserts
  the help contains `--title, --description, --tags, --project,
  --type, --impact, --db` — a lower-bound substring check. Adding
  `--json` to cobra's rendered help does NOT break this test.
  Verified 2026-04-24 by reading the test body.
- `TestAdd_HelpShowsShorthands` (add_test.go:503) asserts the help
  contains `-t, --title`-style pairs for the six field flags —
  `--json` has no shorthand so no pair to check, and the test is
  substring-based, not exhaustive.

**Status-change heuristics** (SPEC-012 ship lesson — grep feature
name across docs for all status claims, not just the primary one):

Explicit greps for the build session to run, with expected
doc-level actions:

```
grep -rn 'brag add' docs/ README.md AGENTS.md
  # → docs/api-contract.md §brag add lines 31–77 (dispatch rule
  #   lines 56–60 grows to three branches; new
  #   "STAGE-003 (JSON stdin form)" subsection added after the
  #   STAGE-002 editor-launch block)
  # → docs/tutorial.md §2/§3 `brag add` examples (§3 "Capture in
  #   $EDITOR" subsection stays; new "Capture from a script:
  #   `--json`" subsection added after it)
  # → docs/tutorial.md §4 lines 209–213 (drop the forward-
  #   reference parenthetical "(SPEC-017)"; make the round-trip
  #   claim present tense)
  # → docs/api-contract.md References list (line 256+) gets
  #   DEC-012 cross-link

grep -rn '\-\-json' docs/ README.md
  # → docs/tutorial.md line 212–213 currently has one mention in a
  #   forward-reference context; rewrite.
  # → no other hits today.

grep -rn 'brag add --json' docs/ README.md
  # → docs/tutorial.md line 213 (forward reference — rewrite to
  #   present tense).
  # → no other hits today (will grow via this spec's edits).

grep -rn 'editor mode\|flag mode' docs/ internal/
  # → docs/api-contract.md dispatch rule lines 56–60 (REWRITE: add
  #   "json mode" as a third mode, dispatch priority: json > flag
  #   > editor).
  # → internal/cli/add.go comments lines 19–23 (update the NewAddCmd
  #   doc comment — "Two modes" becomes "Three modes").

grep -rn 'DEC-011' docs/ decisions/
  # → docs/data-model.md line 146 (already present; no change).
  # → docs/api-contract.md (cross-links already present).
  # → decisions/DEC-011-json-output-shape.md (already references
  #   SPEC-017; "Validation" section says "verified when SPEC-017
  #   lands" — verify section doesn't get edited, but the ship
  #   session can note the validation is complete).
```

**Existing test audit** (addition-case; confirm nothing breaks):

- `internal/cli/add_test.go` — 25 existing tests; none reference
  `--json`. Safe to add alongside. New file `add_json_test.go` is
  cleaner than appending to `add_test.go` (~850 lines already;
  adding 11 tests would push past 1100 — a separate file keeps
  each file's tests topically grouped).
- Every existing dispatcher regression lock stays green unchanged:
  `TestAddCmd_NoFlagsOpensEditor`,
  `TestAddCmd_DbFlagAloneStillOpensEditor`,
  `TestAddCmd_SingleFieldFlagForcesFlagMode`,
  `TestAddCmd_EditorHappyPathPrintsIDToStdout`,
  `TestAddCmd_EditorUnchangedBufferAborts`,
  `TestAddCmd_EditorParseErrorIsUserError`,
  `TestAddCmd_EditorErrorIsInternal`. SPEC-010's design decisions
  1–8 are preserved byte-identically under `--json` absent.

**Symmetric action from `## Outputs`:** every grep hit above maps
to a concrete file modification in Outputs. No discoveries expected
at build time.

## Acceptance Criteria

Every criterion is testable; paired failing test name follows in
italics. Tests use separate `outBuf` / `errBuf` per §9 and
`cmd.SetIn(strings.NewReader(...))` to inject stdin.

- [ ] DEC-012 exists at `/decisions/DEC-012-brag-add-json-stdin-schema.md`
      with six locked choices, five rejected alternatives
      (lenient-accept, batch, tags-as-array, server-field-reject,
      silent-ignore-unknown-keys), honest confidence (0.85), and
      references to SPEC-017 + DEC-004/006/007/011. *[manual: `ls
      decisions/DEC-012*` returns the file; grep for "0.85" and
      "tolerated-and-ignored" in it.]*
- [ ] `brag add --json` with a valid single JSON object on stdin
      inserts one entry and prints the inserted ID on stdout
      (matching `^\d+\n$`). Stderr is empty. Entry in DB has the
      expected user-owned field values.
      *TestAddCmd_JSON_ValidInputInsertsEntryAndEmitsID*
- [ ] `brag add --json` with JSON missing `title` key returns
      `ErrUser`; stdout empty; DB unchanged.
      *TestAddCmd_JSON_MissingTitleIsUserError*
- [ ] `brag add --json` with `"title": ""` (or whitespace-only)
      returns `ErrUser`; stdout empty; DB unchanged.
      *TestAddCmd_JSON_EmptyTitleIsUserError*
- [ ] `brag add --json` with an unknown key (e.g. `"titl": "x"`)
      returns `ErrUser`; the error message contains the literal
      substring `unknown field "titl"` (offending key named); stdout
      empty; DB unchanged.
      *TestAddCmd_JSON_UnknownFieldNamedInError*
- [ ] `brag add --json` with JSON containing `id`, `created_at`,
      `updated_at` alongside valid `title` succeeds; the resulting
      DB row has a fresh ID (NOT the user's) and fresh timestamps
      (NOT the user's). Other user-owned fields are preserved.
      *TestAddCmd_JSON_ServerFieldsToleratedAndIgnored*
- [ ] `brag add --json` with `"tags": ["a", "b"]` (array form)
      returns `ErrUser`; the error message contains the literal
      substring `tags must be a comma-joined string` (naming
      DEC-004's model); stdout empty; DB unchanged.
      *TestAddCmd_JSON_TagsAsArrayRejectedWithDEC004Reference*
- [ ] `brag add --json` with array input (`[{"title":"x"}]`) or
      multiple objects returns `ErrUser`; stdout empty; DB
      unchanged. (Single object only per DEC-012 choice 1.)
      *TestAddCmd_JSON_ArrayInputIsUserError*
- [ ] `brag add --json --title x` (or any other field flag)
      returns `ErrUser`; the error message contains the literal
      substring `--json cannot be combined with --title`; stdout
      empty; DB unchanged; stdin is NOT consumed.
      *TestAddCmd_JSON_MutuallyExclusiveWithFieldFlags*
- [ ] `brag add --json` with invalid JSON syntax (e.g., `{"title":`)
      returns `ErrUser`; stdout empty; DB unchanged.
      *TestAddCmd_JSON_InvalidJSONSyntaxIsUserError*
- [ ] `brag add --json` (with no field flags) dispatches to
      json-mode even though the existing "no field flags" condition
      is otherwise true (it would route to editor-mode per
      SPEC-010 otherwise). Confirms dispatch priority decision.
      *TestAddCmd_JSON_AloneDispatchesToJSONMode*
- [ ] **Load-bearing round-trip:** seed an entry via `Store.Add`
      directly, render it through the same path `brag list
      --format json` uses (i.e., `export.ToJSON([]storage.Entry{e})`
      then take `.[0]` — equivalent to the user's `jq '.[0]'`),
      pipe those bytes into `brag add --json`, verify (a) exit 0;
      (b) stdout is a fresh ID different from the source's ID;
      (c) stderr empty; (d) DB now has two rows whose user-owned
      fields match between source and copy; (e) copy's
      `created_at` / `updated_at` / `id` are freshly assigned, not
      inherited from source. If this test ever fails, DEC-011 or
      DEC-012 has drifted. *TestAddCmd_JSON_RoundTripWithListJSON*
- [ ] `brag add --help` output contains the string `--json` and
      the help-text substring `read a single JSON entry from stdin`
      (assertion-specificity — unique to the feature, not cobra
      boilerplate). *TestAddCmd_JSON_HelpShowsJSONFlag*
- [ ] `gofmt -l .` empty; `go vet ./...` clean; `CGO_ENABLED=0 go
      build ./...` succeeds; `go test ./...` and `just test` green.
- [ ] Doc sweep: `docs/api-contract.md`, `docs/tutorial.md`, and
      `docs/data-model.md` updated per Outputs. The §4 forward
      reference to SPEC-017 becomes present-tense. DEC-012
      cross-links added in api-contract.md and data-model.md.
      *[manual greps listed under Premise audit above.]*

## Failing Tests

Written now, during **design**. Eleven tests total, all in a new
file `/internal/cli/add_json_test.go`. The load-bearing round-trip
test is written **first in the file** per SPEC-014 ship reflection
(Q3) — someone reading the file cold sees the central contract at
the top.

All tests follow §9: separate `outBuf` / `errBuf` with no-cross-
leakage asserts; assertion-specificity on error and help substrings;
every locked decision paired with at least one failing test.

### `internal/cli/add_json_test.go` (new file — 11 tests)

Reuse existing `newRootWithAdd(t)` from `add_test.go` (same
package `cli`). Stdin is injected via `root.SetIn(strings.NewReader(
...))` — cobra honors this through `cmd.InOrStdin()` in `runAddJSON`.

- **`TestAddCmd_JSON_RoundTripWithListJSON`** — THE load-bearing
  test. Written first. Seeds one entry directly via
  `storage.Store.Add` with known user-owned fields (`title:
  "round-trip source"`, `description: "d"`, `tags: "a,b"`,
  `project: "p"`, `type: "shipped"`, `impact: "i"`). Captures the
  source entry's `ID`, `CreatedAt`, `UpdatedAt`.
  Runs `brag list --format json` against the same DB into a
  buffer. Asserts the JSON parses as an array of length 1.
  Extracts element [0] as raw JSON bytes (via
  `json.RawMessage` round-trip or equivalent; alternatively,
  re-marshal via `export.ToJSON([]storage.Entry{source})` and
  strip the array brackets — either is acceptable, both produce
  the bytes `jq '.[0]'` would).
  Runs `brag add --json` with those bytes as stdin against the
  same DB. Asserts:
  - `err == nil`
  - `errBuf.Len() == 0` on the add invocation
  - `outBuf.String()` matches `^\d+\n$`
  - the new ID parses as int64 and is DIFFERENT from source's ID
  - DB now has exactly 2 entries
  - copy entry's `Title`, `Description`, `Tags`, `Project`,
    `Type`, `Impact` all equal source's values (byte-exact)
  - copy entry's `ID` ≠ source's `ID`
  - copy entry's `CreatedAt` ≠ source's `CreatedAt` (sleep 1s
    between the two inserts if needed to force a different RFC3339
    second — or accept equality if they happen to render the same
    second; assert only `ID` differs if a sub-second test-timing
    guarantee is brittle on the CI host)

  Implementation note on the `CreatedAt ≠` assertion: RFC3339 is
  second-precision. If the test runs fast enough that both inserts
  land in the same second, the timestamps will be equal. The safer
  assertion is just "copy's ID differs and user-owned fields match
  byte-exact" — drop the timestamp-inequality check if CI flakes.
  The ID-inequality assertion alone proves the server-field-fresh
  contract (ID comes from AUTOINCREMENT, not the user's input).
  Leave the timestamp-inequality check in as a sleep-guarded
  bonus assertion; if it flakes, remove it without ceremony.

- **`TestAddCmd_JSON_ValidInputInsertsEntryAndEmitsID`** — happy
  path without the round-trip scaffolding. Constructs JSON
  literal:
  `{"title":"j1","description":"body","tags":"a,b","project":"p","type":"shipped","impact":"i"}`.
  Feeds to stdin. Asserts:
  - `err == nil`
  - `errBuf.Len() == 0`
  - `outBuf.String()` matches `^\d+\n$`
  - DB has exactly 1 entry; all 6 user-owned fields match the
    input byte-exact
  Pairs decisions 3 and 6.

- **`TestAddCmd_JSON_MissingTitleIsUserError`** — stdin =
  `{"description":"orphan"}`. Asserts:
  - `err != nil`
  - `errors.Is(err, ErrUser)` true
  - `outBuf.Len() == 0`
  - DB empty (0 entries)
  Pairs decision 2.

- **`TestAddCmd_JSON_EmptyTitleIsUserError`** — stdin =
  `{"title":"","description":"d"}`. Same assertions as above.
  Pairs decision 2. (Whitespace-only title coverage folds into
  this test via a second table-driven sub-case:
  `{"title":"   "}` — same assertion stack.)

- **`TestAddCmd_JSON_UnknownFieldNamedInError`** — stdin =
  `{"title":"x","titl":"typo"}`. Asserts:
  - `err != nil`
  - `errors.Is(err, ErrUser)` true
  - `err.Error()` contains the literal substring `unknown field
    "titl"` (offending key named; the exact substring the stdlib
    decoder emits under `DisallowUnknownFields()`)
  - `outBuf.Len() == 0`
  - DB empty
  Pairs decision 5.

- **`TestAddCmd_JSON_ServerFieldsToleratedAndIgnored`** — stdin =
  `{"id":999,"title":"j2","created_at":"2001-01-01T00:00:00Z","updated_at":"2001-01-01T00:00:00Z"}`.
  Asserts:
  - `err == nil`
  - DB has exactly 1 entry; its `ID` is NOT 999 (it's the next
    AUTOINCREMENT value — typically 1 if the DB is fresh, but
    the test asserts "not 999" to avoid coupling to init state)
  - the entry's `CreatedAt.Year() == 2026` (or any year > 2001
    — proves the user's 2001 value was NOT stored), equivalently
    `!entry.CreatedAt.Equal(time.Date(2001,1,1,0,0,0,0,time.UTC))`
  - same for `UpdatedAt`
  - `Title == "j2"` (user-owned field preserved)
  Pairs decision 4.

- **`TestAddCmd_JSON_TagsAsArrayRejectedWithDEC004Reference`** —
  stdin = `{"title":"x","tags":["a","b"]}`. Asserts:
  - `err != nil`
  - `errors.Is(err, ErrUser)` true
  - `err.Error()` contains the literal substring `tags must be a
    comma-joined string` (message naming DEC-004's model; the
    exact substring emitted by the custom `tagsField.UnmarshalJSON`
    — see Notes for the Implementer)
  - `outBuf.Len() == 0`
  - DB empty
  Pairs decision 3's tags-array sub-case.

- **`TestAddCmd_JSON_ArrayInputIsUserError`** — stdin =
  `[{"title":"x"}]` (array wrapper around a valid object, as
  `jq` emits on `brag list --format json` without `.[0]`).
  Asserts:
  - `err != nil`
  - `errors.Is(err, ErrUser)` true
  - `outBuf.Len() == 0`
  - DB empty
  Pairs decision 1.

  Note: The stdlib decoder's error on an array-into-struct
  mismatch is legible (`json: cannot unmarshal array into Go
  value of type cli.addJSONInput`), so we don't assert on the
  error-message substring here — only on `ErrUser` propagation.
  Contrast with the tags-as-array test, where the custom
  UnmarshalJSON gives us a user-facing message worth asserting.

- **`TestAddCmd_JSON_MutuallyExclusiveWithFieldFlags`** — run
  `brag add --db <path> --json --title x` with stdin = `""` (empty;
  asserting stdin is NOT consumed on this path matters — if the
  parser ran, empty stdin would produce an "invalid JSON" error,
  not a mutual-exclusion error). Asserts:
  - `err != nil`
  - `errors.Is(err, ErrUser)` true
  - `err.Error()` contains the literal substring `--json cannot be
    combined with --title` (first offender from `addFieldFlags`
    iteration)
  - `outBuf.Len() == 0`
  - DB empty
  Pairs decision 7.

  Additional sub-case (same test via table-driven pattern): run
  `brag add --json --description d` and assert the error contains
  `--json cannot be combined with --description` — proves the
  "first offender" naming is accurate when `--title` is NOT the
  one present.

- **`TestAddCmd_JSON_InvalidJSONSyntaxIsUserError`** — stdin =
  `{"title":`(truncated). Asserts:
  - `err != nil`
  - `errors.Is(err, ErrUser)` true
  - `outBuf.Len() == 0`
  - DB empty
  Pairs decision 9.

- **`TestAddCmd_JSON_AloneDispatchesToJSONMode`** — run `brag add
  --json` with valid JSON on stdin AND leave `testEditFunc` nil
  (the same fail-first signal SPEC-010 used in
  `TestAddCmd_SingleFieldFlagForcesFlagMode`). If the dispatcher
  routed to editor-mode, `editor.Default` would try to spawn
  `$EDITOR` and the test would hang or fail. Asserts:
  - `err == nil`
  - `outBuf` matches `^\d+\n$` (json-mode's stdout contract, not
    editor-mode's — editor-mode also prints ID but the test's
    `testEditFunc` is nil, so if we reached editor-mode we'd get
    a different failure shape or a hang)
  - `errBuf.Len() == 0`
  - DB has exactly 1 entry with the JSON-provided title
  Pairs decision 8.

- **`TestAddCmd_JSON_HelpShowsJSONFlag`** — run `brag add --help`.
  Asserts:
  - `err == nil`
  - `errBuf.Len() == 0`
  - `outBuf.String()` contains `--json` AND the specific usage
    substring `read a single JSON entry from stdin` (per
    §9 assertion-specificity: a unique needle, not a generic
    word cobra auto-renders)

### Test count summary

11 failing tests in 1 new file (`internal/cli/add_json_test.go`).
No existing tests modified; all 25 in `add_test.go` stay green as
regression locks for the two original dispatch branches.

## Implementation Context

*Read this section and the files it points to before starting the
build cycle. It is the equivalent of a handoff document, folded into
the spec since there is no separate receiving agent.*

### Decisions that apply

- **`DEC-012`** (emitted in this spec) — six locked JSON stdin
  schema choices. Every deviation is a DEC-012 violation; stop and
  raise a question before changing a test to match.
- **`DEC-011`** — shared JSON output shape. SPEC-017 consumes
  DEC-011 via the round-trip test (the bytes from `list --format
  json` are what `add --json` accepts, minus `id`/`created_at`/
  `updated_at`). SPEC-017 does NOT import from `internal/export`
  at runtime — the test composes the two commands through the
  DB, not through direct helper calls.
- **`DEC-004`** — tags comma-joined TEXT. Array form on stdin is a
  clear reject. The custom `tagsField` UnmarshalJSON surfaces this
  error without leaking decoder jargon; the error message names
  "comma-joined string" (DEC-004's model).
- **`DEC-006`** — cobra framework. `--json` is a bool flag on
  `brag add`, declared in `NewAddCmd` alongside the six field
  flags.
- **`DEC-007`** — required-flag validation in `RunE`. Every
  SPEC-017 user error (missing title, empty title, unknown key,
  tags array, invalid JSON, array input, mutual exclusion) routes
  through `UserErrorf`. No `MarkFlagRequired`; no custom
  sentinels.

### Constraints that apply

For `internal/cli/**`, `docs/**`, `decisions/**`:

- `no-sql-in-cli-layer` — blocking. `add_json.go` works with
  `storage.Entry` + `Store.Add`; it must not import `database/sql`
  or any SQL driver.
- `stdout-is-for-data-stderr-is-for-humans` — blocking. Success
  output (inserted ID) goes to stdout. All error messages reach
  the user via `main.go`'s `fmt.Fprintf(os.Stderr, "brag: %s\n",
  ...)` wrapper — no error text on stdout. Every happy-path test
  asserts `errBuf.Len() == 0`; every error-path test asserts
  `outBuf.Len() == 0`.
- `errors-wrap-with-context` — warning. Any `Store.Add` or IO
  errors wrap with `fmt.Errorf("...: %w", err)` at the handler
  boundary.
- `test-before-implementation` — blocking. Write all 11 tests
  first, run `go test ./internal/cli -run "TestAddCmd_JSON"`,
  confirm every test fails for the expected reason (undefined
  flag `--json`, undefined `runAddJSON`, etc. — NOT a compilation
  error unrelated to the spec), THEN implement.
- `one-spec-per-pr` — blocking. Branch
  `feat/spec-017-brag-add-json-stdin-and-schema-dec`. Diff touches
  only the files in Outputs.
- `no-new-top-level-deps-without-decision` — not triggered;
  `encoding/json` is stdlib.

### AGENTS.md lessons that apply

- **§9 separate `outBuf` / `errBuf`** (SPEC-001) — every new test
  asserts both buffers.
- **§9 fail-first** (SPEC-003) — confirm each of the 11 new tests
  fails for the expected reason before implementation.
- **§9 assertion specificity** (SPEC-005) — error-message
  assertions use load-bearing literal substrings (`unknown field
  "titl"`, `tags must be a comma-joined string`, `--json cannot be
  combined with --title`, `read a single JSON entry from stdin`),
  not generic words.
- **§9 locked-decisions-need-tests** (SPEC-009) — nine locked
  decisions; each paired with at least one of the 11 tests.
- **§9 premise audit — addition case** (SPEC-011) — grepped above;
  `addFieldFlags` is the one tracked collection worth auditing,
  and the audit confirms it stays at 6 elements (`--json` is
  deliberately not in the list).
- **§9 premise audit — status-change case** (SPEC-012) — grepped
  above; doc-level actions enumerated under Outputs, not discovered
  at build time. The `runAdd` dispatcher (two branches → three)
  and the §4 tutorial forward reference are the status-change
  hotspots.
- **Load-bearing test top-billing** (SPEC-014 ship reflection Q3) —
  `TestAddCmd_JSON_RoundTripWithListJSON` is test #1 in the new
  file.

### Prior related work

- **SPEC-003** (shipped 2026-04-20) — original `brag add` flag
  mode. Stdout-is-just-the-ID contract json-mode preserves
  verbatim.
- **SPEC-005** (shipped 2026-04-20) — shorthand flags. The six
  entry-field flags in `addFieldFlags` that json-mode is mutually
  exclusive with.
- **SPEC-010** (shipped 2026-04-21) — editor-launch + two-branch
  dispatch in `runAdd`. Design decision #2 ("`--db` alone does
  NOT trigger flag mode") is the pattern SPEC-017 extends: `--json`
  is outside `addFieldFlags`, so it doesn't suppress editor-mode
  via that list; it's a separate dispatch signal checked first.
- **SPEC-014** (shipped 2026-04-23) — DEC-011 + JSON trio. The
  output-side contract SPEC-017 closes. SPEC-014's
  `TestExportCmd_FormatJSON_ByteIdenticalToListJSON` was the
  template for SPEC-017's round-trip test in structure (same DB,
  two commands, byte-comparable path) if not in mechanism (SPEC-014
  compared two output paths; SPEC-017 closes the output→input
  round-trip).
- **SPEC-015** (shipped 2026-04-24) — markdown export + DEC-013.
  Precedent for design-time DEC emission alongside the spec.

### Out of scope (for this spec specifically)

If any of these feels necessary during build, write a new spec
rather than expanding this one.

- **NDJSON / batch input**. Backlog: "NDJSON / array-batch stdin
  for `brag add --json`". Today's decoder reads exactly one JSON
  value; a second decode attempt is used to detect trailing
  garbage (error if the second `Decode` returns anything other
  than `io.EOF`).
- **Lenient-accept mode**. Backlog: "Lenient-accept mode for
  `brag add --json`". Strict-reject is the default and only mode
  in SPEC-017.
- **Tags as a JSON array**. Would require DEC-004 migration. The
  rejected alternative lives inside DEC-012 itself.
- **Hard-reject on server fields in input** (emit an error if the
  user sends `id`/`created_at`/`updated_at`). Decision 4 is
  tolerate-and-ignore; the round-trip works without `jq del`.
  Rejected alternative in DEC-012.
- **Structured output on success** (e.g. emit the inserted entry's
  JSON on stdout). Decision 6 locks stdout = inserted ID only.
  Future spec if a consumer asks.
- **`internal/ingest` package**. See "Structural recommendation"
  above; parsing lives in `internal/cli/add_json.go` for now.
  One-`git mv` promotion if batch ever lands.
- **Validating impossible field combinations** (type must be one
  of X/Y/Z, etc.). All optionals are free-form text.
- **Strict RFC3339 parsing of user-provided timestamps**. We
  ignore those fields entirely per decision 4.
- **Changes to `internal/cli/add.go`'s editor-mode or flag-mode
  branches**. SPEC-017 only adds the json-mode branch and the
  mutual-exclusion check. SPEC-010's design decisions 1–8 are
  preserved byte-identically under `--json` absent.
- **Changes to `storage.Entry`, `storage.Store`, or
  `internal/export`**. This spec is pure CLI; no storage or export
  modifications.
- **Changes to existing tests in `add_test.go`**. All 25 stay
  green as regression locks.

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities. Read after the
Implementation Context.

- **`internal/cli/add_json.go` layout.** Small file (~80–100
  lines). One input struct, one custom-unmarshal helper, one
  parser, one handler. Suggested sketch:

  ```go
  package cli

  import (
      "encoding/json"
      "errors"
      "fmt"
      "io"
      "strings"

      "github.com/jysf/bragfile000/internal/config"
      "github.com/jysf/bragfile000/internal/storage"
      "github.com/spf13/cobra"
  )

  // tagsField preserves DEC-004 at the ingress boundary: tags must
  // be a comma-joined string, not a JSON array. A dedicated type
  // lets us surface a clear error naming DEC-004's model, rather
  // than leak the stdlib decoder's "cannot unmarshal array into Go
  // struct field" jargon.
  type tagsField string

  func (t *tagsField) UnmarshalJSON(b []byte) error {
      var s string
      if err := json.Unmarshal(b, &s); err != nil {
          return fmt.Errorf("tags must be a comma-joined string, not an array (per DEC-004)")
      }
      *t = tagsField(s)
      return nil
  }

  // addJSONInput is the accepted stdin shape for `brag add --json`.
  // DEC-012: required `title`; optional user-owned text fields;
  // server-owned fields (id, created_at, updated_at) tolerated and
  // ignored via explicit fields with json.RawMessage so the decoder
  // sees them as known-and-accepted but we never store their values.
  type addJSONInput struct {
      Title       string          `json:"title"`
      Description string          `json:"description,omitempty"`
      Tags        tagsField       `json:"tags,omitempty"`
      Project     string          `json:"project,omitempty"`
      Type        string          `json:"type,omitempty"`
      Impact      string          `json:"impact,omitempty"`
      // Tolerated-and-ignored per DEC-012 choice 4. Accepting them
      // via the decoder (rather than rejecting as unknown keys)
      // lets `brag list --format json | jq '.[0]' | brag add --json`
      // round-trip without `jq del(.id, .created_at, .updated_at)`.
      ID          json.RawMessage `json:"id,omitempty"`
      CreatedAt   json.RawMessage `json:"created_at,omitempty"`
      UpdatedAt   json.RawMessage `json:"updated_at,omitempty"`
  }

  func parseAddJSON(r io.Reader) (storage.Entry, error) {
      dec := json.NewDecoder(r)
      dec.DisallowUnknownFields()
      var in addJSONInput
      if err := dec.Decode(&in); err != nil {
          return storage.Entry{}, UserErrorf("invalid JSON input: %v", err)
      }
      // Catch trailing garbage: a second Decode must hit EOF.
      var trailing json.RawMessage
      if err := dec.Decode(&trailing); err != io.EOF {
          if err == nil {
              return storage.Entry{}, UserErrorf("invalid JSON input: expected a single object, got trailing data")
          }
          return storage.Entry{}, UserErrorf("invalid JSON input: %v", err)
      }
      if strings.TrimSpace(in.Title) == "" {
          return storage.Entry{}, UserErrorf("--json input: \"title\" is required and must not be empty")
      }
      return storage.Entry{
          Title:       in.Title,
          Description: in.Description,
          Tags:        string(in.Tags),
          Project:     in.Project,
          Type:        in.Type,
          Impact:      in.Impact,
      }, nil
  }

  func runAddJSON(cmd *cobra.Command, _ []string) error {
      entry, err := parseAddJSON(cmd.InOrStdin())
      if err != nil {
          return err
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

      inserted, err := s.Add(entry)
      if err != nil {
          return fmt.Errorf("add entry: %w", err)
      }
      fmt.Fprintln(cmd.OutOrStdout(), inserted.ID)
      return nil
  }

  // Kept here to satisfy the "errors used in error return tree"
  // convention; io.EOF is compared via ==, not errors.Is.
  var _ = errors.Is
  ```

  The `errors` import / `_ = errors.Is` line is a crutch — drop it
  if the compiler doesn't need it. The sketch is illustrative;
  follow the spirit, not the byte-exact form.

- **`tagsField` unmarshaler error message.** Load-bearing
  substring: `tags must be a comma-joined string`. The test
  `TestAddCmd_JSON_TagsAsArrayRejectedWithDEC004Reference`
  asserts on that exact substring. Include ", not an array (per
  DEC-004)" for clarity; the test only asserts the prefix, not
  the full sentence, so you have room.

- **`json.Decoder.DisallowUnknownFields()` error message.** Load-
  bearing substring: `unknown field "titl"` (with the actual
  offending key). The stdlib emits
  `json: unknown field "titl"` verbatim; wrapping via
  `UserErrorf("invalid JSON input: %v", err)` produces a final
  error like
  `user error: invalid JSON input: json: unknown field "titl"`.
  The test asserts on the substring `unknown field "titl"` —
  present in the wrapped form. Do NOT rephrase — stdlib gives us
  this one for free.

- **Trailing-garbage detection.** Two `Decode` calls: first
  populates `in`, second should return `io.EOF`. If the second
  returns `nil` (i.e., decoded something), treat as trailing data
  (user error). If it returns a non-EOF error, treat as invalid
  JSON (also user error). This is what catches things like
  `{"title":"x"}{"title":"y"}` — which would otherwise silently
  accept the first object and discard the second.

- **`runAdd` dispatcher rewrite.** In `internal/cli/add.go`,
  update `runAdd` (lines 62–70 today) to the three-branch form:

  ```go
  func runAdd(cmd *cobra.Command, args []string) error {
      jsonMode, _ := cmd.Flags().GetBool("json")
      var firstFieldFlag string
      for _, name := range addFieldFlags {
          if cmd.Flags().Changed(name) {
              if firstFieldFlag == "" {
                  firstFieldFlag = name
              }
          }
      }
      if jsonMode && firstFieldFlag != "" {
          return UserErrorf("--json cannot be combined with --%s", firstFieldFlag)
      }
      if jsonMode {
          return runAddJSON(cmd, args)
      }
      if firstFieldFlag != "" {
          return runAddFlags(cmd, args)
      }
      return runAddEditor(cmd)
  }
  ```

  Note the use of `cmd.Flags().Changed("json")` vs `GetBool`: we
  only care about whether the flag was explicitly set. A bool
  default of `false` means `Changed` and `GetBool` both work here
  (`--json` is set only if the user typed it); the existing
  `Changed(name)` pattern in the dispatcher is the right shape to
  match.

- **`--json` flag declaration.** After the six existing field
  flags in `NewAddCmd`:

  ```go
  cmd.Flags().Bool("json", false, "read a single JSON entry from stdin; cannot combine with field flags")
  ```

  Load-bearing help-text substring: `read a single JSON entry from
  stdin`. Test `TestAddCmd_JSON_HelpShowsJSONFlag` asserts on this
  exact needle.

- **`NewAddCmd` Long description update.** The current Long
  (add.go lines 28–45) mentions two modes ("Flag mode" and "Editor
  mode"). Add a brief paragraph for json-mode between them, e.g.:

  ```
  JSON mode (--json set): reads a single JSON entry from stdin and
  inserts it. Mutually exclusive with field flags. See DEC-012 for
  the accepted schema.
  ```

  Add one line to the Examples block:
  ```
    echo '{"title":"shipped"}' | brag add --json
  ```

  `TestAdd_HelpShowsExamples` (add_test.go:524) is substring-
  based; adding examples doesn't break it.

- **Doc updates (execute in order).**

  1. `docs/api-contract.md` — rewrite `brag add` dispatch rule
     (lines 56–60) to three branches; add a new subsection
     "STAGE-003 (JSON stdin form)" after the editor-launch block.
     Pattern:

     ```
     **STAGE-003 (JSON stdin form):**

     ```
     echo '{"title":"shipped"}' | brag add --json
     brag list --format json | jq '.[0]' | brag add --json
     ```

     - `--json` reads a single JSON object from stdin and inserts
       it. Required: `title` (non-empty). Optional: `description`,
       `tags`, `project`, `type`, `impact` — all free-form text.
       Server-owned fields (`id`, `created_at`, `updated_at`) are
       tolerated-and-ignored if present, so `brag list --format
       json | jq '.[0]' | brag add --json` round-trips without
       `jq del`.
     - Unknown keys are strict-rejected with the offending key
       named in the error (catches typos like `"titl"` before they
       become silently-missing entries).
     - `--json` is mutually exclusive with field flags (`--title`,
       `--description`, `--tags`, `--project`, `--type`,
       `--impact`). Combining them exits 1.
     - Stdout on success: the inserted ID, one line, no prefix
       (same as flag-mode).
     - Schema lock: [DEC-012](../decisions/DEC-012-brag-add-json-stdin-schema.md).
     ```

     Update the dispatch rule (the block starting "Dispatch rule:"):

     ```
     Dispatch rule: if `--json` is set, runs in json mode (reads
     stdin). Else if any of `--title/-t`, `--description/-d`,
     `--tags/-T`, `--project/-p`, `--type/-k`, `--impact/-i` is
     set, runs in flag mode. Otherwise runs in editor mode. The
     persistent `--db` flag is a path override, not an entry
     field, so `brag add --db /tmp/x.db` still opens the editor.
     `--json` combined with any field flag exits 1 (user error).
     ```

     Append `DEC-012` to the References list at the bottom of the
     file.

  2. `docs/tutorial.md` §3 — add a new subsection "Capture from a
     script: `--json`" after the `$EDITOR` subsection, before the
     horizontal rule that ends §3:

     ```
     ### Capture from a script: `--json`

     For programmatic capture — a Claude session-end hook, an
     import script, piping from another tool — `brag add --json`
     reads a single JSON object from stdin:

     ```bash
     echo '{"title":"shipped FTS5 search"}' | brag add --json
     # prints the inserted ID on stdout, same as flag mode

     brag list --format json | jq '.[0]' | brag add --json
     # round-trips an entry (without jq del — server fields are
     # tolerated-and-ignored)
     ```

     Required: `title` (non-empty). Optional: `description`,
     `tags`, `project`, `type`, `impact` — all free-form text.
     `tags` stays a comma-joined string (per
     [DEC-004](../decisions/DEC-004-tags-comma-joined-for-mvp.md));
     array form is rejected. Unknown keys are rejected with the
     offending key named — catches `"titl"` typos before they
     become silently-missing entries. Mutually exclusive with
     flag-mode field flags. Schema locked by
     [DEC-012](../decisions/DEC-012-brag-add-json-stdin-schema.md).
     ```

  3. `docs/tutorial.md` §4 (lines 209–213) — the block currently
     reads:

     > The JSON shape is locked by [DEC-011] and is byte-identical
     > between `brag list --format json` and `brag export --format
     > json` on the same rows, so piping one into `brag add --json`
     > (SPEC-017) will round-trip without shape transforms.

     Rewrite as (drop SPEC-017 parenthetical; present tense):

     > The JSON shape is locked by [DEC-011] and is byte-identical
     > between `brag list --format json` and `brag export --format
     > json` on the same rows, so piping one into `brag add --json`
     > round-trips an entry without shape transforms (see [DEC-012]
     > for the stdin schema).

  4. `docs/data-model.md` — append to the References list at the
     bottom:

     ```
     - `DEC-012` — stdin-JSON schema for `brag add --json`:
       user-owned fields only (title required; description, tags,
       project, type, impact optional free-form text); server-owned
       fields (id, created_at, updated_at) tolerated-and-ignored;
       unknown keys strict-rejected.
     ```

- **Stdin injection in tests.** cobra's `cmd.InOrStdin()` returns
  what's been set via `cmd.SetIn(...)`; pass a `*strings.Reader`
  in tests. The `newRootWithAdd(t)` helper returns the root, not
  the `add` subcommand — `root.SetIn(...)` propagates to children
  via cobra's resolution. Sketch:

  ```go
  root, dbPath := newRootWithAdd(t)
  var outBuf, errBuf bytes.Buffer
  root.SetOut(&outBuf)
  root.SetErr(&errBuf)
  root.SetIn(strings.NewReader(`{"title":"x"}`))
  root.SetArgs([]string{"--db", dbPath, "add", "--json"})
  err := root.Execute()
  ```

- **fail-first run.** Before implementing, run:

  ```bash
  go test ./internal/cli -run "TestAddCmd_JSON"
  ```

  Expected: every one of the 11 new tests fails for the expected
  reason (undefined flag `--json`, undefined `runAddJSON`, unknown
  flag error from cobra). If any passes unexpectedly, investigate —
  a passing-before-implementation test is either a too-weak
  assertion or a pre-existing symbol accidentally overlapping
  (SPEC-003 ship lesson).

- **Don't generalize.** Resist the temptation to extract a
  parse-JSON-into-Entry helper shared with any other command. The
  only consumer is `runAddJSON`. If batch mode ever lands, that
  spec will extract the helper and decide its reader shape then
  — extracting now is speculation (see "Structural recommendation"
  above).

- **Branch:** `feat/spec-017-brag-add-json-stdin-and-schema-dec`.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-017-brag-add-json-stdin-and-schema-dec`
- **PR (if applicable):** (filled after `gh pr create`)
- **All acceptance criteria met?** yes
- **New decisions emitted (build-time):** none. DEC-012 was emitted
  during the design cycle and is consumed verbatim here.
- **Deviations from spec:** none. The literal sketches in "Notes for
  the Implementer" applied with one trivial trim — the example sketch
  carried an unused `errors`/`_ = errors.Is` crutch line that the
  spec already flagged as droppable, so it's gone. All six DEC-012
  choices, the mutual-exclusion semantics, and the dispatch priority
  match the spec verbatim.
- **Follow-up work identified:** none beyond the entries already on
  `backlog.md` (NDJSON / array-batch stdin; lenient-accept mode);
  both remain deferred and untouched by this build.

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing slowed the build. The spec's literal code sketches plus
   the round-trip-test-first ordering plus the explicit error-substring
   wordings made the build essentially a transcription job. One minor
   inconsistency to flag for verify: the "Failing Tests" section
   header text said "Eleven tests total" while the body listed twelve
   and the build prompt confirmed twelve — minor typo, no impact.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. `no-sql-in-cli-layer`, `stdout-is-for-data-stderr-is-for-humans`,
   `errors-wrap-with-context`, `test-before-implementation`,
   `one-spec-per-pr` were all listed and applied. DEC-004/006/007/011
   covered the design space; DEC-012 (design-time emission) covered
   the schema lock. Nothing surprised the build.

3. **If you did this task again, what would you do differently?**
   — One thing: I added a `time.Sleep(1 * time.Second)` to the
   round-trip test to force a different RFC3339 second between source
   and copy timestamps. The spec flagged this as a brittle bonus
   assertion that could be dropped without ceremony if CI flakes; a
   more disciplined build would have skipped the sleep entirely and
   relied on the ID-inequality assertion alone (which is what proves
   the server-field-fresh contract). Net cost is one wall-clock
   second per test run — small enough that the bonus assertion earns
   its keep, but a future-me who cares about test latency should
   remove it.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — Nothing about the Design→Build→Verify rhythm — the pattern holds.
   The one micro-tighten: when the design session knows the cleaner
   assertion, prescribe the cleaner assertion rather than offering
   "take it or leave it." The round-trip test spec flagged the 1s
   sleep as "bonus; drop without ceremony if it flakes," which
   transferred the judgment call to build, where path-of-least-
   resistance kept it. Design should have written "assert
   ID-inequality; omit timestamp-inequality (proven elsewhere)" and
   put the sleep in a Rejected-alternative note. Small lesson:
   "either-is-fine" in Notes-for-Implementer quietly off-loads
   decisions to build. When the decision is decidable at design
   time, decide it.

2. **Does any template, constraint, or decision need updating?**
   — Yes — AGENTS.md §9 addendum, not too narrow. SPEC-002 earned
   "use `id DESC` tie-break, not timestamp alone" for ordering.
   SPEC-017 earned the symmetric freshness case: assert `new.ID !=
   source.ID`, not `new.CreatedAt != source.CreatedAt`. Same
   underlying rule (RFC3339 is second-precision; `AUTOINCREMENT`
   is monotonic), two sibling applications. Unified as a one-line
   addendum under the existing SPEC-002 bullet; applied alongside
   this ship.

   The ordering and freshness cases are two faces of the same rule:
   RFC3339 is second-precision, so use the monotonic column for any
   distinctness assertion.

   No template changes. No DEC revision (DEC-012 shipped clean).
   No new constraint.

3. **Is there a follow-up spec I should write now before I forget?**
   — No. Three micro-items, none spec-worthy:
   - Drop the 1s sleep + `CreatedAt.Equal` line from
     `TestAddCmd_JSON_RoundTripWithListJSON` — one-line chore.
     **Applied in this ship commit.**
   - Add the §9 addendum to AGENTS.md — documentation chore.
     **Applied in this ship commit.**
   - Both bundled into this ship chore rather than a follow-up
     micro-PR.

   Deferred backlog entries (NDJSON `--batch`, `--lenient`,
   structured success output) have no revisit-trigger fired —
   leave them dormant. STAGE-003 closes clean.

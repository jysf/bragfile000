---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-012
  type: decision
  confidence: 0.85
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-7
  session_id: null

project:
  id: PROJ-001
repo:
  id: bragfile

created_at: 2026-04-24
supersedes: null
superseded_by: null

tags:
  - cli
  - io-contract
  - json
  - ai-consumer
  - ingress
---

# DEC-012: Stdin-JSON schema for `brag add --json`

## Decision

`brag add --json` reads a single JSON object from stdin, validates it
against a locked schema, and inserts via `Store.Add`. The schema is
the output side of DEC-011 (`brag list --format json`, `brag export
--format json`) MINUS the three server-owned fields, plus the six
choices below.

1. **Single object input only.** The decoder accepts exactly one JSON
   object. Array input (`[{...}]`), NDJSON (one object per line), and
   multiple concatenated objects all return a user error. Batch input
   is deferred to backlog.
2. **`title` is required and non-empty.** Missing key, empty string,
   and whitespace-only values all fail validation with the same
   `strings.TrimSpace` check flag-mode uses (SPEC-003).
3. **Optional user-owned fields are free-form text.** `description`,
   `tags`, `project`, `type`, `impact` accept any string (empty
   included). `tags` remains a comma-joined string per DEC-004; an
   array form is a clear reject with an error that names DEC-004's
   model.
4. **Server-owned fields are tolerated-and-ignored.** `id`,
   `created_at`, `updated_at` on input are silently dropped; the new
   row gets a fresh `ID` and fresh timestamps from `Store.Add`. This
   is the load-bearing choice that lets `brag list --format json | jq
   '.[0]' | brag add --json` round-trip without `jq del(.id,
   .created_at, .updated_at)`.
5. **Unknown keys are strict-rejected with the offending key named.**
   The stdlib decoder's `DisallowUnknownFields()` mode emits
   `json: unknown field "titl"` verbatim; SPEC-017 wraps it via
   `UserErrorf` so `ErrUser` propagates and exit code is 1. The user
   sees the mistyped key in the error — catches `titl` / `descripton`
   typos before they become silently-missing entries.
6. **Output on success is the inserted ID, one line, on stdout.**
   Matches flag-mode's SPEC-003 contract verbatim, so `id=$(echo
   '{"title":"x"}' | brag add --json)` works identically to
   `id=$(brag add --title "x")`.

## Context

STAGE-003 closes the machine-readable I/O loop that makes `brag`
useful to AI agents and pipe-composers. SPEC-014 shipped the output
side (DEC-011). SPEC-017 is the input side. Without a DEC, the six
ingress choices would scatter across the spec, the test file, and
the implementation — and the load-bearing "round-trip works" claim
would have no anchor.

Five of the six choices derive mechanically from DEC-011 + DEC-004:
the shape is DEC-011's shape (1, 2, 3), the output contract mirrors
flag-mode (6), the tags representation matches DEC-004 (3). The
single non-derived choice is #4 — tolerate-and-ignore for server
fields vs hard-reject — which is a deliberate ergonomic tradeoff
earning its DEC weight.

The round-trip test in SPEC-017
(`TestAddCmd_JSON_RoundTripWithListJSON`) is the validation signal: if
DEC-011 or DEC-012 ever drift, that test fails. One test, two DECs,
one contract.

## Alternatives Considered

- **Option A: Lenient-accept mode — ignore unknown keys silently.**
  - What it is: Drop `DisallowUnknownFields()`; unknown keys decode
    into nothing and are silently discarded.
  - Why rejected: At MVP the typo-catching win of strict-reject
    outweighs the "accepts upstream schema evolution gracefully"
    win of lenient mode. A user who types `{"titl": "x"}` today
    gets a missing-title error (their title lost silently); with
    strict-reject they get `unknown field "titl"` which surfaces
    the mistake. Lenient mode is a legitimate future want — captured
    in `backlog.md` under "Lenient-accept mode for `brag add --json`"
    with `--lenient` flag sketch. Revisit if a real pipeline (AI
    agent, another tool's export) emits schema-adjacent-but-extra
    fields AND the user accepts silent field loss as the tradeoff.

- **Option B: Batch input — NDJSON or array of objects.**
  - What it is: Accept `{...}\n{...}\n{...}` (NDJSON) or `[{...},
    {...}]` (JSON array) for bulk import.
  - Why rejected: The shape decision has more weight than it looks
    — NDJSON vs array, transactional vs best-effort, per-entry error
    reporting, partial-failure semantics. Bolting it onto SPEC-017
    would either force premature choices or balloon the spec.
    Single-object is what closes the `jq '.[0]'` round-trip for free;
    batch is a separate feature warranting its own spec. Captured in
    `backlog.md` under "NDJSON / array-batch stdin for `brag add
    --json`" with `--batch` flag sketch. Revisit when a real
    bulk-import workflow appears.

- **Option C: Tags as a JSON array.**
  - What it is: Accept `{"tags": ["auth", "perf"]}` and internally
    join with commas before storing.
  - Why rejected: Puts two different tag representations in play —
    comma-joined inside the binary per DEC-004, array-shaped on the
    ingress boundary — and either creates asymmetry with DEC-011's
    output (where tags render as a string) or forces both sides to
    migrate together. Strict symmetry with DEC-011 is the simpler
    contract: what comes out of `list --format json` is what goes
    into `add --json`. If tags ever become array-shaped, DEC-004 +
    DEC-011 + DEC-012 migrate in lockstep, not one at a time. The
    rejection here is explicitly reject-with-a-clear-message rather
    than silent-coerce, so a user who sends an array gets
    `tags must be a comma-joined string, not an array (per DEC-004)`
    and knows exactly what to change.

- **Option D: Server-field reject — hard error if input contains
  `id`, `created_at`, or `updated_at`.**
  - What it is: Treat the three server-owned fields as "not allowed
    on input" and reject the request when present.
  - Why rejected: Breaks the round-trip ergonomics DEC-011's
    field-names-match-SQL choice was designed to enable. A user
    piping `brag list --format json | jq '.[0]' | brag add --json`
    would have to remember to `jq del(.id, .created_at, .updated_at)`
    to make it work — exactly the shape-transform friction DEC-011
    tried to eliminate. Tolerate-and-ignore keeps the round-trip
    trivial; the cost (silent drop of user-provided server values)
    is not a real cost because those fields are not meaningful input
    — `id` is auto-assigned and timestamps are set by the storage
    layer. The contract is "we render nine fields and accept six" —
    legible without being restrictive at the ingress boundary.

- **Option E: Silent-ignore unknown keys without naming the
  offender.**
  - What it is: Use `DisallowUnknownFields()` but generic error
    message (`invalid JSON input` without the key name).
  - Why rejected: Loses the typo-catching win. The point of strict-
    reject is telling the user which key they mistyped. Stdlib's
    decoder names the key for free; surfacing it in the error costs
    nothing and recovers the UX value. Rejected mostly to make
    explicit that naming the offender is a deliberate choice, not
    an accidental side-effect.

- **Option F (chosen): All six choices above applied together.**
  - What it is: Single object + title required + free-form optionals
    + server fields tolerated + unknown keys strict-rejected with
    the offending key named + stdout = ID.
  - Why selected: Five of six derive from DEC-011/DEC-004/SPEC-003
    and are the only internally consistent choice. The one non-
    derived choice (#4, server-field tolerance) is explicitly for
    round-trip ergonomics; the alternative (Option D) has no
    compensating win. Strict-reject + lenient-as-backlog (#5, with
    Option A deferred) picks the better default for MVP while
    keeping the escape hatch available.

## Consequences

- **Positive:** The round-trip `brag list --format json | jq '.[0]'
  | brag add --json` works without shape transforms — DEC-011's
  field-names-match-SQL + empty-string-not-omit choices cash out at
  the ingress side via DEC-012 choices 3 + 4. One load-bearing test
  (`TestAddCmd_JSON_RoundTripWithListJSON`) locks the full
  bidirectional contract; if it ever fails, the break is in DEC-011
  or DEC-012. Typo-catching on unknown keys (choice 5) is an
  unambiguous UX improvement over silent field loss. The six-choice
  lock prevents the build session from re-litigating any of them.

- **Negative:** A future schema extension (e.g., adding a `links`
  column from the `--link` backlog item) requires coordinated
  updates across DEC-011, DEC-012, the SQL schema, the output
  helper, AND the input parser. Bounded cost; preferable to silent
  drift. The strict-reject default will occasionally surprise a
  user whose upstream tool added a schema-adjacent field; the error
  message names the key, and `--lenient` is backlogged if the pain
  is real. Rejecting tags-as-array means an AI consumer that emits
  structured tags has to join-with-comma before piping — acceptable
  MVP friction.

- **Neutral:** Choice 4 (server-field tolerance) means a user who
  sends a specific `id: 999` expecting to force the ID is silently
  ignored. A user who deliberately attempts ID manipulation is
  outside MVP scope; if this ever becomes a concern, the natural
  move is a follow-up spec and a DEC revision, not a silent
  behavior change.

## Validation

Right if:
- `brag list --format json | jq '.[0]' | brag add --json` inserts a
  new row whose user-owned fields match the source byte-for-byte,
  without `jq del(.id, .created_at, .updated_at)`. Verified by
  `TestAddCmd_JSON_RoundTripWithListJSON` (SPEC-017).
- A user who types `{"titl": "x"}` sees an error naming `titl`, not
  a silent success with a missing-title failure one layer deeper.
- A user who types `{"tags": ["a"]}` sees an error naming DEC-004's
  model (`tags must be a comma-joined string`), not a stdlib decoder
  jargon message.

Revisit if:
- A real pipeline emerges where silent loss of unknown keys is the
  right tradeoff — promote "Lenient-accept mode" from backlog.
- A real bulk-import workflow emerges — promote "NDJSON /
  array-batch stdin" from backlog.
- DEC-011's field names change (e.g., via a DEC-004 migration to
  tags-as-array). DEC-012 migrates in lockstep; the 9-key
  field-names-match-SQL principle still applies, but the tags
  representation changes on both sides.
- An AI consumer asks for structured-output on success (e.g., emit
  the inserted entry's JSON on stdout). Today stdout = ID only, per
  choice 6; revisit as a new spec rather than in-place.

Confidence: 0.85. Five of six choices are near-mechanical derivations
from DEC-011 / DEC-004 / SPEC-003 and individually sit at 0.90+. The
composite softens to 0.85 on choice 4 (server-field tolerate-and-
ignore), which is the one genuine design call — the alternative
(hard-reject) has a defensible argument that round-trip ergonomics
should not paper over user input they thought mattered. The counter
is that `id` / `created_at` / `updated_at` are not meaningful input
fields in any workflow we can name today, so "silently ignore what
the user can't meaningfully provide" is the right default. If
pressed, 0.75 on choice 4 alone is honest; 0.85 composite holds.

## References

- Related specs:
  - SPEC-017 (emits this DEC; adds `brag add --json` with a
    three-branch `runAdd` dispatcher).
  - SPEC-014 (shipped 2026-04-23; emits DEC-011 — the output shape
    SPEC-017 consumes). `TestExportCmd_FormatJSON_ByteIdenticalToListJSON`
    pattern is SPEC-017's round-trip-test template.
  - SPEC-003 (shipped 2026-04-20; flag-mode `brag add` + stdout-
    is-just-the-ID contract that choice 6 preserves).
  - SPEC-010 (shipped 2026-04-21; editor-launch + two-branch
    dispatch that SPEC-017 extends to three branches).
- Related decisions:
  - DEC-011 (shared JSON output shape) — DEC-012 consumes DEC-011's
    9-key shape minus server fields. The byte-identity between the
    two DECs is validated by SPEC-017's round-trip test.
  - DEC-004 (tags comma-joined TEXT) — choice 3's tags-as-string
    rule directly aligns; the error for array input names DEC-004.
  - DEC-006 (cobra framework) — `--json` is a bool flag on `brag
    add`.
  - DEC-007 (required-flag validation in `RunE`) — all DEC-012
    errors route through `UserErrorf`, not `MarkFlagRequired` or
    custom sentinels. Applies to: mutual exclusion with field
    flags, missing title, empty title, unknown key, tags array,
    array input, invalid JSON syntax, trailing garbage.
- Related constraints: `stdout-is-for-data-stderr-is-for-humans`
  (inserted ID to stdout; all errors via stderr through main.go's
  `brag: ...` wrapper).
- Related backlog entries:
  - "NDJSON / array-batch stdin for `brag add --json`" (deferred
    `--batch` mode — single-object ships here).
  - "Lenient-accept mode for `brag add --json`" (deferred
    `--lenient` mode — strict-reject ships here).
- Related docs:
  - `docs/api-contract.md` (SPEC-017 rewrites the `brag add`
    section and adds a STAGE-003 JSON stdin subsection).
  - `docs/tutorial.md` §3 (gets a "Capture from a script: `--json`"
    subsection) and §4 (forward-reference to SPEC-017 becomes
    present tense).
  - `docs/data-model.md` (DEC-012 cross-reference added; no schema
    change).

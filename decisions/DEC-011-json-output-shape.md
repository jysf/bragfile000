---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-011
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

created_at: 2026-04-23
supersedes: null
superseded_by: null

tags:
  - cli
  - io-contract
  - json
  - ai-consumer
---

# DEC-011: Shared JSON output shape for list, export, and stdin round-trip

## Decision

JSON output from `brag list --format json`, `brag export --format json`,
and the output-shape-minus-server-fields contract that `brag add --json`
will consume in SPEC-017 all follow a single documented shape:

1. **Top level is a naked JSON array** of entry objects. No envelope.
2. **Each entry object has exactly nine keys, in this order:** `id`,
   `title`, `description`, `tags`, `project`, `type`, `impact`,
   `created_at`, `updated_at`. Field names match the `entries` SQL
   column names verbatim.
3. **`tags` renders as a comma-joined string** (`"auth,perf"`), not a
   JSON array. Matches DEC-004; no re-normalization at the I/O
   boundary.
4. **Timestamps (`created_at`, `updated_at`) render as RFC3339
   strings**. Matches the storage layer's
   `time.Now().UTC().Format(time.RFC3339)` convention.
5. **Empty / unset fields render as `""`**, not omitted. Schema stays
   fixed-shape so AI consumers don't need key-presence case analysis.
6. **Pretty-printed with 2-space indentation** by default. No
   `--compact` flag in MVP.

## Context

STAGE-003 ships three machine-readable I/O paths that need to agree on
the on-the-wire shape:

- `brag list --format json` (SPEC-014) — for quick piping into `jq` or
  an AI consumer reading recent entries.
- `brag export --format json` (SPEC-014) — for durable dumps to a file.
- `brag add --json` (SPEC-017) — reads stdin JSON representing one
  entry, reusing the shape minus the three server-owned fields (`id`,
  `created_at`, `updated_at`).

Without a single locked shape, each consumer drifts — field names
differ between list and export, the round-trip `list | jq | add` breaks
without hand-written transforms, and AI consumers have to case-analyze
per command. The shape choices are small but load-bearing enough to
warrant a DEC rather than scattering them across three spec files.

Six choices needed locking; each is stated above.

## Alternatives Considered

- **Option A: JSON envelope (`{generated_at, count, filters, entries:
  [...]}`).**
  - What it is: A wrapper object carrying export-time metadata around
    the array.
  - Why rejected: Breaks `jq '.[]'` ergonomics (consumers would need
    `jq '.entries[]'`). Breaks the round-trip with `brag add --json`
    without a `jq .entries` intermediate step. Adds one more shape the
    stdin schema has to special-case. An envelope that carries
    provenance is a legitimate future want — captured in
    `backlog.md` under "JSON output envelope" with `--envelope` /
    `--wrap` flag sketch. Naked array is the default; envelope can
    layer on top later without breaking this shape.

- **Option B: Tags as a JSON array (`["auth","perf"]`).**
  - What it is: Normalize tags at the I/O boundary by splitting the
    comma-joined string into an array of strings.
  - Why rejected: Puts two different tag representations in play —
    comma-joined inside the binary per DEC-004, array-shaped outside —
    and the `brag add --json` round-trip would then need to normalize
    back, silently or otherwise. Keeping tags as a string in JSON
    preserves DEC-004 as the single source of truth; if tag
    normalization ever becomes warranted, it's a coordinated DEC-004
    + DEC-011 migration, not a boundary hack.

- **Option C: Per-format shapes (list has fields A/B/C, export has
  A/B/C/D/E).**
  - What it is: Let each command shape its output independently.
  - Why rejected: The entire point of DEC-011 is shape unification so
    three consumers agree. A single shape is simpler to document,
    simpler to test (one load-bearing byte-identical assertion between
    list-JSON and export-JSON proves no drift), and directly required
    by SPEC-017's round-trip contract.

- **Option D: Omit empty fields (`{"title": "x"}` instead of `{"title":
  "x", "description": "", ...}`).**
  - What it is: Marshal with `omitempty` so zero-value fields drop out.
  - Why rejected: AI consumers benefit from a fixed schema they can
    read without per-key presence checks. Empty-string is unambiguous
    (no field has `""` as a meaningful value — titles are required,
    the rest are free-form optional text). Round-tripping through
    `brag add --json` stays clean because an empty string and a
    missing key both mean "unset" at insert time. Byte size at
    personal-corpus scale is irrelevant.

- **Option E: Compact (no indentation) as default, with an opt-in
  `--pretty` flag.**
  - What it is: `json.Marshal` instead of `json.MarshalIndent`.
  - Why rejected: At personal-corpus scale, readability wins over
    bytes-on-wire. Humans open these files; AI consumers parse either
    shape trivially. `--compact` as an opt-out is captured in
    `backlog.md` under "`--compact` / non-pretty JSON output" with
    trivial implementation note; add if and only if a pipe consumer
    measurably cares.

- **Option F (chosen): All six choices above, applied together.**
  - What it is: Naked array + SQL-matching field names + string tags
    + RFC3339 timestamps + empty-string-not-omit + indent=2.
  - Why selected: Each sub-choice has either a prior DEC it aligns
    with (DEC-004 for tags, storage layer for timestamps), a
    deliberately-deferred alternative captured in `backlog.md`, or a
    load-bearing downstream consumer (SPEC-017 round-trip). The
    combined shape is the simplest thing that supports all three
    STAGE-003 I/O paths without hand-written per-consumer
    transforms.

## Consequences

- **Positive:** One shape, documented once, tested with one
  byte-identical assertion across list and export. `brag list
  --format json | jq '.[0]' | brag add --json` round-trips trivially
  once SPEC-017 lands — no `jq del` of server fields, no shape
  transform. DEC-004's tag representation remains internally
  consistent (no boundary translation). Field-names-match-SQL means a
  JSON consumer can read the data-model doc to learn the schema.

- **Negative:** Future schema extensions require coordinated updates
  across DEC-004, DEC-011, the SQL schema, and every consumer. In
  particular:
  - A future tags-as-array migration would need a paired DEC-004 +
    DEC-011 revision with the `brag add --json` stdin shape updated in
    lockstep. Cost is real but bounded.
  - Adding a new `entries` column (e.g. `links` from the backlog)
    requires extending the 9-key list here and everywhere it's
    asserted. Preferable to silent drift.
  - Distinguishing "field was unset" from "field was set to empty
    string" at the consumer side is not possible under this shape.
    MVP accepts this: every optional field is free-form text where
    empty is functionally equivalent to unset.

- **Neutral:** Naked array + pretty-print makes the output a bit
  chatty for high-volume pipe consumers. At personal-corpus scale
  (hundreds to low thousands of entries) nobody notices; at
  significantly larger scale, `--compact` and `--envelope` (both
  backlogged) layer on without breaking the default.

## Validation

Right if:
- `brag list --format json | jq '.[0]' | brag add --json` works
  without field deletion or shape transforms (verified when SPEC-017
  lands).
- A JSON consumer unmarshaling into a struct of 9 fields needs no
  per-field null / presence checks beyond Go zero-value handling.
- SPEC-014's shape-consistency test (`brag list --format json` and
  `brag export --format json` produce byte-identical output for a
  fixed `[]storage.Entry`) passes. If it ever fails, this DEC has
  been violated.

Revisit if:
- An AI consumer concretely asks for tags-as-array, unset-vs-empty
  distinction, or an envelope with provenance metadata — then the
  corresponding backlog entry gets promoted to a spec.
- DEC-004 migrates (e.g., normalized `tags` / `entry_tags` tables).
  DEC-011 migrates in lockstep; the naked-array and field-names-match-SQL
  choices still apply, but the tags representation changes.
- Pretty-print becomes a measurable bottleneck for a real consumer
  (unlikely at personal scale). Then `--compact` gets promoted from
  backlog.

Confidence: 0.85. Each sub-choice is grounded. The one I'd soften if
pressed is choice 5 (empty-string-not-omit) — pure-JSON style would
favor `omitempty`, and 0.75 feels honest for that sub-choice on its
own. Schema stability for AI consumers + round-trip cleanliness tip
it, and the other five are stronger, so 0.85 for the composite.

## References

- Related specs:
  - SPEC-014 (emits this DEC; wires `brag list --format json|tsv`
    and `brag export --format json`).
  - SPEC-017 (reads this DEC for stdin shape minus server fields;
    emits DEC-012 for the strict-reject / tolerated-server-fields
    input rules).
- Related decisions:
  - DEC-004 (tags comma-joined TEXT) — choice 3 directly aligns.
  - DEC-005 (INTEGER autoincrement IDs) — `id` renders as JSON
    number; no rounding risk at personal-corpus scale.
  - DEC-007 (required-flag validation in `RunE`) — `--format`
    validation on both `brag list` and `brag export` uses the same
    pattern; unknown values return `UserErrorf`.
- Related constraints: `stdout-is-for-data-stderr-is-for-humans`
  (JSON goes to stdout; human-facing error messages stay on stderr).
- Related backlog entries:
  - "JSON output envelope" (deferred `--envelope` / `--wrap` flag).
  - "`--compact` / non-pretty JSON output" (deferred flag).
  - "NDJSON / array-batch stdin" (SPEC-017 ships single-object only).
  - "Lenient-accept mode" (SPEC-017 ships strict-reject only).
- Related docs:
  - `docs/data-model.md` (entries schema — the 9 field names).
  - `docs/api-contract.md` (SPEC-014 adds `--format` subsections).

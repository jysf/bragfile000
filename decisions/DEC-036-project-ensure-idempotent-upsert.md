---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-036                         # stable, never reused
  type: decision
  confidence: 0.85                    # honest: an additive, idempotent primitive
                                      # composed from existing, tested storage
                                      # methods (CreateProject/GetProjectByName/
                                      # the EditLocations owner-query) reusing an
                                      # existing sentinel (ErrLocationOtherProject).
                                      # Nothing novel in the data model. Held below
                                      # 0.9 by two UX sub-choices: the two-line
                                      # stderr shape and the "project still ensured
                                      # when a cross-project location conflicts"
                                      # ordering. Both above §14's 0.7 line, so no
                                      # open question is filed.
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-005
repo:
  id: bragfile

created_at: 2026-07-10
supersedes: null
superseded_by: null

tags:
  - projects
  - cli
  - storage
  - idempotency
  - agent
  - mcp
---

# DEC-036: `brag project ensure` — idempotent project + location upsert, and `brag_add` stays free-text

## Decision

Add an explicit, idempotent primitive to register a project by name, and keep
capture (`brag add` / MCP `brag_add`) free-text. Three things are locked.

1. **Storage `EnsureProject(name string) (Project, bool, error)`** — the
   idempotent upsert. The bool is `created` (true = a row was inserted, false =
   the name already existed). It reads first (`GetProjectByName`); on
   `ErrNotFound` it `CreateProject`s with status `active`; and it **never**
   surfaces `ErrProjectExists` (a defensive dup branch re-fetches and returns
   `created=false`). The no-op path returns the existing row verbatim and does
   **not** bump `updated_at` — ensuring an existing project changed nothing.

2. **Storage `EnsureLocation(projectID int64, path string) (bool, error)`** —
   the idempotent location attach. The bool is `attached` (true = inserted,
   false = already attached to this project). It queries the path's current
   owner (the `EditLocations` pattern): free path → insert, `(true, nil)`; owned
   by this project → no-op, `(false, nil)`; owned by a **different** project →
   `(false, ErrLocationOtherProject)` (reusing the existing sentinel — paths are
   globally UNIQUE). A project may hold **multiple** locations; this guards only
   the same path, never the count. `AddLocation` is left unchanged (it returns
   `ErrLocationExists` for both same- and other-project dups and so cannot
   express the idempotent no-op — which is exactly why `EnsureLocation` exists).

3. **CLI `brag project ensure <name> [--location PATH]`** — calls
   `EnsureProject` then, when `--location` is non-empty, `EnsureLocation`. The
   name is required, trimmed, non-empty, and **≤ 64 characters** (matching the
   cap `brag add --json` / MCP enforce on the `project` field). `--location`
   defaults to `""` (**unset**) and is never defaulted to cwd — ensure must
   never register a surprise directory. Output mirrors `brag project new`:
   stdout stays empty, a human confirmation goes to **stderr**, exit 0 on both
   create and no-op. Two composable stderr lines: a project line (`Created
   project %q.` / `Project %q already exists.`) and, only with `--location`, a
   location line (`Attached location %q.` / `Location %q already attached.`). A
   cross-project location becomes a `UserError` naming the path.

And one boundary is locked:

4. **`brag add` / `brag_add` stay free-text — no silent auto-registration of
   unknown projects on add.** `ensure` is the explicit, safe primitive.
   Auto-registering unknown `project` values at add time would pollute the
   `projects` table with every typo. Whether the MCP server should expose an
   `ensure`-equivalent tool is a candidate follow-up, out of scope here.

## Context

Entries store `project` as free text with no referential check (DEC-017's soft
string match), and MCP `brag_add` — unlike the CLI `brag add` (SPEC-032) — does
**not** auto-fill `project` from cwd. So an agent logging over MCP can write
`project: "standup"` for a project that was never registered, producing an
orphan string that `brag project list` / `brag project status` (which read the
`projects` table) never see. STAGE-015's downstream consumer (the `standup`
portfolio tracker) reads `brag list --format json` + `brag project list --format
json` and maps entries → repos; it silently drops orphan project strings.

STAGE-015 makes `brag mcp serve` first-class. Closing the referential half of
that gap needs a registration primitive an agent can call **safely and
repeatedly** — before a capture, or in a setup script — without a
"project already exists" failure on the second run. `brag project new` errors on
a duplicate name (it is a create), so it cannot be that primitive. Hence an
idempotent `ensure`.

Two soft-link facts fall out of DEC-017 that any entries→repo mapper must
handle, and this DEC commits to **documenting** them where the project commands
live (`docs/api-contract.md`): (a) `brag project list` locations are
**authoritative-but-incomplete** — entries may reference names never registered;
(b) a project may have **multiple** on-disk `locations`.

## Alternatives Considered

- **Option A: make `brag project new` idempotent (swallow `ErrProjectExists`).**
  - Why rejected: overloads a create verb with no-op semantics and changes an
    established command's error contract (`TestProjectNew_DuplicateNameErrUser`
    asserts the failure). `new` and `ensure` are honestly different verbs — one
    asserts novelty, one asserts presence. A separate `ensure` keeps `new`'s
    contract intact and reads correctly at the call site.

- **Option B: auto-register (or warn-and-register) unknown projects on
  `brag add` / `brag_add`.**
  - Why rejected: silent auto-registration pollutes the `projects` table with
    every typo and one-off string, permanently and invisibly. It also makes
    capture non-idempotent in a surprising way (a mistyped project becomes a
    real registered project). Explicit `ensure` gives the same reachability with
    none of the pollution; capture stays a pure append (DEC-017 ethos).

- **Option C: fold location idempotency into the CLI by iterating
  `ListProjects` (the `runProjectNew` pre-check pattern).**
  - Why rejected: it re-hydrates every project in the CLI to answer a
    single-path ownership question and leaves the same-vs-different-project
    decision as untyped CLI logic. A storage method with a typed error
    (`ErrLocationOtherProject`) is cleaner to test at the storage layer, reuses
    `EditLocations`' owner-query, and keeps the CLI thin (`no-sql-in-cli-layer`).

- **Option D: a single `EnsureProject(name, location string)` that does both.**
  - Why rejected: conflates two independent idempotency questions (does the
    project exist? is the path attached?) into one signature and one return, and
    forces a location argument on the common name-only call. Two small methods
    compose cleanly and test independently; the CLI orchestrates them.

- **Option E (chosen): `EnsureProject` + `EnsureLocation` storage primitives,
  a thin `brag project ensure` CLI verb, `brag_add` unchanged.**
  - Why selected: additive, idempotent, composed from existing tested methods,
    reuses an existing sentinel, keeps all SQL in storage, mirrors the
    `runProjectNew` output convention, and gives agents a re-runnable
    registration primitive without touching capture.

## Consequences

- **Positive:** agents (and humans) get an idempotent, scriptable
  `brag project ensure` — safe before every capture, no duplicate rows, no
  duplicate locations. Closes the referential half of the STAGE-015
  unregistered-project gap; `standup` can register the names it sees.
- **Positive:** capture stays a pure free-text append (DEC-017 preserved); the
  `projects` table is only ever grown by an explicit act.
- **Positive:** `EnsureLocation` surfaces the multi-location model and the
  global-path-uniqueness guarantee through one typed error, tested at storage.
- **Neutral:** two new storage methods on `Store`; no schema change, no
  migration. The `project` subcommand set grows by one (no test asserts the
  exact set — premise-audit confirmed).
- **Negative / accepted:** on a cross-project location conflict, the project is
  still (idempotently) ensured before the location step errors — a
  `brag project ensure X --location P` where `P` belongs to another project
  leaves `X` registered without `P`. This is consistent with ensure's idempotent
  contract (creating `X` is the intended, re-runnable effect) and the re-run
  after fixing the path is safe; it is documented rather than rolled back.
- **Negative / accepted:** the ≤64-char name check on `ensure` is stricter than
  `brag project new` (which has no length check today) — a small, documented
  asymmetry, left to a future tidy spec rather than retrofitted here.

## Validation

Right if:
- `EnsureProject` on an absent name creates an `active` project and returns
  `created=true`; on a present name returns the same row, `created=false`, no
  error, no duplicate, `updated_at` unchanged.
- `EnsureLocation` attaches a free path (`true`), no-ops a same-project path
  (`false`, nil), and returns `ErrLocationOtherProject` for a path owned
  elsewhere.
- `brag project ensure <name>` and its `--location` form are byte-for-byte
  re-runnable (exit 0, no error, the "already exists"/"already attached"
  messages, no duplicate rows), with stdout empty and confirmations on stderr.
- `docs/api-contract.md` documents the command and the two soft-link facts.

Revisit if:
- A concrete need appears to register projects over MCP (add a
  `brag_project_ensure` tool with its own small spec).
- Real usage wants `brag add` to nudge/register unknown projects — revisit the
  free-text boundary deliberately, with a DEC, rather than by drift.
- The two-line stderr shape proves awkward to consume — collapse to one line
  (a display-only change; no data contract moves, since stdout is empty).

Confidence: **0.85** — an additive, idempotent primitive built from existing
tested methods and an existing sentinel; nothing novel in the data model. Held
below 0.9 only by two UX sub-choices (the two-line stderr shape; the
"project still ensured on location conflict" ordering), both above §14's 0.7
line, so no open question is filed.

## References

- Related specs:
  - **SPEC-057** (emits this DEC; adds `EnsureProject`/`EnsureLocation`, the
    `brag project ensure` CLI, and the soft-link docs).
  - SPEC-027 (shipped) — `0004_add_projects.sql` + `CreateProject`/`AddLocation`
    this composes.
  - SPEC-028/SPEC-029 (shipped) — the project CRUD CLI whose output convention
    and pre-check pattern `ensure` mirrors.
  - SPEC-031 (shipped) — `brag project here`; the global-path-uniqueness
    guarantee `EnsureLocation`'s cross-project error protects.
  - SPEC-032 (shipped) — `brag add` cwd `--project` auto-fill; the asymmetry
    (MCP `brag_add` does NOT auto-fill) that opens the gap this closes.
  - SPEC-055 / SPEC-058 (STAGE-015 siblings) — `brag mcp install`; the
    agent-facing MCP docs (the full tool contract, out of scope here).
- Related decisions:
  - DEC-017 — `entries.project` ↔ `projects` soft string match; why an
    unregistered string is invisible and why registration is purely additive.
  - DEC-018 — project delete blast radius; ensure is the additive counterpart
    (neither touches entries).
  - DEC-024 — MCP server transport/provenance; the MCP `brag_add` path whose
    lack of cwd auto-fill motivates an explicit registration primitive.
- Related constraints: `no-sql-in-cli-layer`,
  `stdout-is-for-data-stderr-is-for-humans`, `errors-wrap-with-context`,
  `timestamps-in-utc-rfc3339`, `test-before-implementation`,
  `storage-tests-use-tempdir`.
- Related docs: `docs/api-contract.md` (the `brag project ensure` section + the
  two soft-link facts), `docs/tutorial.md` (optional one-line pointer),
  `docs/data-model.md` (the projects/project_locations tables).

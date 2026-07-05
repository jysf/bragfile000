---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-043
  type: story
  cycle: build                     # design + build delivered in one PR (mirrors SPEC-040/041)
  blocked: false
  priority: high
  complexity: S

project:
  id: PROJ-003
  stage: STAGE-009
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # same Claude, different session (claude-only variant)
  created_at: 2026-07-05

references:
  decisions: [DEC-024, DEC-015]
  constraints: [no-sql-in-cli-layer, stdout-is-for-data-stderr-is-for-humans, storage-tests-use-tempdir, one-spec-per-pr]
  related_specs: [SPEC-040, SPEC-041]
---

# SPEC-043: `brag list --author` — provenance-authorship filter

## Context

Cross-project retrospective (PR #64) action-register item **P2** — the
highest-leverage finding — observed that PROJ-003's agent-native thesis is
**unmeasurable**: DEC-024 reserved the `agent:`/`model:` provenance namespace,
but the corpus can't yet answer "how much of this was agent-authored?"

**SPEC-040 already shipped the write half** — the MCP `brag_add` tool stamps
`agent:<id>`/`model:<id>` reserved tags (`internal/mcpserver/{server,provenance}.go`,
with a `clientInfo.Name` fallback), and `reservedTag`/`stampProvenance`
normalization is tested. P2's acceptance bullets 1–3 (MCP stamps provenance;
CLI unchanged; stamping+normalization tested) were therefore satisfied on
`main` before this spec. What was **missing is the read half**: P2's 4th
bullet — *"a corpus query can distinguish agent-authored from human-authored
entries (feeds P3)."* This spec ships that query as a thin filter on the
existing `brag list`, making the thesis measurable and seeding the P3
dogfooding-coverage query.

Parent stage: [STAGE-009](../stages/STAGE-009-mcp-plugin-and-capture-delight.md)
(agent-native spine). Project: PROJ-003.

## Goal

Add `brag list --author agent|human`: `agent` selects entries carrying a
reserved `agent:`/`model:` provenance tag (DEC-024), `human` selects the
complement, and omitting the flag is unchanged (all entries). No schema
change; the human `brag add` path stays byte-parity.

## Inputs

- **Files to read:** `internal/storage/store.go` (the `List` filter builder),
  `internal/storage/entry.go` (`ListFilter`), `internal/cli/list.go` (the
  cobra command + flag wiring), `internal/mcpserver/provenance.go` (the
  `agent:`/`model:` namespace this classifies).
- **Related code paths:** `internal/storage/`, `internal/cli/`.

## Outputs

- **Files modified:**
  - `internal/storage/entry.go` — add `Author string` to `ListFilter`.
  - `internal/storage/store.go` — `authorAgent`/`authorHuman` consts, the
    `provenanceExistsClause` SQL predicate, and the AND-composed WHERE branch
    (invalid value → error).
  - `internal/cli/list.go` — `--author` flag, `agent|human` validation
    (`ErrUser` on anything else), Long/Examples update.
  - `docs/api-contract.md` — `brag list` flag surface documents `--author`.
- **New exports:** none (all additions are package-internal).
- **Database changes:** none (migration-free; classifies existing tags).

## Acceptance Criteria

- [x] An entry carrying `agent:*` or `model:*` is returned by
      `brag list --author agent` and excluded by `--author human`; the
      complement holds for entries carrying neither.
- [x] Classification is prefix-anchored: a topic tag like `agentic` or
      `modeling` (no colon) is **not** provenance → classified `human`.
- [x] `--author` composes (AND) with `--tag`/`--project`/`--type`/`--since`
      and respects `--limit`.
- [x] Omitting `--author` returns all entries — default `brag list`
      output is byte-for-byte unchanged (CLI byte-parity; no provenance
      injected on the human `brag add` path).
- [x] `--author <anything-else>` is a user error (exit 1, message on stderr,
      empty stdout); the storage layer also rejects an invalid `Author`.
- [x] `brag list --author agent --format json | jq length` counts
      agent-authored entries (the measurability payoff; feeds P3).

## Failing Tests

Written during **design**, BEFORE build.

- **`internal/storage/store_test.go`**
  - `TestList_FilterByAuthor` — seeds two provenance entries (one
    `agent:`-only, one `agent:`+`model:`+topic), two plain-human, and one
    false-positive (`agentic,modeling`, no colon). Asserts `Author:"agent"`
    → the 2 provenance rows; `Author:"human"` → the 3 non-provenance rows
    (incl. the false-positive); unset → all 5; `Author:"agent"`+`Tag:"perf"`
    → the single row with both; `Author:"agent"`+`Limit:1` → 1; and
    `Author:"bogus"` → error.
  - `TestList_AuthorComposesWithOtherFilters` — backs the AC3 claim in full:
    `Author` AND-composes with `Project`, `Type`, and `Since` (beyond the
    `Tag`/`Limit` cases above).
- **`internal/mcpserver/server_test.go`**
  - `TestServer_ProvenanceRoundTripToListAuthor` — the cross-package drift
    guard: an entry written via the MCP `brag_add` tool (which stamps
    `agent:`/`model:`) is found by `List{Author:"agent"}` and excluded from
    `List{Author:"human"}`, pinning the stamp literal (mcpserver) to the
    classifier literal (storage), which share no constant.
- **`internal/cli/list_test.go`**
  - `TestListCmd_FilterByAuthor` — `--author agent` shows the agent title,
    not the human title (empty stderr); `--author human` the inverse.
  - `TestListCmd_InvalidAuthorIsUserError` — `--author robot` →
    `errors.Is(err, ErrUser)`, empty stdout.
  - `TestListCmd_HelpShowsFilters` — help output contains `--author`.

## Implementation Context

### Decisions that apply

- `DEC-024` — reserves the `agent:<name>`/`model:<id>` provenance namespace
  the MCP write path stamps; this spec's classifier *reads* it. The classifier
  is a thin query over this convention, so **no new DEC** is warranted.
- `DEC-015` — the normalized `tags`/`taggings` join the provenance predicate
  uses (an `EXISTS` over `taggings` with prefix-anchored `LIKE`), the same
  join the `--tag` filter already uses.

### Constraints that apply

- `no-sql-in-cli-layer` — the classification predicate is SQL and lives in
  `internal/storage`; the CLI only validates the flag value and passes a
  string. Composing with `--limit` also *requires* the filter to be in SQL
  (a CLI post-fetch filter would break `LIMIT`).
- `stdout-is-for-data-stderr-is-for-humans` — the invalid-value error is an
  `ErrUser` (stderr, exit 1); filtered rows go to stdout unchanged.
- `storage-tests-use-tempdir` — the storage test opens under `t.TempDir()`.
- `one-spec-per-pr` — this PR is SPEC-043 only.

### Prior related work

- `SPEC-040` (shipped, PR #61) — the MCP provenance write path this reads.
- `SPEC-041` (shipped, PR #62) — documents the provenance convention in the
  plugin assets.

### Out of scope (for this spec specifically)

- **Time-windowed provenance share / self-reference density** — that is
  action-register **P3** (dogfooding-coverage query, STAGE-010). This spec
  ships the minimal distinguishing filter P3 builds on, not the trend report.
- **First-class `agent`/`model` columns** — the DEC-024 "later, if earned"
  promotion; the tag convention stays the classifier.
- **A provenance breakdown in `brag stats`** — would touch the locked DEC-014
  stats envelope + goldens; a separate spec if wanted.

## Notes for the Implementer

- Prefix-anchored `LIKE 'agent:%' OR LIKE 'model:%'` is deliberate — it
  matches only the reserved namespace, never a topic tag that merely starts
  with those letters (mirrors `TestList_TagFilterNoFalsePositive`). A bare
  `agent:` tag can't occur: `reservedTag` returns `""` for an empty value.
- `--author human` is the SQL negation (`NOT EXISTS ...`); keep the predicate
  a single shared const so the two branches can't drift.
- Reuse the accepted-value validation shape of the existing `--limit`/`--tag`
  guards (`UserErrorf`), so the byte-parity error contract is consistent.

---

## Build Completion

- **Branch:** `feat/spec-043-list-author-provenance-filter`
- **PR (if applicable):** references retro PR #64
- **All acceptance criteria met?** yes (all six; verified by unit tests +
  an end-to-end smoke run against a throwaway `--db`).
- **New decisions emitted:** none — rides DEC-024 (namespace) + DEC-015 (join).
- **Deviations from spec:** none.
- **Follow-up work identified:**
  - Action-register **P3** (dogfooding-coverage query) can now build on
    `--author`: provenance share over time, windowed by month.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing in the spec; the surprise was upstream: P2 assumed the MCP
   write path was unbuilt, but SPEC-040 had already shipped it the same day
   the retro landed. Confirming that (code + tests + `git log`) was the real
   first step, and it re-pointed this spec at the read half.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. `no-sql-in-cli-layer` correctly forced the filter into storage,
   which also happens to be the only way `--author` composes with `--limit`.

3. **If you did this task again, what would you do differently?**
   — Check whether the "emit" half already shipped *before* scoping an
   "emit" spec — a two-minute `git log`/grep that would have re-pointed the
   work immediately.

---

## Reflection (Ship)

1. **What would I do differently next time?**
   — When a retrospective item and the spec it critiques land on the same
   day, re-read the shipped code before treating the item's premise as
   current. P2 read as "emit provenance," but emission was already merged;
   the live gap was the read surface.

2. **Does any template, constraint, or decision need updating?**
   — No. DEC-024's "reserved tags now, first-class columns later if earned"
   path is validated by this read surface working over tags with zero schema
   change.

3. **Is there a follow-up spec I should write now before I forget?**
   — P3 (dogfooding-coverage query) is the natural next spec and now
   unblocked; it extends `--author` classification into a time-windowed
   provenance-share report. Left for STAGE-010 per the register.

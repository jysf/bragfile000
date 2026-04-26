---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-022
  type: story                      # epic | story | task | bug | chore
  cycle: verify
  blocked: false
  priority: medium
  complexity: M                    # S | M | L  (L means split it)
                                   # M (not S): three new artifacts
                                   # (JSON schema + shell hook + slash-
                                   # command template) + a BRAG.md
                                   # cross-reference insertion + an
                                   # extension to scripts/test-docs.sh
                                   # adding ~23 new shell asserts. Each
                                   # artifact is small individually but
                                   # the integration story (schema as
                                   # documented contract; hook +
                                   # slash-command as reference
                                   # implementations consuming it) is
                                   # what earns the M.

project:
  id: PROJ-001
  stage: STAGE-005
repo:
  id: bragfile

agents:
  architect: claude-opus-4-7
  implementer: claude-opus-4-7     # usually same Claude, different session
  created_at: 2026-04-26

references:
  decisions:
    - DEC-011                      # JSON output shape — SQL-matching
                                   # field names; this schema mirrors
                                   # the per-entry shape (minus the
                                   # array wrapping) on the input side.
    - DEC-012                      # Stdin-JSON schema for `brag add
                                   # --json` — six locked choices that
                                   # the JSON Schema vocabulary
                                   # transcribes. The single source of
                                   # truth this spec mirrors. NO new
                                   # DEC: the schema is a JSON-Schema-
                                   # vocabulary expression of DEC-012,
                                   # not an independent decision.
    - DEC-004                      # Tags as comma-joined TEXT — the
                                   # schema constrains `tags` to
                                   # `type: string` (NOT `type: array`)
                                   # so JSON-Schema validators reject
                                   # array form at the document
                                   # boundary; matches DEC-012 choice 3
                                   # and the binary's runtime reject.
  constraints: []                  # No code-layer constraints touch
                                   # this spec; new-artifact work is
                                   # outside `no-sql-in-cli-layer` /
                                   # `storage-tests-use-tempdir` /
                                   # `stdout-is-for-data-stderr-is-for-humans`
                                   # / `test-before-implementation` —
                                   # those govern Go code. The Failing
                                   # Tests below ARE written before the
                                   # build cycle in spirit of TDD;
                                   # artefacts are JSON + shell +
                                   # markdown, not Go.
  related_specs:
    - SPEC-014                     # emitted DEC-011; schema mirrors
                                   # DEC-011's per-entry shape on the
                                   # input side.
    - SPEC-017                     # emitted DEC-012; the binary's
                                   # stdin parser is the runtime
                                   # validator the schema documents.
    - SPEC-018                     # audit-grep cross-check addendum
                                   # (applies forward — see Premise
                                   # audit; mostly no-op here, lighter
                                   # surface than SPEC-021).
    - SPEC-019                     # NOT-contains self-audit pattern
                                   # (applies forward — see § NOT-
                                   # contains self-audit; very few
                                   # NOT-contains assertions in this
                                   # spec, mostly positive shape).
    - SPEC-020                     # negative-substring self-audit
                                   # codified.
    - SPEC-021                     # DIRECT precedent: created
                                   # scripts/test-docs.sh which this
                                   # spec extends (per its docstring
                                   # "single script that grows
                                   # internally as later STAGE-005
                                   # specs add doc-content asserts").
                                   # First confirming application of
                                   # the §12 literal-artifact-as-spec
                                   # pattern post-codification.
    - SPEC-023                     # successor: distribution proper.
                                   # Will reference the schema's
                                   # canonical URL when the README's
                                   # brew-install line activates;
                                   # otherwise independent.
---

# SPEC-022: ai integration distribution asset

## Context

The MVP `brag` CLI ships a strict-shape stdin contract for `brag add
--json` (locked at SPEC-017 / DEC-012, 2026-04-24): one JSON object,
`title` required, optional user-owned fields free-form text, server-
owned fields tolerated-and-ignored, unknown keys strict-rejected with
the offending key named, stdout = inserted ID. The contract is
documented in `BRAG.md` as prose for AI agents to read and in
`docs/api-contract.md` as the CLI surface reference. **What does not
yet exist** is a checked-in JSON Schema document that an AI agent (or
its surrounding tooling) can validate candidate payloads against
*before* piping to `brag add --json`. Today, an AI agent that drafts a
malformed payload (typo in a key, array-shaped tags) discovers the
problem only when the binary rejects it; with a checked-in schema, the
agent's framework can catch the problem at draft time.

`BRAG.md` (2026-04-22, pre-STAGE-003) is the canonical AI-agent
integration guide and already names the field shape as a markdown
table. It does not yet name a schema URL or point at any concrete
hook/slash-command examples — the "how" of integrating with Claude
Code (or any AI assistant with a session-end hook + slash-command
surface) is missing. STAGE-005's framing identified this as the
second of three workstreams before PROJ-001 closes (after SPEC-021's
README rewrite, before SPEC-023's distribution proper).

This spec is **STAGE-005's second workstream** (SPEC-021 shipped
2026-04-25; SPEC-022 ships next; SPEC-023 + SPEC-024 follow). Per the
stage's spec backlog ordering, SPEC-022 is independent of SPEC-023 +
SPEC-024 and depends only on SPEC-021's `scripts/test-docs.sh`
harness existing (it does, post-2026-04-25). SPEC-022 ships **three
artifacts** plus a `BRAG.md` cross-reference insertion plus an
extension to `scripts/test-docs.sh`:

1. **`docs/brag-entry.schema.json`** — JSON Schema (draft 2020-12)
   transcribing DEC-012's six locked stdin choices into JSON Schema
   vocabulary. The schema is documentation of the contract for
   external consumers (AI agents, integrations); the binary's stdin
   parser remains the authoritative validator. Mirrors DEC-011's
   nine-key shape on the input side.
2. **`scripts/claude-code-post-session.sh`** — example shell hook
   reading a session summary from stdin, structuring it as a
   schema-conforming JSON object via `jq`, and emitting a candidate
   `brag add --json` invocation for the user to review. Pure shell +
   `jq`; no Go dependency. Honours BRAG.md's approval loop (does NOT
   auto-execute `brag add`).
3. **`examples/brag-slash-command.md`** — example slash-command
   template (the markdown shape Claude Code's `~/.claude/commands/`
   directory expects). Tight 10-line prompt that triggers Claude to
   draft a brag entry validating against the schema. User copies the
   file to their own `~/.claude/commands/brag.md` to expose `/brag`
   in their Claude Code sessions.

Plus **`BRAG.md` cross-reference**: a new section pointing at the
schema as the contract AI agents validate against, and at the hook
script + slash-command template as concrete reference implementations.
Plus **`scripts/test-docs.sh` extension**: four new groups (H–K) of
shell asserts validating the three artifacts and the BRAG.md insertion
(per SPEC-021's docstring "single script that grows internally as
later STAGE-005 specs add doc-content asserts").

The brief's STAGE-005 sketch (line ~289) and the stage's success
criterion 4 (line ~89) lock this spec's scope. The framing's Q3
answer locks **example + docs only — no new `brag` CLI surface** like
`brag install-claude-hook`. The framing's pre-design Q5 lock holds:
no LLM piping in the binary (PROJ-002 territory).

Parents:
- Project: `projects/PROJ-001-mvp/` (PROJ-001 — MVP wave; closes when
  STAGE-005 ships).
- Stage: `projects/PROJ-001-mvp/stages/STAGE-005-distribution-and-cleanup.md`
  (the stage; SPEC-022 is its second spec). Stage notes lock the
  artifact set (`STAGE-005:Spec Backlog:SPEC-022`, ~lines 217–225)
  and the structural recommendations
  (`STAGE-005:Design Notes:SPEC-022-specific`, ~lines 358–391).
- Repo: `bragfile`.

Stage-level locked decisions that bind this spec:
- **No new DECs expected.** DEC-012 already captures the six locked
  stdin choices; this spec mirrors them as JSON Schema vocabulary —
  that is transcription, not an independent decision.
  (`STAGE-005:Design Notes:Cross-cutting`, line ~252.)
- **Trim heuristic applies cautiously** — SPEC-021 establishes the
  STAGE-005 doc-restructure + shell-asserts construction precedent.
  SPEC-022 inherits it for the prose-around-artifacts and the
  test-docs.sh extension; **literal artifacts get NO trim** (the §12
  pattern requires verbatim embed). Default to fuller skeleton if
  unsure. (`STAGE-005:Design Notes:Cross-cutting`, line ~272 + task-
  brief carry-forward.)
- **Premise audit applicability is asymmetric — light here.**
  SPEC-022 is mostly new-file work (one new JSON file, one new shell
  script, one new markdown file, one new section in BRAG.md, one
  extension to scripts/test-docs.sh). The audit-family rules
  (inversion-removal, count-bump, status-change) have minimal
  triggers — see Premise audit § "Family applicability" below.
  (`STAGE-005:Design Notes:Cross-cutting`, line ~286.)
- **§9 BSD-grep `--exclude-dir` rule applies forward.** SPEC-022's
  test-docs.sh extension uses `grep -rn -F` for some asserts; the
  exclude-dir set is decorative on macOS, the case-statement
  post-filter is the correctness boundary. (`AGENTS.md §9`, line
  ~215, codified at SPEC-021 ship.)
- **§12 literal-artifact-as-spec pattern applies maximally.** Three
  fixed-shape artifacts decidable at design time; all three embedded
  verbatim under Notes for the Implementer. (`AGENTS.md §12`, line
  ~314, codified at SPEC-021 ship — SPEC-022 is the first
  post-codification application.)
- **§10 push-discipline rule applies at build merge time.** Codified
  at STAGE-005 framing (2026-04-25); SPEC-021 was the first proactive
  application (held cleanly); SPEC-022 is the second.
  (`AGENTS.md §10`, line ~242.)

Five framing decisions (Q1–Q5) locked at design 2026-04-26 close
the open structural questions in the stage's SPEC-022-specific
Design Notes. Captured below in § Locked design decisions.

External Claude review (2026-04-24) seeded this spec's existence;
user concurred 2026-04-24; STAGE-005 framed 2026-04-25; SPEC-021
shipped 2026-04-25; SPEC-022 scaffolded + design 2026-04-26.

## Goal

Ship a JSON Schema (draft 2020-12) at `docs/brag-entry.schema.json`
mirroring DEC-012's stdin contract, plus two reference assets
(`scripts/claude-code-post-session.sh` shell hook + `examples/brag-slash-command.md`
slash-command template) demonstrating the schema's use in a Claude
Code session-end + slash-command workflow, plus a `BRAG.md`
cross-reference insertion pointing AI agents at the schema as the
validation contract and at the two assets as reference implementations,
with shape verified by an extension to `scripts/test-docs.sh` (new
groups H–K, ~23 new shell asserts) exposed via the existing `just
test-docs` recipe.

## Inputs

**Files to read (build-cycle reading list):**

- `AGENTS.md` — especially §6 (Cycle Model), §7 (Cross-Reference
  Rules), §8 (Coding Conventions — applies to the shell hook's
  `errors-wrap-with-context`-equivalent shape: `printf … >&2; exit N`
  for shell error reporting), §9 (Testing Conventions — the four
  premise-audit addenda + the audit-grep cross-check + the BSD-grep
  `--exclude-dir` addendum codified at SPEC-021 ship), §10 (Git/PR
  Conventions, particularly the push-discipline rule), §11 (Domain
  Glossary — verify whether new entries are needed for "JSON Schema",
  "Claude Code hook", "slash command"; see Premise audit § "Glossary
  cross-reference check"), §12 (Cycle-Specific Rules — particularly
  the design-time decision rule, the NOT-contains self-audit, and the
  literal-artifact-as-spec pattern).

- `BRAG.md` — repo-root file targeting AI agents. Currently 268 lines
  (2026-04-22 chore + small follow-ups). Documents the `brag` CLI's
  AI-integration story in prose. SPEC-022 inserts a new "## JSON
  contract for programmatic capture" section between the existing "##
  The command" section and the "## Three good examples" section. The
  insertion adds the three pointers (schema + hook + slash-command).
  The rest of the file stays byte-identical — see Outputs.

- `decisions/DEC-011-json-output-shape.md` — the OUTPUT shape (naked
  array of nine-key entry objects). The schema does NOT cover this
  directly (the schema is per-entry, single-object, on the INPUT
  side). The schema mirrors DEC-011's nine-key shape minus the array
  wrapping; the field names match SQL columns verbatim per DEC-011
  choice 2.

- `decisions/DEC-012-brag-add-json-stdin-schema.md` — THE shape this
  spec's JSON schema mirrors. Read verbatim. The six choices in
  DEC-012's "Decision" section translate to JSON Schema vocabulary as
  follows:
  1. Single object input → schema's `type: "object"` (the schema
     describes one object; `brag list --format json`-style array
     wrapping is out of scope).
  2. Title required + non-empty → `required: ["title"]` +
     `properties.title.minLength: 1`.
  3. Optional user-owned fields are free-form text → each listed under
     `properties` with `type: "string"`. Tags is `type: "string"`
     (not `type: "array"`) per DEC-004 alignment; a JSON Schema
     validator presented with `{"tags": ["a"]}` rejects at the type
     check, matching the binary's runtime reject.
  4. Server-owned fields tolerated-and-ignored → `id`, `created_at`,
     `updated_at` are explicitly listed under `properties` so
     `additionalProperties: false` does not reject them. The schema
     documents that they are server-owned and silently dropped on
     input (in the field's `description`); it does not enforce that
     they MUST be absent (the binary tolerates them per DEC-012
     choice 4).
  5. Unknown keys strict-rejected → `additionalProperties: false`.
     A JSON Schema validator presented with `{"titl": "x"}` rejects
     `titl` as an unknown property, matching the binary's runtime
     reject (which surfaces the offending key per DEC-012 choice 5).
  6. Stdout = ID → out of schema scope; the schema is the input
     contract, not the output contract.

- `decisions/DEC-004-tags-comma-joined-for-mvp.md` — the rationale
  for `tags` as a comma-joined TEXT column. Schema constrains `tags`
  to `type: "string"` accordingly (NOT `type: "array"`). The schema's
  `properties.tags.description` references DEC-004 by name so a
  consumer reading the schema (not the DEC) knows why arrays are
  rejected.

- `BRAG.md` (re-read for the cross-reference insertion's exact
  placement) — the existing structure is: What is `brag`? → Your
  role → When to propose → Approval loop → How to compose (Fields +
  Field quality bar) → The command → Three good examples →
  Anti-examples → Reading entries back → If anything goes wrong →
  Short version → Source. Insertion goes between "## The command"
  and "## Three good examples" — see Notes for the Implementer §
  "BRAG.md insertion sketch" for the literal markdown.

- `projects/PROJ-001-mvp/brief.md` — STAGE-005 sketch lines ~289–296
  (the AI-integration distribution asset workstream). Confirms the
  three-artifact shape and the "examples + docs only" framing.

- `projects/PROJ-001-mvp/stages/STAGE-005-distribution-and-cleanup.md`
  — parent stage. Lock points: success criterion 4 (line ~89),
  scope line ~125 (the artifact list), Spec Backlog SPEC-022 entry
  (~line 217), SPEC-022-specific Design Notes (~lines 358–391).

- `projects/PROJ-001-mvp/specs/done/SPEC-021-readme-user-facing-rewrite-and-dev-process-migration.md`
  — DIRECT precedent. Read sections: Build-cycle order; Notes for
  the Implementer (the literal README/CONTRIBUTING/development.md
  sketches); the test-docs.sh sketch + the actual shipped script at
  `scripts/test-docs.sh`. SPEC-022 extends `scripts/test-docs.sh`
  per the script's own docstring; this is the construction precedent
  the trim heuristic carry-forward references.

- `projects/PROJ-001-mvp/specs/done/SPEC-018-brag-summary-aggregate-package-and-shape-dec.md`
  — for the literal-artifact-as-spec pattern in a different language
  (Go test fixtures embedded verbatim and byte-loaded by goldens;
  same pattern, different shape). Read § Notes for the Implementer.

- `projects/PROJ-001-mvp/specs/done/SPEC-017-brag-add-json-stdin-input.md`
  — emitted DEC-012; the binary-side parser the schema documents.
  Useful context but not load-bearing reading; DEC-012 is the
  authoritative source for what the schema must mirror.

- `scripts/test-docs.sh` — exists post-SPEC-021. The shipped 437-line
  POSIX-shell harness with groups A–G. SPEC-022 extends it in place
  per the docstring's invitation. Read all of it before extending —
  the helper functions (`assert_file_exists`, `assert_line_count_band`,
  `assert_contains_literal`, `assert_not_contains_iregex`) are
  reusable; the new groups H–K should reuse them rather than
  inventing new helpers.

- `scripts/status.sh` — reference for shell-script house style
  (`set -eu`, SCRIPT_DIR pattern, POSIX-portable awk).

- `scripts/specs-by-stage.sh` — sibling reference for shell-script
  house style.

- `guidance/constraints.yaml` — repo-level rules. None apply
  directly (this spec touches no Go code; the shell hook is a
  developer-facing example, not a CLI under the
  `stdout-is-for-data-stderr-is-for-humans` contract — its output
  is a candidate command for the user to review, which is human-
  facing, which is fine because it goes to stderr; the JSON
  candidate goes to stdout for piping).

**External APIs:** None. No HTTP, no LLM, no third-party services.

**Related code paths:**

- `internal/cli/add.go` — the binary-side parser DEC-012 governs.
  The schema documents this code's input contract; the schema does
  not modify this code. Read for context only if confirming the
  binary's actual behaviour against the schema's claims; the build
  session does NOT need to touch this file.

- `internal/cli/add_test.go` — `TestAddCmd_JSON_*` tests are the
  authoritative round-trip + validator tests. Schema does not
  change them.

## Outputs

### Files created

- **`docs/brag-entry.schema.json`** (NEW) — JSON Schema document,
  draft 2020-12. Single object schema for the `brag add --json`
  stdin payload. Mirrors DEC-012's six locked choices via JSON
  Schema vocabulary; mirrors DEC-011's per-entry nine-key shape on
  the input side. Sized ~80–110 lines (frontmatter `$schema` + `$id`
  + `title` + `description` + `type` + `required` + `additionalProperties`
  + `properties` block with nine entries each carrying `type` and
  `description`). See Notes for the Implementer § "docs/brag-entry.schema.json
  literal" for the verbatim file content.

- **`scripts/claude-code-post-session.sh`** (NEW) — example shell
  hook script. Reads stdin (the post-session summary text), derives
  a candidate title heuristically, structures the rest of stdin as
  the `description` field, and prints a JSON object validating
  against the schema to stdout (for piping to `brag add --json`)
  plus a hint to stderr explaining the candidate command. **Does
  NOT auto-execute `brag add`** — honours BRAG.md's approval loop.
  Pure shell + `jq`; no Go dependency. POSIX shebang
  (`#!/usr/bin/env bash`); chmod +x. Sized ~50–70 lines (well-
  commented). See Notes for the Implementer §
  "scripts/claude-code-post-session.sh literal" for the verbatim
  file content.

- **`examples/brag-slash-command.md`** (NEW) — example slash-command
  template. The markdown shape Claude Code's `~/.claude/commands/`
  directory expects (YAML frontmatter with `description` + a body
  prompt). Tight 10-line prompt that triggers Claude to draft a
  brag entry validating against the schema. User copies the file
  to their own `~/.claude/commands/brag.md` to expose `/brag` as a
  slash-command in their Claude Code sessions. Sized ~10–15 lines.
  See Notes for the Implementer § "examples/brag-slash-command.md
  literal" for the verbatim file content.

- **`examples/`** (NEW DIRECTORY) — created by the existence of
  the slash-command template above. No `examples/README.md` (locked
  rejection #4 — the BRAG.md cross-reference is the discoverability
  surface; an extra README in `examples/` would be a third
  surface for the same content).

### Files modified

- **`BRAG.md`** — one new section inserted between the existing
  "## The command" section (~line 153) and the existing "## Three
  good examples" section (~line 155). Section title: "## JSON
  contract for programmatic capture". Length: ~35–45 lines.
  Includes:
  - One paragraph naming the schema file and its role (contract for
    AI agents to validate against before piping).
  - The five schema-shape bullets (required title, optional fields,
    server-owned tolerate-and-ignore, unknown-key strict-reject,
    tags as string per DEC-004).
  - A single-line `brag add --json` example showing a minimal
    valid payload.
  - Two pointer bullets to `scripts/claude-code-post-session.sh`
    and `examples/brag-slash-command.md` as reference
    implementations.
  - One closing line reaffirming that both assets honour the
    approval loop.
  Line numbers across the rest of `BRAG.md` shift downward by the
  new section's length; the relative section ordering is otherwise
  unchanged. See Notes for the Implementer § "BRAG.md insertion
  sketch" for the verbatim insertion content.

- **`scripts/test-docs.sh`** — extended in place per the script's
  docstring "single script that grows internally as later STAGE-005
  specs add doc-content asserts". Four new groups appended after
  the existing Group G — Harness ergonomics:
  - **Group H — JSON Schema shape** (10 asserts on
    `docs/brag-entry.schema.json`).
  - **Group I — Hook script shape** (5 asserts on
    `scripts/claude-code-post-session.sh`).
  - **Group J — Slash-command template shape** (4 asserts on
    `examples/brag-slash-command.md`).
  - **Group K — BRAG.md cross-reference** (4 asserts on the new
    `BRAG.md` section).
  Total new asserts: 23. Existing groups A–G unchanged in shape;
  the FAIL_COUNT logic and the final `ALL OK` print at the bottom
  of the script are unchanged. The harness self-pass meta (group
  G's F4 line) stays at the very end. See Notes for the Implementer
  § "test-docs.sh extension sketch" for the literal new shell.

### Files NOT modified by this spec (deferred / byte-identity
invariants)

- **`README.md`** — byte-identical to its post-SPEC-021 state. The
  schema is referenced from `BRAG.md`, not from `README.md`. README's
  "AI integration" pointer (the `BRAG.md` link) already exists
  post-SPEC-021; a deeper drill from README to the schema would be
  redundant.
- **`AGENTS.md`** — byte-identical. SPEC-022 does not touch
  AGENTS.md; the §11 Domain Glossary check (see Premise audit)
  concludes that "JSON Schema", "Claude Code hook", and "slash
  command" do NOT need glossary entries (they are well-known
  industry terms; the glossary is for *project-specific* terms like
  "aggregate", "tap", "Store", "digest").
- **`CONTRIBUTING.md`** — byte-identical (just shipped in SPEC-021).
- **`docs/development.md`** — byte-identical (just shipped in
  SPEC-021).
- **`docs/tutorial.md`** — byte-identical. Line 121's "For
  programmatic capture — a Claude session-end hook, an import from
  another tool, an integration with an issue tracker — pipe a JSON
  object on stdin to `brag add --json`" is an *example list of
  use cases*, not a status claim invalidated by SPEC-022 shipping
  the hook example. The prose stays accurate; pointing tutorial.md
  at the new artifacts would be a doc-sweep, deferred to SPEC-023's
  doc-sweep along with the other tutorial.md updates SPEC-021
  enumerated.
- **`docs/api-contract.md`** — byte-identical. Already references
  DEC-012 at line 416; that reference describes the binary's
  parser, which is unchanged. Adding a schema cross-link here would
  be a doc-sweep deferred to SPEC-023.
- **`docs/architecture.md`** — byte-identical. No mention of JSON
  Schema or hooks or slash commands; nothing to update.
- **`docs/data-model.md`** — byte-identical. Already references
  DEC-012 at line 148; same reasoning as api-contract.md.
- **`docs/CONTEXTCORE_ALIGNMENT.md`** — byte-identical. Out of
  scope.
- **`GETTING_STARTED.md`** — byte-identical.
- **`FIRST_SESSION_PROMPTS.md`** — byte-identical.
- **`justfile`** — byte-identical. The existing `test-docs` recipe
  (added by SPEC-021) already wires `./scripts/test-docs.sh`; the
  in-place extension to that script does not require a new recipe.
- **`scripts/_lib.sh`, `scripts/status.sh`, `scripts/new-spec.sh`,
  `scripts/new-stage.sh`, `scripts/advance-cycle.sh`,
  `scripts/archive-spec.sh`, `scripts/specs-by-stage.sh`,
  `scripts/info.sh`, `scripts/weekly-review.sh`** —
  byte-identical. SPEC-022 extends `scripts/test-docs.sh` only.
- **`LICENSE`, `.gitignore`, `go.mod`, `go.sum`, `cmd/`,
  `internal/`** — byte-identical. No Go code changes; no Go
  dependency added (the schema test approach is `jq`-based, not
  Go-based — Q4 lock).
- **`BRAG.md` content outside the inserted section** — byte-
  identical. Insertion is purely additive between two existing
  section headers.

### New exports / database changes

None.

## Acceptance Criteria

Testable outcomes. Cover positive structure (what the new artifacts
MUST contain), schema-vocabulary fidelity to DEC-012, hook + slash-
command shape, BRAG.md cross-reference completeness, and byte-
identity for the file invariants.

### H. JSON Schema shape

- [ ] **H1.** `docs/brag-entry.schema.json` exists.
- [ ] **H2.** `docs/brag-entry.schema.json` is parseable by `jq` —
      i.e., is syntactically valid JSON. (`jq -e . docs/brag-entry.schema.json
      > /dev/null` exits 0.)
- [ ] **H3.** Schema declares `$schema:
      "https://json-schema.org/draft/2020-12/schema"` (Q3 lock —
      draft 2020-12).
- [ ] **H4.** Schema declares `type: "object"` at the root.
- [ ] **H5.** Schema declares `required: ["title"]` at the root
      (DEC-012 choice 2).
- [ ] **H6.** Schema declares `additionalProperties: false` at the
      root (DEC-012 choice 5).
- [ ] **H7.** `properties.title` declares `type: "string"` AND
      `minLength: 1` (DEC-012 choice 2 — non-empty after trim;
      `minLength: 1` is the JSON Schema vocabulary equivalent of
      `strings.TrimSpace(t) != ""`'s minimum surface guarantee).
- [ ] **H8.** `properties.tags` declares `type: "string"` (NOT
      `type: "array"`). DEC-012 choice 3 / DEC-004 alignment — a
      JSON Schema validator presented with `{"tags": ["a"]}`
      rejects at the type check.
- [ ] **H9.** Schema's `properties` block lists all nine DEC-011
      keys: `title, description, tags, project, type, impact, id,
      created_at, updated_at`. Each has a `type` declaration and a
      `description`. (Server-owned fields explicitly listed so
      `additionalProperties: false` doesn't reject them — DEC-012
      choice 4.)
- [ ] **H10.** Schema declares `$id` as the canonical URL
      `https://github.com/jysf/bragfile000/blob/main/docs/brag-entry.schema.json`
      (the public location once the repo is cloned / browsed via
      GitHub). Stable URL so consumer-side tooling can reference
      `$id` for cache keys.

### I. Hook script shape

- [ ] **I1.** `scripts/claude-code-post-session.sh` exists.
- [ ] **I2.** `scripts/claude-code-post-session.sh` is executable
      (`test -x scripts/claude-code-post-session.sh`).
- [ ] **I3.** `scripts/claude-code-post-session.sh` has a POSIX
      shebang on line 1 — matches `^#!/(usr/bin/env (bash|sh)|bin/sh)`.
      (Mirrors SPEC-021's F3.)
- [ ] **I4.** `scripts/claude-code-post-session.sh` references
      `brag add --json` (the integration target) — at least one
      `grep -F 'brag add --json'` hit.
- [ ] **I5.** `scripts/claude-code-post-session.sh` references
      `jq` (the structuring tool — and the script's only non-stdlib
      dependency, named explicitly so the user knows what to
      install).

### J. Slash-command template shape

- [ ] **J1.** `examples/brag-slash-command.md` exists.
- [ ] **J2.** `examples/brag-slash-command.md` is between 5 and 30
      lines inclusive (target ~10–15; tight per Q5a lock).
- [ ] **J3.** `examples/brag-slash-command.md` references
      `docs/brag-entry.schema.json` (the schema the prompt asks
      Claude to validate against).
- [ ] **J4.** `examples/brag-slash-command.md` references
      `brag add --json` (the integration target the user pipes
      Claude's draft into).

### K. BRAG.md cross-reference

- [ ] **K1.** `BRAG.md` references `docs/brag-entry.schema.json`
      at least once (the schema cross-reference — load-bearing for
      this spec's value proposition).
- [ ] **K2.** `BRAG.md` references `scripts/claude-code-post-session.sh`
      at least once (the hook reference implementation).
- [ ] **K3.** `BRAG.md` references `examples/brag-slash-command.md`
      at least once (the slash-command reference implementation).
- [ ] **K4.** `BRAG.md` contains a heading matching `^## .*JSON`
      (the new section header — load-bearing for visual
      navigability when the user scans BRAG.md).

### L. Byte-identity invariants (manual verify-cycle checklist)

These are no-op-on-unrelated-surfaces sanity checks. They live in
the spec's verify-cycle review, not in `test-docs.sh`, because
comparing byte-identity across multiple unrelated files is more
efficient as a `git diff` review than as scripted asserts. (Same
pattern as SPEC-021's group H.)

- [ ] **L1.** `README.md` byte-identical to its pre-spec state
      (post-SPEC-021).
- [ ] **L2.** `AGENTS.md` byte-identical.
- [ ] **L3.** `CONTRIBUTING.md`, `GETTING_STARTED.md`,
      `FIRST_SESSION_PROMPTS.md` byte-identical.
- [ ] **L4.** `docs/development.md`, `docs/tutorial.md`,
      `docs/api-contract.md`, `docs/architecture.md`,
      `docs/data-model.md`, `docs/CONTEXTCORE_ALIGNMENT.md`
      byte-identical.
- [ ] **L5.** `justfile` byte-identical (the existing `test-docs`
      recipe wires `./scripts/test-docs.sh` already; no recipe
      change needed).
- [ ] **L6.** All `scripts/*.sh` files OTHER than `scripts/test-docs.sh`
      byte-identical.
- [ ] **L7.** `cmd/`, `internal/`, `go.mod`, `go.sum`, `LICENSE`,
      `.gitignore` byte-identical (no Go changes; no dep changes).

### M. Sanity (no Go regression)

- [ ] **M1.** `go test ./...` still passes on the post-spec tree.
- [ ] **M2.** `gofmt -l .` produces no output.
- [ ] **M3.** `go vet ./...` produces no output.
- [ ] **M4.** `just test-docs` exits 0 on the post-spec tree (all
      pre-existing groups A–G + the new groups H–K all OK; total
      60+ asserts pass).

**Total:** 30 acceptance criteria across 6 groupings (H/I/J/K
scripted via test-docs.sh extension; L manual verify-cycle; M Go
sanity). Scripted assertions: 23 (groups H–K). Manual verify-cycle
checks: 7 (group L) + 4 (group M sanity) = 11.

## Failing Tests

Written during **design**, BEFORE build. The implementer's job in
**build** is to make these pass.

This spec extends `scripts/test-docs.sh` rather than writing new Go
tests. The test surface is shell asserts on file-shape invariants
(positive-contains, structural-grep, line-count band) — the same
shape as SPEC-021's groups A–G. There are **zero new Go test
functions** in this spec; the binary's `internal/cli/add_test.go`
already covers the runtime parser side of DEC-012 (SPEC-017's
`TestAddCmd_JSON_*` family).

The new groups H–K are appended to `scripts/test-docs.sh` after the
existing Group G — Harness ergonomics, before the FAIL_COUNT exit
logic. They reuse the existing helper functions (`assert_file_exists`,
`assert_line_count_band`, `assert_contains_literal`,
`assert_not_contains_iregex`); a single new helper
(`assert_jq_eq`) is added for jq-based JSON-shape asserts on the
schema file (group H). See Notes for the Implementer §
"test-docs.sh extension sketch" for the literal scaffold.

### `scripts/test-docs.sh` (EXTENDED)

Extension appends groups H–K after the existing Group G. Each
assertion follows the same `OK: <name>` / `FAIL: <name>: <reason>`
output convention; each increments `FAIL_COUNT` on failure;
script exit code follows the existing rule (0 iff `FAIL_COUNT ==
0`). All group H–K asserts run after groups A–G in order.

#### Group H — JSON Schema shape (positive)

- **H1 — Schema file exists**
  - Assertion: `test -f docs/brag-entry.schema.json`.
  - Failure mode: file absent.

- **H2 — Schema is valid JSON**
  - Assertion: `jq -e . docs/brag-entry.schema.json > /dev/null
    2>&1` exits 0.
  - Failure mode: JSON syntax error → `FAIL: H2: docs/brag-entry.schema.json
    is not valid JSON`.
  - Note: `jq` is required by the harness post-SPEC-022. Pre-spec
    the harness used `jq` only for fenced-block extraction in A5
    (which ran `awk`, not `jq`); SPEC-022 introduces a `jq`
    requirement explicitly. The harness emits a clear error if
    `jq` is missing — see Notes for the Implementer.

- **H3 — Schema declares draft 2020-12**
  - Assertion: `jq -r '."$schema"' docs/brag-entry.schema.json`
    equals `"https://json-schema.org/draft/2020-12/schema"`.
  - Failure mode: `$schema` absent or mismatched.

- **H4 — Schema declares object type at root**
  - Assertion: `jq -r '.type' docs/brag-entry.schema.json` equals
    `"object"`.
  - Failure mode: type absent or mismatched.

- **H5 — Schema requires title**
  - Assertion: `jq -e '.required | index("title")'
    docs/brag-entry.schema.json` exits 0 (returns the index, which
    is non-null and not false).
  - Failure mode: `title` not in `required` array.

- **H6 — Schema disallows additional properties**
  - Assertion: `jq -r '.additionalProperties'
    docs/brag-entry.schema.json` equals `"false"`.
  - Failure mode: `additionalProperties` is `true` or absent.

- **H7 — Title is non-empty string**
  - Assertion: `jq -r '.properties.title.type'
    docs/brag-entry.schema.json` equals `"string"` AND
    `jq -r '.properties.title.minLength'
    docs/brag-entry.schema.json` equals `"1"`.
  - Failure mode: title type wrong or `minLength` missing /
    not 1.

- **H8 — Tags is string (NOT array)**
  - Assertion: `jq -r '.properties.tags.type'
    docs/brag-entry.schema.json` equals `"string"`.
  - Failure mode: tags type is `array` or absent → `FAIL: H8:
    properties.tags.type must be "string" (DEC-004), got <X>`.
  - This is the single most load-bearing schema assertion: it is
    what catches an AI consumer that's trying to send
    `{"tags": ["a","b"]}` at the document boundary BEFORE the
    payload reaches the binary.

- **H9 — All nine DEC-011 keys present**
  - Assertion: for each of `title, description, tags, project,
    type, impact, id, created_at, updated_at`, assert
    `jq -e '.properties.<key>'
    docs/brag-entry.schema.json` exits 0.
  - Failure mode: any key absent → `FAIL: H9: properties.<key>
    missing`.
  - Implementation: a `for` loop over the nine names with one
    `jq -e` invocation per name.

- **H10 — Schema declares canonical $id URL**
  - Assertion: `jq -r '."$id"' docs/brag-entry.schema.json`
    equals `"https://github.com/jysf/bragfile000/blob/main/docs/brag-entry.schema.json"`.
  - Failure mode: `$id` absent or mismatched.

#### Group I — Hook script shape

- **I1 — Hook script exists**
  - Assertion: `test -f scripts/claude-code-post-session.sh`.

- **I2 — Hook script is executable**
  - Assertion: `test -x scripts/claude-code-post-session.sh`.
  - Failure mode: not executable → `FAIL: I2:
    scripts/claude-code-post-session.sh is not executable
    (chmod +x)`.

- **I3 — Hook script has POSIX shebang**
  - Assertion: `head -n 1 scripts/claude-code-post-session.sh`
    matches `^#!(/usr/bin/env (sh|bash)|/bin/sh)`. (Same regex as
    SPEC-021's F3 — the existing `assert_*` infrastructure can
    grow a new `assert_posix_shebang` helper, OR the
    test-docs.sh extension can inline the check; either works,
    inline is simpler given two callsites.)
  - Failure mode: shebang absent or non-POSIX.

- **I4 — Hook script references `brag add --json`**
  - Assertion: `assert_contains_literal "I4"
    scripts/claude-code-post-session.sh "brag add --json"`.
  - Failure mode: literal absent.

- **I5 — Hook script references `jq`**
  - Assertion: `assert_contains_literal "I5"
    scripts/claude-code-post-session.sh "jq"`.
  - Failure mode: literal absent. (`jq` is the script's only
    non-stdlib dependency; naming it explicitly in the
    "missing prereqs" check + the json-construction call site
    is documentation for the user.)

#### Group J — Slash-command template shape

- **J1 — Template file exists**
  - Assertion: `test -f examples/brag-slash-command.md`.

- **J2 — Template is tight (5–30 lines)**
  - Assertion: `assert_line_count_band "J2"
    examples/brag-slash-command.md 5 30`.
  - Failure mode: outside band (Q5a lock — tight, target 10–15).

- **J3 — Template references the schema**
  - Assertion: `assert_contains_literal "J3"
    examples/brag-slash-command.md "docs/brag-entry.schema.json"`.

- **J4 — Template references `brag add --json`**
  - Assertion: `assert_contains_literal "J4"
    examples/brag-slash-command.md "brag add --json"`.

#### Group K — BRAG.md cross-reference

- **K1 — BRAG.md references the schema**
  - Assertion: `assert_contains_literal "K1" BRAG.md
    "docs/brag-entry.schema.json"`.
  - Failure mode: BRAG.md insertion missing the schema pointer.

- **K2 — BRAG.md references the hook script**
  - Assertion: `assert_contains_literal "K2" BRAG.md
    "scripts/claude-code-post-session.sh"`.

- **K3 — BRAG.md references the slash-command template**
  - Assertion: `assert_contains_literal "K3" BRAG.md
    "examples/brag-slash-command.md"`.

- **K4 — BRAG.md has a JSON-contract section heading**
  - Assertion: `grep -E '^## .*JSON' BRAG.md` returns ≥1 hit.
    (Generic enough that the build session can pick the exact
    heading wording — "## JSON contract for programmatic
    capture" is the recommended literal but "## JSON schema"
    or similar passes too.)
  - Failure mode: no `## … JSON …` heading in BRAG.md.

#### Group L — Verify-cycle checklist (NOT in script; manual)

These are no-op-on-unrelated-surfaces sanity checks. They live in
the spec's verify-cycle review, not in `test-docs.sh`. Same shape
as SPEC-021's group H.

- **L1 — `README.md` byte-identical**: `git diff HEAD --
  README.md` empty.
- **L2 — `AGENTS.md` byte-identical**.
- **L3 — `CONTRIBUTING.md`, `GETTING_STARTED.md`,
  `FIRST_SESSION_PROMPTS.md` byte-identical**.
- **L4 — `docs/{development,tutorial,api-contract,architecture,
  data-model,CONTEXTCORE_ALIGNMENT}.md` byte-identical**.
- **L5 — `justfile` byte-identical**.
- **L6 — `scripts/*.sh` other than `scripts/test-docs.sh`
  byte-identical**.
- **L7 — `cmd/`, `internal/`, `go.mod`, `go.sum`, `LICENSE`,
  `.gitignore` byte-identical**.

#### Group M — Sanity (manual)

- **M1 — `go test ./...` passes**.
- **M2 — `gofmt -l .` empty**.
- **M3 — `go vet ./...` empty**.
- **M4 — `just test-docs` exits 0** (full harness — pre-existing
  A–G + new H–K all OK).

### Test count summary

- **Group H:** 10 scripted assertions (jq-based JSON shape
  validation).
- **Group I:** 5 scripted assertions (file-existence + shebang +
  literal-contains).
- **Group J:** 4 scripted assertions (file-existence + line-band
  + literal-contains).
- **Group K:** 4 scripted assertions (literal-contains in
  BRAG.md).
- **Group L:** 7 manual verify-cycle byte-identity checks (NOT in
  script).
- **Group M:** 4 manual Go-sanity checks (NOT in script).

**Scripted total: 23 new assertions in `scripts/test-docs.sh`
groups H–K. Combined with the existing 40 from groups A–G, the
post-spec script runs 63 asserts total. Manual verify-cycle total:
11 (7 byte-identity + 4 Go sanity). Go test count: 0.**

The 1:1 mapping between Acceptance Criteria (30) and the scripted
+ manual asserts (23 + 11 = 34) holds with the gap explained by:
H9 is one criterion / one asserted loop over nine names (counted
as one); I3 inlines a regex check that could have been a helper;
J2's line-band counts as one. The arithmetic delta (30 vs 34) is
the SPEC-021 gap (40 vs 39 there) inverted — multiple sub-
assertions backing fewer criteria, not the other way.

## Implementation Context

*Read this section (and the files it points to) before starting the
build cycle. It is the equivalent of a handoff document, folded into
the spec since there is no separate receiving agent.*

### Decisions that apply

- **DEC-011** (shipped 2026-04-23, conf 0.85) — Shared JSON output
  shape. The schema mirrors DEC-011's nine-key per-entry shape on
  the input side (single object, not naked array). Field names
  match SQL columns verbatim per DEC-011 choice 2 — the schema
  inherits this naming directly.
- **DEC-012** (shipped 2026-04-24, conf 0.85) — Stdin-JSON schema
  for `brag add --json`. THE source of truth this spec mirrors.
  All six choices in DEC-012's "Decision" section translate to
  JSON Schema vocabulary as enumerated under § Inputs above.
  **No new DEC.** This spec is a JSON-Schema-vocabulary expression
  of DEC-012, not an independent decision.
- **DEC-004** (shipped 2026-04-19, conf 0.65) — Tags as comma-
  joined TEXT. Schema constrains `properties.tags.type: "string"`
  (NOT `array`); cross-referenced in the field's `description`
  by name. The H8 assertion is the single load-bearing shape test
  for this alignment.
- **DEC-001 through DEC-014** all apply forward unchanged in the
  broader project. The schema does not reference DECs 001 / 002 /
  003 / 005 / 006 / 007 / 008 / 009 / 010 / 013 / 014 — those
  govern internals (CGO posture, migrations, config resolution,
  IDs, cobra, RunE validation, FTS, etc.) that are below the
  schema's documentation level.

### Constraints that apply

The blocking constraint surface for SPEC-022 is essentially empty
because the spec touches no Go code:

- `no-sql-in-cli-layer` — N/A (no Go code).
- `storage-tests-use-tempdir` — N/A.
- `stdout-is-for-data-stderr-is-for-humans` — N/A as a *blocking*
  constraint (it governs the shipped `brag` binary's output
  channels, not the example hook script). The hook script
  *deliberately* honours the same shape: the JSON candidate goes
  to stdout (pipeable), the human-facing comment goes to stderr
  (advisory). This alignment is good practice, not a constraint
  violation if violated; it is a documentation-by-example of the
  rule.
- `test-before-implementation` — applies in spirit. The 23 new
  shell asserts in groups H–K are written in `scripts/test-docs.sh`
  before the new artifacts exist; running `just test-docs` early
  in the build session should fail-first on H–K (artifacts don't
  exist yet) and pass-first on A–G (SPEC-021's surface still
  intact). See Notes for the Implementer § "Build-cycle order".
- `no-new-top-level-deps-without-decision` — N/A. No Go deps
  added. `jq` is a runtime tooling dep for the shell harness +
  the hook script, not a Go dep; it's already implicitly required
  by SPEC-021's harness (post-extension it becomes explicit; the
  `jq` requirement is named at script entry per Q4 lock — see
  Notes for the Implementer).
- `one-spec-per-pr` — applies as always.
- `errors-wrap-with-context` — applies in spirit to the shell
  hook (errors written to stderr are prefixed with the script
  name `claude-code-post-session: …` so a user wiring it into a
  pipeline knows which script failed).
- `no-cgo` — N/A.

### AGENTS.md lessons that apply

All apply forward unchanged. Cross-refs only:

- **§9 audit-grep cross-check** (SPEC-018) — design enumeration
  in § Premise audit; build re-runs greps 1–4 per build-cycle
  step 3.
- **§9 BSD-grep `--exclude-dir`** (SPEC-021) — does not directly
  fire for groups H–K (no new `grep -r` calls); applies to any
  ad-hoc greps the build session writes.
- **§10 push-discipline** (STAGE-005 framing) — see build-cycle
  step 14 (push HEAD before merge).
- **§12 decide-at-design-time** (SPEC-018) — Q1–Q5 locked; see
  § Locked design decisions + § Rejected alternatives.
- **§12 NOT-contains self-audit** (SPEC-019/020) — does not fire;
  see § NOT-contains self-audit (n/a result).
- **§12 literal-artifact-as-spec** (SPEC-021) — applies maximally.
  Five literals embedded under Notes for the Implementer:
  schema (~85 ln), hook (~60 ln), slash-command (~12 ln),
  BRAG.md insertion (~40 ln), test-docs.sh extension (~120 ln).

### Prior related work

- **SPEC-014** (shipped 2026-04-23) — emitted DEC-011. The schema
  mirrors DEC-011's per-entry shape on the input side; the
  field-names-match-SQL choice (DEC-011 #2) is what makes the
  nine `properties` keys identical to the storage layer's
  column names.
- **SPEC-017** (shipped 2026-04-24) — emitted DEC-012. The
  binary's `internal/cli/add.go` runtime parser is the
  authoritative validator the schema documents. SPEC-022 does
  NOT modify that parser; the round-trip test
  (`TestAddCmd_JSON_RoundTripWithListJSON`) continues to govern
  the binary's behavior; the schema is independent prose-as-
  contract for external consumers.
- **SPEC-018** (shipped 2026-04-25) — origin of the audit-grep
  cross-check addendum (`AGENTS.md §9`, line ~226). SPEC-022's
  audit greps (run at design time below) reconcile to the
  enumerated `## Outputs` list with zero deltas. Build-side
  re-run before locking the artifacts is required per the rule.
- **SPEC-019** (shipped 2026-04-25) — first run of the
  NOT-contains self-audit pattern. SPEC-022 has no NOT-contains
  assertions; the pattern doesn't fire here. Audit captured
  below for completeness.
- **SPEC-020** (shipped 2026-04-25) — design-time pre-emption of
  the NOT-contains pattern.
- **SPEC-021** (shipped 2026-04-25) — DIRECT precedent. Created
  `scripts/test-docs.sh` with the explicit invitation in its
  docstring "single script that grows internally as later
  STAGE-005 specs add doc-content asserts". SPEC-022 is the
  first STAGE-005 spec to extend it post-codification. SPEC-021
  ship reflection's two AGENTS.md addenda (§9 BSD-grep + §12
  literal-artifact-as-spec) both apply forward; the latter is
  what governs SPEC-022's three verbatim-embedded artifacts.

### Out of scope (for this spec specifically)

Explicit list of what this spec does NOT include. If any of these
feel necessary during build, create a new spec rather than
expanding this one.

- **New `brag` CLI surface** for installing the hook or slash
  command (e.g. `brag install-claude-hook`, `brag install-slash-command`).
  Q3 of STAGE-005 framing locked example + docs only — this
  is a personal-tool repo; users copy the artifacts manually
  per the BRAG.md cross-reference's instructions.
- **Multi-file split of the schema.** Single
  `docs/brag-entry.schema.json`; not a per-field spread, not a
  `docs/schemas/` directory. (SPEC-022 is the only schema in
  PROJ-001; a `docs/schemas/` directory is premature.)
- **Build-time JSON Schema validation in `just test`.** Q4 lock
  — `just test` stays Go-only per SPEC-021's Q4 lock. Schema
  shape asserts go in `just test-docs` only. SPEC-023 may
  revisit if/when CI exists.
- **`~/.claude/` path introspection in the hook script.** Hook
  is pure stdin → JSON → `brag add --json`; user wires it into
  their own Claude Code config manually. Q5b lock.
- **Env-var introspection in the hook script** (e.g. reading
  `CLAUDE_PROJECT_DIR`, `CLAUDECODE`). Same Q5b lock — pure
  stdin only.
- **A Go-test that loads the schema and validates fixtures
  against it** (e.g. via `gojsonschema`). Adding a Go schema-
  validation library would be a new top-level dep needing a
  DEC; the binary's `internal/cli/add_test.go` already has
  round-trip tests that prove the parser's behaviour; adding a
  schema-validator on top is redundant and dep-expanding. Q4
  lock.
- **A `python -m jsonschema` or `ajv-cli`-based test in CI.**
  Same Q4 lock — pure jq + structural grep is the test
  approach.
- **Tutorial.md update pointing at the new artifacts.** Line
  121 mentions "a Claude session-end hook" in a generic example
  list; pointing it at the specific artifact would be a
  doc-sweep deferred to SPEC-023 (along with all the other
  tutorial.md updates SPEC-021 enumerated). The current line
  121 prose stays accurate.
- **`docs/api-contract.md` cross-link to the schema.** Same
  reasoning — doc-sweep deferred to SPEC-023.
- **`docs/data-model.md` cross-link to the schema.** Same.
- **`AGENTS.md §11 Domain Glossary` entries for "JSON Schema",
  "Claude Code hook", or "slash command".** Premise audit
  concludes none needed — these are well-known industry terms;
  the glossary is for *project-specific* terms like "aggregate",
  "tap", "Store", "digest".
- **Multi-platform Claude Code variants** (Claude Code Desktop,
  Claude Code Web, etc.). Hook script targets the current
  Claude Code agent shape (CLI / Code-IDE plugin). If Claude
  Code's settings change post-MVP, hook can be revised — out of
  scope for SPEC-022.
- **`brag completion zsh|bash|fish`.** SPEC-024.
- **`goreleaser`, GitHub Actions, homebrew tap wire-up,
  CHANGELOG.** SPEC-023.
- **LLM piping built into the binary.** PROJ-002.
- **Generic webhook / HTTP integrations.** Not the integration
  story; AI-agent + slash-command is the story.
- **Schema versioning / `$id` versioning** (e.g.,
  `docs/brag-entry-v1.schema.json`). Single schema file per
  the brief; if the schema needs to evolve breakingly, that's
  a paired DEC-011/DEC-012 migration with a new schema version
  — out of scope here.
- **A `examples/README.md` describing the examples directory.**
  Locked rejection #4. The BRAG.md cross-reference is the
  discoverability surface.
- **An empty `.gitkeep` in `examples/`.** Not needed — the
  directory has at least one file (the slash-command template).
- **Pulling from `backlog.md`.** STAGE-005's scope is
  distribution-shape per stage out-of-scope rule; backlog items
  are feature-shaped and stay backlogged.
- **Marketing copy / SEO / multi-language / branding.** Out of
  scope per task brief.
- **Tags-normalization / soft-delete / edit-history.** Stage
  out-of-scope.

### Premise audit

The audit is lighter than SPEC-021's because SPEC-022 is mostly
new-file work, not status-change. The audit-grep cross-check (§9
SPEC-018 addendum) was run at design time; results below
reconciled against the enumerated `## Outputs`.

#### Audit-grep enumeration (run at design time, 2026-04-26)

Format: grep + result. All greps run against working tree;
`projects/` excluded for path-naming greps (planning docs are not
the integrity surface). After SPEC-022 ships, the new artifacts
exist at the named paths and BRAG.md references them — that's
the intended delta.

**Grep 1 — `docs/brag-entry.schema.json`:**
```
grep -rn -F 'docs/brag-entry.schema.json' . --include='*.md' \
  --include='*.sh' --include='*.go' --include='*.yaml' \
  --include='*.json' --exclude-dir=projects \
  --exclude-dir=node_modules --exclude-dir=.git \
  --exclude-dir=framework-feedback
```
→ Zero hits in production surface. ✓ (~9 hits in `projects/`
planning docs, correctly excluded.)

**Grep 2 — `claude-code-post-session`:**
```
grep -rn -F 'claude-code-post-session' . --include='*.md' \
  --include='*.sh' --exclude-dir=node_modules --exclude-dir=.git \
  --exclude-dir=framework-feedback
```
→ Zero hits in production surface. ✓ (5 in `projects/`
planning, excluded.)

**Grep 3 — `brag-slash-command`:**
```
grep -rn -F 'brag-slash-command' . --include='*.md' \
  --include='*.sh' --exclude-dir=node_modules --exclude-dir=.git \
  --exclude-dir=framework-feedback
```
→ Zero hits EVERYWHERE. ✓ (Planning docs use the prose form
"`/brag` slash command", not the file slug.)

**Grep 4 — `examples/`:**
```
grep -rn -F 'examples/' . --include='*.md' --include='*.go' \
  --include='*.yaml' --exclude-dir=node_modules \
  --exclude-dir=.git --exclude-dir=framework-feedback
```
→ Zero hits in production surface. ✓ (4 in `projects/`
planning, excluded.) Directory does not yet exist.

**Grep 5 — JSON-Schema collision check:**
```
grep -rn -i 'json schema\|JSON Schema\|jsonschema' BRAG.md \
  docs/ AGENTS.md
```
→ 3 hits, all NO ACTION:
- `docs/api-contract.md:416` + `docs/data-model.md:148` —
  existing DEC-012 cross-references; describe binary parser
  (unchanged); schema cross-link is doc-sweep deferred to
  SPEC-023.
- `AGENTS.md:315–316` — §12 addendum commentary (not
  load-bearing).

**Grep 6 — Claude Code / slash-command collision check:**
```
grep -rn -i 'claude code\|claude-code\|slash command\|slash-command' \
  . --include='*.md' --include='*.sh' --exclude-dir=projects \
  --exclude-dir=node_modules --exclude-dir=.git \
  --exclude-dir=framework-feedback
```
→ 3 hits, all NO ACTION:
- `BRAG.md:3` — existing intro multi-agent framing; new section
  adds Claude Code detail without redacting it.
- `AGENTS.md:316` — §12 addendum commentary.
- `docs/tutorial.md:121` — generic example list; line stays
  accurate; doc-sweep deferred to SPEC-023.

**Grep 7 — `brag add --json` sanity:**
```
grep -rn -F 'brag add --json' BRAG.md docs/ README.md AGENTS.md
```
→ Hits in `docs/api-contract.md`, `docs/tutorial.md`,
`README.md` (post-SPEC-021); zero in BRAG.md (the JSON path is
the implicit topic of SPEC-022's new section). All NO ACTION;
SPEC-022's BRAG.md insertion adds the missing references.

**Grep 8 — Glossary check (AGENTS.md §11):**
```
grep -n -E '^- \*\*' AGENTS.md | head -30
```
→ Existing entries (aggregate, brag/entry, capture, digest,
Store, migration, export, review, summary, stats, tap) are all
*project-specific* terms. "JSON Schema", "Claude Code hook",
"slash command" are industry-standard. **No glossary additions
warranted.** Locked as Q1c.

#### Premise-audit family applicability

- **Inversion/removal** (planned test deletion under
  `## Outputs`): N/A. No existing test's premise is invalidated
  by SPEC-022; `internal/cli/add_test.go` continues to govern
  the binary's behaviour unchanged. No test deletions planned.
- **Addition** (count-bump on tracked collections): N/A for
  count-bumps that would break existing tests. Three "tracked
  collections" SPEC-022 modifies:
  1. **Files in `scripts/`**: pre-spec 9 entries; post-spec 9
     entries (no new file in `scripts/` — only an extension to
     `scripts/test-docs.sh` and one new file at
     `scripts/claude-code-post-session.sh`; the latter
     IS a new file, count goes 9 → 10). No existing test
     asserts on the count of files in `scripts/`. ✓ No
     count-bump needed in any test.
  2. **Files in `docs/`**: pre-spec 6 entries (`api-contract.md`,
     `architecture.md`, `CONTEXTCORE_ALIGNMENT.md`,
     `data-model.md`, `development.md`, `tutorial.md`) plus
     `reports/` subdir; post-spec 7 entries (adds
     `brag-entry.schema.json`). No existing test asserts on
     the count. ✓
  3. **Files in `examples/` (NEW DIRECTORY)**: pre-spec 0;
     post-spec 1. No existing test asserts on the count. ✓
  Other tracked collections (DECs at 14, migrations at 2,
  constraints at 11, `schema_migrations` rows) are untouched. ✓
- **Status change** (planned doc references update): N/A. The
  status of `brag add --json` is unchanged by SPEC-022 (still
  shipped per SPEC-017). The schema documents an existing
  contract; it does not change the contract's status. The
  existing doc references to DEC-012 in `docs/api-contract.md`
  and `docs/data-model.md` continue to be accurate; pointing
  them at the new schema file is the doc-sweep deferred to
  SPEC-023.
- **Audit-grep cross-check (both sides)** — Design-side run
  above; enumeration matches `## Outputs`. **Build-side**:
  re-run greps 1–4 (the four file-naming greps) before locking
  the new artifacts. Treat any new delta (e.g., a planning doc
  that accreted a new `claude-code-post-session.sh` mention
  between design and build) as a question for the spec author
  via the Build Completion reflection, not a unilateral
  expansion of scope.

#### NOT-contains self-audit (SPEC-019/SPEC-020 pattern)

**Result: n/a — no NOT-contains assertions in this spec.**

Groups H, I, J, K are all positive-shape asserts (file exists,
contains literal, declares JSON value). The NOT-contains pattern
does not fire here. SPEC-019/SPEC-020/SPEC-021 all had
NOT-contains assertions where load-bearing prose needed to
exclude specific tokens (e.g. README's group B in SPEC-021); SPEC-022
has no equivalent — the artifacts SPEC-022 ships have no
prose-exclusion requirements. The schema's `additionalProperties:
false` is a positive declaration ("the value of
`additionalProperties` is `false`"), not a NOT-contains assertion
in the SPEC-019/SPEC-020 sense.

For completeness: the BRAG.md insertion is additive (one new
section between two existing sections); no token in the rest of
BRAG.md is invalidated by the new section. No need to grep the
existing BRAG.md for forbidden tokens.

#### Glossary cross-reference check (AGENTS.md §11)

Verified above (grep 8). No glossary entries needed for "JSON
Schema", "Claude Code hook", or "slash command". The glossary's
discriminator is *project-specific*; SPEC-022's three new terms
are industry-standard. Captured as Q1c lock.

## Notes for the Implementer

Gotchas, style preferences, reuse opportunities, and **literal
artifacts** (the load-bearing section). Per the §12 literal-
artifact-as-spec pattern, the three new artifacts + the BRAG.md
insertion + the test-docs.sh extension are embedded verbatim
below; build byte-transcribes; verify diffs against these
literals.

### Locked design decisions (Q1–Q5, locked at design 2026-04-26)

Per the §12 design-time decision rule, alternatives explicitly
considered and rejected at design time (cross-referenced under
§ Rejected alternatives (build-time) below for the canonical
list):

- **Q1 — File locations.**
  - **Q1a (lock):** `scripts/claude-code-post-session.sh` for
    the hook script (path locked by stage success criterion 4
    line ~90 + brief line 139). `examples/brag-slash-command.md`
    for the slash-command template (path matches stage hint
    line ~92 + design Notes line ~383). The split between
    `scripts/` (executables) and `examples/` (reference
    content) is the cleanest semantic separation.
  - **Q1b (lock):** No `examples/README.md` describing the
    directory. The BRAG.md cross-reference is the
    discoverability surface for AI agents (the primary
    audience); contributor-facing discoverability is the
    `docs/development.md`'s "Where to find what" table — out of
    scope for SPEC-022. (See locked rejection #4.)
  - **Q1c (lock):** No `AGENTS.md §11 Domain Glossary` entries
    for the three new terms. See § Premise audit above.

- **Q2 — `scripts/test-docs.sh`: extend or sibling.** Lock:
  **EXTEND in place**, per the script's own docstring
  ("single script that grows internally as later STAGE-005
  specs add doc-content asserts") and per the precedent that
  extending avoids the test-{completions,…}.sh fan-out problem
  for SPEC-024 onwards. (See locked rejection #5.)

- **Q3 — JSON Schema draft version.** Lock: **draft 2020-12**.
  Confirmed against stage Design Notes line ~360. Personal-
  tool repo with no existing schema-consumer ecosystem; 2020-12
  is the current standard and is supported by the major
  validators (Ajv, jsonschema, gojsonschema). Asserted by H3.
  (See locked rejection #6.)

- **Q4 — Schema validator approach.** Lock: **`jq` +
  structural grep asserts in `scripts/test-docs.sh`** (group H).
  No new Go dep. No new tooling beyond `jq` (which the
  pre-existing harness's A5 fenced-block extraction does not
  use, but the new H2/H3/H4/H5/H6/H7/H8/H10 asserts will). The
  binary's `internal/cli/add_test.go` round-trip + parser tests
  are the authoritative runtime validation; the schema is
  documentation, and the H asserts are documentation-shape
  checks. (See locked rejection #7.)

- **Q5 — Slash-command template + hook env-var contract.**
  - **Q5a (lock):** **Tight slash-command prompt** (10–15
    lines, target 12). Easier to maintain; AI agents handle
    terse prompts well; the schema cross-reference does the
    field-by-field work. Asserted by J2 (5–30 line band; target
    inside that band).
  - **Q5b (lock):** **Pure stdin** for the hook script. No
    env-var introspection (`CLAUDE_PROJECT_DIR`,
    `CLAUDECODE`); no `~/.claude/` path lookups. Per stage
    scope guardrail line ~XXX ("No `~/.claude/` path
    introspection in the hook script — hook script is pure
    stdin → JSON → `brag add --json`"). User wires the hook
    into their own Claude Code config manually per BRAG.md's
    cross-reference instructions.

### Build-cycle order

1. **Start a fresh Claude session.** Do not continue from the
   design session. Read the spec's Implementation Context,
   AGENTS.md (especially §6, §9, §10, §12), and the parent
   stage file. Re-read DEC-011 + DEC-012 + DEC-004.

2. **Branch.** `git checkout -b feat/spec-022-ai-integration-distribution-asset`
   (or `chore/spec-022-…` if you prefer — both pass `one-spec-per-pr`;
   `feat/` is the convention since this spec ships new
   user-facing artifacts even though no Go code changes).

3. **Re-run the design-time greps (1–4 from § Premise audit).**
   Reconcile any delta against `## Outputs`. Raise discrepancies
   via Build Completion reflection rather than silently
   expanding scope.

4. **Extend `scripts/test-docs.sh` first** (TDD-in-spirit).
   Append groups H–K per the literal scaffold below in §
   "test-docs.sh extension sketch". Add the `jq` prereq check
   near the top of the script (one-time `command -v jq` check
   with a clear error if missing). Do NOT write the new
   artifacts yet.

5. **Run `just test-docs` against the current tree.** Expect:
   - Groups A–G all OK (SPEC-021's surface unchanged).
   - Groups H–K mostly FAIL (artifacts don't exist yet; BRAG.md
     hasn't been edited). Specifically expect H1, I1, J1, K1
     to FAIL because the files / content don't exist.
   **Confirm the FAILs are for the *expected* reason** (the
   assertion you wrote, not a stray shell error or `jq`
   missing). Per AGENTS.md §9 fail-first discipline.

6. **Write `docs/brag-entry.schema.json`.** Use the verbatim
   literal below. ~85 lines. Validate against `jq -e .` before
   committing.

7. **Write `scripts/claude-code-post-session.sh`.** Use the
   verbatim literal below. ~60 lines. `chmod +x`. Verify the
   shebang on line 1.

8. **Create `examples/` directory and write
   `examples/brag-slash-command.md`.** Use the verbatim literal
   below. ~12 lines.

9. **Update `BRAG.md`.** Insert the new section between the
   existing "## The command" section and the existing "## Three
   good examples" section. Use the verbatim literal below.

10. **Run `just test-docs` again.** Expect all asserts (A–G
    pre-existing + H–K new) to OK. Run `just test` separately
    to confirm `go test ./...` still passes (group M sanity).

11. **Run the manual group L verify-cycle byte-identity checks**
    (seven byte-identity diffs across README/AGENTS/CONTRIBUTING/
    GETTING_STARTED/FIRST_SESSION_PROMPTS/docs-six-files/justfile/
    other-scripts/Go-code). Confirm via `git diff --stat HEAD --`
    that the only modified files are exactly: `BRAG.md`,
    `scripts/test-docs.sh`, plus the four new files
    (`docs/brag-entry.schema.json`,
    `scripts/claude-code-post-session.sh`,
    `examples/brag-slash-command.md`, AND ANY untracked working-
    tree files unrelated to this spec — flag in Build Completion
    if so).

12. **Run `gofmt -l .` and `go vet ./...`** (group M sanity —
    M2/M3).

13. **Fill in `## Build Completion`** including all three
    reflection answers (real, not placeholder — `archive-spec.sh`
    rejects empty `<answer>` placeholders; same discipline for
    the build reflection).

14. **`just advance-cycle SPEC-022 verify`.**

15. **Open PR.** PR description includes `Project: PROJ-001`,
    `Stage: STAGE-005`, `Spec: SPEC-022`, "no DECs referenced or
    emitted (DEC-011/012/004 referenced, not new)",
    "constraints checked: N/A (no Go code)".

### Push-discipline reminder (§10)

**Critical at the merge step.** SPEC-021 was the first proactive
application of the rule post-codification (held cleanly); SPEC-022
is the second test. Concrete heuristic: after any `git commit` on
the feat branch in the same shell session as the PR merge, run
`git push origin HEAD` BEFORE `gh pr merge`. Especially relevant
for this spec because the four-artifact delivery (schema + hook +
slash-command + BRAG.md edit + test-docs.sh extension) creates
multiple commit opportunities that may interleave with verify
feedback.

### docs/brag-entry.schema.json literal

The verbatim file content. Build transcribes byte-for-byte.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://github.com/jysf/bragfile000/blob/main/docs/brag-entry.schema.json",
  "title": "Brag entry — `brag add --json` stdin contract",
  "description": "Single-object stdin payload accepted by `brag add --json`. Mirrors DEC-012's six locked choices in JSON Schema vocabulary; mirrors DEC-011's nine-key per-entry shape on the input side. The binary's runtime parser is the authoritative validator; this schema is the documented contract for AI agents and other external consumers to validate candidate payloads against before piping. See BRAG.md (https://github.com/jysf/bragfile000/blob/main/BRAG.md) for the integration guide.",
  "type": "object",
  "required": ["title"],
  "additionalProperties": false,
  "properties": {
    "title": {
      "type": "string",
      "minLength": 1,
      "description": "Required, non-empty after whitespace trim. Short, action-verb headline. Per DEC-012 choice 2."
    },
    "description": {
      "type": "string",
      "description": "Optional free-form narrative. Per DEC-012 choice 3."
    },
    "tags": {
      "type": "string",
      "description": "Optional comma-joined string of tags (e.g. \"auth,perf,backend\"). Per DEC-004, tags are stored as a single TEXT column inside the binary; per DEC-012 choice 3, the input contract mirrors this shape. Array form (e.g. [\"auth\",\"perf\"]) is rejected at the type check below. To migrate to array-shaped tags, DEC-004 + DEC-011 + DEC-012 + this schema migrate in lockstep."
    },
    "project": {
      "type": "string",
      "description": "Optional work-context label (repo, client, team, initiative). Per DEC-012 choice 3."
    },
    "type": {
      "type": "string",
      "description": "Optional category (e.g. \"shipped\", \"fixed\", \"learned\", \"documented\", \"mentored\", \"unblocked\", \"reviewed\"). Free-form; the binary does not enumerate. Per DEC-012 choice 3."
    },
    "impact": {
      "type": "string",
      "description": "Optional concrete-outcome statement: a metric, an unlock, a named result. The most load-bearing field for review writeups. Per DEC-012 choice 3."
    },
    "id": {
      "type": "integer",
      "description": "Server-owned. Permitted on input but silently dropped per DEC-012 choice 4 (the new row receives a fresh autoincrement ID from the storage layer). Listed here so additionalProperties=false does not reject the round-trip pattern `brag list --format json | jq '.[0]' | brag add --json`."
    },
    "created_at": {
      "type": "string",
      "format": "date-time",
      "description": "Server-owned. Permitted on input but silently dropped per DEC-012 choice 4 (the storage layer sets the timestamp at insert time, RFC3339 UTC per DEC-002 / timestamps-in-utc-rfc3339 constraint)."
    },
    "updated_at": {
      "type": "string",
      "format": "date-time",
      "description": "Server-owned. Permitted on input but silently dropped per DEC-012 choice 4 (the storage layer sets the timestamp at insert time, RFC3339 UTC per the same constraint as created_at)."
    }
  }
}
```

(End of `docs/brag-entry.schema.json` literal.)

**Notes on the schema:**
- `$id` uses `blob/main/` (human-readable GitHub view); tooling
  wanting raw bytes transforms `blob/` → `raw/`.
- Properties listed in DEC-011's nine-key SQL-column order
  (user-owned first, server-owned last) for human readability;
  JSON object key order is semantically unordered.
- `id: integer` per DEC-005; timestamps use `format: date-time`
  per JSON Schema convention for RFC3339.
- Top-level `description` cross-links to BRAG.md via the
  canonical URL so consumers reading the schema file in
  isolation can find the integration guide.

### scripts/claude-code-post-session.sh literal

The verbatim file content. Build transcribes byte-for-byte.
`chmod +x` after writing.

```bash
#!/usr/bin/env bash
# scripts/claude-code-post-session.sh — example Claude Code session-end hook.
#
# Reads a session summary from stdin, structures it as a JSON object
# matching docs/brag-entry.schema.json, and prints a candidate
# `brag add --json` payload to stdout (plus a hint to stderr explaining
# what to do with it).
#
# This script does NOT auto-execute `brag add`. It honours BRAG.md's
# approval loop: you (the user) review the candidate JSON, decide if
# the moment is brag-worthy, then pipe the JSON to `brag add --json`
# yourself.
#
# Wiring: copy this file wherever your Claude Code config wants it.
# As of late 2025, Claude Code reads hook config from
# ~/.claude/settings.json; consult Claude Code docs for the current
# hook-config shape. This script makes no assumptions about the
# invocation interface beyond "stdin carries the session summary".
#
# Dependencies: bash, jq, brag (on $PATH).

set -eu

if ! command -v jq >/dev/null 2>&1; then
    printf 'claude-code-post-session: jq is required but not installed (see https://stedolan.github.io/jq/)\n' >&2
    exit 2
fi

if ! command -v brag >/dev/null 2>&1; then
    printf 'claude-code-post-session: brag is required but not on $PATH (install via brew or `just install`)\n' >&2
    exit 2
fi

# Read the entire stdin into a variable. Empty stdin → no-op exit 0.
SUMMARY=$(cat)
if [ -z "${SUMMARY}" ]; then
    printf 'claude-code-post-session: no summary on stdin; skipping\n' >&2
    exit 0
fi

# Derive a candidate title heuristically: the first non-empty line of
# stdin, trimmed, capped at 100 characters. The user reviews and
# refines before approving.
TITLE=$(printf '%s\n' "${SUMMARY}" | awk 'NF { print; exit }' | cut -c1-100)
if [ -z "${TITLE}" ]; then
    printf 'claude-code-post-session: could not derive a title from stdin; skipping\n' >&2
    exit 0
fi

# Build the JSON payload via jq (handles string escaping safely).
# Fields included: title (required) + description (the full summary)
# + type (default "shipped"). The user can refine project / tags /
# impact before piping to `brag add --json`.
PAYLOAD=$(jq -n \
    --arg title "${TITLE}" \
    --arg description "${SUMMARY}" \
    --arg type "shipped" \
    '{title: $title, description: $description, type: $type}')

# Print the candidate payload to stdout (pipeable). Print the hint to
# stderr (human-facing). This mirrors the brag binary's stdout-is-for-
# data-stderr-is-for-humans contract by example.
printf '%s\n' "${PAYLOAD}"
printf '\nclaude-code-post-session: candidate JSON above. To capture, run:\n' >&2
printf '    <prev-stdout> | brag add --json\n' >&2
printf '(Or copy the JSON, refine project/tags/impact, then pipe to `brag add --json` manually.)\n' >&2

exit 0
```

(End of `scripts/claude-code-post-session.sh` literal.)

**Notes on the hook script:**
- `bash` shebang (not `sh`) because of the conventional `${VAR}`
  patterns; I3 asserts the POSIX-shebang regex which permits
  both forms.
- Exit 2 for missing prereqs; exit 0 for empty stdin (no-op is
  success).
- `jq -n --arg` is the safe JSON-construction idiom — handles
  embedded quotes/backslashes/newlines without manual escaping.
- Candidate JSON to stdout, advisory hint to stderr —
  documentation-by-example of the
  `stdout-is-for-data-stderr-is-for-humans` constraint shape.
- `type: "shipped"` default is the common case at session end;
  user refines before piping.

### examples/brag-slash-command.md literal

The verbatim file content. Build transcribes byte-for-byte. ~12
lines. Tight per Q5a lock.

```markdown
---
description: Draft a brag entry from this session
---

Review what shipped in this session. If a moment is brag-worthy per
[BRAG.md](https://github.com/jysf/bragfile000/blob/main/BRAG.md)
(shipped feature, fixed significant bug, architectural decision,
delivered artifact), draft a single JSON object validating against
[`docs/brag-entry.schema.json`](https://github.com/jysf/bragfile000/blob/main/docs/brag-entry.schema.json):
required `title` (action-verb, ≤100 chars), plus optional
`description`, `project`, `type`, `tags` (comma-joined string per
DEC-004), and `impact` (concrete metric or named outcome). Present
the JSON for my approval. Do not execute `brag add --json` until I
confirm.
```

(End of `examples/brag-slash-command.md` literal. Line count: 13
including the YAML frontmatter delimiters and trailing newline.)

**Notes on the slash-command template:**
- YAML frontmatter matches Claude Code's late-2025 shape
  (`~/.claude/commands/<name>.md` with optional `description` /
  `model` / `allowed-tools`); body remains useful even if
  frontmatter conventions evolve.
- Absolute URLs (not relative paths) because the file is copied
  out of the repo to the user's `~/.claude/commands/`.
- "≤100 chars" mirrors the hook script's title cap intentionally.
- Closing "Do not execute … until I confirm" is defense-in-depth
  for the approval loop in case Claude reads the slash command
  but skips the linked BRAG.md.

### BRAG.md insertion sketch

The verbatim insertion. Goes between the existing `## The command`
section (ends ~line 152, after the `brag show <id>` example block)
and the existing `## Three good examples` section (begins ~line
155). Build inserts the following ~40 lines of new content;
surrounding text byte-identical.

```markdown
## JSON contract for programmatic capture

For programmatic capture (you're a script, an AI agent, or a piped
session — not a human typing flags), `brag add --json` accepts a
single JSON object on stdin. The object's shape is documented in a
checked-in JSON Schema at
[`docs/brag-entry.schema.json`](docs/brag-entry.schema.json).
Validate your payload against the schema before piping; the binary
will reject malformed input but the schema lets your tooling catch
mistakes earlier.

The schema mirrors this guide's field table:

- **Required:** `title` (non-empty string).
- **Optional:** `description`, `tags`, `project`, `type`, `impact`
  (all strings; `tags` is a comma-joined string per DEC-004, NOT a
  JSON array).
- **Server-owned:** `id`, `created_at`, `updated_at` are tolerated
  on input and silently dropped — the new row gets fresh values.
  This makes `brag list --format json | jq '.[0]' | brag add
  --json` round-trip without `jq del(.id, .created_at, .updated_at)`.
- **Unknown keys are strict-rejected.** A typo like `{"titl": "x"}`
  surfaces as `unknown field "titl"` rather than silently losing
  your title.

Minimal valid payload:

```bash
echo '{"title":"shipped FTS5 search end-to-end","project":"bragfile","type":"shipped","tags":"sqlite,fts5,search","impact":"unblocked brag search"}' \
  | brag add --json
# → prints the new entry's ID on stdout
```

Two reference assets in this repo demonstrate AI-agent integration:

- [`scripts/claude-code-post-session.sh`](scripts/claude-code-post-session.sh)
  — example shell hook that reads a session summary from stdin,
  structures it as a schema-conforming JSON object via `jq`, and
  emits a candidate `brag add --json` invocation. Pure shell + `jq`;
  no Go dependency. Copy to wherever your Claude Code config
  expects hook scripts.
- [`examples/brag-slash-command.md`](examples/brag-slash-command.md)
  — example `~/.claude/commands/brag.md` slash-command template.
  Copy to your own `~/.claude/commands/` to expose `/brag` in
  Claude Code sessions.

Both assets honour the approval loop above: they help you draft a
brag entry, but you still review and approve before `brag add
--json` executes.

```

(End of `BRAG.md` insertion. Trailing blank line is intentional —
the next line in BRAG.md begins `## Three good examples`, and
markdown convention is one blank line between sections.)

**Notes on the BRAG.md insertion:**
- Heading `## JSON contract for programmatic capture` satisfies
  K4; three pointer literals satisfy K1/K2/K3.
- Relative paths because BRAG.md lives at repo root; the
  asymmetry with the slash-command template (absolute URLs) is
  deliberate — the slash-command file is copied out of the repo,
  BRAG.md normally is not.
- Example values (FTS5 search etc.) chosen to harmonize with
  the "Three good examples" section that follows.

### test-docs.sh extension sketch

The verbatim shell extension. Append to the existing
`scripts/test-docs.sh` after the existing `# ===== Group G —
Harness ergonomics =====` block ends and BEFORE the existing
`# ===== finalise =====` block begins.

The existing script reads:
```sh
# ... (existing groups A-G) ...

# G2 — Exit-code contract is built-in (FAIL_COUNT-driven exit at the bottom)
ok "G2"

# ===== finalise =====
```

The extension inserts the four new groups H, I, J, K BETWEEN
those two markers, plus a one-time `jq` prereq check inserted
near the top of the script (before the helpers section, after
the FAIL_COUNT initialization). The harness self-pass meta line
(`ok "F4"`) stays at the very end as before.

**Step (a) — `jq` prereq insertion** (added after `FAIL_COUNT=0`
on the existing line 13 of `scripts/test-docs.sh`):

```sh
# Group H asserts require jq for JSON-shape parsing.
if ! command -v jq >/dev/null 2>&1; then
    printf 'test-docs: jq is required but not installed (see https://stedolan.github.io/jq/)\n' >&2
    exit 2
fi
```

**Step (b) — Groups H–K appended** (inserted between
`ok "G2"` and `# ===== finalise =====`):

```sh
# ===== Group H — JSON Schema shape =====

SCHEMA_PATH="docs/brag-entry.schema.json"

# H1 — Schema file exists
assert_file_exists "H1" "$SCHEMA_PATH"

# H2 — Schema is valid JSON
if [ -f "$SCHEMA_PATH" ]; then
    if jq -e . "$SCHEMA_PATH" >/dev/null 2>&1; then
        ok "H2"
    else
        fail "H2" "$SCHEMA_PATH is not valid JSON"
    fi
else
    fail "H2" "$SCHEMA_PATH does not exist"
fi

# Helper for jq-based equality checks against the schema. Compares
# the jq-extracted value against an expected literal string.
assert_jq_eq() {
    name="$1"; expr="$2"; expected="$3"
    if [ ! -f "$SCHEMA_PATH" ]; then
        fail "$name" "$SCHEMA_PATH does not exist"
        return 0
    fi
    actual=$(jq -r "$expr" "$SCHEMA_PATH" 2>/dev/null || echo "<jq-error>")
    if [ "$actual" = "$expected" ]; then
        ok "$name"
    else
        fail "$name" "$expr returned \"$actual\" (expected \"$expected\")"
    fi
}

# H3 — Schema declares draft 2020-12
assert_jq_eq "H3" '."$schema"' "https://json-schema.org/draft/2020-12/schema"

# H4 — Schema declares object type at root
assert_jq_eq "H4" '.type' "object"

# H5 — Schema requires title
if [ -f "$SCHEMA_PATH" ]; then
    if jq -e '.required | index("title")' "$SCHEMA_PATH" >/dev/null 2>&1; then
        ok "H5"
    else
        fail "H5" '"title" not found in .required array'
    fi
else
    fail "H5" "$SCHEMA_PATH does not exist"
fi

# H6 — Schema disallows additional properties
assert_jq_eq "H6" '.additionalProperties' "false"

# H7 — Title is non-empty string
if [ -f "$SCHEMA_PATH" ]; then
    title_type=$(jq -r '.properties.title.type' "$SCHEMA_PATH" 2>/dev/null || echo "")
    title_min=$(jq -r '.properties.title.minLength' "$SCHEMA_PATH" 2>/dev/null || echo "")
    if [ "$title_type" = "string" ] && [ "$title_min" = "1" ]; then
        ok "H7"
    else
        fail "H7" "properties.title.type=\"$title_type\" minLength=\"$title_min\" (want type=\"string\" minLength=\"1\")"
    fi
else
    fail "H7" "$SCHEMA_PATH does not exist"
fi

# H8 — Tags is string (NOT array) — DEC-004 alignment, load-bearing
assert_jq_eq "H8" '.properties.tags.type' "string"

# H9 — All nine DEC-011 keys present in properties
if [ -f "$SCHEMA_PATH" ]; then
    h9_missing=""
    for key in title description tags project type impact id created_at updated_at; do
        if ! jq -e ".properties.$key" "$SCHEMA_PATH" >/dev/null 2>&1; then
            h9_missing="$h9_missing $key"
        fi
    done
    if [ -z "$h9_missing" ]; then
        ok "H9"
    else
        fail "H9" "missing properties:$h9_missing"
    fi
else
    fail "H9" "$SCHEMA_PATH does not exist"
fi

# H10 — Schema declares canonical $id URL
assert_jq_eq "H10" '."$id"' "https://github.com/jysf/bragfile000/blob/main/docs/brag-entry.schema.json"

# ===== Group I — Hook script shape =====

HOOK_PATH="scripts/claude-code-post-session.sh"

# I1 — Hook script exists
assert_file_exists "I1" "$HOOK_PATH"

# I2 — Hook script is executable
if [ -x "$HOOK_PATH" ]; then
    ok "I2"
else
    fail "I2" "$HOOK_PATH is not executable (chmod +x)"
fi

# I3 — Hook script has POSIX shebang on line 1
if [ ! -f "$HOOK_PATH" ]; then
    fail "I3" "$HOOK_PATH does not exist"
elif head -n 1 "$HOOK_PATH" | grep -E -q '^#!(/usr/bin/env (sh|bash)|/bin/sh)'; then
    ok "I3"
else
    fail "I3" "$HOOK_PATH missing POSIX shebang on line 1"
fi

# I4 — Hook script references `brag add --json`
assert_contains_literal "I4" "$HOOK_PATH" "brag add --json"

# I5 — Hook script references `jq`
assert_contains_literal "I5" "$HOOK_PATH" "jq"

# ===== Group J — Slash-command template shape =====

SLASH_PATH="examples/brag-slash-command.md"

# J1 — Template file exists
assert_file_exists "J1" "$SLASH_PATH"

# J2 — Template is tight (5–30 lines)
assert_line_count_band "J2" "$SLASH_PATH" 5 30

# J3 — Template references the schema
assert_contains_literal "J3" "$SLASH_PATH" "docs/brag-entry.schema.json"

# J4 — Template references `brag add --json`
assert_contains_literal "J4" "$SLASH_PATH" "brag add --json"

# ===== Group K — BRAG.md cross-reference =====

# K1 — BRAG.md references the schema
assert_contains_literal "K1" "BRAG.md" "docs/brag-entry.schema.json"

# K2 — BRAG.md references the hook script
assert_contains_literal "K2" "BRAG.md" "scripts/claude-code-post-session.sh"

# K3 — BRAG.md references the slash-command template
assert_contains_literal "K3" "BRAG.md" "examples/brag-slash-command.md"

# K4 — BRAG.md has a JSON-contract section heading
if [ ! -f BRAG.md ]; then
    fail "K4" "BRAG.md does not exist"
elif grep -E -q '^## .*JSON' BRAG.md; then
    ok "K4"
else
    fail "K4" "no '## … JSON …' heading in BRAG.md"
fi

```

(End of test-docs.sh extension. The existing
`# ===== finalise =====` block follows immediately after,
unchanged.)

**Notes on the test-docs.sh extension:**
- `jq` prereq check at top (exit 2 if missing) — without it,
  H2–H10 fail with confusing errors when jq isn't installed.
- Path variables (`SCHEMA_PATH` etc.) make path-renames safe.
- `assert_jq_eq` helper inlined in group H (only consumer);
  migrates to top if a future group reuses it.
- H7 inlines two-value logic (type + minLength) since
  `assert_jq_eq` is single-value.
- I3 inlines the shebang regex (same as SPEC-021 F3); two
  callsites is below the helper-extraction threshold.
- Existing `finalise` exit-code logic unchanged; `ok "F4"`
  meta-line stays at the very end.
- Clean run prints 63 OK lines (40 from A–G + 23 new) + the
  ALL OK footer.

### Style preferences and gotchas

- **JSON.** No JSON files exist in repo today; for the schema,
  follow the literal verbatim — 2-space indent, no trailing
  commas, single-line `description` strings.
- **Shell.** Match `scripts/status.sh` / `_lib.sh` /
  `test-docs.sh` style: `set -eu`, `printf` not `echo`. Hook
  needs no SCRIPT_DIR (pure stdin).
- **Markdown (slash-command + BRAG.md insertion).** Match
  `BRAG.md`'s terse-technical voice; insertion uses the
  bolded-label bullet shape (`**Required:** …`) of surrounding
  sections.
- **No emojis** in any new artifact (user global pref + repo
  emoji-free convention).
- **Link asymmetry is deliberate:** BRAG.md uses relative
  paths (in-repo); slash-command template uses absolute URLs
  (copied out of repo).
- **Re-run the full harness after each artifact added** — not
  at the end. Catches copy-paste contamination early.

### Reuse opportunities

- All four new groups reuse existing `assert_*` helpers from
  SPEC-021. Only `assert_jq_eq` is new.
- `jq` covers both the harness (group H) and the hook script —
  one install, two consumers.

### What NOT to do

See § Out of scope (above) and § Rejected alternatives (below)
for the comprehensive list — every locked rejection is
enumerated there. Three additional mechanical reminders not
covered elsewhere:

- Don't bump any `go.mod` line. No Go dep changes.
- Don't reformat unrelated whitespace in `BRAG.md` — the
  insertion is purely additive between two existing headers;
  `git diff` should show the new section only.
- Don't change the existing `test-docs` recipe in `justfile`
  (Q2 lock — in-place extension only).

### Rejected alternatives (build-time)

Per SPEC-018+ discipline, alternatives explicitly rejected at
design time so they don't slip into Deviations later:

1. **A `brag install-claude-hook` (or `brag install-slash-command`)
   CLI subcommand.** Rejected: Q3 STAGE-005 framing locked
   example + docs only; this is a personal-tool repo where
   users copy artifacts manually. A `brag install-*` surface
   would imply ongoing maintenance of "what does Claude Code
   expect this week?" inside the binary — premature for
   PROJ-001's narrow scope.

2. **A glossary entry in `AGENTS.md §11` for the three new
   terms.** Rejected (Q1c): the glossary's discriminator is
   *project-specific* terms (aggregate, tap, Store, digest);
   "JSON Schema", "Claude Code hook", "slash command" are
   industry-standard.

3. **An `examples/README.md` describing the directory.**
   Rejected (Q1b): the BRAG.md cross-reference is the
   discoverability surface for AI agents; contributor-facing
   discoverability is `docs/development.md`'s "Where to find
   what" table. Third surface for the same content adds no
   marginal value.

4. **A new sibling `scripts/test-ai-integration.sh`** rather
   than extending `scripts/test-docs.sh`. Rejected (Q2):
   SPEC-021's docstring explicitly invites the in-place
   extension; siblings create a fan-out problem
   (`scripts/test-completions.sh` for SPEC-024, then
   `scripts/test-distribution.sh` for SPEC-023).

5. **JSON Schema draft-07** rather than draft 2020-12.
   Rejected (Q3): 2020-12 is the current standard with
   adequate validator support (Ajv ≥7, jsonschema). Personal-
   tool repo with no schema-consumer ecosystem to consider.

6. **A Go-test using `gojsonschema` (or any JSON Schema
   validator library — Python `jsonschema`, `ajv-cli`, etc.)
   to validate fixture entries against the schema.** Rejected
   (Q4): adds a tooling dep (Go dep needs a DEC; Python /
   Node deps add CI complexity); the binary's
   `internal/cli/add_test.go` already has round-trip + parser
   tests that prove behaviour. Schema is documentation, not
   runtime validation.

7. **Updating `docs/tutorial.md`, `docs/api-contract.md`, and
   `docs/data-model.md` to cross-link the new schema.**
   Rejected: doc-sweep is SPEC-023's territory per SPEC-021's
   inherited punch list. Current `docs/tutorial.md:121` is an
   example list that stays accurate; existing DEC-012
   references at `docs/api-contract.md:416` and
   `docs/data-model.md:148` describe binary behavior and are
   not invalidated by the schema's existence.

8. **Reading `CLAUDE_PROJECT_DIR`, `CLAUDECODE`, or other env
   vars in the hook script.** Rejected (Q5b): pure stdin per
   stage scope guardrail. User wires the hook into their
   Claude Code config manually.

9. **Auto-executing `brag add` from the hook script** (rather
   than emitting the candidate JSON for the user to pipe).
   Rejected: BRAG.md's approval loop is non-negotiable
   ("Never post without approval"); the hook script must
   honour it. Auto-execution would be a footgun where a
   not-brag-worthy session-end summary becomes a noisy entry.

10. **Verbose slash-command prompt** (multi-paragraph,
    field-by-field guidance). Rejected (Q5a): tight prompts
    are easier to maintain and AI agents handle terse prompts
    well — the schema cross-link does the field-by-field
    heavy lifting.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to
verify.*

- **Branch:** `feat/spec-022-ai-integration-distribution-asset`
- **PR (if applicable):** opened post-`just advance-cycle` (URL
  added at PR-create time; see ship session).
- **All acceptance criteria met?** yes — 23/23 scripted asserts
  (groups H/I/J/K) pass; 7/7 group L byte-identity invariants
  hold (only modified files: BRAG.md, scripts/test-docs.sh; only
  new files: docs/brag-entry.schema.json,
  scripts/claude-code-post-session.sh,
  examples/brag-slash-command.md); 4/4 group M Go-sanity gates
  green (`go test ./...` cached pass, `gofmt -l .` empty, `go vet
  ./...` empty, `just test-docs` exits 0 with 63 OK lines).
  Pre-existing 40 asserts from groups A–G all still pass — zero
  regression on SPEC-021's harness.
- **New decisions emitted:** NONE — DEC-011 / DEC-012 / DEC-004
  referenced, not new (as expected per consume-only spec).
- **Deviations from spec:** none. All five literal artifacts
  (JSON schema + shell hook + slash-command markdown + BRAG.md
  insertion + test-docs.sh extension) byte-transcribed verbatim
  from the spec's § Notes for the Implementer. The BRAG.md
  insertion went between the existing `---` separator after
  "## The command" and the existing "## Three good examples"
  header, exactly as the spec specified. No deviation from any
  of the 10 locked rejections (no `brag install-claude-hook`
  CLI; no env-var introspection; no `~/.claude/` path lookups;
  no Go schema validator; no draft-07; no doc-sweep; no
  `examples/README.md`; no auto-execute; no verbose prompt; no
  sibling test script). Build-side premise audit greps 1–4
  reconciled to expected `## Outputs` deltas with zero stray
  references in production surface.
- **Follow-up work identified:**
  - SPEC-023 (distribution proper) inherits the deferred
    doc-sweep punch list: cross-link the schema from
    `docs/api-contract.md:416` and `docs/data-model.md:148`;
    point `docs/tutorial.md:121`'s "Claude session-end hook"
    example at the new `scripts/claude-code-post-session.sh`.
    Already enumerated in this spec's § Out of scope.
  - SPEC-024 (shell completions) is the third application of
    the §12 literal-artifact-as-spec pattern — completions
    output is fixed-shape per shell flavor and a strong
    candidate.

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec
create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing slowed the build. The five literal artifacts were
   complete enough that build was pure byte-transcription
   followed by harness-driven verification. One micro-friction:
   the spec literal for the BRAG.md insertion did not include
   the surrounding `---` HR separator that BRAG.md uses between
   sections, so I had to decide whether to add one. The right
   call was to follow byte-transcription discipline (no `---`
   added; the new section directly precedes "## Three good
   examples" without an intervening `---`). The result is one
   tiny stylistic asymmetry vs. the rest of BRAG.md, but the
   spec's locked literal is the source of truth — exactly the
   trade-off the §12 pattern asks for.

2. **Was there a constraint or decision that should have been
   listed but wasn't?**
   — No. The Q1–Q5 lock set + 10 rejected alternatives covered
   every decision the build encountered. The §10
   push-discipline reminder was already explicit. The §9
   BSD-grep `--exclude-dir` warning was a no-op (the new groups
   H/I/J/K use `assert_contains_literal` and inline `grep -E
   -q` — no `-r` calls, no `--exclude-dir` use). The
   NOT-contains self-audit was a no-op (zero negative-shape
   asserts in groups H–K, matching the spec's § NOT-contains
   self-audit n/a result).

3. **If you did this task again, what would you do differently?**
   — Nothing structurally. The build went linearly through the
   spec's 14-step build-cycle order with no surprises and no
   backtracking. **Q1 cross-format verdict:** the §12
   literal-artifact-as-spec pattern transferred cleanly across
   THREE formats simultaneously (JSON schema + bash script +
   markdown template) plus a sub-format insertion (markdown
   into existing markdown) plus a script extension (shell into
   existing shell). All five literals byte-transcribed cleanly
   on first pass; the harness's fail-first → write-artifact →
   pass-confirm rhythm worked identically for each format. The
   pattern is not markdown-specific; it generalizes to any
   fixed-shape artifact whose content is decidable at design
   time. SPEC-024 should adopt it for shell completions
   without hesitation.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection,
distinct from the process-focused build reflection above. NOTE:
`scripts/archive-spec.sh` rejects empty `<answer>` placeholders
(commit `bfa1474`, 2026-04-25) — these MUST be filled with real
answers at ship time or `just archive-spec SPEC-022` will fail.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>

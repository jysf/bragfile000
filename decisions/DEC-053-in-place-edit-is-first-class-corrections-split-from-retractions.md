---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-053
  type: decision
  confidence: 0.80
  audience:
    - developer
    - agent
    - operator

agent:
  id: claude-opus-4-8
  session_id: null

project:
  id: PROJ-008
repo:
  id: bragfile

created_at: 2026-09-14
supersedes: null
superseded_by: null

tags:
  - mcp
  - edit
  - capture
  - honest-corpus
---

# DEC-053: in-place edit is a first-class primitive; fixups split from retractions; delete stays human-gated

## Decision

In-place edit is a **first-class, scriptable primitive on both surfaces** — a
non-interactive CLI edit (`brag edit --set field=value…` and/or `--json`, a
**partial** update) and a `brag_edit` **MCP tool** — operating on **any entry by
id**, bumping `updated_at` with **no shadow-history in MVP**. It is the mechanism
for **fixups** (the record never matched what happened). It is kept **distinct
from `brag link`/supersede** (a *retraction* — a claim that was true and later
was not, worth narrating in `story`) and from **delete**, which stays
**human-gated at the CLI** and is NOT added to MCP.

## Context

Two rounds of heavy-agent field feedback (`~/ContextCore`, 450-entry corpus) plus
a maintainer question converged on the same gap. Low-friction capture is the
whole point of agent-facing `brag_add` and the plugin's capture nudge — the
maintainer does **not** pre-approve each brag — so agents routinely land entries
that are *wrong in some way*: a mis-framed `impact` (a restatement rather than the
value — the exact failure already recorded in memory), the wrong `project`, bad or
missing tags, a typo. The loop that keeps such a corpus honest is **review recent
auto-brags and fix them in place**, cheaply, on the surface the agent wrote
through.

An earlier draft of this project's roadmap positioned `brag link`/supersede
(§3.2) as the agent-correction path. That conflated two different things. The
honest-corpus thesis is about **the corpus telling the truth about the work** — it
is *not* about byte-immutability of every entry. Preserving a typo forever with a
"Correction to #427" beside it does not make the corpus more honest; it makes
`memory`, `export` and `story` carry the typo *and* its correction. Supersede is
right for *retractions*, wrong for *fixups*.

Today `brag edit` mutates in place through `Store.Update` (with `updated_at`
bumped, no history kept) but only via `$EDITOR`; there is no flag/`--json` mode
(§3.4) and no edit tool over MCP. `internal/capture.ValidateChanged(old, new)`
(DEC-046) already validates only changed fields and strips reserved provenance
tags on the edit path — i.e. the partial-update engine already exists.

## Alternatives Considered

- **Option A: supersede/link only — never mutate.**
  - What it is: corrections are always a new, linked entry (`brag link corrects`).
  - Why rejected: turns every typo and mis-framing into a permanent correction
    pair; pollutes the read surfaces `memory`/`story`/`export`; immutability of a
    wrong record is not what "honest corpus" means. Retraction-history is
    meaningful; typo-history is noise.

- **Option B: in-place edit on the CLI only, no MCP tool.**
  - What it is: add `brag edit --set`/`--json`; leave the MCP surface create+read.
  - Why rejected: the surface an agent writes through cannot fix what it wrote,
    forcing a shell-out to an interactive editor or an `$EDITOR`-shim — the exact
    friction observed twice (a reporter, then an agent that could not find how to
    edit at all).

- **Option C: full mutation over MCP, including delete.**
  - What it is: `brag_edit` and `brag_delete` both on the automated surface.
  - Why rejected for delete: hard-deleting corpus rows from an unattended agent is
    the genuinely destructive case; CLI `delete` already gates it with a y/N
    prompt. Edit is recoverable-in-spirit (you can edit again); delete is not.

- **Option D (chosen): in-place edit first-class on CLI + MCP; supersede separate; delete human-gated.**
  - What it is: the Decision above.
  - Why selected: matches how the corpus actually gets corrected (fixups dominate,
    retractions are rare), reuses the existing `ValidateChanged`/`Store.Update`
    engine, and closes the agent capture-then-fix loop without laundering history
    (`updated_at` moves) or handing an agent the destructive verb.

## Consequences

- **Positive:** low-friction, unapproved agent capture becomes *safe to lean on*,
  because mistakes are cheap to fix on the same surface; the CLI gains the
  scriptable edit (§3.4) it has lacked; mostly surface work — the engine exists.
- **Negative:** in-place edit keeps **no pre-edit value** in MVP (only `updated_at`
  moves), so a substantive rewrite leaves no audit trail; and **any-by-id** means
  an agent can edit any entry, including a human's or another agent's — a trust
  surface accepted for parity with CLI `brag edit`, not because it is risk-free.
- **Neutral:** the corpus now has **two** correction mechanisms (in-place fixup vs.
  superseding retraction); they must be documented so a user/agent picks the right
  one — the STAGE-027 specs and SPEC-091's `--help` work carry that.

## Validation

Right if the agent capture-fix loop measurably closes (auto-brags get corrected in
place rather than accumulating errors or spawning correction-pairs). Revisit if:
(a) silent history loss on a *substantive* edit ever matters → add shadow-history
for material fields; (b) any-by-id proves too permissive → tighten to
own-entries-only via the `agent:` tag. Both are additive and do not block the MVP.

## References

- Related stage: STAGE-027 (the correctable corpus — this decision's frame)
- Related decisions: DEC-046 (ValidateChanged — the partial-update engine), DEC-024 (MCP surface framed as write/read — this extends "write" from create to create+update), DEC-052 (`brag edit` emits the mutated id on stdout), DEC-009 (editor buffer format), DEC-017 (`entries.project` free-text)
- Related specs: SPEC-089 (in-place editor edit + dup-header reject), SPEC-091 (documents the two correction mechanisms in `--help`)
- Discussions: round-1/round-2 `~/ContextCore` field feedback (§3.2 supersede, §3.4 edit flag mode, §4.3 CLI/MCP asymmetry); maintainer direction 2026-09-14 ("sometimes you need to edit and a link is not the right mechanism … an agent often gets the brag wrong")

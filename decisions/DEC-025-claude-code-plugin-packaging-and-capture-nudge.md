---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-025
  type: decision
  confidence: 0.8                     # honest: the manifest/marketplace layout
                                       # is high-confidence (validated strict
                                       # against the real loader); the residual
                                       # soft spot is the live delivery of the
                                       # Stop-hook additionalContext, which
                                       # design could not exercise end-to-end
                                       # (build does a §12(b) live check) — see
                                       # Validation.
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-003
repo:
  id: bragfile

created_at: 2026-07-04
supersedes: null
superseded_by: null

tags:
  - plugin
  - claude-code
  - packaging
  - hooks
  - capture-nudge
  - stdout-stderr-spine
  - agent-native
---

# DEC-025: bragfile ships as a Claude Code plugin — layout, MCP-on-PATH, and the capture-nudge delivery model

## Decision

bragfile ships as **one installable Claude Code plugin** that bundles the three
STAGE-009 agent surfaces. Three sub-decisions:

1. **Layout.** The plugin lives at repo-root **`plugin/`** with its manifest at
   `plugin/.claude-plugin/plugin.json`, auto-discovered `commands/brag.md` and
   `hooks/hooks.json`, and a `plugin/README.md`. A same-repo marketplace at
   repo-root **`.claude-plugin/marketplace.json`** lists the one plugin with
   `source: "./plugin"`. Both manifests were **validated strict against the
   real loader** (`claude plugin validate --strict`, Claude Code 2.1.201) at
   design — the §12(b) pre-flight (STAGE-009 Design Note (c)). Install path:
   `claude plugin marketplace add jysf/bragfile000` →
   `claude plugin install brag@bragfile`.

2. **MCP server entry = `{command:"brag", args:["mcp","serve"]}`, declared
   via `plugin/.mcp.json` (bare server-name→config map at the plugin root)
   — the inline `mcpServers` key inside `plugin/.claude-plugin/plugin.json`
   is insufficient for registration; it is kept only for parity with the
   `plugin.json` schema shape (see Amendment below).** The plugin's
   `.mcp.json` `brag` entry runs the released `brag` binary from `PATH`
   (prerequisite: `brew install jysf/bragfile/bragfile`), not a bundled
   binary or a wrapper script. `brag` is already a Homebrew-distributed
   binary, so a direct entry needs no packaged executable and reuses the one
   install. The loader tolerates a bare (non-`${CLAUDE_PLUGIN_ROOT}`) command
   string as long as the binary resolves on `PATH` at launch — confirmed
   behaviorally (`claude plugin details brag` → `MCP servers (1) brag`), not
   assumed; no launcher/shim script is needed.

3. **The capture-nudge Stop hook: once-per-session, commit-gated,
   agent-facing, never-posts, env-silenceable.** The hook
   (`plugin/hooks/capture-nudge.sh`, evolved from SPEC-022's
   `scripts/claude-code-post-session.sh`) fires on **`Stop`** but nudges **at
   most once per session** and **only after a git commit lands in the cwd
   during the session** (baseline HEAD recorded on the first Stop, compared on
   later Stops). On a fire it emits a single
   `hookSpecificOutput.additionalContext` JSON object — **agent-facing**
   context that prompts Claude to propose a brag through BRAG.md's approval
   loop — and **never invokes `brag`**. `BRAG_CAPTURE_NUDGE=off` silences it.
   All fire/silence paths and the never-invokes-`brag` guarantee were exercised
   at design against crafted Stop payloads.

The reserved `agent:<name>`/`model:<id>` provenance convention (DEC-024) is
documented in the shipped `plugin/hooks/capture-nudge.sh`,
`plugin/commands/brag.md`, `plugin/README.md`, and `BRAG.md`.

## Context

STAGE-009 surfaced two design questions this DEC answers: **(c)** the plugin
manifest shape (pre-flight against the current loader, do not transcribe from
memory) and **(d)** the Stop/session-end capture-nudge UX (quiet, skippable,
non-nagging, never auto-posts, degrades on non-TTY; and the explicit
*"fire every session vs only when the session plausibly shipped"* question).

**Two findings from the design-time pre-flight reshaped question (d):**

- **`Stop` fires on every agent turn, not once per session** (confirmed
  against the Claude Code 2.1.201 hooks reference). An ungated nudge would nag
  every turn. So the fire model *must* gate: once per session (a per-session
  marker keyed by `session_id`) + a "plausibly shipped" signal. The signal
  chosen is **"HEAD advanced during the session"** — a strong, cheap proxy for
  "you shipped something," needing only the marker to also store the
  first-seen HEAD. This is the explicit answer to question (d)'s "fire every
  session vs only when plausibly shipped": **only when plausibly shipped.**

- **The Stop-hook surface has no TTY, and its output feeds *Claude*, not the
  user.** The stage note framed the hook as "quiet, TTY-only, degrades on
  non-TTY" — but hooks run headless with JSON on stdin, and a Stop hook's
  `additionalContext` is delivered to the model, not printed to a terminal.
  So "quiet + degrades" is realized by the **fire-gating** (silent unless a
  commit landed, once per session, env-silenceable), not a TTY probe. This is
  a §12(b) *contract discovery* — the assumed surface shape was wrong; the
  real contract differs — exactly the class of thing design-time pre-flight
  exists to surface, and it happens to align better with BRAG.md's model
  (the agent proposes; the user approves).

`SessionEnd` was the intuitive "session-end" event, but its output is not
user- or agent-visible (the session is tearing down), so it cannot deliver a
nudge — `Stop` (gated) is the correct event.

## Alternatives Considered

- **Option A: repo-root *is* the plugin** (`.claude-plugin/plugin.json` at
  root, `commands/`/`hooks/` at root). Rejected: pollutes the app repo root and
  conflates plugin assets with Go source; a dedicated `plugin/` subdir keeps
  the plugin self-contained. Both validate, so this is ergonomics, not
  correctness.

- **Option B: bundle a `brag` binary or a wrapper script in the plugin**
  (à la `formae/scripts/start-mcp.sh`). Rejected: formae ships its binary
  inside the plugin; `brag` is already a released Homebrew binary on `PATH`, so
  a direct `command:"brag"` entry avoids shipping/updating a second copy. The
  PATH prerequisite is documented.

- **Option C: nudge on every `Stop` (ungated).** Rejected: `Stop` fires every
  turn → a nag, the exact failure mode question (d) forbids.

- **Option D: nudge on `SessionEnd`.** Rejected: `SessionEnd` output is not
  surfaced (session ending) — the nudge would go nowhere.

- **Option E: block via exit 2 to force Claude to keep working / surface the
  nudge.** Rejected: exit-2 blocks the stop, forcing continuation every
  session — a nag, and it fights the user's intent to end the session.

- **Option F: user-facing stderr line (TTY-gated, mirroring the SPEC-039
  milestone line).** Rejected: the Stop-hook surface has no TTY, and stderr on
  exit 0 is not reliably surfaced to the user; the agent-facing
  `additionalContext` is the contract-correct channel and matches BRAG.md's
  "agent proposes" model.

- **Option G (chosen): `plugin/` + root marketplace; MCP-on-PATH; Stop hook,
  once-per-session, commit-gated, agent-facing `additionalContext`,
  never-posts, `BRAG_CAPTURE_NUDGE=off`.** Selected: manifests validate strict
  against the real loader; one binary, one install; the hook is a nudge not a
  nag and provably never auto-posts.

## Consequences

- **Positive:** the integration story becomes "install one plugin" instead of
  "copy three files"; the MCP tools, `/brag:brag`, and the capture nudge load
  together; the nudge fires only when a session plausibly shipped, at most
  once, and can be silenced; the approval loop is provably intact (the hook
  never runs `brag`). Manifests are strict-validated, so `claude plugin
  install` / `claude plugin tag` / marketplace CI will not choke on shape drift.
- **Negative:** `commands/brag.md` in a plugin named `brag` invokes as
  `/brag:brag`, not a clean `/brag` (plugin commands are always namespaced);
  the bare `/brag` remains available only via the manual-copy path. Documented.
- **Negative / accepted debt:** the "plausibly shipped" gate is a *commit
  landed mid-session* heuristic — it misses shipped-but-uncommitted work and
  fires at most once even across a long multi-ship session. Acceptable for a
  best-effort nudge; a richer signal is a later refinement if dogfooding asks.
- **Neutral:** the loose SPEC-022 assets stay (manual path + convention proof),
  so `test-docs.sh` groups I/J are untouched. "Evolving
  `claude-code-post-session.sh`" means *derivation*, not deletion — the Stop
  hook is a new contract inspired by it.
- **Residual uncertainty (drives confidence 0.8):** design validated the
  *manifests* against the real loader and the *hook behavior* against crafted
  payloads, but could not exercise a **live** Claude Code session honouring the
  Stop-hook `additionalContext`. Build performs a §12(b) live check (install
  in a scratch dir, `claude plugin details brag`, confirm load) and raises any
  contract delta as a finding.

## Validation

Right if: `claude plugin validate --strict` passes for both manifests (green
at design); the `capture-nudge.sh` behavior harness passes all fire/silence
paths and the PATH-stub sentinel proves the hook never invokes `brag`; the
plugin installs via `claude plugin install brag@bragfile` and
`claude plugin details brag` lists the MCP server + `/brag:brag` command +
Stop hook (confirmed 2026-07-04 in a scratch marketplace install, post
Amendment below — `MCP servers (1) brag`, not `(0)`); and the shipped
hook/command/README/BRAG.md all carry the reserved `agent:`/`model:`
convention.

Revisit if: (a) the "commit-landed" gate proves too eager or too quiet in
dogfooding → refine the shipped signal (e.g. commit *count*, or diff size, or
a SessionStart baseline); (b) the live `additionalContext` delivery differs
from the documented Stop-hook contract → adjust the delivery channel and
re-record here; (c) the Claude Code plugin manifest schema changes under us
(external, moving loader — the §12(b) reason this was pre-flighted) → re-run
`validate --strict` and update the literals; (d) a bare `/brag` becomes a real
ask → consider a differently-named plugin or a shipped command alias.

## Amendment (2026-07-04, SPEC-041 build punch-list)

**Correction to sub-decision 2's manifest shape:** the original literal
declared the MCP server *only* via the inline `mcpServers` key inside
`plugin/.claude-plugin/plugin.json`. `claude plugin validate --strict`
passed against that shape (it validates JSON structure, not runtime
registration), and SPEC-041's own AC3 test (S4, a `jq` structural assertion)
also passed against it — but a scratch-marketplace install showed
`claude plugin details brag` → **`MCP servers (0)`**, confirmed as a real bug
in SPEC-041's Build Completion (Deviation 2) and re-confirmed at this
amendment via a clean before/after scratch install (before: `.mcp.json`
absent → `MCP servers (0)`; after: `.mcp.json` added → `MCP servers (1) brag`).

**Root cause:** Claude Code's plugin loader registers MCP servers from a
separate **`plugin/.mcp.json`** file at the plugin root (a bare
`{"<name>": {"command": ..., "args": ...}}` map) — not from the
`mcpServers` key inside `plugin.json`. Every working reference plugin
inspected (`formae-mcp`) ships a `.mcp.json`; ours did not.

**Fix:** added `plugin/.mcp.json` — `{"brag": {"command": "brag", "args":
["mcp", "serve"]}}` — as the actual registration source. The inline
`mcpServers` key in `plugin.json` is **kept** (does not block registration,
costs nothing, and matches the `formae-mcp` reference plugin, which ships
both `.mcp.json` and an equivalent inline `plugin.json` key) but is
documented here as **non-authoritative for registration**.

**PATH dependency, stated explicitly:** `.mcp.json`'s `command:"brag"` is a
bare command, not `${CLAUDE_PLUGIN_ROOT}`-qualified — it resolves via the
user's `PATH` at MCP-server launch time, same as the inline key always
intended. This requires `brag` to already be installed (`brew install
jysf/bragfile/bragfile`, per `plugin/README.md`'s Prerequisite section) —
the plugin does not bundle or vendor the binary. Confirmed behaviorally that
the loader accepts a bare command here (no launcher/shim script needed).

**Process note for the §12(b) family (flagged WATCH, not codified — N=1):**
the design-time pre-flight ran `claude plugin validate --strict` (the
loader's manifest validator) but not `claude plugin details` (the loader's
*registration* surface) against a real install. Validation passing is
necessary but not sufficient evidence that a manifest's declared components
actually load. This refines the existing §12(b) "run the literal through its
target tool" rule: for a plugin manifest, the target tool for the
MCP-registration claim specifically is `claude plugin details`, not
`validate --strict` — they check different things. One instance so far;
AGENTS.md's own codification meta-rule wants N=2 (same-outcome) or a
paired opposing-outcome N=2 before this earns a codified rule change, so
this is recorded here as a live instance to watch, not folded into AGENTS.md
yet.

**Regression guard:** `scripts/test-docs.sh` group S gained `S12`/`S12-jq`,
asserting `plugin/.mcp.json` exists and declares
`{"brag": {"command": "brag", "args": ["mcp", "serve"]}}` — a cheap guard
against this specific shape silently regressing (it cannot, by itself,
re-prove runtime registration; that still requires the behavioral
`claude plugin details` check performed at this amendment).

## References

- Related specs: SPEC-041 (emits + implements this DEC — the plugin packaging),
  SPEC-042 (the v0.3.0 release cut that tags the release including the plugin),
  SPEC-040 (the `brag mcp serve` the plugin points at), SPEC-039 (the CLI
  milestone line — the TTY-gated mirror surface), SPEC-022 (the convention
  assets the plugin packages)
- Related decisions: DEC-024 (the reserved `agent:`/`model:` provenance
  convention the shipped assets document), DEC-015 (polymorphic tags —
  provenance rides these), DEC-011 (the entry JSON shape), DEC-006 (cobra —
  `brag mcp serve` is the subcommand the MCP entry runs)
- Related constraints: `stdout-is-for-data-stderr-is-for-humans` (blocking —
  the hook's stdout is the hook protocol; only the `hookSpecificOutput` object
  or nothing), `one-spec-per-pr` (blocking — the release cut is peeled to
  SPEC-042)
- External: Claude Code 2.1.201 plugin system — `claude plugin validate
  --strict`, `claude plugin install`, `claude plugin details`, `claude plugin
  tag`; the hooks reference (`Stop` fires per turn; `hookSpecificOutput.
  additionalContext`; `${CLAUDE_PLUGIN_ROOT}`); the `.claude-plugin/
  plugin.json` + `.claude-plugin/marketplace.json` manifest shapes
- Discussions: STAGE-009 Design Notes surfaced questions (c) + (d); PROJ-003
  brief Scope ("Claude Code plugin packaging")

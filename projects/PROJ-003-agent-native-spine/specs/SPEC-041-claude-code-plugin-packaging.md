---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-041
  type: story                      # epic | story | task | bug | chore
  cycle: build
  blocked: false
  priority: high
  complexity: M                    # S | M | L  (L means split it) — see L-watch outcome

project:
  id: PROJ-003
  stage: STAGE-009
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # same Claude, different session (claude-only variant)
  created_at: 2026-07-04

references:
  decisions: [DEC-025, DEC-024, DEC-015, DEC-011]
  constraints: [stdout-is-for-data-stderr-is-for-humans, one-spec-per-pr, no-sql-in-cli-layer]
  related_specs: [SPEC-040, SPEC-039, SPEC-022, SPEC-042]
---

# SPEC-041: Claude Code plugin packaging

## Context

STAGE-009 shipped the agent-native write spine as three loose surfaces:
`brag mcp serve` (SPEC-040, the MCP server + reserved-tag provenance), the
`/brag` slash-command draft (SPEC-022, `examples/brag-slash-command.md`), and
a session-capture helper (SPEC-022, `scripts/claude-code-post-session.sh`).
Today a user wires those in by hand — copy the slash-command into
`~/.claude/commands/`, copy the hook wherever their config wants it, and
hand-configure the MCP server. This spec closes STAGE-009's fourth item: it
**bundles all three into one installable Claude Code plugin**, so the
integration story becomes *"install this plugin"* rather than *"copy these
files."* It also documents the reserved `agent:`/`model:` provenance
convention (DEC-024) in the shipped assets, and folds in one MCP-surface
hardening test (below).

This is the **last committed item in STAGE-009**; shipping it plus the
v0.3.0 release cut closes the stage. Per the stage's PEEL-IF-L discipline
(Spec Backlog "Complexity check"), **the v0.3.0 release cut has been peeled
into its own spec, SPEC-042** — this spec is plugin packaging only. See
[L-watch outcome](#l-watch-outcome-peel-decided-at-design) for the rationale,
decided explicitly at design, not discovered at build.

- Parent: `STAGE-009` (Agent-native spine + capture delight, v0.3.0 core).
- Project: `PROJ-003` (agent-native spine).
- Emits: `DEC-025` (plugin packaging layout + capture-nudge delivery model).
- Depends on: `SPEC-040` (the `brag mcp serve` the plugin points at) and
  `DEC-024` (the provenance convention the shipped assets document).
- Enables / hands off to: `SPEC-042` (the v0.3.0 release cut, which tags the
  release *after* this plugin PR merges to `main`).

## Goal

Bundle the shipped `brag mcp serve`, the `/brag` slash-command, and a quiet
session-capture-nudge Stop hook into one installable Claude Code plugin
(manifest pre-flighted against the real loader), documenting the reserved
`agent:`/`model:` provenance convention in the shipped assets, and add one
regression-guard test asserting `brag_add`'s returned JSON is byte-identical
to `internal/export`'s per-entry rendering of the created row.

## Inputs

- **Files to read:**
  - `examples/brag-slash-command.md` — the slash-command draft the plugin's
    `commands/brag.md` transcribes/refines.
  - `scripts/claude-code-post-session.sh` — the SPEC-022 capture helper the
    plugin's Stop hook evolves *from* (the hook is a new contract; see below).
  - `BRAG.md` — the integration guide that must gain the plugin path + the
    provenance convention.
  - `internal/mcpserver/server.go` (`marshalEntry`/`entryRecord`) and
    `internal/export/json.go` (`ToJSON`/`toEntryRecord`) — for the
    regression-guard test.
  - `internal/mcpserver/server_test.go` — where the regression-guard test lands.
  - `scripts/test-docs.sh` — the shell-assertion harness the plugin docs-tests
    extend (add **group K**).
  - `decisions/DEC-024-...md` — the provenance convention to document.
- **External tools (the §12(b) loaders — already RUN at design, see below):**
  - `claude plugin validate --strict <path>` (Claude Code 2.1.201) — the real
    plugin/marketplace-manifest loader.
- **Related code paths:** `internal/mcpserver/`, `internal/export/`.

## §12(b) design-time pre-flight (RUN at design — results below, all green)

Per STAGE-009 Design Notes (c) and AGENTS.md §12(b), the plugin manifest's
shape depends on an **external, moving loader** (the Claude Code plugin
system), so the manifest was **not transcribed from memory** — it was built
and run through the *real* loader at design. The environment ships
`claude` 2.1.201, whose `claude plugin validate --strict <path>` **is** the
loader's own manifest validator (the goreleaser-check / GenBashCompletion
analogue for this artifact). All literals below were validated before locking.

**Manifest shape learned from real installed manifests** (not memory):
inspected `~/.claude/plugins/**/.claude-plugin/plugin.json` for the
`formae-mcp` (MCP-server), `hookify` (Stop-hook), and `code-simplifier`
plugins, plus the official `.claude-plugin/marketplace.json`. Findings that
shaped the literals:

- Manifest lives at `<plugin-root>/.claude-plugin/plugin.json`. Fields:
  `name` (required), `description`, `version`, `author:{name,email?}`,
  `homepage`, `repository`, `license`, `keywords:[…]`, and `mcpServers`
  (object: name → `{command, args?, env?}`).
- Components are **auto-discovered by directory convention**: `commands/*.md`
  (slash-commands), `hooks/hooks.json` (hooks), and `mcpServers` in the
  manifest (MCP). No component needs a manifest declaration beyond `mcpServers`.
- Hooks JSON shape (from `hookify/hooks/hooks.json`): `{description,
  hooks:{<Event>:[{hooks:[{type:"command",command:"…",timeout:N}]}]}}`.
  `${CLAUDE_PLUGIN_ROOT}` expands to the install dir inside command strings.
- Same-repo marketplace: `<repo-root>/.claude-plugin/marketplace.json` with
  `{name, description, owner:{name}, plugins:[{name, source:"./plugin",
  description, category}]}`; `source` is a repo-relative path to the plugin
  root.

**Pre-flight results (all commands run at design, in `<scratchpad>/preflight`):**

| Artifact | Command | Result |
|---|---|---|
| `plugin/.claude-plugin/plugin.json` | `claude plugin validate --strict plugin` | **✔ Validation passed** |
| `.claude-plugin/marketplace.json` | `claude plugin validate --strict .` | **✔ Validation passed** |

**§12(b) catch — a real strict-mode drift the pre-flight surfaced:** the
first marketplace literal (no `description` key) **failed** `--strict`
(`No marketplace description provided … Validation failed (--strict treats
warnings as errors)`). The embedded literal below carries the added
top-level `description` that makes strict pass. This is precisely the
key-shape drift §12(b) exists to catch before build — mirroring SPEC-023's
`brews:`→`homebrew_casks:` and SPEC-024's cobra bash-marker.

**Hook behavior pre-flight (RUN at design):** the `capture-nudge.sh` literal
below was exercised against crafted Stop payloads in a throwaway git repo
across all fire/silence paths — first-Stop-baseline (silent), no-new-commit
(silent), commit-landed (emits valid `hookSpecificOutput` JSON, verified with
`jq -e`), post-nudge second-fire (silent, once-per-session), env-silenced
(silent), non-git-cwd (silent) — and a PATH-shadowing `brag` stub confirmed
the hook **never invokes `brag`** on any path (approval loop held). All green.

## Outputs

### Files created

- `plugin/.claude-plugin/plugin.json` — the plugin manifest (literal below).
- `plugin/commands/brag.md` — the `/brag:brag` slash-command (literal below).
- `plugin/hooks/hooks.json` — the Stop-hook wiring (literal below).
- `plugin/hooks/capture-nudge.sh` — the capture-nudge hook (literal below, `+x`).
- `plugin/README.md` — plugin install + prerequisites + provenance note (literal below).
- `.claude-plugin/marketplace.json` — same-repo marketplace listing (literal below).

### Files modified

- `BRAG.md` — add a "Install as a Claude Code plugin (recommended)" subsection
  and a "Provenance: reserved `agent:`/`model:` tags" subsection (DEC-024).
  The existing SPEC-022 reference-asset lines (190, 196-199) **stay** and gain
  a cross-reference to the plugin path.
- `README.md` — add a one-line plugin-install pointer under the install
  section (the manual `just install`/`brew` paths stay).
- `internal/mcpserver/server_test.go` — add the regression-guard test
  `TestServer_AddReturnValueParity` (below). No production code changes.
- `scripts/test-docs.sh` — add **group K** (plugin assertions, below).

### Premise-audit enumeration (status-change: "copy these files" → "install this plugin")

Ran `grep -rn` for the SPEC-022 asset names across `BRAG.md`, `docs/`,
`README.md` at design (§9 audit-grep cross-check). Every hit enumerated as
**update** or **stays**:

| Hit | File:line | Disposition |
|---|---|---|
| `scripts/claude-code-post-session.sh` reference asset | `BRAG.md:190` | **stays** + cross-ref to plugin (the helper remains the transport-agnostic pipe helper; the plugin's Stop hook is a *different* contract) |
| `examples/brag-slash-command.md` template | `BRAG.md:196-199` | **stays** + note the packaged `/brag:brag` plugin command as the recommended path |
| both assets, blog-plan row | `docs/blog/README.md:17` | **stays** (a not-yet-drafted blog plan; the assets still exist) |
| both assets, security-review | `docs/reports/security/2026-04-26-*.md:118,141,282,283` | **stays** (historical report; the reviewed files are unchanged) |
| `brag mcp serve` + provenance | `docs/api-contract.md:665-744` | **stays** (SPEC-040 already documented the MCP surface + DEC-024 here; not re-touched) |
| `plugin` (keyword) | none in `BRAG.md`/`docs/`/`README.md` | new mentions added by this spec |

**Deferred doc touches (OUT of this spec, enumerated so they aren't lost):**
`docs/tutorial.md` and `docs/architecture.md` plugin walkthroughs — the
stage's broad doc-completeness goal — are **not** required for the plugin to
install and work, and are deferred to **SPEC-042**'s release doc sweep (which
already touches the tutorial/CHANGELOG at the v0.3.0 cut). BRAG.md + README +
`plugin/README.md` are the load-bearing integration-story docs this spec owns.

### No database / schema changes

None. The plugin is packaging; the MCP server, storage, and export layers are
unchanged except the added *test*.

## Acceptance Criteria

Each criterion pairs with a Failing Test (§9). Tests are shell (`test-docs.sh`
group K + a standalone hook-behavior harness) or Go, per the artifact.

- [ ] **AC1 — Manifest validates strict.** `plugin/.claude-plugin/plugin.json`
  exists and `claude plugin validate --strict plugin` exits 0. → K1, K2.
- [ ] **AC2 — Marketplace validates strict.** `.claude-plugin/marketplace.json`
  exists and `claude plugin validate --strict .` exits 0. → K3.
- [ ] **AC3 — MCP server is wired.** The manifest's `mcpServers.brag` runs
  `brag mcp serve` (command `brag`, args `["mcp","serve"]`). → K4.
- [ ] **AC4 — Slash-command bundled.** `plugin/commands/brag.md` exists, has a
  `description:` front-matter key, and instructs *not* to run `brag add`
  before approval. → K5, K6.
- [ ] **AC5 — Stop hook wired.** `plugin/hooks/hooks.json` registers a `Stop`
  hook whose command references `${CLAUDE_PLUGIN_ROOT}/hooks/capture-nudge.sh`. → K7.
- [ ] **AC6 — Hook is quiet + gated + never posts.** `capture-nudge.sh` is
  executable, silent on the first Stop / no-new-commit / non-git / env-silenced
  paths, emits exactly one valid `hookSpecificOutput` JSON nudge when a commit
  lands mid-session, stays silent on subsequent Stops, and **never invokes
  `brag`**. → H-series (hook harness).
- [ ] **AC7 — Provenance convention documented in shipped assets.** The
  reserved `agent:<name>`/`model:<id>` convention appears in
  `plugin/hooks/capture-nudge.sh`, `plugin/commands/brag.md` **or**
  `plugin/README.md`, and `BRAG.md`. → K8, K9, K10.
- [ ] **AC8 — `brag_add` return-value parity (regression guard).**
  `brag_add`'s returned JSON is byte-identical, field-for-field, to
  `export.ToJSON([]storage.Entry{created})`'s single element. → Go test.
- [ ] **AC9 — No regressions.** `go test ./...`, `gofmt -l .`, `go vet ./...`,
  `just test-docs` all clean.

## Failing Tests

Written at design, BEFORE build. Two are already validated green at design (the
manifest/marketplace validations and the hook-behavior harness were RUN in the
pre-flight); they are **preservation guards** — build transcribes the
pre-flighted literals and they pass. The Go regression-guard is likewise a
preservation guard: the code is already correct (the local `marshalEntry`
mirrors export field-for-field), so it passes immediately — like SPEC-038's ●
tests, not a fail-first case. They "fail" today only because the `plugin/`
files and the test do not exist yet.

### `scripts/test-docs.sh` — new **group K** (plugin)

Uses the harness's existing `assert_file_exists` / `assert_contains_literal` /
`ok` / `fail` helpers plus a new `assert_cmd_ok` for the validator. Guard the
validator behind `command -v claude` so CI without the CLI SKIPs (prints a
`skip` line) rather than failing — the manifest literals are also asserted
structurally with `jq` so coverage does not depend on `claude` being present.

- `K1` — `assert_file_exists "K1" "plugin/.claude-plugin/plugin.json"`.
- `K2` — if `command -v claude`: `claude plugin validate --strict plugin`
  exits 0 (else `skip`). Structural fallback (always runs):
  `jq -e '.name=="brag" and (.mcpServers.brag.command=="brag")'`.
- `K3` — `assert_file_exists "K3" ".claude-plugin/marketplace.json"`; if
  `command -v claude`: `claude plugin validate --strict .` exits 0 (else
  `skip`); structural fallback: `jq -e '.description and (.plugins[0].name=="brag") and (.plugins[0].source=="./plugin")'`.
- `K4` — `jq -e '.mcpServers.brag.args==["mcp","serve"]' plugin/.claude-plugin/plugin.json`.
- `K5` — `assert_file_exists "K5" "plugin/commands/brag.md"`.
- `K6` — `assert_contains_literal "K6" "plugin/commands/brag.md"` for the
  literal `Do not execute` (the approval gate) — and `brag add --json` present
  so the command names the tool it gates.
- `K7` — `jq -e '.hooks.Stop[0].hooks[0].command | test("CLAUDE_PLUGIN_ROOT.*capture-nudge.sh")' plugin/hooks/hooks.json`.
- `K8` — `assert_contains_literal "K8" "plugin/hooks/capture-nudge.sh" "agent:<name>"`
  and `model:<id>` (provenance convention in the shipped hook).
- `K9` — `assert_contains_literal "K9" "plugin/README.md" "brew install"` (the
  brag-on-PATH prerequisite) and `agent:` (provenance note).
- `K10` — `assert_contains_literal "K10" "BRAG.md" "plugin"` and
  `agent:<name>` (BRAG.md documents both the plugin path and the convention).
- `K11` — `plugin/hooks/capture-nudge.sh` is executable
  (`[ -x plugin/hooks/capture-nudge.sh ]`).

### Hook-behavior harness — `scripts/test-capture-nudge.sh` (new)

A standalone shell test (wired into `just test-docs` after the group-K block,
or a sibling recipe `just test-hook`), matching the `ok`/`fail` shape. Builds a
throwaway git repo + a temp `BRAG_STATE_DIR` + a PATH-shadowing `brag` stub
that records invocations to `$BRAG_SENTINEL`. Drives `capture-nudge.sh` with
crafted Stop JSON payloads:

- `H1` — first Stop (no marker): stdout empty (baseline recorded, silent).
- `H2` — second Stop, HEAD unchanged: stdout empty.
- `H3` — a commit lands, next Stop: stdout is valid JSON with
  `.hookSpecificOutput.hookEventName=="Stop"` and
  `.hookSpecificOutput.additionalContext` matching `/brag/`.
- `H4` — another Stop after the nudge: stdout empty (once per session).
- `H5` — `BRAG_CAPTURE_NUDGE=off` with a fresh commit: stdout empty.
- `H6` — cwd is not a git repo: stdout empty.
- `H7` — across all of the above, `$BRAG_SENTINEL` never exists → the hook
  never invoked `brag` (approval loop / never-auto-posts held).

### `internal/mcpserver/server_test.go` — regression guard (Go)

Closes the SPEC-040 verify advisory (its own Ship reflection Q1): the existing
tests assert `brag_add`'s *side effects* (`Store.List`) and `brag_list`
parity, but never assert `brag_add`'s **own returned JSON payload**. This adds
that assertion — a **preservation guard** (the shape is already correct by
construction). Pinned export call: **`export.ToJSON([]storage.Entry{created})`**
— the mcpserver's local `entryRecord`/`marshalEntry` mirrors export's private
`entryRecord`/`toEntryRecord` field-for-field; this test proves the mirror has
not drifted, at the MCP boundary (via `callJSON`), field-for-field byte-wise.

```go
// TestServer_AddReturnValueParity ● preservation guard: brag_add's RETURNED
// JSON is byte-identical, field-for-field, to export.ToJSON's rendering of
// the same created row. Closes the SPEC-040 verify advisory (only brag_add's
// side effects, not its literal return value, were asserted). Passes
// immediately — the local entryRecord mirrors export's field-for-field.
func TestServer_AddReturnValueParity(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	got := callJSON(t, cs, "brag_add", map[string]any{
		"title": "cut p99", "project": "bragfile", "type": "shipped",
		"tags": "perf", "impact": "-80% p99",
		"agent": "claude-code", "model": "claude-opus-4-8",
	})
	// The created row is the sole row in the store.
	rows, err := s.List(storage.ListFilter{})
	if err != nil || len(rows) != 1 {
		t.Fatalf("expected exactly one created row, got %d (err=%v)", len(rows), err)
	}
	// Pinned export call: the per-entry rendering brag_add's return mirrors.
	arr, err := export.ToJSON([]storage.Entry{rows[0]})
	if err != nil {
		t.Fatalf("export.ToJSON: %v", err)
	}
	// Compare field-for-field on the RAW value bytes, so array-wrapper and
	// indentation differences (object vs 1-element array) are irrelevant and
	// only the per-field serialization is asserted.
	var wantArr []map[string]json.RawMessage
	if err := json.Unmarshal(arr, &wantArr); err != nil || len(wantArr) != 1 {
		t.Fatalf("unmarshal export array: %v (len=%d)", err, len(wantArr))
	}
	var gotObj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(got), &gotObj); err != nil {
		t.Fatalf("unmarshal brag_add return: %v", err)
	}
	want := wantArr[0]
	if len(gotObj) != len(want) {
		t.Fatalf("field count: brag_add has %d keys, export has %d", len(gotObj), len(want))
	}
	for k, wv := range want {
		gv, ok := gotObj[k]
		if !ok {
			t.Errorf("brag_add return missing key %q present in export", k)
			continue
		}
		if string(gv) != string(wv) {
			t.Errorf("field %q not byte-identical: brag_add=%s export=%s", k, gv, wv)
		}
	}
}
```

Add `"encoding/json"` to `server_test.go`'s imports (it currently imports
`context`, `path/filepath`, `reflect`, `sort`, `testing`, `time`, plus
`export`, `storage`, `mcp`).

## Locked design decisions

Numbered so each pairs with a test (AGENTS.md §9 "every locked decision has a
failing test"). Rejected alternatives listed to keep them out of Deviations.

1. **Plugin lives at repo-root `plugin/`; marketplace at repo-root
   `.claude-plugin/marketplace.json`.** (→ K1, K3.) Rejected: repo-root *as*
   the plugin (pollutes root with `commands/`/`hooks/`, and the marketplace
   would need to point `source:"."` at a root that also holds Go code — the
   validator is happy but the layout conflates app and plugin). A dedicated
   `plugin/` subdir keeps the plugin self-contained; the marketplace's
   `source:"./plugin"` points at it. Both validate strict (pre-flighted).

2. **MCP server entry = `{command:"brag", args:["mcp","serve"]}`, requiring
   `brag` on PATH.** (→ K4.) Rejected: a wrapper script à la
   `formae/scripts/start-mcp.sh`. Formae ships its binary *inside* the plugin;
   `brag` is a released Homebrew binary already on PATH, so a direct entry is
   simpler and needs no bundled binary. The PATH prerequisite
   (`brew install jysf/bragfile/bragfile`) is documented in `plugin/README.md`
   (→ K9).

3. **Capture-nudge fires on `Stop`, at most once per session, ONLY after a
   commit lands mid-session; it emits agent-facing `additionalContext` and
   NEVER runs `brag add`; silenced by `BRAG_CAPTURE_NUDGE=off`.** (→ H-series;
   DEC-025.) This resolves STAGE-009 surfaced design question (d) —
   *"fire every session vs only when the session plausibly shipped."* **Answer:
   only when it plausibly shipped**, proxied by "HEAD advanced during the
   session" (recorded on the first Stop, compared on later Stops). Rationale
   and the *why not* alternatives (SessionEnd, per-turn nudging, TTY-gating)
   are in DEC-025; the load-bearing constraint is that **`Stop` fires every
   turn**, so an ungated nudge would nag — the once-per-session + commit gate
   is what makes it a nudge, not a nag.

4. **The nudge is delivered as agent-facing `hookSpecificOutput.additionalContext`
   (Claude then proposes a brag through BRAG.md's approval loop), NOT as a
   user-facing TTY line.** (→ H3; DEC-025.) This is a **§12(b) contract
   discovery**: the stage note framed the hook as "quiet, TTY-only, degrades on
   non-TTY" — but the real Stop-hook contract (confirmed against Claude Code
   2.1.201 docs) has **no TTY** (hooks run headless with JSON on stdin) and
   Stop output feeds *Claude*, not the user. So "quiet + degrades" is realized
   by the **fire-gating** (silent unless a commit landed, once per session,
   env-silenceable), not a TTY probe. This aligns better with BRAG.md's model
   anyway: the agent proposes, the user approves. The SPEC-039 milestone line
   remains the TTY-gated *CLI* surface; the plugin hook is a different surface
   with a different contract. Rejected: exit-2 blocking (would force Claude to
   keep working every session — the nag failure mode); a user-facing stderr
   line (Stop stderr on exit 0 is not reliably surfaced to the user).

5. **The slash-command is namespaced `/brag:brag`** (plugin `brag` +
   `commands/brag.md`). (→ K5.) Not a clean `/brag` — plugin commands are
   always `/<plugin>:<command>`. `plugin/README.md` documents the
   `/brag:brag` invocation; the manual-copy path (SPEC-022's
   `examples/brag-slash-command.md` → `~/.claude/commands/brag.md`) still gives
   a bare `/brag` and stays documented as the alternative.

6. **The SPEC-022 loose assets STAY.** `scripts/claude-code-post-session.sh`
   (a transport-agnostic "pipe a summary → candidate JSON" helper) and
   `examples/brag-slash-command.md` remain as the manual/standalone path and
   the convention proof; the plugin is the *recommended* packaged path. This
   keeps `test-docs.sh` groups I/J premises intact (no invalidated tests) and
   honours "evolving `claude-code-post-session.sh`" as *derivation* (the Stop
   hook is a new contract inspired by it), not deletion. Rejected: deleting the
   originals (would invalidate groups I/J and churn BRAG.md's stable references
   for no user benefit — both files still serve non-plugin users).

## L-watch outcome (PEEL — decided at design)

The stage sized SPEC-041 **S/M covering BOTH plugin packaging AND the v0.3.0
cut**, with an explicit instruction to peel the release cut into its own spec
if, after the §12(b) manifest pre-flight, the combined work reads **L**.

**Decision: PEEL. SPEC-041 = plugin packaging only; the v0.3.0 release cut
becomes SPEC-042.** Decided at design, recorded here (not discovered at build).

Rationale (the pre-flight did *not* argue against peeling — it retired the
manifest's L-risk, but the L-pressure now comes from breadth, not the manifest):

1. **The release tag is cut from `main` AFTER this plugin PR merges.** The
   v0.3.0 tag points at a merge commit that does not exist until SPEC-041
   lands. Bundling forces either tagging a pre-merge commit or one spec
   spanning two PRs — the latter violates `one-spec-per-pr`. This is a
   *structural* merge boundary, not a stylistic split.
2. **The coordinator folded in the MCP return-value regression test**, adding
   MCP-surface hardening scope to what was already an S/M packaging job.
3. **The release cut is a different kind of work** — release ops in the
   *separate* `homebrew-bragfile` tap repo, the dual-tag-on-same-commit rule,
   the `brew trust --cask` + Gatekeeper xattr pre-flight, and a clean-upgrade
   verification — a runbook, exactly like **SPEC-037** got its own spec, and
   mirroring the STAGE-007 (SPEC-029→033) and STAGE-008 doc/release-split
   discipline the stage's Complexity check names.

Consequences: STAGE-009 backlog updated to **5 specs** (SPEC-038..042);
SPEC-042 scaffolded as a `proposed` stub (design in a later session). SPEC-041
does **not** touch CHANGELOG, tags, the tap, or release config — those are
SPEC-042's. The `docs/tutorial.md`/`architecture.md` plugin walkthroughs ride
SPEC-042's release doc sweep (enumerated under Outputs).

## Implementation Context

*Read this section (and the files it points to) before starting the build
cycle. It is the handoff, folded into the spec.*

### Decisions that apply

- `DEC-025` (emitted by this spec) — plugin packaging layout (`plugin/` +
  root marketplace), the MCP-server-on-PATH entry, and the capture-nudge
  delivery model (Stop, once-per-session, commit-gated, agent-facing
  `additionalContext`, never-posts, env-silenced). Records the §12(b)
  contract discovery that the Stop-hook surface has no TTY.
- `DEC-024` — the reserved `agent:<name>`/`model:<id>` provenance convention
  the shipped assets must document (the exact literal: lowercase, no spaces,
  `model:` uses the canonical AGENTS.md model id).
- `DEC-011` — the 9-key entry JSON shape the regression-guard asserts parity
  against (`export.ToJSON`).
- `DEC-015` — provenance rides the polymorphic tags path (why the shipped docs
  say "reserved *tags*", zero schema change).

### Constraints that apply

- `stdout-is-for-data-stderr-is-for-humans` (blocking) — the capture-nudge
  hook's stdout is the hook protocol (JSON); it must emit *only* the
  `hookSpecificOutput` object or nothing, never stray human chatter. The
  regression-guard reinforces the MCP transport's stdout-is-protocol spine.
- `one-spec-per-pr` (blocking) — the peel (SPEC-042) keeps this PR
  single-concern; do not fold release work in.
- `no-sql-in-cli-layer` (blocking) — untouched (no CLI changes); the
  regression test lives in `internal/mcpserver`, which wraps `Store`.

### Prior related work

- `SPEC-040` (shipped) — the `brag mcp serve` MCP server + provenance the
  plugin points at; its Ship reflection Q1 named this spec's regression-guard.
- `SPEC-039` (shipped) — the CLI milestone line; the mirror stdout/stderr
  spine (TTY-gated CLI surface vs this spec's headless hook surface).
- `SPEC-022` (shipped) — `BRAG.md`, `examples/brag-slash-command.md`,
  `scripts/claude-code-post-session.sh`, `docs/brag-entry.schema.json`: the
  convention assets this spec packages.
- `SPEC-037` — the release-runbook precedent the peel mirrors.

### Out of scope (for this spec specifically)

- **The v0.3.0 release cut** — CHANGELOG `[0.3.0]`, the RC/final dual-tag,
  the tap bump, `brew trust --cask`/Gatekeeper pre-flight, clean-upgrade
  verification. **→ SPEC-042.**
- **`docs/tutorial.md` / `docs/architecture.md` plugin walkthroughs** — deferred
  to SPEC-042's release doc sweep (enumerated under Outputs).
- **Deleting or rewriting the SPEC-022 loose assets** (Locked decision 6).
- **Promoting provenance to first-class columns** (DEC-024 "later, if earned").
- **A live end-to-end plugin install/session run.** Design validated the
  *manifests* against the real loader and the *hook behavior* against crafted
  payloads; the one thing design cannot exercise is a live Claude Code session
  honouring the Stop-hook `additionalContext`. **Build should do a §12(b)-style
  live check:** `claude plugin marketplace add .` + `claude plugin install
  brag@bragfile` in a scratch dir, confirm the three components load
  (`claude plugin details brag`), and — if feasible — confirm the Stop hook
  fires. If the live `additionalContext` delivery differs from the documented
  contract, that is a build-time finding to raise, not a silent fix.

## Notes for the Implementer

Build transcribes the literals **verbatim** (they were pre-flighted at
design); verify diffs against them. Do not re-derive shapes.

### Literal — `plugin/.claude-plugin/plugin.json`

```json
{
  "name": "brag",
  "description": "Capture and recall career accomplishments — bundles the brag MCP server, the /brag slash-command, and a quiet session-end capture nudge.",
  "version": "0.3.0",
  "author": { "name": "jysf" },
  "homepage": "https://github.com/jysf/bragfile000",
  "repository": "https://github.com/jysf/bragfile000",
  "license": "MIT",
  "keywords": ["brag", "accomplishments", "career", "mcp", "productivity"],
  "mcpServers": {
    "brag": {
      "command": "brag",
      "args": ["mcp", "serve"]
    }
  }
}
```

> `version: "0.3.0"` tracks the release SPEC-042 cuts. If SPEC-042 slips the
> number, bump here to match (the `claude plugin tag` flow validates
> plugin.json ↔ marketplace agreement).

### Literal — `.claude-plugin/marketplace.json`

```json
{
  "name": "bragfile",
  "description": "The bragfile marketplace — install the brag Claude Code plugin (MCP server + /brag slash-command + capture nudge).",
  "owner": { "name": "jysf" },
  "plugins": [
    {
      "name": "brag",
      "source": "./plugin",
      "description": "Capture and recall career accomplishments — MCP server + /brag slash-command + session-end capture nudge.",
      "category": "productivity"
    }
  ]
}
```

> The top-level `description` is load-bearing: without it, `validate --strict`
> FAILS (the §12(b) catch).

### Literal — `plugin/hooks/hooks.json`

```json
{
  "description": "brag capture-nudge: once per session, after a commit lands, nudge Claude to propose a brag (never auto-posts).",
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "\"${CLAUDE_PLUGIN_ROOT}/hooks/capture-nudge.sh\"",
            "timeout": 5
          }
        ]
      }
    ]
  }
}
```

### Literal — `plugin/commands/brag.md`

```markdown
---
description: Draft a brag entry from this session
---

Review what shipped in this session. If a moment is brag-worthy per BRAG.md
(shipped feature, fixed significant bug, architectural decision, delivered
artifact), draft a single JSON object validating against
docs/brag-entry.schema.json: required `title` (action-verb, <=100 chars),
plus optional `description`, `project`, `type`, `tags` (comma-joined string
per DEC-004), and `impact` (concrete metric or named outcome). When the work
was agent-driven, include provenance as reserved tags `agent:<name>` and
`model:<id>` (lowercase, no spaces; e.g. `agent:claude-code`,
`model:claude-opus-4-8`). Present the JSON for my approval. Do not execute
`brag add --json` (or the brag_add MCP tool) until I confirm.
```

### Literal — `plugin/hooks/capture-nudge.sh` (`chmod +x`)

```bash
#!/usr/bin/env bash
# capture-nudge.sh — brag plugin Stop hook.
#
# Fires on every Stop (Claude finishing a turn), but nudges AT MOST ONCE per
# session, and ONLY after a git commit lands in the session's cwd — a
# lightweight "you plausibly shipped something" signal. The nudge is
# AGENT-FACING context (Claude then proposes a brag for the user's approval
# per BRAG.md); this hook NEVER runs `brag add`. Silence it with
# BRAG_CAPTURE_NUDGE=off.
#
# Contract (Claude Code Stop hook): reads a JSON payload on stdin
# ({session_id, cwd, hook_event_name, stop_hook_active}); on a fire it emits
# a hookSpecificOutput.additionalContext JSON object on stdout and exits 0.
# Every non-fire path exits 0 silently so the hook is quiet and never blocks.
#
# Degradation: missing jq, non-git cwd, or no new commit -> silent exit 0.
set -eu

# 1. Silence switch.
case "${BRAG_CAPTURE_NUDGE:-}" in
    off|0|false|no) exit 0 ;;
esac

# 2. jq is required to parse the payload; degrade quietly if absent.
command -v jq >/dev/null 2>&1 || exit 0

PAYLOAD=$(cat)
SESSION_ID=$(printf '%s' "$PAYLOAD" | jq -r '.session_id // empty')
CWD=$(printf '%s' "$PAYLOAD" | jq -r '.cwd // empty')
[ -n "$SESSION_ID" ] || exit 0
[ -n "$CWD" ] || exit 0

# 3. Only meaningful inside a git repo; HEAD is the "shipped" signal.
HEAD=$(git -C "$CWD" rev-parse HEAD 2>/dev/null || true)
[ -n "$HEAD" ] || exit 0

STATE_DIR="${BRAG_STATE_DIR:-$HOME/.bragfile}/capture-nudge"
mkdir -p "$STATE_DIR" 2>/dev/null || exit 0
MARKER="$STATE_DIR/$SESSION_ID"

# 4. First Stop of the session: record the baseline HEAD, never nudge yet.
if [ ! -f "$MARKER" ]; then
    printf 'baseline=%s\n' "$HEAD" > "$MARKER"
    exit 0
fi

# 5. Already nudged this session -> stay silent.
grep -q '^nudged$' "$MARKER" 2>/dev/null && exit 0

# 6. Nudge once, only if a commit landed since the baseline.
BASELINE=$(sed -n 's/^baseline=//p' "$MARKER" | head -1)
if [ -n "$BASELINE" ] && [ "$HEAD" != "$BASELINE" ]; then
    printf 'nudged\n' >> "$MARKER"
    jq -cn '{
        hookSpecificOutput: {
            hookEventName: "Stop",
            additionalContext: "A commit landed during this session. If something brag-worthy shipped, draft a brag entry for the user'"'"'s approval per BRAG.md (you can use the /brag:brag command): a required action-verb title plus optional project, type, tags, and a concrete impact. Stamp provenance as reserved tags agent:<name> and model:<id>. Do NOT run `brag add` until the user explicitly approves."
        }
    }'
fi
exit 0
```

> The embedded single-quote escaping around `user's` is the awkward-but-correct
> `'"'"'` idiom inside the outer single-quoted `jq` program. It was
> pre-flighted; transcribe verbatim.

### Literal — `plugin/README.md`

```markdown
# brag — Claude Code plugin

Bundles three ways to capture accomplishments without leaving a Claude Code
session:

- **MCP server** (`brag mcp serve`) — `brag_add` / `brag_list` /
  `brag_search` / `brag_stats` as typed tools over your `~/.bragfile/db.sqlite`.
- **`/brag:brag` slash-command** — draft a brag from the current session for
  your approval.
- **Capture-nudge hook** — after a commit lands in a session, quietly nudges
  the agent to propose a brag (it never posts one for you).

## Prerequisite

The plugin's MCP server runs the `brag` binary from your `PATH`. Install it
first:

    brew trust --cask jysf/bragfile/bragfile   # one-time (Homebrew 6.0+)
    brew install jysf/bragfile/bragfile

Verify: `brag --version`.

## Install

    claude plugin marketplace add jysf/bragfile000
    claude plugin install brag@bragfile

Then restart Claude Code. `claude plugin details brag` shows the loaded
components.

## Provenance convention

Agent-driven brags carry reserved-namespace tags so multi-agent work is
attributable later: `agent:<name>` (e.g. `agent:claude-code`) and
`model:<id>` (e.g. `model:claude-opus-4-8`) — lowercase, no spaces. The MCP
`brag_add` tool stamps these; query them with
`brag list --tag model:claude-opus-4-8` and `brag tags`.

## Silence the nudge

    export BRAG_CAPTURE_NUDGE=off

## Manual (non-plugin) path

Prefer to wire things by hand? Copy `examples/brag-slash-command.md` to
`~/.claude/commands/brag.md` for a bare `/brag`, and see BRAG.md for the
`scripts/claude-code-post-session.sh` pipe helper.
```

### BRAG.md additions

Add a subsection after the "JSON contract for programmatic capture" block
(around line 205, before "Three good examples"). Keep the existing SPEC-022
reference-asset lines (190, 196-199) and add:

```markdown
## Install as a Claude Code plugin (recommended)

The three integration pieces above — the slash-command, a session-capture
nudge, and a typed MCP write surface — ship as one Claude Code plugin:

    brew install jysf/bragfile/bragfile      # the binary the MCP server runs
    claude plugin marketplace add jysf/bragfile000
    claude plugin install brag@bragfile

This wires the `brag_add` / `brag_list` / `brag_search` / `brag_stats` MCP
tools, the `/brag:brag` slash-command, and a quiet capture-nudge Stop hook.
The loose files above remain for manual/standalone setup.

## Provenance: reserved `agent:` / `model:` tags

Agent-driven brags carry two reserved-namespace tags so multi-agent work is
attributable in hindsight: `agent:<name>` and `model:<id>` — lowercase, no
spaces, e.g. `agent:claude-code`, `model:claude-opus-4-8`. `agent` and
`model` are RESERVED prefixes (never topic tags), normally one of each per
entry, auto-populated by the MCP `brag_add` tool (rarely hand-typed). They
ride the normalized tags model (DEC-015) with zero schema change, so
`brag list --tag model:claude-opus-4-8` filters and `brag tags` counts them.
```

### README.md addition

Under the install section, add one line pointing at the plugin (keep the
`just install` / `brew` lines):

```markdown
- **Claude Code plugin:** `claude plugin marketplace add jysf/bragfile000`
  then `claude plugin install brag@bragfile` — see `plugin/README.md`.
```

### Build order

1. Create `plugin/` + `.claude-plugin/marketplace.json` from the literals.
   `chmod +x plugin/hooks/capture-nudge.sh`.
2. Run the pre-flight yourself: `claude plugin validate --strict plugin` and
   `claude plugin validate --strict .` — both must pass (§12(b) re-verify).
3. Add group K to `scripts/test-docs.sh` + the `scripts/test-capture-nudge.sh`
   harness; run `just test-docs` (and the hook harness).
4. Add `TestServer_AddReturnValueParity` + the `encoding/json` import to
   `server_test.go`; `go test ./internal/mcpserver/...`.
5. Update BRAG.md + README.md + write `plugin/README.md`.
6. Full gates: `go test ./...`, `gofmt -l .`, `go vet ./...`, `just test-docs`.
7. Live check (Out of scope note): `claude plugin marketplace add .` +
   `claude plugin install brag@bragfile` in a scratch checkout; confirm
   `claude plugin details brag` lists the three components.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** feat/spec-041-plugin-and-release
- **PR (if applicable):** https://github.com/jysf/bragfile000/pull/62
- **All acceptance criteria met?** yes — AC1-AC9 all pass. `go test ./...`
  (569 tests, incl. `TestServer_AddReturnValueParity`), `gofmt -l .` (empty),
  `go vet ./...` (clean), `CGO_ENABLED=0 go build ./...` (success),
  `just test-docs` (all OK incl. renamed group S), `just test-hook` (H1-H7
  all OK), `claude plugin validate --strict plugin` and
  `claude plugin validate --strict .` (both exit 0, re-verified at build per
  §12(b)).
- **New decisions emitted:**
  - `DEC-025` — Plugin packaging layout + capture-nudge delivery model
    (already emitted at design; no changes made at build)
- **Deviations from spec:**
  1. **Test-group letter collision: spec's "group K" renamed to "group S"
     in `scripts/test-docs.sh`.** The spec's Failing Tests section names
     the new plugin-assertion block "group K" with IDs K1-K11, but the
     letter `K` was already in use by a pre-existing group ("BRAG.md
     cross-reference", K1-K4, added by a prior spec) that the design
     session's own file did not check against before locking the IDs. Reusing
     `K` would have silently produced two same-named-but-different assertion
     blocks in one script. Resolved by lettering the new block `S` (the next
     unused letter after the existing `R`) and mapping the spec's K1-K11
     1:1 onto S1-S11 (S2/S3 additionally split into a CLI-validator check
     and a `-jq` structural-fallback check, both still under one base ID per
     the spec's own K2/K3 shape). No test *content* changed — only the
     group letter. This is exactly the class of premise-audit miss AGENTS.md
     §9 calls out (design should grep the target file before assigning new
     IDs); worth folding into the design-time pre-flight checklist.
  2. **Live-check finding (not a deviation from *this* build's own scope, but
     a delta from DEC-025's stated Validation bar) — raised per the Out of
     scope note's instruction, not silently fixed:** installed the plugin
     end-to-end in a scratch marketplace (`claude plugin marketplace add
     <scratch-copy-of-.claude-plugin+plugin>` → `claude plugin install
     brag@bragfile`). The install succeeds and `claude plugin details brag`
     correctly shows version 0.3.0, the description, 1 Hook (Stop), and 1
     Skill (the `/brag:brag` command) — but reports **"MCP servers (0)"**
     instead of listing `brag`. Ruled out a stale-binary explanation:
     re-tested with a freshly-built dev binary from this branch (which has
     `mcp serve`, unlike the currently-released Homebrew v0.2.0) shadowed
     onto `PATH`; the count stayed 0. The manifest shape
     (`mcpServers.brag.{command,args}`) is structurally the same pattern
     used by the already-installed `formae-mcp` plugin, which correctly
     shows "MCP servers (1)" for an equivalent `mcpServers.<name>.command`
     entry (formae wraps its args into one script path instead of a
     separate `args` array — not yet isolated as the cause, since further
     root-causing would have meant editing global `~/.claude` plugin-cache
     state, which the user did not want touched mid-investigation).
     `claude plugin validate --strict` (the actual loader gate this spec's
     §12(b) pre-flight targeted) passes clean for both manifests, and this
     spec's own AC3 test (S4) asserts the `mcpServers.brag.args` shape
     structurally via `jq` independent of the `details` display — so no
     AC-mapped test fails. Per user decision (asked directly): **record as
     a finding and proceed**, rather than block the PR or rework the
     manifest shape (DEC-025's locked decision #2 explicitly rejected a
     wrapper-script shape; changing it now on an unconfirmed-cause display
     quirk would be premature). DEC-025's "Revisit if" clause (b) already
     anticipates a live-delivery delta; this finding should be folded in
     there (or as a new revisit clause) at verify or ship.
  3. **Hook harness wired as a sibling `just test-hook` recipe**, not
     appended inside `test-docs.sh`'s group-S block — the spec explicitly
     offered both as acceptable ("wired into `just test-docs` after the
     group-K block, or a sibling recipe `just test-hook`"); the sibling form
     keeps the shell-assertion harness (`test-docs.sh`, `ok`/`fail`/`skip`
     line-per-assertion style) separate from the hook's own
     fixture-heavy scenario harness (temp git repo, PATH-shadowed `brag`
     stub, per-scenario setup), which felt like a different shape of test
     than the doc-content assertions.
- **Follow-up work identified:**
  - SPEC-042 (v0.3.0 release cut) — already scaffolded
  - ~~Root-cause the `claude plugin details` MCP-server-count-0 finding
    above~~ — **resolved below (punch-list fix).**

### Punch-list fix (post-build, pre-verify): MCP-server-count-0 root-caused and fixed

Coordinator-confirmed real bug blocking verify: an installed `brag` plugin
registered **0 MCP servers** (`claude plugin details brag` → `MCP servers
(0)`), so the marquee agent-native surface (`brag_add`/`brag_list`/etc. via
the plugin) was inert, even though `claude plugin validate --strict`, all
AC-mapped tests, and the hook harness were green.

- **Root cause, confirmed (not assumed):** Claude Code's plugin loader
  registers MCP servers from a separate **`plugin/.mcp.json`** file at the
  plugin root (a bare `{"<name>": {command, args?}}` map) — not from the
  inline `mcpServers` key inside `plugin/.claude-plugin/plugin.json`, which
  the loader does not read for registration. Confirmed against the cached
  `formae-mcp` reference plugin, which ships both `.mcp.json` (authoritative)
  and an equivalent inline `plugin.json` key (parity only).
- **Fix:** added `plugin/.mcp.json` — `{"brag": {"command": "brag", "args":
  ["mcp", "serve"]}}`. Kept the inline `mcpServers` key in `plugin.json`
  unchanged, matching the `formae-mcp` reference (ships both); documented as
  non-authoritative for registration in DEC-025's amendment.
- **PATH-vs-launcher question, decided against observed behavior:** tested
  whether the loader requires a `${CLAUDE_PLUGIN_ROOT}`-qualified launcher
  (the `formae` shape) or tolerates a bare command. Confirmed via a clean
  before/after scratch-marketplace install (see behavioral gate below) that
  a **bare `command:"brag"` works** — no `plugin/scripts/` shim needed. The
  PATH dependency on an installed `brag` binary (already documented in
  `plugin/README.md`'s Prerequisite section) is unchanged and is now also
  stated explicitly in DEC-025.
- **Behavioral acceptance gate (the actual pass condition, not
  `validate --strict`):** installed straight from this branch's working
  tree (`claude plugin marketplace add <repo-path>` →
  `claude plugin install brag@bragfile`) both **before** the fix (reproduced
  `MCP servers (0)`) and **after** (confirmed `MCP servers (1) brag`), then
  uninstalled/removed the scratch marketplace registration to leave no
  local `~/.claude` state behind. `claude plugin validate --strict` (both
  manifests) still passes unchanged — it was never the discriminating gate.
- **Regression guard added:** `scripts/test-docs.sh` group S gained
  `S12`/`S12-jq` — fails if `plugin/.mcp.json` goes missing or its
  `brag.command`/`brag.args` drift. This is cheap and structural; it cannot
  by itself re-prove runtime registration (only the behavioral
  `claude plugin details` check above does that), so it's a regression
  guard, not a substitute for the behavioral gate.
- **DEC-025 amended** (see its new "Amendment (2026-07-04, SPEC-041 build
  punch-list)" section): corrects sub-decision 2's manifest shape to
  "declared via `plugin/.mcp.json` (bare map); the inline `plugin.json` key
  is insufficient for registration," states the PATH dependency explicitly,
  and updates Validation to record the confirmed `MCP servers (1)` result.
- **Doc sweep:** checked `plugin/README.md`, `BRAG.md`, `README.md` for
  MCP-wiring descriptions that would now be wrong — none describe the
  manifest's internal registration mechanism (they describe user-facing
  install steps and the PATH prerequisite only, both still accurate), so no
  changes were needed there. Only `DEC-025` and this Build Completion
  described the manifest shape at the level of detail that needed
  correcting.
- **All gates re-run green post-fix:** `go test ./...` (569 tests),
  `gofmt -l .` (empty), `go vet ./...` (clean),
  `CGO_ENABLED=0 go build ./...` (success), `just test-docs` (all OK, incl.
  new S12/S12-jq), `just test-hook` (H1-H7 all OK),
  `claude plugin validate --strict plugin` and
  `claude plugin validate --strict .` (both exit 0).

**§12(b) refinement (flagged WATCH, not codified — this is one instance,
AGENTS.md's codification meta-rule wants N=2 same-outcome or a paired
opposing-outcome N=2):** design's §12(b) pre-flight ran the manifest through
`claude plugin validate --strict` (the loader's manifest-*validator*) but
not `claude plugin details` (the loader's component-*registration* surface)
against a real install — so a literal that validated strict still failed to
register at runtime, and no AC-mapped test caught it because every AC-mapped
test asserted JSON shape, not the loader's runtime behavior. The refinement
to "run the literal through its target tool": for a plugin-manifest MCP
claim specifically, the target tool is `claude plugin details` (does the
component register?), not `validate --strict` (is the JSON shape well
formed?) — they check different things and neither substitutes for the
other. Recorded here as a live instance; do not fold into AGENTS.md's §12(b)
text yet.

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing about the artifacts themselves — every literal transcribed
   clean and validated on the first try (manifests, hooks.json, the
   capture-nudge.sh quote-escaping, the slash-command). The one real
   friction was the test-group letter collision (Deviation 1): the spec
   was written without re-grepping `scripts/test-docs.sh` for the letter
   `K`, which was already spoken for. A two-second `grep -E '^# ===== Group'
   scripts/test-docs.sh` at design time would have caught it, the same
   audit-grep discipline AGENTS.md §9 already codifies for other artifact
   classes (just not yet extended to "next free letter in an
   already-lettered harness").

2. **Was there a constraint or decision that should have been listed but
   wasn't?**
   — The spec's Out of scope note framed the live check as "if feasible,"
   which was the right hedge — a *nested* live Claude Code session
   confirming a Stop hook's `additionalContext` delivery in real time isn't
   something this build session could exercise from inside itself. What
   would have helped: an explicit note that `claude plugin details` is
   itself part of DEC-025's Validation bar (it is — "lists the MCP server +
   `/brag:brag` command + Stop hook") and thus in-scope for the live check,
   not just "does it install." That's what actually surfaced the MCP-count
   finding.

3. **If you did this task again, what would you do differently?**
   — Grep the test harness for existing group letters before the spec locks
   IDs (design-time), and treat "does `claude plugin details` list all
   three components" as an explicit pass/fail live-check line item up front
   rather than discovering its scope mid-build.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>

---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-055
  type: story                      # epic | story | task | bug | chore
  cycle: verify
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-005
  stage: STAGE-015
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-10

references:
  decisions: [DEC-034, DEC-024]
  constraints:
    - stdout-is-for-data-stderr-is-for-humans
    - no-sql-in-cli-layer
    - errors-wrap-with-context
    - no-cgo
    - no-secrets-in-code
    - test-before-implementation
  related_specs: [SPEC-040, SPEC-041, SPEC-053]
---

# SPEC-055: `brag mcp install` — idempotent MCP client-config registration

## Context

The MCP server already exists — `brag mcp serve` (DEC-024/SPEC-040) exposes
`brag_add`/`brag_list`/`brag_search`/`brag_stats` over local stdio — but there
is no ergonomic way to *register* it with a client, so agents can't discover
the path and fall back to the CLI (observed live: a Claude Code agent logging
for the `standup` project used `brag add -p standup …` because nothing told it
how to connect). This spec marks the front door.

- Parent stage: `STAGE-015` (MCP first-class for agents), first backlog item.
  It "writes/merges the correct client config idempotently, never clobbering
  other MCP servers already present, with a `--dry-run` that prints the exact
  JSON + target path" (Success Criteria).
- Project: `PROJ-005` (agent-native depth).
- **The load-bearing property is idempotency + no-clobber** (Stage Design
  Notes): read the target file if present, merge only the `brag` key under
  `mcpServers`, preserve every other server and unrelated key, write back.
  This is a literal-artifact-shaped spec — the emitted JSON is fixed; the
  exact per-case bytes are embedded below (AGENTS.md §12 literal-artifact-as-
  spec) and were pre-flighted through the real merge at design (§12(b)).
- **New decision this spec emits: `DEC-034`** — the config-merge scheme
  (idempotent, no-clobber, semantic-preserving), the per-client target-path
  resolution, and the dry-run/output contract. Confidence 0.80.
- **Reuses `DEC-024`** (the `brag mcp serve` server this registers; the
  `stdout-is-for-data` spine faced at a new surface) and the exact
  `{"command":"brag","args":["mcp","serve"]}` block already shipped in
  `plugin/.mcp.json` and `plugin/.claude-plugin/plugin.json`.

## Goal

Add `brag mcp install [--client claude-code|claude-desktop|cursor] [--scope
user|project] [--dir PATH] [--dry-run]` — a storage-free command that writes or
merges the `brag` MCP server block into the target client's config file
**idempotently** and **without clobbering** any other server or unrelated key.
`--dry-run` prints the exact JSON that would be written (stdout) plus the
resolved target path (stderr) and writes nothing.

## Inputs

- **Files to read:**
  - `internal/cli/mcp.go` — `NewMCPCmd` (the `mcp` cobra parent) and
    `newMCPServeCmd`; `install` is added as a sibling of `serve` via
    `cmd.AddCommand(newMCPInstallCmd())`.
  - `plugin/.mcp.json` and `plugin/.claude-plugin/plugin.json` — the canonical
    `{"command":"brag","args":["mcp","serve"]}` block to reuse verbatim.
  - `internal/cli/errors.go` — `UserErrorf` / `ErrUser` for the incoherent-
    combo errors.
  - `internal/cli/project.go` — the existing `var getCwd = os.Getwd` injectable
    seam (reused for the `--dir` default); the general one-command-per-file
    cobra shape.
  - `internal/config/config.go` — the DEC-003 flag→env→default + tilde/`Abs`
    discipline mirrored (loosely) for target-path resolution.
  - `DEC-034` (this spec's decision), `DEC-024` (the server being registered).
- **External (verified at design, 2026-07-10):** the three clients' config
  file locations — see the DEC-034 path table and Implementation Context.
- **Related code paths:** `internal/cli/` (new `mcp_install.go` +
  `mcp_install_test.go`).

## Outputs

- **Files created:**
  - `internal/cli/mcp_install.go` — the `install` subcommand, the pure
    `mergeMCPConfig` helper, the `resolveInstallTarget` path resolver, the
    `mcpServerBlock` type, and the injectable `var userHomeDir = os.UserHomeDir`
    seam.
  - `internal/cli/mcp_install_test.go` — the Failing Tests below (created at
    **design**; made to pass at build).
  - `decisions/DEC-034-mcp-install-config-merge-scheme.md` — created at design.
- **Files modified:**
  - `internal/cli/mcp.go` — `NewMCPCmd` gains
    `cmd.AddCommand(newMCPInstallCmd())`; update its `Long` to mention
    `install` alongside `serve`.
  - `docs/api-contract.md`, `docs/tutorial.md`, `README.md` — a `brag mcp
    install` entry (flags, per-client snippet, dry-run). Build runs the §12
    audit-grep and enumerates the exact lines under Build Completion; the
    files to touch are listed in Implementation Context.
  - `projects/PROJ-005-agent-native-depth/stages/STAGE-015-mcp-first-class-for-agents.md`
    — the SPEC-055 backlog line flips to a build state at build.
- **New exports:** none exported outside the package. New unexported symbols:
  `mcpServerBlock`, `mergeMCPConfig`, `resolveInstallTarget`,
  `newMCPInstallCmd`, `runMCPInstall`, `var userHomeDir`.
- **Database changes:** none. The command imports neither `internal/storage`
  nor `database/sql` (`no-sql-in-cli-layer` satisfied structurally).

## Acceptance Criteria

- [ ] `brag mcp install` (no flags) registers `brag` for `claude-code` in
      `project` scope — i.e. writes/merges `<cwd>/.mcp.json` — because
      `--client` defaults to `claude-code` and `--scope` defaults to `project`.
- [ ] Merging into an **absent/empty** target file produces a file whose bytes
      are exactly the "Case A" literal below (2-space indent, trailing newline,
      valid JSON).
- [ ] Merging into a file that already contains **another** MCP server and an
      **unrelated top-level key** leaves both intact (values unchanged) and adds
      only the `brag` block — the "Case B" literal below.
- [ ] Running `install` **twice** is idempotent: the second run detects the
      byte-identical result, makes **no change** to the file, reports a no-op,
      and exits 0. A pre-existing unrelated server is still present after both
      runs.
- [ ] A **present-but-different** `brag` block is **overwritten** with the
      canonical block (the stale `args`/`command` do not survive) — the "Case D"
      literal below.
- [ ] `--dry-run` writes **nothing** (an absent target stays absent; an existing
      target is byte-unchanged), prints the exact would-be JSON to **stdout**,
      and prints a `Would write to <path>:` line to **stderr**.
- [ ] A successful real write prints `Registered brag MCP server in <path>` to
      **stderr** with **stdout empty**; the no-op prints `... already registered
      ... (no changes)` to stderr, stdout empty. Both exit 0.
- [ ] Target path resolves per the DEC-034 table for each supported
      client × scope (`.mcp.json`, `~/.claude.json`, `.cursor/mcp.json`,
      `~/.cursor/mcp.json`, and the macOS/Windows Claude Desktop path).
- [ ] Every unsupported combination is a `UserError`
      (`errors.Is(err, ErrUser)`, exit 1, **stdout empty**): unknown
      `--client`, unknown `--scope`, `claude-desktop --scope project`, `--dir`
      with `--scope user`, and `claude-desktop` on a non-macOS/non-Windows OS.
- [ ] `brag mcp install --help` lists the flags and shows a per-client example;
      `brag mcp --help` lists both `serve` and `install`.
- [ ] The command imports no `database/sql`, no SQL driver, and no
      `internal/storage` (`no-sql-in-cli-layer`).

## Failing Tests

Written during **design**, BEFORE build, in `internal/cli/mcp_install_test.go`.
Every embedded expected literal below was produced at design by running the
**real** `mergeMCPConfig` algorithm through a scratch Go program
(`json.Valid` == true on each; byte-equality across two runs == true), so the
literals are faithful, not hand-typed (§12(b) / literal-artifact pre-flight).
Tests use `t.TempDir()` for all paths; CLI tests use separate `outBuf`/`errBuf`
and assert no cross-leakage (§9).

### `internal/cli/mcp_install_test.go` — pure `mergeMCPConfig` (hermetic, no files)

- **`TestMergeMCPConfig_AbsentFile`** (maps to LD3 / AC "absent file"). `mergeMCPConfig(nil, "brag", canonicalBlock)` returns exactly:

  ```json
  {
    "mcpServers": {
      "brag": {
        "command": "brag",
        "args": [
          "mcp",
          "serve"
        ]
      }
    }
  }
  ```

  (byte-exact, with a trailing `\n`). Assert `json.Valid(out)` and
  `string(out) == wantCaseA`.

- **`TestMergeMCPConfig_PreservesOtherServerAndKeys`** (maps to LD3 no-clobber —
  the load-bearing property). Given the existing file

  ```json
  {
    "mcpServers": {
      "other": {
        "command": "other-server",
        "args": ["run"]
      }
    },
    "someTopLevelKey": {"keep": true}
  }
  ```

  `mergeMCPConfig(existing, "brag", canonicalBlock)` returns exactly (byte-exact
  + trailing `\n`):

  ```json
  {
    "mcpServers": {
      "brag": {
        "command": "brag",
        "args": [
          "mcp",
          "serve"
        ]
      },
      "other": {
        "command": "other-server",
        "args": [
          "run"
        ]
      }
    },
    "someTopLevelKey": {
      "keep": true
    }
  }
  ```

  Assert the full byte-exact match. (Note the `other` server's `["run"]` is
  re-indented to multi-line — this is the documented *semantic-not-byte-for-
  byte* preservation: the value survives, the whitespace is canonicalized.)

- **`TestMergeMCPConfig_Idempotent`** (maps to LD4 / AC "idempotent"). Let
  `first := mergeMCPConfig(existingWithOther, "brag", block)`; then
  `second := mergeMCPConfig(first, "brag", block)`; assert
  `bytes.Equal(first, second)` and that `second` still contains `"other"`.

- **`TestMergeMCPConfig_OverwritesDifferentBragBlock`** (maps to LD5). Given
  `{"mcpServers":{"brag":{"command":"OLD","args":["x"]}}}`, the result equals
  the Case A literal (canonical block) and does NOT contain `"OLD"` or `"x"`.

- **`TestMergeMCPConfig_InvalidJSONErrors`** (maps to `errors-wrap-with-
  context`). `mergeMCPConfig([]byte("{not json"), "brag", block)` returns a
  non-nil error whose message contains `parse existing config` (wrapped
  context), and the error is NOT `ErrUser` (it is an internal/IO-class error,
  exit 2 — a corrupt target is not the user's malformed *input*).

### `internal/cli/mcp_install_test.go` — path resolution (injectable-stub table)

- **`TestResolveInstallTarget_Table`** (maps to LD2 / AC "target path
  resolves"). With `userHomeDir` stubbed to return `/home/test` (save/restore),
  table over `{client, scope, dir}` → `wantPath`:

  | client       | scope   | dir     | wantPath |
  |--------------|---------|---------|----------|
  | claude-code  | project | `/proj` | `/proj/.mcp.json` |
  | claude-code  | user    | (n/a)   | `/home/test/.claude.json` |
  | cursor       | project | `/proj` | `/proj/.cursor/mcp.json` |
  | cursor       | user    | (n/a)   | `/home/test/.cursor/mcp.json` |

  Assert each `resolveInstallTarget(client, scope, dir)` returns `wantPath`,
  nil error.

- **`TestResolveInstallTarget_ClaudeDesktopByOS`** (maps to LD2, OS branch).
  With `userHomeDir` stubbed to `/home/test`, `resolveInstallTarget("claude-
  desktop", "user", "")`: when `runtime.GOOS == "darwin"` →
  `/home/test/Library/Application Support/Claude/claude_desktop_config.json`;
  when `"windows"` → the `%APPDATA%\Claude\claude_desktop_config.json` shape;
  otherwise → a `UserError` (`errors.Is(err, ErrUser)`) naming the OS. The test
  switches on `runtime.GOOS` and asserts the branch matching the host.

- **`TestResolveInstallTarget_UnsupportedCombos`** (maps to LD2 UserErrors).
  Each of the following returns `errors.Is(err, ErrUser) == true`:
  `("bogus","project","/p")` (unknown client), `("claude-code","bogus","/p")`
  (unknown scope), `("claude-desktop","project","/p")` (Desktop has no project
  scope).

### `internal/cli/mcp_install_test.go` — CLI command (cobra, `t.TempDir`, split buffers)

Harness: `newMCPInstallTestRoot(t)` builds `NewRootCmd("test")` +
`NewMCPCmd()`, sets separate `outBuf`/`errBuf`; `runMCPInstallCmd(t, args...)`
sets `["mcp","install", …]` and returns `(stdout, stderr, err)`.

- **`TestMCPInstall_WritesAbsentFile`** (maps to AC default + write). In a
  `t.TempDir()` `dir`, run `--client claude-code --scope project --dir <dir>`.
  Assert: `<dir>/.mcp.json` now exists with bytes == the Case A literal;
  `stderr` contains `Registered brag MCP server` and the path; `stdout` is
  empty; `err == nil`.

- **`TestMCPInstall_DefaultsClaudeCodeProject`** (maps to AC default). With
  `getCwd` stubbed to a `t.TempDir()`, run `mcp install` with NO client/scope
  flags; assert `<tmp>/.mcp.json` is written (proves both defaults) and stderr
  names it.

- **`TestMCPInstall_IdempotentPreservesOtherServer`** (maps to AC idempotent +
  no-clobber — LOAD-BEARING). Pre-write `<dir>/.mcp.json` containing only an
  `other` server. Run install once → assert both `brag` and `other` present.
  Capture the file bytes. Run install again → assert the file bytes are
  **byte-identical** to the first run, `stderr` contains `already registered`
  and `(no changes)`, `stdout` empty, exit 0, and `other` still present.

- **`TestMCPInstall_OverwritesStaleBragBlock`** (maps to LD5). Pre-write a
  `.mcp.json` with a `brag` block using `"command":"OLD"`. Run install → the
  file no longer contains `OLD`; it equals the Case A literal.

- **`TestMCPInstall_DryRunWritesNothing`** (maps to AC dry-run — LOAD-BEARING
  for the output contract). In a `t.TempDir()` `dir` with NO existing file, run
  `--dir <dir> --dry-run`. Assert: `<dir>/.mcp.json` does NOT exist afterward;
  `stdout` == the Case A literal (the exact would-be JSON); `stderr` contains
  `Would write to` and the resolved path; `err == nil`.

- **`TestMCPInstall_DryRunPreservesExistingFile`** (maps to AC dry-run). Pre-
  write a `.mcp.json` with an `other` server; run `--dry-run`; assert the file
  on disk is byte-unchanged from the pre-write, and `stdout` shows the merged
  would-be JSON (containing both `brag` and `other`).

- **`TestMCPInstall_UnknownClientIsUserError`** (maps to AC UserError + §9 no
  cross-leakage). Run `--client bogus`; assert `errors.Is(err, ErrUser)`,
  `stdout` empty. (With `SilenceErrors` the message rides `err`, not `errBuf`;
  the invariant under test is stdout purity + the ErrUser class.)

- **`TestMCPInstall_DesktopProjectIsUserError`** (maps to AC UserError). Run
  `--client claude-desktop --scope project`; assert `errors.Is(err, ErrUser)`,
  `stdout` empty.

- **`TestMCPInstall_DirWithUserScopeIsUserError`** (maps to LD2). Run
  `--scope user --dir /somewhere`; assert `errors.Is(err, ErrUser)`, `stdout`
  empty.

- **`TestMCPInstall_HelpListsFlagsAndExample`** (maps to AC help; §9 unique-
  token). `mcp install --help` stdout contains `--client`, `--scope`,
  `--dry-run`, and the distinctive example line `brag mcp install --dry-run`.

- **`TestMCP_InstallRegistered`** (extends existing `mcp_test.go` intent).
  `brag mcp --help` stdout contains both `serve` and `install`.

## Implementation Context

*Read this section (and the files it points to) before starting the build
cycle.*

### Decisions that apply

- `DEC-034` (this spec) — locks the merge scheme (idempotent, no-clobber,
  semantic-preserving via `encoding/json` round-trip), the per-client
  target-path table, the default `--client claude-code` / `--scope project`,
  the `--dir` default (cwd), the unsupported-combo `UserError`s, and the
  stdout(JSON on dry-run)/stderr(human) output contract.
- `DEC-024` — the `brag mcp serve` server being registered; its stdio-only,
  no-network shape; the `stdout-is-for-data-stderr-is-for-humans` spine that
  `install` re-honors (dry-run JSON is the data on stdout).

### Constraints that apply

- `stdout-is-for-data-stderr-is-for-humans` (blocking) — dry-run JSON → stdout;
  `Would write to …` / `Registered …` / `already registered …` → stderr;
  successful real write leaves stdout empty. Pinned by the dry-run and
  stdout-purity assertions.
- `no-sql-in-cli-layer` (blocking) — `install` imports neither
  `internal/storage` nor `database/sql`. It needs no DB at all.
- `errors-wrap-with-context` (warning) — parse/read/write errors wrapped
  (`parse existing config: %w`, `write config: %w`); user-facing incoherent
  combos use `UserErrorf`.
- `no-cgo` (blocking) — stdlib only (`encoding/json`, `os`, `path/filepath`,
  `runtime`); no new dependency, so `no-new-top-level-deps-without-decision`
  stays clean too.
- `no-secrets-in-code` (blocking) — nothing secret; the written block is a
  fixed public command invocation.
- `test-before-implementation` (blocking) — the Failing Tests above are written
  at design.

### Prior related work

- `SPEC-040` (shipped) — `brag mcp serve`; DEC-024; `internal/cli/mcp.go`
  (`NewMCPCmd`/`newMCPServeCmd`) that `install` slots beside.
- `SPEC-041` (shipped) — the Claude Code plugin packaging; `plugin/.mcp.json`
  and `plugin/.claude-plugin/plugin.json` carry the exact `{command,args}`
  block reused here. (Its lesson — §12(b) must target the *behavioral*
  registration surface, not just shape validation — is why the config paths
  here were verified against each client's official docs, not assumed.)
- `SPEC-053` (shipped) — the recent literal-artifact CLI spec whose shape,
  split-buffer test harness, and `UserErrorf`/`errors.Is(err, ErrUser)` idiom
  this spec mirrors.

### Out of scope (for this spec specifically)

- **Closing the unregistered-project gap** (`brag project ensure`, auto-
  register on `brag_add`) — that is SPEC-056.
- **The "For AI agents" docs page + full tool schemas** — that is SPEC-057.
  This spec adds only the minimal `brag mcp install` reference lines.
- **Editing a *running* client's live session.** MCP servers connect at client
  startup; a session must reconnect after install. Docs note it; the command
  does not attempt live re-registration.
- **A format-preserving JSON edit** of the target file (byte-stable outside the
  brag block). Deferred (DEC-034 revisit trigger (a)); v1 preserves values
  semantically, not formatting.
- **Additional clients** (Windsurf, VS Code, Zed, Linux Claude Desktop) and a
  `--type stdio` field. Additive later (DEC-034 revisit triggers (c)/(d)).
- **Uninstall / list-installed.** Not asked for; a future spec if wanted.

## Notes for the Implementer

**This is a literal-artifact-as-spec deliverable (AGENTS.md §12).** The three
JSON literals below are byte-exact outputs of the real `mergeMCPConfig` at
design (`json.Valid` == true; run-twice byte-equality == true). Transcribe the
helper so its output matches these byte-for-byte; verify diffs against them.

### The canonical server block (reuse verbatim from `plugin/.mcp.json`)

Build it from a fixed Go struct so key order is deterministic (`command` then
`args`), matching the plugin:

```go
type mcpServerBlock struct {
    Command string   `json:"command"`
    Args    []string `json:"args"`
}

var canonicalBragBlock = mcpServerBlock{Command: "brag", Args: []string{"mcp", "serve"}}
```

### The pure merge helper (transcribe exactly — pre-flighted at design)

```go
// mergeMCPConfig returns the full bytes to write for the target config file
// after ensuring mcpServers.<serverName> == block, preserving every other
// top-level key and every other server. Absent/empty existing input yields a
// file containing just the mcpServers.<serverName> block. Output is 2-space
// indented with a trailing newline. Preservation is SEMANTIC (values survive);
// encoding/json canonicalizes key order + whitespace on rewrite (DEC-034 #3).
func mergeMCPConfig(existing []byte, serverName string, block mcpServerBlock) ([]byte, error) {
    top := map[string]json.RawMessage{}
    if len(bytes.TrimSpace(existing)) > 0 {
        if err := json.Unmarshal(existing, &top); err != nil {
            return nil, fmt.Errorf("parse existing config: %w", err)
        }
    }
    servers := map[string]json.RawMessage{}
    if raw, ok := top["mcpServers"]; ok {
        if err := json.Unmarshal(raw, &servers); err != nil {
            return nil, fmt.Errorf("parse mcpServers: %w", err)
        }
    }
    blockBytes, err := json.Marshal(block)
    if err != nil {
        return nil, fmt.Errorf("marshal server block: %w", err)
    }
    servers[serverName] = blockBytes
    serversBytes, err := json.Marshal(servers)
    if err != nil {
        return nil, fmt.Errorf("marshal mcpServers: %w", err)
    }
    top["mcpServers"] = serversBytes
    out, err := json.MarshalIndent(top, "", "  ")
    if err != nil {
        return nil, fmt.Errorf("marshal config: %w", err)
    }
    return append(out, '\n'), nil
}
```

Why this shape survives idempotency: the brag block is rebuilt from the fixed
struct (compact `{"command":"brag","args":["mcp","serve"]}` as a RawMessage),
the maps are re-marshaled deterministically (Go sorts map keys), and
`MarshalIndent` re-indents uniformly (it does not reorder keys inside a
RawMessage). So a second run over the first run's output is byte-identical —
that is how the no-op is detected (compare merged bytes to the on-disk bytes).

### Case A — absent/empty target file (byte-exact)

```json
{
  "mcpServers": {
    "brag": {
      "command": "brag",
      "args": [
        "mcp",
        "serve"
      ]
    }
  }
}
```

### Case B — merged into a file with another server + an unrelated key (byte-exact)

Given input `{"mcpServers":{"other":{"command":"other-server","args":["run"]}},"someTopLevelKey":{"keep":true}}`:

```json
{
  "mcpServers": {
    "brag": {
      "command": "brag",
      "args": [
        "mcp",
        "serve"
      ]
    },
    "other": {
      "command": "other-server",
      "args": [
        "run"
      ]
    }
  },
  "someTopLevelKey": {
    "keep": true
  }
}
```

### Case D — stale `brag` block overwritten (byte-exact == Case A)

Given input `{"mcpServers":{"brag":{"command":"OLD","args":["x"]}}}`, the output
is byte-identical to **Case A** (the stale `OLD`/`x` do not survive).

### Path resolution (`resolveInstallTarget`)

```go
var userHomeDir = os.UserHomeDir // injectable seam for tests (AGENTS.md §9)

// resolveInstallTarget maps (client, scope, dir) → the config file path.
// dir defaults (in the caller) to getCwd() when empty; it is only consulted
// for project scope. Unsupported combos return UserErrorf(...).
```

Resolution rules (DEC-034 table):
- `claude-code` + `project` → `filepath.Join(dir, ".mcp.json")`
- `claude-code` + `user`    → `filepath.Join(home, ".claude.json")`
- `cursor` + `project`      → `filepath.Join(dir, ".cursor", "mcp.json")`
- `cursor` + `user`         → `filepath.Join(home, ".cursor", "mcp.json")`
- `claude-desktop` + `user` → by `runtime.GOOS`:
  - `darwin`  → `filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")`
  - `windows` → `filepath.Join(os.Getenv("APPDATA"), "Claude", "claude_desktop_config.json")` (fall back to `filepath.Join(home, "AppData", "Roaming", "Claude", …)` if `APPDATA` is empty)
  - else → `UserErrorf("claude-desktop config path is unknown on %s (supported: macOS, Windows)", runtime.GOOS)`
- `claude-desktop` + `project` → `UserErrorf("claude-desktop has no project scope; use --scope user")`
- unknown client → `UserErrorf("unknown --client %q (accepted: claude-code, claude-desktop, cursor)", client)`
- unknown scope  → `UserErrorf("unknown --scope %q (accepted: user, project)", scope)`

Tilde/`Abs`: `home` comes from `userHomeDir()` (already absolute). For
`project` paths, `dir` is either the user's `--dir` or `getCwd()`; run the
final path through `filepath.Abs` for a clean absolute target (mirrors DEC-003
discipline). Do NOT expand `~` inside `--dir` beyond what `config`-style
handling would do — keep it simple: `filepath.Abs(filepath.Join(dir, …))`.

### The command (`newMCPInstallCmd` / `runMCPInstall`)

- Flags (defaults stated explicitly — flag-default-explicitness §12):
  - `cmd.Flags().String("client", "claude-code", "MCP client to configure (one of: claude-code, claude-desktop, cursor)")`
  - `cmd.Flags().String("scope", "project", "config scope (one of: user, project)")`
  - `cmd.Flags().String("dir", "", "project directory for project scope (default: current directory)")`
  - `cmd.Flags().Bool("dry-run", false, "print the exact JSON + target path without writing")`
- `runMCPInstall` flow:
  1. Read flags. If `scope == "user"` and `--dir` was set
     (`cmd.Flags().Changed("dir")`) → `UserErrorf("--dir applies only to project scope")`.
  2. `dir` default: if empty, `dir, err = getCwd()` (wrap err).
  3. `target, err := resolveInstallTarget(client, scope, dir)` (returns the
     UserErrors for unsupported combos).
  4. Read existing bytes: `existing, err := os.ReadFile(target)`; treat
     `os.IsNotExist` as empty (`existing = nil`), wrap any other error.
  5. `merged, err := mergeMCPConfig(existing, "brag", canonicalBragBlock)`.
  6. **dry-run:** `fmt.Fprintf(cmd.ErrOrStderr(), "Would write to %s:\n", target)`
     then `cmd.OutOrStdout().Write(merged)`; return nil (write nothing).
  7. **no-op:** if the file existed and `bytes.Equal(existing, merged)` →
     `fmt.Fprintf(cmd.ErrOrStderr(), "brag MCP server already registered in %s (no changes)\n", target)`; return nil.
  8. **write:** `os.MkdirAll(filepath.Dir(target), 0o755)` (wrap), then
     `os.WriteFile(target, merged, 0o644)` (wrap `write config: %w`), then
     `fmt.Fprintf(cmd.ErrOrStderr(), "Registered brag MCP server in %s\n", target)`; return nil.
- Wire it: in `NewMCPCmd`, add `cmd.AddCommand(newMCPInstallCmd())` and extend
  the parent `Long` to mention `install`.

### Output-stream discipline (constraint-critical)

- **stdout** receives bytes ONLY on `--dry-run` (the would-be JSON). On a real
  write or no-op, stdout stays empty. Never print human text to stdout.
- **stderr** receives all human lines (`Would write to …`, `Registered …`,
  `already registered …`). The `UserError`s propagate via the returned error
  (main.go maps `ErrUser` → exit 1 and prints to stderr); with
  `SilenceErrors`/`SilenceUsage` on root, cobra prints nothing itself.

### NOT-contains self-audit (§12)

The help-test asserts `--help` stdout contains `--client`, `--scope`,
`--dry-run`, and `brag mcp install --dry-run`. No Failing Test asserts a
token *absent* from the `Long`/help, so no NOT-contains grep is load-bearing
here — but keep the `Long` free of any client name you don't support.

### Docs sweep (build, §12 audit-grep)

At build, `grep -rn "mcp serve\|brag mcp\|mcpServers" docs/ README.md` and add
a `brag mcp install` entry where `brag mcp serve` is documented:
`docs/api-contract.md` (a flags subsection + the per-client target paths),
`docs/tutorial.md` (a short "register the MCP server" note), `README.md` (one
line in the MCP/agents section). Enumerate the exact lines touched under Build
Completion (design lists the files; build re-verifies the hits).

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** feat/spec-055-mcp-install
- **PR (if applicable):** (see PR opened from this branch, base `main`)
- **All acceptance criteria met?** yes — all 22 tests in `mcp_install_test.go`
  pass; Case A/B/D outputs are byte-exact; idempotent re-run is byte-identical;
  dry-run writes nothing (JSON→stdout, path→stderr); every unsupported combo is
  `errors.Is(err, ErrUser)` with empty stdout; `mcp install --help` and `mcp
  --help` list the expected flags/subcommands. Full gate set green (`go test
  ./...` 781 pass, `gofmt -l .` clean, `go vet` clean, `CGO_ENABLED=0 go build`
  OK, `just test-docs` ALL OK, `just test-hook` ALL OK). End-to-end binary run
  confirmed the stdout/stderr discipline live.
- **New decisions emitted:**
  - `DEC-034` — MCP install config-merge scheme (emitted at design)
- **Deviations from spec:**
  - None. `mergeMCPConfig` and `resolveInstallTarget` transcribed per the
    embedded artifacts; the flag set, path table, UserError combos, and output
    contract match DEC-034 exactly.
- **Follow-up work identified:**
  - None new. SPEC-056 (unregistered-project gap) and SPEC-057 (agent docs
    page) already carry the deferred scope noted in "Out of scope".

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing material. The literal-artifact-as-spec approach (byte-exact Case
   A/B/D + the verbatim `mergeMCPConfig`/`resolveInstallTarget` transcriptions
   + the flow numbered 1–8) made the build almost mechanical: transcribe, wire,
   run the pre-authored tests to green. The design-time pre-flight of the merge
   bytes meant zero guesswork about whitespace/key-order.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. The blocking constraints (stdout/stderr spine, no-sql-in-cli-layer,
   errors-wrap-with-context, no-cgo, no-new-deps) were all listed and were the
   right ones; `no-sql-in-cli-layer` is satisfied structurally (the file
   imports only stdlib + cobra).

3. **If you did this task again, what would you do differently?**
   — Nothing significant. One nicety: the spec's flow step 6 puts the dry-run
   `Fprintf` before the `Write`, which I followed verbatim — worth keeping since
   it means the stderr annotation precedes the stdout payload when both are
   viewed interleaved.

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

# Code review — bragfile app (Go CLI + MCP server + plugin)

- **Date:** 2026-10-03
- **Commit reviewed:** `95d7b39` (main)
- **Method:** read-only. Automated gates run first, then six parallel area reviews (storage, CLI, export/aggregate/story, MCP+plugin, security/supply-chain, tests/quality). Headline claims re-checked against the code by the orchestrating agent; severities below are the *post-verification* ratings.
- **Out of scope:** the spec-driven process (`projects/`, `decisions/`, `guidance/`, `scripts/`, templates).
- **Changes made:** none to the app. This file is the only artifact.

## TL;DR

The codebase is in good shape. No critical or high-severity defects were confirmed. The findings are hardening items and edge cases, plus a handful of test-coverage gaps. The top three worth doing:

1. **Tighten file permissions and ignore-errors in `storage.Open`** (existing DB not chmod'd; pre-create error swallowed; backup sidecar mode unspecified).
2. **`(no project)` sentinel collides with a real project of that name** (the only confirmed data-correctness bug).
3. **Add CI steps for `-race` and `govulncheck`, and pin the Go/GoReleaser toolchain** (currently floating).

## Automated baseline (all run on this commit)

| Gate | Result |
|---|---|
| `go build ./...` | clean |
| `go vet ./...` | clean |
| `gofmt -l .` | clean |
| `golangci-lint run` (9 linters) | 0 issues |
| `go test ./...` | all pass |
| `go test -race -count=1 ./...` | all pass |
| `govulncheck ./...` | no vulnerabilities |
| Coverage | 83.9% total. Lowest: `cmd/brag` 23%, `storage` 78.8%, `editor` 77.3%, `cli` 82.4% |

Size: ~11.5k source lines, ~29.7k test lines. No TODO/FIXME/HACK markers in `cmd/` or `internal/`.

## Findings

Severity: **M** medium, **L** low, **N** nit. Confidence is "code-read" unless noted.

### Medium

| # | Area | Finding | Location |
|---|---|---|---|
| M1 | Aggregate | **Sentinel collision.** `NoProjectKey = "(no project)"` is a legal project name. Entries in a project literally named that merge with unassigned entries in `ByProject`, `GroupEntriesByProject`, `GroupForHighlights`, spark rows and story threads. Fix: key by struct/flag and render the sentinel only at output, or reject the name at ingress. | `internal/aggregate/aggregate.go:16-19,73-80,122`; `internal/export/spark.go:53-59` |
| M2 | Storage/Security | **Existing DB file keeps its old mode.** `OpenFile(..., 0o600)` only applies on create, so a pre-existing 0644 DB stays readable, contradicting the "never world-readable" comment. Matters most for `--db` pointing into a shared dir. | `internal/storage/store.go:50-54` |
| M3 | Tests | **`brag spark` renderers have no direct tests** (`ToSparkMarkdown`, `ToSparkJSON`, `sparkRows` at 0% in `internal/export`). Only indirect CLI coverage. | `internal/export/spark.go:90-196` |
| M4 | Tests | **Sleep-based storage tests.** 9 `time.Sleep` calls incl. a 1.1 s sleep to force `UpdatedAt` to change, adds ~3 s and is a flake surface on slow CI. The repo already has a `Backdate` helper and a stated convention against this. | `internal/storage/store_test.go:343,575-594,942-1014` |
| M5 | MCP | **Provenance strings are unbounded.** `agent`/`model`/`session` are normalised but never length-capped before being stamped into tags, so a client can persist an arbitrarily large tag past the limits `capture.Validate` enforces on ordinary fields. | `internal/mcpserver/server.go:71-82,109-133`; `provenance.go:19-76` |
| M6 | MCP | **Request context is dropped.** Handlers take `ctx` as `_`; storage methods use `context.Background()`. A cancelled or disconnected call keeps running, and a cancelled `brag_add` can still commit, so a retrying client may create duplicates. | `internal/mcpserver/server.go:88-90,202-204,283-285,311-313` |

### Low

| # | Area | Finding | Location |
|---|---|---|---|
| L1 | Storage | Pre-create `OpenFile`/`Close` errors are silently ignored, so permission or path failures surface later as a vaguer SQLite error. | `store.go:52-54` |
| L2 | Storage | DSN is built as `path + "?_pragma=..."`. A DB path containing `?` or `#` would be parsed as part of the DSN. Code-read only, not reproduced; self-inflicted via `--db`, so low. | `store.go:81` |
| L3 | Storage/Security | Migration backup sidecar (`VACUUM INTO`) has no explicit 0600; mode depends on driver and umask. It is a full copy of the corpus. | `internal/storage/backup.go:68-76` |
| L4 | Storage | Two processes opening at once can both see pending migrations and collide on the timestamped backup name. Backup and migrate are not serialised across processes. Confidence medium. | `backup.go:58-76`; `migrate.go:62-75` |
| L5 | Storage | `EnsureProject` / `EnsureLocation` are read-then-insert. Under a race the loser gets a raw uniqueness error rather than the documented idempotent or `ErrLocationOtherProject` result. The UNIQUE constraint prevents corruption. | `internal/storage/project.go:83-140` |
| L6 | CLI/Editor | `$EDITOR`/`$VISUAL` is split with `strings.Fields`, so a quoted path with spaces (`"/Applications/My Editor.app/..." --wait`) breaks. Not injection: no shell is used. | `internal/editor/launch.go:73-77` |
| L7 | CLI | `brag export --out` does `defer f.Close()` and discards the close error, so a write-back failure can exit 0. It also truncates through a symlink by design, which is documented overwrite but worth a note. | `internal/cli/export.go:141-150` |
| L8 | CLI | `project edit` with scalar and location changes is not atomic: `UpdateProject` commits, then `EditLocations` may fail, leaving a partial edit. | `internal/cli/project.go:529-567` |
| L9 | CLI | `mcp install` read-merge-rename can lose a concurrent writer's update to the client config (atomic rename prevents corruption but not lost updates). Confidence medium. | `internal/cli/mcp_install.go:180-210` |
| L10 | Story | `LoadProfile(name)` joins `name + ".yaml"` under the override dir with no basename check, so `--audience ../x` reads a YAML outside `~/.bragfile/story-profiles`. Read-only, local, parsed as a profile only. | `internal/story/profile.go:77-80` |
| L11 | Plugin | `capture-nudge.sh` uses `session_id` from the hook payload as a filename. `../` lets it create a `baseline=...` file outside the state dir. The script only writes when the target is **not** an existing file, so it cannot truncate existing files (the subagent's claim was corrected on verification). The input comes from the Claude Code host, not an arbitrary party. Fix: allowlist `[A-Za-z0-9_-]`. | `plugin/hooks/capture-nudge.sh:27-43` |
| L12 | Export/Story | Markdown renderers emit titles, project names and impact text raw. A title starting with `#` or containing a newline alters structure. In LLM-facing output (`story`, `memory`) stored text is also an untrusted-content/prompt-injection surface with no delimiting. JSON is unaffected. | `internal/export/markdown.go:36-62`; `story/bundle.go:121-137`; `export/memory.go:52-80` |
| L13 | MCP | `brag_list`/`brag_search` return everything when `limit` is omitted. A large corpus can exceed client context or memory. | `internal/mcpserver/server.go:191-216,276-304` |
| L14 | Timewindow | `--since Nd` multiplies `time.Duration(n) * 24h` with no range check. A very large `n` overflows and can yield a recent cutoff. | `internal/timewindow/timewindow.go:54-65` |
| L15 | Memory | `memory.Slice` documents a positive budget but does not enforce it; CLI and MCP do, so only direct callers are affected. | `internal/memory/memory.go:129-149` |
| L16 | Spark | `spark.Line` subtracts ints before converting to float, so extreme values overflow. App data is small non-negative counts, so unreachable today. | `internal/spark/spark.go:38-40` |
| L17 | CI/Supply chain | `check-latest: true` on setup-go (CI and release) and `version: '~> v2'` for GoReleaser float the toolchain without a reviewed change. Actions themselves are SHA-pinned. | `.github/workflows/ci.yml:28-33`; `release.yml:31-43` |
| L18 | CI | `-race` and `govulncheck` are not CI steps (both pass locally today). | `.github/workflows/ci.yml` |
| L19 | Tests | `runMCPServe` (production startup) is 0% covered; `storagetest.Backdate` has no direct test; `ResolveDirective` 45%. | `internal/cli/mcp.go:61-76`; `storagetest/storagetest.go`; `story/profile.go:174-191` |
| L20 | Tests | No fuzz tests for the three boundary parsers (`ftsquery.Build`, `editor.Parse`, `timewindow`). | `internal/ftsquery`, `internal/editor`, `internal/timewindow` |
| L21 | Structure | `internal/cli/project.go` is 792 lines with a 1,794-line test file covering nine subcommands. | `internal/cli/project.go` |

### Nits

| # | Finding | Location |
|---|---|---|
| N1 | Editor subprocess uses `os.Stdin/Out/Err` rather than Cobra's streams, inconsistent with the repo's buffer-based test convention. | `internal/editor/launch.go:21-23` |
| N2 | `apply0001Only` test helper leaks the DB handle if setup fails before return. | `internal/storage/store_test.go:1035-1064` |
| N3 | Several `export`/`aggregate` helpers are exported but used only inside the module (`ToJSON`, `ToTagsJSON`, `ToTSVRow`, `ToSparkMarkdown`, `ToSparkJSON`, `Share`). | `internal/export/json.go`, `spark.go`; `aggregate.go:485` |

## Carried over from the 2026-04-26 security review

- **Fixed and still holding:** 0700 dir / 0600 new DB (M1 then, but see M2 now for existing files); export file mode 0600; SHA-pinned actions; release timeout and concurrency.
- **Still open, deferred then:** editor buffer has no size cap (`io.ReadAll`), and no signal-driven temp-file cleanup (`internal/editor/editor.go:112-116`, `launch.go:34-40`). Local trust boundary only.
- **Stale:** the old report's PAT recommendation names `homebrew-bragfile`. The tap is now the shared `jysf/homebrew-tap`. Token scope and branch protection are GitHub settings and could not be verified from the repo.
- **Observation, not a defect:** release artifacts have checksums but no signing or provenance attestation.

## Checked and found sound

- **Layering:** no SQL or driver imports under `internal/cli`; depguard enforces it. No import cycles.
- **SQL safety:** all values are bound parameters; dynamic clauses are fixed fragments; tag match is equality, not interpolated LIKE; `VACUUM INTO` destination is bound.
- **Resource handling:** `rows.Err()` and `defer rows.Close()` are present on all iterators.
- **Migrations and FTS:** each migration and its tracking row run in one transaction; the runner is idempotent. FTS triggers cover entry and tag insert/update/delete/rename, with tests.
- **Concurrency policy:** `busy_timeout`, `_txlock=immediate`, and one pooled connection per store.
- **Dates:** UTC RFC3339 storage, half-open windows, DST-safe day math via `AddDate`, year-rollover `--previous` tested.
- **Ranking and output:** `memory.Slice` ordering is deterministic (ID tie-break); the UTF-8 byte budget never truncates mid-rune; map iteration is always sorted before output.
- **CLI behaviour:** deletes default to abort and accept only `y`/`Y`; `brag add --json` rejects unknown fields and trailing data; stdout/stderr separation holds; the editor is launched with `exec.Command` and no shell, with 0600 temp files.
- **MCP:** exactly one write tool (`brag_add`); no edit/delete/file-write tools; `budget` is a `*int` so omitted and explicit `0` differ; resource names are `PathUnescape`d and validated; a transport test asserts nothing reaches `os.Stdout`.
- **Docs:** `--help` for 10 commands matches `docs/api-contract.md`.

## Suggested order of work

1. **Small and safe (one PR):** M2, L1, L3 (`storage.Open` and backup permissions and errors), L7 (close error), L11 (allowlist session id), L10 (profile name check).
2. **Correctness:** M1 (sentinel collision), L14 (duration overflow), L15 (`Slice` guard).
3. **Agent surface:** M5 (cap provenance), L13 (default/max limit), L12 (label stored text as untrusted data), M6 (thread `ctx` through storage).
4. **CI:** L17, L18.
5. **Tests:** M3, M4, L19, L20.
6. **Refactor when convenient:** L21, N3.

## Limits of this review

- Static reading plus the automated gates above. No fuzzing, no load testing, no live MCP client session, no `goreleaser` dry run.
- Area reviews ran as sub-agent passes; I re-verified M2, L2, L10, L11 and the sentinel mechanics directly, and corrected L11's severity. The remaining findings are code-read by a sub-agent and marked with its stated confidence.
- L2 (DSN with `?`) was not reproduced empirically.

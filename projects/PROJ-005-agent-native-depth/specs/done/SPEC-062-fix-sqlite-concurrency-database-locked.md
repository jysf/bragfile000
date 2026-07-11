---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-062
  type: bug                        # epic | story | task | bug | chore
  cycle: ship
  blocked: false
  priority: high
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-005
  stage: STAGE-016
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-10

references:
  decisions:
    - DEC-038
    - DEC-001
    - DEC-002
    - DEC-021
  constraints:
    - no-cgo
    - storage-tests-use-tempdir
    - errors-wrap-with-context
  related_specs: []
---

# SPEC-062: fix sqlite concurrency database locked

## Context

A pre-release storage audit (HIGH) found `internal/storage/store.go` `Open`
calling `sql.Open("sqlite", path)` with a bare path — no busy timeout, no
connection cap, default (deferred) transactions. So concurrent access fails
immediately with `database is locked (5) (SQLITE_BUSY)` instead of waiting.
Confirmed repro: two `Store` handles on one temp DB doing concurrent `Add`s →
~39/40 writes fail, a second handle's write returning `database is locked` in
~30µs with no wait.

This is reachable and **amplified by v0.5.0**: `brag mcp serve` (a headline
feature, STAGE-016) holds ONE long-lived `Store` for the whole process while
the user also runs `brag add` from a shell, and Claude Code hooks fire `brag
add` — any overlap fails. Concurrent MCP tool calls can also self-conflict
since `database/sql` may open multiple pooled connections.

Part of `PROJ-005` (agent-native depth) / `STAGE-016` (polish) — a reliability
fix hardening the storage layer under the concurrency the MCP server
introduces.

## Goal

Make `storage.Open` configure SQLite so concurrent access from multiple `Store`
handles/processes WAITS and succeeds instead of failing with `database is
locked`, without switching journal mode (keeping the single-file backup story
intact) and without adding a CGO dependency.

## Inputs

- **Files to read:** `internal/storage/store.go` — `Open`; `internal/storage/backup.go`
  (`VACUUM INTO`, must keep the clean path); `internal/config/config.go`
  (`ResolveDBPath`, must keep the clean path).
- **External APIs:** `modernc.org/sqlite` v1.53.0 DSN pragmas — `_pragma` and
  `_txlock`.
- **Related code paths:** `internal/storage/store_test.go`.

## Outputs

- **Files modified:** `internal/storage/store.go` — build a DSN with
  `busy_timeout(5000)` + `_txlock=immediate` in `Open`, call
  `db.SetMaxOpenConns(1)`; keep the on-disk path clean.
- **Files created:** `decisions/DEC-038-sqlite-concurrency-busy-timeout-single-conn.md`.
- **Tests added:** `internal/storage/store_test.go` —
  `TestOpen_ConcurrentWritesAcrossHandles`, `TestOpen_BusyTimeoutPragma`.
- **Database changes:** none (no schema change, no migration).

## Acceptance Criteria

- [x] Two separate `*Store` handles on one temp path running N concurrent
  `Add`s (10 each) complete with zero `database is locked` errors; final row
  count == total inserted. Deterministic across `-count=20`.
- [x] A fresh `Store` reports `PRAGMA busy_timeout == 5000`,
  `db.Stats().MaxOpenConnections == 1`, and `PRAGMA journal_mode == delete`
  (rollback, not WAL).
- [x] The DSN pragma does not leak into the on-disk path; `VACUUM INTO` backup
  and `ResolveDBPath` unaffected (existing storage tests green).
- [x] No CGO added; full gate set passes.

## Failing Tests

Written during **design**, BEFORE build. Confirmed fail-first on the pre-fix
code: `busy_timeout = 0`, `MaxOpenConnections = 0`, and 20/20 concurrent writes
failed with `database is locked`.

- **`internal/storage/store_test.go`**
  - `"TestOpen_ConcurrentWritesAcrossHandles"` — asserts: two handles ×10
    concurrent `Add`s all succeed, no `database is locked`, final count == 20.
  - `"TestOpen_BusyTimeoutPragma"` — asserts: `busy_timeout == 5000`,
    `MaxOpenConnections == 1`, `journal_mode == delete`.

## Implementation Context

### Decisions that apply

- `DEC-038` — the concurrency policy this spec emits: busy_timeout(5000) +
  `_txlock=immediate` + `SetMaxOpenConns(1)`, rollback journal retained, WAL
  deferred.
- `DEC-001` — pure-Go / CGO-off SQLite (modernc); the fix must stay within the
  modernc DSN + `database/sql`.
- `DEC-002` / `DEC-021` — forward-only migrations and the pre-migration
  snapshot; the single-file backup must stay valid (argues against WAL).

### Constraints that apply

- `no-cgo` — modernc stays pure-Go; no CGO driver added.
- `storage-tests-use-tempdir` — the concurrency test uses `t.TempDir()`, never
  touches `~/.bragfile`.
- `errors-wrap-with-context` — `Open` continues to wrap errors with context.

### Out of scope (for this spec specifically)

- Switching to WAL journal mode (considered and deferred in DEC-038).
- Application-level retry loops or mutexes.
- Any schema/migration change.

## Notes for the Implementer

The DSN pragma form is `_pragma=busy_timeout(5000)` (verified on modernc
v1.53.0; the `_busy_timeout=` form is a no-op). `_txlock=immediate` is required
in addition to busy_timeout: busy_timeout alone leaves ~1/20 writes failing on
the deferred-transaction write-upgrade deadlock. Keep pragmas ONLY in the
`sql.Open` DSN — pass the clean `path` to `backupBeforeMigrations` and the
dev/prod guard.

---

## Build Completion

- **Branch:** `fix/spec-062-sqlite-concurrency`
- **PR (if applicable):** see PR opened against `main`.
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - `DEC-038` — SQLite concurrency policy (busy_timeout + IMMEDIATE + single
    conn, rollback retained, WAL deferred)
- **Deviations from spec:**
  - Added `_txlock=immediate` beyond the audit's headline recommendation
    (busy_timeout + single-conn). Justified: busy_timeout alone left ~1/20
    concurrent writes failing on the deferred-transaction upgrade deadlock;
    IMMEDIATE takes the write lock at BEGIN so busy_timeout can serialize
    writers. All of this package's transactions are writes, so it costs
    nothing. Journal mode still unchanged (rollback), so the deviation stays
    within DEC-038's policy envelope.
- **Follow-up work identified:**
  - Revisit WAL only if a workload needs concurrent readers during long writes
    (would require updating the backup docs; own DEC).

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing in the scaffold; the one discovery was that busy_timeout alone is
   insufficient (the deferred-transaction upgrade deadlock), which the
   fail-first test surfaced immediately.

2. **Was there a constraint or decision that should have been listed but
   wasn't?**
   — No; `no-cgo`, `storage-tests-use-tempdir`, and the DEC-001/002/021
   backup lineage covered it.

3. **If you did this task again, what would you do differently?**
   — Reach for `_txlock=immediate` from the start rather than confirming via
   the 1/20 residual failure — though the test-first loop caught it cleanly.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — Outcome is solid: two `Store` handles/processes now WAIT and succeed
   (zero `database is locked`) via `busy_timeout(5000)` + `_txlock=immediate` +
   `SetMaxOpenConns(1)`, journal mode unchanged. The one lesson is reaching for
   `_txlock=immediate` from the start rather than finding the deferred-txn
   upgrade deadlock via the residual 1/20 failure.

2. **Does any template, constraint, or decision need updating?**
   — No. DEC-038 already records the full policy (busy_timeout + IMMEDIATE +
   single-conn, rollback retained, WAL deferred) and its backup-vs-WAL rationale.

3. **Is there a follow-up spec I should write now before I forget?**
   — None now. WAL is deliberately deferred (DEC-038); revisit only if a
   workload needs concurrent readers during long writes, which would also
   require updating the single-file backup docs — its own future DEC.

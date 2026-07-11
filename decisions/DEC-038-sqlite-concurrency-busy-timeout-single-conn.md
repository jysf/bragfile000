---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-038
  type: decision
  confidence: 0.86
  audience:
    - developer
    - agent
    - operator

agent:
  id: claude-opus-4-8
  session_id: null

project:
  id: PROJ-005
repo:
  id: bragfile

created_at: 2026-07-10
supersedes: null
superseded_by: null

tags:
  - storage
  - sqlite
  - concurrency
  - busy-timeout
  - mcp
  - reliability
  - no-cgo
---

# DEC-038: SQLite concurrency policy — busy_timeout(5000) + IMMEDIATE transactions + single pooled connection, rollback journal retained (WAL deferred)

## Decision

`storage.Open` opens the database with a DSN that sets
`busy_timeout(5000)` and `_txlock=immediate`, and calls
`db.SetMaxOpenConns(1)`. The journal mode stays the default rollback
(`delete`) — WAL is considered and deliberately deferred. Concretely the DSN
is `path + "?_pragma=busy_timeout(5000)&_txlock=immediate"`; the on-disk path
stays clean (pragmas live only in the `sql.Open` DSN).

## Context

A pre-release storage audit (HIGH) found `storage.Open` calling
`sql.Open("sqlite", path)` with a bare path: no busy timeout, no connection
cap, default deferred transactions. Two `Store` handles on one file doing
concurrent `Add`s failed immediately with `database is locked (5)
(SQLITE_BUSY)` — a second writer's contention returned in ~30µs with no wait,
so ~39/40 writes were lost.

This is reachable and **amplified by v0.5.0**: `brag mcp serve` (a headline
feature) holds ONE long-lived `Store` for the whole process while the user
also runs `brag add` from a shell, and Claude Code hooks fire `brag add` — any
overlap failed. Concurrent MCP tool calls could also self-conflict because
`database/sql` may open several pooled connections.

The driver is `modernc.org/sqlite` (pure Go, CGO-off per DEC-001 / the
`no-cgo` constraint), so any fix must stay within that driver's DSN and the
`database/sql` API — no CGO driver, no external `sqlite3` tooling.

Three sub-choices had to be made: how a contended writer waits, whether to
serialize the process's own pool, and whether to switch journal mode to WAL.

## Alternatives Considered

- **Option A: busy_timeout alone (no IMMEDIATE), keep deferred transactions.**
  - What it is: append only `_pragma=busy_timeout(5000)` and cap the pool.
  - Why rejected: insufficient. `database/sql`'s default transaction is
    DEFERRED — it takes a read lock and only upgrades to a write lock on the
    first write. When two handles both hold a read lock and both try to
    upgrade, SQLite deadlocks and busy_timeout returns SQLITE_BUSY
    **immediately** (it cannot wait out a deadlock). The concurrency test
    proved this: busy_timeout alone still lost ~1/20 concurrent writes,
    intermittently. `_txlock=immediate` takes the write lock at BEGIN, turning
    write-write contention into a plain wait busy_timeout resolves. Every
    transaction in this package is a write, so serializing at BEGIN costs
    nothing.

- **Option B: switch to WAL journal mode.**
  - What it is: `PRAGMA journal_mode=WAL` for better reader/writer
    concurrency.
  - Why rejected (deferred, not dismissed): WAL adds `-wal` and `-shm` sidecar
    files. The tutorial documents a bare `cp` of the single DB file as a
    backup, and the pre-migration snapshot semantics assume a single file;
    WAL would silently make a bare-`cp` backup inconsistent (uncheckpointed
    frames live in `-wal`). That reopens a documented backup concern for no
    benefit here: busy_timeout + IMMEDIATE + single-conn already resolves the
    reported lock failures for a local, single-user, low-contention CLI. WAL
    is the right lever only if a future workload needs genuine concurrent
    readers during long writes — recorded as the revisit trigger below.

- **Option C: application-level mutex / retry loop in Go.**
  - What it is: wrap every write in a package-level `sync.Mutex` and/or retry
    SQLITE_BUSY in application code.
  - Why rejected: a mutex only serializes ONE process; the real scenario is
    two processes (mcp serve + shell), which a Go mutex cannot coordinate. A
    hand-rolled retry loop reimplements exactly what SQLite's busy_timeout
    already does correctly, with more code and more edge cases.

- **Option D (chosen): busy_timeout(5000) + `_txlock=immediate` +
  SetMaxOpenConns(1), rollback journal retained.**
  - What it is: contended writers WAIT up to 5s (busy_timeout) after taking
    the write lock at BEGIN (IMMEDIATE); the process's own pool is capped at a
    single connection so `database/sql` can't self-conflict; journal mode
    unchanged so the documented backup story is untouched.
  - Why selected: it fixes the reported cross-process failures deterministically
    (0 lost writes across 20×20 concurrent writes), stays pure-Go within the
    modernc driver, adds no dependency and no schema change, and leaves the
    single-file backup invariant intact. 5000ms is generous for a local CLI's
    microsecond-scale contention windows; SetMaxOpenConns(1) is free for a
    single-user tool and a stdio MCP server with no throughput concern.

## Consequences

- **Positive:** Concurrent access from `brag mcp serve` + shell `brag add` +
  Claude Code hooks now WAITS and succeeds instead of failing with `database
  is locked`. No CGO, no new dependency, no migration, no schema change. The
  single-file backup story (bare `cp`, and the `VACUUM INTO` pre-migration
  snapshot) stays valid.
- **Negative:** Every explicit transaction now takes the write lock at BEGIN
  (IMMEDIATE), so concurrent write transactions are fully serialized. This is
  correct for this package (all its transactions are writes) but means no
  write parallelism — acceptable for a single-user local CLI, and the whole
  point under contention.
- **Negative:** A pathological >5s writer would still surface SQLITE_BUSY.
  For this workload (tiny single-row inserts) that is effectively unreachable;
  the value is tunable if a real workload ever approaches it.
- **Neutral:** The on-disk path must stay pragma-free — pragmas live only in
  the `sql.Open` DSN. `config.ResolveDBPath`, `backup.go` (`VACUUM INTO`), and
  the dev/prod migrate guard all continue to receive the clean `path`; only
  `sql.Open` sees the DSN. Enforced by keeping `dsn` a local in `Open`.

## Validation

Right if:
- Two separate `*Store` handles on one temp-DB path running N concurrent
  `Add`s (10 per handle) complete with **zero** `database is locked` errors and
  a final row count equal to the total inserted — deterministically, across
  repeated runs (`-count=20`). (The load-bearing assertion;
  `TestOpen_ConcurrentWritesAcrossHandles`.)
- A fresh `Store` reports `PRAGMA busy_timeout == 5000`,
  `db.Stats().MaxOpenConnections == 1`, and `PRAGMA journal_mode == delete`
  (rollback, not WAL). (`TestOpen_BusyTimeoutPragma`.)
- The pre-migration backup (`VACUUM INTO`) and DB-path resolution are
  unaffected — the DSN pragma never leaks into the file path (existing storage
  tests, all green).

Revisit if:
- A workload needs genuine concurrent readers during long writes, or write
  throughput becomes a bottleneck — then reconsider WAL (Option B), which
  requires updating the backup docs to checkpoint or copy the `-wal`/`-shm`
  sidecars, likely with its own DEC.
- Legitimate contention ever exceeds 5s (raise the timeout, or add bounded
  application-level retry on top).

Confidence: 0.86. The mechanism is well-understood and empirically verified:
the DSN pragma syntax was confirmed by building a scratch program and reading
back `PRAGMA busy_timeout` (`_pragma=busy_timeout(5000)` works on
modernc.org/sqlite v1.53.0; the `_busy_timeout=` form is a no-op), and
`_txlock` is honored (an invalid value errors `unknown _txlock`). The fix
moved the concurrency test from 20/20 failing → 0 failing across 20 repeats.
The residual uncertainty (below 0.9) is the `_txlock=immediate` addition
beyond the audit's headline recommendation (busy_timeout + single-conn): it is
necessary — busy_timeout alone left ~1/20 writes failing on the deferred-
transaction upgrade deadlock — but it is a judgment call to serialize all
transactions at BEGIN rather than change each write's `BeginTx` call site. It
is above §14's 0.7 threshold, so no new open question is required.

## References

- Related specs: SPEC-062 (emits this DEC; the fix + failing-test-first).
- Related decisions:
  - DEC-001 (pure-Go / CGO-off SQLite via modernc) — constrains the fix to the
    modernc DSN and `database/sql`; no CGO driver.
  - DEC-002 (forward-only, irreversible migrations) — why the single-file
    pre-migration backup exists and must stay valid (argues against WAL here).
  - DEC-021 (a failed pre-migration snapshot aborts Open) — the backup path
    this DEC must not disturb.
  - DEC-024 (MCP server) — the long-lived `Store` that amplifies the bug.
- Related constraints: `no-cgo`, `storage-tests-use-tempdir`,
  `errors-wrap-with-context`, `timestamps-in-utc-rfc3339` (unaffected).
- External docs: modernc.org/sqlite DSN pragmas (`_pragma`, `_txlock`);
  SQLite `busy_timeout`, `BEGIN IMMEDIATE`, and rollback-journal locking.

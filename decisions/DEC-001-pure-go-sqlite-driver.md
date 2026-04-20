---
insight:
  id: DEC-001
  type: decision
  confidence: 0.90
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-7
  session_id: null

project:
  id: PROJ-001
repo:
  id: bragfile

created_at: 2026-04-19
supersedes: null
superseded_by: null

tags:
  - storage
  - sqlite
  - build
  - distribution
---

# DEC-001: Use the pure-Go SQLite driver (`modernc.org/sqlite`)

## Decision

`brag` uses `modernc.org/sqlite` as its SQLite driver. No CGO anywhere
in the build.

## Context

The MVP's success criteria include `brew install bragfile` shipping
working macOS binaries for arm64 and x86_64; STAGE-004 plans a
goreleaser pipeline that will also produce Linux arm64 and x86_64.
Go's canonical SQLite bindings (`mattn/go-sqlite3`) are CGO-based,
which means every target arch needs a matching C toolchain at release
time and clean cross-compilation is painful. The project also has no
performance ceiling on the order of what CGO-SQLite provides — the
user is a single person writing a few rows per day.

Triggered by: planning STAGE-001 (storage spec) and knowing STAGE-004
depends on a working cross-compile.

## Alternatives Considered

- **Option A: `mattn/go-sqlite3` (CGO)**
  - What it is: The most widely used Go SQLite driver. Wraps the C
    SQLite library via CGO.
  - Why rejected: CGO complicates goreleaser across arches.
    Static-linking SQLite requires per-arch toolchains in CI. Binary
    size is larger. The performance advantage (it is measurably
    faster) is irrelevant at this scale.

- **Option B: bbolt / pure-Go KV store**
  - What it is: Skip SQL entirely; use an embedded key-value store.
  - Why rejected: The spec already uses FTS5 for `search` (STAGE-002)
    and rule-based group-by counts for `summary` (STAGE-003). Both
    are order-of-magnitude easier in SQL than in a KV store, and
    reinventing them is not in scope.

- **Option C (chosen): `modernc.org/sqlite` (pure Go)**
  - What it is: A pure-Go translation of SQLite's C source. No CGO.
    Supports FTS5.
  - Why selected: Static Go binary, trivial goreleaser cross-compile,
    no toolchain headaches, adequate performance for a personal CLI.
    Compatible with the `database/sql` interface the `internal/storage`
    package will use.

## Consequences

- **Positive:** Release pipeline is a plain `go build` matrix. No CGO
  toolchain required in CI. Homebrew bottle is a single file. Developer
  env needs nothing beyond `go`. FTS5 is still available for STAGE-002.
- **Negative:** The driver is slower than `mattn/go-sqlite3` (order of
  2–3x on write-heavy workloads per published benchmarks). At MVP
  scale (single user, hundreds of rows) the user cannot perceive this.
  If the tool ever becomes multi-tenant or ingests external streams,
  revisit.
- **Neutral:** Adds ~4MB to the binary compared to a CGO build that
  dynamically links. Acceptable for a distributable CLI.

## Validation

This decision is right if:
- `goreleaser` builds all four target binaries in CI without a
  per-arch C toolchain.
- No user-perceivable latency on any operation at MVP scale.

Revisit if:
- The driver stops being maintained (currently active).
- A performance regression appears at personal-use scale (hundreds of
  rows). Unlikely.
- We decide to ship a server/multi-user variant where write throughput
  starts to matter.

## References

- Related specs: SPEC-002 (storage + migrations)
- Related decisions: DEC-002 (embedded migrations, made easier by this)
- External docs:
  - https://pkg.go.dev/modernc.org/sqlite
  - https://goreleaser.com/

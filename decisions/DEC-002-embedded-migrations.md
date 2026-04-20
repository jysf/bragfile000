---
insight:
  id: DEC-002
  type: decision
  confidence: 0.85
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
  - migrations
  - build
---

# DEC-002: Roll our own migrations; embed SQL files via `embed.FS`

## Decision

Schema migrations are plain numbered `.sql` files living in
`internal/storage/migrations/`, compiled into the binary via Go's
`embed.FS` and applied in lexical order by `storage.Open`. Applied
versions are tracked in a `schema_migrations` table. No external
migration library.

## Context

`brag` needs a migration strategy for SPEC-002 onward: the initial
schema ships in STAGE-001, FTS5 virtual table and triggers arrive in
STAGE-002, and any future schema evolution (tag normalization, soft
delete, attachments) will add more migrations. The user runs `brag`
from any terminal with no config or command-line migration step —
"upgrade" means "a newer binary", and the next CLI invocation must
bring the DB up to date silently.

Triggered by: planning SPEC-002 in STAGE-001.

## Alternatives Considered

- **Option A: `golang-migrate/migrate`**
  - What it is: The de-facto Go migration library. Supports many
    databases, up/down, file or embedded sources.
  - Why rejected: It's a dependency aimed at multi-DB server apps
    with operator-driven migrations. We have one DB type, one user,
    no down-migrations policy, and no CLI flag surface for migrations.
    The library's surface area is much larger than we need.

- **Option B: `pressly/goose`**
  - What it is: Similar intent to `golang-migrate`, smaller, Go-first.
  - Why rejected: Same reason — a dependency we don't need. Also pushes
    a particular migration-file format.

- **Option C: GORM / sqlc / ORM-managed schema**
  - Why rejected: No ORM; we use `database/sql` directly. Out of scope
    for MVP.

- **Option D (chosen): ~50 lines of migration code + embedded SQL**
  - What it is: `embed.FS` pointing at `migrations/*.sql`; `Open` reads
    `schema_migrations`, diffs against the embedded filenames, applies
    each missing one inside a transaction together with the tracking
    insert.
  - Why selected: Zero runtime dependencies, all migration logic is
    visible in the storage package, migrations ship automatically with
    the binary (a new `brag` version upgrades on first run).

## Consequences

- **Positive:** Zero migration-related dependencies. The author can
  read every line of the migration runner. `embed.FS` means there is
  no "did you forget to copy the migrations?" deploy failure mode;
  the schema travels inside the binary.
- **Negative:** No `brag migrate down`. If we ever need to reverse a
  bad migration we must write a forward fix. Acceptable for this tool.
- **Negative:** We have to write (and test) the tiny migration runner
  ourselves. Cost: roughly one afternoon of spec + tests.
- **Neutral:** Migration filenames become part of the contract — they
  must never be renamed post-release.

## Validation

Right if:
- Migration bugs remain at zero or near-zero across releases.
- `brag` on a fresh machine auto-creates the DB and applies all
  migrations on first run.
- No user ever has to think about migrations.

Revisit if:
- We add a second database backend (unlikely).
- Migration logic starts growing features (conditionals, data
  backfills) that push it past the "tiny and obvious" threshold.

## References

- Related specs: SPEC-002 (storage + migrations)
- Related decisions: DEC-001 (pure-Go driver — migration runner uses
  `database/sql`)
- External docs:
  - https://pkg.go.dev/embed
  - Discussion of golang-migrate vs roll-your-own:
    https://brandur.org/fragments/go-embed-migrations

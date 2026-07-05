---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-026
  type: decision
  confidence: 0.85                    # honest: the policy + seams are
                                       # high-confidence (hermetic tests cover
                                       # every branch); the residual soft spot
                                       # is the dev-build string heuristic —
                                       # it keys on version markers, which a
                                       # future release-naming change could
                                       # surprise (see Validation).
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

project:
  id: PROJ-003
repo:
  id: bragfile

created_at: 2026-07-05
supersedes: null
superseded_by: null

tags:
  - storage
  - migrations
  - safety
  - dev-prod-isolation
  - guardrail
---

# DEC-026: dev-binary DB-isolation guardrail — refuse to migrate prod from an unreleased build

## Decision

`storage.Open` **refuses to apply a pending migration to the real
`~/.bragfile` database when the running binary is an unreleased/dev build**,
unless the operator sets `BRAG_ALLOW_DEV_PROD_MIGRATE=1`. Concretely, the
guard returns a wrapped `storage.ErrDevProdMigrate` (mapped by `main` to
exit 1) — applying no migration and writing no backup — when **all** of:

1. the build is a **dev build** — `main.version` is empty, `"dev"`, or
   contains a `-dev` / `-dirty` / `-snapshot` marker (goreleaser injects a
   clean semver on release; a release candidate like `0.3.0-rc1` is a real
   artifact and is **not** a dev build);
2. the resolved DB path **is the production default** (`config.DefaultDBPath()`
   — the real `~/.bragfile/db.sqlite`), not a `--db` / `BRAGFILE_DB`
   throwaway; and
3. the DB is **established with work pending** — `applied > 0 && pending > 0`,
   the exact DEC-021 backup trigger; and
4. the override env var is **unset**.

All other cases pass through unchanged: released builds, the override, a
throwaway `--db`, a brand-new default DB (`applied == 0`), and an up-to-date
default DB (`pending == 0`, the read-only case).

The guard runs in `Open` **before** `backupBeforeMigrations` and
`applyMigrations`, so a dev binary against prod refuses cleanly rather than
snapshot-then-migrate.

## Context

At the v0.2.0 cut a **dev binary irreversibly migrated the production
`~/.bragfile`** (`0003` + `0004` applied) because dev/prod isolation was a
procedural discipline with no code enforcement (STAGE-008). DEC-021's
migration auto-backup (SPEC-036) made that mutation *recoverable* but does
not *prevent* it. The cross-project retrospective (2026-07-04) flagged this
as the one production escape with **no code guardrail** (action-register R3).
This DEC is the prevention; DEC-021 remains the belt behind it.

## Alternatives considered

- **Backup-only (status quo, DEC-021).** Rejected as insufficient: it turns
  an irreversible mutation into a recoverable one, but the operator still has
  to notice and restore. Prevention beats recovery for a footgun this sharp.
- **Block *all* dev-build writes to prod (not just migrations).** Rejected as
  too broad: reading/adding to prod with a dev build is a normal dogfooding
  act; only the *irreversible forward-only migration* (DEC-002) is the hazard.
- **Detect "dev" by the goreleaser ldflag being unset only** (`version ==
  "dev"`). Rejected as too narrow: `goreleaser build --snapshot` injects
  `X.Y.Z-SNAPSHOT-<sha>` and a dirty tree yields `-dirty` — both are
  unreleased and must be caught. Hence the marker list.
- **Guard in the CLI layer.** Rejected: the pending-migration state and the
  apply step both live in `storage.Open`; the guard must run pre-apply, and
  `Open` is the single choke point (there is no shared CLI open helper). The
  version + override are plumbed in via injectable package seams (§9).

## Consequences

- New injectable seams in `internal/storage`: `SetBuildVersion` (called once
  by `main`), `defaultDBPathFn` (defaults to `config.DefaultDBPath`), and
  `lookupEnv` — all substitutable in hermetic tests so no test touches the
  real `~/.bragfile` (`storage-tests-use-tempdir`).
- `internal/storage` now imports `internal/config` (one-directional; config
  is a stdlib-only leaf — no cycle).
- `main` gains a `storage` import to set the version and map
  `ErrDevProdMigrate` to exit 1.

## Validation

Hermetic tests (`internal/storage/devguard_test.go`) cover every branch —
dev+prod+pending → refuse (no migration, no sidecar); released / override /
throwaway / fresh / up-to-date → migrate — plus the `isDevBuild` classifier
(`""`, `dev`, `-SNAPSHOT`, `-dirty` are dev; `0.3.0`, `v0.3.0`, `0.3.0-rc1`
are not). **Residual soft spot:** the dev-build test keys on version-string
markers; a future change to how releases are named/stamped could reclassify a
build. If release naming changes, revisit `isDevBuild`.

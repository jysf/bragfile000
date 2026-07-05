---
# Maps to ContextCore task.* semantic conventions.

task:
  id: SPEC-044
  type: story
  cycle: build                     # design + build delivered together (retro R3)
  blocked: false
  priority: high                   # retro P1: a prod-DB safety gap with no code guardrail
  complexity: M

project:
  id: PROJ-003
  stage: STAGE-010                 # hardening spec routed to STAGE-010 (see Notes — stage-fit caveat)
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8
  created_at: 2026-07-05

references:
  decisions: [DEC-026, DEC-021]
  constraints: [storage-tests-use-tempdir, no-sql-in-cli-layer, stdout-is-for-data-stderr-is-for-humans]
  related_specs: [SPEC-036]
---

# SPEC-044: dev-binary DB-isolation guardrail

> **Design + build delivered together (retro R3).** Upgraded from the #70
> draft scaffold. Emits **DEC-026** (the policy) and implements the guard in
> `storage.Open`. Route: the register said "STAGE-009 or 010"; landed as a
> STAGE-010 hardening spec (stage-fit caveat in Notes).

## Context

At the v0.2.0 cut, a **dev binary irreversibly migrated the production
`~/.bragfile` DB** because the dev/prod isolation discipline was procedural,
not enforced (STAGE-008). DEC-021's migration auto-backup (SPEC-036) was the
*after-the-fact* belt — it makes the mutation recoverable but does not
*prevent* a dev build from migrating prod. The retro flagged this as the one
production escape with **no code guardrail**.

## Goal

When an **unreleased/dev build** would apply a pending migration to the
**real `~/.bragfile/db.sqlite`**, refuse — unless an explicit override is set —
so a dev binary can never silently migrate prod. Read-only commands and
released binaries are unaffected.

## Proposed approach (to lock at design)

- **"Dev build" detection.** `main.version` is unset / `dev` / carries a
  `-dev`/`-dirty`/`-SNAPSHOT` suffix (goreleaser injects a clean semver on
  release). Reference it through an injectable package var (per the §9
  os-state seam) so tests can substitute it.
- **Guard condition.** dev build **AND** resolved DB path is the real
  `~/.bragfile/db.sqlite` (not a `--db`/`BRAGFILE_DB` throwaway) **AND** there
  is ≥1 pending migration → `storage.Open` returns a wrapped `ErrUser`
  explaining the override; applies **no** migration.
- **Override.** `BRAG_ALLOW_DEV_PROD_MIGRATE=1` (env) re-permits it for the
  intentional case.
- **Carve-outs.** Read-only opens (no pending migration) are unaffected; the
  SPEC-036 auto-backup remains the belt behind this suspenders.
- **DEC-026** (to emit at design): what counts as a "dev build", the override
  mechanism, and the read-only carve-out (confidence-rated per §14).

## Acceptance Criteria

- [x] Dev build + real `~/.bragfile` (established, `applied>0 && pending>0`) →
      `Open` returns wrapped `storage.ErrDevProdMigrate` (→ exit 1) naming the
      override; **no** migration applied; **no** backup sidecar written.
- [x] Released build (clean version), OR `BRAG_ALLOW_DEV_PROD_MIGRATE=1`, OR a
      throwaway `--db`/`BRAGFILE_DB` path → unchanged (migration applies,
      DEC-021 backup fires as before).
- [x] Brand-new default DB (`applied==0`) by a dev build → unaffected (nothing
      established to protect — mirrors the DEC-021 carve-out).
- [x] Up-to-date default DB (`pending==0`, read-only case) by a dev build →
      unaffected.
- [x] `isDevBuild` classifies `""`/`dev`/`-SNAPSHOT`/`-dirty` as dev and
      `0.3.0`/`v0.3.0`/`0.3.0-rc1` as released.
- [x] Hermetic tests (`t.TempDir()`, injected `buildVersion` +
      `defaultDBPathFn` + `lookupEnv`) cover every branch; no test touches the
      real `~/.bragfile` (`storage-tests-use-tempdir`).

## Failing Tests

- **`internal/storage/devguard_test.go`**
  - `TestDevProdGuard_DevBuildRefusesPendingMigrationOnProdDB` — refuse; no
    migration, no sidecar.
  - `TestDevProdGuard_ReleasedBuildMigrates` / `_OverrideMigrates` /
    `_ThrowawayPathUnaffected` / `_FreshDefaultDBUnaffected` /
    `_UpToDateDefaultDBUnaffected` — the pass-through cases.
  - `TestIsDevBuild` — the classifier literal.

## Implementation Context

### Decisions that apply

- `DEC-026` — the guard policy (this spec emits it): dev build + prod default
  + `applied>0 && pending>0` + no override → refuse.
- `DEC-021` — the migration auto-backup; the guard fires on the *same*
  `applied>0 && pending>0` trigger and runs *before* the backup, so a refused
  open snapshots nothing.

### Constraints that apply

- `storage-tests-use-tempdir` — the injected `defaultDBPathFn` lets tests
  treat a `t.TempDir()` file as "the production default"; nothing touches
  `~/.bragfile`.
- `no-sql-in-cli-layer` — the guard is in `internal/storage`; `main` only sets
  the version and maps the error to an exit code.
- `stdout-is-for-data-stderr-is-for-humans` — `ErrDevProdMigrate` is a user
  error (stderr, exit 1).

### Out of scope

- Blocking non-migration writes to prod from a dev build (reads/adds are
  normal dogfooding; only the irreversible migration is the hazard).
- Symlink-resolving path comparison (`samePath` is lexical `filepath.Clean`;
  documented).

## Notes for the Implementer

- **Stage home caveat.** R3 is a *hardening* spec, off-theme for STAGE-010's
  impact-read-surface; landed here as a ride-along. Reassign if a dedicated
  hardening stage is opened.
- Detecting the real `~/.bragfile` goes through `config.DefaultDBPath()` (the
  same resolver the CLI uses) — the guard asks "is this the default path?",
  never re-hardcodes `~/.bragfile`.

---

## Build Completion

- **Branch:** `feat/spec-044-dev-binary-db-isolation-guardrail`
- **PR:** references retro R3 (action-register).
- **All acceptance criteria met?** yes — 7 hermetic tests (`devguard_test.go`)
  green; full suite 581 pass; gofmt/vet clean; `CGO_ENABLED=0` build ok.
- **New decisions emitted:** `DEC-026` — dev/prod-migration guard policy.
- **Deviations from spec:** none. (The guard reuses the DEC-021
  `applied>0 && pending>0` trigger rather than a bespoke "non-empty" check —
  same set, one source of truth.)
- **Follow-up work identified:** none. `internal/storage` now imports
  `internal/config` (one-directional; config is a stdlib leaf).

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — One real design refinement the sketch missed: the guard must NOT fire on
   a *brand-new* default DB (`applied==0`), only on an *established* one —
   otherwise a dev build could never create `~/.bragfile`. Aligning the
   trigger with DEC-021's `applied>0 && pending>0` resolved it cleanly.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — The `storage → config` import direction is new; worth noting it's
   cycle-free (config is a stdlib-only leaf). Captured in DEC-026.

3. **If you did this task again, what would you do differently?**
   — Nothing. Putting the guard in `Open` (the single pre-apply choke point)
   avoided a 10-site CLI refactor and kept the policy in one testable place.

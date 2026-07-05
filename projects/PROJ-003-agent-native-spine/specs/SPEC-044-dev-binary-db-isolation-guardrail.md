---
# Maps to ContextCore task.* semantic conventions.
# DRAFT SCAFFOLD (retro R3) — frame stub, not yet through design.

task:
  id: SPEC-044
  type: story
  cycle: frame                     # DRAFT — proposed from cross-project-retro R3; needs a design session
  blocked: false
  priority: high                   # retro P1: a prod-DB safety gap with no code guardrail
  complexity: M

project:
  id: PROJ-003
  stage: STAGE-010                 # TENTATIVE — R3 is a hardening spec; stage is a coordinator call (see Notes)
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8
  created_at: 2026-07-05

references:
  decisions: [DEC-021]             # + DEC-026 (dev-build policy) to be emitted at design
  constraints: [storage-tests-use-tempdir, timestamps-in-utc-rfc3339]
  related_specs: [SPEC-036]
---

# SPEC-044: dev-binary DB-isolation guardrail

> **DRAFT SCAFFOLD (retro R3).** Proposed from
> [`2026-07-04-action-register.md`](../../../docs/reports/cross-project/2026-07-04-action-register.md)
> item R3. Not yet designed — this stub captures the scope so it is not lost.
> Route through the normal design→build→verify→ship cycle when picked up.

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

## Acceptance Criteria (sketch)

- [ ] Dev build + real `~/.bragfile` + pending migration → `Open` returns a
      wrapped `ErrUser` naming the override; no migration applied; no sidecar.
- [ ] Released build (clean version), OR override set, OR a throwaway `--db`
      → unchanged behavior (migration applies, SPEC-036 backup fires).
- [ ] Read-only open of an up-to-date real DB by a dev build → unaffected.
- [ ] Hermetic tests (`t.TempDir()`, injected version var + injected
      home/DB-path seam) cover both branches; `storage-tests-use-tempdir`
      respected — the test never touches the real `~/.bragfile`.

## Failing Tests (to write at design)

- **`internal/storage/*_test.go`** — dev+real+pending → `ErrUser`, no
  migration; released/override/throwaway → applies; the "real path" is
  simulated by pointing the injected home/path seam at a `t.TempDir()` file
  the test treats as `~/.bragfile`.

## Notes for the Implementer

- **Stage home is a coordinator call.** R3 is a *hardening* spec, off-theme
  for STAGE-010's impact-read-surface. Options: (a) fold into STAGE-010 as a
  ride-along safety spec, or (b) open a small dedicated hardening stage. The
  `stage: STAGE-010` above is a placeholder.
- Detecting the real `~/.bragfile` must go through the same resolution
  `internal/config` already owns — do not re-hardcode `~/.bragfile`; ask the
  resolver whether the resolved path *is* the default.

---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-008                     # stable, zero-padded, repo-global (never reused)
  status: proposed                  # proposed | active | shipped | cancelled | on_hold
  priority: high                    # release-cutting stage; gates PROJ-002 close (brief target_ship 2026-06-26)
  target_complete: 2026-06-24       # leaves a 2-day buffer before the project's target_ship 2026-06-26

project:
  id: PROJ-002                      # parent project
repo:
  id: bragfile

created_at: 2026-06-12
shipped_at: null
---

# STAGE-008: polish and v0.2.0 release

## What This Stage Is

When this stage ships, **the feature-complete v0.2.0 codebase becomes a
released, installable thing**: the docs describe what the binary actually
does, the CHANGELOG records `[0.2.0]`, the schema migration can no longer
destroy an un-backed-up database without warning, and `v0.2.0` reaches the
public Homebrew tap so `brew upgrade jysf/bragfile/bragfile` moves a v0.1.x
install forward cleanly. Concretely: `docs/tutorial.md` gains a projects
walkthrough (tags landed in STAGE-006) and a WAL-safe backup recipe;
`docs/architecture.md`'s diagram and responsibilities table grow the
`brag project` command surface, the `0004` migration, and the
`projects`/`project_locations` tables; `docs/api-contract.md` gets a final
consistency + stability + version pass (its per-command entries already
landed per-spec in STAGE-007); the CHANGELOG `[0.2.0]` entry is written and
the `[0.1.0]` placeholder reconciled; a **migration auto-backup safety belt**
ships so that applying a schema-bumping migration to an existing DB copies it
to a timestamped sidecar first; and the release is cut `v0.2.0-rc1 → v0.2.0`
per AGENTS.md §4's dual-tag-on-same-commit rule. This is the **last stage of
PROJ-002** — no new feature surface, only the polish and mechanics that turn
"feature-complete on `main`" into "released and safe to upgrade into."

## Why Now

STAGE-006 (tags, shipped 2026-06-07) and STAGE-007 (projects, shipped
2026-06-12) together completed the entire v0.2.0 feature surface; nothing is
left to build, only to document, guard, and ship. Three reasons make this the
moment:

1. **The doc debt is now total and concrete.** STAGE-006/007 deliberately
   folded only *per-spec* doc updates (the premise-audit status-change case)
   and explicitly deferred the comprehensive sweep — the tutorial projects
   walkthrough and the architecture diagram/responsibilities refresh — to this
   stage. That deferred work is now fully scoped and can be done against a
   frozen surface rather than a moving one.

2. **A real migration-safety incident has occurred.** This session, a v0.2.x
   dev binary irreversibly migrated the **production** `~/.bragfile` DB
   (`0003` + `0004` applied) because the dev/prod isolation discipline was
   bypassed and **no backup existed**. The brief listed a migration-prompt
   safety belt as "optional, high safety-vs-cost"; the incident is the
   motivating evidence that promotes it from optional nicety to a stage success
   criterion. Whatever its final UX, v0.2.0 should not be able to mutate an
   un-backed-up DB silently.

3. **It is the only thing between feature-complete and a public release.** The
   v0.2.0 cut, the Homebrew formula bump, and the clean-upgrade guarantee all
   assume tags *and* projects exist and the docs match. They do now.

No external blockers — the Homebrew tap and all release secrets carry from
PROJ-001 (v0.1.0). The §4 release addenda earned at the v0.1.0 cut
(dual-tag-on-same-commit recovery; macOS Gatekeeper xattr workaround) apply
directly to this cut.

## Success Criteria

Concrete and re-verifiable at stage close:

- **Comprehensive doc sweep done.** `docs/tutorial.md` carries a projects
  walkthrough (`new` / `list` / `show` / `status` / `here` / `edit` /
  `archive` / `delete` + `brag add` cwd auto-fill) alongside the existing tags
  section; `docs/architecture.md`'s mermaid diagram and responsibilities table
  reflect the `brag project` command group, the `0004_add_projects` migration,
  and the `projects` + `project_locations` tables; `docs/api-contract.md`
  passes a final consistency + stability-guarantees + version pass. The
  doc-sweep specs run their premise-audit greps (status-change case) and
  reconcile each hit.
- **A WAL-safe backup recipe is documented.** The tutorial's backup guidance
  prefers `sqlite3 ~/.bragfile/db.sqlite ".backup 'backup.db'"` (or
  `VACUUM INTO`) over bare `cp`, and the CHANGELOG/upgrade workflow points at
  it. (Finding from framing: the binary never sets `PRAGMA journal_mode=WAL`,
  so bare `cp` of a quiescent DB is *currently* safe — the upgrade is
  robustness + future-proofing, not a live-bug fix; the recipe should still
  prefer `.backup`.)
- **CHANGELOG `[0.2.0]` is written.** It records tag normalization
  (DEC-015 supersedes DEC-004), the projects entity and full `brag project`
  surface (DEC-017/018/019/020), and the safety belt; the stale `[Unreleased]`
  section and the `[0.1.0]` `YYYY-MM-DD` placeholder + `bragfile000` compare
  links are reconciled.
- **The migration safety belt ships.** Applying a schema-bumping migration to a
  non-empty existing DB makes a timestamped backup copy of the DB file before
  the migration transaction runs. Non-interactive (does not break
  `brag add --json` / non-TTY pipelines). Tested. **(In scope — see Scope; a
  DEC lands if it encodes a durability choice.)**
- **v0.2.0 is tagged and released per §4.** `v0.2.0-rc1` validates the workflow,
  then `v0.2.0` is cut following the dual-tag-on-same-commit rule (default:
  delete the RC tag + release, then tag `v0.2.0` at the same commit); the
  Homebrew formula in `github.com/jysf/homebrew-bragfile` is bumped; a clean
  `brew upgrade` from a v0.1.x install is verified (or the verification
  limitation is documented — see Design Notes on the dev/prod DB story).
- **No regressions.** All STAGE-001..007 success criteria still hold;
  `go test ./...`, `gofmt -l .`, `go vet ./...`, and `CGO_ENABLED=0 go build
  ./...` are clean through every spec.

## Scope

### In scope

- **Comprehensive doc sweep.** `docs/tutorial.md` projects walkthrough + the
  WAL-safe backup recipe upgrade; `docs/architecture.md` diagram +
  responsibilities-table refresh (project command group, `0004` migration,
  `projects`/`project_locations` tables, the project Store methods); a final
  `docs/api-contract.md` consistency / stability / version pass. Each doc spec
  runs the premise-audit status-change grep (`grep -rn -i "project" docs/
  README.md` and the tag equivalent) and enumerates hits under `## Outputs`.
- **CHANGELOG `[0.2.0]`.** The release notes entry (literal-artifact-as-spec),
  plus reconciliation of the stale `[Unreleased]` `brag completion` line, the
  `[0.1.0]` placeholder date, and the compare-link repo slug.
- **Migration auto-backup safety belt.** A `storage.Open`-time guard that, when
  a pending migration would bump a non-empty existing DB, copies the DB file to
  a timestamped sidecar (e.g. `~/.bragfile/backups/db-<ts>.sqlite` via
  `VACUUM INTO` / `.backup`) before applying the migration transaction.
  Non-interactive by default. **Promoted from the brief's "optional" status to
  in-scope on the strength of this session's prod-DB incident.**
- **v0.2.0 release mechanics.** `v0.2.0-rc1` smoke cut → `v0.2.0` per §4;
  goreleaser; Homebrew formula bump; clean-upgrade verification.
- **A documented WAL-safe backup recipe** in the upgrade/backup workflow
  (folds into the doc-sweep spec).

### Explicitly out of scope

- **Automated daily backup (launchd LaunchAgent / `scripts/backup-db.sh`).**
  The PROJ-001 backlog item pairs a *documented recipe* (in scope, above) with
  an *unattended daily backup* (ops, not project code). The ops piece stays
  **out** — it is a dev-machine convenience, removable after PROJ-002, and is
  not needed for a correct release. May ship as an optional repo script only if
  a spec has spare room; default: leave backlogged. (Surfaced in Design Notes.)
- **Goals as a shipped type.** Paper-sketched only (DEC-015); no code.
  PROJ-003 candidate.
- **The `brag project tag` command surface.** Deferred at STAGE-007 (schema is
  ready; no command). Stays out unless dogfooding during this stage demands it
  — it is not a release blocker.
- **Reverse-direction / down migrations.** Forward-only, same as
  PROJ-001/006/007. The documented backup recipe + the new safety belt *are*
  the downgrade story.
- **macOS code signing + notarization.** Separately tracked
  (`docs/macos-notarization-checklist.md`); the xattr workaround documented at
  v0.1.0 still applies. Trigger conditions unchanged.
- **New feature surface of any kind.** This stage ships what STAGE-006/007
  built; it does not add to it.

## Spec Backlog

Framed as **four specs, all S/M (no L)**, splitting the doc sweep, the
CHANGELOG, the safety belt, and the release cut along their natural seams. The
doc sweep is the only M-risk item; it is held to M because `api-contract.md`'s
per-command entries already landed in STAGE-007 (this stage only does the
final pass), leaving the tutorial projects walkthrough + the architecture
diagram/table refresh as the real work. **L-watch on SPEC-034:** if the
tutorial walkthrough and the architecture refresh together read L at design,
peel the architecture diagram/responsibilities refresh into its own S spec
(it is mechanical and cleanly separable) rather than letting SPEC-034 grow to
L — mirroring the STAGE-007 SPEC-029→033 peel discipline.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [x] SPEC-034 (shipped on 2026-06-12) — **M** *(L-watch RESOLVED: held as
      one M, no peel — architecture refresh was mechanical)* — **Comprehensive doc sweep.**
      `docs/tutorial.md` projects walkthrough (as a `### subsection of §4`,
      NOT a new top-level § — `scripts/test-docs.sh` P7/R1 anchor on §9/§10,
      so renumbering would break tests; the §9 grep caught this) + §5
      WAL-safe backup recipe + the SPEC-036 auto-backup behavior;
      `docs/architecture.md` diagram + responsibilities refresh;
      `docs/api-contract.md` + `docs/data-model.md` consistency/version pass.
      No DEC (docs of shipped behavior; DEC-017..021 cited of record).
      22 grep-gradeable ACs + `scripts/test-docs.sh` exits 0. README bare-cp
      flagged as out-of-scope follow-up.
- [ ] SPEC-035 (not yet written) — **S** — **CHANGELOG `[0.2.0]`.** Write the
      `[0.2.0]` entry (tags normalization / projects / safety belt; DECs of
      record 015–020); reconcile the stale `[Unreleased]` `completion` line,
      the `[0.1.0]` `YYYY-MM-DD` placeholder, and the compare-link repo slug.
      Literal-artifact-as-spec. No DEC.
- [x] SPEC-036 (shipped on 2026-06-12) — **S/M** — **Migration auto-backup
      safety belt.** `storage.Open`-time guard: before applying a pending migration to an
      existing DB, snapshot it via `VACUUM INTO` through the open *sql.DB,
      then migrate. **DEC-021 emitted** (durability model): trigger
      discriminator `applied>0 && pending>0`; mechanism `VACUUM INTO`
      (pure-Go, WAL-safe, no sqlite3 CLI); failure ABORTS Open (never
      migrate un-backed-up); sidecar `<dbpath>.pre-<ver>.<UTC>.backup`;
      injectable `clock` seam (injectable-os-var WATCH → N=3); keep-all
      retention; silent. No new migration (schema_migrations stays 4).
      5 tests. Scaffolded out of plan order (ahead of doc sweep).
- [ ] SPEC-037 (not yet written) — **S** — **v0.2.0 release cut.**
      `v0.2.0-rc1` smoke → `v0.2.0` per §4 dual-tag rule; goreleaser; Homebrew
      formula bump; verify (or document the limitation on) a clean
      `brew upgrade` from v0.1.x. No DEC; pure release mechanics.

**Count:** 2 shipped / 0 active / 2 pending

**Complexity check:** 4 specs, all S/M by construction. One above the brief's
"~2–3 specs" estimate — the extra is the migration safety belt promoted from
optional-to-in-scope by the incident, and the release cut broken out from the
doc/CHANGELOG work it does not naturally combine with. The CHANGELOG (SPEC-035)
could fold into either the doc sweep (SPEC-034) or the release cut (SPEC-037)
to return to three specs if a coordinator prefers; kept separate here because
it is a clean literal-artifact spec with its own premise-audit (the placeholder
reconciliations) and conflating it with the M-risk doc sweep risks pushing that
spec toward L.

## Design Notes

Glue and cross-cutting direction. The weighty decision (the safety-belt
durability model) gets its own DEC at SPEC-036 design — **not written at
framing**, and **not pinned to a number** (next free is DEC-021, but earlier
specs do not emit DECs, so it should hold unless SPEC-034/035 surprise us). The
questions below are **surfaced for spec-design time, not decided here.**

### SURFACED design questions (resolve at spec design; do not decide now)

1. **Migration-prompt UX — auto-backup-then-migrate vs. prompt-and-confirm
   (DEC-021, SPEC-036).** Three candidate shapes, none chosen at framing:
   - **Auto-backup-then-migrate (recommended starting point).** Non-interactive;
     copies the DB to a timestamped sidecar before the migration tx, then
     proceeds. Pro: works in `--json`/non-TTY pipelines; strictly safer than
     the isolation discipline the incident bypassed; nothing to dismiss. Con:
     writes a backup file every upgrade (retention/pruning question).
   - **Prompt-and-confirm.** Block on a `[y/N]` before a schema-bumping
     migration. Pro: explicit user awareness. Con: breaks non-interactive
     callers (the `brag add --json` path, cron, CI), and a prompt is exactly
     the kind of guard a hurried user dismisses — the incident's isolation
     discipline was a "prompt" that got bypassed.
   - **Hybrid.** Auto-backup always; additionally prompt when a TTY is present.
   The choice determines where backups land, the retention policy (keep-last-N?
   none?), and whether DEC-021 is warranted (a durability decision likely is).
   Recommendation to weigh at design: lead with auto-backup-then-migrate for
   non-interactivity and incident-fit.

2. **Does v0.2.0 ship an automated backup, or only a documented manual
   recipe?** Framing recommendation: **documented WAL-safe recipe is in scope**
   (doc sweep) and **the migration-time auto-backup is in scope** (SPEC-036);
   the **unattended launchd daily backup stays out** (ops, not release code).
   Confirm at SPEC-034/036 design; the launchd piece can ship as an optional
   `scripts/backup-db.sh` only if a spec has spare room.

3. **Does the release reconcile the dev/prod DB story this session disrupted,
   or defer to PROJ-003?** Context: per the recalled prod-DB-upgrade note, the
   user's production `~/.bragfile` is **already v0.2.x** (the incident migrated
   it), so the "v0.1.x install upgrades cleanly" criterion can no longer be
   dogfooded on the user's own machine — clean-upgrade verification needs a
   throwaway v0.1.x DB, or the limitation must be documented. Separately, the
   "switch the bare `brag` / `~/go/bin` dev binary back to a clean prod story"
   cleanup is a loose end. Surface at SPEC-037 design: reconcile here, or
   explicitly defer to PROJ-003. Do not silently inherit the disrupted state.

4. **CHANGELOG reconciliations (SPEC-035).** The `[Unreleased]` section
   currently lists only `brag completion` — decide whether that folds into
   `[0.2.0]` or belonged to a prior tag; the `[0.1.0]` date is the
   `YYYY-MM-DD` placeholder (set the real date); and the compare links read
   `bragfile000` (confirm the canonical repo slug). These are premise-audit
   reconciliations, not new decisions.

### Premise-audit triggers (every doc/migration spec must reconcile)

Reference `projects/_templates/premise-audit.md`; each spec runs its own greps
at design and enumerates hits under `## Outputs`.

- **Status-change (doc) — the dominant case this stage.** The doc sweep changes
  the shipping status of projects + tags from "per-spec mentions" to "fully
  documented." Run `grep -rn -i "project" docs/ README.md` and the tag
  equivalent; enumerate each hit as "updates" or "stays." Watch specifically
  for the tutorial's "what's NOT there yet" framing and any STAGE-007-era
  "(STAGE-007 later spec)" forward-reference notes in `api-contract.md` that
  are now stale (e.g. the `brag project edit` location-editing note that
  pointed forward to SPEC-033, now shipped).
- **Count-bump (additive) — only if the safety belt touches a migration.**
  SPEC-036 should *not* add a migration (it is an `Open`-time file-copy guard,
  not schema). If a design iteration nonetheless adds a `0005_*`, the §12(a)
  literal-count assertions in `store_test.go` / `fts_test.go` (currently `4`,
  the `0001..0004` list) bump 4→5 at the same sites STAGE-007 exercised — run
  the migration through the real driver and compute the `want` list at design.
- **Inversion/removal.** The safety belt adds behavior to `storage.Open`; if it
  changes the open path's signature or error surface, enumerate existing
  `Open`/migration tests whose premise shifts as planned rewrites, not
  build-time discoveries.

### Architecture refresh — concrete scope (grounded at framing)

`docs/architecture.md` currently omits projects entirely. The refresh adds:
the `brag project` command group to the mermaid diagram; the
`migrations/0004_add_projects.sql` node to the `embed.FS` list; `projects` +
`project_locations` to the DB node; and a responsibilities-table update noting
the project Store methods (`CreateProject` / `GetProject` / `GetProjectByName`
/ `ListProjects` / `AddLocation` / `RemoveLocation` / `UpdateProject` /
`ArchiveProject` / `DeleteProject` / `ProjectForPath` / `EditLocations`).
**Brief drift to flag (see Bookkeeping):** the brief's Scope names a "new
`internal/projects` package," but no such package exists — projects live in
`internal/storage/project.go` + `internal/cli/project.go`. The architecture
refresh should describe the *actual* layout, not the brief's hypothetical one.

### WATCH-list carry-forward (from STAGE-007 close, verbatim)

This is the final stage of PROJ-002. Anything not codified by PROJ-002 close
either codifies at the **project** close (Prompt 1e) or is explicitly dropped —
do not codify mid-stage without the documented trigger. The inherited ledger:

- **Flag-default-explicitness in literal-artifact CLI specs — CODIFIED** at
  STAGE-007 close (AGENTS.md §12, During design). N=3 same-outcome
  (SPEC-026/028/029). Removed from WATCH. *(STAGE-008's CHANGELOG spec is a
  literal-artifact spec but ships no CLI flags, so it does not re-test this.)*
- **Injectable-os-var seam (`var getCwd = os.Getwd` / `addGetCwd`)** — N=2
  (SPEC-031, SPEC-032), same-outcome. Below the N=3 same-outcome bar; **carry
  forward.** STAGE-008 is docs/release-heavy and likely will *not* advance it
  (SPEC-036's safety belt touches the filesystem and *could* — a backup-path
  resolver may want an injectable seam). Candidate shape when earned: a
  one-line §9 testing-conventions note, *not* a blocking constraint.
- **cobra `SilenceErrors:true` + explicit `Fprintln` user-error pattern** —
  N=2 (SPEC-030, SPEC-031). Below the N=3 bar; **carry forward.** Unlikely to
  advance this stage (no new CLI command surface). Cheap fix when earned: a
  concrete example in Notes for the Implementer.
- **Contamination-heuristic exception for literal-artifact builds** — N=2
  (SPEC-026 origin; SPEC-027 confirming). **Carry forward, NOT codified.**
  STAGE-008's doc-sweep + CHANGELOG specs are literal-artifact-as-spec builds
  and the natural next test cases. Discriminator to preserve verbatim:
  "'nothing was unclear' is the expected honest output of a
  literal-artifact-as-spec build; distinguish honest-frictionless from
  mailed-in by whether the reflection surfaces ANY specific observation." Also
  a paired-opposing candidate (N=2 paired) if a future build mails in a
  contaminated "nothing was unclear" with no specific observation.
- **CLI recency-ordering tests need the §9 no-sleep SQL-backdate** — N=1
  (SPEC-030 build). Already implied by the codified §9 no-sleep rule; **carry
  forward** as a candidate one-line generalization to the CLI test layer if it
  recurs.
- **"AC-says-all-N → test each, don't lean on one shared-helper test"** — N=1
  (SPEC-029 verify). **Carry forward**; candidate premise-audit / spec-template
  nudge under Failing Tests.
- **"A design-prompt premise can be wrong — the §9 grep is the source of
  truth"** — N=1+ (SPEC-033). **Carry forward**; reinforces the
  already-codified §9 audit-grep family rather than adding a new rule.
- **Trust-but-verify agent push reports** — N=2 (SPEC-023); applied as a
  coordinator reflex throughout STAGE-006/007. Variant-specific (matters for
  `claude-plus-agents`); **carry forward.** Directly relevant this stage: the
  v0.2.0 release cut (SPEC-037) makes "pushed the tag / formula" claims that
  must be checked via `git ls-remote` / `gh release view`.
- **§13 fresh-session working-tree preservation** — N=2 (SPEC-024 +
  STAGE-007 close, where a parallel session's uncommitted
  `docs/framework-feedback/process-feedback.md` was preserved untouched).
  Below the N=3 bar; **carry forward.** (That same parallel-session file is
  modified in this repo's working tree right now — §13: do not touch it.)

### Bookkeeping (numbering + drift to flag)

- **Next free IDs:** SPEC-034 (this stage's specs are SPEC-034..037), DEC-021
  (consumed only if SPEC-036's safety belt earns a durability DEC — not pinned
  in prose per the standing "don't pin a not-yet-emitted DEC number" rule that
  bit the brief twice in earlier stages).
- **Brief STAGE-008 line — confirmed, with three drifts flagged for the
  coordinator (do not fix unilaterally during framing):**
  1. **Spec count.** The brief estimates "~2–3 specs"; this frame proposes 4
     (doc sweep / CHANGELOG / safety belt / release cut). The delta is the
     incident-promoted safety belt + a separated release cut. Conscious, not
     scope creep.
  2. **"internal/projects package" does not exist.** The brief's Scope (doc
     sweep) and the STAGE-006/007 enabling notes reference a "new
     `internal/projects` package"; projects actually live in
     `internal/storage/project.go` + `internal/cli/project.go`. The
     architecture refresh should document the real layout.
  3. **Stage name.** The brief's Stage Plan calls STAGE-008 "Polish +
     integration"; this file is titled "polish and v0.2.0 release" (the
     release is the defining deliverable). Minor; flagged for consistency.
- **Brief progress markers reconcile at project close.** Per the STAGE-006 and
  STAGE-007 closes, the brief's Stage Plan checkboxes ("0 shipped / 3 pending")
  are reconciled at **PROJ-002 close** (Prompt 1e), not at stage close — left
  as-is here for consistency, flagged for the project close.

## Dependencies

### Depends on

- **STAGE-007 (shipped 2026-06-12).** Projects are first-class; the
  `brag project` surface, `0004_add_projects` migration, and
  `projects`/`project_locations` tables are what the doc sweep documents and
  the safety belt protects. DEC-017/018/019/020 are the decisions of record the
  CHANGELOG lists.
- **STAGE-006 (shipped 2026-06-07).** Tag normalization (DEC-015 supersedes
  DEC-004); the tutorial's tag section already landed here, so the doc sweep
  only adds projects + the backup recipe on top.
- **PROJ-001 (shipped 2026-05-17).** All 14 DECs apply forward. Release-relevant
  directly: the §4 dual-tag-on-same-commit recovery rule and the macOS
  Gatekeeper xattr workaround (both earned at the v0.1.0 cut); DEC-002
  (forward-only embedded migrations — the surface the safety belt guards). The
  Homebrew tap and release secrets (`HOMEBREW_TAP_GITHUB_TOKEN`, etc.) are
  configured.
- **External: none new.** Backup is `VACUUM INTO` / `sqlite3 .backup` — plain
  `database/sql` / stdlib; goreleaser + the existing tap are already wired. Any
  new top-level dependency would need its own DEC.

### Enables

- **PROJ-002 close (Prompt 1e).** This is the last stage; shipping it makes the
  project closeable — the project-level reflection, the WATCH-list final
  disposition (codify or drop), and the brief's progress-marker reconciliation
  all run after this stage ships.
- **PROJ-003 (if framed).** A released, documented v0.2.0 with the polymorphic
  tags + first-class projects model is the substrate a third taggable type
  (`goals`) and richer cross-object views build on. The dev/prod DB cleanup
  deferred here (if deferred) lands as a PROJ-003 housekeeping item.

## Stage-Level Reflection

*Filled in when status moves to shipped. Run Prompt 1c (Stage Ship) in
FIRST_SESSION_PROMPTS.md to draft this.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <one-line items>

---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-034
  type: story                      # epic | story | task | bug | chore
  cycle: design                    # frame | design | build | verify | ship
  blocked: false
  priority: medium
  complexity: M                    # S | M | L  (L means split it)

project:
  id: PROJ-002
  stage: STAGE-008
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-06-12

references:
  decisions: [DEC-017, DEC-018, DEC-019, DEC-020, DEC-021]
  constraints: []
  related_specs: [SPEC-027, SPEC-028, SPEC-029, SPEC-030, SPEC-031, SPEC-032, SPEC-033, SPEC-036]
---

# SPEC-034: comprehensive doc sweep — projects, tags, backup

## Context

STAGE-006 (tags) and STAGE-007 (projects) shipped the entire v0.2.0
feature surface, but deliberately folded only the **per-spec** doc
updates (the premise-audit status-change case) and deferred the
**comprehensive sweep** to this stage (STAGE-008, the v0.2.0 release
stage). That deferred work is now scoped against a frozen surface:

- `docs/tutorial.md` has **no projects walkthrough** and its §5 backup
  guidance is a bare-`cp` recipe (WAL-unsafe and now incomplete — it
  predates the SPEC-036 migration auto-backup safety belt).
- `docs/architecture.md`'s diagram + responsibilities table omit
  **projects entirely** (and, on inspection, the STAGE-006 `brag tags`
  CLI group and the `tags`/`taggings` tables too) and the `0004`
  migration and `backup.go`.
- `docs/api-contract.md` already carries the per-command project
  entries (landed per-spec in STAGE-007); this spec is its **final
  consistency / stability / version pass**, not a rewrite — it fixes a
  stale forward-reference and a missing DEC reference.
- `docs/data-model.md` already documents `projects` +
  `project_locations`; this spec **confirms** that and adds the
  migration-time backup behavior + a WAL-safe backup recipe.

This is documentation reflecting **already-shipped behavior**, so it
emits **no new DEC**. It documents what shipped — `internal/cli/project.go`,
`internal/cli/add.go`'s cwd auto-fill, and `internal/storage/backup.go`
(SPEC-036) — not intent. DEC-017..021 are the **decisions of record**
it cites.

- Parent: `STAGE-008` (polish + v0.2.0 release), backlog item SPEC-034
  ("M — comprehensive doc sweep … L-watch: peel architecture if
  tutorial + architecture together read L").
- Project: `PROJ-002`.

### L-watch resolution (decided at design): **HOLD as one M — do not peel.**

The stage backlog put an L-watch on this spec: peel the architecture
refresh into its own S (next free `SPEC-038`) if tutorial + architecture
together read L. **Call: keep it one M.** Rationale:

- The architecture refresh is **mechanical and bounded** — add two CLI
  nodes (Tags, Project) to a mermaid diagram, extend three existing
  diagram nodes (DB tables, embedded migrations, a backup edge), and
  append project Store methods + a `backup.go` sentence to one
  responsibilities-table row. Every fact is already enumerated in
  `data-model.md` (tables) and `project.go` (methods); nothing must be
  synthesised.
- `api-contract.md` is a verification pass with **two one-line fixes**
  (a stale `STAGE-007 later spec` forward-ref and a missing `DEC-019`
  reference) — near-zero writing.
- `data-model.md` is a **confirm** (projects already documented) plus a
  backup note.
- The only substantial *writing* is the tutorial Projects walkthrough +
  the §5 backup rewrite — comfortably an M's worth on its own; the three
  other docs are transcription/verification.

The backlog said "prefer keeping it one M if the architecture refresh is
mechanical." It is. No `SPEC-038` stub is created.

## Goal

Bring all four `docs/` files into exact agreement with the shipped
v0.2.0 binary: add a projects walkthrough and a WAL-safe + auto-backup
section to the tutorial; refresh the architecture diagram and
responsibilities table for projects, tags, the `0004` migration, and
`backup.go`; and run a final consistency/version pass over
`api-contract.md` and `data-model.md`. No code, no migration, no new DEC.

## Inputs

- **Files to read (shipped behavior to mirror — document what shipped):**
  - `internal/cli/project.go` — the eight `brag project` subcommands
    (`new` / `list` / `show` / `status` / `here` / `edit` / `archive` /
    `delete`) and their exact stderr/stdout/exit behavior + `--add-path`
    / `--remove-path`.
  - `internal/cli/add.go` (`autoFillProject`, lines 21–43; the Long
    string, 63–89) — the cwd `--project` auto-fill in flag / editor /
    JSON modes; explicit `-p` always wins; silent + best-effort.
  - `internal/storage/backup.go` — the migration auto-backup safety belt
    (VACUUM INTO snapshot, `applied>0 && pending>0` trigger, sidecar name
    `<dbpath>.pre-<highestPending>.<UTC>.backup`, failure-abort).
  - `internal/storage/project.go` — Store method names + the `Project` /
    `ProjectStatus` types (for the architecture responsibilities row).
- **Files to read (current docs, the sweep targets):**
  - `docs/tutorial.md` — §4 (the command tour, where `### Tag taxonomy`
    lives at L401), §5 "Where the data lives" (L435, bare-`cp` backup at
    L445–448), §9 "Power-user escape hatch" (L519), §10 "Shell
    completions" (L528).
  - `docs/architecture.md` — the mermaid diagram (L19–58), the
    responsibilities table (L69–77), References (L140–151).
  - `docs/api-contract.md` — the project command entries (L468–629),
    intro (L5), the stale `STAGE-007 later spec` note (L476), tags note
    (L419), References (L675–689).
  - `docs/data-model.md` — projects/project_locations entities
    (L60–87), Schema Evolution (L129–142), Data Lifecycle Backup (L168–182).
- **Decisions of record:** `DEC-017`..`DEC-021` (see Implementation Context).
- **Related code paths:** none modified — `docs/` only.

## Outputs

- **Files created:** none.
- **Files modified:**
  - `docs/tutorial.md` — (a) **new `### Projects` subsection** at the end
    of §4 (after `### Tag taxonomy`, before the `---` preceding §5);
    (b) **§5 "Where the data lives" rewritten** to a WAL-safe manual
    backup recipe + an "automatic backup before an upgrade" subsection;
    (c) two stale status/version flips (§9 `v0.1.0`→`v0.2.0`; Scope
    banner forward-ref). Full literals in Notes for the Implementer.
  - `docs/architecture.md` — mermaid diagram (Tags + Project CLI nodes;
    `0004_add_projects.sql` in the embed list; `tags`/`taggings`/`projects`/
    `project_locations` on the DB node; a `backup.go` pre-migration
    snapshot edge), the caption recount, the `internal/storage`
    responsibilities row (project methods + `backup.go`), and the
    References/Decisions list (add DEC-017..021).
  - `docs/api-contract.md` — intro scope line (L5); strike the
    `STAGE-007 later spec` forward-ref (L476); tense-flip the tags
    `fold in … in STAGE-007` note (L419); add `DEC-019` to References.
  - `docs/data-model.md` — confirm projects/project_locations (no change
    needed there); add a migration-time auto-backup note to Schema
    Evolution; upgrade the Data Lifecycle "Backup" bullet to the WAL-safe
    recipe; add `DEC-021` to References.
- **New exports / Database changes:** none.

### Premise audit (run at design per §9 — STATUS-CHANGE-heavy spec)

The greps below were **run against the repo at design** and reconciled.
The dominant case is status-change (projects/tags moving from "per-spec
mentions" to "fully documented"); the inversion/count-bump cases are
confirmed **empty** (docs-only).

**Inversion / count-bump → NONE (confirmed).** This spec adds no code,
no migration, no test, and no tracked-collection member. `schema_migrations`
stays at 4. No literal-count assertion is touched.

**Doc-content tests DO exist — they constrain placement (key finding).**
The task premise ("confirm no test asserts doc content") is **false**:
`scripts/test-docs.sh` asserts on the swept docs. Every one of these
must stay green; they are the real acceptance boundary for the edits:

| Assertion | What it requires | Impact on this spec |
|---|---|---|
| `P5` | `architecture.md` does NOT contain `sqlite-file-copy` | refresh must not introduce that string |
| `P6` | `architecture.md` does NOT contain `Distribution (STAGE-004)` | unaffected (that text isn't there now) |
| `P7` | `tutorial.md` **§9 body** does NOT contain `brew install` | **§9 must keep number 9**; new content must not add `brew install` to §9 |
| `R1` | `tutorial.md` has `## 10. Shell completions` heading | **§10 must keep number 10 and that exact title** |
| `R2`/`R3`/`R4` | tutorial contains the three `brag completion … ` source lines | untouched |

**Consequence — placement is a locked design decision, not a free
choice.** Any NEW top-level `##` section inserted *before* §10 would
renumber §9 and §10 and break `P7` + `R1`, forcing edits to
`test-docs.sh`. To keep SPEC-034 **purely docs-only** (no test-script
churn), the projects walkthrough is added as a **`###` subsection of §4**
(parallel to the existing `### Tag taxonomy`), and §5's rewrite stays
in place — **no top-level renumbering**. Rejected alternative: a new
top-level `## 5. Working with projects` + renumber + update `test-docs.sh`
`P7`/`R1` anchors — rejected because it converts a docs-only M into a
spec that also edits a test harness, for no reader benefit over a §4
subsection (§4 is already the de-facto command tour: it holds tags,
search, edit, delete, review, stats).

**Status-change grep — `grep -rn -i "project" docs/` (swept docs only),
reconciled:**

- `tutorial.md` — **no** projects content today → the new `### Projects`
  subsection + the auto-fill note are pure **additions** (no stale text
  to strike). The §4 `--project` filter/`-P` mentions (L184–208) are
  about `entries.project` free text and **stay** (DEC-017 soft match).
- `tutorial.md:3–4` Scope banner: "for the full plan" links the
  **PROJ-001-mvp brief** — stale. → **UPDATE** to point at
  `docs/api-contract.md` (the live full surface).
- `tutorial.md:521` (§9): "Everything in this tutorial is shipped in
  **v0.1.0**." → **UPDATE** to `v0.2.0` (tutorial now covers tags,
  projects, completion). Stays in §9, no `brew install` added → `P7` safe.
- `tutorial.md:450` (§5): "useful until `show` exists" — `show` shipped
  in STAGE-002. → **UPDATE** (drop the stale clause) as part of the §5
  rewrite.
- `architecture.md` — diagram + DB node omit projects (and tags). →
  **UPDATE** (the refresh).
- `api-contract.md:476`: "(use `brag project edit` to change them —
  **STAGE-007 later spec**)" — that later spec (SPEC-033) shipped. →
  **UPDATE**: strike the forward-ref.
- `api-contract.md:419` (tags): "`'project'` rows fold in automatically
  **in STAGE-007** with no change here" — STAGE-007 shipped. →
  **UPDATE** the tense.
- `api-contract.md:5`: "across **PROJ-001**" — the contract now spans
  PROJ-002 too. → **UPDATE** to "across PROJ-001 and PROJ-002."
- `api-contract.md:172`: "primary mechanism for updating entries in
  PROJ-001" — historically accurate, not a forward-ref. → **STAYS.**
- `data-model.md:60–87,163–166,189`: projects/project_locations already
  documented. → **CONFIRM (no change).**

**Bare-`cp` backup grep — `grep -rn -iE "cp .*sqlite|copy the file|copying the file" docs/ README.md`:**

- `tutorial.md:445–448` (§5) — bare-`cp` recipe. → **UPGRADE** to
  `sqlite3 .backup` / `VACUUM INTO` (the §5 rewrite).
- `data-model.md:182` ("Backup. Copy the file.") and `:178` ("the
  SQLite file is the backup"). → **UPGRADE** L182 to the WAL-safe
  recipe + safety-belt pointer; L178 (a delete-lifecycle aside) softened.
- `README.md:120,145` — bare-`cp` / "Back up by copying the file." →
  **OUT OF SCOPE.** README is not one of the four swept docs (scope:
  tutorial / architecture / api-contract / data-model). Enumerated here
  as a follow-up flag for the project-close docs pass or a tiny README
  chore; **not** fixed in this spec. (AC "no doc contains a bare-cp
  recipe" is scoped to the swept docs.)

**Test-asserting-docs grep — `grep -rln "docs/" **/*_test.go scripts/`:**

- **No Go test** reads or asserts on any swept doc (`grep -rn
  "tutorial|api-contract|architecture|data-model" **/*_test.go` → none).
- `scripts/test-docs.sh` is the only doc-content test; its swept-doc
  assertions are the P/R table above — all preserved by the chosen
  placement.
- `scripts/claude-code-post-session.sh` references `docs/brag-entry.schema.json`
  only — not swept here.

## Acceptance Criteria

Docs specs have no Go failing tests; these are concrete, cold-gradeable
checks against the shipped binary + the doc text. Each gives the exact
grep a verifier can run from repo root.

**Tutorial — projects walkthrough**

- [ ] **AC1** §4 contains a `### Projects` subsection.
  `grep -nF '### Projects' docs/tutorial.md` → exactly one hit, located
  after `### Tag taxonomy` and before `## 5.`.
- [ ] **AC2** The walkthrough names all eight subcommands. Each of
  `project new`, `project here`, `project status`, `project list`,
  `project show`, `project edit`, `project archive`, `project delete`
  appears in the subsection.
  `for c in "project new" "project here" "project status" "project list" "project show" "project edit" "project archive" "project delete"; do grep -qF "brag $c" docs/tutorial.md || echo "MISSING: $c"; done` → no output.
- [ ] **AC3** The walkthrough documents the cwd `--project` auto-fill and
  that an explicit `-p`/`--project` overrides it.
  `grep -niE 'auto-fill|auto-set|fills? in .*project' docs/tutorial.md` →
  ≥1 hit in the Projects subsection; the "explicit … always wins" fact is
  present.
- [ ] **AC4** `--add-path` and `--remove-path` are shown.
  `grep -cF -e '--add-path' -e '--remove-path' docs/tutorial.md` → ≥1 each.
- [ ] **AC5** archive (recoverable) vs delete (irreversible) is stated,
  and both note that brag **entries are not touched**.

**Tutorial — backup (§5)**

- [ ] **AC6** §5 documents the migration auto-backup safety belt with
  **all four** load-bearing facts from `backup.go`/DEC-021: (i) it fires
  only when an **existing** DB has **pending** migrations; (ii) mechanism
  is a `VACUUM INTO` snapshot of the **pre-migration** state; (iii) the
  sidecar pattern `…db.sqlite.pre-<version>.<UTC>.backup`; (iv) a failed
  snapshot **aborts** the open (no migration runs).
  `grep -niE 'pre-0004_add_projects|\.pre-.*\.backup' docs/tutorial.md` →
  ≥1 hit; manual read confirms (i)–(iv).
- [ ] **AC7** §5 gives a WAL-safe **manual** recipe using `sqlite3
  ".backup"` and/or `VACUUM INTO`, and contains **no** bare-`cp`
  backup recipe.
  `grep -nE "\.backup '|VACUUM INTO" docs/tutorial.md` → ≥1; and
  `grep -nE 'cp .*\.bragfile/db\.sqlite' docs/tutorial.md` → **0 hits**.
- [ ] **AC8** Section numbers are unchanged: `grep -nE '^## 9\.' docs/tutorial.md`
  → "Power-user escape hatch"; `grep -nE '^## 10\. Shell completions' docs/tutorial.md`
  → present. (Guards `P7`/`R1`.)

**Architecture**

- [ ] **AC9** The mermaid diagram shows a `Project` CLI node naming the
  project subcommands and a `Tags` node, both edged to `Store`.
  `grep -niE 'Project\[Project|Tags\[Tags' docs/architecture.md` → both.
- [ ] **AC10** The embedded-migrations diagram node lists
  `0004_add_projects.sql`; the DB node lists `projects` +
  `project_locations` (and `tags` + `taggings`).
  `grep -nF '0004_add_projects.sql' docs/architecture.md` → 1;
  `grep -niE 'projects \+ project_locations|project_locations' docs/architecture.md` → ≥1.
- [ ] **AC11** The responsibilities table's `internal/storage` row names
  the project Store methods and **`backup.go`** (the auto-backup safety
  belt) with its trigger + abort behavior.
  `grep -nF 'backup.go' docs/architecture.md` → ≥1;
  `grep -nF 'ProjectForPath' docs/architecture.md` → ≥1.
- [ ] **AC12** Architecture References list DEC-017, DEC-018, DEC-019,
  DEC-020, DEC-021.
  `for d in 017 018 019 020 021; do grep -qF "DEC-$d" docs/architecture.md || echo "MISSING DEC-$d"; done` → no output.
- [ ] **AC13** `P5` still holds: `grep -iF 'sqlite-file-copy' docs/architecture.md`
  → **0 hits**.

**API contract**

- [ ] **AC14** No stale `STAGE-007 later spec` forward-reference remains.
  `grep -nF 'STAGE-007 later spec' docs/api-contract.md` → **0 hits**.
- [ ] **AC15** All eight `brag project` subcommands + `add` auto-fill +
  the `--add-path`/`--remove-path` flags are present (they already are —
  verify, don't add): `grep -cF 'brag project' docs/api-contract.md` → ≥8;
  `grep -cF -e '--add-path' -e '--remove-path' docs/api-contract.md` → ≥1 each.
- [ ] **AC16** `DEC-019` is in the References list (it is cited in the
  body but was missing from the list).
  `grep -nF 'DEC-019' docs/api-contract.md` → ≥2 (body + References).
- [ ] **AC17** Stability guarantees + version markers are correct for
  v0.2.0: the Pre-1.0 / Post-1.0 block is intact and still says v0.x is
  plastic (v0.2.0 is v0.x — no change, confirm present).

**Data model**

- [ ] **AC18** `projects` and `project_locations` entities are documented
  (confirm): `grep -cE '^### Entity: `project' docs/data-model.md` → 2.
- [ ] **AC19** Schema Evolution notes the migration-time auto-backup
  (existing DB snapshotted before pending migrations apply; failure
  aborts; DEC-021).
- [ ] **AC20** The Data Lifecycle "Backup" bullet prefers the WAL-safe
  recipe (no bare-`cp` as *the* mechanism) and `DEC-021` is in References.

**Repo-wide**

- [ ] **AC21** `bash scripts/test-docs.sh` exits 0 (all doc-content
  assertions, including P5/P6/P7/R1–R4, pass).
- [ ] **AC22** No code/migration/test changed: `git diff --name-only`
  touches only files under `docs/` (plus this spec). `go test ./...`,
  `gofmt -l .`, `go vet ./...` are unaffected (not re-run for a docs-only
  change, but must remain green).

## Failing Tests

**This is a documentation spec — there are no Go failing tests** (no
behavior changes; the §9 `test-before-implementation` constraint applies
to code, and this spec touches none). The Acceptance Criteria above are
the gradeable substitute: each is a concrete grep or a named fact a cold
verifier checks against the shipped code. The one executable gate is
**AC21** (`bash scripts/test-docs.sh` exits 0), which already exists and
must stay green.

## Implementation Context

*Read this section (and the files it points to) before starting the
build cycle. The doc edits are transcription of already-shipped behavior;
the literal artifacts below are authoritative — transcribe them, do not
re-derive from intent.*

### Decisions that apply (docs of record — cite, don't create)

- `DEC-017` — `entries.project` ↔ `projects` **soft string match**
  (free text, opportunistic join on `projects.name`, no FK, no backfill);
  `projects.status` enum; single `state_note`. Why it matters: the
  walkthrough must say renaming/deleting a project does **not** rewrite or
  remove brag entries' project strings.
- `DEC-018` — `brag project delete` blast radius: entries untouched,
  `project_locations` removed in-tx, archive is the recoverable flip.
- `DEC-019` — `brag project here` / add auto-fill **nearest-ancestor
  (longest-prefix)** resolution. Why: the walkthrough's "you may be in
  any subdirectory" claim is DEC-019.
- `DEC-020` — `brag project edit` location editing: verbatim path match;
  `--add-path`/`--remove-path` atomic (removes before adds); location
  edits don't bump `updated_at`.
- `DEC-021` — migration auto-backup durability model: trigger
  `applied>0 && pending>0`; `VACUUM INTO`; failure ABORTS Open; sidecar
  `<dbpath>.pre-<ver>.<UTC>.backup`; keep-all; silent. The §5
  auto-backup subsection mirrors this exactly.

### Constraints that apply

Docs-only — no blocking code constraints fire. The governing discipline
is the **§9 premise-audit family** (status-change case) and the
**§12 literal-artifact-as-spec** rule (the doc bodies are fixed-shape
artifacts embedded below; build transcribes, verify diffs). The one hard
gate is keeping `scripts/test-docs.sh` green (the P/R assertions above).

### Prior related work

- `SPEC-027`..`SPEC-033` (shipped) — the projects feature this documents.
- `SPEC-036` (shipped, PR #48) — the migration auto-backup safety belt;
  `internal/storage/backup.go` is the source of truth for the §5
  auto-backup subsection.
- `SPEC-025` (shipped) — tags normalization; the architecture diagram
  refresh also closes the leftover STAGE-006 gap (tags CLI node +
  `tags`/`taggings` on the DB node).

### Out of scope (for this spec specifically)

- **`README.md`** bare-`cp` backup mention (L120,145) — README is not a
  swept doc; flag for project-close or a tiny chore.
- **CHANGELOG** (`SPEC-035`), the **release cut** (`SPEC-037`), **blog
  posts**, any **code/migration/test** change.
- Any **new behavior** — if the docs and the binary disagree, the
  **binary wins** and the discrepancy is raised as a question, not
  "fixed" by changing code.

---

## Notes for the Implementer

Five literal edits. Match the tutorial's voice: numbered/`###` sections,
fenced shell blocks with `#`-comments showing representative output.

### EDIT 1 — `docs/tutorial.md`: new `### Projects` subsection

Insert **immediately after** the tag-merge bullets that end `### Tag
taxonomy` (current L431, the `…count rises…` bullet) and **before** the
`---` at L433. Verbatim:

```markdown
### Projects

A **project** is a first-class, named workspace you register once and
then attach to brags automatically. Registering a project's directory
lets `brag add` fill in `--project` for you whenever you work inside it.

Register a project and point it at a directory:

```bash
brag project new bragfile --path ~/code/bragfile
# stderr: Created project "bragfile".
```

`--path` is required and stored verbatim; a path already registered to
another project is rejected.

Ask which project the current directory belongs to:

```bash
cd ~/code/bragfile/internal/storage
brag project here
# bragfile	active	-
# name<TAB>status<TAB>state-note. Nearest-ancestor match: any
# subdirectory of a registered path resolves, not just the exact root.
```

Outside any registered project, `brag project here` prints `not inside
any registered project` to stderr and exits 1.

#### Auto-fill `--project` from your working directory

Once a directory is registered, `brag add` fills in `--project` for you
whenever you don't pass one — in flag, editor, and `--json` modes alike:

```bash
cd ~/code/bragfile
brag add -t "shipped the projects walkthrough"
# the entry's project is auto-set to "bragfile" — no -p needed

brag add -t "a cross-cutting note" -p platform
# an explicit -p always wins; auto-fill never overrides it
```

Auto-fill is silent and best-effort: outside any registered project, or
if the directory can't be resolved, the entry is just saved with no
project, exactly as before.

#### Review your projects

```bash
brag project status
# name<TAB>status<TAB>brag-count<TAB>state-note, most-recent first.
# Archived projects are hidden. brag-count is how many entries carry
# that project name (the DEC-017 soft string match).

brag project list
# name<TAB>status<TAB>locations (comma-joined; "-" when none)

brag project show bragfile          # labeled block: Name / Status / State note / Locations
brag project show bragfile --format json
```

`status`, `list`, `show`, and `here` all accept `--format json` for
scripting.

#### Edit a project

```bash
brag project edit bragfile --status paused
brag project edit bragfile --state-note "shipped tags; next: cut v0.2.0"
brag project edit bragfile --name brag-cli
brag project edit bragfile --add-path ~/code/bragfile-fork
brag project edit bragfile --remove-path /srv/old-location
```

- `--status` is one of `active`, `paused`, `done`, `archived`.
- `--add-path` / `--remove-path` are repeatable and apply atomically
  (removes before adds); paths match verbatim against what was
  registered.
- Renaming a project does **not** rewrite the project string on existing
  brags — they keep what they were captured with.

#### Archive vs. delete

```bash
brag project archive bragfile
# status → "archived"; recoverable. Restore it with:
brag project edit bragfile --status active

brag project delete bragfile        # prompts y/N on stdin
brag project delete bragfile --yes  # skip the prompt
```

`archive` is a reversible status flip that hides the project from `brag
project status` but preserves everything. `delete` is **irreversible** —
it removes the project and its registered locations. **Neither touches
your brag entries:** an entry keeps its project string, so `brag list
--project bragfile` still finds those entries afterward.
```

### EDIT 2 — `docs/tutorial.md`: rewrite §5 "Where the data lives"

Replace the current §5 body (L435–454, from the `## 5.` heading through
the closing `sqlite3 … limit 3` block) with:

```markdown
## 5. Where the data lives

```bash
ls -la ~/.bragfile/db.sqlite
```

That's the default, and **every `brag` invocation from any directory
uses it** — the path is absolute and home-expanded, so it doesn't matter
whether you're in the bragfile repo or elsewhere.

### Back up your brags

The database is a single SQLite file, so a backup is a copy of that file
— but take the copy with SQLite's own backup command, not a bare `cp`,
so the snapshot is always transaction-consistent:

```bash
# preferred — a consistent single-file snapshot via the sqlite3 CLI:
sqlite3 ~/.bragfile/db.sqlite ".backup '$HOME/brag-backup.db'"

# equivalent, also consistent:
sqlite3 ~/.bragfile/db.sqlite "VACUUM INTO '$HOME/brag-backup.db'"
```

The result is a portable `.db` you can copy to another machine, commit
to a private repo, or stash anywhere:

```bash
cp ~/brag-backup.db ~/some-private-repo/   # the snapshot is safe to plain-copy
```

> Why not `cp ~/.bragfile/db.sqlite` directly? `brag` doesn't enable WAL
> mode, so a bare `cp` of an idle database is *currently* safe — but
> `.backup` / `VACUUM INTO` stay correct even if that changes or a write
> is in flight. Prefer them. (You can also `brag export --format json
> --out backup.json` for a tool-portable dump — see §4.)

### Automatic backup before an upgrade migrates your DB

When a newer `brag` opens your existing database and needs to apply a
schema migration, it **snapshots the database first** — automatically,
before touching it. The snapshot lands next to your DB as a timestamped
sidecar:

```
~/.bragfile/db.sqlite.pre-0004_add_projects.20260612T093015Z.backup
```

- It fires **only** when an existing database has pending migrations to
  apply — a brand-new DB and an already-up-to-date DB are never copied.
- The copy is a consistent `VACUUM INTO` snapshot of the
  **pre-migration** state; open it with `sqlite3` to inspect or recover.
- If the snapshot can't be written, `brag` **aborts** rather than migrate
  an un-backed-up database (exit 2) — nothing is changed.
- It's silent and non-interactive, so it never breaks `brag add --json`
  or other scripted pipelines.
- Snapshots are **kept, not pruned** — delete old `*.backup` sidecars
  yourself when you no longer need them.

Peek at raw data:

```bash
sqlite3 ~/.bragfile/db.sqlite "select * from entries order by id desc limit 3"
```
```

> Note (heading levels): EDIT 2 introduces `###` subsections **under**
> §5. That does not change any `## N.` number, so `P7` (§9) / `R1` (§10)
> are unaffected.

### EDIT 3 — `docs/tutorial.md`: two status/version flips

1. **Scope banner** (L3–4). Replace:
   ```
   > **Scope:** what you can do with `brag` today. See
   > [`projects/PROJ-001-mvp/brief.md`](../projects/PROJ-001-mvp/brief.md)
   > for the full plan.
   ```
   with:
   ```
   > **Scope:** what you can do with `brag` today. See
   > [`docs/api-contract.md`](./api-contract.md) for the full command
   > surface.
   ```
2. **§9** (L521). Replace `Everything in this tutorial is shipped in
   v0.1.0.` → `Everything in this tutorial is shipped in v0.2.0.`
   (Stays inside §9; adds no `brew install` → `P7` safe.)

### EDIT 4 — `docs/architecture.md`: diagram + table + references refresh

**4a. Mermaid diagram.** After the `Main --> Completion[...]` line (L31),
add two CLI-group nodes and their Store edges:

```
    Main --> Tags[Tags<br/>tags / tag rename / tag merge]
    Main --> Project[Project<br/>project new / list / show / status<br/>/ here / edit / archive / delete]
    Tags --> Store
    Project --> Store
```

Update the **DB node** (L46) to list the STAGE-006/007 tables:

```
    Driver --> DB[(~/.bragfile/db.sqlite<br/>mode 0600<br/>entries + entries_fts + schema_migrations<br/>+ tags + taggings + projects + project_locations)]
```

Update the **embedded-migrations node** (L47) to add `0004`:

```
    Store -.embeds.-> Migrations[migrations/0001_initial.sql<br/>migrations/0002_add_fts.sql<br/>migrations/0003_normalize_tags.sql<br/>migrations/0004_add_projects.sql<br/>via embed.FS]
```

Add a **backup edge + node** (after the `Migrations -.applied on Open…`
line, L48):

```
    Store -.snapshots before migrating.-> Backup[internal/storage/backup.go<br/>VACUUM INTO pre-migration sidecar]
    Backup -.VACUUM INTO.-> DB
```

Extend the **classDef assignments** (L54, L56) so the new nodes are
styled — append `,Tags,Project` to the `cli` class line and `,Backup` to
the `storage` class line:

```
    class Main,Capture,Retrieve,Digest,ExportCmd,DeleteCmd,Completion,Tags,Project cli
    class Store,Driver,Migrations,Backup storage
```

**Do NOT introduce the literal string `sqlite-file-copy`** anywhere
(guards `P5`).

**4b. Caption** (L60–65). The "twelve subcommands … six functional
groups" sentence is now wrong. Replace that clause with a
non-count-brittle phrasing:

```
The command surface is clustered into functional groups — capture,
retrieve, digest, export, delete, tags, projects, completion — rather
than listed individually; see the per-command contracts in
[./api-contract.md](./api-contract.md).
```

**4c. Responsibilities table** — replace the `internal/storage` row
(L74) with (the project methods + `Project` type + `backup.go` are the
additions; keep the existing tags/FTS prose):

```
| `internal/storage` | `Store` struct wrapping `*sql.DB`. Embeds migration SQL files and applies them on `Open`. Exposes typed methods for entries (`Add`, `List`, `Get`, `Update`, `Delete`, `Search`), tags (`TagCounts`, `RenameTag`, `MergeTags`), and projects (`CreateProject`, `GetProject`, `GetProjectByName`, `ListProjects`, `ProjectStatuses`, `AddLocation`, `RemoveLocation`, `EditLocations`, `UpdateProject`, `ArchiveProject`, `DeleteProject`, `ProjectForPath`) — no SQL leaks upward. Owns the `Entry`, `TagCount`, `Project`, and `ProjectStatus` types. Projects are a first-class entity (`projects` + `project_locations` tables, DEC-017 / SPEC-027); `entries.project` joins `projects.name` by soft string match. Tags are stored in a normalized `tags`/`taggings` join (DEC-015 / SPEC-025); `Entry.Tags` is a reconstructed comma-joined projection. `entries_fts` (regular own-content FTS5) indexes title, description, tags projection, project, impact and stays in sync via 6 SQL triggers. On `Open`, `backup.go` snapshots an existing DB via `VACUUM INTO` to a timestamped sidecar **before** any pending migration runs and **aborts** the open if that snapshot fails — never migrating an un-backed-up DB (DEC-021). |
```

**4d. References** (L144–151). Append the projects + backup DECs:

```
  - `DEC-017` — `entries.project` ↔ `projects` soft string match; project status enum + state_note
  - `DEC-018` — `brag project delete` blast radius (entries untouched; archive vs delete)
  - `DEC-019` — `brag project here` / add auto-fill nearest-ancestor resolution
  - `DEC-020` — `brag project edit` location-editing semantics (atomic, verbatim, no updated_at bump)
  - `DEC-021` — migration auto-backup durability model (VACUUM INTO snapshot before migrate; failure aborts)
```

Optionally (nice-to-have, not graded): extend Key Design Principle 3
("Migrations are code") with a closing sentence: *"And migrations
snapshot before they mutate: an existing DB is copied to a timestamped
sidecar before any pending migration runs (DEC-021)."*

### EDIT 5 — `docs/api-contract.md` + `docs/data-model.md`: consistency pass

**api-contract.md:**

1. **L5** — `… frozen spec of that contract across PROJ-001: …` →
   `… frozen spec of that contract across PROJ-001 and PROJ-002: …`.
2. **L476** — strike the stale forward-ref. Current:
   `(use 'brag project edit' to change them — STAGE-007 later spec).`
   →  `(use 'brag project edit' to change them).`
3. **L419** (tags note) — `… 'project' rows fold in automatically in
   STAGE-007 with no change here).` → `… 'project' rows folded in
   automatically with STAGE-007 (shipped), with no change here).`
4. **References** (L675–689) — add the missing `DEC-019` line (it is
   cited in the body at L117/L552 but absent from the list):
   ```
   - `DEC-019` — `brag project here` / `brag add` cwd auto-fill: nearest-ancestor (longest-prefix) match against registered project locations
   ```
   (Verify `DEC-021` is **not** required here — the safety belt is not a
   CLI-contract surface; it belongs in architecture/data-model. Leave the
   contract's command entries otherwise unchanged: they already document
   all eight project subcommands + auto-fill + `--add-path`/`--remove-path`
   — this pass only verifies, per AC15.)
5. **Stability guarantees** (L667–673) — read and confirm intact; v0.2.0
   is still v0.x, so "flag names may change between v0.x releases" stands.
   No edit unless a fact is wrong.

**data-model.md:**

1. **Confirm** `### Entity: projects` (L60) and `### Entity:
   project_locations` (L75) are present and accurate — they are; **no
   change** to the entity tables (AC18 is a confirm).
2. **Schema Evolution** (after L142, the "Backward compatibility" bullet)
   — add a bullet:
   ```
   - **Snapshot before migrate.** On `storage.Open`, if an existing
     database has pending migrations, it is copied to a timestamped
     sidecar (`<db>.pre-<version>.<UTC>.backup`) via `VACUUM INTO`
     **before** the migration transaction runs; a failed snapshot aborts
     `Open` rather than migrate an un-backed-up DB (DEC-021, SPEC-036).
   ```
3. **Data Lifecycle → Backup** (L182). Replace `- **Backup.** Copy the
   file. That is the supported mechanism.` with:
   ```
   - **Backup.** Take a consistent snapshot with `sqlite3 ~/.bragfile/db.sqlite ".backup '…'"` or `VACUUM INTO` (preferred over a bare `cp`); a migration-time auto-backup also snapshots before any schema upgrade (DEC-021). See `docs/tutorial.md` §5.
   ```
   Soften L178 (`… the SQLite file is the backup.`) → `… the SQLite file
   is the source of truth; see the backup recipe below.`
4. **References** (L201–210) — add:
   ```
   - `DEC-021` — migration auto-backup durability model (VACUUM INTO snapshot before applying a pending migration; failure aborts Open)
   ```

### Build checklist

- Make EDITs 1–5 exactly as the literals read. Do not paraphrase the
  backup facts (they mirror `backup.go`/DEC-021 precisely).
- After editing: `bash scripts/test-docs.sh` → must print `ALL OK` and
  exit 0 (AC21 — guards P5/P6/P7/R1–R4).
- `git diff --name-only` must show **only** files under `docs/` (AC22).
- Spot-check the AC greps in this spec.
- Do **not** touch `README.md`, `scripts/test-docs.sh`, any `.go` file,
  any migration, or the uncommitted
  `docs/framework-feedback/process-feedback.md` (a parallel session owns
  it — §13).

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - none expected (documentation reflecting shipped behavior)
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - README.md bare-`cp` backup mention (L120,145) — out of scope here;
    candidate tiny chore or project-close docs pass.

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — <answer>

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — <answer>

3. **If you did this task again, what would you do differently?**
   — <answer>

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

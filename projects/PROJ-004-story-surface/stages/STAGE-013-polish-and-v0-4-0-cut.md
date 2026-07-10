---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-013
  status: shipped
  priority: high                    # closes PROJ-004 and cuts v0.4.0
  target_complete: null             # optional: YYYY-MM-DD

project:
  id: PROJ-004                      # parent project
repo:
  id: bragfile

created_at: 2026-07-06
shipped_at: 2026-07-07             # v0.4.0 cut + published to Homebrew
---

# STAGE-013: Polish + the v0.4.0 cut

## What This Stage Is

The closing stage of PROJ-004. It adds the shareable/visual polish on top of the
read/story surface (`brag impact` + `brag story`, both already on main) and then
**cuts v0.4.0** — the release that finally puts the whole surface into users'
installed CLIs. Per the scoping decision (2026-07-06): **bundle the polish, then
cut one richer v0.4.0** (rather than cutting the core now and following with the
polish).

## Why Now

- The core v0.4.0 value (`impact` + `story`) is built + dogfooded on main; the
  polish makes the release feel complete and shareable before it ships.
- These are the user-requested "fun"/shareable surfaces — a wrapped-style
  review and terminal visuals — that make the story surface land.

## Success Criteria

- **`brag wrapped [year|quarter]`** — a shareable year- or quarter-in-review
  digest over the DEC-014 envelope (quarterly is first-class: many companies
  report by the quarter).
- **A visual/sparklines pass** — in-terminal Unicode sparklines (`▁▂▃▄▅▆▇█`,
  local-first, no dependency, no network) surfacing cadence in `brag stats` /
  `brag wrapped` (and optionally `impact`). An optional external-plotter pipe
  may layer on later.
- **`--previous`** — the last-completed-period window modifier for the
  calendar-windowed commands (`impact`/`story`/`wrapped`), the clean addition
  DEC-028 foresaw.
- **The P3 agent-assist metric** — adopt the drafted **SPEC-045** (a personal
  "how much of my work was agent-assisted" / provenance-share read over the
  `agent:`/`model:` tags).
- **v0.4.0 is cut** — RC → dual-tag Pattern 1 → final, brew-upgrade verify, the
  §12(b) behavioral check, via the `spec-release-cut` template + AGENTS.md §4.
  Ships `impact` + `story` + this stage's polish.

## Scope

### In scope
- `brag wrapped [year|quarter]`; the sparklines/visual pass; `--previous`; the
  P3 agent-assist metric (SPEC-045); the v0.4.0 release cut.

### Explicitly out of scope (→ later)
- Team / multi-user federation, token economics reconciliation (PROJ-005).
- Any CLI→LLM call-out (story stays a pipe).
- An external-plotter integration beyond an optional pipe (defer unless earned).

## Spec Backlog

Ordered; IDs finalize at design emission (§2). Release cut is LAST (it tags what
the feature specs land on main).

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [x] SPEC-051 (shipped on 2026-07-06) — **`brag wrapped [year|quarter]`** — the
      shareable year/quarter-in-review digest (text-first; DEC-014 envelope +
      DEC-030). The headline polish feature. Both byte-goldens verified faithful
      against the real aggregate output; `cadence.series` left sparkline-ready
      for SPEC-052.
- [x] SPEC-052 (shipped on 2026-07-07) — **the sparklines/visual pass** — an
      in-terminal Unicode sparkline primitive (`internal/spark`, local-first, no
      dep) rendered into `wrapped`'s `## Cadence` markdown as a default-on
      `Cadence: <glyphs>` line (escaped by `--no-spark`/`NO_COLOR`); JSON stays
      raw. DEC-031 (min→max normalization, placement, default-on-with-escape);
      `stats`/`impact` deferred. Normalization goldens faithful; JSON byte-unchanged.
- [x] SPEC-053 (shipped on 2026-07-07) — **`--previous`** — the last-completed-
      period window modifier for `impact`/`story`/`wrapped` (DEC-032, extends
      DEC-028; shared `windowCutoff`, bounded, no regression). Surfaced a
      follow-up: the `created_at < end` filter is now a 3rd consumer → a
      `ListFilter.Until` storage promotion (backlog, not blocking v0.4.0).
- [x] SPEC-045 (shipped on 2026-07-07) — **`brag coverage` — the P3 agent-assist
      metric** — the provenance-share read ("how much of my work was
      agent-assisted"), re-homed here from PROJ-003/STAGE-010 as a standalone
      sixth DEC-014 digest: monthly agent-vs-human share + an agent-share
      sparkline + self-reference density, with `IsAgentAuthored` single-sourcing
      the SPEC-043 classifier via a cross-package agreement test (DEC-033).
- [x] SPEC-054 (shipped on 2026-07-07) — **the v0.4.0 release cut** — the stage's
      closing action. Mechanical prep merged to main (CHANGELOG `[0.4.0]` + plugin
      pin `0.4.0` + pre-flight ticked); the irreversible RC → Pattern 1 → final tag,
      brew verify, and §12(b) behavioral check are the orchestrator's at the cut.

**Count:** 5 shipped / 0 active / 0 pending

## Design Notes

- **Sparklines are local-first Unicode** (`▁▂▃▄▅▆▇█`) — no dependency, no
  network (DEC-001). An external-plotter pipe is an optional later layer, not
  the core. Deferred here from `brag impact` (which shipped text-pure) on
  purpose — this is that "dedicated visual pass."
- **`brag wrapped` reuses** `internal/aggregate` + the DEC-014 envelope like the
  other digests; quarterly is first-class (companies report by quarter).
- **The release cut** follows the SPEC-047 (v0.3.1) runbook precedent + the
  operational pre-flight checklist (`spec-release-cut` template).

## Dependencies

### Depends on
- **STAGE-011 (`brag impact`) + STAGE-012 (`brag story`)** — the surfaces this
  polishes and releases.
- **DEC-014** (envelope), **DEC-028** (calendar windows — `--previous` extends
  it), the `agent:`/`model:` provenance corpus (for the P3 metric).

### Enables
- **The v0.4.0 release** — the whole read/story surface reaches users.
- **PROJ-005+** — begins once v0.4.0 ships (team federation + economics).

## Stage-Level Reflection

*Filled at ship (2026-07-07, v0.4.0 published to Homebrew).*

- **Did we deliver the outcome in "What This Stage Is"?** Yes. The polish landed
  and v0.4.0 shipped: `brag wrapped` (year+quarter), the in-terminal sparkline
  pass, `--previous`, and `brag coverage` — then the release cut published the
  whole read/story surface (impact + story + these) to Homebrew. `brew upgrade`
  0.3.1→0.4.0 verified; prod DB opens migration-free.
- **How many specs did it actually take?** Five: SPEC-051 (wrapped), SPEC-052
  (sparklines), SPEC-053 (--previous), SPEC-045 (coverage, adopted from a
  PROJ-003 draft), SPEC-054 (the v0.4.0 cut) + DEC-030/031/032/033.
- **What changed between starting and shipping?** Per the 2026-07-06 scoping
  call we bundled the polish rather than cutting a core-only v0.4.0 — one richer
  release. The digest family's shared primitives (aggregate toolbox, DEC-014
  envelope, calendar-window infra, `spark`) compounded: coverage was assembled
  almost entirely from prior machinery.
- **Lessons that should update AGENTS.md, templates, or constraints?** Two
  observations tracked (both below the §12 bar): window semantics are per-command
  (rolling/calendar-to-now/bounded — DEC-014/028/030/032), and the Go-predicate↔
  SQL-clause agreement-test pattern (DEC-033) is the reusable move for a
  classifier split across `no-sql-in-cli-layer`. One code-quality follow-up
  spawned: promote the duplicated upper-bound filter to `ListFilter.Until`.
- **Should any spec-level reflections be promoted to stage-level lessons?** The
  two above are the stage-level lessons; the "faithful goldens at design time"
  discipline (post-SPEC-049) held across all five specs and is the process win.

---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-013
  status: active
  priority: high                    # closes PROJ-004 and cuts v0.4.0
  target_complete: null             # optional: YYYY-MM-DD

project:
  id: PROJ-004                      # parent project
repo:
  id: bragfile

created_at: 2026-07-06
shipped_at: null
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
- [ ] SPEC-053 (planned) — **`--previous`** — the last-completed-period window
      modifier for `impact`/`story`/`wrapped` (DEC-028 foresaw it).
- [ ] SPEC-045 (adopt draft) — **the P3 agent-assist metric** — complete the
      drafted provenance-share read ("how much of my work was agent-assisted").
- [ ] SPEC-054 (planned) — **the v0.4.0 release cut** — the stage's closing
      action (RC → Pattern 1 → final; CHANGELOG `[0.4.0]`; brew verify).

**Count:** 2 shipped / 0 active / 3 pending

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

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Stage Is"?** <tbd>
- **How many specs did it actually take?** <tbd>
- **What changed between starting and shipping?** <tbd>
- **Lessons that should update AGENTS.md, templates, or constraints?** <tbd>
- **Should any spec-level reflections be promoted to stage-level lessons?** <tbd>

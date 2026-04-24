# PROJ-001-mvp — Session Log

Append-only log of work sessions on PROJ-001. One entry per session,
reverse-chronological (newest first). Each entry captures what shipped,
lessons earned, and a pick-up-here note for the next session so context
survives across fresh Claude sessions and multi-day gaps.

This file complements (does not replace):
- `brag` entries — career-worthy moments; one per ship, not one per session.
- `/decisions/DEC-*.md` — persistent architectural commitments.
- `/AGENTS.md` §9 — process lessons earned from reflection.
- Stage + spec files — artifact-level state.
- Weekly review (Prompt 6 in `FIRST_SESSION_PROMPTS.md`) — not yet ported
  to run against this project; when it is, weekly reviews will read this
  log as their primary input.

**Entry shape** (copy for new entries):

```
## YYYY-MM-DD — <one-line session headline>

**Session duration (approx):** <N hours>
**Branches / PRs landed:** <ids>
**brag entries:** <#ids>

### What shipped
- <bullet per concrete deliverable>

### Framework / AGENTS.md lessons earned
- <bullets; empty if none>

### Stage state at session end
<compact status: X shipped / Y active / Z pending / W deferred>

### Pick up here next session
<2–5 sentences naming the next concrete action, any context that would
disappear otherwise, and any soft decisions that need revisiting>
```

---

## 2026-04-24 — STAGE-003 at 3 of 4, only SPEC-017 left

**Session duration (approx):** ~6 hours (resumed from 2026-04-23
summarized session; real elapsed wall time included multi-session
work earlier in the day).
**Branches / PRs landed:** PR #14 (SPEC-014), PR #15 (SPEC-015).
**brag entries:** #16 (SPEC-014 ship), #17 (SPEC-015 ship).

### What shipped

- **SPEC-014** — JSON trio + DEC-011. `brag list --format json|tsv`,
  `brag export --format json`, shared `internal/export.ToJSON`. DEC-011
  locked at confidence 0.85: naked array, SQL-column field order,
  comma-joined tags, RFC3339 timestamps, empty-string-not-omit, indent=2.
  14 tests including the load-bearing byte-identical cross-path
  assertion.
- **SPEC-015** — markdown export + DEC-013. `brag export --format
  markdown` with provenance + summary + grouped-by-project default,
  `--flat` escape. `renderEntry` lifted from `internal/cli/show.go` to
  `internal/export/markdown.go` with heading-level parameterization
  (level 1 for `brag show`, level 3 for markdown export). Two byte-exact
  goldens (grouped + flat) locked DEC-013.
- **SPEC-016** deferred to `backlog.md` (scope tightening — `cp
  ~/.bragfile/db.sqlite backup.db` already covers the portable-backup
  use case).
- **Session log pattern seeded** (this file). Noted the weekly-review
  framework (Prompt 6) has not been ported into PROJ-001.

### Framework / AGENTS.md lessons earned

- **§9 addendum** under the existing SPEC-005 assertion-specificity
  bullet: for markdown heading-level assertions, split into lines and
  test `ln == "# title"` rather than `strings.Contains(out, "# title")`.
  Every deeper heading is a superstring of a shallower one, so
  substring checks silently false-positive on "heading at level N
  and NOT at level M." Builder caught it; verify independently flagged
  it as generalizable. Applied alongside SPEC-015 ship (commit `85c57ac`).

- **Design-commit-before-branch rhythm** paid off on first reuse
  (SPEC-014) — squash-merge produced a clean feat commit with no
  chore/design bundled, unlike SPEC-013's squash which had included
  triage/framing/design.

- **Branch-created-during-design slip** (SPEC-015) reversed the
  SPEC-014 rhythm — one data point in each direction so far.
  Watching for a third before codifying an explicit rule.
  Captured in SPEC-015 Reflection (Ship) Q2.

- **Advisory size cap worked as intended** (SPEC-015 at 1543 lines,
  143 over the 1400 advisory cap). Session paused, defended, reviewed,
  content stood. Cap is a prompt-to-justify, not a refuse-past-N rule.

- **Non-deviations landing in Deviations section** happened twice
  (SPEC-014's echoFilters extraction; SPEC-015's assertion-shape
  fixes). Minor wording issue; if it recurs in SPEC-017, consider
  updating the spec template's Deviations-section prompt to
  distinguish true deviations from spec-sanctioned choices.

- **Template follow-up noted (not acted on):** the Premise Audit block
  has appeared with the same skeleton in SPEC-011, 012, 014, 015.
  Extracting it into `/projects/_templates/spec.md` as a reusable
  sub-template would thin every future spec by 50–80 lines. Deferred
  to weekly-review or stage-ship.

### Stage state at session end

STAGE-003 — 3 shipped / 0 active / 1 pending / 1 deferred:

- [x] SPEC-013 — `brag list --show-project`          (shipped 2026-04-23)
- [x] SPEC-014 — JSON trio + DEC-011                 (shipped 2026-04-23, M)
- [x] SPEC-015 — markdown export + DEC-013           (shipped 2026-04-24, M)
- [~] SPEC-016 — sqlite export                        (DEFERRED 2026-04-23)
- [ ] SPEC-017 — `brag add --json` + DEC-012         (pending, S, LAST)

Project-wide: STAGE-001 shipped 2026-04-20, STAGE-002 shipped 2026-04-22.
After SPEC-017 ships and STAGE-003 closes (Prompt 1d), the remaining
work in PROJ-001 is STAGE-004 (polish, provisional — may dissolve) and
STAGE-005 (distribution: goreleaser + homebrew tap + CI).

### Pick up here next session

**Next concrete action:** scaffold + design SPEC-017 (`brag add --json`
+ DEC-012). Stage Design Notes has the authoritative scope lock — 6
DEC-012 choices already decided (single-object stdin only, required
`title`, tolerate-and-ignore server fields, strict-reject unknown keys
naming the offender, inserted-ID stdout, comma-joined tags reject
array form). Size S. Structural question for design session to answer:
inline in `internal/cli/add.go` vs. `internal/ingest` package
symmetric with `internal/export`.

**Framework hygiene:** `just new-spec` auto-increments by next free
number, which after SPEC-015 archives will produce SPEC-016 (currently
a deferred slot). Either (a) scaffold, then `git mv` to SPEC-017 plus
a `task.id` frontmatter edit, or (b) accept SPEC-016 re-use and update
`backlog.md` + the stage file's SPEC-016 deferral marker. (a) is
cleaner for traceability; 30-second edit.

**Soft call to revisit:** after SPEC-017 ships, if STAGE-003 feels
complete and the user's review-prep workflow is unblocked, STAGE-004
may dissolve per the 2026-04-22 reshuffle's escape hatch — jumping
directly to STAGE-005 distribution. Decide at stage-ship time based
on whether real usage since shipping surfaces any STAGE-004 polish as
actually-missed vs. speculatively-nice.

**Untracked files worth cleaning up eventually:** `framework-feedback/`
(has legitimate content from earlier session) and
`status-after-nine-specs.md` (stale one-shot snapshot, obsoleted by
this session log). Low priority — one cleanup chore when convenient.

**Porting framework's weekly review (Prompt 6 from
`FIRST_SESSION_PROMPTS.md`) into PROJ-001:** not yet done. When this
session log has 4–6 entries, a weekly review session can read across
them + stage-level reflection + all DECs to produce the kind of
drift-detection report Prompt 6 describes. Don't pre-optimize a
template; let the pattern emerge from 2–3 sessions of log entries.

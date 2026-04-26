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

## 2026-04-26 — STAGE-005 at 2/4 + cross-format literal-artifact-as-spec validated

**Session duration (approx):** continuation of the long-running
coordinator session that started 2026-04-22 (STAGE-003 framing).
Today's segment: SPEC-022 design+build+verify+ship in one
sequence, plus pre-compaction prep.
**Branches / PRs landed:** PR #22 (SPEC-022; squash-merged
`079bb89`).
**brag entries:** #36 (SPEC-022 ship).

### What shipped

- **SPEC-022** — AI-integration distribution asset: three new
  artifacts (`docs/brag-entry.schema.json` 50-line JSON Schema
  draft 2020-12 mirroring DEC-012; `scripts/claude-code-post-session.sh`
  67-line pure-stdin bash hook with `chmod +x`;
  `examples/brag-slash-command.md` 14-line tight slash-command
  template). Plus `BRAG.md` 50-line "## JSON contract for
  programmatic capture" insertion + 161-line scripts/test-docs.sh
  extension (23 new asserts in groups H/I/J/K). All 30 AC met,
  63/63 test-docs pass (40 SPEC-021 + 23 new), no DEC, all 10
  rejected-alternatives held.

- **Examples directory created** (`examples/`) — first use of
  this directory in the repo. Established pattern: `scripts/`
  for executable artifacts, `examples/` for reference content
  the user copies somewhere. Q1 of SPEC-022 framing locked
  this split.

### Framework / AGENTS.md lessons earned

None new (no AGENTS.md addenda landed at SPEC-022 ship). The
quiet positive: every existing addendum proven correct on a
second confirming case.

- **§10 push-discipline rule** held cleanly on its second
  proactive application (SPEC-021 first, SPEC-022 second).
  Third case from SPEC-023 or SPEC-024 promotes the rule from
  "codified at framing" to "load-bearing across the stage"
  without any rewording.
- **`bfa1474` archive-spec empty-`<answer>` rejection**
  exercised second time on a real ship and PASSED (proven
  previously at SPEC-021).
- **§9 audit-grep cross-check + §12 NOT-contains self-audit
  + §9 BSD-grep `--exclude-dir` warning** — all three
  no-op'd at build for the second consecutive spec. Design
  pre-emption working as designed.
- **§12 literal-artifact-as-spec pattern (codified at
  SPEC-021 ship)** — third application, FIRST CROSS-FORMAT.
  Validated cleanly across JSON + bash + markdown + in-place
  markdown insertion + shell-script extension. Empirically
  format-agnostic. Carry to SPEC-024 + PROJ-002 framing.

One Q3 cosmetic observation surfaced: the BRAG.md insertion
literal omitted the `---` HR separator that surrounding
sections use. Verify confirmed asymmetry is cosmetic
(markdown viewers handle `##` as a strong section break with
or without preceding `---`). WATCHED at N=1, not codified
per the three-confirming-cases bar that prior §12
codifications crossed; pattern bloat from a single
observation outweighs marginal signal.

### Stage state at session end

PROJ-001 — 4 stages shipped + STAGE-005 at 2/4:

```
[x] STAGE-001 (shipped 2026-04-20)
[x] STAGE-002 (shipped 2026-04-22)
[x] STAGE-003 (shipped 2026-04-24)
[x] STAGE-004 (shipped 2026-04-25)
[ ] STAGE-005 — distribution + cleanup (2/4)
    [x] SPEC-021 — README rewrite + dev-process migration  (shipped 2026-04-25)
    [x] SPEC-022 — AI-integration distribution asset       (shipped 2026-04-26)
    [ ] SPEC-023 — Distribution proper                      (pending; M, watch-flag for split)
    [ ] SPEC-024 — Shell completions                        (pending; S)
```

Roughly on track for ~2-week MVP target — STAGE-001 shipped
2026-04-20, today 2026-04-26 = 7 days; estimate ~3 more
days for SPEC-023 + SPEC-024 + STAGE-005 close + PROJ-001
close.

### Pick up here next session

**Coordinator session context-compacted** at end of this
entry. SPEC-023 design starts in a fresh coordinator session.

**Next concrete action:** scaffold SPEC-023 + draft Design
prompt + paste into a fresh design Claude session.

```bash
just new-spec "distribution proper goreleaser gha tap changelog" STAGE-005
```

Confirm scaffolder produces SPEC-023 (next free number after
022 archived); `git mv` if not.

**SPEC-023 specifics that will inform Design:**
- Heaviest remaining spec in PROJ-001. Stage framing flagged
  watch-for-split. Design session can split into
  goreleaser+CI vs tap+CHANGELOG if scope inspection argues
  for it; otherwise bundled as one M.
- Inherits SPEC-021's deferred punch list mechanically:
  stale STAGE-NNN refs across 4 doc files (api-contract.md
  13 hits, architecture.md 4 hits including line 45 stale
  "sqlite-file-copy" claim, data-model.md 5 hits,
  tutorial.md:493 brew-install row in "What's NOT there
  yet" table). SPEC-023 doc-sweep activates these references
  the way SPEC-021's brew-install forward-reference becomes
  a live link.
- §12 literal-artifact-as-spec pattern applies directly:
  goreleaser config + GitHub Actions workflow YAMLs + tap
  formula + CHANGELOG seed are all fixed-shape artifacts
  decidable at design time. Trim heuristic: SPEC-021/022
  are construction precedents within stage; signatures +
  invariants compression COULD apply to non-literal prose
  (default fuller skeleton if uncertain — N=2 for the trim
  itself within STAGE-005).
- Homebrew tap repo at `github.com/jysf/homebrew-bragfile`
  was pre-created as a STAGE-005 framing chore (empty repo
  with one-paragraph README); SPEC-023's goreleaser `brews:`
  block points at it.

**Lessons for SPEC-023's design+build+ship discipline
(carry from SPEC-021/022 cycles):**
- §10 push-discipline applies at build (run `git push origin
  HEAD` before `gh pr merge`).
- bfa1474 archive-spec precondition will reject empty
  Reflection (Ship) `<answer>` placeholders at ship.
- §9 audit-grep cross-check + §12 NOT-contains self-audit
  apply at design (run greps; reconcile against `## Outputs`).
- §9 BSD-grep `--exclude-dir` warning applies if test-docs
  extension uses grep -r.
- §12 literal-artifact-as-spec: embed YAMLs + CHANGELOG seed
  + tap formula verbatim under Notes for the Implementer.

**Outstanding STAGE-005 sequence after SPEC-023:**
- SPEC-024 (shell completions, S) — smaller; literal-artifact
  pattern applies; per-shell completion output is fixed-shape.
- STAGE-005 ship (Prompt 1d) — same shape as STAGE-003/004
  closes; stage-level reflection across SPEC-021/022/023/024.
- PROJ-001 ship (Prompt 1e) — FIRST USE in this repo;
  project-level retrospective; closes the MVP.
- Blog post artifact — placement decided as `docs/`; publication
  target TBD. Independent of spec cycles.

**State for fresh-session pickup:**
- Working tree clean (only persistent untracked noise:
  `framework-feedback/`, `revew1.md`, `status-after-nine-specs.md`).
- No daily-status-report refresh today (per user: EOD only,
  not per-ship). Today's snapshot pending if you stop for the
  day.
- Untracked cleanup deferred to PROJ-001 close (`framework-feedback/`
  to evaluate; `revew1.md` to evaluate; `status-after-nine-specs.md`
  is stale and obsoleted by daily-status-report system).

### Compaction note

This session was the long-running coordinator thread that ran
SPEC-013 through SPEC-022 + STAGE-003/004/005 framings + 2
stage closes + ~30 brag entries (#10-#36) + 4 AGENTS.md
addenda earned + 5+ tooling chores (specs-by-stage,
daily-status-report, archive-spec precondition,
push-discipline rule, etc.). Compacted at SPEC-022 ship
boundary by user choice; SPEC-023 design starts in a fresh
coordinator session reading the artifacts above.

---

## 2026-04-25 — STAGE-004 closed + STAGE-005 framed + SPEC-021 shipped

**Session duration (approx):** full day (continuation of
2026-04-24 coordinator session).
**Branches / PRs landed:** PR #18 (SPEC-018), PR #19
(SPEC-019), PR #20 (SPEC-020), PR #21 (SPEC-021); plus
chore commits for STAGE-004 close + STAGE-005 framing
prep.
**brag entries:** #21 (SPEC-018), #23 (SPEC-019), #24
(SPEC-020), #25 (STAGE-004 close), #26-#31 (assorted
chores + tooling), #32 (SPEC-021), #33 (STAGE-005 milestone).

### What shipped

- **SPEC-018** — `brag summary --range week|month` + DEC-014
  (rule-based output envelope) + new `internal/aggregate`
  package (ByType, ByProject, GroupForHighlights, rangeCutoff
  helpers).
- **SPEC-019** — `brag review --week|--month` consuming
  DEC-014; added `GroupEntriesByProject` to
  `internal/aggregate`; refactored `internal/export/json.go`
  to extract `toEntryRecord` helper for review.go reuse.
- **SPEC-020** — `brag stats` six lifetime metrics consuming
  DEC-014; added Streak/MostCommon/Span helpers to
  `internal/aggregate`. Trim experiment validated (signatures
  + invariants sufficient when in-stage precedents exist).
- **STAGE-004 closed** — stage-level reflection drafted;
  status flipped proposed → shipped. SPEC-019 reflection
  recovery (orphaned `bef84b1` commit recovered from
  `git show` and bundled into stage-ship commit; root-cause
  of the missing-push-before-merge bug class).
- **STAGE-005 framed** — 4 specs (021/022/023/024) proposed;
  3 pre-stage chores landed (homebrew tap repo created,
  AGENTS.md §10 push-discipline rule codified,
  scripts/archive-spec.sh empty-`<answer>` rejection
  precondition added).
- **SPEC-021** — README user-facing rewrite + new
  CONTRIBUTING.md + new docs/development.md + new
  scripts/test-docs.sh harness (40 grep-based asserts)
  exposed via new `just test-docs` recipe. First STAGE-005
  spec; first doc-restructure spec for PROJ-001.

### Framework / AGENTS.md lessons earned

Three earned at STAGE-004 close + during STAGE-005 framing
+ at SPEC-021 ship:

- **§9 audit-grep cross-check (SPEC-018-earned, codified at
  SPEC-018 ship 2026-04-25)** — closes the design/build
  premise-audit loop: design enumerates → design verifies
  enumeration → build re-verifies and questions deltas.
  Validated SPEC-019 + SPEC-020 + SPEC-021.
- **§12 NOT-contains self-audit (SPEC-019-earned, codified
  at SPEC-020 ship 2026-04-25)** — when a Failing Test
  asserts output DOES NOT contain "X", grep the spec's
  load-bearing prose for X at design time. Two confirming
  cases (SPEC-019 build self-catch + SPEC-020 design
  pre-empt) crossed the codification bar because the rule
  is concrete (a grep run).
- **§9 BSD grep `--exclude-dir` warning + §12
  literal-artifact-as-spec pattern (both SPEC-021-earned,
  codified at SPEC-021 ship 2026-04-25)** — BSD grep
  matches basenames not path fragments, treat
  `--exclude-dir` as decorative; whitelist tolerable hits
  via case post-filter. Literal-artifact-as-spec: when a
  spec ships a fixed-shape artifact decidable at design
  time, embed the literal artifact under Notes for the
  Implementer; build transcribes verbatim, verify diffs.
  Three confirming cases (SPEC-018 Go fixtures + SPEC-020
  cobra Long string + SPEC-021 markdown documents) crossed
  the codification bar.

Plus operational rules codified at STAGE-005 framing:

- **§10 push-discipline rule** — any commit added to a
  feat branch just before `gh pr merge --squash
  --delete-branch` MUST be pushed to `origin/<feat-branch>`
  first. Three confirming cases (SPEC-013 + SPEC-018 +
  SPEC-019 orphaned reflection commit). Codified
  preemptively at STAGE-005 framing; first proactive
  application held cleanly at SPEC-021 ship.
- **`bfa1474` archive-spec empty-`<answer>` precondition**
  — `scripts/archive-spec.sh` fail-fasts if Reflection
  (Ship) section has unfilled placeholders. Belt-and-
  suspenders for the SPEC-019 reflection-orphan class.
  Validated as PASSING on a real ship at SPEC-021.

### Stage state at session end

```
[x] STAGE-001 (shipped 2026-04-20)
[x] STAGE-002 (shipped 2026-04-22)
[x] STAGE-003 (shipped 2026-04-24)
[x] STAGE-004 (shipped 2026-04-25)
[ ] STAGE-005 — distribution + cleanup (1/4)
    [x] SPEC-021 (shipped 2026-04-25)
    [ ] SPEC-022/023/024 pending
```

### Pick up here next session

Was: SPEC-022 design (AI-integration distribution asset).
Done same day; see 2026-04-26 entry above.

---

## 2026-04-24 (continued) — STAGE-003 closed, STAGE-004/005 cherry-picked

**Session duration (approx):** continued same-day after the
SPEC-017 ship + STAGE-003 close. STAGE-004 plan refinement
session, no code work.
**Branches / PRs landed:** none in this segment (planning only).
**brag entries:** none new (this segment was triage, not ship).

### What shipped

- **STAGE-003 closed** (commit `8120f69`): stage-level reflection
  written, `status: shipped`, `shipped_at: 2026-04-24` in
  frontmatter, brief stage plan updated, brag #20 captured.
- **STAGE-004 cherry-picked** (commit pending — this commit):
  3 specs (`brag summary --range`, `brag review --week`, `brag
  stats`) survived the user filter "will I actually use this?";
  6 items dropped to backlog (emoji 1–4, `brag remind`, Claude
  session-end hook moved to STAGE-005). `brief.md` and
  `backlog.md` updated with full reasoning.
- **STAGE-005 expanded** (same commit): added README rewrite
  (current README documents dev-process not user-facing tool
  use — caught by an external Claude review the user shared),
  `docs/brag-entry.schema.json` (mirrors DEC-012; AI-agent
  validation contract), Claude session-end hook example (moved
  from STAGE-004 as distribution asset), shell completions,
  blog post artifact. Goreleaser + brew tap + CI persist from
  earlier sketches.

### Framework / AGENTS.md lessons earned

None new in this segment. The "either-is-fine in Notes-for-
Implementer off-loads decisions to build" pattern remains at 2
data points (SPEC-014 + SPEC-017); watch for a third before
codifying.

### Stage state at session end

```
STAGE-001 — foundations              shipped 2026-04-20
STAGE-002 — capture & retrieval      shipped 2026-04-22
STAGE-003 — reports + AI-friendly I/O shipped 2026-04-24
STAGE-004 — rule-based polish (3 specs) NOT YET FRAMED
STAGE-005 — distribution + cleanup       NOT YET FRAMED
```

PROJ-001 status: 3 shipped / 0 active / 2 pending.

### Pick up here next session

**Next concrete action:** frame STAGE-004 via Prompt 1c. Stage
sketch in `brief.md` is the authoritative scope (3 specs:
`brag summary --range`, `brag review --week`, `brag stats`).
All three emit clean markdown/JSON; user pipes into external
AI manually when wanted (no LLM in PROJ-001 — that's PROJ-002).

```bash
just new-stage "rule-based polish summary review stats" PROJ-001
```

Then paste a fresh-session Prompt 1c. Framing should produce a
3-spec stage file; spec sizing likely 1×M + 2×S; design sequence
probably independent (all three read existing schema, none depend
on each other).

**Soft decision still open at framing time:** whether to add
`brag add --at <date>` (backdating) as SPEC-021 in STAGE-004.
External Claude review flagged it as "real Friday-recapping-
Tuesday value, ~30 lines." Framer can pull it from backlog
discussion at framing time, or leave it backlogged. User
hasn't strongly committed either way.

**STAGE-005 readiness:** the stage-005 sketch in `brief.md` now
has 5 workstreams enumerated (README rewrite, JSON schema file,
Claude hook, distribution proper, shell completions). Don't
frame STAGE-005 until STAGE-004 ships — keeps focus.

**Untracked files still in working tree:** `framework-feedback/`
and `status-after-nine-specs.md` at repo root. Low-priority
cleanup whenever convenient (the latter is obsoleted by this
session log; the former has legitimate content from earlier
sessions but should probably move out of repo or be added to
`.gitignore`).

**External Claude review absorbed:** an external session
reviewed PROJ-001's current state and surfaced several ideas;
most overlapped with existing backlog entries. The genuinely
new contributions captured: (a) `docs/brag-entry.schema.json`
file (now in STAGE-005); (b) `--at` backdating (open question
for STAGE-004 framing); (c) pressure-tests on DEC-004 tags-as-
string, edit-history, soft-delete (user decision: accept the
v0.1 debt). Several other ideas (`brag last`, `brag tags`,
`brag projects`, `--meta`, `--quiet`, default config file,
auto-tag Claude sessions) are nice-but-not-must, kept out of
scope to preserve STAGE-004's tight 3-spec shape.

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

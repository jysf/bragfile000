---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-018
  status: shipped                   # proposed | active | shipped | cancelled | on_hold
  priority: high
  target_complete: null

project:
  id: PROJ-006
repo:
  id: bragfile

created_at: 2026-07-12
shipped_at: 2026-08-13
---

# STAGE-018: capture-validation caps + the v0.5.0 audit backlog

## What This Stage Is

**Activated 2026-08-09**, and it now gates the v0.6.0 release. STAGE-019
shipped the memory slice, whose per-entry line is
`- <id> <date> [proj/type] <title> — <impact>` — so DEC-044 derived its entire
budget model, including the "one long entry cannot starve a slice" argument,
from the *current* caps (626 bytes / 157 tokens worst case). Cutting v0.6.0
first would ship that model knowing the caps are wrong, then change how many
entries fit a 2000-token budget one release later. The release waits for
**SPEC-075 only** — not for the remaining audit nits, which can trail into
v0.6.1.

**Reframed 2026-08-09.** This began as a parking lot for the v0.5.0 audit's
LOW/NIT items. It now has an anchor that is not a nit: **the capture-validation
caps are wrong, and they are wrong in a way that already rejects a third of the
corpus.** The nits ride along because most of them live in the same code.

The coherent outcome: **`capture.Validate` becomes a rule you can defend** —
every cap derived from what the field is for rather than inherited from a JSON
schema, enforced on *every* ingress path including edit, and expressed once
rather than duplicated. The audit nits are the same "harden the edges" pass they
always were, now with a reason to schedule them.

Three coupled pieces, all touching `internal/capture`:

1. **Re-shape the caps.** `MaxImpact = 256` is exceeded by **74 of 285 entries
   (26%)**, with a **median of 227** — a cap sitting in the middle of ordinary
   usage, not catching outliers. `MaxTitle = 200` is exceeded by 33 of 348.
   `MaxTags = 64` caps the *comma-joined string* rather than each tag, which
   penalises using many tags (legitimate) instead of one absurd tag (the actual
   abuse) — a leftover from DEC-004's comma-joined era that DEC-015's normalised
   `tags` table made obsolete. `project`/`type`/`description` are fine and must
   NOT be widened reflexively (3x, 6x and 48x headroom respectively).
2. **Enforce on the edit path.** `capture.Validate` is called from
   `cli/add.go` (x2), `cli/add_json.go` and MCP `brag_add` — but **not** from
   edit / `Store.Update`. This is the pre-existing backlog item, and it is
   **blocked on (1)**: wiring it up first would make every over-cap entry
   uneditable, so a user could not fix a title without deleting tags.
3. **Express each cap once.** `cli/project.go:162` hard-codes `len(name) > 64`
   instead of referencing `capture.MaxProject`, with a comment directly above it
   reasoning about keeping the two in agreement (DEC-036's soft-match
   invariant) — exactly the case for one constant rather than two a comment asks
   a reader to align.

## Why Now

The original answer was "not now, deliberately." That changed when a user hit
the tags cap in real use and the investigation found the pattern was general,
measured against the live corpus (347 entries): the caps were never derived.
DEC-012 introduced them for the `add --json` schema and records **no rationale
for any value**; SPEC-064 then correctly unified per-path validation and
generalised those numbers to every path — which retroactively made ~26% of the
existing corpus unwritable. Nobody measured that at the time.

It is PROJ-006's business rather than PROJ-007's: the coupling is to items
already in this project, and `impact` is the field this project's own memory
slice reads. Deferring across a project boundary would strand that coupling.

## Success Criteria

- Every cap has a **recorded rationale** tied to what the field is for. No
  surviving number that nobody can defend.
- **No entry in the existing corpus is unwritable.** The measured over-cap
  populations (74 impacts, 33 titles, 7 tag-strings) either fit the new caps or
  are explicitly grandfathered — stated, not discovered.
- `capture.Validate` runs on **every** ingress path, edit included, and that is
  true *after* the caps are fixed, never before.
- Each cap is defined **once**; no hard-coded duplicate survives.
- **DEC-044's budget model is re-derived, not silently invalidated.** It cites
  these caps twice for the memory slice's worst-case line ("626 bytes / 157
  tokens"), the argument that one long entry cannot starve a slice. A new
  `MaxImpact` moves that number.
- Each remaining audit item is fixed with a regression test, or closed WONTFIX
  with a one-line rationale.
- Full gate set green.

## Scope

### In scope

**The anchor — capture-validation caps** (see `capture-caps-were-inherited-not-derived`
in `guidance/questions.yaml` for the full measurement and reasoning):
- Re-shape `MaxImpact` / `MaxTitle`, and re-shape `MaxTags` from a joined-string
  cap into a per-tag + tag-count cap. Emits a DEC (a validation-contract change
  across four ingress paths, plus the DEC-044 re-derivation).
- Leave `MaxProject` / `MaxType` / `MaxDescription` alone — measured, ample.
- Grandfathering policy for existing over-cap rows, decided explicitly.

**The v0.5.0 audit LOW/NIT backlog** (inherited verbatim from the project brief):

- `mcp_install` atomic write (temp + rename).
- ~~MCP `list`/`search` negative-`limit` parity with the CLI.~~ **DONE — folded
  into SPEC-072** (STAGE-019), which already edited both handlers. A negative
  `limit` is now a tool error on `brag_list` and `brag_search`; `0` still means
  unlimited. This is the "fold into the stage touching the same code" rule in
  Design Notes doing its job.
- `search -foo` → clear cobra-flag error (not an FTS query).
- `brag project new` name cap — **identified**: `internal/cli/project.go:162`
  hard-codes `if len(name) > 64` instead of referencing `capture.MaxProject`,
  while the comment above it reasons about keeping the two in agreement
  (DEC-036's soft-match invariant). One-line fix, but it must travel WITH any
  change to the caps — see the question below. Also: run the edit /
  `Store.Update` path through `internal/capture.Validate`.
  > ✅ **RESOLVED by SPEC-075 / DEC-046 (design, 2026-08-09).** Both items are
  > carried by SPEC-075 and are safe there because the caps and the
  > changed-field grandfathering rule land first in the same spec. Do not
  > action either one separately. Original warning retained below for the
  > record.
  > ⚠ **COUPLED — read before actioning.** Wiring the edit path to
  > `capture.Validate` would make a large slice of the corpus
  > **uneditable**: measured today, **74 of 285 impacts (26%)**, 33 titles and
  > 7 tag-strings already exceed their caps, because flag/editor mode did not
  > enforce them until SPEC-064 unified validation. Changing one of those
  > entries' titles would then require truncating impact or deleting tags. Resolve
  > `capture-caps-were-inherited-not-derived` in `guidance/questions.yaml` in the
  > same change, or explicitly grandfather existing rows. Do not action this
  > item alone.
- `brag spark` same-second exclusive-edge.
- backup-filename same-second collision.
- empty-`type` sentinel handling.
- export-md sort id-tiebreak.
- `MergeTags` position dup.
- double-wrapped db-path error.
- `$EDITOR`-with-spaces handling.

#### Survey against v0.6.0 code (2026-08-12) — read before framing

Checked against shipped code rather than trusting the list. Several items are
sharper — or differently shaped — than their one-liners suggest, and **two are
not mechanical fixes at all**. Two remain unverified and are marked as such.

| item | status | what the code actually shows |
|---|---|---|
| `mcp_install` atomic write | **real** | `mcp_install.go:210` is a bare `os.WriteFile(target, merged, 0o644)` — no temp+rename, so an interrupted write truncates the user's MCP client config. |
| double-wrapped db-path error | **real** | `brag --db /nonexistent/x.sqlite list` → `brag: open store: open store: mkdir …`. The prefix is applied twice. |
| backup-filename collision | **real** | `backupTimeFormat = "20060102T150405Z"` (`storage/backup.go:16`) is second-resolution; two backups in the same second collide on filename. |
| export-md id-tiebreak | **real** | `export/markdown.go:220` sorts `SliceStable` on `CreatedAt` alone. Equal timestamps fall back to input (DB) order, so output is not deterministic by id. |
| `search -foo` | **partly done — REFRAME** | The message is already clear: `unknown shorthand flag: 'f' in -foo`. What is wrong is the **exit code — 2 (internal), not 1 (user error)**. A bad flag is user-actionable; `main.go` reserves 2 for internal faults. The remaining fix is exit-code classification, not the message. (`brag search -- -foo` correctly treats it as a query.) |
| `MergeTags` position dup | **real but COUPLED** | `store.go` step 1 grafts `SELECT ?, s.taggable_type, s.taggable_id, s.position` — it copies src's `position` verbatim, so after a merge an object's tag positions can gap or collide. **Do not fix blind:** what the positions should become depends on the open question `tag-ordering-projection` in `guidance/questions.yaml`. Same coupling shape as the caps/edit-path pair — resolve the question in the same change or explicitly defer. |
| `$EDITOR` with spaces | **real but a DECISION, not a fix** | `editor/launch.go:76` is `strings.Fields(v)`. That is what makes `EDITOR="code -w"` work, and it is exactly what breaks `EDITOR="/Applications/My Editor/bin/edit"`. The two cannot both work without a quoting rule, so this needs a chosen behaviour (shell-style quoting? try the whole string as a path first?), not a patch. |
| `brag spark` same-second edge | **UNVERIFIED** | No obvious boundary comparison found in `internal/spark`; needs a targeted read of the window logic before framing. |
| empty-`type` sentinel | **UNVERIFIED** | `cli/list.go:140` gives *project* a `"-"` sentinel; whether *type* is handled the same way (and where it is not) was not confirmed. |

**Framing consequences.** Five items are genuinely mechanical (`mcp_install`,
double-wrap, backup collision, id-tiebreak, `search` exit code) and cluster
cleanly into one small spec — they share nothing but their size, and none needs
a DEC. The other two verified items should NOT ride along: `MergeTags` waits on
`tag-ordering-projection`, and `$EDITOR` needs a decision recorded. Two items
still need a look before anyone commits to a count — "~9 nits" is a list length,
not a scope estimate.

### Explicitly out of scope
- The deeper agent-native pillars (memory / signed provenance / capture
  completeness / benchmark) — separate PROJ-006 stages.
- `ParseSince` wall-clock impurity — **already handled**: STAGE-017 / SPEC-068
  folded in the `since.go` clock-seam fix (audit L4). Verify before re-listing;
  do not double-count.

## Spec Backlog

Ordered list of specs composing this stage. IDs assigned at creation.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [x] SPEC-075 (shipped on 2026-08-10) — **the caps spec, and the v0.6.0 gate.** Design complete
      2026-08-09, build complete 2026-08-10; emits **DEC-046**. Locked: `MaxImpact` 256→**1024**,
      `MaxTitle` 200→**256**, `MaxTags` deleted in favour of
      **`MaxTagLen = 64` + `MaxTagCount = 32`**; `project`/`type`/`description`
      untouched. Grandfathering = **validate on write-of-changed-field** (no
      migration, no column, converges); residual over-cap = 3 impacts, 29
      titles, 0 tag violations. Carries the two items blocked on that decision —
      the edit-path wiring (`capture.ValidateChanged` from `cli/edit.go`) and
      `cli/project.go:162`'s duplicated constant — in the same spec, because
      they are only safe *after* it.
      > **DEC-044 re-derived, not left stale:** worst-case memory line
      > 626 B / 157 tok → **1450 B / 363 tok**; anti-starvation argument holds
      > (18.2% of the default budget, and skip-and-continue still applies);
      > `memory.DefaultBudget` **stays 2000**. Design also found three
      > *pre-existing* errors in DEC-044, corrected by DEC-046: the 626 B bound
      > is already false of the live corpus (measured max line **1483 B /
      > 371 tok**, so the new bound is *tighter* than today's reality); the
      > "~18 tokens/entry" figure is the **8-entry test fixture's** mean
      > (measured corpus mean ≈ **92**); and "2000 buys ≈110 entries" is wrong
      > by ~4.4× (measured `Included: 25`, `Skipped: 175`).
- [x] (shipped on 2026-08-13, no spec) — the remaining audit nits. Landed as a
      single reviewed batch (PR #151) rather than the two clustered specs the
      Design Notes anticipated: the survey found only **seven** items were
      mechanical, they touched four packages with no shared design question
      between them, and a spec per cluster would have cost more review than the
      diff. Each fix carries a regression test. **Two items were deliberately
      NOT fixed** and are carried below rather than silently dropped.

- [x] SPEC-077 (shipped on 2026-08-13) — the **v0.6.1 release cut** that
      delivered this stage's work to users. Filed under STAGE-018 because that
      is what it ships; the stage was already closed on 2026-08-13 when its
      backlog completed, since unlike STAGE-019 its unit is the audit backlog
      rather than a release.

**Count:** 3 shipped (SPEC-075, the audit-nit batch, SPEC-077) / 0 active /
0 pending — **backlog complete.**

### The two items deliberately carried forward, not dropped

Both were verified real. Neither is a patch, and that is why they are not here:

- **`MergeTags` position dup** (`storage/store.go`) — the graft copies
  `s.position` verbatim, so after a merge an object's tag positions can gap or
  collide. What the positions *should* become is the subject of the open
  `tag-ordering-projection` question in `guidance/questions.yaml`. Fixing it
  blind would be inventing that answer in a nit batch. Same coupling shape as
  the caps/edit-path pair this stage already navigated once.
- **`$EDITOR`-with-spaces** (`editor/launch.go:76`) — `strings.Fields` is
  simultaneously what makes `EDITOR="code -w"` work and what breaks
  `EDITOR="/Applications/My Editor/bin/edit"`. The two cannot both work without
  a chosen quoting rule (shell-style parsing? try the whole string as a path
  first?). That is a decision to record, not a line to change.

Whoever picks these up should read them as two small specs, not as leftovers.

## Design Notes

- Batch by touched package, not by audit order — several of these live in the
  same file and share a test. One spec per cluster keeps review cheap.
- Prefer folding an item into a deeper-pillar stage when that stage already
  edits the same code, rather than a standalone visit. (SPEC-072 did exactly
  this with the negative-`limit` item — already ticked above.)
- **Ordering is load-bearing, not a preference.** The caps decision must land
  before the edit-path wiring. Reversed, the wiring makes 26% of the corpus
  uneditable; there is no ordering in which that is acceptable.
- The caps spec is the one that needs a **DEC**; the nits do not. Do not let the
  nits' cheapness set the review altitude for the caps change.
- Resist widening every cap. The measurement says `description` has 48x headroom
  and `impact` is over-subscribed — the fix is per-field, derived from what the
  field is *for*, not a uniform multiplier.

## Dependencies

### Depends on
- Nothing new for the nits — all against shipped v0.5.0 code.
- The caps spec depends on **DEC-044** (shipped, STAGE-019): its worst-case-line
  math has to be re-derived in the same change, not after.

### Enables
- Writing a brag without compressing the one field every digest reads. `impact`
  is what `brag impact`, `wrapped`, `story` and `memory` all surface, and what a
  spec's Reflection-Q4 is transcribed into.
- The edit path finally validating like every other ingress path.
- A marginally more robust substrate for the deeper PROJ-006 stages that build
  on the same MCP / capture / storage paths.

## Stage-Level Reflection

- **Did we deliver the outcome in "What This Stage Is"?** **Yes.**
  `capture.Validate` is now a rule you can defend: every cap derived from what
  its field is for (SPEC-075/DEC-046), enforced on *every* ingress path
  including edit, and expressed once — `cli/project.go`'s hard-coded `64` is
  gone. The measured harm is repaired: 74 of 285 over-cap impacts became 3, and
  the edit path validates on write-of-changed-field so no existing row became
  uneditable. Seven of the nine audit nits shipped with regression tests; the
  other two are carried above as specs-in-waiting, with reasons.

- **How many specs did it actually take?** **One spec plus one unspec'd batch**,
  against a plan of "one anchor spec + 2-ish nit clusters". The anchor
  (SPEC-075) was correctly sized and needed a DEC. The clusters were not: the
  survey found only seven mechanical items across four packages with no shared
  design question, so one reviewed batch beat two ceremonial specs. **The
  Design Notes' "batch by touched package" rule was right in spirit and wrong
  in arithmetic** — it assumed the nine items were nine fixes. Two were
  decisions and two needed verification before anyone could size them.

- **What changed between starting and shipping?** The stage started as a
  parking lot and became the v0.6.0 gate — then, at survey, the nit list itself
  turned out to be partly wrong, which is the actual story.

- **Lessons that should update AGENTS.md, templates, or constraints?**
  - **A backlog list is a hypothesis, not an inventory.** Surveying the nine
    against shipped code before framing changed the plan materially: one item
    (`search -foo`) was already half-fixed and its real defect was a different
    thing entirely (exit code 2 vs 1); two were decisions, not patches; two
    could not be sized without reading the code. "~9 nits" is a list length.
    Cost of the survey: under an hour. Cost of skipping it: two specs framed
    around the wrong work.
  - **A test can pin the right behaviour through the wrong mechanism.** Two
    tests asserted `!errors.Is(err, ErrUser)` as a *proxy* for "cobra produced
    this, not our RunE". The decision they locked was untouched by the fix, but
    the proxy broke — so the correct move was replacing the proxy, not
    overturning the decision. Read what a failing assertion is *for* before
    changing either side.
  - **When a fix and a DEC disagree, the DEC is usually right.** The first
    backup-collision fix probed for a free filename; it broke
    `TestBackup_FailureAbortsOpenAndLeavesDBUnmigrated`, and the failure was
    correct — dodging works around a state DEC-021 exists to stop on. Widening
    the timestamp fixed the same nit while preserving the decision.
  - *(Already actioned during v0.6.0/v0.6.1: release pre-flight now carries
    W1/W2/W3 for the plugin pin, compare-links and doc version claims, plus
    items for cross-shape upgrades and package-manager policy.)*

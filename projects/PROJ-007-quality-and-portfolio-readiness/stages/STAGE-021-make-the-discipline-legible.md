---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-021
  status: shipped
  priority: high
  target_complete: null

project:
  id: PROJ-007
repo:
  id: bragfile

created_at: 2026-08-15
shipped_at: 2026-08-20
---

# STAGE-021: make the discipline legible

## What This Stage Is

A stranger who lands on this repository can find the engineering discipline that
already exists. **Not writing documentation — building an index to work that has
already been done.**

The gap is not quality, it is discoverability. Counted 2026-08-15, the repo
holds **45 decision records, 75 archived specs, a pre-distribution security
review, a three-project cross-repo retrospective, two blog posts, four
framework-feedback documents and four research syntheses** — behind a
**1,186-word README with no entry point to any of it.** Every artifact a
reviewer would find persuasive is present and unreachable.

This stage ships when the evidence is one click from the front door.

## Why Now

**Because it is the differentiated half, and PROJ-007's other stage is not.**
Lint and coverage are table stakes: every competent Go repo has them, and having
them proves only that the author knows the convention. What this repo can show
that almost none can is *mutation-checked tests, mechanical guards that replaced
remembered ones, decision records that log their own corrections, and reflections
that name what was wrong.* A hiring manager can get a coverage badge anywhere.

There is also a sequencing argument: **the docs stage tells you what the lint
stage should claim.** Writing the practices page forces an honest inventory of
what the test regime actually guarantees, and that inventory is what a coverage
number should be presented against. Doing it in the other order invites a
percentage with no story attached — and this repo has already been bitten four
times by a coverage sentence that sounded true and pinned nothing (SPEC-073).

## Success Criteria

- A reader arriving at the README can reach the DEC log, the spec archive, the
  byte-exact test regime, the security posture and the incident→hardening pattern
  **without knowing the repo's directory layout.**
- Every claim on the practices page is **backed by a countable artifact** — a
  path, a count, a named test — not an adjective. No "rigorous", no
  "comprehensive"; if it cannot be pointed at, it does not go on the page.
- `go doc` on the public surface reads as intentional rather than as leftover
  comments.
- `decisions/` has no unexplained numbering gap.
- `guidance/questions.yaml` describes **live uncertainty only** — every entry is
  either genuinely open with a stated resolve condition, or closed.
- The counts on the page are **derived, not typed** (or pinned by `test-docs`),
  so they cannot rot the way the README version line did through two releases.

## Scope

### In scope

1. **An engineering-practices entry point.** A README section plus a document it
   points at, covering: the DEC log and how decisions get corrected rather than
   silently replaced; the spec-driven cycle; byte-exact golden discipline;
   the `test-docs` assertion harness (W1–W4 replaced four previously-remembered
   checks with mechanical ones); the security posture; and the
   incident→hardening pattern (DEC-021 backup, DEC-038 concurrency, the v0.5.2
   tap-token saga, the cask→formula upgrade cliff).
2. **A godoc pass** over the public surface — `internal/` packages that are
   really the API of this codebase, and `cmd/brag`.
3. **`SECURITY.md`** — *(landed 2026-08-15, ahead of framing)*. A real policy
   pointing at the checked-in 2026-04-26 review.
4. **The `DEC-041` gap.** `decisions/` jumps 040 → 042. The reservation belongs
   to the deferred SPEC-070 (`brag project goto`) and its unwritten
   "multi-location primary policy" decision. The gap is explained **in the
   backlog**, which a browsing reader never sees. Land a reservation tombstone
   in `decisions/` that names the holder and links the backlog item, so the gap
   stops reading as a lost decision.
5. **Open-questions hygiene.** **DONE — 6 of 18 open**, closed by SPEC-080 from 8.
   *(The count is now derived by `scripts/inventory.sh` and pinned by `Y4`/`Y5`,
   which is the point: this line was wrong three times in four days before the
   spec that fixed it — 8 of 18 with a wrong total, 9 of 18 wrong on both halves
   from a grep matching the file's own header comment, and 7 of 17 correct until
   a merge landed a new question.)*

   Two of the three long-dormant 2026-04-19 questions were closed as
   **answered-in-practice**: `editor-template-format` by **DEC-009**, written
   2026-04-20 — *the day after the question was raised* — and
   `summary-grouping-heuristics` by `aggregate.GroupForHighlights`.
   `shareable-ids` stays open with a sharpened note; its trigger has not fired.
   No answers were invented to shrink the register.

   *`tag-ordering-projection` was closed ahead of this spec (2026-08-18) and is
   the worked example of what this item is for.* It had been open since
   2026-06-06 carrying its own falsifiable trigger — drop the `position` column
   if no production entry is observed with unsorted tags. Nobody had run it.
   Measured: **260 of 278 tagged entries are not in name-ASC order**, so the
   trigger fails and `position` stays. Resolved by observation, no DEC — the
   pattern this hygiene pass should look for is exactly that: **a question whose
   own stated condition has already been decided by reality, waiting only for
   someone to check.**
6. **The `pr:` cross-repo collision note in `BRAG.md`** — *(landed 2026-08-15,
   ahead of framing)*.

### Explicitly out of scope

- Lint, coverage, and any CI change — STAGE-022.
- Any feature work, any new command, any change to what the binary does.
- A `CODE_OF_CONDUCT.md`. Deliberate: it is boilerplate for a single-maintainer
  project, and adding it is exactly the cargo-culting this repo's story is that
  it does not do. A sharp reviewer reads it as padding.
- Rewriting the tutorial or the blog posts. They exist; this stage links them.

## Spec Backlog

Ordered list of specs composing this stage. IDs assigned at creation.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [x] SPEC-079 (shipped on 2026-08-18) — **the practices entry point.** README section + the
      document it points at; every claim backed by a countable artifact; counts
      derived or pinned rather than typed. Framed 2026-08-16, **GO at complexity
      M** — the counting rule is a real design problem, not a formality. Framing
      caught two claims that do not survive contact with the code (there are no
      "golden files"; the corrections claim has no counting rule) and three
      already-stale counts in this project's own brief.
      **Designed 2026-08-16.** All four forks settled with rejected alternatives
      (LD1 counting rule = one derived script + one whole-table diff; LD2 split
      point = the README carries routing and *zero* numbers, because guard
      coverage is the split; LD3 honest-close = its own named section,
      second-to-last, foreshadowed from the intro, by citation; LD4 = do NOT
      create a `## Amendment` convention — cite, and carry a derived count).
      Plus LD5: `test-docs` A1's band was already at its ceiling (README exactly
      250), so it is **re-pinned tight to 260**, not widened. Design re-measured
      every count: framing's **164 doc assertions is wrong and unstable** (it
      double-counts `S3` and is environment-dependent); the reproducible figure
      is 163 distinct ids, 171 after this spec. Also stale: "four W-series
      guards" (there are six) and the brief's "zero deprecated" (DEC-004 is
      superseded by DEC-015 — which strengthens the story rather than damaging
      it).
      **Built 2026-08-18** on `build/spec-079-practices-entry-point`. All six
      literals transcribed; `just test-docs` green at **171 distinct assertion
      ids** (was 163, delta exactly `X1`–`X8`), `just test` / `gofmt -l .` /
      `go vet ./...` unaffected. `X3` earned its keep on day one: the design
      snapshot recorded **7 projects**, but PR #165 (`PROJ-008`/`PROJ-009`
      scaffolds) merged ~2h before the design PR, so the true count on `main`
      is **9**. Transcribing literal ① verbatim made `X3` the *only* failing
      assertion, naming exactly one drifted row; the fix was the guard's own
      mechanical remedy (`just inventory`, paste). The counting guard caught a
      stale count before a human read the page — which is the whole thesis of
      this spec, demonstrated at its own expense.
      **Verified 2026-08-18 — ⚠ PUNCH LIST (3 items), back to build.** The
      mechanical half is clean: all six literals byte-faithful, `X3` confirmed
      passing for the right reason (both sides non-empty and byte-identical,
      and it goes red when the *repo* drifts, not just the page), the id delta
      exactly `X1`–`X8` in both directions, and all eight Group X assertions
      plus the re-pinned `A1` mutation-tested with the mutant proven real. The
      `Projects` deviation was the correct resolution and is honestly recorded.
      What did not survive is the page's own standard, in three prose claims no
      guard reaches: it says the `TestMemoryCmd_EndToEndMarkdownGolden` comment
      lists **the budget** as unpinned when the comment lists the declared
      Matched order and the golden provably *does* pin the budget
      (mutation-confirmed); it routes the first three "does not measure" gaps to
      STAGE-022 when STAGE-022 lists benchmarks under **Explicitly out of
      scope** and never mentions `test-docs` at all; and it says every archived
      spec ends with a build-phase reflection when **six of 75 lack the standard
      `### Build-phase reflection` heading — of which five carry none at all,
      SPEC-046 having written one under `### Honest reflection`.** Each is a
      claim backed by a path that contradicts it — the exact failure the page
      exists to prevent. Fixed in both the page and literal ①; the delta was
      re-verified (✅ 2026-08-18) before ship.
      *(This sentence itself first read "six of 75 carry none", which the
      re-verify caught: the grep is syntactic, the claim was semantic, and they
      diverge on exactly one spec. Recorded rather than quietly corrected —
      it is the same defect class, in the stage file that owns it.)*
- [x] SPEC-080 (shipped on 2026-08-19) — **the godoc pass + the two legibility repairs.**
      Framed 2026-08-19, **GO at complexity S** — resized *down* by
      measurement. The brief's "no surfaced godoc" implied missing doc
      comments; there are **175 exported declarations and only 3 lack one**,
      all conventional interface methods Go style does not ask you to
      document. The real gap is **7 of 15 packages with no package comment**
      — `cmd/brag`, `cli`, `config`, `export`, `mcpserver`, `storage`,
      `story`, i.e. the largest ones, which is how it went unnoticed. Same
      shape SPEC-079 found: the discipline exists and is not surfaced. Two
      forks left to design — what shape the `DEC-041` marker takes (and what
      it does to `inventory.sh`'s decision count), and whether the hygiene
      pass derives its own number.
      **Designed 2026-08-19.** Re-measurement caught the framing pass's own
      counting error before locking anything else: **191 exported
      declarations (not 175)** — a grep-shaped heuristic does not credit a
      parenthesized `const (…)`/`var (…)` block's doc comment to every name
      inside it, which is the rule `go/doc` actually follows — and **5
      packages missing a comment (not 7)** — `internal/export` and
      `internal/mcpserver` already had one; framing's grep said otherwise.
      The undocumented-symbol list (3) and package count (15) both held.
      Both forks settled with rejected alternatives: **Fork 1** — a
      reservation tombstone lands in `decisions/` (`DEC-041`, marked
      `insight.type: reservation`), and `inventory.sh`'s Decision records
      count is redefined to filter on that field (stays 45; a new row counts
      the reservation, 1) rather than counting every file matching the glob.
      **Fork 2** — yes: `inventory.sh` gains two derived rows (18 questions
      tracked, 6 open), anchored on the real entries' 4-space `status:`
      indent, which structurally cannot match the file's own `#`-prefixed
      header comment — the exact bug that produced "9 of 18." Per-question
      triage closed 2 of the 8 open questions as **answered-in-practice**
      (`editor-template-format` by `DEC-009`, one day after it was raised;
      `summary-grouping-heuristics` by `SPEC-018`/`DEC-014`'s
      `GroupForHighlights`, project-only grouping) — the same shape
      `tag-ordering-projection` closed on, at SPEC-079. One more
      (`shareable-ids`) gets a sharpened resolve-condition note; five stay
      genuinely open on their own evidence. Every literal (five package
      comments, the tombstone, the `inventory.sh`/`test-docs.sh`/
      `questions.yaml`/practices-page diffs) was staged locally and run
      through `gofmt`, `go vet`, `go build`, `go test ./...`, and the full
      `test-docs.sh` harness before being reverted to literals — all green;
      `just test-docs` reaches **176** distinct assertions (was 171).
      Confidence 0.90.
      **Built 2026-08-19** on `build/spec-080-godoc-and-legibility-repairs`.
      All six literals transcribed byte-for-byte — verified by diff against
      each literal, not by eye. Order followed the invoking instruction:
      `inventory.sh` (③) and the `DEC-041` tombstone (②) landed first;
      `./scripts/inventory.sh` was re-run against that intermediate state
      (Decision records 45, reserved 1, both already correct) and again
      after ④/⑤ landed, producing **176 / 18 / 6** — exactly what literal ⑥
      predicted, so no reconciliation was needed before pasting it. `just
      test-docs` is green at **176 distinct assertion ids** (delta exactly
      `Y1`–`Y5`, confirmed by `comm`); `just test`, `gofmt -l .`, and
      `go vet ./...` are unaffected; `go build ./...` exits 0; `go doc` on
      all five named packages was run individually and renders the new
      comment in full. `scripts/status.sh` reports 46 decisions, as the
      design predicted (informational-only off-by-one, left untouched).
      No deviations from the six literals.

- [x] SPEC-081 (shipped on 2026-08-21) — **the README call-to-action + the A1 metric switch.**
      Framed 2026-08-19, **GO at complexity S**. Promotes the existing MCP
      call-to-action out of the `> **Status:**` blockquote (readers skip status
      blocks; the full MCP section is at line 219 of 260), and switches `A1`
      from `wc -l` to `wc -w`. The metric is the real fix: A1 counts lines on a
      **hard-wrapped** file, so a pure rewrap turns CI red having added nothing —
      and a guard that fires on non-events gets disarmed by habit, which is the
      failure SPEC-079's LD5 refused. Framing pre-checked two constraints design
      would otherwise have hit: `assert_line_count_band` has **six callers**, so
      the shape is a sibling helper rather than a repurpose; and `W3` pins the
      literal `> **Status:** v<latest> shipped` line against the CHANGELOG, so
      the blockquote cannot simply be deleted.
      **Designed 2026-08-20.** All four forks settled with rejected
      alternatives. Framing's headline numbers were re-measured and held (260
      lines / 1,268 words; six callers; `W3` at `test-docs.sh:1280`); one
      supporting claim did not — the practices page does **not** cite `A1`
      anywhere, so the id's stability matters for the commit trail and the
      derived id count, not for a citation that does not exist. **The metric
      argument is demonstrated, not asserted:** rewrapping the README's prose
      across 64–100 columns, token stream asserted byte-identical at every
      width, moves `wc -l` across **248…267** while `wc -w` stays at **1268** —
      four of seven rewraps turn CI red having added nothing, and the file is
      not consistently wrapped today (68 vs 77–79 columns by section), so a
      single normalising pass would do it. The honest limit is stated too: a
      `>` or `-` marker is itself a word, so blockquote reflow moves both
      counts, by ±4 against a 500-word band where it was ±4 against a 160-line
      one. **Band = `900 1400`**, chosen without reference to Change 1 (which
      costs +9 words) and argued from three legs — fire before the ~1,500-word
      skim threshold rather than at it; keep the README's tier below the
      smallest deep-dive doc in the repo at 2,120 words; and headroom that only
      content can spend, which is the asymmetry a line budget cannot offer.
      Converted at the README's own 4.88 words/line the old band was
      ~490…1268, so **the admissible span narrows from 780 to 500 — 36% —**
      and the switch tightens the guard overall. Sibling helper
      `assert_word_count_band`; `A1` keeps its id; the other five callers stay
      on lines because the defect is only LIVE where reflow swing exceeds
      headroom, and measured, only A1 qualifies (headroom 0, swing +7) — X7 is
      next at 17/+15 and gets a written trigger rather than an intention. New
      `A11` holds the call-to-action in place (a heading in the first third,
      naming `brag mcp install`), taking `test-docs` to **177**; the Status
      blockquote keeps everything but the call-to-action, with `W3`'s line
      untouched. V2 declined with reasons — `windowFlagNames` is unexported, so
      it sits outside this stage's own `go doc` criterion — and re-recorded
      with its corrected path (`internal/cli/window.go`, not
      `internal/storage/`) and corrected fact (three callers, not two:
      `coverage` shares the flag set). Every literal was staged in a real tree,
      the harness run in three configurations, eight mutants killed, and the
      tree reverted clean. Confidence 0.88.
      **Built 2026-08-20** on `build/spec-081-readme-cta-and-a1-metric`. All
      five literals transcribed byte-for-byte — verified by diff against each
      literal, not by eye. Order followed the invoking instruction: ②③④
      landed on `scripts/test-docs.sh` first, `bash -n` clean, then ① on
      `README.md` (260 → 262 lines, 1268 → 1277 words — exactly the predicted
      +2/+9), then `just test-docs` failed on exactly one assertion (`X3`,
      the stale inventory), then ⑤ (`just inventory`, pasted) — the script
      printed **177**, agreeing with the spec's prediction, so nothing needed
      reconciling by hand. `just test-docs` reached **ALL OK** at **177**
      distinct assertion ids (delta exactly `A11`, confirmed both directions
      via `comm` against `inventory.sh`'s own dedup pipeline — not the raw
      `OK:` line count, which is 178 for unrelated reasons). The mutation
      check requested at build time reproduced §12(b) run 4's last mutant:
      mistyping `wc -w` as `wc -l` inside the new helper sent `A1` red on its
      floor (`262 words` reported against a 900 floor); reverted, and the
      helper re-diffed byte-identical to the literal. `assert_line_count_band`
      diffed byte-identical to `main`, still called by exactly `C2`/`D2`/
      `J2`/`T2`/`X7`. `README.md:10` diffed byte-identical to `main`; `W3`
      green. `just test`, `gofmt -l .`, `go vet ./...` all green. One
      deviation, reported not silently fixed: Fork 3's "third of ten `##`
      headings" undercounts by one — the pre-change README has exactly 10
      `##` headings (confirmed against `main`), so post-change with the new
      section it's the **3rd of 11**, not "3rd of 10." No literal was wrong;
      this is editorial prose in Fork 3's argument. No new DECs. Not pushed;
      no PR opened, per instruction.
      **Verified 2026-08-21 — ✅ APPROVED.** All nine acceptance criteria hold
      and all four gates are green. The mechanical half is clean end to end:
      ①②③④ diff byte-identical against the spec, ⑤ round-trips against
      `./scripts/inventory.sh` with both sides non-empty (matching md5, and
      `X3` carries an explicit empty-block guard so a blank paste would fail
      rather than pass), and the **code-only** diff of `test-docs.sh` against
      `main` — comments stripped — is exactly three changes and nothing else.
      Eight mutants re-run from scratch, each hash-verified as a real change
      before the harness was run and the tree confirmed clean after: `A1` goes
      red on its floor when `wc -w` is mistyped as `wc -l` (`262 words`), and
      responds at **both** band edges reporting **1277** — the word count, not
      262 — so it passes for the right reason; all three `A11` mutants killed.
      Fork 2(a) was checked empirically rather than taken on its word:
      repurposing `assert_line_count_band` to `wc -w` sends exactly
      `C2`/`D2`/`J2`/`T2`/`X7` red and nothing else, so the five untouched
      callers really are the guard the spec declined to add; the helper's body
      diffs byte-identical to `main`. The reflow demonstration was
      re-implemented independently and reproduces the design's table line for
      line (248…267 lines, `wc -w` pinned at 1268 at every width). `W3` green,
      `README.md:1-14` byte-identical to `main`, and the blockquote reads
      coherently with the call-to-action gone — it was a new thought, not a
      completed one, so nothing dangles. Three findings, all **recorded, none
      blocking**, none touching a literal, an assertion or a count. (1) Fork 3's
      ordinal: build fixed the denominator and left the ordinal, so **"3rd of
      11" is true under no single reading** — measured fence-aware, the
      call-to-action is 2nd of 11 `##` headings, 3rd of 12 counting the H1, 4th
      of 13 in GitHub's rendered outline. The correct number strengthens Fork
      3's argument rather than weakening it. **This is the third instance in
      this stage of an `X of N` claim re-derived only on the challenged half**
      (SPEC-079's "six of 75 carry none", and the 8-of-18/9-of-18 open-questions
      family) — N=3 same-outcome on one mechanical sub-rule, which clears §12's
      own codification bar. Recommended for codification at stage ship: *re-derive
      both halves, not only the half that was challenged.* (2) `A1`'s comment
      calls `docs/for-ai-agents.md` (2120) *"the smallest deep-dive doc in this
      repo"*; `docs/architecture.md` (1577) and `docs/data-model.md` (1908) are
      smaller. The claim is exact for the docs the README **routes to**, which is
      the set Fork 1's tier list uses — "in this repo" over-reaches. The
      conclusion survives (1400 still sits below every long-form doc), the stated
      ~700-word cushion does not (47 to the nearest, 177 to the nearest
      unambiguous deep-dive). Not a re-litigation of the band, which the 780→500
      narrowing settles independently. (3) `A11`'s comment says *"A1 above bounds
      that length"*, but `A1` now measures words while `A11`'s threshold is
      `wc -l / 3` — and `test-docs.sh:275` is now the only place the harness
      reads the README's line count at all. Checked before recording: the guard
      does not depend on it — insertions above the heading trip `A11` for any
      `k > 20` with no help from `A1` — so the sentence is decorative, worth a
      word-swap when the comment is next touched. Also sampled seven claims
      neither cycle touched (no golden files, three named tests, archive-spec's
      placeholder rejection, 77/77 ship reflections, the README's "five tool
      schemas"): all held against their citations.

**Count:** 3 shipped / 0 active / 0 pending — **the backlog is complete.**

> `just archive-spec` printed *"All specs for STAGE-021 are shipped"* at each of
> the three ships. It was **false the first two times** — it counts written
> specs, not backlog items, so an unframed entry is invisible to it. This time
> it is true. The warning that stood here through SPEC-079 and SPEC-080 has done
> its job and is retired with the record of why it existed.
>
> **The stage is not closed until its Stage-Level Reflection is written.**

*(The trailing fragment that sat here — the tail of a parenthetical about
SPEC-081 being "agreed and not yet scaffolded" — was orphaned when SPEC-081 was
scaffolded on 2026-08-19 and its opening was replaced by the backlog entry
above. Removed at SPEC-081 design, 2026-08-20: it was a stale claim about this
spec, in the stage file about legibility.)*

> **The stage is NOT complete.** `just archive-spec` printed *"All specs for
> STAGE-021 are shipped"* on an earlier ship — it counts written specs, not
> backlog items, so an unframed entry was invisible to it. SPEC-080 shipped
> 2026-08-19; **SPEC-081 is now framed and designed but not yet built,
> verified, or shipped.** Do not run the Stage Ship prompt until it is.

## Design Notes

- **Point at artifacts, never adjectives.** The persuasive version of this page
  is "45 decisions, 4 of which record their own correction — here they are,"
  not "we take decisions seriously." Every sentence should survive the question
  *what would I click to check that?*
- **Derive the counts.** The README version line was wrong through two releases
  (SPEC-077) until a test derived it from the CHANGELOG. A practices page full
  of hand-typed counts is the same defect waiting to happen — either compute
  them or pin them in `test-docs`.
- **The honest-close material is an asset, not an embarrassment.** PROJ-006
  closed as an explicit scope reduction; DEC-043's first correction was itself
  wrong on a second axis. Most repos cannot show a decision log that records
  being wrong. Surface it deliberately rather than quietly.
- Prefer one document plus a README section over a docs site. A site is a
  distribution project; this is an index.

## Dependencies

### Depends on
- Nothing. Every input already exists in the repo — that is the premise of the
  stage.

### Enables
- **STAGE-022** — the practices page is what a coverage number gets presented
  *against*, so writing it first tells the lint stage what it is allowed to
  claim.
- The PROJ-007 outcome generally: quality that is legible to an outside reader.

## Stage-Level Reflection

*Written at stage ship, 2026-08-20. Every figure below was re-measured against
the repo rather than carried forward — which is, in part, what this stage turned
out to be about.*

- **Did we deliver the outcome in "What This Stage Is"?** **Yes — all six
  success criteria, each checkable.** A reader reaches the DEC log, the spec
  archive, the test regime, the security posture and the incident→hardening
  pattern from a named README section. **15 of 15 packages** render a doc comment
  under `go doc` (5 were missing; the other 10 already had one). `decisions/`
  has no unexplained gap — `DEC-041` carries a reservation tombstone typed
  `insight.type: reservation`, so the decision count stays honest at 45 with a
  sibling row for the reservation. `guidance/questions.yaml` is **6 open of 18**,
  down from 8, and **all six carry a stated resolve condition**. Every
  current-state number on the practices page is derived by `scripts/inventory.sh`
  and diffed by `X3`.

  **What the stage did not achieve, stated plainly:** the guards reach *numbers*,
  not *claims*. Across three specs, **eleven findings** were returned by verify,
  and the ones that mattered were prose contradicted by the path it cited — a
  golden that provably *did* pin what a page said it didn't; a scope sentence
  refuted by the stage it pointed at; "the only layer that imports a SQL driver"
  when four `internal/cli` test files import the driver. **`X3` cannot see any of
  those, and `X5` was green through all of them.** They were caught by
  human-shaped reads in verify, twice returning punch lists. That is the honest
  limit of mechanising a documentation page, and the practices page now says so
  in its own voice rather than implying broader coverage.

- **How many specs did it actually take?** **Three, against a plan of two.**
  SPEC-081 did not exist at framing; it was created *by* SPEC-079's LD5, which
  re-pinned the README's length guard tight at 260 with zero headroom rather than
  widening it. That decision was right and it had a cost — every subsequent
  README edit would trip the guard — so a third spec was needed to fix the
  *metric* rather than the bound. **A correct decision generating necessary
  follow-on work is not scope creep; it is the plan being wrong about cost.**

  The cycle count is the more interesting number: **nine PRs, six verify passes,
  three punch lists, eleven findings.** SPEC-080 alone took three verify passes
  and two return trips.

- **What changed between starting and shipping?** The stage was framed as
  documentation work and turned out to be **measurement work**. Two of its three
  specs were resized *down* by re-measuring their own premise: SPEC-079's framing
  claimed 164 doc assertions (163, and environment-dependent); SPEC-080's claimed
  175 exported declarations with 7 packages missing comments (191 and 5 — a
  per-name grep does not credit a doc comment on a parenthesized `const` block,
  and the package check only read the alphabetically-first file per directory).
  Neither was stale. Both were **grep-shaped heuristics that read as
  authoritative**, against an unchanged tree.

- **Lessons that should update AGENTS.md, templates, or constraints?**
  - **An *X of N* claim is two measurements plus a unit — when one half is
    challenged, re-derive all three.** This reached §12's own N=3 same-outcome
    bar inside a single stage: SPEC-079's "six of 75 carry none" (six lacked the
    *heading*; five carried none); the open-questions family that was wrong three
    times in four days as 8-of-18, 9-of-18 and 7-of-17; and SPEC-081's heading
    ordinal, where build fixed the denominator and carried the wrong ordinal, so
    that **no single reading made the corrected claim true.** Each pass fixed the
    half it was handed. **Recommended for §12.**
  - **When a literal caches derived output, the derivation outranks the cache —
    re-run it as build step one.** Now N=2: SPEC-079's literal cached a project
    count that went stale in a two-hour window between PRs, and SPEC-081's cached
    an assertion id count its own change moved. Below the N=3 bar; **record and
    hold.**
    **RESOLVED at SPEC-082 ship (2026-08-21).** Reached **N=3** — SPEC-082's
    `Y4` pinned the two question-register counts its own new question moved.
    Codified into `AGENTS.md` §9, with the second half the third case added:
    enumerate every guard that caches a value *before* writing `## Failing
    Tests`, not just the one the spec is already thinking about.
  - **`just archive-spec` reports stage completion off written specs, not backlog
    items.** It printed "All specs for STAGE-021 are shipped" at all three ships
    and was **false twice**. Worth a fix in the template's tooling, and worth
    knowing until then.
    **FIXED at SPEC-082 ship (2026-08-21)**, after a third false positive in
    STAGE-022. `scripts/archive-spec.sh` now reads **both** sources — spec files
    *and* unchecked `- [ ]` items in the stage's `## Spec Backlog` — and reports
    "all WRITTEN specs are shipped, but N backlog item(s) remain" instead of
    claiming completion. Fixing it immediately surfaced a stale record this
    stage had carried since its own close: `SPEC-081` was still `- [ ]` at
    `(verify)` while the count line beneath it already read *"the backlog is
    complete."* Corrected in the same pass.

- **Should any spec-level reflections be promoted to stage-level lessons?**
  - **Yes: "test the claim, not the counterexample."** SPEC-080 hit the same
    defect three times inside one spec — a false claim, one counterexample named,
    a fix narrowed just enough to survive *that* example, still false in general.
    `package` → `layer` fixed `storagetest` and stayed false for `internal/cli`.
    The round that worked tested the whole sentence and found a *second* unscoped
    claim eleven lines down that no finding had named. This is the same shape as
    the *X of N* rule above and probably its general case.
  - **Yes: a guard is only as honest as its unit.** SPEC-079 correctly refused to
    widen A1 when it fired; SPEC-081 showed the guard had been measuring the
    wrong *event* all along — a pure rewrap of the same 1,268 words moved
    `wc -l` across a 19-line range while `wc -w` did not move at all. **Refusing
    to weaken a guard and questioning what it measures are not in tension**, and
    the second is the better move when available.
  - **No** to promoting the reservation-type seam or the depguard routing — both
    are STAGE-022's inheritance, already recorded there with evidence.

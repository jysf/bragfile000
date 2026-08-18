---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-079
  type: story                      # epic | story | task | bug | chore
  cycle: build                     # frame | design | build | verify | ship
  blocked: false
  priority: high
  complexity: M                    # S | M | L  (L means split it)

project:
  id: PROJ-007
  stage: STAGE-021
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-16

references:
  decisions:
    # The page asserts what each of these records says. Build must not
    # paraphrase them further than the literal already does; verify diffs the
    # literal, and these are the files a verify session would open to check it.
    - DEC-004                      # superseded record that was kept, not deleted
    - DEC-015                      # the record that superseded it
    - DEC-021                      # migration auto-backup — incident→hardening
    - DEC-025                      # carries the repo's one explicit `## Amendment`
    - DEC-026                      # dev-binary migration guard — incident→hardening
    - DEC-038                      # SQLite concurrency — incident→hardening
    - DEC-040                      # formula-not-cask — the security/distribution claim
    - DEC-043                      # the correction that was itself corrected twice more
    - DEC-044                      # the correction box that traces where a wrong number spread
  constraints:
    - test-before-implementation   # blocking, paths ["**"] — tests below are written at design
    - one-spec-per-pr              # blocking, paths ["**"]
    - no-secrets-in-code           # blocking, paths ["**"] — repo-wide, nothing here handles secrets
  related_specs:
    - SPEC-021                     # the README user-facing rewrite that CREATED test-docs and set A1's band
    - SPEC-073                     # the coverage sentence that pinned nothing, four times — the stage's cautionary case
    - SPEC-076                     # added W1–W4: mechanical forms of hand-ticked release-checklist items
    - SPEC-077                     # W3 — pinned the README version line after it rotted through two releases
    - SPEC-078                     # added W5–W6
---

# SPEC-079: engineering practices entry point

> **Cycle: design.** The go/no-go lives in this spec's git history (framed
> 2026-08-16, GO at complexity M). This revision settles the four forks framing
> left open, embeds the literal artifacts, and writes the failing tests. Per
> AGENTS.md §6, build happens in a **fresh session** and transcribes the
> literals under `## Notes for the Implementer` verbatim.

## Context

STAGE-021's premise is that bragfile's engineering discipline is **strong and
unreachable**. This spec is the stage's primary deliverable: the entry point
that makes it reachable.

The work is an **index, not an essay.** Every artifact a reviewer would find
persuasive already exists in the repository. What does not exist is any path to
it from the front door.

Parent: `STAGE-021-make-the-discipline-legible`, spec 1 of 2.
Project: `PROJ-007`.

## Goal

A first-time reader arriving at the README can reach — and verify — the
decision log, the cycle, the test regime, the security posture and the
incident→hardening pattern, **without knowing the repository's directory
layout.** Every claim on the page points at something countable, and every
count that tracks the repository's current state is **derived by a script and
diffed by a test**, so none of them can rot.

## Inputs

- **Files to read:** the literals under `## Notes for the Implementer` — that is
  the whole input. Build transcribes them.
- **Files to read for verification only:** `scripts/test-docs.sh` (the assertion
  style being extended), `README.md` (the insertion point), and each `DEC-*`
  listed in `references.decisions` (the page asserts what they say).
- **Related code paths:** none. This spec touches no Go source.

## Outputs

- **Files created:**
  - `docs/engineering-practices.md` — the practices page. **268 lines**,
    transcribed verbatim from literal ①.
  - `scripts/inventory.sh` — the single source for every current-state number on
    the page. Transcribed verbatim from literal ③. **Must be `chmod +x`.**
- **Files modified:**
  - `README.md` — one new `## How this repo is built` section inserted
    immediately before `## Where to go next`, from literal ②. Result: **260
    lines** (was 250).
  - `justfile` — one new `inventory:` recipe, from literal ④.
  - `scripts/test-docs.sh` — three changes, from literal ⑤:
    1. **A1's band re-pinned** `100 250` → `100 260` (see LD5).
    2. **E1's source list** gains `docs/engineering-practices.md`, so the new
       page's links are covered by the existing link-integrity assertion.
    3. **New Group X** (`X1`–`X8`) appended after the `W`-series, before
       `# ===== finalise =====`.
  - `guidance/questions.yaml` — one new question,
    `dec-amendment-heading-convention`, from literal ⑥ (see LD4).
- **New exports:** none.
- **Database changes:** none.

### Premise audit (§9), run at design against the repo

The additive case applies: this spec adds to two tracked collections whose
counts are asserted elsewhere. Both greps were **executed**, not merely
enumerated (§9 audit-grep cross-check):

| Grep | Hits | Disposition |
|---|---|---|
| `assert_line_count_band "A1"` in `scripts/test-docs.sh` | 1 (line 117) | **Planned modification.** README goes 250 → 260; A1's ceiling is exactly 250. Confirmed by running the harness against the staged README: `FAIL: A1: README.md has 260 lines (expected 100..250)`. |
| `grep -F 'docs/development.md' -r .` (assertion `E2`'s whitelist) | README.md, CONTRIBUTING.md, docs/development.md | **No change needed.** The new page links `development.md` *relative from `docs/`*, so the literal string `docs/development.md` never appears in it. Build must not "helpfully" rewrite that link to the absolute form — doing so breaks E2. |
| Existing assertions vs. the staged README + page | 1 failure | Full `./scripts/test-docs.sh` run with both artifacts staged: **A1 is the only failure.** Every B-series negative, E1, E2, and Groups C/D/P pass unchanged. |

## Locked design decisions

Framing left four forks open. All four are settled below, with rejected
alternatives listed per AGENTS.md §12 ("Decide at design time when decidable").

### LD1 — How counts are kept true: **derive them in one script, diff the whole table in one assertion**

`scripts/inventory.sh` prints the inventory table. The page embeds that output
verbatim between `<!-- inventory:begin -->` / `<!-- inventory:end -->` markers.
Assertion `X3` re-runs the script and diffs its output against the block; on
mismatch it prints both, and the remedy is `just inventory` + paste.

**The rule that follows from it, and which the page states in its own opening:**
every number describing the repository's **current state** lives inside the
guarded block. A number written into the prose is only ever a **dated historical
fact** (a version that shipped without a compare-link; a figure that was
corrected) — those do not rot. This is why the table carries sixteen rows rather
than the five framing anticipated: every current-state number the page wanted to
assert was pulled into it, including the confidence range and the
`## Amendment` count. (Fifteen at design; the sixteenth — the build-phase
reflection count — was added at the verify punch list, where P3 found a
current-state number that had been typed into the prose as a universal
quantifier. The rule caught its own violation.)

**Why this beats all three candidates framing listed:**

- **(a) one `test-docs` assertion per count — REJECTED.** N counts becomes N
  assertions *each with its own hand-typed expected value*. The expectation is
  then a second copy of the same rotting number, and the maintenance cost is
  paid on every change. `X3` has **zero** hand-maintained expected values: it
  computes the expectation.
- **(b) a `just` recipe, page points at the command — REJECTED as the whole
  answer, ADOPTED as half of it.** Framing was right that (b) is unrottable by
  construction, and right about the cost: "here is how to get them" is a weak
  page for a reader who will not clone the repository. The resolution is that
  (b) is not exclusive with keeping the numbers on the page — the recipe exists
  (`just inventory`), the page names it, *and* the numbers are printed. The
  recipe is also what makes `X3`'s failure remedy mechanical, which is what kills
  option (a)'s maintenance cost.
- **(c) generate the page's table — REJECTED.** A generator implies a build step
  and a second guard asserting the generated file is current; that second guard
  is `X3` doing the same work with more machinery. Paste-and-diff gets the same
  guarantee with no generator: the script is the source, the page is a cached
  copy, and the cache is validated on every run.

**Two sub-decisions inside LD1, both forced by measurement at design:**

1. **The doc-assertion count is DISTINCT ASSERTION IDS, derived statically —
   never a count of `OK:` lines from a run.** Framing recorded **164**. That
   number is wrong twice over. It double-counts: `S3` is one id used by two
   checks (`assert_file_exists` then `assert_cmd_ok claude plugin validate`), so
   it emits two `OK:` lines. Worse, it is **environment-dependent** — the second
   `S3` check is `skip`ped when the `claude` CLI is absent, so a stranger
   cloning the repository and running `just test-docs | grep -c '^OK:'` sees
   **163**, not 164, and would conclude the page is lying. Distinct ids (163
   today) is the only number reproducible off this machine. It is also the only
   one that can be computed **without running the harness**, which matters
   because `X3` runs `inventory.sh` — if `inventory.sh` ran `test-docs.sh`, the
   two would recurse.
2. **Every count is scoped to explicit tracked directories; never `find .`.**
   A background agent session creates a full second copy of the repository under
   `.claude/worktrees/`, and an unscoped walk silently doubles every number.
   This is the same trap already documented in `E2`'s comment.

### LD2 — README section vs. document: **a named `##` section that carries routing and no numbers**

The README gets a `## How this repo is built` section (literal ②) placed
immediately before `## Where to go next`. Everything else lives in
`docs/engineering-practices.md`.

**The split point is not length — it is guard coverage.** `X3` reaches the
page's inventory block and nothing else. Therefore:

- **The README section carries zero numbers.** Any count there would be a
  hand-typed count outside every guard — precisely the defect this spec exists
  to prevent, reintroduced in the most-read file in the repository.
- **The README section carries the one load-bearing sentence and the link.** Its
  job is to name the thing and route to it.
- **The page carries the evidence, the counts and the caveats.**

**Rejected alternatives (build-time):**

- **A sixth bullet in `## Where to go next` — REJECTED.** The goal is that a
  reader *reaches* the material without knowing the layout; the sixth entry of a
  pointer list is not reaching. It also would not have avoided the A1 problem
  (see LD5), so it buys nothing.
- **A `docs/` site or a multi-page split — REJECTED**, per STAGE-021's design
  note: a site is a distribution project, this is an index.
- **Putting the inventory table in the README — REJECTED.** It would move the
  guarded block into the file with the tightest shape constraints (A1's band,
  the B-series negatives) and make every future count change a README change.

### LD3 — Honest-close prominence: **its own named section, second-to-last, foreshadowed from the intro, by citation, in a fixed three-part shape**

`## Where this project was wrong, on the record` is a top-level section placed
after the evidence sections and before `## What this does not measure`. The
page's intro links to it directly ("The section that says the most about how
this repository is run is …").

**Why that position:** placed first, it reads as a disclaimer — the reader has
no frame yet in which "we record our errors" is a claim about discipline rather
than about quality. Placed last-but-one, after the reader has seen the guards
and the incident record, it reads as what it is. Foreshadowing it with a link
from the intro means a skimmer reading headings cannot miss it, so "not first"
does not become "buried."

**Why it is misread-proofed by shape, not by hedging:** every item states *what
was decided → how it was wrong → what changed as a result*, each with a path.
An error with no consequent change is an anecdote; an error with a recorded
repair is a process artifact. The section also states its own limit in its first
paragraph — that it is a list of named records and not a count — so a reader who
greps for a number does not catch the page overclaiming.

**Rejected alternatives (build-time):**

- **Leading with it — REJECTED**, per above.
- **Distributing the corrections into the sections they belong to (the DEC-044
  correction under "Decisions", the PROJ-006 close under "How a change gets
  made") — REJECTED.** It is the differentiated material; scattering it makes it
  invisible, which is the exact failure STAGE-021 exists to fix.
- **Softening it into "lessons learned" — REJECTED.** The persuasive version is
  the one that quotes the brief's own words ("Closed 2026-08-14 as a SCOPE
  REDUCTION, not a completion"), not a paraphrase that sands them off.

### LD4 — Counting rule for corrections: **NO. Do not create the convention. Claim by citation, and make the current state a derived row.**

The page does **not** define a `## Amendment` convention that future DECs must
follow. It names the specific records instead (`DEC-025`, `DEC-043`, `DEC-044`,
`DEC-004`→`DEC-015`, and PROJ-006's close), pinned by assertion `X6`.

**Why not:**

1. **It is a process change riding in on a docs spec.** Framing flagged this and
   was right to; the answer to "should this spec quietly bind every future
   decision record?" is no.
2. **Adoption would be 1 of 45 on day one.** A page advertising a convention with
   1/45 adoption makes a *weaker* claim than a page naming four specific
   records, and invites the reader to compute the ratio.
3. **The repo has a live case of exactly this failure mode.** SPEC-078 found the
   `pr:`/`commit:` evidence-link convention sitting at **zero** adoption because
   it was documented but nothing carried it into the surface that writes. A
   convention announced on a docs page, with no capture-time surface behind it,
   is that same shape.
4. **Retrofitting 45 records is out of scope**, and retrofitting some of them is
   worse than none: a partially-applied convention produces a count that is
   confidently wrong.

**What is done instead, so the deferral is not lost:** the inventory carries a
derived row — `…of those, carrying an explicit ## Amendment section | 1`. It
puts the honest number on the page *today*, and if the convention is ever
adopted the number rises with no edit to the page. The deferral itself is
recorded as a question with a stated resolve condition (literal ⑥), because an
open question with no resolve condition is the defect STAGE-021's other spec is
cleaning up.

**Rejected alternatives (build-time):**

- **Define `## Amendment` as the rule and state "1 of 45 today" — REJECTED**, per
  (2) and (3).
- **Define the rule and retrofit the obvious cases — REJECTED**, per (4); it
  also silently rewrites records, which is the behaviour this whole section is
  claiming the repo does not do.
- **Count by keyword grep — REJECTED and known-false.** A grep for
  correct/amend/supersede matches **all 45** records because ordinary decision
  prose uses those words. There is no number here, and framing was right that
  there is not.

### LD5 — A1's band: **re-pin tight to 260, do not widen, do not trade lines out**

`README.md` is exactly **250** lines and `assert_line_count_band "A1"
"README.md" 100 250` is exactly at its ceiling, so any net-positive addition
fails. A1's band becomes `100 260`, which is the exact line count of the README
after literal ② is inserted — **zero headroom.**

**Why re-pin rather than widen.** A guard that is widened whenever it fires is
not a guard. Re-pinning to the exact new value grants *no* additional room: the
next line added to the README fails A1 again and is again a decision. The
guard's strength strictly increases relative to leaving slack. This is the same
move SPEC-077 made for the version line — replace a claim that drifts with one
pinned to a derived value — applied to a shape constraint.

**Why the README is allowed to grow at all.** A1 was set at SPEC-021 to keep the
README a user-facing quickstart rather than a manual. A ten-line routing section
does not change that; the section adds no commands, no flags and no numbers.
The band's *purpose* is intact, only its *number* moves, and it moves by exactly
the size of the change that justified it.

**The cost, stated honestly:** at zero headroom, any future README edit that adds
a line — including a reflow — fails A1 and must re-pin. That friction is the
point, in a repository whose recorded failure mode is claims drifting silently,
but it is a real cost and a reviewer may reasonably prefer a band.

**Rejected alternatives (build-time):**

- **Widen to a round number with headroom (e.g. `100 300`) — REJECTED.**
  Headroom is permission to sprawl unobserved, and re-earns the exact defect A1
  exists to catch. It is also the "bump the number until the test passes" move.
- **Trade lines out of the README to stay net-zero — REJECTED.** The only
  compressible block of the required size is `### Upgrading from 0.5.1 or
  earlier` (22 lines), which documents a **silent** upgrade failure — `brew
  outdated` stays quiet while a user sits on a stale version. Cutting
  user-protecting content to make room for a link about engineering quality is a
  bad trade and an ironic one.
- **Add nothing net-new (fold the link into an existing line) — REJECTED**, per
  LD2's first rejected alternative; and contorting the README to avoid tripping
  a guard is itself the "make the tests pass" move in a quieter form.

### LD6 — Two known-false claims are not restated; both re-verified at design

Framing flagged these; both were re-verified against the repository in this
session before the page was written.

1. **"Golden files" — there are none.** `find . -name '*.golden'` returns
   nothing and there are no `testdata/` directories (both re-run 2026-08-16, both
   zero). The mechanism is byte-exact expected output **embedded as Go string
   literals**, compared with `!=`. The page says exactly that, and `X4` pins the
   phrase. Note the *word* "golden" does appear in 13 test files as a naming
   convention for those literals — the false claim is "golden **files**", not
   the term.
2. **"Decision records that log their own corrections" is not countable.**
   Re-verified: exactly **1** record (`DEC-025`) has a `^## Amendment` heading;
   a keyword grep matches all 45. Handled by LD4.

**`X4` is deliberately a POSITIVE assertion, and this is the §12 NOT-contains
self-audit result.** The natural form is `assert_not_contains_iregex "X4"
'golden files?'` — and the self-audit kills it: the page's honest sentence *is*
"There are no golden files", so a NOT-contains would forbid the exact sentence
the acceptance criterion requires. Pinning the true mechanism instead is the
form that survives.

## §12(b) design-time verification

Every literal below was run through its target tool during design. Nothing here
is a prediction.

| # | What was pre-flighted | Result |
|---|---|---|
| 1 | `scripts/inventory.sh` executed against the repo | Runs clean; output is literal ③'s `cat` block, values as embedded. |
| 2 | The page's inventory block diffed against `inventory.sh` output | **Exactly one differing line**, the doc-assertion count (page `171`, script `163` — see #3). Every other row byte-identical. |
| 3 | Post-build doc-assertion count | Group X (literal ⑤) concatenated onto `test-docs.sh` and the id-extraction re-run: **171**, and `comm` confirms the delta is exactly `X1 X2 X3 X4 X5 X6 X7 X8` — no accidental ids picked up from `X3`'s multi-line failure message. This is why the embedded block says 171 and not 163. |
| 4 | Static id extraction vs. a real run | `diff` of the statically-extracted id set against the ids in a live `./scripts/test-docs.sh` run: **identical sets** (163 each). The only divergence is multiplicity — `S3` emits twice — which is the finding behind LD1's sub-decision 1. |
| 5 | README line count after literal ② | **260**, measured on an assembled file, not counted by hand. That is A1's new ceiling. |
| 6 | Page line count | **268**, inside `X7`'s `150 300` band. |
| 7 | `X5`'s forbidden-adjective regex against the page literal | **Zero hits.** |
| 8 | `X4`'s literal and `X6`'s ids against the page literal | All present. |
| 9 | E1-equivalent link resolution over every `](…)` in the page | **0 broken links** (24 targets, resolved relative to `docs/`). |
| 10 | `E2`'s grep (`docs/development.md`) against the page literal | **No hit** — the page uses the relative form. E2 needs no whitelist change. |
| 11 | B-series negatives against the assembled README | **No hits** for `spec-driven`, the three cycle-phrase forms, `four habits`, `context contamination`, `claude plays every role`, or the forbidden `just` recipes. |
| 12 | Full `./scripts/test-docs.sh` with both artifacts staged | **A1 is the only failure.** Confirms the Outputs list is complete: nothing else in the harness is disturbed. |

## Acceptance Criteria

- [ ] `docs/engineering-practices.md` exists and matches literal ① byte-for-byte
      (`X1`, and the verify diff).
- [ ] `README.md` contains the section from literal ② and is 260 lines; the
      README links to the page (`X2`).
- [ ] `scripts/inventory.sh` exists, is executable, and its output equals the
      page's inventory block (`X3`).
- [ ] `just inventory` runs it (`X8`).
- [ ] The page describes the test regime as embedded Go string literals, not as
      golden files (`X4`).
- [ ] No bare adjective stands alone on the page (`X5`).
- [ ] The corrections claim is made by citation, naming each record (`X6`).
- [ ] The page stays an index: 150–300 lines (`X7`).
- [ ] Every internal link on the page resolves (`E1`, extended).
- [ ] `just test-docs` passes with 171 distinct assertions; `just test`,
      `gofmt -l .` and `go vet ./...` are unaffected.

## Failing Tests

Written during **design**, BEFORE build. Every locked decision above has at
least one test that would fail without it (AGENTS.md §9).

All assertions live in **`scripts/test-docs.sh`**; the literal source is ⑤.

- **`scripts/test-docs.sh` — Group X (new)**
  - `"X1"` — asserts `docs/engineering-practices.md` exists. *(Fails today: no
    such file.)*
  - `"X2"` — asserts `README.md` contains the literal
    `docs/engineering-practices.md`. Pins **LD2**: without the README section
    there is no entry point, only a document. *(Fails today.)*
  - `"X3"` — **the counting guard.** Asserts the block between
    `<!-- inventory:begin` and `<!-- inventory:end` in the page equals
    `./scripts/inventory.sh` output exactly. Pins **LD1**. Fails if the script
    is missing, not executable, the markers are absent, the block is empty, or
    any value drifts; the failure message prints both sides. *(Fails today: no
    script, no page.)*
  - `"X4"` — asserts the page contains the literal `embedded as Go string
    literals`. Pins **LD6.1** — a page that reverted to "golden files" would go
    red. Positive form by self-audit, see LD6. *(Fails today.)*
  - `"X5"` — asserts the page does NOT match
    `rigorous|comprehensive|world-class|best-in-class|battle-tested|cutting-edge|state-of-the-art`.
    Mechanises STAGE-021's "no adjective stands alone". *(Fails today: no
    page — and see the self-audit note below.)*
  - `"X6"` — asserts the page contains each of `DEC-004`, `DEC-015`, `DEC-025`,
    `DEC-043`, `DEC-044`, `PROJ-006`. Pins **LD4**: the corrections claim is
    made by citation, and stays that way. *(Fails today.)*
  - `"X7"` — asserts the page is 150–300 lines. Pins the spec's stated failure
    mode (an index that grows into an essay). *(Fails today.)*
  - `"X8"` — asserts the `justfile` has an `^inventory:` recipe, so the remedy
    `X3` names actually exists. Same shape as the existing `F1`. *(Fails
    today.)*

- **`scripts/test-docs.sh` — existing assertions modified**
  - `"A1"` — band re-pinned `100 250` → `100 260`. Pins **LD5**. Without the
    re-pin this assertion fails at 260 lines; with a *wider* band the decision
    would be unrecorded. Verified at design: A1 is the only existing assertion
    the change disturbs.
  - `"E1"` — source list gains `docs/engineering-practices.md`, so every
    `](…)` on the page must resolve. Reuses the proven mechanism rather than
    adding a ninth id.

### §12 NOT-contains self-audit for `X5`

`X5` is the spec's only NOT-contains assertion, so the rule applies: grep the
**load-bearing** text — the page literal ① and the README literal ② — for each
forbidden token.

| Token | Hits in ① | Hits in ② |
|---|---|---|
| `rigorous` | 0 | 0 |
| `comprehensive` | 0 | 0 |
| `world-class` | 0 | 0 |
| `best-in-class` | 0 | 0 |
| `battle-tested` | 0 | 0 |
| `cutting-edge` | 0 | 0 |
| `state-of-the-art` | 0 | 0 |

Run at design (§12(b) row 7): **zero hits.** The tokens do appear in *this
spec's* commentary above, which is correct and harmless — commentary does not
reach the artifact, and `X5`'s scope is the page only.

## Implementation Context

*Read this section (and the files it points to) before starting the build
cycle. It is the equivalent of a handoff document, folded into the spec since
there is no separate receiving agent.*

### Decisions that apply

None of these constrains the *implementation*; each is a record whose **content
the page asserts**, so build must not paraphrase beyond the literal and verify
should open them to check the literal is fair:

- `DEC-004` / `DEC-015` — the one superseded record and its successor; the page
  claims DEC-004 was kept, carrying `superseded_by: DEC-015`.
- `DEC-021`, `DEC-026`, `DEC-038` — the three incident→hardening records.
- `DEC-040` — formula-not-cask; the page claims this removes the code-signing
  and Gatekeeper path.
- `DEC-025` — the one record with an explicit `## Amendment` heading.
- `DEC-043`, `DEC-044` — the correction chain the honest-close section cites,
  including that DEC-043 now carries a **third** correction dated 2026-08-10.

### Constraints that apply

These constraints apply to the paths touched by this task (see
`/guidance/constraints.yaml` for full text):

- `test-before-implementation` — blocking, `["**"]`. The Failing Tests section
  above is written; build makes it pass and does not add assertions beyond
  literal ⑤ without saying so under Deviations.
- `one-spec-per-pr` — blocking, `["**"]`. One PR, referencing SPEC-079.
- `no-secrets-in-code` — blocking, `["**"]`. Nothing here handles credentials;
  listed because it is repo-wide.

No Go source is touched, so the Go-scoped constraints (`no-cgo`,
`no-sql-in-cli-layer`, `storage-tests-use-tempdir`, `timestamps-in-utc-rfc3339`,
`migrations-are-append-only`, `errors-wrap-with-context`,
`stdout-is-for-data-stderr-is-for-humans`) do not apply.

### Prior related work

- `SPEC-021` (shipped) — created `scripts/test-docs.sh` and set A1's `100 250`
  band during the README user-facing rewrite. LD5 modifies that band; read
  SPEC-021 before deciding the modification is casual.
- `SPEC-076` (shipped) — added `W1`–`W4`, the mechanical forms of release-
  checklist items that had been ticked by hand and missed.
- `SPEC-077` (shipped) — `W3`; pinned the README version line after it rotted
  through two releases. LD1 is the same move for a table of counts.
- `SPEC-078` (shipped) — added `W5`–`W6`, and is LD4's cautionary case: a
  documented convention at zero adoption because nothing carried it into the
  writing surface.
- `SPEC-073` (shipped) — the stage's cautionary case for coverage claims.

### Out of scope (for this spec specifically)

- The **godoc pass**, the `DEC-041` tombstone, and open-questions **hygiene** —
  STAGE-021's second spec. (Adding one new, well-formed question with a resolve
  condition, per LD4, is not hygiene work on the existing ones.)
- Lint, coverage, badges, CI — STAGE-022. The page **names** these as gaps under
  `## What this does not measure`; it does not close them.
- Rewriting the tutorial or the blog posts. This spec **links** them.
- Any change to what the binary does. No Go file is touched.
- Fixing the brief's three stale counts, or STAGE-021's own stale ones. See
  "Findings" below — recorded, not actioned.
- **Fixing `S3`'s double-emit** in `test-docs.sh`. It is a real wart (LD1 sub-
  decision 1) but it is an existing assertion's behaviour and changing it is not
  this spec's business. The design routes around it instead.

### Findings recorded at design, deliberately not actioned

Each survives contact with the code and is stated here so it is not re-derived:

1. **"Four W-series guards" is stale.** STAGE-021 and the framing spec both say
   `W1–W4`. There are **six** (`W5`/`W6` landed with SPEC-078). The page counts
   them with a derived row rather than repeating the number.
2. **"164 `test-docs` assertions" is wrong and unstable.** See LD1 sub-decision
   1: it double-counts `S3` and is environment-dependent. The true, reproducible
   figure today is 163 distinct ids.
3. **The brief's "zero deprecated" is now false.** `DEC-004` carries
   `superseded_by: DEC-015`. This is an *improvement* to the story, not damage —
   a log that records supersession is stronger than one with nothing to record —
   and the page presents it that way.
4. **STAGE-021 says "8 of 18 [questions] are open"; it is 8 of 17.** Belongs to
   the second spec. *(Corrected 2026-08-16: the design pass first recorded this
   as "9 of 18". Both halves were wrong, from a `grep -c 'status: open'` that
   matched the file's own header comment — `#   - status: open | investigating |
   answered` — inflating the open count, and a total that counted the same
   comment. Filtering comment lines gives **8 open of 17**. A demonstration of
   this spec's thesis at its own expense: the miscount was produced by exactly
   the hand-run-grep method the inventory block exists to replace, while
   auditing a file for stale counts.)*
5. **"Research syntheses — 4" over-describes the directory.** Two of the four
   `docs/research/` files are research *prompts*, not syntheses. The page
   describes them by what they are.
6. **`docs/reports/lifetime/` does not exist** — `just lifetime-save` creates it
   on first run. An early draft of the page linked it; the link check caught it.

## Notes for the Implementer

**Transcribe the six literals below verbatim.** Build should produce zero
prose of its own. Verify diffs the working tree against these literals.

Two traps worth naming before you start:

- The page links `development.md` **relative from `docs/`**. Do not "correct" it
  to `docs/development.md` — assertion `E2` asserts that literal string appears
  only in `README.md`, `CONTRIBUTING.md` and `docs/development.md`, and the
  rewrite would break it.
- `scripts/inventory.sh` must be `chmod +x`. `X3` fails explicitly if it is not.

---

### ① `docs/engineering-practices.md` (new file, 275 lines)

```markdown
# Engineering practices

An index to the engineering artifacts already in this repository — what exists,
where it is, what it actually guarantees, and what it does not.

Two rules govern this page, because a page like this is normally where a
project's claims go to rot:

1. **Every number describing the repository's current state is derived, not
   typed.** They all live in the one table below, printed by
   [`../scripts/inventory.sh`](../scripts/inventory.sh) (`just inventory`) and
   re-derived and diffed on every `just test-docs` run by assertion `X3` in
   [`../scripts/test-docs.sh`](../scripts/test-docs.sh). If a number on this
   page stops matching the repository, the documentation tests fail. A number
   written into the prose below is only ever a **dated historical fact** — a
   version that shipped without a compare-link, a figure that was corrected —
   and those do not rot.
2. **Every claim points at a path, a count, or a named test.** If a sentence
   here cannot be checked by opening something, it does not belong on the page.

The section that says the most about how this repository is run is
[Where this project was wrong, on the record](#where-this-project-was-wrong-on-the-record).

## The inventory

<!-- inventory:begin — generated by scripts/inventory.sh (`just inventory`); pinned by test-docs X3; do not hand-edit -->
| What | Value | Where it lives |
|---|---:|---|
| Decision records | 45 | `decisions/DEC-*.md` |
| …of those, superseded by a later record | 1 | `superseded_by:` in the front-matter |
| …of those, carrying an explicit `## Amendment` section | 1 | `decisions/DEC-*.md` |
| Lowest confidence value on a decision record | 0.65 | `insight.confidence` in the front-matter |
| Highest confidence value on a decision record | 0.95 | `insight.confidence` in the front-matter |
| Decision records claiming confidence 1.0 | 0 | `insight.confidence` in the front-matter |
| Projects | 7 | `projects/PROJ-*/brief.md` |
| Stages | 21 | `projects/*/stages/STAGE-*.md` |
| Specs carried to ship and archived | 75 | `projects/*/specs/done/` |
| …of those, also carrying a build-phase reflection | 69 | `### Build-phase reflection` in those files |
| Go source files | 69 | `internal/`, `cmd/` |
| Go test files | 78 | `internal/`, `cmd/` |
| Go test functions | 812 | `func Test*` in `*_test.go` |
| Documentation assertions (distinct ids) | 171 | `scripts/test-docs.sh`, run by `just test-docs` |
| …of those, replacing a manual release-checklist item | 6 | the `W`-series in `scripts/test-docs.sh` |
| Benchmarks | 0 | none exist — see "What this does not measure" |
<!-- inventory:end -->

Reproduce the whole table with `just inventory`. Each value is scoped to
explicit tracked directories rather than a repo-wide walk, because a background
agent session creates a full second copy of the repository under
`.claude/worktrees/`, and an unscoped walk silently doubles every number — the
same trap already documented at assertion `E2`.

## Decisions

[`../decisions/`](../decisions/) holds one file per non-obvious choice. Ids are
repo-global and never reused, and `DEC-041` is reserved rather than lost — the
reservation is recorded at
[`../projects/PROJ-001-mvp/backlog.md`](../projects/PROJ-001-mvp/backlog.md).

- **Each record carries an honest confidence value.** `insight.confidence` is
  part of every record's front-matter, and no record claims certainty — see the
  lowest, highest and 1.0 rows above. The rule that a decision below 0.7 must
  also open a question in
  [`../guidance/questions.yaml`](../guidance/questions.yaml) is stated in
  [`../AGENTS.md`](../AGENTS.md) §14.
- **Superseded records stay.** `DEC-004` (tags as a comma-joined string) was
  superseded by [`DEC-015`](../decisions/DEC-015-polymorphic-tags-normalization.md)
  when tags were normalised into a `tags` + `taggings` model.
  [`DEC-004`](../decisions/DEC-004-tags-comma-joined-for-mvp.md) is still in the
  directory, carrying `superseded_by: DEC-015`, because the reasoning that was
  later outgrown is the part worth keeping.
- **Corrections are written into the record they correct**, rather than the
  record being quietly rewritten. The specific ones are listed under
  [Where this project was wrong](#where-this-project-was-wrong-on-the-record).

## How a change gets made

The cycle, the work hierarchy and the session rules are documented in
[`development.md`](development.md); [`../AGENTS.md`](../AGENTS.md) is the full
conventions document and [`../CONTRIBUTING.md`](../CONTRIBUTING.md) is the short
version.

Two properties are worth checking directly, because they are what the archived
specs demonstrate:

- **Tests are written before the implementation, into the spec itself.** Each
  spec carries a `## Failing Tests` section written during design; build makes
  them pass. `test-before-implementation` is a blocking constraint in
  [`../guidance/constraints.yaml`](../guidance/constraints.yaml).
- **Every archived spec carries its own reflection.** Every file under
  `projects/*/specs/done/` ends with a ship-phase `## Reflection (Ship)`. The
  build-phase reflection that precedes it — what was unclear, what was missing,
  what would be done differently — is carried by the subset counted in the
  inventory above, not by all of them. `scripts/archive-spec.sh` refuses to
  archive a spec whose reflection still holds template placeholders.

## What the tests actually pin

- **There are no golden files.** `find . -name '*.golden'` returns nothing and
  there are no `testdata/` directories. Byte-exact expected output is instead
  **embedded as Go string literals** in the test file next to the code, and
  compared with `!=` against real output — for example
  `TestToMemoryJSON_BlendedGolden` in
  [`../internal/export/memory_test.go`](../internal/export/memory_test.go).
- **What a golden does *not* pin is written down, established by mutation.**
  [`../internal/cli/memory_test.go`](../internal/cli/memory_test.go) carries a
  comment above `TestMemoryCmd_EndToEndMarkdownGolden` naming three things that
  golden does not pin — the pool composition, the fusion constants, and the
  declared `Matched` order — each verified by mutating the implementation and
  observing the golden stay green, and each paired with the test that does pin
  it. The rule behind it is that a claim about what a test pins is aspirational
  until it has been mutation-checked.
- **Two implementations of the same rule are pinned to each other.**
  `TestProvenanceClassifier_GoPredicateMatchesSQLClause` in
  [`../internal/storage/provenance_agreement_test.go`](../internal/storage/provenance_agreement_test.go)
  fails if the Go predicate that classifies an entry as agent-authored ever
  disagrees with the SQL clause that does the same job.
- **Some invariants are asserted over the package source, not its behaviour.**
  `TestPackageReadsNoWallClock`
  ([`../internal/memory/memory_test.go`](../internal/memory/memory_test.go),
  [`../internal/timewindow/timewindow_test.go`](../internal/timewindow/timewindow_test.go))
  fails if a package that must be time-invariant ever reads the clock;
  `TestPackageEmitsNoReservedTagNamespace` fails if the memory package ever
  writes a reserved tag namespace.
- **Documentation has its own test suite.** `just test-docs` runs the assertions
  counted above, over the README, `CONTRIBUTING.md`, the tutorial, the agent
  docs, the JSON schema, the goreleaser config, the CI workflows and the
  `CHANGELOG`.

## Guards that replaced remembered checks

The `W`-series in [`../scripts/test-docs.sh`](../scripts/test-docs.sh) is the
clearest single artifact on this page: each of these assertions replaced an item
a human used to tick by hand on a release checklist. The comment above each one
states what it would have caught.

- `W1` — the plugin version pin must equal the newest dated `CHANGELOG` section.
  At the v0.5.2 cut the pin sat on 0.5.1, which had a perfectly good dated
  section, so the weaker "the pin has *a* dated section" form would have passed
  and caught nothing.
- `W2` — every dated `CHANGELOG` version heading must have a compare-link.
  v0.5.2 shipped without one.
- `W3` — the README status line and the tutorial's "shipped as of" line must
  name the newest release. Both were wrong through two consecutive releases
  (v0.5.2 and v0.6.0) before this assertion existed, because the release
  pre-flight checked the `CHANGELOG`, the plugin pin and the compare-links, but
  never the prose a reader sees first.
- `W4` — a stage marked `status: shipped` must not still hold reflection
  placeholders. `archive-spec.sh` had this guard for specs; stages had none, so
  one was marked shipped with its entire reflection on template text.
- `W5`, `W6` — the MCP tool description and the agent-facing docs must name the
  evidence-link convention, in the right order.

## Security

- [`../SECURITY.md`](../SECURITY.md) is the policy: supported versions, how to
  report, and what the threat model is for a local-first CLI.
- [`reports/security/2026-04-26-pre-distribution-security-review.md`](reports/security/2026-04-26-pre-distribution-security-review.md)
  is a full review run **before** the first binary was published, triggered by
  pausing the Homebrew deploy to do it. It records the state at the time: no
  tags pushed, no releases published, no users.
- The binary makes no network calls. `SECURITY.md` states plainly that this is
  "enforced by review rather than assumed" — there is no test asserting it.
- [`DEC-040`](../decisions/DEC-040-distribution-binary-formula-over-cask.md)
  records shipping as a Homebrew formula rather than a cask, which removes the
  code-signing and Gatekeeper path entirely.

## When something reached users anyway

Defects that reached released builds or installed users are recorded together
with the hardening they produced. Every one of them was operational or runtime
rather than a logic error — "every prod escape is operational/runtime" is the
headline finding of
[`reports/cross-project/2026-07-04-three-project-retrospective.md`](reports/cross-project/2026-07-04-three-project-retrospective.md),
and the last item below postdates that report and confirms it again.

- A migration could damage an existing database →
  [`DEC-021`](../decisions/DEC-021-migration-auto-backup-durability-model.md)
  added an automatic pre-migration backup with a stated durability model.
- A development build could migrate a real database →
  [`DEC-026`](../decisions/DEC-026-dev-prod-migration-guard.md) makes an
  unreleased binary refuse.
- Concurrent access failed on the MCP path →
  [`DEC-038`](../decisions/DEC-038-sqlite-concurrency-busy-timeout-single-conn.md)
  set the busy timeout, transaction mode and connection policy.
- Moving distribution from a cask to a formula froze every earlier install at
  0.5.1 with **no signal on either side** — `brew outdated` stayed silent. The
  one-time migration is now in the [`../README.md`](../README.md) install
  section and the [`../CHANGELOG.md`](../CHANGELOG.md).

The distribution lessons — including a goreleaser tagging failure and a Homebrew
tap-trust gate that appeared between two releases with no change on this side —
are written up in [`../AGENTS.md`](../AGENTS.md) §4, alongside the rule they
produced: any change to the distribution mechanism is a decision requiring a
downsides pass, regardless of how small the diff is.

## Where this project was wrong, on the record

Each item below is a place the repository records its own error rather than
overwriting it. This is a list of **named records, not a count**: as the
inventory shows, only one decision record carries an explicit `## Amendment`
section, so there is no counting rule to quote, and a keyword search would match
every record because ordinary prose uses those words.

- **A project closed as a scope reduction, not a completion.** PROJ-006 shipped
  two releases and still delivered one of its three stated pillars entire, one
  partly, and one never framed. Its close says so in those terms —
  [`../projects/PROJ-006-agent-native-depth-core/brief.md`](../projects/PROJ-006-agent-native-depth-core/brief.md),
  Project-Level Reflection: "Closed 2026-08-14 as a SCOPE REDUCTION, not a
  completion. Say this plainly, because the numbers below do not otherwise read
  that way."
- **A measured number was wrong, and the record traces where it had spread.**
  [`DEC-044`](../decisions/DEC-044-memory-slice-token-budget-and-line-shape.md)
  carries a correction box: the default memory budget was documented as yielding
  ≈110 entries; re-measured against the real corpus it yields 25. The box also
  names the two other files the wrong figure had already been copied into, and
  why the audit sweep that should have caught it did not.
- **The first repair of that error was itself wrong on a second axis, and the
  second repair is also recorded.**
  [`DEC-043`](../decisions/DEC-043-memory-slice-blended-rank-fusion.md) now
  carries a *third* correction to the same paragraph, dated 2026-08-10, which
  re-measured the headroom figure and then replaced the figure-dependent
  argument with a bound that survives the corpus growing — the property the
  first two versions of the paragraph both lacked.
- **A decision amended after its build punch-list kept the original.**
  [`DEC-025`](../decisions/DEC-025-claude-code-plugin-packaging-and-capture-nudge.md)
  carries an explicit `## Amendment (2026-07-04, SPEC-041 build punch-list)`.
- **A superseded decision was not deleted.** `DEC-004` → `DEC-015`, described
  above.

## What this does not measure

Stated because the page is otherwise an argument that things are checked, and
these are not.

- **No benchmarks.** There is no performance or scale baseline; `func Benchmark`
  appears nowhere in the tree, as the inventory shows.
- **No lint gate and no coverage number.** CI
  ([`../.github/workflows/ci.yml`](../.github/workflows/ci.yml)) runs `gofmt -l
  .`, `go vet ./...` and `go test ./...`. There is no `golangci-lint`, no
  `-cover`, and no badge.
- **The documentation assertions are not run by CI.** `just test-docs` is a
  local command; nothing enforces it on a pull request.
- **The no-network claim is enforced by review**, not by a test — as
  `SECURITY.md` itself says.
- **Known defects are recorded and deliberately left unfixed**, each because it
  needs a decision rather than a patch. They are named in
  [`../projects/PROJ-007-quality-and-portfolio-readiness/brief.md`](../projects/PROJ-007-quality-and-portfolio-readiness/brief.md).

[`STAGE-022`](../projects/PROJ-007-quality-and-portfolio-readiness/stages/STAGE-022-measured-and-enforced.md)
closes the lint gate and the coverage number, and one of the known defects — the
`Entries:` envelope inconsistency. It does not close the rest: its *Explicitly
out of scope* section defers benchmarks to
[`PROJ-009`](../projects/PROJ-009-scale-baseline-and-harness/brief.md), running
the documentation assertions in CI is owned by nothing today, and the no-network
claim stays enforced by review.

## The rest, by path

- [`reports/cross-project/`](reports/cross-project/) — a read-only retrospective
  across three repositories, run 2026-07-04, plus the action register it
  produced and the CSV/JSON data it was derived from.
- [`framework-feedback/`](framework-feedback/) — documents written *back* to the
  process framework this repository uses: after 8 specs, after 3 projects, on
  scaling it, and one proposing a new recipe.
- [`research/`](research/) — a time-boxed DuckDB federation spike with a
  verdict, an idea synthesis, and two reusable research prompts.
- [`blog/`](blog/) — why the tool exists, and how it was built.
- [`reports/daily/`](reports/daily/) — generated status snapshots
  (`just daily-status-report`); derived, not authored. `just lifetime-save`
  writes the whole-repo history report into a `reports/lifetime/` directory it
  creates on first run.
- [`architecture.md`](architecture.md), [`data-model.md`](data-model.md) and
  [`api-contract.md`](api-contract.md) — the design docs.
- [`tutorial.md`](tutorial.md) — the user-facing walkthrough.
```

---

### ② `README.md` — new section

Insert **immediately before** the existing `## Where to go next` heading (after
the blank line that follows the `docs/for-ai-agents.md` paragraph). Result: 260
lines.

```markdown
## How this repo is built

Every change ships through a written specification whose tests are written
before the implementation, and every non-obvious choice becomes a numbered
decision record that stays in the repo when it later turns out to be wrong.
What exists, what the tests actually pin, what is deliberately not measured,
and where this project got something wrong and corrected it are indexed in
[`docs/engineering-practices.md`](docs/engineering-practices.md) — where every
number is printed by `just inventory` and pinned by `just test-docs`.
```

> **Do not add a number to this section.** It sits outside `X3`'s reach; a count
> here is a hand-typed count with nothing holding it. See LD2.

---

### ③ `scripts/inventory.sh` (new file, `chmod +x`)

```bash
#!/usr/bin/env bash
# scripts/inventory.sh — print the engineering-practices inventory table.
#
# SINGLE SOURCE for two consumers:
#   1. the table embedded in docs/engineering-practices.md (paste this output
#      between the `inventory:begin` / `inventory:end` markers), and
#   2. test-docs assertion X3, which re-runs this script and diffs its output
#      against that block — so a number on the page cannot drift from the repo.
#
# Run via `just inventory`.
#
# THE RULE THE PAGE FOLLOWS: every number that tracks the repo's CURRENT state
# is produced here. A number stated inline on the page is only ever a dated
# historical fact (a version that shipped without a compare-link, a figure that
# was corrected) — those do not rot. If you find yourself wanting to type a
# current-state number into the prose, add a row here instead.
#
# TWO RULES THIS SCRIPT ITSELF HOLDS:
#
#   (a) Every count is scoped to explicit tracked directories. NEVER `find .`:
#       when a background agent is running, .claude/worktrees/<name>/ holds a
#       full second copy of the repo, and an unscoped walk silently doubles
#       every number. Same trap already documented at test-docs E2.
#
#   (b) The doc-assertion count is DISTINCT ASSERTION IDS, derived statically
#       from the script text — not a count of `OK:` lines from a run. Two
#       reasons: running test-docs.sh from here would recurse (X3 runs this
#       script), and the OK-line count is environment-dependent — S3 emits an
#       OK per check when the `claude` CLI is installed and OK+SKIP when it is
#       not, so `just test-docs | grep -c '^OK:'` prints one more on a machine
#       with the Claude CLI than without. Distinct ids is the only number a
#       stranger cloning the repo reproduces.

set -eu

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
cd "$REPO_ROOT"

n() { printf '%s' "$1" | tr -d ' '; }

# Confidence values, one per line, numeric, trailing comments stripped.
confidences() {
    grep -h '^  confidence:' decisions/DEC-*.md | sed 's/#.*//' | awk '{print $2+0}'
}

decs=$(n "$(ls decisions/DEC-*.md 2>/dev/null | wc -l)")
decs_superseded=$(n "$(grep -l '^superseded_by: DEC-' decisions/DEC-*.md 2>/dev/null | wc -l)")
decs_amended=$(n "$(grep -l '^## Amendment' decisions/DEC-*.md 2>/dev/null | wc -l)")
conf_min=$(confidences | sort -g | head -1)
conf_max=$(confidences | sort -g | tail -1)
conf_certain=$(n "$(confidences | awk '$1 >= 1.0' | wc -l)")
projects=$(n "$(ls -d projects/PROJ-*/ 2>/dev/null | wc -l)")
stages=$(n "$(ls projects/*/stages/STAGE-*.md 2>/dev/null | wc -l)")
specs_done=$(n "$(ls projects/*/specs/done/SPEC-*.md 2>/dev/null | wc -l)")
specs_build_reflection=$(n "$(grep -l '^### Build-phase reflection' projects/*/specs/done/SPEC-*.md 2>/dev/null | wc -l)")
src=$(n "$(find internal cmd -name '*.go' ! -name '*_test.go' | wc -l)")
tst=$(n "$(find internal cmd -name '*_test.go' | wc -l)")
testfuncs=$(n "$(grep -rh '^func Test' internal cmd --include='*_test.go' | wc -l)")
benchmarks=$(n "$(grep -rh '^func Benchmark' internal cmd --include='*_test.go' | wc -l)")
wseries=$(n "$(grep -cE '^# W[0-9]+ ' scripts/test-docs.sh)")
docasserts=$(n "$(grep -oE '(^|[[:space:]])(ok|fail|skip|assert_[a-z_]+) "[A-Za-z0-9][A-Za-z0-9._-]*"' \
    scripts/test-docs.sh | grep -oE '"[A-Za-z0-9][A-Za-z0-9._-]*"' | tr -d '"' | sort -u | wc -l)")

cat <<EOF
| What | Value | Where it lives |
|---|---:|---|
| Decision records | ${decs} | \`decisions/DEC-*.md\` |
| …of those, superseded by a later record | ${decs_superseded} | \`superseded_by:\` in the front-matter |
| …of those, carrying an explicit \`## Amendment\` section | ${decs_amended} | \`decisions/DEC-*.md\` |
| Lowest confidence value on a decision record | ${conf_min} | \`insight.confidence\` in the front-matter |
| Highest confidence value on a decision record | ${conf_max} | \`insight.confidence\` in the front-matter |
| Decision records claiming confidence 1.0 | ${conf_certain} | \`insight.confidence\` in the front-matter |
| Projects | ${projects} | \`projects/PROJ-*/brief.md\` |
| Stages | ${stages} | \`projects/*/stages/STAGE-*.md\` |
| Specs carried to ship and archived | ${specs_done} | \`projects/*/specs/done/\` |
| …of those, also carrying a build-phase reflection | ${specs_build_reflection} | \`### Build-phase reflection\` in those files |
| Go source files | ${src} | \`internal/\`, \`cmd/\` |
| Go test files | ${tst} | \`internal/\`, \`cmd/\` |
| Go test functions | ${testfuncs} | \`func Test*\` in \`*_test.go\` |
| Documentation assertions (distinct ids) | ${docasserts} | \`scripts/test-docs.sh\`, run by \`just test-docs\` |
| …of those, replacing a manual release-checklist item | ${wseries} | the \`W\`-series in \`scripts/test-docs.sh\` |
| Benchmarks | ${benchmarks} | none exist — see "What this does not measure" |
EOF
```

---

### ④ `justfile` — new recipe

Append to the **HELPERS** section at the end of the file (after `info:`). Do not
place it adjacent to `test:` — assertion `F2` reads the `test:` recipe from its
header to the next blank line and asserts its exact two-line form.

```make
# Print the engineering-practices inventory table (paste into docs/engineering-practices.md)
inventory:
    @./scripts/inventory.sh
```

---

### ⑤ `scripts/test-docs.sh` — three changes

**(i) A1's band.** Replace the existing line (currently line 117):

```bash
# A1 — README line count band 100..250
assert_line_count_band "A1" "README.md" 100 250
```

with:

```bash
# A1 — README line count band 100..260
# Re-pinned 250 -> 260 at SPEC-079 (LD5), TIGHT: 260 is the exact length of the
# README after the `## How this repo is built` section, not a round number with
# headroom. A guard widened whenever it fires is not a guard — the next line
# added to the README fails this again and is again a decision. The band exists
# (SPEC-021) to keep the README a user-facing quickstart, and a ten-line routing
# section with no commands, flags or numbers does not change that.
assert_line_count_band "A1" "README.md" 100 260
```

**(ii) E1's source list.** Replace:

```bash
for src in README.md CONTRIBUTING.md docs/development.md; do
```

with:

```bash
for src in README.md CONTRIBUTING.md docs/development.md docs/engineering-practices.md; do
```

**(iii) Group X.** Insert the block below **after** the `W6` assertion and
**before** the `# ===== finalise =====` header.

```bash
# ===== Group X — engineering-practices entry point (SPEC-079) =====

PRACTICES_DOC="docs/engineering-practices.md"

# X1 — the practices page exists.
assert_file_exists "X1" "$PRACTICES_DOC"

# X2 — the README points at it. This is the whole "entry point" claim: the page
# is worthless if the front door does not name it.
assert_contains_literal "X2" "README.md" "docs/engineering-practices.md"

# X3 — THE COUNTING GUARD. The inventory block on the page must equal
# `scripts/inventory.sh` output byte-for-byte. This is the assertion that makes
# every number on the page unrottable: it does not check one hand-typed count
# against one hand-typed expectation (which is just a second thing to maintain)
# — it recomputes the whole table from the repo and diffs. The remedy on
# failure is mechanical: run `just inventory`, paste between the markers.
if [ ! -f "$PRACTICES_DOC" ]; then
    fail "X3" "$PRACTICES_DOC does not exist"
elif [ ! -x scripts/inventory.sh ]; then
    fail "X3" "scripts/inventory.sh is missing or not executable"
else
    x3_want=$(./scripts/inventory.sh)
    x3_got=$(awk '/<!-- inventory:begin/{f=1; next} /<!-- inventory:end/{f=0} f' "$PRACTICES_DOC")
    if [ -z "$x3_got" ]; then
        fail "X3" "$PRACTICES_DOC has no content between the inventory:begin/inventory:end markers"
    elif [ "$x3_got" = "$x3_want" ]; then
        ok "X3"
    else
        fail "X3" "inventory block is stale — run \`just inventory\` and paste between the markers:
--- on the page ---
$x3_got
--- computed from the repo ---
$x3_want"
    fi
fi

# X4 — the page describes the test mechanism as it ACTUALLY is.
# There are no golden FILES in this repo: `find . -name '*.golden'` is empty and
# there are no testdata/ directories. The mechanism is byte-exact expected
# output embedded as Go string literals. Deliberately a POSITIVE assertion, not
# a NOT-contains on "golden file": the page has to be free to say what the
# mechanism is NOT, and a NOT-contains would forbid exactly that sentence.
assert_contains_literal "X4" "$PRACTICES_DOC" "embedded as Go string literals"

# X5 — no adjective stands alone. STAGE-021's success criterion, mechanised.
# Every token below is a claim a reader cannot check; the page is required to
# point at a path, a count or a named test instead.
assert_not_contains_iregex "X5" "$PRACTICES_DOC" 'rigorous|comprehensive|world-class|best-in-class|battle-tested|cutting-edge|state-of-the-art'

# X6 — the corrections claim is made BY CITATION, not by count.
# There is no counting rule for "decision records that log their own
# correction" (only 1 of 45 DECs carries an explicit `## Amendment` heading, and
# a keyword grep matches all 45 because ordinary prose uses those words), so the
# page names the specific records. This pins that it keeps naming them.
if [ ! -f "$PRACTICES_DOC" ]; then
    fail "X6" "$PRACTICES_DOC does not exist"
else
    x6_missing=""
    for x6_id in "DEC-004" "DEC-015" "DEC-025" "DEC-043" "DEC-044" "PROJ-006"; do
        if ! grep -F -q -- "$x6_id" "$PRACTICES_DOC"; then
            x6_missing="$x6_missing $x6_id"
        fi
    done
    if [ -z "$x6_missing" ]; then
        ok "X6"
    else
        fail "X6" "$PRACTICES_DOC must cite each correction record by id; missing:$x6_missing"
    fi
fi

# X7 — the page stays an INDEX, not an essay. Same idiom as A1/C2/D2. The
# stated failure mode for this document is that it turns into new prose about
# the project instead of a route to artifacts that already exist; a length band
# is the cheap mechanical form of that. Band, not a tight pin, because unlike
# A1 there is no agreed length for a brand-new artifact to be re-pinned to.
assert_line_count_band "X7" "$PRACTICES_DOC" 150 300

# X8 — `just inventory` is wired, so the remedy X3 names actually exists.
if [ ! -f justfile ]; then
    fail "X8" "justfile does not exist"
elif grep -E -q '^inventory:' justfile; then
    ok "X8"
else
    fail "X8" "no '^inventory:' recipe in justfile"
fi
```

---

### ⑥ `guidance/questions.yaml` — one new question

Append to the `questions:` list.

```yaml
  - id: dec-amendment-heading-convention
    question: "Should decision records adopt an explicit `## Amendment` heading as a convention, so 'decisions that record their own corrections' becomes countable rather than citable?"
    priority: low
    status: open
    raised_by: claude-opus-5
    raised_at: 2026-08-16
    assigned_to: null
    notes: |
      Filed at SPEC-079 design (LD4). The claim is TRUE in substance — DEC-044
      carries a correction box, DEC-043 carries a third correction to the same
      paragraph, DEC-025 carries an explicit `## Amendment`, DEC-004 was
      superseded and kept — but NOT COUNTABLE: exactly 1 of 45 records has a
      `^## Amendment` heading, and a keyword grep for correct/amend/supersede
      matches all 45 because ordinary decision prose uses those words.

      DELIBERATELY NOT ADOPTED at SPEC-079, for three reasons: (1) it is a
      process change binding every future DEC, riding in on a docs spec; (2)
      day-one adoption would be 1/45, which is a weaker claim than naming four
      specific records; (3) SPEC-078 is the cautionary case — the pr:/commit:
      evidence-link convention sat at ZERO adoption because it was documented
      but nothing carried it into the surface that writes.

      What was done instead: docs/engineering-practices.md makes the claim by
      citation, and the inventory table carries a DERIVED row counting records
      with an explicit `## Amendment` section. So the current state is on the
      page, honestly, and rises on its own if the convention is ever adopted.

      RESOLVE WHEN: the derived count reaches 4 or more without anyone having
      declared a convention (adopt it — the practice exists, write it down), OR
      a reader/reviewer asks for the number and the citation list is not enough
      (adopt it and retrofit deliberately as its own spec), OR PROJ-007 closes
      with the count still at 1 (close this question: the citation form was
      sufficient and the convention is not worth the process weight).
```

## Confidence

**0.88.**

Strong because the mechanism is not predicted, it is measured: `inventory.sh`
was executed, the block diff was run, the post-build assertion count was
verified by concatenation (`171`, delta exactly `X1`–`X8`), the static id
extraction was proved equal to a live run's id set, and the full harness was run
with both artifacts staged — confirming A1 is the only assertion disturbed.
Every literal is transcribable and every number on the page is derived by code
that ran.

The residual soft spots, all editorial rather than mechanical:

- **LD5 is a judgement call on someone else's guard.** Re-pinning tight is
  defensible and stated, but a reviewer could reasonably prefer a band with
  headroom, and the zero-headroom cost (any future README reflow re-trips A1) is
  real.
- **LD3's placement is taste.** "Second-to-last, foreshadowed from the intro" is
  argued, not verified; only `X6` (the citations are present) is mechanical.
- **`X5` is a heuristic.** Seven forbidden tokens mechanise "no adjective stands
  alone" only approximately; a page could pass `X5` and still hedge. The
  acceptance criterion's two named words are covered; the general rule is not
  fully mechanisable.

Above the 0.8 threshold, so §14 does not require a question on this basis. The
question filed under literal ⑥ is for LD4's deferred convention, which is
genuinely open on its own merits.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `build/spec-079-practices-entry-point`
- **PR (if applicable):** #168 (open).
- **All acceptance criteria met?** yes. `just test-docs` is green at **171
  distinct assertion ids** (was 163; delta is exactly `X1`–`X8`, confirmed by
  `comm` against the pre-build id set). `just test`, `gofmt -l .` and
  `go vet ./...` are unaffected. The static id extraction and a live run agree
  set-for-set (171 each); the run emits 172 `OK:` lines because `S3` still
  double-emits — deliberately untouched, per Out of scope.
- **New decisions emitted:**
  - none. Every choice was settled at design (LD1–LD6); build produced no
    non-trivial decision of its own.
- **Deviations from spec:**
  - **One row of literal ① is not byte-identical: `| Projects | 7 |` was
    replaced with `| Projects | 9 |`.** This is not a build-time judgement
    call — it is `X3`'s own prescribed remedy applied verbatim ("run
    `just inventory` and paste between the markers"). The design session
    measured 7 projects on a branch cut before PR #165 (`PROJ-008` /
    `PROJ-009` scaffolds, merged 2026-08-16 12:41) landed; that PR merged
    ~2h before the design PR (#166, 14:50), so on `main` the true count is 9.
    Fail-first was run and recorded before the remedy: with literal ①
    transcribed verbatim, `X3` was the **only** failing assertion and its
    failure message named exactly one differing row. Every other line of the
    block, and every other literal, transcribed byte-for-byte.
  - **Cosmetic:** a single blank line separates the new
    `dec-amendment-heading-convention` entry from the previous entry in
    `guidance/questions.yaml`. Literal ⑥ carries no leading blank line, but
    every one of the file's 17 prior entries is preceded by one. The entry's
    own 33 lines are byte-identical to the literal.
  - `cycle:` was edited to `build` by hand rather than via
    `just advance-cycle`, per the build instruction.
- **Follow-up work identified:**
  - **`X3` fired on day one, and that is the finding.** The guard caught a
    real drift between the design session's snapshot and `main` before a
    human ever read the page — the same failure mode the spec was written to
    prevent, demonstrated at the spec's own expense (cf. design finding #4).
    Worth quoting in the ship reflection; no new spec needed.
  - **The `Projects` row now counts candidate briefs.** `PROJ-008` and
    `PROJ-009` are scaffolded candidates, not started projects, so "Projects |
    9" is arguably over-counting relative to what a reader infers. If that
    reads wrong at verify, the fix belongs in `inventory.sh` (filter on
    `status:`), not on the page — which is the design working as intended.
  - **`guidance/questions.yaml` will likely conflict at merge.** Open PR #167
    also edits this file (it flips `tag-ordering-projection` to `answered`);
    this branch appends a different, new question. Both changes are wanted;
    the textual conflict is expected and trivially resolved by keeping both.

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Almost nothing, and that is the literal-artifact contract paying out:
   six `sed -n 'A,Bp'` extractions and three insertions, with no prose
   authored at build. The one genuine ambiguity was **what to do when a
   literal disagrees with the repository** — the spec says "transcribe
   verbatim", `X3` says "run `just inventory` and paste", and on the
   `Projects` row those two instructions conflict. The spec never states
   which wins. It resolves cleanly once you notice that literal ① is
   explicitly a *cached copy* of literal ③'s output (LD1: "the script is the
   source, the page is a cached copy"), so ③ outranks ① wherever they
   disagree — but that precedence is inferable from LD1's argument rather
   than stated as a build rule. A one-line note under "Notes for the
   Implementer" ("if `X3` fires, ③ wins; paste and record it as a deviation")
   would have removed the only pause in the build.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No constraint. One small convention was missing: literal ⑥ does not
   carry the blank-line separator that every other entry in
   `guidance/questions.yaml` has, so build had to decide whether "verbatim"
   included the file's separator convention. Trivial, but it is the same
   class as the §12 "flag-default explicitness" rule — a formatting detail
   that is design-decidable and was left to build to infer. Beyond that, the
   §12(b) table was accurate on every row that could be re-checked here
   (A1 at 260, the page at 268, zero `X5` hits, `E2` clean, the id delta
   exactly `X1`–`X8`), which is why there were no other surprises.

3. **If you did this task again, what would you do differently?**
   — Run `scripts/inventory.sh` **first**, before transcribing the page, and
   diff its output against the embedded block. I transcribed all six literals
   and then discovered the `Projects` drift from `X3`'s failure output, which
   was the correct order for producing fail-first evidence but meant the
   discrepancy surfaced later than it needed to. Doing the round-trip check
   as step one costs ten seconds and tells you immediately which rows, if
   any, the design snapshot has aged out of. The broader version: for any
   spec whose literal embeds derived output, the derivation is the first
   thing to re-run at build, because a literal that caches a computation is
   exactly the literal most likely to be stale by the time build starts.

### Punch-list iteration (2026-08-18, second build session)

Verify returned ⚠ PUNCH LIST with three findings, all one defect class: a claim
on the page contradicted by the path it cites. All three originated in
**literal ①** and were faithfully transcribed, so each was fixed in **both**
`docs/engineering-practices.md` and literal ① in this spec. Parity re-confirmed
after editing: the two are byte-identical except the one pre-existing recorded
`| Projects |` row (md5 equal once that row is normalised, 275 lines each). The
`Projects` deviation was deliberately left as verify recorded it — it is settled,
and `X3` self-corrects it on any rebuild.

- **P1 — the third thing the golden does not pin.** The page named "the budget";
  the comment at `internal/cli/memory_test.go:156-178` names the declared
  `Matched` order. Re-ran the mutation rather than trusting the description:
  `memory.DefaultBudget` 2000 → 2048 reddens `TestMemoryCmd_EndToEndMarkdownGolden`
  on its `- Budget: 2048 tokens` line, so the golden **does** pin the budget.
  Two further tests redden — `TestDefaultBudget_Is2000` (the explicit pin) and
  `TestMemoryCmd_BareInvocationIsPlainRecency` (the second golden, which also
  carries the `## Budget` block). Verify's "reddens exactly that test" is
  therefore one test too narrow; the load-bearing half of the finding stands.
  Mutation reverted, tree verified clean.

- **P2 — what `STAGE-022` actually closes.** "Closing the first three is the
  scope of STAGE-022" replaced with a sentence that names the items rather than
  an ordinal range, so it survives the list being reordered. **Sequencing:**
  PR #167 (`chore/defer-stage-022-decisions`) was checked and is **OPEN, not
  merged**; it narrows STAGE-022 to lint + coverage + the `Entries:` envelope
  and moves `MergeTags` / `$EDITOR` to `PROJ-001-mvp/backlog.md`. The
  replacement was written so that **every clause is true both today and after
  #167 merges** — it names lint, coverage and the `Entries:` envelope defect
  (in scope under both states), routes benchmarks to `PROJ-009` via STAGE-022's
  own *Explicitly out of scope* section, and states that gating the
  documentation assertions in CI is owned by nothing. Confirmed: `test-docs`
  appears **zero** times in STAGE-022 under both `main` and #167, and nowhere
  in PROJ-008 or PROJ-009. The sentence does not enumerate how many defects
  STAGE-022 carries, which is the only clause the merge would flip.

- **P3 — a universal quantifier false for 6 of 75.** Verified independently:
  75/75 archived specs end with `## Reflection (Ship)` (it is the last `##`
  section in every one); **69** carry `### Build-phase reflection`. The six that
  do not are SPEC-045, SPEC-046, SPEC-048, SPEC-049, SPEC-050 and SPEC-071 — as
  verify found. One nuance verify did not record: **SPEC-046 does carry a
  build-phase reflection**, under the heading `### Honest reflection`; it is
  prose rather than the three questions. So five specs carry none at all, and
  six lack the standard heading.
  **Chosen fix: derive the count, per LD1 — not soften the wording.**
  `scripts/inventory.sh` (and literal ③) gain a sixteenth row,
  `…of those, also carrying a build-phase reflection`, whose "Where it lives"
  column names the exact heading grepped — which is what makes 69 reproducible
  and what makes SPEC-046's non-standard heading honestly outside the count
  rather than silently miscounted. The prose now points at that row and says
  explicitly "not by all of them".
  **Why derive rather than weaken:** softening to "most" or "generally" fixes
  the false quantifier but breaks the page's *second* rule — every claim points
  at a path, a count, or a named test — trading a checkable falsehood for an
  uncheckable vagueness. Deriving satisfies both rules at once, and it makes the
  claim self-maintaining: when a seventh spec is archived without a build-phase
  reflection, `X3` goes red and the number is re-derived, instead of the prose
  quietly going wrong again. `scripts/inventory.sh`'s own header comment already
  prescribes exactly this ("If you find yourself wanting to type a current-state
  number into the prose, add a row here instead"), so this is the fix the
  artifact asked for, not a new convention.

**The new row was mutation-tested both ways.** Page-side: `69` → `70` turns
`X3` red with the prescribed remedy. Source-side (the check that matters):
renaming one archived spec's `### Build-phase reflection` heading drops the
derived count to 68 and turns `X3` red — so the row pins the page to the
repository, not to itself. Both mutants reverted; tree verified clean after each.

**Gates after the punch list, all green:** `just test` (all packages ok),
`just test-docs` (**ALL OK**, 172 `OK:` lines / **171 distinct ids** — unchanged,
no assertion was added or weakened), `gofmt -l .` (empty), `go vet ./...`
(clean). Bands re-checked: `README.md` is **untouched at 260 lines** (`A1`'s
ceiling, zero headroom, as LD5 requires) and the page is **275 lines**, inside
`X7`'s 150–300. `X5` remains green — none of the new prose uses a forbidden
token.

**One honest limit, not papered over.** All three findings were prose that the
mechanical guards cannot reach: `X3` cannot see them and `X5` was green through
all three. P3 is now genuinely pinned, because its claim was converted into a
derived number. P1 and P2 are **not** pinned by any assertion and cannot be
without one that only *appears* to cover the claim — an `assert_contains` on
"declared `Matched` order" or on "`Entries:` envelope" would go green on the
words while saying nothing about whether the cited path still agrees with them,
which is the exact failure mode `X5` already demonstrates (green, has teeth,
pins the wrong proposition). No such assertion was added. Re-verify remains the
control for this class, which is why `cycle:` stays at `build`.

- **Files changed in this iteration:** `docs/engineering-practices.md`,
  `scripts/inventory.sh`, and this spec (literal ①, literal ③, LD1's row count,
  the stale `PR: none opened` line, and this section).

---

## Verify

**Verdict: ⚠ PUNCH LIST** (fresh independent verify session, 2026-08-18,
`cdc1033` / PR #168).

Gates re-run independently, all green: `just test` (all packages ok),
`just test-docs` (**ALL OK**, 172 `OK:` lines / **171 distinct ids**),
`gofmt -l .` (empty), `go vet ./...` (clean).

### The mechanical diff: six literals, all faithful

Each literal was extracted from this spec by line range and diffed against the
working tree — not read.

| Literal | Range | Result |
|---|---|---|
| ① `docs/engineering-practices.md` | 569–836 (268 lines) | **1 line differs** — the recorded `Projects` deviation. With that row substituted, md5 and byte count match exactly (15658 bytes). |
| ② `README.md` section | 848–856 | Byte-identical, inserted immediately before `## Where to go next`; README = **260** lines. |
| ③ `scripts/inventory.sh` | 867–948 | Byte-identical; committed mode **`100755`**. |
| ④ `justfile` recipe | 960–962 | Byte-identical, appended after `info:` in HELPERS (clear of `F2`'s `test:` read). |
| ⑤ `test-docs.sh` A1 / E1 / Group X | 979–986, 992, 1005–1090 | Byte-identical (Group X differs only by the blank line before `# ===== finalise =====`). |
| ⑥ `guidance/questions.yaml` | 1100–1132 | Byte-identical; the one blank separator line is the file's convention (17 of 18 prior entries carry one) and is recorded under Deviations. |

Literals were also diffed between `cc8f6f6` (design) and `cdc1033` (build): the
build did **not** edit any literal to match what it wrote. Only front-matter
`cycle:` and `## Build Completion` changed.

### The `Projects` deviation was resolved correctly

`| Projects | 7 |` → `9` is the right call, not a build-time judgement. LD1
states the precedence in terms ("the script is the source, the page is a cached
copy"), `X3`'s own failure message prescribes the remedy, and the alternative —
transcribing `7` — ships a red gate. No assertion was weakened. The cause is
confirmed from git: `0bcd124` (#165, PROJ-008/PROJ-009 scaffolds) landed
2026-08-16 12:41, the design PR `cc8f6f6` (#166) at 14:50, off a branch cut
before it. Honestly recorded in Build Completion, the commit message, the PR
body, and the STAGE-021 entry. `Projects: 9` itself is settled as correct.

### `X3` passes for the right reason, and has teeth

- Script output vs. the page block extracted with `X3`'s own `awk` range:
  **byte-identical, 17 lines / 1157 bytes each — both non-empty.** The `[ -z ]`
  guard is not what is passing it.
- The empty-block branch was exercised directly (markers kept, contents
  stripped): fails with `has no content between the inventory:begin/inventory:end
  markers`. A silently-empty diff cannot masquerade as success.
- **`X3` pins the page to the repository, not to itself.** Adding a real
  `decisions/DEC-047-*.md` (45 → 46) — mutating the *source*, not the page —
  turned `X3` red with the prescribed remedy in the message.

### Assertion-id delta is exactly `X1`–`X8`

Static extraction over `cc8f6f6` vs `cdc1033`: **163 → 171**. `comm` both
directions: added = `X1 X2 X3 X4 X5 X6 X7 X8`; removed = **none**. The static
set and the live-run set are identical (171 each); only `S3`'s multiplicity
differs, which is out of scope by design.

### Every Group X assertion was mutation-tested, and every mutant was real

Each mutation was applied, confirmed to have actually changed the artifact
(content hash, or file mode for the `chmod` case), run, and reverted; the tree
was verified clean after each.

| Mutation | Caught by |
|---|---|
| page removed | `X1` (+ `E1`/`X3`–`X7` collaterally) |
| README's link to the page rewritten | `X2` only |
| one digit drifted in the page's block | `X3` only |
| `chmod -x scripts/inventory.sh` | `X3` — *"missing or not executable"* |
| markers deleted | `X3` |
| block emptied, markers kept | `X3` — *"no content between the markers"* |
| a 46th DEC added to the repo | `X3` |
| `embedded as Go string literals` → `stored in golden files` | `X4` |
| `comprehensive` inserted into the page | `X5` |
| `DEC-025` citation removed | `X6` |
| page truncated to 140 lines | `X7` |
| `inventory:` recipe renamed | `X8` |
| one blank line appended to README | `A1` — *"261 lines (expected 100..260)"*, confirming LD5's zero headroom |
| `../SECURITY.md` → a non-existent path **in the page** | `E1`, confirming the source-list extension is live |

### Links

All 41 path links in the page resolve (checked independently of `E1`, resolved
relative to `docs/`); the two anchor links target
`#where-this-project-was-wrong-on-the-record`, which is the correct GitHub slug
for `## Where this project was wrong, on the record`. `E1` strips anchors, so
that one was checked by hand. The README section's single link resolves. The
`E2` trap held: the page uses the relative `development.md` form, zero hits for
the literal `docs/development.md`.

### Claims audited against the repository

Spot-checked ~30 claims. Confirmed true and re-derived: no `*.golden` files and
no `testdata/` directories (both 0); all five named tests exist at the named
paths; `TestPackageReadsNoWallClock` / `TestPackageEmitsNoReservedTagNamespace`
really are source-scanning walkers; all 45 DECs carry `confidence:` and none
claims 1.0; `DEC-041` is the only numbering gap in DEC-001..046 and its
reservation is recorded at `PROJ-001-mvp/backlog.md:1041`; the `## Amendment`
heading in `DEC-025` is verbatim as quoted; `DEC-043`'s third correction is
dated 2026-08-10 and does replace a figure-dependent argument with a
pool-minimum bound; `DEC-044`'s correction box carries ≈110 → 25 and traces
where the figure spread; the PROJ-006 close is quoted word-for-word from
`Project-Level Reflection`; a keyword grep for `correct|amend|supersede` matches
**45 of 45** records, exactly as claimed; all six `W`-series descriptions match
their source comments; CI runs only `gofmt`/`vet`/`test`, with no
`golangci-lint`, no `-cover`, no badge, and no `test-docs`; `SECURITY.md` says
"enforced by review rather than assumed" and no test asserts it; `DEC-040` does
remove the signing/Gatekeeper path; the retro's headline finding is stated as
quoted; the cask→formula item (v0.5.2, 2026-07-30) does postdate the retro
(2026-07-04); `archive-spec.sh` really does reject `<answer>` placeholders;
`lifetime-save` really does `mkdir -p docs/reports/lifetime` on first run; and
the `framework-feedback/` / `research/` / `blog/` / `reports/` descriptions each
match the directory contents.

**Three claims did not survive.** They are the punch list below.

### Punch list

All three are defects in **literal ①**, faithfully transcribed by build. Each
must be fixed in **both** the page and literal ① in this spec, so the next
mechanical diff stays clean.

**P1 — the page names the wrong third item, and the one it names is provably
pinned.** `docs/engineering-practices.md:103-110` says the comment above
`TestMemoryCmd_EndToEndMarkdownGolden` names "three things that golden does not
pin — the pool composition, the fusion constants, and **the budget**."
`internal/cli/memory_test.go:156-178` names: the **pool composition**, the
fusion **constants**, and the declared **Matched ORDER**. The budget is not on
that list — and `memoryCmdGolden1` (`internal/cli/memory_test.go:104`) contains
a whole `## Budget` section including `- Budget: 2000 tokens`, so the golden
*does* pin it. Confirmed by mutation: `memory.DefaultBudget` 2000 → 2048 reddens
exactly `TestMemoryCmd_EndToEndMarkdownGolden`. This is the page's paragraph
about mutation-verified claims, and it is the one claim on the page that fails a
mutation check. Fix: name the declared Matched order as the third item.

**P2 — `Closing the first three is the scope of STAGE-022` is contradicted by
STAGE-022.** `docs/engineering-practices.md:248-249`. The first three bullets of
`## What this does not measure` are benchmarks, lint+coverage, and
documentation-assertions-not-in-CI.
`STAGE-022-measured-and-enforced.md:91-95` heads a section **"Explicitly out of
scope"** with "**Benchmarks and any scale/perf or concurrency harness.**
Deferred to **PROJ-009**." And `test-docs` appears nowhere in STAGE-022 — a grep
across PROJ-007/008/009 finds nothing scoping it into CI. STAGE-022's in-scope
list is golangci-lint, coverage, and the three coupled defects — i.e. it closes
the page's **second** and **fifth** bullets, not the first three. A reader who
clicks the link the sentence provides lands on the sentence's refutation.
Fix: state what STAGE-022 actually closes; benchmarks route to PROJ-009; and
note that gating the documentation assertions in CI is currently owned by
nothing (which is a legitimate thing for this page to say, and stronger than a
wrong routing claim).

**P3 — "each end with build-phase and ship-phase reflections" is false for 6 of
75.** `docs/engineering-practices.md:89-93`. The bolded lead sentence ("Every
archived spec carries its own reflection") is **true** — 75/75 carry
`## Reflection`. The sentence after it is not: the three things it says are
answered ("what was unclear, what was missing, what would be done differently")
are the *build-phase* reflection's three questions, and six archived specs carry
no build-phase reflection at all — `SPEC-045`, `SPEC-046`, `SPEC-048`,
`SPEC-049`, `SPEC-050` (all PROJ-004) and `SPEC-071`. Each has a substantive
`## Build Completion` and a `## Reflection (Ship)`, but none contains the
"what was unclear" question. 69 of 75. Note the shape: this is a **current-state
count written into prose as a universal quantifier**, which is exactly what LD1
says must be derived rather than typed — so either re-word to the claim that is
true without a count, or make it an inventory row and let `X3` hold it.

### Not punch-list items

- **The unstated precedence** (build reflection Q1: "transcribe verbatim" vs
  `X3`'s "run `just inventory` and paste"). Real, honestly raised, and correctly
  resolved anyway. It is a **ship-reflection / AGENTS.md §12 candidate**, not a
  build defect: *when a literal caches derived output, the derivation outranks
  the cache, and re-running it is build step one.* Under the §12 codification
  meta-rule this is N=1 same-outcome — record it at ship, hold for N=3 (or an
  opposing-outcome pair).
- **`F4` after the `FAIL_COUNT` early-exit** — confirmed
  (`scripts/test-docs.sh:1428`, after the `exit 1`). It does not matter here:
  the page's number is derived **statically** precisely so it does not depend on
  a run, and on a red run the harness exits 1, so nothing is comparing sets.
  `F4` also predates this spec. Worth one clause in §12(b) row 4 at most.
- **`PR (if applicable): none opened — build ran to commit only`** in Build
  Completion is now stale: PR #168 is open. Trivial; fold into the punch-list
  commit.
- **`Projects: 9`** — settled as correct by the maintainer; not re-litigated.
  The same reasoning covers `Stages: 21`.

### Also checked

Acceptance criteria: 9 of 10 met. The one that is not is
"no bare adjective stands alone on the page" in its **stage** form — `X5` is
green and the seven forbidden tokens are absent, but STAGE-021's criterion is
"every claim backed by a countable artifact," and P1–P3 are claims backed by an
artifact that contradicts them, which is the criterion's substance rather than
its mechanical proxy.

Constraints: `test-before-implementation` held (tests written at design in
`## Failing Tests`, transcribed from literal ⑤; build ran fail-first and
recorded it). `one-spec-per-pr` held — PR #168 references SPEC-079 only.
`no-secrets-in-code` — nothing here touches credentials. No Go source touched,
so the Go-scoped constraints do not apply. No DEC was owed: every choice traces
to LD1–LD6, and build emitted no non-trivial decision of its own.

Build reflection honesty: **yes.** All three answers are specific and
self-critical, the deviation is disclosed in four places with its cause, and
Q3's answer ("run the derivation first") is the correct general lesson. Build
also flagged the `Projects`-row and `F4` questions to verify rather than
burying them.

**A punch-list delta deserves its own re-verify before this spec advances.**
SPEC-075's re-verify found two further defects; SPEC-078 skipped one. All three
items here are edits to prose that the mechanical guards do **not** reach —
`X3` cannot see them, and `X5` is green through all three — which is precisely
the class that needs a second pair of eyes rather than a re-run.

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

4. **What can a user do now that they couldn't before?** — one sentence,
   before → after; quote the confirming number if one exists, name the outcome
   if not. Write `none` if this spec has no user-visible outcome — that is a
   real, greppable result, not a blank. This is the line a brag's `impact` field
   is transcribed from, and both halves are already written above (## Context is
   the before, ## Goal is the after): confirm the prediction, don't reconstruct
   it from memory.
   — <answer>

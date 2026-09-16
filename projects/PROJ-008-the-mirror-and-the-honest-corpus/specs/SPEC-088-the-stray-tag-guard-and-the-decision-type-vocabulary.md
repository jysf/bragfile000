---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-088
  type: chore                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
                                   # FRAMED 2026-09-15 — GO at S. The file was
                                   # created at SPEC-087 ship to CLAIM the id;
                                   # framing kept two of the four items routed
                                   # here and split two out (SPEC-092, SPEC-093).
  blocked: false
  priority: high                   # RAISED at framing from medium. The stray-tag
                                   # guard has to precede SPEC-086's design: five
                                   # of six historical instances came from a
                                   # session writing a long spec, and SPEC-086's
                                   # design writes the longest one in the stage.
  complexity: S                    # CONFIRMED at framing (was provisional).
                                   # The half that threatened S — a user-facing
                                   # table gaining rows — is the branch framing
                                   # rejected. See ## Complexity.

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-09-07
  framed_at: 2026-09-15

insight:
  confidence: 0.90                 # both items are mechanical and both were
                                   # measured directly on `main` at `6dda56a`;
                                   # the guard was probed under the §12 mutation
                                   # protocol rather than argued (P-1).

references:
  decisions: []                    # Emits none. DEC-050 is spoken for by
                                   # SPEC-086; if design finds it needs a record,
                                   # the next free number is DEC-054, NOT 050.
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-087                     # routed item A here (V-F3); its `inv_row`
                                   # helper and its derive-don't-pin shape are
                                   # the precedent both assertions follow
    - SPEC-082                     # authored Z7, where item A's hard-fail lives
    - SPEC-089                     # routed item B here at its ship; two of the
                                   # six affected files are its own
    - SPEC-092                     # SPLIT OUT at framing — `Y4` derives
    - SPEC-093                     # SPLIT OUT at framing — id reservation
---

# SPEC-088: the stray-tag guard and the decision-type vocabulary

## Context

> **Cycle: frame — FRAMED 2026-09-15. GO at complexity S.** This file was
> created at SPEC-087 ship to *claim* the id; it held transcribed evidence and
> no decisions. Framing has now made them.
>
> **Every number below was measured on 2026-09-15 against `main` at `6dda56a`**,
> and every one is stated as two measurements plus a unit with the command that
> produced it. Where a number in the framing prompt or in this file's own
> pre-framing text turned out to be wrong, the correction is marked
> **CORRECTION** and the old value is kept visible.
>
> **The spec was renamed at framing.** It was *"`Y4` derives and the
> decision-type vocabulary"*; `Y4` is no longer here. Same id, new filename —
> the id stays claimed by a file throughout.

Four items were routed to this id. Framing kept **two** and split **two** out,
under the standing instruction to keep a spec small and split rather than
absorb. What remains is a single coherent job: **two harness assertions, one
template line, and one regeneration of a derived table.**

| Item | Verdict | Home |
|---|---|---|
| A — the decision-type vocabulary (SPEC-087 V-F3) | **GO** | **here** |
| B — a guard against stray tool-call XML (routed at SPEC-089 ship) | **GO** | **here** |
| C — `Y4` derives (SPEC-087 LD6) | **GO** | **SPEC-092** |
| D — id reservation (routed by the v0.7.0 release plan) | **GO** | **SPEC-093** |

The two kept items are not merely co-located. They add one `scripts/test-docs.sh`
assertion each, and **an assertion id is the one thing in this repo that moves a
user-facing derived table**: `Documentation assertions (distinct ids)` is a row
in `docs/engineering-practices.md`. Landing them together pays that cost **once**
(198 → 200) instead of twice. Item C's whole purpose is to *remove* a coupling of
that kind, and item D's fix may *add* rows to the same table — which is exactly
why neither belongs in the same PR under `one-spec-per-pr`.

### Item A — the template advertises five `insight.type` values; the inventory counts two

`decisions/_template.md:6` offers
`decision | analysis | recommendation | observation | reservation`.
`scripts/inventory.sh` emits a row for `decision` and one for `reservation` —
and nothing else. A `DEC-*.md` carrying any of the other three is counted by
**neither** row, vanishes from the page, and hard-fails `Z7`. Measured by
SPEC-087 on `main` at `f4658b0` with a `type: analysis` stub (simulation S-3):

```
FAIL: Z7: the inventory covers 49 of 50 decisions/DEC-*.md files (48 decision + 1 reservation).
```

**This predates SPEC-087.** It arrived with `Z7` at SPEC-082 and fires
identically today, so it is not a regression. The hard fail is the *correct*
behaviour — an uncounted decision is invisible and the page under-reports. What
is wrong is that **the template is where an author picks the value**, and it
currently invites three values that break the harness. SPEC-087 verify added a
warning line to the template naming the consequence; that line is a warning,
**not** the decision.

**M-4 — the three extra values have never been used, in the repo's whole
history.** 52 of 52 `decisions/DEC-*.md` files carry a countable type (51
`decision`, 1 `reservation`), and across every commit on every branch, 65 of 65
`type:` lines ever *added* to a decision record were one of those two (64
`decision`, 1 `reservation`). Zero uses of `analysis`, `recommendation` or
`observation`, ever.

```
$ grep -h '^  type:' decisions/DEC-*.md | sed 's/#.*//' | sort | uniq -c
  51   type: decision      1   type: reservation
$ ls decisions/DEC-*.md | wc -l
  52
$ git log --all -p -- 'decisions/*.md' | grep -E '^\+  type:' | sed 's/#.*//' | sort | uniq -c
  64 +  type: decision      1 +  type: reservation
```

So the template advertises a vocabulary three-fifths of which has no instance
and breaks the build on first use. That is the finding that decides the fork.

### Item B — sessions leave their own tool-call syntax in the files they write

A session writing a file has left a closing tool-call tag as the last line of
it. **M-1 — measured over the full history of every tracked path, both tag
forms, on every branch:** 8 additions of a whole-line closing tag, across 4
commits and **6 distinct files** — and **0** additions outside `*.md`. (The
table below has five rows because `ebdc271` added two of the files.)

```
$ git log --all -p --no-color -- '*.md' \
    | grep -E '^\+[[:space:]]*</[A-Za-z_:.-]+>[[:space:]]*$' | sort | uniq -c
   6 +</content>
   2 +</invoke>
$ git log --all -p --no-color -- . ':(exclude)*.md' \
    | grep -E '^\+[[:space:]]*</[A-Za-z_:.-]+>[[:space:]]*$' | sort | uniq -c
   (no output)
```

| File | Added in | Stripped in | Survived |
|---|---|---|---|
| `SPEC-075` (`content` + `invoke`) | `ebdc271` (#144) | `e0a0a3c` (#145) | its own build |
| `DEC-046` (`content`) | `ebdc271` (#144) | `7c615f5` (#210) | **~66 PRs** |
| `SPEC-089` (`content` + `invoke`) | `55d8925` (#206) | `7c615f5` (#210) | 4 PRs, a full verify |
| `SPEC-091` (`content`) | `73dbf38` (#211 branch) | `ec47382` | pre-merge |
| `DEC-053`, `STAGE-027` (`content`) | `6208eeb` (#212 branch) | `e1436a1` | pre-merge |

**CORRECTION to the framing prompt and to this stage's backlog note: six files,
not five.** Both name `SPEC-089`, `DEC-046`, `DEC-053`, `STAGE-027` and
`SPEC-091`. The sixth is **`SPEC-075`**, which took both tags in the same commit
as `DEC-046` (`ebdc271`, #144) and stripped them one PR later at its own ship —
recorded in that spec's own reflection, and never counted in the running total.
The pattern is therefore older and broader than "five files since #206": it has
recurred across **two projects** over **35 days** (`ebdc271`, 2026-08-10 →
`6208eeb`, 2026-09-14), and the only reason `main` is clean today is that it was
caught by reading rather than by a check — twice in the last eight days.

**M-2 — `main` has zero today.** 0 of 437 tracked files match the whole-line
anchored form; the grep exits 1.

**M-3 — the false-positive surface is real, it moves, and the anchor clears it.**
Measured twice, because this branch changed it.

- **On `main` at `6dda56a`:** 4 of 437 tracked files mention these tags in prose,
  7 lines total — `NEXT-SESSION-PROMPT.md` (2), `SPEC-075` (2), `SPEC-089` (1),
  `STAGE-023` (2).
- **On this branch:** 4 of 439 files (tracked plus the two new spec files), 8
  lines — `NEXT-SESSION-PROMPT.md` (2), `SPEC-075` (2), `SPEC-089` (1), and
  **this file** (3). `STAGE-023` fell to 0 because framing rewrote the backlog
  note that carried its two mentions.

```
$ { git ls-files; git ls-files --others --exclude-standard; } | sort -u \
    | tr '\n' '\0' | xargs -0 grep -cE '</(content|invoke)>' | grep -v ':0$'
```

Every one of these writes the tag inline, inside backticks or mid-sentence; none
writes it alone on a line. **That the set changes composition between two commits
on the same day is the argument for the anchor** rather than for a file
allow-list: an allow-list would have gone stale in a single PR.

**P-1 — the probe, under the AGENTS.md §12 mutation protocol.** Run at framing
rather than asserted, because a guard argued for on paper is what this stage has
already been burned by.

- **Target:** `NEXT-SESSION-PROMPT.md` — chosen deliberately: it is one of the
  four files that mention the tags in prose, so the probe tests the guard and
  its false-positive surface in the same run.
- **The edit, stated because a hash is not a reproduction (§12):** append one
  line to the end of the file whose entire content is the four-character closing
  `content` tag — `printf '</content>\n' >> NEXT-SESSION-PROMPT.md`.
- **Backup:** `/tmp/spec088-probe-backup-1789512746.md`, taken before the edit;
  restored with `cp`, never `git checkout`.
- **Hashes, checked BEFORE the gate ran** (the §12(b) refinement's ordering):
  pre `c9f625f17b05312db10eae8892b092cfb0403cb0ea74069747de69769163a847`,
  post `7ba747a741fb9f8c372b78d3ee3d34b1e114a712a361d369ed56689caf5179d8`,
  restored `c9f625f1…` — identical to pre. `git status --porcelain` empty after.
- **Verdict:** before the mutation the anchored sweep over all tracked files
  exits 1 with no output *while all four prose-mention files are present*; after
  it, exit 0 with exactly one hit, `NEXT-SESSION-PROMPT.md:203`. The guard has
  teeth and does not fire on prose.

**Cost of the sweep: 0.15s wall clock** over 437 tracked files
(`time (git ls-files -z | xargs -0 grep -lE '^[[:space:]]*</(content|invoke)>[[:space:]]*$')`).

## Goal

**Two ways a session can quietly corrupt this repo's own documents are caught by
`just test-docs` instead of by a reader.** After this spec: a tracked file that
ends in the tool-call syntax that wrote it fails the harness on the commit that
introduced it, rather than surviving 66 PRs inside a record that four later
specs cite; and `decisions/_template.md` stops offering an author three
vocabulary values that make their decision record invisible to
`docs/engineering-practices.md` and hard-fail `Z7` on first use.

Neither half touches Go, the binary, the corpus, or any user-facing behaviour.
The single user-visible artifact either half moves is one row of one derived
table — `Documentation assertions (distinct ids)`, 198 → 200 — and it moves once
for both.

## The fork in item A — decided: **(b)**, narrow the template

The pre-framing file left three branches open. Framing measured them.

**(b) — narrow `decisions/_template.md` to the two values the harness counts.
CHOSEN.** The evidence is M-4: the three values being removed have zero
instances in 52 records and zero in 65 historical additions. Removing them
deletes nothing an author has ever wanted and closes the trap at the exact point
where the author chooses. One line of one template.

**(a) — teach `scripts/inventory.sh` three more rows. REJECTED.** It would grow
a **20-row user-facing table to 23** to report a population of **zero, three
times over**, and turn `Z7`'s two-term sum into a five-term one — more harness to
be wrong in, in service of a vocabulary nobody uses. It is also the branch the
pre-framing file named as the half that could push this past **S**, and framing
confirms that judgement rather than overriding it: a user-facing table gaining
rows is a doc change with its own review surface.

**(c) — something else, measured: a *derived* guard instead of a literal pin.
CHOSEN AS THE COMPANION TO (b), not as an alternative.** Narrowing alone leaves
nothing to stop the vocabulary drifting back — and this stage's own recorded
lesson is that *the guard is what catches it, not the note*. But the guard must
not be `assert_contains_literal "…" "decisions/_template.md" "decision | reservation"`:
a literal pin is precisely the anti-pattern SPEC-087 spent a cycle removing from
`Y3`, and it would need a hand-edit the first time a legitimate value is added.

The derived shape, which design locks: **every `insight.type` value the template
advertises must have a row in `scripts/inventory.sh`'s emitted output.** Read the
vocabulary out of the template's own comment, read the rows out of the script's
own emission (`inv_row` already addresses a row by its What-column label), and
compare the two sets. It fails in both directions — a template value with no row,
and a row whose type the template never offers — and it needs no edit when a
future spec legitimately adds a type *with* its row, which is the whole point.

## Complexity

**S. Confirmed at framing; it was provisional at S with a named risk, and the
risk was the branch framing rejected.**

| Signal | Measurement |
|---|---|
| Go code | **none** |
| Decision records emitted | **none** |
| Files modified | **3** — `scripts/test-docs.sh`, `decisions/_template.md`, `docs/engineering-practices.md` (regenerated, not hand-edited) |
| New assertion ids | **2** (198 → 200 distinct) |
| User-facing table rows added | **0** — the branch that added 3 was rejected |
| Precedent at the same size | **SPEC-087** held S on one file, `+77/-39`, no Go, no DEC, table byte-identical |

The two items share one file, one regeneration and one review. Splitting them
would pay the `X3` regeneration twice and produce two PRs that each say "add one
assertion to `test-docs.sh`" — the split would cost more than it saves, which is
the opposite of the reason C and D were split out.

## GO / NO-GO

**GO**, at complexity **S**, scoped to items A and B.

**Sequencing — this spec lands before SPEC-086's design.** Not a preference; the
measurement says so. Five of the six historical instances of item B's defect
were produced by a session writing a long spec or decision file, three of them
in the last eight days (`SPEC-091`, `DEC-053`, `STAGE-027`, all caught by eye
pre-merge). **SPEC-086's design session writes the longest document in this
stage** — a 649-line spec it will substantially rewrite — and it also authors
`DEC-050`, a *second* file of exactly the shape that carried the defect for 66
PRs in `DEC-046`. Landing the guard first costs one small PR; landing it after
means the one session most likely to reproduce the defect runs unguarded.

Item A carries no sequencing claim of its own and rides along.

**What would have made this NO-GO, and did not:** if item B's anchored form had
fired on any of the four prose-mention files, the guard would have been a
heuristic that a future doc about this very defect would trip — P-1 shows it does
not. And if M-4 had found even one legitimate `analysis`/`recommendation`/
`observation` record, fork (b) would have been deleting a used vocabulary and
this would have gone to the user instead of being decided here.

**Explicitly not in this spec** (each has a file, none is reserved in prose):

- **`Y4`'s two literal pins → SPEC-092.** GO at S. Its independent oracle exists
  — it needs a dependency decision this spec does not want.
- **Id reservation → SPEC-093.** GO at M provisional. One branch of it adds rows
  to the same user-facing table fork (a) was rejected for.
- **`guidance/questions.yaml`.** Filing a question here would move the two rows
  `Y4` pins and force a re-pin this spec would then have to carry — the coupling
  SPEC-092 exists to remove. No question was filed at framing.

## Inputs

- **Files to read:** `scripts/test-docs.sh` (`inv_row`, `X3`, `Y3`, `Z7` — the
  derive-don't-pin shape, and the two floors that make an assertion non-vacuous);
  `scripts/inventory.sh` (the two decisions rows and the doc-assertion counter,
  which counts **distinct quoted literal ids**, so a loop-generated id runs but
  is never counted); `decisions/_template.md` (lines 1–16); `decisions/DEC-041-*.md`
  (the one tombstone).
- **Related code paths:** none. No Go.

## Outputs

- **Files modified:**
  - `scripts/test-docs.sh` — two new assertions (one per item), each with a
    non-vacuity floor, following `Y3`/`Z7`'s shape.
  - `decisions/_template.md` — the `insight.type` vocabulary narrowed to the
    values the inventory counts; SPEC-087's warning line adjusted to match.
  - `docs/engineering-practices.md` — the inventory block **regenerated** with
    `just inventory` and pasted whole between the markers. Never hand-edited;
    `X3` compares it byte for byte.
- **Files created:** none.
- **Database changes:** none.

## Acceptance Criteria

*Written at design.* Must be numbers to diff against: the distinct-id count
before and after (198 → 200, re-derived not predicted), the `X3` regeneration
diff, and for each new assertion a mutation that turns it red with the edit
stated, not only its hash.

## Failing Tests

*Written at design.* Three standing traps, all earned in this stage:

- **A green suite is not evidence.** SPEC-089's guard covered two of five keys
  with all 14 packages green. Each assertion needs a positive control that fires.
- **An assertion that enumerates its own scope by hand is the defect it guards
  against.** Derive the scope from the thing being guarded — `git ls-files` for
  item B, the template's own comment and the script's own emitted rows for item A.
- **A floor, or the assertion is vacuously true at the empty state.** `Y3` went
  green on an emptied `decisions/` while borrowing `Z7`'s floor; both now floor
  the same quantity independently.

## Implementation Context

### Decisions that apply

- None binding. This spec emits none. **If design concludes it needs a decision
  record, the next free number is `DEC-054` — `DEC-050` is claimed by SPEC-086
  and was already renumbered around once (#207).**

### Constraints that apply

- `one-spec-per-pr` — blocking. Items A and B ship in one PR under one spec id;
  C and D have their own ids for this reason.

### Prior related work

- **SPEC-087** (shipped) — made `Y3` derive, and proved by mutation M-1 that an
  assertion re-implementing the producer's own filter is blind by construction.
  Both new assertions read what the producer *emits*.
- **SPEC-089** (shipped) — routed item B here; two of the six affected files are
  its own, and its ship is where the sweep was first run.
- **PR #210, #211, #212** — the three most recent strippings, by hand.

### Out of scope (for this spec specifically)

- `Y4` (SPEC-092), id reservation (SPEC-093).
- Front-matter well-formedness. **Found at framing, not fixed:**
  `SPEC-090`'s front-matter is unterminated on `main` — its head carries one
  `---` where every sibling carries two (`for f in projects/PROJ-008-*/specs/*.md;
  do head -60 "$f" | grep -c '^---$'; done` → 2, 2, **1**, 2). It is a different
  defect class from item B (malformed structure, not injected syntax), it belongs
  to another spec's file, and folding it in would put two specs in one PR. Routed
  to STAGE-023's backlog as a candidate, unowned.
- Running the documentation assertions in CI. `just test-docs` is local-only
  (`.github/workflows/ci.yml` has `test`, `lint` and `coverage` jobs and no
  test-docs job; `docs/engineering-practices.md` says so under *What this does
  not measure*, and names it as owned by nothing). **Both new assertions
  therefore gate only what a human or an agent runs locally** — which is the
  honest scope of this spec's claim, and a reason not to overstate it at ship.

## Notes for the Implementer

- The probe in P-1 is reproducible as stated; re-run it rather than trusting it,
  and back up to `/tmp` first.
- Quote `--include` globs; zsh expands unquoted ones and the search silently
  does not run. `git ls-files -z | xargs -0 grep` avoids the class entirely and
  is what P-1 used.
- `$?` after a pipe is the last command's status. The sweep's exit code is the
  assertion's input — read it deliberately.
- Write the tag forms inline in backticks, never alone on a line, or the new
  assertion will fail on the spec that introduces it. This file carries them on
  3 lines and passes the anchored sweep.
- **Two shell traps cost this framing session a re-run each, both on the
  self-check itself.** `FILES=$(…); grep … $FILES` does **not** word-split in
  zsh — the four filenames arrive as one argument and the sweep reports "No such
  file or directory" with exit 2, which is easy to read as a pass. And BSD
  `xargs` has no `-a`, so `xargs -a list grep …` fails the same silent way. The
  form that works on both platforms is
  `tr '\n' '\0' < list | xargs -0 grep -nE …`, and the exit code to read is
  `grep`'s: **1 means clean**.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - `DEC-NNN` — <title> (if any)
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - [any new specs for the stage's backlog]

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

4. **What can a user do now that they couldn't before?** — one sentence,
   before → after; quote the confirming number if one exists, name the outcome
   if not. Write `none` if this spec has no user-visible outcome — that is a
   real, greppable result, not a blank. This is the line a brag's `impact` field
   is transcribed from, and both halves are already written above (## Context is
   the before, ## Goal is the after): confirm the prediction, don't reconstruct
   it from memory.
   **If this answer is not `none`, capture it before closing the cycle** — the
   sentence is the deliverable, and an uncaptured one decays into a
   reconstruction. Evidence ref: under `one-spec-per-pr` this spec has exactly
   one PR by construction, so tag `pr:<n>` rather than a commit hash (a
   squash-merge destroys the branch commit you were looking at).
   — <answer>

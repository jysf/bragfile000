---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-088
  type: chore                      # epic | story | task | bug | chore
  cycle: verify                    # frame | design | build | verify | ship
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
  designed_at: 2026-09-17

insight:
  confidence: 0.95                 # UP from framing's 0.90. Both assertions now
                                   # exist as literals, both were run through the
                                   # real `just test-docs` against the real tree,
                                   # and each fired under every mutation in the
                                   # matrix and stayed green on its negative
                                   # control. Every framing number was re-derived
                                   # at `9f03680`; one moved (437 → 439 tracked
                                   # files) and nothing else did. The residual
                                   # 0.05 is scope, not doubt: `just test-docs`
                                   # is local-only, so neither guard gates CI.

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

## Re-measurement at design — every framing number, re-derived

Framing measured on 2026-09-15 at `main` = `6dda56a`. Design re-derived all of
it on 2026-09-17 at `main` = `9f03680`. **One number moved. Everything else
held**, including both of framing's probe hashes, which reproduced byte-exact
from their stated edits.

| Framing claim | Re-derived at `9f03680` | Verdict |
|---|---|---|
| M-4 — 52 `DEC-*.md`, 51 `decision` + 1 `reservation` | 52 / 51 / 1 | holds |
| M-4 — 65 historical `type:` additions, 64 + 1 | 64 + 1 = 65 | holds |
| M-4 — zero `analysis` / `recommendation` / `observation`, ever | zero | holds |
| M-1 — 8 whole-line tag additions in history | 8: `content` ×6, `invoke` ×2 | holds, **and is sharper** |
| M-1 — 0 additions outside `*.md` | 0 | holds |
| M-2 — `main` is clean today, **0 of 437** tracked files | 0 of **439** | **CORRECTION: 437 → 439** |
| M-3 — 4 files mention the tags inline, 8 lines | 4 files, 8 lines | holds |
| P-1 — pre `c9f625f1…`, post `7ba747a7…` | reproduced exactly, first try (M-B1) | holds |
| `Documentation assertions (distinct ids)` = 198 | 198 | holds |
| → 200 after this spec | **200, regenerated** (see below) | confirmed |
| Files modified = 3 | 3 | holds |

**CORRECTION — the tracked-file count is 439, not 437.** Framing measured 437 at
`6dda56a` and wrote *"4 of 439 files (tracked plus the two new spec files)"* for
its own branch. Both spec files are tracked on `main` now, so the single honest
number at `9f03680` is **439 tracked files, 4 of them carrying 8 inline
mentions**. The distinction framing drew between tracked and untracked no longer
applies, and this spec's assertion reads `git ls-files`, which is the tracked
set only.

**M-1 is sharper than framing stated it.** Framing ran the general form
`</[A-Za-z_:.-]+>` over the full history and reported the two names it found.
Re-run at design, the general form returns *exactly* those 8 additions and
nothing else — so the full historical population of whole-line closing tags in
this repo is **two tag names**, not two among many. That is what makes the
alternation in LD8 a measurement rather than a guess.

## §12(b) design-time pre-flight — both literals, run through the real harness

Per AGENTS.md §12, and specifically the refinement that the pre-flight must
exercise **behaviour**, not shape. `bash -n` was run, and is not the evidence.
What follows was produced by inserting the literal `Group AC` block into
`scripts/test-docs.sh`, applying the literal template narrowing, regenerating
the inventory block, and running `just test-docs` and `just inventory` against
the real tree. Everything was then restored from `/tmp` backups and the tree
verified byte-identical to `main`.

**Green on the clean tree**, with the simulated build state installed:

```
OK:   AC1
OK:   AC2

ALL OK: documentation-content assertions passed.
```

`OK:` lines **199 → 201**, distinct ids **198 → 200**, `SKIP:` 0, `FAIL:` 0,
exit 0. (`OK:` lines and distinct ids differ by one because `S3` double-emits,
and the `OK`/`SKIP` split for `S3` depends on whether the `claude` CLI is
installed — which is why the inventory counts ids, not lines. Measured on a
machine that has it.)

**The inventory, regenerated rather than predicted.** `just inventory` against
the simulated build state, diffed against the block on `main`:

```
17c17
< | Documentation assertions (distinct ids) | 198 | `scripts/test-docs.sh`, run by `just test-docs` |
---
> | Documentation assertions (distinct ids) | 200 | `scripts/test-docs.sh`, run by `just test-docs` |
```

**Exactly one row moves.** No other row is touched — not `Go test functions`,
not `Decision records`, not the two `Y4` pins. This is the part AGENTS.md §9
part (c) says not to reason about: it was regenerated and diffed.

**The literals were then re-extracted from this file and the build rebuilt from
them alone**, to close the literal-artifact contract rather than assume it: both
extracts diff byte-identical against what was pre-flighted, and the rebuild
reproduces `OK: 201 / FAIL: 0 / 200 distinct ids` and the same `--numstat`. What
build transcribes is what design ran.

**Diff size of the simulated build state**, `git diff --numstat`:

```
12      3       decisions/_template.md
1       1       docs/engineering-practices.md
136     0       scripts/test-docs.sh
```

### The mutation matrix

Every row: the exact edit (§12 — *a hash is not a reproduction*), the content
hash **before and after**, and the gate's verdict. The hash was checked **before
the gate ran** in every case, by a helper that refuses to run the gate on an
unmoved hash (the shape SPEC-087 verify built). Every restore was `cp` from a
`/tmp` backup, never `git checkout`, and every restored hash was confirmed equal
to its pre value. Hashes are `shasum -a 256`, first 12 hex digits.

| id | Target | The exact edit | pre → post | Verdict | Collateral |
|---|---|---|---|---|---|
| **M-A0** | `decisions/_template.md` | restore the pre-SPEC-088 line 6: `# decision \| analysis \| recommendation \| observation \| reservation` | `4b99dc8ca8b7` → `89a999aee134` | **FAIL AC1**, naming `analysis`, `observation`, `recommendation` | none |
| **M-A1** | `decisions/_template.md` | `sed -i '' '6s/# decision \| reservation/# decision \| analysis \| reservation/'` | `4b99dc8ca8b7` → `0cd64abdbe46` | **FAIL AC1** — direction 1 | none |
| **M-A2** | `decisions/_template.md` | `sed -i '' '6s/# decision \| reservation/# decision/'` | `4b99dc8ca8b7` → `5e626d725525` | **FAIL AC1** — direction 2 | none |
| **M-A3** | `decisions/_template.md` | `sed -i '' '6s/.*/  type: decision/'` — strip the comment entirely | `4b99dc8ca8b7` → `e98cbf8dccad` | **FAIL AC1** — vocabulary floor | none |
| **M-A4** | `scripts/inventory.sh` | `sed -i '' -e '81s/insight\.type: /type: /' -e '84s/insight\.type: /type: /'` | `e2db95583a9b` → `e98af3fd7306` | **FAIL AC1** — rows floor | **X3 also fires** — expected, the emitted table changed |
| **M-A5** | `scripts/inventory.sh` | `sed -i '' '84s/insight\.type: reservation/the tombstone type/'` | `e2db95583a9b` → `386b6e5a5883` | **FAIL AC1** — `reservation` offered, no row | **X3 also fires** |
| **M-A6** | `decisions/_template.md` | `sed -i '' '6s/# decision \| reservation/# decision \| frobnicate \| reservation/'` | `4b99dc8ca8b7` → `7b564199a0e1` | **FAIL AC1**, naming `frobnicate` | none |
| **M-B1** | `NEXT-SESSION-PROMPT.md` | append one line whose entire content is the closing `content` tag: `printf '</content>\n' >> NEXT-SESSION-PROMPT.md` | `c9f625f17b05` → `7ba747a741fb` | **FAIL AC2**, one hit at `NEXT-SESSION-PROMPT.md:203` | none |
| **M-B3** | `docs/engineering-practices.md` | append one line: two spaces, then the closing `invoke` tag | `78d5e4fb4c82` → `1537fe6df1d8` | **FAIL AC2**, hit at `docs/engineering-practices.md:309` | none |
| **M-B4** | new file `docs/spec088-probe.md` | write two lines — `draft notes`, then the closing `content` tag alone — and `git add` it | absent → `c161a037cb53` | **FAIL AC2**; the message's file count moves **439 → 440** | none |
| **M-B6** | `projects/PROJ-004-story-surface/specs/.gitkeep` | `rm` it, leaving the deletion unstaged | *not a content mutation* — see below | **FAIL AC2**, with grep's own `No such file or directory` in the message | none |
| **M-B2** | a throwaway `git init` directory | run AC2's body where `git ls-files` is empty | *n/a* | **FAIL AC2** — *"listed no tracked files"* | `AC1` also fails (no `scripts/inventory.sh` there) |

Three notes the matrix cannot carry in a cell:

- **M-A5 is the sharp one.** After it, `scripts/inventory.sh` line 50 still
  contains the string `insight.type: reservation` — in a comment. An `AC1` that
  grepped the script's **source** for the vocabulary would read that comment and
  pass the mutant. `AC1` runs the script and parses what it **emits**, so it
  fails. This is the same blindness SPEC-087 measured in `Z7` (M-1 there), one
  file over.
- **M-B6's observable is not a content hash**, because the mutation is the
  file's absence. The pre/post pair is `exists` → `does not exist`, with
  `git ls-files` still listing the path; the restore is `cp` from
  `/tmp/spec088-M-B6-backup`. It is included because it is the only probe that
  exercises LD9, and because the state it creates — a tracked path deleted from
  the worktree without staging the deletion — is one an agent produces routinely.
- **M-B5 was attempted and rejected as a probe, and the rejection is the
  finding.** `chmod 000` on a tracked file does make the gate red, but not via
  `AC2`: `assert_word_count_band` runs first, `wc -w` fails on the unreadable
  path, and `set -eu` kills the whole script with **exit 2 and zero `FAIL:`
  lines**. A red gate with no failing assertion is exactly the shape that reads
  like an environment problem. Do not use `chmod` to probe this harness.

## Locked design decisions

Each one names the Failing Test that fails without it (AGENTS.md §9). Every
pairing below was **executed**, not asserted — the ids in the right-hand column
are rows of the matrix above.

| # | Decision | Fails without it |
|---|---|---|
| **LD1** | `decisions/_template.md` advertises exactly two `insight.type` values: `decision \| reservation`. Framing's fork **(b)**, unchanged. | FT-1 / **M-A0** |
| **LD2** | The vocabulary guard is **derived on both sides and fails in both directions**: a template value with no emitted row fails, and an emitted row whose type the template never offers fails. | FT-2 / **M-A1**, FT-3 / **M-A2** — a one-directional assertion passes M-A2 |
| **LD3** | The row set comes from **running** `scripts/inventory.sh` and parsing its **output** — never from its source text, never from a re-implementation of its filters. | FT-6 / **M-A5** |
| **LD4** | `AC1` carries **two independent floors**, evaluated before the comparison: an empty parsed vocabulary fails, and an empty emitted row set fails. Neither is satisfied by the other's absence. | FT-4 / **M-A3**, FT-5 / **M-A4** |
| **LD5** | The vocabulary is whatever the template's own `  type:` line advertises — split on `\|`, every non-empty token counts. **No allow-list of known values**, which would be LD1's literal pin wearing a loop. | FT-7 / **M-A6** (`frobnicate` appears on no list anywhere and is still named) |
| **LD6** | `AC2`'s scope is **`git ls-files`**, not an enumerated list of files or directories. | FT-11 / **M-B4** (a brand-new tracked file fires it with zero harness edits; the message's count moves 439 → 440), FT-12 / **M-B2** |
| **LD7** | The pattern is **whole-line anchored**: `^[[:space:]]*</(content\|invoke)>[[:space:]]*$`. Inline mentions in backticks are permitted by construction, not by an exception list. | FT-8 (green **while** 4 tracked files carry 8 inline mentions), FT-9 / **M-B1**, FT-10 / **M-B3** |
| **LD8** | The alternation names **exactly the two tag names M-1 found** — `content` and `invoke` — not a general closing-tag form. | FT-9 uses `content`, FT-10 uses `invoke`; dropping either name turns one probe green |
| **LD9** | `AC2` captures **stderr** into the variable it tests for emptiness. The assertion is *"the sweep produces no output"*, not *"grep found no match"*. | FT-13 / **M-B6** |
| **LD10** | The ids are `AC1` and `AC2` — **literal and quoted**, in a new `Group AC` placed immediately before `# ===== finalise =====`. | FT-14 / **X3**: `inventory.sh` counts distinct *quoted literal* ids, so a loop-generated id runs but is never counted and the row stays at 198 (measured at SPEC-085 design: 198 emitted vs 195 counted) |
| **LD11** | `docs/engineering-practices.md`'s inventory block is **regenerated with `just inventory` and pasted whole**. This spec deliberately does **not** embed it as a literal to transcribe. | FT-14 / **X3** |

`AC1` and `AC2` are **one assertion each**, not one per direction: splitting
either one moves the derived row to 201, splitting both moves it to 202, and
neither buys coverage — a single failure message can name both directions, and
does.

### Rejected alternatives (build-time)

Locked here so build does not re-open them.

1. **`assert_contains_literal "AC1" "decisions/_template.md" "decision | reservation"`.**
   The literal pin SPEC-087 spent a cycle removing from `Y3`. It needs a
   hand-edit the first time a legitimate type is added, and it cannot fail in
   direction 2 at all — M-A2 and M-A5 both pass against it.
2. **Teach `scripts/inventory.sh` three more rows** (framing's fork (a)).
   Rejected at framing; restated here because it is the obvious build-time
   shortcut when `AC1` goes red. A 20-row user-facing table would grow to 23 to
   report a population of zero, three times.
3. **Grep `scripts/inventory.sh`'s source instead of running it.** Measured
   wrong by M-A5.
4. **A general closing-tag pattern, `</[A-Za-z_:.-]+>`.** It also has zero hits
   today — but no tracked file is XML, HTML or SVG (234 `.md`, 149 `.go`, 17
   `.sh`, 8 `.yaml`, 6 `.json`, 4 `.yml`, 4 `.sql`, 2 `.csv`, 9 `.gitkeep`, and
   five singletons). The first one added makes a correct document red, and a
   guard that fires on a correct file gets disarmed by habit — the
   bump-the-number-until-green failure SPEC-079 LD5 refused. Widen it when a
   third tag name is actually observed.
5. **A file allow-list for the four prose-mention files.** The membership of
   that set turned over inside a single day at framing (`STAGE-023` fell to 0,
   this file went to 3). It would go stale in one PR. The anchor is what makes
   the allow-list unnecessary.
6. **Narrowing the sweep to `*.md`, or an `--exclude-dir`.** M-1 measured 0
   whole-line tag additions outside `*.md` *so far*, which is not a reason to
   stop looking; AGENTS.md §9 records that BSD grep treats `--exclude-dir` as
   decorative; and the whole sweep costs 0.15s over 439 files. Scope everything.
7. **Suppressing the sweep's stderr with `2>/dev/null`.** Turns an unreadable or
   missing tracked path into a silent green (M-B6).
8. **`chmod 000` as the unreadable-file probe.** See M-B5 above: the harness
   dies at `assert_word_count_band` before `AC2` runs.
9. **Splitting into `AC1a`/`AC1b`/`AC2a`/`AC2b`.** Four ids instead of two moves
   the derived row to 202 for no additional coverage.

## Outputs

### Modified files (11 — design planned 3; 8 added at build by maintainer direction)

- **`scripts/test-docs.sh`** — `+136 / -0`. One new `# ===== Group AC — the
  harness guards (SPEC-088) =====` block, inserted immediately before
  `# ===== finalise =====`. Ids `AC1` and `AC2`, each quoted literally. The
  literal is in `## Notes for the Implementer`; transcribe it verbatim.
- **`decisions/_template.md`** — `+12 / -3`. Lines 6–13 replaced. The literal is
  in `## Notes for the Implementer`.
- **`docs/engineering-practices.md`** — `+1 / -1`. The inventory block
  **regenerated** with `just inventory` and pasted whole between the markers,
  never hand-edited. Exactly one row moves; if a second row moves, the
  regeneration is right and this spec is stale (AGENTS.md §9 part (a)).
- **Eight `decisions/DEC-*.md` files**, `+8 / -8` in total, one line each:
  `DEC-031`, `DEC-037`, `DEC-046`, `DEC-047`, `DEC-048`, `DEC-049`, `DEC-051`
  and `DEC-052`. **This is a maintainer-directed scope addition
  (2026-09-18), not part of design's plan.** Design found these files in the
  premise audit below and routed them out. Only the comment on each file's
  line 6 changes, and the line becomes byte-identical to the template's new
  line 6, `  type: decision                     # decision | reservation`.
  Every `type:` value stays `decision`. No decision's substance changes, and
  none is amended, so this is not a decision record emitted. There is
  deliberately **no** new test-docs assertion for it, because one would move
  the distinct-id count past the 200 this spec pins, which would be a design
  change. Acceptance criterion 8 checks the fold with a grep instead.

### Created files (0) · Decision records (0) · Go (none) · Database (none)

This spec emits no `DEC-*`. If build concludes it needs one, the next free
number is **`DEC-054`** — `DEC-050` is claimed by SPEC-086.

### The premise audit — grep for the value, not the idea

AGENTS.md §9 requires these greps to be **run at design** and their real hits
reconciled against this list. They were. Below is every hit, classified.

**Every count below is measured on `main` at `9f03680`** — i.e. `git grep <pat>
main`, not the working tree. This section's own prose adds hits for all four
patterns, so a count taken on this branch measures the audit, not the repo.

**`git grep -n '198' main`** — the derived count. **12** hits outside
`specs/done/`:

| Hit | Classification | Action |
|---|---|---|
| `docs/engineering-practices.md:43` | the derived row | **regenerated** — in Outputs |
| `NEXT-SESSION-PROMPT.md:28` — *"199 `OK:` lines / 198 distinct ids"* | a **dated snapshot** under a heading that reads *"re-derived 2026-09-14 — every number here will move"*, already stale in its own first bullet (`main` = `7c615f5`; it is `9f03680`) | **examined, not changed** — the orchestrator's working file, not a repo claim; editing it puts another cycle's bookkeeping in this PR. Raised with the maintainer. |
| `scripts/test-docs.sh:2019` — *"measured at design: 198 emitted vs 195 counted"* | a **dated historical measurement** from SPEC-085's design | correct as written after this spec; no edit |
| `STAGE-023:224` — *"(198 → 200) for the pair"* | framing's own prediction | **confirmed by the regeneration**; no edit |
| `SPEC-088` ×4 (lines 95, 235, 282, 350) | this file's own prose | confirmed at design |
| `DEC-049:153`, `guidance/questions.yaml:992`, `STAGE-023:619` | brag-corpus entry counts | **not this number** |
| `SPEC-090:60` — `internal/cli/add.go:198` | a source line number | **not this number** |

**`git grep -n '\b199\b' main`** — the `OK:`-line count, which moves 199 → 201.
Only real hit is `NEXT-SESSION-PROMPT.md:28`, the same dated snapshot. The other
hits are corpus counts in `docs/research/duckdb-federation-spike.md` and
references to PR #199.

**`git grep -n 'distinct id' main`** — 8 hits outside `specs/done/`. Beyond the
above:
`STAGE-021:149` (*"is 163 distinct ids, 171 after this spec"*) is a dated
SPEC-079-era statement; `SPEC-092:150` states that **its** change does not move
the row, which stays true and is unaffected by landing first.

**`git grep -nE 'analysis \| recommendation|recommendation \| observation' main`**
— the vocabulary string itself. This is the grep that found the one thing no
number-grep could reach, and it is the §9 rule working exactly as written:
searching for the *concept* finds the template; only searching for the *value*
finds the eight copies of it. **16 hits** — 9 in live `decisions/*.md` (8
`DEC-*` plus the template), 1 in `STAGE-023:654`, 1 in this file's framing text
at `:102`, and 5 inside archived `specs/done/` files that quote the template.

**`git grep -n 'insight\.type' main`** — the same surface by name rather than by
value. **25** hits outside `specs/done/`, every one classified:

- **Eight shipped `decisions/DEC-*.md` files carry a copy of the old vocabulary
  comment on their own line 6.** `DEC-031`, `DEC-037`, `DEC-046`, `DEC-047`,
  `DEC-048`, `DEC-049` carry `# decision | analysis | recommendation |
  observation`; `DEC-051` and `DEC-052` carry all five. **Examined, not
  changed.** Nothing derives from them: `AC1` reads `decisions/_template.md`
  only, and `inventory.sh`'s `grep -l '^  type: decision'` matches the value
  regardless of the trailing comment. The residual risk is a worse error
  message, not a missed error — an author who copies `DEC-052` and picks
  `analysis` still hard-fails `Z7` on the commit that adds the file, which is
  the pre-existing net and it still holds. Fixing eight records of other specs
  is a second change in one PR under `one-spec-per-pr`. **Routed to STAGE-023's
  backlog as an unowned candidate**, the same treatment framing gave SPEC-090's
  front matter. Raised with the maintainer.

  > **AMENDED at build, 2026-09-18. This is a maintainer-directed scope
  > addition, not a build discovery.** The maintainer answered the question
  > raised here by directing that the eight be folded into this build. In each
  > file only the comment on line 6 changed, and each line is now byte-identical
  > to the template's new line 6,
  > `  type: decision                     # decision | reservation`.
  > Every `type:` value stays `decision`, and the eight moved to `## Outputs`
  > as modified files. Two corrections to this bullet, both measured at build
  > on `main` = `7fa5cae`. First, the eight are hits of the **value** grep
  > above, not of `git grep -n 'insight\.type'`: their line 6 reads `type:`,
  > so this grep never matches them. Second, the STAGE-023 backlog entry named
  > here was never written. The route existed in this file's prose only.
  > The five other `decisions/` hits for these three words (DEC-011, 024, 027,
  > 038, 045) are English prose, not the vocabulary, and are **not** touched.
- `STAGE-023:652` — the routed finding this spec discharges. This stage's own
  convention for these three items is a `> **Status at … ship**` blockquote
  added at ship, not a rewrite of the finding. **No build edit.**
- `docs/engineering-practices.md:65`, `STAGE-021:210/433`, `STAGE-022:478-510`,
  `SPEC-093:87`, `scripts/test-docs.sh:1592/1612/1665/1832/1833/1872`,
  `docs/CONTEXTCORE_ALIGNMENT.md:17` — all describe the **two** counted types or
  the third-type hard fail. Every one stays true after the narrowing. **No
  edits.** In particular `Y3`'s comment at `:1631` (*"`decisions/_template.md`
  says so at the moment `type: reservation` is chosen"*) stays true: the
  tombstone paragraph is preserved verbatim in the new literal.

**No existing test's premise is inverted by this spec**, and no existing
assertion is deleted or rewritten. Both items are purely additive to
`scripts/test-docs.sh`.

## Acceptance Criteria

Numbers to diff against. Every one was measured at design against the real tree;
**re-derive, don't trust** (AGENTS.md §9 part (a)).

1. `just test-docs` exits **0** and prints `OK:   AC1` and `OK:   AC2`.
2. `OK:` lines **199 → 201**; `FAIL:` lines **0**; distinct ids **198 → 200**.
   (`SKIP:` is 0 on a machine with the `claude` CLI installed, 1 without — `S3`
   is the environment-dependent one, which is why the inventory counts ids.)
3. `just inventory` reports `| Documentation assertions (distinct ids) | 200 |`,
   and the regenerated block differs from `main`'s by **exactly that one row**
   (`diff` → `17c17`). If a second row moves, the regeneration is right.
4. `decisions/_template.md` line 6 is exactly
   `  type: decision                     # decision | reservation`.
5. Every mutation in the matrix reproduces its stated post-hash **from its
   stated edit**, turns its named assertion red, and restores to its pre-hash
   from a `/tmp` backup. The hash is checked **before** the gate runs.
6. The anchored sweep over the final tree exits **1**:
   `git ls-files -z | xargs -0 grep -nE '^[[:space:]]*</(content|invoke)>[[:space:]]*$' /dev/null`
7. All five gates green: `just test`, `just test-docs`, `just lint`,
   `gofmt -l .`, `go vet ./...`.
8. **Added at build (2026-09-18) with the maintainer-directed fold of the
   eight stale `DEC-*.md` comments.** This value-grep prints nothing and exits
   **1**:
   `grep -lE '^[[:space:]]+type:.*#.*(analysis|recommendation|observation)' decisions/DEC-*.md`
   Each of the eight changed lines is byte-equal to `decisions/_template.md`
   line 6. No `type:` value moves: 51 `decision` + 1 `reservation`, as on
   `main`. This criterion is checked by hand at build and verify. It is
   deliberately **not** a test-docs assertion, because a new id would move
   the distinct-id row past the 200 this spec pins.

## Failing Tests

There is no Go in this spec, so the failing tests are harness assertions and
their controls. **FT-1 and FT-8 are the two that must be green at the end; every
other FT is a positive control that must fire.** A control that does not fire is
the finding, not a formality — SPEC-089 shipped a guard covering two of five
keys with all 14 packages green.

Run each with the mutation applied, the hash confirmed moved **first**, and the
restore from `/tmp`.

| FT | Assertion | State | Expected |
|---|---|---|---|
| **FT-1** | `AC1` | narrowed template, `Group AC` installed | `OK:   AC1`. **Fails today**: on `main`'s five-value template it names `analysis`, `observation`, `recommendation` (M-A0). |
| **FT-2** | `AC1` | **M-A1** | `FAIL: AC1 … [decisions/_template.md offers 'analysis'; scripts/inventory.sh emits no row for it]` |
| **FT-3** | `AC1` | **M-A2** | `FAIL: AC1 … [scripts/inventory.sh emits a row for 'reservation'; decisions/_template.md never offers it]` |
| **FT-4** | `AC1` | **M-A3** | `FAIL: AC1: no insight.type vocabulary parsed out of decisions/_template.md …` |
| **FT-5** | `AC1` | **M-A4** | `FAIL: AC1: scripts/inventory.sh emitted no row naming an 'insight.type: <value>' …` (plus `X3`) |
| **FT-6** | `AC1` | **M-A5** | `FAIL: AC1 … offers 'reservation'; … emits no row for it` — while `inventory.sh:50` still contains the string in a comment (plus `X3`) |
| **FT-7** | `AC1` | **M-A6** | `FAIL: AC1 … offers 'frobnicate' …` |
| **FT-8** | `AC2` | clean tree | `OK:   AC2`, **while 4 tracked files carry 8 inline mentions of the two tags**. This is the negative control and it is load-bearing: an unanchored pattern fails it. |
| **FT-9** | `AC2` | **M-B1** | `FAIL: AC2 …` with one hit, `NEXT-SESSION-PROMPT.md:203` |
| **FT-10** | `AC2` | **M-B3** | `FAIL: AC2 …` with one hit, `docs/engineering-practices.md:309` — indented, and the other tag name |
| **FT-11** | `AC2` | **M-B4** | `FAIL: AC2 …`, and the message's file count reads **440**, not 439, with no harness edit |
| **FT-12** | `AC2` | **M-B2** | `FAIL: AC2: git ls-files listed no tracked files …` |
| **FT-13** | `AC2` | **M-B6** | `FAIL: AC2 …` carrying grep's own `No such file or directory` |
| **FT-14** | `X3` | block installed, page **not yet** regenerated | `FAIL: X3: inventory block is stale …`; green after `just inventory` is pasted whole |
| **FT-15** | the gate | final tree | exit 0, **201** `OK:`, **0** `FAIL:`, **200** distinct ids |

The three standing traps framing named, now discharged rather than restated:

- **A green suite is not evidence.** Thirteen controls above, each executed at
  design against the real `just test-docs`.
- **An assertion that enumerates its own scope by hand is the defect it guards
  against.** `AC2` derives from `git ls-files` (LD6, proved by M-B4); `AC1`
  derives from the template's own comment and the script's own emission (LD3,
  LD5, proved by M-A5 and M-A6).
- **A floor, or the assertion is vacuously true at the empty state.** Three
  floors, each with its own probe: M-A3, M-A4, M-B2. None borrows another
  assertion's.

## Implementation Context

### Decisions that apply

- None binding. This spec emits none. **If build concludes it needs a decision
  record, the next free number is `DEC-054` — `DEC-050` is claimed by SPEC-086
  and was already renumbered around once (#207).**

### Constraints that apply

- `one-spec-per-pr` — blocking. Items A and B ship in one PR under one spec id;
  C and D have their own ids for this reason. This is also why the eight
  stale-comment `DEC-*.md` files found in the premise audit were routed, not
  absorbed, **at design**. **AMENDED at build, 2026-09-18:** the maintainer
  directed that they be folded into this build. The change still falls under
  this spec's id. It is a comment-only edit to the vocabulary this spec
  narrows, it adds no assertion id, and it moves no inventory row. See
  `## Outputs` and `## Build Completion`.

### Prior related work

- **SPEC-087** (shipped) — made `Y3` derive, and proved by mutation M-1 that an
  assertion re-implementing the producer's own filter is blind by construction.
  Both new assertions read what the producer *emits*. M-A5 is that measurement,
  re-run one file over.
- **SPEC-089** (shipped) — routed item B here; two of the six affected files are
  its own, and its ship is where the sweep was first run.
- **PR #210, #211, #212** — the three most recent strippings, by hand.

### Out of scope (for this spec specifically)

- `Y4` (SPEC-092), id reservation (SPEC-093).
- ~~**The eight `DEC-*.md` files carrying a stale copy of the vocabulary
  comment.** Found at design, enumerated under `## Outputs`, routed to
  STAGE-023's backlog.~~ **Back in scope. The maintainer directed on
  2026-09-18 that they be folded into this build**, so they are now modified
  files under `## Outputs`. The STAGE-023 backlog entry this line routed them
  to was never written. STAGE-023 has no entry for them, so nothing there
  needed removing.
- Front-matter well-formedness. **Found at framing, not fixed:** `SPEC-090`'s
  front matter is unterminated on `main`. Different defect class, another spec's
  file, and folding it in would put two specs in one PR.
- Running the documentation assertions in CI. `just test-docs` is local-only
  (`.github/workflows/ci.yml` has `test`, `lint` and `coverage` jobs and no
  test-docs job). **Both new assertions therefore gate only what a human or an
  agent runs locally** — the honest scope of this spec's claim, and a reason not
  to overstate it at ship. Design did not need this resolved to proceed: the
  assertions are identical either way, and wiring the harness into CI is a
  change to the harness's *reach*, not to these two guards.

## Notes for the Implementer

Two literals. Transcribe both verbatim; both were run through the real tool at
design (§12(b) above).

### Literal 1 — `decisions/_template.md`, lines 5–22 after the change

Lines 6–13 of the current file are replaced by lines 2–18 below. Line 5 (`id:`)
is shown unchanged as an anchor, and **the alignment matters**: every `#` sits
at column 38, as it does everywhere else in this front matter.

```yaml
  id: DEC-XXX                        # stable, never reused
  type: decision                     # decision | reservation
                                     # `reservation` = a TOMBSTONE: a number claimed, not yet decided.
                                     # A tombstone MUST carry a `## This is not a decision` heading in
                                     # its body — scripts/test-docs.sh assertion Y3 counts tombstones by
                                     # that heading, deliberately NOT by this front-matter field, so the
                                     # two counts fail differently. Copy decisions/DEC-041-*.md.
                                     # THESE TWO, AND THE LIST IS NOT FREE-FORM. scripts/inventory.sh
                                     # emits one row per value above; test-docs assertion AC1 compares
                                     # this line against those emitted rows in BOTH directions, so a
                                     # value added here without its row fails, and a row without its
                                     # value here fails too. To add a third type, add its inventory row
                                     # in the SAME edit. A DEC-*.md carrying a value no row counts is
                                     # invisible on docs/engineering-practices.md and hard-fails Z7.
                                     # Narrowed from five values at SPEC-088: `analysis`,
                                     # `recommendation` and `observation` were advertised here, counted
                                     # by no row, and never used — 0 of 52 records, 0 of 65 historical
                                     # `type:` additions across every branch.
```

The tombstone paragraph is preserved **verbatim** — `Y3`'s comment cites it. The
two lines removed are SPEC-087's warning, which pointed at *"the STAGE-023 note
routed to SPEC-088"*; that route closes here, so the warning is replaced by the
rule the new assertion enforces.

### Literal 2 — `scripts/test-docs.sh`, a new `Group AC`

Insert **immediately before** `# ===== finalise =====` (line 2067 on `main` at
`9f03680`), after `AB4`. Nothing else in the file changes.

```bash
# ===== Group AC — the harness guards (SPEC-088) =====
#
# Two ways a session can quietly corrupt this repo's OWN documents, each
# measured before it was guarded. Neither is reachable from CI: `just
# test-docs` is local-only (.github/workflows/ci.yml runs test, lint and
# coverage and has no test-docs job), so both assertions gate what a human or
# an agent runs locally. That is the honest scope, and it is why both failure
# messages name the remedy instead of pointing at a build log.

# AC1 — the `insight.type` vocabulary decisions/_template.md advertises must be
# exactly the set scripts/inventory.sh emits a row for. DERIVED on both sides,
# and deliberately NOT `assert_contains_literal ... "decision | reservation"`:
# a literal pin is the anti-pattern SPEC-087 spent a whole cycle removing from
# Y3, and it would need a hand-edit the first time a legitimate value is added.
#
# WHAT IT CATCHES. Until SPEC-088 the template offered five values and
# inventory.sh emitted rows for two. A DEC-*.md carrying one of the other three
# is counted by NEITHER row, vanishes from docs/engineering-practices.md, and
# hard-fails Z7 — measured at SPEC-087 with a `type: analysis` stub, simulation
# S-3: "the inventory covers 49 of 50". The template is where an author picks
# the value, so the template is where the trap had to close. M-4 at SPEC-088
# framing, re-derived at design: those three values have 0 instances in 52
# records and 0 in 65 historical `type:` additions across every branch.
#
# IT FAILS IN BOTH DIRECTIONS, which is what makes it a vocabulary check rather
# than a spell-check: a template value with no row fails, and a row for a type
# the template never offers fails. A future spec that legitimately adds a third
# type pays one edit to each file in the same commit and pays this assertion
# nothing — no number here moves.
#
# THE TWO FLOORS. Y3 shipped without one and went green against an emptied
# decisions/ while silently borrowing Z7's, so both vacuous states are named
# here explicitly and checked BEFORE the comparison. (a) An unparsed template
# yields the empty vocabulary, and the empty set is trivially covered by any
# set of rows. (b) An inventory emitting no `insight.type:` row at all yields
# the empty row set, for the same reason. Neither floor can be satisfied by the
# other's absence.
if [ ! -x scripts/inventory.sh ]; then
    fail "AC1" "scripts/inventory.sh is missing or not executable"
elif [ ! -f decisions/_template.md ]; then
    fail "AC1" "decisions/_template.md does not exist — the vocabulary an author picks from has no source"
else
    ac1_out=$(./scripts/inventory.sh)
    ac1_vocab=$(awk '
        /^  type: / {
            if (sub(/^[^#]*#[[:space:]]*/, "") == 0) exit
            n = split($0, parts, "|")
            for (i = 1; i <= n; i++) {
                v = parts[i]
                gsub(/^[[:space:]]+|[[:space:]]+$/, "", v)
                if (v != "") print v
            }
            exit
        }
    ' decisions/_template.md | sort -u)
    ac1_rows=$(printf '%s\n' "$ac1_out" \
        | grep -oE 'insight\.type: [A-Za-z][A-Za-z0-9_-]*' \
        | sed 's/^insight\.type: //' | sort -u)
    if [ -z "$ac1_vocab" ]; then
        fail "AC1" "no insight.type vocabulary parsed out of decisions/_template.md: either there is no '  type: ' line, or that line carries no '# a | b' comment. An unparsed template must not pass — the empty set is covered by any set of rows, so the comparison below would assert nothing at all."
    elif [ -z "$ac1_rows" ]; then
        fail "AC1" "scripts/inventory.sh emitted no row naming an 'insight.type: <value>'. The decisions rows were renamed or removed, and an absent row must not pass silently (same reason Y3 and Z7 reject a non-numeric inv_row result)."
    else
        ac1_bad=""
        while IFS= read -r ac1_v; do
            [ -n "$ac1_v" ] || continue
            printf '%s\n' "$ac1_rows" | grep -qxF -- "$ac1_v" \
                || ac1_bad="$ac1_bad [decisions/_template.md offers '$ac1_v'; scripts/inventory.sh emits no row for it]"
        done <<<"$ac1_vocab"
        while IFS= read -r ac1_r; do
            [ -n "$ac1_r" ] || continue
            printf '%s\n' "$ac1_vocab" | grep -qxF -- "$ac1_r" \
                || ac1_bad="$ac1_bad [scripts/inventory.sh emits a row for '$ac1_r'; decisions/_template.md never offers it]"
        done <<<"$ac1_rows"
        if [ -z "$ac1_bad" ]; then
            ok "AC1"
        else
            fail "AC1" "decisions/_template.md and scripts/inventory.sh disagree about the insight.type vocabulary:$ac1_bad. A value with no row is invisible on docs/engineering-practices.md and hard-fails Z7 on first use; a row with no value counts a type no author is offered. Add both halves in one edit, or narrow the template."
        fi
    fi
fi

# AC2 — no tracked file carries a closing tool-call tag ALONE ON A LINE. A
# session writing a file has left its own syntax inside the artifact six times
# across two projects and 35 days (SPEC-088 M-1: 8 additions, 4 commits, 6
# files, 0 outside *.md, and only two tag names in the whole history —
# `content` and `invoke`). One instance sat in DEC-046 for ~66 PRs, inside a
# record four later specs cite. Every one was caught by a human reading the
# file; five gates and ~200 assertions never looked.
#
# THE SCOPE IS DERIVED — `git ls-files`, not a list. An assertion that
# enumerates its own scope by hand is the same defect one level up: SPEC-089's
# duplicate-header guard named two of five keys with all 14 packages green.
#
# THE ANCHOR IS THE WHOLE POINT. Four tracked files legitimately NAME these
# tags in prose, 8 lines between them, and the membership of that set turned
# over inside a single day during framing — so a file allow-list would have
# gone stale in one PR. Every legitimate mention writes the tag inline, in
# backticks or mid-sentence; every leaked one is alone on a line. The
# `^[[:space:]]*...[[:space:]]*$` anchor separates the two, measured rather
# than argued: silent on a clean tree WITH all four prose files present, and
# red on one appended line (framing's P-1, reproduced at design from its
# stated edit as M-B1, plus an indented `invoke` variant as M-B3).
#
# WHY THE TWO NAMES AND NOT `</[A-Za-z_:.-]+>`. The general form also has zero
# hits today, but no tracked file is XML, HTML or SVG — 234 .md, 149 .go, 17
# .sh, and no markup among them. Add one and a legitimately indented closing
# tag trips a guard that has nothing to say about it; a guard that fires on a
# correct file gets disarmed by habit. Widen this when a third tag name is
# actually observed — M-1 says there have only ever been two.
#
# STDERR IS CAPTURED DELIBERATELY. The assertion is "the sweep produces no
# output", not "grep found no match": a tracked path grep cannot read turns
# the gate red instead of passing quietly on an empty result.
#
# THE FLOOR: `git ls-files` returning nothing — outside a work tree, or in a
# repo with nothing tracked. Silence over zero files is not evidence. Cost of
# the sweep, measured at framing: 0.15s over the whole tree.
ac2_pat='^[[:space:]]*</(content|invoke)>[[:space:]]*$'
if ! command -v git >/dev/null 2>&1; then
    fail "AC2" "git is not installed — the sweep's scope is 'git ls-files' and cannot be derived without it"
else
    ac2_files=$(git ls-files | wc -l | tr -d ' ')
    if [ "$ac2_files" -lt 1 ]; then
        fail "AC2" "git ls-files listed no tracked files — the sweep would report clean while reading nothing at all"
    else
        ac2_hits=$(git ls-files -z | xargs -0 grep -nE "$ac2_pat" /dev/null 2>&1 || true)
        if [ -z "$ac2_hits" ]; then
            ok "AC2"
        else
            fail "AC2" "the anchored sweep over $ac2_files tracked files is not silent. Either a closing tool-call tag is alone on a line — session syntax that leaked into the artifact, so delete the line; a document that needs to NAME the tag writes it inline in backticks, which this assertion allows — or the sweep itself could not read a tracked path, which is also not a pass. Output:
$ac2_hits"
        fi
    fi
fi
```

### The order to work in

1. Transcribe Literal 2. `bash -n scripts/test-docs.sh`.
2. Run `just test-docs`. Expect **`FAIL: AC1`** naming three values, and
   **`OK:   AC2`** — `AC1` fails first, for the reason the spec says it should
   (FT-1's "fails today" leg). Do not skip this: a fail-first that reports the
   wrong reason is a spec defect, caught at its cheapest moment.
3. Transcribe Literal 1. `just test-docs` → `AC1` and `AC2` both green,
   `X3` red.
4. `just inventory`, paste the whole block between the
   `<!-- inventory:begin` / `<!-- inventory:end` markers, **no blank lines
   inside them**. `X3` green. **Re-derive the row's value; do not type 200** —
   if the regeneration disagrees with this spec, the regeneration is right.
5. Run the mutation matrix. Back up to `/tmp` first, confirm each hash moved
   **before** running the gate, restore with `cp`.
6. All five gates.

### Traps, all of them paid for already

- **The self-check that works on both zsh and BSD tools.** `FILES=$(…); grep …
  $FILES` does **not** word-split in zsh — the filenames arrive as one argument
  and the sweep reports *"No such file or directory"* with **exit 2**, which
  reads like a pass. BSD `xargs` has no `-a`, and fails the same silent way. Use
  `git ls-files -z | xargs -0 grep …`, or `tr '\n' '\0' < list | xargs -0 grep
  -nE …` for an ad-hoc list. **Read grep's exit code: 1 means clean.**
- **`/dev/null` before the file list is why the hits carry filenames.** With a
  single-file batch, `grep -n` omits the filename prefix; the sentinel argument
  forces it and can never match.
- **`$?` after a pipe is the last command's status.** `xargs` returns 123 when
  grep exits 1, which is why the assertion tests the *output*, not the code.
- **Write the tag forms inline in backticks, never alone on a line**, or `AC2`
  fails on the very commit that introduces it. This file carries them on several
  lines and passes the anchored sweep — verified at design, after writing.
- **Quote `--include` globs.** zsh expands unquoted ones and the search silently
  does not run.
- **`just advance-cycle` strips the inline enum comment** from the line it
  rewrites (`update_frontmatter_scalar`'s `sub(/:[[:space:]]*.*$/, …)`). It did
  so to this file's `cycle:` line at design and the comment was restored by
  hand. Not this spec's to fix; check the line after running it.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

*Built 2026-09-18 in a fresh session, from `main` = `7fa5cae`.*

- **Branch:** `build/spec-088-harness-guards`
- **PR (if applicable):** none opened by build. The orchestrator opens it.
- **All acceptance criteria met?** **No: 7 of 8 in full, and AC5 in part.**
  AC5 holds for every verdict and every restore, but for only **9 of the 10**
  post-hashes: M-A0's stated hash does not reproduce from its stated edit
  (see Deviations). AC1–4 and AC6–8 hold as written.
- **New decisions emitted:** none. The eight `DEC-*.md` edits change a comment
  and no decision.
- **Deviations from spec:**
  - **Maintainer-directed scope addition (2026-09-18), not a build
    discovery: the eight stale `DEC-*.md` comments.** Design found these
    files in its premise audit and routed them out. The maintainer then
    directed that they be folded into this build. `DEC-031`, `037`, `046`,
    `047`, `048`, `049`, `051` and `052` each changed on line 6 only, and each
    line is byte-equal to the template's new line 6.
    `git diff --shortstat main -- 'decisions/DEC-*.md'` → 8 files, +8/−8. The
    value-grep in AC8 prints nothing (exit 1). Every `type:` value stays
    `decision`: 51 `decision` + 1 `reservation`, the same as on `main`. The
    five English-prose hits for these three words (DEC-011, 024, 027, 038,
    045) were not touched. The regenerated inventory is byte-identical before
    and after the fold. No test-docs id was added, so the count stays at 200.
    `## Outputs`, the premise audit, *Constraints that apply* and
    *Out of scope* are amended to match. Each amendment is marked, and the
    design text it replaces is still visible.
  - **AC5 — M-A0's post-hash did not reproduce; reported, not resolved.** I
    ran the stated edit, *"restore the pre-SPEC-088 line 6"*, as
    `sed -i '' '6s/# decision | reservation$/# decision | analysis | recommendation | observation | reservation/' decisions/_template.md`.
    The mutant diff was exactly one line, `6c6`, reading
    `# decision | analysis | recommendation | observation | reservation`.
    Post-hash **`accb16924cb6`**, stated **`89a999aee134`**. The verdict
    reproduced: `FAIL: AC1` naming `analysis`, `observation` and
    `recommendation`, with nothing else red. The restore returned
    `4b99dc8ca8b7`. Per the §12 rule that *the hash does not identify the
    edit*, I did **not** search for an edit that produces the stated hash.
    Whether design's M-A0 changed more than line 6 is for verify to settle.
  - **Cycle bookkeeping.** `main` carried `cycle: design`, because the design
    commit left the advance to build. Build ran one
    `just advance-cycle SPEC-088 verify`, as directed, which strips the inline
    enum comment. The comment was restored by hand.
- **What build measured, for verify to diff against:**
  - **Fail-first, step 2** (Literal 2 installed, Literal 1 not):
    `FAIL: AC1: decisions/_template.md and scripts/inventory.sh disagree about the insight.type vocabulary: [decisions/_template.md offers 'analysis'; scripts/inventory.sh emits no row for it] [decisions/_template.md offers 'observation'; scripts/inventory.sh emits no row for it] [decisions/_template.md offers 'recommendation'; scripts/inventory.sh emits no row for it]. …`
    and `OK:   AC2`. `X3` was also red, which is FT-14's state: the distinct-id
    count moves the moment Literal 2 lands. Totals: 198 `OK:`, 2 `FAIL:`.
    Step 3 (both literals installed): `AC1` and `AC2` green, `X3` red, 199
    `OK:`, 1 `FAIL:`.
  - **Literal fidelity.** Both literals were extracted from this file with
    `awk` (the first fenced block under each `### Literal` heading) and
    written in by `head`/`cat`/`tail`, not retyped. Afterwards, the installed
    `Group AC` block (from its heading down to the blank line before
    `# ===== finalise =====`) diffs **byte-identical** against Literal 2, all
    135 lines. `test-docs.sh` with that block deleted diffs identical to
    `main`'s, with one hunk, `@@ -2066,0 +2067,136 @@`. Template lines 5–22
    diff **byte-identical** against Literal 1, all 18 lines, and every `#`
    sits at column 38. Numstat: `136/0`, `12/3`, `1/1`, as design's
    simulation predicted.
  - **Inventory.** `just inventory` was regenerated and pasted whole, with no
    blank lines inside the markers. Diffed against `main`'s block, exactly one
    row changed, `17c17`: `Documentation assertions (distinct ids)` went from
    198 to 200. It was regenerated again after the DEC fold, and nothing
    else moved.
  - **Mutation matrix.** Each target was backed up to `/tmp` and hashed.
    **The gate was refused unless the hash had moved.** Each run printed the
    mutant's own `diff`, and the target was restored with `cp -p` and
    re-hashed. `git status --porcelain` was also compared before and after
    each probe, to confirm nothing else changed.

    | id | stated post | reproduced post | verdict | restored |
    |---|---|---|---|---|
    | M-A0 | `89a999aee134` | **`accb16924cb6`** ✗ | `FAIL: AC1`, names the 3 values | `4b99dc8ca8b7` = pre |
    | M-A1 | `0cd64abdbe46` | `0cd64abdbe46` | `FAIL: AC1`, offers `'analysis'`, no row | = pre |
    | M-A2 | `5e626d725525` | `5e626d725525` | `FAIL: AC1`, row for `'reservation'`, never offered | = pre |
    | M-A3 | `e98cbf8dccad` | `e98cbf8dccad` | `FAIL: AC1`, vocabulary floor | = pre |
    | M-A4 | `e98af3fd7306` | `e98af3fd7306` | `FAIL: AC1`, rows floor, plus `X3` | `e2db95583a9b` = pre |
    | M-A5 | `386b6e5a5883` | `386b6e5a5883` | `FAIL: AC1`, offers `'reservation'`, no row, plus `X3`. Line 50 still carried the string during the probe. | = pre |
    | M-A6 | `7b564199a0e1` | `7b564199a0e1` | `FAIL: AC1`, names `'frobnicate'` | = pre |
    | M-B1 | `7ba747a741fb` | `7ba747a741fb` | `FAIL: AC2`, one hit, `NEXT-SESSION-PROMPT.md:203` | `c9f625f17b05` = pre |
    | M-B3 | `1537fe6df1d8` | `1537fe6df1d8` | `FAIL: AC2`, one hit, `docs/engineering-practices.md:309` | `78d5e4fb4c82` = pre |
    | M-B4 | `c161a037cb53` | `c161a037cb53` | `FAIL: AC2`, the message reads **440** tracked files | absent again, 439 tracked |
    | M-B6 | exists → absent | exists → absent | `FAIL: AC2`, carrying grep's `No such file or directory` | `e3b0c44298fc` = pre |
    | M-B2 | n/a | n/a | `FAIL: AC2`, *"listed no tracked files"*, plus `AC1` (no `inventory.sh`) | throwaway dir removed |

    Each single-assertion probe went red with only its target assertion: 199
    `OK:` / 1 `FAIL:`. M-A4 and M-A5 went 198 / 2, and their second failure is
    `X3`, as the matrix states. M-B2 ran the installed `Group AC` block, with
    the harness's `ok`/`fail` helpers, inside a fresh `git init` directory.
  - **Negative control FT-8, re-counted.** `git grep -cE` for the two
    closing-tag forms at `7fa5cae` finds **4 files and 9 lines**, not the 8
    lines design measured at `9f03680`. The design commit added a fourth line
    to this file. `AC2` stays green with every one of them present.
  - **Gates.** `just test` exit 0 (14 packages `ok`); `just test-docs` exit 0
    with **201 `OK:`, 0 `FAIL:`, 0 `SKIP:`, 200 distinct ids** (the `claude`
    CLI is installed); `just lint` 0 issues; `gofmt -l .` empty;
    `go vet ./...` clean. The anchored sweep of AC6 exits 1.
- **Follow-up work identified:**
  - None new for the backlog. Verify owns M-A0: decide what its stated edit
    means, then re-pin either the edit or the hash.

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — M-A0 was the only unclear part. It is the one matrix row whose edit is
   prose (*"restore the pre-SPEC-088 line 6"*) rather than a runnable command,
   and it is the one row whose hash did not reproduce. The verdict held, so
   the cost was one hash and no delay. Separately, step 2 of *The order to work
   in* lists only `AC1` and `AC2` as expected, while `X3` is also red at that
   point. FT-14 covers that, so it was not a defect, but it is a sentence
   short.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No constraint was missing. One route existed only in prose: the eight
   `DEC-*.md` files were "routed to STAGE-023's backlog", but no STAGE-023
   entry was ever written. That is the failure mode SPEC-093 exists to
   mechanise. It became moot when the maintainer folded the files in.

3. **If you did this task again, what would you do differently?**
   — I would have the probe helper print the mutant's own `diff` from the
   start, as it did here, and I would ask design for the same: pin a runnable
   command, or better the diff itself, in every matrix row. The recorded diff
   is what made M-A0's mismatch reportable without hunting, because it shows
   exactly what the stated edit produced.

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

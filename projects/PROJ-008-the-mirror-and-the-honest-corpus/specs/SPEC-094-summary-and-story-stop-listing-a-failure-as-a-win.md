---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-094
  type: story                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
                                   # Created at SPEC-086 design (2026-09-18) to
                                   # CLAIM the id in the same edit as the STAGE-023
                                   # line that routes work here. It is not framed:
                                   # it carries the evidence and the posture
                                   # DEC-050 already decided, and owes its own
                                   # framing pass.
  blocked: false
  priority: high                   # every failure written before this ships is
                                   # listed as a win on every `story` profile
                                   # (`summary`: SPEC-095)
  complexity: M                    # framed 2026-09-22: `story` only, after the
                                   # `summary` half split to SPEC-095. Held at M
                                   # by a decision record and a mechanism
                                   # measured against a consuming model.

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-09-18

references:
  decisions:
    - DEC-050                      # rows 3 and 4 are this spec's posture; it
                                   # implements them, it does not re-decide them
    - DEC-029                      # profiles are data — why story's mechanics
                                   # are a decision, not a line of code
    - DEC-014                      # the envelope, incl. part 4's empty-state rule
    - DEC-048                      # a count names what it counted
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-086                     # split this out; implements DEC-050 rows 1-2
    - SPEC-095                     # the `summary` half (DEC-050 row 3), split
                                   # out at this spec's framing
---

# SPEC-094: story stops listing a failure as a win

> **Scope changed at framing (2026-09-22).** This spec is now **`story` only**
> (DEC-050 row 4). The `summary` half (row 3) is **SPEC-095**, split along a
> defect shape this file had mismeasured. See *Framing → The split*. The
> filename still says *summary and story*; it is left as is because the id,
> not the slug, is what every reference uses.

## Context

> **Cycle: frame.** This file exists so the id is claimed by a file rather
> than reserved in prose (SPEC-087 LD6). It was created at SPEC-086's design
> in the same edit as the STAGE-023 backlog line that points here. Nothing
> below is designed.
>
> Measured 2026-09-18 against a frozen copy of the live corpus: 606 entries,
> 4 `type: failed` (ids 420, 433, 465, 473), all four with an impact. Binary
> built from `main` at `3201f50`.

This is the other half of SPEC-086's Fork B. SPEC-086 was split along the
defect shape that framing measured:

- **`impact` and `wrapped` could not represent a failure as one,** in either
  format. DEC-028's 4-key projection carries no `type`. SPEC-086 fixes those
  two surfaces.
- **`summary` and `story` already carry the data and drop it only in
  markdown.** That is this spec.

**The posture is decided.** DEC-050 rows 3 and 4 state it, and this spec
implements it:

| DEC-050 row | Surface | Posture |
|---|---|---|
| 3 | `brag summary` | A section, pulled out of `## Highlights`: `## What didn't work`, rendered only when it has an entry. JSON key `failures_by_project`, because highlights group by project. `By type` is already honest and stays. |
| 4 | `brag story` | Two invariants, and only those: a failure is **never rendered as a win**, and **never silently dropped**. A candid profile (`manager`, `me`) labels its failures where it lists them. A promotional profile (`candor: promotional`, today `exec` and `skip`) **may omit them, but only with a visible note in its output** — a count of the omitted failures is the obvious form. **The mechanism is this spec's to design.** |

## What was measured

**`summary`** (`brag summary --range month`):

- `## Summary → By type` prints `- failed: 4`, and JSON `counts_by_type`
  carries `"failed": 4`. That part is honest.
- `## Highlights` lists all four as `- <id>: <title>`, with no type. The JSON
  highlight entry is `["id","title"]`, which is 2 keys, narrower than
  `impact`'s 4.

**`story`** (`--year`, all four bundled profiles):

- **Every profile** renders each failure as `- ★ <id>: <title>`. Measured
  with `grep -c -E '(473|465|433|420): '`: `exec` 4, `skip` 4, `manager` 4,
  `me` 4.
- The markdown carries no type. The JSON does:
  `.threads[].beats[] | select(.type=="failed")` returns `[420, 433, 465,
  473]` under `--audience exec`.
- `exec` and `skip` are `candor: promotional`, and their directives tell a
  model to promote the beats. `manager` and `me` are `candid`.

## The decision this spec still owes

DEC-050 row 4 leaves `story` two questions, because both turn `Candor` from
metadata into a body rule. `internal/story/profile.go:24` documents `Candor`
as *"metadata surfaced to the LLM, not a body rule,"* and DEC-029 choice 2
makes profiles data rather than code.

1. **What a promotional profile does with its failures.** It can label them
   inline in their thread, move them to a separate block, or omit them — and
   if it omits them, DEC-050 requires a **visible note** in the output, of
   which a count is the obvious form. What that note says, where it sits, and
   whether a bundle that omits a failure can still be read as complete are
   this spec's to settle. Only the two invariants are locked: not rendered as
   a win, and not dropped in silence.
2. **Whether a failure still counts as an impact beat.** `IsImpactBeat` is
   computed inline at `internal/story/thread.go:135` as `e.Impact != ""`.
   It does not call `aggregate.WithImpact`, despite what DEC-029's text
   says. The answer moves `exec`'s `impact_threads_only` selection, the
   throughline's impact-beat counts, and the JSON `is_impact_beat` field.
   That last one is a count-bearing field, so DEC-048 applies.

Row 4 was narrowed by the maintainer on 2026-09-19, after this file was
written: omitting a failure from a promotional profile is allowed, and only
the silence is not. So *"it cannot be done without dropping them"* is no
longer the hand-back. DEC-050's T3 is now the harder question underneath it
— whether a visible note survives the model that consumes the bundle. If
framing finds it does not, that goes back to the maintainer with the
measurement that showed it.

## Reuse from SPEC-086

- `aggregate.FailureType`, `aggregate.IsFailure` and `aggregate.SplitFailures`
  exist. Use them, and do not restate the predicate.
- The heading literal is `## What didn't work`, and it renders only when
  non-empty (DEC-050 rules 4 and 5).
- The standing traps still apply. A NOT-contains assertion passes on a corpus
  with no failure in it, so pair it with a positive. A green suite is not
  evidence: framing's M-1 survived the whole suite.

## Out of scope

- `impact` and `wrapped`. SPEC-086 covers them.
- `--type` negation, which is a separate `bug` entry on STAGE-023.
- The impact-quality classifier (STAGE-024).
- `summary`: SPEC-095, split out at framing (below).

---

## Framing (2026-09-22, `main` at `c0b840e`)

Every number below was taken against a read-only copy of the live corpus
(`sqlite3 ~/.bragfile/db.sqlite ".backup '<tmp>/db.sqlite'"`), with a binary
built from `c0b840e` and run with `--db` pointed at the copy. Before any
equality was believed, each side was checked to be the artifact: the first
line is `# Bragfile Story` (or `# Bragfile Summary`), stderr is empty, and
the JSON parses. The live corpus was not written.

### Re-derived numbers

| What | 2026-09-18 | 2026-09-22 | Command |
|---|---:|---:|---|
| Entries | 606 | **621** | `select count(*) from entries` |
| `type: failed` | 4 (420, 433, 465, 473) | 4 (same ids) | `brag list --type failed --format tsv` |
| Near-miss spellings (`Failed`, `failure`, …) | — | 0 | `select type, count(*) … where lower(type) like '%fail%'` → only `failed\|4` |
| Failures with an impact | 4 of 4 | 4 of 4 | `… where type='failed' and impact=''` → 0 |
| `story --year`, failures rendered `- ★ <id>:` | 4 / 4 / 4 / 4 | 4 / 4 / 4 / 4 | `/usr/bin/grep -c -E '^- ★ (473\|465\|433\|420): '`, per profile: exec / skip / manager / me |
| `story --year` JSON, failed beats | `[420,433,465,473]` | same, **every** profile | `[.threads[].beats[] \| select(.type=="failed") \| .id] \| sort` |
| …of those, `is_impact_beat: true` | — | 4, every profile | `select(.type=="failed" and .is_impact_beat)` |
| `story --year` `Beats:` | — | exec 546/621, skip 619/621, manager 621/621, me 621/621 | header line |
| `summary --range month`, failures in `## Highlights` | 4 | 4 | `/usr/bin/grep -c -E '^- (473\|465\|433\|420): '` |

**Only the corpus size moved (606 → 621).** Nothing about the failures moved:
the same four ids, all with an impact, rendered as `★` on every profile.

### The split: the file's premise was false for `summary`

This file, like DEC-050's Context, says `summary` and `story` *"already carry
the data and drop it only in markdown."* That is true of `story`, whose JSON
beat is 7 keys and carries `"type": "failed"`. **It is false of `summary`.**
Its JSON highlight entry is `{"id","title"}`, so a failure is
indistinguishable from a win *per entry in both formats*. Only the aggregate
`counts_by_type` knows. That is the **unrepresentable** shape SPEC-086
measured on `impact` and `wrapped`, not the lossy-markdown shape.

The split follows that measurement, as SPEC-086's did:

- **SPEC-095 — `summary`.** Unrepresentable per entry. The fix is SPEC-086's
  exact partition (`aggregate.SplitFailures` into `## What didn't work` /
  `failures_by_project`). DEC-050 row 3 states the posture in full, so it
  decides nothing. **S**, GO, and with no dependency on this spec's open
  question.
- **SPEC-094 (this spec) — `story`.** Representable in JSON and dropped in
  markdown, **with a posture that varies by profile**. Its questions are
  open, and they need a decision record and a maintainer confirmation.

The two share no file (`internal/export/summary.go` vs `internal/story/`).
Keeping them together would tie a fully decided change to a question that
waits on the maintainer.

### Question 2 first — a failure stays an impact beat

It comes first because Question 1's mechanism depends on it.

**The discrepancy, confirmed on the tree.** `IsImpactBeat` is set at
`internal/story/thread.go:135` as `e.Impact != ""`, and the same predicate
is restated a second time at `thread.go:87` (`DropImpactlessBeats && e.Impact
== ""`). Neither calls `aggregate.WithImpact`, although DEC-029 choice 1 says
*"reusing `aggregate.WithImpact`"* and `thread.go:18`'s comment says it
*"mirrors aggregate.WithImpact's rule exactly."* **Today the two are
behaviourally identical**: `WithImpact` is also `e.Impact != ""`
(`aggregate.go:297`), and the corpus holds 0 whitespace-only impacts. It is a
drift exposure, not a live bug. The only test that pins it,
`thread_test.go:83`, restates the predicate too, so a change to `WithImpact`
would leave `story` behind with the suite green. Design fixes it by calling
the shared predicate rather than restating it (a per-entry form may need
extracting, since `WithImpact` filters a slice).

**The decision: a failure that carries an impact is still an impact beat.**
`is_impact_beat` keeps meaning *non-empty impact*. What changes is the
**marker**: `★` stops meaning *"is an impact beat"* and starts meaning
*"an impact beat that is not a failure."* This is DEC-050 rule 2's partition
applied to `story`, and it is the answer DEC-050 already gave when it
rejected narrowing `WithImpact` (framing's M-1): that change *"makes a true
headline false."*

**The DEC-048 consequence is none, by construction.** No count changes its
definition, so nothing is renamed: `is_impact_beat`, `impact_beat_count`,
and the throughline's `N with impact` all still count every beat with an
impact, failures included, exactly as `impact`'s `Entries: …with impact`
headline spans both of its sections. The opposite answer was measured and
rejected. It changes what three names count, so DEC-048 would require
renaming all three. It reorders `exec --month`, where `bragfile=9` and
`irradiance=9` tie today and `bragfile` would drop to 6. It also empties
`story --audience exec --type failed`, which is 2 threads and 4 beats today
and would be 0 and 0.

**What the answer does *not* move on this corpus:** `exec`'s
`impact_threads_only` selection. Over every profile and every window
(`--year`, `--quarter`, `--month`), **0 threads** have failures as their only
impact beats. Measured as `[.threads[] | select((impact beats) > 0 and
(non-failure impact beats) == 0)] | length` → 0 in all 12 cells.

### Question 1 — what a promotional profile does with its failures

**Measure first.** The profiles' directives are data. A note is only useful
if it reaches whoever reads the output, and a promotional bundle's reader
is a model (`docs/tutorial.md:651`:
`brag story --audience exec --quarter | claude "weave these threads into one
headline arc"`). So framing ran that pipe. The bundle was
`brag story --audience exec --quarter --project bragfile` (Q3, 42 beats, 3 of
them failures: 433, 465, 473), and the unmodified `exec.md` directive went
into each variant unless noted. Every variant was checked for shape before
any run: its failure-beat count, note line count and clause count. Each ran
5× on Haiku and 5× on Sonnet via `claude -p --tools ""
--setting-sources ""`, with the tutorial's prompt.

Two things were graded. The first was whether a failure appeared in the
prose, using keywords unique to each failure's own title and impact. These
were calibrated to hit v0's failure beats and nothing in v1, and they avoid
phrases the shipped siblings 464, 472 and 632 also use. The second was
whether an omission note appeared. That regex was calibrated to **0 hits on
the unmodified bundle**: an early version matched `13 failed`, `excluding
none` and entry 410's own *"work that didn't work"*. Every hit was then read
by hand.

| Variant | What the bundle carries | Failure in the prose | Omission stated in the prose |
|---|---|---:|---:|
| V0 | status quo: failures as `- ★` | **0 / 10** | — |
| V3 | failures labelled inline, `- ✗ <id> (failed):` | **0 / 10** | — |
| V4 | failures moved to their own `## What didn't work` block | **0 / 10** | — |
| V1 | failures omitted, plus a header line `Omitted: 3 failed entries — not listed for this audience (brag list --type failed)` | n/a (omitted) | **0 / 10** |
| V5 | V1, plus a body line: `Any narrative written from this bundle must state that 3 recorded failures were omitted.` | n/a (omitted) | **4 / 10** (Haiku 0/5, Sonnet 4/5) |
| V2 | V1, plus one clause **appended to the directive**: *"This bundle omits 3 recorded failures … End with one line that says so; do not drop it."* | n/a (omitted) | **10 / 10** |
| V2-full | V2 on the full `exec --quarter` bundle (all projects, 457 in window, 4 failures, 221 KB), Sonnet only | n/a (omitted) | **4 / 4** (5th run invalid: the model attempted a tool call and printed `{}`; excluded, not counted as a miss) |

What the table says:

- **Under a promotional directive, the failures never reach the prose, and
  placement does not change that** (V0, V3, V4: 0 of 30). Labelling them
  inline or moving them to their own block, both options this file named for
  a promotional profile, does not keep them visible downstream. The directive
  says *"No process, no messy middle"*, and the model obeys. So a
  promotional profile is **silently dropping failures today**: not in the
  bundle, where they are listed as wins, but in the prose made from it.
- **A visible note does not survive on its own** (V1: 0 of 10). DEC-050
  called a count of the omitted failures *"the obvious form"*. Measured, it
  is insufficient: it satisfies row 4's letter and not its point, which is
  exactly T3's trigger.
- **An instruction to carry the note does survive, if it sits in the
  directive** (V2: 10 of 10, V2-full: 4 of 4). In the bundle body (V5) it
  held on Sonnet and never on Haiku. Every surviving note was checked by hand
  to be the note, not a false positive.

**The decision: a promotional profile omits its failures, shows a visible
note, and bragfile appends a carry clause to the end of the rendered framing
directive.** A candid profile labels its failures inline. Of the three
options the file named, omit is the only one that removes the failures from
a prompt that tells the model to promote beats. The note is what makes the
omission visible in the bundle a human reads. The appended clause is what
keeps it visible in the prose a model writes. Label and block (V3, V4) would
each still need the same clause to be anything other than silent
downstream, so they cost more and buy nothing.

What design owes inside that decision:

- **The rule keys on exactly `candor: promotional`.** Every other value
  labels, including an empty or unknown one. `Candor` is a free string in a
  user-override profile (`profile.go` parses it unvalidated), and the safe
  default on a typo is the one that never drops anything.
- **Why the clause has to come from the binary, not the directive asset.**
  A user-override profile can point `directive:` at its own file, which will
  not carry a clause. So bragfile appends it to whatever directive text it
  resolved, and only when it omitted something.
- **The note's wording, and where it sits.** The measured form is a header
  line under `Beats:`. It must render in the empty state as well: under
  `--audience exec --type failed`, every beat is an omitted failure, so the
  threads are empty and the note is the only thing that says why. A JSON
  key named for what it counts (DEC-048) and always present (DEC-014
  part 4) goes with it.
- **Can a bundle that omits failures be read as complete? Yes.** `Beats:
  <shown>/<in-window>` already declares every bundle a subset (`exec --year`
  is 546/621), and the note names the failure part of that gap. Design must
  order the omission **before** the `impact_threads_only` fold, so that a
  thread whose only impact beats were failures folds and is still counted in
  the note.
- **The candid label.** `★` is not available (Question 2), and `·` means
  *no recorded impact*, which a failure with an impact is not. Whether a
  labelled failure survives the candid directives (`me.md` asks for *"the
  messy middle"*) was not measured. It is verify's to measure, with this
  section's method.

Limits of the measurement: one scope (one project, one quarter), two models,
5 runs a cell, and clause wording written at framing. It is enough to
reject the bare note and to show a working form. It is not enough to pin
wording, and design should re-run the V2 cell on whatever wording it locks.

### A decision record is needed

**Yes.** This spec turns `Candor` from metadata into a body rule. It
decouples `★` from `is_impact_beat`, a rendering DEC-029 choice 1 locked.
And it has the binary write text into the LLM-facing directive, which
DEC-029 choice 2 made data. Each of those is DEC-029-adjacent, and the third
is a new kind of output. Design writes the record and **takes the next free
number when it writes it**, derived from `ls decisions/`. That is DEC-054
on 2026-09-22, but **SPEC-092's front matter also names DEC-054**
conditionally, as the number its own design would take for an interpreter
dependency. Two specs naming one number in prose is the reservation this
stage keeps paying for, so neither owns it until a file does. Design also
decides whether DEC-029 gets an amendment note, as DEC-028 and DEC-030 did
at SPEC-086.

### Corrections found on the tree

- **`Candor` is not "surfaced to the LLM."** `profile.go:24` says it is.
  Measured, it appears in neither format: there is no `candor` line in any
  profile's markdown and no `candor` key in the JSON envelope (`generated_at,
  scope, audience, filters, threads, throughline, framing_directive`). Every
  `candid`/`promotional` hit in the markdown is either corpus text
  (*"candidates"*) or the directive's own prose. The field is parsed and
  dropped. The only thing telling the model the profile is promotional is the
  directive prose. The comment is wrong today, and this spec makes it doubly
  wrong.
- **`summary`'s defect shape** (above), which moved half this spec to
  SPEC-095.
- **DEC-054 is named in two specs' prose** (above).

### Open question for the maintainer (T3)

DEC-050 T3 reads: *"a promotional `story` bundle's visible note does not
survive the model that consumes it … Then the omission itself goes back to
the maintainer, with the measurement."*

**It fired for the form DEC-050 named**, a bare count (V1: 0 of 10). A form
that survives exists (V2: 14 of 14 valid), so this is not a request to
re-decide row 4. But the working form has bragfile author text that is not
in any profile, append it to the LLM prompt, and do so over a user's own
directive file. That is a new category of output. The maintainer is asked
one question:

> **May `brag story` append a fixed, binary-authored clause to the framing
> directive when a promotional profile omits failures?**
> - **Yes (recommended):** design proceeds as above.
> - **No:** then no measured form keeps a promotional profile's failures
>   visible past the model. Row 4's invariant holds in the bundle only, T3
>   stays open, and that should be written into DEC-050 rather than left
>   implied.

**ANSWERED — YES, as framed. Maintainer, 2026-09-22**, on V2's 14 of 14 and on
the measurement that today's promotional bundle already loses its failures
downstream (V0: 0 of 10). `brag story` may append the fixed clause. The
decision above is unblocked, and design locks the wording, then **re-runs the
V2 cell against the wording it locks** rather than inheriting framing's. The
decision record this spec needs is where the posture is written down; DEC-050's
T3 closes with it.

### What else was found and not routed

- **`brag review`** also lists all four failures as `- <id>: <title>`
  (`brag review --month`). It is not one of DEC-050's seven surfaces, since
  it has no `--type` flag. It lists them under a neutral `## Entries`
  heading followed by reflection questions, not as wins, so it does not
  break DEC-050's invariant as written. It is recorded here rather than
  given an id, because nothing in STAGE-023's criteria asks for it.

### Release impact (v0.7.0)

- **The split adds no time to the release path**, because SPEC-095 (S, no
  open question) can run its cycles alongside this one's. The cost is one
  extra spec's worth of sessions (design, build, verify, ship), and none of
  them wait on the maintainer.
- **The one thing that can delay the cut is the maintainer's answer to T3.**
  Design cannot lock the mechanism without it. It is a yes/no, and with a
  yes the scope is as sized.
- **Window cliffs for verify.** The four failures leave `story --quarter`
  when Q4 starts on **2026-10-01**, and SPEC-095's rolling
  `summary --range month` on about **2026-10-06**. After those dates a
  live-corpus NOT-contains check passes on a window with no failure in it.
  Use `--year`, `--previous`, or a seeded store, and pair every
  NOT-contains with a positive.

## Complexity

**M**, for `story` alone. The work is:

- the promotional omission ordered before the fold;
- the note and its JSON key, including the empty state;
- the appended clause;
- the candid label and the `★` decoupling;
- the shared-predicate fix;
- a decision record;
- goldens across four profiles;
- `docs/api-contract.md`'s `brag story` section (line 832);
- a re-run of the V2 cell on the locked wording.

Nothing touches storage. It stays M rather than L because the posture is
already decided (DEC-050 row 4) and the one open mechanism question is
measured rather than argued.

## GO / NO-GO

**GO**, at complexity **M**, scoped to `story`. **One spec for `story`, and
`summary` split to SPEC-095**, on defect shape (above).

**Conditional on one maintainer answer before design locks the mechanism**:
the T3 question above. Framing recommends *yes*, on V2's 14 of 14. Design can
begin while the answer is pending, because Question 2's answer, the candid
label, the note, the shared-predicate fix and the decision record do not
depend on it. It can only lock the appended clause once answered.

**It gates v0.7.0**, together with SPEC-095. DEC-050's *"until SPEC-094
ships"* consequence now reads *until SPEC-094 and SPEC-095 ship*.

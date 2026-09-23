---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-094
  type: story                      # epic | story | task | bug | chore
  cycle: ship                      # frame | design | build | verify | ship
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
  framed_at: 2026-09-22
  designed_at: 2026-09-22              # after framing merged (#226), main f2c7b9f

references:
  decisions:
    - DEC-050                      # rows 3 and 4 are this spec's posture; it
                                   # implements them, it does not re-decide them
    - DEC-029                      # profiles are data — why story's mechanics
                                   # are a decision, not a line of code
    - DEC-014                      # the envelope, incl. part 4's empty-state rule
    - DEC-048                      # a count names what it counted
    - DEC-054                      # EMITTED at this spec's design: story's
                                   # candor decides what a failure renders
                                   # as; closes DEC-050's T3
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-086                     # split this out; implements DEC-050 rows 1-2
    - SPEC-095                     # the `summary` half (DEC-050 row 3), split
                                   # out at this spec's framing
---

# SPEC-094: story stops listing a failure as a win

> **Cycle: design.** Designed 2026-09-22 against `main` at `f2c7b9f`, after
> framing merged (#226) and the maintainer answered T3 *yes*. **GO at M,
> held.** Framing's record is kept below as it was written. Wherever design
> re-measured a number or overturned a claim, the design sections say so and
> take precedence. **Start at *What design settled*.**

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

---

## What design settled (2026-09-22, `main` at `f2c7b9f`)

> Every number was measured against a **frozen file copy** of
> `~/.bragfile/db.sqlite`, taken with `sqlite3 .backup`, which takes a read
> lock only. The copy holds 621 entries, max id 643, and four `type: failed`
> rows. Its SHA-256 (`2657544f9a74…`) was the same before and after every run
> against it. The live corpus was never opened for writing. Two binaries ran
> against it: `brag-main`, built from `f2c7b9f`, and `brag-proto`, built from
> a prototype worktree carrying every literal in *Notes for the Implementer*.

- **Question 1, the mechanism.** A profile whose `candor` is **exactly**
  `promotional` omits recorded failures **before** threads are built. It
  says so three ways. In markdown, a line directly under `Beats:` reads
  `Omitted: <n> recorded failures, not listed for this audience (brag list --type failed)`,
  and renders only when `n > 0`. In JSON, `omitted_failure_count` is an
  integer between `filters` and `threads`, always present. For the model,
  the binary appends a fixed paragraph to the resolved framing directive:
  `This bundle omits <n> recorded failures for this audience. End with one line that says so; do not drop it.`
  Every other `candor` value labels failures inline, as
  `- ✗ <id> (failed): <title>`.
- **The re-run on that wording (T3).** In **25 of 25** runs the omission was
  stated, as the last line of the model's prose. The runs covered Haiku and
  Sonnet, `exec` and `skip`, and a scoped and a full bundle. Framing's cell
  was 14 of 14, on framing's wording and `exec` only. No omitted failure's
  content reached any prose. One `skip` run named the three failures
  *"production/operational escapes"*: the omission and the count were
  stated, and the noun was wrong. **The locked wording measures no worse
  than framing's.** It is locked.
- **Question 2.** A failure with an impact stays an impact beat. `★` now
  means "an impact beat that is not a failure". `internal/story` calls
  `aggregate.HasImpact`, a new per-entry form of `WithImpact`'s rule, at
  both sites that restated it. **No count changed what it counts, measured
  in 12 cells** (four profiles × `--year`/`--quarter`/`--month`). Candid
  JSON is byte-identical to `main` once the new key is removed.
- **The decision record is DEC-054**, claimed by its file in this commit.
  `next_id DEC` returned `DEC-054` when the file was written, and no branch
  on `origin` carries a `DEC-054` or `DEC-055` file. DEC-029 and DEC-050 each
  gain an `## Amendment (2026-09-22, SPEC-094 design)`, and their original
  text is left as written. DEC-050's T3 closes.
- **Window cliffs.** Every Go test runs on a seeded fixture, so none depends
  on the date. Every live-corpus criterion below uses
  `--since 2026-09-01`, an explicit anchor that holds all four failures for
  good. `--quarter` loses them on 2026-10-01, and `--year` on 2027-01-01.

---

## Re-measurement at design (2026-09-22, `main` at `f2c7b9f`)

Design does not inherit framing's numbers. Each row was measured with
`brag-main` on the frozen copy. Before any equality was believed, each side
was checked to be the artifact: its first line is `# Bragfile Story`, its
stderr is empty, and its JSON parses. One check caught a real defect in the
measurement itself: the first JSON sweep piped `echo "$j" | jq`, and zsh's
`echo` expands the `\n` escapes inside JSON strings, so `jq` failed to parse
all 12 cells. The sweep was re-run through files.

| Framing said (2026-09-22, `c0b840e`) | Measured at design (`f2c7b9f`) |
|---|---|
| 621 entries, 4 `type: failed` (420, 433, 465, 473) | **Same.** Max id 643. `lower(type) like '%fail%'` → only `failed\|4`. 0 failures without an impact. |
| `story --year` renders each failure `- ★ <id>:` on every profile | **Same**: 4 / 4 / 4 / 4 (`exec` / `skip` / `manager` / `me`). Also 4 on each profile under `--quarter`. |
| JSON failed beats `[420,433,465,473]`, all `is_impact_beat: true` | **Same**, in all 12 cells (4 profiles × year/quarter/month) |
| `Beats:` exec 546/621, skip 619/621, manager and me 621/621 | **Same** |
| 0 threads whose only impact beats are failures, in 12 cells | **Same**, 0 in 12 |
| `exec --month`: `bragfile=9` and `irradiance=9` tie | **Same** (`contextcore-pilot-harness` 164, `bragfile-site` 47, `bragfile` 9, `irradiance` 9) |
| JSON envelope keys | `generated_at, scope, audience, filters, threads, throughline, framing_directive` |

**Nothing moved since framing.** The corpus also still holds no failure
without an impact, so DEC-050's T5 has not fired.

**What the prototype does to the same copy** (`brag-proto`, every cell's
shape checked as above):

| Cell | `main` | prototype |
|---|---|---|
| `exec --year` | `Beats: 546/621`, 4 × `★` failure lines | `Beats: 542/621`, `Omitted: 4 recorded failures, …`, 0 failure lines of either kind |
| `skip --year` | `619/621`, 4 × `★` | `615/621`, `Omitted: 4 …`, 0 |
| `exec` / `skip`, `--quarter` | 430/457 and 457/457 | 426/457 and 453/457, `Omitted: 4 …` |
| `exec` / `skip`, `--month` | 230/231 and 231/231 | 226/231 and 227/231, `Omitted: 4 …` |
| `manager`, `me` (all three windows) | 4 × `- ★ <id>:` | 4 × `- ✗ <id> (failed):`, no `★` failure line, no `Omitted:` |
| `exec --year --type failed` | `Threads: 2`, `Beats: 4/4`, four `★` beats | `Threads: 0`, `Beats: 0/4`, `Omitted: 4 …`, then the directive and the clause. JSON: `omitted_failure_count: 4`, `threads: []` |

---

## Question 1 — settled: omit, say so three ways, and the clause measured

### The locked artifacts

| What | Exactly |
|---|---|
| The rule | `Profile.OmitsFailures()` is `p.Candor == CandorPromotional`, with `CandorPromotional = "promotional"`. It is case-sensitive and untrimmed (the parser already trims around the value). |
| When it applies | in `runStory`, **after** `s.List(filter)` and **before** `BuildThreads`: `shown, omitted := story.OmitFailures(entries, profile)`. `EntriesInWindow` stays `len(entries)`. |
| The markdown note | `Omitted: <n> recorded failures, not listed for this audience (brag list --type failed)`, directly under `Beats:`. `n == 1` reads `1 recorded failure`. **Absent when `n == 0`.** |
| The JSON key | `omitted_failure_count`, an integer, between `filters` and `threads`, **always present**, `0` when nothing was omitted (DEC-014 part 4) |
| The clause | `This bundle omits <n> recorded failures for this audience. End with one line that says so; do not drop it.` Appended to the resolved directive as **its own paragraph**: the directive is right-trimmed of `\n`, then `\n\n` + clause + `\n`. When the directive is empty, the clause plus `\n` is the whole directive. Absent when `n == 0`. It is the same in both formats, because both render `framingDirective(opts)`. |
| The candid label | `- ✗ <id> (failed): <title>`, then `  <impact>` when the beat carries one. `✗` is U+2717. |
| `--print-directive` | unchanged: the asset as authored, with no clause, because it reads no window |

**Why the note and the clause are driven by one field.** `StoryOptions`
gains `OmittedFailures int`, and the renderer derives the `Omitted:` line,
the JSON key and the clause from it. No caller can set one without the
other. The alternative, where the CLI appends the clause to `Directive`
itself, leaves a renderer that can print a note with no clause (M-9 below
is that shape, and six tests catch it).

**Why the wording differs from framing's, and what that cost.** Framing's
V2 clause was a bullet appended to `exec.md`'s list:
`- This bundle omits 3 recorded failures (see "Omitted:" above). End with one line that says so; do not drop it.`
Design changed three things:

1. **"(see "Omitted:" above)" is gone.** The same clause lands in the JSON
   `framing_directive`, where there is no `Omitted:` line to point at.
2. **"for this audience" was added.** It is the note's own phrase, and it
   tells the model the omission was deliberate rather than a gap.
3. **It is a paragraph, not a bullet.** A user's `directive:` file need not
   end in a list.

Framing's doctored bundle also carried a throughline that still read
`42 beats, 42 with impact` while listing 39. The prototype's reads
`39 beats, 39 with impact`, because the omission happens before the
skeleton is built. Any of those differences could have moved the result, so
the cell was re-run.

### The re-run (T3), on the locked wording

**Method.** It was framing's method, re-run with bundles the prototype
emitted rather than bundles edited by hand. Each run was one
`claude -p --model <m> --tools "" --no-session-persistence --setting-sources ""`
with the tutorial's prompt, *"weave these threads into one headline arc"*,
and the bundle on stdin. The run used `claude` 2.1.280, five runs a cell,
five in parallel. Every bundle was shape-checked before any run:

| Bundle | Command | Bytes | Failure beat lines | `Omitted:` lines | Clause lines | SHA-256 |
|---|---|---:|---:|---:|---:|---|
| `l2` | `story --audience exec --quarter --project bragfile` | 22 673 | 0 | 1 (`3 recorded failures`) | 1 | `8ff0ffe8eecd` |
| `l2full` | `story --audience exec --quarter` | 220 789 | 0 | 1 (`4 …`) | 1 | `fe57c08d4939` |
| `l2skip` | `story --audience skip --quarter --project bragfile` | 22 909 | 0 | 1 (`3 …`) | 1 | `c4ca4159667c` |
| `l0` (control) | the `l2` command on `brag-main` | 25 080 | 3 | 0 | 0 | — |

(The hashes cover a `Generated:` line, so they identify these files and no
re-run will reproduce them. Before 2026-10-01 the commands above reproduce
the bundles. From 2026-10-01 to 2026-12-31, add `--previous`.)

**Grader, and its calibration.** The note regex is framing's, unchanged:
`omit.{0,40}fail|fail.{0,40}omit|exclud.{0,40}fail|fail.{0,40}exclud|(^|[^0-9])(3|4|three|four) (recorded |failed |documented )*(fail|entr|misstep|setback|dead)|failed entries|recorded failures|type failed|not (listed|included|shown)`.
Run with `/usr/bin/grep -ciE`, it scores **0 hits** on the `l0` control
bundle and on all **30** of framing's outputs that had no omission to report
(V0, V3, V4, still on disk from framing's session). On framing's V5 outputs
it scores 6 of 10, where framing reported 4 of 10 after reading them by
hand. **The regex is an upper bound, so every hit was read.** A leak regex
over keywords unique to each failure's title and impact
(`known duplication|anonymous routing|next spec that opens|in prose|write the mechanism down|reserved an id|two of five|green suite|corrupted artifact|nobody read|worker pool|branch list|unmerged branch`)
checked that no omitted failure reached the prose.

**Result.**

| Cell | Model | Runs | Regex hit | Read by hand: the omission is stated | …as the last non-blank line | Leak |
|---|---|---:|---:|---:|---:|---:|
| `exec`, scoped (`l2`) | Haiku | 5 | 5 | 5 | 5 | 0 |
| `exec`, scoped (`l2`) | Sonnet | 5 | 5 | 5 | 5 | 0 |
| `exec`, full (`l2full`) | Sonnet | 5 | 5 | 5 | 5 | 0 |
| `skip`, scoped (`l2skip`) | Haiku | 5 | **4** | **5** | 5 | 0 |
| `skip`, scoped (`l2skip`) | Sonnet | 5 | 5 | 5 | 5 | 0 |
| **total** | | **25** | **24** | **25** | **25** | **0** |

**Every hit, as read.** Each line below is the model's own, quoted exactly.

- `exec`/Haiku: *"Note: 3 recorded failures omitted for this audience."* ·
  *"(This summary omits 3 recorded failures; see `brag list --type failed`.)"* ·
  *"This summary omits 3 recorded failures; see brag list --type failed for the full operational record."* ·
  *"Summary omits 3 recorded failures."* ·
  *"This summary includes 36 of 39 recorded achievements. 3 recorded failures omitted for this audience (see `brag list --type failed`)."*
  The fifth run's *"36 of 39"* is the model's own arithmetic, and it is
  wrong: the bundle shows 39 of 42.
- `exec`/Sonnet: *"This bundle omits 3 recorded failures for this audience."* ·
  *"3 recorded failures are omitted from this exec view (see `brag list --type failed`)."* ·
  *"3 recorded failures omitted from this exec view (see `brag list --type failed`)."* ·
  *"This summary omits 3 recorded failures not shown for this audience."* ·
  *"3 recorded failures from this quarter are omitted from this exec-audience view."*
- `exec` full/Sonnet: *"This bundle omits 4 recorded failures for this audience."* ·
  *"This summary omits 4 recorded failures not shown for this audience."* ·
  *"4 recorded failures this period are omitted from this summary for this audience."* ·
  *"4 recorded failures are omitted from this summary for this audience."* ·
  *"4 recorded failures were omitted from this summary for this audience."*
  (Framing's full cell lost one run of five to a tool-call attempt. None was
  lost here.)
- `skip`/Haiku: *"**Omitted:** 3 recorded failures for this audience—see `brag list --type failed` for details."* ·
  **the regex miss:** *"**Note:** This summary omits 3 recorded
  production/operational escapes from the same period that did not reach this
  audience."* The omission and the count are stated, and the noun is wrong:
  433, 465 and 473 are process failures, not production escapes. Counted as
  **stated**, and reported here because it is the one line a reader could
  misread. ·
  *"This update omits 3 recorded failures for this audience."* ·
  *"**Omitted:** 3 recorded failures not listed for this audience — see `brag list --type failed`."* ·
  *"3 recorded failures omitted for this audience."*
- `skip`/Sonnet: *"This bundle omits 3 recorded failures for this audience."* ·
  *"Three recorded failures are omitted from this account for this audience."* ·
  *"This bundle omits 3 recorded failures not surfaced for this audience."* ·
  *"This omits 3 recorded failures not shown for this audience."* ·
  *"This bundle omits 3 recorded failures for this audience."*

**Against framing.** Framing's V2 was 10 of 10 plus 4 of 4 valid, so 14 of
14, on `exec` only. Design's is **25 of 25** with the omission stated, or 24
of 25 if the line must also use the word *failure*. It adds a `skip` cell,
which framing did not measure, and it measured *where* the line lands: last,
in every run. **The locked wording measures no worse, so it ships.** The
limits are framing's, narrowed by one: two models, one quarter, one project
scope plus one full bundle, and now two promotional directives rather than
one. DEC-054's T1 is the trigger to re-measure.

**Not measured, and routed to verify:** whether the candid label
`✗ <id> (failed)` survives `me.md` and `manager.md` into the prose. Framing
flagged it. It does not gate the mechanism, because a candid profile omits
nothing, so no silence can arise there. DEC-054's T4 names it.

---

## Question 2 — settled: the predicate is called, and no count changes meaning

**The drift fix.** `aggregate.HasImpact(e storage.Entry) bool` is new. It
holds `WithImpact`'s rule for one entry, and `WithImpact` now calls it, so
there is one definition. `internal/story` calls it at both sites that
restated it:

- `thread.go:135`: `IsImpactBeat: e.Impact != ""` → `IsImpactBeat: aggregate.HasImpact(e)`
- `thread.go:87`: `opts.DropImpactlessBeats && e.Impact == ""` → `opts.DropImpactlessBeats && !aggregate.HasImpact(e)`

The failure flag is called the same way: `Beat.IsFailure` is
`aggregate.IsFailure(e)`, set in `entryToBeat`. It is a Go field and not a
JSON key; the beat projection stays 7 keys, and `type` already carries it.
**`TestStoryPackage_CallsTheSharedPredicates` pins all three.** It scans
the package's non-test sources for `Impact != ""`, `Impact == ""` and the
literal `"failed"`, and M-2, M-3 and M-15 each turn it red while every
behavioural test stays green.

**DEC-048, confirmed by running it and not by asserting it.** Both
binaries were run on the frozen copy in 12 cells, and the JSON compared:

```
for a in exec skip manager me; do for w in year quarter month; do
  brag-main  --db <copy> story --audience $a --$w --format json > m.json
  brag-proto --db <copy> story --audience $a --$w --format json > p.json
  # shape guard: both parse and both have threads; a failure prints SHAPE FAIL
  ...
done; done
```

| Check | Candid (`manager`, `me`) × 3 windows | Promotional (`exec`, `skip`) × 3 windows |
|---|---|---|
| `jq -S 'del(.omitted_failure_count)' p.json` vs `jq -S . m.json` | **identical, 6 of 6** | — (failures removed by design) |
| per-arc `{beat_count, impact_beat_count}` equals `main`'s, recomputed over `main`'s non-failure beats, with threads left with 0 impact beats dropped | — | **match, 6 of 6** |
| each arc's counts equal its own thread's beats (`beats\|length`, `is_impact_beat` count) | match, 6 of 6 | match, 6 of 6 |
| beats where `is_impact_beat != (impact != "")` | **0** | **0** |
| `omitted_failure_count` | 0 | 4 |

So on a candid profile every count is byte-identical. On a promotional
profile every count is the same function of the beats it shows, computed
over fewer beats, which is what "no count changed meaning" means.
`is_impact_beat` means *carries an impact* in all 12 cells.

**What does move, and is reported rather than hidden.** `exec --month`'s
impact-desc order changes. `bragfile` goes from 9 impact beats to 6, because
three of its nine were failures, and it drops below `irradiance` (9). The
omission causes that, not Question 2's answer: under the rejected answer
(failures are not impact beats), `bragfile` would also read 6, *and* the
three candid counts would change meaning. The markdown diff between the two
binaries on a candid profile is exactly **4 lines** (`me --year` and
`manager --year`, each 8 `diff` lines). Each `- ★ <id>:` line becomes
`- ✗ <id> (failed):`, and the impact line under it is unchanged.

---

## The decision record: DEC-054

`decisions/DEC-054-story-candor-decides-what-a-failure-renders-as.md`,
**new** in this commit. How the number was confirmed free, at the moment the
file was written:

```
$ bash -c '. scripts/_lib.sh; next_id DEC ./decisions'
DEC-054
$ git fetch -q origin && git log --all --oneline -- 'decisions/DEC-054*' 'decisions/DEC-055*'
(no output — no branch carries either)
```

The file was created **first**, with its front matter and title. Only then
was `DEC-054` written into any other file: the code comments in the
literals, DEC-029's and DEC-050's amendments, `api-contract.md`, the
CHANGELOG and STAGE-023. **SPEC-092 and SPEC-093 still say** *"the next free
number is DEC-054"* in their front matter (`:42` in each) and SPEC-093 `:77`
and `:139`. Each of those statements was true when it was written and is
false from this commit. **This spec does not edit them**, because they are
SPEC-092's and SPEC-093's and out of scope. They are recorded under
*Corrections*, where SPEC-093, whose subject is exactly this, can take them.

Its posture, one line per claim:

1. **`Candor` is a body rule on exactly `promotional`.** Every other value
   labels, so a typo never drops anything.
2. **An omission is stated three ways**: the `Omitted:` line when `n > 0`,
   `omitted_failure_count` always, and a fixed clause appended to the
   directive, even to a user's own file.
3. **`★` means "an impact beat that is not a failure"**, while
   `is_impact_beat` keeps meaning "carries an impact". It is computed by
   `aggregate.HasImpact`, and nothing is renamed.
4. **DEC-050's T3 is closed**: the omission was stated in 0 of 10 runs with
   the bare note, and in 25 of 25 with the clause.

**DEC-029 is amended rather than superseded**, as DEC-028 and DEC-030 were
at SPEC-086. The maintainer's R3 ruling there is the rule: amend when the
original text goes false. Choice 2's *"candor level"*, choice 5's key list,
and the *"directive appended verbatim"* of choices 5 and 7 each go false.
Choice 1's *"reusing `aggregate.WithImpact`"* becomes true for the first
time. **DEC-050 is amended too.** Its row 3 names SPEC-094 as implementer,
which went false at framing (it is SPEC-095 now). Its row 4 has a mechanism
to point at, and its T3 has fired and closed. A reader of DEC-050 alone
would otherwise see T3 open.

**The inventory moves two rows at design**, regenerated and not predicted:
`Decision records` 52 → **53**, and `…carrying an explicit ## Amendment
section` 3 → **5**. No harness literal pins either (§9(b) below).

---

## Window cliffs, and how this spec avoids them

The four failures are dated 2026-09-06 (420), 2026-09-06 (433), 2026-09-08
(465) and 2026-09-08 (473). A window loses them when it rolls past them:

| Window | Loses them on | Used here? |
|---|---|---|
| `story --quarter` (Q3) | **2026-10-01** | only in the T3 bundles, measured on 2026-09-22. To reproduce, add `--previous` from 2026-10-01 to 2026-12-31. |
| `story --month` (September) | 2026-10-01 | only in the 12-cell design measurement |
| `summary --range month` (rolling) | about 2026-10-06 | not at all (SPEC-095's) |
| `story --year` | 2027-01-01 | only in the design measurement |
| **`story --since 2026-09-01`** | **never** | **every live-corpus acceptance criterion** |

**Chosen: `--since 2026-09-01` for the live corpus, and seeded fixtures for
every Go test.** `--since` is an explicit anchor with an open end, so the
four failures stay in it permanently, and a new failure written later only
raises `n`. AC-2 states `n` as *the count `brag list --type failed --since
2026-09-01` returns*, not as `4`. The Go tests seed their own stores
(`failureStoryFixture`, and `brag learn` in the e2e test) with dates in 2026
and windows of `--year` over a fixed `storyFixedNow`, or `--since
2000-01-01`. None reads the clock or the live corpus. `--year` was
rejected for the live criteria because it has its own cliff on 2027-01-01,
and a seeded store alone would leave AC-1 to AC-4 unmeasured on real data.

---

## §12(b) design-time pre-flight: what the tools actually said

Every literal in *Notes for the Implementer* was run through its real tool.
The Go code went through `go build`, `gofmt -l`, `go vet`, `just lint`
(0 issues) and the full suite: 14 packages `ok`, 1113 passing tests. The help
sentence went through cobra's `--help` (line 9 of `brag story --help`,
verbatim). The markdown and JSON went through `brag-proto` on the frozen
copy. The doc hunks and Group `AE` went through `./scripts/test-docs.sh`
(ALL OK, 210 `OK:` lines). The inventory went through
`scripts/inventory.sh`. And the clause went through two models (above).
**The literals are `git diff` output from the prototype, base `f2c7b9f`.**
`git apply --check` accepts all 15 hunks-files against this design branch,
which changes none of those files.

**Finding 1: zsh's `echo` corrupts JSON on the way to `jq`.** The first
12-cell sweep ran `j=$(brag … --format json); echo "$j" | jq …`. In zsh,
`echo` expands the `\n` escapes inside JSON strings, so all 12 cells printed
`jq: parse error` and an empty result. It was caught because the sweep
printed the shape and not only the comparison. Re-run through files, all 12
parsed. It is the §12 *no difference* clause's family: an empty result on
both sides would have compared equal. **Anyone re-running this spec's jq
commands in zsh should write to a file, or use `print -r --`.**

**Finding 2: the clause measures no worse on the locked wording.** See
*The re-run (T3)* above: 25 of 25 stated, 0 leaked, and one noun wrong.

**Finding 3: two CLI probes failed to compile, and were discarded rather
than credited.** M-6 (`BuildThreads(entries, …)`) and M-7
(`OmittedFailures: 0`) each left a variable unused, and `cmd/brag` and
`internal/cli` failed to build. That is a red for the wrong reason, and in a
log it looks just like a real one (§12, the *no difference* clause and
SPEC-086's Finding 5). They were re-run as M-6′ and M-7′, which compile, and
each fires `TestLearnCmd_StoryLabelsOrOmitsWhatItWrote`.

**Finding 4: M-D1 survived, and the probe was wrong, not the guard.** The
first doc probe renamed one of the three `omitted_failure_count` mentions in
the contract's `story` section, and `AE1` stayed green. That is correct:
`AE1` asserts the section *names* the key, and two mentions still did. The
defect `AE1` exists to catch is a section that stops documenting the key.
M-D1′ renames all three mentions, and `AE1` fires. **AE1 does not check that
every mention is right**, and nothing in this spec claims it does.

**Finding 5: fail-first was run, in two shapes, not predicted.** The new
and modified test files were copied onto an unmodified `main` export.

```
raw (tests only):
internal/aggregate [build failed]  undefined: HasImpact
internal/story     [build failed]  undefined: OmitFailures / CandorPromotional; Beat has no field IsFailure
internal/cli       FAIL  2 tests:  TestLearnCmd_StoryLabelsOrOmitsWhatItWrote, TestStoryCmd_HelpStatesTheFailureRule
(every other package ok)

with symbol-only stubs (HasImpact correct; OmitFailures returns (entries, 0);
OmitsFailures false; Beat.IsFailure and StoryOptions.OmittedFailures unset):
internal/story     FAIL  12 tests: all 10 new, plus the 2 planned rewrites
                   (TestToStoryJSON_MeProfile_ShapeGolden, TestToStory_EmptyWindow)
internal/aggregate ok    (TestHasImpact passes against any correct HasImpact)
```

Every `story` and `cli` failure in the stubbed run is an assertion
failure, for the reason its test states. **`TestHasImpact_IsWithImpactsRuleForOneEntry`
does not fail first against a correct stub, and should not.** It is an
agreement guard between `HasImpact` and `WithImpact`, as SPEC-086's
`CountsByProjectSpansBothSections` was a guard. M-16 turns it red. Build
should expect exactly this picture.

**Finding 6: after the fix, a change to the impact rule carries `story`
with it.** M-16 changes `HasImpact` to trim whitespace. Only
`TestHasImpact` fires, and every story golden stays green, because story no
longer holds its own copy. Before this spec, the same change to
`WithImpact` would have left `story` behind with the suite green, which
framing called the drift exposure. M-2 and M-3 re-introduce the copy, and
only `TestStoryPackage_CallsTheSharedPredicates` catches them.

**Finding 7: the help sentence renders verbatim, and the NOT-contains
self-audit found nothing to fix.** Every NOT-contains this spec adds is
paired (LD16). Their needles are `- ★ <id>:`, `\nOmitted: `,
` <id>: `, ` <id> (failed)`, `This bundle omits` and `"failed"`. They were
grepped against the load-bearing prose that reaches the output: the four
directive assets and four profile assets (0 hits for each), and the new
`Long` (it says `Omitted:` mid-sentence, never at a line start, and no test
reads `Long` for a NOT-contains).

### Mutation matrix: 25 probes, each confirmed by content hash before its gates ran

The probe helper (`probe.py`, in the design session's scratchpad) **refused
to run the gates until the file's hash had moved** (§12(b) clause (1),
refined). It then ran `go test -count=1 ./...` and `./scripts/test-docs.sh`,
restored the file from a backup, and confirmed the hash had returned. The
*Diff* column is the replacement the helper applied, verbatim. Baselines,
measured on the prototype's final files: `thread.go` `b4b18d867c40`,
`profile.go` `2e4b6f59c520`, `bundle.go` `c0decde8de07`, `cli/story.go`
`90683ed9106a`, `aggregate.go` `2249dacd3b3c`, `docs/api-contract.md`
`1cd16e2a3d7a`, `docs/tutorial.md` `3a5ba9709b54`, `AGENTS.md`
`713a717a154f` and `CHANGELOG.md` `1410bc873849`. **Every target was back at
its baseline at the end.**

| # | File | Diff (text replaced → replacement) | Hash | Fired |
|---|---|---|---|---|
| **M-1** | `story/thread.go` | `IsImpactBeat: aggregate.HasImpact(e),` → `IsImpactBeat: aggregate.HasImpact(e) && !aggregate.IsFailure(e),` (Question 2's rejected answer) | `b4b18d867c40`→`8fb51098f865` | `…AFailureWithImpactIsStillAnImpactBeat`, `…RunsBeforeTheFold` (its control no longer keeps `delta`), `…CandidLabelsFailuresGolden` (the throughline counts), the `learn` e2e |
| **M-2** | `story/thread.go` | `IsImpactBeat: aggregate.HasImpact(e),` → `IsImpactBeat: e.Impact != "",` (the restatement) | →`ce6e92db4e64` | **only** `TestStoryPackage_CallsTheSharedPredicates` |
| **M-3** | `story/thread.go` | `if opts.DropImpactlessBeats && !aggregate.HasImpact(e) {` → `if opts.DropImpactlessBeats && e.Impact == "" {` | →`351b8670fae0` | **only** `…CallsTheSharedPredicates` |
| **M-15** | `story/thread.go` | `IsFailure:    aggregate.IsFailure(e),` → `IsFailure:    e.Type == "failed",` | →`ce0ee1372c6b` | **only** `…CallsTheSharedPredicates` |
| **M-4** | `story/profile.go` | `return p.Candor == CandorPromotional` → `return p.Candor != "candid"` | `2e4b6f59c520`→`f2e659c8a795` | `…OnlyExactPromotionalOmits`, `TestLoadProfile_OnlyExactPromotionalOmitsFailures` |
| **M-5** | `story/profile.go` | same line → `return strings.EqualFold(strings.TrimSpace(p.Candor), CandorPromotional)` | →`20af9be5ffc7` | the same two |
| ~~M-6~~ | `cli/story.go` | `BuildThreads(shown, …` → `BuildThreads(entries, …` | →`da6cd10fd74a` | **Discarded**: `shown` unused, a compile red (Finding 3) |
| **M-6′** | `cli/story.go` | `shown, omitted := story.OmitFailures(entries, profile)` + `BuildThreads(shown, ` → `_, omitted := story.OmitFailures(entries, profile)` + `BuildThreads(entries, ` (counts but does not omit) | `90683ed9106a`→`ca1ddc415298` | **only** `TestLearnCmd_StoryLabelsOrOmitsWhatItWrote` |
| ~~M-7~~ | `cli/story.go` | `OmittedFailures: omitted,` → `OmittedFailures: 0,` | →`c23fbf993fed` | **Discarded**: `omitted` unused, a compile red |
| **M-7′** | `cli/story.go` | `OmittedFailures: omitted,` → `OmittedFailures: 0 * omitted,` (omits silently) | →`e635b3be25fc` | **only** the `learn` e2e |
| **M-17** | `cli/story.go` | `…A user profile leaves failures out only when its candor is exactly promotional; any other value lists them.` → `…A user profile leaves failures out when its candor is promotional.` | →`d168c88b2468` | `TestStoryCmd_HelpStatesTheFailureRule` |
| **M-8** | `story/bundle.go` | `if opts.OmittedFailures > 0 {` → `if true {` (the note always renders) | `c0decde8de07`→`b495a25fef42` | 6, including the **existing** `me` and `exec` markdown goldens and `TestToStory_EmptyWindow` |
| **M-9** | `story/bundle.go` | `if opts.OmittedFailures == 0 {` → `if true {` (the clause never renders) | →`11ea798813b0` | 6: `…ClauseAloneWhenDirectiveEmpty`, the e2e, `TestLoadProfile_…` (the user's file), `…PromotionalCountsWhatItOmitted`, the promotional golden, `…EveryBeatOmittedStillSaysWhy` |
| **M-10** | `story/bundle.go` | `End with one line that says so; do not drop it.` → `End with one line that says so.` | →`ceb958cb1441` | the same 6 |
| **M-11** | `story/bundle.go` | `` `json:"omitted_failure_count"` `` → `` `json:"failures_omitted"` `` | →`731e442d6f1a` | 5, including the **existing** `TestToStoryJSON_MeProfile_ShapeGolden` and `TestToStory_EmptyWindow` |
| **M-12** | `story/bundle.go` | `case b.IsFailure:` → `case false:` (a failure renders as `★`) | →`c0ec94cec463` | `…CandidLabelsFailuresGolden`, the e2e |
| **M-13** | `story/bundle.go` | inside the failure case, `if b.IsImpactBeat {` → `if false {` | →`13df49770bfe` | the same two |
| **M-14** | `story/bundle.go` | `if d == "" {` / `return clause + "\n"` → `return ""` | →`b3f09ba2799b` | **only** `…ClauseAloneWhenDirectiveEmpty` |
| **M-16** | `aggregate/aggregate.go` | `return e.Impact != ""` (in `HasImpact`) → `return strings.TrimSpace(e.Impact) != ""` | `2249dacd3b3c`→`147d7cdbc170` | **only** `TestHasImpact_…`. Every story test stays green, which is the drift fix working (Finding 6). |
| M-D1 | `docs/api-contract.md` | **the second of three** (`:938` at `c3f707a`): `` extends DEC-014 with `audience`, `omitted_failure_count` (an integer, `` → `` extends DEC-014 with `audience`, `omitted_failures` (an integer, ``. *Pinned at verify (V-F1): as designed, the row read "one of three", which build could not reproduce.* | `1cd16e2a3d7a`→`471eb6137945` | **none**: the probe was too weak (Finding 4) |
| **M-D1′** | `docs/api-contract.md` | **all three** `` `omitted_failure_count` `` → `` `omitted_failures` `` (every one is in the `story` section) | →`f313c537a4cb` | `AE1` |
| **M-D2** | `docs/tutorial.md` | `…list it where it falls as` / `` `✗ <id> (failed)`, while `` → `…list it where it falls, while` | `3a5ba9709b54`→`927a9f5d3c3b` | `AE2` |
| **M-D3** | `AGENTS.md` | `` `brag story` labels one `` → `Story labels one` | `713a717a154f`→`bcab08c6ce19` | `AE3` |
| **M-D4** | `CHANGELOG.md` | `` `omitted_failure_count`, between `filters` `` → `` `omitted`, between `filters` `` | `1410bc873849`→`cf15beec071c` | `AE4` |
| **M-D5** | `docs/api-contract.md` | `` ### `brag story --audience <name> `` → `` ### `brag  story --audience <name> `` (two spaces) | `1cd16e2a3d7a`→`a1ffcbcc6883` | `AE1: … section not found`, the non-vacuity branch |

**Three probes matter most.** M-2, M-3 and M-15 are caught **only** by the
source scan, because a restated predicate that agrees today is behaviourally
invisible. That is exactly how `thread.go:135` sat for two projects. M-6′
and M-7′ are caught **only** by the end-to-end test. The renderer is correct
in both, and the defect is the CLI wiring, which the `story` package's tests
cannot see.

### Inventory: regenerated and diffed, not predicted

`scripts/inventory.sh` was run on three trees and diffed: a `git archive` of
`main` at `f2c7b9f`, this design branch, and the prototype carrying every
build literal.

| Row | `main` | this design commit | after build |
|---|---:|---:|---:|
| Decision records | 52 | **53** | 53 |
| …of those, carrying an explicit `## Amendment` section | 3 | **5** | 5 |
| Go test files | 80 | 80 | **81** |
| Go test functions | 843 | 843 | **856** |
| Documentation assertions (distinct ids) | 205 | 205 | **209** |

Every other row is unchanged. `Specs carried to ship and archived` stays at
86, because SPEC-094 is not archived. `Stages` stays at 23. No question was
filed, so `Y4`'s pins do not move. Passing tests including subtests, from
`go test -v ./... | grep -c -- '--- PASS'`, go from 1100 to **1113**.

### §9(b): the harness grepped by VALUE

For each value that moves, `scripts/test-docs.sh` was grepped for the
literal, with comment lines excluded:

```
$ for v in '| 52 |' '| 53 |' 'Amendment` section | 3' 'Amendment` section | 5' '!=52' '!=3' \
           '| 80 |' '| 81 |' '| 843 |' '| 856 |' '| 205 |' '| 209 |' '!=205' '!=843' '1100' '1113'; do
    /usr/bin/grep -n -F -- "$v" scripts/test-docs.sh | /usr/bin/grep -v '^[0-9]*:[[:space:]]*#'
  done
(no hits for any value)
$ /usr/bin/grep -rn --include='*_test.go' -E '\b(843|1100|205)\b' internal cmd
internal/storage/store_test.go:950/969/999: time.Sleep(1100 * time.Millisecond)   # a duration, not the count
```

**No literal pin exists for any of them.** `Y3` derives (SPEC-087), and `Z7`
reads what `inventory.sh` emits. Only `X3`'s page block moves, and it is
regenerated. DEC-054 costs zero `Y3` edits, like every record since SPEC-087.

---

## Locked design decisions

**LD1 — the rule keys on exactly `candor: promotional`.**
`story.CandorPromotional = "promotional"` and
`func (p Profile) OmitsFailures() bool { return p.Candor == CandorPromotional }`
are both in `profile.go`. The comparison is case-sensitive and untrimmed,
like DEC-050 rule 1's. Every other value labels inline: `candid`, empty,
unknown, `Promotional`, `promo`, and a value with trailing space set
programmatically. Today `exec` and `skip` omit, and `me` and `manager` label.

**LD2 — the omission runs before threading, in `runStory`, and nowhere
else.** `story.OmitFailures(entries, profile) (shown []storage.Entry,
omitted int)` returns the entries unchanged with `0` for a non-promotional
profile. Otherwise it returns `aggregate.SplitFailures`'s non-failures and
the failure count. `runStory` passes `shown` to `BuildThreads` and keeps
`EntriesInWindow: len(entries)`. So `impact_threads_only` folds a thread
whose only impact beats were failures, and the `--theme` cut never carries
an omitted failure either. A promotional profile omits **every** failure in
window, with or without an impact. An impact-less failure that `exec`'s
`drop_impactless_beats` would have dropped silently is now counted too.

**LD3 — the note.**
`Omitted: <n> recorded failures, not listed for this audience (brag list --type failed)`,
directly under `Beats:`. For one failure it reads `1 recorded failure`.
**It renders only when `n > 0`**, DEC-050 rule 4 applied to a line. It is
markdown only; the JSON carries the count (LD4).

**LD4 — the JSON key.** `omitted_failure_count`, an `int`, the fifth key:
`generated_at, scope, audience, filters, omitted_failure_count, threads,
throughline, framing_directive`. **Always present**, `0` when nothing was
omitted, including on every candid bundle and an empty window (DEC-014
part 4). The name follows the envelope's own `beat_count` and
`impact_beat_count` (DEC-048).

**LD5 — the clause.**
`This bundle omits <n> recorded failures for this audience. End with one line that says so; do not drop it.`
It uses the same `pluralFailures` form as the note. It is appended by
`framingDirective(opts)` as its own paragraph: the resolved directive,
right-trimmed of `\n`, then `\n\n`, the clause and `\n`. When the resolved
directive is empty, the clause plus `\n` becomes the directive, so the
`## Framing directive` section renders. It is appended **whatever the
directive resolved to**, a bundled asset or a user's file. It is absent
when `n == 0`, when the directive is byte-for-byte the asset. **Both formats
render `framingDirective(opts)`.** `--print-directive` is unchanged: it
reads no window, so it never carries the clause. **The wording is locked by
the T3 re-run above.** Changing it requires re-running that measurement
(DEC-054 part 2).

**LD6 — one field drives the note, the key and the clause.** `StoryOptions`
gains `OmittedFailures int`, set from `OmitFailures`' count. The CLI never
touches `Directive` for this. A caller that sets the count gets all three
outputs, and one that does not gets none.

**LD7 — the candid label.** A beat with `IsFailure` renders
`- ✗ <id> (failed): <title>` (`✗` is U+2717), then `  <impact>` when
`IsImpactBeat`. It is checked **before** `IsImpactBeat`, so `★` never
marks a failure. `markerFailure = "✗"` sits beside `markerImpact` and
`markerPlain`. The JSON beat is unchanged: still 7 keys, and `type` already
says `failed`.

**LD8 — Question 2: a failure with an impact is still an impact beat.**
`is_impact_beat`, `impact_beat_count`, the throughline's `<m> with impact`
and `Beats:` keep their definitions, and nothing is renamed (DEC-048). The
one thing that changes is the marker (LD7).

**LD9 — the predicates are called, never restated.**
`aggregate.HasImpact(e storage.Entry) bool` is new, and `WithImpact` calls
it. `thread.go` uses `aggregate.HasImpact` at both former restatement sites
and `aggregate.IsFailure` for `Beat.IsFailure`. No non-test file in
`internal/story` contains `Impact != ""`, `Impact == ""` or the literal
`"failed"`.

**LD10 — the help sentence is a literal**, a new paragraph in `story`'s
`Long`, after the audience table:
*Work recorded with brag learn is never marked as a win. me and manager list
it where it falls, as "✗ <id> (failed)". skip and exec leave it out, print an
Omitted: line counting it, and end the framing directive with a line telling
the LLM to say so. A user profile leaves failures out only when its candor is
exactly promotional; any other value lists them.*
`Short` and every flag's usage are unchanged.

**LD11 — the story fixtures type failures with the literal `"failed"`**,
never `aggregate.FailureType`, as SPEC-086 LD8 did. That way a change to the
persisted value fires the goldens rather than moving with them. The CLI
end-to-end test uses **neither**: it writes through `brag learn` and reads
through `brag story`.

**LD12 — the doc sweep is the premise audit's EDIT list, exactly.** Group
`AE` has four literal ids, placed directly above `# ===== finalise =====`.
It reuses Group `AD`'s `ad_section` and `assert_section_names`, so every id
is counted by `inventory.sh` (SPEC-086 Finding 1) and a missing section is
its own failure.

**LD13 — decision records, at design.** DEC-054 is new. DEC-029 and
DEC-050 each gain `## Amendment (2026-09-22, SPEC-094 design)`, appended
before `## References`, and nothing above those headings is edited
(SPEC-086 LD6).

**LD14 — scope.** Nothing under `internal/story/profiles/` or
`internal/story/directives/` changes. The assets are data, and the clause is
the binary's. `internal/export/` (`summary` is SPEC-095), `internal/mcpserver`
(no `story` tool) and storage are all untouched.

**LD15 — the inventory is regenerated at build**, by pasting
`./scripts/inventory.sh` between the markers. Three rows move. None is
hand-edited, and `X3` diffs the whole block.

**LD16 — every negative is paired with a positive** that fails if the
mechanism under test is absent:

| Negative | Its pair, in the same test |
|---|---|
| `exec` does not list the `brag learn` entry | it **does** list the `brag add` entry, **and** its `Omitted:` line counts 1 (`TestLearnCmd_StoryLabelsOrOmitsWhatItWrote`) |
| `me` neither stars nor omits the failure | it **labels** it `✗ … (failed)` with its impact (same test) |
| `exec` JSON carries no failure beat | `omitted_failure_count` is 3. The `me` half carries failure beats `[2 4 5]` (`TestToStoryJSON_PromotionalCountsWhatItOmitted`). |
| zero omitted renders no `Omitted:` line | two omitted render it, with the clause (`TestToStory_EveryBeatOmittedStillSaysWhy`) |
| seven non-promotional `candor` values omit nothing | `promotional` omits exactly 3 (`TestOmitFailures_OnlyExactPromotionalOmits`) |
| `exec` after omission has no `delta` thread | `exec` over the unfiltered fixture **has** it (`TestOmitFailures_RunsBeforeTheFold`) |
| the package source contains no restated predicate | at least four source files were scanned (`TestStoryPackage_CallsTheSharedPredicates`) |
| `--print-directive` carries no clause | it carries `exec`'s directive (`business impact`) |

### Rejected alternatives (build-time)

- **The CLI appends the clause to `StoryOptions.Directive`.** It works, but
  it puts the note and the clause in two places, where one can ship without
  the other. LD6 keeps them in one field.
- **An `OmitFailures` flag on `ThreadOptions`, applied inside
  `BuildThreads`.** The count would then have to come back out of
  `BuildThreads`, and the theme cut builds from the same entries, so the
  flag would have to be read twice. A separate function called before
  `BuildThreads` makes LD2's order visible at the call site.
- **Returning the omitted entries rather than their count.** Nothing reads
  them. `brag list --type failed` is the sanctioned way to get them, and the
  note says so.
- **An `is_failure` key on the JSON beat.** `type` already carries it, and
  DEC-029 choice 5's 7-key projection would grow for no information.
- **A `candor` key in the envelope.** DEC-054 leaves `candor` unrendered. Its
  effect is visible, and `omitted_failure_count` is the part a program
  needs.
- **Telling the model about the omission in `exec.md` and `skip.md`.** A
  user's own `directive:` file would not carry it (DEC-054).
- **`aggregate.FailureType` in the story fixtures.** Fail-first would be a
  compile error rather than assertion failures, and a changed constant would
  move the goldens with it (LD11).
- **Adding the new tests to `bundle_test.go` and `thread_test.go`.** They
  share one fixture, `failureStoryFixture`, which the existing six-entry
  `storyFixture` is deliberately not. A new file, `candor_test.go`, keeps
  every existing golden's fixture untouched.
- **Appending Group `AE` at the end of the file, or giving it its own
  helper.** An appended group never runs (SPEC-085 LD8), and a helper not
  named `assert_*` is never counted (SPEC-086 Finding 1).

---

## Outputs

### New files (1, at build)

| Path | Lines | What |
|---|---:|---|
| `internal/story/candor_test.go` | 529 | 10 tests over `failureStoryFixture` (LD11). Embedded verbatim in §4. |

### Modified files, at build (14)

| Path | Change | numstat on the prototype |
|---|---|---:|
| `internal/aggregate/aggregate.go` | `HasImpact`; `WithImpact` calls it | +10 / −1 |
| `internal/aggregate/aggregate_test.go` | 1 new test | +25 |
| `internal/story/profile.go` | `CandorPromotional`, `OmitsFailures`, and the struct comment (the *"surfaced to the LLM"* correction) | +17 / −1 |
| `internal/story/thread.go` | `Beat.IsFailure`, `OmitFailures`, both predicate sites, the `Beat` comment | +22 / −5 |
| `internal/story/bundle.go` | `markerFailure`, `OmittedFailures`, `pluralFailures`, `omissionClause`, `framingDirective`, the note, the label, the key | +82 / −26 |
| `internal/story/bundle_test.go` | **planned rewrites:** `TestToStoryJSON_MeProfile_ShapeGolden` +1 `want` line; `TestToStory_EmptyWindow` +1 assertion | +6 |
| `internal/cli/story.go` | `OmitFailures` before `BuildThreads`, `OmittedFailures`, and the LD10 `Long` paragraph | +5 / −1 |
| `internal/cli/story_help_test.go` | 1 new test | +11 |
| `internal/cli/learn_test.go` | `runDigestCorpus` registers `NewStoryCmd()` (**planned modification** of a shared helper), plus 1 new e2e test | +60 / −4 |
| `docs/api-contract.md` | the `brag story` section (LD12) | +33 / −8 |
| `docs/tutorial.md` | the `brag story` paragraph | +5 / −1 |
| `AGENTS.md` | §11, the `learn` glossary entry | ±1 |
| `CHANGELOG.md` | `[Unreleased]` → `### Changed`, two entries | +15 |
| `scripts/test-docs.sh` | Group `AE`, 4 ids, above `finalise` | +26 |
| `docs/engineering-practices.md` | the regenerated inventory block **only** (3 rows) | ±3 |

**Build total, measured on the prototype:** 1 new file (529 lines) and 14
modified. `git diff --stat` over the 15 literal files reads
`15 files changed, 847 insertions(+), 48 deletions(-)`, which counts the new
file but not the 3 inventory rows.

### Modified or created at design (this commit)

| Path | Change |
|---|---|
| `decisions/DEC-054-story-candor-decides-what-a-failure-renders-as.md` | **new**: the mechanism, the measurement, T3 closed |
| `decisions/DEC-029-story-audience-shaping-profiles-and-thread-definition.md` | `## Amendment (2026-09-22, SPEC-094 design)`, +38 lines, appended before `## References`. **Nothing above it is edited.** |
| `decisions/DEC-050-a-failure-is-never-rendered-as-a-win.md` | `## Amendment (2026-09-22, SPEC-094 design)`, +27 lines, the same way |
| `projects/PROJ-008-…/stages/STAGE-023-…md` | the SPEC-094 entry reads `(design)` and gains its design summary, and the count is re-derived: `1 in design / 4 framed` |
| `docs/engineering-practices.md` | the regenerated inventory block: 2 rows (52 → 53, 3 → 5) |
| this file | `cycle: design` (the recipe's stripped comment restored), `designed_at`, DEC-054 in `references`, and the design sections |

### The CHANGELOG entries

Under `## [Unreleased]` → `### Changed`, **before** the existing
`brag memory` entry, and after SPEC-086's two. Embedded verbatim in §7.

### Premise audit (§9), run at design against the repo

**Inversion: tests that assume every impact beat is `★`, that `Candor` is
metadata, or that the story envelope has seven keys.**

```
$ /usr/bin/grep -rln --include='*_test.go' -E '★|markerImpact|IsImpactBeat|is_impact_beat|framing_directive|ToStory(Markdown|JSON)|BuildThreads|runStoryCmd|Candor|NewStoryCmd' internal cmd
internal/cli/story_test.go        internal/cli/story_help_test.go
internal/story/bundle_empty_directive_test.go   internal/story/thread_test.go
internal/story/manager_skip_test.go             internal/story/bundle_test.go
$ /usr/bin/grep -n '"failed"\|FailureType' <those six files>
(no hits: no existing story test has a failure in its fixture)
```

**Found by execution, not by reading:** with the prototype in place, exactly
**one** existing test fails, `TestToStoryJSON_MeProfile_ShapeGolden`, which
is byte-exact over every key. **No test is deleted.**

| Test | Why | Plan |
|---|---|---|
| `TestToStoryJSON_MeProfile_ShapeGolden` | byte-exact over the envelope | **rewrite**: +1 `want` line, `"omitted_failure_count": 0,` after `"filters": {},` |
| `TestToStory_EmptyWindow` | still passes, but does not pin the new key's empty form | **additive**: +1 assertion |
| `runDigestCorpus` (`learn_test.go`) | a shared helper; the new e2e test needs `story` on its root | **modify**: register `NewStoryCmd()`, and update its comment. The two SPEC-086 e2e tests that use it stay green (measured). |
| the `me`/`exec` markdown goldens, `TestToStory_EmptyDirectiveOmitsSection`, `thread_test.go`, `manager_skip_test.go`, `story_test.go` | no failure in any fixture, and `OmittedFailures` is 0 | **none**. M-8 proves the two markdown goldens are live on the new line. |
| `TestWithImpact_*` (2) | `WithImpact`'s behaviour is unchanged | none |

**Addition: the tracked collections this spec adds to.** One DEC and two
Amendments, at design. One test file, 13 test functions and 4 doc ids, at
build. All five were regenerated rather than predicted (see *Inventory*),
and none is hand-pinned (§9(b)).

**Status change: every description of `story`'s output shape.** The
greps, over tracked files only
(`git ls-files -z | xargs -0 /usr/bin/grep -n -F`), excluding
`specs/done/`, test files and this spec:

```
needles: '★'  'is_impact_beat'  'impact beat'  'Beats: '  'framing_directive'  '7-key'
         'candor'  'Candor'  'promotional'  'surfaced to the LLM'  'IsImpactBeat'
plus:    brag story --help, read in full
```

| Hit | Verdict |
|---|---|
| `internal/story/profile.go:24` (*"Candor is metadata surfaced to the LLM, not a body rule."*) | **EDIT**: false today (framing), and false twice after this spec. In the §2 literal. |
| `internal/story/thread.go:18-20` (*"IsImpactBeat mirrors aggregate.WithImpact's rule exactly"*) | **EDIT**: it now *calls* it (§2) |
| `internal/story/bundle.go:20-26` (marker comment), `:125-128` (envelope key order), `:175-177` (`ToStoryJSON` doc) | **EDIT**, in the §3 literal |
| `internal/cli/story.go` `Long` | **EDIT**: LD10 |
| `docs/api-contract.md:861` (provenance), `:863-866` (markers), the directive bullet, the audience list, `:919-921` (JSON keys), `--print-directive`, `:927-931` (empty window) | **EDIT**: §5 |
| `docs/tutorial.md:617-618` (*"impact beats marked `★`"*), `:648` | **EDIT**: one sentence after the audience paragraph (§6). `:617`'s ★ sentence stays true: ★ still marks impact beats, and a failure now carries its own marker. |
| `AGENTS.md:294` (`learn` glossary: *"`brag impact` and `brag wrapped` list a failure…"*) | **EDIT**: it names every reader of the value, and `story` is the third (§6) |
| `CHANGELOG.md` `[Unreleased]` | **EDIT**: two entries (§7) |
| `decisions/DEC-029-…:55-57, :84, :114-124, :315` | **Amendment at design**, original text untouched (LD13) |
| `decisions/DEC-050-…:83` (row 4), `:257` (T3), row 3's *Implemented by* | **Amendment at design** (LD13) |
| `internal/story/directives/exec.md:6` (*"Every beat below carries a ★ impact statement"*) | **NO CHANGE**: still true, because `exec` omits failures and `drop_impactless_beats` removes the rest |
| `internal/story/directives/{me,manager,skip}.md` (`·` beats) | **NO CHANGE**: `·` still means no recorded impact. Whether `me` and `manager` should name `✗` is DEC-054 T4's to measure, not this spec's to guess (LD14). |
| `internal/story/profiles/*.yaml` `candor:` lines | **NO CHANGE**: the data is already right (LD14) |
| `docs/api-contract.md:1499` (the DEC-029 index line: *"impact beats marked via `WithImpact`"*) | **NO CHANGE**: now true through `HasImpact`, which is `WithImpact`'s rule |
| `guidance/questions.yaml:373, :442` | **NO CHANGE**: resolved questions, dated |
| `projects/PROJ-004-…`, `projects/PROJ-003-…/brief.md:57` | **NO CHANGE**: dated project records |
| `projects/PROJ-008-…/stages/STAGE-023-…:295, :599, :703` | **NO CHANGE**: dated measurements. The backlog entry is edited at design (above). |
| `NEXT-SESSION-PROMPT.md:97-98` | **NO CHANGE**: a session hand-off, overwritten per session |
| `README.md:229`, `BRAG.md:442` | **NO CHANGE**: an example with no shape claim |
| `docs/framework-feedback/…`, `docs/reports/…` (*"candor"*) | **NO CHANGE**: unrelated sense of the word |
| `internal/export/markdown.go:55`, `internal/memory/memory.go:240` (`e.Impact != ""`) | **NO CHANGE, recorded.** Both restate the impact rule to decide whether to print an impact line. They are outside `story`, and nothing in DEC-050 or this spec reaches them. See *Corrections*. |

---

## Acceptance Criteria

Numbers to diff against. Each was observed on the prototype or the frozen
copy before it was written here. **Every live-corpus criterion uses
`--since 2026-09-01`**, which holds all four failures (dated 2026-09-06 to
2026-09-08) permanently. See *Window cliffs*.

1. **Candid output changes only on its failure lines.** On one fresh frozen
   copy, run `main`'s binary and this branch's with
   `story --audience me --since 2026-09-01`, then again with `manager`.
   Each `diff` is exactly **8 lines** (4 `<` and 4 `>`): each
   `- ★ <id>: <title>` becomes `- ✗ <id> (failed): <title>`, one per id in
   `brag list --type failed --since 2026-09-01`. With `--format json`, the
   outputs are identical after `jq -S 'del(.omitted_failure_count)'` on this
   branch's side. At design: 473, 465, 433, 420.
2. **Promotional output omits, and says so.** On that copy,
   `story --audience exec --since 2026-09-01` and the same for `skip` list
   no failure id on any beat line. Each has
   `Omitted: 4 recorded failures, not listed for this audience (brag list --type failed)`
   directly under `Beats:`. `<shown>` is `main`'s minus 4, and `<in-window>`
   is unchanged (at design: `exec` 230/231 → 226/231, `skip` 231/231 →
   227/231). The last line of each document is
   `This bundle omits 4 recorded failures for this audience. End with one line that says so; do not drop it.`
   Build's copy may hold more entries. The criterion is that `4` equals the
   number of ids `brag list --type failed --since 2026-09-01` returns.
3. **The empty state explains itself.**
   `story --audience exec --since 2026-09-01 --type failed` renders
   `Threads: 0`, `Beats: 0/4`, the `Omitted:` line, and the directive
   ending in the clause. `--format json` gives `omitted_failure_count: 4`
   and `threads: []`.
4. **No count changes meaning** (DEC-048). The 12-cell check in
   *Question 2 — settled* holds on build's copy: candid JSON is identical
   after deleting the new key, and promotional arc counts equal `main`'s
   recomputed over non-failure beats. The check has to print a shape line
   per cell, so an empty or unparseable side fails loudly (Finding 1).
5. **The JSON key** is the fifth key, always present: `0` on every `me` and
   `manager` bundle and `4` on `exec` and `skip` over the AC-2 window
   (`jq -c keys_unsorted`).
6. **The goldens:** exactly one existing golden changes,
   `TestToStoryJSON_MeProfile_ShapeGolden`, by +1 `want` line. One existing
   test gains one assertion, `TestToStory_EmptyWindow`. **No existing
   markdown golden changes.** Two new full-document markdown goldens exist,
   at 45 and 36 source lines, from `want :=` through the directive splice.
7. **The fixtures hold failures.** `failureStoryFixture` has 3 `failed`
   rows (2 with an impact, 1 without) and 1 near-miss (`Failed`). The e2e
   test writes 1 through `brag learn`. **No assertion that proves new
   behaviour runs on a zero-failure fixture.** The zero-omission half of
   `TestToStory_EveryBeatOmittedStillSaysWhy` is a paired control.
8. **The predicates are called.**
   `/usr/bin/grep -n 'Impact != ""\|Impact == ""\|"failed"' internal/story/*.go | /usr/bin/grep -v _test`
   prints nothing, and `TestStoryPackage_CallsTheSharedPredicates` passes.
9. **Help:** the LD10 paragraph appears verbatim in `brag story --help`.
10. **Test counts:** 843 → **856** top-level test functions, 1100 →
    **1113** passing tests including subtests, and 80 → **81** test files.
11. **`just test-docs`:** ALL OK at **210 `OK:` lines / 209 distinct ids**
    (from 206 / 205). `grep -n '===== finalise =====\|===== Group AE'`
    puts `AE` on the lower line number, and
    `./scripts/test-docs.sh | grep -c 'OK:   AE'` is **4**.
12. **The inventory block is regenerated**, and exactly three rows move at
    build: 80 → 81, 843 → 856 and 205 → 209.
13. **The mutation matrix reproduces.** Each of the 23 non-discarded
    probes reproduces its stated hash from its stated diff **before** its
    gates run, and does what the table says: 22 fire, and M-D1 survives
    (Finding 4). The two discarded probes are not re-run.
14. **All five gates are green,** with no test excluded and no lint
    suppression: `go test ./...` (14 packages `ok`), `gofmt -l .` empty,
    `go vet ./...` clean, `just lint` **0 issues**, and `just test-docs`
    ALL OK.

---

## Failing Tests

Write them first and watch them fail for the stated reason (Finding 5 is the
expected picture), then make them pass. Every literal is in *Notes for the
Implementer*.

### New: `internal/story/candor_test.go` (10)

| Test | Asserts | Fails before because |
|---|---|---|
| `TestOmitFailures_OnlyExactPromotionalOmits` | `promotional` omits 3 (ids 2, 4, 5) and keeps `[1 3 6 7]`. `candid`, `""`, `Promotional`, `PROMOTIONAL`, `"promotional "`, `promo` and `promotion` omit 0 and keep all 7. | `OmitFailures` is undefined; stubbed, it omits 0 |
| `TestOmitFailures_RunsBeforeTheFold` | control: `exec` over the unfiltered fixture keeps `delta`. After `OmitFailures`, `delta` is gone and `omitted == 3`. | the same |
| `TestBuildThreads_AFailureWithImpactIsStillAnImpactBeat` | `{IsImpactBeat, IsFailure}` per beat (2 and 5 are `{true, true}`, 4 is `{false, true}`, 7 is `{true, false}`), and `delta`'s `ImpactBeatCount` is 1 | `Beat.IsFailure` does not exist; stubbed, it is never set |
| `TestStoryPackage_CallsTheSharedPredicates` | no non-test `.go` file in `internal/story` contains `Impact != ""`, `Impact == ""` or `"failed"`, and ≥ 4 files were scanned | `thread.go:87` and `:135` restate the rule |
| `TestToStoryMarkdown_CandidLabelsFailuresGolden` | the full `me` document, 45 lines plus `me.md`: `✗ 2 (failed)` with its impact, `✗ 4 (failed)` without one, `✗ 5 (failed)` beside `· 6`, and `★ 7` for the near-miss | each failure renders `★` or `·` |
| `TestToStoryMarkdown_PromotionalOmitsWithNoteAndClauseGolden` | the full `exec` document, 36 lines plus `exec.md` plus the clause: `Beats: 3/7`, the `Omitted: 3 …` line, `delta` folded, `★ 7` kept | nothing is omitted and there is no note or clause |
| `TestToStoryJSON_PromotionalCountsWhatItOmitted` | `exec`: `omitted_failure_count` is 3, no `failed` beat, the directive ends in `\n\n` + clause, and the key sits between `filters` and `threads`. `me`: the key is present and 0, beats `[2 4 5]` are failures, and the directive is the asset. | the key does not exist |
| `TestToStory_EveryBeatOmittedStillSaysWhy` | two failures only, on `exec`: the header through `## Framing directive` is exact, the document ends with the clause, and the JSON has `"omitted_failure_count": 2,` and `"threads": [],`. With the count at 0: no `Omitted:` line, and the directive is the asset. | the same |
| `TestFramingDirective_ClauseAloneWhenDirectiveEmpty` | an empty directive and 1 omitted: the exact document, singular `1 recorded failure` in both the note and the clause, and `framing_directive` is the clause plus `\n` | the same |
| `TestLoadProfile_OnlyExactPromotionalOmitsFailures` | bundled: `exec` and `skip` omit, `me` and `manager` do not. Overrides: `promotional` omits; `Promotional`, `promo` and no `candor` do not. A promotional override's **own directive file** gets the clause. | `OmitsFailures` is undefined |

### New: `internal/aggregate/aggregate_test.go` (1)

| Test | Asserts | Fails before because |
|---|---|---|
| `TestHasImpact_IsWithImpactsRuleForOneEntry` | `"cut p95 40%"`, `" "` and `"\n"` are true, `""` is false, and `WithImpact` keeps each entry iff `HasImpact` | `HasImpact` is undefined. **Against a correct stub it passes, by design** (Finding 5): it is the M-16 guard. |

### New: `internal/cli` (2)

| Test | File | Asserts | Fails before because |
|---|---|---|---|
| `TestLearnCmd_StoryLabelsOrOmitsWhatItWrote` | `learn_test.go` | `brag add` + `brag learn` on one store, then `story --since 2000-01-01`. `me` shows `★` for the add entry and `✗ <id> (failed)` with its impact for the learn entry, with no `★` on it and no `Omitted:`. `exec` shows the add entry and not the learn entry, `Beats: 1/2`, `Omitted: 1 recorded failure …` and the clause as the last paragraph. `--print-directive` has no clause. The JSON `omitted_failure_count` is 1. | the learn entry renders `- ★ <id>:` on both |
| `TestStoryCmd_HelpStatesTheFailureRule` | `story_help_test.go` | the LD10 paragraph, verbatim | it is absent |

### Changed, as planned rewrites (2 tests, 1 helper)

`TestToStoryJSON_MeProfile_ShapeGolden` gains one `want` line.
`TestToStory_EmptyWindow` gains one assertion. Both are red before build
(Finding 5, stubbed). `runDigestCorpus` registers `NewStoryCmd()`, and its
two SPEC-086 callers stay green.

### New: `scripts/test-docs.sh`, Group `AE` (4 ids)

| Id | Asserts, scoped to a section | Fails before because |
|---|---|---|
| `AE1` | `docs/api-contract.md`'s `brag story` section names `` `- ✗ <id> (failed): <title>` ``, `` `Omitted: <n> recorded failures ``, `` `omitted_failure_count` `` and `DEC-054` | none is present |
| `AE2` | `docs/tutorial.md`'s *Tell your story* section names `` `✗ <id> (failed)` `` and `` `Omitted:` `` | absent |
| `AE3` | the `AGENTS.md` `learn` glossary line names `` `brag story` labels one `` and `DEC-054` | absent |
| `AE4` | `CHANGELOG.md`'s `[Unreleased]` section names `` `omitted_failure_count` `` and `DEC-054` | absent |

A section that is not found is its own failure (M-D5), not a vacuous pass.

### Mutation checks (build re-runs the 23 non-discarded probes, recorded in Build Completion)

The probes, their stated diffs and their design-time hashes are in the
**Mutation matrix** above. Build reproduces each hash from its stated diff
**before** running the gates. If a stated diff does not reproduce its hash,
say so, and do not search for one that does. **M-2, M-3, M-15, M-6′ and
M-7′ carry the most weight**, because each is caught by one test only.

### Decision-to-test mapping (§9)

| Decision | Test(s) that fail without it |
|---|---|
| LD1 exact `promotional` | `…OnlyExactPromotionalOmits`, `TestLoadProfile_OnlyExactPromotionalOmitsFailures`; M-4, M-5 |
| LD2 omit before threading, in `runStory` | `…RunsBeforeTheFold`, the `learn` e2e; M-6′ |
| LD3 the note, its wording, only when `n > 0` | the promotional golden, `…EveryBeatOmittedStillSaysWhy` (both halves), `…ClauseAloneWhenDirectiveEmpty` (singular), the existing `me`/`exec` goldens; M-8 |
| LD4 the JSON key, position, always present | `…MeProfile_ShapeGolden`, `…EmptyWindow`, `…PromotionalCountsWhatItOmitted`, the e2e; M-11 |
| LD5 the clause, placement, user file, empty directive, `--print-directive` | the promotional golden, `…PromotionalCountsWhatItOmitted`, `…EveryBeatOmittedStillSaysWhy`, `…ClauseAloneWhenDirectiveEmpty`, `TestLoadProfile_…` (the user's file), the e2e; M-9, M-10, M-14 |
| LD6 one field drives all three | M-7′ (the count zeroed at the call site silences all three, and the e2e fires) |
| LD7 the candid label, and ★ never on a failure | `…CandidLabelsFailuresGolden`, the e2e; M-12, M-13 |
| LD8 a failure is still an impact beat, and no count changes meaning | `…AFailureWithImpactIsStillAnImpactBeat`, `…CandidLabelsFailuresGolden` (throughline), `…RunsBeforeTheFold` (its control); M-1; AC-4 |
| LD9 predicates called, not restated | `…CallsTheSharedPredicates`, `TestHasImpact_…`; M-2, M-3, M-15, M-16 |
| LD10 help literal | `TestStoryCmd_HelpStatesTheFailureRule`; M-17 |
| LD11 literal fixtures; e2e with no literal | `…CallsTheSharedPredicates` rejects `"failed"` in production code; the e2e writes through `brag learn` |
| LD12 the doc sweep, `AE` | `AE1`–`AE4`; M-D1′, M-D2, M-D3, M-D4, M-D5 |
| LD13 records at design | `X3` (the regenerated rows); the amendments' headings are counted by `inventory.sh` |
| LD14 scope | `git diff --stat main...HEAD` touches nothing under `internal/story/{profiles,directives}/`, `internal/export/` or `internal/mcpserver/` |
| LD15 regenerated block | `X3` |
| LD16 negatives paired | M-7′, M-8 and M-12 show the negatives are live, and the goldens show the positives are |

---

## Implementation Context

*Read this section, and the files it points to, before starting the build
cycle.*

### Decisions that apply

- **DEC-054**, written at this spec's design. **Read it first.** Its four
  parts are the contract, and part 2's wording is locked by measurement.
- **DEC-050**: the posture, and rule 1's predicate. Read its new
  `## Amendment` for how row 4 and T3 now stand.
- **DEC-029**: story profiles are data. Read its new `## Amendment` against
  choices 1, 2, 5 and 7.
- **DEC-014** part 4: the key is always present.
- **DEC-048**: a count names what it counted. It is why LD8 renames
  nothing.
- **DEC-049**: the reserved `failed` value and `brag learn`.

### Constraints that apply

- `one-spec-per-pr`.
- `no-sql-in-cli-layer`: nothing here touches SQL. `OmitFailures` filters
  in Go, over what `Store.List` returned.
- `stdout-is-for-data-stderr-is-for-humans`: the note and the clause are
  bundle content on stdout. Nothing goes to stderr.
- `test-before-implementation`: *Order of work*, step 2.

### Prior related work

- **SPEC-086** (shipped, `c0b840e`): the sibling, and the shape of this
  spec. It built `aggregate.IsFailure`, `SplitFailures` and `FailureType`,
  which this spec reuses unchanged, as well as `runDigestCorpus` and
  `markdownSection` in `learn_test.go`, and Group `AD`'s helpers.
- **SPEC-049 / SPEC-050**: `brag story` and its four profiles.
- **SPEC-085**: `brag learn`.

### Out of scope (for this spec specifically)

- **SPEC-095** (`summary`), `impact` and `wrapped`.
- **`brag review`**, which also lists failures unlabelled. It is recorded at
  framing, has no id, and is not one of DEC-050's seven surfaces.
- The restatements of the impact rule at `export/markdown.go:55` and
  `memory/memory.go:240` (see *Corrections*).
- **Measuring the candid label against a model** (DEC-054 T4). That is
  verify's, with this spec's method.
- `--type` negation, the impact classifier, SPEC-090 to SPEC-093,
  STAGE-027, `brag lint`, STAGE-020, `test-docs` in CI, and
  `guidance/questions.yaml`.
- **Brags.** Capture happens at ship.

---

## Corrections and open questions (design)

- **SPEC-092 and SPEC-093 now name a taken number.** `SPEC-092:42`,
  `SPEC-093:42`, `:77` and `:139` say *"the next free number is DEC-054"*.
  This commit's DEC-054 makes each of those statements false. They are
  statements about the past, which is exactly what SPEC-093 exists to stop
  prose from being. **Not edited here** (out of scope). SPEC-093's design
  should take them as a live instance.
- **The impact rule is still restated twice outside `story`.**
  `internal/export/markdown.go:55` (`brag export`'s per-entry impact line)
  and `internal/memory/memory.go:240` (the memory line) each test
  `e.Impact != ""`. Both agree with `HasImpact` today. Neither decides an
  impact *beat* or a count, only whether to print a line, so the drift
  would be cosmetic. **Recorded, not routed.** The next spec to open either
  file can switch it to `aggregate.HasImpact` in one line.
- **`profile.go:24`'s comment was false before this spec** (framing's
  correction). It is corrected in the §2 literal.
- **Framing's doctored V2 bundle had an inconsistent throughline**
  (`42 beats, 42 with impact` over 39 listed). It did not change framing's
  verdict, and design's re-run on generated bundles agrees with it.
- **Open for verify:** DEC-054 T4, whether `✗ <id> (failed)` survives
  `me.md` and `manager.md` into the prose. Use this spec's method, with
  `me --since 2026-09-01` and `manager --since 2026-09-01`.
- **No open question for the maintainer.** T3 was the only one, and it is
  answered.

---

## Notes for the Implementer

### Order of work

1. **Re-derive before trusting a number.** Take a fresh file copy of
   `~/.bragfile/db.sqlite` with
   `sqlite3 ~/.bragfile/db.sqlite ".backup '<tmp>/db.sqlite'"`, which takes
   a read lock only. Build `main`'s binary for AC-1 through AC-4's *before*
   side. **Never write to the live corpus.** Use `--since 2026-09-01` for
   every live check (*Window cliffs*).
2. **Write the tests first**: §4's new file and its three modified test
   files, plus §1's aggregate test. Run `go test ./...` and compare with
   Finding 5's *raw* picture: `aggregate` and `story` fail to build on the
   new symbols, and `cli` fails exactly 2 tests. To see the assertion-level
   picture before writing production code, add the symbols as stubs:
   `story` then fails exactly 12 tests.
3. **Production code**: §1 (`aggregate.go`), §2 (`profile.go`,
   `thread.go`) and §3 (`bundle.go`, `cli/story.go`). Run `go test ./...`
   until it is green.
4. **Docs**: §5 (`api-contract.md`), §6 (`tutorial.md`, `AGENTS.md`) and
   §7 (`CHANGELOG.md`).
5. **Group `AE`**: §8, inserted **directly above**
   `# ===== finalise =====`, which puts it after Group `AD`. Confirm with
   `grep -n '===== finalise =====\|===== Group AE' scripts/test-docs.sh`.
6. **Inventory, last**: `./scripts/inventory.sh`, pasted between the
   markers (§9). Exactly three rows move.
7. **Mutation matrix**: re-run the 23 non-discarded probes. Reproduce each
   stated hash from its stated diff **before** running the gates. A hash
   that does not move is a no-op, and it produces no evidence.
8. **The live-corpus criteria**, AC-1 to AC-5, on the fresh copy from step 1.
9. **Gates**: `just test`, `just test-docs`, `just lint`, `gofmt -l .` and
   `go vet ./...`.

### Traps

- **`AC2` is live.** A closing tool-call tag alone on a line fails
  `just test-docs` in any file you wrote, staged or not. Run it before every
  commit.
- **Use `/usr/bin/grep` or `git grep` for counts.** In this zsh, bare
  `grep` is a `ugrep` wrapper that respects `.gitignore`.
- **zsh:** never name a variable `path`. An unquoted `$var` does not
  word-split. `${B}:file` needs the braces. **And `echo "$json" | jq` corrupts
  JSON** (Finding 1): write to a file.
- **The flag is `--audience`, not `--profile`.** A comparison run with the
  wrong flag compares two identical usage errors, and they compare equal
  (§12, the *no difference* clause).
- **A NOT-contains check on the live corpus passes vacuously once the
  failures leave the window.** Use `--since 2026-09-01`, and pair every
  negative with the positive in the same AC.
- **`just advance-cycle` strips the inline comment on the `cycle:` line.**
  Restore it by hand.
- **`✗` is U+2717** (`\xe2\x9c\x97`), not U+2718 `✘` and not U+00D7 `×`.
  The goldens are byte-exact.
- **The literals below are `git diff` output from the prototype, base
  `f2c7b9f`.** The design commit touches none of these 15 files, so
  `git apply` accepts them on this branch as-is (checked with
  `git apply --check`). After any rebase that moves one of them,
  transcribe by hand rather than forcing the apply.
- **SPEC-091 (in build) edits `internal/cli/*.go` constructors**, adding
  `GroupID`. If it lands first, `cli/story.go`'s hunk may need a rebase.
  The overlap is textual, not semantic.

### §1. `internal/aggregate/aggregate.go` and its test

`HasImpact` is inserted directly after `WithImpact`, and `WithImpact` calls it. The new test is inserted directly above `TestIsFailure_ExactMatchOnTheReservedValue`.

````diff
diff --git a/internal/aggregate/aggregate.go b/internal/aggregate/aggregate.go
index dbcbd3d..9e2a6b7 100644
--- a/internal/aggregate/aggregate.go
+++ b/internal/aggregate/aggregate.go
@@ -294,13 +294,22 @@ func Span(entries []storage.Entry) CorpusSpan {
 func WithImpact(entries []storage.Entry) []storage.Entry {
 	out := make([]storage.Entry, 0, len(entries))
 	for _, e := range entries {
-		if e.Impact != "" {
+		if HasImpact(e) {
 			out = append(out, e)
 		}
 	}
 	return out
 }
 
+// HasImpact reports whether e carries a recorded impact: its Impact field is
+// non-empty. It is WithImpact's rule for one entry, split out so a caller
+// that projects entries one at a time (brag story's impact beats) calls the
+// rule instead of restating it — a restated copy is one a change to this
+// function leaves behind with every test green (SPEC-094).
+func HasImpact(e storage.Entry) bool {
+	return e.Impact != ""
+}
+
 // FailureType is the reserved entries.type value marking work that did not
 // work (DEC-049) — the ONE type value bragfile pins; `brag add --type` stays
 // free-form. It lives here rather than in internal/cli because it has a
````

````diff
diff --git a/internal/aggregate/aggregate_test.go b/internal/aggregate/aggregate_test.go
index 88ec09d..ff2b6a1 100644
--- a/internal/aggregate/aggregate_test.go
+++ b/internal/aggregate/aggregate_test.go
@@ -649,6 +649,31 @@ func TestWithImpact_EmptyInputAndAllEmptyImpact(t *testing.T) {
 	}
 }
 
+// TestHasImpact_IsWithImpactsRuleForOneEntry pins SPEC-094 LD6: HasImpact is
+// the per-entry form of WithImpact, and the two agree on every seed —
+// including a whitespace-only impact, which both count as an impact today.
+// brag story calls HasImpact rather than restating the rule.
+func TestHasImpact_IsWithImpactsRuleForOneEntry(t *testing.T) {
+	cases := []struct {
+		impact string
+		want   bool
+	}{
+		{"cut p95 40%", true},
+		{" ", true},
+		{"\n", true},
+		{"", false},
+	}
+	for _, c := range cases {
+		e := storage.Entry{ID: 1, Impact: c.impact}
+		if got := HasImpact(e); got != c.want {
+			t.Errorf("HasImpact(Impact=%q) = %v, want %v", c.impact, got, c.want)
+		}
+		if got := len(WithImpact([]storage.Entry{e})) == 1; got != HasImpact(e) {
+			t.Errorf("Impact=%q: WithImpact keeps it = %v, HasImpact = %v; the two must agree", c.impact, got, HasImpact(e))
+		}
+	}
+}
+
 // TestIsFailure_ExactMatchOnTheReservedValue pins SPEC-086 LD1: a failure is
 // exactly the reserved value DEC-049 persists — the literal "failed", matched
 // case-sensitively and untrimmed, the comparison storage's --type filter makes.
````

### §2. `internal/story/profile.go` and `thread.go`

`CandorPromotional` and `OmitsFailures` go directly after the `Profile` struct. The struct comment's *"surfaced to the LLM"* line was false before this spec, and it is corrected here. `OmitFailures` goes directly above `BuildThreads`, and both predicate sites call `aggregate.HasImpact` (LD9).

````diff
diff --git a/internal/story/profile.go b/internal/story/profile.go
index 51ed251..209fdf4 100644
--- a/internal/story/profile.go
+++ b/internal/story/profile.go
@@ -21,7 +21,8 @@ var ErrProfileNotFound = errors.New("story profile not found")
 // struct is DATA, not a Go enum (DEC-029 choice 2). Fields:
 //   - Selection: ImpactThreadsOnly, DropImpactlessBeats
 //   - Threading/altitude: FoldSmallThreads, ThreadOrder
-//   - Candor is metadata surfaced to the LLM, not a body rule.
+//   - Candor is a body rule on exactly one value, CandorPromotional
+//     (DEC-054); it is not rendered in either format.
 //   - Directive points at the framing-directive asset basename.
 type Profile struct {
 	Name                string
@@ -34,6 +35,21 @@ type Profile struct {
 	Directive           string // asset basename: "me.md" | "exec.md" | <user path>
 }
 
+// CandorPromotional is the one Candor value that changes the body (DEC-054):
+// a promotional profile omits recorded failures from the bundle, counts them
+// in an Omitted: line, and appends a clause to the framing directive telling
+// the consuming model to say so. The match is exact. Every other value —
+// candid, empty, unknown, or a misspelling — labels failures inline instead,
+// because Candor is an unvalidated string in a user profile and the default
+// on a typo must be the one that never drops anything.
+const CandorPromotional = "promotional"
+
+// OmitsFailures reports whether p omits recorded failures rather than
+// labelling them (DEC-054).
+func (p Profile) OmitsFailures() bool {
+	return p.Candor == CandorPromotional
+}
+
 // defaultOverrideDir resolves the user override directory
 // (~/.bragfile/story-profiles). A missing HOME degrades to an empty
 // path, which simply means "no override dir" (bundled defaults still
````

````diff
diff --git a/internal/story/thread.go b/internal/story/thread.go
index 91babcf..48587fb 100644
--- a/internal/story/thread.go
+++ b/internal/story/thread.go
@@ -15,9 +15,10 @@ const (
 	KindTheme      = "theme"
 )
 
-// Beat is one entry projected into a thread. IsImpactBeat mirrors
-// aggregate.WithImpact's rule exactly: an entry is an impact beat iff its
-// Impact field is non-empty.
+// Beat is one entry projected into a thread. IsImpactBeat is
+// aggregate.HasImpact, WithImpact's rule, called rather than restated; a
+// failure that carries an impact is still an impact beat (DEC-054).
+// IsFailure is aggregate.IsFailure, and decides only how the beat renders.
 type Beat struct {
 	ID           int64
 	Title        string
@@ -25,6 +26,7 @@ type Beat struct {
 	Type         string
 	Impact       string
 	IsImpactBeat bool
+	IsFailure    bool
 	CreatedAt    time.Time
 }
 
@@ -68,6 +70,20 @@ func impactBeatCount(t Thread) int {
 	return n
 }
 
+// OmitFailures applies a profile's candor to the in-window entries, before
+// they are threaded (DEC-054). A promotional profile loses its recorded
+// failures (aggregate.IsFailure), and omitted counts them; any other profile
+// gets entries back unchanged with omitted == 0. It runs before BuildThreads
+// on purpose: a thread whose only impact beats were failures then folds
+// under impact_threads_only, and its failures are still counted here.
+func OmitFailures(entries []storage.Entry, p Profile) (shown []storage.Entry, omitted int) {
+	if !p.OmitsFailures() {
+		return entries, 0
+	}
+	shown, failures := aggregate.SplitFailures(entries)
+	return shown, len(failures)
+}
+
 // BuildThreads coalesces the already-in-window entries into deterministic
 // threads per the profile's policy (DEC-029 choice 1/3):
 //  1. initiative threads via aggregate.GroupEntriesByProject (alpha-ASC,
@@ -84,7 +100,7 @@ func BuildThreads(entries []storage.Entry, opts ThreadOptions) []Thread {
 	for _, g := range groups {
 		beats := make([]Beat, 0, len(g.Entries))
 		for _, e := range g.Entries {
-			if opts.DropImpactlessBeats && e.Impact == "" {
+			if opts.DropImpactlessBeats && !aggregate.HasImpact(e) {
 				continue
 			}
 			beats = append(beats, entryToBeat(e, g.Project))
@@ -132,7 +148,8 @@ func entryToBeat(e storage.Entry, project string) Beat {
 		Project:      project,
 		Type:         e.Type,
 		Impact:       e.Impact,
-		IsImpactBeat: e.Impact != "",
+		IsImpactBeat: aggregate.HasImpact(e),
+		IsFailure:    aggregate.IsFailure(e),
 		CreatedAt:    e.CreatedAt,
 	}
 }
````

### §3. The renderer and its wiring: `internal/story/bundle.go`, `internal/cli/story.go`

`framingDirective` is the single place the clause is appended, and both renderers call it (LD5, LD6). The failure case is checked before `IsImpactBeat` (LD7). In `runStory`, `OmitFailures` sits between `s.List` and `BuildThreads` (LD2), and the `Long` gains the LD10 paragraph.

````diff
diff --git a/internal/story/bundle.go b/internal/story/bundle.go
index d25820d..4115997 100644
--- a/internal/story/bundle.go
+++ b/internal/story/bundle.go
@@ -15,34 +15,76 @@ import (
 	"bytes"
 	"encoding/json"
 	"fmt"
+	"strings"
 	"time"
 )
 
 // Markers for the standalone-readable markdown body (locked in the
-// goldens): an impact beat leads with ★ (U+2605), a plain beat with ·
-// (U+00B7) — a visible "so what" signal.
+// goldens): an impact beat that is not a failure leads with ★ (U+2605), a
+// plain beat with · (U+00B7) — a visible "so what" signal. A recorded
+// failure leads with ✗ (U+2717) and "(failed)", whether or not it carries an
+// impact, so ★ never marks one (DEC-050, DEC-054).
 const (
-	markerImpact = "★"
-	markerPlain  = "·"
+	markerImpact  = "★"
+	markerPlain   = "·"
+	markerFailure = "✗"
 )
 
 // StoryOptions is the pure renderer's input. The CLI does the windowing,
 // threading, throughline, and directive resolution and passes them in
 // (mirrors export.ImpactOptions). Scope echoes the resolved window token;
 // EntriesInWindow is the raw in-window count for the <shown>/<in-window>
-// beat tally; Now is injected for deterministic goldens.
+// beat tally; Now is injected for deterministic goldens. OmittedFailures
+// is OmitFailures' count, and drives both the Omitted: line and the clause
+// appended to the directive, so a caller cannot set one without the other.
 type StoryOptions struct {
 	Audience        string
 	Scope           string
 	Filters         string            // pre-formatted markdown line ("(none)" or echoed flags)
 	FiltersJSON     map[string]string // JSON filters object (nil → {})
 	EntriesInWindow int
+	OmittedFailures int
 	Now             time.Time
 	Threads         []Thread
 	Throughline     Throughline
 	Directive       string // resolved framing-directive text ("" → section omitted)
 }
 
+// pluralFailures renders a failure count for the Omitted: line and the
+// clause, which must agree on it.
+func pluralFailures(n int) string {
+	if n == 1 {
+		return "1 recorded failure"
+	}
+	return fmt.Sprintf("%d recorded failures", n)
+}
+
+// omissionClause is the fixed, binary-authored text appended to the framing
+// directive when a promotional profile omitted failures (DEC-054). It comes
+// from the binary, not the directive asset, because a user profile can point
+// directive: at its own file, which would not carry it. Its wording was
+// measured against a consuming model at SPEC-094 design; change it only with
+// a re-run of that measurement.
+func omissionClause(n int) string {
+	return fmt.Sprintf("This bundle omits %s for this audience. End with one line that says so; do not drop it.", pluralFailures(n))
+}
+
+// framingDirective is the directive both formats render: the resolved text,
+// with omissionClause appended as its own paragraph when failures were
+// omitted. An empty directive with an omission is the clause alone, so the
+// instruction survives a profile that carries no directive.
+func framingDirective(opts StoryOptions) string {
+	if opts.OmittedFailures == 0 {
+		return opts.Directive
+	}
+	clause := omissionClause(opts.OmittedFailures)
+	d := strings.TrimRight(opts.Directive, "\n")
+	if d == "" {
+		return clause + "\n"
+	}
+	return d + "\n\n" + clause + "\n"
+}
+
 // shownBeats counts the beats surfaced across all threads (the numerator
 // of the Beats: <shown>/<in-window> tally).
 func shownBeats(threads []Thread) int {
@@ -69,6 +111,9 @@ func ToStoryMarkdown(opts StoryOptions) ([]byte, error) {
 	fmt.Fprintf(&buf, "Filters: %s\n", opts.Filters)
 	fmt.Fprintf(&buf, "Threads: %d\n", len(opts.Threads))
 	fmt.Fprintf(&buf, "Beats: %d/%d\n", shownBeats(opts.Threads), opts.EntriesInWindow)
+	if opts.OmittedFailures > 0 {
+		fmt.Fprintf(&buf, "Omitted: %s, not listed for this audience (brag list --type failed)\n", pluralFailures(opts.OmittedFailures))
+	}
 
 	if len(opts.Threads) > 0 {
 		fmt.Fprintln(&buf)
@@ -78,10 +123,16 @@ func ToStoryMarkdown(opts StoryOptions) ([]byte, error) {
 			fmt.Fprintf(&buf, "### %s\n", t.Thread)
 			fmt.Fprintln(&buf)
 			for _, b := range t.Beats {
-				if b.IsImpactBeat {
+				switch {
+				case b.IsFailure:
+					fmt.Fprintf(&buf, "- %s %d (failed): %s\n", markerFailure, b.ID, b.Title)
+					if b.IsImpactBeat {
+						fmt.Fprintf(&buf, "  %s\n", b.Impact)
+					}
+				case b.IsImpactBeat:
 					fmt.Fprintf(&buf, "- %s %d: %s\n", markerImpact, b.ID, b.Title)
 					fmt.Fprintf(&buf, "  %s\n", b.Impact)
-				} else {
+				default:
 					fmt.Fprintf(&buf, "- %s %d: %s\n", markerPlain, b.ID, b.Title)
 				}
 			}
@@ -98,11 +149,11 @@ func ToStoryMarkdown(opts StoryOptions) ([]byte, error) {
 		}
 	}
 
-	if opts.Directive != "" {
+	if directive := framingDirective(opts); directive != "" {
 		fmt.Fprintln(&buf)
 		fmt.Fprintln(&buf, "## Framing directive")
 		fmt.Fprintln(&buf)
-		buf.WriteString(opts.Directive)
+		buf.WriteString(directive)
 	}
 
 	return trimTrailingNewline(buf.Bytes()), nil
@@ -123,16 +174,19 @@ func trimTrailingNewline(b []byte) []byte {
 
 // storyEnvelope is the on-the-wire JSON shape (DEC-029 choice 5). Field
 // order = key order (encoding/json preserves struct-tag declaration
-// order): generated_at, scope, audience, filters, threads, throughline,
-// framing_directive. Extends DEC-014's envelope with the arc-aware body.
+// order): generated_at, scope, audience, filters, omitted_failure_count,
+// threads, throughline, framing_directive. Extends DEC-014's envelope with
+// the arc-aware body. omitted_failure_count is always present, 0 when
+// nothing was omitted (DEC-014 part 4, DEC-054).
 type storyEnvelope struct {
-	GeneratedAt      string            `json:"generated_at"`
-	Scope            string            `json:"scope"`
-	Audience         string            `json:"audience"`
-	Filters          map[string]string `json:"filters"`
-	Threads          []threadJSON      `json:"threads"`
-	Throughline      throughlineJSON   `json:"throughline"`
-	FramingDirective string            `json:"framing_directive"`
+	GeneratedAt         string            `json:"generated_at"`
+	Scope               string            `json:"scope"`
+	Audience            string            `json:"audience"`
+	Filters             map[string]string `json:"filters"`
+	OmittedFailureCount int               `json:"omitted_failure_count"`
+	Threads             []threadJSON      `json:"threads"`
+	Throughline         throughlineJSON   `json:"throughline"`
+	FramingDirective    string            `json:"framing_directive"`
 }
 
 type threadJSON struct {
@@ -174,16 +228,18 @@ type spanJSON struct {
 // ToStoryJSON renders the arc-aware DEC-014-extending envelope with
 // 2-space indent (AC-6/AC-10). Threads/Arcs init to non-nil empty
 // slices; Filters nil → {}; framing_directive is always the resolved
-// directive string (renders even on an empty corpus).
+// directive string, with the omission clause when one applies (renders
+// even on an empty corpus).
 func ToStoryJSON(opts StoryOptions) ([]byte, error) {
 	env := storyEnvelope{
-		GeneratedAt:      opts.Now.UTC().Format(time.RFC3339),
-		Scope:            opts.Scope,
-		Audience:         opts.Audience,
-		Filters:          opts.FiltersJSON,
-		Threads:          make([]threadJSON, 0, len(opts.Threads)),
-		Throughline:      throughlineJSON{Arcs: make([]arcJSON, 0, len(opts.Throughline.Arcs))},
-		FramingDirective: opts.Directive,
+		GeneratedAt:         opts.Now.UTC().Format(time.RFC3339),
+		Scope:               opts.Scope,
+		Audience:            opts.Audience,
+		Filters:             opts.FiltersJSON,
+		OmittedFailureCount: opts.OmittedFailures,
+		Threads:             make([]threadJSON, 0, len(opts.Threads)),
+		Throughline:         throughlineJSON{Arcs: make([]arcJSON, 0, len(opts.Throughline.Arcs))},
+		FramingDirective:    framingDirective(opts),
 	}
 	if env.Filters == nil {
 		env.Filters = map[string]string{}
````

````diff
diff --git a/internal/cli/story.go b/internal/cli/story.go
index 0da5093..75bfbdd 100644
--- a/internal/cli/story.go
+++ b/internal/cli/story.go
@@ -36,6 +36,8 @@ func NewStoryCmd() *cobra.Command {
   skip     skip-level / director — outcomes grouped by initiative, less detail, more "so what"; quarterly
   exec     impact-forward promotion — impact-bearing threads only, one headline arc, terse
 
+Work recorded with brag learn is never marked as a win. me and manager list it where it falls, as "✗ <id> (failed)". skip and exec leave it out, print an Omitted: line counting it, and end the framing directive with a line telling the LLM to say so. A user profile leaves failures out only when its candor is exactly promotional; any other value lists them.
+
 Each audience carries a default window; an explicit window flag overrides it. Windows are CALENDAR periods (like brag impact), mutually exclusive:
   --quarter / --month / --year / --since D   (D: YYYY-MM-DD or Nd/Nw/Nm)
 
@@ -195,7 +197,8 @@ func runStory(cmd *cobra.Command, _ []string) error {
 		return fmt.Errorf("list entries: %w", err)
 	}
 
-	threads := story.BuildThreads(entries, story.ThreadOptionsFromProfile(profile, theme))
+	shown, omitted := story.OmitFailures(entries, profile)
+	threads := story.BuildThreads(shown, story.ThreadOptionsFromProfile(profile, theme))
 	throughline := story.BuildThroughline(threads)
 
 	filtersMD, filtersJSON := echoFiltersForStory(cmd)
@@ -205,6 +208,7 @@ func runStory(cmd *cobra.Command, _ []string) error {
 		Filters:         filtersMD,
 		FiltersJSON:     filtersJSON,
 		EntriesInWindow: len(entries),
+		OmittedFailures: omitted,
 		Now:             now,
 		Threads:         threads,
 		Throughline:     throughline,
````

### §4. Test code, new and modified

`candor_test.go` is new. `bundle_test.go` carries the two planned rewrites, `story_help_test.go` one new test, and `learn_test.go` the helper change plus one new e2e test.

````diff
diff --git a/internal/story/candor_test.go b/internal/story/candor_test.go
new file mode 100644
index 0000000..029f1e9
--- /dev/null
+++ b/internal/story/candor_test.go
@@ -0,0 +1,529 @@
+package story
+
+import (
+	"encoding/json"
+	"os"
+	"path/filepath"
+	"strings"
+	"testing"
+	"time"
+
+	"github.com/jysf/bragfile000/internal/storage"
+)
+
+// failureStoryFixture holds three recorded failures (SPEC-094). Types are the
+// literal "failed", never aggregate.FailureType, so a change to the persisted
+// value fires these goldens rather than moving with them (SPEC-086 LD8).
+//   - 2 (alpha) is a failure WITH an impact: still an impact beat (DEC-054).
+//   - 4 (beta) is a failure with NO impact: labelled, no impact line.
+//   - 5 (delta) is delta's ONLY impact beat, so a promotional profile must
+//     fold delta — and still count 5 in its Omitted: line.
+//   - 7 (gamma) is typed "Failed": a near-miss, not a failure (DEC-050 rule 1).
+var failureStoryFixture = []storage.Entry{
+	{ID: 1, Title: "shipped the cache", Project: "alpha", Type: "shipped",
+		Impact:    "cut p95 latency 40%",
+		CreatedAt: time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)},
+	{ID: 2, Title: "tried a worker pool", Project: "alpha", Type: "failed",
+		Impact:    "cost two days",
+		CreatedAt: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)},
+	{ID: 3, Title: "onboarding guide", Project: "beta", Type: "shipped",
+		Impact:    "onboarding down to 1 day",
+		CreatedAt: time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)},
+	{ID: 4, Title: "abandoned the rewrite", Project: "beta", Type: "failed",
+		CreatedAt: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)},
+	{ID: 5, Title: "migrated to the wrong queue", Project: "delta", Type: "failed",
+		Impact:    "lost a week",
+		CreatedAt: time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)},
+	{ID: 6, Title: "delta notes", Project: "delta", Type: "learned",
+		CreatedAt: time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)},
+	{ID: 7, Title: "gamma near-miss", Project: "gamma", Type: "Failed",
+		Impact:    "shipped anyway",
+		CreatedAt: time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC)},
+}
+
+// candorBundleOpts is the CLI's pipeline over failureStoryFixture, minus the
+// store: OmitFailures, then BuildThreads, then the renderer's options.
+func candorBundleOpts(t *testing.T, audience string) StoryOptions {
+	t.Helper()
+	p, err := LoadProfile(audience)
+	if err != nil {
+		t.Fatalf("LoadProfile(%q): %v", audience, err)
+	}
+	shown, omitted := OmitFailures(failureStoryFixture, p)
+	threads := BuildThreads(shown, ThreadOptionsFromProfile(p, ""))
+	return StoryOptions{
+		Audience:        audience,
+		Scope:           "year",
+		Filters:         "(none)",
+		EntriesInWindow: len(failureStoryFixture),
+		OmittedFailures: omitted,
+		Now:             storyFixedNow,
+		Threads:         threads,
+		Throughline:     BuildThroughline(threads),
+		Directive:       mustDirective(t, p.Directive),
+	}
+}
+
+// TestOmitFailures_OnlyExactPromotionalOmits ▲ SPEC-094 LD1. Candor is an
+// unvalidated string in a user profile, so the rule keys on exactly
+// "promotional" and every other value keeps every entry. The positive case is
+// in the same test: without it, a rule that never omits passes the rest.
+func TestOmitFailures_OnlyExactPromotionalOmits(t *testing.T) {
+	shown, omitted := OmitFailures(failureStoryFixture, Profile{Candor: "promotional"})
+	if omitted != 3 {
+		t.Errorf("promotional: omitted = %d, want 3 (ids 2, 4, 5)", omitted)
+	}
+	if got := entryIDs(shown); !equalInt64s(got, []int64{1, 3, 6, 7}) {
+		t.Errorf("promotional: shown = %v, want [1 3 6 7] (7 is a near-miss, not a failure)", got)
+	}
+
+	for _, candor := range []string{"candid", "", "Promotional", "PROMOTIONAL", "promotional ", "promo", "promotion"} {
+		shown, omitted := OmitFailures(failureStoryFixture, Profile{Candor: candor})
+		if omitted != 0 {
+			t.Errorf("candor %q: omitted = %d, want 0 — only the exact value omits", candor, omitted)
+		}
+		if got := entryIDs(shown); !equalInt64s(got, entryIDs(failureStoryFixture)) {
+			t.Errorf("candor %q: shown = %v, want every entry", candor, got)
+		}
+	}
+}
+
+// TestOmitFailures_RunsBeforeTheFold ▲ SPEC-094 LD2. delta's only impact beat
+// is a failure. Threaded as-is, exec keeps delta (the control); omitted first,
+// delta folds under impact_threads_only and its failure is still counted.
+func TestOmitFailures_RunsBeforeTheFold(t *testing.T) {
+	if got := threadNames(BuildThreads(failureStoryFixture, execThreadOpts)); !containsString(got, "delta") {
+		t.Fatalf("control: exec over the unfiltered fixture should keep delta, got %v", got)
+	}
+	shown, omitted := OmitFailures(failureStoryFixture, Profile{Candor: CandorPromotional})
+	got := threadNames(BuildThreads(shown, execThreadOpts))
+	if containsString(got, "delta") {
+		t.Errorf("exec after OmitFailures still carries delta, whose only impact beat was a failure: %v", got)
+	}
+	if omitted != 3 {
+		t.Errorf("omitted = %d, want 3: delta's failure folds with its thread and is still counted", omitted)
+	}
+}
+
+// TestBuildThreads_AFailureWithImpactIsStillAnImpactBeat ▲ SPEC-094 LD5
+// (framing's Question 2). is_impact_beat keeps meaning "non-empty impact", so
+// no count changes what it counts (DEC-048); IsFailure is what renders it apart.
+func TestBuildThreads_AFailureWithImpactIsStillAnImpactBeat(t *testing.T) {
+	threads := BuildThreads(failureStoryFixture, meThreadOpts)
+	type flags struct{ impact, failure bool }
+	want := map[int64]flags{
+		1: {true, false}, 2: {true, true}, 3: {true, false}, 4: {false, true},
+		5: {true, true}, 6: {false, false}, 7: {true, false},
+	}
+	seen := 0
+	for _, thr := range threads {
+		for _, b := range thr.Beats {
+			seen++
+			if got := (flags{b.IsImpactBeat, b.IsFailure}); got != want[b.ID] {
+				t.Errorf("beat %d: {IsImpactBeat IsFailure} = %v, want %v", b.ID, got, want[b.ID])
+			}
+		}
+	}
+	if seen != 7 {
+		t.Fatalf("me kept %d beats, want all 7", seen)
+	}
+	for _, a := range BuildThroughline(threads).Arcs {
+		if a.Thread == "delta" && a.ImpactBeatCount != 1 {
+			t.Errorf("delta impact_beat_count = %d, want 1: its failure carries an impact", a.ImpactBeatCount)
+		}
+	}
+}
+
+// TestStoryPackage_CallsTheSharedPredicates ▲ SPEC-094 LD6. The impact and
+// failure rules live in internal/aggregate; a copy restated here is one a
+// change there leaves behind with every behavioural test green, which is what
+// thread.go:135 and :87 were until SPEC-094. Scans non-test sources only.
+func TestStoryPackage_CallsTheSharedPredicates(t *testing.T) {
+	forbidden := []string{`Impact != ""`, `Impact == ""`, `"failed"`}
+	files, err := filepath.Glob("*.go")
+	if err != nil {
+		t.Fatal(err)
+	}
+	scanned := 0
+	for _, f := range files {
+		if strings.HasSuffix(f, "_test.go") {
+			continue
+		}
+		b, err := os.ReadFile(f)
+		if err != nil {
+			t.Fatal(err)
+		}
+		scanned++
+		for _, needle := range forbidden {
+			if strings.Contains(string(b), needle) {
+				t.Errorf("%s restates a predicate (%s); call aggregate.HasImpact / aggregate.IsFailure instead", f, needle)
+			}
+		}
+	}
+	if scanned < 4 {
+		t.Fatalf("scanned %d source files, want at least 4 (bundle, embed, profile, thread)", scanned)
+	}
+}
+
+// TestToStoryMarkdown_CandidLabelsFailuresGolden ▲ SPEC-094 LD3/LD5. A candid
+// profile lists every failure where it falls, as "✗ <id> (failed)", with its
+// impact line when it has one. ★ never marks a failure. There is no Omitted:
+// line and the directive is the asset, unchanged.
+func TestToStoryMarkdown_CandidLabelsFailuresGolden(t *testing.T) {
+	withOverrideDir(t, t.TempDir())
+	got, err := ToStoryMarkdown(candorBundleOpts(t, "me"))
+	if err != nil {
+		t.Fatalf("unexpected error: %v", err)
+	}
+	want := `# Bragfile Story
+
+Generated: 2026-07-06T12:00:00Z
+Scope: year
+Audience: me
+Filters: (none)
+Threads: 4
+Beats: 7/7
+
+## Threads
+
+### alpha
+
+- ★ 1: shipped the cache
+  cut p95 latency 40%
+- ✗ 2 (failed): tried a worker pool
+  cost two days
+
+### beta
+
+- ★ 3: onboarding guide
+  onboarding down to 1 day
+- ✗ 4 (failed): abandoned the rewrite
+
+### delta
+
+- ✗ 5 (failed): migrated to the wrong queue
+  lost a week
+- · 6: delta notes
+
+### gamma
+
+- ★ 7: gamma near-miss
+  shipped anyway
+
+## Throughline (skeleton)
+
+- alpha [initiative]: 2 beats, 2 with impact (2026-02-01 → 2026-03-01)
+- beta [initiative]: 2 beats, 1 with impact (2026-04-01 → 2026-05-01)
+- delta [initiative]: 2 beats, 1 with impact (2026-06-01 → 2026-06-10)
+- gamma [initiative]: 1 beat, 1 with impact (2026-06-20 → 2026-06-20)
+
+## Framing directive
+
+` + mustDirectiveTrimmed(t, "me.md")
+	if string(got) != want {
+		t.Errorf("candid markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
+	}
+}
+
+// TestToStoryMarkdown_PromotionalOmitsWithNoteAndClauseGolden ▲ SPEC-094
+// LD1/LD2/LD3/LD4. exec omits the three failures, says so under Beats:, folds
+// delta, and ends its directive with the fixed clause. The near-miss (7) stays.
+func TestToStoryMarkdown_PromotionalOmitsWithNoteAndClauseGolden(t *testing.T) {
+	withOverrideDir(t, t.TempDir())
+	got, err := ToStoryMarkdown(candorBundleOpts(t, "exec"))
+	if err != nil {
+		t.Fatalf("unexpected error: %v", err)
+	}
+	want := `# Bragfile Story
+
+Generated: 2026-07-06T12:00:00Z
+Scope: year
+Audience: exec
+Filters: (none)
+Threads: 3
+Beats: 3/7
+Omitted: 3 recorded failures, not listed for this audience (brag list --type failed)
+
+## Threads
+
+### alpha
+
+- ★ 1: shipped the cache
+  cut p95 latency 40%
+
+### beta
+
+- ★ 3: onboarding guide
+  onboarding down to 1 day
+
+### gamma
+
+- ★ 7: gamma near-miss
+  shipped anyway
+
+## Throughline (skeleton)
+
+- alpha [initiative]: 1 beat, 1 with impact (2026-02-01 → 2026-02-01)
+- beta [initiative]: 1 beat, 1 with impact (2026-04-01 → 2026-04-01)
+- gamma [initiative]: 1 beat, 1 with impact (2026-06-20 → 2026-06-20)
+
+## Framing directive
+
+` + mustDirectiveTrimmed(t, "exec.md") + `
+
+This bundle omits 3 recorded failures for this audience. End with one line that says so; do not drop it.`
+	if string(got) != want {
+		t.Errorf("promotional markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
+	}
+}
+
+// TestToStoryJSON_PromotionalCountsWhatItOmitted ▲ SPEC-094 LD3/LD4. The JSON
+// carries the count as omitted_failure_count, directly after filters; no beat
+// is a failure; and framing_directive carries the same clause the markdown
+// does. The candid half is the pair: the key is present at 0, the failures
+// are in the beats, and the directive is the asset byte for byte.
+func TestToStoryJSON_PromotionalCountsWhatItOmitted(t *testing.T) {
+	withOverrideDir(t, t.TempDir())
+	type envelope struct {
+		OmittedFailureCount *int `json:"omitted_failure_count"`
+		Threads             []struct {
+			Beats []struct {
+				ID   int64  `json:"id"`
+				Type string `json:"type"`
+			} `json:"beats"`
+		} `json:"threads"`
+		FramingDirective string `json:"framing_directive"`
+	}
+	decode := func(audience string) (envelope, string) {
+		body, err := ToStoryJSON(candorBundleOpts(t, audience))
+		if err != nil {
+			t.Fatalf("%s: %v", audience, err)
+		}
+		var env envelope
+		if err := json.Unmarshal(body, &env); err != nil {
+			t.Fatalf("%s: unmarshal: %v\n%s", audience, err, body)
+		}
+		return env, string(body)
+	}
+	failedBeats := func(env envelope) []int64 {
+		var out []int64
+		for _, th := range env.Threads {
+			for _, b := range th.Beats {
+				if b.Type == "failed" {
+					out = append(out, b.ID)
+				}
+			}
+		}
+		return out
+	}
+
+	execEnv, execBody := decode("exec")
+	if execEnv.OmittedFailureCount == nil || *execEnv.OmittedFailureCount != 3 {
+		t.Errorf("exec omitted_failure_count = %v, want 3", execEnv.OmittedFailureCount)
+	}
+	if got := failedBeats(execEnv); len(got) != 0 {
+		t.Errorf("exec JSON still carries failure beats %v", got)
+	}
+	clause := "This bundle omits 3 recorded failures for this audience. End with one line that says so; do not drop it.\n"
+	if !strings.HasSuffix(execEnv.FramingDirective, "\n\n"+clause) {
+		t.Errorf("exec framing_directive should end with the clause as its own paragraph, got:\n%q", execEnv.FramingDirective)
+	}
+	if !strings.Contains(execBody, "\"filters\": {},\n  \"omitted_failure_count\": 3,\n  \"threads\": [") {
+		t.Errorf("omitted_failure_count must sit between filters and threads:\n%s", execBody)
+	}
+
+	meEnv, _ := decode("me")
+	if meEnv.OmittedFailureCount == nil || *meEnv.OmittedFailureCount != 0 {
+		t.Errorf("me omitted_failure_count = %v, want 0 and present", meEnv.OmittedFailureCount)
+	}
+	if got := failedBeats(meEnv); !equalInt64s(got, []int64{2, 4, 5}) {
+		t.Errorf("me JSON failure beats = %v, want [2 4 5]", got)
+	}
+	if meEnv.FramingDirective != mustDirective(t, "me.md") {
+		t.Errorf("me framing_directive must be the asset unchanged, got:\n%q", meEnv.FramingDirective)
+	}
+}
+
+// TestToStory_EveryBeatOmittedStillSaysWhy ▲ SPEC-094 LD4 (the empty state,
+// DEC-014 part 4). When every in-window entry is an omitted failure — `brag
+// story --audience exec --type failed` — the threads are empty and the note
+// is the only thing that says why. The zero-omission run of the same empty
+// shape is the pair: no Omitted: line, and the directive is the asset.
+func TestToStory_EveryBeatOmittedStillSaysWhy(t *testing.T) {
+	p := Profile{Candor: CandorPromotional}
+	onlyFailures := []storage.Entry{failureStoryFixture[1], failureStoryFixture[4]}
+	shown, omitted := OmitFailures(onlyFailures, p)
+	threads := BuildThreads(shown, execThreadOpts)
+	opts := StoryOptions{
+		Audience:        "exec",
+		Scope:           "year",
+		Filters:         "--type failed",
+		FiltersJSON:     map[string]string{"type": "failed"},
+		EntriesInWindow: len(onlyFailures),
+		OmittedFailures: omitted,
+		Now:             storyFixedNow,
+		Threads:         threads,
+		Throughline:     BuildThroughline(threads),
+		Directive:       mustDirective(t, "exec.md"),
+	}
+	md, err := ToStoryMarkdown(opts)
+	if err != nil {
+		t.Fatalf("markdown: %v", err)
+	}
+	wantHead := "# Bragfile Story\n\nGenerated: 2026-07-06T12:00:00Z\nScope: year\nAudience: exec\nFilters: --type failed\nThreads: 0\nBeats: 0/2\nOmitted: 2 recorded failures, not listed for this audience (brag list --type failed)\n\n## Framing directive\n"
+	if !strings.HasPrefix(string(md), wantHead) {
+		t.Errorf("empty-state head mismatch:\n--- got ---\n%s\n--- want prefix ---\n%s", md, wantHead)
+	}
+	if !strings.HasSuffix(string(md), "\n\nThis bundle omits 2 recorded failures for this audience. End with one line that says so; do not drop it.") {
+		t.Errorf("empty-state bundle must still end with the clause:\n%s", md)
+	}
+	body, err := ToStoryJSON(opts)
+	if err != nil {
+		t.Fatalf("json: %v", err)
+	}
+	if !strings.Contains(string(body), `"omitted_failure_count": 2,`) || !strings.Contains(string(body), `"threads": [],`) {
+		t.Errorf("empty-state JSON should carry the count and an empty threads array:\n%s", body)
+	}
+
+	opts.OmittedFailures = 0
+	md, err = ToStoryMarkdown(opts)
+	if err != nil {
+		t.Fatalf("markdown (no omission): %v", err)
+	}
+	if strings.Contains(string(md), "\nOmitted: ") {
+		t.Errorf("zero omitted must render no Omitted: line:\n%s", md)
+	}
+	if !strings.HasSuffix(string(md), mustDirectiveTrimmed(t, "exec.md")) {
+		t.Errorf("zero omitted must leave the directive as the asset:\n%s", md)
+	}
+}
+
+// TestFramingDirective_ClauseAloneWhenDirectiveEmpty ▲ SPEC-094 LD4. A
+// promotional user profile may carry no directive at all; the clause is then
+// the whole directive, so the instruction still reaches the model. Also pins
+// the singular form on both the note and the clause.
+func TestFramingDirective_ClauseAloneWhenDirectiveEmpty(t *testing.T) {
+	opts := StoryOptions{
+		Audience:        "mine",
+		Scope:           "year",
+		Filters:         "(none)",
+		EntriesInWindow: 1,
+		OmittedFailures: 1,
+		Now:             storyFixedNow,
+		Threads:         []Thread{},
+		Throughline:     BuildThroughline(nil),
+		Directive:       "",
+	}
+	md, err := ToStoryMarkdown(opts)
+	if err != nil {
+		t.Fatalf("markdown: %v", err)
+	}
+	want := "# Bragfile Story\n\nGenerated: 2026-07-06T12:00:00Z\nScope: year\nAudience: mine\nFilters: (none)\nThreads: 0\nBeats: 0/1\nOmitted: 1 recorded failure, not listed for this audience (brag list --type failed)\n\n## Framing directive\n\nThis bundle omits 1 recorded failure for this audience. End with one line that says so; do not drop it."
+	if string(md) != want {
+		t.Errorf("clause-only directive mismatch:\n--- got ---\n%s\n--- want ---\n%s", md, want)
+	}
+	body, err := ToStoryJSON(opts)
+	if err != nil {
+		t.Fatalf("json: %v", err)
+	}
+	var env struct {
+		FramingDirective string `json:"framing_directive"`
+	}
+	if err := json.Unmarshal(body, &env); err != nil {
+		t.Fatalf("unmarshal: %v", err)
+	}
+	if env.FramingDirective != "This bundle omits 1 recorded failure for this audience. End with one line that says so; do not drop it.\n" {
+		t.Errorf("framing_directive = %q, want the clause alone", env.FramingDirective)
+	}
+}
+
+// TestLoadProfile_OnlyExactPromotionalOmitsFailures ▲ SPEC-094 LD1. The
+// bundled data decides which built-ins omit (exec, skip) and which label (me,
+// manager); a user profile omits only on the exact value, and a promotional
+// user profile's own directive file still gets the clause appended.
+func TestLoadProfile_OnlyExactPromotionalOmitsFailures(t *testing.T) {
+	withOverrideDir(t, t.TempDir())
+	for name, want := range map[string]bool{"exec": true, "skip": true, "me": false, "manager": false} {
+		p, err := LoadProfile(name)
+		if err != nil {
+			t.Fatalf("LoadProfile(%q): %v", name, err)
+		}
+		if p.OmitsFailures() != want {
+			t.Errorf("bundled %s: OmitsFailures = %v, want %v (candor %q)", name, p.OmitsFailures(), want, p.Candor)
+		}
+	}
+
+	dir := t.TempDir()
+	withOverrideDir(t, dir)
+	directivePath := filepath.Join(dir, "board.md")
+	if err := os.WriteFile(directivePath, []byte("Pitch this to the board.\n"), 0o644); err != nil {
+		t.Fatal(err)
+	}
+	profiles := map[string]string{
+		"board":  "candor: promotional\ndirective: " + directivePath + "\n",
+		"typo":   "candor: Promotional\n",
+		"promo":  "candor: promo\n",
+		"silent": "thread_order: initiative\n",
+	}
+	for name, body := range profiles {
+		if err := os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(body), 0o644); err != nil {
+			t.Fatal(err)
+		}
+	}
+	for name, want := range map[string]bool{"board": true, "typo": false, "promo": false, "silent": false} {
+		p, err := LoadProfile(name)
+		if err != nil {
+			t.Fatalf("LoadProfile(%q): %v", name, err)
+		}
+		if p.OmitsFailures() != want {
+			t.Errorf("override %s (candor %q): OmitsFailures = %v, want %v", name, p.Candor, p.OmitsFailures(), want)
+		}
+	}
+
+	board, _ := LoadProfile("board")
+	directive, err := ResolveDirective(board)
+	if err != nil {
+		t.Fatalf("ResolveDirective: %v", err)
+	}
+	md, err := ToStoryMarkdown(StoryOptions{
+		Audience: "board", Scope: "year", Filters: "(none)", EntriesInWindow: 2,
+		OmittedFailures: 2, Now: storyFixedNow, Directive: directive,
+	})
+	if err != nil {
+		t.Fatalf("markdown: %v", err)
+	}
+	if !strings.HasSuffix(string(md), "## Framing directive\n\nPitch this to the board.\n\nThis bundle omits 2 recorded failures for this audience. End with one line that says so; do not drop it.") {
+		t.Errorf("a user directive file must get the clause appended:\n%s", md)
+	}
+}
+
+// --- helpers ---
+
+func entryIDs(entries []storage.Entry) []int64 {
+	out := make([]int64, len(entries))
+	for i, e := range entries {
+		out[i] = e.ID
+	}
+	return out
+}
+
+func equalInt64s(a, b []int64) bool {
+	if len(a) != len(b) {
+		return false
+	}
+	for i := range a {
+		if a[i] != b[i] {
+			return false
+		}
+	}
+	return true
+}
+
+func containsString(list []string, s string) bool {
+	for _, v := range list {
+		if v == s {
+			return true
+		}
+	}
+	return false
+}
````

````diff
diff --git a/internal/story/bundle_test.go b/internal/story/bundle_test.go
index 8d238f8..b9aae09 100644
--- a/internal/story/bundle_test.go
+++ b/internal/story/bundle_test.go
@@ -168,6 +168,7 @@ func TestToStoryJSON_MeProfile_ShapeGolden(t *testing.T) {
   "scope": "year",
   "audience": "me",
   "filters": {},
+  "omitted_failure_count": 0,
   "threads": [
     {
       "thread": "alpha",
@@ -381,4 +382,9 @@ Beats: 0/0
 	if !strings.Contains(string(jsonBody), `"threads": []`) {
 		t.Errorf("expected empty threads array literal:\n%s", jsonBody)
 	}
+	// SPEC-094: the omission count is always present, 0 on an empty window
+	// (DEC-014 part 4).
+	if !strings.Contains(string(jsonBody), `"omitted_failure_count": 0,`) {
+		t.Errorf("expected omitted_failure_count 0 on an empty window:\n%s", jsonBody)
+	}
 }
````

````diff
diff --git a/internal/cli/story_help_test.go b/internal/cli/story_help_test.go
index 0de1b1e..5fceb25 100644
--- a/internal/cli/story_help_test.go
+++ b/internal/cli/story_help_test.go
@@ -42,3 +42,14 @@ func TestStoryCmd_HelpListsAllBuiltInAudiences(t *testing.T) {
 		t.Errorf("--audience usage must keep the user-profile affordance, got %q", usage)
 	}
 }
+
+// TestStoryCmd_HelpStatesTheFailureRule ▲ SPEC-094 LD8. `brag story --help`
+// says what happens to work recorded with brag learn, per candor, in one
+// literal sentence set — the only place a user reading the CLI learns that a
+// promotional audience leaves failures out.
+func TestStoryCmd_HelpStatesTheFailureRule(t *testing.T) {
+	want := `Work recorded with brag learn is never marked as a win. me and manager list it where it falls, as "✗ <id> (failed)". skip and exec leave it out, print an Omitted: line counting it, and end the framing directive with a line telling the LLM to say so. A user profile leaves failures out only when its candor is exactly promotional; any other value lists them.`
+	if !strings.Contains(NewStoryCmd().Long, want) {
+		t.Errorf("story --help Long is missing the failure rule:\n%s", NewStoryCmd().Long)
+	}
+}
````

````diff
diff --git a/internal/cli/learn_test.go b/internal/cli/learn_test.go
index 4abe209..0ffe2a5 100644
--- a/internal/cli/learn_test.go
+++ b/internal/cli/learn_test.go
@@ -168,16 +168,17 @@ func TestLearnCmd_EmptyTitleIsUserError(t *testing.T) {
 }
 
 // runDigestCorpus runs one brag invocation against dbPath on a fresh root
-// carrying the two writers (add, learn) and the two digests that section
-// failures (impact, wrapped). A fresh root per call, so no flag value leaks
-// from one invocation into the next.
+// carrying the two writers (add, learn), the two digests that section
+// failures (impact, wrapped), and story, which labels or omits them
+// (SPEC-094). A fresh root per call, so no flag value leaks from one
+// invocation into the next.
 func runDigestCorpus(t *testing.T, dbPath string, args ...string) string {
 	t.Helper()
 	t.Setenv("BRAGFILE_DB", "")
 	addStderrIsTTY = func() bool { return false }
 	t.Cleanup(func() { addStderrIsTTY = defaultStderrIsTTY })
 	root := NewRootCmd("test")
-	root.AddCommand(NewAddCmd(), NewLearnCmd(), NewImpactCmd(), NewWrappedCmd())
+	root.AddCommand(NewAddCmd(), NewLearnCmd(), NewImpactCmd(), NewWrappedCmd(), NewStoryCmd())
 	var outBuf, errBuf bytes.Buffer
 	root.SetOut(&outBuf)
 	root.SetErr(&errBuf)
@@ -301,3 +302,58 @@ func TestLearnCmd_WrappedSectionsWhatItWrote(t *testing.T) {
 		t.Errorf("## What didn't work is missing the brag learn entry %s with its impact:\n%s", failID, md)
 	}
 }
+
+// TestLearnCmd_StoryLabelsOrOmitsWhatItWrote ▲ SPEC-094 LD1/LD3/LD4 — the
+// writer-to-reader check on brag story, through a real store and with no
+// constant in sight. A candid audience lists the brag learn entry as
+// "✗ <id> (failed)" with its impact, beside the brag add entry's ★. A
+// promotional one leaves it out, and both its Omitted: line and its closing
+// clause say so; its JSON counts it.
+func TestLearnCmd_StoryLabelsOrOmitsWhatItWrote(t *testing.T) {
+	dbPath := filepath.Join(t.TempDir(), "test.db")
+	winID := strings.TrimSpace(runDigestCorpus(t, dbPath, "add", "-t", "shipped the cache", "-p", "alpha", "-k", "shipped", "-i", "cut p95 40%"))
+	failID := strings.TrimSpace(runDigestCorpus(t, dbPath, "learn", "-t", "tried a worker pool", "-p", "alpha", "-i", "cost two days"))
+
+	me := runDigestCorpus(t, dbPath, "story", "--audience", "me", "--since", "2000-01-01")
+	threads := markdownSection(me, "Threads")
+	if !strings.Contains(threads, "- ★ "+winID+": shipped the cache\n  cut p95 40%") {
+		t.Errorf("me should list the brag add entry %s as ★:\n%s", winID, me)
+	}
+	if !strings.Contains(threads, "- ✗ "+failID+" (failed): tried a worker pool\n  cost two days") {
+		t.Errorf("me should label the brag learn entry %s as failed, with its impact:\n%s", failID, me)
+	}
+	if strings.Contains(me, "- ★ "+failID+":") || strings.Contains(me, "\nOmitted: ") {
+		t.Errorf("me must neither star the failure nor omit it:\n%s", me)
+	}
+
+	exec := runDigestCorpus(t, dbPath, "story", "--audience", "exec", "--since", "2000-01-01")
+	if !strings.Contains(markdownSection(exec, "Threads"), "- ★ "+winID+": shipped the cache") {
+		t.Errorf("exec should still list the brag add entry %s:\n%s", winID, exec)
+	}
+	if strings.Contains(exec, " "+failID+": ") || strings.Contains(exec, " "+failID+" (failed)") {
+		t.Errorf("exec must not list the brag learn entry %s:\n%s", failID, exec)
+	}
+	if !strings.Contains(exec, "\nBeats: 1/2\nOmitted: 1 recorded failure, not listed for this audience (brag list --type failed)\n") {
+		t.Errorf("exec should count the omitted failure under Beats:\n%s", exec)
+	}
+	if !strings.HasSuffix(strings.TrimRight(exec, "\n"), "\n\nThis bundle omits 1 recorded failure for this audience. End with one line that says so; do not drop it.") {
+		t.Errorf("exec should end with the omission clause:\n%s", exec)
+	}
+
+	// --print-directive reads no store, so it prints the asset as authored:
+	// the clause depends on the window, and there is none.
+	directive := runDigestCorpus(t, dbPath, "story", "--audience", "exec", "--print-directive")
+	if !strings.Contains(directive, "business impact") || strings.Contains(directive, "This bundle omits") {
+		t.Errorf("--print-directive should print exec's directive without the clause:\n%s", directive)
+	}
+
+	var env struct {
+		OmittedFailureCount *int `json:"omitted_failure_count"`
+	}
+	if err := json.Unmarshal([]byte(runDigestCorpus(t, dbPath, "story", "--audience", "exec", "--since", "2000-01-01", "--format", "json")), &env); err != nil {
+		t.Fatalf("json unmarshal: %v", err)
+	}
+	if env.OmittedFailureCount == nil || *env.OmittedFailureCount != 1 {
+		t.Errorf("exec omitted_failure_count = %v, want 1", env.OmittedFailureCount)
+	}
+}
````

### §5. `docs/api-contract.md`, the `brag story` section

Six hunks, all inside the section that `AE1` scopes to.

````diff
diff --git a/docs/api-contract.md b/docs/api-contract.md
index c8a05a4..6e17cb7 100644
--- a/docs/api-contract.md
+++ b/docs/api-contract.md
@@ -858,20 +858,27 @@ Document structure (markdown):
 
 - **Provenance:** `Generated:` (RFC3339), `Scope:` (the resolved window
   token), `Audience:` (the requested name), `Filters:` (echoed flags or
-  `(none)`), `Threads: <n>`, and a `Beats: <shown>/<in-window>` tally.
+  `(none)`), `Threads: <n>`, and a `Beats: <shown>/<in-window>` tally. When
+  a promotional profile left recorded failures out, one more line follows,
+  and it is absent otherwise (`1 recorded failure` for one):
+  `Omitted: <n> recorded failures, not listed for this audience (brag list --type failed)`.
 - **Threads** (`## Threads`; omitted on an empty corpus per DEC-014):
   per-thread `### <thread>` blocks. Each beat renders `- ★ <id>: <title>`
   with an indented impact line when it carries a non-empty `impact`, or
   `- · <id>: <title>` when it does not (the ★/· markers are the visible
-  "so what" signal).
+  "so what" signal). A recorded failure renders `- ✗ <id> (failed): <title>`
+  instead, with its impact line when it has one, so ★ never marks a failure.
 - **Throughline** (`## Throughline (skeleton)`; omitted on an empty
   corpus): one line per thread — `- <thread> [<kind>]: <n> beats, <m>
   with impact (<first> → <last>)` — the ordered thread refs + span +
   beat/impact counts. The skeleton is deterministic; the LLM (via the
   directive) finds the actual arc.
 - **Framing directive** (`## Framing directive`): the audience's directive
-  text appended verbatim. Renders even on an empty corpus; omitted only
-  when the resolved directive itself is empty.
+  text appended verbatim. When failures were omitted, bragfile appends one
+  fixed paragraph after it, whatever file the directive came from:
+  `This bundle omits <n> recorded failures for this audience. End with one line that says so; do not drop it.`
+  Renders even on an empty corpus; omitted only when the resolved directive
+  is empty and nothing was omitted.
 
 Audiences are **data-driven shaping profiles**, NOT a Go enum: `me`/`exec`
 load from bundled `embed.FS` assets; a user `<name>.yaml` in the
@@ -894,6 +901,19 @@ are the middle):
   impact-less beats DROPPED, small threads folded, one headline arc
   (threads ordered impact-beat-count DESC); default window `quarter`.
 
+**Recorded failures** (entries whose `type` is exactly `failed`, as `brag
+learn` writes them) are never marked as a win. A profile whose `candor` is
+exactly `promotional` — `skip` and `exec` — leaves them out before threads are
+built and folded, and says how many in the `Omitted:` line, in
+`omitted_failure_count`, and in the paragraph appended to the directive.
+Every other profile — `me`, `manager`, and a user profile whose `candor` is
+anything else, empty or misspelled — lists them where they fall, labelled
+`✗ <id> (failed)`. A failure that carries an impact still counts as an impact
+beat (`is_impact_beat`, `<m> with impact`); only its marker differs. Locked by
+[DEC-054](../decisions/DEC-054-story-candor-decides-what-a-failure-renders-as.md),
+which implements
+[DEC-050](../decisions/DEC-050-a-failure-is-never-rendered-as-a-win.md) row 4.
+
 Flags:
 
 - `--audience <name>` is REQUIRED. An unknown audience (no bundled
@@ -915,12 +935,15 @@ Flags:
   after the initiative threads, grouping every in-window entry carrying
   that tag, time-ordered. Not subject to fold/drop (an explicit opt-in).
 - `--format markdown|json` defaults to `markdown`. The JSON envelope
-  extends DEC-014 with `audience`, `threads` (each `{thread, kind, span,
-  beats:[...]}`), `throughline` (`{arcs:[...]}`), and `framing_directive`,
+  extends DEC-014 with `audience`, `omitted_failure_count` (an integer,
+  always present, `0` when nothing was omitted), `threads` (each `{thread,
+  kind, span, beats:[...]}`), `throughline` (`{arcs:[...]}`), and
+  `framing_directive` (with the appended paragraph when one applies),
   2-space indent. Each beat is a 7-key projection `{id, title, project,
   type, impact, is_impact_beat, created_at}`.
 - `--print-directive` prints ONLY the resolved framing directive to
-  stdout and exits 0 — no window, no DB read.
+  stdout and exits 0 — no window, no DB read, and so never the omission
+  paragraph, which depends on the window.
 - `--tag`/`--project`/`--type` compose with the window and echo into
   `filters`.
 
@@ -928,7 +951,9 @@ Empty-window: provenance renders (`Threads: 0`, `Beats: 0/0`); the
 `## Threads` and `## Throughline` sections are omitted from markdown; the
 `## Framing directive` section still renders. JSON renders `threads` `[]`,
 `throughline.arcs` `[]`, `framing_directive` the directive string,
-`filters` `{}`.
+`filters` `{}`, `omitted_failure_count` `0`. When every in-window entry is an
+omitted failure (`--audience exec --type failed`), the threads are empty in
+the same way and the `Omitted:` line is what says why.
 
 ### `brag memory [--query <text>] [--project <name>] [--budget <n>]` (STAGE-019)
 
````

### §6. `docs/tutorial.md` and `AGENTS.md`

One sentence each. The `AGENTS.md` hunk is a single long glossary line, as every §11 entry is.

````diff
diff --git a/docs/tutorial.md b/docs/tutorial.md
index 0a72f4f..8d5205a 100644
--- a/docs/tutorial.md
+++ b/docs/tutorial.md
@@ -645,7 +645,11 @@ Each audience carries a default window (`me` → year, `manager` → month,
 profile default — to the last-completed period (`--audience me --previous`
 = last year, `--audience exec --quarter --previous` = last quarter).
 Audiences are extensible profiles, not a fixed list — drop a `<name>.yaml`
-in your story-profiles directory to add one. Pipe it into an LLM to finish:
+in your story-profiles directory to add one. Work you recorded with `brag
+learn` is never marked as a win: `me` and `manager` list it where it falls as
+`✗ <id> (failed)`, while `skip` and `exec` leave it out, print an `Omitted:`
+line saying how many, and tell the LLM to end with one line that says so.
+Pipe it into an LLM to finish:
 
 ```bash
 brag story --audience exec --quarter | claude "weave these threads into one headline arc"
````

````diff
diff --git a/AGENTS.md b/AGENTS.md
index b7f0b01..35d52b8 100644
--- a/AGENTS.md
+++ b/AGENTS.md
@@ -291,7 +291,7 @@ DECs are stable; specs come and go. DECs don't reciprocally list specs.
 - **Store** — the `*storage.Store` Go type that owns the `*sql.DB` and all typed methods. The only package that imports a SQL driver.
 - **migration** — a single `NNNN_*.sql` file under `internal/storage/migrations/`, embedded into the binary, applied automatically in lexical order on `storage.Open`.
 - **export** — a one-shot dump of entries, either as a Markdown report (stdout or `--out file.md`) or as a portable SQLite file copy (via `VACUUM INTO`).
-- **learn** — `brag learn`: the capture verb for work that did **not** work, and the only place `entries.type` is pinned rather than free-form. Writes the reserved type value `failed` (DEC-049) from flag mode or editor mode; no `--type` flag, no `-k`, no `--json`, and — deliberately — no milestone nudge on stderr, because every milestone line is a congratulation. `entries.type` stays free-form everywhere else: the verb pins one value, it does not validate the field. Retrieval is the existing `brag list --type failed`; the DEC-044 memory line renders `[<project>/failed]`, so a failure is labelled as one on the agent-facing read surface for free. `brag impact` and `brag wrapped` list a failure that carries an impact under their own `## What didn't work` section, never among the wins (DEC-050, SPEC-086). The constant is `aggregate.FailureType`, read through `aggregate.IsFailure`. PROJ-008 STAGE-023 (SPEC-085).
+- **learn** — `brag learn`: the capture verb for work that did **not** work, and the only place `entries.type` is pinned rather than free-form. Writes the reserved type value `failed` (DEC-049) from flag mode or editor mode; no `--type` flag, no `-k`, no `--json`, and — deliberately — no milestone nudge on stderr, because every milestone line is a congratulation. `entries.type` stays free-form everywhere else: the verb pins one value, it does not validate the field. Retrieval is the existing `brag list --type failed`; the DEC-044 memory line renders `[<project>/failed]`, so a failure is labelled as one on the agent-facing read surface for free. `brag impact` and `brag wrapped` list a failure that carries an impact under their own `## What didn't work` section, never among the wins (DEC-050, SPEC-086). `brag story` labels one `✗ <id> (failed)` on a candid audience (`me`, `manager`) and leaves it out on a promotional one (`skip`, `exec`), with an `Omitted:` count and a directive clause telling the LLM to say so (DEC-054, SPEC-094). The constant is `aggregate.FailureType`, read through `aggregate.IsFailure`. PROJ-008 STAGE-023 (SPEC-085).
 - **review** — `brag review --week | --month`: prints recent entries grouped by project followed by three hard-coded reflection questions ("What pattern do you see in this period?", "What did you underestimate?", "What's missing here that should be?"). Markdown elides per-entry descriptions for compactness; JSON includes the full DEC-011 entry shape. Designed to be pasted into an external AI session for guided self-reflection. STAGE-004 (SPEC-019).
 - **summary** — a rule-based (non-LLM) aggregation of entries grouped by project/type over a rolling 7- or 30-day time window (`brag summary --range week|month`). STAGE-004.
 - **stats** — `brag stats`: six lifetime aggregations (total entries, entries/week rolling average, current streak, longest streak, top-5 most-common tags, top-5 most-common projects, corpus span). STAGE-004 (SPEC-020).
````

### §7. `CHANGELOG.md`

Under `## [Unreleased]` → `### Changed`, after SPEC-086's two entries and before `brag memory`'s.

````diff
diff --git a/CHANGELOG.md b/CHANGELOG.md
index 2eadabd..b7f0717 100644
--- a/CHANGELOG.md
+++ b/CHANGELOG.md
@@ -46,6 +46,21 @@ and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0
   still sums to `entries_with_impact`. A consumer that read every
   with-impact entry from `impact_by_project` or `impact_moments` now reads
   both keys.
+- **`brag story` no longer marks work that did not work as a win**
+  ([DEC-054](decisions/DEC-054-story-candor-decides-what-a-failure-renders-as.md)).
+  An entry recorded with `brag learn` rendered as `- ★ <id>:` on every
+  audience. `me` and `manager` now list it as `- ✗ <id> (failed):`. `skip`
+  and `exec`, whose profiles are `candor: promotional`, leave it out, print
+  `Omitted: <n> recorded failures, not listed for this audience` under
+  `Beats:`, and append one fixed line to the framing directive telling the
+  LLM to say so. That line is what makes the omission survive the model:
+  measured at design, a bare count reached the model's prose in 0 of 10
+  runs, and with the line the omission was stated in 25 of 25.
+- **Breaking for `skip` and `exec`: `brag story --format json` no longer
+  carries failure beats on a promotional audience.** Every envelope gains
+  `omitted_failure_count`, between `filters` and `threads`, always present
+  and `0` when nothing was left out. `is_impact_beat` and every count keep
+  their meaning: a failure with an impact is still an impact beat.
 - **`brag memory`'s headline count is now `Candidates: <N>`, not
   `Entries: <N>`** ([DEC-048](decisions/DEC-048-provenance-count-names-what-it-counted.md)).
   The number was never the corpus size and never a cap: it is the deduped
````

### §8. `scripts/test-docs.sh`, Group `AE`

Inserted immediately above `# ===== finalise =====`. It reuses Group `AD`'s `ad_section` and `assert_section_names`, which are defined earlier in the file.

````diff
diff --git a/scripts/test-docs.sh b/scripts/test-docs.sh
index 24a94f2..31f8f66 100755
--- a/scripts/test-docs.sh
+++ b/scripts/test-docs.sh
@@ -2282,6 +2282,32 @@ assert_section_names "AD4" "$(ad_section docs/tutorial.md '### Your year in brag
 assert_section_names "AD5" "$(grep -F -- '- **wrapped** —' AGENTS.md)" \
     "AGENTS.md, the wrapped glossary entry" "Impact moments → $ad_label"
 
+# ===== Group AE — story labels or omits a failure (SPEC-094 / DEC-054) =====
+#
+# SPEC-094 changes what `brag story` PRINTS: a recorded failure is labelled
+# `✗ <id> (failed)` on a candid audience and left out, with an `Omitted:` line,
+# `omitted_failure_count` and a directive clause, on a promotional one. The Go
+# suite pins the output; nothing but these ids pins the docs that describe it.
+# Each needle is scoped to its command's section with Group AD's helpers, so a
+# neighbouring command's text cannot satisfy it.
+
+# AE1 — the contract documents the label, the note, the JSON key and the record.
+assert_section_names "AE1" "$(ad_section docs/api-contract.md '### `brag story' '### ')" \
+    "docs/api-contract.md, the brag story section" \
+    '`- ✗ <id> (failed): <title>`' '`Omitted: <n> recorded failures' '`omitted_failure_count`' 'DEC-054'
+
+# AE2 — the tutorial tells a user where a promotional audience's failures went.
+assert_section_names "AE2" "$(ad_section docs/tutorial.md '### Tell your story' '### ')" \
+    "docs/tutorial.md, the brag story section" '`✗ <id> (failed)`' '`Omitted:`'
+
+# AE3 — the agent-facing glossary's learn entry names the third reader.
+assert_section_names "AE3" "$(grep -F -- '- **learn** —' AGENTS.md)" \
+    "AGENTS.md, the learn glossary entry" '`brag story` labels one' 'DEC-054'
+
+# AE4 — the unreleased changelog names the new JSON key.
+assert_section_names "AE4" "$(ad_section CHANGELOG.md '## [Unreleased]' '## [')" \
+    "CHANGELOG.md, the [Unreleased] section" '`omitted_failure_count`' 'DEC-054'
+
 # ===== finalise =====
 
 if [ "$FAIL_COUNT" -gt 0 ]; then
````

### §9. `docs/engineering-practices.md`: regenerate, never hand-edit

Paste `./scripts/inventory.sh`'s output between the markers. At design the
three build-time rows went from 80 / 843 / 205 to **81 / 856 / 209**. The
two DEC rows (53 and 5) are already on this branch.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `build/spec-094-story-honesty`
- **PR (if applicable):** none — the orchestrator opens it.
- **All acceptance criteria met?** yes (AC-1 through AC-14, all reproduced on
  a fresh read-only copy of the live corpus and the seeded Go fixtures; see
  the build report for the full numbers).
- **New decisions emitted:** none. DEC-054 was written at design; this cycle
  transcribes its mechanism.
- **Deviations from spec:** none in production code, docs, or tests — all 15
  literal diffs from *Notes for the Implementer* applied via `git apply`
  with a clean `--check` both before and after, byte-identical by
  construction. Two deviations in the build session's own mutation-probe
  tooling (not the spec's artifacts): (1) my first M-D2 probe edit was
  malformed (replaced only half of a line-wrapped phrase, producing a
  duplicated "where it falls" and a hash that did not match
  `927a9f5d3c3b`) — corrected to span both lines, after which the hash
  matched exactly. (2) My first whole-repo `func Test` grep for the 856
  count double-counted by including `.claude/worktrees/*` (full nested
  repo copies); the spec's own `internal cmd`-scoped `inventory.sh` command
  gives the correct, matching count. Neither is a spec defect.
- **Follow-up work identified:** none beyond what the spec already routes
  (DEC-054 T4 to verify; SPEC-092/093's stale "next free number" lines to
  SPEC-093).

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing in the spec itself. The one friction was mechanical: the
   mutation matrix's *Diff* column gives exact edit text for every probe
   except M-D1 ("one of three `omitted_failure_count` mentions" — it doesn't
   say which), so my probe's hash legitimately can't reproduce
   `471eb6137945` (mine: `8d7c04ac6af3`). The *behavior* reproduces exactly
   (AE1 stays green either way), which is what M-D1 exists to show.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No. `one-spec-per-pr`, `no-sql-in-cli-layer`, `stdout-is-for-data-stderr-is-for-humans`,
   and `test-before-implementation` all applied cleanly with no friction.

3. **If you did this task again, what would you do differently?**
   — Scope the first test-function grep to `internal cmd` (as
   `scripts/inventory.sh` does) from the start, rather than the whole repo,
   which double- and triple-counts nested worktree copies under
   `.claude/worktrees/`.

---

## Verification

*Filled in at the end of the **verify** cycle. The orchestrator had already
confirmed the 15 literals' byte-for-byte rebuild from a clean `main` and all
five gates. This record covers what a passing build could not see: DEC-054's
T4, the real output a user reads, mutants nobody wrote, and whether anything
outside `story` moved.*

- **Branch:** `verify/spec-094-story-honesty`, off `main` at `c3f707a`
  (PR #228, the build, merged as a squash).
- **Verdict:** ⚠ **PUNCH LIST: one record correction, fixed in this cycle.
  No functional defect, and no line of Go changed.** Every acceptance
  criterion reproduces on a fresh frozen copy. Nine novel mutants were all
  killed. The M-D1 row now pins one edit and reproduces its hash (V-F1).
  **T4 fired in part** (V-F2). That is a finding about DEC-050's invariant
  at the prose level, not a defect in this build, and it is routed to ship.

Every live-corpus number below came from one frozen copy
(`sqlite3 .backup`, SHA-256 `2657544f9a74…`, the same file design measured:
621 entries, max id 643, four `failed`), which was unchanged at the end. Two
binaries ran against it. `brag-pre` was built from a `git archive` of
`c167764` (the design merge, before `build(SPEC-094)`), and its strings hold
`omitted_failure_count` 0 times. `brag-new` was built from `c3f707a`, and its
strings hold the key once.

### The attack list

| # | Attack | Result |
|---|---|---|
| 1 | DEC-054 T4: does `✗ … (failed)` survive `me` and `manager` into a model's prose? | **Partly. V-F2.** On `manager` it cuts win-crediting from 22/30 to 4/30. On `me` it changes nothing (1/30 → 2/30). All six residual cases are #473 |
| 2 | Read the real output: four audiences, failures and none, `n == 1`, misspelled user `candor` | **HOLDS.** Note and clause placed and worded as LD3/LD5; `--print-directive` byte-identical |
| 3 | Mutants nobody wrote | **HOLDS.** 9 of 9 killed; 2 predictions named more tests than fired |
| 4 | M-D1's stated hash | **V-F1: under-specified. FIXED.** It was the second of three readings. Codification ruling below: **does not clear** |
| 5 | Did any other surface move? | **HOLDS.** 22 invocations over 9 commands byte-identical, each side checked to be the artifact |
| 6 | Who reads `omitted_failure_count` or the story JSON? | **HOLDS.** No consumer outside `internal/story`'s own tests and the docs this spec edited |
| 7 | Is the build reflection honest? | **HOLDS.** Both tooling errors reproduce as described. Neither warrants promotion |

### V-F2: T4, measured with SPEC-094's method

**Method.** Bundles came from the built binaries, not from edited text:
`story --audience <me|manager> --since 2026-09-01 --project bragfile`
(9 beats, 3 failures: 433, 465, 473). Each was built on `brag-new` (`✗`) and
on `brag-pre` (`★`, the counterfactual). The control was the same two
audiences over `--project bragfile-site` (48 beats, no failure). Every
bundle was shape-checked before any run:

| Bundle | Bytes | `✗` lines | `★` failure lines | SHA-256 |
|---|---:|---:|---:|---|
| `new-me` | 7 991 | 3 | 0 | `f2434f67f493` |
| `new-manager` | 8 065 | 3 | 0 | `f6315be65d35` |
| `pre-me` | 7 964 | 0 | 3 | `1d072f16eb34` |
| `pre-manager` | 8 038 | 0 | 3 | `b79a95140438` |
| `ctl-me` / `ctl-manager` | 10 361 / 10 435 | 0 | 0 | `919b007bdd6f` / `3e208c9c3d56` |

(The hashes cover a `Generated:` line, so they identify these files and no
re-run will reproduce them.) Each run was
`claude -p --model <haiku|sonnet> --tools "" --no-session-persistence --setting-sources ""`
with the tutorial's prompt, *"weave these threads into one headline arc"*,
on `claude` 2.1.280. There were five runs per model per bundle, and three
per model per control bundle: 52 runs. **Two `new-me`/Sonnet runs were
lost to tool-call attempts** (11 and 15 words, no prose), the same failure
framing recorded. They were set aside unread and replaced by two fresh runs.

**Grader, calibrated, then overridden by reading.** Two Sonnet graders
scored each failure id per output. Grader 1 asked *absent, failure, win or
neutral*. Grader 2 asked *is this failure credited anywhere as a win, for
example listed as shipped?* On the 12 control outputs (36 judgments, none
of them possible) grader 1 made **2 false hits** and grader 2 made **1**.
Each one credited unrelated `bragfile-site` work to #473. Grader 1 then
missed the one shape that matters: in `new-manager`/Sonnet run 2 it scored
#473 *failure* while the prose lists it under **"What shipped"**. Grader 2
under-counted the counterfactual (10 against my 13 on `pre-manager`/Sonnet).
**So every non-absent judgment in the 40 test outputs was read by hand, in
context, and the table is my reading, not either grader's.** *Credited as a
win* means the prose lists the failure among work shipped, fixed or landed,
or presents it chiefly as an accomplishment. It still counts when the
going-wrong is also stated as a "before".

| Cell | Model | Runs | Framed only as a failure | Credited as a win | Absent |
|---|---|---:|---:|---:|---:|
| `manager`, `✗` (this build) | Haiku | 5 | 15 | **0** | 0 |
| `manager`, `✗` | Sonnet | 5 | 11 | **4** (all #473) | 0 |
| `manager`, `★` (pre-build) | Haiku | 5 | 5 | **9** | 1 |
| `manager`, `★` | Sonnet | 5 | 2 | **13** | 0 |
| `me`, `✗` (this build) | Haiku | 5 | 12 | **2** (both #473) | 1 |
| `me`, `✗` | Sonnet | 5 (+2 lost) | 15 | **0** | 0 |
| `me`, `★` (pre-build) | Haiku | 5 | 13 | **0** | 2 |
| `me`, `★` | Sonnet | 5 | 14 | **1** (#473) | 0 |

A corroborating count needs no judgment. An id followed within 12
characters by `fail`/`failed`/`failure` (`/usr/bin/grep -oiE`) appears
**11 times** in the 20 `✗` outputs, and **0 times** in the 20 `★` outputs
and the 12 controls. Models do carry the label: `new-manager`/Sonnet run 3
writes *"**#433 (failed):** routing fixes by leaving a note in prose doesn't
work"*, and `new-me`/Sonnet run 3 writes *"**#433 failed**, and it's the
pivot"*.

**What it means.**

- **On `manager`, the label works, and it is the label that works.** With
  `★` and the directive's *"Lead with what shipped"*, the failures land
  under **Shipped** in 22 of 30 mentions. With `✗` they land under
  **Blockers**, **Friction**, **Risk** or a "what failed" heading in 26 of
  30.
- **On `me`, the label is inert.** The directive (*"the messy middle"*) and
  the entries' own titles (*"did not fire"*, *"why that fails"*) already
  carry the candour. With `★` too, 27 of 30 mentions are framed as failures.
- **The residue is one entry, and the entry is the cause.** All six
  residual cases are #473. Its recorded impact narrates its own fix (*"The
  guard now derives its keys from the renderer's own output"*). Sonnet on
  `manager` lists it as shipped in 4 of 5 runs: twice outright (*"**Test
  coverage gap closed** (#473)"*, run 3, in the same answer that labels 433
  and 465 *"(failed)"*) and twice alongside a blockers mention. **DEC-050's
  "a failure is never rendered as a win" holds in the bundle and does not
  hold in the prose** for a failure whose impact reads like a win. The
  binary cannot guarantee it downstream, and neither `me.md` nor
  `manager.md` says what `✗` means.
- **An inverse error, outside DEC-050.** Two `✗` runs explicitly call #472,
  a win, a failure: `me`/Sonnet run 3 (*"#472 and #473 are the same failure
  shape"*) and `me`/Haiku run 5 (*"The other failures (472, 473, 643)"*).
  The `★` runs were not audited for this.

**T4's trigger fired in part:** the label does not fully survive `manager`.
**Routed to ship** to decide whether to amend DEC-054 T4 with this result,
and whether to add one line to `me.md` and `manager.md` saying what `✗`
means. That change would edit assets that LD14 keeps out of this spec, and
the wording would need this measurement re-run. **No id is reserved here**
(the prose-reservation failure, #465).

**Limits.** Two models and one scoped window. There are three failures, all
with an impact, so an impact-less failure under `✗` is untested. The
prompt is the tutorial's `exec`-shaped *"one headline arc"*. A
manager-shaped prompt might behave differently, and that was not measured.

### Real output, read as a user would

All runs were on the frozen copy, with a throwaway `HOME` for user profiles.

- **`me` and `manager`, `--since 2026-09-01`.** The `diff` against `brag-pre`
  is exactly 4 changed lines each (8 `diff` lines), and each is
  `- ★ <id>:` → `- ✗ <id> (failed):`. The impact line under it is
  unchanged (AC-1). JSON is identical after
  `del(.omitted_failure_count, .generated_at)` in `--since`, `--year` and
  `--month` (AC-1, AC-4).
- **`exec` and `skip`, same window.** `Beats: 226/231` and `227/231`, then
  `Omitted: 4 recorded failures, not listed for this audience (brag list --type failed)`
  on the line directly below. The document's last line is the clause, a
  paragraph of its own after the asset's bullet list. On `exec`, the
  `bragfile` thread drops from 9 to 6 impact beats and moves below
  `irradiance`, as design reported. That is visible, not wrong. JSON:
  `omitted_failure_count` is the fifth key, `4` on both and `0` on both
  candid audiences (AC-5). In all six promotional cells, the arc counts
  equal `main`'s recomputed over non-failure beats, with 0 failure beats
  left (AC-4).
- **The empty state** (`exec --type failed`): `Threads: 0`, `Beats: 0/4`,
  the `Omitted:` line, and the directive ending in the clause. JSON gives
  `4` and `[]` (AC-3). A user who asked for failures is told why they got
  none, and where to find them.
- **`n == 1`** (`exec --project contextcore-pilot-harness`, #420 only):
  `Omitted: 1 recorded failure, …` and `This bundle omits 1 recorded failure …`.
  The singular holds in both places. `me` over the same window labels
  `✗ 420 (failed)`.
- **A window with no failure** (`--since 2026-09-09`): `exec` and `skip`
  have no `Omitted:` line and no clause, and `omitted_failure_count` is `0`
  and present. The markdown is byte-identical to `brag-pre` apart from
  `Generated:`.
- **User profiles.** Eight exec-shaped profiles differed only in `candor`
  or `directive`:

  | `candor:` as written | Result |
  |---|---|
  | `promotionl` (misspelled) | labels 4 `✗`, no note, no clause |
  | `Promotional` | labels |
  | `"promotional"` (YAML-quoted) | **labels**: the parser keeps the quotes |
  | `promotional # exec-like` (trailing comment) | **labels**: the parser keeps the comment |
  | `promotional   ` (trailing spaces) | omits: the parser trims |

  Each unrecognised value falls to the side that drops nothing (DEC-054
  part 1), and the `✗` lines make that visible. The quoted and commented
  forms are valid YAML that a user could reasonably write. **Recorded, not
  routed:** it is the pre-existing hand parser's behaviour for every key,
  and it fails safe.
- **The clause after a user's own directive.** For a file ending
  `No trailing list.\n\n\n`, the directive is right-trimmed, then `\n\n` and
  the clause, in both formats. For an **empty** directive file, and for a
  profile with **no** `directive:` key, `## Framing directive` renders with
  the clause as its whole body. **`--print-directive`** is byte-identical
  to `brag-pre` for all four bundled audiences, and it prints a user file
  as authored, with no clause.

### Novel mutants: 9, each predicted before it ran

The helper backs up to `/tmp`, applies one exact replacement (it refuses
unless the old text occurs exactly once), **refuses to run the gates until
the content hash has moved**, runs `go test -count=1 ./...` and
`./scripts/test-docs.sh`, restores with `cp`, and confirms the pre-hash
returned. Predictions were written to a file before the first run.
`test-docs` stayed green for every Go mutant, as predicted.

| id | File | Edit (old → new) | Predicted | Fired | pre → post |
|---|---|---|---|---|---|
| **V-N1** | `profile.go` | `return p.Candor == CandorPromotional` → `return strings.TrimSpace(p.Candor) == CandorPromotional` | only `…OnlyExactPromotionalOmits` | **exactly that** | `2e4b6f59c520` → `c30c3e26555d` |
| **V-N2** | `bundle.go` | `if opts.OmittedFailures == 0 {` → `if opts.OmittedFailures < 0 {` (the clause says "omits 0" on every bundle) | many, including the existing goldens | **8**: the `me`/`exec` markdown goldens, the `me` JSON golden, `EmptyDirectiveOmitsSection`, `EmptyWindow`, and 3 new | `c0decde8de07` → `33081fbd15ca` |
| **V-N3** | `bundle.go` | `` `json:"omitted_failure_count"` `` → `` `json:"omitted_failure_count,omitempty"` `` | `MeProfile_ShapeGolden`, `EmptyWindow`, the `me` half of `…PromotionalCountsWhatItOmitted` | **exactly those 3** | → `046f5053e450` |
| **V-N4** | `bundle.go` | `return d + "\n\n" + clause + "\n"` → `… + clause + "\n\n" + clause + "\n"` (clause twice) | 4, and the e2e survives | **2**: `…PromotionalOmitsWithNoteAndClauseGolden`, `TestLoadProfile_OnlyExact…`. The e2e survived, as predicted. **Over-predicted**: `…EveryBeatOmitted…` takes the empty-directive branch, and `…PromotionalCounts…` does not pin the directive's end | → `0e061e2b13f6` |
| **V-N5** | `bundle.go` | in the failure case, `if b.IsImpactBeat {` → `if true {` (a blank impact line under an impact-less failure) | only `…CandidLabelsFailuresGolden` | **exactly that** | → `95e9f4d39bc5` |
| **V-N6** | `bundle.go` | in `pluralFailures`, `if n == 1 {` → `if n == -1 {` (never singular) | `…ClauseAloneWhenDirectiveEmpty` and the e2e | **exactly those 2** | → `4ca63d30d632` |
| **V-N7** | `bundle.go` | `FramingDirective:    framingDirective(opts),` → `FramingDirective:    opts.Directive,` (clause dropped from JSON only) | `ClauseAlone`, `PromotionalCounts`, `EveryBeatOmitted` | **2**: the first two. **Over-predicted**: `EveryBeatOmitted` does not read the JSON directive | → `a219a0ba131f` |
| **V-N8** | `bundle.go` | `if directive := framingDirective(opts); directive != "" {` → `if directive := opts.Directive; directive != "" {` (clause dropped from markdown only) | 5, including the e2e | **exactly those 5** | → `bfc0ee4e7fcb` |
| **V-N9** | `cli/story.go` | `EntriesInWindow: len(entries),` → `EntriesInWindow: len(shown),` | only the `learn` e2e | **exactly that** | `90683ed9106a` → `847e8707877a` |

**What the nine establish.** Each half of the note-and-clause pair is
guarded on its own format (V-N7, V-N8), so LD6's "one field drives all
three" is enforced, not only designed. DEC-014 part 4's always-present key
is guarded against `omitempty` (V-N3). The singular (V-N6), the
denominator (V-N9) and the impact-less failure (V-N5) each have exactly
one guard. **One gap, recorded and not a defect:** a doubled clause
survives the e2e and the JSON test, and only the markdown golden and the
user-file test catch it.

The **first attempt at V-N6 was refused by the helper**, because
`if n == 1 {` occurs twice in `bundle.go`. It was re-anchored on
`func pluralFailures(n int) string {` and run as above. A second helper
fault showed up at item 7: it decoded escapes with `unicode_escape`, which
mangles non-ASCII, so it refused M-D2 (`✗`) as occurring 0 times. The
occurrence guard caught it before any gate ran. None of V-N1 to V-N9
contains non-ASCII, so none of them was affected.

### V-F1: M-D1 pinned. FIXED, and the codification ruling

`docs/api-contract.md` at `c3f707a` (`1cd16e2a3d7a`, the same bytes as
design's prototype) holds `` `omitted_failure_count` `` three times. Each
reading of *"one of three"* was hashed and run through the probe:

| Reading | Line | Hash | `AE1` |
|---|---|---|---|
| 1st | `:908` | `8d7c04ac6af3` (build's) | green |
| **2nd** | **`:938`** | **`471eb6137945` (design's)** | green |
| 3rd | `:954` | `d7c835adfd9f` | green |
| all three (M-D1′) | — | `f313c537a4cb` | **fires** |

Design ran the second reading. The matrix row now states that edit
verbatim, and it reproduces `471eb6137945` from a pristine file. Its
behaviour was never in doubt: every reading leaves `AE1` green, which is
Finding 4.

**Ruling: this does not clear STAGE-023 held candidate 1.** The candidate
clears when *"a second stated edit meets the parent clause and does not
reproduce its hash."* M-D1 does not meet the parent clause. A diff that
admits three literal readings is not pinned, and SPEC-086 ship already drew
that line. Its positives (SPEC-088's `M-B3`/`M-B4` and SPEC-086's five
abbreviated rows) were admitted because each prose description had
**exactly one** literal reading. So M-D1 is the parent clause's own
negative, of the same kind as SPEC-087's `M-6`, which ship declined to
count. `M-A0`, the candidate's one case, has a different defect: a single,
unambiguous stated edit that is not what ran. M-D1's record is incomplete
rather than wrong. What ran is one of its readings, and it reproduces the
moment the reading is named. **The best case for counting it:** the matrix
preamble says *"the Diff column is the replacement the helper applied,
verbatim"*, and for this row it was not, which is composition after the
run in the literal sense. That argument is recorded here for ship, but it
is not the stated condition. Candidate 1 stays at **N=1**. Ship codifies.

### Other surfaces: nothing moved

22 invocations, each run on `brag-pre` and `brag-new` against the frozen
copy, were compared after deleting only `Generated:` / `generated_at`
lines (1 line per document): `impact` (`--year`, `--since`, JSON,
`--type failed`), `wrapped` (md, JSON), `summary --range month` (md, JSON)
and `--range week --type failed`, `export` (markdown, JSON), `coverage`
(md, `--year` JSON), `list` (`--since`, `--type failed`, JSON), `stats`
(md, JSON), `memory` (`--project bragfile`, JSON), and `review` (`--week`,
`--month` JSON). **All 22 are identical.** Before any equality was believed,
each side was checked: exit 0 on both, empty stderr on both, a non-empty
body (87 B to 516 KB), and the expected first line (`# Bragfile Impact`,
`# Bragfile Wrapped`, …, `{`, `[`, or a tab-separated row). **Positive
control:** the same harness on `story --audience exec` reports **DIFF**. So
the comparison can see a difference between these two binaries, and it
compared real documents, not two equal error messages. Flags were taken
from each command's own `--help` first, so no cell ran a flag that does not
exist.

### Consumers of the new key

`git grep omitted_failure_count` outside this spec finds the renderer, its
tests (`bundle_test.go`, `candor_test.go`, `learn_test.go`),
`docs/api-contract.md` (3), `CHANGELOG.md`, DEC-029, DEC-050, DEC-054,
STAGE-023 and `scripts/test-docs.sh` (`AE1`, `AE4`). Every one was written
by this spec's design or build. A wider sweep followed, for
`framing_directive`, `is_impact_beat`, `impact_beat_count`, `"throughline"`
and `story … --format json` outside `internal/story`, `projects/` and
`decisions/`. It finds only `docs/api-contract.md`, `docs/tutorial.md:629`,
`internal/cli/story.go:56` (usage lines) and `CHANGELOG.md`. **`BRAG.md`
(`:442`) and `README.md` (`:229`) name `brag story` in markdown usage
only.** In `internal/mcpserver` the only `story` hits are inside
"hi*story*". There is no story tool (LD14). **The plugin and `examples/`
have no hit. No unupdated reader.**

### The build reflection: HONEST

- **M-D2's line-wrapped edit.** The phrase does span `docs/tutorial.md:649`
  and `:650`. The two-line edit reproduces `927a9f5d3c3b` and fires `AE2`.
  A one-line edit leaves a malformed file with a different hash
  (`56b75290b01f`, on this reading. Build did not record its own malformed
  hash, so its exact text cannot be re-derived). What caught it was
  comparing against the **stated** hash. "Hash moved" alone would have
  passed it.
- **The worktree double-count.** `.claude/worktrees/` holds two nested
  checkouts. `^func Test` over `internal cmd` gives **856**. Over the whole
  repo it gives **2500**.
- **Promotion: neither.** The first is the parent §12 clause working, the
  stated hash catching a malformed probe. It adds no new rule. The second
  is already covered by *"numbers in docs are derived, not typed"*: count
  with the script that derives the row. My own helper fault (V-N6's refusal
  and the `unicode_escape` refusal) is the same shape as the first. An
  exactly-once occurrence guard caught what a hash-moved guard would not
  have. It is recorded, not proposed, and it stays at N=1.

### Gates: all five green on this branch

| gate | result |
|---|---|
| `just test` | exit 0 · **1113** passing (incl. subtests) · **856** top-level `func Test*` · **0** failing · **14** packages `ok` |
| `just test-docs` | exit 0 · **210** `OK:` / **209** distinct ids / **4** `AE` ids / **0** `FAIL:` |
| `just lint` | `0 issues.` |
| `gofmt -l .` | empty |
| `go vet ./...` | exit 0 |

### What this cycle changed

- This spec only: the M-D1 matrix row (V-F1) and this section. **No Go, no
  test, no doc and no `test-docs.sh` change**, so no derived number moves
  and the inventory block is untouched. `cycle:` stays `verify`. Ship
  advances it.

### Not findings, checked and clean

- **The corpus was never written to.** The only invocations against
  `~/.bragfile/db.sqlite` were the read-only `brag memory --project
  bragfile` that §13.5 asks for and the `sqlite3 .backup`. Every other run
  used the frozen copy, whose SHA-256 was unchanged at the end. **No brag
  was captured; the brag comes at ship.**
- **Every probe target is back at its baseline**, and `git status` was
  clean between probes.
- **Counts** were taken with `/usr/bin/grep`, `git grep` or Python. JSON
  went through files, never `echo`.
- **Nothing out of scope was touched:** SPEC-095, SPEC-092/093's stale
  DEC-054 lines, `export/markdown.go:55`, `memory/memory.go:240`, `brag
  review`, AGENTS.md, and `guidance/questions.yaml`.

### Ship addendum (2026-09-23): the matrix's bases are named, and `M-D3` was re-run

**The mutation matrix's nine baselines are now tied to commits on `main`.**
Design measured them *"on the prototype's final files"* and named no commit.
At ship, each of the nine files hashes to its stated baseline at both
`c3f707a` (the build merge) and `cf57ea2` (the verify merge). Those are the
durable bases for every row. A later edit to a target does not invalidate a
row; it means the row is reproduced from `git show <base>:<path>`.

**This ship moves one target itself.** Codifying the §12 refinement edits
`AGENTS.md`, `M-D3`'s target. Rather than re-pin the row against a branch
commit that the squash merge will erase (the way SPEC-086's `"this ship
commit"` row went stale, see STAGE-023 held candidate 2), the row keeps its
base:

- **At `c3f707a`:** `M-D3`'s stated edit occurs exactly once, and reproduces
  `713a717a154f` → `bcab08c6ce19`.
- **On this ship branch, as a behaviour check only:** `00ae4abc713f` →
  `3cb78eea9732`. It fires **`AE3` alone** (208 `OK:`, 1 assertion failed),
  and the pre-hash returned after a `cp` restore. The helper refused to run
  unless the old text occurred exactly once. This hash is not a pin, because
  the branch commit does not survive the squash.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — Measure the candid label at design, in the same run as the clause.
   Design piped 25 promotional bundles through two models and routed the
   candid half to verify as T4. Verify's 52 runs then found the one residue
   the binary cannot fix, #473, whose impact narrates its own fix, and the
   one lever that might, a line in `me.md` and `manager.md`, which LD14 had
   already put out of scope. Measured at design, that line would have been
   in this spec or NO-GO'd here. Instead it is SPEC-096, with a second 52-run
   measurement of its own. Second: pin every matrix row from the helper's own
   replacement, not a summary written afterwards. M-D1's *"one of three"*
   cost build a hash it could not reproduce and cost verify three probes.

2. **Does any template, constraint, or decision need updating?**
   — Two, both done in this cycle. **DEC-054** gains an
   `## Amendment (2026-09-23, SPEC-094 ship)` recording T4's measurement: the
   method, 22/30 → 4/30 on `manager`, and the conclusion that DEC-050's
   invariant holds in the bundle and not always in the prose. **AGENTS.md
   §12**'s *pin its diff* clause gains a refinement, *a pinned diff has
   exactly one literal reading*, cleared at N=2 paired-opposing (M-D1 against
   SPEC-086's and SPEC-088's one-reading rows). No template changes: no
   template carries a mutation matrix, so the rule lives in §12 and in the
   probe helper that enforces it. No constraint changes.

3. **Is there a follow-up spec I should write now before I forget?**
   — Yes, and it is written: **SPEC-096**, *the candid directives say what a
   failure label means*, claimed by its file in the same edit as its
   STAGE-023 line. It carries verify's measurement, the LD14 conflict, and
   the requirement to re-run the 52 runs. It does not gate v0.7.0.
   **SPEC-095** already exists; it is the last spec gating v0.7.0 and the
   next to frame.

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
   — Before, every `brag story` audience listed each recorded failure as a
   `★` win (4 of 4 on all four profiles), and a model writing a `manager`
   update from that bundle credited the failures as shipped in **22 of 30**
   mentions; now `me` and `manager` label them `✗ <id> (failed)`, which cuts
   that to **4 of 30** (every survivor #473, whose impact describes its own
   fix), and `exec` and `skip` leave them out but say so, which a model
   carried into its prose in **25 of 25** runs where a bare note survived
   0 of 10. The *before* is taken from framing's re-measurement and verify's
   `★` counterfactual, not from `## Context`, which predates both. **Not
   captured by this session:** the brag is drafted for the maintainer's
   approval (SPEC-086's precedent, and this cycle's instruction).

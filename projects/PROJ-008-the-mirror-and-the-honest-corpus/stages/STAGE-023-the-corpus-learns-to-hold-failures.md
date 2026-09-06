---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-023                     # stable, zero-padded, repo-global (never reused)
  status: active                    # proposed | active | shipped | cancelled | on_hold
  priority: high                    # critical | high | medium | low
  target_complete: null

project:
  id: PROJ-008                      # parent project
repo:
  id: bragfile

created_at: 2026-09-05
shipped_at: null
---

# STAGE-023: the corpus learns to hold failures

> **Cycle: frame.** GO. This stage is the project's **first**, ahead of impact
> quality — reversing the brief's own stage plan. The ordering call and its
> rejected option are in *The ordering call* below. Every number on this page
> was re-derived on 2026-09-05 against `main` at `1775feb`, corpus **397**.

## What This Stage Is

The corpus can hold work that **did not work**, and an agent reading the
history gets it back. Today `brag` records 397 entries and **none of them is a
failure** — not because failures were suppressed, but because there is no way
to say one. This stage adds the capture verb, decides where a failure lives in
the schema, and makes it survive the read path that agents actually use
(`brag memory`). It stops when the corpus is *capable of being honest*; making
the digests reason about honesty is a later stage.

## Why Now

**The flattery problem is upstream of everything else in PROJ-008.** The brief
argues impact quality goes first — *one classifier, three surfaces* — and that
argument is sound about the mirror and story-surface v2, which both consume a
ranked impact. It is not sound about ordering against `brag learn`, for a
reason the brief could not see when it was written:

**The classifier's calibration data is the corpus itself.** An impact rule
designed against 397 wins is tuned on a population that excludes the case it
will most need to judge. A failure's impact statement reads differently from a
win's — *"cost two days and produced nothing reusable"* is quantified, negative,
and has no before/after — and a rule that never saw one will be re-opened the
first time it does. Building capture first means the classifier is designed
against a corpus that contains both shapes.

Two further weights, both checked rather than inherited:

1. **The promotion condition already fired.** The PROJ-005 synthesis parked
   `brag learn` with *"promote to (A) if the memory work lands."* It landed in
   v0.6.0.
2. **It is the only pillar that changes what the corpus can hold** rather than
   how it is read. Every surface built on a wins-only corpus flatters,
   including the classifier.

And it is the user's stated priority (2026-08-23, at activation).

### The ordering call

**Chosen: `brag learn` first (this stage), impact quality second
(STAGE-024).**

**Rejected: impact quality first, `brag learn` second** — the brief's own stage
plan. Rejected on two grounds, one of principle and one measured:

- *Principle.* The calibration argument above. Ordering the classifier first
  makes it a rule written against half its input domain.
- *Measured.* The brief's premise that impact quality is a **small** primitive
  — *"the highest-leverage primitive is small and blocks all three"* — **does
  not survive measurement.** See *The classifier is not small* below. It is not
  an S, so putting it first delays the whole project rather than unblocking it.

**What the rejection costs, named so it is a choice:** the mirror (stage 3) and
story-surface v2 (stage 4) both consume ranked impact, so impact quality must
land before *them*. This ordering does not change that — it inserts one stage
ahead of the classifier, not ahead of its consumers. The cost is one stage of
delay on `wrapped`'s highlight-reel problem, which has been true since v0.4.0
and is not degrading.

### The classifier is not small — measured 2026-09-05

The brief measured impact quality on 2026-08-15 across 368 entries and reported
**52 (14%)** carrying a number with a unit. Re-derived today across 397, the
answer depends almost entirely on how the rule is worded:

| Candidate rule | Entries | % of 397 |
|---|---:|---:|
| impact contains any digit | 230 | 58% |
| digit + `%` / `x` / `$` | 41 | 10% |
| digit + a unit noun (`ms`, `files`, `tests`, …) | 85 | 21% |
| an `N → N` before/after shape | 11 | 3% |

A twentyfold spread between the loosest and tightest reading, all of them
defensible in prose. Two specific failure modes make this worse than a
threshold-picking exercise:

- **False positives: 141 of the 230 digit-bearing impacts contain a digit that
  is an identifier, not a measurement** — `SPEC-083`, `v0.6.1`, `u16`, `MD5`,
  `#190`. A digit rule is wrong on roughly three fifths of what it selects.
- **False negatives: 27 impacts are quantified only in spelled-out numerals** —
  *"Five subject workspaces"*, *"Three distinct causes"*, *"the two seams"* —
  and are invisible to every digit rule. This is not incidental: **this repo's
  house style spells numerals out in prose**, which is why `root.go:13` carries
  an unguarded `"four"` as a routed comment item.

So "quantified vs. prose" is not a filter someone forgot to write. It is a
stated, tested rule with a documented false-positive posture — an **M**, and
plausibly an L if it must also be applied retroactively. That is a finding for
STAGE-024's framing, recorded here because it is what settled the ordering.

## Success Criteria

- A user can record work that did not work, in one command, without
  contorting it into a win.
- A failure entry is **retrievable as a failure** — some surface can answer
  *"what has not worked?"* rather than requiring the user to remember.
- `brag memory` returns failures to an agent reading the corpus cold, and the
  slice's line shape says which entries they are.
- The celebratory digests (`wrapped`, `impact`) do **not** silently absorb
  failures. Whatever they do is a decision written down, not an accident of
  whichever query happened to match.
- The DEC-014 envelope and DEC-048's provenance-count rule hold on every
  surface this stage touches.
- No LLM in the binary. No network. No migration unless a fork explicitly buys
  one.

## Scope

### In scope
- The capture verb — `brag learn` (or whatever SPEC-085 settles it to be), and
  its MCP counterpart if the fork lands that way.
- **Where a failure lives in the schema** — the `type` / reserved-tag / new-
  column fork, settled with evidence.
- **How failures participate in `brag memory`** — the DEC-043 ranker and the
  DEC-044 line shape.
- **What the digests do with failures** — decided and written down, even where
  the decision is *"nothing, and here is why."*
- The `guidance/questions.yaml` entry `memory-slice-fusion-constants`: check
  whether this stage answers it, and say so either way.

### Explicitly out of scope
- **The impact-quality classifier.** STAGE-024. This stage must not smuggle in
  a ranking rule.
- **The mirror** and **story-surface v2**. Stages 3 and 4.
- Retroactively reclassifying the 8 existing `type: learned` entries. Their
  status is a fork in SPEC-085; *acting* on it beyond a documented position is
  out of scope here.
- Capture completeness (staging inbox, git-import miner) — routed, see
  *Dependencies*.
- The read-path discoverability gap (`BRAG.md` is write-only) — routed, see
  *Dependencies*.

## Spec Backlog

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [x] SPEC-085 (shipped on 2026-09-05) — the capture verb, its schema home,
      and its behaviour in `brag memory`; five forks, all settled. Complexity
      **M**, re-affirmed after the split below. Shipped as `brag learn` +
      **DEC-049**; `pr:199`.
- [ ] SPEC-086 (frame) — digest posture: the `## What didn't work` section on
      `wrapped` and `impact`. **Re-framed 2026-09-06: UNBLOCKED** (SPEC-085
      shipped at `df369e9`), **GO at M after splitting**, and it authors
      **DEC-050**, which states the posture for **all seven** `--type` surfaces
      so the scope guard below is satisfied by the decision even though the
      renderer work lands in two PRs. Framing also shrank the DEC-048
      obligation from *"two count renames"* to *at most one* — see
      *Re-framing corrections* below.
- [ ] SPEC-087 (design) — **`Y3` derives instead of caching.** Split out of
      SPEC-086's re-framing rather than absorbed. **Designed 2026-09-06, all
      three forks settled, complexity S held.** Still **sequence it BEFORE
      SPEC-086 design**, which creates DEC-050 and would otherwise be the
      sixth consecutive hand re-pin. See *Carried into SPEC-086*, item 2.
      Design's headline correction to this page's item 2: the second
      derivation did not have to be invented — **`Z7` already computed it**
      (SPEC-082 LD10) and was measured **blind**, because it re-implements
      `inventory.sh`'s two greps verbatim instead of reading what the script
      emits. Fork 1 is a repair, not an invention. Fork 2 rejected (Fork 1
      gets its benefit without touching a user-facing doc); Fork 3 retired as
      moot. `Y4`'s pin stays and is routed to **SPEC-088** with an id, not to
      an anonymous next spec.
- [ ] (not yet written) — **`summary` + `story` markdown honesty.** The other
      half of SPEC-086's Fork B, split on defect shape: on these two the data is
      already present (`story --format json` carries `"type": "failed"`;
      `summary`'s `## Summary → By type` prints `failed: 1`) and only the
      per-entry markdown drops it. **Unconditional** — DEC-050 will have stated
      their posture, so this implements a decision rather than making one. Its
      `story` half additionally has to turn `Candor` from LLM-facing metadata
      into a body rule (`internal/story/profile.go:24`), which is a
      DEC-029-adjacent decision of its own.
- [ ] (not yet written, `bug`) — **`--type` negation is inexpressible and fails
      silently.** `--type '!failed'`, `'-failed'`, `'shipped,failed'`,
      `'!=failed'` and `'NOT failed'` each return **exit 0 with zero rows, no
      diagnostic**; only `--type ''` errors. `internal/storage/store.go:389` is
      exact-match inclusion. Blast radius is all seven `--type` surfaces.
      Routed here, **not** to `guidance/questions.yaml`: the behaviour, the
      line and the fix shape are all measured, so it is a bug and not a
      question — and filing it as a question would move `Y4`'s pinned counts
      and force an inventory regeneration for zero information gain.

**Count:** 1 shipped / 0 verify / 1 designed / 1 framed / 2 not yet written
(SPEC-088 is reserved-by-routing for `Y4`, not yet written — see SPEC-087 LD6.)

**The stage does NOT close here.** Success Criteria 1, 2 and the DEC-014/
DEC-048 envelope line are met by SPEC-085; Criterion 4 (*the celebratory
digests do not silently absorb failures*) is owed by SPEC-086 **and** the
`summary`/`story` successor, and Criterion 3 is **half met** — see *The Fork 3
finding* below.

### The conditional spec fired (2026-09-05, at SPEC-085 design)

This backlog's second slot was written *"only if SPEC-085's Fork 4 turns out to
need more than a documented default."* It did. Put to the user at design with
the December reading spelled out, Fork 4 resolved to a **named
`## What didn't work` section** on both celebratory digests — rejecting both
exclude-by-default and include-silently.

That is a renderer change on two surfaces **plus two DEC-048 count renames**
(once *Impact moments* stops carrying every with-impact entry, its headline
count no longer means what it says). A second M of work on different files with
its own decision record, so it was split rather than absorbed — which is what
kept SPEC-085 at M.

The measurement that made "include silently" unacceptable, and that framing did
not have: **neither `impact` nor `wrapped` renders `type`, in either format.**
`ToImpactMarkdown` emits `- <id>: <title>` + the impact text, and
`impactEntry` is a deliberately narrow 4-key JSON projection (DEC-028 choice
4). A failure there is not merely unflagged — it is *unrepresentable as a
failure* without a renderer change.

## Design Notes

Four measured facts that constrain every spec in this stage. Each was
re-derived on 2026-09-05 against `1775feb`; none should be re-litigated from
memory.

1. **`type` is free-form, not a closed set.** `internal/cli/add.go:101`
   describes it as *"free-form category (shipped, learned, mentored, ...)"* and
   nothing validates it. The live corpus holds **19 distinct values** across
   397 entries, including near-duplicates (`shipped` 201 / `ship` 15,
   `fixed` 2 / `bugfix` 1) and **113 entries with no type at all**. A `type`
   value therefore carries **no guarantee** today — which is an argument both
   for using it (no migration, no new concept) and for the fork having to say
   what makes this value different.

2. **`learned` already exists and means something else.** Eight entries carry
   `type: learned`, and reading them, they are **lessons framed as wins** —
   *"Verify caught an unpinned test seam…"*, *"Re-running a pre-flight after a
   dependency bump paid for itself."* Not one records something that failed.
   So `learned` is **occupied**, and SPEC-085 cannot quietly adopt it.

3. **Tags are normalized — DEC-004 is superseded.** DEC-004 (comma-joined
   `TEXT`) was superseded by **DEC-015** on 2026-06-06; `0003_normalize_tags.sql`
   creates `tags` + `taggings` with a polymorphic membership join and two
   indexes. A reserved tag is therefore a **first-class indexed lookup**, not a
   `LIKE '%…%'` scan. Any framing input that costs a reserved tag using DEC-004's
   model is costing a schema that has not existed since June.

4. **The memory line renders `type` and does not render tags.** DEC-044's line
   shape, in `internal/memory/memory.go:224`, is:

   ```
   - <id> <YYYY-MM-DD> [<project>/<type>] <title> — <impact>
   ```

   This is the sharpest fact in the stage. A failure marked by **`type`
   travels to the agent-facing read surface for free**; a failure marked by a
   **reserved tag is invisible there** unless DEC-044's locked line shape is
   amended. The two options are not equivalent-cost, and the difference lands
   on the read path rather than the write path.

Two further constraints, inherited rather than discovered:

- **The envelope.** Six `internal/export/` renderers emit a headline count
  under DEC-014, and **DEC-048** (2026-08-23) legislates that the count must
  name what it counted, binding all six forward. Any new surface inherits both.
- **The ranker.** `brag memory` fuses three ordinal lists — recency, relevance,
  project — by RRF (DEC-043), with `PoolLimit = 200`. **There is no type-aware
  or tag-aware signal.** A failure entry competes on recency alone and is
  crowded out by volume; if failures should surface reliably, that needs an
  argued answer, not an assumption that they will.

### Corrections from SPEC-085 design (2026-09-05)

Four of this page's measured facts moved or were wrong when re-derived at
design against `main` at `81e639d`. Recorded here so a later stage does not
re-inherit them; none changes a conclusion this stage reached.

| This page says | Measured at design |
|---|---|
| corpus **397** | **398** (framing's own brag landed) |
| *"**19** distinct values"* | **18** distinct non-empty values. The 19th was the empty bucket — the same 113 entries this page separately reports as *"with no type at all"*, counted twice in two units. |
| *"eight entries carry `type: learned`, and reading them, they are **lessons framed as wins**… Not one records something that failed"* | The eight are **not uniform.** id 18 is a PROJ-001 smoke-test artifact, not a lesson. ids 80/361/383 take a **failure as their subject** and frame it as a win recovered. `learned` is closer to *post-mortems with a recovery* than to *wins*. **The conclusion holds and gets stronger** — redefining it would make `brag list --type learned` return a mix, failing Success Criterion 2. |
| Fork 1's table: `--type` on **5** of 8 commands | **7** — `list`, `export`, `summary`, `impact`, `wrapped`, `story`, `coverage`. The denominator omitted four commands. Retrieval is *more* free than claimed. |

**And one correction to a routing rationale on this page.** *Read-path
discoverability* is routed to STAGE-024 on the premise that `BRAG.md` *"has 22
headings and **none** is a read-the-corpus section."* Measured on the live tree:
**26** markdown headings (16 at `##`), and `## Reading entries back` (line 382)
is exactly such a section — 31 lines listing eight read commands. The literal
string `brag memory` appears **zero** times in the file, not twice.

The gap is real but differently shaped, and smaller: **the read section exists
and omits the one command written for agents.** STAGE-024 should frame it
against that, not against "BRAG.md is write-only."

### Correction from SPEC-085 verify (2026-09-05) — for SPEC-086 framing

**SPEC-086's Fork B splits the seven `--type` surfaces into *celebratory*
(`wrapped`, `impact`) and *neutral* (`summary`, `story`, `export`,
`coverage`), and asks design to decide per surface. Two of those four are
misclassified, measured on a corpus carrying `type: failed` rows.**

`export` and `coverage` are genuinely neutral — `export` renders `type` in its
per-entry table, `coverage` is provenance-only. The other two are not:

| Surface | What it actually does with a failure |
|---|---|
| `story --audience exec` | renders it as a `★` beat with **no `type` in markdown**, under a printed directive that says *"build the narrative from those outcomes"*, *"Terse and promotional… No process, no messy middle"*, *"Quantify wherever the impact beats give you a metric."* The output is a **prompt for an LLM**, so this is a strictly worse leak than `impact`'s: `impact` shows a human a mislabelled row, `story` instructs a model to launder it. |
| `summary` | its section is literally `## Highlights`; the failure appears there with no `type` in either format (`summary`'s JSON highlight is a 2-key `{id,title}`, narrower than `impact`'s 4-key). |

Two constraints this puts on Fork B:

1. **"Is `story` celebratory?" is not a property of the command.**
   `--audience me`'s directive is candid by design (*"Include the messy middle:
   struggles, false starts… are the point"*) while `--audience exec`'s is
   promotional. Fork B's per-surface framing cannot express a per-**profile**
   answer, and needs to.
2. **`story --format json` already carries `"type": "failed"` on each beat.**
   Only the markdown path is lossy — which makes `story` a cheaper fix than
   `impact`/`wrapped`, not a more expensive one.

Reproductions are in SPEC-085's `## Verify`. Nothing here changes SPEC-085,
and none of it was known at the time Fork B was written — it needed a corpus
with failures in it, which is what SPEC-085 shipped the ability to make.

**Also sharpened, not overturned:** SPEC-085's *interim risk* is bounded partly
because *"the user controls when the first [failure with an impact] is
written."* True, but `BRAG.md`'s new section tells the agent *"The `impact`
field is still worth filling in"* and shows `-i` in its example — so the
leaking shape is the **documented default path**. The risk is still acceptable;
what bounds it is the empty corpus, not user restraint.

### Carried into SPEC-086 (added at SPEC-085 ship, 2026-09-05)

Two items with a **named owner**, because the last time one of them was routed
without one it did not fire.

**1. `--type` negation is inexpressible, and it fails silently.** SPEC-086's
Fork A needs *"everything except failures"*; `internal/storage/store.go:388` is
exact-match inclusion on a single string. Re-confirmed at ship on a two-row
corpus: `--type '!failed'`, `--type '-failed'` and `--type 'shipped,failed'`
each return **exit 0 with zero rows** — no error, no diagnostic. Only
`--type ''` errors. A user filtering failures out of a review gets an empty
document rather than a message. This is a **dependency of Fork A**, not a
footnote.

**2. `Y3` and `X3` pin the same numbers by different means, and one of them
should derive rather than cache.** STAGE-022's close promoted this to a
stage-level lesson and then **held** it with the trigger *"routed to the next
spec that opens that file."* SPEC-085 **was** that spec — it inserted 80 lines
of Group `AB` into `scripts/test-docs.sh` — and hand-re-pinned `Y3` anyway,
making it **five consecutive specs** (SPEC-081, 082, 083, 084, 085). The
routing failed for the reason STAGE-022 recorded one line earlier about
`archive-spec`: *"Routing it as a candidate after the first hit did not prevent
the second; the guard did."* An anonymous "next spec" is not an owner. Named
here: **SPEC-086**, which will open the same file. Not codified as a rule —
that is N=2 same-outcome and this repo wants N=3.

Worth pairing with what *did* work, so the lesson is not read as "greps do not
help": AGENTS.md §9 half-(b) (grep for the **value**, not the idea) caught `Y3`
**at design** on SPEC-085, the first time in those five specs that it was found
early rather than late. The rule works; the duplication it keeps finding is the
thing to remove.

**Both items discharged at SPEC-086 re-framing, 2026-09-06:**

- **Item 1 (`--type` negation) — DESIGNED AROUND, not fixed and not blocking.**
  Re-confirmed independently (all five negation spellings: exit 0, zero rows,
  no diagnostic). But it is **not** a dependency of Fork A, which is what it was
  filed as. Mutation **M-1** — `aggregate.WithImpact` changed to
  `if e.Impact != "" && e.Type != "failed"`, confirmed by content hash —
  showed a **one-line in-memory predicate does the entire selection job**;
  `impact` and `wrapped` read all in-window rows once and partition in Go, the
  same shape `brag coverage` already uses because it needs both classes. No
  query in SPEC-086 wants `--type '!failed'`. The bug is real and is now a
  named `bug` entry in the backlog above.
- **Item 2 (`Y3`/`X3`) — SPLIT to SPEC-087, with a file.** Not absorbed:
  SPEC-086 was already splitting on Fork B, and taking this on would have
  pushed it back to L. Not re-routed anonymously either — the file exists,
  which matters mechanically as well as socially: `scripts/_lib.sh:107-119`
  computes `next_id` by scanning **filenames**, so an id reserved only in
  backlog prose gets handed to the next `just new-spec`. **SPEC-087 should
  land before SPEC-086 design**, which authors DEC-050 and would otherwise be
  the sixth consecutive hand re-pin. One correction to this page's framing of
  the item: `X3` and `Y3` are **not** redundant — `X3` catches a stale page,
  `Y3` catches `inventory.sh` itself going wrong, and `Y3`'s own comment names
  that failure mode (*"script and page would still agree, just agree on 48"*).
  The defect is that `Y3`'s **expected value** is a hand-maintained literal,
  not that the assertion exists.

### Re-framing corrections (SPEC-086, 2026-09-06, `main` at `df369e9`)

Six of this page's and SPEC-086's measured facts moved when re-derived. Recorded
so a later stage does not re-inherit them.

**The one that changes a conclusion: the interim risk is REALISED.** SPEC-085
accepted it as bounded because *"the corpus holds zero such entries."* Measured
2026-09-06: **one `type: failed` row exists** — id **420**, project
`contextcore-pilot-harness`, created `2026-09-06T00:44:44Z`, agent-authored,
**carrying a full impact statement**. It renders as an unmarked win on four
surfaces today. This is *not* the drafted `brag learn` entry the orchestrator is
deliberately holding; that one is still held. **The zero was a choice for
exactly one day**, and the argument for SPEC-086's priority is no longer *"this
will happen"* but *"this is happening, and it compounds as the corpus grows."*

| This page / SPEC-086 says | Measured 2026-09-06 |
|---|---|
| corpus **397** / **398**; `impact` `324/398`; `wrapped` `398` | corpus **420**; `Entries: 346/420 with impact`; `Entries: 420` |
| **0** `type: failed` rows | **1** (id 420) |
| *"**18** distinct non-empty type values"* | **19** — `failed` joined the set, as DEC-049 intended |
| `--type` inclusion at `store.go:**388**` | **`:389`**. `:388` is the `if f.Type != ""` guard. The negation precedent at `:404` is cited correctly. |
| Fork B: the `story` leak is an **`exec`** problem | **All four bundled profiles** render `- ★ 420:` with no `type`. **Two** are `candor: promotional` (`exec` **and `skip`**), two are `candid`. The renderer defect is profile-independent; only the *harm* is profile-dependent, because the promotional directives instruct a model to promote the unlabelled beat. |
| Fork B: `summary` renders a failure with no `type` | **Split, not total.** `## Summary → By type` **does** print `failed: 1` (and JSON `counts_by_type` carries it). Only `## Highlights` is lossy — and its JSON highlight is a **2-key** `{id,title}`. `wrapped` has no equivalent honest counterpart: its `Top types` is a **top-3** (`shipped 211`, `milestone 29`, `ship 15`), so `failed: 1` will never appear there. |
| Fork D: *"DEC-014 part 4 has a precedent — `wrapped` omits body sections on an empty period"* | **The precedent covers the wrong case.** DEC-014 part 4 governs the empty **document** (confirmed: `wrapped 2024` ends after `Entries: 0`). For an empty **section** in a non-empty document the two surfaces already **disagree**: `wrapped` renders a bare `## Impact moments` heading, `impact` omits its whole body. Fork D is a live fork, not an application of a settled rule. |
| Fork E: *"five files in `internal/export/` carry the affected strings"* | Grep-shaped. Measured by mutation: the section heading fires **4 tests in 2 files**; a headline rename fires **9 tests in 3 files across 2 packages**, one of them **`internal/cli`**. `memory_test.go:247` (the byte-length assertion §9 cites) **does not move**. |

**And one methodological correction worth keeping.** The claim that
`story --audience exec` renders a failure with *"zero occurrences of `failed`
anywhere in the markdown"* is **false as literally stated** — the string appears
on 3 lines, all incidental prose inside *other* entries' impact text. The
accurate claim is stronger: *the renderer emits no `type` field at all*, so a
grep for `failed` returns hits that have nothing to do with the failure row.
A count from grep is a hypothesis.

**The survived mutant, which is the most useful single result.** Mutation
**M-1** (drop `failed` from `aggregate.WithImpact`) fired **zero** of the
suite's 1070 tests. The repo has **no existing coverage of a `failed` entry
flowing through `WithImpact`, `impact` or `wrapped`.** Every test proving the
new behaviour must be written from scratch; none can be adapted, and a green
suite is not evidence.

### The Fork 3 finding, which belongs to the stage rather than to one spec

This page's *Design Notes* item 4 and its ranker paragraph say a failure
*"competes on recency alone and is crowded out by volume."* Measured, it is
stronger and structurally different: **it is not out-ranked, it is not a
candidate.**

`Gather` reads `List{Limit: PoolLimit}` with `PoolLimit = 200`, so a bare
`brag memory` sees only the 200 most recent entries. On the live corpus the
horizon is entry 207, dated 2026-07-06 — **61 days** — and **198 of 398
entries (49.7%)** cannot be returned at any budget. Five of the eight
`type: learned` entries are already outside it.

This kills the option that reads cleanest: **a fourth ordinal list cannot fix
it.** `buildMatchRank` drops ids not in the pool (`memory.go:191`) and
`buildProjectRank` iterates the pool, so no ranking term can introduce an entry
`Gather` did not fetch. Only a fourth *read* reaches an out-of-pool entry — and
that is pool composition, not fusion, which is why SPEC-085 could answer
`memory-slice-fusion-constants` cleanly in the negative.

**Consequence for this stage:** Success Criterion 3 (*"`brag memory` returns
failures to an agent reading the corpus cold"*) is **half met** by SPEC-085 —
the line shape labels a failure `[project/failed]`, but durability past the
61-day horizon is not addressed. Filed as
`memory-pool-composition-excludes-older-entries` with the measurement, and
deliberately not answered while there are zero failure entries to calibrate a
slot count against.

## Dependencies

### Depends on
- Nothing technical. `internal/storage`, `internal/memory`, `internal/export`
  and the DEC-014 envelope all ship today. Gates green at `1775feb`.

### Routed away from this stage, deliberately
- **Capture completeness** (staging inbox + git-import cold-start miner). Its
  `resume_when` fired on 2026-08-15 and fired again in substance during
  PROJ-007, when a stage close went uncaptured for two days. **Call made at
  framing: it is its own project, not part of this stage.** It is a
  *volume-of-capture* problem; this stage is a *kind-of-capture* problem, and
  the two share no code. Folding them would double this stage and delay the
  honest corpus behind a miner. Recommended as **PROJ-010**, ahead of PROJ-009.
- **Read-path discoverability** (`BRAG.md` is write-only; only a `Stop` hook
  exists, no `SessionStart`). Confirmed on the live tree 2026-09-05: `BRAG.md`
  has 22 headings and **none** is a read-the-corpus section; *"Your role, in one
  sentence"* casts the agent purely as a capturer; `brag memory` appears only
  inside the plugin-install block. **It belongs in PROJ-008 but not in this
  stage** — it should follow the stage that makes the corpus most worth reading,
  so the section written into `BRAG.md` describes a corpus that already holds
  failures. Revisit at STAGE-024 framing.

### Enables
- **STAGE-024 (impact quality)** — with a corpus that contains both shapes, so
  the classifier is calibrated on its whole input domain.
- The mirror, whose sharpest observations (*"three of these went nowhere"*)
  require the corpus to record that they went nowhere.

## Stage-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <one-line items>
- **What can a user do now that they couldn't before, at STAGE scope?** — one
  sentence, before → after; quote the confirming number if one exists, name the
  outcome if not. Write `none` if this stage had no user-visible outcome — a
  real, greppable result, not a blank.

  Not a concatenation of the spec-level answers: those are per-spec, and the
  unit anyone actually records is the stage. Read the spec answers, then say
  what the *stage* bought — the thing that is true now and was not when the
  stage opened. **If this answer is not `none`, capture it before moving status
  to shipped.** Evidence ref: the stage's PRs are known, so tag the one that
  closed it.
  - <answer | none>

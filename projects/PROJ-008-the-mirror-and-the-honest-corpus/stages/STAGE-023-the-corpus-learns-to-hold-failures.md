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

- [ ] SPEC-085 (frame) — the capture verb, its schema home, and its behaviour
      in `brag memory`; five forks, all evidenced.
- [ ] (not yet written) — digest posture: what `wrapped` / `impact` /
      `summary` / `story` do with a failure. **Only if SPEC-085's Fork 4 turns
      out to need more than a documented default.**

**Count:** 0 shipped / 1 active / 1 pending

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

---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-085
  type: story                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
  blocked: false
  priority: high
  complexity: M                    # S | M | L  (L means split it)

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-09-05
  framed_at: 2026-09-05

insight:
  confidence: 0.82                 # honest: the forks are well-evidenced, but
                                   # Fork 3 (memory participation) may split
                                   # into its own spec at design.

references:
  decisions:
    - DEC-014                      # the envelope every rule-based surface inherits
    - DEC-015                      # tags ARE normalized — supersedes DEC-004
    - DEC-043                      # memory RRF over three ordinal lists
    - DEC-044                      # the memory line shape — renders type, not tags
    - DEC-048                      # a provenance count names what it counted
  constraints:
    - one-spec-per-pr
    - no-sql-in-cli-layer
  related_specs:
    - SPEC-073                     # the memory slice this must not destabilise
---

# SPEC-085: `brag learn` — the corpus holds a failure

> **Cycle: frame.** **GO** at complexity **M**. Five forks, each with the
> evidence that decides it; none is settled here — framing's job is to make
> them decidable and to prove they are the right five. Every number was
> re-derived on 2026-09-05 against `main` at `1775feb`, corpus **397**, all
> five gates green.

## Context

`brag` holds **397 entries and not one failure.** That is not a capture-
discipline problem — there is no verb, no field value, and no documented
convention for recording work that did not work, so the absence is structural.
`BRAG.md`'s *"When to propose a brag"* section is entirely about wins, and its
anti-examples are about *vague* wins, not about honesty.

The consequence is that every surface built on the corpus is systematically
optimistic, and the two most important consumers are the ones that suffer most:

- **An agent reading the corpus cold** (`brag memory`, or the `brag://` MCP
  resources that load with no tool call at all, DEC-045) learns only what
  worked. It cannot avoid a path the user already burned two days on, because
  the burn was never written down.
- **The mirror** (STAGE-025) is supposed to say *"this project reads like a
  turnaround."* A turnaround has a downstroke. Over a wins-only corpus the
  mirror can only ever report a rising line.

This spec adds the capture path and decides where a failure lives. It stops at
the point where the corpus is *capable* of honesty. Ranking honesty is
STAGE-024; narrating it is STAGE-025.

### What framing checked and found different from the inputs

Three premises carried into this stage did not survive contact with the tree.
Each changes a fork, so each is recorded here rather than in a note:

1. **`type` is not a closed set.** The framing input described it as *"a small
   closed set."* `internal/cli/add.go:101` calls it *"free-form category"* and
   nothing validates it. Live: **19 distinct values**, 113 entries with none,
   and near-duplicate pairs (`shipped` 201 / `ship` 15; `fixed` 2 / `bugfix`
   1). This makes Fork 1 cheaper (no migration, no new concept) and Fork 2
   harder (a value guarantees nothing on its own).

2. **DEC-004 is superseded.** The input costed a reserved tag against DEC-004's
   comma-joined `TEXT` column. **DEC-015 replaced that on 2026-06-06** with
   normalized `tags` + `taggings` and two indexes. A reserved tag is an indexed
   lookup, not a substring scan — so the tag option is *far* cheaper than the
   input assumed, and Fork 1 has to be decided on the read path instead.

3. **`learned` is already occupied.** Eight entries carry `type: learned` and
   **all eight are wins about learning** — *"Verify caught an unpinned test
   seam that CI, five goldens and two guards all passed"*; *"Re-running a
   pre-flight after a dependency bump paid for itself."* None records a
   failure. Adopting `learned` would overload a value that already means
   something, across eight rows of live data.

## Goal

Give the corpus a first-class way to record work that did not work, and make it
survive the read path an agent actually uses — without ranking it, narrating
it, or changing what the celebratory digests currently say by accident.

## Fork 1 — where does a failure live: `type`, a reserved tag, or a column?

The one fact that decides this is on the **read** path, not the write path.

**DEC-044's memory line shape** (`internal/memory/memory.go:224`) is:

```
- <id> <YYYY-MM-DD> [<project>/<type>] <title> — <impact>
```

**`type` renders. Tags do not.** So a failure marked by `type` reaches the
agent-facing surface for free, and a failure marked by a reserved tag is
invisible there unless DEC-044's locked line shape is amended — which is a
change to the one surface SPEC-073 stabilised and DEC-044 locked.

Supporting evidence, measured across the eight commands:

| Surface | `--type` filter? |
|---|---|
| `list`, `impact`, `wrapped`, `summary`, `story` | yes (5) |
| `memory`, `search`, `stats` | no (3) |

So `brag list --type <value>` gives retrieval **for free** — Success Criterion
2 of the stage is satisfied by an existing flag if this fork lands on `type`.

**Leaning: `type`.** It renders on the read path, it needs no migration, and
retrieval already exists. **But the fork is not free**, and design must cost
the objection: a free-form `type` guarantees nothing, so *"this entry is a
failure"* becomes a convention rather than an invariant. Whether that needs
validation — i.e. promoting `type` to a partly-closed set — is **Fork 2's**
problem, and it may be the real work in this spec.

### Rejected alternatives (to be written up at design)
- **A new column** (`outcome`, or a boolean). A migration, forward-only per
  DEC-002, for a distinction the `type` column already expresses. Buy it only
  if Fork 2 concludes `type` cannot be trusted at all.
- **A reserved tag** (`outcome:failed`, in the `agent:`/`model:` family).
  Cheaper than the input implied, thanks to DEC-015 — but invisible in the
  memory line, which is the surface this stage exists to serve.

## Fork 2 — the verb, the value, and the eight occupied rows

Three sub-questions, and the third is the one with live data behind it.

- **The verb.** `brag learn` is the brief's name. Note it reads as *"record a
  lesson"*, which is what the eight existing `learned` entries already are —
  wins. If the value ends up being `failed`, a verb called `learn` that writes
  `type: failed` is a seam. Design should either align them or say why not.
- **The value.** `learned` is occupied (see Context). Candidates: `failed`,
  `abandoned`, `dead-end`, `learned` redefined. Whichever wins must be
  greppable and must not collide with the 19 values in use.
- **The eight rows.** Options: leave them (two meanings share one value),
  re-type them (touching live user data), or define the new value so it does
  not collide. **Framing's position: leave them and choose a
  non-colliding value.** Rewriting a user's corpus to fit a schema decision is
  a cost this spec should not impose; but the position must be *written down*,
  because a later mirror rule that reads `learned` will otherwise be wrong in
  eight places.

**Open question for design:** does this spec introduce validation of `type` —
a closed set, or a reserved-prefix rule — or does it add a value to a free-form
field and document the convention? This is the single biggest complexity driver
in the spec. Validation is a behaviour change on `brag add` for every existing
user and would need its own DEC.

## Fork 3 — how does a failure participate in `brag memory`?

**Today it would not, reliably.** DEC-043 fuses three ordinal lists — recency,
relevance, project — by RRF, with `PoolLimit = 200`. There is **no type-aware
or tag-aware signal**, and `memory` has no `--type` flag. A failure competes on
recency alone, so in a corpus growing at ~20 entries/week it is crowded out
within weeks of being written.

That is the whole point of the stage failing quietly. Three options:

1. **Do nothing.** Failures compete on recency like everything else. Honest,
   zero risk to SPEC-073, and defensible — but it means the agent-facing
   promise ("learns what didn't work") holds only for recent failures.
2. **A fourth ordinal list.** Structurally consistent with DEC-043, and the
   cleanest fit — but it changes the fusion, and **`memory-slice-fusion-constants`
   is an open `medium` question in `guidance/questions.yaml`** (raised
   2026-08-08, DEC-043 confidence 0.75) asking whether `k=60` and equal weights
   are right at personal scale. Adding a fourth list without answering it
   compounds an unresolved question.
3. **A guaranteed floor** — reserve N slice slots for failures.
   Budget-affecting, so it touches DEC-044's token budget.

**Framing's read: option 1 for this spec, and say so out loud in the docs**, with
the fourth-list option costed and routed to STAGE-024 or its own spec. Rationale:
this spec should not be the one that destabilises the memory slice, and option 2
cannot be done honestly while the fusion constants are an open question.

**Required of design either way:** state explicitly whether this work answers
`memory-slice-fusion-constants`. The stage owes that answer in both directions —
"yes, and here it is" or "no, and it stays open." Silence is the failure mode.

## Fork 4 — what do the digests do with a failure?

The brief flags this as *"decide before something ships that reads badly in
December."* Measured today, the risk is concrete and larger than the brief
stated.

`brag wrapped`'s **"Impact moments"** section and `brag impact`'s body are, at
today's corpus, **identical but for one trailing newline** — verified by
extracting both sections and diffing (694 vs. 695 lines; `diff` output is
exactly one line, `694a695 > `). Both carry **all 323 impact-bearing entries**,
inside a digest whose own help calls it *"shareable, celebratory."*

The header reads `Entries: 323/397 with impact`.

So the moment a failure carries an `impact` — and it should; *"cost two days
and produced nothing reusable"* is a real impact statement — **it appears in
the celebration by default.** Nobody has to do anything wrong for that to
happen.

And it **cannot be fixed with an existing flag**: `--type` is exact-match
*inclusion*. There is no negation, so *"wrapped, minus failures"* is code, not
configuration.

Options: exclude by default; include under a named section (*"what didn't
work"*); include silently. **Framing takes no position** — this is a product
call about what `wrapped` is for, and it belongs to the user. Design must
surface it as a decision with the December reading spelled out, not resolve it
quietly.

**Scope guard:** whatever is chosen must apply to `impact`, `wrapped`,
`summary` and `story` **consistently**, and must not require the impact-quality
classifier (STAGE-024) to exist.

## Fork 5 — the envelope

Any surface this spec adds or changes inherits **DEC-014** (the single-object
JSON envelope and the markdown provenance block) and **DEC-048** (the headline
count must name what it counted, binding all six `internal/export/` renderers
forward).

The trap is specific and DEC-048 exists because of it: if a digest starts
excluding failures, its count changes meaning — `Entries: N` can no longer be
`len(entries)` if failures were dropped, and DEC-048 requires it be **named
something else** rather than silently redefined. `brag impact`'s
`Entries: <shown>/<in-window> with impact` (DEC-028) is the existing precedent
for a two-number form.

Not really a fork so much as a checklist item — recorded as one so it cannot be
skipped.

## Complexity: M — and what would make it L

**M**, on the assumption that Fork 1 lands on `type` and Fork 3 lands on
option 1 (no ranker change). That is: a new verb, a value, doc updates, digest
posture, and tests.

It becomes **L**, and should split, if either:

- **Fork 2 introduces `type` validation.** That is a behaviour change on
  `brag add` for every user and a DEC of its own.
- **Fork 3 takes option 2 or 3.** Touching DEC-043's fusion or DEC-044's budget
  pulls in the open `memory-slice-fusion-constants` question and destabilises
  the slice SPEC-073 stabilised.

Design should re-estimate after Forks 2 and 3 and **split rather than absorb**.

## GO / NO-GO

**GO.**

- The problem is measured, not asserted: 397 entries, zero failures, and a
  read path (DEC-044's line) that determines the schema choice.
- The forks are decidable with evidence that exists today.
- It is the stage's first spec and blocks nothing else in flight.
- Gates are green at `1775feb`; there is no in-flight work to collide with.

**What would flip it to NO-GO:** if design finds that Fork 2 forces `type`
validation *and* Fork 4 forces a digest behaviour change, the spec is two specs
and should be split before build rather than run as an L.

## Acceptance Criteria

*Written at design.* Framing's requirement is that they be stated as
**numbers to diff against**, not prose — the corpus count of failure-typed
entries, the exact `Entries:`/`Candidates:` lines on each touched surface, and
the line count of any changed golden.

## Failing Tests

*Written at design.* Note for design: the stage has a standing trap —
**a NOT-contains assertion proves the old phrase is gone, never that the new
one is true.** Both PROJ-007 stages recorded this and neither resolved it. Any
guard added here needs a positive assertion paired with each negative one.

## Implementation Context

### Decisions that apply
- **DEC-014** — the envelope. Inherited by every rule-based surface.
- **DEC-015** — tags are normalized (`tags` + `taggings`). **Supersedes
  DEC-004**; do not cost a reserved tag against the comma-joined model.
- **DEC-043** — memory RRF over recency/relevance/project, `PoolLimit = 200`.
  No type-aware signal exists.
- **DEC-044** — the memory line shape renders `[<project>/<type>]`. **The fact
  that decides Fork 1.**
- **DEC-048** — a provenance count names what it counted; binds six renderers.

### Constraints that apply
- `one-spec-per-pr`.
- `no-sql-in-cli-layer` — binds **production** code under `internal/cli/`
  (DEC-047 / SPEC-083). Persistence goes through `internal/storage`.

### Prior related work
- **SPEC-073** stabilised the memory slice. Fork 3's default exists to avoid
  destabilising it.
- **`guidance/questions.yaml` → `memory-slice-fusion-constants`** — open,
  `medium`, raised 2026-08-08. Fork 3 must address it explicitly.

### Out of scope (for this spec specifically)
- The impact-quality classifier (STAGE-024). Measured 2026-09-05, it is **not**
  the small primitive the brief assumed: candidate rules select between 11 and
  230 of 397 entries, 141 of 230 digit-bearing impacts contain an *identifier*
  digit rather than a measurement, and 27 quantified impacts use spelled-out
  numerals. That belongs to its own framing.
- The mirror; story-surface v2.
- Retroactively re-typing the eight `learned` entries.
- `BRAG.md`'s read-path gap — routed at stage level.

## Notes for the Implementer

Framing settled nothing except which questions are the right ones. Do not treat
any "leaning" above as decided — each is an argument with its evidence attached
so design can overturn it cheaply if the tree disagrees.

Run the **§12(b) pre-flight**: write every literal into a throwaway worktree and
run it through its own tool, then record what the tool actually said. The last
three specs were near-frictionless at build because of it.

## Build Completion

*Filled at build.*

### Build-phase reflection (3 questions, short answers)

## Reflection (Ship)

- **What can a user do now that they couldn't before?** — one sentence,
  before → after. Capture this before closing the cycle.

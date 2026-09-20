---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-086
  type: story                      # epic | story | task | bug | chore
  cycle: design                    # frame | design | build | verify | ship
  blocked: false                   # UNBLOCKED 2026-09-06: SPEC-085 shipped
                                   # (df369e9, PR #199). The reserved `failed`
                                   # value exists and the corpus now holds one.
  priority: high
  complexity: M                    # S | M | L  (L means split it)
                                   # Re-checked 2026-09-06: L as scoped at
                                   # framing; held at M by splitting the
                                   # summary/story half out. See ## Complexity.

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5
  created_at: 2026-09-05
  framed_at: 2026-09-06                # re-framed; first framed 2026-09-05
  designed_at: 2026-09-18              # after SPEC-088 shipped (#218), main 3201f50

insight:
  confidence: 0.88                 # up from 0.85 at design. Every fork is now
                                   # settled by a measurement on the tree and
                                   # pinned by a test that fires under mutation
                                   # (19 probes). What keeps it below 0.9 is
                                   # DEC-050's row 4 (story), which extends the
                                   # user's reasoning — but that is SPEC-094's
                                   # to implement, not this spec's.

references:
  decisions:
    - DEC-014                      # the envelope, incl. part 4's empty-state rule
    - DEC-028                      # impact's two-number count + the 4-key projection
    - DEC-029                      # story profiles are DATA, not a Go enum
    - DEC-030                      # wrapped's section arc
    - DEC-048                      # a count must name what it counted
    - DEC-049                      # the reserved `failed` value (SPEC-085)
    - DEC-050                      # EMITTED at this spec's design: the failure
                                   # posture for all seven --type surfaces
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-085                     # shipped the verb and the value this consumes
    - SPEC-087                     # split out here: Y3 derives instead of caching
    - SPEC-094                     # split out here: summary + story (DEC-050
                                   # rows 3-4); id claimed by file at design
---

# SPEC-086: what the celebratory digests do with a failure

> **Cycle: design.** Designed 2026-09-18 against `main` at `3201f50`, after
> SPEC-088 shipped (#218). **GO at M, held.** Framing's record is kept below as
> it was written. Wherever design re-measured a number or overturned a claim,
> the design sections say so and take precedence. **Start at *What design
> settled*.**
>
> Every design number was measured against a **frozen file copy** of
> `~/.bragfile/db.sqlite` taken 2026-09-18 with `sqlite3 .backup`, which takes
> a read lock only. The copy holds 606 entries, max id 628, and four
> `type: failed` rows. The live file was never opened for writing, and the
> copy's SHA-256 (`d8189d214c19…`) was the same before and after every run
> against it. A number measured on a toy fixture is labelled as one.
>
> *Framing's own header, kept:* every framing number was re-derived on
> 2026-09-06 against `main` at `df369e9`, on a frozen copy of 420 entries with
> max id 431 and one `type: failed` row.

---

## What design settled

| | Decision | Pinned by |
|---|---|---|
| **Fork A**: the DEC-048 obligation | **Zero renames, measured.** The with-impact subset is **split** between two sections and never narrowed, so every headline and every count key keeps its meaning. `Entries: 531/606 with impact` and `Entries: 606` render byte-identically before and after on the frozen copy. `impact`'s `counts_by_project` is the one count a build could silently narrow, and it stays computed over both sections. There is no per-section count line. The JSON change is that two arrays lose their failure rows, which is breaking, and each envelope gains one key. | `TestToImpactJSON_CountsByProjectSpansBothSections`; mutation M-D |
| **Fork C**: where the predicate lives | `internal/aggregate` holds `FailureType` (moved from `internal/cli`), `IsFailure` and `SplitFailures`. **It does get a drift guard.** Framing said there was no SQL counterpart, and that was wrong: `brag list --type failed`, whose SQL is `e.type = ?`, is DEC-049's own retrieval definition of a failure. | `TestFailureClassifier_GoPredicateMatchesTypeFilter`; M-A, M-B, M-S |
| **Fork D**: the empty section | `## What didn't work` renders **only when it has an entry**, on both surfaces. `impact` likewise drops a bare `## Impact` when every with-impact row is a failure. JSON always carries the key, as `[]` when empty. `wrapped`'s five DEC-030 sections are untouched. | `TestToImpactMarkdown_SectionsRenderOnlyWhenNonEmpty`, `TestToWrappedMarkdown_WhatDidntWorkRendersOnlyWhenNonEmpty`; M-E, M-F, M-G |
| **Fork E**: goldens | Found by running the prototype. **Two existing goldens move, both JSON, both in `internal/export`,** by one line each. **No markdown golden moves and nothing in `internal/cli` moves**, because Fork D omits the empty section and Fork A renames nothing. `memory_test.go:247` does not move. | the full suite on the prototype |
| **DEC-050** | Written as `decisions/DEC-050-a-failure-is-never-rendered-as-a-win.md`, with `insight.type: decision` and confidence **0.85** (0.80 as first written; see *Maintainer rulings*). It states the posture for all seven `--type` surfaces. | `Y3`, `Z7` and `AC1` green on this branch |
| **DEC-030** | Gains `## Amendment (2026-09-18, SPEC-086 design)`: the arc gets a sixth section, and it is the only conditional one. | `TestToWrappedMarkdown_FailureSectionGolden` |
| **DEC-028** | Gains `## Amendment (2026-09-19, SPEC-086 design)`, on the maintainer's R3 ruling: `impact_by_project` excludes failures, `failures_by_project` is the new last key, `counts_by_project` still counts both sections, and the headline is unchanged. Design had filed this as **NO CHANGE**; that was overturned. | `TestToImpactJSON_DEC028ShapeGolden`, `TestToImpactJSON_CountsByProjectSpansBothSections` |
| **Successor** | **SPEC-094**, claimed with a file at `cycle: frame`. STAGE-023's former *(not yet written)* entry now points at it, in the same edit. | — |

---

## Maintainer rulings (2026-09-19)

Design asked four questions. **Three came back**, and each changed a document
rather than the build. **No diff block below moved**: the embedded literals
hashed the same before and after
(`869e3b7f6acb9305…`, from the `awk` recipe that extracts every fenced
`diff` block), and no Go, test or `test-docs.sh` literal changed. **The fourth question needed no
ruling**: whether SPEC-094 gates v0.7.0 was already answered by the plan — it
is the second half of the gate, which is what STAGE-023's backlog line and
DEC-050's *"until SPEC-094 ships"* consequence already say.

| | Question | Ruling | What it moved |
|---|---|---|---|
| **R1** | `story`'s promotional profiles — must a failure survive into every profile? | **Narrower than this design wrote it.** Two invariants only: never rendered as a win, never *silently* dropped. A promotional profile (`candor: promotional`; today `exec`, `skip`) **may omit failures, with a visible note** — a count is the obvious form. The mechanism is SPEC-094's to design. | DEC-050 row 4, its first Alternative, T3, and both confidence statements (0.80 → **0.85**, the row itself ~0.70 → ~0.90 and no longer the softest part). SPEC-094's row-4 input and the decision it still owes. |
| **R2** | A failure with no impact is in neither section but still in the totals — is that right for v0.7.0? | **Accepted**, on the impact-first symmetry (DEC-028 choice 3). It was already a recorded negative consequence and already pinned by an embedded test: `impactFailureFixture`'s id 7, `retry-noimpact`, carries `Impact: ""`, and `TestToImpactMarkdown_FailureSectionGolden` asserts `Entries: 5/8` with that id in **neither** section. | DEC-050's Consequences (the acceptance, dated) and a new revisit trigger **T5**. **No code or test changes.** |
| **R3** | DEC-028 is *refined* by DEC-050 rather than amended — is that enough? | **No. Amend it.** Its key list and its one-section body go stale the moment failures move out. | `decisions/DEC-028-…md` gains `## Amendment (2026-09-19, SPEC-086 design)`, original text untouched (LD6). The premise audit's `DEC-028 … NO CHANGE` row, the design-commit `## Outputs` list, the inventory prediction (`## Amendment` row 1 → **3**, not 2) and DEC-050's DEC-028 reference all move with it. |

**Still open, and not touched here:** `guidance/questions.yaml`'s
`dec-amendment-heading-convention`. Three records now carry a `## Amendment`
heading, which is most of what that question asked for. Closing it would move
`Y4`'s *still open* pin from **8**, and `Y4` is hand-pinned
(`scripts/test-docs.sh:1700-1701`) — SPEC-092 has not made it derive yet. So
the question stays open and the register is untouched.

---

*Everything from here to **## GO / NO-GO** is framing's record (2026-09-05,
re-framed 2026-09-06), kept as written. The design sections after it take
precedence wherever the two disagree, and each disagreement is named there.*

---

## Status change at re-framing: the interim risk is no longer hypothetical

SPEC-085 accepted a named interim risk: between it shipping and this spec
shipping, a `failed` entry carrying an impact statement *would* appear silently
in *Impact moments*. It was bounded on the argument that **the corpus holds zero
such entries**.

**That is no longer true, and it stopped being true before this spec was
re-framed.** Measured 2026-09-06:

```
$ brag list --type failed --format json | grep -c '"id":'
1
```

| | |
|---|---|
| id | **420** |
| project | `contextcore-pilot-harness` |
| type | `failed` |
| created_at | **2026-09-06T00:44:44Z** |
| impact | present, 4 sentences, opens *"Cost about an hour and produced a branch that was dropped."* |
| tags | `…,agent:claude-code,model:claude-opus-5,session:…` — agent-authored |

Its own description records why it exists: *"Recorded as a `failed` entry
because a corpus of only wins is a worse input to a retro than an honest one.
(The installed binary has no `brag learn` verb yet, so this uses the reserved
type directly…)"* — the released binary is v0.6.1; `brag learn` is on `main` and
not yet cut.

**So the risk is realised, not merely real.** The four leaks below are not
predictions about a future corpus; they are what `brag` prints today against the
user's own database. Two consequences:

1. **This is now the argument for the spec's priority**, and it decays in the
   wrong direction: every additional failure written before this ships is
   another row laundered on four surfaces.
2. **The premise that bounded the risk is spent.** SPEC-085's bound was *"the
   user controls when the first one is written."* The first one is written, and
   it was written by an agent following `BRAG.md`'s documented path — which
   SPEC-085 verify had already flagged as the leaking shape.

This is *separate from* the one drafted `brag learn` entry the orchestrator is
deliberately holding until this spec lands. That one is still held; entry 420 is
a different failure, in a different project, and nobody held it. **The zero was
a choice for exactly one day.**

---

## What was re-measured, and what moved

Framing does not inherit numbers. Everything below was re-derived on
2026-09-06 at `df369e9`.

| Claim, as previously written | Where it came from | Measured 2026-09-06 |
|---|---|---|
| corpus **398**; `impact` reads `Entries: 324/398 with impact`; `wrapped` reads `Entries: 398` | this spec, §*What was measured* | corpus **420**; `Entries: 346/420 with impact`; `Entries: 420` |
| corpus **400**; `impact` `326/400`; `wrapped` `400` | framing prompt | same as above — the prompt was already stale |
| **0** rows of `type: failed` | framing prompt | **1** (id 420) — see above |
| **18** distinct non-empty `type` values | STAGE-023 design corrections | **19** — `failed` joined the set, as DEC-049 intended |
| `--type` inclusion is `store.go:388` | this spec Fork A, STAGE-023 | **`store.go:389`**. `:388` is the `if f.Type != ""` guard; `:389` is `conds = append(conds, "e.type = ?")`. The negation precedent at **`:404`** is cited correctly. |
| `story --audience exec` renders a failure with **zero occurrences of `failed`** anywhere in the markdown | framing prompt | **False as stated, true in substance.** The literal string `failed` appears on **3 lines** of that document — all of them incidental prose inside *other* entries' impact text (one quotes `"failed"` as a type name, one is a dashboard's `0 failed`, one is `it failed in all 17 runs`). **None is a type label, and entry 420 carries none.** The accurate claim is stronger: *the renderer emits no `type` field at all*, so a grep for `failed` returns hits that have nothing to do with the failure row. A count from grep is a hypothesis. |
| Fork E: *"five files in `internal/export/` carry the affected strings"* | grep for the idea | **Wrong shape and wrong package set.** Measured by mutation — see **Fork E**. |
| Fork D: *"DEC-014 part 4 has a precedent (`wrapped` omits body sections on an empty period)"* | this spec | **The precedent covers the wrong case.** See **Fork D**. |

Two things that did **not** move, re-confirmed rather than inherited:

- **`wrapped`'s *Impact moments* and `impact`'s body are still the same
  document, rendered twice.** `wrapped.go:92-95` reimplements the
  `aggregate.WithImpact` + `GroupEntriesByProject` pair rather than calling
  `ToImpactMarkdown`. Both render **346** entries today.
- **Neither surface renders `type`, in either format.** `impact.go:60-61` emits
  `- %d: %s` + `  %s`; `impactEntry` is the 4-key `{id,title,project,impact}`
  projection (DEC-028 choice 4), and `wrapped`'s `impact_moments` uses the same
  4 keys. **A failure is unrepresentable as a failure on these two surfaces in
  BOTH formats.** This is the fact that makes them different from `story`.

---

## The decision, and who made it

**Chosen by the user, 2026-09-05:** failures get their **own named section** —
`## What didn't work` — on `brag wrapped` and `brag impact`. They are pulled
*out* of *Impact moments* / the impact body, not hidden and not mixed in.

**Rejected, with the user's reasoning:**

- **Exclude by default.** `wrapped` stays purely celebratory, failures reachable
  only behind a flag. Rejected because it re-creates, inside the digest people
  actually read, the flattery PROJ-008 exists to remove.
- **Include silently** (status quo, zero code). Rejected — and independently
  ruled out by STAGE-023's success criterion that *"the celebratory digests do
  not silently absorb failures."*

The shape the user selected:

```
## Impact moments

### AI Development Factory

- 74: Shipped the single-worker AI dev factory PoC
  Proved the loop end-to-end...

## What didn't work

### AI Development Factory

- 411: Tried a shared-worker pool to cut cold starts
  Cost two days and produced nothing reusable.
```

**Unchanged at re-framing.** Nothing measured since contradicts it.

---

## The forks this spec must settle

### Fork A — the DEC-048 obligation, and its dependency

**Framing's claim was "two DEC-048 count renames." Measured, it is at most
one, and possibly zero renames plus two new per-section counts.**

DEC-048's rule: *"`Entries: <N>` means the number of entries the document
covers, and it narrows when a filter narrows the document. A consumer whose
headline count is not that must name it something else."*

Run that rule against each headline as it actually reads today:

| Surface | Headline today | Does sectioning change what it counts? |
|---|---|---|
| `wrapped` | `Entries: 420` | **No.** `wrapped.go` computes `len(entries)` — every entry in the period. Moving failures between two body sections does not change the document's coverage. **The headline stays true and needs no rename.** |
| `impact` | `Entries: 346/420 with impact` | **Only if the selection changes.** 346 = `len(aggregate.WithImpact(entries))`. If failures are *re-sectioned* but still counted, 346 is unchanged and still true. If failures are *removed* from `WithImpact`, the number becomes 345 while the label still says `with impact` — and entry 420 **does** carry an impact, so the label goes false. |

So the real obligation is narrower and sharper than "rename two counts":

> **Once the body has two sections, the single headline number stops accounting
> for the document.** It is not that `Entries:` becomes a lie — it is that it
> stops telling you how the 346 split. DEC-048's spirit is *a count names what
> it counted*; a document with a `## Impact moments` of 345 and a
> `## What didn't work` of 1, under a single `346`, has a count that names the
> union and nothing else.

**What design must choose between** (both are DEC-048-compliant; this is a
shape call, not a correctness call):

1. **A third number in the headline** — `Entries: 345/1/420` or
   `Entries: 346/420 with impact (1 failed)`. Follows DEC-028's precedent for
   reporting a pair in one line.
2. **Per-section counts**, headline untouched — each `##` heading carries its
   own tally. Leaves `wrapped`'s `Entries: 420` and `impact`'s `346/420`
   verbatim, which is attractive because both are *already correct*.

**The JSON side is additive, not a rename.** `impactEnvelope` has
`entries_in_window` / `entries_with_impact`; `wrapped` has `total_entries` and
`impact_moments`. A new section needs a new key (`failures_by_project` or
similar). Whether `impact_by_project` / `impact_moments` **lose** their failure
rows is the same selection choice as above, and *that* is the breaking wire
change — not the headline.

#### The dependency: `--type` negation is inexpressible and fails silently

Carried onto STAGE-023 with SPEC-086 named as owner. **Re-confirmed
independently on 2026-09-06** against the frozen snapshot:

```
$ for t in '!failed' '-failed' 'shipped,failed' '!=failed' 'NOT failed'; do
    brag --db <snap> list --type "$t"; echo "exit=$?"
  done
# every one: zero rows, exit 0, empty stderr
$ brag --db <snap> list --type ''
brag: user error: --type must not be empty          # the ONLY spelling that errors
```

`internal/storage/store.go:389` is exact-match inclusion on a single string.
A user filtering failures out of a review gets an empty document rather than a
message.

**The call: SPEC-086 DESIGNS AROUND IT. It does not fix it, and it does not
route it as a blocker — because it is not one.**

The evidence that it is not a dependency, run rather than reasoned. Mutation
**M-1** changed `aggregate.WithImpact` (`internal/aggregate/aggregate.go:297`)
from `if e.Impact != ""` to `if e.Impact != "" && e.Type != "failed"`, confirmed
by content hash, and ran the full suite:

```
PRE  ddd33f3ae63a04f2703baea402896bfc267200bfb1332124bdb4e74a51321ba8
POST e7d171eecedadf0a88bbfe0346924ceca2b2c9ca4921c183768ced258b3289cc
$ go test ./...          # 0 failures — see Fork E for what that means
```

**A one-line in-memory predicate does the entire selection job.** These two
renderers read all in-window rows once and partition in Go — exactly as
`brag coverage` does, which needs *both* provenance classes for a share and
therefore deliberately does **not** set `ListFilter.Author`. There is no query
in this spec that wants `--type '!failed'`.

The silent-failure bug is **real, user-facing, and out of scope here**: it is a
storage/CLI input-validation defect whose blast radius is all seven `--type`
surfaces, and fixing it inside a digest-posture spec is the scope creep this
repo splits on principle. **Routed to STAGE-023's backlog as a named `bug`
entry** (see the stage page), not to `guidance/questions.yaml` — a question is
for *"we do not know"*, and here the behaviour, the line, and the fix shape are
all measured. Filing it as a question would also move `Y4`'s pinned counts and
force an inventory regeneration for zero information gain.

---

### Fork B — scope: which of the seven surfaces? *(rewritten; premise overturned)*

#### The framing that is now REJECTED

> ~~*"`wrapped`/`impact` are celebratory; `summary`/`story`/`export`/`coverage`
> are neutral, and a neutral surface arguably needs no section at all."*~~

**Rejected on measurement, not on taste.** Two of those four "neutral" surfaces
render a failure wrong. Recorded here as rejected so design does not re-derive
it, per §12's *decide-at-design-time-when-decidable* discipline.

Two things the rejected framing got structurally wrong, both of which must be
written into whatever replaces it:

1. **"Is this surface celebratory?" is a property of the PROFILE, not the
   command.** `brag story` carries four bundled profiles and is
   **user-extensible** (`~/.bragfile/story-profiles/<name>.yaml`, DEC-029 choice
   2 — profiles are data, not a Go enum). Two are promotional and two are
   candid:

   | profile | `candor:` | `impact_threads_only` | `drop_impactless_beats` |
   |---|---|---|---|
   | `exec` | **promotional** | true | true |
   | `skip` | **promotional** | true | false |
   | `manager` | candid | false | false |
   | `me` | candid | false | false |

   A per-command answer cannot express this. **Note that `skip` is promotional
   too** — the leak is not exec-only, which is how it has been described so far.

2. **The markdown/JSON asymmetry is the actual defect shape on `story`.** The
   data is present and the renderer drops it:

   ```
   $ brag story --audience exec --quarter --format json | grep -c '"type": "failed"'
   1
   ```
   ```json
   {"id": 420, "title": "Rebuilt a colleague's unmerged branch…",
    "project": "contextcore-pilot-harness", "type": "failed",
    "is_impact_beat": true, "created_at": "2026-09-06T00:44:44Z"}
   ```

   That is a *different fix* from `impact`/`wrapped`, where DEC-028's 4-key
   projection means a failure is unrepresentable in **both** formats without a
   renderer change. `story` needs a renderer to stop discarding what it already
   has; `impact`/`wrapped` need a renderer to start carrying what it never had.

#### The answer, for all seven — measured, and stated even where it is "nothing"

The seven were re-derived from the binary, not from the page. Unit stated
explicitly, because framing has already been wrong on this count once (it
claimed five):

```
$ for c in <every subcommand>; do brag $c --help | grep -E '(^|,)\s*(-[a-zA-Z], )?--type\b'; done
```

**Seven commands take `--type` as a FILTER**: `list`, `export`, `summary`,
`impact`, `wrapped`, `story`, `coverage`. An eighth, `brag add`, has `-k,
--type` as a **write** flag — a different unit, and the reason a naive count
lands on eight. `brag learn` deliberately has no `--type` at all (DEC-049).

| # | Surface | What it does with entry 420 today | Verdict |
|---|---|---|---|
| 1 | **`impact`** | `- 420: <title>` + impact text under `## Impact`, no `type`, in markdown **and** in the 4-key JSON. Unrepresentable as a failure. | **CHANGE — this spec.** The user's decision. |
| 2 | **`wrapped`** | Identical rows under `## Impact moments` (346 of them, same as `impact`). `## Rhythm → Top types` is a **top-3** list (`shipped 211`, `milestone 29`, `ship 15`) so `failed: 1` will never surface there. Nothing in the document names it. | **CHANGE — this spec.** Worse than `summary`: no honest counterpart anywhere. |
| 3 | **`summary`** | **Split.** `## Summary → By type` **does** print `failed: 1`, and JSON `counts_by_type` carries `"failed": 1`. But `## Highlights` lists it as `- 420: <title>` with no type, and its JSON highlight is a **2-key** `{id,title}` — narrower than `impact`'s 4-key. | **CHANGE — split to a successor.** The aggregate is honest; the per-entry section calls a two-day dead end a *Highlight*. |
| 4 | **`story`** | `- ★ 420: <title>` in **all four** bundled profiles; no `type` in markdown; `--format json` carries `"type": "failed"`. Under `exec`/`skip` the document is a **prompt** that also prints *"build the narrative from those outcomes"*, *"Terse and promotional… No process, no messy middle"*, *"Quantify wherever the impact beats give you a metric."* Headline is `Beats: 230/256`. | **CHANGE — split to a successor.** Strictly worse than `impact`: `impact` shows a human a mislabelled row, `story` instructs a model to launder it. |
| 5 | **`export`** | Renders `\| type \| failed \|` in the per-entry table **and** `**By type** → failed: 1`; JSON is the full 9-key entry. Already honest. | **NOTHING.** Genuinely neutral — this half of the rejected framing survives, and is now verified rather than assumed. |
| 6 | **`coverage`** | Provenance-only. Never renders per-entry content; `--type` narrows the population and nothing else. | **NOTHING.** Genuinely neutral. |
| 7 | **`list`** | Plain (default) output is 3 tab-separated fields — `id`, `created_at`, `title` — **no type**. `--format tsv` and `--format json` both carry `type`. | **NOTHING, and say why.** `list` is a raw index that makes no claim about its rows; you reach a failure by asking for it (`--type failed`, which is DEC-049's sanctioned retrieval path). An unlabelled row in an unfiltered index is not a mislabelled row in a narrative. Design must state this rather than leave it unanswered. |

**Consistency, per STAGE-023's scope guard.** The guard requires the answer to
apply to `impact`, `wrapped`, `summary` and `story` *consistently*. That is
satisfied by writing **one decision record (DEC-050) covering the posture for
all seven surfaces**, and implementing it in two specs — this one for rows 1–2,
a successor for rows 3–4. The decision is single; only the renderer work splits.
This mirrors DEC-049, written once in SPEC-085 and binding forward.

**What design still owes on `story`, and it is not small.** `internal/story/
profile.go:24` documents `Candor` as *"metadata surfaced to the LLM, **not a
body rule**."* Making a promotional profile behave differently in the body
turns a declarative field into a behavioural one — a semantic change to an
existing field's contract, and precisely the kind of thing DEC-029 made data
rather than code. That is a decision, not an implementation detail, and it is
the single strongest reason `story` does not belong in this spec.

---

### Fork C — where the predicate lives

`aggregate.WithImpact` is the current split point (`internal/aggregate/
aggregate.go:294-302`). A failure predicate must not live in `internal/cli/`
twice, and `internal/aggregate` is SQL-free and dependency-free by design —
which makes it the natural home, but it means the `failed` constant moves out of
`internal/cli`.

**Re-confirmed at re-framing.** M-1 (above) shows a predicate in
`aggregate.WithImpact` is sufficient and touches nothing else. The
single-sourcing precedent still applies: `aggregate.IsAgentAuthored` is kept in
agreement with storage's SQL clause by a cross-package drift-guard test
(`TestProvenanceClassifier_GoPredicateMatchesSQLClause`).

**One nuance design must not skip.** The provenance case single-sources a Go
predicate against a *SQL clause* because both exist. Here there is no SQL
counterpart — nothing in `internal/storage` knows `failed` is special, and
DEC-049 deliberately kept it that way (the verb pins the value; the field is not
validated). So the drift risk is not Go-vs-SQL; it is `internal/cli`'s `learn`
verb writing one literal while `internal/aggregate` reads another. **That** is
what a guard test must pin.

---

### Fork D — the empty case *(the cited precedent covers the wrong case)*

On a corpus with no failures — which is most corpora, and was this one until
yesterday — does `## What didn't work` render empty, or is it omitted?

**Framing cited DEC-014 part 4 (*"`wrapped` omits body sections on an empty
period"*) as the precedent. Measured, that rule fires on a different case than
the one this fork is about, and the two surfaces already disagree.**

Probe: a scratch DB with one entry that has **no impact** — a non-empty period
containing an empty section.

```
$ brag --db <scratch> wrapped              # 1 entry, 0 with impact
Entries: 1
## Cadence  … ## Top initiatives  … ## Impact moments  … ## Rhythm  … ## Span
                                          ^ heading renders, body empty
$ brag --db <scratch> wrapped | sed -n '28,34p' | cat -e
- demo: 1$
$
## Impact moments$
$
## Rhythm$

$ brag --db <scratch> impact --year        # same data
Entries: 0/1 with impact                   # document ENDS here — no body at all

$ brag --db <scratch> wrapped --format json | jq .impact_moments
[]                                         # non-nil empty array, per DEC-014 part 4
```

And the rule that *was* cited, on a genuinely empty period:

```
$ brag --db <scratch> wrapped 2024
Entries: 0                                 # document ends — every section omitted
```

So, precisely:

- **DEC-014 part 4 governs the empty *document*** — both surfaces omit
  everything. Confirmed, unchanged, and **not the case this fork asks about**.
- **For an empty *section* in a non-empty document the two surfaces already
  behave differently**: `wrapped` renders a bare heading (its DEC-030 section
  arc is fixed), `impact` omits its whole body.
- **JSON is settled** for both: empty array, non-nil.

**This is therefore a live fork, and getting it wrong is the "empty accusatory
heading" outcome the spec named.** If `## What didn't work` follows `wrapped`'s
arc convention, every user with a clean quarter grows a permanent empty heading
that reads as an accusation. If it follows `impact`'s convention, the two
surfaces diverge further. Design must pick and pin it with a test on **both**
surfaces; DEC-030's locked arc means the `wrapped` half is a DEC-030 amendment,
not a free choice.

---

### Fork E — goldens *(enumerated by execution, not by grep)*

Framing's Fork E said *"`internal/export`'s byte-exact goldens all move"* and
the prompt carried *"five files in `internal/export/` carry the affected
strings."* Both are grep-shaped hypotheses. **Measured by mutation under §12's
protocol** — content hash before and after, restore from a scratchpad backup,
hash confirmed returned:

| Probe | What it mutated | Tests that fired | Files | Packages |
|---|---|---|---|---|
| **M-1** | `aggregate.WithImpact` drops `Type == "failed"` | **0** | — | — |
| **M-2** | one `## What didn't work` heading emitted in `impact.go` **and** `wrapped.go` | **4** | `impact_test.go`, `wrapped_test.go` | `internal/export` |
| **M-3** | `Entries:` → `Wins:` on both headlines (the DEC-048 rename shape) | **9** | `impact_test.go`, `wrapped_test.go`, **`internal/cli/wrapped_test.go`** | `internal/export`, **`internal/cli`** |

Four corrections to Fork E, each of which changes what design must enumerate:

1. **M-1 is a SURVIVED MUTANT — and that is the most important result here.**
   The entire 1070-test suite has **zero coverage of a `type: failed` entry
   flowing through `WithImpact`, `impact` or `wrapped`.** The selection change
   costs no existing golden churn *because nothing tests it*. Every test that
   proves the new behaviour must be written from scratch; none can be adapted.
   This is also the standing trap SPEC-085 closed with paired assertions: a
   green suite here is not evidence, and a NOT-contains assertion would pass on
   an empty corpus.
2. **The blast radius is not confined to `internal/export/`.** M-3 reaches
   `internal/cli/wrapped_test.go`. Fork E's package assumption was wrong.
3. **It is 2–3 files, not 5.** The five-file grep hit includes
   `coverage_test.go`, `markdown_test.go` and `memory_test.go`, none of which
   moved under either probe.
4. **`memory_test.go:247` does NOT move.** It caches
   `len(got) != 781` — the byte length of the **memory** golden. This spec does
   not touch memory rendering, and the assertion stayed green under M-2 and M-3.
   AGENTS.md §9 cites it as the canonical *not-reachable-by-any-grep* assertion,
   and that lesson still applies as a **shape to look for** — but carrying it
   forward as *"this file moves"* would be inheriting a wrong fact.

**One golden-adjacent surface that nothing guards.** `just test-docs` passed
**unchanged** under M-3, yet prose pins these headlines in **7 sites across 4
files** — unit stated because a first pass at this list got it wrong by
sweeping in `Entries:` lines belonging to *other* commands:

| Site | What it pins |
|---|---|
| `docs/api-contract.md:496` | `impact`'s contract prose — `Entries: <shown>/<in-window> with impact` |
| `docs/api-contract.md:575` | `wrapped`'s contract prose — a headline `Entries: N` count |
| `decisions/DEC-028:76` | the rule fixing `impact`'s two-number form |
| `decisions/DEC-028:145` | a worked example — `Entries: 4/7 with impact` |
| `decisions/DEC-030:140` | a worked `wrapped` example — `Entries: 7` |
| `decisions/DEC-048:72`, `:131` | cite `impact`'s form as *the sanctioned pair form*; a third number reopens that sentence |

**Explicitly NOT affected**, and listed so design does not sweep them in:
`docs/tutorial.md:325` (a `brag export` example), `docs/api-contract.md:670`
(`coverage`), `:746` (`spark`), `decisions/DEC-013:42` and `DEC-048:39/:67/:216`
(`export` and the general rule). DEC-048 states the five existing
`Entries:`-emitting exporters *"are correct and are not changed"* — this spec
must not disturb the four it does not touch.

**A headline change is doc-visible and harness-invisible**, so §9's premise
audit is the only thing that catches it. Design must enumerate these under
`## Outputs`.

---

## Complexity

**Re-estimated 2026-09-06. As framed, this is an L. Held at M by splitting.**

The framing prompt named three conditions that make it an L. Measured:

| Condition | Met? | Evidence |
|---|---|---|
| Fork B lands a change on more than the two originally-named surfaces | **YES** | `summary`'s `## Highlights` and `story`'s markdown both read wrong on the live corpus. Four surfaces need renderer work, not two. |
| Fork A requires fixing `--type` negation rather than routing it | **no** | M-1: a one-line in-memory predicate does the whole selection job. Storage is untouched. |
| The `Y3`/`X3` derivation is taken on here | **no** | Split to **SPEC-087**. See below. |

One condition fires, so the spec splits. **The split runs along the defect
shape the measurement established, not along an arbitrary line:**

- **SPEC-086 (this spec) — the *unrepresentable* pair: `impact` + `wrapped`.**
  A failure cannot be rendered as one in *either* format (DEC-028's 4-key
  projection). Fixing it means adding a section, deciding the accounting, and
  amending DEC-030's locked arc. This is the user's actual decision.
- **A successor — the *lossy-markdown* pair: `summary` + `story`.** The data is
  already present (`story` JSON carries `"type": "failed"`; `summary`'s
  `By type` prints `failed: 1`); only the per-entry markdown drops it. Different
  files, different fix, and `story`'s half additionally requires turning
  `Candor` from LLM-facing metadata into a body rule — a DEC-029-adjacent
  decision on its own.

**DEC-050, authored in this spec's design, states the posture for all seven
surfaces**, so STAGE-023's consistency guard is satisfied by the decision even
though the renderer work lands in two PRs. The successor is carried as an
unconditional entry on STAGE-023's backlog — the same carrier that produced
*this* spec when SPEC-085's Fork 4 fired, which is the mechanism this repo has
measured as working.

Splitting is the repo's normal outcome. SPEC-085 split Fork 4 out and that is
exactly why this spec exists.

---

## GO / NO-GO

**GO**, at complexity **M**, scoped to `impact` + `wrapped`. **No longer
blocked** — SPEC-085 shipped at `df369e9`; `brag learn` and the reserved `failed`
value exist, and the corpus holds a real `failed` row to design and test against.

**What changed the calculus since the first framing:** the interim risk that
justified deferring this work is **realised**. There is a failure entry in the
live corpus today, carrying an impact statement, rendering as a win on four
surfaces — one of which is a prompt instructing a model to promote it. The
argument for priority is no longer *"this will happen"*; it is *"this is
happening, and it compounds at roughly the rate the corpus grows."*

**What still argues for care rather than speed:** the empty-section fork (D) is
now known to be a live disagreement between the two surfaces rather than a
settled precedent, and DEC-030's locked section arc means `wrapped`'s half is an
amendment to a decision record. Neither is a reason to defer; both are reasons
design must pin choices with tests rather than inherit them.

---

## Re-measurement at design (2026-09-18, `main` at `3201f50`)

Design does not inherit framing's numbers. Each row was measured on the frozen
copy with a dev binary built from `3201f50`, which was then the tip of
`main`.

| Framing said (2026-09-06) | Measured 2026-09-18 |
|---|---|
| corpus **420**, max id 431 | **606**, max id **628** |
| **1** `type: failed` row (id 420) | **4**: 420 (`contextcore-pilot-harness`) and 433, 465, 473 (`bragfile`). All four carry an impact and all four are in 2026. |
| `impact` reads `Entries: 346/420 with impact` | `impact --year` reads **`Entries: 531/606 with impact`** |
| `wrapped` reads `Entries: 420` | **`Entries: 606`** |
| **19** distinct non-empty `type` values | **20**. Typeless entries rose from 113 to **136**. |
| `wrapped` `Top types`: `shipped 211`, `milestone 29`, `ship 15` | **`shipped 283`, `fixed 57`, `milestone 50`**. Still a top-3, and still no `failed`. |
| `--type` inclusion at `store.go:389` | Unchanged. `:388` is the guard and `:389` is `conds = append(conds, "e.type = ?")`. |
| the suite is **1070** tests | **1086** passing tests, of which **829** are top-level `func Test*`. |
| M-1 hashes `ddd33f3ae63a…` → `e7d171eecedadf…`, firing 0 tests | **Reproduced byte-exact** from framing's stated edit, and it **still fires 0**: `go test ./...` stays green. |
| `story`: `- ★ 420:` in all four profiles | All **four** failures render as `- ★ <id>:` in all four profiles. `grep -c -E '(473\|465\|433\|420): '` gives 4, 4, 4, 4. The JSON beats carry `"type": "failed"` for all four. |
| `summary`: `By type` shows `failed: 1`, and `## Highlights` lists it | `By type` shows `failed: 4`, and all four appear in `## Highlights` (`--range month`). The highlight JSON is still `["id","title"]`. |
| DEC-029: story's impact beats *"reuse `aggregate.WithImpact`"* | **Not true of the tree.** `internal/story/thread.go:135` computes `IsImpactBeat: e.Impact != ""` inline. Exactly two production files call `WithImpact`: `export/impact.go` (`:39`, `:103`) and `export/wrapped.go` (`:95`, `:225`). This is recorded in SPEC-094. |
| Fork C: *"there is no SQL counterpart"* | **There is one.** See **Fork C — settled**. |

**Measured on the frozen copy, with the prototype built from this spec's
literals:**

```
$ brag-proto --db <copy> impact --year | grep -n -E '^Entries:|^## '
6:Entries: 531/606 with impact            # byte-identical to main's binary
8:## Impact
1112:## What didn't work
$ brag-proto --db <copy> impact --year --format json | jq '{entries_with_impact,
      sum_counts: ([.counts_by_project[]]|add),
      impact_rows: ([.impact_by_project[].entries[]]|length),
      failure_rows: ([.failures_by_project[].entries[]]|length)}'
{ "entries_with_impact": 531, "sum_counts": 531, "impact_rows": 527, "failure_rows": 4 }
$ brag-proto --db <copy> wrapped | grep -n -E '^Entries:|^## '
6:Entries: 606
8:## Cadence  26:## Top initiatives  34:## Impact moments
1138:## What didn't work  1154:## Rhythm  1170:## Span
```

The markdown `## Impact` body of `impact --year` and `## Impact moments` of
`wrapped` are **byte-identical**, 1103 lines each. The two `## What didn't work`
sections are identical except for the blank line that separates sections:
`impact`'s is the last section, while `wrapped`'s is followed by `## Rhythm`.
The property framing measured for the impact rows (*"the same document,
rendered twice"*) now holds for both sections.

---

## Fork A — settled: zero renames, one count kept wide, two keys

**Re-verified on the tree: the obligation is zero renames.** Framing measured
*"at most one."* Both of framing's options assumed failures would be *removed
from* a count and then accounted for. Rather than doing that, design splits
the with-impact subset between two sections that are both shown. Every count
was run through DEC-048's rule:

| Count | Before | After | Meaning changed? |
|---|---|---|---|
| `impact` markdown `Entries: <shown>/<in-window> with impact` | `531/606` | `531/606` | **No.** `<shown>` is still the with-impact subset, and every row of it is still shown, in one section or the other. |
| `impact` JSON `entries_in_window` / `entries_with_impact` | 606 / 531 | 606 / 531 | **No.** |
| `impact` JSON `counts_by_project` | over the with-impact subset | **over the with-impact subset** | **Not if it is computed from `withImpact`.** Derived from the narrowed loop that builds `impact_by_project`, which is the obvious refactor, it silently becomes *"non-failure with-impact"*. That is exactly the kind of redefinition DEC-048 forbids. **LD3 pins it.** |
| `wrapped` `Entries: N` / `total_entries` | 606 | 606 | **No.** Sectioning never changed what the document covers. |
| `wrapped` `top_initiatives`, `longest_streak`, `top_tags`, `top_types`, `span`, `cadence` | — | unchanged | **No.** None of them reads the with-impact subset. `top_types` can now show `failed`. Its rule did not change, and that is honest. |

**So the DEC-048 work reduces to one sentence of code discipline,** that
`counts_by_project` is computed over both sections, plus a test that fails by
name if it is not.

**What was rejected, and why** (recorded in DEC-050's Alternatives):

- **A third number in the headline**, framing's option 1
  (`Entries: 531/606 with impact (4 failed)`). It reopens DEC-048's sentence
  naming the two-number form as *the sanctioned form for reporting a pair*,
  moves seven prose sites and nine tests across two packages (framing's M-3),
  and adds information the section already shows.
- **Per-section count lines**, framing's option 2. DEC-048's own
  Alternative 2 rules them out: it adds a provenance line to an envelope that
  never legislated one. The section arrays have lengths.

**The JSON wire change, enumerated.** Keys in order:

| Envelope | Before | After |
|---|---|---|
| `impact` | `generated_at`, `scope`, `filters`, `entries_in_window`, `entries_with_impact`, `counts_by_project`, `impact_by_project` | same, **then `failures_by_project`** |
| `wrapped` | `generated_at`, `scope`, `filters`, `total_entries`, `cadence`, `top_initiatives`, `impact_moments`, `longest_streak`, `top_tags`, `top_types`, `span` | same, **with `failures_by_project` directly after `impact_moments`** |

- **Breaking:** `impact_by_project` and `impact_moments` no longer carry a
  failure. A consumer that read every with-impact entry from either one
  misses the failures.
- **Additive:** `failures_by_project` has the same
  `[{project, entries:[{id, title, project, impact}]}]` shape and DEC-028
  choice 4's 4-key projection. It is always present, and is `[]` when empty.
  It uses one key name on both surfaces, so `jq .failures_by_project` works
  on both.

`main`'s `[Unreleased]` already carries one breaking rename, `brag memory`'s
`entries` becoming `candidates`. This spec adds a second breaking entry under
`### Changed`. It is embedded verbatim in *Notes for the Implementer*, and
listed in **## Outputs**.

---

## Fork C — settled: `internal/aggregate`, with a drift guard

**Home:** `internal/aggregate`, which is SQL-free and depends only on
`internal/storage`'s types. Three new exports are added: `FailureType`, which
moves from `internal/cli/learn.go:19`, plus `IsFailure(storage.Entry) bool`
and `SplitFailures([]storage.Entry) (others, failures []storage.Entry)`.
`brag learn` writes `aggregate.FailureType`, and the digests read
`aggregate.IsFailure`, so the writer and the readers name one constant.
`internal/cli` already imports `internal/aggregate` (`milestone.go`), so the
move adds no new edge between packages.

**Does it need a drift guard? Yes. Framing's reason for doubting it rested
on a false premise.** Framing wrote that the provenance case pairs a Go
predicate with a SQL clause *"because both exist,"* and that here *"there is
no SQL counterpart."* But DEC-049 defines retrieval as `brag list --type
failed`, and that goes through `store.go:389`, `e.type = ?`, on a column
declared `type TEXT` with no `COLLATE` (`0001_initial.sql:7`). That makes it
a BINARY, case-sensitive, untrimmed comparison. So a failure *does* have two
definitions in two packages, one SQL and one Go. That is exactly the
`IsAgentAuthored` shape. The only difference is that the SQL side is generic
rather than failure-specific, which is why DEC-049 could keep storage
ignorant of `failed`.

**What the guard catches that nothing else does:** a well-meaning change that
widens the Go side. Examples are `strings.EqualFold` to catch `Failed`,
`TrimSpace` to catch `" failed"`, or matching `failure` too, since DEC-049's
Consequences admit an agent can write `brag add --type failure`. Any of these
makes `brag impact` list rows that `brag list --type failed` cannot return.
The guard also catches the opposite drift, a SQL side that starts
case-folding.

**What it does not need to guard:** the writer against the reader. After the
move there is one constant, and a test that compares a constant to itself
passes either way. That is DEC-048's point about byte-identity tests. Two
different tests pin that path. `TestLearnCmd_PinsFailedType` holds the
literal `"failed"`, because the value is persisted and renaming the constant
would orphan every stored row while writer and reader still agreed. The two
`TestLearnCmd_*SectionsWhatItWrote` tests run the real verb and the real
digest against one store, with no constant anywhere in them.

**Evidence the guard has teeth,** by mutation (full matrix below):

- **M-A** makes the Go side `EqualFold`. The guard fires.
- **M-B** makes the Go side `TrimSpace`. The guard fires.
- **M-S** mutates the **SQL** side to `"e.type = ? COLLATE NOCASE"`. The guard
  fires with **`failure sets differ: SQL=3 Go=2`** while its anchored count
  stays satisfied, which proves it compares the two sides rather than only
  counting.

---

## Fork D — settled: the section renders only when it has an entry

**The rule is the same on both surfaces:** `## What didn't work` appears in
markdown only when at least one with-impact failure is in scope. JSON always
carries `failures_by_project`, as `[]` when empty (DEC-014 part 4).

**Why this is not either surface's existing convention.** Framing measured the
two surfaces disagreeing on an empty *section*. `wrapped` renders a bare
`## Impact moments`, because its DEC-030 arc is fixed, while `impact` omits
its whole body. Inheriting each surface's own convention would give the new
section two behaviours for one heading. Following `wrapped`'s convention
would give every clean quarter a permanent empty *"What didn't work"*
heading, the accusatory outcome framing named. So the new section gets one
rule of its own.

**What it does NOT change:** `wrapped`'s five DEC-030 sections keep their
behaviour, including the bare `## Impact moments` on a period with no impact
at all. Changing that is a different spec and would move three markdown
goldens. The DEC-030 Amendment records the asymmetry and argues it: five
celebratory sections that always have something to say, and one honesty
section that only speaks when there is something to be honest about.

**One consequence, pinned rather than left to build:** on `impact`, a window
whose with-impact rows are *all* failures (for example
`brag impact --year --type failed`) renders **no `## Impact` heading**, only
`## What didn't work`. That extends `impact`'s existing rule (*omit the body
when it is empty*) from the body to each section. Probed on a scratch DB with
one `brag learn` entry:

```
Entries: 1/1 with impact

## What didn't work

### alpha

- 1: tried a pool
  cost two days
```

Every Fork D probe ran on a scratch DB, not on the corpus:

| Scratch DB | `impact --year` | `wrapped` headings | JSON |
|---|---|---|---|
| one win with an impact | `## Impact` only | the five DEC-030 sections | `failures_by_project: []` on both |
| one failure with an impact | `## What didn't work` only | `## Impact moments` (bare) → `## What didn't work` → `## Rhythm` | one group on both |
| one entry, no impact | provenance only: `Entries: 0/1 with impact` | the five DEC-030 sections, `## Impact moments` bare | `[]` on both |

**Rejected, with the reason recorded in DEC-050:** always rendering the
heading, and rendering it with a *"None recorded."* line. The line would
also claim more than the section knows, because an impact-less failure is not
listed.

---

## Fork E — settled: goldens found by execution

The prototype carried every production and test change in this spec. The two
kinds of golden that could move were found by **running the prototype's
production code against `main`'s unmodified tests**, not by grep:

```
--- FAIL: TestToImpactJSON_DEC028ShapeGolden     (internal/export)
--- FAIL: TestToWrappedJSON_DEC030ShapeGolden    (internal/export)
FAIL	github.com/jysf/bragfile000/internal/export
(every other package: ok)
```

| Golden | Why it moves | `want` lines, before → after |
|---|---|---:|
| `internal/export/impact_test.go` `TestToImpactJSON_DEC028ShapeGolden` | It enumerates every key byte-for-byte, and `failures_by_project: []` is new, even though the fixture holds no failure. | 41 → **42** |
| `internal/export/wrapped_test.go` `TestToWrappedJSON_DEC030ShapeGolden` | Same reason, with the key inserted after `impact_moments`. | 145 → **146** |

**What does not move, and why:**

- **No markdown golden moves.** The four existing ones
  (`TestToImpactMarkdown_DEC014FullDocumentGolden`,
  `TestToWrappedMarkdown_DEC014FullDocumentGolden`,
  `TestToWrappedMarkdown_QuarterGolden` and the empty-shape pair) all use
  fixtures with no failure, and Fork D omits the empty section. They become
  the regression half of Fork D for free. M-E and M-F show three of them
  fire the moment the heading renders unconditionally.
- **Nothing in `internal/cli` moves.** Framing's M-3 reached
  `internal/cli/wrapped_test.go` through a **headline rename**. Fork A
  renames nothing.
- **`internal/export/memory_test.go:247` does not move.** The full suite was
  green with only the two edits above.
- **Two tests change additively, not because they broke:**
  `TestToImpact_EmptyWindowShape` and `TestToWrapped_EmptyPeriodShape` each
  gain one assertion that the new key is `[]`, so DEC-014 part 4's empty
  shape is pinned for the new key too.

The **new** goldens are listed under **## Failing Tests**: 32, 66 and 60
`want` lines.

---

## §12(b) design-time pre-flight: what the tools actually said

Every literal in *Notes for the Implementer* was run through its real tool:
the Go code through `go build`, `gofmt`, `go vet`, `just lint` and the full
suite; the help strings through cobra's `--help`; the markdown and the JSON
through the prototype binary on the frozen copy; the doc hunks and Group `AD`
through `just test-docs`; and the inventory through `scripts/inventory.sh`.
Seven findings:

**Finding 1: a helper not named `assert_*` produces ids that run and are
never counted.** Group `AD`'s first draft called its helper `ad_names`. The
five ids ran and printed `OK:`, but `Documentation assertions` did **not**
move off 200. `inventory.sh:67` counts only ids passed to `ok`, `fail`,
`skip` or an `assert_[a-z_]+` call. Renamed to `assert_section_names`, the row
moved 200 → **205**. This is the same class as SPEC-085's Finding 2 (a
loop-built id), reached through a different door. **LD11 pins the name.**

**Finding 2: `docs/engineering-practices.md:230` stated a current-state
count in prose, and this design falsifies it.** The sentence was *"as the
inventory shows, only one decision record carries an explicit `## Amendment`
section."* DEC-030's Amendment makes that two in the design commit itself,
and DEC-028's (R3, 2026-09-19) makes it three.
The page's own rule is that a current-state number belongs in the inventory
table, not in prose. So the sentence is rewritten **without** a number, in
this design commit, where the amendment lands. No assertion caught it:
`X6` checks that the section cites records by id, and says nothing about the
sentence's count.

**Finding 3: the obvious `counts_by_project` refactor is a silent DEC-048
violation.** Folding the count loop into the new `impactGroups(worked)`
helper, the natural build, makes alpha 2 instead of 3 on the fixture and
drops gamma entirely. All existing tests stay green, because none has a
failure. That is mutation M-D, and **only the new tests catch it.**

**Finding 4: framing's M-1 is sufficient and still wrong.** Re-run on the
final tree, narrowing `WithImpact` now fires 9 tests. On `main` it fired 0.
Beyond selecting the rows, it turns the headline false: `527/606 with
impact`, while four entries that do carry an impact sit in the next section.
It also makes a function named `WithImpact` return something else. **LD2 pins
the partition.**

**Finding 5: an M-M probe failed to compile, and was discarded rather than
credited.** Its first form replaced `worked` with `aggregate.WithImpact(entries)`,
and the package failed with `declared and not used: worked`. That is a red
for the wrong reason, and in a log it looks just like a real one (§12's third
clause). The probe was re-run as a mutant that compiles cleanly.

**Finding 6: the help strings render through cobra as literal single
paragraphs.** `impact --help` line 9 and `wrapped --help` line 12 carry the
new sentences verbatim. **The NOT-contains self-audit found nothing to fix.**
Every existing help assertion on either command is positive
(`impact_test.go:638/:641`, `wrapped_test.go:179/:404/:407/:506/:509`), and
this spec adds no NOT-contains on help text. The one NOT-contains this spec
does add, *"`## Impact` does not carry the `learn` entry"*, is paired with a
positive in the same test, and its needle (`- <id>:`) cannot occur in prose.

**Finding 7: fail-first was run, not predicted.** The new and modified test
files were copied onto an **unmodified** `main` worktree. `learn_test.go` kept
its old `FailureType` references, which is the state build is in when it
writes tests first. Result:

```
internal/aggregate [build failed]   undefined: IsFailure / FailureType / SplitFailures
internal/storage   [build failed]   undefined: aggregate.FailureType / aggregate.SplitFailures
internal/export    FAIL  10 tests: 6 of the 7 new, the 2 JSON shape goldens, the 2 empty shapes
internal/cli       FAIL  4 tests:  both learn e2e tests, both help tests
```

Every export and cli failure is an assertion failure, which is the reason each
test states. The two build failures are the new symbols not existing yet.
**The seventh new export test, `TestToImpactJSON_CountsByProjectSpansBothSections`,
passes on `main`, and should.** With no failure section, every row is in
`impact_by_project`, so the invariant holds trivially. It is a guard against
the wrong implementation, M-D, and not a test that fails first. Build should
expect exactly this picture.

### Mutation matrix: 19 probes, each confirmed by content hash before its gate ran

The probe helper **refused to run the gate until the hash had moved**
(§12(b) clause (1), refined). It printed the diff it applied, restored from a
`/tmp` backup, and confirmed the hash returned. The *Diff* column is what the
helper printed, not a description written afterwards (the held
codification candidate on STAGE-023). Baselines, measured on the prototype's
final files: `aggregate.go` `a6001a8d00d6`, `export/impact.go`
`f00b77c8e1c1`, `export/wrapped.go` `e507f9ce65a8`, `cli/learn.go`
`545d63eb1ef2`, `cli/impact.go` `71c2dae1fc23`, `cli/wrapped.go`
`5e7887dc464c`, `storage/store.go` `8091ff9d39bd`, `docs/api-contract.md`
`bde92cdba356`, `docs/tutorial.md` `e5aa062fa66c` and `AGENTS.md`
`acc844937b43`. Every target was back at its baseline at the end
(`diff` against the recorded baselines: empty).

| # | File | Diff (the one line replaced → its replacement) | Hash | Tests that fired |
|---|---|---|---|---|
| **M-1** | `aggregate.go` | `if e.Impact != "" {` → `if e.Impact != "" && e.Type != "failed" {` | `a6001a8d00d6`→`2642076c19d3` | **9**: both `learn` e2e, `CountsByProjectSpansBothSections`, both impact failure goldens, `SectionsRenderOnlyWhenNonEmpty`, `FailuresLeaveImpactMoments`, wrapped `FailureSectionGolden`, `WhatDidntWorkRendersOnlyWhenNonEmpty` |
| **M-A** | `aggregate.go` | `return e.Type == FailureType` → `return strings.EqualFold(e.Type, FailureType)` | →`90858b3da15f` | `TestFailureClassifier_…`, `TestIsFailure_…`, `TestSplitFailures_…` |
| **M-B** | `aggregate.go` | `return e.Type == FailureType` → `return strings.TrimSpace(e.Type) == FailureType` | →`b7c212dbc136` | `TestFailureClassifier_…`, `TestIsFailure_…` |
| **M-C** | `aggregate.go` | `const FailureType = "failed"` → `const FailureType = "failure"` | →`725f414acc1c` | **10** in 4 packages: the 8 tests whose expected output needs a literal-`"failed"` row to be a failure, `TestLearnCmd_PinsFailedType`, and the guard. `…CountsByProjectSpansBothSections` stays green, because its invariant also holds with every row in one section. **Both e2e tests stay green**, as designed: writer and reader still agree, and the literal pins catch it. |
| **M-S** | `storage/store.go` | `conds = append(conds, "e.type = ?")` → `conds = append(conds, "e.type = ? COLLATE NOCASE")` | `8091ff9d39bd`→`5e0b2f5c607b` | `TestFailureClassifier_…` (**`failure sets differ: SQL=3 Go=2`**) and the existing `TestList_FilterByType` |
| **M-D** | `export/impact.go` | `for _, group := range aggregate.GroupEntriesByProject(withImpact) {` → `…(worked) {` | `f00b77c8e1c1`→`1985bca9562b` | `TestToImpactJSON_CountsByProjectSpansBothSections`, `TestToImpactJSON_FailureSectionGolden`. **No pre-existing test.** |
| **M-E** | `export/impact.go` | `if len(failed) > 0 {` → `if true {` | →`4f46c0b9d85a` | `TestToImpact_EmptyWindowShape`, `TestToImpactMarkdown_DEC014FullDocumentGolden`, `TestToImpactMarkdown_SectionsRenderOnlyWhenNonEmpty` |
| **M-F** | `export/wrapped.go` | `if len(failed) > 0 {` → `if true {` | `e507f9ce65a8`→`12329eaab606` | `…DEC014FullDocumentGolden`, `…NoSparkOmitsGlyphLine`, `…QuarterGolden`, `…WhatDidntWorkRendersOnlyWhenNonEmpty` |
| **M-G** | `export/impact.go` | `if len(worked) > 0 {` → `if len(withImpact) > 0 {` | →`f7732553c44f` | **only** `TestToImpactMarkdown_SectionsRenderOnlyWhenNonEmpty` (the failures-only case) |
| **M-H** | `cli/learn.go` | flag mode: `Type:        aggregate.FailureType,` → `Type:        "failure",` | `545d63eb1ef2`→`0f91bd325fd6` | `TestLearnCmd_PinsFailedType`, both `learn` e2e tests |
| **M-J** | `export/wrapped.go` | `fmt.Fprintln(&buf, "## What didn't work")` → `…"## What did not work")` | →`b95cb4914bc2` | wrapped `learn` e2e, `…FailureSectionGolden`, `…WhatDidntWorkRendersOnlyWhenNonEmpty` |
| **M-K** | `export/wrapped.go` | `` `json:"failures_by_project"` `` → `` `json:"what_didnt_work"` `` | →`045433f352e9` | `…EmptyPeriodShape`, `…DEC030ShapeGolden`, `…FailuresLeaveImpactMoments` |
| **M-M** | `export/wrapped.go` | `writeImpactGroups(&buf, worked)` → `writeImpactGroups(&buf, append(append([]storage.Entry{}, worked...), failed...))` | →`6f2af25865c0` | wrapped `learn` e2e, wrapped `…FailureSectionGolden` |
| ~~M-M₀~~ | `export/wrapped.go` | `…worked)` → `…aggregate.WithImpact(entries))` | moved | **Discarded**: `declared and not used: worked`, a compile red (Finding 5) |
| **M-L1** | `cli/impact.go` | `…and that heading is left out when there is none.` → `…and that heading is always shown.` | `71c2dae1fc23`→`9f6ec4206bb6` | `TestImpactCmd_HelpNamesTheFailureSection` |
| **M-L2** | `cli/wrapped.go` | `Impact moments, What didn't work (work recorded with brag learn, left out when there is none), Rhythm` → `Impact moments, Rhythm` | `5e7887dc464c`→`7352ab9bf67b` | `TestWrappedCmd_HelpListsTheFailureSectionInArcOrder` |
| **M-D1** | `docs/api-contract.md` | in the **impact** section only, both `` `failures_by_project` `` → `` `failures` `` | `bde92cdba356`→`520445c282a9` | **`AD1` only.** `AD2` stays green, which proves the needle is scoped to its section. |
| **M-D2** | `docs/tutorial.md` | delete `**What didn't work** (anything you recorded with `brag learn`, shown only when there is some), ` | `e5aa062fa66c`→`61767048bed1` | `AD4` |
| **M-D3** | `AGENTS.md` | the arc reordered to `… → Rhythm (…) → What didn't work (…) → Span` | `acc844937b43`→`c0c6bd7a7d58` | `AD5` |
| **M-D4** | `docs/api-contract.md` | `` ### `brag impact --quarter… `` → `` ### `brag  impact --quarter… `` (two spaces) | `bde92cdba356`→`eabca4338391` | `AD1: … section not found`, the non-vacuity branch |

**M-D and M-G are the two that matter most.** Each is caught **only** by a
test this spec adds. M-D is the DEC-048 regression Fork A exists to prevent,
and M-G is Fork D's failures-only case. Without the new tests, both would
survive a green suite, exactly as M-1 did on `main`.

### Inventory: regenerated and diffed, not predicted

Two generations of `scripts/inventory.sh` were diffed: `main` at `3201f50`
against the prototype carrying every artifact in this spec.

| Row | `main` | this design commit | after build |
|---|---:|---:|---:|
| Decision records | 51 | **52** | 52 |
| …of those, carrying an explicit `## Amendment` section | 1 | **3** | 3 |
| Go test files | 79 | 79 | **80** |
| Go test functions | 829 | 829 | **843** |
| Documentation assertions (distinct ids) | 200 | 200 | **205** |

Every other row is unchanged. `Specs carried to ship and archived` stays at 85
and `Stages` at 23, because SPEC-094 is a live spec, not an archived one.
`Questions` stays at 21 / 8, because no question was filed, so `Y4`'s pins do
not move.

### §9(b): the harness grepped by VALUE

For each value that moves, `scripts/test-docs.sh` was grepped for the literal,
with comment lines excluded:

```
$ for v in '| 51 |' '| 52 |' '| 829 |' '| 843 |' '| 79 |' '| 80 |' '| 200 |' '| 205 |' \
           'Amendment` section | 1' '!=51' '!=200' '!=829'; do
    /usr/bin/grep -n -F -- "$v" scripts/test-docs.sh | /usr/bin/grep -v '^[0-9]*:[[:space:]]*#'
  done
(no hits for any value)

$ # re-run 2026-09-19, for the two values R3 moved the `## Amendment` row through
$ for v in '| 2 |' '| 3 |' 'Amendment` section | 2' 'Amendment` section | 3'; do
    /usr/bin/grep -n -F -- "$v" scripts/test-docs.sh | /usr/bin/grep -v '^[0-9]*:[[:space:]]*#'
  done
(no hits for any value)
```

**No literal pin exists for any of them.** `Y3` derives (SPEC-087) and `Z7`
reads what `inventory.sh` emits. Only `X3`'s page block moves, and it is
regenerated. **The design commit re-pins nothing by hand.** Like every
decision record since SPEC-087 made `Y3` derive (DEC-051, DEC-052 and
DEC-053), DEC-050 costs zero `Y3` edits. The `## Amendment` row moves with no
pin either.

---

## Locked design decisions

**LD1 — the failure predicate lives in `internal/aggregate`, and there is one
constant.** `aggregate.FailureType = "failed"` moves from `internal/cli`.
`cli.FailureType` is deleted, not aliased, because two names for one value
is the fragmentation DEC-049 exists to stop. `aggregate.IsFailure` is exact
equality. `brag learn` writes `aggregate.FailureType` in both modes. Fork C.

**LD2 — the section is a partition of the with-impact subset, not a filter.**
Both renderers compute `WithImpact(entries)` and then
`SplitFailures(withImpact)`. The rest render in the impact section, and the
failures render in `## What didn't work`. `aggregate.WithImpact` is **not
modified** (Finding 4). Both sections render through one shared writer,
`writeImpactGroups`, so an entry reads byte-identically in either section and
on either surface. A failure **without** an impact is in neither section: it
is counted and not listed, like any impact-less entry (DEC-028 choice 3).

**LD3 — Fork A: no count changes what it counts, and nothing is renamed.**
Every headline string and count key is unchanged. `impact`'s
`counts_by_project` is computed over `withImpact`, both sections, so it still
sums to `entries_with_impact`, and each project's count equals its rows across
both sections. No per-section count line is added.

**LD4 — the JSON envelopes.** Both gain `failures_by_project`, with the same
group shape and the same 4-key entry projection, always present and `[]` when
empty. On `impact` it is the last key. On `wrapped` it comes directly after
`impact_moments`. `impact_by_project` and `impact_moments` no longer carry a
failure. This is a breaking change, and it is named in the CHANGELOG.

**LD5 — Fork D: `## What didn't work` renders only when it has an entry, on
both surfaces, and `impact`'s `## Impact` follows the same rule.** `wrapped`'s
five DEC-030 sections keep their existing behaviour, including a bare
`## Impact moments`.

**LD6 — DEC-030 and DEC-028 are amended, not superseded.** DEC-030's
`## Amendment (2026-09-18, SPEC-086 design)` places the section between Impact
moments and Rhythm and records the conditional asymmetry. DEC-028's
`## Amendment (2026-09-19, SPEC-086 design)`, added on the R3 ruling, records
the two-key split, the always-present `failures_by_project`, the unnarrowed
`counts_by_project` and the unchanged headline. **On both records the original
text is left as written**, and each amendment is read against it.

**LD7 — the drift guard is
`TestFailureClassifier_GoPredicateMatchesTypeFilter`, in
`internal/storage/failure_agreement_test.go`** (external package
`storage_test`, beside the provenance guard, reusing its `idSet` and
`sameIDSet`). It checks set agreement between `ListFilter{Type:
aggregate.FailureType}` and `SplitFailures(List({}))` over eight seeds, six of
them near-misses, and anchors the count at 2.

**LD8 — the export fixtures type their failures with the literal `"failed"`,
never `aggregate.FailureType`.** That way they also pin the persisted value
(M-C fires them), and they compile on `main`, so fail-first fails on
assertions rather than symbols. The CLI e2e tests use **neither**. They run
`brag learn` and read the digest, so the path from writer to reader is tested
with no shared literal.

**LD9 — the two `--help` sentences are literals.** For `impact`, appended to
the body paragraph: *Work recorded with brag learn is listed under its own
"What didn't work" heading instead of among the impact, and that heading is
left out when there is none.* For `wrapped`, in the section list:
*…Impact moments, What didn't work (work recorded with brag learn, left out
when there is none), Rhythm…*. Both `Short` strings are unchanged.

**LD10 — the doc sweep is the premise audit's EDIT list, exactly.** It covers
both `api-contract.md` sections, both tutorial paragraphs, the two AGENTS.md
glossary entries, and two CHANGELOG entries. The `engineering-practices.md`
sentence is fixed **at design** (Finding 2).

**LD11 — Group `AD` has five literal ids, is inserted above
`# ===== finalise =====`, and its helper is named `assert_section_names`.**
The placement follows SPEC-085 LD8: an appended group never runs. The name
follows Finding 1: only an `assert_*` helper's ids are counted. Each needle
is scoped to its command's section, which M-D1 proves.

**LD12 — every negative assertion is paired with a positive** that fails if
the mechanism under test is absent:

| Negative | Its pair, in the same test |
|---|---|
| a clean window has no `## What didn't work` line | the failure fixture **has** it, in the right position (both surfaces) |
| a failures-only window has no `## Impact` line | it **has** `## What didn't work` |
| `## Impact` / `## Impact moments` does not carry the `learn` entry | `## What didn't work` **does**, with its impact line, and the section above carries the `add` entry |
| `impact_moments` has no failure row | `failures_by_project` has exactly the two failures with an impact |
| `IsFailure` rejects six near-misses | it accepts `"failed"` |

**LD13 — scope.** DEC-050 states the posture for all seven surfaces. This
spec implements rows 1 and 2 only, and SPEC-094 implements rows 3 and 4.
No file under `internal/story` or `internal/export/summary.go` is modified.

**LD14 — the inventory is regenerated at build by pasting
`./scripts/inventory.sh` output between the markers.** Three rows move:
Go test files, Go test functions and Documentation assertions. None is ever
hand-edited, and `X3` diffs the whole block.

### Rejected alternatives (build-time)

- **Folding the `counts_by_project` loop into `impactGroups(worked)`.** It
  is the shortest refactor, and a silent DEC-048 violation (Finding 3, M-D).
- **Keeping `cli.FailureType` as an alias of `aggregate.FailureType`.** It
  gives one value two names, which is the drift LD1 removes.
- **A `type` key on the 4-key projection** instead of a new array. Every
  consumer that reads `impact_by_project` without filtering would still list
  the failure as a win.
- **Putting the drift guard in `internal/aggregate`.** `aggregate`'s tests
  cannot open a store without importing SQL machinery, and an in-package
  `storage` test importing `aggregate` is an import cycle. The external
  `storage_test` package is where the precedent lives, for that reason.
- **Using `aggregate.FailureType` in the export fixtures.** Fail-first would
  become a compile error instead of assertion failures, and M-C would not
  fire a single golden (LD8).
- **Naming the `wrapped` key `what_didnt_work`,** after its section. DEC-050
  rejects it: one datum would have two names across two surfaces (M-K fires
  on it).
- **Appending Group `AD` at the end of the file, or naming its helper
  `ad_*`.** Each fails green in its own way (LD11).
- **Adding `## What didn't work` to `brag learn --help`.** It is not a status
  claim that goes false. `learn --help` names `brag list --type failed` as
  the read-back, and that stays true. Extending it would widen this spec's
  footprint on a file SPEC-091 also edits.

---

## Outputs

### New files (1)

| Path | Lines | What |
|---|---:|---|
| `internal/storage/failure_agreement_test.go` | 70 | The Fork C drift guard (LD7). Embedded verbatim in §7. |

### Modified files, at build (18)

| Path | Change | Δ, measured on the prototype |
|---|---|---:|
| `internal/aggregate/aggregate.go` | `FailureType`, `IsFailure`, `SplitFailures` | +39 |
| `internal/aggregate/aggregate_test.go` | 2 new tests | +74 |
| `internal/cli/learn.go` | const removed, and `aggregate.FailureType` used in both modes | +6 / −10 |
| `internal/cli/learn_test.go` | **planned rewrite:** 7 `FailureType` refs become `aggregate.FailureType`, plus 2 new e2e tests and 2 helpers | +143 / −6 |
| `internal/cli/impact.go` | the `Long` sentence (LD9) | ±1 |
| `internal/cli/impact_test.go` | 1 new help test | +15 |
| `internal/cli/wrapped.go` | the `Long` sentence (LD9) | ±1 |
| `internal/cli/wrapped_test.go` | 1 new help test | +15 |
| `internal/export/impact.go` | partition, the shared writer, `impactGroups`, and the new key | +57 / −26 |
| `internal/export/impact_test.go` | 4 new tests; **planned rewrites** of `TestToImpactJSON_DEC028ShapeGolden` (+1 line) and `TestToImpact_EmptyWindowShape` (+1 assertion) | +250 / −1 |
| `internal/export/wrapped.go` | partition, the shared writer, `wrappedGroups`, and the new key | +60 / −45 |
| `internal/export/wrapped_test.go` | 3 new tests; **planned rewrites** of `TestToWrappedJSON_DEC030ShapeGolden` (+1 line) and `TestToWrapped_EmptyPeriodShape` (+1 assertion) | +183 |
| `docs/api-contract.md` | both sections (LD10) | +26 / −6 |
| `docs/tutorial.md` | both paragraphs | +7 / −3 |
| `AGENTS.md` | §11: the `wrapped` arc and the `learn` entry | +2 / −2 |
| `CHANGELOG.md` | `[Unreleased]` → `### Changed`, two entries | +20 |
| `scripts/test-docs.sh` | Group `AD`, 5 ids, above `finalise` | +71 |
| `docs/engineering-practices.md` | the regenerated inventory block **only** (3 rows) | ±3 |

The regenerated block is listed but is not hand-written. **Build total,
measured on the prototype:** 1 new file (70 lines) and 18 modified, **+973 /
−104** (numstat, counting `engineering-practices.md` as its 3 build-time
inventory rows; the design commit's own sentence fix and 2 rows are excluded).

### Modified or created at design (this commit)

| Path | Change |
|---|---|
| `decisions/DEC-050-a-failure-is-never-rendered-as-a-win.md` | **new**, **302** lines: the posture for all seven surfaces. 277 as first written; R1 rewrote row 4 and both confidence statements, R2 added the acceptance and `T5`, R3 rewrote its DEC-028 reference |
| `decisions/DEC-030-…section-taxonomy.md` | `## Amendment (2026-09-18, SPEC-086 design)`, +38 lines |
| `decisions/DEC-028-…window-and-shape.md` | **R3**: `## Amendment (2026-09-19, SPEC-086 design)`, +40 lines, appended before `## References`. **Nothing above the heading is edited**, which is how DEC-030 is treated (LD6) |
| `projects/PROJ-008-…/specs/SPEC-094-summary-and-story-stop-listing-a-failure-as-a-win.md` | **new**, **142** lines, via `just new-spec`: the id claimed, `cycle: frame`. 134 as first written; R1 rewrote its row-4 input, the decision it still owes, and the `T3` hand-back. **Nothing else of its framing is designed here** |
| `projects/PROJ-008-…/stages/STAGE-023-…md` | the SPEC-086 entry now reads `(design)`, the *(not yet written)* entry now reads **SPEC-094**, and the count is re-derived |
| `docs/engineering-practices.md` | the regenerated inventory block (still 2 rows: `Decision records` 51 → 52 and the `## Amendment` row, which R3 takes to **3** rather than 2), and the `:230` sentence (Finding 2) |
| this file | `cycle: design` (the recipe's stripped comment restored), the design sections, and *Maintainer rulings (2026-09-19)* |

### The CHANGELOG entries

Under `## [Unreleased]` → `### Changed`, **before** the existing `brag memory`
entry. Embedded verbatim in §6.

- **`brag impact` and `brag wrapped` list work that did not work under its own
  `## What didn't work` heading.** A failure that carries an impact moves out
  of `## Impact` and `## Impact moments`. The section appears only when there
  is one, and no headline count changes.
- **Breaking: `impact_by_project` and `impact_moments` no longer list
  failures,** and both envelopes gain `failures_by_project`.

### Premise audit (§9), run at design against the repo

**Inversion: tests that assume every with-impact entry sits in the impact
section.**

```
$ /usr/bin/grep -rln --include='*_test.go' -E 'WithImpact|ToImpact(Markdown|JSON)|ToWrapped(Markdown|JSON)|impact_by_project|impact_moments|runImpactCmd|runWrappedCmd|FailureType' internal cmd
internal/aggregate/aggregate_test.go   internal/cli/wrapped_test.go
internal/cli/impact_test.go            internal/cli/learn_test.go
internal/export/wrapped_test.go        internal/export/impact_test.go
(6 files, 97 matching lines)
```

**By execution, not by reading, none of them holds the premise with a
failure in its fixture.** Framing's M-1 survived the whole suite, and nothing
fed a `failed` row through `WithImpact`, `impact` or `wrapped`. So **no test is
deleted.** What moves:

| Test | Why | Plan |
|---|---|---|
| `TestToImpactJSON_DEC028ShapeGolden` | byte-exact over every key | **rewrite**: +1 line, `"failures_by_project": []` |
| `TestToWrappedJSON_DEC030ShapeGolden` | same | **rewrite**: +1 line |
| `TestToImpact_EmptyWindowShape` | still passes, but does not pin the new key's empty form | **additive**: +1 assertion |
| `TestToWrapped_EmptyPeriodShape` | same | **additive**: +1 `assertRaw` |
| `learn_test.go`, 7 `FailureType` refs | the constant moves (LD1) | **rewrite** to `aggregate.FailureType`, plus the import |
| `TestWithImpact_*` (2) | `WithImpact` is unchanged | none |

**Addition: tracked collections this spec adds to.** One DEC moves
`Decision records` 51 → 52. **Two** Amendments move its row 1 → 3 — DEC-030's
and, on the R3 ruling, DEC-028's. All of that happens in the design commit. At build, 14 test functions move `Go test functions`
829 → 843, one new file moves `Go test files` 79 → 80, and five ids move
`Documentation assertions` 200 → 205. All five were regenerated rather than
predicted, and none is hand-pinned (§9(b) above).

**Status change: every description of `impact`'s or `wrapped`'s output
shape.** The greps, run with `git ls-files -z | xargs -0 /usr/bin/grep -n -F`
(tracked files only), excluding `specs/done/`, test files and this spec's own
framing record:

```
needles: 'Impact moments'  'impact_moments'  'impact_by_project'  'Top initiatives'
         'Rhythm'  'entries_with_impact'  'with impact'  '## Impact'  'Impact body'
plus: every subcommand's --help, swept with
  for c in $(brag --help | awk '/^Available Commands:/{f=1;next} /^$/{f=0} f{print $1}'); do
    brag $c --help | grep -n -i -E 'impact moments|impact_by_project|impact_moments|## impact|with impact|brag wrapped|brag impact|highlight|celebrat'
  done
```

| Hit | Verdict |
|---|---|
| `internal/cli/impact.go:35` (`--help` body paragraph) | **EDIT**: LD9 |
| `internal/cli/wrapped.go:45` (`--help` section list) | **EDIT**: LD9 |
| `docs/api-contract.md:508-512` (impact body), `:546-552` (impact JSON keys), `:560-562` (impact empty window) | **EDIT** |
| `docs/api-contract.md:597-599` (wrapped `## Impact moments`), `:636-637` (wrapped JSON keys) | **EDIT** |
| `docs/tutorial.md:536` (impact), `:568` (wrapped) | **EDIT** |
| `AGENTS.md:299` (`wrapped` glossary arc) | **EDIT** |
| `AGENTS.md:294` (`learn` glossary) | **EDIT**: it names where a failure now renders, and the moved constant |
| `CHANGELOG.md` `[Unreleased]` | **EDIT**: two entries |
| `internal/export/impact.go:30-37`, `wrapped.go:40-46` (doc comments) | **EDIT**, in the code diffs |
| `docs/engineering-practices.md:230` (*"only one decision record carries an explicit `## Amendment`"*) | **EDIT at design**: Finding 2 |
| `decisions/DEC-028-…:76, :100, :125, :145, :303` | ~~**NO CHANGE.** DEC-050 refines them and names what it refines. This follows DEC-048's precedent of not editing the record it binds (DEC-014 was left readable). This is an open question for the user.~~ **OVERTURNED by the maintainer, 2026-09-19 (R3): EDIT at design.** Refine-over-amend leaves DEC-028 stating a key list and a one-section body that go stale the moment failures move out, and DEC-048's precedent does not reach it — DEC-014 was left readable because nothing in it went false, whereas `:100` and `:125` do. The record gains `## Amendment (2026-09-19, SPEC-086 design)`; **those five lines of original text are still not edited** (LD6), which is how DEC-030 is treated. |
| `decisions/DEC-030-…:80, :90, :119, :153, :199, :260` | **NO CHANGE** to the original text. The Amendment carries the change (LD6). |
| `decisions/DEC-029-…:114, :254` | **NO CHANGE**: they concern `story`'s shape |
| `decisions/DEC-048-…:72, :131` | **NO CHANGE**: the pair form is untouched, which is the point of Fork A |
| `README.md:216-217`, `BRAG.md:441` | **NO CHANGE**: examples with no shape claim |
| `BRAG.md:153` (*"The `impact` field is still worth filling in"*) | **NO CHANGE**: still true, and it is now also what routes a failure into the section |
| `projects/PROJ-008-…/brief.md:59` | **NO CHANGE**: a dated framing measurement. Per-section byte-identity still holds (measured above). |
| `internal/cli/learn.go:53-55` (`learn --help` read-back) | **NO CHANGE**: `brag list --type failed` stays true |
| the root `Short` lines for `impact` and `wrapped` | **NO CHANGE**: no shape claim |
| `summary --help`, `story --help` | **NO CHANGE here**: they belong to SPEC-094 |
| `scripts/test-docs.sh:1508` (*"only 1 of 47 DECs"*), `guidance/questions.yaml:883` (*"exactly 1 of 45"*) | **NO CHANGE**: each is dated by its own denominator |
| `docs/data-model.md` DEC list | **NO CHANGE**: it stops at DEC-021 and is not maintained for output-shape records |

---

## Acceptance Criteria

Numbers to diff against. Every one was observed on the prototype or the
frozen copy before it was written here.

1. **The headline strings do not change.** On one fresh frozen copy, run the
   binary from `main` and the binary from this branch. The `Entries:` line of
   `impact --year` and of `wrapped` is byte-identical between them. At design
   that was `Entries: 531/606 with impact` and `Entries: 606`. Build's copy
   will differ, and the **equality** is the criterion.
2. **The partition is exact on real data.** On that copy,
   `impact --year --format json` gives
   `entries_with_impact == sum(counts_by_project) == |impact_by_project rows| + |failures_by_project rows|`
   (531 = 531 = 527 + 4 at design). The `failures_by_project` ids equal the ids
   from `brag list --type failed --format json` that have a non-empty impact
   and fall in the window (433, 465, 473, 420 at design).
3. **Both surfaces agree.** On that copy, the impact section and the failure
   section of `impact --year` are byte-identical to `wrapped --no-spark`'s
   `## Impact moments` and `## What didn't work`, apart from the one blank line
   that separates sections.
4. **The section is conditional.** A clean fixture renders no
   `## What didn't work` line on either surface. A failures-only window renders
   no `## Impact` line. `failures_by_project` is `[]` in both empty-shape
   tests.
5. **The JSON keys are as enumerated** under *Fork A* (in order), and
   `impact_by_project` and `impact_moments` carry no row with a failure's id.
6. **The goldens:** exactly two existing goldens change, 41 → 42 and 145 → 146
   `want` lines, both in `internal/export`. Three new goldens exist, at 32, 66
   and 60 `want` lines. No markdown golden changes. No existing `internal/cli`
   assertion changes; `learn_test.go`'s `FailureType` references are only
   renamed (LD1). `memory_test.go` is untouched.
7. **The fixtures hold failures.** The impact fixture has 3 `failed` rows (2
   with an impact), the wrapped fixture has 3 (2 with an impact), the drift
   guard has 2 exact and 6 near-miss rows, and each e2e test has 1. **No
   assertion that proves new behaviour runs on a zero-failure fixture.** The
   zero-failure cases in the Fork D tests are the paired controls.
8. **The predicate agrees with retrieval.**
   `TestFailureClassifier_GoPredicateMatchesTypeFilter` passes, and M-A, M-B
   and M-S each turn it red.
9. **Help:** both LD9 sentences appear verbatim in `--help`.
10. **Test counts:** 829 → **843** top-level test functions, and 1086 →
    **1100** passing tests including subtests. 79 → **80** test files.
11. **`just test-docs`:** ALL OK at **206 `OK:` lines / 205 distinct ids**
    (from 201 / 200). `grep -n '===== finalise =====\|===== Group AD'` puts
    `AD` **lower**, and `./scripts/test-docs.sh | grep -c 'OK:   AD'` is
    **5**.
12. **The inventory block is regenerated,** and exactly three rows move at
    build: 79 → 80, 829 → 843, 200 → 205.
13. **No count is renamed:** `grep -n 'Entries:' internal/export/impact.go
    internal/export/wrapped.go` shows the same two format strings as `main`.
14. **All five gates are green,** with no test excluded and no lint
    suppression: `go test ./...` (14 packages `ok`), `gofmt -l .` empty,
    `go vet ./...` clean, `just lint` **0 issues**, `just test-docs` ALL OK.

---

## Failing Tests

Write them first and watch them fail for the stated reason (Finding 7 is the
expected picture), then make them pass. Every literal is in *Notes for the
Implementer*.

### New: `internal/aggregate/aggregate_test.go` (2)

| Test | Asserts | Fails before because |
|---|---|---|
| `TestIsFailure_ExactMatchOnTheReservedValue` | `"failed"` → true, and `Failed`, `FAILED`, `" failed"`, `"failed "`, `failure`, `learned`, `shipped`, `""` → false. Also `FailureType == "failed"`. | `IsFailure` and `FailureType` are undefined |
| `TestSplitFailures_PartitionsInOrderNeverNil` | The ids split `[1 3 5]` / `[2 4]` in order, neither half is nil on three edge inputs, and the halves sum to the input | `SplitFailures` is undefined |

### New: `internal/storage/failure_agreement_test.go` (1)

| Test | Asserts | Fails before because |
|---|---|---|
| `TestFailureClassifier_GoPredicateMatchesTypeFilter` | the SQL and Go id sets are equal over 8 seeds, and the count anchor is 2 | `aggregate.FailureType` and `SplitFailures` are undefined |

### New: `internal/export/impact_test.go` (4)

| Test | Asserts | Fails before because |
|---|---|---|
| `TestToImpactMarkdown_FailureSectionGolden` | the full document, 32 lines: `Entries: 5/8`, with ids 6 and 8 under `## What didn't work` and id 7 nowhere | id 6 renders under `## Impact`, and there is no second section |
| `TestToImpactJSON_FailureSectionGolden` | the full envelope, 66 lines: `counts_by_project` alpha 3 / beta 1 / gamma 1, and `failures_by_project` holds ids 6 and 8 | the key does not exist, and `impact_by_project` holds 6 and 8 |
| `TestToImpactJSON_CountsByProjectSpansBothSections` | per project, count = rows across both sections, and the sum is 5 = `entries_with_impact` | **it does not fail before, by design.** On `main` every row is in `impact_by_project`, so the invariant holds. It is the guard against M-D (the narrowed count loop), which it catches and nothing pre-existing does. The JSON failure golden beside it is the half that fails first. |
| `TestToImpactMarkdown_SectionsRenderOnlyWhenNonEmpty` | the `##` headings are `[Impact]` / `[What didn't work]` / `[Impact, What didn't work]` for none / only / both | the failures-only and both cases render only `## Impact` |

### New: `internal/export/wrapped_test.go` (3)

| Test | Asserts | Fails before because |
|---|---|---|
| `TestToWrappedMarkdown_FailureSectionGolden` | the full Q3 document, 60 lines, with the section between Impact moments and Rhythm | the failures sit in `## Impact moments` |
| `TestToWrappedJSON_FailuresLeaveImpactMoments` | compacted `impact_moments` holds only id 1, `failures_by_project` holds ids 3 and 5, and `total_entries` is 5 | `impact_moments` holds 1, 3 and 5, and the key is `<missing key>` |
| `TestToWrappedMarkdown_WhatDidntWorkRendersOnlyWhenNonEmpty` | clean: the five DEC-030 headings. With failures: six, in order. | the with-failures case has five |

### New: `internal/cli` (4)

| Test | File | Asserts | Fails before because |
|---|---|---|---|
| `TestLearnCmd_ImpactSectionsWhatItWrote` | `learn_test.go` | `brag add` and `brag learn` on one store, then `brag impact --since 2000-01-01`: the learn id is under `## What didn't work` with its impact and not under `## Impact`, and the JSON ids split `1` / `2` | the learn entry is under `## Impact`, and `failures_by_project` is empty |
| `TestLearnCmd_WrappedSectionsWhatItWrote` | `learn_test.go` | the same, through `brag wrapped <year of the stored row>` | the learn entry is under `## Impact moments` |
| `TestImpactCmd_HelpNamesTheFailureSection` | `impact_test.go` | the LD9 sentence, verbatim | it is absent |
| `TestWrappedCmd_HelpListsTheFailureSectionInArcOrder` | `wrapped_test.go` | `Impact moments, What didn't work (…), Rhythm`, verbatim | it is absent |

### Changed, as planned rewrites (4)

`TestToImpactJSON_DEC028ShapeGolden` and `TestToWrappedJSON_DEC030ShapeGolden`
each gain one line: the new key, `[]`. `TestToImpact_EmptyWindowShape` and
`TestToWrapped_EmptyPeriodShape` each gain one assertion that the new key is
`[]`. All four are red before build, which Finding 7 observed.

### New: `scripts/test-docs.sh`, Group `AD` (5 ids)

| Id | Asserts, scoped to a section | Fails before because |
|---|---|---|
| `AD1` | `docs/api-contract.md` `brag impact` section: `## What didn't work` and `` `failures_by_project` `` | neither is present |
| `AD2` | the `brag wrapped` section: the same two | neither is present |
| `AD3` | `docs/tutorial.md` `brag impact` section: `What didn't work` | absent |
| `AD4` | the tutorial's `brag wrapped` section: `What didn't work` | absent |
| `AD5` | the `AGENTS.md` `wrapped` glossary line: `Impact moments → What didn't work` | absent |

A section that is not found is its own failure (M-D4), not a vacuous pass.

### Mutation checks (build re-runs all 19, recorded in Build Completion)

The probes, their stated diffs and their design-time hashes are in the
**Mutation matrix** above. Build should reproduce each hash from its stated
diff **before** running the gate. If a stated diff does not reproduce its
hash, say so, and do not search for one that does. M-D and M-G carry the
most weight: each is caught only by this spec's tests.

### Decision-to-test mapping (§9)

| Decision | Test(s) that fail without it |
|---|---|
| LD1 one constant, in `aggregate` | `TestIsFailure_…`, `TestLearnCmd_PinsFailedType`; M-C, M-H |
| LD2 partition, not filter; shared writer | both failure goldens, both e2e tests, `…FailuresLeaveImpactMoments`; M-1, M-M |
| LD3 no count changes meaning | `TestToImpactJSON_CountsByProjectSpansBothSections`; M-D |
| LD4 the JSON keys, their order and emptiness | both shape goldens, both empty-shape tests, `…FailuresLeaveImpactMoments`; M-K |
| LD5 only when non-empty | `…SectionsRenderOnlyWhenNonEmpty`, `…WhatDidntWorkRendersOnlyWhenNonEmpty`, and the existing markdown goldens; M-E, M-F, M-G |
| LD6 the DEC-030 arc position | `TestToWrappedMarkdown_FailureSectionGolden`, `…WhatDidntWorkRendersOnlyWhenNonEmpty`, `AD5`; M-D3 |
| LD7 the drift guard | `TestFailureClassifier_…`; M-A, M-B, M-S |
| LD8 literal fixtures; e2e with no literal | M-C fires the goldens while the e2e tests stay green; M-H fires the e2e tests |
| LD9 help literals | both help tests; M-L1, M-L2 |
| LD10 the doc sweep | `AD1`–`AD5`; M-D1, M-D2, M-D3 |
| LD11 `AD` placement, name and ids | AC-11's `grep -n` and `grep -c 'OK:   AD'` = 5, and 205 distinct ids |
| LD12 negatives paired | M-E and M-F show each negative is live, and the goldens show each positive is |
| LD13 scope | `git diff --stat` touches no `internal/story` or `summary.go` |
| LD14 regenerated block | `X3` |

---

## Implementation Context

*Read this section, and the files it points to, before starting the build
cycle.*

### Decisions that apply

- **DEC-050**, written at this spec's design. **Read it first.** Rows 1 and 2
  are this spec, and its five rules are the contract. Rows 3 and 4 are
  SPEC-094's, so do not touch `summary` or `story`.
- **DEC-030**, including its **`## Amendment (2026-09-18, SPEC-086 design)`**.
  It fixes the section's position in `wrapped`'s arc and why the section
  alone is conditional.
- **DEC-049** covers the reserved value `failed` and its verb. LD1 moves the
  constant; it does not change the value.
- **DEC-048** says a count must name what it counted. LD3 is its whole
  application here.
- **DEC-028** covers `impact`'s two-number headline, its 4-key projection, and
  the impact-first body. All three are kept. **Read its
  `## Amendment (2026-09-19, SPEC-086 design)` too**: the choice 5 key list and
  the one-section body above it predate `failures_by_project`, and the
  amendment is what the envelope actually is.
- **DEC-014** covers the envelope. Part 4's empty-state rule gives the new key
  its `[]` form.
- **DEC-029**: profiles are data. It is why `story` is SPEC-094's and not
  this spec's.

### Constraints that apply

- `no-sql-in-cli-layer` (blocking). The new CLI tests open a store through
  `storage.Open` and `Store.Get` only. No SQL enters `internal/cli`, in
  production or in tests.
- `storage-tests-use-tempdir` (blocking). The drift guard's database is under
  `t.TempDir()`.
- `test-before-implementation` (blocking). Follow *Order of work*, step 2.
- `stdout-is-for-data-stderr-is-for-humans` (blocking). Nothing here writes
  to stderr, and both digests stay stdout-only.
- `one-spec-per-pr` (blocking).

### Prior related work

- **SPEC-085** shipped `brag learn`, `cli.FailureType` and DEC-049. It is
  where the user decided the section.
- **SPEC-045 / `aggregate.IsAgentAuthored`** is the single-source precedent,
  and `internal/storage/provenance_agreement_test.go` is the guard to mirror.
- **SPEC-084 / DEC-048** is the precedent for naming a breaking JSON change in
  the CHANGELOG.
- **SPEC-087** made `Y3` derive. It is why this design re-pins nothing by
  hand.
- **SPEC-088** added the `AC2` stray-tag guard. It is live for every file
  build writes.

### Out of scope (for this spec specifically)

- `summary` and `story` renderer work. It is **SPEC-094**, and DEC-050 rows 3
  and 4 already decide its posture.
- `--type` negation, which is a STAGE-023 `bug` entry, and the `--type` error
  message together with `add --json`'s repeated key, which are SPEC-090.
- STAGE-027 (agent edit, v0.8.0), `brag lint` / STAGE-025, STAGE-020,
  SPEC-092, SPEC-093, and wiring `test-docs` into CI.
- The impact-quality classifier (STAGE-024). **Sectioning by `type` is not
  ranking by quality.**
- `wrapped`'s bare `## Impact moments` on a period with no impact at all. It
  is existing behaviour that LD5 leaves alone.
- `guidance/questions.yaml`. No question was filed, so `Y4`'s pins do not
  move.

---

## Notes for the Implementer

### Order of work

1. **Re-derive before trusting a number.** Take a fresh file copy of
   `~/.bragfile/db.sqlite` with `sqlite3 ~/.bragfile/db.sqlite ".backup
   '<tmp>/db.sqlite'"`, which takes a read lock only. Run `main`'s binary
   against it for AC-1's *before* strings. **Never write to the live
   corpus.**
2. **Write the tests first:** §3's new and modified test code and §7's new
   file. Leave `learn_test.go`'s existing `FailureType` references alone for
   now. Run `go test ./...` and compare with Finding 7: `aggregate` and
   `storage` build-fail on the new symbols, `export` fails 10 tests, and
   `cli` fails 4. `TestToImpactJSON_CountsByProjectSpansBothSections`
   **passes**, which is expected.
3. **Production code:** §1 (`aggregate.go`), §2 (`learn.go`, and only now
   rename `learn_test.go`'s references), §4 (the two renderers) and §5 (the
   two `Long` strings). Then run `go test ./...` until it is green.
4. **Docs:** §6 (`api-contract.md`, `tutorial.md`, `AGENTS.md`,
   `CHANGELOG.md`).
5. **Group `AD`:** §8, inserted **above** `# ===== finalise =====`. Confirm
   with `grep -n '===== finalise =====\|===== Group AD' scripts/test-docs.sh`
   that `AD` has the lower line number.
6. **Inventory, last:**
   `./scripts/inventory.sh`, pasted between the markers with no blank line
   inside. `just inventory` only prints. Exactly three rows move.
7. **Mutation matrix:** re-run all 19 probes. Reproduce each stated hash from
   its stated diff **before** running the gate. A hash that does not move is
   a no-op, and it produces no evidence.
8. **Gates:** `just test`, `just test-docs`, `just lint`, `gofmt -l .` and
   `go vet ./...`.

### Traps

- **`AC2` is live.** A closing tool-call tag alone on a line fails
  `just test-docs` in any file you wrote, staged or not. Run it before every
  commit. Name such a tag only inline, in backticks.
- **Use `/usr/bin/grep` for counts.** In this zsh, bare `grep` is a `ugrep`
  wrapper that respects `.gitignore`.
- **`just advance-cycle` strips the inline comment on the `cycle:` line.**
  Restore it by hand.
- **SPEC-091 (in build) edits the same two constructors,** adding `GroupID`
  to `NewImpactCmd` and `NewWrappedCmd`. That is a textual overlap, not a
  semantic one. Whichever PR lands second rebases, and SPEC-091's LD4 already
  says its doc additions to `Long` strings are additive.
- **The literals below are `git diff` output from the prototype** (base
  `3201f50`), so `git apply` accepts them against that base. After any rebase,
  transcribe them by hand rather than forcing the apply.

### §1. `internal/aggregate/aggregate.go`

Inserted directly after `WithImpact`.

````diff
diff --git a/internal/aggregate/aggregate.go b/internal/aggregate/aggregate.go
index 568c6ab..dbcbd3d 100644
--- a/internal/aggregate/aggregate.go
+++ b/internal/aggregate/aggregate.go
@@ -301,6 +301,45 @@ func WithImpact(entries []storage.Entry) []storage.Entry {
 	return out
 }
 
+// FailureType is the reserved entries.type value marking work that did not
+// work (DEC-049) — the ONE type value bragfile pins; `brag add --type` stays
+// free-form. It lives here rather than in internal/cli because it has a
+// writer and readers: `brag learn` writes it, and the digests that section
+// failures (brag impact, brag wrapped — DEC-050) read it through IsFailure.
+// One constant, so the verb and the digests cannot name two values. The
+// literal is persisted in every user's corpus, so renaming it would orphan
+// every stored row while the writer and the readers still agreed.
+const FailureType = "failed"
+
+// IsFailure reports whether e is a recorded failure: its Type is exactly
+// FailureType. Exact and case-sensitive on purpose — the comparison storage's
+// --type filter makes (`e.type = ?` on a BINARY-collated column), so a
+// digest's "What didn't work" section and `brag list --type failed` (DEC-049's
+// retrieval path) select the same rows. Kept in agreement by
+// TestFailureClassifier_GoPredicateMatchesTypeFilter.
+func IsFailure(e storage.Entry) bool {
+	return e.Type == FailureType
+}
+
+// SplitFailures partitions entries into those that are not failures and those
+// that are (IsFailure), preserving input order within each. Both results are
+// non-nil so JSON callers never see null. The digests call it on the
+// with-impact subset, never in place of WithImpact: WithImpact keeps meaning
+// "non-empty impact", and this decides only which section a row renders in
+// (DEC-050).
+func SplitFailures(entries []storage.Entry) (others, failures []storage.Entry) {
+	others = make([]storage.Entry, 0, len(entries))
+	failures = make([]storage.Entry, 0)
+	for _, e := range entries {
+		if IsFailure(e) {
+			failures = append(failures, e)
+		} else {
+			others = append(others, e)
+		}
+	}
+	return others, failures
+}
+
 // CadenceBucket is one month's entry count in a cadence series: Period
 // is the "YYYY-MM" label, Count the number of entries whose created_at
 // falls in that month. SPEC-051. SPEC-052 renders series[].Count as a
````

### §2. `internal/cli/learn.go`, and the rename in `learn_test.go`

````diff
diff --git a/internal/cli/learn.go b/internal/cli/learn.go
index 8e6fdbe..a916eef 100644
--- a/internal/cli/learn.go
+++ b/internal/cli/learn.go
@@ -6,26 +6,22 @@ import (
 
 	"github.com/spf13/cobra"
 
+	"github.com/jysf/bragfile000/internal/aggregate"
 	"github.com/jysf/bragfile000/internal/capture"
 	"github.com/jysf/bragfile000/internal/config"
 	"github.com/jysf/bragfile000/internal/editor"
 	"github.com/jysf/bragfile000/internal/storage"
 )
 
-// FailureType is the reserved entries.type value marking work that did not
-// work (DEC-049). It is the ONE type value bragfile pins: `brag add --type`
-// stays free-form, and `brag learn` exists so this value cannot fragment the
-// way `shipped`/`ship` and `fixed`/`bugfix` already have in the live corpus.
-const FailureType = "failed"
-
 // learnFieldFlags are the entry-field flags whose presence routes `brag
 // learn` to flag mode. "type" is deliberately ABSENT: the value is pinned,
 // not chosen (DEC-049).
 var learnFieldFlags = []string{"title", "description", "tags", "project", "impact"}
 
 // NewLearnCmd builds `brag learn` — the capture verb for work that did not
-// work. It is `brag add` with entries.type pinned to FailureType and the
-// milestone nudge suppressed; see runLearn for why the nudge is dropped.
+// work. It is `brag add` with entries.type pinned to aggregate.FailureType
+// (DEC-049) and the milestone nudge suppressed; see runLearn for why the
+// nudge is dropped.
 //
 // Two modes, mirroring add's (DEC-007 / DEC-009) minus JSON mode:
 //   - flag mode: any of the five entry-field flags set; --title is required.
@@ -96,7 +92,7 @@ func runLearnFlags(cmd *cobra.Command, _ []string) error {
 		Description: getFlagString(cmd, "description"),
 		Tags:        getFlagString(cmd, "tags"),
 		Project:     getFlagString(cmd, "project"),
-		Type:        FailureType,
+		Type:        aggregate.FailureType,
 		Impact:      getFlagString(cmd, "impact"),
 	}, cmd.Flags().Changed("project"))
 }
@@ -125,7 +121,7 @@ func runLearnEditor(cmd *cobra.Command) error {
 		Description: parsed.Description,
 		Tags:        parsed.Tags,
 		Project:     parsed.Project,
-		Type:        FailureType,
+		Type:        aggregate.FailureType,
 		Impact:      parsed.Impact,
 	}, parsed.Project != "")
 }
````

`learn_test.go`'s rename is part of §3's `learn_test.go` diff: the import,
and the seven `FailureType` references that become `aggregate.FailureType`.

### §3. Test code, new and modified

`internal/aggregate/aggregate_test.go`:

````diff
diff --git a/internal/aggregate/aggregate_test.go b/internal/aggregate/aggregate_test.go
index 1da5069..88ec09d 100644
--- a/internal/aggregate/aggregate_test.go
+++ b/internal/aggregate/aggregate_test.go
@@ -649,6 +649,80 @@ func TestWithImpact_EmptyInputAndAllEmptyImpact(t *testing.T) {
 	}
 }
 
+// TestIsFailure_ExactMatchOnTheReservedValue pins SPEC-086 LD1: a failure is
+// exactly the reserved value DEC-049 persists — the literal "failed", matched
+// case-sensitively and untrimmed, the comparison storage's --type filter makes.
+// Every near-miss below is a real spelling an agent or a user can write with
+// `brag add --type`, and none of them is a failure.
+func TestIsFailure_ExactMatchOnTheReservedValue(t *testing.T) {
+	cases := []struct {
+		typ  string
+		want bool
+	}{
+		{"failed", true},
+		{"Failed", false},
+		{"FAILED", false},
+		{" failed", false},
+		{"failed ", false},
+		{"failure", false},
+		{"learned", false},
+		{"shipped", false},
+		{"", false},
+	}
+	for _, c := range cases {
+		if got := IsFailure(storage.Entry{Type: c.typ}); got != c.want {
+			t.Errorf("IsFailure(Type=%q) = %v, want %v", c.typ, got, c.want)
+		}
+	}
+	if FailureType != "failed" {
+		t.Errorf("FailureType = %q, want %q (the value is persisted in every corpus)", FailureType, "failed")
+	}
+}
+
+// TestSplitFailures_PartitionsInOrderNeverNil pins SPEC-086 LD2: the split is a
+// partition — every entry lands in exactly one half, input order is kept inside
+// each half, and neither half is ever nil, so an empty section marshals as [].
+func TestSplitFailures_PartitionsInOrderNeverNil(t *testing.T) {
+	in := []storage.Entry{
+		{ID: 1, Type: "shipped"},
+		{ID: 2, Type: "failed"},
+		{ID: 3, Type: ""},
+		{ID: 4, Type: "failed"},
+		{ID: 5, Type: "Failed"},
+	}
+	others, failures := SplitFailures(in)
+	ids := func(es []storage.Entry) []int64 {
+		out := []int64{}
+		for _, e := range es {
+			out = append(out, e.ID)
+		}
+		return out
+	}
+	if got, want := ids(others), []int64{1, 3, 5}; !reflect.DeepEqual(got, want) {
+		t.Errorf("others = %v, want %v", got, want)
+	}
+	if got, want := ids(failures), []int64{2, 4}; !reflect.DeepEqual(got, want) {
+		t.Errorf("failures = %v, want %v", got, want)
+	}
+
+	for _, tc := range []struct {
+		name string
+		in   []storage.Entry
+	}{
+		{"nil input", nil},
+		{"no failures", []storage.Entry{{ID: 1, Type: "shipped"}}},
+		{"only failures", []storage.Entry{{ID: 1, Type: "failed"}}},
+	} {
+		o, f := SplitFailures(tc.in)
+		if o == nil || f == nil {
+			t.Errorf("%s: SplitFailures returned a nil half (others nil=%v, failures nil=%v)", tc.name, o == nil, f == nil)
+		}
+		if len(o)+len(f) != len(tc.in) {
+			t.Errorf("%s: halves hold %d entries, input had %d", tc.name, len(o)+len(f), len(tc.in))
+		}
+	}
+}
+
 // --- SPEC-045: provenance coverage helpers ------------------------------
 
 // coverageAggFixture mirrors the export package's coverageYearFixture (kept
````

`internal/export/impact_test.go`:

````diff
diff --git a/internal/export/impact_test.go b/internal/export/impact_test.go
index 45c7605..2e903d3 100644
--- a/internal/export/impact_test.go
+++ b/internal/export/impact_test.go
@@ -2,6 +2,7 @@ package export
 
 import (
 	"encoding/json"
+	"reflect"
 	"strings"
 	"testing"
 	"time"
@@ -137,7 +138,8 @@ func TestToImpactJSON_DEC028ShapeGolden(t *testing.T) {
         }
       ]
     }
-  ]
+  ],
+  "failures_by_project": []
 }`
 	if string(got) != want {
 		t.Errorf("json golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
@@ -191,6 +193,9 @@ Entries: 0/0 with impact`
 	if string(env["impact_by_project"]) != "[]" {
 		t.Errorf("impact_by_project: got %s, want []", env["impact_by_project"])
 	}
+	if string(env["failures_by_project"]) != "[]" {
+		t.Errorf("failures_by_project: got %s, want []", env["failures_by_project"])
+	}
 	if string(env["filters"]) != "{}" {
 		t.Errorf("filters: got %s, want {}", env["filters"])
 	}
@@ -304,3 +309,247 @@ func TestToImpactMarkdown_FiltersEchoed(t *testing.T) {
 		t.Errorf("filters: got %v, want exactly {project: alpha}", env.Filters)
 	}
 }
+
+// impactFailureFixture is impactFixture plus three rows typed "failed" — the
+// literal DEC-049 persists, spelled out rather than taken from
+// aggregate.FailureType so these tests also fail if the constant drifts from
+// the stored value. id 6 is an alpha failure WITH an impact, dated between
+// alpha's two wins, so it must leave their group; id 7 is a failure with NO
+// impact, so it appears nowhere (impact-first, DEC-028 choice 3); id 8 puts
+// gamma in the body only through a failure, so gamma is a failures-only
+// project. 8 in window, 5 with impact: 3 in ## Impact, 2 in ## What didn't
+// work.
+var impactFailureFixture = append(append([]storage.Entry{}, impactFixture...),
+	storage.Entry{ID: 6, Title: "pool-dead-end",
+		Project: "alpha", Type: "failed",
+		Impact:    "cost two days and produced nothing reusable",
+		CreatedAt: time.Date(2026, 7, 2, 11, 0, 0, 0, time.UTC),
+		UpdatedAt: time.Date(2026, 7, 2, 11, 0, 0, 0, time.UTC)},
+	storage.Entry{ID: 7, Title: "retry-noimpact",
+		Project: "delta", Type: "failed",
+		Impact:    "", // a failure with no impact → counted, not shown
+		CreatedAt: time.Date(2026, 7, 3, 11, 0, 0, 0, time.UTC),
+		UpdatedAt: time.Date(2026, 7, 3, 11, 0, 0, 0, time.UTC)},
+	storage.Entry{ID: 8, Title: "vendor-sdk-dead-end",
+		Project: "gamma", Type: "failed",
+		Impact:    "ruled out the vendor SDK",
+		CreatedAt: time.Date(2026, 7, 5, 11, 0, 0, 0, time.UTC),
+		UpdatedAt: time.Date(2026, 7, 5, 11, 0, 0, 0, time.UTC)},
+)
+
+var impactFailureOpts = ImpactOptions{
+	Scope:           "quarter",
+	Filters:         "(none)",
+	EntriesInWindow: 8,
+	Now:             impactFixedNow,
+}
+
+// TestToImpactMarkdown_FailureSectionGolden (LOAD-BEARING, SPEC-086 LD2/LD3).
+// A recorded failure with an impact leaves ## Impact and renders under
+// ## What didn't work, in the same per-entry shape; the Entries: tally is
+// unchanged in meaning — 5 is the with-impact subset, which both sections show.
+func TestToImpactMarkdown_FailureSectionGolden(t *testing.T) {
+	got, err := ToImpactMarkdown(impactFailureFixture, impactFailureOpts)
+	if err != nil {
+		t.Fatalf("unexpected error: %v", err)
+	}
+	want := `# Bragfile Impact
+
+Generated: 2026-07-06T12:00:00Z
+Scope: quarter
+Filters: (none)
+Entries: 5/8 with impact
+
+## Impact
+
+### alpha
+
+- 1: alpha-old
+  cut p95 login latency 40%
+- 4: alpha-new
+  removed the nightly cron entirely
+
+### beta
+
+- 2: beta-mid
+  onboarding time down to 1 day
+
+## What didn't work
+
+### alpha
+
+- 6: pool-dead-end
+  cost two days and produced nothing reusable
+
+### gamma
+
+- 8: vendor-sdk-dead-end
+  ruled out the vendor SDK`
+	if string(got) != want {
+		t.Errorf("markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
+	}
+}
+
+// TestToImpactJSON_FailureSectionGolden (LOAD-BEARING, SPEC-086 LD4). The
+// failures leave impact_by_project and arrive in failures_by_project, the
+// same group shape and 4-key projection; counts_by_project still counts both.
+func TestToImpactJSON_FailureSectionGolden(t *testing.T) {
+	got, err := ToImpactJSON(impactFailureFixture, impactFailureOpts)
+	if err != nil {
+		t.Fatalf("unexpected error: %v", err)
+	}
+	want := `{
+  "generated_at": "2026-07-06T12:00:00Z",
+  "scope": "quarter",
+  "filters": {},
+  "entries_in_window": 8,
+  "entries_with_impact": 5,
+  "counts_by_project": {
+    "alpha": 3,
+    "beta": 1,
+    "gamma": 1
+  },
+  "impact_by_project": [
+    {
+      "project": "alpha",
+      "entries": [
+        {
+          "id": 1,
+          "title": "alpha-old",
+          "project": "alpha",
+          "impact": "cut p95 login latency 40%"
+        },
+        {
+          "id": 4,
+          "title": "alpha-new",
+          "project": "alpha",
+          "impact": "removed the nightly cron entirely"
+        }
+      ]
+    },
+    {
+      "project": "beta",
+      "entries": [
+        {
+          "id": 2,
+          "title": "beta-mid",
+          "project": "beta",
+          "impact": "onboarding time down to 1 day"
+        }
+      ]
+    }
+  ],
+  "failures_by_project": [
+    {
+      "project": "alpha",
+      "entries": [
+        {
+          "id": 6,
+          "title": "pool-dead-end",
+          "project": "alpha",
+          "impact": "cost two days and produced nothing reusable"
+        }
+      ]
+    },
+    {
+      "project": "gamma",
+      "entries": [
+        {
+          "id": 8,
+          "title": "vendor-sdk-dead-end",
+          "project": "gamma",
+          "impact": "ruled out the vendor SDK"
+        }
+      ]
+    }
+  ]
+}`
+	if string(got) != want {
+		t.Errorf("json golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
+	}
+}
+
+// TestToImpactJSON_CountsByProjectSpansBothSections pins SPEC-086 LD3, the
+// DEC-048 obligation: no existing count changes what it counts. DEC-028
+// defines counts_by_project over the with-impact subset, so it keeps summing
+// to entries_with_impact, and each project's count is its rows across BOTH
+// sections. Deriving the map from the narrowed impact_by_project loop — the
+// obvious shortcut — makes alpha 2 and drops gamma, and fails here by name.
+func TestToImpactJSON_CountsByProjectSpansBothSections(t *testing.T) {
+	jsonBytes, err := ToImpactJSON(impactFailureFixture, impactFailureOpts)
+	if err != nil {
+		t.Fatalf("unexpected error: %v", err)
+	}
+	var env struct {
+		EntriesWithImpact int            `json:"entries_with_impact"`
+		CountsByProject   map[string]int `json:"counts_by_project"`
+		ImpactByProject   []struct {
+			Project string            `json:"project"`
+			Entries []json.RawMessage `json:"entries"`
+		} `json:"impact_by_project"`
+		FailuresByProject []struct {
+			Project string            `json:"project"`
+			Entries []json.RawMessage `json:"entries"`
+		} `json:"failures_by_project"`
+	}
+	if err := json.Unmarshal(jsonBytes, &env); err != nil {
+		t.Fatalf("json unmarshal: %v", err)
+	}
+	rows := map[string]int{}
+	for _, g := range env.ImpactByProject {
+		rows[g.Project] += len(g.Entries)
+	}
+	for _, g := range env.FailuresByProject {
+		rows[g.Project] += len(g.Entries)
+	}
+	sum := 0
+	for p, n := range env.CountsByProject {
+		sum += n
+		if rows[p] != n {
+			t.Errorf("counts_by_project[%q] = %d, but the two sections hold %d rows for it", p, n, rows[p])
+		}
+	}
+	if len(rows) != len(env.CountsByProject) {
+		t.Errorf("counts_by_project has %d projects, the two sections have %d", len(env.CountsByProject), len(rows))
+	}
+	if sum != env.EntriesWithImpact || sum != 5 {
+		t.Errorf("counts_by_project sums to %d; entries_with_impact is %d; want both 5", sum, env.EntriesWithImpact)
+	}
+}
+
+// TestToImpactMarkdown_SectionsRenderOnlyWhenNonEmpty pins SPEC-086 LD5 (Fork
+// D) on impact: each section heading appears only when it has an entry. A
+// clean window grows no "## What didn't work"; a failures-only window grows no
+// bare "## Impact". Line equality, not substring (AGENTS.md §9, SPEC-015).
+func TestToImpactMarkdown_SectionsRenderOnlyWhenNonEmpty(t *testing.T) {
+	headings := func(md []byte) []string {
+		var out []string
+		for _, ln := range strings.Split(string(md), "\n") {
+			if strings.HasPrefix(ln, "## ") {
+				out = append(out, ln)
+			}
+		}
+		return out
+	}
+	failuresOnly := []storage.Entry{impactFailureFixture[5], impactFailureFixture[7]}
+	cases := []struct {
+		name    string
+		entries []storage.Entry
+		want    []string
+	}{
+		{"no failures", impactFixture, []string{"## Impact"}},
+		{"failures only", failuresOnly, []string{"## What didn't work"}},
+		{"both", impactFailureFixture, []string{"## Impact", "## What didn't work"}},
+	}
+	for _, c := range cases {
+		opts := impactFailureOpts
+		opts.EntriesInWindow = len(c.entries)
+		md, err := ToImpactMarkdown(c.entries, opts)
+		if err != nil {
+			t.Fatalf("%s: unexpected error: %v", c.name, err)
+		}
+		if got := headings(md); !reflect.DeepEqual(got, c.want) {
+			t.Errorf("%s: ## headings = %q, want %q\n%s", c.name, got, c.want, md)
+		}
+	}
+}
````

`internal/export/wrapped_test.go`:

````diff
diff --git a/internal/export/wrapped_test.go b/internal/export/wrapped_test.go
index cf5be2f..60a8747 100644
--- a/internal/export/wrapped_test.go
+++ b/internal/export/wrapped_test.go
@@ -1,7 +1,9 @@
 package export
 
 import (
+	"bytes"
 	"encoding/json"
+	"reflect"
 	"strings"
 	"testing"
 	"time"
@@ -241,6 +243,7 @@ func TestToWrappedJSON_DEC030ShapeGolden(t *testing.T) {
       ]
     }
   ],
+  "failures_by_project": [],
   "longest_streak": 2,
   "top_tags": [
     {
@@ -504,6 +507,7 @@ Entries: 0`
 	assertRaw(t, env, "total_entries", "0")
 	assertRaw(t, env, "top_initiatives", "[]")
 	assertRaw(t, env, "impact_moments", "[]")
+	assertRaw(t, env, "failures_by_project", "[]")
 	assertRaw(t, env, "longest_streak", "0")
 	assertRaw(t, env, "top_tags", "[]")
 	assertRaw(t, env, "top_types", "[]")
@@ -656,3 +660,182 @@ func assertRaw(t *testing.T, env map[string]json.RawMessage, key, want string) {
 		t.Errorf("key %q: got %s, want %s", key, got, want)
 	}
 }
+
+// wrappedFailureFixture: a Q3 2026 period whose five entries exercise every
+// branch of the failure split. id 1 is a win with an impact (beta); id 2 is a
+// win without one (alpha); ids 3 and 5 are failures WITH an impact (alpha and
+// gamma — gamma appears only as a failure); id 4 is a failure with NO impact,
+// which lands in no per-entry section (DEC-028 choice 3) but is still counted
+// everywhere wrapped counts entries. Type is the literal "failed" DEC-049
+// persists, not aggregate.FailureType, so a drifted constant fails here too.
+var wrappedFailureFixture = []storage.Entry{
+	{ID: 1, Title: "launch", Project: "beta", Type: "shipped", Tags: "api",
+		Impact:    "onboarding time down to 1 day",
+		CreatedAt: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)},
+	{ID: 2, Title: "hotfix", Project: "alpha", Type: "fixed", Tags: "auth",
+		CreatedAt: time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)},
+	{ID: 3, Title: "pool-dead-end", Project: "alpha", Type: "failed", Tags: "perf",
+		Impact:    "cost two days and produced nothing reusable",
+		CreatedAt: time.Date(2026, 8, 10, 10, 0, 0, 0, time.UTC)},
+	{ID: 4, Title: "retry-noimpact", Project: "alpha", Type: "failed", Tags: "perf",
+		CreatedAt: time.Date(2026, 8, 11, 10, 0, 0, 0, time.UTC)},
+	{ID: 5, Title: "vendor-sdk-dead-end", Project: "gamma", Type: "failed", Tags: "vendor",
+		Impact:    "ruled out the vendor SDK",
+		CreatedAt: time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)},
+}
+
+var wrappedFailureOpts = WrappedOptions{
+	Scope:       "2026-Q3",
+	ScopeMonths: q3Months,
+	Filters:     "(none)",
+	Now:         time.Date(2026, 9, 30, 23, 59, 59, 0, time.UTC),
+}
+
+// TestToWrappedMarkdown_FailureSectionGolden (LOAD-BEARING, SPEC-086 LD2/LD6).
+// DEC-030's arc as amended: ## What didn't work sits between Impact moments
+// and Rhythm, carrying the with-impact failures in the impact rendering shape.
+// Nothing else in the document changes rule — Top types honestly reports
+// failed: 3, because it always counted every type.
+func TestToWrappedMarkdown_FailureSectionGolden(t *testing.T) {
+	got, err := ToWrappedMarkdown(wrappedFailureFixture, wrappedFailureOpts)
+	if err != nil {
+		t.Fatalf("unexpected error: %v", err)
+	}
+	want := `# Bragfile Wrapped
+
+Generated: 2026-09-30T23:59:59Z
+Scope: 2026-Q3
+Filters: (none)
+Entries: 5
+
+## Cadence
+
+Busiest month: 2026-07 (2)
+
+- 2026-07: 2
+- 2026-08: 2
+- 2026-09: 1
+
+## Top initiatives
+
+- alpha: 3
+- beta: 1
+- gamma: 1
+
+## Impact moments
+
+### beta
+
+- 1: launch
+  onboarding time down to 1 day
+
+## What didn't work
+
+### alpha
+
+- 3: pool-dead-end
+  cost two days and produced nothing reusable
+
+### gamma
+
+- 5: vendor-sdk-dead-end
+  ruled out the vendor SDK
+
+## Rhythm
+
+Longest streak: 2 days
+
+**Top tags**
+- perf: 2
+- api: 1
+- auth: 1
+- vendor: 1
+
+**Top types**
+- failed: 3
+- fixed: 1
+- shipped: 1
+
+## Span
+
+- First entry: 2026-07-04
+- Last entry: 2026-09-01
+- Active days: 60`
+	if string(got) != want {
+		t.Errorf("markdown golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
+	}
+}
+
+// TestToWrappedJSON_FailuresLeaveImpactMoments pins SPEC-086 LD4 on wrapped:
+// impact_moments no longer carries a failure, and failures_by_project carries
+// exactly the with-impact failures, in impact's group shape. Key ORDER is
+// pinned by TestToWrappedJSON_DEC030ShapeGolden.
+func TestToWrappedJSON_FailuresLeaveImpactMoments(t *testing.T) {
+	jsonBytes, err := ToWrappedJSON(wrappedFailureFixture, wrappedFailureOpts)
+	if err != nil {
+		t.Fatalf("unexpected error: %v", err)
+	}
+	var env map[string]json.RawMessage
+	if err := json.Unmarshal(jsonBytes, &env); err != nil {
+		t.Fatalf("json unmarshal: %v", err)
+	}
+	compact := func(key string) string {
+		raw, ok := env[key]
+		if !ok {
+			return "<missing key>"
+		}
+		var b bytes.Buffer
+		if err := json.Compact(&b, raw); err != nil {
+			t.Fatalf("compact %s: %v", key, err)
+		}
+		return b.String()
+	}
+	wantMoments := `[{"project":"beta","entries":[{"id":1,"title":"launch","project":"beta","impact":"onboarding time down to 1 day"}]}]`
+	wantFailures := `[{"project":"alpha","entries":[{"id":3,"title":"pool-dead-end","project":"alpha","impact":"cost two days and produced nothing reusable"}]},` +
+		`{"project":"gamma","entries":[{"id":5,"title":"vendor-sdk-dead-end","project":"gamma","impact":"ruled out the vendor SDK"}]}]`
+	if got := compact("impact_moments"); got != wantMoments {
+		t.Errorf("impact_moments:\n got %s\nwant %s", got, wantMoments)
+	}
+	if got := compact("failures_by_project"); got != wantFailures {
+		t.Errorf("failures_by_project:\n got %s\nwant %s", got, wantFailures)
+	}
+	if got := string(env["total_entries"]); got != "5" {
+		t.Errorf("total_entries = %s, want 5 (the headline counts every entry, failures included)", got)
+	}
+}
+
+// TestToWrappedMarkdown_WhatDidntWorkRendersOnlyWhenNonEmpty pins SPEC-086 LD5
+// (Fork D) on wrapped: the section is omitted from a period with no recorded
+// failure — no empty heading on a clean quarter — while DEC-030's five arc
+// sections all still render; with a failure it appears between Impact moments
+// and Rhythm. Line equality, not substring.
+func TestToWrappedMarkdown_WhatDidntWorkRendersOnlyWhenNonEmpty(t *testing.T) {
+	headings := func(md []byte) []string {
+		var out []string
+		for _, ln := range strings.Split(string(md), "\n") {
+			if strings.HasPrefix(ln, "## ") {
+				out = append(out, ln)
+			}
+		}
+		return out
+	}
+	clean, err := ToWrappedMarkdown(wrappedYearFixture, WrappedOptions{
+		Scope: "2026", ScopeMonths: yearMonths, Filters: "(none)", Now: wrappedYearNow,
+	})
+	if err != nil {
+		t.Fatalf("clean: unexpected error: %v", err)
+	}
+	wantClean := []string{"## Cadence", "## Top initiatives", "## Impact moments", "## Rhythm", "## Span"}
+	if got := headings(clean); !reflect.DeepEqual(got, wantClean) {
+		t.Errorf("no failures: ## headings = %q, want %q", got, wantClean)
+	}
+
+	withFailures, err := ToWrappedMarkdown(wrappedFailureFixture, wrappedFailureOpts)
+	if err != nil {
+		t.Fatalf("with failures: unexpected error: %v", err)
+	}
+	wantFailures := []string{"## Cadence", "## Top initiatives", "## Impact moments", "## What didn't work", "## Rhythm", "## Span"}
+	if got := headings(withFailures); !reflect.DeepEqual(got, wantFailures) {
+		t.Errorf("with failures: ## headings = %q, want %q", got, wantFailures)
+	}
+}
````

`internal/cli/learn_test.go`:

````diff
diff --git a/internal/cli/learn_test.go b/internal/cli/learn_test.go
index d5b7254..4abe209 100644
--- a/internal/cli/learn_test.go
+++ b/internal/cli/learn_test.go
@@ -2,6 +2,7 @@ package cli
 
 import (
 	"bytes"
+	"encoding/json"
 	"os"
 	"path/filepath"
 	"strconv"
@@ -10,6 +11,7 @@ import (
 
 	"github.com/spf13/cobra"
 
+	"github.com/jysf/bragfile000/internal/aggregate"
 	"github.com/jysf/bragfile000/internal/storage"
 )
 
@@ -53,11 +55,11 @@ func TestLearnCmd_PinsFailedType(t *testing.T) {
 	if err != nil {
 		t.Fatalf("Get(%d): %v", id, err)
 	}
-	if got.Type != FailureType {
-		t.Errorf("Type = %q, want %q", got.Type, FailureType)
+	if got.Type != aggregate.FailureType {
+		t.Errorf("Type = %q, want %q", got.Type, aggregate.FailureType)
 	}
-	if FailureType != "failed" {
-		t.Errorf("FailureType = %q, want %q", FailureType, "failed")
+	if aggregate.FailureType != "failed" {
+		t.Errorf("aggregate.FailureType = %q, want %q", aggregate.FailureType, "failed")
 	}
 }
 
@@ -144,8 +146,8 @@ func TestLearnCmd_EditorModeOverwritesUserType(t *testing.T) {
 	if err != nil {
 		t.Fatalf("Get(%d): %v", id, err)
 	}
-	if got.Type != FailureType {
-		t.Errorf("editor-mode Type = %q, want %q (user's header must be overwritten)", got.Type, FailureType)
+	if got.Type != aggregate.FailureType {
+		t.Errorf("editor-mode Type = %q, want %q (user's header must be overwritten)", got.Type, aggregate.FailureType)
 	}
 }
 
@@ -164,3 +166,138 @@ func TestLearnCmd_EmptyTitleIsUserError(t *testing.T) {
 		t.Errorf("error = %v, want it to mention --title is required", err)
 	}
 }
+
+// runDigestCorpus runs one brag invocation against dbPath on a fresh root
+// carrying the two writers (add, learn) and the two digests that section
+// failures (impact, wrapped). A fresh root per call, so no flag value leaks
+// from one invocation into the next.
+func runDigestCorpus(t *testing.T, dbPath string, args ...string) string {
+	t.Helper()
+	t.Setenv("BRAGFILE_DB", "")
+	addStderrIsTTY = func() bool { return false }
+	t.Cleanup(func() { addStderrIsTTY = defaultStderrIsTTY })
+	root := NewRootCmd("test")
+	root.AddCommand(NewAddCmd(), NewLearnCmd(), NewImpactCmd(), NewWrappedCmd())
+	var outBuf, errBuf bytes.Buffer
+	root.SetOut(&outBuf)
+	root.SetErr(&errBuf)
+	root.SetArgs(append([]string{"--db", dbPath}, args...))
+	if err := root.Execute(); err != nil {
+		t.Fatalf("brag %v: %v (stderr %q)", args, err, errBuf.String())
+	}
+	return outBuf.String()
+}
+
+// markdownSection returns the lines under the `## <heading>` line, up to the
+// next `## ` heading. Found by line equality, not substring (AGENTS.md §9).
+func markdownSection(md, heading string) string {
+	var out []string
+	in := false
+	for _, ln := range strings.Split(md, "\n") {
+		if strings.HasPrefix(ln, "## ") {
+			in = ln == "## "+heading
+			continue
+		}
+		if in {
+			out = append(out, ln)
+		}
+	}
+	return strings.Join(out, "\n")
+}
+
+// TestLearnCmd_ImpactSectionsWhatItWrote ▲ SPEC-086 LD1 — the writer and the
+// reader held to one value by running both, through a real store: the entry
+// `brag learn` wrote is the one `brag impact` lists under "What didn't work",
+// and the entry `brag add` wrote on the same corpus stays under "Impact". No
+// constant appears here, so this fails if the verb and the digest ever name
+// different values, whichever side moves.
+func TestLearnCmd_ImpactSectionsWhatItWrote(t *testing.T) {
+	dbPath := filepath.Join(t.TempDir(), "test.db")
+	winID := strings.TrimSpace(runDigestCorpus(t, dbPath, "add", "-t", "shipped the cache", "-p", "alpha", "-k", "shipped", "-i", "cut p95 40%"))
+	failID := strings.TrimSpace(runDigestCorpus(t, dbPath, "learn", "-t", "tried a worker pool", "-p", "alpha", "-i", "cost two days"))
+
+	md := runDigestCorpus(t, dbPath, "impact", "--since", "2000-01-01")
+	impact := markdownSection(md, "Impact")
+	failed := markdownSection(md, "What didn't work")
+	if !strings.Contains(impact, "- "+winID+": shipped the cache") {
+		t.Errorf("## Impact is missing the brag add entry %s:\n%s", winID, md)
+	}
+	if strings.Contains(impact, "- "+failID+":") {
+		t.Errorf("## Impact still carries the brag learn entry %s:\n%s", failID, md)
+	}
+	if !strings.Contains(failed, "- "+failID+": tried a worker pool\n  cost two days") {
+		t.Errorf("## What didn't work is missing the brag learn entry %s with its impact:\n%s", failID, md)
+	}
+
+	var env struct {
+		ImpactByProject []struct {
+			Entries []struct {
+				ID int64 `json:"id"`
+			} `json:"entries"`
+		} `json:"impact_by_project"`
+		FailuresByProject []struct {
+			Entries []struct {
+				ID int64 `json:"id"`
+			} `json:"entries"`
+		} `json:"failures_by_project"`
+	}
+	if err := json.Unmarshal([]byte(runDigestCorpus(t, dbPath, "impact", "--since", "2000-01-01", "--format", "json")), &env); err != nil {
+		t.Fatalf("json unmarshal: %v", err)
+	}
+	ids := func(groups []struct {
+		Entries []struct {
+			ID int64 `json:"id"`
+		} `json:"entries"`
+	}) string {
+		var out []string
+		for _, g := range groups {
+			for _, e := range g.Entries {
+				out = append(out, strconv.FormatInt(e.ID, 10))
+			}
+		}
+		return strings.Join(out, ",")
+	}
+	if got := ids(env.ImpactByProject); got != winID {
+		t.Errorf("impact_by_project ids = %q, want %q", got, winID)
+	}
+	if got := ids(env.FailuresByProject); got != failID {
+		t.Errorf("failures_by_project ids = %q, want %q", got, failID)
+	}
+}
+
+// TestLearnCmd_WrappedSectionsWhatItWrote ▲ SPEC-086 LD1 — the same
+// writer-to-reader check on brag wrapped. The period is named from the stored
+// row's own created_at, so the test never races a year boundary.
+func TestLearnCmd_WrappedSectionsWhatItWrote(t *testing.T) {
+	dbPath := filepath.Join(t.TempDir(), "test.db")
+	winID := strings.TrimSpace(runDigestCorpus(t, dbPath, "add", "-t", "shipped the cache", "-p", "alpha", "-k", "shipped", "-i", "cut p95 40%"))
+	failID := strings.TrimSpace(runDigestCorpus(t, dbPath, "learn", "-t", "tried a worker pool", "-p", "alpha", "-i", "cost two days"))
+
+	id, err := strconv.ParseInt(failID, 10, 64)
+	if err != nil {
+		t.Fatalf("learn stdout should be the id alone, got %q", failID)
+	}
+	s, err := storage.Open(dbPath)
+	if err != nil {
+		t.Fatalf("storage.Open: %v", err)
+	}
+	got, err := s.Get(id)
+	s.Close()
+	if err != nil {
+		t.Fatalf("Get(%d): %v", id, err)
+	}
+	year := strconv.Itoa(got.CreatedAt.UTC().Year())
+
+	md := runDigestCorpus(t, dbPath, "wrapped", year, "--no-spark")
+	moments := markdownSection(md, "Impact moments")
+	failed := markdownSection(md, "What didn't work")
+	if !strings.Contains(moments, "- "+winID+": shipped the cache") {
+		t.Errorf("## Impact moments is missing the brag add entry %s:\n%s", winID, md)
+	}
+	if strings.Contains(moments, "- "+failID+":") {
+		t.Errorf("## Impact moments still carries the brag learn entry %s:\n%s", failID, md)
+	}
+	if !strings.Contains(failed, "- "+failID+": tried a worker pool\n  cost two days") {
+		t.Errorf("## What didn't work is missing the brag learn entry %s with its impact:\n%s", failID, md)
+	}
+}
````

`internal/cli/impact_test.go`:

````diff
diff --git a/internal/cli/impact_test.go b/internal/cli/impact_test.go
index 154faad..d6c51c9 100644
--- a/internal/cli/impact_test.go
+++ b/internal/cli/impact_test.go
@@ -643,6 +643,21 @@ func TestImpactCmd_HelpShowsPrevious(t *testing.T) {
 	}
 }
 
+// TestImpactCmd_HelpNamesTheFailureSection ▲ SPEC-086 LD9: impact's --help
+// says where a brag learn entry goes, since it is no longer among the impact.
+// "What didn't work" is a token no cobra-generated line can produce.
+func TestImpactCmd_HelpNamesTheFailureSection(t *testing.T) {
+	root, outBuf, _ := newImpactTestRoot(t)
+	root.SetArgs([]string{"impact", "--help"})
+	if err := root.Execute(); err != nil {
+		t.Fatalf("unexpected error: %v", err)
+	}
+	want := `Work recorded with brag learn is listed under its own "What didn't work" heading instead of among the impact, and that heading is left out when there is none.`
+	if !bytes.Contains(outBuf.Bytes(), []byte(want)) {
+		t.Errorf("expected the failure-section sentence in help:\n%s", outBuf.String())
+	}
+}
+
 // TestImpactCmd_StdoutStderrSeparation_Previous: a successful --previous run
 // writes only stdout; the --since --previous combo writes only the returned
 // UserError (main.go routes it to stderr) with empty stdout.
````

`internal/cli/wrapped_test.go`:

````diff
diff --git a/internal/cli/wrapped_test.go b/internal/cli/wrapped_test.go
index 79e7eaf..6aee84f 100644
--- a/internal/cli/wrapped_test.go
+++ b/internal/cli/wrapped_test.go
@@ -511,6 +511,21 @@ func TestWrappedCmd_HelpShowsPrevious(t *testing.T) {
 	}
 }
 
+// TestWrappedCmd_HelpListsTheFailureSectionInArcOrder ▲ SPEC-086 LD9: the
+// section list in --help follows DEC-030's arc as amended, so the new section
+// is named between Impact moments and Rhythm, with its only-when-non-empty rule.
+func TestWrappedCmd_HelpListsTheFailureSectionInArcOrder(t *testing.T) {
+	root, outBuf, _ := newWrappedTestRoot(t)
+	root.SetArgs([]string{"wrapped", "--help"})
+	if err := root.Execute(); err != nil {
+		t.Fatalf("unexpected error: %v", err)
+	}
+	want := "Impact moments, What didn't work (work recorded with brag learn, left out when there is none), Rhythm"
+	if !bytes.Contains(outBuf.Bytes(), []byte(want)) {
+		t.Errorf("expected %q in help:\n%s", want, outBuf.String())
+	}
+}
+
 func TestWrappedCmd_StdoutStderrSeparation(t *testing.T) {
 	dbPath := filepath.Join(t.TempDir(), "test.db")
 	withNowFunc(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC))
````

### §4. The renderers

`internal/export/impact.go`:

````diff
diff --git a/internal/export/impact.go b/internal/export/impact.go
index 33e550a..f03eedf 100644
--- a/internal/export/impact.go
+++ b/internal/export/impact.go
@@ -28,15 +28,18 @@ type ImpactOptions struct {
 }
 
 // ToImpactMarkdown renders the in-window entries as an impact-first
-// digest per DEC-014/DEC-028. The renderer receives the already-in-
+// digest per DEC-014/DEC-028/DEC-050. The renderer receives the already-in-
 // window slice; it selects the with-impact subset (aggregate.WithImpact),
-// groups it by project (aggregate.GroupEntriesByProject), and renders
-// each shown entry's impact text in full. Returns bytes with the
-// trailing "\n" stripped (matches ToSummaryMarkdown). On zero with-
-// impact entries, only the header + provenance block is emitted; the
-// ## Impact body is omitted.
+// splits the recorded failures out of it (aggregate.SplitFailures), and
+// renders each half grouped by project with its impact text in full: the
+// rest under ## Impact, the failures under ## What didn't work. Each section
+// is emitted only when it has an entry (DEC-050), and the Entries: tally
+// still counts both — it is the with-impact subset the body shows. Returns
+// bytes with the trailing "\n" stripped (matches ToSummaryMarkdown). On zero
+// with-impact entries, only the header + provenance block is emitted.
 func ToImpactMarkdown(entries []storage.Entry, opts ImpactOptions) ([]byte, error) {
 	withImpact := aggregate.WithImpact(entries)
+	worked, failed := aggregate.SplitFailures(withImpact)
 
 	var buf bytes.Buffer
 	fmt.Fprintln(&buf, "# Bragfile Impact")
@@ -46,22 +49,34 @@ func ToImpactMarkdown(entries []storage.Entry, opts ImpactOptions) ([]byte, erro
 	fmt.Fprintf(&buf, "Filters: %s\n", opts.Filters)
 	fmt.Fprintf(&buf, "Entries: %d/%d with impact\n", len(withImpact), opts.EntriesInWindow)
 
-	if len(withImpact) == 0 {
-		return trimTrailingNewline(buf.Bytes()), nil
-	}
-
-	fmt.Fprintln(&buf)
-	fmt.Fprintln(&buf, "## Impact")
-	for _, group := range aggregate.GroupEntriesByProject(withImpact) {
+	if len(worked) > 0 {
 		fmt.Fprintln(&buf)
-		fmt.Fprintf(&buf, "### %s\n", group.Project)
+		fmt.Fprintln(&buf, "## Impact")
+		writeImpactGroups(&buf, worked)
+	}
+	if len(failed) > 0 {
 		fmt.Fprintln(&buf)
+		fmt.Fprintln(&buf, "## What didn't work")
+		writeImpactGroups(&buf, failed)
+	}
+	return trimTrailingNewline(buf.Bytes()), nil
+}
+
+// writeImpactGroups renders entries grouped by project as `### <project>`
+// blocks of `- <id>: <title>` plus an indented `  <impact>` line — the
+// per-entry shape DEC-028 choice 4 locks. brag impact and brag wrapped both
+// render their two impact-bearing sections through it, so an entry reads
+// byte-identically on either surface.
+func writeImpactGroups(buf *bytes.Buffer, entries []storage.Entry) {
+	for _, group := range aggregate.GroupEntriesByProject(entries) {
+		fmt.Fprintln(buf)
+		fmt.Fprintf(buf, "### %s\n", group.Project)
+		fmt.Fprintln(buf)
 		for _, e := range group.Entries {
-			fmt.Fprintf(&buf, "- %d: %s\n", e.ID, e.Title)
-			fmt.Fprintf(&buf, "  %s\n", e.Impact)
+			fmt.Fprintf(buf, "- %d: %s\n", e.ID, e.Title)
+			fmt.Fprintf(buf, "  %s\n", e.Impact)
 		}
 	}
-	return trimTrailingNewline(buf.Bytes()), nil
 }
 
 // impactEnvelope is the on-the-wire shape for ToImpactJSON. Field order
@@ -75,6 +90,7 @@ type impactEnvelope struct {
 	EntriesWithImpact int                  `json:"entries_with_impact"`
 	CountsByProject   map[string]int       `json:"counts_by_project"`
 	ImpactByProject   []impactProjectGroup `json:"impact_by_project"`
+	FailuresByProject []impactProjectGroup `json:"failures_by_project"`
 }
 
 type impactProjectGroup struct {
@@ -94,13 +110,16 @@ type impactEntry struct {
 }
 
 // ToImpactJSON renders the DEC-014 envelope with DEC-028's per-spec
-// payload keys: generated_at, scope, filters, entries_in_window,
-// entries_with_impact, counts_by_project (map over the with-impact
-// subset), impact_by_project (array of grouped 4-key projections).
-// 2-space indent. Empty-state per DEC-014 choice (4): counts {},
-// impact_by_project [], filters {}, never null.
+// payload keys plus DEC-050's: generated_at, scope, filters,
+// entries_in_window, entries_with_impact, counts_by_project (map over the
+// whole with-impact subset — failures included, so it still sums to
+// entries_with_impact), impact_by_project (the with-impact entries that are
+// not failures) and failures_by_project (the ones that are), each an array of
+// grouped 4-key projections. 2-space indent. Empty-state per DEC-014 choice
+// (4): counts {}, both arrays [], filters {}, never null.
 func ToImpactJSON(entries []storage.Entry, opts ImpactOptions) ([]byte, error) {
 	withImpact := aggregate.WithImpact(entries)
+	worked, failed := aggregate.SplitFailures(withImpact)
 
 	env := impactEnvelope{
 		GeneratedAt:       opts.Now.UTC().Format(time.RFC3339),
@@ -109,14 +128,26 @@ func ToImpactJSON(entries []storage.Entry, opts ImpactOptions) ([]byte, error) {
 		EntriesInWindow:   opts.EntriesInWindow,
 		EntriesWithImpact: len(withImpact),
 		CountsByProject:   map[string]int{},
-		ImpactByProject:   []impactProjectGroup{},
+		ImpactByProject:   impactGroups(worked),
+		FailuresByProject: impactGroups(failed),
 	}
 	if env.Filters == nil {
 		env.Filters = map[string]string{}
 	}
-
+	// Counted over withImpact, not over worked: DEC-028 defines this map over
+	// the with-impact subset, and narrowing it to one section would change
+	// what an existing key counts without renaming it (DEC-048).
 	for _, group := range aggregate.GroupEntriesByProject(withImpact) {
 		env.CountsByProject[group.Project] = len(group.Entries)
+	}
+	return json.MarshalIndent(env, "", "  ")
+}
+
+// impactGroups projects entries into project groups of the NARROW 4-key
+// entry shape. Non-nil on empty input, so an empty section renders [].
+func impactGroups(entries []storage.Entry) []impactProjectGroup {
+	out := make([]impactProjectGroup, 0)
+	for _, group := range aggregate.GroupEntriesByProject(entries) {
 		g := impactProjectGroup{
 			Project: group.Project,
 			Entries: make([]impactEntry, 0, len(group.Entries)),
@@ -129,7 +160,7 @@ func ToImpactJSON(entries []storage.Entry, opts ImpactOptions) ([]byte, error) {
 				Impact:  e.Impact,
 			})
 		}
-		env.ImpactByProject = append(env.ImpactByProject, g)
+		out = append(out, g)
 	}
-	return json.MarshalIndent(env, "", "  ")
+	return out
 }
````

`internal/export/wrapped.go`:

````diff
diff --git a/internal/export/wrapped.go b/internal/export/wrapped.go
index c9c3e6b..b25d9fb 100644
--- a/internal/export/wrapped.go
+++ b/internal/export/wrapped.go
@@ -39,7 +39,9 @@ type WrappedOptions struct {
 
 // ToWrappedMarkdown renders the in-period entries as the celebratory
 // wrapped digest per DEC-014/DEC-030: provenance, then the section arc
-// Cadence → Top initiatives → Impact moments → Rhythm → Span. Returns
+// Cadence → Top initiatives → Impact moments → What didn't work → Rhythm →
+// Span. What didn't work is the one section rendered only when non-empty
+// (DEC-050; DEC-030's Amendment). Returns
 // bytes with the trailing "\n" stripped (matches every other renderer).
 // On an empty period only the header + provenance block (through
 // "Entries: 0") is emitted; the body sections are omitted (DEC-014 part
@@ -89,17 +91,17 @@ func ToWrappedMarkdown(entries []storage.Entry, opts WrappedOptions) ([]byte, er
 		fmt.Fprintf(&buf, "- %s: %d\n", nc.Name, nc.Count)
 	}
 
-	// Impact moments (with-impact entries grouped by project, full text).
+	// Impact moments (with-impact entries that are not failures, grouped by
+	// project, full text), then What didn't work (the with-impact failures),
+	// which renders only when it has an entry (DEC-050).
+	worked, failed := aggregate.SplitFailures(aggregate.WithImpact(entries))
 	fmt.Fprintln(&buf)
 	fmt.Fprintln(&buf, "## Impact moments")
-	for _, group := range aggregate.GroupEntriesByProject(aggregate.WithImpact(entries)) {
+	writeImpactGroups(&buf, worked)
+	if len(failed) > 0 {
 		fmt.Fprintln(&buf)
-		fmt.Fprintf(&buf, "### %s\n", group.Project)
-		fmt.Fprintln(&buf)
-		for _, e := range group.Entries {
-			fmt.Fprintf(&buf, "- %d: %s\n", e.ID, e.Title)
-			fmt.Fprintf(&buf, "  %s\n", e.Impact)
-		}
+		fmt.Fprintln(&buf, "## What didn't work")
+		writeImpactGroups(&buf, failed)
 	}
 
 	// Rhythm (longest streak, top-5 tags, top-3 types).
@@ -137,17 +139,18 @@ func ToWrappedMarkdown(entries []storage.Entry, opts WrappedOptions) ([]byte, er
 // declaration order is the JSON key order DEC-014/DEC-030 lock
 // (encoding/json preserves it).
 type wrappedEnvelope struct {
-	GeneratedAt    string               `json:"generated_at"`
-	Scope          string               `json:"scope"`
-	Filters        map[string]string    `json:"filters"`
-	TotalEntries   int                  `json:"total_entries"`
-	Cadence        cadenceRecord        `json:"cadence"`
-	TopInitiatives []wrappedInitiative  `json:"top_initiatives"`
-	ImpactMoments  []wrappedImpactGroup `json:"impact_moments"`
-	LongestStreak  int                  `json:"longest_streak"`
-	TopTags        []wrappedNameCount   `json:"top_tags"`
-	TopTypes       []wrappedNameCount   `json:"top_types"`
-	Span           wrappedSpanRecord    `json:"span"`
+	GeneratedAt       string               `json:"generated_at"`
+	Scope             string               `json:"scope"`
+	Filters           map[string]string    `json:"filters"`
+	TotalEntries      int                  `json:"total_entries"`
+	Cadence           cadenceRecord        `json:"cadence"`
+	TopInitiatives    []wrappedInitiative  `json:"top_initiatives"`
+	ImpactMoments     []wrappedImpactGroup `json:"impact_moments"`
+	FailuresByProject []wrappedImpactGroup `json:"failures_by_project"`
+	LongestStreak     int                  `json:"longest_streak"`
+	TopTags           []wrappedNameCount   `json:"top_tags"`
+	TopTypes          []wrappedNameCount   `json:"top_types"`
+	Span              wrappedSpanRecord    `json:"span"`
 }
 
 // cadenceRecord uses *string for BusiestMonth so an empty period renders
@@ -199,16 +202,17 @@ func ToWrappedJSON(entries []storage.Entry, opts WrappedOptions) ([]byte, error)
 	series, busiest := aggregate.Cadence(entries, opts.ScopeMonths)
 
 	env := wrappedEnvelope{
-		GeneratedAt:    opts.Now.UTC().Format(time.RFC3339),
-		Scope:          opts.Scope,
-		Filters:        opts.FiltersJSON,
-		TotalEntries:   len(entries),
-		Cadence:        cadenceRecord{Series: series},
-		TopInitiatives: []wrappedInitiative{},
-		ImpactMoments:  []wrappedImpactGroup{},
-		TopTags:        []wrappedNameCount{},
-		TopTypes:       []wrappedNameCount{},
-		Span:           wrappedSpanRecord{},
+		GeneratedAt:       opts.Now.UTC().Format(time.RFC3339),
+		Scope:             opts.Scope,
+		Filters:           opts.FiltersJSON,
+		TotalEntries:      len(entries),
+		Cadence:           cadenceRecord{Series: series},
+		TopInitiatives:    []wrappedInitiative{},
+		ImpactMoments:     []wrappedImpactGroup{},
+		FailuresByProject: []wrappedImpactGroup{},
+		TopTags:           []wrappedNameCount{},
+		TopTypes:          []wrappedNameCount{},
+		Span:              wrappedSpanRecord{},
 	}
 	if env.Filters == nil {
 		env.Filters = map[string]string{}
@@ -222,21 +226,9 @@ func ToWrappedJSON(entries []storage.Entry, opts WrappedOptions) ([]byte, error)
 		for _, nc := range aggregate.MostCommon(extractProjects(entries), 5) {
 			env.TopInitiatives = append(env.TopInitiatives, wrappedInitiative{Project: nc.Name, Count: nc.Count})
 		}
-		for _, group := range aggregate.GroupEntriesByProject(aggregate.WithImpact(entries)) {
-			g := wrappedImpactGroup{
-				Project: group.Project,
-				Entries: make([]wrappedEntry, 0, len(group.Entries)),
-			}
-			for _, e := range group.Entries {
-				g.Entries = append(g.Entries, wrappedEntry{
-					ID:      e.ID,
-					Title:   e.Title,
-					Project: group.Project,
-					Impact:  e.Impact,
-				})
-			}
-			env.ImpactMoments = append(env.ImpactMoments, g)
-		}
+		worked, failed := aggregate.SplitFailures(aggregate.WithImpact(entries))
+		env.ImpactMoments = wrappedGroups(worked)
+		env.FailuresByProject = wrappedGroups(failed)
 		_, longest := aggregate.Streak(entries, opts.Now)
 		env.LongestStreak = longest
 		for _, nc := range aggregate.MostCommon(extractTags(entries), 5) {
@@ -258,6 +250,29 @@ func ToWrappedJSON(entries []storage.Entry, opts WrappedOptions) ([]byte, error)
 	return json.MarshalIndent(env, "", "  ")
 }
 
+// wrappedGroups projects entries into project groups of the same NARROW 4-key
+// entry shape impact uses. Non-nil on empty input, so an empty section
+// renders [].
+func wrappedGroups(entries []storage.Entry) []wrappedImpactGroup {
+	out := make([]wrappedImpactGroup, 0)
+	for _, group := range aggregate.GroupEntriesByProject(entries) {
+		g := wrappedImpactGroup{
+			Project: group.Project,
+			Entries: make([]wrappedEntry, 0, len(group.Entries)),
+		}
+		for _, e := range group.Entries {
+			g.Entries = append(g.Entries, wrappedEntry{
+				ID:      e.ID,
+				Title:   e.Title,
+				Project: group.Project,
+				Impact:  e.Impact,
+			})
+		}
+		out = append(out, g)
+	}
+	return out
+}
+
 // extractTypes returns each entry's non-empty Type field, suitable for
 // aggregate.MostCommon. Mirrors extractProjects (stats.go): empty Type
 // is excluded from counting.
````

### §5. The two `--help` sentences (LD9)

````diff
diff --git a/internal/cli/impact.go b/internal/cli/impact.go
index 7d23d65..f5c9bf3 100644
--- a/internal/cli/impact.go
+++ b/internal/cli/impact.go
@@ -32,7 +32,7 @@ Output is markdown (default) or a single-object JSON envelope (--format json) pe
   --year      the current calendar year, up to now
   --since D   entries on or after D (YYYY-MM-DD or Nd/Nw/Nm), up to now
 
-Windows are CALENDAR periods, not rolling — this differs from brag summary on purpose (the story surface reports by quarter/month/year). Only entries with a non-empty impact appear in the body; the provenance line tallies how many in-window entries had one. Filter flags --tag/--project/--type compose with the window.
+Windows are CALENDAR periods, not rolling — this differs from brag summary on purpose (the story surface reports by quarter/month/year). Only entries with a non-empty impact appear in the body; the provenance line tallies how many in-window entries had one. Work recorded with brag learn is listed under its own "What didn't work" heading instead of among the impact, and that heading is left out when there is none. Filter flags --tag/--project/--type compose with the window.
 
 --previous shifts the selected window to the last-completed period (bounded on both ends): --quarter --previous is the whole previous calendar quarter, --month --previous the previous month, --year --previous the previous year. It requires a window flag (a modifier is not a window) and is incompatible with --since.
 
````

````diff
diff --git a/internal/cli/wrapped.go b/internal/cli/wrapped.go
index 175672f..39a8057 100644
--- a/internal/cli/wrapped.go
+++ b/internal/cli/wrapped.go
@@ -42,7 +42,7 @@ The window is bounded on both ends: a named period covers only entries created w
 
 --previous (no positional period) covers the last-completed calendar year — brag wrapped --previous in 2026 is identical to brag wrapped 2025. It is valid only with no positional period: pairing it with an explicit year/quarter (brag wrapped 2026 --previous) is an error, since the positional arg already names a bounded period (name the one you want directly, e.g. brag wrapped 2025).
 
-Output is markdown (default) or a single-object JSON envelope (--format json). The digest renders these sections: Cadence (busiest month + per-month counts), Top initiatives, Impact moments, Rhythm (longest streak, top tags, top types), and Span. Filter flags --tag/--project/--type compose with the period.
+Output is markdown (default) or a single-object JSON envelope (--format json). The digest renders these sections: Cadence (busiest month + per-month counts), Top initiatives, Impact moments, What didn't work (work recorded with brag learn, left out when there is none), Rhythm (longest streak, top tags, top types), and Span. Filter flags --tag/--project/--type compose with the period.
 
 Examples:
   brag wrapped                                # the current calendar year, markdown
````

### §6. Documentation

`docs/api-contract.md`:

````diff
diff --git a/docs/api-contract.md b/docs/api-contract.md
index 4e4cca2..3960d94 100644
--- a/docs/api-contract.md
+++ b/docs/api-contract.md
@@ -510,6 +510,15 @@ Document structure:
   `## Impact`, per-project `### <project>` groups (alpha-ASC,
   `(no project)` last; chrono-ASC + ID-tiebreak within group), each
   entry as `- <id>: <title>` followed by an indented `  <impact>` line.
+- **`## What didn't work`** (markdown): the with-impact entries whose
+  `type` is the reserved `failed` that `brag learn` writes, pulled out of
+  `## Impact` and rendered after it in the same grouped shape. Each of the
+  two sections appears only when it has an entry: a window with no
+  recorded failure has no `## What didn't work` heading, and a window
+  holding only failures has no `## Impact` heading. The `Entries:` tally
+  counts both sections. A failure with no impact statement appears in
+  neither, like any other impact-less entry. Locked by
+  [DEC-050](../decisions/DEC-050-a-failure-is-never-rendered-as-a-win.md).
 
 Flags:
 
@@ -545,9 +554,11 @@ Flags:
   [DEC-014](../decisions/DEC-014-rule-based-output-shape.md). Top-level
   keys: `generated_at`, `scope`, `filters`, `entries_in_window`,
   `entries_with_impact`, `counts_by_project` (a `map[string]int` over
-  the with-impact subset, alpha-ASC by key under `encoding/json`), and
+  the with-impact subset, alpha-ASC by key under `encoding/json`; it
+  counts both sections, so it sums to `entries_with_impact`),
   `impact_by_project` (an array of `{project, entries:[...]}` groups in
-  group order). Each entry inside `impact_by_project[].entries` is a
+  group order, failures excluded), and `failures_by_project` (the same
+  shape, holding only the failures). Each entry inside either array is a
   deliberately NARROW 4-key projection `{id, title, project, impact}` —
   NOT DEC-011's 9-key shape — locked by
   [DEC-028](../decisions/DEC-028-impact-digest-window-and-shape.md).
@@ -559,7 +570,7 @@ Flags:
 Empty-window / no-impact: provenance always renders (both tally counts
 `0`); the `## Impact` body is omitted from markdown; JSON renders
 `entries_in_window`/`entries_with_impact` as `0`, `counts_by_project` as
-`{}`, `impact_by_project` as `[]`.
+`{}`, `impact_by_project` and `failures_by_project` as `[]`.
 
 Unknown or missing window flags, or an unknown `--format` value, exit 1
 (user error).
@@ -596,7 +607,14 @@ Document structure:
     excluding `(no project)`.
   - `## Impact moments` — with-impact entries grouped by initiative
     (`### <project>`), each as `- <id>: <title>` plus an indented
-    `  <impact>` line, impact text in full.
+    `  <impact>` line, impact text in full. Failures are excluded.
+  - `## What didn't work` — the with-impact entries typed `failed` (what
+    `brag learn` writes), in the same shape. This is the one section
+    rendered **only when it has an entry**, so a period with no recorded
+    failure has no such heading. Locked by
+    [DEC-050](../decisions/DEC-050-a-failure-is-never-rendered-as-a-win.md)
+    and the Amendment to
+    [DEC-030](../decisions/DEC-030-wrapped-period-selection-and-section-taxonomy.md).
   - `## Rhythm` — `Longest streak: N days`, then `**Top tags**` (top-5)
     and `**Top types**` (top-3) count lists.
   - `## Span` — `- First entry:` / `- Last entry:` (`YYYY-MM-DD`) and
@@ -634,8 +652,10 @@ Flags:
   single-object envelope. Top-level keys: `generated_at`, `scope`,
   `filters`, `total_entries`, `cadence` (`{busiest_month, series:[{period,
   count}]}`), `top_initiatives` (`[{project, count}]`), `impact_moments`
-  (`[{project, entries:[{id, title, project, impact}]}]`),
-  `longest_streak`, `top_tags` (`[{name, count}]`), `top_types`
+  (`[{project, entries:[{id, title, project, impact}]}]`, failures
+  excluded), `failures_by_project` (the same shape, holding only the
+  failures — the key `brag impact` uses too), `longest_streak`,
+  `top_tags` (`[{name, count}]`), `top_types`
   (`[{name, count}]`), `span` (`{first_entry_date, last_entry_date,
   active_days}`). On an empty period arrays are `[]`,
   `busiest_month`/date fields are `null`, numbers `0` — but
````

`docs/tutorial.md`:

````diff
diff --git a/docs/tutorial.md b/docs/tutorial.md
index 8724d6f..229d06a 100644
--- a/docs/tutorial.md
+++ b/docs/tutorial.md
@@ -534,8 +534,10 @@ excluded), `--month --previous` the previous month, `--year --previous`
 last year. It needs a window flag and can't combine with `--since`. Only
 entries with a non-empty impact appear in the body; a
 `<shown>/<in-window> with impact` tally keeps you honest about what was
-left out. Filter flags `--tag`/`--project`/`--type` compose with the
-window. Pipe the JSON form into an LLM to draft the narrative:
+left out. Anything you recorded with `brag learn` gets its own
+`## What didn't work` section after the impact, so a dead end never reads
+as a win; the section only appears when there is one. Filter flags
+`--tag`/`--project`/`--type` compose with the window. Pipe the JSON form into an LLM to draft the narrative:
 
 ```bash
 brag impact --quarter --format json | claude "draft my quarterly impact summary"
@@ -566,7 +568,9 @@ spill into 2027. The digest renders a celebratory arc: **Cadence** (busiest mont
 Unicode block-glyph sparkline `▁▂▃▄▅▆▇█` over the per-month counts, and
 the per-month count series), **Top initiatives** (your top projects),
 **Impact moments** (entries with an impact statement, in full),
-**Rhythm** (longest streak, top tags, top types), and **Span** (first
+**What didn't work** (anything you recorded with `brag learn`, shown only
+when there is some), **Rhythm** (longest streak, top tags, top types),
+and **Span** (first
 and last entry + active days). Filter flags `--tag`/`--project`/`--type`
 compose with the period. The cadence sparkline is on by default in
 markdown; suppress it with `--no-spark` or a `NO_COLOR` env var (it never
````

`AGENTS.md`, §11. Both entries are single long lines:

````diff
diff --git a/AGENTS.md b/AGENTS.md
index 9d66924..9803b3f 100644
--- a/AGENTS.md
+++ b/AGENTS.md
@@ -291,12 +291,12 @@ DECs are stable; specs come and go. DECs don't reciprocally list specs.
 - **Store** — the `*storage.Store` Go type that owns the `*sql.DB` and all typed methods. The only package that imports a SQL driver.
 - **migration** — a single `NNNN_*.sql` file under `internal/storage/migrations/`, embedded into the binary, applied automatically in lexical order on `storage.Open`.
 - **export** — a one-shot dump of entries, either as a Markdown report (stdout or `--out file.md`) or as a portable SQLite file copy (via `VACUUM INTO`).
-- **learn** — `brag learn`: the capture verb for work that did **not** work, and the only place `entries.type` is pinned rather than free-form. Writes the reserved type value `failed` (DEC-049) from flag mode or editor mode; no `--type` flag, no `-k`, no `--json`, and — deliberately — no milestone nudge on stderr, because every milestone line is a congratulation. `entries.type` stays free-form everywhere else: the verb pins one value, it does not validate the field. Retrieval is the existing `brag list --type failed`; the DEC-044 memory line renders `[<project>/failed]`, so a failure is labelled as one on the agent-facing read surface for free. PROJ-008 STAGE-023 (SPEC-085).
+- **learn** — `brag learn`: the capture verb for work that did **not** work, and the only place `entries.type` is pinned rather than free-form. Writes the reserved type value `failed` (DEC-049) from flag mode or editor mode; no `--type` flag, no `-k`, no `--json`, and — deliberately — no milestone nudge on stderr, because every milestone line is a congratulation. `entries.type` stays free-form everywhere else: the verb pins one value, it does not validate the field. Retrieval is the existing `brag list --type failed`; the DEC-044 memory line renders `[<project>/failed]`, so a failure is labelled as one on the agent-facing read surface for free. `brag impact` and `brag wrapped` list a failure that carries an impact under their own `## What didn't work` section, never among the wins (DEC-050, SPEC-086). The constant is `aggregate.FailureType`, read through `aggregate.IsFailure`. PROJ-008 STAGE-023 (SPEC-085).
 - **review** — `brag review --week | --month`: prints recent entries grouped by project followed by three hard-coded reflection questions ("What pattern do you see in this period?", "What did you underestimate?", "What's missing here that should be?"). Markdown elides per-entry descriptions for compactness; JSON includes the full DEC-011 entry shape. Designed to be pasted into an external AI session for guided self-reflection. STAGE-004 (SPEC-019).
 - **summary** — a rule-based (non-LLM) aggregation of entries grouped by project/type over a rolling 7- or 30-day time window (`brag summary --range week|month`). STAGE-004.
 - **stats** — `brag stats`: six lifetime aggregations (total entries, entries/week rolling average, current streak, longest streak, top-5 most-common tags, top-5 most-common projects, corpus span). STAGE-004 (SPEC-020).
 - **tap** — the shared homebrew tap repo (`github.com/jysf/homebrew-tap`) hosting the `bragfile` formula alongside other jysf tools (DEC-040). Originally a per-project tap `homebrew-bragfile` created in STAGE-005; consolidated onto the shared tap at the formula migration.
-- **wrapped** — `brag wrapped [<year>] [Q<n>]`: a shareable, celebratory year- or quarter-in-review digest over a **named** calendar period (default: the current calendar year). The fifth DEC-014 consumer. Unlike `impact`'s `[cutoff, now]` window, `wrapped`'s window is BOUNDED on both ends (`[period-start, next-boundary)`) so a completed year/quarter does not spill past its end. Section arc: Cadence (busiest month + a default-on markdown `sparkline` + per-month `series`) → Top initiatives → Impact moments → Rhythm (longest streak, top tags/types) → Span. PROJ-004 STAGE-013 (SPEC-051), DEC-030; the cadence sparkline is SPEC-052/DEC-031.
+- **wrapped** — `brag wrapped [<year>] [Q<n>]`: a shareable, celebratory year- or quarter-in-review digest over a **named** calendar period (default: the current calendar year). The fifth DEC-014 consumer. Unlike `impact`'s `[cutoff, now]` window, `wrapped`'s window is BOUNDED on both ends (`[period-start, next-boundary)`) so a completed year/quarter does not spill past its end. Section arc: Cadence (busiest month + a default-on markdown `sparkline` + per-month `series`) → Top initiatives → Impact moments → What didn't work (only when non-empty; DEC-050, DEC-030 Amendment) → Rhythm (longest streak, top tags/types) → Span. PROJ-004 STAGE-013 (SPEC-051), DEC-030; the cadence sparkline is SPEC-052/DEC-031.
 - **sparkline** — the Go package `internal/spark/` and its `spark.Line([]int) string` primitive (SPEC-052/DEC-031): a fixed-width Unicode block-glyph string (`▁▂▃▄▅▆▇█`, one glyph per element) computed by min→max linear normalization (`level = round((v-min)/(max-min)*7)`), pure and stdlib-only (`math`, no dependency). `brag wrapped`'s markdown `## Cadence` section renders it as a `Cadence: <glyphs>` line over the zero-filled `cadence.series` counts — default-on, suppressed by `--no-spark` or a present `NO_COLOR` env var (read through the injectable `lookupSparkEnv` package var). Markdown-only: JSON stays raw counts, no glyphs. Empty/flat series → `""`/all `▁`. `stats`/`impact` are deferred (no cadence-series slot). `brag coverage` also renders it — as an `Agent share: <glyphs>` line over the per-month agent share (SPEC-045). PROJ-004 STAGE-013.
 - **--previous** — a boolean window modifier (SPEC-053/DEC-032) shared across `brag impact`, `brag story`, and `brag wrapped` that shifts the selected calendar window from the CURRENT (in-progress) period to the LAST-COMPLETED one, as a BOUNDED `[prev-start, prev-end)` window whose exclusive upper bound is the current period's start (so current-period entries are excluded). The previous-period start is the current-period start shifted back one period via `time.Date`+`AddDate` (never day subtraction; rolls year boundaries). On `impact`/`story` it routes through the extended `windowCutoff` (a `previous bool` param + an exclusive-`end` return; zero `end` = the unchanged current-period path) and requires a window flag on `impact` (a modifier is not a window) while shifting the profile default on `story`; the `scope` echoes `quarter:previous`/`month:previous`/`year:previous`. On `wrapped` it is `now.Year()-1` fed through the existing `parseWrappedPeriod` year path (annual-only; `scope` is the concrete year, e.g. `2025`), valid only with no positional period. INCOMPATIBLE with `--since` (impact/story) or an explicit positional period (wrapped) → `UserError`. PROJ-004 STAGE-013. (`brag coverage` also joins this shared family via `windowCutoff`.)
 - **coverage** — `brag coverage --quarter|--month|--year|--since <date> [--previous]`: a rule-based digest of **provenance share** — how much of the corpus over a calendar window is agent-authored vs human-authored, a per-month trend, and a self-reference (dogfooding) density. The SIXTH DEC-014 consumer (SPEC-045/DEC-033), reusing the shared calendar window (DEC-028) + `--previous` (DEC-032). Body: `## Provenance share` (agent/human counts + %) → `## Monthly trend` (a default-on markdown agent-share `sparkline` over share×100 + a zero-filled per-month `agent / human (%)` series) → `## Self-reference` (entries whose title/description contain `brag`, case-insensitive substring). The agent classifier is **single-sourced** with `brag list --author`: a pure `aggregate.IsAgentAuthored(storage.Entry) bool` predicate kept in agreement with storage's SQL `provenanceExistsClause` by a cross-package drift-guard test (`TestProvenanceClassifier_GoPredicateMatchesSQLClause`); coverage reads all in-window rows once and classifies in Go (it needs BOTH classes for a share), so it does NOT set `ListFilter.Author`. Sparkline is markdown-only (JSON keeps raw counts+shares, DEC-031 choice f), suppressed by `--no-spark`/`NO_COLOR`. PROJ-004 STAGE-013.
````

`CHANGELOG.md`, under `## [Unreleased]` → `### Changed`, before the
`brag memory` entry:

````diff
diff --git a/CHANGELOG.md b/CHANGELOG.md
index 5c4f951..2eadabd 100644
--- a/CHANGELOG.md
+++ b/CHANGELOG.md
@@ -26,6 +26,26 @@ and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0
 
 ### Changed
 
+- **`brag impact` and `brag wrapped` list work that did not work under its
+  own `## What didn't work` heading**
+  ([DEC-050](decisions/DEC-050-a-failure-is-never-rendered-as-a-win.md)).
+  An entry recorded with `brag learn` that carries an impact used to render
+  in `## Impact` and `## Impact moments` exactly like a win. Neither digest
+  shows an entry's type, so a two-day dead end read as an accomplishment in
+  the document built to be shared. It now renders in the new section, after
+  the impact, in the same shape. The section appears only when there is at
+  least one such entry, so a clean quarter has no empty heading. No headline
+  count changes: `impact`'s `Entries: <shown>/<in-window> with impact` still
+  counts both sections.
+- **Breaking: `brag impact --format json` no longer lists failures in
+  `impact_by_project`, and `brag wrapped --format json` no longer lists them
+  in `impact_moments`.** Both envelopes gain a `failures_by_project` key in
+  the same `{project, entries:[{id, title, project, impact}]}` shape. It is
+  always present and `[]` when empty, so one `jq .failures_by_project` works
+  on both. `impact`'s `counts_by_project` still counts both sections and
+  still sums to `entries_with_impact`. A consumer that read every
+  with-impact entry from `impact_by_project` or `impact_moments` now reads
+  both keys.
 - **`brag memory`'s headline count is now `Candidates: <N>`, not
   `Entries: <N>`** ([DEC-048](decisions/DEC-048-provenance-count-names-what-it-counted.md)).
   The number was never the corpus size and never a cap: it is the deduped
````

### §7. `internal/storage/failure_agreement_test.go` (new)

````diff
diff --git a/internal/storage/failure_agreement_test.go b/internal/storage/failure_agreement_test.go
new file mode 100644
index 0000000..452534f
--- /dev/null
+++ b/internal/storage/failure_agreement_test.go
@@ -0,0 +1,70 @@
+package storage_test
+
+import (
+	"path/filepath"
+	"testing"
+
+	"github.com/jysf/bragfile000/internal/aggregate"
+	"github.com/jysf/bragfile000/internal/storage"
+
+	_ "modernc.org/sqlite"
+)
+
+// TestFailureClassifier_GoPredicateMatchesTypeFilter ▲ SPEC-086 — the drift
+// guard for DEC-050's failure predicate, the same shape as
+// TestProvenanceClassifier_GoPredicateMatchesSQLClause in this package.
+//
+// A failure has two definitions in two packages. DEC-049's retrieval path,
+// `brag list --type failed`, is storage's SQL `e.type = ?`. The digests'
+// "What didn't work" section is aggregate.IsFailure in Go. Nothing in storage
+// knows `failed` is special, and that is deliberate (DEC-049 part 3) — but the
+// generic filter is still a definition, and the two must select the same rows,
+// or `brag impact` would list a failure that `brag list --type failed` cannot
+// find. The seeds cover the spellings a well-meaning change would widen the Go
+// side to accept (case, whitespace, the near-synonym), each of which the SQL
+// side rejects on a BINARY-collated column.
+func TestFailureClassifier_GoPredicateMatchesTypeFilter(t *testing.T) {
+	path := filepath.Join(t.TempDir(), "db.sqlite")
+	s, err := storage.Open(path)
+	if err != nil {
+		t.Fatalf("Open: %v", err)
+	}
+	t.Cleanup(func() { _ = s.Close() })
+
+	for _, e := range []storage.Entry{
+		{Title: "failure-with-impact", Type: "failed", Impact: "cost two days"},
+		{Title: "failure-no-impact", Type: "failed"},
+		{Title: "capitalised", Type: "Failed"},
+		{Title: "leading-space", Type: " failed"},
+		{Title: "trailing-space", Type: "failed "},
+		{Title: "near-synonym", Type: "failure"},
+		{Title: "a-lesson", Type: "learned"},
+		{Title: "untyped", Type: ""},
+	} {
+		if _, err := s.Add(e); err != nil {
+			t.Fatalf("add %q: %v", e.Title, err)
+		}
+	}
+
+	// SQL side: the sanctioned retrieval path.
+	sqlFailed, err := s.List(storage.ListFilter{Type: aggregate.FailureType})
+	if err != nil {
+		t.Fatalf("List(Type=%q): %v", aggregate.FailureType, err)
+	}
+
+	// Go side: one unfiltered read, partitioned the way the digests do it.
+	all, err := s.List(storage.ListFilter{})
+	if err != nil {
+		t.Fatalf("List(all): %v", err)
+	}
+	_, goFailed := aggregate.SplitFailures(all)
+
+	if !sameIDSet(idSet(sqlFailed), idSet(goFailed)) {
+		t.Errorf("failure sets differ: SQL=%d Go=%d", len(sqlFailed), len(goFailed))
+	}
+	// Anchor the count so a change that keeps both sides equal-but-wrong still
+	// trips: exactly the two rows typed "failed", nothing else.
+	if len(goFailed) != 2 {
+		t.Errorf("expected 2 failures; got %d", len(goFailed))
+	}
+}
````

### §8. `scripts/test-docs.sh`, Group `AD`

Inserted immediately above `# ===== finalise =====`.

````diff
diff --git a/scripts/test-docs.sh b/scripts/test-docs.sh
index 4487640..24a94f2 100755
--- a/scripts/test-docs.sh
+++ b/scripts/test-docs.sh
@@ -2211,6 +2211,77 @@ $ac2_hits"
     fi
 fi
 
+# ===== Group AD — the digests section failures (SPEC-086 / DEC-050) =====
+#
+# SPEC-086 changes what `brag impact` and `brag wrapped` PRINT: a recorded
+# failure leaves the impact section for its own `## What didn't work`, and
+# both JSON envelopes gain `failures_by_project`. That is visible in every doc
+# that describes the two outputs and invisible to the Go suite, which never
+# reads a doc — the shape SPEC-089 verify found `docs/api-contract.md` stale
+# in. So each place that documents an output shape is asserted to name the
+# change, SCOPED to the right section of its file: an unscoped needle would
+# pass with one of the two commands documented and the other stale.
+#
+# The heading needle carries an apostrophe, so it lives in a double-quoted
+# variable rather than inside a single-quoted literal.
+ad_heading="## What didn't work"
+ad_label="What didn't work"
+
+# ad_section FILE START STOP — FILE's lines from the first line that STARTS
+# with START up to, not including, the next line that starts with STOP.
+# Prefix match via index()==1, so a `####` line never ends a `### ` section.
+ad_section() {
+    awk -v start="$2" -v stop="$3" '
+        !f && index($0, start) == 1 { f = 1; print; next }
+        f && index($0, stop) == 1 { exit }
+        f { print }
+    ' "$1"
+}
+
+# assert_section_names ID TEXT WHERE NEEDLE... — ok when TEXT is non-empty and contains
+# every NEEDLE; otherwise one fail naming each missing needle. An empty TEXT
+# is its own failure: a heading that moved would otherwise make every needle
+# "missing" for the wrong reason, or — for a negative — pass on nothing.
+assert_section_names() {
+    ad_id="$1"; ad_text="$2"; ad_where="$3"; shift 3
+    if [ -z "$ad_text" ]; then
+        fail "$ad_id" "$ad_where: section not found"
+        return 0
+    fi
+    ad_bad=""
+    for ad_needle in "$@"; do
+        printf '%s\n' "$ad_text" | grep -F -q -- "$ad_needle" \
+            || ad_bad="$ad_bad [missing: $ad_needle]"
+    done
+    if [ -z "$ad_bad" ]; then
+        ok "$ad_id"
+    else
+        fail "$ad_id" "$ad_where:$ad_bad"
+    fi
+}
+
+# AD1 / AD2 — the contract documents the section AND the key on BOTH
+# commands. One id per command, two needles each: either alone leaves a
+# reader of the other format wrong.
+assert_section_names "AD1" "$(ad_section docs/api-contract.md '### `brag impact' '### ')" \
+    "docs/api-contract.md, the brag impact section" "$ad_heading" '`failures_by_project`'
+assert_section_names "AD2" "$(ad_section docs/api-contract.md '### `brag wrapped' '### ')" \
+    "docs/api-contract.md, the brag wrapped section" "$ad_heading" '`failures_by_project`'
+
+# AD3 / AD4 — the tutorial tells a user where their failures went, on both
+# commands. It names the section, not the key: the tutorial documents no JSON
+# keys for either command.
+assert_section_names "AD3" "$(ad_section docs/tutorial.md '### Impact by initiative' '### ')" \
+    "docs/tutorial.md, the brag impact section" "$ad_label"
+assert_section_names "AD4" "$(ad_section docs/tutorial.md '### Your year in brags' '### ')" \
+    "docs/tutorial.md, the brag wrapped section" "$ad_label"
+
+# AD5 — the agent-facing glossary states the amended arc, in order. AGENTS.md
+# is what a fresh session reads first, and its `wrapped` entry spells the arc
+# out section by section.
+assert_section_names "AD5" "$(grep -F -- '- **wrapped** —' AGENTS.md)" \
+    "AGENTS.md, the wrapped glossary entry" "Impact moments → $ad_label"
+
 # ===== finalise =====
 
 if [ "$FAIL_COUNT" -gt 0 ]; then
````

### §9. `docs/engineering-practices.md`: regenerate, never hand-edit

Paste `./scripts/inventory.sh`'s output between the markers. At design the
three build-time rows went from 79 / 829 / 200 to **80 / 843 / 205**. The two
DEC rows (52 and 2) and the `:230` sentence are already on this branch.

---

## Build Completion

*Filled at build.*

### Build-phase reflection (3 questions, short answers)

## Reflection (Ship)

- **What can a user do now that they couldn't before?** One sentence,
  before → after. Capture this before closing the cycle.

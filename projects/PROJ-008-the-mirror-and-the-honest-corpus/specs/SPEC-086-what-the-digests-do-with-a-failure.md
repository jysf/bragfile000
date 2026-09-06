---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-086
  type: story                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
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

insight:
  confidence: 0.85                 # up from 0.80. The WHAT is decided by the
                                   # user; Fork B's premise is now measured
                                   # rather than assumed; and the DEC-048
                                   # obligation turned out smaller than framing
                                   # claimed (one count site, not two).

references:
  decisions:
    - DEC-014                      # the envelope, incl. part 4's empty-state rule
    - DEC-028                      # impact's two-number count + the 4-key projection
    - DEC-029                      # story profiles are DATA, not a Go enum
    - DEC-030                      # wrapped's section arc
    - DEC-048                      # a count must name what it counted
    - DEC-049                      # the reserved `failed` value (SPEC-085)
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-085                     # shipped the verb and the value this consumes
    - SPEC-087                     # split out here: Y3 derives instead of caching
---

# SPEC-086: what the celebratory digests do with a failure

> **Cycle: frame.** **GO** at complexity **M**, after splitting the
> `summary`/`story` half out (see **## Complexity**). **Unblocked 2026-09-06** —
> SPEC-085 shipped at `df369e9` (PR #199).
>
> **Every number below was re-derived on 2026-09-06 against `main` at
> `df369e9`.** The corpus is **420** entries, not the 398 this spec was first
> written against and not the 400 the framing prompt carried. Leak measurements
> are taken against a frozen file copy of `~/.bragfile/db.sqlite` (420 entries,
> max id 431, one `type: failed` row) so that every number on this page is
> internally consistent; the live corpus grows during a session.

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

## Acceptance Criteria

*Written at design.* Must be numbers to diff against: the exact `Entries:` /
new-count lines on each touched surface before and after, the line counts of
each changed golden, and the corpus count of `type: failed` entries used in the
fixtures. Note that the fixture count cannot be zero for any assertion that is
supposed to prove the new behaviour — M-1 shows a zero-failure fixture makes the
whole selection change invisible to the suite.

## Failing Tests

*Written at design.* Two standing traps, both earned in this cycle:

- **A NOT-contains assertion proves the old phrase is gone, never that the new
  one is true.** SPEC-085 closed this with paired assertions (its LD7 and
  mutation M-E); follow that shape.
- **A green suite is not evidence.** M-1 survived. Every new assertion must be
  paired with a positive control that fires on a corpus containing a `failed`
  row with an impact.

## Implementation Context

### Decisions that apply
- **DEC-049** — the reserved `failed` value this spec consumes. Read it first.
- **DEC-048** — a count must name what it counted. **The binding constraint** —
  but read Fork A: the obligation is smaller than first written.
- **DEC-028** — `impact`'s two-number count and its 4-key JSON projection.
- **DEC-030** — `wrapped`'s locked section arc. A new section is an **amendment**
  to it, not a free addition.
- **DEC-014** — the envelope, including part 4's empty-state rule (Fork D) —
  which governs the empty *document*, not the empty *section*.
- **DEC-029** — story profiles are data, not a Go enum. Relevant because it is
  the reason the `story` half splits out rather than riding along.

### Prior related work
- **SPEC-085** — the verb, the value, and the `## Verify` reproductions that
  falsified Fork B's original premise.
- **SPEC-084 / DEC-048** — the `Entries:` → `Candidates:` rename. The precedent
  for how this repo renames a count that stopped meaning what it said, and the
  source of the golden-byte-length lesson in Fork E.
- **SPEC-087** — split out of this framing; makes `Y3` derive rather than cache.
  Should land **before** this spec's design, because this spec's design creates
  DEC-050 and would otherwise be the **sixth** consecutive hand re-pin of `Y3`.

### Out of scope
- **`summary` and `story` renderer work.** Split to a successor; DEC-050 still
  states their posture. See **## Complexity**.
- **The `--type` negation silent failure.** Designed around, routed to the stage
  backlog as a `bug`. See **Fork A**.
- **The `Y3`/`X3` derivation.** Split to **SPEC-087**.
- The impact-quality classifier (STAGE-024). **This spec must not smuggle in a
  ranking rule** — sectioning by `type` is not ranking by quality.
- The memory pool fix (`memory-pool-composition-excludes-older-entries`).
- The mirror; story-surface v2.

## Notes for the Implementer

*Written at design.* Run the **§12(b) pre-flight**: SPEC-085's found five things
framing had wrong, two of which would have shipped a guard that could never
fail. This re-framing found six more, listed in *What was re-measured*. Two
mechanical rules that earned their place in this cycle:

- **Quote every `--include` glob.** Unquoted `--include=*.go` is expanded by
  zsh before `grep` runs and the search silently never happens. Prefer
  `grep -rn "x" scripts/ docs/`.
- **Regenerate derived tables; never predict which rows move.** `just inventory`
  only prints — paste it, with no blank lines inside the markers (`X3` compares
  byte-for-byte).

## Build Completion

*Filled at build.*

### Build-phase reflection (3 questions, short answers)

## Reflection (Ship)

- **What can a user do now that they couldn't before?** — one sentence,
  before → after. Capture this before closing the cycle.

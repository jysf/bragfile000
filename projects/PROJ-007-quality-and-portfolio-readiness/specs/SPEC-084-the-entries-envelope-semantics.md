---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-084
  type: bug                        # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-007
  stage: STAGE-022
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-22
  framed_at: 2026-08-22

insight:
  confidence: 0.90

references:
  decisions:
    - DEC-013                      # where `Entries: <N>` was born (export provenance)
    - DEC-014                      # the envelope that never adopted the line
    - DEC-028                      # impact's legislated variant — the pair
    - DEC-043                      # PoolLimit, one-bounded-read-per-list
    - DEC-044                      # memory's redefinition + the Included+Skipped invariant
    - DEC-045                      # MCP resources are byte-identical to `brag memory`
  constraints:
    - one-spec-per-pr
    - test-before-implementation
  related_specs:
    - SPEC-073                     # brag memory
    - SPEC-074                     # MCP resources / brag_memory tool
    - SPEC-082                     # measured the defect and routed it here
---

# SPEC-084: the `Entries:` envelope semantics

> **Cycle: frame.** GO at complexity **S**. This is STAGE-022's fourth
> success criterion and the stage's only remaining correctness item — the
> last spec between here and closing both the stage and PROJ-007.
>
> Framing does **not** settle the fork. It settles what the defect actually
> is, which is not what the backlog entry said it was.

## Context

Six exporters under `internal/export/` emit an `Entries:` provenance line.
Re-measured from source at framing (2026-08-22):

| File:line | Expression |
|---|---|
| `coverage.go:52` | `len(entries)` |
| `impact.go:47` | `len(withImpact)`, `opts.EntriesInWindow` — a pair |
| `markdown.go:89` | `len(entries)` |
| `memory.go:43` | `result.Candidates` |
| `spark.go:97` | `len(entries)` |
| `wrapped.go:54` | `len(entries)` |

`result.Candidates` is the deduped candidate pool (`internal/memory/memory.go:68`).
`PoolLimit = 200` (`internal/memory/memory.go:28`).

### The defect, measured — and it is worse than "a cap mistaken for a corpus"

The backlog entry describes this as *"five use `len(entries)`, and `memory`
alone uses `result.Candidates`, the pool capped at `PoolLimit=200`… the header
reads `Entries: 200`, which a reader takes for the corpus size."* **That
description is incomplete, and the incomplete half is the part that matters.**

Corpus at framing: **387** entries (`brag stats`, 2026-08-22). Measured against
the live corpus with the dev binary built from this branch's `main` (output
byte-identical to released `brag` 0.6.1):

| Command | no filter | `--project bragfile` |
|---|---:|---:|
| `brag export --format markdown` | `Entries: 387` | `Entries: 74` |
| `brag memory` | `Entries: 200` | `Entries: 243` |

**Same word, same corpus, same flag — and the number moves in opposite
directions.** Every sibling's `Entries:` *narrows* when you filter, because
it counts entries in scope. `memory`'s *widens*, because a filter adds a read
to the pool.

The full measured series, all four filter combinations:

- bare → `Entries: 200`
- `--query brag` → `Entries: 232`
- `--project bragfile` → `Entries: 243`
- `--query brag --project bragfile` → `Entries: 247`

So the number is **not a cap** and never was one. `Gather`
(`internal/memory/pool.go:59-89`) performs up to three separate
`PoolLimit`-capped reads — recency, FTS match, project scope — and `Slice`
dedupes their union. `Entries:` on `brag memory` is a **pool-composition
artifact**, bounded by `min(3 × PoolLimit, corpus)`, not by `PoolLimit`.
A fix framed against "200 is a cap" would be false in general — *test the
claim, not the counterexample.*

### `Scope: lifetime` is falsified by the repo's own test

The two lines are adjacent in the rendered document:

```
Scope: lifetime
Entries: 200
```

`Scope: lifetime` is hard-coded (`internal/export/memory.go:41`) and its doc
comment justifies it as *"memory ranks the whole corpus, like stats"*
(`memory.go:14`). It does not. A bare invocation reads
`List(ListFilter{Limit: PoolLimit})` — recency-ordered — so entries older than
the 200th-most-recent are unreachable at any rank. This is not an inference:
`internal/cli/memory_test.go`'s `TestMemoryCmd_ThreeReadsComposeThePool`,
subtest `bare-recency-read-is-capped`, **asserts that unreachability** with two
entries backdated to 2020.

Both header lines are therefore wrong **in the same direction**, and they
reinforce each other: a reader is told the scope is everything and then handed
a number that looks like a total.

### Where the line came from — nobody ever legislated it

This is the root cause, and it is not carelessness.

1. **DEC-013** (markdown export shape) *created* the line: the export
   provenance block is `Exported:`, **`Entries: <N>`**, `Filters:`.
2. **DEC-014** (rule-based output shape) — the envelope every one of these six
   commands inherits — locks only `Generated:`, `Scope:`, `Filters:`.
   **`Entries:` is not in it.** Part 4 even says the empty document *"ends
   after the `Filters:` line."* Every DEC-014 consumer emits an `Entries:` line
   anyway, inherited from DEC-013 by imitation.
3. **DEC-028** (impact) re-specified its own variant:
   `Entries: <shown>/<in-window> with impact`.
4. **DEC-044** (memory) redefined it deliberately and in writing:
   *"`Included + Skipped == Entries` (the deduped candidate-pool size) is an
   invariant"* (DEC-044:173).

So there is no drift to correct and no oversight to blame. There is a **word
with two DEC-level definitions** — DEC-013's *entries in scope* and DEC-044's
*candidate pool* — over an envelope (DEC-014) that never claimed the line and
so has no authority to arbitrate between them. `memory` is the only consumer
that redefined the meaning while keeping the bare word.

Both the code comment (`export/memory.go:29-30`) and the API contract
(`docs/api-contract.md:893-894`, *"the **candidate-pool** size, not the
included count"*) already state the truth. **The only artifact that lies is
the rendered output** — the one place the reader actually is.

### The precedent already in the repo — and the reason it does not transfer for free

`impact.go:47` is not a fifth `len(entries)`. It prints
`Entries: %d/%d with impact` because one number could not carry the meaning,
so it reports a pair with a qualifier. That is an in-repo answer to exactly
this shape of question, and design must be handed it.

**But it is cheap there for a reason that does not hold here.**
`internal/cli/impact.go:130` sets `EntriesInWindow: len(entries)` — both
numbers fall out of **one already-materialized slice**. The denominator is
free.

Memory has no free denominator. There is **no `Store.Count()`** — verified
against every `func (s *Store)` in `internal/storage/` at framing. `brag stats`
gets its 387 by reading the corpus uncapped and taking `len(entries)`
(`export/stats.go:52`). For `memory` to print `of 387` it would need exactly
the unbounded read that `PoolLimit` and DEC-043 sub-decision 5
(*"one bounded read per list"*) exist to prevent — a new storage method or a
deliberate DEC-043 exception, plumbed through both the CLI and the MCP server.

**So the precedent argues for making memory explicit rather than renaming its
label, and simultaneously prices that option well above "follow impact."**
Design owes an honest cost line for it, not an appeal to precedent.

## Goal

Decide what `brag memory`'s `Entries:` line means, make the rendered output say
it, and record the decision at the envelope level so the next DEC-014 consumer
inherits an answer instead of a collision.

## Inputs

- **Files to read:** `decisions/DEC-013` (choice 1), `DEC-014` (choices 3 and 4),
  `DEC-028` (the pair), `DEC-043` (sub-decision 5), `DEC-044` (:146, :173, :251),
  `DEC-045` (byte-identity), `docs/api-contract.md` (:886-905 memory, :693-706
  spark for the contrasting claim).
- **Related code paths:** `internal/export/memory.go`, `internal/memory/pool.go`,
  `internal/memory/memory.go`, `internal/mcpserver/resources.go`,
  `internal/mcpserver/memory.go`.
- **Baseline, all green at framing (2026-08-22):** `gofmt -l .` empty,
  `go vet ./...` clean, `go test ./...` pass, `just lint` **0 issues**,
  `just test-docs` **ALL OK**. Inventory: **187** documentation assertions,
  **46** decision records + 1 reservation (so a new record is **DEC-048**).

## Outputs

*Enumerated at framing, from executed greps — not from the diff imagined.*

- **`internal/export/memory.go:43`** — the markdown header expression.
- **`internal/export/memory.go:89` and `:122`** — the **JSON** envelope's
  `"entries"` key, also fed by `result.Candidates`. **The backlog entry did not
  name this surface.** Whether the JSON key moves with the markdown label is a
  fork, not a given: a renamed JSON key is a harder break than a renamed
  markdown label.
- **`internal/export/memory.go:29-32`** — the doc comment, if the semantics or
  the empty-case sentinel change.
- **A new `DEC-048`** — almost certainly required, because the fix resolves a
  collision between two existing DECs and should bind future DEC-014 consumers.
  Design decides whether it *amends DEC-044* or *legislates `Entries:` at the
  DEC-014 envelope level*. Adding a record moves the inventory's
  `Decision records` row **46 → 47**, which moves **X3** — regenerate with
  `just inventory` and paste between the markers.
- **`decisions/DEC-044.md:173`** — `Included + Skipped == Entries` is stated as
  an invariant against the *label*. If the label changes, this sentence is false
  as written.
- **`docs/api-contract.md:894`** and **`:903`** — the memory provenance
  description and the same invariant restated.
- **`docs/engineering-practices.md:285`** — *status change → planned doc-reference
  update.* It currently names *"the `Entries:` envelope inconsistency is its
  remaining correctness item."* Shipping this makes that sentence false.
- **The stage file** — success criterion 4, the backlog entry, the count line.

### Planned test changes — the premise audit (§9), all three cases

`grep -rn 'Entries:' --include='*_test.go' internal/` returns **45** hits.
Re-derived at framing across **ten** files, not the five the handoff claimed:

| File | Hits | Moves? |
|---|---:|---|
| `internal/aggregate/aggregate_test.go` | 10 | **No** — all ten are `Entries:` Go *struct field* literals |
| `internal/cli/spark_test.go` | 6 | No |
| `internal/export/wrapped_test.go` | 5 | No |
| **`internal/export/memory_test.go`** | **5** | **Yes — 4 assertions + 1 comment** |
| `internal/cli/wrapped_test.go` | 4 | No |
| **`internal/cli/memory_test.go`** | **4** | **Yes — 3 assertions + 1 message** |
| `internal/export/markdown_test.go` | 3 | No |
| `internal/export/coverage_test.go` | 3 | No |
| `internal/cli/export_test.go` | 3 | No |
| `internal/export/impact_test.go` | 2 | No |

**Inversion/removal → planned rewrites.** There are no `.golden` files and no
`testdata/`; expected output is inline string literals, so "byte-exact goldens"
means editing test source. The sites whose premise the change invalidates:

- `internal/export/memory_test.go:45, :70, :95` — `Entries: 8` inside
  `memoryGolden1/2/3`.
- `internal/export/memory_test.go:114` — `Entries: 0` inside `memoryGolden4`.
- `internal/export/memory_test.go:123` — `"entries": 8` in `memoryJSONGolden`.
- `internal/export/memory_test.go:353` — `"entries": 0`.
- `internal/cli/memory_test.go:109, :135` — `Entries: 8`.
- `internal/cli/memory_test.go:591-592` — the `strings.Contains` assertion and
  its failure message.
- `internal/export/memory_test.go:281` — the comment naming the empty-case
  sentinel.

**Thirty-six hits must NOT move**, and `docs/tutorial.md:325` (`Entries: 4`, an
export example) must not either. A change that moves any of them has widened
past the stage criterion.

**Addition → planned count-bump.** If design adds a `test-docs` assertion id,
`Documentation assertions (distinct ids)` moves **187 → 188** and X3's block
must be regenerated. Combined with the DEC row, **two** inventory rows may move
— grep for the *value*, not the idea.

**Self-maintaining, so do not "fix" it:** `internal/mcpserver/` has **zero**
literal `Entries:` assertions.
`TestResourceRecent_IsByteIdenticalToMemoryMarkdown`
(`resources_test.go:154`) compares two renderings of the same function, so it
passes either way. Its passing is **not** evidence the MCP contract is
unaffected — see below.

### Blast radius beyond the diff

- **MCP resource output is an agent-visible contract.**
  `brag://memory/recent` and `brag://memory/project/{name}` are byte-identical
  to `brag memory` by construction (`mcpserver/memory.go:89` calls
  `export.ToMemoryMarkdown`; DEC-045). Changing the header changes what every
  connected agent auto-loads. Design must decide whether that is a breaking
  change and whether anything pins it. `brag_memory`'s **tool description**
  mentions neither `Entries` nor `candidate` (verified), so no description text
  needs updating — only the rendered payload changes.
- **The empty case is load-bearing.** DEC-044:251 and DEC-014 part 4: on an
  empty pool the document **ends after `Entries: 0`** — no `## Slice`, no
  `## Budget` (`export/memory.go:45-47`, pinned by
  `TestToMemoryMarkdown_EmptyCorpusGolden`). Whatever the header becomes, the
  sentinel line and this termination behaviour must be preserved, and design
  must say so explicitly. Note the wrinkle: DEC-014 part 4 as written says the
  empty document ends after the **`Filters:`** line — the implementation ends
  after `Entries: 0`. That gap is itself evidence for the root cause above.

## Acceptance Criteria

*Frame-level. Design tightens these into assertions.*

1. `brag memory`'s rendered header no longer uses the same bare word its five
   sibling exporters use for entries-in-scope — the stage's fourth success
   criterion, met against the **rendered output**, not the doc comment.
2. Whatever the header says is **true under filters**, not just bare. A fix
   validated only against `Entries: 200` has tested the counterexample.
3. `Included + Skipped == <the header's number>` still holds, and is restated
   correctly wherever it is asserted (DEC-044:173, `api-contract.md:903`).
4. The empty case still terminates after the header line with no `## Slice` and
   no `## Budget`.
5. The markdown/JSON relationship is *decided* — either both move or the
   asymmetry is recorded with its reason.
6. `DEC-048` (or an explicit DEC-044 amendment) records the decision, the
   rejected alternatives with reasons, and what binds the *next* DEC-014
   consumer.
7. The five sibling exporters are byte-unchanged; the 36 non-memory `Entries:`
   assertions and `docs/tutorial.md:325` are untouched.
8. `just test`, `just lint` (**0 issues**), `gofmt -l .`, `go vet ./...` all
   clean; `just test-docs` **ALL OK** with the inventory block re-derived and
   pasted if any row moved.

## Forks handed to design

**Framing does not settle these.** It settles that the number is a pool
artifact, not a cap — which reshapes fork 2 and prices it.

1. **What the header should say.** At least four candidates; design owes a
   rejected-alternatives list with a reason on each:
   - **(a) Rename the label** — `Candidates: 200`. Cheapest, and it keeps
     `Entries:` meaning exactly one thing across all six exporters. Costs: a
     word the reader must learn, and the JSON key question.
   - **(b) State the denominator** — `Entries: 200 of 387`. Follows `impact`'s
     precedent, **but the denominator is not free here** (no `Store.Count()`;
     DEC-043 sub-decision 5 forbids the unbounded read). Price it honestly
     before comparing.
   - **(c) Report both** — a `Candidates:` line beside an `Entries:` line.
     Adds a provenance line to an envelope that DEC-014 never legislated, which
     is how this defect started.
   - **(d) Raise or remove `PoolLimit`** — almost certainly wrong; the cap is a
     budget decision (DEC-043). **Write it down as rejected with the reason,
     not unconsidered** — and note that it would not even fix the defect, since
     the number would still be a pool union that *grows* under filters.
2. **Does the JSON `"entries"` key move with the markdown label?**
   (`export/memory.go:89`, `:122`.) Two surfaces, one decision, different
   breakage costs.
3. **Where the decision lives.** A new `DEC-048` legislating `Entries:` at the
   DEC-014 envelope level, versus an `## Amendment` on DEC-044. Note
   `guidance/questions.yaml` carries an **open** question
   `dec-amendment-heading-convention` that this choice would touch — design
   should check whether it answers it, and say so if it does not.
4. **Do the other five need to say anything?** The stage criterion names only
   `memory`. Widening to five working exporters is how a small spec becomes a
   large one. Framing's recommendation: **no** — but the DEC should bind them
   forward, which is free.
5. **Is `Scope: lifetime` in scope?** It is falsified by
   `TestMemoryCmd_ThreeReadsComposeThePool` and it sits one line above the
   defect, reinforcing it. Fixing both together is the honest repair; fixing
   only `Entries:` leaves a false line adjacent to a newly true one.
   **This is the fork that decides S vs M.** If design takes it, say so and
   re-cost. If design defers it, it must be routed as a named follow-up, not
   dropped — a known-false header line is exactly the kind of item this stage
   exists to close.
6. **The second collision, compounding the first: `--project` means two things
   too.** The contrast that makes this defect legible —
   `export --project bragfile` → **74** (down from 387) versus
   `memory --project bragfile` → **243** (up from 200) — is *not* one word
   behaving inconsistently. It is **two** words doing so at once, and both are
   documented:

   | Command | `--project` help text |
   |---|---|
   | `brag export` | *"filter to entries with this project (exact match)"* |
   | `brag memory` | *"boost entries in this project (**a soft boost, not a filter**)"* |

   The boost semantics are deliberate and locked by **DEC-043** — *"a soft
   boost, never a filter"* — and they are exactly why `Gather` adds the third
   read that makes the number grow. So the flag is behaving correctly and the
   header is not.

   **Why this is a fork and not a footnote:** a reader who sees `243` cannot
   tell whether the number rose because `Entries:` means something different
   here or because `--project` does. Repair only the header and that ambiguity
   survives in a subtler form — the number will be correctly labelled and still
   surprising, because the flag that moved it is doing the opposite of what the
   same flag does one command over.

   Design must decide **whether the chosen header wording resolves this or
   merely relabels it**, and say which. Candidate (b) `Entries: N of M` is the
   one most exposed: under a boost there is no meaningful `M`, because nothing
   was filtered out. Testing a fix only against the bare `200` case would miss
   this entirely — *test the claim, not the counterexample.*

   Framing takes no position on changing `--project`'s semantics; DEC-043 locks
   them and this spec must not relitigate a ranking decision. The fork is about
   what the **header** must say to be unambiguous **given** them.

## Verdict

**GO at complexity S.**

The defect is real, measured, and reproduced against the live 387-entry corpus
from a binary built on this branch. The blast radius is one render function,
two envelope surfaces, seven inline test literals in two files, three
documentation artifacts, and one decision record — comparable to SPEC-083.
No ranking behaviour changes and no storage schema moves.

Three things make it non-trivial and design should not treat any of them as
paperwork:

- The number is a **pool-composition artifact**, not a cap. The obvious fix
  ("say 200 is a cap") is false under filters.
- The MCP resources make this an **agent-visible contract change**, not a
  human-facing label edit, and the byte-identity test will stay green through it.
- It resolves a **collision between two DEC-level definitions**, so the record
  outranks the diff — the same shape as SPEC-083, and the reason that one was
  worth its cycle.

**Complexity risk:** fork 5 (`Scope: lifetime`) is the single thing that can
push this to **M**. It belongs in the same repair, and framing's recommendation
is to take it — but design owns that call and must re-cost if it does.

## Implementation Context

*Filled in during the design cycle.*

## Notes for the Implementer

*Filled in during the design cycle.*

## Build Completion

*Filled in during the build cycle.*

- **Deviations from the spec:** <none / list>
- **New DEC-* files created:** <none / list>
- **Constraints checked:** <list>
- **Gates:** <list>

### Build-phase reflection (3 questions, short answers)

- **What was unclear in the spec?**
- **What was missing that you had to decide yourself?**
- **What would you do differently?**

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
   real, greppable result, not a blank. **If this answer is not `none`, capture
   it before closing the cycle.** Evidence ref: tag `pr:<n>`.
   — <answer | none>

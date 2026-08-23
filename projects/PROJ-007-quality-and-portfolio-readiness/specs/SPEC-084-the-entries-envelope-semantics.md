---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-084
  type: bug                        # epic | story | task | bug | chore
  cycle: verify
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
  designed_at: 2026-08-22

insight:
  confidence: 0.93

references:
  decisions:
    - DEC-013                      # where `Entries: <N>` was born (export provenance)
    - DEC-014                      # the envelope that never adopted the line
    - DEC-028                      # impact's legislated variant — the pair
    - DEC-043                      # PoolLimit, one-bounded-read-per-list
    - DEC-044                      # memory's redefinition + the Included+Skipped invariant
    - DEC-045                      # MCP resources are byte-identical to `brag memory`
    - DEC-048                      # NEW — this spec's own record
  constraints:
    - one-spec-per-pr
    - test-before-implementation
  related_specs:
    - SPEC-073                     # brag memory
    - SPEC-074                     # MCP resources / brag_memory tool
    - SPEC-082                     # measured the defect and routed it here
    - SPEC-083                     # the shape this design follows
---

# SPEC-084: the `Entries:` envelope semantics

> **Cycle: design.** GO at complexity **S** — re-costed and **held at S**
> after taking fork 5 (see *Fork 5*). This is STAGE-022's fourth success
> criterion and the stage's only remaining correctness item — the last spec
> between here and closing both the stage and PROJ-007.
>
> **Framing settled what the defect is; design settles all six forks.** Every
> literal below was run through its own tool before it was locked — the
> binary rebuilt on this branch, all four flag combinations and the empty
> corpus measured, both MCP resources driven over a real stdio JSON-RPC
> handshake, and the whole test suite run against the change. See
> [§12(b)](#12b-design-time-pre-flight--what-each-tool-actually-said).
>
> Design found **four things the frame did not have**, and two of them moved
> a fork: the JSON namespace already resolves this collision everywhere except
> `memory` (Finding 1, fork 2); `Scope: lifetime` is the stated premise of a
> DEC-045 sub-decision, so the frame's recommendation would have reopened it
> (Finding 5, fork 5); `Y3` pins the decision count a second time (Finding 3);
> and a **twelfth** literal — a cached golden byte length with no textual link
> to what it caches — that no grep in this spec finds and only running the
> change surfaces (Finding 8).

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

## Outputs (frame)

*Enumerated at framing, from executed greps — not from the diff imagined.*
**Superseded at design** by [*Outputs — re-derived at design*](#outputs--re-derived-at-design),
which adds four sites this list did not name.

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

## Acceptance Criteria (frame)

*Frame-level. Tightened at design into
[*Acceptance Criteria*](#acceptance-criteria) below.*

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


---

## Re-measurement at design (2026-08-22)

*The frame's numbers were taken on 2026-08-22 against a corpus that grows
between sessions. Every one was re-run from a binary built on this branch
before a single decision was locked — a count carried forward from a note is
this repo's most expensive recurring lesson. Prose dates in this spec are the
local working day; the raw tool output below carries its measured UTC stamps,
which fall on 2026-08-23Z.*

`brag stats` → **387** entries. Unchanged since framing, so the frame's series
still holds. Re-measured, not assumed:

| Invocation | `Entries:` | `Included` | `Skipped` | Sum |
|---|---:|---:|---:|---:|
| `brag memory` | **200** | 11 | 189 | 200 |
| `brag memory --query brag` | **232** | — | — | — |
| `brag memory --project bragfile` | **243** | 14 | 229 | 243 |
| `brag memory --query brag --project bragfile` | **247** | — | — | — |
| `brag export --format markdown` | **387** | — | — | — |
| `brag export --format markdown --project bragfile` | **74** | — | — | — |

All four memory values reproduce the frame exactly. `Included + Skipped ==
Entries` holds on both measured invocations. The contrast that names the
defect — `export --project` **387 → 74**, `memory --project` **200 → 243** —
reproduces too.

**What this kills, restated as a rule for the wording below:** any sentence
describing the number as a cap is false the moment a flag is passed. `200` is
not the maximum; `min(3 × PoolLimit, corpus)` is.

---

## §12(b) design-time pre-flight — what each tool actually said

Every literal in this design was run through its own tool before it was
locked. The change was applied to a scratch mutation of
`internal/export/memory.go`, **confirmed present by content hash**
(`17d43e0a…` → `b20c5d9f…`, diff = exactly the three lines intended), measured,
then **restored from a `/tmp` backup** and re-confirmed byte-identical
(`17d43e0a…`). `git checkout` was not used — it discards uncommitted work.

### The post-change output, measured rather than typed

Live corpus, 387 entries, binary built from the mutated tree:

```
brag memory                                  → Candidates: 200
brag memory --query brag                     → Candidates: 232
brag memory --project bragfile               → Candidates: 243
brag memory --query brag --project bragfile  → Candidates: 247
```

Empty corpus, markdown, hexdumped to prove the termination behaviour:

```
# Bragfile Memory

Generated: 2026-08-23T04:01:22Z
Scope: lifetime
Filters: (none)
Candidates: 0
```

`od -c` on the tail: `… e ) \n C a n d i d a t e s :   0 \n` — the document
ends there. No `## Slice`, no `## Budget`. **DEC-014 part 4 and DEC-044:251
are preserved by the rename** (the sentinel is `result.Candidates == 0`, not a
string match), which is Acceptance Criterion 4, verified at design rather than
asserted.

JSON, empty corpus, every key still emitted:

```json
{
  "generated_at": "2026-08-23T04:01:27Z",
  "scope": "lifetime",
  "filters": {},
  "candidates": 0,
  "budget": 2000,
  "estimated_tokens": 0,
  "included": 0,
  "skipped": 0,
  "slice": []
}
```

### The MCP resources, driven over stdio

`brag mcp serve` was driven with a real JSON-RPC handshake
(`initialize` → `notifications/initialized` → `resources/read`). Both
resources, post-change:

```
brag://memory/recent            → Candidates: 200   (mimeType text/markdown)
brag://memory/project/bragfile  → Candidates: 243   (mimeType text/markdown)
```

Pre-change the same two reads returned `Entries: 200` and `Entries: 243`.
**So the defect does reach every connected agent, and so does the fix.** This
is the behavioural surface, not the shape validator (§12(b) refinement): the
byte-identity *test* passes either way, so only actually reading the resource
proves anything.

### The whole cycle was run at design, not just the label mutation

Every literal in *Failing Tests* and *Notes for the Implementer* was written
into the tree, run, and then **restored from `/tmp` with all four content
hashes re-confirmed byte-for-byte**. Three passes:

**Pass 1 — the label mutation alone.** `go test ./...` → **exactly 9 test
functions fail, in exactly 2 packages.** `internal/mcpserver` stays **ok**.

**Pass 2 — fail-first: the tests in place, the implementation untouched.**
`go test ./...` → **exactly 12 test functions fail** (the 9 above plus the 3
new ones), each for the assertion written, none for a compile error. The three
new ones, verbatim:

```
--- FAIL: TestToMemoryMarkdown_HeaderIsCandidatesNotEntries
    memory_test.go:419: provenance still carries an Entries: line: "Entries: 8"
    memory_test.go:426: expected a line "Candidates: 8", got:

--- FAIL: TestResourceRecent_CarriesTheCandidatesHeader
    resources_test.go:395: resource body still carries an Entries: line: "Entries: 3"
    resources_test.go:402: line 6 = "Entries: 3", want "Candidates: 3"

--- FAIL: TestMemoryCmd_CandidateCountGrowsWhenAFlagAddsARead/bare
    memory_test.go:723: provenance still carries an Entries: line: "Entries: 200"
    memory_test.go:733: header = "", want "Candidates: 200"
      …/query_adds_the_match_read      "Entries: 201" → want "Candidates: 201"
      …/project_adds_the_project_read  "Entries: 201" → want "Candidates: 201"
      …/both_reads                     "Entries: 202" → want "Candidates: 202"
```

The four subtests independently re-derive the measured series
**200 / 201 / 201 / 202** from live failure output — the same numbers the
design typed into the table, produced by a different route.

**Pass 3 — implementation applied.** `go test ./...` → **14/14 packages ok**;
`just lint` → **0 issues**; `go vet ./...` → clean; `gofmt -l .` → empty.
`git diff --stat` over `internal/`: four files, +206/−30. And
`git diff --name-only -- internal/export/ internal/aggregate/` returns nothing
outside `export/memory*` — **AC-7 verified at design, not asserted.**

The renamed struct field is gofmt-stable: `Candidates` (10) and
`EstimatedTokens` (15) still set the column, so **not one other line in either
block moves**. Verified with `gofmt -w` and a diff.

Pass 3 did not go green on the first try. What it caught is Finding 8 — a
literal no grep in this spec would have found.

### Finding 1 — the JSON namespace already answered fork 2, in the opposite direction from "leave it alone"

The frame priced fork 2 as *"a renamed JSON key is a harder break than a
renamed markdown label."* True — but it did not look at what `"entries"`
already means in this JSON namespace. Executed at design
(`grep -rn 'Entries .*\`json:' internal/export/*.go`):

| Exporter | markdown line | JSON count key |
|---|---|---|
| `export --format markdown` | `Entries: N` | *(no envelope — the JSON form is a bare array)* |
| `coverage` | `Entries: N` | `total_entries` |
| `impact` | `Entries: X/Y with impact` | `entries_in_window` + `entries_with_impact` |
| `spark` | `Entries: N` | `total.count` |
| `wrapped` | `Entries: N` | `total_entries` |
| **`memory`** | `Entries: N` | **`entries`** |

**`memory` is the only exporter with a bare integer `"entries"` key.** And
`"entries"` *is* already used elsewhere in the same namespace — as an **array
of entry objects**: `impact.go:82`, `review.go:83`, `summary.go:81`,
`wrapped.go:168`.

So `"entries": 200` is wrong twice: wrong meaning, and colliding with the
established array sense of the same key. Renaming it is not "propagating a
markdown decision into JSON at extra cost" — it **removes a pre-existing
collision that the markdown fix alone would leave standing.** Fork 2 settles
for a reason the frame did not have.

### Finding 2 — `Candidates` is not a word the reader must learn; it is the repo's own, on eight surfaces

The frame priced option 1(a) with *"costs: a word the reader must learn."*
Executed at design, that cost is already paid. `candidate` / `Candidates` is
the established term of art for **this exact number**:

| Surface | Text |
|---|---|
| `internal/memory/memory.go:68` | `Candidates      int    // deduped pool size` |
| `internal/memory/memory.go:1` | *"ranks a **candidate pool** of entries"* |
| `internal/memory/memory.go:65` | *"Included+Skipped == **Candidates**"* |
| `internal/export/memory.go:30` | *"`Entries: N` is the **CANDIDATE-POOL** size"* |
| `decisions/DEC-043.md:107` | *"**Candidate pool**: one bounded read per list"* |
| `decisions/DEC-044.md:173` | *"the deduped **candidate-pool** size"* |
| `docs/api-contract.md:895` | *"the **candidate-pool** size, not the included count"* |
| `internal/memory/memory_test.go:353` | `TestSlice_CountsPartitionTheCandidatePool` |

Eight sites across the Go field name, two package/renderer doc comments, two
decision records, the CLI contract, and a test name. **The rendered output is
the only surface in the repo that does not use the word.** Renaming the label
does not introduce a term — it stops the output from being the lone holdout.

### Finding 3 — `Y3` pins the decision count, and the frame's Outputs named only `X3`

The frame wrote: *"Adding a record moves the inventory's `Decision records`
row 46 → 47, which moves **X3**."* Executed at design, `X3` is not the only
guard. `grep -rn "Decision records" scripts/ docs/`:

- `docs/engineering-practices.md:29` — the inventory block (`X3`).
- **`scripts/test-docs.sh:1576`** — `Y3`, a *literal string* pin:
  `grep -F -q 'Decision records | 46 |'`.

This is the identical miss SPEC-083's design caught for itself, and `Y3`'s own
comment says so in the file: *"Note this is the SECOND guard on a number
SPEC-083 moves — Y4 pins the other — which is exactly the pair AGENTS.md §9's
half (b) exists for."* The concept-grep (*"what pins the decision count?"*)
finds `X3`; the value-grep finds both. **N is now 3 for §9 half-(b)'s
one-word sharpening — grep for the value, not the idea** — and it is routed to
the stage close as a promotion candidate rather than re-derived a fourth time.

`X6`'s comment is a third site, prose-only: *"only 1 of 46 DECs carries an
explicit `## Amendment` heading."* Its **numerator** does not move under this
design (see Fork 3); its **denominator** does.

### Finding 4 — DEC-044 restates the invariant twice, not once

The frame's Outputs named `decisions/DEC-044.md:173`. `grep -n 'Entries'
decisions/DEC-044-*.md` returns three:

- **:173** — *"`Included + Skipped == Entries` (the deduped candidate-pool
  size) is an invariant, pinned by a test."*
- **:251** — *"on an **empty candidate pool** the document ends after
  `Entries: 0`"* — a rendering claim; moves with the label.
- **:406** — *"`Included + Skipped == Entries` — pinned by
  `TestSlice_CountsPartitionTheCandidatePool`."* — the Validation section.
  **The frame did not name this one.**

Same class as Finding 3, same remedy: the value-grep found what the
idea-grep missed.

### Finding 5 — `Scope: lifetime` is load-bearing for DEC-045, and the frame's recommendation would have reopened it

The frame recommended taking fork 5 and changing `Scope: lifetime`. Executed
at design, that is a larger move than *"one line above the defect."* Three
artifacts depend on the value being exactly `lifetime`:

1. **`internal/mcpserver/memory_test.go:65-66`** asserts `env.Scope ==
   "lifetime"`. The frame's claim that *"`internal/mcpserver/` has **zero**
   literal `Entries:` assertions… self-maintaining, so do not fix it"* is
   correct for `Entries:` and **false for `Scope:`**. Changing `Scope:` breaks
   a third package that the `Entries:` change leaves untouched — confirmed by
   the pre-flight, where `internal/mcpserver` stayed **ok**.
2. **`internal/mcpserver/memory.go:19`** — `brag_memory` is deliberately built
   *"WITHOUT since/until/day (DEC-045 sub-decision 8) — … a window would make
   the envelope's `Scope: lifetime` line untrue."*
3. **`decisions/DEC-045.md:311`** and **`docs/api-contract.md:931`** state the
   same argument. It is the *stated reason* DEC-045 withheld time-window
   parameters from the MCP tool.

So `Scope: lifetime` is not decorative: it is the premise of a DEC-045
sub-decision. Changing it does not just cost an `M` — it **reopens "should
`brag_memory` take a time window?"**, which this spec must no more relitigate
than it may relitigate DEC-043's soft boost. See Fork 5 for how the fork is
taken anyway.

### Finding 6 — `stats` already carries the correct gloss, one section away

`docs/api-contract.md:392-393`, for `brag stats`:

> `Scope: lifetime` **(hard-coded — stats has no time window)**

`docs/api-contract.md:892-893`, for `brag memory`:

> `Scope: lifetime` **(memory ranks the whole corpus, like `stats`)**

The first is true and says what DEC-014's `Scope:` field actually means. The
second is false *and* mis-cites the first. The repair needs no invention: the
correct gloss is already in the file, 500 lines up, in the voice of the
command being mis-cited.

### Finding 7 — the falsified sentence wraps, so the obvious NOT-contains needle cannot fire

`grep -rn 'ranks the whole corpus' .` returns **zero hits in
`docs/api-contract.md`** — because the phrase spans a line break:

```
892| - **Provenance:** `Generated:` (RFC3339), `Scope: lifetime` (memory ranks
893| the whole corpus, like `stats`), `Filters:` (`--query <text> --project
```

A `U9` written with the natural needle (`memory ranks the whole corpus`) would
have passed **vacuously, forever**, on both the broken and the fixed file.
`grep -F` is line-oriented; a markdown doc is wrapped at 74 columns. The
needle locked below is `(memory ranks`, which is on one line — chosen for that
reason and not for readability.

### Finding 8 — a twelfth literal, found by running the cycle rather than by grepping for it

`internal/export/memory_test.go:247` asserts **`len(got) != 778`** — the byte
length of `memoryGolden1`, cached as a bare integer.

`Candidates:` is three bytes longer than `Entries:`, so the golden becomes
**781** and that assertion fails. **No grep in this spec finds it.** It
contains neither `Entries:` nor `"entries"`; it is a *derived* number with no
textual link to the thing it derives from. It surfaced only because design ran
the implementation and watched `byte length = 781, want 778` come back after
everything else had gone green.

This is AGENTS.md §9's own rule firing on the spec that quotes it: *"when a
literal caches derived output, the derivation outranks the cache — and
enumerate every guard that caches the same value."* Codified at SPEC-082 at
**N=3**; this is the fourth case, and the first where the cached value shares
no token at all with its source. The generalisation worth carrying to the
stage close: half (b) says grep for the value — but **a value you did not know
would move cannot be grepped for, so the backstop is running the change, not a
better grep.**

`grep -rn '\b778\b'` — run once the number was known — finds one further site:
`decisions/DEC-045.md:139`, *"(observed: `size=778` on the wire)"*, a dated
SPEC-074 pre-flight observation proving `Resource.Size` round-trips at all.
**NO CHANGE**: it is evidence about SDK wire behaviour, not a claim about the
current rendering, and the point it makes is unaffected by the byte count.
Named here so build does not "fix" it and verify does not flag it.

---

## Fork 1 — what the header says

**Settled: (a), rename the label to `Candidates:`.**

The number is *how many distinct entries were ranked*. `Candidates:` says
exactly that, it is true in all four flag combinations without qualification,
and it makes no scope claim — so it cannot be read as a corpus total the way
`Entries:` is. Finding 2 removes the cost the frame assigned it: the word is
already the repo's, on eight surfaces; the rendered output is the holdout.

It also gives the five siblings back an unambiguous word. After this change
`Entries:` means *entries in scope* on every surface that emits it, and it
narrows under a filter on every one of them.

### Rejected alternatives

- **(b) State the denominator — `Entries: 200 of 387`.** Rejected on three
  independent grounds, any one sufficient.
  1. **There is no denominator to state.** No `Store.Count()` exists (verified
     against every `func (s *Store)` in `internal/storage/`). `brag stats`
     gets 387 by reading the corpus uncapped (`export/stats.go:52`), which is
     precisely the unbounded read DEC-043 sub-decision 5 exists to prevent —
     *"one bounded read per list… the cap is an implementation bound, not a
     user knob."* Buying the denominator means a new storage method plus a
     deliberate DEC-043 exception, plumbed through the CLI **and** the MCP
     server, in a spec whose subject is a label.
  2. **Under a boost there is no meaningful `M`.** This is fork 6's bite.
     `--project` filters nothing, so "200 of 387" becomes "243 of 387" — where
     243 > the 74 entries actually in the project and the 387 is not a
     population the 243 was drawn *from* under any filter the user applied.
     The form promises a ratio and delivers an artifact.
  3. **The `impact` precedent does not transfer.** `cli/impact.go:130` sets
     `EntriesInWindow: len(entries)` — both numbers fall out of **one already
     materialized slice**, so the pair is free there. Memory materializes a
     *capped* pool and never sees the corpus. Same shape, different price.
- **(c) Report both — a `Candidates:` line beside an `Entries:` line.**
  Rejected: it adds a provenance line to an envelope DEC-014 never legislated,
  which is the exact mechanism that produced this defect (DEC-013's line
  inherited by imitation into six consumers of a DEC that does not contain
  it). Fixing a line-proliferation bug by proliferating a line. It also
  reintroduces the ambiguity it is meant to remove: a reader seeing both
  `Candidates: 243` and `Entries: 74` on one document has to work out which
  the `## Budget` block partitions.
- **(d) Raise or remove `PoolLimit`.** Rejected, and worth the sentence.
  `PoolLimit` is a budget decision with a documented head guarantee (DEC-043:
  *"any entry in the top 26 of any single list cannot be displaced by anything
  the cap excluded"*), and DEC-043's own revisit trigger (d) — *"the pool cap
  becomes observable"* — is about a **caller wanting a bigger slice**, not
  about a mislabelled header. Decisive: **it would not fix the defect.** With
  `PoolLimit = 10000` the number is still a deduped union of up to three
  reads, still grows under flags, and still is not the corpus. Changing a
  ranking constant to repair a word is the wrong instrument.
- **Invented alternatives, also rejected.** `Ranked:` — accurate but a new
  word, where `Candidates:` is the repo's existing one (Finding 2).
  `Pool:` — jargon that names the mechanism rather than the quantity.
  `Considered:` — true, but weaker than `Candidates:` about *what* was
  considered, and unattested anywhere in the repo.

---

## Fork 2 — does the JSON `"entries"` key move with the markdown label?

**Settled: yes. `"entries"` → `"candidates"`.**

Finding 1 is the argument. `memory` is the only exporter with a bare integer
`"entries"`, and `"entries"` in this same namespace already means *an array of
entry objects* on four other exporters. Leaving the key would (1) split one
semantic decision across two surfaces of one document, and (2) preserve a
collision the markdown fix does not touch.

**This is a breaking change to the JSON envelope, and it is named as one**, not
softened: any consumer reading `.entries` from `brag memory --format json` or
from `brag_memory` with `format: "json"` breaks. The repo is pre-1.0 (0.6.1),
the key is 6 weeks old (SPEC-073, PROJ-006), and DEC-011/DEC-014 lock the
envelope *shape*, not the correctness of a payload key a later spec got wrong.
It ships with a CHANGELOG `### Changed` entry under `[Unreleased]` and a
Consequences clause in DEC-048.

### Rejected alternatives

- **Move markdown, leave `"entries"` in JSON.** Rejected: creates a *third*
  meaning rather than removing the second. The document would say
  `Candidates: 243` in markdown and `"entries": 243` in JSON, while
  `"entries"` one command over is a list of objects.
- **Rename to `total_entries`, matching `coverage` and `wrapped`.** Rejected:
  it is not a total. That key means *entries in the window* on both of those
  exporters — the same false claim in a different word.
- **Emit both keys for a deprecation window.** Rejected: a JSON envelope with
  `"entries"` and `"candidates"` carrying the same integer is two contracts,
  and DEC-014 choice 2 pins a flat single-object envelope. There is no
  deprecation machinery in this repo and no evidence of an external consumer;
  inventing one for a 6-week-old key is cost without a beneficiary.

---

## Fork 3 — where the decision lives

**Settled: a new `DEC-048`, legislating the provenance count line at the
DEC-014 envelope level. No `## Amendment` on DEC-044.**

The decision binds **DEC-014 consumers**, not `memory`. Its whole purpose is
that the ninth consumer inherits an answer instead of a collision, which is
not something a `memory`-scoped amendment can do. Three further reasons, each
executed rather than asserted:

1. **`guidance/questions.yaml`'s `dec-amendment-heading-convention` is `open`,
   and SPEC-079 declined the convention deliberately.** Its recorded reason (1)
   is *"it is a process change binding every future DEC, riding in on a docs
   spec."* Adopting `## Amendment` here would be the same move on a **bug**
   spec. **This design does not answer that question** and does not touch its
   `status:` — stated explicitly because the frame asked design to say so
   either way.
2. **The amendment route moves a second inventory row.** `…of those, carrying
   an explicit `## Amendment` section` goes **1 → 2**, and
   `scripts/test-docs.sh:1463`'s comment (*"only 1 of 46 DECs carries an
   explicit `## Amendment` heading"*) becomes false in its numerator as well
   as its denominator. A false comment in a load-bearing script is the exact
   class DEC-047 just spent a cycle repairing.
3. **DEC-044 stays true.** Its `Included + Skipped == Entries` sentences are
   *restated* against the new label (Finding 4: three sites, not one), which
   is a correction of wording, not a reversal of the decision. DEC-044 decided
   the *invariant*; DEC-048 decides the *word*.

### Rejected alternatives

- **`## Amendment` on DEC-044.** Above.
- **Amend DEC-014 in place.** Rejected: DEC-014 is the envelope's founding
  record and six consumers cite it by line. Editing a 2026-04-25 record to
  contain a 2026-08-22 decision destroys the provenance that makes the root
  cause legible — *the line was never in DEC-014* is the finding, and it must
  stay readable in DEC-014.
- **No DEC at all; just fix the code and the docs.** Rejected: the defect is a
  collision between four DEC-level statements (DEC-013 created the line,
  DEC-014 never adopted it, DEC-028 legislated a variant, DEC-044 redefined
  it). Code that is right for reasons nothing records is how the next consumer
  reinvents the collision. Here the record outranks the diff.

---

## Fork 4 — do the other five exporters need to say anything?

**Settled: no code changes to any of the five. DEC-048 binds them forward.**

Their `Entries: N` is `len(entries)` — the number of entries the document
covers. That is what the word means, it narrows under a filter, and it is
correct on all five. Changing them would be a five-command output break for
zero defects, and it is how a small spec becomes a large one.

What DEC-048 *does* do is free: it states the rule they already follow
(`Entries:` = entries in scope, narrows under a filter), records `impact`'s
`Entries: X/Y with impact` as the sanctioned two-number form (DEC-028), and
requires the **next** consumer whose count is not entries-in-scope to name it
something else. Binding forward costs nothing; widening backward costs a cycle.

---

## Fork 5 — is `Scope: lifetime` in scope?

**Taken, and settled at S: the rendered `Scope: lifetime` is CORRECT and stays;
the two artifacts that justify it are FALSE and are repaired here.**

This is not a deferral. The fork asked whether the adjacent header line joins
the repair; it does, and design owns the finding that changes the answer.

**The rendered value is correct in DEC-014's own vocabulary.** DEC-014 choice 3
defines `Scope: <range value>` — the **time window** the document covers.
`brag memory` applies no time window: `Gather` reads
`List(ListFilter{Limit: PoolLimit})` — a *count* bound, not a date bound
(`pool.go:60`). `lifetime` is the settled token for "no window," and `stats`
uses it truthfully for exactly that reason (Finding 6). And per Finding 5 the
value being `lifetime` is the **premise of DEC-045 sub-decision 8** — the
reason `brag_memory` withholds `since`/`until`/`day`.

**What is false is the justification, and it is false in two places:**

- `internal/export/memory.go:14` — *"Scope is always `lifetime` (memory ranks
  the whole corpus, like stats)"*. It does not. `TestMemoryCmd_ThreeReads
  ComposeThePool/bare-recency-read-is-capped` proves an entry at recency rank
  201 is unreachable on a bare invocation.
- `docs/api-contract.md:892-893` — the same sentence, in the contract.

Both are repaired here, using the gloss `stats` already carries one section
away (Finding 6): **`(hard-coded — memory applies no time window)`**. True,
attested, and it preserves the DEC-045 argument verbatim.

**Why this is the honest repair and not the convenient one.** The frame's
worry was that a true `Candidates:` line would sit next to a false
`Scope:` line. After this change neither line is false. The misleading force
of `Scope: lifetime` came from the *pairing* — "everything, and there are 200
of them" — and `Candidates:` dissolves the pairing by no longer claiming to be
a total. What survives is a reader who wants to know whether an old entry
could have been reached; that is a **reachability** question, it is not what
DEC-014's `Scope:` field answers on any of the eight consumers, and DEC-048's
Consequences state it in words plus the renderer's doc comment states it in
code.

**Re-cost: the spec stays S.** Taking the fork this way touches two comments
and one doc sentence. Taking it the other way (changing the rendered value)
breaks `internal/mcpserver/memory_test.go`, falsifies `mcpserver/memory.go:19`,
`DEC-045:311` and `api-contract.md:931`, and reopens a DEC-045 sub-decision —
that is the **M**, and it is an M that relitigates a locked decision.

### Rejected alternatives

- **Change the rendered value to `Scope: pool` / `Scope: ranked` / `Scope:
  recent`.** Rejected: makes `memory` the only DEC-014 consumer whose `Scope:`
  is not a window token, breaking the field DEC-014 *did* legislate in order
  to repair a field it did not. `recent` is additionally false under
  `--project`, where the 2020 entry is reachable.
- **Qualify it in place — `Scope: lifetime (pool capped at 200 per read)`.**
  Rejected twice over: the parenthetical is false under flags (three reads,
  not one cap), and DEC-045 sub-decision 8's argument is stated against the
  literal string `Scope: lifetime`.
- **Defer it as a follow-up spec.** Rejected: the false artifact is two Go/doc
  comments, and repairing them costs three lines. Routing a three-line comment
  repair to a future spec, in a stage whose stated purpose is closing known
  false statements, would be the theatre the stage exists to avoid.

---

## Fork 6 — the second collision: `--project` means two things too

**Settled: the wording RESOLVES the falsehood and RELABELS the surprise. Design
says which, as the frame required, and does not pretend to more.**

DEC-043 sub-decision 4 locks `--project` as *"a soft boost, never a filter"*,
and that is why `Gather` adds the third read that makes the number grow. The
flag is behaving correctly. Design takes no position on changing it.

**What `Candidates:` resolves.** The frame's reader "sees 243 and cannot tell
whether the number moved because `Entries:` means something different or
because `--project` does." Under `Candidates:` there is nothing to
disambiguate, because the label makes **no scope claim at all**. `Entries: 243`
invites exactly one false reading — *243 entries are in this project* — and
that reading is flatly wrong (74 are). `Candidates: 243` says 243 entries were
ranked, which is true, and which is *consistent with* a boost that filtered
nothing out. The specific ambiguity the frame named is gone.

**What it does not resolve, stated plainly.** The header still does not
*explain* why the number grew. A reader who does not know `--project` is a
boost will find `Candidates: 200 → 243` surprising — correctly labelled and
still surprising. That explanation lives where it already lives: the flag help
(*"boost entries in this project (a soft boost, not a filter)"*), the
`Filters:` echo, `docs/for-ai-agents.md` (pinned by `V3`), and now DEC-048's
Consequences. It does not belong in a provenance count line.

Note that candidate 1(b) was the one this fork most exposed — under a boost
there is no denominator, because nothing was filtered out — and 1(b) is
rejected on that ground among others.

### Rejected alternatives

- **Annotate the filters echo — `Filters: --project bragfile (boost)`.**
  Rejected: `Filters:` is a DEC-014-locked field whose contract is *"the
  echoed flag string,"* identical across eight consumers. Diverging one
  consumer's echo to carry semantics is a second envelope change, and it
  would need its own decision — the `M` this spec declined in fork 5, on a
  field with more consumers.
- **Add a prose line explaining the boost to the rendered document.** Rejected
  on the same ground as fork 1(c): the envelope is not the place to
  proliferate lines, and an agent auto-loading `brag://memory/recent` pays for
  every token of it on every session.

---

## Outputs — re-derived at design

*The frame's `## Outputs` above is the frame's enumeration. This one supersedes
it: it is derived from greps executed against the repo at design, and it adds
four sites the frame did not name (Findings 1, 3, 4 and the MCP pin).*

### New files (1)

| Path | What |
|---|---|
| `decisions/DEC-048-provenance-count-names-what-it-counted.md` | The decision record. Embedded verbatim in *Notes for the Implementer* §9. **Must carry `  type: decision`** in its front-matter or `Z7` fails and the inventory silently under-reports. |

### Modified files (10)

| Path | Change |
|---|---|
| `internal/export/memory.go` | Four edits: the markdown label (`:43`), the JSON struct field + tag (`:89`), the envelope assignment (`:122`), and **two** doc comments — `MemoryOptions` (`:13-20`, the false `Scope:` justification) and `ToMemoryMarkdown` (`:27-35`, the count's semantics). |
| `internal/export/memory_test.go` | Seven inline literals, one comment, **one cached byte length** (`778` → `781`, Finding 8), plus **one new test**. |
| `internal/cli/memory_test.go` | Four inline literals, plus **one new test**. |
| `internal/mcpserver/resources_test.go` | **One new test.** Nothing in this package pins the header today — see LD6. |
| `decisions/DEC-044-memory-slice-token-budget-and-line-shape.md` | **Three** sites: `:173`, `:251`, `:406`. The frame named one. |
| `docs/api-contract.md` | Two sites: the memory provenance bullet (`:892-895`, both the `Scope:` gloss and the count) and the invariant (`:903`). |
| `docs/engineering-practices.md` | The STAGE-022 sentence (`:285`) **+ the regenerated inventory block** (three rows move). |
| `scripts/test-docs.sh` | New `U9`; `Y3` re-pinned 46 → 47; `X6` comment denominator 46 → 47. |
| `CHANGELOG.md` | A `### Changed` entry under `[Unreleased]`. The JSON key rename is a breaking change and is named as one. |
| `projects/PROJ-007-quality-and-portfolio-readiness/stages/STAGE-022-measured-and-enforced.md` | Success criterion, the SPEC-084 backlog entry, the count line. |

### Premise audit (§9), run at design against the repo

**Case 1 — inversion/removal → planned test rewrites.** There are no `.golden`
files and no `testdata/` directories; expected output is inline Go string
literals, so "byte-exact goldens" means editing test source.

`grep -rn 'Entries:' --include='*_test.go' internal/` → **45 hits, ten files.**
Executed at design; the frame's re-derivation reproduces exactly.

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

Plus the JSON key, which that grep does **not** find:
`grep -n '"entries"' internal/export/memory_test.go` → `:123`, `:353`. **Two
more literals, in a file the first grep already flagged for other reasons.**

Plus **one that neither grep finds**: `internal/export/memory_test.go:247`
caches `memoryGolden1`'s byte length as `778`. `Candidates:` is three bytes
longer, so it becomes `781`. See Finding 8 — this one was found by running the
change, not by grepping for it, and it is the reason *Order of work* step 1
says run the suite before touching `memory.go`.

**Twelve inline literals move, in two files.**

**Nine test functions fail, measured against the mutation, not predicted:**

| Package | Test | Literal that breaks it |
|---|---|---|
| `export` | `TestToMemoryMarkdown_BlendedGolden` | `memoryGolden1` (`:45`) |
| `export` | `TestToMemoryMarkdown_RecencyGolden` | `memoryGolden2` (`:70`) |
| `export` | `TestToMemoryMarkdown_TightBudgetGolden` | `memoryGolden3` (`:95`) |
| `export` | `TestToMemoryMarkdown_EmptyCorpusGolden` | `memoryGolden4` (`:114`) |
| `export` | `TestToMemoryJSON_BlendedGolden` | `memoryJSONGolden` (`:123`) |
| `export` | `TestToMemoryJSON_EmptyCorpusEmitsEveryKey` | `` `"entries": 0` `` (`:353`) |
| `cli` | `TestMemoryCmd_EndToEndMarkdownGolden` | `memoryCmdGolden1` (`:109`) |
| `cli` | `TestMemoryCmd_BareInvocationIsPlainRecency` | `memoryCmdGolden2` (`:135`) |
| `cli` | `TestMemoryCmd_ProjectIsSoftBoostNotFilter` | `strings.Contains(out, "Entries: 8")` (`:591-592`) |

`internal/mcpserver` reported **ok** through the whole mutation — which is the
gap LD6 closes, not evidence that the contract is unaffected.

**Thirty-six hits must NOT move**, and `docs/tutorial.md:325` (`Entries: 4`, a
`brag export` example) must not either. A change that moves any of them has
widened past the stage criterion.

**Case 2 — addition → planned count-bump, and the value-grep that finds the
second guard.** Two values move. Four guards pin them:

| Value | Decision records 46 → 47 | Doc assertions 187 → 188 |
|---|---|---|
| `X3` (byte-exact inventory block) | ✔ | ✔ |
| `Y3` (literal `'Decision records \| 46 \|'`) | ✔ **← the concept-grep misses this; see Finding 3** | |
| `X6` comment (`1 of 46 DECs`) | ✔ (prose only; the numerator does **not** move) | |
| `X7` word band (2,525 of 1800..2700) | indirect — the practices sentence changes | indirect |

`X7` headroom measured at design: **175 words**. The §285 replacement is
net-neutral to within a few words; if a build-time reword overshoots, the fix
is a shorter sentence, never a wider band.

The `## Amendment` inventory row stays at **1** — DEC-048 carries no
`## Amendment` heading (Fork 3). Verified at design against
`scripts/inventory.sh:49`.

**Case 3 — status change → planned doc-reference update.** Three greps, not
one, and the second and third are the point.

`grep -rn "Entries:" docs/ README.md` → **9 hits.** That grep does **not**
find `api-contract.md:903` (`Included + Skipped == Entries` — no colon) and
does not find `api-contract.md:892` (the `Scope:` gloss — the word `Entries`
is not in it at all). Widening to `grep -rn "Entries" docs/ README.md` adds
`:903` and 11 unrelated uses (section headings, `Entries are packed greedily`,
`Entries live in an embedded SQLite`); the `Scope: lifetime` sweep below is
what surfaces `:892`. All three greps were run at design.

| Hit | Found by | Verdict |
|---|---|---|
| `docs/api-contract.md:894` | `Entries:` | **EDIT** — the memory provenance count. The subject. |
| `docs/api-contract.md:903` | `Entries` (no colon) | **EDIT** — `Included + Skipped == Entries`, false as written after the rename. |
| `docs/api-contract.md:892` | `Scope: lifetime` sweep | **EDIT** — the `Scope:` gloss (Fork 5). Invisible to both `Entries` greps. |
| `docs/engineering-practices.md:285` | `Entries:` | **EDIT** — *"the `Entries:` envelope inconsistency is its remaining correctness item"* stops being true. |
| `docs/api-contract.md:285` | `Entries:` | **NO CHANGE** — DEC-013's `brag export` provenance block. Still `Entries:`, still correct. |
| `docs/api-contract.md:347` | `Entries:` | **NO CHANGE** — the `## Entries` *heading* in a `brag export` document, unrelated. |
| `docs/api-contract.md:451` | `Entries:` | **NO CHANGE** — `impact`'s `Entries: <shown>/<in-window> with impact`, DEC-028's sanctioned variant. |
| `docs/api-contract.md:530`, `:625`, `:701` | `Entries:` | **NO CHANGE** — `wrapped`, `coverage`, `spark`. Each *"a headline `Entries: N` count"*, each true. `:701` even states the contrast this spec relies on: *"(the whole in-window corpus)"*. |
| `docs/tutorial.md:325` | `Entries:` | **NO CHANGE** — a `brag export` sample document. Must stay `Entries: 4`. |
| 11 further `Entries` hits across `docs/` + `README.md` | `Entries` (no colon) | **NO CHANGE.** `## Entries (chronological)` headings (DEC-013's `--flat` wrapper), `- Entries mentioning brag/bragfile: N` (coverage's self-reference row), and ordinary prose uses of the noun. None is a claim about a headline count. |

Repo-wide, outside `docs/`:

| Hit | Verdict |
|---|---|
| `decisions/DEC-044.md:173`, `:251`, `:406` | **EDIT** — three, not one (Finding 4). |
| `decisions/DEC-013.md:42`, `:71` | **NO CHANGE.** The historical record of where the line came from; editing it would erase the root cause. |
| `decisions/DEC-014.md` | **NO CHANGE.** It contains no `Entries:` — that absence *is* the finding (Fork 3's rejected "amend DEC-014 in place"). |
| `decisions/DEC-028.md:76`, `:145`; `DEC-030.md:140` | **NO CHANGE.** `impact` and `wrapped`, both correct. |
| `decisions/DEC-045.md:311` | **NO CHANGE** — the `Scope: lifetime` argument, preserved intact by Fork 5's resolution. |
| `internal/export/{coverage,markdown,spark,wrapped,impact}.go` | **NO CHANGE.** Fork 4. |
| `internal/export/{review,summary,impact,wrapped}.go` (`json:"entries"` **arrays**) | **NO CHANGE.** Different key, different type, correct meaning (Finding 1). |
| `internal/memory/pool.go:89` (`Pool{Entries: …}`) | **NO CHANGE.** A Go struct field on an internal type, not an output label. |
| `projects/PROJ-007-.../brief.md:113` | **NO CHANGE in this spec.** *"STAGE-022 … `[ ]` *(proposed)*"* — a forward-looking plan line describing the stage's scope, not a present-tense status claim. It moves when the stage's checkbox flips, at **STAGE-022 close**. Routed there, not dropped. |
| `projects/PROJ-008-.../brief.md:53` | **NO CHANGE.** Quotes `impact`'s header, which is correct. |
| `projects/PROJ-001-mvp/stages/STAGE-003.md:304`, `STAGE-004.md:366` | **NO CHANGE.** Dated planning records for shipped stages. |
| `projects/**/specs/done/**` | **NO CHANGE.** Archived. |
| `NEXT-SESSION-PROMPT.md` | **NO CHANGE.** Session handoff scratch, not repo content. |

**Case 3, second sweep — `Scope: lifetime`.** Fork 5 required its own audit,
because an `Entries:` grep cannot find it. `grep -rn 'Scope: lifetime\|"scope":
"lifetime"'` → 24 hits.

| Hit | Verdict |
|---|---|
| `internal/export/memory.go:14` | **EDIT** — the false justification. |
| `docs/api-contract.md:892` | **EDIT** — the same false sentence in the contract. |
| `internal/export/memory.go:41`, `:120` | **NO CHANGE.** The rendered value is correct (Fork 5). |
| `internal/mcpserver/memory_test.go:65-66` | **NO CHANGE.** Asserts `env.Scope == "lifetime"` — a test the *other* resolution of Fork 5 would have broken. |
| `internal/mcpserver/memory.go:19`, `DEC-045.md:311`, `docs/api-contract.md:931` | **NO CHANGE.** The DEC-045 sub-decision 8 argument, preserved verbatim. |
| `internal/export/stats.go:19`, `:35`, `:117`; `docs/api-contract.md:392` | **NO CHANGE.** `stats` — and `:392` is the *correct* gloss this design copies (Finding 6). |
| `internal/{cli,export}/{stats,memory}_test.go` (8 hits) | **NO CHANGE** as `Scope:` lines; the memory goldens change on their `Candidates:` line only. |

### §12 NOT-contains self-audit

Two assertions in this spec are negative. Both needles were grepped against
this design's own load-bearing literals at design time.

| Assertion | Forbidden token | Where it is checked | Present in the new literal? |
|---|---|---|---|
| `TestToMemoryMarkdown_HeaderIsCandidatesNotEntries` | a rendered line with prefix `Entries:` | the bytes `ToMemoryMarkdown` emits | **no** — the four `Fprint*` literals in the locked implementation are `"# Bragfile Memory"`, `"Generated: %s\n"`, `"Scope: lifetime"`, `"Filters: %s\n"`, `"Candidates: %d\n"`, `"## Slice"`, `"## Budget"` and four `- <field>:` lines. None begins `Entries:`. |
| `U9` | `(memory ranks` | `docs/api-contract.md` | **no** — the locked replacement reads `(hard-coded — memory applies no time window)`. |

**Scope note, and the reason it matters here.** The Go NOT-contains reads
*rendered output*, which includes each entry's `Line`. A fixture entry whose
title began `Entries:` would false-positive. Audited: the eight SPEC-073
fixture rows are fixed and known (`internal/export/memory_test.go:27-34`);
none contains the string. The `internal/cli/memory.go` cobra `Long` **does**
contain the word *"Entries"* (*"Entries are packed greedily in rank order"*) —
that is help text, it never reaches `ToMemoryMarkdown`'s output, and **no
assertion in this spec reads help output.** Called out because a NOT-contains
pointed at `--help` would fire on it.

**`U9`'s negative needle is deliberately not the readable phrase.** See
Finding 7: `memory ranks the whole corpus` spans a line break in the wrapped
markdown, so `grep -F` never matches it and the assertion would pass
vacuously on the broken file. Confirmed at design — that grep returns zero
hits in `docs/api-contract.md` today, while the false sentence is sitting
there.

---

## Locked design decisions

**LD1 — `brag memory`'s markdown provenance count is `Candidates: <N>`.**
Exact literal in Notes §1. The five sibling exporters are byte-unchanged and
keep `Entries:`. Fork 1(a).

**LD2 — the JSON envelope key is `"candidates"`, and the Go field is
`Candidates`.** One decision, both surfaces. Named as a **breaking change** to
the `brag memory --format json` / `brag_memory` envelope; recorded in
CHANGELOG and in DEC-048 Consequences. Fork 2.

**LD3 — `Included + Skipped == Candidates` is restated at all three sites in
DEC-044 (`:173`, `:251`, `:406`) and at `docs/api-contract.md:903`.** The
invariant does not change; the word it is stated against does. Fork 3 /
Finding 4.

**LD4 — the decision is a new `DEC-048` legislating the DEC-014 provenance
count line, not an `## Amendment` on DEC-044.** DEC-048 binds all DEC-014
consumers forward. It does **not** answer
`guidance/questions.yaml`'s open `dec-amendment-heading-convention`, does not
touch that entry's `status:`, and carries no `## Amendment` heading — so the
inventory's amendment row stays at 1. Fork 3.

**LD5 — the rendered `Scope: lifetime` does not move; the two artifacts that
justify it do.** `internal/export/memory.go:14` and `docs/api-contract.md:892`
lose *"memory ranks the whole corpus, like stats"* and gain
*"(hard-coded — memory applies no time window)"*, the gloss `stats` already
carries at `api-contract.md:392`. `internal/export/memory.go:41` and `:120`,
`internal/mcpserver/memory_test.go:65-66`, `internal/mcpserver/memory.go:19`,
`DEC-045:311` and `api-contract.md:931` are **untouched**. Fork 5.

**LD6 — the MCP resource payload gets its first assertion.** Today nothing in
`internal/mcpserver` would fail if the header regressed: the byte-identity
test compares two renderings of the same function, and it stayed green through
the entire design mutation. `TestResourceRecent_CarriesTheCandidatesHeader`
pins the agent-visible bytes directly. This is the answer to the frame's
question *"whether that is a breaking change and what, if anything, pins it"*:
it is a payload change on an agent-visible contract, the URIs / names /
MIME types / tool schemas are unchanged, and after this spec **one assertion**
pins it instead of zero.

**LD7 — the claim is tested, not the counterexample.**
`TestMemoryCmd_CandidateCountGrowsWhenAFlagAddsARead` proves the number is a
pool artifact that **grows** under flags (200 / 201 / 201 / 202 over a
202-entry corpus) and that it is **budget-independent** (identical at
`--budget 100` and `--budget 20000`, both measured at design). A fix validated
only against the bare `200` would pass a test that cannot distinguish a cap
from a pool.

**LD8 — the header assertions are line-based, never `strings.Contains`.**
`strings.Contains(out, "Entries: 8")` is false against `Candidates: 8`, so a
Contains-shaped guard *looks* sufficient; it is not, because any future
two-line or suffixed variant slips past it in both directions. Line equality
and `strings.HasPrefix` per line, per AGENTS.md §9's heading-level addendum.

**LD9 — `U9` carries both polarities, and its negative needle is
line-scoped.** Positive: `docs/api-contract.md` names `` `Candidates: N` ``.
Negative: it no longer contains `(memory ranks`. Finding 7 is why the needle
is that fragment and not the readable sentence.

**LD10 — `Y3` is re-pinned 46 → 47 and `X6`'s comment denominator moves with
it.** `Y3` pins `Decision records | 46 |` as a literal string; `X6`'s comment
says *"only 1 of 46 DECs."* The numerator stays 1 (LD4). A re-pin note in the
style `Y3`/`Y4` already use records the move as deliberate, not drift.

**LD11 — the practices-page sentence and the inventory block change in the
same edit; the block is regenerated, never hand-edited.** `X3` diffs it
byte-for-byte against `scripts/inventory.sh`. Three rows move.

### Rejected alternatives (build-time)

- **Hand-editing the two moved rows in the inventory block** instead of
  running `just inventory`. `X3` recomputes the whole table and fails with a
  full diff. Run the script; paste between the markers.
- **Widening `X7`'s word band** if the §285 edit overshoots. It does not
  (2,525 of 2,700 measured; the edit is net-neutral). If a build-time reword
  pushed it over, the fix is a shorter sentence — bumping a band to make a
  page fit is the disarming SPEC-079 LD5 refused.
- **`sed -i` across `internal/` for `Entries:` → `Candidates:`.** Rejected
  explicitly: 36 of the 45 hits must **not** move, and 10 of them are Go
  struct-field literals in `internal/aggregate`. Edit the enumerated files by
  hand; the enumeration is in Outputs for exactly this reason.
- **Renaming `memory.Result.Candidates`** to match some new word. It is
  already the right word (Finding 2) and it is what this design renames the
  output *to*. Nothing about the `internal/memory` package changes.
- **Adding a `Store.Count()` "while we are here."** Rejected — fork 1(b)'s
  cost, and it needs a DEC-043 exception, not a drive-by.
- **Touching `guidance/questions.yaml`.** Rejected: LD4. This spec neither
  answers nor advances `dec-amendment-heading-convention`, so the register
  does not move and `Y4` stays at 19/6.

---

## Acceptance Criteria

*Frame criteria 1–8 tightened into checks.*

1. `brag memory`'s markdown provenance count line is exactly
   `Candidates: <N>`; **no** rendered line begins `Entries:`. The five
   sibling exporters are byte-unchanged and still emit `Entries:`.
2. The number is correct **under every flag combination**, proven by
   `TestMemoryCmd_CandidateCountGrowsWhenAFlagAddsARead` (200 / 201 / 201 /
   202) and by its budget-independence sub-assertion — not by the bare case
   alone.
3. `Included + Skipped == Candidates` holds and is restated correctly at
   `DEC-044:173`, `:251`, `:406` and `docs/api-contract.md:903`.
4. The empty case still terminates after the header line: `Candidates: 0`,
   no `## Slice`, no `## Budget`. JSON still emits every key with
   `"candidates": 0`.
5. The JSON envelope key is `"candidates"`; **no** `"entries"` key remains in
   `memoryEnvelope`. The change is in `CHANGELOG.md` under `[Unreleased]`
   as a breaking change.
6. `decisions/DEC-048-provenance-count-names-what-it-counted.md` exists,
   carries `  type: decision`, states the decision, the four rejected
   alternatives with reasons, and what binds the next DEC-014 consumer.
7. The 36 non-memory `Entries:` test assertions and `docs/tutorial.md:325` are
   untouched. `git diff --stat` shows no `internal/export/{coverage,markdown,
   spark,wrapped,impact}.go` and no `internal/aggregate/`. Verified at design:
   the whole `internal/` diff is four files, +206/−30.
   `internal/export/memory_test.go:247`'s length assertion still **exists** and
   reads `781` — re-derived, not deleted.
8. `just test`, `just lint` (**0 issues**), `gofmt -l .` (empty),
   `go vet ./...` all clean; `just test-docs` **ALL OK** at **188** assertion
   ids, inventory block regenerated and pasted (`46 → 47`, `187 → 188`).
9. `./scripts/inventory.sh` and the content between the
   `inventory:begin` / `inventory:end` markers are byte-identical (`X3`).
10. The MCP resource `brag://memory/recent` carries the new header, pinned by
    `TestResourceRecent_CarriesTheCandidatesHeader`, and
    `TestResourceRecent_IsByteIdenticalToMemoryMarkdown` still passes.

---

## Failing Tests

*Every expected value below was measured at design against a real run, not
typed. Write these first, run `go test ./...` once, confirm each fails for the
**stated** reason, then implement.*

### Changed — nine existing test functions (12 inline literals)

| File:line | Now | Becomes |
|---|---|---|
| `internal/export/memory_test.go:45` | `Entries: 8` | `Candidates: 8` |
| `internal/export/memory_test.go:70` | `Entries: 8` | `Candidates: 8` |
| `internal/export/memory_test.go:95` | `Entries: 8` | `Candidates: 8` |
| `internal/export/memory_test.go:114` | `Entries: 0` | `Candidates: 0` |
| `internal/export/memory_test.go:123` | `"entries": 8,` | `"candidates": 8,` |
| `internal/export/memory_test.go:281` | comment: *"ends after `Entries: 0`"* | *"ends after `Candidates: 0`"* |
| `internal/export/memory_test.go:353` | `` `"entries": 0` `` | `` `"candidates": 0` `` |
| `internal/cli/memory_test.go:109` | `Entries: 8` | `Candidates: 8` |
| `internal/cli/memory_test.go:135` | `Entries: 8` | `Candidates: 8` |
| `internal/cli/memory_test.go:591` | `strings.Contains(out, "Entries: 8")` | `"Candidates: 8"` |
| `internal/cli/memory_test.go:592` | failure message | matching text |
| `internal/export/memory_test.go:247-248` | `len(got) != 778` / `want 778` | `781` — **the cached golden byte length** (Finding 8). Not an `Entries:` hit, not an `"entries"` hit. Fails **after** everything else is green, with `byte length = 781, want 778`. |

The full post-change golden bodies are in *Notes for the Implementer* §2 —
captured from the failing tests' own `--- got ---` output at design, so build
transcribes measured bytes rather than retyping eight-line documents.

### New — `internal/export/memory_test.go`

**`TestToMemoryMarkdown_HeaderIsCandidatesNotEntries`** — LD1 + LD8. Renders
the SPEC-073 fixture; asserts one line equals `Candidates: 8` and that **no**
line has prefix `Entries:`. *Fails before the change with*: `provenance still
carries an Entries: line: "Entries: 8"` **and** `expected a line
"Candidates: 8"`.

### New — `internal/cli/memory_test.go`

**`TestMemoryCmd_CandidateCountGrowsWhenAFlagAddsARead`** — LD7, the claim
test. Corpus: 200 fillers (`filler 1`…`filler 200`, all current) plus two
entries backdated to 2020 — one in project `orbit`, one titled
`ancient auth note`. Neither backdated entry is reachable by the bare recency
read, so each flag adds exactly one. Run at `--budget 100`:

| Flags | Header | Included | Skipped | Sum |
|---|---|---:|---:|---:|
| *(none)* | `Candidates: 200` | 11 | 189 | 200 |
| `--query auth` | `Candidates: 201` | 11 | 190 | 201 |
| `--project orbit` | `Candidates: 201` | 10 | 191 | 201 |
| both | `Candidates: 202` | 10 | 192 | 202 |

**All four measured at design**, and re-measured at `--budget 20000` to prove
budget-independence (headers identical: 200 / 201 / 201 / 202). The test
asserts the four header lines, `Included + Skipped == N` in each case, and
`Skipped > 0` so the invariant is not satisfied trivially. *Fails before the
change with* four `header = "Entries: 200", want "Candidates: 200"`-shaped
errors.

### New — `internal/mcpserver/resources_test.go`

**`TestResourceRecent_CarriesTheCandidatesHeader`** — LD6. Reads
`brag://memory/recent` over the in-process client with the standard
three-entry seed and frozen clock; asserts line 6 equals `Candidates: 3` and
no line has prefix `Entries:`. Measured at design: today that line is
`Entries: 3`. *Fails before the change with*: `line 6 = "Entries: 3", want
"Candidates: 3"`.

### New — `scripts/test-docs.sh`, `U9`

Both polarities (LD9). *Fails before the change with*: `docs/api-contract.md:
[memory provenance does not name Candidates: N] [the falsified 'memory ranks
the whole corpus' gloss is still present]`.

### Changed — `Y3`

Re-pin `'Decision records | 46 |'` → `| 47 |` and `decision-records!=46` →
`!=47`, plus a re-pin note. *Fails before `DEC-048` is added with*:
`inventory.sh row value(s) wrong: decision-records!=47`.

### Changed — `X3`

No script change; the **inventory block on the practices page** must be
regenerated. Three rows move (`46 → 47`, `187 → 188`, `812 → 815`). Fails with a full
script-vs-page diff until `just inventory` is run and the output pasted.

### Changed — `X6`

Comment only: `1 of 46 DECs` → `1 of 47 DECs`. No assertion behaviour changes;
DEC-048 carries no `## Amendment` heading, so *"only 1"* still holds.

### Mutation checks (run by build, recorded in Build Completion)

All five ran at design. Build re-runs them and pastes the output.

| # | Mutation | Expected |
|---|---|---|
| **M-A** | revert `memory.go:43` to `"Entries: %d\n"` | `TestToMemoryMarkdown_HeaderIsCandidatesNotEntries` fails on **both** halves; the four goldens fail; `TestResourceRecent_CarriesTheCandidatesHeader` fails |
| **M-B** | revert the JSON tag to `json:"entries"` | `TestToMemoryJSON_BlendedGolden` + `TestToMemoryJSON_EmptyCorpusEmitsEveryKey` fail; markdown tests stay green (proves the two surfaces are independently pinned) |
| **M-C** | delete the `result.Candidates == 0` early return | `TestToMemoryMarkdown_EmptyCorpusGolden` fails with `## Slice` / `## Budget` present — the DEC-014 part 4 sentinel |
| **M-D** | feed `result.Included` instead of `result.Candidates` to the header | `TestMemoryCmd_CandidateCountGrowsWhenAFlagAddsARead` fails on all four rows — the counterexample-vs-claim guard |
| **M-E** | restore `(memory ranks` in `docs/api-contract.md` | `FAIL: U9 … [the falsified 'memory ranks the whole corpus' gloss is still present]` |

**Confirm every mutant actually mutated, by content hash.** `shasum -a 256`
the file before and after; a red with an unmoved hash is a red for the wrong
reason, and `git diff --quiet` does **not** substitute — it is blind to
untracked files. Restore from a `/tmp` backup, never `git checkout` (it
discards uncommitted work), and confirm the hash returns to its pre-mutation
value. This protocol was exercised end-to-end during this design's §12(b)
pre-flight: `17d43e0a…` → `b20c5d9f…` → `17d43e0a…`.

### Decision-to-test mapping (§9)

Every locked decision has an assertion that fails without it.

| Decision | Test |
|---|---|
| LD1 markdown label | `TestToMemoryMarkdown_HeaderIsCandidatesNotEntries`; the four markdown goldens; M-A |
| LD2 JSON key + Go field | `TestToMemoryJSON_BlendedGolden`, `TestToMemoryJSON_EmptyCorpusEmitsEveryKey`; M-B |
| LD3 invariant restated | `U9` (the contract half); `TestMemoryCmd_CandidateCountGrowsWhenAFlagAddsARead`'s `Included + Skipped == N` check (the behavioural half); `TestSlice_CountsPartitionTheCandidatePool` unchanged |
| LD4 new DEC-048, no amendment | `Y3` (the count moves), `Z7` (it must be `type: decision`), `X3` (the amendment row stays 1), `Y4` (unchanged 19/6 — the register did **not** move) |
| LD5 `Scope:` value stays, gloss repaired | `U9`'s negative needle; `internal/mcpserver/memory_test.go:65` **still passing** is the guard that the value did not move; M-E |
| LD6 MCP payload pinned | `TestResourceRecent_CarriesTheCandidatesHeader`; M-A |
| LD7 claim not counterexample | `TestMemoryCmd_CandidateCountGrowsWhenAFlagAddsARead`; M-D |
| LD8 line-based assertions | both new Go tests use line equality / `HasPrefix`; no new `strings.Contains` on a header |
| LD9 `U9` both polarities | `U9`; M-E |
| LD10 `Y3` re-pin + `X6` comment | `Y3` |
| LD11 practices sentence + block | `X3`, `X6`, `X7`, `Z5` |

---

## Implementation Context

### Decisions that apply

- **DEC-048** (this spec creates it) — the subject. Read it first; the forks
  above are its reasoning.
- **DEC-014** (rule-based output shape) — the envelope. Choice 3 locks
  `Generated:` / `Scope:` / `Filters:` and **not** `Entries:`; choice 4 governs
  the empty document. The absence is the root cause; do not "fix" DEC-014.
- **DEC-013** (markdown export shape) — where `Entries: <N>` was born, for
  `brag export`, correctly. Historical; unchanged.
- **DEC-028** (impact) — the sanctioned two-number variant
  (`Entries: X/Y with impact`). Unchanged, and cited by DEC-048 as the
  precedent for *naming* a count rather than overloading a word.
- **DEC-043** (rank fusion) — sub-decision 5 is why the number is a union of
  three capped reads; sub-decision 4 is why `--project` grows it. **Neither is
  relitigated.** No ranking behaviour, constant, or read composition changes.
- **DEC-044** (token budget) — the invariant. The invariant does not change;
  three sentences that state it against the old word do.
- **DEC-045** (MCP push surface) — byte-identity makes this an agent-visible
  contract change, and sub-decision 8 is why `Scope: lifetime` must stay
  exactly that string (LD5).

### Constraints that apply

- `test-before-implementation` — the goldens and all three new tests are
  written and observed failing before `memory.go` is touched.
- `one-spec-per-pr` — one branch, one PR.
- `no-sql-in-cli-layer` — untouched; no SQL is added anywhere.
- `stdout-is-for-data-stderr-is-for-humans` — untouched; no new output stream.

### Prior related work

- **SPEC-073 / DEC-043 / DEC-044** built `brag memory` and chose the word.
- **SPEC-074 / DEC-045** made the same bytes an MCP resource, which is what
  turns a label into a contract.
- **SPEC-082** measured the defect and routed it here.
- **SPEC-083** is the shape this spec follows: a defect whose repair is
  primarily a decision record, with the enforcement mutation-checked.

### Out of scope (for this spec specifically)

- The five sibling exporters' output (Fork 4).
- `--project`'s semantics anywhere (DEC-043 sub-decision 4).
- `PoolLimit`'s value (Fork 1(d)).
- A `Store.Count()` or any new storage method (Fork 1(b)).
- The rendered `Scope:` value and `brag_memory`'s parameter set (Fork 5 /
  DEC-045 sub-decision 8).
- `guidance/questions.yaml`'s `dec-amendment-heading-convention` (LD4).
- Anything under `internal/aggregate/` — its ten `Entries:` hits are Go struct
  fields.

---

## Notes for the Implementer

### Order of work

1. **Tests first.** Change the 12 inline literals (§2, §3) — including the
   cached byte length at `internal/export/memory_test.go:247-248`, which is
   the one no grep in this spec finds — and add the three new tests
   (§2, §3, §4). Run `go test ./...` **once** and confirm **12 test functions**
   fail — the nine from the premise audit plus the three new ones — each for
   the assertion you wrote, not a compile error.
2. **`internal/export/memory.go`** (§1). Re-run `go test ./...`; all green.
3. **`gofmt -l .`** must be empty. The struct alignment is unchanged
   (`Candidates` and `EstimatedTokens` still set the column) — verified at
   design — but run it anyway.
4. **DEC-044** (§5), **`docs/api-contract.md`** (§6).
5. **`scripts/test-docs.sh`** (§7): `U9`, `Y3`, `X6`. Run `just test-docs`;
   `Y3` and `X3` fail until step 6 and step 9.
6. **`decisions/DEC-048-…`** (§9). Then `just test-docs`; `Y3` goes green.
7. **`docs/engineering-practices.md`** (§8) sentence, then **`just inventory`**
   and paste between the markers. Never hand-edit a row. `X3` goes green.
8. **`CHANGELOG.md`** (§10), **the stage file** (§11).
9. Full gate: `just test`, `just lint`, `gofmt -l .`, `go vet ./...`,
   `just test-docs` (**ALL OK, 188 ids**).
10. **Mutations M-A…M-E**, each mutant confirmed present by `shasum -a 256`
    and restored from a `/tmp` backup. Paste the output into Build Completion.

`golangci-lint` is not on `PATH` by default; it lives at `~/go/bin`
(v2.13.1 confirmed at design). `export PATH="$HOME/go/bin:$PATH"`.

---

### 1. `internal/export/memory.go` — four edits

**1a. Replace the `MemoryOptions` doc comment (lines 13-20) with:**

```go
// MemoryOptions controls the rule-based memory-slice digest (SPEC-073), the
// eighth DEC-014 consumer. Scope is always "lifetime" and is hard-coded
// rather than a field: DEC-014's Scope: is the TIME WINDOW a document covers,
// and memory applies none — a count bound (PoolLimit) is not a date bound.
// That is the property DEC-045 sub-decision 8 leans on when it withholds
// since/until/day from brag_memory. It is NOT a claim that every entry was
// ranked: the pool is capped, so entries older than the PoolLimit-th most
// recent are unreachable on a bare invocation (proven by
// TestMemoryCmd_ThreeReadsComposeThePool/bare-recency-read-is-capped).
// Filters is the pre-formatted markdown line ("(none)" or "--query X
// --project Y" in declared order); FiltersJSON is the object the JSON
// envelope renders (Go's map encoder sorts keys alphabetically — the
// documented DEC-014 markdown/JSON ordering asymmetry). Now is injected for a
// deterministic Generated: line.
```

**1b. Replace the `ToMemoryMarkdown` doc comment (lines 27-35) with:**

```go
// ToMemoryMarkdown renders a memory.Result as the DEC-014/DEC-043/DEC-044
// memory-slice digest: header + provenance block, then ## Slice (the ranked,
// budget-trimmed entry lines) and ## Budget (the accounting).
//
// The provenance count is Candidates: N (DEC-048) — how many DISTINCT entries
// were RANKED. It is not the corpus size and not the included count; ##
// Budget decomposes it (Included + Skipped == Candidates). It is a
// pool-composition artifact, not a cap: Gather runs up to three
// PoolLimit-capped reads and Slice dedupes their union, so the number GROWS
// when --query or --project adds a read. It is bounded by
// min(3*PoolLimit, corpus), never by PoolLimit alone. The five sibling
// DEC-014 exporters keep Entries:, which counts entries in scope and NARROWS
// under a filter; this number does the opposite, which is why it does not
// share the word.
//
// On an empty candidate pool only the header block (through "Candidates: 0")
// is emitted — no ## Slice, no ## Budget (DEC-014 part 4). A non-empty pool
// that includes zero entries still renders both sections. Returns bytes with
// the trailing "\n" stripped (matches every other renderer).
```

**1c. Line 43** — one token:

```go
	fmt.Fprintf(&buf, "Candidates: %d\n", result.Candidates)
```

**1d. The envelope — line 89 and line 122.** Both gofmt-verified at design;
no other line in either block moves.

```go
	Candidates      int                `json:"candidates"`
```

```go
		Candidates:      result.Candidates,
```

---

### 2. `internal/export/memory_test.go`

Four `Entries: 8` / `Entries: 0` lines inside `memoryGolden1`…`memoryGolden4`
become `Candidates: …`; `"entries": 8,` in `memoryJSONGolden` becomes
`"candidates": 8,`; the `` `"entries": 0` `` needle at `:353` becomes
`` `"candidates": 0` ``; the comment at `:281` says `Candidates: 0`.

The measured post-change bodies (captured from the failing tests' own
`--- got ---` output at design) — only the sixth line of each moves:

```
memoryGolden1 line 6:  Candidates: 8
memoryGolden2 line 6:  Candidates: 8
memoryGolden3 line 6:  Candidates: 8
memoryGolden4 line 6:  Candidates: 0     (and the document ends there)
memoryJSONGolden:      "candidates": 8,  (4th key, replacing "entries": 8,)
```

**And the cached byte length at `:247-248`** — the one no grep finds
(Finding 8). `Candidates:` is three bytes longer than `Entries:`:

```go
	if len(got) != 781 {
		t.Errorf("byte length = %d, want 781", len(got))
	}
```

Do **not** delete this assertion to make it go away. It is a second,
independent pin on `memoryGolden1` — byte-exact equality plus a length check —
and it is the only guard in the file that would catch a golden edited to match
a wrong render. Re-derive it (`781`, measured at design); do not remove it.

**New test** (append near the other markdown golden tests; `strings` is
already imported):

```go
// TestToMemoryMarkdown_HeaderIsCandidatesNotEntries pins DEC-048 at the
// renderer: the provenance count is a `Candidates:` line and there is no
// `Entries:` line at all. Line equality and HasPrefix, never
// strings.Contains — `Candidates: 8` does not contain `Entries: 8`, so a
// Contains-shaped guard looks sufficient while silently tolerating any future
// two-line or suffixed variant (AGENTS.md §9, the heading-level addendum).
func TestToMemoryMarkdown_HeaderIsCandidatesNotEntries(t *testing.T) {
	result := memory.Slice(memoryFixtureEntries(), memory.Options{Budget: 2000})
	got, err := ToMemoryMarkdown(result, MemoryOptions{Filters: "(none)", Now: memoryFixtureNow})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	sawCandidates := false
	for _, ln := range strings.Split(string(got), "\n") {
		if strings.HasPrefix(ln, "Entries:") {
			t.Errorf("provenance still carries an Entries: line: %q", ln)
		}
		if ln == "Candidates: 8" {
			sawCandidates = true
		}
	}
	if !sawCandidates {
		t.Errorf("expected a line %q, got:\n%s", "Candidates: 8", got)
	}
}
```

---

### 3. `internal/cli/memory_test.go`

`memoryCmdGolden1` (`:109`) and `memoryCmdGolden2` (`:135`) — sixth line
becomes `Candidates: 8`. Lines `:591-592`:

```go
	if !strings.Contains(out, "Candidates: 8") {
		t.Errorf("expected Candidates: 8, got:\n%s", out)
	}
```

*(That one stays a `Contains` because it is asserting the soft boost, not the
header shape — LD8 governs the header guards, which are the two new tests.)*

**New test.** The fixture is the `TestMemoryCmd_ThreeReadsComposeThePool`
shape; the four expected headers were measured at design at `--budget 100`
and re-measured at `--budget 20000`:

```go
// TestMemoryCmd_CandidateCountGrowsWhenAFlagAddsARead is the CLAIM test for
// DEC-048, not the counterexample. A bare `Candidates: 200` is the one case
// that looks like a cap. It is not: the number is the deduped union of up to
// three PoolLimit-capped reads (DEC-043 sub-decision 5), so it GROWS when a
// flag adds a read, and it is bounded by min(3*PoolLimit, corpus).
//
// Corpus: 200 fillers (all current) plus two entries backdated to 2020 — one
// in project orbit, one matching the query "auth". Neither is reachable by
// the bare recency read, so each flag adds exactly one candidate.
//
// The header is also budget-INDEPENDENT: the same four values were measured
// at --budget 100 and --budget 20000 during design. 100 is used here so
// Skipped is non-zero and the Included+Skipped invariant is not satisfied
// trivially.
func TestMemoryCmd_CandidateCountGrowsWhenAFlagAddsARead(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	for i := 1; i <= 200; i++ {
		if _, err := s.Add(storage.Entry{Title: fmt.Sprintf("filler %d", i)}); err != nil {
			t.Fatalf("add filler %d: %v", i, err)
		}
	}
	orbitEntry, err := s.Add(storage.Entry{Project: "orbit", Title: "ancient orbit note"})
	if err != nil {
		t.Fatalf("add orbit entry: %v", err)
	}
	authEntry, err := s.Add(storage.Entry{Title: "ancient auth note"})
	if err != nil {
		t.Fatalf("add auth entry: %v", err)
	}
	s.Close()
	if err := storagetest.Backdate(dbPath, orbitEntry.ID, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("backdate orbit entry: %v", err)
	}
	if err := storagetest.Backdate(dbPath, authEntry.ID, time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("backdate auth entry: %v", err)
	}

	withNowFunc(t, time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC))

	cases := []struct {
		name string
		args []string
		want string
	}{
		{"bare", nil, "Candidates: 200"},
		{"query adds the match read", []string{"--query", "auth"}, "Candidates: 201"},
		{"project adds the project read", []string{"--project", "orbit"}, "Candidates: 201"},
		{"both reads", []string{"--query", "auth", "--project", "orbit"}, "Candidates: 202"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"--budget", "100"}, tc.args...)
			out, errStr, err := runMemoryCmd(t, dbPath, args...)
			if err != nil {
				t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
			}

			var header, included, skipped string
			for _, ln := range strings.Split(out, "\n") {
				switch {
				case strings.HasPrefix(ln, "Entries:"):
					t.Errorf("provenance still carries an Entries: line: %q", ln)
				case strings.HasPrefix(ln, "Candidates: "):
					header = ln
				case strings.HasPrefix(ln, "- Included: "):
					included = strings.TrimPrefix(ln, "- Included: ")
				case strings.HasPrefix(ln, "- Skipped: "):
					skipped = strings.TrimPrefix(ln, "- Skipped: ")
				}
			}
			if header != tc.want {
				t.Errorf("header = %q, want %q\n%s", header, tc.want, out)
			}

			// Included + Skipped == Candidates (DEC-044), restated against the
			// new label at the CLI layer.
			n, err := strconv.Atoi(strings.TrimPrefix(tc.want, "Candidates: "))
			if err != nil {
				t.Fatalf("bad want literal %q: %v", tc.want, err)
			}
			inc, err := strconv.Atoi(included)
			if err != nil {
				t.Fatalf("parse Included %q: %v", included, err)
			}
			skip, err := strconv.Atoi(skipped)
			if err != nil {
				t.Fatalf("parse Skipped %q: %v", skipped, err)
			}
			if inc+skip != n {
				t.Errorf("Included(%d)+Skipped(%d) != %d", inc, skip, n)
			}
			if skip == 0 {
				t.Errorf("Skipped == 0 at --budget 100; the invariant check is trivial")
			}
		})
	}
}
```

**Imports:** this test needs `strconv`. `fmt`, `path/filepath`, `strings`,
`testing`, `time`, `storage` and `storagetest` are already imported in the
file (`TestMemoryCmd_ThreeReadsComposeThePool` uses all of them). Add
`strconv` and let `gofmt` order it.

---

### 4. `internal/mcpserver/resources_test.go` — new test

Place it in the `--- Body parity (LD1, LD5) ---` section, directly under
`TestResourceRecent_IsByteIdenticalToMemoryMarkdown`. Measured at design: that
line is `Entries: 3` today.

```go
// TestResourceRecent_CarriesTheCandidatesHeader pins the AGENT-VISIBLE bytes
// (DEC-045, DEC-048). The byte-identity test above compares two renderings of
// the SAME function, so it passes whatever the header says — it proves parity,
// not correctness. Nothing in this package would have failed if the header
// regressed; this is the assertion that would.
func TestResourceRecent_CarriesTheCandidatesHeader(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one", "two", "three")
	fixed := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	restore := setNowFunc(t, func() time.Time { return fixed })
	defer restore()

	res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uriMemoryRecent})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	body := res.Contents[0].Text
	lines := strings.Split(body, "\n")
	for _, ln := range lines {
		if strings.HasPrefix(ln, "Entries:") {
			t.Errorf("resource body still carries an Entries: line: %q", ln)
		}
	}
	if len(lines) < 6 {
		t.Fatalf("body has %d lines, want at least 6:\n%s", len(lines), body)
	}
	if lines[5] != "Candidates: 3" {
		t.Errorf("line 6 = %q, want %q\nbody:\n%s", lines[5], "Candidates: 3", body)
	}
}
```

*(`t.Fatalf` on the length first, then index — no builtin `min`. Nothing in
`internal/` uses Go's builtin `min`/`max` today, and a spec about house-style
consistency is the wrong place to introduce one.)*

---

### 5. `decisions/DEC-044-memory-slice-token-budget-and-line-shape.md` — three sites

**`:173`:**

```
`Included + Skipped == Candidates` (the deduped candidate-pool size) is an invariant,
```

**`:251`:** `Entries: 0` → `Candidates: 0` in the sentence
*"the document ends after `Candidates: 0` and both body sections are
omitted."*

**`:406`:**

```
- `Included + Skipped == Candidates` — pinned by `TestSlice_CountsPartitionTheCandidatePool`.
```

Add, at the end of DEC-044's Validation section, one line so the record shows
the rename was deliberate:

```
> **Renamed at SPEC-084/DEC-048 (2026-08-22).** The rendered label is
> `Candidates:`, not `Entries:`. The invariant is unchanged — only the word
> it is stated against. `Entries:` now means entries in scope on every
> DEC-014 consumer that emits it.
```

---

### 6. `docs/api-contract.md` — two sites

**Replace lines 892-895** (the memory provenance bullet):

```
- **Provenance:** `Generated:` (RFC3339), `Scope: lifetime` (hard-coded —
  memory applies no time window), `Filters:` (`--query <text> --project
  <name>` in that declared order, or `(none)`), and `Candidates: N` — how
  many distinct entries were **ranked**. Not the corpus size and not the
  included count: the pool is the deduped union of up to three
  200-row-capped reads, so it **grows** when `--query` or `--project` adds
  one (`200` bare, `243` with `--project` on a 387-entry corpus).
  `Entries:` on the other DEC-014 consumers means entries in scope and
  narrows under a filter — see
  [DEC-048](../decisions/DEC-048-provenance-count-names-what-it-counted.md).
```

**Replace line 903:**

```
    and counts). `Included + Skipped == Candidates`.
```

Leave `:931` (*"a window would make `Scope: lifetime` a lie"*) and `:933-935`
(the three-bounded-reads paragraph) exactly as they are — both are true and
`:931` is DEC-045 sub-decision 8's argument.

---

### 7. `scripts/test-docs.sh` — three edits

**7a. New `U9`, appended to Group U after `U8`:**

```sh
# U9 — the memory provenance count is documented under its NEW name, and the
# false gloss on Scope: is gone. Two claims, one id — the doc-level guard on
# LD1 and LD5, the shape V3 already uses for the soft-boost correction.
#
# THE NEGATIVE NEEDLE IS A FRAGMENT ON PURPOSE. The falsified sentence wraps:
#   892| ... `Scope: lifetime` (memory ranks
#   893| the whole corpus, like `stats`), ...
# so `grep -F 'memory ranks the whole corpus'` returns ZERO hits against the
# BROKEN file — a needle spanning the wrap would pass vacuously forever.
# Verified at SPEC-084 design. grep is line-oriented; this doc wraps at 74.
if [ ! -f docs/api-contract.md ]; then
    fail "U9" "docs/api-contract.md does not exist"
else
    u9_bad=""
    if ! grep -F -q -- '`Candidates: N`' docs/api-contract.md; then
        u9_bad="$u9_bad [memory provenance does not name Candidates: N]"
    fi
    if grep -F -q -- '(memory ranks' docs/api-contract.md; then
        u9_bad="$u9_bad [the falsified 'memory ranks the whole corpus' gloss is still present]"
    fi
    if [ -z "$u9_bad" ]; then
        ok "U9"
    else
        fail "U9" "docs/api-contract.md:$u9_bad"
    fi
fi
```

*(`if ! grep …; then` / `if grep …; then` rather than `||` / `&&` — the file
runs under `set -eu`, and the `if` form is the idiom `U3`/`C5`/`AA` already
use.)*

**7b. `Y3` — re-pin and note.** Replace the `RE-PINNED` paragraph and the two
`grep -F -q` lines:

```sh
# RE-PINNED 46 -> 47 at SPEC-084, which adds DEC-048. Deliberate corpus
# change, not drift. This is the THIRD consecutive spec whose §9 half-(b)
# value-grep found Y3 where a concept-grep ("what pins the decision count?")
# found only X3 — SPEC-083 recorded the same miss in this comment. Grep for
# the value, not the idea.
```

```sh
    printf '%s\n' "$y3_out" | grep -F -q 'Decision records | 47 |' || y3_bad="$y3_bad decision-records!=47"
```

The worked-example sentence inside `Y3`'s comment (*"script and page would
still agree, just agree on 47"*) moves to **48** for the same reason SPEC-083
moved it to 47: it illustrates the miscount, and the miscount is
`decisions + the DEC-041 reservation`.

**7c. `X6` comment** — `only 1 of 46 DECs` → `only 1 of 47 DECs`. Numerator
unchanged (LD4). No assertion behaviour changes.

---

### 8. `docs/engineering-practices.md`

**Replace the sentence at `:282-287`:**

```
[`STAGE-022`](../projects/PROJ-007-quality-and-portfolio-readiness/stages/STAGE-022-measured-and-enforced.md)
closed the lint gate and the coverage floor, both described above, and the
`Entries:` envelope inconsistency
([`DEC-048`](../decisions/DEC-048-provenance-count-names-what-it-counted.md)).
It does not close the rest: its *Explicitly out of scope* section defers
benchmarks to
[`PROJ-009`](../projects/PROJ-009-scale-baseline-and-harness/brief.md), running
the documentation assertions in CI is owned by nothing today, and the
no-network claim stays enforced by review.
```

Then run **`just inventory`** and paste its output between the
`<!-- inventory:begin -->` / `<!-- inventory:end -->` markers. **Three** rows
move: `Decision records` 46 → 47, `Documentation assertions (distinct ids)`
187 → 188, and `Go test functions` 812 → 815 — §2/§3/§4 add exactly three test
functions, so this row was derivable and was still missed by reasoning about
which rows the change "touches" instead of running the generator. **Do not
hand-edit a row** — `X3` diffs the whole table, which is why the block was
correct anyway.

`X7` band check: 2,525 words at design, band 1800..2700, 175 words of ceiling
headroom. The replacement is net-neutral.

---

### 9. `decisions/DEC-048-provenance-count-names-what-it-counted.md` (new)

Transcribe verbatim. **The `  type: decision` line is load-bearing** — `Z7`
fails without it and the inventory silently under-reports.

````markdown
---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-048                        # stable, never reused
  type: decision                     # decision | analysis | recommendation | observation
  confidence: 0.93                   # 0.0 - 1.0, honest assessment
  audience:                          # who needs to know?
    - developer
    - agent

agent:
  id: claude-opus-5
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-007                       # the project during which this was decided
repo:
  id: bragfile

created_at: 2026-08-22
supersedes: null
superseded_by: null

tags:
  - output-shape
  - dec-014-envelope
  - memory
  - mcp
  - naming
---

# DEC-048: a provenance count line names what it counted

## Decision

In a DEC-014 envelope, **`Entries: <N>` means the number of entries the
document covers, and it narrows when a filter narrows the document.** A
consumer whose headline count is not that must **name it something else** —
`brag memory` reports **`Candidates: <N>`** (markdown) and **`"candidates"`**
(JSON), the deduped candidate-pool size.

## Context

Six `internal/export/` renderers emit a headline count. Five compute
`len(entries)`. `brag memory` computes `result.Candidates` — the deduped union
of up to three `PoolLimit`-capped reads (DEC-043 sub-decision 5) — and printed
it under the same word.

The two numbers move in **opposite directions under the same flag**. Measured
against the live 387-entry corpus at framing and re-measured at design, both on 2026-08-22:

| Command | bare | `--project bragfile` |
|---|---:|---:|
| `brag export --format markdown` | 387 | **74** |
| `brag memory` | 200 | **243** |

The full memory series: bare `200`, `--query brag` `232`,
`--project bragfile` `243`, both `247`. So the number is **not a cap** — it is
bounded by `min(3 × PoolLimit, corpus)` and it **grows** when a flag adds a
read, because DEC-043 sub-decision 4 makes `--project` a soft boost that adds
a third read rather than a filter that removes rows.

**Nobody legislated the word.** DEC-013 (markdown export shape) created
`Entries: <N>` for `brag export`, correctly. DEC-014 — the envelope every
rule-based consumer inherits — locks `Generated:`, `Scope:` and `Filters:`,
and **does not contain `Entries:` at all**; its part 4 even says the empty
document *"ends after the `Filters:` line,"* while every implementation ends
after the count. DEC-028 re-specified a variant for `impact`
(`Entries: <shown>/<in-window> with impact`). DEC-044 redefined the word for
`memory` in writing (*"the deduped candidate-pool size"*) while keeping it.

So this was never drift. It was **one word with two DEC-level definitions over
an envelope that never claimed the line**, and therefore had no authority to
arbitrate. Both the renderer's own comment and `docs/api-contract.md` already
stated the truth. The rendered output was the only artifact that lied — and
via DEC-045's byte-identity guarantee, that output is what every MCP-connected
agent auto-loads from `brag://memory/recent`.

## Alternatives Considered

- **State a denominator — `Entries: 200 of 387`.** Follows DEC-028's
  two-number precedent. Rejected on three independent grounds. (1) There is no
  denominator available: no `Store.Count()` exists, and `brag stats` gets 387
  by reading the corpus **uncapped**, which is exactly what DEC-043
  sub-decision 5 forbids — buying it means a new storage method plus a
  deliberate DEC-043 exception, plumbed through the CLI and the MCP server.
  (2) Under a soft boost there is **no meaningful denominator**, because
  nothing was filtered out: "243 of 387" is a ratio between two numbers that
  do not stand in that relation. (3) The `impact` precedent is cheap only
  because `cli/impact.go:130` gets both numbers from one already-materialized
  slice; `memory` never materializes the corpus.
- **Report both — a `Candidates:` line beside an `Entries:` line.** Rejected:
  it adds a provenance line to an envelope that never legislated one, which is
  the precise mechanism that produced this defect. It also leaves the reader
  deciding which number `## Budget` partitions.
- **Raise or remove `PoolLimit`.** Rejected. The cap is a budget decision with
  a documented head guarantee, and DEC-043's revisit trigger (d) is about a
  caller wanting a larger slice, not about a mislabelled header. Decisively:
  **it would not fix the defect** — at any cap the number is still a union of
  up to three reads that grows under flags and is not the corpus.
- **Rename the label but leave the JSON key `"entries"`.** Rejected. `memory`
  is the **only** exporter with a bare integer `"entries"` key, and
  `"entries"` in the same JSON namespace already means *an array of entry
  objects* (`impact.go:82`, `review.go:83`, `summary.go:81`,
  `wrapped.go:168`). Leaving it would preserve a collision that the markdown
  fix does not touch, and split one semantic decision across two surfaces of
  one document.
- **Amend DEC-044 with an `## Amendment` heading instead of a new record.**
  Rejected. This decision binds **DEC-014 consumers**, not `memory`, which is
  not something a `memory`-scoped amendment can do. It would also adopt a
  convention `guidance/questions.yaml`'s open `dec-amendment-heading-convention`
  **deliberately declined** at SPEC-079, riding it in on a bug spec — the very
  objection recorded there.
- **Change `Scope: lifetime` too.** Rejected as *rendered value*, taken as
  *justification*. See Consequences.

## Consequences

**What changed.** `brag memory` renders `Candidates: <N>`; its JSON envelope
key is `"candidates"`. `internal/memory` is untouched — the Go field was
already named `Candidates`, and this decision brings the output in line with
the name the rest of the repo has used all along (the field, the package doc,
DEC-043 sub-decision 5's heading, DEC-044, `docs/api-contract.md`, and
`TestSlice_CountsPartitionTheCandidatePool`).

**What binds forward.** Any future DEC-014 consumer whose headline count is
*entries in scope* uses `Entries: <N>`. Any consumer whose count is something
else names it. `impact`'s `Entries: <shown>/<in-window> with impact` (DEC-028)
remains the sanctioned form for reporting a pair. The five existing
`Entries:`-emitting exporters (`export`, `coverage`, `spark`, `wrapped`, and
`impact`'s numerator) are correct and are not changed.

**This is a breaking change to the JSON envelope**, and it is named as one. A
consumer reading `.entries` from `brag memory --format json` or from the
`brag_memory` tool with `format: "json"` breaks. Accepted: the repo is
pre-1.0, the key is six weeks old (SPEC-073), DEC-011/DEC-014 lock the
envelope *shape* rather than the correctness of a payload key, and shipping
two keys for one number would itself violate DEC-014 choice 2's flat
single-object envelope.

**This is an agent-visible contract change.** `brag://memory/recent` and
`brag://memory/project/{name}` are byte-identical to `brag memory` by
construction (DEC-045), so the header change reaches every connected agent.
Resource URIs, names, MIME types and tool schemas are unchanged; only the
markdown payload moves. `brag_memory`'s tool description mentions neither
`Entries` nor `candidate`, so no description text moves. The byte-identity
test compares two renderings of the same function and therefore **passes
either way** — so SPEC-084 adds `TestResourceRecent_CarriesTheCandidatesHeader`,
the first assertion in `internal/mcpserver` that would fail if the header
regressed.

**`Scope: lifetime` stays, and its justification was repaired.** DEC-014's
`Scope:` is the **time window** a document covers, and `brag memory` applies
none — a count bound is not a date bound. That is also the premise of DEC-045
sub-decision 8, which withholds `since`/`until`/`day` from `brag_memory`
because *"a window would make the envelope's `Scope: lifetime` line untrue."*
What was false was the gloss — *"memory ranks the whole corpus, like stats"*
(`internal/export/memory.go:14`, `docs/api-contract.md:892`) — falsified by
`TestMemoryCmd_ThreeReadsComposeThePool/bare-recency-read-is-capped`, which
proves an entry at recency rank 201 is unreachable on a bare invocation. Both
sites now read *"(hard-coded — memory applies no time window)"*, the gloss
`brag stats` already carried at `docs/api-contract.md:392`.

**What this does NOT resolve.** The header does not explain *why* the number
grows under `--project`. `Candidates: 243` is true and makes no scope claim,
so the false reading (*"243 entries are in this project"* — 74 are) is gone;
but a reader who does not know `--project` is a soft boost will still find the
growth surprising. That explanation belongs to the flag help (*"a soft boost,
not a filter"*), the `Filters:` echo, and `docs/for-ai-agents.md` — not to a
provenance count line. DEC-043 sub-decision 4 is not relitigated here.

**What this does not decide.** `dec-amendment-heading-convention` stays
`open`. This record carries no `## Amendment` heading and takes no position on
whether future records should.

## Validation

- `brag memory` prints `Candidates: <N>`; no rendered line begins `Entries:` —
  `TestToMemoryMarkdown_HeaderIsCandidatesNotEntries`.
- The count grows when a flag adds a read, and is budget-independent —
  `TestMemoryCmd_CandidateCountGrowsWhenAFlagAddsARead` (200 / 201 / 201 / 202).
- `Included + Skipped == Candidates` — `TestSlice_CountsPartitionTheCandidatePool`,
  plus the CLI-layer restatement in the test above.
- The empty pool still ends after `Candidates: 0`, no `## Slice`, no
  `## Budget` — `TestToMemoryMarkdown_EmptyCorpusGolden`.
- The JSON key is `"candidates"` — `TestToMemoryJSON_BlendedGolden`,
  `TestToMemoryJSON_EmptyCorpusEmitsEveryKey`.
- The MCP resource carries the header —
  `TestResourceRecent_CarriesTheCandidatesHeader`.
- `docs/api-contract.md` documents the new name and no longer carries the
  false gloss — `U9` in `scripts/test-docs.sh`.

**Revisit triggers.**

- **T1 — a ninth DEC-014 consumer lands whose headline count is neither
  entries-in-scope nor a pair.** Then this record's rule is exercised for the
  first time on a new surface; check it still reads as a rule rather than a
  post-hoc account of `memory`.
- **T2 — `--project` ever becomes a hard filter** (DEC-043's revisit path (b),
  a `--project-only` flag). Then `Candidates:` would narrow rather than grow
  and the *reason* for the different word weakens, though the number is still
  a pool.
- **T3 — a `Store.Count()` arrives for some other reason.** Then the
  denominator becomes free and `Candidates: 200 of 387` is worth re-costing —
  but re-read Alternative (2) first: under a soft boost the ratio is still not
  a ratio.
- **T4 — an external consumer of the JSON envelope appears.** Then the
  "pre-1.0, rename freely" premise expires and the next key rename needs a
  deprecation path this repo does not have.

## References

- **DEC-013** — created `Entries: <N>` for `brag export`. Unchanged.
- **DEC-014** — the envelope. Choice 3 locks `Generated:`/`Scope:`/`Filters:`
  and **not** `Entries:`; choice 4 governs the empty document. The absence is
  the root cause; DEC-014 is deliberately not edited, so it stays readable.
- **DEC-028** — `impact`'s two-number variant, the sanctioned form for a pair.
- **DEC-043** — sub-decision 5 (one bounded read per list, `PoolLimit = 200`)
  is why the number is a union; sub-decision 4 (soft boost, never a filter) is
  why it grows. Neither is relitigated.
- **DEC-044** — the `Included + Skipped` invariant. Restated against the new
  label at `:173`, `:251` and `:406`; the invariant itself is unchanged.
- **DEC-045** — byte-identity makes this an agent-visible contract; sub-decision
  8 is why `Scope: lifetime` stays exactly that string.
- **SPEC-073 / SPEC-074** — built the surfaces. **SPEC-082** measured the
  defect. **SPEC-084** is this record's spec.
````

---

### 10. `CHANGELOG.md` — under `## [Unreleased]`

```markdown
### Changed

- **`brag memory`'s headline count is now `Candidates: <N>`, not
  `Entries: <N>`** ([DEC-048](decisions/DEC-048-provenance-count-names-what-it-counted.md)).
  The number was never the corpus size and never a cap: it is the deduped
  union of up to three 200-row reads, so it *grew* when you passed a flag —
  `200` bare and `243` with `--project` on the same 387-entry corpus, while
  `brag export --project` correctly *narrowed* 387 → 74. Same word, same
  flag, opposite directions. The five other commands that print `Entries:`
  are unchanged and correct; on those it means entries in scope.
  `brag://memory/recent` and `brag://memory/project/{name}` carry the new
  header too, since they are byte-identical to `brag memory`.

### Breaking

- **`brag memory --format json` (and the `brag_memory` MCP tool with
  `format: "json"`) renames the `entries` key to `candidates`.** Same number,
  honest name — and it removes a collision with the `entries` key that means
  *an array of entry objects* on `brag impact`, `brag review`, `brag summary`
  and `brag wrapped`. Update any `jq .entries` to `jq .candidates`.
```

*(If `### Breaking` is not an existing heading in this CHANGELOG's vocabulary,
fold the second bullet into `### Changed` and keep the word "breaking" in the
prose — Keep a Changelog's canonical set is Added/Changed/Deprecated/Removed/
Fixed/Security. Check the file before choosing; do not invent a heading the
release tooling has never seen.)*

---

### 11. `projects/PROJ-007-.../stages/STAGE-022-measured-and-enforced.md`

- **Success criterion** — mark the `Entries:` envelope item met, naming
  `DEC-048` and the measured before/after (`Entries: 200/232/243/247` →
  `Candidates:` on all four, `Included + Skipped` unchanged).
- **The `In scope` item 3** — its description still says *"the pool capped at
  `PoolLimit=200`"*, which framing already corrected in the backlog entry
  below it. Bring the two into agreement rather than leaving the stage
  carrying both readings.
- **The SPEC-084 backlog entry** — `[x] SPEC-084 (ship)`, with what design
  settled: all six forks, `Candidates:` on both surfaces, `Scope:` resolved by
  repairing the justification rather than the value, and the four findings
  the frame did not have (the JSON-namespace collision, the DEC-045 dependency
  on `Scope: lifetime`, `Y3`, and the cached golden byte length).
- **Count line** — `3 shipped / 0 active / 0 pending`.
- **Design Notes** — carry forward the two routed items that are still open
  (`root.go:13` / `.golangci.yml:71`'s unguarded "four"; `store.go:11`'s V1
  package-vs-layer noun), and add the two promotion candidates this design
  earned:
  - **§9 half-(b) at N=3** — *grep for the value, not the idea* (Finding 3:
    `Y3` pinned the decision count a third consecutive time and a concept-grep
    found only `X3`).
  - **§9's cached-literal rule, N=4 — with a limit** (Finding 8). Half (b)
    says grep for every literal occurrence of a value the spec changes. That
    fails when the cached value shares **no token** with its source: nothing
    in `len(got) != 778` names the golden it measures. The candidate wording
    is a clause, not a new rule — *"a value you did not know would move cannot
    be grepped for; run the change before writing `## Failing Tests` closed."*
    Weigh it against the cost of saying so: this spec caught it at design
    precisely by running the cycle, which §12(b) already requires.

## Build Completion

- **Deviations from the spec:**
  1. `docs/api-contract.md:921` (the `--format markdown|json` flags bullet's
     "JSON top-level keys" list) also names `entries` and was not in the
     spec's "two sites" enumeration for that file. Neither `grep -rn
     'Entries:'` nor a case-sensitive `grep 'Entries'` finds it — it is
     lowercase, has no colon, and sits inside a backtick-quoted key list —
     which is the same class of miss Finding 8 names for the cached golden
     byte length. Fixed it (`entries` → `candidates`) since leaving it would
     ship a factually wrong JSON key list. No test guards this specific line;
     `U9` guards the provenance bullet and the `Scope:` gloss, not this one.
  2. STAGE-022's stage file was updated during build rather than left for
     verify/ship. AGENTS.md §12 "During build" lists four actions (fill Build
     Completion, advance-cycle, create DECs, open PR) and none of them is
     "update the parent stage file" — and `git log` on this exact file across
     SPEC-082 and SPEC-083 confirms build never touched the backlog
     entry/checkbox/count line for either; only frame, design, and ship did.
     Design for SPEC-084 skipped its own stage-file update (unlike SPEC-083's
     design, which did), so the entry was still frame-only going into build.
     Order of work step 8 explicitly lists "the stage file (§11)" as build
     work, and §11 itself asks for `[x] SPEC-084 (ship)` and a `3 shipped`
     count line — both would be false statements before verify/ship actually
     happen. Resolved by writing what's true now: the cycle tag moved
     `(frame)` → `(design)`, checkbox stays unchecked, the in-scope item 3
     description was reconciled with the framing text already below it
     (a same-file internal-consistency fix, not a ship-state claim), and a
     consolidated "Designed 2026-08-22, built 2026-08-23" paragraph captures
     what design settled (which its own session never wrote down) plus what
     build measured. Left untouched for verify/ship, matching precedent:
     Success Criteria wording, the checkbox flip, the count-line bump to
     "shipped," and the two AGENTS.md promotion candidates (§9 half-(b) at
     N=3/N=4) that SPEC-083's ship reflection shows get written at ship time.
- **New DEC-* files created:** `decisions/DEC-048-provenance-count-names-what-it-counted.md`.
- **Constraints checked:** `test-before-implementation` (12 failing test
  functions written and observed red for their stated assertion before
  `memory.go` was touched — verbatim match to design's own fail-first count);
  `one-spec-per-pr` (single branch, single PR); `no-sql-in-cli-layer`
  (untouched, no SQL added); `stdout-is-for-data-stderr-is-for-humans`
  (untouched, no new output stream).
- **Gates:** `go test ./...` 14/14 packages ok; `gofmt -l .` empty;
  `go vet ./...` clean; `just lint` 0 issues; `golangci-lint config verify`
  exit 0; `just test-docs` ALL OK at 188 assertion ids (`Decision records`
  46 → 47, `Documentation assertions` 187 → 188, both regenerated via
  `just inventory` and pasted, never hand-edited); `git diff --stat -- internal/`
  confined to exactly the four files the spec named (`memory.go`,
  `memory_test.go` ×2, `resources_test.go`) — no sibling exporter, no
  `internal/aggregate/` file touched. Live-corpus re-measurement (388 entries,
  grown by one since design): `Candidates: 200 / 232 / 243 / 247` across the
  four flag combinations, `brag export --project bragfile` still narrows
  387→74 (388→74 today) unchanged. Empty-corpus markdown and JSON both
  terminate correctly at `Candidates: 0`. All five mutations (M-A…M-E)
  confirmed present by `shasum -a 256` before their failure was credited and
  restored from a `/tmp` backup with the hash returned to its exact
  pre-mutation value afterward (never `git checkout`) — see the transcript
  above for each hash pair.

### Build-phase reflection (3 questions, short answers)

- **What was unclear in the spec?** Whether "Notes for the Implementer"
  §11's stage-file instructions (marking the backlog entry `(ship)`, bumping
  the count line to "shipped") were meant for build to execute literally now,
  or were forward-looking content for verify/ship to write once true. The
  spec's own Order of work bundles it into build's step 8, but AGENTS.md's
  cycle model and two consecutive specs' git history say otherwise. Resolved
  in favor of the documented cycle model and observed precedent — see
  Deviations.
- **What was missing that you had to decide yourself?** The
  `docs/api-contract.md:921` JSON-key-list site (Deviation 1) — the spec's
  own enumeration said "two sites" for this file and this wasn't one of them.
- **What would you do differently?** Nothing in the code/test/doc path — the
  spec's literals, line numbers, and measured values all matched the live
  tree exactly, which made this the fastest build in the stage so far. Flag
  the stage-file-timing ambiguity at design time next, so build doesn't have
  to adjudicate a cycle-ownership question mid-session.

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

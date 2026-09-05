---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-085
  type: story                      # epic | story | task | bug | chore
  cycle: ship
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
  designed_at: 2026-09-05

insight:
  confidence: 0.91                 # up from framing's 0.82. Fork 3 did not
                                   # split this spec — it split out SPEC-086
                                   # instead, and every literal below was run
                                   # through its own tool before locking.

references:
  decisions:
    - DEC-014                      # the envelope every rule-based surface inherits
    - DEC-015                      # tags ARE normalized — supersedes DEC-004
    - DEC-043                      # memory RRF over three ordinal lists
    - DEC-044                      # the memory line shape — renders type, not tags
    - DEC-048                      # a provenance count names what it counted
    - DEC-049                      # THIS SPEC CREATES IT — the reserved type value
  constraints:
    - one-spec-per-pr
    - no-sql-in-cli-layer
  related_specs:
    - SPEC-073                     # the memory slice this must not destabilise
    - SPEC-074                     # Gather is shared verbatim with the MCP resources
    - SPEC-086                     # THE SPLIT — what the digests do with a failure
---

# SPEC-085: `brag learn` — the corpus holds a failure

> **Cycle: design.** Complexity **M**, re-affirmed *after* splitting Fork 4
> out into **SPEC-086**. All five forks settled. Every number below was
> re-derived on 2026-09-05 against `main` at **`81e639d`** (framing measured
> `1775feb`; PR #194 merged in between), corpus **398**. Every literal was
> written into a detached worktree and run through its own tool — see
> *§12(b) design-time pre-flight*, which turned up **six** defects, three of
> which would have shipped a guard that could never fail. (Framing's own
> factual errors are separate, and tabulated under *Re-measurement*.)

## What this spec ships

A capture verb, one reserved value, and the decision record that binds them:

```
brag learn -t "shared-worker pool did not cut cold starts" \
           -i "cost two days and produced nothing reusable"
brag list --type failed
```

It stops at the point where the corpus is *capable* of honesty. What the
celebratory digests do with a failure is **SPEC-086** (split at design, on the
user's Fork 4 decision). Ranking honesty is STAGE-024; narrating it is
STAGE-025.

## Context

`brag` holds **398 entries and not one failure.** That is not a capture-
discipline problem — there is no verb, no field value, and no documented
convention for recording work that did not work, so the absence is structural.

The consequence is that every surface built on the corpus is systematically
optimistic, and the two most important consumers are the ones that suffer most:

- **An agent reading the corpus cold** (`brag memory`, or the `brag://` MCP
  resources that load with no tool call at all, DEC-045) learns only what
  worked. It cannot avoid a path the user already burned two days on, because
  the burn was never written down.
- **The mirror** (STAGE-025) is supposed to say *"this project reads like a
  turnaround."* A turnaround has a downstroke. Over a wins-only corpus the
  mirror can only ever report a rising line.

## Re-measurement at design (2026-09-05, `main` at `81e639d`)

Framing's numbers were re-derived from scratch against a read-only copy of the
live database (`cp ~/.bragfile/db.sqlite`, queried with `sqlite3`), and against
the binary rather than by reading source. **Four of framing's figures moved or
were wrong.** Each is recorded because each changes an argument.

| Claim | Framing said | Measured at design | Effect |
|---|---|---|---|
| corpus size | 397 | **398** | framing's own brag (#408) landed |
| distinct `type` values | 19 | **18** | framing counted the empty bucket as a value |
| entries with no `type` | 113 | **113** ✔ | holds — all 113 are `''`, **zero `NULL`** |
| entries with `type` | — | **285** | 285 + 113 = 398 ✔ |
| surfaces with `--type` | 5 of 8 | **7** | framing's denominator omitted four commands |
| impact-bearing entries | 323/397 | **324/398** | both moved |
| `wrapped` vs `impact` body | 694 / 695 lines | **697 / 696** | same fact, both grew |

### The `type` field, measured

```
$ sqlite3 corpus.sqlite "SELECT COUNT(DISTINCT type) FROM entries WHERE type IS NOT NULL AND type <> '';"
18
$ sqlite3 corpus.sqlite "SELECT COUNT(*) FROM entries WHERE type IS NULL;"
0
$ sqlite3 corpus.sqlite "SELECT COUNT(*) FROM entries WHERE type = '';"
113
```

**18 distinct non-empty values, not 19.** An X-of-N claim is two measurements
plus a unit, and framing's unit was wrong: it counted `(empty)` as the 19th
*value* while separately reporting it as the 113 entries with *none*. The same
bucket, counted twice, in two different units. Nothing downstream depended on
19 rather than 18 — but the *shape* of the error is the one this repo has
recorded five times (brag #383), so it is named rather than quietly corrected.

The distribution, which is the evidence DEC-049 actually rests on:

| value | n | | value | n |
|---|---:|---|---|---:|
| `shipped` | 202 | | `bugfix` | 1 |
| *(empty)* | 113 | | `designed` | 1 |
| `milestone` | 27 | | `discipline` | 1 |
| `ship` | **15** | | `documented` | 1 |
| `feature` | 14 | | `feedback` | 1 |
| `learned` | 8 | | `hardened` | 1 |
| `release` | 4 | | `template` | 1 |
| `fixed` | **2** | | `tooling` | 1 |
| `planning` | 2 | | `verified` | 1 |
| `spike` | 2 | | | |

**`shipped` 202 / `ship` 15** and **`fixed` 2 / `bugfix` 1** are the finding.
An unpinned convention on this field has *already* fragmented, in live data,
with no mechanism that could have stopped it. That is not a hypothetical
argument for pinning the new value — it is the same field, the same user, and
a measured outcome.

### `learned` is occupied — and framing's reading of it is too generous

Framing wrote that all eight `type: learned` entries are *"wins about
learning"*. Reading all eight in full, that is imprecise in a way that matters:

- **id 18** — *"smoke test of brag add --json"*. A PROJ-001 test artifact. Not
  a lesson at all.
- **ids 56, 61, 75, 362** — genuine wins.
- **ids 80, 361, 383** — these take a **failure as their subject** and frame it
  as a win recovered: *"Measured the verification gate's blind spot … escape
  rate ~67%"*, *"Verify caught an unpinned test seam that CI, five goldens and
  two guards all passed"*, *"Five counts in one stretch were wrong the same
  way"*.

So `learned` is not *"wins"* — it is closer to **post-mortems with a
recovery**. That sharpens the argument rather than weakening it: redefining
`learned` to mean *failure* would make `brag list --type learned` return a mix
of wins, one smoke test, and three near-misses — which fails the stage's own
criterion that a failure be **retrievable as a failure**. The conclusion
framing reached is right; the reason is better than the one it gave.

---

## Fork 1 — where does a failure live: `type`, a reserved tag, or a column?

**Settled: `entries.type`.** Framing's leaning, confirmed — and confirmed by
*running* the read path rather than reasoning about it.

DEC-044's memory line (`internal/memory/memory.go:224`) renders
`[<project>/<type>]`. Tags do not appear. So the question is whether a failure
is labelled as one on the surface an agent actually reads. Built and executed
at design:

```
$ brag learn -t "shared-worker pool did not cut cold starts" -p bragfile \
             -i "cost two days and produced nothing reusable"
1
$ brag memory
- 1 2026-09-05 [bragfile/failed] shared-worker pool did not cut cold starts — cost two days and produced nothing reusable
$ brag list --type failed
1	2026-09-05T16:52:58Z	shared-worker pool did not cut cold starts
```

**Framing's supporting table was wrong and the correction strengthens the
fork.** Framing said `--type` exists on 5 of 8 commands. Measured by asking
each binary rather than by grepping source:

```
$ for c in add coverage delete edit export impact list memory review search \
           show spark stats story summary wrapped tags; do
    ./brag "$c" --help | grep -qE '^\s+--type ' && echo "$c YES" || echo "$c no"
  done
```

| `--type` (exact-match inclusion) | commands |
|---|---|
| **yes — 7** | `list`, `export`, `summary`, `impact`, `wrapped`, `story`, `coverage` |
| no — 6 | `memory`, `review`, `search`, `show`, `spark`, `stats` |

(`add` has `--type` as a *write* flag, `-k`; it is not a filter and is not
counted in either column.) Framing's "5 of 8" omitted `export` and `coverage`
from the numerator and four commands from the denominator. Retrieval is
therefore **more** free than framing claimed: Success Criterion 2 of the stage
is met by an existing flag on seven surfaces, and this spec writes no
retrieval code at all.

**`--type` is exact-match inclusion; there is no negation.** Verified in
`internal/storage/store.go:388`:

```go
if f.Type != "" {
    conds = append(conds, "e.type = ?")
    args = append(args, f.Type)
}
```

So *"wrapped, minus failures"* is code, not configuration — which is exactly
why Fork 4 became SPEC-086. Worth noting for that spec: `ListFilter` **does**
already contain a negation precedent two fields down —
`case authorHuman: conds = append(conds, "NOT "+provenanceExistsClause)`
(`store.go:404`). It is class-based rather than a generic `NOT`, but the
pattern SPEC-086 needs exists in the same struct.

### Rejected alternatives

- **A reserved tag** (`outcome:failed`, in the `agent:`/`model:` family).
  Framing was right that DEC-015 makes this far cheaper than an early input
  assumed — normalized `tags` + `taggings` with two indexes since 2026-06-06,
  so it is an indexed lookup and not a `LIKE` scan. Rejected anyway, and on
  the **read** path: tags do not render in the DEC-044 line, so the agent-
  facing surface this stage exists to serve would show nothing without
  amending the one line SPEC-073 stabilised.
- **A new column** (`outcome`, or a boolean). A forward-only migration
  (DEC-002) for a distinction `entries.type` already expresses, plus a new
  concept in every renderer and filter. Buys nothing the value does not.
- **Redefining `learned`.** Eight rows of live data, three of which already
  carry a failure as their subject. See the Context measurement.

---

## Fork 2 — the verb, the value, and the eight occupied rows

**Settled: verb `brag learn`, value `failed`, the eight rows untouched, and
NO validation of `type`.** The last clause is the one that decides the
complexity estimate, so it is stated first.

### 2(a) — this spec does NOT introduce `type` validation

Framing named this *"the single biggest complexity driver"* and it is. The
answer is **no closed set, no reserved-prefix rule, no validation on
`brag add`.**

The insight that makes this an M rather than an L: **pinning one value needs a
verb, not a validator.** `brag learn` always writes `failed` and offers no
`--type` flag, so the reserved value cannot fragment through the path people
will actually use — without touching `brag add`, whose `--type` stays free-form
and accepts anything it accepts today.

Validation was rejected on four measured grounds:

1. It is a **behaviour change on `brag add` for every existing user** and would
   need its own DEC — framing's own stated L-trigger.
2. A closed set would **reject `ship` (15 live entries) and `bugfix` (1)** on
   write. Those are the user's own near-duplicates; a validator's first act
   would be to break the habit that created them, without migrating them.
3. It owes an answer about the **113 entries with `type = ''`** — is empty
   legal? Today it is. Changing that is a separate decision.
4. It is **not necessary for the goal**. The goal is that a failure be
   greppable and retrievable. One verb that always writes one value achieves
   that for the value that matters.

The accepted cost, named: nothing stops `brag add --type failure`. The
mitigation is documentation (`BRAG.md`, and the `brag_add` MCP `type`
parameter description) plus the verb being the obvious path — not a gate. This
is recorded in DEC-049's Consequences, and guarded at the doc level by `AB2`
(the docs must not teach a second spelling), which is the only place a second
spelling can enter systematically.

### 2(b) — the value is `failed`

Non-colliding with all 18 live values, greppable, and it matches the field's
own house style: **7 of the 18 values are past participles** (`shipped`,
`learned`, `fixed`, `designed`, `documented`, `hardened`, `verified`). It also
reads correctly in DEC-044's `[<project>/<type>]` slot — `[bragfile/failed]`,
verified above.

### 2(c) — the verb/value seam, named rather than smoothed over

`brag learn` writes `type: failed`. Framing flagged this as a seam and asked
design to *"either align them or say why not."* **Not aligned, deliberately:**
the verb names what the **user** is doing (extracting the lesson); the value
names what is true of the **entry** (it did not work). They are different
words because the entry is both things.

The alternative — naming the verb for the value — was rejected on product
grounds: `brag fail` describes the entry accurately and discourages the use,
and the whole problem is that failures go unrecorded. The seam is documented
in the command's own `Long` string and in `docs/api-contract.md`, so a reader
meets it once rather than inferring it.

### 2(d) — the eight rows are left alone

Framing's position, adopted. Rewriting a user's corpus to fit a schema
decision is a cost this spec should not impose. Recorded in DEC-049's
Consequences as a known boundary: **a later mirror rule that reads `learned`
expecting failures will be wrong in eight places.**

### Rejected alternatives

- **A documented convention with no verb** — *"just type `brag add --type
  failed`"*. This is the option the corpus itself refutes: `shipped` 202 /
  `ship` 15 is what an unpinned convention produced here already.
- **`abandoned`** — narrower than the case. A wrong call carried to completion
  is not abandoned.
- **`dead-end`** — reads poorly in the `[project/type]` slot, and is the only
  candidate that would introduce a hyphen to a field that has none.
- **A hidden `--type` flag defaulted to `failed`.** Considered as a way to
  reuse `runAdd` wholesale. Rejected: a flag a user can still pass and have
  silently honoured is worse than no flag, and `getFlagString` returns `""`
  for an undefined flag rather than erroring, so the delegation would have
  been silently wrong if the flag were simply omitted.
- **`--json` mode on `learn`.** `brag add --json` with `"type":"failed"`
  already covers programmatic capture. A second JSON ingress is a second place
  to police the pinned value for no new capability.

---

## Fork 3 — how does a failure participate in `brag memory`?

**Settled: option 1 (do nothing to the read path) — but framing's reasoning is
overturned, and so is its ranking of the alternatives.**

### The measurement framing did not take

Framing wrote that a failure *"competes on recency alone"* and is *"crowded out
within weeks."* Measured on the live corpus, it is **stronger and different**:
it is not out-*ranked*, it is **not a candidate**.

`Gather` (`internal/memory/pool.go:59-90`) composes the pool from up to three
reads, the first of which is unconditional:

```go
entries, err := src.List(storage.ListFilter{Limit: PoolLimit})   // PoolLimit = 200
```

So a bare `brag memory` — no `--query`, no `--project`, which is how the
`brag://memory/recent` resource loads — sees **only the 200 most recent
entries**. Measured:

```
$ brag memory | head -6
Candidates: 200
$ sqlite3 corpus.sqlite "SELECT id, created_at FROM entries ORDER BY created_at DESC, id DESC LIMIT 1 OFFSET 199;"
207|2026-07-06T22:39:22Z
```

- corpus **398**; the pool horizon is entry **207**, dated **2026-07-06** —
  **61 days** before today
- **198 of 398 entries (49.7%)** cannot be returned by a bare `brag memory` at
  **any** budget
- of the 8 existing `type: learned` entries — the closest analogue the corpus
  has — **5 are already outside the pool** (ids 18/56/61/75/80, at recency
  ranks 383/349/344/330/325)

The included set is smaller still: a bare `brag memory` reports
`Included: 10, Skipped: 190` at the default 2000-token budget. **10 of 398.**

### Why framing's preferred alternative cannot work

Framing called option 2 — a fourth ordinal list — *"structurally consistent
with DEC-043, and the cleanest fit."* **It is structurally incapable of
surfacing the entries it was proposed to rescue**, and the code says so
directly:

```go
func buildMatchRank(pool []storage.Entry, matched []int64) map[int64]int {
	inPool := make(map[int64]bool, len(pool))
	for _, e := range pool { inPool[e.ID] = true }
	...
	for _, id := range matched {
		if !inPool[id] { continue }        // ← memory.go:191
```

`buildMatchRank` **drops ids that are not in the pool**, and `buildProjectRank`
iterates the pool. `Slice` receives only what `Gather` fetched. **No ranking
term can introduce an entry `Gather` did not read.** A fourth ordinal list
would re-order the most recent 200 and do nothing whatsoever for a failure
written 62 days ago — which is the case the option existed to solve.

The only option that reaches an out-of-pool entry is a **fourth read** in
`Gather` (e.g. `List{Type: FailureType, Limit: N}`). That is **pool
composition, not fusion** — and the distinction is what lets this spec answer
the open question honestly (below).

### Why this spec still does nothing

Not framing's reason (*"cannot be done honestly while the fusion constants are
an open question"* — that reason is void, since a fourth read touches no fusion
constant). Two better ones:

1. **A pool-composition rule has to choose how many slots to spend, and there
   are zero failure-typed entries to calibrate against.** Picking N against
   n=0 is precisely the mistake STAGE-023's own ordering call rejected for the
   impact classifier: a rule tuned on a population that excludes the case it
   exists to judge. The stage cannot make that argument to reorder itself and
   then commit it here.
2. **The blast radius is bigger than a memory change.** `Gather` is shared
   **verbatim** by the CLI and the three `brag://` MCP resources (SPEC-074
   LD9). A fourth read moves `Candidates:` on all four surfaces and their
   byte-exact goldens — and `Candidates:` is a count DEC-048 has *already*
   legislated once. That is its own spec.

**Named honestly: this means STAGE-023's success criterion 3 — *"`brag memory`
returns failures to an agent reading the corpus cold"* — is only half met by
this spec.** The half that IS met is the line shape: a failure in the pool is
labelled `[project/failed]`, verified by running it. The half that is not is
durability past the 61-day horizon. That is filed, with its measurement, as
`memory-pool-composition-excludes-older-entries`.

### The answer `memory-slice-fusion-constants` is owed, in both directions

Framing required an explicit answer either way, and named silence as the
failure mode. **The answer is: this work does NOT answer it, does NOT touch it,
and here is why the two are not the same question.**

| | `memory-slice-fusion-constants` (open since 2026-08-08) | `memory-pool-composition-…` (filed here) |
|---|---|---|
| function | `memory.Slice` | `memory.Gather` |
| subject | `k=60`, three unit weights (DEC-043 sub-decision 2) | `PoolLimit`, which reads compose the pool |
| question | is the *ordering* right at personal scale? | is the *candidate set* right at all? |
| this spec | **untouched** — k stays 60, weights stay 1.0, no fourth term | **filed, not answered** |

The distinction is load-bearing rather than pedantic: it is exactly what kills
framing's option 2. Conflating them would have produced a fourth ordinal list
that compounded an unresolved question *and* did not work.

`guidance/questions.yaml` records both halves — the fusion entry gains an
`EXAMINED AND LEFT OPEN at SPEC-085 design` note (so a still-open entry is
distinguishable from one nobody read), and the pool entry is filed with the
measurement above. Both are asserted by `AB4`, which fires if either is
removed.

### Rejected alternatives

- **Option 2, a fourth ordinal list.** Cannot reach an out-of-pool entry
  (`memory.go:191`). Framing ranked it "cleanest"; it is the one option that
  is provably ineffective.
- **Option 3, a guaranteed floor of N slice slots.** Touches DEC-044's budget,
  and has the same n=0 calibration problem as a fourth read with a larger
  blast radius (the budget is what `Included:`/`Skipped:` report).
- **A `--type` flag on `brag memory`.** Would make failures *findable* but
  only by someone who already suspects they exist — it does not help the cold
  read, which is the whole use case. It also turns the memory slice into a
  filtered query, which DEC-043 sub-decision 4 deliberately avoided for
  `--project`.
- **Raising `PoolLimit`.** Moves the horizon without addressing the shape, and
  changes `Candidates:` on four surfaces for every user.

---

## Fork 4 — what do the digests do with a failure? → **SPEC-086**

**Settled by the user at design, then split out.**

### The measurement

`brag wrapped`'s *"Impact moments"* and `brag impact`'s body are identical but
for one trailing blank line, on today's corpus:

```
$ diff wrapped_impact_section.txt impact_body.txt
697d696
<
$ grep -cE '^- [0-9]+: ' wrapped_impact_section.txt   # 324
$ grep -cE '^- [0-9]+: ' impact_body.txt              # 324
```

**324 of 398 entries carry an impact**, and both surfaces carry all 324 — inside
a digest whose own help calls it *"shareable, celebratory."* `impact`'s header
reads `Entries: 324/398 with impact`; `wrapped`'s reads `Entries: 398`.

**The finding framing did not have:** neither surface renders `type`, in
either format. `ToImpactMarkdown` (`internal/export/impact.go:60-61`) emits:

```go
fmt.Fprintf(&buf, "- %d: %s\n", e.ID, e.Title)
fmt.Fprintf(&buf, "  %s\n", e.Impact)
```

and the JSON projection is deliberately narrow — `impactEntry` is **4 keys**
(`id`, `title`, `project`, `impact`), DEC-028 choice 4. `wrapped` inlines the
same `aggregate.WithImpact` + `GroupEntriesByProject` shape. So a failure with
an impact statement is not merely *unflagged* there: it is
**unrepresentable as a failure** without a renderer change.

### The user's decision

Put to the user at design with the December reading spelled out, because
framing correctly identified this as a product call rather than a technical
one. **Chosen: a named `## What didn't work` section** — failures pulled out of
*Impact moments* / the impact body and rendered under their own heading on both
surfaces. Rejected: exclude-by-default, and include-silently.

The reasoning the user's choice follows: it is what the brief's *"this project
reads like a turnaround"* actually needs, since a turnaround has a downstroke;
exclusion would re-create in the digest the flattery the project exists to
remove; and silent inclusion is ruled out by STAGE-023's own success criterion.

### Why it splits rather than absorbs

The chosen option is a renderer change on two surfaces **plus two DEC-048
count renames** — once *Impact moments* stops carrying all with-impact
entries, its headline count no longer means what it says and must be **named
something else**, not silently redefined (`brag impact`'s
`Entries: <shown>/<in-window> with impact` is the precedent).

That is a second M of work, on different files, with its own decision record.
**STAGE-023's spec backlog already anticipated exactly this**, conditionally:

> *"(not yet written) — digest posture … **Only if SPEC-085's Fork 4 turns out
> to need more than a documented default.**"*

It does. The condition fired, so the spec is written: **SPEC-086**, framed
alongside this one, carrying the user's decision, the measurement, and the
DEC-048 obligation. Splitting keeps both at M; absorbing would have made this
an L, which framing's own guidance says to split rather than run.

**Interim risk, named:** between SPEC-085 shipping and SPEC-086 shipping, a
`failed` entry with an impact statement *would* appear silently in
*Impact moments*. Bounded and acceptable: the corpus holds zero such entries
today, the user controls when the first one is written, and the alternative
(shipping the digest change first) would mean designing a section for a row
type that does not exist yet.

---

## Fork 5 — the envelope

**Settled: a no-op for this spec, and binding on SPEC-086. Recorded so the
no-op is a checked result rather than an omission.**

`brag learn` is a capture verb, not a digest. It adds no `internal/export/`
renderer, emits no DEC-014 envelope, and prints no headline count — stdout is
the new entry's id alone, exactly as `brag add`. So:

- **DEC-014** — no new envelope surface. Nothing inherited, nothing to check.
- **DEC-048** — **no count changes anywhere.** Verified rather than asserted:
  the six `internal/export/` renderers are untouched by this spec, and
  `brag memory`'s `Candidates:` is unmoved because Fork 3 lands on "do
  nothing". The full test suite passes with no golden updated
  (15 packages `ok`).

The DEC-048 obligation lands squarely on **SPEC-086**, where a digest *does*
start excluding rows, and it is written into that spec's front-matter and its
acceptance criteria.

---

## §12(b) design-time pre-flight — what each tool actually said

Every literal this spec embeds was written into a detached `git worktree` at
`81e639d`, run through its own tool, and the worktree removed. Nothing in this
section is predicted. **Six findings below were caught here**, three of which
would have shipped a guard that could never fail (Findings 1, 2 and 3), and one
of which would have shipped a truncated file (Finding 3b).

| Tool | Version | What was run | Result |
|---|---|---|---|
| `gofmt -l .` | Go 1.26.x | full tree, all edits staged | **clean** |
| `go vet ./...` | Go 1.26.x | full tree | **clean** |
| `go build ./...` | Go 1.26.x | full tree | **OK** |
| `go test ./...` | Go 1.26.x | full tree | **15 packages `ok`**, 1 with no test files (`storagetest`) |
| `golangci-lint` (`just lint`) | v2.13.1 | full tree | **0 issues** |
| `./brag learn --help` | cobra | the locked `Long` + flag set | renders; flag set is exactly `-t -d -T -p -i` + `-h`, **no `--type`, no `-k`** |
| `./brag learn` (flag mode) | real binary, temp DB | end-to-end insert | wrote `type='failed'`; stdout `1` |
| `./brag list --type failed` | real binary | retrieval | returned the row — **no new retrieval code needed** |
| `./brag memory` | real binary | the DEC-044 line | `- 1 2026-09-05 [bragfile/failed] … — cost two days and produced nothing reusable` |
| `EDITOR=<stub> ./brag learn` | real binary | editor mode, buffer re-adding `Type: shipped` | stored **`failed`** — the pin holds across both ingress paths |
| `editor.Parse(FailureTemplate())` | stdlib `textproto` | unchanged-buffer abort | `parse buffer: Title header is required and must be non-empty` — same shape as `add`'s |
| MCP `tools/list` | go-sdk 1.7.0 | `brag_add.properties.type` read off a live server | description present, escaped quotes intact |
| `./scripts/inventory.sh` | — | regenerated and **diffed** against `main` | **7 rows move** (table below) |
| `./scripts/test-docs.sh` | — | full harness, all artifacts staged | **ALL OK**, 199 `OK:` lines / **198 distinct ids** |
| exit codes | — | `--type` passed, empty title, success | `1`, `1`, `0` |
| `git apply --check` | git | the three embedded diffs, extracted from this spec, against a pristine `81e639d` worktree | **all 7 file patches apply cleanly** |
| literal round-trip | — | the three embedded file literals re-extracted from this spec and compared to the pre-flighted files | **3 of 3 byte-identical** |

### Finding 1 — a new test-docs group appended to the end of the file NEVER RUNS

The natural way to add Group `AB` is to append it, which is exactly how Group
`AA` reads at the bottom of `scripts/test-docs.sh`. **Group `AA` is not the
last thing in the file.** There is a `# ===== finalise =====` block after it
that exits:

```sh
# ===== finalise =====
if [ "$FAIL_COUNT" -gt 0 ]; then ... exit 1; fi
ok "F4"
printf '\nALL OK: documentation-content assertions passed.\n'
exit 0
```

Appended after that, all ten `AB` assertions were **dead code**, and the
harness said so by saying nothing:

```
$ grep -n '===== finalise =====\|===== Group AB' scripts/test-docs.sh
1897:# ===== finalise =====
1910:# ===== Group AB — the reserved failure type (SPEC-085 / DEC-049) =====
$ ./scripts/test-docs.sh | grep -c 'OK:   AB'
0
$ ./scripts/test-docs.sh | grep -cE '^OK: '
189                                  ← unchanged from main
$ ./scripts/test-docs.sh | tail -1
ALL OK: documentation-content assertions passed.
```

**A green gate, an unchanged assertion count, and ten assertions that had never
executed.** This is the sharpest form of the trap the stage already warned
about — *a guard that is green is not evidence its claim is true* — because
here the guard was not merely weak, it was absent, and every signal available
said fine. Moving the group **above** the finalise block produced the ten `OK:`
lines. **LD8** pins the placement; build must insert, not append.

### Finding 2 — a loop-generated assertion id runs but is never counted

`AB2`'s three positive checks were first written as a loop:

```sh
for ab2_f in "BRAG.md" "README.md" "docs/api-contract.md"; do
    assert_contains_literal "AB2-pos-$ab2_f" "$ab2_f" "--type failed"
done
```

That runs correctly and emits three `OK:` lines. But `scripts/inventory.sh:67`
derives the doc-assertion count **statically**, by matching a *quoted literal*
id:

```sh
grep -oE '(^|[[:space:]])(ok|fail|skip|assert_[a-z_]+) "[A-Za-z0-9][A-Za-z0-9._-]*"'
```

`"AB2-pos-$ab2_f"` matches nothing. Measured: **198 ids emitted, 195 counted**
— the page would have under-reported by exactly 3, and `X3` would have stayed
green because the page is generated from the same filter that lost them. Fixed
by making the three ids literal (`AB2a`/`AB2b`/`AB2c`). This is the same class
as the S3 double-emit that `inventory.sh`'s own header comment documents, in
the opposite direction.

### Finding 3 — the NOT-contains needle had to be the command form, not the word

`AB2`'s negative asserts a second spelling has not entered the docs. The
obvious needle — the bare word `failure` — **is present in this spec's own
BRAG.md prose**, which names `failure` and `dead-end` as the spellings not to
invent, and uses "failure" in an ordinary sentence. Run at design against the
drafted prose *before* locking:

| needle | hits in `BRAG.md` + `README.md` |
|---|---|
| `--type failure` | **0** |
| `--type dead-end` | **0** |
| `--type abandoned` | **0** |
| `brag learn --type` | **0** |
| bare `failure` / `dead-end` | **2** (BRAG.md:147, 153 — line numbers in the POST-change file; the prose does not exist on `main`) |

So the needle is scoped to the invocation. This is §12's NOT-contains
self-audit landing exactly as designed: the assertion is true, the prose stays
free to name the wrong answers, and the naive version would have been red at
build.

### Finding 3b — a 3-backtick fence cannot hold a literal that contains fences

Caught by round-tripping the embedded literals back out of this spec and
diffing them against the pre-flighted files — **`DEC-049` came back at 77 lines
instead of 181.** Its own body contains fenced code blocks, and the first inner
` ``` ` closed the outer wrapper. Build would have transcribed a file truncated
mid-sentence.

The same defect, in a nastier form, in the embedded diffs: three *context*
lines carry a fence with **one leading space** (` ```bash` in `BRAG.md`'s
diff), and CommonMark treats up to three leading spaces as an ordinary fence.
Lines beginning `+```/-```` are safe; context lines are not.

Fixed by wrapping every literal that can contain a fence in a **4-backtick**
outer fence. Re-verified after the fix: 3 of 3 literals byte-identical, and all
7 file patches apply cleanly via `git apply --check` against a pristine
worktree. **The literal-artifact contract is only real if the literal survives
extraction** — checking that is cheap and was not otherwise part of the recipe.

### Finding 4 — `just inventory` does not regenerate the block

`CLAUDE.md` says *"`just inventory` regenerates the inventory block in
`docs/engineering-practices.md`; paste its output"*. The first half is not
true, and build will lose time to it:

```
$ grep -A 2 '^inventory:' justfile
inventory:
    @./scripts/inventory.sh
```

It **prints to stdout only**. Running it and re-running `test-docs` leaves `X3`
red with a full script-vs-page diff. The block must be replaced between the
`inventory:begin` / `inventory:end` markers explicitly — LD9 gives the exact
step.

### Finding 5 — the milestone line makes `brag learn` congratulate you for failing

Not a literal-validation finding but a design defect the pre-flight surfaced,
because writing `insertLearned` meant reading what `runAddFlags` does last:

```go
fmt.Fprintln(cmd.OutOrStdout(), inserted.ID)
emitMilestone(cmd, s, inserted)
```

Every milestone line is a congratulation (`milestone.go:60-78`): `🎉 %d brags
and counting — nice work!`, `🔥 %d-day streak! Keep it going.`, `🎯 %d brags
on %q — a story taking shape.`, `✨ First brag this week.` A verb whose entire
purpose is to remove flattery would have fired one at the moment a user records
a two-day dead end. Confirmed by mutation **M-C**, whose failure output is the
scenario itself:

```
--- FAIL: TestLearnCmd_MilestoneSuppressed_AddStillFires
    learn_test.go:100: brag learn must not congratulate; stderr = "🎉 10 brags and counting — nice work!\n"
```

`insertLearned` therefore does not call `emitMilestone` (**LD5**).

### Mutation checks — 8 probes, every mutant confirmed by content hash

`shasum -a 256` before and after each probe; restore from a `/tmp` backup,
never `git checkout`; hash re-checked after restore. Baselines:
`learn.go` `4ac076c1e792…`, `milestone.go` `a02b53479d37…`.

| # | Mutation | Hash moved | What fired | Restored |
|---|---|---|---|---|
| **M-A** | `FailureType` `"failed"` → `"failure"` | `4ac076c1` → `ec934a84` | `TestLearnCmd_PinsFailedType`: `FailureType = "failure", want "failed"` | ✔ `4ac076c1` |
| **M-B** | give `learn` a `--type`/`-k` flag | `4ac076c1` → `227e785e` | `TestLearnCmd_NoTypeFlag`, **both** sub-checks (`--type` and `-k`) | ✔ `4ac076c1` |
| **M-C** | add `emitMilestone` to `insertLearned` | `4ac076c1` → `09ae64c8` | `…MilestoneSuppressed…`: `stderr = "🎉 10 brags and counting — nice work!"` | ✔ `4ac076c1` |
| **M-D** | editor mode uses `parsed.Type` not `FailureType` | `4ac076c1` → `0b658c58` | `…EditorModeOverwritesUserType`: `Type = "shipped", want "failed"` | ✔ `4ac076c1` |
| **M-E** | change `add`'s milestone **wording** in `milestone.go` | `a02b5347` → `e7212bea` | `control failed: add should still fire the milestone` — **proves the positive half is live** | ✔ `a02b5347` |
| **M-F** | `BRAG.md` teaches `--type failure` | `d0b4618e` → `e23c38c6` | **three** at once: `AB1b` + `AB2a` (positive missing) **and** `AB2-neg` (negative present) | ✔ `d0b4618e` |
| **M-G** | drop the fusion-question note from the register | `7287299e` → `b7736cfc` | `AB4: … [the register does not record that SPEC-085 examined it]` | ✔ `7287299e` |
| **M-H** | remove the `**No milestone line.**` contract note | `1fc45113` → `3d776ec2` | `AB3c: docs/api-contract.md missing literal` | ✔ `1fc45113` |

**M-E is the one that matters most**, and it is the answer to the stage's
standing trap. `TestLearnCmd_MilestoneSuppressed_AddStillFires`'s negative half
(*learn writes nothing to stderr*) passes trivially if the milestone machinery
is broken, absent, or simply not triggered — a green that proves nothing. So the
same test drives the **same corpus, the same forced-TTY seam, and the same
10-entry threshold** through `brag add` and requires the milestone to appear.
M-E breaks `add`'s wording and the control goes red, which is what makes the
silence on the `learn` side evidence rather than coincidence.

**A first attempt at M-E was discarded rather than credited.** Replacing the
milestone's `return fmt.Sprintf(...)` with `return ""` left `t` unused and the
package failed to compile — `vet: declared and not used: t`. A red, on the
right test, for entirely the wrong reason, and indistinguishable from a real
one in a build log. Re-run as a wording-only change so the mutant tested the
assertion instead of the compiler. (AGENTS.md §12's third mutation clause,
earned again.)

### Inventory: regenerated and diffed, not predicted

Run as two generations of the same script — one on `main`, one in the
pre-flight worktree with every artifact staged — and diffed:

```
$ ./scripts/inventory.sh > /tmp/inv_main.txt              # on main @ 81e639d
$ cd <preflight-worktree> && ./scripts/inventory.sh > /tmp/inv_final.txt
$ diff /tmp/inv_main.txt /tmp/inv_final.txt
```

| Row | main (`81e639d`) | after SPEC-085 |
|---|---:|---:|
| Decision records | 47 | **48** |
| Go source files | 69 | **70** |
| Go test files | 78 | **79** |
| Go test functions | 815 | **820** |
| Documentation assertions (distinct ids) | 188 | **198** |
| Questions tracked in guidance/questions.yaml | 19 | **20** |
| …of those, still open | 6 | **7** |

Every other row is unchanged. **`Stages` stays at 22 and `Specs carried to ship
and archived` stays at 81** — adding SPEC-086 moves neither, because
`inventory.sh:60` counts only `projects/*/specs/done/`.

### §9(b) — the harness greps, by VALUE not by concept

Seven values move. Grepping `scripts/test-docs.sh` for each **number**:

| Value | Independently pinned by | Verified |
|---|---|---|
| 47 → 48 (decisions) | **`Y3`** (`'Decision records \| 47 \|'`, `decision-records!=47`) | fired at design: `FAIL: Y3 … decision-records!=47` |
| 19 → 20, 6 → 7 (questions) | **`Y4`** (both halves) | fired: `FAIL: Y4 … questions-total!=19 questions-open!=6` |
| 69, 78, 815, 188 | **nothing but `X3`** — no independent literal pin exists | grep returned no non-comment hits |

All three predicted failures were **observed**, not reasoned about: with the
artifacts staged and the pins unchanged, the harness reported exactly
`X3`, `Y3`, `Y4` and nothing else.

---

## Locked design decisions

**LD1 — a failure lives in `entries.type`, as the reserved value `failed`.**
Not a tag (invisible in DEC-044's line), not a column (a migration for a
distinction `type` already expresses), not `learned` (occupied by 8 rows).
Fork 1 / Fork 2(b). Recorded in **DEC-049**.

**LD2 — the verb pins the value; `type` is NOT validated.** `brag learn` has
no `--type` flag and no `-k` shorthand, and always writes `FailureType`.
`brag add --type` stays free-form and unchanged. This is the decision that
keeps the spec an M — pinning one value needs a verb, not a validator.
Fork 2(a).

**LD3 — `brag learn` has flag mode and editor mode, and no `--json` mode.**
`brag add --json` with `"type":"failed"` already covers programmatic capture;
a second JSON ingress is a second place to police the pinned value for no new
capability. Fork 2(a).

**LD4 — `editor.FailureTemplate()` OMITS the `Type:` header, and
`runLearnEditor` overwrites `parsed.Type` unconditionally.** The two halves are
one decision: a header a user can fill in that is then silently ignored is
worse than no header. The overwrite is still required because a user can
re-add the header by hand — verified end-to-end at design with a stub `$EDITOR`
that does exactly that. Fork 2(c).

**LD5 — `insertLearned` does NOT call `emitMilestone`.** Every milestone line
is a congratulation; firing one when a user records a failure is the flattery
this verb exists to remove. §12(b) Finding 5. Guarded by a **paired**
assertion, not a bare negative — see LD7.

**LD6 — the MCP counterpart is a description on `brag_add`'s existing `type`
parameter, not a sixth tool.** `brag_add` already accepts `type`, and that
field carries **no `jsonschema` description at all** on `main` — so an agent
gets zero signal today. One line names the reserved value. A `brag_learn` tool
would change SPEC-074's stabilised tool surface for no capability the `type`
parameter does not already have. Verified by reading
`brag_add.properties.type` off a live `tools/list`.

**LD7 — every negative assertion in this spec is paired with a positive that
would fail if the mechanism under test were simply absent.** The stage's
standing trap, closed rather than re-recorded:

| Negative | Its pair |
|---|---|
| `learn` writes nothing to stderr | **`add` must still fire `🎉 10 brags and counting`** on the same corpus, same TTY seam, same threshold (M-E proves the control is live) |
| `learn` has no `--type` flag | the five flags it **must** have are asserted present, structurally (`Flags().Lookup`), not by help-text substring |
| the docs contain no `--type failure`/`dead-end`/`abandoned` | `--type failed` **is** present in all three docs (`AB2a`/`AB2b`/`AB2c`) |
| — | `AB4` asserts the fusion question is **still open** *and* that the register **records why** — an open entry alone is indistinguishable from one nobody read |

**LD8 — Group `AB` is INSERTED above `# ===== finalise =====`, never
appended.** §12(b) Finding 1: appended, all ten assertions are dead code and
the harness still prints `ALL OK` at an unchanged 189 `OK:` lines. Build must
verify placement by running `grep -n '===== finalise =====\|===== Group AB'`
and confirming `AB` has the **lower** line number.

**LD9 — assertion ids are literal strings, never built from a loop variable.**
§12(b) Finding 2: `inventory.sh:67` extracts ids statically from quoted
literals, so a loop-generated id runs but is uncounted (198 emitted vs 195
counted, measured).

**LD10 — the inventory block is regenerated by pasting `./scripts/inventory.sh`
output between the markers.** `just inventory` only *prints* (§12(b)
Finding 4). Seven rows move; never hand-edit one. `X3` diffs the whole block
byte-for-byte.

**LD11 — Fork 4 is SPEC-086, not this spec.** The user chose a named
`## What didn't work` section, which is a renderer change on two surfaces plus
two DEC-048 count renames. STAGE-023's backlog pre-authorised exactly this
split.

### Rejected alternatives (build-time)

- **Delegating `runLearn` to `runAdd` with a hidden `--type` defaulted to
  `failed`.** Tempting (it inherits all three modes free) and wrong twice:
  `getFlagString` returns `""` for an *undefined* flag rather than erroring, so
  omitting the flag fails silently; and a hidden flag a user can still pass and
  have honoured re-opens the fragmentation LD2 exists to close.
- **Asserting the bare word `failure` is absent from the docs.** Red at build:
  `BRAG.md`'s own guidance names it as the wrong spelling. §12(b) Finding 3.
- **Appending Group `AB` at the end of `scripts/test-docs.sh`**, matching where
  Group `AA` visually sits. See LD8 — this is the single most likely build
  mistake in the spec, and it fails green.
- **Hand-editing the seven moved inventory rows** instead of pasting the
  script's output. `X3` recomputes the whole table; a hand-edit that gets six
  rows right fails with a full diff anyway.
- **Adding a `brag_learn` MCP tool.** See LD6.
- **Re-typing the eight `type: learned` entries.** Out of scope at stage level
  and explicitly closed at framing. The boundary is recorded in DEC-049.

---

## Outputs

### New files (3)

| Path | Lines | What |
|---|---:|---|
| `internal/cli/learn.go` | 166 | `NewLearnCmd` + `runLearn`/`runLearnFlags`/`runLearnEditor`/`insertLearned`, and the `FailureType` const. Embedded verbatim in *Notes for the Implementer* §1. |
| `internal/cli/learn_test.go` | 166 | 5 test functions. Embedded verbatim in §2. |
| `decisions/DEC-049-a-failure-is-a-reserved-type-value-pinned-by-a-verb.md` | 181 | The decision record. **Must carry `  type: decision`** in its front-matter or `Z7` fails and the inventory under-reports by one. Embedded verbatim in §3. **Corrected at verify:** the under-report is not *silent* — mutation-tested, `X3`, `Y3` and `Z7` all three fire (see *Verify*). |

### Modified files (11)

| Path | Change | Δ lines |
|---|---|---:|
| `internal/editor/editor.go` | `FailureTemplate()` — `EmptyTemplate()` minus the `Type:` header | +7 |
| `cmd/brag/main.go` | `root.AddCommand(cli.NewLearnCmd())`, after `NewAddCmd()` | +1 |
| `internal/mcpserver/server.go` | `addIn.Type` gains a `jsonschema` description naming the reserved value (LD6) | ±1 |
| `BRAG.md` | new `## When it didn't work — \`brag learn\`` section, before `## The command` | +30 |
| `README.md` | `brag learn` + read-back, before the editor-mode paragraph | +13 |
| `docs/api-contract.md` | `### \`brag learn\`` section, before `### \`brag list\`` | +45 |
| `AGENTS.md` | §11 Domain Glossary — a `learn` entry | +1 |
| `CHANGELOG.md` | `## [Unreleased]` → new `### Added` block | +17 |
| `guidance/questions.yaml` | the fusion-question note **+** the new pool-composition question | +54 |
| `scripts/test-docs.sh` | Group `AB` (10 ids) inserted above `finalise`; `Y3` re-pin 47→48; `Y4` re-pin 19→20 and 6→7 | +86 |
| `docs/engineering-practices.md` | the regenerated inventory block **only** (7 rows) | ±14 |

**Totals, measured in the pre-flight tree:** 3 new files (513 lines), 11
modified, **+259 / −11** on the modified set.

### Premise audit (§9), run at design against the repo

**The additive case** — a new subcommand is an addition to a tracked
collection, so every doc that enumerates commands is a planned update. The
sweep used a recently-added command as the probe rather than guessing:

```
$ grep -rln 'brag coverage' --exclude-dir=.git --exclude-dir=.claude \
      --exclude-dir=done --exclude-dir=reports .
```

| Hit class | Verdict |
|---|---|
| `README.md`, `docs/api-contract.md`, `docs/tutorial.md`, `CHANGELOG.md`, `AGENTS.md` | the five places a command lands. **EDIT all except `docs/tutorial.md`** — see below. |
| `BRAG.md` | **EDIT.** Not in the probe's hit list (`coverage`/`spark` are *read* verbs and BRAG.md is the capture guide) — but `brag learn` is a **capture** verb, so it belongs there. Found by asking what kind of command this is, not by the grep. |
| `decisions/DEC-0*.md` (3 hits) | **NO CHANGE.** Historical records of their own features. |
| `internal/aggregate/aggregate.go`, `internal/cli/coverage.go` etc. | **NO CHANGE.** Other commands' implementations. |
| `guidance/questions.yaml` | **EDIT**, for unrelated reasons (Fork 3). |
| `projects/**`, `docs/research/**` | **NO CHANGE.** Dated planning records. |

**`docs/tutorial.md` is deliberately NOT edited**, and this is a scope call
rather than an omission: the tutorial is a narrative walkthrough
(*"1. Check you're wired up" → "2. Capture your first brag" → …*), and every
command in it has a worked example with sample output. A `brag learn` section
would need a fabricated failure entry to show, in a document a new user reads
first. Routed to SPEC-086, which will have both surfaces settled and a reason
to show them together. **No assertion depends on the tutorial mentioning
`learn`** — verified: the full harness is green without it.

**The inversion case:** none. This spec adds a command and inverts no existing
behaviour. `brag add` is untouched — verified by the full suite passing with
no existing test modified or deleted.

**The literal-value case (§9(b)):** covered above in the pre-flight — `Y3` and
`Y4` are the two independent pins; `X3` covers the rest.

---

## Acceptance Criteria

Numbers to diff against, not prose. Every one was observed in the pre-flight
tree before being written here.

1. **The verb writes the reserved value.** `brag learn -t "x"` inserts a row
   with `entries.type = 'failed'`; stdout is the id alone; stderr is **empty**.
   `brag list --type failed` returns it.
2. **The value is pinned.** `NewLearnCmd().Flags().Lookup("type")` is `nil`
   and `ShorthandLookup("k")` is `nil`; the five flags `title`, `description`,
   `tags`, `project`, `impact` are all present. `brag learn --type shipped -t x`
   exits **1** with `unknown flag: --type`.
3. **Editor mode pins it too.** With a `$EDITOR` stub that writes
   `Title: t\nType: shipped\n\nbody\n`, the stored row has
   `type = 'failed'`, not `'shipped'`.
4. **No congratulation.** With the milestone TTY seam forced on and a 10-entry
   threshold crossed: `brag learn` writes **0 bytes** to stderr, while
   `brag add` on an identical corpus writes `🎉 10 brags and counting — nice
   work!`. Both halves in one test.
5. **`DEC-049` exists** and carries `  type: decision` (`Z7`).
6. **Group `AB` actually runs.** `grep -n '===== finalise =====\|===== Group AB'
   scripts/test-docs.sh` reports Group `AB` at the **lower** line number, and
   `./scripts/test-docs.sh | grep -c 'OK:   AB'` is **10**.
7. **`just test-docs` ALL OK** at **199 `OK:` lines / 198 distinct ids**
   (188 → 198; the +1 gap is S3's pre-existing double emit, unchanged).
8. **The inventory block is regenerated**, and
   `diff <(./scripts/inventory.sh) <(block between the markers)` is empty
   (`X3`). Exactly these seven rows moved: decisions **47→48**, Go source
   **69→70**, Go test files **78→79**, Go test functions **815→820**, doc
   assertions **188→198**, questions **19→20**, open questions **6→7**.
9. **`Y3` reads 48 and `Y4` reads 20 / 7.**
10. **The register answers the fusion question in both directions:**
    `memory-slice-fusion-constants` is **still `status: open`** and its notes
    contain `EXAMINED AND LEFT OPEN at SPEC-085 design`; a new entry
    `memory-pool-composition-excludes-older-entries` exists and is `open`.
    `guidance/questions.yaml` parses as YAML: **20 entries, 7 open**.
11. **No count anywhere else moves.** The six `internal/export/` renderers are
    untouched; `brag memory` still reports `Candidates: 200` on the live-shaped
    corpus; **no golden file is updated**.
12. **All five gates green**, with no test excluded and no lint suppression:
    `go test ./...` (15 packages `ok`), `gofmt -l .` empty, `go vet ./...`
    clean, `just lint` **0 issues**, `just test-docs` ALL OK.

---

## Failing Tests

Write them first, watch them fail for the **stated** reason, then make them
pass. Every literal below was run at design; exact text in *Notes for the
Implementer*.

### New — `internal/cli/learn_test.go` (5 functions)

| Test | Asserts | Fails before the change because |
|---|---|---|
| `TestLearnCmd_PinsFailedType` | the row's `Type == FailureType`, and `FailureType == "failed"` | `NewLearnCmd` / `FailureType` do not exist → compile error, then the value check |
| `TestLearnCmd_NoTypeFlag` | `Lookup("type") == nil`, `ShorthandLookup("k") == nil`, **and** the five real flags exist | as above; the positive half is what stops this passing on an empty command |
| `TestLearnCmd_MilestoneSuppressed_AddStillFires` | `learn` → `errBuf.Len() == 0`; **`add` → `errBuf` contains `🎉 10 brags and counting`** | as above; the second half is the control (M-E) |
| `TestLearnCmd_EditorModeOverwritesUserType` | a buffer carrying `Type: shipped` still stores `failed` | as above |
| `TestLearnCmd_EmptyTitleIsUserError` | flag mode with no `--title` errors with `--title is required` | as above |

### New — `scripts/test-docs.sh`, Group `AB` (10 ids)

| Id | Asserts | Fails before because |
|---|---|---|
| `AB1a` / `AB1b` | `BRAG.md` contains `brag learn` and `brag list --type failed` | neither exists in `BRAG.md` |
| `AB2a` / `AB2b` / `AB2c` | `--type failed` present in `BRAG.md`, `README.md`, `docs/api-contract.md` | absent from all three |
| `AB2-neg` | none of `--type failure` / `--type dead-end` / `--type abandoned` / `--type failed-work` appears in those three | passes vacuously before; **the paired positives are what give it meaning** |
| `AB3a` / `AB3b` / `AB3c` | `docs/api-contract.md` records the three deliberate omissions (no `--type` flag / no `--json` / no milestone) | the section does not exist |
| `AB4` | `memory-slice-fusion-constants` is `open` **and** the register records SPEC-085 examined it **and** the pool question is filed | the note and the new question do not exist |

### Changed — `Y3`

Re-pin `'Decision records | 47 |'` → `| 48 |` and `decision-records!=47` →
`!=48`. Observed at design before the re-pin:
`FAIL: Y3: inventory.sh row value(s) wrong: decision-records!=47`.

### Changed — `Y4`

Re-pin `'Questions tracked in guidance/questions.yaml | 19 |'` → `| 20 |` and
`'of those, still open | 6 |'` → `| 7 |`, with both `!=` strings. Observed:
`FAIL: Y4: inventory.sh row value(s) wrong: questions-total!=19 questions-open!=6`.
Add a re-pin note in the style of the existing SPEC-082/SPEC-083 notes: the
total moves because SPEC-085 **files** a question, and the open count moves
with it because that question is open — a deliberate corpus change, not drift.

### Changed — `X3`

No code change; the inventory block on `docs/engineering-practices.md` must be
regenerated (7 rows). Fails with a full script-vs-page diff until
`./scripts/inventory.sh` output is pasted between the markers. **`just
inventory` does not do this for you** (§12(b) Finding 4).

### Mutation checks (run by build, recorded in Build Completion)

All eight ran at design and are tabulated in the §12(b) section with their
hashes. Build re-runs them and pastes the output.

| # | Mutation | Expected |
|---|---|---|
| **M-A** | `FailureType` → `"failure"` | `TestLearnCmd_PinsFailedType` red |
| **M-B** | add `--type`/`-k` to `learn` | `TestLearnCmd_NoTypeFlag` red, both sub-checks |
| **M-C** | `emitMilestone` in `insertLearned` | `…MilestoneSuppressed…` red with the 🎉 line quoted |
| **M-D** | `Type: parsed.Type` in editor mode | `…EditorModeOverwritesUserType` red |
| **M-E** | change `add`'s milestone **wording** | `control failed: add should still fire the milestone` |
| **M-F** | `BRAG.md` teaches `--type failure` | `AB1b`, `AB2a`, `AB2-neg` all red |
| **M-G** | drop the fusion note from the register | `AB4` red |
| **M-H** | remove `**No milestone line.**` | `AB3c` red |

**Confirm every mutant by content hash** (`shasum -a 256` before and after);
`git diff --quiet` is blind to untracked files. **Restore from a `/tmp` backup,
never `git checkout`** — during build most files carry uncommitted work.
Confirm the hash returns to its pre-mutation value, and that the mutant changed
**only** what you meant: M-E's first form broke compilation (`declared and not
used: t`), which is a red for the wrong reason.

### Decision-to-test mapping (§9)

Every locked decision has an assertion that fails without it.

| Decision | Test |
|---|---|
| LD1 value is `failed` in `entries.type` | `TestLearnCmd_PinsFailedType`; M-A |
| LD2 no `--type`; no validation added | `TestLearnCmd_NoTypeFlag`; M-B. (No test asserts `brag add` rejects a type — deliberately: it must not.) |
| LD3 flag + editor mode, no `--json` | `TestLearnCmd_EmptyTitleIsUserError` (flag), `…EditorModeOverwritesUserType` (editor), `AB3b` (the contract records the omission) |
| LD4 template omits `Type:`, overwrite is unconditional | `…EditorModeOverwritesUserType`; M-D |
| LD5 no milestone | `…MilestoneSuppressed_AddStillFires`; M-C, M-E; `AB3c` |
| LD6 MCP description, not a sixth tool | existing `mcpserver` suite stays green (no tool-count change); `AB` does not assert it — the description is verified by the pre-flight `tools/list` read |
| LD7 every negative paired | M-E and M-F are the proofs: each fires a positive **and** a negative |
| LD8 Group `AB` above `finalise` | AC-6's `grep -n` + `grep -c 'OK:   AB'` = 10 |
| LD9 literal ids | AC-7's 198 distinct ids matching `inventory.sh`'s count |
| LD10 block regenerated | `X3` |
| LD11 Fork 4 is SPEC-086 | `SPEC-086` exists at `cycle: frame`; no `internal/export/` file is modified here |

---

## Implementation Context

### Decisions that apply

- **DEC-049** (this spec creates it) — the whole subject. Read it first; the
  forks above are its reasoning.
- **DEC-044** — the memory line renders `[<project>/<type>]`. **The fact that
  decides Fork 1**, and the reason a failure is labelled on the agent-facing
  surface for free.
- **DEC-043** — the RRF fusion. Relevant as the thing this spec does **not**
  touch; `k` and the three weights are unchanged.
- **DEC-015** — tags are normalized (`tags` + `taggings`). **Supersedes
  DEC-004.** Relevant only because it makes the rejected reserved-tag option
  cheap enough to have to reject on merit rather than cost.
- **DEC-014 / DEC-048** — the envelope and the count-naming rule. **No-ops for
  this spec** (Fork 5); binding on SPEC-086.
- **DEC-007 / DEC-009** — `add`'s flag-mode `--title` requirement and the
  editor buffer format that `FailureTemplate` derives from.
- **DEC-002** — migrations are forward-only. Why a new column was expensive.

### Constraints that apply

- **`one-spec-per-pr`** — one PR for this spec. SPEC-086 gets its own.
- **`no-sql-in-cli-layer`** — binds **production** code under `internal/cli/`
  (DEC-047 / SPEC-083). `learn.go` imports no SQL: persistence is
  `storage.Open` + `s.Add`, exactly as `add.go`. Verified — `just lint`
  reports 0 issues with depguard active on the new file.

### Prior related work

- **SPEC-073** stabilised the memory slice; **SPEC-074 LD9** made `Gather`
  shared verbatim between the CLI and the three `brag://` resources. Fork 3's
  "do nothing" exists to keep both stable, and the blast radius that follows
  from LD9 is why the pool fix is its own spec.
- **SPEC-083** — the mutation protocol (hash-confirm, restore from `/tmp`) and
  the Group-`AA` shape that Group `AB` follows.
- **`guidance/questions.yaml` → `memory-slice-fusion-constants`** — open,
  `medium`, raised 2026-08-08. Answered in both directions (Fork 3).

### Out of scope (for this spec specifically)

- **The digest posture.** → **SPEC-086**, with the user's decision recorded.
- **The memory pool fix.** → filed as
  `memory-pool-composition-excludes-older-entries`, with its measurement.
- **The impact-quality classifier** → STAGE-024. The mirror, story v2 →
  STAGE-025/026.
- **Retroactively re-typing the eight `type: learned` entries.** Closed at
  framing; the boundary is recorded in DEC-049.
- **`docs/tutorial.md`.** A scope call with a reason — see the premise audit.
- **`BRAG.md`'s read-path gap.** Routed at stage level to STAGE-024 framing.
  **But the stage's premise for that routing is wrong and should be corrected
  there** — measured on the live tree at design:

  | Stage/brief claim | Measured 2026-09-05 |
  |---|---|
  | *"`BRAG.md` has 22 headings"* | **26** markdown headings (16 at `##`), excluding `#` lines inside code fences |
  | *"none is a read-the-corpus section"* | **false** — `## Reading entries back` (line 382) lists 8 read commands across 31 lines |
  | *"`brag memory` appears twice, both in the plugin-install block"* | the literal string `brag memory` appears **zero** times; one mention of the `brag://` resources at line 215 |

  The gap is real but differently shaped: **the read section exists and omits
  the one command written for agents.** That is a smaller, sharper fix than
  "BRAG.md is write-only", and STAGE-024 should frame it against the measured
  version. (This spec adds a `brag list --type failed` line to BRAG.md's new
  capture section, which does not close that gap and is not claimed to.)

- **The stale `DEC-004` citations.** The `brag_add` MCP `tags` parameter cites
  `DEC-004` for the comma-joined format (`internal/mcpserver/server.go:74`).
  Checked at design: the citation is **stale but not false**. DEC-015
  superseded DEC-004's *storage model* only — DEC-015 itself keeps
  `Entry.Tags` as a comma-joined string for every read, so the *input format*
  claim survives its own supersession. There are **11 `DEC-004` citations in
  Go source**, and one file already shows the right idiom
  (`internal/story/thread.go:143`: `DEC-004/DEC-015`). **Routed, not absorbed:**
  it is an 11-site citation sweep about tags, in a spec about failures, and
  folding it in would violate this spec's own scope discipline. Worth its own
  chore.

---

## Notes for the Implementer

Every literal below was written into a detached worktree at `81e639d`, run
through its own tool, and the tool's output recorded in the §12(b) section.
**Transcribe verbatim; verify diffs against these.**

### Order of work

1. `internal/editor/editor.go` — `FailureTemplate()` (§4).
2. `internal/cli/learn.go` (§1), then `cmd/brag/main.go` (§5).
3. `internal/cli/learn_test.go` (§2). **Run `go test ./internal/cli` and
   confirm the five tests fail for the assertion you wrote**, not a stray
   compile error, before going further.
4. `internal/mcpserver/server.go` (§6).
5. `decisions/DEC-049-*.md` (§3).
6. Docs: `BRAG.md`, `README.md`, `docs/api-contract.md`, `AGENTS.md`,
   `CHANGELOG.md` (§7).
7. `guidance/questions.yaml` (§8).
8. `scripts/test-docs.sh` — Group `AB` **inserted above `finalise`** (LD8),
   `Y3` and `Y4` re-pins (§9).
9. **Last:** regenerate the inventory block (§10). It must be last — steps 5,
   8 and 3 all move rows in it.

### 1. `internal/cli/learn.go` (new, 166 lines)

```go
package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jysf/bragfile000/internal/capture"
	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/editor"
	"github.com/jysf/bragfile000/internal/storage"
)

// FailureType is the reserved entries.type value marking work that did not
// work (DEC-049). It is the ONE type value bragfile pins: `brag add --type`
// stays free-form, and `brag learn` exists so this value cannot fragment the
// way `shipped`/`ship` and `fixed`/`bugfix` already have in the live corpus.
const FailureType = "failed"

// learnFieldFlags are the entry-field flags whose presence routes `brag
// learn` to flag mode. "type" is deliberately ABSENT: the value is pinned,
// not chosen (DEC-049).
var learnFieldFlags = []string{"title", "description", "tags", "project", "impact"}

// NewLearnCmd builds `brag learn` — the capture verb for work that did not
// work. It is `brag add` with entries.type pinned to FailureType and the
// milestone nudge suppressed; see runLearn for why the nudge is dropped.
//
// Two modes, mirroring add's (DEC-007 / DEC-009) minus JSON mode:
//   - flag mode: any of the five entry-field flags set; --title is required.
//   - editor mode: no entry-field flag set; opens $EDITOR on a template whose
//     header block omits Type, because Type is not the user's to set here.
func NewLearnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "learn",
		Short: "Record work that did not work, as a first-class entry",
		Long: `Record work that did not work — a dead end, an abandoned approach, a wrong
call — as a first-class entry rather than contorting it into a win.

Every entry brag learn writes is typed "failed". That is the point of the
verb: the value is pinned, so these stay greppable instead of fragmenting
across near-duplicate spellings the way free-form types do. There is no type
flag here. Use brag add if you want to choose the type yourself.

Flag mode (any of -t, -d, -T, -p, -i set): inserts directly from flag values.
--title is required.

Editor mode (no entry-field flags set): opens $EDITOR on a template buffer.
Save a valid entry to insert it; save unchanged to abort cleanly. The buffer
has no Type header — brag learn sets it for you.

Read them back the same way you read anything else:
  brag list --type failed
  brag list --type failed --since 90d

Examples:
  brag learn                                        # editor mode
  brag learn -t "shared-worker pool did not cut cold starts"
  brag learn -t "tried a bloom filter on the tag join" \
             -i "cost two days and produced nothing reusable"
  brag learn --title "..." --description "..." --tags "..." \
             --project "..." --impact "..."

Short forms: -t title, -d description, -T tags, -p project, -i impact.`,
		RunE: runLearn,
	}
	cmd.Flags().StringP("title", "t", "", "short headline (required in flag mode)")
	cmd.Flags().StringP("description", "d", "", "free-form body — what you tried, and why it did not work")
	cmd.Flags().StringP("tags", "T", "", "comma-joined tag list (e.g. \"auth,perf\")")
	cmd.Flags().StringP("project", "p", "", "project / initiative this belongs to")
	cmd.Flags().StringP("impact", "i", "", "what it cost, or what it ruled out")
	return cmd
}

// runLearn dispatches to flag mode or editor mode. There is no JSON mode:
// `brag add --json` with "type":"failed" already covers programmatic capture,
// and a second JSON ingress would be a second place to police the pinned
// value (DEC-049).
func runLearn(cmd *cobra.Command, args []string) error {
	for _, name := range learnFieldFlags {
		if cmd.Flags().Changed(name) {
			return runLearnFlags(cmd, args)
		}
	}
	return runLearnEditor(cmd)
}

func runLearnFlags(cmd *cobra.Command, _ []string) error {
	title := getFlagString(cmd, "title")
	if strings.TrimSpace(title) == "" {
		return UserErrorf("--title is required and must not be empty")
	}
	return insertLearned(cmd, capture.Fields{
		Title:       title,
		Description: getFlagString(cmd, "description"),
		Tags:        getFlagString(cmd, "tags"),
		Project:     getFlagString(cmd, "project"),
		Type:        FailureType,
		Impact:      getFlagString(cmd, "impact"),
	}, cmd.Flags().Changed("project"))
}

func runLearnEditor(cmd *cobra.Command) error {
	editFn := testEditFunc
	if editFn == nil {
		editFn = editor.Default
	}
	edited, changed, err := editor.Launch(editor.FailureTemplate(), editFn)
	if err != nil {
		return fmt.Errorf("launch editor: %w", err)
	}
	if !changed {
		fmt.Fprintln(cmd.ErrOrStderr(), "Aborted.")
		return nil
	}
	parsed, err := editor.Parse(edited)
	if err != nil {
		return UserErrorf("invalid buffer: %v", err)
	}
	// Type is overwritten, not read. The template omits the header, but a
	// user who adds one back does not get to redirect the pinned value.
	return insertLearned(cmd, capture.Fields{
		Title:       parsed.Title,
		Description: parsed.Description,
		Tags:        parsed.Tags,
		Project:     parsed.Project,
		Type:        FailureType,
		Impact:      parsed.Impact,
	}, parsed.Project != "")
}

// insertLearned validates, opens the store, auto-fills the project and
// inserts. It deliberately does NOT call emitMilestone: every milestone line
// is a congratulation ("🎉 N brags and counting — nice work!", "🔥 N-day
// streak!"), and firing one at the moment a user records a failure is the
// flattery this verb exists to remove. Silence on stderr, the id on stdout.
func insertLearned(cmd *cobra.Command, f capture.Fields, projectSet bool) error {
	if err := capture.Validate(f); err != nil {
		return UserErrorf("%v", err)
	}
	dbFlag := getFlagString(cmd, "db")
	path, err := config.ResolveDBPath(dbFlag)
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}
	s, err := storage.Open(path)
	if err != nil {
		return err
	}
	defer s.Close()

	inserted, err := s.Add(storage.Entry{
		Title:       f.Title,
		Description: f.Description,
		Tags:        f.Tags,
		Project:     autoFillProject(s, f.Project, projectSet),
		Type:        f.Type,
		Impact:      f.Impact,
	})
	if err != nil {
		return fmt.Errorf("add entry: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), inserted.ID)
	return nil
}
```

### 2. `internal/cli/learn_test.go` (new, 166 lines)

```go
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jysf/bragfile000/internal/storage"
)

// newRootWithLearn mirrors newRootWithAdd: milestone TTY seam pinned off by
// default, learn + add + list registered so the milestone PAIR (below) can
// exercise both verbs against one corpus.
func newRootWithLearn(t *testing.T) (*cobra.Command, string) {
	t.Helper()
	addStderrIsTTY = func() bool { return false }
	t.Cleanup(func() { addStderrIsTTY = defaultStderrIsTTY })
	root := NewRootCmd("test")
	root.AddCommand(NewLearnCmd())
	root.AddCommand(NewAddCmd())
	root.AddCommand(NewListCmd())
	dbPath := filepath.Join(t.TempDir(), "test.db")
	return root, dbPath
}

// TestLearnCmd_PinsFailedType is the core claim: the verb writes the reserved
// value, and `brag list --type failed` gets it back (STAGE-023 criterion 2,
// satisfied by an existing flag).
func TestLearnCmd_PinsFailedType(t *testing.T) {
	root, dbPath := newRootWithLearn(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "learn", "--title", "shared-worker pool did not cut cold starts"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	id, err := strconv.ParseInt(strings.TrimSpace(outBuf.String()), 10, 64)
	if err != nil {
		t.Fatalf("stdout should be the ID alone, got %q", outBuf.String())
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	got, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get(%d): %v", id, err)
	}
	if got.Type != FailureType {
		t.Errorf("Type = %q, want %q", got.Type, FailureType)
	}
	if FailureType != "failed" {
		t.Errorf("FailureType = %q, want %q", FailureType, "failed")
	}
}

// TestLearnCmd_NoTypeFlag: the value is pinned, so the flag must not exist.
// A structural check on the flag set, not a substring check on help text —
// the Long deliberately DOES contain "--type" (it teaches `brag list --type
// failed`), so a NOT-contains here would be wrong.
func TestLearnCmd_NoTypeFlag(t *testing.T) {
	cmd := NewLearnCmd()
	if f := cmd.Flags().Lookup("type"); f != nil {
		t.Errorf("brag learn must not define --type, got %v", f)
	}
	if f := cmd.Flags().ShorthandLookup("k"); f != nil {
		t.Errorf("brag learn must not define -k (add's type shorthand), got %v", f)
	}
	for _, name := range []string{"title", "description", "tags", "project", "impact"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("brag learn must define --%s", name)
		}
	}
}

// TestLearnCmd_MilestoneSuppressed_AddStillFires is the PAIR. The negative
// alone proves nothing: stderr is also empty when no milestone would have
// crossed. So the same corpus, the same forced-TTY, the same threshold is
// driven through BOTH verbs — add fires, learn is silent.
func TestLearnCmd_MilestoneSuppressed_AddStillFires(t *testing.T) {
	// negative: learn is the 10th entry, TTY on, nothing on stderr
	root, dbPath := newRootWithLearn(t)
	seedEntries(t, dbPath, 9, "")
	setStderrIsTTY(t, true)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "learn", "--title", "tenth, and it did not work"})
	if err := root.Execute(); err != nil {
		t.Fatalf("learn: unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Errorf("brag learn must not congratulate; stderr = %q", errBuf.String())
	}

	// positive: same threshold, same TTY, via add — the milestone DOES fire,
	// which is what makes the silence above evidence of suppression.
	root2, dbPath2 := newRootWithLearn(t)
	seedEntries(t, dbPath2, 9, "")
	setStderrIsTTY(t, true)
	var outBuf2, errBuf2 bytes.Buffer
	root2.SetOut(&outBuf2)
	root2.SetErr(&errBuf2)
	root2.SetArgs([]string{"--db", dbPath2, "add", "--title", "tenth"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("add: unexpected error: %v", err)
	}
	if !strings.Contains(errBuf2.String(), "🎉 10 brags and counting") {
		t.Fatalf("control failed: add should still fire the milestone, got %q", errBuf2.String())
	}
}

// TestLearnCmd_EditorModeOverwritesUserType: the template omits Type, but a
// user who re-adds the header does not get to redirect the pinned value.
func TestLearnCmd_EditorModeOverwritesUserType(t *testing.T) {
	root, dbPath := newRootWithLearn(t)
	installAddEditFunc(t, func(path string) error {
		return os.WriteFile(path, []byte("Title: bloom filter on the tag join\nType: shipped\n\ntried it; the join was never the bottleneck\n"), 0o600)
	})
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "learn"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	id, err := strconv.ParseInt(strings.TrimSpace(outBuf.String()), 10, 64)
	if err != nil {
		t.Fatalf("stdout should be the ID alone, got %q", outBuf.String())
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	got, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get(%d): %v", id, err)
	}
	if got.Type != FailureType {
		t.Errorf("editor-mode Type = %q, want %q (user's header must be overwritten)", got.Type, FailureType)
	}
}

// TestLearnCmd_EmptyTitleIsUserError: flag mode requires --title, same as add.
func TestLearnCmd_EmptyTitleIsUserError(t *testing.T) {
	root, dbPath := newRootWithLearn(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "learn", "--impact", "cost two days"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected a user error, got nil")
	}
	if !strings.Contains(err.Error(), "--title is required") {
		t.Errorf("error = %v, want it to mention --title is required", err)
	}
}
```

### 3. `decisions/DEC-049-a-failure-is-a-reserved-type-value-pinned-by-a-verb.md` (new, 181 lines)

````markdown
---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-049                        # stable, never reused
  type: decision                     # decision | analysis | recommendation | observation
  confidence: 0.88                   # honest: the schema home is measured and
                                     # near-certain; the choice of the literal
                                     # value "failed" over "abandoned" is the
                                     # soft half.
  audience:
    - developer
    - agent

agent:
  id: claude-opus-5
  session_id: null

project:
  id: PROJ-008
repo:
  id: bragfile

created_at: 2026-09-05
supersedes: null
superseded_by: null

tags:
  - schema
  - capture
  - type
---

# DEC-049: a failure is a reserved `type` value, pinned by a verb

## Decision

Work that did not work is recorded as an ordinary entry whose
`entries.type` is the reserved value **`failed`**, written by a dedicated
verb **`brag learn`** that does not expose `--type`.

Three parts, and the third is the one that carries the weight:

1. **The schema home is `entries.type`** — not a reserved tag, not a new
   column.
2. **The value is `failed`** — a new value, not a redefinition of the
   occupied `learned`.
3. **The verb pins the value; the field is NOT validated.** `brag add
   --type` stays free-form and accepts anything, as it does today. This
   decision adds one reserved value and one verb that always writes it. It
   does **not** introduce a closed set, a reserved-prefix rule, or any
   validation on `brag add`.

## Context

Measured on the live corpus 2026-09-05 (398 entries, `main` at `81e639d`):

- **Zero entries record a failure.** Not suppression — there was no verb, no
  value, and no documented convention for one.
- `entries.type` is free-form. `internal/cli/add.go:101` calls it *"free-form
  category"* and nothing validates it. **18 distinct non-empty values** across
  398 entries, and **113 entries carry no type at all**.
- The field already demonstrates what an unpinned convention does:
  **`shipped` 202 / `ship` 15**, and **`fixed` 2 / `bugfix` 1**. Near-duplicate
  spellings of the same idea, in live data, with no mechanism to have stopped
  them.
- **`learned` is occupied** — 8 entries. Reading all 8, they are lessons, and
  three of them (ids 80, 361, 383) take a *failure as their subject* while
  framing it as a win recovered (*"Verify caught an unpinned test seam…"*).
  One (id 18) is a PROJ-001 smoke-test artifact and not a lesson at all.
  Redefining `learned` would therefore mean `brag list --type learned`
  returning a mix of wins and failures — which fails the stage's own
  criterion that a failure be *retrievable as a failure*.

The read path is what settles part 1. DEC-044's memory line
(`internal/memory/memory.go:224`) is:

```
- <id> <YYYY-MM-DD> [<project>/<type>] <title> — <impact>
```

**`type` renders there; tags do not.** So a failure marked by `type` is
labelled as a failure on the agent-facing read surface at zero cost, while a
failure marked by a reserved tag is invisible there unless DEC-044's locked
line shape is amended. Verified by running it, not by reading it:

```
- 1 2026-09-05 [bragfile/failed] shared-worker pool did not cut cold starts — cost two days and produced nothing reusable
```

Retrieval also needs no new code: `--type` is an exact-match inclusion filter
on **seven** commands (`list`, `export`, `summary`, `impact`, `wrapped`,
`story`, `coverage`), so `brag list --type failed` answers *"what has not
worked?"* the day the verb ships.

## Alternatives Considered

- **A reserved tag (`outcome:failed`), in the `agent:`/`model:` family.**
  Cheaper than an early framing input assumed — DEC-015 normalized tags into
  `tags` + `taggings` with two indexes on 2026-06-06, so a reserved tag is an
  indexed lookup, not a `LIKE` scan. Rejected anyway, and on the read path
  rather than the write path: **tags do not render in the DEC-044 memory
  line**, so the agent-facing surface this stage exists to serve would not
  show it without amending the one line SPEC-073 stabilised.
- **A new column (`outcome`, or a boolean).** A forward-only migration
  (DEC-002) for a distinction `entries.type` already expresses, plus a new
  concept in every renderer and filter. Bought nothing the `type` value does
  not.
- **Redefining `learned`.** Collides with 8 rows of live user data (above).
  Re-typing them was considered and rejected separately: rewriting a user's
  corpus to fit a schema decision is a cost this work should not impose.
- **A documented convention with no verb** — i.e. tell people to type
  `brag add --type failed`. Rejected on the corpus's own evidence: `shipped`
  202 / `ship` 15 is what an unpinned convention produces here. The verb *is*
  the enforcement, scoped to one value.
- **Validating `type` globally** — a closed set or a reserved-prefix rule.
  Rejected as out of proportion and out of scope: it is a behaviour change on
  `brag add` for every existing user, it would reject `ship` (15 live entries)
  and `bugfix` (1) on write, and it owes an answer about the 113 typeless
  entries. It is also unnecessary — pinning ONE value needs one verb, not a
  validator. This is the choice that kept SPEC-085 an M rather than an L.
- **`abandoned` / `dead-end` as the value.** `failed` matches the field's
  existing house style — 7 of the 18 live values are past participles
  (`shipped`, `learned`, `fixed`, `designed`, `documented`, `hardened`,
  `verified`) — and reads correctly in the memory line's `[project/type]`
  slot. `abandoned` is narrower than the case (a wrong call that was carried
  to completion is not abandoned); `dead-end` reads poorly in that slot.

## Consequences

- **The verb and the value are deliberately different words.** `brag learn`
  writes `type: failed`. The verb names what the user is doing (extracting the
  lesson); the value names what is true of the entry (it did not work). The
  seam is real and is documented in the command's own help rather than
  smoothed over: naming the verb `brag fail` would describe the entry and
  discourage the use.
- **`brag learn` emits no milestone line.** `brag add` prints a
  congratulatory nudge to stderr on a TTY. Firing `🎉 N brags and counting —
  nice work!` at the moment someone records a failure is precisely the
  flattery this work exists to remove, so `insertLearned` does not call
  `emitMilestone`. Guarded by a paired assertion — `learn` silent AND `add`
  still firing on the same corpus at the same threshold.
- **One reserved value, not a validated field.** An agent or a user can still
  write `brag add --type failure`, and nothing stops them. The mitigation is
  documentation (`BRAG.md`, the `brag_add` MCP `type` parameter description)
  plus the verb being the obvious path — not a gate. Accepted knowingly.
- **The 8 `type: learned` entries are untouched.** A later rule that reads
  `learned` expecting failures will be wrong in 8 places; that is recorded
  here so it is a known boundary rather than a surprise.
- **The read path is NOT fixed by this decision.** `brag memory`'s candidate
  pool is the 200 most recent entries, so a failure stops being a candidate
  once 200 newer entries exist — measured at a 61-day horizon on a 398-entry
  corpus, with 198 entries already outside it. See
  `memory-pool-composition-excludes-older-entries` in
  `guidance/questions.yaml`. The line shape labels a failure correctly; the
  pool decides whether it is there to label.

## Validation

- `brag learn -t "…"` writes `entries.type = "failed"`; `brag list --type
  failed` returns it.
- `NewLearnCmd().Flags().Lookup("type")` is nil, and so is
  `ShorthandLookup("k")` — a structural check, not a help-text substring
  check, because the help text deliberately *does* contain `--type` (it
  teaches `brag list --type failed`).
- An editor buffer with a user-added `Type: shipped` header still stores
  `failed`.
- With the milestone TTY seam forced on and a 10th-entry threshold crossed,
  `brag learn` writes nothing to stderr while `brag add` writes
  `🎉 10 brags and counting`.

## References

- DEC-002 — migrations are forward-only (why a new column is expensive).
- DEC-015 — tags are normalized; supersedes DEC-004. Why the reserved-tag
  option was costed honestly and still rejected.
- DEC-043 / DEC-044 — the memory slice and its line shape. DEC-044's line is
  the fact that decides the schema home.
- SPEC-085 — the spec that makes this decision.
- SPEC-086 — what the celebratory digests do with a `failed` entry. Decided
  by the user at SPEC-085 design; implemented there.
````

### 4. `internal/editor/editor.go` — insert directly after `EmptyTemplate()`

```go
// FailureTemplate returns the `brag learn` buffer: EmptyTemplate's headers
// MINUS Type, because `brag learn` pins that value (DEC-049) and a header the
// user can fill in but that is then overwritten is a silent-ignore trap.
func FailureTemplate() []byte {
	return []byte("Title: \nTags: \nProject: \nImpact: \n\n")
}
```

### 5. `cmd/brag/main.go` — one line, directly after `NewAddCmd()`

```go
	root.AddCommand(cli.NewAddCmd())
	root.AddCommand(cli.NewLearnCmd())
```

### 6. `internal/mcpserver/server.go` — replace the `Type` field of `addIn`

On `main` the field carries **no** `jsonschema` tag. Replace:

```go
	Type        string `json:"type,omitempty"`
```

with (one line, do not wrap):

```go
	Type        string `json:"type,omitempty" jsonschema:"free-form category (shipped, learned, ...). One value is reserved: \"failed\" marks work that did not work — the MCP counterpart of the brag learn verb (DEC-049). Use it verbatim; do not invent a spelling."`
```

Read back off a live `tools/list` at design as
`brag_add.properties.type.description`, escaped quotes intact.

### 7. Documentation — five hunks

Each was applied and harness-verified in the pre-flight tree. The full unified
diff for all of `BRAG.md`, `README.md`, `docs/api-contract.md`, `AGENTS.md` and
`CHANGELOG.md` follows; apply it as written.

````diff
diff --git a/AGENTS.md b/AGENTS.md
index 3cae5b3..ffcca1a 100644
--- a/AGENTS.md
+++ b/AGENTS.md
@@ -291,6 +291,7 @@ DECs are stable; specs come and go. DECs don't reciprocally list specs.
 - **Store** — the `*storage.Store` Go type that owns the `*sql.DB` and all typed methods. The only package that imports a SQL driver.
 - **migration** — a single `NNNN_*.sql` file under `internal/storage/migrations/`, embedded into the binary, applied automatically in lexical order on `storage.Open`.
 - **export** — a one-shot dump of entries, either as a Markdown report (stdout or `--out file.md`) or as a portable SQLite file copy (via `VACUUM INTO`).
+- **learn** — `brag learn`: the capture verb for work that did **not** work, and the only place `entries.type` is pinned rather than free-form. Writes the reserved type value `failed` (DEC-049) from flag mode or editor mode; no `--type` flag, no `-k`, no `--json`, and — deliberately — no milestone nudge on stderr, because every milestone line is a congratulation. `entries.type` stays free-form everywhere else: the verb pins one value, it does not validate the field. Retrieval is the existing `brag list --type failed`; the DEC-044 memory line renders `[<project>/failed]`, so a failure is labelled as one on the agent-facing read surface for free. PROJ-008 STAGE-023 (SPEC-085).
 - **review** — `brag review --week | --month`: prints recent entries grouped by project followed by three hard-coded reflection questions ("What pattern do you see in this period?", "What did you underestimate?", "What's missing here that should be?"). Markdown elides per-entry descriptions for compactness; JSON includes the full DEC-011 entry shape. Designed to be pasted into an external AI session for guided self-reflection. STAGE-004 (SPEC-019).
 - **summary** — a rule-based (non-LLM) aggregation of entries grouped by project/type over a rolling 7- or 30-day time window (`brag summary --range week|month`). STAGE-004.
 - **stats** — `brag stats`: six lifetime aggregations (total entries, entries/week rolling average, current streak, longest streak, top-5 most-common tags, top-5 most-common projects, corpus span). STAGE-004 (SPEC-020).
diff --git a/BRAG.md b/BRAG.md
index 1922e1f..45ee24d 100644
--- a/BRAG.md
+++ b/BRAG.md
@@ -123,6 +123,36 @@ as a brag?" is always better than posting noise.
 
 ---
 
+## When it didn't work — `brag learn`
+
+Not every session produces a win, and a corpus of only wins is a worse input
+to a retro, a review, and to you-in-a-future-session than an honest one. When
+the session burned real time on something that did not pan out — a dead end, an
+abandoned approach, a wrong call — propose a `brag learn` instead of stretching
+it into a `brag add`.
+
+```bash
+brag learn \
+  -t "Shared-worker pool did not cut cold starts" \
+  -p "project-name" \
+  -T "tag1,tag2" \
+  -i "Cost two days and produced nothing reusable; ruled the approach out." \
+  -d 'What was tried, and why it did not work.'
+```
+
+Same approval loop as `brag add` — propose, wait for the user, then run it.
+
+Every entry `brag learn` writes is typed `failed`. That value is reserved and
+pinned by the verb: there is no `--type` flag on `brag learn`, and you should
+not invent a spelling like `failure` or `dead-end` for it. Read them back with:
+
+```bash
+brag list --type failed
+```
+
+The `impact` field is still worth filling in — for a failure it answers *what
+did this cost, or what did it rule out?* rather than *what did this unlock?*
+
 ## The command
 
 ```bash
diff --git a/CHANGELOG.md b/CHANGELOG.md
index 56fcf7c..1c4fe0c 100644
--- a/CHANGELOG.md
+++ b/CHANGELOG.md
@@ -7,6 +7,23 @@ and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0
 
 ## [Unreleased]
 
+### Added
+
+- **`brag learn` — the corpus can hold work that did not work.** A capture
+  verb that writes the reserved entry type `failed`
+  ([DEC-049](decisions/DEC-049-a-failure-is-a-reserved-type-value-pinned-by-a-verb.md)).
+  Flag mode and editor mode, mirroring `brag add`; no `--type` flag, because
+  the value is pinned rather than chosen — the live corpus already shows what
+  happens to an unpinned convention (`shipped` 202 / `ship` 15, `fixed` 2 /
+  `bugfix` 1). `entries.type` stays free-form everywhere else; this adds one
+  reserved value, not validation. Retrieval needs no new flag:
+  `brag list --type failed`. The `brag_add` MCP tool's `type` parameter now
+  names the reserved value so an agent uses it verbatim.
+- `brag learn` prints **no milestone line**. `brag add`'s congratulatory
+  stderr nudge (`🎉 N brags and counting — nice work!`) is suppressed here;
+  congratulating someone for recording a failure is exactly the flattery the
+  verb exists to remove.
+
 ### Changed
 
 - **`brag memory`'s headline count is now `Candidates: <N>`, not
diff --git a/README.md b/README.md
index 90daf92..aae4423 100644
--- a/README.md
+++ b/README.md
@@ -98,6 +98,19 @@ brag add \
   --impact "unblocked mobile v3 release"
 ```
 
+Work that did **not** work gets its own verb, so the corpus is not a
+wins-only ledger:
+
+```bash
+brag learn -t "shared-worker pool did not cut cold starts" \
+           -i "cost two days and produced nothing reusable"
+brag list --type failed          # read them back
+```
+
+`brag learn` pins the entry type to `failed` — there is no `--type` flag on
+it — so these stay greppable instead of fragmenting across near-duplicate
+spellings.
+
 For longer narrative entries, `brag add` with no flags opens
 `$EDITOR` against a templated buffer:
 
diff --git a/docs/api-contract.md b/docs/api-contract.md
index f21eb12..109407d 100644
--- a/docs/api-contract.md
+++ b/docs/api-contract.md
@@ -130,6 +130,51 @@ resolved or matches no registered project, the entry is saved with an
 empty project, exactly as before. An explicit project always wins over
 the cwd. See `brag project here` (SPEC-031) for the shared resolver.
 
+### `brag learn` — capture work that did not work (STAGE-023)
+
+```
+brag learn [-t title] [-d description] [-T tags] [-p project] [-i impact]
+```
+
+`brag add` with `entries.type` pinned to the reserved value `failed`
+(DEC-049). Two modes, mirroring `add`'s:
+
+- **flag mode** — any of the five entry-field flags set; `--title` is
+  required and must be non-empty, else `UserError` (exit 1).
+- **editor mode** — no entry-field flag set; opens `$EDITOR` on
+  `editor.FailureTemplate()`, which is `EmptyTemplate()` **minus the `Type:`
+  header**. Save unchanged to abort cleanly (`Aborted.` on stderr, exit 0).
+
+**There is no `--type` flag, and no `-k` shorthand.** The value is pinned, not
+chosen — that is the whole reason the verb exists rather than a documented
+`brag add --type failed` convention. `entries.type` remains free-form
+everywhere else; this is the one reserved value, and `brag add --type` can
+still write anything.
+
+A `Type:` header that a user adds back to the editor buffer is **overwritten**,
+not honoured — a header you can fill in that is then silently ignored is worse
+than no header, which is why the template omits it.
+
+**No `--json` mode.** `brag add --json` with `"type":"failed"` already covers
+programmatic capture; a second JSON ingress would be a second place to police
+the pinned value.
+
+**No milestone line.** `brag add` prints a congratulatory nudge to stderr on a
+TTY (`🎉 N brags and counting — nice work!`). `brag learn` prints nothing to
+stderr on success — congratulating someone for recording a failure is the
+flattery this verb exists to remove. stdout is the new entry's id alone, as
+with `add`.
+
+Retrieval needs no new flag: `--type` is exact-match inclusion on five other
+read surfaces plus `export` and `coverage`, so
+
+```
+brag list --type failed
+brag list --type failed --since 90d
+```
+
+already answer *"what has not worked?"*.
+
 ### `brag list` — list entries
 
 ```
````

### 8. `guidance/questions.yaml` — two hunks

A note appended to `memory-slice-fusion-constants`'s existing `notes:` block
(the entry stays `status: open`), and a new question appended at end of file.
Both in the diff below. **Parse-check after applying** — the pre-flight ran it
through PyYAML: **20 entries, 7 open**.

````diff
diff --git a/guidance/questions.yaml b/guidance/questions.yaml
index 7ae12af..ddf6da3 100644
--- a/guidance/questions.yaml
+++ b/guidance/questions.yaml
@@ -631,6 +631,15 @@ questions:
       (Above §14's 0.7 hard threshold; filed proactively per DEC-031's
       precedent of recording the softest sub-choice.)
 
+      EXAMINED AND LEFT OPEN at SPEC-085 design (2026-09-05). STAGE-023 owed
+      this register an answer in both directions, so: SPEC-085 does NOT answer
+      it and does NOT touch it. k stays 60, the three weights stay 1.0, no
+      fourth fusion term is added. The gap SPEC-085 measured on the read path
+      is in Gather's POOL, not in Slice's fusion — see
+      memory-pool-composition-excludes-older-entries, filed alongside this
+      note. Recording the non-answer explicitly because silence here is
+      indistinguishable from having forgotten to look.
+
   - id: capture-caps-were-inherited-not-derived
     question: "The capture.Validate field caps (title 200, tags 64, impact 256) were inherited from the add --json schema and never derived from real use — which of them are load-bearing, which are arbitrary, and what shape should they take?"
     priority: medium
@@ -964,3 +973,48 @@ questions:
       opened list_test.go, could not write a true comment without an answer,
       and the user decided the scope question directly. DEC-047 records it and
       names the revisit triggers for the cost that resolution (A) accepts.
+
+  - id: memory-pool-composition-excludes-older-entries
+    question: "brag memory's candidate pool is the 200 most recent entries (Gather's List{Limit: PoolLimit}) plus the query/project reads — so on a corpus of 398 growing ~20/week, half the corpus is not a candidate at ALL for a bare `brag memory`. Should pool composition be type-aware, and if so how many slots?"
+    priority: medium
+    status: open
+    raised_by: claude-opus-5
+    raised_at: 2026-09-05
+    assigned_to: null
+    notes: |
+      Filed at SPEC-085 design, from a measurement framing did not take.
+      Framing's Fork 3 said a failure "competes on recency alone" and is
+      "crowded out". Measured on the live corpus 2026-09-05, it is stronger
+      and different: it is not out-RANKED, it is not a CANDIDATE.
+
+        - corpus 398; a bare `brag memory` reports `Candidates: 200`
+        - the 200th most recent entry is id 207, 2026-07-06 — a 61-day horizon
+        - 198 of 398 entries (49.7%) cannot be returned by a bare
+          `brag memory` at any budget
+        - 5 of the 8 existing `type: learned` entries are already outside it
+          (ids 18/56/61/75/80 at ranks 383/349/344/330/325)
+
+      WHY THIS IS NOT memory-slice-fusion-constants. That question is about
+      Slice's FUSION — k=60 and the three unit weights (DEC-043 sub-decision
+      2). This one is about Gather's POOL. They are different functions and
+      different decisions, and the distinction is load-bearing because it
+      kills the option that looks cleanest: a fourth ordinal list CANNOT fix
+      this. buildMatchRank skips ids not in the pool (`if !inPool[id] {
+      continue }`) and buildProjectRank iterates the pool, so no ranking term
+      can introduce an entry Gather did not fetch. The only fix that reaches
+      an out-of-pool entry is a fourth READ.
+
+      WHY IT IS NOT ANSWERED HERE: a pool-composition rule has to choose how
+      many slots to spend, and there are ZERO failure-typed entries to
+      calibrate against. That is the same mistake STAGE-023's own ordering
+      call rejected for the impact classifier — a rule tuned on a population
+      that excludes the case it exists to judge. SPEC-085 ships the capture
+      verb and deliberately leaves the read path alone.
+
+      RESOLVE WHEN: the corpus holds enough `type: failed` entries that a
+      slot count can be argued from data rather than picked, or the first
+      time a user reports a failure they know they recorded and `brag memory`
+      will not return. Note the blast radius: Gather is shared VERBATIM by
+      the CLI and the three brag:// MCP resources (SPEC-074 LD9), so any
+      change moves `Candidates:` on all four surfaces and their byte-exact
+      goldens.
````

### 9. `scripts/test-docs.sh` — Group `AB` + two re-pins

**Placement is load-bearing (LD8).** Insert Group `AB` **immediately before**
the `# ===== finalise =====` line. Appending it to the end of the file makes
all ten assertions dead code while the harness still prints `ALL OK` — proven
at design, §12(b) Finding 1. After applying, confirm:

```
$ grep -n '===== finalise =====\|===== Group AB' scripts/test-docs.sh
   # Group AB must have the LOWER line number
$ ./scripts/test-docs.sh | grep -c 'OK:   AB'
10
```

````diff
diff --git a/scripts/test-docs.sh b/scripts/test-docs.sh
index 41c1de2..2807076 100755
--- a/scripts/test-docs.sh
+++ b/scripts/test-docs.sh
@@ -1601,7 +1601,7 @@ if [ ! -x scripts/inventory.sh ]; then
 else
     y3_out=$(./scripts/inventory.sh)
     y3_bad=""
-    printf '%s\n' "$y3_out" | grep -F -q 'Decision records | 47 |' || y3_bad="$y3_bad decision-records!=47"
+    printf '%s\n' "$y3_out" | grep -F -q 'Decision records | 48 |' || y3_bad="$y3_bad decision-records!=48"
     printf '%s\n' "$y3_out" | grep -F -q 'Decision numbers reserved, not yet decided | 1 |' || y3_bad="$y3_bad reserved-decisions!=1"
     if [ -z "$y3_bad" ]; then
         ok "Y3"
@@ -1627,8 +1627,8 @@ if [ ! -x scripts/inventory.sh ]; then
 else
     y4_out=$(./scripts/inventory.sh)
     y4_bad=""
-    printf '%s\n' "$y4_out" | grep -F -q 'Questions tracked in guidance/questions.yaml | 19 |' || y4_bad="$y4_bad questions-total!=19"
-    printf '%s\n' "$y4_out" | grep -F -q 'of those, still open | 6 |' || y4_bad="$y4_bad questions-open!=6"
+    printf '%s\n' "$y4_out" | grep -F -q 'Questions tracked in guidance/questions.yaml | 20 |' || y4_bad="$y4_bad questions-total!=20"
+    printf '%s\n' "$y4_out" | grep -F -q 'of those, still open | 7 |' || y4_bad="$y4_bad questions-open!=7"
     if [ -z "$y4_bad" ]; then
         ok "Y4"
     else
@@ -1894,6 +1894,86 @@ else
     fail "AA3" "a claim SPEC-083 corrected is back in the tree:$aa3_bad"
 fi
 
+# ===== Group AB — the reserved failure type (SPEC-085 / DEC-049) =====
+#
+# SPEC-085 adds ONE reserved value to a field that is otherwise free-form, and
+# pins it with a verb instead of a validator. Nothing in the binary stops a
+# second spelling: `brag add --type failure` is legal and always will be. So
+# the guard that matters is not "is the value valid" (it cannot be) but "do the
+# docs and the binary still name the SAME value" — because the moment a doc
+# teaches `--type failure`, the convention has forked exactly the way
+# `shipped`/`ship` (202/15) and `fixed`/`bugfix` (2/1) already did in the live
+# corpus, which is the evidence DEC-049 rests on.
+#
+# WHY THE NEEDLES ARE COMMAND FORMS, NOT BARE WORDS. Asserting that the word
+# "failure" is absent from BRAG.md would fail on BRAG.md's own guidance, which
+# names "failure" and "dead-end" as the spellings NOT to invent (BRAG.md:147)
+# and uses "failure" in ordinary prose (BRAG.md:153). Verified at design by
+# grepping the drafted prose for each needle before locking: the four `--type
+# <wrong>` forms return 0 hits while the bare words return 2. Scope the needle
+# to the invocation and the assertion stays true without muzzling the prose.
+
+# AB1 — the agent-facing capture guide teaches the verb AND the read-back. The
+# read-back half is the one that rots: BRAG.md documented capture for four
+# months while `brag memory` appeared nowhere in it.
+assert_contains_literal "AB1a" "BRAG.md" "brag learn"
+assert_contains_literal "AB1b" "BRAG.md" "brag list --type failed"
+
+# AB2 — ONE spelling, checked in both directions in the same group. The
+# positives are what make the negatives mean something: a NOT-contains alone
+# passes trivially on a file that never mentions the subject at all.
+# Ids are LITERAL, not loop-generated. scripts/inventory.sh derives the
+# doc-assertion count statically, by matching a quoted literal id in this
+# file, so an id built from a loop variable runs but is never counted — the
+# page would under-report by exactly the number of such ids (measured at
+# design: 198 emitted vs 195 counted). Distinct-ids is the number X3 pins.
+assert_contains_literal "AB2a" "BRAG.md" "--type failed"
+assert_contains_literal "AB2b" "README.md" "--type failed"
+assert_contains_literal "AB2c" "docs/api-contract.md" "--type failed"
+ab2_bad=""
+for ab2_f in "BRAG.md" "README.md" "docs/api-contract.md"; do
+    for ab2_n in "--type failure" "--type dead-end" "--type abandoned" "--type failed-work"; do
+        if grep -F -q -- "$ab2_n" "$ab2_f" 2>/dev/null; then
+            ab2_bad="$ab2_bad [$ab2_f teaches '$ab2_n']"
+        fi
+    done
+done
+if [ -z "$ab2_bad" ]; then
+    ok "AB2-neg"
+else
+    fail "AB2-neg" "a second spelling of the reserved type has entered the docs:$ab2_bad"
+fi
+
+# AB3 — the three deliberate OMISSIONS stay documented. Each is a thing a
+# later contributor would plausibly "fix" by adding it back; each is recorded
+# in DEC-049 as a choice. Naming them in the contract means restoring one has
+# to edit this file too, rather than looking like an oversight being tidied.
+assert_contains_literal "AB3a" "docs/api-contract.md" "There is no \`--type\` flag, and no \`-k\` shorthand."
+assert_contains_literal "AB3b" "docs/api-contract.md" "**No \`--json\` mode.**"
+assert_contains_literal "AB3c" "docs/api-contract.md" "**No milestone line.**"
+
+# AB4 — STAGE-023 owed guidance/questions.yaml an answer on
+# memory-slice-fusion-constants IN BOTH DIRECTIONS. The failure mode named at
+# framing was silence, so this asserts the non-answer is RECORDED: the entry is
+# still open (SPEC-085 did not resolve it) and its notes say SPEC-085 looked.
+# An entry that is merely still open is indistinguishable from one nobody read.
+ab4_bad=""
+ab4_status=$(awk -v want="  - id: memory-slice-fusion-constants" '
+    $0 == want { f=1; next }
+    f && /^  - id: / { exit }
+    f && /^    status: / { print; exit }
+' guidance/questions.yaml)
+[ "$ab4_status" = "    status: open" ] || ab4_bad="$ab4_bad [fusion-constants is no longer open: '$ab4_status']"
+grep -F -q "EXAMINED AND LEFT OPEN at SPEC-085 design" guidance/questions.yaml \
+    || ab4_bad="$ab4_bad [the register does not record that SPEC-085 examined it]"
+grep -F -q "  - id: memory-pool-composition-excludes-older-entries" guidance/questions.yaml \
+    || ab4_bad="$ab4_bad [the pool-composition question was not filed]"
+if [ -z "$ab4_bad" ]; then
+    ok "AB4"
+else
+    fail "AB4" "the fusion-constants answer-in-both-directions is not recorded:$ab4_bad"
+fi
+
 # ===== finalise =====
 
 if [ "$FAIL_COUNT" -gt 0 ]; then
````

### 10. `docs/engineering-practices.md` — regenerate the inventory block, LAST

`just inventory` **prints only** (§12(b) Finding 4). Replace everything between
the `inventory:begin` and `inventory:end` markers with the script's output:

```sh
./scripts/inventory.sh   # then paste between the markers, byte for byte
```

Do this **after** DEC-049, Group `AB` and `learn_test.go` exist — all three
move rows. Never hand-edit a row: `X3` diffs the whole block byte-for-byte.
Seven rows should move, and only these seven:

```
| Decision records | 47 | → | 48 |
| Go source files | 69 | → | 70 |
| Go test files | 78 | → | 79 |
| Go test functions | 815 | → | 820 |
| Documentation assertions (distinct ids) | 188 | → | 198 |
| Questions tracked in guidance/questions.yaml | 19 | → | 20 |
| …of those, still open | 6 | → | 7 |
```

---

## Build Completion

Build ran from `origin/main` at `2913c80` (design PR #196 had already merged;
the scaffold-dirs fix PR #195 was also already in). Every number below was
re-derived on the build tree, not copied from design.

- **Deviations from the spec:** **none.** All three new-file literals
  (`internal/cli/learn.go`, `internal/cli/learn_test.go`, `decisions/DEC-049-*.md`)
  and all applied diffs (`AGENTS.md`, `BRAG.md`, `CHANGELOG.md`, `README.md`,
  `docs/api-contract.md`, `guidance/questions.yaml`, `scripts/test-docs.sh`)
  were transcribed byte-for-byte and applied without modification. Confirmed,
  not assumed: `learn.go` and `milestone.go`'s pre-mutation `shasum -a 256`
  matched the spec's recorded design-time baselines exactly
  (`4ac076c1e792…` / `a02b53479d37…`), which is only possible if the
  transcription introduced zero drift.
- **New DEC-\* files created:** `DEC-049` only — the one this spec creates.
  Front-matter carries `  type: decision` (`Z7` requires it; confirmed by
  inspection and by `Z7` passing in `just test-docs`).
- **Constraints checked:** `one-spec-per-pr` (this PR carries SPEC-085 only;
  SPEC-086 is untouched, still `cycle: frame`); `no-sql-in-cli-layer`
  (`learn.go` imports no SQL driver — `storage.Open` + `s.Add`, exactly as
  `add.go`; `just lint`'s depguard reports 0 issues on the new file).
- **Gates, all five green:**
  - `go test ./...` — **1070 passed in 15 packages** (`storagetest` carries no
    test files, as on `main`).
  - `gofmt -l .` — empty.
  - `go vet ./...` — clean.
  - `just lint` — **0 issues**.
  - `just test-docs` — **ALL OK**, **199 `OK:` lines / 198 distinct ids**
    (188 → 198; the +1 gap is the pre-existing S3 double-emit, unchanged from
    `main`). Group `AB` confirmed running, not merely present:
    `grep -n '===== finalise =====\|===== Group AB' scripts/test-docs.sh`
    reports Group AB at line 1897, `finalise` at line 1977 (AB strictly
    lower), and `./scripts/test-docs.sh | grep -c 'OK:   AB'` returns **10**.
- **Inventory: regenerated, diffed against a saved pre-build baseline, not
  predicted.** `./scripts/inventory.sh` run once on the branch point (before
  any artifact existed) and once after every artifact was staged; `diff`
  between the two shows **exactly the seven predicted rows** moved and
  nothing else: Decision records 47→48, Go source files 69→70, Go test files
  78→79, Go test functions 815→820, Documentation assertions 188→198,
  Questions 19→20, open questions 6→7. `Stages` held at 22 and `Specs carried
  to ship and archived` held at 81, as predicted. `X3` verified directly
  (not just by the harness): `diff <(./scripts/inventory.sh) <(awk
  '/inventory:begin/{f=1;next}/inventory:end/{f=0}f'
  docs/engineering-practices.md)` is empty.
- **`guidance/questions.yaml` parses as YAML:** 20 entries, 7 open (via
  `uv run --with pyyaml python3` — no system PyYAML was installed; used `uv`
  to avoid touching the externally-managed Python env). Matches AC-10 exactly.
- **Mutation checks M-A…M-H: all eight fired, each mutant confirmed present by
  `shasum -a 256` before the test ran and confirmed reverted to the exact
  pre-mutation hash afterward — never by `git diff`.** Full protocol (backup
  to a session scratchpad, not bare `/tmp`, before each probe; restore from
  that backup, never `git checkout`, since the working tree carries the
  spec's own uncommitted artifacts throughout):
  - **M-A** `FailureType` `"failed"` → `"failure"`. Hash `4ac076c1` →
    `ec934a84` (matches design's recorded hash exactly).
    `TestLearnCmd_PinsFailedType` red: `FailureType = "failure", want
    "failed"`. Restored to `4ac076c1`.
  - **M-B** added `--type`/`-k` to `learn`. Hash `4ac076c1` → `35739f1f`
    (design recorded `227e785e` — same mutation class, different literal
    flag description text, so a different hash; the assertion behavior is
    what's being probed, and it matched). `TestLearnCmd_NoTypeFlag` red on
    **both** sub-checks (`--type` present, `-k` present). Restored to
    `4ac076c1`.
  - **M-C** added `emitMilestone` to `insertLearned`. Hash `4ac076c1` →
    `09ae64c8` (matches design's recorded hash exactly).
    `…MilestoneSuppressed…` red: `stderr = "🎉 10 brags and counting — nice
    work!\n"`. Restored to `4ac076c1`.
  - **M-D** editor mode used `parsed.Type` instead of `FailureType`. Hash
    `4ac076c1` → `0b658c58` (matches design's recorded hash exactly).
    `…EditorModeOverwritesUserType` red: `Type = "shipped", want "failed"`.
    Restored to `4ac076c1`.
  - **M-E** changed `add`'s milestone **wording** in `milestone.go`
    (`"…nice work!"` → `"…banked — nice work!"`, a wording-only edit so `t`
    stays referenced and the package still compiles). Hash `a02b5347` →
    `50ab5591` (design's own literal wording differed, hence a different
    hash than design's recorded `e7212bea`; the mechanism is what's probed).
    Fired: `control failed: add should still fire the milestone, got "🎉 10
    brags banked — nice work!\n"` — **the positive control is live.**
    Restored to `a02b5347`.
  - **M-F** `BRAG.md` line 150 changed `brag list --type failed` →
    `brag list --type failure`. Hash `d0b4618e` → `e23c38c6` (matches
    design's recorded hash exactly). **Three** assertions fired at once, as
    predicted: `AB1b` (`brag list --type failed` no longer present), `AB2a`
    (`--type failed` no longer present in `BRAG.md`), `AB2-neg`
    (`--type failure` now present). Restored to `d0b4618e`.
  - **M-G** dropped the `EXAMINED AND LEFT OPEN at SPEC-085 design` note from
    `guidance/questions.yaml`. Hash `7287299e` → `cb0e398f` (design's own
    note text differed slightly in whitespace, hence a different hash than
    design's recorded `b7736cfc`; message is what's probed). Fired: `AB4:
    … [the register does not record that SPEC-085 examined it]`. Restored to
    `7287299e`.
  - **M-H** removed the `**No milestone line.**` sentence from
    `docs/api-contract.md`. Hash `1fc45113` → `d1f853cd` (matches design's
    recorded pre-mutation hash exactly; post-mutation hash not independently
    recorded by design). Fired: `AB3c: docs/api-contract.md missing literal:
    **No milestone line.**`. Restored to `1fc45113`.

  Every one of the five files touched by a mutation (`learn.go`,
  `milestone.go`, `BRAG.md`, `guidance/questions.yaml`,
  `docs/api-contract.md`) was re-hashed after the full mutation sweep and
  confirmed identical to its pre-mutation value — no probe left residue.
- **End-to-end verification, temp DB (`--db`), not the live corpus:**
  `brag learn -t "shared-worker pool did not cut cold starts" -p bragfile -i
  "cost two days and produced nothing reusable"` → stdout `1`; `brag memory`
  renders `- 1 2026-09-05 [bragfile/failed] shared-worker pool did not cut
  cold starts — cost two days and produced nothing reusable`, byte-identical
  to the design-time trace; `brag list --type failed` returns the row;
  `brag learn --type shipped -t x` exits **1** with `unknown flag: --type`;
  `brag learn --help`'s flag table is exactly `-d -h -i -p -T -t` (no
  `--type`, no `-k`).

### Build-phase reflection (3 questions, short answers)

- **What was unclear in the spec?** Nothing. Every literal was pre-flighted
  at design and the recorded hashes let build *verify* the transcription
  instead of trusting it — the learn.go/milestone.go baseline hashes matching
  design's recorded values byte-for-byte is a stronger check than a visual
  diff would have been.
- **What was missing that you had to decide yourself?** The exact wording of
  the mutation text for M-B, M-E, and M-G (the spec names the mutation class
  — "add a `--type`/`-k` flag", "change the wording", "drop the note" — but
  not the literal replacement string, since the assertion being probed is
  behavioral, not textual). Picked minimal, obviously-wrong text in each case
  and confirmed the resulting hash differs from design's recorded value
  (expected, since the literal text differs) while the *fired assertion* is
  identical to what design recorded.
- **What would you do differently?** Nothing on this spec. The one place
  build had to exercise judgment (M-B/M-E/M-G mutation text) is exactly the
  place the spec deliberately left underspecified, because pinning the exact
  mutation string would be design prescribing an implementation detail of a
  throwaway probe.

## Verify

Verified on 2026-09-05 from `main` at `d8c69d7` (**PR #197 had already
merged** — build's "open, not merged" starting state was stale by the time
this session opened; branched `verify/spec-085-brag-learn` from the merged
main rather than from the build branch). Corpus re-derived at **401** in the
probe copy (398 live + the 3 failures this session wrote to a copy);
`~/.bragfile/db.sqlite` was never written.

**Verdict: ⚠ PUNCH LIST — applied here, not sent back.** Everything SPEC-085
claims for itself holds. Two omissions were found in its own artifacts and
fixed in this PR; the substantial finding is about **SPEC-086's premise**, not
about anything SPEC-085 shipped.

### What was attacked

Build's checklist was not re-run. Five attacks were chosen for what a passing
build cannot see.

#### 1. Does a `failed` row degrade any existing consumer? — **two surfaces read wrong, and neither is the pair SPEC-086 names**

A copy of the live corpus was given three `brag learn` entries and driven
through every surface that reads `type`. Minimal standalone reproduction (two
entries, one of each kind):

```
$ brag --db /tmp/min.sqlite add   -t "Shipped the worker pool PoC" -p demo -i "cut p50 by 40%" -k shipped
$ brag --db /tmp/min.sqlite learn -t "shared-worker pool did not cut cold starts" -p demo -i "cost two days and produced nothing reusable"
```

| Surface | Renders `type` in markdown? | Reads |
|---|---|---|
| `memory`, `brag://memory/recent`, `brag_memory` | **yes** — `[demo/failed]` | correct |
| `export`, `show` | **yes** (per-entry table) | correct |
| `list`, `search`, `spark`, `stats`, `coverage`, `tags` | n/a (type-blind or neutral) | correct |
| `review` | no, but heading is `## Entries` + reflection questions | correct |
| `brag_list` / `brag_search` (MCP) | **yes** — full entry JSON carries `"type":"failed"` | correct |
| `impact` | no | **as SPEC-086 describes** |
| `wrapped` | no (in *Impact moments*) | **as SPEC-086 describes** |
| `summary` | no (in *Highlights*) | **reads wrong — not in SPEC-086's pair** |
| `story` markdown | no | **reads wrong — not in SPEC-086's pair** |

**`impact` and `wrapped`: the damage is exactly what SPEC-086 describes and no
worse.** Confirmed by reproduction, not inherited:

```
$ brag --db /tmp/min.sqlite impact --quarter
Entries: 2/2 with impact
## Impact
### demo
- 1: Shipped the worker pool PoC
  cut p50 by 40%
- 2: shared-worker pool did not cut cold starts
  cost two days and produced nothing reusable
```

Silently included, unmarked, unrepresentable as a failure — SPEC-086 §2
verbatim. `wrapped`'s *Impact moments* is the same document. `wrapped`'s
`## Rhythm` → *Top types* **does** render `failed` honestly, so the leak is
confined to the section SPEC-086 already owns. The JSON is the same 4-key
`impactEntry`. **No worse than described.**

**The finding: SPEC-086's Fork B classifies `story` and `summary` as
"neutral", and measured, they are not.** Fork B reads:

> *"`wrapped`/`impact` are celebratory; `summary`/`story`/`export`/`coverage`
> are neutral, and a neutral surface arguably needs no section at all."*

`export` and `coverage` are genuinely neutral — verified above. The other two
are not:

**(a) `brag story --audience exec` is the most promotional surface in the
tool, and it is a prompt, not a display.** Its markdown output is fed to an
LLM together with a directive it also prints:

```
$ brag --db /tmp/min.sqlite story --quarter --audience exec
## Threads
### demo
- ★ 1: Shipped the worker pool PoC
  cut p50 by 40%
- ★ 2: shared-worker pool did not cut cold starts
  cost two days and produced nothing reusable

## Framing directive
# Framing directive — audience: exec (promote, impact-forward)
- Lead with business impact. Every beat below carries a ★ impact
  statement; build the narrative from those outcomes, not the activity.
- Terse and promotional. One or two sentences per outcome. No process,
  no messy middle — the highest-impact thread leads.
- Quantify wherever the impact beats give you a metric.
```

Expected: a failure is distinguishable from a win. Actual: both are `★` beats,
and the directive instructs the model to build a **promotional** narrative
from every one of them and to **quantify** from `"cost two days and produced
nothing reusable"`. This is a strictly worse outcome than `impact`'s, because
`impact` shows a human a mislabelled row while `story` instructs a model to
launder it. Two further notes: `--audience me`'s directive is genuinely candid
(*"Include the messy middle: struggles, false starts… are the point"*), so
**"is `story` celebratory?" is not a property of the command — it is a
property of the profile**, and Fork B's per-surface framing cannot express
that; and `story --format json` **does** carry `"type":"failed"` on each beat,
so the markdown path is the only lossy one.

**(b) `brag summary`'s section is literally named `## Highlights`.**

```
$ brag --db /tmp/min.sqlite summary --range week
**By type**
- failed: 1
- shipped: 1
## Highlights
### demo
- 1: Shipped the worker pool PoC
- 2: shared-worker pool did not cut cold starts
```

The *By type* block is honest; *Highlights* lists a two-day dead end as a
highlight with no type rendered, in either format (`summary`'s JSON highlight
is a 2-key `{id,title}` — narrower than `impact`'s 4-key).

**Consequence for SPEC-086, not for SPEC-085:** Fork B must be answered
against the measured framing of each surface, not an assumed one. Nothing here
asks SPEC-085 to change; it is filed so SPEC-086's design does not inherit a
premise verify has already falsified — the same service SPEC-085's design did
for framing.

**One sharpening of SPEC-085's own "interim risk", which is stated as bounded
because *"the user controls when the first one is written"*.** True, but
`BRAG.md`'s new section (this PR) instructs the agent that *"The `impact`
field is still worth filling in"* and shows `-i` in its example. So the
leaking shape is now the **documented default path**, not merely a possibility.
The risk is still acceptable and still bounded; the reason it is bounded is
the empty corpus, not user restraint.

#### 2. `--type` negation: confirmed inexpressible, and it fails **silently**

`internal/storage/store.go:388` is exact-match inclusion on a single string:

```go
if f.Type != "" {
    conds = append(conds, "e.type = ?")
    args = append(args, f.Type)
}
```

Every obvious attempt at *"everything except failures"* returns **exit 0 and
zero rows** — no error, no diagnostic:

```
$ brag --db /tmp/min.sqlite list                      # 2 rows (baseline)
$ brag --db /tmp/min.sqlite list --type '!failed'     # 0 rows, exit 0
$ brag --db /tmp/min.sqlite list --type '-failed'     # 0 rows, exit 0
$ brag --db /tmp/min.sqlite list --type 'shipped,failed'  # 0 rows, exit 0
$ brag --db /tmp/min.sqlite list --type shipped --type milestone  # 0 rows, exit 0 (cobra last-wins)
$ brag --db /tmp/min.sqlite list --type ''            # exit 1: "--type must not be empty"
```

So the claim holds, and the failure mode is worse than "unsupported": a user
filtering failures out of a review gets an **empty document rather than a
diagnostic**. Only the empty string is rejected.

**Nothing in this PR implies otherwise.** `git diff 2913c80..d8c69d7` over the
doc set returns no added line containing `except`/`exclude`/`minus`/`negat`
in a retrieval claim; `list --help` says `filter to entries with this type
(exact match)`; every doc teaches the positive form `brag list --type failed`
only. Worth carrying to SPEC-086, whose Fork A needs exactly this negation.

#### 3. The Z7 double-claim: **first half true, second half false**

The spec's Outputs table warns DEC-049 *"**Must carry `  type: decision`** …
or `Z7` fails **and the inventory silently under-reports**."* Both halves
tested by mutation (backup to the session scratchpad, `shasum -a 256` before
and after, restore from the backup — never `git checkout`):

```
$ shasum -a 256 decisions/DEC-049-*.md        # be0a6b71…
$ perl -i -pe 's/^  type: decision  .*\n$//' decisions/DEC-049-*.md
$ shasum -a 256 decisions/DEC-049-*.md        # 56f78528…   (mutant landed)
$ ./scripts/inventory.sh | grep 'Decision records'
| Decision records | 47 | ...
$ ./scripts/test-docs.sh
FAIL: X3: inventory block is stale — run `just inventory` and paste between the markers:
FAIL: Y3: inventory.sh row value(s) wrong: decision-records!=48
FAIL: Z7: the inventory covers 48 of 49 decisions/DEC-*.md files (47 decision + 1 reservation)…
FAILED: 3 assertion(s) failed.
$ cp <backup> decisions/DEC-049-*.md
$ shasum -a 256 decisions/DEC-049-*.md        # be0a6b71…  ✔ restored
```

- **Half (a) — Z7 fires: TRUE**, with a message that names the exact cause.
- **Half (b) — "silently under-reports": FALSE.** The row does move (48 → 47),
  but **three** assertions fire, not zero. `Y3` pins the literal and `X3` diffs
  the whole block, so a lost DEC is loud on three independent channels.

The warning is half-right and the wrong half is the one that would matter:
it tells a future author the guard is weak where it is in fact triple-covered.
Corrected in place under *Outputs* below. (The "silent" reading is only
reachable if an author *also* regenerated the block and re-pinned `Y3` to the
under-reported value — and `Z7` exists precisely to be the backstop there,
which is what its own failure message says.)

#### 4. The two routed items, re-checked rather than inherited

**`docs/tutorial.md` — correctly routed; incomplete, not wrong.**

- *No assertion depends on it.* Group `AB` targets exactly `BRAG.md`,
  `README.md`, `docs/api-contract.md` — `grep -c tutorial` over the Group AB
  block returns **0**, and the harness is `ALL OK` without it.
- *Not wrong.* The two places the tutorial talks about `type` are
  `tutorial.md:82` (*"`--type` is free-form text — pick whatever feels useful
  … No enforced enum"*) and `tutorial.md:135`. Both describe **`brag add`**,
  and LD2 keeps `brag add --type` free-form and unvalidated — so both remain
  literally true. The tutorial's scope line already disclaims completeness
  (*"See `docs/api-contract.md` for the full command surface"*), and
  `api-contract.md` **was** updated.
- *One thing to carry forward, not a defect:* `tutorial.md:82`'s *"pick
  whatever feels useful"* is now the only user-facing sentence that invites
  the fragmentation `AB2-neg` guards against elsewhere. It is not false, and
  it is not this spec's to fix; it belongs with the tutorial section
  SPEC-086 will write.
- `W3` (the *"shipped as of v0.6.1"* pin) keys on the latest **dated**
  CHANGELOG section, and SPEC-085's entry went to `## [Unreleased]` — so W3 is
  unaffected until the release cut, as it should be.

**The DEC-004 sweep — count re-derived, routing confirmed.**

```
$ grep -rn 'DEC-004' --include='*.go' . | wc -l
11
```

**11 confirmed**, across 9 files, and `internal/story/thread.go:143` does carry
the `DEC-004/DEC-015` idiom the spec names as the model. **This PR adds no new
stale citation**: the only added `DEC-004` line in `2913c80..d8c69d7` is inside
DEC-049's own prose and reads *"DEC-015 — tags are normalized; supersedes
DEC-004"*, which is the correct form. No assertion depends on the Go-source
citations. **One trap for whoever takes the chore:** `X6` requires
`docs/engineering-practices.md` to keep citing the literal `DEC-004` (it is one
of six ids in X6's list), so a repo-wide "fix the stale citations" pass that
rewrites that page will turn `X6` red for the right reason and the wrong cause.

#### 5. Group `AB`'s unprobed branches — all have teeth

The eight design/build mutations proved `AB1b`, `AB2a`, `AB2-neg`, `AB3c` and
`AB4`. `AB2-neg` is a 3-file × 4-needle nested loop — **12 branches, of which
M-F exercised exactly one** (`BRAG.md` × `--type failure`). Four unprobed
branches were driven, full hash protocol each:

| Probe | Mutation | Hash | Fired |
|---|---|---|---|
| `V-M1` | `README.md` teaches `--type abandoned` | `42d0820e` → `d0d05205` | `AB2-neg: … [README.md teaches '--type abandoned']` |
| `V-M2` | drop the `AB3a` sentence from `docs/api-contract.md` | `1fc45113` → `2f81f69c` | `AB3a: … missing literal` |
| `V-M3` | drop the `AB3b` sentence from `docs/api-contract.md` | `1fc45113` → `1fa8871d` | `AB3b: … missing literal` |
| `V-M4` | remove `README.md`'s `brag list --type failed` line | `42d0820e` → `3f2be8c3` | `AB2b: … missing literal` |

All four restored to their pre-mutation hash. **Nothing found** — the untested
needle (`abandoned`), the untested file (`README.md`) and the two untested
`AB3` positives are all live. Incidental cross-check: `V-M2`/`V-M3`'s
pre-mutation hash `1fc45113` is byte-identical to the baseline design recorded
for `docs/api-contract.md` at `81e639d`, independently confirming the
transcription introduced no drift.

### What did not hold — two omissions in SPEC-085's own artifacts, fixed here

**F1 — the `Y3`/`Y4` re-pin notes the spec asked for were never written.**
Not a build error: the spec **contradicts itself**. Its `## Failing Tests`
says, under *Changed — `Y4`*:

> *"Add a re-pin note in the style of the existing SPEC-082/SPEC-083 notes…"*

while its `## Notes for the Implementer` §9 embedded literal diff changes
**only the two value lines**. Under the literal-artifact contract build
transcribes the literal verbatim — which it did, correctly, and reported
"Deviations: none", also correctly. The result:

```
$ git diff 2913c80..d8c69d7 -- scripts/test-docs.sh | grep -E '^[-+].*RE-PINNED'
(no output)
```

`Y3`'s pin reads `48` while its comment history stops at *"RE-PINNED 46 -> 47
at SPEC-084"*; `Y4`'s reads `20/7` while its history stops at SPEC-083's
`19/6`. That is the exact drift those notes exist to prevent — `Y3`'s own
comment records that three consecutive specs' value-greps found it, and the
next author re-pinning `48 → 49` would read a baseline of 47. **Both notes
added in this PR.**

This is also a gap in the §12(b) recipe worth naming: the design pre-flight's
literal round-trip proved *the literal survives extraction*, which is a
different claim from *the literal implements every instruction the spec gives
elsewhere*. Nothing cross-checks `## Failing Tests` against the embedded diff.

**F2 — the mutation-class hash finding: recorded, deliberately NOT codified.**
Build reproduced 5 of 8 design hashes (M-A/M-C/M-D/M-F/M-H) and not M-B/M-E/
M-G, with the fired-assertion message matching verbatim in all three. The split
is predictable, not random: the five that matched are point mutations with one
obvious form; the three that did not required build to **invent** text, because
the spec named the class.

**The protocol was not violated.** AGENTS.md §12 clause (1) requires only that
the hash **move** — `pre != post` — which is what proves the mutant landed and
what `git diff --quiet` gets wrong on an untracked file. It has never required
`post` to equal a previously recorded value. The mismatch is against a
stronger expectation nobody wrote down.

**The call: no note in the spec template, and no new AGENTS.md §12 clause —
filed as a `guidance/questions.yaml` entry instead.** Reasoning, against this
repo's own codification meta-rule (N=2 paired-opposing, N=3 same-outcome):
this is **N=1** — three instances inside one spec, produced in one sitting by
one design habit, with no opposing case, and the cost was zero because build
noticed and explained it. A template note binds every future spec on that
evidence; that is the objection SPEC-079 raised against the corrections
convention, and the reason that one is still a question. The cheap fix needs
no rule at all: a mutation table can mark which probes are literal-specified
(a hash mismatch is real signal) and which are class-specified (the fired
assertion is the contract). Filed as **`mutation-probe-class-vs-literal`**
with both promotion triggers and the closing trigger, so a second independent
case can settle it either way — the same shape SPEC-085 used for the two
memory questions, and the reason `AB4` asserts *examined* rather than merely
*open*.

Consequence, regenerated and diffed rather than predicted: questions **20 → 21**
and open **7 → 8**; `Y4` re-pinned to `21`/`8`; the inventory block
regenerated. `diff <(./scripts/inventory.sh) <(block)` is empty. **No other
row moved** — the doc-assertion count held at 198 because the `Y3`/`Y4` notes
add comments, not assertion ids. Proved live rather than assumed:

```
$ perl -i -0pe 's/(mutation-probe-class-vs-literal.*?\n    )status: open/${1}status: answered/s' guidance/questions.yaml
$ ./scripts/test-docs.sh
FAIL: X3: inventory block is stale …
FAIL: Y4: inventory.sh row value(s) wrong: questions-open!=8
$ cp <backup> guidance/questions.yaml     # hash 6860b32c restored ✔
```

### What held — re-derived independently, not inherited

- **All five gates green** after the punch-list edits: `go test ./...` **1070
  tests, 14 packages `ok` + 1 with no test files = 15**; `gofmt -l .` empty;
  `go vet ./...` clean; `just lint` **0 issues**; `just test-docs` **ALL OK at
  199 `OK:` lines / 198 distinct ids**, `grep -c 'OK:   AB'` = **10**.
- **Live corpus re-derived** (read-only copy): **398** entries, **18** distinct
  non-empty `type` values, **0** rows of `type='failed'` — design's figures
  hold unchanged.
- **AC-2 / AC-3 / AC-4** re-run end-to-end on a temp DB, matching the recorded
  traces byte-for-byte, including `brag memory`'s
  `- 1 2026-09-05 [bragfile/failed] …` line.
- **LD6 verified live over the wire**, not by reading source: a real
  `tools/list` on `brag mcp serve` returns `brag_add.properties.type` carrying
  *"One value is reserved: \"failed\" … Use it verbatim; do not invent a
  spelling."*, and the tool count is unchanged at five.
- **AC-11 holds:** no golden file is touched by `2913c80..d8c69d7`, and the
  MCP resource surface still renders the memory slice byte-identically.
- `NEXT-SESSION-PROMPT.md` was left modified and uncommitted throughout, as
  instructed — not committed, not reverted.

### Punch list

| # | Item | Status |
|---|---|---|
| F1 | `Y3`/`Y4` re-pin notes missing | **fixed in verify's PR.** Diagnosis **corrected at ship**: only `Y4` is a spec self-contradiction (its `## Failing Tests` asks for the note, its embedded literal omits it). `Y3`'s entry never asks for a note at all — a plain omission of an uncodified convention, not a contradiction. Both comments now name their own cause. |
| F2 | mutation-class hash reproducibility | **filed** as `mutation-probe-class-vs-literal`; not codified, with reasons |
| F3 | `Z7`'s *"silently under-reports"* is false | **corrected in place** under *Outputs* |
| F4 | SPEC-086 Fork B calls `story`/`summary` neutral; measured, they are not | **routed to SPEC-086** — no SPEC-085 change |
| F5 | the leaking shape is `BRAG.md`'s documented default, not merely possible | **routed to SPEC-086** — sharpens the interim risk, does not invalidate it |
| F6 | `X6` pins `DEC-004` in `docs/engineering-practices.md` | **routed** to the DEC-004 sweep chore as a named trap |

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — **Make `## Failing Tests` and the embedded literal say the same thing,
   because nothing cross-checks them.** This spec's one real defect was `Y4`:
   its `## Failing Tests` entry asked in prose for *"a re-pin note in the style
   of the existing SPEC-082/SPEC-083 notes"*, while its own embedded diff under
   `## Notes for the Implementer` §9 changed only the two value lines. Under
   the literal-artifact contract the literal wins, so the instruction was
   unreachable — build transcribed faithfully and reported *"Deviations:
   none"*, and both of those were correct.
   — **Corrected at ship: verify diagnosed `Y3` as the same defect, and it is
   not.** Read line by line, `Y3`'s entry asks only to move the value and the
   `!=` string; **nothing in this spec asks for a `Y3` note at all.** Same
   missing sentence, two different causes — `Y4` is a self-contradiction a
   design review could catch by reading the spec against itself, `Y3` is a
   plain omission no review could catch, because **the re-pin-note convention
   is not a requirement anywhere.** It exists only in the comments it has
   already produced, so it survives exactly as long as the next author happens
   to read one. Both comments in `scripts/test-docs.sh` now name their own
   cause, so the next re-pin inherits the right lesson rather than the
   averaged one. The conclusion verify drew is unchanged and still exonerates
   build; only the diagnosis per pin moved.

2. **Does any template, constraint, or decision need updating?**
   — **No new AGENTS.md §9 clause and no template note, on this repo's own
   codification meta-rule** (N=2 paired-opposing, N=3 same-outcome). The
   prose-vs-literal contradiction is **N=1**, and verify already declined to
   codify the adjacent mutation-hash finding on the same grounds, filing
   `mutation-probe-class-vs-literal` instead. Two N=1 findings do not add up
   to a rule just because they arrived in one cycle.
   — **The §9 half-(b) rule that WAS codified earned its keep here, and this
   is the first spec in five where it did.** `Y3` has now been re-pinned by
   hand at SPEC-081, 082, 083, 084 and 085. At the first four the value-grep
   found it *late* — STAGE-022's close says so in as many words. At SPEC-085 it
   was found **at design**: the §12(b) pre-flight ran the harness before
   locking the spec and watched `Y3` fail with `decision-records!=47`, so the
   re-pin was a planned Output rather than a build-time surprise. Grepping for
   the value and not the idea is now a demonstrated save, not a maxim.
   — **What did not fire is the routing.** STAGE-022's close promoted *"`Y3`
   and `X3` pin the same numbers by different means, and one of them should
   derive rather than cache"* to a stage-level lesson, then **held** it with
   the trigger *"routed to the next spec that opens that file."* SPEC-085 is
   exactly that spec — it opened `scripts/test-docs.sh` to insert 80 lines of
   Group `AB` — and it hand-re-pinned `Y3` anyway, making it five consecutive
   specs. That is the same shape STAGE-022 recorded one line earlier about
   `archive-spec`: *"Routing it as a candidate after the first hit did not
   prevent the second; the guard did."* **Recorded, not codified — N=2
   same-outcome, one short.** Re-routed onto STAGE-023 with a concrete owner
   (SPEC-086) instead of an anonymous "next spec", since that is the half that
   demonstrably failed.

3. **Is there a follow-up spec I should write now before I forget?**
   — **No new spec.** SPEC-086 already exists at `cycle: frame`, blocked on
   this one, and is where the remaining work is owed. Verify routed two
   *measured* corrections to its premise onto the stage page rather than into
   its file — `story --audience exec` and `summary`'s `## Highlights` are not
   the "neutral" surfaces Fork B assumes — and this session adds the `Y3`/`X3`
   de-duplication to the same place. All three are on STAGE-023, so SPEC-086's
   framing inherits them without SPEC-085 having edited a spec it does not own.
   — **One thing SPEC-086 must not inherit as settled:** `--type` negation is
   inexpressible today and fails *silently*. Re-confirmed at ship on the same
   two-row corpus: `--type '!failed'`, `--type '-failed'` and
   `--type 'shipped,failed'` each return **exit 0 with zero rows**; only
   `--type ''` errors. Fork A needs exactly that negation, so it is a
   dependency, not a footnote.

4. **What can a user do now that they couldn't before?**
   — A user can **record work that did not work, and get it back labelled as a
   failure** — in one command, without contorting it into a win. Before:
   **399 entries in the live corpus and 0 of type `failed`** (re-derived at
   ship from a read-only copy; 397 at framing, 398 at design and verify), an
   absence that was structural rather than a discipline problem — no verb, no
   field value, no documented convention. After: `brag learn -t "…" -i "…"`
   writes the reserved value `failed` (DEC-049) with **zero bytes on stderr**,
   because every milestone line is a congratulation; `brag list --type failed`
   reads it back; and `brag memory` renders
   `- 1 2026-09-05 [demo/failed] shared-worker pool did not cut cold starts —
   cost two days and produced nothing reusable`, so an agent reading the corpus
   cold — including over `brag://memory/recent`, which loads with no tool call
   at all — sees a failure marked as one for free. The corpus is now *capable*
   of honesty; what the celebratory digests do with a failure is SPEC-086.
   Evidence ref: `pr:199`.

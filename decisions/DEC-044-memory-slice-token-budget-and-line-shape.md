---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-044
  type: decision
  confidence: 0.80                   # honest: the expression (tokens not rows),
                                     # the enforcement (skip-and-continue), the
                                     # budget-against-markdown-in-both-formats
                                     # rule, and the "never stamped, never a
                                     # tag" scoping are strong and mechanically
                                     # guarded; the soft spots are the 4-bytes-
                                     # per-token divisor and the 2000 default,
                                     # both single numbers with real evidence
                                     # behind them but no measurement — see
                                     # Validation.
  audience:
    - developer
    - agent

agent:
  id: claude-opus-5
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-006
repo:
  id: bragfile

created_at: 2026-08-08
supersedes: null
superseded_by: null

tags:
  - agent-native
  - memory
  - token-budget
  - estimation
  - rendering
  - honesty
---

# DEC-044: The memory slice's token budget — expression, estimation, enforcement, and the line shape that is its unit

## Decision

`brag memory` is bounded by an **estimated token budget**, not a row count. Seven
coupled sub-decisions are locked. They are one decision because the budget is
meaningless without the line shape it is spent on: **the rendering IS the unit of
cost**, so it is locked here rather than left to the surface.

### 1. Expression — a token count, `--budget N`, default **2000**

`--budget` takes a token count. `0` and negatives are a **`UserError`**, not
"unlimited": unlike `--limit` (where `0 = no LIMIT` is an established convention in
this repo), a *budget* has no natural unbounded reading, and `--budget 0` silently
emitting the whole pool into a model's context is exactly the footgun this command
exists to remove. An unbounded dump is spelled `brag list`.

**Why 2000.** It is sized against a real auto-load cost, measured in this repo:

| what a session actually reads today | bytes | ≈ tokens |
|---|---|---|
| `AGENTS.md` | 49,842 | 12,461 |
| `PROJ-006/brief.md` | 15,534 | 3,884 |
| `STAGE-019-corpus-as-agent-memory.md` | 15,278 | 3,820 |
| **the session-opening ritual, total** | **80,654** | **≈20,163** |

The ritual STAGE-019 exists to replace costs ~20k tokens. A 2000-token slice is
**~10% of that**, and ~1% of a 200k context window. At the line shape locked in (5)
a real corpus averages ~18 tokens per entry, so 2000 buys **≈110 entries** — on the
order of a quarter's capture for a daily logger. That is the target: cheap enough
that auto-loading it every session is never a decision, large enough that it is
worth reading. A number chosen from "what fits" rather than "what it displaces"
would have no defense at all.

### 2. Estimation — a documented chars-per-token heuristic: `ceil(len(bytes) / 4)`

`memory.EstimateTokens(s string) int` returns `ceil(len(s)/4)` over **UTF-8 bytes**,
computed **per rendered unit** (one entry line = one unit), and summed. Bytes, not
runes, deliberately: modern BPE tokenizers operate on bytes and non-ASCII text costs
*more* tokens per character, so a byte count degrades in the **safe** direction (it
over-estimates the cost of multi-byte text rather than under-estimating it). `~4
characters per token` is the widely published English-prose approximation. It is
dependency-free, deterministic, and stdlib-free (integer arithmetic only).

> #### ⚠ Honesty clause — this is NOT DEC-027's `tokens` provenance, and must never become it
>
> [DEC-027](./DEC-027-seed-cost-session-token-reserved-tags.md) is unambiguous:
> bragfile **never estimates tokens**, and its Option B ("fabricate/estimate token
> counts in bragfile") is rejected because *"a fabricated number is worse than no
> number — it pollutes the very dataset PROJ-005 will trust."* That rule is intact
> and this decision does not weaken it. The two things are different in kind:
>
> | | DEC-027's reserved token-count tag | DEC-044's estimate |
> |---|---|---|
> | **what it measures** | what a model actually spent producing the work | how much context this retrieval will cost to read |
> | **where it comes from** | the caller, who has the usage accounting | bragfile, from bytes it is about to print |
> | **where it goes** | stamped on an entry, stored, aggregated, reconciled later | printed in one envelope, then discarded |
> | **lifetime** | forever, in the corpus | the duration of one command |
> | **may be wrong?** | no — a wrong value corrupts the dataset | yes — it sizes a retrieval; being 15% off costs a few entries |
>
> **The scoping rules, binding:** the estimate is **never written to the database**,
> **never stamped on an entry**, **never passed to any capture path**, and **never
> enters the reserved tag namespace** in any form. `internal/memory` performs no
> writes. Enforced mechanically, not by convention, by
> `TestPackageEmitsNoReservedTagNamespace`, which walks the package's non-test
> sources and fails on any *quoted* reserved-namespace prefix literal. Naming
> follows: the exported symbol is `EstimateTokens` and the envelope key is
> `estimated_tokens` — both carry "estimate" in the name so no reader can mistake
> the number for a measurement, and neither is ever spelled as the reserved tag.
>
> If DEC-027 is ever read as forbidding this, read them together: DEC-027 governs
> **caller-reported provenance about work**; DEC-044 governs **bragfile sizing its
> own output**. Nothing bragfile prints about the size of what it is printing can
> pollute a dataset it never writes to.

### 3. Enforcement — greedy fill in rank order, **skip the oversize entry and continue**

Walk the ranked candidates in order; include an entry if its line cost fits the
remaining budget, otherwise **skip it and keep going**. Not "stop at the first
overflow."

The fixture makes the difference concrete (budget 40, no query/project; per-entry
costs `8→27, 7→14, 6→15, 5→10, 4→27, 3→15, 2→22, 1→13`):

| policy | included | estimated | budget used |
|---|---|---|---|
| stop at first overflow | 1 (`8`) | 27 | 68% |
| **skip and continue (chosen)** | **2 (`8`, `5`)** | **37** | **93%** |

Skip-and-continue packs better and, more importantly, removes a failure mode:
without it, one long entry landing at rank 1 truncates the entire slice behind it.
The `capture.Validate` field caps bound a single line at **626 bytes / 157 tokens**
worst case, so catastrophic starvation is impossible — but a 27-token entry ahead of
a queue of 10-token entries is ordinary, and stopping there wastes a third of the
budget on nothing.

**Accepted consequence, stated plainly:** a *lower*-ranked short entry can be
included while a *higher*-ranked long one is skipped. In the table above, rank 4
(`5`) is included while ranks 2 and 3 (`7`, `6`) are not. That is the intended
trade — better packing over strict rank monotonicity — and it is why the envelope
**reports `Included` and `Skipped`** rather than leaving the caller to infer them.
Output order remains rank order, so the slice still reads best-first.

`Included + Skipped == Entries` (the deduped candidate-pool size) is an invariant,
pinned by a test.

### 4. The budget is spent on the ENTRY LINES; the envelope is named overhead

`--budget` bounds the rendered `## Slice` body. The DEC-014 header block and the
`## Budget` footer are **fixed overhead outside the budget** (~35 tokens together,
independent of corpus size).

This is the non-circular resolution of a real trap: the footer reports
`Estimated:`/`Included:`/`Skipped:`, so a footer inside its own accounting would
have to converge on a fixed point (the digit count of the number changes the byte
count that produces the number). Every value in the footer is therefore known
**before** the footer renders. In exchange the boundary is stated everywhere it
matters — the DEC, the `--budget` flag help, the cobra `Long`, and
`docs/api-contract.md` — so total cost is `estimated + a small constant`, never a
hidden term. A caller wanting a hard ceiling subtracts ~35.

### 5. Per-entry rendering — one locked line shape, no `--detail` knob

```
- <id> <YYYY-MM-DD> [<project>/<type>] <title> — <impact>
```

- `<id>` bare, first, so the follow-up (`brag show 42`, `brag_list`) is trivial —
  the line is a *pointer* to the full record, which is what lets it stay one line.
- `<YYYY-MM-DD>` is `created_at` in **UTC** (the date part only). UTC because every
  other renderer's `Generated:` line is UTC and the pure package owns no timezone
  policy; the date part only because a timestamp costs ~5 tokens per entry to
  answer a question ("when, roughly?") that a date already answers.
- `[<project>/<type>]` always renders, with **`-`** for an absent side —
  `[bragfile/shipped]`, `[bragfile/-]`, `[-/shipped]`, `[-/-]`. A uniform shape is
  trivially parseable and trivially golden-able. `-` rather than DEC-013/`aggregate`'s
  `(no project)` sentinel: this line is token-budgeted, and `(no project)` costs ~4
  tokens per line to say nothing. `-` is already DEC-014's markdown empty
  convention, so it is a reuse, not an invention.
- `— <impact>` renders **only when `Impact` is non-empty** — no trailing dash
  artifact. Impact is included because it is the outcome line: the single
  highest-value-per-token field in the corpus, and the reason a slice beats a title
  list.
- `Description` and `Tags` are **excluded**. Description is the long field (cap
  100,000 bytes) and would make per-entry cost unpredictable; tags are mostly
  machine namespaces at this altitude. The id is on the line; detail is one call
  away.
- **No truncation, ever.** An entry is included whole or skipped whole. Truncating
  would make the estimate and the render agree while the *content* lied — a
  half-sentence of impact is worse than a reported skip.
- **No `--detail` / `--verbose` knob.** Two line shapes means two token models, two
  golden sets, and a second budget calibration. YAGNI, deliberately.

### 6. The budget is always computed against the MARKDOWN rendering — in both formats

`--format json` selects the *rendering*, never the *selection*. The budget, the
per-item `tokens`, and `estimated_tokens` are all computed from the markdown line in
both formats, so the slice's **ID set and order are identical** across
`--format markdown` and `--format json` for the same options. Pinned by a test.

Markdown is the budgeted rendering because markdown is what a model reads (it is the
body STAGE-019's success criteria are written against, and the rendering SPEC-074's
resources will serve). JSON is a data pipe for tooling; letting a format flag
silently change *which entries you get* would break the "one slice, two renderings"
property that makes the goldens a contract.

### 7. The envelope reports the accounting

Markdown gains a `## Budget` section (after `## Slice`):

```
## Budget

- Budget: 2000 tokens
- Estimated: 143 tokens
- Included: 8
- Skipped: 0
```

JSON carries `budget`, `estimated_tokens`, `included`, `skipped` as flat top-level
keys (DEC-014 choice 2), plus a per-item `tokens`. Per DEC-014 choice 4, on an
**empty candidate pool** the document ends after `Entries: 0` and both body sections
are omitted; JSON still emits every key. A **non-empty** pool that includes zero
entries (budget too small for any line) still renders both sections — `## Slice`
empty, `## Budget` reporting `Included: 0` — because that is precisely the case a
caller needs to see in order to raise the budget.

`internal/memory` owns the line renderer (`RenderLine`) and the estimate, so
`estimated_tokens == Σ EstimateTokens(rendered line)` holds **by construction**, not
by two implementations agreeing. Pinned anyway.

## Context

`Limit` is a row count, and a row count is the wrong unit for a slice whose whole
purpose is to be auto-loaded into a context window. Twenty entries might be 200
tokens or 3,000 depending on whether they carry impact lines — so a caller sizing
against rows is sizing against nothing. STAGE-019 named the budget as the thing that
makes the slice "cheap enough to auto-load every session," which only means something
if the unit is the thing being spent.

Four questions had to be answered together, and the stage's Design Notes enumerated
them: how the budget is *expressed*, how tokens are *estimated* without shipping a
tokenizer, how the budget is *enforced* when an entry does not fit, and what a
rendered entry *is* — because that last one determines the first three. A budget
locked without its line shape is a number attached to nothing.

The estimation question carried a live hazard the stage flagged explicitly: DEC-027
states bragfile never estimates tokens. Getting the scoping wrong here would leave
two contradictory decisions in `/decisions/` — which is why sub-decision 2 spends
more words on the boundary than on the arithmetic, and why the boundary is enforced
by a test rather than by a paragraph.

## Alternatives Considered

- **Option A: express the bound as rows (`--limit N`), as `list`/`search` do.**
  - Why rejected: it is the status quo the stage set out to replace. Rows do not
    predict cost — the fixture's cheapest entry is 10 tokens and its most expensive
    is 27, a 2.7× spread, and the caps allow 157. A caller who wants "about this
    much context" cannot express it, and a caller auto-loading every session has no
    way to bound the damage. Kept available in spirit: `brag list --limit N` still
    exists for anyone who genuinely wants rows.

- **Option B: ship or vendor a real BPE tokenizer.**
  - Why rejected: a tokenizer is model-specific (a count exact for one model is
    wrong for the next), it is a substantial new top-level dependency
    (`no-new-top-level-deps-without-decision`), several carry large embedded
    vocabulary files, and some paths pull CGO (`no-cgo`). It buys accuracy in a
    number that only needs to be *approximately* right to size a retrieval — and it
    would blur exactly the line the honesty clause draws, since a "real" count
    invites stamping it. Explicitly out of scope by stage decree.

- **Option C: characters or bytes as the budget unit (`--budget 8000` chars).**
  - Why rejected — and this is the honest near-miss: it removes the estimate
    entirely, and with it every objection to a heuristic divisor. It loses on
    **calibration against the thing being spent**. The caller's real constraint is a
    context window, which is denominated in tokens; making them divide by 4 in their
    head to translate is the same arithmetic, relocated to the party with less
    information. Recorded as the fallback if the divisor proves indefensible: the
    mechanism is identical, only the reported unit changes.

- **Option D: stop at the first overflow (monotone fill).**
  - What it is: fill greedily and terminate at the first entry that does not fit;
    the slice is then a strict rank prefix.
  - Why rejected: it is genuinely simpler and preserves strict rank monotonicity,
    which is a real property to give up. But it wastes budget in the ordinary case
    (68% vs 93% on the fixture) and, worse, makes the slice's size hostage to one
    long entry's position. The counts in the envelope recover the explainability
    that monotonicity would have provided — a caller sees `Skipped: 6` and knows the
    slice is not a prefix.

- **Option E: truncate an oversize entry to fit.**
  - Why rejected: it makes the accounting honest and the content dishonest. A brag's
    impact line truncated mid-clause ("cut p99 login latency from 600ms…") reads as
    a complete claim and is not one, and an agent consuming it has no signal that
    anything was removed. Skip-and-continue loses the entry but never misrepresents
    one, and the skip is counted.

- **Option F: let the budget cover the whole rendered document, envelope included.**
  - Why rejected: circular. The `## Budget` footer reports numbers whose digit count
    changes the byte count that produces them, so the accounting would need a fixed
    point (or a reserved allowance, which is the same fudge with a nicer name).
    Sub-decision 4's boundary is stated in four places instead, which is a smaller
    cost than a self-referential estimator.

- **Option G: budget against whichever format is being rendered.**
  - Why rejected: `--format json` would then silently return **fewer entries** than
    `--format markdown` for the same budget, because the JSON envelope is several
    times bulkier per entry. A format flag that changes the *answer* rather than its
    presentation is a trap, and it would mean two golden sets that cannot be diffed
    against each other. Sub-decision 6 makes format a pure rendering choice.

- **Option H: `--budget 0` means unlimited (matching `--limit 0`).**
  - Why rejected: the symmetry is superficial and the failure is asymmetric. A
    surprising `--limit 0` prints extra rows to a terminal; a surprising
    `--budget 0` pushes an unbounded corpus into a context window through an
    auto-loading resource. Erroring costs one line of validation and removes the
    class.

- **Option I: a `--detail` / `--verbose` line shape that adds `description`.**
  - Why rejected: YAGNI, and it multiplies the surface — two token models, two
    golden sets, two budget calibrations, and a `--budget` default that is right for
    at most one of them. The id on every line makes the full record one call away,
    which is the right shape for a *slice*.

## Consequences

- **Positive:** a caller can say what they actually mean ("about 2000 tokens of my
  history") and get a slice sized to it, with `Included` / `Skipped` /
  `Estimated` reported so the next invocation is a calibration rather than a guess.
- **Positive:** the estimate and the render cannot disagree — one package owns both,
  and `estimated_tokens == Σ EstimateTokens(rendered line)` is true by construction.
- **Positive:** the DEC-027 boundary is enforced by a test, not by a paragraph
  someone has to read. The two decisions can be read together without contradiction,
  and a future contributor who tries to stamp the estimate hits a red test first.
- **Positive:** locking the line shape here means SPEC-074's MCP resources inherit a
  budgeted, byte-identical body rather than inventing a second rendering — the
  resource *is* the markdown.
- **Negative (accepted):** `ceil(bytes/4)` is a heuristic and will be wrong,
  typically by 10–20% on English prose and further on dense identifiers, URLs, or
  CJK. It over-estimates for multi-byte text (safe) and can under-estimate for text
  full of rare tokens (unsafe, mildly). The mitigation is that the number is
  *reported*, so a caller comparing it against their client's actual accounting can
  correct with `--budget` in one step.
- **Negative (accepted):** the envelope sits outside the budget, so `--budget 2000`
  costs ~2035 tokens. Named in four places; a caller wanting a hard ceiling
  subtracts.
- **Negative (accepted):** skip-and-continue means the slice is not a rank prefix
  (sub-decision 3). Ordinary, intended, and counted.
- **Neutral:** `--budget` gains a `UserError` path that `--limit` does not have. The
  asymmetry is deliberate and documented in the `Long`; it is the DEC-042-style
  "principled asymmetry, written down rather than merely omitted."
- **Neutral:** the fixed line shape freezes a small amount of the CLI contract. It is
  the intended trade — the line shape is the token model, so a stable budget requires
  a stable line. Changing it is a golden diff plus a DEC edit, which is the point.

## Validation

**Right if:**
- `brag memory` with the default budget produces a slice a caller can paste into a
  session without thinking about size, and the reported `Estimated:` lands close
  enough to their client's real accounting that they never need to re-tune.
- `estimated_tokens == Σ EstimateTokens(rendered line)` over the included set —
  pinned by `TestSlice_EstimateMatchesRenderedBody`.
- `Included + Skipped == Entries` — pinned by `TestSlice_CountsPartitionTheCandidatePool`.
- A budget too small for the rank-1 entry still includes a later, cheaper one —
  pinned by `TestSlice_SkipsOversizeAndContinues` and by the `--budget 40` golden
  (`Included: 2`, `Skipped: 6`, `Estimated: 37`).
- The JSON and markdown renderings of the same options select the **same ids in the
  same order** — pinned by `TestMemoryCmd_JSONAndMarkdownSelectTheSameSlice`.
- No non-test file in `internal/memory` contains a quoted reserved-namespace prefix
  literal — pinned by `TestPackageEmitsNoReservedTagNamespace`. If that test ever
  fails, the question is *"is something about to stamp the estimate?"* before it is
  *"should the guard be relaxed?"*.

**Revisit if:**
- **(a) The divisor proves badly wrong in practice** — a user's client reports a
  cost meaningfully different from `Estimated:` on real corpora. First move: adjust
  `CharsPerToken` (it is one exported constant, pinned by goldens). Second move:
  Option C — report the bound in characters and stop estimating at all. Do **not**
  reach for Option B (a real tokenizer); that trade has already been made.
- **(b) The 2000 default is wrong** — slices are routinely truncated at the default
  (raise it), or auto-loading it is noticeably expensive (lower it). The default is
  a single exported constant; the calibration data is already printed in every
  envelope.
- **(c) A caller needs a hard total ceiling** including the envelope. Then add
  `--budget-includes-envelope` (or, better, subtract the envelope's measured cost
  from the body budget up front — the envelope's size is knowable before the footer
  renders as long as the *footer* is excluded from the number it reports). Not
  worth the flag until asked.
- **(d) Descriptions turn out to be the thing agents actually need.** Then it is a
  new line shape and a re-derived default — a DEC edit, not a `--detail` flag bolted
  on (Option I stays rejected; the question is which single shape is right).
- **(e) A second surface wants a different body** (SPEC-074 finds markdown wrong for
  a resource). Then sub-decision 6's "markdown is the budgeted rendering" is what
  moves, and it moves in DEC-045 with this DEC cross-referenced — never by one
  surface quietly rendering its own lines.

**Confidence 0.80**, decomposed honestly:
- Sub-decision 2's **scoping** (never stamped, never a tag, mechanically guarded) —
  **0.95**. The distinction is real, the enforcement is a test, and DEC-027's own
  rejection of Option B is about fabricating *provenance*, not about measuring
  output.
- Sub-decisions 3, 4, 6, 7 (skip-and-continue, the budget region, budget-against-
  markdown, the reported accounting) — **0.88**. Each has a concrete failure it
  prevents and a number or a test behind it.
- Sub-decision 5 (the line shape) — **0.80**. Every field inclusion/exclusion is
  argued, but "which fields at what altitude" is a product judgment that first real
  use will inform.
- Sub-decision 1's **default (2000)** — **0.70**. Grounded in a measured comparison
  (~10% of this repo's 20k-token session-opening ritual) rather than picked, but
  still one number validated against one repo.
- Sub-decision 2's **divisor (4)** — **0.65**. The published English-prose
  approximation, applied to text that is partly identifiers and code fragments.
  Right order of magnitude, certainly not exact.

The composite sits at 0.80 because the load-bearing structural choices are strong
and the two soft spots are single exported constants, each pinned by a golden and
each individually revisitable without touching anything else.

## References

- Related specs: SPEC-073 (emits and implements this DEC), SPEC-074 (pending — the
  MCP resources inherit this budget and this line shape; DEC-045 must not invent a
  second rendering), SPEC-046 (shipped — implemented DEC-027's reserved token-count
  tag, the thing this estimate is scoped away from), SPEC-064 (shipped —
  `internal/capture.Validate`, whose field caps bound a single line at 626 bytes).
- Related decisions: **DEC-027** (reserved cost/session/token-count tags — the
  honesty clause above scopes this estimate away from it; DEC-027's "never
  fabricate" rule governs caller-reported provenance and is untouched),
  **DEC-043** (the ranking this budget trims — emitted together; the line shape is
  the cost the ranking is spent on), DEC-014 (the envelope, the empty-state rule,
  the flat top-level keys, and the markdown `-` convention reused for an absent
  project/type), DEC-013 (`(no project)` / count-block conventions this line shape
  deliberately does **not** reuse, and why), DEC-011 (the 9-key entry shape the
  per-item JSON projection narrows), DEC-031 (`internal/spark` — the precedent for
  a pure primitive package owning a rendering rule, and for JSON staying raw while
  markdown carries the presentation), DEC-042 (`internal/timewindow` — the
  mechanical-guard idiom copied here).
- Related constraints: `no-new-top-level-deps-without-decision` (blocking on Option
  B — no tokenizer, no new dep), `no-cgo`, `no-sql-in-cli-layer`,
  `stdout-is-for-data-stderr-is-for-humans` (the envelope goes to stdout; the
  `--budget` `UserError` to stderr), `test-before-implementation`.
- External: the widely published "~4 characters per token" English-prose
  approximation for BPE tokenizers (the divisor); byte-level BPE (why counting bytes
  rather than runes degrades safely for non-ASCII text).
- Discussions: STAGE-019 Design Notes ("The budget (SPEC-073 / DEC-044)") — this DEC
  adopts its four sub-choices, its skip-and-continue lean, and its instruction that
  budget accounting be a pure function of the rendered bytes; PROJ-006 brief
  (`consult → trust → complete → measure`).

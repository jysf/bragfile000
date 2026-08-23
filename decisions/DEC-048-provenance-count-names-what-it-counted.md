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

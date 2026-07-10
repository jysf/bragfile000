---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-033                         # stable, never reused
  type: decision
  confidence: 0.76                    # honest: the classifier unification + the
                                      # standalone surface + calendar-window reuse
                                      # are high (0.85+); the composite is dragged
                                      # below 0.8 by the sparkline-metric choice
                                      # (agent SHARE vs agent COUNT) and the
                                      # self-reference substring proxy — both
                                      # reversible, both logged in questions.yaml.
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-004
repo:
  id: bragfile

created_at: 2026-07-07
supersedes: null
superseded_by: null

tags:
  - provenance
  - agent-native
  - coverage
  - classifier-unification
  - digest
  - sparkline
  - io-contract
---

# DEC-033: `brag coverage` — the provenance-share metric definition, the `IsAgentAuthored` classifier unification, and its standalone-command surface

## Decision

`brag coverage` is a **new standalone subcommand** — the sixth DEC-014 rule-based
digest consumer — that reports, over a **calendar** reporting window
(`--quarter|--month|--year|--since`, `--previous`, reusing DEC-028/DEC-032
verbatim): the overall **agent-vs-human provenance share**, a **per-month
`{period, agent, human, share}` series** (zero-filled, one bucket per scope
month) rendered with an **agent-SHARE sparkline** (markdown-only, via
`spark.Line`/DEC-031), and a **self-reference density** measure (entries whose
title or description contain the substring `brag`). Three sub-decisions ride
with it:

1. **The classifier is single-sourced.** A pure
   `aggregate.IsAgentAuthored(storage.Entry) bool` predicate is factored into
   `internal/aggregate`, prefix-matching `agent:`/`model:` tokens in
   `Entry.Tags`. It is the Go-side twin of storage's SQL
   `provenanceExistsClause` (`LIKE 'agent:%' OR LIKE 'model:%'`, SPEC-043). The
   two are kept in **agreement by a cross-package test**
   (`TestProvenanceClassifier_GoPredicateMatchesSQLClause`), not by shared code
   — SQL and Go cannot literally share the membership expression. This closes
   the SPEC-043 drift-coupling WATCH.

2. **The trend sparkline encodes agent SHARE, not agent COUNT.** `spark.Line`
   runs over the per-month `share×100` (integer-scaled 0–100), because the
   metric the command exists to surface is *adoption* (what fraction was
   agent-authored) rather than *volume*. JSON stays raw counts + shares, no
   glyphs (DEC-031 choice f); the `--no-spark`/`NO_COLOR` escape is reused
   verbatim from SPEC-052 (`lookupSparkEnv`).

3. **It is a standalone command, not a section of `stats`/`wrapped` or a flag.**
   Provenance-share-over-time is a distinct question; a new command extends
   DEC-014 additively (a new envelope with new keys) without touching any shipped
   digest's locked golden.

No schema change (classifies existing DEC-024 tags); no network; no LLM.

## Context

Action-register **P3**, the read/measure half of the agent-native thesis.
SPEC-043 shipped the SQL classifier + `brag list --author` (the read filter) and
its ship reflection named P3 — "provenance share over time, windowed by month" —
as the natural next spec. At v0.3.0 the baseline is 0% agent-authored (189 human
/ 0 agent); the trend accrues post-v0.3.0 as the MCP write path is used, and this
command is how it becomes visible.

The re-homing (PROJ-003/STAGE-010, never activated → PROJ-004/STAGE-013) and the
maturity of the digest family are what make this cheap now: `impact` (DEC-028)
built the calendar-window core, `wrapped` (DEC-030/SPEC-051) built the monthly
cadence + the DEC-014 renderer pattern, `spark` (DEC-031/SPEC-052) built the
sparkline primitive, and `--previous` (DEC-032/SPEC-053) built the last-completed
modifier. Coverage is assembled almost entirely from these; the genuinely-new
surface is small enough to fit one M spec, but two of its choices — the
classifier unification and the metric definition — are load-bearing and
cross-cutting enough to warrant a DEC rather than being buried in the spec.

Four choices needed a decision:

- **Where the metric lives (surface).** Standalone command vs a `stats`/`wrapped`
  section vs a flag.
- **How the two classifiers stay honest.** Shared code (impossible across the
  SQL/Go boundary) vs a drift-guard test.
- **What the sparkline encodes.** Agent share vs agent count.
- **How to measure self-reference.** Substring vs word-boundary; prose vs tags.

## Alternatives Considered

- **Option A: a `brag stats --provenance` flag / a section in `stats`.**
  - What it is: fold the provenance split into the existing lifetime-stats
    digest.
  - Why rejected: reshapes the locked DEC-014 `stats` envelope and its byte-exact
    goldens; SPEC-043 explicitly deferred "a provenance breakdown in brag stats"
    as needing its own spec. `stats` is also lifetime-scoped and scalar — it has
    no monthly-series slot, so a *trend* would require a new data shape there
    anyway (the exact reason SPEC-052 deferred a `stats` sparkline).

- **Option B: a section inside `brag wrapped`.**
  - What it is: add a provenance-share block to the celebratory year/quarter
    digest.
  - Why rejected: audience mismatch. The PROJ-004 brief separates "reflect / for
    me" (candid) from "promote / for my company" (shareable). A candid
    "how much of my work was agent-assisted" self-metric belongs in the reflect
    lane; `wrapped` is the shareable/celebratory lane. Bundling them conflates two
    purposes the whole project is built to distinguish.

- **Option C: sparkline the agent COUNT per month.**
  - What it is: `spark.Line` over the raw per-month agent-entry counts.
  - Why rejected: conflates *volume* with *adoption*. A month with 10 entries of
    which 2 are agent-authored would out-glyph a month with 2 entries both
    agent-authored, even though the latter is 100% agent-adopted and the former
    20%. The command's thesis is the adoption trend, so the share is the honest
    series. (This is the softest sub-choice — see Validation.)

- **Option D: word-boundary / regex self-reference matching, or matching Tags.**
  - What it is: match `\bbrag\b` or count entries tagged with a brag-related tag.
  - Why rejected: over-engineered for a density *proxy*. "brag" as a substring is
    the simplest honest signal; false positives ("bragging") are near-zero in a
    career-accomplishment corpus and not worth a regex. Tags are the classifier's
    domain (provenance); self-reference is about the prose talking about the tool,
    so Title/Description is the right field.

- **Option E: unify the classifier by sharing one code path across SQL and Go.**
  - What it is: generate the SQL `LIKE` from the Go predicate (or vice versa) so
    there is literally one definition.
  - Why rejected: the two live on opposite sides of the storage boundary
    (`no-sql-in-cli-layer`; SQL only in `internal/storage`, the Go predicate in
    `internal/aggregate`). A shared literal would either leak SQL upward or leak
    a Go dependency into storage's query builder. The honest, low-coupling
    unification is *agreement pinned by a test* — the same shape SPEC-040 used
    for the mcpserver-stamp ↔ storage-classifier drift (`TestServer_
    ProvenanceRoundTripToListAuthor`).

- **Option F (chosen): a standalone `brag coverage` command; `IsAgentAuthored`
  in aggregate pinned to the SQL clause by a cross-package test; the sparkline
  over agent SHARE; self-reference as a `brag` substring in prose.**
  - Why selected: each piece reuses proven machinery (DEC-014 envelope, DEC-028
    window, DEC-031 sparkline, DEC-032 `--previous`, the SPEC-040 drift-guard
    shape) and adds only the small provenance-aware surface; the standalone
    command extends the family additively with zero golden churn on any shipped
    digest.

## Consequences

- **Positive:** the agent-native thesis becomes a *visible trend*, not a claim;
  the classifier drift-coupling WATCH is closed with one predicate + one test;
  `brag list --author` (SQL) and `brag coverage` (Go) can never silently diverge;
  the digest family gains a sixth consumer at low marginal cost (mostly
  transcription from `impact`/`wrapped`).
- **Negative:** a sixth DEC-014 consumer is one more envelope to keep consistent;
  the share-vs-count sparkline choice is a judgment call a reader might disagree
  with (mitigated by the raw per-month counts printed directly beneath the
  glyphs); the self-reference proxy is deliberately coarse.
- **Neutral:** the two classifiers remain two code paths by design — the DEC's
  guarantee is *agreement*, verified continuously by a test, not deduplication.

## Validation

Right if:
- `brag coverage --year` over the real post-v0.3.0 corpus shows a rising
  agent-share trend as the MCP write path is used, and the number reconciles with
  `brag list --author agent --format json | jq length` for the same window (both
  classifiers agree — the whole point).
- The agreement test stays green; if either classifier is edited without the
  other, it fails and names the drift.
- A reader glances at the `Agent share:` sparkline and reads the *adoption*
  shape, with the per-month counts beneath disambiguating any flat/ambiguous
  stretch.

Revisit if:
- The share-vs-count sparkline reads misleadingly in practice (e.g. a
  single-entry 100%-agent month spikes the glyph and overstates adoption) → a
  count-weighted or dual-series variant, or a minimum-entries threshold per
  bucket. Logged as `coverage-sparkline-metric-choice` in
  `guidance/questions.yaml`.
- The `brag`-substring self-reference proxy produces visible false positives →
  tighten to word-boundary or a curated token set (small, localized change).
- Provenance is promoted to first-class `agent`/`model` columns (DEC-024's
  "later, if earned") → `IsAgentAuthored` and `provenanceExistsClause` both
  switch from tag-prefix to a column check, still pinned by the same agreement
  test (which becomes even simpler).

Confidence: **0.76.** The classifier unification (sub-decision 1), the standalone
surface (sub-decision 3), and the calendar-window/`--previous`/envelope reuse are
strong (0.85–0.9) — they reuse shipped, tested machinery and the SPEC-040
drift-guard pattern. The composite is dragged below 0.8 by sub-decision 2 (agent
SHARE vs COUNT for the sparkline): a defensible reader could prefer the count,
and the "single-entry month = 100%" spike is a real if minor readability risk,
mitigated by the adjacent raw counts. The self-reference substring proxy (0.75)
is the second soft spot — coarse by design, reversible. Both soft spots are
escape-hatched and localized; per §14 (< 0.8) a question is logged in
`guidance/questions.yaml` (`coverage-sparkline-metric-choice`).

## References

- Related specs:
  - SPEC-045 (emits this DEC; adds `brag coverage`, `IsAgentAuthored`,
    `CoverageByMonth`, `SelfReferenceCount`, the agreement test).
  - SPEC-043 (shipped) — `provenanceExistsClause` + `brag list --author`, the
    SQL classifier this unifies with; its ship reflection scoped P3.
  - SPEC-040 (shipped) — the mcpserver-stamp ↔ storage-classifier drift-guard
    test shape this reuses for the Go↔SQL agreement.
  - SPEC-048 / SPEC-051 / SPEC-052 / SPEC-053 (shipped) — the calendar-window
    core, the DEC-014 renderer + monthly cadence, the sparkline primitive, and
    `--previous`, all reused.
- Related decisions:
  - DEC-024 — the reserved `agent:`/`model:` provenance namespace both
    classifiers read; the "first-class columns later" revisit trigger.
  - DEC-014 — the rule-based digest envelope `coverage` extends verbatim (sixth
    consumer).
  - DEC-028 — the calendar-window semantics reused unchanged.
  - DEC-032 — `--previous` reused unchanged.
  - DEC-031 — the sparkline primitive + JSON-stays-raw rule (choice f) + the
    `--no-spark`/`NO_COLOR` escape.
  - DEC-015 — the normalized tags/taggings join both classifiers operate over.
- Related constraints: `no-sql-in-cli-layer` (SQL stays in storage; the Go
  predicate stays in aggregate), `stdout-is-for-data-stderr-is-for-humans`,
  `test-before-implementation`, `errors-wrap-with-context`.
- Related questions: `coverage-sparkline-metric-choice`
  (`guidance/questions.yaml`) — the share-vs-count sparkline soft spot.
- Related docs: `docs/api-contract.md` (gains a `brag coverage` section — SPEC-045).

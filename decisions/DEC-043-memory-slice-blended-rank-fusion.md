---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-043
  type: decision
  confidence: 0.75                   # honest: the STRUCTURE (rank fusion, no
                                     # score normalization, clean degeneration
                                     # to recency) is high-confidence and
                                     # forced by the incomparability of the two
                                     # orderings; the residual soft spot is the
                                     # CONSTANTS (k=60 and equal weights are
                                     # borrowed defaults, unvalidated against a
                                     # real personal corpus) — see Validation.
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
  - ranking
  - retrieval
  - determinism
---

# DEC-043: The memory slice's blended ranking — reciprocal-rank fusion over three ordinal lists, with the constants locked

## Decision

`internal/memory.Slice` ranks a candidate pool by **reciprocal-rank fusion (RRF)**
over **three ranked lists** — recency, relevance, and project membership. For an
entry *e*:

```
score(e) = Wr/(k + rank_recency(e))
         + Wm/(k + rank_match(e))
         + Wp/(k + rank_project(e))
```

An entry **absent** from a list contributes **0** for that term. Six coupled
sub-decisions are locked:

1. **The three lists.**
   - *recency* — every candidate, ordered `created_at DESC, id DESC` (`Store.List`'s
     documented order, **re-derived inside `Slice`** from the entries themselves so
     the function is total and depends on no input-order invariant).
   - *relevance* — the `Store.Search` result order (FTS5 `ORDER BY rank, id DESC`),
     passed in as `Options.Matched []int64`, an ordered ID list. Empty/nil when no
     query was given. IDs not present in the pool, and duplicate IDs, are ignored.
   - *project* — when `Options.Project` is non-empty, the candidates whose
     `Project` equals it exactly, **in recency order**. Empty when no project was
     given, or when the named project has no entries.

2. **The constants are locked and pinned by goldens.** `k = 60`, `Wr = Wm = Wp = 1.0`.
   `k = 60` is the published RRF constant (Cormack, Clarke & Buettcher, 2009), taken
   **because it is published and we have no tuning corpus** — a hand-picked value
   would be a number nobody could defend six months from now. Equal weights follow
   from the same reasoning: RRF's premise is that the fused lists are *peers*, and
   nothing in a personal corpus tells us relevance should outweigh recency by
   1.4× rather than 1.7×. Both constants are exported (`memory.FusionK`,
   `memory.WeightRecency` / `WeightMatch` / `WeightProject`) and pinned by
   byte-exact goldens, so any future change to either is a visible golden diff, not
   a silent behavior change.

3. **Recency is ORDINAL, not a time decay — so the ranking reads no clock.** The
   recency signal is a *rank*, never an age in days. Three consequences, all
   intended:
   - The ranking is a pure function of `(entries, options)` with **no `now` input
     at all** — a strictly stronger determinism property than the "inject the now
     used for recency decay" shape STAGE-019 anticipated. Byte-exact goldens for a
     ranking follow trivially. Enforced mechanically by
     `TestPackageReadsNoWallClock`.
   - It **self-normalizes to the developer's capture cadence**. For someone logging
     daily, "recency rank 20" is last week; for someone logging monthly it is last
     year. A half-life decay would need a constant that is wrong for one of them.
   - It is what makes the fusion *principled*. Mixing an ordinal bm25 rank with a
     continuous decay reintroduces exactly the normalization problem that
     disqualifies Option B below.

4. **`--project` is a SOFT boost, never a hard filter.** A named project raises its
   entries; it never removes anyone else's. Reasons, in order of weight: (a) the
   hard filter already exists and is one command away (`brag list --project X`,
   `brag_list {"project":"X"}`) — duplicating it here would make `memory` strictly
   weaker than the tools it composes; (b) `entries.project` is a **soft string
   match** (DEC-017) and is frequently empty, so a hard filter silently drops
   relevant history rather than deprioritizing it; (c) the slice's job is *what
   should I know before I act*, and the cross-project decision that applies here is
   precisely the thing a write-only log fails to surface. No `--project-only` flag
   ships (YAGNI; the revisit trigger is in Validation).

5. **Candidate pool: one bounded read per list, `PoolLimit = 200` each.** The
   caller composes **`Store.List` + `Store.Search` + a project-scoped `Store.List`**
   in Go — no new storage method, no new SQL. Each read is capped at
   `memory.PoolLimit` (200), so the pool is ≤ 600 rows before dedup. The cap is an
   implementation bound, **not a user knob**.

   **What the cap actually guarantees** (corrected at design review — an earlier
   draft of this paragraph claimed the cap "cannot" cost a ranking, which does not
   follow, because the three terms *sum*). The worst-case excluded entry sits at
   rank 201 in all three lists and would have scored `3/261 ≈ 0.011494`; the
   weakest *included* entry sits at recency rank 200 with no match and no project
   and scores `1/260 ≈ 0.003846`. So an excluded entry **can** outrank an included
   one, and the cap is lossy in the tail. The guarantee that does hold is a head
   guarantee: `1/(60 + r) > 3/261` for all `r ≤ 26`, so **any entry in the top 26
   of any single list cannot be displaced by anything the cap excluded.** Below
   that, ordering may be perturbed — and at the default budget (≈110 entries) the
   tail of a typical slice is inside that zone.

   This is accepted, not hidden: the head is what a memory slice is read for, the
   perturbation is confined to entries that are marginal on *all three* axes at
   once, and the alternative (uncapped reads) trades a bounded tail imprecision
   for an unbounded read on every invocation. Raising `PoolLimit` widens the head
   guarantee — it is exported and golden-pinned, so the change is visible.

   One pool per list is what keeps even the head guarantee true — drop the project
   read and old same-project history becomes unreachable at any rank.

6. **Determinism details.** `Slice` **dedupes by ID** (keeping one copy; the sort
   key makes duplicates adjacent), sorts by `score DESC` with a **stable** sort over
   the recency-ordered pool — so equal scores fall back to `created_at DESC, id DESC`
   with no separate comparator — and emits `Items` in final rank order with a 1-based
   `Rank`. Float arithmetic is IEEE-754 and therefore reproducible; the JSON envelope
   exposes `score` **rounded to 6 decimal places** so the goldens stay readable and
   the constants stay observable.

## Context

STAGE-019's premise is that the corpus becomes working memory an agent consults
*before* it acts. The read layer already ships (`Store.List`, `Store.Search`,
`brag_list`, `brag_search`), but they answer along **one axis each** and the two
axes are **incomparable**:

- `Store.List` returns `created_at DESC, id DESC` — an ordinal position in time.
- `Store.Search` returns FTS5 `ORDER BY rank` — bm25, which is **negative**,
  **unbounded**, and **corpus-relative** (its scale depends on document lengths and
  term frequencies across the whole index, so the same entry scores differently as
  the corpus grows).

There is no shared unit. Any blend must either invent one — a normalization that is
guesswork, and guesswork that silently drifts as the corpus grows — or discard the
magnitudes and fuse the **positions**. Rank fusion is the standard answer to exactly
this problem, and it is the reason RRF exists.

Two properties made the choice, beyond "it is standard":

- **One algorithm covers both cases.** The memory slice has two callers with
  different questions: the zero-config session-start auto-load (*"what have I been
  doing?"* — no query, no project) and the targeted ask (*"what do I know about
  auth on orbit?"*). With RRF the first is not a special case: with two empty
  lists the score reduces to `1/(60 + rank_recency)`, a strictly decreasing function
  of rank, so the output is **exactly** plain recency order. No branch, no second
  code path, no second set of goldens to keep honest.
- **Determinism is free.** No normalization means no corpus-relative term, which
  means the same inputs produce the same bytes forever. That is what makes the
  goldens the contract rather than a snapshot.

## Alternatives Considered

- **Option A (chosen): reciprocal-rank fusion over the three ordinal lists.**
  - Why selected: it is the only candidate that needs no invented scale, degenerates
    to plain recency for free, reads no clock, and is cheap to explain (*"we add up
    1/(60+position) across the lists you asked about"*). Its cost — the constants
    are borrowed rather than tuned — is real but bounded, visible in the DEC, and
    pinned by goldens.

- **Option B: weighted linear combination over normalized components.**
  - What it is: normalize bm25 into `[0,1]` (min-max over the result set, or a
    sigmoid), normalize recency into `[0,1]` via an age decay, then take a weighted
    sum.
  - Why rejected: the normalization is the fragile part, and it is fragile in the
    way that is hardest to notice. Min-max over the result set makes an entry's
    score depend on *which other entries matched*, so adding one unrelated brag can
    reorder a slice that did not change; a sigmoid needs a scale constant with even
    less justification than `k`. Worse, both make the ranking **corpus-relative**,
    which breaks byte-exact goldens for anything but a frozen fixture and makes the
    behavior un-explainable to the user at exactly the moment they ask "why is this
    first?". This option trades a defensible borrowed constant for two indefensible
    invented ones.

- **Option C: lexicographic tiers** (`matched ∧ same-project` > `matched` >
  `same-project` > `recent`, recency within tier).
  - What it is: bucket the candidates into four ordered tiers, sort by recency
    inside each.
  - Why rejected — and this is the closest call: it is genuinely the most
    **explainable** option, and it needs no constants at all, which is a real
    virtue for a decision whose weakest point is its constants. It loses on
    **tunability and on truthfulness at the boundary**. Tiers are absolute: the
    single most relevant entry in the corpus loses to *any* same-project-and-matched
    entry, however marginal its match, and a brilliant match from a different
    project can never reach the top. There is no dial between "tier boundaries are
    right" and "tier boundaries are wrong" — the only fix is a different tier
    ordering, i.e. a new DEC. It also throws away the bm25 ordering *within* the
    matched tier (falling back to recency), discarding the one signal the query
    was asked for. Kept on the record because if the k/weights question in
    Validation resolves badly, this is the fallback, not Option B.

- **Option D: recency-only inside an FTS pre-filter.**
  - What it is: when a query is given, `Store.Search` first, then re-sort the
    matches by recency and take the top N.
  - Why rejected: it is not a blend — it is `brag search | sort by date`, which a
    caller can already do. It discards bm25 entirely (the ranking the FTS index
    exists to produce), it cannot include a highly relevant old entry alongside
    recent context, and with no query it degenerates to `brag list`, so the command
    would add nothing over the two tools it wraps. The stage listed it for
    completeness; it fails the stage's own success criterion ("its ordering
    demonstrably differs from both plain recency and plain bm25").

- **Option E: `--project` as a hard filter (or a `--project-only` flag alongside
  the boost).**
  - What it is: set `ListFilter.Project` and rank only that project's entries; or
    ship both behaviors behind two flags.
  - Why rejected: see sub-decision 4. The hard filter is already reachable through
    `brag list`/`brag_list`, so shipping it here buys nothing and costs the
    cross-project recall that is the slice's distinctive value. The two-flag form
    additionally asks the caller to understand a distinction they have no basis to
    make on their first invocation. Pre-authorized reversal in Validation if
    project-scoped slices measurably leak noise.

- **Option F: embeddings / a semantic ranker.**
  - Why rejected: out of scope by stage decree and by constraint — a model in the
    binary breaks local-first, `no-cgo`, and `no-new-top-level-deps-without-decision`
    simultaneously. Recorded so the omission reads as a decision rather than an
    oversight.

## Consequences

- **Positive:** the blend demonstrably differs from **both** inputs. On the spec's
  fixture (`--query auth --project orbit`), plain recency is `8 7 6 5 4 3 2 1`,
  plain bm25 is `5 1 2 4`, and the blend is **`4 2 7 5 1 8 6 3`** — neither list, and
  the entry that leads (`4`, an auth refactor on orbit carrying a real impact line)
  is 5th by recency and *last* by bm25. That inversion is the whole point: bm25
  ranked the shortest document first (`5`, "Read the auth spec" — a title with no
  project and no impact), and the project and recency terms pull the entry that
  actually matters back to the top.
- **Positive:** no clock in the ranking core at all, so the goldens are byte-exact
  without freezing time, and `internal/memory` gets the same mechanical purity guard
  `internal/timewindow` has.
- **Positive:** no storage change. Three existing methods composed in Go, exactly as
  STAGE-019 required.
- **Negative (accepted):** `k = 60` is tuned for TREC-scale corpora (thousands of
  documents) and **compresses the top of a small personal corpus** — ranks 1 and 20
  differ by only ~24% of their own magnitude. In practice the slice is ordered
  correctly but the *scores* look nearly flat, which will read as suspicious to
  anyone inspecting `--format json`. The honest position: the ORDER is what is
  consumed, the scores are exposed for auditability, and a smaller `k` is the first
  thing to try if the ordering itself disappoints (Validation).
- **Negative (accepted):** the project boost is strong. With `k = 60` the whole
  recency term spans `1/61 → 0` (≈0.0164), so a project term of `1/(60+rank_p)`
  is comparable to the entire recency range — a same-project entry effectively
  outranks a non-project entry of similar relevance almost regardless of age. That
  is close to a hard filter in effect while remaining soft in mechanism (nothing is
  removed, and a strong match still competes). Named here so it is not later
  discovered as a bug: it is the direct arithmetic consequence of `k = 60` and
  equal weights, and it moves if either constant moves.
- **Neutral:** `Options.Matched` is an ordered ID list rather than a second entry
  slice. It keeps `Slice`'s signature to the one the stage specified
  (`Slice(entries, opts) Result`) and keeps FTS out of the pure package, at the cost
  of an ordering contract the caller must honor. Documented on the field; SPEC-074's
  MCP caller must honor the same contract (it composes the same three reads).
- **Neutral:** the pool cap means the slice is *not* a whole-corpus ranking. For any
  budget this command can serve, the distinction is unobservable (sub-decision 5),
  but a future caller wanting a full ranking must raise the cap deliberately.

## Validation

**Right if:**
- With no query and no project, `Slice`'s output order is **identical** to
  `Store.List`'s — pinned by `TestSlice_DegeneratesToRecencyWithNoQueryOrProject`.
- With a query and a project, the output order differs from **both** plain recency
  and plain bm25 over the same fixture — pinned by
  `TestSlice_BlendDiffersFromBothInputs` and by the byte-exact markdown golden.
- The same entries in a **different input order**, and with duplicates, produce
  **byte-identical** output — pinned by `TestSlice_InputOrderInvariantAndDedupes`.
- No file in `internal/memory` calls `time.Now(` — pinned by
  `TestPackageReadsNoWallClock`.
- Changing `k` or any weight breaks a golden. If a golden ever changes without a
  DEC edit, this decision has been violated silently, and that is the question to
  answer before updating the golden.

**Revisit if:**
- **(a) The constants disappoint in real use** — the top of the slice feels wrong,
  or `score` inspection shows the ordering is being decided by a term the user did
  not intend. First move: lower `k` (10–20) to sharpen the top of a personal-scale
  corpus, which also *weakens* the project boost relative to recency, addressing
  both accepted negatives at once. Second move: unequal weights, which requires a
  fixture set richer than one hand-built corpus to justify. Logged as
  `memory-slice-fusion-constants` in `guidance/questions.yaml`.
- **(b) Project-scoped slices leak noise** — a `--project X` slice is dominated by
  unrelated entries in practice. Then add the hard filter as an explicit opt-in
  (Option E's two-flag form), **not** by changing the default: the soft boost is
  what makes cross-project recall possible, and a default flip would remove it
  silently.
- **(c) A fourth signal earns a term** (tag membership, `type`, provenance/author).
  RRF extends by adding a list, so this is additive and cheap — but each new list
  dilutes the others at fixed weights, so a fourth term is the moment the equal-
  weights premise must actually be defended rather than inherited.
- **(d) The pool cap becomes observable** — a caller wants a slice larger than
  `PoolLimit` entries, or the corpus grows enough that "top 200 by recency" excludes
  something a user expects. Raise the constant; it is exported and pinned.
- **(e) A semantic ranker stops being out of scope** (a pure-Go, dependency-free
  embedding path). Then it is a fourth list, not a replacement — the fusion
  structure survives.

**Confidence 0.75**, decomposed honestly:
- Sub-decision 1 + 3 + 6 (three ordinal lists, ordinal recency, determinism) —
  **0.92**. Forced by the incomparability of bm25 and recency; the alternatives are
  worse for reasons that are structural, not preferential.
- Sub-decision 4 (project as a soft boost) — **0.80**. Well-argued and cheaply
  reversible, but genuinely a product call rather than a forced one.
- Sub-decision 5 (pool cap, one read per list) — **0.80** (was 0.85 before design
  review). The head guarantee is arithmetic and holds; the tail is knowingly lossy
  and the specific value 200 is round. Lowered because the first draft of this
  sub-decision asserted a stronger property than the arithmetic supports — the
  structure survived the check, the claimed bound did not.
- Sub-decision 2 (`k = 60`, equal weights) — **0.55**. These are borrowed defaults
  applied to a corpus two to three orders of magnitude smaller than the one they
  were derived on, and the accepted negatives above are their direct consequence.
  They are the right *starting* values (published, defensible, not invented) and
  the wrong thing to be confident about.

The composite reflects sub-decision 2 dragging a structurally strong decision down.
Per §14 (< 0.8) a question is logged in `guidance/questions.yaml`.

## References

- Related specs: SPEC-073 (emits and implements this DEC), SPEC-072 (shipped — the
  filter parity and the `internal/timewindow` precedent for a shared pure package),
  SPEC-074 (pending — the MCP resources + `brag_memory` tool, which consumes
  `memory.Slice` and must honor the same `Matched` ordering contract),
  SPEC-011 (the FTS5 index whose bm25 order is fused here).
- Related decisions: **DEC-044** (the token budget and the per-entry line shape —
  emitted with this one; the two are only separable on paper, since the line shape
  *is* the cost that the ranking is trimmed against), DEC-014 (the envelope the
  slice renders into), DEC-010 (`brag search` query transform — `--query` reuses
  `buildFTS5Query` verbatim), DEC-011 (the JSON entry shape `memory`'s per-item
  projection deliberately narrows), DEC-017 (`entries.project` is a soft string
  match — why a hard project filter would silently drop history), DEC-042
  (`internal/timewindow` — the shared-pure-package precedent, and the source of the
  `TestPackageReadsNoWallClock` guard copied here), DEC-024 (the MCP contract
  SPEC-074 extends), DEC-031 (`internal/spark` — the closest structural template: a
  small pure primitive package plus a markdown-only rendering rule).
- Related constraints: `no-sql-in-cli-layer` (blocking — `internal/memory` imports
  `internal/storage` for the `Entry` type only, never a driver),
  `no-new-top-level-deps-without-decision` (none added; stdlib only),
  `no-cgo`, `errors-wrap-with-context`, `test-before-implementation`.
- External: Cormack, Clarke & Buettcher, *"Reciprocal Rank Fusion outperforms Condorcet
  and individual Rank Learning Methods"* (SIGIR 2009) — the source of `k = 60` and of
  the "fuse positions, not scores" premise. SQLite FTS5 `bm25()` / `ORDER BY rank`
  documentation (the negative, corpus-relative score this decision refuses to
  normalize).
- Discussions: STAGE-019 Design Notes ("The blend (SPEC-073 / DEC-043)") — this DEC
  adopts its lean (a) and its "lock the constants, pin them with goldens"
  instruction; PROJ-006 brief (`consult → trust → complete → measure`).

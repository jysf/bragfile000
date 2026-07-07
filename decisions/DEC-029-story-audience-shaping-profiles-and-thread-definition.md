---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-029
  type: decision
  confidence: 0.72
  audience:
    - developer
    - agent
    - executive

agent:
  id: claude-opus-4-8
  session_id: null

project:
  id: PROJ-004
repo:
  id: bragfile

created_at: 2026-07-06
supersedes: null
superseded_by: null

tags:
  - cli
  - io-contract
  - narrative
  - story
  - audience
  - llm-pipe
  - shaping-profile
  - aggregation
  - human-consumer
  - ai-consumer
---

# DEC-029: `brag story --audience` — deterministic threads, the LLM finds the throughline; audiences are data-driven shaping profiles

## Decision

`brag story --audience <name>` emits an **audience-shaped, arc-aware
bundle** that coalesces the corpus into narrative **threads** for an LLM
(already in the caller) to weave into prose. Synthesis stays a **pure
pipe** — no model, no network, no secrets in the binary (preserving
DEC-001, matching `brag review`/`summary`/`impact`). Seven choices are
locked:

1. **A "thread" is a DETERMINISTIC unit bragfile assembles; the LLM
   finds the THROUGHLINE.** The split is load-bearing and is the heart
   of this DEC. bragfile assembles:
   - **threads = initiative (the `project` field, DEC-028's axis),**
     each thread time-ordered ASC, carrying its entries as **beats**;
   - **impact beats** — within a thread, entries that carry a non-empty
     `impact` are marked as impact-bearing beats (reusing
     `aggregate.WithImpact`), so the "so what" of a thread is
     machine-identified, not left for the LLM to guess;
   - an **optional theme-tag cross-cut** — when `--theme <tag>` is
     given, a single extra cross-project thread groups every in-window
     entry carrying that tag, time-ordered, so a cross-cutting arc
     (e.g. `perf`, `reliability`) that spans initiatives is offered as a
     first-class thread;
   - a **throughline skeleton** — the ordered list of thread IDs plus,
     per thread, its span (first/last beat date), its beat count, and
     its impact-beat count. This is the *skeleton* of the arc: which
     threads exist, in what order, how big, where the "so what" lands.

   bragfile does **NOT** decide the single narrative throughline that
   ties threads together, does **NOT** infer cross-project arcs beyond
   the explicit `--theme` cut, and does **NOT** write prose. The
   **framing directive + the caller's LLM** merge threads into one arc,
   pick the headline, and write the words. **bragfile assembles threads;
   the LLM finds the throughline.**

2. **Audiences are extensible, DATA-DRIVEN shaping profiles — NOT a
   hard-coded Go enum.** A profile is a struct populated from a bundled
   default asset (embedded via `embed.FS`) and, if present, a
   user-override file (`~/.bragfile/story-profiles/<name>.yaml` or the
   config dir; see SPEC-049 for the exact path + precedence). Each
   profile carries: a **selection filter** (window default + optional
   type/impact bias), a **threading policy** (how many threads to
   surface + whether to fold small threads), an **altitude/length**
   directive (target arc count + target length + candor level), and a
   pointer to its **framing-directive asset**. Adding an audience is
   adding a profile file — no Go change, no recompile for user profiles.

3. **`--audience` sets HOW MANY arcs and AT WHAT ALTITUDE, not just
   filter and tone.** This is the visible, rule-driven difference the
   stage requires:
   - **`me`** (candid / reflect): every thread surfaces, small threads
     are NOT folded, impact-less beats are KEPT (the messy middle +
     lessons are the point), altitude = low, target = many arcs, candor
     = high. Window default: `--year`.
   - **`exec`** (impact-forward / promote): only threads that carry at
     least one impact beat surface, small threads are folded, impact-less
     beats are DROPPED, altitude = high, target = one headline arc,
     candor = promotional/terse. Window default: `--quarter`.

   Same corpus, same window → demonstrably different bundles
   (thread count, beat inclusion, throughline-skeleton altitude), driven
   by profile rules, not by a tone hint.

4. **SPEC-049 ships the `me` and `exec` profiles — the gradient
   ENDPOINTS.** They maximally exercise "same in theory, diverge in
   practice" (candid-every-thread vs promotional-one-arc). `manager`
   (and optionally `skip`) are pure config additions deferred to
   **SPEC-050**, which then *proves* the mechanism is extensible by
   adding an audience with zero Go change. (Complexity split — see
   SPEC-049.)

5. **The bundle EXTENDS DEC-014's envelope with an arc-aware body; it
   does not fit the flat DEC-028 shape.** A story is threads → beats →
   throughline-skeleton, which the flat `impact_by_project` array cannot
   carry (it has no beat ordering across all beats, no impact-beat
   marking, no skeleton, no directive reference). So `brag story` keeps
   DEC-014's provenance envelope (`generated_at`, `scope`, `filters`,
   2-space JSON, empty-state rules) and adds arc-aware top-level keys:
   `audience`, `threads` (array of `{thread, kind, span, beats:[{id,
   title, project, type, impact, is_impact_beat, created_at}]}`),
   `throughline` (the skeleton: ordered thread refs + counts), and
   `framing_directive` (the resolved directive text OR a reference to
   it). Markdown mirrors this: provenance block, then a per-thread
   section with beats, then a `## Throughline (skeleton)` block, then the
   framing directive appended (or referenced) so the bundle is a
   complete paste-in artifact.

6. **The bundle is USEFUL STANDALONE; the LLM is an optional upgrade.**
   Without any LLM, the markdown bundle is a readable, ordered,
   audience-shaped digest (threads, beats, the "so what" impact beats
   flagged, the skeleton) — pasteable into a review doc or a 1:1 note as
   is. The framing directive is appended as an instruction block a human
   can also just read. The LLM turns the skeleton into woven prose; it
   is never required for the bundle to have value. (Same posture as
   `brag review`'s "paste into an AI session" — the digest stands alone.)

7. **Framing directives are BUNDLED ASSETS, per-audience, embedded in
   the binary** (like the migrations `embed.FS`, and like the plugin's
   checked-in `commands/brag.md` / `BRAG.md` asset convention). They
   live under `internal/story/directives/<audience>.md`, embedded via
   `//go:embed`, resolved by profile name, and either appended to the
   bundle or printed alone via `brag story --print-directive
   --audience <name>`. A user profile may point at its own directive
   file. The directive is DATA (an asset), not Go string literals
   scattered in the renderer.

## Context

STAGE-012 is the narrative headline of v0.4.0. The settled posture (from
the orchestrator + user, not re-litigated here): pure pipe, LLM optional;
the core capability is **coalescing a set of brags into narrative arcs**,
not merely filtering/grouping (that is `brag impact`'s job); audience sets
*how many arcs and at what altitude*; audiences are **extensible profiles,
not an enum**; and it **reuses `brag impact`'s (DEC-028) grouped data**.

The single genuinely-open fork this DEC exists to resolve: **what defines
a thread/arc — the unit bragfile coalesces?** Four candidates were on the
table (initiative/project, theme-tags, time-progression, LLM-inferred).
Getting this wrong in either direction is the risk: too deterministic and
`brag story` is just `brag impact` with a costume; too much inference in
the binary and we either bake in an LLM (violates DEC-001) or ship
non-deterministic, untestable output.

The resolution threads the needle: bragfile owns the **deterministic
scaffolding** (threads = initiative, time-ordered, impact beats marked,
optional theme cross-cut, a throughline *skeleton*) and the LLM owns the
**synthesis** (merging threads into one arc, choosing the headline,
writing prose). Time-progression and theme are folded IN as *properties
of* / *cross-cuts across* the deterministic threads (ordering + the
`--theme` cut), not as a competing thread definition. LLM-inferred
threading is rejected as the primary mechanism because it pushes
coalescing into the model — but it is *reintroduced at the right layer*:
the framing directive explicitly invites the LLM to find cross-cutting
throughlines the deterministic threads only hint at.

## Alternatives Considered

### Thread-definition fork (choice 1)

- **Option A: Initiative/project only (what `brag impact` already
  groups by), no more.**
  - What it is: A thread IS a project group; the bundle is `brag impact`
    plus a tone hint per audience.
  - Why rejected: This is a *report*, not a *story*. It cannot express a
    throughline across initiatives, cannot mark the "so what" beat within
    a thread, and makes audience a cosmetic tone flag — failing the
    stage's explicit "coalesce into arcs, not filter/group" and
    "rule-driven, not a tone hint" criteria. Kept as the *base* thread
    unit (deterministic, reuses DEC-028), but not the whole answer.

- **Option B: Theme-tags as the primary thread axis.**
  - What it is: Threads are tag-clusters spanning projects; the story is
    organized by theme (`perf`, `auth`), not initiative.
  - Why rejected as primary: tags are sparse and inconsistent in the real
    corpus (freeform, DEC-004), so theme-primary threading is fragile and
    would silently drop untagged work. But a theme cross-cut is genuinely
    valuable, so it is folded in as the **optional `--theme` cross-cut
    thread** (choice 1) — offered, not imposed.

- **Option C: Time-progression as the thread axis (the chronological
  climb).**
  - What it is: A thread is a run of related-over-time entries; the story
    is "the climb."
  - Why rejected as a standalone axis: "related over time" with no
    grouping key IS LLM-inference in disguise (what makes two entries
    part of the same climb?). Time-progression is real and is folded in
    as the **within-thread ASC ordering + the span in the skeleton** — a
    *property* of every thread, not a competing definition of one.

- **Option D: LLM-inferred threads.**
  - What it is: The binary asks a model to cluster entries into arcs.
  - Why rejected: two ways to lose. Either bake an LLM into the binary
    (violates DEC-001 / the settled pure-pipe posture / no-network) or
    make the bundle depend on a model to even have structure (breaks
    "useful standalone"). Inference is reintroduced at the correct layer
    — the framing directive tells the *caller's* LLM to find
    cross-cutting throughlines — so the value is captured without the
    cost.

- **Option E (chosen): Deterministic threads (initiative + optional
  theme cross-cut) + time-ordering + impact beats + a throughline
  skeleton; the framing directive + LLM weave the arc.**
  - What it is: choice 1 above.
  - Why selected: it is the only option that (a) keeps the binary
    deterministic and testable (byte-exact goldens over threads/beats/
    skeleton), (b) preserves the pure-pipe / no-model posture, (c)
    delivers a *story* not a *report* (throughline skeleton + marked
    impact beats + cross-cut), and (d) puts the genuinely-inferential
    work (merging threads into ONE arc, picking the headline) where it
    belongs — in the caller's LLM, invited by the directive. It absorbs
    B and C as a cross-cut and an ordering-property rather than
    discarding them.

### Audience-mechanism fork (choice 2)

- **Option F: Hard-coded Go `enum` + a `switch` per audience.**
  - What it is: `type Audience int; const (Me; Manager; Exec)` with the
    shaping rules in Go.
  - Why rejected: the stage explicitly requires "extensible profiles,
    not a locked enum." An enum means every new audience (skip, board,
    peer, self-6-months-from-now) is a Go change + recompile + release.
    Profiles-as-data let a user add `board.yaml` locally.

- **Option G (chosen): Data-driven profiles — bundled defaults
  (`embed.FS`) + optional user-override files.**
  - Why selected: extensibility is the point; a profile is config
    (selection + threading + altitude + directive pointer). Ship
    sensible defaults, let users override/extend without touching Go.
    Mirrors how the plugin already ships behavior as checked-in assets
    (`commands/brag.md`, `hooks.json`) rather than compiled-in strings.

### Bundle-shape fork (choice 5)

- **Option H: Reuse DEC-028's flat `impact_by_project` envelope
  verbatim.**
  - Why rejected: it has no cross-thread beat ordering, no impact-beat
    marking, no throughline skeleton, no directive slot. A story needs
    an arc-aware shape. Reusing it would force the arc structure into the
    LLM's head with no deterministic scaffold — exactly the failure mode
    choice 1 avoids.

- **Option I (chosen): Extend DEC-014's provenance envelope with
  arc-aware top-level keys (`audience`, `threads`, `throughline`,
  `framing_directive`).**
  - Why selected: keeps the stable, tested provenance surface (`jq
    .scope`, `.filters`, empty-state discipline) while adding exactly the
    keys a narrative pipe needs. Backward-compatible extension pattern
    DEC-014 already blesses (consumers ignore unknown keys). The
    per-consumer payload divergence is the same latitude SPEC-018 (maps)
    vs SPEC-020 (arrays) vs SPEC-048 (grouped array) already exercise.

## Consequences

- **Positive:** `brag story` is a genuine narrative surface, not a
  reskinned `brag impact`: the throughline skeleton + marked impact beats
  + optional theme cross-cut give an LLM real arc scaffolding while the
  binary stays deterministic, local, model-free, and byte-testable.
  Audiences-as-data make the taxonomy open-ended (SPEC-050 adds `manager`
  with no Go change — the extensibility proof). The bundle is useful
  standalone (a readable audience-shaped digest) and an optional LLM
  upgrade. Reuses `aggregate` (`GroupEntriesByProject`, `WithImpact`) and
  the DEC-014 envelope + DEC-028 window machinery wholesale.
- **Negative:** A fourth output shape family enters the codebase
  (DEC-011 array; DEC-014 flat digest; DEC-028 impact-grouped; DEC-029
  arc-aware). More surface to keep coherent. Mitigated: DEC-029 shares
  DEC-014's provenance block and DEC-028's window helper, so only the
  *body* is new. A `story-profiles` config path + user-override precedence
  is new config surface (validation, malformed-file handling) — scoped
  and specified in SPEC-049.
- **Negative:** The value of `brag story` is realized only when a capable
  LLM consumes the bundle; a user with no LLM gets "only" a shaped digest.
  Accepted and by design (pure pipe); the standalone-usefulness bar
  (choice 6) keeps the floor high.
- **Neutral:** The `me`/`exec` endpoints ship first; the middle of the
  gradient (`manager`/`skip`) lands in SPEC-050. The gradient is designed
  whole; only the shipped set is sliced.
- **Neutral:** Confidence is dragged below 0.8 by the profile-override
  file *format + precedence* (the mechanism is right; the exact schema is
  a fresh surface) and by whether two audiences suffice to prove
  divergence at design time — both recorded as questions.

## Validation

Right if:
- `brag story --audience me` and `brag story --audience exec` over the
  SAME corpus + window produce **demonstrably different** bundles: `exec`
  surfaces fewer threads (impact-bearing only), folds small threads,
  drops impact-less beats, and its throughline skeleton targets one
  headline arc; `me` surfaces every thread with impact-less beats kept.
  The difference is visible in the deterministic body, not only in the
  appended directive.
- The `threads` array is deterministic and byte-golden-testable: thread
  order (initiative alpha-ASC, `(no project)` last, theme cross-cut
  placement locked), within-thread beat order (ASC + ID tiebreak),
  `is_impact_beat` marking matches `aggregate.WithImpact`.
- `brag story --audience exec --print-directive` prints the bundled exec
  framing directive and nothing else; the same directive is
  referenced/appended in a normal run.
- A user dropping `~/.bragfile/story-profiles/board.yaml` gets a working
  `brag story --audience board` with no recompile.
- The bundle pasted into an LLM session with its directive yields an
  audience-appropriate narrative; the SAME bundle read by a human with no
  LLM is still a coherent shaped digest.

Revisit if:
- Real corpora show initiative-only threading is too coarse (many entries
  land under `(no project)`); then promote theme or a new grouping key to
  a first-class thread axis via a follow-up DEC (do not silently change
  the default).
- The profile-override file format needs to carry logic the flat schema
  can't express (then version the profile schema).
- An audience genuinely needs the previous *complete* period; reuse
  DEC-028's deferred `--previous` rather than changing the window default.
- Two shipped audiences turn out not to prove divergence convincingly in
  practice; add the third (`manager`) earlier.

Confidence: 0.72. The thread-definition core (choice 1) is the
strongest-reasoned part (~0.8): it is the only option preserving
determinism + the pure-pipe posture while still delivering an arc, and it
absorbs the theme/time alternatives rather than discarding them. The
composite is dragged below 0.8 by two softer sub-choices, each filed as a
question (§14): (a) the profile-override **file format + precedence**
(the profiles-as-data *mechanism* is high-confidence, but the concrete
YAML schema, config path, and malformed-file behavior are a fresh surface
— 0.65); (b) whether the `me`/`exec` **two-audience slice** proves
"diverge in practice" at design time or needs `manager` in the same
release to be convincing (0.7). Both are recorded in
`/guidance/questions.yaml`.

## References

- Related specs:
  - SPEC-049 (emits this DEC; wires `brag story --audience me|exec`; adds
    `internal/story/` — profiles, directives asset, bundle renderer;
    reuses `aggregate.GroupEntriesByProject` + `aggregate.WithImpact` and
    DEC-028's `windowCutoff`).
  - SPEC-050 (planned; adds `manager`/`skip` profiles as data — the
    extensibility proof, zero Go change).
  - SPEC-048 (shipped; `brag impact`, DEC-028 — the grouped/impact data
    `brag story` coalesces).
- Related decisions:
  - DEC-014 (rule-based output envelope) — EXTENDED: provenance block +
    empty-state + 2-space JSON inherited; arc-aware body added.
  - DEC-028 (impact digest window + shape) — REUSED: `windowCutoff`
    (calendar windows), `project`=initiative axis, `WithImpact`
    impact-first split, the 4-key entry projection widened to a beat
    projection.
  - DEC-001 (pure-Go, no-CGO / local-first posture) — PRESERVED: no
    model, no network, no secrets in the binary; synthesis is a pipe.
  - DEC-011 (JSON per-entry shape) — the beat projection is a deliberate
    subset+2 (adds `is_impact_beat`, `created_at`), not the 9-key shape.
  - DEC-007 (required-flag validation in RunE) — `--audience` required;
    window flags reuse DEC-028's mutual-exclusion pattern.
  - DEC-025 (plugin packaging / bundled-asset convention) — the
    framing-directive assets follow the checked-in/bundled-asset pattern
    (`embed.FS`), the same posture as `BRAG.md` / `commands/brag.md`.
- Related constraints: `stdout-is-for-data-stderr-is-for-humans` (bundle
  to stdout, errors/directive-preview routing per SPEC-049),
  `no-sql-in-cli-layer` (threading + shaping live in
  `internal/story`/`internal/aggregate`; read path is
  `Store.List(ListFilter{Since})`), `test-before-implementation`,
  `no-cgo` (pure-Go asset embedding, no model).
- Related docs:
  - `projects/PROJ-004-story-surface/stages/STAGE-012-brag-story-audience.md`
  - `projects/PROJ-004-story-surface/brief.md`
  - `/guidance/questions.yaml` (two sub-choice questions filed).

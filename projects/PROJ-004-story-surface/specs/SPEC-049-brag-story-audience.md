---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-049
  type: story                      # epic | story | task | bug | chore
  cycle: verify
  blocked: false
  priority: high                   # the narrative headline of v0.4.0
  complexity: L                    # L — split recommended and taken: SPEC-049 = the profile mechanism + arc-aware bundle + framing-directive asset convention + the me/exec gradient endpoints; SPEC-050 = manager/skip (config-only) + polish. See "Complexity + split" below.

project:
  id: PROJ-004
  stage: STAGE-012
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-06

references:
  decisions:
    - DEC-029   # EMITTED by this spec — deterministic threads / the-LLM-finds-the-throughline; audiences = data-driven shaping profiles; arc-aware bundle
    - DEC-028   # REUSED — calendar windows (windowCutoff), project=initiative axis, WithImpact impact-first split, the narrow entry projection widened to a beat
    - DEC-014   # EXTENDED — provenance envelope (generated_at/scope/filters), 2-space JSON, empty-state rules; arc-aware body added
    - DEC-001   # PRESERVED — pure Go, local-first, no model/network/secrets in the binary; synthesis is a pipe
    - DEC-025   # REUSED — bundled-asset convention (embed.FS); framing directives ship as embedded assets like BRAG.md / commands/brag.md
    - DEC-008   # `--since` date format — reused verbatim via cli.ParseSince (through windowCutoff)
    - DEC-011   # JSON per-entry shape — the beat projection is a deliberate subset+2, not the 9-key shape
    - DEC-006   # cobra framework — new `brag story` subcommand
    - DEC-007   # required-flag validation in RunE — --audience required; window flags mutually exclusive
    - DEC-017   # entries↔project relationship — threads group on the `project` field (the initiative axis)
  constraints:
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
    - errors-wrap-with-context
    - test-before-implementation
    - one-spec-per-pr
    - no-cgo
  related_specs:
    - SPEC-048   # shipped; brag impact — the grouped/impact data brag story coalesces; windowCutoff, WithImpact, GroupEntriesByProject reuse
    - SPEC-018   # shipped; emitted DEC-014; echoFiltersForSummary precedent; maps-vs-arrays per-consumer payload divergence
    - SPEC-041   # shipped; the plugin bundled-asset convention (checked-in commands/brag.md, hooks) the directive assets mirror
    - SPEC-050   # planned; adds manager/skip profiles as data (extensibility proof, zero Go change)
---

# SPEC-049: `brag story --audience` — the extensible shaping-profile mechanism, the coalesce-into-arcs bundle, and the framing-directive assets

## Context

First (and headline) spec of STAGE-012, the narrative headline of
v0.4.0. Where `brag impact` (SPEC-048, STAGE-011) answers *"what did I
move the needle on this quarter, grouped by initiative?"* as a **report**,
`brag story` answers *"tell the story of my work, shaped for who's
listening"* as an **arc**: related brags become **beats in one thread**,
threads carry a **throughline skeleton**, and an audience sets **how many
arcs and at what altitude**. Related brags become beats in one arc, not
bullets in a list — that is the whole difference from `brag impact`.

**Posture (settled by orchestrator + user; NOT re-litigated here):**
synthesis is a **pure pipe** — bragfile owns *data + shaping*, the LLM
(already in the caller: an agent like Claude Code, or a paste-in session)
owns *words*. **No model, no network, no secrets in the binary** (DEC-001,
same posture as `brag review`/`summary`/`impact`). The shaped bundle is
**useful standalone** (readable/pasteable); the LLM is the optional upgrade
to polished narrative.

This spec:

1. **Adds `brag story --audience <name> [--quarter|--month|--year|--since
   <date>] [--theme <tag>] [--format markdown|json] [--print-directive]
   [--project|--type|--tag]`** — the narrative surface. `--audience` is
   **required**. The window flags reuse SPEC-048's calendar
   mutual-exclusion machinery; when no window flag is given, the
   **audience profile's default window** applies (`me` → year, `exec` →
   quarter). `--format` defaults to `markdown`.

2. **Introduces `internal/story/`** — the new package that owns the
   audience **shaping profiles** (data-driven, DEC-029 choice 2), the
   **threading/coalescing** (DEC-029 choice 1), the **arc-aware bundle
   renderer** (DEC-029 choice 5, markdown + JSON), and the embedded
   **framing-directive assets** (DEC-029 choice 7). Threading reuses
   `aggregate.GroupEntriesByProject` + `aggregate.WithImpact` verbatim;
   the window math reuses SPEC-048's `windowCutoff` (lifted to a shared
   helper — see Implementation Context / Rejected alternatives).

3. **Ships two audiences — the gradient endpoints `me` and `exec`** — as
   bundled default profiles + directives. `manager`/`skip` are deferred
   to SPEC-050 (config-only additions, the extensibility proof). This is
   the L→first-coherent-slice split (below).

4. **Emits DEC-029** — the thread definition + the data-driven
   shaping-profile mechanism + the arc-aware bundle shape. Two sub-choices
   land below 0.8 confidence and carry `/guidance/questions.yaml` entries
   (§14): the profile-override file format/precedence, and whether two
   audiences prove divergence.

DEC-014 is **extended, not relitigated**: the provenance envelope
(`generated_at`/`scope`/`filters`), 2-space JSON indent, and empty-state
rules are inherited. DEC-028 is **reused**: calendar windows, the
`project`=initiative axis, and the `WithImpact` impact-first split. Only
the **body** is new — arc-aware (threads → beats → throughline skeleton +
framing directive), which the flat `impact_by_project` array cannot carry.

Parent stage:
[`STAGE-012-brag-story-audience.md`](../stages/STAGE-012-brag-story-audience.md)
— Success Criteria (audience-shaped bundle that coalesces into arcs;
framing-directive asset; LLM optional; same corpus → different stories,
rule-driven) and the OPEN DESIGN FORK this spec resolves in DEC-029 (what
defines a thread). Project:
[`brief.md`](../brief.md) (the audience gradient + the pipe philosophy).

## Complexity + split (L → first coherent slice)

This stage's full ambition is **L**. Cramming it into one spec would
bundle: the profile mechanism, the arc-aware bundle (two formats), the
embedded directive-asset convention, four audiences, and a new command —
too much for one design→build→verify pass, and it would blur the
mechanism-vs-content boundary. Per AGENTS.md §12 (decide-at-design;
propose a split for L) the split is:

- **SPEC-049 (this spec):** the **mechanism + the gradient endpoints.**
  The `internal/story` package (profiles-as-data loader + bundled default
  set + user-override path, the threading/coalescing, the arc-aware
  bundle in markdown + JSON, the embedded framing-directive assets), the
  `brag story` command, and **two shipped audiences — `me` and `exec`**
  (the endpoints, which maximally exercise "diverge in practice").

- **SPEC-050 (planned, STAGE-012):** the **middle of the gradient** —
  `manager` (and optionally `skip`) shipped as **bundled default profiles
  + directive assets only, zero Go change**. This is the extensibility
  proof: if adding an audience touches Go, the mechanism failed. Plus any
  polish (per-profile fold thresholds, doc/tutorial pass).

**Orchestrator sign-off needed on the split** (recorded in DEC-029 choice
4 + the `story-two-audience-slice-proves-divergence` question): ship the
endpoints now, the middle in SPEC-050 — or pull `manager` into SPEC-049.
The design does not foreclose pulling it forward (it is additive config).

## Goal

Ship `brag story --audience me|exec [window] [--theme tag] [--format
markdown|json] [--print-directive] [filters]` as the first narrative
surface: coalesce the in-window corpus into **deterministic threads**
(initiative-grouped, time-ordered, with impact beats marked, plus an
optional `--theme` cross-cut), assemble a **throughline skeleton**, shape
the selection/threading/altitude per a **data-driven audience profile**,
and emit an **arc-aware bundle** (markdown default / JSON envelope) with a
per-audience **framing-directive** asset appended — a complete
paste-into-an-LLM artifact that is also a readable shaped digest
standalone. No model, no network in the binary.

## Inputs

- **Files to read:**
  - `/AGENTS.md` — §6 cycle rules; §9 testing conventions (golden style,
    monotonic-tiebreak, load-bearing-golden-first, injectable-clock seam);
    §12 design rules (decide-at-design-time, NOT-contains self-audit,
    literal-artifact-as-spec, flag-default explicitness, §12b design-time
    pre-flight of embedded literals); §14 confidence.
  - `/guidance/constraints.yaml` — the six referenced constraints.
  - `/decisions/DEC-029-story-audience-shaping-profiles-and-thread-definition.md`
    — the choices this spec implements (emitted by this spec).
  - `/decisions/DEC-028-impact-digest-window-and-shape.md` — the window
    machinery + `project`=initiative + `WithImpact` split reused.
  - `/decisions/DEC-014-rule-based-output-shape.md` — the envelope
    extended.
  - `/decisions/DEC-025-claude-code-plugin-packaging-and-capture-nudge.md`
    — the bundled-asset convention the directive assets mirror.
  - `internal/aggregate/aggregate.go` — `GroupEntriesByProject`,
    `WithImpact`, `NoProjectKey` (all reused verbatim).
  - `internal/export/impact.go` — the closest renderer sibling
    (`ImpactOptions`, `trimTrailingNewline`, the envelope-struct shape);
    `internal/export/summary.go` — provenance-block precedent.
  - `internal/cli/impact.go` — `windowCutoff`, `selectedWindow`,
    `echoFiltersForImpact`, the `nowFunc` clock seam; `brag story`'s
    `runStory` mirrors this plus the audience-profile resolution +
    default-window fallback.
  - `internal/storage/migrate.go` — the `//go:embed` pattern for bundling
    assets into the binary (the directive assets follow it).
  - `cmd/brag/main.go` — subcommand registration (mirror
    `NewImpactCmd()`).
- **Data read:** `Store.List(ListFilter{Since, Project, Type, Tag})` — the
  existing read path. Threading, shaping, and the throughline skeleton
  live ABOVE storage (`internal/story` + `internal/aggregate`); no SQL in
  the CLI or story layers.
- **No schema change.** Threads reuse the `project` field; impact beats
  reuse the `impact` field; the theme cross-cut reuses the normalized
  tags join via `ListFilter.Tag`. No migration.

## Outputs

- **New package `internal/story/`:**
  - `internal/story/profile.go` — the `Profile` struct (data-driven
    shaping profile: selection + threading + altitude + directive
    pointer), `LoadProfile(name)`, the bundled default set, and the
    user-override resolution.
  - `internal/story/profiles/me.yaml`, `internal/story/profiles/exec.yaml`
    — the two shipped default profiles (literal assets, §12
    literal-artifact-as-spec; embedded via `//go:embed`).
  - `internal/story/directives/me.md`, `internal/story/directives/exec.md`
    — the two shipped framing-directive assets (literal assets; embedded
    via `//go:embed`).
  - `internal/story/embed.go` — the `//go:embed profiles/*.yaml
    directives/*.md` FS + accessors.
  - `internal/story/thread.go` — `BuildThreads(entries, opts) []Thread`
    and `BuildThroughline(threads) Throughline` (the deterministic
    coalescing; reuses `aggregate.GroupEntriesByProject` + `WithImpact`).
  - `internal/story/bundle.go` — `ToStoryMarkdown` / `ToStoryJSON` (the
    arc-aware renderer + the `storyEnvelope` JSON structs).
  - `internal/story/*_test.go` — the failing tests below.
- **New file `internal/cli/story.go`** — `NewStoryCmd`, `runStory`, the
  audience-profile resolution + default-window fallback,
  `echoFiltersForStory`. Imports no `database/sql`.
- **New file `internal/cli/story_test.go`** — the CLI-layer failing tests
  below.
- **Edit `internal/cli/impact.go` + `internal/cli/story.go`** — lift
  `windowCutoff` + `selectedWindow` to a shared, unexported CLI helper
  (`window.go`) reused by both `impact` and `story` (this is the
  third-caller threshold — see Implementation Context). *(If build judges
  the lift risky for `one-spec-per-pr`, the fallback — a `story`-local
  copy — is a Rejected alternative below; the lift is preferred and its
  test coverage is Test 14.)*
- **Edit `cmd/brag/main.go`** — `root.AddCommand(cli.NewStoryCmd())`
  (mirror `NewImpactCmd()`).
- **New file `/decisions/DEC-029-...md`** — emitted by this spec (drafted
  this design cycle).
- **Edit `/guidance/questions.yaml`** — the two sub-0.8 questions
  (`story-profile-override-file-format`,
  `story-two-audience-slice-proves-divergence`; added this design cycle).
- **Edit `docs/api-contract.md`** — add a `brag story` section (audience,
  window flags, the arc-aware envelope, `--print-directive`, DEC-029
  cross-link). *(Build transcribes; mirror the `brag impact` section
  shape.)*
- **Edit `docs/tutorial.md` + `README.md`** *(new-command doc-references
  premise audit, §9)* — add `brag story` to any command-surface list. See
  Premise Audit.
- **Edit `BRAG.md`** *(status-change premise audit)* — the "Reading
  entries back" section enumerates read commands; `brag story` belongs in
  the narrative-read family. See Premise Audit for the exact grep.

### Premise audit note (planned, not build-time discovery)

This spec is **additive** (new command, new package, new DEC, new
embedded assets). No existing behavior is inverted or removed. Applicable
§9 additive cases:

- **New command → doc references update.** `brag story` is a new command;
  grep the docs + `BRAG.md` for the command-surface / read-command lists
  and add it. See Premise Audit for the executed greps + expected hits.
- **Addition to a tracked collection.** DEC count: prose in `docs/`, not
  a count-asserted collection. No literal-count assertion over
  `internal/export/*_test.go` goldens is touched (story goldens are
  self-contained, in the new `internal/story` package). The lifted
  `windowCutoff` MUST keep `impact`'s existing tests green — enumerated as
  a planned no-behavior-change refactor (Test 14 + the existing
  `TestWindowCutoff_*` / `TestImpactCmd_CalendarNotRolling` must still
  pass byte-for-byte).

## Acceptance Criteria

1. `brag story --audience me` prints a markdown arc-aware bundle for the
   `me` profile's default window (year): `# Bragfile Story`, the
   provenance block (`Generated:` / `Scope:` / `Audience: me` / `Filters:`
   / `Threads: <n>` / `Beats: <shown>/<in-window>`), then a `## Threads`
   section with one `### <thread>` per thread (initiative-grouped,
   time-ordered beats, impact beats marked), then `## Throughline
   (skeleton)` (the ordered thread refs + span + beat/impact-beat counts),
   then `## Framing directive` (the appended `me` directive text).
2. `brag story --audience exec` over the **same corpus + explicit window**
   produces a **demonstrably different** bundle than `me`: it surfaces
   only threads with ≥1 impact beat, folds/omits threads with no impact
   beat, DROPS impact-less beats from surfaced threads, and its throughline
   skeleton targets one headline arc (the highest-impact thread first).
   The difference is in the deterministic body, not only the directive.
3. `--audience` is **required**: omitting it is a `UserError` on stderr
   (non-zero exit, empty stdout). An unknown audience with no bundled
   default and no user-override file is a `UserError` naming the audience.
4. **Window resolution:** an explicit window flag (`--quarter`/`--month`/
   `--year`/`--since`) overrides the profile default; the window flags are
   **mutually exclusive** (reusing `impact`'s check); with no window flag,
   the **profile's default window** applies (`me` → `year`, `exec` →
   `quarter`), and `scope` echoes the resolved window token.
5. **Threads are deterministic:** initiative threads in alpha-ASC order
   (`(no project)` last, per `GroupEntriesByProject`); within a thread,
   beats ASC by `created_at` with ID tiebreak; a beat is marked
   `is_impact_beat` iff its `impact` is non-empty (matching
   `aggregate.WithImpact`). `--theme <tag>` adds exactly one extra
   cross-project thread (kind `theme`) after the initiative threads,
   grouping every in-window entry carrying that tag, time-ordered.
6. **`--format json`** emits the DEC-014 envelope extended with
   `audience`, `threads` (array of `{thread, kind, span:{first,last},
   beats:[{id, title, project, type, impact, is_impact_beat,
   created_at}]}`), `throughline` (`{arcs:[{thread, kind, beat_count,
   impact_beat_count, span:{first,last}}]}`), and `framing_directive`
   (the resolved directive text), 2-space indent. `audience` echoes the
   requested name; `scope` echoes the resolved window token.
7. **`--print-directive`** prints ONLY the resolved framing directive for
   the audience to stdout (no bundle body) and exits 0; it composes with
   `--audience` and requires no window/DB read.
8. **Standalone usefulness (DEC-029 choice 6):** with no LLM, the markdown
   bundle is a coherent shaped digest — threads, beats (impact beats
   flagged with a visible marker), the throughline skeleton, and the
   directive as a readable instruction block. The bundle body renders even
   when the directive is empty/missing (the `## Framing directive` section
   is omitted, everything else renders).
9. **Profiles are data-driven (DEC-029 choice 2):** `me`/`exec` load from
   the bundled `//go:embed`ded profile assets, NOT from a Go `switch`/enum.
   A user-override profile file at the resolved path shadows/overrides a
   bundled default by name; a malformed override file is a `UserError`
   naming the file (does NOT silently fall back). Loading is exercised
   without any Go type carrying a hard-coded audience list.
10. **Empty window:** provenance renders (`Threads: 0`,
    `Beats: 0/0`), the `## Threads` and `## Throughline` sections are
    OMITTED from markdown (DEC-014 empty-state), the `## Framing directive`
    section still renders (the directive is the audience's instructions,
    valid on an empty corpus — mirrors `brag review`'s always-render
    reflection block); JSON renders `threads` `[]`, `throughline.arcs`
    `[]`, `framing_directive` the directive string, `filters` `{}`.
11. `internal/cli/story.go` and `internal/story/*` import no
    `database/sql` / SQL driver (`no-sql-in-cli-layer`); the bundle body
    goes to stdout, all errors + the directive-preview routing honor
    `stdout-is-for-data-stderr-is-for-humans`; errors wrap with context;
    the build stays pure-Go (`no-cgo` — assets embed via `embed.FS`, no
    model, no network).

## Failing Tests

Written during design (this spec), made to pass during build. Load-bearing
goldens are written FIRST (§9). All fixtures use an injected `Now` so the
window math + `Generated:` line are deterministic. Fixture `Now` is
`2026-07-06T12:00:00Z` (Q3 2026 → calendar quarter start `2026-07-01`;
calendar year start `2026-01-01`).

### Shared threading fixture (`internal/story/thread_test.go`)

```go
// storyFixture: 6 entries across the 2026 calendar year. Two
// initiatives (alpha, beta), plus one (no project) entry. Impact
// beats and impact-less beats are mixed so the me-vs-exec divergence
// is exercised: alpha has 2 beats (1 with impact, 1 without); beta has
// 2 beats (both with impact); (no project) has 1 beat (no impact);
// a 6th entry carries the theme tag `perf` and impact, for the
// --theme cross-cut. Non-monotonic id/time pairing inside alpha
// exercises the ASC + ID-tiebreak path.
var storyFixture = []storage.Entry{
    {ID: 1, Title: "alpha-early", Project: "alpha", Type: "shipped",
        Tags: "perf",
        Impact:    "cut p95 login latency 40%",
        CreatedAt: time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)},
    {ID: 2, Title: "alpha-messy", Project: "alpha", Type: "learned",
        Impact:    "", // impact-less: KEPT by me, DROPPED by exec
        CreatedAt: time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)},
    {ID: 3, Title: "beta-one", Project: "beta", Type: "shipped",
        Impact:    "onboarding time down to 1 day",
        CreatedAt: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)},
    {ID: 4, Title: "beta-two", Project: "beta", Type: "shipped",
        Impact:    "removed the nightly cron entirely",
        CreatedAt: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)},
    {ID: 5, Title: "loose-note", Type: "fixed",
        Impact:    "", // (no project), impact-less
        CreatedAt: time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)},
    {ID: 6, Title: "perf-sweep", Project: "gamma", Type: "shipped",
        Tags:      "perf",
        Impact:    "shaved 200ms off cold start",
        CreatedAt: time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)},
}

var storyFixedNow = time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
```

> Note: the renderer receives the already-in-window slice + the resolved
> `Profile` + the raw in-window count. `storyFixture` stands in for
> "everything in the window"; `Beats in-window` = 6.

#### Test 1 — `TestToStoryMarkdown_MeProfile_FullDocumentGolden` (LOAD-BEARING — write FIRST)

Byte-exact assertion. `me` profile (every thread; impact-less beats KEPT;
low altitude), `--year` window, no theme. Expected document (the `★`
marker flags an impact beat; a plain `·` marks a non-impact beat — the
visible, standalone "so what" signal):

```
# Bragfile Story

Generated: 2026-07-06T12:00:00Z
Scope: year
Audience: me
Filters: (none)
Threads: 4
Beats: 6/6

## Threads

### alpha

- ★ 1: alpha-early
  cut p95 login latency 40%
- · 2: alpha-messy

### beta

- ★ 3: beta-one
  onboarding time down to 1 day
- ★ 4: beta-two
  removed the nightly cron entirely

### gamma

- ★ 6: perf-sweep
  shaved 200ms off cold start

### (no project)

- · 5: loose-note

## Throughline (skeleton)

- alpha [initiative]: 2 beats, 1 with impact (2026-02-01 → 2026-04-01)
- beta [initiative]: 2 beats, 2 with impact (2026-03-01 → 2026-05-01)
- gamma [initiative]: 1 beat, 1 with impact (2026-06-15 → 2026-06-15)
- (no project) [initiative]: 1 beat, 0 with impact (2026-06-01 → 2026-06-01)

## Framing directive

<the full literal text of internal/story/directives/me.md, verbatim>
```

Locks (me): every thread surfaces — **all four initiatives**: alpha, beta,
gamma (a valid initiative thread — entry 6 `perf-sweep` has a project and
an impact beat), and `(no project)` last. `Threads: 4`; `Beats: 6/6` (me
shows all six in-window beats — impact-less beats KEPT). Impact-less beats
render `· <id>: <title>` with no impact line (`· 2: alpha-messy`,
`· 5: loose-note`); impact beats render `★ <id>: <title>` + the indented
impact line in full. Thread order is `GroupEntriesByProject`'s alpha-ASC
with `(no project)` last (alpha, beta, gamma, then (no project)); within
alpha id 1 (Feb) before id 2 (Apr). The throughline skeleton lists all
four threads with counts + span. The directive is appended verbatim; no
trailing newline. **gamma is present without `--theme`** because a project
value makes it an initiative thread — `--theme` adds a *cross-project*
cross-cut (Test 7), it is not what surfaces a project's own thread. This
golden's thread set is `GroupEntriesByProject(storyFixture)` under the
`me` keep-all policy: {alpha, beta, gamma, (no project)} — and it agrees
with Test 2's exec thread set (beta, alpha, gamma — same four minus the
folded impact-less (no project)) and Test 3's JSON (same four threads).

#### Test 2 — `TestToStoryMarkdown_ExecProfile_FullDocumentGolden` (LOAD-BEARING — write SECOND)

Byte-exact assertion. `exec` profile (impact-bearing threads only;
impact-less beats DROPPED; small threads folded; one-headline-arc
altitude: threads ordered by impact-beat count DESC, then alpha-ASC),
same corpus, `--quarter` explicit window → but the fixture's Q3 window
(from 2026-07-01) contains ZERO entries, so to exercise exec meaningfully
the test passes `--year`. Expected document:

```
# Bragfile Story

Generated: 2026-07-06T12:00:00Z
Scope: year
Audience: exec
Filters: (none)
Threads: 3
Beats: 4/6

## Threads

### beta

- ★ 3: beta-one
  onboarding time down to 1 day
- ★ 4: beta-two
  removed the nightly cron entirely

### alpha

- ★ 1: alpha-early
  cut p95 login latency 40%

### gamma

- ★ 6: perf-sweep
  shaved 200ms off cold start

## Throughline (skeleton)

- beta [initiative]: 2 beats, 2 with impact (2026-03-01 → 2026-05-01)
- alpha [initiative]: 1 beat, 1 with impact (2026-02-01 → 2026-02-01)
- gamma [initiative]: 1 beat, 1 with impact (2026-06-15 → 2026-06-15)

## Framing directive

<the full literal text of internal/story/directives/exec.md, verbatim>
```

Locks (exec, contrasted against Test 1 on the SAME corpus): `(no project)`
thread OMITTED (its only beat is impact-less → thread has 0 impact beats →
folded); `alpha-messy` (id 2, impact-less) DROPPED even though alpha
survives (alpha has 1 impact beat); threads ordered by impact-beat count
DESC (beta=2 first, then alpha=1 and gamma=1 tie broken alpha-ASC); the
`Beats: 4/6` tally (4 impact beats shown of 6 in-window); the exec
directive appended. This golden vs Test 1's golden IS the "same corpus,
different story, rule-driven not tone" proof (AC-2): exec = me's four
threads minus the folded impact-less `(no project)` thread, minus dropped
impact-less beats, reordered impact-desc. The exec thread set + order was
computed from the fixture + the exec profile's threading policy and
reconciled against Test 1/3's `me` set at design (§12b below).

#### Test 3 — `TestToStoryJSON_MeProfile_ShapeGolden` (LOAD-BEARING — write THIRD)

Byte-exact JSON assertion on the `me` fixture/window. Expected (abbreviated
skeleton the build fills to byte-exactness from the fixture; the KEY SHAPE
is locked here):

```json
{
  "generated_at": "2026-07-06T12:00:00Z",
  "scope": "year",
  "audience": "me",
  "filters": {},
  "threads": [
    {
      "thread": "alpha",
      "kind": "initiative",
      "span": { "first": "2026-02-01T10:00:00Z", "last": "2026-04-01T10:00:00Z" },
      "beats": [
        { "id": 1, "title": "alpha-early", "project": "alpha", "type": "shipped",
          "impact": "cut p95 login latency 40%", "is_impact_beat": true,
          "created_at": "2026-02-01T10:00:00Z" },
        { "id": 2, "title": "alpha-messy", "project": "alpha", "type": "learned",
          "impact": "", "is_impact_beat": false,
          "created_at": "2026-04-01T10:00:00Z" }
      ]
    }
  ],
  "throughline": {
    "arcs": [
      { "thread": "alpha", "kind": "initiative", "beat_count": 2,
        "impact_beat_count": 1,
        "span": { "first": "2026-02-01T10:00:00Z", "last": "2026-04-01T10:00:00Z" } }
    ]
  },
  "framing_directive": "<the full me.md text as a JSON string>"
}
```

> Only the `alpha` thread + its arc are shown in full above to keep the
> literal readable; the `me` bundle carries **all four** threads. Build
> writes the complete byte-exact golden: `threads` = `[alpha, beta,
> gamma, (no project)]` (each a full `threadJSON` with its beats), and
> `throughline.arcs` = the matching four arcs. The elided threads' shapes
> follow the `alpha` shape exactly:
> - `beta`: beats `[{id:3, is_impact_beat:true}, {id:4, is_impact_beat:true}]`, span `2026-03-01…2026-05-01`, arc `beat_count 2, impact_beat_count 2`.
> - `gamma`: beats `[{id:6, title:"perf-sweep", project:"gamma", type:"shipped", impact:"shaved 200ms off cold start", is_impact_beat:true, created_at:"2026-06-15T10:00:00Z"}]`, span `2026-06-15…2026-06-15`, arc `beat_count 1, impact_beat_count 1`.
> - `(no project)`: beats `[{id:5, title:"loose-note", project:"(no project)", type:"fixed", impact:"", is_impact_beat:false, created_at:"2026-06-01T10:00:00Z"}]`, span `2026-06-01…2026-06-01`, arc `beat_count 1, impact_beat_count 0`. (The `(no project)` group key is `aggregate.NoProjectKey`; a beat under it carries `"project": "(no project)"`.)

Locks: top-level key order (`generated_at`, `scope`, `audience`,
`filters`, `threads`, `throughline`, `framing_directive`); the 7-key beat
projection (`id, title, project, type, impact, is_impact_beat,
created_at` — a deliberate subset+2 of DEC-011, not the 9-key shape); the
`span` object shape; `throughline.arcs` with counts; `framing_directive`
as a JSON string; 2-space indent; `filters` `{}` when absent. The thread
set is the SAME four threads as Test 1's markdown golden (`me` keep-all
policy over `GroupEntriesByProject(storyFixture)`): alpha, beta, gamma,
(no project) — the two goldens agree by construction.

#### Test 4 — `TestToStory_EmptyWindow`

Input: `nil` entries, `me` profile, `Now` = `storyFixedNow`,
`EntriesInWindow: 0`. Markdown = header + provenance
(`Threads: 0`, `Beats: 0/0`) + the `## Framing directive` section only
(no `## Threads`, no `## Throughline`). JSON: `threads` `[]`,
`throughline.arcs` `[]` (non-nil), `framing_directive` = the me directive
string, `filters` `{}`. (DEC-014 empty-state + DEC-029 choice 8: directive
always renders.)

#### Test 5 — `TestBuildThreads_DeterministicOrderAndImpactMarking`

`BuildThreads(storyFixture, meThreadOpts)` (me policy: keep all threads,
keep impact-less beats) returns threads in order `alpha, beta, gamma, (no
project)` (alpha-ASC, `(no project)` last); alpha's beats are `[id1, id2]`
(ASC + ID tiebreak); every beat's `IsImpactBeat` equals
`entry.Impact != ""` (ids 1,3,4,6 true; ids 2,5 false). Confirms the
deterministic coalescing matches `GroupEntriesByProject` + `WithImpact`.

#### Test 6 — `TestBuildThreads_ExecPolicyFoldsAndDrops`

`BuildThreads(storyFixture, execThreadOpts)` (exec policy: impact-bearing
threads only, drop impact-less beats, order by impact-beat-count DESC then
alpha-ASC) returns threads `[beta, alpha, gamma]` — `(no project)` FOLDED
(0 impact beats); alpha's beats = `[id1]` only (`alpha-messy` dropped);
beta before alpha before gamma (2 > 1, tie alpha-ASC). This is the
mechanism half of the me-vs-exec divergence (the render half is Test 2).

#### Test 7 — `TestBuildThreads_ThemeCrossCut`

`BuildThreads(storyFixture, meThreadOptsWithTheme("perf"))` appends
exactly ONE extra thread after the initiative threads: `{thread: "perf",
kind: "theme", beats: [id1, id6]}` (the two `perf`-tagged entries,
time-ordered: id1 Feb, id6 Jun), leaving the initiative threads unchanged.
Confirms `--theme` is an opt-in cross-project cross-cut (DEC-029 choice 1),
placed after initiatives, kind `theme`.

#### Test 8 — `TestBuildThroughline_SkeletonCountsAndSpan`

`BuildThroughline(threads)` produces one arc per thread carrying
`BeatCount`, `ImpactBeatCount`, and `Span{First, Last}` computed from the
thread's beats. For me's alpha thread: `BeatCount 2, ImpactBeatCount 1,
Span{2026-02-01, 2026-04-01}`. Empty threads → non-nil empty `arcs`.

### Profile mechanism (`internal/story/profile_test.go`)

#### Test 9 — `TestLoadProfile_BundledDefaults`

`LoadProfile("me")` and `LoadProfile("exec")` load from the embedded
assets and return profiles whose fields match the shipped `me.yaml`/
`exec.yaml` (default window `year`/`quarter`; me keeps-all-threads +
keeps-impact-less-beats + low altitude; exec impact-threads-only +
drops-impact-less + folds-small + one-arc altitude; each points at its
directive asset). **No Go type enumerates the audience names** — the set
comes from the embedded FS + the override dir (asserted by loading a name
present only as an asset, e.g. by listing the embed FS).

#### Test 10 — `TestLoadProfile_UserOverrideShadowsBundled`

With a `t.TempDir()` override dir seeded with a `me.yaml` that changes the
default window to `quarter`, `LoadProfile("me")` (pointed at that dir via
the injectable path seam) returns the OVERRIDE (window `quarter`), not the
bundled default. Confirms override-by-name precedence (DEC-029 choice 2 /
the `story-profile-override-file-format` question's locked answer:
override-wins).

#### Test 11 — `TestLoadProfile_UnknownAndMalformed`

`LoadProfile("nope")` with no bundled default and no override file →
`ErrProfileNotFound` (surfaced as a `UserError` at the CLI). A malformed
override file (invalid YAML) for an otherwise-known name → an error naming
the file (does NOT silently fall back to the bundled default — DEC-029
choice 9 / AC-9). *(NOT-contains self-audit, §12: `nope` appears only in
this test.)*

#### Test 12 — `TestDirectiveAsset_ResolvesAndIsNonEmpty`

The resolved directive for `me` and for `exec` is the exact byte content
of the embedded `directives/me.md` / `directives/exec.md`, non-empty, and
differs between the two audiences (the directives encode different
altitude/candor instructions — the me directive mentions the messy middle
/ lessons; the exec directive mentions business impact / one headline).
Asserts the two directives are NOT byte-identical.

### CLI layer (`internal/cli/story_test.go`)

CLI tests build a `*cobra.Command` via `NewStoryCmd()` with separate
`outBuf`/`errBuf` (§9), a `t.TempDir()` DB seeded through `storage`, and
the injectable `nowFunc` clock seam.

#### Test 13 — `TestStoryCmd_RequiresAudience`

- No `--audience` → `UserError`, `errBuf`/returned-error mentions
  `--audience` required, `outBuf` empty.
- `--audience nope` (no bundled default, no override) → `UserError` naming
  the audience, `outBuf` empty.

#### Test 14 — `TestStoryCmd_WindowResolutionAndSharedHelper` (LOAD-BEARING for the refactor)

- `--audience me` with NO window flag → resolves to the profile default
  (`year`); `scope` echoes `year`.
- `--audience me --quarter` → explicit window overrides the default;
  `scope` echoes `quarter`.
- `--audience me --quarter --month` → `UserError` (mutual exclusion,
  reusing `impact`'s `selectedWindow`), `outBuf` empty.
- The lifted `windowCutoff`/`selectedWindow` still satisfy `impact`'s
  existing behavior: this test package + the existing `impact` tests both
  pass (asserted by the full suite staying green — enumerated in Premise
  Audit). Confirms the shared helper (Rejected-alternative below: a
  story-local copy).

#### Test 15 — `TestStoryCmd_MeVsExecDivergenceLive` (LOAD-BEARING — the headline assertion)

Seed the `storyFixture` corpus into a `t.TempDir()` DB (via a `created_at`
rewrite seam mirroring `impact`'s `seedImpactEntry`). Run `brag story
--audience me --year` and `brag story --audience exec --year` over the
SAME DB. Assert:
- me output CONTAINS `alpha-messy` and `loose-note` (impact-less beats
  kept) and the `(no project)` thread heading;
- exec output does NOT contain `alpha-messy`, does NOT contain
  `loose-note`, does NOT contain `(no project)` (dropped/folded);
- exec's first thread heading is `### beta` (impact-beat-count DESC);
- both contain their own audience's directive text (me's differs from
  exec's).
This is the live proof of AC-2 (same corpus → different story, rule-driven).
*(NOT-contains self-audit, §12: `alpha-messy`/`loose-note` appear only in
the fixture + this test, never in any `Long` string, profile asset,
directive asset, or rendered literal — confirmed at design; see the
self-audit note in Implementation Context.)*

#### Test 16 — `TestStoryCmd_PrintDirectiveOnly`

`brag story --audience exec --print-directive` writes ONLY the exec
directive text to `outBuf` (no `# Bragfile Story` header, no `## Threads`),
exits 0, does NOT open the DB (works against a non-existent `--db` path).
`--print-directive` with an unknown audience → `UserError`.

#### Test 17 — `TestStoryCmd_FormatDefaultAndUnknown`

No `--format` → markdown (assert `# Bragfile Story` in `outBuf`).
`--format json` → `outBuf` parses as JSON with an `audience` key.
`--format xml` → `UserError`, `outBuf` empty. *(§12 flag-default:
`--format` default is `"markdown"`, stated in the literal `Long` + the
flag registration.)*

#### Test 18 — `TestStoryCmd_StdoutStderrSeparation`

A successful `--audience me --year` run writes the bundle to `outBuf`
only; `errBuf.Len() == 0`. A `UserError` run (missing `--audience`) writes
nothing to `outBuf`; the error is a `UserError` main.go routes to stderr
(assert `errors.Is(err, ErrUser)` + empty stdout, mirroring
`impact_test.go`'s established pattern under `SilenceErrors: true`).

## Implementation Context

Everything build needs without re-discovering it.

### The read path

`brag story` reads via `Store.List(ListFilter{Since, Project, Type, Tag})`
— identical to `brag impact`. The window becomes `ListFilter.Since`;
`--theme` does NOT go into `ListFilter` (the cross-cut is computed in
`internal/story` over the already-fetched in-window set, so a single
`List` call serves both the initiative threads and the theme cross-cut).
Filter flags map to `ListFilter.Project/Type/Tag`. No `Until` bound (the
period end is always "now"; DEC-028's reasoning applies verbatim).

### The shared window helper (lift, third-caller threshold)

`windowCutoff` + `selectedWindow` currently live in `internal/cli/impact.go`
(SPEC-048). `brag story` is the THIRD caller of window logic (impact, story,
and — conceptually — any future windowed digest), which crosses the
established lift threshold (SPEC-018 set "third caller → lift"). Move both
functions verbatim into a new `internal/cli/window.go` (same package,
unexported), and have `impact.go` + `story.go` both call them. **This is a
behavior-preserving refactor**: `impact`'s existing tests
(`TestWindowCutoff_CalendarArithmetic`, `TestImpactCmd_CalendarNotRolling`,
etc.) must pass unchanged. Story's window resolution wraps them:

```
resolveWindow(cmd, profile, now):
  if any of --quarter/--month/--year/--since changed:
      window, err = selectedWindow(cmd)   // reuse; mutual-exclusion + the pair
      ... windowCutoff(window, sinceRaw, now)
  else:
      window = profile.DefaultWindow      // "year" | "quarter" | ...
      ... windowCutoff(window, "", now)   // sinceRaw unused for non-since defaults
```

`selectedWindow` today returns a `UserError` when ZERO window flags are set
(required). For `story`, zero-window is NOT an error — it means "use the
profile default." So `runStory` checks "did the user set any window flag?"
BEFORE calling `selectedWindow`, and only calls `selectedWindow` (which
enforces exactly-one) when at least one is set. This keeps `selectedWindow`
unchanged (still "exactly one among those set") and lets `story` supply the
default. **Do NOT change `selectedWindow`'s signature or its zero-flag
error** — `impact` depends on it. (If build finds `selectedWindow`'s
zero-flag `UserError` awkward to reuse, the clean alternative is a small
`windowFlagsSet(cmd) bool` helper in `window.go` gating the call — that is
a build-time helper, not a decision; add it.)

### The injectable clock + profile-path seams (§9)

Mirror `impact`'s `nowFunc`. Add `var nowFunc = func() time.Time { return
time.Now().UTC() }` referenced once in `runStory` and threaded into
`resolveWindow` + the bundle options (single instant). For profile
override resolution, reference the override directory through a package var
in `internal/story` (e.g. `var overrideDir = defaultOverrideDir` where
`defaultOverrideDir` resolves `~/.bragfile/story-profiles` or the config
dir) so Test 10 can substitute a `t.TempDir()`.

### The profile mechanism (`internal/story/profile.go`) — LOCKED shape

A profile is DATA. The Go struct is the parsed shape; the source of truth
is the YAML asset. Locked struct + YAML schema:

```go
type Profile struct {
    Name          string   // "me" | "exec" | <user name>
    DefaultWindow string   // "year" | "quarter" | "month" | "since:<raw>"
    // Selection
    ImpactThreadsOnly bool  // exec: true (only threads with >=1 impact beat)
    DropImpactlessBeats bool // exec: true (impact-less beats removed from surfaced threads)
    // Threading / altitude
    FoldSmallThreads bool   // exec: true (a thread with 0 impact beats is omitted)
    ThreadOrder     string  // "initiative" (alpha-ASC) | "impact-desc" (impact-beat-count DESC, alpha-ASC tiebreak)
    Candor          string  // "candid" | "promotional" (metadata surfaced to the LLM, not a body rule)
    Directive       string  // asset basename: "me.md" | "exec.md" | <user path>
}
```

```yaml
# internal/story/profiles/me.yaml  (LITERAL — build transcribes verbatim)
name: me
default_window: year
impact_threads_only: false
drop_impactless_beats: false
fold_small_threads: false
thread_order: initiative
candor: candid
directive: me.md
```

```yaml
# internal/story/profiles/exec.yaml  (LITERAL — build transcribes verbatim)
name: exec
default_window: quarter
impact_threads_only: true
drop_impactless_beats: true
fold_small_threads: true
thread_order: impact-desc
candor: promotional
directive: exec.md
```

`LoadProfile(name)`:
1. If `<overrideDir>/<name>.yaml` exists → parse it (malformed → error
   naming the file; NO fallback).
2. Else if `profiles/<name>.yaml` exists in the embedded FS → parse it.
3. Else → `ErrProfileNotFound`.
Override-wins by name (DEC-029 choice 2). YAML parsing: the repo has no
YAML dep in `go.mod`; **check `go.mod` first** — if `gopkg.in/yaml.v3` is
not already an indirect dep, adding it triggers
`no-new-top-level-deps-without-decision` (warning). **Preferred to avoid
the dep:** the profile schema is a flat `key: value` file with no nesting,
lists, or quoting edge cases — a tiny hand-rolled `key: value` line parser
in `profile.go` (≈30 lines, pure stdlib `bufio`/`strings`) suffices and
keeps the dependency surface clean. Build: use the hand-rolled parser
unless a nesting need appears (none here). *(This is a locked decision —
see Locked design decisions #7 + Rejected alternatives.)*

### The threading (`internal/story/thread.go`) — reuses aggregate

```go
type Beat struct {
    ID           int64
    Title        string
    Project      string
    Type         string
    Impact       string
    IsImpactBeat bool      // == (Impact != "")
    CreatedAt    time.Time
}
type Thread struct {
    Thread string    // project name, NoProjectKey, or the theme tag
    Kind   string    // "initiative" | "theme"
    Beats  []Beat
}
type ThreadOptions struct {
    Order              string // profile.ThreadOrder
    ImpactThreadsOnly  bool
    DropImpactlessBeats bool
    FoldSmallThreads   bool
    Theme              string // "" or the --theme tag
}
```

`BuildThreads(entries, opts)`:
1. Initiative threads: `aggregate.GroupEntriesByProject(entries)` gives
   alpha-ASC groups, `(no project)` last, beats ASC+ID-tiebreak already.
   Map each group → `Thread{Kind:"initiative"}`; each entry → `Beat` with
   `IsImpactBeat = e.Impact != ""`.
2. If `opts.DropImpactlessBeats`: drop non-impact beats from each thread.
3. If `opts.ImpactThreadsOnly` / `FoldSmallThreads`: drop threads with 0
   impact beats. (These two co-vary for exec; keep both flags for future
   profiles that might diverge.)
4. If `opts.Order == "impact-desc"`: stable-sort threads by
   impact-beat-count DESC, then existing alpha-ASC order as tiebreak.
   (`(no project)` only appears here if it survived step 3, which it won't
   for exec — but keep the `(no project)`-last rule when `Order ==
   "initiative"`.)
5. If `opts.Theme != ""`: append ONE `Thread{Kind:"theme", Thread:theme}`
   whose beats are every entry in `entries` whose `Tags` contains the theme
   token (reuse the same tag-membership test `ListFilter.Tag` uses — exact
   token match over the comma-joined tags, DEC-004/DEC-015), time-ordered
   ASC+ID-tiebreak. Placed AFTER the initiative threads regardless of
   order. (The theme cross-cut is not subject to fold/drop — it is an
   explicit opt-in.)

`BuildThroughline(threads)` → `Throughline{Arcs: [...]}` one arc per
thread with `BeatCount`, `ImpactBeatCount`, `Span{First,Last}` (first/last
beat `CreatedAt`). Empty → non-nil empty `Arcs`.

### The bundle renderer (`internal/story/bundle.go`) — mirrors impact.go

`StoryOptions{ Audience, Scope, Filters, FiltersJSON, EntriesInWindow,
Now, Threads []Thread, Throughline, Directive string }`. The CLI builds
the threads + throughline + resolves the directive, passes them in (the
renderer is pure, like `ToImpactMarkdown`). Markdown per Test 1/2 golden;
the impact-beat marker is `★` (U+2605), the non-impact marker is `·`
(U+00B7) — both locked in the goldens. `## Framing directive` renders the
directive verbatim; if `Directive == ""` the section is omitted (AC-8).
Empty threads → omit `## Threads` + `## Throughline`, keep provenance +
directive (Test 4).

JSON envelope (field order = key order, DEC-029 choice 5):

```go
type storyEnvelope struct {
    GeneratedAt      string            `json:"generated_at"`
    Scope            string            `json:"scope"`
    Audience         string            `json:"audience"`
    Filters          map[string]string `json:"filters"`
    Threads          []threadJSON      `json:"threads"`
    Throughline      throughlineJSON   `json:"throughline"`
    FramingDirective string            `json:"framing_directive"`
}
type threadJSON struct {
    Thread string   `json:"thread"`
    Kind   string   `json:"kind"`
    Span   spanJSON `json:"span"`
    Beats  []beatJSON `json:"beats"`
}
type beatJSON struct {
    ID           int64  `json:"id"`
    Title        string `json:"title"`
    Project      string `json:"project"`
    Type         string `json:"type"`
    Impact       string `json:"impact"`
    IsImpactBeat bool   `json:"is_impact_beat"`
    CreatedAt    string `json:"created_at"`  // RFC3339
}
type throughlineJSON struct {
    Arcs []arcJSON `json:"arcs"`
}
type arcJSON struct {
    Thread          string   `json:"thread"`
    Kind            string   `json:"kind"`
    BeatCount       int      `json:"beat_count"`
    ImpactBeatCount int      `json:"impact_beat_count"`
    Span            spanJSON `json:"span"`
}
type spanJSON struct {
    First string `json:"first"`  // RFC3339
    Last  string `json:"last"`
}
```

Init `Threads`/`Arcs` to non-nil empty slices; `Filters` nil → `{}`.
Marshal with `json.MarshalIndent(env, "", "  ")`.

### The embedded assets (`internal/story/embed.go`)

```go
//go:embed profiles/*.yaml directives/*.md
var assetsFS embed.FS
```

Mirror `internal/storage/migrate.go`'s `//go:embed migrations/*.sql`
pattern. Accessors: `bundledProfile(name) ([]byte, bool)`,
`directiveAsset(basename) ([]byte, error)`.

### The framing-directive assets (LITERAL — §12 literal-artifact-as-spec)

These are DATA. Build transcribes verbatim; verify diffs. They are the
LLM's per-audience instructions for weaving the arc. Kept concise (the
bundle carries the deterministic scaffold; the directive supplies
altitude + candor + the "find the throughline" ask).

`internal/story/directives/me.md` (LITERAL):

```
# Framing directive — audience: me (reflect, candid)

You are helping the author reflect on their own work. Weave the threads
below into an honest first-person narrative — not a highlight reel.

- Include the messy middle: struggles, false starts, and lessons are the
  point of a reflection. Beats marked · (no recorded impact) matter here.
- Surface every thread. Find the throughline that connects them — the
  skill or theme that grew across initiatives — but do not force one arc;
  parallel arcs are fine for reflection.
- Name what was learned, not just what shipped. Ask what the author would
  do differently.
- Tone: candid, specific, low-altitude. Detail over polish.
```

`internal/story/directives/exec.md` (LITERAL):

```
# Framing directive — audience: exec (promote, impact-forward)

You are drafting an executive-facing summary. Weave the threads below into
ONE headline arc, impact-first.

- Lead with business impact. Every beat below carries a ★ impact
  statement; build the narrative from those outcomes, not the activity.
- Collapse to a single throughline: what shifted for the business this
  period, and why it mattered. Prefer one strong arc over many.
- Terse and promotional. One or two sentences per outcome. No process,
  no messy middle — the highest-impact thread leads.
- Quantify wherever the impact beats give you a metric.
```

### The cobra command (literal `Long`, §12 literal-artifact-as-spec)

```
Long: `Emit an audience-shaped narrative bundle: your brags coalesced into threads (initiatives, time-ordered, with impact beats marked) plus a throughline skeleton and a per-audience framing directive. bragfile shapes the data; an LLM (already in your session, or a paste-in) writes the prose. No model, no network in the binary — the bundle is a complete, readable artifact on its own; the LLM is an optional upgrade.

--audience is required and selects a shaping profile (selection + threading + altitude + framing directive):
  me     candid reflection — every thread, the messy middle and lessons kept, low altitude
  exec   impact-forward promotion — impact-bearing threads only, one headline arc, terse

Each audience carries a default window; an explicit window flag overrides it. Windows are CALENDAR periods (like brag impact), mutually exclusive:
  --quarter / --month / --year / --since D   (D: YYYY-MM-DD or Nd/Nw/Nm)

Audiences are extensible profiles, not a fixed list: drop a <name>.yaml in your story-profiles override directory to add or reshape one.

Output is markdown (default) or a JSON envelope (--format json). --theme <tag> adds a cross-project thread for that tag. --print-directive prints only the audience's framing directive.

Examples:
  brag story --audience me                                   # candid, this year (me's default window)
  brag story --audience exec --quarter                       # exec, this calendar quarter
  brag story --audience exec --year --format json            # arc-aware JSON envelope
  brag story --audience me --theme perf                      # add a cross-project perf arc
  brag story --audience exec --print-directive               # just the framing directive`,
```

Flags (with explicit defaults, §12 flag-default rule):

```go
cmd.Flags().String("audience", "", "shaping profile (required; one of: me, exec, or a user profile)")
cmd.Flags().Bool("quarter", false, "window: the current calendar quarter (overrides the profile default)")
cmd.Flags().Bool("month", false, "window: the current calendar month (overrides the profile default)")
cmd.Flags().Bool("year", false, "window: the current calendar year (overrides the profile default)")
cmd.Flags().String("since", "", "window: entries since a date (YYYY-MM-DD or Nd/Nw/Nm)")
cmd.Flags().String("theme", "", "add a cross-project thread grouping entries with this tag")
cmd.Flags().String("format", "markdown", "output format (one of: markdown, json)")
cmd.Flags().Bool("print-directive", false, "print only the audience's framing directive and exit")
cmd.Flags().String("project", "", "filter to entries with this project (exact match)")
cmd.Flags().String("type", "", "filter to entries with this type (exact match)")
cmd.Flags().String("tag", "", "filter to entries whose tags contain this token")
```

`--audience` required: in `runStory`, `if !cmd.Flags().Changed("audience")`
→ `UserErrorf("--audience is required (one of: me, exec, or a user profile)")`.

### NOT-contains self-audit (§12), run at design

Tests 11/15 assert output does NOT contain certain tokens. Grepping the
load-bearing prose (the `Long` string, the two profile YAML assets, the
two directive `.md` assets, the markdown goldens, the flag help) for each:
- `alpha-messy`, `loose-note` (Test 15 exec NOT-contains): appear ONLY in
  `storyFixture` + Test 15 + the me golden (Test 1, where they SHOULD
  appear). Zero hits in the exec golden (Test 2), the `Long`, the
  profiles, or the directives. Clean.
- `(no project)` (Test 15 exec NOT-contains): appears in the me golden
  (Test 1) + `aggregate.NoProjectKey`; NOT in the exec golden or `Long`.
  The exec NOT-contains holds because exec folds the `(no project)` thread.
  Clean (the token is legitimately in `me`'s output; the assertion is
  scoped to exec's output).
- `nope` (Test 11): only in Test 11. Clean.

### §12b design-time pre-flight (embedded-literal + expected-value)

The goldens above are **correct at design time** (§9: a write-first
load-bearing golden must be right, not aspirational). Their thread sets +
order were computed from `GroupEntriesByProject(storyFixture)` under each
profile's policy and reconciled against each other before locking — build
makes the code produce these goldens, it does not "fix" them:

1. **`me` (Order=initiative, keep-all):** threads =
   {alpha, beta, gamma, (no project)} in alpha-ASC with `(no project)`
   last → **alpha, beta, gamma, (no project)** (4 threads, all 6 beats
   shown, `Beats: 6/6`). gamma is present because entry 6 (`perf-sweep`)
   has a `project` value, so it is a valid initiative thread; `--theme`
   is NOT what surfaces it (that adds a *cross-project* cross-cut, Test 7).
2. **`exec` (Order=impact-desc, impact-threads-only, drop-impactless):**
   threads = {beta(2 impact beats), alpha(1), gamma(1)} → beta first, then
   alpha & gamma tie at 1 broken alpha-ASC → **beta, alpha, gamma** (3
   threads, 4 impact beats shown of 6, `Beats: 4/6`); `(no project)` folded
   (0 impact beats), `alpha-messy` dropped (impact-less).
3. **Golden agreement (SPEC-023 golden-vs-golden cross-check):** Test 1
   (me markdown), Test 3 (me JSON), and Test 5 (me `BuildThreads`) all
   assert the SAME four-thread `me` set. Test 2 (exec markdown) and Test 6
   (exec `BuildThreads`) assert the SAME three-thread exec set. The two
   sets are consistent (exec = me minus the folded impact-less `(no
   project)` thread, minus dropped impact-less beats, reordered by
   impact-desc). Verified at design; no golden contradicts another.
4. **The directive assets are byte-embedded in the goldens, not
   hand-copied.** Tests 1/3 splice `me.md`'s full text; Test 2 splices
   `exec.md`. Build loads the asset bytes and asserts the golden's
   directive section == `directiveAsset("me.md")` / `("exec.md")`, so the
   asset file and the golden cannot silently diverge.

### Registration

Mirror `NewImpactCmd()`'s registration in `cmd/brag/main.go`:
`root.AddCommand(cli.NewStoryCmd())`.

## Locked design decisions

Each has ≥1 paired failing test (§9 traceability).

1. **A thread is a DETERMINISTIC initiative unit (reuse
   `GroupEntriesByProject`), time-ordered, with impact beats marked
   (`WithImpact`); `--theme` adds an opt-in cross-project cross-cut; the
   throughline is a deterministic SKELETON; the LLM (via the directive)
   finds the arc (DEC-029 choice 1).** Paired tests: **Test 5**
   (deterministic order + impact marking), **Test 7** (theme cross-cut),
   **Test 8** (throughline skeleton), **Test 1** (the rendered me arc).

2. **Audiences are DATA-DRIVEN shaping profiles (bundled default assets +
   user override), NOT a Go enum (DEC-029 choice 2).** Paired tests:
   **Test 9** (bundled load, no Go audience enumeration), **Test 10**
   (user override shadows bundled), **Test 11** (unknown/malformed).

3. **`--audience` sets how-many-arcs + altitude, not just tone: me keeps
   every thread + impact-less beats + low altitude; exec keeps
   impact-bearing threads only + drops impact-less beats + one-headline-arc
   altitude (DEC-029 choice 3).** Paired tests: **Test 6** (exec folds +
   drops, mechanism), **Test 2** (exec render), **Test 15** (live me-vs-exec
   divergence — the headline assertion).

4. **The bundle EXTENDS DEC-014 with an arc-aware body (`audience`,
   `threads`, `throughline`, `framing_directive`); the 7-key beat
   projection is a deliberate subset+2 of DEC-011 (DEC-029 choice 5).**
   Paired tests: **Test 3** (JSON shape golden), **Test 1/2** (markdown
   goldens).

5. **The bundle is useful standalone; the directive always renders (even
   empty corpus), the `## Framing directive` section omits only when the
   directive itself is empty (DEC-029 choice 6/8).** Paired tests:
   **Test 4** (empty window: directive renders, threads omitted), **Test 1**
   (standalone-readable body with ★/· markers).

6. **Framing directives are BUNDLED `embed.FS` assets, per-audience,
   printable alone via `--print-directive` (DEC-029 choice 7).** Paired
   tests: **Test 12** (asset resolves, non-empty, me≠exec), **Test 16**
   (`--print-directive` prints only the directive, no DB read).

7. **Profile files use a flat `key: value` schema parsed by a hand-rolled
   stdlib parser — NO new YAML dependency.** The schema has no nesting/
   lists; a ≈30-line `bufio` parser keeps `go.mod` clean
   (`no-new-top-level-deps-without-decision`). Paired tests: **Test 9/10**
   (parse bundled + override). *(If build hits a real nesting need, adding
   `gopkg.in/yaml.v3` requires a DEC — Rejected alternative below.)*

8. **Window logic is LIFTED to a shared `internal/cli/window.go`
   (`windowCutoff`/`selectedWindow`), reused by `impact` + `story`; story
   supplies the profile default when no window flag is set — a
   behavior-preserving refactor for `impact`.** Paired test: **Test 14**
   (window resolution + shared helper; impact's existing tests stay green).

9. **`--audience` required; `--format` defaults to markdown; window flags
   mutually exclusive; stdout/stderr separation; no SQL in CLI/story;
   wrapped errors; pure-Go asset embedding (no model/network).** Paired
   tests: **Test 13** (required audience), **Test 17** (format default +
   unknown), **Test 18** (stdout/stderr separation).

### Rejected alternatives (build-time)

- **LLM-inferred threading in the binary.** Rejected (DEC-029 Option D):
  bakes a model in (violates DEC-001 / the pure-pipe posture) or makes the
  bundle depend on a model to have structure (breaks standalone
  usefulness). Inference is reintroduced at the correct layer — the
  framing directive asks the *caller's* LLM to find the throughline.

- **Theme-tags (or time-progression) as the PRIMARY thread axis.**
  Rejected (DEC-029 Options B/C): tags are sparse/freeform (fragile,
  silently drops untagged work); "related over time" with no key is
  inference in disguise. Both are folded IN as, respectively, the opt-in
  `--theme` cross-cut and the within-thread ASC ordering + skeleton span.

- **Reusing DEC-028's flat `impact_by_project` envelope verbatim.**
  Rejected (DEC-029 Option H): no cross-thread ordering, no impact-beat
  marking, no throughline skeleton, no directive slot — cannot carry an
  arc. The arc-aware body extends DEC-014's provenance envelope instead.

- **A hard-coded Go audience enum + `switch`.** Rejected (DEC-029 Option
  F): the stage requires extensible profiles; an enum means every new
  audience is a Go change + release. Profiles-as-data (bundled + override)
  is the mechanism.

- **Adding `gopkg.in/yaml.v3` for profile parsing.** Rejected at this spec
  (Locked decision #7): the flat `key: value` profile schema needs no YAML
  engine; a stdlib line parser avoids a top-level dep
  (`no-new-top-level-deps-without-decision`). Revisit only if a profile
  needs nested/list config — then emit a DEC for the dep.

- **A story-local COPY of `windowCutoff`/`selectedWindow` (not lifting).**
  Rejected (Locked decision #8): story is the third caller of window logic,
  which crosses SPEC-018's lift threshold; a copy would duplicate the
  calendar-vs-rolling correctness core across two files (drift risk on the
  single most load-bearing bit of DEC-028). Lift to `window.go`; keep
  impact's tests green.

- **Baking the profile's altitude into the BODY as a hard filter for
  `me`.** Rejected: `me`'s "low altitude / detail over polish" is a
  DIRECTIVE instruction to the LLM (candor metadata), not a body-selection
  rule — `me` keeps every thread + beat by design. Only `exec`'s altitude
  is expressed as body rules (fold/drop). Mixing the two would make `me`
  lossy, defeating "candid, complete."

- **Putting `--theme` into `ListFilter.Tag` (a second List call).**
  Rejected: the theme cross-cut is computed over the already-fetched
  in-window set in `internal/story`, so one `List(Since)` call serves both
  the initiative threads and the cross-cut. A second filtered query would
  duplicate the read and risk window skew.

## Premise Audit (AGENTS.md §9 — additive: new-command doc references)

This spec adds a new command (`brag story`). Per §9's new-command case,
grep the docs + `BRAG.md` for the shipped-command surface and enumerate
every hit as a planned Outputs update. **Design-side: greps run against
the repo, expected hits reconciled below (§9 audit-grep cross-check);
build RE-RUNS and reconciles any delta as a question, not silent scope
expansion.**

```
grep -rn "brag summary\|brag review\|brag stats\|brag impact" docs/ README.md BRAG.md
```

Expected: the digest family is documented together in
`docs/api-contract.md` (per-command sections) and enumerated in
`docs/tutorial.md`, `README.md` feature lists, and `BRAG.md`'s "Reading
entries back" section. Build adds a `brag story` section to
`api-contract.md` (mirroring the `brag impact` section) and a line to any
tutorial/README/BRAG.md command list that enumerates the read/digest
commands. **Build re-runs this grep and reconciles the actual hit set
before the doc sweep.**

**Refactor-safety audit (the `windowCutoff` lift):** lifting
`windowCutoff`/`selectedWindow` out of `impact.go` MUST NOT change their
behavior. Enumerate the existing tests whose premise touches them —
`internal/cli/impact_test.go` (`TestImpactCmd_CalendarNotRolling`,
`TestImpactCmd_RequiresExactlyOneWindow`, `TestWindowCutoff_CalendarArithmetic`
if present) — as tests that MUST stay green byte-for-byte after the lift.
No inversion/removal; a pure move. (Grep to run at build:
`grep -rn "windowCutoff\|selectedWindow" internal/cli/`.)

No count-asserted collection is touched (story goldens are self-contained
in the new `internal/story` package; the DEC list in docs is prose).

## Build Completion

Built to the design as specified. All 18 failing tests pass; all six gates
green (`go test ./...` = 647 passed, `gofmt -l .` empty, `go vet ./...`
clean, `CGO_ENABLED=0 go build ./...` success, `just test-docs` ALL OK,
`just test-hook` ALL OK). The me/exec byte-exact goldens (Tests 1/2/3) and
the live me-vs-exec divergence (Test 15) match the design verbatim; the
windowCutoff/selectedWindow lift to `internal/cli/window.go` kept impact's
existing tests green byte-for-byte (349 CLI tests pass).

**Files changed:**
- New: `internal/story/{embed,profile,thread,bundle}.go` +
  `{thread,bundle,profile}_test.go`; assets
  `internal/story/profiles/{me,exec}.yaml`,
  `internal/story/directives/{me,exec}.md`.
- New: `internal/cli/{story.go,story_test.go}`,
  `internal/cli/window.go` (the lifted shared helper).
- Edited: `internal/cli/impact.go` (removed the lifted funcs; behavior
  preserved), `cmd/brag/main.go` (registered `NewStoryCmd`),
  `docs/api-contract.md` (story section + DEC-029 list entry),
  `docs/tutorial.md`, `README.md`, `BRAG.md` (command-surface additions),
  `guidance/questions.yaml` (both SPEC-049 questions resolved).

- **Deviations from spec:** None material. Two build-time helper
  additions the spec pre-authorized: (1) `windowFlagsSet(cmd)` in
  `window.go` gating `selectedWindow` so `story`'s no-window default path
  never trips `selectedWindow`'s zero-flag `UserError` (spec Implementation
  Context named this as the clean alternative — added it). (2)
  `story.ResolveDirective(Profile)` as the public directive-text resolver
  the CLI calls (a bare basename resolves against the embedded assets; a
  path-separator'd `Directive` reads a user file), plus
  `bundledProfileNames()` used by Test 9 to prove the audience set is FS-
  derived, not a Go enum. The hand-rolled profile parser rejects unknown
  keys and non-boolean booleans (beyond the ≈30-line sketch — cheap typo
  protection, no schema change). No `gopkg.in/yaml.v3`; go.mod unchanged.
- **New DEC-* files:** None beyond DEC-029 (emitted during the design
  cycle, already on the branch). No DEC-030 was needed — every build
  decision fell inside DEC-029's locked envelope.
- **Reflection (3 answers):**
  1. *Did the goldens hold?* Yes, byte-for-byte, with one test-only
     nuance: the directive assets end with a trailing newline, but the
     bundle's whole-document trailing-newline trim (matching
     `ToImpactMarkdown`) drops the final one, so the markdown goldens
     splice a `strings.TrimRight(directive, "\n")` tail while the JSON /
     `--print-directive` paths keep the directive verbatim. The renderer
     logic is unchanged; only the golden's spliced tail reflects the trim.
  2. *Was the window lift risky for one-spec-per-pr?* No. The move was
     mechanical (verbatim functions to `window.go`), and impact's existing
     `TestWindowCutoff_*` / `TestImpactCmd_CalendarNotRolling` /
     `_RequiresExactlyOneWindow` stayed green with zero edits — the
     refactor-safety audit's enumerated tests confirmed no behavior change.
  3. *Did two audiences prove divergence?* Yes — the live smoke (me: 4
     threads / 6-6 beats / (no project) kept; exec: 3 threads / 4-6 beats /
     impact-desc / (no project) folded) is the same-corpus-different-story
     proof, rule-driven in the deterministic body, not tone. Recorded in
     the resolved `story-two-audience-slice-proves-divergence` question.

## Verify

*(Filled during verify.)*

## Reflection

*(Filled during ship.)*

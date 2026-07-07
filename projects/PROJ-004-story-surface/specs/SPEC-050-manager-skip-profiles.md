---
# Maps to ContextCore task.* semantic conventions.
# Claude-only variant: the handoff context lives in ## Implementation Context.

task:
  id: SPEC-050
  type: task                       # epic | story | task | bug | chore
  cycle: design
  blocked: false
  priority: high                   # completes v0.4.0's audience gradient + the extensibility proof
  complexity: S                    # S — two bundled asset pairs (config only) + one Go test; ZERO production-Go change to the mechanism.

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
    - DEC-029   # REUSED (the whole mechanism) — profiles-as-data (embed.FS defaults + user override), the flat key:value schema, the thread definition, the arc-aware bundle. SPEC-050 adds two more profiles WITHOUT touching it. This spec is the extensibility PROOF of DEC-029 choice 4.
    - DEC-028   # REUSED — calendar windows (windowCutoff): manager defaults to month, skip to quarter.
    - DEC-014   # REUSED — provenance envelope + empty-state; the empty-directive test asserts an omission branch DEC-014's empty-state already governs.
    - DEC-001   # PRESERVED — pure Go, local-first, no model/network/secrets; no new dependency. The assets embed via the existing embed.FS glob.
  constraints:
    - test-before-implementation
    - one-spec-per-pr
    - no-new-top-level-deps-without-decision   # untouched: no new dep; the assets ride the shipped embed.FS + hand-rolled parser.
    - stdout-is-for-data-stderr-is-for-humans
    - no-cgo
  related_specs:
    - SPEC-049   # shipped; emitted DEC-029, shipped the mechanism + the me/exec endpoints. SPEC-050 fills the middle of the gradient it defined + closes the AC-8 coverage gap SPEC-049 verify flagged.
    - SPEC-048   # shipped; brag impact — windowCutoff / calendar windows reused by the profile default windows.
---

# SPEC-050: `manager` + `skip` audience profiles — the zero-Go-change extensibility proof (+ the empty-directive omission test)

## Context

Second and final spec of STAGE-012. SPEC-049 shipped `brag story
--audience` with a **profiles-as-data** mechanism (bundled `embed.FS`
`profiles/{me,exec}.yaml` + `directives/{me,exec}.md`, a user-override
layer, a hand-rolled flat `key: value` parser, no new dependency) and the
gradient **endpoints** `me` and `exec`. DEC-029 choice 4 deferred the
**middle of the gradient** — `manager` and `skip` — to this spec **as the
extensibility proof**: adding an audience must be **new bundled asset files
with ZERO production-Go change to the mechanism**. That zero-Go-change
property IS the point — it is the living proof that DEC-029's
profiles-as-data design is genuinely extensible ("same in theory, diverge
in practice" without touching code).

This spec:

1. **Ships two new bundled default audiences — `manager` and `skip`** — as
   four new asset files: `internal/story/profiles/manager.yaml`,
   `internal/story/profiles/skip.yaml`,
   `internal/story/directives/manager.md`,
   `internal/story/directives/skip.md`. **No `.go` file is edited.** The
   shipped `//go:embed profiles/*.yaml directives/*.md` glob
   (`internal/story/embed.go:10`) auto-discovers them at build; the shipped
   `LoadProfile(name)` resolves them by name from the embedded FS; the
   shipped `--audience` validation (`internal/cli/story.go`) has **no
   hard-coded audience allowlist** — an unknown name is `ErrProfileNotFound`,
   a resolvable one just works. See the **zero-Go-change trace** below.

2. **Places `manager` and `skip` on the gradient between `me` and `exec`**,
   using ONLY the knobs the shipped `Profile` struct/parser already accepts
   (`default_window`, `impact_threads_only`, `drop_impactless_beats`,
   `fold_small_threads`, `thread_order`, `candor`, `directive`). No new
   knob, no schema change.

3. **Closes the one coverage gap SPEC-049 verify flagged** — a dedicated
   Go test asserting the `## Framing directive` section is **OMITTED when
   the directive is empty** (AC-8's omission branch, `bundle.go:90` `if
   opts.Directive != ""`), and that JSON `framing_directive` is the empty
   string when `Directive == ""`. This is a **new test file addition** (a
   test, written at design — allowed under `test-before-implementation`);
   it touches no production Go.

DEC-029 is **reused wholesale, not amended.** Every shaping knob these two
profiles use already exists and is already tested; SPEC-050 only supplies
new data values for them. If adding these profiles required editing any
`.go` file to accept them, the mechanism would be less extensible than
DEC-029 claimed — that would be a finding. **It does not** (trace below).

Parent stage:
[`STAGE-012-brag-story-audience.md`](../stages/STAGE-012-brag-story-audience.md)
(Success Criteria: `brag story --audience <me|manager|exec>` demonstrably
different per audience; the whole gradient designed, sliced in ship order).
DEC:
[`DEC-029`](../../../decisions/DEC-029-story-audience-shaping-profiles-and-thread-definition.md)
(choice 2 profiles-as-data; choice 4 the me/exec-now, manager/skip-later
slice). Brief:
[`brief.md`](../brief.md) (the four-point audience gradient this completes).

## The zero-Go-change trace (THE HEADLINE — is it truly config-only?)

Traced against the shipped code on `main`. Adding `manager`/`skip` requires
**zero production-`.go` change**:

1. **Asset discovery is a GLOB, not a manifest.**
   `internal/story/embed.go:10` — `//go:embed profiles/*.yaml
   directives/*.md`. Dropping `manager.yaml`/`skip.yaml` +
   `manager.md`/`skip.md` into those directories includes them in the
   binary at build with no edit to the embed directive.

2. **Resolution is by-name from the FS, no enum.**
   `LoadProfile(name)` (`profile.go:61`) checks the override dir, then
   `bundledProfile(name)` (`embed.go:16` → `assetsFS.ReadFile("profiles/"
   + name + ".yaml")`). No Go type enumerates the audience set (SPEC-049's
   `bundledProfileNames()` derives it from the FS). A new asset is a new
   audience.

3. **`--audience` validation has NO hard-coded allowlist.**
   `runStory` (`story.go:91-105`): `Changed("audience")` → non-empty →
   `LoadProfile` → an `ErrProfileNotFound` becomes a `UserError`. There is
   **no `switch audience { case "me", "exec": }`**. `--audience manager`
   resolves the moment `manager.yaml` exists.

4. **Every shaping knob is already data-driven.** `BuildThreads`
   (`thread.go:81`) branches on `opts.Order`,
   `opts.ImpactThreadsOnly`/`FoldSmallThreads`, `opts.DropImpactlessBeats`
   — all lifted off the `Profile` via `ThreadOptionsFromProfile`. The
   parser (`profile.go:98`) already accepts all seven keys. `manager`/`skip`
   set existing knobs to existing values.

**One finding (documentation-freshness, NOT a mechanism failure — see
Locked decisions #4 + Rejected alternatives).** The cobra `Long` help
string and the `--audience` flag-usage/required-error strings in
`internal/cli/story.go` (lines 34-35, 52, 92) literally read **"one of:
me, exec"** and list only `me`/`exec` in the audience table. These are
**help/doc text, not validation** — they do NOT gate `manager`/`skip` from
working; a `--audience manager` run succeeds and its help omission is
purely cosmetic. Editing them to mention the new audiences would be a
`.go` edit, which would technically break the "zero Go change" claim.
**Locked decision (LD4): DO NOT edit `story.go`.** Keep the mechanism
proof clean (zero `.go` change) and update the audience-surface **docs**
(`docs/api-contract.md`, `docs/tutorial.md`, and — for the audience list —
leave the in-binary `Long` as the shipped generic "or a user profile"
phrasing, which already covers extensibility). The help text saying "one
of: me, exec" while `manager`/`skip` also work is the acceptable, honest
cost of the zero-Go-change proof; it is explicitly recorded here, not a
silent gap. (A future cosmetic-only spec may regenerate the help table
from `bundledProfileNames()` — noted, out of scope here.)

**Verdict: config-only is real.** Two asset pairs + one test file. No
production Go touched. DEC-029's extensibility claim holds.

## The audience gradient (where manager/skip land)

Using ONLY the shipped `Profile` knobs. From the brief:

- **me** (shipped) — candid; every thread; impact-less beats KEPT; low
  altitude. `window=year`, all-false, `order=initiative`, `candor=candid`.
- **manager** (this spec) — 1:1 / weekly: *what shipped, blockers, what's
  next; tactical, fairly complete.* Keep every thread + every beat (a
  manager needs the blocked/messy work too, not only the wins), but a
  **tighter reporting cadence** (`window=month`) and a **tactical**
  directive voice. Body policy = me's keep-all; the divergence from `me`
  is the **window** (month vs year → a different beat set from the same
  corpus) and the **directive**. `window=month`, all-false,
  `order=initiative`, `candor=candid`.
- **skip** (this spec) — skip-level / director: *outcomes grouped by
  initiative; less detail, more "so what."* Surface only
  **impact-bearing initiatives** (`impact_threads_only` +
  `fold_small_threads` → the (no project) noise + zero-impact threads fold
  away), but **KEEP the non-impact beats inside a surfaced initiative**
  (`drop_impactless_beats=false`) — a director wants the *shape* of an
  initiative (what it took), not only its impact lines; that KEEP is
  exactly what separates `skip` from `exec`. Group **by initiative**
  (`order=initiative`), NOT one impact-desc headline (that is exec's job).
  `window=quarter`, `order=initiative`, `candor=promotional`.
- **exec** (shipped) — business impact only, terse, promotional; one
  headline arc. `window=quarter`, all-true, `order=impact-desc`,
  `candor=promotional`.

The gradient is monotonic in "how much is dropped":
`me` (keep all) → `manager` (keep all, tighter window) → `skip` (fold
zero-impact threads, keep their non-impact beats, group by initiative) →
`exec` (fold zero-impact threads, drop ALL non-impact beats, one
impact-desc headline). **Each of the four is distinct on at least one
axis** (see the divergence matrix in Implementation Context).

**Ship BOTH `manager` and `skip`? YES (confirmed).** DEC-029 choice 4 said
"`manager` (and optionally `skip`)". Ship both: it completes the brief's
full four-point gradient in one release, and each audience is just two
asset files — two profiles is a **stronger** extensibility proof than one
(it shows the mechanism scales, not that one more happened to fit). Cost is
symmetric and tiny (config + golden). Confirmed: both.

## Goal

Ship `brag story --audience manager` and `brag story --audience skip` as
**bundled default profiles + framing-directive assets ONLY** — four new
asset files, **zero production-Go change** — completing the v0.4.0 audience
gradient and proving DEC-029's profiles-as-data mechanism is extensible.
Plus a dedicated Go test that the `## Framing directive` markdown section
is omitted (and JSON `framing_directive` is empty) when a profile's
directive is empty — closing the AC-8 omission-branch coverage gap
SPEC-049 verify flagged. No new dependency; no model/network; the assets
embed via the shipped `embed.FS` glob.

## Inputs

- **Files to read:**
  - `/AGENTS.md` — §9 (golden style, load-bearing-golden-first,
    literal-artifact-as-spec, NOT-contains self-audit); §12 (decide-at-
    design; literal-artifact-as-spec; §12b design-time pre-flight of the
    embedded goldens against the profile policy); §14 confidence.
  - `/guidance/constraints.yaml` — the referenced constraints (note
    `no-new-top-level-deps-without-decision` stays clean).
  - `/decisions/DEC-029-story-audience-shaping-profiles-and-thread-definition.md`
    — REUSED wholesale (choice 2 profiles-as-data; choice 4 the slice).
  - `internal/story/profile.go` — the shipped `Profile` struct + parser
    (the ONLY knobs `manager`/`skip` may use; the parser rejects unknown
    keys, so a typo'd key fails the load).
  - `internal/story/profiles/{me,exec}.yaml`,
    `internal/story/directives/{me,exec}.md` — the shipped assets to mirror
    (schema + directive voice/shape).
  - `internal/story/embed.go` — the `//go:embed` glob that auto-discovers
    the new assets; `bundledProfileNames()` (proves the set is FS-derived).
  - `internal/story/thread.go` — `BuildThreads` (the shaping policy the new
    profiles drive); `internal/story/bundle.go` — `ToStoryMarkdown`
    (`bundle.go:90` the `if opts.Directive != ""` omission branch the new
    test targets) + `ToStoryJSON`.
  - `internal/story/{profile_test,thread_test,bundle_test}.go` — the
    conventions the new tests mirror (`storyFixture`, `meThreadOpts`,
    `withOverrideDir`, `mustDirective`/`mustDirectiveTrimmed`).
- **Data read:** none new — the profiles drive the SHIPPED read/shaping
  path (`Store.List` → `BuildThreads` → bundle). No new command, flag, or
  query.
- **No schema change. No new dependency. No production-Go change.**

## Outputs

- **New asset `internal/story/profiles/manager.yaml`** (LITERAL below;
  build transcribes verbatim).
- **New asset `internal/story/profiles/skip.yaml`** (LITERAL below).
- **New asset `internal/story/directives/manager.md`** (LITERAL below).
- **New asset `internal/story/directives/skip.md`** (LITERAL below).
- **New test file `internal/story/bundle_empty_directive_test.go`** (the
  AC-8 omission-branch test — Test E1 below). *(A new `_test.go` file
  keeps the addition isolated and reviewable; it may also live appended to
  `bundle_test.go` — build's choice, both are `internal/story` test-only.
  The dedicated file is preferred for traceability.)*
- **New/edited test coverage for the two profiles** — the profile-load
  goldens (Test P1/P2), the divergence assertion (Test P3), and the
  directive-asset assertions (Test P4). These go in a new
  `internal/story/manager_skip_test.go` (or appended to
  `profile_test.go`/`bundle_test.go`; new file preferred for traceability).
  All are `internal/story` test-only — **no production Go.**
- **Edit `docs/api-contract.md`** — add `manager` + `skip` to the
  audience list in the `brag story` section (the `--audience <name>`
  section, ~line 489+; mirror the `me`/`exec` descriptions). Update the
  default-window note (`manager` → month, `skip` → quarter).
- **Edit `docs/tutorial.md`** — extend the audience explanation
  (~line 545-549) to mention `manager` (tactical, month) and `skip`
  (outcomes-by-initiative, quarter) alongside `me`/`exec`.
- **NO edit to `internal/cli/story.go`** (the zero-Go-change proof; LD4).
- **NO edit to `README.md`/`BRAG.md`** unless a command-surface list there
  enumerates the audience *names* (they show only example invocations with
  `me`/`exec` — those stay as illustrative examples; adding manager/skip
  examples is optional polish, not required). **Build re-runs the premise
  grep (below) and reconciles.**
- **NO new DEC.** DEC-029 covers the mechanism; adding data to it is the
  proof, not a new decision. (The documentation-freshness finding is
  recorded in this spec + LD4, below the DEC bar — see §14 note.)

### Premise audit (AGENTS.md §9 — additive: new audience *values*, not a new command)

This spec adds new *audience values* to an existing command — NOT a new
command. The relevant §9 case is the **doc-references / status-change**
audit over the audience-name surface. Design-side grep (RUN at design,
reconciled below; build RE-RUNS and reconciles any delta as a question,
not silent scope expansion):

```
grep -rn "one of: me, exec\|me, exec\|me → year\|exec → quarter\|--audience me\|--audience exec" docs/ README.md BRAG.md internal/cli/story.go
```

Reconciled expected hits (run at design 2026-07-06):
- `internal/cli/story.go` (lines 34-35, 52, 92) — the "one of: me, exec"
  help/error strings. **DELIBERATELY NOT edited (LD4, zero-Go-change).**
- `docs/api-contract.md` (~489-548) — the `brag story` audience section +
  default-window note. **Edited: add manager/skip.**
- `docs/tutorial.md` (~538-549) — the audience explanation +
  default-window line. **Edited: add manager/skip.**
- `README.md:155`, `BRAG.md:319` — illustrative `--audience exec`
  invocation examples only (no audience-name enumeration). **Left as-is**
  (they are examples, not a taxonomy list; adding manager/skip is optional
  polish, not a status claim needing update).

No count-asserted collection is touched (the profile tests are
self-contained in `internal/story`; `bundledProfileNames()` is asserted by
membership, not by a literal count — but SEE the count-coupling note next).

**Count-coupling check (§9 additive case).** `bundledProfileNames()`
returns every `profiles/*.yaml` basename. Grep existing tests for a
literal-count assertion over it before adding two assets:

```
grep -rn "bundledProfileNames\|len(names)\|profiles/\*.yaml" internal/story/*_test.go
```

Design-side finding (run at design): `TestLoadProfile_BundledDefaults`
(`profile_test.go:55-61`) asserts `contains(names, "me") &&
contains(names, "exec")` — a **membership** assertion, NOT a length
assertion, so adding `manager`/`skip` does NOT break it. **Enumerated as
green-after-add** (no edit needed). If build's re-run finds any
`len(names) == 2` style assertion, that is a planned update — none exists
today.

## Acceptance Criteria

1. **`brag story --audience manager`** loads the bundled
   `profiles/manager.yaml` and produces a bundle: keep-all body policy
   (every thread, impact-less beats KEPT — same body rules as `me`),
   `order=initiative` (alpha-ASC, `(no project)` last), default window
   `month`, the `manager` framing directive appended. Over the SAME corpus
   as `me`, manager's body (thread/beat selection) equals me's; the
   divergence from `me` is the default **window** (month vs year) and the
   **directive voice** (tactical vs reflective).

2. **`brag story --audience skip`** loads the bundled
   `profiles/skip.yaml` and produces a bundle that is **demonstrably
   different from both `me` and `exec`** over the same corpus + window:
   surfaces only impact-bearing initiatives (`(no project)` and any
   zero-impact thread FOLDED, like exec) but **KEEPS the non-impact beats
   inside surfaced initiatives** (e.g. `alpha-messy` stays — UNLIKE exec,
   which drops it), grouped **by initiative** (alpha-ASC, NOT exec's
   impact-desc headline order). Default window `quarter`, the `skip`
   directive appended.

3. **Both are added with ZERO production-Go change.** No `.go` file under
   `internal/` or `cmd/` is edited (only new `_test.go` files + the four
   assets + docs). `--audience manager` and `--audience skip` resolve
   purely because the `//go:embed` glob + `LoadProfile` + the
   no-allowlist validation already accept any bundled profile name. The
   full suite (incl. SPEC-049's Tests 1-18) stays green byte-for-byte —
   the new profiles do not perturb `me`/`exec`.

4. **The four profiles are distinct on the gradient.** `me`, `manager`,
   `skip`, `exec` each differ from the other three on at least one shaping
   axis (window, fold/drop policy, thread order, or directive), verified
   by the divergence assertion (Test P3). No two profiles are byte-identical
   in their parsed `Profile` struct.

5. **`manager.yaml`/`skip.yaml` use ONLY the shipped `Profile` keys**
   (`name`, `default_window`, `impact_threads_only`,
   `drop_impactless_beats`, `fold_small_threads`, `thread_order`, `candor`,
   `directive`). No new key; the shipped parser (which rejects unknown
   keys) loads them without error. `skip`'s `drop_impactless_beats` is
   `false` (the me/skip vs exec distinction); `skip`'s `impact_threads_only`
   + `fold_small_threads` are `true` (fold the noise).

6. **Empty-directive omission (AC-8 gap closure).** With a profile whose
   `Directive == ""` (resolved to an empty string), `ToStoryMarkdown`
   OMITS the `## Framing directive` section entirely (no heading, no body)
   while rendering the header + provenance + threads/throughline as normal;
   `ToStoryJSON` renders `framing_directive` as the empty string `""`. A
   non-empty directive renders the section (the positive control). This is
   the dedicated test SPEC-049 verify flagged as missing.

7. **Framing directives are distinct + on-voice.** `manager.md` reads
   tactical (shipped / blockers / next; complete but scoped to the period);
   `skip.md` reads outcomes-by-initiative + "so what" (less detail than
   manager, more than exec's single headline). Both differ from each other
   and from `me.md`/`exec.md` (asserted byte-distinct + substring-on-voice,
   Test P4).

8. **Posture preserved.** No new dependency (`go.mod`/`go.sum` byte-
   unchanged); the assets embed via the shipped `embed.FS`; no model, no
   network, pure Go (`no-cgo`); stdout/stderr separation unchanged (no CLI
   change). `just test-docs` stays green (the new docs additions are
   consistent with the shipped `brag story` surface).

## Failing Tests

Written during design (this spec), made to pass during build. The profile
assets + directive assets are LITERAL (§9 literal-artifact-as-spec) — build
transcribes them verbatim; the tests below assert the loaded/rendered
result. Fixtures reuse SPEC-049's `storyFixture` + `storyFixedNow` (already
in `thread_test.go`, same package).

### The empty-directive omission test (`internal/story/bundle_empty_directive_test.go`)

#### Test E1 — `TestToStory_EmptyDirectiveOmitsSection` (closes the AC-8 gap)

Targets `bundle.go:90` `if opts.Directive != ""`. Uses the `me`-policy
threads over `storyFixture` (a NON-empty corpus, so this is distinct from
SPEC-049 Test 4's empty-*corpus* case) but with `Directive: ""`:

```go
func TestToStory_EmptyDirectiveOmitsSection(t *testing.T) {
    threads := BuildThreads(storyFixture, meThreadOpts)
    opts := StoryOptions{
        Audience:        "me",
        Scope:           "year",
        Filters:         "(none)",
        FiltersJSON:     nil,
        EntriesInWindow: 6,
        Now:             storyFixedNow,
        Threads:         threads,
        Throughline:     BuildThroughline(threads),
        Directive:       "", // <- the omission trigger
    }

    md, err := ToStoryMarkdown(opts)
    if err != nil {
        t.Fatalf("markdown: %v", err)
    }
    // The ## Framing directive section is OMITTED entirely.
    if strings.Contains(string(md), "## Framing directive") {
        t.Errorf("empty directive must omit the ## Framing directive section:\n%s", md)
    }
    // But the rest renders: header, provenance, threads, throughline.
    for _, want := range []string{
        "# Bragfile Story", "Threads: 4", "Beats: 6/6",
        "## Threads", "### alpha", "## Throughline (skeleton)",
    } {
        if !strings.Contains(string(md), want) {
            t.Errorf("empty-directive bundle missing %q:\n%s", want, md)
        }
    }
    // The document must not end with a dangling blank "## " block or a
    // trailing directive artifact — the throughline is the final section.
    if !strings.HasSuffix(strings.TrimRight(string(md), "\n"),
        "(no project) [initiative]: 1 beat, 0 with impact (2026-06-01 → 2026-06-01)") {
        t.Errorf("empty-directive bundle should end at the throughline, got:\n%s", md)
    }

    // JSON: framing_directive is the empty string (not omitted key, not null).
    jsonBody, err := ToStoryJSON(opts)
    if err != nil {
        t.Fatalf("json: %v", err)
    }
    var env struct {
        FramingDirective *string `json:"framing_directive"`
    }
    if err := json.Unmarshal(jsonBody, &env); err != nil {
        t.Fatalf("unmarshal: %v\n%s", err, jsonBody)
    }
    if env.FramingDirective == nil {
        t.Fatalf("framing_directive key must be present (as \"\"), got null/absent:\n%s", jsonBody)
    }
    if *env.FramingDirective != "" {
        t.Errorf("framing_directive: got %q, want empty string", *env.FramingDirective)
    }
    // The literal empty-string form is present (not "null").
    if !strings.Contains(string(jsonBody), `"framing_directive": ""`) {
        t.Errorf("expected framing_directive empty-string literal:\n%s", jsonBody)
    }
}
```

Locks: the omission branch (`bundle.go:90`) is exercised on a non-empty
corpus (complementary to Test 4's empty-corpus-directive-renders case);
the markdown ends at the throughline with no dangling heading; the JSON
carries `framing_directive` as `""` (present, empty, not null). This is
the coverage SPEC-049 verify explicitly flagged (its "Minor observation").

### The manager + skip profile tests (`internal/story/manager_skip_test.go`)

#### Test P1 — `TestLoadProfile_Manager` (bundled load, LOAD-BEARING for the config)

```go
func TestLoadProfile_Manager(t *testing.T) {
    withOverrideDir(t, t.TempDir()) // bundled-only
    p, err := LoadProfile("manager")
    if err != nil {
        t.Fatalf("LoadProfile(manager): %v", err)
    }
    if p.Name != "manager" || p.DefaultWindow != "month" {
        t.Errorf("got name=%q window=%q, want manager/month", p.Name, p.DefaultWindow)
    }
    // Keep-all body policy (same as me): nothing folded, nothing dropped.
    if p.ImpactThreadsOnly || p.DropImpactlessBeats || p.FoldSmallThreads {
        t.Errorf("manager keeps all threads/beats, got %+v", p)
    }
    if p.ThreadOrder != "initiative" || p.Directive != "manager.md" {
        t.Errorf("got order=%q directive=%q, want initiative/manager.md", p.ThreadOrder, p.Directive)
    }
    if p.Candor != "candid" {
        t.Errorf("manager candor: got %q, want candid", p.Candor)
    }
}
```

#### Test P2 — `TestLoadProfile_Skip` (bundled load, LOAD-BEARING for the config)

```go
func TestLoadProfile_Skip(t *testing.T) {
    withOverrideDir(t, t.TempDir())
    p, err := LoadProfile("skip")
    if err != nil {
        t.Fatalf("LoadProfile(skip): %v", err)
    }
    if p.Name != "skip" || p.DefaultWindow != "quarter" {
        t.Errorf("got name=%q window=%q, want skip/quarter", p.Name, p.DefaultWindow)
    }
    // Fold zero-impact threads (like exec) ...
    if !p.ImpactThreadsOnly || !p.FoldSmallThreads {
        t.Errorf("skip folds zero-impact threads, got %+v", p)
    }
    // ... but KEEP the non-impact beats inside surfaced threads (UNLIKE exec).
    if p.DropImpactlessBeats {
        t.Errorf("skip must KEEP non-impact beats (drop_impactless_beats=false), got true")
    }
    // Group by initiative, NOT exec's one-headline impact-desc.
    if p.ThreadOrder != "initiative" {
        t.Errorf("skip order: got %q, want initiative", p.ThreadOrder)
    }
    if p.Directive != "skip.md" || p.Candor != "promotional" {
        t.Errorf("got directive=%q candor=%q, want skip.md/promotional", p.Directive, p.Candor)
    }
}
```

#### Test P3 — `TestProfiles_FourWayGradientDivergence` (LOAD-BEARING — the divergence assertion vs me/exec)

Loads all four bundled profiles and asserts (a) no two are byte-identical
in their parsed struct, (b) the `skip` **body** over `storyFixture`
differs from BOTH `me`'s and `exec`'s body, on the specific axes that
place it between them:

```go
func TestProfiles_FourWayGradientDivergence(t *testing.T) {
    withOverrideDir(t, t.TempDir())
    names := []string{"me", "manager", "skip", "exec"}
    profs := map[string]Profile{}
    for _, n := range names {
        p, err := LoadProfile(n)
        if err != nil {
            t.Fatalf("LoadProfile(%s): %v", n, err)
        }
        profs[n] = p
    }

    // (a) All four distinct as parsed structs (no two identical).
    for i := 0; i < len(names); i++ {
        for j := i + 1; j < len(names); j++ {
            if profs[names[i]] == profs[names[j]] {
                t.Errorf("%s and %s parse to identical profiles: %+v",
                    names[i], names[j], profs[names[i]])
            }
        }
    }

    // (b) Body divergence over the SAME corpus (storyFixture):
    threadsFor := func(n string) []Thread {
        return BuildThreads(storyFixture, ThreadOptionsFromProfile(profs[n], ""))
    }
    me := threadNames(threadsFor("me"))
    skip := threadNames(threadsFor("skip"))
    exec := threadNames(threadsFor("exec"))

    // me keeps all four threads (incl. (no project)); skip folds (no project)
    // but keeps the three impact-bearing initiatives in alpha-ASC order.
    wantMe := []string{"alpha", "beta", "gamma", "(no project)"}
    wantSkip := []string{"alpha", "beta", "gamma"}       // (no project) folded, initiative order
    wantExec := []string{"beta", "alpha", "gamma"}       // impact-desc headline order
    assertEqualSlice(t, "me threads", me, wantMe)
    assertEqualSlice(t, "skip threads", skip, wantSkip)
    assertEqualSlice(t, "exec threads", exec, wantExec)

    // The skip-vs-exec distinction: skip KEEPS alpha-messy (impact-less beat
    // in a surfaced thread); exec DROPS it. Find alpha in each.
    skipAlpha := findThread(threadsFor("skip"), "alpha")
    execAlpha := findThread(threadsFor("exec"), "alpha")
    if beatCount(skipAlpha) != 2 {
        t.Errorf("skip alpha should keep both beats (impact-less kept), got %d", beatCount(skipAlpha))
    }
    if beatCount(execAlpha) != 1 {
        t.Errorf("exec alpha should drop the impact-less beat, got %d", beatCount(execAlpha))
    }
}
```

> `threadNames`, `findThread`, `beatCount`, `assertEqualSlice` are the
> existing/near-existing test helpers in `thread_test.go` (`threadNames`
> is already defined there; build adds the two tiny lookups if not
> present — test-only). The `Profile ==` struct comparison is valid
> because `Profile` is all comparable fields (strings + bools).

Locks: the four-way gradient is real (distinct structs); `skip` sits
strictly between `me` (keeps `(no project)`) and `exec` (drops
`alpha-messy` + reorders impact-desc) — it folds the noise like exec but
keeps initiative detail + initiative order like me. This is the design's
"between the endpoints" claim, made byte-checkable.

#### Test P4 — `TestDirectives_ManagerSkip_ResolveAndVoice`

```go
func TestDirectives_ManagerSkip_ResolveAndVoice(t *testing.T) {
    mgr, err := directiveAsset("manager.md")
    if err != nil {
        t.Fatalf("directiveAsset(manager.md): %v", err)
    }
    skip, err := directiveAsset("skip.md")
    if err != nil {
        t.Fatalf("directiveAsset(skip.md): %v", err)
    }
    me := mustDirective(t, "me.md")
    exec := mustDirective(t, "exec.md")

    for _, d := range [][]byte{mgr, skip} {
        if len(d) == 0 {
            t.Fatal("manager/skip directives must be non-empty")
        }
    }
    // All four directives are pairwise distinct.
    all := map[string]string{"me": me, "exec": exec,
        "manager": string(mgr), "skip": string(skip)}
    seen := map[string]string{}
    for name, body := range all {
        if other, dup := seen[body]; dup {
            t.Errorf("%s and %s directives are byte-identical", name, other)
        }
        seen[body] = name
    }
    // On-voice substrings (the tactical vs outcomes-by-initiative split).
    if !strings.Contains(string(mgr), "blockers") {
        t.Errorf("manager directive should be tactical (mention blockers)")
    }
    if !strings.Contains(string(skip), "initiative") {
        t.Errorf("skip directive should frame outcomes by initiative")
    }
}
```

Locks: both new directives resolve from the embedded FS, are non-empty,
pairwise-distinct from all four, and carry their voice's signature token
(`manager` → "blockers"; `skip` → "initiative").

### The literal profile assets (§9 literal-artifact-as-spec — build transcribes verbatim)

`internal/story/profiles/manager.yaml` (LITERAL):

```
name: manager
default_window: month
impact_threads_only: false
drop_impactless_beats: false
fold_small_threads: false
thread_order: initiative
candor: candid
directive: manager.md
```

`internal/story/profiles/skip.yaml` (LITERAL):

```
name: skip
default_window: quarter
impact_threads_only: true
drop_impactless_beats: false
fold_small_threads: true
thread_order: initiative
candor: promotional
directive: skip.md
```

### The literal directive assets (§9 literal-artifact-as-spec — build transcribes verbatim)

`internal/story/directives/manager.md` (LITERAL):

```
# Framing directive — audience: manager (tactical, 1:1 / weekly)

You are drafting an update for the author's manager — a 1:1 or weekly
check-in. Weave the threads below into a tactical, fairly complete summary.

- Lead with what shipped this period. Then surface blockers and risks —
  a manager needs the friction, not just the wins; beats marked · (no
  recorded impact) often carry the in-progress and blocked work.
- Group by initiative and keep it concrete. This is a working update, not
  a highlight reel or a reflection — enough detail to act on, not every
  thread of reasoning.
- Close each initiative with what is next. Name dependencies the manager
  can unblock.
- Tone: candid, specific, tactical. Complete but scoped to the period.
```

`internal/story/directives/skip.md` (LITERAL):

```
# Framing directive — audience: skip (outcomes by initiative, "so what")

You are drafting an update for the author's skip-level (a director) —
someone one level above the manager. Weave the threads below into
outcomes grouped by initiative.

- Organize by initiative, and lead each with its outcome — the "so what,"
  not the activity. The threads below already surface only the initiatives
  that produced impact; keep that shape.
- Less detail than a 1:1, more than an exec line. Within an initiative you
  may reference the supporting work (including beats marked ·) to show what
  it took, but the outcome leads and the detail supports.
- Do not collapse to a single headline — a director tracks several
  initiatives in parallel. Keep them distinct, each with its own "so what."
- Tone: outcome-forward, measured, promotional. Concise per initiative.
```

## Implementation Context

Everything build needs without re-discovering it.

### The mechanism is shipped — build only adds data + tests

`brag story` (SPEC-049, on main) already reads `Store.List` → `BuildThreads`
→ `ToStoryMarkdown`/`ToStoryJSON`, driven entirely by the loaded `Profile`.
Build's job is: (1) write the four literal asset files verbatim; (2) write
the five tests (E1, P1-P4); (3) update the two docs; (4) run the gates.
**No `.go` under `internal/cli`, `internal/story` (non-test), or `cmd/` is
edited.** Confirm with `git diff --stat` at build: only `_test.go` files +
the four assets + two docs should appear.

### The divergence matrix (design-time pre-flight, §12b)

Over `storyFixture` (alpha: 1 impact + 1 impact-less; beta: 2 impact;
gamma: 1 impact; (no project): 0 impact) — computed at design, build makes
the code produce it:

| profile | window | impact_threads_only | drop_impactless | fold_small | order | threads over fixture | beats |
|---|---|---|---|---|---|---|---|
| me | year | false | false | false | initiative | alpha, beta, gamma, (no project) | 6/6 |
| manager | month | false | false | false | initiative | alpha, beta, gamma, (no project) | 6/6 |
| skip | quarter | true | **false** | true | initiative | alpha, beta, gamma | **5/6** |
| exec | quarter | true | true | true | impact-desc | beta, alpha, gamma | 4/6 |

- **manager's body == me's body** over a shared window (both keep-all).
  manager's real divergence from `me` is the DEFAULT WINDOW (month vs
  year) — over the fixture's calendar, a `month` window (from 2026-07-01,
  given `Now`=2026-07-06) contains ZERO entries, so any manager render
  test that needs a body must pass an explicit `--year`/pass `--year`
  threads (like SPEC-049 Test 2 did for exec's empty-quarter). The profile
  tests (P1) assert the loaded struct, not a windowed render, so they do
  not hit the empty-month issue. This is expected and correct: manager is
  "me, scoped to the reporting cadence + tactically framed" — a legitimate
  gradient point whose body-divergence axis is the window, not folding.
- **skip's `beats 5/6`**: `(no project)` folded (0 impact beats), but
  `alpha-messy` (impact-less, in the surfaced alpha thread) KEPT because
  `drop_impactless_beats=false`. This 5/6 is the exact midpoint between
  me's 6/6 and exec's 4/6 — the checkable "between the endpoints" claim.
- **skip vs exec**: same fold (impact_threads_only + fold_small), but skip
  keeps impact-less beats + initiative order; exec drops them + impact-desc.
- **skip vs me**: skip folds `(no project)`; me keeps it.

Build reconciles the P3 assertions against this table before running — the
`wantSkip = {alpha, beta, gamma}` and the `skip alpha beatCount == 2` vs
`exec alpha beatCount == 1` are the load-bearing rows.

### Why zero-Go-change actually holds (build must PRESERVE this)

The whole spec's value is that build touches no production Go. The temptation
to "fix" the `story.go` help text (which says "one of: me, exec") MUST be
resisted — see LD4 + Rejected alternatives. If build believes a Go change is
required for `--audience manager`/`skip` to WORK (not just to be documented
in help), that is a spec-defect signal: STOP and raise it in
`/guidance/questions.yaml`, because it would contradict the trace above and
mean DEC-029's extensibility claim is weaker than stated. (It does not —
the trace is verified — but the discipline is: a required Go change is a
finding, not a silent edit.)

### NOT-contains self-audit (§9), run at design

Test E1 asserts `## Framing directive` is ABSENT from the empty-directive
markdown. The token `## Framing directive` appears in `bundle.go` (the
heading literal) — that is renderer code the test EXERCISES, not
load-bearing prose the assertion contradicts (the test drives
`Directive:""` precisely to prove the heading is skipped). Clean. Test P3's
NOT-identical assertions compare parsed structs; the profile names
`manager`/`skip` appear in the new assets + tests + docs (where they SHOULD)
— no assertion says a rendered bundle must NOT contain them. Clean.

### §12b design-time pre-flight — the goldens/expected-values are correct at design

- The P3 `wantSkip`/`wantExec`/`wantMe` slices + the skip/exec alpha
  beat-counts were computed from `GroupEntriesByProject(storyFixture)`
  under each profile's policy (the table above) and reconciled against
  SPEC-049's shipped Test 1/2/6 goldens (me=4 threads, exec=beta-alpha-gamma
  3 threads) — they agree by construction.
- Test E1's `HasSuffix` expected tail is the last throughline line from
  SPEC-049 Test 1's me golden (`(no project) [initiative]: 1 beat, 0 with
  impact (2026-06-01 → 2026-06-01)`) — verified against `bundle_test.go:89`.
- The literal `manager.yaml`/`skip.yaml` use only keys the shipped parser
  accepts (`profile.go:114-145` switch) — verified against the shipped
  parser; an unknown key would fail the load, so P1/P2 double as a
  schema-conformance check.

## Locked design decisions

Each has ≥1 paired failing test (§9 traceability).

1. **Ship BOTH `manager` and `skip` as bundled default profiles + directive
   assets (DEC-029 choice 4's "manager and optionally skip" → both).**
   Two profiles complete the brief's full gradient and make a stronger
   extensibility proof. Paired tests: **P1** (manager loads), **P2** (skip
   loads), **P4** (both directives resolve).

2. **The gradient placement: `manager` = keep-all body like `me` + a
   `month` default window + tactical directive; `skip` = fold zero-impact
   threads like `exec` but KEEP their non-impact beats (drop_impactless=
   false) + initiative order + a "so what" directive.** Uses only shipped
   knobs. Paired tests: **P1/P2** (the struct values), **P3** (the four-way
   body divergence, skip strictly between me and exec).

3. **The two new profiles are added with ZERO production-Go change — the
   extensibility PROOF (DEC-029 choice 4).** Only new `_test.go` files, the
   four assets, and docs change. Paired test: **P3/P1/P2** run against the
   UNMODIFIED shipped `LoadProfile`/`BuildThreads`; AC-3 is verified by
   `git diff --stat` showing no non-test `.go` edit + the full SPEC-049
   suite staying green.

4. **DO NOT edit `internal/cli/story.go`'s "one of: me, exec" help/error
   strings — the zero-Go-change proof outranks help-text completeness.**
   The strings are cosmetic (help text, not validation); `manager`/`skip`
   work regardless. Editing them would be a `.go` change that breaks the
   proof. The audience-name surface is updated in DOCS instead
   (`api-contract.md`, `tutorial.md`); the in-binary `Long` keeps its
   shipped generic "or a user profile" phrasing (which already advertises
   extensibility). This is recorded as a **documentation-freshness finding**
   (the help lags the shipped audiences), explicitly accepted, not silent.
   Paired test: AC-3's no-`.go`-edit `git diff --stat` check. *(Rejected
   alternative: regenerate the help table from `bundledProfileNames()` — a
   real improvement but a Go change; deferred to a future cosmetic-only
   spec, out of scope here.)*

5. **The empty-directive omission branch (`bundle.go:90`) gets a dedicated
   test — closing the AC-8 gap SPEC-049 verify flagged.** Distinct from
   SPEC-049 Test 4 (empty *corpus*, directive renders): E1 is a non-empty
   corpus with an empty *directive*, asserting the section is OMITTED
   (markdown) / empty-string (JSON). Paired test: **E1**.

### Rejected alternatives (build-time)

- **Editing `story.go`'s help text to list `manager`/`skip`.** Rejected
  (LD4): it is a production-`.go` change that would break the zero-Go-change
  extensibility proof — the entire point of this spec. The help lag is an
  accepted documentation-freshness cost; the DOCS carry the current
  audience list. A help-table regenerated from `bundledProfileNames()` is a
  genuine future improvement but out of scope (and still a Go change).

- **Giving `skip` `drop_impactless_beats: true` (making it exec-like).**
  Rejected: that would collapse skip into "exec with initiative order" and
  lose the brief's "less detail than manager, more than exec" middle. A
  director wants the *shape* of an initiative (what it took), so skip KEEPS
  the non-impact beats inside surfaced initiatives — the axis that makes it
  distinct from exec (Test P3's skip-alpha-beatCount==2 vs exec==1).

- **Giving `manager` a folding policy (impact_threads_only/drop) to make
  its body differ from `me` over a shared window.** Rejected: a 1:1/weekly
  update needs the blocked + in-progress work (the impact-less beats), so
  manager is legitimately keep-all like `me`. Its gradient distinction is
  the reporting-cadence WINDOW (month) + the tactical directive, not
  body-folding. Forcing a fold would misrepresent the audience to make a
  test prettier.

- **A new `Profile` knob (e.g. `max_threads`, `altitude: N`) for finer
  gradient control.** Rejected: it would be a schema + parser + struct
  change — production Go — defeating the zero-Go-change proof, and DEC-029's
  existing knobs already place all four audiences distinctly. YAGNI; revisit
  only if a real audience cannot be expressed (none here). Would need a DEC.

- **Adding `manager`/`skip` invocation examples to `README.md`/`BRAG.md`.**
  Not rejected but OPTIONAL: those files show illustrative `--audience exec`
  examples, not an audience taxonomy, so they carry no stale status claim.
  Build may add examples as polish; not required for correctness.

## Premise Audit (AGENTS.md §9 — additive: new audience values)

Grep run at design (2026-07-06), reconciled in the Outputs' Premise-audit
note above. Two greps, both re-run at build:

```
grep -rn "one of: me, exec\|--audience me\|--audience exec\|me → year\|exec → quarter" docs/ README.md BRAG.md internal/cli/story.go
grep -rn "bundledProfileNames\|len(names)" internal/story/*_test.go
```

- The audience-surface grep: `story.go` hits are DELIBERATELY left (LD4);
  `api-contract.md` + `tutorial.md` hits are the planned doc edits;
  `README.md`/`BRAG.md` hits are illustrative examples (left as-is).
- The count-coupling grep: `TestLoadProfile_BundledDefaults` asserts
  membership (`contains(names,"me")`), NOT length — adding two assets keeps
  it green (enumerated as green-after-add, no edit). Build reconciles any
  delta as a question.

No inversion/removal (purely additive). No count-asserted collection broken.

## §14 confidence note (no new DEC)

No DEC is emitted: DEC-029 already locks the mechanism, and adding data to
a data-driven mechanism is the PROOF of DEC-029, not a new decision. The
one judgment call — the exact gradient placement of `manager`/`skip` —
sits inside DEC-029 choice 3's "audience sets how-many-arcs + altitude"
envelope; it is design-decidable and pinned by P1/P2/P3, so it needs no
separate decision record. The documentation-freshness finding (help text
lags the shipped audiences, LD4) is recorded here and below the DEC bar (a
cosmetic help-text lag, not an architectural choice). Confidence in the
zero-Go-change proof: **0.95** (traced against shipped code; the only
caveat is the accepted help-text lag, which is cosmetic and explicitly
recorded). Confidence in the gradient placement: **0.85** (manager's
window-not-fold divergence is the honest reading of the brief; a reasonable
alternative would give manager a light fold, but that would misrepresent
the 1:1 audience — recorded in Rejected alternatives).

## Build Completion

*Filled during build.*

## Verify

*Filled during verify.*

## Reflection

*Filled at ship.*

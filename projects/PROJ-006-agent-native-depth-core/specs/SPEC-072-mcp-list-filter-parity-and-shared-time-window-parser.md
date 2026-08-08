---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-072
  type: story                      # epic | story | task | bug | chore
  cycle: ship
  blocked: false
  priority: high
  complexity: M                    # S | M | L  (L means split it)

project:
  id: PROJ-006
  stage: STAGE-019
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-07

references:
  decisions:
    - DEC-042                      # emitted by this spec — the three coupled choices
    - DEC-008                      # --since grammar, reused verbatim
    - DEC-039                      # local calendar day / ParseDay
    - DEC-035                      # ListFilter.Until
    - DEC-033                      # provenance classifier (author)
    - DEC-024                      # MCP server contract this extends
    - DEC-011                      # JSON entry shape (unchanged)
  constraints:
    - no-sql-in-cli-layer
    - test-before-implementation
    - errors-wrap-with-context
    - no-new-top-level-deps-without-decision
    - one-spec-per-pr
  related_specs:
    - SPEC-040                     # created the deferral this retires
    - SPEC-068                     # --day / ParseDay / the L4 clock-seam fix
    - SPEC-056                     # ListFilter.Until
    - SPEC-045                     # ListFilter.Author
---

# SPEC-072: MCP list filter parity and shared time window parser

## Context

STAGE-019's premise is that an agent consults its own history **before** it
acts. The two windows that question needs — "recent" and "agent-authored" —
already exist in `storage.ListFilter` (`Since`, `Until`, `Author`) and are
reachable from `brag list`. Neither is reachable over MCP: `mcpserver.listIn`
carries only `tag`/`project`/`type`/`limit`.

That gap is not an oversight; it is a deferral with a reason recorded in the
code at `internal/mcpserver/server.go:169` —

> *the same exact-match filters `brag list` supports minus `--since` (deferred,
> see Out of scope — `ParseSince` lives in package `cli` and importing it would
> risk a `cli↔mcpserver` cycle).*

The cycle is real and one-directional: `internal/cli/mcp.go` imports
`internal/mcpserver` to run `brag mcp serve`, so `mcpserver` can never import
`cli`. **DEC-042** settles how to resolve it (a shared pure `internal/timewindow`
package taking an explicit `now`, rather than duplicating the parser as DEC-024
already regretted doing for the DEC-010 tokenizer, and rather than inventing an
MCP-only RFC3339 grammar).

This is STAGE-019's first spec: the cheapest of the three, and the one that
removes a structural blocker SPEC-073/SPEC-074 would otherwise hit.

## Goal

Expose `since` / `until` / `day` / `author` on the MCP `brag_list` tool with the
CLI's exact grammar and semantics, by moving `ParseSince` / `ParseDay` into a
new pure `internal/timewindow` package that both `internal/cli` and
`internal/mcpserver` import. No behavior change to any existing surface.

## Inputs

- **Files to read:**
  - `decisions/DEC-042-mcp-time-window-filter-parity.md` — the three coupled
    choices this spec implements verbatim. Read first.
  - `internal/cli/since.go` — `ParseSince` / `ParseDay` / the `clock` seam; the
    code being moved.
  - `internal/cli/since_test.go` — the tests being migrated and simplified.
  - `internal/mcpserver/server.go` — `listIn` / `handleList` / `searchIn` /
    `handleSearch`; the deferral note at line 169 that this retires.
  - `internal/cli/list.go` — the reference implementation of every filter's
    validation, error text, and mutual-exclusion rule (the MCP boundary mirrors
    it, it does not re-invent it).
  - `internal/storage/entry.go` — `ListFilter`'s field semantics (inclusive
    `Since`, **exclusive** `Until`, `Author` ∈ {`agent`,`human`,``}).
  - `internal/mcpserver/server_test.go` — `newTestServer` / `callJSON` /
    `seedViaStore` / `setNowFunc`, the harness the new tests use.
- **External APIs:** none. No new dependency.
- **Related code paths:** `internal/cli/window.go`, `internal/cli/export.go`
  (the two other `ParseSince` call sites).

## Outputs

- **Files created:**
  - `internal/timewindow/timewindow.go` — pure, stdlib-only `ParseSince(s
    string, now time.Time)` and `ParseDay(value string, now time.Time)`. No
    package-level state.
  - `internal/timewindow/timewindow_test.go` — the migrated `since_test.go`
    cases, rewritten to pass `now` explicitly (no clock substitution), plus the
    two new guards (see Failing Tests).
  - `internal/cli/clock.go` — the new home for `var clock = time.Now` and the
    comment distinguishing it from `impact.go`'s UTC `nowFunc`.
  - `internal/mcpserver/list_filters_test.go` — the new MCP boundary tests.
- **Files modified:**
  - `internal/mcpserver/server.go` — `listIn` gains `since`/`until`/`day`/
    `author`; `handleList` validates + parses them; `handleList`/`handleSearch`
    reject a negative `limit`; the line-169 deferral note is deleted.
  - `internal/cli/list.go` — three call sites rewritten to
    `timewindow.ParseSince(raw, clock())` / `timewindow.ParseDay(raw, clock())`.
  - `internal/cli/window.go` — `ParseSince(sinceRaw)` →
    `timewindow.ParseSince(sinceRaw, clock())`. **Mechanical only** — see the
    trap in Notes for the Implementer.
  - `internal/cli/export.go` — same one-line rewrite.
  - `docs/api-contract.md` — the `brag_list` bullet (line ~1162) currently reads
    *"filters `tag`/`project`/`type`/`limit` … minus `--since`"*. Planned update
    (premise audit, status-change case).
  - `docs/for-ai-agents.md` — §3's `brag_list` param table plus the claim at
    lines 100–101, *"There is **no `--since` filter** over MCP (deferred); filter
    by time on the CLI if you need it."* Planned update (premise audit,
    status-change case).
  - `projects/PROJ-006-.../stages/STAGE-019-corpus-as-agent-memory.md` — backlog
    status line for SPEC-072.
  - `projects/PROJ-006-.../stages/STAGE-018-v0-5-0-audit-backlog.md` — tick the
    "MCP `list`/`search` negative-`limit` parity" item as folded into SPEC-072.
- **Files deleted:**
  - `internal/cli/since.go` and `internal/cli/since_test.go` — contents moved
    (premise audit, inversion case: `cli.ParseSince` / `cli.ParseDay` cease to
    exist; every caller is enumerated above).
- **New exports:** `timewindow.ParseSince`, `timewindow.ParseDay`.
- **Database changes:** none. No migration, no schema change, no new storage
  method.

## Acceptance Criteria

- [x] `internal/timewindow` compiles with only `fmt`, `strconv`, `strings`,
      `time` imported, and contains no package-level `var`.
- [x] `timewindow.ParseSince(s, now)` reproduces DEC-008 exactly: `YYYY-MM-DD` →
      that day's UTC midnight; `Nd`/`Nw`/`Nm` → `now - N×{24h,7×24h,30×24h}`
      normalized to UTC; anything else → error.
- [x] `timewindow.ParseDay(value, now)` reproduces DEC-039 exactly: `today` /
      `yesterday` / `YYYY-MM-DD` resolved in `now.Location()`, returning
      `[local midnight, next local midnight)` via `AddDate` (never `+24h`).
- [x] Neither function reads the wall clock; results depend only on arguments.
- [x] `brag_list` accepts `since`, `until`, `day`, `author` and each narrows the
      result set identically to the equivalent `brag list` invocation over the
      same rows.
- [x] `until` is an **exclusive** upper bound (an entry whose `created_at`
      equals `until` is excluded).
- [x] `day` resolves to a LOCAL calendar day via `mcpserver.nowFunc`, so a
      non-UTC caller's "today" is their day, not the UTC day.
- [x] `day` combined with `since` **or** `until` is a tool error naming the
      conflict; the message does not merely report a parse failure.
- [x] An `author` value other than `agent` / `human` is a tool error naming the
      accepted values.
- [x] An unparseable `since` / `until` / `day` is a tool error quoting the bad
      value and the accepted grammar.
- [x] `brag_list` and `brag_search` reject `limit < 0` as a tool error
      (folded-in v0.5.0 audit item); `limit == 0` still means unlimited.
- [x] `tools/list` advertises `since`/`until`/`day`/`author` as `brag_list`
      input properties, and the server still advertises exactly **four** tools
      (this spec adds no tool).
- [x] Output stays byte-identical to `brag list --format json` over the same
      rows — the filters change *which* rows, never their rendering.
- [x] **Every existing `internal/cli` test passes unchanged.** This is the
      no-behavior-change bar for the package move.
- [x] Full gate set green: `gofmt -l .` empty, `go vet ./...`, `go test ./...`,
      `just test-docs`.

## Failing Tests

Written during **design**, BEFORE build. The implementer's job in **build** is
to make these pass.

- **`internal/timewindow/timewindow_test.go`** (migrated from
  `internal/cli/since_test.go`, `now` now explicit — the substitutions in the
  originals disappear rather than move)
  - `"TestParseSince_ISODate"` — `ParseSince("2026-01-01", ref)` returns
    `2026-01-01T00:00:00Z`; asserts the bare-date path ignores `now` entirely.
  - `"TestParseSince_Days"` / `"TestParseSince_Weeks"` /
    `"TestParseSince_Months"` — `"7d"`/`"2w"`/`"3m"` return `ref` minus
    168h/336h/720h, normalized to UTC.
  - `"TestParseSince_InvalidFormat"` — table (`""`, `"x"`, `"0d"`, `"-3d"`,
    `"5y"`, `"2026-13-01"`) each returns a non-nil error.
  - `"TestParseSince_IsPureFunctionOfNow"` — **new**; two different `now`
    values passed to `ParseSince("7d", …)` produce results differing by exactly
    the same delta as the inputs. Pins DEC-042 sub-decision 1 (the seam is the
    caller's, not the package's).
  - `"TestParseDay_ExplicitDate_ResolvesInNowLocation"` — with `now` in a
    fixed UTC-7 zone, `ParseDay("2026-07-05", now)` returns that zone's
    midnight and the next midnight; asserts `end.Sub(start) == 24h` and that
    both carry `now.Location()`.
  - `"TestParseDay_TodayAndYesterday"` — keywords resolve off `now`, not the
    host clock; `yesterday` is exactly one `AddDate(0,0,-1)` day earlier.
  - `"TestParseDay_InvalidValue"` — `"tomorrow"`, `"07/05/2026"`, `""` each
    error, and the message names the accepted forms.
  - `"TestPackageReadsNoWallClock"` — **new**; walks the package's non-test
    `.go` files and fails if any contains `time.Now(`. The mechanical guard for
    "this package holds no clock" (same idiom as
    `internal/mcpserver/import_audit_test.go`).

- **`internal/mcpserver/list_filters_test.go`** (new; uses `newTestServer` /
  `callJSON` / `seedViaStore` / `setNowFunc` from `server_test.go`)
  - `"TestBragList_Since_RelativeDuration"` — three entries backdated to 1d/10d/
    40d ago; `{"since":"7d"}` returns only the 1d row.
  - `"TestBragList_Since_BareDateIsUTCMidnight"` — asserts the DEC-008
    bare-date semantics survive the move to the MCP boundary (an entry at
    `2026-01-01T00:00:00Z` is included by `since:"2026-01-01"`).
  - `"TestBragList_Until_IsExclusiveUpperBound"` — an entry whose `created_at`
    is exactly the `until` instant is **excluded**; the one a second earlier is
    included.
  - `"TestBragList_Day_ResolvesLocalCalendarDay"` — `nowFunc` stubbed to
    `2026-07-05T12:00:00-07:00`; an entry at 23:30 local on 07-05 is returned by
    `{"day":"today"}` and an entry at 00:30 local on 07-06 is not. This is
    DEC-039's local-vs-UTC skew proof, at the MCP boundary.
  - `"TestBragList_Day_ConflictsWithSince"` — `{"day":"today","since":"7d"}` is
    a tool error whose text contains `mutually exclusive`.
  - `"TestBragList_Day_ConflictsWithUntil"` — same for `until`.
  - `"TestBragList_Author_AgentAndHuman"` — one entry inserted through
    `brag_add` (which stamps `agent:`) and one through `seedViaStore` (which
    does not); `{"author":"agent"}` returns exactly the first,
    `{"author":"human"}` exactly the second.
  - `"TestBragList_Author_InvalidValueIsToolError"` — `{"author":"robot"}` errors
    and the message contains both `agent` and `human`.
  - `"TestBragList_InvalidSinceIsToolError"` — `{"since":"5y"}` errors and the
    message quotes `"5y"`.
  - `"TestBragList_NegativeLimitIsToolError"` — `{"limit":-1}` errors instead of
    silently returning everything (folded v0.5.0 audit item).
  - `"TestBragSearch_NegativeLimitIsToolError"` — same for `brag_search`.
  - `"TestBragList_FilteredOutputMatchesStoreParity"` — for a filter set
    exercising all four new fields at once, the tool's text equals
    `export.ToJSON(store.List(equivalentFilter))` byte-for-byte. Pins "the
    filters change which rows, never the rendering."
  - `"TestBragList_ToolSchemaAdvertisesNewFilters"` — `tools/list` →
    `brag_list`'s input schema `properties` contains `since`, `until`, `day`,
    `author`.
  - `"TestServer_StillExactlyFourTools"` — guards the additive case: this spec
    must not change the tool count. (If `server_test.go`'s existing
    `TestServer_ToolsListed` already asserts this, extend a comment there
    instead of duplicating — do not add a second assertion of the same fact.)

- **Existing tests that must pass unchanged** (the no-behavior-change bar; not
  rewritten, not relaxed): all of `internal/cli` — in particular
  `internal/cli/list_test.go`'s `--day` suite (which stubs `cli.clock`, a seam
  that survives the move) and `internal/cli/impact_test.go`'s
  `TestImpactCmd_SinceReusesParseSince`.

## Implementation Context

*Read this section (and the files it points to) before starting the build cycle.*

### Decisions that apply

- `DEC-042` — **the spec's spine.** Parser home + pure-function signature; the
  CLI grammar verbatim at the MCP boundary; `until` MCP-first; the negative-
  `limit` rider. Consume it verbatim; do not re-litigate its alternatives.
- `DEC-008` — the `--since` grammar (`YYYY-MM-DD` | `Nd`/`Nw`/`Nm`) being moved,
  not changed.
- `DEC-039` — the LOCAL calendar day, and the deliberate asymmetry that bare-date
  `--since` stays UTC-midnight. Both survive the move exactly.
- `DEC-035` — `ListFilter.Until` is an **exclusive** upper bound.
- `DEC-033` — `author` classifies on the presence of a reserved `agent:`/`model:`
  tag; storage's `provenanceExistsClause` is the authority.
- `DEC-024` — the MCP server contract being extended: byte-parity with the CLI
  JSON, stdout carries protocol frames only, tool errors are returned as errors
  from the handler.
- `DEC-011` — the 9-key entry shape. Unchanged by this spec.

### Constraints that apply

- `no-sql-in-cli-layer` — `internal/timewindow` is pure stdlib and imports
  nothing from `internal/storage`; `internal/mcpserver` keeps going through
  `*storage.Store` (its `import_audit_test.go` enforces this and must stay
  green).
- `test-before-implementation` — write the Failing Tests above first; run
  `go test ./...` once and confirm each fails for the *expected* reason before
  touching implementation.
- `errors-wrap-with-context` — every handler error stays
  `fmt.Errorf("brag_list: …")` / `fmt.Errorf("brag_search: …")`.
- `no-new-top-level-deps-without-decision` — none added; `go.mod` must be
  untouched.
- `one-spec-per-pr` — one branch, one PR, `feat/spec-072-mcp-list-filter-parity`.

### Prior related work

- `SPEC-040` (shipped) — built the MCP server and wrote the deferral this
  retires.
- `SPEC-068` (shipped) — added `--day`, `ParseDay`, and the injectable `clock`
  seam (v0.5.0 audit item L4).
- `SPEC-056` (shipped) — promoted `Until` into `ListFilter`.
- `SPEC-045` (shipped) — `ListFilter.Author` + the Go/SQL classifier drift guard.

### Out of scope (for this spec specifically)

- **Filters on `brag_search`.** `Store.Search` takes only a query + limit;
  giving search a filter set is a storage change, not a boundary change. Only
  its negative-`limit` guard is in scope.
- **`brag list --until` on the CLI.** DEC-042 Option D, rejected with a
  pre-authorized reversal trigger.
- **New time grammars** (RFC3339, epoch, `Nh`, `Ny`). DEC-042 revisit (c) says
  any new form is added to `timewindow` so *both* surfaces gain it together —
  and not in this spec.
- **Extracting `buildFTS5Query`** (the sibling duplication DEC-024 flagged). Same
  class of fix, different trigger; do not opportunistically bundle it — that is
  a second spec's worth of change in a PR that must stay reviewable.
- **The memory slice and the resources surface** — SPEC-073 / SPEC-074.
- **The rest of the STAGE-018 audit backlog.** Only the `list`/`search`
  negative-`limit` item is folded in, because it lives in the two functions this
  spec already edits.

## Notes for the Implementer

**The one real trap: `window.go`'s unused `now`.** `windowCutoff(window,
sinceRaw, now, previous)` already receives a `now` — and its `--since` branch
does **not** use it, calling `ParseSince(sinceRaw)` (which reads `cli.clock()`)
instead. The move makes that inconsistency visible and it will look like a bug
worth fixing. **Do not fix it here.** In production the two agree (both are
`time.Now`, differing only in `Location`, and `ParseSince` normalizes to UTC),
but the `impact`/`story` tests stub `nowFunc` while leaving `clock` real, so
switching the argument would change test-visible behavior in a spec whose
acceptance bar is "no behavior change." Write `timewindow.ParseSince(sinceRaw,
clock())` — mechanically preserving today's semantics — and record the
observation under Build Completion → Follow-up work. It is a real (small)
follow-up, not this spec's.

**The seams, and which one to use where.** Three now exist and they are not
interchangeable:

| seam | package | value | why |
|---|---|---|---|
| `clock` | `internal/cli` | `time.Now` (LOCAL) | `--day` needs the user's local calendar day (DEC-039) |
| `nowFunc` | `internal/cli` (`impact.go`) | UTC | calendar reporting windows are UTC-anchored (DEC-028) |
| `nowFunc` | `internal/mcpserver` | `time.Now`, deliberately **not** `.UTC()`'d | so the streak buckets by local day (DEC-022) |

`handleList` uses `mcpserver.nowFunc` for both `since` (relative durations) and
`day` (local-day resolution). That is the correct one: it is local, which is
what `ParseDay` needs, and `ParseSince` normalizes to UTC anyway. Extend that
var's existing doc comment to say it now serves the filters too — a future
reader must not "tidy" it to `.UTC()` and silently break `day`.

**Validation order in `handleList` is locked** (so the error a caller sees names
the real problem, not a downstream symptom):

1. `Limit < 0` → error.
2. `Author` not in {`""`, `"agent"`, `"human"`} → error.
3. `Day != ""` and (`Since != ""` or `Until != ""`) → mutual-exclusion error.
   **Before** any parsing — a caller who sends both a bad `since` and a `day`
   should be told about the conflict, not the parse.
4. parse `Since` → set `filter.Since`.
5. parse `Until` → set `filter.Until`.
6. parse `Day` → set **both** `filter.Since` and `filter.Until`.

**Locked error strings** (literal-artifact: transcribe, do not paraphrase — the
Failing Tests assert on substrings of these):

```go
"brag_list: limit must not be negative, got %d (omit limit for unlimited)"
"brag_list: author must be \"agent\" or \"human\", got %q"
"brag_list: day is mutually exclusive with since/until (day sets the full day window)"
"brag_list: invalid since %q: %w"
"brag_list: invalid until %q: %w"
"brag_list: invalid day %q: %w"
"brag_search: limit must not be negative, got %d (omit limit for unlimited)"
```

**Locked `listIn` shape** (the `jsonschema` tags are the agent-facing
documentation — they are the whole point of the spec, so get the wording right):

```go
type listIn struct {
	Tag     string `json:"tag,omitempty"`
	Project string `json:"project,omitempty"`
	Type    string `json:"type,omitempty"`
	Since   string `json:"since,omitempty" jsonschema:"inclusive lower time bound: YYYY-MM-DD (UTC midnight) or a relative Nd/Nw/Nm (e.g. 7d = last week)"`
	Until   string `json:"until,omitempty" jsonschema:"exclusive upper time bound, same grammar as since; combine with since for a bounded window"`
	Day     string `json:"day,omitempty" jsonschema:"scope to a single LOCAL calendar day: YYYY-MM-DD, today, or yesterday; mutually exclusive with since/until"`
	Author  string `json:"author,omitempty" jsonschema:"provenance filter: \"agent\" (entry carries an agent:/model: tag) or \"human\" (carries neither)"`
	Limit   int    `json:"limit,omitempty" jsonschema:"maximum rows to return; omit or 0 for unlimited"`
}
```

**Field order matters for readability, not behavior** — keep the time fields
adjacent and after the exact-match filters, as above.

**Signature shape in `internal/timewindow`.** Move the bodies verbatim; the only
edits are the signature (`now time.Time` replaces the `clock()` calls) and the
doc comments (which currently reference `cli`-local details). Keep the DEC-008 /
DEC-039 references in the comments — they are why the odd bits (30-day "month",
`AddDate` rather than `+24h`, bare-date-is-UTC) are correct rather than sloppy.

**Doc edits are small and specific.** In `docs/for-ai-agents.md` §3, add the four
rows to the `brag_list` table and **replace** the two-line "no `--since` filter
over MCP (deferred)" claim — deleting it silently would leave a reader who
remembers it unsure whether it changed. Say what it is now, including that
`until` has no CLI counterpart on purpose. Same for `docs/api-contract.md`'s
`brag_list` bullet. `scripts/test-docs.sh` group T asserts
`docs/for-ai-agents.md` stays within a 120–500 line band (currently 197) — these
edits keep it well inside; no test-docs change is needed.

**Premise audit (run at design, hits reconciled against Outputs above).**

| case | grep | hits | disposition |
|---|---|---|---|
| inversion | `grep -rn "ParseSince\|ParseDay" --include="*.go" .` | `cli/since.go`, `cli/since_test.go`, `cli/window.go:71`, `cli/list.go:77,90`, `cli/export.go:94`, `cli/impact_test.go` (name only), `mcpserver/server.go:169` (comment) | all enumerated in Outputs; `impact_test.go` is a test *name*, no code change |
| status change | `grep -rn "brag_list" docs/ README.md BRAG.md AGENTS.md plugin/ scripts/test-docs.sh` | `api-contract.md:1162`, `for-ai-agents.md:90`, `architecture.md:85`, `tutorial.md:1003`, `BRAG.md:214`, `plugin/README.md:6`, `test-docs.sh:1018` | first two are param-list claims → planned updates; the rest name the tool *set* (unchanged by this spec) → no edit |
| additive | `grep -rn "four tools\|exactly four" internal/ docs/ scripts/` | `mcpserver/server.go:21`, `server_test.go`, `transport_test.go:21`, `for-ai-agents.md:65`, `tutorial.md:1002` | **no count bump** — this spec adds properties, not tools. Listed so SPEC-074 (which adds a fifth tool + resources) inherits the enumeration. |
| NOT-contains | n/a | — | no test asserts absence of a token in load-bearing prose |

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-072-mcp-list-filter-parity`
- **PR (if applicable):** see stage backlog / PR body
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - `DEC-042` — MCP time-window filter parity (emitted at design, with this spec)
- **Fail-first result (AGENTS.md §9):** run before any implementation.
  `internal/timewindow` failed to build (`undefined: ParseSince` — expected for
  a new package). The 14 `mcpserver` failures were all the *right* failures and
  two distinct kinds, which is the useful signal: the filter tests failed with
  `validating "arguments": unexpected additional properties`, proving the SDK's
  `additionalProperties:false` schema genuinely did not carry the fields; and
  `TestBragList_NegativeLimitIsToolError` failed with *"expected a tool error,
  got success"*, independently confirming the folded-in audit item was a real
  defect and not a hypothetical.
- **Deviations from spec:** none. Two spec instructions were followed as
  written where the tempting alternative was visible:
  - The `window.go` trap: `timewindow.ParseSince(sinceRaw, clock())` preserves
    today's semantics rather than using `windowCutoff`'s `now` argument. The
    reasoning is now a code comment at the call site, so the next reader does
    not "fix" it either.
  - `TestServer_StillExactlyFourTools` was **not** added — `server_test.go`'s
    existing `TestServer_ToolsListed` already asserts it, so per the spec it got
    an extended comment instead, carrying the enumeration of the six other
    places that name the four tools (for SPEC-074, which will bump the count).
- **Follow-up work identified:**
  - **`windowCutoff`'s unused `now` (small, real).** `windowCutoff(window,
    sinceRaw, now, previous)` takes a `now` its `--since` branch does not use,
    reaching for `clock()` instead. Harmless in production (both are
    `time.Now`, and `ParseSince` normalizes to UTC); visible only in tests,
    which stub `nowFunc` while leaving `clock` real. Worth a small spec that
    unifies the seam *with* the test updates that implies — deliberately not
    bundled here, where the acceptance bar was "no behavior change."
  - **`buildFTS5Query` extraction.** DEC-024's noted duplication is now the
    only remaining instance of the pattern this spec retired, and
    `internal/timewindow` is the precedent to copy. Still gated on DEC-024's
    own trigger (a third consumer) — but the cost of doing it dropped.

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — Nothing blocking. The one thing the spec did not pin was the *test* seam
   for backdating rows in `internal/mcpserver`: the package's own
   `import_audit_test.go` forbids `database/sql` in every file including tests,
   so the obvious "just write an UPDATE" is unavailable. The answer already
   existed (`internal/storage/storagetest.Backdate`, which exists precisely so
   non-storage packages can seed past-dated rows), but the spec named
   `server_test.go`'s helpers as the harness and did not mention it. A one-line
   Inputs entry would have saved the lookup.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — The `no-sql-in-cli-layer` entry should have noted that `internal/mcpserver`
   enforces the same rule *by test convention* rather than by the constraint's
   path glob (which covers `internal/cli/**` only), and that the enforcement
   walks test files too. That is the constraint that actually shaped a build
   decision here, and it was listed without the detail that made it binding.

3. **If you did this task again, what would you do differently?**
   — Write the `mcpserver` failing tests *first* and the `timewindow` ones
   second. The MCP tests produced a genuinely informative fail-first (two
   distinct, meaningful failure modes); the `timewindow` ones could only produce
   `undefined: ParseSince`, which is the "stray compilation error" AGENTS.md
   warns is a weak signal. For a package-*move* spec the real fail-first
   evidence lives entirely at the consuming boundary, and leading with it would
   have made that explicit rather than incidental.

---

## Verify

**✅ APPROVED.**

- Acceptance criteria: 15/15 met.
- Gate set green: `gofmt -l .` empty, `go vet ./...` clean, `go test ./...`
  929 pass across 13 packages, `just test-docs` ALL OK.
- No decision drift: DEC-042's three sub-decisions are each implemented as
  written (pure parser with explicit `now`; CLI grammar verbatim; `until`
  MCP-only) and each is pinned by a test.
- No constraint violations. Worth noting explicitly:
  `internal/mcpserver/list_filters_test.go` reaches SQL only through
  `internal/storage/storagetest`, which exists for exactly this purpose, so
  `mcpserver`'s own `import_audit_test.go` (which walks test files too) stays
  green.
- **Live behavioral check beyond the unit tests** (§12(b) discipline applied at
  verify — a real stdio round-trip, not just the in-memory transport). Drove
  `brag mcp serve` over a pipe with a hand-written JSON-RPC session:
  `{"since":"7d","project":"demo"}` → 1 row; `{"day":"today","since":"7d"}` →
  `isError`, *"day is mutually exclusive with since/until"*;
  `brag_search {"limit":-1}` → `isError`, *"limit must not be negative"*;
  `{"author":"human","day":"today"}` → both CLI-authored rows. **stderr was
  empty and stdout carried protocol frames only**, so DEC-024's transport
  purity holds with the new code paths.
  - *Incidental finding, pre-existing, NOT a regression:* if the client closes
    stdin immediately after writing, the server exits without flushing
    responses. Reproduced identically on the released `/opt/homebrew/bin/brag`
    v0.5.2, so it predates this spec. Real clients hold the pipe open, so this
    is a test-harness footgun rather than a user-facing defect — recorded here
    so the next person who writes a pipe-based smoke test does not spend the
    ten minutes I did.
- CLI regression check: `brag list --day today`, `--since 7d --project demo`,
  and the `--day`+`--since` UserError all behave byte-identically to before the
  package move.

## Reflection (Ship)

*Appended during the **ship** cycle.*

1. **What would I do differently next time?**
   — Lead the fail-first run with the *consuming* boundary. For a package-move
   spec, `undefined: ParseSince` in the new package proves nothing; the useful
   evidence was entirely in `mcpserver`, where two distinct failure modes
   (`unexpected additional properties` for the missing fields, *"expected a
   tool error, got success"* for the negative limit) independently confirmed
   both halves of the spec were real problems. I got that signal by accident of
   ordering rather than by design.

2. **Does any template, constraint, or decision need updating?**
   — Two candidates, neither at the codification bar yet (AGENTS.md's N=3
   same-outcome / N=2 paired-opposing rule):
   - **`no-sql-in-cli-layer`'s path glob is narrower than its enforcement.** The
     constraint covers `internal/cli/**`, but `internal/mcpserver` enforces the
     same rule by test convention over *all* its files including tests. That is
     load-bearing (it dictated using `storagetest` here) and invisible from
     `constraints.yaml`. Worth either widening the glob or adding a one-line
     note pointing at `import_audit_test.go`. N=1 so far.
   - **A "purity guard" test idiom.** `TestPackageReadsNoWallClock` (this spec)
     and `TestNoSQLImport` (SPEC-040) are the same mechanical shape: walk the
     package's own sources, fail on a forbidden literal. It is a cheap way to
     make an architectural claim *enforced* rather than documented. N=2, but
     same-outcome, so it holds for a third instance before codifying.

3. **Is there a follow-up spec I should write now before I forget?**
   — Yes, one small one, already recorded under Follow-up work: unify the
   `windowCutoff` seam so its `now` argument is actually used by the `--since`
   branch (currently it reaches for `clock()`), together with the impact/story
   test updates that implies. Deliberately excluded here because this spec's
   acceptance bar was "no behavior change," and that fix *is* a behavior change
   in tests. The `buildFTS5Query` extraction stays gated on DEC-024's own
   third-consumer trigger, but is now cheaper: `internal/timewindow` is the
   precedent to copy.

4. **What can a user do now that they couldn't before?**
   — An agent connected over MCP can now ask for *"the last 7 days on this
   project"* or *"only what agents logged"* in a single `brag_list` call —
   before, MCP exposed only `tag`/`project`/`type`/`limit`, so any time or
   authorship question meant pulling the whole corpus and filtering
   client-side, or dropping to the shell. Four filters (`since`, `until`,
   `day`, `author`), one call, with the CLI's exact grammar because both
   surfaces now share one parser.

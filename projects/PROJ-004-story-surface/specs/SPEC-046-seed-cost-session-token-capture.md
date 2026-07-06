---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-046
  type: story                      # epic | story | task | bug | chore
  cycle: verify
  blocked: false
  priority: high
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-004
  stage: STAGE-014
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # claude-only variant: same model, separate session
  created_at: 2026-07-05

references:
  decisions: [DEC-027, DEC-024, DEC-015, DEC-011, DEC-012, DEC-025]
  constraints: [stdout-is-for-data-stderr-is-for-humans, no-sql-in-cli-layer, no-cgo, test-before-implementation, errors-wrap-with-context]
  related_specs: [SPEC-040, SPEC-041, SPEC-043]
---

# SPEC-046: Seed cost / session / token capture on the MCP `brag_add` path

## Context

PROJ-005 (team + economics) will want per-work **cost / token / session** data.
But history only accrues going forward: when provenance landed in v0.3.0, every
pre-v0.3.0 entry was permanently un-attributable because it was stamped late
(the corpus had **0** agent-authored history in hindsight). If we wait for
PROJ-005 to start capturing economics data, the corpus is empty in hindsight
exactly the same way.

This spec **seeds** that capture now, cheaply, as a **v0.3.x patch**: it extends
the MCP `brag_add` provenance path (SPEC-040 / DEC-024) with three new
**optional** reserved tags — `session:<id>`, `cost:<n>`, `tokens:<n>` —
emitted by the same `stampProvenance` helper that already stamps
`agent:`/`model:`. The real, reliable payload is the **`session:` join-key** (a
stable per-session id the caller forwards); `cost:`/`tokens:` carry real numbers
**only when a caller supplies them**. bragfile **never fabricates** a value:
all three inputs are optional and empty → no tag, exactly like `agent`/`model`
today. Migration-free — the tags ride the DEC-015 taggings join unchanged.

**This spec's governing decision is DEC-027**, which extends DEC-024's reserved
namespace. Read it in full before build.

### Why bragfile cannot self-count (settled — see DEC-027 Context)

- bragfile is a local SQLite CLI/MCP server; it has **no view of the model's
  usage accounting**, so it cannot count tokens or compute cost.
- The stdio MCP transport exposes **no session id** — the SDK's `Implementation`
  carries only `Name/Title/Version/WebsiteURL/Icons` (verified at DEC-024
  design; the exact reason `model:` had to be an explicit param). So a session
  id, like a model id, **cannot come from the transport** and must be a
  caller-supplied param.
- Therefore the honest payload is a **reliable JOIN-KEY now** (`session:`) plus
  **optional** real cost/tokens. Exact-token reconciliation (join `session:<id>`
  → provider usage logs) is **explicitly PROJ-005**; stringly-typed aggregation
  now is fine (DEC-027 accepts the debt, DEC-004→DEC-015 precedent).

### Join-key delivery path (verified in code — no new transport plumbing)

The plugin Stop hook `plugin/hooks/capture-nudge.sh` **already** reads
`session_id` from its stdin payload (`SESSION_ID=$(... jq -r '.session_id ...')`)
and injects agent-facing `additionalContext`; it **never** calls `brag_add`
(Claude does, post-approval, per BRAG.md). So the hook's job here is to
**surface** the `session_id` in its `additionalContext` and instruct Claude to
forward it as a `session:` param on `brag_add`. No stdio-transport plumbing is
invented (it does not exist); the id rides the same nudge that already carries
the `agent:`/`model:` instruction.

This is a small, decoupled seed — it lands as a **v0.3.x patch** (release
vehicle per DEC-027 / the PROJ-004 brief), ahead of and independent of the
v0.4.0 `brag impact` digest that will eventually read it.

## Goal

Extend the MCP `brag_add` provenance path to stamp three new **optional**
reserved tags — `session:<id>`, `cost:<n>`, `tokens:<n>` — via the existing
`stampProvenance` helper, with a defined value-normalization format (session:
opaque; cost: non-negative USD decimal; tokens: non-negative integer; bad
numerics rejected as a tool error), while (a) fabricating nothing (empty →
no tag), (b) keeping `store.go`'s `--author agent|human` classification
**unchanged** (`session:`/`cost:`/`tokens:` are NOT author-provenance tags),
and (c) surfacing `session_id` in the capture-nudge hook so Claude forwards it
— all migration-free, MCP-path-only.

## Inputs

- **Files to read:**
  - `internal/mcpserver/provenance.go` — `reservedTag(prefix, value)` and
    `stampProvenance(tags, agent, model)`. The seed extends `stampProvenance`
    to also stamp `session`/`cost`/`tokens`, and adds numeric validation
    helpers. `reservedTag`'s existing normalization (lowercase, whitespace→`-`,
    commas stripped) is reused for `session:` as-is; the numeric tags need a
    tighter rule (see Locked decisions).
  - `internal/mcpserver/server.go` — `addIn` struct (~50–61: add `Session`,
    `Cost`, `Tokens` optional params); `handleAdd` (~66–113: the length-cap
    guards + the `stampProvenance(in.Tags, agent, in.Model)` call the seed
    extends). Keep the existing pattern: validate, then stamp, then `Store.Add`.
  - `internal/storage/store.go` (308–322) — `provenanceExistsClause` (the
    `--author` classifier) and the `authorAgent`/`authorHuman` constants. **The
    load-bearing correctness point:** this clause is prefix-anchored to
    `t.name LIKE 'agent:%' OR t.name LIKE 'model:%'` and MUST stay that way —
    `session:`/`cost:`/`tokens:` must NOT be added to it (see Locked decision 3).
  - `internal/storage/store_test.go` (~595–655) — `TestList_FilterByAuthor`
    (SPEC-043): the existing guard that a topic tag like `agentic` (no colon)
    doesn't misclassify. The seed adds a sibling guard that a `session:`/`cost:`/
    `tokens:`-only entry doesn't misclassify as agent-authored.
  - `internal/mcpserver/provenance_test.go` / `server_test.go` — the mirror
    coverage to extend (empty→no tag, each field→its tag, order,
    normalization edge cases).
  - `plugin/hooks/capture-nudge.sh` — the Stop hook whose `additionalContext`
    gains the `session:<session_id>` instruction (silent-degradation +
    once-per-session contracts preserved).
  - `scripts/test-capture-nudge.sh` — the hook harness (H1–H7). H3 asserts
    `additionalContext` contains "brag" (case-insensitive) only; it does **not**
    pin the exact wording, so the seed's added sentence keeps it green.
  - `decisions/DEC-027-...` — the governing decision (emitted by this spec).
- **External APIs:** none new. `github.com/modelcontextprotocol/go-sdk` v1.6.1
  (already a dep, DEC-024). No network services; stdio only.
- **Related code paths:** `internal/mcpserver/`, `internal/storage/`
  (read-only — the classifier assertion), `plugin/hooks/`.

## Outputs

- **Files modified:**
  - `internal/mcpserver/provenance.go` — extend `stampProvenance` to accept and
    stamp `session`/`cost`/`tokens`; add `normalizeNumeric`-style validation for
    the numeric tags (see Locked decisions). `reservedTag` is reused unchanged
    for `session:`.
  - `internal/mcpserver/server.go` — add `Session`, `Cost`, `Tokens` optional
    fields to `addIn` (with `jsonschema` doc tags); in `handleAdd`, validate the
    numerics (reject non-numeric/negative as a tool error) and thread them into
    the extended `stampProvenance` call. No change to the milestone-free /
    cwd-free posture; `brag_add`'s **output** shape (DEC-011 object) is
    unchanged — the new params are inputs only.
  - `internal/mcpserver/provenance_test.go` — extend normalization + stamping
    tests to cover `session`/`cost`/`tokens` (see Failing Tests).
  - `internal/mcpserver/server_test.go` — add round-trip tests: each field
    stamps its tag + is `brag_list --tag`-filterable; bad numerics → tool error.
  - `internal/storage/store_test.go` — add `TestList_AuthorIgnoresSeedTags`
    (the regression guard: a `session:`/`cost:`/`tokens:`-only entry classifies
    as `human`, not `agent`).
  - `plugin/hooks/capture-nudge.sh` — the `additionalContext` gains one clause
    instructing Claude to forward `session:<session_id>` (and optional
    cost/tokens) on `brag_add`. The `$SESSION_ID` is already in scope.
  - `docs/api-contract.md` — extend the `### brag mcp serve` section's
    `brag_add` param list + reserved-namespace note with `session:`/`cost:`/
    `tokens:` and their formats.
  - `decisions/DEC-027-seed-cost-session-token-reserved-tags.md` (emitted at
    design).
- **New exports:** none. All additions are package-private in
  `internal/mcpserver` (`stampProvenance` gains params; a small numeric-
  normalizer is unexported). No new CLI surface (MCP-path-only — DEC-027
  Option D rejected).
- **Database changes:** **none. No migration.** The three tags ride the DEC-015
  tags/taggings join unchanged.

### Premise audit (run at design, per §9 — enumerate, don't discover at build)

**Signature change → existing `stampProvenance` callers/tests.** `stampProvenance`
gains parameters. Greps run at design (reconcile actual hits against this
enumeration before locking):

| Grep | Result | Verdict |
|---|---|---|
| `grep -rn "stampProvenance(" internal/` | 2 hits: the def + call in `server.go:98`; tests in `provenance_test.go` (`TestStampProvenance`) | **update all** — the call site gains 3 args; `TestStampProvenance`'s existing cases must be updated to the new arity (they are enumerated as planned rewrites here, not build-time discoveries — §9 inversion case). |
| `grep -rn "reservedTag(" internal/` | def + `TestReservedTag_Normalization` | `reservedTag` is **unchanged** (reused for `session:`); no edit. The numeric tags use a **new** normalizer, not `reservedTag`, so this test is untouched. |

**Author-classification premise (the subtle correctness point).** `store.go`'s
`provenanceExistsClause` and `TestList_FilterByAuthor` (SPEC-043) encode "agent-
authored iff a reserved `agent:`/`model:` tag exists." Adding new reserved
prefixes does **not** touch this — the clause stays `agent:%`/`model:%`-only.
Verified: the clause is prefix-anchored, the new prefixes are distinct, so a
`session:`-only entry is `human`. The seed **adds** a guard test rather than
changing the clause. (Grep: `grep -rn "agent:%\|model:%\|provenanceExists"
internal/storage/` → the single clause; no `session:`/`cost:`/`tokens:` present,
confirming they must stay absent.)

**Hook harness premise.** `scripts/test-capture-nudge.sh` H3 asserts
`additionalContext` matches `brag` (case-insensitive) and H1/H2/H4/H5/H6 assert
silent non-fire paths, H7 asserts `brag add` is never invoked. The seed edits
only the fire-path `additionalContext` **text** (adds a `session:` clause); it
does not touch any non-fire path, the marker/once-per-session logic, or the
"never runs brag" contract. So all seven assertions stay green. (Grep:
`grep -n "additionalContext\|SESSION_ID" plugin/hooks/capture-nudge.sh` →
`$SESSION_ID` already parsed at line 28; available for interpolation.)

**No command-count / no-new-dep / no-migration.** No new subcommand (MCP tool
params only), no new go.mod dependency (reuses DEC-024's SDK), no migration
(rides DEC-015). None of those premise-audit families fire.

**NOT-contains / reserved-namespace literal.** The numeric-normalization rule
(Locked decision 2) is a fixed-shape artifact: `cost:` = non-negative USD
decimal string; `tokens:` = non-negative integer; non-numeric/negative →
rejected. Build transcribes it verbatim; verify diffs against it.

## Acceptance Criteria

- [ ] `brag_add` accepts three new **optional** params — `session`, `cost`,
      `tokens` — alongside the existing `agent`/`model`. All optional; omitting
      any stamps no corresponding tag (parity with `agent`/`model` today).
- [ ] Given `session`, the stored tags include `session:<id>` (normalized via
      the existing `reservedTag` rule: lowercase, whitespace→`-`, commas
      stripped), appended after `agent:`/`model:` and canonicalized by
      `Store.Add` (DEC-015). `brag_list` with `tag: "session:<id>"` returns
      exactly the seed-tagged rows.
- [ ] Given a valid `cost` (non-negative decimal) and/or `tokens` (non-negative
      integer), the stored tags include `cost:<n>` and/or `tokens:<n>` in the
      normalized numeric form. `brag_list --tag` filters them.
- [ ] A non-numeric or negative `cost`/`tokens` is a **tool error**
      (`IsError=true` / non-nil handler error), never a silent insert and never
      a coerced tag. (No `cost:abc` ever lands.)
- [ ] **Append order is deterministic:** user tags, then `agent:`, then
      `model:`, then `session:`, then `cost:`, then `tokens:`.
- [ ] `store.go`'s `--author agent|human` classification is **unchanged**: an
      entry carrying only `session:`/`cost:`/`tokens:` tags (no `agent:`/
      `model:`) classifies as `human`. `provenanceExistsClause` remains
      `agent:%`/`model:%`-only.
- [ ] The capture-nudge hook's `additionalContext` instructs Claude to forward
      `session:<session_id>` (and optional cost/tokens) on `brag_add`; the
      hook's silent-degradation, once-per-session, and never-runs-`brag`
      contracts are intact. `scripts/test-capture-nudge.sh` (H1–H7) stays green.
- [ ] `brag_add`'s **output** (DEC-011 object) and the three read tools are
      unchanged; `brag_add` still emits no milestone and no cwd-project auto-fill.
- [ ] SQL stays in `internal/storage`; no new go.mod dep; no migration.
      `go test ./...`, `gofmt -l .`, `go vet ./...`, `CGO_ENABLED=0 go build
      ./...` clean; `scripts/test-docs.sh` and `scripts/test-capture-nudge.sh` OK.

## Failing Tests

Written during **design**, BEFORE build. Build makes these pass by extending
`internal/mcpserver/provenance.go` + `server.go`, adding the `store_test.go`
guard, and editing the hook. **No `time.Sleep` anywhere.**

### `internal/mcpserver/provenance_test.go` (pure — extend the existing file)

```go
// TestStampProvenance_SeedTags ▲ DEC-027 — session/cost/tokens are appended
// (in that order) after agent:/model:, each omittable; user tags preserved.
// NOTE: this asserts the NEW stampProvenance arity. The existing
// TestStampProvenance is rewritten to the new signature (see premise audit).
func TestStampProvenance_SeedTags(t *testing.T) {
	// signature: stampProvenance(tags, agent, model, session, cost, tokens string)
	if got := stampProvenance("perf", "claude-code", "claude-opus-4-8", "sess-abc", "0.42", "18000"); got !=
		"perf,agent:claude-code,model:claude-opus-4-8,session:sess-abc,cost:0.42,tokens:18000" {
		t.Errorf("all fields: %q", got)
	}
	if got := stampProvenance("", "", "", "sess-abc", "", ""); got != "session:sess-abc" {
		t.Errorf("session-only: %q", got)
	}
	if got := stampProvenance("perf", "claude-code", "", "", "", "18000"); got !=
		"perf,agent:claude-code,tokens:18000" {
		t.Errorf("skip empty middle fields: %q", got)
	}
	if got := stampProvenance("", "", "", "", "", ""); got != "" {
		t.Errorf("nothing → empty: %q", got)
	}
}

// TestNormalizeCost ▲ DEC-027 — non-negative USD decimal string; reject
// non-numeric / negative; trims; empty → ("", ok, no tag).
func TestNormalizeCost(t *testing.T) {
	ok := map[string]string{"0.42": "0.42", "12": "12", "  3.5 ": "3.5", "0": "0"}
	for in, want := range ok {
		got, err := normalizeCost(in)
		if err != nil || got != want {
			t.Errorf("normalizeCost(%q)=%q,%v want %q,nil", in, got, err, want)
		}
	}
	if got, err := normalizeCost(""); err != nil || got != "" {
		t.Errorf("empty cost → (\"\",nil), got %q,%v", got, err)
	}
	for _, bad := range []string{"abc", "-1", "1.2.3", "$5", "1e3"} {
		if _, err := normalizeCost(bad); err == nil {
			t.Errorf("normalizeCost(%q) expected error", bad)
		}
	}
}

// TestNormalizeTokens ▲ DEC-027 — non-negative integer; reject non-integer /
// negative; empty → ("", ok, no tag).
func TestNormalizeTokens(t *testing.T) {
	ok := map[string]string{"18000": "18000", " 0 ": "0", "42": "42"}
	for in, want := range ok {
		got, err := normalizeTokens(in)
		if err != nil || got != want {
			t.Errorf("normalizeTokens(%q)=%q,%v want %q,nil", in, got, err, want)
		}
	}
	if got, err := normalizeTokens(""); err != nil || got != "" {
		t.Errorf("empty tokens → (\"\",nil), got %q,%v", got, err)
	}
	for _, bad := range []string{"abc", "-5", "1.5", "1,000", "0x10"} {
		if _, err := normalizeTokens(bad); err == nil {
			t.Errorf("normalizeTokens(%q) expected error", bad)
		}
	}
}
```

### `internal/mcpserver/server_test.go` (round-trip — extend the existing file)

```go
// TestServer_AddStampsSeedTags ▲ DEC-027 — brag_add with session/cost/tokens
// stamps the reserved tags; brag_list --tag session:<id> finds the row; the
// stored tags carry all three in the locked order after agent:.
func TestServer_AddStampsSeedTags(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	callJSON(t, cs, "brag_add", map[string]any{
		"title": "cut p99", "tags": "perf",
		"session": "sess-abc", "cost": "0.42", "tokens": "18000",
	})
	rows, _ := s.List(storage.ListFilter{Tag: "session:sess-abc"})
	if len(rows) != 1 || rows[0].Title != "cut p99" {
		t.Fatalf("session tag not filterable: %+v", rows)
	}
	// agent: auto-fills from clientInfo.Name; seed tags follow it in order.
	if rows[0].Tags != "perf,agent:claude-code,session:sess-abc,cost:0.42,tokens:18000" {
		t.Errorf("stored tags = %q", rows[0].Tags)
	}
}

// TestServer_AddRejectsBadNumeric ▲ DEC-027 — non-numeric cost / tokens is a
// tool error, not a silent insert.
func TestServer_AddRejectsBadNumeric(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	for _, args := range []map[string]any{
		{"title": "x", "cost": "abc"},
		{"title": "x", "tokens": "-5"},
	} {
		r, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "brag_add", Arguments: args})
		if err != nil {
			t.Fatal(err)
		}
		if !r.IsError {
			t.Errorf("brag_add %v should be a tool error", args)
		}
	}
	if rows, _ := s.List(storage.ListFilter{}); len(rows) != 0 {
		t.Errorf("no row should be inserted on bad numeric, got %d", len(rows))
	}
}

// TestServer_SeedTagsOmittedWhenAbsent ▲ DEC-027 — omitting session/cost/tokens
// stamps no seed tag (parity with agent/model omission).
func TestServer_SeedTagsOmittedWhenAbsent(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	callJSON(t, cs, "brag_add", map[string]any{"title": "shipped"})
	rows, _ := s.List(storage.ListFilter{})
	if rows[0].Tags != "agent:claude-code" {
		t.Errorf("no seed tags expected, got %q", rows[0].Tags)
	}
}
```

### `internal/storage/store_test.go` (the load-bearing regression guard)

```go
// TestList_AuthorIgnoresSeedTags ▲ DEC-027 — session:/cost:/tokens: are
// reserved but NOT author-provenance tags. An entry carrying only these (no
// agent:/model:) classifies as "human", never "agent". Guards that
// provenanceExistsClause stays agent:%/model:%-only.
func TestList_AuthorIgnoresSeedTags(t *testing.T) {
	s := newStore(t)
	addWithTags(t, s, "seed-only", "session:sess-abc,cost:0.42,tokens:18000", "", "")
	addWithTags(t, s, "real-agent", "agent:claude-code,session:sess-abc", "", "")

	got, err := s.List(ListFilter{Author: "agent"})
	if err != nil {
		t.Fatalf("Author=agent: %v", err)
	}
	if len(got) != 1 || got[0].Title != "real-agent" {
		t.Errorf("Author=agent: want {real-agent} (seed-only is human); got %v", titlesOf(got))
	}
	got, _ = s.List(ListFilter{Author: "human"})
	if len(got) != 1 || got[0].Title != "seed-only" {
		t.Errorf("Author=human: want {seed-only}; got %v", titlesOf(got))
	}
}
```

**Fail-first map (§9):** `TestStampProvenance_SeedTags`, `TestNormalizeCost`,
`TestNormalizeTokens` fail to compile on current `main` (new arity / undefined
`normalizeCost`/`normalizeTokens`). `TestServer_AddStampsSeedTags` /
`TestServer_AddRejectsBadNumeric` fail because `addIn` has no
`session`/`cost`/`tokens` fields (the args are ignored → no tag / no error).
`TestList_AuthorIgnoresSeedTags` — this one **passes** on current `main`
already (the clause is already `agent:%`/`model:%`-only, so `seed-only` is
already `human`). That is **intentional**: it is a *regression guard* that must
stay green when the seed tags are added — it fails only if a build wrongly adds
`session:`/`cost:`/`tokens:` to `provenanceExistsClause`. Per §9's
"unexpectedly passing" note: confirm at build that this test would **fail** if
you added `OR t.name LIKE 'session:%'` to the clause (a 10-second mutation
check), proving it actually guards the decision.

## Implementation Context

*Read this section (and the files it points to) before starting the build
cycle. It is the equivalent of a handoff document, folded into the spec.*

### Decisions that apply

- `DEC-027` — **the governing decision** (emitted by this spec). The reserved-
  namespace extension, the "never fabricate / all optional" posture, the
  value-normalization formats, the author-classification isolation, and the
  MCP-only (no CLI flags) scope. Read in full before build.
- `DEC-024` — the path this **extends**: `stampProvenance`, the `agent:`/
  `model:` explicit-params-plus-`agent`-fallback, the stdio transport carrying
  no session/model identity (why `session:` is an explicit param, not
  auto-stamped). Not superseded.
- `DEC-015` — polymorphic tags/taggings. The three seed tags ride this join
  with **no schema change**; `Store.Add` canonicalizes the comma-joined string.
- `DEC-011` — the 9-key JSON entry shape. `brag_add`'s **output** is unchanged;
  the seed adds input params only.
- `DEC-012` — `brag add --json` schema. **Unchanged** — the seed is MCP-only
  (DEC-027 Option D rejected); do not add `--session`/`--cost`/`--tokens` CLI
  flags.
- `DEC-025` — the Claude Code plugin + capture-nudge hook. The hook's
  `additionalContext` gains the `session:` instruction; its contracts are
  preserved.

### Constraints that apply

- `stdout-is-for-data-stderr-is-for-humans` (**blocking**) — the tool errors for
  bad numerics are MCP `IsError` results / returned handler errors, never raw
  `os.Stdout` writes; the stdio transport purity DEC-024 generalized still holds.
- `no-sql-in-cli-layer` (**blocking**) — no `database/sql` under
  `internal/mcpserver/**`; the seed touches only the provenance/handler helpers.
- `no-cgo` (**blocking**) — no new dep; SDK already pure-Go.
- `test-before-implementation` (**blocking**) — the tests above are written
  first.
- `errors-wrap-with-context` (**warning**) — the numeric-validation errors wrap
  with context (`fmt.Errorf("brag_add: cost: %w", err)` or a clear message).
- **Not engaged:** `migrations-are-append-only` (no migration);
  `no-new-top-level-deps-without-decision` (no new dep).

### Prior related work

- `SPEC-040` (shipped) — the MCP server + `stampProvenance` this extends.
- `SPEC-041` (shipped) — the plugin + capture-nudge hook this edits; DEC-025.
- `SPEC-043` (shipped) — `brag list --author` + `TestList_FilterByAuthor`, the
  classifier this must leave untouched and adds a sibling guard to.

### Out of scope (for this spec specifically)

- **First-class cost/tokens/session columns + exact-token reconciliation**
  (join `session:<id>` → provider usage logs) — PROJ-005 (DEC-027 Option A /
  revisit trigger). The seed ships the reserved-tag *convention* only.
- **CLI `--session`/`--cost`/`--tokens` flags** — MCP-only (DEC-027 Option D
  rejected); `brag add --json` (DEC-012) is unchanged.
- **Any aggregation/reporting over cost/tokens** — `brag impact` (STAGE-011,
  separate spec) and PROJ-005 economics; the seed only *captures*.
- **bragfile computing/estimating tokens or cost** — it has no usage view
  (DEC-027 Option B rejected); all three inputs are caller-supplied, optional.
- **A milestone/nudge or cwd-project change at the MCP transport** — unchanged
  from SPEC-040; the seed touches only provenance stamping + the hook text.

## Notes for the Implementer

- **Extend `stampProvenance`, don't fork it.** New signature:
  `stampProvenance(tags, agent, model, session, cost, tokens string) string`.
  Append order is **user tags → agent: → model: → session: → cost: → tokens:**.
  `session:` reuses the existing `reservedTag("session", session)` (opaque id;
  same lowercase/whitespace/comma rule). `cost:`/`tokens:` use the **pre-
  validated, pre-normalized** numeric strings (see next bullet) wrapped as
  `"cost:"+c` / `"tokens:"+tk` only when non-empty — do NOT run them through
  `reservedTag` (its whitespace→`-` / lowercase rules are meaningless for a
  validated number and would mask a bug). Sketch:
  ```go
  func stampProvenance(tags, agent, model, session, cost, tokens string) string {
      toks := splitTags(tags) // existing trim/drop-empty loop
      if a := reservedTag("agent", agent); a != "" { toks = append(toks, a) }
      if m := reservedTag("model", model); m != "" { toks = append(toks, m) }
      if sv := reservedTag("session", session); sv != "" { toks = append(toks, sv) }
      if cost != "" { toks = append(toks, "cost:"+cost) }   // cost already validated+normalized
      if tokens != "" { toks = append(toks, "tokens:"+tokens) }
      return strings.Join(toks, ",")
  }
  ```
- **Validate the numerics at the handler boundary, before stamping.**
  `normalizeCost(raw) (string, error)`: trim; empty → `("", nil)`; else parse as
  a non-negative decimal (`strconv.ParseFloat` on a `[0-9]+(\.[0-9]+)?`-shaped
  string — reject scientific notation, currency symbols, thousands separators,
  and negatives explicitly; return the trimmed canonical string, not a
  re-formatted float, to avoid `0.42`→`0.42000001` drift). `normalizeTokens(raw)
  (string, error)`: trim; empty → `("", nil)`; else `strconv.ParseUint`
  (rejects negatives, decimals, non-digits); return the trimmed string. In
  `handleAdd`: call both, return a tool error on either error (wrap with
  context), then pass the normalized strings into `stampProvenance`.
- **The `session` param needs no length cap** in the numeric sense, but keep the
  existing tags-input length discipline: the caller's own `tags` cap
  (`len(in.Tags) > 64`) is on the **user input**, unchanged — do NOT cap the
  stamped result (provenance is appended after the cap, exactly like today).
  A `session:` id is a reserved tag, not user `tags`, so it is not counted
  against that cap (matching how `agent:`/`model:` aren't).
- **`addIn` additions** (mirror the existing `agent`/`model` doc-tag style):
  ```go
  Session string `json:"session,omitempty" jsonschema:"stable per-session id; stamped as session:<id> — the PROJ-005 reconciliation join-key (no transport fallback; forward the hook-surfaced session_id)"`
  Cost    string `json:"cost,omitempty" jsonschema:"caller-reported cost in USD (non-negative decimal, e.g. 0.42); stamped as cost:<n>. Never estimated by bragfile."`
  Tokens  string `json:"tokens,omitempty" jsonschema:"caller-reported total tokens (non-negative integer); stamped as tokens:<n>. Never estimated by bragfile."`
  ```
  Keep them **strings** (not numeric JSON types) so the normalizer owns the
  format and an omitted field is unambiguously `""` — matching the existing
  all-string `addIn` shape and its validation posture.
- **Do NOT touch `store.go`'s `provenanceExistsClause`.** It must stay
  `agent:%`/`model:%`-only. The `TestList_AuthorIgnoresSeedTags` guard proves
  this; do the 10-second mutation check (add `OR t.name LIKE 'session:%'`,
  confirm the guard fails, revert) to prove the test bites (§9 unexpectedly-
  passing note).
- **The hook edit** (`plugin/hooks/capture-nudge.sh`): `$SESSION_ID` is already
  parsed (line 28). Extend the `additionalContext` string's provenance sentence
  to also tell Claude to pass the session id. Keep it one nudge, fire-path-only,
  and keep the word "brag" in the text (H3). Exact literal — see Locked
  decision 4. Do **not** touch any non-fire path or the marker logic.
- **Fail-first check (build step):** the pure + server tests won't compile until
  `stampProvenance`'s arity changes and `normalizeCost`/`normalizeTokens`/the
  `addIn` fields exist. `TestList_AuthorIgnoresSeedTags` passes immediately (it
  guards, doesn't drive) — run the clause mutation to confirm it bites.

## Locked design decisions

Each behavior decision (1–4) has ≥1 paired test that fails without it (§9),
except decision 3 whose paired test is a *regression guard* (see its note).

1. **Three optional reserved tags via the extended `stampProvenance`
   (DEC-027).** `session:<id>` / `cost:<n>` / `tokens:<n>`, all optional
   (empty → no tag), appended after `agent:`/`model:` in the order
   session→cost→tokens. *Pair (▲):* `TestStampProvenance_SeedTags`,
   `TestServer_AddStampsSeedTags`, `TestServer_SeedTagsOmittedWhenAbsent`.
   - **Rejected alternatives:** first-class columns now (DEC-027 Option A —
     needs a migration; seed is migration-free); bragfile-estimated numbers
     (Option B — no usage view; never fabricate); transport-derived session
     (Option C — impossible, stdio carries no session id).

2. **Numeric normalization: `cost:` = non-negative USD decimal string;
   `tokens:` = non-negative integer; bad numerics rejected as a tool error
   (the §12 literal).** `session:` is opaque (existing `reservedTag` rule).
   Non-numeric / negative / scientific / currency-symbol / thousands-separator
   `cost`/`tokens` → tool error, never a coerced tag. Return the trimmed
   canonical string (no float re-formatting). *Pair (▲):* `TestNormalizeCost`,
   `TestNormalizeTokens`, `TestServer_AddRejectsBadNumeric`.

3. **Author classification is unchanged — `session:`/`cost:`/`tokens:` are NOT
   author-provenance tags.** `provenanceExistsClause` stays `agent:%`/`model:%`-
   only; a seed-tag-only entry is `human`. *Pair (▲, regression guard):*
   `TestList_AuthorIgnoresSeedTags`.
   - **Note:** this test passes on current `main` (the clause is already
     correctly scoped). It is a guard, not a driver — it fails only if a build
     wrongly widens the clause. Build must run the clause-mutation check to
     confirm the guard bites (§9).

4. **The capture-nudge hook surfaces `session_id` for Claude to forward (the
   §12 literal).** The fire-path `additionalContext` gains one clause telling
   Claude to pass `session:<session_id>` (and optional cost/tokens) on
   `brag_add`; silent-degradation, once-per-session, and never-runs-`brag`
   contracts preserved. *Validated by:* `scripts/test-capture-nudge.sh`
   (H1–H7 stay green — H3's "brag" substring holds; the added clause does not
   touch non-fire paths). The literal `additionalContext` text is embedded
   below.

### Locked hook literal (build transcribes verbatim)

The fire-path `jq` block's `additionalContext` becomes (extending the existing
provenance sentence with the session clause — the `\($SESSION_ID)` is
interpolated by `jq` from the already-parsed shell var via `--arg`):

> "A commit landed during this session. If something brag-worthy shipped, draft
> a brag entry for the user's approval per BRAG.md (you can use the /brag:brag
> command): a required action-verb title plus optional project, type, tags, and
> a concrete impact. Stamp provenance as reserved tags agent:<name> and
> model:<id>, and pass session:<id> using this session's id (\(session)) so the
> work is joinable later; include cost:<usd> and tokens:<n> only if you have
> real figures — never estimate them. Do NOT run `brag add` until the user
> explicitly approves."

Build detail: pass the id into `jq` safely with `--arg session "$SESSION_ID"`
(rather than shell-interpolating into the `jq` program), keeping the existing
`jq -cn` invocation shape. The exact phrasing is design-decidable; the
load-bearing parts are (a) the word "brag" survives (H3), (b) it names
`session:<id>` and the real-figures-only caveat, (c) no non-fire path changes.

### S-sizing rationale

SPEC-046 is **S**. It is a bounded extension of an already-shipped, already-
tested path: ~1 line per new tag in `stampProvenance`, two small numeric
normalizers, three `addIn` fields, three handler validation calls, one guard
test in `store_test.go` (which already passes — it just needs to exist), a
one-sentence hook edit, and a docs paragraph. No new package, no new command,
no new dependency, no migration, no new transport surface. The one genuine
design decision (numeric format + the author-classification isolation) is
resolved here in prose + paired tests, per the §12(b) "resolve at design"
discipline.

## Build Completion

*Filled at build (this cycle). The design's plan held; nothing was
re-litigated. No new DEC was needed and no `questions.yaml` entry was raised.*

### What was implemented

- **`internal/mcpserver/provenance.go`** — `stampProvenance` gained three
  trailing params (`session, cost, tokens`) appended in the locked order
  `agent: → model: → session: → cost: → tokens:`; `session:` reuses
  `reservedTag`, while `cost:`/`tokens:` are appended verbatim from the
  pre-validated strings. Added `normalizeCost` (plain non-negative decimal;
  rejects negatives / scientific / currency / thousands separators via a
  tighter `isDecimal` guard run before `strconv.ParseFloat`), `normalizeTokens`
  (`strconv.ParseUint`), and the `isDecimal` helper. Errors wrap with context.
- **`internal/mcpserver/server.go`** — `addIn` gained `Session`/`Cost`/`Tokens`
  optional string fields (mirroring the `agent`/`model` doc-tag style);
  `handleAdd` validates the numerics at the boundary (bad → tool error, no
  insert) and threads all three into the extended `stampProvenance` call. The
  DEC-011 output shape, milestone-free / cwd-free posture, and length caps are
  unchanged (the seed params carry no caps — provenance is appended after the
  `tags` cap, exactly like `agent:`/`model:`).
- **`plugin/hooks/capture-nudge.sh`** — the fire-path `additionalContext` now
  passes the already-parsed `$SESSION_ID` into `jq` via `--arg session` and
  instructs Claude to forward `session:<id>` (plus real-figures-only
  `cost:`/`tokens:`). Silent-degradation, once-per-session, and never-runs-`brag`
  contracts untouched; the word "brag" survives (H3).
- **`docs/api-contract.md`** — the `brag mcp serve` section documents the three
  seed params, their formats, the tool-error-on-bad-numeric rule, and the
  author-classification isolation.
- **Tests added** (all pass): `TestStampProvenance_SeedTags`, `TestNormalizeCost`,
  `TestNormalizeTokens` (provenance_test.go); `TestServer_AddStampsSeedTags`,
  `TestServer_AddRejectsBadNumeric`, `TestServer_SeedTagsOmittedWhenAbsent`
  (server_test.go); `TestList_AuthorIgnoresSeedTags` (store_test.go). The
  existing `TestStampProvenance` was rewritten to the new arity (enumerated in
  the premise audit, not a build-time surprise).

### Fail-first + guard verification

- The pure + server tests failed to compile on the pre-build tree (new arity /
  undefined `normalizeCost`/`normalizeTokens` / missing `addIn` fields), then
  passed after implementation — the §9 fail-first order held.
- `TestList_AuthorIgnoresSeedTags` passed immediately (it is a regression
  guard, not a driver). The §9 mutation check was run: adding
  `OR t.name LIKE 'session:%'` to `provenanceExistsClause` **failed** the guard
  (`seed-only` mis-classified as agent), and the change was reverted. The guard
  bites; the clause stays `agent:%`/`model:%`-only.

### Gate results (all exit 0)

`go test ./...` (588 passed), `gofmt -l .` (empty), `go vet ./...` (no issues),
`CGO_ENABLED=0 go build ./...` (success), `just test-docs` (OK), `just test-hook`
(H1–H7 OK).

### Honest reflection

The spec was unusually precise — the sketch in Notes for the Implementer, the
locked hook literal, and the paired tests left almost no build-time judgment.
The only deviation from a literal transcription was a defensive one: the spec's
prose said "parse as a non-negative decimal … reject scientific notation,
currency symbols, thousands separators, and negatives explicitly," and
`strconv.ParseFloat` alone accepts `1e3`, `+1`, leading `.`/trailing `.`, and
leading whitespace, so I added the small `isDecimal` shape-guard to run first.
That is exactly the `TestNormalizeCost` bad-input set (`1e3`, `$5`, `-1`,
`1.2.3`), so the guard is test-driven, not speculative. No genuinely new
decision arose, so no DEC-028 and no `questions.yaml` stop. The
author-classification isolation is the subtle load-bearing point and the
mutation check confirms the guard actually protects it.

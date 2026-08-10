---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-075
  type: story                      # epic | story | task | bug | chore
  cycle: ship
  blocked: false
  priority: high
  complexity: M                    # S | M | L  (L means split it)

project:
  id: PROJ-006
  stage: STAGE-018
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-09

references:
  decisions:
    - DEC-046                      # emitted by this spec
    - DEC-044                      # re-derived here (worst-case line, budget model)
    - DEC-015                      # tags normalised into a table — no length constraint
    - DEC-012                      # introduced the caps, records no rationale
    - DEC-004                      # comma-joined tags — the era MaxTags belongs to
    - DEC-036                      # soft-match invariant cli/project.go:162 reasons about
    - DEC-017                      # project is a soft match, not a foreign key
    - DEC-027                      # reserved cost:/tokens: tags Validate already checks
  constraints:
    - test-before-implementation
    - no-sql-in-cli-layer
    - stdout-is-for-data-stderr-is-for-humans
  related_specs:
    - SPEC-064                     # unified per-path validation (created the problem)
    - SPEC-073                     # memory slice / DEC-044
    - SPEC-074                     # MCP push surface (inherits the line shape)
---

# SPEC-075: capture validation caps

## Context

`internal/capture.Validate` enforces six byte caps across every capture ingress
path. **None of them was derived.** DEC-012 introduced them for the
`brag add --json` schema and records no rationale for any value; SPEC-064 then
correctly unified per-path validation — flag and editor mode had been skipping
the caps that `--json` and MCP enforced — and in doing so generalised those
numbers to every path. The side effect nobody measured: a large slice of the
**existing** corpus became unwritable.

The full measurement, reasoning, and ordering risk live in
`guidance/questions.yaml` → `capture-caps-were-inherited-not-derived`. That
question is this spec's spine and is not re-derived here. This spec resolves it.

This is the **v0.6.0 gate** (STAGE-018). STAGE-019 shipped the memory slice,
whose per-entry line ends in `<impact>`, and DEC-044 derived its entire budget
model — including the anti-starvation argument — from the *current* caps.
Cutting v0.6.0 first would ship that model knowing the caps are wrong, then
change how many entries fit a 2000-token budget one release later.

**Design-time re-measurement.** The question's figures were taken at 347
entries. Design re-measured against a **359-entry snapshot** (the live corpus is
appended to continuously; the snapshot makes these numbers reproducible). Two
findings changed the shape of the fix, and three pre-existing errors in DEC-044
surfaced — see LD2 and LD5. Where the two measurements disagree, both are
recorded; neither silently replaces the other.

## Goal

Re-shape `MaxImpact`, `MaxTitle` and the tags cap so each is derived from what
its field is for; wire `capture.Validate` onto the edit path without making the
already-over-cap corpus uneditable; and re-derive DEC-044's budget model against
the new caps rather than leaving it silently invalid.

## Inputs

- **Files to read:**
  - `guidance/questions.yaml` (`capture-caps-were-inherited-not-derived`) — the spine.
  - `internal/capture/validate.go` — the caps and `Validate`.
  - `decisions/DEC-044-…` — Consequences + the budget math this spec re-derives.
  - `decisions/DEC-015-…`, `DEC-004`, `DEC-012`, `DEC-036`, `DEC-017`.
  - `internal/cli/add.go` (×2), `internal/cli/add_json.go`, `internal/mcpserver/server.go` — the ingress paths that call `Validate` today.
  - `internal/cli/edit.go`, `internal/storage/store.go` (`Update`) — the path that does not.
- **Related code paths:** `internal/capture/`, `internal/cli/`, `internal/mcpserver/`, `internal/memory/`.
- **External APIs:** none.

## Measurement (design-time, reproducible)

Snapshot: 359 entries, provenance tags (`agent:`/`model:`/`session:`/`cost:`/
`tokens:`) stripped so only user-supplied text counts. Byte counts (`len`),
matching `Validate`.

| field | cap | p50 | p90 | p95 | p99 | max | over cap | verdict |
|---|---|---|---|---|---|---|---|---|
| `impact` | 256 | 228 | 413 | 514 | 932 | 1290 | **74 of 285 (26%)** | re-shape — the sharpest |
| `title` | 200 | 73 | 134 | 541 | 1067 | 1444 | **33 of 359** | re-shape, but see LD2 |
| `tags` (joined) | 64 | 27 (mean) | — | — | — | 89 | **7 of 260** | re-shape, and the wrong shape |
| `project` | 64 | 9 | — | — | — | 22 | 0 | leave (3× headroom) |
| `type` | 64 | 7 | — | — | — | 10 | 0 | leave (6× headroom) |
| `description` | 100000 | 570 | — | — | — | 2071 | 0 | leave (48× headroom) |

Two findings the 347-entry measurement did not separate:

1. **`title`'s over-cap population is not ordinary usage.** All 33 over-cap
   titles belong to a **single project** (`zany-animal-slots`, ids 157–228) — one
   malformed-capture episode where whole brag bodies were written into the
   `title` field. Excluding it, `title` p90 = 134 against a 200 cap. `impact` is
   the opposite: p50 = 228 against a 256 cap (89% of the ceiling), 109 entries
   piled into `[200, 256]`, and over-cap entries spread across **seven**
   projects. One field is genuinely under-provisioned; the other has a data-entry
   bug in its tail.
2. **The tags cap fires on count, never on length.** Every one of the 7 over-cap
   tag strings carries **7–9 individually short tags** (longest single tag among
   them: 17 bytes). The longest single tag in the entire corpus is 21 bytes
   (`model:claude-opus-4-8`); p99 = 18. The cap is exactly inverted from the
   abuse it should catch.

**Delta against the question's figures, recorded not reconciled.** The question
records 81 of 274 over-cap impacts; design measures 74 of 285. The corpus grew
by 12 while the over-cap count fell by 7, and only 3 entries have ever been
edited (`updated_at <> created_at`), so growth alone does not explain it. Both
figures are kept. The grandfathering math below uses the design-time snapshot
because it is reproducible; the discrepancy is not load-bearing for any locked
decision (every option below is chosen on distribution shape, not on the exact
over-cap count).

## Locked design decisions

### LD1 — `MaxImpact`: 256 → **1024**

`impact` is the one-line outcome: before → after, with a confirming number
(the shape a spec's Reflection-Q4 is transcribed into). It is the single field
`brag impact`, `wrapped`, `story`, `coverage` and `memory` all read.

Derivation: 1024 covers measured p99 (932) with headroom, and is **4× the
field's median** — restoring the headroom class `project` (7×) and `type` (9×)
already have and that `impact` uniquely lacked at 1.12×. It is bounded above by
LD5's memory-line argument, which is what stops "widen it" becoming "widen it a
lot". Residual over-cap after the change: **3 entries**.

### LD2 — `MaxTitle`: 200 → **256**

`title` is a headline. Legitimate-use p90 is 134; 256 gives that distribution
~2× headroom and keeps a title inside two 120-column terminal lines.

**Deliberately not widened to fit the tail.** Admitting the 1444-byte outliers
would require ~1536 and would enshrine a capture bug — whole bodies pasted into
`title` — as the contract. The 29 residual over-256 rows are grandfathered by
LD4 and are forced to be fixed if and only if someone edits the title itself,
which is the correct outcome for a field holding a pasted paragraph.

This is a smaller move than the raw "33 of 348 over cap" figure suggests, and
LD2 is the reason: the finding is about *which* entries are over, not how many.

### LD3 — `MaxTags` (joined string) → `MaxTagLen` + `MaxTagCount`

Delete `MaxTags`. Replace with:

- **`MaxTagLen = 64`** — per tag. Longest tag in the corpus is 21 bytes
  (p99 = 18), so 64 is 3× headroom, derived exactly as `project`/`type` are: a
  tag is an identifier-like label, not prose. This is the cap that catches the
  actual abuse (one absurd tag).
- **`MaxTagCount = 32`** — per entry. Max observed is 11 including provenance,
  9 user-supplied; 32 is ~3× headroom.

DEC-015 normalised tags into a `tags` table (`name TEXT NOT NULL UNIQUE`, no
length constraint), so a joined-string cap is a DEC-004-era leftover with no
model behind it. Residual over-cap after the change: **0 entries**.

**The count cap applies pre-stamp.** `Validate` runs on caller-supplied text
before provenance stamping (the `Fields` doc contract). Up to 5 reserved tags
(`agent:`/`model:`/`session:`/`cost:`/`tokens:`) are stamped afterwards and must
not be able to push a legitimate entry over the count cap — 32 against a
9-tag observed maximum leaves that impossible in practice, and the ordering
makes it impossible by construction.

**Neither cap bounds stamped provenance — say so, do not let a reader infer
otherwise.** Because `Validate` runs pre-stamp, `MaxTagLen` never applies to a
`model:`/`agent:`/`session:` value supplied through its dedicated param; there is
no length check on those anywhere (only `cost`/`tokens` get numeric
normalisation). The same-looking token is capped or not depending on which door
it came through: `model:x` in the freeform `tags` field is validated as an
ordinary tag; `model:x` via the param is not. A 500-byte model id is storable
today and still will be after this spec.

This is deliberate and out of scope here — bounding what a caller asserts about
itself belongs to signed provenance (PROJ-006 #2). It matters for the build
because it settles a question that will come up: **`MaxTagLen = 64` does not
need to accommodate long model identifiers**, since it never sees them. 64 is
derived from the 21-byte observed maximum of the tags this cap actually governs.
Do not widen it on provenance grounds.

`project` / `type` / `description` are **not** touched. Their headroom was
measured and is ample; widening them would be the reflexive uniform multiplier
this spec exists to avoid.

### LD4 — Grandfathering: **validate on write-of-changed-field**

On the edit path, a field is validated **only if its value differs from the
stored value**. Nothing is validated that the user did not touch.

Properties: no schema change, no migration, no permanent second class of row;
it **converges** (any field you touch is brought to contract); and the add path
is unaffected — a new entry always validates every field.

Residual over-cap rows after LD1–LD3: **3 impacts, 29 titles, 0 tag
violations.** Each stays editable in every field except the offending one, and
editing the offending one forces a fix.

**Rejected alternatives.**
- *Caps large enough that everything fits* (impact ~1536, title ~1536). Rejected:
  it sets the contract from the corpus's worst data-entry accident rather than
  from what the field is for, and LD2's finding is precisely that the tail is
  not legitimate usage.
- *A row-level grandfather flag / `validated_at` column*. Rejected: a schema
  change plus a migration to encode "these rows are exempt forever", which never
  converges and adds a permanent second class of row to every read path.
- *Validate nothing on edit (status quo)*. Rejected: it is the open backlog item,
  and it leaves edit as the one ingress path where the contract does not hold.

### LD5 — DEC-044 re-derivation

DEC-044 cites the caps twice (lines 136 and 443) for the memory slice's
worst-case line. The line shape (DEC-044 sub-decision 5) is:

```
- <id> <YYYY-MM-DD> [<project>/<type>] <title> — <impact>
```

**Byte budget, term by term** (`—` is 3 bytes UTF-8; `<id>` at int64 max is 19
digits):

| term | bytes |
|---|---|
| `- ` | 2 |
| `<id>` | 19 |
| ` ` + date + ` ` | 12 |
| `[` + project + `/` + type + `]` | 131 |
| ` ` | 1 |
| `<title>` | **256** (was 200) |
| ` — ` | 5 |
| `<impact>` | **1024** (was 256) |
| **total** | **1450** (was 626) |

`EstimateTokens` = `ceil(1450/4)` = **363 tokens** (was `ceil(626/4)` = 157).
The old terms reproduce DEC-044's published 626 / 157 exactly, which confirms
the 19-digit id assumption the DEC left implicit.

**Anti-starvation argument, restated against the new caps.** At 363 tokens a
worst-case entry consumes 18.2% of the 2000-token default, leaving 82% for the
rest of the slice; and DEC-044 sub-decision 3's skip-and-continue means an entry
that does not fit is skipped, not fatal. Catastrophic starvation remains
impossible. The argument survives the change intact.

**Three pre-existing errors in DEC-044 that the re-derivation surfaced.** All
three are independent of this spec's cap change — they are wrong today — but
this spec is where the math is re-derived, so leaving them is exactly the silent
invalidation STAGE-018's success criteria forbid.

- **(i) The 626-byte bound is already false of the live corpus.** Measured
  maximum rendered line today is **1483 bytes / 371 tokens** (entry 172),
  because the over-cap rows predate SPEC-064's unification. So the new
  theoretical bound of 1450 bytes is **below today's observed maximum**: this
  spec *tightens* the true bound while widening the caps that bite in ordinary
  use.
- **(ii) "a real corpus averages ~18 tokens per entry" is the test fixture's
  mean, not a corpus mean.** DEC-044 sub-decision 3's own 8-entry fixture has
  per-entry costs 27, 14, 15, 10, 27, 15, 22, 13 → mean **17.875 ≈ 18**. The
  measured corpus mean is **~92 tokens/entry**.
- **(iii) "2000 buys ≈110 entries" is wrong by ~4.4×.** Measured against the
  real corpus, `brag memory` at the default budget reports
  **`Included: 25`, `Skipped: 175`, `Estimated: 1991`**.

**`memory.DefaultBudget = 2000` HOLDS — the number stays, its stated rationale
is corrected.** The default's defense is *displacement*: 2000 is ~10% of the
measured ~20k-token session-opening ritual it exists to replace, and ~1% of a
200k context window. This spec changes neither. What is false is the *yield*
claim attached to it (≈110 entries), and a yield of 25 best-ranked entries is
still a useful slice. Changing a shipped default is a behavior change to a
one-release-old feature and belongs in its own spec; DEC-044's Revisit-(b)
already names the trigger and the calibration data is printed in every envelope.
Correcting the rationale here is what stops a future reader re-deriving from a
false number.

### LD6 — One constant, not two

`internal/cli/project.go:162` hard-codes `if len(name) > 64` with the message
`"project name exceeds 64-character limit"`, directly beneath a comment
reasoning about keeping it in agreement with the capture cap (DEC-036's
soft-match invariant: an ensured name must be able to soft-match a normally
added entry, both counting bytes — DEC-017). Replace the literal with
`capture.MaxProject` and derive the number in the message from the constant.

`MaxProject` does not move in this spec, so this is behavior-preserving by
construction — which is exactly why it must land *with* the caps change rather
than before it.

### LD7 — Edit-path wiring, ordered strictly after LD1–LD4

Add `capture.ValidateChanged(old, new Fields) error`: validates a field only
when `old` and `new` differ for that field, then applies the same rules
`Validate` does. Call it from `internal/cli/edit.go` before `s.Update`, wrapped
in `UserErrorf` like every other CLI ingress.

**Not in `Store.Update`.** SPEC-064 put ingress validation in `capture`, called
by the boundary layers; storage takes already-validated typed values. Validating
inside `Update` would also validate internal callers (restore, future import)
that legitimately carry grandfathered text, and would need storage to re-read
the row to learn what changed.

There is no MCP edit tool — the server advertises five tools
(`brag_add`/`brag_list`/`brag_search`/`brag_stats`/`brag_memory`) — so
`cli/edit.go` is the only unwired path, and this makes five `capture` call sites
in total.

**Ordering is load-bearing.** LD7 is safe only because LD4 exists. Reversed, it
makes ~30% of the corpus uneditable — a user could not fix a title without
truncating impact or deleting tags. There is no ordering in which that is
acceptable.

## Outputs

- **Files created:**
  - `decisions/DEC-046-capture-field-caps-derived-and-the-edit-path.md` — the validation-contract change + the DEC-044 re-derivation.
  - `internal/capture/validate_changed_test.go` — LD4 tests.
- **Files modified:**
  - `internal/capture/validate.go` — `MaxTitle` 200→256, `MaxImpact` 256→1024; delete `MaxTags`; add `MaxTagLen = 64`, `MaxTagCount = 32`; per-tag + count validation replacing the joined-string check; add `ValidateChanged`. Update the package/const doc comments (they currently assert the caps are "identical to the caps `add --json` and MCP brag_add have always enforced").
  - `internal/cli/edit.go` — call `capture.ValidateChanged` before `s.Update` (LD7).
  - `internal/cli/project.go` — line 162 references `capture.MaxProject` (LD6).
  - `internal/capture/validate_test.go` — **premise audit, planned rewrite:** the all-at-cap fixture (`:12–17`) references `MaxTags`, which is deleted; the over-cap table (`:32–37`) asserts literal `"title" exceeds 200-character limit`, `"tags" exceeds 64-character limit`, `"impact" exceeds 256-character limit`. Three literals change and the `tags` row is *deleted* (different shape, not a different number).
  - `internal/cli/add_json_test.go` — **premise audit, planned rewrite:** `:488–513` duplicates the same six literal error substrings. Same three changes, same `tags` deletion. *(This file was found only by grepping the identifier repo-wide, not by grepping `internal/capture` — see the audit-grep reconciliation below.)*
  - `internal/cli/edit_test.go` — LD7 + LD4 tests (T11, T12).
  - `internal/cli/project_test.go` — T13 (no existing project-name-cap test).
  - `internal/memory/` (test only) — T14, pinning the LD5 re-derivation.
  - `docs/brag-entry.schema.json` — `title.maxLength` 200→256; `impact.maxLength` 256→1024; **remove `tags.maxLength`** and describe the per-tag + count rule in the field description (a joined-string `maxLength` can no longer express the contract).
  - `docs/for-ai-agents.md:71–76` — the `brag_add` param table: `title` ≤256, `impact` ≤1024, `tags` re-worded from "≤64 characters" to the per-tag + count shape.
  - `docs/api-contract.md:1235–1237` — same three values in the `brag mcp serve` tool description.
  - `decisions/DEC-044-…` — lines 136 and 443 (the 626/157 citations) and lines 71–74 (the ~18-tokens / ≈110-entries claim), corrected per LD5 with a pointer to DEC-046.
  - `guidance/questions.yaml` — `capture-caps-were-inherited-not-derived` → `status: resolved`, resolution pointing at SPEC-075 / DEC-046.
  - `projects/PROJ-006-…/stages/STAGE-018-…md` — Spec Backlog line for SPEC-075.
- **New exports:** `capture.MaxTagLen`, `capture.MaxTagCount`, `capture.ValidateChanged(old, new Fields) error`. **Removed export:** `capture.MaxTags`.
- **Database changes:** none. LD4 is chosen partly because it needs none.

### Audit-grep cross-check (§9 — run at design, reconciled here)

Per §9's status-change heuristic, greps target the **feature name** (the cap
identifiers and the field names), never a phrasing of the current number.

| grep | hits | reconciliation |
|---|---|---|
| `grep -rn "MaxTitle\|MaxDescription\|MaxTags\|MaxProject\|MaxType\|MaxImpact" --include="*.go" --include="*.md" --include="*.yaml" .` | `validate.go`, `validate_test.go`, `questions.yaml`, `STAGE-018`, this spec | all enumerated in Outputs |
| `grep -rn "exceeds 200\|exceeds 64\|exceeds 256\|exceeds 100000" --include="*_test.go" internal/` | `validate_test.go:32–37`, **`add_json_test.go:488–513`** | **the second file was missed by a package-scoped grep** — added to Outputs |
| `grep -rn "≤200\|≤256\|≤64\|maxLength\|100000" docs/ README.md BRAG.md AGENTS.md plugin/` | `brag-entry.schema.json`, `for-ai-agents.md`, `api-contract.md`, `docs/reports/security/2026-04-26-…` | first three in Outputs; the security report is a **dated historical record** and is deliberately NOT updated |
| `grep -rn "626\|157 tok" decisions/ guidance/` | `DEC-044:136`, `DEC-044:443`, `questions.yaml:602` | all enumerated in Outputs |
| `grep -rn "capture.Validate" --include="*.go" internal/` | `add.go:143`, `add.go:203`, `add_json.go:74`, `server.go:98` | the four current ingress paths; `edit.go` becomes the fifth (LD7) |

`BRAG.md` and `README.md` were swept and carry **no** cap numbers — a recorded
negative, so build does not re-sweep them.

## Acceptance Criteria

- [x] `MaxTitle == 256`, `MaxImpact == 1024`; `MaxProject`/`MaxType`/`MaxDescription` unchanged.
- [x] `MaxTags` no longer exists; `MaxTagLen == 64` and `MaxTagCount == 32` do.
- [x] A 9-tag comma-joined string longer than 64 bytes **validates successfully** (the reported regression).
- [x] A single tag longer than 64 bytes is rejected, and the error names the offending tag.
- [x] More than 32 tags is rejected with a count-shaped error.
- [x] `capture.ValidateChanged` passes an unchanged over-cap field and rejects a changed one.
- [x] `brag edit` on an entry with a grandfathered over-cap `impact` succeeds when only `title` changes, and the stored `impact` is byte-identical afterwards.
- [x] `brag edit` rejects a changed field that violates a cap, with no database write.
- [x] `cli/project.go` contains no numeric literal for the project cap; a `MaxProject`-length name is accepted and `+1` rejected.
- [x] The worst-case memory line is exactly 1450 bytes / 363 estimated tokens.
- [x] `memory.DefaultBudget` is unchanged at 2000.
- [x] DEC-044's two 626/157 citations and its ~18-tokens/≈110-entries claim are corrected.
- [x] Every doc site enumerated in Outputs is updated; the dated security report is not.
- [x] Full gate set green: `go test ./...`, `gofmt -l .` empty, `go vet ./...`, `scripts/test-docs.sh`.

## Failing Tests

Written during **design**, BEFORE build. Each locked decision has at least one
test that fails without it (§9). Each entry states *only* what a mutation of the
named decision would break — no broader coverage claim is made than the
assertion supports.

- **`internal/capture/validate_test.go`**
  - `"impact at 1024 passes, 1025 fails"` — asserts `Validate(Fields{Impact: 1024×"i"}) == nil` and that 1025 returns an error whose text contains `1024`. Fails if `MaxImpact` is any value other than 1024. *(LD1)*
  - `"title at 256 passes, 257 fails"` — same shape against `MaxTitle`. Fails if `MaxTitle` ≠ 256. *(LD2)*
  - `"a single tag over 64 bytes is rejected and the error names that tag"` — `Tags: "ok," + 65×"z"` returns an error containing the 65-byte token. Fails if the per-tag cap is absent, and fails if the error reports the joined string instead of the tag. *(LD3)*
  - `"33 tags are rejected, 32 pass"` — `Tags` = 32 one-byte tags passes; 33 returns an error. Fails if `MaxTagCount` is absent or ≠ 32. *(LD3)*
  - `"a 9-tag joined string over 64 bytes now passes"` — the exact corpus string from entry 44 (89 bytes, 9 tags, longest tag 16 bytes) returns `nil`. **Fails today**, and fails again if any joined-string cap is reintroduced. *(LD3 — the reported regression)*
  - `"description, project and type caps are unchanged"` — table asserting 100000 / 64 / 64 boundaries still hold. Fails if the uniform-multiplier temptation is acted on. *(LD3's "not touched" clause)*
- **`internal/capture/validate_changed_test.go`** *(new)*
  - `"an unchanged over-cap field passes"` — `old.Impact` = 1290×"i", `new.Impact` identical, `new.Title` different-but-legal → `nil`. Fails if `ValidateChanged` validates untouched fields. *(LD4)*
  - `"a changed over-cap field is rejected"` — same `old`, `new.Impact` a *different* 1290-byte string → error. Fails if `ValidateChanged` compares nothing and passes everything. *(LD4)*
  - `"a changed field must meet the new cap, not merely improve"` — `old.Title` = 300×"t", `new.Title` = 257×"t" → error. Fails if the rule is "must not get worse" rather than "must meet the cap". *(LD4)*
- **`internal/cli/edit_test.go`**
  - `"editing the title of an entry with a grandfathered over-cap impact succeeds and preserves impact byte-for-byte"` — seed via `Store.Add` with a 1290-byte impact (bypassing ingress validation), run `edit` with a buffer changing only `title`, assert exit 0 and `Get(id).Impact` byte-equal to the seed. Fails if edit validates unchanged fields — i.e. this is the test that fails if LD7 lands without LD4. *(LD4 + LD7)*
  - `"editing impact to 1025 bytes is a user error and writes nothing"` — assert `errors.Is(err, ErrUser)`, the error mentions `impact`, stdout empty, and `Get(id)` is unchanged. **Does not and cannot assert stderr**: `root.go` sets `SilenceErrors: true`, so `root.Execute()` never writes the error to `cmd.ErrOrStderr()` in-process — only `cmd/brag/main.go` prints the returned error to `os.Stderr`, outside what this test exercises. The stderr-carries-the-message constraint holds in production; it is a property of `main.go`, not of this test. Fails if edit does not validate at all. *(LD7)*
- **`internal/cli/project_test.go`**
  - `"ensure accepts a name of exactly capture.MaxProject bytes and rejects one byte more"` — the test **references `capture.MaxProject`** rather than a literal, so it cannot drift from the constant. Fails if `project.go` keeps its own literal *and* `MaxProject` ever moves; the reference is what makes LD6 enforced rather than tidied. *(LD6)*
- **`internal/memory/`** (package test)
  - `"the worst-case rendered line is 1450 bytes / 363 estimated tokens"` — build an entry at every cap (`id` = math.MaxInt64, project 64, type 64, title `MaxTitle`, impact `MaxImpact`), render it through the same function the slice renders with, assert `len(line) == 1450` and `EstimateTokens(line) == 363`. Computed from the constants, so it fails if any cap moves without DEC-046 and this number moving together. *(LD5)*
  - `"DefaultBudget is 2000"` — a one-line pin. Fails if the budget is changed as a side effect of the cap change. *(LD5's "the number stays" clause)*

## Implementation Context

### Decisions that apply

- `DEC-046` (emitted here) — the new caps, the tags re-shape, grandfathering, the edit path.
- `DEC-044` — the budget model this spec re-derives; do not edit its numbers without LD5's table.
- `DEC-015` — tags are a normalised table with no length constraint; this is why a joined-string cap has no model behind it.
- `DEC-004` — the comma-joined era `MaxTags` belongs to. The *wire* format stays comma-joined; only the cap's shape changes.
- `DEC-012` — introduced the caps with no rationale. DEC-046 supersedes the values; DEC-012 is cross-referenced, not edited.
- `DEC-036` / `DEC-017` — the soft-match invariant `cli/project.go:162` reasons about; LD6 must preserve byte-counting semantics.
- `DEC-027` — the reserved `cost:`/`tokens:` tag checks inside `Validate` are untouched; the per-tag cap must not break `validateReservedTags`.

### Constraints that apply

- `test-before-implementation` — the tests above are written first and must fail for the *expected* reason before implementation (§12).
- `no-sql-in-cli-layer` — LD7 adds validation, not SQL, to `cli/edit.go`.
- `stdout-is-for-data-stderr-is-for-humans` — edit's rejection goes to stderr; assert `outBuf` stays empty.

### Prior related work

- `SPEC-064` (shipped) — unified per-path validation; created the situation this spec resolves.
- `SPEC-073` / `SPEC-074` (shipped) — the memory slice and the MCP push surface that inherit the line shape.

### Out of scope (for this spec specifically)

- The remaining ~9 v0.5.0 audit nits. They stay in STAGE-018's backlog and trail into v0.6.1.
- Widening `project` / `type` / `description`.
- Changing `memory.DefaultBudget` (LD5 decides it holds).
- The v0.6.0 cut itself — a separate spec after this ships.
- Any schema change or backfill of the grandfathered rows.

## Notes for the Implementer

- **Order the work LD1–LD3 → LD4 → LD7.** Wiring edit before the caps and the
  changed-field rule exist makes ~30% of the corpus uneditable. This is the one
  sequencing rule in the spec that is not a preference.
- `Validate` and `ValidateChanged` must share the field rules — extract the
  per-field check once and have both call it, rather than writing the caps
  twice. This spec exists partly because a cap was written twice.
- The tags error for a too-long tag must name **the tag**, not the joined
  string; that distinction is the whole point of LD3 and it has a test.
- `validateReservedTags` already splits on `,`. The per-tag length and count
  checks want the same split — do it once and pass the tokens to both.
- **Mutation-check the coverage claims.** Before build reports a test as pinning
  a decision, flip the decision and confirm *that* test goes red. A prose claim
  about what a test pins is aspirational until each clause is checked
  individually — SPEC-073's coverage sentence was wrong four times, every time
  caught by a fresh reviewer and never by the writer.
- **For the doc sweep, grep the feature name, not the number's phrasing.**
  SPEC-074's sweep leaked six times from one root cause. The Outputs table above
  was built from identifier greps; re-run them at build and treat any delta as a
  question for the spec author, not a unilateral scope change (§9, both sides).
- The `docs/brag-entry.schema.json` `tags` field loses `maxLength` entirely. A
  JSON Schema `maxLength` on a comma-joined string cannot express "64 per tag,
  32 tags"; the binary stays the authoritative validator (the schema says so
  already) and the description carries the rule.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-075-capture-caps`
- **PR (if applicable):** https://github.com/jysf/bragfile000/pull/144
- **All acceptance criteria met?** yes
- **New decisions emitted:**
  - `DEC-046` — capture field caps, derived; the tags re-shape; the edit path (emitted at design; this cycle implemented it and corrected DEC-044 per LD5)
- **Deviations from spec:**
  - **A third premise-audit miss, found only at fail-first, not by either audit-grep.** `internal/cli/add_hardening_test.go` (SPEC-064-era) hardcodes the pre-DEC-046 cap boundaries as bare numeric input lengths (`strings.Repeat("x", 201)`, `strings.Repeat("i", 257)`, a `200`-byte "at cap" boundary test) with no cap-identifier and no error-message-substring assertion — so neither of the spec's two audit greps (`MaxTitle\|...` identifier grep; `"exceeds 200\|..."` message-substring grep) could find it. It surfaced as 4 unexpectedly-failing tests immediately after implementing LD1/LD2. Fixed in the same shape as the two files the spec did enumerate: `title over 200`→`title over 256` (201→257 bytes), `impact over 256`→`impact over 1024` (257→1025 bytes), the at-cap boundary moved 200→256, and the case names updated to match. Recorded here per §9's audit-grep cross-check rule rather than treated as silent scope creep.
  - **Live-corpus re-measurement drift, reconciled per the build rules.** The corpus is still 359 entries, but `MIN(id)=1, MAX(id)=369` proves genuine churn since design's snapshot (10 rows deleted, 10 added, net count unchanged) — the "359" match is coincidental, not a frozen fixture. Re-running the spec's `## Measurement` queries: **all over-cap counts reproduced exactly** (impact 74/285, title 33/359 — all from `zany-animal-slots` ids 157–228 — tags 7/260), **all max values reproduced exactly** (impact 1290, title 1444, tags 89), the tags mean reproduced exactly (27.2 ≈ 27), and the worst-case *rendered memory line* reproduced exactly (1483 B / 371 tok, entry 172 — DEC-044/LD5 finding (i)). Percentiles (p50/p90/p95/p99) drifted by single-digit-byte amounts on `title`/`impact`, attributable to the churn, not a methodology difference — none of it is load-bearing for any locked decision (each was chosen on distribution *shape*, per the spec's own framing). One figure moved beyond rounding: the design's "longest single non-reserved tag is 21 bytes" now measures as **20 bytes** (`spec-driven-template`) — noted, not acted on; `MaxTagLen=64` keeps 3×+ headroom either way. Re-ran `brag memory` against the live corpus (built from source — the installed Homebrew binary is v0.5.1, which predates the `memory` command) and confirmed DEC-044/LD5's re-derived yield still holds today: `Included: 25, Skipped: 175, Estimated: 1991`.
- **Follow-up work identified:**
  - None beyond STAGE-018's existing backlog (the remaining ~9 v0.5.0 audit nits, explicitly out of scope here).

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing in the spec itself was unclear — LD1–LD7 were unambiguous and the failing tests were largely transcribable. The one real friction was outside the spec's control: its own audit-grep methodology (identifier grep + error-message-substring grep) has a blind spot for tests that assert *only* "an error occurred" against a bare numeric literal, with no cap name and no message substring to grep for (`add_hardening_test.go`). That gap was only visible once tests actually ran red after implementation.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — No — DEC-046 and the spec's own Outputs/audit-grep table covered every load-bearing decision. The `add_hardening_test.go` miss is a gap in *grep coverage*, not a missing constraint or decision.

3. **If you did this task again, what would you do differently?**
   — Add a third audit-grep angle before implementing, not after: grep test files for the *raw numeric cap values themselves* (`200`, `201`, `256`, `257`, `64`, `65`) in addition to the spec's identifier grep and message-substring grep. That third angle would have caught `add_hardening_test.go` before the cap change landed rather than via a fail-first surprise immediately after.

## Punch-List Iteration

*Verify returned five items (two substantive: 1, 3; three prose: 2, 4, 5). Fixed
in place, same PR (#144). No cap value changed; LD1–LD7 stand.*

- **Item 1 — the ≈110-entries figure had propagated into two shipped DECs.**
  DEC-044's own correction blockquote swept the byte/token pair
  (`626\|157 tok`) but never the yield number, so it never caught where the
  number had already been copied onward. Found by re-running the grep against
  the yield figure itself (`≈110`, `110 entries`) across `decisions/`:
  - `DEC-043:122` — the rank-fusion pool-cap paragraph cited "at the default
    budget (≈110 entries) the tail of a typical slice is inside that
    [perturbation] zone." **This was not a number swap — the correction
    inverts the conclusion.** The true default-budget yield is 25, which sits
    *inside* DEC-043's own top-26 head guarantee two sentences earlier, so a
    default-budget slice never reaches the perturbation zone at all. Fixed
    with a correction blockquote (matching DEC-044's own style) that restates
    the consequence rather than only the number, and notes the tail-imprecision
    tradeoff still applies to a wider explicit `--budget`/`--limit`.
  - `DEC-045:270` — sub-decision 7's resource-budget derivation cited "2000 is
    ~10% of it, ≈110 entries" as part of the displacement argument. Fixed with
    a correction blockquote: the displacement defense (2000 is ~10% of the
    ~20k-token ritual) was never about yield and is unaffected; only the
    entry-count aside was false.
  - DEC-044's own correction blockquote gained a second paragraph ("How far
    the ≈110 figure travelled") naming both downstream citations and the
    grep-blindness reason they survived the original sweep, so the record
    shows the figure's full travel rather than just its origin and its two
    silent copies.
  - Both DEC-043 and DEC-045 gained a `DEC-046` cross-reference in their own
    References sections so a future reader following either chain lands on
    the re-derivation.

- **Item 3 — the tag caps DID bound stamped provenance on the edit path.
  Closed by option (a): strip reserved-namespace tags before counting.**
  Two failing tests were written first in
  `internal/capture/validate_changed_test.go`, run red, confirmed failing for
  the exact reason verify described, then made to pass:
  - `TestValidateChanged_StampedTagsDoNotCountAgainstMaxTagCount` — 30 user
    tags + 5 stamped reserved tags (35 stored), editing one user tag.
    **Red:** `tags: more than 32 tags allowed (got 35)`. **Green** after the
    fix.
  - `TestValidateChanged_StampedTagsDoNotCountAgainstMaxTagLen` — a stamped
    `model:` tag 10 bytes over `MaxTagLen`, editing an unrelated user tag.
    **Red:** `tag "model:mmm…" exceeds 64-character limit`. **Green** after
    the fix.
  - Fix: `capture.validateTagsChanged` (new, called from `ValidateChanged`
    only) strips tokens matching the five reserved prefixes
    (`agent:`/`model:`/`session:`/`cost:`/`tokens:`) before applying
    `MaxTagCount`/`MaxTagLen`, mirroring the pre-stamp population `Validate`
    already sees on add. `cost:`/`tokens:` shape validation
    (`validateReservedTags`) still runs against the full token list — that
    check is about value correctness, not the caps, and stripping doesn't
    touch it. Chose (a) over (b) per the spec's own lean: it makes DEC-046's
    "neither cap bounds stamped provenance" sentence true as written on both
    ingress paths, rather than narrowing the sentence to admit an asymmetry
    that was really just an accident of validation order. DEC-046's
    scope-limit paragraph and Consequences section are rewritten to explain
    *why* stripping is correct, not just that it happens, and to name the one
    accepted residual asymmetry (a hand-typed `model:x` in freeform `tags` is
    capped on add but exempt on edit — the opaque comma-joined string cannot
    tell the two apart in either direction).
  - All 1036 repo tests pass after the fix; `gofmt -l .` and `go vet ./...`
    stay clean.

- **Item 2 — `questions.yaml:621` still stated 626 B / 157 tok as current fact.
  Grep ran, hit, mis-reconciled — a different failure than a grep that
  missed.** SPEC-075's own audit-grep table lists `questions.yaml:602` as a
  hit for the `626\|157 tok` grep and reconciles it "all enumerated in
  Outputs" — but the Outputs entry it points at
  (`guidance/questions.yaml` → `status: resolved`) only covers the
  status/resolution block, never the prose notes above it, where the same
  stale figure appears inside a quoted DEC-044 sentence. The grep surfaced the
  line; the reconciliation step didn't read past the block it was checking
  for. Fixed by annotating the quote in place (bracketed correction pointing
  at the 1450/363 re-derivation) rather than rewriting the historical notes
  — the notes describe what was true when the question was raised, and the
  annotation is what stops a reader treating a quoted historical claim as
  current.

- **Item 4 — the Failing Tests entry overclaimed what the edit-rejection test
  asserts.** The spec said "stderr carries the message"; the shipped test
  (`TestEditCmd_ChangedImpactOverCapIsUserErrorNoWrite`) asserts
  `errors.Is(err, ErrUser)`, the error mentions `impact`, `outBuf` is empty,
  and the DB is unchanged — it makes no assertion on `errBuf` at all, and
  could not: `root.go` sets `SilenceErrors: true`, so `root.Execute()` never
  writes to `cmd.ErrOrStderr()` in-process; only `cmd/brag/main.go` prints the
  returned error to `os.Stderr`, outside what the test exercises. The
  production constraint holds; the test just doesn't pin it. Corrected the
  Failing Tests entry to describe what the test actually asserts.

- **Item 5 — SPEC-075 was cited as "shipped" while still in verify.**
  `DEC-044:454` literally said "SPEC-075 (shipped — …)"; corrected to
  "in verify, PR #144 open". `DEC-046:361`'s References line didn't use the
  word "shipped" for SPEC-075, but carried no status at all while every
  sibling entry in the same list did — made explicit ("in verify, PR #144
  open") for the same reason: consistency with how every other related spec
  in that list states its status.

- **Audit-grep re-run, this time including the yield number and other bare
  figures a claim rests on (the sixth grep-blindness mode, per the spec's own
  Note-for-the-Implementer prediction).** Re-ran all five original greps
  (identifiers, message substrings, docs, `626\|157 tok`, ingress paths) —
  no new misses; all hits reconcile to already-known/already-fixed sites.
  Added `grep -rn "≈110\|110 entries" decisions/ guidance/ projects/ docs/
  internal/`, which is what surfaced the DEC-043/DEC-045 propagation above.
  One further hit, correctly left untouched:
  `projects/…/specs/done/SPEC-073-…md:1328` (an archived, shipped spec's Ship
  Reflection, "≈110 entries, about 10% of the ~20k-token ritual"). Same
  category as `docs/reports/security/2026-04-26-…`, which the spec's own
  audit-grep table already carved out as a dated historical record: a done
  spec's reflection is a record of what was believed true at ship time, not
  living documentation, and SPEC-075's Outputs never enumerated it as a file
  to update. Left as-is; recorded here rather than silently skipped.

- **Gates:** `go build`, `go test ./...` (1036 tests), `gofmt -l .` (empty),
  `go vet ./...`, `just test-docs`, `just test-hook` — all six green.

### Second Punch-List Round (2026-08-10)

*Verify returned three more items against the round-one fixes above (one
substantive: 2; two prose: 1, 3). Fixed in place, same PR (#144). No cap
value changed, no LD changed; LD1–LD7 still stand.*

- **Item 2 (code) — `isReservedTag` was a second, unpinned enumeration of
  the five reserved prefixes; stamping built its own independently, with no
  shared definition and no agreement test.** Failing test written first, in
  `internal/mcpserver/provenance_test.go`:
  `TestStampProvenance_UnhandledReservedPrefixIsCaught` mutates
  `capture.ReservedTagPrefixes` (the newly-exported single source, see
  below) by appending a fake sixth prefix `"signature:"`, then expects
  `stampProvenance` to panic. **Red** (confirmed before touching
  `provenance.go`): `stampProvenance` didn't consult
  `capture.ReservedTagPrefixes` at all yet, so the mutation had zero effect
  on it — `t.Fatal("stampProvenance did not notice the new reserved prefix
  \"signature:\" ...")`. **Green** after the fix below.
  - Fix, single-sourced rather than an agreement test (per the punch-list's
    own lean — a shared constant makes the drift impossible, not just
    caught after the fact): `internal/capture/validate.go` now exports
    `ReservedTagPrefixes = []string{"agent:", "model:", "session:", "cost:",
    "tokens:"}`; `isReservedTag` ranges over it instead of its own literal
    array. This half is behavior-preserving by construction — the capture
    suite stayed green before any `mcpserver` change. `internal/mcpserver/
    provenance.go`'s `stampProvenance` now builds its tags by ranging over
    `capture.ReservedTagPrefixes`, with a `switch` case per known prefix and
    a `default: panic(...)` for any prefix with no case. A sixth reserved
    prefix can no longer be added to the shared list without stamping
    either handling it (a case is added) or panicking immediately on the
    very next `brag_add`/`brag edit` call (a case isn't) — closing the exact
    scenario verify described: a future signed-provenance prefix stamped
    while unknown to `validateTagsChanged`'s stripping.
  - A second test closes the loop from the other direction:
    `TestValidateChanged_NewlyRegisteredReservedPrefixIsStrippedAutomatically`
    (`internal/capture/validate_changed_test.go`) registers a prefix in
    `capture.ReservedTagPrefixes` with **no** `validate.go` edit at all, and
    confirms `validateTagsChanged` strips it immediately — proving the fix
    is a true single source in both directions, not a one-way patch.
  - Mutation-check, as instructed: the fake `"signature:"` prefix lives only
    inside the test body (`t.Cleanup` restores the original slice
    afterward) — the red→green pair above *is* the mutation-check, not a
    separate manual step.
  - All 1038 repo tests pass after the fix (1036 baseline + the two new
    tests above); `gofmt -l .` and `go vet ./...` stay clean.

- **Item 1 — DEC-043's round-one correction was itself a logic error: a
  COUNT of included entries compared against a RANK threshold, not a
  score.** Verified the counterexample verify supplied: an entry at recency
  rank 40 *and* match rank 40 scores `1/(60+40) + 1/(60+40) = 0.02000`,
  beating a single-list rank-1 entry (`1/(60+1) ≈ 0.01639`) despite sitting
  outside every list's top 26 — so "25 entries included" does not, by
  itself, imply "everything included is within the top-26 head guarantee,"
  and the round-one conclusion ("a default-budget slice never reaches the
  perturbation zone at all") did not follow from its own premise.
  - Rewrote `decisions/DEC-043-…` (the pool-cap paragraph's correction
    blockquote, the item verify's ID 1 targeted) to make the SCORE argument
    instead: the default-budget slice's included set measures as exactly
    recency ranks 1–25 today, so its weakest member scores exactly
    `1/(60+25) = 1/85 ≈ 0.011765`, above the worst-case pool-cap-excluded
    score of `3/261 ≈ 0.011494` already derived earlier in the same
    paragraph — a real, verified bound, but scoped to the corpus and budget
    as measured, not a structural "never." Named why it isn't structural:
    DEC-044 sub-decision 3's skip-and-continue means the included set is a
    rank prefix only when nothing large enough to skip appears early in the
    order, and this spec just widened the worst-case rendered line from 157
    to 363 tokens — making a long-head/short-tail fill (a departure from the
    clean rank-prefix case) *more* reachable than before, not less.
  - `decisions/DEC-044-…`'s "How far the ≈110 figure travelled" paragraph
    repeated the same "sits inside the top-26 head guarantee" framing when
    summarizing DEC-043's round-one correction. Reworded to describe the
    inversion without restating the flawed arithmetic, pointing at DEC-043
    as the one place the score-based version lives — a third copy of the
    same number is a third place it can drift on its own.
  - Checked `decisions/DEC-045-…` for the same pattern: it does not have
    it — its correction was always a pure number swap (≈110 → 25) with no
    rank/count claim attached, so it needed no change.
  - This item is prose, but a logical claim rather than a wording
    preference, per the spec's own instruction: checked the arithmetic
    (`1/85 > 3/261`, both recomputed above) rather than rewording around it.

- **Item 3 — the edit-path exemption from the caps also exempts the
  control-character check, and neither `validate.go`'s comment nor
  DEC-046's residual-asymmetry paragraph said so.** `validateTagsChanged`
  moves `hasC0Control` inside the same user-only loop as `MaxTagLen`/
  `MaxTagCount` (`validate.go:199–206`), so a reserved-prefixed token
  escapes the control-character check on edit exactly as it escapes the
  caps — a third exempted check, previously unstated in both places.
  - `internal/capture/validate.go`'s `validateTagsChanged` doc comment
    (`:176–187`) now names all three (`MaxTagLen`, `MaxTagCount`,
    `hasC0Control`) instead of just the two caps, and states this is not a
    new gap: a hand-typed C0 byte in a reserved-prefixed `agent`/`model`/
    `session` param is already reachable today through those dedicated
    params, which `Validate` never inspects.
  - `decisions/DEC-046-…`'s residual-asymmetry paragraph is reworded to say
    a hand-typed `model:x` in freeform `tags` is capped *and*
    control-character-checked on add but exempt from **both** on edit, and
    states explicitly that the strip opens no new gap — it extends an
    exemption that already existed at the dedicated-param door to the
    freeform-`tags` door, rather than introducing one.
  - Behavior is unchanged, per the spec's own instruction that this item is
    prose-only: checking a stamped tag's control characters on edit would
    re-create the uneditable-row failure LD4 exists to prevent, the same
    reason the caps are already exempt. If that judgment turns out wrong,
    it is a DEC-046 amendment, not a silent fix — and it is not one here.

- **Audit-grep re-run, second round.** Re-ran all six original greps
  (identifiers, message substrings, docs, `626\|157 tok`, ingress paths,
  and the round-one `≈110\|110 entries` addition) — no new misses; every
  hit reconciles to an already-known or already-fixed site (including
  `projects/…/STAGE-018-…md:179`, a design-notes quote of the *old*,
  already-wrong "≈110" figure being reported as wrong — historical, left
  as-is, same category as the round-one carve-outs). Added a seventh grep
  specific to item 2 — `grep -rn '"agent:", "model:", "session:", "cost:",
  "tokens:"' --include="*.go" internal/` — which now returns exactly one
  hit, `capture/validate.go`'s `ReservedTagPrefixes`, confirming the literal
  no longer exists independently in `mcpserver/provenance.go`.

- **Gates:** `go build`, `go test ./...` (1038 tests), `gofmt -l .` (empty),
  `go vet ./...`, `just test-docs`, `just test-hook` — all six green.

### Re-Verify of the Second Round (2026-08-10)

Independent re-verify of the punch-list #2 delta (`c5bf954`, `3a54a6e`,
`f57f386`). Both round-two fixes are structurally correct — but two of the
claims *made about* them were not, and both are the stage's own process rules
firing on the round that wrote them.

- **Item 2's stripping-half test pinned nothing** (fixed here).
  `TestValidateChanged_NewlyRegisteredReservedPrefixIsStrippedAutomatically`
  was sized on the count axis with `MaxTagCount-1` user tags plus two stamped
  tags, so an unstripped `signature:` landed on **exactly** `MaxTagCount` (32)
  and passed the cap. Mutation-checked by severing `isReservedTag`'s coupling
  to `ReservedTagPrefixes` and re-running: **the test stayed green**, so its
  comment's claim — "registering a sixth reserved prefix there is enough on
  its own" — was aspirational. Re-sized onto the length axis (an over-cap
  `signature:` value), matching the `MaxTagLen` sibling's proven idiom; the
  same mutation now fails with `tag "signature:ddd…" exceeds 64-character
  limit`. The *stamping*-half mutation-check
  (`TestStampProvenance_UnhandledReservedPrefixIsCaught`) was re-run the same
  way and **does** have teeth — neutering the `default:` panic fails it with
  the intended diagnostic. So item 2's closed loop was only half-closed, and
  the half that was pinned is the half the handoff had already verified.

- **DEC-043's headroom figure was wrong** (fixed here). The corrected
  score-based paragraph justified "nothing else displaces in" with *"only 9
  tokens remain after rank 25"*. Re-measured against the live corpus: the
  headroom is **4** tokens (`Estimated: 1996` of 2000; the 25 per-line
  estimates sum to 1996 exactly, reproduced independently). Not staleness — a
  new entry did land ~2 minutes after `3a54a6e`, but it displaced one costing
  an identical 94 tokens, so the figure was 4 at commit time too. The
  conclusion is unaffected and strengthened, and the bound is now stated
  against the **21-token smallest line in the whole 200-entry pool** rather
  than a bare subtraction, so it survives corpus growth — the property both
  earlier versions of the paragraph lacked.

- **What did hold.** The rank-vs-score category error is correctly diagnosed
  and its replacement argument is sound: the arithmetic checks
  (`1/100 + 1/100 = 0.02 > 1/61 ≈ 0.01639`; `1/85 ≈ 0.011765 > 3/261 ≈
  0.011494`), `3/261` traces to its derivation at DEC-043:116, and `157 → 363`
  and `Included: 25` are consistent across DEC-044/046 and this spec. The
  load-bearing structural premise — that the included set is a clean rank
  prefix — was verified *directly* (slice ids compared against the 25 most
  recent ids), not inferred from the count, which is the error the round-two
  fix was itself correcting. DEC-046's "not a new gap" claim holds:
  `Validate(f Fields)` takes no agent/model/session params, so a C0 byte is
  already reachable there today. `aggregate.IsAgentAuthored`'s separate
  `agent:`/`model:` literals are **not** a third drift site — it is an
  authorship subset pinned to a SQL clause by its own cross-check test, and
  must not follow `ReservedTagPrefixes`. No `t.Parallel()` in either package,
  so the mutable package var carries no race hazard.

- **Process rules this round earned.** (1) Rule 1 now applies to *mutation
  checks themselves*: "mutation-verified" is only as good as the axis the
  probe is sized on — a guard that trips on count can be absorbed by an
  off-by-one, so size the probe where the failure is unambiguous. (2) A
  corpus statistic embedded in a decision doc is a measurement of a **live,
  growing** corpus that the tool itself writes to; state such bounds against
  a distribution property (pool minimum) rather than a subtraction, or they
  rot silently. This is process rule 3's fixture-vs-corpus error with the
  time axis added.

- **Gates, re-run after the fixes:** `go build`, `go test ./...`, `gofmt -l .`
  (empty), `go vet ./...`, `just test-docs`, `just test-hook` — all six green.

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — Size a mutation probe on the axis where failure is unambiguous, and make
   the probe fail *before* trusting the claim it supports. This spec produced
   two mutation checks for the same coupling; one had teeth and one did not,
   and the difference was a single off-by-one — `MaxTagCount-1` user tags plus
   two stamped tags put an unstripped prefix on *exactly* the cap, so the test
   passed with the coupling severed. Both were written in the same sitting,
   both were described in prose as pinning the behaviour, and the one that was
   checked was the one that was already fine. The lesson is not "mutation-check
   more" — this stage already knew that — it is that **"mutation-verified" is a
   claim about a probe, and inherits the probe's blind spots.** Run the mutation
   first and watch it fail; a check that has never been seen red is a hypothesis.

2. **Does any template, constraint, or decision need updating?**
   — Yes, and it was fixed here: `projects/_templates/spec.md` and
   `spec-release-cut.md` both ended question 4 with instruction prose and **no
   `— <answer>` placeholder**. `scripts/archive-spec.sh`'s guard greps for
   `^   — <answer>`, so it was structurally incapable of catching an unanswered
   question 4 — the one question whose answer is transcribed into a brag's
   `impact`. Every shipped spec happened to answer it anyway, which is exactly
   how a blind guard stays invisible. Both templates now carry the placeholder.
   Separately, this spec's own file had leaked tool-call markup (`</content>`,
   `</invoke>`) committed at its end via #144; removed here, and a repo-wide
   grep confirms it is not present anywhere else.

3. **Is there a follow-up spec I should write now before I forget?**
   — No new spec, but two open items are now sharper and belong to the
   already-planned pillars. (a) DEC-046's residual asymmetry — a C0 control byte
   is reachable through the `agent`/`model`/`session` params on the add path
   because `Validate(f Fields)` never sees them — is a **signed-provenance**
   concern, not a caps one, and should be resolved when that pillar is framed
   rather than patched piecemeal. (b) DEC-043's `k=60` remains unexamined and is
   already on the open-questions list. The one thing worth writing down now is
   narrower than a spec: any corpus statistic quoted in a decision doc should be
   stated against a distribution property (a pool minimum, a percentile) rather
   than a subtraction against a live total, because the corpus grows underneath
   the claim — three of this stage's documented errors are that shape.

4. **What can a user do now that they couldn't before?** — one sentence,
   before → after; quote the confirming number if one exists, name the outcome
   if not. Write `none` if this spec has no user-visible outcome — that is a
   real, greppable result, not a blank. This is the line a brag's `impact` field
   is transcribed from, and both halves are already written above (## Context is
   the before, ## Goal is the after): confirm the prediction, don't reconstruct
   it from memory.
   — Before, **74 of 285** entries with an `impact` (26%) were over a 256-byte
   cap nobody had derived, so the corpus contained a large slice of its own
   history that capture would refuse to write — and `brag edit` silently
   enforced nothing at all. Now `impact` is capped at 1024 and `title` at 256,
   both derived from measured distributions rather than inherited from DEC-012;
   the joined-string tags cap is re-shaped into `MaxTagLen` 64 + `MaxTagCount`
   32, which catches the actual abuse (one absurd tag) instead of penalising the
   legitimate one (many short tags — all 7 over-cap strings held 7–9 tags whose
   longest was 17 bytes); and `brag edit` validates on write-of-changed-field,
   so the caps reach the edit path **without** making the already-over-cap
   corpus uneditable.

---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-075
  type: story                      # epic | story | task | bug | chore
  cycle: build
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

- [ ] `MaxTitle == 256`, `MaxImpact == 1024`; `MaxProject`/`MaxType`/`MaxDescription` unchanged.
- [ ] `MaxTags` no longer exists; `MaxTagLen == 64` and `MaxTagCount == 32` do.
- [ ] A 9-tag comma-joined string longer than 64 bytes **validates successfully** (the reported regression).
- [ ] A single tag longer than 64 bytes is rejected, and the error names the offending tag.
- [ ] More than 32 tags is rejected with a count-shaped error.
- [ ] `capture.ValidateChanged` passes an unchanged over-cap field and rejects a changed one.
- [ ] `brag edit` on an entry with a grandfathered over-cap `impact` succeeds when only `title` changes, and the stored `impact` is byte-identical afterwards.
- [ ] `brag edit` rejects a changed field that violates a cap, with no database write.
- [ ] `cli/project.go` contains no numeric literal for the project cap; a `MaxProject`-length name is accepted and `+1` rejected.
- [ ] The worst-case memory line is exactly 1450 bytes / 363 estimated tokens.
- [ ] `memory.DefaultBudget` is unchanged at 2000.
- [ ] DEC-044's two 626/157 citations and its ~18-tokens/≈110-entries claim are corrected.
- [ ] Every doc site enumerated in Outputs is updated; the dated security report is not.
- [ ] Full gate set green: `go test ./...`, `gofmt -l .` empty, `go vet ./...`, `scripts/test-docs.sh`.

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
  - `"editing impact to 1025 bytes is a user error and writes nothing"` — assert `UserError`, stderr carries the message, stdout empty, and `Get(id)` is unchanged. Fails if edit does not validate at all. *(LD7)*
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
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - `DEC-046` — capture field caps, derived; the tags re-shape; the edit path
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - [any new specs for the stage's backlog]

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — <answer>

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — <answer>

3. **If you did this task again, what would you do differently?**
   — <answer>

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>

4. **What can a user do now that they couldn't before?** — one sentence,
   before → after; quote the confirming number if one exists, name the outcome
   if not. Write `none` if this spec has no user-visible outcome — that is a
   real, greppable result, not a blank. This is the line a brag's `impact` field
   is transcribed from, and both halves are already written above (## Context is
   the before, ## Goal is the after): confirm the prediction, don't reconstruct
   it from memory.
</content>
</invoke>

---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-046                        # stable, never reused
  type: decision                     # decision | analysis | recommendation | observation
  confidence: 0.85                   # 0.0 - 1.0, honest assessment
  audience:                          # who needs to know?
    - developer
    - agent

agent:
  id: claude-opus-5
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-006                       # the project during which this was decided
repo:
  id: bragfile

created_at: 2026-08-09
supersedes: null                     # DEC-012's cap VALUES are superseded; its schema shape is not
superseded_by: null

tags:
  - validation
  - capture
  - caps
  - tags
  - grandfathering
  - token-budget
---

# DEC-046: The capture field caps, derived — a per-tag tags model, grandfathering by changed-field, and DEC-044 re-derived

## Decision

Each `internal/capture` field cap is set from what its field is **for**, measured
against the live corpus, rather than inherited from DEC-012's `add --json`
schema: **`MaxTitle` 200 → 256**, **`MaxImpact` 256 → 1024**, **`MaxTags` (a cap
on the comma-joined string) deleted and replaced by `MaxTagLen = 64` per tag plus
`MaxTagCount = 32` per entry**, with `MaxProject` / `MaxType` / `MaxDescription`
**deliberately unchanged**. Existing over-cap rows are grandfathered by
**validating only fields whose value changed**, which is what makes it safe to
wire `capture.Validate` onto the edit path — the fifth and last unvalidated
ingress. DEC-044's budget model is **re-derived here, not left stale**: the
worst-case memory line moves from 626 bytes / 157 tokens to **1450 bytes / 363
tokens**, and `memory.DefaultBudget` stays at 2000.

## Context

`capture.Validate` enforces six byte caps across every capture ingress path.
None of them was derived. DEC-012 introduced them for the `brag add --json`
schema and records no rationale for any value. SPEC-064 then correctly unified
per-path validation — flag and editor mode had been skipping the caps that
`--json` and MCP enforced — and in doing so generalised those numbers to every
path. The side effect nobody measured: a large slice of the **existing** corpus
became unwritable.

The trigger was a user report ("Tags have a 64-char limit"). Investigation found
the pattern was general, and `guidance/questions.yaml` →
`capture-caps-were-inherited-not-derived` recorded the measurement and an
explicit ordering risk. This DEC resolves that question.

**The measurement is the argument.** Against a 359-entry snapshot with
provenance tags stripped (byte counts, matching `len`):

| field | cap | p50 | p90 | p95 | p99 | max | over cap |
|---|---|---|---|---|---|---|---|
| `impact` | 256 | 228 | 413 | 514 | 932 | 1290 | 74 of 285 (26%) |
| `title` | 200 | 73 | 134 | 541 | 1067 | 1444 | 33 of 359 |
| `tags` (joined) | 64 | 27 mean | — | — | — | 89 | 7 of 260 |
| `project` | 64 | 9 | — | — | — | 22 | 0 |
| `type` | 64 | 7 | — | — | — | 10 | 0 |
| `description` | 100000 | 570 | — | — | — | 2071 | 0 |

The contrast is the finding, and it is why this is not a uniform multiplier:
`description` is over-provisioned **48×** while `impact` — the field
`brag impact`, `wrapped`, `story`, `coverage` and `memory` all actually read — is
exceeded by a quarter of the corpus and has a median at **89% of its ceiling**.
Each cap was chosen independently with no model of what the field is for.

Two findings separated at design that the original measurement did not:

1. **`title`'s over-cap tail is not ordinary usage.** All 33 over-cap titles come
   from a *single* project (`zany-animal-slots`, ids 157–228) — one
   malformed-capture episode where whole brag bodies were written into `title`.
   Excluding it, `title` p90 = 134 against a 200 cap. `impact`'s over-cap rows,
   by contrast, span **seven** projects with 109 entries piled into `[200, 256]`.
   One field is genuinely under-provisioned; the other has a data-entry bug in
   its tail. They therefore get different treatment.
2. **The tags cap fires on count, never on length.** All 7 over-cap tag strings
   carry 7–9 *individually short* tags (longest among them: 17 bytes). The
   longest single tag in the entire corpus is 21 bytes; p99 is 18. A joined-string
   cap penalises using many tags (legitimate) instead of one absurd tag (the
   actual abuse) — exactly inverted.

DEC-015 normalised tags into a `tags` table (`name TEXT NOT NULL UNIQUE`, no
length constraint), so the joined-string cap is a leftover from DEC-004's
comma-joined era with no model behind it. The wire format stays comma-joined
(DEC-004 is untouched); only the cap's shape changes.

**The coupling that made this non-trivial.** DEC-044 cites these caps twice for
the memory slice's budget model — *"the `capture.Validate` field caps bound a
single line at 626 bytes / 157 tokens worst case"* — which is the argument that
one long entry cannot starve a slice. A new `MaxImpact` moves that number, so
leaving it would silently invalidate a shipped decision one release old.

## Alternatives Considered

- **Option A: a uniform multiplier on every cap.**
  - What it is: scale all six caps by the same factor (say 4×).
  - Why rejected: the measurement says the problem is not scale but *shape*.
    `description` would go from 48× over-provisioned to 192×, `project`/`type`
    would gain headroom nothing asked for, and `tags` would stay the wrong kind
    of cap at a larger number. A uniform factor is precisely the move that would
    reproduce the original defect — a number nobody can defend per field.

- **Option B: caps wide enough that the whole existing corpus fits (impact ~1536, title ~1536).**
  - What it is: pick each cap above the observed maximum so no row is over cap and
    grandfathering is unnecessary.
  - Why rejected: it sets the contract from the corpus's worst data-entry accident.
    `title`'s 1444-byte maximum is a pasted brag body, not a headline; encoding it
    as the contract makes the malformed shape legal forever and removes the only
    signal that would ever get those 29 rows fixed. It also pushes the worst-case
    memory line to ~2000 bytes / ~500 tokens, a quarter of the default budget for
    one entry.

- **Option C: a row-level grandfather flag or a `validated_at` column.**
  - What it is: mark existing rows exempt; enforce the new caps only on rows
    created after the change.
  - Why rejected: a schema change plus a migration to encode "these rows are
    exempt forever". It never converges — an exempt row stays exempt through
    unlimited edits — and it adds a permanent second class of row that every read
    path and every future contributor has to know about.

- **Option D: keep the edit path unvalidated (status quo).**
  - What it is: change the caps, leave `brag edit` / `Store.Update` as the one
    ingress that validates nothing.
  - Why rejected: it is the open backlog item, and it leaves a path by which any
    value — control characters, a 100KB impact, a malformed reserved `cost:` tag —
    enters the database unchecked. The whole point of SPEC-064 was that per-path
    validation drift is a defect.

- **Option E: validate inside `Store.Update`.**
  - What it is: put the check in storage so every caller is covered by
    construction.
  - Why rejected: SPEC-064 deliberately put ingress validation in `capture`,
    called by the boundary layers, with storage taking already-validated typed
    values. Validating in `Update` would also validate internal callers (restore,
    a future import) that legitimately carry grandfathered text, and would force
    storage to re-read the row to learn which fields changed.

- **Option F (chosen): per-field derived caps + a per-tag/count tags model + changed-field validation.**
  - What it is: `MaxTitle = 256`, `MaxImpact = 1024`, `MaxTagLen = 64`,
    `MaxTagCount = 32`, `project`/`type`/`description` untouched; on edit, validate
    a field only when its value differs from what is stored.
  - Why selected: every number is derived from measured usage plus a statement of
    what the field is for, and the grandfathering falls out of the ingress rule
    rather than the schema. It needs **no migration and no new column**; it
    **converges**, because any field a user touches is brought to contract; and it
    makes the edit-path wiring safe, which is the item that was blocked. Residual
    over-cap rows: 3 impacts, 29 titles, 0 tag violations — each still editable in
    every field except the offending one.

### The derivations, per field

- **`impact` = 1024.** The one-line outcome: before → after with a confirming
  number. 1024 covers measured p99 (932) with headroom and is **4× the field's
  median**, restoring the headroom class `project` (7×) and `type` (9×) already
  have and that `impact` uniquely lacked at 1.12×. Bounded above by the
  memory-line argument below, which is what stops "widen it" becoming "widen it
  a lot".
- **`title` = 256.** A headline. Legitimate-use p90 is 134; 256 gives ~2× headroom
  and keeps a title inside two 120-column terminal lines. Deliberately *not*
  widened to admit the malformed tail (see Option B).
- **`MaxTagLen` = 64.** A tag is an identifier-like label, not prose — the same
  thing `project` and `type` are, derived the same way. Corpus maximum is 21
  bytes, so 64 is 3× headroom, and it is the cap that catches the actual abuse.
- **`MaxTagCount` = 32.** Corpus maximum is 11 including provenance, 9
  user-supplied; 32 is ~3× headroom. Applied **pre-stamp** — `Validate` runs on
  caller-supplied text before the up-to-5 reserved provenance tags are added — so
  provenance can never push a legitimate entry over.
- **`project` / `type` / `description` unchanged.** Measured headroom 3×, 6× and
  48×. Widening them would be the reflexive move this decision exists to reject.

**Scope limit, stated so nobody later assumes otherwise: neither `MaxTagLen` nor
`MaxTagCount` bounds STAMPED provenance.** `Validate` runs on caller-supplied
text, and `agent:` / `model:` / `session:` / `cost:` / `tokens:` are appended
*after* it (`cost`/`tokens` get numeric normalisation; the other three are opaque
passthrough with no length check anywhere in the tree). So the same-looking token
is capped or not depending on which door it came through: `model:x` typed into
the freeform `tags` field is validated as an ordinary tag, while `model:x`
supplied via the dedicated param is not. A 500-byte model id is storable today
and remains so after this decision.

That asymmetry predates this decision and is deliberately left in place — it is
**not** a caps problem. Bounding what a caller asserts about itself is the
signed/attestable-provenance pillar's job (PROJ-006 #2, which also subsumes the
reserved-tag forgery gap: `agent:`/`model:`/`session:` smuggled through freeform
`tags` are validated for *shape* but never for *truth*). Sizing `MaxTagLen` to
accommodate long model identifiers would therefore be sizing it against a
population it does not govern — the 64 here is derived from the 21-byte observed
maximum of tags this cap actually sees.

## Consequences

- **Positive:** a user can write the field every digest reads without compressing
  it. `impact`'s median stops sitting at 89% of its ceiling.
- **Positive:** the tags cap now catches what it was meant to catch. Using nine
  short tags is legal; one 65-byte tag is not — the inverse of the shipped
  behavior, and the error names the offending tag rather than the joined string.
- **Positive:** `capture.Validate` runs on **every** ingress path — `cli/add.go`
  (×2), `cli/add_json.go`, MCP `brag_add`, and now `cli/edit.go` — and that became
  true only *after* the caps were fixed, never before.
- **Positive:** each cap is defined once. `cli/project.go:162`'s hard-coded
  `len(name) > 64` now references `capture.MaxProject`, closing the duplication
  whose own comment asked a reader to keep two numbers in agreement (DEC-036's
  soft-match invariant, DEC-017).
- **Positive:** no schema change, no migration, no backfill.
- **Negative (accepted):** 3 impacts and 29 titles remain over cap. Editing the
  offending field forces a fix. For the 29 titles — pasted brag bodies — that is
  the correct outcome; for the 3 impacts it is a real, if rare, friction.
- **Negative (accepted):** the worst-case memory line more than doubles, from 157
  to 363 estimated tokens (18.2% of the default 2000-token budget). See the
  re-derivation below for why the anti-starvation argument survives.
- **Negative (accepted):** `ValidateChanged` means the edit path's behavior depends
  on the *stored* value, so the same buffer can be accepted on one entry and
  rejected on another. That is inherent to grandfathering-by-changed-field and is
  the price of not adding a column.
- **Neutral:** `capture.MaxTags` is a **removed export**. Nothing outside the repo
  consumes it, but it is a breaking change to the package's surface and to
  `docs/brag-entry.schema.json`, whose `tags.maxLength` is deleted — a JSON Schema
  `maxLength` on a comma-joined string cannot express "64 per tag, 32 tags". The
  binary remains the authoritative validator, as that schema already states.
- **Neutral:** DEC-012's cap *values* are superseded; its stdin schema shape
  (six fields, comma-joined tags, server-owned keys dropped) is untouched.

### The DEC-044 re-derivation

The line shape (DEC-044 sub-decision 5) is
`- <id> <YYYY-MM-DD> [<project>/<type>] <title> — <impact>`. With `—` at 3 bytes
UTF-8 and `<id>` at int64 max (19 digits):

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

`ceil(1450/4)` = **363 tokens** (was `ceil(626/4)` = 157). The old terms reproduce
DEC-044's published 626 / 157 exactly, which confirms the 19-digit id assumption
that DEC left implicit.

**The anti-starvation argument, restated.** At 363 tokens a worst-case entry
consumes 18.2% of the 2000-token default, leaving 82% for the rest of the slice;
and DEC-044 sub-decision 3's skip-and-continue means an entry that does not fit is
skipped rather than truncating everything behind it. Catastrophic starvation
remains impossible. The argument survives the change intact.

**Three pre-existing errors in DEC-044 that the re-derivation surfaced.** All are
wrong *today*, independent of this decision; they are corrected here because this
is where the math is re-derived, and leaving them is the silent invalidation
STAGE-018 set out to prevent.

- **(i) The 626-byte bound is already false of the live corpus.** The measured
  maximum rendered line is **1483 bytes / 371 tokens** (entry 172), because the
  over-cap rows predate SPEC-064's unification. The new theoretical bound of 1450
  bytes is therefore *below today's observed maximum*: this change **tightens**
  the true bound while widening the caps that bite in ordinary use.
- **(ii) "a real corpus averages ~18 tokens per entry" is the test fixture's mean.**
  DEC-044 sub-decision 3's own 8-entry fixture has per-entry costs 27, 14, 15, 10,
  27, 15, 22, 13 → mean **17.875 ≈ 18**. The measured corpus mean is **~92
  tokens/entry**.
- **(iii) "2000 buys ≈110 entries" is wrong by ~4.4×.** Measured against the real
  corpus, `brag memory` at the default budget reports `Included: 25`,
  `Skipped: 175`, `Estimated: 1991`.

**`memory.DefaultBudget` stays at 2000; its stated rationale is corrected.** The
default's defense is *displacement* — 2000 is ~10% of the measured ~20k-token
session-opening ritual it exists to replace, and ~1% of a 200k context window —
and this decision changes neither. What was false is the *yield* claim attached
to it, and 25 best-ranked entries is still a useful slice. Changing a shipped
default is a behavior change to a one-release-old feature and belongs in its own
spec; DEC-044's Revisit-(b) already names the trigger, and the calibration data
is printed in every envelope. Correcting the rationale is what stops a future
reader re-deriving from a false number.

## Validation

**Right if:**
- A user writes a normal-length `impact` without hitting a wall — the median stops
  tracking the ceiling. Re-measure `impact` p50 against `MaxImpact` after a
  quarter of capture; a median above ~40% of the cap means this is not settled.
- Nine short tags validate; one 65-byte tag does not — pinned by
  `"a 9-tag joined string over 64 bytes now passes"` and `"a single tag over 64
  bytes is rejected and the error names that tag"`.
- An entry with a grandfathered 1290-byte `impact` can have its title edited, and
  the stored impact is byte-identical afterwards — pinned by the `edit_test.go`
  grandfathering test, which is also the test that fails if the edit-path wiring
  is ever reordered ahead of the changed-field rule.
- The worst-case line is exactly 1450 bytes / 363 tokens, computed from the
  constants — pinned by a `internal/memory` test, so any future cap change that
  does not move this DEC's number turns the build red.
- `cli/project.go` holds no numeric literal for the project cap, and its test
  references `capture.MaxProject` rather than a literal.

**Revisit if:**
- **(a) `impact` at 1024 starts being hit in ordinary use.** First move: re-measure
  p90/p99 as done here. If the field is genuinely becoming multi-sentence prose,
  the question is whether `impact` and `description` still have distinct jobs —
  not whether to add another 768 bytes.
- **(b) The 29 grandfathered titles become friction** rather than a backlog of
  malformed rows. Then the answer is a one-off cleanup pass over that project's
  entries, not a wider `MaxTitle`.
- **(c) `MaxTagCount` is hit by a legitimate workflow.** 32 against a measured max
  of 11 is generous, but an importer or a bulk tagger could exceed it. The count
  is one exported constant.
- **(d) The memory slice's yield (25 entries at the default) proves too small in
  real auto-loading use.** That is DEC-044's Revisit-(b), now with a corrected
  baseline: raise `DefaultBudget`, do not shrink the caps back.
- **(e) A future ingress path appears** (an import command, an MCP edit tool).
  Then it calls `Validate` or `ValidateChanged` at its boundary like the five that
  exist; the rule is the path list, and it is enumerated in SPEC-075's Outputs.

**Confidence 0.85**, decomposed honestly:
- **The tags re-shape (per-tag + count) — 0.95.** The measurement is unambiguous:
  every over-cap string is many-short-tags, the longest single tag in the corpus
  is a third of the proposed per-tag cap, and DEC-015 already removed the model
  that justified a joined-string cap. This is a defect fix, not a judgment call.
- **Grandfathering by changed-field — 0.90.** It needs no schema change,
  converges, and has a clean failure story. The residual asymmetry (same buffer,
  different verdict per entry) is real but bounded and stated.
- **The DEC-044 re-derivation — 0.90.** The arithmetic reproduces the published
  626/157 exactly from the old caps, which is strong evidence the model is being
  re-derived correctly rather than reinvented; and the three corrections were each
  verified against the shipping binary, not inferred.
- **`MaxImpact = 1024` — 0.80.** Grounded in p99 and in a stated headroom class,
  but "how long should an outcome sentence be" is a product judgment, and 1024 is
  a round number near the derived value rather than the derived value itself.
- **`MaxTitle = 256` — 0.75.** The softest number here. The derivation leans on
  excluding a single project as malformed, which is a judgment about intent made
  from length alone. If those 33 entries turn out to reflect how the user actually
  wants to write titles, 256 is wrong and the honest fix is a wider cap plus an
  admission that `title` and `description` overlap.
- **`DefaultBudget` unchanged — 0.75.** Correct on the displacement argument, but
  a yield of 25 rather than the believed 110 is a large enough surprise that the
  right default may genuinely be higher; this decision deliberately does not
  change a shipped default while correcting the number that would inform it.

## References

- Related specs: **SPEC-075** (emits and implements this DEC), SPEC-064 (shipped —
  unified per-path validation; created the situation), SPEC-073 (shipped — the
  memory slice and DEC-044), SPEC-074 (shipped — the MCP push surface inheriting
  the line shape), SPEC-046 (shipped — DEC-027's reserved tags that
  `validateReservedTags` checks).
- Related decisions: **DEC-044** (the budget model re-derived above — its lines
  136, 443 and 71–74 are corrected by this decision), **DEC-012** (introduced the
  caps with no rationale; its *values* are superseded here, its stdin schema shape
  is not), **DEC-015** (tags normalised into a table with no length constraint —
  why a joined-string cap has no model), **DEC-004** (the comma-joined wire format,
  unchanged), **DEC-036** / **DEC-017** (the soft-match invariant
  `cli/project.go:162` reasons about, and why both sides count bytes), **DEC-027**
  (the reserved `cost:`/`tokens:` tag checks inside `Validate`, untouched).
- Discussions: `guidance/questions.yaml` →
  `capture-caps-were-inherited-not-derived` (the originating question, the
  347-entry measurement, and the ordering risk); STAGE-018 (the v0.6.0 gate and
  its success criteria).
- Measurement: 359-entry snapshot of `~/.bragfile/db.sqlite`, provenance tags
  stripped, byte counts via `LENGTH(CAST(x AS BLOB))`; worst-case and yield
  figures verified against the built binary (`brag memory`), not computed only in
  SQL.
</content>

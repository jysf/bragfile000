# PROJ-001 backlog — deferred ideas

Ideas surfaced during PROJ-001 discussions that were consciously
deferred. Each entry names its source, the reason it was deferred,
a trigger for revisiting, and a short implementation sketch so
nothing is lost but the project stays focused.

**Intake rules**

- Items land here when explicitly deferred during design/framing/
  reflection. Never silently.
- Items leave here by being pulled into a framed stage (promote to
  a stage's pre-framing notes), promoted to a future project (e.g.
  PROJ-002 — AI assist), or deleted when no longer relevant.
- Don't treat this as a wishlist. Items here are documented *reasons
  not to ship now*, not aspirations.

---

## Attachments directory (`brag add --attach <file>`)

- **Source:** 2026-04-22 external-Claude evaluation dump; user-triaged
  to backlog.
- **Reason deferred:** Heavy scope — file storage layout,
  attachment lifecycle, export handling (copy vs link vs skip),
  schema change. User's primary workflow is code work, not
  design/screenshot-heavy work.
- **Revisit when:** A real use case arises where a screenshot or
  graph-shaped impact evidence would have meaningfully helped.
- **Sketch:** `brag add --attach screenshot.png` copies the file
  into `~/.bragfile/attachments/<id>/`. New `entry_attachments(
  entry_id, kind, filename)` table or hidden JSON column. Exports
  decide per-format: markdown inlines images; sqlite copies
  attachments dir alongside DB; JSON writes relative paths.

## LLM-backed `brag summarize --since <range>` (with Ollama)

- **Source:** 2026-04-22 external-Claude evaluation dump. Also
  explicitly listed in PROJ-001 brief "Explicitly out of scope"
  section from day one.
- **Reason deferred:** PROJ-001 is about capture + retrieve +
  distribute. LLM integration is a different product-shape
  concern (API-key management, network boundary, prompt design,
  model-choice strategy). Warrants its own project.
- **Revisit when:** STAGE-003/004/005 have shipped, user has real
  corpus to summarize from, ready to design the API-key strategy
  + offline-first story. Frame as **PROJ-002 — AI assist**.
- **Sketch:** `brag summarize --since 90d [--model
  ollama:llama3|openai:gpt-4o]` reads entries via existing
  `Store.List(ListFilter{})`, prompts LLM to draft review section,
  writes markdown. Keys via env or config file. Ollama first-class
  for offline/private workflows. Output feeds into a future
  `brag export --format review`.

## goals / levels — rubric mapping

- **Source:** 2026-04-22 external-Claude evaluation dump;
  user-triaged to backlog.
- **Reason deferred:** Speculative feature for promo-prep use
  cases. User's solo workflow doesn't need rubric mapping yet.
  Adding schema for unused structure is complexity without
  benefit.
- **Revisit when:** A specific review cycle with defined rubrics
  (promo packet, annual self-review, level-up evaluation) is
  imminent AND manual categorization in description/tags has
  proven insufficient.
- **Sketch:** `goals(id, name, description, rubric_source)` table.
  `entry_goals(entry_id, goal_id, weight)` join. Commands: `brag
  goals add`, `brag goals list`, `brag list --goal "L5-scope"`.
  Optionally import rubric from YAML/JSON.

## `peer` / `quote` dedicated field

- **Source:** 2026-04-22 external-Claude evaluation dump.
- **Reason deferred:** Convention in description works today
  (users can write `> "quote" — Alice`). A dedicated column adds
  schema surface for a feature that's usable via convention.
- **Revisit when:** User finds themselves wanting to filter/export
  quotes separately, or AI consumers of JSON export want
  structured access to positive-feedback artifacts.
- **Sketch:** New nullable columns `quote` + `quote_author`.
  Flags: `brag add --quote "..." --quote-by "teammate"`. Render in
  `show` as dedicated section under description. Filter:
  `brag list --has-quote`.

## Full `visibility` enum (public/internal/confidential)

- **Source:** 2026-04-22 external-Claude evaluation dump.
- **Reason deferred:** The simple `--exclude-tag internal`
  approach (captured for STAGE-004) handles ~80% of real
  confidentiality needs with zero schema change. Full enum adds
  column + migration + export-filter complexity; not justified
  until the simple filter proves insufficient.
- **Revisit when:** Using `--exclude-tag` in real review prep
  reveals a concrete case where per-entry categorical visibility
  (public/internal/confidential) would be meaningfully better
  than tag-based filtering.
- **Sketch:** Add `visibility` TEXT column (constrained:
  `CHECK (visibility IN ('public', 'internal', 'confidential'))`).
  Default `public`. Flag `brag add --visibility internal`.
  Export: `brag export --visibility public`.

## `brag export --exclude-tag <tag>` redaction filter

- **Source:** 2026-04-22 ideas dump; initially held for STAGE-004
  polish pass, reshuffled to backlog same day to keep STAGE-004
  tight.
- **Reason deferred:** Solves a real workflow (redact
  `internal`-tagged entries before sharing an export with a
  manager or posting a blog excerpt), but `grep -v` over a
  markdown export handles the 80% case for a solo user. Wait
  until the pain is real before adding export-flag surface.
- **Revisit when:** User does at least one review-doc prep where
  `grep -v ^tag:.*internal` or equivalent manual redaction
  proves insufficient or error-prone — then promote to a small
  spec in STAGE-004 or a later polish stage.
- **Sketch:** `brag export --format markdown --exclude-tag
  internal` filters entries whose `tags` field contains the
  given token (same comma-split semantics as `list --tag`). Flag
  repeatable: `--exclude-tag internal --exclude-tag draft`.
  Apply filter before rendering, not after. Pairs naturally with
  full `visibility` enum (see above) if that ever lands.

## Git-context auto-capture on `brag add`

- **Source:** 2026-04-22 ideas dump; initially held for STAGE-004
  polish, reshuffled to backlog same day.
- **Reason deferred:** Nice convenience — auto-populate project
  from `git remote get-url origin` repo name and stash current
  branch / last commit SHA into description or a new field. But
  the user's `brag` invocations span contexts beyond git repos
  (reading, meetings, mentoring), so auto-capture risks
  wrong-tagging entries that aren't code work. And the manual
  `-p project-name` flag is already 10 characters. Low
  payoff-to-complexity ratio.
- **Revisit when:** User captures 20+ brags across code work and
  observes a consistent pattern of forgetting to set `--project`
  for repo-scoped entries, OR a dedicated AI-agent integration
  (Claude session-end hook) surfaces a concrete need for
  structured git context inside the entry body.
- **Sketch:** `brag add` (no `--project` set) checks
  `git rev-parse --show-toplevel` from CWD; if inside a repo,
  derive project from `basename $(git remote get-url origin)`
  (stripped of `.git`). Optional `--git-context` flag appends
  `[branch @ sha]` to description. Escape hatch: `--no-git` or
  `BRAGFILE_NO_GIT=1`. Keep off by default; opt in via config
  key `autocapture.git = true`.

## `--link` / `--refs` multi-valued field for PR/issue/doc links

- **Source:** 2026-04-22 ideas dump; initially held for STAGE-004
  polish, reshuffled to backlog same day.
- **Reason deferred:** Links are genuinely useful in review prep
  ("here's the PR that shipped it"), but today users inline them
  in `--description` or `--impact` as plain markdown, and that
  works. Adding a dedicated field means a schema migration
  (either a comma-joined TEXT column like `tags` or a proper
  `entry_links` join table), plus rendering decisions in `show`
  / `list --pretty` / `export` / JSON output — lots of surface
  for a problem that convention already solves.
- **Revisit when:** User wants to filter/list entries by
  presence of links (`brag list --has-link`), or a review-doc
  export needs links rendered as a structured "References:"
  section rather than inlined prose. Or an AI consumer of JSON
  export asks for structured access to linked artifacts.
- **Sketch:** Start simple — new `links` TEXT column,
  comma-joined URLs (same shape as `tags`, DEC-004). Flag
  `brag add --link https://... --link https://...` (repeatable,
  joined into the column). Render in `show` as a bulleted
  "Links:" section under description. Escalate to a proper
  `entry_links(entry_id, url, label)` table only if label
  support or per-link metadata becomes a real need.

## NDJSON / array-batch stdin for `brag add --json`

- **Source:** STAGE-003 framing 2026-04-23 (SPEC-017 design
  scope).
- **Reason deferred:** MVP ships single-object stdin only. Batch
  import is a legitimate future workflow but the shape decision
  (NDJSON vs. array, transactional vs. best-effort, error
  reporting per-line) warrants its own spec rather than bolting
  onto SPEC-017.
- **Revisit when:** A real bulk-import workflow appears — e.g.
  importing from a previous tool's export, or a Claude agent
  batching a session's suggested entries in one call.
- **Sketch:** `brag add --json --batch` reads NDJSON from
  stdin, one entry per line; commits transactionally
  (all-or-nothing) or best-effort (`--continue-on-error`).
  Prints inserted IDs one per line. Array input (`[{...},
  {...}]`) as a secondary mode if demand is clear.

## Lenient-accept mode for `brag add --json`

- **Source:** STAGE-003 framing 2026-04-23 (SPEC-017 / DEC-012
  scope).
- **Reason deferred:** Strict-reject-unknown-keys is the MVP
  default because it catches typos (`titl`, `descripton`) before
  they become silently-missing entries. Lenient mode is the
  opposite tradeoff — useful when piping from tools that emit
  extra fields — but only worth adding if strict proves annoying
  in real use.
- **Revisit when:** A real pipeline emerges (AI agent, another
  tool's export) that emits schema-adjacent-but-extra fields,
  AND the user is willing to accept silent field loss as a
  tradeoff.
- **Sketch:** `brag add --json --lenient` ignores unknown keys
  instead of rejecting. No schema evolution story; just skip
  unrecognized keys without error.

## JSON output envelope

- **Source:** STAGE-003 framing 2026-04-23 (DEC-011 scope).
- **Reason deferred:** MVP ships naked JSON array so `jq '.[]'`
  stays trivial and round-trip through `brag add --json` is
  clean. An envelope (`{generated_at, count, filters, entries:
  [...]}`) adds provenance/metadata that a downstream AI
  consumer might want, but nobody has asked yet.
- **Revisit when:** An AI consumer (Claude agent, summary tool,
  analytics pipeline) asks for export-time metadata inside the
  JSON payload, OR a use case emerges where two exports need to
  be correlated by timestamp or filter.
- **Sketch:** `brag export --format json --envelope` (or
  `--wrap`) flag wraps the array in an object with
  `generated_at` (RFC3339), `count`, `filters` (object echoing
  what was passed), `entries` (the array as-is). Consumer can
  unwrap with `jq .entries`.

## `--compact` / non-pretty JSON output

- **Source:** STAGE-003 framing 2026-04-23 (DEC-011 scope).
- **Reason deferred:** MVP ships pretty-printed (indent=2) JSON
  because readability matters more than bytes-on-wire at
  personal-corpus scale. A compact flag is trivial to add but
  zero payoff today.
- **Revisit when:** A pipe consumer is measurably slowed by
  indentation, OR an export writes to a size-constrained
  destination (unlikely at personal scale).
- **Sketch:** `--compact` flag on `brag list --format json` and
  `brag export --format json` toggles `json.Marshal` vs.
  `json.MarshalIndent`. One line of Go.

## Filtered SQLite export

- **Source:** STAGE-003 framing 2026-04-23 (SPEC-016 scope).
- **Reason deferred:** `VACUUM INTO` is one SQLite call and
  ships the full-DB-portable-backup use case cleanly. Filtered
  sqlite (e.g. "last 90 days of project X as a portable DB")
  needs fresh-DB + migration-application + INSERT-SELECT —
  meaningful new code and a migration-ordering coupling.
  Downstream consumers of filtered slices are better served by
  JSON or markdown exports today.
- **Revisit when:** A concrete user need emerges for a filtered
  portable SQLite file (e.g. a multi-laptop workflow where only
  a subset belongs on a machine; a committee that wants
  queryable data but not the full corpus).
- **Sketch:** `brag export --format sqlite --since 90d --out
  slice.db` opens an empty SQLite file, runs the same embedded
  migrations the main `Store.Open` runs, then `INSERT INTO
  new.entries SELECT * FROM old WHERE <filters>`. Handle
  `schema_migrations` table population so the slice opens
  cleanly with current `brag` binary.

## Table of contents in markdown export

- **Source:** STAGE-003 framing 2026-04-23 (SPEC-015 / DEC-013
  scope).
- **Reason deferred:** Markdown headings (`## <project>` + `###
  <title>`) are already scannable with any markdown viewer's
  outline pane. A TOC block adds code for modest payoff at
  typical quarterly-export sizes (~50–100 entries).
- **Revisit when:** Export sizes grow past ~200 entries, OR a
  downstream use case needs stable anchor links (e.g. posting
  exports to a wiki that builds TOCs from headings).
- **Sketch:** `brag export --format markdown --toc` inserts a
  "## Table of Contents" block after the provenance/summary with
  `- [title](#anchor)` lines per entry. Slugify titles for
  anchors; disambiguate collisions with a numeric suffix.

## `--group-by <field>` in markdown export

- **Source:** STAGE-003 framing 2026-04-23 (SPEC-015 / DEC-013
  scope).
- **Reason deferred:** MVP ships group-by-project as the default
  (user ask) with `--flat` as the escape. Multi-axis grouping
  (by type, by tag, by month) is polish work that didn't make
  the must-have cut.
- **Revisit when:** A real review-doc workflow benefits from
  grouping by type (e.g. a promo packet organized by "shipped /
  fixed / learned" buckets) OR by time bucket (a monthly
  retrospective).
- **Sketch:** `brag export --format markdown --group-by type`
  (or `tag`, or `month`). Extends the partition function in
  `internal/export`; keeps `--flat` as the no-grouping escape;
  default stays group-by-project. Document the valid values in
  `--help`.

## `--template <path>` for custom markdown rendering

- **Source:** STAGE-003 framing 2026-04-23 (SPEC-015 scope).
- **Reason deferred:** The default DEC-013 shape covers the four
  named use cases (retro, quarterly review, promo packet, resume
  update). Custom templates introduce stdlib `text/template`
  surface area and template-distribution questions (ship with
  built-ins? load from `~/.bragfile/templates/`?). Not justified
  until a user has a concrete template they can't express via
  the flag matrix.
- **Revisit when:** A user wants to render exports for a
  tool-specific format (e.g. a specific review system's markdown
  dialect) that the flag matrix can't produce.
- **Sketch:** `brag export --format markdown --template
  ~/.bragfile/templates/promo.md.tmpl` parses the file as a Go
  `text/template` with `{{.Entries}}`, `{{.Summary}}`,
  `{{.Provenance}}` bindings. Ship two or three built-ins in the
  binary via `embed.FS` as reference examples.

## Anything unreleased from STAGE-004 at distribution time

- **Source:** User decision 2026-04-22 — STAGE-004 is provisional;
  if STAGE-003 delivers sufficient utility, STAGE-004 items drop
  to backlog in favor of going directly to STAGE-005 distribution.
- **Reason deferred (placeholder):** To be filled in when STAGE-004
  dissolves.
- **Revisit when:** After initial release — real usage reveals
  which STAGE-004 items are actually missed.
- **Sketch:** Populate this section when the decision is made.
  Candidates from the STAGE-004 sketch in `brief.md`:
  `brag summary`, emoji passes 1–4, Claude session-end hook,
  `brag remind`, `brag stats`, `brag review --week`.

---

## Removed / delivered — keep the list honest

*When an item is pulled into a stage and ships, list it here with
the spec IDs that delivered it. Keeps the backlog history tractable
and proves items actually move out, not just in.*

(none yet — this section activates once STAGE-003 ships and
promoted items become traceable.)

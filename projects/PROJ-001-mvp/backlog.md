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

## `brag export --format sqlite` (full-DB VACUUM INTO)

- **Source:** STAGE-003 scope-tightening decision 2026-04-23
  (post-SPEC-013). Originally framed as SPEC-016; the SPEC-016
  slot number is preserved in the stage file as a deferral
  marker (not renumbered).
- **Reason deferred:** The portable-backup use case the brief
  named is already handled by `cp ~/.bragfile/db.sqlite
  backup.db` — documented in the tutorial and works today with
  zero new code. `VACUUM INTO`'s marginal wins over `cp` are
  (a) defragmentation on export (cosmetic) and (b) WAL-flushed
  consistency when another `brag` process is writing
  concurrently (real but narrow — in a single-user personal CLI,
  concurrent writers are effectively never). Not worth a spec
  until one of those marginal wins turns into a real need.
- **Revisit when:** A concrete workflow emerges where `cp`'s
  consistency guarantees are insufficient (a daemon variant of
  `brag`; a multi-process write pattern; a shared-filesystem
  backup pipeline), OR the defragmentation angle becomes user-
  visible (DB file grows pathologically large from churn). Neither
  is imminent for a personal CLI at current usage.
- **Sketch:** `brag export --format sqlite --out backup.db`
  executes `VACUUM INTO '<path>'` via the existing `Store`
  connection. `--out` required (binary output to stdout is
  hostile). Filter flags rejected with a `UserErrorf` pointing
  at `--format markdown` / `--format json` for filtered slices.
  Round-trip smoke: `brag --db backup.db list` returns the same
  entries as the source. One `VACUUM INTO` call, no schema
  duplication, no migration ordering concerns.

## Filtered SQLite export

- **Source:** STAGE-003 framing 2026-04-23 (originally SPEC-016
  scope); reason-for-deferral shape overlaps with the full-DB
  entry above, which was itself deferred later the same day.
- **Reason deferred:** Filtered sqlite (e.g. "last 90 days of
  project X as a portable DB") needs fresh-DB + migration-
  application + INSERT-SELECT — meaningful new code and a
  migration-ordering coupling. Downstream consumers of filtered
  slices are better served by JSON or markdown exports today.
  Now that the full-DB variant is also deferred (see above),
  both would revive together if sqlite export becomes real.
- **Revisit when:** A concrete user need emerges for a filtered
  portable SQLite file (e.g. a multi-laptop workflow where only
  a subset belongs on a machine; a committee that wants
  queryable data but not the full corpus). If the full-DB
  variant revives first, this entry likely follows as the
  filter-aware upgrade.
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

## Emoji decoration passes 1–4

- **Source:** STAGE-003 pre-framing notes (2026-04-21); kept on
  STAGE-004 list through 2026-04-23; cherry-picked OUT on
  2026-04-24 post-STAGE-003 review.
- **Reason deferred:** User wants emoji somewhere in the tool but
  doesn't love this specific palette/scope. The four-pass plan
  (Pass 1 stderr feedback prefixes ✏️/🗑/—/✗; Pass 2 type icons
  in `brag show`; Pass 3 type icons in `brag list --pretty`;
  Pass 4 NO_COLOR + TTY auto-detection) was framed before the
  user had real-usage signal of what would feel right. Shipping
  blind is rule-ahead-of-problem.
- **Revisit when:** User picks a palette + shape they actually
  want. Could be triggered by a specific moment of "I wish this
  output had X icon here," or by a deliberate sit-down to design
  the palette. Likely a 1–2 day pickup once the design is set.
- **Sketch:** Pass 1 is ~5 lines (prefix `fmt.Fprintln(stderr,
  …)` calls in edit/delete/confirmation paths). Pass 2 adds an
  `emojiForType()` helper consumed by `internal/export.RenderEntry`.
  Pass 3 adds a `--pretty` flag on `brag list` that bundles `-P`
  + emoji. Pass 4 adds a TTY/NO_COLOR check that gates the other
  three. Each pass is independent; user can cherry-pick one or
  two without the rest.

## `brag remind` nudge command

- **Source:** STAGE-003 pre-framing notes (2026-04-21); kept on
  STAGE-004 through 2026-04-23; cherry-picked OUT on 2026-04-24.
- **Reason deferred:** User has been logging consistently without
  one (~20 entries across two weeks, no missed days). Habit
  enforcement is a problem the user doesn't currently have.
- **Revisit when:** First week with zero entries — that's the
  signal that habit reinforcement would have helped.
- **Sketch:** Pull-shape: `brag remind` checks last entry's
  `created_at`; if older than N days (configurable via
  `~/.bragfile/config.yaml` or `--days N`, default 3), prints
  `"⏰ no entries in 4 days — last was: '<title>' on <date>"` to
  stderr; otherwise silent (exit 0). User adds to shell prompt
  hook (`zsh precmd`) or daily cron. Push-shape (launchd / cron
  with system notification) is a heavier follow-on if the pull
  shape isn't enough.

## `brag add --at <date>` backdating flag

- **Source:** External Claude review 2026-04-24 surfaced as
  high-value-low-cost during STAGE-004 planning; user-deferred
  to backlog 2026-04-24 to keep STAGE-004 at 3 specs.
- **Reason deferred:** S-sized spec (~25-40 LoC + 5-7 tests +
  small doc updates), genuinely useful, but user has been
  logging at end-of-day reliably for 2+ weeks — no
  Friday-recapping-Tuesday pain has surfaced yet. Adding now
  would push STAGE-004 from 3 specs to 4 (~30% timeline bump)
  for a feature whose value is invisible until needed.
- **Revisit when:** First time the user catches themselves
  wanting to log a brag from a previous day and finds it
  awkward to do via SQL or `--json` with a constructed
  `created_at` (which is currently tolerated-and-ignored per
  DEC-012 anyway, so backdating via stdin doesn't work today
  either — this would be the only path).
- **Sketch:** `brag add --at 2d` (2 days ago) or
  `brag add --at 2026-04-22` (specific date). Reuses the
  existing date parser SPEC-007 built for `--since`. Touches
  `Store.Add` to respect non-zero `CreatedAt` (defaults to
  `time.Now()` when zero, current behavior). One small design
  decision required: if `brag add --json --at 2d` runs AND
  stdin JSON has a `created_at` field, `--at` flag wins
  (more explicit). Document in DEC-012 follow-up note. ~0.75
  day end-to-end.

## Anything unreleased from STAGE-004 at distribution time

- **Source:** User decision 2026-04-22; updated 2026-04-24 after
  cherry-pick. STAGE-004 was provisional; per the
  2026-04-22 escape hatch, items not chosen for STAGE-004 land
  here.
- **Reason deferred:** User-filtered through "will I actually use
  this?" 2026-04-24. Three items (summary, review --week, stats)
  promoted to STAGE-004; six items moved here (emoji 1–4, remind
  — both with their own dedicated entries above; Claude
  session-end hook moved to STAGE-005 as a distribution asset).
- **Revisit when:** Each entry above has its own concrete
  trigger. This umbrella entry exists for traceability —
  individual items live in their own backlog entries.
- **Sketch:** N/A — this is a pointer, not a deferred item.
  See the dedicated entries above for the actual deferred work.

## macOS code signing + notarization (Apple Developer ID)

- **STATUS 2026-06-19: NOW IN SCOPE.** After the v0.2.0 Homebrew release,
  the user hit the macOS install friction first-hand and decided to address
  it "even if it means spending $100." Pursue as a standalone effort — a
  **v0.2.1 "macOS distribution hardening"** mini-stage (or early PROJ-003).
  External lead time on Apple enrollment, so not coupled to anything else.
  NOTE: notarization fixes ONLY the Gatekeeper prompt — it does NOT remove
  the `brew trust --cask` step (see the separate item below); they are
  distinct frictions.
- **Source:** 2026-05-11 Phase 2 ship of bragfile v0.1.0 — smoke-test
  install hit `"Apple could not verify 'brag' is free of malware"`
  on macOS Gatekeeper. Workaround documented in README ("macOS
  Gatekeeper note") and AGENTS.md §4 (xattr -dr quarantine).
- **(Historical) Reason deferred:** Apple Developer Program is $99/year ongoing
  + ~half-day of focused work (account approval, Developer ID
  Application certificate, app-specific password, goreleaser
  `signs:`/`notarize:` blocks, release.yml updates to run signing
  on a macos-latest runner with new secrets MACOS_CERTIFICATE +
  MACOS_CERTIFICATE_PASSWORD + APPLE_ID + APPLE_PASSWORD +
  APPLE_TEAM_ID). For a personal project shipping for learning
  value with no marketing push (per PROJ-001 brief), the cost-
  benefit isn't there yet. The xattr workaround is fine for the
  small audience this binary actually reaches.
- **Revisit when:** Either (a) actual adoption materializes and the
  Gatekeeper UX cliff becomes a real friction point with real
  users, (b) the user enrolls in Apple Developer for another
  reason (iOS / Mac App Store work) and the cert is already
  available, or (c) bragfile is rolled into a paid distribution
  channel where notarization is a hard requirement.
- **Sketch:**
  1. Enroll in Apple Developer Program (~$99/year, ~1-2 day
     approval). Generate a Developer ID Application certificate
     via developer.apple.com. Generate an app-specific password
     for notarytool via appleid.apple.com.
  2. Export certificate as `.p12`, base64-encode it, add as
     repo secret `MACOS_CERTIFICATE` (plus
     `MACOS_CERTIFICATE_PASSWORD`, `APPLE_ID`, `APPLE_PASSWORD`,
     `APPLE_TEAM_ID`).
  3. Add `signs:` block to `.goreleaser.yaml` invoking
     `codesign --options runtime --sign "Developer ID
     Application: <name>" {{ .Path }}` per darwin artifact.
  4. Add `notarize:` block invoking `xcrun notarytool submit
     ... --wait` post-sign, per archive.
  5. Update `.github/workflows/release.yml` to run on
     `macos-latest` instead of (or alongside) `ubuntu-latest`
     for darwin builds — signing requires macOS runner; linux
     builds can stay on ubuntu via a matrix split.
  6. Test: tag a `v0.x.y-rc` pre-release, verify the darwin
     archives pass Gatekeeper on a fresh Mac without the
     xattr workaround.
  7. Remove the "macOS Gatekeeper note" from README; update the
     AGENTS.md §4 lesson to "no longer applies as of vX.Y.Z."

## Homebrew 6.0 `brew trust --cask` friction (cask from third-party tap)

- **Source:** 2026-06-19 v0.2.0 install. On Homebrew 6.0.2, installing the
  cask failed with `Refusing to load cask jysf/bragfile/bragfile from
  untrusted tap jysf/bragfile`; the fix was a one-time
  `brew trust --cask jysf/bragfile/bragfile`. This is a Homebrew tap-source
  policy (arbitrary Ruby in cask defs), **distinct from** Gatekeeper/
  notarization — notarization will NOT remove it.
- **Done now:** README install section documents the `brew trust` step.
- **Open question:** is there a way to avoid making users run `brew trust`
  at all? See the formula-vs-cask item below — switching distribution from
  a cask to a formula may sidestep this entirely.

## Distribution: Homebrew formula vs. cask — RESOLVED, won't help

- **Question was:** would switching goreleaser from a cask to a formula
  avoid the Homebrew 6.0 `brew trust` gate?
- **Answer (researched 2026-06-19): NO.** Homebrew 6.0's tap-trust is a
  **tap-level** security policy — it applies to *both* tap-qualified
  formulae AND casks from any third-party tap (a third-party tap can run
  arbitrary unsandboxed Ruby regardless of artifact type). Switching
  cask→formula would NOT remove the `brew trust` requirement.
- **Implication:** the only ways to drop `brew trust` entirely are getting
  into official `homebrew-core` (high bar) or a non-tap distribution
  channel — neither worth it for a personal project. **Keep the cask**
  (goreleaser's recommended vehicle for prebuilt binaries) and document
  `brew trust` (done in README).
- **Net macOS-friction plan:** (1) README `brew trust` note [done];
  (2) ~~formula-vs-cask~~ [dead end — trust is tap-level]; (3) notarization
  [the $100 item above] still worthwhile — it removes the *Gatekeeper*
  prompt, a separate friction from `brew trust`.
- Sources: Homebrew 6.0.0 release notes (brew.sh/2026/06/11) and coverage
  noting trust covers "tap-qualified formulae and casks."

## govulncheck CI step

- **Source:** 2026-04-26 pre-distribution security review
  (`docs/reports/security/2026-04-26-pre-distribution-security-review.md`),
  prioritized fix list item 8.
- **Reason deferred:** Largely redundant with the Dependabot
  security alerts enabled the same day (item 2 of the report's
  GitHub Advanced Security section). Both consult the same Go
  vulnerability database (vuln.go.dev). Dependabot fires
  continuously on every advisory publication; govulncheck's
  incremental value is its call-graph reachability analysis —
  alerts only fire when the vulnerable function is *actually
  reachable* from the project's call graph. For a personal
  project with a small dep graph and Dependabot already
  surfacing every advisory, the noise reduction is marginal.
- **Revisit when:** Either (a) Dependabot starts producing
  enough advisory noise that reachability filtering would
  meaningfully help; (b) a future project with a much larger
  dep graph (e.g. PROJ-002 if it pulls in LLM client libraries)
  benefits from the call-graph filter; or (c) a CVE is missed
  because Dependabot's advisory-version match doesn't apply
  (rare).
- **Sketch:** Add a workflow step to `.github/workflows/ci.yml`:
  ```yaml
  - name: Run govulncheck
    run: |
      go install golang.org/x/vuln/cmd/govulncheck@latest
      govulncheck ./...
  ```
  Runs on every PR + push-to-main alongside the existing
  test/gofmt/vet steps. Fails CI on any reachable vulnerability.
  ~10s per run on a project this size. Confirmed clean against
  current main at 2026-04-26.

## WAL-safe backup recipe + automated daily backup

- **Source:** 2026-06-07 PROJ-002 coordinator session — user asked to
  back up the production DB before the v0.2.x migrations land; explicitly
  deferred to the end of PROJ-002. Belt-and-suspenders, not urgent: the
  prod DB at `~/.bragfile` is untouched by v0.2.x until a deliberate
  `brew upgrade` (dev/prod DB isolation holds through development).
- **Reason deferred:** Not blocking. The product-facing backup step is
  already in STAGE-008's doc scope (the v0.2.0 upgrade / CHANGELOG /
  tutorial backup workflow), and an unattended dev-machine backup is ops,
  not project code.
- **Revisit when:** STAGE-008 doc sweep, or PROJ-002 close — sooner if
  the user wants an automated daily backup during development.
- **Two distinct pieces:**
  1. **Doc-accuracy fix (STAGE-008):** the tutorial documents backup as
     bare `cp ~/.bragfile/db.sqlite backup.db` (the `brag export --format
     sqlite` entry below repeats the claim). That is **WAL-unsafe** if the
     DB ever runs in WAL mode — recent commits can sit unflushed in
     `db.sqlite-wal`. Upgrade the recipe to `sqlite3
     ~/.bragfile/db.sqlite ".backup 'backup.db'"` (or `VACUUM INTO`),
     which is WAL-aware and yields one consistent file. First confirm
     whether bragfile sets `PRAGMA journal_mode=WAL` at open — if it never
     does, bare `cp` is actually safe and this is a no-op, but the doc
     should still prefer `.backup` to be robust.
  2. **Automated daily backup (ops, optional):** a macOS launchd
     LaunchAgent running `sqlite3 ... ".backup"` to
     `~/.bragfile/backups/db-<date>.db` with keep-last-N pruning; trivial
     to remove after PROJ-002. Could be productized as a
     `scripts/backup-db.sh` shipped with the repo.
- **Related:** the `brag export --format sqlite` (full-DB `VACUUM INTO`)
  entry below — a built-in command version of the same backup need; would
  supersede the manual recipe if built.

---

# Impact & Fun — PROJ-003 candidate cluster

*Captured 2026-06-16 from a brainstorm during the STAGE-008 release hold.
These are post-v0.2.0 feature ideas, grouped because they share two
through-lines the user surfaced: (1) **passive surfacing** — the best
features fire/surface on their own; don't make the user remember to run a
command; (2) **impact over volume** — reward quality of captured outcomes,
not raw entry counts. Likely a PROJ-003 ("delight + impact"). The
polymorphic `taggings` + project substrate from PROJ-002 makes most of
these cheap. Decide scope at PROJ-003 framing, informed by real v0.2.0
dogfooding.*

**Product thesis (why this matters beyond a log).** Recording what you did
is table stakes. The differentiator — the thing that actually *helps*
people, not just documents them — is turning that record into **fun,
interesting, and impactful** output: stories, insight, a mirror on your
own trajectory. North star: **agent-native accomplishment memory** — agents
do the work *and* record why it mattered; you (and agents) read stories
back out. As agents do more of the work, "what got accomplished and why it
mattered" is exactly what they can capture and what humans most need
(reviews, promos, identity), and almost no tool is built for it. The fun /
interesting / impactful surfaces below are the point, not decoration.

## Milestone notifications on `brag add` (liked — "easy and great")

- **Idea:** When `brag add` crosses a threshold, print one celebratory
  line. Total count (10 / 25 / 50 / 100 / 250 / 500 / 1000), streak
  (7 / 30 / 100 days), per-project (10th / 50th brag on a project), and a
  quiet "first brag this week/today."
- **Design constraint (load-bearing):** stderr only, **TTY-only** (skip
  when piped / non-TTY) so `brag add --json` and scripted pipelines stay
  byte-clean — the project's stdout-is-data/stderr-is-humans spine.
- **Cheap:** the counts/streaks already live in `internal/aggregate`.
  Detect "crossed a threshold" from the post-insert total/streak.
- **Why first:** best delight-to-effort ratio; it is *passive* (fires on
  an action the user already takes).

## Impact surfacing & the quarterly "super-brag" (IMPORTANT — the headline)

- **The user's core ask:** pull number/date/impact/description over a
  period and turn it into a *story* for the quarter — gather impact over a
  month/quarter/year and produce insight / analysis / a synthesized
  "super-brag." Surfacing impact is "interesting and important."
- **Distinct from `export` — impact is the LEDE, not a field (2026-06-16).**
  `export` is good but treats `impact` as one column of nine. This feature
  must *lean on impact harder*: each entry renders impact-first (the
  outcome is the headline; title/date are supporting metadata), entries
  with no `impact` are de-emphasized or dropped, and the report is
  organized around outcomes — not a per-entry field dump. The test: it
  should read like an accomplishments narrative, not a table.
- **Why it matters:** `impact` is the most valuable field for perf
  reviews, promo packets, and self-narrative — but today it is only
  retrievable via raw `export`. Nothing *synthesizes* it.
- **Two complementary designs (do both):**
  1. **Rule-based `brag impact --quarter|--month|--year|--since`** — an
     impact-axis digest: every entry with a non-empty `impact`, grouped
     by project/period, rendered as a clean markdown report
     (date · title · impact · description). Performance-review /
     promo-packet ready. Mirrors `brag summary`/`review`'s DEC-014
     envelope, but axis = impact, not project.
  2. **AI-pipe "super-brag"** — emit a clean impact bundle + a synthesis
     prompt (exactly the `brag review` pattern) to pipe into an LLM for
     the narrative quarterly story. Rule-based core, AI via piping —
     consistent with the existing architecture (no built-in LLM).
- **Passive angle (per the user's preference):** surface an impact recap
  automatically — e.g. an end-of-quarter nudge, or an impact summary line
  folded into `brag stats` (which the user already runs) rather than a
  command they must remember.
- **Virtuous loop:** an impact report rewards *filling in* `impact`,
  which improves capture quality — ties to the stats reframe below.
- **Related:** pairs with the shipped "publish your brags to a website"
  tutorial recipe — a quarterly impact post writes itself.

## Storytelling — turn brags into stories (the differentiator)

- **Core reframe:** brags are *atoms*, stories are *molecules*, impact is
  the *bond*. Storage is solved; the unsolved, interesting problem is
  **composition** — assembling atoms into narrative.
- **`brag story` composition axes** (same corpus → different stories):
  - **project = arc** — a project's brags in time order have narrative
    shape (setup → friction/`learned`/blockers → `shipped` → impact);
    render the *journey*, not a list. Detect the genre (turnaround, grind,
    discovery).
  - **tag = capability / identity story** — all "performance" brags across
    projects = "how I became the person who makes things fast." The story
    people can't see about themselves.
  - **time = chapter** — quarter/year as a chapter (`wrapped` is really the
    story of your year).
  - **audience = reshaping** — `brag story --audience promo|resume|blog|1:1`
    — one corpus, many altitudes / voices / lengths.
- **Narrative intelligence / the mirror (the novel bit):** the app
  *notices things about you and tells you* — "your last 6 brags are all
  unblocking others — you're operating like a lead"; "this project reads
  like a turnaround"; "you undersold three of these — impact but no
  numbers." People forget and undersell their own work; an evidence-backed
  mirror on trajectory is emotionally resonant and a category most tools
  don't touch. Promo/review is just the most acute instance of "I can't see
  my own story."
- **Build path:** rule-based composition (ordering, grouping, genre
  heuristics) + AI-pipe for the prose, mirroring `brag review`. No built-in
  LLM.

## Impact depth — the "so what" ladder, density, compounding

- **The "so what" ladder.** Impact is often stated at the wrong altitude:
  *fixed the retry bug* → *cut on-call pages 40%* → *freed the team to ship
  X* → *protected $Y in renewals*. The high-leverage move is *helping climb
  the ladder* — e.g. an AI pass that keeps asking "and why did that matter?"
  up to business altitude.
- **Impact density (maybe the most behavior-changing metric).** Detect
  which brags carry *quantified* impact (number + unit, ideally
  before/after) vs. vague, and surface the gap: "12 of 40 brags this quarter
  are quantified; the other 28 are invisible in a promo packet." Rewards
  good capture; feeds the `stats` reframe.
- **Impact matures — make it a living field.** A bug fix's impact is known
  today; a framework everyone adopts compounds for months. Allow *appending
  to impact later* ("6 months on, three teams build on this") — compounding
  impact is the strongest promo evidence and nothing captures it today.
- **Atoms → molecules (the super-brag umbrella).** Detect clusters of small
  brags that sum to one headline accomplishment with aggregate impact — the
  difference between a changelog and a story.
- **Impact ÷ effort picks the genre.** Capturing rough effort lets the
  narrative choose its angle (smart-leverage win vs. heroic grind).

## `brag wrapped [year]` — shareable year-in-review (liked)

- **Idea:** A "Spotify Wrapped"-style annual recap: total brags, top
  projects/tags, longest streak, activity shape (sparkline), and an
  impact highlight reel.
- **Why:** highly shareable; pairs directly with the publish-to-website
  recipe (a "2026 wrapped" post). Reuses the aggregate + impact work.

## `brag achievements` — badges (liked, keep subtle)

- **Idea:** Earned badges — Prolific (50 brags), Polyglot (5+ projects),
  Consistent (4-week streak), Marathoner (100-day streak),
  Impact-Driven (N entries with impact filled). `brag achievements` lists
  earned + locked.
- **Caution:** highest gimmick-risk of the cluster. Keep tasteful,
  opt-in display; lean on *consistency/impact* badges over *volume* ones
  to stay aligned with quality-over-quantity.

## `brag stats` reframe + activity sparkline (liked, with corrections)

- **Drop the bad metric:** "most brags in a week" / any raw-volume
  leaderboard. The user is explicit: **quality > quantity** — volume is a
  bad achievement signal.
- **Add a sparkline (liked):** a unicode activity sparkline
  (`▁▂▃▅▇▂▁`) of brags per week/month — it shows *shape of activity*, not
  a ranked volume metric, so it sidesteps the quantity trap. Usable in
  `stats` and in `wrapped`.
- **Reframe stats around quality signals:** consistency (streaks),
  **impact density** (% of entries with a non-empty `impact`), and
  project/tag diversity — not totals-as-achievement.

## Resurfacing (`brag random` / `brag on-this-day`) — considered, DEPRIORITIZED

- **Verdict:** rejected as primary features. They require the user to
  *run* them, and the user values passive surfacing over manual recall.
- **Salvageable only passively:** if "on this day" appears, it should be
  surfaced automatically (e.g. a single line in `brag stats`, or a
  notify), never as a command the user must remember to invoke.

## `brag project` usability (informed by dogfooding)

- **Idea:** make the shipped `project` surface more usable. Confirmed-good
  from a v0.2.0 pilot; the surface is complete, but ergonomics can grow:
  - **project → entries shortcut.** Today you read a project's brags via
    `brag list --project X`; nothing in the `project` command points
    there. Option: `brag project show X` optionally tails recent entries,
    or add a `brag project log X`.
  - **macOS symlink papercut.** `/tmp`→`/private/tmp`-style symlinks can
    defeat cwd matching (`EvalSymlinks` was deliberately out of scope,
    DEC-019). Consider an opt-in resolve for symlinked project dirs.
  - **`project status` plain output** has a trailing empty column when
    `state_note` is blank (cosmetic).
  - **auto-fill is silent** by design; consider an opt-in hint.
- **Explicitly:** the user wants to *use* `brag project` on a real
  installed v0.2.0 binary before deciding what else it needs. Let
  dogfooding drive this list.

## Agent-driven capture — brags as a byproduct of work

- **The unlock (already proven in the user's practice):** the agent that
  did the work is best-positioned to log it — it holds title, change,
  project (cwd), tags, and can articulate impact *while context is fresh*.
  Capture-time impact > recalled-weeks-later impact, so this is where
  storytelling quality is actually won.
- **Ask for / confirm impact at capture (near-term, cheap, highest-leverage
  — user idea).** Update the installed agent assets —
  `examples/brag-slash-command.md`, the `scripts/claude-code-post-session.sh`
  hook, and the JSON-schema doc — to instruct the agent to **ask the user
  for the impact of the work, or confirm the impact, when adding a brag.**
  Single biggest lever for downstream storytelling; mostly a prompt/asset
  edit, so it could ship *ahead* of the larger PROJ-003 features.
- **Auto-derive everything except impact:** project from cwd (shipped),
  tags from languages/files touched, provenance (PR URL, commit SHA, files)
  attached automatically — leaving the agent one real job: articulate
  impact.
- **Provenance = sourced stories.** Agent-attached links footnote every
  brag ("shipped X (#123), which cut Y") — perfectly-sourced narrative
  material. Overlaps the parked `--link` / `--refs` item.
- **MCP server — OPEN QUESTION, not a slam dunk (user skeptical).**
  Considered: expose bragfile as a *local* MCP server
  (`brag_add` / `brag_search` / `brag_recent`) so agents capture via native
  tool calls. It **runs locally** (a process over stdio against the same
  `~/.bragfile` SQLite — no cloud). **But for shell-capable agents (Claude
  Code), `brag add --json` over the shell is already a clean, equivalent
  interface — the CLI *is* the agent API, which is why `--json` exists.**
  MCP's real value is narrow: reaching agent surfaces that *can't* run a
  shell (some sandboxed / desktop / hosted tools), typed tool
  discoverability, and `brag_search` as a recall tool. **Verdict:** lower
  priority; only worth building to reach non-shell agents — not a win over
  the shell for the primary workflow.
- **Concurrency wrinkle (real, technical).** `brag` runs in default SQLite
  journal mode — no WAL, no busy-timeout. One human + one agent is fine;
  *several agents writing at once can hit "database is locked."* If
  multi-agent capture becomes real, enable WAL + a busy-timeout (interacts
  with backup semantics — why the safety belt uses `VACUUM INTO`).

## Adjacent data — is bragfile a personal work-log substrate? (strategic)

- **The thread:** for an individual contributor who lives in the CLI (or
  uses AI apps that run the CLI), what accomplishment-adjacent data is
  *not* served by specialized tools and would be handy in a local SQLite +
  CLI? bragfile's polymorphic `taggings` + project model is already a
  generic capture/tag/relate/retrieve substrate; accomplishments are just
  one axis.
- **Candidate adjacent capture types:**
  - **Decisions / ADRs** — a lightweight decision log (this very repo runs
    on `DEC-NNN` files; dogfood it).
  - **Learnings / TILs** — things figured out, searchable later.
  - **Kudos / feedback received** — praise from others; gold for perf
    reviews (overlaps the parked `peer`/`quote` field idea).
  - **Metrics snapshots** — e.g. `p99=120ms` at a date; the before/after
    numbers that make impact stories concrete and quantified.
  - **Artifact links** — PR / commit / doc URLs tied to a brag for
    provenance (overlaps the parked `--link`/`--refs` item).
  - **Blockers / waiting-on**, **open questions / follow-ups**.
  - **Goals / OKRs** — already paper-sketched as a 2nd/3rd taggable object
    type; map brags → goals.
- **For software teams on the CLI:** standup notes, PR-review activity,
  on-call/incidents handled, mentoring/unblocking others, cross-team
  collaboration — all accomplishment-adjacent and capturable.
- **The meta-insight (ties to the passive preference):** the
  highest-leverage capture is *automatic from the dev/AI workflow*. The
  repo already ships `scripts/claude-code-post-session.sh` — passive
  capture of brags / decisions / learnings / metrics from AI coding
  sessions is the killer adjacency for an AI-CLI user, and it directly
  answers "don't make me run things."
- **Strategic question for PROJ-003 framing:** does bragfile stay
  accomplishment-focused, or become a multi-type local-first work-log
  substrate? The polymorphic schema already supports the latter cheaply —
  this is a positioning decision, not a technical blocker.
- **Research prompt parked:** `docs/research/adjacent-data-prompt.md` — a
  self-contained brief for a fresh (or `deep-research`) session to find
  what developers / CLI types / sales engineers track in random files,
  including non-project career-relational data (mentoring, kudos, network).
  Run it, then fold the keepers back into this item.

---

## Removed / delivered — keep the list honest

*When an item is pulled into a stage and ships, list it here with
the spec IDs that delivered it. Keeps the backlog history tractable
and proves items actually move out, not just in.*

(none yet — this section activates once STAGE-003 ships and
promoted items become traceable.)

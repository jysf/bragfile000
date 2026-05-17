---
project:
  id: PROJ-001
  status: shipped
  priority: high
  target_ship: 2026-05-03

repo:
  id: bragfile

created_at: 2026-04-19
shipped_at: 2026-05-17
---

# PROJ-001: MVP — capture, retrieve, export, ship

## What This Project Is

Deliver a usable, distributable MVP of Bragfile: a local-first Go CLI
(`brag`) backed by an embedded SQLite database that lets an engineer
capture a brag-worthy moment in under ten seconds and retrieve, filter,
and export those moments later. The wave of work ends when the author
has been using it daily at work for two weeks and it is installable via
`brew install bragfile`.

## Why Now

Brag-worthy moments evaporate between review cycles. The author
currently reconstructs them from commit logs under time pressure
immediately before self-reviews, and loses most of them. We want the
scaffolding (tool + habit) in place before the next review cycle so the
data exists when it is needed, rather than racing to rebuild it after
the fact. Scope is deliberately narrow — capture, retrieve, export,
distribute — so we can ship in roughly two weeks and start accumulating
real entries.

## Success Criteria

- A new entry can be captured in under 10 seconds from an open terminal
  (either `brag add "..."` one-shot or `brag add` → `$EDITOR` template).
- Past entries can be listed, filtered (tag / project / type / since),
  searched (full-text), shown, edited, and deleted from the CLI.
- `brag export` produces a readable Markdown report. (Portable-copy
  use case handled today by `cp ~/.bragfile/db.sqlite backup.db`;
  `--format sqlite` was scoped out to backlog post-SPEC-013 because
  `cp` already covers it — revisit if VACUUM INTO's defragmentation
  or cross-process-consistency wins ever matter.)
- `brew install bragfile` installs a working binary on macOS (at minimum)
  via a public homebrew tap.
- Author has logged ≥1 entry per working day for two consecutive weeks
  using the shipped binary.

## Scope

### In scope

- Core CRUD: `add`, `list`, `show`, `edit`, `delete` (STAGE-001, STAGE-002).
- Full-text search via SQLite FTS5 (STAGE-002).
- Editor-launch capture: `brag add` with no args opens `$EDITOR` against
  a templated markdown buffer; fields parsed on save (STAGE-002).
- Export: Markdown report and raw SQLite file copy (STAGE-003).
  JSON export added to STAGE-003 scope (2026-04-21) — useful for
  AI/programmatic consumers.
- Rule-based `summary --range=week|month` (group by tag/project, counts,
  rendered as markdown). No AI (STAGE-003).
- UX polish: emoji decoration on stderr feedback messages and the
  `brag show` / `brag list --pretty` output, with `NO_COLOR` +
  TTY-detection escape hatch following industry conventions
  (STAGE-003, added 2026-04-21).
- Distribution: goreleaser build, GitHub release, homebrew tap
  (STAGE-004).
- Data model designed so a future `brag ai-summary` command can read rows
  and POST to an LLM without schema changes. No AI code ships in this
  project.

### Explicitly out of scope

- LLM-backed summaries, narrative generation, resume-bullet rewriting
  — deferred to a future project (PROJ-00N — AI assist).
- Bubble Tea / interactive TUI — deferred. v0.1 is plain cobra CLI.
- Multi-device sync, cloud backup, encrypted-at-rest storage — deferred.
- Rich export targets (HTML, PDF, resume-bullet format). Markdown +
  JSON only for PROJ-001. (SQLite-file export was in scope until
  2026-04-23, then moved to backlog — `cp` already handles the
  portable-backup use case.)
- Non-macOS homebrew (Linuxbrew), apt/yum packages, Windows support.
  macOS-first; Linux is a nice-to-have that falls out of goreleaser if
  it's free, but is not a success criterion.
- Auth, accounts, sharing. This is a single-user local-first tool.

## Stage Plan

- [x] STAGE-001 (shipped on 2026-04-20) — Foundations: repo skeleton,
      Cobra scaffold, SQLite schema + migrations, `add` + `list`.
      4 specs shipped (SPEC-001/002/003/004), 1 new DEC (DEC-007),
      4 lessons landed in AGENTS.md §9/§10.
- [x] STAGE-002 (shipped on 2026-04-22) — Capture & retrieval: `add`
      shorthand flags, `list` filter flags, `show`/`edit`/`delete`,
      editor-launch for `add` and `edit`, FTS5 `search`. All 8 specs
      shipped (SPEC-005–012). 4 new DECs (DEC-007/008/009/010),
      5 new AGENTS.md §9/§12 lessons including the full premise-
      audit family. Target was 2026-05-04; shipped 12 days ahead.
- [x] STAGE-003 (shipped on 2026-04-24) — **Reports + AI-friendly
      I/O + project visibility.** 4 specs shipped
      (SPEC-013/014/015/017) + 3 DECs (DEC-011 JSON shape, DEC-012
      stdin-JSON schema, DEC-013 markdown export shape). SPEC-016
      (`--format sqlite`) deferred to backlog on 2026-04-23. Two
      AGENTS.md §9 addenda earned (substring-trap for markdown
      heading assertions, freshness-assertion for ID-vs-timestamp
      distinctness). Framed 2026-04-22/23, shipped 2026-04-24 —
      two-day wall-clock. Zero rework cycles across all four specs.
      `BRAG.md` shipped as chore pre-framing (2026-04-22).
- [x] STAGE-004 (shipped on 2026-04-25) — **Rule-based polish.**
      3 specs cherry-picked 2026-04-24 from a 9-item provisional
      pool, all shipped 2026-04-25 (one wall-clock day, fastest
      stage cadence so far): `brag summary --range week|month`
      (SPEC-018, M; emitted DEC-014 + seeded `internal/aggregate`),
      `brag review --week|--month` (SPEC-019, S; consumed DEC-014,
      added `GroupEntriesByProject`), `brag stats` (SPEC-020, S;
      consumed DEC-014, added Streak/MostCommon/Span). DEC-014
      locks the rule-based-output envelope across all three. All
      three commands emit clean markdown/JSON the user manually
      pipes into external AI — no LLM integration in PROJ-001 per
      PROJ-002 boundary. Two AGENTS.md addenda earned + codified:
      §9 audit-grep cross-check (SPEC-018 ship) and §12 negative-
      substring self-audit (SPEC-020 ship). Trim experiment
      validated at SPEC-020 (signatures + invariants sufficient
      when in-stage precedents exist as construction reference).
      Zero rework cycles, zero deferrals mid-stage, zero build-
      time DECs.
- [ ] STAGE-005 (not yet framed) — **Distribution + cleanup.**
      Five workstreams: (1) README rewrite — current README
      describes the spec-driven *development process*, not the
      `brag` tool; rewrite as user-facing, move dev-process
      content to `docs/development.md` or `CONTRIBUTING.md`; (2)
      `docs/brag-entry.schema.json` mirroring DEC-012 + BRAG.md
      reference, so AI agents producing entries have a contract
      to validate against; (3) Claude Code session-end hook
      example — `scripts/claude-code-post-session.sh` + a `/brag`
      slash command shape, demonstrates DEC-011/012 round-trip;
      (4) goreleaser cross-compile + GitHub Actions release
      workflow + homebrew tap at
      `github.com/jysf/homebrew-bragfile` + CHANGELOG discipline;
      (5) shell completions (cobra-free). Plus a blog post
      (artifact, not a spec). User explicitly wants the `brew
      install bragfile` milestone for the learning value, not
      adoption — no marketing push.

**Count:** 4 shipped / 0 active / 1 pending (STAGE-005 distribution)

## Dependencies

### Depends on

- None. This is the first project in this repo.
- External-ish: a public GitHub repo at
  `github.com/jysf/bragfile000` (exists) and a second repo for the
  homebrew tap (`github.com/jysf/homebrew-bragfile`, will be created in
  STAGE-004).

### Enables

- Future project: AI-assisted summaries and narrative generation (reads
  rows written by this project).
- Future project: TUI / Bubble Tea polish layer over the same data.
- Future project: sync or export-to-cloud if that ever becomes desired.

## STAGE-003 pre-framing notes

*Seed notes for the STAGE-003 framing session (Prompt 1c). This
section gets consumed into the stage file's Design Notes + Spec
Backlog when STAGE-003 is framed, and trimmed from the brief at
that point. Keep content here accurate; it's the handoff.*

### Post-triage scope (2026-04-22) — authoritative priority

Triaged from a full-ideas-dump via external Claude evaluation.
User filters applied: personal utility first; "nice Go app"
craftsmanship second; adoption pressure explicitly not a driver.
Non-STAGE-003 items land in `backlog.md` or are sketched below
under "STAGE-004 sketch". Detail on individual ideas follows
below this block — the scope list is the authoritative slice.

**STAGE-003 ship-blocking scope (3 items + 1 trivial pre-spec):**

Reshuffled 2026-04-22 (second pass): the "nice-to-have" bucket
moved entirely into STAGE-004 to keep STAGE-003 tight. STAGE-003
is now purely core functionality — must-haves only — with emoji,
summary, and Claude-hook work grouped into a proper polish stage.

0. **`BRAG.md` at repo root.** Trivial (S). **Shipped as a chore
   on 2026-04-22** (commit before STAGE-003 framing); counted as
   pre-spec work rather than a STAGE-003 spec. Delivers the
   agent-integration guide at a canonical location. Content lives
   in `/BRAG.md`; former `docs/agent-brag-guide.md` deleted.

1. **Project visible in `brag list` output.** User flagged
   elevated to must-have ("I want to see quickly what projects
   I have been working on"). Implementation is a design call —
   see "Project column" detail below for the four options
   (default 4-col / `--format` flag / `--pretty` bundle /
   `--columns` list). Framer picks.
2. **Report export pair:** `brag export --format markdown` and
   `--format json`. Originally a trio including `--format sqlite`;
   sqlite dropped to backlog on 2026-04-23 post-SPEC-013 because
   `cp ~/.bragfile/db.sqlite backup.db` already covers the
   portable-backup use case. JSON shape detailed below. Framed
   as SPEC-014 (JSON) + SPEC-015 (markdown).
3. **Machine-readable input/output.** `brag add --json` (stdin
   JSON → `Store.Add`) plus `brag list --format json|tsv`.
   Pairs with the export trio to close the I/O loop. Unblocks
   every downstream AI integration. Can be one M spec or split
   as 2 × S specs.

No nice-to-have bucket for STAGE-003. If the framing session
wants additional scope, take it from STAGE-004's list rather than
adding anew.

### STAGE-004 sketch — Rule-based polish (3 specs)

Cherry-picked 2026-04-24 post-STAGE-003-ship. Original 9-item
list re-triaged through the user filter "will I actually use
this?" Three specs survived; six dropped to backlog (see below).

**The three:**

- **`brag summary --range week|month`** (M) — rule-based
  aggregation grouped by project/type, rendered as markdown.
  Lighter-weight sibling to `brag export --format markdown`.
  Output is markdown that the user can copy-paste into an
  external AI session for deeper reflection — no LLM in the
  tool.

- **`brag review --week`** (S) — print entries from the last 7
  days grouped by project + three static reflection questions
  ("what pattern do you see?", "what did you underestimate?",
  "what's missing that should be here?"). Human-in-the-loop
  reflection; distinct from `summary`'s aggregation. Same AI-
  pipe-out workflow when wanted.

- **`brag stats`** (S) — entries/week, longest streak, most
  common tags/projects. Aggregations over existing schema. Run
  monthly-ish for the chart-of-yourself satisfaction.

**Explicitly dropped to `backlog.md` on 2026-04-24 cherry-pick:**

- Emoji decoration passes 1–4 — user wants emoji but doesn't
  love this specific palette/scope. Backlog with revisit
  trigger "user picks a palette and shape they actually want."
- `brag remind` — user has been logging consistently without
  one. Backlog with revisit trigger "first week with zero
  entries."
- Claude Code session-end hook — moved to STAGE-005 as a
  distribution-asset (demonstrates the AI-integration story for
  the README/blog), not STAGE-004 polish.

**Already on `backlog.md` from prior reshuffles (2026-04-22):**

- `brag export --exclude-tag <tag>` redaction filter
- Git-context auto-capture on `brag add`
- `--link` / `--refs` multi-valued field

**Out of scope for PROJ-001 entirely — PROJ-002 territory:**

- LLM-piping built into the tool (PROJ-002 "AI assist" project
  consumes STAGE-004's rule-based outputs externally).

### STAGE-005 sketch — Distribution + cleanup

Refined 2026-04-24 post-STAGE-003-ship to include three items
that surfaced during real usage of the shipped tool:

**Cleanup (was missing from earlier sketches):**

- **README rewrite.** Current README at repo root describes the
  spec-driven *development process* (frame/design/build/verify/
  ship cycles) rather than the `brag` tool itself. Rewrite as
  user-facing: install, capture, list/search, export. Move
  process-oriented content to `docs/development.md` (or
  `CONTRIBUTING.md`). Brought up by external Claude review
  2026-04-24; user agrees this is the right shape.

- **`docs/brag-entry.schema.json`.** Checked-in JSON schema
  mirroring DEC-012's stdin shape. Reference from BRAG.md as
  the contract AI agents must produce against. Pairs with the
  Claude session-end hook below to make the AI-integration
  story concrete and validatable.

**AI-integration distribution asset (moved from STAGE-004):**

- **Claude Code session-end hook example.** Ships
  `scripts/claude-code-post-session.sh` + a `/brag` slash
  command template. Pure shell + prompt; consumes the JSON
  schema above. Demonstrates the DEC-011/012 round-trip in
  practice — value as a README/blog artifact, not as
  personal-workflow polish.

**Distribution proper (persistent since project start):**

- goreleaser config (cross-compile darwin + linux, arm64 +
  x86_64).
- GitHub Actions: CI (test + lint on PR), release (tag →
  binaries → release notes).
- Homebrew tap at `github.com/jysf/homebrew-bragfile`.
- CHANGELOG discipline.
- Shell completions (`brag completion zsh|bash|fish` —
  cobra-free).

**Plus a blog post** (artifact, not a spec) — write-up of the
spec-driven development process applied to a real project.

**Explicit non-goals for STAGE-005:**

- No marketing push. User wants `brew install bragfile` for the
  learning value, not adoption.
- No pre-1.0 backward-compat promise beyond what's already
  documented in api-contract.md.
- No tags-normalization / soft-delete / edit-history work
  (external review 2026-04-24 raised these as v0.1 migration-
  cost concerns; user decision: accept the debt; revisit if/
  when they bite).

---

### Detail on individual ideas (reference material — framer reads as needed)


### Emoji decoration (accumulated 2026-04-21)

**Four-pass plan, each pass a separate candidate spec.** The
framer can decide how many to include in STAGE-003 vs. defer.

**Pass 1 — stderr feedback (S, low risk, recommended for STAGE-003).**
Prefix emoji on the four existing stderr feedback lines produced
by `brag edit` and `brag delete`:
- `Updated.` → `✏️ Updated.`
- `Deleted.` → `🗑 Deleted.`
- `No changes.` → `— No changes.`
- `Aborted.` → `✗ Aborted.`
- `Delete entry N ("title")? [y/N]` → `⚠️ Delete entry N (...)? [y/N]`

Zero risk to stdout scripting (stderr-for-humans constraint
already protects this). Existing tests assert on substrings
(`"Updated."`, `"Deleted."`) which remain intact — the emoji is
prefix-only. ~5 line code changes. No TTY detection needed for
Pass 1; stderr is always human-facing.

**Pass 2 — type-based emoji in `brag show` (S).**
`brag show <id>` produces markdown; emoji in markdown renders
fine everywhere. Decorate the `type` row of the metadata table:
- `shipped` → 🚀
- `fixed` → 🔧
- `learned` → 🎓
- `documented` → 📝
- `mentored` → 🤝
- `unblocked` → 🔓
- `reviewed` → 👀
- (unknown types) → `•` or no icon

Example:
```
| type        | 🚀 shipped                    |
```
Still valid markdown, still pipeable, still parseable. Copy-paste
into review docs preserves the emoji.

**Pass 3 — `brag list --pretty` (S or S+).**
A new `--pretty` flag that opts into emoji + formatting. Default
`brag list` stays plain for scripting. `--pretty` prefixes each
row with the type's emoji (Pass 2's mapping). Tab-separated
columns stay stable.

Example:
```
🚀  42  2026-04-21  shipped the SPEC-011 FTS5 layer
🔧  41  2026-04-20  fixed the stale .gitignore rule
🎓  40  2026-04-19  learned SQLite's external-content FTS5 pattern
```

**Pass 4 — TTY auto-detection + `NO_COLOR` (S).**
Industry conventions for colored / decorated CLI output:
- If stdout is a terminal (via
  `golang.org/x/term.IsTerminal(int(os.Stdout.Fd()))`), pretty mode
  on by default for `list` and `show`.
- If stdout is piped (redirected to file, `grep`, `cut`), plain
  mode automatically — prevents polluting pipes with emoji that
  downstream tools can't parse.
- Respect `NO_COLOR` env var (see https://no-color.org) — any
  non-empty value forces plain mode. Industry standard; also what
  accessibility-tool users set.
- Optional `BRAG_PLAIN=1` env var as brag-specific override.

Implementation: ~10 lines in a small helper; used by `list`,
`show`, and any future decorated output.

### Emoji caveats worth naming in the spec(s)

- **Cross-platform rendering.** Modern terminals (GNOME Terminal,
  Konsole, Alacritty, Kitty, WezTerm, iTerm2, Windows Terminal,
  st) + any modern Linux/macOS distro with Noto Color Emoji
  handle these fine. Bare Linux TTY, legacy cmd.exe on Windows 10,
  and some enterprise-locked-down terminals don't. `NO_COLOR`
  + `--plain` is the escape hatch for those 2–5% of users.
- **Column width.** Emoji are typically 2 cells wide but some
  terminals compute this as 1, causing subtle misalignment in
  markdown tables. Cosmetic only.
- **Variation selectors + ZWJ sequences.** Avoid complex emoji
  like 👨‍💻 (combined codepoints); prefer single-codepoint
  emoji (🚀, 🔧, 📝) which render consistently.
- **Screen readers.** `"🚀 shipped"` reads as "rocket shipped."
  Slightly noisy; screen-reader users typically set `NO_COLOR=1`
  in their shell rc already.

### JSON export (accumulated 2026-04-21)

**Sibling of markdown / sqlite exports. Primary motivation: AI/
programmatic consumers.** Example flow: a user's downstream AI
agent reads `brag export --format json --since 90d` to produce a
quarterly summary or resume bullet. Also useful for piping into
`jq`, backup tooling, or cross-tool integration.

Shape: an array of entry objects, one per entry. Field names
match the SQL column names (`id`, `title`, `description`, `tags`,
`project`, `type`, `impact`, `created_at`, `updated_at`). RFC3339
timestamps as strings (matches storage layer). `tags` stays as a
comma-joined string (matches DEC-004 — don't normalize to array
here unless the design session has a reason).

Example output:
```json
[
  {
    "id": 11,
    "title": "Shipped FTS5 full-text search index layer...",
    "description": "Shipped SPEC-011 — new 0002_add_fts.sql migration...",
    "tags": "sqlite,fts5,migrations,search",
    "project": "bragfile",
    "type": "shipped",
    "impact": "Prepared the data layer for brag search...",
    "created_at": "2026-04-22T06:30:00Z",
    "updated_at": "2026-04-22T06:30:00Z"
  },
  ...
]
```

Same filter flags as `brag list` (so
`brag export --format json --project bragfile --since 7d` works).
Output to stdout by default; `--out file.json` to write to a file.

Use stdlib `encoding/json` — no new dep needed. Pretty-printed by
default (indent=2) for human readability; `--compact` flag to
disable indentation for pipe consumers if anyone asks.

### `brag list` display options — show project column (accumulated 2026-04-22)

**Motivation.** Today `brag list` prints 3 columns tab-separated:
`<id>\t<created_at>\t<title>`. The `project` field is populated
per-entry but invisible at scan time — users can't see "what
projects have I been working on lately?" without `brag show <id>`
or filtering by a specific `--project`. Adding project to the
list output makes daily scanning more useful.

**Design options for the framer to pick between:**

1. **Part of `--pretty` mode** (leans composable with emoji pass 3).
   `brag list` stays 3-column plain for backwards-compat with
   scripts. `brag list --pretty` adds emoji + project column
   (e.g., `🚀  12  2026-04-22  [bragfile]  shipped FTS5 search`).
   One new flag, one coherent pretty-mode feature bundle.
2. **Dedicated `--show-project` / `-P` flag.** Opts in just the
   project column without emoji. Clean separation from the emoji
   concern; combines freely with `--pretty`.
3. **`--columns <list>` flag.** User picks which columns to show,
   e.g., `--columns id,created_at,project,title`. Most flexible;
   most complex. Probably overkill for MVP.
4. **Change default** to always include project. Breaks existing
   scripts that parse 3 tab-separated columns. Rejected — breaking
   stable contract for a UX nicety isn't worth it.

**Author's lean: option 1 (bundle into `--pretty`)** — simplest
composition with the emoji work, keeps plain mode byte-stable.
Framer can decide otherwise if dogfooding reveals that project-
in-plain-output is a strong standalone ask.

**Visibility of empty project.** When an entry has no `project`
set, the column should render as `-` or `(none)` rather than an
empty field, so the tab-separated shape stays consistent.

### Triage history — ideas-dump resolved 2026-04-22

Full ideas-dump (external Claude evaluation) was triaged on
2026-04-22 into the Post-triage scope block above, the
STAGE-004 sketch above, and `backlog.md` (new file). Nothing
pending. Items in `backlog.md` have revisit triggers and can be
promoted into future stages or projects when the trigger fires.

---

## Project-Level Reflection

*Drafted via Prompt 1e (Project Close — FIRST USE in this repo) in a
fresh session on 2026-05-17.*

### Success Criteria check

The brief enumerated five success bullets at project framing. Four
shipped outright; one is satisfied in spirit but cannot be literally
verified yet because the homebrew-shipped binary has only existed for
6 days.

- **A new entry can be captured in under 10 seconds from an open
  terminal.** ✅ Shipped via SPEC-003 (`brag add --title "..."`
  one-shot, STAGE-001) and SPEC-010 (`brag add` no-args → `$EDITOR`
  on templated markdown buffer, STAGE-002). End-to-end verified on
  installed binary 2026-05-11.
- **Past entries can be listed, filtered, searched, shown, edited,
  and deleted from the CLI.** ✅ Full surface delivered:
  SPEC-004/SPEC-007 (list + filter flags), SPEC-006 (show),
  SPEC-009 (edit, with `$EDITOR` round-trip), SPEC-008 (delete with
  `[y/N]` confirmation), SPEC-011 + SPEC-012 (FTS5 full-text
  search), SPEC-013 (project column in list).
- **`brag export` produces a readable Markdown report.** ✅ Shipped
  via SPEC-014 (JSON shape + DEC-011) + SPEC-015 (markdown shape +
  DEC-013), plus SPEC-017 (`brag add --json` stdin half + DEC-012)
  closing the I/O loop. **Scope-adjusted (deliberate):**
  `--format sqlite` deferred to backlog 2026-04-23 because
  `cp ~/.bragfile/db.sqlite backup.db` already covers the
  portable-backup use case named in the brief; SPEC-016 slot
  preserved in stage file as deferral marker.
- **`brew install bragfile` installs a working binary on macOS via
  a public homebrew tap.** ✅ Shipped via SPEC-023 + v0.1.0 cut on
  2026-05-11; cask at `homebrew-bragfile/Casks/bragfile.rb`;
  `brew install jysf/bragfile/bragfile` works on macOS arm64.
  **Caveat:** unsigned binary triggers macOS Gatekeeper on first
  run; xattr workaround documented in README §Install + AGENTS.md
  §4; full notarization deferred to backlog with an implementer
  checklist at `docs/macos-notarization-checklist.md`.
- **Author has logged ≥1 entry per working day for two consecutive
  weeks using the shipped binary.** ⚠ **Satisfied in spirit; not
  yet literally verifiable.** The dogfooding habit ran continuously
  through PROJ-001 against locally-built binaries (`just install` →
  `~/go/bin/brag`); `framework-feedback/process-feedback.md`
  attests to ~20 entries across two weeks as of 2026-04-20. The
  homebrew-shipped binary only exists since 2026-05-11 (6 working
  days at project close on 2026-05-17), so the criterion's literal
  "two consecutive weeks using the *shipped* binary" cannot yet be
  satisfied. The habit and corpus are real; the binary-source
  switch is recent. Recommend treating this as "shipped, with a
  trailing-clock asterisk" rather than as deferred.

### Three-sentence summary

The brief's framing held: five stages shipped in framing order with
no reordering, no stage cancelled mid-flight, and no DEC deprecated
across the 23 specs and 14 DECs that landed; the only scope
adjustments were small and explicitly logged (JSON export added to
STAGE-003, `--format sqlite` deferred to backlog, three STAGE-004
candidates retriaged out at framing time). STAGES 001–004 ran fast
and smooth (002 shipped 12 days ahead of its target, 003 in 2
wall-clock days, 004 in 1 day); STAGE-005 stretched to +6 days past
the framing-time outside boundary because of a user-requested
pre-distribution security review pause and a Phase 2 manual ship
cycle that surfaced two operational gotchas (goreleaser
dual-tag-on-same-commit; macOS Gatekeeper on unsigned binaries),
both codified into AGENTS.md §4 rather than carried as silent debt.
The deeper signal across the project is that the framework itself
matured measurably mid-stream — ~150 lines of earned codifications
landed in AGENTS.md §9/§10/§12/§4, and PROJ-001 produced the first
"two opposing-outcome cases earn codification at N=2" data point
(§12(b) design-time pre-flight at SPEC-024 ship), which is itself
the strongest meta-finding of the project.

### Reflection questions

- **Did we deliver the outcome in "What This Project Is"?** Yes. A
  local-first Go CLI exists on the public homebrew tap; an engineer
  can capture, retrieve, filter, search, export, summarise, review,
  and stat their brags from a single terminal; entries live at
  `~/.bragfile/db.sqlite`; an AI agent has a checked-in JSON-Schema
  contract and a reference Claude Code session-end hook to integrate
  against without ever needing source. The two operationally-visible
  caveats — macOS Gatekeeper xattr workaround and the 6-day clock
  on the shipped-binary dogfooding criterion — are bounded and
  documented rather than carried as silent debt.

- **How many stages did it actually take?** 5 stages (exactly as
  framed): STAGE-001 foundations → 002 capture-and-retrieval → 003
  reports-and-AI-friendly-I/O → 004 rule-based polish → 005
  distribution-and-cleanup. 23 specs shipped (4/8/4/3/4) plus 1
  deferred slot (SPEC-016 → backlog). 14 DECs emitted, zero
  deprecated, zero superseded. Project-vs-brief timeline: created
  2026-04-19, target ship 2026-05-03, actual v0.1.0 cut 2026-05-11
  (+8 days), project close 2026-05-17 (+14 days). Most of the
  overrun lives in STAGE-005; STAGES 001–004 were on or ahead of
  pace.

- **What changed between starting and shipping?** Three categories
  of change, all small and all explicitly logged at the moment they
  happened.
  - *Additive scope:* JSON export added to STAGE-003 scope
    (2026-04-21) — useful for AI/programmatic consumers; emoji + UX
    polish list accumulated 2026-04-21 then triaged out of STAGE-003
    into a polish bucket; project column in `brag list` added to
    STAGE-003 must-haves (2026-04-22) for daily-scan utility;
    pre-distribution security review *inserted* into STAGE-005
    mid-ship at user request (2026-04-26).
  - *Subtractive scope:* `--format sqlite` export deferred to
    backlog 2026-04-23 post-SPEC-013 because `cp` already covers
    the portable-backup use case; emoji/remind/Claude-hook
    re-triaged out of STAGE-004 to backlog or STAGE-005 on
    2026-04-24 ("will I actually use this?" filter).
  - *Framework maturation:* the most consequential change wasn't to
    scope but to the framework itself — AGENTS.md grew the full
    premise-audit family (inversion/count/status), audit-grep
    cross-check, BSD-grep `--exclude-dir` warning, NOT-contains
    self-audit, literal-artifact-as-spec, §10 push-discipline, §4
    dual-tag and Gatekeeper recipes, and §12(b) design-time
    pre-flight. None of these existed at project framing; all
    earned through actual punch lists or post-incident reflections.

- **Lessons that should update AGENTS.md, templates, or
  constraints?** Almost everything codify-worthy landed at the spec
  ship that surfaced it, not at this project close. Confirming
  STAGE-005's no-codifications-at-stage-close stance at project
  altitude: the lessons live where they were earned.
  - **Already landed mid-project (no further action):**
    - §9 audit family — separate buf/errBuf cross-leakage check
      (SPEC-001); monotonic tie-break (SPEC-002); distinctive-token
      help asserts (SPEC-005); markdown heading line-equality
      (SPEC-015); ID-vs-timestamp freshness (SPEC-017); BSD-grep
      `--exclude-dir` warning (SPEC-021).
    - §9 premise-audit family — inversion-removal (SPEC-010),
      count-bump (SPEC-011), status-change (SPEC-012); plus
      audit-grep cross-check both sides (SPEC-018).
    - §10 anchored-`.gitignore` for binaries (SPEC-003);
      push-discipline between local commits and merge (codified
      STAGE-005 framing 2026-04-25 after three confirming cases
      SPEC-013 + SPEC-018 + SPEC-019).
    - §12 prose-cannot-relax-blocking-constraint (SPEC-007);
      decide-at-design-time when decidable (SPEC-018 after three
      confirming cases); NOT-contains self-audit against
      load-bearing prose (SPEC-019/020); literal-artifact-as-spec
      (SPEC-021, three confirming cases 018/020/021);
      design-time pre-flight (SPEC-024 §12(b), two
      opposing-outcomes cases SPEC-023 D3 NEGATIVE + SPEC-024
      POSITIVE).
    - §4 dual-tag-on-same-commit recovery + macOS Gatekeeper
      unsigned-binary note (both codified 2026-05-11 in `a15e36b`
      mid-STAGE-005 ship).
  - **Proposed at project close — one inline meta-rule (see
    "Proposed AGENTS.md edits at project close" below).** The
    §12(b) codification at N=2 was distinctive enough as a process
    finding to warrant a meta-rule at project altitude.
  - **NOT promoted at project close (carry to PROJ-002 framing):**
    three WATCH-list items — §12 sub-rule (a) literal-test
    assertion pre-flight; trust-but-verify agent push reports; §13
    fresh-session working-tree state preservation. Each has a
    documented third-case trigger and is mechanically adjacent to
    a rule already at AGENTS.md scope; codifying at N=2-within-
    one-spec would over-fit. Carry to PROJ-002 framing per
    STAGE-005's already-explicit ask.
  - **Premise-audit sub-template extraction.** Flagged at SPEC-015
    ship, re-flagged at STAGE-004 ship, deferred at STAGE-005
    framing on grounds STAGE-005 was mostly new-file work where
    the audit-family rules have minimal trigger surface, deferred
    at STAGE-005 close on grounds that SPEC-021 was the only
    audit-heavy spec in the stage and its execution was clean.
    Carry forward to PROJ-002 framing — that is the next project
    where same-shape feature work will repeat the skeleton and
    the extraction's value concentrates.

- **What did we defer to the next project?** Two distinct
  categories: feature scope and framework process. Feature scope
  is enumerated in `backlog.md` (~20 entries with concrete
  trigger conditions) — top-of-mind candidates for PROJ-002
  framing intake are LLM-backed summaries (always was PROJ-002's
  raison d'être), `--at <date>` backdating, the emoji passes (if
  a palette is chosen), attachments, and the `--exclude-tag`
  redaction filter. Framework process carry-forward: three WATCH-
  list AGENTS.md candidates needing third confirming cases; the
  premise-audit sub-template extraction; the STAGE-004
  trim-when-structural-analogy heuristic still at soft N=1.
  Accepted v0.1 debt deliberately not addressed: tags-
  normalization migration; soft-delete + edit-history; macOS
  notarization (with full implementer checklist at
  `docs/macos-notarization-checklist.md` waiting for a trigger).
  See "Carry-forward to PROJ-002" below for the synthesised
  themes.

### Carry-forward to PROJ-002

- **Integration target is now a distributable binary, not a source
  build.** PROJ-002's AI-assist surface designs against
  `brew install jysf/bragfile/bragfile` as the install path; the
  JSON Schema at `docs/brag-entry.schema.json` is the contract AI
  agents validate against; `scripts/claude-code-post-session.sh`
  is the reference implementation new integrations pattern off.
  This is a substantive architectural inheritance, not a
  formality.
- **All 14 DECs apply forward unchanged.** Framing PROJ-002 should
  not require re-litigating DEC-001..014; if any DEC needs
  supersession the framing prompt will surface that explicitly. No
  pre-emptive supersession review needed at project close.
- **Three WATCH-list framework rules with documented third-case
  triggers:** §12 sub-rule (a) (literal-test assertion pre-flight),
  trust-but-verify agent push reports, §13 fresh-session
  working-tree state preservation. PROJ-002's feature work in the
  same shape as STAGE-002/003/004 is the natural next test bed.
- **Premise-audit sub-template extraction** — third deferral. The
  pattern earned three confirming applications in PROJ-001
  (SPEC-010/011/012); a template extraction would compress the
  skeleton across same-shape feature work. PROJ-002 framing should
  decide at framing-time whether to extract before the first
  feature spec or hold one more time.
- **Backlog entries with PROJ-002-shaped triggers:** LLM-backed
  summaries (explicitly PROJ-002 territory); attachments (heavy
  scope, AI agents might need this); `--link`/`--refs` field
  (AI consumers benefit from structured access); JSON envelope
  + lenient mode + NDJSON batch (all pair with AI-piping
  workflows). Backlog entries with later-project triggers stay
  put.

### Proposed AGENTS.md edits at project close

**One inline meta-rule proposed for codification at project
altitude.**

> **Process lessons that produce paired opposing outcomes (a
> costly miss + a cheap save) on the same mechanical sub-rule
> earn codification at N=2; same-outcome confirming cases still
> need N=3.** Earned across SPEC-023 D3 NEGATIVE (design skipped
> pre-flight against `goreleaser check`; deprecated
> `brews:`→`homebrew_casks:` keys surfaced at verify, cost a
> recovery commit) and SPEC-024 §12(b) POSITIVE (design ran a
> scratch Go program against cobra v1.10.2's `GenBashCompletion`,
> caught a bash-marker assumption mid-design, zero deviations at
> build). The two cases form a single mechanical proof — same
> sub-rule, opposing outcomes — and that pairing carries more
> evidence than three same-outcome confirming cases. Apply the
> meta-rule at any future codification decision: paired
> opposing-outcome data points clear the bar at N=2; otherwise
> hold for N=3.

This belongs in AGENTS.md §12 (the cycle-specific rules section
where the codification machinery already lives), likely as a
brief preamble or sidebar to the existing per-rule lessons. The
coordinator commits the exact insertion point.

**Otherwise, no new inline codifications at project close.** Every
codify-worthy lesson earned during PROJ-001 was codified at the
spec ship that surfaced it; three WATCH-list items stay watched;
the STAGE-004 trim heuristic stays soft N=1; the premise-audit
sub-template extraction stays deferred. STAGE-005's close confirmed
this stance at stage altitude, and re-examination at project
altitude reconfirms it.

### Blog-post artifact disposition

**Recommendation: (b) defer post-close as a project-close-adjacent
follow-up.** The brief and STAGE-005 framing both named
`docs/blog-spec-driven-bragfile.md` as a deliverable; it has NOT
been drafted yet.

Drafting now as part of project close would conflate two scoping
acts (project-altitude reflection + audience-shaped write-up) and
risk bloating the close commit. Deferring post-close keeps the
project-close commit clean and gives the blog draft its own session
with proper scope (audience, voice, length, publication target —
none of which were settled at framing).

Raw material already exists in three places: `framework-feedback/`
(458 lines of agent-perspective notes on the framework itself), the
five stage-level reflection sections (cumulative dense narrative),
and the session log (chronological). A dedicated drafting session
with these as inputs is a clean handoff.

Recommend the coordinator opens a separate session post-close
specifically to draft the blog post; publication target
(HN / dev.to / Substack / personal site / GitHub Discussion)
remains TBD per the in-repo-canonical convention named at STAGE-005
framing.

### Untracked-files disposition

Three persistent untracked files have ridden the working tree across
PROJ-001. Per-file recommendations:

- **`framework-feedback/` (`process-feedback.md` + 
  `scale-recommendations.md`, 458 lines total): COMMIT to 
  `docs/framework-feedback/`.** Both files explicitly open with
  "Intended for the template author; not part of the bragfile
  project itself" — they are durable artifacts about the
  spec-driven framework, written from the agent's perspective.
  They are exactly the kind of evidence that should be reachable
  from `git log` when the template author later refines the
  framework. Committing under `docs/framework-feedback/` keeps
  them adjacent to the rest of the docs without claiming they're
  project documentation in the user-facing sense.
- **`revew1.md` (155 lines): DELETE.** The file is a one-shot
  weekly-review prompt template that hardcodes file paths for a
  specific point-in-time spec list (top 9 specs of STAGE-003
  era). Its content is derivable from `FIRST_SESSION_PROMPTS.md`
  Prompt 6 plus `just status`. The session log already flagged
  this category as a cleanup candidate ("Untracked files worth
  cleaning up eventually: framework-feedback/ … and
  status-after-nine-specs.md"). The weekly-review machinery as a
  whole is a framework concern, not a per-project artifact; if
  the user wants to invoke Prompt 6 again, regenerating from
  `FIRST_SESSION_PROMPTS.md` is the right path.
- **`status-after-nine-specs.md` (39 lines): DELETE.** Confirmed
  stale — a `just status` snapshot from the SPEC-009 era (project
  is now at SPEC-024 ship + STAGE-005 close + v0.1.0 cut).
  Explicitly flagged in the session log as a stale snapshot
  obsoleted by the session log itself.

The coordinator runs these as three small commits (or bundled into
the project-close commit) per their judgement.

### Recommendation: mark PROJ-001 shipped and frame PROJ-002

**Ship-and-frame.** "Feature-complete bragfile MVP on homebrew" is
the right scope for PROJ-001 and the brief's success criteria are
substantively met (with the trailing-clock asterisk on the
dogfooding criterion explicitly named above). Extending PROJ-001
into a "phase 2" would violate the framework's hierarchy rule that
*specs do not cross project boundaries* (AGENTS.md §2) — and
PROJ-002's AI-assist scope brings genuinely different architectural
concerns (API-key strategy, network boundary, prompt design,
model-choice strategy, latency/cost tradeoffs, offline-first story)
that warrant their own framing session and their own initial DEC
set.

Concretely: STAGE-005's close already explicitly enumerated PROJ-002
framing as the next step; the brief's "Enables" section names
PROJ-002 by reference; the backlog has been triaged with PROJ-002
markers since 2026-04-22; and the homebrew binary that PROJ-002's
AI agents will integrate against now exists. Frame PROJ-002 in a
separate Prompt 1a cycle.

### Project summary

PROJ-001 delivered the bragfile MVP in five stages across roughly
four weeks of wall-clock (2026-04-19 brief → 2026-05-11 v0.1.0 cut
→ 2026-05-17 project close), shipping 23 specs and 14 DECs with
zero DEC deprecations and zero stage reorderings. The brief's
framing held intact; scope adjustments were small, additive or
clean deferrals, and explicitly logged at the moment they happened.
The most consequential thing that didn't fit cleanly into any
stage's reflection: the framework itself matured measurably mid-
stream, with ~150 lines of earned AGENTS.md codifications, the
first appearance of the §12 literal-artifact-as-spec pattern in
three confirming cross-format applications, and the first
"opposing-outcomes-at-N=2" codification (§12(b)) which generalises
into the project-altitude meta-rule proposed above. The two
operational caveats on the shipped binary (macOS Gatekeeper xattr,
6-day dogfooding clock) are bounded and documented; three WATCH-list
framework patterns and one sub-template extraction carry forward to
PROJ-002 framing rather than getting force-codified at N=2 same-
outcome. PROJ-001 closes; PROJ-002 frames next.

---
project:
  id: PROJ-001
  status: active
  priority: high
  target_ship: 2026-05-03

repo:
  id: bragfile

created_at: 2026-04-19
shipped_at: null
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
- `brag export` produces either (a) a readable Markdown report or
  (b) a portable copy of the SQLite file suitable for backup or another
  machine.
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
- Rule-based `summary --range=week|month` (group by tag/project, counts,
  rendered as markdown). No AI (STAGE-003).
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
- Rich export targets (JSON, HTML, PDF, resume-bullet format). Markdown
  + sqlite-file only.
- Non-macOS homebrew (Linuxbrew), apt/yum packages, Windows support.
  macOS-first; Linux is a nice-to-have that falls out of goreleaser if
  it's free, but is not a success criterion.
- Auth, accounts, sharing. This is a single-user local-first tool.

## Stage Plan

- [ ] STAGE-001 (not yet framed) — Foundations: repo skeleton, Cobra
      scaffold, SQLite schema + migrations, `add` + `list`.
- [ ] STAGE-002 (not yet framed) — Capture & retrieval: `show`, `edit`
      (editor-launch templated), `delete`, FTS5 `search`.
- [ ] STAGE-003 (not yet framed) — Export & summary: markdown export,
      sqlite-file export, rule-based `summary`.
- [ ] STAGE-004 (not yet framed) — Distribution: goreleaser, homebrew
      tap, README, release notes.

**Count:** 0 shipped / 0 active / 4 pending

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

## Project-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Project Is"?** <not yet>
- **How many stages did it actually take?** <not yet>
- **What changed between starting and shipping?** <not yet>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <not yet>
- **What did we defer to the next project?**
  - <not yet>

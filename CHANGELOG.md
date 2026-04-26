# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - YYYY-MM-DD

Initial public release of `brag`, a local-first Go CLI for capturing
and retrieving career-worthy moments. Entries live in an embedded
SQLite database at `~/.bragfile/db.sqlite`. No cloud, no sync, no
account.

### Added

- `brag add` — capture an entry via flags (`-t/--title`, `-d`, `-T`,
  `-p`, `-k`, `-i`) or via `$EDITOR` against a templated markdown
  buffer.
- `brag add --json` — programmatic capture from stdin, validated
  against the DEC-012 single-object schema (title required;
  optional user-owned fields; server-owned fields tolerated and
  ignored; unknown keys strict-rejected).
- `brag list` — list entries newest-first, with `--project`,
  `--tag`, `--type`, `--since` filters and `--show-project / -P`
  for an extra project column. `--format json|tsv` for
  machine-readable output.
- `brag show` — display a single entry by ID with full metadata.
- `brag edit <id>` — round-trip an entry through `$EDITOR`.
- `brag delete` — delete an entry by ID with `[y/N]` confirmation.
- `brag search <query>` — SQLite FTS5 full-text search across
  title, description, tags, project, and impact.
- `brag export` — bulk export with `--format markdown|json` and the
  same filter flags as `list`. `--out file` to write to disk.
- `brag summary --range week|month` — rule-based aggregation
  grouped by project and type, rendered as markdown or JSON
  (DEC-014 envelope).
- `brag review --week|--month` — entries grouped by project plus
  three reflection questions, designed to be piped into an
  external AI session.
- `brag stats` — six lifetime metrics: total entries, weekly
  rolling average, current streak, longest streak, top-5 tags,
  top-5 projects, corpus span.
- `docs/brag-entry.schema.json` — JSON Schema (draft 2020-12)
  mirroring the `brag add --json` stdin contract for AI-agent
  validation.
- `scripts/claude-code-post-session.sh` + `examples/brag-slash-command.md`
  — reference Claude Code session-end hook and slash-command
  template demonstrating the round-trip.
- macOS (arm64, amd64) and Linux (arm64, amd64) binaries via
  goreleaser.
- Homebrew tap at `github.com/jysf/homebrew-bragfile` —
  `brew install jysf/bragfile/bragfile`.

### Decisions of record

The following architectural decisions are committed in this release.
Each decision file under `/decisions/` carries the full rationale.

- DEC-001 — pure-Go SQLite driver (`modernc.org/sqlite`); no CGO.
- DEC-002 — embedded migrations via `embed.FS`, applied on
  `storage.Open`.
- DEC-003 — config resolution order: `--db` flag → `BRAGFILE_DB`
  env → `~/.bragfile/db.sqlite` default.
- DEC-004 — tags stored as a comma-joined string for MVP.
- DEC-005 — integer `AUTOINCREMENT` primary keys.
- DEC-006 — `spf13/cobra` as the CLI framework.
- DEC-007 — required-flag validation in `RunE` (cobra's
  `MarkFlagRequired` reports errors via stderr + non-zero exit;
  the project owns user-error rendering uniformly).
- DEC-008 — `--since` accepts date (`2026-04-19`) or duration
  (`7d`, `30d`).
- DEC-009 — editor buffer format for `brag add` / `brag edit`
  (markdown front-matter on top, free body below).
- DEC-010 — `brag search` query syntax (auto-tokenize whitespace;
  treat hyphens / dots as literal; phrase-quote multi-word
  fragments).
- DEC-011 — JSON output shape for `brag list --format json` and
  `brag export --format json`: naked array of nine-key entry
  objects; field names match SQL columns verbatim.
- DEC-012 — `brag add --json` stdin schema: single object, title
  required, server-owned fields tolerated-and-ignored, unknown
  keys strict-rejected.
- DEC-013 — markdown export shape for `brag export --format
  markdown` (+ `--flat`): per-entry markdown blocks under
  per-project headings; `--flat` flattens.
- DEC-014 — rule-based output envelope for `brag summary` /
  `brag review` / `brag stats`: single-object JSON envelope with
  `generated_at` / `scope` / `filters` provenance + per-spec
  payload keys; markdown convention reuses DEC-013's provenance
  + summary-block style.

[Unreleased]: https://github.com/jysf/bragfile000/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/jysf/bragfile000/releases/tag/v0.1.0

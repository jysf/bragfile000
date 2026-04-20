# Bragfile

A local-first CLI for engineers to capture and retrieve career
accomplishments ("brags") for retros, reviews, and resumes. Go + embedded
SQLite; installs via `brew install bragfile`; binary is `brag`.

> **Status:** in development. See `projects/PROJ-001-mvp/brief.md` for
> what's being built in this wave of work.

This repo uses a spec-driven workflow where Claude plays every role (architect, implementer, reviewer) across different sessions.

## Hierarchy

```
Repo (this app)
 └─ Project (a wave of work: "MVP", "v2 improvements")
     └─ Stage (a coherent chunk within a project)
         └─ Spec (an individual task)
              └─ Cycle (Frame → Design → Build → Verify → Ship)
```

## Getting started

**First time?** Read `GETTING_STARTED.md` — it walks you through your first project end-to-end.

**Daily work?** Run `just --list` to see available commands.

**Common commands:**
```bash
just status                        # See active project, stage, specs by cycle
just new-spec "title" STAGE-001    # Scaffold a new spec
just advance-cycle SPEC-001 verify # Update a spec's cycle
just archive-spec SPEC-001         # Move a shipped spec to done/
just weekly-review                 # Print the weekly review prompt
```

## Key discipline in this variant

Because Claude plays every role, context contamination is the biggest risk. Four habits keep it at bay:

1. **New Claude session per cycle** (especially design → build and build → verify)
2. **The spec file is the source of truth** between sessions — no "as I said earlier"
3. **Weekly review is non-optional** (`just weekly-review`)
4. **Honest confidence values** on decisions

See `AGENTS.md` section 13 for the full discipline.

## The app itself

`brag` is a terminal CLI that stores brag-worthy work moments in a local
SQLite database at `~/.bragfile/db.sqlite`. Core operations are `add`,
`list`, `search`, `show`, `edit`, `delete`, `export`, and `summary`.
`brag add` with no arguments opens `$EDITOR` against a templated
markdown buffer; fields are parsed on save.

Build, test, and run commands will live in `AGENTS.md` Section 4 once
the repo/project design cycle (Prompt 2a) has filled it in.

## Where things live

| Path | Purpose |
|---|---|
| `AGENTS.md` | Conventions for Claude working in this repo |
| `.repo-context.yaml` | Structured metadata about the app |
| `docs/` | Architecture, data model, API contract |
| `guidance/` | Repo-level rules and open questions |
| `decisions/` | Decision log (accumulates across projects) |
| `projects/` | Each project (wave of work) lives here |
| `projects/*/brief.md` | What this project is and why |
| `projects/*/stages/` | Stages within a project |
| `projects/*/specs/` | Specs within a project (with folded-in Implementation Context) |
| `cmd/brag/` | CLI entrypoint (added during STAGE-001) |
| `internal/` | Implementation packages: storage, commands, editor, export |

## License

MIT. See `LICENSE`.

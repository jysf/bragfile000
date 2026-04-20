# My App

*This file replaces the template README when you run `just init` — rename or customize as needed.*

This repo uses a spec-driven multi-agent workflow. Claude handles architecture and review; a separate implementer agent (Kilo Code, Factory Droids, AdaL, Cursor, etc.) handles implementation.

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

## The app itself

[REPLACE: describe what this repo actually builds. The workflow above
is the *meta-process*; this section is about the *app*. Include:]

- What the app does (1 paragraph)
- How to run it locally (link to AGENTS.md Section 4)
- How to run tests

## Where things live

| Path | Purpose |
|---|---|
| `AGENTS.md` | Conventions for agents working in this repo |
| `.repo-context.yaml` | Structured metadata about the app |
| `docs/` | Architecture, data model, API contract |
| `guidance/` | Repo-level rules and open questions |
| `decisions/` | Decision log (accumulates across projects) |
| `projects/` | Each project (wave of work) lives here |
| `projects/*/brief.md` | What this project is and why |
| `projects/*/stages/` | Stages within a project |
| `projects/*/specs/` | Specs within a project |
| `projects/*/handoffs/` | Architect → implementer delegation records |
| `src/` | [REPLACE: the actual app code] |

## License

[REPLACE]

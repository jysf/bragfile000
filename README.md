# Spec-Driven Multi-Agent Repo Template

A GitHub template for running spec-driven development on an app (or apps) where the **repo is the app** and **projects are waves of work** against that app. Works whether you use Claude alone or Claude plus a dedicated implementer agent (Kilo Code, Factory Droids, AdaL, etc.).

## Hierarchy

```
Repo (the app — persists forever)
 └─ Project (a wave of work: "MVP", "v2 improvements", "redesign")
     └─ Stage (a coherent chunk within a project)
         └─ Spec (an individual task)
              └─ Cycle (Frame → Design → Build → Verify → Ship)
```

- **Repo** accumulates architecture, conventions, constraints, and decisions across all projects.
- **Projects** are bounded waves of work. You may have one active project, several sequential projects over time, or (rarely) multiple active projects.
- **Stages** are epic-sized chunks within a project. A project typically has 2–5 stages.
- **Specs** are individual implementable tasks. Each belongs to exactly one stage.
- **Cycles** are the 5-phase lifecycle a spec goes through.

## Using this template

**Option A — GitHub template (recommended):**

Click **"Use this template"** at the top of this repo on GitHub. Create your new repo. Clone it. Then inside your new repo:

```bash
just init
```

This asks whether you want the `claude-only` or `claude-plus-agents` variant, then moves the right files to the repo root and removes what you don't need.

**Option B — Clone and delete git:**

```bash
git clone https://github.com/YOUR-USERNAME/THIS-TEMPLATE.git my-new-repo
cd my-new-repo
rm -rf .git && git init
just init
```

## After `just init`

You'll have a repo root containing:
- `AGENTS.md` — conventions for all agents working here
- `CLAUDE.md` — pointer to `AGENTS.md` for Claude Code
- `GETTING_STARTED.md` — walkthrough for your first project
- `FIRST_SESSION_PROMPTS.md` — copy-paste prompts for each phase
- `.repo-context.yaml` — describes the app (the repo)
- `justfile` — commands you'll run daily
- `docs/`, `guidance/`, `decisions/` — repo-level (accumulate across all projects)
- `projects/PROJ-001-example-mvp/` — example project you can learn from or delete

**Next step:** open `GETTING_STARTED.md` and follow it.

## The two variants at a glance

### `claude-plus-agents/` — Claude architects, a separate agent implements

For workflows where:
- Claude writes specs + reviews PRs (architect + reviewer)
- A different tool (Kilo Code, Factory Droids, AdaL, Cursor, etc.) implements

Adds `/projects/*/handoffs/` — explicit handoff documents that carry context between agents that don't share memory.

### `claude-only/` — Claude does everything

For workflows where Claude plays every role — architect, implementer, reviewer. No separate implementer tool.

No `/handoffs/` folder. The context the implementer needs is folded into each spec's `## Implementation Context` section.

**Not sure which?** Start with `claude-only`. Migration to `claude-plus-agents` later is about an hour of mechanical work.

## `just` commands available

Run `just --list` to see everything. The main ones:

| Command | What it does |
|---|---|
| `just init` | One-time: choose variant, scaffold the repo |
| `just status` | Current state: active project, stage, specs by cycle, stale items |
| `just new-spec "title" STAGE-NNN` | Scaffold a new spec with next available ID |
| `just new-stage "title" PROJ-NNN` | Scaffold a new stage in the active (or named) project |
| `just advance-cycle SPEC-NNN verify` | Update a spec's `task.cycle` field |
| `just archive-spec SPEC-NNN` | Move a shipped spec to `done/` + update stage backlog |
| `just weekly-review` | Load recent activity and print the Weekly Review prompt |

## What ContextCore concepts this template uses

This template is philosophically aligned with [ContextCore](https://github.com/neil-the-nowledgeable/contextcore) — the same vocabulary (`task.*`, `insight.*`, `guidance.*`, `handoff.*`, `project.*`), the same artifact model, the same forward-compatibility to OTel-based observability. But it requires no infrastructure — everything is markdown files until (and only if) you graduate to the full ContextCore stack.

See `docs/CONTEXTCORE_ALIGNMENT.md` for details (created by `just init`).

## License

Do whatever you want with this template.

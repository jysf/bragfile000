# Development — the spec-driven workflow

How this project is built. Read this if you're contributing,
modifying the workflow, or curious why the repo has the structure
it does.

## TL;DR

Each piece of work is a **spec** — a single implementable task —
that moves through five **cycles**:

```
Frame → Design → Build → Verify → Ship
```

A fresh Claude session is started for each cycle. The spec file
itself is the handoff between sessions, so any session can pick
up where the previous left off without "remembering" prior
context.

## Hierarchy

Work organises into four levels:

```
Repo (this app)
 └─ Project (a wave of work — e.g. "MVP", "v2 redesign")
     └─ Stage (a coherent chunk within a project — 2–5 per project)
         └─ Spec (an individual implementable task)
              └─ Cycle (Frame → Design → Build → Verify → Ship)
```

The repo persists across all projects. A project is a bounded wave.
A stage is an epic-sized chunk within a project. A spec is one task.

## Session hygiene

Because one Claude assistant plays multiple roles, four habits keep
work coherent across sessions:

1. **Start a fresh session per cycle.** Especially design → build
   and build → verify.
2. **The spec file is the source of truth between sessions.** Don't
   rely on "as I said earlier" — the next session won't remember.
3. **Run a weekly review.** Without a second agent pushing back,
   drift compounds silently.
4. **Honest confidence values on decisions** — see
   [`AGENTS.md` §14](../AGENTS.md).

## Daily commands

```bash
just status                          # active project, stage, specs by cycle
just new-spec "title" STAGE-001      # scaffold a new spec
just advance-cycle SPEC-001 verify   # advance a spec's cycle
just archive-spec SPEC-001           # move a shipped spec to done/
just weekly-review                   # print the weekly review prompt
just specs-by-stage                  # group all specs by their stage
```

`just --list` shows every recipe.

## Where the conventions live

- [`AGENTS.md`](../AGENTS.md) — full conventions for working in
  this repo. The source of truth. See particularly:
  - **§6** — Cycle Model.
  - **§8/§9** — Coding and Testing Conventions.
  - **§10** — Git/PR Conventions.
  - **§11** — Domain Glossary (what we mean by "aggregate",
    "Store", "tap", and so on).
  - **§13** — Session Hygiene (the four habits above, expanded).
- [`GETTING_STARTED.md`](../GETTING_STARTED.md) — first-project
  walkthrough; if you're forking the framework into a new repo,
  start there.
- [`FIRST_SESSION_PROMPTS.md`](../FIRST_SESSION_PROMPTS.md) — the
  copy-paste prompts that drive each cycle.
- [`docs/CONTEXTCORE_ALIGNMENT.md`](./CONTEXTCORE_ALIGNMENT.md) —
  how this workflow maps to ContextCore's task taxonomy.

## Where to find what

| Looking for… | Look in |
|---|---|
| Architecture overview | [`docs/architecture.md`](./architecture.md) |
| Data model and schema | [`docs/data-model.md`](./data-model.md) |
| Full CLI reference | [`docs/api-contract.md`](./api-contract.md) |
| User tutorial | [`docs/tutorial.md`](./tutorial.md) |
| Decision log | [`decisions/`](../decisions/) |
| Repo-level rules | [`guidance/constraints.yaml`](../guidance/constraints.yaml) |
| Active project brief | [`projects/PROJ-001-mvp/brief.md`](../projects/PROJ-001-mvp/brief.md) |

# Getting Started — Claude + Implementer Variant

Step-by-step walkthrough for your first project. Assumes you just ran `just init` and chose the claude-plus-agents variant.

**Time to first shipped spec:** ~2 hours of your time, plus implementer build time.

---

## Before you start

- Pick your implementer tool (Kilo, Factory, AdaL, Cursor, etc.). Commit to one.
- API access for Claude and your implementer's models.
- A Git repo. Branch protection on `main` recommended.
- `just` installed (`brew install just` on macOS, `cargo install just` elsewhere).

---

## Step 1 — Sanity-check the scaffold

```bash
just --list      # see available commands
just info        # confirm variant and active project
just status      # see the example project and its one example spec
```

Explore the files. Read `AGENTS.md`, `docs/CONTEXTCORE_ALIGNMENT.md`, and
the example `PROJ-001-example-mvp` project. Then come back.

---

## Step 2 — PROJECT FRAME (5 min)

**Goal:** one-paragraph frame for what you want this repo to become. Go/kill decision.

1. Open a Claude session.
2. Copy **Prompt 1a: PROJECT FRAME** from `FIRST_SESSION_PROMPTS.md`.
3. Fill in your raw idea, paste into Claude.
4. Claude produces a frame + go/kill recommendation.

**Decision:**
- ✅ Go → paste the frame into `README.md` (replacing the placeholder). Commit.
- ❌ Kill → you just saved yourself weeks of wasted effort.

---

## Step 3 — FIRST PROJECT (15 min)

**Goal:** create your real first project (PROJ-001). The example project will be deleted.

1. Copy **Prompt 1b: PROJECT BRIEF** from `FIRST_SESSION_PROMPTS.md`.
2. Paste your project frame from Step 2 and the template will guide Claude
   to produce `projects/PROJ-001-<slug>/brief.md`.
3. Delete the example: `rm -rf projects/PROJ-001-example-mvp`
4. Move your real project into place (Claude will have written it).
5. Commit: `git commit -am "chore: first project brief"`.

---

## Step 4 — FIRST STAGE (15 min)

**Goal:** first stage of your first project. Typically "foundational infrastructure" for greenfield.

1. `just new-stage "foundational infra"` — this scaffolds the stage file.
2. Copy **Prompt 1c: STAGE FRAME** from `FIRST_SESSION_PROMPTS.md`.
3. Paste into Claude. It fills in the stage with a proposed spec backlog.
4. Review the backlog carefully. The order is your near-term roadmap.
5. Commit: `git commit -am "design: first stage framed"`.

---

## Step 5 — PROJECT DESIGN (60-90 min)

**Goal:** populate architecture, decisions, constraints, repo context.

1. Copy **Prompt 2a: PROJECT DESIGN** from `FIRST_SESSION_PROMPTS.md`.
2. Paste project frame + first stage.
3. Claude will:
   - Populate `docs/architecture.md` (with Mermaid diagram)
   - Populate `docs/data-model.md` and `docs/api-contract.md` if applicable
   - Create `decisions/DEC-*` files for meaningful choices
   - Update `guidance/constraints.yaml` with derived rules
   - Update `.repo-context.yaml` with your real stack
   - Update `AGENTS.md` with real commands, conventions, glossary
   - Delete the example `DEC-001-example-structured-logging.md` if not relevant
4. Review everything carefully.
5. Commit: `git commit -am "design: project architecture and decisions"`.

---

## Step 6 — FIRST SPEC (20-30 min)

1. Pick the simplest spec from your stage's backlog.
2. `just new-spec "logger module" STAGE-001`
3. Copy **Prompt 2b: SPEC** from `FIRST_SESSION_PROMPTS.md`.
4. Paste spec ID (what `just` just printed), title, stage ID.
5. Claude writes the spec content AND creates the matching handoff file.
6. Review the spec. Tight acceptance criteria? Concrete failing tests?
7. Review the handoff. Does its "Context the Receiving Agent Needs"
   section have everything the implementer needs?
8. Commit: `git commit -am "design: SPEC-001 ready for build"`.

---

## Step 7 — BUILD (implementer time)

1. Create the branch: `git checkout -b feat/spec-001-<slug>`.
2. Open your implementer tool.
3. Copy **Prompt 3: BUILD** from `FIRST_SESSION_PROMPTS.md`. Fill in IDs.
4. Paste. Let it work.
5. Watch for: clarification questions (answer them), scope creep (stop it),
   constraint violations (don't let slide).
6. When done, implementer fills in handoff's Completion section, runs
   `just advance-cycle SPEC-001 verify`, and opens PR.

---

## Step 8 — VERIFY (10-30 min)

1. Back in Claude.
2. Copy **Prompt 4: VERIFY** from `FIRST_SESSION_PROMPTS.md`.
3. Paste PR link or diff.
4. Claude returns: ✅ APPROVED / ⚠ PUNCH LIST / ❌ REJECTED.
5. Punch list → back to implementer, re-verify.
6. Approved → Ship.

---

## Step 9 — SHIP (10 min)

1. Copy **Prompt 5: SHIP** from `FIRST_SESSION_PROMPTS.md`.
2. Pre-ship checklist.
3. Merge, deploy, monitor briefly.
4. Answer three reflection questions. Append to spec as `## Reflection`.
5. `just advance-cycle SPEC-001 ship`
6. `just archive-spec SPEC-001` — this moves to `done/` and reminds you to update stage backlog.
7. Commit.

---

## Step 10 — REPEAT (5-9 for each spec in the stage)

Each subsequent spec: Design → Build → Verify → Ship.

`just status` shows progress at any time.

---

## Step 11 — STAGE SHIP

When all specs in your stage have shipped:

1. Copy **Prompt 1d: STAGE SHIP** from `FIRST_SESSION_PROMPTS.md`.
2. Claude reviews shipped specs and proposes stage-level reflection.
3. Append to the stage file. Update stage status to `shipped`.
4. Commit.

5. Start the next stage with Step 4.

---

## Step 12 — PROJECT SHIP

When all stages in your project have shipped:

1. Copy **Prompt 1e: PROJECT SHIP**.
2. Claude reviews shipped stages and proposes project-level reflection.
3. Append to `projects/PROJ-001-*/brief.md`. Update project status to `shipped`.
4. Commit.

5. Plan the next project (PROJ-002) with Step 2.

---

## Weekly rhythm

Once a week: `just weekly-review` → paste into Claude.

---

## Common first-week stumbles

**Implementer invents nonexistent libraries.** Your `AGENTS.md` stack section needs exact package names + versions.

**Spec was ambiguous; implementer guessed wrong.** Acceptance criteria weren't testable enough. Fix the template.

**Implementer violated a constraint.** It wasn't in the handoff's "applicable constraints" list. Handoff template is the fix.

**Verify keeps finding the same drift.** That's a template or `AGENTS.md` problem. Fix at the source.

**Specs are all complexity L.** You're splitting by feature, not task. Rule of thumb: if you can't write failing tests for it in 20 minutes, split.

**Stages are too big (>8 specs) or too small (1-2 specs).** Target 3-8 specs per stage.

**Projects drag on for months with growing scope.** That means your project's "What This Project Is" paragraph isn't bounded. Rescope or split into multiple projects.

---

## When to graduate

After 10-20 shipped specs, reassess. Triggers:
- Want multiple implementers in parallel → consider OpenSpec
- Decision log unwieldy (>30) → consider ContextCore `InsightQuerier`
- Compliance audit trail needed → full ContextCore
- Second app starting → extract this template's AGENTS.md + templates as a reusable base

Don't graduate prematurely. Each tool is weight.

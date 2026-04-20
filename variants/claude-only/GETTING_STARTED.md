# Getting Started — Claude-Only Variant

Step-by-step walkthrough for your first project. Assumes you just ran `just init` and chose the claude-only variant.

**Time to first shipped spec:** ~90 minutes to a few hours.

---

## Before you start

- Claude API access (Pro, Team, Enterprise, or API key).
- A Git repo. Branch protection on `main` recommended.
- `just` installed (`brew install just` on macOS, `cargo install just` elsewhere).
- **Most important:** a commitment to starting a NEW Claude session for each phase.

---

## Step 1 — Sanity-check the scaffold

```bash
just --list      # see available commands
just info        # confirm variant and active project
just status      # see the example project
```

Explore. Read `AGENTS.md`, `docs/CONTEXTCORE_ALIGNMENT.md`, and
`PROJ-001-example-mvp`. Then come back.

---

## Step 2 — PROJECT FRAME (5 min)

**Goal:** one-paragraph frame for what you want this repo to become.

1. Open a fresh Claude session.
2. Copy **Prompt 1a: PROJECT FRAME** from `FIRST_SESSION_PROMPTS.md`.
3. Paste your raw idea. Claude returns a frame + go/kill.

**Decision:**
- ✅ Go → paste the frame into `README.md`. Commit.
- ❌ Kill → close the tab.

---

## Step 3 — FIRST PROJECT (15 min)

1. Copy **Prompt 1b: PROJECT BRIEF**.
2. Paste project frame.
3. Claude produces `projects/PROJ-001-<slug>/brief.md`.
4. Delete the example: `rm -rf projects/PROJ-001-example-mvp`.
5. Move Claude's real project into place.
6. Commit.

---

## Step 4 — FIRST STAGE (15 min)

1. `just new-stage "foundational infra"` — scaffolds the stage.
2. Copy **Prompt 1c: STAGE FRAME**.
3. Claude fills in the stage with proposed spec backlog.
4. Review carefully. Commit.

---

## Step 5 — PROJECT DESIGN (60-90 min)

1. Copy **Prompt 2a: PROJECT DESIGN**.
2. Paste project frame + first stage.
3. Claude populates:
   - `docs/architecture.md` (with Mermaid)
   - `docs/data-model.md`, `docs/api-contract.md` (if applicable)
   - `decisions/DEC-*` files
   - Updates `guidance/constraints.yaml`, `.repo-context.yaml`, `AGENTS.md`
   - Deletes the example DEC-001 if not relevant
4. Commit.

---

## Step 6 — FIRST SPEC (20-30 min)

1. `just new-spec "logger module" STAGE-001`
2. Copy **Prompt 2b: SPEC**.
3. Claude writes the spec content, including the `## Implementation Context`
   section with everything the build session needs.
4. Review. Does the Implementation Context list every decision and
   constraint the build session (future you) needs?
5. Commit.

---

## Step 7 — BUILD (fresh Claude session!)

> **CRITICAL:** start a NEW Claude session. Don't continue from design.

1. Create the branch: `git checkout -b feat/spec-001-<slug>`.
2. Fresh Claude session.
3. Copy **Prompt 3: BUILD** from `FIRST_SESSION_PROMPTS.md`.
4. Paste spec ID and slug. Claude reads the spec (including its
   Implementation Context) and implements.
5. When done, Claude fills in the spec's `## Build Completion` section,
   runs `just advance-cycle SPEC-001 verify`, opens PR.

---

## Step 8 — VERIFY (another fresh session!)

> **ALSO CRITICAL:** start another NEW Claude session.

1. Fresh session.
2. Copy **Prompt 4: VERIFY**.
3. Paste PR link or diff.
4. Claude returns: ✅ APPROVED / ⚠ PUNCH LIST / ❌ REJECTED.

---

## Step 9 — SHIP (10 min)

1. Copy **Prompt 5: SHIP**.
2. Pre-ship checklist, merge, deploy.
3. Answer three reflection questions.
4. `just advance-cycle SPEC-001 ship`
5. `just archive-spec SPEC-001`
6. Commit.

---

## Step 10 — REPEAT

Each subsequent spec: new session for Build, new session for Verify.
Yes, it feels wasteful. Yes, it's worth it.

`just status` any time to see progress.

---

## Step 11 — STAGE SHIP

When all specs shipped: **Prompt 1d: STAGE SHIP**. Claude drafts stage reflection.

---

## Step 12 — PROJECT SHIP

When all stages shipped: **Prompt 1e: PROJECT SHIP**. Claude drafts project reflection.

Then plan PROJ-002 with Step 2.

---

## Weekly rhythm (non-optional here)

Once a week, always: `just weekly-review`. Without a second agent pushing back, drift compounds silently.

---

## Common first-week stumbles

**Reusing the same Claude session for design AND build.** Start a new one. It's not wasteful.

**Spec's Implementation Context section is too short.** It should list every decision, constraint, and prior spec the build session needs. If it feels redundant, you're doing it right — the build session has no design context.

**Verify finds nothing because "you" built and "you" reviewed.** Same session. Fresh next time.

**Reflections all say "nothing to change."** Either perfect (unlikely) or mailing it in. Write one-line real issues.

**Weekly review keeps finding the same drift.** `AGENTS.md` or `constraints.yaml` is wrong. Fix at source.

---

## When to graduate to claude-plus-agents

- Build is consistently the slowest/most expensive phase
- You want to run multiple builds in parallel
- Verify session can't keep up
- You've stopped reliably using fresh sessions

Migration is about an hour — see this variant's `README.md`.

---

## Trust the discipline

This variant's value is structure without tooling weight. Four non-negotiables:

1. New session per cycle
2. Spec is source of truth between sessions
3. Weekly review mandatory
4. Honest reflections

Skip any and it falls apart. Commit to all four and it's genuinely competitive with multi-agent setups at much lower operational overhead.

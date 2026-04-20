# AGENTS.md — Claude + Implementer Variant

Instructions for any AI agent working in this repository. Read this file first, every session.

> This file contains conventions only. For rules/constraints, see `/guidance/constraints.yaml`. For architectural rationale, see `/decisions/`. For waves of work against this app, see `/projects/`.

---

## 1. Repo Overview

- **Repo (the app):** [REPLACE: My App]
- **Purpose:** [REPLACE: one sentence]
- **Primary stakeholders:** [REPLACE]
- **Active project:** [REPLACE: PROJ-001 — MVP]

See `.repo-context.yaml` for structured metadata.

---

## 2. Work Hierarchy

```
REPO (the app — persists across all projects)
 └─ PROJECT (a wave of work: "MVP", "improvements", "v2 redesign")
     └─ STAGE (a coherent chunk within a project)
         └─ SPEC (an individual task)
              └─ HANDOFF (architect → implementer delegation record)
```

Key distinctions:

- The **repo** is the app. It persists. `AGENTS.md`, `/docs/`, `/guidance/`,
  `/decisions/` live at repo level because they accumulate across all
  projects.
- A **project** (`/projects/PROJ-*/`) is a bounded wave of work. Project
  artifacts (brief, stages, specs, handoffs) live inside the project
  folder.
- A **stage** is an epic-sized chunk within a project. A project typically
  has 2–5 stages.
- A **spec** is a single implementable task. It belongs to exactly one
  stage within one project.
- A **handoff** is an architect-to-implementer delegation document.

**Decisions persist at repo level**, even though they're often made
during a specific project. A decision like "we use pino for logging"
was made during PROJ-001 but binds PROJ-002 and PROJ-003 too. This is
intentional.

**Specs do not cross project boundaries.** If a task isn't finished
when a project ships, either finish it first or defer it explicitly into
the next project's brief.

---

## 3. Tech Stack

Replace this section with your actual stack. Be specific about versions.

- **Language:** [REPLACE]
- **Runtime:** [REPLACE]
- **Framework:** [REPLACE]
- **Database:** [REPLACE]
- **Testing:** [REPLACE]
- **Linter / Formatter:** [REPLACE]
- **Hosting:** [REPLACE]
- **CI:** [REPLACE]

---

## 4. Commands (exact)

These are the APP's commands. For template/workflow commands, see `justfile`.

```bash
[REPLACE: install command]
[REPLACE: dev command]
[REPLACE: test command]
[REPLACE: test single file command]
[REPLACE: lint command]
[REPLACE: typecheck command]
[REPLACE: build command]
```

---

## 5. Directory Structure

```
/
├── AGENTS.md                          # This file
├── CLAUDE.md                          # Pointer to AGENTS.md
├── README.md                          # Human-facing readme
├── GETTING_STARTED.md                 # First-project walkthrough
├── FIRST_SESSION_PROMPTS.md           # Phase prompts
├── .repo-context.yaml                 # Repo (app) metadata
├── .variant                           # "claude-plus-agents"
├── justfile                           # Commands: just status, just new-spec, etc.
├── scripts/                           # Shell scripts powering justfile
├── docs/                              # Architecture, data model, API contract
├── guidance/                          # Repo-level rules (across all projects)
│   ├── constraints.yaml
│   └── questions.yaml
├── decisions/                         # Repo-level DEC-* (across all projects)
├── projects/                          # Waves of work
│   ├── _templates/                    # Shared templates
│   │   ├── spec.md
│   │   ├── stage.md
│   │   ├── handoff.md
│   │   └── project-brief.md
│   ├── PROJ-001-<slug>/
│   │   ├── brief.md
│   │   ├── stages/
│   │   ├── specs/
│   │   │   └── done/
│   │   └── handoffs/
│   └── PROJ-002-<slug>/
│       └── (same structure)
└── src/                               # [REPLACE]
```

---

## 6. Cycle Model

Every spec moves through five cycles. **Cycles are tags, not gates** — edit any artifact anytime. The word "cycle" names what a spec goes through on its way to shipping.

| Cycle | Purpose | Who |
|---|---|---|
| **frame** | Go/no-go on the spec | Human + Claude (1 min) |
| **design** | Spec + failing tests + handoff | Claude (architect) |
| **build** | Make failing tests pass | Implementer agent |
| **verify** | Review + validation | Claude (reviewer) |
| **ship** | Merge, deploy, reflect, archive | Human + light agent |

Valid transitions:
```
frame → design → build → verify → ship
                   ↑       │
                   └───────┘ (verify sends back on punch list)
```

Projects and stages have lighter lifecycles (not full cycles):

- **Project status:** `proposed | active | shipped | cancelled`
- **Stage status:** `proposed | active | shipped | cancelled | on_hold`

A stage is `active` when its first spec enters design. `shipped` when
its spec backlog is complete AND the stage-level reflection is written.

---

## 7. Cross-Reference Rules

Every spec has these relationships, encoded in front-matter:

- `project.id` → the project it belongs to (e.g., `PROJ-001`)
- `project.stage` → the stage within that project (e.g., `STAGE-002`)
- `references.decisions` → DEC-* it was designed against
- `references.constraints` → constraints that apply
- `handoff.from_agent` / `handoff.to_agent` → roles in the delegation

When a spec references a DEC, the DEC does not reciprocally list the
spec. DECs are stable repo-level records; specs come and go.

---

## 8. Coding Conventions

- **Naming:** [REPLACE]
- **File organization:** [REPLACE]
- **Imports:** [REPLACE]
- **Error handling:** [REPLACE]
- **Logging:** [REPLACE]
- **Comments:** Explain *why*, not *what*.
- **No dead code.** Delete, don't comment out.

---

## 9. Testing Conventions

- Every new function gets at least one test.
- Test file naming: [REPLACE]
- Coverage expectations: [REPLACE]
- Must test: happy path, error cases, edge cases from acceptance criteria.
- Need not test: third-party internals, framework behavior.
- **TDD:** Tests live in the spec's `## Failing Tests` section, written
  during **design**, made to pass during **build**.

---

## 10. Git and PR Conventions

- **Branch:** `feat/spec-NNN-<slug>`, `fix/spec-NNN-<slug>`, `chore/<slug>`
- **One spec per branch, one PR per branch.**
- **Commits:** [REPLACE: e.g., Conventional Commits]
- **PR description must include:**
  - Project: `PROJ-NNN`
  - Stage: `STAGE-NNN`
  - Spec: `SPEC-NNN`
  - Handoff: `HANDOFF-NNN`
  - Decisions referenced: `DEC-NNN, DEC-MMM`
  - Constraints checked: `[list]`
  - New `DEC-*` files created during build

---

## 11. Domain Glossary

- **[REPLACE: Term]** — [REPLACE: Definition]

---

## 12. Cycle-Specific Agent Rules

### During **build** (implementer reads this)

Before writing code:
1. Read the `/projects/PROJ-*/handoffs/HANDOFF-*.md` for your spec.
2. Read the linked `SPEC-*.md`, `STAGE-*.md`, and the project's `brief.md`.
3. Read every `DEC-*` listed in the handoff's references.
4. Read `/guidance/constraints.yaml`; check rules for paths you'll touch.
5. If anything is ambiguous, add to `/guidance/questions.yaml` and stop.

When done:
1. Fill in the handoff's `## Completion` section (including reflection).
2. Update `handoff.status` → `completed`; update spec's `task.cycle` → `verify`.
3. Create `DEC-*` files for non-trivial implementer decisions.
4. Open PR following Section 10.

Shortcut: `just advance-cycle SPEC-NNN verify`.

### During **verify** (reviewer reads this)

Check:
1. Acceptance criteria all met and tested?
2. Failing tests from spec now pass?
3. No drift from referenced decisions?
4. No constraint violations?
5. Non-trivial implementer choices have accompanying `DEC-*`?
6. Implementer reflection answered (not mailed in)?

Output: ✅ APPROVED (with SHA) / ⚠ PUNCH LIST / ❌ REJECTED.

### During **ship**

Append a `## Reflection` block to the spec with three answers:
1. What would I do differently next time?
2. Does any template, constraint, or decision need updating?
3. Is there a follow-up spec to write now?

Then:
- Update the spec's `task.cycle` → `ship`.
- Run `just archive-spec SPEC-NNN` (moves to `done/`, updates stage).
- If stage backlog is complete, run the Stage Ship prompt.
- Commit.

---

## 13. Confidence Discipline

Decisions in `/decisions/` have an `insight.confidence` field (0.0–1.0).
Honest values matter — they drive these behaviors:

- **Design phase:** if Claude emits a decision at confidence < 0.7, it
  also adds an entry to `/guidance/questions.yaml` flagging it for
  further investigation.
- **Verify phase:** if a spec references any decision at confidence < 0.6,
  that's a yellow flag worth surfacing in the review.
- **Weekly review:** all decisions at confidence < 0.8 are listed with
  a note on whether recent work has strengthened or weakened them.

Use 1.0 only for decisions that are truly locked (tech stack choice
after it's been installed and working, for example). Most decisions
should land between 0.7 and 0.95.

---

## 14. Pointers

- Constraints: `/guidance/constraints.yaml`
- Open questions: `/guidance/questions.yaml`
- Decisions: `/decisions/`
- Projects: `/projects/`
- Templates: `/projects/_templates/`
- What we're building (architecture): `/docs/architecture.md`
- Phase prompts: `/FIRST_SESSION_PROMPTS.md`
- First walkthrough: `/GETTING_STARTED.md`
- Daily commands: run `just --list`

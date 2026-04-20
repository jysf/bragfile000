# AGENTS.md — Claude-Only Variant

Instructions for Claude working across all phases of this repository. Read this file first, every session.

> This variant assumes Claude plays every role: architect, implementer, reviewer. The context normally in a handoff document lives inside each spec's `## Implementation Context` section.

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
```

- The **repo** is the app. `AGENTS.md`, `/docs/`, `/guidance/`,
  `/decisions/` live at repo level because they accumulate across all
  projects.
- A **project** (`/projects/PROJ-*/`) is a bounded wave of work.
- A **stage** is an epic-sized chunk within a project (2–5 per project).
- A **spec** is a single implementable task. Belongs to one stage in
  one project.

In this variant, Claude plays architect and implementer in **separate
sessions**. The spec file itself carries all the context — see its
`## Implementation Context` section.

**Decisions persist at repo level.** A decision made during PROJ-001
binds PROJ-002 as well.

**Specs do not cross project boundaries.**

---

## 3. Tech Stack

Replace with your actual stack. Be specific about versions.

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
├── .variant                           # "claude-only"
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
│   │   └── project-brief.md
│   ├── PROJ-001-<slug>/
│   │   ├── brief.md
│   │   ├── stages/
│   │   └── specs/
│   │       └── done/
│   └── PROJ-002-<slug>/
└── src/                               # [REPLACE]
```

---

## 6. Cycle Model

Every spec moves through five cycles. **Cycles are tags, not gates**.

| Cycle | Purpose |
|---|---|
| **frame** | Go/no-go on the spec |
| **design** | Write the spec + failing tests + implementation context |
| **build** | Make failing tests pass |
| **verify** | Review + validation in one pass |
| **ship** | Merge, deploy, reflect, archive |

Valid transitions:
```
frame → design → build → verify → ship
                   ↑       │
                   └───────┘ (verify sends back on punch list)
```

**In this variant**, use **separate Claude sessions** for each cycle.
A fresh session prevents design-phase context from contaminating build
decisions, and a fresh verify session catches drift a continuation
session wouldn't.

Project and stage lifecycles are lighter:
- **Project status:** `proposed | active | shipped | cancelled`
- **Stage status:** `proposed | active | shipped | cancelled | on_hold`

---

## 7. Cross-Reference Rules

Every spec has these relationships, encoded in front-matter:
- `project.id` → the project it belongs to
- `project.stage` → the stage within that project
- `references.decisions` → DEC-* it was designed against
- `references.constraints` → constraints that apply

DECs are stable; specs come and go. DECs don't reciprocally list specs.

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
- **TDD:** Tests live in the spec's `## Failing Tests` section, written
  during **design**, made to pass during **build**.

---

## 10. Git and PR Conventions

- **Branch:** `feat/spec-NNN-<slug>`, etc.
- **One spec per branch, one PR per branch.**
- **Commits:** [REPLACE]
- **PR description must include:**
  - Project: `PROJ-NNN`
  - Stage: `STAGE-NNN`
  - Spec: `SPEC-NNN`
  - Decisions referenced, constraints checked, new `DEC-*` files

---

## 11. Domain Glossary

- **[REPLACE: Term]** — [REPLACE: Definition]

---

## 12. Cycle-Specific Rules

### During **build**

Start a **new Claude session**. Do not continue from the design session.

Before writing code:
1. Read the spec's `## Implementation Context` section.
2. Read every `DEC-*` it references.
3. Read the parent `STAGE-*.md` and project `brief.md`.
4. Read `/guidance/constraints.yaml`.
5. If anything is ambiguous, add to `/guidance/questions.yaml` and stop.

When done:
1. Fill in spec's `## Build Completion` (including reflection).
2. `just advance-cycle SPEC-NNN verify`.
3. Create `DEC-*` files for non-trivial build decisions.
4. Open PR.

### During **verify**

Start **another new Claude session**. Do not reuse build session.

Check: acceptance criteria met? tests pass? no decision drift? no
constraint violations? non-trivial choices have DEC-*? build reflection
answered honestly?

Output: ✅ APPROVED / ⚠ PUNCH LIST / ❌ REJECTED.

### During **ship**

Append `## Reflection` to spec. Three answers. Then
`just archive-spec SPEC-NNN`. If stage backlog is complete, run the
Stage Ship prompt.

---

## 13. Session Hygiene (claude-only specific)

Because one agent plays multiple roles, context contamination is a real
risk. Four habits keep it at bay:

1. **New session per cycle where possible.** Especially design → build
   and build → verify.
2. **Never reference "as I said earlier"** in later cycles. The spec
   is the source of truth.
3. **Weekly review is non-optional.** Without a second agent pushing
   back, drift compounds silently. Run `just weekly-review`.
4. **Honest confidence values on decisions.** See Section 14.

---

## 14. Confidence Discipline

Decisions have an `insight.confidence` field (0.0–1.0). Honest values drive:

- **Design:** decisions at confidence < 0.7 also create a question in
  `/guidance/questions.yaml`.
- **Verify:** specs referencing decisions at confidence < 0.6 get a
  yellow flag.
- **Weekly review:** all decisions < 0.8 are listed with strength/weakness trend.

Most decisions should land between 0.7 and 0.95. 1.0 only for truly locked choices.

---

## 15. Pointers

- Constraints: `/guidance/constraints.yaml`
- Open questions: `/guidance/questions.yaml`
- Decisions: `/decisions/`
- Projects: `/projects/`
- Templates: `/projects/_templates/`
- Architecture: `/docs/architecture.md`
- Phase prompts: `/FIRST_SESSION_PROMPTS.md`
- First walkthrough: `/GETTING_STARTED.md`
- Daily commands: run `just --list`

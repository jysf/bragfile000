# First-Session Prompts — Claude + Implementer Variant

Copy-paste-ready prompts for each cycle at the project, stage, and spec levels. Fill in the bracketed parts before pasting.

**Prompts at a glance:**
- **1a:** Project Frame (starting a new project)
- **1b:** Project Brief (approved frame → populated brief.md)
- **1c:** Stage Frame (starting a new stage)
- **1d:** Stage Ship (closing out a completed stage)
- **1e:** Project Ship (closing out a completed project)
- **2a:** Repo/Project Design (initial architecture)
- **2b:** Spec Design (write one spec + handoff)
- **3:** Build (hand off to implementer)
- **4:** Verify (review PR)
- **5:** Ship (merge + reflect)
- **6:** Weekly Review

---

## Prompt 1a — PROJECT FRAME

> **Use when:** Starting a new project (wave of work) in this repo.
> **Agent:** Claude (architect).
> **Time:** 5 min.

```
I want to frame a potential project before committing to it.

A "project" in this repo is a wave of work. Prior projects:
[REPLACE: list any prior projects or say "none, this is the first"]

My raw idea for this project: [REPLACE: 1-3 sentences]

Produce a one-paragraph frame with exactly four elements:
- What: what this wave of work delivers
- For: who the user is
- Why now: what makes this the right wave, now
- Success: one concrete outcome that would mean it worked

Your task:
1. Ask clarifying questions. Be skeptical — it's cheaper to kill a
   project here than after design.
2. Produce the one-paragraph frame.
3. Give a go/kill recommendation with one sentence of reasoning.
4. Suggest a short-slug name for the project directory
   (e.g., "mvp", "improvements-q2", "redesign").

If the idea isn't ready, say "kill" and explain.
```

---

## Prompt 1b — PROJECT BRIEF

> **Use when:** Project frame approved. Producing the full brief.md.
> **Agent:** Claude (architect).
> **Time:** 15 min.

```
Project frame approved. Produce the full brief.md.

Read these first:
- /README.md
- /AGENTS.md (especially the Work Hierarchy section)
- /.repo-context.yaml
- /projects/_templates/project-brief.md
- /projects/PROJ-001-example-mvp/brief.md (delete later)
- Any existing shipped projects in /projects/*/brief.md for continuity

Project frame (from Prompt 1a):
[REPLACE: paste frame]

Your task:

1. Determine the next PROJ-ID. If this is the first real project,
   use PROJ-001. Otherwise find the highest existing ID and increment.

2. Ask clarifying questions, especially:
   - What's the target ship date?
   - What stages do you anticipate? (3-5 typical)
   - Dependencies on prior projects?

3. Create the project directory:
   projects/PROJ-NNN-<slug>/
     ├── brief.md
     ├── stages/
     ├── specs/
     │   └── done/
     └── handoffs/

4. Populate brief.md from the template, including:
   - "What This Project Is" paragraph
   - "Why Now" justification
   - Success criteria (3-5 concrete)
   - Scope (in/out)
   - Stage Plan (2-5 stages, ordered, one-line summary each)
   - Dependencies

5. If there's an example project folder
   (projects/PROJ-001-example-mvp/), propose deleting it.

Stop and let me review before we frame the first stage.
```

---

## Prompt 1c — STAGE FRAME

> **Use when:** Starting a new stage within the active project.
> **Agent:** Claude (architect).
> **Time:** 15 min.

```
I want to frame a new stage in the active project.

Read first:
- /AGENTS.md
- /projects/<active-project>/brief.md
- /projects/_templates/stage.md
- Any already-shipped or in-progress stages in the active project
- Any relevant /decisions/DEC-*.md files

I already ran `just new-stage "<title>"` which created the stage file
with IDs filled in. The file is at:
[REPLACE: paste path, e.g. projects/PROJ-001-mvp/stages/STAGE-001-foundational-infra.md]

The stage I want to define:
- Title: [REPLACE]
- Why this stage is next: [REPLACE]
- Dependencies I know of: [REPLACE or "unknown"]

Your task:

1. Ask clarifying questions. Focus on: what does "done" mean for this
   stage? What does it unblock? What would cause you to split it?

2. Populate the stage file from the template:
   - "What This Stage Is" paragraph
   - "Why Now"
   - Success criteria (3-5)
   - Scope (in/out)
   - Proposed Spec Backlog (3-8 specs, ordered, each with S/M/L)
   - Design Notes (patterns across specs)
   - Dependencies

3. Flag any backlog entry that looks like complexity L — these should
   be split before building.

4. If the stage feels thin (under 3 specs) or sprawling (over 8),
   recommend rescoping.

Stop and let me review before any specs.
```

---

## Prompt 1d — STAGE SHIP

> **Use when:** All specs in a stage have shipped.
> **Agent:** Claude (architect).
> **Time:** 15-20 min.

```
STAGE-NNN is ready to ship. All its specs are in /projects/<active-project>/specs/done/.

Read:
- /projects/<active-project>/stages/STAGE-NNN-<slug>.md
- Every spec in done/ that belongs to this stage (check
  project.stage: STAGE-NNN in front-matter)
- Their Reflection sections especially

Your task:

1. Check the stage's "Success Criteria" against what actually shipped.
   Did we deliver the outcome?

2. Summarize in 3 sentences:
   - Built vs planned
   - Took longer or easier than expected
   - Emergent integration behavior

3. Propose answers for the stage-level reflection fields (## Stage-Level
   Reflection section in the stage file).

4. Flag follow-up work: should any of it become a new stage in the
   current project? A spec in the NEXT stage? Deferred to a future
   project?

5. Propose updates to /AGENTS.md, /guidance/constraints.yaml, or any
   template based on patterns across shipped specs.

I'll review your proposals and write them into the stage file myself.
```

---

## Prompt 1e — PROJECT SHIP

> **Use when:** All stages in a project have shipped.
> **Agent:** Claude (architect).
> **Time:** 20-30 min.

```
PROJ-NNN is ready to ship. All its stages are shipped.

Read:
- /projects/PROJ-NNN-<slug>/brief.md
- Every stage in that project
- Every shipped spec in that project
- Their Reflection sections
- /decisions/DEC-* emitted during this project (most recent N)

Your task:

1. Check the project's "Success Criteria" against what actually
   shipped. Did the wave of work deliver its intended outcome?

2. Summarize in 3-5 sentences:
   - Scope evolution (what changed between brief and reality)
   - Stages that went smoothly vs painfully
   - Any major decisions that were superseded mid-project

3. Propose answers for the project-level reflection (## Project-Level
   Reflection section in the brief).

4. Identify deferred work: what came up during this project that
   should become input to the NEXT project's frame?

5. Recommend: should the active project folder be marked shipped
   (status: shipped) and a new project started, or is there more
   work that should extend into a "PROJ-NNN phase 2"?

6. Propose updates to /AGENTS.md, /guidance/*, or templates.

I'll review and write the proposals into the brief.
```

---

## Prompt 2a — REPO/PROJECT DESIGN

> **Use when:** Project brief + first stage approved; ready to architect.
> **Agent:** Claude (architect).
> **Time:** 60-90 min.

```
Project brief and first stage are approved. Moving to repo-level design.

Read:
1. /README.md
2. /AGENTS.md
3. /.repo-context.yaml
4. /guidance/constraints.yaml
5. /decisions/DEC-001-example-structured-logging.md (format reference)
6. /decisions/_template.md
7. /projects/_templates/spec.md
8. /projects/_templates/handoff.md
9. /projects/<active-project>/brief.md
10. /projects/<active-project>/stages/<first-stage>.md

Project frame / brief summary: already populated; read from files above.

Additional hard constraints:
- Tech choices forced on us: [REPLACE or "none"]
- Timeline: [REPLACE]
- Compliance: [REPLACE or "none"]
- Explicit non-goals: [REPLACE]

Your task:

1. Ask clarifying questions on ambiguities in the architecture space.

2. Populate repo-level docs:
   - /docs/architecture.md (with Mermaid diagram)
   - /docs/data-model.md (or delete if not applicable)
   - /docs/api-contract.md (or delete if not applicable)

3. For every meaningful decision, create
   /decisions/DEC-NNN-<slug>.md using the template. Honest confidence
   values. Use the active project's ID as the origin project.

4. Update /guidance/constraints.yaml with rules derived from your
   decisions. Remove example constraints that don't apply.

5. Update /.repo-context.yaml with real stack values.

6. Update /AGENTS.md with the real tech stack, exact commands,
   conventions, domain glossary.

7. Delete the example files (or keep if they're still relevant):
   - /decisions/DEC-001-example-structured-logging.md
   - /projects/<active-project>/specs/SPEC-001-example-project-logger.md
     (if still present)
   - /projects/<active-project>/handoffs/HANDOFF-001-example-project-logger.md
     (if still present)

8. Cross-check the first stage's backlog against the architecture
   you just designed. Propose updates to the stage backlog if needed.

Stop after step 8 and let me review.
```

---

## Prompt 2b — SPEC DESIGN

> **Use when:** Ready to write one specific spec from a stage's backlog.
> **Agent:** Claude (architect).
> **Time:** 15-30 min.

```
Please write the content of SPEC-NNN for the task
"[REPLACE: task title]" from STAGE-MMM's backlog.

I already ran `just new-spec "<title>" STAGE-MMM` which created the
spec file with IDs filled in:
[REPLACE: paste path]

Cycle starts at "design". Set to "build" after you also create the
handoff.

Read first:
- /AGENTS.md
- /projects/<active-project>/brief.md
- /projects/<active-project>/stages/STAGE-MMM-<slug>.md
- /docs/architecture.md
- /docs/data-model.md and /docs/api-contract.md (if applicable)
- /guidance/constraints.yaml
- All /decisions/DEC-*.md relevant to this task
- Any related specs (check stage backlog for shipped specs)

When writing the spec:
- Target complexity S or M. If L, split into multiple specs first.
- Acceptance criteria must be testable.
- Failing tests must be concrete: file paths + what each asserts.
- List applicable decisions and constraints in front-matter references.
- If you make NEW decisions while writing, emit
  /decisions/DEC-NNN-<slug>.md files for each.

Then create the handoff file at
/projects/<active-project>/handoffs/HANDOFF-NNN-<slug>.md pointing
the implementer at the spec. The handoff's "Context the Receiving
Agent Needs" section is critical.

Finally, run `just advance-cycle SPEC-NNN build` (or update
task.cycle to "build" manually) and update the stage's Spec Backlog
section to reflect the new spec ID.

Stop and let me review before I hand off to the implementer.
```

---

## Prompt 3 — BUILD

> **Use when:** Spec + handoff ready; running the implementer.
> **Agent:** Your implementer (Kilo / Factory / AdaL / etc.).
> **Time:** Task-dependent.

```
You're the implementer for a spec-driven workflow. Cycle: build.

Read files in order, before writing code:

1. /AGENTS.md — conventions you must follow.
2. /projects/<active-project>/handoffs/HANDOFF-NNN-<slug>.md — your
   handoff. Follow its "Context the Receiving Agent Needs" section
   exactly.
3. /projects/<active-project>/brief.md — the project context.
4. /projects/<active-project>/stages/STAGE-MMM-<slug>.md — the stage.
5. Every decision listed in the handoff's references.
6. /guidance/constraints.yaml — constraints for paths you'll touch.
7. /projects/<active-project>/specs/SPEC-NNN-<slug>.md — the spec.

Implement:
- Make the failing tests pass.
- Don't violate constraints. If one needs breaking, STOP and ask.
- For non-trivial decisions (library, pattern, API shape), create
  /decisions/DEC-NNN-<slug>.md files. Use honest confidence values.
- If the spec is ambiguous, STOP and ask. Don't guess.

When done:
1. Fill in the handoff's "Completion" section INCLUDING the three
   implementer reflection questions. Not optional.
2. Update handoff.status → "completed".
3. Run: just advance-cycle SPEC-NNN verify
4. Open PR from feat/spec-NNN-<slug>.
5. PR description references: project ID, stage ID, spec ID, handoff
   ID, decisions used, constraints checked, new DEC-* files.
```

---

## Prompt 4 — VERIFY

> **Use when:** Implementer opens a PR.
> **Agent:** Claude (reviewer).
> **Time:** 10-30 min.

```
Cycle: verify. Implementer completed SPEC-NNN.

Review the PR: [REPLACE: paste PR link or full diff]

Read against:
- /projects/<active-project>/specs/SPEC-NNN-<slug>.md (acceptance
  criteria met? tests pass?)
- /projects/<active-project>/handoffs/HANDOFF-NNN-<slug>.md
  (deliverables, return criteria, reflection answered?)
- /projects/<active-project>/stages/STAGE-MMM-<slug>.md (does this
  advance the stage as intended?)
- Decisions in handoff references (drift? superseding DEC-*?)
- /guidance/constraints.yaml (violations?)

Flag:
- Untested acceptance criteria
- Decision drift without supersession
- Constraint violations
- Non-trivial choices missing DEC-*
- Mailed-in reflection ("nothing was unclear" is suspicious)
- Any decisions referenced at confidence < 0.6 (yellow flag)
- Follow-up specs this PR implies

Output exactly ONE of:

✅ APPROVED — merge at commit <SHA>. No changes needed.

⚠ PUNCH LIST:
   1. [specific item]
   2. [specific item]
   (Don't rewrite the code yourself. List what should change.)

❌ REJECTED because [fundamental reason].
   Recommended: [revise spec | split | revisit design].
```

---

## Prompt 5 — SHIP

> **Use when:** PR approved.
> **Agent:** Human + light agent.
> **Time:** 5-10 min.

```
Cycle: ship. PR for SPEC-NNN is approved.

Pre-ship checklist:
[ ] CI passing?
[ ] Deployment steps known?
[ ] Rollback plan?
[ ] CHANGELOG or release note?

After merge + deploy, help me answer three reflection questions.
Append as a "## Reflection" block at the bottom of the spec:

1. What would I do differently next time?
   [REPLACE: your answer]

2. Does any template, constraint, or decision need updating?
   [REPLACE: your answer — call out specifically so I can update now]

3. Is there a follow-up spec I should write before I forget?
   [REPLACE: your answer — if yes, one-line summary]

After I paste answers:
- Format as ## Reflection block in the spec
- Run: just advance-cycle SPEC-NNN ship
- Run: just archive-spec SPEC-NNN
- If I mentioned a template/constraint/decision update, propose the
  specific edit.
- If I mentioned a follow-up spec, add to the stage's backlog.
- Update the parent STAGE's backlog (mark shipped).

If this was the LAST spec in STAGE-MMM's backlog, remind me to run
Prompt 1d (Stage Ship) next.
```

---

## Prompt 6 — WEEKLY REVIEW

> **Use when:** Weekly, or every 5-10 specs.
> **Agent:** Claude (architect).
> **Time:** 20-30 min.
>
> **Shortcut:** `just weekly-review` pre-loads context — paste its output
> into Claude as the prompt body.

```
Weekly review of the repository.

Read:
- All /decisions/DEC-*.md (repo-level, across projects)
- /guidance/constraints.yaml
- /guidance/questions.yaml
- /AGENTS.md
- /projects/<active-project>/brief.md
- All stages in the active project
- Reflection sections of recent specs in specs/done/

Produce a short report covering:

1. Stale decisions — DEC-* that should be superseded given recent
   learning. Flag with reasoning; don't supersede yet.

2. Low-confidence decisions — any DEC at confidence < 0.8 where
   recent work has strengthened or weakened the underlying bet.

3. Missing constraints — patterns in reflections or PRs that should
   be formalized. Propose specific YAML entries.

4. Resolved questions — items in questions.yaml that recent work
   has answered. Flag which DEC would formalize each.

5. AGENTS.md drift — repo behavior contradicting AGENTS.md. Propose
   specific edits.

6. Template improvements — friction visible in reflections. Specific:
   "add field X to /projects/_templates/spec.md because Y."

7. Stage health — active stage progressing? Stalled? Rescope needed?

8. Cycle health — any cycles consistently skipped or painful?
   Is Frame used as a kill gate? Reflections getting mailed in?

9. Project health — is the active project still well-scoped?
   Scope creep? Time to declare done and start PROJ-NN+1?

Tight report. Actionable in 10 minutes.
```

---

## Quick reference

| You just... | Use this prompt |
|---|---|
| Had a new project idea | 1a (Project Frame) |
| Approved a project frame | 1b (Project Brief) |
| Ready to frame a new stage | 1c (Stage Frame) |
| Brief + first stage approved | 2a (Repo/Project Design) |
| Ready to write a spec | 2b (Spec Design) |
| Spec + handoff ready for implementer | 3 (Build) |
| Implementer opened PR | 4 (Verify) |
| Got APPROVED from Verify | 5 (Ship) |
| All specs in a stage shipped | 1d (Stage Ship) |
| All stages in a project shipped | 1e (Project Ship) |
| It's Friday | 6 (Weekly Review) |

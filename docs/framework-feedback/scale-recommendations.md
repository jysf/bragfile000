# Spec-driven framework — scaling recommendations

*Written 2026-04-20 in response to the question: "We're using this
to build a small app — can we use the same basic process to build
a much bigger and more complex app?" Companion to
`process-feedback.md`. Intended for the template author; not part
of the `bragfile` project itself.*

*Perspective: one agent (Claude) having shipped 7 specs in a solo
personal-CLI project under the `claude-only` variant. Scale claims
below are inference, not experience — caveat accordingly.*

---

## Short verdict

**Yes, with meaningful adaptations.** The framework's core —
hierarchy, cycles, DECs, constraints, accumulating lessons, fresh
sessions, spec-as-source-of-truth — is fundamentally sound and
scales. What breaks isn't the principles; it's operational load and
discoverability.

## What scales as-is

- **Hierarchy (Repo / Project / Stage / Spec / Cycle).** This is
  SAFe-adjacent; works at any size from a CLI tool to a company.
- **Cycle model.** Frame → Design → Build → Verify → Ship is good
  engineering, full stop. Larger orgs might add a pre-Frame "Plan"
  cycle for cross-team coordination, but the core rhythm holds.
- **DECs with confidence values.** Architecture Decision Records
  are already an industry practice — this adds calibration honesty,
  which matters *more* at scale because decisions propagate to
  people who don't have your context.
- **AGENTS.md accumulating rules.** Codifying lessons scales with
  any size of team.
- **Fresh sessions for Design → Build → Verify.** At scale this is
  MORE valuable, because more work is autonomous and context
  contamination compounds.
- **Spec-as-handoff document.** Fundamental to async / multi-agent
  collaboration. Scales cleanly.

## What breaks at scale

1. **One agent plays every role.** The `claude-only` variant
   assumes one developer. At scale you need the
   `claude-plus-agents` variant — or humans-plus-Claude, or
   multiple agents each owning a role. The framework already
   anticipates this.

2. **AGENTS.md bloat.** 300 lines after 8 specs. At 100 specs that
   's 1000+ lines, and every fresh session reads it. At some
   threshold you need:
   - Section folding or collapsible structure.
   - Per-stage or per-module AGENTS.md overrides (only the team
     working on this area reads the full rules).
   - Automated extraction — e.g. `just rules-for internal/storage`
     should return just the rules relevant to that path.

3. **DEC proliferation.** At 40+ DECs, the `references.decisions`
   lists in specs become unwieldy. Needs:
   - An auto-generated `decisions/INDEX.md` with one-line summaries.
   - An explicit **deprecated** status (not just `superseded_by`)
     for DECs that are stale but historically important.
   - Tag-based filtering (e.g. "all storage DECs") to narrow
     reading load.

4. **Constraint proliferation.** 11 constraints is manageable. At
   50+ you want:
   - **Tiers**: universal (security), repo-level, path-scoped.
   - **Automated linting**: pre-commit hooks that parse
     `constraints.yaml` and enforce the `paths:` glob rules for
     enforceable ones (e.g., "no `database/sql` under
     `internal/cli/**`"), not honor-system.
   - A `just constraint-check` that runs the enforceable rules.

5. **Cross-project coordination.** The framework supports
   PROJ-001 → PROJ-002, but at scale you have multiple projects in
   flight simultaneously, sometimes sharing specs' implications.
   Current design has no explicit dependency tracking beyond the
   brief's prose. Needs:
   - `depends_on: [PROJ-002/SPEC-005]` in spec front-matter.
   - Cross-project dependency reports.

6. **Discoverability.** `find /specs/done -name '*.md'` doesn't
   scale past ~20 shipped specs. New joiners (human or agent)
   can't find prior art efficiently. Needs:
   - Indexing, search, tagging.
   - `just find-spec "auth refactor"` style commands.

7. **Verify fatigue.** With 100+ specs over a year, verify sessions
   become routine and rubber-stamping risk grows. The framework's
   "honest reflection" discipline needs structural support at
   scale — options:
   - Random-sampling peer review in addition to per-spec verify.
   - CI-enforced structural checks (did build complete the
     reflection? is it >50 words? does it avoid generic phrases
     like "nothing was unclear"?).

8. **Atomic one-spec-per-PR assumption breaks for cross-cutting
   changes.** "Add tracing to all services" legitimately touches
   N specs' worth of code. The framework needs a notion of
   **meta-spec** or **epic** that coordinates atomic children —
   not one-PR-per-epic, but one-PR-per-child-spec with the epic
   as the coordinating artifact.

9. **Long-running architecture drift.** Our 8 DECs are 2 days old;
   they haven't been stress-tested by time. At a year out, some
   of DEC-001's assumptions (e.g. "single user, O(hundreds) of
   rows") will no longer hold. The `Revisit if:` sections are
   honest but **unenforced** — no one reviews them unless
   something breaks. Needs:
   - A quarterly consolidation cycle that reviews all DECs with
     `confidence < 0.8` and all with `Revisit if:` triggers that
     may have fired.
   - Automated triggers (e.g., "row count passed 10,000 — DEC-001's
     revisit condition fires, notify the team").

10. **Test coverage taxonomy.** Our MVP has unit tests. A bigger
    app needs unit + integration + e2e + load + security. The
    framework's "Failing Tests" section assumes one class of test.
    Needs extension — per-spec declaration of which test tier(s)
    apply, enforced by CI.

## Concrete adaptations by project tier

### Small app (1 dev, 1–2 months) — framework works as-is

Current `bragfile` scale. No adaptations needed beyond the small
bug-fixes and quick-spec template option noted in
`process-feedback.md`.

### Medium app (1–3 engineers, 3–6 months)

1. **Switch to `claude-plus-agents` variant.** Explicit architect-
   implementer separation even when both are Claude sessions.
2. **Add `decisions/INDEX.md` and `specs/INDEX.md`.** Auto-
   generated, updated on ship. ~30 minutes of tooling.
3. **Add `just constraint-check`** that parses `constraints.yaml`
   and runs glob-based greps for enforceable rules.
4. **Add a quarterly consolidation cycle.** Review stale DECs,
   prune AGENTS.md redundancy, close resolved questions. The
   framework has `weekly-review`; this is `quarterly-consolidation`.

### Larger app (team + sustained work across quarters)

5. **Introduce `epic` as a tier between Project and Stage.** Cross-
   cutting initiatives that span multiple stages. Not one-PR-per-
   epic; one-PR-per-child-spec with the epic as the coordinating
   artifact.
6. **Formalize spec ownership.** `assignee:` field in frontmatter;
   `just unassigned` command to show open work.
7. **CI enforcement beyond tests.** Pre-commit hooks verify: every
   PR has a `SPEC-NNN` reference, every new dep has a
   corresponding DEC, every DEC passes schema validation.
8. **Multi-repo patterns.** If the app spans services, decide
   where AGENTS.md / constraints.yaml live — per-repo, shared, or
   hybrid. Non-trivial architectural question.

### Very large app (50+ engineers)

9. **Consider whether the framework is still earning its keep vs.
   conventional practices.** At very large scale the overhead of a
   shared AGENTS.md + DEC corpus might exceed the coordination
   benefit. The framework is optimized for small-team + AI-agent
   collaboration; at very large scale you might want only the best
   parts (DEC discipline, constraint YAML, spec-as-handoff) grafted
   onto standard practices, not the whole framework.

## Practical advice

**The best test of "does this scale" is trying it at the next tier
up.** `bragfile` is ~1000 lines of Go across 4 shipped specs. The
next tier is something like 5k–10k lines — a small service with
HTTP + DB + auth. Build that same way (multi-stage project, same
framework) and note what hurts. You'll have real data instead of
speculation.

**Strong hunch**: the framework will hold at that tier with minor
adaptations (mostly the indexing/discoverability items under
"Medium app" above). It would need more substantial reshaping for
50k+ line apps, but the core ideas — specs as source of truth,
accumulating lessons, honest DECs, fresh-session discipline — stay
right.

## One signal to watch for

The `## Reflection (Ship)` Q2 answers ("Does any template,
constraint, or decision need updating?").

- If they say "no, framework still fits" for 5 specs in a row, the
  framework is in steady state and you can scale further without
  restructuring.
- If they start saying "yes, this constraint should be tiered / this
  DEC should be archived / AGENTS.md needs restructuring" — that's
  the framework telling you it's hit its next growth threshold.
  Pay attention.

The framework's own reflection mechanism is the earliest signal of
when it's outgrowing itself.

## Summary

The framework is approximately right for its target scale (small
team + AI agents). It needs **operational adaptation, not
principle replacement**, for medium/large scale. The most
underdeveloped areas for scaling are:

1. Discoverability (indexing, search)
2. Archival / pruning of stale artifacts
3. Enforcement (constraint linting, not just honor-system)
4. Cross-project and cross-agent coordination

Of these, #1 and #2 are tooling projects; #3 is a few CI scripts;
#4 is the most substantive redesign. None are blockers for
starting at the next tier up.

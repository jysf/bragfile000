# Spec-driven process — feedback after 8 specs / 1 stage close

*Written 2026-04-20 while working through STAGE-002 of the `bragfile`
MVP (personal CLI project). Feedback is from the agent (Claude) that
played every role — architect, implementer, reviewer — across design,
build, verify, and ship cycles under the `claude-only` variant of the
spec-driven-template. Intended for the template author; not part of
the `bragfile` project itself.*

*Scope of experience: PROJ-001 through ship of SPEC-007 (4 shipped
specs in STAGE-001, 3 shipped in STAGE-002 at time of writing).*

---

## Pace assessment: healthy — don't push faster

Average ~2.5–3 hours per spec (design + build + verify + ship, across
fresh sessions). For a personal tool with accumulating quality
discipline, that's remarkably productive. Most solo side-projects
this size wouldn't have 8 DECs, 6 formalized lessons, and a passing
test suite with meaningful edge-case coverage.

The framework's overhead is real but it's buying real things:

- **2 of 8 specs caught real defects during verify that would have
  silently shipped** without fresh-session discipline:
  - SPEC-001: tests collapsed stdout + stderr into one buffer, so
    the `stdout-is-for-data-stderr-is-for-humans` constraint was
    unverified empirically.
  - SPEC-007: spec's Notes-for-the-Implementer offered "Either is
    acceptable" for a test helper, and one option would have
    imported `database/sql` under `internal/cli/`, violating a
    blocking constraint.
  Both were design-session blind spots. Without a fresh verify
  session, both would be in main today.
- **Lesson compounding is real.** SPEC-004, SPEC-005, SPEC-006 all
  shipped clean-first-try specifically because they inherited rules
  earned by earlier specs. AGENTS.md §9's accumulated bullets + §10
  + §12 are all earned through punch lists and reflections, not
  invented up front.

"Move faster" isn't the right lens for this scale of project. Cost
per spec is dominated by spec-writing (~30 min design) + human merge
+ ship mechanics. Pushing harder on batch size or skipping verify
erodes the thing that's paying interest.

## What's genuinely working

- **Prescriptive Implementation Context = mechanical build.** SPEC-005
  pre-locked the shorthand-flag letters and the `Long` description's
  copy; the build session spent ~30 min on mostly mechanical
  transcription. SPEC-006 did the same. This pattern is the single
  highest-leverage design-phase investment.
- **DEC system with honest confidence values.** DEC-004 at 0.65 with
  a paired question (`tags-storage-model` in `questions.yaml`)
  correctly gated SPEC-007's design — the revisit trigger fired, the
  question got answered, DEC-004 stayed in place with an amended
  Validation section. DEC-007 informally extending via SPEC-006 and
  then formally having its References amended during SPEC-006 ship
  also went smoothly.
- **Ship commits direct to main — no separate PR.** The one-spec-
  per-PR constraint is per-spec (feature branch), not per-commit.
  Ship work (reflection + archive + AGENTS.md edit + lesson
  promotion) as one commit on main is the right scoping. Saves
  significant ceremony.
- **Fresh-session discipline — with a caveat.** See #5 under
  "Friction" below.

## Friction I've actually noticed

1. **`just archive-spec` prints a false "All specs for STAGE-XXX are
   shipped" message after every archive.** The heuristic is "empty
   active `specs/` dir," not "all backlog items completed." Fires
   spuriously on every ship. Worked around ~6 times. Fix: compare
   shipped-archive count against stage-backlog `[x]` markers (or
   count of backlog items).

2. **Spec-numbering vs stage-backlog aspirational numbering
   misalignment.** When a stage backlog is written before the specs
   it references exist, numbering in the backlog (e.g. "SPEC-007
   show") becomes aspirational — `just new-spec` picks the actual
   next-available number (SPEC-006), forcing a backlog renumber.
   Low-frequency, annoying when it hits. Fix: use labels in stage
   backlogs (`brag-show`, `list-filters`) rather than predicted
   numbers.

3. **DEC emission timing is inconsistent.** DEC-007 emitted during
   SPEC-003 *build* (discovered when cobra's required-flag
   validator didn't compose with `ErrUser`). DEC-008 emitted during
   SPEC-007 *design* (anticipated up front). Both legitimate, but
   no rule says which is preferred. Mild friction on when to stop
   and write a DEC vs barrel ahead.

4. **Ship reflection Q1 feels ceremonial when truly nothing changed.**
   SPEC-004 and SPEC-006's answers were essentially "the pattern
   that worked keeps working." Honest but repetitive. Consider
   letting Q1 be skippable when the honest answer is "nothing new"
   — the template currently forces a ritual that sometimes has no
   content.

5. **"Fresh session" is weaker than it looks with a single agent.**
   Claude is the same agent across sessions even when conversation
   history is cleared. The verify sessions catch real defects
   because the spec file is the handoff, not because the agent is
   meaningfully different. The framework's value at this scale
   comes from **spec-as-source-of-truth**, not from agent-identity
   change. Worth naming honestly — it doesn't diminish the value
   but it reframes what's actually happening. (Implication: the
   `claude-plus-agents` variant, where different agents play
   different roles, is closer to the framework's theoretical ideal
   than the `claude-only` variant under high stakes.)

6. **Cost-per-cycle scales with accumulated context.** Each fresh
   session reads AGENTS.md (now 300+ lines), the spec (~400–600
   lines for M specs), stage, brief, and every referenced DEC.
   SPEC-007's verify pass pulled ~3000 lines of context. Not a
   blocker at current scale, but noticeable.

## Improvements I'd prioritize (in order)

1. **Fix `just archive-spec`'s end-of-stage heuristic.** ~15 minute
   chore. Track specs-shipped against stage-backlog `[x]` counts
   rather than `specs/` directory contents.

2. **Formalize "lightweight verify" for chore/XS specs.** Specs
   under ~50 lines of code change with no new behavior (e.g.,
   justfile additions, gitignore fixes, doc-only changes) don't
   need a separate verify session — the design session can
   self-verify. Keep the rigor for S/M/L specs. This would have
   saved a fresh-session cycle on the `just install` chore commit
   here. **Caveat: define the boundary strictly** ("no constraint-
   touching code, no new cobra command, no new storage method") so
   the exception doesn't leak into real specs.

3. **Add a "quick-spec" template variant.** For specs like SPEC-005
   (add shorthand flags), the full spec format produced 378 lines
   for ~15 lines of implementation. A lighter template with just
   Goal / Outputs / Acceptance / Failing Tests / Implementation
   Context would cut design-session time in half for simple
   additions. Don't make it universal — reserve for explicit
   chore/XS.

4. **Track cycle duration in the spec itself.** Build Completion
   gains a one-line `- Build duration: ~7m`. Over 5+ specs you get
   calibration data for future sizing. Requires discipline to log
   it, but low cost.

5. **Collapse Build Completion + Build-phase Reflection.** They
   bleed together in practice (Q1 about friction + the Deviations
   list cover overlapping ground). One tighter section would reduce
   spec-file bloat (which feeds back into the "cost-per-cycle"
   friction above).

6. **Add `<repo-id>` auto-substitution to `decisions/_template.md`.**
   Same bug that was fixed for `just new-stage` / `just new-spec`
   during PROJ-001 setup. DEC templates still hardcode `<repo-id>`
   and require manual edit.

## What I would NOT change

- **Fresh sessions for design → build and build → verify.** Even
  with the caveat above, genuinely catches real defects. Non-
  negotiable at this scale.
- **One PR per spec.** Keeps review tractable.
- **DEC emission with confidence values.** Honest confidence is
  actually honest — DEC-004 at 0.65 vs DEC-006 at 0.85 vs DEC-003
  at 0.95 tracked real uncertainty, and the low-confidence DECs
  have been the ones that earned revisit triggers.
- **Reflection questions in general.** They're ~80% of the way to
  the right shape. The Q1-when-nothing-changed issue is a trim,
  not a structural problem.
- **Spec prescriptiveness.** The more prescriptive specs shipped
  faster and cleaner. The one time the design session got loose
  ("either is acceptable" in SPEC-007), verify caught it
  immediately — discipline earned its keep.

## One meta-observation

The process became noticeably better over 8 specs **because rules
compounded**. That's the intended payoff and it's real. But the
biggest lever isn't framework changes — it's keeping the habits
already in place. Most of what's listed above is ~1-hour fixes to
friction, not fundamental rethinking. The framework works; the only
risk is complacency eroding the rules (skipping verify "just this
once," writing "either is acceptable" again, etc.).

The framework's discipline is self-reinforcing when followed and
self-eroding when not. The claude-only variant makes both directions
easier: there's no team pressure to cut corners, but also no team
pressure to maintain rigor. That's worth naming in the template's
onboarding docs — the discipline is entirely self-imposed.

---

## Addendum after SPEC-008 ship

Two observations worth logging that came up in SPEC-008's verify and
ship cycles:

**1. The `agents.architect` / `agents.implementer` front-matter fields
are template noise under `claude-only`, not signal.** Every spec
file produced under this variant has `architect: claude-opus-4-7`
and `implementer: claude-opus-4-7` because the spec.md template
hardcodes both to the same value. SPEC-008's verify session read
"architect == implementer" as evidence of same-session design+build
contamination and raised a yellow flag — but it was actually just
template default. Real contamination evidence would have to come
from session-specific markers or commit timestamps, not from these
fields. Recommendation: either (a) add a one-line clarifying note
to AGENTS.md §2 (done in `bragfile`'s SPEC-008 ship commit) so
future verifiers don't misread the fields, or (b) auto-populate
them with session-specific identifiers so they actually track which
session did which work (more valuable in the `claude-plus-agents`
variant).

**2. "Build did almost no design work" is not an anti-pattern; it's
the intended framework behavior.** Prescriptive Implementation
Context (hand-fed code blocks, exact test assertions, locked tables
of flag letters / field mappings / output shapes) is what lets
build sessions run clean and fast. SPEC-005, SPEC-006, SPEC-008 all
benefited from this. A verify session might look at a thorough spec
and conclude "the build session just transcribed, that's
contamination-like" — but the framework's value here is that build
is supposed to be mechanical when design is thorough. The friction
the framework protects against is design decisions drifting between
spec-writing and code-writing, not the absence of novel decisions
in build. Worth naming explicitly in the template (perhaps in
`FIRST_SESSION_PROMPTS.md` Prompt 3 or AGENTS.md §12 "During build")
so future verifiers don't over-correct into penalizing thorough
specs.

**Subtle risk worth noting anyway:** a thorough spec + a lazy build
could in theory pass a subtly-wrong implementation because the
failing tests were designed against the code the spec prescribed
(both drafted by the same design session). Fresh-session
verification protects against design-build contamination; it
doesn't inherently protect against spec-tests-and-spec-code being
designed together. Mitigation in practice: the fail-first discipline
(§9) and distinctive-token assertions (§9, SPEC-005 lesson) catch
most of this because they force the test author to reason about the
actual output rather than the prescribed output. But at scale, a
second-pass "test quality" review during verify might be worth
formalizing as its own check.

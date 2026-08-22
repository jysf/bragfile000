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

---

# Learning session — after ~31 specs / 6 stages (2026-06-12)

*Second structured learning session, run during STAGE-007 (PROJ-002).
Mined all 30 shipped-spec reflections (SPEC-001…031), all 7 stage
reflections, the session-log and backlog. Same lens as the round above:
feedback for the template author, not part of `bragfile` itself.*

**Big picture: the reflection→codification loop is working very well.**
Of ~90 distinct findings across the corpus, the large majority were
*already* harvested into AGENTS.md §9/§12, the premise-audit sub-template,
or §4 release ops. The loop the template prescribes (Ship-Q2 "does any
template/constraint/decision need updating?" → WATCH → codify at a stage
close) is the engine doing this, and it is genuinely paying. So this round
is deliberately narrow: (A) lessons that have hit their bar but are
**stranded**, (B) **structural** gaps that keep regenerating the same class
of friction, and (C) **reconciling** the first feedback round.

## Headline: WATCH items lack cross-stage visibility (minor — the loop is not losing lessons)

*Note (corrected after author feedback): an earlier draft of this section
called pending lessons "stranded." That overstated it. The mechanism is
working as designed — see below.*

Lessons accumulate as "WATCH at N=1 / N=2" notes inside the **stage file**
that owns them, and are promoted to AGENTS.md at a stage **close**.
`flag-default-explicitness` is the live example: it reached **N=3**
(SPEC-026 `--format ""`, SPEC-028, SPEC-029) and is correctly recorded as a
WATCH item in STAGE-007 and queued to codify at that stage's close —
SPEC-029's own reflection explicitly says *"do NOT codify mid-stage."*
STAGE-007's `## Stage-Level Reflection` is blank only because **the stage
is still open**, which is the expected, correct state. Nothing is lost.

So this is a *visibility* nicety, not a defect: a WATCH item lives in one
stage file, so there is no single cross-stage view of "what's queued to
codify and at what N." That only bites in narrow cases — a lesson confirmed
by specs spanning two stages, or a long-open stage accumulating several
ready rules.

**Optional recommendation (low priority):** a small roll-up table
(`lesson | confirming cases | bar | owning stage | status`) in
`docs/framework-feedback/`, regenerated at each stage close, would give an
at-a-glance view across stages. Genuinely optional — the per-stage WATCH
notes already do the load-bearing work; this is convenience, not a fix.

## A. Pending codifications (queued — confirm they land at their stage close)

*These are tracked correctly as per-stage WATCH items; the note here is
just to make sure they actually get promoted when their stage closes,
rather than to suggest mid-stage action.*

1. **Flag-default-explicitness (N=3 — queued for STAGE-007 close).** Specs
   for literal-artifact CLI commands repeatedly name a flag's *accepted
   values* but not its *default*; build then guesses (`--format` defaulted
   to `""` vs `"json"`). When STAGE-007 closes, land it: spec template /
   Prompt 2b should require stating each flag's default, not just its
   accepted values.
2. **cobra `SilenceErrors:true` + errBuf-asserting user-error tests
   (N=2, SPEC-030/031).** Non-obvious pattern; the spec's "SilentErr or
   equivalent" hint cost a test-run failure to surface twice. A concrete
   example in Notes (parallel to the existing `"Aborted."` example) removes
   the round-trip.
3. **os-dependent test seams (SPEC-031).** Testing `os.Getwd`/`os.Getenv`
   needs an indirection seam (`var getCwd = os.Getwd`) that the literal
   spec omitted. When a spec touches os-level calls, Notes should call out
   the seam explicitly.
4. **"AC-says-all-N → test each" (SPEC-029).** When an acceptance criterion
   applies to several commands but coverage leans on one shared-helper
   test, verify catches the gap. Candidate spec-template nudge under
   Failing Tests.

## B. Structural gaps (beyond appending more AGENTS.md prose)

1. **Recurring spec-clarity misses cluster on stack specifics.** The same
   *classes* keep recurring — flag defaults, `SilenceErrors`, os seams,
   FTS5 hyphen-as-NOT, bm25 ordering nondeterminism, nullable-column scan
   types. A generic spec template can't encode these, so they re-surface
   at build/verify each time. **Recommend a stack-specific "gotchas
   checklist"** (Go + cobra + sqlite/FTS5), a companion to
   `premise-audit.md`, that the design cycle runs through. Pre-empts the
   single most common recurring friction at design instead of discovering
   it downstream.
2. **AGENTS.md §9/§12 accretion → extract another sub-template.** The
   premise-audit extraction into a referenced sub-doc was the right move;
   repeat it for the test-assertion rule family (split-buffer, monotonic
   tie-break, distinctive-token, heading line-equality, NOT-contains self-
   audit, fail-first run). It's now large enough to live as a referenced
   "test-assertion checklist," leaving §9 a pointer — which also relieves
   the cold-read context cost flagged as friction #6 in the round above.
3. **Deviations-section noise.** Non-deviations keep landing in the
   Deviations list (SPEC-014/015, session-log notes it twice). Sharpen the
   spec template's Deviations prompt to distinguish a *true* deviation from
   a spec-sanctioned choice, so the section stays a real signal.
4. **Role-boundary nit (matters most for `claude-plus-agents`).** A build
   session marked the stage backlog `[x]/shipped` (SPEC-029) — that is the
   ship step's job. One explicit line in Prompt 3: "do not mark the backlog
   shipped; leave the in-progress marker."
5. **trust-but-verify agent reports (N=2, SPEC-023).** Two sub-agent
   sessions reported "pushed" while `origin` was still at the prior SHA.
   For the multi-agent variant, a coordinator should `git ls-remote origin
   <branch>` after any agent reports a push. Carry toward a §13 coordinator-
   discipline line (variant-specific).

## C. Reconcile the first feedback round (2026-04-20)

Several round-1 recommendations appear **un-adopted**; they deserve an
explicit accept-or-reject rather than silent drift:

- **Quick-spec / lightweight-verify for chore/XS — not adopted.** Live
  example *from this very session*: the `just specs-by-stage` name column
  was a ~40-line tooling change that still ran the full feature-branch +
  PR + CI path (PR #45). Either adopt a bounded lightweight lane (with the
  strict boundary round 1 already proposed — "no constraint-touching code,
  no new cobra command, no new storage method") or explicitly reject it and
  record why.
- **Cycle-duration line in Build Completion — not adopted.**
- **Collapse Build Completion + Build-phase reflection — not adopted**
  (still two overlapping sections; round 1's friction #5 stands).
- **archive-spec false "all specs shipped" message** — confirm whether
  fixed. The empty-`<answer>` ship precondition *did* land (good).
- **`decisions/_template.md` `<repo-id>` auto-substitution** — confirm
  whether fixed.

## What's validated and should NOT change

- **literal-artifact-as-spec** is now empirically *format-agnostic* (JSON
  schema, bash hooks, markdown, YAML workflow, shell scripts, completion
  scripts — all transcribed with zero drift). It is the highest-leverage
  pattern in the framework; keep it the default for fixed-shape deliverables.
- **The codification bar** (N=3 same-outcome / N=2 paired-opposing-outcome)
  is a genuinely good meta-rule and the right guard against single-
  observation pattern bloat. Its only weakness is the missing ledger
  (Headline, above) — fix the visibility, not the bar.
- **Fresh-session + spec-as-source-of-truth, one-PR-per-spec, DEC
  confidence values** — all still paying; nothing here argues against them.

## One meta-observation

Round 1's risk was *complacency eroding the rules*. Two stages and ~23
more specs later, the opposite also shows: the rules are so reliably
harvested that the bottleneck has moved from "do we codify lessons?" to
"can we *see* the lessons that are ready to codify?" The discipline
graduated from a habits problem to a bookkeeping one — which is a good
problem to have, and the WATCH-list ledger is the cheap fix for it.

---

# Field note — one build session under the literal-artifact pattern (2026-08-22, ~79 specs / 21 stages)

*Not a mining round. A single build session (SPEC-082, a lint + coverage
gating spec: 2 new files, 19 modified, 19 lint findings fixed, 5 mutation
checks) produced five findings that generalise past `bragfile`, so they are
recorded here while they are cheap. Same lens as the rounds above: feedback
for the template author. Deliberately excluded — feedback on `BRAG.md`'s
`impact` field, which is a product concern of this repo, not a template one.*

**Headline: `literal-artifact-as-spec` held again, and the interesting
failures were all in the machinery *around* the literals, not the literals.**
Every one of the nine embedded artifacts transcribed with zero drift — the
practices page landed at **301 lines / 2,468 words**, matching design's
measured prediction to the word. What went wrong went wrong in how the spec
*reasoned about* its own change, and in how build *confirmed* its own
mutations. That is a healthy place for the failures to be, but three of these
are cheap to fix in the template.

## 1. A spec that moves a derived value must enumerate every guard that caches it

The strongest recommendation here.

The derive-and-diff pattern (a script derives current-state numbers; an
assertion diffs the document against the script) is working well and should
not change. But it creates a specific hazard the template does not warn about:
**a second assertion can also cache the same derived number as a literal**, and
nothing enumerates the cachers.

Concretely: this spec adds one entry to a question register. It correctly
reasoned that this moves two rows of a generated inventory table, and correctly
routed that through the round-trip assertion (`X3`) and the regenerated block.
It missed that a *different* assertion (`Y4`) also pins those same two counts
as literals, for a different and legitimate reason — `Y4` exists to catch the
derivation script itself drifting. Build hit the red, diagnosed it, and
re-pinned with a comment naming the spec. Cost: minutes. But it was found by
running the harness, not by the spec's own *Failing Tests* section, which is
where design is supposed to have predicted it.

**Recommended, cheap:** add to the design-phase checklist — *before writing the
Failing Tests section, grep the harness for every literal occurrence of any
value this spec changes.* One `grep` catches the whole class. This is the
second-order form of a rule already at N=3 in this repo (*when a literal caches
derived output, the derivation outranks the cache*); the first-order rule is
about documents, and this is the same rule pointed at the test harness.

## 2. Mutation-check confirmation must not use `git diff` on files the spec creates

A real gap in the `§12(b)` / mutation-check protocol, and the kind that fails
silently in the safe-looking direction.

The protocol correctly insists the mutant be **confirmed to have mutated**
before its failure is credited — good rule, earned. But the obvious
implementation (`git diff --quiet <file>`) returns **clean for an untracked
file**. This spec's mutation target was `scripts/coverage.sh`, a file the spec
itself creates, so the first run reported *"mutant did not mutate"* while the
mutant was sitting on disk. The detector was wrong, not the mutant.

The failure mode is nasty precisely because it is conservative-looking: it
reports a *missing* mutation, so it reads as a careful guard working, and the
natural next move is to go re-write a mutation that was already correct.

**Recommended:** the protocol should prescribe **content-hash comparison**
(capture a hash before, after, and after-revert), not `git diff`. It is
strictly more general, works identically for tracked and untracked files, and
additionally proves the *revert* was exact — which `git diff` gives you for
free on tracked files and not at all on new ones. This session used it to
confirm the reverted file was byte-identical to the spec's literal.

## 3. "Confirm the mutant mutated" is necessary but not sufficient — confirm it mutated *only* what you intended

Same session, different mutation. A probe file was created to trip a totality
assertion; it carried a plausible-looking `confidence: 0.50` in its front
matter. It tripped the intended assertion — and also moved a *lowest-confidence*
row in the same generated table, turning a second assertion red for an entirely
unrelated reason.

The mutation "passed" in the sense that the target failed. But the run no longer
demonstrated the proposition the design decision actually claims, which was
*this assertion catches what that one cannot*. Re-run with the confidence value
inside the existing range, the second assertion stayed green and the target was
the only failure — which is the real claim.

**Recommended:** extend the existing rule to two clauses — *confirm the mutant
mutated, and confirm it changed exactly one thing.* A probe with side effects
produces a red for the wrong reason, and a red for the wrong reason is
indistinguishable from a red for the right one in a build log. This pairs
naturally with the repo-level lesson **test the claim, not the counterexample**.

## 4. Strengthen `literal-artifact-as-spec` at the *application* step, not the authoring step

The pattern is validated and should stay the default. This is about how build
applies it.

Build extracted every literal from the spec's own fenced blocks **by line
range**, then applied each with an assertion that the target text occurred
**exactly once** before writing. That is mechanically better than retyping, and
it earned its keep immediately: one hunk was first anchored on a range that
resolved to a closing ``` fence — which occurs **seven times** in the target
file. A manual application would have silently corrupted it in seven places.
This repo has already truncated a spec once this way (1,779 lines → 715), so
the failure is not hypothetical.

**Recommended, optional:** the implementer-notes convention could make this
easier by giving each fenced block a stable label, so build can address blocks
by name rather than by line range. Even without that, one line of build
guidance — *never apply a literal without asserting the anchor's match count* —
would carry most of the value. The assertion caught this, not the care.

## 5. Minor: a spec's own self-counts are unguarded prose

This spec's *Outputs* section says "Modified files (**18**)" while its own list
enumerates **19**; its punch-list section is titled "**19** edits" while
specifying **21** replacement blocks (three import-block additions serve 19
findings). Nothing depended on either number and build simply followed the
enumerated list.

Worth one sentence only because it is the exact class this repo has been burned
by five times: *a count produced by hand is a hypothesis*. Specs are documents
that make counted claims about themselves, and unlike the documents this
framework has taught us to guard, **nothing derives a spec's self-counts.** Not
proposing a mechanism — a linter for prose arithmetic is worse than the disease.
Proposing only that spec templates prefer **"the files below"** over **"Modified
files (18)"** where a list is present anyway. Delete the count, keep the list.

## What this session validates and argues *not* to change

- **Separate sessions per cycle.** Build genuinely did not know design's
  reasoning and had to re-derive it from the spec. Every gap above was found
  because the spec had to stand alone, which is the point.
- **`§12(b)` design-time pre-flight — run every literal through its own tool at
  design and record the output.** This is the single highest-value thing in the
  spec. It let build *diff against a prediction* rather than judge a result:
  "19 issues before, 0 after" and "301 lines / 2,468 words" are falsifiable
  claims, and matching them exactly is strong evidence the transcription is
  correct. Every finding above is about machinery this practice does not cover;
  none is an argument against it.
- **Mutation-checking the enforcement, not just asserting it exists.** Three of
  five mutations produced output that differed from naive expectation. Without
  them the gate would have been believed rather than known.

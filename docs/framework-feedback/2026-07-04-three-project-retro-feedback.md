# Spec-driven process — feedback after 3 projects / 9 stages / 42 specs / 2 releases

*Written 2026-07-04 from a read-only cross-project retrospective over the
`bragfile` repo (PROJ-001 MVP shipped, PROJ-002 projects-and-tags shipped,
PROJ-003 agent-native-spine in progress). Feedback is from the agent
(Claude) that played every role across design/build/verify/ship under the
`claude-only` variant of the spec-driven-template. Intended for the template
author; not part of the `bragfile` project itself. This is the second
installment — the first (`process-feedback.md`, 2026-04-20) was written after
8 specs; this one has a much larger sample and the benefit of two shipped
releases and one supersession.*

*Full evidence and citations: `docs/reports/cross-project/2026-07-04-three-project-retrospective.md` (same PR).*

---

## Headline

The template works. Across 3 projects it produced **40/42 specs shipped, every
planned stage shipped in order with zero reordering/cancellation, 25 DECs with
exactly one supersession (which fired on its own pre-declared triggers) and
zero design-lock→ship drift, mean decision confidence 0.823 with no 1.0s.**
Those are the numbers of a process operating as designed, not one being fought.

The single structural finding worth your attention: **the design→build→verify
net is dense on spec-logic correctness and sparse on runtime/operational
behavior.** Every defect that escaped a cycle across all three projects was
operational/runtime, not logic. Two of the recommendations below are about
closing that gap at the *template* level so every project on the template
inherits the fix instead of re-earning it. The rest is validation — things
the template does that are demonstrably paying off, which you should protect.

---

## What's proven at scale (protect these)

1. **The WATCH → codification pipeline with the N=3-same / N=2-paired-opposing
   meta-rule is the template's best feature.** Over 3 projects it codified ~a
   dozen rules and produced **one** supersession and **zero** rule reversals —
   because it deliberately lags codification (premise-audit family took 3 specs;
   §12(b) took 2 across a stage boundary; injectable-os-var took 3 across two
   stages). The paired-opposing-outcome bar in particular is doing real work:
   its canonical instance (SPEC-023 NEGATIVE — skipped pre-flight, paid a
   recovery commit — paired with SPEC-024 POSITIVE — ran it, zero deviations)
   is more convincing than three successes because each outcome independently
   constrains the rule. Keep this exactly as is.

2. **The premise-audit family (§9) is the highest-ROI defect catcher, and it's
   self-improving.** Two of its four cases (count-bump, status-change adjacency)
   were *born from misses* — a test the audit failed to flag — which is why the
   audit-grep cross-check now demands the greps be *executed*, not described.
   A methodology that gets stronger when it fails is rare; the template should
   hold this up as a worked example of how a WATCH lesson should evolve.

3. **PEEL-IF-L, decided-at-design, keeps sizing honest.** Both peels in the
   corpus (SPEC-029→033, SPEC-041→042) were taken at design and *logged*, not
   discovered at build. Net effect: under-estimates surface as a visible peel
   and over-estimates as a logged rescope — sizing self-corrects without silent
   scope creep. The "release cut as its own spec" shape (SPEC-037, SPEC-042) has
   become a stable, repeatable precedent.

4. **§14 confidence discipline with forced candor works.** No DEC sat at 1.0;
   the two sub-0.75 decisions each carried a revisit trigger or a paired open
   question, and the one supersession (DEC-004 0.65 → DEC-015 0.80) fired on
   the exact triggers DEC-004 named for itself. The confidence field is not
   decoration — it gated real design sessions.

5. **Literal-artifact-as-spec compounds rather than degrades.** It held across
   Go fixtures, cobra `Long` strings, README prose, YAML (goreleaser + CI),
   JSON Schema, and shell templates — value grew with artifact count instead of
   fraying with format heterogeneity. This is the highest-leverage design-phase
   investment the template teaches; it generalized cleanly.

---

## The one gap the template should close

**Finding: the process catches spec-logic defects early (design premise audits,
§12(b) pre-flight, TDD, verify punch-lists) but is structurally weak on
runtime/operational behavior — because those defects only exist once the
artifact meets its real host (a shell, a release runner, a package manager, a
plugin loader, a user's timezone).** The escape-stage distribution across 3
projects:

- caught at design/build/verify: all the spec-logic defects.
- **escaped to production/runtime: every single operational defect** — a
  timezone/day-boundary bug in a derived metric (streak read 0), goreleaser
  dual-tag, macOS Gatekeeper, Homebrew tap-trust, a dev binary migrating the
  *production* DB, and a plugin that registered 0 MCP servers despite its
  manifest validating `--strict`.

Two upstream candidates:

### Upstream candidate A — generalize §12(b) to "target the behavioral surface, not the shape validator"

The template already teaches §12(b) ("run the literal through its target tool
before declaring design done"). The corpus shows the rule needs one refinement
that is **project-agnostic** and belongs in the template, not just in bragfile:

> When a spec's literal makes a claim about *runtime behavior* — a component
> registers, a hook fires, a binary resolves on PATH, a server answers — the
> pre-flight must run the literal through the surface that *exercises that
> behavior*, not merely the surface that *validates its shape*. They are
> different checks; neither substitutes for the other.

Canonical opposing pair, ready-made: SPEC-024 ran cobra's actual
`GenBashCompletion` (behavioral) and caught a marker mismatch at design;
SPEC-041 ran `claude plugin validate --strict` (shape-only) but not
`claude plugin details` (registration), so a manifest that validated still
registered zero servers — the defect escaped to build. This is a template-level
lesson because *any* project that emits a manifest/config/registration artifact
can make the same mistake.

### Upstream candidate B — ship a release-spec template with a runtime/operational pre-flight checklist by default

Every §4 release gotcha in the corpus (dual-tag-on-same-commit, code-signing /
Gatekeeper quarantine, package-manager trust gates, dev/prod data isolation)
was **earned in production, then codified after the fact.** They are also
largely *portable across projects* — any tool that ships binaries via a release
runner and a package manager will hit the same class. The template could ship a
release-spec template whose "Notes for the Implementer" already contains a
runtime/operational pre-flight checklist, so template users inherit the lessons
instead of re-discovering them one release at a time. (bragfile's concrete
version of this checklist is in the retro's action register, item R2.)

---

## A subtler methodology lesson: "reserved but not wired"

DEC-024 reserved an `agent:`/`model:` provenance tag namespace for the MCP write
path. The corpus shows **zero** entries using it — not from neglect, but because
the write path only just shipped and nothing populates the namespace yet. The
generalizable lesson: **a decision that reserves a capability for later should
also name the observability check that will confirm adoption once it lands.**
Consider having the template prompt, when a DEC reserves-for-later, for a paired
"how will we know it's actually being used?" line — otherwise reservations
become invisible debt. (This is the kind of thing that only shows up in a
cross-project retro, which is itself a hint — see below.)

---

## Meta: the cross-project retro is a repeatable template artifact

Two process observations about *doing this retrospective*:

1. **"Defect-escape distribution" deserves to be a standard reflection metric.**
   The most valuable single output here — "where do our defects actually get
   caught, and what escapes?" — is not something any per-spec or per-stage
   reflection surfaces; it only appears when you look across projects. The
   template could add a defect-catch-stage tag to ship reflections (design /
   build / verify / ship / escaped) so this distribution is cheap to compute
   later instead of requiring a full mining pass.

2. **The read-only-worktree + parallel-fan-out retro is itself a recipe worth
   templatizing.** Isolation (a throwaway worktree/branch for all output, a
   read-only DB copy, no writes to the primary checkout), one extraction agent
   per project plus one for cross-cutting artifacts plus one for the data
   corpus, then synthesis — that shape is reusable for any multi-project repo on
   the template. It might belong as a template skill/recipe ("cross-project
   retrospective").

---

## Concrete nits to check (low-priority, operator-reported)

- **Brief `status:` resolver is comment-sensitive.** Per the operator's notes,
  a project brief's `status: active` had to be kept comment-free for the
  stage/status resolver to pick it up (a trailing inline comment caused a
  status-drift miss). If the resolver parses YAML front-matter loosely, this is
  a small robustness fix worth confirming.
- **`just archive-spec` false "all specs shipped" message** was reported in the
  first feedback note (2026-04-20). Worth confirming whether it's been
  addressed; it was worked around repeatedly.

---

## Bottom line

Don't push the template to move faster or codify sooner — the deliberate lag and
the confidence/peel/premise-audit discipline are exactly what produced a
one-supersession, zero-drift, 40/42-shipped record across three projects. The
one place to *add* is runtime/operational coverage: generalize §12(b) to target
behavioral surfaces (A), and give the release-spec template a runtime pre-flight
checklist by default (B). Both are portable across every project the template
serves.

# Next session — open PROJ-007, decide its scope and order

Paste this whole file as the opening message of a fresh session, from
`/Users/jyashinsky/PSeven/experiments/bragfile000`.

---

## Where the repo is

Everything is landed. `main` = `235bb52`, clean tree, **zero open PRs**, all six
gates green. Seventeen PRs (#144–#160) merged on 2026-08-13/14.

- **v0.6.0** shipped 2026-08-12 — `brag memory`, the MCP push surface
  (`brag://` resources + `brag_memory`), `brag_list` filter parity, and the
  re-derived capture caps.
- **v0.6.1** shipped 2026-08-13 — seven audit-nit fixes each with a regression
  test, plus `go install …@latest` finally reporting a real version.
- **PROJ-006 closed** 2026-08-14, explicitly as a **scope reduction**, not a
  completion. Read its Project-Level Reflection before assuming it "finished".
- **No project has `status: active`**, so `just status` misleadingly reports
  `PROJ-001-mvp`. That is a resolver artifact and it resolves the moment
  PROJ-007 opens. Do not chase it.

Run `just status` and read
`projects/PROJ-006-agent-native-depth-core/brief.md` §Project-Level Reflection
first. It names what was deferred and why.

---

## TASK 1 — capture what shipped (do this first, it is stale already)

**Nothing from the 2026-08-13/14 session was ever bragged.** The last
`bragfile` entry is #363 from 2026-08-09. Two releases, a project closure and a
live prod-escape discovery went unrecorded.

Per BRAG.md, **draft and get approval — do not run `brag add` unheard.** Four
candidates, roughly in value order:

1. **v0.6.0 — the corpus became something an agent reads before it works.**
   `brag memory` blends recency/match/project by reciprocal-rank fusion and
   trims to a *token budget* rather than a row count; three `brag://` resources
   let an MCP client auto-load it with **no tool call**. Impact should carry the
   measured figure: a 2000-token default yields 25 entries.
2. **The cask→formula upgrade cliff — found by running the pre-flight, not
   reading it.** `brew upgrade` does not cross the boundary; the retired tap
   still pinned a 0.5.1 cask that Homebrew considered current, so any
   pre-v0.5.2 install was frozen with **no signal on either side** —
   `brew outdated` silent, release side healthy. Found on the maintainer's own
   machine. This is the sharpest single finding of the session.
3. **v0.6.1 — seven audit nits, each with a regression test**, plus `go install`
   reporting its version without disabling the DEC-026 dev/prod guard.
4. **PROJ-006 closed as a scope reduction** — one of three pillars fully
   delivered (*consulted*), one partial (*trustworthy*), one never framed
   (*complete*). Worth capturing precisely because it is an honest close rather
   than a victory lap.

**Use the evidence-link convention SPEC-078 just shipped** — `pr:160`,
`pr:158`, etc. Two reasons: it is correct, and it is the first opportunity to
move the adoption baseline off zero.

⚠️ **But flag it honestly in the entry or to the user:** if the same agent that
wrote the instruction is the one following it, that is a weak signal for the
2026-09-14 measurement. Note these entries as instruction-authored so the
number is not later read as organic adoption.

---

## TASK 2 — open PROJ-007 and decide its shape

`projects/PROJ-007-quality-and-portfolio-readiness/brief.md` exists as
`status: proposed`, with **no `stages/` directory yet**. Its own header says:
*"CANDIDATE — proposed, not opened … Treat everything below as scope input, not
a plan."*

Its framing is **portfolio**: bragfile is a portfolio project, so the goal is
making engineering quality *measured, enforced, and legible to an outside
reader*. Three sketched stages (sequence/split at framing):

- lint + coverage in CI (golangci-lint config, `-cover`, badge)
- discipline-surfacing docs (engineering-practices section + godoc pass)
- a time-boxed scale/perf + concurrency baseline harness

**What to actually decide:**

1. **Does it open?** Opening it also fixes the `just status` artifact.
2. **Scope and order.** The brief ranks (1) as highest ROI and time-boxes (3).
   Challenge that ordering rather than inheriting it.
3. **Does story-surface v2 go in PROJ-007 or a new PROJ-008?** This is
   explicitly unresolved. The user's rule: *if PROJ-007's scope is small, pull
   story-surface in; if PROJ-007 already has a lot, open PROJ-008 and park it
   there.* **A fast-follow PROJ-008 is expected and fine** — do not cram.

**A specific argument worth weighing at framing.** This repo's real quality
story is not lint coverage — it is a discipline almost invisible from outside:
mutation-checked tests, mechanical guards replacing remembered ones, decision
records that log their own corrections, reflections that name what was wrong.
Four `test-docs` assertions (W1–W4) were added in three days, each replacing
something a human previously had to remember. **Making that legible may be
worth more than a coverage badge**, and it is the thing a hiring manager cannot
get from any other repo. Consider ordering the docs stage first.

---

## Outstanding elsewhere (none blocking)

| item | state |
|---|---|
| **2026-09-14 measurement** | The one calendar-bound thing. Parked in `projects/PROJ-001-mvp/backlog.md` as a `⏱ DATED` entry. Re-run `brag tags \| grep -cE '^(commit\|pr\|issue):'` on **2026-09-14 OR after +50 entries, whichever is LATER**. Baseline 0. It decides whether STAGE-020's stamping spec is written or dropped. |
| **STAGE-020** | `on_hold` under closed PROJ-006. Deliberate: not shipped (1 of 6 criteria), not cancelled (question is live). |
| **Retired-tap cleanup** | `retired-tap-migration-prompt.md` in the repo root. `tap_migrations.json` + archive `jysf/homebrew-bragfile`. Not a blocker; measured exposure ~zero. |
| **Two audit nits needing decisions** | `MergeTags` position dup (coupled to the open `tag-ordering-projection` question) and `$EDITOR`-with-spaces (`strings.Fields` makes `"code -w"` work and breaks paths with spaces — needs a chosen quoting rule). Written up in STAGE-018. |
| **8 open questions** | `guidance/questions.yaml`, incl. DEC-043's unexamined `k=60` and the low-priority JSON-transport one. |
| **Stray branches** | ~10 merged branches linger on `origin` from earlier sessions. A stale remote-tracking ref briefly muddied the squash investigation. `git fetch --prune && git branch --merged main \| grep -v main \| xargs -r git branch -d`. |
| **SPEC-070** | Deferred to `PROJ-001/backlog.md` with its full design. Note it reserved **DEC-041**, which is unwritten — `decisions/` has a real gap between 040 and 042, and the backlog entry says the gap *is* that item. |

---

## The one lesson to carry in

**The same defect recurred five times in three days, in different costumes: a
claim that sounds true and that no test actually pins.**

1. SPEC-073's coverage sentence — wrong four times, caught by a fresh reviewer
   every time, never by the writer.
2. SPEC-075's punch-list mutation check — green with the coupling severed;
   pinned nothing.
3. The JSON schema's `maxLength` — an unguarded duplicate of a Go constant.
4. README/tutorial version claims — stale through **two** releases.
5. SPEC-078's `H9` — green, *had teeth*, and still pinned the wrong
   proposition: it asserted "no hex-ish token", passing only because the fixture
   used `session-1` instead of a real UUID.

Four are now mechanical (`test-docs` W1–W4). **The generalisation: run the
mutation and watch it fail before trusting the check — and verify the mutant
actually mutated.** One attempted mutation this session silently did nothing
(`$HEAD` inside a single-quoted `jq` program stays literal), which looks
identical to a test with no teeth.

There is also a **gap in the ladder**: `archive-spec` blocks unanswered spec
reflections, `W4` blocks shipped stages with placeholders, but **nothing checks
that a punch-list delta got its own look before the spec advances**. SPEC-075
got that re-verify and it found two more defects; SPEC-078 did not. Worth a small
assertion if someone is in `test-docs` anyway.

---

## Suggested order

1. Draft the brags (Task 1) — get approval, then capture with `pr:` links.
2. Open PROJ-007, frame its stages, decide the story-surface home (Task 2).
3. Everything else is opportunistic.

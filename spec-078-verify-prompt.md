# Prompt — verify SPEC-078 (tell agents to record evidence links)

Run this in a **fresh session**, from the repo root. The branch is
`feat/spec-078-build`, open as PR #158, cycle `verify`.

You are the fresh reviewer. The session that wrote this spec also built it,
which is exactly the situation the verify cycle exists for — and in this stage's
recorded history, **every prose error was caught by a fresh reviewer and none by
the writer** (SPEC-073's coverage sentence was wrong four times; SPEC-075
shipped a mutation-check that pinned nothing; SPEC-077's first W1 draft would
have passed against the defect it targeted).

## Read first

- `projects/PROJ-006-agent-native-depth-core/specs/SPEC-078-tell-agents-to-record-evidence-links.md`
  — especially LD1–LD4 and Build Completion.
- `BRAG.md` §"Evidence links" and its "Which ref to use" subsection.
- The diff: `git diff main...feat/spec-078-build`

## Already mechanically confirmed — do not just re-run these and call it done

All six gates green. All six acceptance criteria checked, including a live one:
`brag add -T "…,pr:158,commit:4ebc7cb"` then `brag list --tag pr:158` returns
the entry, so the hand-typed convention is untouched. W5/W6/H8/H9 were each
verified failing before implementation, and H9 was verified failing against a
correctly-injected hash.

Re-running those is cheap and fine. It is not the job.

## What to actually attack

**1. Is the premise true, or just plausible?** The spec claims zero adoption is
a *missing instruction* rather than rejected friction. That reframing is the
whole basis for LD1 (instruction only, no stamping). Check it yourself: was
`commit:` genuinely absent from every surface an agent reads at capture time, or
did the writer miss a surface? Candidates not examined in the build: the plugin
README, `docs/api-contract.md`, `AGENTS.md`, the `brag_add` input-schema field
descriptions (as opposed to the tool description), `GETTING_STARTED.md`. **If
the instruction did exist somewhere an agent reads, LD1's reasoning collapses
and the spec should be sent back.**

**2. Does the `brag_add` description actually survive contact?** It went from 9
words to a long sentence. Read it as an agent would, in full, alongside the
other four tool descriptions. Is it still a *description*, or has it become a
manual? Does its length crowd the schema? The spec's own Notes flag this risk —
judge whether the implementation respected it.

**3. Is LD3 stated correctly, and is H9 the right assertion?** LD3 forbids the
hook proposing a hash. H9 enforces it with `grep -oE '\b[0-9a-f]{7,40}\b'`.
Consider: does that regex have false positives on ordinary English in the nudge
text (7+ letters drawn only from a–f)? If someone later adds the word
"defaced" or "accede…" to the nudge, does H9 fail spuriously? Decide whether
that is acceptable or should be tightened.

**4. Is the guidance itself correct?** BRAG.md and the four surfaces all tell
agents to prefer `pr:` because squash-merge orphans branch commits. That was
verified once, on one PR (#151 → `4d6d24c` unreachable, `29b6e88` on main).
Confirm it independently. And consider the case the guidance does *not* cover:
what should an agent record in a repo that **merge-commits** rather than
squashes, where the branch hash survives? Is the advice wrong there, or merely
conservative?

**5. Does the spec claim anything the tests do not pin?** This is the stage's
signature failure. For each claim in Build Completion, ask which assertion would
fail if it were false. The claim "nothing about existing entries changes" is
worth attacking specifically.

## Punch-list, then decide

Write findings as a numbered punch-list with a severity for each, and state
plainly whether SPEC-078 is ready to advance to `ship`. If it is, the closing
sequence is:

```
just advance-cycle SPEC-078 ship     # after filling ## Reflection (Ship)
just archive-spec SPEC-078
```

Then STAGE-020's backlog needs its count updated, and — because SPEC-078 is its
only shipped spec while two remain unwritten — a judgement call on whether the
stage closes or stays active. **Do not close STAGE-020 by reflex**: the stage's
Success Criteria describe automatic stamping and a verification surface, neither
of which this spec delivers. Closing it would mark shipped a stage that met one
of its six criteria. The honest options are to rewrite the stage's criteria
around the instruction-first approach, or leave it active pending the
measurement.

## The measurement that decides what happens next

SPEC-078's whole point is LD4. Baseline: **0 evidence tags** at 2026-08-13.

```bash
brag tags | grep -cE '^(commit|pr|issue):'
```

That number decides whether the expensive stamping spec is needed at all. It
cannot be read today — it needs entries captured *after* this ships. Note in the
ship reflection when it should next be checked, and against how many new
entries, so the follow-up is scheduled rather than remembered.

---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-093
  type: chore                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
                                   # Created at SPEC-088 framing (2026-09-15) to
                                   # CLAIM the id in the same edit as the
                                   # sentence that routes work here — which is
                                   # the rule this spec exists to mechanise.
                                   # Carries SPEC-088's verdict (GO); still owes
                                   # its own framing pass for the fork below.
  blocked: false
  priority: medium                 # the failure has cost one renumbering (#207)
                                   # and no lost work; it is chronic, not urgent.
  complexity: M                    # PROVISIONAL. S for the mechanism; the
                                   # backfill branch moves a user-facing table.
                                   # See ## Complexity.

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-09-15

insight:
  confidence: 0.80                 # the measurements are exact; the right fix
                                   # shape is genuinely open, and one branch has
                                   # a cost this file deliberately does not pay.

references:
  decisions: []                    # emits none yet. If design needs a record,
                                   # the next free number is DEC-054 — NOT
                                   # DEC-050, which SPEC-086 has claimed.
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-088                     # split this out at framing, 2026-09-15
    - SPEC-087                     # LD6: an id reserved in prose gets reallocated
    - SPEC-089                     # paid the cost — DEC-050 renumbered at #207
---

# SPEC-093: ids are claimed by files, not by prose

## Context

> **Cycle: frame.** Created at SPEC-088's framing to claim the id with a file,
> in the same edit as the sentence routing work here. **The routing verdict is
> GO**; the fork below is its own framing pass.
>
> Measured 2026-09-15 against `main` at `6dda56a`. Note that `next_id` reads the
> **working tree**, so the SPEC values below already include this file.

This repo knows the rule. It is written in `AGENTS.md`, recorded as a corpus
failure entry, and codified in a design decision — and it has still been broken
**four times**, the most recent inside the work that explains why it fails.

**M-8a — how ids are chosen.** `scripts/_lib.sh:107` `next_id` takes the highest
number among **filenames in the working tree** and adds one. Exactly two callers:
`scripts/new-spec.sh:48` and `scripts/new-stage.sh:30`. **No recipe creates a
DEC**, so every decision number in this repo is typed by hand
(`grep -n 'new-' justfile` lists `new-spec` and `new-stage` only).

**M-8b — ids reserved only in prose, today.**

| Id(s) | Files on disk | Tracked files mentioning them | Reachable by `next_id`? |
|---|---|---|---|
| `DEC-050` | **0** | **7** | no — `next_id DEC` returns `DEC-054` |
| `STAGE-024`, `STAGE-025`, `STAGE-026` | **0** | **7** (union) | no — `next_id STAGE` returns `STAGE-028` |

```
$ git grep -ln 'DEC-050' | wc -l                    # 7
$ git grep -ln -E 'STAGE-02[456]' | wc -l           # 7
$ ls decisions/DEC-050-*.md projects/*/stages/STAGE-02[456]-*.md   # no matches
```

The inventory's one "reserved" number is `DEC-041`'s tombstone — a *file* with
`insight.type: reservation` — not `DEC-050`. So the page's own count of reserved
numbers is **1** while **four** numbers are in fact spoken for in prose.

**M-8c — the two failure modes are opposite, and only one has bitten.** `DEC-050`
and `STAGE-024/025/026` are **holes**: `next_id` skips past them and can never
reissue them, so they are currently harmless-but-invisible. Holes already exist
harmlessly elsewhere — `SPEC-016` and `SPEC-070` have no files and nothing has
ever noticed. The failure that *did* cost something was a **collision**:
SPEC-089's design took `DEC-050` while SPEC-086 had already claimed it in prose,
and had to renumber to `DEC-051`/`DEC-052` in its own PR (#207).

**M-8d — nothing detects a collision.** There is no assertion that two artifacts
claim one id, and `next_id` cannot help: an id claimed on an **unmerged branch**
is invisible to a working tree cut from `main`, which is the exact shape of #207.

## The fork this spec owes

**What is the mechanism, and what happens to the four ids already reserved?**

- **The mechanism.** A `just new-dec` recipe (so DEC numbers derive like SPEC and
  STAGE ones), and/or a `test-docs` assertion that no two tracked artifacts claim
  the same id. Note what an assertion can and cannot see: it reads the working
  tree, so it catches a collision **after** both files exist, not the cross-branch
  race that produced #207. Design should say plainly which failure it closes.
- **The backfill, which is the expensive half.** Creating `STAGE-024/025/026`
  files would move the inventory's `Stages` row **23 → 26** — a user-facing
  table change, which is the same cost that got fork (a) rejected in SPEC-088.
  Creating a `DEC-050` tombstone would move `Decision numbers reserved, not yet
  decided` 1 → 2 **and** collide with SPEC-086, which is about to author that
  number for real. **Neither was done at framing, deliberately: choosing that is
  what this spec is for.**

## Complexity

**M, provisional.** The mechanism alone is **S** — a recipe plus one assertion,
no Go. What makes it M is that the spec cannot avoid deciding the backfill, and
one branch of that decision moves rows in `docs/engineering-practices.md` and
touches SPEC-086's in-flight decision number. If its own framing pass chooses
the no-backfill branch (record the holes, guard the collisions), it holds at S —
and that is the likelier outcome, since M-8c shows holes have never cost
anything and collisions have.

## GO / NO-GO

**GO** at **M provisional**. **Not a blocker for SPEC-086's design**, for a
reason worth stating precisely rather than assuming:

`DEC-050` is a hole, and `next_id` never returns a hole — so SPEC-086 authoring
`DEC-050` is safe **by construction**, with or without this spec. The risk that
actually materialised at #207 was a *second, concurrent* spec taking a
prose-claimed number, and between now and SPEC-086's design there is exactly one
such spec: **SPEC-088, which emits no decision record**, and whose front-matter
records that if it ever needs one the next free number is `DEC-054`. The same
line is in this file and in SPEC-092. That is a sentence in three files rather
than a guard — adequate for one release, and precisely the kind of "note instead
of mechanism" this stage has already measured as insufficient in the long run,
which is why the spec exists at all.

**Not NO-GO**, despite being the least urgent of the four items: the defect is
chronic (four instances, the fourth inside the work explaining the third), the
mechanism is cheap, and "no check detects two artifacts claiming the same id" is
a one-line gap in a repo whose entire harness thesis is that a rule without a
check is a note.

## Out of scope at framing, deliberately

- **Creating `STAGE-024`, `STAGE-025`, `STAGE-026` or a `DEC-050` tombstone.**
  That is the decision this spec owns, not a tidy-up its framing may perform.
- **`STAGE-025`'s subject (`brag lint`) and `STAGE-020`.** Both parked by the
  release plan pending maintainer review; this spec is about the *numbers*, not
  the work they name.

## Prior related work

- **SPEC-087 LD6** — the rule: an id reserved in prose gets reallocated, because
  `next_id` scans filenames. It is why SPEC-088 and SPEC-090 exist as files.
- **SPEC-089 / PR #207** — the renumbering that paid for the rule's absence.
- **Corpus entry 465** (`bragfile/failed`, 2026-09-08) — "Reserved an id in prose
  three times inside the work that explains why that fails", which already names
  this spec's finding: *"nothing derives the next DEC id and nothing detects two
  specs claiming one."*

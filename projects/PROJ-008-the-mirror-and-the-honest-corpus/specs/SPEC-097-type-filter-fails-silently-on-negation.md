---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-097
  type: bug                        # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
                                   # NOT YET FRAMED. Created 2026-09-24 to
                                   # CLAIM the id and give the item a file,
                                   # when SPEC-090's framing moved it out.
  blocked: false
  priority: medium                 # a silent empty result, not data loss.
                                   # Does NOT gate v0.7.0 (maintainer,
                                   # 2026-09-24): parked in STAGE-027 for a
                                   # decision on where it belongs.
  complexity: S                    # PROVISIONAL. Framing re-checks.

project:
  id: PROJ-008
  stage: STAGE-027                 # parked here for decision, not a thematic
                                   # fit: STAGE-027 is the correctable corpus.
repo:
  id: bragfile

agents:
  architect: claude-opus-5-5
  implementer: claude-opus-5-5     # usually same Claude, different session
  created_at: 2026-09-24

references:
  decisions:
    - DEC-050                      # names the seven --type surfaces
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-086                     # measured it as a dependency of its Fork A
    - SPEC-090                     # carried it until its framing moved it out
---

# SPEC-097: `--type` negation fails silently

## Context

> **Cycle: frame, not yet framed.** This file exists so the item has a
> claimed id and a named owner rather than a sentence in a prompt. Nothing
> below is decided. Framing decides GO/NO-GO, the rule and the blast radius.

**The defect.** `--type` is an exact-match inclusion filter on a single
string (`internal/storage/store.go`, the `--type` clause). A user who tries to
*exclude* a type, or to name two, gets an empty result with exit 0 and no
diagnostic:

```
$ brag list --type '!failed'        # exit 0, no rows, empty stderr
```

Measured the same way for `'-failed'`, `'shipped,failed'`, `'!=failed'` and
`'NOT failed'` (STAGE-023, SPEC-086 ship). Only `--type ''` errors.
Re-confirmed 2026-09-24 for `'!failed'` on a binary built from `main`
(`5edc40e`). It reaches all seven `--type` surfaces, plus MCP `brag_list`.

**Why it matters.** Someone filtering failures *out* of a review gets an
empty document, where they expected a message. After STAGE-023, "everything
except failures" is a thing people will reach for.

**The recommended rule, carried from its first routing** (the maintainer,
2026-09-14, `NEXT-SESSION-PROMPT.md` in #213): when a `--type` value contains
`!`, `,`, whitespace or a leading `-` **and** matches zero rows, exit 1 with a
message saying `--type` is an exact match. That is the error-message fix. It
does not add negation.

## History

- **2026-09-08, STAGE-023:** measured, and filed as a `(not yet written,
  bug)` backlog entry.
- **2026-09-14:** the maintainer routed its error message to SPEC-090, with
  the rule above and *"split if it grows past M"*. The stage entry was never
  updated.
- **2026-09-24, SPEC-090 framing:** moved it out. It shares no code with the
  JSON duplicate-key reject, and the two together would pass M.
- **2026-09-24, the maintainer:** give it its own spec, and push it to the
  next stage for a decision. This file.

## What framing must settle

1. **Error message only, or real negation?** The recommended rule only makes
   the silence loud. Negation (`--type '!failed'` meaning "not failed") is a
   feature, and it touches all seven surfaces and MCP.
2. **Where it belongs.** It is parked in STAGE-027 (the correctable corpus,
   targeting v0.8.0) because that is the next stage with a file, not because
   it fits there. Framing proposes a home, and the maintainer decides.
3. **False positives.** A real type containing a comma or a space: does any
   exist in the corpus? Measure on a `sqlite3 .backup` copy, never the live
   file.

## Prior related work

- **STAGE-023**, the `--type` negation backlog entry and *The Fork 3
  finding*.
- **SPEC-086**, whose Fork A needed "everything except failures" and
  measured this as a dependency.
- **SPEC-090**, `## Scope: the --type error message is out`.

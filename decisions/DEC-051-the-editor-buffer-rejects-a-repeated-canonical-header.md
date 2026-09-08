---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-051                        # stable, never reused
  type: decision                     # decision | analysis | recommendation | observation | reservation
  confidence: 0.9                    # small, well-precedented contract tightening —
                                     # the defect and the fix are both mechanical
                                     # and fully covered by the paired tests.
  audience:
    - developer
    - agent

agent:
  id: claude-sonnet-5
  session_id: null

project:
  id: PROJ-008
repo:
  id: bragfile

created_at: 2026-09-07
supersedes: null
superseded_by: null

tags:
  - editor
  - capture-integrity
  - parsing
---

# DEC-051: the editor buffer rejects a repeated canonical header rather than silently keeping the first

## Decision

`editor.Parse` now rejects a buffer that repeats any of the five canonical
headers (`Title`, `Tags`, `Project`, `Type`, `Impact`) with a descriptive
error naming the offending header, instead of silently keeping the first
value and dropping the rest. A repeated **unknown** header is unaffected —
it stays silently ignored, per the pre-existing contract.

## Context

`Parse` reads the DEC-009 buffer with `net/textproto` and pulled each
canonical field via `hdr.Get(key)`. `Get` returns only the first value for a
repeated key; every later value was dropped with no error. This is
undefined behavior that happened to resolve to first-wins — nothing in
DEC-009 specified it, and nothing signalled it to the caller.

Reported from a field audit of a ~450-entry corpus driven under script: a
scripted `$EDITOR` shim that inserts a header line rather than replacing it
produced a buffer with two `Type:` lines. The duplicate happened to carry
the same value, so nothing broke — but the same mechanism on a duplicated
`Title:` or `Impact:` would have discarded real content silently, which is
exactly the corpus-honesty defect PROJ-008 exists to close. Reachable from
both `brag add` (editor mode) and `brag edit`.

## Alternatives Considered

- **Option A: keep first-wins, document it.** What it is: leave `hdr.Get`
  as-is and add a doc comment noting the behavior. Why rejected: does not
  fix the data-loss case; a user or script that produces a duplicate header
  still loses content with no signal.

- **Option B (chosen): reject a duplicated canonical header as a parse
  error.** What it is: after `ReadMIMEHeader`, check `len(hdr[key]) > 1` for
  each of the five canonical keys (iterated as a fixed slice, not by
  ranging `hdr`) and return an error naming the header before any field
  extraction happens. Why selected: turns silent data loss into a surfaced,
  actionable error; the fixed-slice iteration keeps the "unknown headers
  are silently ignored" contract intact by construction, since only the
  five canonical keys are ever checked.

- **Option C: last-wins instead of first-wins.** What it is: prefer the
  last occurrence of a repeated header instead of the first. Why rejected:
  still silently drops a value; swapping which value survives does not fix
  the defect, it just moves it.

## Consequences

- **Positive:** A duplicated canonical header can no longer discard a field
  silently on either `brag add` (editor mode) or `brag edit` — both go
  through the same `Parse`. `brag learn` (SPEC-085) inherits the fix for
  free, since it uses the same parser.
- **Negative:** A buffer that previously "worked" (silently, on first-wins)
  because of a duplicated header now fails with a `UserError` on `edit`, or
  aborts the add on `add` (editor mode). This is the intended behavior
  change — the prior "success" was silent data loss.
- **Neutral:** No change to `Fields`, `Parse`'s signature, or the buffer
  format itself (DEC-009 stands). No database or migration impact.

## Validation

Covered by `TestParse_DuplicateTitleHeaderIsError`,
`TestParse_DuplicateImpactHeaderIsError`,
`TestParse_DuplicateHeaderIsCaseInsensitive`, and the boundary guard
`TestParse_DuplicateUnknownHeaderStillIgnored` (proves the guard's scope is
exactly the five canonical keys). Revisit if a future buffer format change
(e.g. a repeatable header, such as multiple `Tags:` lines meant to
accumulate) needs a header to legitimately repeat — none does today.

## References

- Related specs: SPEC-089, SPEC-085, SPEC-064
- Related decisions: DEC-009 (buffer format), DEC-046 (validate-on-changed;
  this check runs strictly before `capture.ValidateChanged`), DEC-007
  (`ErrUser` sentinel, which is how this surfaces on `edit`)
- External docs: none

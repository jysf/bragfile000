---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-090
  type: bug                        # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
                                   # NOT YET FRAMED. This file was created at
                                   # SPEC-089 ship to CLAIM the id and to give
                                   # the routed item a named owner. SPEC-089's
                                   # own DEC renumbering (DEC-050 → DEC-051/052)
                                   # was the third instance of an id reserved in
                                   # prose rather than by a file; this is the
                                   # cheapest place to stop the fourth.
  blocked: false
  priority: high                   # the silent ingress is the SCRIPTED one —
                                   # the path an agent drives unattended.
  complexity: S                    # PROVISIONAL — framing re-checks. The fork
                                   # (reject vs. warn) and the MCP ingress are
                                   # the two halves that could push this past S.

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-09-08

references:
  decisions:
    - DEC-051                      # the editor half of exactly this rule
    - DEC-007                      # the stdout-is-data contract
  constraints:
    - one-spec-per-pr
    - stdout-is-for-data-stderr-is-for-humans
  related_specs:
    - SPEC-089                     # fixed the editor half; routed this out

# SPEC-090: the `--json` ingress drops a repeated key

## Context

> **Cycle: frame — not yet framed.** This file exists so the id is *claimed*
> rather than *reserved in prose*, and so the item has a **named owner** rather
> than "the next spec that touches it" — the routing failure STAGE-023 has
> already named twice (corpus entry 433, and again at SPEC-087). Everything
> below is transcribed or re-measured evidence. Nothing here is a decision.
> Framing decides GO/NO-GO, the fork, and the blast radius.

SPEC-089 fixed a silent field drop in `editor.Parse`: a buffer repeating any of
the five canonical headers kept the first value and discarded the rest with no
error. The fix reaches all three **editor** ingresses — `internal/cli/edit.go:94`,
`internal/cli/add.go:198`, `internal/cli/learn.go:117` — which each wrap it as
`UserErrorf("invalid buffer: %v", err)` before any `storage.Open`, so rejection
is atomic by construction.

**There is a fourth ingress, and it still drops the field silently.**
`internal/cli/add_json.go` decodes with `encoding/json`, whose documented
behaviour for a repeated object key is **last-wins, with no error**.

Re-measured at SPEC-089 ship (2026-09-08), against a binary built from the
shipped code, on a scratch `--db` under the session scratchpad:

```
$ echo '{"title":"json dup probe","impact":"REAL VALUE","impact":"CLOBBERED"}' \
    | brag --db "$DB" add --json
exit=0
stdout=[1]
stderr=[]
$ sqlite3 "$DB" 'select id, title, impact from entries;'
1|json dup probe|CLOBBERED
```

**The asymmetry is the defect, not the drop on its own.** After SPEC-089 the two
write ingresses to one corpus *disagree* about what a repeated field means:

| ingress | repeated field | signal |
|---|---|---|
| `editor.Parse` (`edit`, `add` editor mode, `learn`) | rejected, nothing written | exit 1, named header on stderr |
| `add --json` | **last value silently stored** | exit 0, empty stderr |

and the silent one is the **scripted** path — the one an agent drives
unattended, and the one the SPEC-089 field report was generated from. This is
the defect PROJ-008 is named for ("the corpus must not quietly hold something
other than what the author wrote"), on the surface least able to notice it.

SPEC-089 scoped this out **correctly**: it framed Bug A as a `Parse` defect, and
its `## Inputs` says so explicitly. The framing that justifies the fix is what
does not stop at `Parse`.

## Why it is not a one-line fix

`encoding/json` has **no duplicate-key hook** — no `DisallowDuplicateFields`
analogue to `DisallowUnknownFields`, and no callback on repeat. Rejecting means
a `json.Decoder`-token pre-pass over the object before the struct decode, which
is real code with its own edge cases (nested objects, arrays of objects, the
`tagsField` custom unmarshaler, the DEC-012 server-owned `json.RawMessage`
fields). That cost is why it needs its own spec rather than an amendment to
SPEC-089.

## The fork framing must settle

**Reject (exit 1, nothing written), or warn (store last, diagnose on stderr)?**

- *Reject* matches DEC-051 and makes the two ingresses agree, which is the whole
  argument. It is also a **breaking change** for any existing script emitting a
  duplicate key today — and unlike the editor case, no interactive human is in
  the loop to read the message and retry.
- *Warn* is non-breaking and still ends the silence, but leaves the corpus
  holding a value the author did not necessarily intend, and puts the signal on
  stderr where a scripted caller is least likely to read it.

Framing should also decide whether the answer is DEC-051 **amended** to cover
every ingress, or a second record. Neither is assumed here.

## Open, and explicitly NOT measured yet

- **The MCP `brag_add` ingress.** SPEC-089 verify recorded "same shape on the
  MCP `brag_add` ingress, which the SDK decodes with the same package." Plausible
  — the SDK unmarshals tool arguments with `encoding/json` — but it was **not
  driven end to end**, and `internal/mcpserver` does not do the decode itself.
  Framing must measure it rather than inherit the claim; it is the difference
  between an S and an M, because the MCP tool has no exit code to carry a
  rejection and would need the error in the tool result envelope.
- Whether any other stdin/JSON ingress exists (`brag import`, if it has one).

## Prior related work

- **SPEC-089** — the editor half. Its `## Verification` §H carries the original
  measurement and the routing; its `## Reflection (Ship)` Q3 names this file.
- **DEC-051** — the rule this would extend, and the precedent for "reject, do
  not guess."
- **DEC-052** — the sibling fix in the same PR (`brag edit` stdout signal). Its
  own unfixed mirror, `brag delete`, is a *separate* STAGE-023 backlog item and
  is **not** owned here.

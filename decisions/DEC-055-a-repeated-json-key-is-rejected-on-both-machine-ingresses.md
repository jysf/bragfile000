---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-055                        # stable, never reused
  type: decision                     # decision | reservation
  confidence: 0.85                   # the reject posture is the maintainer's
                                     # ruling and the mechanics were measured on
                                     # both ingresses, end to end. The soft spot
                                     # is the key-identity rule: it borrows
                                     # encoding/json's fold, which a Go release
                                     # could change (a test fails if it does).
  audience:
    - developer
    - agent

agent:
  id: claude-opus-5-5
  session_id: null

project:
  id: PROJ-008
repo:
  id: bragfile

created_at: 2026-09-27
supersedes: null
superseded_by: null

tags:
  - capture-integrity
  - json
  - mcp
  - ingress
---

# DEC-055: a repeated JSON key is rejected on both machine ingresses

## Decision

`brag add --json` and the MCP `brag_add` tool reject a JSON object whose top
level names the same key twice, and write nothing: the CLI exits 1 with a
`UserError` (DEC-007) and MCP returns a tool error (`isError: true`). Two keys
are the same key when they are equal **after unescaping and under
encoding/json's case fold** (so `"impact"`, `"Impact"` and `"\u0069mpact"` are
one key). Every top-level key counts, whatever field it names and whatever its
values are; a repeat inside a nested value does not. One function,
`capture.CheckRepeatedKeys`, makes the call on the raw bytes before either
ingress decodes them, and its message names the key:
`key "impact" appears more than once`.

## Context

Both machine ingresses decode last-wins and say nothing. `brag add --json`
uses `encoding/json`; the MCP SDK (go-sdk v1.8.0) decodes `brag_add`'s
arguments with `segmentio/encoding/json` after collapsing them into a map for
schema validation. So `{"title":"x","impact":"A","impact":"B"}` stored `B`
with exit 0 on the CLI and a success result on MCP, and one more `"type"` in a
templated payload could turn a win into a `failed` entry (DEC-049) with no
signal. A trailing `null` kept the *first* value on the CLI instead.

The editor ingresses already reject a repeated header (DEC-051). After SPEC-089
the five write ingresses to one corpus disagreed about what a repeated field
means, and the two that stored silently were the two a script or an agent
drives unattended. The maintainer ruled on 2026-09-24: reject, not warn, as a
named breaking change (SPEC-090's fork question).

The two decoders also disagree with each other. `encoding/json` matches a key
to a field case-insensitively, with Unicode folding (`tagſ` fills `tags`); the
SDK matches exactly and its inferred schema rejects any other spelling. And
the SDK validates the *last* value of a repeat, so `{"tags":["a"],"tags":"b"}`
was stored on MCP and rejected on the CLI. SPEC-090's framing measured six of
twelve key-shape variants coming out differently on the two ingresses.

## Alternatives Considered

- **Option A: warn, store the last value.** Rejected by the maintainer. The
  signal would land on CLI stderr and in a successful MCP result, the two
  places an unattended caller does not act on, and the corpus would still hold
  a value its author may not have meant.
- **Option B: amend DEC-051.** Rejected at framing. DEC-051 is about an
  editor buffer read with `net/textproto`; this rule has different mechanics
  (escaped and folded keys, server-owned and provenance fields, the MCP
  envelope, a pre-pass rather than a map-length check) and changes DEC-012's
  schema, not DEC-009's buffer.
- **Option C: compare keys byte for byte after unescaping.** It matches the
  SDK's own rule, and it misses `"impact"` + `"Impact"` on the CLI, where
  `encoding/json` puts both into one field: the CLI keeps clobbering on the
  exact case the rule exists for.
- **Option D: one rule per ingress** (fold on the CLI, exact on MCP). Both
  would reject every repeat each decoder can collapse, but the two ingresses
  would then answer differently on `"impact"` + `"Impact"` for a different
  reason than today, with two rules to keep in step instead of one.
- **Option E: check nested objects too.** No field `brag` stores holds an
  object. A repeat below the top level sits either in a string field, which
  both ingresses already reject for its type, or in a server-owned value that
  DEC-012 choice 4 discards. Rejecting it would refuse input for a value that
  is never stored.
- **Option F: put the MCP check in `handleAdd`.** The raw arguments are still
  visible there, but the SDK has already validated and decoded them. Six
  variants (`"Impact"`, `"IMPACT"`, `"tagſ"`, a trailing `null`, a trailing
  object, a trailing array) would still be rejected, but by the SDK, with a
  message about something else, one of them garbled
  (`type: <invalid reflect.Value> has type "null"`). The check is SDK
  receiving middleware instead, so it runs before validation.
- **Option G (chosen): reject; fold like encoding/json; top level only; one
  shared check before either decoder; the key named.**

## The six choices

1. **Reject, nothing written.** CLI: `UserErrorf("--json input: %v", err)`,
   exit 1, empty stdout, returned before `storage.Open`. MCP: a
   `CallToolResult` with `isError: true` and one text block, the shape
   `brag_add` already uses for an empty title, returned before the tool
   handler runs. The rejection is atomic by construction on both.
2. **Key identity is encoding/json's.** Two keys are one key when their
   unescaped strings are equal under `strings.EqualFold`, which is exactly the
   fold `encoding/json` uses to match a key to a struct field (ASCII
   upper-casing, then the smallest rune in each `unicode.SimpleFold` orbit).
   It is the wider of the two decoders' rules, so no pair either decoder could
   collapse into one field escapes it. `TestCheckRepeatedKeys_FoldsExactlyLikeEncodingJSON`
   checks the detector against the stdlib's own field match, so a Go release
   that changes the fold fails a test instead of reopening the hole.
3. **Top level only** (Option E).
4. **Every top-level key counts.** User-owned, server-owned, provenance and
   unknown keys alike, and a repeat with the same value. This narrows DEC-012
   choice 4 in one case: a server-owned key is still tolerated and ignored,
   but sent twice it is an error, so "a repeated key is an error" has no
   exception to document. A repeated *unknown* key now reports the repeat
   rather than `unknown field` (choice 5); both are rejections.
5. **One check, first.** `capture.CheckRepeatedKeys(raw []byte) error` walks
   the top-level object with a `json.Decoder` token pass, skipping each value
   whole. The CLI calls it on stdin's bytes before its struct decode; MCP calls
   it from receiving middleware on `brag_add`'s raw `arguments`, before the
   SDK's schema validation and decode. So a repeat is reported before anything
   else about the input, on both ingresses. Input that is not a well-formed
   object (an array, a scalar, a syntax error anywhere) gets `nil` and keeps
   the message its decoder already gives it.
6. **The message names the key**, as sent at its second occurrence:
   `key "impact" appears more than once`, and, when the spellings differ,
   `key "Impact" appears more than once (first as "impact"; keys match
   case-insensitively)`. Each ingress adds its existing prefix:
   `--json input: ` on the CLI, `brag_add: ` on MCP.

## Consequences

- **Positive:** all five write ingresses now agree that a repeated field is an
  error (DEC-051 for the editor three, this record for the machine two). The
  two machine ingresses give the same answer, with the same message, on every
  repeated-key variant SPEC-090 measured, where before they disagreed on six
  of twelve. A repeated `type` can no longer reclassify an entry, and a
  repeated `cost`/`tokens`/`session`/`agent`/`model` can no longer rewrite
  DEC-024/DEC-027 provenance.
- **Negative:** a breaking change for any script that emits a repeated key
  today, including one whose repeats carry the same value. It bites only
  hand-built or string-templated JSON: a JSON library cannot emit a repeat,
  and a library that parses one upstream (`jq .`, `json.load`) collapses it
  before brag sees it, where no check can reach. `add --json` now reads stdin
  fully before decoding instead of streaming it, which the DEC-046 field caps
  keep small.
- **Neutral:** the CLI and MCP still disagree where no key repeats: a lone
  `"Impact"` (the CLI accepts it, MCP's schema rejects it) and a lone `"id"`
  (DEC-012 tolerates it, MCP rejects it). So a repeat inside a nested `"id"`
  value is accepted on the CLI and rejected on MCP, by that lone-`"id"`
  difference, not by this rule. The read-only MCP tools are not checked; a
  repeated filter key there changes a read, not the corpus.

## Validation

Right if every repeated-key variant is refused on both ingresses with the same
message and nothing written, and nothing else changes. Pinned by
`TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant` (one table, both ingresses,
the twelve framing variants plus the headline case),
`TestAddCmd_JSON_RepeatOfEveryFieldIsRejected` and
`TestServer_AddRejectsARepeatOfEveryField` (every key, derived from each
ingress's struct tags with a non-vacuity floor), and the five
`TestCheckRepeatedKeys_*` unit tests. Measured at SPEC-090 design against the
real `brag mcp serve`: a rejected write leaves the database file's SHA-256
unchanged on both ingresses.

Revisit if a schema field ever holds an object (choice 3 then needs a nested
rule for it), if the SDK starts rejecting duplicate keys itself, or if a caller
shows a legitimate need to repeat a key.

## References

- Related specs: SPEC-090 (designs this), SPEC-089 (the editor half)
- Related decisions: DEC-051 (the editor instance of the same rule), DEC-012
  (the schema this adds a rule to; see its amendment), DEC-007 (the exit-1
  path), DEC-024 and DEC-027 (the `brag_add` provenance fields it protects),
  DEC-049 (the `failed` type a repeated `type` could flip)
- External: go-sdk v1.8.0 `mcp/server.go` (`AddReceivingMiddleware`,
  `toolForErr`'s validate-then-decode), `encoding/json` `fold.go`

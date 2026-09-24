---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-090
  type: bug                        # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
                                   # FRAMED 2026-09-24 against main c9de864:
                                   # GO at M. Gates v0.7.0, by the maintainer's
                                   # pull-in of 2026-09-24. The fork (reject vs.
                                   # warn) is the maintainer's; framing
                                   # recommends reject. Created at SPEC-089 ship
                                   # (2026-09-08) to CLAIM the id; that record
                                   # is kept below as written.
  blocked: false
  priority: high                   # the silent ingress is the SCRIPTED one —
                                   # the path an agent drives unattended.
  complexity: M                    # re-checked at framing, up from a
                                   # provisional S: two ingresses that decode
                                   # with different libraries and disagree with
                                   # each other, a key-matching rule the CLI
                                   # pre-pass must mirror (case-folding, escaped
                                   # keys), and a decision record. Not L.

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
    - DEC-012                      # the `add --json` schema this would extend
    - DEC-007                      # ErrUser → exit 1: how a CLI rejection
                                   # surfaces. (First written here as "the
                                   # stdout-is-data contract", which is the
                                   # constraint below, not DEC-007.)
  constraints:
    - one-spec-per-pr
    - stdout-is-for-data-stderr-is-for-humans
  related_specs:
    - SPEC-089                     # fixed the editor half; routed this out
---

# SPEC-090: the `--json` ingress drops a repeated key

## Context

> **Cycle: frame — framed 2026-09-24** against `main` at `c9de864`. **GO at
> M. Gates v0.7.0.** Start at *Framing (2026-09-24)* below; it re-measures
> everything in this section and takes precedence wherever the two differ.
> The section from here to *Prior related work* is the id-claiming record
> written at SPEC-089 ship, kept as written.

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

---

## Framing (2026-09-24, `main` at `c9de864`)

**Why now.** The maintainer pulled SPEC-090 into v0.7.0 on 2026-09-24. With
SPEC-095 shipped (#235), it is the only spec that gates the release cut.
`#236` (`chore(ci): run test-docs in CI on both OSes`) was **open, not
merged** when this framing ran, so `just test-docs` still gates locally only.

### How it was measured

Every number below comes from a binary built from `main` at `c9de864`
(`go build -o $SP/brag ./cmd/brag`, go1.26.6), run against a **fresh scratch
`--db` per case** under the session scratchpad. Nothing touched
`~/.bragfile/db.sqlite`. Each payload is a file written with `printf '%s'`
and fed with `<`, never `echo`. The stored row is read back with the
binary's own `list --format json`, not `sqlite3`. (Tags live in a join
table, so a `select tags from entries` probe fails with `no such column`.)

CLI runner, one case:

```sh
# run.sh <payload.json>
DB="$D/$n.sqlite"; rm -f "$DB"
out=$("$B" --db "$DB" add --json < "$f" 2>"$D/$n.err"); rc=$?
[ -f "$DB" ] && row=$("$B" --db "$DB" list --format json \
  | jq -c ".[] | {id,title,description,tags,project,type,impact,created_at}")
```

MCP driver: a Python script spawns `brag --db <scratch> mcp serve`, keeps
stdin open, and sends `initialize`, then `notifications/initialized`, then
`tools/call brag_add`. The `arguments` object is **spliced into the request
as raw bytes** and never passes through a JSON library, because a library
would collapse the duplicate before it reached the wire (see *Who can emit a
duplicate*). It records the `tools/call` result envelope and then reads the
row back the same way. The first try piped a three-line file to
`mcp serve`. It returned nothing, because stdin EOF shut the server down
before it answered. That run is **not** a measurement and is not counted.

Every row below was checked as its own artifact before being compared:
exit code or `isError`, stderr or envelope text, and the stored row (or the
absence of a DB row).

### Re-measurement 1: every field, both ingresses

Payload shape: `{"title":"<id>","<field>":"REAL","<field>":"CLOBBERED"}`
(for `title`: `{"title":"REAL","title":"CLOBBERED"}`).

| field | `add --json` (CLI): stored | CLI signal | `brag_add` (MCP): stored | MCP signal |
|---|---|---|---|---|
| `title` | `CLOBBERED` | exit 0, stdout `1`, stderr empty | `CLOBBERED` | no `isError`; the result echoes the stored entry |
| `description` | `CLOBBERED` | exit 0, stderr empty | `CLOBBERED` | success |
| `tags` (`tagsField` custom unmarshaler) | `clobbered` | exit 0, stderr empty | `clobbered,agent:frame-probe` | success |
| `project` | `clobbered` | exit 0, stderr empty | `clobbered` | success |
| `type` (`shipped` then `failed`) | **`failed`** | exit 0, stderr empty | **`failed`** | success |
| `impact` | `CLOBBERED` | exit 0, stderr empty | `CLOBBERED` | success |
| `id` (DEC-012 server-owned) | nothing from either value; fresh id `1` | exit 0, stderr empty | n/a: `id` is not a `brag_add` field. A single `"id":1` is rejected: `unexpected additional properties ["id"]` | `isError: true` |
| `created_at` (server-owned) | nothing from either value; fresh timestamp | exit 0, stderr empty | n/a (not a field) | n/a |
| `updated_at` (server-owned) | nothing from either value | exit 0, stderr empty | n/a (not a field) | n/a |
| `agent` (MCP provenance) | n/a: `unknown field "agent"`, exit 1 | exit 1 | **`agent:clobbered`** (the clientInfo fallback does not fire) | success |
| `model` | n/a | n/a | **`model:clobbered`** | success |
| `session` | n/a | n/a | **`session:clobbered`** | success |
| `cost` (`1.00` then `9.99`) | n/a | n/a | **`cost:9.99`** | success |
| `tokens` (`100` then `999`) | n/a | n/a | **`tokens:999`** | success |

**Result: the defect reproduces on every user-owned field, on both
ingresses, and on every provenance field MCP stamps.** The spec's original
reproduction (`impact` → `CLOBBERED`, exit 0, empty stderr) holds exactly.
Two rows go beyond "a value was lost":

- A repeated **`type`** silently decides whether an entry is a failure. One
  more `"type"` in a templated payload turns a win into a `failed` entry or
  the reverse. That is the classification DEC-049 and DEC-050 are built on.
- A repeated **`cost`, `tokens`, `session`, `agent` or `model`** silently
  rewrites the provenance DEC-024 and DEC-027 stamp. `cost` is the one field
  bragfile promises it "never estimates", and a duplicate replaces it with no
  signal.

Server-owned duplicates on the CLI **lose nothing**. Both values are thrown
away under DEC-012 choice 4 whether or not the key repeats.

### Re-measurement 2: key shape and value shape change the answer

| case | CLI | MCP |
|---|---|---|
| `"impact":"REAL","Impact":"CLOBBERED"` (case-fold) | **stores `CLOBBERED`**, exit 0 | **rejected**: `validating "arguments": validating root: unexpected additional properties ["Impact"]`, `isError: true`, nothing stored |
| `"IMPACT":"REAL","impact":"CLOBBERED"` | stores `CLOBBERED`, exit 0 | rejected: `unexpected additional properties ["IMPACT"]`, nothing stored |
| `"tags":"real","tagſ":"clobbered"` (U+017F long s) | **stores `clobbered`**, exit 0 | rejected, `additional properties ["tagſ"]` |
| `"impact":"REAL","impact":"CLOBBERED"` (escaped key) | stores `CLOBBERED`, exit 0 | **stores `CLOBBERED`**, success |
| `"impact":"REAL","impact":null` | **stores `REAL`**: first-wins, exit 0, silent | rejected: `validating /properties/impact: type: <invalid reflect.Value> has type …` |
| `"impact":"REAL","impact":""` | stores `""`, exit 0 | stores `""`, success |
| `"impact":{"x":1},"impact":"CLOBBERED"` (nested, then string) | **rejected**: `cannot unmarshal object into Go struct field addJSONInput.impact of type string`, exit 1 | **stores `CLOBBERED`**, success |
| `"impact":"REAL","impact":{"x":1}` (string, then nested) | rejected, exit 1 | rejected: `type: map[x:1] has type "object"` |
| `"tags":["a"],"tags":"b"` (array, then string) | **rejected**: `tags must be a comma-joined string, not an array (per DEC-004)`, exit 1 | **stores `b`**, success |
| `"tags":"a","tags":["b"]` (string, then array) | rejected (the same DEC-004 message), exit 1 | rejected: `type: [b] has type "array", want "string"` |
| `"type":"shipped","type":"shipped"` (same value) | stores `shipped`, exit 0 | stores `shipped`, success |
| `"id":{"a":1,"a":2}` / `"id":[{"a":1,"a":2}]` (duplicate **inside** a server-owned value) | accepted, nothing stored from it, exit 0 | n/a (`id` is rejected outright) |

**Yes, nested and array values change the behaviour, differently on each
ingress.** The CLI decodes into the struct in order. A wrong-typed value in
*either* position records an error, and `Decode` returns it at the end, so
the CLI rejects. The MCP SDK first collapses the arguments into a
`map[string]any` for schema validation (`go-sdk@v1.8.0 mcp/tool.go:98`).
The earlier, invalid value is gone before the validator runs, so MCP
**accepts** an array or object followed by a string. Its errors are all
about the *last* value.

**The key-matching rule differs by ingress, and it matters for design.**
`encoding/json` matches struct fields **case-insensitively, with Unicode
folding** (`ſ` → `s`), and `DisallowUnknownFields` does not flag a case
variant. The MCP SDK decodes with `segmentio/encoding/json` set to
`DontMatchCaseInsensitiveStructFields` (`go-sdk internal/json/json.go`), and
its inferred schema rejects additional properties. So a case-variant
duplicate is **already rejected on MCP and silently clobbers on the CLI**.
A CLI pre-pass that compares keys byte for byte would miss `Impact`/`impact`
and `tags`/`tagſ`. Both ingresses unescape keys, so both need the pre-pass
to compare **decoded** token strings, not raw bytes.

**`null` is a second silent drop, in the other direction.** On the CLI a
trailing `null` is a no-op for a string field, so the *first* value wins,
silently. That is DEC-051's exact first-wins shape, on the JSON ingress. Any
rule keyed on "a repeated key" covers it. A rule keyed on "the last value
differs" would not.

### Re-measurement 3: the MCP `brag_add` ingress, driven end to end

**Driven, not asserted.** It was the real `brag mcp serve` from the `main`
binary, pointed at a scratch DB with `--db` (DEC-026's dev guard is why the
default config resolution was not used). SPEC-089 verify's claim, "same shape
on the MCP `brag_add` ingress", **holds for the plain duplicate and does not
hold in general**. The two ingresses disagree on six of the twelve variant
rows above. The SDK claim it rested on is also wrong in detail: the SDK
does not decode "with the same package". It uses `segmentio/encoding/json`,
case-sensitive, after a map round-trip.

What an MCP caller sees for `{"title":"m01","impact":"REAL","impact":"CLOBBERED"}`:
no `isError`, and `content[0].text` is the stored entry as DEC-011 JSON with
`"impact": "CLOBBERED"`. **The result echoes the clobbered value back.** A
careful agent *could* notice by diffing what it sent against what came back.
Nothing tells it to.

**Where a fix can sit.** A throwaway module outside the repo (go-sdk v1.8.0,
same `go.sum`, `GOPROXY=off`) registered a `brag_add`-shaped typed handler,
and the same raw-byte driver sent it the duplicate. The handler saw:

```
RAW={"title":"t","impact":"REAL","impact":"CLOBBERED"}
DECODED={Title:t Impact:CLOBBERED}
```

So `req.Params.Arguments` still carries the **raw wire bytes with the
duplicate intact**, even though the typed `In` has already decoded last-wins.
A handler-side pre-pass over `req.Params.Arguments` can see the repeat and
reject it. It needs no SDK change and no fork.

**What an MCP rejection looks like.** It is the shape `brag_add` already
uses for every other input error. The handler returns a plain Go `error`;
the SDK's `ToolHandlerFor` wraps it as `CallToolResult{IsError: true,
Content: [TextContent{Text: err.Error()}]}` (`go-sdk@v1.8.0 mcp/server.go:423–433`). The
JSON-RPC response is a *success* whose result is marked as an error. It is
not a JSON-RPC `error` object. Measured on the live handler:

```json
{"content":[{"type":"text","text":"brag_add: title is required and must not be empty"}],"isError":true}
{"content":[{"type":"text","text":"brag_add: cost \"-1\": must be a non-negative decimal (e.g. 0.42)"}],"isError":true}
```

A duplicate-key rejection would be one more of these:
`brag_add: <message naming the key>`, `isError: true`, nothing written. The
throwaway handler's `brag_add: probe rejection` came back in exactly that
envelope. There is no `structuredContent` on an error, and no exit code.

**Not measured:** whether a real MCP client ever forwards a model's
duplicate key verbatim. Claude Code receives tool input as an
already-parsed object from the model API, so a model-written duplicate is
probably collapsed before it reaches brag. The raw-stdio driver here shows
that the **server** accepts it, not that a given client sends it. Design
should not claim more than that.

### Re-measurement 4: other ingresses

By grep over non-test Go (`git grep -n -E 'json\.(Unmarshal|NewDecoder)|\.Decode\(|mcp\.AddTool'`)
and by running `brag --help`:

- **`brag add --json`** (`internal/cli/add_json.go:52–56`): an entry ingress.
  Affected.
- **MCP `brag_add`** (`internal/mcpserver/server.go:88`): an entry ingress.
  Affected. The other four MCP tools are read-only.
- **`brag import` does not exist.** `brag --help` lists no import command,
  and nothing else decodes user JSON into an entry.
- `internal/cli/mcp_install.go:38,44` decodes the **MCP client's config
  file** into `map[string]json.RawMessage`. That is last-wins too, but it is
  not an entry ingress and does not write to the corpus. It is out of scope
  and is named here so nobody re-finds it as a gap.
- The three editor ingresses (`edit`, `add` editor mode, `learn`) go through
  `editor.Parse` and are covered by DEC-051.

So the corpus has **five write ingresses**. After SPEC-089 three reject a
repeated field, and **two, both the machine-driven ones, silently keep one
value.**

### Who can emit a duplicate

A serializer cannot. Measured: `jq -nc '{title:"t",impact:"REAL",impact:"CLOBBERED"}'`
→ `{"title":"t","impact":"CLOBBERED"}`, and Python
`json.dumps({"impact":"REAL","impact":"CLOBBERED"})` → `{"impact": "CLOBBERED"}`.
Both *also* collapse a duplicate they parse (`jq -c . < dup.json`,
`json.load`) last-wins, **upstream of brag**, where no pre-pass can see it.
So a duplicate reaches brag only from **hand-written or string-templated
JSON**: `printf`, heredocs, concatenation. That is the shape of the
SPEC-089 field report's shim, and it is the input a pre-pass can reach.

---

## The fork: reject or warn

**Framing recommends. The maintainer decides.**

### Reject: nothing written; CLI exit 1, MCP `isError: true`

- **Makes all five ingresses agree.** That is DEC-051's rule, and PROJ-008's
  name: the corpus does not quietly hold something other than what the
  author wrote. It also removes the six CLI/MCP disagreements in
  *Re-measurement 2* that involve a repeated key, because both sides reject
  before either decoder's quirks matter.
- **It breaks only input that is already ambiguous.** Per *Who can emit a
  duplicate*, a caller that serializes with a JSON library can never trip
  it. The callers it breaks are exactly the templated ones whose entries are
  being silently rewritten today. That includes same-value duplicates
  (`"type":"shipped","type":"shipped"`), which would now fail. DEC-051
  already made the same call on the editor side: the field report's
  duplicate `Type:` lines *had the same value*, and SPEC-089 rejects them
  anyway.
- **It is a breaking change, and v0.7.0 is where one costs least.**
  `[Unreleased]` already holds **four** breaking JSON changes (`impact`,
  `summary`, `story`, `memory`: `CHANGELOG.md` lines 33, 50, 66, 81), so
  v0.7.0 is a minor whose readers are already told to check their JSON
  consumers. This is an
  argument, not a conclusion. The counter-argument is that it puts one more
  breaking line in front of a user who is already reading several.
- **On MCP it has a native shape** (above): a tool error the calling model
  reads as the result of its call, the same channel it already gets for an
  empty title or a bad `cost`. An agent that sees
  `brag_add: "impact" appears twice` can fix it and retry in the same turn.

### Warn: store the last value, diagnose on stderr, exit 0

- **Non-breaking.** No script that works today stops working.
- **On the CLI the signal lands where a scripted caller does not look.** The
  contract is `id=$(… | brag add --json)`. The caller reads stdout for the id
  and the exit code for success, and both say success. stderr is typically
  unread, or discarded with `2>/dev/null`. That is the reason
  `stdout-is-for-data-stderr-is-for-humans` exists, and DEC-052's argument
  for why `brag edit` needed a stdout signal: a batch driver cannot act on
  what it does not read.
- **On MCP there is no stderr the model sees.** A server's stderr goes to the
  client's log. A warning would have to be a second content block on a
  **success** result (`isError` absent), which a calling model may or may not
  act on. It is a weaker signal than the echo it already gets and ignores.
- **The corpus still holds a value the author may not have meant**, and on
  the CLI's `null` case that is the *first* value, not the last. Warn has to
  say which value it kept, per case.
- It makes the ingresses **disagree on purpose**: the editor rejects and
  JSON warns. That would have to be written down as a deliberate asymmetry
  in the record, not left as a gap.

### Recommendation: reject, on both ingresses, in v0.7.0

The deciding asymmetry is who reads the signal. Warn's signal goes to the
two places an unattended caller, script or agent, is least able to act on:
CLI stderr, and a success result on MCP. Reject's signal goes to the place
both are built to act on: the CLI exit code, and a tool error on MCP. The
break falls only on input that no JSON library can produce and that is
already being stored wrong. **Reject top-level keys only.** A duplicate
inside a server-owned value (`"id":{"a":1,"a":2}`) is thrown away under
DEC-012 and loses nothing. For design to confirm: the CLI also rejects a
repeated server-owned key (`"id":1,"id":2`), even though it loses nothing,
so that "a repeated key is an error" has no exceptions to document.

**The question for the maintainer (yes/no):**

> **Should `brag add --json` and MCP `brag_add` reject a repeated top-level
> key in v0.7.0, with nothing written (CLI exit 1, MCP `isError: true`), as
> a named breaking change in the CHANGELOG?**
>
> *Yes* → reject, as recommended. *No* → warn: store the last value, name
> the key and the kept value on stderr (CLI) or in a second content block
> (MCP), exit 0 / success. The asymmetry with DEC-051 is recorded as
> deliberate.

---

## The record: a new decision record, not an amendment to DEC-051

**Recommended: a new record at design, taking whatever id `next_id DEC`
returns when design writes it.** It is `DEC-055` today: `next_id DEC`
returns `DEC-055`, `decisions/` stops at `DEC-054`, and no ref or branch on
`origin` carries a `DEC-055*` file (`git log --all -- 'decisions/DEC-055*'`
is empty after `git fetch`). **This framing does not claim it.** The fork is
open, so the record's content is not known. Claiming the id means writing
the file, and that is design's job, in the same edit as the text that
names it.

Why not amend DEC-051:

- **DEC-051 is an editor-buffer decision.** It is about `net/textproto`,
  five canonical headers, and a fixed-slice check. The JSON rule has
  different mechanics: case-folded and escaped keys, server-owned fields,
  MCP provenance fields the editor has no header for, the MCP envelope, and
  a pre-pass rather than a map-length check. It also has a different
  consequence: a scripted contract breaks under DEC-012. An amendment would
  roughly double DEC-051 with content that is not about its subject.
- **The schema being changed is DEC-012's.** DEC-012's six choices lock the
  `add --json` ingress. "A repeated key is rejected" is a seventh choice on
  that schema, and on `brag_add`'s mirror of it. The new record should
  **cite DEC-051 as the editor instance** of the same principle and state
  how it bears on DEC-012 choice 4 (server-owned fields). Whether DEC-012
  also gets an amendment line pointing at it is design's call. DEC-012 has
  no `## Amendment` today, and one would move the inventory's amendment row.
- A new record keeps the inventory honest: a decision count plus one, which
  `Y3`/`Z7` absorb with zero hand edits since SPEC-087.

---

## Scope: the `--type` error message is **out**

**Where it came from.** It is not in this file and never was. It first
appears in `NEXT-SESSION-PROMPT.md` at `6dda56a` (#213, 2026-09-15), as
*"`--type` error message (added to SPEC-090 by the user, 2026-09-14)"*. That
entry carries a recommended rule: when a `--type` value contains `!`, `,`,
whitespace or a leading `-` **and** matches zero rows, exit 1. It lists the
sites (seven read commands plus MCP `brag_list`) and says *"Update
STAGE-023's backlog: its unwritten `--type` item is now part of SPEC-090."*
The same prompt's table says **"Split if it grows past M."** SPEC-086's
`### Out of scope` (line 1518) recorded it the same way. The STAGE-023 update
never happened: the stage still carries `--type` negation as its own
`(not yet written, bug)` entry. The current prompt (`0f34fe4`, #233) still
lists it under SPEC-090.

**Why out:**

1. **It trips the maintainer's own split rule.** SPEC-090 alone re-checks at
   **M**. The `--type` work is a separate S: a shared helper across eight
   call sites, its own zero-rows rule, and its own tests. M plus an
   unrelated S is past M.
2. **The two share no code, no ingress and no decision.** One is a write
   ingress's JSON pre-pass. The other is a read-side filter's diagnostic in
   seven CLI commands and `brag_list`. Bundled, a verify failure in either
   holds the other, and v0.7.0 is gated on this spec.
3. **Its design question is its own.** "Only on zero rows, because `type` is
   free-form under DEC-049" is a posture call with its own evidence, and
   none of it overlaps this fork.

It stays where STAGE-023 already has it: the `(not yet written, bug)`
`--type` negation entry, with the 2026-09-14 recommended rule attached.
**Whether that item also gates v0.7.0 is the maintainer's call, and framing
does not make it.** The 2026-09-14 note's argument that it should is
sound: *"v0.7.0 is the release that gives users a reason to exclude
failures."* If it does, it needs its own file and id, claimed by writing the
file.

---

## Out of scope

- Designing the pre-pass, choosing its key-comparison rule, or writing Go.
  That is design and build.
- `brag delete`'s stdout mirror of DEC-052: a separate STAGE-023 item. (The
  2026-09-15 prompt suggested folding it in. This framing does not.)
- The `--type` error message: see above.
- CLI/MCP disagreements that involve **no** repeated key: a lone
  `"Impact"` (CLI accepts, MCP rejects), and a lone `"id"` (CLI tolerates
  under DEC-012, MCP rejects). They are real and they are named here, but
  they are not this defect.
- A duplicate that a JSON library collapsed **upstream** of brag
  (`jq '.' | brag add --json`). No ingress can see it.
- `mcp_install`'s last-wins decode of the client config file.
- SPEC-091, SPEC-092, SPEC-093, SPEC-096, the DEC-014 key-list question,
  and the v0.7.0 release cut itself.

## What design must settle

1. **The comparison rule for "the same key".** The CLI must fold the way
   `encoding/json` matches fields: case-insensitive, with Unicode folding
   (`ſ`). Otherwise `impact`/`Impact` keeps clobbering. MCP already rejects
   case variants, so byte-equal-after-unescape is enough there. One helper
   with the rule as a parameter, or two? Either way, keys are compared as
   **decoded token strings**, never raw bytes (`impact`).
2. **Where the MCP check sits.** Measured feasible in `handleAdd` over
   `req.Params.Arguments`. It runs after the SDK's schema validation, so
   cases the SDK already rejects (`m17`, `m24`) never reach it. That is
   fine, because they are rejected either way.
3. **The key set a test derives from.** SPEC-089's ship lesson applies:
   derive the fields to test from `addJSONInput`'s and `addIn`'s struct tags
   (with a non-vacuity floor), not a hand-typed list. The MCP set is 11
   keys; the CLI set is 9.
4. **The message text**, and whether it names the key as sent or as
   canonical (`Impact` vs `impact`).
5. **Docs and CHANGELOG.** `docs/api-contract.md`'s `brag add --json`
   section (lines 92–108) and its DEC-012 line (1530), and `brag_add` in the
   MCP section (line 1362). The CHANGELOG entry is named as breaking if the answer is
   reject.
6. **Atomicity.** Both checks must run before `storage.Open` or `s.Add`.
   The CLI's `parseAddJSON` already runs before `storage.Open`
   (`add_json.go:95–105`).

## Complexity

**M**, up from a provisional S. It is not S because:

- There are **two ingresses with two decoders** (`encoding/json` and
  `segmentio`), which measurably disagree with each other on six variant
  rows. The fix has to be right for both, and tested on both.
- The CLI pre-pass has to mirror `encoding/json`'s case- and Unicode-folding
  field match. A naive pre-pass passes the headline test and misses
  `Impact` and `tagſ`. That is exactly the "tested two of five keys" trap
  SPEC-089 fell into.
- A new decision record, plus docs and a CHANGELOG breaking line.

It is not L because the MCP half needs **no SDK change**: the raw bytes
reach the handler (measured), and the rejection envelope already exists
(measured). The `--type` error message stays out. With it in, this would be
past M, and the maintainer's own rule says to split.

## GO / NO-GO

**GO, at M.** Nothing blocks design except the fork. Design can start on
the shared mechanics (the pre-pass, the key-derivation test, MCP placement)
before the answer arrives, because both branches need the same detector.
Only the action on detection differs.

**It gates v0.7.0**, by the maintainer's 2026-09-24 pull-in. It is the last
spec before the cut.

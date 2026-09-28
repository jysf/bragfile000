---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-090
  type: bug                        # epic | story | task | bug | chore
  cycle: build                     # frame | design | build | verify | ship
                                   # DESIGNED 2026-09-27 against main 01b10ca
                                   # on the maintainer's reject ruling; DEC-055
                                   # is claimed by the design commit's file.
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
  designed_at: 2026-09-27              # main 01b10ca

references:
  decisions:
    - DEC-055                      # WRITTEN at design: the rule this spec
                                   # implements (and a pointer amendment on
                                   # DEC-012)
    - DEC-051                      # the editor half of exactly this rule
    - DEC-012                      # the `add --json` schema this would extend
    - DEC-007                      # ErrUser → exit 1: how a CLI rejection
                                   # surfaces. (First written here as "the
                                   # stdout-is-data contract", which is the
                                   # constraint below, not DEC-007.)
    - DEC-024                      # brag_add's provenance params, which a
    - DEC-027                      # repeat could rewrite
  constraints:
    - one-spec-per-pr
    - stdout-is-for-data-stderr-is-for-humans
    - no-sql-in-cli-layer          # the detector lives in internal/capture
  related_specs:
    - SPEC-089                     # fixed the editor half; routed this out
---

# SPEC-090: the `--json` ingress drops a repeated key

## Context

> **Cycle: design — designed 2026-09-27** against `main` at `01b10ca`, on the
> maintainer's *reject* ruling (2026-09-24). **DEC-055** is written in the
> design commit. **Start at *What design settled*** near the end. Framing's
> record (from *Framing (2026-09-24)*) and the id-claiming record written at
> SPEC-089 ship (from here to *Prior related work*) are kept as written;
> wherever design re-measured or overturned a claim, the design sections say
> so and take precedence.

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

**Answered, 2026-09-24: yes, reject.** The maintainer, in the orchestration
session: *"if you mean a repeated key in the json, then I think reject is
fine."* It is a repeated key in the JSON object (`{"impact":"A","impact":"B"}`),
so the condition holds. Design proceeds on the reject branch.

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

*Maintainer, 2026-09-24: it does **not** gate v0.7.0. It has its own file
now, **SPEC-097**, parked in STAGE-027 for a decision on where it belongs.*

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

---

## What design settled (2026-09-27, `main` at `01b10ca`)

> Every measurement below comes from two binaries, both go1.26.6:
> **`brag-main`**, built from `01b10ca`, and **`brag-proto`**, built from a
> detached prototype worktree at `01b10ca` carrying every literal in *Notes for
> the Implementer*. Each ran against a fresh scratch `--db` per case under the
> session scratchpad. Nothing touched `~/.bragfile/db.sqlite`, and the MCP
> server was always started with an explicit `--db`. Payloads were files,
> written from Python and fed with `<` or spliced as raw bytes, never passed
> through `echo`.

- **Key identity: `encoding/json`'s fold, on both ingresses.** Two top-level
  keys are the same key when their **unescaped** strings are equal under
  `strings.EqualFold`. That is exactly how `encoding/json` matches a key to a
  struct field (its `fold.go`: ASCII upper-cased, every other rune replaced by
  the smallest rune in its `unicode.SimpleFold` orbit). It is the wider of the
  two decoders' rules. `segmentio`'s exact match is contained in it, so one
  rule covers every pair that either decoder could collapse into one field.
  `"impact"` + `"Impact"` **is** a repeat, and so is `"tags"` + `"tagſ"`. An
  escaped spelling (`"\u0069mpact"`) is compared decoded. A trailing `null`
  is a repeat like any other value. So is a nested or array value followed by
  a string. The rule is pinned to its source:
  `TestCheckRepeatedKeys_FoldsExactlyLikeEncodingJSON` asks `encoding/json`
  itself, through a `reflect.StructOf` field, whether each pair matches one
  field, and requires the detector to agree.
- **Where the detector lives: `internal/capture`, as one function,
  `capture.CheckRepeatedKeys(raw []byte) error`.** The package already holds
  the validation shared by every capture ingress, and both `internal/cli` and
  `internal/mcpserver` import it. It imports nothing but the standard library,
  so `no-sql-in-cli-layer` is untouched. The CLI calls it in `parseAddJSON` on
  stdin's bytes, before the struct decode. MCP calls it from **SDK receiving
  middleware** on `brag_add`'s raw `arguments`, **before the SDK's schema
  validation and decode**. The rule is stated once, the way DEC-050 rule 1
  keeps one predicate.
- **Top level only.** No field that `brag` stores holds an object. `tags` is
  a string (an array is rejected on both ingresses), every `brag_add` field is
  a string, and the only fields that accept an object are DEC-012's three
  server-owned `json.RawMessage` fields on the CLI, whose values are
  discarded. A repeat below the top level therefore sits either in a value
  whose type is already rejected or in one that is never stored. Checking it
  would refuse input for no stored consequence. Measured: `"id":{"a":1,"a":2}`
  is accepted on the CLI after the change, as before.
- **Every top-level key counts**, including a server-owned one sent twice
  (`"id":1,"id":2`), a provenance one (`cost`), an unknown one, and a repeat
  with the same value. That is framing's *"a repeated key is an error, with
  no exceptions to document"*, confirmed.
- **The error text names the key**, as sent at its second occurrence:
  `key "impact" appears more than once`. When the spellings differ it also
  names the first: `key "Impact" appears more than once (first as "impact";
  keys match case-insensitively)`. Each ingress adds the prefix it already
  uses. The CLI prints `brag: user error: --json input: key "impact" appears
  more than once` and exits 1, the shape of its empty-title error. MCP returns
  `brag_add: key "impact" appears more than once` with `isError: true`, the
  shape of its empty-title error.
- **DEC-055 is written in this commit**, as
  `decisions/DEC-055-a-repeated-json-key-is-rejected-on-both-machine-ingresses.md`.
  `next_id DEC` returned `DEC-055` on this branch, `decisions/` stopped at
  `DEC-054`, and `git log --all -- 'decisions/DEC-055*'` was empty after
  `git fetch`. **DEC-012 gains a pointer amendment** in the same commit,
  `## Amendment (2026-09-27, SPEC-090 design)`. A reader of choice 4 (*"server-owned
  fields tolerated-and-ignored"*) would otherwise not learn that
  `"id":1,"id":2` is now an error. The amendment moves the inventory, so this
  commit regenerates the inventory block.
- **Atomicity is measured, not argued.** Each database was seeded with one
  row and hashed before and after the call. A rejected write left the
  SHA-256 unchanged on both ingresses, for all eleven repeated-key rows and
  for the four extra cases below.

---

## The twelve rows, before and after, on both ingresses

`before` is `brag-main`, `after` is `brag-proto`. **CLI** is
`brag --db <scratch> add --json < payload`. **MCP** is the real
`brag --db <scratch> mcp serve`, driven by a Python client that holds stdin
open, sends `initialize`, `notifications/initialized` and `tools/call
brag_add`, and splices the payload in as raw bytes. Each cell was checked as
its own artifact before any comparison: the exit code or `isError`, the
stderr or result text, and the stored row. `after` rows were run on a
database pre-seeded with one row, and **every rejected `after` cell left that
database's SHA-256 unchanged.**

`R` = `brag: user error: --json input: ` on the CLI and `brag_add: ` on MCP.
`K(x)` = `key "x" appears more than once`, and `K(x←y)` = `key "x" appears
more than once (first as "y"; keys match case-insensitively)`.

| # | payload (after `{"title":"rNN",`) | CLI before | MCP before | CLI after | MCP after |
|---|---|---|---|---|---|
| 1 | `"impact":"REAL","Impact":"CLOBBERED"` | stores `CLOBBERED`, exit 0 | rejected: `unexpected additional properties ["Impact"]` | exit 1, `R K(Impact←impact)` | `isError`, `R K(Impact←impact)` |
| 2 | `"IMPACT":"REAL","impact":"CLOBBERED"` | stores `CLOBBERED`, exit 0 | rejected: additional properties `["IMPACT"]` | exit 1, `R K(impact←IMPACT)` | `isError`, `R K(impact←IMPACT)` |
| 3 | `"tags":"real","tagſ":"clobbered"` | stores `clobbered`, exit 0 | rejected: additional properties `["tagſ"]` | exit 1, `R K(tagſ←tags)` | `isError`, `R K(tagſ←tags)` |
| 4 | `"impact":"REAL","\u0069mpact":"CLOBBERED"` | stores `CLOBBERED`, exit 0 | stores `CLOBBERED`, success | exit 1, `R K(impact)` | `isError`, `R K(impact)` |
| 5 | `"impact":"REAL","impact":null` | stores **`REAL`** (first-wins), exit 0 | rejected: `type: <invalid reflect.Value> has type "null", want "string"` | exit 1, `R K(impact)` | `isError`, `R K(impact)` |
| 6 | `"impact":"REAL","impact":""` | stores `""`, exit 0 | stores `""`, success | exit 1, `R K(impact)` | `isError`, `R K(impact)` |
| 7 | `"impact":{"x":1},"impact":"CLOBBERED"` | rejected: `cannot unmarshal object into … addJSONInput.impact` | stores `CLOBBERED`, success | exit 1, `R K(impact)` | `isError`, `R K(impact)` |
| 8 | `"impact":"REAL","impact":{"x":1}` | rejected (the same message) | rejected: `type: map[x:1] has type "object"` | exit 1, `R K(impact)` | `isError`, `R K(impact)` |
| 9 | `"tags":["a"],"tags":"b"` | rejected: `tags must be a comma-joined string, not an array (per DEC-004)` | stores `b,agent:design-probe`, success | exit 1, `R K(tags)` | `isError`, `R K(tags)` |
| 10 | `"tags":"a","tags":["b"]` | rejected (the DEC-004 message) | rejected: `type: [b] has type "array"` | exit 1, `R K(tags)` | `isError`, `R K(tags)` |
| 11 | `"type":"shipped","type":"shipped"` | stores `shipped`, exit 0 | stores `shipped`, success | exit 1, `R K(type)` | `isError`, `R K(type)` |
| 12a | `"id":{"a":1,"a":2}` | accepted, `id` discarded, exit 0 | rejected: additional properties `["id"]` | **unchanged**: accepted, exit 0 | **unchanged**: rejected, `["id"]` |
| 12b | `"id":[{"a":1,"a":2}]` | accepted, exit 0 | rejected: additional properties `["id"]` | **unchanged** | **unchanged** |

**Before: the two ingresses disagreed on six rows (1, 2, 3, 5, 7, 9), which
reproduces framing exactly. After: rows 1 to 11 get the same answer on both
ingresses. That answer is a rejection, nothing written, with the same message
after each ingress's prefix.** Rows 12a and 12b repeat a key only inside a
value, which is not this rule's (LD3). Neither ingress calls them a repeat, and
their outcomes still differ by the lone-`"id"` disagreement framing put out of
scope. They cannot be made to agree without touching that disagreement, so the
parity test pins only that neither ingress reports a repeat for them.

Four more cases, run the same way (after):

| payload | CLI after | MCP after | why it is here |
|---|---|---|---|
| `{"title":"r00","impact":"REAL","impact":"CLOBBERED"}` | exit 1, `R K(impact)` | `isError`, `R K(impact)` | the headline case. Before: `CLOBBERED` stored on both |
| `{"title":"r13","title":""}` | exit 1, `R K(title)` | `isError`, `R K(title)` | the repeat is reported **before** the empty-title check (LD5). Before: the title error on both |
| `{"title":"r14","cost":"1.00","cost":"9.99"}` | exit 1, `R K(cost)` | `isError`, `R K(cost)` | a provenance key. Before: MCP stored `cost:9.99`, and the CLI said `unknown field "cost"` |
| `{"title":"r15","id":1,"id":2}` | exit 1, `R K(id)` | `isError`, `R K(id)` | a server-owned key (LD4). Before: the CLI accepted it, and MCP rejected `["id"]` |

A clean payload, `{"title":"ok","impact":"REAL"}`, stores `REAL` on both,
before and after. That is the positive control. The row, the stamped
`agent:design-probe` tag on MCP and stdout `2` on the seeded CLI database are
all unchanged.

---

## The MCP reject path, driven end to end

The real `brag-proto --db <scratch> mcp serve`, one session, five calls in
order, with the exact bytes on the wire:

```
→ {"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"brag_add","arguments":{"title":"r00","impact":"REAL","impact":"CLOBBERED"}}}
← {"jsonrpc":"2.0","id":2,"result":{"content":[{"type":"text","text":"brag_add: key \"impact\" appears more than once"}],"isError":true}}
→ {"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"brag_list","arguments":{"type":"x","type":"shipped"}}}
← {"jsonrpc":"2.0","id":3,"result":{"content":[{"type":"text","text":"[]"}]}}
→ {"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"brag_add"}}
← {"jsonrpc":"2.0","id":4,"result":{"content":[{"type":"text","text":"validating \"arguments\": validating root: required: missing properties: [\"title\"]"}],"isError":true}}
→ {"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"brag_add","arguments":{"title":"clean","impact":"REAL"}}}
← {"jsonrpc":"2.0","id":5,"result":{"content":[{"type":"text","text":"{\n  \"id\": 1,\n  \"title\": \"clean\", …"}]}}
```

What this shows, measured rather than inferred:

1. **The middleware is registered and fires** on the served binary (R1:
   validation is not registration). The response is a JSON-RPC *success*
   whose result carries `isError: true` and one text block, with no
   `structuredContent`. It is byte-for-byte the envelope `brag_add` already
   returns for an empty title.
2. **It fires only for `brag_add`.** `brag_list` with a repeated `type` still
   runs, last-wins, and returns `[]`. The read tools are out of scope (DEC-055
   *Consequences*).
3. **An absent `arguments` object passes through** to the SDK's own
   `required: missing properties` error. The detector returns `nil` on empty
   input.
4. **A rejected call leaves the session usable**: call 5 succeeds on the same
   connection, and id `1` shows the rejected call wrote nothing.

This ran on go-sdk **v1.8.0** (`go.mod`), where `AddReceivingMiddleware`
wraps the method handler and a `tools/call` request reaches it as
`*mcp.CallToolRequest` (`ServerRequest[*CallToolParamsRaw]`), whose
`Params.Arguments` is the raw `json.RawMessage` off the wire. The SDK's
validate-then-decode step is `toolForErr`'s wrapper, inside the per-tool
handler, so it runs after receiving middleware. **An SDK bump must re-run this
section's calls, not just the Go suite**, because the in-memory Go tests go
through the same middleware path but not the same stdio framing.

---

## §12(b) design-time pre-flight: what the tools actually said

Every literal in *Notes for the Implementer* ran through its real tool:
`go build`, `gofmt -l .` (empty), `go vet ./...` (clean), `just lint`
(**0 issues**), `go test -count=1 ./...` (14 packages `ok`; **1150** passing
tests counting subtests, up from 1119), `./scripts/test-docs.sh` (ALL OK,
**219** `OK:` lines, up from 213), `scripts/inventory.sh`, the served
`brag mcp serve` binary (*The MCP reject path*), and the CLI binary (*The
twelve rows*).

**The literals are `git diff` output from the prototype, base `01b10ca` on
`main`.** They cover 13 files and 19 hunks: 3 new files and 16 hunks in 10
modified files. `internal/mcpserver/server_test.go` is diffed at `-U4` and
the rest at the default `-U3`, because at `-U3` that file's end-of-file
append had a 3-line context occurring twice. **Every modified file's hunk has
an old side that occurs exactly once in its base file.** A script checked
this by rebuilding each hunk's context-plus-`-` text, treating a bare blank
line as empty context, and counting it in `git show 01b10ca:<file>`: 16 of 16
unique. `git apply --check` accepts the whole patch against this design
branch. Applied to a pristine `git archive 01b10ca`, it reproduces all 13
prototype files byte for byte, and `go test ./...` is green there.

**Finding 1: fail-first was run, not predicted.** The four new and changed
test files were copied onto an unmodified `git archive 01b10ca`, with one stub
file, `internal/capture/repeated_key.go`, whose `CheckRepeatedKeys` returns
`nil` (without it the capture package would fail to compile, which is a red
for the wrong reason):

```
internal/capture    FAIL 4: …FoldsExactlyLikeEncodingJSON, …ComparesDecodedKeys,
                            …NamesTheKey, …TopLevelOnly
                    (…LeavesOtherInputToTheDecoder passes: it asserts nil, which the stub returns)
internal/cli        FAIL 2: TestAddCmd_JSON_RepeatOfEveryFieldIsRejected (all 9 subtests),
                            TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant (the 12 repeat rows;
                            nested_object and nested_array pass, as they must)
internal/mcpserver  FAIL 1: TestServer_AddRejectsARepeatOfEveryField
(every other package ok; nothing failed to compile)
```

`test-docs.sh` from the prototype, run on the same archive with its
untouched docs, fails `AG1` to `AG6`, each for its named reason. It also fails
`X3` (the new ids move the inventory) and `AC2` (an archive tracks no files),
both expected.

**Finding 2: M-7 survived the first run, and the guard was wrong, not the
probe.** M-7 deletes the check that the object closes. The first fixture for
that case, `{"a":1,"a"`, breaks *inside* the second value, so the loop
returned `nil` before ever reaching the close check. That left the guard
untested. The fixture `{"a":1,"a":2` was added (the repeat is complete and
the object never closes), and M-7 now fires
`TestCheckRepeatedKeys_LeavesOtherInputToTheDecoder`. The whole matrix was then
re-run on the final prototype, and every other row fired exactly as before.

**Finding 3: M-D1 was a probe too weak, and it is discarded, not credited.**
It removed the link *text* `DEC-055` but left the link's URL, which also
contains `DEC-055`, so `AG1` stayed green, correctly: `AG1` asserts that the
block names the record. M-D1′ removes the whole link, and `AG1` fires.

**Finding 4: the detector first reported a repeat inside a truncated
object.** The first prototype returned the repeat as soon as it saw it, so
`{"a":1,"a"` said *"appears more than once"* where `main` says *"unexpected
EOF"*. That is a second message on input already broken another way. It now
records the first repeat and returns it only after the object closes (LD5).
That change is what M-7 probes.

**Finding 5: two design choices were reversed by a gate, not by argument.**
A named `addToolName` constant failed `W5`, which greps the literal
`Name:        "brag_add"` line. A `(error, int)` helper failed `just lint`
(staticcheck ST1008). Both are in *Rejected alternatives*.

**Finding 6: the escaped-key fixtures were silently de-escaped twice.** See
*Traps*. The Go fixtures now assert their own backslash.

### Mutation matrix: 19 probes, 18 kept, each confirmed by content hash before its gates ran

The probe helper (`probe.py`, in the design session's scratchpad, not
committed: SPEC-093 owns that) **refused any edit whose old text does not occur
exactly once**. It **refused to run the gates until the file's hash had
moved**, then printed the unified diff it had applied, ran
`go test -count=1 ./...` and `./scripts/test-docs.sh`, restored the file from
a backup, and confirmed the hash was back at baseline. **Every target was back
at its baseline after every probe.** Each cell's diff below is exactly what
the helper printed.

**Baselines.** Each baseline is the file at `01b10ca` (on `main`) with this
spec's literals applied, so it is reproduced as
`git show 01b10ca:<path>` + the §-numbered diff for that file. The hashes are
SHA-256, first 12 hex digits:

| File | at `01b10ca` | baseline (`01b10ca` + literals) |
|---|---|---|
| `internal/cli/add_json.go` | `dbd598537f6d` | `333e7c341785` |
| `internal/mcpserver/server.go` | `c5eb1a851848` | `d24ee0b3ee64` |
| `internal/capture/repeated_key.go` | (new file) | `9e37c53c7f54` |
| `docs/api-contract.md` | `3da3b8d8a35f` | `b42abc011274` |
| `docs/for-ai-agents.md` | `bc917e101d07` | `97f3574a1e8f` |
| `BRAG.md` | `d0b4618efea5` | `9383a86b91b9` |
| `CHANGELOG.md` | `33c7ebd14725` | `9428a0d70cef` |

A later cycle that edits one of these files re-derives the baseline from
`01b10ca` and the literal, not from the file on disk. After build merges,
the build merge commit on `main` becomes the durable base.

| # | File | What it breaks | Hash | Fired |
|---|---|---|---|---|
| **M-1** | `internal/cli/add_json.go` | CLI unwired: the detector reaches MCP only | `333e7c341785`→`45ed8d1489a3` | `TestAddCmd_JSON_RepeatOfEveryFieldIsRejected`, `TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant` (12 rows) |
| **M-2** | `internal/mcpserver/server.go` | MCP unwired: the detector reaches the CLI only | `d24ee0b3ee64`→`deaec4e0492e` | `TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant` (12 rows), `TestServer_AddRejectsARepeatOfEveryField` |
| **M-3** | `internal/mcpserver/server.go` | Option F: the MCP check in handleAdd, after the SDK | `d24ee0b3ee64`→`bd41ee6336ce` | `TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant` (6 rows: case_fold, upper_first, unicode_fold, trailing_null, string_then_object, string_then_array) |
| **M-4** | `internal/capture/repeated_key.go` | Option C: keys compared exactly after unescaping | `9e37c53c7f54`→`5e74cc8e9f8f` | `TestCheckRepeatedKeys_FoldsExactlyLikeEncodingJSON`, `TestCheckRepeatedKeys_NamesTheKey`, `TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant` (3 rows: case_fold, upper_first, unicode_fold) |
| **M-5** | `internal/capture/repeated_key.go` | ASCII-only fold: non-ASCII runes pass through unfolded | `9e37c53c7f54`→`34a4057b5f16` | `TestCheckRepeatedKeys_FoldsExactlyLikeEncodingJSON`, `TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant` (1 row: unicode_fold) |
| **M-6** | `internal/capture/repeated_key.go` | the message never names the first spelling | `9e37c53c7f54`→`8007e8b40264` | `TestCheckRepeatedKeys_NamesTheKey`, `TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant` (3 rows: case_fold, upper_first, unicode_fold) |
| **M-7** | `internal/capture/repeated_key.go` | a repeat is reported even in an object that never closes | `9e37c53c7f54`→`dd0b58a2f1f2` | `TestCheckRepeatedKeys_LeavesOtherInputToTheDecoder` |
| **M-8** | `internal/mcpserver/server.go` | registered, never fires: the middleware matches no tool name | `d24ee0b3ee64`→`5e8af181537c` | `TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant` (12 rows), `TestServer_AddRejectsARepeatOfEveryField` |
| **M-9** | `internal/cli/add_json.go` | the CLI check runs after the struct decode, not before | `333e7c341785`→`61328934e3d8` | `TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant` (4 rows: object_then_string, string_then_object, array_then_string, string_then_array) |
| **M-10** | `internal/mcpserver/server.go` | the MCP message loses its brag_add: prefix | `d24ee0b3ee64`→`7b35f9e21af2` | `TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant` (12 rows), `TestServer_AddRejectsARepeatOfEveryField` |
| **M-11** | `internal/capture/repeated_key.go` | a nested value is walked token by token instead of skipped whole | `9e37c53c7f54`→`a78dd6ed671c` | `TestCheckRepeatedKeys_TopLevelOnly`, `TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant` (2 rows: object_then_string, nested_object) |
| ~~M-D1~~ | `docs/api-contract.md` | the add --json bullet loses its record | `b42abc011274`→`9df7a53d98a9` | **none: discarded**, the probe was too weak (Finding 3) |
| **M-D2** | `docs/api-contract.md` | the brag_add bullet's example message | `b42abc011274`→`1212b0395232` | `AG2` |
| **M-D3** | `docs/for-ai-agents.md` | the tool reference's example message | `97f3574a1e8f`→`025b3c8b6410` | `AG3` |
| **M-D4** | `BRAG.md` | the JSON contract bullet's head | `9383a86b91b9`→`141fd4cc14f2` | `AG4` |
| **M-D5** | `CHANGELOG.md` | SPEC-089's stale sentence restored | `9428a0d70cef`→`c41d716a4c01` | `AG6` |
| **M-D6** | `CHANGELOG.md` | the breaking bullet's head reworded | `9428a0d70cef`→`1d5405fbcd76` | `AG5` |
| **M-D7** | `docs/api-contract.md` | the add --json block's opening marker renamed (non-vacuity) | `b42abc011274`→`0c93224b63ad` | `AG1` |
| **M-D1′** | `docs/api-contract.md` | the add --json bullet loses its record, link and all | `b42abc011274`→`d467478eca95` | `AG1` |

**Three probes matter most.** **M-1 and M-2 are the one-ingress shortcut:** a
detector wired into only one ingress. Each is caught by the parity test, and
each also by its own ingress's per-field test. **M-3 is Option F**, the check
in `handleAdd`. It rejects everything the middleware rejects, and the parity
test still catches it on exactly the six rows where the SDK answers first:
`case_fold`, `upper_first`, `unicode_fold`, `trailing_null`,
`string_then_object` and `string_then_array`. **M-8 is R1 in miniature:** the
middleware is registered and never fires. It is caught by the same two tests
as M-2.

The cells, each the diff the helper printed and applied:

**M-1**, `internal/cli/add_json.go`, `333e7c341785` → `45ed8d1489a3`:

````diff
--- a/internal/cli/add_json.go
+++ b/internal/cli/add_json.go
@@ -56,9 +56,6 @@
 	}
 	// Before the decode, which keeps the last of a repeated key and says
 	// nothing (DEC-055). MCP brag_add runs the same check.
-	if err := capture.CheckRepeatedKeys(raw); err != nil {
-		return storage.Entry{}, UserErrorf("--json input: %v", err)
-	}
 	dec := json.NewDecoder(bytes.NewReader(raw))
 	dec.DisallowUnknownFields()

````

**M-2**, `internal/mcpserver/server.go`, `d24ee0b3ee64` → `deaec4e0492e`:

````diff
--- a/internal/mcpserver/server.go
+++ b/internal/mcpserver/server.go
@@ -60,8 +60,6 @@
 	}, handleMemory(s))

 	addResources(srv, s)
-
-	srv.AddReceivingMiddleware(rejectRepeatedKeys)

 	return srv
 }
````

**M-3**, `internal/mcpserver/server.go`, `d24ee0b3ee64` → `bd41ee6336ce`:

````diff
--- a/internal/mcpserver/server.go
+++ b/internal/mcpserver/server.go
@@ -61,8 +61,6 @@

 	addResources(srv, s)

-	srv.AddReceivingMiddleware(rejectRepeatedKeys)
-
 	return srv
 }

@@ -107,6 +105,9 @@
 // entry as a single DEC-011 object.
 func handleAdd(s *storage.Store) func(context.Context, *mcp.CallToolRequest, addIn) (*mcp.CallToolResult, any, error) {
 	return func(_ context.Context, req *mcp.CallToolRequest, in addIn) (*mcp.CallToolResult, any, error) {
+		if err := capture.CheckRepeatedKeys(req.Params.Arguments); err != nil {
+			return nil, nil, fmt.Errorf("brag_add: %w", err)
+		}
 		if strings.TrimSpace(in.Title) == "" {
 			return nil, nil, fmt.Errorf("brag_add: title is required and must not be empty")
 		}
````

**M-4**, `internal/capture/repeated_key.go`, `9e37c53c7f54` → `5e74cc8e9f8f`:

````diff
--- a/internal/capture/repeated_key.go
+++ b/internal/capture/repeated_key.go
@@ -81,7 +81,7 @@
 func foldKey(key string) string {
 	var b []byte
 	for _, r := range key {
-		b = utf8.AppendRune(b, foldRune(r))
+		b = utf8.AppendRune(b, r)
 	}
 	return string(b)
 }
````

**M-5**, `internal/capture/repeated_key.go`, `9e37c53c7f54` → `34a4057b5f16`:

````diff
--- a/internal/capture/repeated_key.go
+++ b/internal/capture/repeated_key.go
@@ -90,7 +90,7 @@
 	if 'a' <= r && r <= 'z' {
 		return r - ('a' - 'A')
 	}
-	if r < utf8.RuneSelf {
+	if true {
 		return r
 	}
 	for {
````

**M-6**, `internal/capture/repeated_key.go`, `9e37c53c7f54` → `8007e8b40264`:

````diff
--- a/internal/capture/repeated_key.go
+++ b/internal/capture/repeated_key.go
@@ -65,7 +65,7 @@
 // repeatedKeyError names the key as sent at its second occurrence, and the
 // first spelling too when the two differ only by case.
 func repeatedKeyError(first, again string) error {
-	if first == again {
+	if true {
 		return fmt.Errorf("key %q appears more than once", again)
 	}
 	return fmt.Errorf("key %q appears more than once (first as %q; keys match case-insensitively)", again, first)
````

**M-7**, `internal/capture/repeated_key.go`, `9e37c53c7f54` → `dd0b58a2f1f2`:

````diff
--- a/internal/capture/repeated_key.go
+++ b/internal/capture/repeated_key.go
@@ -56,9 +56,6 @@
 	}
 	// Report a repeat only in an object that closes: a truncated one is the
 	// decoder's to report, whatever it repeated before it broke.
-	if _, err := dec.Token(); err != nil {
-		return nil
-	}
 	return repeat
 }

````

**M-8**, `internal/mcpserver/server.go`, `d24ee0b3ee64` → `5e8af181537c`:

````diff
--- a/internal/mcpserver/server.go
+++ b/internal/mcpserver/server.go
@@ -73,7 +73,7 @@
 // else. `brag add --json` runs the same check before its own decode.
 func rejectRepeatedKeys(next mcp.MethodHandler) mcp.MethodHandler {
 	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
-		if call, ok := req.(*mcp.CallToolRequest); ok && call.Params.Name == "brag_add" {
+		if call, ok := req.(*mcp.CallToolRequest); ok && call.Params.Name == "brag-add" {
 			if err := capture.CheckRepeatedKeys(call.Params.Arguments); err != nil {
 				res := &mcp.CallToolResult{}
 				res.SetError(fmt.Errorf("brag_add: %w", err))
````

**M-9**, `internal/cli/add_json.go`, `333e7c341785` → `61328934e3d8`:

````diff
--- a/internal/cli/add_json.go
+++ b/internal/cli/add_json.go
@@ -56,9 +56,6 @@
 	}
 	// Before the decode, which keeps the last of a repeated key and says
 	// nothing (DEC-055). MCP brag_add runs the same check.
-	if err := capture.CheckRepeatedKeys(raw); err != nil {
-		return storage.Entry{}, UserErrorf("--json input: %v", err)
-	}
 	dec := json.NewDecoder(bytes.NewReader(raw))
 	dec.DisallowUnknownFields()

@@ -66,6 +63,9 @@
 	if err := dec.Decode(&in); err != nil {
 		return storage.Entry{}, UserErrorf("invalid JSON input: %v", err)
 	}
+	if err := capture.CheckRepeatedKeys(raw); err != nil {
+		return storage.Entry{}, UserErrorf("--json input: %v", err)
+	}
 	// Trailing-garbage detection: a second Decode must hit EOF. Anything
 	// else (a second object, a stray token, or a parse error) means the
 	// stdin contained more than the single JSON value we accept.
````

**M-10**, `internal/mcpserver/server.go`, `d24ee0b3ee64` → `7b35f9e21af2`:

````diff
--- a/internal/mcpserver/server.go
+++ b/internal/mcpserver/server.go
@@ -76,7 +76,7 @@
 		if call, ok := req.(*mcp.CallToolRequest); ok && call.Params.Name == "brag_add" {
 			if err := capture.CheckRepeatedKeys(call.Params.Arguments); err != nil {
 				res := &mcp.CallToolResult{}
-				res.SetError(fmt.Errorf("brag_add: %w", err))
+				res.SetError(err)
 				return res, nil
 			}
 		}
````

**M-11**, `internal/capture/repeated_key.go`, `9e37c53c7f54` → `a78dd6ed671c`:

````diff
--- a/internal/capture/repeated_key.go
+++ b/internal/capture/repeated_key.go
@@ -49,8 +49,7 @@
 		} else if repeat == nil {
 			repeat = repeatedKeyError(first, key)
 		}
-		var skip json.RawMessage
-		if err := dec.Decode(&skip); err != nil {
+		if _, err := dec.Token(); err != nil {
 			return nil
 		}
 	}
````

**M-D1**, `docs/api-contract.md`, `b42abc011274` → `9df7a53d98a9`:

````diff
--- a/docs/api-contract.md
+++ b/docs/api-contract.md
@@ -109,7 +109,7 @@
   (`"impact"`, `"Impact"` and an escaped spelling are one key), and
   a repeated server-owned key counts too. A repeat inside a nested
   value does not. The MCP `brag_add` tool applies the same rule
-  ([DEC-055](../decisions/DEC-055-a-repeated-json-key-is-rejected-on-both-machine-ingresses.md)).
+  ([the record](../decisions/DEC-055-a-repeated-json-key-is-rejected-on-both-machine-ingresses.md)).
 - `tags` stays a comma-joined string per
   [DEC-004](../decisions/DEC-004-tags-comma-joined-for-mvp.md); array
   form (`["a","b"]`) is rejected with an error naming DEC-004.
````

**M-D2**, `docs/api-contract.md`, `b42abc011274` → `1212b0395232`:

````diff
--- a/docs/api-contract.md
+++ b/docs/api-contract.md
@@ -1376,7 +1376,7 @@
   as a single [DEC-011](../decisions/DEC-011-json-output-shape.md) object.
   A missing/empty `title` is a tool error, never a silent insert. So is a
   repeated top-level key, with nothing written:
-  `brag_add: key "impact" appears more than once`. It is checked on the raw
+  `brag_add: key "impact" is repeated`. It is checked on the raw
   arguments, before the SDK validates or decodes them, by the same rule as
   `brag add --json`
   ([DEC-055](../decisions/DEC-055-a-repeated-json-key-is-rejected-on-both-machine-ingresses.md)).
````

**M-D3**, `docs/for-ai-agents.md`, `97f3574a1e8f` → `025b3c8b6410`:

````diff
--- a/docs/for-ai-agents.md
+++ b/docs/for-ai-agents.md
@@ -87,7 +87,7 @@
 never a silent insert. So is a key sent twice, whether as
 `"impact"` and `"impact"` or as `"impact"` and `"Impact"`: nothing is written,
 and the tool error names the key, e.g.
-`brag_add: key "impact" appears more than once`. Send each key once and retry.
+`brag_add: key "impact" is repeated`. Send each key once and retry.
 Unlike the CLI `brag add`, `brag_add` does **not** emit
 a milestone line.

````

**M-D4**, `BRAG.md`, `9383a86b91b9` → `141fd4cc14f2`:

````diff
--- a/BRAG.md
+++ b/BRAG.md
@@ -206,7 +206,7 @@
 - **Unknown keys are strict-rejected.** A typo like `{"titl": "x"}`
   surfaces as `unknown field "titl"` rather than silently losing
   your title.
-- **A repeated key is rejected.** `{"title":"x","impact":"A","impact":"B"}`
+- **Repeated keys are rejected.** `{"title":"x","impact":"A","impact":"B"}`
   writes nothing and exits 1 with `key "impact" appears more than once`,
   where it used to store `B` silently. Keys match case-insensitively, so
   `"Impact"` repeats `"impact"`. A JSON library never emits a repeat, so
````

**M-D5**, `CHANGELOG.md`, `9428a0d70cef` → `c41d716a4c01`:

````diff
--- a/CHANGELOG.md
+++ b/CHANGELOG.md
@@ -119,8 +119,7 @@
   user error (exit 1) with nothing written; the rejection is atomic. Repeating
   an *unknown* header is still ignored, as before. Reaches all three editor
   ingresses: `brag edit`, `brag add` (editor mode) and `brag learn`. The
-  `brag add --json` and MCP `brag_add` ingresses are separate decoders; they
-  get the same rule under *Changed* (DEC-055).
+  `brag add --json` ingress is a separate decoder and is unchanged.
 - **`brag edit` now prints the edited entry's id to stdout on a write**
   ([DEC-052](decisions/DEC-052-brag-edit-emits-the-mutated-id-on-stdout.md)),
   and still prints nothing on a no-op. Both outcomes previously produced empty
````

**M-D6**, `CHANGELOG.md`, `9428a0d70cef` → `1d5405fbcd76`:

````diff
--- a/CHANGELOG.md
+++ b/CHANGELOG.md
@@ -91,8 +91,8 @@
   that means *an array of entry objects* on `brag impact`, `brag review`,
   `brag summary` and `brag wrapped`. Update any `jq .entries` to
   `jq .candidates`.
-- **Breaking: `brag add --json` and the MCP `brag_add` tool reject a
-  repeated key**
+- **Breaking: a repeated key is rejected by `brag add --json` and the MCP
+  `brag_add` tool**
   ([DEC-055](decisions/DEC-055-a-repeated-json-key-is-rejected-on-both-machine-ingresses.md)).
   Both decoders kept the last of a repeated key and said nothing, so
   `{"title":"x","impact":"A","impact":"B"}` stored `B` with exit 0, and a
````

**M-D7**, `docs/api-contract.md`, `b42abc011274` → `0c93224b63ad`:

````diff
--- a/docs/api-contract.md
+++ b/docs/api-contract.md
@@ -86,7 +86,7 @@
 - Editor exec failure (e.g. `:cq` in vim with a modified buffer)
   exits 2 (internal error); the DB is unchanged.

-**STAGE-003 (JSON stdin form):**
+**STAGE-003 (JSON on stdin):**

 ```
 echo '{"title":"shipped"}' | brag add --json
````

**M-D1′**, `docs/api-contract.md`, `b42abc011274` → `d467478eca95`:

````diff
--- a/docs/api-contract.md
+++ b/docs/api-contract.md
@@ -108,8 +108,7 @@
   the same key when they match case-insensitively after unescaping
   (`"impact"`, `"Impact"` and an escaped spelling are one key), and
   a repeated server-owned key counts too. A repeat inside a nested
-  value does not. The MCP `brag_add` tool applies the same rule
-  ([DEC-055](../decisions/DEC-055-a-repeated-json-key-is-rejected-on-both-machine-ingresses.md)).
+  value does not. The MCP `brag_add` tool applies the same rule.
 - `tags` stays a comma-joined string per
   [DEC-004](../decisions/DEC-004-tags-comma-joined-for-mvp.md); array
   form (`["a","b"]`) is rejected with an error naming DEC-004.
````


### Inventory: regenerated and diffed, not predicted

`scripts/inventory.sh` was run on three trees and diffed: `main` at
`01b10ca`, this design branch, and the prototype.

| Row | `main` | this design commit | after build |
|---|---:|---:|---:|
| Decision records | 53 | **54** (DEC-055) | 54 |
| …of those, carrying an explicit `## Amendment` section | 6 | **7** (DEC-012) | 7 |
| Go source files | 70 | 70 | **71** |
| Go test files | 81 | 81 | **83** |
| Go test functions | 862 | 862 | **870** |
| Documentation assertions (distinct ids) | 212 | 212 | **218** |

Every other row is unchanged. The lowest and highest confidence rows hold,
because DEC-055's `0.85` is inside `0.65`–`0.95`. **This design commit moves
two rows, so it carries a regenerated block**, and `X3` is green on it. Build
regenerates the block again with `just inventory`. Passing tests counting
subtests (`go test -v ./... | grep -c -- '--- PASS'`) go from 1119 to
**1150**, and `test-docs` `OK:` lines go from 213 to **219**.

### §9(b): the harness grepped by VALUE

```
$ for v in '| 53 |' '| 54 |' '| 6 |' '| 7 |' '| 862 |' '| 870 |' '| 212 |' '| 218 |' \
           '| 81 |' '| 83 |' '| 70 |' '| 71 |' '1119' '1150' '!=53' '!=862' '!=212'; do
    /usr/bin/grep -n -F -- "$v" scripts/test-docs.sh scripts/inventory.sh \
      | /usr/bin/grep -v '^[^:]*:[0-9]*:[[:space:]]*#'
  done
(no hits for any value)
$ /usr/bin/grep -rn --include='*_test.go' -E '\b(53|862|212|1119)\b' internal cmd
(no hits)
```

**No literal pin exists for any value that moves.** `Y3` and `Z7` derive the
decision count (SPEC-087) and absorb DEC-055 with no edit. Only `X3`'s page
block moves, and it is regenerated.

---

## Locked design decisions

**LD1: reject, nothing written.** Maintainer ruling, 2026-09-24. The CLI
returns `UserErrorf("--json input: %v", err)` from `parseAddJSON`, which runs
before `storage.Open`: exit 1, empty stdout, and no database file is even
created on a fresh path. MCP returns a `*mcp.CallToolResult` built with
`SetError(fmt.Errorf("brag_add: %w", err))` from middleware, so the tool
handler and `s.Add` never run. A **named breaking change** under
`CHANGELOG.md` `[Unreleased]` → `### Changed`.

**LD2: key identity is `encoding/json`'s fold, on both ingresses.** Unescaped
keys, equal under `strings.EqualFold`, implemented as `foldKey`, a transcription
of `encoding/json`'s `appendFoldedName` (go1.26.6 `src/encoding/json/fold.go`).
It is pinned against the stdlib's actual field match, not against a copy of
the rule.

**LD3: top level only.** See *What design settled*. A value is skipped whole
with `dec.Decode(&json.RawMessage{})`, so a key inside it never collides with
a top-level key.

**LD4: every top-level key counts:** user-owned, server-owned, provenance,
unknown, and a same-value repeat. The detector has no key list at all. The
per-ingress tests derive their keys from each struct's tags anyway, so a field
added later is exercised on both wirings with no test edit.

**LD5: one check, first, on the raw bytes.** `capture.CheckRepeatedKeys`
runs before any decoder on both ingresses. A repeat is therefore reported
ahead of every other complaint about the same input: ahead of the CLI's
`unknown field` and type errors, and ahead of MCP's schema errors and its
empty-title check. Input that is **not a well-formed object** (an array, a
scalar, empty input, a syntax error anywhere in the object) gets `nil`, and
the decoder reports it exactly as today. A repeat is reported only in an object
that closes: `{"a":1,"a"` is the decoder's syntax error, not a repeat.

**LD6: MCP's check is receiving middleware, scoped to `brag_add`.**
`srv.AddReceivingMiddleware(rejectRepeatedKeys)` in `mcpserver.New`. It
matches `*mcp.CallToolRequest` with `Params.Name == "brag_add"` and otherwise
calls `next`. The tool name stays a literal in both places. A named constant
would move the `Name:        "brag_add"` line that test-docs `W5` greps for
(tried in the prototype: `W5` failed).

**LD7: the message.** `key %q appears more than once`, and
`key %q appears more than once (first as %q; keys match case-insensitively)`
when the two spellings differ. The key is quoted with `%q`, as sent at its
second occurrence and after unescaping. Prefixes are the existing ones:
`--json input: ` (CLI) and `brag_add: ` (MCP).

**LD8: docs and guards.** `docs/api-contract.md`: a bullet in the `add --json`
block, a sentence in the `brag_add` bullet, and a `DEC-055` line in
*References*. `docs/for-ai-agents.md`: a sentence in `brag_add`.
`BRAG.md`: a bullet in the JSON contract. `docs/brag-entry.schema.json`: a
sentence in the root `description`. `CHANGELOG.md`: the breaking bullet, and
SPEC-089's *"`brag add --json` … is unchanged"* rewritten. `scripts/test-docs.sh`
gains Group `AG` (6 ids), scoped with Group AD's `ad_section` and
`assert_section_names`. **`AGENTS.md` does not change**: no glossary entry
describes `add --json` or `brag_add` input (grep below).

**LD9: records.** DEC-055 and DEC-012's pointer amendment are written in the
design commit. Build only checks them (AC-10).

### Rejected alternatives (build-time)

- **Put the MCP check in `handleAdd`.** It is Option F in DEC-055, and mutant
  M-3 below: six parity rows go red, because the SDK has already rejected or
  decoded those inputs.
- **A byte-for-byte or exact-string comparison.** That is M-4. The CLI keeps
  clobbering on `"Impact"`.
- **Fold with `strings.ToLower`, or ASCII only.** That is M-5. `ſ`, `K`
  (Kelvin) and every other non-ASCII fold are missed on the CLI, whose decoder
  folds them.
- **A named `addToolName` constant.** It breaks `W5` (LD6).
- **Return `(error, int)` from the parity test's CLI helper.** `staticcheck`
  ST1008 fails `just lint`. The error comes last.
- **Stream stdin into the detector instead of reading it whole.** The decoder
  would then need a second read of the same bytes. `io.ReadAll` costs nothing
  measurable at DEC-046's caps. A read error keeps its old classification,
  `UserErrorf("invalid JSON input: %v", err)`.
- **Write the escaped-key fixtures with a `\u` escape in Go or zsh source.**
  Both zsh's `printf` and this session's tool layer decoded `\u0069` to `i`
  before it reached the file, so the escaped-key row silently became a plain
  repeat. The fixtures spell the backslash as `"\x5c"` and assert the escape
  survived (see *Traps*).

---

## Outputs

### New files at build (3)

| File | Contents |
|---|---|
| `internal/capture/repeated_key.go` | `CheckRepeatedKeys`, `repeatedKeyError`, `foldKey`, `foldRune` |
| `internal/capture/repeated_key_test.go` | 5 tests |
| `internal/cli/add_json_parity_test.go` | `package cli_test`: 1 table test over both ingresses, 3 helpers |

### Modified files at build (11)

| File | Change |
|---|---|
| `internal/cli/add_json.go` | read stdin whole; call the detector before the decode (LD1, LD5) |
| `internal/cli/add_json_test.go` | 1 test; imports `reflect` |
| `internal/mcpserver/server.go` | `rejectRepeatedKeys` middleware and its registration (LD6) |
| `internal/mcpserver/server_test.go` | 1 test; imports `strings` |
| `docs/api-contract.md` | 3 hunks (LD8) |
| `docs/for-ai-agents.md` | 1 hunk |
| `BRAG.md` | 1 hunk |
| `docs/brag-entry.schema.json` | 1 hunk (the root `description`, one line) |
| `CHANGELOG.md` | 2 hunks: the breaking bullet, and SPEC-089's sentence rewritten |
| `scripts/test-docs.sh` | Group `AG` |
| `docs/engineering-practices.md` | the inventory block, **regenerated with `just inventory`, never hand-edited** |

### Modified or created at design (this commit)

| File | Change |
|---|---|
| this spec | `cycle: design`, and everything from *What design settled* on |
| `decisions/DEC-055-a-repeated-json-key-is-rejected-on-both-machine-ingresses.md` | **new**: claims the id |
| `decisions/DEC-012-brag-add-json-stdin-schema.md` | `## Amendment (2026-09-27, SPEC-090 design)` |
| `docs/engineering-practices.md` | inventory regenerated: decision records 53 → 54, amendments 6 → 7 |
| `projects/PROJ-008-…/stages/STAGE-023-…md` | SPEC-090's backlog line: `(frame)` → `(design)`, plus a design note |

### The CHANGELOG entry

`[Unreleased]` → `### Changed` gains one bullet, last in the section:

> **Breaking: `brag add --json` and the MCP `brag_add` tool reject a repeated
> key** (DEC-055). Both decoders kept the last of a repeated key and said
> nothing, so `{"title":"x","impact":"A","impact":"B"}` stored `B` with exit
> 0, and a second `"type"` could silently turn a win into a `failed` entry.
> Now nothing is written: the CLI exits 1 with
> `--json input: key "impact" appears more than once` on stderr, and
> `brag_add` returns that message as a tool error. Keys are compared after
> unescaping and case-insensitively, so `"Impact"` repeats `"impact"`, and a
> repeat with the same value is rejected too. Only top-level keys count. A
> payload built with a JSON library cannot contain a repeat; one built by
> string templating (`printf`, a heredoc) can, and now fails instead of
> storing a value its author may not have meant.

`### Fixed`'s SPEC-089 bullet ended *"The `brag add --json` ingress is a
separate decoder and is unchanged."* That sentence is false in the same
release once this ships, so it becomes *"The `brag add --json` and MCP
`brag_add` ingresses are separate decoders; they get the same rule under
*Changed* (DEC-055)."* `AG6` fails if the old sentence returns.

### Premise audit (§9), run at design against the repo

**Case 1: inversion or removal → planned test rewrites. None, by
execution.** The change inverts *"a repeated key is accepted"*. With every
literal applied to `01b10ca`, `go test -count=1 ./...` passed in all 14
packages with no pre-existing test edited. A grep agrees: no test or script
sends a repeated key.

```
$ git grep -n -P '"([A-Za-z_]+)"\s*:[^{}\n]*?,\s*"(?i:\1)"\s*:' -- '*_test.go' '*.sh'
(no hits)
```

The ordering change (LD5) moves no existing assertion either. Every
existing CLI test that asserts `unknown field`, the DEC-004 message or an
invalid-syntax error uses a payload with no repeat.

**Case 2: addition → planned count bumps.** Regenerated and diffed, not
predicted: see *Inventory* below. No literal pin exists on any moving value
(*§9(b)*).

**Case 3: behaviour change → doc references, grepped by value.**

| Grep | Hit | Lands |
|---|---|---|
| `git grep -n -i -E 'add --json\|brag_add' -- ':!projects' ':!.claude' ':!*.go' ':!decisions'` | `docs/api-contract.md:92-110` (the `add --json` block) | **updated** (LD8), `AG1` |
| same | `docs/api-contract.md:1362` (the `brag_add` bullet) | **updated**, `AG2` |
| same | `docs/api-contract.md:1530` (*References*, DEC-012's line) | **a DEC-055 line added after it**. DEC-012's own line stays true |
| same | `docs/for-ai-agents.md:68-89` (`brag_add`'s param table and prose) | **updated**, `AG3` |
| same | `BRAG.md:185-210` (the JSON contract, which tells a caller to validate against the schema first) | **updated**, `AG4`: a schema validator cannot see a repeat |
| same | `docs/brag-entry.schema.json:5` (root `description`) | **updated**: one sentence. JSON Schema cannot express the rule. `internal/capture/schema_pin_test.go` reads only `properties` and stays green |
| same | `CHANGELOG.md:108` (SPEC-089's *"… is unchanged"*) | **rewritten**, `AG6` |
| same | `README.md:122-154`, `docs/tutorial.md:122-130, 295`: examples and the round-trip | **no change**: every example is a single object with no repeat, and no text claims a repeat is accepted |
| same | `docs/data-model.md:216`, `SECURITY.md:39`, `docs/architecture.md:22,26,87`: DEC-012 summary, threat surface, diagrams | **no change**: nothing contradicted |
| same | `scripts/claude-code-post-session.sh`, `examples/brag-slash-command.md`, `plugin/commands/brag.md` | **no change**: the hook builds its payload with `jq`, which cannot emit a repeat |
| same | `CHANGELOG.md` released sections, `docs/reports/`, `docs/research/`, `docs/roadmap/`, `docs/blog/`, `docs/launch/`, `guidance/questions.yaml`, `NEXT-SESSION-PROMPT.md`, `spec-078-verify-prompt.md` | **no change**: dated records |
| `/usr/bin/grep -n -i 'add --json\|brag_add\|stdin json' AGENTS.md` | **no hits** | **`AGENTS.md` does not change** |
| `git grep -n 'mcp.AddTool\|AddReceivingMiddleware'` (non-test) | `internal/mcpserver/server.go`, `resources.go` | the only server. No other middleware exists to order against |

---

## Acceptance Criteria

Run against a binary built from the build branch, `B`, with a scratch store.
Payloads are written to files by a script, never with `echo`. First check each
artifact's shape: stderr is one line starting `brag: `, and a stored row is
read back with `$B --db "$T/db.sqlite" list --format json` and parses with
`jq -e`.

- [ ] **AC-1: the CLI rejects a repeat, atomically.** Seed one row, hash the
  database, and run `$B --db "$T/db.sqlite" add --json < dup.json` with
  `{"title":"t","impact":"A","impact":"B"}`. Exit **1**, stdout empty,
  stderr exactly
  `brag: user error: --json input: key "impact" appears more than once`, and
  the database hash is **unchanged**.
- [ ] **AC-2: the case-fold spelling.** `{"title":"t","impact":"A","Impact":"B"}`
  gives exit 1 and stderr ending
  `key "Impact" appears more than once (first as "impact"; keys match case-insensitively)`.
- [ ] **AC-3: MCP rejects a repeat, on the real server.** Drive
  `$B --db "$T/m.sqlite" mcp serve` with a client that holds stdin open and
  sends the AC-1 payload as `brag_add`'s raw `arguments`. The result is
  `{"content":[{"type":"text","text":"brag_add: key \"impact\" appears more than once"}],"isError":true}`,
  and the database hash is unchanged. A following clean `brag_add` on the same
  session succeeds.
- [ ] **AC-4: the twelve rows.** All thirteen payloads of *The twelve rows*
  (12a and 12b separately) give the `after` columns on both ingresses.
- [ ] **AC-5: unchanged where no key repeats.** `{"title":"ok","impact":"REAL"}`
  stores `REAL` on both ingresses. `{"titl":"x"}` still says
  `unknown field "titl"`. `{"title":`, `[{"title":"x"}]` and two concatenated
  objects give the same stderr as `main`, compared byte for byte after each
  side is checked non-empty.
- [ ] **AC-6: the Go suite.** The 8 new tests in *Failing Tests* exist under
  those names and pass. No pre-existing test is edited, except the two files
  that gain one appended test and one import each.
- [ ] **AC-7: docs.** `./scripts/test-docs.sh` prints `OK:` for `AG1`–`AG6`
  and `ALL OK`. The inventory block equals `just inventory` (`X3`).
- [ ] **AC-8: all five gates:** `just test`, `just test-docs`, `just lint`,
  `gofmt -l .` (empty) and `go vet ./...`.
- [ ] **AC-9: the matrix.** The 18 kept probes, re-run from their stated
  diffs (*Notes for the Implementer*), fire what the matrix says.
- [ ] **AC-10: records.** DEC-055 exists as written in the design commit.
  DEC-012's text above `## Amendment (2026-09-27, SPEC-090 design)` is
  byte-identical to `01b10ca`.

---

## Failing Tests

Written at build first, and observed failing against `01b10ca` with a stub
`CheckRepeatedKeys` that returns `nil` (design ran this: *Finding 1*).

### New: `internal/capture/repeated_key_test.go` (5)

- `TestCheckRepeatedKeys_FoldsExactlyLikeEncodingJSON` (LD2, LOAD-BEARING):
  11 key pairs. `encoding/json` itself decides whether each pair fills one
  field, through a `reflect.StructOf` struct, and the detector must agree. A
  non-vacuity floor requires at least 6 matches and at least 1 non-match.
- `TestCheckRepeatedKeys_ComparesDecodedKeys` (LD2): `"impact"` + its escaped
  spelling. The fixture asserts its own backslash survived.
- `TestCheckRepeatedKeys_NamesTheKey` (LD7): 5 payloads, exact messages.
- `TestCheckRepeatedKeys_TopLevelOnly` (LD3): 4 nested shapes are `nil`, and a
  repeat after nested values is still found.
- `TestCheckRepeatedKeys_LeavesOtherInputToTheDecoder` (LD5): 6 malformed or
  non-object inputs are `nil`. **Passes on the stub by design**, since it
  asserts `nil`. M-7 kills it.

### New: `internal/cli/add_json_test.go` (1)

- `TestAddCmd_JSON_RepeatOfEveryFieldIsRejected` (LD1, LD4): every
  `addJSONInput` json key (floor 9), repeated. `ErrUser`, the exact message,
  empty stdout, 0 rows.

### New: `internal/cli/add_json_parity_test.go` (1)

- `TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant` (LD1–LD7, LOAD-BEARING):
  the twelve framing rows plus the headline, through the real `add` cobra
  command and a real `mcpserver.New` over in-memory transports. Arguments are
  sent as `json.RawMessage`, which design confirmed reaches the server with the
  repeat intact. Rows 1–11 and the headline must give exactly
  `user error: --json input: <msg>` and `brag_add: <msg>`, with nothing stored
  on either side. Rows 12a/12b must not be called a repeat by either ingress.

### New: `internal/mcpserver/server_test.go` (1)

- `TestServer_AddRejectsARepeatOfEveryField` (LD1, LD4, LD6): every `addIn`
  json key (floor 11), repeated, on one session. Each result is `isError` with
  the exact message, and the store holds 0 rows at the end.

### New: `scripts/test-docs.sh`, Group `AG` (6 ids)

- `AG1`: the contract's `**STAGE-003 (JSON stdin form):**` block names
  `appears more than once` and `DEC-055`.
- `AG2`: the contract's `- **\`brag_add\`**` bullet names the same two.
- `AG3`: `docs/for-ai-agents.md`'s `### \`brag_add\`` section names
  `brag_add: key "impact" appears more than once`.
- `AG4`: `BRAG.md`'s `## JSON contract` section names
  `**A repeated key is rejected.**`.
- `AG5`: `CHANGELOG.md` `[Unreleased]` names the breaking bullet's head and
  `DEC-055`.
- `AG6`: `[Unreleased]` no longer contains `is a separate decoder and is
  unchanged`. It is a negative, paired with `AG5`'s positive on the same
  section.

### Decision-to-test mapping (§9)

| Decision | Test |
|---|---|
| LD1 reject, atomic | the parity test (0 rows on both ingresses), both per-field tests, AC-1/AC-3 (hash) |
| LD2 key identity | `…FoldsExactlyLikeEncodingJSON`, `…ComparesDecodedKeys`, parity rows 1–4. M-4, M-5 |
| LD3 top level only | `…TopLevelOnly`, parity rows 12a/12b. M-11 |
| LD4 every key | both per-field tests (derived key sets) |
| LD5 first, on raw bytes; malformed left alone | parity rows 5, 7–10 (the CLI decoder would say something else), `…LeavesOtherInputToTheDecoder`. M-7, M-9 |
| LD6 MCP middleware | parity rows 1, 2, 3, 5, 8, 10, `TestServer_AddRejectsARepeatOfEveryField`. M-2, M-3, M-8 |
| LD7 message | `…NamesTheKey`, the exact-string checks in the parity and per-field tests. M-6, M-10 |
| LD8 docs | `AG1`–`AG6`. M-D1–M-D7 |
| DEC-055 as a whole, one rule on both ingresses | the parity test. M-1 and M-2 are the one-ingress shortcut |

---

## Implementation Context

### Decisions that apply

- **DEC-055** (written at design): the whole rule. Its six choices are LD1–LD7
  here.
- **DEC-012**, with its 2026-09-27 amendment: the schema this adds a rule to.
  Choices 4 and 5 are narrowed, and the rest is unchanged.
- **DEC-051**: the editor instance. Its `canonicalHeaders` check is unaffected.
- **DEC-007**: the exit-1 path (`UserErrorf`).
- **DEC-024 / DEC-027**: the `brag_add` provenance fields, repeated like any
  other key.

### Constraints that apply

- `no-sql-in-cli-layer`: `internal/capture` imports only the standard library.
  The parity test lives in `internal/cli` as `package cli_test` and imports
  `internal/storage` and `internal/mcpserver`, never `database/sql`.
- `stdout-is-for-data-stderr-is-for-humans`: a rejection prints nothing on
  stdout.
- `one-spec-per-pr`.
- `AGENTS.md` §12: every pinned diff has one literal reading, and a baseline
  hash names its commit on `main`.

### Prior related work

- **SPEC-089**: the editor half, and the lesson that a guard derives its scope
  from the thing it guards (hence the struct-tag-derived key sets).
- **SPEC-095**: the matrix and literal conventions this spec follows.

### Out of scope (for this spec specifically)

SPEC-097 (`--type`), SPEC-091, SPEC-092, SPEC-093 (the probe helper, which
stays in the design session's scratchpad), SPEC-096, and the release cut.
CLI/MCP disagreements that involve no repeated key (a lone `"Impact"`, a lone
`"id"`). `mcp_install`'s last-wins decode of the client config. The read-only
MCP tools (`brag_list`, `brag_search`, `brag_memory`). A repeat that a JSON
library collapsed upstream of brag. Brags.

---

## Corrections and open questions (design)

1. **Framing's row 4 is the escaped key `"\u0069mpact"`.** It was printed as
   `impact` in the framing table (and at *What design must settle* item 1),
   because the escape was lost in writing it down. The same loss hit this session twice: zsh's `printf '%s'` and the
   session's own tool layer both decoded `\u0069` to `i` before it reached a
   file. It is a trap for build (*Traps*), and the fixtures guard against it.
2. **Framing's *"MCP already rejects case variants, so byte-equal-after-unescape
   is enough there"* holds only for the SDK's current order of checks.** With
   the check in middleware ahead of the SDK (LD6), MCP now reports
   `"Impact"` + `"impact"` as a repeat, not as an additional property. That
   is what makes the two messages identical.
3. **Framing's *"the MCP set is 11 keys; the CLI set is 9"*** was re-derived
   from the struct tags: `addIn` has 11 json keys and `addJSONInput` has 9.
   Both tests floor at those numbers.
4. **Framing located the SDK's map collapse at `go-sdk@v1.8.0 mcp/tool.go:98`.**
   That is `applySchema`. The validate-then-decode wrapper that calls it is
   `toolForErr` in `mcp/server.go` (lines 395–420 at v1.8.0), and it is the
   reason the check cannot sit in `handleAdd` (M-3).
5. **The parity test cannot make rows 12a and 12b agree**, and does not try
   to. It pins only that neither ingress calls a nested repeat a repeat. If a
   later spec resolves the lone-`"id"` disagreement, those two rows can gain
   an outcome assertion.
6. **Open, not blocking:** whether the read-only MCP tools should get the same
   middleware. A repeated `brag_list` filter key reads last-wins today (driven
   above). It changes a read, not the corpus, so it is outside PROJ-008's
   honesty line. It is named here so nobody re-finds it as a gap.
7. **Open, not blocking: an SDK bump.** go-sdk's middleware contract (a
   `tools/call` arriving as `*mcp.CallToolRequest` with raw `Arguments`, ahead
   of `toolForErr`) is what LD6 rests on. The Go tests cover it through the
   in-memory transport. A bump should also re-run *The MCP reject path* on the
   stdio binary, per the project's R1 lesson and the #124 bump memory.

---

## Notes for the Implementer

### Order of work

1. `git switch main && git pull --ff-only`. Confirm this design merged and that
   `decisions/DEC-055-*.md` exists. Then branch `build/spec-090-json-duplicate-key`.
2. **Apply the test literals first** (§4 to §7), plus a one-line stub
   `func CheckRepeatedKeys(raw []byte) error { return nil }` in
   `internal/capture/repeated_key.go`. Run `go test ./...` and expect exactly
   *Finding 1*'s reds. Then replace the stub with §1.
3. **Apply the rest.** Each `diff` block below is `git diff` output against
   **`01b10ca`** (on `main`). None of the build files is touched by the design
   commit, so every block applies unchanged to `main` after this design
   merges. Save them to one file and `git apply` it. `git apply --check`
   accepted the whole patch against this design branch. If `main` has moved
   and a hunk no longer applies, apply it by hand from its context: every
   modified file's hunk has an old side that occurs exactly once in its base
   file.
4. Run `just inventory` and paste its output between the markers in
   `docs/engineering-practices.md`. **Do not hand-edit it.** *Inventory* says
   which rows should move.
5. Re-run the matrix (18 kept probes) from the stated diffs, the five gates,
   and AC-1 to AC-5 on a scratch store and the real `mcp serve`.
6. `just advance-cycle SPEC-090 build`, then **restore the inline enum comment
   it strips** from `cycle:` (column 36).

### Traps

- **`\u` escapes do not survive a trip through zsh or this tool layer.**
  zsh's builtin `printf '%s'` decodes `\u0069` to `i`, and so did this
  session's file-writing tool, silently turning the escaped-key row into a
  plain repeat that still passes. The Go fixtures spell the backslash as
  `"\x5c"` and assert that the escape survived. Write payload files from
  Python with `chr(92)`, and check them with `xxd`.
- **Non-ASCII fixtures are written as `string(rune(0x017F))`** and the like,
  for the same reason. Do not "tidy" them into literals.
- **`W5` greps `Name:        "brag_add"`** in `server.go`. Keep the literal
  (LD6).
- **Driving `brag mcp serve` needs a held-open stdin.** A piped file hits EOF
  and the server exits before it answers. Always pass `--db` to a scratch path.
- **An MCP client library collapses a repeat before sending it.** The Go tests
  pass `Arguments: json.RawMessage(payload)`, which design confirmed reaches
  the middleware intact. A `map[string]any` cannot carry the repeat.
- **A *no difference* is a measurement.** For AC-5, check each side's exit
  code, stderr and stored row before comparing.
- **zsh:** write JSON to files, and never pipe it through `echo`. Never name a
  variable `path`. Use `/usr/bin/diff`, since the shell's `diff` is rewritten
  by a hook.
- **AC2** fails on any tree that is not a git checkout. Run `test-docs` in a
  worktree.

### §1. The detector (LD2, LD3, LD5, LD7)

`internal/capture/repeated_key.go`: 1 hunk, base `01b10ca`, file hash after `9e37c53c7f54`.

````diff
diff --git a/internal/capture/repeated_key.go b/internal/capture/repeated_key.go
new file mode 100644
index 0000000..cfb0d6a
--- /dev/null
+++ b/internal/capture/repeated_key.go
@@ -0,0 +1,103 @@
+package capture
+
+import (
+	"bytes"
+	"encoding/json"
+	"fmt"
+	"unicode"
+	"unicode/utf8"
+)
+
+// CheckRepeatedKeys rejects a JSON object whose top level names the same key
+// twice (DEC-055). It runs on the raw bytes, before any decoder sees them,
+// because both machine ingresses decode last-wins with no error: by the time
+// a struct holds the value, the repeat is gone.
+//
+// Two keys are the same key when they are equal after unescaping and case
+// folding — the rule encoding/json uses to match a key to a struct field
+// (foldKey). The MCP SDK matches case-sensitively, but one rule for both
+// ingresses has to be the wider one, or `add --json` keeps clobbering on
+// "impact" + "Impact".
+//
+// Only the top level is checked. No field brag stores holds an object, so a
+// repeat inside a nested value either fails that field's type check or sits
+// in a server-owned value DEC-012 discards.
+//
+// Anything that is not a well-formed object is left to the caller's decoder,
+// which already reports it: this returns nil for an array, a scalar, or a
+// syntax error anywhere in the object, so input that is broken in another way
+// keeps the message it has today.
+func CheckRepeatedKeys(raw []byte) error {
+	dec := json.NewDecoder(bytes.NewReader(raw))
+	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
+		return nil
+	}
+	seen := map[string]string{} // folded key → the key as first sent
+	var repeat error
+	for dec.More() {
+		tok, err := dec.Token()
+		if err != nil {
+			return nil
+		}
+		key, ok := tok.(string)
+		if !ok {
+			return nil
+		}
+		folded := foldKey(key)
+		if first, dup := seen[folded]; !dup {
+			seen[folded] = key
+		} else if repeat == nil {
+			repeat = repeatedKeyError(first, key)
+		}
+		var skip json.RawMessage
+		if err := dec.Decode(&skip); err != nil {
+			return nil
+		}
+	}
+	// Report a repeat only in an object that closes: a truncated one is the
+	// decoder's to report, whatever it repeated before it broke.
+	if _, err := dec.Token(); err != nil {
+		return nil
+	}
+	return repeat
+}
+
+// repeatedKeyError names the key as sent at its second occurrence, and the
+// first spelling too when the two differ only by case.
+func repeatedKeyError(first, again string) error {
+	if first == again {
+		return fmt.Errorf("key %q appears more than once", again)
+	}
+	return fmt.Errorf("key %q appears more than once (first as %q; keys match case-insensitively)", again, first)
+}
+
+// foldKey maps every spelling encoding/json would match to one struct field
+// onto the same string, so that foldKey(a) == foldKey(b) exactly when
+// strings.EqualFold(a, b). It is encoding/json's own fold (appendFoldedName in
+// its fold.go): ASCII letters upper-cased, and any other rune replaced by the
+// smallest rune in its unicode.SimpleFold orbit, so "ſ" (U+017F) and "s" both
+// fold to "S". TestCheckRepeatedKeys_FoldsExactlyLikeEncodingJSON holds the
+// two in step.
+func foldKey(key string) string {
+	var b []byte
+	for _, r := range key {
+		b = utf8.AppendRune(b, foldRune(r))
+	}
+	return string(b)
+}
+
+func foldRune(r rune) rune {
+	if 'a' <= r && r <= 'z' {
+		return r - ('a' - 'A')
+	}
+	if r < utf8.RuneSelf {
+		return r
+	}
+	for {
+		next := unicode.SimpleFold(r)
+		if next <= r {
+			return next
+		}
+		r = next
+	}
+}
````

### §2. The CLI wiring (LD1, LD5)

`internal/cli/add_json.go`: 2 hunks, base `01b10ca`, file hash after `333e7c341785`.

````diff
diff --git a/internal/cli/add_json.go b/internal/cli/add_json.go
index 954ca7d..c60814c 100644
--- a/internal/cli/add_json.go
+++ b/internal/cli/add_json.go
@@ -1,6 +1,7 @@
 package cli

 import (
+	"bytes"
 	"encoding/json"
 	"fmt"
 	"io"
@@ -49,7 +50,16 @@ type addJSONInput struct {
 // hydrated storage.Entry. Server-owned fields on the input are dropped.
 // All errors route through UserErrorf so ErrUser propagates.
 func parseAddJSON(r io.Reader) (storage.Entry, error) {
-	dec := json.NewDecoder(r)
+	raw, err := io.ReadAll(r)
+	if err != nil {
+		return storage.Entry{}, UserErrorf("invalid JSON input: %v", err)
+	}
+	// Before the decode, which keeps the last of a repeated key and says
+	// nothing (DEC-055). MCP brag_add runs the same check.
+	if err := capture.CheckRepeatedKeys(raw); err != nil {
+		return storage.Entry{}, UserErrorf("--json input: %v", err)
+	}
+	dec := json.NewDecoder(bytes.NewReader(raw))
 	dec.DisallowUnknownFields()

 	var in addJSONInput
````

### §3. The MCP wiring (LD1, LD6)

`internal/mcpserver/server.go`: 1 hunk, base `01b10ca`, file hash after `d24ee0b3ee64`.

````diff
diff --git a/internal/mcpserver/server.go b/internal/mcpserver/server.go
index 4ca8927..c19ab57 100644
--- a/internal/mcpserver/server.go
+++ b/internal/mcpserver/server.go
@@ -61,9 +61,29 @@ func New(s *storage.Store) *mcp.Server {

 	addResources(srv, s)

+	srv.AddReceivingMiddleware(rejectRepeatedKeys)
+
 	return srv
 }

+// rejectRepeatedKeys runs the repeated-key check (DEC-055) on brag_add's raw
+// argument bytes. It is middleware, not a line in handleAdd, because the SDK
+// validates the arguments against the schema and decodes them last-wins before
+// the handler runs: by then a repeat is either gone or reported as something
+// else. `brag add --json` runs the same check before its own decode.
+func rejectRepeatedKeys(next mcp.MethodHandler) mcp.MethodHandler {
+	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
+		if call, ok := req.(*mcp.CallToolRequest); ok && call.Params.Name == "brag_add" {
+			if err := capture.CheckRepeatedKeys(call.Params.Arguments); err != nil {
+				res := &mcp.CallToolResult{}
+				res.SetError(fmt.Errorf("brag_add: %w", err))
+				return res, nil
+			}
+		}
+		return next(ctx, method, req)
+	}
+}
+
 // addIn is brag_add's input shape. Title has no `,omitempty` so the SDK's
 // inferred schema marks it required; the other user-owned fields mirror
 // parseAddJSON's addJSONInput (DEC-012). Agent/Model are the explicit
````

### §4. Detector tests

`internal/capture/repeated_key_test.go`: 1 hunk, base `01b10ca`, file hash after `a9bada643c1f`.

````diff
diff --git a/internal/capture/repeated_key_test.go b/internal/capture/repeated_key_test.go
new file mode 100644
index 0000000..63dcf32
--- /dev/null
+++ b/internal/capture/repeated_key_test.go
@@ -0,0 +1,150 @@
+package capture
+
+import (
+	"encoding/json"
+	"reflect"
+	"strings"
+	"testing"
+)
+
+// encodingJSONMatches reports whether encoding/json, decoding
+// {"<sent>":"v"} into a struct whose only field is tagged `json:"<field>"`,
+// puts the value in that field: the stdlib's own answer to "is this the same
+// key", which the CLI ingress inherits.
+func encodingJSONMatches(t *testing.T, field, sent string) bool {
+	t.Helper()
+	typ := reflect.StructOf([]reflect.StructField{{
+		Name: "F",
+		Type: reflect.TypeFor[string](),
+		Tag:  reflect.StructTag(`json:"` + field + `"`),
+	}})
+	v := reflect.New(typ)
+	raw, err := json.Marshal(map[string]string{sent: "v"})
+	if err != nil {
+		t.Fatalf("marshal %q: %v", sent, err)
+	}
+	if err := json.Unmarshal(raw, v.Interface()); err != nil {
+		t.Fatalf("unmarshal %s: %v", raw, err)
+	}
+	return v.Elem().Field(0).String() == "v"
+}
+
+// TestCheckRepeatedKeys_FoldsExactlyLikeEncodingJSON pins DEC-055's
+// key-identity rule to its source: for every pair, the detector calls the
+// second key a repeat of the first exactly when encoding/json would decode
+// both into one field. If a Go release changes how the stdlib folds, this
+// fails instead of the CLI quietly clobbering again.
+func TestCheckRepeatedKeys_FoldsExactlyLikeEncodingJSON(t *testing.T) {
+	longS := string(rune(0x017F))     // ſ, LATIN SMALL LETTER LONG S
+	kelvin := string(rune(0x212A))    // K, KELVIN SIGN
+	sharpS := string(rune(0x00DF))    // ß: does not simple-fold to "ss"
+	cyrillicE := string(rune(0x0435)) // е: a homoglyph of "e", not a fold
+	pairs := []struct{ first, second string }{
+		{"impact", "impact"},
+		{"impact", "Impact"},
+		{"IMPACT", "impact"},
+		{"tags", "tag" + longS},
+		{"kind", kelvin + "ind"},
+		{"title", "TITLE"},
+		{"strasse", "stra" + sharpS + "e"},
+		{"impact", "impacts"},
+		{"impact", "imp_act"},
+		{"type", "typ" + cyrillicE},
+		{string(rune(0x01C5)), string(rune(0x01C6))}, // ǅ and ǆ: title case, lower case
+	}
+	matched := 0
+	for _, p := range pairs {
+		want := encodingJSONMatches(t, p.first, p.second)
+		if want {
+			matched++
+		}
+		raw := `{"` + p.first + `":"a","` + p.second + `":"b"}`
+		got := CheckRepeatedKeys([]byte(raw)) != nil
+		if got != want {
+			t.Errorf("%s: detector says repeat=%v, encoding/json matches one field=%v", raw, got, want)
+		}
+	}
+	// Non-vacuity: the table must hold both answers, or it pins nothing.
+	if matched < 6 || matched == len(pairs) {
+		t.Fatalf("encoding/json matched %d of %d pairs; the table no longer tests both sides", matched, len(pairs))
+	}
+}
+
+// TestCheckRepeatedKeys_ComparesDecodedKeys: an escaped spelling of a key
+// is the same key. Both ingresses unescape before matching, so a raw-byte
+// comparison would miss it.
+func TestCheckRepeatedKeys_ComparesDecodedKeys(t *testing.T) {
+	escaped := `{"impact":"REAL","` + "\x5c" + `u0069mpact":"CLOBBERED"}`
+	if !strings.Contains(escaped, "\x5cu0069") {
+		t.Fatalf("fixture lost its escape: %s", escaped)
+	}
+	err := CheckRepeatedKeys([]byte(escaped))
+	if err == nil {
+		t.Fatalf("%s: want a repeat, got nil", escaped)
+	}
+	if want := `key "impact" appears more than once`; err.Error() != want {
+		t.Errorf("got %q, want %q", err.Error(), want)
+	}
+}
+
+// TestCheckRepeatedKeys_NamesTheKey locks the two message shapes: the key as
+// sent at its second occurrence, and the first spelling when they differ.
+func TestCheckRepeatedKeys_NamesTheKey(t *testing.T) {
+	cases := []struct{ raw, want string }{
+		{`{"title":"t","impact":"A","impact":"B"}`, `key "impact" appears more than once`},
+		{`{"title":"t","impact":"A","Impact":"B"}`, `key "Impact" appears more than once (first as "impact"; keys match case-insensitively)`},
+		{`{"impact":"A","impact":null}`, `key "impact" appears more than once`},
+		{`{"type":"shipped","type":"shipped"}`, `key "type" appears more than once`},
+		{`{"a":1,"b":2,"c":3,"b":4,"a":5}`, `key "b" appears more than once`},
+	}
+	for _, tc := range cases {
+		err := CheckRepeatedKeys([]byte(tc.raw))
+		if err == nil {
+			t.Errorf("%s: want %q, got nil", tc.raw, tc.want)
+			continue
+		}
+		if err.Error() != tc.want {
+			t.Errorf("%s: got %q, want %q", tc.raw, err.Error(), tc.want)
+		}
+	}
+}
+
+// TestCheckRepeatedKeys_TopLevelOnly: a repeat inside a nested value is not
+// this rule's (DEC-055). Every value shape is skipped whole, so a key inside
+// one never collides with a top-level key either.
+func TestCheckRepeatedKeys_TopLevelOnly(t *testing.T) {
+	for _, raw := range []string{
+		`{"title":"t","id":{"a":1,"a":2}}`,
+		`{"title":"t","id":[{"a":1,"a":2}]}`,
+		`{"title":"t","id":{"title":"nested"}}`,
+		`{"a":[1,{"b":2}],"b":null,"c":{"d":{"a":1}},"d":"x"}`,
+	} {
+		if err := CheckRepeatedKeys([]byte(raw)); err != nil {
+			t.Errorf("%s: want nil, got %v", raw, err)
+		}
+	}
+	// ...and a repeat that follows nested values is still found.
+	raw := `{"a":[1,{"b":2}],"c":{"d":{"a":1}},"a":3}`
+	if err := CheckRepeatedKeys([]byte(raw)); err == nil {
+		t.Errorf("%s: want a repeat of \"a\", got nil", raw)
+	}
+}
+
+// TestCheckRepeatedKeys_LeavesOtherInputToTheDecoder: what is not a
+// well-formed object gets no second message from here; the ingress's own
+// decoder already reports it (DEC-012 choice 1, invalid syntax).
+func TestCheckRepeatedKeys_LeavesOtherInputToTheDecoder(t *testing.T) {
+	for _, raw := range []string{
+		``,
+		`[{"a":1,"a":2}]`,
+		`"a"`,
+		`{"title":`,
+		`{"a":1,"a"`,
+		`{"a":1,"a":2`,
+		`{"a":1}{"a":1}`,
+	} {
+		if err := CheckRepeatedKeys([]byte(raw)); err != nil {
+			t.Errorf("%q: want nil, got %v", raw, err)
+		}
+	}
+}
````

### §5. The CLI per-field test

`internal/cli/add_json_test.go`: 2 hunks, base `01b10ca`, file hash after `1cf1ca59f5af`.

````diff
diff --git a/internal/cli/add_json_test.go b/internal/cli/add_json_test.go
index 5ddb22d..2b4d3b2 100644
--- a/internal/cli/add_json_test.go
+++ b/internal/cli/add_json_test.go
@@ -4,6 +4,7 @@ import (
 	"bytes"
 	"encoding/json"
 	"errors"
+	"reflect"
 	"regexp"
 	"strconv"
 	"strings"
@@ -627,3 +628,52 @@ func TestAddJSON_ExplicitProjectWins(t *testing.T) {
 		t.Errorf("Project: got %q want %q", entries[0].Project, "explicit")
 	}
 }
+
+// TestAddCmd_JSON_RepeatOfEveryFieldIsRejected: a repeat of any key the
+// schema accepts, server-owned ones included, writes nothing and names the
+// key (DEC-055). The keys come from addJSONInput's own tags rather than a
+// typed list, so a field added later is covered with no edit here.
+func TestAddCmd_JSON_RepeatOfEveryFieldIsRejected(t *testing.T) {
+	var keys []string
+	typ := reflect.TypeFor[addJSONInput]()
+	for i := range typ.NumField() {
+		name, _, _ := strings.Cut(typ.Field(i).Tag.Get("json"), ",")
+		keys = append(keys, name)
+	}
+	// Non-vacuity: DEC-012's six user-owned keys and three server-owned ones.
+	if len(keys) < 9 {
+		t.Fatalf("addJSONInput has %d json keys, want at least 9: %v", len(keys), keys)
+	}
+	for _, k := range keys {
+		t.Run(k, func(t *testing.T) {
+			stdin := `{"title":"t","` + k + `":"a","` + k + `":"b"}`
+			if k == "title" {
+				stdin = `{"title":"a","title":"b"}`
+			}
+			root, dbPath := newRootWithAdd(t)
+			var outBuf, errBuf bytes.Buffer
+			root.SetOut(&outBuf)
+			root.SetErr(&errBuf)
+			root.SetIn(strings.NewReader(stdin))
+			root.SetArgs([]string{"--db", dbPath, "add", "--json"})
+
+			err := root.Execute()
+			if err == nil {
+				t.Fatalf("%s: expected error, got nil", stdin)
+			}
+			if !errors.Is(err, ErrUser) {
+				t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
+			}
+			want := `--json input: key "` + k + `" appears more than once`
+			if !strings.Contains(err.Error(), want) {
+				t.Errorf("got %q, want it to contain %q", err.Error(), want)
+			}
+			if outBuf.Len() != 0 {
+				t.Errorf("expected stdout empty, got %q", outBuf.String())
+			}
+			if got := len(listAll(t, dbPath)); got != 0 {
+				t.Errorf("expected 0 entries, got %d", got)
+			}
+		})
+	}
+}
````

### §6. The parity test (LOAD-BEARING)

`internal/cli/add_json_parity_test.go`: 1 hunk, base `01b10ca`, file hash after `ca2bfb03601c`.

````diff
diff --git a/internal/cli/add_json_parity_test.go b/internal/cli/add_json_parity_test.go
new file mode 100644
index 0000000..9d06aa2
--- /dev/null
+++ b/internal/cli/add_json_parity_test.go
@@ -0,0 +1,144 @@
+package cli_test
+
+import (
+	"bytes"
+	"context"
+	"encoding/json"
+	"errors"
+	"path/filepath"
+	"strings"
+	"testing"
+
+	"github.com/jysf/bragfile000/internal/cli"
+	"github.com/jysf/bragfile000/internal/mcpserver"
+	"github.com/jysf/bragfile000/internal/storage"
+	"github.com/modelcontextprotocol/go-sdk/mcp"
+)
+
+// TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant runs one table against
+// both machine ingresses (DEC-055). Before it, the two decoders disagreed on
+// six of these rows, because encoding/json folds case and the MCP SDK does
+// not, and because the SDK drops an earlier value before its schema check.
+// Now every row that repeats a top-level key is refused on both, with the
+// same message after each ingress's prefix, and nothing is written.
+//
+// The last two rows repeat a key only inside a value, which is not this
+// rule's: neither ingress may call them a repeat. What each then does with
+// "id" is the lone-"id" difference DEC-012 already has, and is not pinned
+// here.
+func TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant(t *testing.T) {
+	longS := string(rune(0x017F)) // ſ, which encoding/json folds to "s"
+	escapedImpact := "\x5c" + "u0069mpact"
+	rows := []struct {
+		name, payload, want string
+	}{
+		{"plain", `{"title":"t","impact":"REAL","impact":"CLOBBERED"}`, `key "impact" appears more than once`},
+		{"case_fold", `{"title":"t","impact":"REAL","Impact":"CLOBBERED"}`, `key "Impact" appears more than once (first as "impact"; keys match case-insensitively)`},
+		{"upper_first", `{"title":"t","IMPACT":"REAL","impact":"CLOBBERED"}`, `key "impact" appears more than once (first as "IMPACT"; keys match case-insensitively)`},
+		{"unicode_fold", `{"title":"t","tags":"real","tag` + longS + `":"clobbered"}`, `key "tag` + longS + `" appears more than once (first as "tags"; keys match case-insensitively)`},
+		{"escaped_key", `{"title":"t","impact":"REAL","` + escapedImpact + `":"CLOBBERED"}`, `key "impact" appears more than once`},
+		{"trailing_null", `{"title":"t","impact":"REAL","impact":null}`, `key "impact" appears more than once`},
+		{"trailing_empty", `{"title":"t","impact":"REAL","impact":""}`, `key "impact" appears more than once`},
+		{"object_then_string", `{"title":"t","impact":{"x":1},"impact":"CLOBBERED"}`, `key "impact" appears more than once`},
+		{"string_then_object", `{"title":"t","impact":"REAL","impact":{"x":1}}`, `key "impact" appears more than once`},
+		{"array_then_string", `{"title":"t","tags":["a"],"tags":"b"}`, `key "tags" appears more than once`},
+		{"string_then_array", `{"title":"t","tags":"a","tags":["b"]}`, `key "tags" appears more than once`},
+		{"same_value", `{"title":"t","type":"shipped","type":"shipped"}`, `key "type" appears more than once`},
+		{"nested_object", `{"title":"t","id":{"a":1,"a":2}}`, ""},
+		{"nested_array", `{"title":"t","id":[{"a":1,"a":2}]}`, ""},
+	}
+	if !strings.Contains(rows[4].payload, "\x5cu0069") {
+		t.Fatalf("escaped_key fixture lost its escape: %s", rows[4].payload)
+	}
+	for _, row := range rows {
+		t.Run(row.name, func(t *testing.T) {
+			cliRows, cliErr := runAddJSON(t, row.payload)
+			mcpText, mcpIsError, mcpRows := callBragAdd(t, row.payload)
+
+			if row.want == "" {
+				if cliErr != nil && strings.Contains(cliErr.Error(), "appears more than once") {
+					t.Errorf("add --json called a nested repeat a repeat: %v", cliErr)
+				}
+				if strings.Contains(mcpText, "appears more than once") {
+					t.Errorf("brag_add called a nested repeat a repeat: %s", mcpText)
+				}
+				return
+			}
+			if !errors.Is(cliErr, cli.ErrUser) {
+				t.Fatalf("add --json: want a user error (exit 1), got %v", cliErr)
+			}
+			if got, want := cliErr.Error(), "user error: --json input: "+row.want; got != want {
+				t.Errorf("add --json:\n got %q\nwant %q", got, want)
+			}
+			if !mcpIsError {
+				t.Fatalf("brag_add: want isError, got a success: %s", mcpText)
+			}
+			if want := "brag_add: " + row.want; mcpText != want {
+				t.Errorf("brag_add:\n got %q\nwant %q", mcpText, want)
+			}
+			if cliRows != 0 || mcpRows != 0 {
+				t.Errorf("want nothing written; add --json stored %d, brag_add stored %d", cliRows, mcpRows)
+			}
+		})
+	}
+}
+
+// runAddJSON runs `brag add --json` with payload on stdin against a fresh
+// store, and returns how many rows it left and the command's error.
+func runAddJSON(t *testing.T, payload string) (int, error) {
+	t.Helper()
+	dbPath := filepath.Join(t.TempDir(), "cli.db")
+	root := cli.NewRootCmd("test")
+	root.AddCommand(cli.NewAddCmd())
+	var outBuf, errBuf bytes.Buffer
+	root.SetOut(&outBuf)
+	root.SetErr(&errBuf)
+	root.SetIn(strings.NewReader(payload))
+	root.SetArgs([]string{"--db", dbPath, "add", "--json"})
+	err := root.Execute()
+	return countRows(t, dbPath), err
+}
+
+// callBragAdd sends payload as brag_add's raw argument bytes to a server on a
+// fresh store, and returns the result text, its isError flag, and how many
+// rows it left. A map could not carry a repeated key, so the arguments are a
+// json.RawMessage.
+func callBragAdd(t *testing.T, payload string) (string, bool, int) {
+	t.Helper()
+	dbPath := filepath.Join(t.TempDir(), "mcp.db")
+	s, err := storage.Open(dbPath)
+	if err != nil {
+		t.Fatalf("storage.Open: %v", err)
+	}
+	ctx := context.Background()
+	ct, st := mcp.NewInMemoryTransports()
+	if _, err := mcpserver.New(s).Connect(ctx, st, nil); err != nil {
+		t.Fatalf("server connect: %v", err)
+	}
+	cs, err := mcp.NewClient(&mcp.Implementation{Name: "parity", Version: "1"}, nil).Connect(ctx, ct, nil)
+	if err != nil {
+		t.Fatalf("client connect: %v", err)
+	}
+	r, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "brag_add", Arguments: json.RawMessage(payload)})
+	if err != nil {
+		t.Fatalf("brag_add transport error: %v", err)
+	}
+	_ = cs.Close()
+	s.Close()
+	text := r.Content[0].(*mcp.TextContent).Text
+	return text, r.IsError, countRows(t, dbPath)
+}
+
+func countRows(t *testing.T, dbPath string) int {
+	t.Helper()
+	s, err := storage.Open(dbPath)
+	if err != nil {
+		t.Fatalf("storage.Open: %v", err)
+	}
+	defer s.Close()
+	got, err := s.List(storage.ListFilter{})
+	if err != nil {
+		t.Fatalf("List: %v", err)
+	}
+	return len(got)
+}
````

### §7. The MCP per-field test (diffed at `-U4`)

`internal/mcpserver/server_test.go`: 2 hunks, base `01b10ca`, file hash after `82cd9b29cfb7`.

````diff
diff --git a/internal/mcpserver/server_test.go b/internal/mcpserver/server_test.go
index c874aea..607e44a 100644
--- a/internal/mcpserver/server_test.go
+++ b/internal/mcpserver/server_test.go
@@ -5,8 +5,9 @@ import (
 	"encoding/json"
 	"path/filepath"
 	"reflect"
 	"sort"
+	"strings"
 	"testing"
 	"time"

 	"github.com/jysf/bragfile000/internal/export"
@@ -314,4 +315,48 @@ func TestServer_AddReturnValueParity(t *testing.T) {
 			t.Errorf("field %q not byte-identical: brag_add=%s export=%s", k, gv, wv)
 		}
 	}
 }
+
+// TestServer_AddRejectsARepeatOfEveryField: a repeat of any brag_add key,
+// the provenance ones included, is a tool error naming the key, and nothing
+// is written (DEC-055). The keys come from addIn's own tags, so a field added
+// later is covered with no edit here. The arguments go over the wire as raw
+// bytes: a map could not hold the repeat.
+func TestServer_AddRejectsARepeatOfEveryField(t *testing.T) {
+	var keys []string
+	typ := reflect.TypeFor[addIn]()
+	for i := range typ.NumField() {
+		name, _, _ := strings.Cut(typ.Field(i).Tag.Get("json"), ",")
+		keys = append(keys, name)
+	}
+	// Non-vacuity: six entry fields and DEC-024/DEC-027's five provenance ones.
+	if len(keys) < 11 {
+		t.Fatalf("addIn has %d json keys, want at least 11: %v", len(keys), keys)
+	}
+	cs, s := newTestServer(t, "repeat-probe")
+	for _, k := range keys {
+		raw := `{"title":"t","` + k + `":"1","` + k + `":"2"}`
+		if k == "title" {
+			raw = `{"title":"a","title":"b"}`
+		}
+		r, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "brag_add", Arguments: json.RawMessage(raw)})
+		if err != nil {
+			t.Fatalf("%s: transport error: %v", raw, err)
+		}
+		if !r.IsError {
+			t.Errorf("%s: want isError, got a success", raw)
+			continue
+		}
+		text := r.Content[0].(*mcp.TextContent).Text
+		if want := `brag_add: key "` + k + `" appears more than once`; text != want {
+			t.Errorf("%s: got %q, want %q", raw, text, want)
+		}
+	}
+	rows, err := s.List(storage.ListFilter{})
+	if err != nil {
+		t.Fatalf("List: %v", err)
+	}
+	if len(rows) != 0 {
+		t.Errorf("want nothing written, got %d rows", len(rows))
+	}
+}
````

### §8. `docs/api-contract.md` (LD8)

`docs/api-contract.md`: 3 hunks, base `01b10ca`, file hash after `b42abc011274`.

````diff
diff --git a/docs/api-contract.md b/docs/api-contract.md
index 5b1c2c7..11ca872 100644
--- a/docs/api-contract.md
+++ b/docs/api-contract.md
@@ -102,6 +102,14 @@ brag list --format json | jq '.[0]' | brag add --json
 - Unknown keys are strict-rejected with the offending key named in
   the error (catches typos like `"titl"` before they become silently-
   missing entries).
+- A repeated top-level key is rejected before anything is decoded:
+  exit 1, nothing written, and stderr names the key
+  (`--json input: key "impact" appears more than once`). Two keys are
+  the same key when they match case-insensitively after unescaping
+  (`"impact"`, `"Impact"` and an escaped spelling are one key), and
+  a repeated server-owned key counts too. A repeat inside a nested
+  value does not. The MCP `brag_add` tool applies the same rule
+  ([DEC-055](../decisions/DEC-055-a-repeated-json-key-is-rejected-on-both-machine-ingresses.md)).
 - `tags` stays a comma-joined string per
   [DEC-004](../decisions/DEC-004-tags-comma-joined-for-mvp.md); array
   form (`["a","b"]`) is rejected with an error naming DEC-004.
@@ -1366,8 +1374,13 @@ same `~/.bragfile/db.sqlite` the CLI uses:
   and the optional `session`/`cost`/`tokens` seed provenance params
   (DEC-027). Inserts via `Store.Add` and returns the created entry
   as a single [DEC-011](../decisions/DEC-011-json-output-shape.md) object.
-  A missing/empty `title` is a tool error, never a silent insert. Unlike
-  `brag add`, the MCP tool does **not** emit a SPEC-039 milestone line and
+  A missing/empty `title` is a tool error, never a silent insert. So is a
+  repeated top-level key, with nothing written:
+  `brag_add: key "impact" appears more than once`. It is checked on the raw
+  arguments, before the SDK validates or decodes them, by the same rule as
+  `brag add --json`
+  ([DEC-055](../decisions/DEC-055-a-repeated-json-key-is-rejected-on-both-machine-ingresses.md)).
+  Unlike `brag add`, the MCP tool does **not** emit a SPEC-039 milestone line and
   does **not** auto-fill `project` from a server-side cwd — the MCP server
   has no meaningful cwd relative to the calling agent.
 - **`brag_list`** — filters `tag`/`project`/`type` (exact match), the time
@@ -1528,6 +1541,7 @@ Machine-parseable output is stdout only; stderr is for humans.
 - `DEC-011` — shared JSON output shape for `brag list --format json` and `brag export --format json`
 - `DEC-013` — markdown export shape for `brag export --format markdown` (+`--flat`)
 - `DEC-012` — stdin-JSON schema for `brag add --json` (single object, title required, server-owned fields tolerated-and-ignored)
+- `DEC-055` — a repeated top-level key is rejected on `brag add --json` and MCP `brag_add`, with nothing written: keys are compared after unescaping and case-insensitively (encoding/json's own field match), top level only, by one shared check that runs before either decoder.
 - `DEC-014` — rule-based output shape for `brag summary`, `brag review`, `brag stats`, and `brag impact`: single-object JSON envelope with `generated_at` / `scope` / `filters` provenance + per-spec payload keys; markdown convention reuses DEC-013's provenance + summary-block style.
 - `DEC-016` — tag mutation semantics: `brag tags` in-use-only taxonomy (count-DESC/name-ASC; `{tag,count}` JSON shape), rename-errors-into-existing, merge via DELETE+INSERT, orphan tags invisible (no GC).
 - `DEC-017` — `entries.project` ↔ `projects` relationship (soft string match) + `projects.status` enum + single `state_note`; the data `brag project show`/`list` render.
````

### §9. `docs/for-ai-agents.md` (LD8)

`docs/for-ai-agents.md`: 1 hunk, base `01b10ca`, file hash after `97f3574a1e8f`.

````diff
diff --git a/docs/for-ai-agents.md b/docs/for-ai-agents.md
index 17624a5..82045bf 100644
--- a/docs/for-ai-agents.md
+++ b/docs/for-ai-agents.md
@@ -84,7 +84,11 @@ the same `~/.bragfile/db.sqlite` the CLI uses (see §7 to change that).
 Returns the created entry as a single JSON object with the nine standard keys:
 `id`, `title`, `description`, `tags`, `project`, `type`, `impact`,
 `created_at`, `updated_at`. A missing or empty `title` is a **tool error**,
-never a silent insert. Unlike the CLI `brag add`, `brag_add` does **not** emit
+never a silent insert. So is a key sent twice, whether as
+`"impact"` and `"impact"` or as `"impact"` and `"Impact"`: nothing is written,
+and the tool error names the key, e.g.
+`brag_add: key "impact" appears more than once`. Send each key once and retry.
+Unlike the CLI `brag add`, `brag_add` does **not** emit
 a milestone line.

 ### `brag_list` — list entries
````

### §10. `BRAG.md` (LD8)

`BRAG.md`: 1 hunk, base `01b10ca`, file hash after `9383a86b91b9`.

````diff
diff --git a/BRAG.md b/BRAG.md
index 45ee24d..b5db6df 100644
--- a/BRAG.md
+++ b/BRAG.md
@@ -206,6 +206,13 @@ The schema mirrors this guide's field table:
 - **Unknown keys are strict-rejected.** A typo like `{"titl": "x"}`
   surfaces as `unknown field "titl"` rather than silently losing
   your title.
+- **A repeated key is rejected.** `{"title":"x","impact":"A","impact":"B"}`
+  writes nothing and exits 1 with `key "impact" appears more than once`,
+  where it used to store `B` silently. Keys match case-insensitively, so
+  `"Impact"` repeats `"impact"`. A JSON library never emits a repeat, so
+  this only bites a payload built by string templating (`printf`, a
+  heredoc). A schema validator cannot catch it either: it parses the
+  payload first, and parsing keeps one value.

 Minimal valid payload:

````

### §11. `docs/brag-entry.schema.json` (LD8)

`docs/brag-entry.schema.json`: 1 hunk, base `01b10ca`, file hash after `ed2c7264fc78`.

````diff
diff --git a/docs/brag-entry.schema.json b/docs/brag-entry.schema.json
index 2e441b7..da0f593 100644
--- a/docs/brag-entry.schema.json
+++ b/docs/brag-entry.schema.json
@@ -2,7 +2,7 @@
   "$schema": "https://json-schema.org/draft/2020-12/schema",
   "$id": "https://github.com/jysf/bragfile000/blob/main/docs/brag-entry.schema.json",
   "title": "Brag entry — `brag add --json` stdin contract",
-  "description": "Single-object stdin payload accepted by `brag add --json`. Mirrors DEC-012's six locked choices in JSON Schema vocabulary; mirrors DEC-011's nine-key per-entry shape on the input side. The binary's runtime parser is the authoritative validator; this schema is the documented contract for AI agents and other external consumers to validate candidate payloads against before piping. See BRAG.md (https://github.com/jysf/bragfile000/blob/main/BRAG.md) for the integration guide.",
+  "description": "Single-object stdin payload accepted by `brag add --json`. Mirrors DEC-012's six locked choices in JSON Schema vocabulary; mirrors DEC-011's nine-key per-entry shape on the input side. The binary's runtime parser is the authoritative validator; this schema is the documented contract for AI agents and other external consumers to validate candidate payloads against before piping. The binary also rejects a repeated top-level key, compared case-insensitively (DEC-055); JSON Schema cannot express that, and a validator that parses the payload first never sees the repeat. See BRAG.md (https://github.com/jysf/bragfile000/blob/main/BRAG.md) for the integration guide.",
   "type": "object",
   "required": ["title"],
   "additionalProperties": false,
````

### §12. `CHANGELOG.md` (LD8)

`CHANGELOG.md`: 2 hunks, base `01b10ca`, file hash after `9428a0d70cef`.

````diff
diff --git a/CHANGELOG.md b/CHANGELOG.md
index 7ed8f79..bebfa75 100644
--- a/CHANGELOG.md
+++ b/CHANGELOG.md
@@ -91,6 +91,20 @@ and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0
   that means *an array of entry objects* on `brag impact`, `brag review`,
   `brag summary` and `brag wrapped`. Update any `jq .entries` to
   `jq .candidates`.
+- **Breaking: `brag add --json` and the MCP `brag_add` tool reject a
+  repeated key**
+  ([DEC-055](decisions/DEC-055-a-repeated-json-key-is-rejected-on-both-machine-ingresses.md)).
+  Both decoders kept the last of a repeated key and said nothing, so
+  `{"title":"x","impact":"A","impact":"B"}` stored `B` with exit 0, and a
+  second `"type"` could silently turn a win into a `failed` entry. Now
+  nothing is written: the CLI exits 1 with
+  `--json input: key "impact" appears more than once` on stderr, and
+  `brag_add` returns that message as a tool error. Keys are compared after
+  unescaping and case-insensitively, so `"Impact"` repeats `"impact"`, and a
+  repeat with the same value is rejected too. Only top-level keys count. A
+  payload built with a JSON library cannot contain a repeat; one built by
+  string templating (`printf`, a heredoc) can, and now fails instead of
+  storing a value its author may not have meant.

 ### Fixed

@@ -105,7 +119,8 @@ and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0
   user error (exit 1) with nothing written; the rejection is atomic. Repeating
   an *unknown* header is still ignored, as before. Reaches all three editor
   ingresses: `brag edit`, `brag add` (editor mode) and `brag learn`. The
-  `brag add --json` ingress is a separate decoder and is unchanged.
+  `brag add --json` and MCP `brag_add` ingresses are separate decoders; they
+  get the same rule under *Changed* (DEC-055).
 - **`brag edit` now prints the edited entry's id to stdout on a write**
   ([DEC-052](decisions/DEC-052-brag-edit-emits-the-mutated-id-on-stdout.md)),
   and still prints nothing on a no-op. Both outcomes previously produced empty
````

### §13. `scripts/test-docs.sh`, Group `AG` (LD8)

`scripts/test-docs.sh`: 1 hunk, base `01b10ca`, file hash after `7a82e50a2733`.

````diff
diff --git a/scripts/test-docs.sh b/scripts/test-docs.sh
index f528042..7f2e946 100755
--- a/scripts/test-docs.sh
+++ b/scripts/test-docs.sh
@@ -2328,6 +2328,42 @@ assert_section_names "AF2" "$(grep -F -- '- **learn** —' AGENTS.md)" \
 assert_section_names "AF3" "$(ad_section CHANGELOG.md '## [Unreleased]' '## [')" \
     "CHANGELOG.md, the [Unreleased] section" '`brag summary --format json` no longer lists failures'

+# ===== Group AG — a repeated JSON key is rejected (SPEC-090 / DEC-055) =====
+#
+# SPEC-090 makes `brag add --json` and MCP `brag_add` refuse a repeated
+# top-level key instead of keeping the last value. The Go suite pins the
+# behaviour on both ingresses; these ids pin the docs a caller reads, each
+# scoped with Group AD's helpers to the one section that documents that
+# ingress, so the other ingress's text cannot satisfy it.
+
+# AG1 — the contract's `add --json` block states the rule and the record.
+assert_section_names "AG1" "$(ad_section docs/api-contract.md '**STAGE-003 (JSON stdin form):**' '**STAGE-')" \
+    "docs/api-contract.md, the brag add --json block" 'appears more than once' 'DEC-055'
+
+# AG2 — the contract's `brag_add` bullet states it for the MCP tool.
+assert_section_names "AG2" "$(ad_section docs/api-contract.md '- **`brag_add`**' '- **`brag_list`**')" \
+    "docs/api-contract.md, the brag_add bullet" 'appears more than once' 'DEC-055'
+
+# AG3 — the agent-facing tool reference tells a caller what comes back.
+assert_section_names "AG3" "$(ad_section docs/for-ai-agents.md '### `brag_add`' '### ')" \
+    "docs/for-ai-agents.md, the brag_add section" 'brag_add: key "impact" appears more than once'
+
+# AG4 — BRAG.md's JSON contract, which tells a caller to validate against the
+# schema, says the schema cannot catch a repeat.
+assert_section_names "AG4" "$(ad_section BRAG.md '## JSON contract' '## ')" \
+    "BRAG.md, the JSON contract section" '**A repeated key is rejected.**'
+
+# AG5 — the unreleased changelog names the breaking change, and no longer
+# says the JSON ingress is unchanged (SPEC-089's bullet, rewritten).
+ag_unreleased=$(ad_section CHANGELOG.md '## [Unreleased]' '## [')
+assert_section_names "AG5" "$ag_unreleased" "CHANGELOG.md, the [Unreleased] section" \
+    '**Breaking: `brag add --json` and the MCP `brag_add` tool reject a' 'DEC-055'
+if printf '%s\n' "$ag_unreleased" | grep -F -q 'is a separate decoder and is unchanged'; then
+    fail "AG6" "CHANGELOG.md [Unreleased] still says the --json ingress is unchanged"
+else
+    ok "AG6"
+fi
+
 # ===== finalise =====

 if [ "$FAIL_COUNT" -gt 0 ]; then
````

### §14. `docs/engineering-practices.md`: regenerate with `just inventory`, never hand-edit

After §1–§13, `just inventory` should move exactly four rows from this design
commit's block: `Go source files` 70 → 71, `Go test files` 81 → 83,
`Go test functions` 862 → 870, and `Documentation assertions (distinct ids)`
212 → 218. Paste its output between the markers. If another row moves, `main`
has moved; trust the script, and say so in the build reflection.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `build/spec-090-json-duplicate-key`
- **PR (if applicable):** none — the orchestrator opens it.
- **All acceptance criteria met?** yes (AC-1 through AC-10), all reproduced on
  a scratch store (never `~/.bragfile/db.sqlite`) and, for AC-3/AC-4, against
  the real `brag mcp serve` binary with a held-open stdin (Traps). AC-1: exit
  1, empty stdout, exact stderr, DB hash unchanged (`bd107ab6ee48…` before and
  after). AC-2: the case-fold message names both spellings. AC-3: the real
  server's rejected-call result is byte-for-byte
  `{"content":[{"type":"text","text":"brag_add: key \"impact\" appears more than once"}],"isError":true}`,
  the DB hash held across the reject, and a following clean `brag_add` on the
  same session succeeded. AC-4: all twelve rows plus 12a/12b reproduced the
  `after` columns on both the CLI binary and the real server; the four extra
  cases (r00/r13/r14/r15) too, with the DB hash unmoved across all four
  rejections on each ingress. AC-5: the clean payload stores `REAL` on both
  ingresses; the four malformed/non-repeat payloads (`titl` typo, truncated
  object, array-wrapped, two concatenated objects) gave byte-identical stderr
  on a binary built from `main` (`95d7b39`) and the build binary. AC-6:
  `go test -count=1 ./...` is 1150 pass / 0 fail across 14 packages, the 8 new
  tests exist under their spec'd names, and `git diff --stat` on the two
  touched test files shows pure insertions (one import + one appended test
  each). AC-7: `./scripts/test-docs.sh` is 219 OK / 0 FAIL including
  `AG1`–`AG6`, and the inventory block matches `just inventory`'s output
  verbatim. AC-8: all five gates green (below). AC-9: all 18 kept probes
  reproduced exactly, hash-gated before and after each — see *Deviations* for
  the one self-inflicted extraction bug caught before any probe ran. AC-10:
  `git diff main -- decisions/DEC-055-*.md decisions/DEC-012-*.md` is empty,
  and DEC-012's text above its amendment header is byte-identical to
  `01b10ca` through line 284 (line 285 is the ordinary blank separator before
  the new heading).
- **New decisions emitted:** none. DEC-055 and DEC-012's amendment were
  written at design; this cycle only implements the rule.
- **Deviations from spec:**
  - None in the shipped diff: all 13 `git diff` blocks in §1–§13 applied via
    `git apply` with a clean `--check` both before and after, and every one of
    the 13 resulting file hashes matches the spec's stated hash exactly. No
    hunk was applied by hand.
  - One build-tooling-only note, not a spec defect: my first extraction of the
    13 diff blocks from the spec's markdown fencing used a non-greedy regex
    that swallowed the trailing blank line of §10 (`BRAG.md`), whose block
    happens to end on a blank line right before the closing fence. Caught
    immediately by validating each hunk's old/new line counts against its
    `@@` header before ever calling `git apply` — the mismatch (`old=5/6,
    new=12/13`) pointed straight at the missing line. Re-extracted by
    splitting on the fence lines' positions instead of a regex spanning them,
    and all 19 hunks (13 build blocks plus every mutation-matrix cell)
    balanced after that. The spec's own stated trap — that the repo's
    trailing-whitespace strip turns a hunk's blank context line into an empty
    line, needing a restored single space — held for 13 lines across the
    build blocks and 7 more across the mutation-matrix diffs; every one sat
    inside a hunk body, confirmed by parsing rather than by eye.
- **Follow-up work identified:** none beyond what the spec already routes
  (SPEC-097 for `--type`, SPEC-093 for the probe helper, SPEC-091/092/096, and
  the v0.7.0 release cut). The two "open, not blocking" items in *Corrections
  and open questions* (read-only MCP tools parity, an SDK bump's re-run
  obligation) are noted there and not reopened here.

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing in the spec's own mechanics — every literal, hash, and matrix
   cell was exact. The only friction was self-inflicted tooling: extracting
   markdown-fenced diff blocks programmatically needs the extraction verified
   by hunk arithmetic (old/new line counts against `@@`), not trusted by
   inspection, because a boundary bug in a regex can silently drop a line at a
   fence edge and nothing about the resulting patch looks wrong until
   `git apply` either fails elsewhere or, worse, succeeds with a shifted hunk.

2. **Was there a constraint or decision that should have been listed but
   wasn't?**
   — No. `no-sql-in-cli-layer`, `stdout-is-for-data-stderr-is-for-humans`,
   `one-spec-per-pr`, and AGENTS.md §12's mutation clauses all applied
   cleanly. Every one of the 13 build diffs and 18 probe diffs had an old side
   occurring exactly once in its base file, so nothing was ever applied by
   hand or by inference.

3. **If you did this task again, what would you do differently?**
   — Validate each embedded diff's hunk line-count arithmetic immediately
   after parsing it out of the markdown, before the first `git apply --check`,
   rather than after a hash mismatch. That check is cheap (pure arithmetic on
   the `@@` header vs. counted lines) and would have caught the §10
   extraction bug at parse time instead of at verification time.

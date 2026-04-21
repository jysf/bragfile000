---
insight:
  id: DEC-009
  type: decision
  confidence: 0.80
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-7
  session_id: null

project:
  id: PROJ-001
repo:
  id: bragfile

created_at: 2026-04-20
supersedes: null
superseded_by: null

tags:
  - cli
  - editor
  - parsing
  - format
---

# DEC-009: Editor buffer format uses `net/textproto` header + markdown body

## Decision

The editable buffer for `brag edit <id>` (and, later, `brag add`
with no args) is a plain text file with **RFC822-style headers**
parseable via Go's stdlib `net/textproto.Reader.ReadMIMEHeader`,
followed by a **blank line**, followed by a **markdown body** that
becomes the entry's `Description` field.

Example:

```
Title: cut p99 login latency
Tags: auth,perf
Project: platform
Type: shipped
Impact: unblocked mobile v3 release

Replaced the join-on-every-request with a redis lookup.
Full writeup: postmortem in doc-042.
```

Header keys are case-insensitive on read (per RFC822 / `net/textproto`
semantics) but rendered with the canonical case `Title`, `Tags`,
`Project`, `Type`, `Impact`. Empty-valued fields are omitted from the
render entirely (buffer doesn't show `Tags:` if the entry has no
tags). The body after the blank line is the full description, newlines
preserved.

## Context

`brag edit <id>` needs a buffer format that:
1. Round-trips reliably (render → edit → parse produces the same
   structured state, assuming the user doesn't change anything).
2. Is ergonomic to type — users will be editing this in vi/emacs/VS
   Code, not through a GUI form.
3. Keeps structured fields (title, tags, project, type, impact)
   distinct from free-form narrative (description).
4. Parses with a small amount of code (ideally stdlib-only) so the
   implementation surface is minimal.
5. Doesn't introduce a new top-level dependency — honors the
   `no-new-top-level-deps-without-decision` constraint at the
   weakest possible point (a YAML parser would need its own DEC).

The format choice compounds: `brag add` no-args (SPEC-010) will
reuse the same buffer shape, and a future `brag export --format
markdown` (STAGE-003) could write entries back out in this form.
Whatever we pick will be in the binary for a while.

## Alternatives Considered

- **Option A: YAML front-matter**
  ```
  ---
  title: cut p99 login latency
  tags: [auth, perf]
  ---
  body here
  ```
  - What it is: The Jekyll/Hugo convention.
  - Why rejected: Requires `gopkg.in/yaml.v3` (a new top-level dep
    needing its own DEC). YAML typing rules (lists vs scalars, quote
    escaping, indentation sensitivity) are more than the format
    needs. Users editing by hand in vi have to remember YAML's
    whitespace rules. The project is explicitly avoiding YAML
    outside of the metadata layer (`.repo-context.yaml`,
    `constraints.yaml`) which is generated/machine-read, not hand-
    edited.

- **Option B: TOML front-matter (`+++` delimiters)**
  - What it is: Hugo-style, less whitespace-sensitive than YAML.
  - Why rejected: Same dep concern (no TOML in stdlib) plus less
    familiar to users than YAML or headers.

- **Option C: Just `key=value` lines, no blank-line separator**
  - What it is: Minimal custom format.
  - Why rejected: Conflates structure with prose. How does a
    description containing `=` signs interact? Fragile.

- **Option D: Body-only free-form prose (no header block)**
  - What it is: Parse title from first `# ...` line, tags from a
    trailing `tags: ...` line, etc.
  - Why rejected: Can't cleanly round-trip structured fields.
    Every parse is ambiguous and context-sensitive.

- **Option E (chosen): `net/textproto` header + markdown body**
  - What it is: RFC822-style headers (the same mechanism email and
    HTTP headers use), terminated by a blank line, then free-form
    body.
  - Why selected:
    - **Stdlib parser.** `net/textproto.Reader.ReadMIMEHeader()`
      handles the header block in ~5 lines of caller code. No new
      dep.
    - **User-familiar.** Anyone who's written HTTP requests, email
      templates, or `git commit` trailers already knows the shape.
    - **Whitespace-tolerant.** Headers are trivial to type; no
      indentation rules. Body preserves newlines verbatim.
    - **Forgiving with `grep`.** A user can `grep '^Tags:'
      ~/.bragfile/*.md` style search if they ever export entries.
    - **Extensible.** New optional headers (e.g. a future
      `Visibility: private` or `Mood: tired`) land as new keys
      without format migration.

## Consequences

- **Positive:** Parser is ~20 lines of Go. Renderer is similar.
  Tests are fully stdlib. Users can edit the buffer in any editor
  without surprises. No new go.mod entries. Reusable for SPEC-010
  (add no-args editor) and potentially STAGE-003 markdown export.
- **Negative:** Headers can't natively express nested structure
  (e.g., multiple impact statements as a list). If we ever want
  that, either (a) a single header value uses some delimiter (the
  same way `Tags: a,b,c` does today) or (b) we migrate to YAML
  with a DEC justifying the dep. Cross that bridge when/if we
  reach it.
- **Negative:** Case-insensitive header reads mean `TAGS: foo` and
  `tags: foo` both parse. Rendering always uses the canonical form,
  so round-trips normalize silently. Worth documenting in
  user-facing help so nobody is surprised.
- **Neutral:** Commits us to the header block being visually
  separated from the body by a single blank line. Editor config
  that strips trailing blank lines on save could conflict — we'll
  need to tolerate zero or more blank lines between the header
  block and the first content line during parse.

## Validation

This decision is right if:
- `brag edit <id>` ships with a parser + renderer under ~100 lines
  of code combined (implementation sign).
- Users don't ask for YAML or TOML variants during PROJ-001.
- Round-trip (render → no-op save → parse) produces byte-identical
  or semantically-identical output for typical entries.

Revisit if:
- A header field needs structured content that comma-delimited
  strings can't represent ergonomically.
- `brag export` in STAGE-003 finds the header format awkward to
  emit (unlikely — it'll just reuse the same renderer).
- A collaboration use case emerges where users exchange buffer
  files across tools expecting a specific format (e.g. Obsidian
  with YAML front-matter).

## References

- Related specs: SPEC-009 (this format's origin), SPEC-010 (add
  no-args editor — reuses this format), future STAGE-003 export
  spec.
- Related constraints: `no-new-top-level-deps-without-decision`
  (DEC-009 is the decision that keeps `yaml.v3` out of go.mod).
- External docs:
  - https://pkg.go.dev/net/textproto#Reader.ReadMIMEHeader
  - https://datatracker.ietf.org/doc/html/rfc822 (header format)

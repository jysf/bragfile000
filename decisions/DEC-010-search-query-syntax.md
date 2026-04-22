---
insight:
  id: DEC-010
  type: decision
  confidence: 0.85
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

created_at: 2026-04-22
supersedes: null
superseded_by: null

tags:
  - cli
  - search
  - fts5
---

# DEC-010: `brag search` auto-tokenizes + phrase-quotes user input

## Decision

`brag search "query"` takes a **single positional argument** from
the CLI, splits it on whitespace into tokens, wraps each token in
FTS5 phrase-quotes, joins them with a single space, and passes the
result to `MATCH` in `Store.Search`. Empty, whitespace-only, or
quote-containing queries return `ErrUser` from the CLI before
reaching storage.

Example transformations:

| User types | argv | FTS5 MATCH argument |
|---|---|---|
| `brag search latency` | `["latency"]` | `"latency"` |
| `brag search "cut latency"` | `["cut latency"]` | `"cut" "latency"` |
| `brag search auth-refactor` | `["auth-refactor"]` | `"auth-refactor"` |
| `brag search ""` | `[""]` | ErrUser — empty query |
| `brag search 'with "quote"'` | `["with \"quote\""]` | ErrUser — quote in query |

## Context

SPEC-011's build session empirically discovered that FTS5's `-`
operator is **binary NOT**, so a raw `MATCH 'auth-refactor'` parses
as `auth NOT refactor` and returns nothing. A user typing
`brag search auth-refactor` expects to find their entries about
auth refactoring, not the NOT semantics. Similar surprises lurk
with `*` (prefix), `^` (column filter), `AND`/`OR`/`NOT` as
literals, and parentheses.

FTS5 is a powerful query language, but it's a query language — not
what users expect from a `grep`-like command. Most CLI search tools
(`grep`, `ripgrep`, `fzf`, `ack`) treat user input as a literal
pattern, not a mini-language.

SPEC-011 ship reflection flagged the choice as an open design
question for SPEC-012. This DEC records the answer.

## Alternatives Considered

- **Option A: Expose raw FTS5 syntax**
  - What it is: pass the user's query to `MATCH` verbatim.
  - Why rejected: user types `auth-refactor` → gets zero results
    silently (`auth NOT refactor`). Every CLI tool that exposes
    a query DSL acquires a FAQ entry about this. Power users get
    NEAR/OR/column-filter for free; everyone else is confused.

- **Option B: Auto-quote the entire query as a single phrase**
  - What it is: wrap the whole argv-single-arg in `"..."` →
    `MATCH '"cut latency"'` → phrase search requiring the exact
    sequence of adjacent tokens.
  - Why rejected: too restrictive for multi-word queries. Users
    typing `brag search cut latency` expect "find entries with
    both words somewhere" (AND), not "find entries with the
    exact phrase `cut latency`."

- **Option C (chosen): Tokenize + per-token phrase-quote + AND-join**
  - What it is: `strings.Fields(query)` → wrap each non-empty
    token in `"..."` → `strings.Join(tokens, " ")` → pass to
    `MATCH`. FTS5's default AND semantics across space-separated
    phrases gives "find entries containing all of these terms."
  - Why selected:
    - Hyphens, asterisks, and other FTS5 operators inside a
      phrase-quoted token are treated as literal text, not
      operators. No NOT-operator surprise.
    - Multi-word queries get sensible AND semantics per user
      intuition.
    - FTS5 still tokenizes the content inside each quoted phrase
      using the default unicode61 tokenizer, so
      `"auth-refactor"` matches rows where `auth` and `refactor`
      are adjacent tokens (which is how the indexed content
      parses anyway).
    - Zero new dependencies; ~5 lines of Go.

- **Option D: Hybrid — detect "simple" vs "FTS5 syntax" queries**
  - What it is: if the query contains no `"`, `*`, `AND`, `OR`,
    `NOT`, or parentheses, auto-quote; otherwise pass raw.
  - Why rejected: detection heuristic has edge cases (user has a
    legit entry about "OR" programming logic; user types a
    parenthesized phrase naturally). Two semantics for one
    command surprises users when their query accidentally
    crosses the threshold. Save the power-user path for an
    explicit `--raw` / `--fts5` flag in a future polish spec if
    demand materializes.

## Consequences

- **Positive:** 95%+ of user queries "just work." Hyphens, asterisks,
  common punctuation in titles (e.g., `SPEC-011`, `p99`, `CI/CD`)
  become first-class search terms. Multi-word queries get the
  intuitive AND semantics. No new deps.
- **Negative:** Power users can't use FTS5's advanced operators
  (NEAR, explicit NOT, OR, column filter, prefix matching `foo*`)
  through `brag search`. Acceptable for MVP; add `--raw` flag in a
  future polish spec if anyone asks.
- **Negative:** Quote characters inside the query fail with
  `ErrUser`. Users who want to search for a literal `"` (rare —
  entry titles usually don't quote) are blocked. Could escape
  FTS5-style (double the `"`) in a future enhancement; reject is
  simpler and safer for MVP.
- **Neutral:** Empty-string queries are rejected. A user intent to
  "list everything" is served by `brag list`, not `brag search`.

## Validation

Right if:
- `brag search <word>` always finds rows containing that word,
  regardless of what other punctuation surrounds it in either the
  query or the indexed content.
- `brag search "multi word"` finds rows containing both words.
- Users never ask "why did my search return zero results?"
  referencing a hyphen or special character.
- The parity check: `brag list --tag auth` and `brag search auth`
  both find rows with `auth` in tags. (Both do, under this
  decision — SPEC-007's sentinel-comma tag filter + SPEC-011's
  FTS5 unicode61 tokenizer agree that commas are word separators.)

Revisit if:
- Users demand FTS5 operators in search — add `--raw` (or
  `--advanced`) flag as an escape hatch, documented explicitly
  as "passes query through to FTS5 MATCH verbatim; you own the
  syntax."
- FTS5 tokenizer ever changes from default unicode61 (e.g.,
  adding porter stemmer for typo tolerance) — the quoting
  behavior still holds, but match semantics shift.
- Search performance degrades on large DBs such that a smarter
  query optimizer is warranted. Unlikely at personal-use scale.

## References

- Related specs: SPEC-012 (this DEC's first consumer), SPEC-011
  (whose build surfaced the hyphen-as-NOT concern).
- Related constraints:
  `stdout-is-for-data-stderr-is-for-humans` (search output goes
  to stdout; any ErrUser messages go to stderr via main.go).
- External docs:
  [SQLite FTS5 query syntax](https://sqlite.org/fts5.html#full_text_query_syntax)
  — §3 covers the phrase-quoting semantics; §4 covers operators.

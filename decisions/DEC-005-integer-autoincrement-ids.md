---
insight:
  id: DEC-005
  type: decision
  confidence: 0.70
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

created_at: 2026-04-19
supersedes: null
superseded_by: null

tags:
  - data-model
  - mvp-tradeoff
---

# DEC-005: Entries use `INTEGER PRIMARY KEY AUTOINCREMENT` IDs for MVP

## Decision

`entries.id` is `INTEGER PRIMARY KEY AUTOINCREMENT`. Printed as a
decimal number (`42`). Used as the argument to `brag show`, `edit`,
`delete`.

## Context

Every entry needs a stable, user-visible identifier. Two serious
options: short monotonic integers (SQLite's `ROWID`/autoincrement) or
opaque sortable IDs (ULID/KSUID/nanoid).

The MVP is a single-user tool with one database per user. The ID is
meaningful only to that user on that machine. There is no sharing, no
merge-between-dbs, no URL surface — all reasons that normally push
toward opaque IDs do not apply yet.

Confidence is 0.70: we are making the ergonomic choice knowing the
portability choice exists.

## Alternatives Considered

- **Option A: ULID (26-char, sortable by time)**
  - Why rejected (for MVP): `brag show 01HFXY... `is a worse daily
    ergonomic than `brag show 42`. No current need for global
    uniqueness or database merging. Would be the right call if
    cross-machine sync or shareable URLs ever become a feature.

- **Option B: UUIDv4**
  - Why rejected: All the costs of ULID (ugly long IDs) with none of
    the sort-by-time benefit.

- **Option C: Nanoid**
  - Why rejected: Shorter than ULID but still worse than an integer
    for CLI typing, with no current offsetting benefit.

- **Option D (chosen): `INTEGER PRIMARY KEY AUTOINCREMENT`**
  - Why selected: Trivial to type, short, monotonic, zero extra code.
    Exactly matches how the user reads entries back out of `brag list`.

## Consequences

- **Positive:** Minimal friction. `brag list` output already contains
  the IDs the user will re-type. Fits the 10-second-capture aesthetic.
- **Negative:** IDs are meaningless outside one user's DB. If we ever
  want cross-machine sync, sharing, or import/export that preserves
  IDs across databases, we'll need a stable secondary identifier
  (likely a ULID column added in a migration).
- **Neutral:** `AUTOINCREMENT` (as distinct from plain `INTEGER PRIMARY
  KEY`) guarantees monotonically increasing IDs across deletes. Small
  overhead, avoids the surprise of a reused ID.

## Validation

Right if:
- No user confusion from IDs that "look like a row number".
- No feature in PROJ-001 requires IDs to be portable.

Revisit if:
- Sync, sharing, or multi-db export enters scope.
- Users ask for shareable short-links to entries.

## References

- Related specs: SPEC-002 (initial schema)
- Related docs: `./docs/data-model.md`

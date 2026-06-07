# Premise-audit sub-template

A reusable design-cycle checklist distilled from AGENTS.md §9. A spec
**references** this file from its `## Outputs` section rather than
re-deriving the cases each time. Extracted at STAGE-006 framing
(2026-06-06) after three deferrals (SPEC-015 ship, STAGE-004 ship,
STAGE-005 framing); first applied by SPEC-025.

> **What a premise audit is.** When a spec changes existing behavior,
> some existing tests, counts, or docs rest on the *old* premise and
> will silently break (or silently keep passing when they shouldn't).
> The audit enumerates those at **design** time and lists each as a
> planned `## Outputs` item, so build transcribes a plan instead of
> discovering breakage. Enumeration without execution is aspirational —
> **run every grep below at design and reconcile its hits against your
> enumerated Outputs.**

## The three cases

Walk each locked design decision against the repo. For each, ask which
case(s) it triggers, then add the result to `## Outputs`.

### 1. Inversion / removal → planned test deletion or rewrite

The decision **inverts or removes** existing behavior (a new default, a
replaced algorithm, a dropped flag/column).

- Enumerate every existing test whose *premise* the change invalidates.
- List them under `## Outputs` as planned deletions / rewrites — not as
  build-time discoveries disclosed under Deviations.
- Grep the affected files for tests asserting the old behavior.

*Origin: SPEC-010 (no-flags → editor mode inverted `TestAdd_MissingTitleIsUserError`).*

### 2. Addition → planned count-bump

The decision **adds to a tracked collection** whose size is asserted
somewhere: migrations, DECs, constraints, queued entries, any
fixed-count structure.

- `grep -rn "<collection>" <test globs>` and audit each hit for literal
  count coupling (e.g. `count == N`, an exact-list assertion).
- List every coupled assertion under `## Outputs` as a planned bump.

*Origin: SPEC-011 (`0002_add_fts.sql` broke `TestOpen_MigrationsTracked`'s "exactly 1 row").*
*Concrete trigger: anything added to `schema_migrations`, the DECs list, or `constraints.yaml`.*

### 3. Status change → planned doc-references update

The decision **changes a feature's shipping status** (strikes a
"not-yet" row, introduces/removes a command, realizes a future-work
section).

- `grep -rn "<feature-name>" docs/ README.md` and audit each hit for a
  status claim — including *secondary* claims (scope blurbs, intro
  lines), not just the primary one.
- List every hit under `## Outputs` as "updates" or "stays here."

*Origin: SPEC-012 (struck `search` from the primary status table but missed the §-3 scope blurb).*

## The audit-grep cross-check (both sides)

A grep written in the spec but never run is decorative.

- **Design side:** RUN each grep above against the repo and reconcile
  actual hits against your enumerated `## Outputs`. If a grep surfaces a
  hit the enumeration missed, add it before locking the spec.
- **Build side:** before the doc/test sweep, RE-RUN the spec's audit
  greps. Treat any delta between actual hits and the spec's enumerated
  Outputs as a **question for the spec author** (raise in build
  reflection or a clarifying check) — not a unilateral scope expansion.

*Origin: SPEC-018 (two greps enumerated with expected hits, neither run at design; both would have caught real misses).*

> **macOS grep caveat (§9).** BSD grep matches basenames, not path
> fragments — `--exclude-dir=docs/reports` is a silent no-op. Bound
> tolerable hits with an explicit whitelist / `case` post-filter, not
> with `--exclude-dir`.

## Checklist to paste into a spec's `## Outputs`

```
Premise audit (projects/_templates/premise-audit.md), run at design:
- [ ] Inversion/removal: greps run, invalidated tests listed as planned rewrites/deletions
- [ ] Addition/count-bump: greps run, literal-count assertions listed as planned bumps
- [ ] Status-change: greps run, every doc hit listed as updates/stays
- [ ] Cross-check: actual grep hits reconciled against the lists above
```

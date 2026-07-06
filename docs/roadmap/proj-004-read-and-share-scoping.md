# PROJ-004 (candidate) — "read & share" scoping note

> **Status: SCOPING / pre-brief — NOT framed.** This is a candidate-scope
> stash, not a project brief. PROJ-004's real brief is written at framing,
> **after the DuckDB federation spike** returns (a separate session). `PROJ-004`
> is the next free repo-global id (§2); the slug `read-and-share` is
> provisional. Captured 2026-07-05 so the scope survives the PROJ-003 close.

## The reframe

The multi-user "federation" question and the long-deferred single-user
"impact read surface" are **the same theme: read & share.** Federation is
plumbing; what you federate *for* is reading and distributing the corpus.
And a read/output surface is the **original north-star of bragfile** —
"capture accomplishments for retros, reviews, and resumes" (AGENTS.md §1) —
which has been deferred through PROJ-001/002/003. v0.3.0 built the
agent-native *write* spine and made provenance measurable (`brag list
--author`); PROJ-004 is the natural *read & share* payoff.

## Candidate features (rated by fit)

Grounded in already-deferred work: the PROJ-003 brief's out-of-scope list,
the cross-project retro action-register (P1/P3), and the PROJ-001 backlog
storytelling cluster.

| Feature | Why it fits | Fit |
|---|---|---|
| `brag impact --quarter\|--month\|--year\|--since` (rule-based impact digest, retro **P1**) | The read foundation; the warehouse's cross-source queries reuse the same aggregation | **Core** |
| Provenance-share / dogfooding query (retro **P3**, drafted as SPEC-045) | "% agent- vs human-authored over time" — now measurable via `--author` | **Core** |
| Storytelling: `brag story --audience promo\|resume\|blog\|1:1` | Same corpus → different stories; the brief's named "differentiator" and the original review/resume/promo payoff | **Core** |
| AI-pipe "super-brag" (impact bundle + synthesis; atoms→molecules clustering) | Quarterly synthesis — the backlog "headline" | **Core** |
| `brag wrapped [year]`, activity sparkline, impact density, `brag achievements` | The insight / shareable cluster | **Core** |
| Federation: export-with-source (`host:`/`user:`) → DuckDB warehouse → pull-in | The multi-user half; **shape decided by the DuckDB spike** | **Core** |
| Output adapters (Notion; resume / review / promo-packet generators) | The "share" endpoint — turn the corpus into hand-off artifacts | **Core** |
| First-class `agent`/`model` columns (the DEC-004→DEC-015 "later, if earned" promotion) | v0.3.0 makes provenance flow *and* measurable — the revisit trigger may now fire | **Core-ish** (earned-by-data) |
| WAL + busy-timeout concurrency hardening | Needed **only if** federation is a *shared-write* DB (or heavy multi-agent). Export-based aggregation → **not needed** | **Conditional** (on the spike's architecture) |
| Goals as an object type (`brag goals add/list`, `brag list --goal`) | Synergy with impact ("did I hit my goal?") but a *planning/forward* object — a different shape | **Separate** (own project, ~PROJ-005) |
| macOS notarization | Pure distribution/ops (v0.2.1 track) — unrelated to read/share | **Separate** |

## Proposed shape (to confirm at framing)

Two stages, sequenced so the second reuses the first:

1. **Single-user read/story surface** (`brag impact` → storytelling → wrapped
   / adapters) — independently valuable and shippable as **v0.4.0** (additive
   read commands = a *minor* bump, not a v0.3.1 patch). This is the
   aggregation logic the warehouse later unions.
2. **Federation / warehouse** (export-with-source → DuckDB → pull-in) —
   turns those same queries multi-source (v0.4.x / v0.5.0).

Alternatively split into two projects (ship the read surface, pause before
federation). Either way the read surface comes first.

## Open decisions (resolve at framing / post-spike)

- **Federation architecture** (from the DuckDB spike): federated **export →
  DuckDB warehouse, preserving local-first** (leading hypothesis; additive,
  keeps DEC-001) **vs** a shared remote DB (rejected — supersedes DEC-001) **vs**
  local-first sync (libsql/Turso — only if real-time *shared write* is a real
  need). This choice determines whether concurrency hardening is in-scope.
- **Identity dimension:** stamp `host:`/`user:` at export time (label at the
  boundary — don't pollute every row) — settle the exact convention.
- **First-class provenance columns:** run the v0.3.x dogfooding first; promote
  only if provenance filtering/reporting becomes a real ask or `agent:`/`model:`
  tags visibly pollute the `brag tags` taxonomy (the stated revisit trigger).
- **Goals**: keep out of PROJ-004 (separate project) unless framing decides the
  goals×impact synergy is worth bundling.

## Sources

- [`projects/PROJ-003-agent-native-spine/brief.md`](../../projects/PROJ-003-agent-native-spine/brief.md) — Scope (out-of-scope list), Project-Level Reflection (deferred items).
- [`docs/reports/cross-project/2026-07-04-action-register.md`](../reports/cross-project/2026-07-04-action-register.md) — P1 (impact digest), P3 (dogfooding query).
- [`projects/PROJ-001-mvp/backlog.md`](../../projects/PROJ-001-mvp/backlog.md) — the impact-surfacing / storytelling cluster.

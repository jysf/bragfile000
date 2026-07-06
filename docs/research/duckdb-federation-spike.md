# Findings — DuckDB federation spike (PROJ-004 pre-framing)

*Throwaway, time-boxed spike. Run 2026-07-05 by Claude (opus-4-8) against the
real corpus (`~/.bragfile/db.sqlite`, treated read-only — prod never mutated;
md5 verified unchanged before/after). Scripts live in the session scratchpad,
not the repo; only this report is durable. Goal: validate the leading PROJ-004
hypothesis — **federated daily export → DuckDB warehouse, preserving
local-first** (each person/machine keeps their own `~/.bragfile`; a warehouse
unions everyone's exports) — before the framing session.*

> **Doc-path note.** The brief pointed at
> `docs/roadmap/proj-004-read-and-share-scoping.md` as "read this first"; that
> file does not exist in the repo (there is no `docs/roadmap/` directory). The
> spike proceeded from the brief's inline context. Worth reconciling the path
> at framing — either the roadmap doc is unwritten or it lives elsewhere.

---

## TL;DR — Verdict

**Federated export → DuckDB is viable and, once the data lands in native
DuckDB tables, genuinely pleasant.** Every candidate warehouse query — time
series, project/type rollups, streaks, provenance share, tag taxonomy, impact
density, and all their cross-source/whole-org variants — was clean SQL, mostly
one short statement each. DuckDB's `FILTER (WHERE …)`, `string_agg`, `unnest`,
and window functions make the multi-source read surface expressive.

**But the naive "`ATTACH` every `~/.bragfile` and `UNION` live" path is a
trap** and must not be the architecture. Two hard blockers surfaced
immediately (details below):

1. **Raw bragfile DBs don't attach** — DuckDB's SQLite reader parses all of
   `sqlite_master`, and the FTS5 trigger DDL crashes its parser.
2. **Same-schema multi-attach silently returns wrong data** — with three
   `~/.bragfile` files attached at once, every source alias resolved to the
   *first* attached DB. Counts looked plausible and were wrong.

Both are dodged by the same design: **each source materializes independently
(one attach at a time, or — better — exports to Parquet), stamped with its
identity, and the warehouse unions the materialized copies.** That is also the
more sensible warehouse anyway (daily export → durable columnar store).

The one thing the export path *must* add for federation is a **file-level
source/identity dimension** (a stamped column / export manifest — **not** a
per-entry tag). And PROJ-004 has to confront that **bragfile has no global
entry identity**: local `id`s are per-file autoincrement, so dedup across a
person's machines needs a content key or a capture-time UUID.

---

## Method (what was actually run)

- DuckDB v1.5.4 (`brew install duckdb`), SQLite reader via
  `INSTALL sqlite; LOAD sqlite;`.
- Real corpus snapshot (read-only `cp`): **199 entries** after simulation
  seed; source corpus = 195 entries, 226 tags, 620 taggings, 5 projects,
  span 2026-04-20 … 2026-07-06. **5 entries are agent-authored**
  (`agent:claude-code` + `model:claude-opus-4-8`) — the brief expected 0; the
  corpus has moved since. Good: the provenance query now has real signal.
- **Multi-source simulation** (only one real corpus exists): partitioned the
  195 entries disjointly by `id % 3` into three separate SQLite files —
  `alice@laptop` (65), `bob@desktop` (65), `alice@agent-box` (65) — mimicking
  three people's separate `~/.bragfile`s. Then injected a deliberate
  **cross-machine overlap**: copied 4 of alice@laptop's entries into
  alice@agent-box (same content, new local id) to simulate the *same brag
  synced across one person's two machines* → agent-box = 69, union = 199.
- **Both label mechanisms tried**: (A) source stamped as a column at
  materialize/export time; (B) identity carried as reserved `user:`/`host:`
  tags and recovered by parsing.

---

## Surprises (the two blockers, in detail)

### Blocker 1 — raw bragfile DBs fail to `ATTACH`

```
ATTACH 'source-me.db' AS db (TYPE sqlite);
-- Invalid Error: Failed to prepare query "SELECT type FROM sqlite_master …":
--   malformed database schema (entries_ai) - near "ORDER": syntax error
```

DuckDB's SQLite extension reads the whole `sqlite_master` catalog on attach.
The FTS5 maintenance trigger `entries_ai` contains
`GROUP_CONCAT(t.name, ',' ORDER BY tg.position)` — valid SQLite, but DuckDB's
embedded SQLite-compat parser rejects the `ORDER BY` inside the aggregate. The
FTS5 virtual table (`entries_fts`) and its six triggers are dead weight for a
warehouse anyway. **Workaround:** drop `entries_fts` + all six triggers from
the export copy before it reaches DuckDB. This is a concrete requirement on
the export path (see below) — a bare `VACUUM INTO` / `brag export --sqlite`
file is **not** DuckDB-attachable as-is.

### Blocker 2 — same-schema multi-attach conflates sources (silent, worse)

Attaching three `~/.bragfile` files with identical table names in one session
and counting each by fully-qualified name:

```sql
ATTACH 'alice-laptop.db' AS s1 (TYPE sqlite, READ_ONLY);
ATTACH 'bob-desktop.db'  AS s2 (TYPE sqlite, READ_ONLY);
ATTACH 'alice-agent.db'  AS s3 (TYPE sqlite, READ_ONLY);
SELECT count(*), max(id) FROM s1.entries;   -- ground truth 65 / 201
SELECT count(*), max(id) FROM s2.entries;   -- ground truth 65 / 202
SELECT count(*), max(id) FROM s3.entries;   -- ground truth 69 / 207
```

| alias | DuckDB returned | sqlite3 ground truth |
|---|---|---|
| s1 (laptop) | 65 / max 201 | 65 / max 201 |
| s2 (desktop) | **65 / max 201** | 65 / max 202 |
| s3 (agent) | **65 / max 201** | 69 / max 207 |

Every alias returned the **first-attached** DB's data. The union looked
healthy (3 × 65 = 195, distinct-looking) and was wrong — agent-box's 69 rows
and the overlap were invisible. This is the dangerous kind of bug: no error,
just silently incorrect rollups. It is specific to multiple attachments that
share table names — exactly the federation case.

**Fix (verified):** attach **one source at a time**, materialize into a native
DuckDB table with the source stamped, `DETACH`, repeat; then union the native
tables. Correct counts (65 / 65 / 69 = 199, distinct maxids) returned
immediately.

```sql
ATTACH 'alice-laptop.db' AS s (TYPE sqlite, READ_ONLY);
CREATE TABLE m_laptop AS SELECT *, 'alice@laptop' AS source FROM s.entries;
DETACH s;   -- repeat per source, then UNION ALL the m_* tables
```

---

## The query surface — clean vs painful

**All candidate queries were clean.** Nothing in the target set was painful
*once the data was in native DuckDB tables*. The pain was entirely in the
attach/ingest layer (blockers above), not the read layer. Representative SQL
that worked (full set + captured output in the scratchpad `query-results.txt`):

**Cross-source time series** — the `FILTER` pivot is the workhorse; whole-org
and per-source come from one statement:

```sql
SELECT strftime(local_day, '%Y-%m') AS month,
       count(*) FILTER (WHERE source='alice@laptop')   AS alice_laptop,
       count(*) FILTER (WHERE source='bob@desktop')     AS bob_desktop,
       count(*) FILTER (WHERE source='alice@agent-box') AS alice_agent,
       count(*)                                          AS whole_org
FROM wh_entries_local GROUP BY month ORDER BY month;
```
```
 month   │ alice_laptop │ bob_desktop │ alice_agent │ whole_org
 2026-04 │ 13           │ 12          │ 16          │ 41
 2026-06 │ 40           │ 41          │ 40          │ 121
 2026-07 │ 10           │ 10          │ 11          │ 31
```

**Streak by LOCAL day (DEC-022 semantics)** — gaps-and-islands is a two-line
idiom in DuckDB; local-day bucketing is `derive-local, store-UTC` exactly per
DEC-022 (`CAST(timezone('America/Los_Angeles', created_at::TIMESTAMPTZ) AS DATE)`):

```sql
WITH days AS (SELECT DISTINCT local_day FROM wh_entries_local),
islands AS (SELECT local_day,
              local_day - CAST(row_number() OVER (ORDER BY local_day) AS INTEGER) AS grp
            FROM days)
SELECT max(run) AS longest_streak_days FROM (SELECT count(*) run FROM islands GROUP BY grp);
-- whole-org longest streak = 13 days
```

**Provenance share (agent vs human) over time** — reuses the reserved-tag
predicate; agent authorship shows up only in 2026-07 (16.1% that month),
matching the 5 agent entries:

```sql
CASE WHEN tags LIKE 'agent:%' OR tags LIKE '%,agent:%'
       OR tags LIKE 'model:%' OR tags LIKE '%,model:%' THEN 'agent' ELSE 'human' END
```

**Impact density**, **by-project / by-type rollups**, **tag taxonomy top-N**
(via `unnest(string_split(tags, ','))` into a long-form `wh_tags` view) all
worked as single statements with per-user/whole-org breakdowns. Whole-org
impact density = **65.8%** (131/199).

**One real friction in the read layer — the reserved-tag family.** The tag
taxonomy top-N and any `string_agg` tag projection must exclude the entire
reserved namespace — `agent:`, `model:`, and (if Mechanism B is used)
`user:`/`host:` — or identity/provenance tags pollute the taxonomy. This is a
mild but pervasive `NOT LIKE 'user:%' AND NOT LIKE 'host:%' AND …` that rides
on every tag rollup. It is an argument *against* carrying identity as tags.

---

## Identity, attribution & dedup (the genuinely hard part)

This is where federation gets real, and where bragfile's current schema is
under-specified for it.

**bragfile has no global entry identity.** `entries.id` is per-file
`AUTOINCREMENT` (DEC-005) — every `~/.bragfile` reuses `1, 2, 3, …`, so `id`
means nothing across sources and cannot be a join/dedup key.

**Overlap double-counts, and dedup must be identity-scoped.** The 4 synced
entries (alice's brag on both her laptop and her agent-box) inflate the naive
union 195 → 199. Collapsing them requires a **(user + content-hash)** key:

```sql
-- naive 199 → deduped 195
SELECT count(*) FILTER (WHERE rn=1) AS deduped_total FROM (
  SELECT row_number() OVER (
    PARTITION BY wh_user,
                 md5(coalesce(title,'')||'|'||coalesce(created_at,'')||'|'||coalesce(description,''))
    ORDER BY source) AS rn
  FROM wh_entries);
```

Two subtleties the probe surfaced:

- **Dedup must be scoped to the same person.** A content key alone is unsafe:
  the probe found a `(title, created_at)` pair shared across *different* users.
  Bob legitimately shipping "Fixed the flaky test" the same day as Alice must
  **not** collapse into one org-level row. Dedup key = `wh_user` +
  content-hash, never content alone, never `id`.
- **Content-hash is a heuristic, not a guarantee.** If Alice `edit`s the entry
  on one machine after syncing, `updated_at`/`description` diverge and the hash
  splits — the warehouse sees two rows. A capture-time stable UUID would make
  this deterministic; a content-hash is the best available today.

---

## Label mechanism: stamp-at-export (A) beats identity-as-tag (B)

The brief asked which way the source/identity dimension should attach. Both
were built and compared.

| | **A — stamp at export (file-level)** | **B — `user:`/`host:` tags (per-entry)** |
|---|---|---|
| Coverage | Every row in the export inherits identity uniformly | Only rows that *carry* the tag; any entry missing it → **NULL source** |
| Taxonomy | Untouched | Pollutes tags; needs `NOT LIKE` filters everywhere |
| Recovery | Plain column | Per-row `regexp_extract` out of a comma-joined blob |
| Bloat | One value per file | +2 tags × N entries in tags/taggings (here +390 taggings) |
| Mutation risk | None | `brag tags rename/merge` (DEC-016) could corrupt identity |

Mechanism B failed concretely in the probe: the 4 synced entries that arrived
**without** their `taggings` came back as `user=NULL, host=NULL` — unknown
source. Identity is a property of *where an export came from*, not of each
entry, so it belongs on the **file/export**, not on the rows.

```
-- Mechanism B, recovering identity by parsing tags:
 wh_user │ wh_host   │ n
 alice   │ agent-box │ 65
 NULL    │ NULL      │ 4     ← the synced entries lost their source
```

**Recommendation: Mechanism A.** Stamp identity at the file/export level.

---

## Latency: daily export is fine; real-time is not the question

Federation here is a **read/reporting** surface (retros, org rollups), not an
operational store. A daily (or on-demand) `export → warehouse rebuild` is
well-matched: the warehouse is derived, disposable, and idempotent — rebuild
from scratch each run, no incremental-merge machinery, no CDC. The only
latency that matters is "how stale is the org rollup," and a day is fine for
retro/review cadence. Real-time sync would *reintroduce* the shared-remote-DB
coupling that DEC-001 (local-first) exists to avoid. Keep exports as flat
snapshots; let the warehouse be a pure function of them.

---

## What the bragfile export path must add for federation

Concrete, minimal shape:

1. **A warehouse-friendly export** that is DuckDB-attachable — i.e. **strip
   `entries_fts` + its six triggers** from the emitted file (Blocker 1), or
   export the base tables (`entries`, `tags`, `taggings`, `projects`) to
   Parquet directly. A new `brag export --format duckdb|parquet` (or a
   `--warehouse` flag on the sqlite export) is the natural home.
2. **A file-level source/identity stamp** (Mechanism A) — user + host (+
   machine role), written once into an export manifest or a stamped column,
   not as per-entry tags. This is *the* thing the current export path lacks.
3. **A dedup key.** Short-term: document that the warehouse dedups on
   `(user, content-hash)`. Longer-term, if federation ships: consider a
   **capture-time UUID** on `entries` so cross-machine dedup is deterministic
   rather than heuristic. (Schema change — its own DEC.)
4. **Reserved-namespace hygiene.** The warehouse's tag rollups must exclude the
   reserved tag family (`agent:`, `model:`, and — only if B were ever used —
   `user:`/`host:`). Since we're choosing A, `user:`/`host:` never enter the
   tag stream, which is one more reason to prefer A.

---

## Recommendation for the PROJ-004 framing

**Pursue federated export → DuckDB. It validates the local-first federation
hypothesis: each person keeps their own `~/.bragfile`; a derived warehouse
unions stamped exports; the read surface is excellent.** Shape it as:

- **Ingest = materialize-per-source, never live multi-attach.** Each source
  exports a warehouse-friendly snapshot (FTS stripped, identity-stamped);
  the warehouse loads them one at a time into native DuckDB tables (or reads
  per-source Parquet) and unions. This sidesteps both blockers by design.
- **Warehouse is derived & disposable** — rebuilt from the day's exports,
  a pure function of the snapshots. No incremental merge, no CDC, no
  shared remote DB (DEC-001 preserved).
- **Identity is file-level (Mechanism A).** Add the source/identity stamp to
  the export path — this is the single required change to bragfile itself for
  a *minimal* federation MVP.
- **Name the identity gap explicitly in the frame.** "No global entry
  identity" is the load-bearing open question. A content-hash dedup unblocks
  an MVP; a capture-time UUID is the durable fix and deserves a decision.

**Biggest surprises to carry into the frame:** (1) raw bragfile DBs are *not*
DuckDB-attachable (FTS trigger DDL); (2) same-schema multi-attach returns
**silently wrong** data — the single most important reason the architecture
must be materialize-then-union, not attach-and-union.

---

## Appendix — repro (scratch only; not committed)

Scripts in the session scratchpad (`scratchpad/duckdb-spike/`):
`build-warehouse.sql` (materialize-per-source union), `queries.sql`
(candidate query set), `query-results.txt` / `dedup-results.txt` (captured
output). The three simulated source DBs and the `warehouse.duckdb` are
throwaway. Nothing here touches `~/.bragfile` (prod md5 verified unchanged).

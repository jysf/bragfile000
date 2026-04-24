---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-013
  type: decision
  confidence: 0.82
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

created_at: 2026-04-23
supersedes: null
superseded_by: null

tags:
  - cli
  - io-contract
  - markdown
  - export
  - human-consumer
---

# DEC-013: Markdown export shape — provenance + summary + grouped entries

## Decision

`brag export --format markdown` emits a review-ready document with six
locked shape choices:

1. **Document structure** is fixed. Top of document is a `# Bragfile
   Export` level-1 heading, followed by a provenance block
   (`Exported: <RFC3339>`, `Entries: <N>`, `Filters: <echoed flags or
   "(none)">`), then an executive summary block with `**By type**` and
   `**By project**` bulleted counts.
2. **Default grouping is by project**, `## <project name>` level-2
   sections in alphabetical-ASC order. Entries without a project are
   grouped under `## (no project)` and rendered **last** regardless of
   count.
3. **Within each group, entries render chronologically ASCENDING** by
   `created_at` (not DESC — review docs read forward through time, so
   the oldest entry of the period appears first).
4. **Entries render via the lifted `export.RenderEntry` helper**, with
   title at `### <title>` (level 3, nested under the level-2 project
   heading) and description at `#### Description` (level 4). Consecutive
   entries **within the same group** are separated by `---`; transitions
   between groups rely on the `## <project>` heading as the visual
   separator (no `---` between groups).
5. **`--flat` flag skips grouping** and renders all entries
   chronologically ASC under a single `## Entries (chronological)`
   wrapper section. Entry heading level stays at `###` (consistent with
   grouped mode). `---` separators between every pair of consecutive
   entries.
6. **`--out <path>` writes to the file** instead of stdout. Matches
   SPEC-014's `--out` semantics: overwrite existing files without
   prompt; directory or unwritable path returns an internal error.
   Absent `--out` writes to stdout.

Summary-block ordering: within `**By type**` entries sort DESC by
count with alphabetical-ASC tie-break. Within `**By project**` the
same rule applies, except `(no project)` is forced last regardless
of count (matching choice 2). When the entry set is empty (`Entries:
0`), the summary block and groups section are both omitted — the
document ends with the provenance block.

## Context

STAGE-003 ships three exports that share filter semantics but serve
different consumers: `--format json` (AI / machine — DEC-011,
SPEC-014), `--format markdown` (human review prep — this DEC,
SPEC-015), and a deferred `--format sqlite` (backup / portable copy —
scoped out 2026-04-23, `cp ~/.bragfile/db.sqlite` handles the use
case). The markdown export is the form users will paste into retro
docs, self-review writeups, and promo packets; the shape choices are
load-bearing for those workflows and need to be documented once,
cross-referenced from the spec and the doc sweep.

Six choices needed locking; each is stated above.

## Alternatives Considered

- **Option A: Chronological DESC within groups.**
  - What it is: Mirror `brag list`'s `created_at DESC` ordering so the
    most recent entry of a period appears first in each group.
  - Why rejected: Review docs read forward through time. A reader
    opening a quarterly export wants to start at the beginning of the
    quarter and follow the narrative to the end. DESC-within-group
    would force readers to scroll to the bottom of each group to find
    the starting point. `brag list`'s DESC is correct for "what
    happened recently"; the export's ASC is correct for "here's the
    story of the period."

- **Option B: Table of contents block after the summary.**
  - What it is: Auto-generate a `## Table of Contents` section linking
    to each entry's heading anchor.
  - Why rejected: Any modern markdown viewer (glow, Obsidian, VS Code
    preview, GitHub's rendered view) has an outline pane that
    navigates the heading tree for free. At typical quarterly-export
    sizes (~50–100 entries) the built-in outline is sufficient. A TOC
    block would add slug-generation and anchor-collision code for
    marginal payoff. Deferred to `backlog.md` under "Table of contents
    in markdown export"; revisit if exports grow past ~200 entries or
    a consumer needs stable anchor links.

- **Option C: Grouping by type (or `--group-by <field>`).**
  - What it is: Group entries by `type` (shipped / learned / fixed)
    or by another field, with `--group-by` selecting the axis.
  - Why rejected: Project is the primary "what are you working on"
    axis users care about for review prep — the stage brief explicitly
    flagged project visibility as the load-bearing grouping. Type-
    grouping or multi-axis grouping can follow as a later polish;
    captured in `backlog.md` under "`--group-by <field>` in markdown
    export" with concrete revisit trigger (a promo-packet workflow
    organized by type buckets).

- **Option D: Custom templates via `--template <path>`.**
  - What it is: Let users point at a `text/template` file to render
    exports however they like.
  - Why rejected: The DEC-013 shape covers the four named use cases
    (retro, quarterly review, promo packet, resume update). Custom
    templates introduce stdlib `text/template` surface area plus
    template-distribution questions (ship built-ins? load from
    `~/.bragfile/templates/`?). Not justified until a real
    tool-specific format emerges that the flag matrix can't produce.
    Backlog entry exists.

- **Option E: Flat-only (skip grouping entirely as the default).**
  - What it is: Chronologically sorted single-section output as the
    default; `--group` as opt-in.
  - Why rejected: The user's ask was specifically project visibility
    ("I want to see quickly what projects I have been working on").
    Grouping as default directly serves that. `--flat` is preserved as
    the opt-out for users who want a pure timeline view.

- **Option F (chosen): All six choices above, applied together.**
  - What it is: Grouped-by-project default, ASC within group, level-3
    entry titles, `---` within-group separators, `--flat` escape,
    `--out` writer.
  - Why selected: Each sub-choice has either a user-named motivation
    (project visibility, forward chronological reading), a prior DEC
    it aligns with (SPEC-014's `--out` semantics), or a
    deliberately-deferred alternative captured in `backlog.md`. The
    combined shape is the simplest thing that serves the four named
    review-prep workflows without pushing polish concerns into MVP.

## Consequences

- **Positive:** One shape, documented once. Reusable between
  `brag export --format markdown` and (later, if wanted) `brag
  summary`'s aggregation rendering. `renderEntry` is now a shared
  helper, so any future metadata field (`links`, `quote`, etc.) lands
  in one place and flows through both `brag show` and `brag export
  --format markdown` automatically.

- **Negative:** The fixed shape is a commitment. Future schema
  additions require revising DEC-013 if they need a new rendering
  surface. In particular:
  - Adding a new `entries` column means updating `renderEntry`'s
    metadata table and (if filterable) the provenance filters echo.
  - Shipping `--group-by <field>` or `--template <path>` revises this
    DEC; their backlog entries assume DEC-013 stays the default.
  - Entries with no description render the same as entries without
    description plus optional metadata — a small cosmetic surface
    that's fine at MVP scale but could feel inconsistent if the
    metadata table ever grows significantly.

- **Neutral:** The `---` separator between entries makes the rendered
  markdown slightly chattier than strictly necessary; users who care
  can strip it with `sed`. The separator earns its keep in viewers
  that render `---` as a horizontal rule (glow, Obsidian, GitHub) —
  makes scanning faster.

## Validation

Right if:
- A user opening a `brag export --format markdown --since 90d` output
  can read it top-to-bottom and follow the quarter's narrative without
  jumping around. Confirmed by the load-bearing golden test
  (`TestToMarkdown_DEC013FullDocumentGolden` in SPEC-015 locks shape
  on a fixture with four entries across three groups including
  `(no project)` last).
- `brag show` and `brag export --format markdown` render the same
  entry identically except for the heading level and description
  heading (level 1/2 vs level 3/4). Confirmed by
  `TestRenderEntry_HeadingLevel1` + `TestRenderEntry_HeadingLevel3`
  asserting on both forms of the same fixture entry.
- A reviewer scanning the summary block can see at a glance the
  shape of the period's work (which projects, which types, how many
  of each) before drilling into entries.

Revisit if:
- A user reports that DESC-within-group would read better for their
  workflow (e.g., "latest wins" for promo packets). Then split
  ordering into a flag rather than changing the default.
- Export sizes routinely exceed ~200 entries and users ask for a
  TOC. Promote the backlog entry.
- A concrete need for type-grouping or custom templates emerges from
  a real review-prep session. Promote `--group-by` or `--template`.
- A future schema change (new column, new relationship) requires a
  rendering surface this DEC doesn't predict. Revise DEC-013 in
  lockstep.

Confidence: 0.82. Each sub-choice has a grounded rationale, but two
softer sub-choices keep the composite below 0.90:
- Choice 3 (ASC-within-group) is a user-ergonomic judgment that could
  reasonably go either way for a reader who prefers "most recent
  first"; 0.75 on its own.
- Choice 5 (`--flat` heading wrapper as `## Entries (chronological)`
  vs. no wrapper) is a cosmetic micro-choice; 0.70 on its own.
The other four are stronger (0.85–0.90 range), so 0.82 for the
composite.

## References

- Related specs:
  - SPEC-015 (emits this DEC; wires `brag export --format markdown`
    and lifts `renderEntry` into `internal/export`).
  - SPEC-014 (shipped; DEC-011 JSON shape, the sibling export format;
    `--out` semantics locked here carry forward verbatim).
  - SPEC-006 (shipped; original `renderEntry` home in
    `internal/cli/show.go`, whose tests stay green after the lift).
- Related decisions:
  - DEC-011 (shared JSON shape) — sibling export format in STAGE-003;
    `--out` semantics, filter-flag reuse, and the overall
    "locked-shape-tested-by-golden" pattern carry across.
  - DEC-004 (tags comma-joined TEXT) — tags render as comma-joined
    string in the entry metadata table, matching storage.
  - DEC-006 (cobra framework) — `--format markdown` and `--flat` are
    declared on the existing `brag export` command; no new command.
  - DEC-007 (required-flag validation in `RunE`) — `--flat`
    validation (reject when combined with `--format json`) uses
    `UserErrorf`.
- Related constraints: `stdout-is-for-data-stderr-is-for-humans`
  (markdown body to stdout; error messages to stderr).
- Related backlog entries:
  - "Table of contents in markdown export" (deferred `--toc` flag).
  - "`--group-by <field>` in markdown export" (deferred type/tag/
    month grouping).
  - "`--template <path>` for custom markdown rendering" (deferred
    text/template support).
- Related docs:
  - `docs/api-contract.md` (`brag export` section gains markdown
    subsection).
  - `docs/tutorial.md` (§4 gains a markdown example).

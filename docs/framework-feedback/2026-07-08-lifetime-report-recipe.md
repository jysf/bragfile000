# Feedback: add a `lifetime-report` recipe to the template

*Written 2026-07-08 against `bragfile` (claude-only variant), after four
shipped projects (PROJ-001…004) / 14 stages / 52 shipped specs / 5
releases. Feedback is from the agent (Claude) playing every role.
Intended for the spec-driven-template author; not part of the `bragfile`
project itself.*

---

## The gap

The template ships two reporting recipes and they cover two of three
natural horizons:

- **`just status`** — *now*. Point-in-time snapshot: active project,
  specs by cycle, stale items.
- **`just weekly-review`** — *the recent slice*. Active project only,
  last 7 days, framed as a review-and-act prompt.

There is no **lifetime** horizon — "tell the story of this repo from the
first commit to now, across *every* project." A user asked for exactly
this ("something like `just status` but more about the lifetime of the
project so far"), and producing it meant hand-assembling data from
`status`, `specs-by-stage`, `CHANGELOG.md`, every `brief.md`'s frontmatter,
the `decisions/` dir, and `git tag`. That assembly is mechanical and
identical for any repo built on the template — it should be a recipe.

## What was added downstream (candidate to upstream)

`scripts/lifetime-report.sh` + `just lifetime-report`. It follows the
**`weekly-review` pattern deliberately**: bragfile does the *aggregation*
(release timeline from tags + CHANGELOG headings, per-project
created→shipped from brief frontmatter, all DECs, git span, plus the
existing `status` and `specs-by-stage` output inlined), and prints a
**synthesis prompt** — an LLM writes the narrative arc. This matches the
template's own "tool owns data + shaping, LLM owns prose" posture that
bragfile's `review`/`summary`/`story` commands already embody. It is
whole-repo (loops `projects/*/brief.md`), not active-project-scoped,
which is the one structural thing `weekly-review` can't be reused for.

Files, if the author wants to lift them verbatim:
- `scripts/lifetime-report.sh` (uses only existing `_lib.sh` helpers:
  `require_initialized`, `get_repo_id`, `get_variant`, `get_active_project`)
- `justfile` recipe block next to `weekly-review` / `specs-by-stage`

## Why it belongs in the template, not just here

1. **It generalizes with zero project-specific logic.** Everything it
   reads is template-standard structure (brief frontmatter, `decisions/`,
   `CHANGELOG.md`, tags). Nothing in the script knows it's bragfile.
2. **It's the natural artifact at a project/version boundary** — the same
   moment you'd cut a release or open a retro. A template that models
   Repo→Project→Stage→Spec should have a recipe that reads *up* to the
   repo level, and currently the ladder tops out at the weekly slice.
3. **It composes with the retro flow.** The three-project retrospective
   (`docs/reports/cross-project/`) had to re-derive the same timeline by
   hand; a standing recipe removes that toil each time.

## Minor observations while doing this

- **`get_active_project` returned `PROJ-001-mvp`** even though all four
  projects are shipped and PROJ-004 is newest. For a *status* recipe the
  "active" resolver picking the oldest-shipped project when none is
  `active` is misleading. Not a blocker for the lifetime report (it reads
  all projects), but the resolver's "no active project → fall back to
  first" branch is worth a look — it also mislabels the header of
  `just status`.
- **Brief `status:` lives under a nested `project:` block** and some
  briefs carry an inline `# comment` after the value (the documented
  status-drift gotcha). The `grep | awk '{print $1}'` parse in the script
  tolerates the comment, but the template's own status resolver has been
  bitten by this before — a note in `projects/_templates/brief.md` that
  `status:` must be comment-free (or a resolver that strips comments)
  would prevent the recurring paper-cut.

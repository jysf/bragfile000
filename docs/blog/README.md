# bragfile blog

Long-form write-ups about why bragfile exists, who it's for, how to use
it, how it was built, and what was learned shipping it. Posts live here
as markdown for now; publication target (HN / dev.to / Substack /
personal site / GitHub Discussions) is decided per-post and may differ
between posts. The in-repo copy is canonical regardless.

## Planned posts

Numbered in the order they were brainstormed; publish order may differ.

| # | Slug | Status | Notes |
|---|---|---|---|
| 1 | `why-bragfile.md` | **draft** | Motivation. Why keep a brag log at all; the gap between "I shipped a lot" and "I can articulate what I shipped"; the local-first / no-cloud framing. |
| 2 | `who-is-bragfile-for.md` | not yet drafted | Audience targeting. Mid+ career engineers, performance-review-anxious folks, the "I forget what I worked on three months ago" crowd. Honest about who it's NOT for. |
| 3 | `using-brag-with-ai-tools.md` | not yet drafted | Working title `using-brag-with-claude-codex-cursor.md`. The AI-integration story: `docs/brag-entry.schema.json` as the contract, `scripts/claude-code-post-session.sh` as the reference hook, `examples/brag-slash-command.md` as the template. Concrete walkthroughs for Claude Code + Cursor + Codex if applicable. |
| 4 | `how-brag-was-built.md` | **draft** | Process narrative. The spec-driven framework (Repo → Project → Stage → Spec → Cycle), ~4 weeks from brief to v0.1.0, 5 stages × 23 specs × 14 DECs. This was the deferred `docs/blog-spec-driven-bragfile.md` artifact mentioned in PROJ-001's brief — new canonical path. |
| 5 | `capturing-a-good-brag-entry.md` | not yet drafted | Practical how-to. Field-by-field walkthrough; impact-framing prompts; examples of strong vs weak entries. Closes the gap between #2 (who) and #3 (AI capture). |
| 6 | `the-weekly-brag-review-rhythm.md` | not yet drafted | Workflow post on `brag review --week`, piping into Claude for synthesis, the rhythm itself. Highlights an underused feature. |
| 7 | `why-pure-go-sqlite.md` | not yet drafted | Technical deep-dive on DEC-001 + `modernc.org/sqlite` + CGO-free distribution. Different audience (Go folks); SEO-discoverable. Probably 600-800 words. |
| 8 | `codification-meta-rules.md` | not yet drafted | Synthesis post on §12 literal-artifact-as-spec + §12(b) design-time pre-flight + the paired-opposing-outcome N=2 codification meta-rule. Patterns extracted from the narrative in #4; referenceable independent of bragfile. |
| 9 | `whats-deferred-and-why.md` | not yet drafted | Anti-roadmap post. The backlog as durable artifact: ~20 entries with explicit trigger conditions, including macOS notarization, attachments, LLM-backed summaries, tags normalisation. Counter-cultural in software writing where roadmaps usually aspire. |

## Raw material

When drafting any of the above, mine these sources:

- `docs/framework-feedback/process-feedback.md` — process notes
  written ~2 weeks into the build, captures the practitioner-
  perspective on the framework while it was still being formed.
- `docs/framework-feedback/scale-recommendations.md` — companion
  notes on whether/how the framework scales beyond small projects.
- `projects/PROJ-001-mvp/brief.md` — particularly the `##
  Project-Level Reflection` section drafted at project close.
- The five shipped stage files at
  `projects/PROJ-001-mvp/stages/` — each has a `## Stage-Level
  Reflection` section.
- The 23 shipped specs under
  `projects/PROJ-001-mvp/specs/done/` — each has both a
  build-phase reflection and a Reflection (Ship) section.
- `projects/PROJ-001-mvp/session-log.md` — append-only working
  log; captures cross-spec context and lessons that don't live
  in any single artifact.
- `AGENTS.md` §4 / §9 / §10 / §12 — the codified addenda
  themselves, with the post-mortem framing of when each landed
  and why.
- `docs/reports/security/2026-04-26-pre-distribution-security-review.md` —
  280-line security review report; raw material for any
  security-flavoured posts.
- `CHANGELOG.md` — the public-facing v0.1.0 history.

## Conventions

- Markdown for now. Hugo / mdBook / static-site generator
  decided per-publication-target.
- Filename slug = lowercase, dashes between words, `.md`.
- Each post opens with a one-line tagline + the publication
  target (or `target: TBD`) in a YAML front-matter block
  once we settle on one.
- Date-stamps on completed posts; not on drafts.
- Cross-references between posts use relative links
  (`./other-post.md`).
- Inline links to repo artifacts use repo-root-relative paths
  (`../../projects/PROJ-001-mvp/brief.md`).
- Code blocks use fenced syntax with language hints; bragfile
  commands shown verbatim, not paraphrased.

## Out of scope (don't write yet)

- v0.1.0 announcement / launch-post / Show HN. Explicitly
  counter to PROJ-001's brief framing of "no marketing push;
  ship for the learning value."
- Roadmap / "coming soon" posts. The backlog is the durable
  artifact; an anti-roadmap post (#9) covers this honestly.
- Anything that requires bragfile to be more popular than it
  is. Posts that name an audience honestly are fine; posts
  that aspire to be referenced by a community that doesn't
  exist yet are not.

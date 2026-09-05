# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

This project uses `AGENTS.md` as the **single source of truth** for agent
instructions. This file is a pointer on purpose.

**Do not expand it into a copy of AGENTS.md.** A second copy of derived guidance
is the exact drift pattern `AGENTS.md` §9 exists to prevent — nothing diffs the
two, so the copy goes stale silently the first time the original changes. If you
want them genuinely unified, symlink rather than duplicate:

```bash
rm CLAUDE.md && ln -s AGENTS.md CLAUDE.md
```

**Read `/AGENTS.md` now.** It contains the work hierarchy (Repo → Project →
Stage → Spec → Cycle), tech stack and versions, exact build/test/run commands,
coding and testing conventions, git and PR conventions, and cycle-specific rules.

## Three things that surprise a fresh session

- **`just test` is Go-only.** `just test-docs` is a *separate* harness of
  documentation-content assertions that also gates CI, and it is the one most
  likely to fail on a docs-only change. `just lint` (golangci-lint, version
  pinned in `.github/workflows/ci.yml`) is the third gate. Run all three.
- **Cycles run in separate sessions** (§6). Frame, design, build, verify and ship
  each get a fresh one — a continuation session misses the drift the split
  exists to catch.
- **Numbers in docs are derived, not typed.** `just inventory` regenerates the
  inventory block in `docs/engineering-practices.md`; paste its output and never
  hand-edit a row. The same rule is why this file states no counts of its own.

## Where to look

- Rules and constraints: `/guidance/constraints.yaml`
- Architectural rationale: `/decisions/`
- Current work: `just status` names the active project — then read its `brief.md`

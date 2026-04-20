# CLAUDE.md

This project uses `AGENTS.md` as the single source of truth for agent instructions.

**Please read `/AGENTS.md` now.** It contains:
- The work hierarchy (Repo → Project → Stage → Spec → Cycle)
- Tech stack and versions
- Build, test, and run commands
- Coding and testing conventions
- Git / PR conventions
- Phase-specific rules

For project rules and constraints, read `/guidance/constraints.yaml`.
For architectural rationale, browse `/decisions/`.
For what we're currently building, read `/projects/<active-project>/brief.md`.

Run `just status` to see the current state of the repo.

---

*This file exists because some tools look for `CLAUDE.md` specifically. To keep a single source of truth, consider making this a symlink to `AGENTS.md`:*

```bash
rm CLAUDE.md
ln -s AGENTS.md CLAUDE.md
```

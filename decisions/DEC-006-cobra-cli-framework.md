---
insight:
  id: DEC-006
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

created_at: 2026-04-19
supersedes: null
superseded_by: null

tags:
  - cli
  - dependency
---

# DEC-006: Use `spf13/cobra` as the CLI framework

## Decision

`brag` uses `github.com/spf13/cobra` (and its underlying flag library
`github.com/spf13/pflag`) for argv parsing, subcommand wiring, and
auto-generated help/usage text.

## Context

SPEC-001 is about to add a top-level dependency to `go.mod`, which the
`no-new-top-level-deps-without-decision` constraint says requires a
DEC first. The CLI surface documented in `docs/api-contract.md` has
eight user-facing subcommands (`add`, `list`, `show`, `edit`, `delete`,
`search`, `export`, `summary`), each with its own flags, plus
persistent flags (`--db`, `--version`, `--help`). Rolling that by hand
against `flag` is a meaningful amount of busywork that doesn't differentiate
the product.

## Alternatives Considered

- **Option A: stdlib `flag` + hand-rolled dispatch**
  - What it is: Parse `os.Args`, route the second arg to a subcommand
    handler, each handler owns its own `flag.NewFlagSet`.
  - Why rejected: Workable but every subcommand re-implements help
    formatting, flag inheritance, and error reporting. Roughly 200–300
    lines of framework-ish code we don't need to own.

- **Option B: `urfave/cli`**
  - What it is: Mature alternative to cobra, lighter API surface.
  - Why rejected: Smaller community footprint in the Go-tool ecosystem
    and less convention overlap with the tools we'll interact with
    (goreleaser's own CLI uses cobra; many examples in the Go CLI
    world assume it). No technical dealbreaker — this is a tie-breaker
    on familiarity and ecosystem.

- **Option C: `alecthomas/kong`**
  - What it is: Struct-tag-based CLI parser. Very ergonomic.
  - Why rejected: Ergonomic but niche. Harder to find examples of
    patterns like `PersistentPreRunE` or composable command
    constructors when things get weird.

- **Option D (chosen): `spf13/cobra`**
  - What it is: The de-facto Go CLI framework. Used by `kubectl`, `gh`,
    `goreleaser`, `hugo`, etc.
  - Why selected: Maximum familiarity; examples for every pattern we'd
    need (subcommand trees, persistent flags, shell completion for
    future polish); no CGO; actively maintained.

## Consequences

- **Positive:** Subcommand scaffolding is essentially free. Shell
  completion (`brag completion zsh`) is one line if we ever want it.
  Help/usage is consistent across commands. Common patterns have
  obvious examples in other well-known repos.
- **Negative:** Cobra pulls in `pflag` and a bit of transitive weight
  (`inflection`, `viper` is NOT pulled unless used — we don't use it).
  The API has some legacy cruft (`Run` vs `RunE`, `Args` validators)
  that new contributors sometimes trip over.
- **Neutral:** Commits us to the cobra conventions for subcommand file
  organization. Matches `internal/cli/<name>.go` one-per-file plan in
  `docs/architecture.md`.

## Validation

Right if:
- Adding a new subcommand in a future spec takes only the subcommand
  file itself plus one `root.AddCommand(...)` line.
- Help/usage output never has to be hand-formatted.

Revisit if:
- We ever need argv parsing semantics cobra doesn't support (unlikely;
  none anticipated).
- Cobra stops being maintained (currently very active).

## References

- Related specs: SPEC-001 (introduces the dep), SPEC-003+ (every new
  subcommand uses it)
- Related constraints: `no-cgo` (cobra is pure Go ✓),
  `no-new-top-level-deps-without-decision` (this DEC satisfies it for
  cobra)
- External docs: https://cobra.dev/

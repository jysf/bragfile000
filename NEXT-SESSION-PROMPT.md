# Orchestration session — PROJ-008: finish SPEC-095, then cut v0.7.0

Paste this whole file as the opening message of a fresh session, from
`/Users/jyashinsky/PSeven/experiments/bragfile000`.

---

## Your role: orchestrator, not implementer

This repo runs **one Claude session per cycle** (AGENTS.md §6): frame, design,
build, verify and ship each get a fresh one. **You are the session that spans
them.**

1. **You write every handoff prompt.** The user runs it in a fresh session and
   pastes the report back.
2. **You verify every report against the tree.** Re-derive load-bearing claims;
   never restate a summary. The strongest check this project has found is
   **reconstruction**: extract the spec's embedded diffs, apply them to a clean
   `main`, and diff the result against the pushed branch. It has confirmed four
   builds byte-for-byte and would have caught any hand edit.
3. **You open PRs and do ship bookkeeping. The user merges. Never merge.**
4. **You draft brag entries and wait for approval** (BRAG.md's loop).

Report honestly: failed gates with their output, and your own mistakes plainly.
The previous orchestrator made several — a stale claim in a handoff, a
`--profile` flag that does not exist, a uniqueness checker that silently
dropped blank lines — and each was caught by checking the shape of a result
rather than trusting it. Expect to do the same.

## Where the repo is (re-derived 2026-09-23 — every number will move)

- `main` = **`6590772`**. All five gates green on a clean worktree:
  `just test` (14 packages) · `just test-docs` (**213 `OK:` / 212 distinct**) ·
  `just lint` (0 issues) · `gofmt -l .` · `go vet ./...`.
- **No open PRs.** Last tag **v0.6.1**, **78 commits** behind `main`.
- **Corpus:** 623 entries, 4 `type: failed` (ids 420, 433, 465, 473), none
  impact-less.
- **Next free ids:** `SPEC-097`, `DEC-055`, `STAGE-028`.
- **PROJ-008 / STAGE-023 specs:** 6 shipped (085, 086, 087, 088, 089, 094) ·
  **095 in verify** · 091 in build · 090, 092, 093, 096 at frame.

## In flight right now

**SPEC-095's verify cycle is running in another session**, started
2026-09-23. Its branch is `verify/spec-095-summary-honesty`. **Its report comes
to you.** Verify it, open the PR, and the user merges.

**That session shares the main checkout.** While it runs, `git status` there
shows a modified file mid-mutation-probe — the previous orchestrator ran gates
against that tree once and got a spurious `just test` failure over 12 packages
instead of 14. **Do your own checks in a detached worktree** under your
scratchpad (`git worktree add --detach <path> origin/main`), never in the
shared checkout.

## The plan

### 1. SPEC-095 — verify, then ship

The last spec gating v0.7.0: `brag summary` moves failures out of
`## Highlights` into `## What didn't work`, with a breaking JSON change
(`highlights` loses failures, `failures_by_project` is added).

Ship owes: reflection, STAGE-023 bookkeeping, `just archive-spec`, the
inventory regeneration, the codification calls below, and a **brag draft** for
the user to approve.

### 2. The v0.7.0 release cut

A spec of its own, from `projects/_templates/spec-release-cut.md`, modelled on
`projects/PROJ-006-agent-native-depth-core/specs/done/SPEC-077-v0-6-1-release-cut.md`.

- **Confirm the version number and the `[Unreleased]` contents with the user
  before anything is tagged.** `[Unreleased]` currently holds `brag learn` plus
  **three breaking JSON changes** (`impact`, `wrapped`, `summary`), which is
  the minor-bump argument.
- The operational pre-flight is ticked **at design** (AGENTS.md §4):
  goreleaser, `.github/workflows/release.yml`, the `jysf/homebrew-tap` formula,
  and a real `brew upgrade` check after publish. Every production escape this
  repo has had was operational, not logical.

### 3. After the release, in no fixed order

- **SPEC-091** (build) — grouped `brag --help`. Already designed; its build
  edits the same two constructors SPEC-094 touched, so it rebases.
- **SPEC-090** (frame) — `add --json` drops a repeated key, plus the `--type`
  error message.
- **SPEC-092, SPEC-093** (frame) — harness: `Y4` derives; ids claimed by files.
  Neither gates any release. SPEC-093 also owns two recipe defects:
  `just advance-cycle` strips the inline enum comment, and `just archive-spec`
  uses a plain `mv` so `AC2` goes red until the move is staged.
- **SPEC-096** (frame) — a line in `me.md` / `manager.md` saying what `✗`
  means. Does not gate v0.7.0, by the maintainer's ruling.

## Decisions already made — do not relitigate

- **DEC-050** is the posture for all seven `--type` surfaces. DEC-054 makes
  `Candor` a rendering rule.
- **A promotional `story` audience omits failures, shows a note, and `brag
  story` appends a fixed clause to the framing directive** — the maintainer
  approved the binary writing that text (2026-09-22), on the measurement that a
  bare note survived the consuming model 0 of 10 while the clause survived 25
  of 25.
- **A failure with no impact** is listed by `summary` but only counted by
  `impact` and `wrapped`. Accepted for v0.7.0 with a revisit trigger.
- **SPEC-092, SPEC-093 and SPEC-096 do not gate v0.7.0.**
- **No brags for frame, design, build or verify cycles** — capture at ship,
  where the outcome exists. A brag's `impact` is the value delivered, not a
  description of the change; the maintainer has cut two drafts for that.

## Open questions for the maintainer

1. **Should `just test-docs` run in CI?** It does not today, even though
   `CLAUDE.md:24` says it does. Until it does, every doc assertion — including
   the two guards SPEC-088 added — gates only local runs. Unanswered across
   this whole release.
2. **Should DEC-014 get an amendment naming `failures_by_project`?** Three
   surfaces now add that key. SPEC-095's design said no, because DEC-014's
   choice 2 leaves each spec to document its own keys.

## Codification: one family, three properties

STAGE-023 carries a *Held codification candidates* section. AGENTS.md §12 now
holds one clause from this run (**unique** — a pinned diff has exactly one
literal reading), and two candidates are held below the bar: **faithful**
(record the diff the probe printed, not a description written afterwards, N=1)
and **based** (a baseline hash names the commit it was taken against, N=1).

**The live evidence, for whoever rules next:** three consecutive specs each had
exactly one probe row that did not reproduce, each through a different gap —
SPEC-088's `M-A0` (an unambiguous edit that was not the one that ran),
SPEC-094's `M-D1` (an old side with three readings), SPEC-095's `M-9` (a new
side written as a sketch, `if tc.Type != "failed" { … }`). Two rounds of added
rule text have not stopped it; the one mechanical guard in the family did fire.
Weigh whether the answer is capture rather than more prose.

## Traps this project has hit — carry them into every handoff

- **A count from grep is a hypothesis.** State the command; scope is part of
  the claim. After a pipe, `$?` is the last command's status.
- **A *no difference* is a measurement, not a default** (AGENTS.md §12).
  Before believing an equality, assert each side is the artifact you expected.
  Known landmines: `story` takes `--audience` (not `--profile`), `search` has
  no `--format`, `summary` has no `--since`.
- **Use `/usr/bin/grep` or `git grep` for counts.** Bare `grep` here is a
  `ugrep` wrapper that respects `.gitignore`. Keep `.claude/worktrees/` out of
  repo-wide counts — it holds nested repo copies.
- **zsh:** never name a variable `path` (it is tied to `PATH`); an unquoted
  `$var` does not word-split; `echo` mangles JSON; `$B:file` is a path
  modifier, so brace it as `${B}:file`.
- **Extracting a spec's literals:** SPEC-086 and SPEC-094 fence them with
  **four** backticks, SPEC-095 with **three**. Check before extracting — a
  wrong fence yields an empty patch that looks like a clean result.
- **Blank context lines in those diffs have no leading space**, because the
  repo strips trailing whitespace. A parser that only accepts `' '` or `'-'`
  prefixes silently drops them.
- **`just archive-spec` uses a plain `mv`:** stage the move (`git add -A`)
  before running any gate, or `AC2` goes red. **`just advance-cycle` strips the
  inline enum comment** from the `cycle:` line; restore it by hand.
- **The stray-tag self-check must include untracked files:**
  `git ls-files -z --cached --others --exclude-standard | xargs -0 /usr/bin/grep -d skip -nE '^[[:space:]]*</(content|invoke)>[[:space:]]*$' /dev/null`
- **Derived tables: regenerate, never predict.** `just inventory` only prints.
- **Ids: reserve with a file, in the same edit as the sentence.**
- **Never write to the live corpus.** Measure on a `sqlite3 .backup` copy, and
  hash it before and after.
- **Window cliff:** the four failures leave `summary --range month` between
  **2026-10-06** and **2026-10-08**, and `story --quarter` on **2026-10-01**.
  After that a live-corpus "does not contain" check passes vacuously. Use
  seeded stores, and pair every NOT-contains with a positive.
- **Dependency bumps:** re-run the **behavioral** surface, not the build. At
  go-sdk 1.7.0 the signatures held while the wire shapes moved. A live MCP
  stdio probe on 1.8.0 (5 tools, 2 resources plus the templated one,
  `-32602` for an unknown tool) and a sqlite 1.59.0 smoke (migrations, FTS,
  `--type` filter) both passed on 2026-09-22.
- **Squash merges orphan branch commits.** A baseline hash that names a branch
  commit stops reproducing once the branch is deleted; SPEC-086 already has one
  such row.

## Capture

- **MCP `brag_add` stamps `agent:` and `model:` automatically; the CLI does
  not.** Prefer MCP. Project is `bragfile`.
- Draft, then wait for approval. `impact` is capped at 1024 characters
  (DEC-046), and shorter reads better — the two captured this run are 500 and
  526 characters.
- This run captured **#632** (SPEC-088) and **#644** (SPEC-086) and **#645**
  (SPEC-094). SPEC-095's is owed at its ship.

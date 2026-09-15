# Orchestration session — PROJ-008: finish STAGE-023 and cut v0.7.0

Paste this whole file as the opening message of a fresh session, from
`/Users/jyashinsky/PSeven/experiments/bragfile000`.

---

## Your role: orchestrator, not implementer

This repo runs **one Claude session per cycle** (AGENTS.md §6): frame, design,
build, verify and ship each get a fresh session. **You are the session that
spans them.**

1. **You write every handoff prompt.** The user runs it in a fresh session and
   pastes the report back.
2. **You verify every report against the tree before accepting it.** Re-derive
   load-bearing claims; never restate a summary. In the last run this caught
   real errors in both directions — cycles corrected the orchestrator's counts,
   and the orchestrator corrected cycles' diagnoses.
3. **You open PRs and do ship bookkeeping. The user merges. Never merge.**
4. **You draft brag entries and wait for approval** (BRAG.md's loop).

Report honestly: failed gates with their output, and your own mistakes plainly.

## Where the repo is (re-derived 2026-09-14 — every number here will move)

- `main` = **`7c615f5`**. All five gates green: `just test` · `just test-docs`
  (**199 `OK:` lines / 198 distinct ids** — two units, `S3` double-emits) ·
  `just lint` · `gofmt -l .` · `go vet ./...`.
- **Two open PRs**, both mergeable, 7/7 CI:
  - **#211** (`73dbf38`) — SPEC-091 design: grouped `brag --help` + three
    help-text gaps. Frame and design were collapsed; the spec is at `cycle: build`.
  - **#212** (`6208eeb`) — DEC-053 (in-place edit posture), STAGE-027 frame,
    round-2 field feedback routed onto the PROJ-008 brief.
- **Release:** last tag **v0.6.1** (2026-08-13), **56 commits behind `main`**.
  `brag learn` and SPEC-089's fixes are unreleased.
- **Corpus:** 570 entries, 4 `type: failed`. ~130 entries/week recently, most
  from one project. `brag memory`'s 200-entry pool now covers only ~30 days.
- **`next_id` on `main`:** SPEC-091, DEC-053, STAGE-024 — all already claimed by
  the open PRs or by prose. See *Traps*.

## Decisions already made — do not relitigate

- **v0.7.0 ships the honest-corpus work.** Scope and order below.
- **`brag edit` for agents (STAGE-027 / DEC-053) is v0.8.0**, not v0.7.0.
- **Projects need no work.** `brag add -p <any name>` needs no registration;
  `brag project` exists only to track where local in-progress work lives, so
  `project new` requiring `--path` is correct.
- **`brag lint` (STAGE-025, in #212's brief) waits for the maintainer's review.**
  Do not design, reword or act on it. Unregistered project names are legitimate.
- **STAGE-020 (evidence links) is parked.** Do not touch it.

## The plan

### Step 0 — tidy the open PRs, then the user merges them

- Strip the stray tool-call XML: a trailing `</content>` line at
  `DEC-053:128`, `STAGE-027:152`, and `SPEC-091:224`.
- #212's body tells the field reporter to "upgrade" for SPEC-089's fixes.
  Nothing has been released; correct it.
- Mark STAGE-027 as targeting v0.8.0. **Leave the lint item's text alone.**
- The two PRs don't overlap, so order is free. Merge one at a time and re-check
  gates on `main` between them.

### Steps 1–6 — v0.7.0

| # | Spec | State | Notes |
|---|---|---|---|
| 1 | **SPEC-088** — harness guards | frame, S (provisional) | `Y4` derives; the decision-type vocabulary (template lists 5 types, inventory counts 2, a `type: analysis` DEC hard-fails `Z7`); a guard against stray tool-call XML in tracked markdown; id reservation. Not user-facing, but every later spec hits these. **Keep it small — split rather than absorb.** |
| 2 | **SPEC-086** — `impact` + `wrapped` show failures honestly | frame, M | The release gate. Authors **DEC-050** (still unused; reserved). See notes below. |
| 3 | **`summary` + `story` follow-up** | not yet written | Second half of the gate. **Claim its id with a file** at SPEC-086 design. |
| 4 | **SPEC-091** — grouped help | build, M | Already designed (#211). Gives `learn` a visible home. |
| 5 | **SPEC-090** — `add --json` repeated key, plus a `--type` error message | frame, S→M, bug | Also fold in `delete`'s missing stdout signal, as SPEC-089 bundled two bugs. **Split if it grows past M.** |
| 6 | **v0.7.0 release cut** | not yet written | `projects/_templates/spec-release-cut.md`; SPEC-077 (v0.6.1) is the model. **Confirm with the user before tagging.** |

**STAGE-023 closes with one criterion half-met** (failures surviving in
`brag memory` beyond the pool horizon). Record it honestly, the way PROJ-007
closed with an explicit scope reduction.

### Notes per step

**SPEC-086** — framed twice, 649 lines; read all of it. The hard parts:
- **Fork A is the real work.** Once *Impact moments* stops carrying every entry
  with an impact, its count changes meaning, and DEC-048 forbids silently
  redefining it. The JSON envelope moves too — a breaking wire change.
- **Fork C:** put the failure predicate in `internal/aggregate`, single-sourced;
  `aggregate.IsAgentAuthored` + its drift-guard test is the precedent.
- **Fork D:** the empty *section* case. The cited precedent (DEC-014 part 4)
  covers the empty *document*, and the two surfaces already disagree.
- **Fork E:** re-framing already found the moving goldens by running the code:
  2–3 files across 2 packages (one in `internal/cli`), and
  `internal/export/memory_test.go:247` does **not** move. Re-verify at design.
- `wrapped`'s change amends **DEC-030**'s locked section arc.

**`summary` + `story` follow-up** — implements DEC-050's posture.
`summary`'s section is literally `## Highlights`. **All four** bundled `story`
profiles drop `type` in markdown (`exec` and `skip` are `candor: promotional`);
`--format json` already carries it. Turning `Candor` from "metadata surfaced to
the LLM" (`internal/story/profile.go:24`) into a body rule is a decision.

**SPEC-091** — relocates `cmd/brag/main.go`'s `AddCommand` block into a
testable `cli.AssembleRoot`. `test-docs.sh` greps `--help`; the spec carries a
premise-audit step. Its `edit --help` text documents the `$EDITOR` edit that
ships in v0.7.0 — correct, since agent edit is deferred.

**SPEC-090** — `encoding/json` silently keeps the last duplicate key and has no
hook for it; rejecting needs a `json.Decoder` token pre-pass. The MCP `brag_add`
ingress decodes with the same package — **framing must drive it, not assume it.**
`delete.go:75` (`Aborted.`) and `:86` (`Deleted.`) both print to stderr and
both return nil, so a script can't tell them apart — DEC-052's defect, on delete.

**`--type` error message (added to SPEC-090 by the user, 2026-09-14).** Today
`--type '!failed'`, `'-failed'`, `'shipped,failed'` return exit 0 and zero rows
with no diagnostic — and v0.7.0 is the release that gives users a reason to
exclude failures. Recommended rule, for framing to confirm: **if a `--type`
value contains `!`, `,`, whitespace or a leading `-`, and the query matches
zero rows, fail with a user error (exit 1)** saying `--type` matches one exact
value and has no negation or lists.
- *Only on zero rows,* because `type` is free-form on write (DEC-049), so a
  stored type could legitimately contain those characters. This rule can never
  reject a query that would have returned rows. None of the 20 stored types
  contain them today — re-check.
- *An error, not a stderr hint,* because an empty JSON envelope with exit 0 is
  the silent failure itself; scripts need the exit code (DEC-052's reasoning).
- *Sites:* the seven read commands (`list`, `story`, `export`, `coverage`,
  `wrapped`, `summary`, `impact`) each copy the same 3-line flag block, and MCP
  `brag_list` sets it at `internal/mcpserver/server.go:246` beside its existing
  input checks. One shared helper, not eight copies. Keep it out of
  `internal/storage` — storage doesn't emit user diagnostics.
- Real negation (an `--exclude-type`-style flag) stays out of v0.7.0.
- Update STAGE-023's backlog: its unwritten `--type` item is now part of SPEC-090.

**Release cut** — the operational pre-flight is ticked at design (AGENTS.md §4):
goreleaser, `.github/workflows/release.yml`, `jysf/homebrew-tap`, a real
`brew upgrade` check after publish. `[Unreleased]` holds a **breaking** JSON key
rename (`brag memory` → `candidates`), which is the minor-bump argument.

## Traps this project has hit — carry them into every handoff

- **A count from grep is a hypothesis.** An X-of-N claim is two measurements
  plus a unit. Grep scope is part of the claim: a `projects/`-only grep reported
  10 where the repo had 11, and a `head -30`-truncated list was once reported as
  "none".
- **Measure on the real corpus, not a 2-row fixture.** Two claims last run were
  false only because they were measured on a toy database.
- **Quote `--include` globs** — unquoted, zsh expands them and the search
  silently never runs. `find` works directly; the `rtk` proxy intercepts some
  forms. `$?` after a pipe is the last command's exit, not yours.
- **Mutation protocol (§12 + AGENTS.md:364):** confirm the target's hash
  **moved before running the gate** — a no-op mutant's green half looks exactly
  like a working guard. Restore from a `/tmp` backup, **never `git checkout`**.
- **A green guard is not evidence.** SPEC-089's guard tested 2 of 5 keys;
  deleting a key left all 14 packages green.
- **Derived tables: regenerate, never predict.** `just inventory` **only
  prints** — paste it with no blank lines inside the markers (`X3` is
  byte-for-byte). `Y3`/`Z7` now derive (SPEC-087): adding a DEC costs zero
  harness edits. **If either ever needs a hand-edit, that is a regression.**
- **Ids: reserve with a file, in the same edit as the sentence.** `next_id`
  scans filenames **in the working tree** — a file on an unmerged branch
  reserves nothing, and there's no guard for DEC ids. After #212 merges, the
  next stage id is **STAGE-028**; STAGE-024/025/026 exist only in prose and
  `just new-stage` can't produce them. Don't create stages this release;
  SPEC-088 owns the fix.
- **Stray tool-call XML.** Files written by sessions have ended in
  `</content>` / `</invoke>` — five files so far, surviving CI and multiple PRs. Until
  SPEC-088's guard lands, every handoff must grep
  `^\s*</(content|invoke)>\s*$` before committing a written document.
- **Squash merges detach stacked PRs.** Don't stack. If it happens, cherry-pick
  the commit onto `main`, prove the tree is unchanged, and `--force-with-lease`.
  Retargeting alone doesn't re-trigger CodeQL; a push does.
- **`just archive-spec` — run it once.** The active-project resolver needs a
  comment-free `status: active`.
- **Never write to the live corpus** from a test. Use `--db` or `t.TempDir()`.

## Not in v0.7.0 — do not fold in

- **STAGE-027 / agent edit → v0.8.0.** Two design gaps to settle then:
  `Store.Update` overwrites the whole row (`WHERE id = ?`) and a partial edit is
  read-then-write across two transactions, so concurrent agent edits lose one
  silently (fix: expected-`updated_at` check, or merge inside one transaction);
  and nothing records who edited an entry.
- **`brag lint` / STAGE-025** — maintainer review first.
- **Real `--type` negation** (e.g. an `--exclude-type` flag). Only the error
  message for negation/list syntax is in v0.7.0, as part of SPEC-090.
- **Memory pool durability** (`memory-pool-composition-excludes-older-entries`).
- **Housekeeping:** `projects/PROJ-001-mvp/backlog.md` still lists the shipped
  streak bug as open and its "Removed / delivered" section is empty; two used
  prompts in the repo root (`retired-tap-migration-prompt.md`,
  `spec-078-verify-prompt.md`); `internal/cli/root.go:13`'s unguarded "four";
  goreleaser's `brews:` deprecation and no shell completions in the formula
  (may surface at the release cut — route, don't absorb).
- **PROJ-009** — still zero benchmarks.

## Capture

- **MCP `brag_add` stamps `agent:`/`model:` automatically; the CLI doesn't.**
  Prefer MCP. Project is `bragfile`.
- Draft, then wait for approval. Transcribe `impact` from each spec's
  *"what can a user do now that they couldn't before"* answer.
- Record honest failures as `type: failed`. `brag learn` isn't in the released
  binary until v0.7.0 — use `brag_add` with `type: failed` until then.
- `impact` is capped at **1024 characters** (DEC-046); MCP rejects longer.

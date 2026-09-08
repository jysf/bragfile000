---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-091
  type: task                       # epic | story | task | bug | chore
  cycle: build                     # frame | design | build | verify | ship
  blocked: false
  priority: high
  complexity: M                    # S | M | L  (L means split it)

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-09-08

references:
  decisions: [DEC-011, DEC-012, DEC-046]
  constraints: [stdout-is-for-data-stderr-is-for-humans, test-before-implementation]
  related_specs: [SPEC-089, SPEC-090]
---

# SPEC-091: the help surface reads as families and tells the truth

> **Cycle: frame+design (collapsed by user direction).** **GO** at complexity
> **M**. Written in one session while triaging round-2 field feedback from heavy
> agent use of `brag` (`~/ContextCore`, 450-entry corpus, 2026-09-06→08). A build
> session picks it up: the `## Failing Tests` are written; make them pass. Two
> cohesive halves on one surface (`brag --help` grouping + three per-command
> `--help` truth-gaps) — one PR, because both are about the same artifact
> (`--help`) and neither is worth a branch of its own.

## Context

Round-2 field feedback retracted an earlier wrong claim ("the reporting surface
is thin") and, in doing so, produced the highest-ROI finding in either round:
the read/digest surface is not thin — `story`, `wrapped`, `coverage`, `impact`,
`memory`, `summary`, `review`, `spark` all exist and all take `--project` — it
is **undiscoverable**. The reporter hand-built in Python what `coverage` and
`story` already ship, and worked for two days without finding them.

**Root cause (verified):** `brag --help` renders all 22 subcommands as one flat
alphabetical block. cobra v1.10.2 supports command groups
([root.go](../../../internal/cli/root.go) registers none). The reporter's own
diagnosis: *"the digest family does not read as a family… grouping the help
output — write / read / digest / admin — would likely have prevented every
mistake."*

This is the **third instance** of a class this project's own brief already
named — *"a capability exists and the artifact that would surface it doesn't"*
(Scope item 5: `BRAG.md`'s write-only read path; and
`docs/framework-feedback/process-feedback.md` §6: the spec-template work-log
hook). Round 2 is instance 3, on the CLI's own front door.

Three companion **truth-gaps** in per-command `--help`, all cheap, all filed by
the same reporter:

1. **§4.1** — `impact` caps at 1024 bytes ([capture/validate.go](../../../internal/capture/validate.go) `MaxImpact`), enforced, but `add --help` never says so; the only signal is a rejection after composing. No field limit is documented.
2. **§4.2** — the `$EDITOR` buffer format (headers, blank line, body) is undocumented; the reporter reverse-engineered it with a shim. `edit --help` should state it — including that a repeated header is now rejected (SPEC-089) and that the entry id prints to stdout on save (SPEC-089's applied signal, §2.1).
3. **§4.3** — a documentation *trap*: `brag_add` over MCP stamps `agent:` and takes `model`/`session`; CLI `brag add` stamps **none** and `add --help` doesn't say so. The reporter concluded stamping didn't exist and *wrote that into another project's instructions* before catching it. Naming the asymmetry in `add --help` is the fix (CLI provenance flags for parity — §4.3 option B — are out of scope; see below).

The help grouping also gives `brag learn` a prominent home in the **Write**
group, partially answering §3.6 (`failed` is under-used largely because `learn`
is buried in a flat list).

**Stage placement.** A `task` riding the active stage (STAGE-023), like SPEC-089
before it — not part of the failure-capture narrative backlog and not a gate on
the stage's ship. It rides here because it is current work and because it is a
concrete instance of Scope item 5 (read-path discoverability). It is a CLI-human
surface; the agent-facing `BRAG.md` read path (Scope item 5) stays its own work.

## Goal

`brag --help` presents its commands in four labelled groups (Write / Read /
Digest / Admin) instead of one flat list, and `add`/`edit` `--help` document the
field limits, the editor buffer format, and the CLI-vs-MCP provenance asymmetry
that today are only learnable by trial and error.

## Inputs

- **Files to read:** `cmd/brag/main.go` — the current `root.AddCommand(...)` block (the untestable assembly this spec relocates).
- **Files to read:** `internal/cli/root.go` — `NewRootCmd`; groups get registered here or in the new assembler.
- **Files to read:** `internal/cli/root_test.go` — the existing `NewRootCmd("test-v0")` help-flag tests; the new grouped-help test lives alongside.
- **Files to read:** `internal/cli/add.go`, `internal/cli/edit.go` — the two `Long` strings to extend.
- **Files to read:** `internal/capture/validate.go` — the exact caps to quote (Max* consts).
- **Files to read:** `docs/for-ai-agents.md` — the provenance-stamping doc `add --help` should cross-reference; confirm the section that describes MCP stamping.
- **Related code paths:** `scripts/test-docs.sh` — greps `--help` output; premise-audit target (see Implementation Context).

## Outputs

- **New export:** `cli.AssembleRoot(version string) *cobra.Command` — creates the root, registers the four groups, adds every subcommand with its `GroupID`, and sets the help/completion command group ids. `cmd/brag/main.go` collapses its 22-line `AddCommand` block to `root := cli.AssembleRoot(version)`. This co-locates the command list with the constructors that own it and makes the whole assembled tree — and its grouped help — unit-testable in the `cli` package (it is not today; assembly lives in `package main`).
- **Files modified:** `internal/cli/root.go` (or a new `internal/cli/assemble.go`) — `AssembleRoot` + group registration; `cmd/brag/main.go` — call it; `internal/cli/add.go`, `internal/cli/edit.go` — extended `Long` strings.
- **New tests:** `internal/cli/root_test.go` (grouped help), `internal/cli/add_test.go` + `internal/cli/edit_test.go` (help truth-gaps).
- **Database changes:** none.
- **New decisions (emit at build):**
  - `DEC-NNN` — the four-group help taxonomy (Write/Read/Digest/Admin) and the assignment of each command to a group (a small, user-visible IA decision worth a record so a new command knows which group it joins). Claim a free id at build (frame reserved none).

## Locked design decisions

1. **Four groups, fixed taxonomy.** Group ids / titles and membership are locked
   here (literal-artifact §12): a new command must be assigned deliberately, not
   defaulted into "Additional Commands".
   - **`write` — "Write commands:"** — `add`, `learn`, `edit`, `delete`
   - **`read` — "Read commands:"** — `list`, `show`, `search`, `export`, `tags`
   - **`digest` — "Digest commands:"** — `summary`, `review`, `stats`, `impact`, `wrapped`, `coverage`, `spark`, `story`, `memory`
   - **`admin` — "Admin commands:"** — `project`, `tag`, `mcp`, `completion`, `help`
2. **`GroupID` is set in each command's constructor**, not in `AssembleRoot`, so the group a command belongs to lives next to the command. `AssembleRoot` owns only the `AddGroup` registration + `SetHelpCommandGroupID("admin")` / `SetCompletionCommandGroupID("admin")`.
3. **No command may land ungrouped.** cobra renders ungrouped commands under an "Additional Commands:" header; that header appearing is a bug. A NOT-contains test guards it (§12 NOT-contains rule: audited — "Additional Commands" appears in no load-bearing `Long`/help prose we author).
4. **Doc-gap content is additive to the existing `Long` strings**, not a rewrite — the Examples blocks and existing guidance stay. Rejected alternative: a separate `brag help-fields` command — heavier, and the reporter's pain was that `--help` itself was silent, so the fix belongs in `--help`.

## Acceptance Criteria

- [ ] `brag --help` output contains the four group headers `Write commands:`, `Read commands:`, `Digest commands:`, `Admin commands:` and does NOT contain `Additional Commands:`.
- [ ] `story`, `memory`, `coverage` render under the Digest group; `add`/`learn` under Write (proving the digest family reads as a family).
- [ ] `cli.AssembleRoot` returns a root whose subcommand set is identical to today's `main.go` assembly (no command dropped or added).
- [ ] `add --help` states the `impact` 1024 limit (and the other field limits) and names the CLI-vs-MCP provenance asymmetry (CLI stamps no `agent:`/`model:`; MCP `brag_add` does).
- [ ] `edit --help` documents the buffer format (headers → blank line → body; the five canonical headers; a repeated header is rejected; save-unchanged aborts; the entry id prints to stdout on save).
- [ ] `just test`, `just lint`, `just test-docs` all pass; no existing help/root/add/edit test regresses.

## Failing Tests

Written during design, BEFORE build. Make these pass in build.

- **`internal/cli/root_test.go`**
  - `"TestAssembleRoot_HelpIsGrouped"` — build `AssembleRoot("test-v0")`, capture `--help` on an in-memory `outBuf`; assert it contains all four group headers and does NOT contain `Additional Commands:`. Fails today (no groups; assembly not even in-package).
  - `"TestAssembleRoot_DigestFamilyGrouped"` — line-scan the help: assert `story`, `memory`, and `coverage` each appear on a command line that falls after the `Digest commands:` header and before the next group header. Use line-based positional checks (AGENTS.md §9 heading-assert lesson), not bare `Contains`.
  - `"TestAssembleRoot_CommandSetUnchanged"` — assert the set of immediate subcommand `Name()`s equals the known 22 (list them literally) plus cobra's `help`/`completion`, so a future accidental drop is caught.

- **`internal/cli/add_test.go`**
  - `"TestAddCmd_HelpDocumentsImpactCap"` — `add --help` contains `1024` (distinctive token; not produced by cobra otherwise).
  - `"TestAddCmd_HelpDocumentsProvenanceAsymmetry"` — `add --help` contains a distinctive phrase naming the asymmetry (e.g. `only when capturing over MCP`); pick a token unique to this content, not a generic word (§9).

- **`internal/cli/edit_test.go`**
  - `"TestEditCmd_HelpDocumentsBufferFormat"` — `edit --help` contains `blank line` and `repeated header` (distinctive tokens covering the format + the SPEC-089 reject).

## Implementation Context

*Read before starting build.*

### Decisions that apply
- `DEC-011` / `DEC-012` — the export/`--json` schema `add --help` already cites; the field-limits note must not contradict the schema keys.
- `DEC-046` — the caps live in `internal/capture`; quote `MaxImpact` etc. from there, do not hard-type a number that could drift (state the value AND that it is enforced by `internal/capture`).
- (SPEC-089) — `edit` already rejects duplicate headers and prints the id on save; `edit --help` documents behavior that now exists, so its assertions are truthful.

### Constraints that apply
- `stdout-is-for-data-stderr-is-for-humans` — unaffected (help is help), but the `edit --help` text that says "the id prints to stdout on save" must match SPEC-089's actual behavior; do not restate it wrongly.
- `test-before-implementation` — run the new tests first; the grouped-help ones fail because groups don't exist and `AssembleRoot` isn't defined; the help-text ones fail on the missing substrings.

### Premise audit (run at design AND re-run at build — §9)
`scripts/test-docs.sh` greps `brag --help` / per-command help. Before locking,
`grep -n -- '--help\|brag help\|Available Commands\|Usage:' scripts/test-docs.sh`
and reconcile: any assertion that pins the flat "Available Commands:" block or a
command's alphabetical position will break when grouping lands, and belongs in
`## Outputs` as a planned update. (The grouped output still lists every command;
what changes is the section headers and ordering.)

### Out of scope (route elsewhere — see the PROJ-008 brief's field-feedback backlog)
- **§4.3 option B** — CLI `--agent`/`--model`/`--session` flags for MCP parity. A capture/provenance feature, larger than documenting the asymmetry. Not here.
- **§3.5 `brag lint`**, **§2.2 unregistered-project reporting**, **§3.3 `verified_at`/staleness** → STAGE-025 (the mirror).
- **§3.2 `brag link` relationship primitive** → STAGE-026 (story-surface v2).
- **§3.1 duplicate-detection on `add`**, **§3.4 `edit` flag mode** → smaller ergonomics / capture-completeness; unscheduled.
- Do NOT restructure any command's own behavior; this spec changes only help *presentation* and help *text*, plus the assembly seam.

## Notes for the Implementer

- **cobra grouping.** `root.AddGroup(&cobra.Group{ID: "write", Title: "Write commands:"}, ...)` before `Execute`; set `cmd.GroupID = "write"` in each constructor. For the auto `help` command call `root.SetHelpCommandGroupID("admin")`; for the explicitly-added completion command either set its `GroupID` in `NewCompletionCmd` or call `root.SetCompletionCommandGroupID("admin")`. A command with a `GroupID` whose group isn't registered makes cobra print a warning to stderr at help time — the `TestAssembleRoot_HelpIsGrouped` assertion of no "Additional Commands:" plus a clean errBuf catches that.
- **AssembleRoot signature.** `func AssembleRoot(version string) *cobra.Command`. It calls `NewRootCmd(version)` internally, registers groups, and adds the same commands `main.go` adds today (move the list verbatim, then attach GroupIDs via the constructors). `main.go` keeps `storage.SetBuildVersion` and `root.Execute()`; only the `NewRootCmd`+`AddCommand` block collapses to the one call.
- **Doc strings.** Keep each `Long`'s existing Examples block. Add a short "Field limits:" line to `add --help` (title 256, impact 1024, project/type 64, tags 32×64, description 100000 bytes — sourced from `internal/capture`), and a "Provenance:" line naming the MCP-only stamping with a pointer to `docs/for-ai-agents.md`. Add a "Buffer format:" block to `edit --help` per §4.2, ending with the repeated-header reject and the stdout-id-on-save note.
- **Fail-first.** Run the three test files; confirm the grouped-help tests fail on undefined `AssembleRoot` / missing group headers and the help-text tests fail on missing substrings — not on a compile error elsewhere.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `feat/spec-091-help-surface`
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - `DEC-NNN` — help group taxonomy (claim a free id)
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - [any]

### Build-phase reflection (3 questions, short answers)

1. **What was unclear in the spec that slowed you down?**
   — <answer>

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — <answer>

3. **If you did this task again, what would you do differently?**
   — <answer>

---

## Reflection (Ship)

*Appended during the **ship** cycle.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>

4. **What can a user do now that they couldn't before?** — one sentence.
   — <answer>
</content>

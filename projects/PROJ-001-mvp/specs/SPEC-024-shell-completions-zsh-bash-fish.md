---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-024
  type: story                      # epic | story | task | bug | chore
  cycle: build                     # design complete; ready for a fresh build session
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-001
  stage: STAGE-005
repo:
  id: bragfile

agents:
  architect: claude-sonnet-4-6
  implementer: claude-sonnet-4-6   # fresh session for build
  created_at: 2026-05-10

references:
  decisions: []                    # no new DECs; cobra already a dep
  constraints:
    - no-new-top-level-deps-without-decision
    - stdout-is-for-data-stderr-is-for-humans
    - test-before-implementation
  related_specs:
    - SPEC-021
    - SPEC-022
    - SPEC-023
---

# SPEC-024: Shell completions — zsh, bash, fish

## Context

SPEC-024 is the last spec in STAGE-005 and the last spec of PROJ-001-mvp.
It adds `brag completion zsh|bash|fish` using cobra's built-in completion
generators, plus a tutorial §10 addendum and api-contract section. After
SPEC-024 ships, STAGE-005 closes and PROJ-001 closes.

See the STAGE-005 backlog entry and the `### SPEC-024-specific` subsection
of `STAGE-005-distribution-and-cleanup.md` Design Notes for stage-level
framing. Three same-stage construction precedents (SPEC-021/022/023) make
the §12 trim heuristic applicable: this spec embeds literal artifacts in
Notes for the Implementer and compresses to signatures + invariants elsewhere.

## Goal

Expose cobra's built-in shell completion generators as `brag completion
<shell>`, writing the script to stdout, with help text describing the
per-shell sourcing pattern and a `docs/tutorial.md` §10 addendum for
persistent setup.

## Inputs

- `internal/cli/root.go` — existing cobra root setup; completion wires
  alongside other subcommands
- `cmd/brag/main.go` — the addition site: `root.AddCommand(cli.NewCompletionCmd(root))`
- `docs/tutorial.md` — §10 addendum destination; current last numbered
  section is §9 (Power-user escape hatch), followed by unnumbered
  "Further reading"
- `docs/api-contract.md` — needs a `### brag completion` section appended
  after the existing `### brag stats` section
- `scripts/test-docs.sh` — extend with groups Q (5 asserts) and R (4 asserts)
- `CHANGELOG.md` — `[Unreleased]` section gets a `### Added` entry

## Outputs

**New files:**

- `internal/cli/completion.go` — `NewCompletionCmd(root *cobra.Command) *cobra.Command`
  + unexported `completionRun(root, w, shell)` dispatcher
- `internal/cli/completion_test.go` — 6 tests (Zsh, Bash, Fish,
  UnsupportedShell, NoArgs, HelpShowsSourcingInstructions)

**Modified files:**

- `cmd/brag/main.go` — one line: `root.AddCommand(cli.NewCompletionCmd(root))`
- `docs/tutorial.md` — new `## 10. Shell completions` section inserted
  before the existing "## Further reading" heading
- `docs/api-contract.md` — new `### brag completion` section appended
  after `### brag stats` (before `## Error output`)
- `scripts/test-docs.sh` — groups Q+R (9 new asserts) inserted before
  the finalize block
- `CHANGELOG.md` — `[Unreleased]` → `### Added` → `` `brag completion` ``

**No database changes. No new exports beyond `NewCompletionCmd`.**

**Premise audit (§9 audit-grep cross-check, run at design time):**

`grep -rn "completion" docs/ README.md AGENTS.md` produced one hit:
`docs/CONTEXTCORE_ALIGNMENT.md:35: stable ID, from/to agents, and completion tracking.`
— this is "task completion tracking" (different sense); no update needed.

All other doc updates are enumerated explicitly above.

## Acceptance Criteria

- [ ] AC-1: `brag completion zsh` → stdout contains `#compdef brag`; stderr
      empty; exit 0.
- [ ] AC-2: `brag completion bash` → stdout contains `__start_brag`; stderr
      empty; exit 0.
- [ ] AC-3: `brag completion fish` → stdout contains `complete -c brag`; stderr
      empty; exit 0.
- [ ] AC-4: `brag completion powershell` → exit 1 (ErrUser), error names
      `powershell`, stdout empty.
- [ ] AC-5: `brag completion` (no arg) → cobra error (ExactArgs(1) violation),
      stdout empty.
- [ ] AC-6: `brag completion --help` → stdout contains
      `source <(brag completion zsh)`, `source <(brag completion bash)`, and
      `brag completion fish | source`; stderr empty.
- [ ] AC-7: `brag --help` output includes `completion` in the listed commands.
- [ ] AC-8: `go.mod` unchanged — no new top-level dependency.
- [ ] AC-9: `cmd/brag/main.go` wires `cli.NewCompletionCmd(root)`.
- [ ] AC-10: `docs/tutorial.md` contains `## 10. Shell completions` with
      zsh, bash, and fish sourcing examples.
- [ ] AC-11: `docs/api-contract.md` contains `### \`brag completion` section.
- [ ] AC-12: `scripts/test-docs.sh` groups Q1–Q5 + R1–R4 all pass.
- [ ] AC-13: `go test ./internal/cli/...` passes with the 6 new tests.
- [ ] AC-14: `gofmt -l .` and `go vet ./...` clean.

## Failing Tests

Written during design. The build session's job is to make these pass.

**`internal/cli/completion_test.go`** (`package cli`)

- `TestCompletionCmd_Zsh` — `root.Execute()` with args `["completion", "zsh"]`:
  asserts no error, `errBuf.Len() == 0`, `outBuf.String()` contains
  `"#compdef brag"`. Pairs AC-1.

- `TestCompletionCmd_Bash` — args `["completion", "bash"]`: asserts no error,
  `errBuf.Len() == 0`, `outBuf.String()` contains `"__start_brag"`.
  **Note: the stage Design Notes suggested `_brag_completion()` as the
  bash marker — that is WRONG. Design-time §12(b) verification against
  cobra v1.10.2 shows `__start_brag` is the actual entry-point function
  (appears in both `__start_brag()` function def and the final
  `complete -F __start_brag brag` registration line).** Pairs AC-2.

- `TestCompletionCmd_Fish` — args `["completion", "fish"]`: asserts no error,
  `errBuf.Len() == 0`, `outBuf.String()` contains `"complete -c brag"`.
  Pairs AC-3.

- `TestCompletionCmd_UnsupportedShell` — args `["completion", "powershell"]`:
  asserts `err != nil`, `errors.Is(err, ErrUser)`, `outBuf.Len() == 0`,
  `err.Error()` contains `"powershell"`. Pairs AC-4.

- `TestCompletionCmd_NoArgs` — args `["completion"]`: asserts `err != nil`
  (cobra ExactArgs violation), `outBuf.Len() == 0`. Pairs AC-5.

- `TestCompletionCmd_HelpShowsSourcingInstructions` — args
  `["completion", "--help"]`: asserts no error, `errBuf.Len() == 0`,
  `outBuf.String()` contains each of `"source <(brag completion zsh)"`,
  `"source <(brag completion bash)"`, `"brag completion fish | source"`.
  Pairs AC-6. (§12 NOT-contains self-audit: no NOT-contains assertions
  in this spec; self-audit is a no-op here.)

**`scripts/test-docs.sh`** groups Q+R (see Notes for the Implementer for
the exact bash literal; 9 new assertions):

- Q1: `internal/cli/completion.go` exists
- Q2: `internal/cli/completion_test.go` exists
- Q3: `completion.go` contains `GenZshCompletion`
- Q4: `completion.go` contains `GenBashCompletion`
- Q5: `completion.go` contains `GenFishCompletion`
- R1: `docs/tutorial.md` has `## 10. Shell completions` heading (line-regex)
- R2: `docs/tutorial.md` contains `source <(brag completion zsh)`
- R3: `docs/tutorial.md` contains `source <(brag completion bash)`
- R4: `docs/tutorial.md` contains `brag completion fish | source`

## Locked design decisions

1. **PowerShell skipped.** `cobra.Command.GenPowerShellCompletion` is
   available and would be trivial to add (one switch arm + one test).
   Skipped because bragfile is distributed for macOS+Linux only
   (goreleaser targets darwin+linux); adding Windows completion support to
   a non-Windows-distributed binary is incongruous and creates a support
   surface for an unsupported platform. PowerShell is intentionally
   surfaced as an "unsupported shell" user error so a Windows user gets a
   clear message rather than silently failing. Confidence: 0.90.
   Rejected alternative: ship powershell (trivial cost but wrong signal).

2. **`NewCompletionCmd(root *cobra.Command)` signature.** The cobra gen
   methods (`GenZshCompletion`, `GenBashCompletion`, `GenFishCompletion`)
   must be called on the root command to include all registered
   subcommands in the output. Calling on the completion subcommand itself
   would produce a degenerate script that only completes `brag completion`.
   Passing root through the constructor is the minimal correct signature.
   Confidence: 1.0.
   Rejected alternative: retrieve root via `cmd.Root()` inside RunE —
   equally correct but forces the implementer to trace cobra's call chain
   to understand why; explicit parameter is clearer.

3. **Completion scripts → stdout.** Completion scripts are piped into
   `source` by users; they are data, not human-readable status. Per the
   `stdout-is-for-data-stderr-is-for-humans` constraint. Error messages
   (unsupported shell) surface through the standard RunE error-return
   path and land on stderr via main.go's `fmt.Fprintf(os.Stderr, ...)`.
   Confidence: 1.0.

4. **Presence-of-marker tests, not goldens.** Cobra's completion output
   drifts with version bumps (comment whitespace, internal function names
   not part of the shell protocol). The three marker strings
   (`#compdef brag`, `__start_brag`, `complete -c brag`) are
   structurally load-bearing in their respective shells and effectively
   stable across cobra minor versions. Goldens would require manual
   regeneration on every cobra bump. Confidence: 0.95.

5. **Tutorial addendum as `## 10. Shell completions`.** Adding to the
   existing §9 "Power-user escape hatch" (which covers sqlite3 direct
   access) would muddle two unrelated concepts. A new §10 before
   "Further reading" keeps the sections coherent. The existing P7
   test-docs assertion (§9 body must NOT contain `brew install`) remains
   passing because the new §10 heading shifts out of its awk pattern
   range. Confidence: 0.90.

6. **`api-contract.md` gets `### brag completion` section.** Every other
   command has a dedicated section; omitting it would leave the contract
   incomplete. Confidence: 1.0.

### Rejected alternatives (build-time)

- Auto-sourcing helper (`brag completion --install`): adds complexity;
  the one-liner is copy-paste trivial; deferred to backlog if users
  request it.
- Generating to a file via `--out`: adds a flag; users can redirect
  stdout (`brag completion zsh > ~/.zsh/completions/_brag`). Rejected.
- `cmd.Root()` inside RunE instead of explicit root parameter: correct
  but less transparent. Rejected in favor of explicit parameter (see
  decision 2).

## Implementation Context

### Decisions that apply

No DECs directly govern this spec. Cobra is an existing dependency
(`github.com/spf13/cobra v1.10.2` in go.mod); no new top-level dep is
added. The `no-new-top-level-deps-without-decision` constraint is
satisfied without a DEC.

### Constraints that apply

- `no-new-top-level-deps-without-decision` — cobra already present; zero
  new entries in go.mod. ✓
- `stdout-is-for-data-stderr-is-for-humans` — completion scripts go to
  stdout (`cmd.OutOrStdout()`). ✓
- `test-before-implementation` — 6 failing tests written in design; make
  them pass in build. ✓

### Prior related work

- SPEC-021 (shipped) — README rewrite; established `scripts/test-docs.sh`
  with groups A–G
- SPEC-022 (shipped) — AI-integration assets; extended test-docs.sh with
  groups H–K
- SPEC-023 (shipped) — distribution proper; extended test-docs.sh with
  groups L–P (current tail: P10 at line ~815, finalize block follows)
- PR #21, #22, #23 — same-stage construction precedents for the
  §12 literal-artifact-as-spec pattern

### Out of scope (for this spec specifically)

- PowerShell support — explicitly skipped (decision 1)
- Linuxbrew / apt / Windows distribution — out of scope for PROJ-001
- Auto-sourcing (`brag completion --install`) — backlog
- `--out <file>` flag on completion — redirect with `>` suffices
- PROJ-001 close / STAGE-005 ship — separate prompts after this spec ships

## Notes for the Implementer

Three same-stage construction precedents justify the §12 trim heuristic.
The literal artifacts below are the ground truth; transcribe them verbatim.

### §12(b) design-time verification: cobra v1.10.2 actual markers

Run against cobra v1.10.2 at design time (scratch Go program). Verified:

| Shell | Marker | Source |
|-------|--------|--------|
| zsh   | `#compdef brag` | Line 1 of `GenZshCompletion` output |
| bash  | `__start_brag`  | Function def + final `complete -F __start_brag brag` |
| fish  | `complete -c brag` | Multiple lines in `GenFishCompletion` output |

**The stage Design Notes suggested `_brag_completion()` as the bash marker —
that is incorrect for cobra v1.10.2. Use `__start_brag`.**

### Literal: `internal/cli/completion.go`

```go
package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// NewCompletionCmd returns a subcommand that generates shell completion scripts
// using cobra's built-in generators. root must be the root brag command so the
// generated scripts include all registered subcommands, not just this one.
func NewCompletionCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "completion <shell>",
		Short: "Generate shell completion script",
		Long: `Generate a shell completion script for brag and print it to stdout.

Supported shells: zsh, bash, fish.

To load completions in your current shell session:

  zsh:
    source <(brag completion zsh)

  bash:
    source <(brag completion bash)

  fish:
    brag completion fish | source

To load completions permanently, add the sourcing line above to your
shell's startup file (~/.zshrc, ~/.bashrc, or ~/.config/fish/config.fish).`,
		ValidArgs: []string{"zsh", "bash", "fish"},
		Args:      cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return completionRun(root, cmd.OutOrStdout(), args[0])
		},
	}
}

func completionRun(root *cobra.Command, w io.Writer, shell string) error {
	switch shell {
	case "zsh":
		return root.GenZshCompletion(w)
	case "bash":
		return root.GenBashCompletion(w)
	case "fish":
		return root.GenFishCompletion(w, true)
	default:
		return fmt.Errorf("completion: unsupported shell %q (supported: zsh, bash, fish): %w",
			shell, ErrUser)
	}
}
```

### Literal: `internal/cli/completion_test.go`

```go
package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newCompletionTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	root := NewRootCmd("test")
	root.AddCommand(NewCompletionCmd(root))
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

// TestCompletionCmd_Zsh pairs locked decision §2 (root parameter) and
// verifies the §12(b) zsh marker.
func TestCompletionCmd_Zsh(t *testing.T) {
	root, outBuf, errBuf := newCompletionTestRoot(t)
	root.SetArgs([]string{"completion", "zsh"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	if !strings.Contains(outBuf.String(), "#compdef brag") {
		t.Errorf("zsh completion missing '#compdef brag'; got prefix %q", firstChars(outBuf.String(), 80))
	}
}

// TestCompletionCmd_Bash pairs locked decision §2 and verifies the §12(b) bash
// marker (__start_brag, NOT _brag_completion — design-time verified).
func TestCompletionCmd_Bash(t *testing.T) {
	root, outBuf, errBuf := newCompletionTestRoot(t)
	root.SetArgs([]string{"completion", "bash"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	if !strings.Contains(outBuf.String(), "__start_brag") {
		t.Errorf("bash completion missing '__start_brag'; got %d bytes", outBuf.Len())
	}
}

// TestCompletionCmd_Fish pairs locked decision §2 and verifies the §12(b) fish
// marker.
func TestCompletionCmd_Fish(t *testing.T) {
	root, outBuf, errBuf := newCompletionTestRoot(t)
	root.SetArgs([]string{"completion", "fish"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	if !strings.Contains(outBuf.String(), "complete -c brag") {
		t.Errorf("fish completion missing 'complete -c brag'; got %d bytes", outBuf.Len())
	}
}

// TestCompletionCmd_UnsupportedShell pairs locked decision §1 (powershell
// skipped) and §3 (stdout empty on error).
func TestCompletionCmd_UnsupportedShell(t *testing.T) {
	root, outBuf, _ := newCompletionTestRoot(t)
	root.SetArgs([]string{"completion", "powershell"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unsupported shell, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Errorf("expected ErrUser for unsupported shell, got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected empty stdout on error, got %q", outBuf.String())
	}
	if !strings.Contains(err.Error(), "powershell") {
		t.Errorf("error should name the unsupported shell, got %q", err.Error())
	}
}

// TestCompletionCmd_NoArgs pairs locked decision §2 (ExactArgs enforcement).
func TestCompletionCmd_NoArgs(t *testing.T) {
	root, outBuf, _ := newCompletionTestRoot(t)
	root.SetArgs([]string{"completion"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no shell arg given, got nil")
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected empty stdout on arg error, got %q", outBuf.String())
	}
}

// TestCompletionCmd_HelpShowsSourcingInstructions pairs locked decision §3
// (Long string contains per-shell sourcing pattern).
func TestCompletionCmd_HelpShowsSourcingInstructions(t *testing.T) {
	root, outBuf, errBuf := newCompletionTestRoot(t)
	root.SetArgs([]string{"completion", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
	out := outBuf.String()
	for _, needle := range []string{
		"source <(brag completion zsh)",
		"source <(brag completion bash)",
		"brag completion fish | source",
	} {
		if !strings.Contains(out, needle) {
			t.Errorf("help text missing sourcing instruction %q", needle)
		}
	}
}
```

### Literal: `cmd/brag/main.go` addition

Add ONE line immediately after the existing `root.AddCommand(cli.NewStatsCmd())` line:

```go
root.AddCommand(cli.NewCompletionCmd(root))
```

### Literal: `docs/tutorial.md` §10 addendum

Insert the following block immediately before the `## Further reading`
heading (currently at line ~495). The preceding line is `---` (end of §9).

`````markdown
## 10. Shell completions

`brag completion` generates tab-completion scripts for zsh, bash, and fish.

**zsh** — add to `~/.zshrc`:

```zsh
source <(brag completion zsh)
```

**bash** — add to `~/.bashrc`:

```bash
source <(brag completion bash)
```

**fish** — add to `~/.config/fish/config.fish`:

```fish
brag completion fish | source
```

After sourcing, `brag <tab>` and `brag add --<tab>` show available commands
and flags. Run `brag completion --help` for details.

---
`````

### Literal: `docs/api-contract.md` section

Insert the following block immediately after the `### \`brag stats\`` section
and before `## Error output`:

`````markdown
### `brag completion <shell>` — generate shell completion script (STAGE-005)

```
brag completion zsh|bash|fish
```

Writes a shell completion script to stdout. Supported shells: `zsh`, `bash`,
`fish`.

To source in the current session:

- **zsh:** `source <(brag completion zsh)`
- **bash:** `source <(brag completion bash)`
- **fish:** `brag completion fish | source`

For permanent setup, add the sourcing line to the shell's startup file
(`~/.zshrc`, `~/.bashrc`, or `~/.config/fish/config.fish`).

- Stdout on success: the completion script (pipe to `source` or redirect to a
  file).
- Stderr on success: empty.
- Exit 0 on success; 1 if an unsupported shell name is given (user error); cobra
  arg-count error if the shell arg is omitted.
- No `--db` / `BRAGFILE_DB` dependency — completion generation is stateless.
- PowerShell is not supported (bragfile distributes for macOS + Linux only).
`````

### Literal: `scripts/test-docs.sh` extension

Insert the following block immediately before the `# ===== finalise =====`
comment (currently at line ~816). The new groups append after group P (P10
is the current last assertion).

```bash
# ===== Group Q — completion subcommand source shape =====

# Q1 — internal/cli/completion.go exists
assert_file_exists "Q1" "internal/cli/completion.go"

# Q2 — internal/cli/completion_test.go exists
assert_file_exists "Q2" "internal/cli/completion_test.go"

# Q3 — completion.go wires GenZshCompletion
assert_contains_literal "Q3" "internal/cli/completion.go" "GenZshCompletion"

# Q4 — completion.go wires GenBashCompletion
assert_contains_literal "Q4" "internal/cli/completion.go" "GenBashCompletion"

# Q5 — completion.go wires GenFishCompletion
assert_contains_literal "Q5" "internal/cli/completion.go" "GenFishCompletion"

# ===== Group R — tutorial §10 shell completions addendum =====

# R1 — tutorial §10 heading exists (line-regex avoids substring trap)
if [ ! -f docs/tutorial.md ]; then
    fail "R1" "docs/tutorial.md does not exist"
elif grep -E -q '^## 10\. Shell completions' docs/tutorial.md; then
    ok "R1"
else
    fail "R1" "docs/tutorial.md missing '## 10. Shell completions' heading"
fi

# R2 — tutorial contains zsh sourcing example
assert_contains_literal "R2" "docs/tutorial.md" "source <(brag completion zsh)"

# R3 — tutorial contains bash sourcing example
assert_contains_literal "R3" "docs/tutorial.md" "source <(brag completion bash)"

# R4 — tutorial contains fish sourcing example
assert_contains_literal "R4" "docs/tutorial.md" "brag completion fish | source"

```

### Literal: `CHANGELOG.md` addition

Add the following under `## [Unreleased]`:

```markdown
### Added

- `brag completion` — generate tab-completion scripts for zsh, bash, and fish
  via `brag completion <shell>`. Source into your shell rc for `brag <tab>`
  and flag completion.
```

### Wire-up sequence

1. Write `internal/cli/completion.go` (transcribe literal above).
2. Write `internal/cli/completion_test.go` (transcribe literal above).
3. Run `go test ./internal/cli/...` — confirm 6 new tests fail for the right
   reason (undefined `completionRun` or missing file — not a stray compile
   error from unrelated packages).
4. `go test ./...` — confirm nothing else broke.
5. Add `root.AddCommand(cli.NewCompletionCmd(root))` to `cmd/brag/main.go`.
6. Run `go test ./...` — confirm all pass.
7. Apply tutorial.md, api-contract.md, CHANGELOG.md additions (transcribe
   literals above).
8. Apply test-docs.sh extension (transcribe literal above).
9. Run `just test-docs` — confirm Q1–Q5 + R1–R4 + all prior groups pass.
10. `gofmt -w .` and `go vet ./...` clean.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - (none expected)
- **Deviations from spec:**
  - [list]
- **Follow-up work identified:**
  - Stage-005 ship prompt (STAGE-005-distribution-and-cleanup.md)
  - PROJ-001 close prompt (projects/PROJ-001-mvp/brief.md)

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

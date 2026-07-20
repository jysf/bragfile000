---
# Maps to ContextCore task.* semantic conventions.
# DRAFT — floated 2026-07-16 as an ergonomics backlog candidate. PROJ-006's
# stages are not framed yet; `stage` below is a placeholder to be set when this
# is pulled into a stage. Everything through ## Goal is design-ready; the lower
# sections are a first sketch, not a locked design.

task:
  id: SPEC-070
  type: story
  cycle: design                    # frame | design | build | verify | ship
  blocked: false
  priority: low
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-006
  stage: STAGE-XXX                 # TBD — ergonomics/navigation stage, unframed
repo:
  id: bragfile

agents:
  architect: claude-opus-4-8
  implementer: claude-opus-4-8     # usually same Claude, different session
  created_at: 2026-07-16

references:
  decisions: []                    # will emit DEC-040 (multi-location primary policy)
  constraints: [one-spec-per-pr]
  related_specs: []                # SPEC-031 (project here / ProjectForPath), SPEC-032 (cwd auto-fill), SPEC-024 (completion / shell-script generation)
---

# SPEC-070: `brag project goto` + `brag shell-init` — jump to a project's directory

## Context

Projects already carry their directories: `Project.Locations []string`
(one-to-many, `project_locations`, DEC-017/019/020), and `brag project here`
does the *reverse* today — cwd → project (`ProjectForPath`, SPEC-031). What's
missing is the forward move: **name → directory, then cd there.** This is a
proven, heavily-used pattern (zoxide / autojump / `z`), and the author wants it.

There is a real synergy, not just keystroke savings: `brag add` already
auto-fills `--project` from cwd when inside a registered location (SPEC-032). So
jumping *into* a project's directory is also jumping into the place where
capture auto-tags itself. `goto` closes a small loop between navigation and
frictionless capture — squarely the "capture without discipline" pillar.

**The one real constraint (why this is a spec, not a one-liner):** a subprocess
cannot change its parent shell's working directory. `brag project goto foo` runs
as a child, changes *its own* cwd, and exits — the shell stays put. The
established, correct solution (zoxide/autojump all do this) is two parts:
1. a command that **emits the resolved path** to stdout (nothing else), and
2. a tiny **shell function** (`bragcd`) that runs `cd "$(brag project goto …)"`,
   shipped via `brag shell-init <shell>` — same family as `brag completion`
   (SPEC-024).

Trying to make `goto` literally `cd` would silently no-op and confuse everyone;
the spec's job is to ship the resolver + the wrapper, not the trap.

## Goal

Add `brag project goto <name|id>` that prints a registered project's primary
directory to stdout (exit 1 if unknown or location-less), and `brag shell-init
<zsh|bash|fish>` that emits a `bragcd` shell function wrapping it — so `bragcd
<name>` changes the user's directory to that project.

## Inputs

- **Files to read:**
  - `internal/cli/project.go` — the `brag project` command tree and
    `runProjectHere` (nearest analog; reuse its store-open + resolve shape)
  - `internal/storage/project.go` — `GetProjectByName` (name→project, hydrates
    `Locations`), `GetProject` (id fallback), the `project_locations` ordering
  - `internal/cli/` completion command (SPEC-024) — the shell-script generation
    pattern to mirror for `shell-init`
- **Related code paths:** `internal/storage/entry.go` (`Project` shape)

## Outputs

- **Files created:** `internal/cli/project_goto.go`, `internal/cli/shell_init.go`
  (or fold `shell-init` next to the existing completion command), plus tests.
- **Files modified:** the `brag project` command registration; docs
  (README §Install/Usage, tutorial, api-contract) documenting `goto` + the
  one-line `bragcd` install.
- **New exports:** possibly `Store.PrimaryLocation(projectID)` if a helper reads
  cleaner than sorting `Locations` at the call site.
- **Database changes:** none — pure read over existing schema.

## Acceptance Criteria

- [ ] `brag project goto <name>` prints the project's **primary** location and
      exits 0; stdout is **exactly** the path + a trailing newline, nothing else
      (byte-exact, so `$(...)` wrapping is safe).
- [ ] `<id>` also resolves (name-first, id-fallback — parity with `brag project
      show <name|id>`).
- [ ] Unknown project → exit 1, message on **stderr**, stdout empty.
- [ ] Project exists but has **no** registered location → exit 1, a distinct
      stderr message, stdout empty.
- [ ] Multi-location project → prints the **primary** (first-registered) path
      per DEC-040; `--all` lists every location, one per line, exit 0.
- [ ] `brag shell-init zsh|bash|fish` prints a syntactically valid `bragcd`
      function for that shell to stdout; an unknown shell → exit 1 (user error).
- [ ] The emitted `bragcd` function, sourced, changes directory to the project
      on success and leaves the shell put on failure (non-zero from `goto`).

## Failing Tests

Written during **design**, before build.

- **`internal/cli/project_goto_test.go`**
  - `"goto by name prints primary location, exit 0, byte-exact path\n"`
  - `"goto by id resolves same as name"`
  - `"goto unknown project → exit 1, stderr set, stdout empty"`
  - `"goto project with zero locations → exit 1, distinct message"`
  - `"goto multi-location prints first-registered (DEC-040)"`
  - `"goto --all lists every location, one per line"`
- **`internal/cli/shell_init_test.go`** (golden per shell)
  - `"shell-init zsh emits bragcd (golden)"` / bash / fish
  - `"shell-init unknown-shell → exit 1 user error"`

## Implementation Context

### Decisions that apply / to emit

- **DEC-040 (to emit) — multi-location primary policy.** When a project has
  more than one location, `goto` returns the **first-registered** (lowest
  `project_locations.position`/rowid). Rationale: deterministic, matches the
  `new --path` seed, zero prompts, keeps the stdout-is-just-a-path contract
  clean; `--all` is the escape hatch for the multi-dir case. Rejected: an
  interactive picker (breaks the pipe contract), and "most-recent-brag dir"
  (needs a join and a frecency model — see Out of scope).

### Constraints that apply

- `one-spec-per-pr`.
- Stdout of `goto` is a machine contract (a shell `$(...)` consumes it) — no
  decoration, no `--format` needed; errors and diagnostics go to stderr.

### Prior related work

- `brag project here` (SPEC-031) — the reverse resolver; copy its store-open +
  resolve + exit-1-on-no-match shape.
- `brag completion` (SPEC-024) — the shell-script generation precedent for
  `shell-init`.
- cwd auto-fill (SPEC-032) — the synergy `goto` completes.

### Out of scope (for this spec specifically)

- Fuzzy / frecency matching (zoxide-style ranking by recent use) — deferred; a
  possible follow-up once it earns its keep.
- Project-name tab-completion for `bragcd <TAB>` — desirable, but its own spec.
- Any change to how locations are stored or ordered.

## Notes for the Implementer

- Keep `goto` output to the bare path. The single most common failure mode for
  cd-wrappers is a stray log line or prompt on stdout breaking `$(...)`.
- `bragcd` name: fixed for now (don't over-parameterize). Fish syntax differs
  from zsh/bash — the golden tests pin all three.
- Suggested wrapper shape (zsh/bash):
  `bragcd() { local d; d="$(brag project goto "$1")" || return; cd "$d"; }`

---

## Build Completion

*Filled at the end of the build cycle, before verify.*

- **Branch:**
- **PR (if applicable):**
- **All acceptance criteria met?** yes/no
- **New decisions emitted:**
  - `DEC-040` — multi-location primary policy (if emitted as specified)
- **Deviations from spec:**
- **Follow-up work identified:**

---

## Reflection (Ship)

*Appended during the **ship** cycle. Outcome-focused reflection, distinct
from the process-focused build reflection above.*

1. **What would I do differently next time?**
   — <answer>

2. **Does any template, constraint, or decision need updating?**
   — <answer>

3. **Is there a follow-up spec I should write now before I forget?**
   — <answer>

4. **What can a user do now that they couldn't before?** — one sentence,
   before → after; quote the confirming number if one exists, name the outcome
   if not. Write `none` if this spec has no user-visible outcome — that is a
   real, greppable result, not a blank. This is the line a brag's `impact` field
   is transcribed from, and both halves are already written above (## Context is
   the before, ## Goal is the after): confirm the prediction, don't reconstruct
   it from memory.

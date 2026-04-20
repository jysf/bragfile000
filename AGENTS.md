# AGENTS.md — Claude-Only Variant

Instructions for Claude working across all phases of this repository. Read this file first, every session.

> This variant assumes Claude plays every role: architect, implementer, reviewer. The context normally in a handoff document lives inside each spec's `## Implementation Context` section.

> This file contains conventions only. For rules/constraints, see `/guidance/constraints.yaml`. For architectural rationale, see `/decisions/`. For waves of work against this app, see `/projects/`.

---

## 1. Repo Overview

- **Repo (the app):** Bragfile (`brag` CLI, homebrew formula `bragfile`)
- **Purpose:** Local-first Go CLI for engineers to capture and retrieve career accomplishments ("brags") for retros, reviews, and resumes. Stores entries in embedded SQLite at `~/.bragfile/db.sqlite`.
- **Primary stakeholders:** solo developer (jysf). No other stakeholders in PROJ-001.
- **Active project:** `PROJ-001-mvp` — ship a usable, distributable MVP via `brew install bragfile` within ~2 weeks.

See `.repo-context.yaml` for structured metadata.

---

## 2. Work Hierarchy

```
REPO (the app — persists across all projects)
 └─ PROJECT (a wave of work: "MVP", "improvements", "v2 redesign")
     └─ STAGE (a coherent chunk within a project)
         └─ SPEC (an individual task)
```

- The **repo** is the app. `AGENTS.md`, `/docs/`, `/guidance/`,
  `/decisions/` live at repo level because they accumulate across all
  projects.
- A **project** (`/projects/PROJ-*/`) is a bounded wave of work.
- A **stage** is an epic-sized chunk within a project (2–5 per project).
- A **spec** is a single implementable task. Belongs to one stage in
  one project.

In this variant, Claude plays architect and implementer in **separate
sessions**. The spec file itself carries all the context — see its
`## Implementation Context` section.

**Decisions persist at repo level.** A decision made during PROJ-001
binds PROJ-002 as well.

**Specs do not cross project boundaries.**

---

## 3. Tech Stack

- **Language:** Go
- **Runtime:** Go 1.26.x (latest stable; pinned via `go.mod` `go 1.26`). Update via `brew upgrade go`.
- **Framework:** `spf13/cobra` for CLI argv + subcommands.
- **Database:** SQLite 3, embedded, accessed via `modernc.org/sqlite` (pure Go, **no CGO**). See DEC-001.
- **Testing:** Go stdlib `testing` package. Storage tests use `t.TempDir()` (enforced by `storage-tests-use-tempdir` constraint).
- **Linter / Formatter:** `gofmt` (enforced) + `go vet`. `golangci-lint` welcome but not required in CI yet.
- **Hosting:** None. Local CLI only.
- **Distribution:** `goreleaser` → GitHub Releases → homebrew tap at `github.com/jysf/homebrew-bragfile` (arriving in STAGE-004).
- **CI:** GitHub Actions (to be set up in STAGE-004). Must run `go test ./...` and `gofmt -l .` and fail on any diff.

---

## 4. Commands (exact)

These are the APP's commands. For template/workflow commands, see `justfile`.

```bash
# --- once, to bootstrap (SPEC-001 creates go.mod; nothing to install until then) ---
go mod download                              # install module deps after go.mod exists

# --- daily development ---
just build                                   # build ./brag in repo root (wraps `go build ./cmd/brag`)
just install                                 # install brag to ~/go/bin (wraps `go install ./cmd/brag`)
just uninstall                               # remove installed brag binary
just test                                    # run all tests (wraps `go test ./...`)
just run -- list                             # run without installing (e.g. `just run list`, `just run add --title "x"`)
# direct-go equivalents are also fine:
go build ./cmd/brag                          # build binary into ./brag
go install ./cmd/brag                        # install to $GOBIN (default ~/go/bin)
go run ./cmd/brag <args>                     # run without building
go test ./...                                # run all tests
go test ./internal/storage -run TestAdd -v   # run a single test in a package
gofmt -w .                                   # format (write in place)
gofmt -l .                                   # lint (list unformatted files; CI fails if non-empty)
go vet ./...                                 # static checks

# --- release (STAGE-004) ---
goreleaser release --clean                   # full release (tag triggers this in CI)
goreleaser build --snapshot --clean          # local cross-compile smoke test

# --- runtime (once built) ---
./brag --version
./brag add --title "..."
./brag list
```

macOS note: `brew upgrade go` to move to the latest Go. `brew install goreleaser` once we hit STAGE-004.

---

## 5. Directory Structure

```
/
├── AGENTS.md                          # This file
├── CLAUDE.md                          # Pointer to AGENTS.md
├── README.md                          # Human-facing readme
├── GETTING_STARTED.md                 # First-project walkthrough
├── FIRST_SESSION_PROMPTS.md           # Phase prompts
├── .repo-context.yaml                 # Repo (app) metadata
├── .variant                           # "claude-only"
├── justfile                           # Commands: just status, just new-spec, etc.
├── scripts/                           # Shell scripts powering justfile
├── docs/                              # Architecture, data model, API contract
├── guidance/                          # Repo-level rules (across all projects)
│   ├── constraints.yaml
│   └── questions.yaml
├── decisions/                         # Repo-level DEC-* (across all projects)
├── projects/                          # Waves of work
│   ├── _templates/                    # Shared templates
│   │   ├── spec.md
│   │   ├── stage.md
│   │   └── project-brief.md
│   ├── PROJ-001-<slug>/
│   │   ├── brief.md
│   │   ├── stages/
│   │   └── specs/
│   │       └── done/
│   └── PROJ-002-<slug>/
├── cmd/
│   └── brag/                          # CLI entrypoint (main package)
└── internal/                          # implementation packages (not importable externally)
    ├── cli/                           # one file per subcommand; no SQL here
    ├── config/                        # --db / BRAGFILE_DB / default resolution
    ├── storage/                       # Store, Entry, embedded migrations
    │   └── migrations/                # NNNN_*.sql, embedded via embed.FS
    ├── editor/                        # (STAGE-002) $EDITOR launch + markdown parse
    └── export/                        # (STAGE-003) markdown + sqlite-file exporters
```

---

## 6. Cycle Model

Every spec moves through five cycles. **Cycles are tags, not gates**.

| Cycle | Purpose |
|---|---|
| **frame** | Go/no-go on the spec |
| **design** | Write the spec + failing tests + implementation context |
| **build** | Make failing tests pass |
| **verify** | Review + validation in one pass |
| **ship** | Merge, deploy, reflect, archive |

Valid transitions:
```
frame → design → build → verify → ship
                   ↑       │
                   └───────┘ (verify sends back on punch list)
```

**In this variant**, use **separate Claude sessions** for each cycle.
A fresh session prevents design-phase context from contaminating build
decisions, and a fresh verify session catches drift a continuation
session wouldn't.

Project and stage lifecycles are lighter:
- **Project status:** `proposed | active | shipped | cancelled`
- **Stage status:** `proposed | active | shipped | cancelled | on_hold`

---

## 7. Cross-Reference Rules

Every spec has these relationships, encoded in front-matter:
- `project.id` → the project it belongs to
- `project.stage` → the stage within that project
- `references.decisions` → DEC-* it was designed against
- `references.constraints` → constraints that apply

DECs are stable; specs come and go. DECs don't reciprocally list specs.

---

## 8. Coding Conventions

- **Naming:** standard Go. Exported identifiers in `CamelCase`, unexported in `camelCase`. Package names are short, lowercase, no underscores (`cli`, `storage`, `config`).
- **File organization:** one subcommand per file under `internal/cli/` (`add.go`, `list.go`, …). One concept per file under `internal/storage/` (`store.go`, `entry.go`, `migrate.go`).
- **Imports:** `gofmt`-sorted. Grouped by stdlib / third-party / internal with a blank line between groups.
- **Error handling:** wrap with context: `fmt.Errorf("add entry: %w", err)`. Return errors; don't panic. No custom error types in PROJ-001 unless a DEC justifies one.
- **Logging:** none in MVP. If we ever add structured logging it gets a dedicated DEC. CLI talks to users via stdout/stderr directly (see `stdout-is-for-data-stderr-is-for-humans` constraint).
- **No SQL outside `internal/storage`.** Enforced by the `no-sql-in-cli-layer` constraint.
- **Comments:** Explain *why*, not *what*. No doc comments on unexported helpers unless they are non-obvious.
- **No dead code.** Delete, don't comment out.

---

## 9. Testing Conventions

- Every new exported function gets at least one test.
- Test file naming: Go convention — `foo_test.go` next to `foo.go` in the same package.
- Storage tests use `t.TempDir()` for the DB path. Never touch `~/.bragfile`.
- CLI command tests construct a `*cobra.Command` with in-memory buffers for stdout/stderr and assert on them, not on the host's terminal. Use **separate `outBuf` and `errBuf`** (`cmd.SetOut(&outBuf)` / `cmd.SetErr(&errBuf)`) and assert no cross-leakage — e.g., a `--version` test must assert `errBuf.Len() == 0` in addition to the stdout substring check. This enforces the `stdout-is-for-data-stderr-is-for-humans` constraint at the test layer. Lesson earned in SPEC-001 verify punch list (2026-04-20).
- Time-based ordering tests must use a monotonic tie-break column (e.g., `id DESC` under DEC-005) in addition to `created_at DESC`. Sleep-based timestamp separation alone is insufficient because RFC3339 is second-precision — a test that sleeps 10ms between rows will see identical `created_at` strings and produce non-deterministic ordering without a tie-break. Lesson earned in SPEC-002 build reflection (2026-04-20).
- Coverage expectations: no hard threshold. Every storage method and every command has at least one happy-path and one error-path test. Migration runner has a "runs twice is a no-op" test.
- **TDD:** Tests live in the spec's `## Failing Tests` section, written
  during **design**, made to pass during **build**. Enforced by the `test-before-implementation` constraint.
- In build: after writing the failing tests and before touching implementation, run `go test ./...` once and confirm the tests fail for the *expected* reason (the assertion you wrote, not a stray compilation error or undefined-symbol). Catches spec defects at the cheapest moment. If fail-first reports an "unexpectedly passing" test, investigate before proceeding — usually the assertion is too weak (see next bullet). Lesson earned in SPEC-003 Q3 ship reflection and validated by SPEC-004 build (2026-04-20).
- When a test asserts that help or documentation output contains a substring, pick a token that is unique to the content under test — e.g. a distinctive example phrase or an explicit label like `"Examples:"` — not a generic word that cobra or another auto-rendering layer may already produce. Generic substring asserts give false-positive passes: SPEC-005's `TestAdd_HelpShowsExamples` asserted `outBuf` contained `"brag add"` but cobra's `Usage: brag add [flags]` line already contains that string, so the test would have passed without the implementation. Lesson earned in SPEC-005 build reflection (2026-04-20).

---

## 10. Git and PR Conventions

- **Branch:** `feat/spec-NNN-<slug>` for feature specs. `fix/spec-NNN-<slug>` for bug specs. `chore/<slug>` is acceptable for non-spec infra work (e.g., CI changes).
- **One spec per branch, one PR per branch.** Enforced by the `one-spec-per-pr` constraint.
- **Commits:** Conventional-style short subject. `feat(storage): add Entry type`, `fix(cli): exit 1 when --title empty`, `docs: update architecture.md`, `chore(ci): add go-test job`. Body optional; when present, explain *why*. Avoid "WIP" commits on branches that will be opened as PRs — squash or rebase first.
- **PR description must include:**
  - Project: `PROJ-001`
  - Stage: `STAGE-NNN`
  - Spec: `SPEC-NNN`
  - Decisions referenced, constraints checked, new `DEC-*` files
- Do not force-push to `main`. Branch protection on `main` should require PR + passing CI.
- **`.gitignore` for compiled binaries must be anchored.** Write `/brag`, not `brag` — an unanchored pattern matches every `brag` path at any depth and will silently shadow a directory of the same name (e.g., `cmd/brag/`). Lesson earned in SPEC-003 build reflection (2026-04-20, when SPEC-001's unanchored `brag` pattern hid `cmd/brag/main.go`).

---

## 11. Domain Glossary

- **brag / entry** — one captured moment worth remembering for a retro, review, or resume. Single row in the `entries` table.
- **capture** — the act of creating an entry, either via `brag add --title ...` (flags form) or `brag add` with no args opening `$EDITOR` on a templated markdown buffer (STAGE-002).
- **Store** — the `*storage.Store` Go type that owns the `*sql.DB` and all typed methods. The only package that imports a SQL driver.
- **migration** — a single `NNNN_*.sql` file under `internal/storage/migrations/`, embedded into the binary, applied automatically in lexical order on `storage.Open`.
- **export** — a one-shot dump of entries, either as a Markdown report (stdout or `--out file.md`) or as a portable SQLite file copy (via `VACUUM INTO`).
- **summary** — a rule-based (non-LLM) aggregation of entries grouped by project/type over a time range. STAGE-003.
- **tap** — a homebrew tap repo (`github.com/jysf/homebrew-bragfile`) hosting the `bragfile.rb` formula. Created in STAGE-004.

---

## 12. Cycle-Specific Rules

### During **build**

Start a **new Claude session**. Do not continue from the design session.

Before writing code:
1. Read the spec's `## Implementation Context` section.
2. Read every `DEC-*` it references.
3. Read the parent `STAGE-*.md` and project `brief.md`.
4. Read `/guidance/constraints.yaml`.
5. If anything is ambiguous, add to `/guidance/questions.yaml` and stop.

When done:
1. Fill in spec's `## Build Completion` (including reflection).
2. `just advance-cycle SPEC-NNN verify`.
3. Create `DEC-*` files for non-trivial build decisions.
4. Open PR.

### During **verify**

Start **another new Claude session**. Do not reuse build session.

Check: acceptance criteria met? tests pass? no decision drift? no
constraint violations? non-trivial choices have DEC-*? build reflection
answered honestly?

Output: ✅ APPROVED / ⚠ PUNCH LIST / ❌ REJECTED.

### During **ship**

Append `## Reflection` to spec. Three answers. Then
`just archive-spec SPEC-NNN`. If stage backlog is complete, run the
Stage Ship prompt.

---

## 13. Session Hygiene (claude-only specific)

Because one agent plays multiple roles, context contamination is a real
risk. Four habits keep it at bay:

1. **New session per cycle where possible.** Especially design → build
   and build → verify.
2. **Never reference "as I said earlier"** in later cycles. The spec
   is the source of truth.
3. **Weekly review is non-optional.** Without a second agent pushing
   back, drift compounds silently. Run `just weekly-review`.
4. **Honest confidence values on decisions.** See Section 14.

---

## 14. Confidence Discipline

Decisions have an `insight.confidence` field (0.0–1.0). Honest values drive:

- **Design:** decisions at confidence < 0.7 also create a question in
  `/guidance/questions.yaml`.
- **Verify:** specs referencing decisions at confidence < 0.6 get a
  yellow flag.
- **Weekly review:** all decisions < 0.8 are listed with strength/weakness trend.

Most decisions should land between 0.7 and 0.95. 1.0 only for truly locked choices.

---

## 15. Pointers

- Constraints: `/guidance/constraints.yaml`
- Open questions: `/guidance/questions.yaml`
- Decisions: `/decisions/`
- Projects: `/projects/`
- Templates: `/projects/_templates/`
- Architecture: `/docs/architecture.md`
- Data model: `/docs/data-model.md`
- CLI contract: `/docs/api-contract.md`
- User tutorial (how to use `brag`): `/docs/tutorial.md`
- Phase prompts: `/FIRST_SESSION_PROMPTS.md`
- First walkthrough: `/GETTING_STARTED.md`
- Daily commands: run `just --list`

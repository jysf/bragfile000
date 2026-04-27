# Pre-distribution security review — bragfile v0.1.0

**Date:** 2026-04-26
**Branch reviewed:** `main` at `1582572` (post-merge of PR #23 + post-fix punch-list)
**Reviewer:** coordinator session (Claude Opus 4.7) + Tier 2 subagent + `/security-review` skill + `govulncheck`
**State at time of review:** no tags pushed, no releases published, no PAT created, no users
**Trigger:** user-requested pause of the Homebrew deploy specifically to do a full security analysis before any binary lands on user machines

---

## Executive summary

**Posture: solid for a personal-scale Homebrew CLI.**

- **0 critical, 0 high, 3 medium, 5 low, ~13 info findings.**
- **No findings met the `/security-review` skill's strict ≥8-confidence threshold** with its hard exclusions applied — i.e. no concrete exploitable vulnerabilities introduced in the recent SPEC-022 + SPEC-023 distribution work.
- **`govulncheck ./...` clean** — no vulnerabilities in the Go module graph.
- **Parameterized SQL throughout, narrow FTS5 input handling, no panic surface, no shell-string concatenation around `$EDITOR`, no network I/O, no unsafe deserialization.**

Findings cluster around **hardening** (action SHA-pinning, file mode tightening, JSON schema length limits) and **operational controls** (PAT scoping, tap branch unprotected). None are blockers; all are recommended pre-tag-push fixes ranked below.

**Highest-leverage fixes before `git tag v0.1.0`:**
1. Tighten DB file mode from 0644 → 0600 ([internal/storage/store.go:Open](../../../internal/storage/store.go))
2. Tighten export file mode from 0644 → 0600 ([internal/cli/export.go:142](../../../internal/cli/export.go))
3. Pin `goreleaser/goreleaser-action@v6` to a commit SHA ([.github/workflows/release.yml:34](../../../.github/workflows/release.yml))
4. Add `maxLength` constraints to [docs/brag-entry.schema.json](../../../docs/brag-entry.schema.json)
5. Create the `HOMEBREW_TAP_GITHUB_TOKEN` PAT with minimum scope (see recommendation below)

**Ongoing automated coverage enabled post-review** (see "GitHub Advanced Security features enabled" section): CodeQL Code Scanning (Go + Actions), Dependabot security alerts, Dependabot version updates (gomod + github-actions), Secret scanning + push protection.

---

## Methodology

Three parallel review tracks, plus one user-driven track still pending:

| Track | Reviewer | Scope | Status |
|---|---|---|---|
| Tier 1A | coordinator (direct) | GitHub Actions workflows, `.goreleaser.yaml`, tap-repo ACLs, PAT scope recommendation | ✅ done |
| Tier 1B | coordinator (direct) | Privacy posture: SQLite file mode, telemetry surface, error/panic leakage | ✅ done |
| Tier 1C | coordinator (direct) | Repo hygiene: committed-secrets scan, `.gitignore`, untracked-files content | ✅ done |
| Tier 2 | general-purpose subagent (background) | Binary attack surface (SQL/FTS5/path/editor/JSON), supply chain (deps), AI-integration paths (SPEC-022 hook + schema + slash-command) | ✅ done |
| `/security-review` skill | skill-invoked sub-agent | Independent third-pass with strict ≥8-confidence + hard-exclusion filter | ✅ done — no qualifying findings |
| `govulncheck ./...` | tool | Go vulnerability database lookup against current module graph | ✅ done — clean |
| `/ultrareview` | user-driven (billed) | Multi-agent cloud review of current `main` | ⏭️ skipped — sufficient coverage from prior tracks (see section below) |

The three Claude-driven tracks (Tier 1, Tier 2, `/security-review`) ran with mostly non-overlapping scope by design — overlap was used as cross-validation (DB file mode, action pinning) rather than work duplication.

---

## Findings (sectioned by severity)

### MEDIUM (3)

#### M1 — SQLite DB file is readable by other local users

- **Location:** [internal/storage/store.go:28-32](../../../internal/storage/store.go) — `os.MkdirAll(filepath.Dir(path), 0o755)` creates `~/.bragfile/` group+other-readable; the SQLite file inside inherits the `modernc.org/sqlite` driver default of `0o644`.
- **Vector:** On a shared/multi-user macOS or Linux host, a stolen laptop without full-disk encryption, or any system where the user's home is not strictly per-user-readable, any other local user can read the brag corpus. Brag entries are personal professional reflection content — names of projects, evaluative narrative, possibly performance-cycle context. On a personal single-user laptop the exposure is largely cosmetic, hence medium not high.
- **Fix:** create the directory `0o700` and chmod the DB file to `0o600` immediately after creation:
  ```go
  // in storage.Open, before sql.Open:
  if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil { ... }
  // ensure the DB file exists with 0o600 before the lazy driver creates it 0o644:
  f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
  if err != nil { ... }
  _ = f.Close()
  // then proceed with sql.Open
  ```
  Add a test that asserts the post-Open mode equals `0o600`.

#### M2 — Third-party action `goreleaser/goreleaser-action@v6` floats

- **Location:** [.github/workflows/release.yml:34](../../../.github/workflows/release.yml)
- **Vector:** If goreleaser's GitHub org is compromised and a malicious `v6` is published (or the `v6` tag is moved to a malicious commit), the next tag-triggered release pulls the malicious action and can exfiltrate `HOMEBREW_TAP_GITHUB_TOKEN`. With the PAT, the attacker pushes a malicious cask file to `github.com/jysf/homebrew-bragfile`; subsequent `brew install jysf/bragfile/bragfile` users get a backdoored binary. Probability low (well-known maintainer); impact high (tap takeover → user-machine compromise).
- **Fix:** pin to a commit SHA. Look up the current `v6.x.y` release SHA on `github.com/goreleaser/goreleaser-action/releases` and use:
  ```yaml
  uses: goreleaser/goreleaser-action@<40-char-sha>  # v6.x.y
  ```
  Add Renovate or Dependabot to keep the SHA current with PR review.

#### M3 — Tap repo `main` branch is unprotected

- **Location:** `github.com/jysf/homebrew-bragfile`, branch `main` (`protected: false` per `gh api repos/jysf/homebrew-bragfile/branches`)
- **Vector:** Compromise of `HOMEBREW_TAP_GITHUB_TOKEN` (token leak via local `git credential` store, malicious goreleaser-action, etc.) → attacker pushes malicious cask file directly to `main` → users `brew install` get the malicious binary. No PR review interposes between PAT compromise and a poisoned cask.
- **Fix:** Standard branch protection (require PR + reviews) would BLOCK goreleaser's direct push, so it's not viable. Compensating controls:
  1. Tight PAT scope (see below — fine-grained PAT, single repo, Contents:r/w only).
  2. Short PAT expiration (90 days max).
  3. Periodic audit: `gh api user/keys` and the user's PAT list; rotate PAT after any suspicion of local-machine compromise.
  4. Watch the tap repo for unexpected commits: enable "Watch all activity" in the tap repo's notifications, or automate via GitHub Actions running on tap-repo pushes.

### LOW (5)

#### L1 — `brag export --out FILE` writes 0644

- **Location:** [internal/cli/export.go:142](../../../internal/cli/export.go) — `os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)`
- **Vector:** Same shape as M1 but for user-named export files. `brag export --out /tmp/review.json` leaks the export to all local users.
- **Fix:** change `0o644` → `0o600`. One-character edit. The user can chmod up if they want to share.

#### L2 — Editor temp file leaks on SIGINT/SIGKILL

- **Location:** [internal/editor/launch.go:34-40](../../../internal/editor/launch.go) — `os.CreateTemp("", "brag-edit-*.md")` + `defer os.Remove(path)`
- **Vector:** Ctrl-C in the editor (or editor crash that kills the whole process group) bypasses the `defer`. The temp file (containing in-progress entry text — possibly sensitive) is left in `$TMPDIR` and survives in `/var/folders/...` for the duration of macOS retention. The file is mode 0600 (Go stdlib default for `os.CreateTemp`) so the privacy exposure is bounded to the same user, not other users.
- **Fix:** install a `signal.Notify(SIGINT, SIGTERM)` handler in `Launch` that removes the temp file on signal then re-raises. Borderline gold-plate for a personal CLI; defer is also a reasonable choice. If deferred, document in `--help`.

#### L3 — `io.ReadAll` on editor buffer is unbounded

- **Location:** [internal/editor/editor.go:89](../../../internal/editor/editor.go) — `remaining, err := io.ReadAll(tp.R)`
- **Vector:** Local trust boundary — a user with a malicious or buggy `$EDITOR` could write a multi-GB file and OOM the process. Since `$EDITOR` is trusted (precedent: env vars are trusted inputs), this is not exploitable by an external attacker. Worth wrapping defensively.
- **Fix:** wrap with `io.LimitReader`:
  ```go
  remaining, err := io.ReadAll(io.LimitReader(tp.R, 10<<20))  // 10 MiB cap
  ```
  Reject exact-cap reads as "buffer too large" to distinguish a legitimate-but-huge entry from a malicious truncation.

#### L4 — JSON schema lacks `maxLength` constraints

- **Location:** [docs/brag-entry.schema.json](../../../docs/brag-entry.schema.json)
- **Vector:** A malformed-but-strict-schema-compliant JSON payload with a 100 MB `description` field validates and gets persisted via `INSERT`. Becomes part of FTS5 index; bloats the corpus. Not a security vulnerability per se — the user controls their own input — but the AI-integration hook (`scripts/claude-code-post-session.sh`) currently dumps the entire session SUMMARY into `description`, making accidental over-large payloads plausible.
- **Fix:** add `maxLength` constraints to the schema:
  - `title`: 200
  - `tags` / `project` / `type`: 64
  - `impact`: 256
  - `description`: 100000 (100 KB)

  And add matching length checks in [internal/cli/add_json.go:parseAddJSON](../../../internal/cli/add_json.go) — the binary is the authoritative validator; the schema alone is documentation.

#### L5 — GitHub-owned actions float (`@v4`, `@v5`)

- **Location:** [.github/workflows/ci.yml:25,28](../../../.github/workflows/ci.yml), [.github/workflows/release.yml:22,27](../../../.github/workflows/release.yml)
- **Vector:** Same shape as M2, lower probability (GitHub-verified-creator program; `actions/checkout` and `actions/setup-go` are heavily monitored).
- **Fix:** SHA-pin alongside the M2 fix. Same Renovate/Dependabot story.

### INFO (selected)

- **I1** — No `timeout-minutes` on release workflow. Recommend `timeout-minutes: 30` to prevent runaway-cost stalls.
- **I2** — No `concurrency:` block on release workflow. Consider `concurrency: { group: release, cancel-in-progress: false }` as cheap insurance.
- **I3** — No code signing / SLSA provenance. Acceptable to defer; if you later want canonical "shipped right" demonstration, goreleaser supports `signs:` (cosign/gpg) and SLSA out of the box.
- **I4** — Persistent untracked files visible to `git status`: `framework-feedback/`, `revew1.md`, `status-after-nine-specs.md`. Verified content; no secrets. Either commit to `docs/` or add to `.gitignore`. No security risk.
- **I5** — Goreleaser binary version `~> v2` floats within v2.x at [.github/workflows/release.yml:37](../../../.github/workflows/release.yml). Pin to specific patch (e.g. `version: '2.15.4'`) for reproducibility.
- **I6** — Schema documentation contract: `additionalProperties: false` is correctly enforced both in the schema and at the runtime parser via `dec.DisallowUnknownFields()`.
- **I7** — Hook script trust boundary (SPEC-022): `scripts/claude-code-post-session.sh` correctly delegates JSON construction to `jq -n --arg` (safe escaping) and does NOT auto-execute `brag add`. Approval gate is the user pasting the JSON manually. Worth a one-line note in `BRAG.md` that "approval means you read the JSON, not just hit y" — defends against cargo-cult approval of prompt-injected entries.

---

## Clean areas (verified, no findings)

- ✅ **No real secrets in git history.** Targeted scan for `ghp_*`, `gho_*`, `github_pat_*`, `AKIA*`, `BEGIN ... PRIVATE KEY`, password/token patterns. Single regex match was a SPEC-001 example file talking about *redacting* `password`/`token` field names — fake data, not credentials.
- ✅ **Zero network surface in the binary.** No `net/http`, `http.`, `url.`, `net.Dial` imports anywhere in `internal/` or `cmd/`. Truly local-first; no telemetry path.
- ✅ **No `panic(` calls** in user-input paths (or anywhere in `internal/` and `cmd/`). Only `os.Exit(1/2)` at top of [cmd/brag/main.go:36-38](../../../cmd/brag/main.go) for top-level error handling. No goroutine-stack leakage on user input.
- ✅ **`.gitignore` covers basics.** `.env`, `.env.local`, `.env.*.local`, `*.pem`, `*.key`, build artifacts, IDE files, `.DS_Store`.
- ✅ **`permissions: contents: read` on ci.yml** (least privilege at workflow level). Fork PRs run with read-only token; no secrets exposed; cannot exfiltrate.
- ✅ **`permissions: contents: write` on release.yml** (minimum needed for release publishing).
- ✅ **`HOMEBREW_TAP_GITHUB_TOKEN` sourced from `secrets.*`**, never embedded in YAML.
- ✅ **`skip_upload: auto` correctly tied to `release.prerelease: auto`** in `.goreleaser.yaml`. RC tags don't push to tap; production tags do.
- ✅ **SQL: every query in `internal/storage/*.go` uses parameterized `?` placeholders.** The dynamic `WHERE` builder in `Store.List` joins with literal `" AND "`, with all user values bound separately. The `Tag` filter constructs `"%," + tag + ",%"` but binds it as a parameter, not inline SQL.
- ✅ **FTS5 input handling in [internal/cli/search.go:43-56](../../../internal/cli/search.go)** — `buildFTS5Query` rejects any `"`, splits on whitespace, and wraps each token in double quotes. Without `"` an attacker cannot terminate the phrase to inject FTS5 operators or column qualifiers; phrase contents are interpreted literally.
- ✅ **JSON parser** uses `DisallowUnknownFields()`, custom `tagsField` enforcing DEC-004 string-not-array, and trailing-garbage detection. Server-owned fields (`id`, `created_at`, `updated_at`) are `json.RawMessage` and discarded.
- ✅ **`$EDITOR` argv handling** uses `strings.Fields` + `exec.Command(argv[0], argv[1:]...)` with no `sh -c`. `EDITOR='vi; whoami'` becomes a failed exec of literal `vi;`, not a shell injection.
- ✅ **Editor tempfile** created via `os.CreateTemp` (mode 0600, randomized name). No symlink-followup or TOCTOU vector.
- ✅ **Migrations** are embedded `*.sql` files only; no user input feeds the migrator.
- ✅ **Direct + transitive Go deps all reputable.** `github.com/spf13/cobra`, `modernc.org/sqlite`, `github.com/google/uuid`, `golang.org/x/sys`, `github.com/spf13/pflag`, modernc.org family, mattn/go-isatty, dustin/go-humanize, ncruces/go-strftime, remyoudompheng/bigfft (transitive pseudo-version, used by modernc, h1-verified). No surprises.
- ✅ **`govulncheck ./...`** — *No vulnerabilities found.*
- ✅ **LICENSE present at repo root** (1056 bytes; bundled into archives per goreleaser config).

---

## PAT scope recommendation (since not yet created)

When you create `HOMEBREW_TAP_GITHUB_TOKEN`, use a **fine-grained PAT** (not classic):

- **Resource owner:** `jysf`
- **Repository access:** "Only select repositories" → check `homebrew-bragfile` ONLY (do NOT grant access to `bragfile000` or anything else)
- **Repository permissions:**
  - **Contents:** Read and write
  - **Metadata:** Read-only (auto-required)
  - **All others:** No access
- **Expiration:** 90 days max — set a calendar reminder to rotate
- **Name:** descriptive, e.g. `bragfile-tap-publisher`

**Avoid classic PAT.** Classic PATs cannot be scoped to a single repo; the closest scope is `public_repo`, which still grants write to all your public repos. If you must use classic, do not select `repo` (full access including private).

After creation:
- Settings → Secrets and variables → Actions → New repository secret
- Name: `HOMEBREW_TAP_GITHUB_TOKEN` (exact)
- Value: the PAT

---

## Prioritized fix list (recommended order before `git tag v0.1.0`)

| # | Effort | Severity | Fix |
|---|---|---|---|
| 1 | 5 min | Medium | M1 — DB file mode → 0600 in [internal/storage/store.go:Open](../../../internal/storage/store.go); add chmod-after-create + a test asserting the post-Open mode |
| 2 | 1 min | Low | L1 — Export file mode `0o644` → `0o600` in [internal/cli/export.go:142](../../../internal/cli/export.go) |
| 3 | 10 min | Medium | M2 — SHA-pin `goreleaser/goreleaser-action@v6` in [.github/workflows/release.yml:34](../../../.github/workflows/release.yml) |
| 4 | 5 min | Low | L5 — SHA-pin `actions/checkout@v4`, `actions/setup-go@v5` in both workflow files |
| 5 | 5 min | Low | L4a — Add `maxLength` constraints to [docs/brag-entry.schema.json](../../../docs/brag-entry.schema.json) |
| 6 | 15 min | Low | L4b — Add length checks in [internal/cli/add_json.go:parseAddJSON](../../../internal/cli/add_json.go) matching schema (binary is the authoritative validator) |
| 7 | 10 min | Info | I1+I2 — Add `timeout-minutes: 30` and `concurrency:` block to [.github/workflows/release.yml](../../../.github/workflows/release.yml) |
| 8 | — | — | ~~Add a govulncheck step to ci.yml~~ → **deferred to backlog** ([projects/PROJ-001-mvp/backlog.md](../../../projects/PROJ-001-mvp/backlog.md) "govulncheck CI step"). Largely redundant with Dependabot security alerts (item 2 of GitHub Advanced Security section below) — both consult the same Go vulnerability database; Dependabot fires continuously. Govulncheck's incremental value (call-graph reachability filter) is marginal for a small dep graph. |
| 9 | 10 min | Medium (compensating) | M3 — Create the PAT with the fine-grained scope above; set 90-day expiration; calendar-remind for rotation |

Total: ~1 hour of work for a complete pre-tag-push hardening pass.

**Optional / deferred:**
- L2, L3 — editor signal-handling + `LimitReader`. Local-trust-bounded; not exploitable.
- I3 — code signing / SLSA. Defer past v0.1.0 unless you want the learning value.
- I4 — `framework-feedback/` etc. — `.gitignore` or commit to `docs/`.
- I5 — Pin specific goreleaser patch version.

---

## Out of scope / not assessed

- **Penetration testing of the install path** (`brew install` end-to-end on a fresh macOS box). Recommended as the v0.1.0-rc1 smoke test step (already in the spec's Ship Checklist).
- **Threat-model document.** Implicit threat model used: attacker controls JSON payload to `brag add --json` (including via Claude-Code hook) and FTS5 query input; does NOT control env vars, `$PATH`, local filesystem, or other local users (single-user-laptop assumption modulo M1's multi-user concern).
- **License compliance audit** of transitive deps (none flagged as obviously incompatible with MIT, but no formal SPDX scan run).
- **Cryptographic posture review.** No crypto in the binary; supply-chain hashing covered via goreleaser's `checksum.txt` SHA-256 output.

---

## GitHub Advanced Security features enabled (post-review hardening, 2026-04-26)

After the four review tracks landed, the following native GitHub features were enabled on `github.com/jysf/bragfile000` to provide ongoing automated coverage. Verified via `gh api repos/jysf/bragfile000` and `gh api repos/jysf/bragfile000/code-scanning/default-setup`.

| # | Feature | Status | Notes |
|---|---|---|---|
| 1 | **CodeQL (Code Scanning)** | ✅ enabled | `state: configured`, `languages: ["actions", "go"]`, `query_suite: default`, `threat_model: remote`, weekly schedule. Scans both Go source AND workflow YAMLs — partially mitigates M2/L5 by alerting on action-pinning patterns automatically. **Currently 0 alerts; first scan pending** (will populate within ~24h or on next push). |
| 2 | **Dependabot security alerts** | ✅ enabled | `dependabot_security_updates: enabled`; `vulnerability-alerts` endpoint returns 204 (active). Watches `go.mod` against the GitHub Advisory Database. Auto-PRs on advisory publication. Replaces the deferred govulncheck CI step (item 8 above). |
| 3 | **Dependabot version updates** | ✅ enabled | Configured via [.github/dependabot.yml](../../../.github/dependabot.yml). Watches both `gomod` (weekly Go-module bumps) and `github-actions` (weekly action-version bumps). The `github-actions` watcher is the maintenance story for M2/L5 — once actions are SHA-pinned, Dependabot opens an auto-PR on each upstream version. *(Initial commit `df6213b` shipped GitHub's empty placeholder template; corrected in a follow-up commit before any Dependabot run.)* |
| 4 | **Secret scanning + push protection** | ✅ enabled | `secret_scanning: enabled`, `secret_scanning_push_protection: enabled`. Scans real-time and historical commits for known secret patterns (PATs, AWS keys, etc.); blocks pushes containing detected secrets at git-push time. The historical scan in track 1C confirmed no past leaks; this adds forward protection. Two minor sub-options remain default-disabled: `secret_scanning_non_provider_patterns` (high-entropy generic strings) and `secret_scanning_validity_checks` (live-token-validation API calls). |

**Net effect on the prioritized fix list:**
- Item 8 (govulncheck CI step) deferred to backlog — Dependabot covers the advisory-detection use case continuously.
- Items 3/4 (action SHA-pinning) keep their priority — pinning is what catches malicious-tag-move attacks; Dependabot then handles ongoing maintenance.
- All other items unchanged.

---

## `/ultrareview` (track c, skipped)

After the four Claude-driven review tracks (Tier 1A/B/C, Tier 2 subagent, `/security-review` skill, `govulncheck`) returned no critical/high findings and only hardening-shape mediums, the user weighed the marginal value of a fifth track against the per-run cost and **chose to skip `/ultrareview`** (which is user-triggered and billed). Justification: the existing tracks plus the GitHub Advanced Security features above provide sufficient coverage for a personal project shipping for learning value. If a future review (PROJ-002 framing, post-incident, or pre-1.0 stability) warrants belt-and-suspenders, `/ultrareview <PR#>` remains the canonical path.

---

## Verdict

**Cleared for tag-push after items 1–6 of the prioritized fix list.**
Items 7–9 are recommended but not blocking. The codebase has no critical or high-severity findings; the discipline applied during construction (parameterized SQL, narrow FTS5 input handling, no `panic`, no shell-string composition, no network surface) shows up clearly in the review. The medium-severity findings are hardening / privacy posture rather than active exploitation paths.

The PAT creation step (item 9) gates the `git tag v0.1.0` step regardless — without `HOMEBREW_TAP_GITHUB_TOKEN` set in the repo's Actions secrets, the release workflow's tap-formula push will fail. Use the recommended fine-grained scope.

---

## Appendices

### A. govulncheck output

```
$ govulncheck ./...
No vulnerabilities found.
```

(Run 2026-04-26 against `main` at `1582572`. Tool: `golang.org/x/vuln/cmd/govulncheck@latest` installed during the review.)

### B. File inventory reviewed

**Go source (security-relevant):**
- `cmd/brag/main.go`
- `internal/cli/`: `add.go`, `add_json.go`, `delete.go`, `edit.go`, `export.go`, `root.go`, `search.go`
- `internal/config/config.go`
- `internal/editor/editor.go`, `internal/editor/launch.go`
- `internal/storage/`: `store.go`, `migrate.go`, `entry.go`, `migrations/0001_initial.sql`, `migrations/0002_add_fts.sql`

**Distribution pipeline:**
- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `.goreleaser.yaml`

**AI-integration surface (SPEC-022):**
- `scripts/claude-code-post-session.sh`
- `examples/brag-slash-command.md`
- `docs/brag-entry.schema.json`

**Repo hygiene:**
- `.gitignore`
- `LICENSE`
- `go.mod`, `go.sum`
- `framework-feedback/process-feedback.md`, `framework-feedback/scale-recommendations.md`, `revew1.md`, `status-after-nine-specs.md` (untracked; verified contents)

**External:**
- `github.com/jysf/homebrew-bragfile` (tap repo) — visibility, default branch, branch protection state

### C. Severity calibration used

- **Critical:** users get owned (RCE on install, secret theft, supply-chain compromise of users)
- **High:** data loss, privilege escalation on user machine, MITM on install
- **Medium:** exploit requires unusual circumstances; data integrity issue with non-trivial impact
- **Low:** best-practice violation, hardening opportunity
- **Info:** noted, not a vulnerability

Calibrated for a **personal project shipping for learning value, not enterprise distribution** per the stated PROJ-001 brief. Enterprise calibration would shift several lows to mediums.

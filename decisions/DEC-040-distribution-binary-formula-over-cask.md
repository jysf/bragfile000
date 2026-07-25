---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-040
  type: decision
  confidence: 0.82
  audience:
    - developer
    - agent
    - operator

agent:
  id: claude-opus-4-8
  session_id: null

project:
  id: PROJ-006
repo:
  id: bragfile

created_at: 2026-07-16
supersedes: null                     # the original cask choice was never a DEC (commit 1582572, a deprecation fix)
superseded_by: null

tags:
  - distribution
  - homebrew
  - macos
  - signing
  - packaging
  - no-cgo
---

# DEC-040: Distribute via a binary Homebrew formula, not a cask — to avoid code-signing entirely

## Decision

bragfile is distributed as a **binary Homebrew formula** (goreleaser installs the
pre-built release binary via a `brews:`-generated formula), **not** a cask —
because Homebrew does not quarantine formula-installed binaries, so no macOS
code-signing or notarization is required. The formula lives on the **shared
`jysf/homebrew-tap`** (install: `brew install jysf/tap/bragfile`), consolidating
off the per-project `homebrew-bragfile` tap so one tap and one trust cover all
jysf tools. `go install` remains a valid secondary path for Go developers.

## Context

The v0.1.0 release shipped a formula, then switched to a cask in commit
`1582572` — a verify-phase fix to silence a goreleaser `brews:` deprecation
warning. That switch was never a decision (a four-line YAML diff, no DEC) and it
silently changed the artifact's OS-trust path: **casks are quarantined by
Gatekeeper, formulae are not.** The entire signing/notarization problem — the
`xattr -dr com.apple.quarantine` README workaround and a deferred $99/yr Apple
Developer backlog item — traces to that one line. This DEC is the downsides pass
that decision never got (AGENTS.md §4, "Packaging-shape changes are decisions").

Two independent macOS gates must not be conflated:
- **Gatekeeper quarantine** — applies to casks (and browser downloads), not to
  formulae or locally-built binaries. This is the signing trigger.
- **Homebrew tap trust** (6.0+) — a one-time `brew trust` for any third-party
  tap, cask *or* formula. Not solvable by artifact type; out of scope here.

The goal: a `brew install` that needs no code-signing. That is a Gatekeeper
question, and it is a cask problem.

## Alternatives Considered

- **Cask + notarization**
  - What it is: sign with a Developer ID cert + notarize via Apple.
  - Why rejected: $99/yr + cert management + Apple approval latency + CI secrets
    — disproportionate remediation for "let a stranger run a CLI." Disproportion
    is a signal to re-examine the upstream decision, not pay the bill.

- **Cask + `xattr` workaround (status quo)**
  - What it is: each user runs `sudo xattr -dr com.apple.quarantine …` per install.
  - Why rejected: a scary manual command on every install/upgrade; the current pain.

- **`go install …@latest` as primary**
  - What it is: users build from source via the Go toolchain.
  - Why rejected as *primary*: requires Go, and the no-cgo `modernc.org/sqlite`
    driver (DEC-001) is a heavy compile (30s–minutes cold). Fine as a *secondary*
    developer path; wrong as the default.

- **Build-from-source Homebrew formula**
  - What it is: the formula runs `go build` on the user's machine.
  - Why rejected: same Go-toolchain + heavy-compile cost as `go install`, plus a
    hand-maintained formula — worst UX with no offsetting benefit.

- **`curl | sh` install script**
  - What it is: script downloads the pre-built binary (curl doesn't set quarantine).
  - Why rejected: dodges both gates and needs no compile, but the `curl|sh` trust
    smell and hand-rolled upgrades are worse than a formula for the same benefit.

- **Binary Homebrew formula (chosen)**
  - What it is: goreleaser `brews:` installs the pre-built, ldflags-stamped
    release binary via a formula.
  - Why selected: formulae are not quarantined → **no signing, no Gatekeeper
    prompt**; pre-built → **no compile**; it is the release binary → **`--version`
    works with no code change**; reuses goreleaser automation; best `brew install`
    UX. The only cost is a goreleaser deprecation warning — a *nudge, not a wall*.

## Consequences

- **Positive:** No code-signing/notarization on the critical path; no Gatekeeper
  prompt; no compile; `--version` stamps correctly from the release ldflags (no
  `ReadBuildInfo` change needed); the $99/yr backlog item and the README `xattr`
  workaround are retired.
- **Negative:** Re-introduces the goreleaser `brews:` deprecation warning. We
  **accept it deliberately** — a cosmetic warning is a better consequence than a
  real signing/UX cost, which is the exact reasoning the original switch inverted.
  If goreleaser removes `brews:`, migrate to a hand-maintained tap formula (still
  a binary formula, still unquarantined) — tracked as the fallback.
- **Neutral:** The Homebrew tap-trust gate (if it fires) is independent of
  artifact type. `go install` stays available for Go devs. The cask, and the
  per-project `homebrew-bragfile` tap, are retired.

## Validation

Right if: on a clean Mac, `brew install jysf/tap/bragfile` followed by
`brag --version` runs with **no Gatekeeper prompt** and prints the tagged
version. Revisit if: (a) goreleaser removes `brews:` → switch to a hand-maintained
tap formula; (b) a large non-Homebrew, non-Go audience emerges → reconsider a
signed+notarized cask for that channel specifically.

**Open question (verify at the clean-host cut):** does a *formula* from
`jysf/homebrew-tap` trigger Homebrew's untrusted-tap gate, and what is the exact
trust command? The gate and its `brew trust --cask …` command were researched
against a **cask** (AGENTS.md §4); a formula may not trigger it. Docs currently
tell the user to run whatever command Homebrew prints, rather than asserting one.

Confidence: 0.82. The mechanism is well-understood and established Homebrew
behavior (formulae are not quarantined; casks are). Residual uncertainty is the
longevity of goreleaser's `brews:` support and whether a non-Go/non-brew audience
ever appears — both have named fallbacks above, so neither blocks the decision.

## References

- Related specs: SPEC-071 (the migration), SPEC-023 (original distribution)
- Related decisions: DEC-001 (no-cgo driver — the reason the compile-based paths
  are costly)
- Code: `.goreleaser.yaml`; commit `1582572` (the undocumented cask switch)
- Docs: `docs/distribution-decisions.md`, `AGENTS.md` §4, `docs/macos-notarization-checklist.md`

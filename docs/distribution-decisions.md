# Distribution decisions — a checklist before you change how the app ships

**When to use this:** any change to *how bragfile reaches a user's machine* — the
goreleaser packaging blocks (`brews:`, `homebrew_casks:`, `nfpms:`, `archives:`,
`signs:`/`notarize:`), the tap/channel, or the install path. This includes
one-line deprecation fixes: a change that alters the artifact's **type or trust
path** is a decision, not hygiene (see AGENTS.md §4, "Packaging-shape changes are
decisions").

**Why this file exists:** the v0.1.0 formula→cask switch (commit `1582572`) was a
four-line deprecation fix that silently introduced the entire macOS
signing/Gatekeeper problem. It never ran a downsides pass because nothing in its
path forced one. This checklist is that forced pass, extracted from the release
pre-flight (`projects/_templates/spec-release-cut.md`, DEC-006) so *any*
distribution change can consult it — not just a release-cut spec.

---

## The four framing questions (answer all before merging)

1. **What artifact does the user actually receive?** A pre-built binary? A
   source build? A `.pkg`/`.app`? (bragfile: a single pure-Go binary.)
2. **Through what channel?** Homebrew tap (cask vs formula), `go install`,
   install script, direct download. Each channel has its own trust gates.
3. **Does a clean host trust it on first run?** The load-bearing question — see
   the trust matrix below. This is the one that was skipped.
4. **What does the audience already have?** A Go toolchain? Homebrew? This
   decides whether "build from source" or "`go install`" is friction-free or a
   blocker.

## Clean-host trust matrix (macOS)

Two **independent** gates — do not conflate them:

| Gate | Triggers on | Cleared by | NOT cleared by |
|---|---|---|---|
| **Gatekeeper quarantine** | a **cask** (or any browser/quarantine-aware download) of an unsigned/un-notarized binary | notarization ($99/yr) · `xattr -dr com.apple.quarantine` · shipping a **formula** or **locally-built** binary (not quarantined) · `go install` | switching taps; `brew trust` |
| **Homebrew tap trust** (6.0+) | installing a cask **or formula** from a third-party tap | `brew trust --cask <tap>/<name>` (one-time) · not using a Homebrew tap at all (`go install`, install script) | notarization; artifact type (formula does **not** dodge this — the gate is on the tap source) |

The takeaway that keeps getting relearned: **notarization ≠ tap-trust ≠
artifact-type.** Only leaving Homebrew entirely (`go install`, install script)
clears *both* gates without paying for signing.

## Before merge

- [ ] All four framing questions answered in the PR/spec.
- [ ] Clean-host trust checked for every OS the change targets (run the shipped
      artifact on a machine that didn't build it — or state honestly it's
      deferred, per the release pre-flight `[cut]` convention).
- [ ] If the change alters artifact type or trust path: **emit a DEC** (or a
      one-line downsides record) — this is a decision, not a lint fix.
- [ ] Disproportion check: if the remediation cost dwarfs the problem, re-examine
      the upstream decision before paying it.

## Related

- `AGENTS.md` §4 — the packaging-tripwire rule + the Gatekeeper / dual-tag /
  brew-trust lessons-earned addenda.
- `projects/_templates/spec-release-cut.md` — the full release pre-flight this
  checklist is extracted from.
- `docs/macos-notarization-checklist.md` — step-by-step *if* you choose to sign.

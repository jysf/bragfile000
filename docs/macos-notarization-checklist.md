# macOS code signing + notarization — implementer checklist

**Status:** deferred. See [backlog entry](../projects/PROJ-001-mvp/backlog.md#macos-code-signing--notarization-apple-developer-id) for "why deferred" + trigger conditions for revisiting.

**Why this file exists:** the backlog entry captures the *decision* to defer; this file captures the *step-by-step actions* a future implementer (you, or a fresh agent session) walks through to actually do the work. Pick up cold from here.

**Outcome when this is complete:** `brew install jysf/bragfile/bragfile` on a fresh Mac → `brag --version` runs without Gatekeeper prompting `"Apple could not verify…"`. The xattr workaround in `README.md` §Install becomes obsolete.

**Effort estimate:** ~3-5 hours of focused work spread across "wait for Apple approval" + "edit configs" + "test on a clean Mac." Plus $99/year ongoing for Apple Developer Program.

**Verify against current goreleaser docs at implementation time:** goreleaser's signing/notarization config schema evolves. The literal blocks below match goreleaser v2.15.x circa 2026-05; check <https://goreleaser.com/customization/notarize/> and <https://goreleaser.com/customization/sign/> for any drift before transcribing.

---

## Phase A — one-time prereqs (do these first; Apple approval is the long pole)

- [ ] **A1. Enroll in Apple Developer Program.** <https://developer.apple.com/programs/enroll/>. $99/year. Requires an Apple ID. Approval is typically 1-2 business days. You can proceed with A2 after enrollment confirms.

- [ ] **A2. Generate a "Developer ID Application" certificate.**
  - <https://developer.apple.com/account/resources/certificates/list>
  - Click "+" → **Developer ID Application** (NOT "Mac Development" or "Mac App Distribution"). This is the cert type for binaries distributed outside the Mac App Store.
  - Generate a Certificate Signing Request (CSR) from Keychain Access → Certificate Assistant → Request a Certificate from a Certificate Authority. Save the `.certSigningRequest` file.
  - Upload the CSR; Apple issues the certificate (`.cer`). Download it.
  - Double-click the `.cer` to install into Keychain Access (default "login" keychain is fine).

- [ ] **A3. Export the certificate as a `.p12` for CI use.**
  - In Keychain Access, find the new "Developer ID Application: <Your Name> (TEAMID)" certificate.
  - Right-click → Export → save as `developer-id.p12`. **Set a strong password** when prompted; you'll need it in A5.
  - Keep the `.p12` and password somewhere safe (1Password, etc.) — not in git.

- [ ] **A4. Generate an app-specific password for notarytool.**
  - <https://appleid.apple.com/account/manage> → Sign-In and Security → App-Specific Passwords.
  - Label: `bragfile-notarytool` (any descriptive name).
  - Copy the generated password (4-group hyphenated form). You'll see it once.

- [ ] **A5. Note your Apple Team ID.**
  - <https://developer.apple.com/account/>. Membership Details → Team ID (10-character alphanumeric, e.g. `ABCDE12345`).

---

## Phase B — GitHub repo secrets (5 new secrets on bragfile000)

All five live at <https://github.com/jysf/bragfile000/settings/secrets/actions>. Add exactly these names:

- [ ] **B1. `MACOS_CERTIFICATE`** — base64-encoded `.p12` contents. Generate via:
  ```bash
  base64 -i developer-id.p12 | pbcopy
  ```
  Paste from clipboard into the secret value field.

- [ ] **B2. `MACOS_CERTIFICATE_PASSWORD`** — the password you set when exporting the `.p12` in A3.

- [ ] **B3. `APPLE_ID`** — the Apple ID email associated with your Developer account (e.g. `jyashinsky@gmail.com`).

- [ ] **B4. `APPLE_PASSWORD`** — the app-specific password from A4 (NOT your Apple ID login password).

- [ ] **B5. `APPLE_TEAM_ID`** — the 10-char Team ID from A5.

Verify all five present:
```bash
gh api repos/jysf/bragfile000/actions/secrets --jq '.secrets[] | .name'
# expect: APPLE_ID, APPLE_PASSWORD, APPLE_TEAM_ID, HOMEBREW_TAP_GITHUB_TOKEN, MACOS_CERTIFICATE, MACOS_CERTIFICATE_PASSWORD
```

---

## Phase C — `.goreleaser.yaml` updates

- [ ] **C1. Add `signs:` block for codesign.** Insert after the existing `archives:` block, before `checksum:`:
  ```yaml
  signs:
    - id: macos-codesign
      cmd: codesign
      args:
        - "--options=runtime"
        - "--sign={{ .Env.APPLE_DEVELOPER_ID }}"
        - "{{ .Path }}"
      artifacts: binary
      ids: [brag]
      stdin: '{{ .Env.MACOS_CERTIFICATE_PASSWORD }}'
  ```
  Note: this signs the *binary* per macOS conventions; the archives are then re-built around the signed binary. Some setups sign the archive itself — verify which goreleaser expects in the current docs.

- [ ] **C2. Add `notarize:` block.** Insert near the top-level, alongside `release:`:
  ```yaml
  notarize:
    macos:
      - enabled: true
        sign:
          certificate: "{{ .Env.MACOS_CERTIFICATE }}"
          password: "{{ .Env.MACOS_CERTIFICATE_PASSWORD }}"
        notarize:
          issuer_id: "{{ .Env.APPLE_TEAM_ID }}"
          key_id: "{{ .Env.APPLE_ID }}"
          key: "{{ .Env.APPLE_PASSWORD }}"
          wait: true
  ```
  Wait for Apple's notarization API to complete before publishing (typically 2-15 minutes per archive). **Verify the exact schema** — goreleaser's notarize block has evolved; some versions use `notarytool:` instead of `notarize:`.

- [ ] **C3. Optionally add stapling.** After notarization, `xcrun stapler staple` attaches the notarization ticket to the artifact so Gatekeeper checks work offline. Goreleaser may do this automatically depending on version; verify.

---

## Phase D — `.github/workflows/release.yml` updates

- [ ] **D1. Switch the goreleaser job to `macos-latest` runner.** Signing requires Apple's `codesign` + `xcrun notarytool`, which only exist on macOS GitHub Actions runners. Current workflow runs on `ubuntu-latest`.

  Options:
  - **Simpler:** flip the whole `goreleaser` job to `runs-on: macos-latest`. Linux cross-compiles still work from macOS via Go's toolchain.
  - **Cleaner but more YAML:** matrix-split — one job builds linux on ubuntu, another signs+notarizes darwin on macos. Goreleaser supports a `--split` mode that's intended for this. Defer matrix-split unless macos-latest minute costs become noticeable.

- [ ] **D2. Expose the new secrets via `env:` on the goreleaser step.** Add to the existing `env:` block (alongside `GITHUB_TOKEN` and `HOMEBREW_TAP_GITHUB_TOKEN`):
  ```yaml
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
    MACOS_CERTIFICATE: ${{ secrets.MACOS_CERTIFICATE }}
    MACOS_CERTIFICATE_PASSWORD: ${{ secrets.MACOS_CERTIFICATE_PASSWORD }}
    APPLE_ID: ${{ secrets.APPLE_ID }}
    APPLE_PASSWORD: ${{ secrets.APPLE_PASSWORD }}
    APPLE_TEAM_ID: ${{ secrets.APPLE_TEAM_ID }}
    APPLE_DEVELOPER_ID: "Developer ID Application: <Your Name> (${{ secrets.APPLE_TEAM_ID }})"
  ```
  The `APPLE_DEVELOPER_ID` is the human-readable cert name codesign expects — replace `<Your Name>` with the name on your Developer Program account.

- [ ] **D3. Bump `timeout-minutes` if notarization is slow.** Current is 30 minutes; notarization can add 5-15 minutes per archive × 4 archives = up to an hour. Bump to `timeout-minutes: 90`.

---

## Phase E — test cycle (RC before production)

- [ ] **E1. Cut a pre-release tag.** `git tag v0.2.0-rc1 && git push origin v0.2.0-rc1`. Watch release workflow.

- [ ] **E2. Verify the signing step ran cleanly** in the GitHub Actions log. Look for `codesign --verify` exit 0 and `xcrun notarytool submit … --wait` returning `status: Accepted`.

- [ ] **E3. Download the darwin_arm64 archive from the GitHub Release.** Extract on a fresh Mac (or a Mac that's never run unsigned bragfile builds — clear `~/Library/Saved Application State/com.bragfile.*` if it exists, and `xattr -dr com.apple.quarantine` should NOT be needed on the new binary). Run:
  ```bash
  ./brag --version
  ```
  Expected: prints version. No Gatekeeper prompt. **This is the success criterion.**

- [ ] **E4. Verify spctl acceptance.**
  ```bash
  spctl --assess --type execute /path/to/brag
  # expect: "accepted" + signing identity reported
  ```

- [ ] **E5. If RC clean, delete RC tag/release per the AGENTS.md §4 dual-tag rule, then cut v0.2.0 proper.** (Note: don't reuse v0.1.x — v0.1.0 already shipped unsigned. Bump to v0.2.0 to signal the artifact format change.)

---

## Phase F — documentation cleanup post-success

- [ ] **F1. Remove the "macOS Gatekeeper note" section from `README.md` §Install.** It's obsolete once notarization ships.

- [ ] **F2. Update `AGENTS.md` §4** — change the "macOS Gatekeeper on unsigned binaries" lesson note from active guidance to historical: prefix with `**Historical (resolved in vX.Y.Z):**` and keep for traceability.

- [ ] **F3. Move the backlog entry to the "Removed / delivered" section** at the bottom of `projects/PROJ-001-mvp/backlog.md` with the SPEC ID that delivered it (if you wrap this work in a spec) or the bare commit SHA otherwise.

- [ ] **F4. Auto-brag the ship** — separate brag for "shipped Apple-signed + notarized bragfile binaries" with proper `--type "shipped"` and a description of the effort.

---

## References

- [backlog entry — macOS code signing + notarization](../projects/PROJ-001-mvp/backlog.md#macos-code-signing--notarization-apple-developer-id) — the "why deferred" record
- [AGENTS.md §4 — macOS Gatekeeper on unsigned binaries](../AGENTS.md) — the current workaround documentation
- [README.md §Install — macOS Gatekeeper note](../README.md#install) — the user-facing workaround
- Goreleaser docs (verify at implementation time):
  - <https://goreleaser.com/customization/sign/>
  - <https://goreleaser.com/customization/notarize/>
- Apple's notarization docs:
  - <https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution>
  - `xcrun notarytool --help` on macOS

---

## Implementation note for whoever picks this up

This checklist was drafted 2026-05-11 immediately after the v0.1.0 ship encountered the Gatekeeper prompt. The friction it describes is real but bounded; the work to remove it is real but bounded. If you're picking this up cold, read the backlog entry first to confirm the "why now" trigger has actually fired — don't do this work just because the checklist exists. The xattr workaround stays viable indefinitely; the only reason to ship notarization is when its absence is causing material friction for actual users.

# Prompt — retire `jysf/homebrew-bragfile` properly (tap migration, then archive)

Run this in a fresh session. It operates on the **`jysf/homebrew-bragfile`
GitHub repo**, not on `bragfile000`. Nothing here touches the bragfile000
working tree.

## Context you need

`jysf/homebrew-bragfile` is the retired per-project Homebrew tap. It was
replaced at v0.5.2 by a binary **formula** on the shared `jysf/homebrew-tap`
(DEC-040 / SPEC-071). The old tap still contains `Casks/bragfile.rb` pinned at
version **0.5.1**.

That pin is the bug. Homebrew treats it as current, so anyone who still has the
old tap gets:

- no upgrade signal — `brew outdated` is completely silent
- no migration hint — nothing tells them the package moved
- a permanently frozen 0.5.1 install, missing everything from v0.5.2 onward

This was found at the v0.6.0 cut and is written up as **finding F4** in
`projects/PROJ-006-agent-native-depth-core/specs/SPEC-076-v0-6-0-release-cut.md`.
v0.6.0 documents the manual migration in README §Install and the `[0.6.0]`
CHANGELOG entry; this prompt fixes the mechanism instead of documenting around
it.

## What to do

**Do NOT use a `deprecate!` stanza.** That is for sunsetting a package, and it
only surfaces on `brew install`/`brew info` — which a frozen user never runs, so
it would not solve this. This is a *move*, and Homebrew's mechanism for a move
is `tap_migrations.json`.

Two facts verified against Homebrew 6.0.15's own source, worth re-confirming if
your Homebrew is newer:

- `Library/Homebrew/cask/cask_loader.rb` consults `tap.tap_migrations[token]`
  and redirects to the new tap — cask migrations are supported, not just
  formulae.
- `Library/Homebrew/cask/audit.rb` raises `"<token> is listed in
  tap_migrations.json"` when a cask file *and* a migration entry both exist.
  So the cask file must be **removed**, not left in place beside the entry.

### Steps

1. Clone `jysf/homebrew-bragfile` fresh (it is no longer tapped locally).
2. `git rm Casks/bragfile.rb`
3. Create `tap_migrations.json` at the **repo root**:
   ```json
   {
     "bragfile": "jysf/tap"
   }
   ```
   The value is the destination **tap**, not `jysf/tap/bragfile`. A value with
   no slash would mean "renamed inside this same tap", which is not the case
   here.
4. Update `README.md` to say the tap is retired and point at
   `brew install jysf/tap/bragfile`.
5. Commit and push to `main`.

### Verify before archiving

On a machine with the old tap added back temporarily:

```bash
brew tap jysf/bragfile
brew update
brew info bragfile
```

Expect Homebrew to resolve `bragfile` to `jysf/tap/bragfile` rather than the
old 0.5.1 cask. If it still reports 0.5.1, the migration is not taking effect —
stop and diagnose before archiving, because archiving makes the repo read-only
and you lose the ability to fix it.

Then `brew untap jysf/bragfile` again.

### Then archive

**Order matters: push and verify the migration BEFORE archiving.** An archived
repo is read-only.

Archive (do not delete) `jysf/homebrew-bragfile`:

- Deleting breaks `brew update` for anyone still tapped, with a recurring
  cryptic fetch error that tells them nothing actionable. Archiving keeps the
  tap resolvable and quiet.
- Archiving makes the repo read-only, which closes a live vector: `main` is
  currently **unprotected** and the repo is public and writable. The
  2026-04-26 pre-distribution security review names this exact scenario —
  *"attacker pushes a malicious cask file to `github.com/jysf/homebrew-bragfile`;
  subsequent users get a backdoored binary."*
- DEC-040, that security review, and two CHANGELOG entries all link to it.
  Archiving keeps those links alive; deletion breaks them after a short restore
  window.

## Honest scoping

Measured exposure is close to zero: 0 stars, 0 forks, 0 watchers, **0 views**,
and 5 clones from 5 unique actors over 14 days — clones with no views is the
signature of `brew update` fetches and scrapers, not people. The local machine
has already been untapped manually.

So this is **defensive hygiene, not an incident.** The archive is the part with
real value (it closes the unprotected-write vector). The migration entry is
worth doing because it is cheap and it is the correct mechanism — but if you
only do one thing, archive.

Do not hold the v0.6.0 tag for this.

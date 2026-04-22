# Guide for AI agents: helping the user capture brag-worthy work

*Share this file (or its URL) with any Claude/AI session working on
the user's projects. It's self-contained and can be dropped into a
session context at any time.*

---

## What is `brag`?

`brag` is a local-first CLI tool the user built for tracking
career-worthy work moments. It stores entries in a SQLite database
at `~/.bragfile/db.sqlite` on the user's machine. The user
accumulates these over time and uses them for retrospectives,
quarterly reviews, promotion packets, and resume updates.

The user is now using `brag` across multiple projects, and wants AI
sessions working on those projects to **propose brag entries at
meaningful moments**. This guide tells you how.

---

## Your role, in one sentence

When significant work ships in a session you're part of, **propose a
brag entry — don't post unilaterally — and only execute `brag add`
after the user explicitly approves.**

The user's trust in the tool depends on not getting spammed with
low-value entries. Your job is to recognize the significant moments
and help capture them well.

---

## When to propose a brag

### ✅ YES — propose a brag when the session produced:

- **A shipped feature** — real user-facing change, merged and
  deployed (or equivalent — for personal projects, working on the
  user's main branch counts).
- **A fixed significant bug** — not a typo, a real debug-and-repair
  effort with a concrete outcome.
- **An architectural decision** — especially one captured as an ADR
  or DEC that will bind future work.
- **A meaningful piece of documentation** — a migration guide, an
  onboarding doc, a postmortem, a process proposal.
- **Mentoring / unblocking others** — concrete help delivered to a
  named teammate or team, with a specific outcome.
- **A measurable learning applied** — the user learned something
  concrete (a new framework, a pattern, a deep-dive) AND applied it
  to produce something specific.
- **A delivered artifact** — PR merged, presentation given, proposal
  adopted, playbook published.

### ❌ NO — skip the brag when the work is:

- **Routine maintenance** — typo fixes, trivial refactors, updating
  copyright years, bumping a patch-level dep.
- **Not the user's contribution** — if the user pointed at something
  and you did it with minimal input, that's probably your work, not
  theirs. Err toward skipping.
- **Small and incremental** — every commit isn't brag-worthy. One
  good brag per shipped feature, not per commit.
- **Time spent without work achieved** — "spent 3 hours debugging"
  is not a brag. "Root-caused the OAuth latency regression to a
  missed index" is.
- **Vague in scope** — if you can't articulate a specific outcome in
  one sentence, it's not a brag yet. Wait until you can.

### Rule of thumb

Would the user mention this in a weekly status update to their
manager? If yes, probably brag-worthy. If they'd say "just cleaned
up some stuff," definitely not.

---

## The approval loop (non-negotiable)

1. **You identify** a candidate brag moment during or after a piece
   of work.
2. **You draft** the full entry with all fields (format below) and
   present it to the user.
3. **You ask explicitly:** "Want me to record this as a brag?"
4. **The user responds:** approve, edit, or skip.
5. **Only then** do you run `brag add`.

**Never post without approval.** If you're unsure whether a moment
qualifies, describe it and ask. "Is this worth capturing as a brag?"
is always better than posting noise.

---

## How to compose a good brag

### Fields

| Flag | Required? | Purpose |
|---|---|---|
| `-t` / `--title` | **Required** | Short, specific, action-verb. The headline. |
| `-p` / `--project` | Strongly recommended | Work context: repo name, client, team, initiative. Enables future `brag list --project X` filtering. |
| `-k` / `--type` | Recommended | Category: `shipped`, `fixed`, `learned`, `documented`, `mentored`, `unblocked`, `proposed`, `reviewed`, etc. Free-form — pick whatever's useful. |
| `-T` / `--tags` | Recommended | 2–4 comma-joined topic tags for future filtering. Example: `auth,perf,backend`. |
| `-i` / `--impact` | **Important** | The concrete outcome: metric, quote, unlock, business result. **This is the most load-bearing field for reviews.** |
| `-d` / `--description` | Optional | Free-form narrative. What you did, why it mattered, relevant context. Use single quotes to embed `"`. |

### Field quality bar

- **Title**: specific enough that the user could re-read it in 6
  months and know what it meant. "Shipped auth refactor cutting
  login latency 80%" beats "Worked on auth."
- **Impact**: a metric or a named outcome. "Unblocked mobile v3
  release" beats "improved performance." "Reduced p99 from 600ms to
  120ms" beats "made it faster."
- **Description**: the story. Tell why it mattered, how you
  approached it, what specifically changed. Think 2–5 sentences.

---

## The command

```bash
brag add \
  -t "Title here" \
  -p "project-name" \
  -k "shipped" \
  -T "tag1,tag2,tag3" \
  -i "Concrete impact statement with a metric or named outcome." \
  -d 'Free-form description. Use single quotes to safely embed "double quotes".'
```

The binary is typically at `~/go/bin/brag`. Find it via:

```bash
which brag
# or, if PATH isn't set up:
ls ~/go/bin/brag
```

On successful add, `brag` prints the new entry's numeric ID to
stdout. Verify:

```bash
brag show <id>
```

---

## Three good examples

### Example 1 — shipped

```bash
brag add \
  -t "Shipped OAuth refactor reducing login latency p99 from 600ms to 120ms" \
  -p "platform-backend" \
  -k "shipped" \
  -T "auth,perf,backend" \
  -i "Unblocked mobile v3 release (hard dependency). Reduced aggregate login infra load by 40%." \
  -d 'Replaced per-request JOIN between users and sessions with a Redis-backed session cache. Changed query pattern in internal/auth/session.go and rolled out behind feature flag over a week. Full postmortem in doc-042.'
```

### Example 2 — documented

```bash
brag add \
  -t "Authored migration guide that shipped 3 teams from legacy SOAP to internal gRPC service" \
  -p "platform-infra" \
  -k "documented" \
  -T "documentation,grpc,migration" \
  -i "Three teams (Fulfillment, Returns, Inventory) migrated within the quarter, eliminating the last SOAP services in the fleet." \
  -d 'Wrote comprehensive migration guide with per-language examples (Java, Python, Go) and a runbook for cutover. Ran 2 office-hour sessions answering team-specific questions. Published in internal wiki at /docs/soap-to-grpc.'
```

### Example 3 — fixed

```bash
brag add \
  -t "Fixed 3-year-old flaky test that blocked nightly deploy pipeline" \
  -p "observability" \
  -k "fixed" \
  -T "testing,ci,reliability" \
  -i "Restored nightly deploy pipeline to 100% pass rate from ~75%. Unblocked release cadence for 8 teams depending on the artifact." \
  -d 'Root-caused race condition in shared fixture initialization via targeted instrumentation. Flakiness appeared only under 4+ parallel test workers due to a Go channel buffer size of 1. Fix in internal/testutil/fixtures.go with an added sync.WaitGroup barrier.'
```

---

## Anti-examples (avoid these)

```bash
# ❌ Vague title
brag add -t "Did some coding today"

# ❌ No context, no impact
brag add -t "Fixed bug"

# ❌ Generic impact
brag add -t "..." -i "Improved things"

# ❌ Non-specific description
brag add -t "..." -d "worked on stuff"
```

The pattern in all of these: **no specific outcome.** A brag entry
without a specific outcome is a reminder, not an artifact.

---

## Reading entries back

```bash
brag list                                   # all entries, newest first
brag list --project platform-backend        # filter by project
brag list --tag auth --since 30d            # filter by tag + time
brag list --type shipped --limit 10         # most recent 10 shipped
brag show <id>                              # full entry as markdown
```

The user can export a filtered slice to paste into review documents:

```bash
brag list --project X --since 90d | while read -r line; do
  id=$(echo "$line" | cut -f1)
  brag show "$id"
  echo "---"
done > /tmp/review-draft.md
```

---

## If anything goes wrong

- `brag --version` fails: the binary may not be installed on PATH.
  Tell the user; don't try to install it from a non-canonical path.
- `brag add` returns exit 1: user error (missing title, invalid
  flag). Show the error, let the user decide.
- `brag add` returns exit 2: internal error (DB issue, disk full).
  Report to the user; don't silently retry.

---

## Short version (for session contexts)

> "After significant work ships, draft a `brag add` command with
> title, project, type, tags, impact, description. Present it to
> me, wait for approval, then execute. Don't post without approval.
> Focus the impact field on a specific metric or named outcome."

---

## Source

This guide lives in the bragfile project repo:
`docs/agent-brag-guide.md`. For the canonical up-to-date version,
see `github.com/jysf/bragfile000/blob/main/docs/agent-brag-guide.md`.

---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-023
  type: decision
  confidence: 0.8                     # honest: the gate/crossing mechanics are
                                       # high-confidence; the residual soft spots are
                                       # the copy wording and the precedence order
                                       # among simultaneously-crossed thresholds
                                       # (low-stakes, rare) — see Validation.
  audience:
    - developer
    - agent

agent:
  id: claude-opus-4-8
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-003
repo:
  id: bragfile

created_at: 2026-07-03
supersedes: null
superseded_by: null

tags:
  - cli
  - add
  - milestone
  - delight
  - stdout-stderr-spine
  - tty
---

# DEC-023: milestone notifications — TTY/stderr gate, crossing semantics, copy

## Decision

`brag add` prints **exactly one** celebratory line to **stderr, only when
stderr is a terminal**, on crossing a milestone threshold — lifetime total
`{10, 25, 50, 100, 250, 500, 1000}`, current-streak `{7, 30, 100}`, or
per-project `{10, 50}` (non-empty project only) — else a quiet
"first brag this week / today" nudge. "Crossing" is `before < T <= after`
(not equality), so a same-day re-add never re-fires. Precedence, when
several cross at once, is **total → streak → per-project → first-this-week →
first-today**. The line is **silent** under `--json`, on any non-TTY/piped
invocation, and when nothing crosses; it never touches stdout (the ID). TTY
detection is stdlib `os.Stderr.Stat().Mode()&os.ModeCharDevice` — **no new
dependency**. The streak value comes from the SPEC-038/DEC-022 corrected
`aggregate.Streak` (local day, alive through yesterday); the milestone clock
is a single local `time.Now()` so the zone rides on `now.Location()`. "This
week" is the ISO-8601 week (`time.Time.ISOWeek`, Monday-start) in that same
local zone. No persistent "already celebrated" state (migration-free core).

## Context

STAGE-009 (PROJ-003 v0.3.0 core) makes capture *delightful*: milestone
notifications were the 2026-06-16 brainstorm's best effort-to-payoff pick
because they fire on an action the user (or an agent) already takes. SPEC-039
implements them; this DEC settles the decidable design choices so the spec's
literal artifact (the copy) and its gating tests have a durable home.

The load-bearing constraint is `stdout-is-for-data-stderr-is-for-humans`
(blocking): `brag add`'s stdout is the entry ID (`id=$(brag add ...)`), so a
milestone — pure human chatter — must live on stderr and must be *silent* on
the machine-facing paths (`--json`, pipes, non-TTY) so scripted and
agent-driven capture stays byte-clean. STAGE-009's Design Notes frame this
line as the CLI-side mirror of the "stdout/stderr spine at a new transport"
that SPEC-040's MCP server faces.

The streak tier is the reason **SPEC-038 blocked SPEC-039**: before DEC-022,
current-streak read 0 for the whole part of a day before the first re-log, so
a streak milestone would have fired on a wrong number. With SPEC-038 shipped
(PR #57), the corrected metric is on `main` and this DEC consumes it.

## Alternatives Considered

- **Option A: fire on equality (`post == T`), not crossing.**
  - What it is: emit whenever the post-add metric equals a threshold.
  - Why rejected: double-fires the streak milestone on *every* same-day
    re-add once the streak sits at a threshold value (add twice on your 7th
    day → two "7-day streak!" lines). `before < T <= after` fires once, on
    the transition. (Total/per-project increment by exactly 1 per add, so for
    them equality and crossing coincide; the crossing rule is uniform and
    correct for all three.)

- **Option B: gate on `cmd.ErrOrStderr()` being a terminal.**
  - What it is: probe the command's configured error writer.
  - Why rejected: in tests that writer is a `bytes.Buffer`; in prod it is
    `os.Stderr`. Probing it couples the gate to the test harness. Gating on
    the real `os.Stderr` fd through an injectable seam
    (`addStderrIsTTY`/`defaultStderrIsTTY`) is correct in prod *and*
    deterministic in tests (which force the seam and pin it off by default in
    the shared harness).

- **Option C: promote `mattn/go-isatty` (already an indirect dep) to a
  direct dependency for TTY detection.**
  - What it is: `isatty.IsTerminal(os.Stderr.Fd())`.
  - Why rejected: it would fire `no-new-top-level-deps-without-decision` and
    add a direct dependency for a cosmetic feature. The stdlib
    `os.ModeCharDevice` probe answers "is stderr a terminal" adequately,
    stays pure-Go (`no-cgo`), and needs no dep DEC. (go-isatty handles a few
    exotic Windows console cases the stdlib check doesn't, but bragfile's
    milestone is a best-effort delight on a local dev CLI — the stdlib check
    is the right cost/benefit.)

- **Option D: persist a "celebrated this threshold" flag so a milestone
  never repeats.**
  - What it is: a table/column recording which thresholds a corpus has
    crossed.
  - Why rejected: needs a schema migration, and the STAGE-009 core is
    explicitly migration-free. A re-fire is only reachable by deleting the
    corpus back below a threshold and re-crossing it — rare, and it genuinely
    *is* a re-crossing. Deferred to "later, if earned."

- **Option E (chosen): TTY/stderr-gated, crossing-based, stdlib-detected,
  fixed-copy, migration-free.**
  - What it is: the Decision above.
  - Why selected: honors the stdout/stderr spine at zero risk to the ID
    contract; keeps every scripted/agent path byte-clean; adds no dependency
    and no migration; and localizes the whole threshold/precedence/copy
    matrix into a pure, exhaustively-unit-testable function with a thin
    Store/clock/TTY glue layer.

## Consequences

- **Positive:** capture gains a passive, no-new-command delight that rides an
  action already taken; `--json`/piped/non-TTY capture is provably unaffected
  (the `errBuf.Len()==0`-under-`--json` test); the ID contract
  (`id=$(brag add ...)`) is untouched because stderr, not stdout, carries the
  line; the pure decision function makes the copy/threshold/precedence matrix
  cheap to test and to tune later.
- **Negative:** the copy strings and the precedence order among
  simultaneously-crossed thresholds are judgment calls baked into a literal;
  changing them is a code+test edit (cheap, but not config). A milestone can
  re-fire after a corpus is deleted below a threshold and re-crosses it
  (accepted). Emoji may render as tofu on a threadbare terminal (cosmetic;
  a plain-mode toggle is out of scope).
- **Neutral:** the milestone does a second `Store.List` after the insert
  (one extra read on an interactive add — negligible). The MCP `brag_add`
  tool (SPEC-040) deliberately does **not** emit a milestone; its stdout is
  the protocol stream, a different spine handled at that transport.

## Validation

Right if: the split-buffer tests hold (`--json` and non-TTY leave `errBuf`
empty even with the TTY seam forced on; a TTY add at a threshold puts one
line on `errBuf` and the ID alone on stdout); the pure-function tests lock
each threshold set, the crossing-not-equality rule, the precedence order, and
the exact copy; and no go.mod change appears. Revisit if: (a) users ask to
silence or reshape the milestone (add an env var / `--quiet` / plain-mode —
the TTY gate already covers scripts/agents, so this is a human-preference
ask, not a correctness one); (b) the precedence order feels wrong in real
dogfooding (low-stakes; a one-line reorder + test edit); (c) a
multi-host/synced corpus makes "this week / today" ambiguous (out of scope —
the CLI is single-host, local zone via `now.Location()`); or (d) re-fires
after corpus deletion become annoying enough to justify the deferred
"celebrated-once" state (Option D — carries a migration).

## References

- Related specs: SPEC-039 (emits + implements this DEC), SPEC-038 (the
  corrected streak this consumes; the blocker), SPEC-020 (streak/aggregate
  origin + single-`time.Now()`-source pattern), SPEC-017 (`brag add --json`
  machine-path contract this keeps clean)
- Related decisions: DEC-022 (local-day alive-through-yesterday streak),
  DEC-011 (stored UTC RFC3339 entry shape), DEC-012 (`brag add --json`
  schema)
- Related constraints: `stdout-is-for-data-stderr-is-for-humans` (blocking —
  the spine this DEC is built around), `no-new-top-level-deps-without-decision`
  (warning — recorded as *not fired*: stdlib TTY probe), `no-cgo`,
  `test-before-implementation`
- Discussions: backlog.md "milestone notifications" (2026-06-16 brainstorm);
  STAGE-009 Spec Backlog SPEC-039 entry + the "stdout/stderr spine at a new
  transport" Design Note

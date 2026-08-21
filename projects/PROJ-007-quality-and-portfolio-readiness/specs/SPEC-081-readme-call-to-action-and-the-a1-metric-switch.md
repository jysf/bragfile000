---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-081
  type: chore                      # epic | story | task | bug | chore
  cycle: verify
  blocked: false
  priority: medium
  complexity: S                    # S | M | L  (L means split it)

project:
  id: PROJ-007
  stage: STAGE-021
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-08-19

references:
  decisions:
    # The promoted call-to-action makes two factual claims — "five typed
    # tools" and "an auto-loadable memory resource". Both were checked
    # against these records and against the code at design (see §12(b)).
    - DEC-024                      # the MCP server: SDK, stdio transport, provenance — the original "four typed tools" line
    - DEC-045                      # the push surface: three `brag://` resources + `brag_memory` as the FIFTH tool; amends DEC-024
  constraints:
    - test-before-implementation   # the A1 switch and A11 are written here, made to pass at build
    - one-spec-per-pr              # relevant because this spec deliberately declines to fold in V2
  related_specs:
    - SPEC-021                     # the README rewrite that created test-docs and set A1's band
    - SPEC-077                     # added W3, which pins the `> **Status:** v<latest> shipped` line
    - SPEC-079                     # added `## How this repo is built`; re-pinned A1 to 260 (LD5)
    - SPEC-080                     # the stage's other legibility spec; recorded V2, which this spec declines
---

# SPEC-081: the README call-to-action, and A1 measured in words

> **Cycle: design.** Framed 2026-08-19 (GO at S). This document settles the
> four forks framing left open, each with rejected alternatives per AGENTS.md
> §12, and carries the literal artifacts build transcribes.

## Context

STAGE-021's third and final spec, and the one that closes the stage.

Two changes that touch the same two files. The second is what makes the first
affordable — but, critically, **not what justifies it.** The frame set the
no-go condition itself: if the `wc -l` → `wc -w` switch is argued as *"raise
the ceiling so the call-to-action fits"*, this spec becomes the thing that
disarmed a working guard, which is the bump-until-green move SPEC-079's LD5
refused. So Change 2 is argued first, on its own evidence, and the argument is
**demonstrated rather than asserted** (see the next section).

Parent: `STAGE-021-make-the-discipline-legible`, spec 3 of 3.
Project: `PROJ-007`.

## Goal

An agent-using developer landing on the README sees the one command that makes
`brag` useful to their agent in a container they actually read, and the guard
protecting the README's size measures content rather than typography.

---

## Re-measurement

Framing's headline numbers **held**. Two of its supporting claims did not.

| Framing said | Measured 2026-08-20 | Verdict |
|---|---|---|
| README is 260 lines / 1,268 words | `wc -l` = 260, `wc -w` = 1268 | **HOLDS** |
| A1's ceiling is 260, zero headroom | `scripts/test-docs.sh:123` | **HOLDS** |
| The full MCP section is at line 219 | `## Using brag from an AI agent (MCP)` is README.md:219 | **HOLDS** |
| `assert_line_count_band` has six callers | A1:123, C2:271, D2:292, J2:588, T2:1027, X7:1410 — six call sites plus the definition at :45 | **HOLDS** |
| `W3` greps the literal `^> \*\*Status:\*\* v<latest> shipped`, deriving `<latest>` from the CHANGELOG | `scripts/test-docs.sh:1280`, exactly as described | **HOLDS** |
| "the practices page … reference[s] `A1` by name" | `grep -n 'A1' docs/engineering-practices.md` → **zero hits**. The page names `X3`, `E2`, `W3` and the W-series, never A1. | **WRONG** |
| "several commit messages reference `A1` by name" | True — 14 distinct commit-message lines on `main` name A1 | **HOLDS** |

The one wrong claim does not change the fork-2 answer (A1 keeps its id) but it
does change the *reason*: the id's stability matters for the commit trail and
for the inventory count, not for a page citation that does not exist.

The rest of framing's prose reads as an argument, not a measurement — "roughly
900–1,400 words is where it still gets read" is an editorial judgment. It is
treated as one below, and grounded against numbers this repo can actually check.

---

## The reflow demonstration

Framing's central claim: **`wc -l` measures the wrong event**, because on a
hard-wrapped file it conflates adding content with rewrapping a paragraph.
That claim is either demonstrable or the spec is weaker than it sounds. It is
demonstrable.

### Method

A pure rewrap of `README.md` — every unprefixed prose paragraph rejoined and
re-broken at a given column, code fences and blockquotes untouched, **no word
added, removed, or altered.** Invariance of the content is not assumed: the
script asserts `new.split() == original.split()` at every width and aborts if
it ever differs. It never did.

(BSD `fmt`, the obvious tool, is **not** UTF-8-safe — it silently eats the
README's em dashes, which is itself a content change. That first attempt was
discarded; the rewrap is done with an explicit greedy wrapper instead. Recorded
because it is exactly the kind of tool artifact that would have produced a
false demonstration.)

### Result

Rewrapping only the prose paragraphs, at widths from 64 to 100 columns:

| Width | `wc -l` | `wc -w` | token stream |
|---:|---:|---:|---|
| baseline | 260 | 1268 | — |
| 64 | 267 | 1268 | identical |
| 68 | 265 | 1268 | identical |
| 72 | 263 | 1268 | identical |
| 76 | 259 | 1268 | identical |
| 84 | 254 | 1268 | identical |
| 92 | 250 | 1268 | identical |
| 100 | 248 | 1268 | identical |

**`wc -l` moves across a 19-line range (248…267). `wc -w` does not move at
all.** Under the current band `100..260`, four of those seven rewraps turn CI
red having added nothing.

This is not a hypothetical. `README.md` is **not** consistently wrapped today:
measured over its 65 prose lines, the intro section maxes at 68 columns while
`### Upgrading from 0.5.1 or earlier`, `## Capture an entry`, `## Using brag
from an AI agent (MCP)` and `## How this repo is built` all run to 77–79.
Normalising the file to a single width — a formatting change with no content in
it, and the first thing any markdown formatter would do — fails A1.

### The honest limit of the claim

`wc -w` is reflow-invariant on **unprefixed** lines. It is not universally
reflow-invariant, and the spec should not pretend otherwise: a blockquote's `>`
and a list bullet's `-` are themselves whitespace-delimited tokens, so
rewrapping *those* moves both counts by the same absolute amount. Measured on
this README's 16 blockquote lines:

| Width | `wc -l` delta | `wc -w` delta |
|---:|---:|---:|
| 58 | +3 | +3 |
| 70 | +1 | +1 |
| 82 | −2 | −2 |
| 100 | −4 | −4 |

So the improvement is not "invariant everywhere". It is: **strictly invariant
over the 82 unprefixed content lines (65 prose + 17 list), and equally
sensitive in absolute terms over the 16 blockquote lines** — where ±4 lands
against a 500-word band instead of a 160-line one, i.e. 0.8% of the budget
instead of 2.5%. Stated in the helper's own comment so a future reader does not
have to rediscover it.

### What this does *not* argue

It does not argue that the README needs more room. The switch is justified by
the instrument being wrong, and it would be justified in exactly the same terms
if Change 1 did not exist. That is checkable, and was checked: with the metric
switch applied and **the README left completely untouched at 1,268 words**, the
full harness is green (see §12(b), run 2). The two changes are independent.

---

## Scope

- `README.md` — the call-to-action leaves the Status blockquote and becomes its
  own section (Change 1).
- `scripts/test-docs.sh` — a new `assert_word_count_band` helper, A1 switched
  to it with a new band, and a new `A11` holding the call-to-action in place
  (Change 2 plus the guard Change 1 needs).
- `docs/engineering-practices.md` — **one derived number**, re-pasted, not
  rewritten. Adding `A11` moves the distinct-assertion-id count from 176 to
  177, and that count lives in the page's guarded inventory block. This is the
  block's own documented remedy (`just inventory`, paste), not a redesign of
  it; see the Implementer note ⑤.

---

## Fork 1 — the band: **`900 1400`, in words**

`assert_word_count_band "A1" "README.md" 900 1400`.

**The number is chosen, and the choice is testable for honesty.** The test:
would this ceiling be the same if Change 1 did not exist? Yes — 1,400 is
derived below from the README's job and from this repo's own document tiers,
with no reference to what Change 1 costs. (For the record, Change 1 costs
**+9 words**: 41 out of the blockquote, 50 into the new section. The ceiling
would have to be *below 1,277* for the two to interact at all, and it is not
close.) The commitment that goes with that test: had the honest ceiling landed
under the post-change count, the correct move would have been to cut README
words, not to move the ceiling.

### Why 1,400 for the ceiling — three independent legs

1. **The guard must fire before the failure, not at it.** Framing's editorial
   judgment is that past roughly 1,500 words a README stops being read and
   starts being skimmed. A ceiling *at* the failure point is useless: it
   permits the README to become unreadable and only objects afterwards. 1,400
   leaves the guard firing while the file is still fixable.
2. **The README must keep its tier in this repo's own document set.** Measured
   2026-08-20: `README.md` 1,268 · `GETTING_STARTED.md` 828 ·
   `docs/for-ai-agents.md` 2,120 · `docs/engineering-practices.md` 2,245 ·
   `BRAG.md` 2,916 · `docs/tutorial.md` 6,407 · `docs/api-contract.md` 10,403.
   A1 exists (SPEC-021) to keep the README a front door rather than a manual.
   A ceiling of 1,400 preserves a ~700-word gap to the smallest deep-dive doc;
   a ceiling of 2,000 would erase the tier and, with it, A1's purpose.
3. **The headroom is a stated budget, not slack — and this is the asymmetry
   framing identified.** Under `wc -l`, granted headroom is spent mostly on
   non-events: reflow, code-fence delimiters, blank lines. Under `wc -w`,
   every word of headroom is a word of *content*, so headroom cannot be
   consumed by typography at all. The gap between the post-change 1,277 and
   1,400 is 123 words ≈ **one short section** (SPEC-079's `## How this repo is
   built` is 82 words). That is a legible allowance: a reflow, a typo fix, or a
   clarifying sentence never trips the guard; a second new section does, and is
   again a decision — which is precisely what LD5 wanted and could not get in
   lines.

### Why 900 for the floor

The floor is a backstop against the README being **hollowed out** — a bad
merge, an over-eager trim, a section deleted in a refactor and never noticed.
The A-series above it already pins *which* sections exist (A4 install, A5 the
seven verbs, A6 the data path, A7–A9 the pointers); the floor pins that those
sections still have content in them. Below ~900 words the README can no longer
carry install, capture, retrieval and pointers at usable depth.

### The band is narrower than the one it replaces — measured, not asserted

This is the fact that settles whether the switch disarms anything. Converted at
the README's own words-per-line ratio (1268 / 260 = 4.88 w/l):

| | floor | ceiling | span |
|---|---:|---:|---:|
| old, in lines | 100 | 260 | 160 |
| old, in word-equivalents | ~490 | ~1268 | ~780 |
| new, in words | 900 | 1400 | 500 |

**The admissible band narrows by 36%.** The ceiling gains ~130 words of
budget; the floor tightens by ~410. A change that shrinks the guard's
admissible region by more than a third is not a relaxation, whatever it does to
one edge. (Caveat stated plainly: the conversion is only meaningful at today's
line/word mix — it is an approximation used to compare, not a second guard.)

### Rejected alternatives (build-time)

- **`900 1268` — ceiling at today's exact count — REJECTED.** It repeats LD5's
  zero-headroom trade in a new unit. LD5 accepted that cost knowingly and
  stated it ("any future README edit that adds a line — including a reflow —
  fails A1 and must re-pin"), and this spec is the bill arriving. Repeating it
  reproduces the friction *and* wastes the one property that makes headroom
  safe in words but not in lines.
- **`900 1500` — ceiling at the readability limit — REJECTED.** Puts the alarm
  at the fire rather than before it, and buys a further 100 words for nothing:
  nothing in the backlog needs them.
- **`1000 1500`, `800 1600`, or any round-number pair — REJECTED.** Roundness
  is not a justification, and both ends here are load-bearing in different
  ways; picking them together for symmetry hides that.
- **No ceiling at all; keep only the floor — REJECTED.** The README's failure
  mode in this repo is accretion (four specs have added to it), not deletion.
  Dropping the ceiling drops the half that has actually fired.
- **A percentage-of-baseline band recomputed from git history — REJECTED.**
  Unrottable in principle, unreadable in practice, and it makes every failure
  message require an archaeology session. The existing helpers' idiom is two
  integers; a guard nobody can read in one line does not get respected.

---

## Fork 2 — shape: **a sibling helper; A1 keeps its id; the other five stay on lines**

Three sub-decisions.

### (a) Sibling helper, not a repurpose — confirmed against the callers

Framing's pre-check holds: `assert_line_count_band` has **six call sites** —
`A1` (:123), `C2` (:271), `D2` (:292), `J2` (:588), `T2` (:1027), `X7` (:1410)
— plus its definition at :45. Changing its body would silently re-scope five
unrelated guards. `assert_word_count_band` is added beside it.

**This needs no new assertion to enforce, because the five untouched callers
already are that test.** If build repurposes the helper instead of adding a
sibling, all five go red immediately, since none of those files' word counts
falls inside its line band:

| id | file | line band | words |
|---|---|---|---:|
| C2 | `CONTRIBUTING.md` | 30..120 | 253 |
| D2 | `docs/development.md` | 50..200 | 474 |
| J2 | `examples/brag-slash-command.md` | 5..30 | 77 |
| T2 | `docs/for-ai-agents.md` | 120..500 | 2120 |
| X7 | `docs/engineering-practices.md` | 150..300 | 2245 |

There is a matching free property on the other side: a `wc -w`/`wc -l`
transcription slip inside the *new* helper is caught by A1's own floor, because
the README's 262 lines fall below the 900-word floor. Mutation-confirmed below.

### (b) A1 keeps its id

**Yes.** Two reasons, one of them mechanical:

- The id is a stable handle in 14 commit-message lines on `main` and in three
  spec files. Renaming it breaks the trail for no gain. (It is *not* cited on
  the practices page — framing said it was; it is not. See Re-measurement.)
- Mechanically, the id names the **slot** — "the README's size guard" — not the
  instrument. Keeping it means `scripts/inventory.sh`'s distinct-id set gains
  only `A11` (176 → 177) instead of churning A1 in and out.

### (c) The other five stay on `assert_line_count_band` — the scope line is measured, not asserted

Framing said "resist scope creep here". The measured version of that: **the
reflow defect is only LIVE where a file's reflow swing exceeds its guard's
remaining headroom.** Measured 2026-08-20, rewrapping each guarded file's prose
across 64–100 columns (token stream asserted identical every time):

| id | file | band | now | headroom | reflow range | max swing | live? |
|---|---|---|---:|---:|---|---:|---|
| A1 | `README.md` | 100..260 | 260 | **0** | 248…267 | **+7** | **YES** |
| C2 | `CONTRIBUTING.md` | 30..120 | 55 | 65 | 52…56 | +1 | no |
| D2 | `docs/development.md` | 50..200 | 91 | 109 | 87…94 | +3 | no |
| J2 | `examples/brag-slash-command.md` | 5..30 | 14 | 16 | 9…14 | +0 | no |
| T2 | `docs/for-ai-agents.md` | 120..500 | 284 | 216 | 259…311 | +27 | no |
| X7 | `docs/engineering-practices.md` | 150..300 | 283 | **17** | 271…298 | **+15** | no |

A1 is the only caller where the defect can fire today. **X7 is two lines from
joining it** and is named as the next one, with a stated trigger rather than a
vague intention: *switch a caller when its reflow swing exceeds its headroom,
not before.* That trigger is written into the helper's comment, where whoever
next touches the harness will read it.

### Rejected alternatives (build-time)

- **Repurpose `assert_line_count_band` to count words — REJECTED**, per (a);
  five unrelated guards would silently change meaning, and all five would fail.
- **Switch all six now — REJECTED.** Four of them have 40–55% headroom; the
  change would be motion without evidence, and it would put five files' size
  bands up for renegotiation inside a spec framed at S. The trigger above is
  the disciplined version.
- **Retire `A1` and add `A12` for the word guard — REJECTED**, per (b): it
  breaks the commit trail and churns the derived id count for nothing.
- **Add a `--metric` parameter to one helper (`assert_size_band "A1" f words
  900 1400`) — REJECTED.** It is the repurpose with extra steps: one body, six
  callers, and a new way to get the argument order wrong. Two four-line
  functions are cheaper to read than one parameterised one.
- **Count prose words only, stripping fences and markers with `awk` —
  REJECTED.** It would make the guard's number impossible to reproduce by hand
  (`wc -w README.md` would no longer match the failure message), and code
  examples *are* content a size guard should count. The blockquote-marker
  imprecision it would fix is quantified above at ±4 words in 500.

---

## Fork 3 — placement and volume: **its own `##` section immediately after `## Install`**

`## Wire it into your AI agent`, inserted after the Install section's last line
(`README.md:77`) and before `## Capture an entry`. Four lines of copy, one
inline command, one link down. Literal ① carries it verbatim.

**Why a named section rather than a lead-in paragraph.** The problem being
fixed is that the copy is invisible to a *skimmer*. On GitHub a `##` heading
appears in the rendered outline; a sentence — in a blockquote or in a
paragraph — does not. A lead-in paragraph would move the copy without fixing
the thing that made it unread.

**The trade, stated rather than glossed.** This moves the copy *down* the page:
from line 16 to line 72. That is a real cost and the spec should not claim
otherwise. What it buys, measured in the outline the skimmer actually uses: the
call-to-action goes from **absent from the outline entirely** to the **third of
ten `##` headings**. The bet is that presence-in-the-outline beats
height-in-a-blockquote — which is the same bet framing made, and, like
framing's, it is reasoned rather than validated. bragfile has one human author
and no traffic; there is no signal to read. Saying so is the point.

**Why after Install rather than before it.** The section's whole content is a
command to run. Telling a reader to run `brag mcp install` before they have the
binary is the wrong order, and Install is only 56 lines.

**What it must not duplicate from the section at line 219.** The lower section
keeps *all* the detail — the fenced command with its client/scope comment, the
"reconnect your client" instruction, the enumeration of tool schemas and
`brag://` resources, and the pointer to `docs/for-ai-agents.md`. The new
section carries the command **inline, not fenced**, and one link down. Detail
belongs low; the budget is words; and two byte-identical fenced blocks 140
lines apart is redundancy nobody should pay for.

**Heading voice.** The README's `##` headings are verb phrases — Install,
Capture an entry, Read entries back, Export for reviews. "Wire it into your AI
agent" matches, and is deliberately *not* a near-duplicate of `## Using brag
from an AI agent (MCP)`; the anchor `#using-brag-from-an-ai-agent-mcp` that the
copy links to is unaffected.

### Rejected alternatives (build-time)

- **A lead-in paragraph under `## Install` with no heading — REJECTED**, per
  above: invisible to the outline, which is the failure being fixed.
- **A section before `## Install` — REJECTED.** Instructs the reader to run a
  command they cannot run yet.
- **Splitting `## Install` so the call-to-action lands higher (before
  `### Upgrading from 0.5.1 or earlier`) — REJECTED.** It orphans the Upgrading
  subsection under the wrong `##`, and that subsection documents a *silent*
  upgrade failure. Restructuring Install to win 20 lines of scroll is a README
  restructure, which this spec is not.
- **Repeating the fenced `brag mcp install` block — REJECTED**, per above.
- **Moving the full MCP section up to line 79 instead — REJECTED**, and
  explicitly out of scope per framing. Detail belongs low; the point is a
  signpost, not relocation.
- **A badge or a callout admonition — REJECTED.** GitHub's `> [!NOTE]` syntax
  renders as a blockquote, which is the container this spec is getting the copy
  *out* of.

---

## Fork 4 — the Status blockquote: **keep it exactly as it is, minus the call-to-action and its `>` separator**

`README.md` lines 15–19 are deleted (the bare `>` separator plus the four
call-to-action lines). Lines 10–14 are **untouched**, including —
load-bearingly — line 10:

```
> **Status:** v0.6.1 shipped. Capture, retrieve, search, export, digests,
```

`W3` (`scripts/test-docs.sh:1280`) greps `^> \*\*Status:\*\* v${latest} shipped`
with `${latest}` derived from the CHANGELOG's newest dated heading. Confirmed
at design: the line survives byte-identical, and `W3` is green in every
pre-flight run below. This is SPEC-077's guard, added after the version claim
rotted through two releases — the blockquote is not a candidate for removal and
the prefix is not decoration.

**Why keep the rest.** The remaining blockquote is a *release-state* claim —
what v0.6.1 contains. The new section is an *instruction*. They overlap in
subject (both mention the MCP server) but not in function, and the blockquote's
capability list is the thing `W3` exists to keep honest.

### Rejected alternatives (build-time)

- **Also trimming the blockquote's two MCP sentences now that a dedicated
  section exists — REJECTED**, though the duplication is real and worth naming:
  the blockquote spends about half its words on MCP and `brag memory`, and the
  new section says a version of the same thing 60 lines later. Rejected anyway
  because (i) it deletes accurate release-state content to save ~30 words
  against a 123-word budget that is not under pressure, (ii) every edit inside
  that blockquote is an edit next to the one guard in this file that exists
  because this exact region rotted twice, and minimum-touch is the right
  posture there, and (iii) framing scoped this as a move plus a heading and
  explicitly warned against creep. If the duplication becomes annoying it is a
  one-line follow-up with the guard already in place.
- **Deleting the blockquote and putting Status in a table or badge —
  REJECTED.** Breaks `W3` outright.
- **Leaving the `>` separator line behind — REJECTED.** A blockquote ending in
  a bare `>` renders as a trailing empty line inside the quote.

---

## V2 (`internal/cli/window.go`) — **not folded in**

Framing routed this as design's call. The decision is **no**, and the routing
note is corrected on the way past.

What V2 actually is, verified at design: `internal/cli/window.go:12-13` reads
*"windowFlagNames is the canonical, ordered set of calendar-window flags shared
by `brag impact` and `brag story`"* — but `internal/cli/coverage.go:59,68` calls
`selectedWindow`/`windowCutoff` too. **Three consumers, not two.** (`spark.go`
is *not* one: its comments at :73 and :166 say explicitly that it governs its
own flag set.) Note also that SPEC-080's re-verify recorded the path as
`internal/storage/window.go`; the file is `internal/cli/window.go`.

**Why not fold it in:**

1. `windowFlagNames` is **unexported**, so its comment never reaches `go doc`.
   STAGE-021's success criterion is *"`go doc` on the public surface reads as
   intentional"* — this symbol is outside it. Folding it in would be doing
   STAGE-021's godoc work in the spec that is not the godoc spec.
2. The spec's surface is two files, and its S sizing rests on that. A Go source
   file is a third gate surface for a one-line comment.
3. It has no guard and cannot get one, in a spec whose entire thesis is that
   claims are held by mechanical checks rather than by remembering. It would
   ride in as exactly the kind of unguarded prose claim SPEC-079's verify
   punch-listed.

**What it costs:** the item stays unrouted as STAGE-021 closes. That is a real
cost, mitigated the only way available here — it is re-recorded below under
"Out of scope" with the corrected path, the corrected fact (three callers), and
the exact line, so whoever picks it up does not have to re-derive any of it.

---

## §12(b) design-time verification

Every literal in this spec was applied to a real working tree, run through the
real harness, mutation-tested, and then reverted. The working tree is clean;
build transcribes.

**Run 1 — Change 1 alone, old harness.** Staged literal ① only.
`./scripts/test-docs.sh` →
`FAIL: A1: README.md has 262 lines (expected 100..260)` and **nothing else**.
Confirms framing's sequencing claim exactly: the call-to-action is blocked by
A1's zero headroom, and by nothing else in the harness.

**Run 2 — Change 2 alone, README untouched.** Reverted ①, staged literals ②③④.
Result: `A1` **passes** (1268 words, inside 900..1400); `A11` fails
(`first '## …agent…' heading is at line 219 of 260 (must be within the first
third, i.e. line 86)`) — the correct fail-first, naming the real defect; `X3`
fails on the id count. **This run is the load-bearing one for the no-go
condition:** the metric switch is green against the *unmodified* README, so it
is not carrying Change 1.

**Run 3 — everything staged.** ①②③④ plus the inventory re-paste (⑤).
`./scripts/test-docs.sh` → **ALL OK**. `bash -n scripts/test-docs.sh` clean.
`./scripts/inventory.sh` prints 177 for the assertion row; the practices page
diff is exactly one character (`176` → `177`) and its line count is unchanged
at 283, so `X7` is unaffected.

**Run 4 — mutation testing.** Each mutant applied to the fully-staged tree, then
reverted:

| Mutant | Result |
|---|---|
| README padded to 1401 words | `FAIL: A1: README.md has 1401 words (expected 900..1400)` |
| README at exactly 1400 words | green — the ceiling is inclusive |
| README truncated to 899 words | `FAIL: A1: README.md has 899 words (expected 900..1400)` |
| README at exactly 900 words | green — the floor is inclusive |
| `## Wire it into your AI agent` heading deleted | `FAIL: A11: first '## …agent…' heading is at line 219 of 260…` |
| the section moved down to just before `## License` | `FAIL: A11: first '## …agent…' heading is at line 215 of 262…` |
| `brag mcp install` removed from the section body | `FAIL: A11: the agent call-to-action section must name 'brag mcp install' within 8 lines of its heading` |
| `wc -w` mistyped as `wc -l` inside the new helper | `FAIL: A1: README.md has 262 words (expected 900..1400)` |

Every mutant is killed, and the last one confirms the transcription-error
property claimed in Fork 2(a).

**Run 5 — factual claims in the promoted copy.** "five typed tools":
`internal/mcpserver/server.go` registers `brag_add`, `brag_list`, `brag_search`,
`brag_stats`, `brag_memory` at :38–:58 — five (DEC-045 added the fifth,
amending DEC-024's "four"). "an auto-loadable memory resource":
`brag://memory/recent`, per DEC-045. Both hold; the copy moves verbatim.

**Run 6 — Go gates, with everything staged.** `go test ./...` all pass
(one `[no test files]` note for `internal/storage/storagetest`, pre-existing);
`gofmt -l .` empty; `go vet ./...` clean. Expected — this spec touches no Go —
but run rather than assumed.

---

## Outputs

- **`README.md`** — lines 15–19 deleted; a six-line `## Wire it into your AI
  agent` section inserted after line 77. Net **+2 lines, +9 words**: 260 → 262
  lines, 1,268 → 1,277 words.
- **`scripts/test-docs.sh`** — three edits:
  1. new helper `assert_word_count_band`, inserted between
     `assert_line_count_band` and `assert_contains_literal`;
  2. `A1`'s comment block and call replaced —
     `assert_line_count_band "A1" "README.md" 100 260` becomes
     `assert_word_count_band "A1" "README.md" 900 1400`;
  3. new assertion `A11`, appended to Group A immediately before the
     `# ===== Group B` header.
- **`docs/engineering-practices.md`** — the guarded inventory block re-pasted;
  one number changes (176 → 177). No prose change, no line-count change.
- **New exports:** none. **Database changes:** none. **New DEC files:** none —
  nothing here is a repo-level architectural choice; the reasoning is local to
  one assertion and is written into the assertion's own comment, which is where
  the next person to touch it will look.

### Premise audit (§9), run at design against the repo

The **additive** case applies (this spec adds to a tracked collection whose
count is asserted elsewhere) and the **status-change** case was checked and does
not. Both greps were **executed**, not merely enumerated:

| Grep | Hits | Disposition |
|---|---|---|
| `grep -n 'assert_line_count_band' scripts/test-docs.sh` | 7 (definition :45 + six call sites) | **One planned modification** (A1, :123). The other five are deliberately untouched — and are themselves the test that the helper was not repurposed (Fork 2a). |
| distinct assertion ids via `scripts/inventory.sh`'s own pipeline | 176 → 177 | **Planned update.** `A11` is the only new id; `A1` is reused. Confirmed by running the pipeline against the staged script. The count is asserted by `X3` against the practices page → literal ⑤. |
| `grep -rn 'line count\|line-count\|wc -l' docs/ README.md CONTRIBUTING.md AGENTS.md BRAG.md justfile` | 2, both in `docs/reports/cost-estimate-2026-06-11T233221.md` (an unrelated LOC methodology note) | **No change needed.** No document describes A1's instrument, so switching it makes no prose stale. |
| `grep -rn 'A1' docs/engineering-practices.md` | 0 | **No change needed** — and it corrects framing (see Re-measurement). |
| `grep -n 'README' docs/engineering-practices.md` | 3 (:134 a coverage list, :151 the W3 description, :196 an install-note link) | **No change needed.** None describes the README's size or A1. |
| `grep -cE '^# W[0-9]+ ' scripts/test-docs.sh` (the `wseries` inventory row) | 6, unchanged | **No change needed.** The new comments do not open with a `# W<n> ` marker, so the W-series row is untouched — confirmed by running `inventory.sh` against the staged script. |

---

## Acceptance Criteria

1. `README.md` contains no call-to-action inside the `> **Status:**`
   blockquote; lines 15–19 of the pre-change file are gone, and the blockquote
   ends at the sentence naming `brew install jysf/tap/bragfile`.
2. `README.md` contains a `## Wire it into your AI agent` section, byte-identical
   to literal ①, positioned after the Install section and before
   `## Capture an entry`.
3. `README.md:10` is byte-identical to its pre-change self, and `W3` is green.
4. `scripts/test-docs.sh` defines `assert_word_count_band` and it is called by
   exactly one assertion, `A1`, with the band `900 1400`.
5. `assert_line_count_band` is unchanged in body, and still called by exactly
   five assertions: `C2`, `D2`, `J2`, `T2`, `X7`.
6. `A11` exists, passes on the new README, and fails on the old one.
7. `just test-docs` reports **ALL OK** at **177 distinct assertion ids**.
8. `docs/engineering-practices.md`'s inventory block equals `./scripts/inventory.sh`
   output byte-for-byte (`X3` green), and the page's line count is unchanged
   at 283 (`X7` green).
9. `just test`, `gofmt -l .`, `go vet ./...` unaffected.

---

## Failing Tests

Written here per `test-before-implementation`; made to pass at build. All three
were run in the states described — this is a record of observed output, not a
prediction.

### Changed: `A1` — instrument and band

```
- assert_line_count_band "A1" "README.md" 100 260
+ assert_word_count_band "A1" "README.md" 900 1400
```

**Fails before, for the right reason.** With literal ① staged and the harness
unchanged:

```
FAIL: A1: README.md has 262 lines (expected 100..260)
FAILED: 1 assertion(s) failed.
```

That is the *only* failure, which is what makes the sequencing claim in
framing's Go/no-go concrete: A1's zero headroom, and nothing else, blocks
Change 1.

**Passes after**, and its boundaries are inclusive and mutation-proven (1401
red / 1400 green / 899 red / 900 green — §12(b) run 4).

### Added: `A11` — the call-to-action is a scannable heading near the top

Full literal at ④. **Fails before**, against the unmodified README:

```
FAIL: A11: first '## …agent…' heading is at line 219 of 260 (must be within the first third, i.e. line 86)
```

**Passes after** (heading at line 72 of 262, threshold 87; `brag mcp install`
found within 8 lines). Three mutants killed — heading deleted, section moved to
the bottom, command dropped — §12(b) run 4.

`A11` pins no line number: its threshold is derived from the file's own length,
and `A1` bounds that length, so the two compose rather than drifting apart.

### Changed by re-derivation, not by hand: `X3`

`X3` goes red the moment `A11` lands — *"inventory block is stale — run `just
inventory` and paste between the markers"* — and green again after the paste. It is listed here because it *will* fail during build and its failure is
expected and mechanical — not because its expectation is being edited. **Nobody
types 177.**

### Not added, deliberately

No assertion was added for Fork 2(a) ("build must add a sibling, not repurpose
the helper"). The five untouched callers already fail if it is repurposed
(Fork 2a's table), so a dedicated assertion would be a guard on a guard, and it
would move the derived id count a second time for zero coverage.

### §12 NOT-contains self-audit

This spec adds no NOT-contains assertion. The audit is run in the other
direction, which is the load-bearing one here: every token the harness already
asserts **absent** from `README.md` was grepped against the exact literal this
spec adds to `README.md` (load-bearing prose — it renders into the file the
assertions read). Executed at design against literal ①:

| Assertion | Pattern | Hits |
|---|---|---:|
| B1 | `spec-driven` | 0 |
| B2 | `frame.*design.*build.*verify.*ship` / `frame → design` / `frame -> design` | 0 |
| B3 | `four habits` | 0 |
| B4 | `context contamination` | 0 |
| B5 | `just (new-spec\|advance-cycle\|archive-spec\|weekly-review\|new-stage)` | 0 |
| B6 | `claude plays every role` | 0 |
| B7-heading | `^## .*table of contents` | 0 |
| B7-toc | `^- \[` (4+ contiguous, first 50 lines) | 0 |
| A3 | `spec-driven\|architect\|implementer\|reviewer\|cycle\|hierarchy` (first 12 lines) | 0 — and README.md:1-12 is untouched by this spec |

Total hits: **0**. Confirmed independently by run 3 (whole harness green).

---

## Implementation Context

### Decisions that apply

- **`DEC-024`** — the MCP server's SDK, stdio transport and provenance
  stamping. The original "four typed tools" line lives here.
- **`DEC-045`** — the push surface: three `brag://` resources plus
  `brag_memory` as the **fifth** tool, amending DEC-024 without rewriting it.
  Together these are what make the promoted copy's "five typed tools plus an
  auto-loadable memory resource" true; both were re-checked against
  `internal/mcpserver/server.go` at design (§12(b) run 5).

No new DEC. This is a change to one assertion's instrument and one README
section's placement — local reasoning, recorded at the site.

### Constraints that apply

- **`test-before-implementation`** (blocking) — the A1 switch and A11 are
  written above and were observed failing before implementation.
- **`one-spec-per-pr`** (blocking) — one spec, one branch, one PR. Part of why
  V2 is declined rather than folded in.

No constraint in `guidance/constraints.yaml` touches `README.md`,
`scripts/test-docs.sh`, or `docs/`; the remaining nine are Go/SQL/secrets rules
on paths this spec does not touch. Checked, not assumed.

### Prior related work

- **`SPEC-021`** — created `scripts/test-docs.sh` and set A1's original band.
  A1 exists to keep the README a user-facing quickstart. That purpose is
  carried forward unchanged; only the instrument moves.
- **`SPEC-077`** — added `W1`–`W3` after the README status line and the
  tutorial's "shipped as of" line rotted through v0.5.2 *and* v0.6.0. `W3` is
  why Fork 4 is a minimum-touch decision.
- **`SPEC-079`** — added `## How this repo is built` and re-pinned A1 tight to
  260 (LD5). **LD5 is the decision this spec revisits, and it is not
  overturned.** LD5's reasoning — a guard widened whenever it fires is not a
  guard — is accepted in full; what this spec disputes is the *instrument*, and
  it pays LD5's own stated cost forward by narrowing the band overall.
- **`SPEC-080`** — the stage's godoc pass; recorded V2, which this spec
  declines with reasons.

### Out of scope (for this spec specifically)

- **Moving or rewriting the full MCP section** (`README.md:219`). It stays.
- **`docs/engineering-practices.md`'s prose, and `scripts/inventory.sh`.** The
  only permitted touch is re-pasting the guarded block after `just inventory` —
  the block's own documented remedy.
- **Anything in STAGE-022**, including the four items SPEC-080 routed there.
- **`internal/cli/window.go:12-13`** — the doc comment on `windowFlagNames`
  names `impact` and `story` but the flag set is shared by **three** commands:
  `coverage` calls `selectedWindow`/`windowCutoff` at
  `internal/cli/coverage.go:59,68`. (`spark` does not — see its notes at
  `internal/cli/spark.go:73,166`.) A one-line prose fix. Recorded here with the
  corrected path — SPEC-080's re-verify wrote `internal/storage/window.go`,
  which does not exist — and the corrected count, so it can be picked up
  without re-derivation. Deliberately not folded in; reasons above.
- **Switching C2/D2/J2/T2/X7 to word bands.** Trigger stated in Fork 2(c) and
  written into the helper's comment.

---

## Notes for the Implementer

Five literals. **Build transcribes; verify diffs.** Nothing below is a sketch.

**Two mechanical warnings before you start.**

1. **Anchor on line-start matches and check the match count before writing.**
   This spec embeds literal copies of markdown headings and shell comments, so
   several strings below occur more than once *in this file*. Every edit
   described here has exactly one anchor in its target file; if an anchor
   matches zero or two times, stop — the target has drifted from what design
   measured.
2. **The derivation outranks the cache.** Numbers appear inside literals ②③④
   as **dated measurements** (`measured 2026-08-20`) — historical facts that do
   not rot, and you should transcribe them verbatim even if the README's
   current count differs by then. Literal ⑤ is the opposite: it is *derived
   output*, and the value quoted there (177) is a prediction. Run
   `just inventory` and paste **whatever it prints**. If it prints something
   other than 177, that is unrelated drift on `main` and the paste is still the
   correct action; do not reconcile it by hand.

### ① `README.md` — the call-to-action moves

**Delete lines 15–19** (the bare `>` separator plus the four call-to-action
lines). The blockquote must end at line 14 — the line beginning
`> client can auto-load with no tool call.` — and **line 10 must not be
touched**; `W3` greps it literally.

Removed block, for identification:

```
>
> Working with an AI agent? `brag mcp install` wires brag into Claude Code,
> Cursor, or Claude Desktop as five typed tools plus an auto-loadable memory
> resource — see
> [Using brag from an AI agent (MCP)](#using-brag-from-an-ai-agent-mcp) below.
```

**Insert**, after the Install section's last line (`  preview) — see
[`docs/api-contract.md`](docs/api-contract.md).`) and its following blank line,
and before `## Capture an entry`:

```markdown
## Wire it into your AI agent

Working with an AI agent? `brag mcp install` wires brag into Claude Code,
Cursor, or Claude Desktop as five typed tools plus an auto-loadable memory
resource. The flags, the resources and the gotchas are in
[Using brag from an AI agent (MCP)](#using-brag-from-an-ai-agent-mcp) below.
```

Exactly one blank line before the heading and one after the paragraph. Result:
**262 lines, 1,277 words.** Check with `wc -l README.md; wc -w README.md`
before running the harness.

### ② `scripts/test-docs.sh` — the new helper

Insert immediately after `assert_line_count_band`'s closing `}` and its blank
line, i.e. directly before the line `assert_contains_literal() {` (one match in
the file):

```bash
# Word count, not line count. `wc -l` on a HARD-WRAPPED prose file conflates two
# different events: adding content (what a size guard exists to catch) and
# rewrapping a paragraph (which adds nothing). Measured at SPEC-081 design,
# 2026-08-20: rewrapping README.md's prose across 64..100 columns — token stream
# provably byte-identical, not one word added, removed or changed — moves
# `wc -l` from 248 to 267, while `wc -w` stays at exactly 1268 every time. A
# guard that fires on a non-event gets disarmed by habit, which is the
# "bump the number until green" failure SPEC-079's LD5 refused.
#
# The honest limit: `wc -w` is reflow-invariant on UNPREFIXED lines only. A
# blockquote's `>` and a list's `-` are themselves words, so rewrapping those
# moves both counts by the same absolute amount — measured on README.md, at
# most ±4. That is ±4 against a 500-word band here, where it was ±4 against a
# 160-line one before.
#
# The other five callers stay on `assert_line_count_band` deliberately: the
# defect is only LIVE where a file's reflow swing exceeds its guard's headroom,
# and measured 2026-08-20 only A1 qualified (headroom 0, swing +7). The next
# closest is X7 (headroom 17, swing +15). Switch a caller to this helper when
# its swing exceeds its headroom, not before.
assert_word_count_band() {
    name="$1"; path="$2"; min="$3"; max="$4"
    if [ ! -f "$path" ]; then
        fail "$name" "file does not exist: $path"
        return 0
    fi
    n=$(wc -w < "$path" | tr -d ' ')
    if [ "$n" -ge "$min" ] && [ "$n" -le "$max" ]; then
        ok "$name"
    else
        fail "$name" "$path has $n words (expected $min..$max)"
    fi
}
```

Followed by one blank line. `assert_line_count_band` itself is **not modified**.

### ③ `scripts/test-docs.sh` — A1's block, replaced

Replace this exact block (the eight lines from `# A1 — README line count band`
through the `assert_line_count_band "A1"` call; one match):

```bash
# A1 — README line count band 100..260
# Re-pinned 250 -> 260 at SPEC-079 (LD5), TIGHT: 260 is the exact length of the
# README after the `## How this repo is built` section, not a round number with
# headroom. A guard widened whenever it fires is not a guard — the next line
# added to the README fails this again and is again a decision. The band exists
# (SPEC-021) to keep the README a user-facing quickstart, and a ten-line routing
# section with no commands, flags or numbers does not change that.
assert_line_count_band "A1" "README.md" 100 260
```

with:

```bash
# A1 — README word count band 900..1400.
#
# The INSTRUMENT changed at SPEC-081, from `wc -l` to `wc -w`; the band's
# purpose (SPEC-021) is unchanged — keep the README a user-facing front door
# rather than a manual. See `assert_word_count_band` above for why a line count
# is the wrong instrument on a hard-wrapped file.
#
# This is NOT the old band widened. Converted at the README's own words-per-line
# ratio measured 2026-08-20 (1268 words / 260 lines = 4.88), the old 100..260
# lines was ~490..1268 words: a span of 780. This band spans 500 — 36% narrower.
# The ceiling gains ~130 words of budget; the floor tightens by ~410.
#   ceiling 1400 — past roughly 1500 words a README gets skimmed rather than
#     read, so the guard fires BEFORE that point, not at it. The smallest
#     deep-dive doc in this repo (`docs/for-ai-agents.md`, 2120 words on
#     2026-08-20) sits well clear of it, so the README keeps its tier as the
#     front door rather than becoming a second manual. The gap between the
#     README's count and 1400 is a budget of about one short section —
#     SPEC-079's `## How this repo is built` is 82 words — not slack: under
#     `wc -w` headroom can only ever be spent on content, which is why a word
#     budget tolerates headroom that a line budget cannot.
#   floor 900 — a backstop against the README being hollowed out by a bad merge
#     or an over-eager trim. Below 900 words it can no longer carry install,
#     capture, retrieval and pointers at usable depth. The A-series above pins
#     WHICH sections exist; this pins that they still have content in them.
assert_word_count_band "A1" "README.md" 900 1400
```

### ④ `scripts/test-docs.sh` — the new `A11`

Insert immediately before the line
`# ===== Group B — README shape (negative — load-bearing) =====` (one match),
followed by one blank line — i.e. `A11` is the last assertion in Group A:

```bash
# A11 — the agent call-to-action is a scannable top-level heading near the top
# of the README, not a sentence inside the Status blockquote.
#
# SPEC-081. The copy already existed — at README.md:16-19, inside
# `> **Status:** …`, where a heading-skimmer never sees it at all and the reader
# who does see it is skipping. Promoting it out of the blockquote is the change;
# this assertion is what holds it there. The threshold is derived from the
# file's own length rather than a pinned line number, so it cannot rot, and A1
# above bounds that length so the two guards compose.
if [ ! -f README.md ]; then
    fail "A11" "README.md does not exist"
else
    a11_line=$(grep -n -i -E '^## .*agent' README.md | head -n 1 | cut -d: -f1)
    a11_total=$(wc -l < README.md | tr -d ' ')
    a11_third=$((a11_total / 3))
    if [ -z "$a11_line" ]; then
        fail "A11" "README.md has no '## …agent…' heading"
    elif [ "$a11_line" -gt "$a11_third" ]; then
        fail "A11" "first '## …agent…' heading is at line $a11_line of $a11_total (must be within the first third, i.e. line $a11_third)"
    elif ! sed -n "$((a11_line + 1)),$((a11_line + 8))p" README.md | grep -F -q 'brag mcp install'; then
        fail "A11" "the agent call-to-action section must name 'brag mcp install' within 8 lines of its heading"
    else
        ok "A11"
    fi
fi
```

Run `bash -n scripts/test-docs.sh` after ②③④ land, before running the harness.

### ⑤ `docs/engineering-practices.md` — re-paste the derived block

**Not a literal. Derived output — run it.**

```bash
just inventory
```

Paste its output between the `<!-- inventory:begin -->` and
`<!-- inventory:end -->` markers, replacing everything between them. Design's
prediction is that exactly one number changes,
`| Documentation assertions (distinct ids) | 176 |` → `| … | 177 |`, and that
the page stays 283 lines so `X7` is unaffected — but **the script is the
source**: paste what it prints. Do not edit the block by hand, and do not touch
any prose on the page.

### Order

②③④ (harness) → `bash -n` → ① (README) → `just test-docs` (expect exactly one
failure, `X3`) → ⑤ (`just inventory`, paste) → `just test-docs` (expect ALL OK
at 177).

Doing ① first is also fine and is a useful fail-first observation
(`A1: README.md has 262 lines`), but then the harness is red for two different
reasons at once, which is noisier.

---

## Confidence

**0.88.**

The mechanical half is as close to certain as this repo gets: every literal was
staged in a real working tree, the full harness was run in three configurations,
eight mutants were killed, the Go gates were run, and the tree was reverted
clean. If build transcribes ①–④ and runs ⑤, the result is known.

The 0.12 is one judgment: **the two numbers.** 1,400 and 900 are argued from
the README's job, this repo's document tiers, and a stated reading-length
heuristic — but "past ~1,500 words a README gets skimmed" is an editorial claim
inherited from framing, not something measured. A reviewer could reasonably
land on 1,300 or 1,500. The design is cheap to be wrong about (a band is one
line, and the narrowing argument survives ±100 at either end), but it is the
part of this spec that is reasoned rather than demonstrated, and the confidence
should say so.

Above 0.8, so no entry is added to `guidance/questions.yaml` — and deliberately
not: STAGE-021's own success criterion is that the register describes **live
uncertainty only**, and "the exact ceiling could be ±100" is a judgment already
recorded here with its reasoning, not an open question with a resolve
condition. The one thing that *does* have a falsifiable trigger — when to
switch the remaining five callers — is written into the helper's comment at the
site where it will be read, per Fork 2(c).

---

## Gates

```
just test
just test-docs
gofmt -l .
go vet ./...
```

All four were run at design against the fully-staged tree: `just test-docs`
ALL OK at 177 ids; `go test ./...` all pass; `gofmt -l .` empty; `go vet ./...`
clean.

---

## Build Completion

*Filled in at the end of the **build** cycle, before advancing to verify.*

- **Branch:** `build/spec-081-readme-cta-and-a1-metric`
- **PR (if applicable):** not opened — per instruction, no push/PR this session.
- **All acceptance criteria met?** yes — all 9 confirmed:
  1. Lines 15–19 of the pre-change `README.md` are gone; the blockquote now
     ends at line 14 (`… brew install jysf/tap/bragfile\`.`).
  2. `## Wire it into your AI agent` lands byte-identical to literal ①
     (diffed), between `## Install` and `## Capture an entry`.
  3. `README.md:10` diffed byte-identical against `main`; `W3` green.
  4. `assert_word_count_band` defined, called by exactly one assertion (`A1`,
     band `900 1400`) — grepped and confirmed one call site.
  5. `assert_line_count_band`'s body diffed byte-identical against `main`;
     still called by exactly `C2`, `D2`, `J2`, `T2`, `X7` (five call sites,
     grepped).
  6. `A11` exists; passes on the new README; confirmed failing on the
     pre-change README during the ordered build sequence (Run 2 shape).
  7. `just test-docs` → **ALL OK** at **177** distinct assertion ids
     (computed via `inventory.sh`'s own dedup pipeline, not a raw `OK:` line
     count — those differ because some ids emit multiple runtime `OK:` lines).
  8. `docs/engineering-practices.md`'s pasted inventory block diffed
     byte-for-byte against a fresh `just inventory` run; page line count
     unchanged at 283 (`X7` green).
  9. `just test`, `gofmt -l .`, `go vet ./...` all green, unaffected.
- **New decisions emitted:** none — matches the spec's own prediction (local
  reasoning, recorded in the assertion comments).
- **Deviations from spec:**
  - None in the five literals — all diffed byte-identical against the spec
    text before and after the mutation check.
  - **One narrative claim in Fork 3 does not hold and is reported, not
    fixed, per the transcribe-verbatim-and-report rule.** The spec says the
    call-to-action goes "to the third of ten `##` headings." Measured: the
    pre-change README (`git show main:README.md`) has exactly **10** `##`
    headings, so post-change, with the new section added, there are **11**
    total, and the call-to-action is the **3rd of 11**, not "3rd of 10." The
    literal artifacts (①–⑤) are unaffected — this is editorial prose in
    Fork 3's argument, not a transcribed literal or a tested assertion — but
    it's a factual slip worth naming since Fork 3 leans on the "third of
    ten" framing to argue the outline win.
- **Follow-up work identified:**
  - None beyond what's already recorded in the spec's own "Out of scope"
    section (`internal/cli/window.go:12-13` V2, and switching C2/D2/J2/T2/X7
    to word bands per their stated trigger).

### Build-phase reflection (3 questions, short answers)

Process-focused: how did the build go? What friction did the spec create?

1. **What was unclear in the spec that slowed you down?**
   — Nothing. The five literals, the anchor line ranges, and the ordered
   sequence (②③④ → `bash -n` → ① → `just test-docs` → ⑤ → `just test-docs`)
   were all specific enough to follow mechanically — every anchor matched
   exactly once, every diff came back clean on the first attempt.

2. **Was there a constraint or decision that should have been listed but wasn't?**
   — Not a constraint, but one useful thing to confirm empirically rather
   than assume: the "177 distinct assertion ids" claim is about
   `inventory.sh`'s deduped extraction from source, not a count of runtime
   `OK:` lines from `just test-docs` (which is 178, because some ids can
   legitimately emit more than one `OK:` at runtime). A build session that
   grep-counted `OK:` lines instead of running `inventory.sh`'s own
   extraction would have seen a mismatch and wrongly suspected a defect.

3. **If you did this task again, what would you do differently?**
   — Nothing procedurally. The one thing worth carrying forward: the Fork 3
   "third of ten" miscount was only caught by re-deriving the pre-change
   heading count from `git show main:README.md` instead of trusting the
   spec's arithmetic — which is exactly the re-measurement discipline this
   spec itself modeled against framing. Good habit to keep applying even
   inside a "transcribe verbatim" build.

---

## Verify Findings

*Filled in during the **verify** cycle (fresh session, per AGENTS.md §6).*

**Verdict: ✅ APPROVED.** All nine acceptance criteria hold, all four gates are
green, and the mechanical half is clean end to end: literals ①②③④ diff
byte-identical against this document, literal ⑤ round-trips against
`./scripts/inventory.sh` with both sides non-empty (21 lines, 1543 bytes,
matching md5), and the code-only diff of `scripts/test-docs.sh` against `main`
— every comment line stripped — is **exactly three changes** and nothing else:
the new helper, `A1`'s call line, and `A11`. Eight mutants were re-run from
scratch in this session; every one was hash-verified as a real change to the
file before the harness was run, and the tree was confirmed clean by
`git status --porcelain` after each revert.

Three findings, all **recorded rather than blocking**. None touches a literal,
an assertion, a derived count, or a gate; all three are editorial claims in
prose, two of which now ship inside harness comments. Nothing was fixed in
place — the design sections stay as design wrote them, and the corrections live
here.

### O1 — "3rd of 11 `##` headings" is true under no single reading

Fork 3 argues the outline win with *"the third of ten `##` headings."* Build
challenged the denominator, corrected it to 11, and reported **"3rd of 11"** —
which fixes one half and leaves the other wrong. Measured here with a
fence-aware parser (`grep -c '^# '` returns **2** on this README, because the
bash comment at `README.md:87` inside a fenced block matches it; `grep -c '^## '`
happens to return the right answer, 11, only because no `## ` line falls inside
a fence):

| Unit | Post-change total | CTA's position |
|---|---:|---|
| `##` headings only | 11 | **2nd** |
| headings counting the `# Bragfile` H1 | 12 | **3rd** |
| all headings as GitHub's outline renders them (H1 + H2 + the one H3) | 13 | **4th** |

"3rd of 11" pairs the ordinal from the second row with the denominator from the
first. The accurate statements are *2nd of 11 `##` headings*, *3rd of 12
counting the H1*, or *4th of 13 in the rendered outline*. Design's original
"3rd of ten" was wrong on both halves under the `##`-only reading it named.

**Not blocking, and not a punch list.** No assertion, literal, or derived count
depends on the figure; `A11` pins the heading by position-in-file, not by
ordinal. The correct number also makes Fork 3's argument *stronger* rather than
weaker — 2nd of 11 sits higher in the outline than 3rd of 11 — so nothing that
was decided on this sentence would have been decided differently.

**Stage-level note, since approving this closes STAGE-021.** This is the third
recorded instance in this stage of the same mechanical failure: an `X of N`
claim re-derived only on the half that was challenged. SPEC-079's verify
correction — *"six of 75 carry none"* → *"six lack the heading, of which five
carry none"* — is recorded in the stage file as the same defect class, and the
open-questions item records a third family of it (`8 of 18` with a wrong total,
then `9 of 18` wrong on both halves). Under §12's own codification meta-rule
that is N=3 same-outcome on one mechanical sub-rule, which clears the bar:
**re-derive both halves of an `X of N` claim, not only the challenged half.**
Recommended for codification at stage ship, not asserted here.

### O2 — "the smallest deep-dive doc in this repo" is scoped wider than its evidence

Fork 1's second leg, now shipped verbatim in `A1`'s comment, reads: *"The
smallest deep-dive doc in this repo (`docs/for-ai-agents.md`, 2120 words on
2026-08-20) sits well clear of it"*, and the spec draws a *"~700-word gap"* from
it. Measured across every long-form document in the tree:

| doc | words |
|---|---:|
| `docs/macos-notarization-checklist.md` | 1447 |
| `docs/architecture.md` | 1577 |
| `docs/data-model.md` | 1908 |
| `docs/for-ai-agents.md` | 2120 |

Two deep-dive documents are smaller than the one named. The claim is exact
under the reading Fork 1's tier list actually uses — *the docs the README routes
to* (`BRAG.md`, `docs/api-contract.md`, `docs/engineering-practices.md`,
`docs/for-ai-agents.md`, `docs/tutorial.md`) — where `for-ai-agents.md` at 2120
genuinely is the smallest; it is the phrase "in this repo" that over-reaches.

**The conclusion survives; the margin does not.** A ceiling of 1400 still sits
below *every* long-form doc in the repo (1400 < 1447 < 1577 < 1908 < 2120), so
the tier the leg exists to preserve is preserved. What is not true is the size
of the cushion: the nearest long-form doc is 47 words away, and the nearest
unambiguous deep-dive is 177 — not ~700. Recorded, not re-litigated: this does
not reopen `900 1400`, which the invoking brief placed out of scope and which
the narrowing argument (780 → 500, verified) settles independently.

### O3 — `A11`'s "A1 above bounds that length" no longer describes what A1 measures

`A11`'s comment closes: *"The threshold is derived from the file's own length
rather than a pinned line number, so it cannot rot, and A1 above bounds that
length so the two guards compose."* `A11`'s threshold is
`$(wc -l < README.md) / 3`. As of this spec `A1` measures **words**, so it does
not bound that quantity — and `scripts/test-docs.sh:275` (`a11_total`) is now
the only place in the harness that reads `README.md`'s line count at all.
Nothing bounds it.

**Checked before recording: the guard does not depend on the claim.** Inserting
`k` lines above the heading moves the heading to `74 + k` and the threshold to
`(262 + k) / 3`; `A11` fails whenever `74 + k > (262 + k) / 3`, i.e. for any
`k > 20`, with no help from `A1`. Insertions *below* the heading raise the
threshold without moving the heading, which only loosens a guard that is
already satisfied. So the composition sentence is decorative rather than
load-bearing, and reads as true only if "length" is taken loosely as "size".
Worth a word-swap whenever `A11`'s comment is next touched; not worth a return
trip.

### What was checked, and what it showed

**Gates.** `just test` (all packages ok), `just test-docs` (**ALL OK**),
`gofmt -l .` (empty), `go vet ./...` (clean), plus `bash -n scripts/test-docs.sh`
(clean). `just test-docs` emits **178 `OK:` lines against 177 distinct ids**;
`S3` is the sole duplicate, confirmed by `uniq -d` — the documented pre-existing
wart, not this spec's business. `./scripts/inventory.sh` independently prints
177.

**A1 passes for the right reason — six mutants, each proven real.** Every
mutation below was applied, its md5 confirmed changed against the pre-mutation
hash (a no-op `sed` would have been caught, not assumed), the harness run, then
reverted with `git status --porcelain` confirmed empty:

| Mutant | Observed |
|---|---|
| `wc -w` → `wc -l` in `assert_word_count_band` (line 85) | `FAIL: A1: README.md has 262 words (expected 900..1400)` — red on the **floor**, and the only failure. Reproduces the build's report and §12(b) run 4's last row exactly. |
| band → `900 1276` (ceiling one below actual) | `FAIL: A1: README.md has 1277 words (expected 900..1276)` |
| band → `900 1277` (ceiling at actual) | ALL OK — ceiling inclusive |
| band → `1278 1400` (floor one above actual) | `FAIL: A1: README.md has 1277 words (expected 1278..1400)` |
| band → `1277 1400` (floor at actual) | ALL OK — floor inclusive |

The band mutants matter beyond inclusivity: `A1` reports **1277**, the word
count, not 262, the line count, so it is evaluating the metric it claims to.
The spec's own four README-side edge mutants were reproduced as well — padded
to 1401 → `FAIL … has 1401 words`; trimmed to exactly 1400 → green; truncated
to 899 → `FAIL … has 899 words`; exactly 900 → green — all four matching §12(b)
run 4's recorded messages verbatim.

**`A11` fails when what it pins is broken — three mutants, each proven real.**
Heading line deleted → `FAIL: A11: first '## …agent…' heading is at line 220 of
261 (must be within the first third, i.e. line 87)`. Whole section relocated to
just above `## License` → `FAIL: A11: … at line 214 of 262 …`. `brag mcp
install` replaced with `brag setup` on line 76 only → `FAIL: A11: the agent
call-to-action section must name 'brag mcp install' within 8 lines of its
heading`. Tree clean after each.

**The other five callers are untouched, and the no-guard-needed argument holds
empirically.** `assert_line_count_band`'s body diffs byte-identical against
`main` (matching md5, 13 lines), and its call sites are exactly `C2` (:348),
`D2` (:369), `J2` (:665), `T2` (:1104), `X7` (:1487) — five, with the sixth
(`A1`) migrated and the only other hit on `main` being the new helper's own
comment. Fork 2(a)'s claim was not taken on the design's word: mutating
`assert_line_count_band` itself to `wc -w` — the repurpose the sibling exists to
avoid — sends **exactly those five red and nothing else**:

```
FAIL: C2: CONTRIBUTING.md has 253 lines (expected 30..120)
FAIL: D2: docs/development.md has 474 lines (expected 50..200)
FAIL: J2: examples/brag-slash-command.md has 77 lines (expected 5..30)
FAIL: T2: docs/for-ai-agents.md has 2120 lines (expected 120..500)
FAIL: X7: docs/engineering-practices.md has 2245 lines (expected 150..300)
FAILED: 5 assertion(s) failed.
```

Each file's word count reproduces the design's table (253 / 474 / 77 / 2120 /
2245) and each falls outside its own line band, so the five callers really are
the test, and the decision not to add a guard is sound.

**The reflow demonstration reproduces.** Re-implemented independently — greedy
rewrap of unprefixed prose paragraphs only, fences and blockquotes untouched,
`new.split() == original.split()` asserted at every width — against
`git show main:README.md`: **248 / 250 / 254 / 259 / 263 / 265 / 267** lines at
widths 100 / 92 / 84 / 76 / 72 / 68 / 64, with `wc -w` pinned at **1268** and
the token stream identical every time. That is the spec's table line for line,
and A1's `+7` swing against zero headroom. `docs/engineering-practices.md` came
out at 270…295 (swing +12) against the design's 271…298 (+15) — a difference in
which lines each implementation treats as reflowable prose, not a disagreement:
X7's headroom is 17 either way, so Fork 2(c)'s written trigger ("switch a caller
when its swing exceeds its headroom") is not yet met under either measurement.

**`W3` and the blockquote.** `W3` is green (`OK:   W3`); `CHANGELOG.md`'s newest
dated heading is `0.6.1` and `README.md:10` claims `v0.6.1`. `README.md:1-14`
diff byte-identical against `main`, so the guarded line is untouched and the
blockquote ends at line 14 on `brew install jysf/tap/bragfile`, with no orphan
`>` separator. Read back with the CTA gone, it is a self-contained release-state
claim — *what v0.6.1 contains*, then how to install it — with no forward
pointer, no "see below", and no demonstrative left dangling: the removed
paragraph began a new thought (*"Working with an AI agent?"*) rather than
completing one. The blockquote loses its only link to the MCP section, which is
the trade Fork 3 states and prices; the link now lives in the new section at
`README.md:79`. It resolves — `## Using brag from an AI agent (MCP)` at
`README.md:221` slugs to `using-brag-from-an-ai-agent-mcp` — though note `E1`
does **not** check it: `check_link_target` strips `#…` and returns early on an
empty target, so pure-anchor links are unverified by the harness. Pre-existing,
identical on `main`, and out of scope here.

**Reported numbers all reproduce.** README 260 → 262 lines and 1268 → 1277
words; the removed block is 41 words and the new section 50, for the predicted
+9; `A11` is the only new id (`comm` against `main`'s extraction); the practices
page is unchanged at 283 lines with a one-number diff. `X3` carries an explicit
`[ -z "$x3_got" ]` guard, so an empty block fails rather than passing silently —
the failure mode worth ruling out on a derived literal.

**The promoted copy's factual claims hold against the code, not just the DECs.**
`internal/mcpserver/server.go` makes exactly **five** `mcp.AddTool` calls —
`brag_add`, `brag_list`, `brag_search`, `brag_stats`, `brag_memory` (:38–:58),
five in the package as a whole — matching "five typed tools". "An auto-loadable
memory resource" is `brag://memory/recent`, registered at
`internal/mcpserver/resources.go:29` with the description *"Load this before you
start work"*; `brag mcp install` is a real command
(`internal/cli/mcp_install_test.go`). The claim sentence is also byte-unchanged
from the copy that already stood on `main` — only the pointer sentence was
rewritten.

**Untouched-claim spot-check (seven sampled, all held).** *"There are no golden
files"* — `find . -name '*.golden'` and `find -type d -name testdata` both
return zero. `TestProvenanceClassifier_GoPredicateMatchesSQLClause` exists at
`internal/storage/provenance_agreement_test.go:55`.
`TestPackageReadsNoWallClock` exists in both files cited, and
`TestPackageEmitsNoReservedTagNamespace` in `internal/memory`.
`scripts/archive-spec.sh:33-37` really does reject `<answer>` placeholders.
*"Every file under `projects/*/specs/done/` ends with a ship-phase
`## Reflection (Ship)`"* — 77 of 77 carry the heading, zero missing. The
README's *"five tool schemas"* pointer matches the five registrations, and
`## How this repo is built`'s 82 words match the figure `A1`'s comment cites for
it. No claim in this sample was contradicted by its own citation.

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
   if not.

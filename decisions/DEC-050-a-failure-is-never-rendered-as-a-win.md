---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-050                        # stable, never reused
  type: decision                     # decision | reservation
  confidence: 0.85                   # honest: the impact/wrapped posture is the
                                     # user's call and measured on the live
                                     # corpus (~0.92); row 4 is now the
                                     # maintainer's own ruling too, narrower than
                                     # the one this record first extrapolated
                                     # (~0.90, 2026-09-19); the zero-rename
                                     # accounting is measured, not argued
                                     # (~0.88); and the omit-when-empty rule has
                                     # a defensible opposite (~0.78), which is
                                     # now the softest part. See Validation.
  audience:
    - developer
    - agent

agent:
  id: claude-opus-5
  session_id: null

project:
  id: PROJ-008
repo:
  id: bragfile

created_at: 2026-09-18
supersedes: null
superseded_by: null

tags:
  - output-shape
  - dec-014-envelope
  - digest
  - failure
  - type
---

# DEC-050: a failure is never rendered as a win — the posture on all seven `--type` surfaces

## Decision

**No `brag` surface lists a recorded failure, unlabelled, among the entries it
presents as having gone well.** Where a surface does that today, the failure
moves into a section of its own, headed **`## What didn't work`**, in the same
per-entry shape the surface already uses. A surface that already says an
entry's type, or never lists entries one by one, changes nothing.

Five rules make that concrete:

1. **A failure is `aggregate.IsFailure(e)`, which is `e.Type == "failed"`,
   exactly.** One Go constant, `aggregate.FailureType`, is written by
   `brag learn` and read by every digest. The match is case-sensitive and
   untrimmed: the comparison storage's `--type` filter already makes
   (`e.type = ?` on a BINARY-collated column). So a digest's failure section
   and `brag list --type failed`, DEC-049's retrieval path, select the same
   rows. `Failed`, `failure` and `learned` are not failures.
2. **The section is a partition, not a filter.** The rows that leave a
   celebratory section are exactly the rows that land in `## What didn't
   work`. The with-impact subset is split in two by `aggregate.SplitFailures`,
   never narrowed, and `aggregate.WithImpact` keeps meaning *non-empty
   impact*.
3. **No existing count changes what it counts** (DEC-048). Every headline and
   every count key keeps its definition. A count that spanned the rows now
   split still spans both sections.
4. **The section renders only when it has an entry.** That rule is for
   markdown. JSON always carries the section's key, as `[]` when it is empty
   (DEC-014 part 4). A period with no recorded failure grows no empty heading.
5. **One name everywhere.** The markdown heading is `## What didn't work` on
   every surface that has the section. The JSON key is `failures_by_project`
   on every surface that groups the section by project.

The seven surfaces that take `--type` as a filter, each with its posture:

| # | Surface | What it did with a failure, measured | Posture | Implemented by |
|---|---|---|---|---|
| 1 | `brag impact` | Listed under `## Impact` as `- <id>: <title>` plus its impact. No type in either format: the JSON entry is DEC-028's 4-key `{id, title, project, impact}`. | **Section.** The failures leave `## Impact` for `## What didn't work`; JSON `impact_by_project` loses them and `failures_by_project` carries them. `Entries: <shown>/<in-window> with impact`, `entries_with_impact` and `counts_by_project` all still count both sections. | SPEC-086, plus the Amendment to DEC-028 |
| 2 | `brag wrapped` | Identical rows under `## Impact moments`. `Top types` is a top-3, so `failed` never surfaced there. | **Section**, between Impact moments and Rhythm. JSON `impact_moments` loses them and `failures_by_project` carries them. `Entries: N` and `total_entries` are unchanged. | SPEC-086, plus the Amendment to DEC-030 |
| 3 | `brag summary` | `By type` printed `failed: 4` and JSON `counts_by_type` carried it. `## Highlights` listed all four as `- <id>: <title>`, and its JSON highlight is a 2-key `{id, title}`. | **Section**, pulled out of `## Highlights`, with the same heading, the same omission rule and, because highlights group by project, the same `failures_by_project` key. `By type` is already honest and stays. | SPEC-094 |
| 4 | `brag story` | All four bundled profiles rendered each failure as `- ★ <id>: <title>`, with no type in markdown. JSON beats carry `"type": "failed"`. Two profiles are `candor: promotional` (`exec`, `skip`), and their directives tell a model to promote the beats. | **Never rendered as a win, and never silently dropped.** Those are the two invariants, and they are all this record locks. A **candid** profile (`manager`, `me`) labels its failures where it lists them. A **promotional** profile (`candor: promotional`, today `exec` and `skip`) **may omit them, but only with a visible note in its output** — a count of the omitted failures is the obvious form. **SPEC-094 designs the mechanism**: where the label goes, what the note says and where it sits, and whether a failure still counts as an impact beat. Each turns `Candor` from LLM-facing metadata into a body rule, and DEC-029 choice 2 makes profiles data. | SPEC-094 |
| 5 | `brag export` | The per-entry table carried a `type` row reading `failed`, and `**By type**` listed `failed: 4`. JSON is the full 9-key entry. | **Nothing.** It already labels every entry. | — |
| 6 | `brag coverage` | Provenance counts only. It never lists an entry. | **Nothing.** | — |
| 7 | `brag list` | Plain output is `id`, `created_at`, `title`. `--format tsv` and `json` carry `type`. | **Nothing.** A raw index makes no claim about its rows, and `--type failed` is the sanctioned way to ask for failures. | — |

The measurements are from a frozen copy of the live corpus taken on
2026-09-18: 606 entries, four of them `type: failed`, all four carrying an
impact, all four in 2026.

## Context

DEC-049 gave the corpus a way to hold work that did not work. It is a
reserved `type` value, `failed`, written by `brag learn`. It deliberately
did not decide what the digests do with one. The user decided that at
SPEC-085 design, for the two surfaces built to be shared (`impact` and
`wrapped`). Failures get their own named section, pulled out of the impact
body. They are not hidden and not mixed in. Both alternatives were
rejected: excluding failures by default, and including them silently as
the status quo did.

Three facts made this more than a renderer tweak:

- **Neither `impact` nor `wrapped` renders `type`, in either format.** A
  failure there was not merely unflagged. It could not be told apart from a
  win without a renderer change.
- **The risk was already realised.** Entry 420, typed `failed` and carrying
  an impact, reached the corpus the day SPEC-085 shipped. It was written by
  an agent following `BRAG.md`'s documented path. By 2026-09-18 there were
  four, and all four rendered as wins on four surfaces.
- **"Celebratory" is not a property of a command.** STAGE-023's scope guard
  asks for one consistent answer across `impact`, `wrapped`, `summary` and
  `story`. Framing had split the seven surfaces into *celebratory* and
  *neutral* and measured two of them wrong: `summary`'s `## Highlights` and
  every `story` profile list a failure as a win. On `story`, harm varies by
  profile, since two profiles are promotional and two are candid.

The renderer work lands in two specs. SPEC-086 covers rows 1 and 2, and
SPEC-094 covers rows 3 and 4. The split is along the measured defect shape.
On `impact` and `wrapped` a failure cannot be represented in either format.
On `summary` and `story` the data is present and only the per-entry
markdown drops it. This one record states the posture for all seven, which
is what makes the answer consistent while the code ships twice. DEC-049 set
the precedent: written once, binding forward.

## Alternatives Considered

- **Exclude failures by default, and reach them only behind a flag.**
  Rejected by the user: it re-creates, inside the digest people actually
  read, the flattery PROJ-008 exists to remove. On `story` the maintainer
  ruled narrower (row 4, 2026-09-19): a promotional profile may omit its
  failures, but not silently — the omission has to show in the output. What
  is rejected there is the *unannounced* exclusion, not every exclusion.
- **Include them silently.** This was the status quo and cost no code. The
  user rejected it, and STAGE-023's criterion that *"the celebratory digests
  do not silently absorb failures"* rules it out independently.
- **Decide per command: celebratory surfaces get a section, neutral ones
  nothing.** Rejected on measurement. `summary` and `story` were filed as
  neutral and both list a failure as a win. On `story` the answer depends on
  the profile, which a per-command rule cannot express.
- **Narrow `aggregate.WithImpact` to drop failures.** This one-line change
  is sufficient for selection, and was framing's mutation M-1. Rejected
  because it makes a true headline false: `Entries: 531/606 with impact`
  would read `527/606 with impact` while four entries that do carry an
  impact sit in the next section. It would also change what a function
  named `WithImpact` returns. The partition (rule 2) leaves both alone.
- **A third number in `impact`'s headline**, such as
  `Entries: 531/606 with impact (4 failed)`. Rejected on three grounds.
  First, it reopens DEC-048's sentence naming the two-number form as *the
  sanctioned form for reporting a pair*. Second, it moves seven prose sites
  and, by framing's measurement, nine tests across two packages. Third, it
  carries information the section already shows.
- **A per-section count line under each heading.** Rejected on DEC-048's
  own Alternative 2: it adds a provenance line to an envelope that never
  legislated one. Each JSON section array already has a length.
- **Redefine `counts_by_project` to count only `impact_by_project`.**
  Rejected because it is a silent redefinition. DEC-028 defines the map over
  the with-impact subset, and narrowing it without renaming it is exactly
  what DEC-048 forbids. It keeps its definition, so it still sums to
  `entries_with_impact`.
- **Always render the heading,** as `wrapped` already does for an empty
  `## Impact moments`. Rejected: every user with a clean quarter would grow
  a permanent empty heading that reads as an accusation or a template
  artifact. The JSON key is always present, so a program can still tell
  *"none"* from *"not reported"*.
- **Render the heading with a *"None recorded."* line.** Rejected for the
  same noise, and because the line would claim more than the section
  knows: a failure with no impact is not listed there (rule 2).
- **List every failure in `wrapped`'s section, with or without an impact.**
  Rejected for three reasons. The section would differ from `impact`'s for
  the same period. It would need a new per-entry shape with no impact line.
  And on `impact` it would break DEC-028's impact-first body and its
  headline. A failure without an impact is treated like any other entry
  without one: it is counted and not listed. `brag list --type failed` is
  the complete list.
- **Name `wrapped`'s key after its section (`what_didnt_work`).** That
  matches `wrapped`'s convention (`impact_moments`, `top_tags`). Rejected
  because the same data would carry two names on two surfaces, and one
  `jq .failures_by_project` would stop working on both.
- **Add `type` to the 4-key projection instead of adding a key.** Rejected:
  every consumer that reads `impact_by_project` without knowing to filter
  would still list the failure as a win. The key-level split makes it
  impossible to miss.
- **Widen `IsFailure` to catch `Failed` or `failure`.** Rejected. DEC-049
  part 3 keeps `type` unvalidated with exactly one reserved value. A wider
  predicate would also make a digest list rows that `brag list --type
  failed` cannot return.

## Consequences

- **Positive:** On the two surfaces built to be shared, a dead end no longer
  reads as a win. On the measured corpus, `impact --year` and `wrapped` each
  list 527 entries under their impact section and 4 under `## What didn't
  work`, where every other count is unchanged. The two surfaces render each
  section byte-identically, as they already did for the impact rows.
- **Positive:** Adding the section costs no count rename, so every
  `Entries:` line and count key means what it meant yesterday.
- **Positive:** The failure predicate has one definition, one constant and a
  drift guard, `TestFailureClassifier_GoPredicateMatchesTypeFilter`, which
  holds it to the retrieval path. `aggregate.IsAgentAuthored` set the
  precedent.
- **Negative — a breaking JSON change on two envelopes.** A consumer that
  read every with-impact entry from `impact_by_project` or `impact_moments`
  now has to read `failures_by_project` too. It is named in the CHANGELOG.
  The repo is pre-1.0 and has no known external JSON consumer, which is
  DEC-048 T4's premise. If that premise expires, the next change of this
  kind needs a deprecation path.
- **Negative — a failure without an impact is listed on neither digest.**
  That is the same treatment an impact-less win gets. `BRAG.md` already tells
  an agent to fill in `impact` on a failure, and `brag list --type failed`
  still returns all of them. **Accepted by the maintainer on 2026-09-19 for
  v0.7.0**, on the symmetry: the digests are impact-first (DEC-028 choice 3),
  and a failure is not made an exception to that in either direction. It is
  still counted in the headline totals, exactly as an impact-less win is. The
  cost is bounded by the corpus — **0 of the 4 recorded failures lack an
  impact** — and T5 is what re-opens it. No code or test changes: the
  behaviour is already pinned.
- **Negative — until SPEC-094 ships, rows 3 and 4 still list a failure as a
  win.** This record decides their posture now, so SPEC-094 implements a
  decision rather than making one. The gap is real for as long as it is
  open.
- **Neutral:** `wrapped`'s `Top types` can now show `failed`. Its rule did
  not change: it always counted every type, and a period heavy with failures
  says so.
- **Neutral:** `aggregate.FailureType` replaces `cli.FailureType`. The
  constant moved to the package that both its writer and its readers
  import. `internal/storage` still does not know `failed` is special.

## Validation

**Right if:**

- `brag impact` and `brag wrapped` list a `brag learn` entry under
  `## What didn't work` and not in the impact section, on one real store.
  Tested by `TestLearnCmd_ImpactSectionsWhatItWrote` and
  `TestLearnCmd_WrappedSectionsWhatItWrote`.
- `aggregate.IsFailure` and `ListFilter{Type: aggregate.FailureType}`
  select the same rows over every near-miss spelling. Tested by
  `TestFailureClassifier_GoPredicateMatchesTypeFilter` and
  `TestIsFailure_ExactMatchOnTheReservedValue`.
- `counts_by_project` sums to `entries_with_impact` and spans both
  sections. Tested by `TestToImpactJSON_CountsByProjectSpansBothSections`.
- A period with no failure renders no `## What didn't work`, and a
  failures-only window renders no bare `## Impact`. Tested by
  `TestToImpactMarkdown_SectionsRenderOnlyWhenNonEmpty` and
  `TestToWrappedMarkdown_WhatDidntWorkRendersOnlyWhenNonEmpty`.
- SPEC-094's summary and story work lands without revisiting this table.

**Revisit if:**

- **T1** — a user reads a clean period's missing section as *"this digest
  does not show failures"*. Then render the heading with an explanation,
  and re-cost the noise.
- **T2** — a second reserved value arrives (say `abandoned`). Then
  `IsFailure` matches a set, and the drift guard's seeds grow with it.
- **T3** — a promotional `story` bundle's visible note does not survive the
  model that consumes it, so the omission is silent again by the time anyone
  reads it. A note that a downstream synthesizer drops satisfies row 4's
  letter and not its point. Then the omission itself goes back to the
  maintainer, with the measurement that showed it.
- **T4** — an external consumer of either JSON envelope appears
  (DEC-048 T4).
- **T5** — an impact-less `failed` entry appears in the corpus. 0 of 4 carry
  that shape on 2026-09-19, which is why the acceptance above was cheap to
  give. The first one is a row that is counted, invisible on both digests,
  and reachable only by `brag list --type failed` — so re-cost listing every
  failure in the section then, against the per-entry shape and the headline
  that rejected it (see the *"List every failure in `wrapped`'s section"*
  alternative).

**Confidence: 0.85**, raised from 0.80 on 2026-09-19. Rows 1 and 2 carry the
user's own decision, measured on the live corpus (~0.92). Row 4 was the
softest part of this record at ~0.70, because it extended their reasoning to
a surface they had not ruled on. **They have now ruled on it** — narrower
than this record first wrote it — so it is maintainer-ruled rather than
extrapolated (~0.90). What is left there is not the posture but the
mechanism, which belongs to SPEC-094, and T3 is now about whether a visible
note actually stays visible. Rule 3's zero-rename result was measured on the
tree rather than argued (~0.88). **Rule 4 is now the softest part (~0.78)**:
omitting an empty heading has a defensible opposite, and T1 is its trigger.

## Amendment (2026-09-22, SPEC-094 design)

**Rows 3 and 4 have separate implementers now, row 4 has its mechanism, and
T3 has fired and closed.** Everything above this heading is left as written,
so the amendment can be read against it.

- **Row 3 (`brag summary`) is implemented by SPEC-095, not SPEC-094.**
  SPEC-094's framing (2026-09-22) found that the premise shared by this
  record's Context and row 3 is false for `summary`: its JSON highlight is
  `{id, title}` with no `type`. On `summary` a failure cannot be told from a
  win per entry in either format, which is the *unrepresentable* shape rows
  1 and 2 fixed, not the lossy-markdown shape `story` has. The posture in
  row 3 is unchanged. The *"until SPEC-094 ships"* consequence now reads
  *until SPEC-094 and SPEC-095 ship*.
- **Row 4's mechanism is
  [DEC-054](DEC-054-story-candor-decides-what-a-failure-renders-as.md).** A
  profile whose `candor` is exactly `promotional` omits its failures, with
  an `Omitted:` line, an always-present `omitted_failure_count`, and a fixed
  clause appended to the framing directive. Every other profile labels them
  `✗ <id> (failed)`. A failure with an impact is still an impact beat, so
  rule 3 holds on `story` with no rename.
- **T3 fired for the form this record named, and is closed.** A bare count
  survived the consuming model in 0 of 10 runs. The maintainer allowed a
  binary-authored clause on 2026-09-22. With it, the omission was stated in
  14 of 14 runs at framing, and in 25 of 25 at design on the wording DEC-054
  locks. DEC-054's T1 is the trigger that would reopen it.

## References

- **DEC-014:** the envelope. Part 4's empty-state rule governs the empty
  *document*, and this record's rule 4 governs the empty *section*.
- **DEC-028:** `impact`'s two-number headline, its 4-key projection, and
  the impact-first body. All kept, and the record **amended** for the new
  key — `## Amendment (2026-09-19, SPEC-086 design)`, added on the
  maintainer's ruling. Its choice 5 key list predates `failures_by_project`,
  and its choice 3 body is now two sections. The original text is untouched,
  as it is on DEC-030.
- **DEC-029:** story profiles are data. That is why row 4's mechanics
  belong to SPEC-094.
- **DEC-030:** `wrapped`'s section arc, amended for the new section.
- **DEC-048:** a count names what it counted. It is rule 3's authority, and
  its Alternative 2 is why there is no per-section count line.
- **DEC-049:** the reserved value and its verb. Rule 1 holds the digests to
  its retrieval path.
- **SPEC-085:** the user's decision, taken at its design.
- **SPEC-086:** this record's spec. It implements rows 1 and 2.
- **SPEC-094:** implements rows 3 and 4.

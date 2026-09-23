---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-054                        # stable, never reused
  type: decision                     # decision | reservation
  confidence: 0.80                   # honest: the posture is DEC-050 row 4 and
                                     # the maintainer's own yes (2026-09-22), so
                                     # what this record adds is mechanism. The
                                     # count accounting is measured on the frozen
                                     # corpus in 12 cells (~0.9). The clause's
                                     # survival is measured, 25 of 25, but on two
                                     # models, two profiles and one quarter
                                     # (~0.75). The candid label was not measured
                                     # against a model at all (~0.7): that is
                                     # verify's, and T4 below is its trigger.
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

created_at: 2026-09-22
supersedes: null
superseded_by: null

tags:
  - story
  - failure
  - output-shape
  - dec-014-envelope
---

# DEC-054: `brag story`'s candor decides what a failure renders as

## Decision

**A `brag story` profile whose `candor` is exactly `promotional` leaves
recorded failures out of the bundle and says so three ways. Every other
profile lists them where they fall, labelled as failures. No profile marks
one as a win.** This is the mechanism for
[DEC-050](DEC-050-a-failure-is-never-rendered-as-a-win.md) row 4. That row
fixed two invariants: a failure is never rendered as a win, and it is never
dropped silently. The maintainer then ruled on the one open question on
2026-09-22.

Four parts:

1. **`Candor` is a body rule, on exactly one value.** A failure is
   `aggregate.IsFailure(e)` (DEC-050 rule 1). When a profile's `candor` is
   the exact string `promotional` (`story.CandorPromotional`), its failures
   are removed from the in-window entries before threads are built. That
   order matters: `impact_threads_only` then folds a thread whose only
   impact beats were failures, and those failures are still counted as
   omitted. **Every other value labels them inline**, as
   `- ✗ <id> (failed): <title>` with the impact line when there is one. That
   covers `candid`, an empty value, an unknown value, and a case or spelling
   variant such as `Promotional` or `promo`. `Candor` is an unvalidated
   string in a user profile, so the default on a typo has to be the rule
   that never drops anything. Today `exec` and `skip` omit, and `me` and
   `manager` label.

2. **An omission is stated in the bundle, in the JSON, and to the model.**
   - Markdown: a line directly under `Beats:`,
     `Omitted: <n> recorded failures, not listed for this audience (brag list --type failed)`
     (`1 recorded failure` for one). It renders **only when `n > 0`**,
     which is DEC-050 rule 4's rule for a section, applied to a line.
   - JSON: `omitted_failure_count`, an integer that is **always present**,
     `0` when nothing was omitted, and placed between `filters` and
     `threads` (DEC-014 part 4). It is named for what it counts (DEC-048).
   - The model: when `n > 0`, the binary appends one fixed paragraph to the
     resolved framing directive, in both formats:
     `This bundle omits <n> recorded failures for this audience. End with one line that says so; do not drop it.`
     It is appended to **whatever the directive resolved to**, including a
     user's own `directive:` file. When the directive is empty, the clause
     becomes the whole directive. `--print-directive` reads no window, so it
     prints the directive as authored, without the clause.

   **The wording is locked by measurement, not by taste.** Changing it
   means re-running the measurement below.

3. **`★` means "an impact beat that is not a failure". `is_impact_beat`
   keeps meaning "carries an impact".** A failure with an impact is still an
   impact beat. It is counted in `is_impact_beat`, `impact_beat_count` and
   the throughline's `<m> with impact`, and only its marker changes. The
   predicate is `aggregate.HasImpact`, which is `WithImpact`'s rule for a
   single entry. `internal/story` calls it rather than restating it. **No
   count changes its definition, so nothing is renamed** (DEC-048). On a
   promotional profile, `Beats: <shown>/<in-window>` and every per-thread
   count still count the beats shown. There are simply fewer of them, and
   the `Omitted:` line accounts for the difference.

4. **DEC-050's T3 is closed by part 2.** T3 fired for the form DEC-050
   named, a bare count, which survived the consuming model in 0 of 10
   runs. The maintainer allowed a binary-authored clause. The locked
   wording was then measured, and the note was stated in 25 of 25 runs.

## Context

DEC-050 row 4 left `story` two questions. Both turned `Candor` from
metadata into a body rule, and DEC-029 choice 2 makes profiles data rather
than code. SPEC-094's framing measured them on 2026-09-22, on a frozen copy
of 621 entries, 4 of them `failed` and all 4 carrying an impact:

- **All four bundled profiles rendered every failure as `- ★ <id>:`.** The
  markdown carries no type. The JSON beats do.
- **Under a promotional directive, the failures never reached the prose,
  wherever they sat.** They were listed as `★` in 0 of 10 runs, labelled
  inline in 0 of 10, and moved to their own block in 0 of 10. The `exec`
  directive says *"No process, no messy middle"*, and the model obeys it. So
  a promotional profile was already dropping its failures silently, in the
  prose made from the bundle rather than in the bundle itself.
- **A visible note did not survive on its own** (0 of 10). With a line in
  the bundle body asking for it, it survived in 4 of 10 (Sonnet 4 of 5,
  Haiku 0 of 5). **With a clause appended to the directive, it survived in
  14 of 14.**
- **`Candor` was not surfaced to the LLM** at all, despite `profile.go`'s
  comment saying so. It appears in neither format. The field was parsed and
  dropped.

The maintainer was asked whether `brag story` may append a fixed,
binary-authored clause to the framing directive when a promotional profile
omits failures. They answered **yes, as framed**, on 2026-09-22.

**Design's measurement, on the wording this record locks** (SPEC-094,
§12(b) Finding 2). The bundles were produced by the prototype binary, not
doctored by hand. Each was piped through
`claude -p --model <m> --tools "" --no-session-persistence --setting-sources ""`
with the tutorial's prompt, *"weave these threads into one headline arc"*:

| Cell | Bundle | Runs | Omission stated, as the prose's last line |
|---|---|---:|---:|
| `exec`, scoped | `--audience exec --quarter --project bragfile`, 39/42 beats, 3 omitted | Haiku 5, Sonnet 5 | **10 / 10** |
| `exec`, full | `--audience exec --quarter`, 426/457 beats, 4 omitted, 221 KB | Sonnet 5 | **5 / 5** |
| `skip`, scoped | `--audience skip --quarter --project bragfile`, 39/42, 3 omitted | Haiku 5, Sonnet 5 | **10 / 10**, one of them calling the failures *"production/operational escapes"* |

No omitted failure's content reached any prose (0 of 25 on per-failure
keywords). The grader's regex was calibrated to 0 hits on the 30 framing
outputs made without an omission (V0, V3, V4) and on the unmodified input
bundle. Every hit was read by hand.

## Alternatives Considered

- **Label a promotional profile's failures inline**, as a candid profile
  does. Rejected on measurement: framing's V3 put them in the prose in 0 of
  10 runs. Under a promotional directive it is silent downstream, and it
  would still need the clause.
- **Move a promotional profile's failures to a `## What didn't work`
  block.** This is SPEC-086's shape. Rejected on measurement: V4 scored 0 of
  10, for the same reason.
- **A bare count note**, which DEC-050 called *"the obvious form"*.
  Rejected on measurement: V1 scored 0 of 10. It satisfies row 4's letter
  and not its point, which is exactly T3.
- **Put the carry instruction in the bundle body rather than in the
  directive.** Rejected on measurement: V5 scored 4 of 10, and 0 of 5 on
  Haiku.
- **Put the clause in the bundled `exec.md` and `skip.md` assets** instead
  of having the binary append it. Rejected: a user profile can point
  `directive:` at its own file, and that file would carry no clause. The
  clause has to come from whatever did the omitting.
- **Stop counting a failure as an impact beat.** Rejected. It changes what
  `is_impact_beat`, `impact_beat_count` and `<m> with impact` count, so
  DEC-048 would require renaming all three. It is also the answer DEC-050
  already rejected as M-1, because it makes a true headline false: a
  failure that carries an impact does carry one.
- **Key the rule on `candor != "candid"`.** Rejected: then a typo omits. A
  rule that drops output has to be the one reached only on purpose.
- **Validate `candor` at parse time, and reject unknown values.** Rejected:
  DEC-029 made the profile schema free-form data, and existing user
  profiles may carry other values. Rejecting them would break
  `brag story --audience <theirs>`. An exact match gives the same safety
  without breaking anything.
- **Render `Omitted: 0 …` on every promotional bundle.** Rejected, for the
  reason DEC-050 rule 4 gives against empty headings: the JSON key is always
  present, and markdown grows no line that reports nothing.
- **Name the key `failures_omitted`, or put it last.** `omitted_failure_count`
  matches the envelope's own `beat_count` and `impact_beat_count`. It sits
  where the markdown line does, among the provenance and before the body.

## Consequences

- **Positive:** no `story` profile renders a dead end as a win. A
  promotional bundle tells its reader, its program and its model what it
  left out.
- **Positive:** the impact-beat predicate has one definition. `thread.go`
  restated it twice. Because `WithImpact` was also `e.Impact != ""`, no
  output changed, but the one test that pinned it restated it too, so a
  change to `WithImpact` would have left `story` behind with the suite
  green.
- **Negative: for the first time, bragfile writes text that a model reads
  as an instruction.** It is one fixed sentence, it is appended only when
  something was omitted, and it is appended even to a directive the user
  wrote. The only way to turn it off is to stop omitting, by setting a
  profile's `candor` to anything other than `promotional`.
- **Negative: on a promotional profile, the shown counts and `exec`'s
  thread order move**, because they count the beats shown. Measured on the
  frozen corpus: `exec --year` `Beats:` goes from 546/621 to 542/621, and
  `skip --year` from 619/621 to 615/621. `exec --month`'s `bragfile` thread
  goes from 9 impact beats to 6, which drops it below `irradiance`. Nothing
  changed what it counts. The rows it counts are fewer.
- **Neutral:** `Candor` is still not rendered in either format. Its effect
  is visible, and the value is not.
- **Neutral:** candid output changes in exactly one way. The failure lines
  swap `★` for `✗ … (failed)`. On the frozen corpus, `me --year` and
  `manager --year` differ from `main` in 4 lines each, and their JSON is
  identical once `omitted_failure_count` is removed.

## Validation

**Right if:**

- A promotional profile omits exactly the rows `brag list --type failed`
  returns in its window, and a candid one labels them. Tested by
  `TestOmitFailures_OnlyExactPromotionalOmits`,
  `TestLoadProfile_OnlyExactPromotionalOmitsFailures` and
  `TestLearnCmd_StoryLabelsOrOmitsWhatItWrote`.
- A thread whose only impact beats are failures folds on `exec`, and those
  failures are still counted. Tested by `TestOmitFailures_RunsBeforeTheFold`.
- The note, the key and the clause agree in both formats, including when
  every beat was omitted. Tested by the two SPEC-094 goldens,
  `TestToStoryJSON_PromotionalCountsWhatItOmitted` and
  `TestToStory_EveryBeatOmittedStillSaysWhy`.
- `internal/story` restates neither predicate. Tested by
  `TestStoryPackage_CallsTheSharedPredicates`.

**Revisit if:**

- **T1:** a verify or later run of the SPEC-094 method finds a promotional
  bundle's omission missing from the model's prose. The clause is then no
  longer sufficient. Re-measure before changing its wording, and change it
  only with a new measurement.
- **T2:** someone wants a promotional voice that still lists failures. That
  is a second key, such as `failures: omit | label`, and not a third
  `candor` value.
- **T3:** a third `candor` value arrives. Part 1 then has to say which rule
  it takes, and the answer is never "omit" by default.
- **T4:** the candid label does not survive the candid directives
  (`me.md` asks for *"the messy middle"*, and `manager.md` for
  *"blockers and risks"*). That was not measured at design. It is verify's
  to measure with SPEC-094's method.
- **T5:** an external consumer of `brag story --format json` appears. The
  removal of failure beats from `exec` and `skip` is then a break to
  announce, not only a CHANGELOG line (DEC-048 T4, DEC-050 T4).

## Amendment (2026-09-23, SPEC-094 ship)

**T4 fired in part.** SPEC-094's verify measured it (V-F2), and the
maintainer ruled on 2026-09-23 that the result is recorded here and the
remedy gets its own spec. Everything above this heading is left as written,
so the amendment can be read against it.

**Method.** It was SPEC-094's method, pointed at the candid profiles. Bundles
were built by the binaries on the same frozen corpus copy (621 entries, 4
`failed`):
`story --audience <me|manager> --since 2026-09-01 --project bragfile`, which
is 9 beats, 3 of them failures (433, 465, 473). Each bundle was built twice,
on the SPEC-094 binary (`✗ <id> (failed)`) and on the pre-build binary (`★`,
the counterfactual). A no-failure control used `--project bragfile-site`. Each
run was `claude -p --model <haiku|sonnet> --tools "" --no-session-persistence --setting-sources ""`
with the tutorial's prompt, *"weave these threads into one headline arc"*.
There were **52 runs**: five per model per bundle, and three per model per
control. Two model graders each made false hits on the controls, so every
non-absent judgment was read by hand. *Credited as a win* means the prose
lists the failure among work shipped, fixed or landed, or presents it chiefly
as an accomplishment.

| Profile | Failures credited as wins, `★` (before) | …with `✗` (this record) |
|---|---:|---:|
| `manager` (Haiku + Sonnet) | **22 of 30** mentions | **4 of 30**, all #473 |
| `me` (Haiku + Sonnet) | 1 of 30 | 2 of 30, both #473 |

An id followed within 12 characters by `fail`/`failed`/`failure` appears 11
times in the 20 `✗` outputs and 0 times in the 20 `★` outputs and the 12
controls, so models do carry the label.

**What it establishes:**

- **On `manager`, part 1's label does the work.** With `★` and the
  directive's *"Lead with what shipped"*, the failures landed under
  **Shipped**. With `✗` they land under blockers, friction or risk in 26 of
  30 mentions.
- **On `me`, the label is inert.** The directive's *"messy middle"* and the
  entries' own titles already carry the candour.
- **The residue is one entry.** All six residual cases are #473, whose
  recorded impact narrates its own fix.
- **So DEC-050's invariant, *a failure is never rendered as a win*, holds in
  the bundle and not always in the model's prose**, when a failure's impact
  reads like a win. The binary cannot guarantee it downstream. Neither
  `me.md` nor `manager.md` says what `✗` means.

**Where it goes.** A line in `me.md` and `manager.md` telling the model what
`✗` means is **SPEC-096**, claimed by its file at SPEC-094's ship. It edits
shipped assets that SPEC-094's LD14 kept out of scope, and it must re-run
this 52-run measurement on the wording it locks. **It does not gate
v0.7.0.** T4 stays open until SPEC-096 ships or is closed NO-GO.

**Limits.** Two models, one scoped window, and three failures, all with an
impact, so an impact-less failure under `✗` is untested. The prompt is
`exec`-shaped, and a manager-shaped prompt was not measured. Two `✗` runs on
`me` also called #472, a win, a failure. That error is outside DEC-050, and
the `★` runs were not audited for it.

## References

- **DEC-050**: the posture. This record implements row 4 and closes T3.
- **DEC-029**: story profiles are data. Amended at SPEC-094 design for the
  new key, the rule and the clause. Its original text is left as written.
- **DEC-014**: the envelope. Part 4 is why `omitted_failure_count` is always
  present.
- **DEC-048**: a count names what it counted. It is part 3's authority.
- **DEC-049**: the reserved value, written by `brag learn`.
- **SPEC-094**: this record's spec, with both measurements in full.
- **SPEC-096**: the directive line T4's measurement asks for. Claimed at
  SPEC-094's ship and not yet framed.
- **SPEC-086**: the sibling. It implements rows 1 and 2, and its shape set
  this spec's shape.

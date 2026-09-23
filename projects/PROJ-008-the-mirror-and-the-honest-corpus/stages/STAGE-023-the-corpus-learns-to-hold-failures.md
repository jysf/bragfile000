---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-023                     # stable, zero-padded, repo-global (never reused)
  status: active                    # proposed | active | shipped | cancelled | on_hold
  priority: high                    # critical | high | medium | low
  target_complete: null

project:
  id: PROJ-008                      # parent project
repo:
  id: bragfile

created_at: 2026-09-05
shipped_at: null
---

# STAGE-023: the corpus learns to hold failures

> **Cycle: frame.** GO. This stage is the project's **first**, ahead of impact
> quality — reversing the brief's own stage plan. The ordering call and its
> rejected option are in *The ordering call* below. Every number on this page
> was re-derived on 2026-09-05 against `main` at `1775feb`, corpus **397**.

## What This Stage Is

The corpus can hold work that **did not work**, and an agent reading the
history gets it back. Today `brag` records 397 entries and **none of them is a
failure** — not because failures were suppressed, but because there is no way
to say one. This stage adds the capture verb, decides where a failure lives in
the schema, and makes it survive the read path that agents actually use
(`brag memory`). It stops when the corpus is *capable of being honest*; making
the digests reason about honesty is a later stage.

## Why Now

**The flattery problem is upstream of everything else in PROJ-008.** The brief
argues impact quality goes first — *one classifier, three surfaces* — and that
argument is sound about the mirror and story-surface v2, which both consume a
ranked impact. It is not sound about ordering against `brag learn`, for a
reason the brief could not see when it was written:

**The classifier's calibration data is the corpus itself.** An impact rule
designed against 397 wins is tuned on a population that excludes the case it
will most need to judge. A failure's impact statement reads differently from a
win's — *"cost two days and produced nothing reusable"* is quantified, negative,
and has no before/after — and a rule that never saw one will be re-opened the
first time it does. Building capture first means the classifier is designed
against a corpus that contains both shapes.

Two further weights, both checked rather than inherited:

1. **The promotion condition already fired.** The PROJ-005 synthesis parked
   `brag learn` with *"promote to (A) if the memory work lands."* It landed in
   v0.6.0.
2. **It is the only pillar that changes what the corpus can hold** rather than
   how it is read. Every surface built on a wins-only corpus flatters,
   including the classifier.

And it is the user's stated priority (2026-08-23, at activation).

### The ordering call

**Chosen: `brag learn` first (this stage), impact quality second
(STAGE-024).**

**Rejected: impact quality first, `brag learn` second** — the brief's own stage
plan. Rejected on two grounds, one of principle and one measured:

- *Principle.* The calibration argument above. Ordering the classifier first
  makes it a rule written against half its input domain.
- *Measured.* The brief's premise that impact quality is a **small** primitive
  — *"the highest-leverage primitive is small and blocks all three"* — **does
  not survive measurement.** See *The classifier is not small* below. It is not
  an S, so putting it first delays the whole project rather than unblocking it.

**What the rejection costs, named so it is a choice:** the mirror (stage 3) and
story-surface v2 (stage 4) both consume ranked impact, so impact quality must
land before *them*. This ordering does not change that — it inserts one stage
ahead of the classifier, not ahead of its consumers. The cost is one stage of
delay on `wrapped`'s highlight-reel problem, which has been true since v0.4.0
and is not degrading.

### The classifier is not small — measured 2026-09-05

The brief measured impact quality on 2026-08-15 across 368 entries and reported
**52 (14%)** carrying a number with a unit. Re-derived today across 397, the
answer depends almost entirely on how the rule is worded:

| Candidate rule | Entries | % of 397 |
|---|---:|---:|
| impact contains any digit | 230 | 58% |
| digit + `%` / `x` / `$` | 41 | 10% |
| digit + a unit noun (`ms`, `files`, `tests`, …) | 85 | 21% |
| an `N → N` before/after shape | 11 | 3% |

A twentyfold spread between the loosest and tightest reading, all of them
defensible in prose. Two specific failure modes make this worse than a
threshold-picking exercise:

- **False positives: 141 of the 230 digit-bearing impacts contain a digit that
  is an identifier, not a measurement** — `SPEC-083`, `v0.6.1`, `u16`, `MD5`,
  `#190`. A digit rule is wrong on roughly three fifths of what it selects.
- **False negatives: 27 impacts are quantified only in spelled-out numerals** —
  *"Five subject workspaces"*, *"Three distinct causes"*, *"the two seams"* —
  and are invisible to every digit rule. This is not incidental: **this repo's
  house style spells numerals out in prose**, which is why `root.go:13` carries
  an unguarded `"four"` as a routed comment item.

So "quantified vs. prose" is not a filter someone forgot to write. It is a
stated, tested rule with a documented false-positive posture — an **M**, and
plausibly an L if it must also be applied retroactively. That is a finding for
STAGE-024's framing, recorded here because it is what settled the ordering.

## Success Criteria

- A user can record work that did not work, in one command, without
  contorting it into a win.
- A failure entry is **retrievable as a failure** — some surface can answer
  *"what has not worked?"* rather than requiring the user to remember.
- `brag memory` returns failures to an agent reading the corpus cold, and the
  slice's line shape says which entries they are.
- The celebratory digests (`wrapped`, `impact`) do **not** silently absorb
  failures. Whatever they do is a decision written down, not an accident of
  whichever query happened to match.
- The DEC-014 envelope and DEC-048's provenance-count rule hold on every
  surface this stage touches.
- No LLM in the binary. No network. No migration unless a fork explicitly buys
  one.

## Scope

### In scope
- The capture verb — `brag learn` (or whatever SPEC-085 settles it to be), and
  its MCP counterpart if the fork lands that way.
- **Where a failure lives in the schema** — the `type` / reserved-tag / new-
  column fork, settled with evidence.
- **How failures participate in `brag memory`** — the DEC-043 ranker and the
  DEC-044 line shape.
- **What the digests do with failures** — decided and written down, even where
  the decision is *"nothing, and here is why."*
- The `guidance/questions.yaml` entry `memory-slice-fusion-constants`: check
  whether this stage answers it, and say so either way.

### Explicitly out of scope
- **The impact-quality classifier.** STAGE-024. This stage must not smuggle in
  a ranking rule.
- **The mirror** and **story-surface v2**. Stages 3 and 4.
- Retroactively reclassifying the 8 existing `type: learned` entries. Their
  status is a fork in SPEC-085; *acting* on it beyond a documented position is
  out of scope here.
- Capture completeness (staging inbox, git-import miner) — routed, see
  *Dependencies*.
- The read-path discoverability gap (`BRAG.md` is write-only) — routed, see
  *Dependencies*.

## Spec Backlog

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [x] SPEC-085 (shipped on 2026-09-05) — the capture verb, its schema home,
      and its behaviour in `brag memory`; five forks, all settled. Complexity
      **M**, re-affirmed after the split below. Shipped as `brag learn` +
      **DEC-049**; `pr:199`.
- [x] SPEC-086 (shipped on 2026-09-20) — **the celebratory digests stop
      rendering a failure as a win.** `brag impact` and `brag wrapped` gain a
      `## What didn't work` section, and it authors **DEC-050**, which states
      the posture for **all seven** `--type` surfaces, so the scope guard below
      is satisfied by the decision even though the renderer work lands in two
      PRs. Framing shrank the DEC-048 obligation from *"two count renames"* to
      *at most one*; design measured it at **zero**, because the with-impact
      subset is **split** between two sections rather than narrowed —
      `Entries: 538/613 with impact` is byte-identical on the `main` binary and
      this branch's, and `entries_with_impact` 538 = `sum(counts_by_project)`
      538 = 534 impact rows + 4 failure rows. Complexity **M** held by splitting
      the `summary`/`story` half out as **SPEC-094**. `DEC-030` and — on the
      maintainer's **R3** ruling, which overturned design's own NO CHANGE call —
      `DEC-028` each carry an `## Amendment` with their original text untouched
      (LD6), so the inventory's `## Amendment` row moved 2 → 3. **The JSON
      change is breaking:** failures leave `impact_by_project` and
      `impact_moments`, both envelopes gain `failures_by_project` (always
      present, `[]` when empty, DEC-014 part 4), and `counts_by_project` still
      spans both sections — guarded from both directions by `M-D` and `N-6`.
      **Verify returned a punch list of three record-and-doc corrections, all
      fixed in-cycle, and no functional defect: not one line of Go changed at
      verify.** The renderers survived **31 mutants** — the 19 in the matrix,
      each reproduced independently from its stated edit, plus 12 novel ones —
      and every one is killed by a named test. `V-F1` was the real one: the
      `impact` hunk of `docs/api-contract.md` cited DEC-050 alone where the
      `wrapped` hunk cited DEC-050 **and** the amended DEC-030, so a reader
      following the `impact` links landed on original text the amendment had
      since changed. `V-F2` corrected this spec's own §9 prose, which still said
      the DEC rows were *"52 and 2"* after R3 made it 3. `V-F3` amended AC-13's
      literal command, which matched one more line than the criterion meant.
      `V-F4` was ruled **not a defect**: the header blockquote records the cycle
      that wrote it, and four of four shipped siblings hold that convention.
      **`V-F1` is also SPEC-088's `V-F1` in a new guise, and its own class:**
      SPEC-088's was a guard blind to a file its session had just written;
      this one is a *record* blind to the base it was taken against — the V-F1
      fix moved `docs/api-contract.md`, and `M-D1` and `M-D4` had to be
      re-pinned. `V-F0` names the mechanism, and this ship re-pinned a third
      row (`M-D3`) because codifying a §12 clause edits `AGENTS.md`, `M-D3`'s
      own target. It is **held at N=1** — see *Held codification candidates*
      below, where the count stays at one case because the unit is the spec,
      not the row. One §12 clause **cleared at N=3** and is written into
      AGENTS.md; the zsh quirks behind it went into
      `projects/_templates/spec.md`'s new `### Traps` subsection instead, which
      is the gap build's Q2 named. `pr:219` (design, with framing folded in),
      `pr:220` (build), `pr:221` (verify), `pr:222` (ship).
- [x] SPEC-087 (shipped on 2026-09-07) — **`Y3` derives instead of caching.**
      Split out of SPEC-086's re-framing rather than absorbed; sequenced
      BEFORE SPEC-086 design and landed there, so DEC-050 will be the **first**
      decision in six that costs no hand re-pin. Complexity **S** held: one
      file, `+77/-39`, no Go, no DEC, no user-facing doc, inventory table
      byte-identical. Design's headline correction to this page's item 2: the
      second derivation did not have to be invented — **`Z7` already computed
      it** (SPEC-082 LD10) and was measured **blind**, because it
      re-implemented `inventory.sh`'s two greps verbatim instead of reading
      what the script emits. Fork 1 was a repair, not an invention; Fork 2
      rejected; Fork 3 retired as moot. Verify then overturned build's own
      floor argument by measurement — `Y3` had been borrowing `Z7`'s
      non-vacuity floor and went vacuously green when that floor was deleted —
      and gave `Y3` a floor of its own. `Y4`'s pin stays, routed to
      **SPEC-088**, which now exists as a file. `pr:202` (build), `pr:203`
      (verify), `pr:205` (ship).
- [x] SPEC-088 (shipped on 2026-09-18) — **the stray-tag guard and the
      decision-type vocabulary.** It added two `scripts/test-docs.sh`
      assertions, changed one template line and regenerated one row:
      `Documentation assertions (distinct ids)` moved 198 → 200, once for
      the pair. Framing kept two of the four items routed here and split two
      out, `Y4` to SPEC-092 and id reservation to SPEC-093. **`AC1`** reads
      the `insight.type` vocabulary from `decisions/_template.md`'s own
      comment, and the rows from what `scripts/inventory.sh` *emits*. It
      fails in both directions. The template narrowed from five values to the
      two the inventory counts, because the three it dropped had **0** uses
      in 52 records. **`AC2`** fails when a closing `content` or `invoke` tag
      sits alone on a line of any file. The whole-line anchor stays green
      while 4 files mention those tags inline. At the maintainer's direction,
      build also folded in the eight `DEC-*.md` line-6 comments that still
      carried the old vocabulary. Complexity **S** held, with no Go and no
      DEC. **Verify's punch list had one real gap, V-F1.** `AC2` swept
      `git ls-files`, which reads the index, so it could not see a new file
      the session had not yet `git add`-ed. Five of the six historical leaks
      were in exactly such a file. So is `DEC-050`, the file the sequencing
      argument named. Verify fixed it in-cycle by widening the sweep to
      `--cached --others --exclude-standard` with `grep -d skip`, and the
      unstaged-file probe V-B7 now fires. The other findings were record
      corrections, including M-A0's re-pin: its stated edit was not the one
      that ran. Both guards are local gates, because `just test-docs` has no
      CI job. SPEC-086's design was sequenced behind this spec at framing,
      and it now runs guarded. Two codification candidates are held, not
      written. See *Held codification candidates* below. `pr:214` (frame),
      `pr:215` (design), `pr:216` (build), `pr:217` (verify), `pr:218`
      (ship).
- [x] SPEC-089 (shipped on 2026-09-08) — **the buffer drops a field silently,
      and `edit` gives no applied/no-op signal.** Two capture-integrity defects
      from a field report, one PR. Attached to this stage **without gating it**:
      it advances no Success Criterion, it arrived from heavy agent use of
      `brag` rather than from the stage plan, and it was taken here because the
      defect shape — *the corpus quietly holding something other than what the
      author wrote* — is the one this project is named for. `editor.Parse` now
      rejects a repeated canonical header (**DEC-051**) across all three editor
      ingresses; `brag edit` prints the mutated id on stdout on a write and
      nothing on a no-op (**DEC-052**). Complexity **S** held. Verify returned a
      punch list of four findings, three fixed in-cycle: the duplicate guard
      covered only two of five keys with a fully green suite, two of three
      buffer templates could have broken `brag add` and `brag learn` for every
      user without a test firing, and `docs/api-contract.md` had gone stale
      because AGENTS.md §9's premise audit was not run at build. The fourth was
      routed — see SPEC-090 below. `pr:206` (design), `pr:207` (DEC renumber),
      `pr:208` (build), `pr:209` (verify), `pr:210` (ship).

- [x] SPEC-094 (shipped on 2026-09-23) — **`story` stops listing a failure as
      a win.** Given a file at SPEC-086 design (2026-09-18), in the same edit as
      this line, and split at its framing (2026-09-22) on defect shape: `story`
      carries `"type": "failed"` in JSON and dropped it only in markdown, while
      `summary` cannot represent a failure per entry at all, so `summary` went
      to **SPEC-095**. **The mechanism is DEC-054**, and it implements DEC-050
      row 4. A profile whose `candor` is exactly `promotional` (`exec`, `skip`)
      omits recorded failures before threads are built and says so three ways:
      an `Omitted: <n> recorded failures …` line under `Beats:`, an
      always-present `omitted_failure_count`, and a fixed clause the binary
      appends to the framing directive. Every other value labels them
      `- ✗ <id> (failed):`, so a typo never drops anything. A failure with an
      impact is still an impact beat, so no count was renamed (DEC-048), and
      `internal/story` now calls `aggregate.HasImpact` instead of restating it.
      **The clause was measured, not assumed.** A bare note survived the
      consuming model in 0 of 10 runs at framing. With the clause, on the
      wording design locked, the omission was stated in **25 of 25** runs
      (Haiku and Sonnet, `exec` and `skip`), as the prose's last line every
      time. **Verify measured the candid half (DEC-054 T4) and it fired in
      part.** On `manager`, failures credited as wins fell from **22/30 to
      4/30** mentions; on `me` the label is inert (1/30 → 2/30). All six
      residual cases are #473, the one failure whose impact describes its own
      fix. So DEC-050's invariant holds in the bundle and not always in the
      prose. That is recorded in DEC-054's ship amendment, and the directive
      line it asks for is **SPEC-096** (below). **Verify's punch list was one
      record fix and no functional defect:** the `M-D1` matrix row said *"one
      of three"* mentions, which build could not reproduce, and verify pinned
      it to the second (`:938` at `c3f707a`), which reproduces design's hash.
      Nine novel mutants were all killed. That row is the negative of the §12
      refinement codified at this ship, *a pinned diff has exactly one literal
      reading* (see *Held codification candidates*, item 5). `pr:226` (frame),
      `pr:227` (design), `pr:228` (build), `pr:229` (verify), `pr:SHIPPR`
      (ship).
- [ ] SPEC-095 (frame) — **`summary` moves a failure out of `## Highlights`.**
      Given a file at SPEC-094's framing (2026-09-22), in the same edit as this
      line. It is DEC-050 row 3, which states the posture in full: SPEC-086's
      partition into `## What didn't work` / `failures_by_project`, on a third
      surface, reusing `aggregate.SplitFailures`. **GO at S**, with nothing
      blocking it and no dependency on SPEC-094's maintainer question.
      Measured: 4 of 4 failures listed as highlights under
      `summary --range month`. **It gates v0.7.0 with SPEC-094**, since the two
      together are the half of Success Criterion 4 still owed. Verify trap:
      `--range month` is rolling, and the four failures leave it on the live
      corpus around 2026-10-06.
      **It is now the last spec gating v0.7.0, and the next spec to frame.**
      Recorded at SPEC-094 ship (2026-09-23), on this entry rather than only in
      that spec's reflection, for SPEC-088 Q1's reason. The cut itself is its
      own spec after this one.
- [ ] SPEC-096 (frame) — **the candid directives say what a failure label
      means.** Given a file at SPEC-094 ship (2026-09-23), in the same edit as
      this line, on the maintainer's ruling of that day. SPEC-094's verify
      measured DEC-054 T4: the `✗ <id> (failed)` label cut failures credited as
      wins on `manager` from 22/30 to 4/30, did nothing on `me`, and left one
      failure, #473, credited as shipped because its own impact reads like a
      fix. Neither `me.md` nor `manager.md` tells the model what `✗` means.
      The candidate is one line in each. It edits **shipped profile assets**,
      which SPEC-094's LD14 kept out of scope, and it **requires re-running the
      52-run measurement** on whatever wording it locks. Unframed, and
      provisionally S. **Does not gate v0.7.0**, by the maintainer's ruling: the
      bundle already satisfies DEC-050, and this is about the model's prose.
- [ ] (not yet written, `bug`) — **`--type` negation is inexpressible and fails
      silently.** `--type '!failed'`, `'-failed'`, `'shipped,failed'`,
      `'!=failed'` and `'NOT failed'` each return **exit 0 with zero rows, no
      diagnostic**; only `--type ''` errors. `internal/storage/store.go:389` is
      exact-match inclusion. Blast radius is all seven `--type` surfaces.
      Routed here, **not** to `guidance/questions.yaml`: the behaviour, the
      line and the fix shape are all measured, so it is a bug and not a
      question — and filing it as a question would move `Y4`'s pinned counts
      and force an inventory regeneration for zero information gain.

- [ ] SPEC-090 (frame) — **`brag add --json` drops a repeated key silently,
      the way the editor buffer used to.** Routed out of SPEC-089 verify
      (2026-09-08); **given a file at SPEC-089 ship (2026-09-08) so the id is
      claimed and the item has a named owner**, per SPEC-087 LD6 and this
      spec's own DEC-050→DEC-051/052 renumbering — the third instance of an id
      reserved in prose. The file carries the transcribed evidence, a re-measured
      reproduction, and an explicit fork; it is `cycle: frame` and genuinely
      unframed. SPEC-089 fixed the silent field drop in
      `editor.Parse`, which reaches all three editor ingresses (`edit`, `add`
      editor mode, `learn`). The fourth ingress does not go through `Parse`:
      `internal/cli/add_json.go:24` uses `encoding/json`, whose documented
      behaviour for a repeated object key is **last-wins, silently**.
      Measured: `echo '{"title":"t","impact":"REAL","impact":"CLOBBERED"}' |
      brag add --json` exits 0 and stores `CLOBBERED`. So the two write
      ingresses to the same corpus now *disagree* about what a repeated field
      means — one rejects first-wins, the other silently takes last — and the
      silent one is the scripted path an agent uses. SPEC-089 scoped this out
      correctly (it framed Bug A as a `Parse` defect), but the framing that
      justifies the fix — *the corpus must not quietly hold something other
      than what the author wrote* — does not stop at `Parse`. Needs a decision
      (reject vs. warn) because `encoding/json` has no duplicate-key hook;
      rejecting means a `json.Decoder`-token pre-pass over the object.
      Same shape on the MCP `brag_add` ingress, which the SDK decodes with the
      same package.

- [ ] SPEC-091 (build) — **the help surface reads as families and tells the
      truth.** Groups `brag --help` into Write / Read / Digest / Admin (moving
      `cmd/brag/main.go`'s `AddCommand` block into a testable
      `cli.AssembleRoot`) and closes three per-command `--help` truth-gaps: the
      enforced `impact` cap, the `$EDITOR` buffer format, and the CLI-vs-MCP
      provenance asymmetry. Gives `brag learn` a visible home in the Write
      group, which partly answers round-2 feedback §3.6 (`failed` is
      under-used largely because it is hard to find). Attached to this stage
      **without gating it**, like SPEC-089. Frame and design were collapsed by
      user direction; complexity **M**. `pr:211` (design). **Entry added by the
      orchestrator in SPEC-088's framing PR:** the file has existed since #211
      but had no entry here, which is why the previous count carried a `0` for
      it.

- [ ] SPEC-092 (frame) — **`Y4` derives against a parsed register.** Split out
      of SPEC-088 at framing (2026-09-15) and **given a file in the same edit as
      this sentence**, per SPEC-087 LD6. **GO at S.** SPEC-087's open question
      (LD6) is **answered**: an independent oracle does exist — a real YAML parse
      builds a document model where `inventory.sh` matches lines at an exact
      indent, which is the same class of independence `Y3` gets from a body
      heading. Measured: `/usr/bin/ruby` (2.6.10) and `/usr/bin/python3` (3.9.6 +
      PyYAML 6.0.3) each reproduce **21 questions / 8 open**, and
      `scripts/test-docs.sh:15-19` already hard-requires `jq`, so a harness tool
      dependency is a decision this repo has made once before. What the spec
      still owes is the fork *which* oracle (interpreter, skip-if-absent, or a
      structural `awk` partition check). **Not a blocker for SPEC-086** — `Y4`
      pins only the two `guidance/questions.yaml` rows, which SPEC-086 does not
      move *unless its design files a question*; if that happens, pull this
      forward rather than paying a sixth hand re-pin. `Y4` keeps its id, so this
      spec leaves the inventory table byte-identical.
      **Does not gate v0.7.0.** This is the maintainer's call of 2026-09-18,
      recorded at SPEC-088 ship: the spec is harness-only.

- [ ] SPEC-093 (frame) — **ids are claimed by files, not by prose.** Split out
      of SPEC-088 at framing (2026-09-15), file created in the same edit — which
      is the rule it exists to mechanise. **GO at M provisional** (S for the
      mechanism; M because it must also decide the backfill, and that branch
      moves a user-facing table). Measured on `6dda56a`: `next_id`
      (`scripts/_lib.sh:107`) reads **filenames in the working tree** and has
      exactly two callers, `new-spec.sh:48` and `new-stage.sh:30` — **no recipe
      creates a DEC**, so decision numbers are typed by hand. `DEC-050` has no
      file and 7 tracked files mention it; `STAGE-024/025/026` have no files and
      7 tracked files mention them; `next_id` returns `DEC-054` and `STAGE-028`,
      so none of the four is reachable. The two modes are opposite: those four
      are **holes** (harmless-but-invisible — `SPEC-016` and `SPEC-070` have been
      holes for months unnoticed), while the failure that actually cost
      something was a **collision**, SPEC-089 taking `DEC-050` on an unmerged
      branch and renumbering at #207. Nothing detects a collision today. **Not a
      blocker for SPEC-086:** `next_id` never returns a hole, so SPEC-086
      authoring `DEC-050` is safe by construction, and the only spec in between
      is SPEC-088, which emits no decision record.
      **Does not gate v0.7.0.** This is the maintainer's call of 2026-09-18,
      recorded at SPEC-088 ship: the spec is harness-only.
      **Also owned here, routed at SPEC-088 ship (2026-09-18): `just
      advance-cycle` strips the inline comment from the line it rewrites.**
      `update_frontmatter_scalar` (`scripts/_lib.sh:173`) runs
      `sub(/:[[:space:]]*.*$/, ": " val)`, which replaces everything after the
      first colon, the `# frame | design | build | verify | ship` comment
      included. Its only caller is `scripts/advance-cycle.sh:30`. It fired on
      SPEC-088 at design, at build and at ship, and each time the comment was
      restored by hand. Across `specs/done/`, **80 of 84** archived specs carry
      a bare `cycle:` line, while `projects/_templates/spec.md:10` carries the
      comment. It belongs here because this spec already opens
      `scripts/_lib.sh`, for `next_id` at `:107`. Framing decides whether it
      is a line in this spec's build or a split of its own. Either way it has
      a named owner. **Not fixed at SPEC-088 ship**, by instruction.
      **Also owned here, found at SPEC-088 ship (2026-09-18): `just
      archive-spec` turns `AC2` red on every archive.**
      `scripts/archive-spec.sh` moves the spec with a plain `mv`, not
      `git mv`, so the old path stays in the index with its deletion
      unstaged. That is SPEC-088's M-B6 state, which LD9 makes red by design.
      Measured on the first archive since `AC2` landed: `FAIL: AC2`, naming
      the old path with grep's `No such file or directory`, and green once the
      move was staged. Nothing leaked. `git mv` in `archive-spec.sh` is the
      fix that keeps LD9 intact. Skipping `git ls-files --deleted` inside
      `AC2` would overturn M-B6, a locked decision. It is routed here, beside
      `advance-cycle`, because both are recipes in `scripts/` that leave the
      tree in a state the repo then has to repair by hand, and this spec
      already works on that recipe layer (`next_id`, a possible `new-dec`).
      Framing may re-route or split it. **Until it lands, a ship stages the
      move before it runs the gates.**

- [ ] (not yet written, `bug`) — **`brag delete` has Bug B's defect, unfixed.**
      Identified at SPEC-089 build as follow-up ("own spec if wanted") and
      recorded only inside that spec, which is now archived — routed here at
      ship so it does not die in `specs/done/`. Measured, not assumed:
      `internal/cli/delete.go:75` prints `Aborted.` and `:86` prints `Deleted.`,
      both to **stderr**, and both paths `return nil`. So a batch driver cannot
      tell a completed delete from a declined confirmation on stdout *or* on the
      exit code — the exact three-state ambiguity DEC-052 just closed for
      `edit`. **No owner assigned**, deliberately: it is a near-mechanical
      application of an existing decision, and the honest question at framing is
      whether it is a spec at all or a line in whichever spec next opens
      `delete.go`. Cheaper to leave unowned and visible than to invent a spec id
      for it now.

**Count:** 6 shipped / 0 verify / 1 in build / 0 in design / 5 framed / 2 not
yet written
(Re-derived at SPEC-094 ship, 2026-09-23. It comes from each file's own
`cycle:` field, read with
`awk '/^---$/{f=!f; next} f && /^[[:space:]]+cycle:/{print $2; exit}'` over
`projects/PROJ-008-*/specs/*.md` and `specs/done/*.md`, and is not
incremented. Every file's `stage:` is STAGE-023, read the same way. shipped =
SPEC-085, SPEC-086, SPEC-087, SPEC-088, SPEC-089 and SPEC-094, the six
`cycle: ship` files in `specs/done/`. in build = SPEC-091. **Nothing is in
design or verify.** framed = SPEC-090, SPEC-092, SPEC-093, SPEC-095 and
SPEC-096, the five `cycle: frame` files in `specs/`. Only SPEC-095 has had a
framing pass and carries a `## GO / NO-GO`; it was created at SPEC-094's
framing. The other four are id-claiming files rather than fully framed specs:
SPEC-090 was created at SPEC-089's ship, SPEC-092 and SPEC-093 at SPEC-088's
framing, and SPEC-096 at SPEC-094's ship. `not yet written` is the two
`- [ ] (not yet written…` entries: `--type` negation and `brag delete`. The
two recipe defects routed at SPEC-088's ship, `advance-cycle` and
`archive-spec`, add no entry, because both have an owner, SPEC-093. Both fired
again at SPEC-094's ship, as they did at SPEC-086's: `advance-cycle` stripped
the `cycle:` comment, which was restored by hand, and the archive's plain `mv`
was staged with `git add -A` before any gate ran.)

**The stage does NOT close here.** Success Criteria 1, 2 and the DEC-014/
DEC-048 envelope line are met by SPEC-085. **Criterion 4** (*the celebratory
digests do not silently absorb failures*) is now **three quarters met**:
SPEC-086 shipped `wrapped` and `impact` on 2026-09-20, and SPEC-094 shipped
`story` on 2026-09-23. The `summary` quarter is owed by **SPEC-095**, which is
why it is now the last spec gating v0.7.0. SPEC-096 does not gate it: it acts
on the prose a model writes from a `story` bundle, not on the bundle.
Criterion 3 is **half met** — see *The Fork 3 finding* below.

### The conditional spec fired (2026-09-05, at SPEC-085 design)

This backlog's second slot was written *"only if SPEC-085's Fork 4 turns out to
need more than a documented default."* It did. Put to the user at design with
the December reading spelled out, Fork 4 resolved to a **named
`## What didn't work` section** on both celebratory digests — rejecting both
exclude-by-default and include-silently.

That is a renderer change on two surfaces **plus two DEC-048 count renames**
(once *Impact moments* stops carrying every with-impact entry, its headline
count no longer means what it says). A second M of work on different files with
its own decision record, so it was split rather than absorbed — which is what
kept SPEC-085 at M.

The measurement that made "include silently" unacceptable, and that framing did
not have: **neither `impact` nor `wrapped` renders `type`, in either format.**
`ToImpactMarkdown` emits `- <id>: <title>` + the impact text, and
`impactEntry` is a deliberately narrow 4-key JSON projection (DEC-028 choice
4). A failure there is not merely unflagged — it is *unrepresentable as a
failure* without a renderer change.

## Design Notes

Four measured facts that constrain every spec in this stage. Each was
re-derived on 2026-09-05 against `1775feb`; none should be re-litigated from
memory.

1. **`type` is free-form, not a closed set.** `internal/cli/add.go:101`
   describes it as *"free-form category (shipped, learned, mentored, ...)"* and
   nothing validates it. The live corpus holds **19 distinct values** across
   397 entries, including near-duplicates (`shipped` 201 / `ship` 15,
   `fixed` 2 / `bugfix` 1) and **113 entries with no type at all**. A `type`
   value therefore carries **no guarantee** today — which is an argument both
   for using it (no migration, no new concept) and for the fork having to say
   what makes this value different.

2. **`learned` already exists and means something else.** Eight entries carry
   `type: learned`, and reading them, they are **lessons framed as wins** —
   *"Verify caught an unpinned test seam…"*, *"Re-running a pre-flight after a
   dependency bump paid for itself."* Not one records something that failed.
   So `learned` is **occupied**, and SPEC-085 cannot quietly adopt it.

3. **Tags are normalized — DEC-004 is superseded.** DEC-004 (comma-joined
   `TEXT`) was superseded by **DEC-015** on 2026-06-06; `0003_normalize_tags.sql`
   creates `tags` + `taggings` with a polymorphic membership join and two
   indexes. A reserved tag is therefore a **first-class indexed lookup**, not a
   `LIKE '%…%'` scan. Any framing input that costs a reserved tag using DEC-004's
   model is costing a schema that has not existed since June.

4. **The memory line renders `type` and does not render tags.** DEC-044's line
   shape, in `internal/memory/memory.go:224`, is:

   ```
   - <id> <YYYY-MM-DD> [<project>/<type>] <title> — <impact>
   ```

   This is the sharpest fact in the stage. A failure marked by **`type`
   travels to the agent-facing read surface for free**; a failure marked by a
   **reserved tag is invisible there** unless DEC-044's locked line shape is
   amended. The two options are not equivalent-cost, and the difference lands
   on the read path rather than the write path.

Two further constraints, inherited rather than discovered:

- **The envelope.** Six `internal/export/` renderers emit a headline count
  under DEC-014, and **DEC-048** (2026-08-23) legislates that the count must
  name what it counted, binding all six forward. Any new surface inherits both.
- **The ranker.** `brag memory` fuses three ordinal lists — recency, relevance,
  project — by RRF (DEC-043), with `PoolLimit = 200`. **There is no type-aware
  or tag-aware signal.** A failure entry competes on recency alone and is
  crowded out by volume; if failures should surface reliably, that needs an
  argued answer, not an assumption that they will.

### Corrections from SPEC-085 design (2026-09-05)

Four of this page's measured facts moved or were wrong when re-derived at
design against `main` at `81e639d`. Recorded here so a later stage does not
re-inherit them; none changes a conclusion this stage reached.

| This page says | Measured at design |
|---|---|
| corpus **397** | **398** (framing's own brag landed) |
| *"**19** distinct values"* | **18** distinct non-empty values. The 19th was the empty bucket — the same 113 entries this page separately reports as *"with no type at all"*, counted twice in two units. |
| *"eight entries carry `type: learned`, and reading them, they are **lessons framed as wins**… Not one records something that failed"* | The eight are **not uniform.** id 18 is a PROJ-001 smoke-test artifact, not a lesson. ids 80/361/383 take a **failure as their subject** and frame it as a win recovered. `learned` is closer to *post-mortems with a recovery* than to *wins*. **The conclusion holds and gets stronger** — redefining it would make `brag list --type learned` return a mix, failing Success Criterion 2. |
| Fork 1's table: `--type` on **5** of 8 commands | **7** — `list`, `export`, `summary`, `impact`, `wrapped`, `story`, `coverage`. The denominator omitted four commands. Retrieval is *more* free than claimed. |

**And one correction to a routing rationale on this page.** *Read-path
discoverability* is routed to STAGE-024 on the premise that `BRAG.md` *"has 22
headings and **none** is a read-the-corpus section."* Measured on the live tree:
**26** markdown headings (16 at `##`), and `## Reading entries back` (line 382)
is exactly such a section — 31 lines listing eight read commands. The literal
string `brag memory` appears **zero** times in the file, not twice.

The gap is real but differently shaped, and smaller: **the read section exists
and omits the one command written for agents.** STAGE-024 should frame it
against that, not against "BRAG.md is write-only."

### Correction from SPEC-085 verify (2026-09-05) — for SPEC-086 framing

**SPEC-086's Fork B splits the seven `--type` surfaces into *celebratory*
(`wrapped`, `impact`) and *neutral* (`summary`, `story`, `export`,
`coverage`), and asks design to decide per surface. Two of those four are
misclassified, measured on a corpus carrying `type: failed` rows.**

`export` and `coverage` are genuinely neutral — `export` renders `type` in its
per-entry table, `coverage` is provenance-only. The other two are not:

| Surface | What it actually does with a failure |
|---|---|
| `story --audience exec` | renders it as a `★` beat with **no `type` in markdown**, under a printed directive that says *"build the narrative from those outcomes"*, *"Terse and promotional… No process, no messy middle"*, *"Quantify wherever the impact beats give you a metric."* The output is a **prompt for an LLM**, so this is a strictly worse leak than `impact`'s: `impact` shows a human a mislabelled row, `story` instructs a model to launder it. |
| `summary` | its section is literally `## Highlights`; the failure appears there with no `type` in either format (`summary`'s JSON highlight is a 2-key `{id,title}`, narrower than `impact`'s 4-key). |

Two constraints this puts on Fork B:

1. **"Is `story` celebratory?" is not a property of the command.**
   `--audience me`'s directive is candid by design (*"Include the messy middle:
   struggles, false starts… are the point"*) while `--audience exec`'s is
   promotional. Fork B's per-surface framing cannot express a per-**profile**
   answer, and needs to.
2. **`story --format json` already carries `"type": "failed"` on each beat.**
   Only the markdown path is lossy — which makes `story` a cheaper fix than
   `impact`/`wrapped`, not a more expensive one.

Reproductions are in SPEC-085's `## Verify`. Nothing here changes SPEC-085,
and none of it was known at the time Fork B was written — it needed a corpus
with failures in it, which is what SPEC-085 shipped the ability to make.

**Also sharpened, not overturned:** SPEC-085's *interim risk* is bounded partly
because *"the user controls when the first [failure with an impact] is
written."* True, but `BRAG.md`'s new section tells the agent *"The `impact`
field is still worth filling in"* and shows `-i` in its example — so the
leaking shape is the **documented default path**. The risk is still acceptable;
what bounds it is the empty corpus, not user restraint.

### Carried into SPEC-086 (added at SPEC-085 ship, 2026-09-05)

Two items with a **named owner**, because the last time one of them was routed
without one it did not fire.

**1. `--type` negation is inexpressible, and it fails silently.** SPEC-086's
Fork A needs *"everything except failures"*; `internal/storage/store.go:388` is
exact-match inclusion on a single string. Re-confirmed at ship on a two-row
corpus: `--type '!failed'`, `--type '-failed'` and `--type 'shipped,failed'`
each return **exit 0 with zero rows** — no error, no diagnostic. Only
`--type ''` errors. A user filtering failures out of a review gets an empty
document rather than a message. This is a **dependency of Fork A**, not a
footnote.

**2. `Y3` and `X3` pin the same numbers by different means, and one of them
should derive rather than cache.** STAGE-022's close promoted this to a
stage-level lesson and then **held** it with the trigger *"routed to the next
spec that opens that file."* SPEC-085 **was** that spec — it inserted 80 lines
of Group `AB` into `scripts/test-docs.sh` — and hand-re-pinned `Y3` anyway,
making it **five consecutive specs** (SPEC-081, 082, 083, 084, 085). The
routing failed for the reason STAGE-022 recorded one line earlier about
`archive-spec`: *"Routing it as a candidate after the first hit did not prevent
the second; the guard did."* An anonymous "next spec" is not an owner. Named
here: **SPEC-086**, which will open the same file. Not codified as a rule —
that is N=2 same-outcome and this repo wants N=3.

Worth pairing with what *did* work, so the lesson is not read as "greps do not
help": AGENTS.md §9 half-(b) (grep for the **value**, not the idea) caught `Y3`
**at design** on SPEC-085, the first time in those five specs that it was found
early rather than late. The rule works; the duplication it keeps finding is the
thing to remove.

**Both items discharged at SPEC-086 re-framing, 2026-09-06:**

- **Item 1 (`--type` negation) — DESIGNED AROUND, not fixed and not blocking.**
  Re-confirmed independently (all five negation spellings: exit 0, zero rows,
  no diagnostic). But it is **not** a dependency of Fork A, which is what it was
  filed as. Mutation **M-1** — `aggregate.WithImpact` changed to
  `if e.Impact != "" && e.Type != "failed"`, confirmed by content hash —
  showed a **one-line in-memory predicate does the entire selection job**;
  `impact` and `wrapped` read all in-window rows once and partition in Go, the
  same shape `brag coverage` already uses because it needs both classes. No
  query in SPEC-086 wants `--type '!failed'`. The bug is real and is now a
  named `bug` entry in the backlog above.
- **Item 2 (`Y3`/`X3`) — SPLIT to SPEC-087, with a file.** Not absorbed:
  SPEC-086 was already splitting on Fork B, and taking this on would have
  pushed it back to L. Not re-routed anonymously either — the file exists,
  which matters mechanically as well as socially: `scripts/_lib.sh:107-119`
  computes `next_id` by scanning **filenames**, so an id reserved only in
  backlog prose gets handed to the next `just new-spec`. **SPEC-087 should
  land before SPEC-086 design**, which authors DEC-050 and would otherwise be
  the sixth consecutive hand re-pin. One correction to this page's framing of
  the item: `X3` and `Y3` are **not** redundant — `X3` catches a stale page,
  `Y3` catches `inventory.sh` itself going wrong, and `Y3`'s own comment names
  that failure mode (*"script and page would still agree, just agree on 48"*).
  The defect is that `Y3`'s **expected value** is a hand-maintained literal,
  not that the assertion exists.

### Re-framing corrections (SPEC-086, 2026-09-06, `main` at `df369e9`)

Six of this page's and SPEC-086's measured facts moved when re-derived. Recorded
so a later stage does not re-inherit them.

**The one that changes a conclusion: the interim risk is REALISED.** SPEC-085
accepted it as bounded because *"the corpus holds zero such entries."* Measured
2026-09-06: **one `type: failed` row exists** — id **420**, project
`contextcore-pilot-harness`, created `2026-09-06T00:44:44Z`, agent-authored,
**carrying a full impact statement**. It renders as an unmarked win on four
surfaces today. This is *not* the drafted `brag learn` entry the orchestrator is
deliberately holding; that one is still held. **The zero was a choice for
exactly one day**, and the argument for SPEC-086's priority is no longer *"this
will happen"* but *"this is happening, and it compounds as the corpus grows."*

| This page / SPEC-086 says | Measured 2026-09-06 |
|---|---|
| corpus **397** / **398**; `impact` `324/398`; `wrapped` `398` | corpus **420**; `Entries: 346/420 with impact`; `Entries: 420` |
| **0** `type: failed` rows | **1** (id 420) |
| *"**18** distinct non-empty type values"* | **19** — `failed` joined the set, as DEC-049 intended |
| `--type` inclusion at `store.go:**388**` | **`:389`**. `:388` is the `if f.Type != ""` guard. The negation precedent at `:404` is cited correctly. |
| Fork B: the `story` leak is an **`exec`** problem | **All four bundled profiles** render `- ★ 420:` with no `type`. **Two** are `candor: promotional` (`exec` **and `skip`**), two are `candid`. The renderer defect is profile-independent; only the *harm* is profile-dependent, because the promotional directives instruct a model to promote the unlabelled beat. |
| Fork B: `summary` renders a failure with no `type` | **Split, not total.** `## Summary → By type` **does** print `failed: 1` (and JSON `counts_by_type` carries it). Only `## Highlights` is lossy — and its JSON highlight is a **2-key** `{id,title}`. `wrapped` has no equivalent honest counterpart: its `Top types` is a **top-3** (`shipped 211`, `milestone 29`, `ship 15`), so `failed: 1` will never appear there. |
| Fork D: *"DEC-014 part 4 has a precedent — `wrapped` omits body sections on an empty period"* | **The precedent covers the wrong case.** DEC-014 part 4 governs the empty **document** (confirmed: `wrapped 2024` ends after `Entries: 0`). For an empty **section** in a non-empty document the two surfaces already **disagree**: `wrapped` renders a bare `## Impact moments` heading, `impact` omits its whole body. Fork D is a live fork, not an application of a settled rule. |
| Fork E: *"five files in `internal/export/` carry the affected strings"* | Grep-shaped. Measured by mutation: the section heading fires **4 tests in 2 files**; a headline rename fires **9 tests in 3 files across 2 packages**, one of them **`internal/cli`**. `memory_test.go:247` (the byte-length assertion §9 cites) **does not move**. |

**And one methodological correction worth keeping.** The claim that
`story --audience exec` renders a failure with *"zero occurrences of `failed`
anywhere in the markdown"* is **false as literally stated** — the string appears
on 3 lines, all incidental prose inside *other* entries' impact text. The
accurate claim is stronger: *the renderer emits no `type` field at all*, so a
grep for `failed` returns hits that have nothing to do with the failure row.
A count from grep is a hypothesis.

**The survived mutant, which is the most useful single result.** Mutation
**M-1** (drop `failed` from `aggregate.WithImpact`) fired **zero** of the
suite's 1070 tests. The repo has **no existing coverage of a `failed` entry
flowing through `WithImpact`, `impact` or `wrapped`.** Every test proving the
new behaviour must be written from scratch; none can be adapted, and a green
suite is not evidence.

### The Fork 3 finding, which belongs to the stage rather than to one spec

This page's *Design Notes* item 4 and its ranker paragraph say a failure
*"competes on recency alone and is crowded out by volume."* Measured, it is
stronger and structurally different: **it is not out-ranked, it is not a
candidate.**

`Gather` reads `List{Limit: PoolLimit}` with `PoolLimit = 200`, so a bare
`brag memory` sees only the 200 most recent entries. On the live corpus the
horizon is entry 207, dated 2026-07-06 — **61 days** — and **198 of 398
entries (49.7%)** cannot be returned at any budget. Five of the eight
`type: learned` entries are already outside it.

This kills the option that reads cleanest: **a fourth ordinal list cannot fix
it.** `buildMatchRank` drops ids not in the pool (`memory.go:191`) and
`buildProjectRank` iterates the pool, so no ranking term can introduce an entry
`Gather` did not fetch. Only a fourth *read* reaches an out-of-pool entry — and
that is pool composition, not fusion, which is why SPEC-085 could answer
`memory-slice-fusion-constants` cleanly in the negative.

**Consequence for this stage:** Success Criterion 3 (*"`brag memory` returns
failures to an agent reading the corpus cold"*) is **half met** by SPEC-085 —
the line shape labels a failure `[project/failed]`, but durability past the
61-day horizon is not addressed. Filed as
`memory-pool-composition-excludes-older-entries` with the measurement, and
deliberately not answered while there are zero failure entries to calibrate a
slot count against.

### Findings routed out of SPEC-087 verify (2026-09-06)

Three items that SPEC-087 surfaced and deliberately did not decide. Each has a
named owner, because *"the next spec that touches it"* is the routing failure
this stage already named once.

> **Status at SPEC-087 ship (2026-09-07):** item 1 **discharged** — SPEC-088
> now exists as a file (renamed at SPEC-088's framing, 2026-09-15, to
> `specs/SPEC-088-the-stray-tag-guard-and-the-decision-type-vocabulary.md`),
> so the id is claimed rather than reserved in prose, and this note's evidence
> is transcribed into it. Item 2 **discharged** — both candidates were written
> into `AGENTS.md` §12; see the closing note under item 2. Item 3 stays open by
> design, with no owner.
>
> **Status at SPEC-088 ship (2026-09-18):** item 1 is **discharged**.
> `decisions/_template.md` now offers `decision | reservation`, the fork's
> option (b). `AC1` compares that line against the rows
> `scripts/inventory.sh` emits, in both directions, so a value added on one
> side without the other fails the gate. SPEC-087 verify's stopgap warning on
> the template was replaced by the rule `AC1` enforces. Item 3 is unchanged.

**1. The decision template advertises five `insight.type` values; the inventory
tolerates two. → SPEC-088.** `decisions/_template.md` offers `decision |
analysis | recommendation | observation | reservation`; `scripts/inventory.sh`
has a row for `decision` and for `reservation` only. A `DEC-*.md` carrying any
of the other three is counted by neither row and hard-fails `Z7`, which is the
correct behaviour — but the template is where an author picks the value, and it
currently invites three that break the harness. Measured on `main` at
`f4658b0`, so this **predates SPEC-087** and arrived with `Z7` at SPEC-082:

```
FAIL: Z7: the inventory covers 49 of 50 decisions/DEC-*.md files (48 decision + 1 reservation).
```

The choice — teach `inventory.sh` three more rows, or narrow the template's
vocabulary — moves a user-facing table on one side and a template on the other,
so it is a design call, not a verify fix. SPEC-088 already owns `Y4` per
SPEC-087 LD6 and reads the same two files. SPEC-087 verify added a template
line naming the consequence in the meantime; that line is a warning, not the
decision.

**2. Two AGENTS.md §12 candidates from SPEC-087's mutation work. → SPEC-087
ship.** Codification in this repo lands at ship or stage close, never at
verify, so both are carried rather than written:

- **"A mutation pinned by a hash must also pin its diff." Clears the bar at
  N=2 paired-opposing.** NEGATIVE: SPEC-087's M-6 recorded `2eebe0e644ee` with
  no edit text, and build had to hash-search plausible renames to recover
  design's mutant. POSITIVE: the five mutations whose text *was* stated
  reproduced first try. Verify independently re-paid the same cost, producing a
  different-hashed rename (`353b013924bc`) that fires both guards identically —
  a third occurrence of the negative, not a third case. A hash is a checksum of
  a reproduction, not a reproduction.
- **"A no-op mutant voids the probe." Does NOT clear as a new rule.** §12
  clause (1) already requires `shasum -a 256` before and after, and it caught
  both of build's near-misses. Two same-outcome confirming cases for an existing
  promoted clause are not evidence for a new one, and the meta-rule wants N=3.
  What is missing is one sentence of *consequence and order*: clause (1) says to
  confirm, never that a target whose hash did not move produces **no evidence**,
  so the probe's result — including any *expected-green* half, which is where a
  no-op is invisible — must be discarded rather than recorded. Recommended as a
  refinement of clause (1), the shape the existing *"§12(b) refinement"* has.

**Both were written at SPEC-087 ship (2026-09-07), each at the strength verify
argued for and no higher.** The hash-must-pin-its-diff candidate landed as its
own §12 paragraph, *"A mutation pinned by a hash must also pin its diff,"*
labelled **N=2 paired-opposing**. The no-op candidate landed as
*"§12(b) refinement of clause (1) — a no-op mutant produces no evidence, and
the check comes BEFORE the record,"* explicitly **not** a new clause and
explicitly recording that its two cases are same-outcome and one short of N=3;
what it adds is the consequence-and-order sentence clause (1) never had. Verify
declining to write them itself was the right call for a different reason than
the one it gave: codification lands at ship *and* a rule written by the cycle
that discovered its evidence has no second reader.

**3. `inv_row`'s four latent limits, recorded not fixed.** The helper silently
takes the first of duplicate labels; truncates at a `|` inside a cell (and if
the truncated prefix is digits, the callers' `^[0-9]+$` guard passes a *wrong*
number); cannot address a label containing `|`; and inherits `awk -v`'s escape
processing, so a label containing `\t` addresses a different row. All four
require a `|` or a backslash inside a table cell, and `scripts/inventory.sh`
emits neither — every Value is `n "$(… wc -l)"` or `awk '{print $2+0}'`, every
label is literal prose in one heredoc. No owner assigned deliberately: hardening
a helper against inputs its only producer cannot emit is speculative work. This
note exists so that a future spec which teaches `inventory.sh` a computed or
quoted label knows it is the change that makes them reachable.

### Held codification candidates (SPEC-088 ship, 2026-09-18; updated SPEC-086 ship, 2026-09-20, and SPEC-094 ship, 2026-09-23)

Three `AGENTS.md` §12 candidates below the codification bar, plus the two this
stage has cleared. They are listed here rather than in an archived spec because a
candidate routed that way never arrived: SPEC-087's ship routed one to SPEC-088
in its own reflection only, and none of SPEC-088's four cycles saw it. **When a
spec produces a matching case, add it here with its evidence and move the
count.** The stage close reads this list for *Lessons that should update
AGENTS.md*.

**Bar, restated so nobody re-derives it:** paired opposing-outcome cases clear
at **N=2**; same-outcome confirming cases need **N=3**. A **refinement of an
already-promoted clause** clears at **N=2 same-outcome**, on the §12(b)
precedent — but only when both cases exercise the new sentence's own gap, and a
case the parent clause already closes is not one of them (SPEC-088 ship's ruling
on `M-6`).

1. **A refinement of *"A mutation pinned by a hash must also pin its diff"*:
   the record must be **captured from the run**, not composed after it.** Held
   at **N=1**. *(Sentence narrowed at SPEC-086 ship — see the positives.)*
   - **The negative, N=1: SPEC-088's `M-A0`.** Design stated *"restore the
     pre-SPEC-088 line 6"* with hash `89a999aee134`. That is the hash of a
     whole-file revert to the pre-SPEC-088 template, which also rewrites lines
     12–22. The stated edit hashes to `accb16924cb6`. The parent clause's letter
     was met, because an edit *was* stated, and the record still was not what
     ran.
   - **Not counted: SPEC-087's `M-6`.** It stated no edit at all. The parent
     clause closes that on its own, and `M-6` is the NEGATIVE the parent was
     written from. SPEC-088 verify counted it as the second case. Ship does not,
     because `M-A0` *"exposes a different hole"*, in verify's own words.
   - **An adjacent surface, noted and not counted: SPEC-088 framing's `M-4`.**
     Its printed output was hand-summed from three buckets into two, and the
     hand-cleaning hid the eight stale DEC files (SPEC-088 `V-F4`).
   - **Positives, now 2, and they are why the sentence changed.** Verify's
     proposed form was *paste the `diff` the probe helper printed, not a
     description written afterward*. SPEC-086 rules that out as too broad, at
     **2 positives / 0 negatives** against the hypothesis that *prose* is the
     defect. **SPEC-088's `M-B3` and `M-B4`** were prose-described and
     reproduced, each having exactly one literal reading. **SPEC-086's five
     abbreviated rows** (`M-L1`, `M-D1`, `M-D2`, `M-D3`, `M-D4`) reproduced
     their stated hash first try at build; verify then re-derived all five
     *independently from the row's own description*, without reading build's
     transcript, and ran each one verbatim against a pristine `152dbe9`
     checkout, where all five reproduced again. A *paste the diff* rule would
     have failed all ten of those rows for a fault none of them has. So the
     candidate keeps SPEC-088 verify's diagnosis and drops verify's wording:
     the defect is **composition after the run**, and the remedy is capture,
     of which pasting the helper's `diff` is one form.
   - **Not counted, ruled at SPEC-094 ship (2026-09-23): SPEC-094's `M-D1`.**
     Design's row read *"one of three `omitted_failure_count` mentions"*; build
     could not reproduce its hash, and verify pinned it to the second reading.
     Verify ruled it **does not** clear this candidate, and **ship confirms**.
     The candidate's hazard is a record that reads as exactly one edit and is
     not the edit that ran, so a reader is misled with no signal. `M-D1` misled
     nobody: the row disclosed its own ambiguity, and design's hash reproduced
     the moment its reading was named. It is *faithful but ambiguous*, which is
     the parent clause failing, not this refinement. **The counter-argument,
     weighed and rejected:** the matrix preamble says the Diff column is *"the
     replacement the helper applied, verbatim"*, and for this row it was not,
     which is composition after the run in the literal sense. That is true, but
     it is a false claim about a column, and the row beside it said so. Had the
     row been captured from the helper, it would also have been unique, so
     uniqueness is the property that failed, and item 5 is where it is written.
   - **Re-scoped at SPEC-094 ship, so the three items do not overlap.** A
     reproducible probe record has three properties, and each has one home:
     **unique** (the stated edit has exactly one literal reading, item 5,
     codified), **faithful** (that one reading is the edit that ran, this
     item) and **based** (the baseline names the commit it was taken against,
     item 2). `M-A0` is a fidelity failure; `M-D1` a uniqueness failure;
     `V-F0` a base failure. A future case is counted under exactly one of them.
   - **It clears when** a second stated edit with exactly one literal reading
     does not reproduce its hash, because the edit that ran was a different
     one. That is N=2 distinct, and the §12(b) refinement precedent then
     applies.

2. **A baseline hash must name the commit it was taken against.** Held at
   **N=1**. New at SPEC-086 ship (2026-09-20). A refinement of the same promoted
   clause as candidate 1, so its bar is **N=2**, not 3 — it is one case short,
   not two.
   - **The case, N=1: SPEC-086's `V-F0`.** A matrix row pins a probe to a
     *file state* and never says which commit that state came from, so any
     later cycle that edits the probe target invalidates the row **silently**.
     It happened inside the cycle that wrote the rows: verify's `V-F1` citation
     fix moved `docs/api-contract.md` from `bde92cdba356` to `68ae22d061da`, and
     `M-D1` and `M-D4` were instantly stated against a base the branch no longer
     had. Verify re-ran both on the new baseline so the next cycle would not
     hunt a hash that cannot exist. The failure mode is the nasty direction: a
     reproduction failure reads as a defect in the mutant.
   - **Two more rows inside the same spec, deliberately not counted.**
     Codifying candidate 4 below edits `AGENTS.md`, which is `M-D3`'s own
     target, so SPEC-086's ship re-pinned it too (`acc844937b43` at both
     `152dbe9` and `b9d9681` → **`520208b59f9b`** after the clause →
     `b3869cc4cfbb` under the mutant, still firing `AD5` alone). **The counting
     unit is the spec that paid the lesson, not the row that got invalidated**
     — `V-F0`'s two rows were one case, and these are a third and fourth row in
     the same spec. What they add is scope rather than count: the mechanism is
     not specific to a citation fix, and any cycle that edits a document is one
     edit away from invalidating a probe it never read.
   - **Weighed at SPEC-094 ship (2026-09-23): verify's corrected `M-D1`
     names its base** (`:938` at `c3f707a`). That is the fix shape applied,
     and `c3f707a` is a squash commit on `main`, so it will survive. **It is
     an application, not yet a positive.** A positive counts toward a
     paired-opposing N=2 only when it pays off, meaning a later cycle
     reproduces the row after its target has moved. `docs/api-contract.md` has
     not moved since `c3f707a`. It is likely to move soon, because SPEC-095
     edits its `brag summary` section.
   - **Found at SPEC-094 ship: SPEC-086's own named bases, one save and one
     miss.** SPEC-094's build moved both of SPEC-086's re-pinned targets.
     - *Save:* `docs/api-contract.md` was stated as **`68ae22d061da` as of
       `b9d9681`**. The file is `1cd16e2a3d7a` on `main` now, and the row still
       reproduces: `git show b9d9681:docs/api-contract.md` hashes to
       `68ae22d061da`, and applying M-D1's stated edit to it gives exactly the
       stated **`fbadf13d66f1`**.
     - *Miss:* `AGENTS.md` was stated as **`520208b59f9b` as of "this ship
       commit"**. At `c0b840e`, the squash merge on `main`, `AGENTS.md` hashes
       to **`8367a0159c19`**. A second commit in the same PR (`935c8f0`, the
       attribution fix) edited `AGENTS.md` after the re-pin, and the squash
       then erased the commit the row meant. `520208b59f9b` exists only at
       `0568e7a`, reachable today only through the unmerged branch
       `ship/spec-086-digest-failures`.
     **Both rows are SPEC-086's, so they add scope and not count.** The
     paired-opposing precedent (SPEC-023 against SPEC-024) is two specs, each
     of which decided whether to apply the rule. Here one spec wrote both rows,
     and SPEC-094 only moved the files. What they add is the sentence's missing
     half: **the named base must be a commit on `main`**, because a squash
     merge destroys "this commit", and the template's Q4 already says so about
     `pr:` tags.
   - **Invalidation alone is not a case.** SPEC-094's build silently
     invalidated both SPEC-086 rows above, and nobody paid, because nobody
     re-ran them. Every spec that edits a document does this to every earlier
     spec's probes on it. Counting invalidations would clear the candidate on
     noise; the case is a cycle that **pays**, meaning it tries to reproduce a
     row and cannot, because the base moved.
   - **SPEC-094's own rows are named, not re-pinned.** This ship's §12 edit
     moves `AGENTS.md`, which is `M-D3`'s target. Rather than re-pin (SPEC-086
     ship's move), ship named the base: all nine of the matrix's baselines hash
     identically at `c3f707a` and at `cf57ea2`, both on `main`. `M-D3` was also
     re-run against the new file, and it still fires `AE3` alone (SPEC-094,
     *Ship addendum*). Not counted: it was caused by the codifying ship itself,
     inside the same cycle.
   - **Still N=1.** It clears at a **second spec that pays** for an unnamed
     or non-durable base, **or** at **N=2 paired-opposing** when a spec other
     than SPEC-086 reproduces a `main`-named row after its target moved. The
     nearest test is SPEC-094's `M-D1` at `c3f707a`, once SPEC-095 has edited
     `docs/api-contract.md`.

3. **`scripts/test-docs.sh` is itself an input to the inventory table it
   guards.** A harness spec is always one assertion id away from moving
   `Documentation assertions (distinct ids)`. Held at **N=2 same-outcome**,
   against a bar of 3.
   - **SPEC-087's `LD5`.** *"Keep both ids"* held the row at 198 as a side
     effect of a coverage argument. Held at N=1 at SPEC-087's ship.
   - **SPEC-088's `AC8`.** The check on the eight-DEC fold was left as a
     hand-run grep, because a new id *"would move the distinct-id row past
     the 200 this spec pins."* A guard went unbuilt to hold a derived number
     steady.
   - **SPEC-086 is a counter-case, and it is recorded rather than counted
     against the candidate.** It added Group `AD`, five new ids, moved the row
     200 → **205**, and regenerated the block instead of arguing the row
     steady. Nothing went unbuilt. What made that cheap is that SPEC-087 had
     already made `Y3` and `Z7` derive, so the move cost zero assertion edits —
     which sharpens the candidate's shape: the pressure appears when a spec
     both *pins* a derived row and *moves* it, not whenever a spec adds an id.
   - **It clears at a third case** of the pressure, not of the move. The rule's
     shape is still open. SPEC-087's build asked for a constraint, and the
     codifying session should choose it.

4. **CLEARED at SPEC-086 ship (2026-09-20), N=3 same-outcome, and written into
   AGENTS.md §12:** *a **no difference** is a measurement, not a default —
   validate the inputs before you believe it.* Kept here for the stage close's
   record, because this is where its third case was argued.
   - **The three cases are one mechanism:** a *no difference* verdict determined
     by something other than the content being compared. **SPEC-082** —
     `git diff --quiet` reported *clean* for an untracked `scripts/coverage.sh`,
     so a real mutant on disk read as a missing one. **SPEC-086 build** — a zsh
     snippet named a variable `path`, which is tied to `PATH`, so every command
     in the loop failed and the comparison reported *0 differences* from `""`
     against `""`. **SPEC-086 verify** — the orchestrator compared `brag story`
     across two binaries with a `--profile` flag that does not exist, and four
     identical `brag: user error: unknown flag: --profile` lines compared equal.
   - **The third case set the sentence.** Its inputs were **non-empty**, so an
     emptiness assert would have passed it; only a plausibility check on the
     output caught it. The clause therefore says *the shape you expected*, of
     which non-emptiness is the weakest instance.
   - **Why SPEC-082 is not a reused case.** Clause (1) of the mutation protocol
     is scoped to mutation probes, and two of these three cases are not mutation
     probes — both are verification equalities across two binaries, which no
     promoted clause reached. SPEC-082's remedy was codified narrowly, as *use
     `shasum` inside the mutation protocol*, and that narrowness is what failed
     to transfer: the project paid the general case twice inside one cycle, in a
     place the special case does not reach. Clause (1) is named in the new clause
     as its **special case**, not its parent.
   - **The zsh quirks themselves are not in §12.** They went to
     `projects/_templates/spec.md`'s new `### Traps` subsection, which had no
     `### Traps` at all before — which is why two consecutive specs wrote one
     from scratch, and why build's Q2 named the Traps list rather than a
     constraint or a decision as the gap.

5. **CLEARED at SPEC-094 ship (2026-09-23), N=2 paired-opposing, and written
   into AGENTS.md §12** as a refinement of *"A mutation pinned by a hash must
   also pin its diff"*: ***a pinned diff has exactly one literal reading.***
   SPEC-094's verify suggested the sentence and deliberately did not propose
   it. Ship decided it, on these grounds.
   - **NEGATIVE: SPEC-094 design's `M-D1`.** Its `old → new` pair occurred
     three times in the target. Build reproduced the behaviour and not the hash
     (`8d7c04ac6af3` against the stated `471eb6137945`), and verify had to hash
     all three readings to recover design's.
   - **POSITIVE: SPEC-088's `M-B3`/`M-B4` and SPEC-086's five abbreviated
     rows.** Each was prose with exactly one literal reading, and each
     reproduced first try. SPEC-086's ship admitted them on exactly that
     property, before anyone had named it as a rule. SPEC-094's verify then
     enforced it mechanically: its probe helper refused its own ambiguous
     `V-N6` anchor (`if n == 1 {`, twice in `bundle.go`) before any gate ran.
   - **Why it clears, and at what bar.** A costly miss and cheap saves, on the
     same surface (a matrix row a later cycle reads to reproduce a hash), from
     different specs, is the meta-rule's paired-opposing N=2. It also meets the
     refinement bar, because the negative exercises the new sentence's own gap:
     `M-D1` met the parent's letter, *the line changed and its replacement
     text*, as an `old → new` pair, and still pinned nothing.
   - **What it absorbs, and what it does not.** It absorbs verify's
     suggestion. It does not absorb item 1 (fidelity) or item 2 (base), which
     stay held, each with its own clearing condition. **SPEC-087's `M-6` is not
     counted**, because it stated no edit, which the parent clause closes.
     SPEC-088's `M-A0` is not a case, because it had one reading.

## Dependencies

### Depends on
- Nothing technical. `internal/storage`, `internal/memory`, `internal/export`
  and the DEC-014 envelope all ship today. Gates green at `1775feb`.

### Routed away from this stage, deliberately
- **Capture completeness** (staging inbox + git-import cold-start miner). Its
  `resume_when` fired on 2026-08-15 and fired again in substance during
  PROJ-007, when a stage close went uncaptured for two days. **Call made at
  framing: it is its own project, not part of this stage.** It is a
  *volume-of-capture* problem; this stage is a *kind-of-capture* problem, and
  the two share no code. Folding them would double this stage and delay the
  honest corpus behind a miner. Recommended as **PROJ-010**, ahead of PROJ-009.
- **Read-path discoverability** (`BRAG.md` is write-only; only a `Stop` hook
  exists, no `SessionStart`). Confirmed on the live tree 2026-09-05: `BRAG.md`
  has 22 headings and **none** is a read-the-corpus section; *"Your role, in one
  sentence"* casts the agent purely as a capturer; `brag memory` appears only
  inside the plugin-install block. **It belongs in PROJ-008 but not in this
  stage** — it should follow the stage that makes the corpus most worth reading,
  so the section written into `BRAG.md` describes a corpus that already holds
  failures. Revisit at STAGE-024 framing.

### Enables
- **STAGE-024 (impact quality)** — with a corpus that contains both shapes, so
  the classifier is calibrated on its whole input domain.
- The mirror, whose sharpest observations (*"three of these went nowhere"*)
  require the corpus to record that they went nowhere.

## Stage-Level Reflection

*Filled in when status moves to shipped.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <one-line items>
- **What can a user do now that they couldn't before, at STAGE scope?** — one
  sentence, before → after; quote the confirming number if one exists, name the
  outcome if not. Write `none` if this stage had no user-visible outcome — a
  real, greppable result, not a blank.

  Not a concatenation of the spec-level answers: those are per-spec, and the
  unit anyone actually records is the stage. Read the spec answers, then say
  what the *stage* bought — the thing that is true now and was not when the
  stage opened. **If this answer is not `none`, capture it before moving status
  to shipped.** Evidence ref: the stage's PRs are known, so tag the one that
  closed it.
  - <answer | none>

---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-087
  type: chore                      # epic | story | task | bug | chore
  cycle: design                    # frame | design | build | verify | ship
  blocked: false
  priority: high                   # sequencing, not size: it should land
                                   # BEFORE SPEC-086 design, which creates
                                   # DEC-050 and would otherwise be the sixth
                                   # consecutive hand re-pin of Y3.
  complexity: S                    # S | M | L  (L means split it)
                                   # RE-CHECKED at design and HELD at S: one
                                   # file changes, +77/-39, no Go, no DEC, no
                                   # user-facing doc, and the inventory table
                                   # comes out byte-identical.

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5
  created_at: 2026-09-06
  framed_at: 2026-09-06
  designed_at: 2026-09-06

insight:
  confidence: 0.9                  # up from framing's 0.75. Framing's open
                                   # question was "what independent source can
                                   # Y3 derive from without going vacuous" —
                                   # settled by measurement, with a six-way
                                   # mutation matrix and five corpus
                                   # simulations. The residual 0.1 is the ONE
                                   # new convention this introduces (a
                                   # tombstone must carry its marker heading),
                                   # which rests on an N=1 population.

references:
  decisions: []                    # THIS SPEC AUTHORS NO DECISION RECORD.
                                   # DEC-050 is spoken for by SPEC-086; taking
                                   # it here would collide on the id. See LD7.
  constraints:
    - one-spec-per-pr
  related_specs:
    - SPEC-086                     # split out of its re-framing, 2026-09-06;
                                   # this spec must land BEFORE its design
    - SPEC-085                     # the fifth consecutive hand re-pin
    - SPEC-084                     # the third; DEC-048 and the golden-length lesson
    - SPEC-082                     # authored Z7 (LD10) — the assertion this
                                   # spec found already half-solved the problem
---

# SPEC-087: `Y3` derives instead of caching

> **Cycle: design.** Complexity **S**, re-checked and held. All three forks
> settled. Every number below was re-derived on 2026-09-06 against `main` at
> **`516dd28`**; every literal was written into a detached `git worktree` and
> run through its own tool — see *§12(b) design-time pre-flight*.
>
> **The pre-flight overturned framing's central premise.** Framing asked
> whether an independent derivation exists *at all*, and treated it as
> genuinely open. It exists, and **half of it was already in the repo**:
> assertion `Z7`, added at SPEC-082, already computes the partition identity
> framing was reaching for. It was measured **blind** — it re-implements
> `inventory.sh`'s two greps as a byte-identical copy, so it cannot catch
> `inventory.sh` going wrong, which is the one thing `Y3` exists for. Fork 1
> is therefore not "invent a second derivation" but "**point the derivation
> that already exists at the right input**."

## What this spec ships

`Y3` stops caching two hand-maintained literals and starts deriving. `Z7`
stops re-deriving `inventory.sh`'s numbers with a copy of its filters and
starts reading the numbers `inventory.sh` actually **emits**.

| | Before | After |
|---|---|---|
| `Y3` | `grep -F` for `Decision records \| 48 \|` and `… reserved … \| 1 \|` | the **emitted** reserved count must equal the tombstone files on disk, counted by a **body-region** marker |
| `Z7` | harness's own copy of `inventory.sh`'s two greps, summed against the file count | `inventory.sh`'s **emitted** two values, summed against the file count |
| Cost of adding a `DEC-*` | **hand-edit a number in `Y3`** (done 5× consecutively) | **zero edits** — measured, simulation S-1 |
| Cost of adding a *tombstone* | zero | must carry `## This is not a decision` (the new convention; simulation S-5) |

**The lesson is closed, not carried a sixth time.**

## Re-measurement at design — every framing premise, checked

Framing's factual claims, re-derived against `516dd28`. Five held; **three
moved**, one of them decisively.

| # | Framing said | Verdict |
|---|---|---|
| 1 | `Decision records` is `48`, `reserved` is `1`; `X3` green, table byte-identical | **HELD.** `./scripts/inventory.sh` re-run; all five gates green (1070 tests / 15 pkgs, 199 `OK:` / 198 distinct ids, lint 0 issues, `gofmt -l` empty, `go vet` clean). |
| 2 | `X3` and `Y3` are not redundant | **HELD, and now proven** rather than asserted — mutations M-4 and M-5 fire each guard through its own oracle only. |
| 3 | `Y3` cannot derive from `inventory.sh` without going vacuous | **HELD, and sharpened.** The rule is narrower than framing stated: `Y3` must not re-derive the *expectation* from `inventory.sh`'s **filters**. Reading `inventory.sh`'s **output** is not only allowed, it is mandatory — that output is the thing under test. Framing conflated input and output. |
| 4 | "The obvious candidate is counting `decisions/DEC-*.md` front-matter directly in the harness, by a different expression" | **MOVED — this is the trap, not the fix.** A second front-matter grep is exactly the §9 failure. And the repo already contains one: `Z7`. Measured blind (Finding 1). |
| 5 | Fork 1 needs a newly invented second derivation | **MOVED.** `Z7` (SPEC-082 LD10) already computes `decisions + reserved == files on disk`. The work is repair, not invention. |
| 6 | Fork 2 would change a user-facing document | **HELD** — and Fork 1's answer delivers Fork 2's benefit *without* that cost, which is what decides the fork. |
| 7 | Files this spec opens: `X3`, `Y3`, `Y4`, `inventory.sh`, `engineering-practices.md` | **MOVED.** `Z7` is in scope and is the centre of the fix; `inventory.sh` and `engineering-practices.md` are **not** modified at all. `Y4` stays out of scope (LD6). |
| 8 | Complexity **S** | **HELD.** One file, +77/−39. |

**A ninth, found by measuring rather than reasoning:** the whole inventory
table comes out **byte-identical** after the change, so no row moves, `X3`
needs no regeneration, and there is no paste step. Framing did not claim
otherwise, but the *Notes for the Implementer* warned about pasting between
the markers as though there would be — there is not. Verified by
`diff <(main inventory.sh) <(fixed worktree inventory.sh)`: empty.

---

## Fork 1 — is there an independent derivation at all? **YES — two, and one already existed**

### The finding that decides the fork

`scripts/test-docs.sh` already contains this, added at SPEC-082 (LD10) and
routed there by STAGE-022's Design Notes:

```sh
z7_files=$(ls decisions/DEC-*.md 2>/dev/null | wc -l | tr -d ' ')
z7_decisions=$(grep -l '^  type: decision' decisions/DEC-*.md 2>/dev/null | wc -l | tr -d ' ')
z7_reserved=$(grep -l '^  type: reservation' decisions/DEC-*.md 2>/dev/null | wc -l | tr -d ' ')
z7_sum=$((z7_decisions + z7_reserved))
```

Those two `grep -l` expressions are **byte-identical** to `inventory.sh:47`
and `inventory.sh:54`. `Z7` therefore does not check `inventory.sh` at all —
it checks a copy of `inventory.sh` against the filesystem. A copy agrees with
the original by construction.

**Measured, not reasoned (mutation M-1).** Broadening `inventory.sh`'s decision
filter from `'^  type: decision'` to `'^  type: '` — the exact failure mode
`Y3`'s own comment names, *"someone edits the type filter in inventory.sh and
it silently starts counting the tombstone as a decision"* — produced:

```
FAIL: X3: inventory block is stale …
FAIL: Y3: inventory.sh row value(s) wrong: decision-records!=48
OK:   Z7                              ← blind
```

`Z7` is the assertion whose entire subject is the decisions rows, and it sat
green through the canonical corruption of the decisions rows. **A green guard
is not evidence its claim is true**, demonstrated on this repo's own harness.

### The fix, and why the two derivations fail differently

The independence framing was looking for does not come from writing a
*cleverer grep*. It comes from **changing what the assertion reads**:

- Both guards now read the two numbers **out of `inventory.sh`'s emitted
  table** (`inv_row`, a new helper). That is the thing under test, and it is
  correct — indeed necessary — for the *subject* of an assertion to be the
  thing under test.
- Each guard checks that subject against an **oracle in a different
  mechanical class from `inventory.sh`'s content filters**:

| Guard | Oracle | Mechanical class | An edit to `inventory.sh`'s front-matter filters… |
|---|---|---|---|
| `Z7` | `ls decisions/DEC-*.md \| wc -l` | **directory entries** — filenames, never file contents | …cannot move it. A filter is a content pattern; the file set is not. |
| `Y3` | `grep -l '^## This is not a decision' decisions/DEC-*.md \| wc -l` | **prose body** — a markdown heading, never front-matter | …cannot move it. The marker is below the `---` fence the filters live above. |

This is the argument framing demanded — *why does the second derivation fail
differently from the first?* Because the two oracles are not greps of the same
region: one reads the directory, one reads the body, and `inventory.sh` reads
the YAML block. A pattern error in the YAML block moves neither oracle, which
is precisely why they can catch it.

### The paired mutation — the acceptance test for this fork

Framing's bar: *"break `inventory.sh`'s filter and confirm `Y3` fires; break
`Y3`'s own derivation and confirm it fires too. A guard that survives both
mutations is vacuous."* Constructed, run, and both halves fire. Full matrix in
§12(b); the two halves of the pair are:

| Half | Mutation | Result |
|---|---|---|
| **break the thing under test** | M-1 broaden `inventory.sh`'s decision filter | `Z7` **FAIL** (`covers 50 of 49`) — where the pre-fix `Z7` was green |
| | M-2 break `inventory.sh`'s reservation filter | `Y3` **FAIL** (`reports 0 … but 1 file(s) carry the marker`), `Z7` **FAIL** |
| **break the guard's own derivation** | M-4 break `Y3`'s body-heading oracle | `Y3` **FAIL** — `X3` and `Z7` stay green, so the failure is attributable |
| | M-5 break `Z7`'s file-glob oracle | `Z7` **FAIL** — `X3` and `Y3` stay green |

Neither guard survives having its own oracle broken. Neither is vacuous.

### Why `Y3` is kept as a second assertion rather than collapsed into `Z7`

Because **M-3 proves it is not redundant.** Swapping `inventory.sh`'s two
filters (`decision` ↔ `reservation`) preserves the sum — 48 + 1 is still 49 —
so `Z7` holds, while the table now reports 1 decision record and 48 reserved
numbers:

```
FAIL: Y3: inventory.sh reports 48 reserved decision number(s), but 1 …file(s)
          carry the tombstone marker '## This is not a decision'.
OK:   Z7
```

`Z7` constrains the two values **jointly** (their sum). `Y3` constrains one of
them **individually**. Collapsing to a single assertion would silently drop
the individual constraint that today's `Y3` literal provides — a coverage
regression dressed as simplification.

### The non-vacuity floor

If `decisions/` ever emptied or the glob broke, `0 + 0 == 0` would pass while
asserting nothing. `Z7` gains an explicit `-lt 1` floor for that, and both
guards **reject a non-numeric or absent row value** rather than comparing empty
strings — measured by M-6, which renames a row label in `inventory.sh` and
fires both guards on the numeric guard rather than passing silently.

### Rejected alternatives

| Alternative | Why rejected |
|---|---|
| **A second front-matter grep in the harness, by "a different expression"** (framing's own suggested candidate) | This is the §9 trap stated in framing and then walked into. Two greps of the same YAML region by the same author fail the same way. **The repo already ran this experiment**: `Z7` is that second grep, and M-1 measured it blind. |
| **Derive `Y3`'s expectation by re-running `inventory.sh`'s filter** | Vacuously true by construction — strictly worse than a stale literal, which at least fails loudly. Framing called this correctly. |
| **Oracle = the filename convention `decisions/DEC-*-reserved-*.md`** | **Measured and wrong: it matches THREE files** — `DEC-027-seed-cost-session-token-reserved-tags.md` and `DEC-049-a-failure-is-a-reserved-type-value-pinned-by-a-verb.md` carry "reserved" in their slugs. A grep-shaped heuristic that reads as authoritative; §9's named failure. Caught only by running it. |
| **Oracle = the `tags: - reservation` front-matter entry** | Correct today (1 hit, `DEC-041`) but lives **inside the front-matter**, the same region `inventory.sh` filters — weaker independence than a body heading for no gain. Also more plausibly attracted to an ordinary decision *about* reservations. |
| **Collapse `Y3` into `Z7`, one assertion** | M-3 shows `Z7` cannot catch a filter swap. Also drops `Documentation assertions (distinct ids)` 198 → 197, moving an inventory row and forcing an `X3` regeneration — a larger blast radius for less coverage. |
| **Assert `reserved < decs`** as the direction check | A heuristic about how rare tombstones are, not an invariant. Exactly the "reads as authoritative" shape §9 warns about. |

---

## Fork 2 — delete `Y3`'s value pin and strengthen `X3` instead? **REJECTED**

Framing's shape: give `inventory.sh` a self-check row (files examined, files
rejected and why) so a filter change moves the rejection count and `X3`'s
byte-for-byte diff catches it. Framing named the cost honestly — *"it changes
the inventory table, which is a user-facing document, for a harness-internal
reason"* — and asked design to weigh it.

**Rejected, because Fork 1's answer delivers the whole benefit at none of the
cost.** The quantity Fork 2 wanted to publish is *"how many `DEC-*.md` files
were examined"*. That is exactly `Z7`'s `z7_files` oracle. Fork 1 keeps it in
the harness, where a harness-internal number belongs, and the reader of
`docs/engineering-practices.md` never sees a row that exists to make a test
work.

Two further measurements against Fork 2, neither available at framing:

1. **`X3` cannot do this job even with the row.** `X3` compares the script's
   output to the page. Both sides are generated from the same filters, so a
   filter change moves both together the moment the page is regenerated — and
   regenerating the page is the *mechanical, always-correct* remedy `X3` itself
   prints on failure. Fork 2 would install a guard whose documented remedy
   silences it. `Z7`'s independent oracle has no such property.
2. **The change is not free even in `X3`'s own terms.** Adding a row moves the
   inventory block, which every spec that regenerates it must then carry.
   Fork 1 leaves the table **byte-identical** (verified by diff), so this spec
   ships without touching `docs/engineering-practices.md` at all.

### Rejected alternatives

| Alternative | Why rejected |
|---|---|
| Fork 2 as framed (a rejection-count row on the page) | Above: user-facing cost, and `X3`'s remedy silences the guard. |
| A self-check row emitted by `inventory.sh` but **excluded** from the `X3` block | Two outputs from one script, only one of them diffed, is a second thing to maintain and a new way for the two to drift. |
| Fork 2 **in addition to** Fork 1 | Redundant once `Z7` reads the emitted values, and it would move an inventory row for zero added coverage. |

---

## Fork 3 — the re-pin-note convention. **MOOT, and formally retired**

Framing: *"Moot if Fork 1 or Fork 2 lands. If neither does, the convention is
the only thing between the next re-pin and a wrong baseline, and it should
become a requirement rather than a habit."*

Fork 1 landed. **There is no next re-pin**, so there is nothing for a re-pin
note to protect. The convention is retired rather than promoted:

- `Y3`'s comment keeps its *history* (46→47 at SPEC-084, 47→48 at SPEC-085) as
  the record of why the pin is gone — that is a dated historical fact, which
  does not rot — but it stops being an instruction to future authors.
- SPEC-085's ship correction established that nothing ever required the note;
  it *"lives only in the comments it has already produced."* This spec removes
  the thing the note was about, which is the only way that ends.
- `Y4` still carries a hand-maintained pin and still carries its re-pin notes.
  Those notes stay, and stay a habit rather than a requirement — see LD6 for
  why `Y4` is out of scope.

### Rejected alternatives

| Alternative | Why rejected |
|---|---|
| Codify the re-pin-note convention in `AGENTS.md` anyway | Framing forbade it and the framing is right: STAGE-023 recorded this at **N=2 same-outcome** and this repo's §12 meta-rule wants N=3. Also pointless — this spec removes the pin the convention protects. |
| Extend the convention to `Y4` as a requirement | `Y4` is out of scope (LD6), and a requirement with no enforcing check is a habit with a stronger adjective. |
| Delete `Y3`'s comment history along with the pin | The history is *why* the guard is shaped this way. Dated facts do not rot; `inventory.sh`'s own header comment states this rule. |

---

## §12(b) design-time pre-flight — what each tool actually said

Every literal this spec embeds was written into a detached `git worktree` at
`516dd28`, run through its own tool, and the result recorded below. Nothing in
this section is predicted. **Four findings were caught here**, one of which
(Finding 1) overturned the fork this spec exists to settle.

| Tool | What was run | Result |
|---|---|---|
| `go test ./...` | full tree, `-count=1` | **15 packages `ok`** + `storagetest` (no test files); **1070** `=== RUN` |
| `just lint` (golangci-lint) | full tree | **0 issues.** |
| `gofmt -l .` | full tree | **empty** |
| `go vet ./...` | full tree | **clean** |
| `bash -n scripts/test-docs.sh` | the spliced script | **syntax OK** |
| `./scripts/test-docs.sh` | full harness, fix applied | **ALL OK**, **199** `OK:` lines / **198** distinct ids — unchanged from `main` |
| `./scripts/inventory.sh` | fixed worktree, **diffed** against `main` | **BYTE-IDENTICAL — zero rows move** |
| `./scripts/inventory.sh` | `main`, live values | `Decision records \| 48 \|`, `… reserved … \| 1 \|` |
| `ls decisions/DEC-*.md \| wc -l` | the `Z7` oracle | **49** — and 48 + 1 = 49, the identity holds |
| `grep -h '^  type: ' decisions/DEC-*.md \| sort \| uniq -c` | type vocabulary, live | `42 decision`, `6 decision  # <comment>`, `1 reservation` = **49** |
| per-file `grep -c '^  type: '` over all 49 | classification is total and unambiguous | **every file has exactly one** — 0 missing, 0 multiple |
| `grep -l '^## This is not a decision'` | the `Y3` oracle | **1** — `DEC-041` exactly |
| `ls decisions/DEC-*-reserved-*.md` | the *rejected* filename oracle | **3** — `DEC-027`, `DEC-041`, `DEC-049` |
| six mutations, `shasum -a 256` before/after | the mutation matrix | **6 of 6 applied and restored**; table below |
| five corpus simulations | added DECs / tombstones / a third type | table below |

### Finding 1 — `Z7` already computes the partition identity, and is blind

The decisive one, described in full under *Fork 1*. Framing treated "is there
an independent derivation at all?" as open and proposed inventing one. The
repo already had it, in the same file, **160 lines below `Y3`** — and it was
wired to the wrong input. Two consequences:

- The fix is **repair, not invention**, which is why complexity holds at S.
- Framing's proposed candidate ("count `DEC-*.md` front-matter directly in the
  harness, by a different expression") is precisely what `Z7` does, and M-1
  measured that it does not work. Had design followed framing's suggestion
  without looking, it would have added a **third** copy of the same grep and
  called the duplication removed.

`Z7` was not in framing's *"Files this spec opens"* list. It was found by
grepping the harness for the string `inventory.sh` rather than for `Y3` — the
§9 half-(b) move (grep for the **value**, not the idea) applied to a
mechanism instead of a number.

### Finding 2 — the obvious oracle for the reservation bucket is wrong

`decisions/DEC-*-reserved-*.md` reads like the natural filename oracle for
"tombstones". It matches **three** files, because `DEC-027`
(`…-reserved-tags`) and `DEC-049` (`…-a-reserved-type-value-…`) carry the word
in their slugs. Had it shipped, `Y3` would have asserted `1 == 3` and failed
immediately — a cheap failure, but only because it was run. The body-heading
oracle was measured to hit exactly `DEC-041`.

### Finding 3 — no inventory row moves, so there is no paste step

Framing's *Notes for the Implementer* carried the standing warning to paste
`just inventory` output between the markers with no blank lines. **It does not
apply to this spec.** The change touches only `scripts/test-docs.sh`, and
neither of the two rows derived from that file moves:
`Documentation assertions (distinct ids)` stays **198** (the `inv_row` helper
takes no assertion id, and both `ok`/`fail` calls keep their literal `"Y3"` /
`"Z7"` ids, which is what `inventory.sh:67` counts statically), and the
`W`-series count stays **6**. Verified by whole-table `diff`: empty. **Build
must not regenerate or paste the inventory block.**

### Finding 4 — two environment traps hit during the pre-flight itself

Recorded because both wasted time and neither is in the spec's warning list:

- **`find` is filtered in this session's shell.** `find . -name "STAGE-023*"`
  returned nothing for a file that exists. Every path lookup in this pre-flight
  used `ls` and `grep -rn` instead. If a build session sees an empty `find`,
  the file is probably there.
- **`cat -A` does not exist on macOS** (BSD `cat`); it errors rather than
  showing line endings. Use `cat -e` or `sed -n l`.

### The mutation matrix

Protocol per §12: content hash before and after, restore from a `/tmp` backup
(never `git checkout`), confirm the hash returns and that the mutant changed
only what was intended. **All six applied and all six restored** — every
`RESTORED` hash matched its pre-mutation value, and every mutant's `diff` was
inspected line-by-line. `inventory.sh` pre/post-restore
`e2db95583a9b…`; `scripts/test-docs.sh` (fixed) `bf94233efe68…`.

| # | Mutation | Target | Mutant hash | `X3` | `Y3` | `Z7` |
|---|---|---|---|---|---|---|
| **M-1** | decision filter broadened `'^  type: decision'` → `'^  type: '` (the failure mode `Y3`'s comment names) | `inventory.sh` | `6e28f6585f0f` | FAIL | ok | **FAIL** |
| **M-1′** | *the same mutation, against `main`'s unfixed harness* | `inventory.sh` | `6e28f6585f0f` | FAIL | FAIL | **ok ← blind** |
| **M-2** | reservation filter broken → `'^  type: reservationX'` | `inventory.sh` | `c820a68e4387` | FAIL | **FAIL** | **FAIL** |
| **M-3** | the two filters **swapped** | `inventory.sh` | `7d37e1ad73d4` | FAIL | **FAIL** | ok |
| **M-4** | `Y3`'s **own** oracle broken (body heading + `X`) | `test-docs.sh` | `437643a8a967` | ok | **FAIL** | ok |
| **M-5** | `Z7`'s **own** oracle narrowed (`DEC-*` → `DEC-04*`) | `test-docs.sh` | `d942e09b4950` | ok | ok | **FAIL** |
| **M-6** | a row **label renamed** in the emitted table | `inventory.sh` | `2eebe0e644ee` | FAIL | **FAIL** | **FAIL** |

Read the matrix three ways:

- **M-1 vs M-1′ is the whole spec in one row.** Same mutant, same hash; the
  pre-fix harness says `OK: Z7`, the post-fix harness says
  `FAIL: Z7: the inventory covers 50 of 49 decisions/DEC-*.md files`.
- **M-4 and M-5 are the "break your own derivation" half.** Each fires exactly
  one guard and leaves the others green, so a failure is attributable to the
  oracle that broke.
- **M-3 is why two assertions survive** rather than one: the sum is preserved,
  `Z7` cannot see it, `Y3` can.
- **M-6 is the vacuity guard.** Without the numeric check, a renamed row would
  make `inv_row` return `""` and both comparisons would be against empty
  strings. Both guards fail loudly instead, naming the renamed row.

*No mutant was discarded.* Every probe applied on the first attempt and
compiled/parsed; `bash -n` was run on the spliced script before any mutation.

### Corpus simulations — what this costs a future spec

Run in the pre-flight worktree by adding a real file to `decisions/` and
running the full harness. **S-1 is the acceptance test for the whole spec.**

| # | Simulation | `X3` | `Y3` | `Z7` | Hand-edits to an assertion |
|---|---|---|---|---|---|
| **S-1** | a new **decision** lands — *simulates SPEC-086 authoring `DEC-050`*; table moves 48 → 49 | FAIL | ok | ok | **ZERO** |
| **S-2** | S-1, then the mechanical remedy (`just inventory`, paste) | ok | ok | ok | **ZERO** — `ALL OK` |
| **S-3** | a DEC lands carrying a **third type** (`analysis`) | FAIL | ok | **FAIL** (`covers 49 of 50`) | zero — the message names both remedies |
| **S-4** | a new **tombstone** lands **with** the marker heading | ok | ok | ok | **ZERO** |
| **S-5** | a new tombstone lands **without** the marker heading | ok | **FAIL** (`reports 2 … but 1 file(s) carry the marker`) | ok | the marker, not a number |

S-1 and S-2 are the answer to framing's question *"what does `Y3` cost a
future spec that adds a `DEC-*`?"* — **nothing.** `X3` fires, and its remedy is
the mechanical regeneration it already prints. That is a generated-artifact
refresh, not a hand-maintained literal, and it is the behaviour every other
row on the page already has.

S-5 is the honest cost, disclosed rather than buried: the spec trades a
**per-decision** hand edit (5 of the last 5 specs paid it) for a
**per-tombstone** convention (paid once in the repo's history, at `DEC-041`).
The failure message names the remedy and the file to copy.

---

## Locked design decisions

**LD1 — `Z7` reads `inventory.sh`'s emitted values, not a copy of its
filters.** The two `grep -l` lines move out of `test-docs.sh`; the two numbers
come from the emitted table via `inv_row`. The oracle (`ls decisions/DEC-*.md
| wc -l`) is unchanged. *Test: M-1, M-1′.*

**LD2 — `Y3` derives from a body-region marker.** The expected value is
`grep -l '^## This is not a decision' decisions/DEC-*.md | wc -l`, compared
against the reserved count `inventory.sh` emits. No literal `48`, no literal
`1`. *Tests: M-2, M-3, M-4, S-5.*

**LD3 — both guards reject a non-numeric or absent row value.** `inv_row`
prints nothing for an unknown label; a bare `[ "$a" -eq "$b" ]` on empty
strings is a shell error under `set -eu`, and a bare string compare would pass
vacuously. Each guard checks `^[0-9]+$` first and fails with the row name.
*Test: M-6.*

**LD4 — `Z7` gains a `-lt 1` non-vacuity floor.** `0 + 0 == 0` must not pass.

**LD5 — two assertion ids are kept, not collapsed to one.** `Z7` constrains the
sum, `Y3` constrains the reserved value individually. *Test: M-3 (`Z7` green,
`Y3` red).* Keeping both also holds `Documentation assertions (distinct ids)`
at 198, so no inventory row moves.

**LD6 — `Y4` is OUT of scope and keeps its literal pin.** Framing scoped it out
*"unless Fork 2 makes it free"*; Fork 2 was rejected, so it is not free. `Y4`
pins `Questions tracked … | 21 |` and `… still open | 8 |` against
`guidance/questions.yaml`, and the equivalent oracle would be a YAML parse of a
register whose entries are hand-written prose — a materially different problem
from counting files in a directory. Taking it here would push this spec past S.
**Routed to SPEC-088** (id verified free) rather than to an anonymous "next
spec" — that is the routing failure STAGE-023 named, and this spec exists
because of it.

**LD7 — this spec authors NO decision record.** `DEC-050` is the next free
number and is **spoken for by SPEC-086's design**. Authoring one here would
collide on the id, and `scripts/_lib.sh:107-119` derives `next_id` from
filenames, so the collision would be real rather than notional. This is a
harness-mechanics repair, not an architectural choice — and framing was
explicit that it is **not a rule to codify** (N=2 same-outcome; §12's
meta-rule wants N=3).

**LD8 — `inventory.sh` and `docs/engineering-practices.md` are NOT modified.**
The inventory table is byte-identical before and after. Build must not paste.

### Rejected alternatives (build-time)

Locked here so build does not re-decide them:

| Build might be tempted to… | Do not, because |
|---|---|
| put `inv_row` next to `Y3` instead of with the other helpers | `Z7` is 160 lines further down and also uses it. It belongs with `assert_*` in the helper block (after `assert_not_contains_iregex`). |
| give `inv_row` an assertion id, or route its failures through `fail` | `inventory.sh:67` counts **literal quoted ids**; a helper that emits one would move the 198. `inv_row` is a pure value extractor and returns nothing on miss — the *callers* decide the verdict. |
| use `grep`/`cut` instead of `awk -F'\|'` to pull the value cell | The label must match the **whole trimmed cell**, not a substring. `Decision records` is a prefix of nothing here today, but `…of those, …` rows and a future row could collide. Exact-cell match is the point. |
| collapse the numeric check and the comparison into one `if` | The failure messages differ and both are load-bearing: one says *the row is gone*, the other says *the number is wrong*. M-6 vs M-2. |
| "simplify" by deleting `Y3` now that `Z7` is fixed | M-3. See LD5. |
| regenerate the inventory block "to be safe" | It is byte-identical; a regeneration would either be a no-op or a sign something else changed. Finding 3. |
| add a `shellcheck` gate while in the file | Out of scope; CI runs golangci-lint (Go-only) and there is no shell linter. |

---

## Outputs

### New files (0)

None. No `DEC-*`, no new script, no new doc.

### Modified files (1)

| Path | Change | Δ lines |
|---|---|---:|
| `scripts/test-docs.sh` | `inv_row` helper inserted after `assert_not_contains_iregex` (**+18**); `Y3` block replaced (41 → 41); `Z7` block replaced (25 → 45) | **+77 / −39** |

**Totals, measured in the pre-flight tree:** 1 file, `2017 → 2055` lines.
`git diff --numstat` reports exactly `77  39  scripts/test-docs.sh`, and
`git status --short` in the pre-flight worktree shows exactly one modified
path.

**Explicitly NOT modified**, each verified: `scripts/inventory.sh`,
`docs/engineering-practices.md`, `AGENTS.md`, `guidance/questions.yaml`,
`decisions/` (any file), anything under `internal/` or `cmd/`.

### Premise audit (§9), run at design against the repo

**The literal-value case, §9 half-(b) — grep for the value, not the idea.**
The two values this spec removes are `48` and `1`. Run with quoted globs:

```
$ grep -rn "Decision records | 48" scripts/ docs/ AGENTS.md
$ grep -rn "decision-records!=48" scripts/
$ grep -rn "reserved, not yet decided | 1" scripts/ docs/ AGENTS.md
```

| Hit | Verdict |
|---|---|
| `scripts/test-docs.sh` `Y3` — the literal `grep -F` needles and the `!=48` string | **REMOVED.** This is the spec. |
| `scripts/test-docs.sh` `Y3` comment — `"just agree on 48"` | **KEPT, rewritten.** Commentary, not an expected value; the rewritten comment keeps the history (46→47→48) as a dated fact. |
| `docs/engineering-practices.md` inventory block — `| Decision records | 48 |` | **NO CHANGE.** Generated by `X3`'s mechanism; byte-identical after this spec. |
| `projects/**`, `decisions/**` | **NO CHANGE.** Dated planning records. |

**The mechanism case — the move framing missed.** Grepping for the *value*
finds `Y3`. Grepping for the *mechanism* is what finds `Z7`:

```
$ grep -n "inventory.sh" scripts/test-docs.sh
```

Six call sites across `X3`, `Y3`, `Y4`, `Z7`. `Z7` is the one framing's file
list omitted, and it is the centre of the fix. **Generalised heuristic for a
harness spec: grep for the artifact under test, not only for the literal.**

**The additive case:** none — no collection gains a member. No new assertion
id, no new DEC, no new question, no new migration.

**The inversion case:** `Z7`'s *implementation* is inverted (it stops trusting
its own copy of the filters). Its *claim* is unchanged and strengthened, so no
existing test's premise is invalidated. Verified: the full harness is `ALL OK`
with no other assertion touched.

---

## Acceptance Criteria

Numbers to diff against, not prose. Every one was observed in the pre-flight
tree before being written here.

1. **`Y3` contains no numeric literal expectation.**
   `grep -n "Decision records | 48\|decision-records!=48\|reserved, not yet decided | 1" scripts/test-docs.sh`
   returns **0 hits** inside the `Y3` assertion body.
2. **`Z7` no longer greps `decisions/` front-matter.**
   `grep -c "type: decision' decisions" scripts/test-docs.sh` is **0**;
   the same expression in `scripts/inventory.sh` still returns **1**.
3. **`inv_row` exists once** and is defined **before** its first use:
   `grep -n "^inv_row()" scripts/test-docs.sh` reports exactly one line, at a
   **lower** line number than either `y3_reserved=` or `z7_decisions=`.
4. **The harness is green at the same size.** `just test-docs` → `ALL OK`,
   **199** `OK:` lines / **198** distinct ids — both unchanged from `main`.
5. **`bash -n scripts/test-docs.sh`** exits 0.
6. **The inventory table is byte-identical to `main`'s.**
   `diff <(./scripts/inventory.sh) <(git show main:scripts/inventory.sh …)` —
   in practice: `X3` is green **without** the block being re-pasted, and
   `git status --short` shows `docs/engineering-practices.md` **unmodified**.
7. **Exactly one file is modified.** `git diff --numstat` reports one row:
   `77  39  scripts/test-docs.sh`. `scripts/test-docs.sh` is **2055** lines.
8. **All five gates green:** `go test ./...` 15 pkgs ok / **1070** tests ·
   `just test-docs` ALL OK · `just lint` **0 issues** · `gofmt -l .` empty ·
   `go vet ./...` clean.
9. **The six mutations reproduce the §12(b) matrix**, including **M-1′**
   against `main` (`OK: Z7`) versus M-1 against the fix (`FAIL: Z7`).
10. **Simulation S-1 costs zero hand-edits**: dropping a `DEC-050` stub with
    `type: decision` into `decisions/` leaves `Y3` and `Z7` **green** and fires
    only `X3`; after regenerating the block, `ALL OK`.
11. **No `DEC-*` file is created.** `ls decisions/DEC-*.md | wc -l` is **49**,
    and `DEC-050` does not exist — it stays free for SPEC-086.
12. **`Y4` is untouched.** Its two literals (`| 21 |`, `| 8 |`) and its re-pin
    notes are byte-identical to `main`.

---

## Failing Tests

This is a harness spec: the "tests" are the harness's own assertions plus the
mutation probes that give them teeth. **Every negative is paired with a
positive**, per §12.

### Changed — `Y3` (assertion id retained)

| Check | Asserts | Fails before the change because |
|---|---|---|
| `Y3` **positive** | `inv_row(reserved) == count('^## This is not a decision')` → `1 == 1` | *Passes* on `main` too — the pin is currently correct. The positive alone proves nothing; **M-2/M-3/M-4 are what give it meaning.** |
| `Y3` **numeric guard** | the reserved row's value matches `^[0-9]+$` | on `main` there is no such guard: a renamed row makes `grep -F` miss and the message says *"row value(s) wrong"*, blaming the number rather than the rename (M-6) |

### Changed — `Z7` (assertion id retained)

| Check | Asserts | Fails before the change because |
|---|---|---|
| `Z7` **positive** | `inv_row(decisions) + inv_row(reserved) == ls decisions/DEC-*.md` → `48 + 1 == 49` | passes on `main` as well — but on `main` the two addends come from the harness's own copy of the filters, so the sum is **structurally incapable** of disagreeing with `inventory.sh`. **M-1′ is the proof.** |
| `Z7` **floor** | `files >= 1` | absent on `main`: an empty `decisions/` would give `0 + 0 == 0` and pass |
| `Z7` **numeric guard** | both row values match `^[0-9]+$` | absent on `main` (M-6) |

### Mutation checks (run by build, recorded in Build Completion)

All seven ran at design and are tabulated in §12(b) with hashes. Build re-runs
them and pastes the output.

| # | Mutation | Expected |
|---|---|---|
| **M-1** | `inventory.sh` decision filter → `'^  type: '` | `Z7` **red** (`covers 50 of 49`), `X3` red, `Y3` green |
| **M-1′** | the same, on a pristine `main` worktree | `Z7` **green** — the regression this spec removes. Run it, or the fix has no evidence. |
| **M-2** | `inventory.sh` reservation filter → `reservationX` | `Y3` **red** and `Z7` **red** |
| **M-3** | swap the two `inventory.sh` filters | `Y3` **red**, `Z7` **green** (proves LD5) |
| **M-4** | `Y3`'s body-heading oracle → `…decisionX` | `Y3` **red**, `X3`/`Z7` green |
| **M-5** | `Z7`'s glob → `DEC-04*` | `Z7` **red** (`covers 49 of 10`), `X3`/`Y3` green |
| **M-6** | rename `Decision numbers reserved, not yet decided` in `inventory.sh` | `Y3` **red** and `Z7` **red**, both naming the *row*, not the number |

**Confirm every mutant by content hash** (`shasum -a 256` before and after);
`git diff --quiet` is blind to untracked files. **Restore from a `/tmp` backup,
never `git checkout`** — `git checkout` would discard build's own uncommitted
edits to `scripts/test-docs.sh`, which is the single file this spec touches.
Confirm the hash returns to its pre-mutation value and that the mutant changed
**only** what you meant. Run `bash -n scripts/test-docs.sh` after each restore.

### Simulation checks (run by build)

| # | Simulation | Expected |
|---|---|---|
| **S-1** | add a `DEC-050` stub with `type: decision` | `Y3`/`Z7` **green**, `X3` red. **Zero assertion edits** — the headline claim. |
| **S-3** | change that stub to `type: analysis` | `Z7` **red** (`covers 49 of 50`) |
| **S-5** | change it to `type: reservation` **without** the marker heading | `Y3` **red**, naming the marker and `DEC-041` |

Delete the stub afterwards and confirm `ls decisions/DEC-*.md | wc -l` is **49**.

### Decision-to-test mapping (§9)

Every locked decision has a check that fails without it.

| Decision | Check |
|---|---|
| LD1 `Z7` reads emitted values | **M-1 / M-1′** — the paired opposing outcome |
| LD2 `Y3` derives from the body marker | M-2, M-3, M-4; S-5 |
| LD3 numeric/absent-row guard | M-6 |
| LD4 `-lt 1` floor | AC-3's ordering + the floor branch; no live mutant (an empty `decisions/` cannot be simulated without deleting 49 tracked files — **stated, not tested**) |
| LD5 two ids, not one | M-3, plus AC-4's 198 distinct ids |
| LD6 `Y4` out of scope | AC-12 — `Y4` byte-identical to `main` |
| LD7 no DEC authored | AC-11 — `DEC-050` does not exist |
| LD8 `inventory.sh` / the page untouched | AC-6, AC-7 |

**One honest gap, disclosed:** LD4's floor has no executed mutant. Simulating
it means emptying `decisions/`, which would delete 49 tracked files, and a
probe whose blast radius exceeds the thing it proves is not worth running. The
branch is three lines and was read, not run. Build may prove it cheaply by
pointing `z7_files`' glob at an empty directory instead — that is M-5's shape
with a different target, and is **optional**.

---

## Implementation Context

### Decisions that apply

None directly. `Z7`'s claim originates in **SPEC-082 LD10** (a hard fail, not a
warning tier) and is preserved verbatim in the rewritten comment — build must
not soften it. No `DEC-*` is referenced or created (LD7).

### Constraints that apply

- `one-spec-per-pr` — one branch, one PR, this spec only.
- No Go code is touched, so `no-sql-in-cli-layer` and the testing conventions
  in §9 do not bite. `just test` is unaffected; the gate that matters is
  `just test-docs`.

### Prior related work

- **SPEC-082 LD10** — authored `Z7`, including the correct oracle
  (`ls decisions/DEC-*.md`) and the correct failure message. The defect this
  spec repairs is only its *input*, not its design.
- **SPEC-080** — deferred the sum check; **STAGE-022 Design Notes** routed it
  to `Z7`.
- **SPEC-081…085** — the five consecutive hand re-pins of `Y3`.
- **SPEC-085 ship** — corrected verify's diagnosis of why the `Y3` re-pin note
  was missing, and established the note convention is nowhere a requirement.
  That correction is what makes Fork 3 retirable rather than promotable.
- **STAGE-023** — named SPEC-086 as owner; SPEC-086's re-framing split this
  file out rather than absorbing it.

### Out of scope

- **`Y4`** (the open-questions pin) → **SPEC-088**. See LD6.
- `scripts/inventory.sh` — the thing under test. Not modified (LD8).
- `docs/engineering-practices.md` — byte-identical (Finding 3).
- Anything under `internal/` or `cmd/`. This is a harness spec.
- Codifying anything into `AGENTS.md`. N=2 same-outcome; the repo wants N=3.

### Sequencing — the deadline this spec exists to beat

**Land this before SPEC-086 design.** SPEC-086 authors `DEC-050`, moving
`Decision records` 48 → 49. If `Y3` still caches a literal at that moment,
SPEC-086 becomes the **sixth** consecutive hand re-pin — and the first under a
*named* owner, which removes "nobody owned it" as the explanation and makes the
result materially worse than the current N=5. Simulation **S-1** is exactly
that event, run in advance: with this spec landed it costs zero assertion
edits.

---

## Notes for the Implementer

**Literal-artifact-as-spec.** All three blocks below were written into the
pre-flight worktree, run through `bash -n`, the full harness, six mutations and
five simulations, and are reproduced here **verbatim**. Transcribe them; do not
paraphrase. Verify diffs against these literals.

### Order of work

1. Insert `inv_row` (§1) directly after the `assert_not_contains_iregex`
   function, before the `# Resolve $1 against $2` comment.
2. Replace the `Y3` block (§2) — comment through closing `fi`.
3. Replace the `Z7` block (§3) — comment through closing `fi`.
4. `bash -n scripts/test-docs.sh`, then `just test-docs`.
5. Run the mutation matrix and the simulations; paste the output into
   *Build Completion*.
6. **Do not run `just inventory`. Do not touch the inventory block.**

Two mechanical rules that bit earlier in this cycle, carried forward:

- **Quote your `--include` globs.** Unquoted `--include=*.sh` is expanded by
  zsh before `grep` runs and the search silently never happens.
  `grep -rn "x" scripts/ docs/` is the safe form.
- **`find` may return nothing for files that exist** in this environment
  (§12(b) Finding 4). Use `ls` and `grep -rn`.

### 1. `scripts/test-docs.sh` — new helper, inserted after `assert_not_contains_iregex`

```sh

# Pull the Value cell out of one `scripts/inventory.sh` table row, addressed by
# its exact What-column label. Used by Y3 and Z7 to check inventory.sh's OWN
# EMITTED numbers against an independent oracle, rather than re-deriving those
# numbers with a copy of inventory.sh's filters — a copy agrees with the
# original by construction, and so cannot catch the original going wrong
# (measured: SPEC-087 mutation M-1, recorded in Z7's comment).
#
# Prints nothing when the label is absent. Callers MUST reject an empty or
# non-numeric result, or a renamed row would make the assertion vacuously true.
inv_row() {
    awk -F'|' -v want="$2" '
        {
            lab = $2; gsub(/^[ \t]+|[ \t]+$/, "", lab)
            if (lab == want) { v = $3; gsub(/^[ \t]+|[ \t]+$/, "", v); print v; exit }
        }
    ' <<<"$1"
}
```

### 2. `scripts/test-docs.sh` — the `Y3` block, replacing lines 1586–1626

```sh
# Y3 — the reservation tombstone is not counted as a decision. DERIVED, not
# pinned: until SPEC-087 this assertion hard-coded `Decision records | 48 |`
# and `Decision numbers reserved, not yet decided | 1 |`, so every spec that
# added a DEC-*.md had to hand-edit a number here. SPEC-081, 082, 083, 084 and
# 085 each did exactly that — five consecutive re-pins. The literal is gone.
#
# WHAT IT ASSERTS NOW: the reserved count inventory.sh EMITS must equal the
# number of tombstone files on disk, counted by a marker in the file BODY
# (`## This is not a decision`) rather than by the front-matter `insight.type`
# filter inventory.sh itself uses. That is the independence that makes this
# non-vacuous: an edit to inventory.sh's front-matter filters cannot move a
# body heading, so the two counts fail differently. Z7 owns the complementary
# half — that the two emitted rows account for every file on disk.
#
# WHY NOT RE-RUN INVENTORY.SH'S OWN FILTER HERE: a copy of the filter agrees
# with the original by construction. Z7 did precisely that until SPEC-087 and
# was measured blind; see mutation M-1 in Z7's comment.
#
# WHY THE FILENAME IS NOT THE ORACLE: `decisions/DEC-*-reserved-*.md` matches
# THREE files (DEC-027, DEC-041, DEC-049), because two ordinary decisions carry
# "reserved" in their slug. Measured at SPEC-087 design and rejected — a
# grep-shaped heuristic that reads as authoritative is the AGENTS.md §9 trap.
#
# WHAT THIS COSTS A FUTURE SPEC: nothing at all when it adds a decision. A spec
# that adds a new TOMBSTONE must give it the `## This is not a decision`
# heading — copy `decisions/DEC-041-*.md`, the only one today. The failure
# message below says so, which is where the convention is discoverable.
if [ ! -x scripts/inventory.sh ]; then
    fail "Y3" "scripts/inventory.sh is missing or not executable"
else
    y3_out=$(./scripts/inventory.sh)
    y3_reserved=$(inv_row "$y3_out" 'Decision numbers reserved, not yet decided')
    y3_tombstones=$(grep -l '^## This is not a decision' decisions/DEC-*.md 2>/dev/null | wc -l | tr -d ' ')
    if ! printf '%s' "$y3_reserved" | grep -qE '^[0-9]+$'; then
        fail "Y3" "inventory.sh emitted no numeric value for the 'Decision numbers reserved, not yet decided' row (got: '$y3_reserved'). The row was renamed or removed — an absent row must not pass silently."
    elif [ "$y3_reserved" -eq "$y3_tombstones" ]; then
        ok "Y3"
    else
        fail "Y3" "inventory.sh reports $y3_reserved reserved decision number(s), but $y3_tombstones decisions/DEC-*.md file(s) carry the tombstone marker '## This is not a decision'. Either inventory.sh's 'insight.type: reservation' filter is wrong, or a tombstone is missing that heading (copy decisions/DEC-041-*.md)."
    fi
fi
```

### 3. `scripts/test-docs.sh` — the `Z7` block, replacing lines 1789–1813

```sh
# Z7 — the two decisions/ rows in the inventory must ADD UP to the files on
# disk. Deferred at SPEC-080, routed here by STAGE-022's Design Notes.
# scripts/inventory.sh counts `insight.type: decision` and
# `insight.type: reservation` separately; a DEC-*.md carrying a third type, or
# none at all, is counted by NEITHER row and vanishes from the page silently —
# while X3 still round-trips green, because the page is generated from the same
# two filters that lost it.
#
# THE MEANING, decided at SPEC-082 LD10: a hard fail. The two readings of an
# untyped file — a typo, or a category inventory.sh has not been taught yet —
# differ in intent, not in consequence: either way the file is invisible and
# the page under-reports. This harness has no warning tier, so the failure
# message names both remedies instead of inventing one.
#
# CHANGED AT SPEC-087 — THE NUMBERS NOW COME FROM INVENTORY.SH'S OUTPUT.
# Until SPEC-087 this assertion re-ran inventory.sh's two greps here, as a
# byte-identical copy of them. A copy agrees with the original by construction,
# so the sum stayed 49 no matter what inventory.sh did. MEASURED, mutation M-1:
# broadening inventory.sh's decision filter to `^  type: ` made it report 49
# decisions and 1 reservation over 49 files — the exact failure Y3 exists to
# catch — and this assertion stayed GREEN while X3 and Y3 both fired. Reading
# the EMITTED values makes the same arithmetic catch it: 49 + 1 != 49.
#
# The `-lt 1` floor is the non-vacuity guard: if the glob ever matched nothing,
# 0 + 0 == 0 would pass while asserting nothing at all.
#
# decisions/_template.md is deliberately outside every count (it does not match
# DEC-*.md) and stays outside this one.
if [ ! -x scripts/inventory.sh ]; then
    fail "Z7" "scripts/inventory.sh is missing or not executable"
else
    z7_files=$(ls decisions/DEC-*.md 2>/dev/null | wc -l | tr -d ' ')
    z7_out=$(./scripts/inventory.sh)
    z7_decisions=$(inv_row "$z7_out" 'Decision records')
    z7_reserved=$(inv_row "$z7_out" 'Decision numbers reserved, not yet decided')
    if ! printf '%s' "$z7_decisions" | grep -qE '^[0-9]+$' || ! printf '%s' "$z7_reserved" | grep -qE '^[0-9]+$'; then
        fail "Z7" "inventory.sh emitted no numeric value for 'Decision records' (got: '$z7_decisions') and/or for 'Decision numbers reserved, not yet decided' (got: '$z7_reserved'). A renamed or removed row must not pass silently."
    elif [ "$z7_files" -lt 1 ]; then
        fail "Z7" "no decisions/DEC-*.md files matched — the glob found nothing, which would make the sum check vacuously true"
    elif [ $((z7_decisions + z7_reserved)) -eq "$z7_files" ]; then
        ok "Z7"
    else
        fail "Z7" "the inventory covers $((z7_decisions + z7_reserved)) of $z7_files decisions/DEC-*.md files ($z7_decisions decision + $z7_reserved reservation). A DEC-*.md with a missing or unknown 'insight.type' is counted by neither row: fix its front-matter, or teach scripts/inventory.sh a row for the new type."
    fi
fi
```

---

## Build Completion

*Filled at build.*

### Build-phase reflection (3 questions, short answers)

## Reflection (Ship)

- **What can a user do now that they couldn't before?** — one sentence,
  before → after. Capture this before closing the cycle.

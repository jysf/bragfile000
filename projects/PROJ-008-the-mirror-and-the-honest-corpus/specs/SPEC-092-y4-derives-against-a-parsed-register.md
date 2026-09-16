---
# Maps to ContextCore task.* semantic conventions.
# This variant assumes Claude plays every role. The context normally
# in a separate handoff doc lives in the ## Implementation Context
# section below.

task:
  id: SPEC-092
  type: chore                      # epic | story | task | bug | chore
  cycle: frame                     # frame | design | build | verify | ship
                                   # Created at SPEC-088 framing (2026-09-15) to
                                   # CLAIM the id in the same edit as the
                                   # sentence that routes work here. It carries
                                   # SPEC-088's framing verdict — GO, S, with the
                                   # measurement — and still owes its own framing
                                   # pass for the one fork named below.
  blocked: false
  priority: medium                 # `Y4`'s pin is CORRECT today. It costs only
                                   # the next spec that files a question.
  complexity: S                    # one file, one assertion, no Go, no user-
                                   # facing table row. See ## Complexity.

project:
  id: PROJ-008
  stage: STAGE-023
repo:
  id: bragfile

agents:
  architect: claude-opus-5
  implementer: claude-opus-5       # usually same Claude, different session
  created_at: 2026-09-15

insight:
  confidence: 0.85                 # the oracle question is answered by
                                   # measurement; the dependency choice is a
                                   # judgement its own design pass should make.

references:
  decisions: []                    # emits none unless design picks an
                                   # interpreter dependency — then a DEC, and the
                                   # next free number is DEC-054, NOT DEC-050.
  constraints:
    - one-spec-per-pr
    - no-new-top-level-deps-without-decision   # severity: warning; scoped to
                                               # go.mod/go.sum, so it does not
                                               # formally bind a shell harness
                                               # dependency — but its rationale
                                               # ("dependencies are forever")
                                               # is the argument design must
                                               # answer. See ## The one fork.
  related_specs:
    - SPEC-087                     # made `Y3` derive; left `Y4` pinned (LD6) and
                                   # left this open question behind
    - SPEC-088                     # split this out at framing, 2026-09-15
    - SPEC-080                     # authored `Y4`'s pin and the questions rows
---

# SPEC-092: `Y4` derives against a parsed register

## Context

> **Cycle: frame.** Created at SPEC-088's framing to claim the id with a file
> rather than reserve it in prose. **The routing verdict is GO at S**, with the
> open question from SPEC-087 LD6 **answered by measurement** below. What this
> file does *not* yet settle is the one fork named at the end; that is its own
> framing pass.
>
> Measured 2026-09-15 against `main` at `6dda56a`.

`Y4` is the last assertion in `scripts/test-docs.sh` that caches a derived
number. It pins two of `scripts/inventory.sh`'s emitted rows as literals:

```
printf '%s\n' "$y4_out" | grep -F -q 'Questions tracked in guidance/questions.yaml | 21 |' || …
printf '%s\n' "$y4_out" | grep -F -q 'of those, still open | 8 |' || …
```

Its comment records **five re-pins across four specs** (18/6 → 19/7 → 19/6 →
20/7 → 21/8), each one a hand-edit paid by a spec whose real subject was
something else. `Y3` stopped doing this at SPEC-087; `Y4` did not, on
measurement rather than preference — the equivalent oracle for a hand-written
YAML register was a materially different problem from counting files in a
directory, and taking it inside SPEC-087 would have pushed that spec past S.

**The open question SPEC-087 left (LD6): is there an independent oracle for a
hand-written YAML register that is not just `inventory.sh`'s two `grep -cE`
expressions copied?**

### Answer: yes — a real YAML parse. Measured, not asserted.

`inventory.sh` derives both numbers by matching *lines* at an exact indent:
`grep -cE '^  - id: '` and `grep -cE '^    status: open$'`. An oracle that
builds a **document model** and counts *entries* and *field values* is in a
different mechanical class by construction: an indentation change, a quoted
scalar, a duplicated id, a commented-out entry or a second YAML document moves
one and not the other. That is exactly the independence `Y3` gets from reading a
body heading where `inventory.sh` reads front-matter, and `Z7` gets from reading
what the script *emits*.

Both interpreters that ship on the two platforms this repo targets parse the
register today and reproduce both numbers — **21 questions, 8 open**:

```
$ /usr/bin/ruby -ryaml -e 'q=YAML.load_file("guidance/questions.yaml")["questions"];
    puts "total=#{q.size} open=#{q.count{|x| x["status"]=="open"}}"'
  total=21 open=8                      # ruby 2.6.10, /usr/bin/ruby, macOS
$ /usr/bin/python3 -c 'import yaml; q=yaml.safe_load(open("guidance/questions.yaml"))["questions"];
    print(len(q), sum(1 for x in q if x["status"]=="open"))'
  21 8                                 # Python 3.9.6 + PyYAML 6.0.3, /usr/bin/python3
```

The parse also surfaces two facts the greps cannot see, both of which are
candidate assertions in their own right: all 21 ids are **unique**, all 21
entries **have** a `status`, and the register carries **three** status values
(`open` 8, `answered` 7, `resolved` 6) where the file's own header comment
documents only `open | investigating | answered` — so `resolved` is an
undocumented value used by 6 of 21 entries.

**A harness dependency on an external tool has direct precedent.**
`scripts/test-docs.sh:15-19` already hard-requires `jq` and exits 2 with an
install pointer when it is missing. This is not a new class of decision for this
file; it is the second instance of one already made.

## The one fork this spec still owes

**Which oracle, and what happens when its tool is absent?**

- **(i) A YAML parse via a system interpreter** (`ruby`, or `python3` with
  PyYAML), hard-required like `jq`. Strongest independence. Cost: a second
  interpreter dependency on a shell harness, and `/usr/bin/ruby` is 2.6.10 —
  long deprecated by Apple and a plausible removal in a future macOS.
- **(ii) The same parse, but `skip`-ing when the tool is absent.** `test-docs.sh`
  has a `skip()` and S3 uses it. Cheaper politically, weaker as a guard: a
  skipped assertion is green and proves nothing, which is the vacuity failure
  `Y3` was repaired for.
- **(iii) A structural `awk` oracle** — count entries by their sequence marker
  and cross-check that every entry has exactly one `status`, asserting the
  partition (`total == open + answered + resolved`) rather than re-running the
  producer's filters. No dependency at all; independence is real but thinner,
  since it is still line-matching. `AB4` already reads `questions.yaml` with
  `awk` this way, so the idiom exists in the file.

Design decides, with the §12(b) pre-flight run against whichever tool it picks.

## Complexity

**S.** One file modified (`scripts/test-docs.sh`), no Go, no user-facing doc, no
decision record unless fork (i) is chosen. **`Y4` keeps its id**, so the
`Documentation assertions (distinct ids)` row does **not** move — this spec, by
design, leaves the inventory table byte-identical, which is the same profile
SPEC-087 held S on.

## GO / NO-GO

**GO** at **S**, with no sequencing claim on SPEC-086.

**Not a blocker for SPEC-086's design.** `Y4` pins only the two
`guidance/questions.yaml` rows, and SPEC-086 moves `Decision records` (DEC-050)
and `Documentation assertions`, neither of which `Y4` reads. **The one condition
that would change that:** if SPEC-086's design files a question in
`guidance/questions.yaml`, it moves both of `Y4`'s pins and pays the sixth
hand-re-pin. If that comes up, this spec is cheap to pull forward.

**Why it is not folded into SPEC-088.** Not size — the dependency. SPEC-088 is
two assertions with no external tool and no decision record; this one may need
both, and `no-new-top-level-deps-without-decision`'s rationale ("dependencies are
forever, force deliberation") deserves its own review rather than a paragraph
inside a spec about something else.

## Prior related work

- **SPEC-087** (shipped) — `Y3` derives. Its Fork 1 is the template for this
  work: *find an oracle in a different mechanical class from the producer's own
  filter*, and its verify cycle is where the non-vacuity floor was learned.
- **SPEC-080** — authored the questions rows and `Y4`'s pin.
- **SPEC-088** — split this out; its `## GO / NO-GO` records why.

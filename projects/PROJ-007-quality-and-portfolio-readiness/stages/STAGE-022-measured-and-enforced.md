---
# Maps to ContextCore epic-level conventions.
# A Stage is a coherent chunk of work within a Project.
# It has a spec backlog and ships as a unit when the backlog is done.

stage:
  id: STAGE-022
  status: proposed                  # proposed | active | shipped | cancelled | on_hold
  priority: medium
  target_complete: null

project:
  id: PROJ-007
repo:
  id: bragfile

created_at: 2026-08-15
shipped_at: null
---

# STAGE-022: quality measured and enforced, known defects closed

## What This Stage Is

The standard signals a reviewer greps for exist and **gate CI** rather than
being available to run: a lint configuration, a coverage number with an honest
floor, and an empty known-defect list.

Two halves, both small:

- **Measured and enforced** — `golangci-lint` configured and gating,
  `go test -cover` reporting a number, a badge, and a floor chosen from the
  measured value rather than aspiration.
- **The known-defect list emptied** — three defects are currently *recorded as
  real and deliberately not fixed*, each because it needs a decision rather
  than a patch. Carrying them past this project would mean the quality pass
  left the honest list of open bugs untouched, which is the one outcome that
  would make the whole project read as theatre.

## Why Now

Verified 2026-08-15, all four gaps still hold: no `golangci-lint` config, no
`-cover` anywhere, `func Benchmark` count **0**, no surfaced godoc. CI runs
`gofmt`, `go vet` and `go test` and nothing else.

**Second, not first** — see STAGE-021's *Why Now*. Lint and coverage are table
stakes; the discipline page is the differentiated half, and it establishes what
a coverage number is allowed to claim. Reversing the order invites a percentage
with no story attached, which is the exact failure mode SPEC-073 hit four times.

## Success Criteria

- `golangci-lint` runs in CI and **fails the build**, with a config whose
  enabled linters were each chosen, not inherited from a default set.
- Coverage is measured, reported, and carries a **floor that the current number
  actually clears** — set from measurement, and stated as a floor rather than a
  target.
- The coverage claim on the practices page is **mutation-checked**: break
  something the tests supposedly pin and watch the check fail, then confirm the
  mutant actually mutated. (SPEC-073's coverage sentence was wrong four times;
  SPEC-078's H9 was green, had teeth, and still pinned the wrong proposition.)
- `guidance/questions.yaml` no longer lists `tag-ordering-projection` as open —
  it is answered by a DEC, and `MergeTags` is fixed in the same change.
- `$EDITOR` handling has a **recorded quoting rule**, and both
  `EDITOR="code -w"` and an editor path containing a space work.
- `brag memory`'s header no longer describes its candidate pool using the same
  word its five sibling exporters use for entries-in-scope.

## Scope

### In scope

1. **`golangci-lint`** — config, per-linter rationale, CI gate.
2. **Coverage** — `-cover` in CI, a reported number, a badge, an honest floor.
3. **The three coupled defects**, each needing a decision recorded:
   - **`MergeTags` position dup** (`storage/store.go`) — the graft copies
     `s.position` verbatim, so after a merge an object's tag positions gap or
     collide. Blocked on the open `tag-ordering-projection` question. Resolve
     the question in the same change (see Design Notes for the recommendation).
   - **`$EDITOR` with spaces** (`editor/launch.go:76`) — `strings.Fields(v)` is
     simultaneously what makes `EDITOR="code -w"` work and what breaks
     `EDITOR="/Applications/My Editor/bin/edit"`. Needs a chosen rule.
   - **The `Entries:` envelope inconsistency** — six exporters emit that header
     line; five use `len(entries)`, and `memory` alone uses `result.Candidates`,
     the pool capped at `PoolLimit=200`. On a 368-entry corpus the header reads
     `Entries: 200`, which a reader takes for the corpus size. Changing it moves
     byte-exact goldens **and MCP resource output** (the resources are
     byte-identical to `brag memory`), so it is an envelope decision, not a
     label edit.

### Explicitly out of scope

- **Benchmarks and any scale/perf or concurrency harness.** Deferred to
  **PROJ-009** with the crustyimg methodology harvest that it depends on. This
  keeps PROJ-007 short and leaves room for what comes up.
- Heavy performance optimization — never this project's game.
- Raising coverage by writing tests to hit a number. Measure what is there; if
  the number is embarrassing, say so on the practices page rather than gaming it.
- Any feature work.

## Spec Backlog

Ordered list of specs composing this stage. IDs assigned at creation.

Format: `- [status] SPEC-ID (cycle) — one-line summary`

- [ ] (not yet framed) — **lint + coverage in CI.** golangci-lint config with
      per-linter rationale, `-cover`, badge, measured floor; the coverage claim
      mutation-checked before it is written down.
- [ ] (not yet framed) — **the three coupled defects.** Emits two DECs
      (tag-ordering/`MergeTags`, `$EDITOR` quoting) and settles the `Entries:`
      envelope semantics; each fix lands with a regression test.

**Count:** 0 shipped / 0 active / 2 pending

## Design Notes

- **Recommended answer for `tag-ordering-projection`: keep `position`.** Tag
  order is user-visible today — `brag show 374` renders tags in the order they
  were supplied (topic tags, then `agent:`/`model:`/`session:`/`cost:`/
  `tokens:`, then evidence tags), which is insertion order, not name-ASC.
  Dropping the column would silently reorder every entry's display. So the
  question resolves to *preserve original order*, and `MergeTags`' fix follows
  from it: after grafting, **renumber each affected taggable's positions densely
  from 0 in current position order**, with a deterministic tie-break. That
  removes gaps and collisions without inventing a new ordering semantic.
- **Recommended rule for `$EDITOR`: try the whole string as a path first, then
  split.** If the entire value names an existing executable, run it with no
  arguments; otherwise fall back to `strings.Fields`. Both real cases then work
  with no configuration and no quoting rule for the user to learn:
  `/Applications/My Editor/bin/edit` stats successfully and is used whole;
  `code -w` does not stat and is split. The failure mode — a full command line
  that happens to name an existing file — is vanishingly unlikely. Rejected
  alternative: shell-style quoting, which needs a parser or a dependency and
  pushes a quoting rule onto the user; also rejected: invoking a shell, which
  this codebase deliberately avoids.
- **Pick linters deliberately.** A default `golangci-lint` enable-all produces a
  wall of findings that gets silenced with `//nolint`, which is worse than no
  lint. Enable a small set, justify each, and let the config itself be
  legible — it is going to be read by the same person the practices page is for.
- **The coverage number is not the deliverable; the honest floor is.** A repo
  with 78 test files against 69 source files does not need to defend a
  percentage. State the number, set the floor below it, and explain in the
  practices page what the tests actually pin — which STAGE-021 will have already
  worked out.

## Dependencies

### Depends on
- **STAGE-021** — the practices page establishes what the coverage number is
  presented against, and what the test regime genuinely guarantees.
- Prior art, not a blocker: crustyimg already wired lint and coverage. Read its
  config as reference when framing the first spec. The **formal methodology
  harvest** (harness structure, perf approach) travels with the deferred
  benchmark work to PROJ-009.

### Enables
- **PROJ-009** — a scale/perf and concurrency baseline, on a substrate whose
  quality is already measured and whose known-defect list is empty.

## Stage-Level Reflection

*Filled in when status moves to shipped. Run Prompt 1c (Stage Ship) in
FIRST_SESSION_PROMPTS.md to draft this.*

- **Did we deliver the outcome in "What This Stage Is"?** <yes/no + notes>
- **How many specs did it actually take?** <number vs. plan>
- **What changed between starting and shipping?** <one sentence>
- **Lessons that should update AGENTS.md, templates, or constraints?**
  - <one-line updates>
- **Should any spec-level reflections be promoted to stage-level lessons?**
  - <one-line items>

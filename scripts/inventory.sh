#!/usr/bin/env bash
# scripts/inventory.sh — print the engineering-practices inventory table.
#
# SINGLE SOURCE for two consumers:
#   1. the table embedded in docs/engineering-practices.md (paste this output
#      between the `inventory:begin` / `inventory:end` markers), and
#   2. test-docs assertion X3, which re-runs this script and diffs its output
#      against that block — so a number on the page cannot drift from the repo.
#
# Run via `just inventory`.
#
# THE RULE THE PAGE FOLLOWS: every number that tracks the repo's CURRENT state
# is produced here. A number stated inline on the page is only ever a dated
# historical fact (a version that shipped without a compare-link, a figure that
# was corrected) — those do not rot. If you find yourself wanting to type a
# current-state number into the prose, add a row here instead.
#
# TWO RULES THIS SCRIPT ITSELF HOLDS:
#
#   (a) Every count is scoped to explicit tracked directories. NEVER `find .`:
#       when a background agent is running, .claude/worktrees/<name>/ holds a
#       full second copy of the repo, and an unscoped walk silently doubles
#       every number. Same trap already documented at test-docs E2.
#
#   (b) The doc-assertion count is DISTINCT ASSERTION IDS, derived statically
#       from the script text — not a count of `OK:` lines from a run. Two
#       reasons: running test-docs.sh from here would recurse (X3 runs this
#       script), and the OK-line count is environment-dependent — S3 emits an
#       OK per check when the `claude` CLI is installed and OK+SKIP when it is
#       not, so `just test-docs | grep -c '^OK:'` prints one more on a machine
#       with the Claude CLI than without. Distinct ids is the only number a
#       stranger cloning the repo reproduces.

set -eu

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
cd "$REPO_ROOT"

n() { printf '%s' "$1" | tr -d ' '; }

# Confidence values, one per line, numeric, trailing comments stripped.
confidences() {
    grep -h '^  confidence:' decisions/DEC-*.md | sed 's/#.*//' | awk '{print $2+0}'
}

decs=$(n "$(grep -l '^  type: decision' decisions/DEC-*.md 2>/dev/null | wc -l)")
decs_superseded=$(n "$(grep -l '^superseded_by: DEC-' decisions/DEC-*.md 2>/dev/null | wc -l)")
decs_amended=$(n "$(grep -l '^## Amendment' decisions/DEC-*.md 2>/dev/null | wc -l)")
# A tombstone (insight.type: reservation, e.g. DEC-041) holds a number
# without deciding anything — SPEC-080. Counted separately so it does not
# silently inflate "Decision records", and so its own existence is a derived
# number too, not a fact only readable in prose.
decs_reserved=$(n "$(grep -l '^  type: reservation' decisions/DEC-*.md 2>/dev/null | wc -l)")
conf_min=$(confidences | sort -g | head -1)
conf_max=$(confidences | sort -g | tail -1)
conf_certain=$(n "$(confidences | awk '$1 >= 1.0' | wc -l)")
projects=$(n "$(ls -d projects/PROJ-*/ 2>/dev/null | wc -l)")
stages=$(n "$(ls projects/*/stages/STAGE-*.md 2>/dev/null | wc -l)")
specs_done=$(n "$(ls projects/*/specs/done/SPEC-*.md 2>/dev/null | wc -l)")
specs_build_reflection=$(n "$(grep -l '^### Build-phase reflection' projects/*/specs/done/SPEC-*.md 2>/dev/null | wc -l)")
src=$(n "$(find internal cmd -name '*.go' ! -name '*_test.go' | wc -l)")
tst=$(n "$(find internal cmd -name '*_test.go' | wc -l)")
testfuncs=$(n "$(grep -rh '^func Test' internal cmd --include='*_test.go' | wc -l)")
benchmarks=$(n "$(grep -rh '^func Benchmark' internal cmd --include='*_test.go' | wc -l)")
wseries=$(n "$(grep -cE '^# W[0-9]+ ' scripts/test-docs.sh)")
docasserts=$(n "$(grep -oE '(^|[[:space:]])(ok|fail|skip|assert_[a-z_]+) "[A-Za-z0-9][A-Za-z0-9._-]*"' \
    scripts/test-docs.sh | grep -oE '"[A-Za-z0-9][A-Za-z0-9._-]*"' | tr -d '"' | sort -u | wc -l)")

# Open-questions hygiene (SPEC-080). STAGE-021's own count was wrong three
# times in three days — once from a plain `grep -c 'status: open'` that
# matched this file's OWN header comment (`#   - status: open |
# investigating | answered`). Anchored to the exact 4-space indent real
# entries use, which the `#`-prefixed header comment can never match.
questions_total=$(n "$(grep -cE '^  - id: ' guidance/questions.yaml 2>/dev/null || true)")
questions_open=$(n "$(grep -cE '^    status: open$' guidance/questions.yaml 2>/dev/null || true)")

cat <<EOF
| What | Value | Where it lives |
|---|---:|---|
| Decision records | ${decs} | \`decisions/DEC-*.md\` (\`insight.type: decision\`) |
| …of those, superseded by a later record | ${decs_superseded} | \`superseded_by:\` in the front-matter |
| …of those, carrying an explicit \`## Amendment\` section | ${decs_amended} | \`decisions/DEC-*.md\` |
| Decision numbers reserved, not yet decided | ${decs_reserved} | \`decisions/DEC-*.md\` (\`insight.type: reservation\`) |
| Lowest confidence value on a decision record | ${conf_min} | \`insight.confidence\` in the front-matter |
| Highest confidence value on a decision record | ${conf_max} | \`insight.confidence\` in the front-matter |
| Decision records claiming confidence 1.0 | ${conf_certain} | \`insight.confidence\` in the front-matter |
| Projects | ${projects} | \`projects/PROJ-*/brief.md\` |
| Stages | ${stages} | \`projects/*/stages/STAGE-*.md\` |
| Specs carried to ship and archived | ${specs_done} | \`projects/*/specs/done/\` |
| …of those, also carrying a build-phase reflection | ${specs_build_reflection} | \`### Build-phase reflection\` in those files |
| Go source files | ${src} | \`internal/\`, \`cmd/\` |
| Go test files | ${tst} | \`internal/\`, \`cmd/\` |
| Go test functions | ${testfuncs} | \`func Test*\` in \`*_test.go\` |
| Documentation assertions (distinct ids) | ${docasserts} | \`scripts/test-docs.sh\`, run by \`just test-docs\` |
| …of those, replacing a manual release-checklist item | ${wseries} | the \`W\`-series in \`scripts/test-docs.sh\` |
| Questions tracked in guidance/questions.yaml | ${questions_total} | \`guidance/questions.yaml\` |
| …of those, still open | ${questions_open} | \`status: open\` in the same file |
| Benchmarks | ${benchmarks} | none exist — see "What this does not measure" |
EOF

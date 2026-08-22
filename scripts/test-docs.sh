#!/usr/bin/env bash
# scripts/test-docs.sh — documentation-content assertions for the
# bragfile repo. Exits 0 iff all assertions pass.
#
# Run via `just test-docs`. Not wired into `just test` (Go-only).

set -eu

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
cd "$REPO_ROOT"

FAIL_COUNT=0

# Group H asserts require jq for JSON-shape parsing.
if ! command -v jq >/dev/null 2>&1; then
    printf 'test-docs: jq is required but not installed (see https://stedolan.github.io/jq/)\n' >&2
    exit 2
fi

ok() {
    printf 'OK:   %s\n' "$1"
}

fail() {
    printf 'FAIL: %s: %s\n' "$1" "$2"
    FAIL_COUNT=$((FAIL_COUNT + 1))
}

skip() {
    printf 'SKIP: %s: %s\n' "$1" "$2"
}

# --- helpers ---

assert_file_exists() {
    name="$1"; path="$2"
    if [ -f "$path" ]; then
        ok "$name"
    else
        fail "$name" "file does not exist: $path"
    fi
}

assert_line_count_band() {
    name="$1"; path="$2"; min="$3"; max="$4"
    if [ ! -f "$path" ]; then
        fail "$name" "file does not exist: $path"
        return 0
    fi
    n=$(wc -l < "$path" | tr -d ' ')
    if [ "$n" -ge "$min" ] && [ "$n" -le "$max" ]; then
        ok "$name"
    else
        fail "$name" "$path has $n lines (expected $min..$max)"
    fi
}

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

assert_contains_literal() {
    name="$1"; path="$2"; pattern="$3"
    if [ ! -f "$path" ]; then
        fail "$name" "file does not exist: $path"
        return 0
    fi
    if grep -F -q -- "$pattern" "$path"; then
        ok "$name"
    else
        fail "$name" "$path missing literal: $pattern"
    fi
}

assert_cmd_ok() {
    name="$1"; shift
    if "$@" >/dev/null 2>&1; then
        ok "$name"
    else
        fail "$name" "command failed: $*"
    fi
}

assert_not_contains_iregex() {
    name="$1"; path="$2"; pattern="$3"
    if [ ! -f "$path" ]; then
        fail "$name" "file does not exist: $path"
        return 0
    fi
    if grep -i -E -q -- "$pattern" "$path"; then
        hit=$(grep -i -n -E -- "$pattern" "$path" | head -n 1)
        fail "$name" "$path contains forbidden pattern: $pattern (first hit: $hit)"
    else
        ok "$name"
    fi
}

# Resolve $1 against $2 (source file's dir) and check existence.
# Strips #anchor, skips http/https/mailto and bare anchors.
check_link_target() {
    src="$1"; src_dir="$2"; link="$3"
    target=$(printf '%s' "$link" | sed 's/#.*$//')
    case "$target" in
        http://*|https://*|mailto:*) return 0 ;;
        '') return 0 ;;
    esac
    if [ "$src_dir" = "." ]; then
        resolved="$target"
    else
        resolved="$src_dir/$target"
    fi
    if [ ! -e "$resolved" ]; then
        fail "E1" "$src: link \"$link\" → \"$resolved\" not found"
    fi
}

# ===== Group A — README shape (positive) =====

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

# A2 — README opens with H1 in line 1 or 2
if [ ! -f README.md ]; then
    fail "A2" "README.md does not exist"
elif head -n 2 README.md | grep -E -q '^# '; then
    ok "A2"
else
    fail "A2" "no '# ' heading in first 2 lines of README.md"
fi

# A3 — Above-the-fold is user-facing
if [ ! -f README.md ]; then
    fail "A3" "README.md does not exist"
else
    above=$(head -n 12 README.md)
    has_brag=no; has_uf=no; has_forbidden=no
    if printf '%s\n' "$above" | grep -i -q 'brag'; then has_brag=yes; fi
    if printf '%s\n' "$above" | grep -i -E -q 'capture|retrieve|accomplishment|retro|review|resume'; then has_uf=yes; fi
    if printf '%s\n' "$above" | grep -i -E -q 'spec-driven|architect|implementer|reviewer|cycle|hierarchy'; then has_forbidden=yes; fi
    if [ "$has_brag" = yes ] && [ "$has_uf" = yes ] && [ "$has_forbidden" = no ]; then
        ok "A3"
    else
        fail "A3" "above-the-fold gate (brag=$has_brag user-facing-word=$has_uf forbidden-token=$has_forbidden)"
    fi
fi

# A4 — Install section with both paths
if [ ! -f README.md ]; then
    fail "A4" "README.md does not exist"
else
    has_heading=no; has_brew=no; has_local=no
    if grep -E -q '^## .*[Ii]nstall' README.md; then has_heading=yes; fi
    if grep -F -q 'brew install jysf/tap/bragfile' README.md; then has_brew=yes; fi
    if grep -F -q 'go install ./cmd/brag' README.md || grep -F -q 'just install' README.md; then
        has_local=yes
    fi
    if [ "$has_heading" = yes ] && [ "$has_brew" = yes ] && [ "$has_local" = yes ]; then
        ok "A4"
    else
        fail "A4" "install section (heading=$has_heading brew=$has_brew local=$has_local)"
    fi
fi

# A5 — Workflow-demo command coverage (all 7 brag verbs in fenced shell blocks)
if [ ! -f README.md ]; then
    fail "A5" "README.md does not exist"
else
    fenced=$(awk '/^```/{f=!f; next} f' README.md)
    missing=""
    for cmd in "brag add" "brag list" "brag search" "brag export" "brag summary" "brag review" "brag stats"; do
        if ! printf '%s\n' "$fenced" | grep -F -q -- "$cmd"; then
            missing="$missing $cmd"
        fi
    done
    if [ -z "$missing" ]; then
        ok "A5"
    else
        fail "A5" "missing in fenced blocks:$missing"
    fi
fi

# A6 — Where-data-lives reference
assert_contains_literal "A6" "README.md" "~/.bragfile/db.sqlite"

# A7 — Tutorial pointer
assert_contains_literal "A7" "README.md" "docs/tutorial.md"

# A8 — BRAG.md pointer
assert_contains_literal "A8" "README.md" "BRAG.md"

# A9 — CONTRIBUTING.md pointer
assert_contains_literal "A9" "README.md" "CONTRIBUTING.md"

# A10 — License section
if [ ! -f README.md ]; then
    fail "A10" "README.md does not exist"
else
    has_heading=no; has_mit=no
    if grep -E -q '^## [Ll]icense' README.md; then has_heading=yes; fi
    if grep -F -q 'MIT' README.md; then has_mit=yes; fi
    if [ "$has_heading" = yes ] && [ "$has_mit" = yes ]; then
        ok "A10"
    else
        fail "A10" "license section (heading=$has_heading mit=$has_mit)"
    fi
fi

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

# ===== Group B — README shape (negative — load-bearing) =====

# B1 — No `spec-driven` token
assert_not_contains_iregex "B1" "README.md" 'spec-driven'

# B2 — No cycle phrase (any of three forms)
if [ ! -f README.md ]; then
    fail "B2" "README.md does not exist"
else
    hit=""
    if grep -i -E -q 'frame.*design.*build.*verify.*ship' README.md; then
        hit="${hit} regex-form"
    fi
    if grep -i -F -q 'frame → design' README.md; then
        hit="${hit} unicode-arrow-form"
    fi
    if grep -i -F -q 'frame -> design' README.md; then
        hit="${hit} ascii-arrow-form"
    fi
    if [ -z "$hit" ]; then
        ok "B2"
    else
        fail "B2" "cycle phrase present:$hit"
    fi
fi

# B3 — No `four habits` phrase
assert_not_contains_iregex "B3" "README.md" 'four habits'

# B4 — No `context contamination` phrase
assert_not_contains_iregex "B4" "README.md" 'context contamination'

# B5 — No contributor-shaped just-recipe refs
assert_not_contains_iregex "B5" "README.md" 'just (new-spec|advance-cycle|archive-spec|weekly-review|new-stage)'

# B6 — No `Claude plays every role` phrase
assert_not_contains_iregex "B6" "README.md" 'claude plays every role'

# B7-heading — No `## … table of contents` heading
assert_not_contains_iregex "B7-heading" "README.md" '^## .*table of contents'

# B7-toc — No 4+ contiguous `- [` lines in first 50 lines
if [ ! -f README.md ]; then
    fail "B7-toc" "README.md does not exist"
else
    streak=$(head -n 50 README.md | awk '
        /^- \[/ { s += 1; if (s > max) max = s; next }
        { s = 0 }
        END { print (max ? max : 0) }
    ')
    if [ "$streak" -lt 4 ]; then
        ok "B7-toc"
    else
        fail "B7-toc" "found contiguous run of $streak '- [' lines in first 50 lines (TOC block)"
    fi
fi

# ===== Group C — CONTRIBUTING.md =====

assert_file_exists "C1" "CONTRIBUTING.md"
assert_line_count_band "C2" "CONTRIBUTING.md" 30 120
assert_contains_literal "C3" "CONTRIBUTING.md" "docs/development.md"
assert_contains_literal "C4" "CONTRIBUTING.md" "AGENTS.md"

# C5 — Setup commands: just install AND just test
if [ ! -f CONTRIBUTING.md ]; then
    fail "C5" "CONTRIBUTING.md does not exist"
else
    has_install=no; has_test=no
    if grep -F -q 'just install' CONTRIBUTING.md; then has_install=yes; fi
    if grep -F -q 'just test' CONTRIBUTING.md; then has_test=yes; fi
    if [ "$has_install" = yes ] && [ "$has_test" = yes ]; then
        ok "C5"
    else
        fail "C5" "setup commands (just install=$has_install just test=$has_test)"
    fi
fi

# ===== Group D — docs/development.md =====

assert_file_exists "D1" "docs/development.md"
assert_line_count_band "D2" "docs/development.md" 50 200

# D3 — Hierarchy present (Repo + Project + Stage + Spec)
if [ ! -f docs/development.md ]; then
    fail "D3" "docs/development.md does not exist"
else
    missing=""
    for tok in Repo Project Stage Spec; do
        if ! grep -i -F -q -- "$tok" docs/development.md; then
            missing="$missing $tok"
        fi
    done
    if [ -z "$missing" ]; then
        ok "D3"
    else
        fail "D3" "hierarchy tokens missing:$missing"
    fi
fi

# D4 — Cycle phrase present (Unicode-arrow form, exact substring)
assert_contains_literal "D4" "docs/development.md" "Frame → Design → Build → Verify → Ship"

assert_contains_literal "D5" "docs/development.md" "AGENTS.md"
assert_contains_literal "D6" "docs/development.md" "GETTING_STARTED.md"
assert_contains_literal "D7" "docs/development.md" "FIRST_SESSION_PROMPTS.md"

# D8 — Glossary cross-link: AGENTS.md mention within ±5 lines of 'glossary'
if [ ! -f docs/development.md ]; then
    fail "D8" "docs/development.md does not exist"
else
    agents_lines=$(grep -n -F 'AGENTS.md' docs/development.md | cut -d: -f1)
    glossary_lines=$(grep -n -i -F 'glossary' docs/development.md | cut -d: -f1)
    if [ -z "$agents_lines" ] || [ -z "$glossary_lines" ]; then
        fail "D8" "missing AGENTS.md mention or 'glossary' mention"
    else
        min_diff=999999
        for a in $agents_lines; do
            for g in $glossary_lines; do
                d=$(( a > g ? a - g : g - a ))
                if [ "$d" -lt "$min_diff" ]; then
                    min_diff=$d
                fi
            done
        done
        if [ "$min_diff" -le 5 ]; then
            ok "D8"
        else
            fail "D8" "closest AGENTS.md/glossary line distance is $min_diff (>5)"
        fi
    fi
fi

# ===== Group E — Link integrity =====

# E1 — Internal links resolve in README, CONTRIBUTING, development.md
e1_pre_count=$FAIL_COUNT
for src in README.md CONTRIBUTING.md docs/development.md docs/engineering-practices.md; do
    [ -f "$src" ] || continue
    src_dir=$(dirname "$src")
    # Extract every ](url) and strip surrounding markers
    while IFS= read -r raw; do
        [ -n "$raw" ] || continue
        link=$(printf '%s' "$raw" | sed -E 's/^]\((.*)\)$/\1/')
        check_link_target "$src" "$src_dir" "$link"
    done <<EOF
$(grep -oE '\]\([^)]+\)' "$src" || true)
EOF
done
if [ "$FAIL_COUNT" -eq "$e1_pre_count" ]; then
    ok "E1"
fi

# E2 — docs/development.md only referenced by this spec's outputs
# NOTE on the exclude list (see AGENTS.md §9): BSD grep matches --exclude-dir
# against BASENAMES, so `--exclude-dir=docs/reports` is a silent no-op — it is
# retained only as intent. The correctness boundary is the `case` whitelist
# below. `.claude` IS a valid basename exclude and is load-bearing: background
# agent sessions create git worktrees under .claude/worktrees/<name>/, which is
# a full second copy of the repo. Without this, E2 scans that copy and fails on
# the worktree's own CONTRIBUTING.md — a FAIL that depends on whether an agent
# happens to be running, which is the worst kind of flake to hand a verify
# session. .claude/ is gitignored, so nothing in it is repo content.
hits=$(grep -rn -F 'docs/development.md' . \
    --include='*.md' \
    --exclude-dir=projects \
    --exclude-dir=node_modules \
    --exclude-dir=.git \
    --exclude-dir=.claude \
    --exclude-dir=framework-feedback \
    --exclude-dir=docs/reports 2>/dev/null || true)
unexpected=""
if [ -n "$hits" ]; then
    while IFS= read -r line; do
        [ -n "$line" ] || continue
        path=$(printf '%s' "$line" | cut -d: -f1)
        case "$path" in
            ./README.md|./CONTRIBUTING.md|./docs/development.md) ;;
            *) unexpected="${unexpected}\n  $line" ;;
        esac
    done <<EOF
$hits
EOF
fi
if [ -z "$unexpected" ]; then
    ok "E2"
else
    fail "E2" "unexpected references to docs/development.md:$(printf '%b' "$unexpected")"
fi

# E3 — CONTRIBUTING.md is brand-new (no prior deletion in git history)
prior=$(git log --all --diff-filter=D --pretty=format:%H -- CONTRIBUTING.md 2>/dev/null || true)
if [ -z "$prior" ]; then
    ok "E3"
else
    fail "E3" "prior deletion(s) of CONTRIBUTING.md in git history: $prior"
fi

# ===== Group F — Just-recipe wiring =====

# F1 — `test-docs` recipe defined
if [ ! -f justfile ]; then
    fail "F1" "justfile does not exist"
elif grep -E -q '^test-docs:' justfile; then
    ok "F1"
else
    fail "F1" "no '^test-docs:' recipe in justfile"
fi

# F2 — `test:` recipe unchanged from pre-spec form (header + `    @go test ./...`)
if [ ! -f justfile ]; then
    fail "F2" "justfile does not exist"
else
    actual=$(awk '/^test:$/{f=1; print; next} f && /^$/{exit} f{print}' justfile)
    expected="$(printf 'test:\n    @go test ./...')"
    if [ "$actual" = "$expected" ]; then
        ok "F2"
    else
        fail "F2" "test: recipe diverged from pre-spec form"
    fi
fi

# F3 — scripts/test-docs.sh executable + POSIX-headed shebang
if [ ! -f scripts/test-docs.sh ]; then
    fail "F3" "scripts/test-docs.sh does not exist"
else
    is_exec=no; shebang_ok=no
    if [ -x scripts/test-docs.sh ]; then is_exec=yes; fi
    if head -n 1 scripts/test-docs.sh | grep -E -q '^#!(/usr/bin/env (sh|bash)|/bin/sh)'; then
        shebang_ok=yes
    fi
    if [ "$is_exec" = yes ] && [ "$shebang_ok" = yes ]; then
        ok "F3"
    else
        fail "F3" "executable=$is_exec posix-shebang=$shebang_ok"
    fi
fi

# ===== Group G — Harness ergonomics =====

# G3 — Works from any cwd (verified by SCRIPT_DIR pattern at top of script)
if grep -q 'SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)' scripts/test-docs.sh; then
    ok "G3"
else
    fail "G3" "scripts/test-docs.sh missing SCRIPT_DIR resolution pattern"
fi

# G2 — Exit-code contract is built-in (FAIL_COUNT-driven exit at the bottom)
ok "G2"

# ===== Group H — JSON Schema shape =====

SCHEMA_PATH="docs/brag-entry.schema.json"

# H1 — Schema file exists
assert_file_exists "H1" "$SCHEMA_PATH"

# H2 — Schema is valid JSON
if [ -f "$SCHEMA_PATH" ]; then
    if jq -e . "$SCHEMA_PATH" >/dev/null 2>&1; then
        ok "H2"
    else
        fail "H2" "$SCHEMA_PATH is not valid JSON"
    fi
else
    fail "H2" "$SCHEMA_PATH does not exist"
fi

# Helper for jq-based equality checks against the schema. Compares
# the jq-extracted value against an expected literal string.
assert_jq_eq() {
    name="$1"; expr="$2"; expected="$3"
    if [ ! -f "$SCHEMA_PATH" ]; then
        fail "$name" "$SCHEMA_PATH does not exist"
        return 0
    fi
    actual=$(jq -r "$expr" "$SCHEMA_PATH" 2>/dev/null || echo "<jq-error>")
    if [ "$actual" = "$expected" ]; then
        ok "$name"
    else
        fail "$name" "$expr returned \"$actual\" (expected \"$expected\")"
    fi
}

# H3 — Schema declares draft 2020-12
assert_jq_eq "H3" '."$schema"' "https://json-schema.org/draft/2020-12/schema"

# H4 — Schema declares object type at root
assert_jq_eq "H4" '.type' "object"

# H5 — Schema requires title
if [ -f "$SCHEMA_PATH" ]; then
    if jq -e '.required | index("title")' "$SCHEMA_PATH" >/dev/null 2>&1; then
        ok "H5"
    else
        fail "H5" '"title" not found in .required array'
    fi
else
    fail "H5" "$SCHEMA_PATH does not exist"
fi

# H6 — Schema disallows additional properties
assert_jq_eq "H6" '.additionalProperties' "false"

# H7 — Title is non-empty string
if [ -f "$SCHEMA_PATH" ]; then
    title_type=$(jq -r '.properties.title.type' "$SCHEMA_PATH" 2>/dev/null || echo "")
    title_min=$(jq -r '.properties.title.minLength' "$SCHEMA_PATH" 2>/dev/null || echo "")
    if [ "$title_type" = "string" ] && [ "$title_min" = "1" ]; then
        ok "H7"
    else
        fail "H7" "properties.title.type=\"$title_type\" minLength=\"$title_min\" (want type=\"string\" minLength=\"1\")"
    fi
else
    fail "H7" "$SCHEMA_PATH does not exist"
fi

# H8 — Tags is string (NOT array) — DEC-004 alignment, load-bearing
assert_jq_eq "H8" '.properties.tags.type' "string"

# H9 — All nine DEC-011 keys present in properties
if [ -f "$SCHEMA_PATH" ]; then
    h9_missing=""
    for key in title description tags project type impact id created_at updated_at; do
        if ! jq -e ".properties.$key" "$SCHEMA_PATH" >/dev/null 2>&1; then
            h9_missing="$h9_missing $key"
        fi
    done
    if [ -z "$h9_missing" ]; then
        ok "H9"
    else
        fail "H9" "missing properties:$h9_missing"
    fi
else
    fail "H9" "$SCHEMA_PATH does not exist"
fi

# H10 — Schema declares canonical $id URL
assert_jq_eq "H10" '."$id"' "https://github.com/jysf/bragfile000/blob/main/docs/brag-entry.schema.json"

# ===== Group I — Hook script shape =====

HOOK_PATH="scripts/claude-code-post-session.sh"

# I1 — Hook script exists
assert_file_exists "I1" "$HOOK_PATH"

# I2 — Hook script is executable
if [ -x "$HOOK_PATH" ]; then
    ok "I2"
else
    fail "I2" "$HOOK_PATH is not executable (chmod +x)"
fi

# I3 — Hook script has POSIX shebang on line 1
if [ ! -f "$HOOK_PATH" ]; then
    fail "I3" "$HOOK_PATH does not exist"
elif head -n 1 "$HOOK_PATH" | grep -E -q '^#!(/usr/bin/env (sh|bash)|/bin/sh)'; then
    ok "I3"
else
    fail "I3" "$HOOK_PATH missing POSIX shebang on line 1"
fi

# I4 — Hook script references `brag add --json`
assert_contains_literal "I4" "$HOOK_PATH" "brag add --json"

# I5 — Hook script references `jq`
assert_contains_literal "I5" "$HOOK_PATH" "jq"

# ===== Group J — Slash-command template shape =====

SLASH_PATH="examples/brag-slash-command.md"

# J1 — Template file exists
assert_file_exists "J1" "$SLASH_PATH"

# J2 — Template is tight (5–30 lines)
assert_line_count_band "J2" "$SLASH_PATH" 5 30

# J3 — Template references the schema
assert_contains_literal "J3" "$SLASH_PATH" "docs/brag-entry.schema.json"

# J4 — Template references `brag add --json`
assert_contains_literal "J4" "$SLASH_PATH" "brag add --json"

# ===== Group K — BRAG.md cross-reference =====

# K1 — BRAG.md references the schema
assert_contains_literal "K1" "BRAG.md" "docs/brag-entry.schema.json"

# K2 — BRAG.md references the hook script
assert_contains_literal "K2" "BRAG.md" "scripts/claude-code-post-session.sh"

# K3 — BRAG.md references the slash-command template
assert_contains_literal "K3" "BRAG.md" "examples/brag-slash-command.md"

# K4 — BRAG.md has a JSON-contract section heading
if [ ! -f BRAG.md ]; then
    fail "K4" "BRAG.md does not exist"
elif grep -E -q '^## .*JSON' BRAG.md; then
    ok "K4"
else
    fail "K4" "no '## … JSON …' heading in BRAG.md"
fi

# ===== Group L — .goreleaser.yaml shape =====

GORELEASER="${REPO_ROOT}/.goreleaser.yaml"

# L1 — config file exists
assert_file_exists "L1" "$GORELEASER"

# L2 — opens with `version: 2`
if [ ! -f "$GORELEASER" ]; then
    fail "L2" "$GORELEASER does not exist"
elif head -n 5 "$GORELEASER" | grep -E -q '^version:[[:space:]]+2[[:space:]]*$'; then
    ok "L2"
else
    fail "L2" "$GORELEASER does not declare 'version: 2' in first 5 lines"
fi

# L3 — declares CGO_ENABLED=0
assert_contains_literal "L3" "$GORELEASER" "CGO_ENABLED=0"

# L4 — declares both darwin and linux goos values
assert_contains_literal "L4a" "$GORELEASER" "- darwin"
assert_contains_literal "L4b" "$GORELEASER" "- linux"

# L5 — declares both amd64 and arm64 goarch values
assert_contains_literal "L5a" "$GORELEASER" "- amd64"
assert_contains_literal "L5b" "$GORELEASER" "- arm64"

# L6 — declares a top-level `brews:` block (binary formula, not a cask — DEC-040)
if [ ! -f "$GORELEASER" ]; then
    fail "L6" "$GORELEASER does not exist"
elif grep -E -q '^brews:[[:space:]]*$' "$GORELEASER"; then
    ok "L6"
else
    fail "L6" "$GORELEASER does not declare a top-level 'brews:' block"
fi

# L7 — brews block points at the shared homebrew-tap (DEC-040)
assert_contains_literal "L7" "$GORELEASER" "name: homebrew-tap"

# L8 — brews block has `skip_upload: auto`
assert_contains_literal "L8" "$GORELEASER" "skip_upload: auto"

# L9 — declares `-X main.version=` ldflag
assert_contains_literal "L9" "$GORELEASER" "-X main.version="

# L10 — archive format is `tar.gz` (goreleaser v2 list form)
assert_contains_literal "L10" "$GORELEASER" "formats: [tar.gz]"

# L11 — brews block installs the `brag` binary onto $PATH
assert_contains_literal "L11" "$GORELEASER" 'bin.install "brag"'

# ===== Group M — .github/workflows/ci.yml shape =====

CI_WORKFLOW=".github/workflows/ci.yml"

# M1 — file exists
assert_file_exists "M1" "$CI_WORKFLOW"

# M2 — triggers on pull_request
assert_contains_literal "M2" "$CI_WORKFLOW" "pull_request:"

# M3 — triggers on push to main (push: stanza followed by branches: + main)
if [ ! -f "$CI_WORKFLOW" ]; then
    fail "M3" "$CI_WORKFLOW does not exist"
elif awk '
        /^on:/ { in_on=1; next }
        in_on && /^[a-zA-Z]/ { in_on=0 }
        in_on && /push:/ { in_push=1; next }
        in_push && /branches:/ { in_branches=1; next }
        in_branches && /- main/ { print "ok"; exit }
    ' "$CI_WORKFLOW" | grep -q ok; then
    ok "M3"
else
    fail "M3" "$CI_WORKFLOW does not trigger on push to main"
fi

# M4 — matrix includes macos-latest
assert_contains_literal "M4" "$CI_WORKFLOW" "macos-latest"

# M5 — matrix includes ubuntu-latest
assert_contains_literal "M5" "$CI_WORKFLOW" "ubuntu-latest"

# M6 — runs `go test ./...`
assert_contains_literal "M6" "$CI_WORKFLOW" "go test ./..."

# M7 — runs `gofmt -l .`
assert_contains_literal "M7" "$CI_WORKFLOW" "gofmt -l ."

# M8 — runs `go vet ./...`
assert_contains_literal "M8" "$CI_WORKFLOW" "go vet ./..."

# M9 — uses actions/setup-go (SHA-pinned; version in comment)
assert_contains_literal "M9" "$CI_WORKFLOW" "actions/setup-go@"

# ===== Group N — .github/workflows/release.yml shape =====

RELEASE_WORKFLOW=".github/workflows/release.yml"

# N1 — file exists
assert_file_exists "N1" "$RELEASE_WORKFLOW"

# N2 — triggers on tag push pattern v*
if [ ! -f "$RELEASE_WORKFLOW" ]; then
    fail "N2" "$RELEASE_WORKFLOW does not exist"
elif grep -E -q "^[[:space:]]+- 'v\\*'" "$RELEASE_WORKFLOW"; then
    ok "N2"
else
    fail "N2" "$RELEASE_WORKFLOW does not trigger on tag pattern 'v*'"
fi

# N3 — uses goreleaser/goreleaser-action (SHA-pinned; version in comment)
assert_contains_literal "N3" "$RELEASE_WORKFLOW" "goreleaser/goreleaser-action@"

# N4 — passes HOMEBREW_TAP_GITHUB_TOKEN env
assert_contains_literal "N4" "$RELEASE_WORKFLOW" "HOMEBREW_TAP_GITHUB_TOKEN"

# N5 — checkout uses fetch-depth: 0
assert_contains_literal "N5" "$RELEASE_WORKFLOW" "fetch-depth: 0"

# N6 — job declares timeout-minutes: 30
assert_contains_literal "N6" "$RELEASE_WORKFLOW" "timeout-minutes: 30"

# N7 — workflow declares concurrency block
assert_contains_literal "N7" "$RELEASE_WORKFLOW" "concurrency:"

# ===== Group O — CHANGELOG.md shape =====

# O1 — file exists
assert_file_exists "O1" "CHANGELOG.md"

# O2 — references Keep-A-Changelog
assert_contains_literal "O2" "CHANGELOG.md" "keepachangelog.com"

# O3 — has `## [0.1.0]` heading (line-based equality avoids substring trap)
if [ ! -f CHANGELOG.md ]; then
    fail "O3" "CHANGELOG.md does not exist"
elif grep -E -q '^## \[0\.1\.0\]' CHANGELOG.md; then
    ok "O3"
else
    fail "O3" "CHANGELOG.md missing '## [0.1.0]' heading"
fi

# O4 — lists each shipped command verb under Added (single named
# assertion that iterates internally over the ten verbs)
o4_failed=""
for verb in "brag add" "brag list" "brag show" "brag edit" "brag delete" \
            "brag search" "brag export" "brag summary" "brag review" "brag stats"; do
    if ! grep -F -q -- "\`${verb}\`" CHANGELOG.md; then
        o4_failed="${o4_failed} ${verb}"
    fi
done
if [ -z "$o4_failed" ]; then
    ok "O4"
else
    fail "O4" "CHANGELOG.md missing command refs:$o4_failed"
fi

# O5 — has [Unreleased] and [0.1.0] link reference definitions
o5_failed=""
for ref in "[Unreleased]:" "[0.1.0]:"; do
    if ! grep -F -q -- "$ref" CHANGELOG.md; then
        o5_failed="${o5_failed} ${ref}"
    fi
done
if [ -z "$o5_failed" ]; then
    ok "O5"
else
    fail "O5" "CHANGELOG.md missing link refs:$o5_failed"
fi

# ===== Group P — Doc sweep + tense flips =====

# P1 — README.md status banner does NOT contain "in progress"
assert_not_contains_iregex "P1" "README.md" "in progress"

# P2 — README.md does NOT contain "in active development"
assert_not_contains_iregex "P2" "README.md" "in active development"

# P3 — AGENTS.md does NOT contain "arriving in STAGE"
assert_not_contains_iregex "P3" "AGENTS.md" "arriving in STAGE"

# P4 — AGENTS.md does NOT contain literal `(STAGE-004)`
assert_not_contains_iregex "P4" "AGENTS.md" "\\(STAGE-004\\)"

# P5 — docs/architecture.md does NOT contain "sqlite-file-copy"
assert_not_contains_iregex "P5" "docs/architecture.md" "sqlite-file-copy"

# P6 — docs/architecture.md does NOT contain "Distribution (STAGE-004)"
assert_not_contains_iregex "P6" "docs/architecture.md" "Distribution \\(STAGE-004\\)"

# P7 — docs/tutorial.md §9 body does NOT contain `brew install`
# (sectioned slice from the §9 heading to the next ## or --- divider)
if [ ! -f docs/tutorial.md ]; then
    fail "P7" "docs/tutorial.md does not exist"
else
    section_body=$(awk '
        /^## 9\./ { in_section=1; next }
        in_section && /^## / { in_section=0 }
        in_section && /^---[[:space:]]*$/ { in_section=0 }
        in_section { print }
    ' docs/tutorial.md)
    if printf '%s' "$section_body" | grep -F -q "brew install"; then
        fail "P7" "docs/tutorial.md §9 body contains 'brew install'"
    else
        ok "P7"
    fi
fi

# P8 — cmd/brag/main.go does NOT contain literal STAGE-004
assert_not_contains_iregex "P8" "cmd/brag/main.go" "STAGE-004"

# P9 — BRAG.md still references SPEC-022 artifacts (regression check)
p9_failed=""
for art in "docs/brag-entry.schema.json" \
           "scripts/claude-code-post-session.sh" \
           "examples/brag-slash-command.md"; do
    if ! grep -F -q -- "$art" BRAG.md; then
        p9_failed="${p9_failed} ${art}"
    fi
done
if [ -z "$p9_failed" ]; then
    ok "P9"
else
    fail "P9" "BRAG.md missing SPEC-022 artifact refs:$p9_failed"
fi

# P10 — README.md does NOT contain "(recommended once available)"
assert_not_contains_iregex "P10" "README.md" "recommended once available"

# ===== Group Q — completion subcommand source shape =====

# Q1 — internal/cli/completion.go exists
assert_file_exists "Q1" "internal/cli/completion.go"

# Q2 — internal/cli/completion_test.go exists
assert_file_exists "Q2" "internal/cli/completion_test.go"

# Q3 — completion.go wires GenZshCompletion
assert_contains_literal "Q3" "internal/cli/completion.go" "GenZshCompletion"

# Q4 — completion.go wires GenBashCompletion
assert_contains_literal "Q4" "internal/cli/completion.go" "GenBashCompletion"

# Q5 — completion.go wires GenFishCompletion
assert_contains_literal "Q5" "internal/cli/completion.go" "GenFishCompletion"

# ===== Group R — tutorial §10 shell completions addendum =====

# R1 — tutorial §10 heading exists (line-regex avoids substring trap)
if [ ! -f docs/tutorial.md ]; then
    fail "R1" "docs/tutorial.md does not exist"
elif grep -E -q '^## 10\. Shell completions' docs/tutorial.md; then
    ok "R1"
else
    fail "R1" "docs/tutorial.md missing '## 10. Shell completions' heading"
fi

# R2 — tutorial contains zsh sourcing example
assert_contains_literal "R2" "docs/tutorial.md" "source <(brag completion zsh)"

# R3 — tutorial contains bash sourcing example
assert_contains_literal "R3" "docs/tutorial.md" "source <(brag completion bash)"

# R4 — tutorial contains fish sourcing example
assert_contains_literal "R4" "docs/tutorial.md" "brag completion fish | source"

# ===== Group S — Claude Code plugin packaging (SPEC-041) =====
#
# NOTE: SPEC-041 names this "group K" in its own numbering, but the letter K
# was already taken by the pre-existing "BRAG.md cross-reference" group above
# (K1-K4). Reusing K here would collide with those assertion names, so this
# group is lettered S (the next unused letter after R) instead; the test IDs
# below (S1-S11) map 1:1 to the spec's K1-K11. Deviation recorded in SPEC-041
# Build Completion.

PLUGIN_MANIFEST="plugin/.claude-plugin/plugin.json"
MARKETPLACE_MANIFEST=".claude-plugin/marketplace.json"

# S1 — plugin manifest file exists
assert_file_exists "S1" "$PLUGIN_MANIFEST"

# S2 — claude plugin validate --strict plugin exits 0 (skip if claude CLI absent)
if command -v claude >/dev/null 2>&1; then
    assert_cmd_ok "S2" claude plugin validate --strict plugin
else
    skip "S2" "claude CLI not installed"
fi

# S2-jq — structural fallback (always runs, independent of the claude CLI)
if [ -f "$PLUGIN_MANIFEST" ] && jq -e '.name=="brag" and (.mcpServers.brag.command=="brag")' "$PLUGIN_MANIFEST" >/dev/null 2>&1; then
    ok "S2-jq"
else
    fail "S2-jq" "$PLUGIN_MANIFEST missing name==\"brag\" or mcpServers.brag.command==\"brag\""
fi

# S3 — marketplace manifest exists + validates strict (skip if claude CLI absent)
assert_file_exists "S3" "$MARKETPLACE_MANIFEST"
if command -v claude >/dev/null 2>&1; then
    assert_cmd_ok "S3" claude plugin validate --strict .
else
    skip "S3" "claude CLI not installed"
fi

# S3-jq — structural fallback (always runs)
if [ -f "$MARKETPLACE_MANIFEST" ] && jq -e '.description and (.plugins[0].name=="brag") and (.plugins[0].source=="./plugin")' "$MARKETPLACE_MANIFEST" >/dev/null 2>&1; then
    ok "S3-jq"
else
    fail "S3-jq" "$MARKETPLACE_MANIFEST missing description, plugins[0].name==\"brag\", or plugins[0].source==\"./plugin\""
fi

# S4 — mcpServers.brag runs `brag mcp serve`
if [ -f "$PLUGIN_MANIFEST" ] && jq -e '.mcpServers.brag.args==["mcp","serve"]' "$PLUGIN_MANIFEST" >/dev/null 2>&1; then
    ok "S4"
else
    fail "S4" "$PLUGIN_MANIFEST missing mcpServers.brag.args==[\"mcp\",\"serve\"]"
fi

# S5 — slash-command file exists
assert_file_exists "S5" "plugin/commands/brag.md"

# S6 — slash-command names the approval gate + the tool it gates
if [ ! -f plugin/commands/brag.md ]; then
    fail "S6" "plugin/commands/brag.md does not exist"
else
    has_gate=no; has_tool=no
    if grep -F -q 'Do not execute' plugin/commands/brag.md; then has_gate=yes; fi
    if grep -F -q 'brag add --json' plugin/commands/brag.md; then has_tool=yes; fi
    if [ "$has_gate" = yes ] && [ "$has_tool" = yes ]; then
        ok "S6"
    else
        fail "S6" "approval gate (gate=$has_gate tool=$has_tool)"
    fi
fi

# S7 — Stop hook wired to capture-nudge.sh via ${CLAUDE_PLUGIN_ROOT}
if [ -f plugin/hooks/hooks.json ] && jq -e '.hooks.Stop[0].hooks[0].command | test("CLAUDE_PLUGIN_ROOT.*capture-nudge.sh")' plugin/hooks/hooks.json >/dev/null 2>&1; then
    ok "S7"
else
    fail "S7" "plugin/hooks/hooks.json missing Stop hook wired to CLAUDE_PLUGIN_ROOT/.../capture-nudge.sh"
fi

# S8 — provenance convention documented in the shipped hook
if [ ! -f plugin/hooks/capture-nudge.sh ]; then
    fail "S8" "plugin/hooks/capture-nudge.sh does not exist"
else
    has_agent=no; has_model=no
    if grep -F -q 'agent:<name>' plugin/hooks/capture-nudge.sh; then has_agent=yes; fi
    if grep -F -q 'model:<id>' plugin/hooks/capture-nudge.sh; then has_model=yes; fi
    if [ "$has_agent" = yes ] && [ "$has_model" = yes ]; then
        ok "S8"
    else
        fail "S8" "provenance convention (agent:<name>=$has_agent model:<id>=$has_model)"
    fi
fi

# S9 — plugin README documents the PATH prerequisite + provenance
if [ ! -f plugin/README.md ]; then
    fail "S9" "plugin/README.md does not exist"
else
    has_brew=no; has_agent=no
    if grep -F -q 'brew install' plugin/README.md; then has_brew=yes; fi
    if grep -F -q 'agent:' plugin/README.md; then has_agent=yes; fi
    if [ "$has_brew" = yes ] && [ "$has_agent" = yes ]; then
        ok "S9"
    else
        fail "S9" "plugin/README.md (brew install=$has_brew agent:=$has_agent)"
    fi
fi

# S10 — BRAG.md documents both the plugin path and the provenance convention
if [ ! -f BRAG.md ]; then
    fail "S10" "BRAG.md does not exist"
else
    has_plugin=no; has_agent=no
    if grep -F -q 'plugin' BRAG.md; then has_plugin=yes; fi
    if grep -F -q 'agent:<name>' BRAG.md; then has_agent=yes; fi
    if [ "$has_plugin" = yes ] && [ "$has_agent" = yes ]; then
        ok "S10"
    else
        fail "S10" "BRAG.md (plugin=$has_plugin agent:<name>=$has_agent)"
    fi
fi

# S11 — capture-nudge.sh is executable
if [ -x plugin/hooks/capture-nudge.sh ]; then
    ok "S11"
else
    fail "S11" "plugin/hooks/capture-nudge.sh is not executable (chmod +x)"
fi

# S12 — plugin/.mcp.json declares the brag MCP server (SPEC-041 punch-list
# regression guard). `claude plugin validate --strict` and S2/S2-jq/S4 all
# stayed green while the loader registered 0 MCP servers at runtime, because
# registration reads plugin/.mcp.json, not the inline plugin.json mcpServers
# key — this assertion fails loudly if .mcp.json goes missing or drifts.
PLUGIN_MCP_JSON="plugin/.mcp.json"
assert_file_exists "S12" "$PLUGIN_MCP_JSON"
if [ -f "$PLUGIN_MCP_JSON" ] && jq -e '.brag.command=="brag" and (.brag.args==["mcp","serve"])' "$PLUGIN_MCP_JSON" >/dev/null 2>&1; then
    ok "S12-jq"
else
    fail "S12-jq" "$PLUGIN_MCP_JSON missing brag.command==\"brag\" or brag.args==[\"mcp\",\"serve\"]"
fi

# ===== Group T — for-ai-agents docs (SPEC-058) =====

AGENT_DOC="docs/for-ai-agents.md"

# T1 — page exists
assert_file_exists "T1" "$AGENT_DOC"

# T2 — page is a full playbook, not a stub
assert_line_count_band "T2" "$AGENT_DOC" 120 500

# T3 — names all five MCP tools
if [ ! -f "$AGENT_DOC" ]; then
    fail "T3" "$AGENT_DOC does not exist"
else
    t3_missing=""
    for tool in brag_add brag_list brag_memory brag_search brag_stats; do
        if ! grep -F -q -- "$tool" "$AGENT_DOC"; then
            t3_missing="$t3_missing $tool"
        fi
    done
    if [ -z "$t3_missing" ]; then
        ok "T3"
    else
        fail "T3" "$AGENT_DOC missing tool names:$t3_missing"
    fi
fi

# T4 — documents the registration command
assert_contains_literal "T4" "$AGENT_DOC" "brag mcp install"

# T5 — gives the manual mcpServers JSON snippet
assert_contains_literal "T5" "$AGENT_DOC" '{"mcpServers":{"brag":{"command":"brag","args":["mcp","serve"]}}}'

# T6 — states the client-startup-reconnect note
if [ ! -f "$AGENT_DOC" ]; then
    fail "T6" "$AGENT_DOC does not exist"
else
    has_startup=no; has_reconnect=no
    if grep -F -q -- "connect at client startup" "$AGENT_DOC"; then has_startup=yes; fi
    if grep -F -q -- "reconnect" "$AGENT_DOC"; then has_reconnect=yes; fi
    if [ "$has_startup" = yes ] && [ "$has_reconnect" = yes ]; then
        ok "T6"
    else
        fail "T6" "reconnect note (startup=$has_startup reconnect=$has_reconnect)"
    fi
fi

# T7 — states the project-not-auto-filled gotcha (names the field)
if [ ! -f "$AGENT_DOC" ]; then
    fail "T7" "$AGENT_DOC does not exist"
else
    has_phrase=no; has_field=no
    if grep -F -q -- "does not auto-fill" "$AGENT_DOC"; then has_phrase=yes; fi
    if grep -F -q -- "project" "$AGENT_DOC"; then has_field=yes; fi
    if [ "$has_phrase" = yes ] && [ "$has_field" = yes ]; then
        ok "T7"
    else
        fail "T7" "gotcha (phrase=$has_phrase field=$has_field)"
    fi
fi

# T8 — cross-links the fix
assert_contains_literal "T8" "$AGENT_DOC" "brag project ensure"

# T9 — documents provenance stamping (all five reserved namespaces)
if [ ! -f "$AGENT_DOC" ]; then
    fail "T9" "$AGENT_DOC does not exist"
else
    t9_missing=""
    for tok in "agent:<name>" "model:<id>" "session:<id>" "cost:<n>" "tokens:<n>"; do
        if ! grep -F -q -- "$tok" "$AGENT_DOC"; then
            t9_missing="$t9_missing $tok"
        fi
    done
    if [ -z "$t9_missing" ]; then
        ok "T9"
    else
        fail "T9" "$AGENT_DOC missing provenance tokens:$t9_missing"
    fi
fi

# T10 — documents the DB resolution order
if [ ! -f "$AGENT_DOC" ]; then
    fail "T10" "$AGENT_DOC does not exist"
else
    t10_missing=""
    for tok in "--db" "BRAGFILE_DB" "~/.bragfile/db.sqlite"; do
        if ! grep -F -q -- "$tok" "$AGENT_DOC"; then
            t10_missing="$t10_missing $tok"
        fi
    done
    if [ -z "$t10_missing" ]; then
        ok "T10"
    else
        fail "T10" "$AGENT_DOC missing db-resolution tokens:$t10_missing"
    fi
fi

# T11 — carries the impact-framing convention (distinctive phrase)
assert_contains_literal "T11" "$AGENT_DOC" "a metric or a named outcome"

# T12 — gives the resolved per-client config paths
if [ ! -f "$AGENT_DOC" ]; then
    fail "T12" "$AGENT_DOC does not exist"
else
    t12_missing=""
    for tok in ".mcp.json" "~/.claude.json" ".cursor/mcp.json" "claude_desktop_config.json"; do
        if ! grep -F -q -- "$tok" "$AGENT_DOC"; then
            t12_missing="$t12_missing $tok"
        fi
    done
    if [ -z "$t12_missing" ]; then
        ok "T12"
    else
        fail "T12" "$AGENT_DOC missing config-path tokens:$t12_missing"
    fi
fi

# T13 — README links the page
assert_contains_literal "T13" "README.md" "docs/for-ai-agents.md"

# T14 — README has an agent/MCP section heading (line-regex avoids substring trap)
if [ ! -f README.md ]; then
    fail "T14" "README.md does not exist"
elif grep -E -q '^## .*[Aa]gent' README.md; then
    ok "T14"
else
    fail "T14" "README.md missing an agent/MCP '## ' section heading"
fi

# ===== Group U — brag memory docs (SPEC-073) =====

# U1 — api-contract.md has the brag memory section heading
assert_contains_literal "U1" "docs/api-contract.md" "### \`brag memory"

# U2 — api-contract.md claims the eighth DEC-014 consumer ordinal
assert_contains_literal "U2" "docs/api-contract.md" "**eighth** DEC-014 consumer"

# U3 — api-contract.md names both DEC-043 and DEC-044
if [ ! -f docs/api-contract.md ]; then
    fail "U3" "docs/api-contract.md does not exist"
else
    u3_missing=""
    for tok in "DEC-043" "DEC-044"; do
        if ! grep -F -q -- "$tok" docs/api-contract.md; then
            u3_missing="$u3_missing $tok"
        fi
    done
    if [ -z "$u3_missing" ]; then
        ok "U3"
    else
        fail "U3" "docs/api-contract.md missing:$u3_missing"
    fi
fi

# U4 — README.md fenced blocks mention brag memory
if [ ! -f README.md ]; then
    fail "U4" "README.md does not exist"
else
    fenced=$(awk '/^```/{f=!f; next} f' README.md)
    if printf '%s\n' "$fenced" | grep -F -q -- "brag memory"; then
        ok "U4"
    else
        fail "U4" "README.md fenced blocks missing 'brag memory'"
    fi
fi

# U5 — tutorial.md mentions brag memory
assert_contains_literal "U5" "docs/tutorial.md" "brag memory"

# U6 — AGENTS.md carries the glossary entry
assert_contains_literal "U6" "AGENTS.md" "- **memory** —"

# U7 — api-contract.md states the DEC-044 honesty scoping user-visibly
assert_contains_literal "U7" "docs/api-contract.md" "never stamped on an entry"

# U8 — architecture.md's package table names internal/memory
assert_contains_literal "U8" "docs/architecture.md" "internal/memory"

# ===== Group V — MCP resources docs (SPEC-074 / DEC-045) =====

# V1 — for-ai-agents.md names all three resource URIs
if [ ! -f "$AGENT_DOC" ]; then
    fail "V1" "$AGENT_DOC does not exist"
else
    v1_missing=""
    for uri in "brag://memory/recent" "brag://memory/project/{name}" "brag://projects"; do
        if ! grep -F -q -- "$uri" "$AGENT_DOC"; then
            v1_missing="$v1_missing $uri"
        fi
    done
    if [ -z "$v1_missing" ]; then
        ok "V1"
    else
        fail "V1" "$AGENT_DOC missing resource URIs:$v1_missing"
    fi
fi

# V2 — api-contract.md documents the recent-memory resource
assert_contains_literal "V2" "docs/api-contract.md" "brag://memory/recent"

# V3 — for-ai-agents.md states the soft-boost correction (the doc-level guard
# on LD4: the docs cannot describe the project resource as a scope)
assert_contains_literal "V3" "$AGENT_DOC" "a soft boost, not a filter"

# V4 — architecture.md names the internal/ftsquery extraction
assert_contains_literal "V4" "docs/architecture.md" "internal/ftsquery"

# ===== release-cut hygiene (SPEC-076 / F3) =====
#
# Both of these would have FAILED at the v0.5.2 cut, which is the point: the
# release pre-flight carried "plugin version pin matches the tag" and
# "compare-links repointed" as items a human ticked by hand, and both were
# missed. These are the mechanical form of those two items.

# W1 — the plugin version pin equals the LATEST dated CHANGELOG section.
# Deliberately not "the pin has *a* dated section": at the v0.5.2 cut the pin
# sat on 0.5.1, which had a perfectly good dated section, so the weaker form
# would have passed and caught nothing. The defect is the pin falling BEHIND
# the newest release, so the newest release is what it must be compared to.
w1_pin=$(sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
    plugin/.claude-plugin/plugin.json | head -1)
w1_latest=$(sed -n 's/^## \[\([0-9][^]]*\)\] - .*/\1/p' CHANGELOG.md | head -1)
if [ -z "$w1_pin" ]; then
    fail "W1" "could not read version pin from plugin/.claude-plugin/plugin.json"
elif [ -z "$w1_latest" ]; then
    fail "W1" "CHANGELOG.md has no dated '## [x.y.z] - ' section to compare the pin against"
elif [ "$w1_pin" = "$w1_latest" ]; then
    ok "W1"
else
    fail "W1" "plugin.json pins ${w1_pin} but the latest released CHANGELOG section is ${w1_latest}"
fi

# W2 — every dated CHANGELOG version heading has a matching compare-link.
# Catches a released section whose link was never added (v0.5.2 had none).
w2_missing=""
while IFS= read -r w2_ver; do
    if ! grep -q "^\[${w2_ver}\]: " CHANGELOG.md; then
        w2_missing="$w2_missing $w2_ver"
    fi
done <<EOF
$(sed -n 's/^## \[\([0-9][^]]*\)\] - .*/\1/p' CHANGELOG.md)
EOF
if [ -z "$w2_missing" ]; then
    ok "W2"
else
    fail "W2" "CHANGELOG.md version heading(s) with no compare-link:$w2_missing"
fi

# W3 — user-facing "current version" claims match the latest release.
# The README status line and the tutorial's "shipped as of" line are the two
# places a reader is told what version this is. Both silently rotted through
# v0.5.2 AND v0.6.0 (they still said v0.5.1) because the release pre-flight
# checked the CHANGELOG, the plugin pin and the compare-links — but never the
# prose that users actually read first. Same defect class as W1/W2: a claim
# duplicating the release version with nothing holding it there.
w3_latest=$(sed -n 's/^## \[\([0-9][^]]*\)\] - .*/\1/p' CHANGELOG.md | head -1)
w3_bad=""
if [ -z "$w3_latest" ]; then
    fail "W3" "CHANGELOG.md has no dated '## [x.y.z] - ' section to check version claims against"
else
    if ! grep -q "^> \*\*Status:\*\* v${w3_latest} shipped" README.md; then
        w3_bad="$w3_bad README.md(status-line)"
    fi
    if ! grep -q "shipped as of v${w3_latest}" docs/tutorial.md; then
        w3_bad="$w3_bad docs/tutorial.md(shipped-as-of)"
    fi
    if [ -z "$w3_bad" ]; then
        ok "W3"
    else
        fail "W3" "latest release is ${w3_latest} but these still claim an older version:$w3_bad"
    fi
fi

# W4 — a stage marked `status: shipped` has no reflection placeholders left.
# scripts/archive-spec.sh refuses to archive a SPEC whose reflection still has
# `<answer>` placeholders, and that guard has held since STAGE-004. Stages have
# no equivalent, because stages are never archived — so STAGE-019 was set to
# `status: shipped` with its entire Stage-Level Reflection still on template
# placeholders, and nothing noticed. Same defect class, different artifact.
w4_bad=""
for w4_f in projects/*/stages/STAGE-*.md; do
    [ -f "$w4_f" ] || continue
    grep -q '^  status: shipped' "$w4_f" || continue
    if grep -qE '<yes/no \+ notes>|<number vs\. plan>|<one sentence>|<one-line updates>|<one-line items>|^   — <answer>' "$w4_f"; then
        w4_bad="$w4_bad $(basename "$w4_f")"
    fi
done
if [ -z "$w4_bad" ]; then
    ok "W4"
else
    fail "W4" "shipped stage(s) still carrying reflection placeholders:$w4_bad"
fi

# W5 — brag_add's MCP tool description tells the agent about evidence links.
# SPEC-078 LD1/LD2. This is the single highest-leverage string in the project:
# it is read on EVERY tool call by the agent that authors most of the corpus.
# It said nothing about evidence links while adoption sat at zero.
if grep -A 1 'Name:        "brag_add"' internal/mcpserver/server.go | grep -q 'pr:'; then
    ok "W5"
else
    fail "W5" "brag_add's Description must name the evidence-link convention (pr:/commit:)"
fi

# W6 — the agent-facing docs carry the ref-preference ORDER, not just a mention.
# pr: must be recommended ahead of commit:, because a squash-merge orphans the
# branch commit an agent would otherwise record (BRAG.md documents the case).
w6_pr=$(grep -n 'pr:' docs/for-ai-agents.md | head -1 | cut -d: -f1)
w6_commit=$(grep -n 'commit:' docs/for-ai-agents.md | head -1 | cut -d: -f1)
if [ -n "$w6_pr" ] && [ -n "$w6_commit" ] && [ "$w6_pr" -lt "$w6_commit" ]; then
    ok "W6"
else
    fail "W6" "docs/for-ai-agents.md must name pr: BEFORE commit: (pr=${w6_pr:-none} commit=${w6_commit:-none})"
fi

# ===== Group X — engineering-practices entry point (SPEC-079) =====

PRACTICES_DOC="docs/engineering-practices.md"

# X1 — the practices page exists.
assert_file_exists "X1" "$PRACTICES_DOC"

# X2 — the README points at it. This is the whole "entry point" claim: the page
# is worthless if the front door does not name it.
assert_contains_literal "X2" "README.md" "docs/engineering-practices.md"

# X3 — THE COUNTING GUARD. The inventory block on the page must equal
# `scripts/inventory.sh` output byte-for-byte. This is the assertion that makes
# every number on the page unrottable: it does not check one hand-typed count
# against one hand-typed expectation (which is just a second thing to maintain)
# — it recomputes the whole table from the repo and diffs. The remedy on
# failure is mechanical: run `just inventory`, paste between the markers.
if [ ! -f "$PRACTICES_DOC" ]; then
    fail "X3" "$PRACTICES_DOC does not exist"
elif [ ! -x scripts/inventory.sh ]; then
    fail "X3" "scripts/inventory.sh is missing or not executable"
else
    x3_want=$(./scripts/inventory.sh)
    x3_got=$(awk '/<!-- inventory:begin/{f=1; next} /<!-- inventory:end/{f=0} f' "$PRACTICES_DOC")
    if [ -z "$x3_got" ]; then
        fail "X3" "$PRACTICES_DOC has no content between the inventory:begin/inventory:end markers"
    elif [ "$x3_got" = "$x3_want" ]; then
        ok "X3"
    else
        fail "X3" "inventory block is stale — run \`just inventory\` and paste between the markers:
--- on the page ---
$x3_got
--- computed from the repo ---
$x3_want"
    fi
fi

# X4 — the page describes the test mechanism as it ACTUALLY is.
# There are no golden FILES in this repo: `find . -name '*.golden'` is empty and
# there are no testdata/ directories. The mechanism is byte-exact expected
# output embedded as Go string literals. Deliberately a POSITIVE assertion, not
# a NOT-contains on "golden file": the page has to be free to say what the
# mechanism is NOT, and a NOT-contains would forbid exactly that sentence.
assert_contains_literal "X4" "$PRACTICES_DOC" "embedded as Go string literals"

# X5 — no adjective stands alone. STAGE-021's success criterion, mechanised.
# Every token below is a claim a reader cannot check; the page is required to
# point at a path, a count or a named test instead.
assert_not_contains_iregex "X5" "$PRACTICES_DOC" 'rigorous|comprehensive|world-class|best-in-class|battle-tested|cutting-edge|state-of-the-art'

# X6 — the corrections claim is made BY CITATION, not by count.
# There is no counting rule for "decision records that log their own
# correction" (only 1 of 46 DECs carries an explicit `## Amendment` heading, and
# a keyword grep matches all of them because ordinary prose uses those words), so
# page names the specific records. This pins that it keeps naming them.
if [ ! -f "$PRACTICES_DOC" ]; then
    fail "X6" "$PRACTICES_DOC does not exist"
else
    x6_missing=""
    for x6_id in "DEC-004" "DEC-015" "DEC-025" "DEC-043" "DEC-044" "PROJ-006"; do
        if ! grep -F -q -- "$x6_id" "$PRACTICES_DOC"; then
            x6_missing="$x6_missing $x6_id"
        fi
    done
    if [ -z "$x6_missing" ]; then
        ok "X6"
    else
        fail "X6" "$PRACTICES_DOC must cite each correction record by id; missing:$x6_missing"
    fi
fi

# X7 — the page stays an INDEX, not an essay. Same idiom as A1/C2/D2. The
# stated failure mode for this document is that it turns into new prose about
# the project instead of a route to artifacts that already exist; a size band
# is the cheap mechanical form of that.
#
# WORDS, NOT LINES, since SPEC-082. The switch trigger is the one SPEC-081
# wrote when it built assert_word_count_band: "switch a caller to this helper
# when its swing exceeds its headroom, not before." X7 was named there as the
# next closest caller — headroom 17 lines, measured reflow swing +15 — and
# SPEC-082's own edit to this page took it to 301 lines, one over the old
# ceiling. The alternative was to raise the line band, which is the
# bump-the-number-until-green disarming that SPEC-079 LD5 refused.
#
# The band tightens the guard rather than loosening it. Converted at this
# file's own 7.93 words/line, the old 150..300 line band was ~1,190..2,380
# words, an admissible span of ~1,190. 1800..2700 is a span of 900 — 24%
# narrower — with 232 words of ceiling headroom, against a reflow swing of
# at most ±4 words (only the `-`/`>` prefixes on wrapped list items move).
assert_word_count_band "X7" "$PRACTICES_DOC" 1800 2700

# X8 — `just inventory` is wired, so the remedy X3 names actually exists.
if [ ! -f justfile ]; then
    fail "X8" "justfile does not exist"
elif grep -E -q '^inventory:' justfile; then
    ok "X8"
else
    fail "X8" "no '^inventory:' recipe in justfile"
fi

# ===== Group Y — godoc pass + legibility repairs (SPEC-080) =====

# Y1 — the five packages the godoc pass names now carry a package doc
# comment: a `//` comment line immediately preceding `package X`, the exact
# adjacency rule go/doc (and `go doc`) key off. Framing's "7 missing" was
# wrong — internal/export and internal/mcpserver already had one; the real
# gap, re-measured at design, was these five.
y1_missing=""
for y1_target in \
    "cmd/brag/main.go" \
    "internal/cli/root.go" \
    "internal/config/config.go" \
    "internal/storage/store.go" \
    "internal/story/bundle.go"
do
    if [ ! -f "$y1_target" ]; then
        y1_missing="$y1_missing $y1_target(missing-file)"
        continue
    fi
    y1_pkgline=$(grep -n '^package ' "$y1_target" | head -1 | cut -d: -f1)
    if [ -z "$y1_pkgline" ]; then
        y1_missing="$y1_missing $y1_target(no-package-decl)"
        continue
    fi
    y1_prev=$((y1_pkgline - 1))
    if [ "$y1_prev" -lt 1 ] || ! sed -n "${y1_prev}p" "$y1_target" | grep -q '^//'; then
        y1_missing="$y1_missing $y1_target"
    fi
done
if [ -z "$y1_missing" ]; then
    ok "Y1"
else
    fail "Y1" "missing a doc comment immediately before 'package':$y1_missing"
fi

# Y2 — the DEC-041 gap is explained IN decisions/, not only in the backlog:
# a tombstone file exists and is explicitly marked `insight.type:
# reservation`, not `decision` — the marker the Decision records count
# (Y3/inventory.sh) relies on to NOT count it as a decision.
y2_file=$(ls decisions/DEC-041-*.md 2>/dev/null | head -1)
if [ -z "$y2_file" ]; then
    fail "Y2" "no decisions/DEC-041-*.md tombstone file"
elif grep -q '^  type: reservation' "$y2_file"; then
    ok "Y2"
else
    fail "Y2" "$y2_file exists but is missing '  type: reservation' in its front-matter"
fi

# Y3 — the tombstone does not inflate the Decision records count, and its
# own reservation is counted separately. Pins the SEMANTIC values, not just
# X3-style script-vs-page self-consistency, which would happily pass even if
# both sides agreed on a wrong number (the failure mode this pin exists to
# catch: someone edits the type filter in inventory.sh and it silently starts
# counting the tombstone as a decision — script and page would still agree,
# just agree on 47).
#
# RE-PINNED 45 -> 46 at SPEC-083, which adds DEC-047. Deliberate corpus
# change, not drift. Note this is the SECOND guard on a number SPEC-083
# moves — Y4 pins the other — which is exactly the pair AGENTS.md §9's
# half (b) exists for: grep the harness for every literal the spec moves.
if [ ! -x scripts/inventory.sh ]; then
    fail "Y3" "scripts/inventory.sh is missing or not executable"
else
    y3_out=$(./scripts/inventory.sh)
    y3_bad=""
    printf '%s\n' "$y3_out" | grep -F -q 'Decision records | 46 |' || y3_bad="$y3_bad decision-records!=46"
    printf '%s\n' "$y3_out" | grep -F -q 'Decision numbers reserved, not yet decided | 1 |' || y3_bad="$y3_bad reserved-decisions!=1"
    if [ -z "$y3_bad" ]; then
        ok "Y3"
    else
        fail "Y3" "inventory.sh row value(s) wrong:$y3_bad"
    fi
fi

# Y4 — the open-questions count is DERIVED (inventory.sh), not restated: the
# fix for the STAGE-021 line that has been wrong three times in three days
# (8 of 18 wrong total; 9 of 18 wrong on both halves, from a grep that
# matched the file's own header comment; 7 of 17 correct until a merge
# landed a new question). Pins the semantic values, same rationale as Y3.
#
# RE-PINNED 18/6 -> 19/7 at SPEC-082, which appends the
# no-sql-in-cli-layer-test-scope question (LD7). This is a deliberate
# corpus change, not drift: the pin moves with the register it describes.
#
# RE-PINNED AGAIN 19/7 -> 19/6 at SPEC-083, which ANSWERS that same question
# (DEC-047). The total does not move: the entry is closed, not removed.
if [ ! -x scripts/inventory.sh ]; then
    fail "Y4" "scripts/inventory.sh is missing or not executable"
else
    y4_out=$(./scripts/inventory.sh)
    y4_bad=""
    printf '%s\n' "$y4_out" | grep -F -q 'Questions tracked in guidance/questions.yaml | 19 |' || y4_bad="$y4_bad questions-total!=19"
    printf '%s\n' "$y4_out" | grep -F -q 'of those, still open | 6 |' || y4_bad="$y4_bad questions-open!=6"
    if [ -z "$y4_bad" ]; then
        ok "Y4"
    else
        fail "Y4" "inventory.sh row value(s) wrong:$y4_bad"
    fi
fi

# Y5 — the two questions answered in practice by this spec (editor-template-
# format by DEC-009; summary-grouping-heuristics by SPEC-018/DEC-014 +
# aggregate.GroupForHighlights) are marked closed in the register, not left
# stale the way both sat since 2026-04-19 despite the answer already
# existing.
y5_missing=""
for y5_id in "editor-template-format" "summary-grouping-heuristics"; do
    y5_status=$(awk -v want="  - id: $y5_id" '
        $0 == want { f=1; next }
        f && /^  - id: / { exit }
        f && /^    status: / { print; exit }
    ' guidance/questions.yaml)
    if [ "$y5_status" != "    status: answered" ]; then
        y5_missing="$y5_missing $y5_id"
    fi
done
if [ -z "$y5_missing" ]; then
    ok "Y5"
else
    fail "Y5" "guidance/questions.yaml entries not marked 'status: answered':$y5_missing"
fi

# ===== Group Z — lint gate + coverage floor (SPEC-082) =====

GOLANGCI=".golangci.yml"
COVERAGE_SH="scripts/coverage.sh"

# Z1 — the lint config exists at all. Before SPEC-082 there was none.
assert_file_exists "Z1" "$GOLANGCI"

# Z2 — the enabled set is CHOSEN, not inherited: `default: none` plus exactly
# the nine linters SPEC-082 locked, each with its own argument in the file.
# This is the assertion that fires if someone "just turns on a few more":
# `linters.default: all` reports 15,300 issues on this tree, and a gate that
# size gets silenced with //nolint, which is worse than no gate.
if [ ! -f "$GOLANGCI" ]; then
    fail "Z2" "$GOLANGCI does not exist"
elif ! grep -Eq '^  default: none$' "$GOLANGCI"; then
    fail "Z2" "$GOLANGCI must declare '  default: none' — nothing inherited"
else
    z2_want="depguard errcheck errorlint ineffassign nolintlint rowserrcheck sqlclosecheck staticcheck unused"
    z2_got=$(awk '/^  enable:$/{f=1; next} f && /^  [a-z]/{f=0} f' "$GOLANGCI" \
        | grep -oE '^    - [a-z]+' | awk '{print $2}' | sort | tr '\n' ' ' | sed 's/ $//')
    if [ "$z2_got" = "$z2_want" ]; then
        ok "Z2"
    else
        fail "Z2" "enabled linters are [$z2_got]; expected [$z2_want]"
    fi
fi

# Z3 — depguard denies BOTH halves of the no-sql-in-cli-layer constraint, and
# is scoped to the package its path glob names. The driver-only half is the
# whole reason depguard was chosen over porting internal/mcpserver's
# TestNoSQLImport: that test greps the literal "database/sql" and nothing
# else, so an import of modernc.org/sqlite alone walks straight past it.
if [ ! -f "$GOLANGCI" ]; then
    fail "Z3" "$GOLANGCI does not exist"
else
    z3_missing=""
    for z3_needle in "no-sql-in-cli-layer:" "internal/cli/*.go" "!\$test" \
                     'pkg: "database/sql"' 'pkg: "modernc.org/sqlite"'; do
        if ! grep -F -q -- "$z3_needle" "$GOLANGCI"; then
            z3_missing="$z3_missing [$z3_needle]"
        fi
    done
    if [ -z "$z3_missing" ]; then
        ok "Z3"
    else
        fail "Z3" "$GOLANGCI depguard rule is missing:$z3_missing"
    fi
fi

# Z4 — both gates actually run in CI. A config nobody runs is documentation.
if [ ! -f "$CI_WORKFLOW" ]; then
    fail "Z4" "$CI_WORKFLOW does not exist"
else
    z4_missing=""
    grep -F -q -- "golangci/golangci-lint-action@" "$CI_WORKFLOW" || z4_missing="$z4_missing [golangci-lint-action]"
    grep -F -q -- "./scripts/coverage.sh" "$CI_WORKFLOW" || z4_missing="$z4_missing [scripts/coverage.sh]"
    if [ -z "$z4_missing" ]; then
        ok "Z4"
    else
        fail "Z4" "$CI_WORKFLOW does not run:$z4_missing"
    fi
fi

# Z5 — THE FLOOR GUARD, same idiom as X3: derive, then diff. The floor is
# written down once, in scripts/coverage.sh. This reads it back out and
# asserts the practices page states the same number, so the page cannot claim
# a floor CI is not enforcing (or the reverse).
if [ ! -f "$COVERAGE_SH" ]; then
    fail "Z5" "$COVERAGE_SH does not exist"
elif [ ! -f "$PRACTICES_DOC" ]; then
    fail "Z5" "$PRACTICES_DOC does not exist"
else
    z5_floor=$(grep -E '^FLOOR=' "$COVERAGE_SH" | head -1 | cut -d= -f2)
    if [ -z "$z5_floor" ]; then
        fail "Z5" "$COVERAGE_SH has no '^FLOOR=' line"
    elif grep -F -q -- "${z5_floor}%" "$PRACTICES_DOC"; then
        ok "Z5"
    else
        fail "Z5" "$PRACTICES_DOC does not state the enforced floor ${z5_floor}% from $COVERAGE_SH"
    fi
fi

# Z6 — the remedies named in the config and on the page actually exist as
# commands. Same idiom as X8: `just inventory` had to be real for X3's failure
# message to mean anything.
if [ ! -f justfile ]; then
    fail "Z6" "justfile does not exist"
else
    z6_missing=""
    grep -E -q '^lint:' justfile || z6_missing="$z6_missing [lint]"
    grep -E -q '^coverage:' justfile || z6_missing="$z6_missing [coverage]"
    if [ -z "$z6_missing" ]; then
        ok "Z6"
    else
        fail "Z6" "justfile is missing recipe(s):$z6_missing"
    fi
fi

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
# decisions/_template.md is deliberately outside every count (it does not match
# DEC-*.md) and stays outside this one.
z7_files=$(ls decisions/DEC-*.md 2>/dev/null | wc -l | tr -d ' ')
z7_decisions=$(grep -l '^  type: decision' decisions/DEC-*.md 2>/dev/null | wc -l | tr -d ' ')
z7_reserved=$(grep -l '^  type: reservation' decisions/DEC-*.md 2>/dev/null | wc -l | tr -d ' ')
z7_sum=$((z7_decisions + z7_reserved))
if [ "$z7_sum" -eq "$z7_files" ]; then
    ok "Z7"
else
    fail "Z7" "the inventory covers $z7_sum of $z7_files decisions/DEC-*.md files ($z7_decisions decision + $z7_reserved reservation). A DEC-*.md with a missing or unknown 'insight.type' is counted by neither row: fix its front-matter, or teach scripts/inventory.sh a row for the new type."
fi

# ===== Group AA — the constraint amendment (SPEC-083 / DEC-047) =====
#
# SPEC-083 NARROWED a blocking constraint's prose. That is the manoeuvre where
# enforcement quietly erodes: the mechanism keeps running while the rule it
# points at drifts out from under it. These three assertions pin the narrowing
# to exactly what was decided, so a later widening, a severity downgrade or a
# revert of the repaired comments has to be a deliberate edit to this file too.
#
# What is deliberately NOT asserted: anything about the test half. It is
# unguarded BY DECISION (DEC-047), and both available assertions would be
# wrong. "No CLI test imports SQL" contradicts the decision. "Some CLI test
# imports SQL" fires the day someone finally ports them to storagetest, which
# is the outcome the decision's own revisit triggers want. An unenforced half
# stays unenforced; it is documented in .golangci.yml and in DEC-047 instead.

CONSTRAINTS_YAML="guidance/constraints.yaml"

# AA1 — THE ANTI-EROSION GUARD. Reads the no-sql-in-cli-layer entry as a block
# (its `- id:` line to the next one) and checks four things inside it: the rule
# text is production-scoped, the severity is STILL blocking, the path glob is
# unchanged, and the rationale cites the record that authorised the change.
# Severity is the load-bearing one — the spec that narrowed this rule must not
# be remembered as the spec that softened it.
if [ ! -f "$CONSTRAINTS_YAML" ]; then
    fail "AA1" "$CONSTRAINTS_YAML does not exist"
else
    aa1_block=$(awk '
        /^  - id: no-sql-in-cli-layer$/ { f=1; print; next }
        f && /^  - id: / { exit }
        f { print }
    ' "$CONSTRAINTS_YAML")
    aa1_bad=""
    if [ -z "$aa1_block" ]; then
        aa1_bad=" [no constraint with id no-sql-in-cli-layer]"
    else
        printf '%s\n' "$aa1_block" | grep -q '^    rule: "Production files under internal/cli/' \
            || aa1_bad="$aa1_bad [rule text is not scoped to production files]"
        printf '%s\n' "$aa1_block" | grep -q '^    severity: blocking$' \
            || aa1_bad="$aa1_bad [severity is no longer blocking]"
        printf '%s\n' "$aa1_block" | grep -F -q 'paths: ["internal/cli/**"]' \
            || aa1_bad="$aa1_bad [paths glob changed]"
        printf '%s\n' "$aa1_block" | grep -F -q 'DEC-047' \
            || aa1_bad="$aa1_bad [rationale does not cite DEC-047]"
    fi
    if [ -z "$aa1_bad" ]; then
        ok "AA1"
    else
        fail "AA1" "$CONSTRAINTS_YAML no-sql-in-cli-layer:$aa1_bad"
    fi
fi

# AA2 — the register entry is closed, cites the record that closed it, and that
# record exists on disk. Y5's idiom plus the citation: a question marked
# answered with no pointer to the answer is how `editor-template-format` sat
# stale for four months with its answer already shipped.
aa2_block=$(awk '
    /^  - id: no-sql-in-cli-layer-test-scope$/ { f=1; print; next }
    f && /^  - id: / { exit }
    f { print }
' guidance/questions.yaml)
aa2_bad=""
if [ -z "$aa2_block" ]; then
    aa2_bad=" [no question with id no-sql-in-cli-layer-test-scope]"
else
    printf '%s\n' "$aa2_block" | grep -q '^    status: answered$' \
        || aa2_bad="$aa2_bad [status is not answered]"
    printf '%s\n' "$aa2_block" | grep -F -q 'DEC-047' \
        || aa2_bad="$aa2_bad [does not cite DEC-047]"
fi
ls decisions/DEC-047-*.md >/dev/null 2>&1 \
    || aa2_bad="$aa2_bad [decisions/DEC-047-*.md does not exist]"
if [ -z "$aa2_bad" ]; then
    ok "AA2"
else
    fail "AA2" "guidance/questions.yaml no-sql-in-cli-layer-test-scope:$aa2_bad"
fi

# AA3 — the five comments this spec repaired stay repaired. Every phrase below
# was a FALSE claim standing on main: two package comments said the boundary
# was held "by convention and review, not an automated test" months after
# SPEC-082 made it a lint gate; store.go's Store TYPE comment said "no other
# package imports a SQL driver" while internal/storage/storagetest does (the
# import graph says exactly two non-test files import a driver, and that is the
# other one); list_test.go said CLI tests "cannot import database/sql", which
# was wrong under BOTH readings of the rule; and storagetest's package comment
# gave the constraint as the reason CLI tests route through it, which stopped
# being a reason when DEC-047 scoped the constraint to production files.
#
# NOT-contains by design: each needle is the false claim itself, so a revert
# fails while the replacement prose stays free to be reworded — just not back.
aa3_bad=""
while IFS='|' read -r aa3_file aa3_phrase; do
    [ -n "$aa3_file" ] || continue
    if [ ! -f "$aa3_file" ]; then
        aa3_bad="$aa3_bad [$aa3_file is missing]"
    elif grep -F -q -- "$aa3_phrase" "$aa3_file"; then
        aa3_bad="$aa3_bad [$aa3_file still says \"$aa3_phrase\"]"
    fi
done <<EOF
internal/cli/root.go|convention and review
internal/storage/store.go|convention and review
internal/storage/store.go|no other package imports a
internal/cli/list_test.go|cannot import database/sql
internal/storage/storagetest/storagetest.go|without violating
EOF
if [ -z "$aa3_bad" ]; then
    ok "AA3"
else
    fail "AA3" "a claim SPEC-083 corrected is back in the tree:$aa3_bad"
fi

# ===== finalise =====

if [ "$FAIL_COUNT" -gt 0 ]; then
    printf '\nFAILED: %d assertion(s) failed.\n' "$FAIL_COUNT" >&2
    exit 1
fi

# F4 — harness self-pass meta (printed last, after all assertions OK)
ok "F4"

printf '\nALL OK: documentation-content assertions passed.\n'
exit 0

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

ok() {
    printf 'OK:   %s\n' "$1"
}

fail() {
    printf 'FAIL: %s: %s\n' "$1" "$2"
    FAIL_COUNT=$((FAIL_COUNT + 1))
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

# A1 — README line count band 100..250
assert_line_count_band "A1" "README.md" 100 250

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
    if grep -F -q 'brew install jysf/bragfile/bragfile' README.md; then has_brew=yes; fi
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
for src in README.md CONTRIBUTING.md docs/development.md; do
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
hits=$(grep -rn -F 'docs/development.md' . \
    --include='*.md' \
    --exclude-dir=projects \
    --exclude-dir=node_modules \
    --exclude-dir=.git \
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

# ===== finalise =====

if [ "$FAIL_COUNT" -gt 0 ]; then
    printf '\nFAILED: %d assertion(s) failed.\n' "$FAIL_COUNT" >&2
    exit 1
fi

# F4 — harness self-pass meta (printed last, after all assertions OK)
ok "F4"

printf '\nALL OK: documentation-content assertions passed.\n'
exit 0

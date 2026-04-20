#!/usr/bin/env bash
# scripts/weekly-review.sh — print the Weekly Review prompt with recent context.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/_lib.sh"

require_initialized

ACTIVE_PROJECT=$(get_active_project)
ACTIVE_PROJECT_DIR="${REPO_ROOT}/projects/${ACTIVE_PROJECT}"

cat <<'EOF'
# ============================================================
# Weekly Review — copy everything below this line into Claude
# ============================================================

Weekly review of the repo. Read the files below (paths relative to repo root)
and produce the report as defined in FIRST_SESSION_PROMPTS.md, Prompt 6.

Active project for this review: 
EOF

echo "${ACTIVE_PROJECT}"
echo ""
echo "## Repo-level files to read"
echo "- AGENTS.md"
echo "- .repo-context.yaml"
echo "- guidance/constraints.yaml"
echo "- guidance/questions.yaml"
echo ""

echo "## Decision files (across all projects)"
find "${REPO_ROOT}/decisions" -name "DEC-*.md" 2>/dev/null | sort | sed 's|^|- |' || true
echo ""

echo "## Active project files"
echo "- projects/${ACTIVE_PROJECT}/brief.md"
find "${ACTIVE_PROJECT_DIR}/stages" -name "STAGE-*.md" 2>/dev/null | sort | sed 's|^|- |' || true
echo ""

echo "## Recently shipped specs in ${ACTIVE_PROJECT} (read Reflection sections)"
if [ -d "${ACTIVE_PROJECT_DIR}/specs/done" ]; then
    # Last 10 by mtime
    if [ "$(uname)" = "Darwin" ]; then
        find "${ACTIVE_PROJECT_DIR}/specs/done" -name "SPEC-*.md" -print0 2>/dev/null \
            | xargs -0 stat -f "%m %N" 2>/dev/null \
            | sort -rn | head -n 10 | awk '{ $1=""; print "- " substr($0, 2) }'
    else
        find "${ACTIVE_PROJECT_DIR}/specs/done" -name "SPEC-*.md" -printf '%T@ %p\n' 2>/dev/null \
            | sort -rn | head -n 10 | awk '{ $1=""; print "- " substr($0, 2) }'
    fi
fi
echo ""

echo "## Currently active specs (not yet shipped)"
find "${ACTIVE_PROJECT_DIR}/specs" -maxdepth 1 -name "SPEC-*.md" 2>/dev/null | sort | sed 's|^|- |' || true
echo ""

# --- Recent git activity (last 7 days) ---
if command -v git >/dev/null 2>&1 && [ -d "${REPO_ROOT}/.git" ]; then
    echo "## Git activity (last 7 days)"
    echo '```'
    git -C "$REPO_ROOT" log --since="7 days ago" --oneline --no-decorate 2>/dev/null || echo "(no git activity or not a git repo)"
    echo '```'
    echo ""
fi

cat <<'EOF'
## Report to produce

1. Stale decisions — DEC-* that should be superseded given recent learning.
   Flag with reasoning; don't supersede yet.

2. Missing constraints — patterns in reflections or PRs that should be
   formalized as constraints. Propose specific YAML entries.

3. Resolved questions — items in questions.yaml that recent work has
   answered. Flag which DEC-* would formalize each.

4. AGENTS.md drift — anything in the repo contradicting AGENTS.md.
   Propose specific edits.

5. Template improvements — friction visible in reflections. Be specific:
   "add field X to projects/_templates/spec.md because Y."

6. Stage health — active stages progressing? Any stalled? Any to rescope?

7. Cycle health — are any cycles consistently skipped or painful? Is Frame
   used as a kill gate? Are reflections getting mailed in?

8. (claude-only variant only) Session hygiene — any signs that build and
   verify are happening in the same Claude session? Reflections that
   consistently say "nothing was unclear" are a warning sign.

Tight report. Actionable in 10 minutes.

# ============================================================
# End of Weekly Review prompt
# ============================================================
EOF

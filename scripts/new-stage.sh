#!/usr/bin/env bash
# scripts/new-stage.sh — scaffold a new stage.
# Usage: new-stage.sh "short title" [PROJ-NNN]

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/_lib.sh"

require_initialized

TITLE="${1:-}"
PROJECT_ID="${2:-}"

if [ -z "$TITLE" ]; then
    die "Usage: just new-stage \"title\" [PROJ-NNN]"
fi

if [ -z "$PROJECT_ID" ]; then
    PROJECT_ID=$(get_active_project | awk -F- '{print $1"-"$2}')
fi

PROJECT_DIR=$(find "${REPO_ROOT}/projects" -maxdepth 1 -type d -name "${PROJECT_ID}-*" | head -n1)
if [ -z "$PROJECT_DIR" ]; then
    die "Project not found: ${PROJECT_ID}"
fi

STAGE_ID=$(next_id STAGE "${PROJECT_DIR}/stages")
SLUG=$(slugify "$TITLE")
STAGE_FILE="${PROJECT_DIR}/stages/${STAGE_ID}-${SLUG}.md"

if [ -f "$STAGE_FILE" ]; then
    die "Stage file already exists: ${STAGE_FILE}"
fi

TEMPLATE="${REPO_ROOT}/projects/_templates/stage.md"
if [ ! -f "$TEMPLATE" ]; then
    die "Template not found: ${TEMPLATE}. Did init run correctly?"
fi

cp "$TEMPLATE" "$STAGE_FILE"

sed_inplace() {
    if [ "$(uname)" = "Darwin" ]; then
        sed -i '' "$@"
    else
        sed -i "$@"
    fi
}

REPO_ID=$(get_repo_id)

sed_inplace "s|STAGE-XXX|${STAGE_ID}|g" "$STAGE_FILE"
sed_inplace "s|PROJ-XXX|${PROJECT_ID}|g" "$STAGE_FILE"
sed_inplace "s|<repo-id>|${REPO_ID}|g" "$STAGE_FILE"
sed_inplace "s|<Short Title — the coherent outcome>|${TITLE}|g" "$STAGE_FILE"
sed_inplace "s|YYYY-MM-DD|$(today)|g" "$STAGE_FILE"

success "Created ${STAGE_FILE}"
echo ""
echo "Next steps:"
echo "  1. Fill in the stage with Claude (use Prompt 1b: STAGE FRAME from FIRST_SESSION_PROMPTS.md)"
echo "  2. When ready, scaffold the first spec:"
echo "       just new-spec \"first task title\" ${STAGE_ID}"

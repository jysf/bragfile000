#!/usr/bin/env bash
# scripts/new-spec.sh — scaffold a new spec.
# Usage: new-spec.sh "short title" STAGE-NNN [PROJ-NNN]

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/_lib.sh"

require_initialized

TITLE="${1:-}"
STAGE_ID="${2:-}"
PROJECT_ID="${3:-}"

if [ -z "$TITLE" ] || [ -z "$STAGE_ID" ]; then
    die "Usage: just new-spec \"title\" STAGE-NNN [PROJ-NNN]"
fi

if [ -z "$PROJECT_ID" ]; then
    PROJECT_ID=$(get_active_project | sed -E 's/-.*//')
    # get_active_project returns PROJ-001-slug; we only want PROJ-001
    PROJECT_ID=$(get_active_project | awk -F- '{print $1"-"$2}')
fi

PROJECT_DIR=$(find "${REPO_ROOT}/projects" -maxdepth 1 -type d -name "${PROJECT_ID}-*" | head -n1)
if [ -z "$PROJECT_DIR" ]; then
    die "Project not found: ${PROJECT_ID}"
fi

# Verify stage exists in this project.
#
# The trailing `|| true` is load-bearing, not decoration. `find` exits 1 when
# ${PROJECT_DIR}/stages does not exist — which is precisely the state of a
# project whose first stage has not been framed yet — and under `set -e` plus
# `pipefail` that non-zero status propagates out of the command substitution
# and kills the script HERE, before the `die` below can name the cause. The
# symptom was a bare "exit code 1" with no message at all, strictly less
# legible than the raw `cp` failure this commit is fixing. `next_id` in
# _lib.sh already guards the same pipeline the same way.
STAGE_FILE=$(find "${PROJECT_DIR}/stages" -type f -name "${STAGE_ID}-*.md" 2>/dev/null | head -n1 || true)
if [ -z "$STAGE_FILE" ]; then
    die "Stage not found in ${PROJECT_ID}: ${STAGE_ID}"
fi

# Spec IDs are repo-global and monotonic (like DEC-* and STAGE-*), not
# reset per project — so search all projects, not just this one. This
# keeps every SPEC-NNN globally unique across the repo's history.
SPEC_ID=$(next_id SPEC "${REPO_ROOT}/projects")
SLUG=$(slugify "$TITLE")
SPEC_FILE="${PROJECT_DIR}/specs/${SPEC_ID}-${SLUG}.md"
VARIANT=$(get_variant)

if [ -f "$SPEC_FILE" ]; then
    die "Spec file already exists: ${SPEC_FILE}"
fi

# Choose template based on variant
if [ "$VARIANT" = "claude-plus-agents" ]; then
    TEMPLATE="${REPO_ROOT}/projects/_templates/spec.md"
else
    TEMPLATE="${REPO_ROOT}/projects/_templates/spec.md"
fi

if [ ! -f "$TEMPLATE" ]; then
    die "Template not found: ${TEMPLATE}. Did init run correctly?"
fi

# A project has no specs/ directory until its first spec is framed, so a
# missing directory here is the NORMAL state for spec one, not an error —
# create it rather than refusing. Same shape, same session as the new-stage.sh
# guard: the bare `cp` failed with a raw "No such file or directory" naming the
# *file* and never the missing parent. Hit during PROJ-008 framing (2026-09-05),
# immediately after the stages/ one, because the workaround for the first
# (mkdir -p by hand) did not generalise to the second.
SPEC_DIR=$(dirname "$SPEC_FILE")
if [ ! -d "$SPEC_DIR" ]; then
    mkdir -p "$SPEC_DIR" || die "Could not create spec directory: ${SPEC_DIR}"
    info "Created ${SPEC_DIR} (first spec in ${PROJECT_ID})"
fi

# Copy template, substitute placeholders
cp "$TEMPLATE" "$SPEC_FILE"

# Use sed to substitute. Portable across macOS/Linux using a wrapper.
sed_inplace() {
    if [ "$(uname)" = "Darwin" ]; then
        sed -i '' "$@"
    else
        sed -i "$@"
    fi
}

REPO_ID=$(get_repo_id)

sed_inplace "s|SPEC-XXX|${SPEC_ID}|g" "$SPEC_FILE"
sed_inplace "s|STAGE-XXX|${STAGE_ID}|g" "$SPEC_FILE"
sed_inplace "s|PROJ-XXX|${PROJECT_ID}|g" "$SPEC_FILE"
sed_inplace "s|<repo-id>|${REPO_ID}|g" "$SPEC_FILE"
sed_inplace "s|<Short Title>|${TITLE}|g" "$SPEC_FILE"
sed_inplace "s|YYYY-MM-DD|$(today)|g" "$SPEC_FILE"

success "Created ${SPEC_FILE}"
echo ""
echo "Next steps:"
echo "  1. Fill in the spec with Claude (use Prompt 2b: SPEC from FIRST_SESSION_PROMPTS.md)"
echo "  2. Update the stage's backlog in ${STAGE_FILE}"
echo "  3. When ready for build, run:"
echo "       just advance-cycle ${SPEC_ID} build"

#!/usr/bin/env bash
# scripts/archive-spec.sh — move a shipped spec to done/ and update stage backlog.
# Usage: archive-spec.sh SPEC-NNN

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/_lib.sh"

require_initialized

SPEC_ID="${1:-}"

if [ -z "$SPEC_ID" ]; then
    die "Usage: just archive-spec SPEC-NNN"
fi

SPEC_FILE=$(find_spec "$SPEC_ID")
if [ -z "$SPEC_FILE" ]; then
    die "Spec not found: ${SPEC_ID}"
fi

# Check cycle is ship
CYCLE=$(awk '/^---$/{f=!f; next} f && /^[[:space:]]+cycle:/{print $2; exit}' "$SPEC_FILE" 2>/dev/null || echo "")
if [ "$CYCLE" != "ship" ]; then
    warn "Spec cycle is '${CYCLE}', not 'ship'. Continue anyway? [y/N]"
    read -r answer
    if [ "$answer" != "y" ] && [ "$answer" != "Y" ]; then
        echo "Aborted."
        exit 0
    fi
fi

SPEC_DIR=$(dirname "$SPEC_FILE")
DONE_DIR="${SPEC_DIR}/done"
mkdir -p "$DONE_DIR"

SPEC_BASENAME=$(basename "$SPEC_FILE")
TARGET="${DONE_DIR}/${SPEC_BASENAME}"

mv "$SPEC_FILE" "$TARGET"
success "Archived: ${SPEC_FILE} → ${TARGET}"

# Try to update the parent stage's backlog.
# Get the stage ID from the spec's front-matter (project.stage field).
STAGE_ID=$(awk '/^---$/{f=!f; next} f && /^[[:space:]]+stage:/{print $2; exit}' "$TARGET" 2>/dev/null || echo "")
if [ -n "$STAGE_ID" ]; then
    STAGE_FILE=$(find_stage "$STAGE_ID")
    if [ -n "$STAGE_FILE" ]; then
        echo ""
        echo "Parent stage: ${STAGE_ID} (${STAGE_FILE})"
        echo "${DIM}Remember to update the stage's Spec Backlog section manually:"
        echo "  - Change '[ ] ${SPEC_ID}' to '[x] ${SPEC_ID} (shipped on $(today))'"
        echo "  - Update the count summary at the bottom of the backlog.${RESET}"
    fi
fi

# Check if this was the last active spec in the stage.
if [ -n "$STAGE_ID" ]; then
    REMAINING=$(find "$SPEC_DIR" -maxdepth 1 -name "SPEC-*.md" 2>/dev/null \
                | xargs -I{} awk -v sid="$STAGE_ID" '/^---$/{f=!f; next} f && /^[[:space:]]+stage:/ && $2 == sid {print FILENAME; exit}' {} \
                | wc -l | tr -d ' ')
    if [ "$REMAINING" = "0" ]; then
        echo ""
        echo "${GREEN}All specs for ${STAGE_ID} are shipped.${RESET}"
        echo "Consider running the Stage Ship prompt (Prompt 1c) in FIRST_SESSION_PROMPTS.md."
    fi
fi

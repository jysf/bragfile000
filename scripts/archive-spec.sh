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

# IDEMPOTENCY GUARD. find_spec searches recursively, so it also finds a spec
# that is ALREADY archived — and the move below derives DONE_DIR from the
# file's own directory, which would then be specs/done/done/. A second run
# therefore buried the spec one level deeper and silently broke the layout.
# Hit twice: SPEC-083 ship and SPEC-084 ship, both repaired by hand.
# Archiving is a one-way step; running it again is a no-op, not an error.
case "$SPEC_FILE" in
    */done/*)
        echo "${SPEC_ID} is already archived: ${SPEC_FILE}"
        echo "Nothing to do."
        exit 0
        ;;
esac

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

# Reject empty <answer> placeholders in reflection sections.
# Lesson earned at STAGE-004 ship (SPEC-019 orphaned reflection commit
# was recovered, but the same shape could otherwise ship empty).
if grep -q "^   — <answer>" "$SPEC_FILE"; then
    echo "✗ Reflection section has unanswered <answer> placeholders."
    echo "  Fill in the answers before archiving."
    exit 1
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
#
# TWO SOURCES, BOTH REQUIRED. Written spec FILES are only half the stage: a
# stage's Spec Backlog can carry items that have no spec file yet (planned,
# not yet framed). Counting files alone reports a stage complete while planned
# work remains — this claimed "All specs are shipped" three times against a
# non-empty backlog (twice in STAGE-021, once in STAGE-022) before being fixed.
if [ -n "$STAGE_ID" ]; then
    REMAINING=$(find "$SPEC_DIR" -maxdepth 1 -name "SPEC-*.md" 2>/dev/null \
                | xargs -I{} awk -v sid="$STAGE_ID" '/^---$/{f=!f; next} f && /^[[:space:]]+stage:/ && $2 == sid {print FILENAME; exit}' {} \
                | wc -l | tr -d ' ')

    # Unchecked "- [ ]" items in the stage file's ## Spec Backlog section,
    # including entries that name no SPEC id yet.
    UNPLANNED=0
    if [ -n "$STAGE_FILE" ] && [ -f "$STAGE_FILE" ]; then
        UNPLANNED=$(awk '/^## Spec Backlog/{f=1; next} f && /^## /{exit} f && /^- \[ \]/{c++} END{print c+0}' "$STAGE_FILE")
    fi

    if [ "$REMAINING" = "0" ] && [ "$UNPLANNED" = "0" ]; then
        echo ""
        echo "${GREEN}All specs for ${STAGE_ID} are shipped.${RESET}"
        echo "Consider running the Stage Ship prompt (Prompt 1c) in FIRST_SESSION_PROMPTS.md."
    elif [ "$REMAINING" = "0" ]; then
        echo ""
        echo "${GREEN}All WRITTEN specs for ${STAGE_ID} are shipped${RESET}, but its Spec"
        echo "Backlog still lists ${UNPLANNED} unchecked item(s) with no spec file yet."
        echo "${DIM}The stage is NOT complete. Frame the remaining item(s), or strike them"
        echo "from the backlog deliberately, before running the Stage Ship prompt.${RESET}"
    fi
fi

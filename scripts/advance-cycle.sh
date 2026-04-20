#!/usr/bin/env bash
# scripts/advance-cycle.sh — update a spec's task.cycle field.
# Usage: advance-cycle.sh SPEC-NNN <new-cycle>

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/_lib.sh"

require_initialized

SPEC_ID="${1:-}"
NEW_CYCLE="${2:-}"

if [ -z "$SPEC_ID" ] || [ -z "$NEW_CYCLE" ]; then
    die "Usage: just advance-cycle SPEC-NNN <frame|design|build|verify|ship>"
fi

case "$NEW_CYCLE" in
    frame|design|build|verify|ship) ;;
    *) die "Invalid cycle: ${NEW_CYCLE}. Must be one of: frame, design, build, verify, ship." ;;
esac

SPEC_FILE=$(find_spec "$SPEC_ID")
if [ -z "$SPEC_FILE" ]; then
    die "Spec not found: ${SPEC_ID}"
fi

OLD_CYCLE=$(awk '/^---$/{f=!f; next} f && /^[[:space:]]+cycle:/{print $2; exit}' "$SPEC_FILE" 2>/dev/null || echo "unknown")

update_frontmatter_scalar "$SPEC_FILE" "task.cycle" "$NEW_CYCLE"

success "Advanced ${SPEC_ID}: ${OLD_CYCLE} → ${NEW_CYCLE}"
echo "  File: ${SPEC_FILE}"

# Helpful next-step hints based on the new cycle.
echo ""
case "$NEW_CYCLE" in
    design)
        echo "Next: use Prompt 2b (Spec Design) in FIRST_SESSION_PROMPTS.md"
        ;;
    build)
        echo "Next: use Prompt 3 (Build) in FIRST_SESSION_PROMPTS.md"
        echo "  ${DIM}Claude-only variant reminder: start a NEW session for build.${RESET}"
        ;;
    verify)
        echo "Next: use Prompt 4 (Verify) in FIRST_SESSION_PROMPTS.md"
        echo "  ${DIM}Claude-only variant reminder: start a NEW session for verify.${RESET}"
        ;;
    ship)
        echo "Next: use Prompt 5 (Ship), then run 'just archive-spec ${SPEC_ID}'"
        ;;
esac

#!/usr/bin/env bash
# scripts/info.sh — print active project and variant.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/_lib.sh"

require_initialized

echo "Variant:        $(get_variant)"
echo "Active project: $(get_active_project)"
echo ""
echo "To switch active project, set ACTIVE_PROJECT env var:"
echo "  export ACTIVE_PROJECT=PROJ-002-whatever"

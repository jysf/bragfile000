#!/usr/bin/env bash
# scripts/_lib.sh — shared helpers sourced by other scripts.
# Sources are bash-only. Keep this minimal.

set -euo pipefail

REPO_ROOT="$(pwd)"

# Colors (fall back to no-op if terminal doesn't support color).
if [ -t 1 ] && command -v tput >/dev/null 2>&1; then
    BOLD=$(tput bold 2>/dev/null || printf '')
    DIM=$(tput dim 2>/dev/null || printf '')
    RED=$(tput setaf 1 2>/dev/null || printf '')
    GREEN=$(tput setaf 2 2>/dev/null || printf '')
    YELLOW=$(tput setaf 3 2>/dev/null || printf '')
    BLUE=$(tput setaf 4 2>/dev/null || printf '')
    RESET=$(tput sgr0 2>/dev/null || printf '')
else
    BOLD=''; DIM=''; RED=''; GREEN=''; YELLOW=''; BLUE=''; RESET=''
fi

die() {
    echo "${RED}ERROR:${RESET} $*" >&2
    exit 1
}

info() {
    echo "${BLUE}•${RESET} $*"
}

success() {
    echo "${GREEN}✓${RESET} $*"
}

warn() {
    echo "${YELLOW}⚠${RESET} $*"
}

# Require that the repo has been initialized (AGENTS.md at root).
require_initialized() {
    if [ ! -f "${REPO_ROOT}/AGENTS.md" ]; then
        die "Repo not initialized. Run 'just init' first."
    fi
}

# Get the active variant (claude-only or claude-plus-agents).
get_variant() {
    if [ -f "${REPO_ROOT}/.variant" ]; then
        cat "${REPO_ROOT}/.variant"
    else
        # Fallback: detect based on whether any project has a handoffs/ folder.
        if find "${REPO_ROOT}/projects" -maxdepth 2 -type d -name handoffs 2>/dev/null | grep -q .; then
            echo "claude-plus-agents"
        else
            echo "claude-only"
        fi
    fi
}

# Find the active project directory. Default heuristic: the lexically first
# project folder that doesn't start with PROJ-ZZZZ-archive or similar.
# Users can override by setting ACTIVE_PROJECT env var.
get_active_project() {
    if [ -n "${ACTIVE_PROJECT:-}" ]; then
        echo "${ACTIVE_PROJECT}"
        return
    fi
    # Look for the first PROJ-* directory that isn't the example.
    local first
    first=$(find "${REPO_ROOT}/projects" -maxdepth 1 -type d -name "PROJ-*" 2>/dev/null \
            | grep -v "example" | sort | head -n1)
    if [ -z "$first" ]; then
        # Fall back to the example if nothing else exists.
        first=$(find "${REPO_ROOT}/projects" -maxdepth 1 -type d -name "PROJ-*" 2>/dev/null \
                | sort | head -n1)
    fi
    if [ -z "$first" ]; then
        die "No projects found in ./projects/. Create one with 'just new-project' (or check GETTING_STARTED.md)."
    fi
    basename "$first"
}

# Return the next ID for a given prefix (SPEC, STAGE, PROJ, DEC, HANDOFF)
# across the entire repo (or within a project, for SPEC/STAGE/HANDOFF).
# Usage: next_id SPEC ./projects/PROJ-001-foo
next_id() {
    local prefix="$1"
    local search_dir="${2:-$REPO_ROOT}"
    local max
    max=$(find "$search_dir" -type f -name "${prefix}-*.md" 2>/dev/null \
          | sed -E "s|.*/${prefix}-([0-9]+).*|\\1|" \
          | sort -n \
          | tail -n1 || true)
    if [ -z "$max" ]; then
        printf "%s-%03d" "$prefix" 1
    else
        # Strip leading zeros for arithmetic, then reformat.
        max=$((10#$max))
        printf "%s-%03d" "$prefix" $((max + 1))
    fi
}

# Slugify a string. "Foo Bar Baz" -> "foo-bar-baz"
slugify() {
    echo "$1" \
        | tr '[:upper:]' '[:lower:]' \
        | sed -E 's/[^a-z0-9]+/-/g' \
        | sed -E 's/^-+|-+$//g'
}

# Find a spec file by ID. Searches all projects.
# Usage: find_spec SPEC-001
find_spec() {
    local spec_id="$1"
    find "${REPO_ROOT}/projects" -type f -name "${spec_id}-*.md" 2>/dev/null | head -n1
}

# Find a stage file by ID.
find_stage() {
    local stage_id="$1"
    find "${REPO_ROOT}/projects" -type f -name "${stage_id}-*.md" 2>/dev/null | head -n1
}

# Today's date in YYYY-MM-DD format.
today() {
    date +%Y-%m-%d
}

# Read the app's repo.id from .repo-context.yaml. Falls back to "my-app" if
# the file is missing or the key cannot be found. Assumes the standard
# template shape (metadata.repo.id at 4-space indent).
get_repo_id() {
    local ctx="${REPO_ROOT}/.repo-context.yaml"
    if [ ! -f "$ctx" ]; then
        echo "my-app"
        return
    fi
    local id
    id=$(grep -E '^    id:' "$ctx" | head -n1 \
         | sed -E 's/^    id:[[:space:]]*//; s/[[:space:]]*#.*$//; s/["'"'"']//g; s/[[:space:]]+$//')
    if [ -z "$id" ]; then
        echo "my-app"
    else
        echo "$id"
    fi
}

# Update a YAML front-matter scalar in a markdown file.
# Usage: update_frontmatter_scalar path/to/file.md task.cycle verify
# This is a deliberately simple awk-based updater for flat YAML. Requires
# the key to already exist in the front-matter.
update_frontmatter_scalar() {
    local file="$1"
    local key="$2"        # e.g. task.cycle or handoff.status
    local value="$3"

    # Split the key into top-level and leaf
    local top="${key%%.*}"
    local leaf="${key##*.}"

    # awk script that walks the front-matter (between the first two ---
    # delimiters) and replaces the target key's value.
    awk -v top="$top" -v leaf="$leaf" -v val="$value" '
        BEGIN { in_fm = 0; fm_seen = 0; in_top = 0 }
        /^---$/ {
            if (!fm_seen) { in_fm = 1; fm_seen = 1 }
            else if (in_fm) { in_fm = 0 }
            print; next
        }
        in_fm {
            if ($0 ~ "^" top ":") { in_top = 1; print; next }
            if ($0 ~ "^[a-zA-Z_]+:") { in_top = 0 }
            if (in_top && $0 ~ "^[[:space:]]+" leaf ":") {
                sub(/:[[:space:]]*.*$/, ": " val)
            }
        }
        { print }
    ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
}

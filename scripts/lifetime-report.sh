#!/usr/bin/env bash
# scripts/lifetime-report.sh — print the Lifetime Report prompt with the
# whole-repo history pre-loaded.
#
# Unlike `weekly-review` (active project, last 7 days) this looks across
# EVERY project and stage and assembles the raw lifetime data — release
# timeline, per-project dates, specs-by-stage, decision log — then prints a
# synthesis prompt. bragfile does the aggregation; an LLM writes the
# narrative arc (same pipe posture as review/summary/story).
#
# Usage:
#   just lifetime-report              # print the prompt to stdout
#   just lifetime-report | pbcopy     # ...and copy it for a Claude session

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/_lib.sh"

require_initialized

REPO_ID=$(get_repo_id)
VARIANT=$(get_variant)
ACTIVE_PROJECT=$(get_active_project)
GENERATED=$(today)

cat <<'EOF'
# ============================================================
# Lifetime Report — copy everything below this line into Claude
# ============================================================

Produce a full lifetime report on this repo: the arc of every project,
stage, spec, decision, and release since the first commit. Not a status
snapshot (`just status` already does that) — a narrative of how the tool
got from nothing to where it is now. Use the pre-loaded data below;
read the project briefs and CHANGELOG for the "why" behind each wave.

Head the report with `# <Repo> — Lifetime Report` and, directly under it, a
`Generated: <date>` line — use the `Generated:` date given in the context block
below the prompt (this is the date the report was produced; the git-span and
project dates below are the history it covers, which is a different thing).

Structure to produce:
  1. What this is — one paragraph on the product and the lifecycle model.
  2. Lifetime at a glance — a metrics table (lifespan, #projects, #stages,
     #specs shipped/deferred, #decisions, #releases, open low-confidence debt).
  3. Release timeline — version / date / theme / which project shipped it.
  4. The waves — one tight paragraph per project: dates, stage/spec counts,
     what it delivered, and WHY now (pull from each brief's "Why Now").
  5. Arc of the whole thing — the progression across waves in one line,
     plus how each wave made the next cheap.
  6. Decision debt & loose ends — low-confidence DECs, deferred/cut specs,
     stale flags.
  7. What's next — the next unframed wave (see docs/roadmap/ if present).

Keep it skimmable: tables where the data is tabular, prose only for the arc.
EOF

echo ""
echo "Generated: ${GENERATED}"
echo "Repo: ${REPO_ID}   ·   variant: ${VARIANT}   ·   active project: ${ACTIVE_PROJECT}"
echo ""

# --- Repo-level context to read ---
echo "## Repo-level files to read (for the 'why')"
echo "- AGENTS.md"
echo "- CHANGELOG.md"
for f in projects/*/brief.md; do
    [ -e "$REPO_ROOT/$f" ] && echo "- $f"
done
if [ -d "${REPO_ROOT}/docs/roadmap" ]; then
    find "${REPO_ROOT}/docs/roadmap" -name "*.md" 2>/dev/null | sort \
        | sed "s|^${REPO_ROOT}/|- |"
fi
echo ""

# --- Per-project timeline (created / shipped / status from brief frontmatter) ---
echo "## Projects (created → shipped, from brief frontmatter)"
for brief in "${REPO_ROOT}"/projects/*/brief.md; do
    [ -e "$brief" ] || continue
    proj_dir=$(basename "$(dirname "$brief")")
    created=$(grep -m1 -E '^created_at:'  "$brief" | sed 's/created_at:[[:space:]]*//' || true)
    shipped=$(grep -m1 -E '^shipped_at:'  "$brief" | sed 's/shipped_at:[[:space:]]*//' || true)
    status=$(grep -m1  -E '^[[:space:]]*status:' "$brief" | sed 's/.*status:[[:space:]]*//' | awk '{print $1}' || true)
    echo "- ${proj_dir}: created ${created:-?} → shipped ${shipped:-—}  (status: ${status:-?})"
done
echo ""

# --- Releases: git tags + CHANGELOG headings ---
echo "## Releases (git tags)"
if command -v git >/dev/null 2>&1 && [ -d "${REPO_ROOT}/.git" ]; then
    git -C "$REPO_ROOT" tag --sort=version:refname 2>/dev/null | sed 's|^|- |' || true
fi
echo ""
echo "## Release headings (from CHANGELOG.md)"
grep -E '^## ' "${REPO_ROOT}/CHANGELOG.md" 2>/dev/null | sed 's|^## |- |' || true
echo ""

# --- Decisions across all projects ---
echo "## Decisions (across all projects)"
find "${REPO_ROOT}/decisions" -name "DEC-*.md" 2>/dev/null | sort \
    | sed "s|^${REPO_ROOT}/decisions/|- |" || true
echo ""

# --- Git span ---
if command -v git >/dev/null 2>&1 && [ -d "${REPO_ROOT}/.git" ]; then
    echo "## Git span"
    first=$(git -C "$REPO_ROOT" log --reverse --format='%ad' --date=short 2>/dev/null | head -1)
    last=$(git  -C "$REPO_ROOT" log          --format='%ad' --date=short 2>/dev/null | head -1)
    count=$(git -C "$REPO_ROOT" rev-list --count HEAD 2>/dev/null)
    echo "- first commit: ${first:-?}   last commit: ${last:-?}   (${count:-?} commits)"
    echo "  (note: project dates in briefs predate the git history if the repo was re-inited)"
    echo ""
fi

# --- The mechanical aggregates the report tables are built from ---
echo "## Repo status snapshot"
echo '```'
"${SCRIPT_DIR}/status.sh"
echo '```'
echo ""

echo "## Specs by stage (full history)"
echo '```'
"${SCRIPT_DIR}/specs-by-stage.sh"
echo '```'
echo ""

cat <<'EOF'
# ============================================================
# End of Lifetime Report prompt
# ============================================================
EOF

#!/usr/bin/env bash
# scripts/status.sh — print repo state report.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/_lib.sh"

require_initialized

VARIANT=$(get_variant)
ACTIVE_PROJECT=$(get_active_project)
ACTIVE_PROJECT_DIR="${REPO_ROOT}/projects/${ACTIVE_PROJECT}"

# Optional `activity:` under `project:` in the brief front-matter — the type of
# work currently active within the project (requirements/design/build/test/…).
# Coarse `status` stays what the resolver keys on; this is human-facing detail.
ACTIVE_ACTIVITY=$(awk '
    /^---$/ { f = !f; next }
    f && /^project:/ { inproj = 1; next }
    f && inproj && /^[a-zA-Z_]+:/ { inproj = 0 }
    f && inproj && /^[[:space:]]+activity:/ { print $2; exit }
' "${ACTIVE_PROJECT_DIR}/brief.md" 2>/dev/null || echo "")

echo "${BOLD}Repo status${RESET}"
echo ""
echo "  Variant:         ${VARIANT}"
echo "  Active project:  ${ACTIVE_PROJECT}"
if [ -n "$ACTIVE_ACTIVITY" ]; then
    echo "  Activity:        ${ACTIVE_ACTIVITY}"
fi
echo ""

# --- Active project: stages ---
echo "${BOLD}Stages in ${ACTIVE_PROJECT}${RESET}"
stages_dir="${ACTIVE_PROJECT_DIR}/stages"
if [ -d "$stages_dir" ]; then
    for s in "${stages_dir}"/STAGE-*.md; do
        [ -f "$s" ] || continue
        sname=$(basename "$s" .md)
        pstatus=$(awk '/^---$/{f=!f; next} f && /^[[:space:]]+status:/{print $2; exit}' "$s" 2>/dev/null || echo "unknown")
        printf "  %-44s  status: %s\n" "$sname" "$pstatus"
    done
else
    echo "  ${DIM}(no stages dir yet)${RESET}"
fi
echo ""

# --- Active project: specs by cycle ---
echo "${BOLD}Specs in ${ACTIVE_PROJECT} by cycle${RESET}"
specs_dir="${ACTIVE_PROJECT_DIR}/specs"
if [ -d "$specs_dir" ]; then
    for cycle in frame design build verify ship; do
        count=0
        names=""
        for f in "${specs_dir}"/SPEC-*.md; do
            [ -f "$f" ] || continue
            spec_cycle=$(awk '/^---$/{f=!f; next} f && /^[[:space:]]+cycle:/{print $2; exit}' "$f" 2>/dev/null || echo "")
            if [ "$spec_cycle" = "$cycle" ]; then
                count=$((count + 1))
                names="${names}    - $(basename "$f" .md)\n"
            fi
        done
        # Also count done/ as ship
        if [ "$cycle" = "ship" ] && [ -d "${specs_dir}/done" ]; then
            for f in "${specs_dir}/done"/SPEC-*.md; do
                [ -f "$f" ] || continue
                count=$((count + 1))
                names="${names}    - $(basename "$f" .md) ${DIM}(archived)${RESET}\n"
            done
        fi
        printf "  ${BOLD}%-8s${RESET} (%d)\n" "$cycle" "$count"
        if [ -n "$names" ]; then
            printf "%b" "$names"
        fi
    done
else
    echo "  ${DIM}(no specs yet)${RESET}"
fi
echo ""

# --- Low-confidence decisions ---
echo "${BOLD}Low-confidence decisions (< 0.7)${RESET}"
decisions_dir="${REPO_ROOT}/decisions"
found_any=false
if [ -d "$decisions_dir" ]; then
    for d in "${decisions_dir}"/DEC-*.md; do
        [ -f "$d" ] || continue
        conf=$(awk '/^---$/{f=!f; next} f && /^[[:space:]]+confidence:/{print $2; exit}' "$d" 2>/dev/null || echo "")
        if [ -n "$conf" ]; then
            # Use awk for float comparison (portable)
            low=$(awk -v c="$conf" 'BEGIN { print (c + 0 < 0.7) ? "1" : "0" }')
            if [ "$low" = "1" ]; then
                printf "  %-42s  confidence: %s\n" "$(basename "$d" .md)" "$conf"
                found_any=true
            fi
        fi
    done
fi
if [ "$found_any" = "false" ]; then
    echo "  ${DIM}(none — or no decisions yet)${RESET}"
fi
echo ""

# --- Stale specs (no commits on their branch in 7 days, approximate) ---
echo "${BOLD}Possibly stale specs${RESET}"
echo "  ${DIM}(heuristic: specs in build/verify with file mtime > 7 days)${RESET}"
found_stale=false
if [ -d "$specs_dir" ]; then
    for f in "${specs_dir}"/SPEC-*.md; do
        [ -f "$f" ] || continue
        cycle=$(awk '/^---$/{fm=!fm; next} fm && /^[[:space:]]+cycle:/{print $2; exit}' "$f" 2>/dev/null || echo "")
        if [ "$cycle" = "build" ] || [ "$cycle" = "verify" ]; then
            # Age in days (portable across macOS and Linux).
            if [ "$(uname)" = "Darwin" ]; then
                age_days=$(( ( $(date +%s) - $(stat -f %m "$f") ) / 86400 ))
            else
                age_days=$(( ( $(date +%s) - $(stat -c %Y "$f") ) / 86400 ))
            fi
            if [ "$age_days" -gt 7 ]; then
                printf "  %-40s  cycle: %-8s  age: %d days\n" "$(basename "$f" .md)" "$cycle" "$age_days"
                found_stale=true
            fi
        fi
    done
fi
if [ "$found_stale" = "false" ]; then
    echo "  ${DIM}(none)${RESET}"
fi
echo ""

# --- Summary counts ---
total_specs=$(find "${ACTIVE_PROJECT_DIR}/specs" -name "SPEC-*.md" 2>/dev/null | wc -l | tr -d ' ')
shipped_specs=$(find "${ACTIVE_PROJECT_DIR}/specs/done" -name "SPEC-*.md" 2>/dev/null | wc -l | tr -d ' ')
total_decisions=$(find "$decisions_dir" -name "DEC-*.md" 2>/dev/null | wc -l | tr -d ' ')
echo "${BOLD}Summary${RESET}"
echo "  Total specs in ${ACTIVE_PROJECT}:     ${total_specs}"
echo "  Shipped (archived):                   ${shipped_specs}"
echo "  Total decisions (across all projects): ${total_decisions}"
echo ""

# --- All projects (overview — kept at the bottom) ---
echo "${BOLD}All projects${RESET}"
for p in "${REPO_ROOT}"/projects/PROJ-*; do
    [ -d "$p" ] || continue
    pname=$(basename "$p")
    brief="${p}/brief.md"
    status="unknown"
    if [ -f "$brief" ]; then
        # Grep for "status:" nested under "project:" in the front-matter
        status=$(awk '
            /^---$/ { f = !f; next }
            f && /^project:/ { inproj = 1; next }
            f && inproj && /^[a-zA-Z_]+:/ { inproj = 0 }
            f && inproj && /^[[:space:]]+status:/ { print $2; exit }
        ' "$brief" 2>/dev/null || echo "unknown")
    fi
    marker=" "
    if [ "$pname" = "$ACTIVE_PROJECT" ]; then marker="${GREEN}*${RESET}"; fi
    st_count=$(find "${p}/stages" -name "STAGE-*.md" 2>/dev/null | wc -l | tr -d ' ')
    sp_total=$(find "${p}/specs" -name "SPEC-*.md" 2>/dev/null | wc -l | tr -d ' ')
    sp_done=$(find "${p}/specs/done" -name "SPEC-*.md" 2>/dev/null | wc -l | tr -d ' ')
    printf "  %s %-40s  status: %-9s  ${DIM}%s stages · %s/%s specs shipped${RESET}\n" \
        "$marker" "$pname" "$status" "$st_count" "$sp_done" "$sp_total"
done
echo ""

# --- Completed in prior projects (the accomplishment trail) ---
# Non-active projects only (the active project gets its own deep-dive
# above). Stage granularity, not every spec — shipped stages + dates.
echo "${BOLD}Completed in prior projects${RESET}"
prior_found=false
for p in "${REPO_ROOT}"/projects/PROJ-*; do
    [ -d "$p" ] || continue
    pname=$(basename "$p")
    [ "$pname" = "$ACTIVE_PROJECT" ] && continue
    pstages_dir="${p}/stages"
    [ -d "$pstages_dir" ] || continue
    stage_lines=""
    for s in "${pstages_dir}"/STAGE-*.md; do
        [ -f "$s" ] || continue
        sstatus=$(awk '/^---$/{f=!f; next} f && /^[[:space:]]+status:/{print $2; exit}' "$s" 2>/dev/null || echo "")
        sshipped=$(awk '/^---$/{f=!f; next} f && /^shipped_at:/{print $2; exit}' "$s" 2>/dev/null || echo "")
        sname=$(basename "$s" .md)
        if [ -n "$sshipped" ] && [ "$sshipped" != "null" ]; then
            stage_lines="${stage_lines}    - ${sname}  ${DIM}(shipped ${sshipped})${RESET}\n"
        else
            stage_lines="${stage_lines}    - ${sname}  ${DIM}(${sstatus})${RESET}\n"
        fi
    done
    [ -z "$stage_lines" ] && continue
    st_count=$(find "$pstages_dir" -name "STAGE-*.md" 2>/dev/null | wc -l | tr -d ' ')
    sh_count=$(find "${p}/specs/done" -name "SPEC-*.md" 2>/dev/null | wc -l | tr -d ' ')
    pshipped=$(awk '/^---$/{f=!f; next} f && /^shipped_at:/{print $2; exit}' "${p}/brief.md" 2>/dev/null || echo "")
    if [ -n "$pshipped" ] && [ "$pshipped" != "null" ]; then
        printf "  ${BOLD}%s${RESET}  ${DIM}(shipped %s — %s stages, %s specs)${RESET}\n" "$pname" "$pshipped" "$st_count" "$sh_count"
    else
        printf "  ${BOLD}%s${RESET}  ${DIM}(%s stages, %s specs shipped)${RESET}\n" "$pname" "$st_count" "$sh_count"
    fi
    printf "%b" "$stage_lines"
    prior_found=true
done
if [ "$prior_found" = "false" ]; then
    echo "  ${DIM}(none — this is the first project)${RESET}"
fi
echo ""

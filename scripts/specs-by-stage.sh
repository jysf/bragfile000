#!/usr/bin/env bash
# scripts/specs-by-stage.sh — print all specs grouped by stage with
# ship dates and complexity sizes, derived from the stage files'
# Spec Backlog sections.
#
# Usage: just specs-by-stage   (or ./scripts/specs-by-stage.sh)
#
# Reads from: projects/PROJ-*-*/stages/STAGE-*.md
# Output: human-readable text report on stdout.

set -eu

# Find all stages across all projects (handles future PROJ-002+ cleanly).
shopt -s nullglob 2>/dev/null || true
stages=(projects/PROJ-*/stages/STAGE-*.md)

if [ ${#stages[@]} -eq 0 ]; then
    echo "No stage files found under projects/PROJ-*/stages/"
    exit 0
fi

printf "Specs by stage\n"
printf "==============\n"

total_shipped=0
total_deferred=0
total_pending=0

for stage in "${stages[@]}"; do
    sname=$(basename "$stage" .md | cut -d- -f1-2)
    sst=$(grep -E "^  status:" "$stage" | head -1 | awk '{print $2}')
    sdate=$(grep -E "^shipped_at:" "$stage" | awk '{print $2}')
    [ "$sdate" = "null" ] && sdate="—"

    printf "\n=== %s [%s, %s] ===\n" "$sname" "$sst" "$sdate"

    while IFS= read -r line; do
        sid=$(echo "$line" | grep -oE 'SPEC-[0-9]+' | head -1)
        shipped=$(echo "$line" | grep -oE '20[0-9][0-9]-[0-9]+-[0-9]+' | head -1)
        ssize=$(echo "$line" | grep -oE '\*\*[SML]\*\*' | head -1 | sed 's/\*//g')
        [ -z "$ssize" ] && ssize=$(echo "$line" | grep -oE '\([SML]\)' | head -1 | sed 's/[()]//g')
        [ -z "$ssize" ] && ssize="?"

        if echo "$line" | grep -q '^- \[x\]'; then
            total_shipped=$((total_shipped + 1))
            printf "  %-10s  %-10s  %-12s  %s\n" "$sid" "shipped" "${shipped:-—}" "$ssize"
        elif echo "$line" | grep -q '^- \[~\]'; then
            total_deferred=$((total_deferred + 1))
            printf "  %-10s  %-10s  %-12s  %s\n" "$sid" "deferred" "—" "$ssize"
        elif echo "$line" | grep -q '^- \[ \]'; then
            total_pending=$((total_pending + 1))
            printf "  %-10s  %-10s  %-12s  %s\n" "$sid" "pending" "—" "$ssize"
        fi
    done < <(grep -E "^- \[[x~ ]\] SPEC-" "$stage")
done

printf "\n— Totals: %d shipped, %d pending, %d deferred\n" \
    "$total_shipped" "$total_pending" "$total_deferred"

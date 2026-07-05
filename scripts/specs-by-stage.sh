#!/usr/bin/env bash
# scripts/specs-by-stage.sh — print all specs grouped by stage with
# ship dates and complexity sizes, derived from the stage files'
# Spec Backlog sections.
#
# Usage: just specs-by-stage [--no-names]   (or ./scripts/specs-by-stage.sh)
#
# By default each spec's name (its slug, from the spec filename) is shown
# in a trailing column. Pass --no-names for the compact id-only view.
#
# Reads from: projects/PROJ-*-*/stages/STAGE-*.md
#         and: projects/PROJ-*-*/specs/{,done/}SPEC-*.md  (for names)
# Output: human-readable text report on stdout.

set -eu

show_names=1
for arg in "$@"; do
    case "$arg" in
        --no-names) show_names=0 ;;
        *) echo "Unknown argument: $arg (usage: specs-by-stage [--no-names])" >&2; exit 2 ;;
    esac
done

# Find all stages across all projects (handles future PROJ-002+ cleanly).
shopt -s nullglob 2>/dev/null || true
stages=(projects/PROJ-*/stages/STAGE-*.md)

# spec_name SPEC-NNN — derive a human-readable name from the spec's
# filename slug (looks in both active specs/ and archived specs/done/).
# Prints the empty string when no spec file exists (e.g. deferred specs).
spec_name() {
    local id="$1"
    local matches=(projects/PROJ-*/specs/"$id"-*.md projects/PROJ-*/specs/done/"$id"-*.md)
    [ ${#matches[@]} -eq 0 ] && return 0
    local fn
    fn=$(basename "${matches[0]}" .md)
    printf '%s' "${fn#"$id"-}" | tr '-' ' '
}

# spec_size SPEC-NNN — the authoritative complexity (S/M/L) from the spec
# file's `complexity:` front-matter (active specs/ or archived specs/done/).
# Reading front-matter — not the stage backlog prose — is robust: backlog
# lines wrap across multiple lines and write the size as "**M (headline) —"
# or "**XS/S —", neither of which is a fixed, greppable token. Prints ""
# when no spec file exists yet (e.g. a pending/deferred stub).
spec_size() {
    local id="$1"
    local matches=(projects/PROJ-*/specs/"$id"-*.md projects/PROJ-*/specs/done/"$id"-*.md)
    [ ${#matches[@]} -eq 0 ] && return 0
    grep -m1 -E '^  complexity:' "${matches[0]}" | sed 's/#.*//' | awk '{print $2}'
}

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
        ssize=$(spec_size "$sid")
        [ -z "$ssize" ] && ssize="?"

        # Trailing column: just the size, or "size  name" when names are on.
        if [ "$show_names" -eq 1 ]; then
            name=$(spec_name "$sid")
            [ -z "$name" ] && name="—"
            tail=$(printf "%-4s%s" "$ssize" "$name")
        else
            tail="$ssize"
        fi

        if echo "$line" | grep -q '^- \[x\]'; then
            total_shipped=$((total_shipped + 1))
            printf "  %-10s  %-10s  %-12s  %s\n" "$sid" "shipped" "${shipped:-—}" "$tail"
        elif echo "$line" | grep -q '^- \[~\]'; then
            total_deferred=$((total_deferred + 1))
            printf "  %-10s  %-10s  %-12s  %s\n" "$sid" "deferred" "—" "$tail"
        elif echo "$line" | grep -q '^- \[ \]'; then
            total_pending=$((total_pending + 1))
            printf "  %-10s  %-10s  %-12s  %s\n" "$sid" "pending" "—" "$tail"
        fi
    done < <(grep -E "^- \[[x~ ]\] SPEC-" "$stage")
done

printf "\n— Totals: %d shipped, %d pending, %d deferred\n" \
    "$total_shipped" "$total_pending" "$total_deferred"

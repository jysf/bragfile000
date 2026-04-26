#!/usr/bin/env bash
# scripts/claude-code-post-session.sh — example Claude Code session-end hook.
#
# Reads a session summary from stdin, structures it as a JSON object
# matching docs/brag-entry.schema.json, and prints a candidate
# `brag add --json` payload to stdout (plus a hint to stderr explaining
# what to do with it).
#
# This script does NOT auto-execute `brag add`. It honours BRAG.md's
# approval loop: you (the user) review the candidate JSON, decide if
# the moment is brag-worthy, then pipe the JSON to `brag add --json`
# yourself.
#
# Wiring: copy this file wherever your Claude Code config wants it.
# As of late 2025, Claude Code reads hook config from
# ~/.claude/settings.json; consult Claude Code docs for the current
# hook-config shape. This script makes no assumptions about the
# invocation interface beyond "stdin carries the session summary".
#
# Dependencies: bash, jq, brag (on $PATH).

set -eu

if ! command -v jq >/dev/null 2>&1; then
    printf 'claude-code-post-session: jq is required but not installed (see https://stedolan.github.io/jq/)\n' >&2
    exit 2
fi

if ! command -v brag >/dev/null 2>&1; then
    printf 'claude-code-post-session: brag is required but not on $PATH (install via brew or `just install`)\n' >&2
    exit 2
fi

# Read the entire stdin into a variable. Empty stdin → no-op exit 0.
SUMMARY=$(cat)
if [ -z "${SUMMARY}" ]; then
    printf 'claude-code-post-session: no summary on stdin; skipping\n' >&2
    exit 0
fi

# Derive a candidate title heuristically: the first non-empty line of
# stdin, trimmed, capped at 100 characters. The user reviews and
# refines before approving.
TITLE=$(printf '%s\n' "${SUMMARY}" | awk 'NF { print; exit }' | cut -c1-100)
if [ -z "${TITLE}" ]; then
    printf 'claude-code-post-session: could not derive a title from stdin; skipping\n' >&2
    exit 0
fi

# Build the JSON payload via jq (handles string escaping safely).
# Fields included: title (required) + description (the full summary)
# + type (default "shipped"). The user can refine project / tags /
# impact before piping to `brag add --json`.
PAYLOAD=$(jq -n \
    --arg title "${TITLE}" \
    --arg description "${SUMMARY}" \
    --arg type "shipped" \
    '{title: $title, description: $description, type: $type}')

# Print the candidate payload to stdout (pipeable). Print the hint to
# stderr (human-facing). This mirrors the brag binary's stdout-is-for-
# data-stderr-is-for-humans contract by example.
printf '%s\n' "${PAYLOAD}"
printf '\nclaude-code-post-session: candidate JSON above. To capture, run:\n' >&2
printf '    <prev-stdout> | brag add --json\n' >&2
printf '(Or copy the JSON, refine project/tags/impact, then pipe to `brag add --json` manually.)\n' >&2

exit 0

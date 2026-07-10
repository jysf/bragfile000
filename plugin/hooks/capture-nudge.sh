#!/usr/bin/env bash
# capture-nudge.sh — brag plugin Stop hook.
#
# Fires on every Stop (Claude finishing a turn), but nudges AT MOST ONCE per
# session, and ONLY after a git commit lands in the session's cwd — a
# lightweight "you plausibly shipped something" signal. The nudge is
# AGENT-FACING context (Claude then proposes a brag for the user's approval
# per BRAG.md); this hook NEVER runs `brag add`. Silence it with
# BRAG_CAPTURE_NUDGE=off.
#
# Contract (Claude Code Stop hook): reads a JSON payload on stdin
# ({session_id, cwd, hook_event_name, stop_hook_active}); on a fire it emits
# a hookSpecificOutput.additionalContext JSON object on stdout and exits 0.
# Every non-fire path exits 0 silently so the hook is quiet and never blocks.
#
# Degradation: missing jq, non-git cwd, or no new commit -> silent exit 0.
set -eu

# 1. Silence switch.
case "${BRAG_CAPTURE_NUDGE:-}" in
    off|0|false|no) exit 0 ;;
esac

# 2. jq is required to parse the payload; degrade quietly if absent.
command -v jq >/dev/null 2>&1 || exit 0

PAYLOAD=$(cat)
SESSION_ID=$(printf '%s' "$PAYLOAD" | jq -r '.session_id // empty')
CWD=$(printf '%s' "$PAYLOAD" | jq -r '.cwd // empty')
[ -n "$SESSION_ID" ] || exit 0
[ -n "$CWD" ] || exit 0

# 3. Only meaningful inside a git repo; HEAD is the "shipped" signal.
HEAD=$(git -C "$CWD" rev-parse HEAD 2>/dev/null || true)
[ -n "$HEAD" ] || exit 0

STATE_DIR="${BRAG_STATE_DIR:-$HOME/.bragfile}/capture-nudge"
mkdir -p "$STATE_DIR" 2>/dev/null || exit 0
MARKER="$STATE_DIR/$SESSION_ID"

# 4. First Stop of the session: record the baseline HEAD, never nudge yet.
if [ ! -f "$MARKER" ]; then
    printf 'baseline=%s\n' "$HEAD" > "$MARKER"
    exit 0
fi

# 5. Already nudged this session -> stay silent.
grep -q '^nudged$' "$MARKER" 2>/dev/null && exit 0

# 6. Nudge once, only if a commit landed since the baseline.
BASELINE=$(sed -n 's/^baseline=//p' "$MARKER" | head -1)
if [ -n "$BASELINE" ] && [ "$HEAD" != "$BASELINE" ]; then
    printf 'nudged\n' >> "$MARKER"
    jq -cn --arg session "$SESSION_ID" '{
        hookSpecificOutput: {
            hookEventName: "Stop",
            additionalContext: "A commit landed during this session. If something brag-worthy shipped, draft a brag entry for the user'"'"'s approval per BRAG.md (you can use the /brag:brag command): a required action-verb title plus optional project, type, tags, and a concrete impact. Stamp provenance as reserved tags agent:<name> and model:<id>, and pass session:<id> using this session'"'"'s id (\($session)) so the work is joinable later; include cost:<usd> and tokens:<n> only if you have real figures — never estimate them. Do NOT run `brag add` until the user explicitly approves."
        }
    }'
fi
exit 0

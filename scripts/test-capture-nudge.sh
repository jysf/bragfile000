#!/usr/bin/env bash
# scripts/test-capture-nudge.sh — behavior harness for
# plugin/hooks/capture-nudge.sh (SPEC-041 / DEC-025).
#
# Builds a throwaway git repo + a temp BRAG_STATE_DIR + a PATH-shadowing
# `brag` stub that records invocations to $BRAG_SENTINEL, then drives the
# hook with crafted Stop-event JSON payloads across every fire/silence path.
# Exits 0 iff all assertions pass.
#
# Run via `just test-hook`.

set -eu

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
HOOK="$REPO_ROOT/plugin/hooks/capture-nudge.sh"

FAIL_COUNT=0

if ! command -v jq >/dev/null 2>&1; then
    printf 'test-capture-nudge: jq is required but not installed\n' >&2
    exit 2
fi

ok() {
    printf 'OK:   %s\n' "$1"
}

fail() {
    printf 'FAIL: %s: %s\n' "$1" "$2"
    FAIL_COUNT=$((FAIL_COUNT + 1))
}

if [ ! -f "$HOOK" ]; then
    fail "H0" "hook does not exist: $HOOK"
    printf '\nFAILED: %d assertion(s) failed.\n' "$FAIL_COUNT" >&2
    exit 1
fi
if [ ! -x "$HOOK" ]; then
    fail "H0" "hook is not executable: $HOOK"
fi

# --- fixture setup ---

TMPDIR_ROOT=$(mktemp -d)
trap 'rm -rf "$TMPDIR_ROOT"' EXIT

REPO="$TMPDIR_ROOT/repo"
NOTGIT="$TMPDIR_ROOT/notgit"
STATE_DIR="$TMPDIR_ROOT/state"
BIN_DIR="$TMPDIR_ROOT/bin"
SENTINEL="$TMPDIR_ROOT/sentinel"

mkdir -p "$REPO" "$NOTGIT" "$STATE_DIR" "$BIN_DIR"

git -C "$REPO" init -q
git -C "$REPO" config user.email test@example.com
git -C "$REPO" config user.name test
echo "one" > "$REPO/file.txt"
git -C "$REPO" add -A
git -C "$REPO" commit -q -m "init"

# PATH-shadowing brag stub: records every invocation to $SENTINEL. If this
# file ever appears, the hook broke the approval loop by running `brag`.
cat > "$BIN_DIR/brag" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${BRAG_SENTINEL:?}"
exit 0
EOF
chmod +x "$BIN_DIR/brag"

export PATH="$BIN_DIR:$PATH"
export BRAG_SENTINEL="$SENTINEL"
export BRAG_STATE_DIR="$STATE_DIR"

payload() {
    sid="$1"; cwd="$2"
    jq -n --arg sid "$sid" --arg cwd "$cwd" \
        '{session_id:$sid, cwd:$cwd, hook_event_name:"Stop", stop_hook_active:false}'
}

run_hook() {
    sid="$1"; cwd="$2"
    payload "$sid" "$cwd" | "$HOOK"
}

# ===== H1 — first Stop of a session: baseline recorded, silent =====

SID1="d838146a-116c-4021-9099-892972ca5c5a"  # UUID-shaped: a short id hid the H9 bug
out=$(run_hook "$SID1" "$REPO")
if [ -z "$out" ]; then
    ok "H1"
else
    fail "H1" "expected empty stdout on first Stop, got: $out"
fi

# ===== H2 — second Stop, HEAD unchanged: silent =====

out=$(run_hook "$SID1" "$REPO")
if [ -z "$out" ]; then
    ok "H2"
else
    fail "H2" "expected empty stdout when HEAD unchanged, got: $out"
fi

# ===== H3 — a commit lands, next Stop: fires exactly one nudge =====

echo "two" >> "$REPO/file.txt"
git -C "$REPO" add -A
git -C "$REPO" commit -q -m "change"

out=$(run_hook "$SID1" "$REPO")
event=$(printf '%s' "$out" | jq -r '.hookSpecificOutput.hookEventName // empty' 2>/dev/null || true)
ctx=$(printf '%s' "$out" | jq -r '.hookSpecificOutput.additionalContext // empty' 2>/dev/null || true)
if [ "$event" = "Stop" ] && printf '%s' "$ctx" | grep -qi 'brag'; then
    ok "H3"
else
    fail "H3" "expected valid hookSpecificOutput nudge, got: $out"
fi

# The two hashes the hook actually holds at nudge time: BASELINE (stored on the
# first Stop) and HEAD (read fresh). H9a asserts neither reaches the output.
HEAD_SHA=$(git -C "$REPO" rev-parse HEAD)
BASE_SHA=$(sed -n 's/^baseline=//p' "$BRAG_STATE_DIR"/*"$SID1"* 2>/dev/null | head -1 || true)

# ===== H8 — the nudge tells the agent to record an evidence link =====
#
# SPEC-078 LD1: the instruction has to reach the moment of capture. Before this,
# `commit:` appeared only in BRAG.md — never in the tool description, /brag, or
# here — and adoption was zero across the whole corpus.

if printf '%s' "$ctx" | grep -q 'pr:'; then
    ok "H8"
else
    fail "H8" "nudge should tell the agent to record an evidence link (pr:/commit:); got: $ctx"
fi

# ===== H9 — the nudge proposes NO commit hash (SPEC-078 LD3) =====
#
# The hook knows BASELINE and HEAD, so it knows every commit from the session —
# and must not suggest them. Under squash-merge those are branch commits about
# to be rewritten and orphaned, so the moment the hook knows most is exactly
# when its hashes are least likely to survive. Emitting one would manufacture
# unverifiable evidence automatically, at scale.
#
# TWO ASSERTIONS, and the order matters.
#
# H9a is the PRECISE property: the two hashes the hook actually holds must not
# appear. H9b is a generic backstop for any other hash-shaped token.
#
# The first version of this test had only the generic sweep, and it was WRONG
# in a way that passed: it asserted "no hex-ish token", which is not the LD3
# property, and the hook violates it on every real invocation — the nudge
# interpolates the session id, and a real Claude Code session id is a UUID whose
# segments are hex. It stayed green only because the fixture used the toy id
# "session-1". Verified against a real UUID, the old assertion found
# "d838146a" and "892972ca5c5a" and would have failed a correctly-behaving
# hook. So: the fixture is now UUID-shaped so the case cannot hide again, and
# the session id is scrubbed before the backstop runs.

if printf '%s' "$ctx" | grep -qF "$HEAD_SHA"; then
    fail "H9a" "nudge leaked HEAD ($HEAD_SHA) — LD3 forbids proposing a hash the hook holds"
elif [ -n "$BASE_SHA" ] && printf '%s' "$ctx" | grep -qF "$BASE_SHA"; then
    fail "H9a" "nudge leaked BASELINE ($BASE_SHA) — LD3 forbids proposing a hash the hook holds"
else
    ok "H9a"
fi

# Backstop: any OTHER hash-shaped token, with the session id removed first
# (it is legitimately present and legitimately hex-segmented).
scrubbed=$(printf '%s' "$ctx" | sed "s/$SID1//g")
hashes=$(printf '%s' "$scrubbed" | grep -oE '\b[0-9a-f]{7,40}\b' || true)
if [ -z "$hashes" ]; then
    ok "H9b"
else
    fail "H9b" "nudge contains an unexpected hash-shaped token (LD3): $hashes"
fi

# ===== H4 — another Stop after the nudge: silent (once per session) =====

out=$(run_hook "$SID1" "$REPO")
if [ -z "$out" ]; then
    ok "H4"
else
    fail "H4" "expected empty stdout after nudge already fired, got: $out"
fi

# ===== H5 — BRAG_CAPTURE_NUDGE=off with a fresh commit: silent =====

SID2="session-2"
echo "three" >> "$REPO/file.txt"
git -C "$REPO" add -A
git -C "$REPO" commit -q -m "change 2"

out=$(BRAG_CAPTURE_NUDGE=off run_hook "$SID2" "$REPO")
if [ -z "$out" ]; then
    ok "H5"
else
    fail "H5" "expected empty stdout when BRAG_CAPTURE_NUDGE=off, got: $out"
fi

# ===== H6 — cwd is not a git repo: silent =====

SID3="session-3"
out=$(run_hook "$SID3" "$NOTGIT")
if [ -z "$out" ]; then
    ok "H6"
else
    fail "H6" "expected empty stdout in non-git cwd, got: $out"
fi

# ===== H7 — across all of the above, the hook never invoked `brag` =====

if [ ! -e "$SENTINEL" ]; then
    ok "H7"
else
    fail "H7" "brag stub was invoked (sentinel contents: $(cat "$SENTINEL"))"
fi

# ===== finalise =====

if [ "$FAIL_COUNT" -gt 0 ]; then
    printf '\nFAILED: %d assertion(s) failed.\n' "$FAIL_COUNT" >&2
    exit 1
fi

printf '\nALL OK: capture-nudge hook-behavior assertions passed.\n'
exit 0

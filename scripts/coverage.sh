#!/usr/bin/env bash
# scripts/coverage.sh — measure Go statement coverage and enforce the floor.
#
# SINGLE SOURCE for three consumers:
#   1. `just coverage` (local),
#   2. the `coverage` job in .github/workflows/ci.yml, and
#   3. test-docs assertion Z5, which reads FLOOR out of this file and asserts
#      docs/engineering-practices.md states the same number.
# The floor is written down once, here. Nothing else may restate it.
#
# THE UNIT — because "coverage" is not one number. This script means exactly
# one thing by it: the `total:` line of `go tool cover -func` over a profile
# from `go test ./... -covermode=set`, with NO -coverpkg. Each package is
# credited only for what its OWN tests execute. Adding `-coverpkg=./...`
# reports 86.2% on the same tree (measured 2026-08-21) because it credits a
# package for statements some other package's tests happened to run — a bigger
# number for no additional test, so it is not the one enforced.
#
# IT IS A FLOOR, NOT A TARGET. It sits BELOW the measured value on purpose,
# with headroom, so that ordinary refactoring never fails CI and nobody is ever
# rewarded for writing a test that exists to move a percentage. It catches one
# thing: a large untested surface landing at once. See SPEC-082 LD5.

set -eu

FLOOR=80.0

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
cd "$REPO_ROOT"

PROFILE=$(mktemp)
trap 'rm -f "$PROFILE"' EXIT

go test ./... -covermode=set -coverprofile="$PROFILE"

total=$(go tool cover -func="$PROFILE" | awk '/^total:/ {print $NF}')
pct=${total%\%}

printf '\ntotal: %s   floor: %s%%\n' "$total" "$FLOOR"

if awk -v p="$pct" -v f="$FLOOR" 'BEGIN { exit !(p < f) }'; then
    printf 'coverage %s is below the floor of %s%% — see docs/engineering-practices.md\n' \
        "$total" "$FLOOR" >&2
    exit 1
fi

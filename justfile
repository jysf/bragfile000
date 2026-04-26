# Spec-driven repo template — command runner
#
# This justfile works BOTH before and after `just init`:
# - Before init: only `init` and `list-variants` are expected to work.
# - After init: all the daily commands work (status, new-spec, etc.)
#
# Run `just --list` to see everything.

# Show all commands
default:
    @just --list

# ----------------------------------------------------------------------------
# ONE-TIME SETUP
# ----------------------------------------------------------------------------

# Initialize the repo: pick a variant and scaffold files to the root.
init:
    @echo "Spec-driven repo template — init"
    @echo ""
    @if [ -f AGENTS.md ]; then \
        echo "⚠  Already initialized (AGENTS.md exists at repo root)."; \
        echo "   If you want to re-init, remove AGENTS.md first."; \
        exit 1; \
    fi
    @echo "Pick a variant:"
    @echo "  1) claude-only         (Claude plays every role; simpler)"
    @echo "  2) claude-plus-agents  (Claude architects, separate agent implements)"
    @echo ""
    @printf "Enter 1 or 2: "
    @read variant_choice; \
    if [ "$variant_choice" = "1" ]; then \
        VARIANT="claude-only"; \
    elif [ "$variant_choice" = "2" ]; then \
        VARIANT="claude-plus-agents"; \
    else \
        echo "Invalid choice: $variant_choice"; exit 1; \
    fi; \
    echo ""; \
    echo "Scaffolding $VARIANT to repo root..."; \
    cp -r variants/$VARIANT/. .; \
    rm -rf variants/; \
    echo "$VARIANT" > .variant; \
    echo ""; \
    echo "✓ Done. Your variant: $VARIANT"; \
    echo ""; \
    echo "Next steps:"; \
    echo "  1. Open GETTING_STARTED.md"; \
    echo "  2. Work through the PROJECT FRAME prompt in FIRST_SESSION_PROMPTS.md"; \
    echo "  3. Commit the scaffolded repo:"; \
    echo "       git add . && git commit -m 'chore: initialize spec-driven scaffold'"

# List the available variants (useful before init)
list-variants:
    @echo "Available variants:"
    @echo "  claude-only         — Claude plays every role; no handoff documents"
    @echo "  claude-plus-agents  — Claude architects, separate agent implements; adds /handoffs/"
    @echo ""
    @echo "Run 'just init' to pick one."

# ----------------------------------------------------------------------------
# DAILY COMMANDS (work after `just init`)
# ----------------------------------------------------------------------------

# Print repo state: active project, stage, specs by cycle, stale items
status:
    @./scripts/status.sh

# Scaffold a new spec. Usage: just new-spec "short title" STAGE-NNN [PROJ-NNN]
new-spec TITLE STAGE_ID PROJECT_ID="":
    @./scripts/new-spec.sh "{{TITLE}}" "{{STAGE_ID}}" "{{PROJECT_ID}}"

# Scaffold a new stage. Usage: just new-stage "short title" [PROJ-NNN]
new-stage TITLE PROJECT_ID="":
    @./scripts/new-stage.sh "{{TITLE}}" "{{PROJECT_ID}}"

# Advance a spec's cycle. Usage: just advance-cycle SPEC-NNN verify
advance-cycle SPEC_ID NEW_CYCLE:
    @./scripts/advance-cycle.sh "{{SPEC_ID}}" "{{NEW_CYCLE}}"

# Archive a shipped spec: move to done/ and update stage backlog.
# Usage: just archive-spec SPEC-NNN
archive-spec SPEC_ID:
    @./scripts/archive-spec.sh "{{SPEC_ID}}"

# Print the Weekly Review prompt with recent activity pre-loaded
weekly-review:
    @./scripts/weekly-review.sh

# Print all specs grouped by stage with ship dates and complexity sizes
specs-by-stage:
    @./scripts/specs-by-stage.sh

# Snapshot `just status` to docs/reports/daily/YYYY-MM-DD.md (overwrites if run twice in one day)
daily-status-report:
    @mkdir -p docs/reports/daily
    @D="$(date +%Y-%m-%d)"; \
        { echo "# Daily status - $D"; echo; ./scripts/status.sh; } > "docs/reports/daily/$D.md"; \
        echo "✓ Wrote docs/reports/daily/$D.md"

# ----------------------------------------------------------------------------
# APP COMMANDS — bragfile itself
# ----------------------------------------------------------------------------

# Build ./brag in the repo root (no install; for ad-hoc testing)
build:
    @go build -o brag ./cmd/brag
    @echo "✓ Built ./brag (current dir)"

# Install brag to Go's GOBIN (~/go/bin by default). Ensure that directory is on your $PATH.
install:
    @go install ./cmd/brag
    @echo "✓ Installed brag. Verify with: brag --version"

# Remove the installed binary from GOBIN / ~/go/bin.
uninstall:
    @rm -f "$(go env GOBIN 2>/dev/null || echo "$(go env GOPATH)/bin")/brag"
    @echo "✓ Uninstalled brag (if present)"

# Run the full Go test suite.
test:
    @go test ./...

# Run brag without installing. Usage: just run list   |   just run add --title "x"
run *ARGS:
    @go run ./cmd/brag {{ARGS}}

# ----------------------------------------------------------------------------
# HELPERS
# ----------------------------------------------------------------------------

# Print the active project and variant
info:
    @./scripts/info.sh

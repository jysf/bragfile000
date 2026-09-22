package cli

import (
	"strings"
	"testing"
)

// Test H1 — TestStoryCmd_HelpListsAllBuiltInAudiences (the LD4 tested
// contract). Asserts `brag story --help` (via NewStoryCmd().Long + the
// --audience flag usage) enumerates all four built-in audiences, so the two
// new built-in defaults (manager, skip) are discoverable at the CLI. If a
// future audience ships as a built-in without a help refresh, this fails.
func TestStoryCmd_HelpListsAllBuiltInAudiences(t *testing.T) {
	cmd := NewStoryCmd()

	// The Long audience mini-table lists all four built-ins.
	for _, name := range []string{"me", "manager", "skip", "exec"} {
		if !strings.Contains(cmd.Long, name) {
			t.Errorf("story --help Long must list built-in audience %q:\n%s", name, cmd.Long)
		}
	}
	// The two NEW built-ins specifically are present (the regression this
	// test guards): manager + skip appear in the shaping mini-table.
	for _, want := range []string{
		"manager  tactical",   // the manager row's leading label + descriptor
		"skip     skip-level", // the skip row's leading label + descriptor
	} {
		if !strings.Contains(cmd.Long, want) {
			t.Errorf("story --help Long missing new-audience row %q:\n%s", want, cmd.Long)
		}
	}

	// The --audience flag usage names all four built-ins.
	usage := cmd.Flags().Lookup("audience").Usage
	for _, name := range []string{"me", "manager", "skip", "exec"} {
		if !strings.Contains(usage, name) {
			t.Errorf("--audience usage must name built-in %q, got %q", name, usage)
		}
	}
	// The extensibility affordance is preserved (a user's own profile).
	if !strings.Contains(usage, "user profile") {
		t.Errorf("--audience usage must keep the user-profile affordance, got %q", usage)
	}
}

// TestStoryCmd_HelpStatesTheFailureRule ▲ SPEC-094 LD8. `brag story --help`
// says what happens to work recorded with brag learn, per candor, in one
// literal sentence set — the only place a user reading the CLI learns that a
// promotional audience leaves failures out.
func TestStoryCmd_HelpStatesTheFailureRule(t *testing.T) {
	want := `Work recorded with brag learn is never marked as a win. me and manager list it where it falls, as "✗ <id> (failed)". skip and exec leave it out, print an Omitted: line counting it, and end the framing directive with a line telling the LLM to say so. A user profile leaves failures out only when its candor is exactly promotional; any other value lists them.`
	if !strings.Contains(NewStoryCmd().Long, want) {
		t.Errorf("story --help Long is missing the failure rule:\n%s", NewStoryCmd().Long)
	}
}

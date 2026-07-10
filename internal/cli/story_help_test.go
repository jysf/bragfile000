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

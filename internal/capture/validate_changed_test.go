package capture

import (
	"strings"
	"testing"
)

// TestValidateChanged_UnchangedOverCapFieldPasses ▲ LD4 grandfathering: a
// field whose value is identical between old and new is never validated,
// even when it is over today's cap and would fail Validate directly.
func TestValidateChanged_UnchangedOverCapFieldPasses(t *testing.T) {
	old := Fields{Title: "orig", Impact: strings.Repeat("i", 1290)}
	newf := Fields{Title: "new legal title", Impact: old.Impact}
	if err := ValidateChanged(old, newf); err != nil {
		t.Errorf("unchanged over-cap impact should not be validated, got %v", err)
	}
}

// TestValidateChanged_ChangedOverCapFieldIsRejected ▲ LD4: touching an
// over-cap field — even replacing it with a different over-cap value —
// is validated like any ingress path.
func TestValidateChanged_ChangedOverCapFieldIsRejected(t *testing.T) {
	old := Fields{Title: "orig", Impact: strings.Repeat("i", 1290)}
	newf := Fields{Title: "orig", Impact: strings.Repeat("j", 1290)}
	if err := ValidateChanged(old, newf); err == nil {
		t.Fatalf("changed over-cap impact should be rejected")
	}
}

// TestValidateChanged_ChangedFieldMustMeetCapNotMerelyImprove ▲ LD4: the
// rule is "meet the cap", not "not worse than before". A title shortened
// from 300 to 257 bytes is still over MaxTitle and must still be rejected.
func TestValidateChanged_ChangedFieldMustMeetCapNotMerelyImprove(t *testing.T) {
	old := Fields{Title: strings.Repeat("t", 300)}
	newf := Fields{Title: strings.Repeat("t", MaxTitle+1)}
	if err := ValidateChanged(old, newf); err == nil {
		t.Fatalf("a changed title still over cap should be rejected, even though shorter than before")
	}
}

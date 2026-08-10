package capture

import (
	"fmt"
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

// TestValidateChanged_StampedTagsDoNotCountAgainstMaxTagCount ▲ DEC-046
// punch-list item 3. On the edit path, old.Tags/new.Tags are the STORED,
// post-stamp string (ValidateChanged sees what capture.Validate on the add
// path never does: MaxTagCount is meant to apply pre-stamp — DEC-046's own
// "neither cap bounds stamped provenance" scope limit). Without stripping,
// 30 user tags + 5 stamped reserved tags = 35 stored tags, and editing a
// single user tag makes new.Tags != old.Tags, re-counting all 35 against a
// cap of 32 — even though the user only ever supplied 30. This must pass.
func TestValidateChanged_StampedTagsDoNotCountAgainstMaxTagCount(t *testing.T) {
	userTags := make([]string, MaxTagCount-2) // 30 user tags
	for i := range userTags {
		userTags[i] = fmt.Sprintf("t%d", i)
	}
	reserved := "agent:claude-code,model:claude-opus-5,session:sess-abc,cost:0.42,tokens:18000"
	old := Fields{Title: "ok", Tags: strings.Join(userTags, ",") + "," + reserved}

	edited := append([]string(nil), userTags...)
	edited[len(edited)-1] = edited[len(edited)-1] + "b" // change exactly one user tag
	newf := Fields{Title: "ok", Tags: strings.Join(edited, ",") + "," + reserved}

	if old.Tags == newf.Tags {
		t.Fatalf("test fixture bug: old and new Tags must differ to exercise the Tags branch")
	}
	if err := ValidateChanged(old, newf); err != nil {
		t.Errorf("30 user tags + 5 stamped should not trip MaxTagCount=%d, got %v", MaxTagCount, err)
	}
}

// TestValidateChanged_StampedTagsDoNotCountAgainstMaxTagLen ▲ DEC-046
// punch-list item 3. A long stamped model: id (arbitrary length — DEC-046
// line 190-192: no length check on it anywhere) must not make an unrelated
// tag edit fail MaxTagLen, and must not require deleting the provenance tag
// itself to converge.
func TestValidateChanged_StampedTagsDoNotCountAgainstMaxTagLen(t *testing.T) {
	longModel := "model:" + strings.Repeat("m", MaxTagLen+10)
	old := Fields{Title: "ok", Tags: "work," + longModel}
	newf := Fields{Title: "ok", Tags: "work,extra," + longModel}

	if err := ValidateChanged(old, newf); err != nil {
		t.Errorf("editing a user tag should not re-trip MaxTagLen on a stamped model: tag, got %v", err)
	}
}

// TestValidateChanged_NewlyRegisteredReservedPrefixIsStrippedAutomatically
// ▲ DEC-046 punch-list item 2, the closed-loop half. isReservedTag ranges
// over the exported ReservedTagPrefixes rather than its own literal list, so
// registering a sixth reserved prefix there is enough on its own — no
// validate.go edit required — for validateTagsChanged to start stripping it.
// (The other half — that stamping cannot emit an unregistered prefix without
// a loud panic — is pinned in mcpserver's
// TestStampProvenance_UnhandledReservedPrefixIsCaught, which is where the
// mutation lives on the stamping side.)
// The probe is sized on the LENGTH axis, like the MaxTagLen sibling above,
// rather than the count axis. That is what gives this test teeth: an earlier
// draft used MaxTagCount-1 user tags plus two stamped tags, so an unstripped
// "signature:" landed on exactly MaxTagCount and still passed — the test went
// green with isReservedTag's coupling to ReservedTagPrefixes severed, pinning
// nothing. An over-cap signature value cannot be absorbed by an off-by-one:
// if the new prefix is not stripped, MaxTagLen trips on it.
func TestValidateChanged_NewlyRegisteredReservedPrefixIsStrippedAutomatically(t *testing.T) {
	orig := ReservedTagPrefixes
	ReservedTagPrefixes = append(append([]string{}, orig...), "signature:")
	t.Cleanup(func() { ReservedTagPrefixes = orig })

	stamped := "agent:claude-code,signature:" + strings.Repeat("d", MaxTagLen+10)
	old := Fields{Title: "ok", Tags: "work," + stamped}
	newf := Fields{Title: "ok", Tags: "work,extra," + stamped}

	if err := ValidateChanged(old, newf); err != nil {
		t.Errorf("a newly-registered reserved prefix should be stripped with no validate.go edit, got %v", err)
	}
}

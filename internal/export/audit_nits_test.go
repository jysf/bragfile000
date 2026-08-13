package export

import (
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// TestToMarkdown_ChronoTieBreaksOnID is the STAGE-018 audit-nit regression.
//
// The chronological section sorted on CreatedAt alone with sort.SliceStable,
// so entries sharing a timestamp kept whatever order the DB handed back —
// making the rendered document depend on query plan rather than on data.
// created_at is second-resolution, so a tie is not exotic: anything that
// captures two entries in one second (a script, an import, the MCP server
// answering twice) produces one.
//
// Feeding the SAME entries in two different input orders must produce
// byte-identical output. Without the ID tie-break this fails, because
// SliceStable faithfully preserves the two different input orders.
func TestToMarkdown_ChronoTieBreaksOnID(t *testing.T) {
	ts := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	mk := func(id int64, title string) storage.Entry {
		return storage.Entry{ID: id, Title: title, CreatedAt: ts, UpdatedAt: ts}
	}
	ascending := []storage.Entry{mk(1, "first"), mk(2, "second"), mk(3, "third")}
	shuffled := []storage.Entry{mk(3, "third"), mk(1, "first"), mk(2, "second")}

	render := func(entries []storage.Entry) string {
		b, err := ToMarkdown(entries, MarkdownOptions{Now: fixedNow})
		if err != nil {
			t.Fatalf("ToMarkdown: %v", err)
		}
		return string(b)
	}

	got, want := render(shuffled), render(ascending)
	if got != want {
		t.Errorf("same entries in a different input order rendered differently —\n"+
			"the chronological sort is not total, so output depends on DB order.\n"+
			"--- from shuffled ---\n%s\n--- from ascending ---\n%s", got, want)
	}

	// And the tie must resolve by ascending ID, not merely deterministically.
	iFirst := strings.Index(want, "first")
	iSecond := strings.Index(want, "second")
	iThird := strings.Index(want, "third")
	if !(iFirst < iSecond && iSecond < iThird) {
		t.Errorf("same-timestamp entries must order by ascending ID (insertion order); "+
			"got positions first=%d second=%d third=%d", iFirst, iSecond, iThird)
	}
}

// TestToMarkdown_UntypedEntryRendersSentinelInByType is the STAGE-018
// audit-nit regression for empty-`type` handling.
//
// writeSummaryByType counted on e.Type directly, so an entry with no type
// produced a bare, blank-labelled row — literally "- : 3" — in the By type
// summary. The three renderers disagreed on empty type: wrapped skipped it,
// memory substituted "-", and markdown rendered a blank. This pins markdown to
// the "-" sentinel the others already use.
func TestToMarkdown_UntypedEntryRendersSentinelInByType(t *testing.T) {
	ts := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	entries := []storage.Entry{
		{ID: 1, Title: "typed", Type: "shipped", CreatedAt: ts, UpdatedAt: ts},
		{ID: 2, Title: "untyped", CreatedAt: ts, UpdatedAt: ts},
	}
	b, err := ToMarkdown(entries, MarkdownOptions{Now: fixedNow})
	if err != nil {
		t.Fatalf("ToMarkdown: %v", err)
	}
	out := string(b)

	if strings.Contains(out, "- : ") {
		t.Errorf("By type rendered a blank label for an untyped entry (bare \"- : \"):\n%s", out)
	}
	if !strings.Contains(out, "- -: 1") {
		t.Errorf("untyped entry should count under the \"-\" sentinel, matching "+
			"`brag memory` and `brag list`; got:\n%s", out)
	}
}

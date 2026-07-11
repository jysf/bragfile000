package mcpserver

import (
	"context"
	"testing"

	"github.com/jysf/bragfile000/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SPEC-064 — capture input hardening at the MCP ingress. brag_add already
// enforced byte caps + numeric provenance validation; these tests pin the
// added control-char rejection (single-line fields reject C0 incl NUL;
// description rejects NUL only) and reserved-numeric-tag validation in the
// freeform tags field.

// (B) Control chars in the single-line title must be a tool error on the MCP
// path: newline, tab, and NUL.
func TestServer_AddRejectsControlCharsInTitle(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	for _, title := range []string{"a\nb", "a\tb", "a\x00b"} {
		r, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "brag_add",
			Arguments: map[string]any{"title": title},
		})
		if err != nil {
			t.Fatal(err)
		}
		if !r.IsError {
			t.Errorf("brag_add title %q should be a tool error", title)
		}
	}
	if rows, _ := s.List(storage.ListFilter{}); len(rows) != 0 {
		t.Errorf("no row should be inserted on control-char title, got %d", len(rows))
	}
}

// (B) description is multi-line: a newline is accepted, a NUL is rejected.
func TestServer_AddDescriptionNewlineOkNulRejected(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")

	callJSON(t, cs, "brag_add", map[string]any{
		"title":       "ok",
		"description": "line1\nline2",
	})
	if rows, _ := s.List(storage.ListFilter{}); len(rows) != 1 {
		t.Fatalf("expected 1 entry after newline description, got %d", len(rows))
	}

	r, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "brag_add",
		Arguments: map[string]any{"title": "ok2", "description": "bad\x00body"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !r.IsError {
		t.Error("brag_add with NUL in description should be a tool error")
	}
	if rows, _ := s.List(storage.ListFilter{}); len(rows) != 1 {
		t.Errorf("NUL description must not insert; want 1 total row, got %d", len(rows))
	}
}

// (C) A reserved cost:/tokens: token smuggled through the freeform tags
// field is validated with the dedicated-param rules and rejected when bad,
// so garbage cannot bypass the validated cost/tokens params.
func TestServer_AddRejectsBadReservedTagInTags(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	for _, tags := range []string{"cost:-9", "tokens:xyz", "perf,cost:$5"} {
		r, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "brag_add",
			Arguments: map[string]any{"title": "x", "tags": tags},
		})
		if err != nil {
			t.Fatal(err)
		}
		if !r.IsError {
			t.Errorf("brag_add tags %q should be a tool error", tags)
		}
	}
	if rows, _ := s.List(storage.ListFilter{}); len(rows) != 0 {
		t.Errorf("no row should be inserted on bad reserved tag, got %d", len(rows))
	}
}

// (C) A valid cost: token in freeform tags is accepted (and coexists with a
// non-numeric provenance token) — reserved tokens are validated, not banned.
func TestServer_AddAcceptsValidReservedTagInTags(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	callJSON(t, cs, "brag_add", map[string]any{
		"title": "ok", "tags": "cost:12.50,agent:manual",
	})
	rows, _ := s.List(storage.ListFilter{})
	if len(rows) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(rows))
	}
}

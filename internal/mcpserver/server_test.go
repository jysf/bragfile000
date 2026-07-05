package mcpserver

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// newTestServer opens a *storage.Store on t.TempDir(), builds the server via
// New(s), wires mcp.NewInMemoryTransports(), and connects a client
// identifying as clientName. Returns the connected *mcp.ClientSession + the
// *storage.Store (for out-of-band seeding/assertions).
func newTestServer(t *testing.T, clientName string) (*mcp.ClientSession, *storage.Store) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "db.sqlite")
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	srv := New(s)
	ctx := context.Background()
	ct, stt := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, stt, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: clientName, Version: "1.0"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs, s
}

// callJSON calls the named tool and returns the first TextContent's text.
// Fails the test on a transport error or a tool-level IsError result.
func callJSON(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) string {
	t.Helper()
	r, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	if r.IsError {
		t.Fatalf("call %s: tool error: %+v", name, r.Content)
	}
	tc, ok := r.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("call %s: content[0] not TextContent: %T", name, r.Content[0])
	}
	return tc.Text
}

// seedViaStore inserts one entry per title directly through the storage
// layer (bypassing the MCP tools, so no provenance is stamped).
func seedViaStore(t *testing.T, s *storage.Store, titles ...string) {
	t.Helper()
	for _, title := range titles {
		if _, err := s.Add(storage.Entry{Title: title}); err != nil {
			t.Fatalf("seed %q: %v", title, err)
		}
	}
}

// setNowFunc overrides the package's nowFunc seam for one test and returns a
// restore function.
func setNowFunc(t *testing.T, fn func() time.Time) func() {
	t.Helper()
	old := nowFunc
	nowFunc = fn
	return func() { nowFunc = old }
}

// TestServer_ToolsListed ▲ exactly the four tool names, nothing else.
func TestServer_ToolsListed(t *testing.T) {
	cs, _ := newTestServer(t, "claude-code")
	lt, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, x := range lt.Tools {
		names = append(names, x.Name)
	}
	sort.Strings(names)
	want := []string{"brag_add", "brag_list", "brag_search", "brag_stats"}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("tools = %v, want %v", names, want)
	}
}

// TestServer_AddRequiresTitle ▲ schema validation: no title → IsError.
func TestServer_AddRequiresTitle(t *testing.T) {
	cs, _ := newTestServer(t, "claude-code")
	r, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "brag_add", Arguments: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if !r.IsError {
		t.Error("brag_add with no title should be a tool error")
	}
}

// TestServer_AddStampsProvenanceAndListParity ▲ the headline: brag_add with
// explicit agent+model stamps the reserved tags; brag_list --tag model:<id>
// finds it; and the list payload is byte-identical to export.ToJSON of the row.
func TestServer_AddStampsProvenanceAndListParity(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	callJSON(t, cs, "brag_add", map[string]any{
		"title": "cut p99", "tags": "perf",
		"agent": "claude-code", "model": "claude-opus-4-8",
	})
	// stored via Store → provenance rode the DEC-015 tags path
	rows, _ := s.List(storage.ListFilter{Tag: "model:claude-opus-4-8"})
	if len(rows) != 1 || rows[0].Title != "cut p99" {
		t.Fatalf("provenance tag not filterable: %+v", rows)
	}
	if rows[0].Tags != "perf,agent:claude-code,model:claude-opus-4-8" {
		t.Errorf("stored tags = %q", rows[0].Tags)
	}
	got := callJSON(t, cs, "brag_list", map[string]any{"tag": "model:claude-opus-4-8"})
	want, _ := export.ToJSON(rows)
	if got != string(want) {
		t.Errorf("brag_list not byte-parity with export.ToJSON:\n got=%s\nwant=%s", got, want)
	}
}

// TestServer_AddAutoStampsAgentFromClientInfo ▲ omit the agent param → agent:
// auto-fills from clientInfo.Name; model omitted → no model: tag.
func TestServer_AddAutoStampsAgentFromClientInfo(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	callJSON(t, cs, "brag_add", map[string]any{"title": "shipped"})
	rows, _ := s.List(storage.ListFilter{})
	if rows[0].Tags != "agent:claude-code" {
		t.Errorf("auto-stamp: tags = %q, want %q", rows[0].Tags, "agent:claude-code")
	}
}

// TestServer_SearchParity ▲ DEC-010 tokenization; brag_search parity with
// Store.Search on the same query.
func TestServer_SearchParity(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "cut p99 latency", "shipped auth refactor")
	got := callJSON(t, cs, "brag_search", map[string]any{"query": "latency"})
	m, _ := buildMatch("latency")
	rows, _ := s.Search(m, 0)
	want, _ := export.ToJSON(rows)
	if got != string(want) {
		t.Errorf("search parity:\n got=%s\nwant=%s", got, want)
	}
}

// TestServer_StatsParityWithCLI ▲ brag_stats byte-identical to the DEC-014
// envelope for the same corpus + pinned Now (nowFunc seam).
func TestServer_StatsParityWithCLI(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one", "two")
	fixed := time.Date(2026, 7, 4, 12, 0, 0, 0, time.Local)
	restore := setNowFunc(t, func() time.Time { return fixed })
	defer restore()
	got := callJSON(t, cs, "brag_stats", map[string]any{})
	rows, _ := s.List(storage.ListFilter{})
	want, _ := export.ToStatsJSON(rows, export.StatsOptions{Now: fixed})
	if got != string(want) {
		t.Errorf("stats parity:\n got=%s\nwant=%s", got, want)
	}
}

// TestServer_AddReturnValueParity ● preservation guard: brag_add's RETURNED
// JSON is byte-identical, field-for-field, to export.ToJSON's rendering of
// the same created row. Closes the SPEC-040 verify advisory (only brag_add's
// side effects, not its literal return value, were asserted). Passes
// immediately — the local entryRecord mirrors export's field-for-field.
func TestServer_AddReturnValueParity(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	got := callJSON(t, cs, "brag_add", map[string]any{
		"title": "cut p99", "project": "bragfile", "type": "shipped",
		"tags": "perf", "impact": "-80% p99",
		"agent": "claude-code", "model": "claude-opus-4-8",
	})
	// The created row is the sole row in the store.
	rows, err := s.List(storage.ListFilter{})
	if err != nil || len(rows) != 1 {
		t.Fatalf("expected exactly one created row, got %d (err=%v)", len(rows), err)
	}
	// Pinned export call: the per-entry rendering brag_add's return mirrors.
	arr, err := export.ToJSON([]storage.Entry{rows[0]})
	if err != nil {
		t.Fatalf("export.ToJSON: %v", err)
	}
	// Compare field-for-field on the RAW value bytes, so array-wrapper and
	// indentation differences (object vs 1-element array) are irrelevant and
	// only the per-field serialization is asserted.
	var wantArr []map[string]json.RawMessage
	if err := json.Unmarshal(arr, &wantArr); err != nil || len(wantArr) != 1 {
		t.Fatalf("unmarshal export array: %v (len=%d)", err, len(wantArr))
	}
	var gotObj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(got), &gotObj); err != nil {
		t.Fatalf("unmarshal brag_add return: %v", err)
	}
	want := wantArr[0]
	if len(gotObj) != len(want) {
		t.Fatalf("field count: brag_add has %d keys, export has %d", len(gotObj), len(want))
	}
	for k, wv := range want {
		gv, ok := gotObj[k]
		if !ok {
			t.Errorf("brag_add return missing key %q present in export", k)
			continue
		}
		if string(gv) != string(wv) {
			t.Errorf("field %q not byte-identical: brag_add=%s export=%s", k, gv, wv)
		}
	}
}

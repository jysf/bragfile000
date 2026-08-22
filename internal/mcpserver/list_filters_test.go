package mcpserver

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/jysf/bragfile000/internal/storage/storagetest"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SPEC-072 / DEC-042 — time-window and provenance filter parity at the MCP
// boundary. brag_list now accepts since/until/day/author with the CLI's exact
// grammar (DEC-008 for since/until, DEC-039 for day) via the shared pure
// internal/timewindow package, plus the folded-in v0.5.0 audit item: a
// negative limit is a tool error, not a silent "unlimited".

// pdtZone is a fixed UTC-7 zone standing in for a non-UTC local zone, so the
// local-vs-UTC day assertions do not depend on the host's TZ.
var pdtZone = time.FixedZone("PDT", -7*3600)

// newFilterTestServer is newTestServer plus the on-disk DB path, which the
// filter tests need for storagetest.Backdate (Store.Add always stamps now).
func newFilterTestServer(t *testing.T) (*mcp.ClientSession, *storage.Store, string) {
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
	client := mcp.NewClient(&mcp.Implementation{Name: "claude-code", Version: "1.0"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs, s, dbPath
}

// seedAt inserts one entry and rewrites its created_at to at, returning the id.
func seedAt(t *testing.T, s *storage.Store, dbPath, title string, at time.Time) int64 {
	t.Helper()
	e, err := s.Add(storage.Entry{Title: title})
	if err != nil {
		t.Fatalf("seed %q: %v", title, err)
	}
	if err := storagetest.Backdate(dbPath, e.ID, at); err != nil {
		t.Fatalf("backdate %q: %v", title, err)
	}
	return e.ID
}

// callToolErrText calls a tool expecting a tool-level error and returns the
// concatenated text of the error result. Fails if the call SUCCEEDS.
func callToolErrText(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) string {
	t.Helper()
	r, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: transport error: %v", name, err)
	}
	if !r.IsError {
		t.Fatalf("call %s with %v: expected a tool error, got success", name, args)
	}
	var sb strings.Builder
	for _, c := range r.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			sb.WriteString(tc.Text)
		}
	}
	return sb.String()
}

// titlesOf extracts the "title" values from a brag_list JSON payload, in order.
func titlesOf(t *testing.T, payload string) []string {
	t.Helper()
	var rows []struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal([]byte(payload), &rows); err != nil {
		t.Fatalf("decode list payload: %v\npayload: %s", err, payload)
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Title)
	}
	return out
}

func assertTitles(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("titles = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("titles = %v, want %v", got, want)
		}
	}
}

// --- since ------------------------------------------------------------------

func TestBragList_Since_RelativeDuration(t *testing.T) {
	cs, s, dbPath := newFilterTestServer(t)
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	defer setNowFunc(t, func() time.Time { return now })()

	seedAt(t, s, dbPath, "one-day-ago", now.Add(-24*time.Hour))
	seedAt(t, s, dbPath, "ten-days-ago", now.Add(-10*24*time.Hour))
	seedAt(t, s, dbPath, "forty-days-ago", now.Add(-40*24*time.Hour))

	got := titlesOf(t, callJSON(t, cs, "brag_list", map[string]any{"since": "7d"}))
	assertTitles(t, got, []string{"one-day-ago"})
}

// A bare date is UTC midnight (DEC-008), unchanged by the move to MCP.
func TestBragList_Since_BareDateIsUTCMidnight(t *testing.T) {
	cs, s, dbPath := newFilterTestServer(t)

	seedAt(t, s, dbPath, "at-midnight", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	seedAt(t, s, dbPath, "one-second-before", time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC))

	got := titlesOf(t, callJSON(t, cs, "brag_list", map[string]any{"since": "2026-01-01"}))
	assertTitles(t, got, []string{"at-midnight"})
}

func TestBragList_InvalidSinceIsToolError(t *testing.T) {
	cs, _, _ := newFilterTestServer(t)
	msg := callToolErrText(t, cs, "brag_list", map[string]any{"since": "5y"})
	if !strings.Contains(msg, `"5y"`) {
		t.Errorf("error %q should quote the offending value", msg)
	}
	if !strings.Contains(msg, "since") {
		t.Errorf("error %q should name the since field", msg)
	}
}

// --- until ------------------------------------------------------------------

func TestBragList_Until_IsExclusiveUpperBound(t *testing.T) {
	cs, s, dbPath := newFilterTestServer(t)

	seedAt(t, s, dbPath, "one-second-before", time.Date(2026, 7, 5, 23, 59, 59, 0, time.UTC))
	seedAt(t, s, dbPath, "exactly-at-bound", time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC))

	got := titlesOf(t, callJSON(t, cs, "brag_list", map[string]any{"until": "2026-07-06"}))
	assertTitles(t, got, []string{"one-second-before"})
}

func TestBragList_InvalidUntilIsToolError(t *testing.T) {
	cs, _, _ := newFilterTestServer(t)
	msg := callToolErrText(t, cs, "brag_list", map[string]any{"until": "next tuesday"})
	if !strings.Contains(msg, "until") {
		t.Errorf("error %q should name the until field", msg)
	}
}

// --- day --------------------------------------------------------------------

// The DEC-039 local-vs-UTC skew, proven at the MCP boundary: for a UTC-7
// caller, an entry logged at 23:30 local belongs to THAT local day even though
// its UTC timestamp already rolled into the next date.
func TestBragList_Day_ResolvesLocalCalendarDay(t *testing.T) {
	cs, s, dbPath := newFilterTestServer(t)
	defer setNowFunc(t, func() time.Time {
		return time.Date(2026, 7, 5, 12, 0, 0, 0, pdtZone)
	})()

	seedAt(t, s, dbPath, "in-day-2330-local", time.Date(2026, 7, 5, 23, 30, 0, 0, pdtZone))
	seedAt(t, s, dbPath, "next-day-0030-local", time.Date(2026, 7, 6, 0, 30, 0, 0, pdtZone))

	got := titlesOf(t, callJSON(t, cs, "brag_list", map[string]any{"day": "today"}))
	assertTitles(t, got, []string{"in-day-2330-local"})

	got = titlesOf(t, callJSON(t, cs, "brag_list", map[string]any{"day": "2026-07-06"}))
	assertTitles(t, got, []string{"next-day-0030-local"})
}

func TestBragList_Day_ConflictsWithSince(t *testing.T) {
	cs, _, _ := newFilterTestServer(t)
	msg := callToolErrText(t, cs, "brag_list", map[string]any{"day": "today", "since": "7d"})
	if !strings.Contains(msg, "mutually exclusive") {
		t.Errorf("error %q should name the conflict", msg)
	}
}

func TestBragList_Day_ConflictsWithUntil(t *testing.T) {
	cs, _, _ := newFilterTestServer(t)
	msg := callToolErrText(t, cs, "brag_list", map[string]any{"day": "today", "until": "2026-07-06"})
	if !strings.Contains(msg, "mutually exclusive") {
		t.Errorf("error %q should name the conflict", msg)
	}
}

// The conflict is reported BEFORE parsing, so a caller sending both a bad
// since and a day is told about the real problem.
func TestBragList_Day_ConflictReportedBeforeParse(t *testing.T) {
	cs, _, _ := newFilterTestServer(t)
	msg := callToolErrText(t, cs, "brag_list", map[string]any{"day": "today", "since": "garbage"})
	if !strings.Contains(msg, "mutually exclusive") {
		t.Errorf("error %q should report the conflict, not the parse failure", msg)
	}
}

func TestBragList_InvalidDayIsToolError(t *testing.T) {
	cs, _, _ := newFilterTestServer(t)
	msg := callToolErrText(t, cs, "brag_list", map[string]any{"day": "tomorrow"})
	if !strings.Contains(msg, "day") {
		t.Errorf("error %q should name the day field", msg)
	}
}

// --- author -----------------------------------------------------------------

func TestBragList_Author_AgentAndHuman(t *testing.T) {
	cs, s, _ := newFilterTestServer(t)

	// brag_add stamps agent:<clientInfo.Name> (DEC-024); a direct Store.Add
	// carries no provenance tag at all.
	callJSON(t, cs, "brag_add", map[string]any{"title": "agent-authored"})
	if _, err := s.Add(storage.Entry{Title: "human-authored"}); err != nil {
		t.Fatalf("seed human entry: %v", err)
	}

	assertTitles(t, titlesOf(t, callJSON(t, cs, "brag_list", map[string]any{"author": "agent"})),
		[]string{"agent-authored"})
	assertTitles(t, titlesOf(t, callJSON(t, cs, "brag_list", map[string]any{"author": "human"})),
		[]string{"human-authored"})
}

func TestBragList_Author_InvalidValueIsToolError(t *testing.T) {
	cs, _, _ := newFilterTestServer(t)
	msg := callToolErrText(t, cs, "brag_list", map[string]any{"author": "robot"})
	if !strings.Contains(msg, "agent") || !strings.Contains(msg, "human") {
		t.Errorf("error %q should name both accepted values", msg)
	}
}

// --- limit (folded-in v0.5.0 audit item) ------------------------------------

func TestBragList_NegativeLimitIsToolError(t *testing.T) {
	cs, s, _ := newFilterTestServer(t)
	seedViaStore(t, s, "a", "b")
	msg := callToolErrText(t, cs, "brag_list", map[string]any{"limit": -1})
	if !strings.Contains(msg, "negative") {
		t.Errorf("error %q should say the limit must not be negative", msg)
	}
}

func TestBragSearch_NegativeLimitIsToolError(t *testing.T) {
	cs, s, _ := newFilterTestServer(t)
	seedViaStore(t, s, "alpha")
	msg := callToolErrText(t, cs, "brag_search", map[string]any{"query": "alpha", "limit": -1})
	if !strings.Contains(msg, "negative") {
		t.Errorf("error %q should say the limit must not be negative", msg)
	}
}

// limit 0 remains "unlimited" — the zero value must not be swept up by the
// negative guard.
func TestBragList_ZeroLimitStillMeansUnlimited(t *testing.T) {
	cs, s, _ := newFilterTestServer(t)
	seedViaStore(t, s, "a", "b", "c")
	got := titlesOf(t, callJSON(t, cs, "brag_list", map[string]any{"limit": 0}))
	if len(got) != 3 {
		t.Errorf("limit 0 returned %d rows, want 3 (unlimited)", len(got))
	}
}

// --- parity + schema --------------------------------------------------------

// The filters change WHICH rows come back, never how they are rendered: the
// tool payload stays byte-identical to export.ToJSON over the equivalent
// storage.ListFilter (DEC-024's parity-by-construction contract).
func TestBragList_FilteredOutputMatchesStoreParity(t *testing.T) {
	cs, s, dbPath := newFilterTestServer(t)
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	defer setNowFunc(t, func() time.Time { return now })()

	for i, title := range []string{"recent-a", "recent-b", "old-c"} {
		id := seedAt(t, s, dbPath, title, now.Add(-time.Duration(i+1)*24*time.Hour))
		_ = id
	}
	seedAt(t, s, dbPath, "way-old", now.Add(-90*24*time.Hour))

	args := map[string]any{"since": "7d", "until": "2026-07-10", "author": "human", "limit": 2}
	payload := callJSON(t, cs, "brag_list", args)

	sinceT := now.Add(-7 * 24 * time.Hour).UTC()
	untilT := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	rows, err := s.List(storage.ListFilter{Since: sinceT, Until: untilT, Author: "human", Limit: 2})
	if err != nil {
		t.Fatalf("store list: %v", err)
	}
	want, err := export.ToJSON(rows)
	if err != nil {
		t.Fatalf("export.ToJSON: %v", err)
	}
	if payload != string(want) {
		t.Errorf("tool payload differs from export.ToJSON parity\n got: %s\nwant: %s", payload, want)
	}
	if len(rows) == 0 {
		t.Fatal("parity fixture selected zero rows; the test would pass vacuously")
	}
}

func TestBragList_ToolSchemaAdvertisesNewFilters(t *testing.T) {
	cs, _, _ := newFilterTestServer(t)
	lt, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	for _, tool := range lt.Tools {
		if tool.Name == "brag_list" {
			m, ok := tool.InputSchema.(map[string]any)
			if !ok {
				t.Fatalf("brag_list InputSchema is %T, want map[string]any", tool.InputSchema)
			}
			schema = m
		}
	}
	if schema == nil {
		t.Fatal("brag_list not advertised")
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("inputSchema.properties is %T, want map[string]any", schema["properties"])
	}
	for _, want := range []string{"tag", "project", "type", "since", "until", "day", "author", "limit"} {
		if _, ok := props[want]; !ok {
			t.Errorf("brag_list input schema missing property %q", want)
		}
	}
}

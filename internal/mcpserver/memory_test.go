package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/memory"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// parseBudgetField extracts the integer value off a "- <field>: <n>[ tokens]"
// line in a DEC-044 ## Budget section.
func parseBudgetField(t *testing.T, body, field string) int {
	t.Helper()
	prefix := "- " + field + ": "
	for _, ln := range strings.Split(body, "\n") {
		if strings.HasPrefix(ln, prefix) {
			rest := strings.TrimSuffix(strings.TrimPrefix(ln, prefix), " tokens")
			n, err := strconv.Atoi(rest)
			if err != nil {
				t.Fatalf("parse %q from line %q: %v", field, ln, err)
			}
			return n
		}
	}
	t.Fatalf("field %q not found in body:\n%s", field, body)
	return 0
}

func TestMemoryTool_ByteIdenticalToCLIMarkdown(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one", "two", "three")
	fixed := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	restore := setNowFunc(t, func() time.Time { return fixed })
	defer restore()

	got := callJSON(t, cs, "brag_memory", map[string]any{})
	want := wantMemoryMarkdown(t, s, "", "", memory.DefaultBudget, "(none)", map[string]string{}, fixed)
	if got != string(want) {
		t.Errorf("body mismatch:\n got=%s\nwant=%s", got, want)
	}
}

func TestMemoryTool_JSONFormat(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one", "two")

	got := callJSON(t, cs, "brag_memory", map[string]any{"format": "json"})
	var env struct {
		Scope  string `json:"scope"`
		Budget int    `json:"budget"`
		Slice  []any  `json:"slice"`
	}
	if err := json.Unmarshal([]byte(got), &env); err != nil {
		t.Fatalf("unmarshal: %v (body=%s)", err, got)
	}
	if env.Scope != "lifetime" {
		t.Errorf("scope = %q, want lifetime", env.Scope)
	}
	if env.Budget != memory.DefaultBudget {
		t.Errorf("budget = %d, want %d", env.Budget, memory.DefaultBudget)
	}
	if env.Slice == nil {
		t.Error("slice is nil, want a (possibly empty) array")
	}

	md := callJSON(t, cs, "brag_memory", map[string]any{"format": "markdown"})
	if !strings.HasPrefix(md, "# Bragfile Memory") {
		t.Errorf("markdown body = %q", md)
	}
}

func TestMemoryTool_DefaultFormatIsMarkdown(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one")
	got := callJSON(t, cs, "brag_memory", map[string]any{})
	if !strings.HasPrefix(got, "# Bragfile Memory") {
		t.Errorf("expected markdown, got %q", got)
	}
	var v any
	if err := json.Unmarshal([]byte(got), &v); err == nil {
		t.Errorf("default-format body should not be valid JSON, got %q", got)
	}
}

func TestMemoryTool_JSONAndMarkdownSelectTheSameSlice(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one", "two", "three", "four", "five")

	mdBody := callJSON(t, cs, "brag_memory", map[string]any{"format": "markdown"})
	jsonBody := callJSON(t, cs, "brag_memory", map[string]any{"format": "json"})

	var env struct {
		Slice []struct {
			ID int64 `json:"id"`
		} `json:"slice"`
	}
	if err := json.Unmarshal([]byte(jsonBody), &env); err != nil {
		t.Fatalf("unmarshal json body: %v", err)
	}
	var jsonIDs []int64
	for _, it := range env.Slice {
		jsonIDs = append(jsonIDs, it.ID)
	}

	var mdIDs []int64
	inSlice := false
	for _, ln := range strings.Split(mdBody, "\n") {
		if ln == "## Slice" {
			inSlice = true
			continue
		}
		if ln == "## Budget" {
			break
		}
		if inSlice && strings.HasPrefix(ln, "- ") {
			fields := strings.Fields(ln)
			id, err := strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				t.Fatalf("parse id from line %q: %v", ln, err)
			}
			mdIDs = append(mdIDs, id)
		}
	}

	if !reflect.DeepEqual(jsonIDs, mdIDs) {
		t.Errorf("json ids = %v, markdown ids = %v", jsonIDs, mdIDs)
	}
}

func TestMemoryTool_BudgetOmittedUsesDefault(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one")
	got := callJSON(t, cs, "brag_memory", map[string]any{})
	want := fmt.Sprintf("- Budget: %d tokens", memory.DefaultBudget)
	if !strings.Contains(got, want) {
		t.Errorf("body missing %q:\n%s", want, got)
	}
}

func TestMemoryTool_BudgetZeroIsAnError(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one")
	r, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "brag_memory", Arguments: map[string]any{"budget": 0}})
	if err != nil {
		t.Fatal(err)
	}
	if !r.IsError {
		t.Error("budget=0 should be a tool error (distinguishable from omitted via *int)")
	}
}

func TestMemoryTool_BudgetNegativeIsAnError(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one")
	r, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "brag_memory", Arguments: map[string]any{"budget": -1}})
	if err != nil {
		t.Fatal(err)
	}
	if !r.IsError {
		t.Error("budget=-1 should be a tool error")
	}
}

func TestMemoryTool_BudgetHonoured(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	for i := 0; i < 10; i++ {
		if _, err := s.Add(storage.Entry{
			Title:  fmt.Sprintf("entry number %d with a fairly long title to cost real tokens", i),
			Impact: "a moderately long impact clause to add even more estimated tokens",
		}); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	got := callJSON(t, cs, "brag_memory", map[string]any{"budget": 40})
	if !strings.Contains(got, "- Budget: 40 tokens") {
		t.Errorf("body missing budget line:\n%s", got)
	}
	if skipped := parseBudgetField(t, got, "Skipped"); skipped <= 0 {
		t.Errorf("expected Skipped > 0 with a 40-token budget over 10 verbose entries, got %d:\n%s", skipped, got)
	}
}

func TestMemoryTool_ProjectIsASoftBoost(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	if _, err := s.Add(storage.Entry{Title: "orbit work", Project: "orbit"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := s.Add(storage.Entry{Title: "bragfile work", Project: "bragfile"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	got := callJSON(t, cs, "brag_memory", map[string]any{"project": "orbit"})
	if !strings.Contains(got, "[bragfile/") {
		t.Errorf("expected entries from other projects too (soft boost), got:\n%s", got)
	}
}

func TestMemoryTool_UnknownFormatIsAnError(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one")
	r, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "brag_memory", Arguments: map[string]any{"format": "xml"}})
	if err != nil {
		t.Fatal(err)
	}
	if !r.IsError {
		t.Error("format=xml should be a tool error")
	}
}

func TestMemoryTool_MalformedQueryIsAnError(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one")
	r, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "brag_memory", Arguments: map[string]any{"query": `has"quote`}})
	if err != nil {
		t.Fatal(err)
	}
	if !r.IsError {
		t.Error(`query with an unbalanced quote should be a tool error`)
	}
}

// TestMemoryTool_SchemaHasNoTimeWindowParams ▲ reads brag_memory's InputSchema
// off tools/list and asserts its properties keys are EXACTLY
// {budget, format, project, query} — since/until/day absent by construction,
// stronger than three separate NOT-contains checks (LD6).
func TestMemoryTool_SchemaHasNoTimeWindowParams(t *testing.T) {
	cs, _ := newTestServer(t, "claude-code")
	lt, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	for _, tool := range lt.Tools {
		if tool.Name == "brag_memory" {
			schema, _ = tool.InputSchema.(map[string]any)
		}
	}
	if schema == nil {
		t.Fatal("brag_memory not found in tools/list, or its InputSchema was not a map[string]any")
	}
	props, _ := schema["properties"].(map[string]any)
	var keys []string
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	want := []string{"budget", "format", "project", "query"}
	if !reflect.DeepEqual(keys, want) {
		t.Errorf("properties = %v, want %v", keys, want)
	}
}

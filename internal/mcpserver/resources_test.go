package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/memory"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// wantMemoryMarkdown computes the DEC-043/044 memory-slice body independently
// of the handler under test: Options and MemoryOptions are hard-coded here,
// not read back from the server, so a handler that silently changes the
// budget, the pool composition, or the filters produces a diff rather than
// two implementations agreeing with each other. Shared with memory_test.go.
func wantMemoryMarkdown(t *testing.T, s *storage.Store, query, project string, budget int, filters string, filtersJSON map[string]string, now time.Time) []byte {
	t.Helper()
	pool, err := memory.Gather(s, memory.GatherOptions{Query: query, Project: project})
	if err != nil {
		t.Fatalf("memory.Gather: %v", err)
	}
	result := memory.Slice(pool.Entries, memory.Options{
		Query:   query,
		Project: project,
		Matched: pool.Matched,
		Budget:  budget,
	})
	body, err := export.ToMemoryMarkdown(result, export.MemoryOptions{
		Filters:     filters,
		FiltersJSON: filtersJSON,
		Now:         now,
	})
	if err != nil {
		t.Fatalf("export.ToMemoryMarkdown: %v", err)
	}
	return body
}

// assertInvalidParams fails the test unless err unwraps (via errors.As, NOT a
// bare type assertion — the client-side error is a *fmt.wrapError, per
// SPEC-074's §12(b) pre-flight) to a *jsonrpc.Error whose Code is
// jsonrpc.CodeInvalidParams (-32602, per SEP-2164; NOT the deprecated
// mcp.CodeResourceNotFound).
func assertInvalidParams(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var jerr *jsonrpc.Error
	if !errors.As(err, &jerr) {
		t.Fatalf("error does not unwrap to *jsonrpc.Error: %v (%T)", err, err)
	}
	if jerr.Code != jsonrpc.CodeInvalidParams {
		t.Errorf("code = %d, want %d (jsonrpc.CodeInvalidParams)", jerr.Code, jsonrpc.CodeInvalidParams)
	}
}

// --- Registration (the SPEC-041 bar: registration, not shape) ---

func TestResources_Listed(t *testing.T) {
	cs, _ := newTestServer(t, "claude-code")
	lr, err := cs.ListResources(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var uris []string
	for _, r := range lr.Resources {
		uris = append(uris, r.URI)
		if r.MIMEType != "text/markdown" {
			t.Errorf("%s: MIMEType = %q, want text/markdown", r.URI, r.MIMEType)
		}
		if r.Description == "" {
			t.Errorf("%s: Description is empty", r.URI)
		}
		if r.Size != 0 {
			t.Errorf("%s: Size = %d, want 0 (LD2 — Size is left at its zero value)", r.URI, r.Size)
		}
	}
	sort.Strings(uris)
	want := []string{"brag://memory/recent", "brag://projects"}
	if !reflect.DeepEqual(uris, want) {
		t.Errorf("resources = %v, want %v", uris, want)
	}
}

func TestResourceTemplates_Listed(t *testing.T) {
	cs, _ := newTestServer(t, "claude-code")
	lt, err := cs.ListResourceTemplates(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(lt.ResourceTemplates) != 1 {
		t.Fatalf("resource templates = %d, want 1: %+v", len(lt.ResourceTemplates), lt.ResourceTemplates)
	}
	rt := lt.ResourceTemplates[0]
	if rt.URITemplate != tmplMemoryProject {
		t.Errorf("URITemplate = %q, want %q", rt.URITemplate, tmplMemoryProject)
	}
	if rt.MIMEType != "text/markdown" {
		t.Errorf("MIMEType = %q, want text/markdown", rt.MIMEType)
	}
	if !strings.Contains(rt.Description, "a soft boost, not a filter") {
		t.Errorf("Description missing %q: %q", "a soft boost, not a filter", rt.Description)
	}
}

func TestResourceTemplate_MatchesConcreteURI(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "seed")
	res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "brag://memory/project/bragfile"})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(res.Contents) != 1 {
		t.Fatalf("contents = %d, want 1", len(res.Contents))
	}
	if res.Contents[0].URI != "brag://memory/project/bragfile" {
		t.Errorf("Contents[0].URI = %q, want the concrete URI", res.Contents[0].URI)
	}
}

func TestResources_ReadRoundTripsAllThree(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "seed")
	for _, uri := range []string{uriMemoryRecent, "brag://memory/project/bragfile", uriProjects} {
		res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uri})
		if err != nil {
			t.Fatalf("read %s: %v", uri, err)
		}
		if len(res.Contents) != 1 {
			t.Fatalf("read %s: contents = %d, want 1", uri, len(res.Contents))
		}
		c := res.Contents[0]
		if c.MIMEType != "text/markdown" {
			t.Errorf("read %s: MIMEType = %q, want text/markdown", uri, c.MIMEType)
		}
		if c.Text == "" {
			t.Errorf("read %s: Text is empty", uri)
		}
	}
}

// --- Body parity (LD1, LD5) ---

func TestResourceRecent_IsByteIdenticalToMemoryMarkdown(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one", "two", "three")
	fixed := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	restore := setNowFunc(t, func() time.Time { return fixed })
	defer restore()

	res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uriMemoryRecent})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	want := wantMemoryMarkdown(t, s, "", "", memory.DefaultBudget, "(none)", map[string]string{}, fixed)
	if res.Contents[0].Text != string(want) {
		t.Errorf("body mismatch:\n got=%s\nwant=%s", res.Contents[0].Text, want)
	}
}

// TestResourceRecent_CarriesTheCandidatesHeader pins the AGENT-VISIBLE bytes
// (DEC-045, DEC-048). The byte-identity test above compares two renderings of
// the SAME function, so it passes whatever the header says — it proves parity,
// not correctness. Nothing in this package would have failed if the header
// regressed; this is the assertion that would.
func TestResourceRecent_CarriesTheCandidatesHeader(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one", "two", "three")
	fixed := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	restore := setNowFunc(t, func() time.Time { return fixed })
	defer restore()

	res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uriMemoryRecent})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	body := res.Contents[0].Text
	lines := strings.Split(body, "\n")
	for _, ln := range lines {
		if strings.HasPrefix(ln, "Entries:") {
			t.Errorf("resource body still carries an Entries: line: %q", ln)
		}
	}
	if len(lines) < 6 {
		t.Fatalf("body has %d lines, want at least 6:\n%s", len(lines), body)
	}
	if lines[5] != "Candidates: 3" {
		t.Errorf("line 6 = %q, want %q\nbody:\n%s", lines[5], "Candidates: 3", body)
	}
}

func TestResourceRecent_UsesDefaultBudget(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one")
	res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uriMemoryRecent})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	want := fmt.Sprintf("- Budget: %d tokens", memory.DefaultBudget)
	if !strings.Contains(res.Contents[0].Text, want) {
		t.Errorf("body missing %q:\n%s", want, res.Contents[0].Text)
	}
}

// --- The soft boost, proven positively (LD4) ---

func TestResourceProject_IncludesOtherProjects(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	if _, err := s.Add(storage.Entry{Title: "orbit work", Project: "orbit"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := s.Add(storage.Entry{Title: "bragfile work", Project: "bragfile"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "brag://memory/project/orbit"})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	body := res.Contents[0].Text
	if !strings.Contains(body, "[bragfile/") {
		t.Errorf("expected at least one [bragfile/...] line (the soft boost), got:\n%s", body)
	}

	var firstSliceLine string
	inSlice := false
	for _, ln := range strings.Split(body, "\n") {
		if ln == "## Slice" {
			inSlice = true
			continue
		}
		if inSlice && strings.HasPrefix(ln, "- ") {
			firstSliceLine = ln
			break
		}
	}
	if !strings.Contains(firstSliceLine, "[orbit/") {
		t.Errorf("expected the first slice line to be an [orbit/...] one, got %q in body:\n%s", firstSliceLine, body)
	}
}

func TestResourceProject_UnknownNameStillReturnsASlice(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one", "two")
	res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "brag://memory/project/zzz-nonexistent"})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(res.Contents[0].Text, "## Slice") {
		t.Errorf("expected a non-empty ## Slice section, got:\n%s", res.Contents[0].Text)
	}
}

func TestResourceProject_EchoesTheProjectInFilters(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one")
	res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "brag://memory/project/orbit"})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	want := "Filters: --project orbit"
	if !strings.Contains(res.Contents[0].Text, want) {
		t.Errorf("body missing %q:\n%s", want, res.Contents[0].Text)
	}
}

func TestResourceProject_EmptyNameIsInvalidParams(t *testing.T) {
	cs, _ := newTestServer(t, "claude-code")
	_, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "brag://memory/project/"})
	assertInvalidParams(t, err)
}

func TestResourceProject_PercentDecodesTheName(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	if _, err := s.Add(storage.Entry{Title: "space project work", Project: "my project"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "brag://memory/project/my%20project"})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	want := "Filters: --project my project"
	if !strings.Contains(res.Contents[0].Text, want) {
		t.Errorf("body missing %q:\n%s", want, res.Contents[0].Text)
	}
}

func TestResourceProject_SlashInNameDoesNotMatch(t *testing.T) {
	cs, _ := newTestServer(t, "claude-code")
	_, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "brag://memory/project/a/b"})
	assertInvalidParams(t, err)
}

func TestResource_UnregisteredURIIsInvalidParams(t *testing.T) {
	cs, _ := newTestServer(t, "claude-code")
	_, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "brag://nope"})
	assertInvalidParams(t, err)
}

// --- Cache semantics (LD3) ---

// TestResources_CacheHints ▲ ttlMs is OURS and chosen (0 — the corpus changes
// on every capture, so a cached slice is worse than no slice at all);
// cacheScope is the SDK's and OBSERVED — Server.readResource clobbers it to
// "public" unconditionally regardless of what a handler sets. If this test
// ever reddens on the cacheScope clause specifically, the SDK stopped
// clobbering and DEC-045 sub-decision 5 should then set "private".
func TestResources_CacheHints(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "one")
	for _, uri := range []string{uriMemoryRecent, "brag://memory/project/orbit", uriProjects} {
		res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uri})
		if err != nil {
			t.Fatalf("read %s: %v", uri, err)
		}
		if res.TTLMs != 0 {
			t.Errorf("%s: TTLMs = %d, want 0", uri, res.TTLMs)
		}
		if res.CacheScope != "public" {
			t.Errorf("%s: CacheScope = %q, want %q (observed, not chosen)", uri, res.CacheScope, "public")
		}
	}
}

// --- The projects resource (LD10) ---

func TestResourceProjects_ListsNamesStatusAndCount(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	if _, err := s.CreateProject(storage.Project{Name: "bragfile", Status: "active"}); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := s.CreateProject(storage.Project{Name: "orbit", Status: "paused"}); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := s.Add(storage.Entry{Title: "bragfile work", Project: "bragfile"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uriProjects})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	body := res.Contents[0].Text
	for _, want := range []string{"bragfile", "active", "orbit", "paused", "1 brags"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q:\n%s", want, body)
		}
	}
}

func TestResourceProjects_OmitsLocations(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	p, err := s.CreateProject(storage.Project{Name: "bragfile", Status: "active"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if err := s.AddLocation(p.ID, "/tmp/somewhere-distinctive"); err != nil {
		t.Fatalf("add location: %v", err)
	}
	res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uriProjects})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if strings.Contains(res.Contents[0].Text, "/tmp/somewhere-distinctive") {
		t.Errorf("body must not contain the location path:\n%s", res.Contents[0].Text)
	}
}

func TestResourceProjects_ExcludesArchived(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	p, err := s.CreateProject(storage.Project{Name: "gone", Status: "active"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if err := s.ArchiveProject(p.ID); err != nil {
		t.Fatalf("archive: %v", err)
	}
	res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uriProjects})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if strings.Contains(res.Contents[0].Text, "gone") {
		t.Errorf("body must not contain the archived project name:\n%s", res.Contents[0].Text)
	}
}

func TestResourceProjects_EmptyState(t *testing.T) {
	cs, _ := newTestServer(t, "claude-code")
	res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uriProjects})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(res.Contents[0].Text, "brag project ensure") {
		t.Errorf("empty-state body missing 'brag project ensure':\n%s", res.Contents[0].Text)
	}
}

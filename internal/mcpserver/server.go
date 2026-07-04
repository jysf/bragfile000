package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// nowFunc is the clock brag_stats reads (package-level injectable seam, per
// AGENTS.md §9). Kept local (not .UTC()'d) so the streak buckets by local
// day, matching cli.runStats / DEC-022.
var nowFunc = time.Now

// New builds the MCP server advertising exactly four typed tools —
// brag_add, brag_list, brag_search, brag_stats — as thin wrappers over s.
func New(s *storage.Store) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{Name: "brag", Version: "mcp"}, nil)

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "brag_add",
		Description: "Capture a new brag entry. Requires a non-empty title.",
	}, handleAdd(s))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "brag_list",
		Description: "List brag entries, optionally filtered by tag/project/type, capped by limit.",
	}, handleList(s))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "brag_search",
		Description: "Full-text search brag entries (DEC-010 whitespace-tokenized, phrase-AND query).",
	}, handleSearch(s))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "brag_stats",
		Description: "Lifetime stats over the entire corpus (same shape as `brag stats --format json`).",
	}, handleStats(s))

	return srv
}

// addIn is brag_add's input shape. Title has no `,omitempty` so the SDK's
// inferred schema marks it required; the other user-owned fields mirror
// parseAddJSON's addJSONInput (DEC-012). Agent/Model are the explicit
// provenance params (DEC-024 locked decision 4).
type addIn struct {
	Title       string `json:"title" jsonschema:"the entry title (required, non-empty)"`
	Description string `json:"description,omitempty"`
	Tags        string `json:"tags,omitempty" jsonschema:"comma-joined tag list (DEC-004)"`
	Project     string `json:"project,omitempty"`
	Type        string `json:"type,omitempty"`
	Impact      string `json:"impact,omitempty"`
	Agent       string `json:"agent,omitempty" jsonschema:"caller agent name; stamped as agent:<name> (falls back to the MCP client's clientInfo.Name when omitted)"`
	Model       string `json:"model,omitempty" jsonschema:"caller model id; stamped as model:<id> (no transport fallback)"`
}

// handleAdd mirrors parseAddJSON's validation (DEC-012), stamps provenance
// (DEC-024 locked decision 4), inserts via Store, and returns the created
// entry as a single DEC-011 object.
func handleAdd(s *storage.Store) func(context.Context, *mcp.CallToolRequest, addIn) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, req *mcp.CallToolRequest, in addIn) (*mcp.CallToolResult, any, error) {
		if strings.TrimSpace(in.Title) == "" {
			return nil, nil, fmt.Errorf("brag_add: title is required and must not be empty")
		}
		if len(in.Title) > 200 {
			return nil, nil, fmt.Errorf("brag_add: title exceeds 200-character limit")
		}
		if len(in.Description) > 100000 {
			return nil, nil, fmt.Errorf("brag_add: description exceeds 100000-character limit")
		}
		if len(in.Tags) > 64 {
			return nil, nil, fmt.Errorf("brag_add: tags exceeds 64-character limit")
		}
		if len(in.Project) > 64 {
			return nil, nil, fmt.Errorf("brag_add: project exceeds 64-character limit")
		}
		if len(in.Type) > 64 {
			return nil, nil, fmt.Errorf("brag_add: type exceeds 64-character limit")
		}
		if len(in.Impact) > 256 {
			return nil, nil, fmt.Errorf("brag_add: impact exceeds 256-character limit")
		}

		agent := in.Agent
		if agent == "" {
			agent = clientInfoName(req)
		}

		inserted, err := s.Add(storage.Entry{
			Title:       in.Title,
			Description: in.Description,
			Tags:        stampProvenance(in.Tags, agent, in.Model),
			Project:     in.Project,
			Type:        in.Type,
			Impact:      in.Impact,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("brag_add: %w", err)
		}

		b, err := marshalEntry(inserted)
		if err != nil {
			return nil, nil, fmt.Errorf("brag_add: %w", err)
		}
		return textResult(b), nil, nil
	}
}

// clientInfoName reads the MCP client's application name from the session's
// initialize params, nil-safe. Returns "" when the transport carries no
// client identity.
func clientInfoName(req *mcp.CallToolRequest) string {
	ip := req.Session.InitializeParams()
	if ip == nil || ip.ClientInfo == nil {
		return ""
	}
	return ip.ClientInfo.Name
}

// entryRecord mirrors export's DEC-011 9-key shape for the single created
// entry brag_add returns (export.entryRecord is package-private, so this is
// the small local marshal the spec's Notes for the Implementer permit).
type entryRecord struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	Project     string `json:"project"`
	Type        string `json:"type"`
	Impact      string `json:"impact"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func marshalEntry(e storage.Entry) ([]byte, error) {
	return json.MarshalIndent(entryRecord{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		Tags:        e.Tags,
		Project:     e.Project,
		Type:        e.Type,
		Impact:      e.Impact,
		CreatedAt:   e.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt.UTC().Format(time.RFC3339),
	}, "", "  ")
}

// listIn is brag_list's input shape: the same exact-match filters `brag
// list` supports minus --since (deferred, see Out of scope — ParseSince
// lives in package cli and importing it would risk a cli↔mcpserver cycle).
type listIn struct {
	Tag     string `json:"tag,omitempty"`
	Project string `json:"project,omitempty"`
	Type    string `json:"type,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

func handleList(s *storage.Store) func(context.Context, *mcp.CallToolRequest, listIn) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, in listIn) (*mcp.CallToolResult, any, error) {
		rows, err := s.List(storage.ListFilter{
			Tag:     in.Tag,
			Project: in.Project,
			Type:    in.Type,
			Limit:   in.Limit,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("brag_list: %w", err)
		}
		b, err := export.ToJSON(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("brag_list: %w", err)
		}
		return textResult(b), nil, nil
	}
}

// searchIn is brag_search's input shape: query is required (no
// `,omitempty`); limit mirrors `brag search --limit` (0 = unlimited).
type searchIn struct {
	Query string `json:"query" jsonschema:"FTS query, whitespace-tokenized and AND-joined (DEC-010, required)"`
	Limit int    `json:"limit,omitempty"`
}

func handleSearch(s *storage.Store) func(context.Context, *mcp.CallToolRequest, searchIn) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, in searchIn) (*mcp.CallToolResult, any, error) {
		match, err := buildMatch(in.Query)
		if err != nil {
			return nil, nil, fmt.Errorf("brag_search: %w", err)
		}
		rows, err := s.Search(match, in.Limit)
		if err != nil {
			return nil, nil, fmt.Errorf("brag_search: %w", err)
		}
		b, err := export.ToJSON(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("brag_search: %w", err)
		}
		return textResult(b), nil, nil
	}
}

// statsIn is brag_stats's input shape: no arguments.
type statsIn struct{}

func handleStats(s *storage.Store) func(context.Context, *mcp.CallToolRequest, statsIn) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, _ statsIn) (*mcp.CallToolResult, any, error) {
		rows, err := s.List(storage.ListFilter{})
		if err != nil {
			return nil, nil, fmt.Errorf("brag_stats: %w", err)
		}
		b, err := export.ToStatsJSON(rows, export.StatsOptions{Now: nowFunc()})
		if err != nil {
			return nil, nil, fmt.Errorf("brag_stats: %w", err)
		}
		return textResult(b), nil, nil
	}
}

// textResult wraps pre-marshaled CLI JSON bytes in the locked tool-output
// shape: a single explicit TextContent block, Out=any (no structured
// content), giving byte-parity with `brag <cmd> --format json`.
func textResult(b []byte) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}
}

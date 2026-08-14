package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jysf/bragfile000/internal/capture"
	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/ftsquery"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/jysf/bragfile000/internal/timewindow"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// nowFunc is the clock brag_stats reads (package-level injectable seam, per
// AGENTS.md §9). Kept local (not .UTC()'d) so the streak buckets by local
// day, matching cli.runStats / DEC-022.
//
// Since SPEC-072 it also anchors brag_list's time filters: `day` resolves its
// calendar-day boundary off nowFunc().Location() (DEC-039) and `since`/`until`
// resolve relative Nd/Nw/Nm durations from it. Do NOT "tidy" this to
// .UTC() — that would silently turn a caller's local day into the UTC day,
// which is the exact skew DEC-039 exists to prevent.
var nowFunc = time.Now

// New builds the MCP server advertising exactly five typed tools —
// brag_add, brag_list, brag_search, brag_stats, brag_memory — as thin
// wrappers over s, plus the SPEC-074 push surface: three brag:// resources
// (a ranked recent-memory slice, a project-boosted template, and a projects
// lookup) an MCP client can auto-load with no tool call. See DEC-045.
func New(s *storage.Store) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{Name: "brag", Version: "mcp"}, nil)

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "brag_add",
		Description: "Capture a new brag entry. Requires a non-empty title. If the work has a proving artifact, add an evidence link as a tag so the claim is checkable: prefer `pr:<number>`, else `commit:<hash on the default branch>` — a squash-merge orphans the branch commit you had while working, so a hash captured mid-session often resolves for nobody. Never a range, and record nothing rather than a hash you cannot confirm. See BRAG.md.",
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

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "brag_memory",
		Description: "Ranked, token-budgeted slice of your history — the corpus as working memory (same shape as `brag memory`).",
	}, handleMemory(s))

	addResources(srv, s)

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
	Session     string `json:"session,omitempty" jsonschema:"stable per-session id; stamped as session:<id> — the PROJ-005 reconciliation join-key (no transport fallback; forward the hook-surfaced session_id)"`
	Cost        string `json:"cost,omitempty" jsonschema:"caller-reported cost in USD (non-negative decimal, e.g. 0.42); stamped as cost:<n>. Never estimated by bragfile."`
	Tokens      string `json:"tokens,omitempty" jsonschema:"caller-reported total tokens (non-negative integer); stamped as tokens:<n>. Never estimated by bragfile."`
}

// handleAdd mirrors parseAddJSON's validation (DEC-012), stamps provenance
// (DEC-024 locked decision 4), inserts via Store, and returns the created
// entry as a single DEC-011 object.
func handleAdd(s *storage.Store) func(context.Context, *mcp.CallToolRequest, addIn) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, req *mcp.CallToolRequest, in addIn) (*mcp.CallToolResult, any, error) {
		if strings.TrimSpace(in.Title) == "" {
			return nil, nil, fmt.Errorf("brag_add: title is required and must not be empty")
		}
		// Byte caps, control-char rejection, and reserved-numeric-tag validation
		// are shared with every capture ingress path (SPEC-064). Validate the
		// caller's raw tags here, before stampProvenance appends the reserved
		// agent:/model:/session:/cost:/tokens: tokens (so the 64-byte tags cap
		// applies to what the caller sent, not the post-stamp string).
		if err := capture.Validate(capture.Fields{
			Title:       in.Title,
			Description: in.Description,
			Tags:        in.Tags,
			Project:     in.Project,
			Type:        in.Type,
			Impact:      in.Impact,
		}); err != nil {
			return nil, nil, fmt.Errorf("brag_add: %w", err)
		}

		agent := in.Agent
		if agent == "" {
			agent = clientInfoName(req)
		}

		// Validate the seed numerics at the handler boundary, before stamping,
		// so a bad value is a tool error and never a coerced/inserted tag
		// (DEC-027). session: is opaque and needs no numeric check.
		cost, err := capture.NormalizeCost(in.Cost)
		if err != nil {
			return nil, nil, fmt.Errorf("brag_add: %w", err)
		}
		tokens, err := capture.NormalizeTokens(in.Tokens)
		if err != nil {
			return nil, nil, fmt.Errorf("brag_add: %w", err)
		}

		inserted, err := s.Add(storage.Entry{
			Title:       in.Title,
			Description: in.Description,
			Tags:        stampProvenance(in.Tags, agent, in.Model, in.Session, cost, tokens),
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

// listIn is brag_list's input shape: full parity with `brag list`'s filter
// set, plus `until` (which storage supports and the CLI deliberately does
// not — DEC-042). The time fields use the CLI's exact grammar because they
// are parsed by the same shared package: DEC-008 for since/until, DEC-039
// for day.
type listIn struct {
	Tag     string `json:"tag,omitempty"`
	Project string `json:"project,omitempty"`
	Type    string `json:"type,omitempty"`
	Since   string `json:"since,omitempty" jsonschema:"inclusive lower time bound: YYYY-MM-DD (UTC midnight) or a relative Nd/Nw/Nm (e.g. 7d = last week)"`
	Until   string `json:"until,omitempty" jsonschema:"exclusive upper time bound, same grammar as since; combine with since for a bounded window"`
	Day     string `json:"day,omitempty" jsonschema:"scope to a single LOCAL calendar day: YYYY-MM-DD, today, or yesterday; mutually exclusive with since/until"`
	Author  string `json:"author,omitempty" jsonschema:"provenance filter: \"agent\" (entry carries an agent:/model: tag) or \"human\" (carries neither)"`
	Limit   int    `json:"limit,omitempty" jsonschema:"maximum rows to return; omit or 0 for unlimited"`
}

func handleList(s *storage.Store) func(context.Context, *mcp.CallToolRequest, listIn) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, in listIn) (*mcp.CallToolResult, any, error) {
		filter, err := listFilterFrom(in)
		if err != nil {
			return nil, nil, err
		}
		rows, err := s.List(filter)
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

// listFilterFrom validates brag_list's input and translates it into a
// storage.ListFilter. The validation ORDER is locked (DEC-042) so the error a
// caller sees names the real problem rather than a downstream symptom:
// structural checks (limit, author) first, then the day/since-until conflict
// BEFORE any parsing — a caller who sends both a malformed since and a day
// should be told about the conflict, not the parse failure.
//
// Time values are parsed by internal/timewindow against nowFunc(), which is
// LOCAL by design (see nowFunc): ParseDay reads the calendar-day boundary off
// now.Location(), and ParseSince normalizes to UTC on its own.
func listFilterFrom(in listIn) (storage.ListFilter, error) {
	if in.Limit < 0 {
		return storage.ListFilter{}, fmt.Errorf("brag_list: limit must not be negative, got %d (omit limit for unlimited)", in.Limit)
	}
	switch in.Author {
	case "", "agent", "human":
	default:
		return storage.ListFilter{}, fmt.Errorf("brag_list: author must be %q or %q, got %q", "agent", "human", in.Author)
	}
	if in.Day != "" && (in.Since != "" || in.Until != "") {
		return storage.ListFilter{}, fmt.Errorf("brag_list: day is mutually exclusive with since/until (day sets the full day window)")
	}

	filter := storage.ListFilter{
		Tag:     in.Tag,
		Project: in.Project,
		Type:    in.Type,
		Author:  in.Author,
		Limit:   in.Limit,
	}
	now := nowFunc()
	if in.Since != "" {
		t, err := timewindow.ParseSince(in.Since, now)
		if err != nil {
			return storage.ListFilter{}, fmt.Errorf("brag_list: invalid since %q: %w", in.Since, err)
		}
		filter.Since = t
	}
	if in.Until != "" {
		t, err := timewindow.ParseSince(in.Until, now)
		if err != nil {
			return storage.ListFilter{}, fmt.Errorf("brag_list: invalid until %q: %w", in.Until, err)
		}
		filter.Until = t
	}
	if in.Day != "" {
		start, end, err := timewindow.ParseDay(in.Day, now)
		if err != nil {
			return storage.ListFilter{}, fmt.Errorf("brag_list: invalid day %q: %w", in.Day, err)
		}
		filter.Since = start
		filter.Until = end
	}
	return filter, nil
}

// searchIn is brag_search's input shape: query is required (no
// `,omitempty`); limit mirrors `brag search --limit` (0 = unlimited).
type searchIn struct {
	Query string `json:"query" jsonschema:"FTS query, whitespace-tokenized and AND-joined (DEC-010, required)"`
	Limit int    `json:"limit,omitempty"`
}

func handleSearch(s *storage.Store) func(context.Context, *mcp.CallToolRequest, searchIn) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, in searchIn) (*mcp.CallToolResult, any, error) {
		// A negative limit is a caller mistake, not "unlimited": Store.Search
		// treats limit <= 0 as no LIMIT, so without this guard `-1` silently
		// returns the whole corpus while `brag search --limit -1` errors
		// (v0.5.0 audit item, folded in per DEC-042).
		if in.Limit < 0 {
			return nil, nil, fmt.Errorf("brag_search: limit must not be negative, got %d (omit limit for unlimited)", in.Limit)
		}
		match, err := ftsquery.Build(in.Query)
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

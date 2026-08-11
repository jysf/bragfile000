package mcpserver

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/memory"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// The three SPEC-074 / DEC-045 URIs. brag:// is a custom scheme, not file://
// — nothing on disk corresponds to these; the bytes are computed per read
// from the SQLite corpus (DEC-045 sub-decision 2).
const (
	uriMemoryRecent        = "brag://memory/recent"
	uriProjects            = "brag://projects"
	tmplMemoryProject      = "brag://memory/project/{name}"
	uriMemoryProjectPrefix = "brag://memory/project/"
)

// addResources registers the push surface: two static resources and one
// template, all text/markdown, all pinned to memory.DefaultBudget (LD5). See
// DEC-045 for the reasoning behind every literal below.
func addResources(srv *mcp.Server, s *storage.Store) {
	srv.AddResource(&mcp.Resource{
		Name:        "memory-recent",
		Title:       "Recent memory slice",
		Description: "A ranked, token-budgeted slice of the developer's own work history — the corpus read back as working memory. Blended by recency (reciprocal-rank fusion, DEC-043) and trimmed to an estimated 2000-token budget. Load this before you start work: it is what the developer already knows, so you do not re-derive it. Markdown; the same bytes as `brag memory`, minus the trailing newline that command appends when printing.",
		MIMEType:    "text/markdown",
		URI:         uriMemoryRecent,
	}, handleMemoryRecentResource(s))

	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "memory-project",
		Title:       "Memory slice, boosted toward a project",
		Description: "The same memory slice, with entries in {name} ranked higher. {name} is a soft boost, not a filter: entries from other projects still appear, ranked lower, because a decision made elsewhere is often the one that matters here. Use `brag://projects` to get real project names. Markdown; the same bytes as `brag memory --project <name>`, minus the trailing newline that command appends when printing.",
		MIMEType:    "text/markdown",
		URITemplate: tmplMemoryProject,
	}, handleMemoryProjectResource(s))

	srv.AddResource(&mcp.Resource{
		Name:        "projects",
		Title:       "Registered projects",
		Description: "Every registered, non-archived project with its status and brag count. Use these names verbatim for the `project` field on brag_add and brag_memory, and for the {name} slot of brag://memory/project/{name} — bragfile never auto-registers a name you invent.",
		MIMEType:    "text/markdown",
		URI:         uriProjects,
	}, handleProjectsResource(s))

	// Size is deliberately left at its zero value on all three (LD2): the
	// contents are generated per read from a corpus that changes on every
	// capture, so any byte count set at registration time is a stale
	// prediction. Do not add it — see DEC-045 sub-decision 4.
}

// handleMemoryRecentResource serves brag://memory/recent: the same
// composition and rendering as `brag memory` with no query and no project.
func handleMemoryRecentResource(s *storage.Store) mcp.ResourceHandler {
	return func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		body, err := renderMemory(s, "", "", memory.DefaultBudget, "markdown")
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", uriMemoryRecent, err)
		}
		return resourceResult(uriMemoryRecent, body), nil
	}
}

// handleMemoryProjectResource serves brag://memory/project/{name}: the same
// slice with {name} as a SOFT boost (LD4) — never a filter. The {name}
// segment arrives percent-encoded and unrejected by the SDK for an empty
// segment (both observed at design, §12(b)), so this handler percent-decodes
// it itself (LD7) and treats an empty name as invalid params (LD4's second
// error semantic — the URI promises a project scope and names none).
func handleMemoryProjectResource(s *storage.Store) mcp.ResourceHandler {
	return func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		uri := req.Params.URI
		raw := strings.TrimPrefix(uri, uriMemoryProjectPrefix)
		name, err := url.PathUnescape(raw)
		if err != nil || name == "" {
			// Unwrapped deliberately (errors-wrap-with-context's documented
			// exception): wrapping would hide the *jsonrpc.Error from the
			// client's errors.As and lose the -32602 code (DEC-045 Notes).
			return nil, mcp.ResourceNotFoundError(uri)
		}
		body, rerr := renderMemory(s, "", name, memory.DefaultBudget, "markdown")
		if rerr != nil {
			return nil, fmt.Errorf("read %s: %w", uri, rerr)
		}
		return resourceResult(uri, body), nil
	}
}

// handleProjectsResource serves brag://projects: the template's lookup
// table, so an agent fills {name} with a real project name instead of
// guessing one (a guessed name silently produces an unboosted slice).
func handleProjectsResource(s *storage.Store) mcp.ResourceHandler {
	return func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		rows, err := s.ProjectStatuses()
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", uriProjects, err)
		}
		body, err := export.ToProjectsMarkdown(rows)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", uriProjects, err)
		}
		return resourceResult(uriProjects, body), nil
	}
}

// resourceResult wraps a rendered markdown body in the locked cache-hints
// shape (LD3). TTLMs: 0 is OURS and chosen — the corpus changes on every
// capture, so a cached memory slice is worse than no slice at all.
// CacheScope is deliberately left unset: Server.readResource clobbers it to
// "public" unconditionally regardless of what a handler sets (observed,
// DEC-045 sub-decision 5), so setting anything here would be a fiction.
func resourceResult(uri string, body []byte) *mcp.ReadResourceResult {
	return &mcp.ReadResourceResult{
		Cacheable: mcp.Cacheable{TTLMs: 0},
		Contents: []*mcp.ResourceContents{
			{URI: uri, MIMEType: "text/markdown", Text: string(body)},
		},
	}
}

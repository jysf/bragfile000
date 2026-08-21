package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/memory"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// memoryIn is brag_memory's input shape: full parity with `brag memory`'s
// flag set (query/project/budget/format), deliberately WITHOUT
// since/until/day (DEC-045 sub-decision 8) — the memory slice already prefers
// recent entries, hard windows are brag_list's job, and a window would make
// the envelope's `Scope: lifetime` line untrue. Budget is a *int, not an int,
// so an omitted budget (-> memory.DefaultBudget) is distinguishable from an
// explicit 0 (-> tool error, mirroring DEC-044's UserError) — under a plain
// int the two are the same value.
type memoryIn struct {
	Query   string `json:"query,omitempty" jsonschema:"boost entries matching this full-text query (FTS5, DEC-010 semantics)"`
	Project string `json:"project,omitempty" jsonschema:"boost entries in this project (a soft boost, not a filter)"`
	Budget  *int   `json:"budget,omitempty" jsonschema:"token budget for the entry lines; omitted uses the default (2000); 0 or negative is a tool error"`
	Format  string `json:"format,omitempty" jsonschema:"output format: markdown (default) or json"`
}

// handleMemory mirrors runMemory's validation and composition (DEC-024
// parity contract) over the fifth tool, brag_memory.
func handleMemory(s *storage.Store) func(context.Context, *mcp.CallToolRequest, memoryIn) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, in memoryIn) (*mcp.CallToolResult, any, error) {
		budget := memory.DefaultBudget
		if in.Budget != nil {
			if *in.Budget <= 0 {
				return nil, nil, fmt.Errorf("brag_memory: budget must be greater than zero, got %d (an unbounded slice is brag_list)", *in.Budget)
			}
			budget = *in.Budget
		}

		format := in.Format
		if format == "" {
			format = "markdown"
		}
		if format != "markdown" && format != "json" {
			return nil, nil, fmt.Errorf("brag_memory: unknown format %q (accepted: markdown, json)", format)
		}

		body, err := renderMemory(s, in.Query, in.Project, budget, format)
		if err != nil {
			var qe *memory.ErrQuery
			if errors.As(err, &qe) {
				return nil, nil, fmt.Errorf("brag_memory: %w", qe)
			}
			return nil, nil, fmt.Errorf("brag_memory: %w", err)
		}
		return textResult(body), nil, nil
	}
}

// renderMemory is the ONE composition + rendering path brag_memory and the
// two memory resources all share (SPEC-074 LD9): memory.Gather for the
// three-read pool, memory.Slice for the ranking, then the requested
// export.ToMemory* rendering. DEC-045's "the resource IS the markdown" parity
// claim is only true if one composition and one tokenizer run on every
// caller — not two that happen to agree today.
func renderMemory(s *storage.Store, query, project string, budget int, format string) ([]byte, error) {
	pool, err := memory.Gather(s, memory.GatherOptions{Query: query, Project: project})
	if err != nil {
		return nil, err
	}
	result := memory.Slice(pool.Entries, memory.Options{
		Query:   query,
		Project: project,
		Matched: pool.Matched,
		Budget:  budget,
	})

	filtersMD, filtersJSON := memoryFilters(query, project)
	opts := export.MemoryOptions{
		Filters:     filtersMD,
		FiltersJSON: filtersJSON,
		Now:         nowFunc(),
	}
	if format == "json" {
		return export.ToMemoryJSON(result, opts)
	}
	return export.ToMemoryMarkdown(result, opts)
}

// memoryFilters mirrors cli.echoFiltersForMemory's output shape inline.
// cli.echoFiltersForMemory is package-private in internal/cli, and importing
// it would create the one-way dependency SPEC-074's Notes for the Implementer
// explicitly reject — this is four lines with no drift risk worth a shared
// package, and the parity tests assert the result byte-exactly. Declared
// order is query then project for markdown; the JSON object follows Go's
// alphabetical map encoding (project, then query).
func memoryFilters(query, project string) (string, map[string]string) {
	jsonObj := map[string]string{}
	var parts []string
	if query != "" {
		jsonObj["query"] = query
		parts = append(parts, fmt.Sprintf("--query %s", query))
	}
	if project != "" {
		jsonObj["project"] = project
		parts = append(parts, fmt.Sprintf("--project %s", project))
	}
	if len(parts) == 0 {
		return "(none)", jsonObj
	}
	return strings.Join(parts, " "), jsonObj
}

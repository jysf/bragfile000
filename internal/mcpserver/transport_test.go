package mcpserver

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestServer_StdoutCarriesNoStrayBytes ▲ a full five-tool + resources/read
// round-trip over the in-memory transport must write NOTHING to the process
// os.Stdout. The in-memory transport does not use os.Stdout, so any captured
// bytes are stray human/log pollution — which in production (stdio transport)
// would corrupt the protocol frame stream. The transport-side twin of
// SPEC-039's errBuf.Len()==0. The resources/read path is a NEW code path into
// the same stdio transport (SPEC-074) — without driving it here it would be
// unguarded.
func TestServer_StdoutCarriesNoStrayBytes(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "seed")
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	// drive all five tools
	callJSON(t, cs, "brag_add", map[string]any{"title": "x", "agent": "claude-code"})
	callJSON(t, cs, "brag_list", map[string]any{})
	callJSON(t, cs, "brag_search", map[string]any{"query": "seed"})
	callJSON(t, cs, "brag_stats", map[string]any{})
	callJSON(t, cs, "brag_memory", map[string]any{})
	// and a resources/read on each of the three resources
	for _, uri := range []string{uriMemoryRecent, "brag://memory/project/bragfile", uriProjects} {
		if _, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: uri}); err != nil {
			t.Fatalf("read %s: %v", uri, err)
		}
	}
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if buf.Len() != 0 {
		t.Errorf("os.Stdout must be empty during MCP handling, got %q", buf.String())
	}
}

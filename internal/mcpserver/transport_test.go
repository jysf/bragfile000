package mcpserver

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// TestServer_StdoutCarriesNoStrayBytes ▲ a full four-tool round-trip over the
// in-memory transport must write NOTHING to the process os.Stdout. The
// in-memory transport does not use os.Stdout, so any captured bytes are stray
// human/log pollution — which in production (stdio transport) would corrupt the
// protocol frame stream. The transport-side twin of SPEC-039's errBuf.Len()==0.
func TestServer_StdoutCarriesNoStrayBytes(t *testing.T) {
	cs, s := newTestServer(t, "claude-code")
	seedViaStore(t, s, "seed")
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	// drive all four tools
	callJSON(t, cs, "brag_add", map[string]any{"title": "x", "agent": "claude-code"})
	callJSON(t, cs, "brag_list", map[string]any{})
	callJSON(t, cs, "brag_search", map[string]any{"query": "seed"})
	callJSON(t, cs, "brag_stats", map[string]any{})
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	if buf.Len() != 0 {
		t.Errorf("os.Stdout must be empty during MCP handling, got %q", buf.String())
	}
}

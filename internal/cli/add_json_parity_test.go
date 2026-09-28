package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jysf/bragfile000/internal/cli"
	"github.com/jysf/bragfile000/internal/mcpserver"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant runs one table against
// both machine ingresses (DEC-055). Before it, the two decoders disagreed on
// six of these rows, because encoding/json folds case and the MCP SDK does
// not, and because the SDK drops an earlier value before its schema check.
// Now every row that repeats a top-level key is refused on both, with the
// same message after each ingress's prefix, and nothing is written.
//
// The last two rows repeat a key only inside a value, which is not this
// rule's: neither ingress may call them a repeat. What each then does with
// "id" is the lone-"id" difference DEC-012 already has, and is not pinned
// here.
func TestAddJSONAndBragAdd_AgreeOnEveryRepeatVariant(t *testing.T) {
	longS := string(rune(0x017F)) // ſ, which encoding/json folds to "s"
	escapedImpact := "\x5c" + "u0069mpact"
	rows := []struct {
		name, payload, want string
	}{
		{"plain", `{"title":"t","impact":"REAL","impact":"CLOBBERED"}`, `key "impact" appears more than once`},
		{"case_fold", `{"title":"t","impact":"REAL","Impact":"CLOBBERED"}`, `key "Impact" appears more than once (first as "impact"; keys match case-insensitively)`},
		{"upper_first", `{"title":"t","IMPACT":"REAL","impact":"CLOBBERED"}`, `key "impact" appears more than once (first as "IMPACT"; keys match case-insensitively)`},
		{"unicode_fold", `{"title":"t","tags":"real","tag` + longS + `":"clobbered"}`, `key "tag` + longS + `" appears more than once (first as "tags"; keys match case-insensitively)`},
		{"escaped_key", `{"title":"t","impact":"REAL","` + escapedImpact + `":"CLOBBERED"}`, `key "impact" appears more than once`},
		{"trailing_null", `{"title":"t","impact":"REAL","impact":null}`, `key "impact" appears more than once`},
		{"trailing_empty", `{"title":"t","impact":"REAL","impact":""}`, `key "impact" appears more than once`},
		{"object_then_string", `{"title":"t","impact":{"x":1},"impact":"CLOBBERED"}`, `key "impact" appears more than once`},
		{"string_then_object", `{"title":"t","impact":"REAL","impact":{"x":1}}`, `key "impact" appears more than once`},
		{"array_then_string", `{"title":"t","tags":["a"],"tags":"b"}`, `key "tags" appears more than once`},
		{"string_then_array", `{"title":"t","tags":"a","tags":["b"]}`, `key "tags" appears more than once`},
		{"same_value", `{"title":"t","type":"shipped","type":"shipped"}`, `key "type" appears more than once`},
		{"nested_object", `{"title":"t","id":{"a":1,"a":2}}`, ""},
		{"nested_array", `{"title":"t","id":[{"a":1,"a":2}]}`, ""},
	}
	if !strings.Contains(rows[4].payload, "\x5cu0069") {
		t.Fatalf("escaped_key fixture lost its escape: %s", rows[4].payload)
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			cliRows, cliErr := runAddJSON(t, row.payload)
			mcpText, mcpIsError, mcpRows := callBragAdd(t, row.payload)

			if row.want == "" {
				if cliErr != nil && strings.Contains(cliErr.Error(), "appears more than once") {
					t.Errorf("add --json called a nested repeat a repeat: %v", cliErr)
				}
				if strings.Contains(mcpText, "appears more than once") {
					t.Errorf("brag_add called a nested repeat a repeat: %s", mcpText)
				}
				return
			}
			if !errors.Is(cliErr, cli.ErrUser) {
				t.Fatalf("add --json: want a user error (exit 1), got %v", cliErr)
			}
			if got, want := cliErr.Error(), "user error: --json input: "+row.want; got != want {
				t.Errorf("add --json:\n got %q\nwant %q", got, want)
			}
			if !mcpIsError {
				t.Fatalf("brag_add: want isError, got a success: %s", mcpText)
			}
			if want := "brag_add: " + row.want; mcpText != want {
				t.Errorf("brag_add:\n got %q\nwant %q", mcpText, want)
			}
			if cliRows != 0 || mcpRows != 0 {
				t.Errorf("want nothing written; add --json stored %d, brag_add stored %d", cliRows, mcpRows)
			}
		})
	}
}

// runAddJSON runs `brag add --json` with payload on stdin against a fresh
// store, and returns how many rows it left and the command's error.
func runAddJSON(t *testing.T, payload string) (int, error) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "cli.db")
	root := cli.NewRootCmd("test")
	root.AddCommand(cli.NewAddCmd())
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(strings.NewReader(payload))
	root.SetArgs([]string{"--db", dbPath, "add", "--json"})
	err := root.Execute()
	return countRows(t, dbPath), err
}

// callBragAdd sends payload as brag_add's raw argument bytes to a server on a
// fresh store, and returns the result text, its isError flag, and how many
// rows it left. A map could not carry a repeated key, so the arguments are a
// json.RawMessage.
func callBragAdd(t *testing.T, payload string) (string, bool, int) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "mcp.db")
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	ctx := context.Background()
	ct, st := mcp.NewInMemoryTransports()
	if _, err := mcpserver.New(s).Connect(ctx, st, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "parity", Version: "1"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	r, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "brag_add", Arguments: json.RawMessage(payload)})
	if err != nil {
		t.Fatalf("brag_add transport error: %v", err)
	}
	_ = cs.Close()
	s.Close()
	text := r.Content[0].(*mcp.TextContent).Text
	return text, r.IsError, countRows(t, dbPath)
}

func countRows(t *testing.T, dbPath string) int {
	t.Helper()
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	got, err := s.List(storage.ListFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	return len(got)
}

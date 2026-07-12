package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// canonicalBlock mirrors the fixed server block build wires as
// canonicalBragBlock ({"command":"brag","args":["mcp","serve"]}).
var testCanonicalBlock = mcpServerBlock{Command: "brag", Args: []string{"mcp", "serve"}}

// The byte-exact literals below were produced at design by running the real
// mergeMCPConfig through a scratch program (json.Valid == true on each;
// run-twice byte-equality == true). See SPEC-055 Notes for the Implementer.

const wantCaseA = `{
  "mcpServers": {
    "brag": {
      "command": "brag",
      "args": [
        "mcp",
        "serve"
      ]
    }
  }
}
`

const wantCaseB = `{
  "mcpServers": {
    "brag": {
      "command": "brag",
      "args": [
        "mcp",
        "serve"
      ]
    },
    "other": {
      "command": "other-server",
      "args": [
        "run"
      ]
    }
  },
  "someTopLevelKey": {
    "keep": true
  }
}
`

const caseBInput = `{
  "mcpServers": {
    "other": {
      "command": "other-server",
      "args": ["run"]
    }
  },
  "someTopLevelKey": {"keep": true}
}`

// --- pure mergeMCPConfig (hermetic, no files) ---

func TestMergeMCPConfig_AbsentFile(t *testing.T) {
	out, err := mergeMCPConfig(nil, "brag", testCanonicalBlock)
	if err != nil {
		t.Fatalf("mergeMCPConfig: %v", err)
	}
	if !json.Valid(out) {
		t.Fatalf("output is not valid JSON: %s", out)
	}
	if string(out) != wantCaseA {
		t.Errorf("absent-file merge mismatch:\n got: %q\nwant: %q", out, wantCaseA)
	}
}

func TestMergeMCPConfig_PreservesOtherServerAndKeys(t *testing.T) {
	out, err := mergeMCPConfig([]byte(caseBInput), "brag", testCanonicalBlock)
	if err != nil {
		t.Fatalf("mergeMCPConfig: %v", err)
	}
	if string(out) != wantCaseB {
		t.Errorf("no-clobber merge mismatch:\n got: %q\nwant: %q", out, wantCaseB)
	}
}

func TestMergeMCPConfig_Idempotent(t *testing.T) {
	first, err := mergeMCPConfig([]byte(caseBInput), "brag", testCanonicalBlock)
	if err != nil {
		t.Fatalf("first merge: %v", err)
	}
	second, err := mergeMCPConfig(first, "brag", testCanonicalBlock)
	if err != nil {
		t.Fatalf("second merge: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Errorf("merge not idempotent:\nfirst:  %q\nsecond: %q", first, second)
	}
	if !strings.Contains(string(second), `"other"`) {
		t.Errorf("second merge dropped the other server: %s", second)
	}
}

func TestMergeMCPConfig_OverwritesDifferentBragBlock(t *testing.T) {
	stale := []byte(`{"mcpServers":{"brag":{"command":"OLD","args":["x"]}}}`)
	out, err := mergeMCPConfig(stale, "brag", testCanonicalBlock)
	if err != nil {
		t.Fatalf("mergeMCPConfig: %v", err)
	}
	if string(out) != wantCaseA {
		t.Errorf("stale brag block not overwritten to canonical:\n got: %q\nwant: %q", out, wantCaseA)
	}
	if strings.Contains(string(out), "OLD") || strings.Contains(string(out), `"x"`) {
		t.Errorf("stale values survived: %s", out)
	}
}

func TestMergeMCPConfig_InvalidJSONErrors(t *testing.T) {
	_, err := mergeMCPConfig([]byte("{not json"), "brag", testCanonicalBlock)
	if err == nil {
		t.Fatal("expected error for malformed existing config, got nil")
	}
	if !strings.Contains(err.Error(), "parse existing config") {
		t.Errorf("error should wrap parse context, got: %v", err)
	}
	if errors.Is(err, ErrUser) {
		t.Errorf("a corrupt target file is an internal error, not ErrUser: %v", err)
	}
}

// --- path resolution (injectable-stub table) ---

func TestResolveInstallTarget_Table(t *testing.T) {
	restore := userHomeDir
	userHomeDir = func() (string, error) { return "/home/test", nil }
	t.Cleanup(func() { userHomeDir = restore })

	cases := []struct {
		client, scope, dir string
		want               string
	}{
		{"claude-code", "project", "/proj", "/proj/.mcp.json"},
		{"claude-code", "user", "", "/home/test/.claude.json"},
		{"cursor", "project", "/proj", "/proj/.cursor/mcp.json"},
		{"cursor", "user", "", "/home/test/.cursor/mcp.json"},
	}
	for _, c := range cases {
		got, err := resolveInstallTarget(c.client, c.scope, c.dir)
		if err != nil {
			t.Errorf("resolveInstallTarget(%q,%q,%q): unexpected err %v", c.client, c.scope, c.dir, err)
			continue
		}
		if got != c.want {
			t.Errorf("resolveInstallTarget(%q,%q,%q) = %q, want %q", c.client, c.scope, c.dir, got, c.want)
		}
	}
}

func TestResolveInstallTarget_ClaudeDesktopByOS(t *testing.T) {
	restore := userHomeDir
	userHomeDir = func() (string, error) { return "/home/test", nil }
	t.Cleanup(func() { userHomeDir = restore })

	got, err := resolveInstallTarget("claude-desktop", "user", "")
	switch runtime.GOOS {
	case "darwin":
		want := "/home/test/Library/Application Support/Claude/claude_desktop_config.json"
		if err != nil || got != want {
			t.Errorf("darwin: got (%q, %v), want %q", got, err, want)
		}
	case "windows":
		if err != nil || !strings.HasSuffix(filepath.ToSlash(got), "Claude/claude_desktop_config.json") {
			t.Errorf("windows: got (%q, %v), want a .../Claude/claude_desktop_config.json path", got, err)
		}
	default:
		if !errors.Is(err, ErrUser) {
			t.Errorf("%s: expected ErrUser for unsupported OS, got (%q, %v)", runtime.GOOS, got, err)
		}
	}
}

func TestResolveInstallTarget_UnsupportedCombos(t *testing.T) {
	restore := userHomeDir
	userHomeDir = func() (string, error) { return "/home/test", nil }
	t.Cleanup(func() { userHomeDir = restore })

	cases := []struct{ client, scope, dir string }{
		{"bogus", "project", "/p"},
		{"claude-code", "bogus", "/p"},
		{"claude-desktop", "project", "/p"},
	}
	for _, c := range cases {
		_, err := resolveInstallTarget(c.client, c.scope, c.dir)
		if !errors.Is(err, ErrUser) {
			t.Errorf("resolveInstallTarget(%q,%q,%q): expected ErrUser, got %v", c.client, c.scope, c.dir, err)
		}
	}
}

// --- CLI command (cobra, t.TempDir, split buffers) ---

func newMCPInstallTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	root := NewRootCmd("test")
	root.AddCommand(NewMCPCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

func runMCPInstallCmd(t *testing.T, args ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf := newMCPInstallTestRoot(t)
	root.SetArgs(append([]string{"mcp", "install"}, args...))
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

func TestMCPInstall_WritesAbsentFile(t *testing.T) {
	dir := t.TempDir()
	stdout, stderr, err := runMCPInstallCmd(t, "--client", "claude-code", "--scope", "project", "--dir", dir)
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	target := filepath.Join(dir, ".mcp.json")
	got, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("read %s: %v", target, readErr)
	}
	if string(got) != wantCaseA {
		t.Errorf("written file mismatch:\n got: %q\nwant: %q", got, wantCaseA)
	}
	if stdout != "" {
		t.Errorf("stdout should be empty on a real write, got %q", stdout)
	}
	if !strings.Contains(stderr, "Registered brag MCP server") || !strings.Contains(stderr, target) {
		t.Errorf("stderr should confirm the write + path, got %q", stderr)
	}
}

func TestMCPInstall_DefaultsClaudeCodeProject(t *testing.T) {
	dir := t.TempDir()
	restore := getCwd
	getCwd = func() (string, error) { return dir, nil }
	t.Cleanup(func() { getCwd = restore })

	stdout, stderr, err := runMCPInstallCmd(t) // no flags -> defaults
	if err != nil {
		t.Fatalf("install (defaults): %v", err)
	}
	target := filepath.Join(dir, ".mcp.json")
	if _, statErr := os.Stat(target); statErr != nil {
		t.Errorf("default client/scope should write %s: %v", target, statErr)
	}
	if stdout != "" {
		t.Errorf("stdout should be empty, got %q", stdout)
	}
	if !strings.Contains(stderr, target) {
		t.Errorf("stderr should name the default target, got %q", stderr)
	}
}

func TestMCPInstall_IdempotentPreservesOtherServer(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".mcp.json")
	if err := os.WriteFile(target, []byte(`{"mcpServers":{"other":{"command":"other-server","args":["run"]}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, _, err := runMCPInstallCmd(t, "--dir", dir); err != nil {
		t.Fatalf("first install: %v", err)
	}
	first, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(first), `"brag"`) || !strings.Contains(string(first), `"other"`) {
		t.Fatalf("first run should contain both brag and other: %s", first)
	}

	_, stderr, err := runMCPInstallCmd(t, "--dir", dir)
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	second, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Errorf("re-run not idempotent:\nfirst:  %q\nsecond: %q", first, second)
	}
	if !strings.Contains(stderr, "already registered") || !strings.Contains(stderr, "no changes") {
		t.Errorf("second run should report a no-op, got %q", stderr)
	}
	if !strings.Contains(string(second), `"other"`) {
		t.Errorf("other server lost after re-run: %s", second)
	}
}

func TestMCPInstall_OverwritesStaleBragBlock(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".mcp.json")
	if err := os.WriteFile(target, []byte(`{"mcpServers":{"brag":{"command":"OLD","args":["x"]}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runMCPInstallCmd(t, "--dir", dir); err != nil {
		t.Fatalf("install: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "OLD") {
		t.Errorf("stale brag block survived: %s", got)
	}
	if string(got) != wantCaseA {
		t.Errorf("overwrite mismatch:\n got: %q\nwant: %q", got, wantCaseA)
	}
}

func TestMCPInstall_DryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".mcp.json")
	stdout, stderr, err := runMCPInstallCmd(t, "--dir", dir, "--dry-run")
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Errorf("--dry-run must not write; %s exists (stat err %v)", target, statErr)
	}
	if stdout != wantCaseA {
		t.Errorf("dry-run stdout should be the would-be JSON:\n got: %q\nwant: %q", stdout, wantCaseA)
	}
	if !strings.Contains(stderr, "Would write to") || !strings.Contains(stderr, target) {
		t.Errorf("dry-run stderr should name the target path, got %q", stderr)
	}
}

func TestMCPInstall_DryRunPreservesExistingFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".mcp.json")
	original := []byte(`{"mcpServers":{"other":{"command":"other-server","args":["run"]}}}`)
	if err := os.WriteFile(target, original, 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := runMCPInstallCmd(t, "--dir", dir, "--dry-run")
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, after) {
		t.Errorf("--dry-run altered the existing file:\n before: %q\n after:  %q", original, after)
	}
	if !strings.Contains(stdout, `"brag"`) || !strings.Contains(stdout, `"other"`) {
		t.Errorf("dry-run stdout should show the merged would-be JSON, got %q", stdout)
	}
}

func TestMCPInstall_UnknownClientIsUserError(t *testing.T) {
	stdout, _, err := runMCPInstallCmd(t, "--client", "bogus")
	if !errors.Is(err, ErrUser) {
		t.Errorf("unknown --client should be ErrUser, got %v", err)
	}
	if stdout != "" {
		t.Errorf("stdout must stay empty on error, got %q", stdout)
	}
}

func TestMCPInstall_DesktopProjectIsUserError(t *testing.T) {
	stdout, _, err := runMCPInstallCmd(t, "--client", "claude-desktop", "--scope", "project")
	if !errors.Is(err, ErrUser) {
		t.Errorf("claude-desktop project scope should be ErrUser, got %v", err)
	}
	if stdout != "" {
		t.Errorf("stdout must stay empty on error, got %q", stdout)
	}
}

func TestMCPInstall_DirWithUserScopeIsUserError(t *testing.T) {
	stdout, _, err := runMCPInstallCmd(t, "--scope", "user", "--dir", "/somewhere")
	if !errors.Is(err, ErrUser) {
		t.Errorf("--dir with --scope user should be ErrUser, got %v", err)
	}
	if stdout != "" {
		t.Errorf("stdout must stay empty on error, got %q", stdout)
	}
}

func TestMCPInstall_HelpListsFlagsAndExample(t *testing.T) {
	root, out, _ := newMCPInstallTestRoot(t)
	root.SetArgs([]string{"mcp", "install", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"--client", "--scope", "--dry-run", "brag mcp install --dry-run"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("install --help missing %q:\n%s", want, out.String())
		}
	}
}

func TestMCP_InstallRegistered(t *testing.T) {
	root, out, _ := newMCPInstallTestRoot(t)
	root.SetArgs([]string{"mcp", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"serve", "install"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("`brag mcp --help` should list %q, got %q", want, out.String())
		}
	}
}

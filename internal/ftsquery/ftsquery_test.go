package ftsquery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The union of the two deleted test sets (SPEC-074 LD8): cli's six
// TestBuildFTS5Query_* cases and mcpserver's TestBuildMatch table. Neither
// covered a case the other missed, but both are kept — they were written
// against the same contract from two sides, and the union is the contract.

func TestBuild_SingleWord(t *testing.T) {
	got, err := Build("latency")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := `"latency"`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuild_MultiWordAnd(t *testing.T) {
	got, err := Build("cut latency")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := `"cut" "latency"`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// A hyphen must survive as a literal, not become FTS5's NOT operator — the
// reason every token is phrase-quoted rather than passed through.
func TestBuild_HyphenatedLiteral(t *testing.T) {
	got, err := Build("auth-refactor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := `"auth-refactor"`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuild_EmptyIsError(t *testing.T) {
	got, err := Build("")
	if err == nil {
		t.Fatalf("expected error, got nil (output=%q)", got)
	}
	if got != "" {
		t.Errorf("expected empty output on error, got %q", got)
	}
}

func TestBuild_WhitespaceOnlyIsError(t *testing.T) {
	if _, err := Build("   \t  "); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBuild_QuoteInQueryIsError(t *testing.T) {
	if _, err := Build(`with "quote"`); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestBuild_Table is mcpserver's deleted TestBuildMatch, retargeted. Kept
// alongside the per-case tests above rather than folded into them: it is the
// compact statement of the same contract, and the two sets were written
// independently.
func TestBuild_Table(t *testing.T) {
	ok := map[string]string{
		"auth":          `"auth"`,
		"cut latency":   `"cut" "latency"`,
		"auth-refactor": `"auth-refactor"`,
	}
	for in, want := range ok {
		got, err := Build(in)
		if err != nil || got != want {
			t.Errorf("Build(%q)=%q,%v want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "   ", `has"quote`} {
		if _, err := Build(bad); err == nil {
			t.Errorf("Build(%q) expected error", bad)
		}
	}
}

// TestPackageIsStdlibOnly pins the "small pure package both cli and mcpserver
// can import" property that makes this extraction safe — an internal import
// here could reintroduce the cli<->mcpserver cycle DEC-042 resolved. Same
// mechanical-guard idiom as internal/timewindow and internal/mcpserver.
func TestPackageIsStdlibOnly(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(b), "bragfile000/internal/") {
			t.Errorf("%s imports another internal package; ftsquery must stay stdlib-only", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

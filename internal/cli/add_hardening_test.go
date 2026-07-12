package cli

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// SPEC-064 — capture input hardening. These tests assert that flag mode
// and editor mode enforce the same byte caps, control-char rejection, and
// reserved-numeric-tag validation that --json / MCP already applied, so
// every add ingress path validates consistently.

// (A) Length caps — flag mode must reject an over-cap field the same way
// --json does. Byte counts (len), matching the SPEC-061 byte decision.
func TestAdd_FlagFieldCapsAreUserError(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"title over 200", []string{"--title", strings.Repeat("x", 201)}},
		{"description over 100000", []string{"--title", "ok", "--description", strings.Repeat("d", 100001)}},
		{"tags over 64", []string{"--title", "ok", "--tags", strings.Repeat("t", 65)}},
		{"project over 64", []string{"--title", "ok", "--project", strings.Repeat("p", 65)}},
		{"type over 64", []string{"--title", "ok", "--type", strings.Repeat("y", 65)}},
		{"impact over 256", []string{"--title", "ok", "--impact", strings.Repeat("i", 257)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, dbPath := newRootWithAdd(t)
			var outBuf, errBuf bytes.Buffer
			root.SetOut(&outBuf)
			root.SetErr(&errBuf)
			root.SetArgs(append([]string{"--db", dbPath, "add"}, tc.args...))

			err := root.Execute()
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, ErrUser) {
				t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
			}
			if outBuf.Len() != 0 {
				t.Errorf("expected stdout empty, got %q", outBuf.String())
			}
			if got := len(listAll(t, dbPath)); got != 0 {
				t.Errorf("expected 0 entries, got %d", got)
			}
		})
	}
}

// (A) A field exactly at its cap must still be accepted (boundary check).
func TestAdd_FlagTitleAtCapAccepted(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", strings.Repeat("x", 200)})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := len(listAll(t, dbPath)); got != 1 {
		t.Fatalf("expected 1 entry, got %d", got)
	}
}

// (A) Editor mode must enforce the same caps as flag mode.
func TestAdd_EditorOverCapFieldIsUserError(t *testing.T) {
	installAddEditFunc(t, func(path string) error {
		body := "Title: ok\nImpact: " + strings.Repeat("i", 257) + "\n\n"
		return os.WriteFile(path, []byte(body), 0o600)
	})
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
	}
	if got := len(listAll(t, dbPath)); got != 0 {
		t.Errorf("expected 0 entries, got %d", got)
	}
}

// (B) Control chars in a single-line field (title) — newline, tab, and NUL
// must all be rejected at ingress on the flag path.
func TestAdd_FlagControlCharSingleLineFieldIsUserError(t *testing.T) {
	cases := []struct {
		name  string
		value string
	}{
		{"newline", "line1\nline2"},
		{"tab", "a\tb"},
		{"nul", "a\x00b"},
		{"carriage return", "a\rb"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, dbPath := newRootWithAdd(t)
			var outBuf, errBuf bytes.Buffer
			root.SetOut(&outBuf)
			root.SetErr(&errBuf)
			root.SetArgs([]string{"--db", dbPath, "add", "--title", tc.value})

			err := root.Execute()
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, ErrUser) {
				t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
			}
			if got := len(listAll(t, dbPath)); got != 0 {
				t.Errorf("expected 0 entries, got %d", got)
			}
		})
	}
}

// (B) description is multi-line: newline and tab are ALLOWED, but an
// embedded NUL is rejected (SQLite would silently truncate it).
func TestAdd_FlagDescriptionAllowsNewlineRejectsNUL(t *testing.T) {
	t.Run("newline+tab accepted", func(t *testing.T) {
		root, dbPath := newRootWithAdd(t)
		var outBuf, errBuf bytes.Buffer
		root.SetOut(&outBuf)
		root.SetErr(&errBuf)
		root.SetArgs([]string{"--db", dbPath, "add", "--title", "ok", "--description", "line1\n\tline2"})

		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		entries := listAll(t, dbPath)
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
	})
	t.Run("nul rejected", func(t *testing.T) {
		root, dbPath := newRootWithAdd(t)
		var outBuf, errBuf bytes.Buffer
		root.SetOut(&outBuf)
		root.SetErr(&errBuf)
		root.SetArgs([]string{"--db", dbPath, "add", "--title", "ok", "--description", "bad\x00body"})

		err := root.Execute()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrUser) {
			t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
		}
		if got := len(listAll(t, dbPath)); got != 0 {
			t.Errorf("expected 0 entries, got %d", got)
		}
	})
}

// (B) An accepted entry's `brag list` output stays exactly one line — the
// direct consequence of rejecting embedded newlines at ingress.
func TestAdd_AcceptedEntryListsOnOneLine(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "no controls here"})
	if err := root.Execute(); err != nil {
		t.Fatalf("add: unexpected error: %v", err)
	}

	root2 := newRootForList(t)
	var listOut, listErr bytes.Buffer
	root2.SetOut(&listOut)
	root2.SetErr(&listErr)
	root2.SetArgs([]string{"--db", dbPath, "list"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("list: unexpected error: %v", err)
	}
	got := strings.TrimRight(listOut.String(), "\n")
	if strings.Contains(got, "\n") {
		t.Errorf("expected single-line list output, got %q", listOut.String())
	}
	if !strings.Contains(got, "no controls here") {
		t.Errorf("expected list output to contain the title, got %q", got)
	}
}

// (C) A reserved cost:/tokens: token smuggled through the freeform --tags
// field must be validated with the same rules as the dedicated params and
// rejected when invalid.
func TestAdd_FlagReservedNumericTagIsUserError(t *testing.T) {
	cases := []struct {
		name string
		tags string
	}{
		{"negative cost", "cost:-9"},
		{"non-numeric tokens", "tokens:xyz"},
		{"cost with currency", "cost:$5"},
		{"cost among others", "perf,cost:-1,shipped"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, dbPath := newRootWithAdd(t)
			var outBuf, errBuf bytes.Buffer
			root.SetOut(&outBuf)
			root.SetErr(&errBuf)
			root.SetArgs([]string{"--db", dbPath, "add", "--title", "ok", "--tags", tc.tags})

			err := root.Execute()
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, ErrUser) {
				t.Fatalf("expected errors.Is(err, ErrUser); got %v", err)
			}
			if got := len(listAll(t, dbPath)); got != 0 {
				t.Errorf("expected 0 entries, got %d", got)
			}
		})
	}
}

// (C) A valid cost: token and a non-numeric provenance token (agent:) in
// freeform tags must both be accepted — the CLI has no dedicated provenance
// flags, so --tags is the documented provenance path and must keep working.
func TestAdd_FlagValidReservedTagsAccepted(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "ok", "--tags", "cost:12.50,tokens:18000,agent:claude"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries := listAll(t, dbPath)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Tags != "cost:12.50,tokens:18000,agent:claude" {
		t.Errorf("stored tags = %q", entries[0].Tags)
	}
}

// newRootForList builds a root command with the list subcommand attached.
// Separate from the add helpers so the one-line-output assertion above can
// drive `brag list` against the same temp DB.
func newRootForList(t *testing.T) *cobra.Command {
	t.Helper()
	root := NewRootCmd("test")
	root.AddCommand(NewListCmd())
	return root
}

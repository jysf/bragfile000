package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
	"github.com/jysf/bragfile000/internal/storage/storagetest"
	"github.com/spf13/cobra"
)

// newMemoryTestRoot builds a fresh root with the memory subcommand attached
// and isolates BRAGFILE_DB. Mirrors newCoverageTestRoot/newSparkTestRoot.
func newMemoryTestRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	root := NewRootCmd("test")
	root.AddCommand(NewMemoryCmd())
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	return root, outBuf, errBuf
}

func runMemoryCmd(t *testing.T, dbPath string, args ...string) (stdout, stderr string, runErr error) {
	t.Helper()
	root, outBuf, errBuf := newMemoryTestRoot(t)
	full := append([]string{"--db", dbPath, "memory"}, args...)
	root.SetArgs(full)
	runErr = root.Execute()
	return outBuf.String(), errBuf.String(), runErr
}

// memoryFixtureRow is one row of the shared 8-entry fixture (SPEC-073 Notes
// → "The fixture"), keyed by its intended id so seedMemoryFixture can add
// rows in ascending-id order (Store.Add's own autoincrement) and then
// storagetest.Backdate each to its fixture created_at.
type memoryFixtureRow struct {
	entry     storage.Entry
	createdAt string
}

func memoryFixtureRows() []memoryFixtureRow {
	return []memoryFixtureRow{
		{storage.Entry{Title: "Auth is mostly caching", Type: "learned"}, "2026-01-09T21:45:00Z"},
		{storage.Entry{Title: "Auth token rotation", Project: "orbit", Type: "shipped", Impact: "removed the last shared secret"}, "2026-03-14T10:10:00Z"},
		{storage.Entry{Title: "FTS5 triggers fight attach", Project: "bragfile", Type: "learned"}, "2026-05-02T19:30:00Z"},
		{storage.Entry{Title: "Auth refactor for the gateway", Project: "orbit", Type: "shipped", Impact: "cut p99 login latency from 600ms to 120ms"}, "2026-06-11T14:25:00Z"},
		{storage.Entry{Title: "Read the auth spec"}, "2026-07-20T08:00:00Z"},
		{storage.Entry{Title: "Sparkline pulse command", Project: "bragfile", Type: "shipped"}, "2026-08-01T11:02:00Z"},
		{storage.Entry{Title: "Retry storms need jitter", Project: "orbit", Type: "learned"}, "2026-08-05T17:40:00Z"},
		{storage.Entry{Title: "MCP list filter parity", Project: "bragfile", Type: "shipped", Impact: "agents can ask for a bounded window in one call"}, "2026-08-07T09:15:00Z"},
	}
}

// seedMemoryFixture adds the 8 fixture rows (ids assigned 1..8 by insertion
// order) then storagetest.Backdates each to its fixture created_at — the
// sanctioned way for a non-storage test to seed past-dated rows.
func seedMemoryFixture(t *testing.T, dbPath string) {
	t.Helper()
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	rows := memoryFixtureRows()
	added := make([]storage.Entry, len(rows))
	for i, r := range rows {
		e, err := s.Add(r.entry)
		if err != nil {
			t.Fatalf("add entry %q: %v", r.entry.Title, err)
		}
		added[i] = e
	}
	s.Close()

	for i, r := range rows {
		at, err := time.Parse(time.RFC3339, r.createdAt)
		if err != nil {
			t.Fatalf("parse fixture createdAt %q: %v", r.createdAt, err)
		}
		if err := storagetest.Backdate(dbPath, added[i].ID, at); err != nil {
			t.Fatalf("backdate id %d: %v", added[i].ID, err)
		}
	}
}

const memoryCmdGolden1 = `# Bragfile Memory

Generated: 2026-08-08T12:00:00Z
Scope: lifetime
Filters: --query auth --project orbit
Entries: 8

## Slice

- 4 2026-06-11 [orbit/shipped] Auth refactor for the gateway — cut p99 login latency from 600ms to 120ms
- 2 2026-03-14 [orbit/shipped] Auth token rotation — removed the last shared secret
- 7 2026-08-05 [orbit/learned] Retry storms need jitter
- 5 2026-07-20 [-/-] Read the auth spec
- 1 2026-01-09 [-/learned] Auth is mostly caching
- 8 2026-08-07 [bragfile/shipped] MCP list filter parity — agents can ask for a bounded window in one call
- 6 2026-08-01 [bragfile/shipped] Sparkline pulse command
- 3 2026-05-02 [bragfile/learned] FTS5 triggers fight attach

## Budget

- Budget: 2000 tokens
- Estimated: 143 tokens
- Included: 8
- Skipped: 0
`

const memoryCmdGolden2 = `# Bragfile Memory

Generated: 2026-08-08T12:00:00Z
Scope: lifetime
Filters: (none)
Entries: 8

## Slice

- 8 2026-08-07 [bragfile/shipped] MCP list filter parity — agents can ask for a bounded window in one call
- 7 2026-08-05 [orbit/learned] Retry storms need jitter
- 6 2026-08-01 [bragfile/shipped] Sparkline pulse command
- 5 2026-07-20 [-/-] Read the auth spec
- 4 2026-06-11 [orbit/shipped] Auth refactor for the gateway — cut p99 login latency from 600ms to 120ms
- 3 2026-05-02 [bragfile/learned] FTS5 triggers fight attach
- 2 2026-03-14 [orbit/shipped] Auth token rotation — removed the last shared secret
- 1 2026-01-09 [-/learned] Auth is mostly caching

## Budget

- Budget: 2000 tokens
- Estimated: 143 tokens
- Included: 8
- Skipped: 0
`

// TestMemoryCmd_EndToEndMarkdownGolden proves the declared Matched order is
// real: the CLI does not receive [5 1 2 4], it derives it from FTS5.
// Verified at design (SPEC-073 Notes → §12(b) pre-flight → B).
func TestMemoryCmd_EndToEndMarkdownGolden(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedMemoryFixture(t, dbPath)
	withNowFunc(t, time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC))

	out, errStr, err := runMemoryCmd(t, dbPath, "--query", "auth", "--project", "orbit")
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
	}
	if out != memoryCmdGolden1 {
		t.Errorf("stdout mismatch:\n--- got ---\n%s\n--- want ---\n%s", out, memoryCmdGolden1)
	}
}

// TestMemoryCmd_BareInvocationIsPlainRecency: brag memory with no flags
// produces Golden 2, and Filters: (none).
func TestMemoryCmd_BareInvocationIsPlainRecency(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedMemoryFixture(t, dbPath)
	withNowFunc(t, time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC))

	out, errStr, err := runMemoryCmd(t, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
	}
	if out != memoryCmdGolden2 {
		t.Errorf("stdout mismatch:\n--- got ---\n%s\n--- want ---\n%s", out, memoryCmdGolden2)
	}
}

// TestMemoryCmd_JSONAndMarkdownSelectTheSameSlice pins DEC-044 sub-decision
// 6: the same options select the same ids in the same order across formats;
// fails if the budget is ever measured against the JSON bytes.
func TestMemoryCmd_JSONAndMarkdownSelectTheSameSlice(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedMemoryFixture(t, dbPath)
	withNowFunc(t, time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC))

	mdOut, _, err := runMemoryCmd(t, dbPath, "--query", "auth", "--project", "orbit", "--budget", "60")
	if err != nil {
		t.Fatalf("markdown run: %v", err)
	}
	jsonOut, _, err := runMemoryCmd(t, dbPath, "--query", "auth", "--project", "orbit", "--budget", "60", "--format", "json")
	if err != nil {
		t.Fatalf("json run: %v", err)
	}

	var mdIDs []int64
	for _, ln := range strings.Split(mdOut, "\n") {
		if !strings.HasPrefix(ln, "- ") {
			continue
		}
		var id int64
		if _, err := fmt.Sscanf(ln, "- %d", &id); err == nil {
			mdIDs = append(mdIDs, id)
		}
	}

	var env struct {
		Slice []struct {
			ID int64 `json:"id"`
		} `json:"slice"`
	}
	if err := json.Unmarshal([]byte(jsonOut), &env); err != nil {
		t.Fatalf("unmarshal json output: %v\n%s", err, jsonOut)
	}
	var jsonIDs []int64
	for _, it := range env.Slice {
		jsonIDs = append(jsonIDs, it.ID)
	}

	if len(mdIDs) == 0 {
		t.Fatal("no markdown bullet ids parsed")
	}
	if fmt.Sprint(mdIDs) != fmt.Sprint(jsonIDs) {
		t.Errorf("markdown ids = %v, json ids = %v", mdIDs, jsonIDs)
	}
}

// TestMemoryCmd_ThreeReadsComposeThePool proves DEC-043 sub-decision 5's
// "one bounded read per list": 200 filler entries (recency-capped) plus two
// entries backdated to 2020 (outside PoolLimit) that are reachable ONLY via
// their own list's read.
func TestMemoryCmd_ThreeReadsComposeThePool(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	for i := 1; i <= 200; i++ {
		if _, err := s.Add(storage.Entry{Title: fmt.Sprintf("filler %d", i)}); err != nil {
			t.Fatalf("add filler %d: %v", i, err)
		}
	}
	orbitEntry, err := s.Add(storage.Entry{Project: "orbit", Title: "ancient orbit note"})
	if err != nil {
		t.Fatalf("add id 201: %v", err)
	}
	authEntry, err := s.Add(storage.Entry{Title: "ancient auth note"})
	if err != nil {
		t.Fatalf("add id 202: %v", err)
	}
	s.Close()

	if err := storagetest.Backdate(dbPath, orbitEntry.ID, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("backdate id 201: %v", err)
	}
	if err := storagetest.Backdate(dbPath, authEntry.ID, time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("backdate id 202: %v", err)
	}

	withNowFunc(t, time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC))

	orbitBullet := fmt.Sprintf("- %d ", orbitEntry.ID)
	authBullet := fmt.Sprintf("- %d ", authEntry.ID)

	t.Run("project-read-reaches-201", func(t *testing.T) {
		out, _, err := runMemoryCmd(t, dbPath, "--project", "orbit", "--budget", "20000")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, orbitBullet) {
			t.Errorf("expected %q in output, missing it:\n%s", orbitBullet, out)
		}
	})

	t.Run("match-read-reaches-202", func(t *testing.T) {
		out, _, err := runMemoryCmd(t, dbPath, "--query", "auth", "--budget", "20000")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, authBullet) {
			t.Errorf("expected %q in output, missing it:\n%s", authBullet, out)
		}
	})

	t.Run("bare-recency-read-is-capped", func(t *testing.T) {
		out, _, err := runMemoryCmd(t, dbPath, "--budget", "20000")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(out, orbitBullet) {
			t.Errorf("bare invocation unexpectedly reached id %d (PoolLimit not enforced):\n%s", orbitEntry.ID, out)
		}
		if strings.Contains(out, authBullet) {
			t.Errorf("bare invocation unexpectedly reached id %d (PoolLimit not enforced):\n%s", authEntry.ID, out)
		}
	})
}

// TestMemoryCmd_ZeroBudgetIsUserError: --budget 0 returns an error wrapping
// ErrUser; outBuf is empty and the message names the flag.
func TestMemoryCmd_ZeroBudgetIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	root, outBuf, _ := newMemoryTestRoot(t)
	root.SetArgs([]string{"--db", dbPath, "memory", "--budget", "0"})
	err := root.Execute()
	if err == nil || !errors.Is(err, ErrUser) {
		t.Fatalf("expected UserError, got %v", err)
	}
	if !strings.Contains(err.Error(), "--budget") {
		t.Errorf("expected error to name --budget, got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected empty stdout, got %q", outBuf.String())
	}
}

// TestMemoryCmd_NegativeBudgetIsUserError: same for --budget -1.
func TestMemoryCmd_NegativeBudgetIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	root, outBuf, _ := newMemoryTestRoot(t)
	root.SetArgs([]string{"--db", dbPath, "memory", "--budget", "-1"})
	err := root.Execute()
	if err == nil || !errors.Is(err, ErrUser) {
		t.Fatalf("expected UserError, got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected empty stdout, got %q", outBuf.String())
	}
}

// TestMemoryCmd_UnknownFormatIsUserError: --format yaml errors, naming
// markdown and json.
func TestMemoryCmd_UnknownFormatIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	root, outBuf, _ := newMemoryTestRoot(t)
	root.SetArgs([]string{"--db", dbPath, "memory", "--format", "yaml"})
	err := root.Execute()
	if err == nil || !errors.Is(err, ErrUser) {
		t.Fatalf("expected UserError, got %v", err)
	}
	if !strings.Contains(err.Error(), "markdown") || !strings.Contains(err.Error(), "json") {
		t.Errorf("expected error to name markdown and json, got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected empty stdout, got %q", outBuf.String())
	}
}

// TestMemoryCmd_EmptyQueryIsUserError: --query "   " errors via
// buildFTS5Query (DEC-010), stdout empty.
func TestMemoryCmd_EmptyQueryIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	root, outBuf, _ := newMemoryTestRoot(t)
	root.SetArgs([]string{"--db", dbPath, "memory", "--query", "   "})
	err := root.Execute()
	if err == nil || !errors.Is(err, ErrUser) {
		t.Fatalf("expected UserError, got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected empty stdout, got %q", outBuf.String())
	}
}

// TestMemoryCmd_QuoteInQueryIsUserError: --query 'a"b' errors (DEC-010).
func TestMemoryCmd_QuoteInQueryIsUserError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	root, outBuf, _ := newMemoryTestRoot(t)
	root.SetArgs([]string{"--db", dbPath, "memory", "--query", `a"b`})
	err := root.Execute()
	if err == nil || !errors.Is(err, ErrUser) {
		t.Fatalf("expected UserError, got %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected empty stdout, got %q", outBuf.String())
	}
}

// TestMemoryCmd_QueryWithNoMatchesStillRenders: --query zzzznomatch produces
// a pure-recency slice and still echoes Filters: --query zzzznomatch.
// Honest degradation, not a silent empty result.
func TestMemoryCmd_QueryWithNoMatchesStillRenders(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedMemoryFixture(t, dbPath)
	withNowFunc(t, time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC))

	out, errStr, err := runMemoryCmd(t, dbPath, "--query", "zzzznomatch")
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
	}
	if !strings.Contains(out, "Filters: --query zzzznomatch") {
		t.Errorf("expected filters echo, got:\n%s", out)
	}
	wantOrder := []string{"- 8 ", "- 7 ", "- 6 ", "- 5 ", "- 4 ", "- 3 ", "- 2 ", "- 1 "}
	lastIdx := -1
	for _, bullet := range wantOrder {
		idx := strings.Index(out, bullet)
		if idx == -1 {
			t.Fatalf("missing bullet %q in:\n%s", bullet, out)
		}
		if idx <= lastIdx {
			t.Errorf("bullet %q out of recency order in:\n%s", bullet, out)
		}
		lastIdx = idx
	}
}

// TestMemoryCmd_ProjectIsSoftBoostNotFilter pins DEC-043 sub-decision 4:
// --project orbit over the fixture returns all 8 entries, with the orbit
// entries leading; fails if ListFilter.Project is ever used as a filter on
// the recency read.
func TestMemoryCmd_ProjectIsSoftBoostNotFilter(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	seedMemoryFixture(t, dbPath)
	withNowFunc(t, time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC))

	out, errStr, err := runMemoryCmd(t, dbPath, "--project", "orbit")
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr: %s)", err, errStr)
	}
	if !strings.Contains(out, "Entries: 8") {
		t.Errorf("expected Entries: 8, got:\n%s", out)
	}
	i7 := strings.Index(out, "- 7 ")
	i4 := strings.Index(out, "- 4 ")
	i2 := strings.Index(out, "- 2 ")
	if i7 == -1 || i4 == -1 || i2 == -1 || !(i7 < i4 && i4 < i2) {
		t.Errorf("expected orbit entries 7, 4, 2 leading in that order, got:\n%s", out)
	}
}

// TestMemoryCmd_HelpShape: help contains the locked phrases and does NOT
// contain --limit or --detail. errBuf.Len() == 0 (AGENTS.md §9 separate-
// buffers rule).
func TestMemoryCmd_HelpShape(t *testing.T) {
	root, outBuf, errBuf := newMemoryTestRoot(t)
	root.SetArgs([]string{"memory", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	help := outBuf.String()
	for _, want := range []string{"reciprocal-rank fusion", "soft boost, not a filter", "not a tokenizer", "Examples:"} {
		if !strings.Contains(help, want) {
			t.Errorf("help missing %q:\n%s", want, help)
		}
	}
	for _, forbidden := range []string{"--limit", "--detail"} {
		if strings.Contains(help, forbidden) {
			t.Errorf("help unexpectedly contains %q:\n%s", forbidden, help)
		}
	}
	if errBuf.Len() != 0 {
		t.Errorf("expected empty stderr, got %q", errBuf.String())
	}
}

// TestMemoryCmd_StdoutOnlyCarriesTheDocument: on success errBuf.Len() == 0;
// on a UserError outBuf.Len() == 0 (stdout-is-for-data-stderr-is-for-humans).
func TestMemoryCmd_StdoutOnlyCarriesTheDocument(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	withNowFunc(t, time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC))

	t.Run("success", func(t *testing.T) {
		root, _, errBuf := newMemoryTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "memory"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if errBuf.Len() != 0 {
			t.Errorf("expected empty stderr, got %q", errBuf.String())
		}
	})

	t.Run("usererror", func(t *testing.T) {
		root, outBuf, _ := newMemoryTestRoot(t)
		root.SetArgs([]string{"--db", dbPath, "memory", "--budget", "0"})
		err := root.Execute()
		if err == nil || !errors.Is(err, ErrUser) {
			t.Fatalf("expected UserError, got %v", err)
		}
		if outBuf.Len() != 0 {
			t.Errorf("expected empty stdout, got %q", outBuf.String())
		}
	})
}

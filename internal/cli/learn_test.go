package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jysf/bragfile000/internal/aggregate"
	"github.com/jysf/bragfile000/internal/storage"
)

// newRootWithLearn mirrors newRootWithAdd: milestone TTY seam pinned off by
// default, learn + add + list registered so the milestone PAIR (below) can
// exercise both verbs against one corpus.
func newRootWithLearn(t *testing.T) (*cobra.Command, string) {
	t.Helper()
	addStderrIsTTY = func() bool { return false }
	t.Cleanup(func() { addStderrIsTTY = defaultStderrIsTTY })
	root := NewRootCmd("test")
	root.AddCommand(NewLearnCmd())
	root.AddCommand(NewAddCmd())
	root.AddCommand(NewListCmd())
	dbPath := filepath.Join(t.TempDir(), "test.db")
	return root, dbPath
}

// TestLearnCmd_PinsFailedType is the core claim: the verb writes the reserved
// value, and `brag list --type failed` gets it back (STAGE-023 criterion 2,
// satisfied by an existing flag).
func TestLearnCmd_PinsFailedType(t *testing.T) {
	root, dbPath := newRootWithLearn(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "learn", "--title", "shared-worker pool did not cut cold starts"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	id, err := strconv.ParseInt(strings.TrimSpace(outBuf.String()), 10, 64)
	if err != nil {
		t.Fatalf("stdout should be the ID alone, got %q", outBuf.String())
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	got, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get(%d): %v", id, err)
	}
	if got.Type != aggregate.FailureType {
		t.Errorf("Type = %q, want %q", got.Type, aggregate.FailureType)
	}
	if aggregate.FailureType != "failed" {
		t.Errorf("aggregate.FailureType = %q, want %q", aggregate.FailureType, "failed")
	}
}

// TestLearnCmd_NoTypeFlag: the value is pinned, so the flag must not exist.
// A structural check on the flag set, not a substring check on help text —
// the Long deliberately DOES contain "--type" (it teaches `brag list --type
// failed`), so a NOT-contains here would be wrong.
func TestLearnCmd_NoTypeFlag(t *testing.T) {
	cmd := NewLearnCmd()
	if f := cmd.Flags().Lookup("type"); f != nil {
		t.Errorf("brag learn must not define --type, got %v", f)
	}
	if f := cmd.Flags().ShorthandLookup("k"); f != nil {
		t.Errorf("brag learn must not define -k (add's type shorthand), got %v", f)
	}
	for _, name := range []string{"title", "description", "tags", "project", "impact"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("brag learn must define --%s", name)
		}
	}
}

// TestLearnCmd_MilestoneSuppressed_AddStillFires is the PAIR. The negative
// alone proves nothing: stderr is also empty when no milestone would have
// crossed. So the same corpus, the same forced-TTY, the same threshold is
// driven through BOTH verbs — add fires, learn is silent.
func TestLearnCmd_MilestoneSuppressed_AddStillFires(t *testing.T) {
	// negative: learn is the 10th entry, TTY on, nothing on stderr
	root, dbPath := newRootWithLearn(t)
	seedEntries(t, dbPath, 9, "")
	setStderrIsTTY(t, true)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "learn", "--title", "tenth, and it did not work"})
	if err := root.Execute(); err != nil {
		t.Fatalf("learn: unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Errorf("brag learn must not congratulate; stderr = %q", errBuf.String())
	}

	// positive: same threshold, same TTY, via add — the milestone DOES fire,
	// which is what makes the silence above evidence of suppression.
	root2, dbPath2 := newRootWithLearn(t)
	seedEntries(t, dbPath2, 9, "")
	setStderrIsTTY(t, true)
	var outBuf2, errBuf2 bytes.Buffer
	root2.SetOut(&outBuf2)
	root2.SetErr(&errBuf2)
	root2.SetArgs([]string{"--db", dbPath2, "add", "--title", "tenth"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("add: unexpected error: %v", err)
	}
	if !strings.Contains(errBuf2.String(), "🎉 10 brags and counting") {
		t.Fatalf("control failed: add should still fire the milestone, got %q", errBuf2.String())
	}
}

// TestLearnCmd_EditorModeOverwritesUserType: the template omits Type, but a
// user who re-adds the header does not get to redirect the pinned value.
func TestLearnCmd_EditorModeOverwritesUserType(t *testing.T) {
	root, dbPath := newRootWithLearn(t)
	installAddEditFunc(t, func(path string) error {
		return os.WriteFile(path, []byte("Title: bloom filter on the tag join\nType: shipped\n\ntried it; the join was never the bottleneck\n"), 0o600)
	})
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "learn"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	id, err := strconv.ParseInt(strings.TrimSpace(outBuf.String()), 10, 64)
	if err != nil {
		t.Fatalf("stdout should be the ID alone, got %q", outBuf.String())
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	got, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get(%d): %v", id, err)
	}
	if got.Type != aggregate.FailureType {
		t.Errorf("editor-mode Type = %q, want %q (user's header must be overwritten)", got.Type, aggregate.FailureType)
	}
}

// TestLearnCmd_EmptyTitleIsUserError: flag mode requires --title, same as add.
func TestLearnCmd_EmptyTitleIsUserError(t *testing.T) {
	root, dbPath := newRootWithLearn(t)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "learn", "--impact", "cost two days"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected a user error, got nil")
	}
	if !strings.Contains(err.Error(), "--title is required") {
		t.Errorf("error = %v, want it to mention --title is required", err)
	}
}

// runDigestCorpus runs one brag invocation against dbPath on a fresh root
// carrying the two writers (add, learn), the three digests that section
// failures (impact, wrapped, summary — SPEC-095), and story, which labels or
// omits them (SPEC-094). A fresh root per call, so no flag value leaks from
// one invocation into the next.
func runDigestCorpus(t *testing.T, dbPath string, args ...string) string {
	t.Helper()
	t.Setenv("BRAGFILE_DB", "")
	addStderrIsTTY = func() bool { return false }
	t.Cleanup(func() { addStderrIsTTY = defaultStderrIsTTY })
	root := NewRootCmd("test")
	root.AddCommand(NewAddCmd(), NewLearnCmd(), NewImpactCmd(), NewWrappedCmd(), NewSummaryCmd(), NewStoryCmd())
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs(append([]string{"--db", dbPath}, args...))
	if err := root.Execute(); err != nil {
		t.Fatalf("brag %v: %v (stderr %q)", args, err, errBuf.String())
	}
	return outBuf.String()
}

// markdownSection returns the lines under the `## <heading>` line, up to the
// next `## ` heading. Found by line equality, not substring (AGENTS.md §9).
func markdownSection(md, heading string) string {
	var out []string
	in := false
	for _, ln := range strings.Split(md, "\n") {
		if strings.HasPrefix(ln, "## ") {
			in = ln == "## "+heading
			continue
		}
		if in {
			out = append(out, ln)
		}
	}
	return strings.Join(out, "\n")
}

// TestLearnCmd_ImpactSectionsWhatItWrote ▲ SPEC-086 LD1 — the writer and the
// reader held to one value by running both, through a real store: the entry
// `brag learn` wrote is the one `brag impact` lists under "What didn't work",
// and the entry `brag add` wrote on the same corpus stays under "Impact". No
// constant appears here, so this fails if the verb and the digest ever name
// different values, whichever side moves.
func TestLearnCmd_ImpactSectionsWhatItWrote(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	winID := strings.TrimSpace(runDigestCorpus(t, dbPath, "add", "-t", "shipped the cache", "-p", "alpha", "-k", "shipped", "-i", "cut p95 40%"))
	failID := strings.TrimSpace(runDigestCorpus(t, dbPath, "learn", "-t", "tried a worker pool", "-p", "alpha", "-i", "cost two days"))

	md := runDigestCorpus(t, dbPath, "impact", "--since", "2000-01-01")
	impact := markdownSection(md, "Impact")
	failed := markdownSection(md, "What didn't work")
	if !strings.Contains(impact, "- "+winID+": shipped the cache") {
		t.Errorf("## Impact is missing the brag add entry %s:\n%s", winID, md)
	}
	if strings.Contains(impact, "- "+failID+":") {
		t.Errorf("## Impact still carries the brag learn entry %s:\n%s", failID, md)
	}
	if !strings.Contains(failed, "- "+failID+": tried a worker pool\n  cost two days") {
		t.Errorf("## What didn't work is missing the brag learn entry %s with its impact:\n%s", failID, md)
	}

	var env struct {
		ImpactByProject []struct {
			Entries []struct {
				ID int64 `json:"id"`
			} `json:"entries"`
		} `json:"impact_by_project"`
		FailuresByProject []struct {
			Entries []struct {
				ID int64 `json:"id"`
			} `json:"entries"`
		} `json:"failures_by_project"`
	}
	if err := json.Unmarshal([]byte(runDigestCorpus(t, dbPath, "impact", "--since", "2000-01-01", "--format", "json")), &env); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	ids := func(groups []struct {
		Entries []struct {
			ID int64 `json:"id"`
		} `json:"entries"`
	}) string {
		var out []string
		for _, g := range groups {
			for _, e := range g.Entries {
				out = append(out, strconv.FormatInt(e.ID, 10))
			}
		}
		return strings.Join(out, ",")
	}
	if got := ids(env.ImpactByProject); got != winID {
		t.Errorf("impact_by_project ids = %q, want %q", got, winID)
	}
	if got := ids(env.FailuresByProject); got != failID {
		t.Errorf("failures_by_project ids = %q, want %q", got, failID)
	}
}

// TestLearnCmd_WrappedSectionsWhatItWrote ▲ SPEC-086 LD1 — the same
// writer-to-reader check on brag wrapped. The period is named from the stored
// row's own created_at, so the test never races a year boundary.
func TestLearnCmd_WrappedSectionsWhatItWrote(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	winID := strings.TrimSpace(runDigestCorpus(t, dbPath, "add", "-t", "shipped the cache", "-p", "alpha", "-k", "shipped", "-i", "cut p95 40%"))
	failID := strings.TrimSpace(runDigestCorpus(t, dbPath, "learn", "-t", "tried a worker pool", "-p", "alpha", "-i", "cost two days"))

	id, err := strconv.ParseInt(failID, 10, 64)
	if err != nil {
		t.Fatalf("learn stdout should be the id alone, got %q", failID)
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	got, err := s.Get(id)
	s.Close()
	if err != nil {
		t.Fatalf("Get(%d): %v", id, err)
	}
	year := strconv.Itoa(got.CreatedAt.UTC().Year())

	md := runDigestCorpus(t, dbPath, "wrapped", year, "--no-spark")
	moments := markdownSection(md, "Impact moments")
	failed := markdownSection(md, "What didn't work")
	if !strings.Contains(moments, "- "+winID+": shipped the cache") {
		t.Errorf("## Impact moments is missing the brag add entry %s:\n%s", winID, md)
	}
	if strings.Contains(moments, "- "+failID+":") {
		t.Errorf("## Impact moments still carries the brag learn entry %s:\n%s", failID, md)
	}
	if !strings.Contains(failed, "- "+failID+": tried a worker pool\n  cost two days") {
		t.Errorf("## What didn't work is missing the brag learn entry %s with its impact:\n%s", failID, md)
	}
}

// TestLearnCmd_StoryLabelsOrOmitsWhatItWrote ▲ SPEC-094 LD1/LD3/LD4 — the
// writer-to-reader check on brag story, through a real store and with no
// constant in sight. A candid audience lists the brag learn entry as
// "✗ <id> (failed)" with its impact, beside the brag add entry's ★. A
// promotional one leaves it out, and both its Omitted: line and its closing
// clause say so; its JSON counts it.
func TestLearnCmd_StoryLabelsOrOmitsWhatItWrote(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	winID := strings.TrimSpace(runDigestCorpus(t, dbPath, "add", "-t", "shipped the cache", "-p", "alpha", "-k", "shipped", "-i", "cut p95 40%"))
	failID := strings.TrimSpace(runDigestCorpus(t, dbPath, "learn", "-t", "tried a worker pool", "-p", "alpha", "-i", "cost two days"))

	me := runDigestCorpus(t, dbPath, "story", "--audience", "me", "--since", "2000-01-01")
	threads := markdownSection(me, "Threads")
	if !strings.Contains(threads, "- ★ "+winID+": shipped the cache\n  cut p95 40%") {
		t.Errorf("me should list the brag add entry %s as ★:\n%s", winID, me)
	}
	if !strings.Contains(threads, "- ✗ "+failID+" (failed): tried a worker pool\n  cost two days") {
		t.Errorf("me should label the brag learn entry %s as failed, with its impact:\n%s", failID, me)
	}
	if strings.Contains(me, "- ★ "+failID+":") || strings.Contains(me, "\nOmitted: ") {
		t.Errorf("me must neither star the failure nor omit it:\n%s", me)
	}

	exec := runDigestCorpus(t, dbPath, "story", "--audience", "exec", "--since", "2000-01-01")
	if !strings.Contains(markdownSection(exec, "Threads"), "- ★ "+winID+": shipped the cache") {
		t.Errorf("exec should still list the brag add entry %s:\n%s", winID, exec)
	}
	if strings.Contains(exec, " "+failID+": ") || strings.Contains(exec, " "+failID+" (failed)") {
		t.Errorf("exec must not list the brag learn entry %s:\n%s", failID, exec)
	}
	if !strings.Contains(exec, "\nBeats: 1/2\nOmitted: 1 recorded failure, not listed for this audience (brag list --type failed)\n") {
		t.Errorf("exec should count the omitted failure under Beats:\n%s", exec)
	}
	if !strings.HasSuffix(strings.TrimRight(exec, "\n"), "\n\nThis bundle omits 1 recorded failure for this audience. End with one line that says so; do not drop it.") {
		t.Errorf("exec should end with the omission clause:\n%s", exec)
	}

	// --print-directive reads no store, so it prints the asset as authored:
	// the clause depends on the window, and there is none.
	directive := runDigestCorpus(t, dbPath, "story", "--audience", "exec", "--print-directive")
	if !strings.Contains(directive, "business impact") || strings.Contains(directive, "This bundle omits") {
		t.Errorf("--print-directive should print exec's directive without the clause:\n%s", directive)
	}

	var env struct {
		OmittedFailureCount *int `json:"omitted_failure_count"`
	}
	if err := json.Unmarshal([]byte(runDigestCorpus(t, dbPath, "story", "--audience", "exec", "--since", "2000-01-01", "--format", "json")), &env); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if env.OmittedFailureCount == nil || *env.OmittedFailureCount != 1 {
		t.Errorf("exec omitted_failure_count = %v, want 1", env.OmittedFailureCount)
	}
}

// TestLearnCmd_SummarySectionsWhatItWrote ▲ SPEC-095 LD1/LD2/LD5 — the
// writer-to-reader check on brag summary, through a real store and with no
// constant in sight. The brag learn entry carries NO impact on purpose: it is
// the row impact and wrapped leave out, and summary must still list it, under
// "What didn't work" and not under "Highlights" (LD1). Both rows are written
// now and the window is --range week, so the test has no window cliff (LD5):
// it cannot pass on a window that holds no failure. Every negative is paired
// with a positive on the same section.
func TestLearnCmd_SummarySectionsWhatItWrote(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	winID := strings.TrimSpace(runDigestCorpus(t, dbPath, "add", "-t", "shipped the cache", "-p", "alpha", "-k", "shipped"))
	failID := strings.TrimSpace(runDigestCorpus(t, dbPath, "learn", "-t", "tried a worker pool", "-p", "alpha"))

	md := runDigestCorpus(t, dbPath, "summary", "--range", "week")
	highlights := markdownSection(md, "Highlights")
	failed := markdownSection(md, "What didn't work")
	if !strings.Contains(highlights, "- "+winID+": shipped the cache") {
		t.Errorf("## Highlights is missing the brag add entry %s:\n%s", winID, md)
	}
	if strings.Contains(highlights, "- "+failID+":") {
		t.Errorf("## Highlights still carries the brag learn entry %s:\n%s", failID, md)
	}
	if !strings.Contains(failed, "- "+failID+": tried a worker pool") {
		t.Errorf("## What didn't work is missing the brag learn entry %s:\n%s", failID, md)
	}
	if strings.Contains(failed, "- "+winID+":") {
		t.Errorf("## What didn't work carries the brag add entry %s:\n%s", winID, md)
	}
	if !strings.Contains(markdownSection(md, "Summary"), "- failed: 1\n") {
		t.Errorf("By type should still count the brag learn entry:\n%s", md)
	}

	type group struct {
		Entries []struct {
			ID int64 `json:"id"`
		} `json:"entries"`
	}
	var env struct {
		CountsByType      map[string]int `json:"counts_by_type"`
		Highlights        []group        `json:"highlights"`
		FailuresByProject []group        `json:"failures_by_project"`
	}
	if err := json.Unmarshal([]byte(runDigestCorpus(t, dbPath, "summary", "--range", "week", "--format", "json")), &env); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	ids := func(groups []group) string {
		var out []string
		for _, g := range groups {
			for _, e := range g.Entries {
				out = append(out, strconv.FormatInt(e.ID, 10))
			}
		}
		return strings.Join(out, ",")
	}
	if got := ids(env.Highlights); got != winID {
		t.Errorf("highlights ids = %q, want %q", got, winID)
	}
	if got := ids(env.FailuresByProject); got != failID {
		t.Errorf("failures_by_project ids = %q, want %q", got, failID)
	}
	if env.CountsByType["failed"] != 1 || env.CountsByType["shipped"] != 1 {
		t.Errorf("counts_by_type = %v, want failed:1 and shipped:1", env.CountsByType)
	}
}

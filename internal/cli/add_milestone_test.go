package cli

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
	"github.com/jysf/bragfile000/internal/storage/storagetest"
)

// setStderrIsTTY overrides the milestone TTY seam for one test and restores
// the real detector on cleanup.
func setStderrIsTTY(t *testing.T, v bool) {
	t.Helper()
	addStderrIsTTY = func() bool { return v }
	t.Cleanup(func() { addStderrIsTTY = defaultStderrIsTTY })
}

// seedEntries inserts n rows directly through the storage layer (no CLI, so
// no milestone fires during setup) and returns their IDs in order. project
// may be "".
func seedEntries(t *testing.T, dbPath string, n int, project string) []int64 {
	t.Helper()
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer s.Close()
	ids := make([]int64, 0, n)
	for range n {
		e, err := s.Add(storage.Entry{Title: "seed", Project: project})
		if err != nil {
			t.Fatalf("seed Add: %v", err)
		}
		ids = append(ids, e.ID)
	}
	return ids
}

// TestAddMilestone_FiresOnTTY: 9 prior entries, add the 10th on a TTY → the
// total line on stderr; stdout is still just the new ID.
func TestAddMilestone_FiresOnTTY(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	seedEntries(t, dbPath, 9, "") // 9 rows, no project
	setStderrIsTTY(t, true)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "tenth"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errBuf.String(), "🎉 10 brags and counting") {
		t.Errorf("expected total milestone on stderr, got %q", errBuf.String())
	}
	if _, err := strconv.ParseInt(strings.TrimSpace(outBuf.String()), 10, 64); err != nil {
		t.Errorf("stdout should be the ID alone, got %q", outBuf.String())
	}
}

// TestAddMilestone_SilentUnderJSON (the §9 split-buffer core + NOT-contains
// enforcement): a milestone WOULD cross (10th entry) but --json is silent
// even with TTY forced ON.
func TestAddMilestone_SilentUnderJSON(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	seedEntries(t, dbPath, 9, "")
	setStderrIsTTY(t, true)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetIn(strings.NewReader(`{"title":"tenth"}`))
	root.SetArgs([]string{"--db", dbPath, "add", "--json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr empty under --json, got %q", errBuf.String())
	}
	if _, err := strconv.ParseInt(strings.TrimSpace(outBuf.String()), 10, 64); err != nil {
		t.Errorf("stdout should be the ID alone, got %q", outBuf.String())
	}
}

// TestAddMilestone_SilentWhenNotTTY: same crossing, TTY off → nothing.
func TestAddMilestone_SilentWhenNotTTY(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	seedEntries(t, dbPath, 9, "")
	setStderrIsTTY(t, false) // explicit; also the harness default
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "tenth"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected stderr empty when not a TTY, got %q", errBuf.String())
	}
}

// TestAddMilestone_EditorModeFires: editor path also emits (the other human
// path). 9 prior entries; editor writes a valid Title; TTY on.
func TestAddMilestone_EditorModeFires(t *testing.T) {
	installAddEditFunc(t, func(path string) error {
		return os.WriteFile(path, []byte("Title: tenth\n\n"), 0o600)
	})
	root, dbPath := newRootWithAdd(t)
	seedEntries(t, dbPath, 9, "")
	setStderrIsTTY(t, true)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add"}) // no field flags → editor mode
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errBuf.String(), "🎉 10 brags and counting") {
		t.Errorf("expected total milestone via editor mode, got %q", errBuf.String())
	}
}

// TestAddMilestone_PerProjectFires: 9 prior entries in project "platform",
// add a 10th in "platform" → per-project line. Pad with 13 no-project rows so
// the post-add TOTAL is 23 (13+9+1), NOT a total-threshold — isolating the
// per-project tier as the line we observe (10 and 50 are themselves total
// thresholds, so the padding must keep total off the set).
func TestAddMilestone_PerProjectFires(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	seedEntries(t, dbPath, 13, "")        // 13 no-project rows (total padding)
	seedEntries(t, dbPath, 9, "platform") // 9 in "platform"; total 22, → 23 after add
	setStderrIsTTY(t, true)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "tenth-plat", "--project", "platform"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errBuf.String(), `🎯 10 brags on "platform"`) {
		t.Errorf("expected per-project milestone, got %q", errBuf.String())
	}
}

// TestAddMilestone_StreakEndToEnd proves the CLI reads the SPEC-038 corrected
// streak: 6 entries on the 6 LOCAL days ending YESTERDAY (streak
// alive-through-yesterday = 6), add today → crosses 7. Uses the DEFAULT
// addClock (real now) because Store.Add server-stamps the triggering entry's
// created_at to real now and the milestone clock must agree; Backdate seeds
// the six priors relative to that same now. No time.Sleep.
func TestAddMilestone_StreakEndToEnd(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	now := time.Now()
	ids := seedEntries(t, dbPath, 6, "")
	for i, id := range ids { // i=0..5 → local days -6..-1 at local noon
		d := now.AddDate(0, 0, -6+i)
		at := time.Date(d.Year(), d.Month(), d.Day(), 12, 0, 0, 0, now.Location())
		if err := storagetest.Backdate(dbPath, id, at); err != nil {
			t.Fatalf("backdate id=%d: %v", id, err)
		}
	}
	setStderrIsTTY(t, true)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "today"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errBuf.String(), "🔥 7-day streak!") {
		t.Errorf("expected 7-day streak milestone, got %q", errBuf.String())
	}
}

// TestAddMilestone_OrdinaryAddSilentOnTTY: 2 prior entries today, add a 3rd on
// a TTY → nothing (total 3 not a threshold; not first today/week; streak
// steady). Proves we don't celebrate every add.
func TestAddMilestone_OrdinaryAddSilentOnTTY(t *testing.T) {
	root, dbPath := newRootWithAdd(t)
	seedEntries(t, dbPath, 2, "") // both stamped ~now (today)
	setStderrIsTTY(t, true)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"--db", dbPath, "add", "--title", "third"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected no milestone for an ordinary add, got %q", errBuf.String())
	}
}

package storage

import (
	"database/sql"
	"errors"
	"testing"

	_ "modernc.org/sqlite"
)

func TestTagCounts_SortedAcrossEntries(t *testing.T) {
	s, _ := newTestStore(t)
	if _, err := s.Add(Entry{Title: "e1", Tags: "auth,perf"}); err != nil {
		t.Fatalf("Add e1: %v", err)
	}
	if _, err := s.Add(Entry{Title: "e2", Tags: "perf"}); err != nil {
		t.Fatalf("Add e2: %v", err)
	}
	if _, err := s.Add(Entry{Title: "e3", Tags: "auth,backend"}); err != nil {
		t.Fatalf("Add e3: %v", err)
	}

	got, err := s.TagCounts()
	if err != nil {
		t.Fatalf("TagCounts: %v", err)
	}
	want := []TagCount{
		{Name: "auth", Count: 2},
		{Name: "perf", Count: 2},
		{Name: "backend", Count: 1},
	}
	if len(got) != len(want) {
		t.Fatalf("TagCounts: got %d rows, want %d: %v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].Name != w.Name || got[i].Count != w.Count {
			t.Errorf("TagCounts[%d]: got {%q,%d}, want {%q,%d}",
				i, got[i].Name, got[i].Count, w.Name, w.Count)
		}
	}
}

func TestTagCounts_ExcludesOrphans(t *testing.T) {
	s, path := newTestStore(t)
	e, err := s.Add(Entry{Title: "entry", Tags: "a,b"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	// Drop tag "a" by updating to only "b" — a becomes an orphan.
	if _, err := s.Update(e.ID, Entry{Title: "entry", Tags: "b"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := s.TagCounts()
	if err != nil {
		t.Fatalf("TagCounts: %v", err)
	}
	if len(got) != 1 || got[0].Name != "b" {
		t.Fatalf("TagCounts: expected [{b,1}], got %v", got)
	}

	// Orphan tag row "a" still persists in the tags table.
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("raw open: %v", err)
	}
	defer db.Close()
	var cnt int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tags WHERE name='a'`).Scan(&cnt); err != nil {
		t.Fatalf("raw query: %v", err)
	}
	if cnt != 1 {
		t.Errorf("expected orphan tag 'a' to persist, got count %d", cnt)
	}
}

func TestTagCounts_EmptyCorpus(t *testing.T) {
	s, _ := newTestStore(t)
	got, err := s.TagCounts()
	if err != nil {
		t.Fatalf("TagCounts: %v", err)
	}
	if got == nil {
		t.Fatal("TagCounts: expected non-nil empty slice, got nil")
	}
	if len(got) != 0 {
		t.Fatalf("TagCounts: expected empty slice, got %v", got)
	}
}

func TestRenameTag_GlobalAndFTSReSync(t *testing.T) {
	s, _ := newTestStore(t)
	// Title-free so only the tag column can match in FTS.
	e1, err := s.Add(Entry{Tags: "auth"})
	if err != nil {
		t.Fatalf("Add e1: %v", err)
	}
	e2, err := s.Add(Entry{Tags: "auth"})
	if err != nil {
		t.Fatalf("Add e2: %v", err)
	}
	if _, err := s.Add(Entry{Tags: "perf"}); err != nil {
		t.Fatalf("Add e3: %v", err)
	}

	if err := s.RenameTag("auth", "authz"); err != nil {
		t.Fatalf("RenameTag: %v", err)
	}

	// Each formerly-auth entry now reads authz.
	for _, id := range []int64{e1.ID, e2.ID} {
		got, err := s.Get(id)
		if err != nil {
			t.Fatalf("Get(%d): %v", id, err)
		}
		if got.Tags != "authz" {
			t.Errorf("Get(%d).Tags = %q, want %q", id, got.Tags, "authz")
		}
	}

	// FTS: authz finds both; auth finds none.
	authzHits, err := s.Search("authz", 0)
	if err != nil {
		t.Fatalf("Search(authz): %v", err)
	}
	if len(authzHits) != 2 {
		t.Errorf("Search(authz): got %d, want 2", len(authzHits))
	}
	authHits, err := s.Search("auth", 0)
	if err != nil {
		t.Fatalf("Search(auth): %v", err)
	}
	if len(authHits) != 0 {
		t.Errorf("Search(auth): got %d, want 0", len(authHits))
	}
}

func TestRenameTag_IntoExistingErrors(t *testing.T) {
	s, path := newTestStore(t)
	if _, err := s.Add(Entry{Title: "e1", Tags: "auth"}); err != nil {
		t.Fatalf("Add e1: %v", err)
	}
	if _, err := s.Add(Entry{Title: "e2", Tags: "perf"}); err != nil {
		t.Fatalf("Add e2: %v", err)
	}

	err := s.RenameTag("auth", "perf")
	if !errors.Is(err, ErrTagExists) {
		t.Fatalf("RenameTag into existing: want ErrTagExists, got %v", err)
	}

	// Both tags and all memberships are intact.
	db, err2 := sql.Open("sqlite", path)
	if err2 != nil {
		t.Fatalf("raw open: %v", err2)
	}
	defer db.Close()
	for _, name := range []string{"auth", "perf"} {
		var cnt int
		if err3 := db.QueryRow(`SELECT COUNT(*) FROM tags WHERE name=?`, name).Scan(&cnt); err3 != nil {
			t.Fatalf("count tag %q: %v", name, err3)
		}
		if cnt != 1 {
			t.Errorf("expected tag %q to still exist, count=%d", name, cnt)
		}
	}
}

func TestRenameTag_MissingOldErrors(t *testing.T) {
	s, _ := newTestStore(t)
	err := s.RenameTag("nope", "x")
	if !errors.Is(err, ErrTagNotFound) {
		t.Fatalf("RenameTag missing old: want ErrTagNotFound, got %v", err)
	}
}

func TestMergeTags_FoldsDeDupsAndDropsSrc(t *testing.T) {
	s, path := newTestStore(t)
	// e1: auth + perf (tagged both src and dst)
	e1, err := s.Add(Entry{Tags: "auth,perf"})
	if err != nil {
		t.Fatalf("Add e1: %v", err)
	}
	// e2: perf only
	if _, err := s.Add(Entry{Tags: "perf"}); err != nil {
		t.Fatalf("Add e2: %v", err)
	}
	// e3: auth only
	e3, err := s.Add(Entry{Tags: "auth"})
	if err != nil {
		t.Fatalf("Add e3: %v", err)
	}

	if err := s.MergeTags("auth", "perf"); err != nil {
		t.Fatalf("MergeTags: %v", err)
	}

	// (a) TagCounts: only perf with count 3.
	counts, err := s.TagCounts()
	if err != nil {
		t.Fatalf("TagCounts: %v", err)
	}
	if len(counts) != 1 || counts[0].Name != "perf" || counts[0].Count != 3 {
		t.Errorf("TagCounts after merge: got %v, want [{perf 3}]", counts)
	}

	// (b) e1 de-duped: one perf, no auth.
	got1, err := s.Get(e1.ID)
	if err != nil {
		t.Fatalf("Get(e1): %v", err)
	}
	if got1.Tags != "perf" {
		t.Errorf("e1.Tags after merge: got %q, want %q", got1.Tags, "perf")
	}

	// (c) raw: exactly one perf tagging for e1.
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("raw open: %v", err)
	}
	defer db.Close()
	var rawCnt int
	err = db.QueryRow(
		`SELECT COUNT(*) FROM taggings tg JOIN tags t ON t.id=tg.tag_id
		  WHERE tg.taggable_id=? AND t.name='perf'`, e1.ID,
	).Scan(&rawCnt)
	if err != nil {
		t.Fatalf("raw tagging count: %v", err)
	}
	if rawCnt != 1 {
		t.Errorf("e1 perf tagging count: got %d, want 1", rawCnt)
	}

	// (d) auth tag row is gone.
	var authCnt int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tags WHERE name='auth'`).Scan(&authCnt); err != nil {
		t.Fatalf("auth tag count: %v", err)
	}
	if authCnt != 0 {
		t.Errorf("auth tag should be deleted, count=%d", authCnt)
	}

	// (e) FTS: Search("perf") → 3, Search("auth") → 0.
	perfHits, err := s.Search("perf", 0)
	if err != nil {
		t.Fatalf("Search(perf): %v", err)
	}
	if len(perfHits) != 3 {
		t.Errorf("Search(perf): got %d, want 3", len(perfHits))
	}
	authHits, err := s.Search("auth", 0)
	if err != nil {
		t.Fatalf("Search(auth): %v", err)
	}
	if len(authHits) != 0 {
		t.Errorf("Search(auth): got %d, want 0 (e3 was auth only, now perf)", len(authHits))
	}
	_ = e3
}

func TestMergeTags_MissingErrors(t *testing.T) {
	s, _ := newTestStore(t)
	if _, err := s.Add(Entry{Title: "e", Tags: "perf"}); err != nil {
		t.Fatalf("Add: %v", err)
	}

	// Missing src.
	err := s.MergeTags("nope", "perf")
	if !errors.Is(err, ErrTagNotFound) {
		t.Fatalf("MergeTags missing src: want ErrTagNotFound, got %v", err)
	}

	// Missing dst.
	err = s.MergeTags("perf", "nope")
	if !errors.Is(err, ErrTagNotFound) {
		t.Fatalf("MergeTags missing dst: want ErrTagNotFound, got %v", err)
	}
}

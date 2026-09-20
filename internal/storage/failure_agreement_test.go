package storage_test

import (
	"path/filepath"
	"testing"

	"github.com/jysf/bragfile000/internal/aggregate"
	"github.com/jysf/bragfile000/internal/storage"

	_ "modernc.org/sqlite"
)

// TestFailureClassifier_GoPredicateMatchesTypeFilter ▲ SPEC-086 — the drift
// guard for DEC-050's failure predicate, the same shape as
// TestProvenanceClassifier_GoPredicateMatchesSQLClause in this package.
//
// A failure has two definitions in two packages. DEC-049's retrieval path,
// `brag list --type failed`, is storage's SQL `e.type = ?`. The digests'
// "What didn't work" section is aggregate.IsFailure in Go. Nothing in storage
// knows `failed` is special, and that is deliberate (DEC-049 part 3) — but the
// generic filter is still a definition, and the two must select the same rows,
// or `brag impact` would list a failure that `brag list --type failed` cannot
// find. The seeds cover the spellings a well-meaning change would widen the Go
// side to accept (case, whitespace, the near-synonym), each of which the SQL
// side rejects on a BINARY-collated column.
func TestFailureClassifier_GoPredicateMatchesTypeFilter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.sqlite")
	s, err := storage.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	for _, e := range []storage.Entry{
		{Title: "failure-with-impact", Type: "failed", Impact: "cost two days"},
		{Title: "failure-no-impact", Type: "failed"},
		{Title: "capitalised", Type: "Failed"},
		{Title: "leading-space", Type: " failed"},
		{Title: "trailing-space", Type: "failed "},
		{Title: "near-synonym", Type: "failure"},
		{Title: "a-lesson", Type: "learned"},
		{Title: "untyped", Type: ""},
	} {
		if _, err := s.Add(e); err != nil {
			t.Fatalf("add %q: %v", e.Title, err)
		}
	}

	// SQL side: the sanctioned retrieval path.
	sqlFailed, err := s.List(storage.ListFilter{Type: aggregate.FailureType})
	if err != nil {
		t.Fatalf("List(Type=%q): %v", aggregate.FailureType, err)
	}

	// Go side: one unfiltered read, partitioned the way the digests do it.
	all, err := s.List(storage.ListFilter{})
	if err != nil {
		t.Fatalf("List(all): %v", err)
	}
	_, goFailed := aggregate.SplitFailures(all)

	if !sameIDSet(idSet(sqlFailed), idSet(goFailed)) {
		t.Errorf("failure sets differ: SQL=%d Go=%d", len(sqlFailed), len(goFailed))
	}
	// Anchor the count so a change that keeps both sides equal-but-wrong still
	// trips: exactly the two rows typed "failed", nothing else.
	if len(goFailed) != 2 {
		t.Errorf("expected 2 failures; got %d", len(goFailed))
	}
}

package storage_test

import (
	"path/filepath"
	"testing"

	"github.com/jysf/bragfile000/internal/aggregate"
	"github.com/jysf/bragfile000/internal/storage"

	_ "modernc.org/sqlite"
)

// This is an EXTERNAL test package (storage_test) on purpose: it imports
// internal/aggregate, which imports internal/storage. An in-package
// (package storage) test file importing aggregate would form an import
// cycle, so the drift-guard lives here and uses only the exported API.

// seedTag adds an entry with the given title + tags via the exported Store API.
func seedTag(t *testing.T, s *storage.Store, title, tags string) {
	t.Helper()
	if _, err := s.Add(storage.Entry{Title: title, Tags: tags}); err != nil {
		t.Fatalf("add %q: %v", title, err)
	}
}

func idSet(entries []storage.Entry) map[int64]struct{} {
	m := make(map[int64]struct{}, len(entries))
	for _, e := range entries {
		m[e.ID] = struct{}{}
	}
	return m
}

func sameIDSet(a, b map[int64]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for id := range a {
		if _, ok := b[id]; !ok {
			return false
		}
	}
	return true
}

// TestProvenanceClassifier_GoPredicateMatchesSQLClause ▲ SPEC-045 — the
// cross-package agreement test that closes the SPEC-043 drift-coupling WATCH.
// It seeds a corpus covering every classification edge, then runs BOTH
// classifiers over the SAME rows and asserts identical partitions:
//   - SQL side:  Store.List(ListFilter{Author:"agent"}) → provenanceExistsClause
//   - Go side:   Store.List(ListFilter{}) partitioned by aggregate.IsAgentAuthored
//
// If either classifier ever changes without the other, this fails — the
// single-source guarantee (SPEC-045 AC 3 / LD2).
func TestProvenanceClassifier_GoPredicateMatchesSQLClause(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.sqlite")
	s, err := storage.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	// Cover every edge: agent-only, model-only, both, plain-human, no-tags,
	// the agentic/modeling false-positive, and an agent tag mid-list.
	seedTag(t, s, "agent-only", "agent:claude-code")
	seedTag(t, s, "model-only", "model:claude-opus-4-8")
	seedTag(t, s, "both", "perf,agent:claude-code,model:claude-opus-4-8")
	seedTag(t, s, "plain-human", "perf,api")
	seedTag(t, s, "no-tags", "")
	seedTag(t, s, "false-positive", "agentic,modeling") // no colon → human
	seedTag(t, s, "agent-mid-list", "a,agent:claude-code,b")

	// SQL side.
	sqlAgent, err := s.List(storage.ListFilter{Author: "agent"})
	if err != nil {
		t.Fatalf("List(Author=agent): %v", err)
	}
	sqlHuman, err := s.List(storage.ListFilter{Author: "human"})
	if err != nil {
		t.Fatalf("List(Author=human): %v", err)
	}

	// Go side: one query, partition in Go.
	all, err := s.List(storage.ListFilter{})
	if err != nil {
		t.Fatalf("List(all): %v", err)
	}
	var goAgent, goHuman []storage.Entry
	for _, e := range all {
		if aggregate.IsAgentAuthored(e) {
			goAgent = append(goAgent, e)
		} else {
			goHuman = append(goHuman, e)
		}
	}

	if !sameIDSet(idSet(sqlAgent), idSet(goAgent)) {
		t.Errorf("agent sets differ: SQL=%d Go=%d", len(sqlAgent), len(goAgent))
	}
	if !sameIDSet(idSet(sqlHuman), idSet(goHuman)) {
		t.Errorf("human sets differ: SQL=%d Go=%d", len(sqlHuman), len(goHuman))
	}
	// Anchor the counts so a silent classifier change that keeps both sides
	// equal-but-wrong still trips: 4 agent, 3 human.
	if len(goAgent) != 4 {
		t.Errorf("expected 4 agent-authored; got %d", len(goAgent))
	}
	if len(goHuman) != 3 {
		t.Errorf("expected 3 human; got %d", len(goHuman))
	}
}

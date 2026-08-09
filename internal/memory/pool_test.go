package memory

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jysf/bragfile000/internal/storage"
)

// fakeSource records every read Gather makes, so a test can assert on the
// COMPOSITION (which reads ran, with what filters) without a database. Keeping
// these tests DB-free is deliberate: internal/memory's other tests are pure,
// and Gather is the package's only I/O — it should not drag a store into the
// package's test dependencies (SPEC-074 LD9).
type fakeSource struct {
	listCalls   []storage.ListFilter
	searchCalls []string
	searchLimit []int
	listRet     [][]storage.Entry
	searchRet   []storage.Entry
	listErr     error
	searchErr   error
}

func (f *fakeSource) List(filter storage.ListFilter) ([]storage.Entry, error) {
	f.listCalls = append(f.listCalls, filter)
	if f.listErr != nil {
		return nil, f.listErr
	}
	if len(f.listRet) == 0 {
		return nil, nil
	}
	out := f.listRet[0]
	if len(f.listRet) > 1 {
		f.listRet = f.listRet[1:]
	}
	return out, nil
}

func (f *fakeSource) Search(q string, limit int) ([]storage.Entry, error) {
	f.searchCalls = append(f.searchCalls, q)
	f.searchLimit = append(f.searchLimit, limit)
	if f.searchErr != nil {
		return nil, f.searchErr
	}
	return f.searchRet, nil
}

func entries(ids ...int64) []storage.Entry {
	out := make([]storage.Entry, len(ids))
	for i, id := range ids {
		out[i] = storage.Entry{ID: id}
	}
	return out
}

// No query and no project: exactly ONE read, and no relevance ranking.
func TestGather_RecencyOnlyRunsOneRead(t *testing.T) {
	src := &fakeSource{listRet: [][]storage.Entry{entries(3, 2, 1)}}
	pool, err := Gather(src, GatherOptions{})
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	if len(src.listCalls) != 1 || len(src.searchCalls) != 0 {
		t.Errorf("reads = %d list / %d search, want 1 / 0", len(src.listCalls), len(src.searchCalls))
	}
	if pool.Matched != nil {
		t.Errorf("Matched = %v, want nil (no query → no relevance list)", pool.Matched)
	}
	if len(pool.Entries) != 3 {
		t.Errorf("Entries = %d, want 3", len(pool.Entries))
	}
}

// A query adds the relevance read, and Matched carries the bm25 ORDER — not a
// set. This is the property DEC-043 sub-decision 1 depends on.
func TestGather_QueryAddsMatchedInSearchOrder(t *testing.T) {
	src := &fakeSource{
		listRet:   [][]storage.Entry{entries(9, 8)},
		searchRet: entries(5, 1, 2, 4),
	}
	pool, err := Gather(src, GatherOptions{Query: "auth"})
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	want := []int64{5, 1, 2, 4}
	if len(pool.Matched) != len(want) {
		t.Fatalf("Matched = %v, want %v", pool.Matched, want)
	}
	for i := range want {
		if pool.Matched[i] != want[i] {
			t.Fatalf("Matched = %v, want %v (order is the ranking)", pool.Matched, want)
		}
	}
	if got := src.searchCalls[0]; got != `"auth"` {
		t.Errorf("search query = %q, want %q (DEC-010 transform applied)", got, `"auth"`)
	}
}

// A project adds the third read, scoped and capped.
func TestGather_ProjectAddsScopedRead(t *testing.T) {
	src := &fakeSource{listRet: [][]storage.Entry{entries(3), entries(7)}}
	if _, err := Gather(src, GatherOptions{Project: "orbit"}); err != nil {
		t.Fatalf("Gather: %v", err)
	}
	if len(src.listCalls) != 2 {
		t.Fatalf("list reads = %d, want 2", len(src.listCalls))
	}
	if src.listCalls[1].Project != "orbit" {
		t.Errorf("second read project = %q, want orbit", src.listCalls[1].Project)
	}
}

// Every read is capped at PoolLimit (DEC-043 sub-decision 5). An uncapped read
// would make the pool unbounded on a large corpus.
func TestGather_AllReadsCappedAtPoolLimit(t *testing.T) {
	src := &fakeSource{listRet: [][]storage.Entry{entries(1), entries(2)}, searchRet: entries(3)}
	if _, err := Gather(src, GatherOptions{Query: "auth", Project: "orbit"}); err != nil {
		t.Fatalf("Gather: %v", err)
	}
	for i, c := range src.listCalls {
		if c.Limit != PoolLimit {
			t.Errorf("list read %d limit = %d, want %d", i, c.Limit, PoolLimit)
		}
	}
	if src.searchLimit[0] != PoolLimit {
		t.Errorf("search limit = %d, want %d", src.searchLimit[0], PoolLimit)
	}
}

// A malformed query is classified as *ErrQuery so callers can surface it as a
// USER error. Without the distinct type both surfaces would have to
// string-match, or would report a caller mistake as an internal failure.
func TestGather_MalformedQueryIsErrQuery(t *testing.T) {
	src := &fakeSource{listRet: [][]storage.Entry{entries(1)}}
	_, err := Gather(src, GatherOptions{Query: `bad"quote`})
	if err == nil {
		t.Fatal("expected an error")
	}
	var qe *ErrQuery
	if !errors.As(err, &qe) {
		t.Fatalf("error %T (%v) is not *ErrQuery", err, err)
	}
}

// An infrastructure failure must NOT be classified as a user error.
func TestGather_ReadFailureIsNotErrQuery(t *testing.T) {
	for _, tc := range []struct {
		name string
		src  *fakeSource
		opts GatherOptions
	}{
		{"list", &fakeSource{listErr: fmt.Errorf("disk gone")}, GatherOptions{}},
		{"search", &fakeSource{listRet: [][]storage.Entry{entries(1)}, searchErr: fmt.Errorf("fts broken")}, GatherOptions{Query: "auth"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Gather(tc.src, tc.opts)
			if err == nil {
				t.Fatal("expected an error")
			}
			var qe *ErrQuery
			if errors.As(err, &qe) {
				t.Errorf("%v classified as a user error; it is infrastructure", err)
			}
		})
	}
}

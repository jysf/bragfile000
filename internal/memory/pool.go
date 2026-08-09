package memory

import (
	"fmt"

	"github.com/jysf/bragfile000/internal/ftsquery"
	"github.com/jysf/bragfile000/internal/storage"
)

// This file is the ONE exception to internal/memory's purity: Gather performs
// reads. It is here rather than in a caller because DEC-043 sub-decision 5's
// "one bounded read per ranked list" composition has to be identical on every
// surface — the CLI and, from SPEC-074, the MCP resources and the brag_memory
// tool. The parity property those surfaces claim ("the resource IS the
// markdown") is only true if one composition runs on both sides, not two that
// happen to agree today (SPEC-074 LD9 / DEC-045).
//
// Slice itself stays pure and clock-free; both mechanical guards
// (TestPackageReadsNoWallClock, TestPackageEmitsNoReservedTagNamespace) still
// hold, because reading is not the same as reading a clock and this file
// writes nothing.

// Source is the read surface Gather needs. *storage.Store satisfies it; a test
// can substitute a fake without a database.
type Source interface {
	List(f storage.ListFilter) ([]storage.Entry, error)
	Search(query string, limit int) ([]storage.Entry, error)
}

// GatherOptions selects which of the three reads run.
type GatherOptions struct {
	Query   string // raw user query; "" skips the relevance read
	Project string // "" skips the project read
}

// Pool is Gather's result: the union of the reads, plus the bm25 id order that
// becomes Options.Matched. The order of Matched is load-bearing — it is the
// relevance ranking itself, not a set (DEC-043 sub-decision 1).
type Pool struct {
	Entries []storage.Entry
	Matched []int64
}

// ErrQuery marks a malformed query so callers can surface it as a USER error
// (a UserError on the CLI, a tool error over MCP) rather than an internal one.
// Everything else Gather returns is an infrastructure failure.
type ErrQuery struct{ Err error }

func (e *ErrQuery) Error() string { return e.Err.Error() }
func (e *ErrQuery) Unwrap() error { return e.Err }

// Gather performs DEC-043 sub-decision 5's one-bounded-read-per-list
// composition and returns the union pool plus the bm25 id order.
//
// Each read is capped at PoolLimit. Dropping the project read makes old
// same-project history unreachable at any rank; dropping the match read makes
// an old top-relevance entry unreachable. Slice dedupes the union by id, so
// the overlap between reads costs nothing but rows.
func Gather(src Source, opts GatherOptions) (Pool, error) {
	entries, err := src.List(storage.ListFilter{Limit: PoolLimit})
	if err != nil {
		return Pool{}, fmt.Errorf("list entries: %w", err)
	}

	var matched []int64
	if opts.Query != "" {
		fts5, qerr := ftsquery.Build(opts.Query)
		if qerr != nil {
			return Pool{}, &ErrQuery{Err: qerr}
		}
		hits, serr := src.Search(fts5, PoolLimit)
		if serr != nil {
			return Pool{}, fmt.Errorf("search entries: %w", serr)
		}
		for _, e := range hits {
			matched = append(matched, e.ID)
		}
		entries = append(entries, hits...)
	}

	if opts.Project != "" {
		scoped, perr := src.List(storage.ListFilter{Project: opts.Project, Limit: PoolLimit})
		if perr != nil {
			return Pool{}, fmt.Errorf("list entries: %w", perr)
		}
		entries = append(entries, scoped...)
	}

	return Pool{Entries: entries, Matched: matched}, nil
}

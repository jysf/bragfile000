package storage

import "time"

// Entry is one captured brag-worthy moment.
//
// Tags is stored as a comma-joined string for MVP (DEC-004). The
// caller's input is persisted verbatim; splitting/normalization is the
// caller's problem.
type Entry struct {
	ID          int64
	Title       string
	Description string
	Tags        string
	Project     string
	Type        string
	Impact      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ListFilter is the filter passed to Store.List. A zero value means
// "no filter": every field's zero value contributes no WHERE clause.
// Populated fields combine via AND.
type ListFilter struct {
	Tag     string    // sentinel-comma token match against comma-joined tags
	Project string    // exact equality on entries.project
	Type    string    // exact equality on entries.type
	Since   time.Time // entries.created_at >= Since (RFC3339 UTC)
	Limit   int       // LIMIT N; 0 = no limit
}

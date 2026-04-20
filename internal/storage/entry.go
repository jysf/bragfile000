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

// ListFilter is the filter passed to Store.List. Empty for MVP —
// filter fields (tag, project, since, …) arrive in STAGE-002.
type ListFilter struct{}

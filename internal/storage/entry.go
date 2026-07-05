package storage

import "time"

// Entry is one captured brag-worthy moment.
//
// Tags is the comma-joined, whitespace-trimmed, deduplicated tag
// projection reconstructed from the tags/taggings join (DEC-015).
// Store.Add and Store.Update canonicalize raw input; Store.Get, List,
// and Search reconstruct this value from the normalized join.
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
	Tag     string    // exact tag-name membership via normalized join (DEC-015)
	Project string    // exact equality on entries.project
	Type    string    // exact equality on entries.type
	Since   time.Time // entries.created_at >= Since (RFC3339 UTC)
	Limit   int       // LIMIT N; 0 = no limit
	Author  string    // "agent" | "human" | "" (all); classifies by presence of a reserved agent:/model: provenance tag (DEC-024). Invalid values are an error.
}

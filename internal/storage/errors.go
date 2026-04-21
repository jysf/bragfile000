package storage

import "errors"

// ErrNotFound is returned (wrapped) by Store.Get when no row matches
// the requested ID. Callers use errors.Is(err, ErrNotFound) to map to
// a user-facing "no such entry" error.
var ErrNotFound = errors.New("entry not found")

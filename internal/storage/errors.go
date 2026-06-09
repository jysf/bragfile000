package storage

import "errors"

// ErrNotFound is returned (wrapped) by Store.Get when no row matches
// the requested ID. Callers use errors.Is(err, ErrNotFound) to map to
// a user-facing "no such entry" error.
var ErrNotFound = errors.New("entry not found")

// ErrTagNotFound is returned (wrapped) when a named tag does not exist
// (RenameTag old name, MergeTags src/dst). Callers map it to a
// user-facing "no tag named X" error.
var ErrTagNotFound = errors.New("tag not found")

// ErrTagExists is returned (wrapped) by RenameTag when the target name
// already names a tag. Callers map it to "use merge".
var ErrTagExists = errors.New("tag already exists")

// ErrProjectExists is returned (wrapped) by CreateProject when name is taken.
var ErrProjectExists = errors.New("project already exists")

// ErrLocationExists is returned (wrapped) by AddLocation when the path is
// already attached to some project (paths are globally unique).
var ErrLocationExists = errors.New("location already exists")

// Package editor renders and parses the header+body buffer format
// used by `brag edit` (and, eventually, `brag add` no-args). The
// format is pinned by DEC-009: RFC822-style headers parseable via
// net/textproto.Reader.ReadMIMEHeader, followed by a blank line,
// followed by a free-form markdown body (the entry's Description).
//
// editor does NOT import internal/storage: Fields mirrors the
// user-editable subset of storage.Entry, and the CLI layer translates
// between the two. This keeps the format concern independent and
// avoids a dependency cycle risk.
package editor

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"strings"
)

// Fields is the user-editable subset of a brag entry. ID, CreatedAt,
// and UpdatedAt are intentionally absent — they're managed by the
// storage layer and never round-tripped through the editor buffer.
type Fields struct {
	Title       string
	Description string
	Tags        string
	Project     string
	Type        string
	Impact      string
}

// Render produces the editable buffer for f. Headers are emitted in a
// fixed canonical order (Title, Tags, Project, Type, Impact); empty-
// valued fields are omitted entirely. Description is written verbatim
// after a blank-line separator, with a trailing newline appended when
// not already present so the file plays nicely with editors that
// complain about missing final newlines.
func Render(f Fields) []byte {
	var b bytes.Buffer
	write := func(k, v string) {
		if v != "" {
			fmt.Fprintf(&b, "%s: %s\n", k, v)
		}
	}
	write("Title", f.Title)
	write("Tags", f.Tags)
	write("Project", f.Project)
	write("Type", f.Type)
	write("Impact", f.Impact)
	b.WriteByte('\n')
	b.WriteString(f.Description)
	if f.Description != "" && !strings.HasSuffix(f.Description, "\n") {
		b.WriteByte('\n')
	}
	return b.Bytes()
}

// EmptyTemplate returns a buffer shaped per DEC-009 with all five
// editable headers pre-listed but empty. Distinct from Render(Fields{}),
// which omits empty-valued headers — the template is a UX hint shown
// to users in editor-launch mode (e.g. `brag add` with no flags), not
// a renderer output.
func EmptyTemplate() []byte {
	return []byte("Title: \nTags: \nProject: \nType: \nImpact: \n\n")
}

// Parse reads the header block and body out of buf and returns the
// populated Fields. Header keys are case-insensitive (canonicalized by
// net/textproto). Unknown headers are silently ignored. A missing or
// whitespace-only Title returns an error mentioning "title".
func Parse(buf []byte) (Fields, error) {
	tp := textproto.NewReader(bufio.NewReader(bytes.NewReader(buf)))
	hdr, err := tp.ReadMIMEHeader()
	if err != nil && err != io.EOF {
		return Fields{}, fmt.Errorf("parse buffer headers: %w", err)
	}
	f := Fields{
		Title:   strings.TrimSpace(hdr.Get("Title")),
		Tags:    strings.TrimSpace(hdr.Get("Tags")),
		Project: strings.TrimSpace(hdr.Get("Project")),
		Type:    strings.TrimSpace(hdr.Get("Type")),
		Impact:  strings.TrimSpace(hdr.Get("Impact")),
	}
	if f.Title == "" {
		return Fields{}, fmt.Errorf("parse buffer: Title header is required and must be non-empty")
	}
	remaining, err := io.ReadAll(tp.R)
	if err != nil {
		return Fields{}, fmt.Errorf("parse buffer body: %w", err)
	}
	f.Description = string(remaining)
	return f, nil
}

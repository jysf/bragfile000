package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// tagsField preserves DEC-004 at the ingress boundary: tags must be a
// comma-joined string, not a JSON array. A dedicated type lets us
// surface a clear error naming DEC-004's model rather than leak the
// stdlib decoder's "cannot unmarshal array into Go struct field"
// jargon.
type tagsField string

func (t *tagsField) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("tags must be a comma-joined string, not an array (per DEC-004)")
	}
	*t = tagsField(s)
	return nil
}

// addJSONInput is the accepted stdin shape for `brag add --json`.
// DEC-012: required `title`; optional user-owned text fields;
// server-owned fields (id, created_at, updated_at) tolerated and
// ignored via explicit fields with json.RawMessage so the decoder
// sees them as known-and-accepted but their values are never stored.
type addJSONInput struct {
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Tags        tagsField       `json:"tags,omitempty"`
	Project     string          `json:"project,omitempty"`
	Type        string          `json:"type,omitempty"`
	Impact      string          `json:"impact,omitempty"`
	ID          json.RawMessage `json:"id,omitempty"`
	CreatedAt   json.RawMessage `json:"created_at,omitempty"`
	UpdatedAt   json.RawMessage `json:"updated_at,omitempty"`
}

// parseAddJSON reads exactly one JSON object from r and returns the
// hydrated storage.Entry. Server-owned fields on the input are dropped.
// All errors route through UserErrorf so ErrUser propagates.
func parseAddJSON(r io.Reader) (storage.Entry, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var in addJSONInput
	if err := dec.Decode(&in); err != nil {
		return storage.Entry{}, UserErrorf("invalid JSON input: %v", err)
	}
	// Trailing-garbage detection: a second Decode must hit EOF. Anything
	// else (a second object, a stray token, or a parse error) means the
	// stdin contained more than the single JSON value we accept.
	var trailing json.RawMessage
	if err := dec.Decode(&trailing); err != io.EOF {
		if err == nil {
			return storage.Entry{}, UserErrorf("invalid JSON input: expected a single object, got trailing data")
		}
		return storage.Entry{}, UserErrorf("invalid JSON input: %v", err)
	}
	if strings.TrimSpace(in.Title) == "" {
		return storage.Entry{}, UserErrorf(`--json input: "title" is required and must not be empty`)
	}
	if len(in.Title) > 200 {
		return storage.Entry{}, UserErrorf(`--json input: "title" exceeds 200-character limit`)
	}
	if len(in.Description) > 100000 {
		return storage.Entry{}, UserErrorf(`--json input: "description" exceeds 100000-character limit`)
	}
	if len(string(in.Tags)) > 64 {
		return storage.Entry{}, UserErrorf(`--json input: "tags" exceeds 64-character limit`)
	}
	if len(in.Project) > 64 {
		return storage.Entry{}, UserErrorf(`--json input: "project" exceeds 64-character limit`)
	}
	if len(in.Type) > 64 {
		return storage.Entry{}, UserErrorf(`--json input: "type" exceeds 64-character limit`)
	}
	if len(in.Impact) > 256 {
		return storage.Entry{}, UserErrorf(`--json input: "impact" exceeds 256-character limit`)
	}
	return storage.Entry{
		Title:       in.Title,
		Description: in.Description,
		Tags:        string(in.Tags),
		Project:     in.Project,
		Type:        in.Type,
		Impact:      in.Impact,
	}, nil
}

func runAddJSON(cmd *cobra.Command, _ []string) error {
	entry, err := parseAddJSON(cmd.InOrStdin())
	if err != nil {
		return err
	}

	dbFlag := getFlagString(cmd, "db")
	path, err := config.ResolveDBPath(dbFlag)
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}
	s, err := storage.Open(path)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	inserted, err := s.Add(entry)
	if err != nil {
		return fmt.Errorf("add entry: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), inserted.ID)
	return nil
}

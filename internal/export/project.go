package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// projectRecord is the DEC-011-family serialization shape for a project:
// keys in `projects`-column order, locations as a JSON array of strings,
// timestamps pre-formatted RFC3339. Shared by ToProjectsJSON (list) and
// ToProjectJSON (show) so the array elements and the single-show object
// are byte-identical in shape.
type projectRecord struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	StateNote string   `json:"state_note"`
	Locations []string `json:"locations"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

func toProjectRecord(p storage.Project) projectRecord {
	locs := p.Locations
	if locs == nil {
		locs = []string{}
	}
	return projectRecord{
		ID:        p.ID,
		Name:      p.Name,
		Status:    p.Status,
		StateNote: p.StateNote,
		Locations: locs,
		CreatedAt: p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// ToProjectsJSON renders projects as a naked JSON array (DEC-011 shape;
// 2-space indent). Empty/nil input renders "[]", never "null".
func ToProjectsJSON(projects []storage.Project) ([]byte, error) {
	out := make([]projectRecord, 0, len(projects))
	for _, p := range projects {
		out = append(out, toProjectRecord(p))
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal projects json: %w", err)
	}
	return b, nil
}

// projectStatusRecord is the DEC-011-family serialization for a `brag
// project status` dashboard row: the project's identity fields plus
// brag_count. state_note is carried in full (the plain-output truncation
// does not apply to JSON). Locations are omitted (see storage.ProjectStatus).
type projectStatusRecord struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	StateNote string `json:"state_note"`
	BragCount int    `json:"brag_count"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ToProjectStatusesJSON renders dashboard rows as a naked JSON array
// (DEC-011 shape; 2-space indent). Empty/nil input renders "[]", never "null".
func ToProjectStatusesJSON(statuses []storage.ProjectStatus) ([]byte, error) {
	out := make([]projectStatusRecord, 0, len(statuses))
	for _, st := range statuses {
		out = append(out, projectStatusRecord{
			ID:        st.ID,
			Name:      st.Name,
			Status:    st.Status,
			StateNote: st.StateNote,
			BragCount: st.BragCount,
			CreatedAt: st.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: st.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal project statuses json: %w", err)
	}
	return b, nil
}

// ToProjectJSON renders a single project as a JSON object (not an array)
// for `brag project show --format json`. 2-space indent; an empty
// Locations renders "[]", never "null".
func ToProjectJSON(p storage.Project) ([]byte, error) {
	b, err := json.MarshalIndent(toProjectRecord(p), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal project json: %w", err)
	}
	return b, nil
}

// ToProjectsMarkdown renders statuses as the `brag://projects` resource body
// (SPEC-074 LD10): every non-archived project, most-recently-updated first,
// with its status and brag count. Locations and state_note are deliberately
// excluded (DEC-045 sub-decision 3) — this is a name-selection lookup for the
// `brag://memory/project/{name}` template, not a dashboard. Returns bytes
// with the trailing "\n" stripped (matches every other renderer).
func ToProjectsMarkdown(statuses []storage.ProjectStatus) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "# Bragfile Projects")
	fmt.Fprintln(&buf)

	if len(statuses) == 0 {
		fmt.Fprint(&buf, "No projects registered yet. Register one with `brag project ensure <name>`.")
		return trimTrailingNewline(buf.Bytes()), nil
	}

	fmt.Fprintln(&buf, "Registered projects, most-recently-updated first. Use these names verbatim.")
	fmt.Fprintln(&buf)
	for _, st := range statuses {
		fmt.Fprintf(&buf, "- %s — %s, %d brags\n", st.Name, st.Status, st.BragCount)
	}
	return trimTrailingNewline(buf.Bytes()), nil
}

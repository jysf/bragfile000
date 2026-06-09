package export

import (
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

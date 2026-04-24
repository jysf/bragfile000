// Package export renders []storage.Entry into the machine-readable
// output formats used by `brag list --format ...` and `brag export
// --format ...`. The JSON shape is locked by DEC-011; the TSV shape
// mirrors the JSON field order with a header row.
package export

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// entryRecord mirrors storage.Entry with the exact JSON struct tags
// that DEC-011 locks: 9 keys in SQL-column order, tags as string,
// timestamps pre-formatted as RFC3339 strings.
type entryRecord struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	Project     string `json:"project"`
	Type        string `json:"type"`
	Impact      string `json:"impact"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ToJSON marshals entries per DEC-011: naked array, 9 keys in SQL-
// column order, tags comma-joined string, timestamps RFC3339, empty
// fields as "", pretty-printed with 2-space indent. Empty/nil input
// returns exactly "[]" (never "null").
func ToJSON(entries []storage.Entry) ([]byte, error) {
	if len(entries) == 0 {
		return []byte("[]"), nil
	}
	records := make([]entryRecord, 0, len(entries))
	for _, e := range entries {
		records = append(records, entryRecord{
			ID:          e.ID,
			Title:       e.Title,
			Description: e.Description,
			Tags:        e.Tags,
			Project:     e.Project,
			Type:        e.Type,
			Impact:      e.Impact,
			CreatedAt:   e.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:   e.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	return json.MarshalIndent(records, "", "  ")
}

// TSVHeader is the first line of `brag list --format tsv` output: 9
// column names in the same order as DEC-011, separated by 8 tabs. No
// trailing newline — the caller writes it with one.
const TSVHeader = "id\ttitle\tdescription\ttags\tproject\ttype\timpact\tcreated_at\tupdated_at"

// ToTSVRow renders one storage.Entry as a tab-separated data row. Field
// order matches TSVHeader and DEC-011. Embedded tabs in user text are
// NOT escaped — same accepted MVP trade-off as `brag list` plain mode.
func ToTSVRow(e storage.Entry) string {
	return strings.Join([]string{
		strconv.FormatInt(e.ID, 10),
		e.Title,
		e.Description,
		e.Tags,
		e.Project,
		e.Type,
		e.Impact,
		e.CreatedAt.UTC().Format(time.RFC3339),
		e.UpdatedAt.UTC().Format(time.RFC3339),
	}, "\t")
}

package export

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/jysf/bragfile000/internal/storage"
)

// TestToJSON_DEC011ShapeGolden locks all six DEC-011 choices against a
// literal golden string. If any whitespace, key, or value drifts, the
// fix is in internal/export/json.go, not here.
func TestToJSON_DEC011ShapeGolden(t *testing.T) {
	fixture := []storage.Entry{
		{
			ID:          1,
			Title:       "shipped FTS5",
			Description: "migration 0002 landed",
			Tags:        "sqlite,fts5",
			Project:     "bragfile",
			Type:        "shipped",
			Impact:      "unblocked search",
			CreatedAt:   time.Date(2026, 4, 22, 6, 30, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 4, 22, 6, 30, 0, 0, time.UTC),
		},
		{
			ID:    2,
			Title: "note",
		},
	}

	want := `[
  {
    "id": 1,
    "title": "shipped FTS5",
    "description": "migration 0002 landed",
    "tags": "sqlite,fts5",
    "project": "bragfile",
    "type": "shipped",
    "impact": "unblocked search",
    "created_at": "2026-04-22T06:30:00Z",
    "updated_at": "2026-04-22T06:30:00Z"
  },
  {
    "id": 2,
    "title": "note",
    "description": "",
    "tags": "",
    "project": "",
    "type": "",
    "impact": "",
    "created_at": "0001-01-01T00:00:00Z",
    "updated_at": "0001-01-01T00:00:00Z"
  }
]`

	got, err := ToJSON(fixture)
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}
	if !bytes.Equal(got, []byte(want)) {
		t.Fatalf("DEC-011 shape drift:\nwant:\n%s\n\ngot:\n%s", want, got)
	}

	// Secondary: parse and walk key order to confirm the 9 keys
	// appear in DEC-011 order (guards against future maintenance that
	// might swap struct-tag order without updating the golden).
	dec := json.NewDecoder(bytes.NewReader(got))
	dec.UseNumber()
	if tok, err := dec.Token(); err != nil || tok != json.Delim('[') {
		t.Fatalf("expected array open, got %v / %v", tok, err)
	}
	wantKeys := []string{"id", "title", "description", "tags", "project", "type", "impact", "created_at", "updated_at"}
	for i := 0; i < 2; i++ {
		if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
			t.Fatalf("entry %d: expected object open, got %v / %v", i, tok, err)
		}
		for j, wantKey := range wantKeys {
			tok, err := dec.Token()
			if err != nil {
				t.Fatalf("entry %d key %d: %v", i, j, err)
			}
			gotKey, ok := tok.(string)
			if !ok || gotKey != wantKey {
				t.Fatalf("entry %d key %d: want %q, got %v", i, j, wantKey, tok)
			}
			// discard value token (string or number, no nested objects here)
			if _, err := dec.Token(); err != nil {
				t.Fatalf("entry %d value %d: %v", i, j, err)
			}
		}
		if tok, err := dec.Token(); err != nil || tok != json.Delim('}') {
			t.Fatalf("entry %d: expected object close, got %v / %v", i, tok, err)
		}
	}
}

// TestToJSON_EmptyInputEmitsEmptyArray locks DEC-011 choice 1 (naked
// array) for the empty case: must be `[]`, never `null`.
func TestToJSON_EmptyInputEmitsEmptyArray(t *testing.T) {
	for _, input := range []struct {
		name    string
		entries []storage.Entry
	}{
		{"nil", nil},
		{"empty", []storage.Entry{}},
	} {
		t.Run(input.name, func(t *testing.T) {
			got, err := ToJSON(input.entries)
			if err != nil {
				t.Fatalf("ToJSON: %v", err)
			}
			if string(got) != "[]" {
				t.Fatalf("want %q, got %q", "[]", string(got))
			}
		})
	}
}

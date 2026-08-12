package capture

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// schemaPath is the checked-in JSON Schema BRAG.md tells callers to validate
// against before piping to `brag add --json`. Relative to this package.
const schemaPath = "../../docs/brag-entry.schema.json"

func loadSchemaProps(t *testing.T) map[string]map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Clean(schemaPath))
	if err != nil {
		t.Fatalf("read %s: %v", schemaPath, err)
	}
	var doc struct {
		Properties map[string]map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("parse %s: %v", schemaPath, err)
	}
	if len(doc.Properties) == 0 {
		t.Fatalf("%s has no properties", schemaPath)
	}
	return doc.Properties
}

// TestSchemaCapsMatchTheGoConstants pins docs/brag-entry.schema.json's
// maxLength values to this package's caps.
//
// The schema is a SECOND source of truth for every cap, and until this test
// existed nothing held the two in agreement. `just test-docs` asserts the
// schema's SHAPE ($schema, type, additionalProperties, tags.type, $id) but
// never its NUMBERS, and the Go tests that mention a cap assert the error
// STRING. Mutation-checked: changing MaxTitle 256→512 fails
// TestValidate_CapsRejectedOverBoundary on its hardcoded "256" message while
// `just test-docs` stays entirely green — so the live drift path was "change
// the constant, update the error-string expectation (which you must, to go
// green), and the schema silently keeps the old number with every gate
// passing."
//
// Not hypothetical: SPEC-075 moved MaxTitle 200→256 and MaxImpact 256→1024,
// and the schema was updated correctly — by a human remembering to sweep it,
// which is the exact mechanism DEC-046's punch-list item 2 removed for the
// reserved tag prefixes. Same defect class, same fix: make the coupling
// mechanical instead of remembered.
func TestSchemaCapsMatchTheGoConstants(t *testing.T) {
	props := loadSchemaProps(t)

	for _, tc := range []struct {
		field string
		want  int
	}{
		{"title", MaxTitle},
		{"description", MaxDescription},
		{"project", MaxProject},
		{"type", MaxType},
		{"impact", MaxImpact},
	} {
		p, ok := props[tc.field]
		if !ok {
			t.Errorf("%s: absent from the schema", tc.field)
			continue
		}
		raw, ok := p["maxLength"]
		if !ok {
			t.Errorf("%s: schema has no maxLength; Go caps it at %d", tc.field, tc.want)
			continue
		}
		got, ok := raw.(float64) // encoding/json numbers land as float64
		if !ok {
			t.Errorf("%s: maxLength is %T, want a number", tc.field, raw)
			continue
		}
		if int(got) != tc.want {
			t.Errorf("%s: schema maxLength=%d but the Go constant is %d — the schema drifted "+
				"from capture's caps (or vice versa); they must move together",
				tc.field, int(got), tc.want)
		}
	}
}

// TestSchemaTagsHasNoMaxLength pins the deliberate ABSENCE of a tags cap in
// the schema, so a future reader does not "helpfully" restore one.
//
// tags is the single field whose rule the schema cannot express. DEC-046
// replaced the joined-string cap with per-tag length (MaxTagLen) plus tag
// count (MaxTagCount), and no maxLength describes that shape — SPEC-075
// therefore removed the old `maxLength: 64` on purpose and said so in the
// property's own description.
//
// A `pattern` looks like it would close the gap and was tried:
// `^$|^[^,]{1,64}(,[^,]{1,64}){0,31}$` is correct under ECMA-262 (JSON
// Schema's dialect) and was verified against every boundary in Python. It is
// rejected here for a reason worth recording, because it is not obvious and
// someone will try again: **Go's regexp cannot compile it.** RE2 caps
// repetition expansion, and 64x32 overflows it —
// `error parsing regexp: invalid repeat count: {0,31}`. Shipping a constraint
// this project's own language cannot evaluate would mean the numbers inside it
// could never be pinned the way TestSchemaCapsMatchTheGoConstants pins the
// others, recreating in a regex exactly the unguarded second source of truth
// that test exists to eliminate.
//
// So the binary stays the authoritative validator for tags, the description
// states the rule for humans, and this test keeps the absence intentional
// rather than accidental. If a future change makes the rule expressible as a
// simple maxLength again, delete this test and extend the one above.
func TestSchemaTagsHasNoMaxLength(t *testing.T) {
	props := loadSchemaProps(t)
	tags, ok := props["tags"]
	if !ok {
		t.Fatal("tags: absent from the schema")
	}
	if v, present := tags["maxLength"]; present {
		t.Errorf("tags has maxLength=%v, but DEC-046 caps tags per-tag (MaxTagLen=%d) and "+
			"by count (MaxTagCount=%d) — no single maxLength expresses that, so any value "+
			"here is wrong. See this test's comment before restoring one.",
			v, MaxTagLen, MaxTagCount)
	}
	if _, present := tags["pattern"]; present {
		t.Errorf("tags has a pattern; a correct one (ECMA-262) does not compile under Go's " +
			"RE2 repetition limit, so its numbers could never be pinned to MaxTagLen/" +
			"MaxTagCount. See this test's comment.")
	}
}

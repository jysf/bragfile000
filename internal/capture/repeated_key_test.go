package capture

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// encodingJSONMatches reports whether encoding/json, decoding
// {"<sent>":"v"} into a struct whose only field is tagged `json:"<field>"`,
// puts the value in that field: the stdlib's own answer to "is this the same
// key", which the CLI ingress inherits.
func encodingJSONMatches(t *testing.T, field, sent string) bool {
	t.Helper()
	typ := reflect.StructOf([]reflect.StructField{{
		Name: "F",
		Type: reflect.TypeFor[string](),
		Tag:  reflect.StructTag(`json:"` + field + `"`),
	}})
	v := reflect.New(typ)
	raw, err := json.Marshal(map[string]string{sent: "v"})
	if err != nil {
		t.Fatalf("marshal %q: %v", sent, err)
	}
	if err := json.Unmarshal(raw, v.Interface()); err != nil {
		t.Fatalf("unmarshal %s: %v", raw, err)
	}
	return v.Elem().Field(0).String() == "v"
}

// TestCheckRepeatedKeys_FoldsExactlyLikeEncodingJSON pins DEC-055's
// key-identity rule to its source: for every pair, the detector calls the
// second key a repeat of the first exactly when encoding/json would decode
// both into one field. If a Go release changes how the stdlib folds, this
// fails instead of the CLI quietly clobbering again.
func TestCheckRepeatedKeys_FoldsExactlyLikeEncodingJSON(t *testing.T) {
	longS := string(rune(0x017F))     // ſ, LATIN SMALL LETTER LONG S
	kelvin := string(rune(0x212A))    // K, KELVIN SIGN
	sharpS := string(rune(0x00DF))    // ß: does not simple-fold to "ss"
	cyrillicE := string(rune(0x0435)) // е: a homoglyph of "e", not a fold
	pairs := []struct{ first, second string }{
		{"impact", "impact"},
		{"impact", "Impact"},
		{"IMPACT", "impact"},
		{"tags", "tag" + longS},
		{"kind", kelvin + "ind"},
		{"title", "TITLE"},
		{"strasse", "stra" + sharpS + "e"},
		{"impact", "impacts"},
		{"impact", "imp_act"},
		{"type", "typ" + cyrillicE},
		{string(rune(0x01C5)), string(rune(0x01C6))}, // ǅ and ǆ: title case, lower case
	}
	matched := 0
	for _, p := range pairs {
		want := encodingJSONMatches(t, p.first, p.second)
		if want {
			matched++
		}
		raw := `{"` + p.first + `":"a","` + p.second + `":"b"}`
		got := CheckRepeatedKeys([]byte(raw)) != nil
		if got != want {
			t.Errorf("%s: detector says repeat=%v, encoding/json matches one field=%v", raw, got, want)
		}
	}
	// Non-vacuity: the table must hold both answers, or it pins nothing.
	if matched < 6 || matched == len(pairs) {
		t.Fatalf("encoding/json matched %d of %d pairs; the table no longer tests both sides", matched, len(pairs))
	}
}

// TestCheckRepeatedKeys_ComparesDecodedKeys: an escaped spelling of a key
// is the same key. Both ingresses unescape before matching, so a raw-byte
// comparison would miss it.
func TestCheckRepeatedKeys_ComparesDecodedKeys(t *testing.T) {
	escaped := `{"impact":"REAL","` + "\x5c" + `u0069mpact":"CLOBBERED"}`
	if !strings.Contains(escaped, "\x5cu0069") {
		t.Fatalf("fixture lost its escape: %s", escaped)
	}
	err := CheckRepeatedKeys([]byte(escaped))
	if err == nil {
		t.Fatalf("%s: want a repeat, got nil", escaped)
	}
	if want := `key "impact" appears more than once`; err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

// TestCheckRepeatedKeys_NamesTheKey locks the two message shapes: the key as
// sent at its second occurrence, and the first spelling when they differ.
func TestCheckRepeatedKeys_NamesTheKey(t *testing.T) {
	cases := []struct{ raw, want string }{
		{`{"title":"t","impact":"A","impact":"B"}`, `key "impact" appears more than once`},
		{`{"title":"t","impact":"A","Impact":"B"}`, `key "Impact" appears more than once (first as "impact"; keys match case-insensitively)`},
		{`{"impact":"A","impact":null}`, `key "impact" appears more than once`},
		{`{"type":"shipped","type":"shipped"}`, `key "type" appears more than once`},
		{`{"a":1,"b":2,"c":3,"b":4,"a":5}`, `key "b" appears more than once`},
	}
	for _, tc := range cases {
		err := CheckRepeatedKeys([]byte(tc.raw))
		if err == nil {
			t.Errorf("%s: want %q, got nil", tc.raw, tc.want)
			continue
		}
		if err.Error() != tc.want {
			t.Errorf("%s: got %q, want %q", tc.raw, err.Error(), tc.want)
		}
	}
}

// TestCheckRepeatedKeys_TopLevelOnly: a repeat inside a nested value is not
// this rule's (DEC-055). Every value shape is skipped whole, so a key inside
// one never collides with a top-level key either.
func TestCheckRepeatedKeys_TopLevelOnly(t *testing.T) {
	for _, raw := range []string{
		`{"title":"t","id":{"a":1,"a":2}}`,
		`{"title":"t","id":[{"a":1,"a":2}]}`,
		`{"title":"t","id":{"title":"nested"}}`,
		`{"a":[1,{"b":2}],"b":null,"c":{"d":{"a":1}},"d":"x"}`,
	} {
		if err := CheckRepeatedKeys([]byte(raw)); err != nil {
			t.Errorf("%s: want nil, got %v", raw, err)
		}
	}
	// ...and a repeat that follows nested values is still found.
	raw := `{"a":[1,{"b":2}],"c":{"d":{"a":1}},"a":3}`
	if err := CheckRepeatedKeys([]byte(raw)); err == nil {
		t.Errorf("%s: want a repeat of \"a\", got nil", raw)
	}
}

// TestCheckRepeatedKeys_LeavesOtherInputToTheDecoder: what is not a
// well-formed object gets no second message from here; the ingress's own
// decoder already reports it (DEC-012 choice 1, invalid syntax).
func TestCheckRepeatedKeys_LeavesOtherInputToTheDecoder(t *testing.T) {
	for _, raw := range []string{
		``,
		`[{"a":1,"a":2}]`,
		`"a"`,
		`{"title":`,
		`{"a":1,"a"`,
		`{"a":1,"a":2`,
		`{"a":1}{"a":1}`,
	} {
		if err := CheckRepeatedKeys([]byte(raw)); err != nil {
			t.Errorf("%q: want nil, got %v", raw, err)
		}
	}
}

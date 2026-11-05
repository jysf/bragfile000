package capture

import (
	"bytes"
	"encoding/json"
	"fmt"
	"unicode"
	"unicode/utf8"
)

// CheckRepeatedKeys rejects a JSON object whose top level names the same key
// twice (DEC-055). It runs on the raw bytes, before any decoder sees them,
// because both machine ingresses decode last-wins with no error: by the time
// a struct holds the value, the repeat is gone.
//
// Two keys are the same key when they are equal after unescaping and case
// folding — the rule encoding/json uses to match a key to a struct field
// (foldKey). The MCP SDK matches case-sensitively, but one rule for both
// ingresses has to be the wider one, or `add --json` keeps clobbering on
// "impact" + "Impact".
//
// Only the top level is checked. No field brag stores holds an object, so a
// repeat inside a nested value either fails that field's type check or sits
// in a server-owned value DEC-012 discards.
//
// Anything that is not a well-formed object is left to the caller's decoder,
// which already reports it: this returns nil for an array, a scalar, or a
// syntax error anywhere in the object, so input that is broken in another way
// keeps the message it has today.
func CheckRepeatedKeys(raw []byte) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return nil
	}
	seen := map[string]string{} // folded key → the key as first sent
	var repeat error
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil
		}
		key, ok := tok.(string)
		if !ok {
			return nil
		}
		folded := foldKey(key)
		if first, dup := seen[folded]; !dup {
			seen[folded] = key
		} else if repeat == nil {
			repeat = repeatedKeyError(first, key)
		}
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return nil
		}
	}
	// Report a repeat only in an object that closes: a truncated one is the
	// decoder's to report, whatever it repeated before it broke.
	if _, err := dec.Token(); err != nil {
		return nil
	}
	return repeat
}

// repeatedKeyError names the key as sent at its second occurrence, and the
// first spelling too when the two differ only by case.
func repeatedKeyError(first, again string) error {
	if first == again {
		return fmt.Errorf("key %q appears more than once", again)
	}
	return fmt.Errorf("key %q appears more than once (first as %q; keys match case-insensitively)", again, first)
}

// foldKey maps every spelling encoding/json would match to one struct field
// onto the same string, so that foldKey(a) == foldKey(b) exactly when
// strings.EqualFold(a, b). It is encoding/json's own fold (appendFoldedName in
// its fold.go): ASCII letters upper-cased, and any other rune replaced by the
// smallest rune in its unicode.SimpleFold orbit, so "ſ" (U+017F) and "s" both
// fold to "S". TestCheckRepeatedKeys_FoldsExactlyLikeEncodingJSON holds the
// two in step.
func foldKey(key string) string {
	var b []byte
	for _, r := range key {
		b = utf8.AppendRune(b, foldRune(r))
	}
	return string(b)
}

func foldRune(r rune) rune {
	if 'a' <= r && r <= 'z' {
		return r - ('a' - 'A')
	}
	if r < utf8.RuneSelf {
		return r
	}
	for {
		next := unicode.SimpleFold(r)
		if next <= r {
			return next
		}
		r = next
	}
}

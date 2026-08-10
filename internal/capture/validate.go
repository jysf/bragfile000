// Package capture holds the input-validation rules shared by every brag
// capture ingress path (flag mode, editor mode, `add --json`, MCP brag_add,
// and — since DEC-046 — `brag edit`). It has no SQL and no cobra dependency
// so both the cli and mcpserver layers can call it, ending the per-path
// validation drift that previously let flag/editor mode skip the caps that
// --json and MCP enforced (SPEC-064). Each boundary wraps the returned error
// in its own type (cli.ErrUser for a user-facing exit code; a tool error for
// MCP).
package capture

import (
	"fmt"
	"strconv"
	"strings"
)

// Field byte caps. These are byte counts (len), matching the SPEC-061 byte
// decision. Each is derived from what its field is FOR, measured against the
// live corpus — not inherited from DEC-012's add --json schema, which is
// where the original values came from with no rationale recorded. See
// DEC-046 for the full derivation.
//
// MaxTags (a cap on the comma-joined string) is deliberately gone: DEC-015
// normalised tags into a table with no length constraint, so a joined-string
// cap penalised using many tags (legitimate) instead of one absurd tag (the
// actual abuse). MaxTagLen + MaxTagCount replace it with a shape that
// catches the real problem.
const (
	MaxTitle       = 256
	MaxDescription = 100000
	MaxTagLen      = 64
	MaxTagCount    = 32
	MaxProject     = 64
	MaxType        = 64
	MaxImpact      = 1024
)

// Fields is the raw, user-supplied entry text to validate at a capture
// ingress boundary. Validate is called on these values BEFORE any provenance
// stamping, so the tags cap and reserved-tag checks apply to what the caller
// actually typed, not to the post-stamp string.
type Fields struct {
	Title       string
	Description string
	Tags        string
	Project     string
	Type        string
	Impact      string
}

// Validate enforces the caps, control-char, and reserved-tag rules across
// every field, returning the first violation as a descriptive (unwrapped-
// sentinel) error the caller re-wraps.
func Validate(f Fields) error {
	if err := validateSingleLine("title", f.Title, MaxTitle); err != nil {
		return err
	}
	if err := validateTags(f.Tags); err != nil {
		return err
	}
	if err := validateSingleLine("project", f.Project, MaxProject); err != nil {
		return err
	}
	if err := validateSingleLine("type", f.Type, MaxType); err != nil {
		return err
	}
	if err := validateSingleLine("impact", f.Impact, MaxImpact); err != nil {
		return err
	}
	return validateDescription(f.Description)
}

// ValidateChanged enforces the same rules as Validate, but validates each
// field only when its value in new differs from old (DEC-046 LD4:
// grandfathering by validate-on-write-of-changed-field). A field the caller
// did not touch is never checked, even when its stored value is over a cap
// that predates this decision — which is what makes it safe to wire onto the
// edit path without making the already-over-cap corpus uneditable. The rule
// is "the changed value must meet today's cap", not "must not get worse
// than before": a field replaced with a different, still-over-cap value is
// rejected exactly like any other ingress.
func ValidateChanged(old, new Fields) error {
	if new.Title != old.Title {
		if err := validateSingleLine("title", new.Title, MaxTitle); err != nil {
			return err
		}
	}
	if new.Tags != old.Tags {
		if err := validateTags(new.Tags); err != nil {
			return err
		}
	}
	if new.Project != old.Project {
		if err := validateSingleLine("project", new.Project, MaxProject); err != nil {
			return err
		}
	}
	if new.Type != old.Type {
		if err := validateSingleLine("type", new.Type, MaxType); err != nil {
			return err
		}
	}
	if new.Impact != old.Impact {
		if err := validateSingleLine("impact", new.Impact, MaxImpact); err != nil {
			return err
		}
	}
	if new.Description != old.Description {
		if err := validateDescription(new.Description); err != nil {
			return err
		}
	}
	return nil
}

// validateSingleLine checks a single-line field (title, project, type,
// impact — tags has its own per-tag rule, see validateTags) against C0
// control characters and its byte cap.
func validateSingleLine(name, val string, max int) error {
	if hasC0Control(val) {
		return fmt.Errorf("%q must not contain control characters", name)
	}
	if len(val) > max {
		return fmt.Errorf("%q exceeds %d-character limit", name, max)
	}
	return nil
}

// validateDescription checks the multi-line description field: tabs and
// newlines are allowed, an embedded NUL is not (SQLite would truncate it),
// and it must not exceed MaxDescription.
func validateDescription(val string) error {
	if strings.IndexByte(val, 0x00) >= 0 {
		return fmt.Errorf("%q must not contain a NUL byte", "description")
	}
	if len(val) > MaxDescription {
		return fmt.Errorf("%q exceeds %d-character limit", "description", MaxDescription)
	}
	return nil
}

// validateTags enforces DEC-046's per-tag shape: each individual tag must
// not exceed MaxTagLen bytes (the cap that catches the actual abuse — one
// absurd tag), the entry must not carry more than MaxTagCount tags (applied
// PRE-STAMP, before the up-to-5 reserved provenance tags are appended, so
// legitimate provenance can never push an entry over), and any cost:/
// tokens: token is validated with the dedicated-param normalizers
// (DEC-027). A too-long tag's error names the offending TAG, not the
// joined string — that distinction is the whole point of the per-tag model.
func validateTags(tags string) error {
	toks := splitTags(tags)
	if len(toks) > MaxTagCount {
		return fmt.Errorf("tags: more than %d tags allowed (got %d)", MaxTagCount, len(toks))
	}
	for _, tok := range toks {
		if hasC0Control(tok) {
			return fmt.Errorf("%q must not contain control characters", "tags")
		}
		if len(tok) > MaxTagLen {
			return fmt.Errorf("tag %q exceeds %d-character limit", tok, MaxTagLen)
		}
	}
	return validateReservedTags(toks)
}

// splitTags splits a comma-joined tag string (DEC-004) into trimmed,
// non-empty tokens. An empty or whitespace-only field yields no tokens, so
// it never counts as a tag toward MaxTagCount.
func splitTags(tags string) []string {
	raw := strings.Split(tags, ",")
	toks := make([]string, 0, len(raw))
	for _, r := range raw {
		t := strings.TrimSpace(r)
		if t == "" {
			continue
		}
		toks = append(toks, t)
	}
	return toks
}

// hasC0Control reports whether s contains any C0 control byte (0x00–0x1F).
// Byte-wise is correct here: every C0 code point is a single UTF-8 byte, and
// no continuation byte of a multi-byte rune falls in 0x00–0x1F.
func hasC0Control(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 {
			return true
		}
	}
	return false
}

// validateReservedTags checks, for each already-split-and-trimmed tag token,
// whether it falls in the reserved cost:/tokens: namespaces, and if so
// validates the value with the dedicated-param normalizers. Invalid values
// are rejected; agent:/model:/session: tokens are intentionally not checked
// here.
func validateReservedTags(toks []string) error {
	for _, tok := range toks {
		if v, ok := strings.CutPrefix(tok, "cost:"); ok {
			if _, err := NormalizeCost(v); err != nil {
				return fmt.Errorf("invalid reserved tag %q: %w", tok, err)
			}
		} else if v, ok := strings.CutPrefix(tok, "tokens:"); ok {
			if _, err := NormalizeTokens(v); err != nil {
				return fmt.Errorf("invalid reserved tag %q: %w", tok, err)
			}
		}
	}
	return nil
}

// NormalizeCost validates a caller-reported cost as a non-negative USD decimal
// string (DEC-027). It trims whitespace; empty input yields ("", nil) so an
// omitted cost stamps no tag. A non-empty value must match a plain decimal
// shape ([0-9]+(\.[0-9]+)?) — rejecting negatives, scientific notation,
// currency symbols, and thousands separators — and is returned as the trimmed
// canonical string (never a re-formatted float, to avoid drift). Anything else
// is an error.
func NormalizeCost(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", nil
	}
	if !isDecimal(v) {
		return "", fmt.Errorf("cost %q: must be a non-negative decimal (e.g. 0.42)", raw)
	}
	if _, err := strconv.ParseFloat(v, 64); err != nil {
		return "", fmt.Errorf("cost %q: %w", raw, err)
	}
	return v, nil
}

// NormalizeTokens validates a caller-reported token count as a non-negative
// integer string (DEC-027). It trims whitespace; empty input yields ("", nil)
// so an omitted value stamps no tag. A non-empty value is parsed with
// strconv.ParseUint (rejecting negatives, decimals, non-digits, and radix
// prefixes) and returned as the trimmed string. Anything else is an error.
func NormalizeTokens(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", nil
	}
	if _, err := strconv.ParseUint(v, 10, 64); err != nil {
		return "", fmt.Errorf("tokens %q: must be a non-negative integer (e.g. 18000)", raw)
	}
	return v, nil
}

// isDecimal reports whether s is a plain non-negative decimal: one or more
// digits, optionally followed by a single '.' and one or more digits. No sign,
// exponent, separators, or leading/trailing dot. strconv.ParseFloat accepts
// "1e3", "+1", "  1", ".5" etc., so this tighter guard runs first.
func isDecimal(s string) bool {
	if s == "" {
		return false
	}
	dot := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			if dot || i == 0 || i == len(s)-1 {
				return false
			}
			dot = true
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

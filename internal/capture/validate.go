// Package capture holds the input-validation rules shared by every brag
// capture ingress path (flag mode, editor mode, `add --json`, and the MCP
// brag_add tool). It has no SQL and no cobra dependency so both the cli and
// mcpserver layers can call it, ending the per-path validation drift that
// previously let flag/editor mode skip the caps that --json and MCP enforced
// (SPEC-064). Each boundary wraps the returned error in its own type
// (cli.ErrUser for a user-facing exit code; a tool error for MCP).
package capture

import (
	"fmt"
	"strconv"
	"strings"
)

// Field byte caps. These are byte counts (len), matching the SPEC-061 byte
// decision, and are identical to the caps `add --json` and MCP brag_add have
// always enforced — the point of this package is that all paths share them.
const (
	MaxTitle       = 200
	MaxDescription = 100000
	MaxTags        = 64
	MaxProject     = 64
	MaxType        = 64
	MaxImpact      = 256
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

// Validate enforces three rules across the fields, returning the first
// violation as a descriptive (unwrapped-sentinel) error the caller re-wraps:
//
//   - Byte caps (A): each field must not exceed its MaxX cap.
//   - Control chars (B): single-line fields (title, tags, project, type,
//     impact) must contain no C0 control byte (0x00–0x1F, which includes
//     NUL, tab, newline, and carriage return) — one would break line-oriented
//     `brag list`/TSV output or be silently truncated by SQLite. description
//     is multi-line, so it permits tabs and newlines but still rejects an
//     embedded NUL (0x00), which SQLite would truncate.
//   - Reserved numeric tags (C): any cost:/tokens: token in the freeform tags
//     field is validated with the same rules as the dedicated cost/tokens
//     params, so a caller cannot smuggle garbage past them. agent:/model:/
//     session: tokens are left untouched (opaque, and the CLI's documented
//     provenance path).
func Validate(f Fields) error {
	singleLine := []struct {
		name string
		val  string
		max  int
	}{
		{"title", f.Title, MaxTitle},
		{"tags", f.Tags, MaxTags},
		{"project", f.Project, MaxProject},
		{"type", f.Type, MaxType},
		{"impact", f.Impact, MaxImpact},
	}
	for _, c := range singleLine {
		if hasC0Control(c.val) {
			return fmt.Errorf("%q must not contain control characters", c.name)
		}
		if len(c.val) > c.max {
			return fmt.Errorf("%q exceeds %d-character limit", c.name, c.max)
		}
	}

	// description is multi-line: tabs/newlines are allowed, NUL is not.
	if strings.IndexByte(f.Description, 0x00) >= 0 {
		return fmt.Errorf("%q must not contain a NUL byte", "description")
	}
	if len(f.Description) > MaxDescription {
		return fmt.Errorf("%q exceeds %d-character limit", "description", MaxDescription)
	}

	return validateReservedTags(f.Tags)
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

// validateReservedTags splits the freeform tags on ',' (DEC-004) and, for any
// token in the reserved cost:/tokens: namespaces, validates the value with the
// dedicated-param normalizers. Invalid values are rejected; agent:/model:/
// session: tokens are intentionally not checked here.
func validateReservedTags(tags string) error {
	for _, raw := range strings.Split(tags, ",") {
		tok := strings.TrimSpace(raw)
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

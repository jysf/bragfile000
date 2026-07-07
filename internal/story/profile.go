package story

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ErrProfileNotFound is returned by LoadProfile when a name has neither
// a user-override file nor a bundled default asset. The CLI surfaces it
// as a UserError naming the audience (AC-3).
var ErrProfileNotFound = errors.New("story profile not found")

// Profile is the parsed shape of an audience shaping profile. The source
// of truth is the YAML asset (bundled default or user override); this
// struct is DATA, not a Go enum (DEC-029 choice 2). Fields:
//   - Selection: ImpactThreadsOnly, DropImpactlessBeats
//   - Threading/altitude: FoldSmallThreads, ThreadOrder
//   - Candor is metadata surfaced to the LLM, not a body rule.
//   - Directive points at the framing-directive asset basename.
type Profile struct {
	Name                string
	DefaultWindow       string // "year" | "quarter" | "month" | "since:<raw>"
	ImpactThreadsOnly   bool
	DropImpactlessBeats bool
	FoldSmallThreads    bool
	ThreadOrder         string // "initiative" | "impact-desc"
	Candor              string // "candid" | "promotional"
	Directive           string // asset basename: "me.md" | "exec.md" | <user path>
}

// defaultOverrideDir resolves the user override directory
// (~/.bragfile/story-profiles). A missing HOME degrades to an empty
// path, which simply means "no override dir" (bundled defaults still
// load). Resolved once and cached in overrideDir.
func defaultOverrideDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".bragfile", "story-profiles")
}

// overrideDir is the injectable seam (§9) for the user-override profile
// directory. Tests substitute a t.TempDir(); production resolves the
// real config path. A package var (not a const) so Test 10 can point it
// at a seeded temp dir.
var overrideDir = defaultOverrideDir()

// LoadProfile resolves an audience shaping profile by name, override-wins
// (DEC-029 choice 2):
//  1. <overrideDir>/<name>.yaml, if present → parse it (malformed → error
//     naming the file; NO fallback to the bundled default).
//  2. else profiles/<name>.yaml in the embedded FS → parse it.
//  3. else → ErrProfileNotFound.
func LoadProfile(name string) (Profile, error) {
	if overrideDir != "" {
		path := filepath.Join(overrideDir, name+".yaml")
		if data, err := os.ReadFile(path); err == nil {
			p, perr := parseProfile(data)
			if perr != nil {
				return Profile{}, fmt.Errorf("parse override profile %q: %w", path, perr)
			}
			if p.Name == "" {
				p.Name = name
			}
			return p, nil
		} else if !os.IsNotExist(err) {
			return Profile{}, fmt.Errorf("read override profile %q: %w", path, err)
		}
	}

	if data, ok := bundledProfile(name); ok {
		p, err := parseProfile(data)
		if err != nil {
			return Profile{}, fmt.Errorf("parse bundled profile %q: %w", name, err)
		}
		if p.Name == "" {
			p.Name = name
		}
		return p, nil
	}

	return Profile{}, fmt.Errorf("%w: %q", ErrProfileNotFound, name)
}

// parseProfile is the hand-rolled, dependency-free parser for the flat
// `key: value` profile schema (Locked decision #7 — no YAML dep). The
// schema has no nesting, lists, or quoting edge cases, so a bufio line
// scan suffices. Blank lines and `#` comments are ignored. An unknown
// key is an error (typo protection); a non-boolean value for a boolean
// key is an error (naming the offending line).
func parseProfile(data []byte) (Profile, error) {
	var p Profile
	sc := bufio.NewScanner(bytes.NewReader(data))
	line := 0
	for sc.Scan() {
		line++
		raw := strings.TrimSpace(sc.Text())
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		key, val, ok := strings.Cut(raw, ":")
		if !ok {
			return Profile{}, fmt.Errorf("line %d: expected 'key: value', got %q", line, raw)
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch key {
		case "name":
			p.Name = val
		case "default_window":
			p.DefaultWindow = val
		case "thread_order":
			p.ThreadOrder = val
		case "candor":
			p.Candor = val
		case "directive":
			p.Directive = val
		case "impact_threads_only":
			b, err := parseBool(val, key, line)
			if err != nil {
				return Profile{}, err
			}
			p.ImpactThreadsOnly = b
		case "drop_impactless_beats":
			b, err := parseBool(val, key, line)
			if err != nil {
				return Profile{}, err
			}
			p.DropImpactlessBeats = b
		case "fold_small_threads":
			b, err := parseBool(val, key, line)
			if err != nil {
				return Profile{}, err
			}
			p.FoldSmallThreads = b
		default:
			return Profile{}, fmt.Errorf("line %d: unknown profile key %q", line, key)
		}
	}
	if err := sc.Err(); err != nil {
		return Profile{}, fmt.Errorf("scan profile: %w", err)
	}
	return p, nil
}

// ResolveDirective returns the framing-directive text for a profile. An
// empty Directive pointer yields an empty string (the bundle then omits
// the ## Framing directive section, AC-8). A user profile may point at an
// absolute/relative path outside the bundle; a bare basename resolves
// against the embedded directive assets.
func ResolveDirective(p Profile) (string, error) {
	if p.Directive == "" {
		return "", nil
	}
	// A path separator signals a user-supplied directive file on disk;
	// a bare basename resolves against the bundled assets.
	if strings.ContainsRune(p.Directive, os.PathSeparator) {
		b, err := os.ReadFile(p.Directive)
		if err != nil {
			return "", fmt.Errorf("read directive file %q: %w", p.Directive, err)
		}
		return string(b), nil
	}
	b, err := directiveAsset(p.Directive)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func parseBool(val, key string, line int) (bool, error) {
	b, err := strconv.ParseBool(val)
	if err != nil {
		return false, fmt.Errorf("line %d: key %q wants a boolean, got %q", line, key, val)
	}
	return b, nil
}

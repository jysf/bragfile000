// Package story builds and renders `brag story` (SPEC-049): audience-shaped
// bundles of Threads and Beats over a window of entries. BuildThreads and
// BuildThroughline (thread.go) project []storage.Entry into the threading
// shape a Profile selects; ToStoryMarkdown/ToStoryJSON (bundle.go, this
// file) render the result per DEC-014's envelope. Profile (profile.go) is
// DATA, not a Go enum — bundled defaults live under profiles/*.yaml and
// directives/*.md (embed.go, DEC-029 choice 2), with an optional
// ~/.bragfile/story-profiles/ user override shadowing a bundled name by
// name. The package reads storage.Entry values but never a *storage.Store
// or a SQL driver, and reads no clock itself — the CLI resolves the window,
// the directive text, and "now", and passes them in via StoryOptions.
package story

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Markers for the standalone-readable markdown body (locked in the
// goldens): an impact beat that is not a failure leads with ★ (U+2605), a
// plain beat with · (U+00B7) — a visible "so what" signal. A recorded
// failure leads with ✗ (U+2717) and "(failed)", whether or not it carries an
// impact, so ★ never marks one (DEC-050, DEC-054).
const (
	markerImpact  = "★"
	markerPlain   = "·"
	markerFailure = "✗"
)

// StoryOptions is the pure renderer's input. The CLI does the windowing,
// threading, throughline, and directive resolution and passes them in
// (mirrors export.ImpactOptions). Scope echoes the resolved window token;
// EntriesInWindow is the raw in-window count for the <shown>/<in-window>
// beat tally; Now is injected for deterministic goldens. OmittedFailures
// is OmitFailures' count, and drives both the Omitted: line and the clause
// appended to the directive, so a caller cannot set one without the other.
type StoryOptions struct {
	Audience        string
	Scope           string
	Filters         string            // pre-formatted markdown line ("(none)" or echoed flags)
	FiltersJSON     map[string]string // JSON filters object (nil → {})
	EntriesInWindow int
	OmittedFailures int
	Now             time.Time
	Threads         []Thread
	Throughline     Throughline
	Directive       string // resolved framing-directive text ("" → section omitted)
}

// pluralFailures renders a failure count for the Omitted: line and the
// clause, which must agree on it.
func pluralFailures(n int) string {
	if n == 1 {
		return "1 recorded failure"
	}
	return fmt.Sprintf("%d recorded failures", n)
}

// omissionClause is the fixed, binary-authored text appended to the framing
// directive when a promotional profile omitted failures (DEC-054). It comes
// from the binary, not the directive asset, because a user profile can point
// directive: at its own file, which would not carry it. Its wording was
// measured against a consuming model at SPEC-094 design; change it only with
// a re-run of that measurement.
func omissionClause(n int) string {
	return fmt.Sprintf("This bundle omits %s for this audience. End with one line that says so; do not drop it.", pluralFailures(n))
}

// framingDirective is the directive both formats render: the resolved text,
// with omissionClause appended as its own paragraph when failures were
// omitted. An empty directive with an omission is the clause alone, so the
// instruction survives a profile that carries no directive.
func framingDirective(opts StoryOptions) string {
	if opts.OmittedFailures == 0 {
		return opts.Directive
	}
	clause := omissionClause(opts.OmittedFailures)
	d := strings.TrimRight(opts.Directive, "\n")
	if d == "" {
		return clause + "\n"
	}
	return d + "\n\n" + clause + "\n"
}

// shownBeats counts the beats surfaced across all threads (the numerator
// of the Beats: <shown>/<in-window> tally).
func shownBeats(threads []Thread) int {
	n := 0
	for _, t := range threads {
		n += len(t.Beats)
	}
	return n
}

// ToStoryMarkdown renders the arc-aware bundle as markdown (AC-1/AC-2/
// AC-8/AC-10). Header + provenance always render; the ## Threads and
// ## Throughline sections omit on an empty corpus (DEC-014 empty-state);
// the ## Framing directive section renders whenever the directive is
// non-empty (it is valid on an empty corpus — DEC-029 choice 8). Returns
// bytes with the trailing "\n" stripped (matches ToImpactMarkdown).
func ToStoryMarkdown(opts StoryOptions) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "# Bragfile Story")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "Generated: %s\n", opts.Now.UTC().Format(time.RFC3339))
	fmt.Fprintf(&buf, "Scope: %s\n", opts.Scope)
	fmt.Fprintf(&buf, "Audience: %s\n", opts.Audience)
	fmt.Fprintf(&buf, "Filters: %s\n", opts.Filters)
	fmt.Fprintf(&buf, "Threads: %d\n", len(opts.Threads))
	fmt.Fprintf(&buf, "Beats: %d/%d\n", shownBeats(opts.Threads), opts.EntriesInWindow)
	if opts.OmittedFailures > 0 {
		fmt.Fprintf(&buf, "Omitted: %s, not listed for this audience (brag list --type failed)\n", pluralFailures(opts.OmittedFailures))
	}

	if len(opts.Threads) > 0 {
		fmt.Fprintln(&buf)
		fmt.Fprintln(&buf, "## Threads")
		for _, t := range opts.Threads {
			fmt.Fprintln(&buf)
			fmt.Fprintf(&buf, "### %s\n", t.Thread)
			fmt.Fprintln(&buf)
			for _, b := range t.Beats {
				switch {
				case b.IsFailure:
					fmt.Fprintf(&buf, "- %s %d (failed): %s\n", markerFailure, b.ID, b.Title)
					if b.IsImpactBeat {
						fmt.Fprintf(&buf, "  %s\n", b.Impact)
					}
				case b.IsImpactBeat:
					fmt.Fprintf(&buf, "- %s %d: %s\n", markerImpact, b.ID, b.Title)
					fmt.Fprintf(&buf, "  %s\n", b.Impact)
				default:
					fmt.Fprintf(&buf, "- %s %d: %s\n", markerPlain, b.ID, b.Title)
				}
			}
		}

		fmt.Fprintln(&buf)
		fmt.Fprintln(&buf, "## Throughline (skeleton)")
		fmt.Fprintln(&buf)
		for _, a := range opts.Throughline.Arcs {
			fmt.Fprintf(&buf, "- %s [%s]: %s, %d with impact (%s → %s)\n",
				a.Thread, a.Kind, pluralBeats(a.BeatCount), a.ImpactBeatCount,
				a.Span.First.UTC().Format("2006-01-02"),
				a.Span.Last.UTC().Format("2006-01-02"))
		}
	}

	if directive := framingDirective(opts); directive != "" {
		fmt.Fprintln(&buf)
		fmt.Fprintln(&buf, "## Framing directive")
		fmt.Fprintln(&buf)
		buf.WriteString(directive)
	}

	return trimTrailingNewline(buf.Bytes()), nil
}

func pluralBeats(n int) string {
	if n == 1 {
		return "1 beat"
	}
	return fmt.Sprintf("%d beats", n)
}

// trimTrailingNewline drops a single trailing "\n" so the CLI's
// Fprintln adds exactly one (matches export.trimTrailingNewline).
func trimTrailingNewline(b []byte) []byte {
	return bytes.TrimRight(b, "\n")
}

// storyEnvelope is the on-the-wire JSON shape (DEC-029 choice 5). Field
// order = key order (encoding/json preserves struct-tag declaration
// order): generated_at, scope, audience, filters, omitted_failure_count,
// threads, throughline, framing_directive. Extends DEC-014's envelope with
// the arc-aware body. omitted_failure_count is always present, 0 when
// nothing was omitted (DEC-014 part 4, DEC-054).
type storyEnvelope struct {
	GeneratedAt         string            `json:"generated_at"`
	Scope               string            `json:"scope"`
	Audience            string            `json:"audience"`
	Filters             map[string]string `json:"filters"`
	OmittedFailureCount int               `json:"omitted_failure_count"`
	Threads             []threadJSON      `json:"threads"`
	Throughline         throughlineJSON   `json:"throughline"`
	FramingDirective    string            `json:"framing_directive"`
}

type threadJSON struct {
	Thread string     `json:"thread"`
	Kind   string     `json:"kind"`
	Span   spanJSON   `json:"span"`
	Beats  []beatJSON `json:"beats"`
}

// beatJSON is the 7-key beat projection (DEC-029 choice 5) — a deliberate
// subset+2 of DEC-011's 9-key shape.
type beatJSON struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Project      string `json:"project"`
	Type         string `json:"type"`
	Impact       string `json:"impact"`
	IsImpactBeat bool   `json:"is_impact_beat"`
	CreatedAt    string `json:"created_at"` // RFC3339
}

type throughlineJSON struct {
	Arcs []arcJSON `json:"arcs"`
}

type arcJSON struct {
	Thread          string   `json:"thread"`
	Kind            string   `json:"kind"`
	BeatCount       int      `json:"beat_count"`
	ImpactBeatCount int      `json:"impact_beat_count"`
	Span            spanJSON `json:"span"`
}

type spanJSON struct {
	First string `json:"first"` // RFC3339
	Last  string `json:"last"`
}

// ToStoryJSON renders the arc-aware DEC-014-extending envelope with
// 2-space indent (AC-6/AC-10). Threads/Arcs init to non-nil empty
// slices; Filters nil → {}; framing_directive is always the resolved
// directive string, with the omission clause when one applies (renders
// even on an empty corpus).
func ToStoryJSON(opts StoryOptions) ([]byte, error) {
	env := storyEnvelope{
		GeneratedAt:         opts.Now.UTC().Format(time.RFC3339),
		Scope:               opts.Scope,
		Audience:            opts.Audience,
		Filters:             opts.FiltersJSON,
		OmittedFailureCount: opts.OmittedFailures,
		Threads:             make([]threadJSON, 0, len(opts.Threads)),
		Throughline:         throughlineJSON{Arcs: make([]arcJSON, 0, len(opts.Throughline.Arcs))},
		FramingDirective:    framingDirective(opts),
	}
	if env.Filters == nil {
		env.Filters = map[string]string{}
	}

	for _, t := range opts.Threads {
		tj := threadJSON{
			Thread: t.Thread,
			Kind:   t.Kind,
			Span:   spanToJSON(threadSpan(t)),
			Beats:  make([]beatJSON, 0, len(t.Beats)),
		}
		for _, b := range t.Beats {
			tj.Beats = append(tj.Beats, beatJSON{
				ID:           b.ID,
				Title:        b.Title,
				Project:      b.Project,
				Type:         b.Type,
				Impact:       b.Impact,
				IsImpactBeat: b.IsImpactBeat,
				CreatedAt:    b.CreatedAt.UTC().Format(time.RFC3339),
			})
		}
		env.Threads = append(env.Threads, tj)
	}

	for _, a := range opts.Throughline.Arcs {
		env.Throughline.Arcs = append(env.Throughline.Arcs, arcJSON{
			Thread:          a.Thread,
			Kind:            a.Kind,
			BeatCount:       a.BeatCount,
			ImpactBeatCount: a.ImpactBeatCount,
			Span:            spanToJSON(a.Span),
		})
	}

	return json.MarshalIndent(env, "", "  ")
}

func spanToJSON(s Span) spanJSON {
	return spanJSON{
		First: s.First.UTC().Format(time.RFC3339),
		Last:  s.Last.UTC().Format(time.RFC3339),
	}
}

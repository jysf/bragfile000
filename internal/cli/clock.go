package cli

import "time"

// clock is this package's injectable LOCAL-wall-clock seam (AGENTS.md §9). It
// returns the current instant carrying the LOCAL location (time.Local in
// production), which is what `brag list --day` needs: the zone rides on the
// injected now, so `internal/timewindow.ParseDay` reads the day boundary off
// `clock().Location()` (DEC-039). Tests substitute a fixed instant+zone to
// make keyword/day resolution deterministic across timezones.
//
// Distinct from impact.go's `nowFunc`, which deliberately returns UTC for the
// calendar-window commands: a "day" is a local concept, a reporting
// quarter/month/year is anchored in UTC there. Do not merge the two.
//
// Lives here rather than in since.go because SPEC-072/DEC-042 moved the
// parsers themselves into the pure `internal/timewindow` package, which holds
// no clock — each caller owns its own policy and passes `now` in.
var clock = time.Now

package main

import (
	"regexp"
	"runtime/debug"
	"strings"
)

// releaseTagPattern matches ONLY a clean release version — "1.2.3", with no
// pre-release suffix and no build metadata.
//
// This is deliberately strict, and the strictness is the whole point. Go
// synthesises a PSEUDO-VERSION for anything built outside a tagged release —
// `v0.6.1-0.20260812074351-43675b5b3b48` for a build off an untagged commit,
// and the literal `(devel)` for a plain local `go build`. Those look like
// versions but are not releases, and storage.isDevBuild only recognises the
// markers "-dev", "-dirty" and "-snapshot", none of which a pseudo-version
// carries. Accepting one would therefore report an unreleased build as a
// release AND silently switch off the DEC-026 guard that stops a dev binary
// migrating the real ~/.bragfile. Matching only vX.Y.Z keeps both behaviours
// correct without teaching isDevBuild about Go's version grammar.
var releaseTagPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// resolveVersion decides what this binary reports as its version, and what the
// DEC-026 dev/prod-migration guard is told.
//
// Two build paths produce a real release and they stamp it differently:
//
//   - goreleaser sets main.version via ldflags (-X main.version=0.6.0). That
//     wins whenever present — it is the authoritative release stamp.
//   - `go install github.com/jysf/bragfile000/cmd/brag@latest` does NOT run
//     ldflags, so main.version stays "dev" even though the module was fetched
//     at a real tag. Before this, every go-install user saw "brag version dev".
//     The module version is recoverable from the embedded build info, so use it
//     — but only when it is a clean release tag (see releaseTagPattern).
//
// Anything else — a pseudo-version, "(devel)", an unreadable build info —
// falls through to the ldflags default, which is "dev", which is exactly what
// an unreleased build should say.
func resolveVersion(ldflagsVersion string, readBuildInfo func() (*debug.BuildInfo, bool)) string {
	if ldflagsVersion != "" && ldflagsVersion != "dev" {
		return ldflagsVersion // goreleaser stamped a real release.
	}
	bi, ok := readBuildInfo()
	if !ok || bi == nil {
		return ldflagsVersion
	}
	// Module versions carry a leading "v"; goreleaser's {{ .Version }} does
	// not. Normalise to goreleaser's shape so `brag --version` reads the same
	// however the binary was produced.
	v := strings.TrimPrefix(bi.Main.Version, "v")
	if releaseTagPattern.MatchString(v) {
		return v
	}
	return ldflagsVersion
}

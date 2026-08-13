package main

import (
	"runtime/debug"
	"testing"
)

func buildInfo(moduleVersion string) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: moduleVersion}}, true
	}
}

// TestResolveVersion covers the three build paths that reach this binary and,
// most importantly, the ones that must NOT be mistaken for a release.
//
// The stakes are not cosmetic. storage.SetBuildVersion feeds the DEC-026
// guard: a build whose version reads as a release is ALLOWED to migrate the
// real ~/.bragfile, and storage.isDevBuild only recognises "-dev", "-dirty"
// and "-snapshot". A Go pseudo-version (`v0.6.1-0.20260812074351-43655b3b48`)
// carries none of those markers, so returning one here would report an
// unreleased build as a release and silently switch the guard off. That is the
// case this table exists for.
func TestResolveVersion(t *testing.T) {
	for _, tc := range []struct {
		name     string
		ldflags  string
		module   string
		want     string
		wantWhy  string
		devGuard bool // true = must still look like a dev build downstream
	}{
		{
			name: "goreleaser ldflags win", ldflags: "0.6.1", module: "v0.5.2",
			want: "0.6.1", wantWhy: "the ldflags stamp is authoritative for a real release",
		},
		{
			name:    "go install at a tag recovers the module version",
			ldflags: "dev", module: "v0.6.0",
			want: "0.6.0", wantWhy: "go install runs no ldflags; the module version is the real one",
		},
		{
			name:    "leading v is stripped to match goreleaser's shape",
			ldflags: "dev", module: "v1.2.3",
			want: "1.2.3", wantWhy: "goreleaser's {{ .Version }} has no leading v",
		},
		{
			name:    "PSEUDO-VERSION stays dev",
			ldflags: "dev", module: "v0.6.1-0.20260812074351-43655b3b4812",
			want: "dev", devGuard: true,
			wantWhy: "an untagged build is not a release; accepting it would disable the DEC-026 guard",
		},
		{
			name: "(devel) stays dev", ldflags: "dev", module: "(devel)",
			want: "dev", devGuard: true,
			wantWhy: "a plain local go build is not a release",
		},
		{
			name: "pre-release tag stays dev", ldflags: "dev", module: "v0.7.0-rc1",
			want: "dev", devGuard: true,
			wantWhy: "only a clean vX.Y.Z counts; an RC is stamped by goreleaser via ldflags anyway",
		},
		{
			name: "empty module version stays dev", ldflags: "dev", module: "",
			want: "dev", devGuard: true, wantWhy: "nothing to recover",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveVersion(tc.ldflags, buildInfo(tc.module))
			if got != tc.want {
				t.Errorf("resolveVersion(%q, module=%q) = %q, want %q — %s",
					tc.ldflags, tc.module, got, tc.want, tc.wantWhy)
			}
			if tc.devGuard && got != "dev" {
				t.Errorf("%q must remain \"dev\" so the DEC-026 dev/prod-migration guard "+
					"still fires; got %q", tc.module, got)
			}
		})
	}
}

// TestResolveVersion_UnreadableBuildInfoFallsBack pins the degenerate case:
// if the build info cannot be read at all, report whatever ldflags said rather
// than inventing something.
func TestResolveVersion_UnreadableBuildInfoFallsBack(t *testing.T) {
	none := func() (*debug.BuildInfo, bool) { return nil, false }
	if got := resolveVersion("dev", none); got != "dev" {
		t.Errorf("got %q, want \"dev\"", got)
	}
	if got := resolveVersion("0.6.1", none); got != "0.6.1" {
		t.Errorf("got %q, want the ldflags value \"0.6.1\"", got)
	}
}

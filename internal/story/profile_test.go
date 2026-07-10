package story

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withOverrideDir swaps the package override-dir seam for a test.
func withOverrideDir(t *testing.T, dir string) {
	t.Helper()
	prev := overrideDir
	overrideDir = dir
	t.Cleanup(func() { overrideDir = prev })
}

// Test 9 — TestLoadProfile_BundledDefaults.
func TestLoadProfile_BundledDefaults(t *testing.T) {
	// Point the override dir at an empty temp dir so only bundled assets
	// resolve.
	withOverrideDir(t, t.TempDir())

	me, err := LoadProfile("me")
	if err != nil {
		t.Fatalf("LoadProfile(me): %v", err)
	}
	if me.Name != "me" || me.DefaultWindow != "year" {
		t.Errorf("me: got name=%q window=%q, want me/year", me.Name, me.DefaultWindow)
	}
	if me.ImpactThreadsOnly || me.DropImpactlessBeats || me.FoldSmallThreads {
		t.Errorf("me should keep all threads/beats, got %+v", me)
	}
	if me.ThreadOrder != "initiative" || me.Directive != "me.md" {
		t.Errorf("me: got order=%q directive=%q, want initiative/me.md", me.ThreadOrder, me.Directive)
	}

	exec, err := LoadProfile("exec")
	if err != nil {
		t.Fatalf("LoadProfile(exec): %v", err)
	}
	if exec.DefaultWindow != "quarter" {
		t.Errorf("exec window: got %q, want quarter", exec.DefaultWindow)
	}
	if !exec.ImpactThreadsOnly || !exec.DropImpactlessBeats || !exec.FoldSmallThreads {
		t.Errorf("exec should fold/drop, got %+v", exec)
	}
	if exec.ThreadOrder != "impact-desc" || exec.Directive != "exec.md" {
		t.Errorf("exec: got order=%q directive=%q, want impact-desc/exec.md", exec.ThreadOrder, exec.Directive)
	}

	// The audience set is data-driven: it comes from the embedded FS, not
	// a Go enum. Assert both names are present as assets by listing the FS.
	names, err := bundledProfileNames()
	if err != nil {
		t.Fatalf("bundledProfileNames: %v", err)
	}
	if !contains(names, "me") || !contains(names, "exec") {
		t.Errorf("embedded profile names %v must include me + exec", names)
	}
}

// Test 10 — TestLoadProfile_UserOverrideShadowsBundled.
func TestLoadProfile_UserOverrideShadowsBundled(t *testing.T) {
	dir := t.TempDir()
	// An override me.yaml that changes the default window to quarter.
	override := "name: me\ndefault_window: quarter\nthread_order: initiative\ndirective: me.md\n"
	if err := os.WriteFile(filepath.Join(dir, "me.yaml"), []byte(override), 0o644); err != nil {
		t.Fatal(err)
	}
	withOverrideDir(t, dir)

	me, err := LoadProfile("me")
	if err != nil {
		t.Fatalf("LoadProfile(me): %v", err)
	}
	if me.DefaultWindow != "quarter" {
		t.Errorf("override should win: got window %q, want quarter", me.DefaultWindow)
	}
}

// Test 11 — TestLoadProfile_UnknownAndMalformed.
func TestLoadProfile_UnknownAndMalformed(t *testing.T) {
	withOverrideDir(t, t.TempDir())

	t.Run("unknown", func(t *testing.T) {
		_, err := LoadProfile("nope")
		if !errors.Is(err, ErrProfileNotFound) {
			t.Errorf("expected ErrProfileNotFound, got %v", err)
		}
	})

	t.Run("malformed-override-no-fallback", func(t *testing.T) {
		dir := t.TempDir()
		// A malformed override for an otherwise-known name (bad key line).
		if err := os.WriteFile(filepath.Join(dir, "me.yaml"), []byte("this is not key value\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		withOverrideDir(t, dir)
		_, err := LoadProfile("me")
		if err == nil {
			t.Fatal("expected error for malformed override, got nil (must NOT fall back)")
		}
		if !containsSubstr(err.Error(), "me.yaml") {
			t.Errorf("malformed override error should name the file, got %v", err)
		}
	})
}

// Test 12 — TestDirectiveAsset_ResolvesAndIsNonEmpty.
func TestDirectiveAsset_ResolvesAndIsNonEmpty(t *testing.T) {
	me, err := directiveAsset("me.md")
	if err != nil {
		t.Fatalf("directiveAsset(me.md): %v", err)
	}
	exec, err := directiveAsset("exec.md")
	if err != nil {
		t.Fatalf("directiveAsset(exec.md): %v", err)
	}
	if len(me) == 0 || len(exec) == 0 {
		t.Fatal("directives must be non-empty")
	}
	if string(me) == string(exec) {
		t.Error("me and exec directives must differ")
	}
	// The me directive mentions the messy middle / lessons; exec mentions
	// business impact / one headline.
	if !containsSubstr(string(me), "messy middle") {
		t.Errorf("me directive should mention the messy middle")
	}
	if !containsSubstr(string(exec), "business impact") {
		t.Errorf("exec directive should mention business impact")
	}
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

func containsSubstr(s, sub string) bool {
	return strings.Contains(s, sub)
}

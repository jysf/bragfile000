package storage

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

// withGuardSeams sets the dev/prod-migration guard's package seams for one
// test and restores them on cleanup. defaultPath is what the guard treats as
// the real production DB; override toggles BRAG_ALLOW_DEV_PROD_MIGRATE=1.
func withGuardSeams(t *testing.T, version, defaultPath string, override bool) {
	t.Helper()
	ov, odp, ole := buildVersion, defaultDBPathFn, lookupEnv
	buildVersion = version
	defaultDBPathFn = func() (string, error) { return defaultPath, nil }
	lookupEnv = func(k string) (string, bool) {
		if k == envAllowDevProdMigrate && override {
			return "1", true
		}
		return "", false
	}
	t.Cleanup(func() { buildVersion, defaultDBPathFn, lookupEnv = ov, odp, ole })
}

func migrationCount(t *testing.T, path string) int {
	t.Helper()
	db := openRawDB(t, path)
	var n int
	if err := db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM schema_migrations").Scan(&n); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	return n
}

// TestDevProdGuard_DevBuildRefusesPendingMigrationOnProdDB ▲ SPEC-044 — a dev
// build opening the real ~/.bragfile with pending migrations is refused: no
// migration applied, no backup sidecar, ErrDevProdMigrate returned.
func TestDevProdGuard_DevBuildRefusesPendingMigrationOnProdDB(t *testing.T) {
	rawDB, path := apply0001Only(t)
	_ = rawDB.Close()
	withGuardSeams(t, "dev", path, false) // `path` IS the production default

	_, err := Open(path)
	if !errors.Is(err, ErrDevProdMigrate) {
		t.Fatalf("want ErrDevProdMigrate, got %v", err)
	}
	if n := migrationCount(t, path); n != 1 {
		t.Errorf("guard should apply no migration: schema_migrations=%d, want 1", n)
	}
	matches, _ := filepath.Glob(path + ".pre-*.backup")
	if len(matches) != 0 {
		t.Errorf("guard should write no backup sidecar, got %v", matches)
	}
}

// TestDevProdGuard_ReleasedBuildMigrates ▲ a clean release version is not a dev
// build → the migration (and its DEC-021 backup) proceeds normally.
func TestDevProdGuard_ReleasedBuildMigrates(t *testing.T) {
	rawDB, path := apply0001Only(t)
	_ = rawDB.Close()
	withGuardSeams(t, "0.3.0", path, false)

	s, err := Open(path)
	if err != nil {
		t.Fatalf("released build should migrate, got %v", err)
	}
	defer s.Close()
	if n := migrationCount(t, path); n != 4 {
		t.Errorf("released build should apply all migrations: got %d, want 4", n)
	}
}

// TestDevProdGuard_OverrideMigrates ▲ BRAG_ALLOW_DEV_PROD_MIGRATE=1 re-permits
// the intentional case.
func TestDevProdGuard_OverrideMigrates(t *testing.T) {
	rawDB, path := apply0001Only(t)
	_ = rawDB.Close()
	withGuardSeams(t, "dev", path, true) // override ON

	s, err := Open(path)
	if err != nil {
		t.Fatalf("override should permit migration, got %v", err)
	}
	defer s.Close()
	if n := migrationCount(t, path); n != 4 {
		t.Errorf("override should apply all migrations: got %d, want 4", n)
	}
}

// TestDevProdGuard_ThrowawayPathUnaffected ▲ a dev build against a NON-default
// DB (a --db/BRAGFILE_DB throwaway) is unaffected.
func TestDevProdGuard_ThrowawayPathUnaffected(t *testing.T) {
	rawDB, path := apply0001Only(t)
	_ = rawDB.Close()
	// The "production default" is some OTHER path, so `path` is a throwaway.
	withGuardSeams(t, "dev", filepath.Join(t.TempDir(), "prod.db"), false)

	s, err := Open(path)
	if err != nil {
		t.Fatalf("throwaway --db should be unaffected, got %v", err)
	}
	defer s.Close()
	if n := migrationCount(t, path); n != 4 {
		t.Errorf("throwaway path should migrate: got %d, want 4", n)
	}
}

// TestDevProdGuard_FreshDefaultDBUnaffected ▲ a dev build creating a BRAND-NEW
// default DB (applied==0) is fine — there is no established data to protect
// (mirrors the DEC-021 backup carve-out).
func TestDevProdGuard_FreshDefaultDBUnaffected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.sqlite")
	withGuardSeams(t, "dev", path, false) // dev build, path is the default, but brand-new

	s, err := Open(path)
	if err != nil {
		t.Fatalf("fresh default DB should open for a dev build, got %v", err)
	}
	defer s.Close()
	if n := migrationCount(t, path); n != 4 {
		t.Errorf("fresh DB should apply all migrations: got %d, want 4", n)
	}
}

// TestDevProdGuard_UpToDateDefaultDBUnaffected ▲ a dev build opening an
// already-migrated default DB (pending==0) is a read-only case — unaffected.
func TestDevProdGuard_UpToDateDefaultDBUnaffected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.sqlite")
	// First open (as a released build) brings it to head.
	withGuardSeams(t, "0.3.0", path, false)
	s0, err := Open(path)
	if err != nil {
		t.Fatalf("initial Open: %v", err)
	}
	_ = s0.Close()
	// Now a dev build opens the up-to-date default DB: no pending → no guard.
	withGuardSeams(t, "dev", path, false)
	s, err := Open(path)
	if err != nil {
		t.Fatalf("up-to-date default DB should open for a dev build, got %v", err)
	}
	defer s.Close()
}

// TestIsDevBuild ▲ locks the dev-build classifier.
func TestIsDevBuild(t *testing.T) {
	dev := []string{"", "dev", "0.3.0-dev", "0.2.0-SNAPSHOT-7fdfe29", "0.3.0-dirty"}
	rel := []string{"0.3.0", "v0.3.0", "1.2.3", "0.3.0-rc1"}
	for _, v := range dev {
		if !isDevBuild(v) {
			t.Errorf("isDevBuild(%q) = false, want true", v)
		}
	}
	for _, v := range rel {
		if isDevBuild(v) {
			t.Errorf("isDevBuild(%q) = true, want false", v)
		}
	}
}

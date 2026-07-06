package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jysf/bragfile000/internal/config"
)

// ErrDevProdMigrate is returned (wrapped) by Open when an unreleased/dev
// build would apply a pending migration to the real ~/.bragfile database.
// It is a user-actionable condition (the message names the override), so the
// CLI maps it to exit 1. See DEC-026 / SPEC-044.
var ErrDevProdMigrate = errors.New("dev build refused to migrate the production database")

// envAllowDevProdMigrate is the override env var: set it to "1" to permit a
// dev build to migrate the real ~/.bragfile.
const envAllowDevProdMigrate = "BRAG_ALLOW_DEV_PROD_MIGRATE"

// buildVersion is the running binary's version, set by main via
// SetBuildVersion. goreleaser injects a clean semver on release; local
// builds keep "dev". Package-level so tests can substitute it (§9 seam).
var buildVersion = "dev"

// SetBuildVersion records the running binary's version for the dev/prod
// migration guard. main calls it once at startup, from the same ldflags
// `version` it passes to cli.NewRootCmd.
func SetBuildVersion(v string) { buildVersion = v }

// defaultDBPathFn resolves the real production DB path. Injectable so tests
// can treat a t.TempDir() file as "the production DB" without touching the
// user's real ~/.bragfile.
var defaultDBPathFn = config.DefaultDBPath

// lookupEnv reads the override env var; injectable for hermetic tests.
var lookupEnv = os.LookupEnv

// isDevBuild reports whether v denotes an unreleased build: empty, "dev", or
// a version carrying a pre-release/dirty/snapshot marker (as goreleaser's
// snapshot build and a dirty tree produce). A clean release tag ("0.3.0",
// "v0.3.0") — and a release candidate ("0.3.0-rc1", which is a real
// goreleaser artifact) — are NOT dev builds.
func isDevBuild(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" || v == "dev" {
		return true
	}
	for _, marker := range []string{"-dev", "-dirty", "-snapshot"} {
		if strings.Contains(v, marker) {
			return true
		}
	}
	return false
}

// devProdMigrateGuard refuses to migrate the real production DB from an
// unreleased build. It fires on exactly the DEC-021 backup trigger — an
// established DB about to be mutated (applied>0 && pending>0) — but ONLY when
// the build is a dev build, the path is the resolved production default, and
// the override is unset. All other cases (released build, override, throwaway
// --db, brand-new DB, up-to-date DB) return nil and let Open proceed.
func devProdMigrateGuard(ctx context.Context, db *sql.DB, path string, src fs.FS) error {
	if !isDevBuild(buildVersion) {
		return nil
	}
	if v, ok := lookupEnv(envAllowDevProdMigrate); ok && v == "1" {
		return nil
	}
	def, err := defaultDBPathFn()
	if err != nil || !samePath(path, def) {
		return nil // can't resolve the default, or this is not the prod DB.
	}
	applied, pending, err := migrationStatus(ctx, db, src)
	if err != nil {
		return nil // let the normal apply path surface the real error.
	}
	if len(applied) == 0 || len(pending) == 0 {
		return nil // brand-new or already at head — nothing established to protect.
	}
	return fmt.Errorf("%w: %q has %d pending migration(s) and this is an unreleased build (%q). "+
		"Refusing to migrate your real ~/.bragfile. Use a released `brag`, point --db / %s at a "+
		"throwaway copy, or set %s=1 to override.",
		ErrDevProdMigrate, path, len(pending), buildVersion, "BRAGFILE_DB", envAllowDevProdMigrate)
}

// samePath compares two DB paths for equality after cleaning. It does not
// resolve symlinks — the default path and a resolved path are constructed the
// same way, so lexical equality is sufficient in practice.
func samePath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

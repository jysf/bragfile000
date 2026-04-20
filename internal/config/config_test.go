package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveDBPath_Default(t *testing.T) {
	t.Setenv("BRAGFILE_DB", "")

	got, err := ResolveDBPath("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("expected absolute path, got %q", got)
	}
	if strings.Contains(got, "~") {
		t.Errorf("expected no tilde in path, got %q", got)
	}
	if !strings.HasSuffix(got, "/.bragfile/db.sqlite") {
		t.Errorf("expected path ending with /.bragfile/db.sqlite, got %q", got)
	}
}

func TestResolveDBPath_FlagWins(t *testing.T) {
	t.Setenv("BRAGFILE_DB", "")

	got, err := ResolveDBPath("/tmp/foo.db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/tmp/foo.db" {
		t.Errorf("expected /tmp/foo.db, got %q", got)
	}
}

func TestResolveDBPath_EnvWhenFlagEmpty(t *testing.T) {
	t.Setenv("BRAGFILE_DB", "/tmp/env.db")

	got, err := ResolveDBPath("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/tmp/env.db" {
		t.Errorf("expected /tmp/env.db, got %q", got)
	}
}

func TestResolveDBPath_FlagBeatsEnv(t *testing.T) {
	t.Setenv("BRAGFILE_DB", "/tmp/env.db")

	got, err := ResolveDBPath("/tmp/flag.db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/tmp/flag.db" {
		t.Errorf("expected /tmp/flag.db, got %q", got)
	}
}

func TestResolveDBPath_ExpandsTilde(t *testing.T) {
	t.Setenv("BRAGFILE_DB", "")

	got, err := ResolveDBPath("~/weird.db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(got, "~") {
		t.Errorf("expected no tilde in path, got %q", got)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("expected absolute path, got %q", got)
	}
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, "weird.db")
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

package cli

import (
	"errors"
	"strings"
	"testing"
)

func TestErrUser_IsDetectable(t *testing.T) {
	err := ErrUser
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected errors.Is(err, ErrUser) to be true")
	}
}

func TestUserErrorf_FormatsAndWraps(t *testing.T) {
	err := UserErrorf("bad %q", "x")
	if !errors.Is(err, ErrUser) {
		t.Fatalf("expected UserErrorf to wrap ErrUser; got %v", err)
	}
	if !strings.Contains(err.Error(), `bad "x"`) {
		t.Fatalf("expected error message to contain %q, got %q", `bad "x"`, err.Error())
	}
}

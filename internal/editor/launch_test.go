package editor

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestLaunch_ChangedBufferReturnsNewBytes(t *testing.T) {
	fake := func(path string) error {
		return os.WriteFile(path, []byte("NEW"), 0o600)
	}
	edited, changed, err := Launch([]byte("OLD"), fake)
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	if !changed {
		t.Errorf("changed = false, want true")
	}
	if !bytes.Equal(edited, []byte("NEW")) {
		t.Errorf("edited = %q, want %q", string(edited), "NEW")
	}
}

func TestLaunch_UnchangedBufferReturnsFalse(t *testing.T) {
	fake := func(path string) error { return nil }
	edited, changed, err := Launch([]byte("X"), fake)
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	if changed {
		t.Errorf("changed = true, want false")
	}
	if !bytes.Equal(edited, []byte("X")) {
		t.Errorf("edited = %q, want %q", string(edited), "X")
	}
}

func TestLaunch_EditFuncErrorPropagates(t *testing.T) {
	// Write modified bytes before returning an error, to exercise the
	// "error + changed buffer" path. The "error + unchanged buffer" path
	// is the :cq abort covered by TestLaunch_EditFuncErrorOnUnmodifiedFileIsAborted.
	fake := func(path string) error {
		if err := os.WriteFile(path, []byte("PARTIAL"), 0o600); err != nil {
			return err
		}
		return errors.New("boom")
	}
	_, _, err := Launch([]byte("X"), fake)
	if err == nil {
		t.Fatal("Launch: expected error, got nil")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("boom")) {
		t.Errorf("error = %q; expected to contain %q", err.Error(), "boom")
	}
}

// TestLaunch_EditFuncErrorOnUnmodifiedFileIsAborted pins locked design
// decision #6: when the editor exits non-zero but the buffer is byte-
// identical to the initial state (the git :cq convention), Launch treats
// it as an aborted edit — returns the initial bytes, changed=false, nil error.
func TestLaunch_EditFuncErrorOnUnmodifiedFileIsAborted(t *testing.T) {
	fake := func(path string) error { return errors.New(":cq") }
	edited, changed, err := Launch([]byte("ORIGINAL"), fake)
	if err != nil {
		t.Fatalf("Launch: expected nil error for aborted edit, got %v", err)
	}
	if changed {
		t.Errorf("changed = true, want false")
	}
	if !bytes.Equal(edited, []byte("ORIGINAL")) {
		t.Errorf("edited = %q, want %q", string(edited), "ORIGINAL")
	}
}

func TestResolveEditor_UsesEditor(t *testing.T) {
	t.Setenv("EDITOR", "foo")
	t.Setenv("VISUAL", "")
	got := resolveEditor()
	want := []string{"foo"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("resolveEditor = %v, want %v", got, want)
	}
}

func TestResolveEditor_FallsBackToVisual(t *testing.T) {
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "bar")
	got := resolveEditor()
	want := []string{"bar"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("resolveEditor = %v, want %v", got, want)
	}
}

func TestResolveEditor_DefaultsToVi(t *testing.T) {
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "")
	got := resolveEditor()
	want := []string{"vi"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("resolveEditor = %v, want %v", got, want)
	}
}

func TestResolveEditor_SplitsOnWhitespace(t *testing.T) {
	t.Setenv("EDITOR", "code --wait")
	t.Setenv("VISUAL", "")
	got := resolveEditor()
	want := []string{"code", "--wait"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("resolveEditor = %v, want %v", got, want)
	}
}

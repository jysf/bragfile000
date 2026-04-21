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
	fake := func(path string) error { return errors.New("boom") }
	_, _, err := Launch([]byte("X"), fake)
	if err == nil {
		t.Fatal("Launch: expected error, got nil")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("boom")) {
		t.Errorf("error = %q; expected to contain %q", err.Error(), "boom")
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

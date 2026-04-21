package storage

import (
	"errors"
	"testing"
)

func TestGet_RoundTripsAllFields(t *testing.T) {
	s, _ := newTestStore(t)

	in := Entry{
		Title:       "shipped widget v1",
		Description: "did the thing, saved the day",
		Tags:        "auth,perf",
		Project:     "platform",
		Type:        "shipped",
		Impact:      "p99 -30%",
	}
	inserted, err := s.Add(in)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := s.Get(inserted.ID)
	if err != nil {
		t.Fatalf("Get(%d): %v", inserted.ID, err)
	}

	if got.ID != inserted.ID {
		t.Errorf("ID: got %d want %d", got.ID, inserted.ID)
	}
	if got.Title != in.Title {
		t.Errorf("Title: got %q want %q", got.Title, in.Title)
	}
	if got.Description != in.Description {
		t.Errorf("Description: got %q want %q", got.Description, in.Description)
	}
	if got.Tags != in.Tags {
		t.Errorf("Tags: got %q want %q", got.Tags, in.Tags)
	}
	if got.Project != in.Project {
		t.Errorf("Project: got %q want %q", got.Project, in.Project)
	}
	if got.Type != in.Type {
		t.Errorf("Type: got %q want %q", got.Type, in.Type)
	}
	if got.Impact != in.Impact {
		t.Errorf("Impact: got %q want %q", got.Impact, in.Impact)
	}
	if loc := got.CreatedAt.Location().String(); loc != "UTC" {
		t.Errorf("CreatedAt.Location = %q, want UTC", loc)
	}
	if !got.UpdatedAt.Equal(got.CreatedAt) {
		t.Errorf("UpdatedAt (%v) != CreatedAt (%v)", got.UpdatedAt, got.CreatedAt)
	}
}

func TestGet_NotFoundReturnsErrNotFound(t *testing.T) {
	s, _ := newTestStore(t)

	got, err := s.Get(42)
	if err == nil {
		t.Fatalf("expected error for missing id, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected errors.Is(err, ErrNotFound) to be true; got %v", err)
	}
	if got != (Entry{}) {
		t.Errorf("expected zero-value Entry on not-found, got %+v", got)
	}
}

func TestGet_PartiallyEmptyFieldsHydrateAsEmptyStrings(t *testing.T) {
	s, _ := newTestStore(t)

	inserted, err := s.Add(Entry{Title: "only title"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := s.Get(inserted.ID)
	if err != nil {
		t.Fatalf("Get(%d): %v", inserted.ID, err)
	}

	if got.Title != "only title" {
		t.Errorf("Title: got %q want %q", got.Title, "only title")
	}
	if got.Description != "" {
		t.Errorf("Description: got %q want empty", got.Description)
	}
	if got.Tags != "" {
		t.Errorf("Tags: got %q want empty", got.Tags)
	}
	if got.Project != "" {
		t.Errorf("Project: got %q want empty", got.Project)
	}
	if got.Type != "" {
		t.Errorf("Type: got %q want empty", got.Type)
	}
	if got.Impact != "" {
		t.Errorf("Impact: got %q want empty", got.Impact)
	}
}

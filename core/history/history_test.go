package history

import (
	"path/filepath"
	"testing"
)

func TestHistoryStoreAddListUpdateClear(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "history.json"))

	entry, err := store.Add(Entry{
		Direction: DirectionSent,
		FileName:  "video.mp4",
		SizeBytes: 123,
		PeerName:  "PC-Max",
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if entry.ID == "" {
		t.Fatal("expected generated ID")
	}

	updated, err := store.UpdateStatus(entry.ID, StatusCompleted, "abc123", "")
	if err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}
	if updated.Status != StatusCompleted || updated.SHA256 != "abc123" {
		t.Fatalf("unexpected updated entry: %+v", updated)
	}

	entries, err := store.List(10)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if err := store.Clear(); err != nil {
		t.Fatalf("Clear returned error: %v", err)
	}
	entries, _ = store.List(10)
	if len(entries) != 0 {
		t.Fatalf("expected empty history, got %d", len(entries))
	}
}

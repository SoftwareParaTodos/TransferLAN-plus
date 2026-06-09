package notify

import (
	"path/filepath"
	"testing"
)

func TestStoreAddListMarkReadClear(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "notifications.json"))
	_, err := store.Add(TypeTransferDone, "Listo", "Archivo recibido", "a.mp4", "/tmp/a.mp4", 123)
	if err != nil {
		t.Fatal(err)
	}
	events, err := store.List(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Read {
		t.Fatalf("evento inesperado: %#v", events)
	}
	if err := store.MarkAllRead(); err != nil {
		t.Fatal(err)
	}
	events, _ = store.List(10)
	if !events[0].Read {
		t.Fatal("debería estar leído")
	}
	if err := store.Clear(); err != nil {
		t.Fatal(err)
	}
	events, _ = store.List(10)
	if len(events) != 0 {
		t.Fatalf("esperaba 0 eventos, got %d", len(events))
	}
}

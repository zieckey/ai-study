package memory

import (
	"context"
	"path/filepath"
	"testing"
)

func TestStoreSetGetListDelete(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "memory.json"))
	ctx := context.Background()

	if err := store.Set(ctx, "language", "Go"); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	entry, ok, err := store.Get(ctx, "language")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !ok {
		t.Fatal("Get did not find entry")
	}
	if entry.Value != "Go" {
		t.Fatalf("Value = %q", entry.Value)
	}

	entries, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d", len(entries))
	}

	deleted, err := store.Delete(ctx, "language")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if !deleted {
		t.Fatal("Delete returned false")
	}
}

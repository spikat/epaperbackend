package cache_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonathanribas/epaperbackend/pkg/cache"
)

func TestStoreSetGetAndTTL(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store, err := cache.Open(dir, "test", 1, cache.DefaultMigrate)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	key := "abc"
	payload := []byte(`{"hello":"world"}`)

	if err := store.Set(ctx, key, payload, time.Minute); err != nil {
		t.Fatalf("set: %v", err)
	}

	entry, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry")
	}
	if string(entry.Payload) != string(payload) {
		t.Fatalf("payload mismatch: %s", entry.Payload)
	}
	if !store.IsValid(entry, time.Now().UTC()) {
		t.Fatal("entry should be valid")
	}
}

func TestStoreExpiredEntry(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store, err := cache.Open(dir, "test", 1, cache.DefaultMigrate)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	key := "expired"
	if err := store.Set(ctx, key, []byte("x"), -time.Second); err != nil {
		t.Fatalf("set: %v", err)
	}

	entry, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if store.IsValid(entry, time.Now().UTC()) {
		t.Fatal("entry should be expired")
	}
}

func TestStorePurgeExpired(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store, err := cache.Open(dir, "test", 1, cache.DefaultMigrate)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	if err := store.Set(ctx, "old", []byte("1"), -time.Second); err != nil {
		t.Fatalf("set old: %v", err)
	}
	if err := store.Set(ctx, "new", []byte("2"), time.Minute); err != nil {
		t.Fatalf("set new: %v", err)
	}
	if err := store.PurgeExpired(ctx); err != nil {
		t.Fatalf("purge: %v", err)
	}

	old, err := store.Get(ctx, "old")
	if err != nil {
		t.Fatalf("get old: %v", err)
	}
	if old != nil {
		t.Fatal("old entry should be purged")
	}

	newEntry, err := store.Get(ctx, "new")
	if err != nil {
		t.Fatalf("get new: %v", err)
	}
	if newEntry == nil {
		t.Fatal("new entry should remain")
	}
}

func TestStoreSchemaVersionMismatch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "broken.db")
	store, err := cache.Open(dir, "broken", 1, cache.DefaultMigrate)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	_ = store.Close()

	dbPath := path
	_ = dbPath

	store2, err := cache.Open(dir, "broken", 2, cache.DefaultMigrate)
	if err == nil {
		_ = store2.Close()
		t.Fatal("expected schema version mismatch error")
	}
}

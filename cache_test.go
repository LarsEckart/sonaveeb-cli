package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	// Create temp directory for test cache
	tmpDir, err := os.MkdirTemp("", "sonaveeb-cache-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cachePath := filepath.Join(tmpDir, "test.db")
	cache, err := OpenCacheAt(cachePath)
	if err != nil {
		t.Fatalf("failed to open cache: %v", err)
	}
	defer cache.Close()

	t.Run("get missing key returns nil", func(t *testing.T) {
		entry, err := cache.Get("nonexistent")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if entry != nil {
			t.Errorf("expected nil, got %v", entry)
		}
	})

	t.Run("set and get", func(t *testing.T) {
		key := "search:puu"
		value := []byte(`{"words":[{"wordId":123}]}`)

		if err := cache.Set(key, value); err != nil {
			t.Fatalf("failed to set: %v", err)
		}

		entry, err := cache.Get(key)
		if err != nil {
			t.Fatalf("failed to get: %v", err)
		}
		if string(entry.Value) != string(value) {
			t.Errorf("got %s, want %s", entry.Value, value)
		}
	})

	t.Run("set overwrites existing", func(t *testing.T) {
		key := "details:123"
		cache.Set(key, []byte("old"))
		cache.Set(key, []byte("new"))

		entry, _ := cache.Get(key)
		if string(entry.Value) != "new" {
			t.Errorf("got %s, want new", entry.Value)
		}
	})

	t.Run("created_at is set", func(t *testing.T) {
		key := "test:timestamp"
		before := time.Now().Add(-time.Second)
		cache.Set(key, []byte("value"))
		after := time.Now().Add(time.Second)

		entry, _ := cache.Get(key)
		if entry.CreatedAt.Before(before) || entry.CreatedAt.After(after) {
			t.Errorf("created_at %v not in expected range [%v, %v]", entry.CreatedAt, before, after)
		}
	})

	t.Run("delete", func(t *testing.T) {
		key := "todelete"
		cache.Set(key, []byte("value"))
		cache.Delete(key)

		entry, _ := cache.Get(key)
		if entry != nil {
			t.Errorf("expected nil after delete, got %v", entry)
		}
	})

	t.Run("clear", func(t *testing.T) {
		cache.Set("key1", []byte("val1"))
		cache.Set("key2", []byte("val2"))
		cache.Clear()

		entry1, _ := cache.Get("key1")
		entry2, _ := cache.Get("key2")
		if entry1 != nil || entry2 != nil {
			t.Errorf("expected all keys cleared")
		}
	})
}

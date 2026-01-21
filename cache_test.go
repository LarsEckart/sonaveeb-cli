package main

import (
	"os"
	"path/filepath"
	"testing"
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
		val, err := cache.Get("nonexistent")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != nil {
			t.Errorf("expected nil, got %v", val)
		}
	})

	t.Run("set and get", func(t *testing.T) {
		key := "search:puu"
		value := []byte(`{"words":[{"wordId":123}]}`)

		if err := cache.Set(key, value); err != nil {
			t.Fatalf("failed to set: %v", err)
		}

		got, err := cache.Get(key)
		if err != nil {
			t.Fatalf("failed to get: %v", err)
		}
		if string(got) != string(value) {
			t.Errorf("got %s, want %s", got, value)
		}
	})

	t.Run("set overwrites existing", func(t *testing.T) {
		key := "details:123"
		cache.Set(key, []byte("old"))
		cache.Set(key, []byte("new"))

		got, _ := cache.Get(key)
		if string(got) != "new" {
			t.Errorf("got %s, want new", got)
		}
	})

	t.Run("delete", func(t *testing.T) {
		key := "todelete"
		cache.Set(key, []byte("value"))
		cache.Delete(key)

		got, _ := cache.Get(key)
		if got != nil {
			t.Errorf("expected nil after delete, got %v", got)
		}
	})

	t.Run("clear", func(t *testing.T) {
		cache.Set("key1", []byte("val1"))
		cache.Set("key2", []byte("val2"))
		cache.Clear()

		got1, _ := cache.Get("key1")
		got2, _ := cache.Get("key2")
		if got1 != nil || got2 != nil {
			t.Errorf("expected all keys cleared")
		}
	})
}

package main

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Cache provides a simple key-value store backed by SQLite.
type Cache struct {
	db *sql.DB
}

// OpenCache opens or creates a cache at the default location.
// Returns nil cache (not an error) if the path cannot be created.
func OpenCache() (*Cache, error) {
	path, err := defaultCachePath()
	if err != nil {
		return nil, nil // No cache, not an error
	}
	return OpenCacheAt(path)
}

// OpenCacheAt opens or creates a cache at the given path.
func OpenCacheAt(path string) (*Cache, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := initSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Cache{db: db}, nil
}

func initSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cache (
			key   TEXT PRIMARY KEY,
			value BLOB
		)
	`)
	return err
}

func defaultCachePath() (string, error) {
	// Prefer XDG_CACHE_HOME, fall back to ~/.cache
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		cacheDir = filepath.Join(home, ".cache")
	}
	return filepath.Join(cacheDir, "sonaveeb", "cache.db"), nil
}

// Get retrieves a value from the cache. Returns nil if not found.
func (c *Cache) Get(key string) ([]byte, error) {
	var value []byte
	err := c.db.QueryRow("SELECT value FROM cache WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Set stores a value in the cache.
func (c *Cache) Set(key string, value []byte) error {
	_, err := c.db.Exec(
		"INSERT OR REPLACE INTO cache (key, value) VALUES (?, ?)",
		key, value,
	)
	return err
}

// Delete removes a key from the cache.
func (c *Cache) Delete(key string) error {
	_, err := c.db.Exec("DELETE FROM cache WHERE key = ?", key)
	return err
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() error {
	_, err := c.db.Exec("DELETE FROM cache")
	return err
}

// Close closes the cache database.
func (c *Cache) Close() error {
	return c.db.Close()
}

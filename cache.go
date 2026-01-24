package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Cache provides a simple key-value store backed by SQLite.
type Cache struct {
	db *sql.DB
}

// CacheEntry holds a cached value and its metadata.
type CacheEntry struct {
	Value     []byte
	CreatedAt time.Time
}

// OpenCache opens or creates a cache at the default location.
// Returns an error if the cache directory or database cannot be created.
func OpenCache() (*Cache, error) {
	path, err := defaultCachePath()
	if err != nil {
		return nil, err
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
		_ = db.Close()
		return nil, err
	}

	return &Cache{db: db}, nil
}

func initSchema(db *sql.DB) error {
	// Check if we need to migrate (old schema without created_at)
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('cache') WHERE name = 'created_at'
	`).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return err // Propagate unexpected errors
	}

	if err == nil && count == 0 {
		// Old schema exists, drop it (it's just a cache)
		if _, execErr := db.Exec("DROP TABLE IF EXISTS cache"); execErr != nil {
			return execErr
		}
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cache (
			key        TEXT PRIMARY KEY,
			value      BLOB,
			created_at INTEGER
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
func (c *Cache) Get(key string) (*CacheEntry, error) {
	var value []byte
	var createdAt int64
	err := c.db.QueryRow(
		"SELECT value, created_at FROM cache WHERE key = ?", key,
	).Scan(&value, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &CacheEntry{
		Value:     value,
		CreatedAt: time.Unix(createdAt, 0),
	}, nil
}

// Set stores a value in the cache.
func (c *Cache) Set(key string, value []byte) error {
	_, err := c.db.Exec(
		"INSERT OR REPLACE INTO cache (key, value, created_at) VALUES (?, ?, ?)",
		key, value, time.Now().Unix(),
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

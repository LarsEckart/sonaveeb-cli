package cache

import (
	"context"
	"database/sql"
	"fmt"
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
func OpenCache() (*Cache, error) {
	path, err := defaultCachePath()
	if err != nil {
		return nil, fmt.Errorf("resolve cache path: %w", err)
	}
	return OpenCacheAt(path)
}

// OpenCacheAt opens or creates a cache at the given path.
func OpenCacheAt(path string) (*Cache, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create cache directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open cache database: %w", err)
	}

	if err := initSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize cache schema: %w", err)
	}

	return &Cache{db: db}, nil
}

func initSchema(db *sql.DB) error {
	ctx := context.Background()

	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pragma_table_info('cache') WHERE name = 'created_at'
	`).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == nil && count == 0 {
		if _, execErr := db.ExecContext(ctx, "DROP TABLE IF EXISTS cache"); execErr != nil {
			return execErr
		}
	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS cache (
			key        TEXT PRIMARY KEY,
			value      BLOB,
			created_at INTEGER
		)
	`)
	return err
}

func defaultCachePath() (string, error) {
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
	ctx := context.Background()

	var value []byte
	var createdAt int64
	err := c.db.QueryRowContext(
		ctx,
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
	ctx := context.Background()

	_, err := c.db.ExecContext(
		ctx,
		"INSERT OR REPLACE INTO cache (key, value, created_at) VALUES (?, ?, ?)",
		key, value, time.Now().Unix(),
	)
	return err
}

// Delete removes a key from the cache.
func (c *Cache) Delete(key string) error {
	_, err := c.db.ExecContext(context.Background(), "DELETE FROM cache WHERE key = ?", key)
	return err
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() error {
	_, err := c.db.ExecContext(context.Background(), "DELETE FROM cache")
	return err
}

// Close closes the cache database.
func (c *Cache) Close() error {
	return c.db.Close()
}

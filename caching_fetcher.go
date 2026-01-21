package main

import (
	"fmt"
	"log"
)

// CachingFetcher wraps a Fetcher with a cache layer.
// On cache hit, returns cached data. On miss, fetches from upstream and caches.
type CachingFetcher struct {
	upstream Fetcher
	cache    *Cache
	refresh  bool // If true, bypass cache reads (but still write)
}

// NewCachingFetcher creates a caching fetcher.
// If cache is nil, it behaves like the upstream fetcher.
// If refresh is true, it bypasses cache reads but still updates the cache.
func NewCachingFetcher(upstream Fetcher, cache *Cache, refresh bool) *CachingFetcher {
	return &CachingFetcher{
		upstream: upstream,
		cache:    cache,
		refresh:  refresh,
	}
}

func (f *CachingFetcher) Search(word string) ([]byte, error) {
	return f.cachedFetch("search:"+word, func() ([]byte, error) {
		return f.upstream.Search(word)
	})
}

func (f *CachingFetcher) WordDetails(wordID int64) ([]byte, error) {
	return f.cachedFetch(fmt.Sprintf("details:%d", wordID), func() ([]byte, error) {
		return f.upstream.WordDetails(wordID)
	})
}

func (f *CachingFetcher) ParadigmDetails(wordID int64) ([]byte, error) {
	return f.cachedFetch(fmt.Sprintf("paradigm:%d", wordID), func() ([]byte, error) {
		return f.upstream.ParadigmDetails(wordID)
	})
}

func (f *CachingFetcher) cachedFetch(key string, fetch func() ([]byte, error)) ([]byte, error) {
	// No cache? Just fetch.
	if f.cache == nil {
		return fetch()
	}

	// Check cache (unless refresh mode)
	if !f.refresh {
		entry, err := f.cache.Get(key)
		if err != nil {
			log.Printf("cache get error for %q: %v", key, err)
		} else if entry != nil {
			return entry.Value, nil
		}
	}

	// Fetch from upstream
	data, err := fetch()
	if err != nil {
		return nil, err
	}

	// Store in cache (best-effort; log errors)
	if err := f.cache.Set(key, data); err != nil {
		log.Printf("cache set error for %q: %v", key, err)
	}

	return data, nil
}

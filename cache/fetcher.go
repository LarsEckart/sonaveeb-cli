package cache

import (
	"fmt"

	"github.com/LarsEckart/sonaveeb-cli/sonaveeb"
)

// CachingFetcher wraps a Fetcher with a cache layer.
type CachingFetcher struct {
	upstream sonaveeb.Fetcher
	cache    *Cache
	refresh  bool
}

func NewCachingFetcher(upstream sonaveeb.Fetcher, cache *Cache, refresh bool) *CachingFetcher {
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
	if f.cache == nil {
		return fetch()
	}

	if !f.refresh {
		entry, err := f.cache.Get(key)
		if err == nil && entry != nil {
			return entry.Value, nil
		}
	}

	data, err := fetch()
	if err != nil {
		return nil, err
	}

	_ = f.cache.Set(key, data)
	return data, nil
}

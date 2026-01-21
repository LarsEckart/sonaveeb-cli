package main

import (
	"os"
	"path/filepath"
	"testing"
)

// MockFetcher records calls and returns canned responses.
type MockFetcher struct {
	SearchCalls   []string
	DetailsCalls  []int64
	ParadigmCalls []int64

	SearchResponse   []byte
	DetailsResponse  []byte
	ParadigmResponse []byte
}

func (m *MockFetcher) Search(word string) ([]byte, error) {
	m.SearchCalls = append(m.SearchCalls, word)
	return m.SearchResponse, nil
}

func (m *MockFetcher) WordDetails(wordID int64) ([]byte, error) {
	m.DetailsCalls = append(m.DetailsCalls, wordID)
	return m.DetailsResponse, nil
}

func (m *MockFetcher) ParadigmDetails(wordID int64) ([]byte, error) {
	m.ParadigmCalls = append(m.ParadigmCalls, wordID)
	return m.ParadigmResponse, nil
}

func TestCachingFetcher(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "sonaveeb-caching-test")
	defer os.RemoveAll(tmpDir)

	cache, _ := OpenCacheAt(filepath.Join(tmpDir, "test.db"))
	defer cache.Close()

	mock := &MockFetcher{
		SearchResponse:   []byte(`{"words":[]}`),
		DetailsResponse:  []byte(`{"wordClass":"noun"}`),
		ParadigmResponse: []byte(`[]`),
	}

	fetcher := NewCachingFetcher(mock, cache, false)

	t.Run("caches search results", func(t *testing.T) {
		// First call — should hit upstream
		data1, _ := fetcher.Search("puu")
		if len(mock.SearchCalls) != 1 {
			t.Errorf("expected 1 upstream call, got %d", len(mock.SearchCalls))
		}
		if string(data1) != `{"words":[]}` {
			t.Errorf("unexpected response: %s", data1)
		}

		// Second call — should hit cache
		data2, _ := fetcher.Search("puu")
		if len(mock.SearchCalls) != 1 {
			t.Errorf("expected still 1 upstream call (cache hit), got %d", len(mock.SearchCalls))
		}
		if string(data2) != string(data1) {
			t.Errorf("cache returned different data")
		}
	})

	t.Run("caches word details", func(t *testing.T) {
		fetcher.WordDetails(123)
		fetcher.WordDetails(123)
		if len(mock.DetailsCalls) != 1 {
			t.Errorf("expected 1 upstream call, got %d", len(mock.DetailsCalls))
		}
	})

	t.Run("caches paradigm details", func(t *testing.T) {
		fetcher.ParadigmDetails(456)
		fetcher.ParadigmDetails(456)
		if len(mock.ParadigmCalls) != 1 {
			t.Errorf("expected 1 upstream call, got %d", len(mock.ParadigmCalls))
		}
	})

	t.Run("refresh bypasses cache read", func(t *testing.T) {
		refreshFetcher := NewCachingFetcher(mock, cache, true)

		// Pre-populate cache
		cache.Set("search:maja", []byte(`cached`))

		// With refresh=true, should still call upstream
		callsBefore := len(mock.SearchCalls)
		data, _ := refreshFetcher.Search("maja")
		if len(mock.SearchCalls) != callsBefore+1 {
			t.Errorf("expected upstream call with refresh=true")
		}
		// Should return fresh data, not cached
		if string(data) != `{"words":[]}` {
			t.Errorf("expected fresh data, got %s", data)
		}

		// Cache should be updated with fresh data
		entry, _ := cache.Get("search:maja")
		if string(entry.Value) != `{"words":[]}` {
			t.Errorf("cache not updated after refresh")
		}
	})

	t.Run("works without cache", func(t *testing.T) {
		noCacheFetcher := NewCachingFetcher(mock, nil, false)
		data, err := noCacheFetcher.Search("test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if string(data) != `{"words":[]}` {
			t.Errorf("unexpected response: %s", data)
		}
	})
}

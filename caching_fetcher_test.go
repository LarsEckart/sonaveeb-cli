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
	tmpDir, err := os.MkdirTemp("", "sonaveeb-caching-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cache, err := OpenCacheAt(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open cache: %v", err)
	}
	defer cache.Close()

	mock := &MockFetcher{
		SearchResponse:   []byte(`{"words":[]}`),
		DetailsResponse:  []byte(`{"wordClass":"noun"}`),
		ParadigmResponse: []byte(`[]`),
	}

	fetcher := NewCachingFetcher(mock, cache, false)

	t.Run("caches search results", func(t *testing.T) {
		// First call — should hit upstream
		data1, err := fetcher.Search("puu")
		if err != nil {
			t.Fatalf("fetcher.Search() error: %v", err)
		}
		if len(mock.SearchCalls) != 1 {
			t.Errorf("expected 1 upstream call, got %d", len(mock.SearchCalls))
		}
		if string(data1) != `{"words":[]}` {
			t.Errorf("unexpected response: %s", data1)
		}

		// Second call — should hit cache
		data2, err := fetcher.Search("puu")
		if err != nil {
			t.Fatalf("fetcher.Search() error: %v", err)
		}
		if len(mock.SearchCalls) != 1 {
			t.Errorf("expected still 1 upstream call (cache hit), got %d", len(mock.SearchCalls))
		}
		if string(data2) != string(data1) {
			t.Errorf("cache returned different data")
		}
	})

	t.Run("caches word details", func(t *testing.T) {
		if _, err := fetcher.WordDetails(123); err != nil {
			t.Fatalf("fetcher.WordDetails() error: %v", err)
		}
		if _, err := fetcher.WordDetails(123); err != nil {
			t.Fatalf("fetcher.WordDetails() error: %v", err)
		}
		if len(mock.DetailsCalls) != 1 {
			t.Errorf("expected 1 upstream call, got %d", len(mock.DetailsCalls))
		}
	})

	t.Run("caches paradigm details", func(t *testing.T) {
		if _, err := fetcher.ParadigmDetails(456); err != nil {
			t.Fatalf("fetcher.ParadigmDetails() error: %v", err)
		}
		if _, err := fetcher.ParadigmDetails(456); err != nil {
			t.Fatalf("fetcher.ParadigmDetails() error: %v", err)
		}
		if len(mock.ParadigmCalls) != 1 {
			t.Errorf("expected 1 upstream call, got %d", len(mock.ParadigmCalls))
		}
	})

	t.Run("refresh bypasses cache read", func(t *testing.T) {
		refreshFetcher := NewCachingFetcher(mock, cache, true)

		// Pre-populate cache
		if err := cache.Set("search:maja", []byte(`cached`)); err != nil {
			t.Fatalf("cache.Set() error: %v", err)
		}

		// With refresh=true, should still call upstream
		callsBefore := len(mock.SearchCalls)
		data, err := refreshFetcher.Search("maja")
		if err != nil {
			t.Fatalf("refreshFetcher.Search() error: %v", err)
		}
		if len(mock.SearchCalls) != callsBefore+1 {
			t.Errorf("expected upstream call with refresh=true")
		}
		// Should return fresh data, not cached
		if string(data) != `{"words":[]}` {
			t.Errorf("expected fresh data, got %s", data)
		}

		// Cache should be updated with fresh data
		entry, err := cache.Get("search:maja")
		if err != nil {
			t.Fatalf("cache.Get() error: %v", err)
		}
		if entry == nil {
			t.Fatalf("expected search:maja to be cached")
		}
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

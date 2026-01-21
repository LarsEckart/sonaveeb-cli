package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Fetcher abstracts API access for testability and caching.
type Fetcher interface {
	Search(word string) ([]byte, error)
	WordDetails(wordID int64) ([]byte, error)
	ParadigmDetails(wordID int64) ([]byte, error)
}

// APIFetcher fetches data directly from the Ekilex API.
type APIFetcher struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

func NewAPIFetcher(apiKey string) *APIFetcher {
	return &APIFetcher{
		client:  &http.Client{},
		apiKey:  apiKey,
		baseURL: apiBaseURL,
	}
}

func (f *APIFetcher) Search(word string) ([]byte, error) {
	return f.get("/word/search/" + url.PathEscape(word))
}

func (f *APIFetcher) WordDetails(wordID int64) ([]byte, error) {
	return f.get(fmt.Sprintf("/word/details/%d", wordID))
}

func (f *APIFetcher) ParadigmDetails(wordID int64) ([]byte, error) {
	return f.get(fmt.Sprintf("/paradigm/details/%d", wordID))
}

func (f *APIFetcher) get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", f.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("ekilex-api-key", f.apiKey)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

package sonaveeb

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const apiBaseURL = "https://ekilex.ee/api"

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
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, f.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if f.apiKey != "" {
		req.Header.Set("ekilex-api-key", f.apiKey)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return body, nil
}

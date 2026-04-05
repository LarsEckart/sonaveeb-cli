//go:build integration

package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/lars/sonaveeb-cli/cache"
	"github.com/lars/sonaveeb-cli/config"
	"github.com/lars/sonaveeb-cli/sonaveeb"
)

func getAPIKey(t *testing.T) string {
	key := os.Getenv("EKILEX_API_KEY")
	if key == "" {
		key = config.LoadAPIKey()
	}
	if key == "" {
		t.Skip("EKILEX_API_KEY not set, skipping integration test")
	}
	return key
}

func TestIntegration_NounPuu(t *testing.T) {
	apiKey := getAPIKey(t)
	cfg := cliConfig{Homonym: 1}

	fetcher := sonaveeb.NewAPIFetcher(apiKey)
	var buf bytes.Buffer
	if err := run("puu", cfg, fetcher, &buf); err != nil {
		t.Fatalf("run() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "puu") {
		t.Errorf("expected output to contain 'puu', got:\n%s", output)
	}
	if !strings.Contains(output, "noun") {
		t.Errorf("expected output to contain 'noun', got:\n%s", output)
	}
	if !strings.Contains(output, "ainsuse nimetav") {
		t.Errorf("expected output to contain 'ainsuse nimetav', got:\n%s", output)
	}
	if !strings.Contains(output, "ainsuse omastav") {
		t.Errorf("expected output to contain 'ainsuse omastav', got:\n%s", output)
	}
}

func TestIntegration_VerbTegema(t *testing.T) {
	apiKey := getAPIKey(t)
	cfg := cliConfig{Homonym: 1}

	fetcher := sonaveeb.NewAPIFetcher(apiKey)
	var buf bytes.Buffer
	if err := run("tegema", cfg, fetcher, &buf); err != nil {
		t.Fatalf("run() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tegema") {
		t.Errorf("expected output to contain 'tegema', got:\n%s", output)
	}
	if !strings.Contains(output, "verb") {
		t.Errorf("expected output to contain 'verb', got:\n%s", output)
	}
	if !strings.Contains(output, "ma-tegevusnimi") {
		t.Errorf("expected output to contain 'ma-tegevusnimi', got:\n%s", output)
	}
	if !strings.Contains(output, "da-tegevusnimi") {
		t.Errorf("expected output to contain 'da-tegevusnimi', got:\n%s", output)
	}
}

func TestIntegration_AllForms(t *testing.T) {
	apiKey := getAPIKey(t)
	cfg := cliConfig{Homonym: 1, All: true}

	fetcher := sonaveeb.NewAPIFetcher(apiKey)
	var buf bytes.Buffer
	if err := run("kass", cfg, fetcher, &buf); err != nil {
		t.Fatalf("run() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "kass") {
		t.Errorf("expected output to contain 'kass', got:\n%s", output)
	}
	if !strings.Contains(output, "mitmuse nimetav") {
		t.Errorf("expected output to contain 'mitmuse nimetav', got:\n%s", output)
	}
	if !strings.Contains(output, "mitmuse omastav") {
		t.Errorf("expected output to contain 'mitmuse omastav', got:\n%s", output)
	}
}

func TestIntegration_CachePopulated(t *testing.T) {
	apiKey := getAPIKey(t)

	tmpDir, err := os.MkdirTemp("", "sonaveeb-integration-cache")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cachePath := filepath.Join(tmpDir, "cache.db")
	store, err := cache.OpenCacheAt(cachePath)
	if err != nil {
		t.Fatalf("failed to open cache: %v", err)
	}
	defer func() { _ = store.Close() }()

	apiFetcher := sonaveeb.NewAPIFetcher(apiKey)
	fetcher := cache.NewCachingFetcher(apiFetcher, store, false)

	cfg := cliConfig{Homonym: 1}
	var buf bytes.Buffer
	before := time.Now()
	if err := run("puu", cfg, fetcher, &buf); err != nil {
		t.Fatalf("run() error: %v", err)
	}

	t.Run("search cached", func(t *testing.T) {
		entry, err := store.Get("search:puu")
		if err != nil {
			t.Fatalf("store.Get error: %v", err)
		}
		if entry == nil {
			t.Fatal("expected search:puu to be cached")
		}

		var result sonaveeb.WordSearchResult
		if err := json.Unmarshal(entry.Value, &result); err != nil {
			t.Errorf("cached value is not valid JSON: %v", err)
		}
		if entry.CreatedAt.Before(before.Add(-2 * time.Second)) {
			t.Errorf("created_at %v is too old (test started %v)", entry.CreatedAt, before)
		}
	})

	t.Run("details cached", func(t *testing.T) {
		searchEntry, err := store.Get("search:puu")
		if err != nil {
			t.Fatalf("store.Get error: %v", err)
		}
		if searchEntry == nil {
			t.Fatal("expected search:puu to be cached")
		}
		var result sonaveeb.WordSearchResult
		if err := json.Unmarshal(searchEntry.Value, &result); err != nil {
			t.Fatalf("failed to unmarshal search entry: %v", err)
		}

		estWords := sonaveeb.FilterEstonianWords(result.Words)
		if len(estWords) == 0 {
			t.Skip("no Estonian words found")
		}
		wordID := estWords[0].WordID

		entry, err := store.Get("details:" + toString(wordID))
		if err != nil {
			t.Fatalf("store.Get error: %v", err)
		}
		if entry == nil {
			t.Fatalf("expected details:%d to be cached", wordID)
		}
	})

	t.Run("paradigm cached", func(t *testing.T) {
		searchEntry, err := store.Get("search:puu")
		if err != nil {
			t.Fatalf("store.Get error: %v", err)
		}
		if searchEntry == nil {
			t.Fatal("expected search:puu to be cached")
		}
		var result sonaveeb.WordSearchResult
		if err := json.Unmarshal(searchEntry.Value, &result); err != nil {
			t.Fatalf("failed to unmarshal search entry: %v", err)
		}

		estWords := sonaveeb.FilterEstonianWords(result.Words)
		if len(estWords) == 0 {
			t.Skip("no Estonian words found")
		}
		wordID := estWords[0].WordID

		entry, err := store.Get("paradigm:" + toString(wordID))
		if err != nil {
			t.Fatalf("store.Get error: %v", err)
		}
		if entry == nil {
			t.Fatalf("expected paradigm:%d to be cached", wordID)
		}
	})
}

func toString(id int64) string {
	return strconv.FormatInt(id, 10)
}

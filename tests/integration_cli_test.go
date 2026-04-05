//go:build integration

package main_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LarsEckart/sonaveeb-cli/cache"
	"github.com/LarsEckart/sonaveeb-cli/config"
	"github.com/LarsEckart/sonaveeb-cli/sonaveeb"
)

func getAPIKey(t *testing.T) string {
	t.Helper()

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

	stdout, stderr, exitCode := runCLIWithEnv(t, []string{"EKILEX_API_KEY=" + apiKey}, "puu")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s stdout=%s", exitCode, stderr, stdout)
	}
	if !strings.Contains(stdout, "puu") {
		t.Errorf("expected output to contain 'puu', got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "noun") {
		t.Errorf("expected output to contain 'noun', got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "ainsuse nimetav") {
		t.Errorf("expected output to contain 'ainsuse nimetav', got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "ainsuse omastav") {
		t.Errorf("expected output to contain 'ainsuse omastav', got:\n%s", stdout)
	}
}

func TestIntegration_VerbTegema(t *testing.T) {
	apiKey := getAPIKey(t)

	stdout, stderr, exitCode := runCLIWithEnv(t, []string{"EKILEX_API_KEY=" + apiKey}, "tegema")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s stdout=%s", exitCode, stderr, stdout)
	}
	if !strings.Contains(stdout, "tegema") {
		t.Errorf("expected output to contain 'tegema', got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "verb") {
		t.Errorf("expected output to contain 'verb', got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "ma-tegevusnimi") {
		t.Errorf("expected output to contain 'ma-tegevusnimi', got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "da-tegevusnimi") {
		t.Errorf("expected output to contain 'da-tegevusnimi', got:\n%s", stdout)
	}
}

func TestIntegration_AllForms(t *testing.T) {
	apiKey := getAPIKey(t)

	stdout, stderr, exitCode := runCLIWithEnv(t, []string{"EKILEX_API_KEY=" + apiKey}, "--all", "kass")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s stdout=%s", exitCode, stderr, stdout)
	}
	if !strings.Contains(stdout, "kass") {
		t.Errorf("expected output to contain 'kass', got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "mitmuse nimetav") {
		t.Errorf("expected output to contain 'mitmuse nimetav', got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "mitmuse omastav") {
		t.Errorf("expected output to contain 'mitmuse omastav', got:\n%s", stdout)
	}
}

func TestIntegration_CachePopulated(t *testing.T) {
	apiKey := getAPIKey(t)

	tmpDir, err := os.MkdirTemp("", "sonaveeb-integration-cache")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	before := time.Now()
	stdout, stderr, exitCode := runCLIWithEnv(t, []string{
		"EKILEX_API_KEY=" + apiKey,
		"XDG_CACHE_HOME=" + tmpDir,
	}, "puu")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s stdout=%s", exitCode, stderr, stdout)
	}

	cachePath := filepath.Join(tmpDir, "sonaveeb", "cache.db")
	store, err := cache.OpenCacheAt(cachePath)
	if err != nil {
		t.Fatalf("failed to open cache: %v", err)
	}
	defer func() { _ = store.Close() }()

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

		entry, err := store.Get("details:" + strconv.FormatInt(wordID, 10))
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

		entry, err := store.Get("paradigm:" + strconv.FormatInt(wordID, 10))
		if err != nil {
			t.Fatalf("store.Get error: %v", err)
		}
		if entry == nil {
			t.Fatalf("expected paradigm:%d to be cached", wordID)
		}
	})
}

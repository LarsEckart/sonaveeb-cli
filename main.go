package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	version    = "0.1.1"
	apiBaseURL = "https://ekilex.ee/api"
)

func main() {
	cfg := Config{Homonym: 1}
	flag.BoolVar(&cfg.JSON, "json", false, "Output raw JSON")
	flag.BoolVar(&cfg.All, "all", false, "Show all forms")
	flag.BoolVar(&cfg.Quiet, "quiet", false, "Minimal output")
	flag.BoolVar(&cfg.Quiet, "q", false, "Minimal output (shorthand)")
	flag.BoolVar(&cfg.Version, "version", false, "Print version")
	flag.IntVar(&cfg.Homonym, "homonym", 1, "Select homonym (when multiple exist)")
	flag.BoolVar(&cfg.Refresh, "refresh", false, "Bypass cache and fetch fresh data")
	flag.BoolVar(&cfg.ClearCache, "clear-cache", false, "Clear the cache and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: sonaveeb <word> [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Query Estonian word forms from Ekilex API\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment:\n")
		fmt.Fprintf(os.Stderr, "  EKILEX_API_KEY    API key (required)\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  sonaveeb puu\n")
		fmt.Fprintf(os.Stderr, "  sonaveeb --all tegema\n")
		fmt.Fprintf(os.Stderr, "  sonaveeb --json puu\n")
		fmt.Fprintf(os.Stderr, "  sonaveeb --refresh puu    # bypass cache\n")
	}
	flag.Parse()

	if cfg.Version {
		fmt.Println(version)
		os.Exit(0)
	}

	// Handle --clear-cache before requiring a word
	if cfg.ClearCache {
		cache, err := OpenCache()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening cache: %v\n", err)
			os.Exit(3)
		}

		clearErr := cache.Clear()
		closeErr := cache.Close()

		hadError := false
		if clearErr != nil {
			fmt.Fprintf(os.Stderr, "error clearing cache: %v\n", clearErr)
			hadError = true
		}
		if closeErr != nil {
			fmt.Fprintf(os.Stderr, "error closing cache: %v\n", closeErr)
			hadError = true
		}
		if hadError {
			os.Exit(3)
		}

		fmt.Println("Cache cleared")
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}

	cfg.APIKey = os.Getenv("EKILEX_API_KEY")
	if cfg.APIKey == "" {
		cfg.APIKey = loadConfigFile()
	}
	if cfg.APIKey == "" {
		fmt.Fprintln(os.Stderr, "error: EKILEX_API_KEY not set (use env var or ~/.config/sonaveeb/config)")
		os.Exit(2)
	}

	// Open cache (nil is fine — caching is optional)
	cache, err := OpenCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: cache unavailable: %v\n", err)
	}
	if cache != nil {
		defer cache.Close()
	}

	word := flag.Arg(0)
	apiFetcher := NewAPIFetcher(cfg.APIKey)
	fetcher := NewCachingFetcher(apiFetcher, cache, cfg.Refresh)
	if err := run(word, cfg, fetcher, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if strings.Contains(err.Error(), "not found") {
			os.Exit(1)
		}
		os.Exit(3)
	}
}

func run(word string, cfg Config, fetcher Fetcher, w io.Writer) error {
	searchData, err := fetcher.Search(word)
	if err != nil {
		return err
	}

	searchResult, err := ParseSearchResult(searchData)
	if err != nil {
		return err
	}

	estWords := FilterEstonianWords(searchResult.Words)
	if len(estWords) == 0 {
		return fmt.Errorf("word not found: %s", word)
	}

	selectedWord, err := SelectHomonym(estWords, cfg.Homonym)
	if err != nil {
		return err
	}

	detailsData, err := fetcher.WordDetails(selectedWord.WordID)
	if err != nil {
		return err
	}

	details, err := ParseWordDetails(detailsData)
	if err != nil {
		return err
	}

	paradigmsData, err := fetcher.ParadigmDetails(selectedWord.WordID)
	if err != nil {
		return err
	}

	if cfg.JSON {
		var prettyJSON interface{}
		json.Unmarshal(paradigmsData, &prettyJSON)
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(prettyJSON)
	}

	paradigms, err := ParseParadigms(paradigmsData)
	if err != nil {
		return err
	}
	details.Paradigms = paradigms

	output := FormatOutput(selectedWord.WordValue, details, cfg.Homonym, len(estWords), cfg.All)
	rendered := RenderOutput(output, cfg.Quiet)
	fmt.Fprint(w, rendered)
	return nil
}

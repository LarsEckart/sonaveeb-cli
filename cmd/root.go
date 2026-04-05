package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LarsEckart/sonaveeb-cli/cache"
	"github.com/LarsEckart/sonaveeb-cli/config"
	"github.com/LarsEckart/sonaveeb-cli/render"
	"github.com/LarsEckart/sonaveeb-cli/sonaveeb"
)

type cliConfig struct {
	JSON       bool
	All        bool
	Quiet      bool
	Version    bool
	Homonym    int
	Refresh    bool
	ClearCache bool
}

func Run(args []string, stdout, stderr io.Writer) error {
	cfg := cliConfig{Homonym: 1}
	flagSet := flag.NewFlagSet("sonaveeb-cli", flag.ContinueOnError)
	flagSet.SetOutput(stderr)
	flagSet.BoolVar(&cfg.JSON, "json", false, "Output raw JSON")
	flagSet.BoolVar(&cfg.All, "all", false, "Show all forms")
	flagSet.BoolVar(&cfg.Quiet, "quiet", false, "Minimal output")
	flagSet.BoolVar(&cfg.Quiet, "q", false, "Minimal output (shorthand)")
	flagSet.BoolVar(&cfg.Version, "version", false, "Print version")
	flagSet.IntVar(&cfg.Homonym, "homonym", 1, "Select homonym (when multiple exist)")
	flagSet.BoolVar(&cfg.Refresh, "refresh", false, "Bypass cache and fetch fresh data")
	flagSet.BoolVar(&cfg.ClearCache, "clear-cache", false, "Clear the cache and exit")
	flagSet.Usage = func() {
		_, _ = fmt.Fprintf(stderr, "Usage: sonaveeb-cli <word> [flags]\n\n")
		_, _ = fmt.Fprintf(stderr, "Query Estonian word forms from Ekilex API\n\n")
		_, _ = fmt.Fprintf(stderr, "Flags:\n")
		flagSet.PrintDefaults()
		_, _ = fmt.Fprintf(stderr, "\nEnvironment:\n")
		_, _ = fmt.Fprintf(stderr, "  EKILEX_API_KEY    API key (required)\n")
		_, _ = fmt.Fprintf(stderr, "\nExamples:\n")
		_, _ = fmt.Fprintf(stderr, "  sonaveeb-cli puu\n")
		_, _ = fmt.Fprintf(stderr, "  sonaveeb-cli --all tegema\n")
		_, _ = fmt.Fprintf(stderr, "  sonaveeb-cli --json puu\n")
		_, _ = fmt.Fprintf(stderr, "  sonaveeb-cli --refresh puu    # bypass cache\n")
	}

	if err := flagSet.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return exitWith("", exitCodeUsage)
	}

	if cfg.Version {
		_, _ = fmt.Fprintf(stdout, "sonaveeb-cli version %s\n", appVersion())
		return nil
	}

	if cfg.ClearCache {
		return clearCache(stdout)
	}

	if flagSet.NArg() < 1 {
		flagSet.Usage()
		return exitWith("", exitCodeUsage)
	}

	apiKey := os.Getenv("EKILEX_API_KEY")
	if apiKey == "" {
		apiKey = config.LoadAPIKey()
	}
	if apiKey == "" {
		return exitWith("error: EKILEX_API_KEY not set (use env var or ~/.config/sonaveeb/config)", exitCodeUsage)
	}

	word := flagSet.Arg(0)
	apiFetcher := sonaveeb.NewAPIFetcher(apiKey)

	cacheStore, err := cache.OpenCache()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "warning: cache unavailable: %v\n", err)
	}
	if cacheStore != nil {
		defer func() { _ = cacheStore.Close() }()
	}

	fetcher := cache.NewCachingFetcher(apiFetcher, cacheStore, cfg.Refresh)
	if err := run(word, cfg, fetcher, stdout); err != nil {
		return classifyRunError(err)
	}

	return nil
}

func run(word string, cfg cliConfig, fetcher sonaveeb.Fetcher, stdout io.Writer) error {
	result, err := sonaveeb.LookupWord(word, cfg.Homonym, fetcher)
	if err != nil {
		return err
	}

	if cfg.JSON {
		var prettyJSON any
		if err := json.Unmarshal(result.RawParadigms, &prettyJSON); err != nil {
			return fmt.Errorf("failed to parse paradigms JSON: %w", err)
		}
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(prettyJSON)
	}

	output := sonaveeb.FormatOutput(
		result.SelectedWord.WordValue,
		result.Details,
		cfg.Homonym,
		result.TotalHomonyms,
		cfg.All,
	)
	_, _ = fmt.Fprint(stdout, render.RenderOutput(output, cfg.Quiet))
	return nil
}

func classifyRunError(err error) error {
	if strings.Contains(err.Error(), "not found") {
		return exitWith("error: "+err.Error(), exitCodeNotFound)
	}
	return exitWith("error: "+err.Error(), exitCodeFailure)
}

func clearCache(stdout io.Writer) error {
	cacheStore, err := cache.OpenCache()
	if err != nil {
		return exitWith(fmt.Sprintf("error opening cache: %v", err), exitCodeFailure)
	}

	clearErr := cacheStore.Clear()
	closeErr := cacheStore.Close()

	if clearErr != nil {
		return exitWith(fmt.Sprintf("error clearing cache: %v", clearErr), exitCodeFailure)
	}
	if closeErr != nil {
		return exitWith(fmt.Sprintf("error closing cache: %v", closeErr), exitCodeFailure)
	}

	_, _ = fmt.Fprintln(stdout, "Cache cleared")
	return nil
}

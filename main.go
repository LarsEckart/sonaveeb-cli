package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	version    = "0.1.0"
	apiBaseURL = "https://ekilex.ee/api"
)

type Config struct {
	APIKey     string `toml:"api_key"`
	JSON       bool
	All        bool
	Quiet      bool
	Version    bool
	Homonym    int
	Refresh    bool
	ClearCache bool
}

func loadConfigFile() string {
	// Try local config first
	if key := readKeyFromFile("config"); key != "" {
		return key
	}
	// Then try XDG config
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	configPath := filepath.Join(home, ".config", "sonaveeb", "config")
	return readKeyFromFile(configPath)
}

func readKeyFromFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

type WordSearchResult struct {
	Words []WordMatch `json:"words"`
}

type WordMatch struct {
	WordID    int64  `json:"wordId"`
	WordValue string `json:"wordValue"`
	Lang      string `json:"lang"`
}

type WordDetails struct {
	WordClass string     `json:"wordClass"`
	Paradigms []Paradigm `json:"paradigms"`
	Lexemes   []Lexeme   `json:"lexemes"`
}

type Lexeme struct {
	Pos []PosInfo `json:"pos"`
}

type PosInfo struct {
	Code  string `json:"code"`
	Value string `json:"value"`
}

type Paradigm struct {
	Title            string `json:"title"`
	InflectionTypeNr string `json:"inflectionTypeNr"`
	InflectionType   string `json:"inflectionType"`
	WordClass        string `json:"wordClass"`
	Forms            []Form `json:"paradigmForms"`
}

type Form struct {
	Value     string `json:"value"`
	MorphCode string `json:"morphCode"`
}

var morphLabels = map[string]string{
	// Singular noun cases
	"SgN":   "ainsuse nimetav",
	"SgG":   "ainsuse omastav",
	"SgP":   "ainsuse osastav",
	"SgAdt": "ainsuse lühike sisseütlev",
	"SgIll": "ainsuse sisseütlev",
	"SgIn":  "ainsuse seesütlev",
	"SgEl":  "ainsuse seestütlev",
	"SgAll": "ainsuse alaleütlev",
	"SgAd":  "ainsuse alalütlev",
	"SgAbl": "ainsuse alaltütlev",
	"SgTr":  "ainsuse saav",
	"SgTer": "ainsuse rajav",
	"SgEs":  "ainsuse olev",
	"SgAb":  "ainsuse ilmaütlev",
	"SgKom": "ainsuse kaasaütlev",
	// Plural noun cases
	"PlN":   "mitmuse nimetav",
	"PlG":   "mitmuse omastav",
	"PlP":   "mitmuse osastav",
	"PlIll": "mitmuse sisseütlev",
	"PlIn":  "mitmuse seesütlev",
	"PlEl":  "mitmuse seestütlev",
	"PlAll": "mitmuse alaleütlev",
	"PlAd":  "mitmuse alalütlev",
	"PlAbl": "mitmuse alaltütlev",
	"PlTr":  "mitmuse saav",
	"PlTer": "mitmuse rajav",
	"PlEs":  "mitmuse olev",
	"PlAb":  "mitmuse ilmaütlev",
	"PlKom": "mitmuse kaasaütlev",
	"Rpl":   "mitmuse tüvi",
	// Verb forms
	"Sup":         "ma-tegevusnimi",
	"SupAb":       "ma-tegevusnimi ilmaütlev",
	"SupIn":       "ma-tegevusnimi seesütlev",
	"SupEl":       "ma-tegevusnimi seestütlev",
	"SupTr":       "ma-tegevusnimi saav",
	"SupIps":      "ma-tegevusnimi umbisikuline",
	"Inf":         "da-tegevusnimi",
	"Ger":         "des-vorm",
	"PtsPrPs":     "oleviku kesksõna isikuline",
	"PtsPrIps":    "oleviku kesksõna umbisikuline",
	"PtsPtPs":     "mineviku kesksõna isikuline",
	"PtsPtPsNeg":  "mineviku kesksõna isikuline eitav",
	"PtsPtIps":    "mineviku kesksõna umbisikuline",
	"PtsPtIpsNeg": "mineviku kesksõna umbisikuline eitav",
	"IndPrSg1":    "kindel kõneviis olevikus 1.p ainsus",
	"IndPrSg2":    "kindel kõneviis olevikus 2.p ainsus",
	"IndPrSg3":    "kindel kõneviis olevikus 3.p ainsus",
	"IndPrPl1":    "kindel kõneviis olevikus 1.p mitmus",
	"IndPrPl2":    "kindel kõneviis olevikus 2.p mitmus",
	"IndPrPl3":    "kindel kõneviis olevikus 3.p mitmus",
	"IndPrIps":    "kindel kõneviis olevikus umbisikuline",
	"IndPrIpsNeg": "kindel kõneviis olevikus umbisikuline eitav",
	"IndIpfSg1":   "kindel kõneviis minevikus 1.p ainsus",
	"IndIpfSg2":   "kindel kõneviis minevikus 2.p ainsus",
	"IndIpfSg3":   "kindel kõneviis minevikus 3.p ainsus",
	"IndIpfPl1":   "kindel kõneviis minevikus 1.p mitmus",
	"IndIpfPl2":   "kindel kõneviis minevikus 2.p mitmus",
	"IndIpfPl3":   "kindel kõneviis minevikus 3.p mitmus",
	"IndIpfIps":   "kindel kõneviis minevikus umbisikuline",
	"KndPrSg1":    "tingiv kõneviis olevikus 1.p ainsus",
	"KndPrSg2":    "tingiv kõneviis olevikus 2.p ainsus",
	"KndPrSg3":    "tingiv kõneviis olevikus 3.p ainsus",
	"KndPrPl1":    "tingiv kõneviis olevikus 1.p mitmus",
	"KndPrPl2":    "tingiv kõneviis olevikus 2.p mitmus",
	"KndPrPl3":    "tingiv kõneviis olevikus 3.p mitmus",
	"KndPrIps":    "tingiv kõneviis olevikus umbisikuline",
	"KndPtSg1":    "tingiv kõneviis minevikus 1.p ainsus",
	"KndPtSg2":    "tingiv kõneviis minevikus 2.p ainsus",
	"KndPtSg3":    "tingiv kõneviis minevikus 3.p ainsus",
	"KndPtPl1":    "tingiv kõneviis minevikus 1.p mitmus",
	"KndPtPl2":    "tingiv kõneviis minevikus 2.p mitmus",
	"KndPtPl3":    "tingiv kõneviis minevikus 3.p mitmus",
	"KndPtIps":    "tingiv kõneviis minevikus umbisikuline",
	"KvtPrSg2":    "käskiv kõneviis 2.p ainsus",
	"KvtPrPl1":    "käskiv kõneviis 1.p mitmus",
	"KvtPrPl2":    "käskiv kõneviis 2.p mitmus",
	"KvtPrIps":    "käskiv kõneviis umbisikuline",
	"Neg":         "eitav vorm",
}

var nounMorphCodes = []string{"SgN", "SgG", "SgP", "PlP"}
var verbMorphCodes = []string{"Sup", "Inf", "IndPrSg3", "PtsPtIps"}

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
		if cache != nil {
			cache.Clear()
			cache.Close()
			fmt.Println("Cache cleared")
		} else {
			fmt.Println("No cache to clear")
		}
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
		fmt.Fprintln(os.Stderr, "error: EKILEX_API_KEY not set (use env var or ~/.config/sonaveeb/config.toml)")
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





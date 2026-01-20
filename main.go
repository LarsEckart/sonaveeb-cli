package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	version    = "0.1.0"
	apiBaseURL = "https://ekilex.ee/api"
)

type Config struct {
	APIKey  string `toml:"api_key"`
	JSON    bool
	All     bool
	Quiet   bool
	Version bool
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
	WordID   int64  `json:"wordId"`
	WordValue string `json:"wordValue"`
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
var nounLabels = map[string]string{
	"SgN": "ainsuse nimetav",
	"SgG": "ainsuse omastav",
	"SgP": "ainsuse osastav",
	"PlP": "mitmuse osastav",
}

var verbMorphCodes = []string{"Sup", "Inf", "IndPrSg3", "PtsPtIps"}
var verbLabels = map[string]string{
	"Sup":      "ma-tegevusnimi",
	"Inf":      "da-tegevusnimi",
	"IndPrSg3": "kindel kõneviis olevikus 3.p",
	"PtsPtIps": "mineviku kesksõna umbisikuline",
}

func main() {
	cfg := Config{}
	flag.BoolVar(&cfg.JSON, "json", false, "Output raw JSON")
	flag.BoolVar(&cfg.All, "all", false, "Show all forms")
	flag.BoolVar(&cfg.Quiet, "quiet", false, "Minimal output")
	flag.BoolVar(&cfg.Quiet, "q", false, "Minimal output (shorthand)")
	flag.BoolVar(&cfg.Version, "version", false, "Print version")
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
	}
	flag.Parse()

	if cfg.Version {
		fmt.Println(version)
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

	word := flag.Arg(0)
	if err := run(word, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if strings.Contains(err.Error(), "not found") {
			os.Exit(1)
		}
		os.Exit(3)
	}
}

func run(word string, cfg Config) error {
	client := &http.Client{}

	wordID, wordValue, err := searchWord(client, cfg.APIKey, word)
	if err != nil {
		return err
	}

	details, err := getWordDetails(client, cfg.APIKey, wordID)
	if err != nil {
		return err
	}

	paradigms, rawBody, err := getParadigms(client, cfg.APIKey, wordID)
	if err != nil {
		return err
	}
	details.Paradigms = paradigms

	if cfg.JSON {
		var prettyJSON interface{}
		json.Unmarshal(rawBody, &prettyJSON)
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(prettyJSON)
	}

	printForms(wordValue, details, cfg)
	return nil
}

func searchWord(client *http.Client, apiKey, word string) (int64, string, error) {
	reqURL := fmt.Sprintf("%s/word/search/%s", apiBaseURL, url.PathEscape(word))
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("ekilex-api-key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("API error: %s", resp.Status)
	}

	var result WordSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Words) == 0 {
		return 0, "", fmt.Errorf("word not found: %s", word)
	}

	return result.Words[0].WordID, result.Words[0].WordValue, nil
}

func getWordDetails(client *http.Client, apiKey string, wordID int64) (*WordDetails, error) {
	reqURL := fmt.Sprintf("%s/word/details/%d", apiBaseURL, wordID)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("ekilex-api-key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var details WordDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &details, nil
}

func getParadigms(client *http.Client, apiKey string, wordID int64) ([]Paradigm, []byte, error) {
	reqURL := fmt.Sprintf("%s/paradigm/details/%d", apiBaseURL, wordID)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("ekilex-api-key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("API error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	var paradigms []Paradigm
	if err := json.Unmarshal(body, &paradigms); err != nil {
		return nil, body, fmt.Errorf("failed to parse response: %w", err)
	}

	return paradigms, body, nil
}

func printForms(word string, details *WordDetails, cfg Config) {
	if len(details.Paradigms) == 0 {
		fmt.Println("No paradigm data available")
		return
	}

	paradigm := details.Paradigms[0]
	isVerb := strings.TrimSpace(details.WordClass) == "verb"

	posLabel := "noun"
	if isVerb {
		posLabel = "verb"
	} else if len(details.Lexemes) > 0 && len(details.Lexemes[0].Pos) > 0 {
		code := strings.TrimSpace(details.Lexemes[0].Pos[0].Code)
		switch code {
		case "adj":
			posLabel = "adj"
		case "s":
			posLabel = "noun"
		case "v":
			posLabel = "verb"
			isVerb = true
		}
	}

	if !cfg.Quiet {
		fmt.Printf("%s (%s, type %s)\n", word, posLabel, strings.TrimSpace(paradigm.InflectionTypeNr))
	}

	formMap := make(map[string]string)
	for _, f := range paradigm.Forms {
		code := strings.TrimSpace(f.MorphCode)
		formMap[code] = strings.TrimSpace(f.Value)
	}

	if cfg.All {
		for _, f := range paradigm.Forms {
			code := strings.TrimSpace(f.MorphCode)
			value := strings.TrimSpace(f.Value)
			label := morphLabels[code]
			if label == "" {
				label = code
			}
			if cfg.Quiet {
				fmt.Printf("%s\t%s\n", code, value)
			} else {
				fmt.Printf("  %-45s %s\n", label+":", value)
			}
		}
		return
	}

	var codes []string
	var labels map[string]string
	if isVerb {
		codes = verbMorphCodes
		labels = verbLabels
	} else {
		codes = nounMorphCodes
		labels = nounLabels
	}

	for _, code := range codes {
		value, ok := formMap[code]
		if !ok {
			value = "-"
		}
		label := labels[code]
		if cfg.Quiet {
			fmt.Println(value)
		} else {
			fmt.Printf("  %-35s %s\n", label+":", value)
		}
	}
}

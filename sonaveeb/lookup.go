package sonaveeb

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ParseSearchResult(data []byte) (*WordSearchResult, error) {
	var result WordSearchResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

func ParseWordDetails(data []byte) (*WordDetails, error) {
	var details WordDetails
	if err := json.Unmarshal(data, &details); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &details, nil
}

func ParseParadigms(data []byte) ([]Paradigm, error) {
	var paradigms []Paradigm
	if err := json.Unmarshal(data, &paradigms); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return paradigms, nil
}

func FilterEstonianWords(words []WordMatch) []WordMatch {
	estWords := make([]WordMatch, 0, len(words))
	for _, word := range words {
		if word.Lang == "est" {
			estWords = append(estWords, word)
		}
	}
	return estWords
}

func SelectHomonym(words []WordMatch, homonymIndex int) (WordMatch, error) {
	if len(words) == 0 {
		return WordMatch{}, fmt.Errorf("no words available")
	}

	idx := homonymIndex - 1
	if idx < 0 || idx >= len(words) {
		return WordMatch{}, fmt.Errorf("homonym %d not found (have %d)", homonymIndex, len(words))
	}

	return words[idx], nil
}

func LookupWord(word string, homonymIndex int, fetcher Fetcher) (*LookupResult, error) {
	searchData, err := fetcher.Search(word)
	if err != nil {
		return nil, err
	}

	searchResult, err := ParseSearchResult(searchData)
	if err != nil {
		return nil, err
	}

	estWords := FilterEstonianWords(searchResult.Words)
	if len(estWords) == 0 {
		return nil, fmt.Errorf("word not found: %s", word)
	}

	selectedWord, err := SelectHomonym(estWords, homonymIndex)
	if err != nil {
		return nil, err
	}

	detailsData, err := fetcher.WordDetails(selectedWord.WordID)
	if err != nil {
		return nil, err
	}

	details, err := ParseWordDetails(detailsData)
	if err != nil {
		return nil, err
	}

	paradigmsData, err := fetcher.ParadigmDetails(selectedWord.WordID)
	if err != nil {
		return nil, err
	}

	paradigms, err := ParseParadigms(paradigmsData)
	if err != nil {
		return nil, err
	}
	details.Paradigms = paradigms

	return &LookupResult{
		SelectedWord:  selectedWord,
		TotalHomonyms: len(estWords),
		Details:       details,
		RawParadigms:  paradigmsData,
	}, nil
}

func ExtractEnglishTranslations(details *WordDetails) []string {
	seen := make(map[string]bool)
	translations := make([]string, 0)

	for _, lex := range details.Lexemes {
		for _, group := range lex.SynonymLangGroups {
			if group.Lang != "eng" {
				continue
			}
			for _, syn := range group.Synonyms {
				for _, word := range syn.Words {
					if word.Lang == "eng" && word.WordValue != "" && !seen[word.WordValue] {
						seen[word.WordValue] = true
						translations = append(translations, word.WordValue)
					}
				}
			}
		}
	}

	return translations
}

func DeterminePartOfSpeech(details *WordDetails) (label string, isVerb bool) {
	isVerb = strings.TrimSpace(details.WordClass) == "verb"
	label = "noun"

	if isVerb {
		label = "verb"
	} else if len(details.Lexemes) > 0 && len(details.Lexemes[0].Pos) > 0 {
		code := strings.TrimSpace(details.Lexemes[0].Pos[0].Code)
		switch code {
		case "adj":
			label = "adj"
		case "s":
			label = "noun"
		case "v":
			label = "verb"
			isVerb = true
		}
	}

	return label, isVerb
}

func SelectMorphCodes(isVerb bool) []string {
	if isVerb {
		return verbMorphCodes
	}
	return nounMorphCodes
}

func BuildFormMap(forms []Form) map[string]string {
	formMap := make(map[string]string)
	for _, form := range forms {
		code := strings.TrimSpace(form.MorphCode)
		formMap[code] = strings.TrimSpace(form.Value)
	}
	return formMap
}

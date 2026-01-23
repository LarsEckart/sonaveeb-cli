package main

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

func FilterEstonianWords(words []WordMatch) []WordMatch {
	var estWords []WordMatch
	for _, w := range words {
		if w.Lang == "est" {
			estWords = append(estWords, w)
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

func GetMorphLabel(code string) string {
	if label, ok := morphLabels[code]; ok {
		return label
	}
	return code
}

func ExtractEnglishTranslations(details *WordDetails) []string {
	seen := make(map[string]bool)
	var translations []string

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
	for _, f := range forms {
		code := strings.TrimSpace(f.MorphCode)
		formMap[code] = strings.TrimSpace(f.Value)
	}
	return formMap
}

type FormattedOutput struct {
	Header       string
	Translations []string
	Lines        []FormLine
}

type FormLine struct {
	Code  string
	Label string
	Value string
}

func FormatOutput(word string, details *WordDetails, homonymIndex, totalHomonyms int, showAll bool) FormattedOutput {
	output := FormattedOutput{}

	if len(details.Paradigms) == 0 {
		output.Header = "No paradigm data available"
		return output
	}

	paradigm := details.Paradigms[0]
	posLabel, isVerb := DeterminePartOfSpeech(details)
	inflectionType := strings.TrimSpace(paradigm.InflectionTypeNr)

	if totalHomonyms > 1 {
		output.Header = fmt.Sprintf("%s (%s, type %s)  [%d of %d — use --homonym=N for others]",
			word, posLabel, inflectionType, homonymIndex, totalHomonyms)
	} else {
		output.Header = fmt.Sprintf("%s (%s, type %s)", word, posLabel, inflectionType)
	}

	output.Translations = ExtractEnglishTranslations(details)

	formMap := BuildFormMap(paradigm.Forms)

	if showAll {
		for _, f := range paradigm.Forms {
			code := strings.TrimSpace(f.MorphCode)
			value := strings.TrimSpace(f.Value)
			output.Lines = append(output.Lines, FormLine{
				Code:  code,
				Label: GetMorphLabel(code),
				Value: value,
			})
		}
		return output
	}

	codes := SelectMorphCodes(isVerb)
	for _, code := range codes {
		value, ok := formMap[code]
		if !ok {
			value = "-"
		}
		output.Lines = append(output.Lines, FormLine{
			Code:  code,
			Label: GetMorphLabel(code),
			Value: value,
		})
	}

	return output
}

func RenderOutput(output FormattedOutput, quiet bool) string {
	var sb strings.Builder

	if !quiet && output.Header != "" {
		sb.WriteString(output.Header)
		sb.WriteString("\n")
	}

	if !quiet && len(output.Translations) > 0 {
		sb.WriteString(fmt.Sprintf("  English: %s\n", strings.Join(output.Translations, ", ")))
	}

	for _, line := range output.Lines {
		if quiet {
			sb.WriteString(fmt.Sprintf("%s\t%s\n", line.Code, line.Value))
		} else {
			sb.WriteString(fmt.Sprintf("  %-45s %s\n", line.Label+":", line.Value))
		}
	}

	return sb.String()
}

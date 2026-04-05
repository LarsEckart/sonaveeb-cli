package sonaveeb

import (
	"fmt"
	"strings"
)

func FormatOutput(word string, details *WordDetails, homonymIndex, totalHomonyms int, showAll bool) Output {
	output := Output{}

	if len(details.Paradigms) == 0 {
		output.Header = "No paradigm data available"
		return output
	}

	posLabel, isVerb := DeterminePartOfSpeech(details)

	types := make([]string, 0, len(details.Paradigms))
	seenTypes := make(map[string]bool)
	for _, paradigm := range details.Paradigms {
		inflectionType := strings.TrimSpace(paradigm.InflectionTypeNr)
		if !seenTypes[inflectionType] {
			seenTypes[inflectionType] = true
			types = append(types, inflectionType)
		}
	}
	typeLabel := strings.Join(types, ", ")

	if totalHomonyms > 1 {
		output.Header = fmt.Sprintf("%s (%s, type %s)  [%d of %d — use --homonym=N for others]",
			word, posLabel, typeLabel, homonymIndex, totalHomonyms)
	} else {
		output.Header = fmt.Sprintf("%s (%s, type %s)", word, posLabel, typeLabel)
	}

	output.Translations = ExtractEnglishTranslations(details)

	mergedForms := make(map[string][]string)
	seenValues := make(map[string]map[string]bool)
	for _, paradigm := range details.Paradigms {
		for _, form := range paradigm.Forms {
			code := strings.TrimSpace(form.MorphCode)
			value := strings.TrimSpace(form.Value)

			if seenValues[code] == nil {
				seenValues[code] = make(map[string]bool)
			}
			if !seenValues[code][value] {
				seenValues[code][value] = true
				mergedForms[code] = append(mergedForms[code], value)
			}
		}
	}

	if showAll {
		allCodes := make([]string, 0)
		seenCodes := make(map[string]bool)
		for _, paradigm := range details.Paradigms {
			for _, form := range paradigm.Forms {
				code := strings.TrimSpace(form.MorphCode)
				if !seenCodes[code] {
					seenCodes[code] = true
					allCodes = append(allCodes, code)
				}
			}
		}

		for _, code := range allCodes {
			output.Lines = append(output.Lines, OutputLine{
				Code:  code,
				Label: GetMorphLabel(code),
				Value: strings.Join(mergedForms[code], ", "),
			})
		}

		return output
	}

	for _, code := range SelectMorphCodes(isVerb) {
		value := "-"
		if values, ok := mergedForms[code]; ok {
			value = strings.Join(values, ", ")
		}
		output.Lines = append(output.Lines, OutputLine{
			Code:  code,
			Label: GetMorphLabel(code),
			Value: value,
		})
	}

	return output
}

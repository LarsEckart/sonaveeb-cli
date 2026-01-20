package main

import (
	"testing"
)

func TestFilterEstonianWords(t *testing.T) {
	words := []WordMatch{
		{WordID: 1, WordValue: "puu", Lang: "est"},
		{WordID: 2, WordValue: "tree", Lang: "eng"},
		{WordID: 3, WordValue: "puu", Lang: "est"},
	}

	result := FilterEstonianWords(words)

	if len(result) != 2 {
		t.Errorf("expected 2 Estonian words, got %d", len(result))
	}
	for _, w := range result {
		if w.Lang != "est" {
			t.Errorf("expected lang 'est', got %q", w.Lang)
		}
	}
}

func TestFilterEstonianWords_Empty(t *testing.T) {
	result := FilterEstonianWords([]WordMatch{})
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

func TestSelectHomonym(t *testing.T) {
	words := []WordMatch{
		{WordID: 1, WordValue: "puu1"},
		{WordID: 2, WordValue: "puu2"},
	}

	tests := []struct {
		name    string
		index   int
		wantID  int64
		wantErr bool
	}{
		{"first homonym", 1, 1, false},
		{"second homonym", 2, 2, false},
		{"zero index", 0, 0, true},
		{"out of range", 3, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectHomonym(words, tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectHomonym() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.WordID != tt.wantID {
				t.Errorf("SelectHomonym() = %d, want %d", got.WordID, tt.wantID)
			}
		})
	}
}

func TestSelectHomonym_EmptyList(t *testing.T) {
	_, err := SelectHomonym([]WordMatch{}, 1)
	if err == nil {
		t.Error("expected error for empty list")
	}
}

func TestParseSearchResult(t *testing.T) {
	json := `{"words":[{"wordId":123,"wordValue":"tere","lang":"est"}]}`

	result, err := ParseSearchResult([]byte(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Words) != 1 {
		t.Fatalf("expected 1 word, got %d", len(result.Words))
	}
	if result.Words[0].WordID != 123 {
		t.Errorf("expected wordId 123, got %d", result.Words[0].WordID)
	}
}

func TestParseSearchResult_InvalidJSON(t *testing.T) {
	_, err := ParseSearchResult([]byte("invalid"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDeterminePartOfSpeech(t *testing.T) {
	tests := []struct {
		name      string
		details   *WordDetails
		wantLabel string
		wantVerb  bool
	}{
		{
			name:      "verb by word class",
			details:   &WordDetails{WordClass: "verb"},
			wantLabel: "verb",
			wantVerb:  true,
		},
		{
			name:      "noun default",
			details:   &WordDetails{},
			wantLabel: "noun",
			wantVerb:  false,
		},
		{
			name: "adjective by pos code",
			details: &WordDetails{
				Lexemes: []Lexeme{{Pos: []PosInfo{{Code: "adj"}}}},
			},
			wantLabel: "adj",
			wantVerb:  false,
		},
		{
			name: "verb by pos code",
			details: &WordDetails{
				Lexemes: []Lexeme{{Pos: []PosInfo{{Code: "v"}}}},
			},
			wantLabel: "verb",
			wantVerb:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label, isVerb := DeterminePartOfSpeech(tt.details)
			if label != tt.wantLabel {
				t.Errorf("label = %q, want %q", label, tt.wantLabel)
			}
			if isVerb != tt.wantVerb {
				t.Errorf("isVerb = %v, want %v", isVerb, tt.wantVerb)
			}
		})
	}
}

func TestSelectMorphCodes(t *testing.T) {
	nounCodes := SelectMorphCodes(false)
	verbCodes := SelectMorphCodes(true)

	if len(nounCodes) != 4 {
		t.Errorf("expected 4 noun codes, got %d", len(nounCodes))
	}
	if len(verbCodes) != 4 {
		t.Errorf("expected 4 verb codes, got %d", len(verbCodes))
	}
	if nounCodes[0] != "SgN" {
		t.Errorf("first noun code should be SgN, got %s", nounCodes[0])
	}
	if verbCodes[0] != "Sup" {
		t.Errorf("first verb code should be Sup, got %s", verbCodes[0])
	}
}

func TestGetMorphLabel(t *testing.T) {
	if got := GetMorphLabel("SgN"); got != "ainsuse nimetav" {
		t.Errorf("expected 'ainsuse nimetav', got %q", got)
	}
	if got := GetMorphLabel("unknown"); got != "unknown" {
		t.Errorf("expected 'unknown' for unknown code, got %q", got)
	}
}

func TestBuildFormMap(t *testing.T) {
	forms := []Form{
		{Value: " puu ", MorphCode: " SgN "},
		{Value: "puu", MorphCode: "SgG"},
	}

	result := BuildFormMap(forms)

	if result["SgN"] != "puu" {
		t.Errorf("expected trimmed value 'puu', got %q", result["SgN"])
	}
	if result["SgG"] != "puu" {
		t.Errorf("expected 'puu', got %q", result["SgG"])
	}
}

func TestFormatOutput_NoParadigms(t *testing.T) {
	details := &WordDetails{}
	output := FormatOutput("test", details, 1, 1, false)

	if output.Header != "No paradigm data available" {
		t.Errorf("expected 'No paradigm data available', got %q", output.Header)
	}
}

func TestFormatOutput_WithHomonyms(t *testing.T) {
	details := &WordDetails{
		Paradigms: []Paradigm{{
			InflectionTypeNr: "22",
			Forms: []Form{
				{Value: "puu", MorphCode: "SgN"},
				{Value: "puu", MorphCode: "SgG"},
				{Value: "puud", MorphCode: "SgP"},
				{Value: "puid", MorphCode: "PlP"},
			},
		}},
	}

	output := FormatOutput("puu", details, 1, 3, false)

	if output.Header == "" {
		t.Error("expected non-empty header")
	}
	if len(output.Lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(output.Lines))
	}
}

func TestRenderOutput_Quiet(t *testing.T) {
	output := FormattedOutput{
		Header: "puu (noun, type 22)",
		Lines: []FormLine{
			{Code: "SgN", Label: "ainsuse nimetav", Value: "puu"},
		},
	}

	result := RenderOutput(output, true)

	if result != "SgN\tpuu\n" {
		t.Errorf("unexpected quiet output: %q", result)
	}
}

func TestRenderOutput_Normal(t *testing.T) {
	output := FormattedOutput{
		Header: "puu (noun, type 22)",
		Lines: []FormLine{
			{Code: "SgN", Label: "ainsuse nimetav", Value: "puu"},
		},
	}

	result := RenderOutput(output, false)

	if result == "" {
		t.Error("expected non-empty output")
	}
	if result[:3] != "puu" {
		t.Errorf("expected header to start with 'puu', got %q", result[:3])
	}
}

package sonaveeb

import (
	"strings"
	"testing"
)

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

func TestFormatOutput_MultipleParadigms(t *testing.T) {
	details := &WordDetails{
		Paradigms: []Paradigm{
			{
				InflectionTypeNr: "12",
				Forms: []Form{
					{Value: "väike", MorphCode: "SgN"},
					{Value: "väikese", MorphCode: "SgG"},
					{Value: "väikest", MorphCode: "SgP"},
					{Value: "väikesi", MorphCode: "PlP"},
				},
			},
			{
				InflectionTypeNr: "10",
				Forms: []Form{
					{Value: "väike", MorphCode: "SgN"},
					{Value: "väikse", MorphCode: "SgG"},
					{Value: "väikest", MorphCode: "SgP"},
					{Value: "väikseid", MorphCode: "PlP"},
				},
			},
		},
	}

	output := FormatOutput("väike", details, 1, 1, false)

	if len(output.Lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(output.Lines))
	}
	if !strings.Contains(output.Header, "type 12, 10") {
		t.Errorf("expected header to contain 'type 12, 10', got %q", output.Header)
	}
	for _, line := range output.Lines {
		if line.Code == "SgG" && line.Value != "väikese, väikse" {
			t.Errorf("expected genitive 'väikese, väikse', got %q", line.Value)
		}
	}
}

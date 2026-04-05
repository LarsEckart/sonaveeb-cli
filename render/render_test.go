package render

import (
	"testing"

	"github.com/LarsEckart/sonaveeb-cli/sonaveeb"
)

func TestRenderOutput_Quiet(t *testing.T) {
	output := sonaveeb.Output{
		Header: "puu (noun, type 22)",
		Lines: []sonaveeb.OutputLine{
			{Code: "SgN", Label: "ainsuse nimetav", Value: "puu"},
		},
	}

	result := RenderOutput(output, true)

	if result != "SgN\tpuu\n" {
		t.Errorf("unexpected quiet output: %q", result)
	}
}

func TestRenderOutput_Normal(t *testing.T) {
	output := sonaveeb.Output{
		Header: "puu (noun, type 22)",
		Lines: []sonaveeb.OutputLine{
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

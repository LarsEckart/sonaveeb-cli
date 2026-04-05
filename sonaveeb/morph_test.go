package sonaveeb

import "testing"

func TestGetMorphLabel(t *testing.T) {
	if got := GetMorphLabel("SgN"); got != "ainsuse nimetav" {
		t.Errorf("expected 'ainsuse nimetav', got %q", got)
	}
	if got := GetMorphLabel("unknown"); got != "unknown" {
		t.Errorf("expected 'unknown' for unknown code, got %q", got)
	}
}

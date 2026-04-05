package cmd

import "testing"

func TestAppVersionUsesInjectedVersion(t *testing.T) {
	previous := version
	version = "v9.9.9"
	t.Cleanup(func() {
		version = previous
	})

	if got := appVersion(); got != "v9.9.9" {
		t.Fatalf("expected injected version, got %q", got)
	}
}

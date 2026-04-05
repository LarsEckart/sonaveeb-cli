package main

import (
	"os/exec"
	"strings"
	"testing"
)

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

func TestVersionFlag(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected version command to succeed: %v\n%s", err, out)
	}

	stdout := string(out)
	if !strings.Contains(stdout, "sonaveeb-cli version ") {
		t.Fatalf("expected version output, got:\n%s", stdout)
	}
}

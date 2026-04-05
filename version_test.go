package main

import (
	"os/exec"
	"strings"
	"testing"
)

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

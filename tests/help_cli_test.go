package main_test

import (
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t, "--version")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s stdout=%s", exitCode, stderr, stdout)
	}
	if !strings.Contains(stdout, "sonaveeb-cli version ") {
		t.Fatalf("expected version output, got:\n%s", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", stderr)
	}
}

func TestHelpShowsEnvironmentSection(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t, "-h")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s stdout=%s", exitCode, stderr, stdout)
	}
	if !strings.Contains(stderr, "Environment:") {
		t.Fatalf("expected Environment section, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "EKILEX_API_KEY") {
		t.Fatalf("expected EKILEX_API_KEY help text, got:\n%s", stderr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout for help, got:\n%s", stdout)
	}
}

func TestMissingWordShowsUsageExitCode(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t)
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d, stderr=%s stdout=%s", exitCode, stderr, stdout)
	}
	if !strings.Contains(stderr, "Usage: sonaveeb-cli <word> [flags]") {
		t.Fatalf("expected usage in stderr, got:\n%s", stderr)
	}
}

func TestMissingAPIKeyShowsHelpfulError(t *testing.T) {
	stdout, stderr, exitCode := runCLIWithEnv(t, []string{"EKILEX_API_KEY=", "XDG_CONFIG_HOME=/nonexistent", "HOME=/nonexistent"}, "puu")
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d, stderr=%s stdout=%s", exitCode, stderr, stdout)
	}
	if !strings.Contains(stderr, "error: EKILEX_API_KEY not set") {
		t.Fatalf("expected missing API key error, got:\n%s", stderr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout, got:\n%s", stdout)
	}
}

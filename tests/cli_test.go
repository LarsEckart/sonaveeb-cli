package main_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "sonaveeb-cli-test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	binaryPath = filepath.Join(dir, "sonaveeb-cli")
	build := exec.CommandContext(context.Background(), "go", "build", "-o", binaryPath, "..")
	build.Dir = filepath.Join("..", "tests")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

func runCLI(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	return runCLIWithEnv(t, nil, args...)
}

func runCLIWithEnv(t *testing.T, env []string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), binaryPath, args...)
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run CLI: %v", err)
		}
	}

	return stdout.String(), stderr.String(), exitCode
}

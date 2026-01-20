//go:build integration

package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func getAPIKey(t *testing.T) string {
	key := os.Getenv("EKILEX_API_KEY")
	if key == "" {
		key = loadConfigFile()
	}
	if key == "" {
		t.Skip("EKILEX_API_KEY not set, skipping integration test")
	}
	return key
}

func TestIntegration_NounPuu(t *testing.T) {
	apiKey := getAPIKey(t)

	cfg := Config{
		APIKey:  apiKey,
		Homonym: 1,
	}

	var buf bytes.Buffer
	err := run("puu", cfg, &buf)
	if err != nil {
		t.Fatalf("run() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "puu") {
		t.Errorf("expected output to contain 'puu', got:\n%s", output)
	}
	if !strings.Contains(output, "noun") {
		t.Errorf("expected output to contain 'noun', got:\n%s", output)
	}
	if !strings.Contains(output, "ainsuse nimetav") {
		t.Errorf("expected output to contain 'ainsuse nimetav', got:\n%s", output)
	}
	if !strings.Contains(output, "ainsuse omastav") {
		t.Errorf("expected output to contain 'ainsuse omastav', got:\n%s", output)
	}
}

func TestIntegration_VerbTegema(t *testing.T) {
	apiKey := getAPIKey(t)

	cfg := Config{
		APIKey:  apiKey,
		Homonym: 1,
	}

	var buf bytes.Buffer
	err := run("tegema", cfg, &buf)
	if err != nil {
		t.Fatalf("run() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "tegema") {
		t.Errorf("expected output to contain 'tegema', got:\n%s", output)
	}
	if !strings.Contains(output, "verb") {
		t.Errorf("expected output to contain 'verb', got:\n%s", output)
	}
	if !strings.Contains(output, "ma-tegevusnimi") {
		t.Errorf("expected output to contain 'ma-tegevusnimi', got:\n%s", output)
	}
	if !strings.Contains(output, "da-tegevusnimi") {
		t.Errorf("expected output to contain 'da-tegevusnimi', got:\n%s", output)
	}
}

func TestIntegration_AllForms(t *testing.T) {
	apiKey := getAPIKey(t)

	cfg := Config{
		APIKey:  apiKey,
		Homonym: 1,
		All:     true,
	}

	var buf bytes.Buffer
	err := run("kass", cfg, &buf)
	if err != nil {
		t.Fatalf("run() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "kass") {
		t.Errorf("expected output to contain 'kass', got:\n%s", output)
	}
	if !strings.Contains(output, "mitmuse nimetav") {
		t.Errorf("expected output to contain 'mitmuse nimetav' (plural form), got:\n%s", output)
	}
	if !strings.Contains(output, "mitmuse omastav") {
		t.Errorf("expected output to contain 'mitmuse omastav', got:\n%s", output)
	}
}

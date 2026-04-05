package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func LoadAPIKey() string {
	if key := readKeyFromFile("config"); key != "" {
		return key
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	configPath := filepath.Join(home, ".config", "sonaveeb", "config")
	return readKeyFromFile(configPath)
}

func readKeyFromFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}

	return ""
}

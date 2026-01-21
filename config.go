package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	APIKey     string
	JSON       bool
	All        bool
	Quiet      bool
	Version    bool
	Homonym    int
	Refresh    bool
	ClearCache bool
}

func loadConfigFile() string {
	// Try local config first
	if key := readKeyFromFile("config"); key != "" {
		return key
	}
	// Then try XDG config
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
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

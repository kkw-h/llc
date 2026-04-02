package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Config holds configuration
type Config struct {
	Color          string
	Sort           string
	GroupDirsFirst bool
	HumanReadable  bool
	ShowHidden     bool
	TimeStyle      string
	IgnorePatterns []string
}

// loadConfig loads configuration from ~/.llcrc
func loadConfig() Config {
	config := Config{
		Color:          "auto",
		Sort:           "name",
		TimeStyle:      "default",
		IgnorePatterns: []string{},
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return config
	}

	configPath := filepath.Join(home, configFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "color":
			config.Color = strings.ToLower(value)
		case "sort":
			config.Sort = strings.ToLower(value)
		case "group-directories-first":
			config.GroupDirsFirst = value == "true" || value == "1"
		case "human-readable":
			config.HumanReadable = value == "true" || value == "1"
		case "show-hidden":
			config.ShowHidden = value == "true" || value == "1"
		case "time-style":
			config.TimeStyle = strings.ToLower(value)
		case "ignore":
			config.IgnorePatterns = append(config.IgnorePatterns, value)
		}
	}

	return config
}

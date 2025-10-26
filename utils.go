package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ensureAbsolutePath(path string) string {
	// Check if the path is already absolute
	if filepath.IsAbs(path) {
		return path
	}

	// Get the current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		return "-1"
	}

	// Join the working directory with the relative path
	absolutePath := filepath.Join(workingDir, path)
	return absolutePath
}

func dataPadding(data string, length int) string {
	data = strings.Trim(data, " ")
	if len([]rune(data)) > length {
		data = "#OF#"
	}
	for len([]rune(data)) < length {
		data = data + " "
	}

	return data
}

// ApplyReplacements applies post-processing string replacements to the output
func ApplyReplacements(content string, replacements map[string]string) string {
	result := content
	for search, replace := range replacements {
		result = strings.ReplaceAll(result, search, replace)
	}
	return result
}

// Convert date format YYYY-MM-DD
func DateConvert(dateStr string) (date string, err error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", err
	}

	// Format to DD-MM-YY
	return t.Format("02.01.06"), nil

}

// LoadConfig loads the server configuration from a JSON file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	// Set defaults for missing values
	if config.Delimiter == "" {
		config.Delimiter = ","
	}
	if config.Extensions == nil {
		config.Extensions = []string{"txt", "rmg"}
	}
	if config.Port == 0 {
		config.Port = 8080
	}
	if config.OutputPattern == "" {
		config.OutputPattern = "%s-converted"
	}
	if config.FormatString == "" {
		config.FormatString = "P:14 Y:12 X:12 H:10 MC:6 DT:10"
	}
	if config.Replacements == nil {
		config.Replacements = make(map[string]string)
	}

	return &config, nil
}

package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// loadEnvFile loads environment variables from .env file if it exists
func loadEnvFile() error {
	envPath := filepath.Join(currentDir, ".env")

	// Check if .env file exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(envPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Remove quotes if present
			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}

			// Set environment variable (don't override existing)
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}

	return scanner.Err()
}

// expandEnvVars expands environment variables in the format ${VAR} or $VAR
func expandEnvVars(text string) string {
	// First, expand ${VAR} format
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		varName := strings.TrimPrefix(strings.TrimSuffix(match, "}"), "${")
		if value := os.Getenv(varName); value != "" {
			return value
		}
		// Return empty string if variable doesn't exist to avoid PowerShell errors
		return ""
	})

	// Then, expand $VAR format (but not ${VAR} which we already handled)
	// Need to avoid matching ${VAR} patterns we already processed
	re2 := regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)
	text = re2.ReplaceAllStringFunc(text, func(match string) string {
		varName := strings.TrimPrefix(match, "$")
		if value := os.Getenv(varName); value != "" {
			return value
		}
		// Return empty string if variable doesn't exist to avoid PowerShell errors
		return ""
	})

	return text
}

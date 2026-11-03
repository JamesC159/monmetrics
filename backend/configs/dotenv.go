package configs

import (
	"bufio"
	"os"
	"strings"
)

// loadDotEnv reads KEY=VALUE pairs from the first existing candidate path into the
// process environment. Variables already set in the environment are left untouched,
// so explicit exports and docker-compose env still win over the file.
func loadDotEnv(candidates ...string) {
	for _, path := range candidates {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			key, value, ok := parseDotEnvLine(scanner.Text())
			if !ok {
				continue
			}
			if _, set := os.LookupEnv(key); !set {
				os.Setenv(key, value)
			}
		}
		return
	}
}

// parseDotEnvLine parses one line. Blank lines and # comments are skipped, an optional
// "export " prefix is allowed, and values may be wrapped in single or double quotes.
func parseDotEnvLine(line string) (key, value string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	line = strings.TrimPrefix(line, "export ")

	key, value, found := strings.Cut(line, "=")
	key = strings.TrimSpace(key)
	if !found || key == "" {
		return "", "", false
	}

	value = strings.TrimSpace(value)
	if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
		return key, value[1 : len(value)-1], true
	}
	// Unquoted values may carry a trailing " # comment".
	if idx := strings.Index(value, " #"); idx >= 0 {
		value = strings.TrimSpace(value[:idx])
	}
	return key, value, true
}

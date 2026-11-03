package configs

import (
	"os"
	"testing"
)

func TestParseDotEnvLine(t *testing.T) {
	tests := []struct {
		name, line, key, value string
		ok                     bool
	}{
		{"plain", "SMTP_HOST=localhost", "SMTP_HOST", "localhost", true},
		{"export prefix", "export SMTP_PORT=1025", "SMTP_PORT", "1025", true},
		{"double quoted", `FROM="MonMetrics <a@b.test>"`, "FROM", "MonMetrics <a@b.test>", true},
		{"single quoted", `DB_NAME='monmetrics'`, "DB_NAME", "monmetrics", true},
		{"inline comment", "PORT=8080 # api port", "PORT", "8080", true},
		{"value with equals", "JWT_SECRET=a=b=c", "JWT_SECRET", "a=b=c", true},
		{"comment line", "# SMTP_HOST=x", "", "", false},
		{"blank line", "   ", "", "", false},
		{"no equals", "JUSTTEXT", "", "", false},
		{"empty key", "=value", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value, ok := parseDotEnvLine(tt.line)
			if ok != tt.ok || key != tt.key || value != tt.value {
				t.Errorf("parseDotEnvLine(%q) = (%q, %q, %v), want (%q, %q, %v)",
					tt.line, key, value, ok, tt.key, tt.value, tt.ok)
			}
		})
	}
}

func TestLoadDotEnvDoesNotOverrideExisting(t *testing.T) {
	t.Setenv("SMTP_HOST", "from-shell")
	dir := t.TempDir()
	path := dir + "/.env"
	if err := writeFile(path, "SMTP_HOST=localhost\nSMTP_PORT=1025\n"); err != nil {
		t.Fatal(err)
	}
	loadDotEnv(path)
	if got := getEnv("SMTP_HOST", ""); got != "from-shell" {
		t.Errorf("SMTP_HOST = %q, want shell value to win", got)
	}
	if got := getEnv("SMTP_PORT", ""); got != "1025" {
		t.Errorf("SMTP_PORT = %q, want 1025 from file", got)
	}
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o600)
}

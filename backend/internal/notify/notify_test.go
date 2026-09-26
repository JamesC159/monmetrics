package notify

import (
	"strings"
	"testing"
)

func TestBuildMessageStripsHeaderInjection(t *testing.T) {
	msg := buildMessage("a@example.com", "b@example.com", "Hi\r\nBcc: evil@example.com", "line1\nline2")
	headers, body, _ := strings.Cut(msg, "\r\n\r\n")
	if strings.Contains(headers, "\r\nBcc:") {
		t.Fatalf("header injection not stripped:\n%s", headers)
	}
	if body != "line1\r\nline2" {
		t.Fatalf("body not normalized to CRLF: %q", body)
	}
}

package smtp

import (
	"testing"
	"time"
)

func TestSettingsRejectNonCanonicalBaseURL(t *testing.T) {
	for _, baseURL := range []string{
		"https://app.example.invalid/",
		"https://user@app.example.invalid",
		"https://app.example.invalid/path",
		"ftp://app.example.invalid",
	} {
		if _, err := NewSettings(
			"smtp.example.invalid", 587, "worker", "secret",
			"security@example.invalid", "Harpy Security", baseURL,
			"Invitation", "Open {{INVITATION_URL}}", "<a href=\"{{INVITATION_URL}}\">Open</a>",
			5*time.Second,
		); err == nil {
			t.Fatalf("base URL %q accepted", baseURL)
		}
	}
}

func smtpSettingsFixture(t *testing.T) Settings {
	t.Helper()
	settings, err := NewSettings(
		"smtp.example.invalid", 587, "worker", "secret",
		"security@example.invalid", "Harpy Security",
		"https://app.example.invalid", "Invitation",
		"Open {{INVITATION_URL}}", "<a href=\"{{INVITATION_URL}}\">Open</a>",
		5*time.Second,
	)
	if err != nil {
		t.Fatal(err)
	}
	return settings
}

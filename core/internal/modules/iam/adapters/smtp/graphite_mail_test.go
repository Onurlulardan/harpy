package smtp

import (
	"strings"
	"testing"
)

func TestGraphiteActionHTMLKeepsSinglePlaceholder(t *testing.T) {
	html := GraphiteActionHTML(
		"You're invited",
		"Join your Harpy workspace.",
		"Accept invitation",
		invitationPlaceholder,
	)
	if strings.Count(html, invitationPlaceholder) != 1 {
		t.Fatalf("placeholder count = %d", strings.Count(html, invitationPlaceholder))
	}
	if !strings.Contains(html, "#18181b") || !strings.Contains(html, "Harpy") {
		t.Fatal("graphite invitation HTML is missing theme markers")
	}
	if _, err := newBodyTemplate(html); err != nil {
		t.Fatalf("invitation HTML rejected: %v", err)
	}
}

func TestGraphiteActionHTMLAcceptsLifecyclePlaceholder(t *testing.T) {
	html := GraphiteActionHTML(
		"Reset your password",
		"Choose a new password to continue.",
		"Reset password",
		lifecyclePlaceholder,
	)
	if _, err := newBodyTemplateFor(html, lifecyclePlaceholder); err != nil {
		t.Fatalf("lifecycle HTML rejected: %v", err)
	}
}

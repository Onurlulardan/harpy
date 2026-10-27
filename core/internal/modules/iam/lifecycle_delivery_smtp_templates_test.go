package iam

import (
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/smtp"
)

func TestLifecycleSMTPTemplatesBuildCanonicalSettings(t *testing.T) {
	settings, err := smtp.NewLifecycleSettings(
		"mailpit", 1025, "harpy-local", "secret",
		"security@local.harpy.test", "Harpy Security",
		"http://localhost:5173", 5*time.Second, lifecycleTemplates(),
	)
	if err != nil {
		t.Fatalf("lifecycle SMTP templates: %v", err)
	}
	if _, err := smtp.OpenLifecycleSender(settings); err != nil {
		t.Fatalf("open lifecycle SMTP sender: %v", err)
	}
}

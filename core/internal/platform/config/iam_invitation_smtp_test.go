package config

import (
	"strings"
	"testing"
)

func TestIAMInvitationSMTPRequiresSecretFile(t *testing.T) {
	entries := withoutEnvironmentKey(
		validWorkerEnvironment(),
		"HARPY_IAM_INVITATION_SMTP_PASSWORD_FILE",
	)
	if _, err := loadWorker(entries); err == nil ||
		!strings.Contains(err.Error(), "SMTP_PASSWORD_FILE") {
		t.Fatalf("missing SMTP secret-file error = %v", err)
	}
}

func TestIAMInvitationSMTPRejectsDirectSecret(t *testing.T) {
	entries := append(
		validWorkerEnvironment(),
		"HARPY_IAM_INVITATION_SMTP_PASSWORD=plaintext",
	)
	if _, err := loadWorker(entries); err == nil ||
		!strings.Contains(err.Error(), "unknown key") {
		t.Fatalf("direct SMTP secret error = %v", err)
	}
}

func TestIAMInvitationSMTPRequiresHTTPSInProduction(t *testing.T) {
	entries := withEnvironment(
		validWorkerEnvironment(),
		"HARPY_ENVIRONMENT=production",
		"HARPY_IAM_INVITATION_PUBLIC_BASE_URL=http://app.example.invalid",
	)
	values, err := collectEnvironment(entries, profileWorker)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := loadIAMInvitationSMTP(values, EnvironmentProduction); err == nil {
		t.Fatal("production HTTP invitation base URL was accepted")
	}
}

func TestIAMInvitationSMTPRejectsInvalidTemplate(t *testing.T) {
	entries := withEnvironment(
		validWorkerEnvironment(),
		"HARPY_IAM_INVITATION_TEMPLATE_FILE=C:\\missing-template.json",
	)
	if _, err := loadWorker(entries); err == nil ||
		!strings.Contains(err.Error(), "TEMPLATE_FILE") {
		t.Fatalf("invalid template error = %v", err)
	}
}

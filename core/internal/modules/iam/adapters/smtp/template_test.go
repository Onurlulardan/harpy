package smtp

import (
	"bytes"
	"testing"
)

func TestTemplateWritesCredentialOnlyIntoConfiguredInvitationURL(t *testing.T) {
	template, err := newBodyTemplate("Open {{INVITATION_URL}} now")
	if err != nil {
		t.Fatal(err)
	}
	credential := []byte("abcdefghijklmnopqrstuv.ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopq")
	var output bytes.Buffer
	if _, err := template.writer(
		"https://app.example.invalid", credential, "tenant_a",
	)(&output); err != nil {
		t.Fatal(err)
	}
	want := "Open https://app.example.invalid/accept-invitation/" +
		string(credential) + "?tenant=tenant_a now"
	if output.String() != want {
		t.Fatalf("template output = %q", output.String())
	}
}

func TestTemplateEscapesTenantSlugInLifecycleURL(t *testing.T) {
	template, err := newBodyTemplateFor(
		"Open {{LIFECYCLE_URL}} now", lifecyclePlaceholder,
	)
	if err != nil {
		t.Fatal(err)
	}
	credential := []byte("abcdefghijklmnopqrstuv.ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopq")
	var output bytes.Buffer
	if _, err := template.writerRoute(
		"https://app.example.invalid", "/reset-password/", credential,
		"tenant with+reserved",
	)(&output); err != nil {
		t.Fatal(err)
	}
	want := "Open https://app.example.invalid/reset-password/" +
		string(credential) + "?tenant=tenant+with%2Breserved now"
	if output.String() != want {
		t.Fatalf("template output = %q", output.String())
	}
}

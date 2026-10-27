package billing

import (
	"testing"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

func TestProviderJobEffectKeysAreBoundToReferenceIdentity(t *testing.T) {
	identifier := uuid.MustParse("018f1000-0000-7000-8000-000000000001")
	command := "billing:provider:command:manual:" + identifier.String()
	reconcile := "billing:provider:reconcile:" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if !validProviderEffectKey(
		providerapp.ProviderCommandJobKind, "manual", identifier, command,
	) || !validProviderEffectKey(
		providerapp.ProviderReconciliationJobKind, "manual", identifier, reconcile,
	) {
		t.Fatal("canonical provider effect key was rejected")
	}
	if validProviderEffectKey(
		providerapp.ProviderCommandJobKind, "other", identifier, command,
	) || validProviderEffectKey(
		providerapp.ProviderReconciliationJobKind, "manual", identifier,
		"billing:provider:reconcile:not-a-digest",
	) {
		t.Fatal("mismatched provider effect key was accepted")
	}
}

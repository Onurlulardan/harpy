package providerapp

import "testing"

func TestRegistryCapabilityProviderListsAreSortedAndDefensive(t *testing.T) {
	registry, _, _ := webhookRegistryFixture(
		t, verifiedWebhookFixture(t),
		normalizedWebhookFixture(t, verifiedWebhookFixture(t)),
	)
	commands := registry.CommandProviders()
	reconciliation := registry.ReconciliationProviders()
	webhooks := registry.WebhookProviders()
	if len(commands) != 2 || commands[0] != "manual" || commands[1] != "testpay" {
		t.Fatalf("command providers = %v", commands)
	}
	if len(reconciliation) != 2 || reconciliation[0] != "manual" ||
		reconciliation[1] != "testpay" {
		t.Fatalf("reconciliation providers = %v", reconciliation)
	}
	if len(webhooks) != 1 || webhooks[0] != "testpay" {
		t.Fatalf("webhook providers = %v", webhooks)
	}
	commands[0] = "changed"
	if registry.CommandProviders()[0] != "manual" {
		t.Fatal("registry capability list was not defensive")
	}
}

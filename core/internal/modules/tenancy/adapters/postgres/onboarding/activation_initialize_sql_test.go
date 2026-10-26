package onboarding

import (
	"strings"
	"testing"
)

func TestActivationInitializeUsesTransactionScopedV5Wrapper(t *testing.T) {
	if !strings.Contains(
		activationInitializeSQL,
		"platform.tenancy_activation_initialize_v5(",
	) || strings.Contains(
		activationInitializeSQL,
		"platform.tenancy_activation_initialize_v4(",
	) || !strings.Contains(
		activationInitializeSQL,
		"$10::text",
	) {
		t.Fatal("activation initialize SQL is not pinned to the transaction-scoped v5 wrapper")
	}
}

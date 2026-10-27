package installationactivation

import (
	"strings"
	"testing"
)

func TestBindingUsesInstallationCapabilitySuccessors(t *testing.T) {
	if !strings.Contains(recordEventSQL,
		"platform_async_installation_activation_event_v2") ||
		strings.Contains(recordEventSQL,
			"platform_async_installation_activation_event_v1") {
		t.Fatal("activation event binding does not use capability successor")
	}
	if !strings.Contains(enqueueInitialAdminSetupSQL,
		"platform_async_initial_admin_setup_enqueue_v2") ||
		strings.Contains(enqueueInitialAdminSetupSQL,
			"platform_async_initial_admin_setup_enqueue_v1") {
		t.Fatal("initial admin enqueue binding does not use capability successor")
	}
}

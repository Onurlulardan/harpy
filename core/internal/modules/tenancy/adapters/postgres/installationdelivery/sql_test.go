package installationdelivery

import (
	"strings"
	"testing"
)

func TestInstallationDeliverySQLUsesOnlyNarrowOwnerRoutines(t *testing.T) {
	documents := map[string]string{
		"stage": stageSQL, "read": readSQL, "claim": claimSQL,
		"ack": acknowledgeSQL, "unknown": unknownSQL,
		"secret_deleted": markSecretDeletedSQL, "complete": completeSQL,
		"recover": recoverSQL, "record": recordRecoverySQL,
	}
	for name, document := range documents {
		lower := strings.ToLower(document)
		if !strings.Contains(lower, "platform.tenancy_installation_delivery_") ||
			strings.Contains(lower, "password") || strings.Contains(lower, "plaintext") {
			t.Fatalf("%s SQL contract drifted: %s", name, document)
		}
	}
}

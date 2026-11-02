package postgres

import (
	"strings"
	"testing"
)

func TestIAMInstallationBindingUsesCapabilitySuccessors(t *testing.T) {
	queries := map[string]string{
		"activation": installationActivationAssertSQL,
		"epoch":      installationSecurityEpochSQL,
		"identity":   installationInitializeSQL,
		"directory":  installationDirectorySQL,
		"owner":      installationPlatformOwnerSQL,
	}
	for name, query := range queries {
		if !strings.Contains(query, "installation") ||
			!strings.Contains(query, "_v2") &&
				!strings.Contains(query, "_v3") &&
				!strings.Contains(query, "_v4") &&
				!strings.Contains(query, "directory_upsert_v1") {
			t.Fatalf("%s binding does not use an installation capability successor", name)
		}
	}
}

func readIAMCapabilityMigration(
	t *testing.T, scope, origin, name string,
) []byte {
	t.Helper()
	if scope == "platform" {
		return readIAMPlatformAuthMigration(t, origin, name)
	}
	return readIAMIdentityMigration(t, origin, name)
}
